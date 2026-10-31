package main

import (
	"context"
	"fmt"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/httpapi/graphread"
	"github.com/steveyegge/beads/internal/resourcequery"
)

// This concrete adapter borrows the already opened checked local Reader. The
// existing caller owns OpenExisting/Close-before-deliver; this adapter neither
// opens a fallback store nor closes somebody else's store or caches capability.
type canonicalLocalQueryAdapter struct{ reader *graphread.Reader }

func newCanonicalLocalQueryAdapter(reader *graphread.Reader) (canonicalLocalQueryAdapter, error) {
	if reader == nil {
		return canonicalLocalQueryAdapter{}, fmt.Errorf("%w: checked local Query reader required", resourcequery.ErrInvalid)
	}
	if err := graph.ValidateScopeURL(reader.ScopeURL()); err != nil {
		return canonicalLocalQueryAdapter{}, fmt.Errorf("%w: invalid checked local Query Scope: %v", resourcequery.ErrDataset, err)
	}
	return canonicalLocalQueryAdapter{reader: reader}, nil
}

func (a canonicalLocalQueryAdapter) scopeURL() string { return a.reader.ScopeURL() }
func (a canonicalLocalQueryAdapter) execute(ctx context.Context, raw []byte) (resourcequery.Response, error) {
	if a.reader == nil {
		return resourcequery.Response{}, fmt.Errorf("%w: checked local Query reader required", resourcequery.ErrDataset)
	}
	return a.reader.ExecuteResourceQuery(ctx, raw)
}

// invokeCanonicalGraphLocal takes already-resolved explicit operands and an
// already selected literal Scope. It really executes the checked Reader through
// the SAME kernel as Query AST. It chooses no Graph CLI defaults/descriptor,
// rewrites no compatibility graph, resolves no selectors and activates no HTTP.
func invokeCanonicalGraphLocal(ctx context.Context, owningScope string, reader *graphread.Reader, options canonicalGraphRequestOptions) (canonicalLocalQueryResult, error) {
	request, err := buildCanonicalGraphRequest(options)
	if err != nil {
		return canonicalLocalQueryResult{}, err
	}
	invocation, err := canonicalQueryInvocationFromRequest(request)
	if err != nil {
		return canonicalLocalQueryResult{}, err
	}
	adapter, err := newCanonicalLocalQueryAdapter(reader)
	if err != nil {
		return canonicalLocalQueryResult{}, err
	}
	return invokeCanonicalQueryLocal(ctx, owningScope, invocation, adapter)
}
