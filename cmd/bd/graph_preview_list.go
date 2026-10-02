package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/utils"
	"github.com/steveyegge/beads/internal/validation"
)

const graphIssueListOutputLimit = 16 << 20

var errGraphListSelector = errors.New("invalid Issue list selector")

func runGraphPreviewList(cmd *cobra.Command) error {
	in, structured, err := graphIssueListInput(cmd, os.Args[1:])
	if err != nil {
		return err
	}
	selectedType := ""
	if cmd.Flags().Changed("bead-type") {
		selector, _ := cmd.Flags().GetString("bead-type")
		selectedType, err = graphPreviewTypeURL(graphPreviewConfig.GraphScopeURL, selector)
		if err != nil {
			return graphFailure("invalid_selector", err.Error(), 2)
		}
	}
	issueFilters := graphIssueListFiltersSelected(cmd, in)
	if issueFilters && selectedType != "" && selectedType != graphstore.IssueTypeURL(graphPreviewConfig.GraphScopeURL) {
		return graphFailure("capability_unavailable", "Issue list filters cannot be combined with a non-Issue --bead-type", 5)
	}
	if !issueFilters {
		return withGraphStoreOutput(func(ctx context.Context, s *graphstore.Store) (any, string, error) {
			snapshot, err := s.CurrentSnapshot(ctx)
			if err != nil {
				return nil, "", err
			}
			page, err := graphBeadListFromSnapshot(snapshot, selectedType, *in.Limit, in.MaxRows)
			if err != nil {
				return nil, "", err
			}
			output, err := renderGraphBeadList(page, s.ScopeURL(), structured, quietFlag)
			return nil, output, err
		}, func(_ any, output string) error { _, err := fmt.Fprint(cmd.OutOrStdout(), output); return err })
	}
	return withGraphStoreOutput(func(ctx context.Context, s *graphstore.Store) (any, string, error) {
		page, err := s.ListIssues(ctx, in.ListRequest)
		if errors.Is(err, storage.ErrValidation) {
			return nil, "", fmt.Errorf("%w: %w", errGraphListSelector, err)
		}
		if err != nil {
			return nil, "", err
		}
		output, err := renderGraphIssueList(page, structured, quietFlag)
		return nil, output, err
	}, func(_ any, output string) error { _, err := fmt.Fprint(cmd.OutOrStdout(), output); return err })
}

func graphIssueListInput(cmd *cobra.Command, argv []string) (listInput, bool, error) {
	fail := func(code, message string, exit int) (listInput, bool, error) {
		return listInput{}, false, graphFailure(code, message, exit)
	}
	if err := graphPreviewFlags(cmd, "flat", "format", "bead-type", "status", "state", "type", "all", "limit", "title", "title-contains", "priority", "priority-min", "priority-max", "label", "label-any", "exclude-label", "pinned", "no-pinned", "sort", "reverse", "assignee", "no-assignee", "due-before", "due-after", "overdue"); err != nil {
		return listInput{}, false, err
	}
	format, _ := cmd.Flags().GetString("format")
	structured := format == "records-json"
	if cmd.Flags().Changed("json") || (jsonOutput && !structured) {
		return fail("capability_unavailable", "graph list preserves the Issue JSON compatibility gate; use --format records-json for experimental graph records", 5)
	}
	flat, _ := cmd.Flags().GetBool("flat")
	if (cmd.Flags().Changed("format") && !structured) || (cmd.Flags().Changed("flat") && !flat) {
		return fail("capability_unavailable", "graph list supports flat output or --format records-json; tree and legacy JSON are unavailable", 5)
	}
	if err := graphIssueListRepeatedFilters(cmd, argv); err != nil {
		return fail("capability_unavailable", err.Error(), 5)
	}
	if cmd.Flags().Changed("status") && cmd.Flags().Changed("state") {
		return fail("invalid_selector", "graph list accepts --status or --state, not both", 2)
	}
	for _, name := range []string{"status", "state", "type", "title", "title-contains", "sort", "assignee", "due-before", "due-after"} {
		value, _ := cmd.Flags().GetString(name)
		if !utf8.ValidString(value) || len(value) > 4096 {
			return fail("invalid_selector", "--"+name+" must be UTF-8 and at most 4096 bytes", 2)
		}
	}
	sort, _ := cmd.Flags().GetString("sort")
	if !map[string]bool{"": true, "priority": true, "created": true, "updated": true, "title": true, "status": true, "type": true}[sort] {
		return fail("capability_unavailable", "graph list supports sort priority, created, updated, title, status or type", 5)
	}
	for _, name := range []string{"priority", "priority-min", "priority-max"} {
		if cmd.Flags().Changed(name) {
			value, _ := cmd.Flags().GetString(name)
			if _, err := validation.ValidatePriority(value); err != nil {
				return fail("invalid_selector", err.Error(), 2)
			}
		}
	}
	pinned, _ := cmd.Flags().GetBool("pinned")
	unpinned, _ := cmd.Flags().GetBool("no-pinned")
	if pinned && unpinned {
		return fail("invalid_selector", "--pinned and --no-pinned are mutually exclusive", 2)
	}
	for _, name := range []string{"label", "label-any", "exclude-label"} {
		if !cmd.Flags().Changed(name) {
			continue
		}
		values, _ := cmd.Flags().GetStringSlice(name)
		if len(values) > 256 {
			return fail("invalid_selector", "too many label values", 2)
		}
		for _, value := range values {
			if !utf8.ValidString(value) || len(value) > 4096 {
				return fail("invalid_selector", "labels must be UTF-8 and at most 4096 bytes each", 2)
			}
		}
		if len(utils.NormalizeLabels(values)) == 0 {
			return fail("invalid_selector", "--"+name+" requires a nonempty label", 2)
		}
	}
	// Reuse the existing limit/config/directory-label policy after unsupported
	// inputs are fenced. Neither the ordinary parser nor its output is replaced.
	in, err := gatherListInput(cmd)
	if err != nil {
		return listInput{}, false, err
	}
	// pflag returns a non-nil empty slice for an omitted StringSlice flag.
	// Admission already refused explicit --exclude-type, including an empty one.
	// Canonicalize only the absent value before the strict storage API boundary.
	if !cmd.Flags().Changed("exclude-type") && len(in.ExcludeTypes) == 0 {
		in.ExcludeTypes = nil
	}
	if in.Limit == nil || *in.Limit < 0 {
		return fail("invalid_selector", "list limit must be non-negative", 2)
	}
	return in, structured, nil
}

// A read-only replay detects spellings the current ordinary parser collapses.
// It does not replace contributor-owned flag values or implement union semantics.
// pflag itself consumes values, optional values and short aliases, so text inside
// a title is never mistaken for another flag. The original FlagSet is untouched.
func graphIssueListRepeatedFilters(cmd *cobra.Command, argv []string) error {
	probe := pflag.NewFlagSet("graph-list-admission", pflag.ContinueOnError)
	probe.SetOutput(io.Discard)
	counts := map[string]*graphIssueListFlagCount{}
	add := func(f *pflag.Flag) {
		if probe.Lookup(f.Name) != nil {
			return
		}
		value := &graphIssueListFlagCount{kind: f.Value.Type()}
		counts[f.Name] = value
		probe.AddFlag(&pflag.Flag{Name: f.Name, Shorthand: f.Shorthand, Value: value, NoOptDefVal: f.NoOptDefVal})
	}
	cmd.Flags().VisitAll(add)
	cmd.InheritedFlags().VisitAll(add)
	if err := probe.Parse(argv); err != nil {
		return fmt.Errorf("cannot validate graph list filter occurrences: %w", err)
	}
	for _, name := range []string{"status", "state", "type", "assignee", "bead-type"} {
		if count := counts[name]; count != nil && count.n > 1 {
			return fmt.Errorf("graph list does not yet support repeated --%s; supply one filter", name)
		}
	}
	return nil
}

type graphIssueListFlagCount struct {
	kind string
	n    int
}

func (v *graphIssueListFlagCount) Set(string) error { v.n++; return nil }
func (v *graphIssueListFlagCount) String() string   { return "" }
func (v *graphIssueListFlagCount) Type() string     { return v.kind }

// The existing Issue filters keep their native query and result contract.
// An unfiltered graph list, or one narrowed only by nominal Bead Type, reads
// the complete checked current inventory and projects its Beads. In
// particular, implicit directory labels must not be silently ignored.
func graphIssueListFiltersSelected(cmd *cobra.Command, in listInput) bool {
	for _, name := range []string{"status", "state", "type", "title", "title-contains", "priority", "priority-min", "priority-max", "label", "label-any", "exclude-label", "pinned", "no-pinned", "sort", "reverse", "assignee", "no-assignee", "due-before", "due-after", "overdue"} {
		if cmd.Flags().Changed(name) {
			return true
		}
	}
	return in.Status != "" || in.IssueType != "" || in.TitleSearch != "" || in.TitleContains != "" || in.Assignee != "" || in.NoAssignee || len(in.Labels) != 0 || len(in.LabelsAny) != 0 || len(in.ExcludeLabels) != 0 || in.Priority != nil || in.PriorityMin != nil || in.PriorityMax != nil || in.PinnedFlag || in.NoPinnedFlag || in.SortBy != "" || in.Reverse || in.DueBefore != nil || in.DueAfter != nil || in.OverdueFlag
}

type graphBeadListPage struct {
	Items   []any `json:"items"`
	HasMore bool  `json:"hasMore"`
}

func graphBeadListFromSnapshot(snapshot graphstore.Snapshot, selectedType string, limit, maxRows int) (graphBeadListPage, error) {
	if selectedType != "" {
		installed := false
		for _, descriptor := range snapshot.Types {
			if descriptor.ID() == selectedType && descriptor.Describes() == graph.KindBead {
				installed = true
				break
			}
		}
		if !installed {
			return graphBeadListPage{}, fmt.Errorf("%w: --bead-type must name an installed Bead Type", graphstore.ErrCapabilityUnavailable)
		}
	}
	page := graphBeadListPage{Items: []any{}}
	for _, value := range snapshot.Records {
		switch record := value.(type) {
		case graphstore.Record:
			if selectedType == "" || record.Type == selectedType {
				page.Items = append(page.Items, record)
			}
		case graphstore.IssueRecord:
			if selectedType == "" || record.Type == selectedType {
				page.Items = append(page.Items, record)
			}
		case graphstore.LinkRecord:
			// A Link is a Resource, but never a Bead.
		default:
			return graphBeadListPage{}, fmt.Errorf("%w: unsupported current Resource projection", graphstore.ErrInvalidStore)
		}
	}
	if maxRows > 0 && len(page.Items) > maxRows {
		return graphBeadListPage{}, fmt.Errorf("%w: Bead list exceeds configured maximum of %d rows", graphstore.ErrLimitExceeded, maxRows)
	}
	if limit > 0 && len(page.Items) > limit {
		page.Items = page.Items[:limit]
		page.HasMore = true
	}
	return page, nil
}

func renderGraphBeadList(page graphBeadListPage, scope string, structured, quiet bool) (string, error) {
	var human strings.Builder
	if !structured && !quiet {
		fmt.Fprintf(&human, "Beads (%d; more: %t; graph preview)\n", len(page.Items), page.HasMore)
		for _, value := range page.Items {
			switch record := value.(type) {
			case graphstore.Record:
				fmt.Fprintf(&human, "  %s  Memory  %s\n", graphMemoryDisplayText(strings.TrimPrefix(record.ID, scope)), graphMemoryDisplayText(record.Properties.Title))
			case graphstore.IssueRecord:
				fmt.Fprintf(&human, "  %s  Issue   %s\n", graphMemoryDisplayText(strings.TrimPrefix(record.ID, scope)), graphMemoryDisplayText(record.Properties.Title))
			}
		}
		if page.HasMore {
			human.WriteString("More Beads exist; increase --limit or use --all within preview bounds.\n")
		}
	}
	var output bytes.Buffer
	if err := graphPrintTo(&output, page, strings.TrimSuffix(human.String(), "\n"), quiet, structured); err != nil {
		return "", err
	}
	if output.Len() > graphIssueListOutputLimit {
		return "", fmt.Errorf("%w: Bead list output exceeds %d bytes; narrow the query", graphstore.ErrLimitExceeded, graphIssueListOutputLimit)
	}
	return output.String(), nil
}

func renderGraphIssueList(page graphstore.IssueListPage, structured, quiet bool) (string, error) {
	var human strings.Builder
	if !structured && !quiet {
		fmt.Fprintf(&human, "Issues (%d; more: %t; graph preview)\n", len(page.Items), page.HasMore)
		for _, item := range page.Items {
			fmt.Fprintf(&human, "%q %q P%d %q\n", item.ID, item.Properties.Status, item.Properties.Priority, item.Properties.Title)
		}
		if page.HasMore {
			human.WriteString("More matching Issues exist; increase --limit or use --all within preview bounds.\n")
		}
	}
	var output bytes.Buffer
	if err := graphPrintTo(&output, page, strings.TrimSuffix(human.String(), "\n"), quiet, structured); err != nil {
		return "", err
	}
	if output.Len() > graphIssueListOutputLimit {
		return "", fmt.Errorf("%w: Issue list output exceeds %d bytes; narrow the query", graphstore.ErrLimitExceeded, graphIssueListOutputLimit)
	}
	return output.String(), nil
}
