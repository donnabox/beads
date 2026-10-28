package graphstore

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"sort"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage/issueops"
	"github.com/steveyegge/beads/internal/types"
)

// nativeFamilyReference is the private semantic adapter: a concrete native
// declaration retains its public identity, while inherited endpoint/ownership
// constraints operate on the native family. This never authorizes admission.
// nativeFamilyReference uses exact native descriptor facts, not URI syntax.
// With no captured inventory only the explicit fixed native catalog is known.
func nativeFamilyReference(scope, id string, role graph.ResourceKind, facts ...graph.TypeDescriptor) string {
	if len(facts) == 0 {
		if role == graph.KindBead {
			for _, value := range append(append([]types.IssueType(nil), types.AllIssueTypes...), types.TypeEvent) {
				uri, err := NativeIssueTypeURL(scope, string(value))
				if err == nil && uri == id {
					return IssueTypeURL(scope)
				}
			}
		} else if role == graph.KindLink {
			for _, value := range types.WellKnownDependencyTypes() {
				uri, err := NativeDependencyTypeURL(scope, string(value))
				if err == nil && uri == id {
					return DependencyTypeURL(scope)
				}
			}
		}
		return id
	}
	for _, d := range facts {
		if d.ID() == id && d.Describes() == role && IsNativeConcreteTypeDescriptor(scope, d) {
			if role == graph.KindBead {
				return IssueTypeURL(scope)
			}
			return DependencyTypeURL(scope)
		}
	}
	return id
}
func isConcreteNativeIssue(scope, id string, facts ...graph.TypeDescriptor) bool {
	return id != IssueTypeURL(scope) && nativeFamilyReference(scope, id, graph.KindBead, facts...) == IssueTypeURL(scope)
}
func isConcreteNativeDependency(scope, id string, facts ...graph.TypeDescriptor) bool {
	return id != DependencyTypeURL(scope) && nativeFamilyReference(scope, id, graph.KindLink, facts...) == DependencyTypeURL(scope)
}

// selectedDescriptorsInTx combines checked immutable bindings with configured
// unused native Issue classes. Used bindings never disappear after a config
// edit or reclassification, and enumeration never scans retained payloads.
func (s *Store) selectedDescriptorsInTx(ctx context.Context, tx *sql.Tx) ([]graph.TypeDescriptor, error) {
	result, err := installedPreviewTypes(ctx, tx, s.ScopeURL())
	if err != nil {
		return nil, err
	}
	seen := map[string]graph.TypeDescriptor{}
	for _, d := range result {
		seen[d.ID()] = d
	}
	custom, err := issueops.ResolveCustomTypesInTx(ctx, tx)
	if err != nil {
		return nil, err
	}
	for _, value := range custom {
		id, err := configuredNativeIssueTypeURL(s.ScopeURL(), value)
		if err != nil {
			return nil, err
		}
		d, err := nativeConcreteDescriptor(s.ScopeURL(), id, graph.KindBead)
		if err != nil {
			return nil, err
		}
		if previous, exists := seen[id]; exists {
			if !bytes.Equal(previous.CanonicalJSON(), d.CanonicalJSON()) {
				return nil, fmt.Errorf("%w: configured native classification conflicts with an installed Type", ErrAlreadyExists)
			}
		} else {
			// The complete checked installed set above already proves this
			// exact Type identity absent; do not rescan every descriptor for
			// each configured unused class. Other reservations stay checked.
			owner, err := allocatedResourceIDOwnerInTx(ctx, tx, id[len(s.ScopeURL()):])
			if err != nil {
				return nil, err
			}
			if owner != "" {
				return nil, fmt.Errorf("%w: configured Type identity is already reserved by a Resource or discovery", ErrAlreadyExists)
			}
			result = append(result, d)
			seen[id] = d
		}
	}
	if len(result) > len(previewTypeDefinitions())+PreviewSnapshotLimit {
		return nil, fmt.Errorf("%w: complete Type inventory exceeds acquisition bound", ErrLimitExceeded)
	}
	sort.Slice(result, func(i, j int) bool { return graph.CompareCodeUnits(result[i].ID(), result[j].ID()) < 0 })
	return result, nil
}

func nativeIssueTypeMatches(scope, actual, family string) bool {
	base, _, ok := NativeTypeClassification(scope, actual, graph.KindBead)
	return family == IssueTypeURL(scope) && ok && base == family
}

func (s *Store) nativeDependencyTypeIDsInTx(ctx context.Context, tx *sql.Tx) ([]string, error) {
	descriptors, err := s.selectedDescriptorsInTx(ctx, tx)
	if err != nil {
		return nil, err
	}
	var result []string
	for _, d := range descriptors {
		if isConcreteNativeDependency(s.ScopeURL(), d.ID(), d) {
			result = append(result, d.ID())
		}
	}
	return result, nil
}

// selectedFamilyIDsInTx resolves a family against the complete checked descriptor
// inventory. It does not infer semantics from a URI basename or native label.
func (s *Store) selectedFamilyIDsInTx(ctx context.Context, tx *sql.Tx, required string, kind graph.ResourceKind) (map[string]bool, error) {
	if required == "" {
		return nil, nil
	}
	descriptors, err := s.selectedDescriptorsInTx(ctx, tx)
	if err != nil {
		return nil, err
	}
	byID := map[string]graph.TypeDescriptor{}
	for _, descriptor := range descriptors {
		byID[descriptor.ID()] = descriptor
	}
	root, exists := byID[required]
	if !exists || root.Describes() != kind {
		return nil, fmt.Errorf("%w: selected family is not an installed %s Type", ErrCapabilityUnavailable, kind)
	}
	result := map[string]bool{}
	for _, descriptor := range descriptors {
		if descriptor.Describes() != kind {
			continue
		}
		pending, seen := []string{descriptor.ID()}, map[string]bool{}
		for len(pending) != 0 {
			id := pending[len(pending)-1]
			pending = pending[:len(pending)-1]
			if seen[id] {
				continue
			}
			seen[id] = true
			parent, exists := byID[id]
			if !exists || parent.Describes() != kind {
				return nil, fmt.Errorf("%w: incomplete Type ancestry", ErrInvalidStore)
			}
			if id == required {
				result[descriptor.ID()] = true
			}
			pending = append(pending, parent.ConformsTo()...)
		}
	}
	return result, nil
}
