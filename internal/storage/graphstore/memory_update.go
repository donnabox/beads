package graphstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"unicode/utf8"

	graph "github.com/steveyegge/beads/graphops"

	"github.com/steveyegge/beads/internal/graphpatch"
	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/storage/issueops"
	publicops "github.com/steveyegge/beads/issueops"
)

// MemoryUpdateRequest replaces the complete selected Memory properties object.
// Empty strings are values; the caller must distinguish omitted JSON members.
type MemoryUpdateRequest struct {
	TypeURL          string
	Path             string
	Properties       Properties
	Actor            string
	ExpectedRevision string
	Unconditional    bool
	Metadata         publicops.MetadataPatch
}

type MemoryMutationResult struct {
	Memory   Record          `json:"memory"`
	Changed  bool            `json:"changed"`
	Replaced *ReplacedMemory `json:"replaced,omitempty"`
}

// UpdateMemory retains the complete accepted Memory, including unchanged owned
// Links, in the same transaction as its properties and revision. This preview
// operation establishes no public History ordering or change-context contract.
func (s *Store) UpdateMemory(ctx context.Context, request MemoryUpdateRequest) (MemoryMutationResult, error) {
	return s.writeMemory(ctx, memoryWriteRequest{
		typeURL: request.TypeURL, completeProperties: &request.Properties,
		path: request.Path, actor: request.Actor, expectedRevision: request.ExpectedRevision,
		unconditional: request.Unconditional, title: request.Properties.Title, body: request.Properties.Body,
		hasTitle: true, hasBody: true, metadataPatch: request.Metadata,
	})
}

// MemoryPatchRequest changes only supplied fields. A nil pointer preserves the
// checked predecessor's field; a pointer to an empty string clears that field.
// At least one field must be supplied. This is an internal preview API.
type MemoryPatchRequest struct {
	TypeURL            string
	PropertiesJSON     json.RawMessage
	PropertiesMerge    json.RawMessage
	Path               string
	Title              *string
	Body               *string
	PropertiesProvided bool
	Actor              string
	ExpectedRevision   string
	Unconditional      bool
	Metadata           publicops.MetadataPatch
}

// PatchMemory resolves omitted fields from the actual predecessor inside the
// mutation transaction, under the same guard and retention rules as UpdateMemory.
func (s *Store) PatchMemory(ctx context.Context, request MemoryPatchRequest) (MemoryMutationResult, error) {
	patch := memoryWriteRequest{path: request.Path, actor: request.Actor,
		typeURL:          request.TypeURL,
		expectedRevision: request.ExpectedRevision, unconditional: request.Unconditional, metadataPatch: request.Metadata,
		propertiesProvided: request.PropertiesProvided}
	if len(request.PropertiesMerge) != 0 {
		if len(request.PropertiesJSON) != 0 || request.Title != nil || request.Body != nil {
			return MemoryMutationResult{}, fmt.Errorf("%w: property merge cannot accompany replacement fields", storage.ErrValidation)
		}
		canonical, err := graph.CanonicalizeJSON(request.PropertiesMerge)
		if err != nil {
			return MemoryMutationResult{}, err
		}
		if len(canonical) == 0 || canonical[0] != '{' {
			return MemoryMutationResult{}, fmt.Errorf("%w: property merge requires an object", storage.ErrValidation)
		}
		patch.propertiesMerge = string(canonical)
		patch.propertiesProvided = true
	}
	if len(request.PropertiesJSON) != 0 {
		if request.Title != nil || request.Body != nil {
			return MemoryMutationResult{}, fmt.Errorf("%w: complete properties cannot accompany individual fields", storage.ErrValidation)
		}
		var complete Properties
		if err := json.Unmarshal(request.PropertiesJSON, &complete); err != nil {
			return MemoryMutationResult{}, err
		}
		patch.completeProperties = &complete
	}
	// Capture caller-owned pointers before entering the transaction. No pointer is
	// retained or read by the transaction callback.
	if request.Title != nil {
		patch.title, patch.hasTitle = *request.Title, true
	}
	if request.Body != nil {
		patch.body, patch.hasBody = *request.Body, true
	}
	return s.writeMemory(ctx, patch)
}

type memoryWriteRequest struct {
	path, actor, expectedRevision string
	unconditional                 bool
	title, body                   string
	typeURL                       string
	completeProperties            *Properties
	propertiesMerge               string
	hasTitle, hasBody             bool
	propertiesPatch               *graphpatch.Patch
	propertiesProvided            bool
	metadataPatch                 publicops.MetadataPatch
}

func (s *Store) writeMemory(ctx context.Context, request memoryWriteRequest) (MemoryMutationResult, error) {
	if err := validatePath(request.path); err != nil {
		return MemoryMutationResult{}, fmt.Errorf("%w: %v", storage.ErrValidation, err)
	}
	if request.typeURL == "" && request.completeProperties == nil && !request.hasTitle && !request.hasBody && request.propertiesPatch == nil && !request.propertiesProvided && !hasCommonMetadataPatch(request.metadataPatch) {
		return MemoryMutationResult{}, fmt.Errorf("%w: at least one Memory field must be supplied", storage.ErrValidation)
	}
	if request.propertiesPatch != nil && (request.hasTitle || request.hasBody) {
		return MemoryMutationResult{}, fmt.Errorf("%w: ordered Memory properties patch cannot be combined with field replacement", storage.ErrValidation)
	}
	if err := validateCommonMetadataPatch(request.metadataPatch); err != nil {
		return MemoryMutationResult{}, err
	}
	if !utf8.ValidString(request.title) || !utf8.ValidString(request.body) || !utf8.ValidString(request.actor) {
		return MemoryMutationResult{}, fmt.Errorf("%w: Memory title, body and actor must be UTF-8", storage.ErrValidation)
	}
	var result MemoryMutationResult
	err := s.withTypeValidationTx(ctx, func(tx *sql.Tx) error {
		if err := checkBinding(ctx, tx, s.options); err != nil {
			return err
		}
		current, revision, err := s.beadEndpointInTx(ctx, tx, request.path)
		if err != nil {
			return err
		}
		memory, ok := current.(Record)
		if !ok {
			return fmt.Errorf("%w: property replacement supports the experimental Memory Type only", ErrCapabilityUnavailable)
		}
		if err := checkRevisionGuard(request.expectedRevision, request.unconditional, revision, true, "Memory"); err != nil {
			return err
		}
		next := memory.Properties
		nextType := memory.Type
		if request.typeURL != "" {
			nextType = request.typeURL
		}
		if request.completeProperties != nil {
			next = *request.completeProperties
		}
		if request.propertiesMerge != "" {
			next, err = mergeMemoryProperties(next, []byte(request.propertiesMerge))
			if err != nil {
				return err
			}
		}
		if request.hasTitle {
			next.Title = request.title
		}
		if request.hasBody {
			next.Body = request.body
		}
		if request.propertiesPatch != nil {
			next, err = applyMemoryPropertiesPatch(request.propertiesPatch, memory.Properties)
			if err != nil {
				return err
			}
		}
		metadata, metadataChanged, err := issueops.ApplyMetadataPatch(memory.Metadata, request.metadataPatch)
		if err != nil {
			return err
		}
		metadata, err = commonMetadata(metadata)
		if err != nil {
			return err
		}
		properties, err := canonicalJSON(next)
		if err != nil {
			return err
		}
		catalog, err := s.selectedTypes()
		if err != nil {
			return err
		}
		if !catalog.memoryType(nextType) {
			return fmt.Errorf("%w: target Type requires its native domain writer", storage.ErrValidation)
		}
		if err := catalog.validateProperties(nextType, properties, graph.KindBead); err != nil {
			return err
		}
		if catalog.evaluationType(nextType) {
			if catalog.evaluationType(memory.Type) {
				if err := preserveEvaluationSubjectGroup(memory.Properties, next); err != nil {
					return err
				}
			} else if err := s.preserveRetainedEvaluationGroup(ctx, tx, request.path, revision, next); err != nil {
				return err
			}
		}
		if memory.Properties == next && memory.Type == nextType && !metadataChanged {
			result = MemoryMutationResult{Memory: memory}
			return nil
		}
		if err := s.touchCoordination(ctx, tx); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE graph_preview_payloads SET properties=?, metadata=? WHERE path=?`, properties, metadata, request.path); err != nil {
			return err
		}
		if err := s.afterStage("memory-payload"); err != nil {
			return err
		}
		replaced := replacedMemory(memory, request.unconditional)
		memory.Properties = next
		memory.Type = nextType
		if _, err := tx.ExecContext(ctx, `UPDATE graph_preview_catalog SET type_url=? WHERE path=?`, nextType, request.path); err != nil {
			return err
		}
		memory.Metadata = metadata
		accepted, err := s.recordOwnedMemoryInTx(ctx, tx, request.path, request.actor, memory)
		if err != nil {
			return err
		}
		// New ordered patches must not publish a state that exceeds the existing
		// current-read acquisition budget. Existing replacement/field routes retain
		// their prior admission behavior. A no-op above writes nothing.
		if request.propertiesPatch != nil || hasCommonMetadataPatch(request.metadataPatch) {
			if err := checkCurrentReadBytes(ctx, tx); err != nil {
				return err
			}
		}
		result = MemoryMutationResult{Memory: accepted.(Record), Changed: true, Replaced: replaced}
		return nil
	})
	if err != nil {
		return MemoryMutationResult{}, err
	}
	return result, nil
}

// Claims belong to one collective group. Changing ordinary Evaluation content
// may preserve or reorder that group, but may not retarget existing claims.
func preserveEvaluationSubjectGroup(previous, next Properties) error {
	group := func(properties Properties) (map[exactSubjectBinding]struct{}, error) {
		var fields struct {
			Subjects []exactSubjectBinding `json:"subjects"`
		}
		if err := json.Unmarshal([]byte(properties.extra), &fields); err != nil || len(fields.Subjects) == 0 {
			return nil, fmt.Errorf("%w: Evaluation requires exact subject bindings", storage.ErrValidation)
		}
		members := make(map[exactSubjectBinding]struct{}, len(fields.Subjects))
		for _, subject := range fields.Subjects {
			members[subject] = struct{}{}
		}
		return members, nil
	}
	before, err := group(previous)
	if err != nil {
		return err
	}
	after, err := group(next)
	if err != nil {
		return err
	}
	if len(before) == len(after) {
		same := true
		for member := range before {
			if _, present := after[member]; !present {
				same = false
				break
			}
		}
		if same {
			return nil
		}
	}
	return fmt.Errorf("%w: a different subject group requires a separate Evaluation", storage.ErrValidation)
}

// Retyping away from Evaluation does not release the identity's first subject
// group. Ordinary Memory extension fields never establish that binding. Retained
// history is the authority; incomplete or oversized history refuses admission.
func (s *Store) preserveRetainedEvaluationGroup(ctx context.Context, tx *sql.Tx, path, head string, next Properties) error {
	versions, err := previewVersionsInTx(ctx, tx, path, head, "", shippedVersionBounds())
	if err != nil {
		return err
	}
	var size uint64
	if err := tx.QueryRowContext(ctx, `SELECT CAST(LEAST(COALESCE(SUM(OCTET_LENGTH(snapshot)),0),?) AS UNSIGNED)
 FROM graph_preview_versions WHERE path=?`, PreviewCurrentReadByteLimit+1, path).Scan(&size); err != nil {
		return err
	}
	if size > PreviewCurrentReadByteLimit {
		return fmt.Errorf("%w: Evaluation binding history exceeds acquisition budget", ErrLimitExceeded)
	}
	catalog, err := s.selectedTypes()
	if err != nil {
		return err
	}
	for i := len(versions) - 1; i >= 0; i-- {
		version := versions[i].Version
		raw, actor, err := readVersionBytes(ctx, tx, path, version, head)
		if err != nil {
			return err
		}
		memory, err := s.decodeMemoryVersion(ctx, tx, path, "", version, raw, actor)
		if err != nil {
			return err
		}
		if catalog.evaluationType(memory.Type) {
			return preserveEvaluationSubjectGroup(memory.Properties, next)
		}
	}
	return nil
}
