package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func memoryLinkTestCommand(t *testing.T, op memoryLinkOperation, flags ...string) *cobra.Command {
	t.Helper()
	c := newMemoryLinkCommand(op)
	root := &cobra.Command{Use: "bd"}
	root.PersistentFlags().Bool("json", false, "")
	root.PersistentFlags().String("format", "", "")
	parent := &cobra.Command{Use: "memory", Annotations: memoryGraphCmd.Annotations}
	root.AddCommand(parent)
	parent.AddCommand(c)
	t.Cleanup(func() { delete(canonicalFailureFactories, c) })
	if err := c.ParseFlags(flags); err != nil {
		t.Fatal(err)
	}
	return c
}

func TestMemoryLinkInput(t *testing.T) {
	const scope = "https://example.invalid/memory-create/"
	for _, op := range memoryLinkOperations {
		c := memoryLinkTestCommand(t, op, "--"+op.targetFlag+"=b", "--id=relationship", "--if-source-revision=owner-now")
		r, err := memoryLinkRequest(c, []string{"a"}, scope, op)
		if err != nil || r.SourcePath != "beads/a" || r.TargetPath != "beads/b" || r.Path != "links/relationship" || r.TypeURL != scope+"types/"+op.family || r.ExpectedSourceRevision != "owner-now" || r.UnconditionalSource {
			t.Fatalf("%s: %+v %v", op.name, r, err)
		}
	}
	op := memoryLinkOperations[0]
	for _, tc := range []struct{ flags, args []string }{
		{}, {flags: []string{"--to=b"}}, {flags: []string{"--to=b"}, args: []string{"a", "other"}},
		{flags: []string{"--to="}, args: []string{"a"}}, {flags: []string{"--to=links/b"}, args: []string{"a"}},
		{flags: []string{"--to=https://foreign.invalid/beads/b"}, args: []string{"a"}},
		{flags: []string{"--to=b@old", "--target-revision=other"}, args: []string{"a"}},
		{flags: []string{"--to=b", "--id=beads/wrong"}, args: []string{"a"}},
	} {
		c := memoryLinkTestCommand(t, op, tc.flags...)
		r, err := memoryLinkRequest(c, tc.args, scope, op)
		if err == nil || !reflect.DeepEqual(r, graphstore.LinkCreateRequest{}) {
			t.Fatalf("invalid %v: %+v %v", tc, r, err)
		}
	}
	c := memoryLinkTestCommand(t, op, "--to=b@current")
	r, err := memoryLinkRequest(c, []string{"a"}, scope, op)
	if err != nil || !r.UnconditionalSource || r.Path != "" {
		t.Fatalf("shared allocation/current intent: %+v %v", r, err)
	}
}

func TestMemoryLinkProperties(t *testing.T) {
	op := memoryLinkOperations[0]
	for _, tc := range []struct {
		flags []string
		want  map[string]any
		bad   bool
	}{
		{want: map[string]any{}},
		{flags: []string{"--title=Background", "--body=Reason", "--status=reviewed"}, want: map[string]any{"title": "Background", "body": "Reason", "status": "reviewed"}},
		{flags: []string{"--properties={\"title\":\"From JSON\",\"extra\":true}", "--status=active"}, want: map[string]any{"title": "From JSON", "extra": true, "status": "active"}},
		{flags: []string{"--properties={\"title\":\"JSON\"}", "--title=Flag"}, bad: true},
		{flags: []string{"--properties={\"status\":null}", "--status=active"}, bad: true},
		{flags: []string{"--properties={\"x\":1,\"x\":2}"}, bad: true},
	} {
		c := memoryLinkTestCommand(t, op, tc.flags...)
		got, err := memoryLinkProperties(c)
		if (err != nil) != tc.bad || (!tc.bad && !reflect.DeepEqual(got, tc.want)) {
			t.Fatalf("%v: %+v %v", tc.flags, got, err)
		}
	}
	c := memoryLinkTestCommand(t, op, "--properties=@-")
	c.SetIn(strings.NewReader(`{"body":"stdin"}`))
	if got, err := memoryLinkProperties(c); err != nil || got["body"] != "stdin" {
		t.Fatalf("shared properties reader: %+v %v", got, err)
	}
}

func TestMemoryLinkPropertiesCanonicalNumbersAndFiles(t *testing.T) {
	op := memoryLinkOperations[0]
	raw := `{"n":9007199254740994,"nested":[1.25,1e-7],"unicode":"記憶"}`
	file := filepath.Join(t.TempDir(), "properties.json")
	if err := os.WriteFile(file, []byte(raw), 0600); err != nil {
		t.Fatal(err)
	}
	for _, input := range []string{raw, "@" + file, "@-"} {
		c := memoryLinkTestCommand(t, op, "--properties="+input)
		c.SetIn(strings.NewReader(raw))
		got, err := memoryLinkProperties(c)
		if err != nil || got["n"] != json.Number("9007199254740994") || got["nested"].([]any)[0] != json.Number("1.25") {
			t.Fatalf("%s: %#v %v", input, got, err)
		}
	}
	for _, raw := range []string{`{"n":9007199254740993}`, `{"n":1} {}`, `{"n":"\ud800"}`, `[]`, `null`, `{"nested":{"x":1,"x":2}}`} {
		c := memoryLinkTestCommand(t, op, "--properties="+raw)
		if got, err := memoryLinkProperties(c); err == nil || got != nil {
			t.Fatalf("accepted invalid canonical properties %s: %#v %v", raw, got, err)
		}
	}
}

func TestMemoryLinkHoistedPropertiesRejectInvalidUTF8(t *testing.T) {
	for _, name := range []string{"title", "body", "status"} {
		c := memoryLinkTestCommand(t, memoryLinkOperations[0], "--"+name+"="+string([]byte{0xff}))
		if got, err := memoryLinkProperties(c); err == nil || got != nil {
			t.Fatalf("--%s silently repaired invalid UTF-8: %#v %v", name, got, err)
		}
	}
}
func TestMemoryLinkCanonicalBoundary(t *testing.T) {
	for _, tc := range []struct {
		flags               []string
		refuse, fallbackNil bool
	}{
		{}, {flags: []string{"--json"}}, {flags: []string{"--output-format=json"}},
		{flags: []string{"--json", "--output-format=json"}},
		{flags: []string{"--output-format=detail"}},
		{flags: []string{"--json", "--omit=envelope"}},
		{flags: []string{"--include=/link/id"}, refuse: true},
		{flags: []string{"--json", "--omit=/link"}, refuse: true},
		{flags: []string{"--omit=envelope"}, refuse: true, fallbackNil: true},
		{flags: []string{"--output-format=tree"}, refuse: true},
		{flags: []string{"--json", "--output-format=tree"}, refuse: true},
		{flags: []string{"--json", "--output-format=detail"}, refuse: true},
		{flags: []string{"--output-format=json", "--output-format=detail"}, refuse: true},
		{flags: []string{"--format=json"}, refuse: true, fallbackNil: true},
	} {
		c := memoryLinkTestCommand(t, memoryLinkOperations[0], tc.flags...)
		opts, selected, err := canonicalOutputOptions(c)
		if err == nil {
			_, err = newMemoryMutationRenderer(opts)
		}
		if !selected || (err != nil) != tc.refuse {
			t.Fatalf("%v selected=%v err=%v", tc.flags, selected, err)
		}
		if tc.refuse && (canonicalFailureRendererFor(c) == nil) != tc.fallbackNil {
			t.Fatalf("wrong failure renderer boundary %v", tc.flags)
		}
	}
	for _, op := range memoryLinkOperations {
		cmd, _, err := rootCmd.Find([]string{"memory", op.name})
		if err != nil || cmd == nil || cmd.Name() != op.name || !canonicalOutputDefault(cmd) || cmd.InheritedFlags().Lookup("json") == nil {
			t.Fatalf("registration lost contract: %s %v", op.name, err)
		}
	}
}

func TestMemoryLinkSourceGuardAdmission(t *testing.T) {
	for _, guard := range []string{string([]byte{0xff}), strings.Repeat("x", graphstore.PreviewVersionTokenLimit+1)} {
		c := memoryLinkTestCommand(t, memoryLinkOperations[0], "--to=b", "--if-source-revision="+guard)
		got, err := memoryLinkRequest(c, []string{"a"}, "https://example.invalid/memory-create/", memoryLinkOperations[0])
		if err == nil || !reflect.DeepEqual(got, graphstore.LinkCreateRequest{}) {
			t.Fatalf("invalid guard admitted: %#v %v", got, err)
		}
	}
	c := memoryLinkTestCommand(t, memoryLinkOperations[0], "--to=b", "--if-source-revision="+strings.Repeat("x", graphstore.PreviewVersionTokenLimit))
	if _, err := memoryLinkRequest(c, []string{"a"}, "https://example.invalid/memory-create/", memoryLinkOperations[0]); err != nil {
		t.Fatalf("valid maximum-size opaque guard refused: %v", err)
	}
}
