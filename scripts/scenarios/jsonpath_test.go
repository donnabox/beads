package main

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func mustDecode(t *testing.T, s string) any {
	t.Helper()
	dec := json.NewDecoder(strings.NewReader(s))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		t.Fatalf("decode %q: %v", s, err)
	}
	return v
}

func TestJSONPathEval(t *testing.T) {
	doc := mustDecode(t, `{
		"a": {"b": [{"c": "x"}, {"c": "y"}]},
		"items": [{"id": "i1"}, {"id": "i2"}, {"id": "i3"}],
		"ids": ["p", "q"],
		"n": 3
	}`)
	tests := []struct {
		path string
		want []any
	}{
		{"$.a.b[0].c", []any{"x"}},
		{"$.a.b[1].c", []any{"y"}},
		{"$.items[*].id", []any{"i1", "i2", "i3"}},
		{"$.items[1].id", []any{"i2"}},
		{"$.n", []any{json.Number("3")}},
		{"$.ids", []any{[]any{"p", "q"}}},
	}
	for _, tc := range tests {
		p, err := parseJSONPath(tc.path)
		if err != nil {
			t.Errorf("parseJSONPath(%q): %v", tc.path, err)
			continue
		}
		got, err := p.eval(doc)
		if err != nil {
			t.Errorf("eval(%q): %v", tc.path, err)
			continue
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("eval(%q) = %#v, want %#v", tc.path, got, tc.want)
		}
	}
}

func TestJSONPathWildcardFlag(t *testing.T) {
	for path, want := range map[string]bool{"$.a.b": false, "$.items[*].id": true, "$.items[0].id": false} {
		p, err := parseJSONPath(path)
		if err != nil {
			t.Fatalf("parseJSONPath(%q): %v", path, err)
		}
		if p.hasWildcard() != want {
			t.Errorf("hasWildcard(%q) = %v, want %v", path, p.hasWildcard(), want)
		}
	}
}

// Only $.a.b[0].c and the [*] projection are supported; everything else must
// be refused when the scenario is loaded, never quietly evaluated differently.
func TestJSONPathRejectsUnsupportedSyntax(t *testing.T) {
	for _, path := range []string{
		"",
		"$",
		"a.b",
		"$.",
		"$..a",
		"$.a..b",
		"$.a[?(@.x)]",
		"$.a[0:2]",
		"$.a[-1]",
		"$['a']",
		"$.a[*].b[*].c",
		"$.a b",
		"$.[0]",
		"$.a[x]",
		"$.a[]",
		"$.a[0",
		"$.a.*",
	} {
		if _, err := parseJSONPath(path); err == nil {
			t.Errorf("parseJSONPath(%q) accepted unsupported syntax", path)
		}
	}
}

func TestJSONPathEvalMissing(t *testing.T) {
	doc := mustDecode(t, `{"a": {"b": [{"c": "x"}]}, "items": [{"id": "i1"}, {"other": 1}], "n": 3}`)
	for _, path := range []string{
		"$.zzz",
		"$.a.zzz",
		"$.a.b[5].c",
		"$.n.x",
		"$.a.b.c",
		"$.items[*].id", // one element has no id: an incomplete projection is an error, not a shorter list
		"$.n[0]",
	} {
		p, err := parseJSONPath(path)
		if err != nil {
			t.Fatalf("parseJSONPath(%q): %v", path, err)
		}
		if got, err := p.eval(doc); err == nil {
			t.Errorf("eval(%q) = %#v, want an error", path, got)
		}
	}
}
