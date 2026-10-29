package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// Resolve local aliases before any Reference reaches a wire or writer. The
// immutable allocation lookup and bounded History reader never select by row
// index. An ordinal may resolve to D; the chosen operation then decides whether
// terminal evidence is admissible (guards yes, retained live payload no).
func graphPreviewResolveParsedReference(ref graphPreviewReference) (result graphPreviewReference, err error) {
	if !ref.Explicit || ref.currentSelection() || (ref.Ordinal == 0 && !ref.LiteralRevision) {
		return ref, nil
	}
	if strings.Contains(ref.Identity, "://") && !graphPreviewCanonicalResourceAddress(graphPreviewConfig.GraphScopeURL, ref.Identity) {
		return ref, fmt.Errorf("retained revision selectors require a canonical URL in this Scope")
	}
	defer func() {
		if err != nil {
			err = graphPreviewReferenceResolutionError{err}
		}
	}()
	ctx, cancel := context.WithTimeout(getRootContext(), 30*time.Second)
	defer cancel()
	store, err := openSelectedGraphStore(ctx, graphOptions(graphPreviewConfig))
	if err != nil {
		return ref, err
	}
	defer func() { err = errors.Join(err, store.Close()) }()
	var identity graphstore.AllocatedResourceIdentity
	if graphPreviewCanonicalResourceAddress(graphPreviewConfig.GraphScopeURL, ref.Identity) {
		identity, err = store.ResolveCanonicalPublicResourceIdentity(ctx, ref.Identity)
	} else {
		identity, err = store.ResolveAllocatedResourceIdentity(ctx, ref.Identity)
	}
	if err != nil {
		return ref, err
	}
	ref.Identity = graphPreviewConfig.GraphScopeURL + identity.PublicID
	if ref.Ordinal != 0 {
		row, readErr := store.ResolveLocalRevision(ctx, identity.PrivatePath, ref.Ordinal)
		if readErr != nil {
			return ref, readErr
		}
		ref.Revision, ref.Ordinal, ref.LiteralRevision = row.Version, 0, true
	}
	return ref, nil
}

type graphPreviewReferenceResolutionError struct{ error }

func (e graphPreviewReferenceResolutionError) Unwrap() error { return e.error }

func graphPreviewReferenceStorageError(err error) bool {
	var resolution graphPreviewReferenceResolutionError
	return errors.As(err, &resolution) || errors.Is(err, graphstore.ErrVersionUnknown) || errors.Is(err, graphstore.ErrNotFound) || errors.Is(err, graphstore.ErrGone) || errors.Is(err, graphstore.ErrInvalidStore) || errors.Is(err, graphstore.ErrLimitExceeded) || errors.Is(err, graphstore.ErrCapabilityUnavailable)
}

func graphPreviewReferenceFailure(err error) error {
	var exit *exitError
	if errors.As(err, &exit) {
		return err
	}
	if graphPreviewReferenceStorageError(err) {
		return graphStorageError(err)
	}
	return graphFailure("invalid_selector", err.Error(), 2)
}

func graphPreviewResolveReference(input string) (graphPreviewReference, error) {
	ref, err := graphPreviewParseReference(input)
	if err != nil {
		return ref, err
	}
	return graphPreviewResolveParsedReference(ref)
}
