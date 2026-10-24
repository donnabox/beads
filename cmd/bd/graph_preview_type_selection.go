package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// These names describe the Type system itself, not user-declared Resource Types.
// Public Conformance is a different (abstract) Link Type and stays selectable.
func graphPreviewReservedTypeSelector(selector string) bool {
	name := selector[strings.LastIndex(selector, "/")+1:]
	name = strings.ToLower(strings.NewReplacer("-", "", "_", "", " ", "").Replace(name))
	switch name {
	case "type", "beadtype", "linktype", "conformsto", "refines":
		return true
	}
	return false
}

// Selection inspects only the checked installed descriptor inventory. It does
// not treat a partial legacy schema or ownership projection as admission proof.
func graphPreviewValidateSelectedType(scope string, installed []graph.TypeDescriptor, id string, kind graph.ResourceKind) error {
	if graphPreviewReservedTypeSelector(id) {
		return fmt.Errorf("%w: intrinsic Type is not a public selector", graphstore.ErrCapabilityUnavailable)
	}
	for _, descriptor := range installed {
		if descriptor.ID() != id {
			continue
		}
		if descriptor.Describes() != kind || (!graphstore.IsNativeConcreteTypeDescriptor(scope, descriptor) && graphPreviewIntrinsicType(descriptor)) {
			return fmt.Errorf("%w: installed Type does not describe the selected Resource kind", graphstore.ErrCapabilityUnavailable)
		}
		if !strings.HasPrefix(id, scope+"types/") {
			break
		}
		return nil
	}
	return fmt.Errorf("%w: Type is not installed in this workspace", graphstore.ErrCapabilityUnavailable)
}

func graphPreviewTypeRefines(installed []graph.TypeDescriptor, actual, required string) bool {
	byID := make(map[string]graph.TypeDescriptor, len(installed))
	for _, descriptor := range installed {
		byID[descriptor.ID()] = descriptor
	}
	seen := map[string]bool{}
	pending := []string{actual}
	for len(pending) != 0 {
		id := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		if seen[id] {
			continue
		}
		seen[id] = true
		descriptor, present := byID[id]
		if !present {
			continue
		}
		if id == required {
			return true
		}
		pending = append(pending, descriptor.ConformsTo()...)
	}
	return false
}

func graphPreviewRequireInstalledType(ctx context.Context, store *graphstore.Store, id string, kind graph.ResourceKind, memoryFamily bool) error {
	installed, err := store.ListInstalledTypes(ctx)
	if err != nil {
		return err
	}
	if err := graphPreviewValidateSelectedType(store.ScopeURL(), installed, id, kind); err != nil {
		return err
	}
	if memoryFamily && !graphPreviewTypeRefines(installed, id, graphstore.MemoryTypeURL(store.ScopeURL())) {
		return fmt.Errorf("%w: this operation requires the installed Memory family; native Types keep their domain writer", graphstore.ErrCapabilityUnavailable)
	}
	return nil
}

func graphPreviewMemoryTypeFlag(cmd *cobra.Command) (string, error) {
	if !cmd.Flags().Changed("type") {
		return "", nil
	}
	selector, _ := cmd.Flags().GetString("type")
	if selector == "" {
		return "", graphFailure("invalid_selector", "--type requires a value", 2)
	}
	typeURL, err := graphPreviewTypeURL(graphPreviewConfig.GraphScopeURL, selector)
	if err != nil {
		return "", graphFailure("invalid_selector", err.Error(), 2)
	}
	return typeURL, nil
}

// The fresh-generation catalog is a cheap preflight, not a substitute for the
// checked persisted descriptor read before an operation or final validation.
func graphPreviewSelectedTypeCandidate(scope, id string) bool {
	for _, name := range graphstore.FreshTypeNames() {
		if id == scope+"types/"+name {
			return true
		}
	}
	return false
}
