package canonicaledit

import (
	"strings"
	"testing"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/graphpatch"
)

func TestKeyedOperationsPreserveLiteralKeysAndOrder(t *testing.T) {
	raw, err := KeyedOperations([]string{"b=first", "a/b~c=3", "b=last=value", "gone=true"}, []string{"missing", "gone", "missing"})
	if err != nil {
		t.Fatal(err)
	}
	want := `[{"op":"add","path":"/b","value":"last=value"},{"op":"add","path":"/a~1b~0c","value":3},{"op":"add","path":"/gone","value":true},{"op":"add","path":"/missing","value":null},{"op":"remove","path":"/missing"},{"op":"add","path":"/gone","value":null},{"op":"remove","path":"/gone"}]`
	if string(raw) != want {
		t.Fatalf("got %s, want %s", raw, want)
	}
	patch, err := graphpatch.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	before, err := graph.NewProperties([]byte(`{"keep":{"nested":true},"gone":"old"}`))
	if err != nil {
		t.Fatal(err)
	}
	after, err := patch.Apply(before)
	if err != nil || string(after.Bytes()) != `{"a/b~c":3,"b":"last=value","keep":{"nested":true}}` {
		t.Fatalf("after=%s err=%v", after.Bytes(), err)
	}
	if !strings.Contains(string(before.Bytes()), `"gone":"old"`) {
		t.Fatal("modified predecessor")
	}
}

func TestScalarExactNumbersAndText(t *testing.T) {
	for _, tc := range []struct{ input, want string }{
		{"null", "null"}, {"true", "true"}, {"false", "false"}, {"17", "17"},
		{"hello=world", `"hello=world"`}, {"[1,2]", `"[1,2]"`}, {"01", `"01"`},
	} {
		got, err := Scalar(tc.input)
		if err != nil || string(got) != tc.want {
			t.Errorf("Scalar(%q) = %s, %v, want %s", tc.input, got, err, tc.want)
		}
	}
	for _, value := range []string{"1e999", "\xff"} {
		if _, err := Scalar(value); err == nil {
			t.Errorf("accepted %q", value)
		}
	}
	if _, err := KeyedOperations([]string{"missing-equals"}, nil); err == nil {
		t.Fatal("accepted malformed assignment")
	}
	if _, err := KeyedOperations([]string{"k=1e999", "k=valid"}, nil); err == nil {
		t.Fatal("overwritten invalid number escaped admission")
	}
}

func TestMergeOperationsAreShallowAndCodeUnitOrdered(t *testing.T) {
	raw, err := MergeOperations([]byte(`{"\uE000":1,"\uD800\uDC00":2,"nested":{"keep":"replacement"}}`))
	if err != nil {
		t.Fatal(err)
	}
	want := `[{"op":"add","path":"/nested","value":{"keep":"replacement"}},{"op":"add","path":"/𐀀","value":2},{"op":"add","path":"/","value":1}]`
	if string(raw) != want {
		t.Fatalf("merge order = %s, want %s", raw, want)
	}
	raw, err = MergeOperations([]byte(`{}`))
	if err != nil || string(raw) != "[]" {
		t.Fatalf("empty merge = %s %v", raw, err)
	}
	if _, err := MergeOperations([]byte(`{"x":1,"x":2}`)); err == nil {
		t.Fatal("merge collapsed duplicate keys")
	}
}
