package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"runtime/debug"
	"slices"
	"strconv"
	"strings"
)

// The Core read surface of the Beads Protocol: the Scope root, the discovery
// document, the two fixed inventories and GET of a canonical Resource URL. See
// specs/profiles/core/core.md and specs/types/issue/issue.md.
//
// It serves the same store over the same roles v0's read handlers use, and it
// holds no storage of its own. What it adds is the projection: a row becomes a
// Bead record, an edge becomes a Link record, and refusals use Core's closed
// Problem table rather than v0's vocabulary. A bd serve process serves one
// workspace, so it is one Scope, rooted at the server root.
const (
	coreDiscoveryPath = "/bdp"
	coreBeadsPath     = "/bdp/beads"
	coreLinksPath     = "/bdp/links"

	// coreSpecRevision is Core discovery's `specRevision`: the commit of
	// gastownhall/bdp this surface was written against, branch
	// trish/preview1-issue-callables-20261104 (specs/profiles/core/core.md,
	// specs/types/issue/issue.md, specs/types/issue/callables.md). It names protocol
	// text, not this build.
	coreSpecRevision = "7620a095a9ce5bb4209775c17340e4f493c67cee"

	coreBDPVersion = "0.1"
	coreProfile    = "core"
	coreOrder      = "canonical-uri"

	// callablesCapability is Callables §1. Version "1" has a closed empty
	// detail, because invocation lives at the fixed <scope>/bdp/call.
	callablesCapability        = "https://github.com/gastownhall/bdp/capabilities/callables"
	callablesCapabilityVersion = "1"

	corePageDefaultItems = 100
	corePageMaximumItems = 1000
	coreSelectorBytes    = 4096
	coreSelectorDepth    = 32
	coreSelectorNodes    = 256

	opCoreScopeRoot = "coreScopeRoot"
	opCoreDiscovery = "coreDiscovery"
	opCoreBeads     = "coreListBeads"
	opCoreLinks     = "coreListLinks"
	opCoreResource  = "coreReadResource"

	coreCodeMalformedRequest = "malformed-request"
	coreCodeInvalidParameter = "invalid-parameter"
	coreCodeUnauthenticated  = "unauthenticated"
	coreCodeForbidden        = "forbidden"
	coreCodeNotFound         = "resource-not-found"
	coreCodeCursorExpired    = "cursor-expired"
	coreCodeLimitExceeded    = "limit-exceeded"
	coreCodeUnavailable      = "temporarily-unavailable"

	coreRetryAfterSeconds = 2
)

// coreProblemShape is one row of Core's closed Problem table.
type coreProblemShape struct {
	family string
	status int
	retry  string
}

var coreProblemShapes = map[string]coreProblemShape{
	coreCodeMalformedRequest: {"request", http.StatusBadRequest, "never"},
	coreCodeInvalidParameter: {"request", http.StatusBadRequest, "never"},
	coreCodeUnauthenticated:  {"authentication", http.StatusUnauthorized, "after-state-change"},
	coreCodeForbidden:        {"authorization", http.StatusForbidden, "after-state-change"},
	coreCodeNotFound:         {"not-found", http.StatusNotFound, "after-state-change"},
	coreCodeCursorExpired:    {"gone", http.StatusGone, "after-state-change"},
	coreCodeLimitExceeded:    {"size", http.StatusRequestEntityTooLarge, "never"},
	coreCodeUnavailable:      {"unavailable", http.StatusServiceUnavailable, "after-delay"},
}

// coreProblem is an RFC 9457 body with Core's required `code` and `retry`.
type coreProblem struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Code   string `json:"code"`
	Retry  string `json:"retry"`
	Detail string `json:"detail,omitempty"`
}

// writeCoreProblem answers one row of the table. The request line records the
// code, as it does for a v0 refusal. HEAD gets the same status and headers; the
// server drops the body.
func writeCoreProblem(w http.ResponseWriter, r *http.Request, code, detail string) {
	shape := coreProblemShapes[code]
	requestInfo(r.Context()).code = Code(code)
	h := w.Header()
	h.Set("Content-Type", problemMediaType)
	h.Set("Cache-Control", "private, no-store")
	if code == coreCodeUnauthenticated {
		h.Set("WWW-Authenticate", "Bearer")
	}
	if shape.retry == "after-delay" {
		h.Set("Retry-After", strconv.Itoa(coreRetryAfterSeconds))
	}
	w.WriteHeader(shape.status)
	_ = json.NewEncoder(w).Encode(coreProblem{
		Type:   bdpProblemBase + shape.family,
		Title:  http.StatusText(shape.status),
		Status: shape.status,
		Code:   code,
		Retry:  shape.retry,
		Detail: detail,
	})
}

// failCoreErr answers a failure from a role. A store that is busy or out of
// reach is Core's temporary failure; a miss is the uniform 404. Anything else is
// the unexpected fault Core says is a bodyless 500 with its details off the
// wire, and the log carries them instead.
func (s *Server) failCoreErr(w http.ResponseWriter, r *http.Request, err error) {
	rec := requestInfo(r.Context())
	res := ClassifyError(err)
	switch Code(res.Problem.Code) {
	case CodeNotFound:
		writeCoreProblem(w, r, coreCodeNotFound, "")
		return
	case CodeBusy, CodeDBUnavailable:
		writeCoreProblem(w, r, coreCodeUnavailable, "the store is temporarily unavailable")
		return
	}
	rec.code = Code(res.Problem.Code)
	if errors.Is(err, context.Canceled) {
		rec.code = codeClientClosed
	} else {
		s.event("request_error", "request_id", rec.id, "error", err.Error())
	}
	w.Header().Set("Cache-Control", "private, no-store")
	w.WriteHeader(http.StatusInternalServerError)
}

// coreGate runs serve behind the checks a v0 row gets, answered in Core's
// vocabulary: the request deadline, the credential unless the route is public,
// and the workspace stamp. A panic is the bodyless 500 too, rather than v0's
// `internal` Problem.
func (s *Server) coreGate(op string, public bool, serve http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := requestInfo(r.Context())
		rec.op = op
		defer func() {
			p := recover()
			if p == nil {
				return
			}
			if p == http.ErrAbortHandler {
				panic(p)
			}
			s.event("panic", "request_id", rec.id, "op", rec.op, "method", r.Method, "path", r.URL.Path,
				"error", fmt.Sprint(p), "stack", string(debug.Stack()))
			w.WriteHeader(http.StatusInternalServerError)
		}()

		ctx, cancel := context.WithTimeout(r.Context(), requestDeadline)
		defer cancel()
		r = r.WithContext(ctx)

		if !public && s.auth != nil && !s.authorize(discardResponse{}, r, rec) {
			writeCoreProblem(w, r, coreCodeUnauthenticated, "a bearer credential is required")
			return
		}
		if !public && s.checkProjectStamp(r) != nil {
			writeCoreProblem(w, r, coreCodeForbidden, "the "+ProjectIDHeader+" header names another workspace")
			return
		}
		serve(w, r)
	})
}

// discardResponse takes the v0 refusal authorize writes, whose log line and
// request id are what the caller wants from it; Core's Problem goes on the wire.
type discardResponse struct{}

func (discardResponse) Header() http.Header         { return http.Header{} }
func (discardResponse) Write(b []byte) (int, error) { return len(b), nil }
func (discardResponse) WriteHeader(int)             {}

// coreMethodNotAllowed is Core's bodyless 405: GET and HEAD are the only
// methods here, and with cross-origin access not enabled OPTIONS is refused too.
func coreMethodNotAllowed(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Allow", "GET, HEAD")
	w.WriteHeader(http.StatusMethodNotAllowed)
}

// mountCoreReads registers the Core read surface on the server's mux.
//
// Go's mux takes the most specific pattern, so these sit beside every /v0 row
// and beside /bdp/call without being able to shadow one. Resource URLs have no
// pattern of their own; see coreResourcesOr.
func (s *Server) mountCoreReads(mux *http.ServeMux) {
	mux.Handle("GET /{$}", s.coreGate(opCoreScopeRoot, true, s.serveCoreScopeRoot))
	mux.Handle("GET "+coreDiscoveryPath, s.coreGate(opCoreDiscovery, true, s.serveCoreDiscovery))
	mux.Handle("GET "+coreBeadsPath, s.coreGate(opCoreBeads, false, s.serveCoreBeads))
	mux.Handle("GET "+coreLinksPath, s.coreGate(opCoreLinks, false, s.serveCoreLinks))
	// The rest of `bdp/` is protocol navigation this Scope does not advertise.
	mux.Handle(coreDiscoveryPath+"/{rest...}", s.coreGate(opCoreResource, true, func(w http.ResponseWriter, r *http.Request) {
		writeCoreProblem(w, r, coreCodeNotFound, "no such route on this Scope")
	}))
	for _, method := range coreRefusedMethods {
		mux.HandleFunc(method+" "+coreDiscoveryPath, coreMethodNotAllowed)
		mux.HandleFunc(method+" "+coreBeadsPath, coreMethodNotAllowed)
		mux.HandleFunc(method+" "+coreLinksPath, coreMethodNotAllowed)
	}
}

// coreRefusedMethods are the methods Core's 405 answers. Anything rarer stays
// the catch-all's 404.
var coreRefusedMethods = []string{
	http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions,
}

// coreResourcesOr is the server's catch-all, because a Resource URL can be any
// path: a method-carrying wildcard at the root would conflict with the
// method-less /bdp/call patterns. Every registered route has already had its
// turn by the time a request lands here.
//
// A Resource address is one path segment, under neither a v0 route's first segment
// nor bdp. Any other path stays unrouted, answered the way it always was and with
// no credential asked for it.
func (s *Server) coreResourcesOr(unrouted http.Handler) http.Handler {
	reserved := map[string]bool{strings.TrimPrefix(coreDiscoveryPath, "/"): true}
	for _, rt := range routeTable {
		first, _, _ := strings.Cut(strings.TrimPrefix(rt.pattern, "/"), "/")
		reserved[first] = true
	}
	resources := s.coreGate(opCoreResource, false, s.serveCoreResource)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		segment := strings.TrimPrefix(r.URL.EscapedPath(), "/")
		switch {
		case reserved[segment] || strings.Contains(segment, "/"):
			unrouted.ServeHTTP(w, r)
		case r.Method == http.MethodGet || r.Method == http.MethodHead:
			resources.ServeHTTP(w, r)
		case slices.Contains(coreRefusedMethods, r.Method):
			coreMethodNotAllowed(w, r)
		default:
			unrouted.ServeHTTP(w, r)
		}
	})
}

// coreScope is the canonical Scope URL of this request: its scheme and Host.
//
// bd serve has no configured public URL, so the Host header is all there is.
// checkHost has already held it to the loopback spellings, the bind address and
// the operator's --allowed-host list, so a caller can pick among the names this
// server answers to and nothing else. Two of those names are two spellings of
// the Scope, and every id the request is told differs between them.
func coreScope(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	host := strings.ToLower(r.Host)
	if name, port, err := net.SplitHostPort(host); err == nil &&
		((scheme == "http" && port == "80") || (scheme == "https" && port == "443")) {
		host = name
		if strings.Contains(name, ":") {
			host = "[" + name + "]"
		}
	}
	return (&url.URL{Scheme: scheme, Host: host, Path: "/"}).String()
}

// setCoreNavigation sets what every successful read and every 304 carries.
func setCoreNavigation(h http.Header, scope string) {
	h.Set("Cache-Control", "private, no-store")
	h.Set("Link", "<"+scope+strings.TrimPrefix(coreDiscoveryPath, "/")+`>; rel="service-desc"`)
}

// serveCoreScopeRoot is Core's GET of the Scope URL: it must succeed and carry
// the service-desc Link, and its body is never machine discovery, so it has
// none. It discloses nothing but the address the caller just used, which is why
// no credential is asked of it.
func (s *Server) serveCoreScopeRoot(w http.ResponseWriter, r *http.Request) {
	setCoreNavigation(w.Header(), coreScope(r))
	if coreNotModified(w, r, "") {
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// coreDiscovery is the closed coreDiscovery schema, less the members this
// Scope has nothing to say with.
type coreDiscovery struct {
	BDPVersion   string                               `json:"bdpVersion"`
	Profile      string                               `json:"profile"`
	Scope        string                               `json:"scope"`
	SpecRevision string                               `json:"specRevision"`
	Order        string                               `json:"order"`
	Limits       coreDiscoveryLimits                  `json:"limits"`
	Capabilities map[string]map[string]map[string]any `json:"capabilities"`
}

type coreDiscoveryLimits struct {
	Page     corePageLimits     `json:"page"`
	Selector coreSelectorLimits `json:"selector"`
}

type corePageLimits struct {
	DefaultItems int `json:"defaultItems"`
	MaximumItems int `json:"maximumItems"`
}

type coreSelectorLimits struct {
	Bytes int `json:"bytes"`
	Depth int `json:"depth"`
	Nodes int `json:"nodes"`
}

// serveCoreDiscovery answers GET <scope>/bdp. Everything in it is either the
// address the caller used or a constant of this build, so it asks for no
// credential; the v0 identity handshake, which names the repo and the database,
// stays guarded.
//
// The profile is `core` and Callables is advertised beside it. Write's generic
// verbs are not served, so the Write literal would be a promise nothing keeps.
func (s *Server) serveCoreDiscovery(w http.ResponseWriter, r *http.Request) {
	if len(r.URL.RawQuery) != 0 {
		writeCoreProblem(w, r, coreCodeInvalidParameter, "discovery takes no query parameters")
		return
	}
	scope := coreScope(r)
	if !acceptsMedia(r.Header.Values("Accept"), bdpMediaType) {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}
	setCoreNavigation(w.Header(), scope)
	if coreNotModified(w, r, "") {
		return
	}
	writeCoreJSON(w, coreDiscovery{
		BDPVersion:   coreBDPVersion,
		Profile:      coreProfile,
		Scope:        scope,
		SpecRevision: coreSpecRevision,
		Order:        coreOrder,
		Limits: coreDiscoveryLimits{
			Page:     corePageLimits{DefaultItems: corePageDefaultItems, MaximumItems: corePageMaximumItems},
			Selector: coreSelectorLimits{Bytes: coreSelectorBytes, Depth: coreSelectorDepth, Nodes: coreSelectorNodes},
		},
		Capabilities: map[string]map[string]map[string]any{
			callablesCapability: {callablesCapabilityVersion: {}},
		},
	})
}

// marshalCoreJSON is the one spelling of a BDP JSON body: no HTML escaping, so
// a title carrying `<` is the bytes the store holds.
func marshalCoreJSON(value any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(value); err != nil {
		return nil, err
	}
	return bytes.TrimRight(buf.Bytes(), "\n"), nil
}

// writeCoreJSON answers 200 with one BDP JSON representation. A value that will
// not marshal is a fault of this build, so it panics into coreGate's bodyless
// 500 before any byte has gone out.
func writeCoreJSON(w http.ResponseWriter, value any) {
	body, err := marshalCoreJSON(value)
	if err != nil {
		panic(fmt.Sprintf("marshal BDP response: %v", err))
	}
	w.Header().Set("Content-Type", bdpMediaType)
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.WriteHeader(http.StatusOK)
	//nolint:gosec // G705: BDP JSON body, JSON-encoded content, no HTML sink.
	_, _ = w.Write(body)
}

// coreNotModified applies RFC 9110's GET and HEAD preconditions as Core spells
// them, answering the bodyless 304 or 412 itself and reporting whether it did.
// etag is the representation's quoted entity tag, or "" when it has none. It
// runs after authorization, query and cursor checks, so an ordinary refusal is
// never replaced by a 304.
//
// No representation here has a modification date, so If-Modified-Since and
// If-Unmodified-Since have nothing to compare and do not apply.
func coreNotModified(w http.ResponseWriter, r *http.Request, etag string) bool {
	if values := r.Header.Values("If-Match"); len(values) > 0 && !coreETagMatches(values, etag, true) {
		w.WriteHeader(http.StatusPreconditionFailed)
		return true
	}
	if values := r.Header.Values("If-None-Match"); len(values) > 0 && coreETagMatches(values, etag, false) {
		w.WriteHeader(http.StatusNotModified)
		return true
	}
	return false
}

// coreETagMatches is the list comparison of RFC 9110 §13.1. `*` matches any
// representation, and no tag matches a representation that has none. If-Match
// compares strongly, so a `W/` candidate never matches; If-None-Match compares
// weakly. Every tag served here is strong.
func coreETagMatches(values []string, etag string, strong bool) bool {
	for _, value := range values {
		for _, candidate := range strings.Split(value, ",") {
			candidate = strings.TrimSpace(candidate)
			if candidate == "*" {
				return true
			}
			if !strong {
				candidate = strings.TrimPrefix(candidate, "W/")
			}
			if etag != "" && candidate == etag {
				return true
			}
		}
	}
	return false
}
