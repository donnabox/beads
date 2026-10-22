//go:build cgo

package main

import (
	"os"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
)

func TestGraphPreviewCreateInitialStatusWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			initArgs := []string{"init", "--graph-mode", "link", "--prefix", "status", "--scope-url", "https://example.invalid/create-status/", "--skip-hooks", "--skip-agents", "--non-interactive", "--json"}
			if engine == "server" {
				port := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
				if port == "" {
					t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for ordinary shared-server qualification")
				}
				initArgs = append(initArgs, "--server", "--external", "--server-host", "127.0.0.1", "--server-port", port, "--server-user", "root")
			}
			graphPolicyCLI(t, bd, work, home, nil, "", initArgs...)
			for _, tc := range []struct {
				name   string
				status types.Status
				flag   string
			}{
				{"default", types.StatusOpen, ""},
				{"empty", types.StatusOpen, "--status="},
				{"closed", types.StatusClosed, "--status=closed"},
				{"deferred", types.StatusDeferred, "--status=deferred"},
				{"blocked", types.StatusBlocked, "--status=blocked"},
			} {
				t.Run(tc.name, func(t *testing.T) {
					id := "status-" + tc.name
					args := []string{"create", "Status " + tc.name, "--id", id, "--notes=initial", "--json"}
					if tc.flag != "" {
						args = append(args, tc.flag)
					}
					created := graphMixedResult[graphstore.IssueRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", args...))
					if created.Properties.ID != id || created.Properties.Status != tc.status || (created.Properties.ClosedAt != nil) != (tc.status == types.StatusClosed) || created.Properties.Notes != "initial" {
						t.Fatalf("initial status or native Issue fields changed: %+v", created.Properties)
					}
					shown := graphMixedResult[graphstore.IssueRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", "show", id, "--format", "graph-json", "--json"))
					if shown.Revision != created.Revision || shown.Properties.Status != tc.status || (shown.Properties.ClosedAt != nil) != (tc.status == types.StatusClosed) {
						t.Fatalf("retained create status differs from current: %+v", shown.Properties)
					}
				})
			}
			child := graphMixedResult[graphstore.IssueRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", "create", "Status child", "--parent", "status-default", "--status=deferred", "--json"))
			if child.Properties.ID != "status-default.1" || child.Properties.Status != types.StatusDeferred {
				t.Fatalf("child create lost native dotted ID or initial status: %+v", child.Properties)
			}
			deferred := graphMixedResult[graphstore.IssueRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", "create", "Deferred by date", "--id", "status-defer-date", "--defer=+24h", "--json"))
			if deferred.Properties.Status != types.StatusDeferred || deferred.Properties.DeferUntil == nil || !deferred.Properties.DeferUntil.After(time.Now()) {
				t.Fatalf("create --defer did not retain a future deferred Issue: %+v", deferred.Properties)
			}
			shownDeferred := graphMixedResult[graphstore.IssueRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", "show", "status-defer-date", "--format", "graph-json", "--json"))
			if shownDeferred.Revision != deferred.Revision || shownDeferred.Properties.DeferUntil == nil || !shownDeferred.Properties.DeferUntil.Equal(*deferred.Properties.DeferUntil) {
				t.Fatalf("defer timestamp changed between creation and read: %+v", shownDeferred.Properties)
			}
			blockedWithDefer := graphMixedResult[graphstore.IssueRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", "create", "Explicit status wins", "--id", "status-defer-blocked", "--status=blocked", "--defer=+24h", "--json"))
			if blockedWithDefer.Properties.Status != types.StatusBlocked || blockedWithDefer.Properties.DeferUntil == nil {
				t.Fatalf("explicit status did not win over --defer: %+v", blockedWithDefer.Properties)
			}
			graphPolicyCLI(t, bd, work, home, nil, "invalid_properties", "create", "Invalid", "--id", "status-invalid", "--status=not-configured", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "not_found", "show", "status-invalid", "--json")
		})
	}
}
