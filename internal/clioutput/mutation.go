package clioutput

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// MutationDescriptor identifies the exact contract of the operation result in
// data. The adapter owns that result's schema and supplies the selected pin;
// this renderer only frames it and never infers whether a write was applied.
type MutationDescriptor struct {
	BDPSpec  string
	Required []string
}

// MutationRenderer buffers a complete operation result before the caller writes
// it. It does not execute, retry, or verify a mutation's transaction outcome.
type MutationRenderer struct {
	bdpSpec  string
	format   string
	raw      bool
	required []string
}

// NewMutation validates output options before the operation begins. Mutation
// projection is not admitted until an adapter declares protected postimage and
// effect roots; only the complete native result or its envelope can be selected.
func NewMutation(descriptor MutationDescriptor, options Options) (*MutationRenderer, error) {
	if strings.TrimSpace(descriptor.BDPSpec) == "" {
		return nil, fmt.Errorf("an exact BDP payload contract identifier is required")
	}
	if len(descriptor.Required) == 0 {
		return nil, fmt.Errorf("mutation output requires declared result members")
	}
	required := make([]string, 0, len(descriptor.Required))
	seen := make(map[string]bool, len(descriptor.Required))
	for _, member := range descriptor.Required {
		if strings.TrimSpace(member) == "" || seen[member] {
			return nil, fmt.Errorf("invalid required mutation member %q", member)
		}
		seen[member] = true
		required = append(required, member)
	}
	format, err := resolveFormat(options.Formats, options.JSON)
	if err != nil {
		return nil, err
	}
	if len(options.Include) != 0 {
		return nil, fmt.Errorf("mutation output does not yet admit field projection")
	}
	raw := false
	for _, path := range options.Omit {
		if path != "envelope" {
			return nil, fmt.Errorf("mutation output does not yet admit field projection")
		}
		raw = true
	}
	if raw && format != "json" {
		return nil, fmt.Errorf("omit envelope requires JSON output")
	}
	return &MutationRenderer{bdpSpec: descriptor.BDPSpec, format: format, raw: raw, required: required}, nil
}

// RenderSuccess returns a complete successful document. Applied versus no-op
// and any predecessor or affected-source values remain in the native result.
// A nil, scalar, or array result is refused rather than inventing a write DTO.
func (r *MutationRenderer) RenderSuccess(result json.RawMessage) ([]byte, error) {
	value, err := mutationObject(result)
	if err != nil {
		return nil, err
	}
	if err := r.validateResult(value); err != nil {
		return nil, err
	}
	if r.format == "detail" {
		return renderDetail(value)
	}
	if r.raw {
		return document(value)
	}
	return document(envelope{
		SchemaVersion: SchemaVersion, BDPSpec: r.bdpSpec, ResultKind: "mutation",
		Outcome: "success", Data: value,
	})
}

// RenderFailure frames a known failure. The caller retains responsibility for
// diagnostics and a nonzero exit status. Human and raw modes emit no stdout.
func (r *MutationRenderer) RenderFailure(origin string, problem json.RawMessage) ([]byte, error) {
	return r.renderProblem("failure", origin, problem, nil)
}

// RenderFailureWithResult preserves an available operation result on a known
// failure, such as a transaction result that reports actual partial effects.
// It never manufactures a postimage or implies that the effects were rolled back.
func (r *MutationRenderer) RenderFailureWithResult(origin string, problem, result json.RawMessage) ([]byte, error) {
	value, err := mutationObject(result)
	if err != nil {
		return nil, err
	}
	if err := r.validateResult(value); err != nil {
		return nil, err
	}
	return r.renderProblem("failure", origin, problem, value)
}

// RenderUnknown marks an indeterminate write. Callers must not automatically
// replay it; a lost commit response cannot be converted into a known failure.
func (r *MutationRenderer) RenderUnknown(origin string, problem json.RawMessage) ([]byte, error) {
	return r.renderProblem("unknown", origin, problem, nil)
}

func (r *MutationRenderer) renderProblem(outcome, origin string, problem, result json.RawMessage) ([]byte, error) {
	if origin != "bdp" && origin != "cli" {
		return nil, fmt.Errorf("problem origin must be bdp or cli")
	}
	value, err := mutationObject(problem)
	if err != nil {
		return nil, fmt.Errorf("problem: %w", err)
	}
	if origin == "cli" {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(value, &fields); err != nil {
			return nil, err
		}
		for _, name := range []string{"code", "message"} {
			var field string
			if err := json.Unmarshal(fields[name], &field); err != nil || strings.TrimSpace(field) == "" {
				return nil, fmt.Errorf("CLI problem requires nonempty string %s", name)
			}
		}
	}
	if r.raw || r.format != "json" {
		return nil, nil
	}
	return document(envelope{
		SchemaVersion: SchemaVersion, BDPSpec: r.bdpSpec, ResultKind: "mutation",
		Outcome: outcome, Data: result, Error: &problemWrapper{Origin: origin, Problem: value},
	})
}

func mutationObject(raw json.RawMessage) (json.RawMessage, error) {
	value := bytes.TrimSpace(raw)
	if len(value) == 0 || value[0] != '{' || !json.Valid(value) {
		return nil, fmt.Errorf("mutation output requires a JSON object")
	}
	return value, nil
}

func (r *MutationRenderer) validateResult(value json.RawMessage) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(value, &fields); err != nil {
		return err
	}
	for _, member := range r.required {
		if _, present := fields[member]; !present {
			return fmt.Errorf("mutation result is missing declared member %q", member)
		}
	}
	return nil
}
