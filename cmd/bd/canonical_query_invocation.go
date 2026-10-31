package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/resourcequery"
)

// canonicalQueryInvocation owns an already admitted immutable Request. It
// contains no provider/configuration/default, opens nothing and grants no Scope
// ownership. Query and resolved Graph share this request boundary.
type canonicalQueryInvocation struct{ request *resourcequery.Request }

func admitCanonicalQueryInvocation(raw []byte) (canonicalQueryInvocation, error) {
	request, err := resourcequery.Decode(raw)
	if err != nil {
		return canonicalQueryInvocation{}, err
	}
	return canonicalQueryInvocationFromRequest(request)
}

func canonicalQueryInvocationFromRequest(request *resourcequery.Request) (canonicalQueryInvocation, error) {
	// Admit also refuses a caller's nil/zero Request. Its private fields and
	// detached Bytes accessor are the existing request immutability boundary.
	if err := resourcequery.Admit(request, resourcequery.DefaultLimits()); err != nil {
		return canonicalQueryInvocation{}, err
	}
	return canonicalQueryInvocation{request: request}, nil
}

func (i canonicalQueryInvocation) Bytes() []byte                    { return i.request.Bytes() }
func (i canonicalQueryInvocation) Profile() string                  { return i.request.Profile() }
func (i canonicalQueryInvocation) ResultForm() string               { return i.request.ResultForm() }
func (i canonicalQueryInvocation) ResourceKind() graph.ResourceKind { return i.request.ResourceKind() }

// This is a LOCAL execution seam. Implementations cannot stand in for HTTP
// responses: configured transport must preserve its own actual status, headers,
// body, Problem and bodyless disposition instead of flattening them here.
type canonicalQueryLocalExecutor interface {
	scopeURL() string
	execute(context.Context, []byte) (resourcequery.Response, error)
}

// canonicalLocalQueryResult is detached local success data, not an observed
// HTTP response. Logical200 has no fabricated HTTP headers. Native Result and
// Evaluation are retained together as bytes; no evaluation Scope/Complete/depth
// or completion/closure interpretation occurs at this invocation boundary.
type canonicalLocalQueryResult struct {
	owningScope string
	form        string
	kind        graph.ResourceKind
	body        []byte
}

func (r canonicalLocalQueryResult) ScopeURL() string                 { return r.owningScope }
func (r canonicalLocalQueryResult) ResultForm() string               { return r.form }
func (r canonicalLocalQueryResult) ResourceKind() graph.ResourceKind { return r.kind }
func (r canonicalLocalQueryResult) StatusCode() int {
	if r.body == nil {
		return 0
	}
	return 200
}
func (r canonicalLocalQueryResult) Headers() map[string][]string { return nil }
func (r canonicalLocalQueryResult) Bytes() []byte                { return append([]byte(nil), r.body...) }
func (r canonicalLocalQueryResult) Value() json.RawMessage {
	return append(json.RawMessage(nil), r.body...)
}

func invokeCanonicalQueryLocal(ctx context.Context, owningScope string, invocation canonicalQueryInvocation, executor canonicalQueryLocalExecutor) (canonicalLocalQueryResult, error) {
	if ctx == nil {
		return canonicalLocalQueryResult{}, fmt.Errorf("%w: Query invocation requires a context", resourcequery.ErrInvalid)
	}
	if err := ctx.Err(); err != nil {
		return canonicalLocalQueryResult{}, err
	}
	// Do not infer admission from an opaque pointer's presence, or consult a
	// provider to rescue malformed/unsupported input.
	if err := resourcequery.Admit(invocation.request, resourcequery.DefaultLimits()); err != nil {
		return canonicalLocalQueryResult{}, err
	}
	if err := graph.ValidateScopeURL(owningScope); err != nil {
		return canonicalLocalQueryResult{}, fmt.Errorf("%w: invalid selected Query Scope: %v", resourcequery.ErrInvalid, err)
	}
	if executor == nil || executor.scopeURL() != owningScope {
		return canonicalLocalQueryResult{}, fmt.Errorf("%w: local Query Scope differs from selected Scope", resourcequery.ErrDataset)
	}
	response, err := executor.execute(ctx, invocation.Bytes())
	if err != nil {
		return canonicalLocalQueryResult{}, err
	}
	if err := ctx.Err(); err != nil {
		return canonicalLocalQueryResult{}, err
	}
	if executor.scopeURL() != owningScope {
		return canonicalLocalQueryResult{}, fmt.Errorf("%w: local Query Scope changed during execution", resourcequery.ErrDataset)
	}
	// Match only dispatch facts of the actual typed native result. The checked
	// Reader/evaluator owns record/Type/graph semantics and evaluation metadata.
	form, kind := "", graph.ResourceKind("")
	switch result := response.Result.(type) {
	case resourcequery.ResourcesResult:
		if result.Form != "resources" {
			return canonicalLocalQueryResult{}, fmt.Errorf("%w: invalid local Resources form", resourcequery.ErrDataset)
		}
		form, kind = result.Form, result.ResourceKind
	case resourcequery.CountResult:
		if result.Form != "count" {
			return canonicalLocalQueryResult{}, fmt.Errorf("%w: invalid local Count form", resourcequery.ErrDataset)
		}
		form, kind = result.Form, result.ResourceKind
	case resourcequery.GraphResult:
		if result.Form != "graph" {
			return canonicalLocalQueryResult{}, fmt.Errorf("%w: invalid local Graph form", resourcequery.ErrDataset)
		}
		form, kind = result.Form, graph.KindLink
	default:
		return canonicalLocalQueryResult{}, fmt.Errorf("%w: local Query returned no native result", resourcequery.ErrDataset)
	}
	if form != invocation.ResultForm() || kind != invocation.ResourceKind() {
		return canonicalLocalQueryResult{}, fmt.Errorf("%w: local Query result form/kind differs from admitted request", resourcequery.ErrDataset)
	}
	var body bytes.Buffer
	encoder := json.NewEncoder(&body)
	// Match the existing CLI renderer's serialization, including HTML bytes.
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(response); err != nil {
		return canonicalLocalQueryResult{}, fmt.Errorf("%w: detach local Query response: %v", resourcequery.ErrDataset, err)
	}
	if err := ctx.Err(); err != nil {
		return canonicalLocalQueryResult{}, err
	}
	return canonicalLocalQueryResult{owningScope: owningScope, form: form, kind: kind, body: append([]byte(nil), body.Bytes()...)}, nil
}
