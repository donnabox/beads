package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/httpapi/apigen"
)

// newBDPServer is a test server with both surfaces on, for a test about BDP.
func newBDPServer(t *testing.T, cfg Config, tune ...func(*Server)) *testServer {
	t.Helper()
	cfg.Protocol = ProtocolBoth
	return newTestServer(t, cfg, tune...)
}

func answerTo(t *testing.T, ts *testServer, method, path string) (int, apigen.Problem) {
	t.Helper()
	req, err := http.NewRequest(method, ts.base+path, strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := ts.client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	var problem apigen.Problem
	_ = json.Unmarshal(raw, &problem)
	return resp.StatusCode, problem
}

// The default is what this server answered before BDP: not one /bdp path, not the
// Scope root and not a Resource address is any different from an unrouted path.
func TestProtocolV0IsAServerWithoutBDP(t *testing.T) {
	for name, cfg := range map[string]Config{"zero": {}, "v0": {Protocol: ProtocolV0}} {
		t.Run(name, func(t *testing.T) {
			ts := newTestServer(t, cfg)
			for _, c := range [][2]string{
				{http.MethodGet, "/"},
				{http.MethodGet, "/bdp"},
				{http.MethodGet, "/bdp/beads"},
				{http.MethodGet, "/bdp/links"},
				{http.MethodGet, "/bdp/call"},
				{http.MethodPost, "/bdp/call/issue/show"},
				{http.MethodGet, "/sw-1"},
				{http.MethodPost, "/bdp"},
				{http.MethodGet, "/bdp/nonsense"},
			} {
				status, problem := answerTo(t, ts, c[0], c[1])
				if status != http.StatusNotFound || problem.Code != string(CodeNotFound) || detailOf(problem) != "no such route on this server" {
					t.Errorf("%s %s = %d %q %q, want v0's unrouted answer", c[0], c[1], status, problem.Code, detailOf(problem))
				}
			}
			if resp := ts.get(t, "/v0/beads/context"); resp.StatusCode != http.StatusOK {
				t.Errorf("GET /v0/beads/context = %d", resp.StatusCode)
			}
		})
	}
}

func detailOf(problem apigen.Problem) string {
	if problem.Detail == nil {
		return ""
	}
	return *problem.Detail
}

// With v0 off, no route in the table answers on its own path. Each says why, and not
// with v0's not_found, which a client reads as "no such issue".
func TestProtocolBDPTakesEveryV0RouteOffTheWire(t *testing.T) {
	ts := newTestServer(t, Config{Protocol: ProtocolBDP})
	checked := 0
	for _, rt := range routeTable {
		if !isV0Route(rt) {
			continue
		}
		checked++
		path := strings.NewReplacer("{id}", "x-1", "{key}", "k", "{idAction}", "x-1").Replace(rt.pattern) + rt.customMethod
		status, problem := answerTo(t, ts, rt.method, path)
		if status != http.StatusNotFound || problem.Code != codeProtocolNotServed {
			t.Errorf("%s %s = %d %q, want 404 %s", rt.method, path, status, problem.Code, codeProtocolNotServed)
		}
		if !strings.Contains(detailOf(problem), "--protocol bdp") {
			t.Errorf("%s %s: the detail does not say why: %q", rt.method, path, detailOf(problem))
		}
	}
	if checked == 0 {
		t.Fatal("no v0 route in the table")
	}
	for _, path := range []string{"/v0", "/v0/", "/v0/nonsense"} {
		if status, problem := answerTo(t, ts, http.MethodGet, path); status != http.StatusNotFound || problem.Code != codeProtocolNotServed {
			t.Errorf("GET %s = %d %q", path, status, problem.Code)
		}
	}
	if _, problem := answerTo(t, ts, http.MethodGet, "/v0/beads/events:watch"); !strings.Contains(detailOf(problem), "no Callable") {
		t.Errorf("the stream's refusal says %q", detailOf(problem))
	}
}

// What a BDP-only server still answers: liveness, Core, and the Callables, the
// handshake among them, on the rows the mux no longer routes.
func TestProtocolBDPStillServesTheCallablesAndLiveness(t *testing.T) {
	ts := newTestServer(t, Config{Protocol: ProtocolBDP})
	for _, path := range []string{"/healthz", "/bdp", "/bdp/call"} {
		if resp := ts.get(t, path); resp.StatusCode != http.StatusOK {
			t.Errorf("GET %s = %d", path, resp.StatusCode)
		}
	}
	if resp := ts.get(t, "/"); resp.StatusCode != http.StatusNoContent {
		t.Errorf("GET / = %d", resp.StatusCode)
	}
	req, err := http.NewRequest(http.MethodPost, ts.base+"/bdp/call/issue/context", strings.NewReader(`{"arguments":{}}`))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", bdpMediaType)
	req.Header.Set("Accept", bdpMediaType)
	resp, err := ts.client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var answer struct {
		Result apigen.ContextResponse `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&answer); err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("issue/context = %d", resp.StatusCode)
	}
	if slices.Contains(answer.Result.Capabilities, "events.watch") {
		t.Error("a server with no /v0 advertises the stream, which only /v0 has")
	}
	if !slices.Contains(answer.Result.Capabilities, "events.list") {
		t.Error("a server with BDP on dropped events.list, which has a Callable")
	}
}

func TestProtocolBothServesBoth(t *testing.T) {
	ts := newTestServer(t, Config{Protocol: ProtocolBoth})
	for _, path := range []string{"/healthz", "/v0/beads/context", "/bdp", "/bdp/call"} {
		if resp := ts.get(t, path); resp.StatusCode != http.StatusOK {
			t.Errorf("GET %s = %d", path, resp.StatusCode)
		}
	}
	var ctx apigen.ContextResponse
	resp := ts.get(t, "/v0/beads/context")
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&ctx); err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(ctx.Capabilities, "events.watch") {
		t.Error("a server with /v0 on dropped the stream's capability")
	}
}

func TestAnUnknownProtocolIsRefused(t *testing.T) {
	if _, err := ParseProtocol("smoke"); err == nil {
		t.Error("ParseProtocol took smoke")
	}
	if _, err := ParseProtocol(""); err == nil {
		t.Error(`ParseProtocol took "": the flag defaults to v0, so an empty value was typed`)
	}
	if _, err := Listen(Config{Addr: "127.0.0.1:0", Provider: &fakeProvider{}, Protocol: "smoke"}); err == nil {
		t.Error("Listen took an unknown Protocol")
	}
}
