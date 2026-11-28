//go:build cgo

package main

import (
	"os"
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
			graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "q", "Child", "--parent", created, "--json")
			graphPolicyCLI(t, bd, work, home, nil, "permission_denied", "q", "Forbidden", "--readonly", "--json")
			// The helper decodes stderr as one JSON object; a second plain or JSON
			// diagnostic would make this assertion fail.
			graphPolicyCLI(t, bd, work, home, nil, "invalid_properties", "q", "   ", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "invalid_properties", "q", "--json", "--", "   ")
		})
	}
}
