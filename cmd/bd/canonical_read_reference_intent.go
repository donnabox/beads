package main

import (
	"context"
	"fmt"
	"strings"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/httpapi/bdpwire"
)

// canonicalReadReferenceIntent is CLI syntax, not a wire Reference, verified
// Scope binding, kind assertion or permission to acquire a local store.
// Construction uses the existing Reference parser exactly once.
type canonicalReadReferenceIntent struct {
	reference graphPreviewReference
	absolute  bool
	ready     bool
}

// canonicalReadReferenceResolvers contains only selected-authority facts. The
// caller binds locateLocal to its actual selected Scope/identity authority;
// it must preserve whole local IDs without guessing kind prefixes. No default
// Scope or graphPreviewConfig is consulted here. resolveOrdinal is an actual
// stable mapping for this exact Resource authority, never an Nth History row.
// Returning only a token prevents ordinal resolution from retargeting a URI.
// Callbacks must honor the exact supplied context/deadline and perform no
// fallback to a different authority, profile or compatibility representation.
type canonicalReadReferenceResolvers struct {
	locateLocal    func(context.Context, string) (string, error)
	resolveOrdinal func(context.Context, string, int64) (string, error)
}

func parseCanonicalReadReferenceIntent(input string) (canonicalReadReferenceIntent, error) {
	ref, err := graphPreviewParseReferenceWithLocatorPolicy(input, graphPreviewReferenceCanonicalReadLocator)
	if err != nil {
		return canonicalReadReferenceIntent{}, err
	}
	absolute := strings.Contains(ref.Identity, "://")
	if absolute {
		if err := validateCanonicalReadLocator(ref.Identity); err != nil {
			return canonicalReadReferenceIntent{}, err
		}
	} else if err := graph.ValidateCanonicalResourceID(ref.Identity); err != nil {
		return canonicalReadReferenceIntent{}, err
	}
	return canonicalReadReferenceIntent{reference: ref, absolute: absolute, ready: true}, nil
}

// validateCanonicalReadLocator reuses the existing canonical HTTP(S) URL law
// without normalizing or inferring Resource kind, owning Scope, existence or
// authorization. As at the current Resource boundary, identity has no query;
// native GET selectors/encoding belong to later request preparation.
func validateCanonicalReadLocator(uri string) error {
	if err := graph.ValidateTypeURL(uri); err != nil {
		return err
	}
	if strings.Contains(uri, "?") {
		return fmt.Errorf("%w: Resource identity must not contain a query", graph.ErrValidation)
	}
	return nil
}

// resolveCanonicalReadReferenceIntent returns the existing Reference VALUE
// carrier only. Its v0 codec does not establish selected retained-Core binding
// or prove a usable GET. Absolute current/exact locators need no default Scope
// or local callback. Kind assertions and native request/capability admission
// belong to the later read invoker, not URI spelling or this resolver.
//
// The resolver opens no store or transport. It never invokes the old
// graphPreviewResolveParsedReference, which acquires a local graphstore even
// for an exact selector. Every callback error is returned unchanged. Nil or
// canceled contexts and a zero intent fail before callbacks; cancellation is
// checked again after successful callbacks and before returning a value.
func resolveCanonicalReadReferenceIntent(ctx context.Context, intent canonicalReadReferenceIntent,
	resolvers canonicalReadReferenceResolvers) (bdpwire.Reference, error) {
	if ctx == nil {
		return bdpwire.Reference{}, fmt.Errorf("canonical read requires a context")
	}
	if err := ctx.Err(); err != nil {
		return bdpwire.Reference{}, err
	}
	if !intent.ready {
		return bdpwire.Reference{}, fmt.Errorf("canonical read requires a parsed Reference intent")
	}
	ref := intent.reference
	uri := ref.Identity
	if intent.absolute {
		if err := validateCanonicalReadLocator(uri); err != nil {
			return bdpwire.Reference{}, err
		}
	} else {
		if resolvers.locateLocal == nil {
			return bdpwire.Reference{}, fmt.Errorf("selected-authority local Resource locator is unavailable")
		}
		var err error
		uri, err = resolvers.locateLocal(ctx, ref.Identity)
		if err != nil {
			return bdpwire.Reference{}, err
		}
		if err := ctx.Err(); err != nil {
			return bdpwire.Reference{}, err
		}
		if err := validateCanonicalReadLocator(uri); err != nil {
			return bdpwire.Reference{}, err
		}
	}
	var token string
	if ref.Ordinal != 0 {
		if resolvers.resolveOrdinal == nil {
			return bdpwire.Reference{}, fmt.Errorf("selected-authority stable ordinal mapping is unavailable")
		}
		var err error
		token, err = resolvers.resolveOrdinal(ctx, uri, ref.Ordinal)
		if err != nil {
			return bdpwire.Reference{}, err
		}
		if err := ctx.Err(); err != nil {
			return bdpwire.Reference{}, err
		}
		if err := graphPreviewReferenceRevision(token); err != nil {
			return bdpwire.Reference{}, err
		}
	} else if ref.LiteralRevision {
		token = ref.Revision
		if err := graphPreviewReferenceRevision(token); err != nil {
			return bdpwire.Reference{}, err
		}
	}
	if err := ctx.Err(); err != nil {
		return bdpwire.Reference{}, err
	}
	return bdpwire.Reference{URI: uri, Revision: token}, nil
}
