package graphstore

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/workapi"
)

// ListInstalledTypes returns checked fixed/retained bindings and currently
// configured unused native Issue declarations in stable ID order. This read
// never installs a binding; partial or older generations are refused.
func (s *Store) ListInstalledTypes(ctx context.Context) ([]graph.TypeDescriptor, error) {
	var descriptors []graph.TypeDescriptor
	err := s.withTx(ctx, false, func(tx *sql.Tx) error {
		if err := checkBinding(ctx, tx, s.options); err != nil {
			return err
		}
		var err error
		descriptors, err = s.selectedDescriptorsInTx(ctx, tx)
		return err
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(descriptors, func(i, j int) bool {
		return graph.CompareCodeUnits(descriptors[i].ID(), descriptors[j].ID()) < 0
	})
	return descriptors, nil
}

// ReadType reads a checked selected declaration. Native adapter definitions are
// deterministic; used concrete bindings persist immutably across configuration
// edits and retained reads. Configured unused classes may be derived, and their
// visibility does not authorize new native use. This accessor never registers
// or edits a Type, and every record validates its concrete binding.
func (s *Store) ReadType(ctx context.Context, path string) (graph.TypeDescriptor, error) {
	// Request acquisition budget, independent of the shared identity grammar.
	if len(path) > 1024 {
		return graph.TypeDescriptor{}, fmt.Errorf("%w: Type selector exceeds 1024-byte request budget", ErrLimitExceeded)
	}
	if err := graph.ValidateCanonicalResourceID(path); err != nil {
		return graph.TypeDescriptor{}, err
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
// Concrete native identities are authoritative adapter-derived descriptors;
// their existence does not bypass native classification admission.
func (s *Store) readTypeInTx(ctx context.Context, tx *sql.Tx, id string) (graph.TypeDescriptor, error) {
	// Resolve against the complete selected installer definition set, rather than
	// a hardcoded command alias switch or an unbounded persisted descriptor.
	var name string
	var expected graph.TypeDescriptor
	for _, definition := range previewTypeDefinitions() {
		if id != s.ScopeURL()+definition.name {
			continue
		}
		candidate, err := definition.build(s.ScopeURL())
		if err != nil {
			return graph.TypeDescriptor{}, err
		}
		if candidate.ID() != id {
			return graph.TypeDescriptor{}, fmt.Errorf("%w: selected Type identity differs", ErrInvalidStore)
		}
		name, expected = definition.name, candidate
		break
	}
	if name == "" {
		descriptors, err := s.selectedDescriptorsInTx(ctx, tx)
		if err != nil {
			return graph.TypeDescriptor{}, err
		}
		for _, d := range descriptors {
			if d.ID() == id {
				return d, nil
			}
		}
		return graph.TypeDescriptor{}, ErrNotFound
	}
	var raw []byte
	var fingerprint string
	if err := tx.QueryRowContext(ctx, `SELECT descriptor,fingerprint FROM graph_preview_types WHERE name=? AND OCTET_LENGTH(descriptor)=?`, name, len(expected.CanonicalJSON())).Scan(&raw, &fingerprint); err != nil {
		return graph.TypeDescriptor{}, fmt.Errorf("%w: read installed %s descriptor: %v", ErrInvalidStore, name, err)
	}
	parsed, err := graph.ParseTypeDescriptor(raw)
	if err != nil {
		return graph.TypeDescriptor{}, fmt.Errorf("%w: invalid installed %s descriptor: %v", ErrInvalidStore, name, err)
	}
	if parsed.ID() != id || !bytes.Equal(parsed.CanonicalJSON(), raw) || !bytes.Equal(raw, expected.CanonicalJSON()) || parsed.Fingerprint() != fingerprint || fingerprint != expected.Fingerprint() {
		return graph.TypeDescriptor{}, fmt.Errorf("%w: installed %s descriptor identity or fingerprint differs", ErrInvalidStore, name)
	}
	return parsed, nil
}

// ScopeURL returns the workspace-bound canonical identity, not a transport or
// listener address. Every read verifies this binding against persisted state.
func (s *Store) ScopeURL() string { return s.options.Binding.ScopeURL }

// ResourceProjectionContext contains only the immutable declaration and effective
// ownership slots captured with one authoritative Resource. It is adapter context,
// not stored record content or a public wire field.
type ResourceProjectionContext struct {
	Descriptor     graph.TypeDescriptor
	OwnedLinkTypes []string
}

// ReadForProjection captures one current Resource and its projection context in one transaction.
func (s *Store) ReadForProjection(ctx context.Context, path string) (any, ResourceProjectionContext, error) {
	var projection ResourceProjectionContext
	record, err := s.readWithProjectionContext(ctx, path, &projection)
	if err != nil {
		return nil, ResourceProjectionContext{}, err
	}
	return record, projection, nil
}

// ReadSelectedForProjection resolves a public ID and captures the same snapshot context.
func (s *Store) ReadSelectedForProjection(ctx context.Context, path, localID string) (string, any, ResourceProjectionContext, error) {
	var projection ResourceProjectionContext
	selected, record, _, err := s.readResourceSelected(ctx, path, localID, workapi.DetailOptions{}, false, &projection)
	if err != nil {
		return "", nil, ResourceProjectionContext{}, err
	}
	return selected, record, projection, nil
}

// ReadVersionForProjection captures exact retained content without current owned state.
func (s *Store) ReadVersionForProjection(ctx context.Context, path, revision string) (any, ResourceProjectionContext, error) {
	var projection ResourceProjectionContext
	record, err := s.readVersionWithProjectionContext(ctx, path, revision, &projection)
	if err != nil {
		return nil, ResourceProjectionContext{}, err
	}
	return record, projection, nil
}

func (s *Store) captureResourceProjectionInTx(ctx context.Context, tx *sql.Tx, record any, out *ResourceProjectionContext) error {
	if out == nil {
		return nil
	}
	var typ string
	var owned []json.RawMessage
	switch value := record.(type) {
	case Record:
		typ, owned = value.Type, value.Owned
	case IssueRecord:
		typ, owned = value.Type, value.Owned
	case LinkRecord:
		return nil
	default:
		return fmt.Errorf("%w: unsupported Resource projection context", ErrInvalidStore)
	}
	facts, err := s.selectedDescriptorsInTx(ctx, tx)
	if err != nil {
		return err
	}
	var descriptor graph.TypeDescriptor
	for _, fact := range facts {
		if fact.ID() == typ {
			descriptor = fact
			break
		}
	}
	if descriptor.ID() != typ || descriptor.Describes() != graph.KindBead {
		return fmt.Errorf("%w: captured Bead Type absent from checked inventory", ErrInvalidStore)
	}
	observed := make([]string, 0, len(owned))
	for _, raw := range owned {
		var link LinkRecord
		if err := json.Unmarshal(raw, &link); err != nil {
			return fmt.Errorf("%w: malformed captured ownership", ErrInvalidStore)
		}
		observed = append(observed, link.Type)
	}
	slots, err := s.SelectedOwnedLinkTypesWithDescriptors(typ, facts, observed...)
	if err != nil {
		return err
	}
	*out = ResourceProjectionContext{Descriptor: descriptor, OwnedLinkTypes: slots}
	return nil
}
