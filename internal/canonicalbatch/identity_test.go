package canonicalbatch_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/canonicalbatch"
)

func compileDecision(t *testing.T, input string) (*canonicalbatch.Batch, []byte) {
	t.Helper()
	b, err := canonicalbatch.Compile([]byte(input), canonicalbatch.Options{Scope: scope})
	if err != nil {
		t.Fatal(err)
	}
	identity, err := b.DecisionJSON()
	if err != nil {
		t.Fatal(err)
	}
	return b, identity
}

func TestDecisionIdentityCaptureRenaming(t *testing.T) {
	input := strings.Join([]string{
		`{"operation":"createBead","name":"first","id":"one","type":"note"}`,
		`{"operation":"createBead","name":"second","id":"two","type":"note"}`,
		`{"operation":"createLink","name":"edge","type":"references","source":{"binding":"first"},"target":{"binding":"second"}}`,
		`{"operation":"select","name":"picked","query":{"ir":"bdp-query-design-1","profile":"query","selection":{"op":"refs","collection":"beads","references":{"binding":"first"}}}}`,
		`{"operation":"select","name":"again","query":{"ir":"bdp-query-design-1","profile":"query","selection":{"op":"refs","collection":"beads","references":{"binding":"picked"}}}}`,
		`{"operation":"updateBead","bead":{"binding":"again"},"propertiesChange":[{"op":"add","path":"/done","value":true}]}`,
		`{"operation":"deleteLink","link":{"binding":"edge"}}`,
	}, "\n")
	_, want := compileDecision(t, input)
	// Every batch-capture definition and use changes, including refs sources and
	// a set-valued mutation target. No labels remain in request identity.
	renamed := strings.NewReplacer(`"first"`, `"alpha"`, `"second"`, `"beta"`, `"edge"`, `"gamma"`, `"picked"`, `"delta"`, `"again"`, `"epsilon"`).Replace(input)
	b, got := compileDecision(t, renamed)
	if !bytes.Equal(want, got) {
		t.Fatalf("capture renaming changed identity:\n%s\n%s", want, got)
	}
	if !bytes.Contains(b.NativeJSON(), []byte(`"name":"alpha"`)) || bytes.Contains(got, []byte(`"name":`)) {
		t.Fatalf("presentation labels not separated: native=%s decision=%s", b.NativeJSON(), got)
	}
	// Reusing the old label on a different creator must not make the two
	// sources equivalent: references identify operation slots, not spellings.
	swapped := strings.Replace(input, `"source":{"binding":"first"}`, `"source":{"binding":"second"}`, 1)
	_, changed := compileDecision(t, swapped)
	if bytes.Equal(want, changed) {
		t.Fatal("different creator slot aliased the original decision")
	}
}

func TestDecisionIdentityUnusedCapture(t *testing.T) {
	plain := `{"operation":"createBead","type":"note","properties":{"title":"kept"}}`
	_, want := compileDecision(t, plain)
	for _, name := range []string{"unused", "other"} {
		_, got := compileDecision(t, strings.Replace(plain, `"type":`, `"name":"`+name+`","type":`, 1))
		if !bytes.Equal(want, got) {
			t.Fatalf("unused capture %q changed identity", name)
		}
	}
}

func TestDecisionIdentityLiteralDataAndSourceArePreserved(t *testing.T) {
	input := strings.Join([]string{
		`{"operation":"createBead","name":"capture","type":"note","properties":{"binding":"capture","number":9007199254740992},"metadata":{"name":"capture"}}`,
		`{"operation":"select","name":"picked","query":{"ir":"bdp-query-design-1","profile":"list","selection":{"op":"filter","input":{"op":"refs","collection":"beads","references":{"binding":"capture"}},"bind":"row","predicate":{"op":"compare","operator":"eq","left":{"op":"field","of":{"op":"binding","name":"row"},"path":["properties","name"]},"right":{"op":"literal","value":{"binding":"capture","op":"refs","number":9007199254740992}}}}}}`,
		`{"operation":"updateBead","bead":{"binding":"picked"},"metadataChange":[{"op":"add","path":"/capture","value":{"binding":"capture"}}],"revisionProvenance":{"message":"capture"}}`,
	}, "\n")
	b, identity := compileDecision(t, input)
	before := b.NativeJSON()
	if bytes.Count(identity, []byte(`"binding":"capture"`)) != 3 || !bytes.Contains(identity, []byte(`"name":"row"`)) || bytes.Count(identity, []byte(`9007199254740992`)) != 2 {
		t.Fatalf("literal data or exact numbers changed: %s", identity)
	}
	// Rename capture positions only; leave every ordinary value byte-for-byte.
	renamed := strings.NewReplacer(`"name":"capture"`, `"name":"renamed"`, `"references":{"binding":"capture"}`, `"references":{"binding":"renamed"}`).Replace(input)
	// The first replacement also changed metadata.name; undo that ordinary value.
	renamed = strings.Replace(renamed, `"metadata":{"name":"renamed"}`, `"metadata":{"name":"capture"}`, 1)
	_, same := compileDecision(t, renamed)
	if !bytes.Equal(identity, same) {
		t.Fatalf("renaming rewrote literal data:\n%s\n%s", identity, same)
	}
	identity[0] = 'x'
	if !bytes.Equal(before, b.NativeJSON()) || b.Members()[0].Name != "capture" || b.Members()[1].Name != "picked" {
		t.Fatal("identity normalization mutated compiled input")
	}
	_, original := compileDecision(t, input)
	again, err := b.DecisionJSON()
	if err != nil || !bytes.Equal(original, again) {
		t.Fatalf("identity returned aliased memory: %s %v", again, err)
	}
}

func TestDecisionIdentityPreservesWriteSemantics(t *testing.T) {
	base := `{"operation":"updateBead","bead":"one","type":{"uri":"https://work.example/acme/note","revision":"type-r1"},"propertiesChange":[{"op":"add","path":"/first","value":1},{"op":"add","path":"/second","value":2}],"metadataChange":[],"expectedRevision":"r1","revisionProvenance":{"message":"why"}}`
	_, want := compileDecision(t, base)
	for name, changed := range map[string]string{
		"literal identity": strings.Replace(base, `"bead":"one"`, `"bead":"two"`, 1),
		"guard":            strings.Replace(base, `"expectedRevision":"r1"`, `"expectedRevision":"r2"`, 1),
		"type pin":         strings.Replace(base, `"revision":"type-r1"`, `"revision":"type-r2"`, 1),
		"edit value":       strings.Replace(base, `"value":1`, `"value":2`, 1),
		"patch ordering":   strings.Replace(base, `{"op":"add","path":"/first","value":1},{"op":"add","path":"/second","value":2}`, `{"op":"add","path":"/second","value":2},{"op":"add","path":"/first","value":1}`, 1),
		"presence":         strings.Replace(base, `,"metadataChange":[]`, ``, 1),
		"replacement":      strings.Replace(base, `"metadataChange":[]`, `"metadata":{}`, 1),
		"provenance":       strings.Replace(base, `"message":"why"`, `"message":"because"`, 1),
	} {
		t.Run(name, func(t *testing.T) {
			_, got := compileDecision(t, changed)
			if bytes.Equal(want, got) {
				t.Fatal("changed write semantics retained the same identity")
			}
		})
	}
	// Property/object member order follows the same canonical normalization as
	// ordinary compilation; unlike patch order it is not execution semantics.
	_, reordered := compileDecision(t, strings.Replace(base, `"op":"add","path":"/first","value":1`, `"value":1,"path":"/first","op":"add"`, 1))
	if !bytes.Equal(want, reordered) {
		t.Fatal("object member ordering changed identity")
	}
}

func TestDecisionIdentityRequiresCompiledBatch(t *testing.T) {
	for _, b := range []*canonicalbatch.Batch{nil, {}} {
		if raw, err := b.DecisionJSON(); raw != nil || !errors.Is(err, canonicalbatch.ErrInvalid) {
			t.Fatalf("zero batch decision: %s %v", raw, err)
		}
	}
}
