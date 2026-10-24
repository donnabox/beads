//go:build cgo

package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/configfile"
	"github.com/steveyegge/beads/internal/storage/doltutil"
	"github.com/steveyegge/beads/internal/storage/embeddeddolt"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// Capture every authoritative/retained write table between installed commands,
// including the writer fence, tombstones, native history and journal counters.
func compatibilityDiagnosticState(t *testing.T, work string) map[string][]string {
	t.Helper()
	cfg, err := configfile.LoadForDiscovery(filepath.Join(work, ".beads"))
	if err != nil || cfg == nil {
		t.Fatalf("load workspace: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	var db *sql.DB
	var closeDB func() error
	if cfg.DoltMode == configfile.DoltModeServer {
		db, err = sql.Open("mysql", doltutil.ServerDSN{Host: cfg.DoltServerHost, Port: cfg.DoltServerPort, User: cfg.DoltServerUser, Database: cfg.DoltDatabase, Socket: cfg.DoltServerSocket, TLS: cfg.DoltServerTLS}.String())
		if err == nil {
			closeDB = db.Close
		}
	} else {
		db, closeDB, err = embeddeddolt.OpenSQL(ctx, filepath.Join(cfg.GraphWorkspace, "embeddeddolt"), cfg.DoltDatabase, "main")
	}
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := closeDB(); err != nil {
			t.Error(err)
		}
	}()
	state := map[string][]string{}
	for _, table := range []string{"issues", "dependencies", "events", "bd_events_journal", "bd_events_seq", "leases", "issue_versions", "store_epoch", "local_metadata", "graph_preview_scope", "graph_preview_types", "graph_preview_catalog", "graph_preview_versions", "graph_preview_issue_versions", "graph_preview_links", "graph_preview_payloads"} {
		rows, err := db.QueryContext(ctx, "SELECT * FROM "+table)
		if err != nil {
			t.Fatal(err)
		}
		columns, err := rows.Columns()
		if err != nil {
			_ = rows.Close()
			t.Fatal(err)
		}
		values := []string{}
		for rows.Next() {
			fields := make([]any, len(columns))
			pointers := make([]any, len(columns))
			for i := range fields {
				pointers[i] = &fields[i]
			}
			if err := rows.Scan(pointers...); err != nil {
				_ = rows.Close()
				t.Fatal(err)
			}
			raw, err := json.Marshal(fields)
			if err != nil {
				_ = rows.Close()
				t.Fatal(err)
			}
			values = append(values, string(raw))
		}
		if err := errors.Join(rows.Err(), rows.Close()); err != nil {
			t.Fatal(err)
		}
		sort.Strings(values)
		state[table] = values
	}
	return state
}

func TestGraphPreviewRememberCompatibilityDiagnostics(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			const scope = "https://example.invalid/compatibility-diagnostics/"
			work, home := graphListWorkspace(t, bd, engine, scope)
			created := graphMixedResult[graphstore.Record](t, graphPolicyCLI(t, bd, work, home, nil, "", "remember", "Exact body", "--key", "TWO", "--json"))
			if created.ID != scope+"TWO" {
				t.Fatal("valid key changed", created.ID)
			}
			for _, flag := range []string{"--key", "--id"} {
				for _, tc := range []struct{ key, reason string }{{"..", "dot segment"}, {"a//b", "nonempty"}} {
					for _, guarded := range []bool{false, true} {
						before := compatibilityDiagnosticState(t, work)
						args := []string{"remember", flag, tc.key, "--title", "Changed title", "--json"}
						if guarded {
							args = append([]string{"remember", "Guarded changed body"}, args[1:]...)
							args = append(args, "--if-revision", created.Revision)
						}
						out, stderr, exit := graphGenerationRun(t, bd, work, home, args...)
						var diagnostic struct {
							Code, Message string
							Retryable     bool
						}
						if exit != 2 || out != "" || json.Unmarshal([]byte(stderr), &diagnostic) != nil || diagnostic.Code != "invalid_properties" || diagnostic.Retryable {
							t.Fatalf("refusal: exit%d out%q stderr%s", exit, out, stderr)
						}
						for _, want := range []string{strconv.Quote(tc.key), tc.reason, "choose an explicit compatible replacement ID"} {
							if !strings.Contains(diagnostic.Message, want) {
								t.Errorf("%v diagnostic %q omits %q", args, diagnostic.Message, want)
							}
						}
						if after := compatibilityDiagnosticState(t, work); !reflect.DeepEqual(before, after) {
							t.Fatal("rejected identity changed authoritative or retained state")
						}
					}
				}
			}
			updated := graphMixedResult[graphstore.MemoryMutationResult](t, graphPolicyCLI(t, bd, work, home, nil, "", "remember", "--id", "TWO", "--title", "Independent title", "--if-revision", created.Revision, "--json"))
			if updated.Memory.ID != created.ID || updated.Memory.Properties.Body != created.Properties.Body || updated.Memory.Properties.Title != "Independent title" {
				t.Fatal("valid exact ID update changed identity/body", updated)
			}
		})
	}
}
