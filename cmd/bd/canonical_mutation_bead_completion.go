package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/steveyegge/beads/internal/clioutput"
)

// This is a prepared output completion, not an operation executor or the
// root-owned store-lifetime hook. Complete receives the original typed callback
// value after the single checked Close. It never opens a store or retries.
type canonicalBeadMutationCompletion struct {
	mu       sync.Mutex
	mapper   *canonicalBeadMutationMapper
	renderer *clioutput.MutationRenderer
	out      io.Writer
	used     bool
	result   canonicalBeadMutationProjection
}

// Prepare before acquisition or effects. The mapper is CLOSED over admitted
// actual writer result types; no arbitrary payload/projector callback is taken.
// Existing renderer capabilities remain detail/JSON only. This adapter does not
// claim the remaining canonical default/context/status/JSONL framing surface.
func newCanonicalBeadMutationCompletion(operation string, options clioutput.Options, out io.Writer) (*canonicalBeadMutationCompletion, error) {
	if out == nil {
		return nil, fmt.Errorf("mutation completion requires an output writer")
	}
	mapper, err := newCanonicalBeadMutationMapper(operation)
	if err != nil {
		return nil, err
	}
	renderer, err := clioutput.NewMutation(mapper.Descriptor(), options)
	if err != nil {
		return nil, err
	}
	return &canonicalBeadMutationCompletion{mapper: mapper, renderer: renderer, out: out}, nil
}

// Complete retains the original Value and all independent causes alongside the
// separate public Payload. A successful operation return stays known even when
// capture, Close, rendering or output fails. Unknown comes only from the
// original operation error through mapper02, never the joined diagnostics.
// One buffered document is written at most once, including concurrent reentry.
func (c *canonicalBeadMutationCompletion) Complete(value any, human string, operationErr, closeErr error) canonicalBeadMutationProjection {
	result, document, first := c.prepare(value, human, operationErr, closeErr)
	if !first {
		return result
	}
	// Publish the prepared result before calling external output code. A
	// writer may reenter Complete synchronously; holding mu across Write
	// would deadlock instead of returning the consumed refusal. An in-flight
	// duplicate gets the available first-result causes, never a second Write.
	if result.RenderErr == nil && len(document) != 0 {
		n, err := c.out.Write(document)
		if err == nil && n != len(document) {
			err = io.ErrShortWrite
		}
		result.OutputErr = err
	}
	c.mu.Lock()
	c.result = canonicalBeadMutationCompletionCopy(result)
	c.mu.Unlock()
	return result
}

func (c *canonicalBeadMutationCompletion) prepare(value any, human string, operationErr, closeErr error) (canonicalBeadMutationProjection, []byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.used {
		result := canonicalBeadMutationCompletionCopy(c.result)
		result.OutputErr = errors.Join(result.OutputErr, errCanonicalMutationOutputConsumed)
		return result, nil, false
	}
	c.used = true
	result := c.mapper.Project(value, human, operationErr, closeErr)
	var document []byte
	switch result.Disposition {
	case canonicalMutationOutputUnknown:
		document, result.RenderErr = c.renderer.RenderUnknown("cli", canonicalMutationOutputProblem("outcome_unknown", "operation outcome is unknown; do not execute it again"))
	case canonicalMutationOutputFailed:
		document, result.RenderErr = c.renderer.RenderFailure("cli", canonicalMutationOutputProblem("operation_error", operationErr.Error()))
	default:
		switch {
		case result.PayloadErr != nil:
			document, result.RenderErr = c.renderer.RenderFailure("cli", canonicalMutationOutputProblem("output_unavailable", "operation returned successfully but its checked result is unavailable"))
		case closeErr != nil:
			document, result.RenderErr = c.renderer.RenderFailureWithResult("cli", canonicalMutationOutputProblem("cleanup_error", "operation returned successfully but store cleanup failed"), result.Payload)
		default:
			document, result.RenderErr = c.renderer.RenderSuccess(result.Payload)
		}
	}
	// Defend the retained raw payload from mutation by a result recipient.
	// Value deliberately remains the exact callback value, not a reconstructed
	// resource or a replacement with the projected wire payload.
	c.result = canonicalBeadMutationCompletionCopy(result)
	return result, document, true
}

func canonicalBeadMutationCompletionCopy(result canonicalBeadMutationProjection) canonicalBeadMutationProjection {
	result.Payload = bytes.Clone(result.Payload)
	return result
}
