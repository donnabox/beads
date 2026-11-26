package bdpclient_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/bdp"
	"github.com/steveyegge/beads/internal/bdp/bdpclient"
)

// TestCallDecodesOutcome exercises RemoteScope.Call against a MOCK server.
// cleanroom-v023 offers no call at all, so there is no live oracle for the call
// protocol — this validates the client's encode/decode against the rc2 wire
// shapes (core.lumen §call/§Reply/§Outcome) only, not against a real server.
func TestCallDecodesOutcome(t *testing.T) {
	mux := http.NewServeMux()
	// A sync success → 200 with the Reply's Outcome.success.
	mux.HandleFunc("/issueops/closeIssue/-/call", func(w http.ResponseWriter, r *http.Request) {
		var req bdp.Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		if len(req.Arguments) == 0 || req.Arguments[0].Name != "issue" {
			t.Errorf("server saw unexpected arguments: %+v", req.Arguments)
		}
		_, _ = w.Write([]byte(`{"outcome":{"kind":"success","value":{"id":"be-1","closed":true}},"cause":"c1","callId":"k1"}`))
	})
	// A declared domain failure → 409 with the Reply's Outcome.failure (NOT an
	// out-of-band fault): the client returns the Reply, not an error.
	mux.HandleFunc("/issueops/claimIssue/-/call", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"outcome":{"kind":"failure","error":{"code":"already-claimed","issue":"be-1"}},"cause":"c2","callId":"k2"}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	s := bdpclient.New(srv.URL)

	// Success arm.
	rep, err := s.Call(context.Background(), bdp.ResourceID{ID: "issueops/closeIssue"}, bdp.Request{
		Arguments: []bdp.Argument{{Name: "issue", Value: json.RawMessage(`"be-1"`)}},
	})
	if err != nil {
		t.Fatalf("Call(close): %v", err)
	}
	if rep.Outcome.Kind != bdp.OutcomeSuccess {
		t.Errorf("Outcome.Kind = %q, want success", rep.Outcome.Kind)
	}
	if !strings.Contains(string(rep.Outcome.Value), "closed") {
		t.Errorf("success value = %s", rep.Outcome.Value)
	}
	if rep.CallId != "k1" {
		t.Errorf("CallId = %q, want k1", rep.CallId)
	}

	// Declared-failure arm: 409 is a Reply with Outcome.failure, not an error.
	rep2, err := s.Call(context.Background(), bdp.ResourceID{ID: "issueops/claimIssue"}, bdp.Request{})
	if err != nil {
		t.Fatalf("Call(claim): a 409 declared failure must return a Reply, not an error: %v", err)
	}
	if rep2.Outcome.Kind != bdp.OutcomeFailure {
		t.Errorf("Outcome.Kind = %q, want failure", rep2.Outcome.Kind)
	}
	if !strings.Contains(string(rep2.Outcome.Error), "already-claimed") {
		t.Errorf("failure error = %s", rep2.Outcome.Error)
	}
}
