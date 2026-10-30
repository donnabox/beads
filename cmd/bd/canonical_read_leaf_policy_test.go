package main

import (
	"errors"
	"reflect"
	"testing"
)

func TestCanonicalReadLeafExactCoverageAndIsolation(t *testing.T) {
	for _, tc := range []struct {
		path        []string
		op          canonicalReadLeafOperation
		positionals int
	}{
		{[]string{"list"}, canonicalReadLeafListBeads, 0},
		{[]string{"link", "list"}, canonicalReadLeafListLinks, 2},
		{[]string{"query"}, canonicalReadLeafQuery, 0},
	} {
		first, ok := lookupCanonicalReadLeaf(tc.path)
		if !ok || first.Operation != tc.op || first.MaximumPositionals != tc.positionals {
			t.Fatalf("wrong descriptor: %#v %v", first, ok)
		}
		first.Flags[0] = "compat-only"
		first.Path[0] = "changed"
		again, ok := lookupCanonicalReadLeaf(tc.path)
		if !ok || again.Flags[0] == "compat-only" || !reflect.DeepEqual(again.Path, tc.path) {
			t.Fatalf("descriptor state leaked: %#v", again)
		}
	}
	for _, path := range [][]string{{"ls"}, {"link"}, {"link", "ls"}, {"memory", "query"}, {"query", "nested"}, {"link list"}, {"link\x00list"}, nil} {
		if _, ok := lookupCanonicalReadLeaf(path); ok {
			t.Fatalf("uncovered/alias path admitted: %q", path)
		}
	}
}

func TestCanonicalReadLeafCompatDoesNotDecorateOrReinterpret(t *testing.T) {
	result, err := admitCanonicalReadLeaf(canonicalReadLeafInvocation{Profile: cliSurfaceCompat, Path: []string{"unknown"}, AliasUsed: true, Positionals: 100, Flags: []canonicalReadExplicitFlag{{Name: "status", Shorthand: "s", Inherited: true}}})
	if err != nil || result.Canonical || result.Descriptor.Operation != "" {
		t.Fatalf("compat was changed: %#v %v", result, err)
	}
}

func TestCanonicalReadLeafRejectsExplicitCompatFlagsIndependentlyOfValueOrigin(t *testing.T) {
	for _, use := range []canonicalReadExplicitFlag{
		{Name: "status"}, // includes --status=false/empty/default: explicitness is presence
		{Name: "status", Shorthand: "s"},
		{Name: "status", Inherited: true},
		{Name: "status", Shorthand: "s", Inherited: true},
		{Name: "format"},
		{Name: "conforms-to"},
	} {
		result, err := admitCanonicalReadLeaf(canonicalReadLeafInvocation{Profile: cliSurfaceCanonical, Path: []string{"list"}, Flags: []canonicalReadExplicitFlag{use}})
		var refusal *canonicalReadLeafRefusal
		if !errors.As(err, &refusal) || refusal.Code != "invalid_selector" || refusal.ExitCode != 2 || result.Canonical {
			t.Fatalf("compat flag silently accepted: %#v %#v %v", use, result, err)
		}
	}
}

func TestCanonicalReadLeafRequiresQueryASTAndLeavesItsContentsToCheckedReader(t *testing.T) {
	input := canonicalReadLeafInvocation{Profile: cliSurfaceCanonical, Path: []string{"query"}}
	if _, err := admitCanonicalReadLeaf(input); err == nil {
		t.Fatal("legacy expression fallback permitted")
	}
	input.Flags = []canonicalReadExplicitFlag{{Name: "ast"}}
	result, err := admitCanonicalReadLeaf(input)
	if err != nil || !result.Canonical || result.Descriptor.Operation != canonicalReadLeafQuery {
		t.Fatalf("AST surface refused: %#v %v", result, err)
	}
	input.Flags = append(input.Flags, canonicalReadExplicitFlag{Name: "include"})
	if _, err := admitCanonicalReadLeaf(input); err == nil {
		t.Fatal("Query output projection was promoted out of AST")
	}
}

func TestCanonicalReadLeafAliasesAndPositionalsAreExplicit(t *testing.T) {
	for _, input := range []canonicalReadLeafInvocation{
		{Profile: cliSurfaceCanonical, Path: []string{"list"}, AliasUsed: true},
		{Profile: cliSurfaceCanonical, Path: []string{"list"}, Positionals: 1},
		{Profile: cliSurfaceCanonical, Path: []string{"link", "list"}, Positionals: 3},
		{Profile: cliSurfaceCanonical, Path: []string{"query"}, Positionals: 1, Flags: []canonicalReadExplicitFlag{{Name: "ast"}}},
	} {
		if result, err := admitCanonicalReadLeaf(input); err == nil || result.Canonical {
			t.Fatalf("bad invocation admitted: %#v %#v %v", input, result, err)
		}
	}
	input := canonicalReadLeafInvocation{Profile: cliSurfaceCanonical, Path: []string{"link", "list"}, Positionals: 2, Flags: []canonicalReadExplicitFlag{{Name: "type"}, {Name: "from"}}}
	if result, err := admitCanonicalReadLeaf(input); err != nil || !result.Canonical {
		t.Fatalf("existing link endpoint surface unavailable: %#v %v", result, err)
	}
	// Type values, positional/flag endpoint conflicts and pins still belong to
	// the existing pure request and checked store admission; this policy has no
	// values or authority and cannot declare those inputs semantically valid.
}

func TestCanonicalReadLeafUnknownCoverageAndProfileNeverFallBack(t *testing.T) {
	for _, input := range []canonicalReadLeafInvocation{
		{Profile: cliSurfaceCanonical, Path: []string{"show"}},
		{Profile: cliSurfaceCanonical, Path: []string{"ready"}},
		{Profile: cliSurfaceCanonical, Path: []string{"memory", "references"}},
		{Profile: cliSurfaceProfile("invalid"), Path: []string{"list"}},
	} {
		if result, err := admitCanonicalReadLeaf(input); err == nil || result.Canonical {
			t.Fatalf("unknown surface/profile fell back: %#v %#v %v", input, result, err)
		}
	}
}
