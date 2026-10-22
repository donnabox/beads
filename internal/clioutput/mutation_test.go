package clioutput_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/steveyegge/beads/internal/clioutput"
)

func mutationRenderer(t *testing.T, options clioutput.Options) *clioutput.MutationRenderer {
	t.Helper()
	renderer, err := clioutput.NewMutation(clioutput.MutationDescriptor{BDPSpec: "fixture-contract-not-a-release-pin", Required: []string{"changed"}}, options)
	if err != nil {
		t.Fatal(err)
	}
	return renderer
}

func TestMutationRendererPreservesNativeResultAndNoop(t *testing.T) {
	renderer := mutationRenderer(t, clioutput.Options{JSON: true})
	result := json.RawMessage(`{"link":{"id":"l","revision":"r2","source":{"id":"s","revision":"r3"}},"source":{"id":"s","revision":"r3"},"changed":false,"replacedSource":{"revision":"r1"},"large":9007199254740993}`)
	out, err := renderer.RenderSuccess(result)
	if err != nil {
		t.Fatal(err)
	}
	equalJSON(t, out, `{"schema_version":2,"bdp_spec":"fixture-contract-not-a-release-pin","result_kind":"mutation","outcome":"success","data":`+string(result)+`}`)
	if !bytes.Contains(out, []byte(`9007199254740993`)) || !bytes.Contains(out, []byte(`"changed":false`)) {
		t.Fatalf("lost native mutation meaning: %s", out)
	}
	raw := mutationRenderer(t, clioutput.Options{JSON: true, Omit: []string{"envelope"}})
	bare, err := raw.RenderSuccess(result)
	if err != nil {
		t.Fatal(err)
	}
	equalJSON(t, bare, string(result))
}

func TestMutationRendererDistinguishesFailureAndUnknown(t *testing.T) {
	renderer := mutationRenderer(t, clioutput.Options{JSON: true})
	problem := json.RawMessage(`{"code":"outcome_unknown","message":"inspect before retry","detail":{"token":9007199254740993}}`)
	unknown, err := renderer.RenderUnknown("cli", problem)
	if err != nil {
		t.Fatal(err)
	}
	equalJSON(t, unknown, `{"schema_version":2,"bdp_spec":"fixture-contract-not-a-release-pin","result_kind":"mutation","outcome":"unknown","error":{"origin":"cli","problem":`+string(problem)+`}}`)
	if bytes.Contains(unknown, []byte(`"data"`)) || !bytes.Contains(unknown, []byte(`9007199254740993`)) {
		t.Fatalf("unknown write looks like a success or lost the problem: %s", unknown)
	}
	partial := json.RawMessage(`{"changed":true,"affectedSource":{"id":"s","revision":"r3"}}`)
	failure, err := renderer.RenderFailureWithResult("bdp", json.RawMessage(`{"code":"partial_effect"}`), partial)
	if err != nil {
		t.Fatal(err)
	}
	equalJSON(t, failure, `{"schema_version":2,"bdp_spec":"fixture-contract-not-a-release-pin","result_kind":"mutation","outcome":"failure","data":`+string(partial)+`,"error":{"origin":"bdp","problem":{"code":"partial_effect"}}}`)
	for _, options := range []clioutput.Options{{}, {JSON: true, Omit: []string{"envelope"}}} {
		r := mutationRenderer(t, options)
		for _, render := range []func() ([]byte, error){
			func() ([]byte, error) { return r.RenderFailure("cli", problem) },
			func() ([]byte, error) { return r.RenderUnknown("cli", problem) },
		} {
			out, err := render()
			if err != nil || len(out) != 0 {
				t.Fatalf("human/raw problem leaked to stdout: %s, %v", out, err)
			}
		}
	}
}

func TestMutationRendererRejectsUnsafeProjectionAndMalformedPayloads(t *testing.T) {
	for _, options := range []clioutput.Options{
		{JSON: true, Include: []string{"/link/properties"}},
		{JSON: true, Omit: []string{"/changed"}},
		{Omit: []string{"envelope"}},
		{Formats: []string{"json", "detail"}},
	} {
		if _, err := clioutput.NewMutation(clioutput.MutationDescriptor{BDPSpec: "fixture", Required: []string{"changed"}}, options); err == nil {
			t.Fatalf("accepted unsafe output options: %+v", options)
		}
	}
	if _, err := clioutput.NewMutation(clioutput.MutationDescriptor{}, clioutput.Options{}); err == nil {
		t.Fatal("accepted missing contract identifier")
	}
	if _, err := clioutput.NewMutation(clioutput.MutationDescriptor{BDPSpec: "fixture"}, clioutput.Options{}); err == nil {
		t.Fatal("accepted undeclared mutation result")
	}
	renderer := mutationRenderer(t, clioutput.Options{JSON: true})
	for _, value := range []string{"", `null`, `[]`, `"text"`, `{broken`} {
		if out, err := renderer.RenderSuccess(json.RawMessage(value)); err == nil || len(out) != 0 {
			t.Fatalf("accepted malformed result %s as %s, %v", value, out, err)
		}
	}
	if out, err := renderer.RenderSuccess(json.RawMessage(`{"link":{"id":"l"}}`)); err == nil || len(out) != 0 {
		t.Fatalf("accepted result missing declared effect: %s, %v", out, err)
	}
	for _, problem := range []string{"", `null`, `[]`, `{"code":"x"}`, `{"code":"x","message":" "}`} {
		if out, err := renderer.RenderUnknown("cli", json.RawMessage(problem)); err == nil || len(out) != 0 {
			t.Fatalf("accepted malformed problem %s as %s, %v", problem, out, err)
		}
	}
	if out, err := renderer.RenderFailure("invalid", json.RawMessage(`{"code":"x"}`)); err == nil || len(out) != 0 {
		t.Fatalf("accepted problem origin: %s, %v", out, err)
	}
}
