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
	marker, err := os.ReadFile(markerPath)
	if err != nil || string(marker) != graphPreviewGeneration {
		t.Fatalf("fresh marker = %q (%v), want %q", marker, err, graphPreviewGeneration)
	}
	cfg, err := configfile.LoadForDiscovery(beadsDir)
	if err != nil || cfg == nil {
		t.Fatalf("read initialized binding: %v", err)
	}
	if cfg.GraphSchemaVersion != 11 {
		t.Fatalf("graph_schema_version = %d, want 11", cfg.GraphSchemaVersion)
	}
	if persisted := graphGenerationQueryInt(t, cfg, `SELECT schema_version FROM graph_preview_scope WHERE singleton = 1`); persisted != 11 {
		t.Fatalf("persisted schema_version = %d, want 11", persisted)
	}
	// Availability changes discovery, not the immutable installed catalog.
	if registered := graphGenerationQueryInt(t, cfg, `SELECT COUNT(*) FROM graph_preview_types`); registered != 50 {
		t.Fatalf("fresh registered Types = %d; want all 50 definitions", registered)
	}
	var installed struct {
		Result struct {
			BeadTypes []json.RawMessage `json:"beadTypes"`
			LinkTypes []json.RawMessage `json:"linkTypes"`
		} `json:"result"`
	}
	out := graphPolicyCLI(t, bd, work, home, nil, "", "types", "--json")
	if err := json.Unmarshal([]byte(out), &installed); err != nil {
		t.Fatalf("read fresh Type inventory: %v\n%s", err, out)
	}
	if beads, links := len(installed.Result.BeadTypes), len(installed.Result.LinkTypes); beads != 22 || links != 25 {
		t.Fatalf("fresh available Types = %d Bead, %d Link; want 20 and 25", beads, links)
	}

	refuse := func(marker, want string) {
		t.Helper()
		writeFile(t, markerPath, []byte(marker))
		before := legacyUpgradeTreeDigest(t, beadsDir)
		if got := graphGenerationRefusal(t, bd, work, home, "types"); got != want {
			t.Fatalf("marker %q: refusal = %q, want %q", marker, got, want)
		}
		if after := legacyUpgradeTreeDigest(t, beadsDir); after != before {
			t.Fatalf("marker %q: refused read changed workspace", marker)
		}
	}
	for _, old := range []string{"link-preview-v5\n", "link-preview-v6\n", "link-preview-v7\n", "link-preview-v8\n", "link-preview-v9\n", "link-preview-v10\n", "link-preview-v11\n", "link-preview-v12\n"} {
		refuse(old, graphGenerationFirstGateRefusal)
	}
	refuse("link-preview-v14\n", "graph_mode workspace format link-preview-v14 is newer than this bd supports (link-preview-v13); upgrade bd; no database was opened")
	refuse("link-preview-v9", graphGenerationFirstGateRefusal)
	refuse("link-preview-v9x\n", graphGenerationFirstGateRefusal)

	if err := os.Remove(markerPath); err != nil {
		t.Fatal(err)
	}
	before := legacyUpgradeTreeDigest(t, beadsDir)
	if got := graphGenerationRefusal(t, bd, work, home, "types"); got != graphGenerationSecondGateRefusal {
		t.Fatalf("missing marker: refusal = %q, want %q", got, graphGenerationSecondGateRefusal)
	}
	if after := legacyUpgradeTreeDigest(t, beadsDir); after != before {
		t.Fatal("missing marker read changed workspace")
	}

	writeFile(t, markerPath, []byte(graphPreviewGeneration))
	graphGenerationExec(t, cfg, `DELETE FROM graph_preview_types WHERE name='dependency'`)
	stdout, stderr, exit := graphGenerationRun(t, bd, work, home, "types", "--json")
	if exit != 5 || stdout != "" || !strings.Contains(stderr, `"code":"invalid_store"`) {
		t.Fatalf("partial installation admitted: exit=%d stdout=%q stderr=%q", exit, stdout, stderr)
	}
}

// Inspect the same native rows the installed CLI must have written. Close the
// connection before the next CLI process, particularly for embedded workspaces.
type graphInstalledRelationshipState struct {
	NativeType, SourceID, TargetID, LinkType              string
	NativeVersions, GraphVersions, LinkVersions, PairRows int
}

func graphInstalledRelationship(t *testing.T, work, linkPublicID, sourceID, targetID string) graphInstalledRelationshipState {
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
	localID := strings.TrimPrefix(linkPublicID, cfg.GraphScopeURL)
	if localID == linkPublicID {
		t.Fatalf("Link %q is not in Scope %q", linkPublicID, cfg.GraphScopeURL)
	}
	var linkPath string
	if err := db.QueryRowContext(ctx, `SELECT path FROM graph_preview_catalog WHERE public_id=?`, localID).Scan(&linkPath); err != nil {
		t.Fatalf("resolve allocated Link %q: %v", linkPublicID, err)
	}
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

func TestGraphPreviewInstalledFreshTypedRelationships(t *testing.T) {
	bd := buildBDUnderTest(t)
	const scope = "https://example.invalid/typed-relationships/"
	for _, engine := range []string{"embedded", "server"} {
		t.Run("v11-"+engine, func(t *testing.T) {
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
				{"dep-custom", "tracks", false},
				{"link-duplicate", types.DepDuplicates, true},
				{"link-custom", "tracks", true},
			} {
				t.Run(tc.name, func(t *testing.T) {
					sourcePath, targetPath, linkPath := "beads/source-"+tc.name, "beads/target-"+tc.name, "links/"+tc.name
					source := graphMixedResult[graphstore.IssueRecord](t, call("create", "Source", "--graph-path", sourcePath))
					target := graphMixedResult[graphstore.IssueRecord](t, call("create", "Target", "--graph-path", targetPath))
					typeURL, _ := graphstore.RelationshipDependencyTypeURL(scope, tc.kind)
					addArgs := []string{"dep", "add", source.ID, target.ID, "--type", string(tc.kind), "--id", linkPath}
					if tc.viaLink {
						addArgs = []string{"link", "create", "--from", source.ID, "--to", target.ID, "--type", string(tc.kind), "--id", linkPath, "--if-source-revision", source.Revision}
					}
					added := graphMixedResult[graphstore.DependencyResult](t, call(addArgs...))
					if !added.Changed || added.Link.Type != typeURL || added.Link.Properties["dep_type"] != string(tc.kind) || added.Link.Source != source.ID || added.Link.Target != target.ID || len(added.Source.Owned) != 1 {
						t.Fatalf("rendered native Link differs: %+v", added)
					}
					shown := graphMixedResult[graphstore.LinkRecord](t, call("show", scope+linkPath))
					if shown.Type != typeURL || shown.Properties["dep_type"] != string(tc.kind) || shown.Source != source.ID || shown.Target != target.ID {
						t.Fatalf("shown Link differs from native endpoints: %+v", shown)
					}
					listed := graphMixedResult[[]graphstore.LinkRecord](t, call("link", "list", "--from", source.ID, "--type", "dependency", "--exact-type", typeURL))
					counted := graphMixedResult[map[string]int](t, call("link", "count", "--from", source.ID, "--type", "dependency", "--exact-type", typeURL))
					if len(listed) != 1 || listed[0].ID != added.Link.ID || counted["count"] != 1 {
						t.Fatalf("Dependency type list/count did not select %s: list=%+v count=%+v", tc.kind, listed, counted)
					}
					shorthandListed := graphMixedResult[[]graphstore.LinkRecord](t, call("link", "list", "--from", source.ID, "--type", string(tc.kind)))
					shorthandCounted := graphMixedResult[map[string]int](t, call("link", "count", "--from", source.ID, "--type", string(tc.kind)))
					if len(shorthandListed) != 1 || shorthandListed[0].ID != added.Link.ID || shorthandCounted["count"] != 1 {
						t.Fatalf("native Dependency shorthand did not select %s: list=%+v count=%+v", tc.kind, shorthandListed, shorthandCounted)
					}
					state := graphInstalledRelationship(t, work, added.Link.ID, source.Properties.ID, target.Properties.ID)
					if state.NativeType != string(tc.kind) || state.SourceID != source.Properties.ID || state.TargetID != target.Properties.ID || state.LinkType != typeURL ||
						state.NativeVersions != 2 || state.GraphVersions != 2 || state.LinkVersions != 1 || state.PairRows != 1 {
						t.Fatalf("native Link mapping or History after add: %+v", state)
					}
					beforeNoOp := call("show", source.ID, "--format", "graph-json")
					if tc.viaLink {
						addArgs[len(addArgs)-1] = added.Source.Revision
					}
					repeated := graphMixedResult[graphstore.DependencyResult](t, call(addArgs...))
					if repeated.Changed || repeated.Link.ID != added.Link.ID || repeated.Source.Revision != added.Source.Revision {
						t.Fatalf("same pair was not an exact no-op: %+v", repeated)
					}
					if again := graphInstalledRelationship(t, work, added.Link.ID, source.Properties.ID, target.Properties.ID); again != state {
						t.Fatalf("no-op changed native or retained state: before=%+v after=%+v", state, again)
					}
					if afterNoOp := call("show", source.ID, "--format", "graph-json"); afterNoOp != beforeNoOp {
						t.Fatal("same-pair no-op changed the source record")
					}
					refuse("revision_conflict", "link", "delete", "--force", linkPath, "--if-revision", "="+added.Link.Revision, "--if-source-revision", source.Revision)
					removed := graphMixedResult[graphstore.LinkDeleteResult](t, call("link", "delete", "--force", added.Link.ID, "--if-revision", "="+added.Link.Revision, "--if-source-revision", added.Source.Revision))
					if !removed.Changed || removed.Link.Type != typeURL || removed.Link.State != "deleted" {
						t.Fatalf("canonical unlink did not tombstone typed Link: %+v", removed)
					}
					refuse("gone", "show", added.Link.ID)
					if owner := graphMixedResult[graphstore.IssueRecord](t, call("show", source.ID, "--format", "graph-json")); len(owner.Owned) != 0 {
						t.Fatalf("unlinked native Link remains owned: %+v", owner.Owned)
					}
					state = graphInstalledRelationship(t, work, added.Link.ID, source.Properties.ID, target.Properties.ID)
					if state.NativeType != "" || state.PairRows != 0 || state.NativeVersions != 3 || state.GraphVersions != 3 || state.LinkVersions != 2 {
						t.Fatalf("native unlink or retained History differs: %+v", state)
					}
				})
			}
		})
	}

}
