package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/internal/utils"
	"github.com/steveyegge/beads/internal/validation"
)

const graphIssueListOutputLimit = 16 << 20

var errGraphListSelector = errors.New("invalid Issue list selector")

func runGraphPreviewList(cmd *cobra.Command) error {
	resourceWorkspace := graphPreviewConfig.GraphSchemaVersion == graphstore.TypesFreshSchemaVersion
	if canonicalOutputSelected(cmd) {
		return runCanonicalCollection(cmd, false)
	}
	// gatherListInput lowers the Changed state of the Issue filters once it has read
	// them, so what the caller typed is read here, before anything gathers.
	typed := graphIssueListTypedOptions(cmd)
	exactType, familyType, err := graphPreviewCollectionTypes(cmd, graphPreviewConfig.GraphScopeURL, false)
	if err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	in, structured, err := graphIssueListInput(cmd, os.Args[1:])
	if err != nil {
		return err
	}
	var parent *graphstore.IssueSelector
	if in.ParentID != "" {
		selected, err := graphPreviewIssueSelector(in.ParentID)
		if err != nil {
			return graphFailure("invalid_selector", err.Error(), 2)
		}
		parent = &selected
		// The graph store resolves the selected path to a native Issue ID in
		// the same checked transaction as the list and its counted projection.
		in.ParentID = ""
	}
	jsonFlag, _ := cmd.Flags().GetBool("json")
	legacyJSON := !resourceWorkspace && !structured && (jsonOutput || jsonFlag)
	if resourceWorkspace && (jsonOutput || jsonFlag || strings.EqualFold(in.formatStr, "json")) {
		structured = true
	}
	selectedType := familyType
	if exactType != "" {
		selectedType = exactType
	}
	// Native Issue-domain filters remain native. Concrete selectors translate to
	// the same authoritative classification column, never a parallel tag.
	in.IssueType = ""
	issueFamily := selectedType == graphstore.IssueTypeURL(graphPreviewConfig.GraphScopeURL)
	if base, classification, native := graphstore.NativeTypeClassification(graphPreviewConfig.GraphScopeURL, selectedType, graph.KindBead); native && base == graphstore.IssueTypeURL(graphPreviewConfig.GraphScopeURL) {
		in.IssueType = classification
		issueFamily = true
	}
	if !issueFamily {
		in.IssueType = ""
	}
	typed = slices.DeleteFunc(typed, func(option string) bool { return option == "type" })
	if resourceWorkspace {
		typed = slices.DeleteFunc(typed, func(option string) bool { return option == "long" })
	}
	planningType := selectedType
	if issueFamily {
		planningType = graphstore.IssueTypeURL(graphPreviewConfig.GraphScopeURL)
	}
	plan := planGraphList(typed, graphIssueListConfigLabel(cmd, in), planningType, graphPreviewIssueTypeURL(graphPreviewConfig.GraphScopeURL))
	if plan.refuse {
		return graphFailure("capability_unavailable", "Issue list filters cannot be combined with a non-Issue --type", 5)
	}
	intersection := exactType != "" && familyType != ""
	if resourceWorkspace {
		if err := graphResourceListCapabilities(in, graphPreviewConfig.GraphScopeURL, exactType, plan.issueQuery); err != nil {
			return graphFailure("capability_unavailable", err.Error(), 5)
		}
	}
	// The Issue query and its legacy JSON renderer cannot apply both Type
	// predicates. Use the Bead inventory for an otherwise unfiltered
	// intersection, and refuse Issue-only filters rather than dropping one.
	if intersection && (plan.issueQuery || legacyJSON) {
		return graphFailure("capability_unavailable", "combined --type/--exact-type selection requires a Bead collection read without Issue-only filters or legacy JSON", 5)
	}
	if legacyJSON {
		if !issueFamily || exactType == graphstore.IssueTypeURL(graphPreviewConfig.GraphScopeURL) {
			return graphFailure("capability_unavailable", "legacy list JSON contains Issues only", 5)
		}
		return withGraphStoreOutput(func(ctx context.Context, s *graphstore.Store) (any, string, error) {
			var page graphstore.LegacyIssueListPage
			var err error
			if parent != nil {
				page, err = s.ListIssuesLegacySelected(ctx, in.ListRequest, *parent)
			} else {
				page, err = s.ListIssuesLegacy(ctx, in.ListRequest)
			}
			if errors.Is(err, storage.ErrValidation) {
				return nil, "", fmt.Errorf("%w: %w", errGraphListSelector, err)
			}
			return page, "", err
		}, func(value any, _ string) error {
			page := value.(graphstore.LegacyIssueListPage)
			encoded, err := json.Marshal(page.Items)
			if err != nil {
				return err
			}
			if len(encoded) > graphIssueListOutputLimit {
				return graphFailure("capability_unavailable", fmt.Sprintf("Issue list JSON exceeds %d bytes; narrow the query", graphIssueListOutputLimit), 5)
			}
			if err := outputJSON(page.Items); err != nil {
				return err
			}
			printTruncationHint(page.HasMore, in.effectiveLimit)
			return nil
		})
	}
	if !resourceWorkspace && issueFamily && exactType != graphstore.IssueTypeURL(graphPreviewConfig.GraphScopeURL) && !intersection {
		plan.issueQuery = true
	}
	if !plan.issueQuery {
		return withGraphStoreOutput(func(ctx context.Context, s *graphstore.Store) (any, string, error) {
			if resourceWorkspace {
				page, err := graphPreviewReadBeadDocuments(ctx, s, graphstore.BeadListRequest{TypeURL: exactType, ConformsTo: familyType, All: in.AllFlag, Limit: *in.Limit, MaxRows: in.MaxRows})
				if err != nil {
					return nil, "", err
				}
				output, err := renderGraphResourceList(page, s.ScopeURL(), in.AllFlag, plan.notice, structured, quietFlag, in.longFormat)
				return nil, output, err
			}
			if selectedType != "" {
				if err := graphPreviewRequireInstalledType(ctx, s, selectedType, graph.KindBead, false); err != nil {
					return nil, "", err
				}
			}
			page, err := s.ListBeads(ctx, graphstore.BeadListRequest{TypeURL: exactType, ConformsTo: familyType, All: in.AllFlag, Limit: *in.Limit, MaxRows: in.MaxRows})
			if err != nil {
				return nil, "", err
			}
			output, err := renderGraphBeadList(page, s.ScopeURL(), in.AllFlag, plan.notice, structured, quietFlag)
			return nil, output, err
		}, func(_ any, output string) error { _, err := fmt.Fprint(cmd.OutOrStdout(), output); return err })
	}
	if resourceWorkspace {
		if tree, _ := cmd.Flags().GetBool("tree"); tree && cmd.Flags().Changed("tree") {
			return graphFailure("capability_unavailable", "Types-aware Issue trees are not available; use the ordinary flat list or graph", 5)
		}
		return withGraphStoreOutput(func(ctx context.Context, s *graphstore.Store) (any, string, error) {
			page, err := s.ListIssueDocuments(ctx, in.ListRequest, parent)
			if errors.Is(err, storage.ErrValidation) {
				return nil, "", fmt.Errorf("%w: %w", errGraphListSelector, err)
			}
			if err != nil {
				return nil, "", err
			}
			output, err := renderGraphResourceList(page, s.ScopeURL(), in.AllFlag, plan.notice, structured, quietFlag, in.longFormat)
			return nil, output, err
		}, func(_ any, output string) error { _, err := fmt.Fprint(cmd.OutOrStdout(), output); return err })
	}
	if parent != nil && !in.ReadyFlag && in.prettyFormat && !structured && !legacyJSON && !in.longFormat && !quietFlag {
		return withGraphStoreOutput(func(ctx context.Context, s *graphstore.Store) (any, string, error) {
			tree, err := s.ListIssueTreeSelected(ctx, in.ListRequest, *parent)
			if errors.Is(err, storage.ErrValidation) {
				return nil, "", fmt.Errorf("%w: %w", errGraphListSelector, err)
			}
			if err != nil {
				return nil, "", err
			}
			output, err := renderGraphIssueTree(tree, plan.notice, in.SortBy, in.Reverse)
			return nil, output, err
		}, func(_ any, output string) error { _, err := fmt.Fprint(cmd.OutOrStdout(), output); return err })
	}
	if in.longFormat && !structured && !quietFlag {
		return withGraphStoreOutput(func(ctx context.Context, s *graphstore.Store) (any, string, error) {
			var page graphstore.LegacyIssueListPage
			var err error
			if parent != nil {
				page, err = s.ListIssuesLegacySelected(ctx, in.ListRequest, *parent)
			} else {
				page, err = s.ListIssuesLegacy(ctx, in.ListRequest)
			}
			if errors.Is(err, storage.ErrValidation) {
				return nil, "", fmt.Errorf("%w: %w", errGraphListSelector, err)
			}
			if err != nil {
				return nil, "", err
			}
			output, err := renderGraphIssueLongList(page, plan.notice)
			return nil, output, err
		}, func(_ any, output string) error { _, err := fmt.Fprint(cmd.OutOrStdout(), output); return err })
	}
	return withGraphStoreOutput(func(ctx context.Context, s *graphstore.Store) (any, string, error) {
		var page graphstore.IssueListPage
		var err error
		if parent != nil {
			page, err = s.ListIssuesSelected(ctx, in.ListRequest, *parent)
		} else {
			page, err = s.ListIssues(ctx, in.ListRequest)
		}
		if errors.Is(err, storage.ErrValidation) {
			return nil, "", fmt.Errorf("%w: %w", errGraphListSelector, err)
		}
		if err != nil {
			return nil, "", err
		}
		output, err := renderGraphIssueList(page, plan.notice, structured, quietFlag, s.ScopeURL())
		return nil, output, err
	}, func(_ any, output string) error { _, err := fmt.Fprint(cmd.OutOrStdout(), output); return err })
}

// The filtered native Issue seam cannot silently drop enrichment or replace an
// exact abstract nominal selector with its concrete family. Check before any
// read, not after pagination or after producing part of the output.
func graphResourceListCapabilities(in listInput, scope, exact string, issueQuery bool) error {
	if in.IncludeComments {
		return fmt.Errorf("Types-aware complete Resource lists do not support --include-comments")
	}
	if issueQuery && exact == graphstore.IssueTypeURL(scope) {
		return fmt.Errorf("Issue-only filters cannot be combined with exact abstract Issue selection; use --type issue for its family")
	}
	return nil
}

func graphIssueListInput(cmd *cobra.Command, argv []string) (listInput, bool, error) {
	fail := func(code, message string, exit int) (listInput, bool, error) {
		return listInput{}, false, graphFailure(code, message, exit)
	}
	if err := graphPreviewFlags(cmd, "flat", "tree", "format", "status", "state", "type", "exact-type", "conforms-to", "all", "limit", "id", "title", "title-contains", "priority", "priority-min", "priority-max", "label", "label-any", "exclude-label", "pinned", "no-pinned", "parent", "filter-parent", "no-parent", "ready", "sort", "reverse", "assignee", "no-assignee", "due-before", "due-after", "overdue", "include-comments", "long"); err != nil {
		return listInput{}, false, err
	}
	format, _ := cmd.Flags().GetString("format")
	structured := format == "records-json"
	legacyFormatJSON := strings.EqualFold(format, "json")
	jsonFlag, _ := cmd.Flags().GetBool("json")
	if cmd.Flags().Changed("json") && !jsonFlag && !jsonOutput {
		return fail("capability_unavailable", "graph list does not support explicit --json=false", 5)
	}
	if cmd.Flags().Changed("include-comments") && (structured || !(jsonOutput || jsonFlag || legacyFormatJSON)) {
		return fail("capability_unavailable", "--include-comments requires legacy Issue --json", 5)
	}
	flat, _ := cmd.Flags().GetBool("flat")
	tree, _ := cmd.Flags().GetBool("tree")
	parentValue, _ := cmd.Flags().GetString("parent")
	filterParentValue, _ := cmd.Flags().GetString("filter-parent")
	parentTree := parentValue != "" || filterParentValue != ""
	if (cmd.Flags().Changed("format") && !structured && !legacyFormatJSON) || (cmd.Flags().Changed("flat") && !flat) || (cmd.Flags().Changed("tree") && tree && !parentTree) {
		return fail("capability_unavailable", "graph list supports flat output, Issue-only --json/--format json, or --format records-json; tree output is unavailable", 5)
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
	if !map[string]bool{"": true, "priority": true, "created": true, "updated": true, "closed": true, "title": true, "status": true, "type": true, "id": true, "assignee": true}[sort] {
		return fail("capability_unavailable", "graph list supports sort priority, created, updated, closed, title, status, type, id or assignee", 5)
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
	if (cmd.Flags().Changed("parent") || cmd.Flags().Changed("filter-parent")) && in.ParentID == "" {
		return fail("invalid_selector", "--parent requires an Issue ID", 2)
	}
	if in.ParentID != "" && !in.ReadyFlag && quietFlag && !structured && !in.jsonOutput && !jsonFlag && !legacyFormatJSON && !flat && !(cmd.Flags().Changed("tree") && !tree) {
		return fail("capability_unavailable", "graph list --parent --quiet currently requires --flat or JSON", 5)
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
	for _, name := range []string{"status", "state", "type", "assignee"} {
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
// the complete checked current inventory and projects its Beads. The options
// below select the Issue query instead, and the notice names them in this order.
var graphIssueListOptions = []string{"status", "state", "type", "id", "title", "title-contains", "priority", "priority-min", "priority-max", "label", "label-any", "exclude-label", "pinned", "no-pinned", "parent", "filter-parent", "no-parent", "ready", "sort", "reverse", "assignee", "no-assignee", "due-before", "due-after", "overdue", "long"}

// graphIssueListTypedOptions names the Issue options the caller supplied, in
// canonical order. A supplied flag counts even when its value, such as an empty
// --assignee=, adds no restriction.
func graphIssueListTypedOptions(cmd *cobra.Command) []string {
	typed := []string{}
	for _, name := range graphIssueListOptions {
		if cmd.Flags().Changed(name) {
			typed = append(typed, name)
		}
	}
	return typed
}

// graphIssueListConfigLabel is the directory label the ordinary list parser
// applied from configuration, which it does only while neither --label nor
// --label-any was typed. Implicit directory labels must not be silently ignored,
// so it selects the Issue query like a typed option does.
func graphIssueListConfigLabel(cmd *cobra.Command, in listInput) string {
	if len(in.LabelsAny) == 0 || cmd.Flags().Changed("label") || cmd.Flags().Changed("label-any") {
		return ""
	}
	return in.LabelsAny[0]
}

// graphListPlan says which query a list runs and what its human output tells
// the caller about that choice.
type graphListPlan struct {
	issueQuery bool   // read through the native Issue query rather than the Bead inventory
	refuse     bool   // typed Issue options cannot be combined with a non-Issue Bead Type
	notice     string // line directly under the human header, empty for none
}

// planGraphList routes a list. Typed Issue options select the Issue query and a
// non-Issue Bead Type refuses them. A configured directory label also selects
// it, but a Memory carries no labels, so the label never refuses a Bead Type: it
// is simply not applied to one, and the notice says so. Naming the Issue Type
// already narrowed the list to Issues, so that route needs no notice.
func planGraphList(typed []string, label, selectedType, issueType string) graphListPlan {
	switch {
	case len(typed) > 0 && selectedType != "" && selectedType != issueType:
		return graphListPlan{refuse: true}
	case len(typed) > 0 || (label != "" && (selectedType == "" || selectedType == issueType)):
		plan := graphListPlan{issueQuery: true}
		if selectedType == "" {
			plan.notice = graphIssueQueryNotice(typed, label)
		}
		return plan
	case label != "":
		return graphListPlan{notice: fmt.Sprintf("directory.labels %q is not applied to this Bead Type.", label)}
	}
	return graphListPlan{}
}

func graphIssueQueryNotice(typed []string, label string) string {
	selectedBy := make([]string, 0, len(typed)+1)
	for _, name := range typed {
		selectedBy = append(selectedBy, "--"+name)
	}
	if label != "" {
		selectedBy = append(selectedBy, fmt.Sprintf("directory.labels %q", label))
	}
	return "Memories are not listed (Issue query selected by: " + strings.Join(selectedBy, ", ") + ")."
}

func renderGraphBeadList(page graphstore.BeadListPage, scope string, all bool, notice string, structured, quiet bool) (string, error) {
	var human strings.Builder
	if !structured && !quiet {
		fmt.Fprintf(&human, "Beads (%d; more: %t; graph preview)\n", len(page.Items), page.HasMore)
		if notice != "" {
			human.WriteString(notice + "\n")
		}
		for _, value := range page.Items {
			fmt.Fprintln(&human, graphHumanRow(scope, value))
		}
		if page.HasMore {
			if all {
				human.WriteString("More Beads exist; increase --limit within preview bounds.\n")
			} else {
				human.WriteString("More Beads exist; increase --limit or use --all within preview bounds.\n")
			}
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

func renderGraphResourceList(page graphstore.BeadDocumentPage, scope string, all bool, notice string, structured, quiet, long bool) (string, error) {
	ordinary := graphstore.BeadListPage{Items: make([]any, 0, len(page.Items)), HasMore: page.HasMore}
	for _, document := range page.Items {
		ordinary.Items = append(ordinary.Items, document)
	}
	if !long || structured || quiet {
		return renderGraphBeadList(ordinary, scope, all, notice, structured, quiet)
	}
	var human strings.Builder
	fmt.Fprintf(&human, "Beads (%d; more: %t; graph preview)\n", len(page.Items), page.HasMore)
	if notice != "" {
		human.WriteString(notice + "\n")
	}
	for _, document := range page.Items {
		human.WriteString(graphHumanDocument(scope, document, "", false, true) + "\n\n")
	}
	if page.HasMore {
		human.WriteString("More Beads exist; increase --limit or use --all within preview bounds.\n")
	}
	if human.Len() > graphIssueListOutputLimit {
		return "", fmt.Errorf("%w: Bead list output exceeds %d bytes; narrow the query", graphstore.ErrLimitExceeded, graphIssueListOutputLimit)
	}
	return human.String(), nil
}

func renderGraphIssueList(page graphstore.IssueListPage, notice string, structured, quiet bool, scope string) (string, error) {
	var human strings.Builder
	if !structured && !quiet {
		fmt.Fprintf(&human, "Issues (%d; more: %t; graph preview)\n", len(page.Items), page.HasMore)
		if notice != "" {
			human.WriteString(notice + "\n")
		}
		for _, item := range page.Items {
			fmt.Fprintln(&human, graphHumanRow(scope, item))
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

func renderGraphIssueLongList(page graphstore.LegacyIssueListPage, notice string) (string, error) {
	var human strings.Builder
	fmt.Fprintf(&human, "\nFound %d issues:\n\n", len(page.Items))
	if notice != "" {
		human.WriteString(notice + "\n\n")
	}
	for _, row := range page.Items {
		if row == nil || row.Issue == nil {
			return "", fmt.Errorf("%w: Issue list returned an empty row", graphstore.ErrInvalidStore)
		}
		formatIssueLong(&human, row.Issue, row.Labels, false)
	}
	if page.HasMore {
		human.WriteString("More matching Issues exist; increase --limit or use --all within preview bounds.\n")
	}
	if human.Len() > graphIssueListOutputLimit {
		return "", fmt.Errorf("%w: Issue list output exceeds %d bytes; narrow the query", graphstore.ErrLimitExceeded, graphIssueListOutputLimit)
	}
	return human.String(), nil
}

func renderGraphIssueTree(tree graphstore.IssueTreeSelected, notice, sortBy string, reverse bool) (string, error) {
	if tree.Root == nil {
		return "", fmt.Errorf("%w: Issue tree has no root", graphstore.ErrInvalidStore)
	}
	var output strings.Builder
	if notice != "" {
		output.WriteString(notice + "\n")
	}
	if len(tree.Items) == 0 {
		fmt.Fprintf(&output, "Issue '%s' has no children\n", tree.Root.ID)
		return output.String(), nil
	}
	byID := map[string]*types.Issue{tree.Root.ID: tree.Root}
	for _, item := range tree.Items {
		if item == nil || item.ID == "" || byID[item.ID] != nil {
			return "", fmt.Errorf("%w: invalid Issue tree row", graphstore.ErrInvalidStore)
		}
		byID[item.ID] = item
	}
	compare := compareIssuesForTree(sortBy, reverse)
	var walk func(string, string, map[string]bool) error
	walk = func(parentID, prefix string, onPath map[string]bool) error {
		childIDs := tree.Children[parentID]
		children := make([]*types.Issue, 0, len(childIDs))
		for _, childID := range childIDs {
			child := byID[childID]
			if child == nil {
				return fmt.Errorf("%w: Issue tree references an absent child", graphstore.ErrInvalidStore)
			}
			children = append(children, child)
		}
		slices.SortFunc(children, compare)
		for i, child := range children {
			last := i == len(children)-1
			connector, extension := "├── ", "│   "
			if last {
				connector, extension = "└── ", "    "
			}
			fmt.Fprintf(&output, "%s%s%s", prefix, connector, formatPrettyIssueGated(child, nil))
			if onPath[child.ID] {
				output.WriteString(" " + treeCycleMarker + "\n")
				if output.Len() > graphIssueListOutputLimit {
					return fmt.Errorf("%w: Issue tree output exceeds %d bytes; narrow the query", graphstore.ErrLimitExceeded, graphIssueListOutputLimit)
				}
				continue
			}
			output.WriteByte('\n')
			if output.Len() > graphIssueListOutputLimit {
				return fmt.Errorf("%w: Issue tree output exceeds %d bytes; narrow the query", graphstore.ErrLimitExceeded, graphIssueListOutputLimit)
			}
			onPath[child.ID] = true
			if err := walk(child.ID, prefix+extension, onPath); err != nil {
				return err
			}
			delete(onPath, child.ID)
		}
		return nil
	}
	fmt.Fprintln(&output, formatPrettyIssueGated(tree.Root, nil))
	if err := walk(tree.Root.ID, "", map[string]bool{tree.Root.ID: true}); err != nil {
		return "", err
	}
	open, inProgress := 0, 0
	for _, item := range append([]*types.Issue{tree.Root}, tree.Items...) {
		switch item.Status {
		case types.StatusOpen:
			open++
		case types.StatusInProgress:
			inProgress++
		}
	}
	output.WriteString("\n" + strings.Repeat("-", 80) + "\n")
	output.WriteString(listFooterLine(len(tree.Items)+1, open, inProgress, false, false, "") + "\n\n")
	output.WriteString("Status: ○ open  ◐ in_progress  ● blocked  ✓ closed  ❄ deferred\n")
	output.WriteString("Priority: P0–P4 (label only; not a status icon)\n")
	if output.Len() > graphIssueListOutputLimit {
		return "", fmt.Errorf("%w: Issue tree output exceeds %d bytes; narrow the query", graphstore.ErrLimitExceeded, graphIssueListOutputLimit)
	}
	return output.String(), nil
}
