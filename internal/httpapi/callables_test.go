package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"slices"
	"sort"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/issueops"
)

// The pins for the Callables layer (callables.go). Pure, on fake roles, like the
// rest of this package: a fake cannot show that a routed call commits against
// real storage, only that it reaches the v0 handler that does.

// callSend posts one invocation with the media type the profile requires.
func (ts *testServer) callSend(t *testing.T, name, body string, header http.Header) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, ts.base+callableMount+"/"+name, strings.NewReader(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", bdpMediaType)
	for key, values := range header {
		req.Header[key] = values
	}
	resp, err := ts.client.Do(req)
	if err != nil {
		t.Fatalf("POST %s: %v", name, err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })
	return resp
}

func (ts *testServer) call(t *testing.T, name, arguments string) *http.Response {
	t.Helper()
	return ts.callSend(t, name, `{"arguments":`+arguments+`}`, nil)
}

// callAs invokes a Callable with revisionProvenance naming the actor.
func (ts *testServer) callAs(t *testing.T, name, actor, arguments string) *http.Response {
	t.Helper()
	return ts.callSend(t, name, `{"arguments":`+arguments+`,"revisionProvenance":{"actor":{"id":"`+actor+`","basis":"writer-supplied"}}}`, nil)
}

// wantProblem asserts a mount's Problem and returns its body.
func wantProblem(t *testing.T, resp *http.Response, status int, code string) map[string]any {
	t.Helper()
	raw := readAll(t, resp)
	if resp.StatusCode != status {
		t.Fatalf("status = %d, want %d: %s", resp.StatusCode, status, raw)
	}
	if got := resp.Header.Get("Content-Type"); got != problemMediaType {
		t.Errorf("Content-Type = %q, want %q", got, problemMediaType)
	}
	if got := resp.Header.Get("Cache-Control"); got != "private, no-store" {
		t.Errorf("Cache-Control = %q", got)
	}
	var body map[string]any
	if err := json.Unmarshal([]byte(raw), &body); err != nil {
		t.Fatalf("body is not JSON: %v: %s", err, raw)
	}
	if body["code"] != code {
		t.Fatalf("code = %v, want %s: %s", body["code"], code, raw)
	}
	shape := bdpProblemShapes[code]
	if body["type"] != bdpProblemBase+shape.family || body["retry"] != shape.retry || body["status"] != float64(status) {
		t.Errorf("type/retry/status do not match the code's fixed shape: %s", raw)
	}
	return body
}

// wantDiagnosticNaming asserts a validation-failed names one parameter.
func wantDiagnosticNaming(t *testing.T, body map[string]any, param string) {
	t.Helper()
	diagnostics, _ := body["diagnostics"].([]any)
	for _, raw := range diagnostics {
		entry, _ := raw.(map[string]any)
		if message, _ := entry["message"].(string); strings.Contains(message, "`"+param+"`") {
			return
		}
	}
	t.Errorf("no diagnostic names %q: %v", param, body)
}

// callResult asserts a 200 and returns the raw `result` and the envelope.
func callResult(t *testing.T, resp *http.Response) (json.RawMessage, map[string]json.RawMessage) {
	t.Helper()
	raw := readAll(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", resp.StatusCode, raw)
	}
	if got := resp.Header.Get("Content-Type"); got != bdpMediaType {
		t.Errorf("Content-Type = %q, want %q", got, bdpMediaType)
	}
	if got := resp.Header.Get("Cache-Control"); got != "private, no-store" {
		t.Errorf("Cache-Control = %q", got)
	}
	if resp.Header.Get("ETag") != "" || resp.Header.Get("Location") != "" {
		t.Errorf("a result carries no ETag or Location: %v", resp.Header)
	}
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &envelope); err != nil {
		t.Fatalf("envelope is not JSON: %v: %s", err, raw)
	}
	for name := range envelope {
		if name != "result" && name != "changed" {
			t.Errorf("callableResult is closed, but it has %q", name)
		}
	}
	return envelope["result"], envelope
}

func specParameterNames(t *testing.T, doc map[string]any, op specOp) (path, query, body []string) {
	t.Helper()
	if list, ok := op.op["parameters"].([]any); ok {
		for _, raw := range list {
			node, _ := raw.(map[string]any)
			param := resolveRef(t, doc, node)
			name, _ := param["name"].(string)
			switch param["in"] {
			case "path":
				path = append(path, name)
			case "query":
				query = append(query, name)
			}
		}
	}
	if raw, ok := op.op["requestBody"].(map[string]any); ok {
		requestBody := resolveRef(t, doc, raw)
		media := mapAt(t, mapAt(t, requestBody, "content"), claimContentType)
		schema := resolveRef(t, doc, mapAt(t, media, "schema"))
		for name := range mapAt(t, schema, "properties") {
			body = append(body, name)
		}
	}
	sort.Strings(path)
	sort.Strings(query)
	sort.Strings(body)
	return path, query, body
}

func sortedCopy(names []string) []string {
	out := slices.Clone(names)
	sort.Strings(out)
	return out
}

// TestCallableTableMatchesTheDocument holds the one hand-written table to
// openapi.v0.yaml in both directions: every operation but the two plain routes is
// declared, and each row lists exactly its operation's parameters, each at its
// location. It also proves flattening is unambiguous: no operation uses one name
// in two locations.
func TestCallableTableMatchesTheDocument(t *testing.T) {
	doc := loadSpec(t)
	ops := specOps(t, doc)

	names := map[string]bool{}
	declared := map[string]bool{}
	for i := range callableTable {
		row := &callableTable[i]
		if names[row.name] {
			t.Errorf("name %q is declared twice", row.name)
		}
		names[row.name] = true
		declared[row.op] = true

		op, ok := ops[row.op]
		if !ok {
			t.Errorf("%s routes to %q, which the document does not declare", row.name, row.op)
			continue
		}
		rt, ok := routeForOp(row.op)
		if !ok || !rt.implemented || rt.streaming {
			t.Errorf("%s routes to %q, which is no implemented document route", row.name, row.op)
		}
		path, query, body := specParameterNames(t, doc, op)
		for _, part := range []struct {
			location   string
			table, doc []string
		}{{"path", row.path, path}, {"query", row.query, query}, {"body", row.v0BodyMembers(), body}} {
			if got := sortedCopy(part.table); !reflect.DeepEqual(got, part.doc) && (len(got) > 0 || len(part.doc) > 0) {
				t.Errorf("%s %s parameters:\n table %v\n   doc %v", row.name, part.location, got, part.doc)
			}
		}
		all := row.declaredParameters()
		sort.Strings(all)
		if len(slices.Compact(slices.Clone(all))) != len(all) {
			t.Errorf("%s uses one argument name twice: %v", row.name, all)
		}
		for _, name := range row.optional {
			if !slices.Contains(all, name) {
				t.Errorf("%s: optional %q is no parameter", row.name, name)
			}
		}
		// No declaration has an actor or an author (§2), and every operation that
		// records one says so on its row: the document's body list above holds that.
		if slices.Contains(all, "actor") || slices.Contains(all, "author") {
			t.Errorf("%s declares an actor or an author: %v", row.name, all)
		}
		if row.actorMember != "" {
			if _, v0Required := v0Parameters(doc, op.op); v0Required[row.actorMember] == row.actorOptional {
				t.Errorf("%s: actorOptional = %v, v0 requires %s = %v", row.name, row.actorOptional, row.actorMember, v0Required[row.actorMember])
			}
		}
		for name, value := range row.defaults {
			if !slices.Contains(all, name) {
				t.Errorf("%s: default for %q, which is no parameter", row.name, name)
			}
			if !json.Valid(value) {
				t.Errorf("%s: default for %q is %s", row.name, name, value)
			}
		}
	}
	if len(callableTable) != 45 || len(declared) != 44 {
		t.Errorf("table has %d names over %d operations, want 45 over 44", len(callableTable), len(declared))
	}
	for op := range ops {
		if declared[op] == (op == OpHealth || op == OpWatchEvents) {
			t.Errorf("operation %q: declared = %v", op, declared[op])
		}
	}
	if callableIndex["issue/purge"].op != OpSweepIssues || callableIndex["issue/prune"].op != OpSweepIssues {
		t.Error("issue/purge and issue/prune must both route to sweepIssues")
	}
}

// TestCallableReasonsMatchTheOperationCodes: every code an operation can answer
// is either mapped to a Core or Write code by §5, or is a reason its row declares.
// precondition_failed is exempt where the only guard is expected_version, which
// §5 maps per response.
func TestCallableReasonsMatchTheOperationCodes(t *testing.T) {
	versionGuardOnly := map[string]bool{OpCloseIssue: true, OpReopenIssue: true, OpDeleteIssues: true}
	for i := range callableTable {
		row := &callableTable[i]
		codes := operationCodes[row.op]
		for _, code := range codes {
			if bdpCodeFor(code, false) != bdpCallableError {
				continue
			}
			if code == CodePreconditionFailed && versionGuardOnly[row.op] {
				continue
			}
			if !slices.Contains(row.reasons, string(code)) {
				t.Errorf("%s: v0 answers %s, but the row does not declare it", row.name, code)
			}
		}
		for _, reason := range row.reasons {
			if !slices.Contains(codes, Code(reason)) {
				t.Errorf("%s: declares %s, which v0's operation does not answer", row.name, reason)
			}
		}
	}
}

// TestEveryCallableReachesItsV0Handler replaces every v0 handler with a recorder,
// so what is asserted is the route table's own dispatch: each name lands on its
// operation's row, by that row's method and path, with each argument at the
// location the document gives it.
func TestEveryCallableReachesItsV0Handler(t *testing.T) {
	type seen struct {
		op, method, path, id, key, contentType string
		query                                  map[string][]string
		body                                   map[string]json.RawMessage
	}
	var last atomic.Pointer[seen]

	original := routeTable
	t.Cleanup(func() { routeTable = original })
	swapped := slices.Clone(original)
	for i := range swapped {
		op := swapped[i].op
		swapped[i].handler = func(_ *Server, w http.ResponseWriter, r *http.Request) {
			got := &seen{op: op, method: r.Method, path: r.URL.EscapedPath(), id: r.PathValue("id"),
				key: r.PathValue("key"), contentType: r.Header.Get("Content-Type"), query: r.URL.Query()}
			if data, _ := io.ReadAll(r.Body); len(data) > 0 {
				if err := json.Unmarshal(data, &got.body); err != nil {
					panic(err)
				}
			}
			last.Store(got)
			writeJSON(w, map[string]string{"op": op})
		}
	}
	routeTable = swapped
	ts := newBDPServer(t, rolesConfig(Config{}))

	for i := range callableTable {
		row := &callableTable[i]
		t.Run(row.name, func(t *testing.T) {
			last.Store(nil)
			args := map[string]string{}
			for _, name := range row.declaredParameters() {
				if !slices.Contains(row.adapterParams, name) {
					args[name] = "v-" + name
				}
			}
			encoded, _ := json.Marshal(args)
			var resp *http.Response
			if row.actorMember != "" {
				resp = ts.callAs(t, row.name, "v-"+row.actorMember, string(encoded))
			} else {
				resp = ts.call(t, row.name, string(encoded))
			}
			result, envelope := callResult(t, resp)
			last := last.Load()
			if last == nil {
				t.Fatal("no v0 handler ran")
			}
			if last.op != row.op {
				t.Errorf("reached %s, want %s", last.op, row.op)
			}
			if want := fmt.Sprintf(`{"op":%q}`, row.op); strings.TrimSpace(string(result)) != want {
				t.Errorf("result = %s, want v0's bytes %s", result, want)
			}
			if _, has := envelope["changed"]; has != row.batch {
				t.Errorf("changed present = %v on a row with batch = %v", has, row.batch)
			}

			rt, _ := routeForOp(row.op)
			if last.method != rt.method {
				t.Errorf("method = %s, want %s", last.method, rt.method)
			}
			wantPath := rt.specPathOf()
			for _, name := range row.path {
				wantPath = strings.Replace(wantPath, "{"+name+"}", "v-"+name, 1)
			}
			if last.path != wantPath {
				t.Errorf("path = %s, want %s", last.path, wantPath)
			}
			if slices.Contains(row.path, "id") && last.id != "v-id" {
				t.Errorf("path value id = %q", last.id)
			}
			if slices.Contains(row.path, "key") && last.key != "v-key" {
				t.Errorf("path value key = %q", last.key)
			}
			if len(last.query) != len(row.query) {
				t.Errorf("query = %v, want %v", last.query, row.query)
			}
			for _, name := range row.query {
				if !slices.Equal(last.query[name], []string{"v-" + name}) {
					t.Errorf("query %s = %v", name, last.query[name])
				}
			}
			members := row.v0BodyMembers()
			if len(last.body) != len(members) {
				t.Errorf("body = %v, want %v", last.body, members)
			}
			for _, name := range members {
				if string(last.body[name]) != `"v-`+name+`"` {
					t.Errorf("body %s = %s", name, last.body[name])
				}
			}
			if len(members) > 0 && last.contentType != claimContentType {
				t.Errorf("Content-Type = %q", last.contentType)
			}
		})
	}
}

// directAndRouted sends one operation both ways against one server and returns
// the direct body and the routed `result`.
func directAndRouted(t *testing.T, direct, routed *http.Response) (string, json.RawMessage, map[string]json.RawMessage) {
	t.Helper()
	body := readAll(t, direct)
	if direct.StatusCode != http.StatusOK {
		t.Fatalf("direct status = %d: %s", direct.StatusCode, body)
	}
	result, envelope := callResult(t, routed)
	if strings.TrimSpace(body) != string(result) {
		t.Errorf("routed result differs from the direct body:\n direct %s\n routed %s", body, result)
	}
	return body, result, envelope
}

func TestRoutedResponseEqualsDirect(t *testing.T) {
	t.Run("issue/show", func(t *testing.T) {
		reader := &roleReader{details: &issueops.IssueDetails{
			Issue: *seededIssue("bd-1", "alice", types.StatusOpen),
		}}
		ts := newBDPServer(t, rolesConfig(Config{Reader: reader}))
		direct := ts.get(t, "/v0/beads/issues/bd-1?include_comments=true")
		routed := ts.call(t, "issue/show", `{"id":"bd-1","include_comments":true}`)
		_, _, envelope := directAndRouted(t, direct, routed)
		if _, has := envelope["changed"]; has {
			t.Error("a CallableQuery result has no `changed`")
		}
		got := reader.getRequests()
		if len(got) != 2 || !reflect.DeepEqual(got[0], got[1]) {
			t.Errorf("the role saw different requests: %+v", got)
		}
	})

	t.Run("issue/update", func(t *testing.T) {
		lifecycle := &roleLifecycle{updateResult: issueops.UpdateResult{Issue: updatedIssue("bd-1"), Changed: true}}
		ts := newBDPServer(t, rolesConfig(Config{Lifecycle: lifecycle}))
		direct := ts.updateIssue(t, updatePath, `{"actor":"alice","patch":{"title":"new"}}`)
		routed := ts.callAs(t, "issue/update", "alice", `{"id":"bd-1","patch":{"title":"new"}}`)
		_, _, envelope := directAndRouted(t, direct, routed)
		if string(envelope["changed"]) != "true" {
			t.Errorf("changed = %s, want true", envelope["changed"])
		}
		got := lifecycle.updateRequests()
		if len(got) != 2 || !reflect.DeepEqual(got[0], got[1]) {
			t.Errorf("the role saw different requests: %+v", got)
		}

		lifecycle.updateResult.Changed = false
		_, envelope = callResult(t, ts.callAs(t, "issue/update", "alice", `{"id":"bd-1","patch":{"title":"new"}}`))
		if string(envelope["changed"]) != "false" {
			t.Errorf("a no-op update: changed = %s, want false", envelope["changed"])
		}
	})

	for _, tc := range []struct {
		name         string
		direct, sent string
		dryRun       bool
		wantChanged  string
	}{
		{"issue/delete real run", `"dry_run":false,`, `"dry_run":false,`, false, "true"},
		{"issue/delete dry run", `"dry_run":true,`, `"dry_run":true,`, true, "false"},
		{"issue/delete dry_run omitted previews", `"dry_run":true,`, "", true, "false"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			deleter := &roleDeleter{result: issueops.DeleteResult{Deleted: 2, DryRun: tc.dryRun}}
			ts := newBDPServer(t, rolesConfig(Config{Deleter: deleter}))
			direct := ts.delete(t, `{`+tc.direct+`"ids":["bd-1","bd-2"],"force":true}`)
			routed := ts.call(t, "issue/delete", `{`+tc.sent+`"ids":["bd-1","bd-2"],"force":true}`)
			_, _, envelope := directAndRouted(t, direct, routed)
			if string(envelope["changed"]) != tc.wantChanged {
				t.Errorf("changed = %s, want %s", envelope["changed"], tc.wantChanged)
			}
			got := deleter.requests()
			if len(got) != 2 || !reflect.DeepEqual(got[0], got[1]) {
				t.Fatalf("the role saw different requests: %+v", got)
			}
			if got[1].DryRun != tc.dryRun {
				t.Errorf("DryRun reaching the role = %v, want %v", got[1].DryRun, tc.dryRun)
			}
		})
	}
}

func TestCallableUnknownName(t *testing.T) {
	ts := newBDPServer(t, rolesConfig(Config{}))
	for _, name := range []string{"issue/nope", "nope", "issue", "issue/show/extra", ""} {
		wantProblem(t, ts.call(t, name, `{}`), http.StatusNotFound, bdpCallableNotFound)
	}
	// An encoded separator is not a second spelling of a name.
	wantProblem(t, ts.call(t, "issue%2Fshow", `{"id":"bd-1"}`), http.StatusNotFound, bdpCallableNotFound)
}

// TestCallableWrongMethod: a mount admits POST whatever v0's own method is, for a
// query as for a batch. GET of a mount is never an invocation.
func TestCallableWrongMethod(t *testing.T) {
	reader := &roleReader{}
	lifecycle := &roleLifecycle{}
	ts := newBDPServer(t, rolesConfig(Config{Reader: reader, Lifecycle: lifecycle}))

	for _, name := range []string{"issue/show", "issue/update", "issue/nope"} {
		for _, method := range []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodDelete, http.MethodPatch} {
			resp := ts.do(t, method, callableMount+"/"+name, nil)
			if resp.StatusCode != http.StatusMethodNotAllowed {
				t.Errorf("%s %s = %d, want 405", method, name, resp.StatusCode)
			}
			if got := resp.Header.Get("Allow"); got != "POST" {
				t.Errorf("%s %s: Allow = %q", method, name, got)
			}
			if body := readAll(t, resp); body != "" {
				t.Errorf("%s %s: the 405 carries a body: %s", method, name, body)
			}
		}
	}
	if len(reader.getRequests()) != 0 || len(lifecycle.updateRequests()) != 0 {
		t.Error("a wrong method reached a role")
	}

	resp := ts.do(t, http.MethodPost, callableMount, nil)
	if resp.StatusCode != http.StatusMethodNotAllowed || resp.Header.Get("Allow") != "GET, HEAD" {
		t.Errorf("POST /bdp/call = %d, Allow %q", resp.StatusCode, resp.Header.Get("Allow"))
	}
}

func TestCallableEnvelopeRefusals(t *testing.T) {
	reader := &roleReader{}
	lifecycle := &roleLifecycle{}
	ts := newBDPServer(t, rolesConfig(Config{Reader: reader, Lifecycle: lifecycle}))

	for _, media := range []string{"", "application/json", "text/plain"} {
		resp := ts.postBody(t, callableMount+"/issue/show", media, `{"arguments":{"id":"bd-1"}}`)
		wantProblem(t, resp, http.StatusUnsupportedMediaType, bdpUnsupportedMedia)
		resp = ts.postBody(t, callableMount+"/issue/update", media, `{"arguments":{"id":"bd-1"}}`)
		wantProblem(t, resp, http.StatusUnsupportedMediaType, bdpUnsupportedMedia)
	}

	for name, body := range map[string]string{
		"no arguments":        `{}`,
		"arguments is null":   `{"arguments":null}`,
		"arguments an array":  `{"arguments":[]}`,
		"unknown member":      `{"arguments":{"id":"bd-1"},"operation":"x"}`,
		"name member":         `{"arguments":{"id":"bd-1"},"name":"x"}`,
		"duplicate member":    `{"arguments":{"id":"bd-1"},"arguments":{"id":"bd-2"}}`,
		"nested duplicate":    `{"arguments":{"id":"bd-1","id":"bd-2"}}`,
		"deep duplicate":      `{"arguments":{"id":"bd-1","patch":{"a":[{"b":1,"b":2}]}}}`,
		"escaped duplicate":   `{"arguments":{"id":"bd-1","id":"bd-2"}}`,
		"lone high surrogate": `{"arguments":{"id":"\ud800"}}`,
		"lone low surrogate":  `{"arguments":{"id":"\udc00x"}}`,
		"not JSON":            `{"arguments":`,
		"trailing data":       `{"arguments":{"id":"bd-1"}} {}`,
		"an array":            `[]`,
		"provenance on query": `{"arguments":{"id":"bd-1"},"revisionProvenance":{}}`,
	} {
		t.Run(name, func(t *testing.T) {
			wantProblem(t, ts.callSend(t, "issue/show", body, nil), http.StatusBadRequest, bdpMalformedRequest)
		})
	}
	if len(reader.getRequests()) != 0 {
		t.Error("a malformed invocation reached the role")
	}

	wantProblem(t, ts.callSend(t, "issue/show", `{"arguments":{"id":"bd-1"}}`,
		http.Header{"Idempotency-Key": {"k"}}), http.StatusBadRequest, bdpMalformedRequest)

	// Not retained, and admitted: the Scope does not advertise idempotency.
	lifecycle.updateResult = issueops.UpdateResult{Issue: updatedIssue("bd-1"), Changed: true}
	callResult(t, ts.callSend(t, "issue/update", `{"arguments":{"id":"bd-1","patch":{"title":"t"}},"revisionProvenance":{"actor":{"id":"a","basis":"unknown"}}}`,
		http.Header{"Idempotency-Key": {"k"}}))

	resp := ts.callSend(t, "issue/show", `{"arguments":{"id":"bd-1"}}`, http.Header{"Accept": {"application/json"}})
	if resp.StatusCode != http.StatusNotAcceptable || readAll(t, resp) != "" {
		t.Errorf("Accept application/json: status %d, want a bodyless 406", resp.StatusCode)
	}
	for _, accept := range []string{bdpMediaType, "application/*", "*/*", "application/json, " + bdpMediaType + ";q=0.5"} {
		if !acceptsMedia([]string{accept}) {
			t.Errorf("Accept %q was refused", accept)
		}
	}
	for _, accept := range []string{"text/plain", bdpMediaType + ";q=0", "*/*, " + bdpMediaType + ";q=0"} {
		if acceptsMedia([]string{accept}) {
			t.Errorf("Accept %q was admitted", accept)
		}
	}
}

// TestCallableCheckingArguments is Callables §4: unknown and missing parameters
// are validation-failed naming the parameter, and nothing runs.
func TestCallableCheckingArguments(t *testing.T) {
	reader := &roleReader{}
	claimer := &roleClaimer{}
	ts := newBDPServer(t, rolesConfig(Config{Reader: reader, Claimer: claimer}))

	body := wantProblem(t, ts.call(t, "issue/show", `{"id":"bd-1","zebra":1,"apple":2}`),
		http.StatusUnprocessableEntity, bdpValidationFailed)
	wantDiagnosticNaming(t, body, "zebra")
	wantDiagnosticNaming(t, body, "apple")

	for _, tc := range []struct{ name, arguments, param string }{
		{"issue/show", `{}`, "id"},
		{"issue/show", `{"id":""}`, "id"},
		{"issue/show", `{"id":null}`, "id"},
		{"issue/show", `{"id":7}`, "id"},
		{"issue/show", `{"id":"."}`, "id"},
		{"issue/show", `{"id":".."}`, "id"},
		{"issue/claim", `{"id":"bd-1"}`, "revisionProvenance.actor"},
		{"issue/claim", `{"id":"bd-1","actor":"sam"}`, "actor"},
		{"issue/claim", `{"id":"bd-1","actor":null}`, "actor"},
		{"issue/claim", `{"id":"bd-1","actor":7}`, "actor"},
		{"issue/dep/tree", `{}`, "root_id"},
		{"issue/show", `{"id":"bd-1","include_comments":null}`, "include_comments"},
		{"issue/show", `{"id":"bd-1","include_comments":"perhaps"}`, "include_comments"},
		{"issue/list", `{"label":{"a":1}}`, "label"},
		{"issue/list", `{"label":[["a"]]}`, "label"},
	} {
		body := wantProblem(t, ts.call(t, tc.name, tc.arguments), http.StatusUnprocessableEntity, bdpValidationFailed)
		wantDiagnosticNaming(t, body, tc.param)
	}
	claimer.mu.Lock()
	claims := len(claimer.claims)
	claimer.mu.Unlock()
	if len(reader.getRequests()) != 0 || claims != 0 {
		t.Error("a rejected argument reached a role")
	}
}

func TestCallableQueryArguments(t *testing.T) {
	reader := &roleReader{page: issueops.IssuePage{}}
	ts := newBDPServer(t, rolesConfig(Config{Reader: reader}))

	// sort=priority is issue/list's default, and not v0's.
	direct := ts.get(t, "/v0/beads/issues?label=a&label=b%26c&limit=5&all=true&status=open&sort=priority")
	routed := ts.call(t, "issue/list", `{"label":["a","b&c"],"limit":5,"all":true,"status":["open"]}`)
	directAndRouted(t, direct, routed)
	reader.mu.Lock()
	defer reader.mu.Unlock()
	if len(reader.list) != 2 || !reflect.DeepEqual(reader.list[0], reader.list[1]) {
		t.Errorf("the role saw different requests: %+v", reader.list)
	}
}

// TestCallablePathArgumentsAreEscaped: an id cannot climb out of its segment.
func TestCallablePathArgumentsAreEscaped(t *testing.T) {
	settings := &roleSettings{}
	reader := &roleReader{err: fmt.Errorf("get: %w", issueops.ErrNotFound)}
	ts := newBDPServer(t, rolesConfig(Config{Settings: settings, Reader: reader}))
	for _, id := range []string{"a/b", "a?x=1", "a#b", "a b", "a:close"} {
		encoded, _ := json.Marshal(map[string]string{"id": id})
		wantProblem(t, ts.call(t, "issue/show", string(encoded)), http.StatusNotFound, bdpResourceNotFound)
		got := reader.getRequests()
		if len(got) == 0 || got[len(got)-1].ID != id {
			t.Errorf("id %q did not reach the role whole: %+v", id, got)
		}
	}
}

// TestCallableReasonTakesV0sBounds: a Callable runs on its v0 handler, so a close or
// reopen reason with a control character, or past what v0's reason column holds, is
// refused here as it is there, with v0's refusal carried, and the role never runs.
func TestCallableReasonTakesV0sBounds(t *testing.T) {
	long := strings.Repeat("x", 300)
	for _, tc := range []struct{ name, wire string }{
		{"a trailing newline", `done\n`},
		{"several lines", `two\nlines`},
		{"past the 255-character columns", long},
	} {
		t.Run(tc.name, func(t *testing.T) {
			wantRefused := func(resp *http.Response, callable, param string) {
				t.Helper()
				body := wantProblem(t, resp, http.StatusUnprocessableEntity, bdpValidationFailed)
				if carried, _ := body["v0"].(map[string]any); carried["param"] != param {
					t.Errorf("%s: %v, want v0's refusal of `%s`", callable, body, param)
				}
			}
			arguments := `{"id":"bd-1","reason":"` + tc.wire + `"}`

			closing := &roleLifecycle{closeResult: issueops.CloseResult{Issue: closedIssue("bd-1"), Changed: true}}
			wantRefused(newBDPServer(t, rolesConfig(Config{Lifecycle: closing})).callAs(t, "issue/close", "alice", arguments), "issue/close", "reason")
			if got := closing.closeRequests(); len(got) != 0 {
				t.Errorf("issue/close: the role ran: %+v", got)
			}

			reopening := &roleLifecycle{reopenResult: issueops.ReopenResult{Issue: reopenedIssue("bd-1"), Changed: true}}
			wantRefused(newBDPServer(t, rolesConfig(Config{Lifecycle: reopening})).callAs(t, "issue/reopen", "alice", arguments), "issue/reopen", "reason")
			if got := reopening.reopenRequests(); len(got) != 0 {
				t.Errorf("issue/reopen: the role ran: %+v", got)
			}

			closer := &roleBatchCloser{result: issueops.CloseBatchResult{Outcomes: []issueops.CloseOutcome{{IssueID: "bd-1", Issue: closedIssue("bd-1"), Changed: true}}}}
			wantRefused(newBDPServer(t, rolesConfig(Config{BatchCloser: closer})).callAs(t, "issue/close-many", "alice", `{"items":[{"id":"bd-1","reason":"`+tc.wire+`"}]}`), "issue/close-many", "items[0].reason")
			if got := closer.closeBatchRequests(); len(got) != 0 {
				t.Errorf("issue/close-many: the role ran: %+v", got)
			}
		})
	}
}

func TestCallableFlatDependencyAdapter(t *testing.T) {
	editor := &roleDependencyEditor{}
	ts := newBDPServer(t, rolesConfig(Config{DependencyEditor: editor}))
	lastEdges := func() []issueops.DependencyEdge {
		editor.mu.Lock()
		defer editor.mu.Unlock()
		return editor.adds[len(editor.adds)-1].Edges
	}

	_, envelope := callResult(t, ts.callAs(t, "issue/dep/add", "a", `{"source":"bd-1","target":"bd-2"}`))
	if string(envelope["changed"]) != "true" {
		t.Errorf("changed = %s", envelope["changed"])
	}
	want := issueops.DependencyEdge{IssueID: "bd-1", DependsOnID: "bd-2", Type: types.DepBlocks}
	if got := lastEdges(); len(got) != 1 || got[0].IssueID != want.IssueID || got[0].DependsOnID != want.DependsOnID || got[0].Type != want.Type {
		t.Errorf("omitted type: edges = %+v, want %+v", got, want)
	}

	callResult(t, ts.callAs(t, "issue/dep/add", "a", `{"source":"bd-1","target":"bd-2","type":"related"}`))
	if got := lastEdges(); got[0].Type != types.DependencyType("related") {
		t.Errorf("flat type: edges = %+v", got)
	}

	callResult(t, ts.callAs(t, "issue/dep/add", "a",
		`{"edges":[{"issue_id":"bd-3","depends_on_id":"bd-4","type":"blocks"},{"issue_id":"bd-5","depends_on_id":"bd-6","type":"related"}]}`))
	if got := lastEdges(); len(got) != 2 || got[1].IssueID != "bd-5" {
		t.Errorf("edges form: %+v", got)
	}

	addCount := func() int {
		editor.mu.Lock()
		defer editor.mu.Unlock()
		return len(editor.adds)
	}
	before := addCount()
	for _, tc := range []struct{ arguments, param string }{
		{`{}`, "edges"},
		{`{"edges":[{"issue_id":"x","depends_on_id":"y","type":"blocks"}],"source":"bd-1","target":"bd-2"}`, "source"},
		{`{"edges":[{"issue_id":"x","depends_on_id":"y","type":"blocks"}],"type":"blocks"}`, "type"},
		{`{"source":"bd-1"}`, "target"},
		{`{"target":"bd-1"}`, "source"},
		{`{"type":"blocks"}`, "source"},
		{`{"source":7,"target":"bd-2"}`, "source"},
		{`{"source":"bd-1","target":null}`, "target"},
		{`{"source":"","target":"bd-2"}`, "source"},
		{`{"source":"bd-1","target":""}`, "target"},
		{`{"source":"bd-1","target":"bd-2","type":""}`, "type"},
		{`{"source":"` + strings.Repeat("a", 256) + `","target":"bd-2"}`, "source"},
		{`{"source":"bd-1","target":"` + strings.Repeat("a", 256) + `"}`, "target"},
		{`{"source":"bd-1","target":"bd-2","type":"` + strings.Repeat("a", 256) + `"}`, "type"},
	} {
		body := wantProblem(t, ts.callAs(t, "issue/dep/add", "a", tc.arguments), http.StatusUnprocessableEntity, bdpValidationFailed)
		wantDiagnosticNaming(t, body, tc.param)
	}
	wantDiagnosticNaming(t, wantProblem(t, ts.call(t, "issue/dep/add", `{"source":"bd-1","target":"bd-2"}`),
		http.StatusUnprocessableEntity, bdpValidationFailed), "revisionProvenance.actor")
	if addCount() != before {
		t.Error("a refused form reached the role")
	}
}

// TestCallableCommentIsSignedByTheActor: `author` is no parameter of issue/comment.
// The actor is what v0 stores under `author`, and v0's own rules for that name are
// the check.
func TestCallableCommentIsSignedByTheActor(t *testing.T) {
	commenter := &roleCommenter{comment: storedComment("bd-1")}
	ts := newBDPServer(t, rolesConfig(Config{Commenter: commenter}))

	callResult(t, ts.callAs(t, "issue/comment", "sam", `{"id":"bd-1","text":"hi"}`))
	if got := commenter.commentRequests(); got[len(got)-1].Author != "sam" {
		t.Errorf("the comment is signed %+v, want the invocation's actor", got[len(got)-1])
	}
	callResult(t, ts.callAs(t, "issue/comment", "  sam  ", `{"id":"bd-1","text":"hi"}`))
	if got := commenter.commentRequests(); got[len(got)-1].Author != "sam" {
		t.Errorf("v0 trims what it stores, and it did not here: %+v", got[len(got)-1])
	}

	before := len(commenter.commentRequests())
	wantDiagnosticNaming(t, wantProblem(t, ts.call(t, "issue/comment", `{"id":"bd-1","text":"hi"}`),
		http.StatusUnprocessableEntity, bdpValidationFailed), "revisionProvenance.actor")
	for _, arguments := range []string{
		`{"id":"bd-1","text":"hi","author":"kim"}`,
		`{"id":"bd-1","text":"hi","author":null}`,
		`{"id":"bd-1","text":"hi","actor":"kim"}`,
	} {
		body := wantProblem(t, ts.callAs(t, "issue/comment", "sam", arguments), http.StatusUnprocessableEntity, bdpValidationFailed)
		if body["reason"] != string(ReasonUnknownParameter) {
			t.Errorf("%s: reason = %v, want the undeclared-argument refusal", arguments, body["reason"])
		}
	}
	// v0's refusal, carried: a name that is blank once trimmed.
	body := wantProblem(t, ts.callAs(t, "issue/comment", "   ", `{"id":"bd-1","text":"hi"}`), http.StatusUnprocessableEntity, bdpValidationFailed)
	if carried, _ := body["v0"].(map[string]any); carried["param"] != "author" {
		t.Errorf("a blank actor: %v, want v0's refusal of `author`", body)
	}
	if len(commenter.commentRequests()) != before {
		t.Error("a refused comment reached the role")
	}
}

// TestCallableProvenanceIsHonoredOrRefused: v0 has nowhere to carry
// revisionProvenance but the name it records, so an input it cannot honor is refused
// rather than dropped.
func TestCallableProvenanceIsHonoredOrRefused(t *testing.T) {
	lifecycle := &roleLifecycle{updateResult: issueops.UpdateResult{Issue: updatedIssue("bd-1"), Changed: true}}
	ts := newBDPServer(t, rolesConfig(Config{Lifecycle: lifecycle}))
	args := `"arguments":{"id":"bd-1","patch":{"title":"t"}}`

	callResult(t, ts.callSend(t, "issue/update",
		`{`+args+`,"revisionProvenance":{"actor":{"id":"sam","basis":"writer-supplied"},"agent":null,"message":null}}`, nil))
	if got := lifecycle.updateRequests(); got[len(got)-1].Actor != "sam" {
		t.Errorf("the role was told the actor is %q", got[len(got)-1].Actor)
	}
	callResult(t, ts.callSend(t, "issue/update", `{`+args+`,"revisionProvenance":{"actor":{"id":"kim","basis":"unknown"}}}`, nil))
	if got := lifecycle.updateRequests(); got[len(got)-1].Actor != "kim" {
		t.Errorf("a basis of unknown: the role was told the actor is %q", got[len(got)-1].Actor)
	}

	before := len(lifecycle.updateRequests())
	// `revisionProvenance` is present and names nobody, which is no actor at all.
	wantDiagnosticNaming(t, wantProblem(t, ts.callSend(t, "issue/update", `{`+args+`,"revisionProvenance":{}}`, nil),
		http.StatusUnprocessableEntity, bdpValidationFailed), "revisionProvenance.actor")
	for _, provenance := range []string{
		`{"actor":{"id":"sam","basis":"writer-supplied"},"agent":"claude"}`,
		`{"actor":{"id":"sam","basis":"writer-supplied"},"message":"why"}`,
	} {
		wantDiagnosticNaming(t, wantProblem(t, ts.callSend(t, "issue/update", `{`+args+`,"revisionProvenance":`+provenance+`}`, nil),
			http.StatusUnprocessableEntity, bdpValidationFailed), "revisionProvenance")
	}
	for _, provenance := range []string{
		`null`,
		`[]`,
		`{"who":"x"}`,
		`{"actor":"sam"}`,
		`{"actor":{"id":"sam"}}`,
		`{"actor":{"id":"","basis":"unknown"}}`,
		`{"actor":{"id":"sam","basis":"attested"}}`,
		`{"actor":{"id":"sam","basis":"unknown","extra":1}}`,
		`{"agent":""}`,
		`{"message":7}`,
	} {
		wantProblem(t, ts.callSend(t, "issue/update", `{`+args+`,"revisionProvenance":`+provenance+`}`, nil),
			http.StatusBadRequest, bdpMalformedRequest)
	}
	// What the id must be is v0's to say, as it was when the same name came in `actor`.
	for _, id := range []string{`   `, `a\u0007b`, strings.Repeat("x", 300)} {
		body := wantProblem(t, ts.callSend(t, "issue/update",
			`{`+args+`,"revisionProvenance":{"actor":{"id":"`+id+`","basis":"writer-supplied"}}}`, nil),
			http.StatusUnprocessableEntity, bdpValidationFailed)
		if carried, _ := body["v0"].(map[string]any); carried["param"] != "actor" {
			t.Errorf("id %q: %v, want v0's refusal of `actor`", id, body)
		}
	}
	if len(lifecycle.updateRequests()) != before {
		t.Error("a refused provenance reached the role")
	}
}

// TestCallableActorIsDeclaredNowhere is §2 and §3 across the table: `actor` inside
// `arguments` is any other undeclared argument, an operation that requires an actor
// refuses its absence in words this caller can act on, the three that do not run
// without one, and the four that record none refuse one.
func TestCallableActorIsDeclaredNowhere(t *testing.T) {
	var reached atomic.Int32
	original := routeTable
	t.Cleanup(func() { routeTable = original })
	swapped := slices.Clone(original)
	for i := range swapped {
		swapped[i].handler = func(_ *Server, w http.ResponseWriter, _ *http.Request) {
			reached.Add(1)
			writeJSON(w, map[string]string{})
		}
	}
	routeTable = swapped
	ts := newBDPServer(t, rolesConfig(Config{}))

	runsWithout := []string{"issue/purge", "issue/prune", "issue/delete"}
	refusesOne := []string{"config/set", "config/unset", "memory/remember", "memory/forget"}
	attributed := 0
	for i := range callableTable {
		row := &callableTable[i]
		if !row.batch {
			continue
		}
		t.Run(row.name, func(t *testing.T) {
			args := map[string]any{}
			for _, name := range row.path {
				args[name] = "v-" + name
			}
			if row.name == "issue/dep/add" {
				args["source"], args["target"] = "v-source", "v-target"
			}
			arguments := func(extra ...string) string {
				withExtra := maps.Clone(args)
				for _, name := range extra {
					withExtra[name] = "kim"
				}
				encoded, _ := json.Marshal(withExtra)
				return string(encoded)
			}
			for _, name := range []string{"actor", "author"} {
				before := reached.Load()
				resp := ts.call(t, row.name, arguments(name))
				if row.actorMember != "" {
					resp = ts.callAs(t, row.name, "sam", arguments(name))
				}
				body := wantProblem(t, resp, http.StatusUnprocessableEntity, bdpValidationFailed)
				wantDiagnosticNaming(t, body, name)
				if body["reason"] != string(ReasonUnknownParameter) {
					t.Errorf("`%s` in arguments: reason = %v", name, body["reason"])
				}
				if reached.Load() != before {
					t.Errorf("`%s` in arguments reached v0", name)
				}
			}

			before := reached.Load()
			switch {
			case slices.Contains(refusesOne, row.name):
				if row.actorMember != "" {
					t.Errorf("actorMember = %q", row.actorMember)
				}
				wantDiagnosticNaming(t, wantProblem(t, ts.callAs(t, row.name, "sam", arguments()),
					http.StatusUnprocessableEntity, bdpValidationFailed), "revisionProvenance")
				if reached.Load() != before {
					t.Error("a refused actor reached v0")
				}
				callResult(t, ts.call(t, row.name, arguments()))
			case slices.Contains(runsWithout, row.name):
				attributed++
				callResult(t, ts.call(t, row.name, arguments()))
				callResult(t, ts.callAs(t, row.name, "sam", arguments()))
			default:
				attributed++
				body := wantProblem(t, ts.call(t, row.name, arguments()), http.StatusUnprocessableEntity, bdpValidationFailed)
				wantDiagnosticNaming(t, body, "revisionProvenance.actor")
				if _, carried := body["v0"]; carried {
					t.Errorf("v0 ran: %v", body)
				}
				if reached.Load() != before {
					t.Error("an invocation with no actor reached v0")
				}
				callResult(t, ts.callAs(t, row.name, "sam", arguments()))
			}
		})
	}
	if attributed != 20 {
		t.Errorf("%d operations take an actor, want the 19 with `actor` and issue/comment", attributed)
	}
}

// TestCallableServiceDefaults is §4: an omitted parameter gets its declared value at
// the route, and neither a supplied value nor `null` is replaced. The direct route is
// held to v0's own omission alongside it.
func TestCallableServiceDefaults(t *testing.T) {
	type seen struct {
		query map[string][]string
		body  map[string]json.RawMessage
	}
	var last atomic.Pointer[seen]
	original := routeTable
	t.Cleanup(func() { routeTable = original })
	swapped := slices.Clone(original)
	for i := range swapped {
		swapped[i].handler = func(_ *Server, w http.ResponseWriter, r *http.Request) {
			got := &seen{query: r.URL.Query()}
			if data, _ := io.ReadAll(r.Body); len(data) > 0 {
				if err := json.Unmarshal(data, &got.body); err != nil {
					panic(err)
				}
			}
			last.Store(got)
			writeJSON(w, map[string]string{})
		}
	}
	routeTable = swapped
	ts := newBDPServer(t, rolesConfig(Config{}))

	want := map[string]map[string]string{
		"issue/list":     {"sort": `"priority"`},
		"issue/dep/list": {"direction": `"out"`},
		"issue/close":    {"reason": `"Closed"`},
		"issue/create":   {"issue_type": `"task"`},
		"issue/purge":    {"dry_run": `true`, "tier": `"ephemeral"`, "protect_referenced": `false`, "protect_live_dependents": `true`},
		"issue/prune":    {"dry_run": `true`, "tier": `"durable"`, "protect_referenced": `true`, "protect_live_dependents": `false`},
		"issue/delete":   {"dry_run": `true`},
		"issue/reclaim":  {"older_than_seconds": `600`},
	}
	for i := range callableTable {
		row := &callableTable[i]
		declared, ok := want[row.name]
		if len(row.defaults) != len(declared) {
			t.Errorf("%s: %d defaults, want %d", row.name, len(row.defaults), len(declared))
		}
		if !ok {
			continue
		}
		t.Run(row.name, func(t *testing.T) {
			invoke := func(extra map[string]any) *http.Response {
				args := map[string]any{}
				for _, name := range row.path {
					args[name] = "v-" + name
				}
				maps.Copy(args, extra)
				encoded, _ := json.Marshal(args)
				if row.actorMember != "" {
					return ts.callAs(t, row.name, "sam", string(encoded))
				}
				return ts.call(t, row.name, string(encoded))
			}
			valueOf := func(name string) string {
				got := last.Load()
				if slices.Contains(row.query, name) {
					return strings.Join(got.query[name], ",")
				}
				return string(got.body[name])
			}

			last.Store(nil)
			callResult(t, invoke(nil))
			for name, value := range declared {
				wanted := value
				if slices.Contains(row.query, name) {
					_ = json.Unmarshal([]byte(value), &wanted)
				}
				if got := valueOf(name); got != wanted {
					t.Errorf("omitted %s reached v0 as %q, want %q", name, got, wanted)
				}
			}

			for name := range declared {
				callResult(t, invoke(map[string]any{name: "supplied"}))
				supplied := `"supplied"`
				if slices.Contains(row.query, name) {
					supplied = "supplied"
				}
				if got := valueOf(name); got != supplied {
					t.Errorf("supplied %s reached v0 as %q, want %q", name, got, supplied)
				}
				// null is a supplied value (§4). A body member goes to v0 as it is, which
				// refuses it. A query has no spelling of null, and is refused here.
				last.Store(nil)
				resp := invoke(map[string]any{name: nil})
				if slices.Contains(row.query, name) {
					wantDiagnosticNaming(t, wantProblem(t, resp, http.StatusUnprocessableEntity, bdpValidationFailed), name)
					if last.Load() != nil {
						t.Errorf("null %s reached v0", name)
					}
					continue
				}
				callResult(t, resp)
				if got := valueOf(name); got != "null" {
					t.Errorf("null %s reached v0 as %q", name, got)
				}
			}
		})
	}
}

// TestV0ProblemSurfacesAsItsDeclaredError drives real handlers into real v0
// refusals, and compares each to the same refusal answered directly.
func TestV0ProblemSurfacesAsItsDeclaredError(t *testing.T) {
	t.Run("a reason is callable-error with v0's Problem verbatim", func(t *testing.T) {
		claimer := &roleClaimer{err: &issueops.ClaimConflictError{Err: issueops.ErrAlreadyClaimed, Assignee: "kim"}}
		ts := newBDPServer(t, rolesConfig(Config{Claimer: claimer}))
		directResp := ts.postBody(t, "/v0/beads/issues/bd-1:claim", "application/json", `{"actor":"sam"}`)
		direct := decodeBody(t, directResp)
		if direct["code"] != string(CodeAlreadyClaimed) {
			t.Fatalf("direct code = %v", direct)
		}
		body := wantProblem(t, ts.callAs(t, "issue/claim", "sam", `{"id":"bd-1"}`), http.StatusConflict, bdpCallableError)
		inner, ok := body["error"].(map[string]any)
		if !ok {
			t.Fatalf("no error member: %v", body)
		}
		delete(direct, "request_id")
		if inner["request_id"] == nil {
			t.Error("the verbatim Problem lost its request_id")
		}
		delete(inner, "request_id")
		if !reflect.DeepEqual(inner, direct) {
			t.Errorf("error is not v0's Problem:\n direct %v\n  error %v", direct, inner)
		}
		if inner["assignee"] != "kim" {
			t.Errorf("assignee = %v", inner["assignee"])
		}
		if slices.Contains(callableIndex["issue/claim"].reasons, "already_claimed") == false {
			t.Error("issue/claim does not declare already_claimed")
		}
	})

	t.Run("not_found is resource-not-found", func(t *testing.T) {
		reader := &roleReader{err: fmt.Errorf("get: %w", issueops.ErrNotFound)}
		ts := newBDPServer(t, rolesConfig(Config{Reader: reader}))
		body := wantProblem(t, ts.call(t, "issue/show", `{"id":"bd-1"}`), http.StatusNotFound, bdpResourceNotFound)
		if _, has := body["error"]; has {
			t.Errorf("only callable-error carries `error`: %v", body)
		}
		if v0, _ := body["v0"].(map[string]any); v0["code"] != string(CodeNotFound) {
			t.Errorf("v0 = %v", body["v0"])
		}
	})

	t.Run("a stale expected_version is revision-mismatch, another guard a reason", func(t *testing.T) {
		lifecycle := &roleLifecycle{updateErr: fmt.Errorf("update: %w", issueops.ErrVersionMismatch)}
		ts := newBDPServer(t, rolesConfig(Config{Lifecycle: lifecycle}))
		wantProblem(t, ts.callAs(t, "issue/update", "a", `{"id":"bd-1","patch":{"title":"t"},"expected_version":"7"}`),
			http.StatusConflict, bdpRevisionMismatch)

		lifecycle.updateErr = fmt.Errorf("update: %w", issueops.ErrStatusMismatch)
		body := wantProblem(t, ts.callAs(t, "issue/update", "a", `{"id":"bd-1","patch":{"title":"t"},"expected_status":"open"}`),
			http.StatusConflict, bdpCallableError)
		if inner, _ := body["error"].(map[string]any); inner["code"] != string(CodePreconditionFailed) {
			t.Errorf("error = %v", body["error"])
		}
	})

	t.Run("a role's system failure is temporarily-unavailable, with no error member", func(t *testing.T) {
		reader := &roleReader{err: errors.New("dial tcp 10.0.0.1:3306: secret-dsn")}
		ts := newBDPServer(t, rolesConfig(Config{Reader: reader}))
		resp := ts.call(t, "issue/show", `{"id":"bd-1"}`)
		raw := readAll(t, resp)
		if resp.StatusCode != http.StatusServiceUnavailable || strings.Contains(raw, "secret-dsn") {
			t.Fatalf("status %d: %s", resp.StatusCode, raw)
		}
		var body map[string]any
		_ = json.Unmarshal([]byte(raw), &body)
		if body["code"] != bdpUnavailable || body["error"] != nil {
			t.Errorf("body = %s", raw)
		}
	})

	t.Run("a panicking handler answers the same way and is logged", func(t *testing.T) {
		original := routeTable
		t.Cleanup(func() { routeTable = original })
		swapped := slices.Clone(original)
		for i := range swapped {
			if swapped[i].op == OpGetIssue {
				swapped[i].handler = func(*Server, http.ResponseWriter, *http.Request) { panic("boom-marker") }
			}
		}
		routeTable = swapped
		ts := newBDPServer(t, rolesConfig(Config{}))
		body := wantProblem(t, ts.call(t, "issue/show", `{"id":"bd-1"}`), http.StatusServiceUnavailable, bdpUnavailable)
		if !strings.Contains(ts.stderr.String(), "boom-marker") {
			t.Error("the panic left no log line")
		}
		// What a direct call would have said, so the client reads a server fault it
		// does not retry and a request id that finds this log line.
		v0, _ := body["v0"].(map[string]any)
		if v0["code"] != string(CodeInternal) || v0["status"] != float64(http.StatusInternalServerError) || v0["request_id"] == "" || v0["request_id"] == nil {
			t.Errorf("v0 = %v", body["v0"])
		}
		if strings.Contains(fmt.Sprint(body), "boom-marker") {
			t.Errorf("the panic text reached the answer: %v", body)
		}
	})

	t.Run("the request line books the v0 operation", func(t *testing.T) {
		claimer := &roleClaimer{err: &issueops.ClaimConflictError{Err: issueops.ErrAlreadyClaimed, Assignee: "kim"}}
		ts := newBDPServer(t, rolesConfig(Config{Claimer: claimer}))
		wantProblem(t, ts.callAs(t, "issue/claim", "sam", `{"id":"bd-1"}`), http.StatusConflict, bdpCallableError)
		line := findLogLine(t, ts.stderr.String(), "path=/bdp/call/issue/claim")
		if !strings.Contains(line, "op="+OpClaimIssue) {
			t.Errorf("request line = %s", line)
		}
	})
}

// TestEveryV0CodeHasAnAnswer is §5 as a table, over every code the server can
// produce, including the inner status that differs from the invocation's 409.
func TestEveryV0CodeHasAnAnswer(t *testing.T) {
	want := map[Code]string{
		CodeInvalidArgument:        bdpValidationFailed,
		CodeInvalidCursor:          bdpInvalidParameter,
		CodeUnauthenticated:        bdpUnauthenticated,
		CodeNotFound:               bdpResourceNotFound,
		CodeAlreadyExists:          bdpIdentityTaken,
		CodeIdempotencyExpired:     bdpIdentityTaken,
		CodeBusy:                   bdpUnavailable,
		CodeDBUnavailable:          bdpUnavailable,
		CodeInternal:               bdpUnavailable,
		CodeAlreadyClaimed:         bdpCallableError,
		CodeNotClaimable:           bdpCallableError,
		CodeNotClosable:            bdpCallableError,
		CodeNotesOverwrite:         bdpCallableError,
		CodeTemplateReadOnly:       bdpCallableError,
		CodeIssuePinned:            bdpCallableError,
		CodeNotAssignee:            bdpCallableError,
		CodeNotReleasable:          bdpCallableError,
		CodeLeaseNotHeld:           bdpCallableError,
		CodeDependencyCycle:        bdpCallableError,
		CodeDependencyExists:       bdpCallableError,
		CodePreconditionFailed:     bdpCallableError,
		CodeEventsJournalDisabled:  bdpCallableError,
		CodeEventsJournalTruncated: bdpCallableError,
		CodeEventsWatchSaturated:   bdpCallableError,
	}
	for code := range codeStatus {
		expected, ok := want[code]
		if !ok {
			t.Errorf("v0 code %s has no expectation here", code)
			continue
		}
		body, _ := json.Marshal(newResult(code, "d").Problem)
		problem := translateV0Problem(body)
		if problem.Code != expected {
			t.Errorf("%s -> %s, want %s", code, problem.Code, expected)
		}
		if (problem.Code == bdpCallableError) != (problem.Error != nil) {
			t.Errorf("%s: `error` present = %v on %s", code, problem.Error != nil, problem.Code)
		}
		if problem.Status != bdpProblemShapes[expected].status {
			t.Errorf("%s: status = %d", code, problem.Status)
		}
		if problem.Code == bdpValidationFailed && len(problem.Diagnostics) == 0 {
			t.Errorf("%s: validation-failed must carry diagnostics", code)
		}
	}

	body, _ := json.Marshal(newResult(CodeEventsJournalTruncated, "gap").Problem)
	problem := translateV0Problem(body)
	var inner map[string]any
	_ = json.Unmarshal(problem.Error, &inner)
	if problem.Status != http.StatusConflict || inner["status"] != float64(http.StatusGone) {
		t.Errorf("outer %d, inner %v: §5 keeps v0's own status inside `error`", problem.Status, inner["status"])
	}
}

// TestCallablesGateBeforeInterpretation: the credential, the project stamp and the
// wire floor are the route's own, and a refusal comes before any of the body is
// read, so an unauthenticated caller learns nothing about names or shapes.
func TestCallablesGateBeforeInterpretation(t *testing.T) {
	ts := newBDPServer(t, rolesConfig(Config{Auth: authTokenFile(t, "real-token")}))

	for _, name := range []string{"issue/show", "issue/nope"} {
		resp := ts.callSend(t, name, `not json`, nil)
		wantProblem(t, resp, http.StatusUnauthorized, bdpUnauthenticated)
		if resp.Header.Get("WWW-Authenticate") != "Bearer" {
			t.Errorf("%s: no challenge on the 401", name)
		}
	}
	wantProblem(t, ts.do(t, http.MethodGet, callableMount, nil), http.StatusUnauthorized, bdpUnauthenticated)

	authed := bearer("real-token")
	wantProblem(t, ts.callSend(t, "issue/nope", `{"arguments":{}}`, authed), http.StatusNotFound, bdpCallableNotFound)
	if resp := ts.do(t, http.MethodGet, callableMount, authed); resp.StatusCode != http.StatusOK {
		t.Errorf("GET /bdp/call with the credential = %d", resp.StatusCode)
	}

	ctx := http.Header{"Authorization": authed["Authorization"], ProjectIDHeader: {"not-this-project"}}
	body := wantProblem(t, ts.callSend(t, "issue/show", `{"arguments":{"id":"bd-1"}}`, ctx),
		http.StatusUnprocessableEntity, bdpValidationFailed)
	wantDiagnosticNaming(t, body, ProjectIDHeader)

	// The handshake is exempt, as v0's row is: it is how a client learns the id, and
	// what its rich mismatch error (database, repo root) is built from.
	result, _ := callResult(t, ts.callSend(t, "issue/context", `{"arguments":{}}`, ctx))
	var served map[string]any
	if err := json.Unmarshal(result, &served); err != nil {
		t.Fatal(err)
	}
	if _, ok := served["project_id"]; !ok {
		t.Errorf("issue/context answered no project_id: %s", result)
	}
}

// TestAnInvocationRunsItsRouteOnce: one credential check per request, not one per
// layer. Every other check the route does costs the same, and this one can be counted:
// with the reload gate open, each Verify reports the token file it could not read.
func TestAnInvocationRunsItsRouteOnce(t *testing.T) {
	ts := newBDPServer(t, rolesConfig(Config{Auth: authTokenFile(t, "real-token")}))
	ts.Server.auth.reloadGate = time.Nanosecond
	if err := os.Remove(ts.Server.auth.path); err != nil {
		t.Fatal(err)
	}
	before := strings.Count(ts.stderr.String(), "auth_reload_error")
	resp := ts.callSend(t, "issue/context", `{"arguments":{}}`, bearer("real-token"))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d: %s", resp.StatusCode, readAll(t, resp))
	}
	_ = readAll(t, resp)
	if got := strings.Count(ts.stderr.String(), "auth_reload_error") - before; got != 1 {
		t.Errorf("the credential was checked %d times for one invocation", got)
	}
}

// countingBody records whether anyone asked it for bytes.
type countingBody struct {
	io.Reader
	reads *atomic.Int32
}

func (b countingBody) Read(p []byte) (int, error) {
	b.reads.Add(1)
	return b.Reader.Read(p)
}

// TestAnInvocationsBodyIsReadInsideItsSlot is what keeps the memory bound at
// maxInflight: a request waiting for a slot holds no body. It is refused for being
// busy, the way a direct v0 call is, without having read a byte.
func TestAnInvocationsBodyIsReadInsideItsSlot(t *testing.T) {
	ts := newBDPServer(t, rolesConfig(Config{}), func(s *Server) { s.semTimeout = 50 * time.Millisecond })
	for range cap(ts.Server.sem) {
		ts.Server.sem <- struct{}{}
	}
	t.Cleanup(func() {
		for range cap(ts.Server.sem) {
			<-ts.Server.sem
		}
	})

	var reads atomic.Int32
	body := countingBody{Reader: strings.NewReader(`{"arguments":{"id":"bd-1"}}`), reads: &reads}
	invoke := func(name string, body io.Reader, media string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, callableMount+"/"+name, body)
		req.Host = "127.0.0.1"
		req.Header.Set("Content-Type", media)
		rec := httptest.NewRecorder()
		ts.Server.handler().ServeHTTP(rec, req)
		return rec
	}
	rec := invoke("issue/show", body, bdpMediaType)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body)
	}
	var problem map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &problem); err != nil {
		t.Fatal(err)
	}
	if v0, _ := problem["v0"].(map[string]any); v0["code"] != string(CodeBusy) {
		t.Errorf("answer = %s", rec.Body)
	}
	if got := reads.Load(); got != 0 {
		t.Errorf("the body was read %d times outside a slot", got)
	}

	t.Run("a refused media type costs no slot", func(t *testing.T) {
		rec := invoke("issue/show", strings.NewReader(`{}`), "application/json")
		if rec.Code != http.StatusUnsupportedMediaType {
			t.Errorf("status = %d, want 415 at saturation: %s", rec.Code, rec.Body)
		}
	})
	t.Run("issue/context holds no slot", func(t *testing.T) {
		rec := invoke("issue/context", strings.NewReader(`{"arguments":{}}`), bdpMediaType)
		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want the handshake to answer at saturation: %s", rec.Code, rec.Body)
		}
	})
}

// TestIssueContextReadsOnlyTheEnvelopesSlack: it holds no slot, so nothing but its own
// limit bounds what it reads.
func TestIssueContextReadsOnlyTheEnvelopesSlack(t *testing.T) {
	ts := newBDPServer(t, rolesConfig(Config{}))
	huge := `{"arguments":{},"revisionProvenance":"` + strings.Repeat("x", callableBodySlack) + `"}`
	wantProblem(t, ts.callSend(t, "issue/context", huge, nil), http.StatusRequestEntityTooLarge, bdpRequestTooLarge)
}

// TestScanInvocationKeepsTheBodysOwnBytes: the members it answers are spans of the one
// buffer, whatever the spacing, so nothing about an invocation is held twice.
func TestScanInvocationKeepsTheBodysOwnBytes(t *testing.T) {
	body := []byte(" {\n \"arguments\" :\t{ \"id\" : \"bd-1\" , \"nested\" : [ {\"a\":1} , \"s\" ] , \"n\":-1.5e3 , \"null\":null } , \"revisionProvenance\": {\"actor\":{\"id\":\"a\",\"basis\":\"unknown\"}} } ")
	members, args, detail := scanInvocation(body)
	if detail != "" {
		t.Fatalf("detail = %q", detail)
	}
	for name, want := range map[string]string{
		"id": `"bd-1"`, "nested": `[ {"a":1} , "s" ]`, "n": `-1.5e3`, "null": `null`,
	} {
		if got := string(args[name]); got != want {
			t.Errorf("args[%s] = %q, want %q", name, got, want)
		}
	}
	if got := string(members["revisionProvenance"]); got != `{"actor":{"id":"a","basis":"unknown"}}` {
		t.Errorf("revisionProvenance = %q", got)
	}
	var whole map[string]any
	if err := json.Unmarshal(members["arguments"], &whole); err != nil || len(whole) != 4 {
		t.Errorf("arguments = %q: %v", members["arguments"], err)
	}
	if cap(args["id"]) != len(args["id"]) {
		t.Errorf("a member can grow into its neighbor: cap %d, len %d", cap(args["id"]), len(args["id"]))
	}

	for _, malformed := range []string{``, `[]`, `1`, `null`, `"x"`, `{} {}`, `{"a":1}x`, `{"a":1,}`, `{"a":1,"a":2}`, `{"arguments":{"x":1,"x":2}}`} {
		if members, _, detail := scanInvocation([]byte(malformed)); detail == "" && members != nil {
			t.Errorf("%q passed the scan: %v", malformed, members)
		}
	}
	if _, args, detail := scanInvocation([]byte(`{"arguments":[]}`)); detail != "" || args != nil {
		t.Errorf("arguments that are no object: args %v, detail %q", args, detail)
	}
}

// TestV0BodyIsTheRequestV0WouldHaveReceived: spliced rather than re-encoded, so it has
// to come out as the same JSON either way.
func TestV0BodyIsTheRequestV0WouldHaveReceived(t *testing.T) {
	bodyOf := func(row *callableRow, document string) map[string]any {
		t.Helper()
		_, args, _ := scanInvocation([]byte(document))
		inv := &invocation{args: args}
		if detail := func() string {
			members, _, detail := scanInvocation([]byte(document))
			inv.arguments = members["arguments"]
			return detail
		}(); detail != "" {
			t.Fatal(detail)
		}
		if problem := row.supplyActor(inv, json.RawMessage(`"a b"`)); problem != nil {
			t.Fatal(problem.Detail)
		}
		row.supplyDefaults(inv)
		head, rest := row.v0Body(inv)
		var got map[string]any
		if err := json.Unmarshal(append(slices.Clone(head), rest...), &got); err != nil {
			t.Fatalf("%s%s: %v", head, rest, err)
		}
		return got
	}
	got := bodyOf(callableIndex["issue/update"], `{"arguments":{"id":"bd-1","patch":{"title":"t\"x"}}}`)
	if want := (map[string]any{"actor": "a b", "patch": map[string]any{"title": `t"x`}}); !reflect.DeepEqual(got, want) {
		t.Errorf("a path and a body: %v, want %v", got, want)
	}
	got = bodyOf(callableIndex["issue/batch"], `{"arguments": { "items" : [] } }`)
	if want := (map[string]any{"actor": "a b", "items": []any{}}); !reflect.DeepEqual(got, want) {
		t.Errorf("a body-only Callable: %v, want %v", got, want)
	}
	got = bodyOf(callableIndex["issue/delete"], `{"arguments":{ }}`)
	if want := (map[string]any{"actor": "a b", "dry_run": true}); !reflect.DeepEqual(got, want) {
		t.Errorf("nothing but what the layer supplied: %v, want %v", got, want)
	}

	apply := callableIndex["issue/batch"]
	arguments := json.RawMessage(`{"items":[]}`)
	if head, rest := apply.v0Body(&invocation{arguments: arguments}); len(head) != 0 || &rest[0] != &arguments[0] {
		t.Error("a body-only Callable copied its arguments")
	}
	inv := &invocation{arguments: arguments, args: map[string]json.RawMessage{"items": json.RawMessage(`[]`)}}
	inv.supply("actor", json.RawMessage(`"a"`))
	if _, rest := apply.v0Body(inv); &rest[0] != &arguments[1] {
		t.Error("supplying the actor copied a body-only Callable's arguments")
	}
}

// TestCallableListIsTheOpenAPIProjection is Callables §3 and §14 over the served
// document.
func TestCallableListIsTheOpenAPIProjection(t *testing.T) {
	ts := newBDPServer(t, rolesConfig(Config{}))
	resp := ts.get(t, callableMount)
	raw := readAll(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d: %s", resp.StatusCode, raw)
	}
	if got := resp.Header.Get("Content-Type"); got != callableDocMediaType {
		t.Errorf("Content-Type = %q", got)
	}
	if got := resp.Header.Get("Cache-Control"); got != "private, no-store" {
		t.Errorf("Cache-Control = %q", got)
	}
	head := ts.do(t, http.MethodHead, callableMount, nil)
	if head.StatusCode != http.StatusOK || readAll(t, head) != "" {
		t.Errorf("HEAD = %d", head.StatusCode)
	}

	var doc map[string]any
	if err := json.Unmarshal([]byte(raw), &doc); err != nil {
		t.Fatalf("document is not JSON: %v", err)
	}
	if doc["openapi"] != "3.1.0" {
		t.Errorf("openapi = %v", doc["openapi"])
	}
	paths := mapAt(t, doc, "paths")
	if len(paths) != len(callableTable) {
		t.Errorf("%d paths for %d names", len(paths), len(callableTable))
	}
	for path, item := range paths {
		post, _ := item.(map[string]any)["post"].(map[string]any)
		extension, _ := post["x-bdp-callable"].(map[string]any)
		if extension["compiled"] != true {
			t.Errorf("%s isn't marked compiled: %v", path, extension)
		}
	}
	for _, name := range []string{"source", "target", "type"} {
		schema, _ := adapterParameterSchemas[name].(map[string]any)
		if schema["minLength"] != 1 || schema["maxLength"] != 255 {
			t.Errorf("issue/dep/add %s = %v, want the declaration's bounds 1 and 255", name, schema)
		}
	}
	v0 := loadSpec(t)
	ops := specOps(t, v0)
	ids := map[string]bool{}
	for i := range callableTable {
		row := &callableTable[i]
		item, ok := paths[callableMount+"/"+row.name].(map[string]any)
		if !ok {
			t.Errorf("no path for %s", row.name)
			continue
		}
		if len(item) != 1 {
			t.Errorf("%s: a mount admits POST only: %v", row.name, item)
		}
		op := mapAt(t, item, "post")
		id, _ := op["operationId"].(string)
		if ids[id] {
			t.Errorf("operationId %q is used twice", id)
		}
		ids[id] = true
		if op["summary"] != ops[row.op].op["summary"] {
			t.Errorf("%s: summary = %v", row.name, op["summary"])
		}

		schema := mapAt(t, mapAt(t, mapAt(t, mapAt(t, op, "requestBody"), "content"), bdpMediaType), "schema")
		if schema["additionalProperties"] != false {
			t.Errorf("%s: the invocation must be closed", row.name)
		}
		invocation := mapAt(t, schema, "properties")
		if _, has := invocation["revisionProvenance"]; has != row.batch {
			t.Errorf("%s: revisionProvenance present = %v", row.name, has)
		}
		arguments := mapAt(t, invocation, "arguments")
		if arguments["additionalProperties"] != false {
			t.Errorf("%s: arguments must be closed", row.name)
		}
		properties := mapAt(t, arguments, "properties")
		if len(properties) != len(row.declaredParameters()) {
			t.Errorf("%s: %d properties for parameters %v", row.name, len(properties), row.declaredParameters())
		}

		_, hasHeader := op["parameters"]
		if hasHeader != row.batch {
			t.Errorf("%s: Idempotency-Key present = %v", row.name, hasHeader)
		}
		ok200 := mapAt(t, mapAt(t, mapAt(t, mapAt(t, op, "responses"), "200"), "content"), bdpMediaType)
		result := mapAt(t, mapAt(t, ok200, "schema"), "properties")
		if _, has := result["changed"]; has != row.batch {
			t.Errorf("%s: changed declared = %v", row.name, has)
		}

		extension := mapAt(t, op, "x-bdp-callable")
		wantType := callableTypeBase + "callable-query"
		if row.batch {
			wantType = callableTypeBase + "callable-batch"
		}
		if extension["type"] != wantType {
			t.Errorf("%s: type = %v", row.name, extension["type"])
		}
	}
	for _, tc := range []struct{ name, id string }{
		{"issue/show", "issueShow"},
		{"issue/metadata/compare-and-set", "issueMetadataCompareAndSet"},
		{"issue/dep/add", "issueDepAdd"},
		{"events/list", "eventsList"},
	} {
		if got := callableOperationID(tc.name); got != tc.id {
			t.Errorf("operationId(%s) = %s, want %s", tc.name, got, tc.id)
		}
	}

	declaredArguments := func(name string) map[string]any {
		op := mapAt(t, mapAt(t, paths, callableMount+"/"+name), "post")
		schema := mapAt(t, mapAt(t, mapAt(t, op, "requestBody"), "content"), bdpMediaType)
		return mapAt(t, mapAt(t, mapAt(t, schema, "schema"), "properties"), "arguments")
	}
	required := func(name string) []any {
		list, _ := declaredArguments(name)["required"].([]any)
		return list
	}
	if got := required("issue/comment"); !reflect.DeepEqual(got, []any{"id", "text"}) {
		t.Errorf("issue/comment required = %v", got)
	}
	if got := required("issue/dep/add"); len(got) != 0 {
		t.Errorf("issue/dep/add required = %v", got)
	}
	if got := required("issue/claim"); !reflect.DeepEqual(got, []any{"id"}) {
		t.Errorf("issue/claim required = %v", got)
	}
	// A parameter v0 requires is not required where the service has a default for it.
	if got := required("issue/dep/list"); !reflect.DeepEqual(got, []any{"id"}) {
		t.Errorf("issue/dep/list required = %v", got)
	}
	if got := required("issue/purge"); len(got) != 0 {
		t.Errorf("issue/purge required = %v", got)
	}
	for i := range callableTable {
		row := &callableTable[i]
		properties := mapAt(t, declaredArguments(row.name), "properties")
		for _, name := range []string{"actor", "author"} {
			if _, has := properties[name]; has {
				t.Errorf("%s publishes `%s`", row.name, name)
			}
		}
		for name, raw := range properties {
			value, defaulted := row.defaults[name]
			schema, _ := raw.(map[string]any)
			if !defaulted {
				continue
			}
			var want any
			if err := json.Unmarshal(value, &want); err != nil {
				t.Fatal(err)
			}
			if schema["default"] != want {
				t.Errorf("%s: %s publishes default %v, want %v", row.name, name, schema["default"], want)
			}
			if got := schema["description"]; got != "Omitted, this Callable uses `"+string(value)+"`." {
				t.Errorf("%s: %s is described as %v", row.name, name, got)
			}
			if slices.Contains(required(row.name), any(name)) {
				t.Errorf("%s: %s has a default and is required", row.name, name)
			}
		}
		op := mapAt(t, mapAt(t, paths, callableMount+"/"+row.name), "post")
		if cli, ok := mapAt(t, op, "x-bdp-callable")["cliMapping"].(map[string]any); ok {
			if _, has := cli["defaults"]; has {
				t.Errorf("%s: cliMapping still has `defaults`: %v", row.name, cli)
			}
		}
	}

	// Every $ref resolves inside the document (§2.3), and 3.0's spellings are gone.
	var walk func(node any)
	walk = func(node any) {
		switch value := node.(type) {
		case []any:
			for _, item := range value {
				walk(item)
			}
		case map[string]any:
			if ref, ok := value["$ref"].(string); ok {
				rest, local := strings.CutPrefix(ref, "#/components/schemas/")
				if !local || mapAt(t, mapAt(t, doc, "components"), "schemas")[rest] == nil {
					t.Errorf("$ref %q does not resolve inside the document", ref)
				}
			}
			for key, member := range value {
				if strings.HasPrefix(key, "x-go-") || key == "x-bd-codes" {
					t.Errorf("code-generation annotation %q survived", key)
				}
				if _, isBool := member.(bool); isBool && key == "nullable" {
					t.Errorf("OpenAPI 3.0 `nullable` survived: %v", value)
				}
				walk(member)
			}
		}
	}
	walk(doc)
}

// TestStorageSentinelsStayUnderProblemMapping keeps one real sentinel route to a
// Core code, so the mapping is shown to compose with ClassifyError rather than
// only with synthetic bodies.
func TestStorageSentinelsStayUnderProblemMapping(t *testing.T) {
	lifecycle := &roleLifecycle{updateErr: fmt.Errorf("update: %w", storage.ErrNotFound)}
	ts := newBDPServer(t, rolesConfig(Config{Lifecycle: lifecycle}))
	wantProblem(t, ts.callAs(t, "issue/update", "a", `{"id":"bd-1","patch":{"title":"t"}}`),
		http.StatusNotFound, bdpResourceNotFound)
}

// TestCallableGuardTakesACoreRevision: the Issue document's routing rule. A token
// from a generic Core record works as `expected_version` on the Callable path, with
// both of its halves checked; a bare row lock goes on working, and v0's own route is
// not touched.
func TestCallableGuardTakesACoreRevision(t *testing.T) {
	const patch = `"patch":{"title":"t"}`
	setup := func(t *testing.T, rowVersion int64) (*testServer, *coreStore, *types.Issue, *roleLifecycle, *roleDeleter) {
		t.Helper()
		issue := coreIssue("sw-1", types.TypeTask)
		issue.RowVersion = rowVersion
		store := &coreStore{issues: []*types.Issue{issue}}
		lifecycle := &roleLifecycle{updateResult: issueops.UpdateResult{Issue: issue}, inTxRow: store.rowInTx}
		deleter := &roleDeleter{result: issueops.DeleteResult{Deleted: 1}, inTxRow: store.rowInTx}
		return newCoreServer(t, store, Config{Lifecycle: lifecycle, Deleter: deleter}), store, issue, lifecycle, deleter
	}
	revision := func(t *testing.T, ts *testServer) string {
		t.Helper()
		_, record := coreGetJSON(t, ts, "/sw-1")
		return record["revision"].(string)
	}

	t.Run("a token never starts with a dash, and works with a negative row lock", func(t *testing.T) {
		for _, rowVersion := range []int64{7, -4107163796693034294} {
			ts, _, _, lifecycle, _ := setup(t, rowVersion)
			token := revision(t, ts)
			if strings.HasPrefix(token, "-") {
				t.Fatalf("token %q reads as a flag", token)
			}
			if resp := ts.callAs(t, "issue/update", "a", `{"id":"sw-1",`+patch+`,"expected_version":"`+token+`"}`); resp.StatusCode != http.StatusOK {
				t.Fatalf("%q: status %d", token, resp.StatusCode)
			}
			got := lifecycle.updateRequests()
			if len(got) != 1 || got[0].ExpectedVersion == nil || *got[0].ExpectedVersion != rowVersion {
				t.Fatalf("%q reached the role as %+v", token, got)
			}
		}
	})

	t.Run("a change the row lock does not see is refused, and writes nothing", func(t *testing.T) {
		ts, store, issue, lifecycle, _ := setup(t, 7)
		token := revision(t, ts)
		store.mu.Lock()
		issue.Labels = []string{"new"} // A label-only write: the row lock stays at 7.
		store.mu.Unlock()
		for _, name := range []string{"issue/update", "issue/close", "issue/reopen"} {
			arguments := `{"id":"sw-1",`
			if name == "issue/update" {
				arguments += patch + `,`
			}
			wantProblem(t, ts.callAs(t, name, "a", arguments+`"expected_version":"`+token+`"}`), http.StatusConflict, bdpRevisionMismatch)
		}
		if got := len(lifecycle.updateRequests()) + len(lifecycle.closeRequests()) + len(lifecycle.reopenRequests()); got != 0 {
			t.Errorf("%d requests reached a role past a stale digest", got)
		}
	})

	t.Run("a moved row lock is refused", func(t *testing.T) {
		ts, store, issue, lifecycle, _ := setup(t, 7)
		token := revision(t, ts)
		store.mu.Lock()
		issue.RowVersion = 8
		store.mu.Unlock()
		wantProblem(t, ts.callAs(t, "issue/update", "a", `{"id":"sw-1",`+patch+`,"expected_version":"`+token+`"}`), http.StatusConflict, bdpRevisionMismatch)
		if got := lifecycle.updateRequests(); len(got) != 0 {
			t.Errorf("reached the role: %+v", got)
		}
	})

	// The race the outside comparison had: a label-only write that lands after the mount
	// has the request and before the role's transaction reads the row. The row lock half
	// of the token still matches, so only the whole token, compared in the transaction,
	// can catch it.
	t.Run("a change landing on the way to the write is refused", func(t *testing.T) {
		for _, name := range []string{"issue/update", "issue/close", "issue/reopen", "issue/delete"} {
			ts, store, issue, lifecycle, deleter := setup(t, 7)
			token := revision(t, ts)
			read := store.rowInTx
			landing := func(id string) (*types.Issue, error) {
				store.mu.Lock()
				issue.Labels = []string{"landed"}
				store.mu.Unlock()
				return read(id)
			}
			lifecycle.inTxRow, deleter.inTxRow = landing, landing
			arguments := `{"id":"sw-1",`
			switch name {
			case "issue/update":
				arguments += patch + `,`
			case "issue/delete":
				arguments = `{"ids":["sw-1"],`
			}
			wantProblem(t, ts.callAs(t, name, "a", arguments+`"expected_version":"`+token+`"}`), http.StatusConflict, bdpRevisionMismatch)
			if got := len(lifecycle.updateRequests()) + len(lifecycle.closeRequests()) + len(lifecycle.reopenRequests()) + len(deleter.requests()); got != 0 {
				t.Errorf("%s: %d writes got past a stale guard", name, got)
			}
		}
	})

	t.Run("a bare row lock keeps working, and v0 reads it", func(t *testing.T) {
		ts, _, _, lifecycle, _ := setup(t, 7)
		if resp := ts.callAs(t, "issue/update", "a", `{"id":"sw-1",`+patch+`,"expected_version":"7"}`); resp.StatusCode != http.StatusOK {
			t.Fatalf("status %d", resp.StatusCode)
		}
		got := lifecycle.updateRequests()
		if len(got) != 1 || got[0].ExpectedVersion == nil || *got[0].ExpectedVersion != 7 {
			t.Fatalf("reached the role as %+v", got)
		}
		lifecycle.updateErr = fmt.Errorf("update: %w", issueops.ErrVersionMismatch)
		wantProblem(t, ts.callAs(t, "issue/update", "a", `{"id":"sw-1",`+patch+`,"expected_version":"6"}`), http.StatusConflict, bdpRevisionMismatch)
	})

	t.Run("v0's refusal stands for what is no token", func(t *testing.T) {
		ts, _, _, lifecycle, _ := setup(t, 7)
		for _, value := range []string{`"nope"`, `7`, `null`, `"-4107163796693034294-85a3c1d2e3f405162738495a6b7c8d9e"`} {
			wantProblem(t, ts.callAs(t, "issue/update", "a", `{"id":"sw-1",`+patch+`,"expected_version":`+value+`}`), http.StatusUnprocessableEntity, bdpValidationFailed)
		}
		if got := lifecycle.updateRequests(); len(got) != 0 {
			t.Errorf("reached the role: %+v", got)
		}
	})

	t.Run("issue/delete takes one, with its defaults intact, and two ids are v0's to refuse", func(t *testing.T) {
		ts, store, issue, _, deleter := setup(t, 7)
		token := revision(t, ts)
		if resp := ts.call(t, "issue/delete", `{"ids":["sw-1"," sw-1"],"expected_version":"`+token+`"}`); resp.StatusCode != http.StatusOK {
			t.Fatalf("status %d", resp.StatusCode)
		}
		got := deleter.requests()
		if len(got) != 1 || got[0].ExpectedVersion == nil || *got[0].ExpectedVersion != 7 || !got[0].DryRun {
			t.Fatalf("reached the role as %+v", got)
		}
		wantProblem(t, ts.call(t, "issue/delete", `{"ids":["sw-1","sw-2"],"expected_version":"`+token+`"}`), http.StatusUnprocessableEntity, bdpValidationFailed)
		store.mu.Lock()
		issue.Labels = []string{"new"}
		store.mu.Unlock()
		wantProblem(t, ts.call(t, "issue/delete", `{"ids":["sw-1"],"expected_version":"`+token+`"}`), http.StatusConflict, bdpRevisionMismatch)
		if got := deleter.requests(); len(got) != 1 {
			t.Errorf("%d requests reached the role, want 1", len(got))
		}
	})

	t.Run("a missing record is v0's answer", func(t *testing.T) {
		ts, _, _, lifecycle, _ := setup(t, 7)
		token := revision(t, ts)
		lifecycle.updateErr = fmt.Errorf("update: %w", storage.ErrNotFound)
		wantProblem(t, ts.callAs(t, "issue/update", "a", `{"id":"sw-404",`+patch+`,"expected_version":"`+token+`"}`), http.StatusNotFound, bdpResourceNotFound)
	})

	t.Run("a guard that cannot be read writes nothing", func(t *testing.T) {
		ts, store, _, lifecycle, _ := setup(t, 7)
		token := revision(t, ts)
		store.mu.Lock()
		store.getErr = errors.New("db down")
		store.mu.Unlock()
		wantProblem(t, ts.callAs(t, "issue/update", "a", `{"id":"sw-1",`+patch+`,"expected_version":"`+token+`"}`), http.StatusServiceUnavailable, bdpUnavailable)
		if got := lifecycle.updateRequests(); len(got) != 0 {
			t.Errorf("reached the role: %+v", got)
		}
	})

	t.Run("v0's own route does not take a Core token", func(t *testing.T) {
		ts, _, _, lifecycle, _ := setup(t, 7)
		token := revision(t, ts)
		req, err := http.NewRequest(http.MethodPatch, ts.base+"/v0/beads/issues/sw-1", strings.NewReader(`{"actor":"a",`+patch+`,"expected_version":"`+token+`"}`))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := ts.client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("status %d, want 400", resp.StatusCode)
		}
		if got := lifecycle.updateRequests(); len(got) != 0 {
			t.Errorf("reached the role: %+v", got)
		}
	})
}
