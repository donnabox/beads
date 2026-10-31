package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/clioutput"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/resourceget"
)

type mutationOutputTestWriter struct {
	calls int
	data  []byte
	err   error
	short bool
}

func (w *mutationOutputTestWriter) Write(p []byte) (int, error) {
	w.calls++
	w.data = bytes.Clone(p)
	if w.err != nil {
		return 0, w.err
	}
	if w.short {
		return len(p) - 1, nil
	}
	return len(p), nil
}

func mutationOutputTestAdapter(t *testing.T, shape canonicalMutationOutputShape, out io.Writer) *canonicalMutationOutput {
	t.Helper()
	required, ok := canonicalMutationOutputMembers(shape)
	if !ok {
		t.Fatal("test requested an unsupported shape")
	}
	// This pin labels a private fixture native payload, not public Write.
	adapter, err := newCanonicalMutationOutput(shape, clioutput.MutationDescriptor{
		BDPSpec: "urn:bd-test:captured-native-output:1", Required: required,
	}, clioutput.Options{Formats: []string{"json"}}, out)
	if err != nil {
		t.Fatal(err)
	}
	return adapter
}

// Constructor-created documents below are REPRESENTATION FIXTURES ONLY.
// They do not populate graphstore's private checked14 capture and never prove
// transaction, materialization, installed-Type authority or commit.
func mutationOutputTestDocument(t *testing.T, kind graph.ResourceKind, id, revision string) resourceget.ResourceDocument {
	t.Helper()
	fields := map[string]any{
		"id": id, "revision": revision,
		"type":       map[string]any{"uri": "https://types.example.test/Task", "revision": "type <>& /@= exact"},
		"properties": map[string]any{"title": "original", "body": "R <>&"},
	}
	if kind == graph.KindLink {
		fields["source"] = map[string]any{"uri": "https://example.test/work/a", "revision": "from <>& /@= opaque"}
		fields["target"] = map[string]any{"uri": "https://remote.example.test/b", "revision": "to <>& /@= opaque"}
	}
	raw, err := canonicalMutationMarshalNative(fields)
	if err != nil {
		t.Fatal(err)
	}
	document, err := resourceget.NewResourceDocument("https://example.test/work/", kind, raw)
	if err != nil {
		t.Fatal(err)
	}
	return document
}

func TestCanonicalMutationOutputPreparesBeforeOperation(t *testing.T) {
	out := &mutationOutputTestWriter{}
	cases := []struct {
		shape      canonicalMutationOutputShape
		descriptor clioutput.MutationDescriptor
		options    clioutput.Options
	}{
		{"unsupported", clioutput.MutationDescriptor{BDPSpec: "fixture", Required: []string{"id"}}, clioutput.Options{}},
		{canonicalMutationOutputBead, clioutput.MutationDescriptor{Required: []string{"id", "type", "revision"}}, clioutput.Options{}},
		{canonicalMutationOutputBead, clioutput.MutationDescriptor{BDPSpec: "fixture", Required: []string{"id"}}, clioutput.Options{}},
		{canonicalMutationOutputBead, clioutput.MutationDescriptor{BDPSpec: "fixture", Required: []string{"id", "type", "revision"}}, clioutput.Options{Include: []string{"/properties"}}},
	}
	for _, test := range cases {
		adapter, err := newCanonicalMutationOutput(test.shape, test.descriptor, test.options, out)
		if err == nil || adapter != nil {
			t.Fatalf("invalid preparation admitted: %#v", test)
		}
	}
	if out.calls != 0 {
		t.Fatal("preparation produced output")
	}
}

func TestCanonicalMutationOutputCapturedDocumentExactFacts(t *testing.T) {
	id := "https://example.test/work/link"
	document := mutationOutputTestDocument(t, graph.KindLink, id, "R")
	raw, err := canonicalMutationCapturedDocument(graph.KindLink, id, "R", document.Type().URI, document, true)
	if err != nil || !bytes.Equal(raw, document.NativeJSON()) {
		t.Fatalf("captured representation changed: %s, %v", raw, err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	if _, present := fields["metadata"]; present {
		t.Fatal("absent metadata was manufactured")
	}
	var typ resourceget.TypeReference
	if err := json.Unmarshal(fields["type"], &typ); err != nil || typ.Revision != "type <>& /@= exact" {
		t.Fatalf("exact affiliation lost: %#v %v", typ, err)
	}
	for name, expected := range map[string]struct{ uri, revision string }{
		"source": {"https://example.test/work/a", "from <>& /@= opaque"},
		"target": {"https://remote.example.test/b", "to <>& /@= opaque"},
	} {
		var endpoint struct {
			URI      string `json:"uri"`
			Revision string `json:"revision"`
		}
		if err := json.Unmarshal(fields[name], &endpoint); err != nil || (endpoint.Revision != expected.revision || endpoint.URI != expected.uri) {
			t.Fatalf("%s whole Reference changed: %#v %v", name, endpoint, err)
		}
	}
	// Exercise the SAME native wrapper encoder plus existing renderer. This
	// remains a representation fixture, not a producer-capture admission.
	sourceDocument := mutationOutputTestDocument(t, graph.KindBead, "https://example.test/work/source", "source <>& R")
	payload, err := canonicalMutationMarshalNative(map[string]any{"link": json.RawMessage(raw), "source": json.RawMessage(sourceDocument.NativeJSON()), "changed": false})
	if err != nil {
		t.Fatal(err)
	}
	renderer, err := clioutput.NewMutation(clioutput.MutationDescriptor{BDPSpec: "urn:bd-test:native-fixture:1", Required: []string{"link", "source", "changed"}}, clioutput.Options{Formats: []string{"json"}})
	if err != nil {
		t.Fatal(err)
	}
	rendered, err := renderer.RenderSuccess(payload)
	if err != nil {
		t.Fatal(err)
	}
	var envelope struct {
		Data struct {
			Link json.RawMessage `json:"link"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rendered, &envelope); err != nil || !bytes.Equal(envelope.Data.Link, document.NativeJSON()) {
		t.Fatalf("native wrapper/renderer changed full captured Resource bytes: %s %v", rendered, err)
	}
	// Independent raw numeric/HTML sentinel makes this control fail if the
	// wrapper encoder is changed back to default HTML escaping.
	literal := json.RawMessage("{\"literal\":\"<>&\",\"number\":1.00}")
	encoded, err := canonicalMutationMarshalNative(map[string]any{"value": literal})
	if err != nil || !bytes.Contains(encoded, literal) {
		t.Fatalf("native RawMessage spelling changed: %s %v", encoded, err)
	}
	// NativeJSON is a detached byte carrier.
	raw[0] = '['
	if document.NativeJSON()[0] != '{' {
		t.Fatal("returned bytes alias document storage")
	}
}

func TestCanonicalMutationOutputRefusesWrongOrAbsentCapturedFacts(t *testing.T) {
	document := mutationOutputTestDocument(t, graph.KindBead, "https://example.test/work/a", "R")
	cases := []struct {
		kind                  graph.ResourceKind
		id, revision, typeURI string
		document              resourceget.ResourceDocument
		present               bool
	}{
		{graph.KindBead, document.ID(), "R", document.Type().URI, document, false},
		{graph.KindLink, document.ID(), "R", document.Type().URI, document, true},
		{graph.KindBead, "https://example.test/work/other", "R", document.Type().URI, document, true},
		{graph.KindBead, document.ID(), "R+1", document.Type().URI, document, true},
		{graph.KindBead, document.ID(), "R", "https://types.example.test/Other", document, true},
		{graph.KindBead, document.ID(), "R", document.Type().URI, resourceget.ResourceDocument{}, true},
	}
	for _, test := range cases {
		if _, err := canonicalMutationCapturedDocument(test.kind, test.id, test.revision, test.typeURI, test.document, test.present); !errors.Is(err, errCanonicalMutationOutputCapture) {
			t.Fatalf("mismatched fixture admitted: %#v, %v", test, err)
		}
	}
	// A constructor-built document and a caller-built DTO cannot use the
	// production type admission path, however valid their JSON shape.
	for _, value := range []any{document, document.NativeJSON(), map[string]any{"id": document.ID()}, struct{ Resource resourceget.ResourceDocument }{document}} {
		if _, err := canonicalMutationOutputPayload(canonicalMutationOutputBead, value); !errors.Is(err, errCanonicalMutationOutputCapture) {
			t.Fatalf("untrusted result %T admitted: %v", value, err)
		}
	}
	legacy := graphstore.Record{ID: document.ID(), Revision: "R", Type: document.Type().URI}
	if _, err := canonicalMutationCapturedResource(legacy); !errors.Is(err, errCanonicalMutationOutputCapture) {
		t.Fatalf("legacy record without private producer facet admitted: %v", err)
	}
}

func TestCanonicalMutationOutputKnownReturnSurvivesCloseAndOutputErrors(t *testing.T) {
	closeErr := errors.New("checked store Close IO failure")
	writeErr := errors.New("stdout IO failure")
	out := &mutationOutputTestWriter{err: writeErr}
	adapter := mutationOutputTestAdapter(t, canonicalMutationOutputBead, out)
	original := graphstore.Record{ID: "https://example.test/work/a", Type: "https://types.example.test/Task", Revision: "R"}
	current := original
	// Hypothetical later state is intentionally not made available through any
	// callback/store. This adapter has no readback or execution seam.
	current.Revision = "R+1"
	result := adapter.Complete(original, "original human R", nil, closeErr)
	if result.Disposition != canonicalMutationOutputReturned || result.OperationErr != nil {
		t.Fatalf("cleanup/output failure changed successful operation return: %#v", result)
	}
	got, ok := result.Value.(graphstore.Record)
	if !ok || got.Revision != "R" || current.Revision != "R+1" || result.Human != "original human R" {
		t.Fatalf("original result erased or reread: %#v", result)
	}
	if !errors.Is(result.Err(), closeErr) || !errors.Is(result.Err(), writeErr) || !errors.Is(result.PayloadErr, errCanonicalMutationOutputCapture) {
		t.Fatalf("independent causes lost: %#v / %v", result, result.Err())
	}
	// This unfaceted fixture intentionally produces a named output gap,
	// not a fabricated successful Resource. Actual checked14 producer/output
	// integration remains a root-owned engine qualification.
	if out.calls != 1 || !bytes.Contains(out.data, []byte("output_unavailable")) {
		t.Fatalf("missing bounded output-gap document: %d %s", out.calls, out.data)
	}
}

func TestCanonicalMutationOutputKeepsOperationAndCloseErrorsIndependent(t *testing.T) {
	opErr, closeErr := errors.New("validation failed"), errors.New("Close failed")
	out := &mutationOutputTestWriter{}
	adapter := mutationOutputTestAdapter(t, canonicalMutationOutputIssue, out)
	value := graphstore.IssueMutationResult{Changed: false}
	result := adapter.Complete(value, "original", opErr, closeErr)
	if result.Disposition != canonicalMutationOutputFailed || result.OperationErr != opErr || result.CloseErr != closeErr ||
		!errors.Is(result.Err(), opErr) || !errors.Is(result.Err(), closeErr) {
		t.Fatalf("error provenance collapsed: %#v", result)
	}
	if _, ok := result.Value.(graphstore.IssueMutationResult); !ok {
		t.Fatal("available value discarded")
	}
	// A valid caller-built representation is still NOT proof of a native
	// commit. A known operation error never projects it into applied data.
	document := mutationOutputTestDocument(t, graph.KindBead, "https://example.test/work/a", "R")
	out2 := &mutationOutputTestWriter{}
	adapter2 := mutationOutputTestAdapter(t, canonicalMutationOutputBead, out2)
	failed := adapter2.Complete(document, "unconfirmed", opErr, closeErr)
	var envelope struct {
		Outcome string          `json:"outcome"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(out2.data, &envelope); err != nil || envelope.Outcome != "failure" ||
		(len(envelope.Data) != 0 && string(envelope.Data) != "null") || failed.PayloadErr != nil {
		t.Fatalf("known error promoted an available representation: %s %v", out2.data, err)
	}
	if kept, ok := failed.Value.(resourceget.ResourceDocument); !ok || !bytes.Equal(kept.NativeJSON(), document.NativeJSON()) {
		t.Fatal("original unconfirmed representation was erased")
	}
	// Root must still exercise an actual checked14 native failure. This pure
	// constructor fixture intentionally claims no native producer authority.
}

func TestCanonicalMutationOutputUnknownIsNeverSuccessFromValue(t *testing.T) {
	closeErr := errors.New("Close failed")
	opErr := errors.Join(errors.New("native commit response lost"), graphstore.ErrOutcomeUnknown)
	out := &mutationOutputTestWriter{}
	adapter := mutationOutputTestAdapter(t, canonicalMutationOutputBead, out)
	value := map[string]any{"revision": "unconfirmed"}
	result := adapter.Complete(value, "unconfirmed original", opErr, closeErr)
	if result.Disposition != canonicalMutationOutputUnknown || !errors.Is(result.Err(), graphstore.ErrOutcomeUnknown) ||
		!errors.Is(result.Err(), closeErr) || result.PayloadErr != nil {
		t.Fatalf("unknown was projected/promoted or causes lost: %#v", result)
	}
	var envelope struct {
		Outcome string          `json:"outcome"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(out.data, &envelope); err != nil || envelope.Outcome != "unknown" {
		t.Fatalf("unknown envelope changed: %s %v", out.data, err)
	}
	if len(envelope.Data) != 0 && string(envelope.Data) != "null" {
		t.Fatalf("unconfirmed value promoted into data: %s", envelope.Data)
	}
}

func TestCanonicalMutationOutputNoopAndUnsupportedEvidenceRemainOriginal(t *testing.T) {
	out := &mutationOutputTestWriter{}
	adapter := mutationOutputTestAdapter(t, canonicalMutationOutputMemory, out)
	value := graphstore.MemoryMutationResult{
		Memory:  graphstore.Record{ID: "https://example.test/work/a", Revision: "same", Type: "https://types.example.test/Task"},
		Changed: false,
	}
	result := adapter.Complete(value, "unchanged", nil, nil)
	got, ok := result.Value.(graphstore.MemoryMutationResult)
	if !ok || got.Changed || got.Memory.Revision != "same" || got.Replaced != nil ||
		result.Disposition != canonicalMutationOutputReturned || !errors.Is(result.PayloadErr, errCanonicalMutationOutputCapture) {
		t.Fatalf("no-op facts changed or missing capture manufactured: %#v", result)
	}
	for _, value := range []any{graphstore.LinkDeleteResult{}, graphstore.DependencyRemoveResult{}, graphstore.LinkMutationResult{Source: nil}} {
		if _, err := canonicalMutationOutputPayload(canonicalMutationOutputLinkChange, value); !errors.Is(err, errCanonicalMutationOutputCapture) {
			t.Fatalf("unsupported/absent affected-source evidence admitted: %T %v", value, err)
		}
	}
}

func TestCanonicalMutationOutputRendererRetainsFixtureNoopAndExactNativeBody(t *testing.T) {
	// This exercises pure representation/renderer correspondence, not a native
	// producer or the closed production completion admission.
	document := mutationOutputTestDocument(t, graph.KindBead, "https://example.test/work/a", "same")
	native, err := canonicalMutationCapturedDocument(graph.KindBead, document.ID(), document.Revision(), document.Type().URI, document, true)
	if err != nil {
		t.Fatal(err)
	}
	payload, err := canonicalMutationMarshalNative(map[string]any{"memory": native, "changed": false})
	if err != nil {
		t.Fatal(err)
	}
	renderer, err := clioutput.NewMutation(clioutput.MutationDescriptor{BDPSpec: "urn:bd-test:native-fixture:1", Required: []string{"memory", "changed"}}, clioutput.Options{Formats: []string{"json"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, cleanupFailure := range []bool{false, true} {
		var rendered []byte
		if cleanupFailure {
			rendered, err = renderer.RenderFailureWithResult("cli", canonicalMutationOutputProblem("cleanup_error", "cleanup failed"), payload)
		} else {
			rendered, err = renderer.RenderSuccess(payload)
		}
		if err != nil {
			t.Fatal(err)
		}
		var envelope struct {
			Data struct {
				Changed bool            `json:"changed"`
				Memory  json.RawMessage `json:"memory"`
			} `json:"data"`
		}
		if err := json.Unmarshal(rendered, &envelope); err != nil || envelope.Data.Changed ||
			!bytes.Equal(envelope.Data.Memory, document.NativeJSON()) {
			t.Fatalf("pure renderer changed original no-op Resource: %s %v", rendered, err)
		}
	}
}

func TestCanonicalMutationOutputFailureWritesOnceWithoutResubmission(t *testing.T) {
	for _, short := range []bool{false, true} {
		writeErr := errors.New("broken stdout")
		out := &mutationOutputTestWriter{short: short}
		if !short {
			out.err = writeErr
		}
		adapter := mutationOutputTestAdapter(t, canonicalMutationOutputBead, out)
		// The helper has no execution callback. Exercise a second delivery
		// attempt and prove that neither a new output nor value replacement occurs.
		first := adapter.Complete("original opaque value", "original", graphstore.ErrOutcomeUnknown, nil)
		if short && !errors.Is(first.OutputErr, io.ErrShortWrite) || !short && !errors.Is(first.OutputErr, writeErr) {
			t.Fatalf("IO error not retained: %v", first.Err())
		}
		second := adapter.Complete("replacement must be ignored", "replacement", nil, nil)
		if out.calls != 1 || !errors.Is(second.OutputErr, errCanonicalMutationOutputConsumed) ||
			second.Value != first.Value || second.Disposition != first.Disposition ||
			!strings.Contains(string(out.data), "unknown") {
			t.Fatalf("delivery retried or completion replaced: calls=%d %#v", out.calls, second)
		}
	}
}

// This writer deliberately calls the SAME adapter synchronously. It represents
// external output code, not a producer/executor callback or checked-facet fake.
type mutationOutputReentrantTestWriter struct {
	adapter *canonicalMutationOutput
	calls   int
	nested  canonicalMutationOutputResult
	err     error
	short   bool
}

func (w *mutationOutputReentrantTestWriter) Write(p []byte) (int, error) {
	w.calls++
	w.nested = w.adapter.Complete("replacement", "replacement", graphstore.ErrOutcomeUnknown, nil)
	if w.err != nil {
		return 0, w.err
	}
	if w.short {
		return len(p) - 1, nil
	}
	return len(p), nil
}

func TestCanonicalMutationOutputSynchronousWriterReentryRefusesWithoutDeadlock(t *testing.T) {
	for _, mode := range []string{"success", "error", "short"} {
		t.Run(mode, func(t *testing.T) {
			writeErr := errors.New("writer failed after reentry")
			opErr, closeErr := errors.New("original operation failure"), errors.New("original Close failure")
			out := &mutationOutputReentrantTestWriter{short: mode == "short"}
			if mode == "error" {
				out.err = writeErr
			}
			out.adapter = mutationOutputTestAdapter(t, canonicalMutationOutputBead, out)
			original := &struct{ Revision string }{Revision: "original R"}
			done := make(chan canonicalMutationOutputResult, 1)
			go func() { done <- out.adapter.Complete(original, "first human", opErr, closeErr) }()
			deadline := time.NewTimer(3 * time.Second)
			defer deadline.Stop()
			var first canonicalMutationOutputResult
			select {
			case first = <-done:
			case <-deadline.C:
				t.Fatal("synchronous writer reentry did not return within 3 seconds")
			}
			// The nested call happens BEFORE Write returns. It can preserve
			// available first channels but cannot invent the future Write error.
			nested := out.nested
			if nested.Value != original || nested.Human != "first human" || nested.OperationErr != opErr || nested.CloseErr != closeErr ||
				nested.Disposition != canonicalMutationOutputFailed || nested.PayloadErr != nil || nested.RenderErr != nil ||
				!errors.Is(nested.OutputErr, errCanonicalMutationOutputConsumed) || errors.Is(nested.OutputErr, writeErr) ||
				errors.Is(nested.OutputErr, io.ErrShortWrite) || !errors.Is(nested.Err(), opErr) || !errors.Is(nested.Err(), closeErr) {
				t.Fatalf("reentry lost first channels or invented unsettled output status: %#v", nested)
			}
			var expected error
			if mode == "error" {
				expected = writeErr
			} else if mode == "short" {
				expected = io.ErrShortWrite
			}
			if first.Value != original || first.Human != "first human" || first.OperationErr != opErr || first.CloseErr != closeErr ||
				first.Disposition != canonicalMutationOutputFailed || first.PayloadErr != nil || first.RenderErr != nil ||
				(expected == nil && first.OutputErr != nil) || (expected != nil && !errors.Is(first.OutputErr, expected)) ||
				errors.Is(first.OutputErr, errCanonicalMutationOutputConsumed) {
				t.Fatalf("settled original completion changed: %#v", first)
			}
			settled := out.adapter.Complete("later replacement", "later", nil, nil)
			if settled.Value != original || settled.Human != "first human" || settled.OperationErr != opErr || settled.CloseErr != closeErr ||
				settled.Disposition != first.Disposition || !errors.Is(settled.OutputErr, errCanonicalMutationOutputConsumed) ||
				(expected != nil && !errors.Is(settled.OutputErr, expected)) || out.calls != 1 {
				t.Fatalf("settled duplicate retried output or lost actual output status: calls=%d %#v", out.calls, settled)
			}
		})
	}
}

type mutationOutputInflightTestWriter struct {
	mu      sync.Mutex
	once    sync.Once
	calls   int
	entered chan struct{}
	release chan struct{}
	err     error
}

func (w *mutationOutputInflightTestWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	w.calls++
	w.mu.Unlock()
	w.once.Do(func() { close(w.entered) })
	<-w.release
	return 0, w.err
}

func (w *mutationOutputInflightTestWriter) count() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.calls
}

func TestCanonicalMutationOutputConcurrentInflightRefusalKeepsAvailableChannels(t *testing.T) {
	writeErr := errors.New("eventual output failure")
	closeErr := errors.Join(errors.New("original Close diagnostic"), graphstore.ErrOutcomeUnknown)
	out := &mutationOutputInflightTestWriter{entered: make(chan struct{}), release: make(chan struct{}), err: writeErr}
	var release sync.Once
	defer release.Do(func() { close(out.release) })
	adapter := mutationOutputTestAdapter(t, canonicalMutationOutputBead, out)
	// This is an unfaceted original value. Its missing-capture gap is known
	// before Write, without pretending a successful native producer fixture.
	original := &struct{ Revision string }{Revision: "original R"}
	firstDone := make(chan canonicalMutationOutputResult, 1)
	go func() { firstDone <- adapter.Complete(original, "first human", nil, closeErr) }()
	deadline := time.NewTimer(3 * time.Second)
	defer deadline.Stop()
	select {
	case <-out.entered:
	case <-deadline.C:
		t.Fatal("original completion did not enter Write within 3 seconds")
	}
	const followers = 8
	results := make(chan canonicalMutationOutputResult, followers)
	for i := 0; i < followers; i++ {
		go func() { results <- adapter.Complete("replacement", "replacement", graphstore.ErrOutcomeUnknown, nil) }()
	}
	for i := 0; i < followers; i++ {
		var duplicate canonicalMutationOutputResult
		select {
		case duplicate = <-results:
		case <-deadline.C:
			t.Fatal("in-flight refusal waited on external writer beyond 3 seconds")
		}
		if duplicate.Value != original || duplicate.Human != "first human" || duplicate.OperationErr != nil || duplicate.CloseErr != closeErr ||
			duplicate.Disposition != canonicalMutationOutputReturned || !errors.Is(duplicate.PayloadErr, errCanonicalMutationOutputCapture) ||
			duplicate.RenderErr != nil || !errors.Is(duplicate.OutputErr, errCanonicalMutationOutputConsumed) ||
			errors.Is(duplicate.OutputErr, writeErr) || !errors.Is(duplicate.Err(), closeErr) {
			t.Fatalf("in-flight first channels lost or future IO failure invented: %#v", duplicate)
		}
	}
	// All duplicate calls MUST return while the first external Write is still
	// blocked. Only after that check does the writer's actual error exist.
	release.Do(func() { close(out.release) })
	var first canonicalMutationOutputResult
	select {
	case first = <-firstDone:
	case <-deadline.C:
		t.Fatal("original completion did not settle after writer release")
	}
	if first.Value != original || first.Human != "first human" || first.OperationErr != nil || first.CloseErr != closeErr ||
		first.Disposition != canonicalMutationOutputReturned || !errors.Is(first.PayloadErr, errCanonicalMutationOutputCapture) ||
		!errors.Is(first.OutputErr, writeErr) || errors.Is(first.OutputErr, errCanonicalMutationOutputConsumed) {
		t.Fatalf("actual settled first channels lost: %#v", first)
	}
	settled := adapter.Complete(nil, "late replacement", nil, nil)
	if settled.Value != original || settled.Human != "first human" || settled.OperationErr != nil || settled.CloseErr != closeErr ||
		settled.Disposition != canonicalMutationOutputReturned || !errors.Is(settled.PayloadErr, errCanonicalMutationOutputCapture) ||
		!errors.Is(settled.OutputErr, writeErr) || !errors.Is(settled.OutputErr, errCanonicalMutationOutputConsumed) || out.count() != 1 {
		t.Fatalf("settled refusal lost actual status or wrote again: calls=%d %#v", out.count(), settled)
	}
	// No operation callback, readback or executor counter is introduced.
}
