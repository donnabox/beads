// Package wiretest lets a fake bd serve in a test keep speaking v0's vocabulary
// while the client it talks to speaks Callables. A test's canned handler asserts
// and answers in the operation's own terms, which is what it is testing, so
// wrapping it in V0 is the one change the client's move to /bdp/call asks of it.
//
// V0 checks what every invocation owes the mount (the method, both media types, the
// closed envelope), so a client that stops sending them fails every such test.
// The translation it does is the mount's, written out here because the mount lives
// in internal/httpapi, which a client test should not need. internal/httpapi's
// TestWireTestRoutesMatchTheServerTable holds Routes to the server's table.
//
// It does not import the wire package, whose own tests use it.
package wiretest

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
)

// Route is where one Callable lives on v0.
type Route struct {
	Method string
	// Path is the document's spelling, with its one parameter in braces.
	Path string
	// Body lists the arguments that go in the body when a Callable has both query
	// and body parameters. Empty means the method decides: a GET's arguments are
	// its query, and any other method's are its body.
	Body []string
}

// Routes is every declared Callable.
var Routes = map[string]Route{
	"issue/context":                  {Method: http.MethodGet, Path: "/v0/beads/context"},
	"issue/ready":                    {Method: http.MethodGet, Path: "/v0/beads/ready"},
	"issue/ready-count":              {Method: http.MethodGet, Path: "/v0/beads/ready:count"},
	"issue/stats":                    {Method: http.MethodGet, Path: "/v0/beads/stats"},
	"issue/blocked":                  {Method: http.MethodGet, Path: "/v0/beads/blocked"},
	"issue/dep/cycles":               {Method: http.MethodGet, Path: "/v0/beads/dependencies/cycles"},
	"issue/dep/tree":                 {Method: http.MethodGet, Path: "/v0/beads/dependencies/tree"},
	"issue/list":                     {Method: http.MethodGet, Path: "/v0/beads/issues"},
	"issue/create":                   {Method: http.MethodPost, Path: "/v0/beads/issues"},
	"issue/query":                    {Method: http.MethodGet, Path: "/v0/beads/issues:query"},
	"issue/count":                    {Method: http.MethodGet, Path: "/v0/beads/issues:count"},
	"issue/show":                     {Method: http.MethodGet, Path: "/v0/beads/issues/{id}"},
	"issue/update":                   {Method: http.MethodPatch, Path: "/v0/beads/issues/{id}"},
	"issue/dep/list":                 {Method: http.MethodGet, Path: "/v0/beads/issues/{id}/related"},
	"issue/comment":                  {Method: http.MethodPost, Path: "/v0/beads/issues/{id}/comments"},
	"issue/claim":                    {Method: http.MethodPost, Path: "/v0/beads/issues/{id}:claim"},
	"issue/unclaim":                  {Method: http.MethodPost, Path: "/v0/beads/issues/{id}:release"},
	"issue/heartbeat":                {Method: http.MethodPost, Path: "/v0/beads/issues/{id}:heartbeat"},
	"issue/close":                    {Method: http.MethodPost, Path: "/v0/beads/issues/{id}:close"},
	"issue/advance":                  {Method: http.MethodPost, Path: "/v0/beads/issues/{id}:advanceMolecule"},
	"issue/reopen":                   {Method: http.MethodPost, Path: "/v0/beads/issues/{id}:reopen"},
	"issue/metadata/compare-and-set": {Method: http.MethodPost, Path: "/v0/beads/issues/{id}:casMetadata"},
	"issue/purge":                    {Method: http.MethodPost, Path: "/v0/beads/issues:sweep"},
	"issue/prune":                    {Method: http.MethodPost, Path: "/v0/beads/issues:sweep"},
	"issue/delete":                   {Method: http.MethodPost, Path: "/v0/beads/issues:delete"},
	"issue/get-many":                 {Method: http.MethodPost, Path: "/v0/beads/issues:batchGet"},
	"issue/reclaim":                  {Method: http.MethodPost, Path: "/v0/beads/issues:reclaim"},
	"issue/create-many":              {Method: http.MethodPost, Path: "/v0/beads/issues:batchCreate"},
	"issue/batch":                    {Method: http.MethodPost, Path: "/v0/beads/issues:batchApply"},
	"issue/close-many":               {Method: http.MethodPost, Path: "/v0/beads/issues:batchClose"},
	"issue/claim-next":               {Method: http.MethodPost, Path: "/v0/beads/issues:claimNext", Body: []string{"actor"}},
	"issue/dep/edges":                {Method: http.MethodGet, Path: "/v0/beads/dependencies"},
	"issue/dep/count":                {Method: http.MethodGet, Path: "/v0/beads/dependencies:count"},
	"issue/blocking-annotations":     {Method: http.MethodGet, Path: "/v0/beads/dependencies/blocking"},
	"issue/dep/add":                  {Method: http.MethodPost, Path: "/v0/beads/dependencies:add"},
	"issue/dep/remove":               {Method: http.MethodPost, Path: "/v0/beads/dependencies:remove"},
	"config/list":                    {Method: http.MethodGet, Path: "/v0/beads/config"},
	"config/get":                     {Method: http.MethodGet, Path: "/v0/beads/config/{key}"},
	"config/set":                     {Method: http.MethodPut, Path: "/v0/beads/config/{key}"},
	"config/unset":                   {Method: http.MethodDelete, Path: "/v0/beads/config/{key}"},
	"memory/list":                    {Method: http.MethodGet, Path: "/v0/beads/memories"},
	"memory/remember":                {Method: http.MethodPost, Path: "/v0/beads/memories"},
	"memory/recall":                  {Method: http.MethodGet, Path: "/v0/beads/memories/{key}"},
	"memory/forget":                  {Method: http.MethodDelete, Path: "/v0/beads/memories/{key}"},
	"events/list":                    {Method: http.MethodGet, Path: "/v0/beads/events"},
}

const (
	mount     = "/bdp/call/"
	mediaType = "application/vnd.bdp.v0.1+json"
)

// V0 serves a test's v0-speaking handler to a client that invokes Callables. An
// invocation reaches h as the request v0 would have received, with v0's own
// response headers and problem bodies, and comes back under the Callables envelope.
// Anything not under the mount (/healthz, the events stream) reaches h unchanged.
func V0(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		at := strings.Index(r.URL.Path, mount)
		if at < 0 {
			h.ServeHTTP(w, r)
			return
		}
		route, ok := Routes[r.URL.Path[at+len(mount):]]
		switch {
		case !ok:
			writeRefusal(w, http.StatusNotFound, "callable-not-found", "no reachable Callable has this name")
			return
		case r.Method != http.MethodPost:
			w.Header().Set("Allow", http.MethodPost)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		case r.Header.Get("Content-Type") != mediaType:
			writeRefusal(w, http.StatusUnsupportedMediaType, "unsupported-media-type", "got "+r.Header.Get("Content-Type"))
			return
		case !strings.Contains(r.Header.Get("Accept"), mediaType):
			w.WriteHeader(http.StatusNotAcceptable)
			return
		}
		inner, detail := v0Request(r, r.URL.Path[:at], route)
		if detail != "" {
			writeRefusal(w, http.StatusBadRequest, "malformed-request", detail)
			return
		}
		env := &envelope{w: w}
		h.ServeHTTP(env, inner)
		env.finish()
	})
}

// Logical is what one request asks for in v0's terms.
type Logical struct {
	Method string
	Path   string
	// RequestURI is the path as escaped, and the query.
	RequestURI string
	Query      url.Values
}

// LogicalRequest spells a request the way v0 would have received it, for a test that
// watches real traffic pass and asserts on the operations it saw. Its Body is left
// readable. A request that is no invocation is its own spelling.
func LogicalRequest(r *http.Request) Logical {
	if at := strings.Index(r.URL.Path, mount); at >= 0 {
		if route, ok := Routes[r.URL.Path[at+len(mount):]]; ok && r.Method == http.MethodPost {
			raw, err := io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewReader(raw))
			if err == nil {
				copied := r.Clone(r.Context())
				copied.Body = io.NopCloser(bytes.NewReader(raw))
				if inner, detail := v0Request(copied, r.URL.Path[:at], route); detail == "" {
					return Logical{inner.Method, inner.URL.Path, inner.URL.RequestURI(), inner.URL.Query()}
				}
			}
		}
	}
	return Logical{r.Method, r.URL.Path, r.URL.RequestURI(), r.URL.Query()}
}

// HandlerFunc is V0 for a handler function, so that a fake written inline needs
// one word changed.
func HandlerFunc(f func(http.ResponseWriter, *http.Request)) http.Handler {
	return V0(http.HandlerFunc(f))
}

// v0Request is the request v0 would have received. prefix is where the server is
// mounted, which is whatever came before /bdp/call.
func v0Request(r *http.Request, prefix string, route Route) (*http.Request, string) {
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, "the invocation body could not be read"
	}
	var invocation struct {
		Arguments map[string]json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(raw, &invocation); err != nil || invocation.Arguments == nil {
		return nil, "`arguments` is required and is an object"
	}
	args := invocation.Arguments

	path, rawPath := route.Path, route.Path
	if open := strings.IndexByte(path, '{'); open >= 0 {
		end := strings.IndexByte(path, '}')
		var value string
		if json.Unmarshal(args[path[open+1:end]], &value) != nil || value == "" {
			return nil, "`" + path[open+1:end] + "` is required"
		}
		delete(args, path[open+1:end])
		rawPath = path[:open] + escapeSegment(value) + path[end+1:]
		path = path[:open] + value + path[end+1:]
	}

	query := url.Values{}
	body := map[string]json.RawMessage{}
	for name, value := range args {
		inBody := route.Method != http.MethodGet
		if len(route.Body) > 0 {
			inBody = false
			for _, member := range route.Body {
				inBody = inBody || member == name
			}
		}
		if inBody {
			body[name] = value
			continue
		}
		var one string
		var many []string
		switch {
		case json.Unmarshal(value, &one) == nil:
			query.Set(name, one)
		case json.Unmarshal(value, &many) == nil:
			query[name] = many
		default:
			return nil, "`" + name + "` is not a query value"
		}
	}

	inner := r.Clone(r.Context())
	inner.Method = route.Method
	inner.URL.Path, inner.URL.RawPath = prefix+path, prefix+rawPath
	inner.URL.RawQuery = query.Encode()
	inner.RequestURI = inner.URL.RequestURI()
	inner.Header.Set("Accept", "application/json, application/problem+json")
	inner.Header.Del("Content-Type")
	inner.Body = http.NoBody
	inner.ContentLength = 0
	if len(body) > 0 || route.Method == http.MethodPost || route.Method == http.MethodPatch {
		encoded, err := json.Marshal(body)
		if err != nil {
			return nil, "the arguments could not be encoded"
		}
		inner.Body = io.NopCloser(bytes.NewReader(encoded))
		inner.ContentLength = int64(len(encoded))
		inner.Header.Set("Content-Type", "application/json")
	}
	return inner, ""
}

// escapeSegment is the escaping paths.go applies to an id or key, so a fake that
// asserts on the escaped path sees the v0 spelling of that path.
func escapeSegment(s string) string {
	seg := strings.ReplaceAll(url.PathEscape(s), ":", "%3A")
	switch seg {
	case ".":
		return "%2E"
	case "..":
		return "%2E%2E"
	}
	return seg
}

// envelope turns a v0 answer into the mount's. A success streams through, so a test
// of a slow or truncated body still sees one; a refusal is buffered until it can be
// translated.
type envelope struct {
	w       http.ResponseWriter
	status  int
	opened  bool
	refused bytes.Buffer
}

func (e *envelope) Header() http.Header { return e.w.Header() }

func (e *envelope) WriteHeader(status int) {
	if e.status == 0 {
		e.status = status
	}
}

func (e *envelope) success() bool { return e.status >= 200 && e.status < 300 }

func (e *envelope) Write(b []byte) (int, error) {
	e.WriteHeader(http.StatusOK)
	if !e.success() {
		return e.refused.Write(b)
	}
	e.open()
	return e.w.Write(b)
}

func (e *envelope) open() {
	if e.opened {
		return
	}
	e.opened = true
	// A reverse proxy behind this copies v0's length, which the envelope makes wrong.
	e.w.Header().Del("Content-Length")
	e.w.Header().Set("Content-Type", mediaType)
	e.w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(e.w, `{"result":`)
}

// Flush sends a success's headers early, for a test that holds its body back.
func (e *envelope) Flush() {
	e.WriteHeader(http.StatusOK)
	if !e.success() {
		return
	}
	e.open()
	if f, ok := e.w.(http.Flusher); ok {
		f.Flush()
	}
}

// Hijack hands a test the connection, for one that resets it.
func (e *envelope) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return http.NewResponseController(e.w).Hijack()
}

func (e *envelope) finish() {
	switch {
	case e.status == 0:
		// A handler that wrote nothing: net/http's empty 200.
	case e.success():
		if e.opened {
			_, _ = io.WriteString(e.w, "}\n")
			return
		}
		e.w.WriteHeader(e.status)
	default:
		e.translate()
	}
}

// translate is the mount's §5: v0's body goes under `v0`, and the status is the
// BDP code's where it differs on the wire, which is the case that matters: a client
// has to read v0's status, not this one. An answer with no v0 code is something in
// front of bd serve, and the mount does not touch it.
func (e *envelope) translate() {
	raw := e.refused.Bytes()
	var v0 struct {
		Code string `json:"code"`
	}
	if json.Unmarshal(raw, &v0) != nil || v0.Code == "" {
		e.w.WriteHeader(e.status)
		_, _ = e.w.Write(raw)
		return
	}
	status := e.status
	switch v0.Code {
	case "internal", "busy", "db_unavailable":
		status = http.StatusServiceUnavailable
	}
	problem, err := json.Marshal(struct {
		Status int             `json:"status"`
		Code   string          `json:"code"`
		V0     json.RawMessage `json:"v0"`
	}{status, "translated", raw})
	if err != nil {
		panic(err)
	}
	e.w.Header().Del("Content-Length")
	e.w.Header().Set("Content-Type", "application/problem+json")
	e.w.WriteHeader(status)
	_, _ = e.w.Write(problem)
}

func writeRefusal(w http.ResponseWriter, status int, code, detail string) {
	body, err := json.Marshal(map[string]any{"status": status, "code": code, "detail": detail})
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}
