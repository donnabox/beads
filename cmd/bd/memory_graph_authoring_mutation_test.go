package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/clioutput"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

type memoryAuthoringStore struct {
	events                         []string
	types                          []graph.TypeDescriptor
	owned                          []string
	result                         graphstore.LinkMutationResult
	contextErr, writeErr, closeErr error
	seen                           graphstore.LinkCreateRequest
}

func (s *memoryAuthoringStore) ListInstalledTypes(context.Context) ([]graph.TypeDescriptor, error) {
	s.events = append(s.events, "types")
	return s.types, s.contextErr
}
func (s *memoryAuthoringStore) SelectedOwnedLinkTypes(string) ([]string, error) {
	s.events = append(s.events, "ownership")
	return s.owned, nil
}
func (s *memoryAuthoringStore) AddInformationalLink(_ context.Context, r graphstore.LinkCreateRequest) (graphstore.LinkMutationResult, error) {
	s.events = append(s.events, "write")
	s.seen = r
	return s.result, s.writeErr
}
func (s *memoryAuthoringStore) Close() error { s.events = append(s.events, "close"); return s.closeErr }

func memoryAuthoringFixture(t *testing.T) (*memoryAuthoringStore, graphstore.LinkCreateRequest) {
	t.Helper()
	const scope = "https://example.invalid/writer/"
	descriptors := []graph.TypeDescriptor{}
	for _, spec := range []graph.TypeDescriptorSpec{
		{ID: scope + "memory", Name: "Memory", Describes: graph.KindBead},
		{ID: scope + "references", Name: "References", Describes: graph.KindLink, Source: &graph.EndpointConstraint{}, Target: &graph.EndpointConstraint{}},
	} {
		d, err := graph.NewTypeDescriptor(spec)
		if err != nil {
			t.Fatal(err)
		}
		descriptors = append(descriptors, d)
	}
	attribution := graphstore.Attribution{Actor: "author", Status: "claimed", RecordedAt: "2026-10-21T12:00:00Z"}
	link := graphstore.LinkRecord{ID: scope + "links/created", Type: scope + "references", Revision: "link-r2", Version: "link-r2", Source: scope + "beads/a", Target: scope + "beads/b", Properties: map[string]any{"title": "Context", "n": json.Number("9007199254740994")}, Metadata: json.RawMessage(`{"arbitrary":{"keep":true}}`), Attribution: attribution}
	owned, err := json.Marshal(link)
	if err != nil {
		t.Fatal(err)
	}
	source := graphstore.Record{ID: link.Source, Type: scope + "memory", Revision: "source-r2", Version: "source-r2", Properties: graphstore.Properties{Title: "Memory", Body: "Primary content"}, Metadata: json.RawMessage(`{}`), Owned: []json.RawMessage{owned}, Attribution: attribution}
	return &memoryAuthoringStore{types: descriptors, owned: []string{link.Type}, result: graphstore.LinkMutationResult{Link: link, Source: source, Changed: true}}, graphstore.LinkCreateRequest{SourcePath: "beads/a", TargetPath: "beads/b", UnconditionalSource: true, Actor: "author"}
}
func memoryAuthoringRenderer(t *testing.T, options clioutput.Options) *clioutput.MutationRenderer {
	t.Helper()
	r, err := newMemoryMutationRenderer(options)
	if err != nil {
		t.Fatal(err)
	}
	return r
}
func memoryAuthoringPublish(t *testing.T, state memoryLinkExecution, out io.Writer) (string, string, error) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	if out == nil {
		out = &stdout
	}
	cmd := &cobra.Command{}
	cmd.SetOut(out)
	cmd.SetErr(&stderr)
	options := clioutput.Options{JSON: true}
	err := publishMemoryLinkExecution(cmd, memoryAuthoringRenderer(t, options), options, state)
	return stdout.String(), stderr.String(), err
}
func memoryAuthoringExit(t *testing.T, err error, want int) {
	t.Helper()
	var exit *exitError
	if !errors.As(err, &exit) || exit.Code != want {
		t.Fatalf("exit=%v want %d", err, want)
	}
}

func TestMemoryLinkCreateLifecycleAndProjection(t *testing.T) {
	s, req := memoryAuthoringFixture(t)
	state := executeMemoryLinkCreate(t.Context(), s, req, encodeMemoryGraphJSON)
	if state.err != nil || !state.applied || !reflect.DeepEqual(s.events, []string{"types", "ownership", "write", "close"}) || !reflect.DeepEqual(s.seen, req) {
		t.Fatalf("lifecycle: %+v events=%v", state, s.events)
	}
	out, stderr, err := memoryAuthoringPublish(t, state, nil)
	if err != nil || stderr != "" {
		t.Fatalf("publish: %s %s %v", out, stderr, err)
	}
	var frame struct {
		Kind    string          `json:"result_kind"`
		Outcome string          `json:"outcome"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal([]byte(out), &frame); err != nil || frame.Kind != "mutation" || frame.Outcome != "success" {
		t.Fatalf("frame %s %v", out, err)
	}
	for _, want := range []string{`"revision":"link-r2"`, `"revision":"source-r2"`, `"changed":true`, `9007199254740994`, `"principal":"author"`, `"basis":"writer-supplied"`, `"ownedLinks"`, `"arbitrary":{"keep":true}`} {
		if !strings.Contains(string(frame.Data), want) {
			t.Fatalf("lost %s: %s", want, frame.Data)
		}
	}
	for _, private := range []string{`"version":`, `"actor":`, `"recordedAt":`, `"owned":`} {
		if strings.Contains(string(frame.Data), private) {
			t.Fatalf("private DTO field %s: %s", private, frame.Data)
		}
	}
	// A real writer-provided predecessor is surfaced; it is not inferred from the guard.
	s, req = memoryAuthoringFixture(t)
	s.result.ReplacedSource = &graphstore.ReplacedMemory{ID: s.result.Link.Source, Version: "prior", Attribution: graphstore.Attribution{Actor: "previous-author", Status: "claimed", RecordedAt: "2026-10-21T11:00:00Z"}}
	state = executeMemoryLinkCreate(t.Context(), s, req, encodeMemoryGraphJSON)
	if state.err != nil || !bytes.Contains(state.raw, []byte(`"replacedSource"`)) || !bytes.Contains(state.raw, []byte(`"principal":"previous-author"`)) {
		t.Fatalf("predecessor: %s %v", state.raw, state.err)
	}
}

func TestMemoryLinkCreateFailuresRetainActualOutcome(t *testing.T) {
	fault := errors.New("injected close failure")
	for _, which := range []string{"context", "empty-context", "write", "close", "projection", "encoding", "unknown", "unknown-close"} {
		t.Run(which, func(t *testing.T) {
			s, req := memoryAuthoringFixture(t)
			encode := encodeMemoryGraphJSON
			wantExit := 5
			wantApplied := false
			switch which {
			case "context":
				s.contextErr = graphstore.ErrInvalidStore
			case "empty-context":
				s.types = nil
			case "write":
				s.writeErr = graphstore.ErrConflict
				wantExit = 4
			case "close":
				s.closeErr = fault
				wantApplied = true
			case "projection":
				s.result.Source = nil
				wantApplied = true
				wantExit = 1
			case "encoding":
				encode = func(any) ([]byte, error) { return nil, errors.New("encode failed") }
				wantApplied = true
				wantExit = 1
			case "unknown":
				s.result = graphstore.LinkMutationResult{}
				s.writeErr = errors.Join(graphstore.ErrOutcomeUnknown, graphstore.ErrConflict)
				wantExit = 6
			case "unknown-close":
				s.closeErr = errors.Join(graphstore.ErrOutcomeUnknown, fault)
				wantApplied = true
				wantExit = 6
			}
			state := executeMemoryLinkCreate(t.Context(), s, req, encode)
			if state.err == nil || state.applied != wantApplied {
				t.Fatalf("state %+v", state)
			}
			writes, closes := 0, 0
			for _, event := range s.events {
				if event == "write" {
					writes++
				}
				if event == "close" {
					closes++
				}
			}
			if closes != 1 || writes > 1 || (strings.Contains(which, "context") && writes != 0) {
				t.Fatalf("retry/close: %v", s.events)
			}
			out, stderr, err := memoryAuthoringPublish(t, state, nil)
			memoryAuthoringExit(t, err, wantExit)
			if !json.Valid([]byte(out)) {
				t.Fatalf("invalid frame %s; %s", out, stderr)
			}
			var frame map[string]json.RawMessage
			if err := json.Unmarshal([]byte(out), &frame); err != nil {
				t.Fatal(err)
			}
			if strings.HasPrefix(which, "unknown") {
				if string(frame["outcome"]) != `"unknown"` || frame["data"] != nil || strings.Contains(out, `"applied":true`) || strings.Contains(stderr, "operation completed") {
					t.Fatalf("unknown overstated %s %s", out, stderr)
				}
				if !strings.Contains(stderr, "inspect") {
					t.Fatal(stderr)
				}
				if which == "unknown" && strings.Contains(out, `"changed":false`) {
					t.Fatalf("fabricated zero changed: %s", out)
				}
			} else if wantApplied {
				if !strings.Contains(out, `"applied":true`) || !strings.Contains(stderr, "inspect before retry") {
					t.Fatalf("lost known effect: %s %s", out, stderr)
				}
				if which == "close" && frame["data"] == nil {
					t.Fatalf("lost actual result on close: %s", out)
				}
				if which != "close" && frame["data"] != nil {
					t.Fatalf("invented result: %s", out)
				}
			} else if frame["data"] != nil || strings.Contains(out, `"effect"`) {
				t.Fatalf("refusal looks applied: %s", out)
			}
		})
	}
}

type memoryAuthoringShortWriter struct{ calls int }

func (w *memoryAuthoringShortWriter) Write(p []byte) (int, error) {
	w.calls++
	return len(p) / 2, io.ErrShortWrite
}
func TestMemoryLinkCreateOutputFailureNeverReplays(t *testing.T) {
	for _, unknown := range []bool{false, true} {
		s, req := memoryAuthoringFixture(t)
		if unknown {
			s.closeErr = graphstore.ErrOutcomeUnknown
		}
		state := executeMemoryLinkCreate(t.Context(), s, req, encodeMemoryGraphJSON)
		writer := &memoryAuthoringShortWriter{}
		_, stderr, err := memoryAuthoringPublish(t, state, writer)
		want := 1
		if unknown {
			want = 6
		}
		memoryAuthoringExit(t, err, want)
		if writer.calls != 1 || !strings.Contains(stderr, "inspect") || strings.Count(stderr, "outcome_unknown:") > 1 || strings.Count(stderr, "operation completed") > 1 {
			t.Fatalf("writes=%d stderr=%s", writer.calls, stderr)
		}
		if unknown && strings.Contains(stderr, "operation completed") {
			t.Fatalf("unknown claimed completed: %s", stderr)
		}
	}
}

func TestMemoryLinkCreateInvalidOutputKeepsJSONFailure(t *testing.T) {
	for _, options := range []clioutput.Options{
		{JSON: true, Formats: []string{"tree"}},
		{Formats: []string{"tree", "json"}},
		{JSON: true, Include: []string{"/link/id"}},
	} {
		if _, err := newMemoryMutationRenderer(options); err == nil {
			t.Fatal("accepted unsupported mutation selector", options)
		}
		fallback, err := newMemoryMutationFailureRenderer(options)
		if err != nil {
			t.Fatal(err)
		}
		raw, err := fallback.RenderFailure("cli", json.RawMessage(`{"code":"invalid_selector","message":"refused before write"}`))
		if err != nil || !json.Valid(raw) || !bytes.Contains(raw, []byte(`"outcome":"failure"`)) || bytes.Contains(raw, []byte(`"data"`)) {
			t.Fatalf("invalid error framing %s %v", raw, err)
		}
	}
	for _, options := range []clioutput.Options{{Formats: []string{"tree"}}, {JSON: true, Formats: []string{"tree"}, Omit: []string{"envelope"}}} {
		fallback, err := newMemoryMutationFailureRenderer(options)
		if err != nil {
			t.Fatal(err)
		}
		raw, err := fallback.RenderFailure("cli", json.RawMessage(`{"code":"invalid_selector","message":"refused"}`))
		if err != nil || len(raw) != 0 {
			t.Fatalf("human/raw failure stdout %s %v", raw, err)
		}
	}
}

func TestMemoryLinkCreateRefusalDeliveryKeepsDiagnosis(t *testing.T) {
	s, req := memoryAuthoringFixture(t)
	s.writeErr = graphstore.ErrConflict
	state := executeMemoryLinkCreate(t.Context(), s, req, encodeMemoryGraphJSON)
	writer := &memoryAuthoringShortWriter{}
	_, stderr, err := memoryAuthoringPublish(t, state, writer)
	memoryAuthoringExit(t, err, 1) // the document could not be delivered
	if writer.calls != 1 || !strings.Contains(stderr, "revision_conflict:") || !strings.Contains(stderr, "output_error:") || strings.Index(stderr, "revision_conflict:") > strings.Index(stderr, "output_error:") || strings.Contains(stderr, "operation completed") {
		t.Fatalf("lost refusal diagnosis: calls=%d %s", writer.calls, stderr)
	}
}
