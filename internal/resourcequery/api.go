// Package resourcequery admits and evaluates the frozen logical Query AST over
// one caller-supplied, complete current dataset. It opens no providers, files,
// transactions or network connections. Dataset authority belongs to its caller.
package resourcequery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	graph "github.com/steveyegge/beads/graphops"
)

const IR = "bdp-query-design-1"

var (
	ErrInvalid     = errors.New("invalid query")
	ErrUnsupported = errors.New("unsupported query capability")
	ErrUnavailable = errors.New("required Resource unavailable")
	ErrLimit       = errors.New("query limit exceeded")
	ErrDataset     = errors.New("invalid query dataset")
)

// Failure classifies an internal failure; Code is not an assigned public Problem.
// At identifies request syntax, never an unavailable/hidden Resource identity.
type Failure struct {
	Code, At, Message string
	cause             error
}

func (f *Failure) Error() string { return fmt.Sprintf("%s: %s", f.Code, f.Message) }
func (f *Failure) Unwrap() error { return f.cause }
func fail(code, at, message string, cause error) error {
	return &Failure{Code: code, At: at, Message: message, cause: cause}
}

// Limits are operational bounds, separate from a walk's semantic hop depth.
// Zero Limits selects defaults; otherwise every field must be positive.
type Limits struct{ RequestBytes, Nodes, Nesting, MaxWork, MaxResults int }

func DefaultLimits() Limits { return Limits{262144, 4096, 64, 1000000, 10000} }
func normalizeLimits(l Limits) (Limits, error) {
	if l == (Limits{}) {
		l = DefaultLimits()
	}
	if l.RequestBytes < 1 || l.Nodes < 1 || l.Nesting < 1 || l.MaxWork < 1 || l.MaxResults < 1 || l.RequestBytes > 1048576 || l.Nodes > 16384 || l.Nesting > 128 {
		return l, fail("invalid-limits", "", "invalid operational limits", ErrInvalid)
	}
	return l, nil
}

// Resource is a complete current record. NewResource validates and copies Data;
// Execute validates again because callers may construct/mutate exported fields.
type Resource struct {
	Kind graph.ResourceKind
	Data json.RawMessage
}

// Dataset is complete for this authorized current read context. TypeFamilies
// contains full reflexive ancestry keyed by nominal Type URI. TypeKinds records
// all known Types, even abstract Types with no current instances.
type Dataset struct {
	Resources []Resource
	// TypeEditions is exact captured authority for recorded Type affiliations.
	// A nonnil slice selects this mechanism, including an empty inventory;
	// TypeFamilies/TypeKinds are never used as a fallback in that mode.
	TypeEditions []TypeEditionFact
	// Nil authority maps mean the family mechanism is unavailable. Nonnil
	// maps assert a complete authorized inventory, including an empty one.
	// Missing queried Types are unavailable data; missing required ancestry
	// for a known Type is an integrity failure in that asserted inventory.
	TypeFamilies map[string][]string
	TypeKinds    map[string]graph.ResourceKind
}

// TypeEditionFact describes one semantic Type edition in the same captured
// context as Resources. Reference is the affiliation, not a local definition
// alias. Ancestors is complete, exact and reflexive. Current requires captured
// live-selection evidence; a retained edition does not imply it. These facts
// do not provide historical Resource acquisition or authorize URI rewriting.
type TypeEditionFact struct {
	Reference Edition
	Kind      graph.ResourceKind
	Current   bool
	Ancestors []Edition
}

// Request is immutable after Decode. Accessors expose only dispatch metadata.
type Request struct {
	raw            []byte
	profile, form  string
	root           *node
	nodes, nesting int
	kind           graph.ResourceKind
	families       []familyUse
	references     []*node
	fields         []*node
}

func (r *Request) Profile() string {
	if r == nil {
		return ""
	}
	return r.profile
}
func (r *Request) ResultForm() string {
	if r == nil {
		return ""
	}
	return r.form
}
func (r *Request) ResourceKind() graph.ResourceKind {
	if r == nil {
		return ""
	}
	return r.kind
}
func (r *Request) Bytes() []byte {
	if r == nil {
		return nil
	}
	return append([]byte(nil), r.raw...)
}

// FieldRoots returns the literal object member at the start of every field
// access, including fields in short-circuited branches. Adapters can refuse a
// known representation mismatch before acquiring data, without reparsing the
// AST or changing evaluation. Numeric first segments have no object root.
// The returned slice is detached; ordinary unknown roots keep their usual
// missing-field behavior unless an adapter explicitly refuses them.
func (r *Request) FieldRoots() []string {
	if r == nil {
		return nil
	}
	roots := make([]string, 0, len(r.fields))
	for _, field := range r.fields {
		if len(field.path) != 0 && !field.path[0].indexed {
			roots = append(roots, field.path[0].key)
		}
	}
	return roots
}

type Edition struct {
	ID       string `json:"id"`
	Revision string `json:"revision"`
}
type EndpointMapping struct {
	Link   Edition `json:"link"`
	Source Edition `json:"source"`
	Target Edition `json:"target"`
}
type ResourcesResult struct {
	Form         string             `json:"form"`
	ResourceKind graph.ResourceKind `json:"resourceKind"`
	Items        []json.RawMessage  `json:"items"`
	Next         *string            `json:"next"`
}
type CountResult struct {
	Form         string             `json:"form"`
	ResourceKind graph.ResourceKind `json:"resourceKind"`
	Count        int                `json:"count"`
}
type GraphResult struct {
	Form      string            `json:"form"`
	Roots     []Edition         `json:"roots"`
	Beads     []json.RawMessage `json:"beads"`
	Links     []json.RawMessage `json:"links"`
	Endpoints []EndpointMapping `json:"endpoints"`
}
type Traversal struct {
	At         string `json:"at"`
	Depth      any    `json:"depth"`
	Completion string `json:"completion"`
	Closure    string `json:"closure"`
}
type Evaluation struct {
	Scope      string      `json:"scope"`
	Complete   bool        `json:"complete"`
	Traversals []Traversal `json:"traversals"`
}

// Response is exactly the proposed native result/evaluation carrier.
type Response struct {
	Result     any        `json:"result"`
	Evaluation Evaluation `json:"evaluation"`
}

// Execute defensively repeats admission before inspecting the dataset. Callers
// can call Decode+Admit first, before acquiring their storage context.
func Execute(ctx context.Context, raw []byte, data Dataset, limits Limits) (Response, error) {
	if ctx == nil {
		return Response{}, fail("invalid-context", "", "a context is required", ErrInvalid)
	}
	if err := ctx.Err(); err != nil {
		return Response{}, err
	}
	l, err := normalizeLimits(limits)
	if err != nil {
		return Response{}, err
	}
	request, err := Decode(raw)
	if err != nil {
		return Response{}, err
	}
	if err = Admit(request, l); err != nil {
		return Response{}, err
	}
	e, err := newEvaluation(ctx, request, data, l)
	if err != nil {
		return Response{}, err
	}
	return e.execute()
}
