package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/resourcequery"
)

const queryASTListFixture = `{"query":{"ir":"bdp-query-design-1","profile":"list","selection":{"op":"scan","collection":"beads"},"result":{"form":"resources"}}}`

func TestQueryASTReferenceInputClassification(t *testing.T) {
	for _, tc := range []struct {
		name, references, code string
		status                 int
	}{
		{"malformed", `7`, "invalid-parameter", http.StatusBadRequest},
		{"batch", `{"binding":"prior"}`, "capability-not-supported", http.StatusUnprocessableEntity},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, q, _, executed := queryASTTestAdapter(nil)
			body := fmt.Sprintf(`{"query":{"ir":"bdp-query-design-1","profile":"query","selection":{"op":"refs","collection":"beads","references":%s},"result":{"form":"resources"}}}`, tc.references)
			r := queryASTTestRequest(http.MethodPost, "/read/?readVersion=1&operation=query", body)
			w := httptest.NewRecorder()
			q.serve(g, w, r)
			var problem struct{ Code string }
			if json.Unmarshal(w.Body.Bytes(), &problem) != nil || w.Code != tc.status || problem.Code != tc.code || *executed != 0 {
				t.Fatalf("reference admission reached executor or wrong classification: status=%d body=%s executions=%d", w.Code, w.Body.String(), *executed)
			}
		})
	}
}

func queryASTTestAdapter(executeErr error) (*GraphRead, *QueryAST, *int, *int) {
	validated, executed := new(int), new(int)
	g := &GraphRead{scope: "https://example.test/read/", path: "/read/"}
	q := &QueryAST{
		executions:        make(chan struct{}, QueryASTMaxConcurrentExecutions),
		validateAuthority: func(context.Context) error { *validated++; return nil },
		execute: func(context.Context, []byte) (resourcequery.Response, error) {
			*executed++
			return resourcequery.Response{Result: resourcequery.ResourcesResult{Form: "resources", ResourceKind: graph.KindBead, Items: []json.RawMessage{}}, Evaluation: resourcequery.Evaluation{Scope: "response", Complete: true, Traversals: []resourcequery.Traversal{}}}, executeErr
		},
	}
	return g, q, validated, executed
}

func TestQueryASTConcurrentExecutionBound(t *testing.T) {
	g, q, _, _ := queryASTTestAdapter(nil)
	q.validateAuthority = func(context.Context) error { return nil }
	var calls atomic.Int32
	entered, release, done := make(chan struct{}), make(chan struct{}), make(chan struct{})
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	q.execute = func(ctx context.Context, _ []byte) (resourcequery.Response, error) {
		if calls.Add(1) == 1 {
			close(entered)
			select {
			case <-release:
			case <-ctx.Done():
				return resourcequery.Response{}, ctx.Err()
			}
		}
		return resourcequery.Response{Result: resourcequery.ResourcesResult{Form: "resources", ResourceKind: "bead", Items: []json.RawMessage{}},
			Evaluation: resourcequery.Evaluation{Scope: "response", Complete: true, Traversals: []resourcequery.Traversal{}}}, nil
	}
	first := httptest.NewRecorder()
	go func() {
		defer close(done)
		q.serve(g, first, queryASTTestRequest("POST", "/read/?readVersion=1&operation=list", queryASTListFixture).WithContext(ctx))
	}()
	t.Cleanup(func() {
		cancel()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Error("Query test executor did not stop after cancellation")
		}
	})
	select {
	case <-entered:
	case <-ctx.Done():
		t.Fatal("first Query did not enter executor")
	}
	second := httptest.NewRecorder()
	q.serve(g, second, queryASTTestRequest("POST", "/read/?readVersion=1&operation=list", queryASTListFixture).WithContext(ctx))
	if second.Code != http.StatusServiceUnavailable || calls.Load() != 1 || second.Header().Get("Retry-After") != "1" || !strings.Contains(second.Body.String(), `"code":"temporarily-unavailable"`) {
		t.Fatalf("concurrent Query bypassed execution bound: status=%d calls=%d body=%s", second.Code, calls.Load(), second.Body.String())
	}
	close(release)
	select {
	case <-done:
	case <-ctx.Done():
		t.Fatal("first Query did not finish")
	}
	if first.Code != http.StatusOK || len(q.executions) != 0 {
		t.Fatalf("Query slot not released: status=%d active=%d", first.Code, len(q.executions))
	}
	third := httptest.NewRecorder()
	q.serve(g, third, queryASTTestRequest("POST", "/read/?readVersion=1&operation=list", queryASTListFixture).WithContext(ctx))
	if third.Code != http.StatusOK || calls.Load() != 2 || len(q.executions) != 0 {
		t.Fatalf("Query execution slot not reusable: status=%d calls=%d active=%d", third.Code, calls.Load(), len(q.executions))
	}
}

func TestQueryASTMissingExecutorConfiguration(t *testing.T) {
	for _, missing := range []string{"authority", "execute", "slots"} {
		t.Run(missing, func(t *testing.T) {
			g, q, validated, executed := queryASTTestAdapter(nil)
			switch missing {
			case "authority":
				q.validateAuthority = nil
			case "execute":
				q.execute = nil
			case "slots":
				q.executions = nil
			}
			w := httptest.NewRecorder()
			q.serve(g, w, queryASTTestRequest("POST", "/read/?readVersion=1&operation=list", queryASTListFixture))
			if w.Code != http.StatusServiceUnavailable || *validated != 0 || *executed != 0 {
				t.Fatalf("misconfigured handler reached callbacks: status=%d validations=%d executions=%d", w.Code, *validated, *executed)
			}
		})
	}
}

func queryASTTestRequest(method, target, body string) *http.Request {
	r := httptest.NewRequest(method, target, strings.NewReader(body))
	r.Header.Set("Content-Type", queryASTMedia)
	r.Header.Set("Accept", queryASTMedia)
	return r
}

func TestQueryASTTransport(t *testing.T) {
	cases := []struct {
		name, method, target, body, contentType, accept string
		status, calls                                   int
		code                                            string
	}{
		{name: "success", method: "POST", target: "/read/?readVersion=1&operation=list", body: queryASTListFixture, status: 200, calls: 1},
		{name: "wrong method", method: "GET", target: "/read/?readVersion=1&operation=list", status: 405},
		{name: "missing version", method: "POST", target: "/read/?operation=list", body: queryASTListFixture, status: 400, code: "invalid-parameter"},
		{name: "duplicate version", method: "POST", target: "/read/?readVersion=1&readVersion=1&operation=list", body: queryASTListFixture, status: 400, code: "invalid-parameter"},
		{name: "future version", method: "POST", target: "/read/?readVersion=2&operation=list", body: queryASTListFixture, status: 422, code: "capability-not-supported"},
		{name: "extra parameter", method: "POST", target: "/read/?readVersion=1&operation=list&x=1", body: queryASTListFixture, status: 400, code: "invalid-parameter"},
		{name: "malformed JSON", method: "POST", target: "/read/?readVersion=1&operation=list", body: `{`, status: 400, code: "malformed-request"},
		{name: "duplicate JSON", method: "POST", target: "/read/?readVersion=1&operation=list", body: `{"query":{},"query":{}}`, status: 400, code: "malformed-request"},
		{name: "unknown body member", method: "POST", target: "/read/?readVersion=1&operation=list", body: strings.TrimSuffix(queryASTListFixture, "}") + `,"extra":1}`, status: 400, code: "invalid-parameter"},
		{name: "wrong profile", method: "POST", target: "/read/?readVersion=1&operation=query", body: queryASTListFixture, status: 400, code: "invalid-parameter"},
		{name: "wrong media", method: "POST", target: "/read/?readVersion=1&operation=list", body: queryASTListFixture, contentType: "application/json", status: 415},
		{name: "media parameter", method: "POST", target: "/read/?readVersion=1&operation=list", body: queryASTListFixture, contentType: queryASTMedia + "; charset=utf-8", status: 415},
		{name: "not acceptable", method: "POST", target: "/read/?readVersion=1&operation=list", body: queryASTListFixture, accept: "application/json", status: 406, calls: 0},
		{name: "q zero overrides wildcard", method: "POST", target: "/read/?readVersion=1&operation=list", body: queryASTListFixture, accept: queryASTMedia + ";q=0, */*;q=1", status: 406, calls: 0},
		{name: "oversized body", method: "POST", target: "/read/?readVersion=1&operation=list", body: strings.Repeat(" ", queryASTRequestLimit+1), status: 413, code: "request-too-large"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g, q, validated, executed := queryASTTestAdapter(nil)
			r := queryASTTestRequest(tc.method, tc.target, tc.body)
			if tc.contentType != "" {
				r.Header.Set("Content-Type", tc.contentType)
			}
			if tc.accept != "" {
				r.Header.Set("Accept", tc.accept)
			}
			w := httptest.NewRecorder()
			q.serve(g, w, r)
			if w.Code != tc.status || *executed != tc.calls || *validated != 1 {
				t.Fatalf("status=%d executions=%d authority=%d body=%s", w.Code, *executed, *validated, w.Body.String())
			}
			if tc.code != "" {
				var p map[string]any
				if err := json.Unmarshal(w.Body.Bytes(), &p); err != nil {
					t.Fatal(err)
				}
				if p["code"] != tc.code {
					t.Fatalf("problem=%v", p)
				}
				if tc.code == "capability-not-supported" && (p["capability"] != queryASTCapability || p["capabilityVersion"] != "1") {
					t.Fatalf("capability problem=%v", p)
				}
			}
			if tc.status == 200 && (w.Header().Get("Content-Type") != queryASTMedia || w.Header().Get("BDP-Collection-Kind") != "bead" || !strings.Contains(w.Header().Get("Link"), "service-desc")) {
				t.Fatal(w.Header())
			}
			if tc.status == 405 && w.Header().Get("Allow") != "POST" {
				t.Fatal(w.Header())
			}
		})
	}
}

func TestQueryASTRouteIsolation(t *testing.T) {
	g, q, _, _ := queryASTTestAdapter(nil)
	for _, tc := range []struct {
		method, target string
		matches        bool
	}{
		{"GET", "/read/", false}, {"HEAD", "/read/", false}, {"POST", "/read/", true},
		{"GET", "/read/?operation=list", false}, {"DELETE", "/read/?readVersion=1&operation=list", false},
		{"GET", "/read/?readVersion=1", false}, {"HEAD", "/read/?cursor=x", false},
		{"POST", "/read/operations/read?readVersion=1&operation=query", false},
		{"POST", "/read/beads/a?readVersion=1&operation=query", false},
	} {
		r := queryASTTestRequest(tc.method, tc.target, "")
		if q.matches(g, r) != tc.matches {
			t.Fatalf("route %s %s", tc.method, tc.target)
		}
	}
}

func TestQueryASTStrictBodyAndDeferredControls(t *testing.T) {
	cases := []struct{ tail, code string }{
		{`,"sort":{"path":"/properties/title","direction":"asc"}`, "capability-not-supported"},
		{`,"sort":{"path":null,"direction":"asc"}`, "invalid-parameter"},
		{`,"sort":{"path":"/bad~2","direction":"asc"}`, "invalid-parameter"},
		{`,"limit":1`, "capability-not-supported"},
		{`,"limit":0`, "invalid-parameter"},
		{`,"sort":{"path":"","direction":"asc"},"limit":0`, "invalid-parameter"},
	}
	for _, tc := range cases {
		_, err := decodeQueryASTBody([]byte(strings.TrimSuffix(queryASTListFixture, "}")+tc.tail+"}"), "list")
		var input *queryASTInputFailure
		if !errors.As(err, &input) || input.code != tc.code {
			t.Fatalf("%s: %v", tc.tail, err)
		}
	}
}

func TestQueryASTAuthorityAndCancellation(t *testing.T) {
	for _, cancelled := range []bool{false, true} {
		g, q, _, executed := queryASTTestAdapter(nil)
		if !cancelled {
			q.validateAuthority = func(context.Context) error { return context.DeadlineExceeded }
		}
		r := queryASTTestRequest("POST", "/read/?readVersion=1&operation=list", queryASTListFixture)
		if cancelled {
			ctx, cancel := context.WithCancel(r.Context())
			cancel()
			r = r.WithContext(ctx)
		}
		w := httptest.NewRecorder()
		q.serve(g, w, r)
		if w.Code != 503 || *executed != 0 {
			t.Fatalf("cancelled=%t status=%d executions=%d", cancelled, w.Code, *executed)
		}
	}
}

func TestQueryASTUnavailableDoesNotDisclosePointer(t *testing.T) {
	g, q, _, _ := queryASTTestAdapter(resourcequery.ErrUnavailable)
	r := queryASTTestRequest("POST", "/read/?readVersion=1&operation=list", queryASTListFixture)
	w := httptest.NewRecorder()
	q.serve(g, w, r)
	var problem map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &problem); err != nil {
		t.Fatal(err)
	}
	if w.Code != 404 || problem["at"] != nil || problem["detail"] != nil {
		t.Fatalf("%d %v", w.Code, problem)
	}
}

func TestQueryASTMediaAndParameters(t *testing.T) {
	for _, tc := range []struct {
		accept  string
		allowed bool
	}{
		{queryASTMedia, true}, {"*/*", true}, {"application/*", true}, {"application/json", false},
		{queryASTMedia + ";q=0, */*;q=1", false}, {queryASTMedia + ";q=0.5", true},
		{queryASTMedia + ";profile=other", false}, {queryASTMedia + ";q=2", false},
	} {
		if queryASTAccepts(tc.accept) != tc.allowed {
			t.Fatalf("Accept %q", tc.accept)
		}
	}
	for _, raw := range []string{"readVersion=&operation=list", "readVersion=x&operation=list", "readVersion=1&operation=other", "readVersion=1&operation=list&cursor=x"} {
		values, err := url.ParseQuery(raw)
		if err != nil {
			t.Fatal(err)
		}
		if _, code := queryASTParameters(values); code != "invalid-parameter" {
			t.Fatalf("%s: %s", raw, code)
		}
	}
}

func TestQueryASTHeadersAndRawTarget(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*http.Request)
		status int
	}{
		{"duplicate content type", func(r *http.Request) { r.Header.Add("Content-Type", queryASTMedia) }, 415},
		{"conditional POST", func(r *http.Request) { r.Header.Set("If-None-Match", "*") }, 400},
		{"compressed body", func(r *http.Request) { r.Header.Set("Content-Encoding", "gzip") }, 415},
		{"raw target mismatch", func(r *http.Request) { r.RequestURI = "/read/%2e/?readVersion=1&operation=list" }, 404},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, q, _, executed := queryASTTestAdapter(nil)
			r := queryASTTestRequest("POST", "/read/?readVersion=1&operation=list", queryASTListFixture)
			tc.mutate(r)
			w := httptest.NewRecorder()
			q.serve(g, w, r)
			if w.Code != tc.status || *executed != 0 {
				t.Fatalf("status=%d executions=%d", w.Code, *executed)
			}
		})
	}
}

// A body that would fail if read proves negotiation precedes even JSON parsing,
// not merely the final write. The provider fake also returns sensitive failure
// data, which must never be requested for an unacceptable representation.
type queryASTBodyReadTripwire struct{ reads int }

func (b *queryASTBodyReadTripwire) Read([]byte) (int, error) {
	b.reads++
	return 0, errors.New("body must not be read for bad Accept")
}
func (b *queryASTBodyReadTripwire) Close() error { return nil }

func TestQueryASTUnacceptableMediaBeforeBodyAndExecution(t *testing.T) {
	for _, accept := range []string{"application/json", queryASTMedia + ";q=0, */*;q=1"} {
		g, q, validated, executed := queryASTTestAdapter(errors.New("internal provider detail"))
		request := queryASTTestRequest("POST", "/read/?readVersion=1&operation=list", "")
		body := &queryASTBodyReadTripwire{}
		request.Body = body
		request.ContentLength = -1
		request.Header.Set("Accept", accept)
		response := httptest.NewRecorder()
		q.serve(g, response, request)
		if response.Code != 406 || body.reads != 0 || *executed != 0 || *validated != 1 {
			t.Fatalf("Accept=%q status=%d bodyReads=%d execution=%d authority=%d", accept, response.Code, body.reads, *executed, *validated)
		}
		if response.Body.Len() != 0 || response.Header().Get("Content-Type") != "" || response.Header().Get("Content-Length") != "0" {
			t.Fatalf("unacceptable response leaked representation: headers=%v body=%s", response.Header(), response.Body.String())
		}
	}
}

func TestQueryASTExecutionFailureClassesAndSanitization(t *testing.T) {
	const sensitive = "private-query-storage-identity-and-driver-details"
	for _, tc := range []struct {
		name   string
		err    error
		status int
		code   string
	}{
		{"unavailable", errors.Join(resourcequery.ErrUnavailable, &resourcequery.Failure{Code: "hidden-lookup", At: "/query/selection/references/0", Message: sensitive}), 404, "resource-not-found"},
		{"unsupported", resourcequery.ErrUnsupported, 422, "capability-not-supported"},
		{"invalid", resourcequery.ErrInvalid, 400, "invalid-parameter"},
		{"limit", resourcequery.ErrLimit, 413, "limit-exceeded"},
		{"dataset", resourcequery.ErrDataset, 500, ""},
		{"unknown implementation fault", errors.New("unclassified implementation fault"), 500, ""},
		{"cancelled", context.Canceled, 503, "temporarily-unavailable"},
		{"deadline", context.DeadlineExceeded, 503, "temporarily-unavailable"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, q, validated, executed := queryASTTestAdapter(fmt.Errorf("%s: %w", sensitive, tc.err))
			request := queryASTTestRequest("POST", "/read/?readVersion=1&operation=list", queryASTListFixture)
			response := httptest.NewRecorder()
			q.serve(g, response, request)
			if response.Code != tc.status || *executed != 1 || *validated != 1 {
				t.Fatalf("status=%d execution=%d authority=%d body=%s", response.Code, *executed, *validated, response.Body.String())
			}
			if strings.Contains(response.Body.String(), sensitive) || strings.Contains(response.Body.String(), "unclassified implementation fault") || strings.Contains(response.Body.String(), `"result"`) {
				t.Fatalf("failure leaked implementation/success data: %s", response.Body.String())
			}
			if tc.status == 500 {
				if response.Body.Len() != 0 || response.Header().Get("Content-Type") != "" || response.Header().Get("Content-Length") != "0" {
					t.Fatalf("implementation fault must remain sanitized bodyless500: %v %s", response.Header(), response.Body.String())
				}
				return
			}
			var problem map[string]any
			if err := json.Unmarshal(response.Body.Bytes(), &problem); err != nil {
				t.Fatal(err)
			}
			if problem["code"] != tc.code || problem["status"] != float64(tc.status) || response.Header().Get("Content-Type") != "application/problem+json" {
				t.Fatalf("wrong failure class: %v", problem)
			}
			if problem["detail"] != nil || problem["result"] != nil {
				t.Fatalf("failure leaked detail or successful payload: %v", problem)
			}
			if tc.code == "resource-not-found" && problem["at"] != nil {
				t.Fatalf("unavailable Resource disclosed branch: %v", problem)
			}
			if tc.code == "capability-not-supported" && (problem["capability"] != queryASTCapability || problem["capabilityVersion"] != "1") {
				t.Fatalf("missing selected capability detail: %v", problem)
			}
		})
	}
}
