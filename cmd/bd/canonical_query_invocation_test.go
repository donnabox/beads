package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/httpapi/bdpwire"
	"github.com/steveyegge/beads/internal/resourcequery"
)

const canonicalQueryInvocationScope = "https://invoke.example/"

type canonicalQueryLocalControl struct {
	scope    string
	response resourcequery.Response
	failure  error
	calls    int
	after    func()
	request  []byte
}

func (c *canonicalQueryLocalControl) scopeURL() string { return c.scope }
func (c *canonicalQueryLocalControl) execute(_ context.Context, raw []byte) (resourcequery.Response, error) {
	c.calls++
	c.request = append([]byte(nil), raw...)
	if c.after != nil {
		c.after()
	}
	return c.response, c.failure
}
func canonicalQueryInvocationForTest(t *testing.T, raw string) canonicalQueryInvocation {
	t.Helper()
	invocation, err := admitCanonicalQueryInvocation([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	return invocation
}
func canonicalQueryGraphInvocationForTest(t *testing.T) canonicalQueryInvocation {
	t.Helper()
	r, err := buildCanonicalGraphRequest(canonicalGraphRequestOptions{Roots: []bdpwire.Reference{{URI: canonicalQueryInvocationScope + "root"}}, Direction: "outgoing", Depth: json.RawMessage(`1`)})
	if err != nil {
		t.Fatal(err)
	}
	i, err := canonicalQueryInvocationFromRequest(r)
	if err != nil {
		t.Fatal(err)
	}
	return i
}
func canonicalQueryControlResponse(result any) resourcequery.Response {
	return resourcequery.Response{Result: result, Evaluation: resourcequery.Evaluation{Scope: "response", Complete: true, Traversals: []resourcequery.Traversal{{At: "/query/selection", Depth: 1, Completion: "depth", Closure: "unknown"}}}}
}

func TestCanonicalQueryInvocationOwnsAdmittedRequest(t *testing.T) {
	raw := []byte(queryASTCountFixture)
	i, err := admitCanonicalQueryInvocation(raw)
	if err != nil {
		t.Fatal(err)
	}
	before := i.Bytes()
	raw[0] = 'x'
	exposed := i.Bytes()
	exposed[0] = 'x'
	if !bytes.Equal(before, i.Bytes()) || i.Profile() != "list" || i.ResultForm() != "count" || i.ResourceKind() != graph.KindBead {
		t.Fatal("request aliases caller bytes or lost dispatch facts")
	}
	for _, request := range []*resourcequery.Request{nil, &resourcequery.Request{}} {
		if got, err := canonicalQueryInvocationFromRequest(request); !errors.Is(err, resourcequery.ErrInvalid) || got.request != nil {
			t.Fatalf("unadmitted Request accepted: %+v %v", got, err)
		}
	}
	if got, err := admitCanonicalQueryInvocation([]byte(`{"ir":"other"}`)); err == nil || got.request != nil {
		t.Fatal("invalid raw request admitted")
	}
}

func TestCanonicalQueryLocalScopeAdmissionAndCancellation(t *testing.T) {
	i := canonicalQueryInvocationForTest(t, queryASTCountFixture)
	for _, mode := range []string{"nil-context", "canceled", "zero-invocation", "empty-scope", "noncanonical-scope", "different-scope", "nil-executor"} {
		t.Run(mode, func(t *testing.T) {
			ctx := context.Background()
			scope := canonicalQueryInvocationScope
			invocation := i
			control := &canonicalQueryLocalControl{scope: scope}
			var executor canonicalQueryLocalExecutor = control
			switch mode {
			case "nil-context":
				ctx = nil
			case "canceled":
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			case "zero-invocation":
				invocation = canonicalQueryInvocation{}
			case "empty-scope":
				scope = ""
			case "noncanonical-scope":
				scope = "https://invoke.example"
			case "different-scope":
				control.scope = "https://other.example/"
			case "nil-executor":
				executor = nil
			}
			got, err := invokeCanonicalQueryLocal(ctx, scope, invocation, executor)
			if err == nil || !reflect.DeepEqual(got, canonicalLocalQueryResult{}) || control.calls != 0 {
				t.Fatalf("invalid/canceled input acquired or yielded data: %+v calls%d %v", got, control.calls, err)
			}
		})
	}
}

func TestCanonicalQueryLocalResponseDeeplyDetached(t *testing.T) {
	for _, form := range []string{"resources", "count", "graph"} {
		t.Run(form, func(t *testing.T) {
			i := canonicalQueryInvocationForTest(t, queryASTCountFixture)
			items := []json.RawMessage{json.RawMessage(`{"id":"https://invoke.example/a","properties":{"title":"<keep&>"}}`)}
			roots := []resourcequery.Edition{{ID: canonicalQueryInvocationScope + "a", Revision: "r1"}}
			endpoints := []resourcequery.EndpointMapping{{Link: resourcequery.Edition{ID: canonicalQueryInvocationScope + "link", Revision: "l1"}, Source: roots[0], Target: roots[0]}}
			var selected any = resourcequery.CountResult{Form: "count", ResourceKind: graph.KindBead, Count: 1}
			if form == "resources" {
				i = canonicalQueryInvocationForTest(t, `{"ir":"bdp-query-design-1","profile":"list","selection":{"op":"scan","collection":"beads"},"result":{"form":"resources"}}`)
				selected = resourcequery.ResourcesResult{Form: "resources", ResourceKind: graph.KindBead, Items: items}
			} else if form == "graph" {
				i = canonicalQueryGraphInvocationForTest(t)
				selected = resourcequery.GraphResult{Form: "graph", Roots: roots, Beads: items, Links: items, Endpoints: endpoints}
			}
			response := canonicalQueryControlResponse(selected)
			// The boundary preserves metadata from its Reader, rather than
			// comparing Evaluation.Scope to the owning service URL or inventing
			// a completion rule. This control claims no evaluator semantics.
			response.Evaluation.Complete = false
			control := &canonicalQueryLocalControl{scope: canonicalQueryInvocationScope, response: response}
			got, err := invokeCanonicalQueryLocal(context.Background(), canonicalQueryInvocationScope, i, control)
			if err != nil {
				t.Fatal(err)
			}
			before := got.Bytes()
			response.Evaluation.Traversals[0].Closure = "changed"
			items[0][0] = 'x'
			roots[0].ID = "changed"
			endpoints[0].Source.ID = "changed"
			exposed := got.Bytes()
			exposed[0] = 'x'
			exposedValue := got.Value()
			exposedValue[0] = 'x'
			if !bytes.Equal(before, got.Bytes()) || got.ScopeURL() != canonicalQueryInvocationScope || got.StatusCode() != 200 || got.Headers() != nil || got.ResultForm() != form || !bytes.Equal(control.request, i.Bytes()) {
				t.Fatal("local result aliases mutable response/exposed bytes or lost native dispatch")
			}
			var detached struct{ Evaluation resourcequery.Evaluation }
			if err := json.Unmarshal(got.Bytes(), &detached); err != nil || detached.Evaluation.Complete || detached.Evaluation.Scope != "response" || detached.Evaluation.Traversals[0].Closure != "unknown" {
				t.Fatalf("evaluation was reinterpreted: %+v %v", detached, err)
			}
		})
	}
}

func TestCanonicalQueryLocalErrorsAndResultDispatch(t *testing.T) {
	i := canonicalQueryInvocationForTest(t, queryASTCountFixture)
	for _, cause := range []error{resourcequery.ErrInvalid, resourcequery.ErrUnsupported, resourcequery.ErrUnavailable, resourcequery.ErrLimit, resourcequery.ErrDataset, context.Canceled} {
		c := &canonicalQueryLocalControl{scope: canonicalQueryInvocationScope, response: canonicalQueryControlResponse(resourcequery.CountResult{Form: "count", ResourceKind: graph.KindBead}), failure: cause}
		got, err := invokeCanonicalQueryLocal(context.Background(), c.scope, i, c)
		if !errors.Is(err, cause) || !reflect.DeepEqual(got, canonicalLocalQueryResult{}) || c.calls != 1 {
			t.Fatalf("lost error or partial result: %+v %v", got, err)
		}
	}
	for _, result := range []any{nil, resourcequery.ResourcesResult{Form: "count", ResourceKind: graph.KindBead}, resourcequery.CountResult{Form: "resources", ResourceKind: graph.KindBead}, resourcequery.CountResult{Form: "count", ResourceKind: graph.KindLink}, resourcequery.GraphResult{Form: "count"}} {
		c := &canonicalQueryLocalControl{scope: canonicalQueryInvocationScope, response: canonicalQueryControlResponse(result)}
		got, err := invokeCanonicalQueryLocal(context.Background(), c.scope, i, c)
		if !errors.Is(err, resourcequery.ErrDataset) || !reflect.DeepEqual(got, canonicalLocalQueryResult{}) {
			t.Fatalf("wrong actual result type/form/kind admitted: %+v %v", got, err)
		}
	}
	bad := canonicalQueryControlResponse(resourcequery.CountResult{Form: "count", ResourceKind: graph.KindBead})
	bad.Evaluation.Traversals[0].Depth = make(chan int)
	control := &canonicalQueryLocalControl{scope: canonicalQueryInvocationScope, response: bad}
	if got, err := invokeCanonicalQueryLocal(context.Background(), control.scope, i, control); !errors.Is(err, resourcequery.ErrDataset) || !reflect.DeepEqual(got, canonicalLocalQueryResult{}) {
		t.Fatalf("serialization failure returned partial success: %+v %v", got, err)
	}
	for _, mode := range []string{"late-cancel", "scope-changed"} {
		ctx, cancel := context.WithCancel(context.Background())
		c := &canonicalQueryLocalControl{scope: canonicalQueryInvocationScope, response: canonicalQueryControlResponse(resourcequery.CountResult{Form: "count", ResourceKind: graph.KindBead})}
		want := resourcequery.ErrDataset
		if mode == "late-cancel" {
			c.after = cancel
			want = context.Canceled
		} else {
			c.after = func() { c.scope = "https://changed.example/" }
		}
		got, err := invokeCanonicalQueryLocal(ctx, canonicalQueryInvocationScope, i, c)
		cancel()
		if !errors.Is(err, want) || !reflect.DeepEqual(got, canonicalLocalQueryResult{}) {
			t.Fatalf("late change returned success: %+v %v", got, err)
		}
	}
}

func TestCanonicalQueryLocalBodyPreservesRendererOutput(t *testing.T) {
	for _, form := range []string{"count", "resources", "graph"} {
		i := canonicalQueryInvocationForTest(t, queryASTCountFixture)
		var result any = resourcequery.CountResult{Form: "count", ResourceKind: graph.KindBead, Count: 1}
		items := []json.RawMessage{json.RawMessage(`{"properties":{"title":"<keep&> 記憶"}}`)}
		if form == "resources" {
			i = canonicalQueryInvocationForTest(t, `{"ir":"bdp-query-design-1","profile":"list","selection":{"op":"scan","collection":"beads"},"result":{"form":"resources"}}`)
			result = resourcequery.ResourcesResult{Form: "resources", ResourceKind: graph.KindBead, Items: items}
		} else if form == "graph" {
			i = canonicalQueryGraphInvocationForTest(t)
			result = resourcequery.GraphResult{Form: "graph", Roots: []resourcequery.Edition{}, Beads: items, Links: []json.RawMessage{}, Endpoints: []resourcequery.EndpointMapping{}}
		}
		response := canonicalQueryControlResponse(result)
		c := &canonicalQueryLocalControl{scope: canonicalQueryInvocationScope, response: response}
		got, err := invokeCanonicalQueryLocal(context.Background(), c.scope, i, c)
		if err != nil {
			t.Fatal(err)
		}
		for _, format := range []string{"json", "default", "detail"} {
			for _, raw := range []bool{false, true} {
				if raw && format != "json" {
					continue
				} // Existing renderer admits raw JSON only.
				var want bytes.Buffer
				cmd := queryASTTestCommand()
				if err := cmd.Flags().Set("output-format", format); err != nil {
					t.Fatal(err)
				}
				if raw {
					if err := cmd.Flags().Set("omit", "envelope"); err != nil {
						t.Fatal(err)
					}
				}
				renderer, err := queryASTRenderer(cmd, form)
				if err != nil {
					t.Fatal(err)
				}
				cmd.SetOut(&want)
				if err := renderCanonicalResult(cmd, renderer, response); err != nil {
					t.Fatal(err)
				}
				var actual bytes.Buffer
				cmd.SetOut(&actual)
				if err := renderCanonicalResult(cmd, renderer, got.Value()); err != nil || !bytes.Equal(want.Bytes(), actual.Bytes()) {
					t.Fatalf("changed %s/%s raw%t bytes: %q/%q %v", form, format, raw, want.Bytes(), actual.Bytes(), err)
				}
			}
		}
	}
}

func TestCanonicalGraphLocalAdmitsBeforeReader(t *testing.T) {
	options := canonicalGraphRequestOptions{Roots: []bdpwire.Reference{{URI: canonicalQueryInvocationScope + "root", Revision: "opaque-pin"}}, Direction: "outgoing", Depth: json.RawMessage(`1`)}
	got, err := invokeCanonicalGraphLocal(context.Background(), canonicalQueryInvocationScope, nil, options)
	if !errors.Is(err, resourcequery.ErrUnsupported) || !reflect.DeepEqual(got, canonicalLocalQueryResult{}) {
		t.Fatalf("Graph admission reached reader or dropped pin: %+v %v", got, err)
	}
	options.Roots[0].Revision = ""
	got, err = invokeCanonicalGraphLocal(context.Background(), canonicalQueryInvocationScope, nil, options)
	if !errors.Is(err, resourcequery.ErrInvalid) || !reflect.DeepEqual(got, canonicalLocalQueryResult{}) {
		t.Fatalf("Graph fabricated a local reader: %+v %v", got, err)
	}
}
