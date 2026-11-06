package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

// coreRequest sends one request with the given headers and returns the response
// with its whole body.
func coreRequest(t *testing.T, ts *testServer, method, path string, header http.Header) (*http.Response, []byte) {
	t.Helper()
	req, err := http.NewRequest(method, ts.base+path, nil)
	if err != nil {
		t.Fatal(err)
	}
	for name, values := range header {
		req.Header[name] = values
	}
	resp, err := ts.client.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return resp, body
}

// wantCoreProblem checks one refusal against Core's closed Problem table.
func wantCoreProblem(t *testing.T, resp *http.Response, body []byte, code string) {
	t.Helper()
	shape, ok := coreProblemShapes[code]
	if !ok {
		t.Fatalf("no shape for %q", code)
	}
	if resp.StatusCode != shape.status {
		t.Fatalf("status = %d, want %d; body %s", resp.StatusCode, shape.status, body)
	}
	if got := resp.Header.Get("Content-Type"); got != problemMediaType {
		t.Errorf("Content-Type = %q, want %q", got, problemMediaType)
	}
	var problem map[string]any
	if err := json.Unmarshal(body, &problem); err != nil {
		t.Fatalf("decode %q: %v", body, err)
	}
	if problem["code"] != code || problem["retry"] != shape.retry || problem["type"] != bdpProblemBase+shape.family {
		t.Errorf("problem = %v, want code %s, retry %s, family %s", problem, code, shape.retry, shape.family)
	}
}

func wantServiceDesc(t *testing.T, ts *testServer, resp *http.Response) {
	t.Helper()
	want := "<" + ts.base + `/bdp>; rel="service-desc"`
	if got := resp.Header.Get("Link"); got != want {
		t.Errorf("Link = %q, want %q", got, want)
	}
	if got := resp.Header.Get("Cache-Control"); got != "private, no-store" {
		t.Errorf("Cache-Control = %q", got)
	}
}

// TestCoreScopeRootCarriesServiceDesc: GET of the Scope URL must succeed and
// carry the discovery Link, with no body, and asks for no credential.
func TestCoreScopeRootCarriesServiceDesc(t *testing.T) {
	ts := newTestServer(t, Config{Auth: authTokenFile(t, "real-token")})
	for _, method := range []string{http.MethodGet, http.MethodHead} {
		resp, body := coreRequest(t, ts, method, "/", nil)
		if resp.StatusCode != http.StatusNoContent || len(body) != 0 {
			t.Errorf("%s / = %d with %d bytes, want an empty 204", method, resp.StatusCode, len(body))
		}
		wantServiceDesc(t, ts, resp)
	}
	resp, body := coreRequest(t, ts, http.MethodPost, "/", nil)
	if resp.StatusCode != http.StatusMethodNotAllowed || resp.Header.Get("Allow") != "GET, HEAD" || len(body) != 0 {
		t.Errorf("POST / = %d, Allow %q, %d bytes; want a bodyless 405", resp.StatusCode, resp.Header.Get("Allow"), len(body))
	}
}

// TestCoreDiscoveryDocument pins the closed document: Core plus Callables1, the
// Scope derived from the request, and no credential needed to read it.
func TestCoreDiscoveryDocument(t *testing.T) {
	ts := newTestServer(t, Config{Auth: authTokenFile(t, "real-token")})
	resp, body := coreRequest(t, ts, http.MethodGet, "/bdp", http.Header{"Accept": {bdpMediaType}})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /bdp = %d: %s", resp.StatusCode, body)
	}
	if got := resp.Header.Get("Content-Type"); got != bdpMediaType {
		t.Errorf("Content-Type = %q", got)
	}
	wantServiceDesc(t, ts, resp)

	var doc map[string]any
	if err := json.Unmarshal(body, &doc); err != nil {
		t.Fatal(err)
	}
	if doc["bdpVersion"] != "0.1" || doc["profile"] != "core" || doc["scope"] != ts.base+"/" {
		t.Errorf("discovery = %s", body)
	}
	if rev, _ := doc["specRevision"].(string); len(rev) != 40 || strings.ToLower(rev) != rev {
		t.Errorf("specRevision = %q, want a full lowercase commit", rev)
	}
	capabilities, _ := doc["capabilities"].(map[string]any)
	versions, _ := capabilities[callablesCapability].(map[string]any)
	detail, ok := versions["1"].(map[string]any)
	if !ok || len(detail) != 0 {
		t.Errorf("Callables advertisement = %v, want version 1 with empty detail", capabilities)
	}
	for _, closedOut := range []string{"operations", "beads", "links", "events", "historicalResolution", "aliases", "types"} {
		if _, present := doc[closedOut]; present {
			t.Errorf("discovery advertises %q, which this Scope does not serve", closedOut)
		}
	}

	resp, body = coreRequest(t, ts, http.MethodGet, "/bdp", http.Header{"Accept": {"text/plain"}})
	if resp.StatusCode != http.StatusNotAcceptable || len(body) != 0 {
		t.Errorf("Accept text/plain = %d with %d bytes, want a bodyless 406", resp.StatusCode, len(body))
	}
	resp, body = coreRequest(t, ts, http.MethodGet, "/bdp?x=1", nil)
	wantCoreProblem(t, resp, body, coreCodeInvalidParameter)
	resp, body = coreRequest(t, ts, http.MethodDelete, "/bdp", nil)
	if resp.StatusCode != http.StatusMethodNotAllowed || resp.Header.Get("Allow") != "GET, HEAD" || len(body) != 0 {
		t.Errorf("DELETE /bdp = %d, Allow %q", resp.StatusCode, resp.Header.Get("Allow"))
	}
	resp, _ = coreRequest(t, ts, http.MethodGet, "/bdp", http.Header{"If-None-Match": {"*"}})
	if resp.StatusCode != http.StatusNotModified {
		t.Errorf("If-None-Match: * = %d, want 304", resp.StatusCode)
	}
	wantServiceDesc(t, ts, resp)
	resp, _ = coreRequest(t, ts, http.MethodGet, "/bdp", http.Header{"If-Match": {`"x"`}})
	if resp.StatusCode != http.StatusPreconditionFailed {
		t.Errorf("If-Match of a tag on a representation with none = %d, want 412", resp.StatusCode)
	}
}
