package graphstore

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/memorycatalog"
	"github.com/steveyegge/beads/internal/typecatalog"
	"github.com/steveyegge/beads/resourceget"
)

// installedTypeSource is initializer input, never a second mutable runtime
// catalog. The stored record/contract must be acquired and checked together
// before constructing the operation's compiler. Native nominal validation is
// retained; a partial legacy descriptor is never an executable properties schema.
type installedTypeSource struct {
	Name       string
	Descriptor graph.TypeDescriptor
	Contract   typecatalog.Definition
	Memory     *memorycatalog.Definition
}

func selectedInstalledTypeSources(scope string) ([]installedTypeSource, error) {
	selected, err := newSelectedTypeCatalog(scope)
	if err != nil {
		return nil, err
	}
	result := make([]installedTypeSource, 0, len(previewTypeDefinitions()))
	seen := map[string]bool{}
	for _, factory := range previewTypeDefinitions() {
		descriptor, err := factory.build(scope)
		if err != nil {
			return nil, err
		}
		if descriptor.IsZero() || seen[descriptor.ID()] {
			return nil, fmt.Errorf("%w: duplicate installed Type source", ErrInvalidStore)
		}
		seen[descriptor.ID()] = true
		definition, ok := selected.types.Definition(descriptor.ID())
		if !ok {
			family := nativeFamilyReference(scope, descriptor.ID(), descriptor.Describes(), descriptor)
			if family == descriptor.ID() {
				return nil, fmt.Errorf("%w: installed Type has no selected source contract", ErrInvalidStore)
			}
			definition, ok = selected.types.Definition(family)
			if !ok || !definition.NominalOnly {
				return nil, fmt.Errorf("%w: invalid native Type source", ErrInvalidStore)
			}
			definition.Identity = typecatalog.Reference{URI: descriptor.ID()}
			definition.Name = descriptor.Name()
			definition.Abstract = false
			definition.Parents = []typecatalog.Reference{{URI: family}}
		}
		if definition.Identity.URI != descriptor.ID() || definition.Kind != descriptor.Describes() {
			return nil, fmt.Errorf("%w: installed descriptor/source mismatch", ErrInvalidStore)
		}
		source := installedTypeSource{Name: factory.name, Descriptor: descriptor, Contract: definition}
		if declaration, ok := selected.definitions[descriptor.ID()]; ok {
			copy := declaration
			source.Memory = &copy
		}
		result = append(result, source)
	}
	return result, nil
}

// bindInstalledTypeContract closes every Type reference against allocations
// already chosen in this initializer transaction. It neither makes aliases nor
// substitutes the current definition for a recorded exact reference.
func bindInstalledTypeContract(source typecatalog.Definition, refs map[string]resourceget.TypeReference) (typecatalog.Definition, error) {
	raw, err := json.Marshal(source)
	if err != nil {
		return typecatalog.Definition{}, err
	}
	var result typecatalog.Definition
	if err = json.Unmarshal(raw, &result); err != nil {
		return typecatalog.Definition{}, err
	}
	bind := func(ref typecatalog.Reference) (typecatalog.Reference, error) {
		actual, ok := refs[ref.URI]
		if !ok || actual.URI != ref.URI || actual.Revision == "" || ref.Revision != "" && ref.Revision != actual.Revision {
			return typecatalog.Reference{}, fmt.Errorf("%w: incomplete exact installed Type closure", ErrInvalidStore)
		}
		return typecatalog.Reference{URI: actual.URI, Revision: actual.Revision}, nil
	}
	result.Identity, err = bind(result.Identity)
	if err != nil {
		return typecatalog.Definition{}, err
	}
	for i, ref := range result.Parents {
		result.Parents[i], err = bind(ref)
		if err != nil {
			return typecatalog.Definition{}, err
		}
	}
	bindSelector := func(selector *typecatalog.Selector) error {
		for i, ref := range selector.AnyOf {
			selector.AnyOf[i], err = bind(ref)
			if err != nil {
				return err
			}
		}
		return nil
	}
	for i := range result.OwnsOutgoing {
		if err := bindSelector(&result.OwnsOutgoing[i]); err != nil {
			return typecatalog.Definition{}, err
		}
	}
	for i := range result.EndpointPairs {
		if err := bindSelector(&result.EndpointPairs[i].Source.Types); err != nil {
			return typecatalog.Definition{}, err
		}
		if err := bindSelector(&result.EndpointPairs[i].Target.Types); err != nil {
			return typecatalog.Definition{}, err
		}
	}
	return result, nil
}

// installedTypeProperties emits local declarations from the complete selected
// source contract. Additional native/Memory domain checks remain source-bound
// requirements; this does not manufacture a replacement general JSON schema.
func installedTypeProperties(source installedTypeSource, contract typecatalog.Definition) (json.RawMessage, error) {
	properties := map[string]any{"abstract": contract.Abstract}
	if description, ok := source.Descriptor.Description(); ok {
		properties["description"] = description
	}
	if len(contract.PropertiesSchema) != 0 {
		var schema map[string]json.RawMessage
		if json.Unmarshal(contract.PropertiesSchema, &schema) != nil || schema == nil {
			return nil, fmt.Errorf("%w: selected installed properties schema", ErrInvalidStore)
		}
		schema["$schema"] = json.RawMessage(`"https://json-schema.org/draft/2020-12/schema"`)
		base, err := json.Marshal(contract.SchemaBase)
		if err != nil {
			return nil, err
		}
		schema["$id"] = base
		properties["propertiesSchema"] = schema
		if len(contract.SchemaResources) != 0 {
			properties["schemaResources"] = contract.SchemaResources
		}
	}
	selector := func(value typecatalog.Selector) map[string]any {
		if value.Any {
			return map[string]any{"any": true}
		}
		refs := make([]resourceget.TypeReference, 0, len(value.AnyOf))
		for _, ref := range value.AnyOf {
			refs = append(refs, resourceget.TypeReference{URI: ref.URI, Revision: ref.Revision})
		}
		return map[string]any{"anyOf": refs}
	}
	if contract.Kind == graph.KindBead {
		if len(contract.OwnsOutgoing) != 0 {
			owns := make([]map[string]any, 0, len(contract.OwnsOutgoing))
			for _, value := range contract.OwnsOutgoing {
				owns = append(owns, selector(value))
			}
			properties["ownsOutgoing"] = owns
		}
	} else {
		endpoint := func(value typecatalog.EndpointConstraint) map[string]any {
			if value.ResourceKind == "resource" {
				return map[string]any{"resourceKind": "resource"}
			}
			return map[string]any{"resourceKind": value.ResourceKind, "scope": value.Scope, "types": selector(value.Types)}
		}
		pairs := make([]map[string]any, 0, len(contract.EndpointPairs))
		for _, value := range contract.EndpointPairs {
			pairs = append(pairs, map[string]any{"source": endpoint(value.Source), "target": endpoint(value.Target)})
		}
		if len(pairs) == 0 {
			return nil, fmt.Errorf("%w: installed Link Type lacks its actual endpoint contract", ErrInvalidStore)
		}
		constraints := map[string]any{"typeConstraints": pairs}
		if source.Memory != nil {
			rule := source.Memory.LocalGraphRule
			if rule.SourceMax.Present {
				constraints["sourceMax"] = rule.SourceMax.Value
			}
			if rule.TargetMax.Present {
				constraints["targetMax"] = rule.TargetMax.Value
			}
			if rule.Unique {
				constraints["unique"] = true
			}
			if rule.Acyclic {
				constraints["acyclic"] = true
			}
			if rule.ForbidIdentitySelf {
				constraints["allowSelf"] = false
			}
		}
		properties["linkConstraints"] = constraints
	}
	return canonicalJSON(properties)
}

// Additional concrete native classifications are installed only by their real
// writer. Configured-but-unused declarations are not materialized implicitly.
// The fixed source plus exact native binding rows bounds the allowed records;
// no current payload or basename supplies a new executable contract.
func selectedInstalledTypeSourcesInTx(ctx context.Context, tx *sql.Tx, scope string) ([]installedTypeSource, error) {
	sources, err := selectedInstalledTypeSources(scope)
	if err != nil {
		return nil, err
	}
	baseline := make([]graph.TypeDescriptor, 0, len(sources))
	for _, source := range sources {
		baseline = append(baseline, source.Descriptor)
	}
	bindings, err := installedNativeBindings(ctx, tx, scope, baseline)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	for _, source := range sources {
		seen[source.Descriptor.ID()] = true
	}
	for _, descriptor := range bindings {
		if seen[descriptor.ID()] {
			return nil, ErrInvalidStore
		}
		var name, fingerprint string
		var raw []byte
		if err = tx.QueryRowContext(ctx, `SELECT name,descriptor,fingerprint FROM graph_preview_types WHERE fingerprint=? AND OCTET_LENGTH(descriptor)=?`, descriptor.Fingerprint(), len(descriptor.CanonicalJSON())).Scan(&name, &raw, &fingerprint); err != nil {
			return nil, err
		}
		if !bytes.Equal(raw, descriptor.CanonicalJSON()) || fingerprint != descriptor.Fingerprint() {
			return nil, ErrInvalidStore
		}
		source, err := nativeInstalledTypeSource(scope, name, descriptor)
		if err != nil {
			return nil, err
		}
		seen[descriptor.ID()] = true
		sources = append(sources, source)
	}
	return sources, nil
}

func nativeInstalledTypeSource(scope, name string, descriptor graph.TypeDescriptor) (installedTypeSource, error) {
	selected, err := newSelectedTypeCatalog(scope)
	if err != nil {
		return installedTypeSource{}, err
	}
	family := nativeFamilyReference(scope, descriptor.ID(), descriptor.Describes(), descriptor)
	definition, known := selected.types.Definition(family)
	if !known || !definition.NominalOnly || family == descriptor.ID() {
		return installedTypeSource{}, ErrInvalidStore
	}
	definition.Identity = typecatalog.Reference{URI: descriptor.ID()}
	definition.Name = descriptor.Name()
	definition.Abstract = false
	definition.Parents = []typecatalog.Reference{{URI: family}}
	return installedTypeSource{Name: name, Descriptor: descriptor, Contract: definition}, nil
}
