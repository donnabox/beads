//go:build cgo

package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/httpapi/bdpwire"
	"github.com/steveyegge/beads/internal/httpapi/graphread"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// This exercises the actual listener, security gates and graph store. The CLI
// and independent public client are separately exercised by graph-bdp-read-smoke.
func TestGraphReadHTTPAuthorityAndSecurity(t *testing.T) {
	port, err := strconv.Atoi(os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT"))
	if err != nil || port == 0 {
		t.Skip("ordinary shared-server Dolt port not configured")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	workspace, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	const scope = "https://authority.example/read/"
	options := graphstore.Options{Backend: "server", ServerHost: "127.0.0.1", ServerPort: port, ServerUser: "root", Database: fmt.Sprintf("graph_http_%d", time.Now().UnixNano()), Branch: "main", DataDir: filepath.Join(workspace, "dolt"), Binding: graphstore.Binding{WorkspaceID: workspace, ScopeURL: scope, AuthorityID: "0123456789abcdef0123456789abcdef", SchemaVersion: graphstore.SchemaVersion}}
	if err := graphstore.Init(ctx, options); err != nil {
		t.Fatal(err)
	}
	store, err := graphstore.OpenExisting(ctx, options)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	for _, path := range []string{"beads/plan", "beads/other"} {
		if _, err := store.Create(ctx, graphstore.CreateRequest{Path: path, Title: "記憶"}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := store.AddInformationalLink(ctx, graphstore.LinkCreateRequest{Path: "links/context", SourcePath: "beads/plan", TargetPath: "beads/other", UnconditionalSource: true, Properties: map[string]any{"note": "initial"}}); err != nil {
		t.Fatal(err)
	}
	graph, err := NewGraphRead(graphread.New(store), scope)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(graph.Close)
	tokenFile := filepath.Join(workspace, "tokens")
	if err := os.WriteFile(tokenFile, []byte("first-token\n"), 0600); err != nil {
		t.Fatal(err)
	}
	auth, err := NewTokenFileAuth(tokenFile)
	if err != nil {
		t.Fatal(err)
	}
	auth.reloadGate = time.Nanosecond // exercise existing reload policy without a sleep
	srv := newTestServer(t, Config{GraphRead: graph, Auth: auth, AllowedHosts: []string{"transport.example"}})
	request := func(method, path, token, host string, headers http.Header) (*http.Response, []byte) {
		t.Helper()
		req, err := http.NewRequestWithContext(ctx, method, srv.base+path, nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header = headers.Clone()
		if req.Header == nil {
			req.Header = http.Header{}
		}
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		if host != "" {
			req.Host = host
		}
		response, err := srv.client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(response.Body)
		_ = response.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		if response.Header.Get("Cache-Control") != "private, no-store" {
			t.Fatal("authorization-dependent response is cacheable")
		}
		return response, body
	}
	assertProblem := func(response *http.Response, body []byte, code bdpwire.ReadProblemCode) {
		t.Helper()
		var problem bdpwire.ReadProblem
		if response.StatusCode != code.Status() || bdpwire.Unmarshal(body, &problem) != nil || problem.Code != code {
			t.Fatalf("want %s, got %d %s", code, response.StatusCode, body)
		}
	}
	response, body := request("GET", "/read/beads/plan", "", "", http.Header{"If-None-Match": {"*"}})
	assertProblem(response, body, bdpwire.CodeUnauthenticated)
	response, body = request("GET", "/read/beads/plan", "first-token", "evil.example", nil)
	assertProblem(response, body, bdpwire.CodeMalformedRequest)
	response, body = request("GET", "/read/beads/plan", "first-token", "transport.example", nil)
	var bead bdpwire.BeadRecord
	if response.StatusCode != 200 || bdpwire.Unmarshal(body, &bead) != nil || bead.ID != scope+"beads/plan" {
		t.Fatalf("transport alias reassigned identity: %d %s", response.StatusCode, body)
	}
	getLength := len(body)
	response, body = request("HEAD", "/read/beads/plan", "first-token", "", nil)
	if response.StatusCode != 200 || len(body) != 0 || response.ContentLength != int64(getLength) {
		t.Fatalf("Unicode HEAD: %d %d %d", response.StatusCode, len(body), response.ContentLength)
	}
	for _, test := range []struct {
		path    string
		headers http.Header
		code    bdpwire.ReadProblemCode
	}{
		{"/read/beads/missing", http.Header{"Accept": {"image/png"}, "If-None-Match": {"*"}}, bdpwire.CodeResourceNotFound},
		{"/read/beads/?cursor=unknown", http.Header{"If-None-Match": {"*"}}, bdpwire.CodeCursorExpired},
		{"/read/beads/?limit=1&limit=2", nil, bdpwire.CodeInvalidParameter},
		{"/read/beads/?bad=%zz", nil, bdpwire.CodeInvalidParameter},
		{"/read/beads/plan?revision=old", nil, bdpwire.CodeInvalidParameter},
		{"/v0/beads/context", nil, bdpwire.CodeResourceNotFound},
	} {
		response, body = request("GET", test.path, "first-token", "", test.headers)
		assertProblem(response, body, test.code)
	}
	response, body = request("GET", "/read/beads/?limit=1", "first-token", "", nil)
	var first bdpwire.BeadCollection
	if response.StatusCode != 200 || bdpwire.Unmarshal(body, &first) != nil || first.Next == nil {
		t.Fatalf("first page: %d %s", response.StatusCode, body)
	}
	next, _ := url.Parse(*first.Next)
	// Atomic token rotation changes credential acceptance, not the full-workspace
	// authorization projection. Revoked credentials fail before cursor/conditions.
	newToken := filepath.Join(workspace, "tokens-new")
	if err := os.WriteFile(newToken, []byte("second-token-longer\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(newToken, tokenFile); err != nil {
		t.Fatal(err)
	}
	response, body = request("GET", next.RequestURI(), "first-token", "", http.Header{"If-None-Match": {"*"}})
	assertProblem(response, body, bdpwire.CodeUnauthenticated)
	response, body = request("GET", next.RequestURI(), "second-token-longer", "", nil)
	if response.StatusCode != 200 {
		t.Fatalf("accepted rotated token lost same projection: %d %s", response.StatusCode, body)
	}
	// Every aggregate must close over a single inventory even while writes race
	// the HTTP request. The owned and incident copy of a Link must agree.
	writerCtx, stopWriter := context.WithCancel(ctx)
	writerDone := make(chan error, 1)
	go func() {
		for i := 0; ; i++ {
			_, err := store.UpdateLink(writerCtx, graphstore.LinkUpdateRequest{Path: "links/context", Unconditional: true, UnconditionalSource: true, Properties: map[string]any{"note": strconv.Itoa(i)}})
			if err != nil {
				if writerCtx.Err() != nil {
					err = nil
				}
				writerDone <- err
				return
			}
		}
	}()
	stopAndWait := sync.OnceFunc(func() {
		stopWriter()
		if err := <-writerDone; err != nil {
			t.Errorf("concurrent writer: %v", err)
		}
	})
	defer stopAndWait()
	for range 12 {
		response, body = request("GET", "/read/beads/plan?include=links", "second-token-longer", "", nil)
		var aggregate bdpwire.BeadRecord
		if response.StatusCode != 200 || bdpwire.Unmarshal(body, &aggregate) != nil || aggregate.Links == nil || len(aggregate.Links.Items) != 1 {
			t.Fatalf("aggregate: %d %s", response.StatusCode, body)
		}
		owned := aggregate.OwnedLinks[graphstore.RelatedTypeURL(scope)]
		if len(owned) != 1 || owned[0].Revision != aggregate.Links.Items[0].Revision || string(owned[0].Properties["note"]) != string(aggregate.Links.Items[0].Properties["note"]) {
			t.Fatal("aggregate mixed two storage snapshots")
		}
	}
	stopAndWait()
	// Retained bytes never substitute for a currently usable authority. Even a
	// condition that would otherwise return 304 must fail after storage closes.
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	response, body = request("GET", next.RequestURI(), "second-token-longer", "", http.Header{"If-None-Match": {"*"}})
	if response.StatusCode != http.StatusInternalServerError || len(body) != 0 {
		t.Fatalf("closed authority served retained page: %d %s", response.StatusCode, body)
	}
	if strings.Contains(srv.stderr.String(), "first-token") || strings.Contains(srv.stderr.String(), "second-token-longer") {
		t.Fatal("credential leaked to request logs")
	}
	if !strings.Contains(srv.stderr.String(), "db=graph-read") || strings.Contains(srv.stderr.String(), "capabilities=issue") {
		t.Fatal("startup reported a legacy source")
	}
	encoded, _ := json.Marshal(map[string]any{"checks": "real authority, Host, auth rotation, query/condition precedence, Unicode HEAD, aggregate concurrent snapshot", "passed": true})
	t.Logf("HTTP_RECEIPT %s", encoded)
}

func TestGraphReadRejectsMixedSources(t *testing.T) {
	for _, cfg := range []Config{
		{GraphRead: &GraphRead{}},
		{GraphRead: &GraphRead{}, Provider: &fakeProvider{}},
		{GraphRead: &GraphRead{}, EventsJournalEnabled: true},
	} {
		if _, err := Listen(cfg); err == nil {
			t.Fatal("invalid graph source bound a listener")
		}
	}
}
