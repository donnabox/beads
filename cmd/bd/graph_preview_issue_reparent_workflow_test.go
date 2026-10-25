//go:build cgo

package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func TestGraphPreviewIssueReparentWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			const scope = "https://example.invalid/reparent/"
			initArgs := []string{"init", "--graph-mode", "link", "--prefix", "demo", "--scope-url", scope,
				"--skip-hooks", "--skip-agents", "--non-interactive", "--json"}
			if engine == "server" {
				port := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
				if port == "" {
					t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for isolated shared-server qualification")
				}
				initArgs = append(initArgs, "--server", "--external", "--server-host", "127.0.0.1", "--server-port", port, "--server-user", "root")
			}
			call := func(code string, args ...string) string {
				return graphPolicyCLI(t, bd, work, home, nil, code, args...)
			}
			call("", initArgs...)
			oldParent := graphMixedResult[graphstore.IssueRecord](t, call("", "create", "Old", "--id", "demo-old", "--type", "epic", "--json"))
			newParent := graphMixedResult[graphstore.IssueRecord](t, call("", "create", "New", "--id", "demo-new", "--type", "epic", "--json"))
			child := graphMixedResult[graphstore.IssueRecord](t, call("", "create", "Child", "--parent", oldParent.Properties.ID, "--json"))
			if !strings.HasPrefix(child.Properties.ID, oldParent.Properties.ID+".") || len(child.Owned) != 1 {
				t.Fatalf("child was not born under old parent: %+v", child)
			}
			var priorLink graphstore.LinkRecord
			if err := json.Unmarshal(child.Owned[0], &priorLink); err != nil {
				t.Fatal(err)
			}
			call("revision_conflict", "update", child.Properties.ID, "--parent", newParent.Properties.ID,
				"--if-revision", "=00000000000000000000000000000000", "--json")
			call("capability_unavailable", "update", child.Properties.ID, "--parent=", "--json")
			result := graphMixedResult[graphstore.IssueMutationResult](t, call("", "update", child.Properties.ID,
				"--parent", newParent.Properties.ID, "--title", "Moved child", "--priority", "P1",
				"--if-revision", "="+child.Revision, "--json"))
			if !result.Changed || result.Issue.Revision == child.Revision || len(result.Issue.Owned) != 1 ||
				result.Issue.Properties.Title != "Moved child" || result.Issue.Properties.Priority != 1 {
				t.Fatalf("reparent did not update Issue and owned Link: %+v", result)
			}
			var newLink graphstore.LinkRecord
			if err := json.Unmarshal(result.Issue.Owned[0], &newLink); err != nil ||
				newLink.ID == priorLink.ID || newLink.Source != child.ID || newLink.Target != newParent.ID ||
				newLink.Properties["dep_type"] != "parent-child" {
				t.Fatalf("reparent did not replace canonical Link: %+v %v", newLink, err)
			}
			var oldChildren, newChildren []struct {
				ID string `json:"id"`
			}
			if err := json.Unmarshal([]byte(call("", "children", oldParent.Properties.ID, "--json")), &oldChildren); err != nil || len(oldChildren) != 0 {
				t.Fatalf("old parent retained child: %+v %v", oldChildren, err)
			}
			if err := json.Unmarshal([]byte(call("", "children", newParent.Properties.ID, "--json")), &newChildren); err != nil ||
				len(newChildren) != 1 || newChildren[0].ID != strings.TrimPrefix(child.ID, scope) {
				t.Fatalf("new parent did not gain child: %+v %v", newChildren, err)
			}
			noOp := graphMixedResult[graphstore.IssueMutationResult](t, call("", "update", child.Properties.ID,
				"--parent", newParent.Properties.ID, "--if-revision", "="+result.Issue.Revision, "--json"))
			if noOp.Changed || noOp.Issue.Revision != result.Issue.Revision {
				t.Fatalf("same-parent retry was not a no-op: %+v", noOp)
			}
			sameParentEdit := graphMixedResult[graphstore.IssueMutationResult](t, call("", "update", child.Properties.ID,
				"--parent", newParent.Properties.ID, "--description", "Edited in place", "--if-revision", "="+result.Issue.Revision, "--json"))
			if !sameParentEdit.Changed || sameParentEdit.Issue.Properties.Description != "Edited in place" ||
				len(sameParentEdit.Issue.Owned) != 1 {
				t.Fatalf("same-parent scalar edit failed: %+v", sameParentEdit)
			}
			var stableLink graphstore.LinkRecord
			if err := json.Unmarshal(sameParentEdit.Issue.Owned[0], &stableLink); err != nil || stableLink.ID != newLink.ID {
				t.Fatalf("same-parent scalar edit replaced Link: %+v %v", stableLink, err)
			}
			call("revision_conflict", "update", child.Properties.ID, "--parent", newParent.Properties.ID,
				"--if-revision", "1", "--json")
		})
	}
}
