//go:build cgo

package main

import (
	"testing"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func TestGraphPreviewReadyIgnoresInformationalLinks(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			const scope = "https://example.invalid/ready-nondependency/"
			work, home := graphListWorkspace(t, bd, engine, scope)
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			issue := graphMixedResult[graphstore.IssueRecord](t, call("create", "Work", "--id", "beads/work", "--labels", "selected"))
			call("create", "--type", "types/memory", "--id", "context", "--properties", `{"title":"Context","body":"Informational context."}`)
			ready := func() []graphstore.IssueRecord {
				t.Helper()
				return graphMixedResult[[]graphstore.IssueRecord](t, call("ready", "--label", "selected"))
			}
			if got := ready(); len(got) != 1 || got[0].ID != issue.ID {
				t.Fatalf("fixture Issue is not ready before informational Links: %+v", got)
			}
			call("link", "create", "work", "context", "--type", "types/references", "--id", "links/outgoing-context")
			call("link", "create", "context", "work", "--type", "types/references", "--id", "links/incoming-context")
			links := graphMixedResult[[]graphstore.LinkRecord](t, call("link", "list", "--incident-to", "work"))
			if len(links) != 2 {
				t.Fatalf("fixture did not create both informational Links: %+v", links)
			}
			if got := ready(); len(got) != 1 || got[0].ID != issue.ID {
				t.Fatalf("informational Links changed dependency readiness: %+v", got)
			}
			claimed := graphMixedResult[[]graphstore.IssueRecord](t, call("ready", "--claim", "--label", "selected", "--actor", "ready-agent"))
			if len(claimed) != 1 || claimed[0].ID != issue.ID {
				t.Fatalf("informational Links prevented an otherwise-ready claim: %+v", claimed)
			}
		})
	}
}
