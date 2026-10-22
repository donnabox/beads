package main

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/internal/ui"
	"github.com/steveyegge/beads/internal/uimd"
)

// Familiar Issue commands keep the ordinary top-level JSON contract. Generic
// graph commands continue to use graphPrint's preview envelope.
func withGraphFamiliarStore(fn func(context.Context, *graphstore.Store) (any, string, error)) error {
	return withGraphStoreOutput(fn, graphPrintFamiliar)
}

// Claim already resolves its audit identity this way. Coordination commands
// must use the same default, or an unconfigured --actor cannot heartbeat the
// claim that this CLI just created.
func graphPreviewCoordinationActor() string { return getActorWithGit() }

func graphPrintFamiliar(result any, human string) error {
	if jsonOutput {
		return outputJSON(result)
	}
	if !quietFlag {
		_, err := fmt.Println(human) //nolint:forbidigo // Familiar commands use the existing graph output seam, which has no Cobra writer.
		return err
	}
	return nil
}

type graphIssueDetailOptions struct {
	Long         bool
	LocalTime    bool
	CommentsTail int
}

// renderGraphIssueDetails uses the familiar Issue presentation for the
// current-state human show route. The detail assembler has already translated
// native backing IDs to graph Bead paths in one checked read transaction.
func renderGraphIssueDetails(details *types.IssueDetails, opts graphIssueDetailOptions) string {
	issue := &details.Issue
	var out strings.Builder
	formatTime := func(at time.Time) string {
		if opts.LocalTime {
			at = at.Local()
		}
		return at.Format("2006-01-02 15:04")
	}
	gates := make([]*types.Issue, 0, len(details.GatedBy))
	for _, gate := range details.GatedBy {
		gates = append(gates, &types.Issue{ID: gate.ID, AwaitType: gate.Type})
	}
	fmt.Fprintln(&out, formatIssueHeaderWithGates(issue, gates))
	fmt.Fprintln(&out, formatIssueMetadata(issue))
	for _, gate := range details.GatedBy {
		kind := gate.Type
		if kind == "" {
			kind = "gate"
		}
		if gate.Reason != "" {
			kind += ": " + gate.Reason
		}
		fmt.Fprintf(&out, "Gated by: %s (%s)\n", gate.ID, kind)
	}
	writeSection := func(heading, body string) {
		fmt.Fprintf(&out, "\n%s\n%s\n", ui.RenderBold(heading), uimd.RenderMarkdown(body))
	}
	if issue.Description != "" {
		writeSection("DESCRIPTION", issue.Description)
	} else {
		fmt.Fprintf(&out, "\n%s\n  %s\n", ui.RenderBold("DESCRIPTION"), ui.RenderMuted("(none)"))
	}
	if issue.Design != "" {
		writeSection("DESIGN", issue.Design)
	}
	if issue.Notes != "" {
		writeSection("NOTES", issue.Notes)
	}
	if issue.AcceptanceCriteria != "" {
		writeSection("ACCEPTANCE CRITERIA", issue.AcceptanceCriteria)
	}
	if len(details.Labels) > 0 {
		fmt.Fprintf(&out, "\n%s %s\n", ui.RenderBold("LABELS:"), strings.Join(details.Labels, ", "))
	}
	if metadata := formatIssueCustomMetadata(issue); metadata != "" {
		fmt.Fprintf(&out, "\n%s\n", metadata)
	}
	related := make(map[string]*types.IssueWithDependencyMetadata)
	for _, relation := range []struct {
		rows     []*types.IssueWithDependencyMetadata
		outgoing bool
	}{{details.Dependencies, true}, {details.Dependents, false}} {
		for _, sec := range groupDepSections(relation.rows, relation.outgoing, related) {
			fmt.Fprintf(&out, "\n%s\n", ui.RenderBold(sec.Heading))
			for _, dep := range sec.Deps {
				fmt.Fprintln(&out, formatDependencyLine(sec.Glyph, dep))
			}
		}
	}
	if len(related) > 0 {
		fmt.Fprintf(&out, "\n%s\n", ui.RenderBold("RELATED"))
		ids := make([]string, 0, len(related))
		for id := range related {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		for _, id := range ids {
			fmt.Fprintln(&out, formatDependencyLine("↔", related[id]))
		}
	}
	if len(details.Comments) > 0 {
		fmt.Fprintf(&out, "\n%s\n", ui.RenderBold("COMMENTS"))
		comments := details.Comments
		if opts.CommentsTail > 0 && opts.CommentsTail < len(comments) {
			hidden := len(comments) - opts.CommentsTail
			noun := "comments"
			if hidden == 1 {
				noun = "comment"
			}
			fmt.Fprintf(&out, "  %s\n", ui.RenderMuted(fmt.Sprintf("… %d older %s hidden — bd show %s for the full record", hidden, noun, issue.ID)))
			comments = comments[hidden:]
		}
		for _, comment := range comments {
			fmt.Fprintf(&out, "  %s %s\n", ui.RenderMuted(formatTime(comment.CreatedAt)), comment.Author)
			for _, line := range strings.Split(strings.TrimRight(uimd.RenderMarkdown(comment.Text), "\n"), "\n") {
				fmt.Fprintf(&out, "    %s\n", line)
			}
		}
	}
	if opts.Long {
		fmt.Fprint(&out, formatIssueLongExtras(issue, formatTime))
	}
	return strings.TrimSuffix(out.String(), "\n")
}
