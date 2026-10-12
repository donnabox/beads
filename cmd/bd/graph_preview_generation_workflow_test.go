//go:build cgo

package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/steveyegge/beads/internal/configfile"
	"github.com/steveyegge/beads/internal/storage/doltutil"
	"github.com/steveyegge/beads/internal/storage/embeddeddolt"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
)

// The two refusals that existed before generations were admitted by one
// predicate. A marker that is merely unsupported keeps these texts; only a
// well-formed newer generation earns a message of its own.
const (
	graphGenerationFirstGateRefusal  = "graph_mode marker and metadata disagree; automatic recovery is not supported"
	graphGenerationSecondGateRefusal = "graph_mode link metadata is missing, incomplete, or unsupported; no database was opened"
)

// graphGenerationRun runs bd in a disposable environment and returns its stdout,
// stderr and exit status. The environment is the whitelist graphPolicyCLI uses,
// so an operator's workspace, routing, credentials and telemetry never enter it;
// unlike graphPolicyCLI it hands back the message and status of a refusal.
func graphGenerationRun(t *testing.T, bd, work, home string, args ...string) (stdout, stderr string, exit int) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bd, args...)
	cmd.Dir = work
	cmd.Env = []string{
		"PATH=" + os.Getenv("PATH"), "HOME=" + home,
		"TMPDIR=" + os.TempDir(), "TMP=" + os.TempDir(), "TEMP=" + os.TempDir(),
		"XDG_CONFIG_HOME=" + filepath.Join(home, ".config"),
		"XDG_CACHE_HOME=" + filepath.Join(home, ".cache"),
		"GIT_CONFIG_GLOBAL=" + filepath.Join(home, "missing-gitconfig"), "GIT_CONFIG_NOSYSTEM=1",
		"BD_DISABLE_METRICS=1", "BD_DISABLE_EVENT_FLUSH=1", "BD_NON_INTERACTIVE=1",
		"BEADS_DOLT_AUTO_START=0", "DOLT_METRICS_DISABLED=1", "NO_COLOR=1",
	}
	if developerDir := os.Getenv("DEVELOPER_DIR"); developerDir != "" {
		cmd.Env = append(cmd.Env, "DEVELOPER_DIR="+developerDir)
	}
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	err := cmd.Run()
	if ctx.Err() != nil {
		t.Fatalf("CLI did not complete within deadline: %v\n%s", ctx.Err(), errOut.String())
	}
	var exitErr *exec.ExitError
	if err != nil && !errors.As(err, &exitErr) {
		t.Fatalf("run %v: %v", args, err)
	}
	if exitErr != nil {
		exit = exitErr.ExitCode()
	}
	return out.String(), errOut.String(), exit
}

// graphGenerationRefusal runs a command that admission must refuse and returns
// the refusal message, after checking it is the typed, non-retryable
// graph_not_initialized refusal with exit status 5 and no output on stdout.
func graphGenerationRefusal(t *testing.T, bd, work, home string, args ...string) string {
	t.Helper()
	stdout, stderr, exit := graphGenerationRun(t, bd, work, home, append(args, "--json")...)
	if exit != 5 || stdout != "" {
		t.Fatalf("%v: want an exit 5 refusal with empty stdout, got exit %d stdout=%q stderr=%q", args, exit, stdout, stderr)
	}
	var diagnostic struct {
		Code      string `json:"code"`
		Message   string `json:"message"`
		Retryable bool   `json:"retryable"`
	}
	if err := json.Unmarshal([]byte(stderr), &diagnostic); err != nil {
		t.Fatalf("%v: expected a typed refusal: %v\n%s", args, err, stderr)
	}
	if diagnostic.Code != "graph_not_initialized" || diagnostic.Retryable {
		t.Fatalf("%v: want a non-retryable graph_not_initialized refusal, got %+v", args, diagnostic)
	}
	return diagnostic.Message
}

// graphGenerationExec runs one statement against the workspace's embedded
// database and closes it again, for the one edit the CLI cannot make. The
// database is closed before returning because no bd process can open it while
// this process holds it.
func graphGenerationExec(t *testing.T, cfg *configfile.Config, statement string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
	defer cancel()
	db, closeDB, err := embeddeddolt.OpenSQL(ctx, filepath.Join(cfg.GraphWorkspace, "embeddeddolt"), cfg.DoltDatabase, "main")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := closeDB(); err != nil {
			t.Error(err)
		}
	}()
	if _, err := db.ExecContext(ctx, statement); err != nil {
		t.Fatal(err)
	}
}

// graphGenerationQueryInt reads one integer from the workspace's embedded
// database, closing it again before returning.
func graphGenerationQueryInt(t *testing.T, cfg *configfile.Config, query string) int {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
	defer cancel()
	db, closeDB, err := embeddeddolt.OpenSQL(ctx, filepath.Join(cfg.GraphWorkspace, "embeddeddolt"), cfg.DoltDatabase, "main")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := closeDB(); err != nil {
			t.Error(err)
		}
	}()
	var value int
	if err := db.QueryRowContext(ctx, query).Scan(&value); err != nil {
		t.Fatal(err)
	}
	return value
}

// This workflow follows one embedded workspace through the generations of its
// format marker: what a fresh init writes, which generations admission accepts and
// reads, which it refuses and with what message, and that a newer generation is
// refused before the store is touched. The marker is engine independent (one
// writer, one admission function), so there is no server variant.
func TestGraphPreviewMarkerGenerationWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	work, home := t.TempDir(), t.TempDir()
	const scope = "https://example.invalid/generation/"
	graphPolicyCLI(t, bd, work, home, nil, "", "init", "--graph-mode", "link", "--scope-url", scope, "--skip-hooks", "--skip-agents", "--non-interactive")
	beadsDir := filepath.Join(work, ".beads")
	markerPath := filepath.Join(beadsDir, graphPreviewMarker)
	dataDir := filepath.Join(beadsDir, "embeddeddolt")

	loadConfig := func(t *testing.T) *configfile.Config {
		t.Helper()
		cfg, err := configfile.LoadForDiscovery(beadsDir)
		if err != nil || cfg == nil {
			t.Fatalf("read initialized binding: %v", err)
		}
		return cfg
	}
	// readTypes runs a read command against the workspace under the given marker
	// and returns how many Bead and Link Types it reports installed.
	readTypes := func(t *testing.T, marker string) (beadTypes, linkTypes int) {
		t.Helper()
		writeFile(t, markerPath, []byte(marker))
		var installed struct {
			Result struct {
				BeadTypes []json.RawMessage `json:"beadTypes"`
				LinkTypes []json.RawMessage `json:"linkTypes"`
			} `json:"result"`
		}
		out := graphPolicyCLI(t, bd, work, home, nil, "", "types", "--json")
		if err := json.Unmarshal([]byte(out), &installed); err != nil {
			t.Fatalf("types under marker %q: %v\n%s", marker, err, out)
		}
		return len(installed.Result.BeadTypes), len(installed.Result.LinkTypes)
	}
	// refuse asserts that a command is refused at admission under the marker, with
	// exactly the wanted message, and that the refusal changed nothing on disk.
	refuse := func(t *testing.T, marker, want string, args ...string) {
		t.Helper()
		writeFile(t, markerPath, []byte(marker))
		before := legacyUpgradeTreeDigest(t, beadsDir)
		if got := graphGenerationRefusal(t, bd, work, home, args...); got != want {
			t.Fatalf("marker %q: refusal message\n got: %s\nwant: %s", marker, got, want)
		}
		if after := legacyUpgradeTreeDigest(t, beadsDir); after != before {
			t.Fatalf("marker %q: refused command changed the workspace", marker)
		}
	}
	newerRefusal := func(number string) string {
		return "graph_mode workspace format link-preview-v" + number + " is newer than this bd supports (link-preview-v7); upgrade bd; no database was opened"
	}

	t.Run("fresh init writes the current generation", func(t *testing.T) {
		marker, err := os.ReadFile(markerPath)
		if err != nil || string(marker) != "link-preview-v7\n" {
			t.Fatalf("fresh workspace marker = %q (%v), want link-preview-v7 and a newline", marker, err)
		}
		// Common metadata changes the storage layout independently of the
		// installed Type generation.
		if cfg := loadConfig(t); cfg.GraphSchemaVersion != 7 {
			t.Fatalf("graph_schema_version = %d, want 7", cfg.GraphSchemaVersion)
		}
		if persisted := graphGenerationQueryInt(t, loadConfig(t), `SELECT schema_version FROM graph_preview_scope WHERE singleton = 1`); persisted != 7 {
			t.Fatalf("persisted schema_version = %d, want 7", persisted)
		}
	})

	t.Run("admits link-preview-v7 over eleven Types", func(t *testing.T) {
		if beads, links := readTypes(t, "link-preview-v7\n"); beads != 3 || links != 8 {
			t.Fatalf("installed Types = %d Bead, %d Link; want 3 and 8", beads, links)
		}
	})

	// Recreate the exact preceding six-Type installation. The new binary reads
	// it as v6 and v5 without inserting any v7 Type or changing old fingerprints.
	graphGenerationExec(t, loadConfig(t), `DELETE FROM graph_preview_types WHERE name IN ('issue-v3','blocks-v2','parent-child','duplicates','supersedes')`)
	t.Run("admits link-preview-v6 over six Types", func(t *testing.T) {
		if beads, links := readTypes(t, "link-preview-v6\n"); beads != 2 || links != 4 {
			t.Fatalf("installed Types = %d Bead, %d Link; want 2 and 4", beads, links)
		}
	})

	t.Run("admits link-preview-v5 over six Types", func(t *testing.T) {
		if beads, links := readTypes(t, "link-preview-v5\n"); beads != 2 || links != 4 {
			t.Fatalf("installed Types = %d Bead, %d Link; want 2 and 4", beads, links)
		}
	})
	t.Run("v7 marker does not adopt a v6 installation", func(t *testing.T) {
		writeFile(t, markerPath, []byte("link-preview-v7\n"))
		before := legacyUpgradeTreeDigest(t, beadsDir)
		stdout, stderr, exit := graphGenerationRun(t, bd, work, home, "types", "--json")
		if exit != 5 || stdout != "" || !strings.Contains(stderr, `"code":"invalid_store"`) ||
			!strings.Contains(stderr, "marker and installed Type generation differ") {
			t.Fatalf("v7 marker over six Types: exit=%d stdout=%q stderr=%q", exit, stdout, stderr)
		}
		if after := legacyUpgradeTreeDigest(t, beadsDir); after != before {
			t.Fatal("generation mismatch read changed the workspace")
		}
	})

	t.Run("refuses a newer generation at the first gate", func(t *testing.T) {
		// 10 sorts before 7 as text and 18446744073709551616 overflows 64 bits;
		// both are still newer, and the write is refused the same way as a read.
		for _, number := range []string{"8", "10", "18446744073709551616"} {
			refuse(t, "link-preview-v"+number+"\n", newerRefusal(number), "types")
		}
		refuse(t, "link-preview-v8\n", newerRefusal("8"), "remember", "Refused write", "--id", "beads/refused", "--title", "Refused")
	})

	t.Run("keeps the existing refusal for every other marker", func(t *testing.T) {
		for _, tc := range []struct{ name, marker string }{
			{"older generation", "link-preview-v4\n"},
			{"empty file", ""},
			{"current generation without a newline", "link-preview-v7"},
			{"newer generation without a newline", "link-preview-v8"},
			{"leading zero", "link-preview-v06\n"},
			{"trailing text on the current generation", "link-preview-v7x\n"},
			{"trailing text on a newer generation", "link-preview-v8x\n"},
			{"no number", "link-preview-v"},
			{"no number with a newline", "link-preview-v\n"},
			{"newer generation with a second newline", "link-preview-v8\n\n"},
		} {
			t.Run(tc.name, func(t *testing.T) {
				refuse(t, tc.marker, graphGenerationFirstGateRefusal, "types")
			})
		}
	})

	t.Run("a missing marker is refused at the second gate", func(t *testing.T) {
		if err := os.Remove(markerPath); err != nil {
			t.Fatal(err)
		}
		before := legacyUpgradeTreeDigest(t, beadsDir)
		if got := graphGenerationRefusal(t, bd, work, home, "types"); got != graphGenerationSecondGateRefusal {
			t.Fatalf("missing marker: refusal message\n got: %s\nwant: %s", got, graphGenerationSecondGateRefusal)
		}
		if after := legacyUpgradeTreeDigest(t, beadsDir); after != before {
			t.Fatal("refused command changed the workspace")
		}
	})

	t.Run("generation-specific routes refuse before the store is opened", func(t *testing.T) {
		away := filepath.Join(beadsDir, "embeddeddolt.away")
		if err := os.Rename(dataDir, away); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := os.Rename(away, dataDir); err != nil {
				t.Error(err)
			}
		})
		// Control: with the store gone, a supported marker fails for some other
		// reason, so the refusal below cannot come from looking at the store.
		writeFile(t, markerPath, []byte("link-preview-v6\n"))
		stdout, stderr, exit := graphGenerationRun(t, bd, work, home, "types", "--json")
		if exit == 0 || stdout != "" {
			t.Fatalf("a missing store was not noticed: exit %d stdout=%q stderr=%q", exit, stdout, stderr)
		}
		for _, marker := range []string{graphGenerationFirstGateRefusal, graphGenerationSecondGateRefusal, "newer than this bd supports"} {
			if strings.Contains(stderr, marker) {
				t.Fatalf("a supported marker was refused as a marker problem: %s", stderr)
			}
		}
		refuse(t, "link-preview-v8\n", newerRefusal("8"), "types")
		for _, marker := range []string{"link-preview-v5\n", "link-preview-v6\n"} {
			writeFile(t, markerPath, []byte(marker))
			for _, args := range [][]string{
				{"children", "beads/parent", "--json"},
				{"epic", "status", "--json"},
				{"epic", "close-eligible", "--dry-run", "--json"},
				{"epic", "close-eligible", "--json"},
			} {
				graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", args...)
			}
		}
	})

	t.Run("admits link-preview-v5 over the legacy four Types", func(t *testing.T) {
		// Removing the two unused example rows recreates the original four-Type
		// installation, which this bd still opens and never upgrades.
		graphGenerationExec(t, loadConfig(t), `DELETE FROM graph_preview_types WHERE name IN ('example-follows','example-cites')`)
		if beads, links := readTypes(t, "link-preview-v5\n"); beads != 2 || links != 2 {
			t.Fatalf("installed Types = %d Bead, %d Link; want the original 2 and 2", beads, links)
		}
	})
}

// Inspect the same native rows the installed CLI must have written. Close the
// connection before the next CLI process, particularly for embedded workspaces.
type graphInstalledRelationshipState struct {
	NativeType, SourceID, TargetID, LinkType              string
	NativeVersions, GraphVersions, LinkVersions, PairRows int
}

func graphInstalledRelationship(t *testing.T, work, linkPath, sourceID, targetID string) graphInstalledRelationshipState {
	t.Helper()
	cfg, err := configfile.LoadForDiscovery(filepath.Join(work, ".beads"))
	if err != nil || cfg == nil {
		t.Fatalf("load graph workspace binding: %v", err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
	defer cancel()
	var db *sql.DB
	var closeDB func() error
	if cfg.DoltMode == configfile.DoltModeServer {
		user := cfg.DoltServerUser
		if user == "" {
			user = configfile.DefaultDoltServerUser
		}
		db, err = sql.Open("mysql", doltutil.ServerDSN{
			Host: cfg.DoltServerHost, Port: cfg.DoltServerPort, User: user,
			Database: cfg.DoltDatabase, Socket: cfg.DoltServerSocket, TLS: cfg.DoltServerTLS,
		}.String())
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
	state := graphInstalledRelationshipState{}
	query := `SELECT d.type,d.issue_id,d.depends_on_issue_id,c.type_url FROM graph_preview_catalog c
 JOIN dependencies d ON d.id=c.backing_key WHERE c.path=? AND c.backing='dependency' AND c.allocation_state='live'`
	err = db.QueryRowContext(ctx, query, linkPath).Scan(&state.NativeType, &state.SourceID, &state.TargetID, &state.LinkType)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		t.Fatal(err)
	}
	for _, item := range []struct {
		query string
		args  []any
		out   *int
	}{
		{`SELECT COUNT(*) FROM issue_versions WHERE issue_id=?`, []any{sourceID}, &state.NativeVersions},
		{`SELECT COUNT(*) FROM graph_preview_issue_versions WHERE issue_id=?`, []any{sourceID}, &state.GraphVersions},
		{`SELECT COUNT(*) FROM graph_preview_versions WHERE path=?`, []any{linkPath}, &state.LinkVersions},
		{`SELECT COUNT(*) FROM dependencies WHERE issue_id=? AND depends_on_issue_id=?`, []any{sourceID, targetID}, &state.PairRows},
	} {
		if err := db.QueryRowContext(ctx, item.query, item.args...).Scan(item.out); err != nil {
			t.Fatal(err)
		}
	}
	return state
}

func TestGraphPreviewInstalledV7TypedRelationshipsAndLegacyRefusal(t *testing.T) {
	bd := buildBDUnderTest(t)
	const scope = "https://example.invalid/typed-relationships/"
	for _, engine := range []string{"embedded", "server"} {
		t.Run("v7-"+engine, func(t *testing.T) {
			work, home := graphListWorkspace(t, bd, engine, scope)
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			refuse := func(code string, args ...string) {
				t.Helper()
				graphPolicyCLI(t, bd, work, home, nil, code, append(args, "--json")...)
			}
			for _, tc := range []struct {
				name    string
				kind    types.DependencyType
				viaLink bool
			}{
				{"dep-parent", types.DepParentChild, false},
				{"dep-duplicate", types.DepDuplicates, false},
				{"dep-supersede", types.DepSupersedes, false},
				{"link-duplicate", types.DepDuplicates, true},
			} {
				t.Run(tc.name, func(t *testing.T) {
					sourcePath, targetPath, linkPath := "beads/source-"+tc.name, "beads/target-"+tc.name, "links/"+tc.name
					source := graphMixedResult[graphstore.IssueRecord](t, call("create", "Source", "--id", sourcePath))
					target := graphMixedResult[graphstore.IssueRecord](t, call("create", "Target", "--id", targetPath))
					typeURL, _ := graphstore.RelationshipDependencyTypeURL(scope, tc.kind)
					addArgs := []string{"dep", "add", sourcePath, targetPath, "--type", string(tc.kind), "--id", linkPath}
					if tc.viaLink {
						addArgs = []string{"link", sourcePath, targetPath, "--link-type", typeURL, "--id", linkPath, "--if-source-revision", source.Revision}
					}
					added := graphMixedResult[graphstore.DependencyResult](t, call(addArgs...))
					if !added.Changed || added.Link.Type != typeURL || added.Link.Source != source.ID || added.Link.Target != target.ID || len(added.Source.Owned) != 1 {
						t.Fatalf("rendered native Link differs: %+v", added)
					}
					shown := graphMixedResult[graphstore.LinkRecord](t, call("show", linkPath))
					if shown.Type != typeURL || shown.Source != source.ID || shown.Target != target.ID {
						t.Fatalf("shown Link differs from native endpoints: %+v", shown)
					}
					state := graphInstalledRelationship(t, work, linkPath, source.Properties.ID, target.Properties.ID)
					if state.NativeType != string(tc.kind) || state.SourceID != source.Properties.ID || state.TargetID != target.Properties.ID || state.LinkType != typeURL ||
						state.NativeVersions != 2 || state.GraphVersions != 2 || state.LinkVersions != 1 || state.PairRows != 1 {
						t.Fatalf("native Link mapping or History after add: %+v", state)
					}
					beforeNoOp := call("show", sourcePath, "--format", "graph-json")
					if tc.viaLink {
						addArgs[len(addArgs)-1] = added.Source.Revision
					}
					repeated := graphMixedResult[graphstore.DependencyResult](t, call(addArgs...))
					if repeated.Changed || repeated.Link.ID != added.Link.ID || repeated.Source.Revision != added.Source.Revision {
						t.Fatalf("same pair was not an exact no-op: %+v", repeated)
					}
					if again := graphInstalledRelationship(t, work, linkPath, source.Properties.ID, target.Properties.ID); again != state {
						t.Fatalf("no-op changed native or retained state: before=%+v after=%+v", state, again)
					}
					if afterNoOp := call("show", sourcePath, "--format", "graph-json"); afterNoOp != beforeNoOp {
						t.Fatal("same-pair no-op changed the source record")
					}
					refuse("revision_conflict", "unlink", linkPath, "--if-revision", added.Link.Revision, "--if-source-revision", source.Revision)
					removed := graphMixedResult[graphstore.LinkDeleteResult](t, call("unlink", added.Link.ID, "--if-revision", added.Link.Revision, "--if-source-revision", added.Source.Revision))
					if !removed.Changed || removed.Link.Type != typeURL || removed.Link.State != "deleted" {
						t.Fatalf("canonical unlink did not tombstone typed Link: %+v", removed)
					}
					refuse("gone", "show", linkPath)
					if owner := graphMixedResult[graphstore.IssueRecord](t, call("show", sourcePath, "--format", "graph-json")); len(owner.Owned) != 0 {
						t.Fatalf("unlinked native Link remains owned: %+v", owner.Owned)
					}
					state = graphInstalledRelationship(t, work, linkPath, source.Properties.ID, target.Properties.ID)
					if state.NativeType != "" || state.PairRows != 0 || state.NativeVersions != 3 || state.GraphVersions != 3 || state.LinkVersions != 2 {
						t.Fatalf("native unlink or retained History differs: %+v", state)
					}
				})
			}
		})
	}

	for _, generation := range []string{"v5", "v6"} {
		t.Run(generation+"-pinned", func(t *testing.T) {
			work, home := graphListWorkspace(t, bd, "embedded", scope)
			beadsDir := filepath.Join(work, ".beads")
			cfg, err := configfile.LoadForDiscovery(beadsDir)
			if err != nil || cfg == nil {
				t.Fatalf("load graph workspace binding: %v", err)
			}
			graphGenerationExec(t, cfg, `DELETE FROM graph_preview_types WHERE name IN ('issue-v3','blocks-v2','parent-child','duplicates','supersedes')`)
			writeFile(t, filepath.Join(beadsDir, graphPreviewMarker), []byte("link-preview-"+generation+"\n"))
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			call("create", "Source", "--id", "beads/source")
			call("create", "Target", "--id", "beads/target")
			before := call("show", "beads/source", "--format", "graph-json")
			for _, args := range [][]string{
				{"children", "beads/source", "--json"},
				{"epic", "status", "--json"},
				{"epic", "close-eligible", "--dry-run", "--json"},
				{"epic", "close-eligible", "--json"},
			} {
				graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", args...)
			}
			for _, kind := range []types.DependencyType{types.DepParentChild, types.DepDuplicates, types.DepSupersedes} {
				graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "dep", "add", "beads/source", "beads/target", "--type", string(kind), "--json")
				typeURL, _ := graphstore.RelationshipDependencyTypeURL(scope, kind)
				graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "link", "beads/source", "beads/target", "--link-type", typeURL, "--unconditional-source", "--json")
			}
			if after := call("show", "beads/source", "--format", "graph-json"); after != before || graphGenerationQueryInt(t, cfg, `SELECT COUNT(*) FROM dependencies`) != 0 ||
				graphGenerationQueryInt(t, cfg, `SELECT COUNT(*) FROM graph_preview_types`) != 6 {
				t.Fatal("pinned workspace changed after typed relationship refusals")
			}
			blocking := graphMixedResult[graphstore.DependencyResult](t, call("dep", "add", "beads/source", "beads/target"))
			if blocking.Link.Type != graphstore.DependencyTypeURL(scope) {
				t.Fatalf("pinned workspace lost its original blocking Type: %+v", blocking.Link)
			}
		})
	}
}
