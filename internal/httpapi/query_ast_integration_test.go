//go:build cgo

package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/httpapi/bdpwire"
	"github.com/steveyegge/beads/internal/httpapi/graphread"
	"github.com/steveyegge/beads/internal/resourcequery"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/issueops"
)

// This opt-in proof uses real ordinary graphstore operations and the actual
// GraphRead dispatcher/Reader executor, including the configured Server handler
// authentication/Host gates. httptest recorders avoid a listener: socket behavior
// and an installed CLI remain separate proofs.
// No raw SQL, substitute provider, private candidate or payload seeding is used.
func TestQueryASTEngineRoundTrip(t *testing.T) {
	backend := os.Getenv("BEADS_QUERY_ENGINE_TEST_BACKEND")
	if backend == "" {
		t.Skip("set owned Query engine backend/DB/receipt custody for this proof")
	}
	if backend != "embedded" && backend != "server" {
		t.Fatal("Query engine backend must be embedded or server")
	}
	database := os.Getenv("BEADS_QUERY_ENGINE_TEST_DATABASE")
	if !regexp.MustCompile(`^bds_exp_query_[a-z0-9_]{1,49}$`).MatchString(database) {
		t.Fatal("explicit fresh bds_exp_query_* database name required")
	}
	receiptDir := os.Getenv("BEADS_QUERY_ENGINE_HTTP_RECEIPTS")
	if receiptDir == "" && os.Getenv("TEST_UNDECLARED_OUTPUTS_DIR") != "" {
		receiptDir = filepath.Join(os.Getenv("TEST_UNDECLARED_OUTPUTS_DIR"), "query-http-receipts")
	}
	if !filepath.IsAbs(receiptDir) || filepath.Clean(receiptDir) != receiptDir {
		t.Fatal("explicit absolute new receipt directory required")
	}
	if err := os.Mkdir(receiptDir, 0700); err != nil {
		t.Fatalf("create exclusive owned receipt directory: %v", err)
	}
	workspace, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()
	const scope = "https://query-engine.example/read/"
	options := graphstore.Options{Backend: backend, Database: database, Branch: "main", DataDir: filepath.Join(workspace, "dolt"), IssuePrefix: "exp",
		Binding: graphstore.Binding{WorkspaceID: workspace, ScopeURL: scope, AuthorityID: "9123456789abcdef0123456789abcdef", SchemaVersion: graphstore.SchemaVersion}}
	if backend == "server" {
		port, err := strconv.Atoi(os.Getenv("BEADS_QUERY_ENGINE_TEST_SERVER_PORT"))
		if err != nil || port < 1 || port > 65535 {
			t.Fatal("server proof requires explicit owned loopback port")
		}
		if os.Getenv("BEADS_QUERY_ENGINE_SERVER_CUSTODY") != "owned-disposable" {
			t.Fatal("server/database cleanup custody must be acknowledged by finite runner")
		}
		options.ServerHost, options.ServerPort, options.ServerUser = "127.0.0.1", port, "root"
	}
	queryEngineWriteReceipt(t, receiptDir, "00-target", map[string]any{"backend": backend, "database": database, "host": options.ServerHost, "port": options.ServerPort, "workspace": workspace, "dataDir": options.DataDir, "scope": scope, "serverCleanup": "runner owns exact database/server disposal after store Close; test does not issue admin SQL"})
	if err := graphstore.Init(ctx, options); err != nil {
		t.Fatal(err)
	}
	store, err := graphstore.OpenExisting(ctx, options)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		closeErr := store.Close()
		queryEngineWriteReceipt(t, receiptDir, "99-store-close", map[string]any{"backend": backend, "database": database, "closed": closeErr == nil, "serverDisposalOwnedByRunner": backend == "server"})
		if closeErr != nil {
			t.Error(closeErr)
		}
	})

	issueA, err := store.CreateIssue(ctx, "beads/issue-a", issueops.CreateRequest{Actor: "human:Donna", Issue: &issueops.Issue{Title: "Issue A", Status: "open", IssueType: "task", Priority: 2}})
	if err != nil {
		t.Fatal(err)
	}
	issueB, err := store.CreateIssue(ctx, "beads/issue-b", issueops.CreateRequest{Actor: "human:Donna", Issue: &issueops.Issue{Title: "Issue B", Status: "open", IssueType: "task", Priority: 2}})
	if err != nil {
		t.Fatal(err)
	}
	memories := map[string]graphstore.Record{}
	for _, input := range []struct{ path, publicID, typ, title, status string }{
		{"beads/policy", "query-policy", "policy", "Follow the rails", "active"},
		{"beads/requirement", "query-requirement", "requirement", "Final requirement", "draft"},
		{"beads/noise", "query-noise", "memory", "Unrelated note", "inactive"},
	} {
		properties, err := json.Marshal(map[string]any{"title": input.title, "body": "Engine fixture content", "status": input.status})
		if err != nil {
			t.Fatal(err)
		}
		record, err := store.Create(ctx, graphstore.CreateRequest{Path: input.path, PublicID: input.publicID, TypeURL: scope + input.typ, PropertiesJSON: properties, Actor: "agent:Melinda", Metadata: json.RawMessage(`{"fixture":"query-http"}`)})
		if err != nil {
			t.Fatal(err)
		}
		memories[input.path] = record
	}
	links := map[string]graphstore.LinkRecord{}
	for _, input := range []struct{ path, source, target, status string }{
		{"links/one", "beads/issue-a", "beads/policy", "active"},
		{"links/two", "beads/policy", "beads/issue-b", "active"},
		{"links/three", "beads/issue-b", "beads/requirement", "draft"},
		{"links/noise", "beads/issue-a", "beads/noise", "inactive"},
	} {
		result, err := store.AddInformationalLink(ctx, graphstore.LinkCreateRequest{Path: input.path, TypeURL: graphstore.RelatedTypeURL(scope), SourcePath: input.source, TargetPath: input.target, Actor: "agent:Melinda", UnconditionalSource: true, Properties: map[string]any{"status": input.status, "body": "Reference"}})
		if err != nil {
			t.Fatal(err)
		}
		links[input.path] = result.Link
	}
	dependency, err := store.AddDependency(ctx, graphstore.DependencyRequest{Path: "links/native", SourcePath: "beads/issue-a", TargetPath: "beads/issue-b", Actor: "human:Donna"})
	if err != nil {
		t.Fatal(err)
	}
	handler, err := NewGraphReadWithQuery(graphread.New(store), scope)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { handler.Close() })
	sequence := 0
	send := func(t *testing.T, label, method, target string, body any, requestCtx context.Context) *httptest.ResponseRecorder {
		t.Helper()
		sequence++
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		if body == nil {
			raw = nil
		}
		request := httptest.NewRequest(method, target, strings.NewReader(string(raw))).WithContext(requestCtx)
		request.Header.Set("Accept", queryASTMedia)
		if method == "POST" {
			request.Header.Set("Content-Type", queryASTMedia)
		} else {
			request.Header.Set("Accept", "application/json")
		}
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		var responseJSON any
		if response.Body.Len() > 0 {
			if err := json.Unmarshal(response.Body.Bytes(), &responseJSON); err != nil {
				t.Fatalf("non-JSON response: %s", response.Body.String())
			}
		}
		queryEngineWriteReceipt(t, receiptDir, fmt.Sprintf("%02d-%s", sequence, label), map[string]any{"method": method, "requestTarget": target, "requestBody": json.RawMessage(raw), "status": response.Code, "headers": response.Header(), "responseBody": responseJSON})
		return response
	}
	post := func(t *testing.T, label string, ast map[string]any) *httptest.ResponseRecorder {
		t.Helper()
		operation := "query"
		if ast["profile"] == "list" {
			operation = "list"
		}
		return send(t, label, "POST", "/read/?readVersion=1&operation="+operation, map[string]any{"query": ast}, ctx)
	}
	request := func(profile, form string, selection any) map[string]any {
		return map[string]any{"ir": "bdp-query-design-1", "profile": profile, "selection": selection, "result": map[string]any{"form": form}}
	}
	binding := func(name string) map[string]any { return map[string]any{"op": "binding", "name": name} }
	family := func(name, typ string) map[string]any {
		return map[string]any{"op": "typeIs", "of": binding(name), "type": typ, "match": "family"}
	}
	status := func(name, value string) map[string]any {
		return map[string]any{"op": "compare", "operator": "eq", "left": map[string]any{"op": "field", "of": binding(name), "path": []string{"properties", "status"}}, "right": map[string]any{"op": "literal", "value": value}}
	}
	walk := func(through any) map[string]any {
		return map[string]any{"op": "walk", "roots": map[string]any{"op": "refs", "collection": "beads", "references": []string{issueA.ID}}, "direction": "outgoing", "depth": 3, "bind": "l", "through": through}
	}
	referenceWalk := walk(family("l", graphstore.RelatedTypeURL(scope)))

	t.Run("mixed-inventory-family-and-count", func(t *testing.T) {
		for _, item := range []struct {
			collection, typ string
			want            int
		}{
			{"beads", graphstore.MemoryTypeURL(scope), 3}, {"beads", graphstore.IssueTypeURL(scope), 2}, {"links", graphstore.DependencyTypeURL(scope), 1},
		} {
			selection := map[string]any{"op": "filter", "input": map[string]any{"op": "scan", "collection": item.collection}, "bind": "r", "predicate": family("r", item.typ)}
			response := post(t, "family-count-"+item.collection+"-"+strconv.Itoa(item.want), request("list", "count", selection))
			result := queryEngineResult(t, response)
			if result["count"] != float64(item.want) || result["resourceKind"] != strings.TrimSuffix(item.collection, "s") {
				t.Fatalf("wrong exact family count: %s", response.Body.String())
			}
		}
		response := post(t, "mixed-bead-inventory", request("list", "resources", map[string]any{"op": "scan", "collection": "beads"}))
		queryEngineAssertIDs(t, response, []string{issueA.ID, issueB.ID, memories["beads/policy"].ID, memories["beads/requirement"].ID, memories["beads/noise"].ID})
		if strings.Contains(response.Body.String(), `"attribution":`) || strings.Contains(response.Body.String(), `"changeContext":`) || !strings.Contains(response.Body.String(), `"revisionProvenance":`) {
			t.Fatalf("wrong selected carrier: %s", response.Body.String())
		}
	})
	t.Run("cross-family-through-versus-outer-selection", func(t *testing.T) {
		filtered := map[string]any{"op": "filter", "input": referenceWalk, "bind": "selected", "predicate": status("selected", "draft")}
		queryEngineAssertIDs(t, post(t, "outer-draft", request("query", "resources", filtered)), []string{links["links/three"].ID})
		restricted := walk(map[string]any{"op": "and", "operands": []any{family("l", graphstore.RelatedTypeURL(scope)), status("l", "draft")}})
		queryEngineAssertIDs(t, post(t, "through-draft", request("query", "resources", restricted)), nil)
		active := walk(map[string]any{"op": "and", "operands": []any{family("l", graphstore.RelatedTypeURL(scope)), status("l", "active")}})
		queryEngineAssertIDs(t, post(t, "through-active", request("query", "resources", active)), []string{links["links/one"].ID, links["links/two"].ID})
	})
	t.Run("typed-graph-and-native-dependency", func(t *testing.T) {
		response := post(t, "reference-graph", request("graph", "graph", referenceWalk))
		result := queryEngineResult(t, response)
		beads, ok := result["beads"].([]any)
		if !ok || len(beads) != 5 {
			t.Fatalf("graph beads: %s", response.Body.String())
		}
		graphLinks, ok := result["links"].([]any)
		if !ok || len(graphLinks) != 4 {
			t.Fatalf("graph Links: %s", response.Body.String())
		}
		mappings, ok := result["endpoints"].([]any)
		if !ok || len(mappings) != 4 {
			t.Fatalf("graph mappings: %s", response.Body.String())
		}
		roots, ok := result["roots"].([]any)
		if !ok || len(roots) != 1 || roots[0].(map[string]any)["id"] != issueA.ID {
			t.Fatalf("graph roots: %s", response.Body.String())
		}
		beadEditions := map[string]string{}
		for _, value := range beads {
			bead := value.(map[string]any)
			beadEditions[bead["id"].(string)] = bead["revision"].(string)
		}
		represented := map[string]map[string]any{}
		for _, value := range graphLinks {
			link := value.(map[string]any)
			id := link["id"].(string)
			if id == dependency.Link.ID {
				t.Fatal("native Dependency leaked into Reference family traversal")
			}
			represented[id] = link
		}
		mapped := map[string]bool{}
		for _, mapping := range mappings {
			item := mapping.(map[string]any)
			edition := item["link"].(map[string]any)
			link := represented[edition["id"].(string)]
			if link == nil || link["revision"] != edition["revision"] {
				t.Fatal("mapping selected another Link edition")
			}
			id := edition["id"].(string)
			if mapped[id] {
				t.Fatal("duplicate Link edition mapping")
			}
			mapped[id] = true
			for _, side := range []string{"source", "target"} {
				endpoint := item[side].(map[string]any)
				if link[side] != endpoint["id"] || beadEditions[endpoint["id"].(string)] != endpoint["revision"] {
					t.Fatal("graph mapping did not preserve floating reference/selected edition")
				}
			}
		}
		if len(mapped) != len(represented) {
			t.Fatal("graph mappings do not cover every represented Link")
		}
	})
	t.Run("legacy-get-and-discovery-unchanged", func(t *testing.T) {
		parsed, err := url.Parse(memories["beads/policy"].ID)
		if err != nil {
			t.Fatal(err)
		}
		response := send(t, "old-current-get", "GET", parsed.RequestURI(), nil, ctx)
		var bead bdpwire.BeadRecord
		if response.Code != 200 || bdpwire.Unmarshal(response.Body.Bytes(), &bead) != nil || bead.ID != memories["beads/policy"].ID || bead.Attribution == nil || response.Header().Get("Content-Type") != "application/json" {
			t.Fatalf("legacy GET changed: %s", response.Body.String())
		}
		response = send(t, "old-discovery", "GET", "/read/bdp.json", nil, ctx)
		var discovery bdpwire.ReadDiscovery
		if response.Code != 200 || bdpwire.Unmarshal(response.Body.Bytes(), &discovery) != nil || discovery.BDPVersion != bdpwire.BDPVersion || discovery.Scope != scope {
			t.Fatalf("legacy discovery changed: %s", response.Body.String())
		}
	})
	t.Run("refusals-cancellation-and-reopen", func(t *testing.T) {
		pin := map[string]any{"op": "refs", "collection": "beads", "references": []any{map[string]any{"uri": issueA.ID, "revision": issueA.Revision}}}
		queryEngineAssertProblem(t, post(t, "explicit-pin-refusal", request("query", "resources", pin)), 422, "capability-not-supported")
		queryEngineAssertProblem(t, post(t, "unknown-operator", request("query", "resources", map[string]any{"op": "not-an-operator"})), 400, "invalid-parameter")
		missing := map[string]any{"op": "refs", "collection": "beads", "references": []string{scope + "missing"}}
		queryEngineAssertProblem(t, post(t, "missing-refusal", request("query", "resources", missing)), 404, "resource-not-found")
		cancelled, stop := context.WithCancel(ctx)
		stop()
		response := send(t, "cancelled", "POST", "/read/?readVersion=1&operation=list", map[string]any{"query": request("list", "count", map[string]any{"op": "scan", "collection": "beads"})}, cancelled)
		queryEngineAssertProblem(t, response, 503, "temporarily-unavailable")
		before := post(t, "before-reopen", request("query", "resources", referenceWalk))
		queryEngineResult(t, before)
		handler.Close()
		queryEngineAssertProblem(t, post(t, "closed-handler", request("query", "resources", referenceWalk)), 503, "temporarily-unavailable")
		if err := store.Close(); err != nil {
			t.Fatal(err)
		}
		store, err = graphstore.OpenExisting(ctx, options)
		if err != nil {
			t.Fatal(err)
		}
		handler, err = NewGraphReadWithQuery(graphread.New(store), scope)
		if err != nil {
			t.Fatal(err)
		}
		after := post(t, "after-reopen", request("query", "resources", referenceWalk))
		queryEngineResult(t, after)
		var a, b any
		if json.Unmarshal(before.Body.Bytes(), &a) != nil || json.Unmarshal(after.Body.Bytes(), &b) != nil || !reflect.DeepEqual(a, b) {
			t.Fatalf("reopen changed complete Query answer:\n%s\n%s", before.Body.String(), after.Body.String())
		}
	})

	t.Run("configured-security-handler-chain", func(t *testing.T) {
		// Reuse the same normally opened real store and current GraphRead. Assemble
		// the actual handler chain without Listen/Serve or an HTTP client/listener.
		tokenPath := filepath.Join(workspace, "query-security-token")
		const token = "query-engine-owned-test-token"
		if err := os.WriteFile(tokenPath, []byte(token+"\n"), 0600); err != nil {
			t.Fatal(err)
		}
		auth, err := NewTokenFileAuth(tokenPath)
		if err != nil {
			t.Fatal(err)
		}
		cfg := Config{GraphRead: handler, Addr: "127.0.0.1:0", Auth: auth, AllowedHosts: []string{"query-transport.example"}}
		if err := checkDatabaseSource(cfg); err != nil {
			t.Fatal(err)
		}
		bindIP, err := ValidateBindAddr(cfg.Addr, false)
		if err != nil {
			t.Fatal(err)
		}
		if !bindIP.Equal(net.ParseIP("127.0.0.1")) {
			t.Fatal("security fixture is not loopback")
		}
		if err := ValidateAuthPosture(false, true, false); err != nil {
			t.Fatal(err)
		}
		service := &Server{cfg: cfg, auth: auth, hosts: newHostPolicy(bindIP, cfg.AllowedHosts), sem: make(chan struct{}, maxInflight), semTimeout: time.Second, log: newTestLogger(&lockedBuffer{}), idPrefix: "query-security"}
		actualAuthority, actualExecute := handler.queryAST.validateAuthority, handler.queryAST.execute
		authorities, executions := 0, 0
		handler.queryAST.validateAuthority = func(ctx context.Context) error {
			authorities++
			return actualAuthority(ctx)
		}
		handler.queryAST.execute = func(ctx context.Context, raw []byte) (resourcequery.Response, error) {
			executions++
			deadline, ok := ctx.Deadline()
			if remaining := time.Until(deadline); !ok || remaining <= 0 || remaining > requestDeadline {
				t.Errorf("Query did not inherit the actual server request deadline: %v %t", deadline, ok)
			}
			if len(service.sem) != 1 {
				t.Error("Query executed outside the actual request semaphore")
			}
			return actualExecute(ctx, raw)
		}
		defer func() { handler.queryAST.validateAuthority, handler.queryAST.execute = actualAuthority, actualExecute }()
		chain := service.handler()
		payload, err := json.Marshal(map[string]any{"query": request("list", "count", map[string]any{"op": "scan", "collection": "beads"})})
		if err != nil {
			t.Fatal(err)
		}
		for _, input := range []struct {
			name, host string
			credential bool
			status     int
			code       string
			calls      int
		}{
			{"bad-host", "unallowed.example", true, 400, "malformed-request", 0},
			{"unauthenticated", "query-transport.example", false, 401, "unauthenticated", 0},
			{"authorized", "query-transport.example", true, 200, "", 1},
		} {
			r := httptest.NewRequest("POST", "/read/?readVersion=1&operation=list", strings.NewReader(string(payload))).WithContext(ctx)
			r.Host = input.host
			r.Header.Set("Content-Type", queryASTMedia)
			r.Header.Set("Accept", queryASTMedia)
			if input.credential {
				r.Header.Set("Authorization", "Bearer "+token)
			}
			oldAuthorities, oldExecutions := authorities, executions
			response := httptest.NewRecorder()
			chain.ServeHTTP(response, r)
			if authorities-oldAuthorities != input.calls || executions-oldExecutions != input.calls {
				t.Fatalf("%s reached Query before security: authority=%d execution=%d", input.name, authorities-oldAuthorities, executions-oldExecutions)
			}
			if input.status == 200 {
				if result := queryEngineResult(t, response); result["count"] != float64(5) {
					t.Fatalf("authorized Query did not use real store: %s", response.Body.String())
				}
			} else {
				queryEngineAssertProblem(t, response, input.status, input.code)
			}
			if len(service.sem) != 0 {
				t.Fatal("security request retained a semaphore slot")
			}
			var responseJSON any
			if err := json.Unmarshal(response.Body.Bytes(), &responseJSON); err != nil {
				t.Fatal(err)
			}
			sequence++
			// Credential contents are deliberately absent from receipts and logs.
			queryEngineWriteReceipt(t, receiptDir, fmt.Sprintf("%02d-security-%s", sequence, input.name), map[string]any{"method": "POST", "requestTarget": r.RequestURI, "host": input.host, "authorizationSupplied": input.credential, "requestBody": json.RawMessage(payload), "status": response.Code, "headers": response.Header(), "responseBody": responseJSON})
		}
	})
	t.Logf("exact engine/HTTP receipts: %s (not yet independently checked against selected v0.1 schema)", receiptDir)
}

func queryEngineWriteReceipt(t *testing.T, dir, name string, value any) {
	t.Helper()
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	file, err := os.OpenFile(filepath.Join(dir, name+".json"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		t.Fatal(err)
	}
	_, writeErr := file.Write(append(raw, '\n'))
	closeErr := file.Close()
	if writeErr != nil {
		t.Fatal(writeErr)
	}
	if closeErr != nil {
		t.Fatal(closeErr)
	}
}
func queryEngineResult(t *testing.T, response *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var envelope struct {
		Result     map[string]any `json:"result"`
		Evaluation struct {
			Scope    string `json:"scope"`
			Complete bool   `json:"complete"`
		} `json:"evaluation"`
	}
	if response.Code != 200 || response.Header().Get("Content-Type") != queryASTMedia || json.Unmarshal(response.Body.Bytes(), &envelope) != nil || !envelope.Evaluation.Complete || envelope.Evaluation.Scope != "response" || envelope.Result == nil {
		t.Fatalf("incomplete Query response %d %s", response.Code, response.Body.String())
	}
	return envelope.Result
}
func queryEngineAssertIDs(t *testing.T, response *httptest.ResponseRecorder, want []string) {
	t.Helper()
	result := queryEngineResult(t, response)
	items, ok := result["items"].([]any)
	if !ok {
		t.Fatalf("missing Resource items: %s", response.Body.String())
	}
	got := make([]string, 0, len(items))
	for _, item := range items {
		got = append(got, item.(map[string]any)["id"].(string))
	}
	expected := append([]string{}, want...)
	sort.Strings(got)
	sort.Strings(expected)
	if !reflect.DeepEqual(got, expected) || result["next"] != nil {
		t.Fatalf("wrong complete identities got=%v want=%v body=%s", got, expected, response.Body.String())
	}
}
func queryEngineAssertProblem(t *testing.T, response *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	var problem map[string]any
	if response.Code != status || json.Unmarshal(response.Body.Bytes(), &problem) != nil || problem["code"] != code || problem["result"] != nil {
		t.Fatalf("wrong refusal %d %s", response.Code, response.Body.String())
	}
	if code == "resource-not-found" && problem["at"] != nil {
		t.Fatal("missing Resource disclosed request branch")
	}
	if code == "capability-not-supported" && (problem["capability"] != queryASTCapability || problem["capabilityVersion"] != "1") {
		t.Fatal("unsupported capability omitted binding")
	}
}
