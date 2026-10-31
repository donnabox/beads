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
		{"conflicting-atomic-on-off", []string{"--graph-mode", "link", "--atomic-batch=true", "--atomic-batch=false"}, 2},
		{"conflicting-atomic-off-on", []string{"--graph-mode", "link", "--atomic-batch=false", "--atomic-batch=true"}, 2},
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
		call(t, 0, "init", "--graph-mode", "link", "--graph-profile", "types", "--scope-url", scope, "--prefix", "bd", "--json")
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
	// The independent installed-name census catches an acquisition that exposes
	// only the four intrinsic definitions. One complete read avoids an expensive
	// point-Get loop over every installed Type in this native fixture.
	wanted := map[string]bool{"bead-type": true, "type": true, "link-type": true, "refines": true}
	for _, id := range graphstore.FreshTypeNames() {
		if wanted[id] {
			t.Fatalf("fresh Type identity collision: %q", id)
		}
		wanted[id] = true
	}
	capture, err := store.CurrentResourceDocuments(ctx)
	if err != nil {
		_ = store.Close()
		t.Fatal(err)
	}
	expected := make(map[string][]byte, len(wanted))
	for _, document := range capture.Documents {
		if document.Kind() != graph.KindBead {
			continue
		}
		id := strings.TrimPrefix(document.ID(), scope)
		if !wanted[id] || expected[id] != nil {
			_ = store.Close()
			t.Fatalf("unexpected or duplicate fresh Bead: %s", document.ID())
		}
		if _, definition := document.Definition(); !definition {
			_ = store.Close()
			t.Fatalf("fresh installed Type lost checked definition provenance: %s", document.ID())
		}
		expected[id] = document.NativeJSON()
	}
	for id := range wanted {
		if expected[id] == nil {
			_ = store.Close()
			t.Fatalf("installed Type lacks ordinary stored Bead: %s", id)
		}
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	return expected, call
}

func TestIntrinsicDefinitionsOrdinaryShowReopen(t *testing.T) {
	expected, call := intrinsicDefinitionCLI(t, "BEADS_INTRINSIC_DEMO_DIR")
	ids := []string{"bead-type", "type", "link-type", "refines"}
	// The first requested command must work without an output-mode opt-in.
	plain := call(t, 0, "show", ids[0])
	if !strings.Contains(plain, "Local ID: "+ids[0]+"\n") || !strings.Contains(plain, "Type: ") || !strings.Contains(plain, "propertiesSchema") || strings.Contains(plain, "Metatype:") {
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
	plain := call(t, 0, "list", "--type", "type", "--limit", "0")
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
		{"type-family", []string{"--type", "type"}, intrinsicDefinitionFamilyIDs(t, expected, "type")},
		{"bead-type-family", []string{"--type", "bead-type"}, intrinsicDefinitionFamilyIDs(t, expected, "bead-type")},
		{"link-type-family", []string{"--type", "link-type"}, intrinsicDefinitionFamilyIDs(t, expected, "link-type")},
		{"exact-abstract-type", []string{"--exact-type", "type"}, nil},
		{"intersected-families", []string{"--type", "link-type", "--exact-type", "bead-type"}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			args := append([]string{"list", "--output-format", "json", "--unpaged", "--limit", "0"}, tc.args...)
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

func TestIntrinsicFooDefinitionIsOrdinaryStoredBead(t *testing.T) {
	testIntrinsicFooDefinitionIsOrdinaryStoredBead(t, "BEADS_INTRINSIC_FOO_DEMO_DIR", false)
}

// Public setup owns a distinct process and readiness-publication boundary. No
// metadata, marker or definition row is installed by the test for this journey.
func TestIntrinsicDefinitionsPublicInit(t *testing.T) {
	testIntrinsicFooDefinitionIsOrdinaryStoredBead(t, "BEADS_INTRINSIC_PUBLIC_DEMO_DIR", true)
}

func testIntrinsicFooDefinitionIsOrdinaryStoredBead(t *testing.T, demoVariable string, publicInit bool) {
	expected, call := intrinsicDefinitionCLIWithInit(t, demoVariable, publicInit)
	if publicInit {
		for _, id := range []string{"bead-type", "type", "link-type", "refines"} {
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
		out := call(t, 0, "list", "--type", "link-type", "--json", "--unpaged", "--limit", "0")
		var envelope struct {
			Data struct {
				Items []json.RawMessage `json:"items"`
			} `json:"data"`
		}
		if err := json.Unmarshal([]byte(out), &envelope); err != nil || len(envelope.Data.Items) != len(intrinsicDefinitionFamilyIDs(t, expected, "link-type")) {
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
	}{{"type", len(expected) + 1}, {"bead-type", len(intrinsicDefinitionFamilyIDs(t, expected, "bead-type")) + 1}} {
		out := call(t, 0, "list", "--type", tc.family, "--json", "--unpaged", "--limit", "0")
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
	// Definition persistence is the existing feature under test. Donna replaced
	// the unsaved-instance demo contract with real persisted custom create/update
	// in a later increment; this definition journey must not assert fake success.
	assertSame(call(t, 0, "show", "foo", "--json"))
	call(t, 5, "type", "create", "--id", "readonly-control", "--readonly", "--json")
	call(t, 3, "show", "readonly-control", "--json")
}

// Expected family membership comes from the stored definition's actual global
// metatype, independently of the collection selector being exercised.
func intrinsicDefinitionFamilyIDs(t *testing.T, definitions map[string][]byte, family string) []string {
	t.Helper()
	ids := []string{}
	for id, raw := range definitions {
		var record struct {
			Type struct {
				URI string `json:"uri"`
			} `json:"type"`
		}
		if err := json.Unmarshal(raw, &record); err != nil {
			t.Fatal(err)
		}
		if family == "type" || record.Type.URI == "https://github.com/gastownhall/bdp/types/type-system/"+family {
			ids = append(ids, id)
		}
	}
	return ids
}

// This is the first useful increment, using public init and a separate actual
// CLI process for every operation. It does not require custom-instance or batch
// work, or treat the old unsaved preview as a successful creation.
func TestTypesOrdinaryReadsPreserveExistingCreate(t *testing.T) {
	typesOrdinaryReadCreateJourney(t, "bd-333")
}

// Donna explicitly retained native prefixed Issue IDs for this increment.
// The original unforced333 positive journey is retained in Source03; this
// refusal is the accepted contract, without Force or a public/native ID alias.
func TestTypesOrdinaryReadsRejectInvalidIssueID(t *testing.T) {
	_, call := intrinsicDefinitionCLIWithInit(t, "BEADS_TYPES_INVALID_ID_DEMO_DIR", true)
	snapshot := func() []byte {
		t.Helper()
		output := call(t, 0, "list", "--json", "--limit", "0")
		var envelope struct {
			Result struct {
				Items   json.RawMessage `json:"items"`
				HasMore *bool           `json:"hasMore"`
			} `json:"result"`
		}
		if err := json.Unmarshal([]byte(output), &envelope); err != nil || len(envelope.Result.Items) == 0 || envelope.Result.HasMore == nil || *envelope.Result.HasMore {
			t.Fatalf("ordinary snapshot is not complete: %v %s", err, output)
		}
		canonical, err := graph.CanonicalizeJSON(envelope.Result.Items)
		if err != nil {
			t.Fatal(err)
		}
		return canonical
	}
	before := snapshot()
	output := call(t, 2, "create", "--id", "333", "--type", "task", "Hello world", "--json")
	if strings.TrimSpace(output) != "" {
		t.Fatalf("refused creation exposed a result: %s", output)
	}
	if !bytes.Equal(before, snapshot()) {
		t.Fatal("invalid Issue ID changed ordinary records or installed definitions")
	}
	call(t, 3, "show", "333", "--json")
}

func typesOrdinaryReadCreateJourney(t *testing.T, localID string) {
	t.Helper()
	definitions, call := intrinsicDefinitionCLIWithInit(t, "BEADS_TYPES_ORDINARY_DEMO_DIR", true)
	const scope = "https://example.invalid/intrinsic-show/"
	show := func(selector string) json.RawMessage {
		t.Helper()
		out := call(t, 0, "show", selector, "--json")
		var envelope struct {
			Outcome string          `json:"outcome"`
			Data    json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal([]byte(out), &envelope); err != nil || envelope.Outcome != "success" || len(envelope.Data) == 0 {
			t.Fatalf("ordinary Show: %v %s", err, out)
		}
		canonical, err := graph.CanonicalizeJSON(envelope.Data)
		if err != nil {
			t.Fatal(err)
		}
		return canonical
	}
	count := func(want int, selectors ...string) {
		t.Helper()
		args := append([]string{"count", "--json"}, selectors...)
		out := call(t, 0, args...)
		var value struct {
			Count *int `json:"count"`
		}
		if err := json.Unmarshal([]byte(out), &value); err != nil || value.Count == nil || *value.Count != want {
			t.Fatalf("ordinary count, want %d: %v %s", want, err, out)
		}
	}
	count(len(definitions))
	created := call(t, 0, "create", "--id", localID, "--type", "task", "Hello world")
	if !strings.Contains(created, localID) || strings.Contains(created, "no Bead was saved") {
		t.Fatalf("existing creation became a demo: %s", created)
	}
	task := show(localID)
	var record struct {
		ID       string `json:"id"`
		Revision string `json:"revision"`
		Type     struct {
			URI      string `json:"uri"`
			Revision string `json:"revision"`
		} `json:"type"`
		Properties struct {
			Title  string `json:"title"`
			Status string `json:"status"`
		} `json:"properties"`
	}
	if err := json.Unmarshal(task, &record); err != nil {
		t.Fatal(err)
	}
	taskType, err := graphstore.NativeIssueTypeURL(scope, "task")
	if err != nil {
		t.Fatal(err)
	}
	if record.ID != scope+localID || record.Revision == "" || record.Type.URI != taskType || record.Type.Revision == "" || record.Properties.Title != "Hello world" || record.Properties.Status != "open" {
		t.Fatalf("lost persisted native identity, properties or exact Type: %s", task)
	}
	// A first-used native classification may install its actual definition in
	// the same write. It must then be an ordinary stored Bead too.
	typeID := strings.TrimPrefix(taskType, scope)
	if definitions[typeID] == nil {
		definitions[typeID] = show(taskType)
	}
	var actualDefinition struct {
		ID       string `json:"id"`
		Revision string `json:"revision"`
	}
	if err := json.Unmarshal(definitions[typeID], &actualDefinition); err != nil || actualDefinition.ID != record.Type.URI || actualDefinition.Revision != record.Type.Revision {
		t.Fatalf("Task affiliation does not pin its actual installed definition: %v %s", err, definitions[typeID])
	}
	wanted := make(map[string][]byte, len(definitions)+1)
	for id, raw := range definitions {
		wanted[scope+id] = raw
	}
	wanted[record.ID] = task
	assertRows := func(rows []json.RawMessage, expected map[string][]byte) {
		t.Helper()
		if len(rows) != len(expected) {
			t.Fatalf("ordinary collection cardinality %d, want %d", len(rows), len(expected))
		}
		seen := map[string]bool{}
		for _, raw := range rows {
			var value struct {
				ID string `json:"id"`
			}
			if err := json.Unmarshal(raw, &value); err != nil {
				t.Fatal(err)
			}
			canonical, err := graph.CanonicalizeJSON(raw)
			if err != nil || seen[value.ID] || expected[value.ID] == nil || !bytes.Equal(canonical, expected[value.ID]) {
				t.Fatalf("ordinary path changed, duplicated or hid a complete record: %v %s", err, raw)
			}
			seen[value.ID] = true
		}
	}
	// Ordinary JSON and explicit canonical output are distinct existing CLI
	// contracts. Exercise both against the same complete persisted records.
	list := func(selectors ...string) []json.RawMessage {
		t.Helper()
		args := append([]string{"list", "--json", "--limit", "0"}, selectors...)
		out := call(t, 0, args...)
		var envelope struct {
			Result struct {
				Items   []json.RawMessage `json:"items"`
				HasMore *bool             `json:"hasMore"`
			} `json:"result"`
		}
		if err := json.Unmarshal([]byte(out), &envelope); err != nil || envelope.Result.Items == nil || envelope.Result.HasMore == nil || *envelope.Result.HasMore {
			t.Fatalf("ordinary list lost completeness: %v %s", err, out)
		}
		return envelope.Result.Items
	}
	canonicalList := func(selectors ...string) []json.RawMessage {
		t.Helper()
		args := append([]string{"list", "--output-format", "json", "--unpaged", "--limit", "0"}, selectors...)
		out := call(t, 0, args...)
		var envelope struct {
			Outcome string `json:"outcome"`
			Data    struct {
				Items []json.RawMessage `json:"items"`
				Next  json.RawMessage   `json:"next"`
			} `json:"data"`
			Completion struct {
				Returned  int  `json:"returned"`
				Exhausted bool `json:"exhausted"`
			} `json:"completion"`
		}
		if err := json.Unmarshal([]byte(out), &envelope); err != nil || envelope.Outcome != "success" || !envelope.Completion.Exhausted || envelope.Completion.Returned != len(envelope.Data.Items) || string(envelope.Data.Next) != "null" {
			t.Fatalf("canonical list lost completeness: %v %s", err, out)
		}
		return envelope.Data.Items
	}
	onlyDefinitions := make(map[string][]byte, len(definitions))
	for id, raw := range definitions {
		onlyDefinitions[scope+id] = raw
	}
	// Keep the canonical-only flag guard. Subsequent complete reads also
	// verify that this refusal changed no records.
	if out := call(t, 2, "list", "--json", "--unpaged", "--limit", "0"); strings.TrimSpace(out) != "" {
		t.Fatalf("invalid output combination exposed rows: %s", out)
	}
	for _, read := range []func(...string) []json.RawMessage{list, canonicalList} {
		assertRows(read(), wanted)
		assertRows(read("--type", "task"), map[string][]byte{record.ID: task})
		assertRows(read("--type", "type"), onlyDefinitions)
	}
	count(len(wanted))
	count(1, "--type", "task")
	count(len(definitions), "--type", "type")
	for _, tc := range []struct {
		args     []string
		expected map[string][]byte
	}{
		{[]string{"graph", "--all", "--json"}, wanted},
		{[]string{"graph", localID, "--json"}, map[string][]byte{record.ID: task}},
	} {
		out := call(t, 0, tc.args...)
		var envelope struct {
			Result struct {
				Beads     []json.RawMessage `json:"beads"`
				Traversal struct {
					Complete bool `json:"complete"`
				} `json:"traversal"`
			} `json:"result"`
		}
		if err := json.Unmarshal([]byte(out), &envelope); err != nil || !envelope.Result.Traversal.Complete {
			t.Fatalf("ordinary graph lost complete selected records: %v %s", err, out)
		}
		assertRows(envelope.Result.Beads, tc.expected)
	}
	if !bytes.Equal(show(record.ID), task) {
		t.Fatal("reopen changed the persisted Task or exact affiliation")
	}
}
