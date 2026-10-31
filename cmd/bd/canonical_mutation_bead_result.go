package main

import (
	"encoding/json"
	"errors"
	"fmt"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/clioutput"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// This is the existing owner-selected PR329 native schema identity, not a
// read-framing schema, new profile, moving-main pin or runtime adoption.
const canonicalBeadMutationSchema = "https://github.com/gastownhall/bdp/schemas/bdp-v0.1.schema.json"

var errCanonicalBeadMutationResult = errors.New("selected public Bead mutation result is unavailable")

// Construct before dispatch from the actual command operation. No operation
// is inferred from a returned DTO, a changed flag or a request reconstruction.
type canonicalBeadMutationMapper struct {
	operation string
}

func newCanonicalBeadMutationMapper(operation string) (*canonicalBeadMutationMapper, error) {
	switch operation {
	case "create", "update", "delete":
		return &canonicalBeadMutationMapper{operation: operation}, nil
	default:
		return nil, fmt.Errorf("%w: unsupported operation %q", errCanonicalBeadMutationResult, operation)
	}
}

func (m *canonicalBeadMutationMapper) Descriptor() clioutput.MutationDescriptor {
	if m == nil {
		return clioutput.MutationDescriptor{}
	}
	switch m.operation {
	case "create", "update":
		return clioutput.MutationDescriptor{BDPSpec: canonicalBeadMutationSchema, Required: []string{"outcome", "changed", "resource"}}
	case "delete":
		return clioutput.MutationDescriptor{BDPSpec: canonicalBeadMutationSchema, Required: []string{"outcome", "changed", "deleted"}}
	default:
		return clioutput.MutationDescriptor{}
	}
}

// Original values/causes remain distinct from the public payload. Root owns
// store lifetime, renderer preparation, output, context and exit classification;
// it must not replace Value with Payload or infer durable recovery from this.
type canonicalBeadMutationProjection struct {
	canonicalMutationOutputResult
	Payload json.RawMessage
}

// Project performs no IO, readback, effect, rendering or retry. Unknown in the
// ORIGINAL operation error chain wins over other operation diagnoses. Close or
// output failure does not turn an already successful operation into unknown.
func (m *canonicalBeadMutationMapper) Project(value any, human string, operationErr, closeErr error) canonicalBeadMutationProjection {
	result := canonicalBeadMutationProjection{
		canonicalMutationOutputResult: canonicalMutationOutputResult{
			Value: value, Human: human, OperationErr: operationErr, CloseErr: closeErr,
			Disposition: canonicalMutationOutputReturned,
		},
	}
	if errors.Is(operationErr, graphstore.ErrOutcomeUnknown) {
		result.Disposition = canonicalMutationOutputUnknown
		return result
	}
	if operationErr != nil {
		result.Disposition = canonicalMutationOutputFailed
		return result
	}
	if m == nil {
		result.PayloadErr = fmt.Errorf("%w: mapper was not prepared", errCanonicalBeadMutationResult)
		return result
	}
	selection, err := canonicalBeadMutationSelect(m.operation, value)
	if err != nil {
		result.PayloadErr = err
		return result
	}
	// Only the CLOSED actual-result selector reaches this checked14 facet
	// extractor. A caller-built ResourceDocument/interface is not admitted.
	captured, err := canonicalMutationCapturedResource(selection.record)
	if err != nil {
		result.PayloadErr = fmt.Errorf("%w: %w", errCanonicalBeadMutationResult, err)
		return result
	}
	result.Payload, result.PayloadErr = canonicalBeadMutationWire(m.operation, selection.changed, graph.KindBead, captured)
	return result
}

type canonicalBeadMutationSelection struct {
	record  any
	changed bool
}

// DTO selection alone proves neither capture lineage nor commit. This private
// step is separately testable with ordinary DTOs; Project MUST then require the
// actual private checked14 ResourceDocument facet through the existing helper.
func canonicalBeadMutationSelect(operation string, value any) (canonicalBeadMutationSelection, error) {
	switch operation {
	case "create":
		switch value.(type) {
		case graphstore.Record, graphstore.IssueRecord:
			return canonicalBeadMutationSelection{record: value, changed: true}, nil
		}
	case "update":
		switch r := value.(type) {
		case graphstore.MemoryMutationResult:
			if r.Replaced != nil {
				return canonicalBeadMutationSelection{}, fmt.Errorf("%w: replacement variant has no selected capture mapping", errCanonicalBeadMutationResult)
			}
			return canonicalBeadMutationSelection{record: r.Memory, changed: r.Changed}, nil
		case graphstore.IssueMutationResult:
			return canonicalBeadMutationSelection{record: r.Issue, changed: r.Changed}, nil
		}
	case "delete":
		switch r := value.(type) {
		case graphstore.MemoryDeleteResult:
			if r.Preview || !r.Deleted {
				return canonicalBeadMutationSelection{}, fmt.Errorf("%w: Memory deletion was not applied", errCanonicalBeadMutationResult)
			}
			return canonicalBeadMutationSelection{record: r.Memory, changed: true}, nil
		case graphstore.IssueDeleteResult:
			if r.Preview || !r.Deleted || len(r.IncidentLinkIDs) != 0 {
				return canonicalBeadMutationSelection{}, fmt.Errorf("%w: Issue deletion is preview, unapplied or composed with incident Links", errCanonicalBeadMutationResult)
			}
			return canonicalBeadMutationSelection{record: r.Issue, changed: true}, nil
		}
	}
	return canonicalBeadMutationSelection{}, fmt.Errorf("%w: %s does not admit %T", errCanonicalBeadMutationResult, operation, value)
}

// Typed wire structures encode only the EXISTING selected schema members.
// They are not new public result types or an alternate JSON-schema framework.
type canonicalBeadMutationResourceWire struct {
	Outcome  string          `json:"outcome"`
	Changed  bool            `json:"changed"`
	Resource json.RawMessage `json:"resource"`
}

type canonicalBeadMutationIdentityWire struct {
	ID       json.RawMessage `json:"id"`
	Type     json.RawMessage `json:"type"`
	Revision json.RawMessage `json:"revision"`
}

type canonicalBeadMutationDeletedWire struct {
	ResourceKind string                            `json:"resourceKind"`
	Resource     canonicalBeadMutationIdentityWire `json:"resource"`
}

type canonicalBeadMutationDeleteWire struct {
	Outcome string                           `json:"outcome"`
	Changed bool                             `json:"changed"`
	Deleted canonicalBeadMutationDeletedWire `json:"deleted"`
}

// Representation-only serialization of an ALREADY checked full capture.
// Production calls only through Project's closed selection+facet extraction.
// Tests may supply constructor-built shape fixtures here, never to prove a
// checked producer or to bypass Project. No storage/authentication occurs here.
func canonicalBeadMutationWire(operation string, changed bool, kind graph.ResourceKind, captured json.RawMessage) (json.RawMessage, error) {
	if kind != graph.KindBead {
		return nil, fmt.Errorf("%w: a Bead operation requires an actual Bead", errCanonicalBeadMutationResult)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(captured, &fields); err != nil || fields == nil {
		return nil, fmt.Errorf("%w: captured Resource is not an object", errCanonicalBeadMutationResult)
	}
	// Exact identity bytes come ONLY from the full capture. No terminal
	// revision, local nominal Type, URI fallback or extra native wrapper field.
	for _, name := range []string{"id", "type", "revision"} {
		if len(fields[name]) == 0 {
			return nil, fmt.Errorf("%w: captured Resource lacks %s", errCanonicalBeadMutationResult, name)
		}
	}
	switch operation {
	case "create":
		if !changed {
			return nil, fmt.Errorf("%w: create requires changed=true", errCanonicalBeadMutationResult)
		}
		return canonicalMutationMarshalNative(canonicalBeadMutationResourceWire{Outcome: "created", Changed: true, Resource: captured})
	case "update":
		return canonicalMutationMarshalNative(canonicalBeadMutationResourceWire{Outcome: "updated", Changed: changed, Resource: captured})
	case "delete":
		if !changed {
			return nil, fmt.Errorf("%w: applied delete requires changed=true", errCanonicalBeadMutationResult)
		}
		return canonicalMutationMarshalNative(canonicalBeadMutationDeleteWire{
			Outcome: "deleted", Changed: true,
			Deleted: canonicalBeadMutationDeletedWire{
				ResourceKind: "bead",
				Resource:     canonicalBeadMutationIdentityWire{ID: fields["id"], Type: fields["type"], Revision: fields["revision"]},
			},
		})
	default:
		return nil, fmt.Errorf("%w: unsupported operation %q", errCanonicalBeadMutationResult, operation)
	}
}
