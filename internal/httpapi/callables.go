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
// business logic. An invocation runs under its v0 row's own route: the credential,
// the project and wire checks, and the database slot are that row's, once, with
// the row's own exemptions. Inside the slot it decodes the envelope, rebuilds the
// request v0 would have received, hands it to the row's handler, and puts the
// answer under the envelope the Callables profile requires. Validation and every
// refusal are therefore v0's. What this layer adds is what the document gives it:
// revisionProvenance.actor.id becomes v0's `actor` or `author`, and an omitted
// parameter with a service default gets that default, on this surface alone.
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
	bdpIdentityTaken:    {"conflict", http.StatusConflict, retryAfterStateMoved},
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
// `reason` is v0's discriminator, on the one refusal this layer makes in v0's place
// and that a client tells apart by it: an undeclared argument.
type callableProblem struct {
	Type        string               `json:"type"`
	Status      int                  `json:"status"`
	Title       string               `json:"title"`
	Code        string               `json:"code"`
	Retry       string               `json:"retry"`
	Detail      string               `json:"detail,omitempty"`
	Reason      Reason               `json:"reason,omitempty"`
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
	// row is set when the name resolved to a Callable: a 200 is then v0's result.
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
	mux.Handle("GET "+callableMount, s.callableGate(func(c *callableCapture, _ *http.Request) route {
		return route{op: opCallableList, bypassSemaphore: true, handler: func(s *Server, _ http.ResponseWriter, r *http.Request) {
			s.serveCallableList(c, r)
		}}
	}))
	mux.Handle(callableMount, methodNotAllowed("GET, HEAD"))
	mux.Handle(http.MethodPost+" "+callableMount+"/{name...}", s.callableGate(invocationRoute))
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

// callableGate runs one request behind the route pick chooses for it and writes
// what that route captured under the mount's answer shape, after the route has
// given its database slot back: a slow reader of the answer holds no slot.
//
// A panic in v0 is recovered here rather than by withRequestContext, whose answer
// would be v0's bare Problem outside the Callables shape. What it answers is what
// a direct v0 call would have said, `internal` with its request id, under §5's
// translation of that code, so that id is not lost and one log line and one stack
// trace stay the whole record.
func (s *Server) callableGate(pick func(*callableCapture, *http.Request) route) http.Handler {
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
			problem := newCallableProblem(bdpUnavailable, "")
			if raw, err := json.Marshal(newResult(CodeInternal, "").WithRequestID(rec.id).Problem); err == nil {
				problem = translateV0Problem(raw)
			}
			fresh.refuse(r, problem)
			writeCallableResponse(w, r, fresh)
		}()
		s.route(pick(capture, r)).ServeHTTP(capture, r)
		writeCallableResponse(w, r, capture)
	})
}

// invocationRoute is the route one invocation runs under: Callables §3, on the
// lifecycle of the v0 row the name routes to. That row's exemptions are the
// Callable's by construction, so issue/context answers a wrong project stamp as
// v0's handshake does, which is how a client learns the id and the rich mismatch
// error it reports.
//
// The media checks run before the slot, because they cost nothing and read
// nothing. The body is read inside the slot, as v0 reads its own.
func invocationRoute(c *callableCapture, r *http.Request) route {
	refusal := func(problem *callableProblem) route {
		return route{op: opCallableInvoke, bypassSemaphore: true, handler: func(_ *Server, _ http.ResponseWriter, r *http.Request) {
			c.refuse(r, problem)
		}}
	}
	row, ok := callableFromPath(r)
	if !ok {
		return refusal(newCallableProblem(bdpCallableNotFound, "no reachable Callable has this name"))
	}
	rt, ok := routeForOp(row.op)
	if !ok {
		return refusal(newCallableProblem(bdpUnavailable, ""))
	}
	c.row = row
	v0Handler := rt.handler
	rt.admit = func(_ http.ResponseWriter, r *http.Request) bool {
		return row.admitInvocation(c, r)
	}
	rt.handler = func(s *Server, _ http.ResponseWriter, r *http.Request) {
		inner, problem := row.invocationRequest(s, c, r)
		if problem != nil {
			c.refuse(r, problem)
			return
		}
		// A custom method's id gets the bound its shared wildcard gets on a direct
		// call (customMethodTarget), under the same answer.
		if rt.customMethod != "" {
			_, id, res := customMethodTarget([]route{rt}, inner.PathValue(customMethodIDValue)+rt.customMethod)
			if res != nil {
				s.fail(c, inner, *res)
				return
			}
			inner.SetPathValue(customMethodIDValue, id)
		}
		v0Handler(s, c, inner)
	}
	return rt
}

// admitInvocation is what an invocation is refused for by its headers alone. It
// reports whether the invocation goes on.
func (row *callableRow) admitInvocation(c *callableCapture, r *http.Request) bool {
	if media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || media != bdpMediaType {
		c.refuse(r, newCallableProblem(bdpUnsupportedMedia, "an invocation's Content-Type is "+bdpMediaType))
		return false
	}
	if !acceptsMedia(r.Header.Values("Accept"), bdpMediaType) {
		c.final = true
		c.status = http.StatusNotAcceptable
		return false
	}
	if !row.batch && r.Header.Get("Idempotency-Key") != "" {
		c.refuse(r, newCallableProblem(bdpMalformedRequest, "a CallableQuery invocation carries no Idempotency-Key"))
		return false
	}
	return true
}

// invocationRequest decodes and checks the envelope, and answers the request v0
// would have received.
func (row *callableRow) invocationRequest(s *Server, c *callableCapture, r *http.Request) (*http.Request, *callableProblem) {
	inv, problem := decodeInvocation(c, r, row)
	if problem != nil {
		return nil, problem
	}
	actor, problem := checkProvenance(row, inv.provenance)
	if problem != nil {
		return nil, problem
	}
	if problem := row.checkArgumentNames(inv.args); problem != nil {
		return nil, problem
	}
	if row.adapt != nil {
		if problem := row.adapt(inv.args); problem != nil {
			return nil, problem
		}
	}
	if problem := row.supplyActor(inv, actor); problem != nil {
		return nil, problem
	}
	row.supplyDefaults(inv)
	if problem := row.settleCoreGuard(s, r, inv); problem != nil {
		return nil, problem
	}
	return row.v0Request(r, inv)
}

// settleCoreGuard is the routing rule's revision tokens: `expected_version` also
// takes a token from a generic Core record. v0 knows only the row lock, so this is
// done here, before the request v0 would have received is built, and v0's own
// routes go on refusing such a token. A bare row lock, or anything else that is no
// Core token, is left for v0 to read.
//
// A Core token is compared whole against the record as it stands, which is what
// catches a label-only write, and v0 is then sent its row lock half, which v0 checks
// inside the write's own transaction. The comparison is outside that transaction: a
// write that moves only the digest can land between the two. A record that cannot be
// named here, or does not exist, goes on to v0, whose refusal is the answer.
func (row *callableRow) settleCoreGuard(s *Server, r *http.Request, inv *invocation) *callableProblem {
	raw, sent := inv.args[expectedVersionMember]
	if !sent || !slices.Contains(row.body, expectedVersionMember) {
		return nil
	}
	var token string
	if json.Unmarshal(raw, &token) != nil {
		return nil
	}
	rowLock, isCore := coreRevisionRowLock(token)
	if !isCore {
		return nil
	}
	if id, named := row.guardedID(inv.args); named {
		p := newCoreProjector(s, r)
		found, err := p.lookup(r.Context(), id)
		if err != nil {
			return newCallableProblem(bdpUnavailable, "the record's revision could not be read")
		}
		if found != nil && found.bead != nil {
			record, err := p.bead(r.Context(), found.bead)
			if err != nil {
				return newCallableProblem(bdpUnavailable, "the record's revision could not be read")
			}
			if record.Revision != token {
				return newCallableProblem(bdpRevisionMismatch, "`"+expectedVersionMember+"` is a revision this record no longer carries")
			}
		}
	}
	replaced, err := json.Marshal(rowLock)
	if err != nil {
		return newCallableProblem(bdpUnavailable, "")
	}
	inv.args[expectedVersionMember] = replaced
	inv.rebuild = true
	return nil
}

// guardedID is the one record a row's guard is about: its path `id`, or the single
// distinct id of `ids`. More than one is v0's to refuse.
func (row *callableRow) guardedID(args map[string]json.RawMessage) (string, bool) {
	if slices.Contains(row.path, "id") {
		var id string
		return id, json.Unmarshal(args["id"], &id) == nil && id != ""
	}
	var ids []string
	if json.Unmarshal(args["ids"], &ids) != nil || len(ids) == 0 || distinctDeleteIDs(ids) != 1 {
		return "", false
	}
	return strings.TrimSpace(ids[0]), true
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

// invocation is one decoded envelope. Its members are subslices of the one buffer
// the body was read into: nothing here copies it, because an issue/batch body is
// 16 MiB and every copy held in a slot is a copy maxInflight times over.
type invocation struct {
	// arguments is the `arguments` member's own bytes, and args its members.
	arguments  json.RawMessage
	args       map[string]json.RawMessage
	provenance map[string]json.RawMessage
	// supplied names what this layer added to args beyond the caller's `arguments`:
	// the actor and any defaults.
	supplied []string
	// rebuild says a member of args differs from `arguments`, so v0's body cannot be
	// `arguments` itself.
	rebuild bool
}

// supply adds an argument the caller did not send.
func (inv *invocation) supply(name string, value json.RawMessage) {
	inv.args[name] = value
	inv.supplied = append(inv.supplied, name)
}

// decodeInvocation is the closed callableInvocation of Callables §3. Unknown
// members, duplicate members and invalid strings are malformed-request, at every
// depth, because v0's decoders would keep the last duplicate without a word.
func decodeInvocation(c *callableCapture, r *http.Request, row *callableRow) (*invocation, *callableProblem) {
	limit := row.bodyLimit() + callableBodySlack
	// Sized up front when the client says, so that reading a large body does not
	// hold both the old and the new buffer each time one fills.
	buffer := bytes.NewBuffer(make([]byte, 0, min(max(r.ContentLength, 0), limit)+bytes.MinRead))
	_, err := buffer.ReadFrom(http.MaxBytesReader(c, r.Body, limit))
	body := buffer.Bytes()
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			return nil, newCallableProblem(bdpRequestTooLarge, fmt.Sprintf("the invocation is larger than %d bytes", limit))
		}
		return nil, newCallableProblem(bdpMalformedRequest, "the invocation body could not be read")
	}
	members, args, detail := scanInvocation(body)
	if detail != "" {
		return nil, newCallableProblem(bdpMalformedRequest, detail)
	}
	if members == nil {
		return nil, newCallableProblem(bdpMalformedRequest, "the invocation is a JSON object")
	}
	for name := range members {
		if name != "arguments" && name != "revisionProvenance" {
			return nil, newCallableProblem(bdpMalformedRequest, "`"+name+"` is not an invocation member")
		}
	}
	if args == nil {
		return nil, newCallableProblem(bdpMalformedRequest, "`arguments` is required and is an object")
	}
	inv := &invocation{arguments: members["arguments"], args: args}
	raw, present := members["revisionProvenance"]
	if !present {
		return inv, nil
	}
	if !row.batch {
		return nil, newCallableProblem(bdpMalformedRequest, "`revisionProvenance` is rejected on a CallableQuery")
	}
	if err := json.Unmarshal(raw, &inv.provenance); err != nil || inv.provenance == nil {
		return nil, newCallableProblem(bdpMalformedRequest, "`revisionProvenance` is an object")
	}
	return inv, nil
}

// scanInvocation walks the whole document for what encoding/json lets through: a
// duplicate member, invalid UTF-8 and an unpaired surrogate escape. It answers
// what is wrong in detail, or else the document's members and the members of its
// `arguments`, as subslices of body. Either map is nil when the document, or its
// `arguments`, is not an object.
func scanInvocation(body []byte) (members, args map[string]json.RawMessage, detail string) {
	if !utf8.Valid(body) {
		return nil, nil, "the invocation is not valid UTF-8"
	}
	if !json.Valid(body) {
		return nil, nil, "the invocation is not valid JSON"
	}
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	// One entry per open container; names is nil on an array, whose strings are
	// values. into is where an object's member values are kept, when it is one of
	// the two objects anyone reads.
	type frame struct {
		names     map[string]bool
		wantValue bool
		name      string
		start     int64
		into      map[string]json.RawMessage
	}
	var stack []*frame
	valueDone := func() {
		n := len(stack)
		if n == 0 || stack[n-1].names == nil {
			return
		}
		top := stack[n-1]
		top.wantValue = false
		if top.into != nil {
			top.into[top.name] = body[top.start:dec.InputOffset():dec.InputOffset()]
		}
	}
	// valueStart is where the value after a member name begins.
	valueStart := func() int64 {
		i := dec.InputOffset()
		for i < int64(len(body)) && strings.IndexByte(" \t\r\n:", body[i]) >= 0 {
			i++
		}
		return i
	}
	for {
		token, err := dec.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, nil, "the invocation is not valid JSON"
		}
		switch t := token.(type) {
		case json.Delim:
			switch t {
			case '{':
				opened := &frame{names: map[string]bool{}}
				switch n := len(stack); {
				case n == 0:
					members = map[string]json.RawMessage{}
					opened.into = members
				case n == 1 && stack[0].name == "arguments":
					args = map[string]json.RawMessage{}
					opened.into = args
				}
				stack = append(stack, opened)
			case '[':
				stack = append(stack, &frame{})
			default:
				stack = stack[:len(stack)-1]
				valueDone()
			}
		case string:
			if n := len(stack); n > 0 && stack[n-1].names != nil && !stack[n-1].wantValue {
				top := stack[n-1]
				if top.names[t] {
					return nil, nil, "the invocation repeats the member `" + t + "`"
				}
				top.names[t] = true
				top.wantValue = true
				top.name = t
				top.start = valueStart()
				continue
			}
			valueDone()
		default:
			valueDone()
		}
	}
	if hasUnpairedSurrogate(body) {
		return nil, nil, "the invocation carries an unpaired surrogate escape"
	}
	return members, args, ""
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
// the `id` of the invocation's actor, as sent, or nil if it names none.
//
// V0 HAS NOWHERE TO CARRY THE REST. Callables §3 says the input is "carried onto
// every revision the batch commits", and v0 records only the name it is given as
// `actor` or `author`. So the rule here is honored or refused, never silently
// dropped: an agent or message, which v0 cannot store at all, is refused, and so is
// an actor on the four operations of §3 that record none.
func checkProvenance(row *callableRow, provenance map[string]json.RawMessage) (json.RawMessage, *callableProblem) {
	if provenance == nil {
		return nil, nil
	}
	malformed := func(detail string) (json.RawMessage, *callableProblem) {
		return nil, newCallableProblem(bdpMalformedRequest, "`revisionProvenance`: "+detail)
	}
	for name := range provenance {
		if name != "actor" && name != "agent" && name != "message" {
			return malformed("`" + name + "` is not a member")
		}
	}
	var actor json.RawMessage
	if raw, ok := provenance["actor"]; ok {
		var members map[string]json.RawMessage
		if err := json.Unmarshal(raw, &members); err != nil || members == nil {
			return malformed("`actor` is an object")
		}
		var id, basis string
		if len(members) != 2 || json.Unmarshal(members["id"], &id) != nil || id == "" ||
			json.Unmarshal(members["basis"], &basis) != nil || (basis != "writer-supplied" && basis != "unknown") {
			return malformed("`actor` is exactly a nonempty `id` and a `basis` of writer-supplied or unknown")
		}
		actor = members["id"]
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
		return nil, invalidArgument("revisionProvenance",
			"`"+member+"` cannot be recorded: this server's operations carry no such attribution")
	}
	if actor != nil && row.actorMember == "" {
		return nil, invalidArgument("revisionProvenance",
			"`actor` cannot be recorded: this operation records no actor")
	}
	return actor, nil
}

// supplyActor is the actor row of §3: revisionProvenance.actor.id goes into v0's
// body as the row's member, where v0's own validation of a name is the check, and
// nothing about that check is restated here. With no actor where v0 requires one
// the refusal is this layer's, because v0's would name a member this caller cannot
// send.
func (row *callableRow) supplyActor(inv *invocation, actor json.RawMessage) *callableProblem {
	switch {
	case row.actorMember == "" || (actor == nil && row.actorOptional):
		return nil
	case actor != nil:
		inv.supply(row.actorMember, actor)
		return nil
	}
	message := "`revisionProvenance.actor` is required: " + row.name + " has to say who is acting, and its `id` is what v0 records as `" + row.actorMember + "`"
	problem := newCallableProblem(bdpValidationFailed, message)
	problem.Diagnostics = []callableDiagnostic{{Message: message, InstanceLocation: "/revisionProvenance/actor"}}
	return problem
}

// supplyDefaults is §4: an omitted parameter with a declared default is supplied
// that value, and anything sent, null included, is the caller's.
func (row *callableRow) supplyDefaults(inv *invocation) {
	for name, value := range row.defaults {
		if _, sent := inv.args[name]; !sent {
			inv.supply(name, value)
		}
	}
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
	problem.Reason = ReasonUnknownParameter
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
func adaptFlatDependency(args map[string]json.RawMessage) *callableProblem {
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
		if n := utf8.RuneCountInString(value); n < 1 || n > 255 {
			return invalidArgument(member.arg, "must be 1 to 255 characters")
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

// bodyLimit is the v0 operation's largest body class. A Callable with no
// parameters has no body class, and only the envelope's slack: issue/context, the
// one Callable whose row holds no database slot, is among them, so what it reads
// is bounded by that slack and maxConns rather than by any semaphore.
func (row *callableRow) bodyLimit() int64 {
	if len(row.path)+len(row.query)+len(row.body)+len(row.adapterParams) == 0 {
		return 0
	}
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
func (row *callableRow) v0Request(r *http.Request, inv *invocation) (*http.Request, *callableProblem) {
	rt, ok := routeForOp(row.op)
	if !ok {
		return nil, newCallableProblem(bdpUnavailable, "")
	}
	args := inv.args
	target := rt.specPathOf()
	pathValues := make(map[string]string, len(row.path))
	for _, name := range row.path {
		raw, ok := args[name]
		if !ok {
			return nil, invalidArgument(name, "is required")
		}
		var value string
		// `.` and `..` are excluded because a direct call cannot name them either:
		// ServeMux cleans them into a redirect to another route.
		if json.Unmarshal(raw, &value) != nil || string(raw) == "null" || value == "" || value == "." || value == ".." {
			return nil, invalidArgument(name, "must be a nonempty string")
		}
		pathValues[name] = value
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
	var bodyLength int64
	if len(row.body) > 0 || row.actorMember != "" {
		head, rest := row.v0Body(inv)
		body = io.MultiReader(bytes.NewReader(head), bytes.NewReader(rest))
		bodyLength = int64(len(head) + len(rest))
	}

	inner, err := http.NewRequestWithContext(r.Context(), rt.method, target, body)
	if err != nil {
		return nil, newCallableProblem(bdpMalformedRequest, "the arguments do not make a valid request")
	}
	if body == nil {
		inner.Body = http.NoBody // A server request's Body is never nil.
	}
	for name, value := range pathValues {
		inner.SetPathValue(name, value)
	}
	inner.ContentLength = bodyLength
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

// v0Body is the body v0 would have received, in two pieces. When every argument is a
// body member, which is every large operation but issue/update, the second piece is
// `arguments` itself: checkArgumentNames has held its names to the row's, and it needs
// no second buffer. What this layer supplied is the first, in front of it.
func (row *callableRow) v0Body(inv *invocation) (head, rest []byte) {
	if len(row.path) == 0 && len(row.query) == 0 && row.adapt == nil && !inv.rebuild {
		if len(inv.supplied) == 0 {
			return nil, inv.arguments
		}
		head = []byte{'{'}
		for _, name := range inv.supplied {
			head = appendBodyMember(head, name, inv.args[name])
		}
		if len(inv.args) == len(inv.supplied) {
			return append(head, '}'), nil
		}
		// Past the opening brace, whatever comes next is a member: args is not empty.
		return append(head, ','), inv.arguments[1:]
	}
	size := 2
	for _, raw := range inv.args {
		size += len(raw) + 64
	}
	head = append(make([]byte, 0, size), '{')
	for _, name := range row.body {
		if raw, ok := inv.args[name]; ok {
			head = appendBodyMember(head, name, raw)
		}
	}
	if raw, ok := inv.args[row.actorMember]; ok {
		head = appendBodyMember(head, row.actorMember, raw)
	}
	return append(head, '}'), nil
}

// appendBodyMember writes one member after the opening brace or an earlier member.
// The name is a declared one, which needs no escaping.
func appendBodyMember(body []byte, name string, raw json.RawMessage) []byte {
	if len(body) > 1 {
		body = append(body, ',')
	}
	body = append(body, '"')
	body = append(body, name...)
	body = append(body, '"', ':')
	return append(body, raw...)
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
		writeCallableResult(w, c.row, c.body.Bytes())
	default:
		problem := translateV0Problem(c.body.Bytes())
		requestInfo(r.Context()).code = Code(problem.Code)
		h.Set("Content-Type", problemMediaType)
		w.WriteHeader(problem.Status)
		_ = json.NewEncoder(w).Encode(problem)
	}
}

// writeCallableResult is the closed callableResult of Callables §5, with v0's
// bytes as `result`, written around them rather than copied.
func writeCallableResult(w io.Writer, row *callableRow, result []byte) {
	result = bytes.TrimSpace(result)
	if row.batch {
		changed := true
		if row.changed != nil {
			var members map[string]json.RawMessage
			if json.Unmarshal(result, &members) == nil {
				changed = row.changed(members)
			}
		}
		_, _ = fmt.Fprintf(w, `{"changed":%t,"result":`, changed)
	} else {
		_, _ = io.WriteString(w, `{"result":`)
	}
	_, _ = w.Write(result)
	_, _ = io.WriteString(w, "}\n")
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
