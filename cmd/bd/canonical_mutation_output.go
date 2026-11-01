package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/clioutput"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/resourceget"
)

// This is an output adapter, not the root-owned store completion hook. The root
// passes the ORIGINAL callback value/human/op error and checked Close error to
// Complete exactly once. It must construct this adapter before the operation.
// There is no operation callback, store/read API, recovery key or retry here.
type canonicalMutationOutputShape string

const (
	canonicalMutationOutputRetained      canonicalMutationOutputShape = "retained-singleton"
	canonicalMutationOutputRetainedBatch canonicalMutationOutputShape = "retained-batch1"
	canonicalMutationOutputBead          canonicalMutationOutputShape = "bead-resource"
	canonicalMutationOutputLink          canonicalMutationOutputShape = "link-resource"
	canonicalMutationOutputMemory        canonicalMutationOutputShape = "memory-mutation"
	canonicalMutationOutputIssue         canonicalMutationOutputShape = "issue-mutation"
	canonicalMutationOutputLinkChange    canonicalMutationOutputShape = "link-mutation"
	canonicalMutationOutputDependency    canonicalMutationOutputShape = "dependency-mutation"
	canonicalMutationOutputMemoryEnd     canonicalMutationOutputShape = "memory-delete"
	canonicalMutationOutputIssueEnd      canonicalMutationOutputShape = "issue-delete"
)

var errCanonicalMutationOutputConsumed = errors.New("mutation output completion already consumed")
var errCanonicalMutationOutputCapture = errors.New("selected mutation output lacks a checked captured Resource")

// These are operation-return dispositions, not claims that a value proves
// commit, that cleanup succeeded, or that a durable retry record exists.
type canonicalMutationOutputDisposition string

const (
	canonicalMutationOutputReturned canonicalMutationOutputDisposition = "operation-returned-successfully"
	canonicalMutationOutputFailed   canonicalMutationOutputDisposition = "operation-returned-error"
	canonicalMutationOutputUnknown  canonicalMutationOutputDisposition = "operation-outcome-unknown"
)

type canonicalMutationOutputResult struct {
	Value        any
	Human        string
	OperationErr error
	ExchangeErr  error
	ReadErr      error
	LimitErr     error
	CloseErr     error
	PayloadErr   error
	RenderErr    error
	OutputErr    error
	Disposition  canonicalMutationOutputDisposition
}

// Err retains independent causes. Root classification must inspect OperationErr
// separately: a Close, payload or output error cannot change its known outcome.
func (r canonicalMutationOutputResult) Err() error {
	return errors.Join(r.OperationErr, r.ExchangeErr, r.ReadErr, r.LimitErr, r.CloseErr, r.PayloadErr, r.RenderErr, r.OutputErr)
}

type canonicalMutationOutput struct {
	mu       sync.Mutex
	shape    canonicalMutationOutputShape
	renderer *clioutput.MutationRenderer
	out      io.Writer
	used     bool
	result   canonicalMutationOutputResult
}

// The descriptor names the TRUE selected native payload contract, not a newly
// asserted public Write profile. Required fields must match this closed shape.
// The existing native wrapper keys and deletion evidence remain native; only
// checked producer-captured Resource members replace legacy display DTOs.
func newCanonicalMutationOutput(shape canonicalMutationOutputShape, descriptor clioutput.MutationDescriptor, options clioutput.Options, out io.Writer) (*canonicalMutationOutput, error) {
	required, ok := canonicalMutationOutputMembers(shape)
	if !ok || out == nil {
		return nil, fmt.Errorf("unsupported mutation output shape or missing writer")
	}
	if len(descriptor.Required) != len(required) {
		return nil, fmt.Errorf("mutation descriptor does not declare the selected native shape")
	}
	seen := make(map[string]bool, len(descriptor.Required))
	for _, name := range descriptor.Required {
		seen[name] = true
	}
	for _, name := range required {
		if !seen[name] {
			return nil, fmt.Errorf("mutation descriptor is missing %q", name)
		}
	}
	renderer, err := clioutput.NewMutation(descriptor, options)
	if err != nil {
		return nil, err
	}
	return &canonicalMutationOutput{shape: shape, renderer: renderer, out: out}, nil
}

func canonicalMutationOutputMembers(shape canonicalMutationOutputShape) ([]string, bool) {
	switch shape {
	case canonicalMutationOutputRetainedBatch:
		return []string{"capabilityVersion", "changed", "results", "resources"}, true
	case canonicalMutationOutputRetained:
		return []string{"outcome", "changed"}, true
	case canonicalMutationOutputBead, canonicalMutationOutputLink:
		return []string{"id", "type", "revision"}, true
	case canonicalMutationOutputMemory:
		return []string{"memory", "changed"}, true
	case canonicalMutationOutputIssue:
		return []string{"issue", "changed"}, true
	case canonicalMutationOutputLinkChange, canonicalMutationOutputDependency:
		return []string{"link", "source", "changed"}, true
	case canonicalMutationOutputMemoryEnd:
		return []string{"memory", "preview", "deleted"}, true
	case canonicalMutationOutputIssueEnd:
		return []string{"issue", "preview", "deleted"}, true
	default:
		return nil, false
	}
}

// Complete preserves Value and Human even if payload admission or output fails.
// It buffers one document and makes at most ONE Write call. Re-entry returns the
// first completion plus an error; it never writes or executes an operation again.
// The prepared first result is published before external Write. In-flight or
// synchronous writer reentry refuses without waiting for future output status.
func (o *canonicalMutationOutput) Complete(value any, human string, operationErr, closeErr error) canonicalMutationOutputResult {
	o.mu.Lock()
	if o.used {
		result := o.result
		result.OutputErr = errors.Join(result.OutputErr, errCanonicalMutationOutputConsumed)
		o.mu.Unlock()
		return result
	}
	o.used = true
	result := canonicalMutationOutputResult{Value: value, Human: human, OperationErr: operationErr, CloseErr: closeErr}
	var document []byte
	if o.shape == canonicalMutationOutputRetained || o.shape == canonicalMutationOutputRetainedBatch {
		document, result = o.completeRetained(value, human, operationErr, closeErr)
	} else if errors.Is(operationErr, graphstore.ErrOutcomeUnknown) {
		result.Disposition = canonicalMutationOutputUnknown
		document, result.RenderErr = o.renderer.RenderUnknown("cli", canonicalMutationOutputProblem("outcome_unknown", "operation outcome is unknown; do not execute it again"))
	} else {
		result.Disposition = canonicalMutationOutputReturned
		if operationErr != nil {
			result.Disposition = canonicalMutationOutputFailed
		}
		var payload json.RawMessage
		// A captured facet precedes commit. It cannot prove publishable effects
		// after an operation error. No allocated result contract proves partial
		// effects, so only successful operation returns admit a payload.
		if operationErr == nil {
			payload, result.PayloadErr = canonicalMutationOutputPayload(o.shape, value)
		}
		switch {
		case operationErr != nil:
			problem := canonicalMutationOutputProblem("operation_error", operationErr.Error())
			document, result.RenderErr = o.renderer.RenderFailure("cli", problem)
		case result.PayloadErr != nil:
			// Successful operation return remains known. This is an output
			// capability gap, not a rollback or an unknown operation.
			document, result.RenderErr = o.renderer.RenderFailure("cli", canonicalMutationOutputProblem("output_unavailable", "operation returned successfully but its checked native result is unavailable"))
		case closeErr != nil:
			document, result.RenderErr = o.renderer.RenderFailureWithResult("cli", canonicalMutationOutputProblem("cleanup_error", "operation returned successfully but store cleanup failed"), payload)
		default:
			document, result.RenderErr = o.renderer.RenderSuccess(payload)
		}
	}
	// Cache all available first channels before calling external writer code.
	// Reentry sees a consumed refusal, not a second Write or a future IO error.
	o.result = result
	o.mu.Unlock()
	if result.RenderErr == nil && len(document) != 0 {
		n, err := o.out.Write(document)
		if err == nil && n != len(document) {
			err = io.ErrShortWrite
		}
		result.OutputErr = err
	}
	o.mu.Lock()
	o.result = result
	o.mu.Unlock()
	return result
}

func canonicalMutationOutputProblem(code, message string) json.RawMessage {
	// Fixed string-valued CLI problem fields are always JSON encodable.
	raw, _ := json.Marshal(struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}{code, message})
	return raw
}

// Admission is closed over the actual Source07 writer result types. Arbitrary
// ResourceDocument values, interfaces, maps, raw JSON and caller-built wrappers
// are NOT accepted as producer authority. Pointer/legacy result forms also
// refuse. Root must preserve typed writer results before legacy projection.
func canonicalMutationOutputPayload(shape canonicalMutationOutputShape, value any) (json.RawMessage, error) {
	switch shape {
	case canonicalMutationOutputBead:
		switch value.(type) {
		case graphstore.Record, graphstore.IssueRecord:
			return canonicalMutationCapturedResource(value)
		}
	case canonicalMutationOutputLink:
		if _, ok := value.(graphstore.LinkRecord); ok {
			return canonicalMutationCapturedResource(value)
		}
	case canonicalMutationOutputMemory:
		if r, ok := value.(graphstore.MemoryMutationResult); ok {
			if r.Replaced != nil {
				return nil, fmt.Errorf("%w: replacement evidence has no selected checked capture mapping", errCanonicalMutationOutputCapture)
			}
			return canonicalMutationCapturedMembers(r, map[string]any{"memory": r.Memory})
		}
	case canonicalMutationOutputIssue:
		if r, ok := value.(graphstore.IssueMutationResult); ok {
			return canonicalMutationCapturedMembers(r, map[string]any{"issue": r.Issue})
		}
	case canonicalMutationOutputLinkChange:
		if r, ok := value.(graphstore.LinkMutationResult); ok {
			if r.ReplacedSource != nil {
				return nil, fmt.Errorf("%w: replaced source has no selected checked capture mapping", errCanonicalMutationOutputCapture)
			}
			return canonicalMutationCapturedMembers(r, map[string]any{"link": r.Link, "source": r.Source})
		}
	case canonicalMutationOutputDependency:
		if r, ok := value.(graphstore.DependencyResult); ok {
			// Target is deliberately json:"-" in the actual native contract.
			return canonicalMutationCapturedMembers(r, map[string]any{"link": r.Link, "source": r.Source})
		}
	case canonicalMutationOutputMemoryEnd:
		if r, ok := value.(graphstore.MemoryDeleteResult); ok {
			// Keep final-live Memory separate from original terminal evidence.
			return canonicalMutationCapturedMembers(r, map[string]any{"memory": r.Memory})
		}
	case canonicalMutationOutputIssueEnd:
		if r, ok := value.(graphstore.IssueDeleteResult); ok {
			return canonicalMutationCapturedMembers(r, map[string]any{"issue": r.Issue})
		}
	}
	// LinkDeleteResult and DependencyRemoveResult expose tombstones without a
	// complete captured Resource. Do not reconstruct a postimage from them.
	return nil, fmt.Errorf("%w: unsupported %s result %T", errCanonicalMutationOutputCapture, shape, value)
}

func canonicalMutationCapturedMembers(value any, members map[string]any) (json.RawMessage, error) {
	// Extract EVERY required affected Resource before encoding any wrapper.
	captured := make(map[string]json.RawMessage, len(members))
	for name, member := range members {
		if name == "source" {
			switch member.(type) {
			case graphstore.Record, graphstore.IssueRecord:
			default:
				return nil, fmt.Errorf("%w: required affected source must be a captured Bead, got %T", errCanonicalMutationOutputCapture, member)
			}
		}
		raw, err := canonicalMutationCapturedResource(member)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		captured[name] = raw
	}
	raw, err := canonicalMutationMarshalNative(value)
	if err != nil {
		return nil, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil, err
	}
	for name, native := range captured {
		if _, present := fields[name]; !present {
			return nil, fmt.Errorf("selected native wrapper lacks %s", name)
		}
		fields[name] = native
	}
	return canonicalMutationMarshalNative(fields)
}

// Match clioutput’s existing encoder policy. Default json.Marshal would HTML-
// escape RawMessage bodies and opaque tokens before the renderer sees them.
func canonicalMutationMarshalNative(value any) ([]byte, error) {
	var out bytes.Buffer
	encoder := json.NewEncoder(&out)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return bytes.TrimSuffix(out.Bytes(), []byte("\n")), nil
}

func canonicalMutationCapturedResource(value any) (json.RawMessage, error) {
	var kind graph.ResourceKind
	var id, revision, typeURI string
	var document resourceget.ResourceDocument
	var ok bool
	switch r := value.(type) {
	case graphstore.Record:
		kind, id, revision, typeURI = graph.KindBead, r.ID, r.Revision, r.Type
		document, ok = r.ResourceDocument()
	case graphstore.IssueRecord:
		kind, id, revision, typeURI = graph.KindBead, r.ID, r.Revision, r.Type
		document, ok = r.ResourceDocument()
	case graphstore.LinkRecord:
		kind, id, revision, typeURI = graph.KindLink, r.ID, r.Revision, r.Type
		document, ok = r.ResourceDocument()
	default:
		return nil, fmt.Errorf("%w: required Resource has unsupported type %T", errCanonicalMutationOutputCapture, value)
	}
	return canonicalMutationCapturedDocument(kind, id, revision, typeURI, document, ok)
}

// This checks representation correspondence, not producer lineage by itself.
// Only canonicalMutationCapturedResource's CLOSED actual-record switch supplies
// production facts. Tests may construct shape fixtures for this pure comparison;
// that never demonstrates a checked14 write or authorizes a caller-built DTO.
func canonicalMutationCapturedDocument(kind graph.ResourceKind, id, revision, typeURI string, document resourceget.ResourceDocument, present bool) (json.RawMessage, error) {
	if !present || id == "" || revision == "" || typeURI == "" ||
		(kind != graph.KindBead && kind != graph.KindLink) ||
		document.Kind() != kind || document.ID() != id || document.Revision() != revision ||
		document.Type().URI != typeURI || document.Type().Revision == "" {
		return nil, errCanonicalMutationOutputCapture
	}
	native := document.NativeJSON()
	marshaled, err := document.MarshalJSON()
	if err != nil || len(native) == 0 || !bytes.Equal(native, marshaled) {
		return nil, fmt.Errorf("%w: captured native bytes are unavailable", errCanonicalMutationOutputCapture)
	}
	return json.RawMessage(native), nil
}

// Retained HTTP completion keeps original status/body/Problem and independent
// exchange, parse, Close, render and Write causes in the first result. It never
// changes a known operation into unknown because cleanup or output failed.
func (o *canonicalMutationOutput) completeRetained(value any, human string, operationErr, closeErr error) ([]byte, canonicalMutationOutputResult) {
	result := canonicalMutationOutputResult{Value: value, Human: human, OperationErr: operationErr, CloseErr: closeErr}
	reply, ok := value.(canonicalRetainedReply)
	if !ok {
		result.PayloadErr = errCanonicalMutationOutputCapture
		result.Disposition = canonicalMutationOutputFailed
		raw, err := o.renderer.RenderFailure("cli", canonicalMutationOutputProblem("output_unavailable", "missing original retained response"))
		result.RenderErr = err
		return raw, result
	}
	result.ExchangeErr, result.ReadErr, result.LimitErr = reply.ExchangeErr, reply.ReadErr, reply.LimitErr
	result.PayloadErr = reply.ParseErr
	var raw []byte
	switch {
	case len(reply.Mutation) != 0 && reply.Checked:
		result.Disposition = canonicalMutationOutputReturned
		if closeErr != nil {
			raw, result.RenderErr = o.renderer.RenderFailureWithResult("cli", canonicalMutationOutputProblem("cleanup_error", "retained operation returned successfully but response cleanup failed"), reply.Mutation)
		} else {
			raw, result.RenderErr = o.renderer.RenderSuccess(reply.Mutation)
		}
	case len(reply.Problem) != 0:
		result.Disposition = canonicalMutationOutputFailed
		// The original native Problem carries its own retry advice. Do not assert
		// durable key retention, pre-admission refusal or rollback from HTTP alone.
		raw, result.RenderErr = o.renderer.RenderFailure("bdp", reply.Problem)
	case reply.Sent:
		result.Disposition = canonicalMutationOutputUnknown
		raw, result.RenderErr = o.renderer.RenderUnknown("cli", canonicalMutationOutputProblem("outcome_unknown", "no checked retained outcome; recover only with the saved same invocation"))
	default:
		result.Disposition = canonicalMutationOutputFailed
		message := "retained request refused before submission"
		if operationErr != nil {
			message = operationErr.Error()
		}
		raw, result.RenderErr = o.renderer.RenderFailure("cli", canonicalMutationOutputProblem("request_refused", message))
	}
	return raw, result
}
