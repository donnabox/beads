package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/clioutput"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// Pure completion tests: ordinary DTOs deliberately lack producer-private
// checked14 capture. Successful producer/CloseWithResult and root lifecycle
// integration require a separate actual integration gate; no fixture forges it.
type beadCompletionTestWriter struct {
	mu      sync.Mutex
	calls   int
	data    []byte
	err     error
	short   bool
	entered chan struct{}
	release chan struct{}
}

func (w *beadCompletionTestWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	w.calls++
	w.data = bytes.Clone(p)
	entered, release := w.entered, w.release
	err, short := w.err, w.short
	w.mu.Unlock()
	if entered != nil {
		close(entered)
		<-release
	}
	if err != nil {
		return 0, err
	}
	if short {
		return len(p) - 1, nil
	}
	return len(p), nil
}

func (w *beadCompletionTestWriter) snapshot() (int, []byte) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.calls, bytes.Clone(w.data)
}

func beadCompletionTestPrepare(t *testing.T, options clioutput.Options, out io.Writer) *canonicalBeadMutationCompletion {
	t.Helper()
	completion, err := newCanonicalBeadMutationCompletion("update", options, out)
	if err != nil {
		t.Fatal(err)
	}
	return completion
}

func beadCompletionTestEnvelope(t *testing.T, raw []byte, outcome, code string) {
	t.Helper()
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatalf("invalid envelope: %s %v", raw, err)
	}
	var actualOutcome, spec string
	if err := json.Unmarshal(fields["outcome"], &actualOutcome); err != nil || actualOutcome != outcome {
		t.Fatalf("wrong outcome: %s %v", raw, err)
	}
	if err := json.Unmarshal(fields["bdp_spec"], &spec); err != nil || spec != canonicalBeadMutationSchema {
		t.Fatalf("wrong payload pin: %s %v", raw, err)
	}
	if data := fields["data"]; len(data) != 0 && string(data) != "null" {
		t.Fatalf("unconfirmed value promoted into data: %s", raw)
	}
	var wrapper map[string]json.RawMessage
	if err := json.Unmarshal(fields["error"], &wrapper); err != nil {
		t.Fatal(err)
	}
	var origin string
	if err := json.Unmarshal(wrapper["origin"], &origin); err != nil || origin != "cli" {
		t.Fatalf("wrong problem origin: %s %v", raw, err)
	}
	var problem map[string]json.RawMessage
	if err := json.Unmarshal(wrapper["problem"], &problem); err != nil {
		t.Fatal(err)
	}
	var actualCode string
	if err := json.Unmarshal(problem["code"], &actualCode); err != nil || actualCode != code {
		t.Fatalf("wrong problem code: %s %v", raw, err)
	}
}

func TestCanonicalBeadCompletionRejectsInvalidPreparationWithoutOutput(t *testing.T) {
	out := &beadCompletionTestWriter{}
	for _, operation := range []string{"", "Create", "link create", "batch"} {
		if completion, err := newCanonicalBeadMutationCompletion(operation, clioutput.Options{Formats: []string{"json"}}, out); err == nil || completion != nil {
			t.Fatalf("invalid operation %q admitted", operation)
		}
	}
	for _, options := range []clioutput.Options{
		{Formats: []string{"unsupported"}},
		{Formats: []string{"json"}, Include: []string{"/properties"}},
		{Formats: []string{"json"}, Omit: []string{"/properties"}},
		{Formats: []string{"detail"}, Omit: []string{"envelope"}},
	} {
		if completion, err := newCanonicalBeadMutationCompletion("create", options, out); err == nil || completion != nil {
			t.Fatalf("invalid options admitted: %#v", options)
		}
	}
	if completion, err := newCanonicalBeadMutationCompletion("delete", clioutput.Options{}, nil); err == nil || completion != nil {
		t.Fatal("nil writer admitted")
	}
	if calls, data := out.snapshot(); calls != 0 || len(data) != 0 {
		t.Fatalf("preparation wrote output: %d %s", calls, data)
	}
	// No operation/store callback exists here: caller prepares before effects.
}

func TestCanonicalBeadCompletionKnownFailureKeepsOriginalAndCloseCause(t *testing.T) {
	opErr, closeErr := errors.New("validation failed"), errors.New("checked Close failed")
	original := &struct{ Revision string }{Revision: "original R"}
	out := &beadCompletionTestWriter{}
	completion := beadCompletionTestPrepare(t, clioutput.Options{Formats: []string{"json"}}, out)
	result := completion.Complete(original, "original human", opErr, closeErr)
	if result.Value != original || result.Human != "original human" || result.OperationErr != opErr || result.CloseErr != closeErr ||
		result.Disposition != canonicalMutationOutputFailed || !errors.Is(result.Err(), opErr) || !errors.Is(result.Err(), closeErr) ||
		result.PayloadErr != nil || len(result.Payload) != 0 || result.RenderErr != nil || result.OutputErr != nil {
		t.Fatalf("failure channels collapsed: %#v / %v", result, result.Err())
	}
	calls, data := out.snapshot()
	if calls != 1 {
		t.Fatalf("wanted one buffered write, got %d", calls)
	}
	beadCompletionTestEnvelope(t, data, "failure", "operation_error")
}

func TestCanonicalBeadCompletionUnknownComesOnlyFromOriginalOperationError(t *testing.T) {
	opCause, closeCause := errors.New("response lost"), errors.New("Close failed")
	unknown := fmt.Errorf("outer: %w", errors.Join(opCause, fmt.Errorf("commit: %w", graphstore.ErrOutcomeUnknown)))
	out := &beadCompletionTestWriter{}
	completion := beadCompletionTestPrepare(t, clioutput.Options{Formats: []string{"json"}}, out)
	original := &struct{ ID string }{ID: "unconfirmed"}
	result := completion.Complete(original, "original unknown", unknown, closeCause)
	if result.Disposition != canonicalMutationOutputUnknown || result.Value != original || result.OperationErr != unknown ||
		result.CloseErr != closeCause || !errors.Is(result.Err(), opCause) || !errors.Is(result.Err(), graphstore.ErrOutcomeUnknown) ||
		!errors.Is(result.Err(), closeCause) || result.PayloadErr != nil || len(result.Payload) != 0 {
		t.Fatalf("nested original unknown lost/promoted: %#v", result)
	}
	_, data := out.snapshot()
	beadCompletionTestEnvelope(t, data, "unknown", "outcome_unknown")

	knownErr := errors.New("known operation failure")
	closeUnknown := fmt.Errorf("Close diagnostic: %w", graphstore.ErrOutcomeUnknown)
	out2 := &beadCompletionTestWriter{}
	completion2 := beadCompletionTestPrepare(t, clioutput.Options{Formats: []string{"json"}}, out2)
	known := completion2.Complete(original, "known failure", knownErr, closeUnknown)
	if known.Disposition != canonicalMutationOutputFailed || known.OperationErr != knownErr || known.CloseErr != closeUnknown ||
		!errors.Is(known.Err(), graphstore.ErrOutcomeUnknown) {
		t.Fatalf("Close sentinel reclassified original failure: %#v", known)
	}
	_, data = out2.snapshot()
	beadCompletionTestEnvelope(t, data, "failure", "operation_error")
}

func TestCanonicalBeadCompletionUncapturedKnownNoopStaysKnown(t *testing.T) {
	original := graphstore.MemoryMutationResult{Memory: graphstore.Record{
		ID: "https://example.test/work/a", Type: "https://types.example.test/Task", Revision: "original R",
	}, Changed: false}
	closeErr := fmt.Errorf("Close only: %w", graphstore.ErrOutcomeUnknown)
	out := &beadCompletionTestWriter{}
	completion := beadCompletionTestPrepare(t, clioutput.Options{Formats: []string{"json"}}, out)
	result := completion.Complete(original, "original unchanged", nil, closeErr)
	kept, ok := result.Value.(graphstore.MemoryMutationResult)
	if !ok || kept.Memory.Revision != "original R" || kept.Changed ||
		result.Human != "original unchanged" || result.OperationErr != nil || result.CloseErr != closeErr ||
		result.Disposition != canonicalMutationOutputReturned || !errors.Is(result.PayloadErr, errCanonicalMutationOutputCapture) ||
		!errors.Is(result.Err(), closeErr) || len(result.Payload) != 0 || result.RenderErr != nil || result.OutputErr != nil {
		t.Fatalf("known no-op/output gap erased, reread or made unknown: %#v", result)
	}
	calls, data := out.snapshot()
	if calls != 1 {
		t.Fatalf("output gap wrote %d times", calls)
	}
	beadCompletionTestEnvelope(t, data, "failure", "output_unavailable")
	// This unfaceted DTO proves preservation/refusal only, not native capture.
}

type beadCompletionReentrantWriter struct {
	completion *canonicalBeadMutationCompletion
	reentered  canonicalBeadMutationProjection
	calls      int
	err        error
}

func (w *beadCompletionReentrantWriter) Write(p []byte) (int, error) {
	w.calls++
	if w.calls != 1 {
		return 0, errors.New("duplicate output attempt")
	}
	w.reentered = w.completion.Complete("replacement", "replacement", graphstore.ErrOutcomeUnknown, nil)
	return 0, w.err
}

func TestCanonicalBeadCompletionReentrantWriterRefusesWithoutDeadlock(t *testing.T) {
	opErr, closeErr, writeErr := errors.New("first operation"), errors.New("first Close"), errors.New("first output")
	out := &beadCompletionReentrantWriter{err: writeErr}
	out.completion = beadCompletionTestPrepare(t, clioutput.Options{Formats: []string{"json"}}, out)
	original := &struct{ ID string }{ID: "original"}
	done := make(chan canonicalBeadMutationProjection, 1)
	go func() { done <- out.completion.Complete(original, "first human", opErr, closeErr) }()
	var first canonicalBeadMutationProjection
	select {
	case first = <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("writer reentry deadlocked instead of refusing a second completion")
	}
	if out.calls != 1 || first.Value != original || first.OperationErr != opErr || first.CloseErr != closeErr || !errors.Is(first.OutputErr, writeErr) {
		t.Fatalf("outer completion lost channels or repeated Write: %#v, calls=%d", first, out.calls)
	}
	reentered := out.reentered
	if reentered.Value != original || reentered.Human != "first human" || reentered.OperationErr != opErr || reentered.CloseErr != closeErr ||
		reentered.Disposition != canonicalMutationOutputFailed || !errors.Is(reentered.OutputErr, errCanonicalMutationOutputConsumed) || errors.Is(reentered.OutputErr, writeErr) {
		t.Fatalf("in-flight refusal replaced original channels or invented future output result: %#v", reentered)
	}
	after := out.completion.Complete(nil, "replacement", nil, nil)
	if !errors.Is(after.OutputErr, writeErr) || !errors.Is(after.OutputErr, errCanonicalMutationOutputConsumed) || out.calls != 1 {
		t.Fatalf("settled reentry lost actual output cause or repeated Write: %#v", after)
	}
}

func TestCanonicalBeadCompletionDetailAndRawFailuresEmitNoStdout(t *testing.T) {
	for _, options := range []clioutput.Options{
		{Formats: []string{"detail"}},
		{Formats: []string{"json"}, Omit: []string{"envelope"}},
	} {
		for _, operationErr := range []error{errors.New("known failure"), graphstore.ErrOutcomeUnknown, nil} {
			out := &beadCompletionTestWriter{}
			completion := beadCompletionTestPrepare(t, options, out)
			result := completion.Complete(graphstore.MemoryMutationResult{}, "original", operationErr, nil)
			if result.OperationErr != operationErr || result.RenderErr != nil || result.OutputErr != nil {
				t.Fatalf("quiet mode lost channels: %#v", result)
			}
			if calls, data := out.snapshot(); calls != 0 || len(data) != 0 {
				t.Fatalf("detail/raw failure leaked stdout: %d %s", calls, data)
			}
		}
	}
}

func TestCanonicalBeadCompletionOutputFailureWritesOnceAndKeepsFirstChannels(t *testing.T) {
	for _, short := range []bool{false, true} {
		writeErr := errors.New("stdout failed")
		opErr, closeErr := errors.New("original failure"), errors.New("original Close failed")
		out := &beadCompletionTestWriter{short: short}
		if !short {
			out.err = writeErr
		}
		completion := beadCompletionTestPrepare(t, clioutput.Options{Formats: []string{"json"}}, out)
		original := &struct{ ID string }{ID: "original"}
		first := completion.Complete(original, "first human", opErr, closeErr)
		expected := writeErr
		if short {
			expected = io.ErrShortWrite
		}
		if !errors.Is(first.OutputErr, expected) || errors.Is(first.OutputErr, errCanonicalMutationOutputConsumed) ||
			first.OperationErr != opErr || first.CloseErr != closeErr || !errors.Is(first.Err(), opErr) || !errors.Is(first.Err(), closeErr) {
			t.Fatalf("write failure lost causes: %#v", first)
		}
		second := completion.Complete("replacement", "replacement", graphstore.ErrOutcomeUnknown, nil)
		third := completion.Complete(nil, "third", nil, nil)
		for _, result := range []canonicalBeadMutationProjection{second, third} {
			if result.Value != original || result.Human != "first human" || result.OperationErr != opErr || result.CloseErr != closeErr ||
				result.Disposition != canonicalMutationOutputFailed || !errors.Is(result.OutputErr, expected) ||
				!errors.Is(result.OutputErr, errCanonicalMutationOutputConsumed) || result.PayloadErr != nil || result.RenderErr != nil {
				t.Fatalf("reentry replaced first completion: %#v", result)
			}
		}
		if calls, _ := out.snapshot(); calls != 1 {
			t.Fatalf("failed delivery wrote again: %d", calls)
		}
		// No execution callback exists. The actual control is one Write and
		// first-result preservation, not a fabricated execution counter.
	}
}

func TestCanonicalBeadCompletionConcurrentReentryKeepsFirstAndWritesOnce(t *testing.T) {
	out := &beadCompletionTestWriter{entered: make(chan struct{}), release: make(chan struct{})}
	completion := beadCompletionTestPrepare(t, clioutput.Options{Formats: []string{"json"}}, out)
	original := &struct{ Revision string }{Revision: "first R"}
	opErr, closeErr := errors.New("first failure"), errors.New("first Close failure")
	firstDone := make(chan canonicalBeadMutationProjection, 1)
	go func() { firstDone <- completion.Complete(original, "first human", opErr, closeErr) }()
	// Block the real first Write while followers contend with Complete.
	<-out.entered
	const followers = 8
	results := make(chan canonicalBeadMutationProjection, followers)
	var ready sync.WaitGroup
	ready.Add(followers)
	for i := 0; i < followers; i++ {
		go func() {
			ready.Done()
			results <- completion.Complete("replacement", "replacement", graphstore.ErrOutcomeUnknown, nil)
		}()
	}
	ready.Wait()
	close(out.release)
	first := <-firstDone
	if first.OutputErr != nil || first.Value != original || first.OperationErr != opErr || first.CloseErr != closeErr {
		t.Fatalf("first completion changed: %#v", first)
	}
	for i := 0; i < followers; i++ {
		result := <-results
		if result.Value != original || result.Human != "first human" || result.OperationErr != opErr || result.CloseErr != closeErr ||
			result.Disposition != canonicalMutationOutputFailed || !errors.Is(result.OutputErr, errCanonicalMutationOutputConsumed) ||
			!errors.Is(result.Err(), opErr) || !errors.Is(result.Err(), closeErr) {
			t.Fatalf("concurrent reentry replaced channels: %#v", result)
		}
	}
	if calls, data := out.snapshot(); calls != 1 {
		t.Fatalf("concurrent completion wrote %d times: %s", calls, data)
	}
}

func TestCanonicalBeadCompletionCopyDetachesPayloadRepresentation(t *testing.T) {
	// Representation-only copy, never supplied to Complete or internal mapper
	// state: authenticates no producer and does not fabricate checked capture.
	payload := json.RawMessage("{\"revision\":\"R <>&\",\"n\":1.00}")
	value := &struct{ ID string }{ID: "original"}
	opErr, closeErr := errors.New("operation"), errors.New("Close")
	original := canonicalBeadMutationProjection{
		canonicalMutationOutputResult: canonicalMutationOutputResult{
			Value: value, Human: "original", OperationErr: opErr, CloseErr: closeErr, Disposition: canonicalMutationOutputFailed,
		},
		Payload: payload,
	}
	copied := canonicalBeadMutationCompletionCopy(original)
	if !bytes.Equal(copied.Payload, payload) || copied.Value != value || copied.OperationErr != opErr || copied.CloseErr != closeErr {
		t.Fatal("copy changed representation or original channels")
	}
	copied.Payload[0] = '['
	if payload[0] != '{' {
		t.Fatal("copy aliases original payload")
	}
	original.Payload[1] = '!'
	if copied.Payload[1] == '!' {
		t.Fatal("original aliases copied payload")
	}
	if copied := canonicalBeadMutationCompletionCopy(canonicalBeadMutationProjection{}); copied.Payload != nil {
		t.Fatal("absent payload manufactured")
	}
}
