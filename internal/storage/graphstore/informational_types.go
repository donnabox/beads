package graphstore

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/memorycatalog"
	"github.com/steveyegge/beads/internal/types"
)

// Historical identity helpers remain for callers testing refusal. These Types
// are never installed or admitted by the selected catalog.
func ExampleFollowsTypeURL(scope string) string { return scope + "types/example-follows" }
func ExampleCitesTypeURL(scope string) string   { return scope + "types/example-cites" }

func IsInformationalTypeURL(scope, id string) bool {
	for _, symbol := range selectedMemoryLinkNames {
		if id == scope+symbol {
			return true
		}
	}
	return false
}

var selectedMemoryBeadNames = []string{
	"memory", "evaluation", "constraint", "policy",
	"requirement", "scenario", "specification",
}

var selectedMemoryLinkNames = []string{
	"references", "superseded-by", "requires", "defines", "subject-to",
	"conformance", "addresses", "satisfies", "violates",
}

type previewTypeDefinition struct {
	name  string
	build func(string) (graph.TypeDescriptor, error)
}

// Fresh workspaces install the selected Memory catalog, abstract native family
// bases and fixed concrete native children. Earlier generations are refused.
func previewTypeDefinitions() []previewTypeDefinition {
	definitions := make([]previewTypeDefinition, 0, 20)
	for _, name := range append(append([]string(nil), selectedMemoryBeadNames...), selectedMemoryLinkNames...) {
		definitions = append(definitions, previewTypeDefinition{name, func(scope string) (graph.TypeDescriptor, error) {
			return selectedMemoryDescriptor(scope, name)
		}})
	}
	definitions = append(definitions, previewTypeDefinition{"issue", issueDescriptor}, previewTypeDefinition{"dependency", dependencyDescriptor})
	for _, classification := range append(append([]types.IssueType(nil), types.AllIssueTypes...), types.TypeEvent) {
		name := string(classification)
		definitions = append(definitions, previewTypeDefinition{name, func(scope string) (graph.TypeDescriptor, error) {
			return nativeConcreteDescriptor(scope, scope+name, graph.KindBead)
		}})
	}
	for _, kind := range types.WellKnownDependencyTypes() {
		name := string(kind)
		definitions = append(definitions, previewTypeDefinition{name, func(scope string) (graph.TypeDescriptor, error) {
			return nativeConcreteDescriptor(scope, scope+name, graph.KindLink)
		}})
	}
	return definitions
}

// selectedMemoryDescriptor is a deliberately partial inventory projection. The
// legacy descriptor cannot represent correlated endpoint alternatives or
// positive family ownership without invented finite maxima. Admission uses the
// selectedTypeCatalog sidecar, never these omitted/relaxed details.
func selectedMemoryDescriptor(scope, name string) (graph.TypeDescriptor, error) {
	projection, err := newSelectedMemoryProjection(scope)
	if err != nil {
		return graph.TypeDescriptor{}, err
	}
	declaration, ok := projection.Definition(memorycatalog.Symbol(name))
	if !ok {
		return graph.TypeDescriptor{}, fmt.Errorf("%w: unknown selected Memory Type", ErrNotFound)
	}
	spec := graph.TypeDescriptorSpec{
		ID: declaration.Binding.Type.URI, Name: name, Describes: declaration.Kind,
		Description: fmt.Sprintf("Selected Memory Type; partial inventory descriptor. Property, correlated endpoint, graph and ownership constraints are enforced by the selected catalog. Source %s@%s:%s SHA256 %s.", projection.Source().Repository, projection.Source().Commit, projection.Source().Path, projection.Source().SHA256),
	}
	if declaration.Abstract {
		spec.Description += " Abstract: cannot be used as a declared Type."
	}
	for _, parent := range declaration.Parents {
		spec.ConformsTo = append(spec.ConformsTo, parent.URI)
	}
	if declaration.Kind == graph.KindLink {
		// No independent source/target approximation of correlated pairs.
		endpoint := graph.EndpointConstraint{}
		spec.Source, spec.Target = &endpoint, &endpoint
	}
	return graph.NewTypeDescriptor(spec)
}

// FreshTypeNames lists, in installation order, the Types a fresh init installs.
// The workspace format generation recorded by the CLI must change whenever this
// list does, so the CLI checks the generation against it.
func FreshTypeNames() []string {
	definitions := previewTypeDefinitions()
	names := make([]string, len(definitions))
	for i, definition := range definitions {
		names[i] = definition.name
	}
	return names
}

func exampleFollowsDescriptor(scope string) (graph.TypeDescriptor, error) {
	return exampleInformationalDescriptor(ExampleFollowsTypeURL(scope), "Example follows", "The source follows a policy or model described by the target.")
}

func exampleCitesDescriptor(scope string) (graph.TypeDescriptor, error) {
	return exampleInformationalDescriptor(ExampleCitesTypeURL(scope), "Example cites", "The source cites the target as supporting context.")
}

func exampleInformationalDescriptor(id, name, meaning string) (graph.TypeDescriptor, error) {
	// conformsTo lists are conjunctions, not alternative Memory/Issue Types.
	endpoint, err := graph.NewEndpointConstraint(nil, graph.ExternalNone)
	if err != nil {
		return graph.TypeDescriptor{}, err
	}
	return graph.NewTypeDescriptor(graph.TypeDescriptorSpec{
		ID: id, Name: name, Describes: graph.KindLink,
		Description: meaning + " Informational example with live unpinned local Issue or Memory endpoints and independent multiedge identity. Closed properties admit only an optional UTF-8 string note. Memory sources own this Link; Issue sources do not. No scheduling effect.",
		Source:      &endpoint, Target: &endpoint,
	})
}

// installedPreviewTypes admits only the current fresh Type set. Count and SQL
// byte-length guards reject partial, unknown, replaced or oversized installs.
func installedPreviewTypes(ctx context.Context, tx *sql.Tx, scope string) ([]graph.TypeDescriptor, error) {
	var count int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM graph_preview_types`).Scan(&count); err != nil {
		return nil, err
	}
	definitions := previewTypeDefinitions()
	if count < len(definitions) || count > len(definitions)+PreviewSnapshotLimit {
		return nil, fmt.Errorf("%w: unsupported Type installation", ErrInvalidStore)
	}
	result := make([]graph.TypeDescriptor, 0, count)
	expectedByName := make(map[string]graph.TypeDescriptor, len(definitions))
	args := make([]any, 0, 2*len(definitions))
	for _, definition := range definitions {
		expected, err := definition.build(scope)
		if err != nil {
			return nil, err
		}
		expectedByName[definition.name] = expected
		args = append(args, definition.name, len(expected.CanonicalJSON()))
	}
	// Acquire the fixed set together, retaining each descriptor's exact SQL
	// length guard. Repeated complete-record checks must not issue one query
	// per declaration; all validation still belongs to this transaction.
	rows, err := tx.QueryContext(ctx, fixedTypeDescriptorQuery(len(definitions)), args...)
	if err != nil {
		return nil, fmt.Errorf("%w: read fixed Type descriptors: %v", ErrInvalidStore, err)
	}
	checked := make(map[string]graph.TypeDescriptor, len(definitions))
	for rows.Next() {
		var name string
		var raw []byte
		var fingerprint string
		if err := rows.Scan(&name, &raw, &fingerprint); err != nil {
			return nil, errors.Join(fmt.Errorf("%w: read fixed Type descriptor: %v", ErrInvalidStore, err), rows.Close())
		}
		expected, present := expectedByName[name]
		if !present {
			return nil, errors.Join(fmt.Errorf("%w: unexpected fixed Type name", ErrInvalidStore), rows.Close())
		}
		if _, duplicate := checked[name]; duplicate {
			return nil, errors.Join(fmt.Errorf("%w: duplicate fixed Type name", ErrInvalidStore), rows.Close())
		}
		parsed, err := graph.ParseTypeDescriptor(raw)
		if err != nil || !bytes.Equal(raw, expected.CanonicalJSON()) || fingerprint != expected.Fingerprint() || fingerprint != parsed.Fingerprint() {
			return nil, errors.Join(fmt.Errorf("%w: %s descriptor differs", ErrInvalidStore, name), rows.Close())
		}
		checked[name] = parsed
	}
	if err := errors.Join(rows.Err(), rows.Close()); err != nil {
		return nil, fmt.Errorf("%w: read fixed Type descriptors: %v", ErrInvalidStore, err)
	}
	for _, definition := range definitions {
		parsed, present := checked[definition.name]
		if !present {
			return nil, fmt.Errorf("%w: read %s descriptor: missing or incorrect byte length", ErrInvalidStore, definition.name)
		}
		result = append(result, parsed)
	}
	extra, err := installedNativeBindings(ctx, tx, scope, result)
	if err != nil {
		return nil, err
	}
	result = append(result, extra...)
	return result, nil
}

// Only the number of fixed placeholders varies. Names and exact byte lengths
// remain bound arguments, and no caller text is part of the SQL statement.
func fixedTypeDescriptorQuery(count int) string {
	return `SELECT name,descriptor,fingerprint FROM graph_preview_types WHERE ` + strings.TrimSuffix(strings.Repeat("(name=? AND OCTET_LENGTH(descriptor)=?) OR ", count), " OR ")
}

func (s *Store) informationalTypeInTx(ctx context.Context, tx *sql.Tx, id string) error {
	if !IsInformationalTypeURL(s.ScopeURL(), id) {
		return fmt.Errorf("%w: invalid informational Type", ErrInvalidStore)
	}
	if _, err := s.readTypeInTx(ctx, tx, id); err != nil {
		return fmt.Errorf("%w: informational Type is not installed: %v", ErrInvalidStore, err)
	}
	return nil
}

// Keep collection integrity based on the installed set, not every Type the
// binary knows.
func (s *Store) installedInformationalTypes(ctx context.Context, tx *sql.Tx) ([]string, error) {
	descriptors, err := installedPreviewTypes(ctx, tx, s.ScopeURL())
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, descriptor := range descriptors {
		if IsInformationalTypeURL(s.ScopeURL(), descriptor.ID()) {
			ids = append(ids, descriptor.ID())
		}
	}
	return ids, nil
}

func informationalTypeSQL(ids []string) (string, []any) {
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	return strings.TrimSuffix(strings.Repeat("?,", len(ids)), ","), args
}

// User selection of a supported but absent example is unavailable, whereas an
// authoritative record referring to that absent Type is corrupt store state.
func (s *Store) selectedLinkTypeInTx(ctx context.Context, tx *sql.Tx, id string) error {
	if _, err := s.readTypeInTx(ctx, tx, id); err != nil {
		if errors.Is(err, ErrNotFound) {
			return fmt.Errorf("%w: Link Type is not in the installed selected catalog", ErrCapabilityUnavailable)
		}
		return err
	}
	return nil
}
