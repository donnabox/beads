package bdpclient_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/bdp"
	"github.com/steveyegge/beads/internal/bdp/bdpclient"
)

// TestSmokeAgainstCleanroomServer round-trips the mapped primitives against a
// RUNNING clean-room BDP server. It is skipped unless BDP_SMOKE_ADDR names one
// (e.g. "http://127.0.0.1:8799"), so it is inert in CI, which has no such server.
// Stand the server up with the harness in the session scratchpad
// (bdp-smoke-server), then: BDP_SMOKE_ADDR=http://127.0.0.1:8799 go test ./internal/bdp/bdpclient/ -run Smoke -v
func TestSmokeAgainstCleanroomServer(t *testing.T) {
	base := os.Getenv("BDP_SMOKE_ADDR")
	if base == "" {
		t.Skip("set BDP_SMOKE_ADDR to a running clean-room server to run the smoke test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s := bdpclient.New(base)

	// Capabilities advertises exactly the mapped set, get included.
	caps, err := s.Capabilities(ctx)
	if err != nil {
		t.Fatalf("Capabilities: %v", err)
	}
	if !caps.Offers(bdp.PrimitiveGet) || !caps.Offers(bdp.PrimitivePut) {
		t.Fatalf("Capabilities missing get/put: %+v", caps.Core)
	}

	// Discover: the full conformance handshake — decode the ServiceDescriptor off
	// the live server and negotiate its protocol version (0.2.x).
	desc, dcaps, err := s.Discover(ctx)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if desc.Version == "" {
		t.Error("Discover: ServiceDescriptor carried no protocol version")
	}
	if !dcaps.Offers(bdp.PrimitiveGet) {
		t.Error("Discover: capabilities must offer get (the floor)")
	}

	note1 := bdp.ResourceID{ID: "note-1"}

	// 1) CREATE note-1 → pinned ref with a minted rev.
	createRef, err := s.Create(ctx, bdp.CreateInput{
		ID:   bdp.IDProposal{Exact: "note-1"},
		Type: "issue/task",
		Body: json.RawMessage(`{"title":"hello","done":false}`),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if createRef.ID != "note-1" {
		t.Errorf("Create ref id = %q, want note-1", createRef.ID)
	}
	if createRef.Rev == "" {
		t.Error("Create must return a pinned ref (minted rev)")
	}

	// 2) GET note-1 → representation round-trips id, rev, and a body field.
	rep, err := s.Get(ctx, note1, nil)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if rep.ID.ID != "note-1" {
		t.Errorf("Get id = %q, want note-1", rep.ID.ID)
	}
	if rep.ID.Rev != createRef.Rev {
		t.Errorf("Get rev = %q, want %q (the created rev)", rep.ID.Rev, createRef.Rev)
	}
	var body struct {
		Title string `json:"title"`
		Done  bool   `json:"done"`
	}
	if err := json.Unmarshal(rep.Body, &body); err != nil {
		t.Fatalf("Get body not JSON object: %v (%s)", err, rep.Body)
	}
	if body.Title != "hello" || body.Done {
		t.Errorf("Get body = %+v, want {hello false}", body)
	}

	// 3) Pinned GET at that exact revision still resolves.
	pinned := bdp.ResourceID{ID: "note-1", Rev: createRef.Rev}
	if _, err := s.Get(ctx, pinned, nil); err != nil {
		t.Errorf("pinned Get: %v", err)
	}

	// 4) PUT note-1 wholesale → a new rev, different from create.
	putRef, err := s.Put(ctx, bdp.PutInput{
		Target: note1,
		Type:   "issue/task",
		Body:   json.RawMessage(`{"title":"hello again","done":true}`),
	})
	if err != nil {
		t.Fatalf("Put: %v", err)
	}
	if putRef.Rev == "" || putRef.Rev == createRef.Rev {
		t.Errorf("Put rev = %q, want a fresh rev (!= %q)", putRef.Rev, createRef.Rev)
	}

	// 4b) LIST with a SERVER-SIDE filter → confirm the clean-room evaluates
	//     body-field Compare conditions (the push-down finding). After the put,
	//     note-1's body/title is "hello again".
	hit, err := s.List(ctx, bdp.Namespace{}, bdp.ListFilter{{Field: "body/title", Op: "eq", Value: json.RawMessage(`"hello again"`)}}, nil, nil)
	if err != nil {
		t.Fatalf("List(matching filter): %v", err)
	}
	if !containsID(hit.Items, "note-1") {
		t.Error("server-side filter body/title=\"hello again\" should include note-1")
	}
	miss, err := s.List(ctx, bdp.Namespace{}, bdp.ListFilter{{Field: "body/title", Op: "eq", Value: json.RawMessage(`"no-such-title"`)}}, nil, nil)
	if err != nil {
		t.Fatalf("List(non-matching filter): %v", err)
	}
	if containsID(miss.Items, "note-1") {
		t.Error("server-side filter should EXCLUDE note-1 for a non-matching title — clean-room is not filtering body fields")
	}

	// 5) GET a missing resource → ErrNotFound, classifiable with errors.Is.
	_, err = s.Get(ctx, bdp.ResourceID{ID: "does-not-exist"}, nil)
	if !errors.Is(err, bdp.ErrNotFound) {
		t.Errorf("Get missing: err = %v, want ErrNotFound", err)
	}

	// 6) CREATE note-1 again → identity conflict (409) → ErrAlreadyExists.
	_, err = s.Create(ctx, bdp.CreateInput{
		ID:   bdp.IDProposal{Exact: "note-1"},
		Type: "issue/task",
		Body: json.RawMessage(`{"title":"dup"}`),
	})
	if !errors.Is(err, bdp.ErrAlreadyExists) {
		t.Errorf("duplicate Create: err = %v, want ErrAlreadyExists", err)
	}

	// 7) DELETE note-1 → tombstone rev; then GET → gone.
	tombstone, err := s.Delete(ctx, note1, "", nil)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if tombstone == "" {
		t.Error("Delete must return a tombstone rev")
	}
	if _, err := s.Get(ctx, note1, nil); !errors.Is(err, bdp.ErrNotFound) {
		t.Errorf("Get after delete: err = %v, want ErrNotFound", err)
	}

	// 8) An unmapped primitive faults as unsupported (server answers 501; the
	//    client short-circuits via UnimplementedScope).
	_, err = s.Call(ctx, bdp.ResourceID{ID: "op"}, bdp.Request{})
	if err == nil {
		t.Error("Call should fault (unmapped primitive)")
	}
}

func containsID(items []bdp.Representation, id string) bool {
	for _, it := range items {
		if it.ID.ID == id {
			return true
		}
	}
	return false
}
