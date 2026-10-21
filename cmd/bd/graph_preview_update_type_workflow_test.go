//go:build cgo

package main

import (
	"testing"

	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
)

func TestGraphPreviewUpdateTypeWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			const scope = "https://example.invalid/update-type/"
			work, home := graphListWorkspace(t, bd, engine, scope)
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			call("create", "Work", "--id", "beads/work")
			before := graphMixedResult[graphstore.IssueRecord](t, call("show", "beads/work", "--format", "graph-json"))
			changed := graphMixedResult[graphstore.IssueMutationResult](t, call("update", "beads/work", "--type", "enhancement", "--if-revision", before.Revision))
			if !changed.Changed || changed.Issue.Properties.IssueType != types.TypeFeature || changed.Issue.Type != before.Type {
				t.Fatalf("type alias update: %+v", changed)
			}
			repeated := graphMixedResult[graphstore.IssueMutationResult](t, call("update", "beads/work", "--type", "feature", "--if-revision", changed.Issue.Revision))
			if repeated.Changed || repeated.Issue.Revision != changed.Issue.Revision {
				t.Fatalf("type no-op: %+v", repeated)
			}
			graphPolicyCLI(t, bd, work, home, nil, "invalid_properties", "update", "beads/work", "--type", "not-configured", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "revision_conflict", "update", "beads/work", "--type", "bug", "--if-revision", before.Revision, "--json")
		})
	}
}
