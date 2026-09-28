package graphstore

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"strings"

	graph "github.com/steveyegge/beads/graphops"
)

// ReadType reads an admitted installed descriptor, never a reconstructed stand-in.
// The preview fixes its four descriptors at initialization; every record read
// and write checks that immutable binding. Separate Type and record reads thus
// cannot silently adopt a changed Type contract. Custom Type installation is not
// supported by this accessor or by the preview writer.
func (s *Store) ReadType(ctx context.Context, path string) (graph.TypeDescriptor, error) {
	if !strings.HasPrefix(path, "types/") {
		return graph.TypeDescriptor{}, fmt.Errorf("%w: expected a local types/ path", graph.ErrValidation)
	}
	// These private installed paths share the canonical identifier-segment
	// grammar of Bead paths; do not normalize or admit URL query selectors.
	if err := validatePath("beads/" + strings.TrimPrefix(path, "types/")); err != nil {
		return graph.TypeDescriptor{}, fmt.Errorf("%w: invalid Type path: %v", graph.ErrValidation, err)
	}
	id := graph.CanonicalURL(s.options.Binding.ScopeURL, path)
	var descriptor graph.TypeDescriptor
	err := s.withTx(ctx, false, func(tx *sql.Tx) error {
		if err := checkBinding(ctx, tx, s.options); err != nil {
			return err
		}
		var err error
		descriptor, err = s.readTypeInTx(ctx, tx, id)
		return err
	})
	if err != nil {
		return graph.TypeDescriptor{}, err
	}
	return descriptor, nil
}

// readTypeInTx requires the caller to validate the binding in this transaction.
func (s *Store) readTypeInTx(ctx context.Context, tx *sql.Tx, id string) (graph.TypeDescriptor, error) {
	var name string
	switch id {
	case MemoryTypeURL(s.options.Binding.ScopeURL):
		name = "memory"
	case IssueTypeURL(s.options.Binding.ScopeURL):
		name = "issue"
	case DependencyTypeURL(s.options.Binding.ScopeURL):
		name = "dependency"
	case RelatedTypeURL(s.options.Binding.ScopeURL):
		name = "related"
	default:
		return graph.TypeDescriptor{}, ErrNotFound
	}
	var raw []byte
	var fingerprint string
	if err := tx.QueryRowContext(ctx, `SELECT descriptor,fingerprint FROM graph_preview_types WHERE name=?`, name).Scan(&raw, &fingerprint); err != nil {
		return graph.TypeDescriptor{}, fmt.Errorf("%w: read installed %s descriptor: %v", ErrInvalidStore, name, err)
	}
	parsed, err := graph.ParseTypeDescriptor(raw)
	if err != nil {
		return graph.TypeDescriptor{}, fmt.Errorf("%w: invalid installed %s descriptor: %v", ErrInvalidStore, name, err)
	}
	if parsed.ID() != id || !bytes.Equal(parsed.CanonicalJSON(), raw) || parsed.Fingerprint() != fingerprint {
		return graph.TypeDescriptor{}, fmt.Errorf("%w: installed %s descriptor identity or fingerprint differs", ErrInvalidStore, name)
	}
	return parsed, nil
}

// ScopeURL returns the workspace-bound canonical identity, not a transport or
// listener address. Every read verifies this binding against persisted state.
func (s *Store) ScopeURL() string { return s.options.Binding.ScopeURL }
