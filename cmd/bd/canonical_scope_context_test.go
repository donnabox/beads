package main

import (
	"errors"
	"strings"
	"testing"

	graph "github.com/steveyegge/beads/graphops"
)

func TestCanonicalScopeContextPrecedenceAndAbsence(t *testing.T) {
	explicit, environment, configuration := "https://explicit.example/team/", "https://environment.example/team/", "https://configuration.example/team/"
	for _, tc := range []struct {
		name   string
		input  canonicalScopeContextInput
		url    string
		source canonicalScopeContextSource
	}{
		{"absent", canonicalScopeContextInput{}, "", canonicalScopeContextNone},
		{"configuration", canonicalScopeContextInput{Configuration: &configuration}, configuration, canonicalScopeContextConfiguration},
		{"environment", canonicalScopeContextInput{Environment: &environment, Configuration: &configuration}, environment, canonicalScopeContextEnvironment},
		{"explicit", canonicalScopeContextInput{Explicit: &explicit, Environment: &environment, Configuration: &configuration}, explicit, canonicalScopeContextExplicit},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := selectCanonicalScopeContext(tc.input)
			if err != nil || got.URL != tc.url || got.Source != tc.source || got.Present() != (tc.source != canonicalScopeContextNone) {
				t.Fatalf("context=%+v err=%v want URL=%q source=%q", got, err, tc.url, tc.source)
			}
		})
	}
}

func TestCanonicalScopeContextSuppliedEmptyDoesNotFallThrough(t *testing.T) {
	empty, valid := "", "https://work.example/team/"
	for _, tc := range []struct {
		name  string
		input canonicalScopeContextInput
	}{
		{"explicit", canonicalScopeContextInput{Explicit: &empty, Environment: &valid, Configuration: &valid}},
		{"environment", canonicalScopeContextInput{Environment: &empty, Configuration: &valid}},
		{"configuration", canonicalScopeContextInput{Configuration: &empty}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := selectCanonicalScopeContext(tc.input)
			if !errors.Is(err, graph.ErrValidation) || got.Present() || got.URL != "" || got.Source != canonicalScopeContextNone {
				t.Fatalf("empty value became absent/fallback: %+v %v", got, err)
			}
		})
	}
}

func TestCanonicalScopeContextUsesExactClientURLLaw(t *testing.T) {
	validFallback := "https://fallback.example/"
	for _, raw := range []string{
		"relative", "https://work.example/team", " https://work.example/team/",
		"https://work.example/team/?q=1", "https://work.example/team/#fragment",
		"https://user:secret@work.example/team/", "https://work.example/a/../b/",
		"https://work.example/a%2Fb/", "https://work.example/%61/",
	} {
		t.Run(raw, func(t *testing.T) {
			got, err := selectCanonicalScopeContext(canonicalScopeContextInput{Explicit: &raw, Environment: &validFallback})
			if !errors.Is(err, graph.ErrValidation) || got.Present() || got.URL != "" {
				t.Fatalf("invalid/noncanonical context was repaired or fell through: %+v %v", got, err)
			}
			if strings.Contains(err.Error(), "secret") || strings.Contains(err.Error(), raw) {
				t.Fatalf("rejected context was echoed: %v", err)
			}
		})
	}
	for _, raw := range []string{"https://work.example/", "https://work.example/team/%E2%82%AC/", "http://localhost:8080/local-test/demo/"} {
		got, err := selectCanonicalScopeContext(canonicalScopeContextInput{Explicit: &raw})
		if err != nil || got.URL != raw || !got.Present() {
			t.Fatalf("valid CLIENT context changed or persistence rules leaked: %+v %v", got, err)
		}
	}
}

func TestCanonicalScopeContextDoesNotValidateUnusedDefaults(t *testing.T) {
	valid, invalid := "https://work.example/team/", "not a Scope"
	for _, input := range []canonicalScopeContextInput{
		{Explicit: &valid, Environment: &invalid, Configuration: &invalid},
		{Environment: &valid, Configuration: &invalid},
	} {
		got, err := selectCanonicalScopeContext(input)
		if err != nil || got.URL != valid || !got.Present() {
			t.Fatalf("unused lower-precedence input affected selected context: %+v %v", got, err)
		}
	}
}

func TestCanonicalScopeContextIsDetachedAndAllowsAbsentDefault(t *testing.T) {
	selected := "https://work.example/team/"
	input := canonicalScopeContextInput{Explicit: &selected}
	got, err := selectCanonicalScopeContext(input)
	if err != nil {
		t.Fatal(err)
	}
	if selected != "https://work.example/team/" || input.Explicit != &selected {
		t.Fatal("selector changed caller input")
	}
	selected = "https://changed.example/"
	if got.URL != "https://work.example/team/" || got.Source != canonicalScopeContextExplicit {
		t.Fatalf("result retained mutable input storage: %+v", got)
	}
	// This selector returns only a Scope context; it has no Resource operand,
	// default-scope requirement or URL rewriting operation. Absolute point
	// identity/reference behavior is qualified by the separate intent/invoker.
	absent, err := selectCanonicalScopeContext(canonicalScopeContextInput{})
	if err != nil || absent.Present() || absent.URL != "" {
		t.Fatalf("absence wrongly required a default Scope: %+v %v", absent, err)
	}
}
