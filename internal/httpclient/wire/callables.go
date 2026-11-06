package wire

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/steveyegge/beads/issueops"
)

// Every operation but two reaches bd serve as a Callable: POST /bdp/call/NAME with
// its arguments in the invocation envelope, answered under a result envelope or a
// Problem carrying v0's own body. Request stays the vocabulary of everything above
// this file. Its Method, Path and Query still say which v0 operation this is, and
// the read memo and the two retries key on them, so a read is still a read. Only
// what goes on the wire differs, and everything that differs is here.
//
// The two exceptions: `health` is not an operation on a Scope and needs no
// credential, and `watchEvents` holds a stream open, which an invocation cannot
// carry; it goes through its own door in stream.go.
//
// bd connect talks only to the bd serve of its own release. There is no v0
// fallback and no capability for this: a server without the mount is told apart
// by its answer and refused in words (CallablesAbsentError).
const (
	callableMount     = "/bdp/call"
	callableMediaType = "application/vnd.bdp.v0.1+json"
	callableAccept    = callableMediaType + ", application/problem+json"
	// callableEnvelopeSlack is what the result envelope adds around a v0 body:
	// {"changed":false,"result":...}\n needs less than this. It keeps a body exactly
	// at Options.MaxResponseBytes readable, as it was before the envelope.
	callableEnvelopeSlack = 64
)

// callable is how one operation is invoked: the Callable's name, and which of its
// arguments a Request carries in PathArg.
type callable struct {
	name     string
	pathName string
}

// callables is the operation-to-Callable map, written from
// specs/types/issue/callables.md. internal/httpapi's
// TestTheClientCallableMapMatchesTheServerTable holds it to the server's table.
// issue/prune shares sweepIssues with issue/purge; see callableFor.
var callables = map[string]callable{
	OpGetContext:              {name: "issue/context"},
	OpListReadyWork:           {name: "issue/ready"},
	OpCountReadyWork:          {name: "issue/ready-count"},
	OpGetStats:                {name: "issue/stats"},
	OpListBlockedIssues:       {name: "issue/blocked"},
	OpListDependencyCycles:    {name: "issue/dep/cycles"},
	OpGetDependencyTree:       {name: "issue/dep/tree"},
	OpListIssues:              {name: "issue/list"},
	OpCreateIssue:             {name: "issue/create"},
	OpQueryIssues:             {name: "issue/query"},
	OpCountIssues:             {name: "issue/count"},
	OpGetIssue:                {name: "issue/show", pathName: "id"},
	OpUpdateIssue:             {name: "issue/update", pathName: "id"},
	OpListRelatedIssues:       {name: "issue/dep/list", pathName: "id"},
	OpAddComment:              {name: "issue/comment", pathName: "id"},
	OpClaimIssue:              {name: "issue/claim", pathName: "id"},
	OpReleaseIssue:            {name: "issue/unclaim", pathName: "id"},
	OpHeartbeatIssue:          {name: "issue/heartbeat", pathName: "id"},
	OpCloseIssue:              {name: "issue/close", pathName: "id"},
	OpAdvanceMolecule:         {name: "issue/advance", pathName: "id"},
	OpReopenIssue:             {name: "issue/reopen", pathName: "id"},
	OpCompareAndSetMetadata:   {name: "issue/metadata/compare-and-set", pathName: "id"},
	OpSweepIssues:             {name: "issue/purge"},
	OpDeleteIssues:            {name: "issue/delete"},
	OpBatchGetIssues:          {name: "issue/get-many"},
	OpReclaimIssues:           {name: "issue/reclaim"},
	OpBatchCreateIssues:       {name: "issue/create-many"},
	OpApplyBatch:              {name: "issue/batch"},
	OpBatchCloseIssues:        {name: "issue/close-many"},
	OpClaimNextIssue:          {name: "issue/claim-next"},
	OpListDependencies:        {name: "issue/dep/edges"},
	OpCountDependencyEdges:    {name: "issue/dep/count"},
	OpListBlockingAnnotations: {name: "issue/blocking-annotations"},
	OpAddDependencies:         {name: "issue/dep/add"},
	OpRemoveDependency:        {name: "issue/dep/remove"},
	OpListSettings:            {name: "config/list"},
	OpGetSetting:              {name: "config/get", pathName: "key"},
	OpSetSetting:              {name: "config/set", pathName: "key"},
	OpUnsetSetting:            {name: "config/unset", pathName: "key"},
	OpListMemories:            {name: "memory/list"},
	OpRememberMemory:          {name: "memory/remember"},
	OpGetMemory:               {name: "memory/recall", pathName: "key"},
	OpForgetMemory:            {name: "memory/forget", pathName: "key"},
	OpListEvents:              {name: "events/list"},
}

// outbound is what one Request puts on the wire.
type outbound struct {
	method string
	url    *url.URL
	body   []byte
	// callable names the Callable invoked, and says the answer comes back under
	// its envelope. "" is a plain v0 request.
	callable string
}

// outboundFor turns a Request into its wire form. logical is the v0 URL, and body
// the marshaled Body. An operation with no Callable that is not one of the
// exceptions is refused rather than sent to /v0: nothing reaches v0 by accident.
func (c *Client) outboundFor(r Request, logical *url.URL, body []byte) (*outbound, error) {
	if r.Op == OpHealth {
		return &outbound{method: r.Method, url: logical, body: body}, nil
	}
	call, err := callableFor(r, body)
	if err != nil {
		return nil, err
	}
	envelope, err := invocationEnvelope(r, call, body)
	if err != nil {
		return nil, err
	}
	return &outbound{
		method:   http.MethodPost,
		url:      c.base.JoinPath(callableMount, call.name),
		body:     envelope,
		callable: call.name,
	}, nil
}

// callableFor resolves a Request's Callable. sweepIssues is the one operation two
// Callables route to, told apart by what the sweep clears: `ephemeral` is bd purge
// and `durable` is bd prune; the wider `wisps-plane` starts from the wisps table,
// and so is a purge.
func callableFor(r Request, body []byte) (callable, error) {
	call, ok := callables[r.Op]
	if !ok {
		return callable{}, fmt.Errorf("%s: no Callable carries this operation", r.Op)
	}
	if r.Op == OpSweepIssues {
		var sweep struct {
			Tier string `json:"tier"`
		}
		if err := json.Unmarshal(body, &sweep); err != nil {
			return callable{}, fmt.Errorf("%s: reading the sweep's tier: %w", r.Op, err)
		}
		if sweep.Tier == "durable" {
			call.name = "issue/prune"
		}
	}
	return call, nil
}

// invocationEnvelope spells a v0 request's parameters as one `arguments` object:
// the path segment under its name, query values as strings (arrays when repeated,
// which is how bd serve reads them back) and the body's members as they are.
func invocationEnvelope(r Request, call callable, body []byte) ([]byte, error) {
	if call.pathName == "" && r.PathArg != "" {
		return nil, fmt.Errorf("%s: %s takes no path argument", r.Op, call.name)
	}
	if call.pathName != "" && r.PathArg == "" {
		return nil, fmt.Errorf("%s: %s needs its %s", r.Op, call.name, call.pathName)
	}

	var query map[string]json.RawMessage
	for name, values := range r.Query {
		var (
			raw []byte
			err error
		)
		switch len(values) {
		case 0:
			continue
		case 1:
			raw, err = json.Marshal(values[0])
		default:
			raw, err = json.Marshal(values)
		}
		if err != nil {
			return nil, fmt.Errorf("%s: encoding query parameter %s: %w", r.Op, name, err)
		}
		if query == nil {
			query = make(map[string]json.RawMessage, len(r.Query))
		}
		query[name] = raw
	}

	// The common large case, a body and nothing else, needs no decoding: marshaled
	// from a Go value, it is already one object with unique members.
	if call.pathName == "" && len(query) == 0 && bytes.HasPrefix(body, []byte("{")) {
		envelope := make([]byte, 0, len(body)+len(`{"arguments":}`))
		envelope = append(envelope, `{"arguments":`...)
		envelope = append(envelope, body...)
		return append(envelope, '}'), nil
	}

	arguments := map[string]json.RawMessage{}
	if len(body) > 0 {
		if err := json.Unmarshal(body, &arguments); err != nil {
			return nil, fmt.Errorf("%s: a request body must be a JSON object to be invoked: %w", r.Op, err)
		}
		if arguments == nil { // A nil map marshals to null.
			arguments = map[string]json.RawMessage{}
		}
	}
	for name, raw := range query {
		if _, taken := arguments[name]; taken {
			return nil, fmt.Errorf("%s: %s is both a query parameter and a body member", r.Op, name)
		}
		arguments[name] = raw
	}
	if call.pathName != "" {
		if _, taken := arguments[call.pathName]; taken {
			return nil, fmt.Errorf("%s: %s is both a path argument and another parameter", r.Op, call.pathName)
		}
		raw, err := json.Marshal(r.PathArg)
		if err != nil {
			return nil, fmt.Errorf("%s: encoding %s: %w", r.Op, call.pathName, err)
		}
		arguments[call.pathName] = raw
	}
	return json.Marshal(struct {
		Arguments map[string]json.RawMessage `json:"arguments"`
	}{arguments})
}

// callableResult takes v0's own body out of a result envelope. The envelope's
// `changed` is dropped: a v0 body already says what it did, and nothing above this
// layer reads anything else.
func callableResult(body []byte) ([]byte, error) {
	var envelope struct {
		Result json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, err
	}
	if envelope.Result == nil {
		return nil, errors.New("the answer has no `result`")
	}
	return envelope.Result, nil
}

// mountProblem is the part of the mount's Problem this client reads. V0 and Error
// are v0's own answer, verbatim, when v0 gave one: `error` on a callable-error and
// `v0` on everything else.
type mountProblem struct {
	Code        string          `json:"code"`
	V0          json.RawMessage `json:"v0"`
	Error       json.RawMessage `json:"error"`
	Diagnostics []struct {
		Message          string `json:"message"`
		InstanceLocation string `json:"instanceLocation"`
	} `json:"diagnostics"`
}

// The mount's own codes that say nothing about the operation.
const (
	callableCodeNotFound         = "callable-not-found"
	callableCodeValidationFailed = "validation-failed"
	callableCodeRequestTooLarge  = "request-too-large"
)

// callableProblem maps a refused invocation; call is the Callable's name. It returns either the *ProblemError a
// direct v0 call would have produced or a *CallablesAbsentError, which is not the
// server refusing the operation at all.
//
// When v0 answered, its body and its status are what is mapped, because everything
// above keys on them. The mount's own status is a coarser statement: §5 answers
// v0's `internal` 500 with a 503, and reading that would newly retry a request that
// earned a server fault.
//
// When v0 did not answer, the mount did, about the invocation. The status class is
// the right reading of each of these (a malformed invocation or wrong media type is
// a client bug, ErrBadRequest; a 503 says come back) except the two that mean what
// v0's own invalid_argument means, and those get its sentinel so a caller sees the
// same refusal. An answer that is not a Problem of either kind is something in front
// of bd serve, and reads as it always has.
func callableProblem(t target, call string, res *response, maxRetryAfter time.Duration) error {
	var p mountProblem
	_ = json.Unmarshal(res.body, &p)

	inner := p.V0
	if len(inner) == 0 {
		inner = p.Error
	}
	if len(inner) > 0 {
		var v0 struct {
			Status int `json:"status"`
		}
		status := res.status
		if json.Unmarshal(inner, &v0) == nil && v0.Status != 0 {
			status = v0.Status
		}
		return mapProblem(t, status, res.header, inner, maxRetryAfter)
	}

	// The mount never answers a top-level not_found: that is v0's catch-all, on a
	// server with no /bdp/call for this path to reach.
	if res.status == http.StatusNotFound && p.Code == "not_found" {
		return &CallablesAbsentError{Op: t.op, ServerURL: t.serverURL, Callable: call, NoMount: true}
	}
	if p.Code == callableCodeNotFound {
		return &CallablesAbsentError{Op: t.op, ServerURL: t.serverURL, Callable: call}
	}

	prob := mapProblem(t, res.status, res.header, res.body, maxRetryAfter)
	switch p.Code {
	case callableCodeValidationFailed, callableCodeRequestTooLarge:
		prob.Err = issueops.ErrValidation
		var details []string
		for i, d := range p.Diagnostics {
			details = append(details, stripControlRunes(d.Message))
			if i == 0 {
				prob.Param = stripControlRunes(strings.TrimPrefix(d.InstanceLocation, "/arguments/"))
			}
		}
		if len(details) > 0 {
			prob.Detail = strings.TrimSpace(prob.Detail + ": " + strings.Join(details, "; "))
		}
	}
	if len(bytes.TrimSpace(res.body)) == 0 {
		switch res.status {
		case http.StatusMethodNotAllowed:
			prob.Detail = "this path does not take a POST: the URL may not point at a bd serve"
		case http.StatusNotAcceptable:
			prob.Detail = "the answer this client accepts, " + callableAccept + ", is not on offer"
		}
	}
	return prob
}

// ErrCallablesAbsent reports that a bd serve lacks what this client needs to invoke
// an operation. It is deliberately not issueops.ErrNotFound: the operation's subject
// may exist, and "no such issue" would send the reader looking for the wrong thing.
var ErrCallablesAbsent = errors.New("bd serve does not serve the Callables this client invokes")

// CallablesAbsentError is a server from another release: it has no /bdp/call at all
// (NoMount), or has one without the Callable this client invoked. bd connect talks
// only to the bd serve of its own release, so the recovery is the same for both.
type CallablesAbsentError struct {
	Op        string
	ServerURL string
	Callable  string
	NoMount   bool
}

func (e *CallablesAbsentError) Error() string {
	if e.NoMount {
		return fmt.Sprintf("bd serve at %s predates Callables: it has no %s to invoke %s on. bd connect talks only to the bd serve of its own release, so upgrade that server",
			e.ServerURL, callableMount, e.Callable)
	}
	return fmt.Sprintf("bd serve at %s has no Callable %s, so it is from another release. bd connect talks only to the bd serve of its own release, so upgrade that server",
		e.ServerURL, e.Callable)
}

func (e *CallablesAbsentError) Unwrap() error { return ErrCallablesAbsent }
