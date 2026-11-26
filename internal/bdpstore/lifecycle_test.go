package bdpstore_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/steveyegge/beads/internal/bdp/bdpclient"
	"github.com/steveyegge/beads/internal/bdpstore"
	"github.com/steveyegge/beads/issueops"
)

// TestLifecycleCallMapping validates the lifecycle→`call` mapping against a MOCK
// server. There is NO live oracle (cleanroom-v023 offers no call), so this
// exercises the pattern only: Close succeeds via call issueops/closeIssue, and a
// claim maps a declared Outcome.failure (409) back onto the issueops
// ClaimConflictError (the fault mapping the ruling calls out).
func TestLifecycleCallMapping(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/issueops/closeIssue/-/call", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"outcome":{"kind":"success","value":{"id":"be-1","title":"done","closed_at":"2026-11-26T00:00:00Z"}},"cause":"c","callId":"k"}`))
	})
	mux.HandleFunc("/issueops/claimIssue/-/call", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"outcome":{"kind":"failure","error":{"code":"already-claimed","assignee":"bob"}},"cause":"c","callId":"k"}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	store := bdpstore.New(bdpclient.New(srv.URL))
	lc, err := store.IssueLifecycle()
	if err != nil {
		t.Fatalf("IssueLifecycle: %v", err)
	}

	// Close → success → CloseResult carries the returned Issue.
	cres, err := lc.Close(context.Background(), issueops.CloseRequest{Actor: "sheila", IssueID: "be-1", Reason: "fixed"})
	if err != nil {
		t.Fatalf("Close: %v", err)
	}
	if cres.Issue == nil || cres.Issue.ID != "be-1" || !cres.Changed {
		t.Errorf("Close result = %+v", cres)
	}

	// Claim (Update{Claim:true}) → 409 declared failure → ClaimConflictError,
	// classifiable by errors.Is and carrying the typed fields.
	_, err = lc.Update(context.Background(), issueops.UpdateRequest{Actor: "sheila", IssueID: "be-1", Claim: true})
	if !errors.Is(err, issueops.ErrAlreadyClaimed) {
		t.Errorf("claim conflict: err = %v, want ErrAlreadyClaimed", err)
	}
	var cce *issueops.ClaimConflictError
	if !errors.As(err, &cce) || cce.Assignee != "bob" {
		t.Errorf("claim conflict should be a *ClaimConflictError with assignee bob; got %v", err)
	}

	// A plain (non-claim) Update is not the state machine → unsupported, not a
	// silent no-op or a field-patch of the lease.
	_, err = lc.Update(context.Background(), issueops.UpdateRequest{Actor: "sheila", IssueID: "be-1"})
	if err == nil {
		t.Error("non-claim Update should be unsupported for now")
	}
}
