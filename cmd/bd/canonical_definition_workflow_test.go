//go:build cgo

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/configfile"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// The fixture selects the new fresh layout through its production initializer.
// No definition row is injected by this test. Each ordinary CLI command then
// opens and closes that same persisted workspace in a separate process.
func intrinsicDefinitionCLI(t *testing.T, demoVariable string) (map[string][]byte, func(*testing.T, int, ...string) string) {
	return intrinsicDefinitionCLIWithInit(t, demoVariable, false)
}

// These refusals must precede workspace, guidance and engine creation. Test the
// actual process boundary so ordinary legacy init cannot swallow the selector.
func TestIntrinsicInitRefusesBeforeEffects(t *testing.T) {
	isolateBeadsDirForTest(t)
	bd := buildBDUnderTest(t)
	for _, tc := range []struct {
		name string
		args []string
		exit int
	}{
		{"missing-mode", []string{"--graph-profile", "types"}, 2},
		{"dependency-mode", []string{"--graph-mode", "dependency", "--graph-profile", "types"}, 2},
		{"empty-profile", []string{"--graph-mode", "link", "--graph-profile="}, 2},
		{"unknown-profile", []string{"--graph-mode", "link", "--graph-profile", "14"}, 2},
		{"duplicate-profile", []string{"--graph-mode", "link", "--graph-profile", "types", "--graph-profile=types"}, 2},
		{"conflicting-mode", []string{"--graph-mode", "dependency", "--graph-mode", "link", "--graph-profile", "types"}, 2},
		{"readonly", []string{"--graph-mode", "link", "--graph-profile", "types", "--readonly"}, 5},
		{"invalid-scope", []string{"--graph-mode", "link", "--graph-profile", "types", "--scope-url", "not-a-scope"}, 2},
		{"server", []string{"--graph-mode", "link", "--graph-profile", "types", "--server", "--external"}, 5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			args := append([]string{"init", "--scope-url", "https://example.invalid/init-refusal/"}, tc.args...)
			command := exec.CommandContext(ctx, bd, args...)
			command.Dir = work
			command.Env = []string{"PATH=" + os.Getenv("PATH"), "HOME=" + home,
				"XDG_CONFIG_HOME=" + filepath.Join(home, ".config"), "DOLT_ROOT_PATH=" + filepath.Join(home, "dolt-root"),
				"GIT_CONFIG_GLOBAL=" + filepath.Join(home, "missing-gitconfig"), "GIT_CONFIG_NOSYSTEM=1",
				"BD_DISABLE_METRICS=1", "BD_NON_INTERACTIVE=1", "BEADS_DOLT_AUTO_START=0", "DOLT_METRICS_DISABLED=1"}
			var stdout, stderr bytes.Buffer
			command.Stdout, command.Stderr = &stdout, &stderr
			err := command.Run()
			var failed *exec.ExitError
			if !errors.As(err, &failed) || failed.ExitCode() != tc.exit || ctx.Err() != nil || stdout.Len() != 0 || stderr.Len() == 0 {
				t.Fatalf("init refusal exit=%v context=%v stdout=%q stderr=%q", err, ctx.Err(), stdout.String(), stderr.String())
			}
			entries, err := os.ReadDir(work)
			if err != nil || len(entries) != 0 {
				t.Fatalf("refused init changed the workspace: %v %v", entries, err)
			}
			if _, err := os.Stat(filepath.Join(home, "dolt-root")); !os.IsNotExist(err) {
				t.Fatalf("refused init reached the engine: %v", err)
			}
		})
	}
}

func intrinsicDefinitionCLIWithInit(t *testing.T, demoVariable string, publicInit bool) (map[string][]byte, func(*testing.T, int, ...string) string) {
	t.Helper()
	isolateBeadsDirForTest(t)
	bd := buildBDUnderTest(t)
	work, home := t.TempDir(), t.TempDir()
	// An explicit development demo retains only a newly created private fixture.
	// Never initialize, replace or clean up a preexisting path or workspace.
	if demo := os.Getenv(demoVariable); demo != "" {
		if !filepath.IsAbs(demo) || filepath.Clean(demo) != demo {
			t.Fatalf("%s must be a clean absolute fresh path", demoVariable)
		}
		if err := os.Mkdir(demo, 0o700); err != nil {
			t.Fatalf("refuse nonfresh demo path: %v", err)
		}
		work, home = filepath.Join(demo, "work"), filepath.Join(demo, "home")
		for _, directory := range []string{work, home} {
			if err := os.Mkdir(directory, 0o700); err != nil {
				t.Fatal(err)
			}
		}
		t.Logf("RETAINED_FRESH_DEMO work=%s home=%s binary=%s", work, home, bd)
	}
	call := func(t *testing.T, wantExit int, args ...string) string {
		t.Helper()
		commandContext, stop := context.WithTimeout(context.Background(), 60*time.Second)
		defer stop()
		command := exec.CommandContext(commandContext, bd, args...)
		command.Dir = work
		command.Env = []string{
			"PATH=" + os.Getenv("PATH"), "HOME=" + home,
			"TMPDIR=" + os.TempDir(), "TMP=" + os.TempDir(), "TEMP=" + os.TempDir(),
			"XDG_CONFIG_HOME=" + filepath.Join(home, ".config"),
			"XDG_CACHE_HOME=" + filepath.Join(home, ".cache"),
			"DOLT_ROOT_PATH=" + filepath.Join(home, "dolt-root"),
			"GIT_CONFIG_GLOBAL=" + filepath.Join(home, "missing-gitconfig"), "GIT_CONFIG_NOSYSTEM=1",
			"BD_DISABLE_METRICS=1", "BD_DISABLE_EVENT_FLUSH=1", "BD_NON_INTERACTIVE=1",
			"BEADS_DOLT_AUTO_START=0", "DOLT_METRICS_DISABLED=1", "NO_COLOR=1",
		}
		var stdout, stderr bytes.Buffer
		command.Stdout, command.Stderr = &stdout, &stderr
		err := command.Run()
		t.Logf("ACTUAL_CLI args=%q exit=%v\nSTDOUT\n%s\nSTDERR\n%s", args, err, stdout.String(), stderr.String())
		exit := 0
		if err != nil {
			var failed *exec.ExitError
			if !errors.As(err, &failed) {
				t.Fatalf("CLI launch failed: %v", err)
			}
			exit = failed.ExitCode()
		}
		if exit != wantExit || commandContext.Err() != nil || wantExit == 0 && stderr.Len() != 0 {
			t.Fatalf("ordinary command exit %d, want %d: %v context=%v", exit, wantExit, err, commandContext.Err())
		}
		return stdout.String()
	}
	const scope = "https://example.invalid/intrinsic-show/"
	beadsDir := filepath.Join(work, ".beads")
	var cfg *configfile.Config
	if publicInit {
		call(t, 0, "init", "--graph-mode", "link", "--graph-profile", "types", "--scope-url", scope, "--json")
		var err error
		cfg, err = configfile.LoadForDiscovery(beadsDir)
		if err != nil || cfg == nil || !cfg.GraphReady || cfg.GraphSchemaVersion != graphstore.TypesFreshSchemaVersion || cfg.GraphScopeURL != scope {
			t.Fatalf("public initializer did not publish the selected ready profile: %+v %v", cfg, err)
		}
		guidance, err := os.ReadFile(filepath.Join(work, "AGENTS.md"))
		if err != nil || !strings.Contains(string(guidance), "profile:graph-types") || strings.Contains(string(guidance), "bd remember") || strings.Contains(string(guidance), "bd setup") {
			t.Fatalf("public Types guidance: %v %s", err, guidance)
		}
		if _, err := os.Stat(filepath.Join(work, ".claude", "settings.json")); !os.IsNotExist(err) {
			t.Fatalf("Types-only initialization installed unsupported hooks: %v", err)
		}
		metadata, err := os.ReadFile(filepath.Join(beadsDir, "metadata.json"))
		if err != nil {
			t.Fatal(err)
		}
		call(t, 5, "init", "--graph-mode", "link", "--graph-profile", "types", "--scope-url", scope, "--json")
		after, err := os.ReadFile(filepath.Join(beadsDir, "metadata.json"))
		if err != nil || !bytes.Equal(metadata, after) {
			t.Fatalf("repeated init changed existing metadata: %v", err)
		}
	} else {
		if err := os.Mkdir(beadsDir, 0o700); err != nil {
			t.Fatal(err)
		}
		real, err := filepath.EvalSymlinks(beadsDir)
		if err != nil {
			t.Fatal(err)
		}
		cfg = &configfile.Config{
			Backend: "dolt", GraphMode: "link", GraphScopeURL: scope,
			GraphAuthorityID: strings.Repeat("b", 32), GraphWorkspace: real,
			GraphSchemaVersion: graphstore.TypesFreshSchemaVersion,
			DoltMode:           "embedded", DoltDatabase: "intrinsic_show", ProjectID: strings.Repeat("b", 32),
		}
		options := graphOptions(cfg)
		options.DataDir = filepath.Join(real, "embeddeddolt")
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()
		if err := graphstore.Init(ctx, options); err != nil {
			t.Fatal(err)
		}
		cfg.GraphReady = true
		if err := cfg.Save(real); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(real, graphPreviewMarker), []byte(graphPreviewGeneration), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	options := graphOptions(cfg)
	options.DataDir = filepath.Join(cfg.GraphWorkspace, "embeddeddolt")
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	store, err := graphstore.OpenExisting(ctx, options)
	if err != nil {
		t.Fatal(err)
	}
	ids := []string{"type-for-bead-type", "type-for-type", "type-for-link-type", "type-for-refines"}
	expected := make(map[string][]byte, len(ids))
	for _, id := range ids {
		document, err := store.GetDefinition(ctx, id, "")
		if err != nil {
			_ = store.Close()
			t.Fatal(err)
		}
		expected[id] = document.NativeJSON()
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	return expected, call
}

func TestIntrinsicDefinitionsOrdinaryShowReopen(t *testing.T) {
	expected, call := intrinsicDefinitionCLI(t, "BEADS_INTRINSIC_DEMO_DIR")
	ids := []string{"type-for-bead-type", "type-for-type", "type-for-link-type", "type-for-refines"}
	// The first requested command must work without an output-mode opt-in.
	plain := call(t, 0, "show", ids[0])
	if !strings.Contains(plain, "https://example.invalid/intrinsic-show/"+ids[0]) || !strings.Contains(plain, "propertiesSchema") {
		t.Fatalf("ordinary Show omitted the actual definition: %s", plain)
	}
	for _, id := range ids {
		t.Run(id, func(t *testing.T) {
			out := call(t, 0, "show", id, "--json")
			var envelope struct {
				Outcome string          `json:"outcome"`
				Spec    string          `json:"bdp_spec"`
				Data    json.RawMessage `json:"data"`
			}
			if err := json.Unmarshal([]byte(out), &envelope); err != nil || envelope.Outcome != "success" || envelope.Spec != intrinsicDefinitionContract {
				t.Fatalf("ordinary JSON Show: %v %s", err, out)
			}
			actual, err := graph.CanonicalizeJSON(envelope.Data)
			if err != nil || !bytes.Equal(actual, expected[id]) {
				t.Fatalf("CLI changed the persisted definition: %v\n%s\nexpected %s", err, actual, expected[id])
			}
		})
	}
	// Another process proves reopen preserves the first record and exact pins.
	again := call(t, 0, "show", "https://example.invalid/intrinsic-show/"+ids[0], "--output-format=json")
	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal([]byte(again), &envelope); err != nil {
		t.Fatal(err)
	}
	actual, err := graph.CanonicalizeJSON(envelope.Data)
	if err != nil || !bytes.Equal(actual, expected[ids[0]]) {
		t.Fatalf("reopen changed the BeadType definition: %v %s", err, again)
	}
}

func TestIntrinsicDefinitionsOrdinaryList(t *testing.T) {
	expected, call := intrinsicDefinitionCLI(t, "BEADS_INTRINSIC_LIST_DEMO_DIR")
	plain := call(t, 0, "list", "--type", "type")
	for id := range expected {
		if !strings.Contains(plain, id) {
			t.Fatalf("ordinary List omitted %s: %s", id, plain)
		}
	}
	for _, tc := range []struct {
		name string
		args []string
		ids  []string
	}{
		{"type-family", []string{"--type", "type"}, []string{"type-for-bead-type", "type-for-type", "type-for-link-type", "type-for-refines"}},
		{"bead-type-family", []string{"--type", "bead-type"}, []string{"type-for-bead-type", "type-for-type", "type-for-link-type"}},
		{"link-type-family", []string{"--type", "link-type"}, []string{"type-for-refines"}},
		{"exact-abstract-type", []string{"--exact-type", "type"}, nil},
		{"intersected-families", []string{"--type", "link-type", "--exact-type", "bead-type"}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			args := append([]string{"list", "--json"}, tc.args...)
			out := call(t, 0, args...)
			var envelope struct {
				Outcome string `json:"outcome"`
				Spec    string `json:"bdp_spec"`
				Data    struct {
					Items []json.RawMessage `json:"items"`
					Next  json.RawMessage   `json:"next"`
				} `json:"data"`
				Completion struct {
					Returned  int  `json:"returned"`
					Exhausted bool `json:"exhausted"`
				} `json:"completion"`
			}
			if err := json.Unmarshal([]byte(out), &envelope); err != nil || envelope.Outcome != "success" || envelope.Spec != intrinsicDefinitionContract {
				t.Fatalf("ordinary List envelope: %v %s", err, out)
			}
			if len(envelope.Data.Items) != len(tc.ids) || envelope.Completion.Returned != len(tc.ids) || !envelope.Completion.Exhausted || string(envelope.Data.Next) != "null" {
				t.Fatalf("incomplete or wrong List selection: %s", out)
			}
			wanted := make(map[string]bool, len(tc.ids))
			for _, id := range tc.ids {
				wanted["https://example.invalid/intrinsic-show/"+id] = true
			}
			for _, raw := range envelope.Data.Items {
				var identity struct {
					ID string `json:"id"`
				}
				if err := json.Unmarshal(raw, &identity); err != nil || !wanted[identity.ID] {
					t.Fatalf("unexpected or duplicate definition: %v %s", err, raw)
				}
				delete(wanted, identity.ID)
				local := strings.TrimPrefix(identity.ID, "https://example.invalid/intrinsic-show/")
				actual, err := graph.CanonicalizeJSON(raw)
				if err != nil || !bytes.Equal(actual, expected[local]) {
					t.Fatalf("List changed persisted definition: %v %s", err, raw)
				}
			}
		})
	}
}

func TestIntrinsicFooDefinitionCreateAndPrototype(t *testing.T) {
	testIntrinsicFooDefinitionCreateAndPrototype(t, "BEADS_INTRINSIC_FOO_DEMO_DIR", false)
}

// Public setup owns a distinct process and readiness-publication boundary. No
// metadata, marker or definition row is installed by the test for this journey.
func TestIntrinsicDefinitionsPublicInit(t *testing.T) {
	testIntrinsicFooDefinitionCreateAndPrototype(t, "BEADS_INTRINSIC_PUBLIC_DEMO_DIR", true)
}

func testIntrinsicFooDefinitionCreateAndPrototype(t *testing.T, demoVariable string, publicInit bool) {
	expected, call := intrinsicDefinitionCLIWithInit(t, demoVariable, publicInit)
	if publicInit {
		for _, id := range []string{"type-for-bead-type", "type-for-type", "type-for-link-type", "type-for-refines"} {
			out := call(t, 0, "show", id, "--json")
			var envelope struct {
				Data json.RawMessage `json:"data"`
			}
			if err := json.Unmarshal([]byte(out), &envelope); err != nil {
				t.Fatal(err)
			}
			actual, err := graph.CanonicalizeJSON(envelope.Data)
			want, expectedErr := graph.CanonicalizeJSON(expected[id])
			if err != nil || expectedErr != nil || !bytes.Equal(actual, want) {
				t.Fatalf("public setup Show differs from persistence: %s %v %v", id, err, expectedErr)
			}
		}
		out := call(t, 0, "list", "--type", "link-type", "--json")
		var envelope struct {
			Data struct {
				Items []json.RawMessage `json:"items"`
			} `json:"data"`
		}
		if err := json.Unmarshal([]byte(out), &envelope); err != nil || len(envelope.Data.Items) != 1 {
			t.Fatalf("public LinkType family: %v %s", err, out)
		}
	}
	created := call(t, 0, "type", "create", "--id", "foo", "--json")
	decode := func(out string) json.RawMessage {
		t.Helper()
		var result struct {
			Outcome string          `json:"outcome"`
			Spec    string          `json:"bdp_spec"`
			Data    json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal([]byte(out), &result); err != nil || result.Outcome != "success" || result.Spec != intrinsicDefinitionContract {
			t.Fatalf("definition result: %v %s", err, out)
		}
		return result.Data
	}
	definition := decode(created)
	var identity struct {
		ID         string                     `json:"id"`
		Revision   string                     `json:"revision"`
		Properties map[string]json.RawMessage `json:"properties"`
		Type       struct {
			URI      string `json:"uri"`
			Revision string `json:"revision"`
		} `json:"type"`
	}
	if err := json.Unmarshal(definition, &identity); err != nil || identity.ID != "https://example.invalid/intrinsic-show/foo" || identity.Revision == "" || identity.Type.URI != "https://github.com/gastownhall/bdp/types/type-system/bead-type" || identity.Type.Revision != "bead-type-2" {
		t.Fatalf("new local definition identity: %v %s", err, definition)
	}
	if identity.Properties == nil || len(identity.Properties) != 0 {
		t.Fatalf("fresh definition manufactured local schema or properties: %s", definition)
	}
	assertSame := func(out string) {
		t.Helper()
		actual, err := graph.CanonicalizeJSON(decode(out))
		expected, expectedErr := graph.CanonicalizeJSON(definition)
		if err != nil || expectedErr != nil || !bytes.Equal(actual, expected) {
			t.Fatalf("persisted Foo changed: %v %v %s", err, expectedErr, out)
		}
	}
	assertSame(call(t, 0, "show", "foo", "--json"))
	for _, tc := range []struct {
		family string
		count  int
	}{{"type", 5}, {"bead-type", 4}} {
		out := call(t, 0, "list", "--type", tc.family, "--json")
		var page struct {
			Items []json.RawMessage `json:"items"`
		}
		if err := json.Unmarshal(decode(out), &page); err != nil || len(page.Items) != tc.count {
			t.Fatalf("Foo absent from %s family: %v %s", tc.family, err, out)
		}
		found := 0
		for _, raw := range page.Items {
			var item struct {
				ID string `json:"id"`
			}
			if err := json.Unmarshal(raw, &item); err != nil {
				t.Fatal(err)
			}
			if item.ID == identity.ID {
				found++
				actual, err := graph.CanonicalizeJSON(raw)
				expected, expectedErr := graph.CanonicalizeJSON(definition)
				if err != nil || expectedErr != nil || !bytes.Equal(actual, expected) {
					t.Fatalf("List rewrote Foo: %v %v %s", err, expectedErr, raw)
				}
			}
		}
		if found != 1 {
			t.Fatalf("Foo occurs %d times", found)
		}
	}
	call(t, 4, "type", "create", "--id", "foo", "--json")
	preview := call(t, 0, "create", "--type", "foo", "--properties", `{"note":"prototype"}`, "--json")
	var envelope struct {
		Preview bool                       `json:"preview"`
		Schema  int                        `json:"schemaVersion"`
		Result  map[string]json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal([]byte(preview), &envelope); err != nil || !envelope.Preview || envelope.Schema != 1 || envelope.Result == nil {
		t.Fatalf("prototype lost the existing preview envelope: %v %s", err, preview)
	}
	value := envelope.Result
	if string(value["prototype"]) != "true" || string(value["persisted"]) != "false" || value["id"] != nil || value["revision"] != nil {
		t.Fatalf("prototype falsely claims persistence: %s", preview)
	}
	var selected struct {
		URI      string `json:"uri"`
		Revision string `json:"revision"`
	}
	if err := json.Unmarshal(value["type"], &selected); err != nil || selected.URI != identity.ID || selected.Revision != identity.Revision {
		t.Fatalf("prototype lost exact Type: %v %s", err, preview)
	}
	if publicInit {
		const proposed = "Examples/My%20Bead"
		out := call(t, 0, "create", "--id", proposed, "--type", "foo", "--json")
		var proposal struct {
			Result struct {
				Prototype bool            `json:"prototype"`
				Persisted bool            `json:"persisted"`
				Proposed  string          `json:"proposedId"`
				ID        json.RawMessage `json:"id"`
				Revision  json.RawMessage `json:"revision"`
			} `json:"result"`
		}
		if err := json.Unmarshal([]byte(out), &proposal); err != nil || !proposal.Result.Prototype || proposal.Result.Persisted || proposal.Result.Proposed != "https://example.invalid/intrinsic-show/Examples/My%20Bead" || proposal.Result.ID != nil || proposal.Result.Revision != nil {
			t.Fatalf("proposed identity escaped the unsaved boundary: %v %s", err, out)
		}
		call(t, 3, "show", proposed, "--json")
	}
	plain := call(t, 0, "create", "--type", "foo")
	if !strings.Contains(plain, "no Bead was saved") {
		t.Fatalf("prototype missing its unsaved boundary: %s", plain)
	}
	assertSame(call(t, 0, "show", "foo", "--json"))
	call(t, 5, "type", "create", "--id", "readonly-control", "--readonly", "--json")
	call(t, 3, "show", "readonly-control", "--json")
}
