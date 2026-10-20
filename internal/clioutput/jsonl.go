package clioutput

import (
	"encoding/json"
	"fmt"
	"io"
)

// JSONLWriter owns successful delivery accounting and stream lifecycle, but not
// fetching or process status. A failed output write poisons the stream: attempting
// another record could append valid JSON after a partial line. It is not safe for
// concurrent use. The caller must also check its writer's flush/close errors.
type JSONLWriter struct {
	renderer *CollectionRenderer
	output   io.Writer
	returned int
	closed   bool
	writeErr error
}

// NewJSONL validates before writing the one header (unless raw). The writer accepts
// projected delivery units through Item and exactly one success or failure terminal.
// No outcome is claimed by the header. This initial stream contract admits only
// resource collections: a version/row adapter must specify how native enclosing
// subject/read context travels before those result kinds become streamable.
func (r *CollectionRenderer) NewJSONL(output io.Writer) (*JSONLWriter, error) {
	if r.format != "jsonl" {
		return nil, fmt.Errorf("JSONL writer requires JSONL format")
	}
	if r.kind != "resource_collection" || len(r.contextMembers) != 0 {
		return nil, fmt.Errorf("JSONL requires an admitted resource collection stream contract")
	}
	if output == nil {
		return nil, fmt.Errorf("JSONL requires an output writer")
	}
	stream := &JSONLWriter{renderer: r, output: output}
	if !r.raw {
		err := stream.write(struct {
			RecordType    string              `json:"record_type"`
			SchemaVersion int                 `json:"schema_version"`
			BDPSpec       string              `json:"bdp_spec"`
			ResultKind    string              `json:"result_kind"`
			Projection    *projectionMetadata `json:"projection,omitempty"`
		}{"header", 2, r.item.bdpSpec, r.kind, r.item.projection})
		if err != nil {
			return nil, err
		}
	}
	return stream, nil
}

// Item writes exactly one complete item record. Projection failure emits nothing
// and permits Failure; an IO error prevents any subsequent stream writes.
func (s *JSONLWriter) Item(item json.RawMessage) error {
	if err := s.open(); err != nil {
		return err
	}
	projected, err := s.renderer.item.Render(item)
	if err != nil {
		return err
	}
	var record any = json.RawMessage(projected)
	if !s.renderer.raw {
		record = struct {
			RecordType string          `json:"record_type"`
			Data       json.RawMessage `json:"data"`
		}{"item", projected}
	}
	if err := s.write(record); err != nil {
		return err
	}
	s.returned++
	return nil
}

// Returned counts only complete item records accepted by the output writer.
// A buffered writer's later flush error still makes process completion fail.
func (s *JSONLWriter) Returned() int { return s.returned }

// Finish closes a successfully handled page, limit or exhausted collection.
// Continuation is a protocol JSON value at the delivered boundary; the renderer
// neither interprets nor fabricates it. Nil means unavailable, not exhausted.
func (s *JSONLWriter) Finish(exhausted bool, stopReason string, continuation json.RawMessage) error {
	c := Completion{Returned: s.returned, Exhausted: &exhausted, StopReason: stopReason}
	if err := validateCompletion(c, s.returned, false); err != nil {
		return err
	}
	return s.terminal("success", c, continuation, nil)
}

// Failure retains prior items and, when framed, emits an error terminal. Raw
// streams emit no exceptional data record. In both cases callers must report
// diagnostics and nonzero status; successful terminal delivery is not success of
// the operation. Exhaustion is unknown because execution did not finish.
func (s *JSONLWriter) Failure(origin string, problem json.RawMessage) error {
	return s.FailureWithContinuation(origin, problem, nil)
}

// FailureWithContinuation preserves an available provider continuation at the
// last successfully delivered boundary. The caller must not pass the next-page
// token when part of that page was not delivered. Nil reports no continuation.
func (s *JSONLWriter) FailureWithContinuation(origin string, problem, continuation json.RawMessage) error {
	wrapper, err := collectionProblem(origin, problem)
	if err != nil {
		return err
	}
	return s.terminal("failure", Completion{Returned: s.returned, StopReason: "error"}, continuation, wrapper)
}

func (s *JSONLWriter) terminal(outcome string, completion Completion, continuation json.RawMessage, problem *problemWrapper) error {
	if err := s.open(); err != nil {
		return err
	}
	if continuation != nil && !json.Valid(continuation) {
		return fmt.Errorf("continuation must be a protocol JSON value")
	}
	if !s.renderer.raw {
		if err := s.write(struct {
			RecordType   string          `json:"record_type"`
			Outcome      string          `json:"outcome"`
			Completion   Completion      `json:"completion"`
			Continuation json.RawMessage `json:"continuation,omitempty"`
			Error        *problemWrapper `json:"error,omitempty"`
		}{"terminal", outcome, completion, continuation, problem}); err != nil {
			return err
		}
	}
	s.closed = true
	return nil
}

func (s *JSONLWriter) open() error {
	if s.writeErr != nil {
		return s.writeErr
	}
	if s.closed {
		return fmt.Errorf("JSONL stream is already complete")
	}
	return nil
}

func (s *JSONLWriter) write(value any) error {
	record, err := document(value)
	if err != nil {
		return err
	}
	n, err := s.output.Write(record)
	if err == nil && n != len(record) {
		err = io.ErrShortWrite
	}
	if err != nil {
		s.writeErr = err
	}
	return err
}
