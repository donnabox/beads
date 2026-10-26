package graphstore

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"
	"unicode/utf8"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/types"
)

// NativeIssueTypeURL returns the Scope-local concrete identity of a native
// classification. Mapping is injective and does not authorize a custom class:
// the native Issue writer still validates its workspace configuration.
func NativeIssueTypeURL(scope, classification string) (string, error) {
	// The native column is VARCHAR(32), measured in Unicode scalars. Refuse
	// unrepresentable values instead of truncating or assigning an unusable ID.
	if utf8.RuneCountInString(classification) > 32 {
		return "", fmt.Errorf("%w: native Issue classification exceeds 32 characters", graph.ErrValidation)
	}
	return nativeConcreteTypeURL(scope, graph.KindBead, classification)
}

// NativeDependencyTypeURL preserves the exact validated native kind. Its
// identity is the meaningful selected name. Conflicting selected names refuse;
// custom cross-role bindings are checked by the storage transaction.
func NativeDependencyTypeURL(scope, kind string) (string, error) {
	if !types.DependencyType(kind).IsValid() {
		return "", fmt.Errorf("%w: native Dependency kind exceeds its native storage bound", graph.ErrValidation)
	}
	return nativeConcreteTypeURL(scope, graph.KindLink, kind)
}

// nativeConcreteTypeURL is a representation operation, not installation or
// native classification admission. The role comes from the native writer.
func nativeConcreteTypeURL(scope string, role graph.ResourceKind, classification string) (string, error) {
	if classification == "" || !utf8.ValidString(classification) {
		return "", fmt.Errorf("%w: native classification must be nonempty UTF-8", graph.ErrValidation)
	}
	if nativeNameConflicts(classification, role) {
		return "", fmt.Errorf("%w: native classification %q conflicts with a selected Type", ErrAlreadyExists, classification)
	}
	segments := strings.Split(classification, "/")
	for i, segment := range segments {
		segments[i] = graph.CanonicalSegment(segment)
	}
	return graph.CanonicalPublicResourceURL(scope, strings.Join(segments, "/"))
}

// NativeTypeClassification reverses a generated identity for an explicit native
// role supplied by its caller. The full local path is preserved; no basename or
// namespace infers a kind. This does not establish a checked native binding:
// storage must still compare the exact descriptor/backing in its transaction.
func NativeTypeClassification(scope, id string, role graph.ResourceKind) (base, classification string, ok bool) {
	local, err := graph.ScopeLocalPublicResourceID(scope, id)
	if err != nil {
		return "", "", false
	}
	parts := strings.Split(local, "/")
	for i, part := range parts {
		parts[i], err = url.PathUnescape(part)
		if err != nil {
			return "", "", false
		}
	}
	classification = strings.Join(parts, "/")
	var canonical string
	if role == graph.KindBead {
		base = IssueTypeURL(scope)
		canonical, err = NativeIssueTypeURL(scope, classification)
	} else if role == graph.KindLink {
		base = DependencyTypeURL(scope)
		canonical, err = NativeDependencyTypeURL(scope, classification)
	} else {
		return "", "", false
	}
	if err != nil || canonical != id {
		return "", "", false
	}
	return base, classification, true
}

func nativeNameConflicts(name string, role graph.ResourceKind) bool {
	for _, fixed := range append(append([]string(nil), selectedMemoryBeadNames...), selectedMemoryLinkNames...) {
		if name == fixed {
			return true
		}
	}
	if name == "issue" || name == "dependency" || IsReservedTypeName(name) {
		return true
	}

	if role == graph.KindBead {
		for _, kind := range types.WellKnownDependencyTypes() {
			if name == string(kind) {
				return true
			}
		}
	} else if role == graph.KindLink {
		for _, kind := range append(append([]types.IssueType(nil), types.AllIssueTypes...), types.TypeEvent) {
			if name == string(kind) {
				return true
			}
		}
	} else {
		return true
	}
	return false
}

func nativeConcreteDescriptor(scope, id string, role graph.ResourceKind) (graph.TypeDescriptor, error) {
	base, classification, ok := NativeTypeClassification(scope, id, role)
	if !ok {
		return graph.TypeDescriptor{}, fmt.Errorf("%w: invalid native concrete %s Type identity %q", graph.ErrValidation, role, id)
	}
	kind := graph.KindBead
	spec := graph.TypeDescriptorSpec{ID: id, Name: classification, Describes: kind, ConformsTo: []string{base}, Description: "Concrete native adapter Type. The native classification column is authoritative; its domain validator controls admission. This Scope-local identity does not assign a global Type identity."}
	if base == DependencyTypeURL(scope) {
		spec.Describes = graph.KindLink
		endpoint, err := graph.NewEndpointConstraint([]string{IssueTypeURL(scope)}, graph.ExternalNone)
		if err != nil {
			return graph.TypeDescriptor{}, err
		}
		spec.Source, spec.Target = &endpoint, &endpoint
	}
	return graph.NewTypeDescriptor(spec)
}

// IsNativeConcreteTypeDescriptor recognizes only the exact authoritative native
// adapter declaration. It does not authorize use of a custom classification.
func IsNativeConcreteTypeDescriptor(scope string, descriptor graph.TypeDescriptor) bool {
	expected, err := nativeConcreteDescriptor(scope, descriptor.ID(), descriptor.Describes())
	return err == nil && bytes.Equal(expected.CanonicalJSON(), descriptor.CanonicalJSON())
}

// IsReservedTypeName names only the hidden intrinsic local identities. It does
// not strip path components or classify foreign identities by their basename.
func IsReservedTypeName(name string) bool {
	switch name {
	case "type", "bead-type", "link-type", "conforms-to", "refines", "Type", "BeadType", "LinkType", "ConformsTo", "Refines":
		return true
	}
	return false
}
