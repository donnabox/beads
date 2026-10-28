package graphread

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/httpapi/bdpwire"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/resourceget"
)

var _ resourceget.Reader = (*Reader)(nil)

// Get acquires the existing checked current projection in one storage snapshot.
// It is independent of CLI selector parsing and of retained-version reads. The
// reader borrows its store; callers must keep that source open for this call.
func (r *Reader) Get(ctx context.Context, request resourceget.Request) (resourceget.NativeDocument, error) {
	if ctx == nil {
		return resourceget.NativeDocument{}, resourceget.ErrInvalidRequest
	}
	if err := ctx.Err(); err != nil {
		return resourceget.NativeDocument{}, err
	}
	if r == nil || r.store == nil {
		return resourceget.NativeDocument{}, errors.Join(resourceget.ErrInvalidSource, graphstore.ErrInvalidStore)
	}
	path, localID, err := currentSelector(r.ScopeURL(), request)
	if err != nil {
		return resourceget.NativeDocument{}, err
	}
	_, projected, err := r.ResourceSelected(ctx, path, localID)
	if err != nil {
		return resourceget.NativeDocument{}, currentSourceError(err)
	}
	var kind resourceget.Kind
	switch projected.(type) {
	case bdpwire.BeadRecord:
		kind = resourceget.Bead
	case bdpwire.LinkRecord:
		kind = resourceget.Link
	default:
		return resourceget.NativeDocument{}, errors.Join(resourceget.ErrInvalidSource, graphstore.ErrInvalidStore)
	}
	if request.AssertKind != resourceget.Any && request.AssertKind != kind {
		// Preserve the existing CLI Link assertion's native error identity.
		return resourceget.NativeDocument{}, errors.Join(resourceget.ErrKindMismatch, graphstore.ErrCapabilityUnavailable)
	}
	raw, err := json.Marshal(projected)
	if err != nil {
		return resourceget.NativeDocument{}, fmt.Errorf("%w: %w: %v", resourceget.ErrInvalidSource, graphstore.ErrInvalidStore, err)
	}
	doc, err := resourceget.ParseLegacyCurrent(r.ScopeURL(), kind, raw)
	if err != nil {
		return resourceget.NativeDocument{}, fmt.Errorf("%w: %w: %w", resourceget.ErrInvalidSource, graphstore.ErrInvalidStore, err)
	}
	if localID == "" && doc.ID() != request.ID {
		return resourceget.NativeDocument{}, errors.Join(resourceget.ErrInvalidSource, graphstore.ErrInvalidStore)
	}
	if err := ctx.Err(); err != nil {
		return resourceget.NativeDocument{}, err
	}
	return doc, nil
}

func currentSelector(scope string, request resourceget.Request) (path, localID string, err error) {
	if request.AssertKind != resourceget.Any && request.AssertKind != resourceget.Bead && request.AssertKind != resourceget.Link {
		return "", "", resourceget.ErrInvalidRequest
	}
	if err := graph.ValidateScopeURL(scope); err != nil {
		return "", "", errors.Join(resourceget.ErrInvalidSource, graphstore.ErrInvalidStore, err)
	}
	if strings.HasPrefix(request.ID, scope) {
		if _, err := graph.ScopeLocalPublicResourceID(scope, request.ID); err != nil {
			return "", "", errors.Join(resourceget.ErrInvalidRequest, err)
		}
		return request.ID, "", nil
	}
	if err := graph.ValidateCanonicalResourceID(request.ID); err != nil {
		return "", "", errors.Join(resourceget.ErrInvalidRequest, err)
	}
	// This placeholder is the existing selector's internal lookup convention,
	// not a public ID prefix or an assertion that the Resource is a Bead.
	return "beads/" + request.ID, request.ID, nil
}

func currentSourceError(err error) error {
	switch {
	case errors.Is(err, graphstore.ErrNotFound):
		return errors.Join(resourceget.ErrNotFound, err)
	case errors.Is(err, graphstore.ErrGone):
		return errors.Join(resourceget.ErrGone, err)
	case errors.Is(err, graphstore.ErrInvalidStore):
		return errors.Join(resourceget.ErrInvalidSource, err)
	default:
		return err
	}
}
