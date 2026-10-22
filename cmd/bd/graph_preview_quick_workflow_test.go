//go:build cgo

package main

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func TestGraphPreviewQuickCaptureWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			initArgs := []string{"init", "--graph-mode", "link", "--prefix", "quick", "--scope-url", "https://example.invalid/quick-capture/", "--skip-hooks", "--skip-agents", "--non-interactive"}
			if engine == "server" {
				port := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
				if port == "" {
					t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for ordinary shared-server qualification")
				}
				initArgs = append(initArgs, "--server", "--external", "--server-host", "127.0.0.1", "--server-port", port, "--server-user", "root")
			}
			graphPolicyCLI(t, bd, work, home, nil, "", append(initArgs, "--json")...)
			created := strings.TrimSpace(graphPolicyCLI(t, bd, work, home, nil, "", "q", "Quick", "capture", "--priority", "P1", "--type", "bug", "--labels", "triage", "--json"))
			if !strings.HasPrefix(created, "beads/quick-") || strings.Contains(created, "\n") {
				t.Fatalf("quick capture did not print exactly one canonical ID: %q", created)
			}
			record := graphMixedResult[graphstore.IssueRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", "show", created, "--format", "graph-json"))
			if record.ID != "https://example.invalid/quick-capture/"+created || record.Properties.Title != "Quick capture" || record.Properties.Priority != 1 || record.Properties.IssueType != "bug" || len(record.Properties.Labels) != 1 || record.Properties.Labels[0] != "triage" {
				t.Fatalf("quick capture lost Issue fields: ID=%q title=%q priority=%d type=%q labels=%v", record.ID, record.Properties.Title, record.Properties.Priority, record.Properties.IssueType, record.Properties.Labels)
			}
			// Like ordinary q, -- ends option parsing so a leading hyphen is title text.
			hyphenID := strings.TrimSpace(graphPolicyCLI(t, bd, work, home, nil, "", "q", "--", "-follow-up"))
			hyphen := graphMixedResult[graphstore.IssueRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", "show", hyphenID, "--format", "graph-json"))
			if hyphen.Properties.Title != "-follow-up" {
				t.Fatalf("quick capture changed a flag-looking title: %q", hyphen.Properties.Title)
			}
			childID := strings.TrimSpace(graphPolicyCLI(t, bd, work, home, nil, "", "q", "Child", "--parent", created, "--labels", "own", "--json"))
			if strings.Contains(childID, "\n") || !strings.HasPrefix(childID, created+".") {
				t.Fatalf("quick child did not print one dotted ID: %q", childID)
			}
			child := graphMixedResult[graphstore.IssueRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", "show", childID, "--format", "graph-json"))
			if child.Properties.Title != "Child" || !reflect.DeepEqual(child.Properties.Labels, []string{"own", "triage"}) || len(child.Owned) != 1 {
				t.Fatalf("quick child lost title, inherited labels, or owner Link: title=%q labels=%v owned=%d", child.Properties.Title, child.Properties.Labels, len(child.Owned))
			}
			var link graphstore.LinkRecord
			if err := json.Unmarshal(child.Owned[0], &link); err != nil || link.Source != child.ID || link.Target != record.ID || link.Properties["dep_type"] != "parent-child" {
				t.Fatalf("quick child has no canonical parent-child Link: %+v %v", link, err)
			}
			graphPolicyCLI(t, bd, work, home, nil, "invalid_selector", "q", "Refused", "--parent=", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "not_found", "q", "Refused", "--parent", "missing", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "permission_denied", "q", "Forbidden", "--readonly", "--json")
			// The helper decodes stderr as one JSON object; a second plain or JSON
			// diagnostic would make this assertion fail.
			graphPolicyCLI(t, bd, work, home, nil, "invalid_properties", "q", "   ", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "invalid_properties", "q", "--json", "--", "   ")
		})
	}
}
