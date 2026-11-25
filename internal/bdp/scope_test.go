package bdp

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/steveyegge/beads/beadserrors"
)

// minimalScope is the smallest conformant scope: it embeds UnimplementedScope
// (inheriting the refusing default for every optional primitive) and writes
// only the two mandatory members, Capabilities and Get. That this compiles is
// the point — the compiler enforces "get is the only mandatory primitive" (plus
// discovery), because UnimplementedScope provides neither.
type minimalScope struct {
	UnimplementedScope
}

func (minimalScope) Capabilities(context.Context) (Capabilities, error) {
	return Capabilities{Core: CoreCapabilities{Get: GetCaps{Selector: true}}}, nil
}

func (minimalScope) Get(_ context.Context, ref ResourceID, _ *Selector) (Representation, error) {
	if ref.ID == "" {
		return Representation{}, ErrNotFound
	}
	return Representation{ID: ref, Type: "text/plain", Body: json.RawMessage(`"hello"`)}, nil
}

// Compile-time proof that the minimal scope satisfies the full contract.
var _ Scope = minimalScope{}

func TestMinimalScopeGet(t *testing.T) {
	var s Scope = minimalScope{}
	rep, err := s.Get(context.Background(), ResourceID{ID: "x"}, nil)
	if err != nil {
		t.Fatalf("Get: unexpected error: %v", err)
	}
	if rep.Type != "text/plain" {
		t.Errorf("Get: type = %q, want text/plain", rep.Type)
	}
}

func TestCapabilitiesOffers(t *testing.T) {
	caps, err := minimalScope{}.Capabilities(context.Background())
	if err != nil {
		t.Fatalf("Capabilities: %v", err)
	}
	if !caps.Offers(PrimitiveGet) {
		t.Error("minimal scope must offer get")
	}
	if caps.Offers(PrimitiveCall) {
		t.Error("minimal scope must not offer call")
	}
}

// An optional primitive a scope does not implement faults with a typed,
// classifiable *ErrUnsupported naming the operation and the backend.
func TestUnimplementedOptionalFaults(t *testing.T) {
	s := struct{ UnimplementedScope }{UnimplementedScope{Backend: "test scope"}}

	_, err := s.Create(context.Background(), CreateInput{Type: "issue/task"})
	var unsup *beadserrors.ErrUnsupported
	if !errors.As(err, &unsup) {
		t.Fatalf("Create on an unimplemented scope: err = %v, want *ErrUnsupported", err)
	}
	if unsup.Op != "create" {
		t.Errorf("ErrUnsupported.Op = %q, want create", unsup.Op)
	}
	if unsup.Backend != "test scope" {
		t.Errorf("ErrUnsupported.Backend = %q, want \"test scope\"", unsup.Backend)
	}

	_, err = s.Call(context.Background(), ResourceID{ID: "op"}, Request{})
	if !errors.As(err, &unsup) || unsup.Op != "call" {
		t.Errorf("Call on an unimplemented scope: err = %v, want ErrUnsupported{Op:call}", err)
	}
}

// Faults map onto sentinels so a caller classifies them with errors.Is, and a
// FaultError still exposes its Problem via errors.As.
func TestFaultErrorClassification(t *testing.T) {
	fe := NewFaultError(Problem{Status: 412, Code: "revision-mismatch", Title: "stale"}, ErrRevisionMismatch)
	if !errors.Is(fe, ErrRevisionMismatch) {
		t.Error("FaultError must match its mapped sentinel via errors.Is")
	}
	var got *FaultError
	if !errors.As(fe, &got) || got.Problem.Status != 412 {
		t.Errorf("FaultError must expose its Problem via errors.As; got %+v", got)
	}
}

func TestNamespaceLocal(t *testing.T) {
	if !(Namespace{}).Local() {
		t.Error("zero Namespace must be local")
	}
	if (Namespace{Store: "https://x"}).Local() {
		t.Error("a store namespace is not local")
	}
	if (Namespace{URI: true}).Local() {
		t.Error("a uri namespace is not local")
	}
}
