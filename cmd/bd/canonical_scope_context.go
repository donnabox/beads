package main

import (
	"fmt"

	graph "github.com/steveyegge/beads/graphops"
)

// These inputs have already been captured by the CLI composition root. A nil
// pointer means absent; a pointer to an empty string means explicitly supplied.
// The pure selector never reads environment, configuration, files or a store.
type canonicalScopeContextInput struct {
	Explicit      *string
	Environment   *string
	Configuration *string
}

type canonicalScopeContextSource string

const (
	canonicalScopeContextNone          canonicalScopeContextSource = ""
	canonicalScopeContextExplicit      canonicalScopeContextSource = "--scope"
	canonicalScopeContextEnvironment   canonicalScopeContextSource = "BDP_SCOPE_URL"
	canonicalScopeContextConfiguration canonicalScopeContextSource = "scope.url"
)

// This detached value records client context, not an ownership, credential,
// discovery, capability or native-store authority assertion. URL is unchanged
// from the selected input. The zero value is an absent default, not a Scope.
type canonicalScopeContext struct {
	URL    string
	Source canonicalScopeContextSource
}

func (c canonicalScopeContext) Present() bool { return c.Source != canonicalScopeContextNone }

// selectCanonicalScopeContext selects the first supplied value, then validates
// it by the existing CLIENT Scope law. Invalid/empty input never falls through.
// It does not normalize a URL, apply persistence restrictions, infer a Scope
// from a Resource URL, or resolve any Resource/local ordinal.
//
// Absence is permitted here. An inventory or relative-identity caller must
// require a present context before invocation. Absolute point reads use their
// literal Resource identity and need not call this default-context selector.
func selectCanonicalScopeContext(input canonicalScopeContextInput) (canonicalScopeContext, error) {
	for _, candidate := range []struct {
		value  *string
		source canonicalScopeContextSource
	}{
		{input.Explicit, canonicalScopeContextExplicit},
		{input.Environment, canonicalScopeContextEnvironment},
		{input.Configuration, canonicalScopeContextConfiguration},
	} {
		if candidate.value == nil {
			continue
		}
		if err := graph.ValidateScopeURL(*candidate.value); err != nil {
			// Do not echo userinfo or arbitrary rejected context text. This is
			// local input admission, not an invented native Problem mapping.
			return canonicalScopeContext{}, fmt.Errorf("%w: invalid %s Scope context", graph.ErrValidation, candidate.source)
		}
		return canonicalScopeContext{URL: *candidate.value, Source: candidate.source}, nil
	}
	return canonicalScopeContext{}, nil
}
