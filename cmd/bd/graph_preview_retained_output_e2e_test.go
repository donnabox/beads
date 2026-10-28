//go:build cgo

package main

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// This process test covers command registration, legacy/canonical dispatch,
// root/Link alias wiring and the actual retained provider. Pure descriptor tests
// cover masks and exact-value preservation without opening any store.
func TestGraphCanonicalRetainedOutputCLI(t *testing.T) {
	isolateBeadsDirForTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) { graphCanonicalRetainedOutputCLI(t, engine) })
	}
}

func graphCanonicalRetainedOutputCLI(t *testing.T, engine string) {
	t.Helper()
	bd := buildBDUnderTest(t)
	const scope = "https://example.invalid/canonical-retained/"
	work, home := graphListWorkspace(t, bd, engine, scope)
	call := func(args ...string) string {
		t.Helper()
		return graphVersionsOK(t, bd, work, home, args...)
	}
	call("remember", "Source", "--id", "beads/source", "--json")
	call("create", "Target", "--id", "beads/target", "--type", "constraint", "--json")
	call("link", "create", "beads/source", "beads/target", "--id", "links/edge", "--type", graphstore.RelatedTypeURL(scope), "--json")
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
	for _, route := range [][]string{{"link", "versions", "links/edge"}, {"versions", "links/edge"}, {"history", "links/edge"}} {
		for _, raw := range []bool{false, true} {
			args := append(append([]string(nil), route...), "--output-format", "jsonl", "--include", "/local_revision")
			if raw {
				args = append(args, "--omit", "envelope")
			}
			lines := strings.Split(strings.TrimSpace(call(args...)), "\n")
			want := 4 // header, two versions, terminal
			if raw {
				want = 2
			}
			if len(lines) != want {
				t.Fatalf("%v: got %d JSONL records: %v", args, len(lines), lines)
			}
			for i := range versions.Versions {
				line := lines[i]
				if !raw {
					line = lines[i+1]
					var wrapper struct {
						RecordType string          `json:"record_type"`
						Data       json.RawMessage `json:"data"`
					}
					if err := json.Unmarshal([]byte(line), &wrapper); err != nil || wrapper.RecordType != "item" {
						t.Fatalf("%v: invalid item wrapper: %v %s", args, err, line)
					}
					line = string(wrapper.Data)
				}
				var item struct{ Resource, Kind, Version string }
				if err := json.Unmarshal([]byte(line), &item); err != nil || item.Resource != "links/edge" || item.Kind != "link" || item.Version != versions.Versions[i].Version {
					t.Fatalf("%v: item lost subject, kind or revision: %v %s", args, err, line)
				}
			}
		}
	}
	for _, flags := range [][]string{{}, {"--omit", "envelope"}} {
		args := append([]string{"history", "missing", "--output-format", "jsonl"}, flags...)
		out, stderr, code := graphVersionsProcess(t, bd, work, home, args...)
		if code == 0 || stderr == "" {
			t.Fatalf("%v: missing subject accepted: %q %q", args, out, stderr)
		}
		lines := strings.Split(strings.TrimSpace(out), "\n")
		if len(flags) == 0 && (len(lines) != 2 || !strings.Contains(lines[1], `"record_type":"terminal"`) || !strings.Contains(lines[1], `"outcome":"failure"`)) {
			t.Fatalf("%v: missing failed terminal: %q", args, out)
		}
		if len(flags) != 0 && out != "" {
			t.Fatalf("%v: raw stream emitted error data: %q", args, out)
		}
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

	for _, route := range [][]string{{"compare", "links/edge"}, {"link", "compare", "links/edge"}} {
		base := append(append([]string(nil), route...), "--from", from, "--to", to)
		repeated := append(append([]string(nil), base...), "--from="+from, "--to="+to, "--output-format", "json")
		if got := call(repeated...); got != compared {
			t.Fatalf("identical repeated comparison operands changed result: %v\n%s\n%s", repeated, got, compared)
		}
		for _, conflict := range []string{"--from=" + to, "--to=" + from} {
			args := append(append([]string(nil), base...), conflict, "--output-format", "json")
			stdout, stderr, code := graphVersionsProcess(t, bd, work, home, args...)
			var failed struct {
				Kind    string `json:"result_kind"`
				Outcome string `json:"outcome"`
				Error   struct {
					Problem struct {
						Code string `json:"code"`
					} `json:"problem"`
				} `json:"error"`
			}
			if err := json.Unmarshal([]byte(stdout), &failed); err != nil || code != 2 || stderr == "" || failed.Kind != "comparison" || failed.Outcome != "failure" || failed.Error.Problem.Code != "invalid_selector" {
				t.Fatalf("conflicting operands lost canonical failure contract: %v code=%d out=%q err=%q decode=%v", args, code, stdout, stderr, err)
			}
		}
		// Canonical occurrence checks must not change legacy last-value parsing.
		legacyBase := append(append([]string(nil), base...), "--json")
		legacyRepeated := append(append([]string(nil), route...), "--from", "unknown", "--to", "unknown", "--from", from, "--to", to, "--json")
		if got, want := call(legacyRepeated...), call(legacyBase...); got != want {
			t.Fatalf("legacy repeated operand behavior changed: %v\n%s\n%s", route, got, want)
		}
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
		{"compare", "missing", "--from", from, "--to", to, "--output-format", "json", "--include", "/properties"},
	} {
		stdout, stderr, code := graphVersionsProcess(t, bd, work, home, args...)
		if code != 2 || stderr == "" {
			t.Fatalf("unsupported projection/format was not refused before read: %v code=%d out=%q err=%q", args, code, stdout, stderr)
		}
		var failed struct {
			Outcome string `json:"outcome"`
			Error   struct {
				Problem struct {
					Code string `json:"code"`
				} `json:"problem"`
			} `json:"error"`
		}
		if err := json.Unmarshal([]byte(stdout), &failed); err != nil || failed.Outcome != "failure" || failed.Error.Problem.Code != "invalid_selector" {
			t.Fatalf("invalid mask lost JSON failure contract: %v code=%d out=%q err=%q decode=%v", args, code, stdout, stderr, err)
		}
	}
	stdout, stderr, code := graphVersionsProcess(t, bd, work, home, "versions", "missing", "--output-format", "json", "--omit", "envelope")
	if code != 3 || stdout != "" || stderr == "" {
		t.Fatalf("raw failure wrote result or lost status: code=%d out=%q err=%q", code, stdout, stderr)
	}
	removedLink := graphMixedResult[graphstore.LinkDeleteResult](t, call("link", "delete", "links/edge", "--force", "--json"))
	var beforeLink, afterLink struct {
		Data struct {
			Versions     []map[string]any                  `json:"versions"`
			CurrentState *graphstore.ResourceTerminalState `json:"currentState"`
			Terminals    []map[string]any                  `json:"terminal_versions"`
		} `json:"data"`
	}
	afterDelete := call("link", "versions", "links/edge", "--output-format", "json")
	if err := json.Unmarshal([]byte(listed), &beforeLink); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(afterDelete), &afterLink); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(beforeLink.Data.Versions, afterLink.Data.Versions) || afterLink.Data.CurrentState == nil || afterLink.Data.CurrentState.Revision != removedLink.Link.Revision || len(afterLink.Data.Terminals) != 1 {
		t.Fatalf("Link deletion context missing or became citable: before=%s after=%s", listed, afterDelete)
	}
	linkStream, linkStderr, linkCode := graphVersionsProcess(t, bd, work, home, "link", "versions", "links/edge", "--output-format", "jsonl")
	if linkCode != 5 || linkStderr == "" || strings.Contains(linkStream, `"record_type":"item"`) || !strings.Contains(linkStream, `"record_type":"terminal"`) {
		t.Fatalf("Link terminal context was silently dropped from JSONL: code=%d out=%s err=%s", linkCode, linkStream, linkStderr)
	}
	// The merged History adapter must preserve real deletion context in JSON,
	// and refuse JSONL before emitting items it cannot interpret independently.
	original := graphMixedResult[graphstore.Record](t, call("remember", "Retained context", "--id", "beads/context", "--json"))
	deleted := graphMixedResult[graphstore.MemoryDeleteResult](t, call("delete", "beads/context", "--force", "--json"))
	if deleted.Terminal == nil {
		t.Fatal("context fixture did not mint a terminal state")
	}
	assertHistoryContext := func(wantCurrent *graphstore.MemoryTerminalState, wantVersions int) {
		t.Helper()
		decode("versions", "beads/context", "--output-format", "json")
		var history struct {
			CurrentState *graphstore.MemoryTerminalState `json:"currentState"`
			Terminals    []map[string]any                `json:"terminal_versions"`
			Versions     []map[string]any                `json:"versions"`
		}
		// The comparison oracle uses private graphstore attribution vocabulary;
		// normalize only that carried envelope as other workflow oracles do.
		projected, err := graphMixedStoredAttribution(canonical.Data)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(projected, &history); err != nil || !reflect.DeepEqual(history.CurrentState, wantCurrent) || len(history.Terminals) != 1 || history.Terminals[0]["version"] != deleted.Terminal.Revision || len(history.Versions) != wantVersions {
			t.Fatalf("canonical JSON lost native history context: %v %s", err, canonical.Data)
		}
		for _, route := range []string{"versions", "history"} {
			for _, raw := range []bool{false, true} {
				args := []string{route, "beads/context", "--output-format", "jsonl"}
				if raw {
					args = append(args, "--omit", "envelope")
				}
				stdout, stderr, code := graphVersionsProcess(t, bd, work, home, args...)
				if code != 5 || stderr == "" {
					t.Fatalf("native context was not refused: %v code=%d out=%q err=%q", args, code, stdout, stderr)
				}
				if raw {
					if stdout != "" {
						t.Fatalf("raw context refusal emitted bytes: %v %q", args, stdout)
					}
					continue
				}
				lines := strings.Split(strings.TrimSpace(stdout), "\n")
				if len(lines) != 2 || !strings.Contains(lines[0], `"record_type":"header"`) {
					t.Fatalf("context refusal emitted items or lost header: %v %q", args, stdout)
				}
				var terminal struct {
					RecordType string `json:"record_type"`
					Outcome    string `json:"outcome"`
					Error      struct {
						Problem struct {
							Code string `json:"code"`
						} `json:"problem"`
					} `json:"error"`
				}
				if err := json.Unmarshal([]byte(lines[1]), &terminal); err != nil || terminal.RecordType != "terminal" || terminal.Outcome != "failure" || terminal.Error.Problem.Code != "capability_unavailable" {
					t.Fatalf("context refusal lost failure terminal: %v %v %q", args, err, stdout)
				}
			}
		}
	}
	assertHistoryContext(deleted.Terminal, 1)
	call("update", "beads/context", "--from-revision", "="+original.Revision, "--if-revision", "="+deleted.Terminal.Revision, "--json")
	// Recovery removes currentState but must not drop the retained terminal.
	assertHistoryContext(nil, 2)
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
