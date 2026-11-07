package httpapi

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/issueops"
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
	ts := newBDPServer(t, Config{Auth: authTokenFile(t, "real-token")})
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
	ts := newBDPServer(t, Config{Auth: authTokenFile(t, "real-token")})
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

// coreStore is the store under a Core read: the four roles the surface uses, over
// one set of rows.
type coreStore struct {
	issueops.WorkspaceConfig
	mu       sync.Mutex
	issues   []*types.Issue
	edges    []*types.Dependency
	settings map[string]string
	lists    []issueops.ListRequest
	getErr   error
}

func (c *coreStore) Ready(context.Context, issueops.ReadyRequest) (issueops.IssuePage, error) {
	return issueops.IssuePage{}, nil
}

func (c *coreStore) Get(context.Context, issueops.GetRequest) (*issueops.IssueDetails, error) {
	return nil, errors.New("Core reads do not use the detail view")
}

// List answers in a deliberately non-canonical order when asked for id order, so
// a server that trusts it fails the order test.
func (c *coreStore) List(_ context.Context, req issueops.ListRequest) (issueops.IssuePage, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lists = append(c.lists, req)
	var page issueops.IssuePage
	for i := len(c.issues) - 1; i >= 0; i-- {
		issue := c.issues[i]
		if req.IssueType == "" || req.IssueType == string(issue.IssueType) {
			page.Items = append(page.Items, &types.IssueWithCounts{Issue: &types.Issue{ID: issue.ID}})
		}
	}
	return page, nil
}

func (c *coreStore) GetMany(_ context.Context, req issueops.GetManyRequest) (issueops.GetManyResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := issueops.GetManyResult{Issues: []*types.Issue{}, Missing: []string{}}
	if c.getErr != nil {
		return result, c.getErr
	}
	for _, id := range req.IDs {
		index := slices.IndexFunc(c.issues, func(issue *types.Issue) bool { return issue.ID == id })
		if index < 0 {
			result.Missing = append(result.Missing, id)
			continue
		}
		clone := *c.issues[index]
		result.Issues = append(result.Issues, &clone)
	}
	return result, nil
}

func (c *coreStore) ReadEdges(_ context.Context, req issueops.EdgeReadRequest) (issueops.EdgeReadResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	var result issueops.EdgeReadResult
	for _, id := range req.IDs {
		anchor := issueops.AnchorEdges{ID: id, Edges: []*types.Dependency{}}
		for _, edge := range c.edges {
			if edge.IssueID == id {
				anchor.Edges = append(anchor.Edges, edge)
			}
		}
		result.Anchors = append(result.Anchors, anchor)
	}
	return result, nil
}

func (c *coreStore) Related(_ context.Context, req issueops.RelatedRequest) ([]*issueops.RelatedIssue, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	related := []*issueops.RelatedIssue{}
	for _, edge := range c.edges {
		if req.Direction == issueops.RelationIn && edge.DependsOnID == req.ID {
			related = append(related, &issueops.RelatedIssue{Issue: types.Issue{ID: edge.IssueID}})
		}
	}
	return related, nil
}

func (c *coreStore) GetSetting(_ context.Context, req issueops.GetSettingRequest) (issueops.SettingResult, error) {
	return issueops.SettingResult{Key: req.Key, Value: c.settings[req.Key]}, nil
}

func newCoreServer(t *testing.T, store *coreStore, cfg Config) *testServer {
	t.Helper()
	cfg.Reader, cfg.BatchGetter, cfg.EdgeReader, cfg.Relations, cfg.Settings = store, store, store, store, store
	return newBDPServer(t, rolesConfig(cfg))
}

var coreClock = time.Date(2026, 11, 6, 9, 30, 0, 0, time.UTC)

func coreIssue(id string, kind types.IssueType) *types.Issue {
	return &types.Issue{
		ID: id, Title: "title of " + id, Status: types.StatusOpen, Priority: 2, IssueType: kind,
		CreatedAt: coreClock, UpdatedAt: coreClock, RowVersion: 7,
	}
}

func coreEdge(source, target string, kind types.DependencyType) *types.Dependency {
	return &types.Dependency{IssueID: source, DependsOnID: target, Type: kind, CreatedAt: coreClock, CreatedBy: "sam"}
}

func coreGetJSON(t *testing.T, ts *testServer, path string) (*http.Response, map[string]any) {
	t.Helper()
	resp, body := coreRequest(t, ts, http.MethodGet, path, http.Header{"Accept": {bdpMediaType}})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET %s = %d: %s", path, resp.StatusCode, body)
	}
	if got := resp.Header.Get("Content-Type"); got != bdpMediaType {
		t.Errorf("GET %s Content-Type = %q", path, got)
	}
	var value map[string]any
	if err := json.Unmarshal(body, &value); err != nil {
		t.Fatalf("decode %s: %v", body, err)
	}
	return resp, value
}

func coreItemIDs(t *testing.T, page map[string]any) []string {
	t.Helper()
	items, ok := page["items"].([]any)
	if !ok {
		t.Fatalf("page has no items array: %v", page)
	}
	ids := make([]string, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.(map[string]any)["id"].(string))
	}
	return ids
}

// TestCoreBeadPointRead pins the record: identity, the Type table, what §3.4 keeps
// out of properties, metadata beside them, and the read metadata Core requires.
func TestCoreBeadPointRead(t *testing.T) {
	issue := coreIssue("sw-1", types.TypeBug)
	issue.Labels = []string{"zeta", "alpha"}
	issue.Metadata = json.RawMessage(`{"big":9007199254740993,"nested":{"a":1}}`)
	lease := coreClock.Add(time.Hour)
	issue.LeaseExpiresAt, issue.HeartbeatAt, issue.LeaseGrantedNode, issue.IsBlocked = &lease, &lease, "node", true
	due := time.Date(2026, 12, 1, 8, 0, 0, 0, time.FixedZone("x", 3600))
	issue.DueAt = &due
	store := &coreStore{issues: []*types.Issue{issue}}
	ts := newCoreServer(t, store, Config{})

	resp, record := coreGetJSON(t, ts, "/sw-1")
	wantServiceDesc(t, ts, resp)
	if got := resp.Header.Get("BDP-Resource-Kind"); got != "bead" {
		t.Errorf("BDP-Resource-Kind = %q", got)
	}
	if resp.Header.Get("BDP-Collection-Kind") != "" {
		t.Error("a point read carries BDP-Collection-Kind")
	}
	if record["id"] != ts.base+"/sw-1" || record["type"] != coreBeadTypeNamespace+"bug" {
		t.Errorf("record = %v", record)
	}
	revision, _ := record["revision"].(string)
	if revision == "" || resp.Header.Get("ETag") != `"`+revision+`"` {
		t.Errorf("revision %q, ETag %q", revision, resp.Header.Get("ETag"))
	}
	properties := record["properties"].(map[string]any)
	for _, absent := range []string{"id", "issue_type", "metadata", "lease_expires_at", "heartbeat_at", "lease_granted_node", "is_blocked", "dependencies", "comments"} {
		if _, present := properties[absent]; present {
			t.Errorf("properties carry %q", absent)
		}
	}
	if properties["title"] != "title of sw-1" || properties["status"] != "open" || properties["priority"] != float64(2) {
		t.Errorf("properties = %v", properties)
	}
	if properties["due_at"] != "2026-12-01T07:00:00Z" {
		t.Errorf("due_at = %v, want the instant in UTC", properties["due_at"])
	}
	if !slices.Equal(properties["labels"].([]any), []any{"alpha", "zeta"}) {
		t.Errorf("labels = %v", properties["labels"])
	}
	_, raw := coreRequest(t, ts, http.MethodGet, "/sw-1", nil)
	if !strings.Contains(string(raw), `"big":9007199254740993`) {
		t.Errorf("metadata lost a number's literal: %s", raw)
	}

	head, headBody := coreRequest(t, ts, http.MethodHead, "/sw-1", nil)
	if head.StatusCode != http.StatusOK || len(headBody) != 0 || head.Header.Get("BDP-Resource-Kind") != "bead" {
		t.Errorf("HEAD = %d, %d bytes", head.StatusCode, len(headBody))
	}

	view, viewed := coreGetJSON(t, ts, "/sw-1?view=properties")
	if viewed["title"] != "title of sw-1" || viewed["properties"] != nil {
		t.Errorf("view=properties = %v", viewed)
	}
	if view.Header.Get("ETag") != resp.Header.Get("ETag") || view.Header.Get("BDP-Resource-Kind") != "bead" {
		t.Errorf("view=properties headers: %v", view.Header)
	}

	etag := resp.Header.Get("ETag")
	cached, cachedBody := coreRequest(t, ts, http.MethodGet, "/sw-1", http.Header{"If-None-Match": {etag}})
	if cached.StatusCode != http.StatusNotModified || len(cachedBody) != 0 {
		t.Errorf("If-None-Match = %d, %d bytes", cached.StatusCode, len(cachedBody))
	}
	wantServiceDesc(t, ts, cached)
	if cached.Header.Get("BDP-Resource-Kind") != "bead" || cached.Header.Get("ETag") != etag {
		t.Errorf("304 headers: %v", cached.Header)
	}
	stale, _ := coreRequest(t, ts, http.MethodGet, "/sw-1", http.Header{"If-Match": {`"other"`}})
	if stale.StatusCode != http.StatusPreconditionFailed {
		t.Errorf("stale If-Match = %d", stale.StatusCode)
	}
	weak, _ := coreRequest(t, ts, http.MethodGet, "/sw-1", http.Header{"If-Match": {"W/" + etag}})
	if weak.StatusCode != http.StatusPreconditionFailed {
		t.Errorf("weak If-Match = %d, want 412: If-Match compares strongly", weak.StatusCode)
	}

	miss, missBody := coreRequest(t, ts, http.MethodGet, "/sw-404", http.Header{"If-None-Match": {"*"}})
	wantCoreProblem(t, miss, missBody, coreCodeNotFound)
	if miss.Header.Get("BDP-Resource-Kind") != "" {
		t.Error("a Problem carries BDP-Resource-Kind")
	}
	notAcceptable, _ := coreRequest(t, ts, http.MethodGet, "/sw-1", http.Header{"Accept": {"text/plain"}})
	if notAcceptable.StatusCode != http.StatusNotAcceptable {
		t.Errorf("Accept text/plain = %d", notAcceptable.StatusCode)
	}
	hidden, hiddenBody := coreRequest(t, ts, http.MethodGet, "/sw-404", http.Header{"Accept": {"text/plain"}})
	wantCoreProblem(t, hidden, hiddenBody, coreCodeNotFound)
}

// TestCoreRevisionCoversTheRecordAndNothingElse: Issue §7. The native row_lock
// does not move on a label-only write, and the revision must. It does not cover
// the lease fields, which change without changing the record.
func TestCoreRevisionCoversTheRecordAndNothingElse(t *testing.T) {
	issue := coreIssue("sw-1", types.TypeTask)
	store := &coreStore{issues: []*types.Issue{issue}}
	ts := newCoreServer(t, store, Config{})
	revision := func() string {
		t.Helper()
		_, record := coreGetJSON(t, ts, "/sw-1")
		return record["revision"].(string)
	}
	first := revision()
	if again := revision(); again != first {
		t.Fatalf("revision moved between two reads: %q, %q", first, again)
	}

	store.mu.Lock()
	beat := coreClock.Add(time.Minute)
	issue.HeartbeatAt, issue.LeaseExpiresAt = &beat, &beat
	store.mu.Unlock()
	if got := revision(); got != first {
		t.Errorf("a heartbeat moved the revision: %q to %q", first, got)
	}

	store.mu.Lock()
	issue.Labels = []string{"new"}
	store.mu.Unlock()
	labeled := revision()
	if labeled == first {
		t.Error("a label-only write left the revision where it was")
	}

	store.mu.Lock()
	issue.Labels = nil
	issue.RowVersion = 8
	store.mu.Unlock()
	if got := revision(); got == first || got == labeled {
		t.Errorf("a row_lock change left the revision at %q", got)
	}
}

// TestCoreBeadInventoryPages walks the inventory: canonical order whatever the
// store answered, the collection metadata, and a continuation that resizes.
func TestCoreBeadInventoryPages(t *testing.T) {
	store := &coreStore{}
	for _, id := range []string{"sw-c", "sw-a", "sw-e", "sw-b", "sw-d"} {
		store.issues = append(store.issues, coreIssue(id, types.TypeTask))
	}
	ts := newCoreServer(t, store, Config{})

	resp, page := coreGetJSON(t, ts, "/bdp/beads?limit=2")
	wantServiceDesc(t, ts, resp)
	if resp.Header.Get("BDP-Collection-Kind") != "bead" || resp.Header.Get("BDP-Resource-Kind") != "" {
		t.Errorf("collection headers: %v", resp.Header)
	}
	if got := coreItemIDs(t, page); !slices.Equal(got, []string{ts.base + "/sw-a", ts.base + "/sw-b"}) {
		t.Errorf("first page = %v", got)
	}
	next, _ := page["next"].(string)
	if !strings.HasPrefix(next, ts.base+"/bdp/beads?") {
		t.Fatalf("next = %q", next)
	}

	// Core lets a client replace the limit on an issued continuation.
	parsed, err := url.Parse(next)
	if err != nil {
		t.Fatal(err)
	}
	query := parsed.Query()
	query.Set("limit", "1")
	parsed.RawQuery = query.Encode()
	_, resized := coreGetJSON(t, ts, parsed.RequestURI())
	if got := coreItemIDs(t, resized); !slices.Equal(got, []string{ts.base + "/sw-c"}) {
		t.Errorf("resized page = %v", got)
	}

	var seen []string
	path := "/bdp/beads?limit=2"
	for pages := 0; path != ""; pages++ {
		if pages > 5 {
			t.Fatal("the walk did not end")
		}
		_, page := coreGetJSON(t, ts, path)
		seen = append(seen, coreItemIDs(t, page)...)
		path = ""
		if link, ok := page["next"].(string); ok {
			path = strings.TrimPrefix(link, ts.base)
		} else if page["next"] != nil {
			t.Fatalf("next = %v", page["next"])
		}
	}
	if !slices.IsSorted(seen) || len(seen) != 5 {
		t.Errorf("walk = %v", seen)
	}

	// The list role is asked for every row, in both planes.
	store.mu.Lock()
	req := store.lists[0]
	store.mu.Unlock()
	if !req.AllFlag || !req.IncludeTemplates || !req.IncludeGates || !req.IncludeInfra || !req.IncludeEphemeral {
		t.Errorf("the inventory hid something: %+v", req)
	}
	if req.Limit == nil || *req.Limit != 0 {
		t.Errorf("the id projection was limited: %+v", req.Limit)
	}

	for _, tc := range []struct{ name, path, code string }{
		{"garbage cursor", "/bdp/beads?cursor=@@", coreCodeInvalidParameter},
		{"unknown parameter", "/bdp/beads?status=open", coreCodeInvalidParameter},
		{"Link parameter", "/bdp/beads?source=sw-a", coreCodeInvalidParameter},
		{"repeated", "/bdp/beads?limit=1&limit=2", coreCodeInvalidParameter},
		{"zero limit", "/bdp/beads?limit=0", coreCodeInvalidParameter},
		{"negative limit", "/bdp/beads?limit=-1", coreCodeInvalidParameter},
		{"view on an inventory", "/bdp/beads?view=links", coreCodeInvalidParameter},
		{"over the maximum", "/bdp/beads?limit=1001", coreCodeLimitExceeded},
		{"unknown route", "/bdp/nothing", coreCodeNotFound},
	} {
		resp, body := coreRequest(t, ts, http.MethodGet, tc.path, nil)
		t.Run(tc.name, func(t *testing.T) { wantCoreProblem(t, resp, body, tc.code) })
	}

	// A continuation is bound to the predicates it was issued for.
	_, typed := coreGetJSON(t, ts, "/bdp/beads?limit=1&type="+url.QueryEscape(coreBeadTypeNamespace+"task"))
	issued, err := url.Parse(typed["next"].(string))
	if err != nil {
		t.Fatal(err)
	}
	foreign, body := coreRequest(t, ts, http.MethodGet, "/bdp/beads?cursor="+issued.Query().Get("cursor"), nil)
	wantCoreProblem(t, foreign, body, coreCodeInvalidParameter)
}

// TestCoreTypeSelection is Core's selection rule for a Type this Scope cannot
// hold: `type` is an exact comparison, so it is a true empty page, while
// conformsTo needs an interpretation this Scope has only for its fixed catalog.
func TestCoreTypeSelection(t *testing.T) {
	store := &coreStore{issues: []*types.Issue{
		coreIssue("sw-1", types.TypeTask), coreIssue("sw-2", types.TypeBug), coreIssue("sw-3", "widget"),
	}, settings: map[string]string{"types.custom": "widget,other"}}
	ts := newCoreServer(t, store, Config{})
	local := ts.base + "/types/issue/widget"
	memory := "https://github.com/gastownhall/bdp/types/memory/memory"

	for _, tc := range []struct {
		name, query string
		want        []string
	}{
		{"exact real", "type=" + coreBeadTypeNamespace + "task", []string{"sw-1"}},
		{"exact unsupported Type is empty", "type=" + memory, nil},
		{"exact abstract base admits no affiliation", "type=" + coreBeadFamily, nil},
		{"exact custom", "type=" + local, []string{"sw-3"}},
		{"family", "conformsTo=" + coreBeadFamily, []string{"sw-1", "sw-2", "sw-3"}},
		{"family child is reflexive", "conformsTo=" + coreBeadTypeNamespace + "bug", []string{"sw-2"}},
		{"custom is reflexive", "conformsTo=" + local, []string{"sw-3"}},
		{"both predicates must hold", "conformsTo=" + coreBeadFamily + "&type=" + coreBeadTypeNamespace + "bug", []string{"sw-2"}},
		{"a Link family holds no Bead", "conformsTo=" + coreLinkFamily, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, page := coreGetJSON(t, ts, "/bdp/beads?"+tc.query)
			want := []string{}
			for _, id := range tc.want {
				want = append(want, ts.base+"/"+id)
			}
			if got := coreItemIDs(t, page); !slices.Equal(got, want) {
				t.Errorf("got %v, want %v", got, want)
			}
			if next, present := page["next"]; !present || next != nil {
				t.Errorf("next = %v, want an explicit null", next)
			}
		})
	}
	for _, tc := range []struct{ name, query string }{
		{"conformsTo outside the catalog", "conformsTo=" + memory},
		{"conformsTo an unconfigured local name", "conformsTo=" + ts.base + "/types/issue/nope"},
		{"relative type", "type=task"},
		{"relative conformsTo", "conformsTo=issue"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resp, body := coreRequest(t, ts, http.MethodGet, "/bdp/beads?"+tc.query, nil)
			wantCoreProblem(t, resp, body, coreCodeInvalidParameter)
		})
	}
}

// TestCoreUnservableClassificationRefuses: Issue §2. A classification with no
// identity is not served and gets none invented, so the read refuses rather than
// mapping it to a base or dropping the row from a page.
func TestCoreUnservableClassificationRefuses(t *testing.T) {
	store := &coreStore{issues: []*types.Issue{coreIssue("sw-1", types.TypeTask), coreIssue("sw-2", "zork")}}
	ts := newCoreServer(t, store, Config{})
	for _, path := range []string{"/sw-2", "/bdp/beads"} {
		resp, body := coreRequest(t, ts, http.MethodGet, path, nil)
		wantCoreProblem(t, resp, body, coreCodeUnavailable)
		if !strings.Contains(string(body), "zork") {
			t.Errorf("GET %s does not name the classification: %s", path, body)
		}
	}
	coreGetJSON(t, ts, "/sw-1")

	store.mu.Lock()
	store.settings = map[string]string{"types.custom": "zork"}
	store.mu.Unlock()
	_, record := coreGetJSON(t, ts, "/sw-2")
	if record["type"] != ts.base+"/types/issue/zork" {
		t.Errorf("type = %v", record["type"])
	}
}

// TestCoreSelectorOnInventories runs Selectors through the HTTP surface.
func TestCoreSelectorOnInventories(t *testing.T) {
	open := coreIssue("sw-1", types.TypeTask)
	closed := coreIssue("sw-2", types.TypeBug)
	closed.Status = types.StatusClosed
	closed.Priority = 0
	store := &coreStore{issues: []*types.Issue{open, closed}}
	ts := newCoreServer(t, store, Config{})

	get := func(selector string) []string {
		t.Helper()
		_, page := coreGetJSON(t, ts, "/bdp/beads?selector="+url.QueryEscape(selector))
		return coreItemIDs(t, page)
	}
	want := []string{ts.base + "/sw-2"}
	for _, selector := range []string{
		`$[?@.properties.status == "closed"]`,
		`$[?@.properties.priority < 1]`,
		`$[?@.id == "` + ts.base + `/sw-2"]`,
		`$[?@.type == "` + coreBeadTypeNamespace + `bug"]`,
	} {
		if got := get(selector); !slices.Equal(got, want) {
			t.Errorf("%s = %v", selector, got)
		}
	}
	for _, tc := range []struct{ selector, code string }{
		{`status == "closed"`, coreCodeInvalidParameter},
		{`$..*`, coreCodeInvalidParameter},
		{`$[?match(@.properties.title, "a")]`, coreCodeInvalidParameter},
		{`$[?@.properties.a == @.properties.b]`, coreCodeInvalidParameter},
		{`$[?@.properties.a == 9007199254740993]`, coreCodeInvalidParameter},
		{`$[?@.a == "` + strings.Repeat("x", coreSelectorBytes) + `"]`, coreCodeLimitExceeded},
		{`$[?` + strings.Repeat("(", coreSelectorDepth+1) + `@.a` + strings.Repeat(")", coreSelectorDepth+1) + `]`, coreCodeLimitExceeded},
	} {
		resp, body := coreRequest(t, ts, http.MethodGet, "/bdp/beads?selector="+url.QueryEscape(tc.selector), nil)
		t.Run(tc.selector[:min(len(tc.selector), 30)], func(t *testing.T) { wantCoreProblem(t, resp, body, tc.code) })
	}
}

func TestCoreSelectorEvaluation(t *testing.T) {
	candidate := map[string]any{
		"id":         "https://s/sw-1",
		"properties": map[string]any{"n": json.Number("1.0"), "s": "b", "flag": true, "none": nil, "list": []any{"x", "y"}, "o": map[string]any{"k": "v"}},
		"metadata":   map[string]any{"metadata": "m"},
	}
	for selector, want := range map[string]bool{
		`$[?@.properties.n == 1]`:             true,
		`$[?@.properties.n == 1e0]`:           true,
		`$[?@.properties.n != 1]`:             false,
		`$[?1 <= @.properties.n]`:             true,
		`$[?@.properties.n > 0.5]`:            true,
		`$[?@.properties.s < "c"]`:            true,
		`$[?@.properties.s < 1]`:              false,
		`$[?@.properties.flag == true]`:       true,
		`$[?@.properties.flag <= true]`:       true,
		`$[?@.properties.flag < true]`:        false,
		`$[?@.properties.none == null]`:       true,
		`$[?@.properties.missing == null]`:    false,
		`$[?@.properties.missing != "x"]`:     true,
		`$[?@.properties.missing]`:            false,
		`$[?@.properties.none]`:               true,
		`$[?!@.properties.missing]`:           true,
		`$[?@.properties.list[1] == "y"]`:     true,
		`$[?@.properties.list[-1] == "y"]`:    true,
		`$[?@.properties.list[5]]`:            false,
		`$[?@['properties']["s"] == 'b']`:     true,
		`$[?@.properties.o == "v"]`:           false,
		`$[?@.metadata.metadata == "m"]`:      true,
		`$[?@.properties.s=="b"&&@.nope]`:     false,
		`$[?@.nope || @.properties.s == "b"]`: true,
		`$[?!(@.nope || @.properties.flag)]`:  false,
		` $ [ ? @.properties.s == "b" ] `:     true,
		`$[?@.revision]`:                      false,
	} {
		parsed, err := parseCoreSelector(selector)
		if err != nil {
			t.Errorf("%s: %v", selector, err)
			continue
		}
		if got := parsed.matches(candidate); got != want {
			t.Errorf("%s = %v, want %v", selector, got, want)
		}
	}
	for _, selector := range []string{
		"", "$", "$[?]", "$[?@.a", "$[?@.a]x", "$[*]", "$[?@.*]", "$[?@..a]", "$[?@.a[0:1]]",
		"$[?$.a]", "$[?1 == 1]", `$[?@.a == {"b":1}]`, "$[?@.a == [1]]", "$[?length(@.a) > 1]",
		`$[?@.a == "\ud800"]`, `$[?@.a == "\q"]`, "$[?@.a = 1]", "$[?@.a == 01]", "$[?@.a == 1e999]",
	} {
		if _, err := parseCoreSelector(selector); err == nil {
			t.Errorf("%q parsed, want a refusal", selector)
		}
	}
}

// TestCoreLinks covers the Link inventory, its predicates, a point read and the
// target spellings of Issue §4.
func TestCoreLinks(t *testing.T) {
	a, b, c := coreIssue("sw-a", types.TypeTask), coreIssue("sw-b", types.TypeTask), coreIssue("sw-c", types.TypeEpic)
	blocks := coreEdge("sw-a", "sw-b", types.DepBlocks)
	blocks.Metadata = `{"weight":2}`
	blocks.ThreadID = "t-1"
	parent := coreEdge("sw-a", "sw-c", types.DepParentChild)
	parent.Metadata = "not an object"
	external := coreEdge("sw-b", "external:proj:cap", types.DepBlocks)
	foreign := coreEdge("sw-b", "other-1", types.DepRelated)
	custom := coreEdge("sw-c", "sw-b", "my-kind")
	store := &coreStore{
		issues: []*types.Issue{a, b, c},
		edges:  []*types.Dependency{blocks, parent, external, foreign, custom},
	}
	ts := newCoreServer(t, store, Config{})

	resp, page := coreGetJSON(t, ts, "/bdp/links")
	wantServiceDesc(t, ts, resp)
	if resp.Header.Get("BDP-Collection-Kind") != "link" || resp.Header.Get("BDP-Resource-Kind") != "" {
		t.Errorf("headers: %v", resp.Header)
	}
	ids := coreItemIDs(t, page)
	if len(ids) != 5 || !slices.IsSorted(ids) {
		t.Fatalf("ids = %v", ids)
	}
	byPair := map[string]map[string]any{}
	for _, item := range page["items"].([]any) {
		record := item.(map[string]any)
		byPair[record["source"].(string)+" "+record["target"].(string)] = record
	}
	base := ts.base + "/"
	record := byPair[base+"sw-a "+base+"sw-b"]
	if record == nil {
		t.Fatalf("no in-Scope record in %v", byPair)
	}
	if record["type"] != coreLinkTypeNamespace+"blocks" {
		t.Errorf("type = %v", record["type"])
	}
	properties := record["properties"].(map[string]any)
	if properties["created_by"] != "sam" || properties["thread_id"] != "t-1" || properties["created_at"] != "2026-11-06T09:30:00Z" {
		t.Errorf("properties = %v", properties)
	}
	if _, present := properties["dep_type"]; present {
		t.Error("properties carry dep_type")
	}
	if record["metadata"].(map[string]any)["weight"] != float64(2) {
		t.Errorf("metadata = %v", record["metadata"])
	}
	if meta := byPair[base+"sw-a "+base+"sw-c"]["metadata"].(map[string]any); len(meta) != 0 {
		t.Errorf("a non-object native metadata = %v, want {}", meta)
	}
	if byPair[base+"sw-b external:proj:cap"] == nil {
		t.Errorf("the external target is not its opaque URI: %v", byPair)
	}
	if byPair[base+"sw-b beads-id:other-1"] == nil {
		t.Errorf("the foreign target is not its opaque URI: %v", byPair)
	}
	if got := byPair[base+"sw-c "+base+"sw-b"]["type"]; got != base+"types/dependency/my-kind" {
		t.Errorf("custom kind type = %v", got)
	}

	// Point read, including what a Link refuses.
	linkPath := strings.TrimPrefix(record["id"].(string), ts.base)
	point, read := coreGetJSON(t, ts, linkPath)
	if point.Header.Get("BDP-Resource-Kind") != "link" || point.Header.Get("ETag") != `"`+record["revision"].(string)+`"` {
		t.Errorf("point headers: %v", point.Header)
	}
	wantServiceDesc(t, ts, point)
	if !reflect.DeepEqual(read, record) {
		t.Errorf("point read %v differs from its collection entry %v", read, record)
	}
	_, viewed := coreGetJSON(t, ts, linkPath+"?view=properties")
	if viewed["thread_id"] != "t-1" {
		t.Errorf("view=properties = %v", viewed)
	}
	for _, query := range []string{"?view=links", "?include=links"} {
		resp, body := coreRequest(t, ts, http.MethodGet, linkPath+query, nil)
		wantCoreProblem(t, resp, body, coreCodeInvalidParameter)
	}

	for _, tc := range []struct {
		query string
		want  int
	}{
		{"source=sw-a", 2},
		{"source=" + url.QueryEscape(base+"sw-a"), 2},
		{"target=sw-b", 2},
		{"endpoint=sw-b", 4},
		{"target=" + url.QueryEscape("external:proj:cap"), 1},
		{"source=" + url.QueryEscape("https://elsewhere.example/x"), 0},
		{"type=" + url.QueryEscape(coreLinkTypeNamespace+"blocks"), 2},
		{"type=" + url.QueryEscape(coreBeadTypeNamespace+"task"), 0},
		{"conformsTo=" + url.QueryEscape(coreLinkFamily), 5},
		{"conformsTo=" + url.QueryEscape(coreLinkTypeNamespace+"related"), 1},
		{"conformsTo=" + url.QueryEscape(coreBeadFamily), 0},
		{"selector=" + url.QueryEscape(`$[?@.properties.thread_id == "t-1"]`), 1},
		{"source=sw-a&type=" + url.QueryEscape(coreLinkTypeNamespace+"blocks"), 1},
	} {
		_, got := coreGetJSON(t, ts, "/bdp/links?"+tc.query)
		if ids := coreItemIDs(t, got); len(ids) != tc.want {
			t.Errorf("%s = %d items, want %d", tc.query, len(ids), tc.want)
		}
	}
	bad, body := coreRequest(t, ts, http.MethodGet, "/bdp/links?conformsTo="+url.QueryEscape("https://elsewhere.example/t"), nil)
	wantCoreProblem(t, bad, body, coreCodeInvalidParameter)

	var walked []string
	path := "/bdp/links?limit=2"
	for pages := 0; path != ""; pages++ {
		if pages > 5 {
			t.Fatal("the walk did not end")
		}
		_, got := coreGetJSON(t, ts, path)
		walked = append(walked, coreItemIDs(t, got)...)
		path = ""
		if link, ok := got["next"].(string); ok {
			path = strings.TrimPrefix(link, ts.base)
		}
	}
	if !slices.Equal(walked, ids) {
		t.Errorf("walk = %v, want %v", walked, ids)
	}
}

// TestCoreLinkIdentityAcrossRemoveAndAdd is IT-9's read half: a pair's next Link
// has a new identity and the old address is the uniform miss.
func TestCoreLinkIdentityAcrossRemoveAndAdd(t *testing.T) {
	first := coreEdge("sw-a", "sw-b", types.DepBlocks)
	store := &coreStore{
		issues: []*types.Issue{coreIssue("sw-a", types.TypeTask), coreIssue("sw-b", types.TypeTask)},
		edges:  []*types.Dependency{first},
	}
	ts := newCoreServer(t, store, Config{})
	_, page := coreGetJSON(t, ts, "/bdp/links")
	old := strings.TrimPrefix(coreItemIDs(t, page)[0], ts.base)
	coreGetJSON(t, ts, old)

	second := coreEdge("sw-a", "sw-b", types.DepBlocks)
	second.CreatedAt = coreClock.Add(time.Minute)
	store.mu.Lock()
	store.edges = []*types.Dependency{second}
	store.mu.Unlock()

	resp, body := coreRequest(t, ts, http.MethodGet, old, nil)
	wantCoreProblem(t, resp, body, coreCodeNotFound)
	_, page = coreGetJSON(t, ts, "/bdp/links")
	fresh := strings.TrimPrefix(coreItemIDs(t, page)[0], ts.base)
	if fresh == old {
		t.Errorf("the pair's second Link reused %q", old)
	}
	coreGetJSON(t, ts, fresh)
}

// TestCoreIncidentViews covers view=links and the include=links aggregate.
func TestCoreIncidentViews(t *testing.T) {
	store := &coreStore{
		issues: []*types.Issue{coreIssue("sw-a", types.TypeTask), coreIssue("sw-b", types.TypeTask), coreIssue("sw-c", types.TypeTask)},
		edges: []*types.Dependency{
			coreEdge("sw-a", "sw-b", types.DepBlocks),
			coreEdge("sw-b", "sw-c", types.DepBlocks),
			coreEdge("sw-b", "external:proj:cap", types.DepBlocks),
			coreEdge("sw-a", "sw-c", types.DepRelated),
		},
	}
	ts := newCoreServer(t, store, Config{})
	for _, tc := range []struct {
		query string
		want  int
	}{
		{"", 3}, {"&direction=both", 3}, {"&direction=outbound", 2}, {"&direction=inbound", 1},
		{"&direction=outbound&target=external:proj:cap", 1},
		{"&type=" + url.QueryEscape(coreLinkTypeNamespace+"related"), 0},
		{"&direction=outbound&target=sw-c", 1},
	} {
		resp, page := coreGetJSON(t, ts, "/sw-b?view=links"+tc.query)
		if ids := coreItemIDs(t, page); len(ids) != tc.want {
			t.Errorf("view=links%s = %d items, want %d", tc.query, len(ids), tc.want)
		}
		if resp.Header.Get("BDP-Collection-Kind") != "link" || resp.Header.Get("BDP-Resource-Kind") != "bead" {
			t.Errorf("headers: %v", resp.Header)
		}
	}

	_, first := coreGetJSON(t, ts, "/sw-b?view=links&limit=1&direction=outbound")
	next, _ := first["next"].(string)
	if !strings.Contains(next, "view=links") || !strings.Contains(next, "direction=outbound") {
		t.Fatalf("next = %q", next)
	}
	_, second := coreGetJSON(t, ts, strings.TrimPrefix(next, ts.base))
	if ids := coreItemIDs(t, second); len(ids) != 1 || ids[0] == coreItemIDs(t, first)[0] || second["next"] != nil {
		t.Errorf("second page = %v", second)
	}
	switched := strings.Replace(next, "direction=outbound", "direction=inbound", 1)
	resp, body := coreRequest(t, ts, http.MethodGet, strings.TrimPrefix(switched, ts.base), nil)
	wantCoreProblem(t, resp, body, coreCodeInvalidParameter)

	resp, aggregate := coreGetJSON(t, ts, "/sw-b?include=links&limit=1")
	if aggregate["revision"] == nil || resp.Header.Get("BDP-Collection-Kind") != "" || resp.Header.Get("BDP-Resource-Kind") != "bead" {
		t.Errorf("aggregate = %v, %v", aggregate, resp.Header)
	}
	links := aggregate["links"].(map[string]any)
	if ids := coreItemIDs(t, links); len(ids) != 1 {
		t.Errorf("embedded = %v", ids)
	}
	if next, _ := links["next"].(string); !strings.Contains(next, "view=links") {
		t.Errorf("embedded next = %q, want the Link view", next)
	}
	if resp.Header.Get("ETag") == `"`+aggregate["revision"].(string)+`"` {
		t.Error("the aggregate's entity tag is the Bead's revision")
	}
	_, plain := coreGetJSON(t, ts, "/sw-b")
	if _, present := plain["links"]; present {
		t.Error("a default read embeds the aggregate")
	}

	for _, query := range []string{
		"?view=events", "?view=nope", "?include=nope", "?revision=r1", "?view=versions", "?nope=1", "?direction=both",
		"?view=properties&limit=1", "?view=links&direction=sideways", "?include=links&view=links",
		"?include=links&cursor=x", "?include=links&type=https://x.example/t", "?view=links&cursor=@@",
		"?view=properties&view=properties",
	} {
		resp, body := coreRequest(t, ts, http.MethodGet, "/sw-b"+query, nil)
		t.Run(query, func(t *testing.T) { wantCoreProblem(t, resp, body, coreCodeInvalidParameter) })
	}
}

// TestCoreAuthAndRouting: resources and inventories take v0's token, the Scope
// root and discovery do not, and nothing here shadows a route it sits beside.
func TestCoreAuthAndRouting(t *testing.T) {
	store := &coreStore{issues: []*types.Issue{coreIssue("sw-1", types.TypeTask)}}
	ts := newCoreServer(t, store, Config{Auth: authTokenFile(t, "real-token")})
	credential := http.Header{"Authorization": {"Bearer real-token"}}

	for _, path := range []string{"/sw-1", "/sw-404", "/bdp/beads", "/bdp/links", "/sw-1?view=properties"} {
		resp, body := coreRequest(t, ts, http.MethodGet, path, nil)
		wantCoreProblem(t, resp, body, coreCodeUnauthenticated)
		if resp.Header.Get("WWW-Authenticate") != "Bearer" {
			t.Errorf("GET %s has no challenge", path)
		}
		wrong, wrongBody := coreRequest(t, ts, http.MethodGet, path, http.Header{"Authorization": {"Bearer nope"}})
		wantCoreProblem(t, wrong, wrongBody, coreCodeUnauthenticated)
		if strings.Contains(string(wrongBody), "nope") {
			t.Errorf("GET %s echoed the credential", path)
		}
	}
	for _, path := range []string{"/sw-1", "/bdp/beads", "/bdp/links"} {
		if resp, body := coreRequest(t, ts, http.MethodGet, path, credential); resp.StatusCode != http.StatusOK {
			t.Errorf("GET %s with the token = %d: %s", path, resp.StatusCode, body)
		}
	}
	for _, path := range []string{"/", "/bdp", "/healthz"} {
		if resp, _ := coreRequest(t, ts, http.MethodGet, path, nil); resp.StatusCode >= 300 {
			t.Errorf("GET %s with no credential = %d", path, resp.StatusCode)
		}
	}

	// A v0 route keeps its own answer, and a v0 path that is not a route is
	// still the catch-all's uncredentialed 404 in v0's vocabulary.
	if resp, _ := coreRequest(t, ts, http.MethodGet, "/v0/beads/context", credential); resp.StatusCode != http.StatusOK {
		t.Errorf("GET /v0/beads/context = %d", resp.StatusCode)
	}
	if resp, _ := coreRequest(t, ts, http.MethodGet, "/v0/beads/context", nil); resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("GET /v0/beads/context with no credential = %d", resp.StatusCode)
	}
	for _, tc := range []struct{ method, path string }{
		{http.MethodGet, "/v0/nonsense"}, {http.MethodPost, "/v0/nonsense"}, {http.MethodGet, "/v0"}, {http.MethodPost, "/healthz"},
	} {
		resp, body := coreRequest(t, ts, tc.method, tc.path, nil)
		if resp.StatusCode != http.StatusNotFound || !strings.Contains(string(body), string(CodeNotFound)) {
			t.Errorf("%s %s = %d: %s", tc.method, tc.path, resp.StatusCode, body)
		}
	}
	if resp, _ := coreRequest(t, ts, http.MethodGet, "/bdp/call", nil); resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("GET /bdp/call with no credential = %d: Callables' own gate was shadowed", resp.StatusCode)
	}
	if resp, _ := coreRequest(t, ts, http.MethodPut, "/bdp/call", nil); resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("PUT /bdp/call = %d", resp.StatusCode)
	}

	for _, path := range []string{"/sw-1", "/bdp/beads", "/bdp/links"} {
		for _, method := range coreRefusedMethods {
			resp, body := coreRequest(t, ts, method, path, nil)
			if resp.StatusCode != http.StatusMethodNotAllowed || resp.Header.Get("Allow") != "GET, HEAD" || len(body) != 0 {
				t.Errorf("%s %s = %d, Allow %q, %d bytes", method, path, resp.StatusCode, resp.Header.Get("Allow"), len(body))
			}
		}
	}

	for _, path := range []string{"/a%2Fb", "/%00", "/a%5Cb"} {
		resp, body := coreRequest(t, ts, http.MethodGet, path, credential)
		wantCoreProblem(t, resp, body, coreCodeNotFound)
	}
}

// TestCoreFaultsOffTheWire: a store that is out of reach is Core's temporary
// failure, and an unexpected fault is a bodyless 500 with its details in the log.
func TestCoreFaultsOffTheWire(t *testing.T) {
	store := &coreStore{issues: []*types.Issue{coreIssue("sw-1", types.TypeTask)}, getErr: ErrBusy}
	ts := newCoreServer(t, store, Config{})
	resp, body := coreRequest(t, ts, http.MethodGet, "/sw-1", nil)
	wantCoreProblem(t, resp, body, coreCodeUnavailable)
	if resp.Header.Get("Retry-After") == "" {
		t.Error("a temporary failure has no Retry-After")
	}

	store.mu.Lock()
	store.getErr = errors.New("secret table name")
	store.mu.Unlock()
	resp, body = coreRequest(t, ts, http.MethodGet, "/sw-1", nil)
	if resp.StatusCode != http.StatusInternalServerError || len(body) != 0 {
		t.Errorf("fault = %d: %s", resp.StatusCode, body)
	}
}

// TestCoreSparseSelectorAdvances: a Selector matching little answers a short page
// whose next has moved, rather than holding the request open over the whole store
// or claiming exhaustion it has not observed.
func TestCoreSparseSelectorAdvances(t *testing.T) {
	store := &coreStore{}
	for i := range coreSelectorScanBudget + 10 {
		store.issues = append(store.issues, coreIssue(fmt.Sprintf("sw-%05d", i), types.TypeTask))
	}
	last := store.issues[len(store.issues)-1]
	last.Priority = 0
	ts := newCoreServer(t, store, Config{})

	selector := url.QueryEscape(`$[?@.properties.priority == 0]`)
	_, first := coreGetJSON(t, ts, "/bdp/beads?selector="+selector)
	if ids := coreItemIDs(t, first); len(ids) != 0 {
		t.Fatalf("first page = %v, want the budget spent with nothing found", ids)
	}
	next, ok := first["next"].(string)
	if !ok {
		t.Fatalf("next = %v, want a continuation: the scan has not reached the end", first["next"])
	}
	_, second := coreGetJSON(t, ts, strings.TrimPrefix(next, ts.base))
	if ids := coreItemIDs(t, second); !slices.Equal(ids, []string{ts.base + "/" + last.ID}) {
		t.Errorf("second page = %v", ids)
	}
	if second["next"] != nil {
		t.Errorf("next = %v, want null at the end", second["next"])
	}
}

func TestCoreScopeFromRequest(t *testing.T) {
	for _, tc := range []struct {
		host string
		tls  bool
		want string
	}{
		{"127.0.0.1:8080", false, "http://127.0.0.1:8080/"},
		{"LOCALHOST:8080", false, "http://localhost:8080/"},
		{"bd.svc:80", false, "http://bd.svc/"},
		{"bd.svc:443", false, "http://bd.svc:443/"},
		{"bd.svc:443", true, "https://bd.svc/"},
		{"[::1]:80", false, "http://[::1]/"},
		{"[::1]:8080", false, "http://[::1]:8080/"},
		{"bd.svc", false, "http://bd.svc/"},
	} {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Host = tc.host
		if tc.tls {
			r.TLS = &tls.ConnectionState{}
		}
		if got := coreScope(r); got != tc.want {
			t.Errorf("Host %q, TLS %v = %q, want %q", tc.host, tc.tls, got, tc.want)
		}
	}
}

func TestCoreLinkIDRoundTrip(t *testing.T) {
	segment := coreLinkID("sw-a.1", "external:proj:cap", coreClock)
	if !coreSafeSegment(segment) || strings.ContainsAny(segment, "/:=+") {
		t.Errorf("%q is not one plain segment", segment)
	}
	source, target, generation, ok := coreParseLinkID(segment)
	if !ok || source != "sw-a.1" || target != "external:proj:cap" || generation != coreClock.Unix() {
		t.Errorf("round trip = %q, %q, %d, %v", source, target, generation, ok)
	}
	for _, bad := range []string{"sw-a", "dep~", "dep~@@", coreLinkIDPrefix + "c3c"} {
		if _, _, _, ok := coreParseLinkID(bad); ok {
			t.Errorf("%q parsed as a Link id", bad)
		}
	}
}
