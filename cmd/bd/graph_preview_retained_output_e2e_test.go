//go:build cgo

package main

import (
	"encoding/json"
	"reflect"
	"testing"
)

// This process test covers command registration, legacy/canonical dispatch,
// root/Link alias wiring and the actual retained provider. Pure descriptor tests
// cover masks and exact-value preservation without opening any store.
func TestGraphCanonicalRetainedOutputCLI(t *testing.T) {
	isolateBeadsDirForTest(t)
	bd := buildBDUnderTest(t)
	work, home := t.TempDir(), t.TempDir()
	const scope = "https://example.invalid/canonical-retained/"
	call := func(args ...string) string {
		t.Helper()
		return graphVersionsOK(t, bd, work, home, args...)
	}
	call("init", "--graph-mode", "link", "--scope-url", scope, "--skip-hooks", "--skip-agents", "--non-interactive", "--json")
	call("create", "Source", "--id", "beads/source", "--json")
	call("create", "Target", "--id", "beads/target", "--json")
	call("link", "create", "beads/source", "beads/target", "--id", "links/edge", "--type", scope+"types/preview-related", "--json")
	call("link", "update", "links/edge", "--properties", `{"note":"second"}`, "--json")

	var legacy struct {
		SchemaVersion int             `json:"schemaVersion"`
		Result        json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal([]byte(call("link", "versions", "links/edge", "--json")), &legacy); err != nil || legacy.SchemaVersion != 1 {
		t.Fatalf("legacy contract drifted: %v %+v", err, legacy)
	}
	var canonical struct {
		SchemaVersion int             `json:"schema_version"`
		BDPSpec       string          `json:"bdp_spec"`
		ResultKind    string          `json:"result_kind"`
		Data          json.RawMessage `json:"data"`
	}
	decode := func(args ...string) string {
		t.Helper()
		out := call(args...)
		if err := json.Unmarshal([]byte(out), &canonical); err != nil {
			t.Fatal(err)
		}
		if canonical.SchemaVersion != 2 || canonical.BDPSpec == "" {
			t.Fatalf("canonical contract identity missing: %s", out)
		}
		return out
	}
	// The root commands also retain the native Bead result contract.
	decode("versions", "beads/source", "--output-format", "json")
	var beadVersions struct {
		Resource string `json:"resource"`
		Kind     string `json:"kind"`
		Versions []struct {
			Version string `json:"version"`
		} `json:"versions"`
	}
	if err := json.Unmarshal(canonical.Data, &beadVersions); err != nil || beadVersions.Resource != "beads/source" || len(beadVersions.Versions) != 1 {
		t.Fatalf("Bead version result changed: %v %s", err, canonical.Data)
	}
	beadToken := beadVersions.Versions[0].Version
	decode("compare", "beads/source", "--from", beadToken, "--to", beadToken, "--output-format", "json")
	if canonical.ResultKind != "comparison" {
		t.Fatalf("Bead comparison result kind: %s", canonical.ResultKind)
	}
	listed := decode("link", "versions", "links/edge", "--output-format", "json")
	if canonical.ResultKind != "version_collection" {
		t.Fatalf("wrong version result kind: %s", listed)
	}
	var oldValue, newValue any
	if err := json.Unmarshal(legacy.Result, &oldValue); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(canonical.Data, &newValue); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(oldValue, newValue) {
		t.Fatalf("native version payload rewritten: %s", listed)
	}
	for _, route := range [][]string{{"versions", "links/edge"}, {"history", "links/edge"}} {
		if got := call(append(route, "--output-format", "json")...); got != listed {
			t.Fatalf("alias differs: %v\n%s\n%s", route, got, listed)
		}
	}
	var versions struct {
		Versions []struct {
			Version string `json:"version"`
		} `json:"versions"`
	}
	if err := json.Unmarshal(canonical.Data, &versions); err != nil || len(versions.Versions) != 2 {
		t.Fatalf("version fixture: %v %s", err, canonical.Data)
	}
	from, to := versions.Versions[0].Version, versions.Versions[1].Version
	compared := decode("link", "compare", "links/edge", "--from", from, "--to", to, "--output-format", "json")
	if canonical.ResultKind != "comparison" {
		t.Fatalf("wrong comparison result kind: %s", compared)
	}
	if got := call("compare", "links/edge", "--from", from, "--to", to, "--output-format", "json"); got != compared {
		t.Fatalf("root/link compare differ: %s\n%s", got, compared)
	}
	var operands struct {
		From, To struct {
			Version string `json:"version"`
		}
	}
	if err := json.Unmarshal(canonical.Data, &operands); err != nil || operands.From.Version != from || operands.To.Version != to {
		t.Fatalf("operand direction changed: %v %s", err, canonical.Data)
	}

	for _, flags := range [][]string{{"--json"}, {"--format", "json"}, {"--json", "--format", "json"}} {
		args := append([]string{"link", "versions", "links/edge", "--output-format", "json"}, flags...)
		stdout, stderr, code := graphVersionsProcess(t, bd, work, home, args...)
		if code == 0 || stdout != "" || stderr == "" {
			t.Fatalf("ambiguous format accepted: %v code=%d out=%q err=%q", args, code, stdout, stderr)
		}
	}
	for _, args := range [][]string{
		{"versions", "missing", "--output-format", "json"},
		{"history", "missing", "--output-format", "json", "--limit", "1"},
		{"link", "versions", "beads/source", "--output-format", "json"},
		{"link", "compare", "beads/source", "--from", from, "--to", to, "--output-format", "json"},
		{"versions", "--output-format", "json"},
		{"history", "--output-format", "json"},
		{"link", "versions", "--output-format", "json"},
		{"compare", "--from", from, "--to", to, "--output-format", "json"},
		{"link", "compare", "--from", from, "--to", to, "--output-format", "json"},
	} {
		stdout, stderr, code := graphVersionsProcess(t, bd, work, home, args...)
		var failed struct {
			Outcome string `json:"outcome"`
			Error   struct {
				Origin string `json:"origin"`
			} `json:"error"`
		}
		if err := json.Unmarshal([]byte(stdout), &failed); err != nil || code == 0 || stderr == "" || failed.Outcome != "failure" || failed.Error.Origin != "cli" {
			t.Fatalf("read refusal lost structured failure: %v code=%d out=%q err=%q decode=%v", args, code, stdout, stderr, err)
		}
	}
	for _, args := range [][]string{
		{"versions", "missing", "--output-format", "json", "--omit", "/version"},
		{"versions", "missing", "--output-format", "jsonl"},
		{"compare", "missing", "--from", from, "--to", to, "--output-format", "json", "--include", "/properties"},
	} {
		stdout, stderr, code := graphVersionsProcess(t, bd, work, home, args...)
		if code != 2 || stdout != "" || stderr == "" {
			t.Fatalf("unsupported projection/format was not refused before read: %v code=%d out=%q err=%q", args, code, stdout, stderr)
		}
	}
	stdout, stderr, code := graphVersionsProcess(t, bd, work, home, "versions", "missing", "--output-format", "json", "--omit", "envelope")
	if code != 3 || stdout != "" || stderr == "" {
		t.Fatalf("raw failure wrote result or lost status: code=%d out=%q err=%q", code, stdout, stderr)
	}
	call("link", "delete", "links/edge", "--force", "--json")
	if afterDelete := call("link", "versions", "links/edge", "--output-format", "json"); afterDelete != listed {
		t.Fatalf("private deletion marker became a citable version: before=%s after=%s", listed, afterDelete)
	}
	// Admission fails before the provider opens, but must still use the chosen
	// result contract rather than a Resource-shaped failure or stderr alone.
	uninitialized := t.TempDir()
	for _, request := range []struct {
		args []string
		kind string
	}{
		{[]string{"versions", "missing"}, "version_collection"},
		{[]string{"link", "versions", "missing"}, "version_collection"},
		{[]string{"compare", "missing", "--from", "a", "--to", "b"}, "comparison"},
	} {
		stdout, stderr, code := graphVersionsProcess(t, bd, uninitialized, home, append(request.args, "--output-format", "json")...)
		var failed struct {
			Kind    string `json:"result_kind"`
			Outcome string `json:"outcome"`
		}
		if err := json.Unmarshal([]byte(stdout), &failed); err != nil || code == 0 || stderr == "" || failed.Outcome != "failure" || failed.Kind != request.kind {
			t.Fatalf("admission failure lost retained result contract: %v code=%d out=%q err=%q decode=%v", request.args, code, stdout, stderr, err)
		}
	}
}
