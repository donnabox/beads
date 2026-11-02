//go:build cgo

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/configfile"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// graphModeAdmissionFixture is a workspace whose metadata or format marker
// declares a format this build does not open.
type graphModeAdmissionFixture struct {
	name string
	// filename is the metadata file; "" lays down no metadata.
	filename string
	metadata string
	// embedded adds the embedded Dolt root of a graph database. Without the
	// guard, an ordinary bd opens that database as an Issue store and
	// migrates it.
	embedded bool
	// marker adds the graph-preview-format sentinel graph-preview builds
	// write before metadata.json.
	marker bool
	// refusal lists what the refusal must mention besides "this bd build
	// does not support": the declared format and the start of its hint.
	refusal []string
}

// write lays the fixture down under beadsDir and returns the metadata path,
// or "" when the fixture has no metadata.
func (f graphModeAdmissionFixture) write(t *testing.T, beadsDir string) string {
	t.Helper()
	if err := os.MkdirAll(beadsDir, 0o700); err != nil {
		t.Fatal(err)
	}
	path := ""
	if f.filename != "" {
		path = filepath.Join(beadsDir, f.filename)
		writeFile(t, path, []byte(f.metadata))
	}
	if f.embedded {
		if err := os.MkdirAll(filepath.Join(beadsDir, "embeddeddolt", "beads_graph_0123456789abcdef0123456789abcdef", ".dolt"), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	if f.marker {
		writeFile(t, filepath.Join(beadsDir, "graph-preview-format"), []byte("link-preview-v9\n"))
	}
	if err := os.Chmod(beadsDir, 0o700); err != nil {
		t.Fatal(err)
	}
	return path
}

func graphModeAdmissionFixtures() []graphModeAdmissionFixture {
	graphModeRefusal := func(mode string) []string {
		return []string{"graph_mode", mode, "use a bd build that supports graph_mode"}
	}
	var fixtures []graphModeAdmissionFixture
	fixtures = append(fixtures, graphModeAdmissionFixture{
		name: "metadata.json/link/embedded", filename: "metadata.json",
		metadata: graphLinkWorkspaceMetadata, embedded: true, marker: true,
		refusal: graphModeRefusal("link"),
	})
	for _, filename := range []string{"metadata.json", "config.json"} {
		for _, mode := range []string{"link", "future-format"} {
			fixtures = append(fixtures, graphModeAdmissionFixture{
				name: filename + "/" + mode + "/server", filename: filename,
				metadata: `{"backend":"dolt","dolt_mode":"server","graph_mode":"` + mode + `"}`,
				refusal:  graphModeRefusal(mode),
			})
		}
	}
	// Only the format marker declares these: graph init stopped before
	// metadata.json, metadata.json was lost, or a bd that predates
	// graph_mode rewrote it without the field.
	markerRefusal := []string{"graph-preview-format", "restore metadata.json"}
	fixtures = append(fixtures,
		graphModeAdmissionFixture{
			name: "marker/only", marker: true,
			refusal: markerRefusal,
		},
		graphModeAdmissionFixture{
			name: "marker/metadata-lost/embedded", embedded: true, marker: true,
			refusal: markerRefusal,
		},
		graphModeAdmissionFixture{
			name: "marker/metadata-stripped/embedded", filename: "metadata.json",
			metadata: graphStrippedWorkspaceMetadata, embedded: true, marker: true,
			refusal: markerRefusal,
		},
		graphModeAdmissionFixture{
			name: "marker/dependency/server", filename: "metadata.json",
			metadata: `{"backend":"dolt","dolt_mode":"server","graph_mode":"dependency"}`, marker: true,
			refusal: markerRefusal,
		},
	)
	return fixtures
}

// runGraphModeAdmissionBD runs bd in repoDir with a private HOME and no
// ambient BEADS_* selection, and returns its combined output and exit code.
func runGraphModeAdmissionBD(t *testing.T, bd, repoDir string, args ...string) (string, int) {
	t.Helper()
	return runGraphModeAdmissionBDEnv(t, bd, repoDir, nil, args...)
}

// runGraphModeAdmissionBDEnv is runGraphModeAdmissionBD run in dir with env
// added to bd's environment.
func runGraphModeAdmissionBDEnv(t *testing.T, bd, dir string, env []string, args ...string) (string, int) {
	t.Helper()
	// A refused run takes well under a second. An unrefused one (what these
	// tests catch) goes on to create or open a database, which under load
	// can approach 10s; the bound leaves room for it to fail on the
	// assertion that names the regression rather than on the timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bd, args...)
	cmd.Dir = dir
	cmd.Env = append(bdEnv(t.TempDir()),
		"BEADS_DOLT_SERVER_MODE=0",
		"BEADS_DOLT_SHARED_SERVER=0",
	)
	cmd.Env = append(cmd.Env, env...)
	var output, stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &output
	err := cmd.Run()
	if ctx.Err() != nil {
		t.Fatalf("bd %s did not finish promptly: %v\n%s", strings.Join(args, " "), ctx.Err(), output.String())
	}
	var exitErr *exec.ExitError
	switch {
	case err == nil:
		return stdout.String() + output.String(), 0
	case errors.As(err, &exitErr):
		if (exitErr.ExitCode() == 5 || exitErr.ExitCode() == ExitUnsupportedWorkspaceFormat) && stdout.Len() != 0 {
			t.Fatalf("workspace admission refusal wrote stdout: %q; stderr=%q", stdout.String(), output.String())
		}
		return stdout.String() + output.String(), exitErr.ExitCode()
	default:
		t.Fatalf("bd %s: %v", strings.Join(args, " "), err)
		return "", -1
	}
}

// requireGraphModeRefusal checks the fork's graph-aware early admission,
// independently of the ordinary Issue/foreign-workspace format refusal.
func requireGraphModeRefusal(t *testing.T, fixture graphModeAdmissionFixture, args []string, output string, code int) {
	t.Helper()
	var metadata struct {
		GraphMode string `json:"graph_mode"`
	}
	if fixture.metadata != "" {
		if err := json.Unmarshal([]byte(fixture.metadata), &metadata); err != nil {
			t.Fatal(err)
		}
	}
	wantCode, wantMessage := "graph_not_initialized", graphGenerationSecondGateRefusal
	switch {
	case metadata.GraphMode == "future-format":
		wantCode = "capability_unavailable"
		wantMessage = `unsupported workspace graph_mode "future-format"; use a compatible bd build without changing the workspace marker`
	case fixture.marker:
		wantMessage = graphGenerationFirstGateRefusal
	}
	requireExactGraphAdmissionFailure(t, args, output, code, wantCode, wantMessage)
}

func requireExactGraphAdmissionFailure(t *testing.T, args []string, output string, exit int, wantCode, wantMessage string) {
	t.Helper()
	if exit != 5 {
		t.Fatalf("bd %v exit=%d, want5\n%s", args, exit, output)
	}
	jsonMode := false
	for _, arg := range args {
		if arg == "--json" {
			jsonMode = true
		}
	}
	if !jsonMode {
		if strings.TrimSpace(output) != wantCode+": "+wantMessage {
			t.Fatalf("bd %v refusal=%q, want %q", args, output, wantCode+": "+wantMessage)
		}
		return
	}
	var got struct {
		Code      string `json:"code"`
		Message   string `json:"message"`
		Retryable *bool  `json:"retryable"`
	}
	dec := json.NewDecoder(strings.NewReader(output))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&got); err != nil {
		t.Fatalf("decode graph refusal: %v\n%s", err, output)
	}
	if got.Code != wantCode || got.Message != wantMessage || got.Retryable == nil || *got.Retryable {
		t.Fatalf("wrong graph refusal: %+v, want %s: %s", got, wantCode, wantMessage)
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		t.Fatalf("trailing graph diagnostic: %v", err)
	}
}

// Foreign Issue opens keep upstream's typed exit16 format guard. The direct
// graph front door above must not weaken this guard or pretend it shares G1/G2.
func requireLegacyGraphFormatRefusal(t *testing.T, fixture graphModeAdmissionFixture, beadsDir string, args []string, output string, code int) {
	t.Helper()
	if code != ExitUnsupportedWorkspaceFormat {
		t.Fatalf("bd %v exit=%d, want%d\n%s", args, code, ExitUnsupportedWorkspaceFormat, output)
	}
	var metadata struct {
		GraphMode string `json:"graph_mode"`
	}
	if fixture.metadata != "" {
		if err := json.Unmarshal([]byte(fixture.metadata), &metadata); err != nil {
			t.Fatal(err)
		}
	}
	want := &configfile.UnsupportedWorkspaceFormatError{BeadsDir: beadsDir}
	if metadata.GraphMode != "" && metadata.GraphMode != configfile.GraphModeDependency {
		want.GraphMode = metadata.GraphMode
	} else {
		want.FormatMarker = true
	}
	jsonMode := false
	for _, arg := range args {
		if arg == "--json" {
			jsonMode = true
		}
	}
	if jsonMode {
		var got struct {
			SchemaVersion int    `json:"schema_version"`
			Error         string `json:"error"`
			Hint          string `json:"hint"`
		}
		dec := json.NewDecoder(strings.NewReader(output))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&got); err != nil {
			t.Fatalf("decode ordinary format refusal: %v\n%s", err, output)
		}
		if got.SchemaVersion != JSONSchemaVersion || !strings.Contains(got.Error, want.Error()) || got.Hint != want.Hint() {
			t.Fatalf("ordinary format refusal differs: %+v, want %q / %q", got, want.Error(), want.Hint())
		}
		var extra any
		if err := dec.Decode(&extra); err != io.EOF {
			t.Fatalf("trailing format diagnostic: %v", err)
		}
	} else if !strings.Contains(output, want.Error()) || !strings.Contains(output, want.Hint()) {
		t.Fatalf("ordinary format refusal lacks exact typed error/hint: %s", output)
	}
}

// A refusal in a helper is insufficient: ordinary commands must reach it before
// metadata migration, version tracking, or an attempt to provision/open storage.
func TestGraphModeCLIRefusesLegacyStoreWithoutWorkspaceEffects(t *testing.T) {
	bd := buildBDUnderTest(t)
	commands := [][]string{
		{"show", "beads/plan"},
		{"list", "--json"},
		{"create", "graph fence probe"},
		{"context", "--json"},
		{"where"},
		{"dolt", "start"},
		{"doctor", "--json"},
		{"bootstrap", "--dry-run", "--yes"},
		{"init", "--quiet", "--non-interactive", "--skip-hooks", "--skip-agents"},
		{"init", "--force", "--quiet", "--non-interactive", "--skip-hooks", "--skip-agents"},
		// What the installed git hooks run; their shim passes exit 16 on.
		{"hooks", "run", "pre-commit"},
		// Nothing listens on the url: a connect that got as far as its
		// handshake fails there instead.
		{"connect", "http://127.0.0.1:1"},
	}
	for _, fixture := range graphModeAdmissionFixtures() {
		for _, args := range commands {
			name := args[0]
			if name == "init" && args[1] == "--force" {
				name = "init-force"
			}
			t.Run(fixture.name+"/"+name, func(t *testing.T) {
				repoDir := t.TempDir()
				initGitRepoAt(t, repoDir)
				beadsDir := filepath.Join(repoDir, ".beads")
				metadataPath := fixture.write(t, beadsDir)
				before := legacyUpgradeTreeDigest(t, beadsDir)

				output, code := runGraphModeAdmissionBD(t, bd, repoDir, args...)
				requireGraphModeRefusal(t, fixture, args, output, code)
				if after := legacyUpgradeTreeDigest(t, beadsDir); after != before {
					t.Fatalf("refused bd %s changed the workspace\n%s", strings.Join(args, " "), output)
				}
				if metadataPath != "" {
					if got, err := os.ReadFile(metadataPath); err != nil || string(got) != fixture.metadata {
						t.Fatalf("refused bd %s rewrote %s: %q, %v", strings.Join(args, " "), fixture.filename, got, err)
					}
				}
			})
		}
	}
}

// Commands also reach a workspace they do not run in: create --repo, from
// another workspace or from none and as a dry run, --db and BEADS_DB,
// init --proxied-server, bootstrap in a workspace with no metadata of its
// own, which takes the nearest ancestor's (and refuses a graph ancestor even
// when only its format marker declares it, rather than skip past it), and
// connect in a linked worktree, which attaches the main checkout's .beads.
// Each must refuse a graph workspace the same way, before resolving it
// migrates a legacy config.json or anything opens its store.
func TestGraphModeOutsideEntryPointsRefuseWithoutWorkspaceEffects(t *testing.T) {
	bd := buildBDUnderTest(t)
	sourceDir, _, _ := bdInit(t, bd, "--prefix", "src")
	outsideDir := t.TempDir()
	entries := []struct {
		name string
		// run returns where bd runs, what it adds to bd's environment and
		// bd's arguments, for a graph workspace in repoDir.
		run func(t *testing.T, repoDir string) (dir string, env, args []string)
	}{
		{"create --repo from a workspace", func(t *testing.T, repoDir string) (string, []string, []string) {
			return sourceDir, nil, []string{"create", "--repo", repoDir, "graph fence probe"}
		}},
		{"create --repo from outside", func(t *testing.T, repoDir string) (string, []string, []string) {
			return outsideDir, nil, []string{"create", "--repo", repoDir, "graph fence probe"}
		}},
		{"create --dry-run --repo", func(t *testing.T, repoDir string) (string, []string, []string) {
			// Without --parent the preview opens no target at all.
			return sourceDir, nil, []string{"create", "--dry-run", "--repo", repoDir, "graph fence probe"}
		}},
		{"create --dry-run --parent --repo", func(t *testing.T, repoDir string) (string, []string, []string) {
			// A dry run looks the parent up in the target's store.
			return sourceDir, nil, []string{"create", "--dry-run", "--parent", "src-abc", "--repo", repoDir, "graph fence probe"}
		}},
		{"--db", func(t *testing.T, repoDir string) (string, []string, []string) {
			return outsideDir, nil, []string{"--db", filepath.Join(repoDir, ".beads"), "list", "--json"}
		}},
		{"BEADS_DB", func(t *testing.T, repoDir string) (string, []string, []string) {
			return outsideDir, []string{"BEADS_DB=" + filepath.Join(repoDir, ".beads")}, []string{"list", "--json"}
		}},
		{"init --proxied-server", func(t *testing.T, repoDir string) (string, []string, []string) {
			return repoDir, nil, []string{"init", "--proxied-server", "--quiet", "--non-interactive", "--skip-hooks", "--skip-agents"}
		}},
		{"bootstrap below it", func(t *testing.T, repoDir string) (string, []string, []string) {
			innerDir := filepath.Join(repoDir, "inner")
			writeFile(t, filepath.Join(innerDir, ".beads", "config.yaml"), []byte("# bd workspace config\n"))
			initGitRepoAt(t, innerDir)
			return innerDir, nil, []string{"bootstrap", "--dry-run", "--yes"}
		}},
		// connect writes to the .beads it resolves itself, which from a
		// linked worktree with none of its own is the main checkout's, even
		// where discovery passes over it (a marker or a legacy config.json
		// alone). Nothing listens on the url, so a connect that got as far as
		// its handshake fails there instead.
		{"connect from a linked worktree", func(t *testing.T, repoDir string) (string, []string, []string) {
			return graphModeLinkedWorktree(t, repoDir), nil, []string{"connect", "http://127.0.0.1:1"}
		}},
		{"connect --clear from a linked worktree", func(t *testing.T, repoDir string) (string, []string, []string) {
			return graphModeLinkedWorktree(t, repoDir), nil, []string{"connect", "--clear"}
		}},
	}
	for _, fixture := range graphModeAdmissionFixtures() {
		for _, entry := range entries {
			t.Run(fixture.name+"/"+entry.name, func(t *testing.T) {
				repoDir := t.TempDir()
				initGitRepoAt(t, repoDir)
				beadsDir := filepath.Join(repoDir, ".beads")
				metadataPath := fixture.write(t, beadsDir)
				dir, env, args := entry.run(t, repoDir)
				before := legacyUpgradeTreeDigest(t, beadsDir)

				output, code := runGraphModeAdmissionBDEnv(t, bd, dir, env, args...)
				switch entry.name {
				case "--db", "BEADS_DB", "init --proxied-server":
					requireGraphModeRefusal(t, fixture, args, output, code)
				default:
					requireLegacyGraphFormatRefusal(t, fixture, beadsDir, args, output, code)
				}
				if after := legacyUpgradeTreeDigest(t, beadsDir); after != before {
					t.Fatalf("bd %s changed the graph workspace\n%s", strings.Join(args, " "), output)
				}
				if metadataPath != "" {
					if got, err := os.ReadFile(metadataPath); err != nil || string(got) != fixture.metadata {
						t.Fatalf("refused bd %s rewrote %s: %q, %v", strings.Join(args, " "), fixture.filename, got, err)
					}
				}
			})
		}
	}
}

// bd connect checks the format of the workspace it attaches before its
// handshake, but a graph build can declare that format while the handshake is
// in flight. connect checks again once the handshake returns. It must render
// that refusal as every other graph refusal is rendered (the format, its hint
// and exit 16), not as a generic failure, and refuse before it writes anything
// there: the workspace keeps exactly what the graph build declared, with no
// .gitignore, sidecar or metadata.json of connect's.
func TestGraphModeConnectRefusesAFormatDeclaredDuringItsHandshake(t *testing.T) {
	bd := buildBDUnderTest(t)
	const handshake = `{"api_version":"v0","backend":"dolt","bd_version":"9.9.9",` +
		`"beads_dir":"/srv/.beads","capabilities":["issues.get"],"database":"beads",` +
		`"dolt_mode":"server","project_id":"proj-graph-fence","repo_root":"/srv",` +
		`"schema_version":1,"wire_revision":1,"min_client_wire_revision":1}`
	for _, fixture := range graphModeAdmissionFixtures() {
		if fixture.embedded {
			// A database is not what a concurrent writer adds mid-handshake.
			continue
		}
		for _, flags := range [][]string{nil, {"--json"}} {
			t.Run(fixture.name+"/connect"+strings.Join(flags, ""), func(t *testing.T) {
				repoDir := t.TempDir()
				initGitRepoAt(t, repoDir)
				beadsDir := filepath.Join(repoDir, ".beads")
				// The workspace as the graph build alone leaves it.
				declaredDir := filepath.Join(t.TempDir(), ".beads")
				if err := fixture.declare(declaredDir); err != nil {
					t.Fatal(err)
				}
				want := legacyUpgradeTreeDigest(t, declaredDir)
				declared := make(chan error, 1)
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path != "/v0/beads/context" {
						http.NotFound(w, r)
						return
					}
					select {
					case declared <- fixture.declare(beadsDir):
					default:
					}
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte(handshake))
				}))
				t.Cleanup(server.Close)
				args := append(append([]string{"connect"}, flags...), server.URL)

				output, code := runGraphModeAdmissionBD(t, bd, repoDir, args...)
				select {
				case err := <-declared:
					if err != nil {
						t.Fatalf("declaring the graph workspace during the handshake: %v", err)
					}
				default:
					t.Fatalf("bd %s never reached its handshake, so the late refusal went unexercised\n%s", strings.Join(args, " "), output)
				}
				requireLegacyGraphFormatRefusal(t, fixture, beadsDir, args, output, code)
				if _, err := os.Lstat(filepath.Join(beadsDir, "http_target.json")); !os.IsNotExist(err) {
					t.Fatalf("refused bd %s wrote the http sidecar (%v)\n%s", strings.Join(args, " "), err, output)
				}
				if fixture.filename != "" {
					if got, err := os.ReadFile(filepath.Join(beadsDir, fixture.filename)); err != nil || string(got) != fixture.metadata {
						t.Fatalf("refused bd %s rewrote %s: %q, %v", strings.Join(args, " "), fixture.filename, got, err)
					}
				}
				if fixture.filename != "metadata.json" {
					if _, err := os.Lstat(filepath.Join(beadsDir, "metadata.json")); !os.IsNotExist(err) {
						t.Fatalf("refused bd %s wrote metadata.json (%v)", strings.Join(args, " "), err)
					}
				}
				if got := legacyUpgradeTreeDigest(t, beadsDir); got != want {
					t.Fatalf("refused bd %s changed the graph workspace\n%s", strings.Join(args, " "), output)
				}
			})
		}
	}
}

// declare lays down the fixture's metadata and marker under beadsDir, as a
// graph build writing concurrently would, without a *testing.T: it runs on
// a server goroutine.
func (f graphModeAdmissionFixture) declare(beadsDir string) error {
	if err := os.MkdirAll(beadsDir, 0o700); err != nil {
		return err
	}
	if f.filename != "" {
		if err := os.WriteFile(filepath.Join(beadsDir, f.filename), []byte(f.metadata), 0o600); err != nil {
			return err
		}
	}
	if f.marker {
		return os.WriteFile(filepath.Join(beadsDir, "graph-preview-format"), []byte("link-preview-v9\n"), 0o600)
	}
	return nil
}

// graphModeLinkedWorktree commits once in repoDir and adds a linked worktree
// of it. The worktree has no .beads of its own: the workspace fixture under
// repoDir is not committed.
func graphModeLinkedWorktree(t *testing.T, repoDir string) string {
	t.Helper()
	worktreeDir := filepath.Join(t.TempDir(), "worktree")
	runGitForWorktreeTest(t, repoDir, "commit", "--allow-empty", "--quiet", "-m", "init")
	runGitForWorktreeTest(t, repoDir, "worktree", "add", "--quiet", "--detach", worktreeDir)
	return worktreeDir
}

// Commands that never open a store keep working in a workspace this build
// refuses, and leave it untouched.
func TestGraphModeWorkspaceAdmitsStorelessCommands(t *testing.T) {
	bd := buildBDUnderTest(t)
	var fixtures []graphModeAdmissionFixture
	for _, fixture := range graphModeAdmissionFixtures() {
		if fixture.name == "metadata.json/link/embedded" || fixture.name == "marker/only" {
			fixtures = append(fixtures, fixture)
		}
	}
	if len(fixtures) != 2 {
		t.Fatalf("store-less fixtures = %d, want 2", len(fixtures))
	}
	for _, fixture := range fixtures {
		for _, args := range [][]string{
			{"version"},
			{"--version"},
			{"help"},
			{"completion", "bash"},
		} {
			t.Run(fixture.name+"/"+strings.Join(args, " "), func(t *testing.T) {
				repoDir := t.TempDir()
				initGitRepoAt(t, repoDir)
				beadsDir := filepath.Join(repoDir, ".beads")
				fixture.write(t, beadsDir)
				before := legacyUpgradeTreeDigest(t, beadsDir)

				output, code := runGraphModeAdmissionBD(t, bd, repoDir, args...)
				if args[0] == "completion" {
					requireGraphModeRefusal(t, fixture, args, output, code)
				} else if code != 0 {
					t.Fatalf("bd %s exit = %d, want 0\n%s", strings.Join(args, " "), code, output)
				}
				if strings.Contains(output, "this bd build does not support") {
					t.Fatalf("store-less bd %s reported the workspace format:\n%s", strings.Join(args, " "), output)
				}
				if after := legacyUpgradeTreeDigest(t, beadsDir); after != before {
					t.Fatalf("bd %s changed the workspace", strings.Join(args, " "))
				}
			})
		}
	}
}

// A normal current-generation initialization must reach the graph provider,
// despite ordinary Issue opens refusing graph formats.
func TestGraphModeCLIAdmitsCurrentGenerationWithoutLegacyOpen(t *testing.T) {
	bd := buildBDUnderTest(t)
	work, home := t.TempDir(), t.TempDir()
	graphPolicyCLI(t, bd, work, home, nil, "", "init", "--graph-mode", "link", "--scope-url", "https://example.invalid/admission/", "--skip-hooks", "--skip-agents", "--non-interactive")
	beadsDir := filepath.Join(work, ".beads")
	cfg, err := configfile.LoadForDiscovery(beadsDir)
	if err != nil || cfg == nil {
		t.Fatalf("fresh binding: %v", err)
	}
	real, err := filepath.EvalSymlinks(beadsDir)
	if err != nil {
		t.Fatal(err)
	}
	marker, err := os.ReadFile(filepath.Join(beadsDir, graphPreviewMarker))
	if err != nil || string(marker) != graphPreviewGeneration || cfg.GraphSchemaVersion != graphstore.SchemaVersion || !cfg.GraphReady || cfg.GraphWorkspace != real || cfg.GraphScopeURL == "" || cfg.GraphAuthorityID == "" || cfg.DoltDatabase == "" {
		t.Fatalf("fresh graph binding/marker invalid: %+v marker=%q err=%v", cfg, marker, err)
	}
	out := graphPolicyCLI(t, bd, work, home, nil, "", "types", "--json")
	var inventory struct {
		Result struct {
			Beads []json.RawMessage `json:"beadTypes"`
			Links []json.RawMessage `json:"linkTypes"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(out), &inventory); err != nil || len(inventory.Result.Beads) != 20 || len(inventory.Result.Links) != 25 {
		t.Fatalf("current graph Type inventory: %v\n%s", err, out)
	}
	original, err := os.ReadFile(configfile.ConfigPath(beadsDir))
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name          string
		edit          func(*configfile.Config)
		code, message string
	}{
		{"not-ready", func(c *configfile.Config) { c.GraphReady = false }, "graph_not_initialized", graphGenerationSecondGateRefusal},
		{"schema", func(c *configfile.Config) { c.GraphSchemaVersion++ }, "graph_not_initialized", graphGenerationSecondGateRefusal},
		{"scope", func(c *configfile.Config) { c.GraphScopeURL = "" }, "graph_not_initialized", graphGenerationSecondGateRefusal},
		{"authority", func(c *configfile.Config) { c.GraphAuthorityID = "" }, "graph_not_initialized", graphGenerationSecondGateRefusal},
		{"workspace", func(c *configfile.Config) { c.GraphWorkspace = "" }, "graph_not_initialized", graphGenerationSecondGateRefusal},
		{"database", func(c *configfile.Config) { c.DoltDatabase = "" }, "graph_not_initialized", graphGenerationSecondGateRefusal},
		{"physical-binding", func(c *configfile.Config) { c.GraphWorkspace = filepath.Join(real, "other") }, "not_authority", "graph_mode workspace binding differs; copied/moved workspaces cannot claim this authority"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			changed := *cfg
			tc.edit(&changed)
			if err := changed.Save(beadsDir); err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := os.WriteFile(configfile.ConfigPath(beadsDir), original, 0o600); err != nil {
					t.Fatal(err)
				}
			}()
			before := legacyUpgradeTreeDigest(t, beadsDir)
			stdout, stderr, exit := graphGenerationRun(t, bd, work, home, "types", "--json")
			if stdout != "" {
				t.Fatalf("refusal stdout=%q", stdout)
			}
			requireExactGraphAdmissionFailure(t, []string{"types", "--json"}, stderr, exit, tc.code, tc.message)
			if after := legacyUpgradeTreeDigest(t, beadsDir); after != before {
				t.Fatal("refused graph read changed workspace")
			}
		})
	}
}
