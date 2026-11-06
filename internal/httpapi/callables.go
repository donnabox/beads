package httpapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"runtime/debug"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"
)

// The Callables layer: POST <scope>/bdp/call/NAME routes one Callable to its v0
// operation, and GET <scope>/bdp/call says what can be called. It holds no
// business logic. It decodes the invocation envelope, rebuilds the request v0
// would have received, sends it through the same ServeMux a direct v0 request
// goes through, and puts the answer under the envelope the Callables profile
// requires. Auth, the project and wire checks, the database slot, validation
// and every refusal are therefore v0's.
//
// <scope> is the server root. A bd serve process serves one workspace, so it is
// one Scope and needs no path of its own; the Scope URL is http://HOST:PORT/.
const (
	callableMount = "/bdp/call"

	bdpMediaType     = "application/vnd.bdp.v0.1+json"
	problemMediaType = "application/problem+json"
	bdpProblemBase   = "https://github.com/gastownhall/bdp/problems/"

	// callableBodySlack is what the envelope may add around arguments that sit
	// exactly at the v0 operation's own body limit.
	callableBodySlack = 64 << 10

	opCallableInvoke = "callableInvoke"
	opCallableList   = "callableList"

	bdpMalformedRequest  = "malformed-request"
	bdpInvalidParameter  = "invalid-parameter"
	bdpUnauthenticated   = "unauthenticated"
	bdpResourceNotFound  = "resource-not-found"
	bdpRequestTooLarge   = "request-too-large"
	bdpUnavailable       = "temporarily-unavailable"
	bdpUnsupportedMedia  = "unsupported-media-type"
	bdpValidationFailed  = "validation-failed"
	bdpIdentityTaken     = "identity-taken"
	bdpRevisionMismatch  = "revision-mismatch"
	bdpCallableNotFound  = "callable-not-found"
	bdpCallableError     = "callable-error"
	retryNever           = "never"
	retryAfterStateMoved = "after-state-change"
	retryAfterDelay      = "after-delay"
)

// bdpProblemShape is what Core, Write and Callables §9 fix for a code.
type bdpProblemShape struct {
	family string
	status int
	retry  string
}

var bdpProblemShapes = map[string]bdpProblemShape{
	bdpMalformedRequest: {"request", http.StatusBadRequest, retryNever},
	bdpInvalidParameter: {"request", http.StatusBadRequest, retryNever},
	bdpUnauthenticated:  {"authentication", http.StatusUnauthorized, retryAfterStateMoved},
	bdpResourceNotFound: {"not-found", http.StatusNotFound, retryAfterStateMoved},
	bdpRequestTooLarge:  {"size", http.StatusRequestEntityTooLarge, retryNever},
	bdpUnavailable:      {"unavailable", http.StatusServiceUnavailable, retryAfterDelay},
	bdpUnsupportedMedia: {"request", http.StatusUnsupportedMediaType, retryNever},
	bdpValidationFailed: {"validation", http.StatusUnprocessableEntity, retryNever},
	bdpIdentityTaken:    {"conflict", http.StatusConflict, retryNever},
	bdpRevisionMismatch: {"conflict", http.StatusConflict, retryAfterStateMoved},
	bdpCallableNotFound: {"not-found", http.StatusNotFound, retryAfterStateMoved},
	bdpCallableError:    {"conflict", http.StatusConflict, retryAfterStateMoved},
}

// v0ProblemCodes is specs/types/issue/callables.md §5: the v0 codes that meet
// the condition of a Core or Write code. Every other v0 code is a reason of its
// Callable and answers callable-error. precondition_failed is decided per
// response, in bdpCodeFor, because only a stale expected_version is listed.
var v0ProblemCodes = map[Code]string{
	CodeAlreadyExists:      bdpIdentityTaken,
	CodeIdempotencyExpired: bdpIdentityTaken,
	CodeInvalidCursor:      bdpInvalidParameter,
	CodeNotFound:           bdpResourceNotFound,
	CodeInternal:           bdpUnavailable,
	CodeBusy:               bdpUnavailable,
	CodeDBUnavailable:      bdpUnavailable,
	CodeUnauthenticated:    bdpUnauthenticated,
	CodeInvalidArgument:    bdpValidationFailed,
}

// callableProblem is the Problem a mount answers. `v0` is an RFC 9457 extension
// carrying v0's own body on every code but callable-error, whose `error` member
// already carries it, so a translation loses nothing a direct call would show.
type callableProblem struct {
	Type        string               `json:"type"`
	Status      int                  `json:"status"`
	Title       string               `json:"title"`
	Code        string               `json:"code"`
	Retry       string               `json:"retry"`
	Detail      string               `json:"detail,omitempty"`
	Diagnostics []callableDiagnostic `json:"diagnostics,omitempty"`
	Error       json.RawMessage      `json:"error,omitempty"`
	V0          json.RawMessage      `json:"v0,omitempty"`
}

// callableDiagnostic is one entry of Write's `diagnostics`, which
// validation-failed must carry.
type callableDiagnostic struct {
	Message          string `json:"message"`
	InstanceLocation string `json:"instanceLocation,omitempty"`
}

func newCallableProblem(code, detail string) *callableProblem {
	shape := bdpProblemShapes[code]
	return &callableProblem{
		Type:   bdpProblemBase + shape.family,
		Status: shape.status,
		Title:  http.StatusText(shape.status),
		Code:   code,
		Retry:  shape.retry,
		Detail: detail,
	}
}

// invalidArgument is validation-failed with a diagnostic naming the parameter,
// which is every rejection of Callables §4.
func invalidArgument(param, message string) *callableProblem {
	problem := newCallableProblem(bdpValidationFailed, message)
	problem.Diagnostics = []callableDiagnostic{{
		Message:          "`" + param + "`: " + message,
		InstanceLocation: "/arguments/" + param,
	}}
	return problem
}

// callableCapture buffers what the gate or v0 wrote, so that it can be put under
// the envelope once the handler has returned. Every operation on the table
// answers a whole document; the one that streams, watchEvents, is not declared.
type callableCapture struct {
	header http.Header
	status int
	body   bytes.Buffer
	// row is set once the request reached v0: a 200 is then v0's result.
	row *callableRow
	// final means the body is already in its answer shape.
	final bool
}

func (c *callableCapture) Header() http.Header { return c.header }

func (c *callableCapture) WriteHeader(status int) {
	if c.status == 0 {
		c.status = status
	}
}

func (c *callableCapture) Write(b []byte) (int, error) {
	c.WriteHeader(http.StatusOK)
	return c.body.Write(b)
}

// refuse records one of this layer's own Problems.
func (c *callableCapture) refuse(r *http.Request, problem *callableProblem) {
	requestInfo(r.Context()).code = Code(problem.Code)
	c.header.Set("Content-Type", problemMediaType)
	c.final = true
	c.status = problem.Status
	c.body.Reset()
	_ = json.NewEncoder(&c.body).Encode(problem)
}

// mountCallables registers the fixed namespace on the mux v0's routes live on.
// Neither route is in routeTable, which is v0's document: the parity tests hold
// every row there to openapi.v0.yaml.
func (s *Server) mountCallables(mux *http.ServeMux) {
	mux.Handle("GET "+callableMount, s.callableGate(opCallableList, s.serveCallableList))
	mux.Handle(callableMount, methodNotAllowed("GET, HEAD"))
	mux.Handle(http.MethodPost+" "+callableMount+"/{name...}", s.callableGate(opCallableInvoke,
		func(c *callableCapture, r *http.Request) { s.invokeCallable(mux, c, r) }))
	mux.Handle(callableMount+"/{name...}", methodNotAllowed(http.MethodPost))
}

// methodNotAllowed is the bodyless 405 of Callables §3. v0 has no 405; this
// namespace does.
func methodNotAllowed(allow string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Allow", allow)
		w.WriteHeader(http.StatusMethodNotAllowed)
	})
}

// callableGate runs serve behind the same checks a v0 row gets, by being a row:
// s.route does the credential, the project stamp and the wire floor, and the
// database slot is left to the v0 row the request reaches.
//
// A panic in v0 is recovered here rather than by withRequestContext, whose
// answer is v0's `internal` Problem: §5 maps that code, and this keeps one log
// line and one stack trace for it.
func (s *Server) callableGate(op string, serve func(*callableCapture, *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capture := &callableCapture{header: http.Header{}}
		defer func() {
			p := recover()
			if p == nil {
				return
			}
			if p == http.ErrAbortHandler {
				panic(p)
			}
			rec := requestInfo(r.Context())
			s.event("panic", "request_id", rec.id, "op", rec.op, "method", r.Method, "path", r.URL.Path,
				"error", fmt.Sprint(p), "stack", string(debug.Stack()))
			fresh := &callableCapture{header: http.Header{}}
			fresh.refuse(r, newCallableProblem(bdpUnavailable, ""))
			writeCallableResponse(w, r, fresh)
		}()
		s.route(route{
			op:              op,
			bypassSemaphore: true,
			handler: func(_ *Server, _ http.ResponseWriter, r *http.Request) {
				serve(capture, r)
			},
		}).ServeHTTP(capture, r)
		writeCallableResponse(w, r, capture)
	})
}

// invokeCallable is Callables §3: the mount gives the operation, the body gives
// the arguments, and v0 does the rest.
func (s *Server) invokeCallable(mux http.Handler, c *callableCapture, r *http.Request) {
	row, ok := callableFromPath(r)
	if !ok {
		c.refuse(r, newCallableProblem(bdpCallableNotFound, "no reachable Callable has this name"))
		return
	}
	if media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || media != bdpMediaType {
		c.refuse(r, newCallableProblem(bdpUnsupportedMedia, "an invocation's Content-Type is "+bdpMediaType))
		return
	}
	if !acceptsMedia(r.Header.Values("Accept"), bdpMediaType) {
		c.final = true
		c.status = http.StatusNotAcceptable
		return
	}
	if !row.batch && r.Header.Get("Idempotency-Key") != "" {
		c.refuse(r, newCallableProblem(bdpMalformedRequest, "a CallableQuery invocation carries no Idempotency-Key"))
		return
	}

	args, provenance, problem := decodeInvocation(c, r, row)
	if problem != nil {
		c.refuse(r, problem)
		return
	}
	actor, problem := checkProvenance(row, args, provenance)
	if problem != nil {
		c.refuse(r, problem)
		return
	}
	if problem := row.checkArgumentNames(args); problem != nil {
		c.refuse(r, problem)
		return
	}
	if row.adapt != nil {
		if problem := row.adapt(args, actor); problem != nil {
			c.refuse(r, problem)
			return
		}
	}
	inner, problem := row.v0Request(r, args)
	if problem != nil {
		c.refuse(r, problem)
		return
	}

	c.row = row
	mux.ServeHTTP(c, inner)
}

// callableFromPath finds the row a mount names. An encoded separator names
// nothing: Core refuses one in a local spelling, and a decoded lookup would make
// issue%2Fshow a second spelling of issue/show.
func callableFromPath(r *http.Request) (*callableRow, bool) {
	escaped := strings.ToLower(r.URL.EscapedPath())
	if strings.Contains(escaped, "%2f") || strings.Contains(escaped, "%5c") {
		return nil, false
	}
	row, ok := callableIndex[r.PathValue("name")]
	return row, ok
}

// acceptsMedia is Core's Accept rule for an endpoint that offers one media type:
// missing permits it, and the most specific matching range decides, q=0 included.
func acceptsMedia(values []string, media string) bool {
	if strings.TrimSpace(strings.Join(values, "")) == "" {
		return true
	}
	mainType, _, _ := strings.Cut(media, "/")
	best, accepted := -1, false
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			mediaRange, params, err := mime.ParseMediaType(strings.TrimSpace(part))
			if err != nil {
				continue
			}
			specificity := -1
			switch mediaRange {
			case media:
				specificity = 2
			case mainType + "/*":
				specificity = 1
			case "*/*":
				specificity = 0
			}
			if specificity > best {
				q, err := strconv.ParseFloat(params["q"], 64)
				best, accepted = specificity, params["q"] == "" || (err == nil && q > 0)
			}
		}
	}
	return accepted
}

// decodeInvocation is the closed callableInvocation of Callables §3. Unknown
// members, duplicate members and invalid strings are malformed-request, at every
// depth, because v0's decoders would keep the last duplicate without a word.
func decodeInvocation(c *callableCapture, r *http.Request, row *callableRow) (args, provenance map[string]json.RawMessage, problem *callableProblem) {
	limit := row.bodyLimit() + callableBodySlack
	body, err := io.ReadAll(http.MaxBytesReader(c, r.Body, limit))
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			return nil, nil, newCallableProblem(bdpRequestTooLarge, fmt.Sprintf("the invocation is larger than %d bytes", limit))
		}
		return nil, nil, newCallableProblem(bdpMalformedRequest, "the invocation body could not be read")
	}
	if detail := invalidJSONCarrier(body); detail != "" {
		return nil, nil, newCallableProblem(bdpMalformedRequest, detail)
	}

	var members map[string]json.RawMessage
	if err := json.Unmarshal(body, &members); err != nil || members == nil {
		return nil, nil, newCallableProblem(bdpMalformedRequest, "the invocation is a JSON object")
	}
	for name := range members {
		if name != "arguments" && name != "revisionProvenance" {
			return nil, nil, newCallableProblem(bdpMalformedRequest, "`"+name+"` is not an invocation member")
		}
	}
	if err := json.Unmarshal(members["arguments"], &args); err != nil || args == nil {
		return nil, nil, newCallableProblem(bdpMalformedRequest, "`arguments` is required and is an object")
	}
	raw, present := members["revisionProvenance"]
	if !present {
		return args, nil, nil
	}
	if !row.batch {
		return nil, nil, newCallableProblem(bdpMalformedRequest, "`revisionProvenance` is rejected on a CallableQuery")
	}
	if err := json.Unmarshal(raw, &provenance); err != nil || provenance == nil {
		return nil, nil, newCallableProblem(bdpMalformedRequest, "`revisionProvenance` is an object")
	}
	return args, provenance, nil
}

// invalidJSONCarrier walks the whole document for what encoding/json lets
// through: a duplicate member, invalid UTF-8 and an unpaired surrogate escape.
// It returns what is wrong, or "".
func invalidJSONCarrier(body []byte) string {
	if !utf8.Valid(body) {
		return "the invocation is not valid UTF-8"
	}
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	// One entry per open container; nil marks an array, whose strings are values.
	type frame struct {
		names     map[string]bool
		wantValue bool
	}
	var stack []*frame
	valueDone := func() {
		if n := len(stack); n > 0 && stack[n-1].names != nil {
			stack[n-1].wantValue = false
		}
	}
	for {
		token, err := dec.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "the invocation is not valid JSON"
		}
		switch t := token.(type) {
		case json.Delim:
			switch t {
			case '{':
				stack = append(stack, &frame{names: map[string]bool{}})
			case '[':
				stack = append(stack, &frame{})
			default:
				stack = stack[:len(stack)-1]
				valueDone()
			}
		case string:
			if n := len(stack); n > 0 && stack[n-1].names != nil && !stack[n-1].wantValue {
				if stack[n-1].names[t] {
					return "the invocation repeats the member `" + t + "`"
				}
				stack[n-1].names[t] = true
				stack[n-1].wantValue = true
				continue
			}
			valueDone()
		default:
			valueDone()
		}
	}
	if hasUnpairedSurrogate(body) {
		return "the invocation carries an unpaired surrogate escape"
	}
	return ""
}

// hasUnpairedSurrogate scans valid JSON, where a backslash occurs only inside a
// string and always starts an escape.
func hasUnpairedSurrogate(body []byte) bool {
	unit := func(i int) (rune, bool) {
		if i+6 > len(body) || body[i] != '\\' || body[i+1] != 'u' {
			return 0, false
		}
		v, err := strconv.ParseUint(string(body[i+2:i+6]), 16, 16)
		return rune(v), err == nil
	}
	for i := 0; i < len(body); i++ {
		if body[i] != '\\' {
			continue
		}
		u, ok := unit(i)
		switch {
		case !ok:
			i++
		case u >= 0xDC00 && u <= 0xDFFF:
			return true
		case u >= 0xD800 && u <= 0xDBFF:
			low, ok := unit(i + 6)
			if !ok || low < 0xDC00 || low > 0xDFFF {
				return true
			}
			i += 11
		default:
			i += 5
		}
	}
	return false
}

// checkProvenance validates Write's closed revisionProvenanceInput and answers
// the invocation's actor.
//
// V0 HAS NOWHERE TO CARRY IT. Callables §3 says the input is "carried onto every
// revision the batch commits", and v0 records only its own `actor` argument. So
// the rule here is honored or refused, never silently dropped: an actor must
// agree with the `actor` argument that v0 will record, and an agent or message,
// which v0 cannot store at all, is refused. issue/comment is the exception by
// §3 of the Issue Callables: it uses the actor to fill `author`.
func checkProvenance(row *callableRow, args, provenance map[string]json.RawMessage) (string, *callableProblem) {
	if provenance == nil {
		return "", nil
	}
	malformed := func(detail string) (string, *callableProblem) {
		return "", newCallableProblem(bdpMalformedRequest, "`revisionProvenance`: "+detail)
	}
	for name := range provenance {
		if name != "actor" && name != "agent" && name != "message" {
			return malformed("`" + name + "` is not a member")
		}
	}
	actor := ""
	if raw, ok := provenance["actor"]; ok {
		var members map[string]json.RawMessage
		if err := json.Unmarshal(raw, &members); err != nil || members == nil {
			return malformed("`actor` is an object")
		}
		var basis string
		if len(members) != 2 || json.Unmarshal(members["id"], &actor) != nil || actor == "" ||
			json.Unmarshal(members["basis"], &basis) != nil || (basis != "writer-supplied" && basis != "unknown") {
			return malformed("`actor` is exactly a nonempty `id` and a `basis` of writer-supplied or unknown")
		}
	}
	for _, member := range []string{"agent", "message"} {
		raw, ok := provenance[member]
		if !ok || string(raw) == "null" {
			continue
		}
		var value string
		if json.Unmarshal(raw, &value) != nil || (member == "agent" && value == "") {
			return malformed("`" + member + "` is not a valid value")
		}
		return "", invalidArgument("revisionProvenance",
			"`"+member+"` cannot be recorded: this server's operations carry no such attribution")
	}
	if actor == "" || row.op == OpAddComment {
		return actor, nil
	}
	var named string
	if json.Unmarshal(args["actor"], &named) != nil || named != actor {
		return "", invalidArgument("revisionProvenance",
			"`actor` must equal the `actor` argument, the only attribution this operation records")
	}
	return actor, nil
}

// checkArgumentNames is the first rejection of Callables §4; every unknown name
// gets a diagnostic. Required parameters and value schemas are v0's to check, and
// its invalid_argument becomes the same validation-failed.
func (row *callableRow) checkArgumentNames(args map[string]json.RawMessage) *callableProblem {
	var unknown []string
	for name := range args {
		if !slices.Contains(row.path, name) && !slices.Contains(row.query, name) &&
			!slices.Contains(row.body, name) && !slices.Contains(row.adapterParams, name) {
			unknown = append(unknown, name)
		}
	}
	if len(unknown) == 0 {
		return nil
	}
	slices.Sort(unknown)
	problem := newCallableProblem(bdpValidationFailed, "these arguments are not declared parameters of "+row.name)
	for _, name := range unknown {
		problem.Diagnostics = append(problem.Diagnostics, callableDiagnostic{
			Message:          "`" + name + "` is not a declared parameter",
			InstanceLocation: "/arguments/" + name,
		})
	}
	return problem
}

// adaptFlatDependency is the issue/dep/add adapter: flat source, target and
// optional type, or v0's edges, exactly one form per call. An omitted type is
// blocks, where v0 itself requires one on every edge.
func adaptFlatDependency(args map[string]json.RawMessage, _ string) *callableProblem {
	_, hasEdges := args["edges"]
	flat := ""
	for _, name := range []string{"source", "target", "type"} {
		if _, ok := args[name]; ok {
			flat = name
			break
		}
	}
	switch {
	case hasEdges && flat != "":
		return invalidArgument(flat, "`edges` and the flat form cannot be combined; send one form")
	case !hasEdges && flat == "":
		return invalidArgument("edges", "send `edges`, or `source` and `target`")
	case hasEdges:
		return nil
	}
	edge := map[string]json.RawMessage{"type": json.RawMessage(`"blocks"`)}
	for _, member := range []struct{ arg, wire string }{
		{"source", "issue_id"}, {"target", "depends_on_id"}, {"type", "type"},
	} {
		raw, ok := args[member.arg]
		if !ok {
			if member.arg == "type" {
				continue
			}
			return invalidArgument(member.arg, "the flat form requires `source` and `target`")
		}
		var value string
		if json.Unmarshal(raw, &value) != nil || string(raw) == "null" {
			return invalidArgument(member.arg, "must be a string")
		}
		edge[member.wire] = raw
		delete(args, member.arg)
	}
	edges, err := json.Marshal([]map[string]json.RawMessage{edge})
	if err != nil {
		return invalidArgument("edges", "the flat form could not be encoded")
	}
	args["edges"] = edges
	return nil
}

// adaptCommentAuthor is the issue/comment adapter: an omitted `author` is the
// invocation's actor, revisionProvenance.actor.id. With no actor either, nothing
// is filled and v0 refuses the missing `author`.
func adaptCommentAuthor(args map[string]json.RawMessage, actor string) *callableProblem {
	if _, supplied := args["author"]; supplied || actor == "" {
		return nil
	}
	author, err := json.Marshal(actor)
	if err != nil {
		return invalidArgument("author", "the actor could not be encoded")
	}
	args["author"] = author
	return nil
}

// bodyLimit is the v0 operation's largest body class, so that an oversized
// invocation is refused here, before it is held in memory outside a database slot.
func (row *callableRow) bodyLimit() int64 {
	switch row.op {
	case OpApplyBatch:
		return maxApplyBatchBodyBytes
	case OpBatchCreateIssues, OpBatchCloseIssues, OpCreateIssue, OpUpdateIssue:
		return maxBatchCreateBodyBytes
	}
	return maxJSONBodyBytes
}

// v0Request flattens the arguments back into the request a direct v0 caller
// would have sent. The row's route gives the method and path, so the two cannot
// disagree.
func (row *callableRow) v0Request(r *http.Request, args map[string]json.RawMessage) (*http.Request, *callableProblem) {
	rt, ok := routeForOp(row.op)
	if !ok {
		return nil, newCallableProblem(bdpUnavailable, "")
	}
	target := rt.specPathOf()
	for _, name := range row.path {
		raw, ok := args[name]
		if !ok {
			return nil, invalidArgument(name, "is required")
		}
		var value string
		// `.` and `..` are excluded because ServeMux would clean them into a
		// redirect to another route.
		if json.Unmarshal(raw, &value) != nil || string(raw) == "null" || value == "" || value == "." || value == ".." {
			return nil, invalidArgument(name, "must be a nonempty string")
		}
		target = strings.Replace(target, "{"+name+"}", url.PathEscape(value), 1)
	}

	query := url.Values{}
	for _, name := range row.query {
		raw, ok := args[name]
		if !ok {
			continue
		}
		values, valid := queryValues(raw)
		if !valid {
			return nil, invalidArgument(name, "must be a string, number, boolean or array of those")
		}
		query[name] = values
	}
	if encoded := query.Encode(); encoded != "" {
		target += "?" + encoded
	}

	var body io.Reader
	if len(row.body) > 0 {
		members := make(map[string]json.RawMessage, len(row.body))
		for _, name := range row.body {
			if raw, ok := args[name]; ok {
				members[name] = raw
			}
		}
		encoded, err := json.Marshal(members)
		if err != nil {
			return nil, newCallableProblem(bdpMalformedRequest, "the arguments could not be encoded")
		}
		body = bytes.NewReader(encoded)
	}

	inner, err := http.NewRequestWithContext(r.Context(), rt.method, target, body)
	if err != nil {
		return nil, newCallableProblem(bdpMalformedRequest, "the arguments do not make a valid request")
	}
	if body == nil {
		inner.Body = http.NoBody // A server request's Body is never nil.
	}
	inner.Host = r.Host
	inner.RemoteAddr = r.RemoteAddr
	inner.Header = r.Header.Clone()
	inner.Header.Del("Accept")
	inner.Header.Del("Idempotency-Key")
	inner.Header.Del("Content-Type")
	if body != nil {
		inner.Header.Set("Content-Type", claimContentType)
	}
	return inner, nil
}

// routeForOp finds a v0 row by operation.
func routeForOp(op string) (route, bool) {
	for _, rt := range routeTable {
		if rt.op == op {
			return rt, true
		}
	}
	return route{}, false
}

// queryValues spells one JSON argument as query values. Numbers keep their
// literal digits. An empty array is no parameter at all.
func queryValues(raw json.RawMessage) ([]string, bool) {
	var items []json.RawMessage
	if json.Unmarshal(raw, &items) == nil && items != nil {
		values := make([]string, 0, len(items))
		for _, item := range items {
			value, ok := queryScalar(item)
			if !ok {
				return nil, false
			}
			values = append(values, value)
		}
		return values, true
	}
	value, ok := queryScalar(raw)
	if !ok {
		return nil, false
	}
	return []string{value}, true
}

func queryScalar(raw json.RawMessage) (string, bool) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return "", false
	}
	switch trimmed[0] {
	case '"':
		var value string
		return value, json.Unmarshal(trimmed, &value) == nil
	case '{', '[', 'n':
		return "", false
	}
	return string(trimmed), true
}

// writeCallableResponse puts what was captured under the mount's answer shape.
func writeCallableResponse(w http.ResponseWriter, r *http.Request, c *callableCapture) {
	h := w.Header()
	h.Set("Cache-Control", "private, no-store")
	for _, name := range []string{"Retry-After", "WWW-Authenticate", "Content-Type"} {
		if value := c.header.Get(name); value != "" {
			h.Set(name, value)
		}
	}
	status := c.status
	if status == 0 {
		status = http.StatusOK
	}
	switch {
	case c.final:
		w.WriteHeader(status)
		_, _ = w.Write(c.body.Bytes())
	case status == http.StatusOK && c.row != nil:
		h.Set("Content-Type", bdpMediaType)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(callableResultBody(c.row, c.body.Bytes()))
	default:
		problem := translateV0Problem(c.body.Bytes())
		requestInfo(r.Context()).code = Code(problem.Code)
		h.Set("Content-Type", problemMediaType)
		w.WriteHeader(problem.Status)
		_ = json.NewEncoder(w).Encode(problem)
	}
}

// callableResultBody is the closed callableResult of Callables §5, with v0's
// bytes as `result`.
func callableResultBody(row *callableRow, result []byte) []byte {
	result = bytes.TrimSpace(result)
	var out bytes.Buffer
	out.WriteByte('{')
	if row.batch {
		changed := true
		if row.changed != nil {
			var members map[string]json.RawMessage
			if json.Unmarshal(result, &members) == nil {
				changed = row.changed(members)
			}
		}
		fmt.Fprintf(&out, `"changed":%t,`, changed)
	}
	out.WriteString(`"result":`)
	out.Write(result)
	out.WriteString("}\n")
	return out.Bytes()
}

// translateV0Problem is specs/types/issue/callables.md §5.
func translateV0Problem(body []byte) *callableProblem {
	var v0 struct {
		Code            string  `json:"code"`
		Detail          *string `json:"detail"`
		Param           *string `json:"param"`
		ExpectedVersion *string `json:"expected_version"`
	}
	raw := json.RawMessage(bytes.TrimSpace(body))
	if json.Unmarshal(raw, &v0) != nil || v0.Code == "" {
		// Unreachable for a v0 route: every non-2xx byte is a Problem.
		return newCallableProblem(bdpUnavailable, "")
	}
	code := bdpCodeFor(Code(v0.Code), v0.ExpectedVersion != nil || (v0.Param != nil && *v0.Param == expectedVersionMember))
	if code == bdpCallableError {
		problem := newCallableProblem(code, "")
		problem.Error = raw
		return problem
	}
	detail := ""
	if v0.Detail != nil {
		detail = *v0.Detail
	}
	problem := newCallableProblem(code, detail)
	problem.V0 = raw
	if code == bdpValidationFailed {
		problem.Diagnostics = []callableDiagnostic{v0Diagnostic(v0.Param, detail)}
	}
	return problem
}

// bdpCodeFor is the §5 table: staleVersion says a precondition_failed was the
// expected_version guard, which is the only one that table lists.
func bdpCodeFor(code Code, staleVersion bool) string {
	if code == CodePreconditionFailed && staleVersion {
		return bdpRevisionMismatch
	}
	if mapped, ok := v0ProblemCodes[code]; ok {
		return mapped
	}
	return bdpCallableError
}

// v0Diagnostic names the parameter v0 refused. v0's `param` can be a nested
// spelling such as items[0].title or a header, so only a plain name becomes a
// JSON Pointer into the arguments.
func v0Diagnostic(param *string, detail string) callableDiagnostic {
	if detail == "" {
		detail = "the value was refused"
	}
	if param == nil {
		return callableDiagnostic{Message: detail}
	}
	diagnostic := callableDiagnostic{Message: "`" + *param + "`: " + detail}
	if !strings.ContainsAny(*param, ".[-") {
		diagnostic.InstanceLocation = "/arguments/" + *param
	}
	return diagnostic
}
