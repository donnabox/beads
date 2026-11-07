package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/issueops"
)

// writeRoles answers the three write roles over a coreStore as the real legs do, in
// the respects Write depends on: InTx runs over the row the role is about to write,
// and its answer decides what happens. It is a double of their contract. The contract
// itself is held on real Dolt by internal/storage/uow's TestGuardComparedInsideTheWrite
// and by the live run in this work's report.
type writeRoles struct {
	store *coreStore
	mu    sync.Mutex
	next  int
	// retired ids refuse a create as a retaining sweep's do.
	retired map[string]bool
	// race lands just before a role reads the row, the moment an outside comparison
	// cannot see.
	race func()
	// actors records who each write named.
	actors []string
	// now stamps new rows.
	now func() time.Time
}

func newWriteServer(t *testing.T, cfg Config) (*testServer, *writeRoles) {
	t.Helper()
	store := &coreStore{}
	roles := &writeRoles{store: store, retired: map[string]bool{}, now: func() time.Time { return coreClock }}
	cfg.Lifecycle, cfg.Deleter = roles, roles
	return newCoreServer(t, store, cfg), roles
}

func (f *writeRoles) find(id string) *types.Issue {
	f.store.mu.Lock()
	defer f.store.mu.Unlock()
	for _, issue := range f.store.issues {
		if issue.ID == id {
			clone := *issue
			clone.Labels = slices.Clone(issue.Labels)
			return &clone
		}
	}
	return nil
}

func (f *writeRoles) put(next *types.Issue) {
	f.store.mu.Lock()
	defer f.store.mu.Unlock()
	for i, issue := range f.store.issues {
		if issue.ID == next.ID {
			f.store.issues[i] = next
			return
		}
	}
	f.store.issues = append(f.store.issues, next)
}

func (f *writeRoles) landRace() {
	f.mu.Lock()
	race := f.race
	f.race = nil
	f.mu.Unlock()
	if race != nil {
		race()
	}
}

func (f *writeRoles) Create(_ context.Context, req issueops.CreateRequest) (issueops.CreateResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.actors = append(f.actors, req.Actor)
	if !req.StoredTimestamps {
		return issueops.CreateResult{}, errors.New("a Write create asks for the stored timestamps")
	}
	issue := *req.Issue
	if issue.ID == "" {
		f.next++
		issue.ID = fmt.Sprintf("wr-%d", f.next)
	}
	if f.retired[issue.ID] {
		return issueops.CreateResult{}, &issueops.IdempotencyExpiredError{ID: issue.ID}
	}
	if f.find(issue.ID) != nil {
		return issueops.CreateResult{}, fmt.Errorf("create %s: %w", issue.ID, issueops.ErrAlreadyExists)
	}
	issue.CreatedBy = req.Actor
	issue.CreatedAt, issue.UpdatedAt = coreClock, coreClock
	issue.RowVersion = 7
	f.put(&issue)
	return issueops.CreateResult{Issue: f.find(issue.ID)}, nil
}

func (f *writeRoles) Update(_ context.Context, req issueops.UpdateRequest) (issueops.UpdateResult, error) {
	f.landRace()
	f.mu.Lock()
	defer f.mu.Unlock()
	f.actors = append(f.actors, req.Actor)
	current := f.find(req.IssueID)
	if current == nil {
		return issueops.UpdateResult{}, fmt.Errorf("update %s: %w", req.IssueID, storage.ErrNotFound)
	}
	if req.ExpectedVersion != nil && *req.ExpectedVersion != current.RowVersion {
		return issueops.UpdateResult{}, &issueops.VersionMismatchError{Expected: *req.ExpectedVersion, Current: current.RowVersion}
	}
	attempt := req
	if req.InTx != nil {
		if err := req.InTx(f.find(req.IssueID), &attempt); err != nil {
			return issueops.UpdateResult{}, err
		}
	}
	if attempt.Patch.Notes.Set && current.Notes != "" && !attempt.ForceNotesOverwrite {
		return issueops.UpdateResult{}, storage.ErrNotesOverwrite
	}
	changed := applyPatchToIssue(current, attempt.Patch)
	if changed {
		current.RowVersion++
		current.UpdatedAt = current.UpdatedAt.Add(time.Second)
		f.put(current)
	}
	return issueops.UpdateResult{Issue: f.find(req.IssueID), Changed: changed}, nil
}

func applyPatchToIssue(issue *types.Issue, patch issueops.IssuePatch) bool {
	before, _ := json.Marshal(issue)
	set := func(field issueops.Field[string], into *string) {
		if field.Set {
			*into = field.Value
		}
	}
	set(patch.Title, &issue.Title)
	set(patch.Description, &issue.Description)
	set(patch.Design, &issue.Design)
	set(patch.AcceptanceCriteria, &issue.AcceptanceCriteria)
	set(patch.Notes, &issue.Notes)
	set(patch.SpecID, &issue.SpecID)
	set(patch.Assignee, &issue.Assignee)
	set(patch.Owner, &issue.Owner)
	if patch.Status.Set {
		issue.Status = patch.Status.Value
	}
	if patch.Priority.Set {
		issue.Priority = patch.Priority.Value
	}
	if patch.IssueType.Set {
		issue.IssueType = patch.IssueType.Value
	}
	if patch.EstimatedMinutes.Set {
		issue.EstimatedMinutes = patch.EstimatedMinutes.Value
	}
	if patch.ExternalRef.Set {
		issue.ExternalRef = patch.ExternalRef.Value
	}
	if patch.DueAt.Set {
		issue.DueAt = patch.DueAt.Value
	}
	if patch.DeferUntil.Set {
		issue.DeferUntil = patch.DeferUntil.Value
	}
	if patch.Labels.Replace.Set {
		issue.Labels = patch.Labels.Replace.Value
	}
	if patch.Metadata.Replace.Set {
		issue.Metadata = patch.Metadata.Replace.Value
	}
	after, _ := json.Marshal(issue)
	return !bytes.Equal(before, after)
}

func (f *writeRoles) Close(context.Context, issueops.CloseRequest) (issueops.CloseResult, error) {
	return issueops.CloseResult{}, errors.New("not reached by Write")
}

func (f *writeRoles) Reopen(context.Context, issueops.ReopenRequest) (issueops.ReopenResult, error) {
	return issueops.ReopenResult{}, errors.New("not reached by Write")
}

func (f *writeRoles) Delete(_ context.Context, req issueops.DeleteRequest) (issueops.DeleteResult, error) {
	f.landRace()
	f.mu.Lock()
	defer f.mu.Unlock()
	f.actors = append(f.actors, req.Actor)
	id := req.IDs[0]
	current := f.find(id)
	if current == nil {
		return issueops.DeleteResult{}, &issueops.NotFoundError{IDs: req.IDs}
	}
	if req.InTx != nil {
		if err := req.InTx(current); err != nil {
			return issueops.DeleteResult{}, err
		}
	}
	f.store.mu.Lock()
	defer f.store.mu.Unlock()
	if req.RefuseIncidentEdges {
		for _, edge := range f.store.edges {
			if edge.IssueID == id || edge.DependsOnID == id {
				return issueops.DeleteResult{}, &issueops.IncidentEdgesError{IssueID: id}
			}
		}
	}
	f.store.issues = slices.DeleteFunc(f.store.issues, func(issue *types.Issue) bool { return issue.ID == id })
	return issueops.DeleteResult{Deleted: 1, DeletedIDs: []string{id}}, nil
}

func (f *writeRoles) lastActor() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.actors) == 0 {
		return ""
	}
	return f.actors[len(f.actors)-1]
}

type writeReply struct {
	status int
	header http.Header
	body   map[string]any
	raw    string
}

// post sends one mutation with the media Write selects.
func writePost(t *testing.T, ts *testServer, member, body string) writeReply {
	t.Helper()
	return writePostAs(t, ts, member, body, http.Header{"Content-Type": {bdpMediaType}, "Accept": {bdpMediaType}})
}

func writePostAs(t *testing.T, ts *testServer, member, body string, header http.Header) writeReply {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, ts.base+coreOperationsPath+"/"+member, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header = header
	resp, err := ts.client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	reply := writeReply{status: resp.StatusCode, header: resp.Header, raw: string(raw)}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &reply.body); err != nil {
			t.Fatalf("POST %s answered undecodable JSON %q: %v", member, raw, err)
		}
	}
	return reply
}

// refused checks one refusal against the Write table, and that the store did not move.
func (reply writeReply) refused(t *testing.T, code string) {
	t.Helper()
	shape, ok := bdpProblemShapes[code]
	if !ok {
		t.Fatalf("no shape for %q", code)
	}
	if reply.status != shape.status || reply.body["code"] != code {
		t.Fatalf("got %d %s, want %d %s", reply.status, reply.raw, shape.status, code)
	}
	if reply.body["retry"] != shape.retry || reply.body["type"] != bdpProblemBase+shape.family {
		t.Errorf("problem shape: %s", reply.raw)
	}
	if got := reply.header.Get("Content-Type"); got != problemMediaType {
		t.Errorf("Content-Type = %q", got)
	}
	if got := reply.header.Get("Cache-Control"); got != "private, no-store" {
		t.Errorf("Cache-Control = %q", got)
	}
	if code == bdpValidationFailed {
		diagnostics, _ := reply.body["diagnostics"].([]any)
		if len(diagnostics) == 0 {
			t.Errorf("validation-failed without diagnostics: %s", reply.raw)
		}
	}
}

// succeeded checks the common shape of a result and answers its Resource.
func (reply writeReply) succeeded(t *testing.T, outcome string, changed bool) map[string]any {
	t.Helper()
	if reply.status != http.StatusOK {
		t.Fatalf("got %d: %s", reply.status, reply.raw)
	}
	if reply.body["outcome"] != outcome || reply.body["changed"] != changed {
		t.Fatalf("result = %s, want %s changed %t", reply.raw, outcome, changed)
	}
	if got := reply.header.Get("Cache-Control"); got != "private, no-store" {
		t.Errorf("Cache-Control = %q", got)
	}
	if reply.header.Get("ETag") != "" || reply.header.Get("Location") != "" {
		t.Errorf("a command URL carries no ETag or Location: %v", reply.header)
	}
	if _, has := reply.body["ownerRevisions"]; has {
		t.Errorf("preview-1 declares no ownership: %s", reply.raw)
	}
	if outcome == writeOutcomeDeleted {
		return nil
	}
	resource, _ := reply.body["resource"].(map[string]any)
	if resource == nil {
		t.Fatalf("no resource: %s", reply.raw)
	}
	return resource
}

// wantDiagnostic checks that some diagnostic's message contains text.
func (reply writeReply) wantDiagnostic(t *testing.T, text string) {
	t.Helper()
	diagnostics, _ := reply.body["diagnostics"].([]any)
	for _, entry := range diagnostics {
		if message, _ := entry.(map[string]any)["message"].(string); strings.Contains(message, text) {
			return
		}
	}
	t.Errorf("no diagnostic contains %q: %s", text, reply.raw)
}

const (
	writeTask  = coreBeadTypeNamespace + "task"
	writeBlock = coreLinkTypeNamespace + "blocks"
)

func writeBody(id, extra string) string {
	body := `{"type":"` + writeTask + `","properties":{"title":"t","status":"open","priority":2`
	if extra != "" {
		body += "," + extra
	}
	body += `}`
	if id != "" {
		body += `,"id":"` + id + `"`
	}
	return body + `}`
}

// seedBead creates a Bead through the route and answers its record.
func seedBead(t *testing.T, ts *testServer, id string) map[string]any {
	t.Helper()
	return writePost(t, ts, "createBead", writeBody(id, "")).succeeded(t, writeOutcomeCreated, true)
}

func recordRevision(t *testing.T, record map[string]any) string {
	t.Helper()
	revision, _ := record["revision"].(string)
	if revision == "" {
		t.Fatalf("no revision on %v", record)
	}
	return revision
}

// TestWriteIsNotOnV0: with --protocol v0 nothing here exists, and the answer is the one
// any other unrouted path gets.
func TestWriteIsNotOnV0(t *testing.T) {
	ts := newTestServer(t, rolesConfig(Config{}))
	baseline, baselineBody := coreRequest(t, ts, http.MethodPost, "/no-such-route/anywhere", nil)
	for _, path := range []string{coreOperationsPath, coreOperationsPath + "/createBead", coreOperationsPath + "/updateLink", "/bdp"} {
		for _, method := range []string{http.MethodGet, http.MethodPost} {
			resp, body := coreRequest(t, ts, method, path, nil)
			if resp.StatusCode != baseline.StatusCode {
				t.Errorf("%s %s = %d, want the unrouted %d", method, path, resp.StatusCode, baseline.StatusCode)
			}
			var problem, want map[string]any
			_ = json.Unmarshal(body, &problem)
			_ = json.Unmarshal(baselineBody, &want)
			if problem["code"] != want["code"] {
				t.Errorf("%s %s = %s, want %s", method, path, body, baselineBody)
			}
		}
	}
}

// TestWriteDiscovery pins the advertisement: Write, its directory of six, and what is
// not offered.
func TestWriteDiscovery(t *testing.T) {
	ts := newBDPServer(t, Config{Auth: authTokenFile(t, "real-token")})
	resp, body := coreRequest(t, ts, http.MethodGet, "/bdp", http.Header{"Accept": {bdpMediaType}})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /bdp = %d: %s", resp.StatusCode, body)
	}
	var doc map[string]any
	if err := json.Unmarshal(body, &doc); err != nil {
		t.Fatal(err)
	}
	if doc["profile"] != "write" || doc["operations"] != ts.base+coreOperationsPath {
		t.Errorf("discovery = %s", body)
	}
	for _, absent := range []string{"aliases", "events", "historicalResolution", "beads", "links"} {
		if _, has := doc[absent]; has {
			t.Errorf("discovery advertises %s: %s", absent, body)
		}
	}
	capabilities, _ := doc["capabilities"].(map[string]any)
	for name := range capabilities {
		if strings.Contains(name, "idempotency") || strings.Contains(name, "atomic-batch") || strings.Contains(name, "id-root") {
			t.Errorf("discovery advertises %s, which is not implemented: %s", name, body)
		}
	}
	exceptions, _ := capabilities[writeExceptionsCapability].(map[string]any)
	detail, _ := exceptions[writeExceptionsCapabilityVersion].(map[string]any)
	notSupported, _ := detail["operationsNotSupported"].([]any)
	if len(notSupported) != 3 || notSupported[0] != "createLink" || notSupported[1] != "updateLink" || notSupported[2] != "deleteLink" {
		t.Errorf("discovery does not say the three Link writes are refused: %s", body)
	}

	resp, body = coreRequest(t, ts, http.MethodGet, coreOperationsPath, http.Header{"Accept": {bdpMediaType}})
	if resp.StatusCode != http.StatusOK || resp.Header.Get("Content-Type") != bdpMediaType {
		t.Fatalf("GET directory = %d %q: %s", resp.StatusCode, resp.Header.Get("Content-Type"), body)
	}
	var directory map[string]string
	if err := json.Unmarshal(body, &directory); err != nil {
		t.Fatal(err)
	}
	want := []string{"createBead", "updateBead", "deleteBead", "createLink", "updateLink", "deleteLink"}
	seen := map[string]bool{}
	for _, member := range want {
		if !strings.HasPrefix(directory[member], ts.base+"/") {
			t.Errorf("%s = %q", member, directory[member])
		}
		seen[directory[member]] = true
	}
	if len(directory) != len(want) || len(seen) != len(want) {
		t.Errorf("directory is not six distinct members: %s", body)
	}

	resp, _ = coreRequest(t, ts, http.MethodPost, coreOperationsPath, nil)
	if resp.StatusCode != http.StatusMethodNotAllowed || resp.Header.Get("Allow") != "GET, HEAD" {
		t.Errorf("POST directory = %d, Allow %q", resp.StatusCode, resp.Header.Get("Allow"))
	}
	for _, member := range want {
		resp, body = coreRequest(t, ts, http.MethodGet, coreOperationsPath+"/"+member, nil)
		if resp.StatusCode != http.StatusMethodNotAllowed || resp.Header.Get("Allow") != "POST" || len(body) != 0 {
			t.Errorf("GET %s = %d, Allow %q, %d bytes; want a bodyless 405", member, resp.StatusCode, resp.Header.Get("Allow"), len(body))
		}
	}
}

// TestWriteAdmission is the carrier, in order, before anything is read.
func TestWriteAdmission(t *testing.T) {
	ts, roles := newWriteServer(t, Config{})
	t.Run("generic JSON is 415 and creates nothing", func(t *testing.T) {
		writePostAs(t, ts, "createBead", writeBody("wr-a", ""), http.Header{"Content-Type": {"application/json"}}).refused(t, bdpUnsupportedMedia)
		writePostAs(t, ts, "createBead", writeBody("wr-a", ""), http.Header{}).refused(t, bdpUnsupportedMedia)
		if roles.find("wr-a") != nil {
			t.Fatal("created")
		}
	})
	t.Run("an unacceptable Accept is a bodyless 406", func(t *testing.T) {
		reply := writePostAs(t, ts, "createBead", writeBody("wr-a", ""), http.Header{"Content-Type": {bdpMediaType}, "Accept": {"text/plain"}})
		if reply.status != http.StatusNotAcceptable || reply.raw != "" {
			t.Fatalf("got %d %q", reply.status, reply.raw)
		}
	})
	t.Run("the carrier is closed", func(t *testing.T) {
		for _, body := range []string{
			`[]`, `{`, `{"type":"` + writeTask + `","type":"` + writeTask + `"}`,
			`{"type":"` + writeTask + `","operation":"createBead"}`,
			`{"type":"` + writeTask + `","name":"x"}`,
			`{}`,
			`{"type":"task"}`,
			`{"type":3}`,
			`{"type":"` + writeTask + `","properties":[]}`,
			`{"type":"` + writeTask + `","metadata":null}`,
		} {
			writePost(t, ts, "createBead", body).refused(t, bdpMalformedRequest)
		}
	})
	t.Run("a query parameter is refused", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, ts.base+coreOperationsPath+"/createBead?x=1", strings.NewReader(writeBody("wr-a", "")))
		req.Header.Set("Content-Type", bdpMediaType)
		resp, err := ts.client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("got %d", resp.StatusCode)
		}
	})
	t.Run("Idempotency-Key is ignored, whatever it holds", func(t *testing.T) {
		reply := writePostAs(t, ts, "createBead", writeBody("", ""), http.Header{
			"Content-Type": {bdpMediaType}, "Idempotency-Key": {"!!! not a key", "two"},
		})
		reply.succeeded(t, writeOutcomeCreated, true)
	})
}

// TestWriteAsksACredential: a mutation target asks for what any Core read does.
func TestWriteAsksACredential(t *testing.T) {
	ts, roles := newWriteServer(t, Config{Auth: authTokenFile(t, "real-token")})
	reply := writePost(t, ts, "createBead", writeBody("wr-a", ""))
	if reply.status != http.StatusUnauthorized || reply.body["code"] != "unauthenticated" {
		t.Fatalf("got %d %s", reply.status, reply.raw)
	}
	for _, member := range writeLinkExceptions {
		if reply := writePost(t, ts, member, `{}`); reply.status != http.StatusUnauthorized {
			t.Fatalf("%s answered an anonymous caller: %d %s", member, reply.status, reply.raw)
		}
	}
	if roles.find("wr-a") != nil {
		t.Fatal("created")
	}
}

// TestCreateBead is IT-3, IT-4 and the identity rules, with Donna's ruling on a deleted id.
func TestCreateBead(t *testing.T) {
	ts, roles := newWriteServer(t, Config{})

	t.Run("the result is the record a read serves", func(t *testing.T) {
		resource := writePost(t, ts, "createBead", `{"id":"wr-read","type":"`+writeTask+`",
			"properties":{"title":"T","status":"open","priority":0,"labels":["b","a"],"estimated_minutes":0,"due_at":"2026-12-01T00:00:00Z"},
			"metadata":{"k":{"n":1.50}}}`).succeeded(t, writeOutcomeCreated, true)
		_, read := coreGetJSON(t, ts, "/wr-read")
		got, _ := json.Marshal(resource)
		want, _ := json.Marshal(read)
		if !bytes.Equal(got, want) {
			t.Errorf("result %s\nread   %s", got, want)
		}
		if resource["id"] != ts.base+"/wr-read" || resource["type"] != writeTask {
			t.Errorf("identity: %v", resource)
		}
	})
	t.Run("an omitted id is allocated", func(t *testing.T) {
		resource := writePost(t, ts, "createBead", writeBody("", "")).succeeded(t, writeOutcomeCreated, true)
		if id, _ := resource["id"].(string); !strings.HasPrefix(id, ts.base+"/wr-") {
			t.Errorf("id = %q", id)
		}
	})
	t.Run("IT-3 abstract and foreign Types refuse", func(t *testing.T) {
		for uri, why := range map[string]string{
			coreBeadFamily:                    "abstract",
			coreLinkFamily:                    "Link Type",
			writeBlock:                        "Link Type",
			coreBeadTypeNamespace + "nosuch":  "no Type",
			"https://example.invalid/t/x":     "only",
			ts.base + "/types/issue/not-here": "configures",
		} {
			reply := writePost(t, ts, "createBead", `{"type":"`+uri+`","properties":{"title":"t","status":"open","priority":2}}`)
			reply.refused(t, bdpValidationFailed)
			reply.wantDiagnostic(t, why)
		}
	})
	t.Run("a pinned affiliation cannot be stored, so it refuses", func(t *testing.T) {
		writePost(t, ts, "createBead", `{"type":{"uri":"`+writeTask+`","revision":"task-preview-1"},"properties":{"title":"t","status":"open","priority":2}}`).refused(t, bdpValidationFailed)
	})
	t.Run("a configured Type is served under the Scope", func(t *testing.T) {
		roles.store.mu.Lock()
		roles.store.settings = map[string]string{"types.custom": "custom-kind"}
		roles.store.mu.Unlock()
		// The table is read per request, and a request that needs it asks.
		uri := ts.base + "/types/issue/custom-kind"
		resource := writePost(t, ts, "createBead", `{"type":"`+uri+`","properties":{"title":"t","status":"open","priority":2}}`).succeeded(t, writeOutcomeCreated, true)
		if resource["type"] != uri {
			t.Errorf("type = %v", resource["type"])
		}
	})
	t.Run("IT-4 every missing required member is named", func(t *testing.T) {
		reply := writePost(t, ts, "createBead", `{"type":"`+writeTask+`","properties":{"description":"d"}}`)
		reply.refused(t, bdpValidationFailed)
		for _, name := range issueRequired {
			reply.wantDiagnostic(t, "`"+name+"` is required")
		}
		writePost(t, ts, "createBead", `{"type":"`+writeTask+`"}`).refused(t, bdpValidationFailed)
	})
	t.Run("service-managed, classification and unknown members refuse", func(t *testing.T) {
		for _, extra := range []string{
			`"created_at":"2026-01-01T00:00:00Z"`, `"closed_at":null`, `"close_reason":"x"`, `"arrival_seq":1`,
			`"issue_type":"task"`, `"ephemeral":true`, `"not_a_field":1`,
		} {
			writePost(t, ts, "createBead", writeBody("wr-x", extra)).refused(t, bdpValidationFailed)
		}
		if roles.find("wr-x") != nil {
			t.Fatal("created")
		}
	})
	t.Run("values are checked, and none is rounded or dropped", func(t *testing.T) {
		for _, extra := range []string{
			`"labels":["a","a"]`, `"labels":[""]`, `"labels":"a"`,
			`"estimated_minutes":-1`, `"estimated_minutes":1.5`,
			`"due_at":"tomorrow"`, `"due_at":"2026-12-01T00:00:00.5Z"`,
			`"description":null`,
		} {
			writePost(t, ts, "createBead", writeBody("wr-x", extra)).refused(t, bdpValidationFailed)
		}
		for _, priority := range []string{`5`, `-1`, `1.5`, `"2"`, `null`} {
			writePost(t, ts, "createBead", `{"type":"`+writeTask+`","properties":{"title":"t","status":"open","priority":`+priority+`}}`).refused(t, bdpValidationFailed)
		}
		writePost(t, ts, "createBead", `{"type":"`+writeTask+`","properties":{"title":" ","status":"open","priority":2}}`).refused(t, bdpValidationFailed)
		writePost(t, ts, "createBead", `{"type":"`+writeTask+`","properties":{"title":"t","status":"","priority":2}}`).refused(t, bdpValidationFailed)
	})
	t.Run("status closed needs closed_at, which is service-managed", func(t *testing.T) {
		reply := writePost(t, ts, "createBead", `{"type":"`+writeTask+`","properties":{"title":"t","status":"closed","priority":2}}`)
		reply.refused(t, bdpValidationFailed)
		reply.wantDiagnostic(t, "issue/close")
	})
	t.Run("an inadmissible number refuses, naming its document", func(t *testing.T) {
		reply := writePost(t, ts, "createBead", `{"type":"`+writeTask+`","properties":{"title":"t","status":"open","priority":2},"metadata":{"a":{"b":9007199254740993}}}`)
		reply.refused(t, bdpValidationFailed)
		reply.wantDiagnostic(t, "metadata")
		writePost(t, ts, "createBead", `{"type":"`+writeTask+`","properties":{"title":"t","status":"open","priority":2},"metadata":{"a":1e300}}`).succeeded(t, writeOutcomeCreated, true)
	})
	t.Run("id spellings", func(t *testing.T) {
		seedBead(t, ts, "wr-live")
		for _, id := range []string{"wr-live", ts.base + "/wr-live", coreLinkID("wr-a", "wr-b", coreClock)} {
			writePost(t, ts, "createBead", writeBody(id, "")).refused(t, bdpIdentityTaken)
		}
		for _, id := range []string{"a/b", "wr%40x", "bdp", "v0", coreLinkIDPrefix + "not-base64-pair", strings.Repeat("x", types.MaxFieldLen+1)} {
			writePost(t, ts, "createBead", writeBody(id, "")).refused(t, bdpValidationFailed)
		}
		for _, id := range []string{"a//b", "/a", "a/", "a/../b", ".", "a%2Fb", "a%2fb", "@name", "a?b", "a#b", "https://elsewhere.invalid/x"} {
			writePost(t, ts, "createBead", writeBody(id, "")).refused(t, bdpMalformedRequest)
		}
	})
	t.Run("a retired id stays identity-taken", func(t *testing.T) {
		roles.mu.Lock()
		roles.retired["wr-retired"] = true
		roles.mu.Unlock()
		writePost(t, ts, "createBead", writeBody("wr-retired", "")).refused(t, bdpIdentityTaken)
	})
	t.Run("Donna's ruling: a supplied id on a deleted Bead adds a revision under it", func(t *testing.T) {
		first := seedBead(t, ts, "wr-again")
		writePost(t, ts, "deleteBead", `{"bead":"wr-again"}`).succeeded(t, writeOutcomeDeleted, true)
		second := writePost(t, ts, "createBead", `{"id":"wr-again","type":"`+writeTask+`","properties":{"title":"second","status":"open","priority":3}}`).succeeded(t, writeOutcomeCreated, true)
		if second["id"] != first["id"] || second["revision"] == first["revision"] {
			t.Errorf("first %v, second %v", first, second)
		}
	})
}

// TestUpdateBead is IT-5, IT-11, the patch forms, no-ops and the guard.
func TestUpdateBead(t *testing.T) {
	ts, roles := newWriteServer(t, Config{})
	update := func(body string) writeReply { return writePost(t, ts, "updateBead", body) }

	t.Run("a replacement replaces what is writable and keeps the service's", func(t *testing.T) {
		before := writePost(t, ts, "createBead", `{"id":"wr-1","type":"`+writeTask+`","properties":{"title":"t","status":"open","priority":2,"description":"d","labels":["x"]},"metadata":{"m":1}}`).succeeded(t, writeOutcomeCreated, true)
		after := update(`{"bead":"wr-1","properties":{"title":"new","status":"open","priority":1},"revisionProvenance":{"actor":{"id":"sam","basis":"writer-supplied"}}}`).succeeded(t, writeOutcomeUpdated, true)
		properties := after["properties"].(map[string]any)
		if properties["title"] != "new" || properties["description"] != nil || properties["labels"] != nil {
			t.Errorf("properties = %v", properties)
		}
		if properties["created_at"] != before["properties"].(map[string]any)["created_at"] {
			t.Errorf("created_at moved: %v", properties)
		}
		if meta := after["metadata"].(map[string]any); meta["m"] != float64(1) {
			t.Errorf("metadata was not preserved: %v", meta)
		}
		if roles.lastActor() != "sam" {
			t.Errorf("actor = %q", roles.lastActor())
		}
		if after["revision"] == before["revision"] {
			t.Error("an effectful update kept its revision")
		}
		_, read := coreGetJSON(t, ts, "/wr-1")
		if read["revision"] != after["revision"] {
			t.Errorf("result revision %v, read %v", after["revision"], read["revision"])
		}
	})
	t.Run("a replacement of required members refuses", func(t *testing.T) {
		reply := update(`{"bead":"wr-1","properties":{"title":"only"}}`)
		reply.refused(t, bdpValidationFailed)
		reply.wantDiagnostic(t, "`status` is required")
	})
	t.Run("IT-5 a service-managed member refuses, equal or undone", func(t *testing.T) {
		_, read := coreGetJSON(t, ts, "/wr-1")
		createdAt := read["properties"].(map[string]any)["created_at"].(string)
		for _, body := range []string{
			`{"bead":"wr-1","properties":{"title":"t","status":"open","priority":2,"created_at":"` + createdAt + `"}}`,
			`{"bead":"wr-1","propertiesChange":[{"op":"replace","path":"/updated_at","value":"2026-01-01T00:00:00Z"}]}`,
			`{"bead":"wr-1","propertiesChange":[{"op":"remove","path":"/created_at"}]}`,
			`{"bead":"wr-1","propertiesChange":[{"op":"add","path":"/closed_at","value":null},{"op":"remove","path":"/closed_at"}]}`,
			`{"bead":"wr-1","propertiesChange":[{"op":"replace","path":"","value":{}}]}`,
			`{"bead":"wr-1","propertiesChange":[{"op":"add","path":"/issue_type","value":"bug"}]}`,
		} {
			update(body).refused(t, bdpValidationFailed)
		}
	})
	t.Run("a patch applies in order to the state the write lands on", func(t *testing.T) {
		seedBead(t, ts, "wr-2")
		after := update(`{"bead":"wr-2","propertiesChange":[
			{"op":"add","path":"/labels","value":["a"]},
			{"op":"add","path":"/labels/-","value":"b"},
			{"op":"add","path":"/labels/0","value":"z"},
			{"op":"remove","path":"/labels/1"},
			{"op":"replace","path":"/title","value":"patched"}]}`).succeeded(t, writeOutcomeUpdated, true)
		properties := after["properties"].(map[string]any)
		if properties["title"] != "patched" {
			t.Errorf("title = %v", properties["title"])
		}
		if labels, _ := properties["labels"].([]any); len(labels) != 2 || labels[0] != "b" || labels[1] != "z" {
			t.Errorf("labels = %v", properties["labels"])
		}
	})
	t.Run("a patch's targets and final result", func(t *testing.T) {
		for _, body := range []string{
			`{"bead":"wr-2","propertiesChange":[{"op":"replace","path":"/description","value":"x"}]}`,
			`{"bead":"wr-2","propertiesChange":[{"op":"remove","path":"/nope"}]}`,
			`{"bead":"wr-2","propertiesChange":[{"op":"replace","path":"/labels/9","value":"x"}]}`,
			`{"bead":"wr-2","propertiesChange":[{"op":"add","path":"/not_writable","value":1}]}`,
			`{"bead":"wr-2","propertiesChange":[{"op":"remove","path":"/title"}]}`,
			`{"bead":"wr-2","metadataChange":[{"op":"replace","path":"","value":3}]}`,
			`{"bead":"wr-2","metadataChange":[{"op":"remove","path":""}]}`,
			`{"bead":"wr-2","metadataChange":[{"op":"remove","path":"/nope"}]}`,
		} {
			update(body).refused(t, bdpValidationFailed)
		}
	})
	t.Run("a patch's syntax is the carrier", func(t *testing.T) {
		for _, body := range []string{
			`{"bead":"wr-2","propertiesChange":{}}`,
			`{"bead":"wr-2","propertiesChange":[{"op":"move","path":"/a","from":"/b"}]}`,
			`{"bead":"wr-2","propertiesChange":[{"op":"test","path":"/a","value":1}]}`,
			`{"bead":"wr-2","propertiesChange":[{"op":"add","path":"/a"}]}`,
			`{"bead":"wr-2","propertiesChange":[{"op":"remove","path":"/a","value":1}]}`,
			`{"bead":"wr-2","propertiesChange":[{"op":"remove","path":"a"}]}`,
			`{"bead":"wr-2","propertiesChange":[{"op":"remove","path":"/a~2"}]}`,
			`{"bead":"wr-2","propertiesChange":[{"op":"add","path":"/a","value":1,"extra":true}]}`,
			`{"bead":"wr-2","properties":{},"propertiesChange":[]}`,
			`{"bead":"wr-2","metadata":{},"metadataChange":[]}`,
			`{"bead":"wr-2"}`,
			`{"bead":"wr-2","expectedRevision":"x"}`,
			`{"bead":"@local","properties":{}}`,
			`{"bead":{"uri":"wr-2"},"properties":{}}`,
			`{"bead":"wr-2","expectedRevision":"","metadata":{}}`,
		} {
			update(body).refused(t, bdpMalformedRequest)
		}
	})
	t.Run("IT-11 a label-only change serves a new revision", func(t *testing.T) {
		before := seedBead(t, ts, "wr-3")
		after := update(`{"bead":"wr-3","propertiesChange":[{"op":"add","path":"/labels","value":["l"]}],"expectedRevision":"`+recordRevision(t, before)+`"}`).succeeded(t, writeOutcomeUpdated, true)
		if after["revision"] == before["revision"] {
			t.Error("a label change kept the revision")
		}
	})
	t.Run("a no-op mints no revision", func(t *testing.T) {
		before := seedBead(t, ts, "wr-4")
		token := recordRevision(t, before)
		for _, body := range []string{
			`{"bead":"wr-4","properties":{"title":"t","status":"open","priority":2},"expectedRevision":"` + token + `"}`,
			`{"bead":"wr-4","properties":{"title":"t","status":"open","priority":2.0,"description":"","labels":[],"estimated_minutes":null}}`,
			`{"bead":"wr-4","propertiesChange":[]}`,
			`{"bead":"wr-4","metadataChange":[]}`,
			`{"bead":"wr-4","metadata":{}}`,
			`{"bead":"wr-4","type":"` + writeTask + `"}`,
			`{"bead":"wr-4","propertiesChange":[{"op":"add","path":"/title","value":"x"},{"op":"replace","path":"/title","value":"t"}]}`,
			`{"bead":"` + ts.base + `/wr-4","propertiesChange":[]}`,
		} {
			after := update(body).succeeded(t, writeOutcomeUpdated, false)
			if after["revision"] != before["revision"] {
				t.Errorf("%s moved the revision", body)
			}
		}
	})
	t.Run("a nominal Type change is an effect", func(t *testing.T) {
		before := seedBead(t, ts, "wr-5")
		after := update(`{"bead":"wr-5","type":"`+coreBeadTypeNamespace+`bug"}`).succeeded(t, writeOutcomeUpdated, true)
		if after["type"] != coreBeadTypeNamespace+"bug" || after["revision"] == before["revision"] {
			t.Errorf("after = %v", after)
		}
		update(`{"bead":"wr-5","type":"`+coreBeadFamily+`"}`).refused(t, bdpValidationFailed)
	})
	t.Run("metadata keeps its literals and compares by value", func(t *testing.T) {
		seedBead(t, ts, "wr-6")
		after := update(`{"bead":"wr-6","metadata":{"n":1.50,"s":"<x>"}}`).succeeded(t, writeOutcomeUpdated, true)
		if !strings.Contains(string(roles.find("wr-6").Metadata), "1.50") {
			t.Errorf("stored %s", roles.find("wr-6").Metadata)
		}
		again := update(`{"bead":"wr-6","metadata":{"s":"<x>","n":1.5}}`).succeeded(t, writeOutcomeUpdated, false)
		if again["revision"] != after["revision"] {
			t.Error("an equal metadata moved the revision")
		}
		patched := update(`{"bead":"wr-6","metadataChange":[{"op":"add","path":"/k","value":{"deep":[1]}},{"op":"replace","path":"/k/deep/0","value":2}]}`).succeeded(t, writeOutcomeUpdated, true)
		if meta := patched["metadata"].(map[string]any); meta["k"].(map[string]any)["deep"].([]any)[0] != float64(2) {
			t.Errorf("metadata = %v", meta)
		}
		update(`{"bead":"wr-6","metadata":{"a":1e400}}`).refused(t, bdpValidationFailed)
		update(`{"bead":"wr-6","metadataChange":[{"op":"add","path":"/a","value":9007199254740993}]}`).refused(t, bdpValidationFailed)
	})
	t.Run("clears, and what a clear is", func(t *testing.T) {
		writePost(t, ts, "createBead", `{"id":"wr-7","type":"`+writeTask+`","properties":{"title":"t","status":"open","priority":2,"estimated_minutes":0,"external_ref":"gh-1","due_at":"2026-12-01T00:00:00Z"}}`).succeeded(t, writeOutcomeCreated, true)
		after := update(`{"bead":"wr-7","propertiesChange":[{"op":"replace","path":"/estimated_minutes","value":null},{"op":"remove","path":"/external_ref"},{"op":"replace","path":"/due_at","value":null}]}`).succeeded(t, writeOutcomeUpdated, true)
		properties := after["properties"].(map[string]any)
		for _, name := range []string{"estimated_minutes", "external_ref", "due_at"} {
			if _, has := properties[name]; has {
				t.Errorf("%s survived: %v", name, properties)
			}
		}
		update(`{"bead":"wr-7","propertiesChange":[{"op":"replace","path":"/title","value":null}]}`).refused(t, bdpValidationFailed)
	})
	t.Run("a direct edit to closed refuses", func(t *testing.T) {
		seedBead(t, ts, "wr-8")
		reply := update(`{"bead":"wr-8","propertiesChange":[{"op":"replace","path":"/status","value":"closed"}]}`)
		reply.refused(t, bdpValidationFailed)
		reply.wantDiagnostic(t, "issue/close")
		update(`{"bead":"wr-8","propertiesChange":[{"op":"replace","path":"/status","value":"in_progress"}]}`).succeeded(t, writeOutcomeUpdated, true)
	})
	t.Run("the guard", func(t *testing.T) {
		before := seedBead(t, ts, "wr-9")
		token := recordRevision(t, before)
		update(`{"bead":"wr-9","properties":{"title":"x","status":"open","priority":2},"expectedRevision":"nope"}`).refused(t, bdpRevisionMismatch)
		update(`{"bead":"wr-9","propertiesChange":[{"op":"replace","path":"/title","value":"x"}],"expectedRevision":"`+strings.Replace(token, token[:1], "f", 1)+`f"}`).refused(t, bdpRevisionMismatch)
		if roles.find("wr-9").Title != "t" {
			t.Fatal("a stale guard wrote")
		}
		update(`{"bead":"wr-9","propertiesChange":[{"op":"replace","path":"/title","value":"x"}],"expectedRevision":"`+token+`"}`).succeeded(t, writeOutcomeUpdated, true)
		// A repeat with the prior guard fails, as Write's repetition rules say.
		update(`{"bead":"wr-9","propertiesChange":[{"op":"replace","path":"/title","value":"y"}],"expectedRevision":"`+token+`"}`).refused(t, bdpRevisionMismatch)
	})
	t.Run("the guard is compared inside the write", func(t *testing.T) {
		before := seedBead(t, ts, "wr-race")
		token := recordRevision(t, before)
		// A label-only write that lands as the role opens its transaction: the row
		// lock half of the token stays right, so only the whole token can catch it.
		roles.mu.Lock()
		roles.race = func() {
			issue := roles.find("wr-race")
			issue.Labels = []string{"landed"}
			roles.put(issue)
		}
		roles.mu.Unlock()
		update(`{"bead":"wr-race","propertiesChange":[{"op":"replace","path":"/title","value":"stale"}],"expectedRevision":"`+token+`"}`).refused(t, bdpRevisionMismatch)
		if got := roles.find("wr-race"); got.Title != "t" || len(got.Labels) != 1 {
			t.Fatalf("row = %+v", got)
		}
	})
	t.Run("the native notes fence is what a replacement says, and it is said", func(t *testing.T) {
		writePost(t, ts, "createBead", `{"id":"wr-notes","type":"`+writeTask+`","properties":{"title":"t","status":"open","priority":2,"notes":"old"}}`).succeeded(t, writeOutcomeCreated, true)
		update(`{"bead":"wr-notes","propertiesChange":[{"op":"replace","path":"/notes","value":"new"}]}`).succeeded(t, writeOutcomeUpdated, true)
	})
	t.Run("provenance", func(t *testing.T) {
		seedBead(t, ts, "wr-prov")
		body := func(provenance string) string {
			return `{"bead":"wr-prov","metadata":{"k":1},"revisionProvenance":` + provenance + `}`
		}
		for _, provenance := range []string{`{"agent":"claude"}`, `{"message":"why"}`} {
			reply := update(body(provenance))
			reply.refused(t, bdpValidationFailed)
			reply.wantDiagnostic(t, "cannot be recorded")
		}
		for _, provenance := range []string{`[]`, `{"actor":"sam"}`, `{"actor":{"id":"sam"}}`, `{"actor":{"id":"sam","basis":"attested"}}`, `{"who":"sam"}`} {
			update(body(provenance)).refused(t, bdpMalformedRequest)
		}
		update(body(`{"actor":{"id":"sam\nnewline","basis":"writer-supplied"}}`)).refused(t, bdpValidationFailed)
		update(body(`{"agent":null,"message":null}`)).succeeded(t, writeOutcomeUpdated, true)
		if roles.lastActor() != writeNoActor {
			t.Errorf("actor = %q, want the store's no-actor name", roles.lastActor())
		}
	})
	t.Run("an unknown, deleted or wrong-kind subject is the uniform miss", func(t *testing.T) {
		for _, subject := range []string{"wr-none", "a/b", "https://elsewhere.invalid/wr-1", coreLinkID("wr-1", "wr-2", coreClock)} {
			update(`{"bead":"`+subject+`","metadata":{}}`).refused(t, bdpResourceNotFound)
		}
		if roles.lastActor() == "wr-none" {
			t.Fatal("unreachable")
		}
	})
}

// TestDeleteBead is IT-12, the guard and the deleted identity.
func TestDeleteBead(t *testing.T) {
	ts, roles := newWriteServer(t, Config{})
	t.Run("the identity reported is the final live edition", func(t *testing.T) {
		record := seedBead(t, ts, "wr-1")
		reply := writePost(t, ts, "deleteBead", `{"bead":"wr-1","expectedRevision":"`+recordRevision(t, record)+`"}`)
		reply.succeeded(t, writeOutcomeDeleted, true)
		deleted := reply.body["deleted"].(map[string]any)
		name := deleted["resource"].(map[string]any)
		if deleted["resourceKind"] != "bead" || name["id"] != record["id"] || name["type"] != record["type"] || name["revision"] != record["revision"] {
			t.Errorf("deleted = %v, record %v", deleted, record)
		}
		if _, has := reply.body["resource"]; has {
			t.Errorf("a delete carries no resource: %s", reply.raw)
		}
		resp, body := coreRequest(t, ts, http.MethodGet, "/wr-1", nil)
		wantCoreProblem(t, resp, body, coreCodeNotFound)
		writePost(t, ts, "deleteBead", `{"bead":"wr-1"}`).refused(t, bdpResourceNotFound)
	})
	t.Run("a stale guard deletes nothing", func(t *testing.T) {
		seedBead(t, ts, "wr-2")
		writePost(t, ts, "deleteBead", `{"bead":"wr-2","expectedRevision":"nope"}`).refused(t, bdpRevisionMismatch)
		if roles.find("wr-2") == nil {
			t.Fatal("deleted")
		}
	})
	t.Run("the guard is compared inside the delete", func(t *testing.T) {
		record := seedBead(t, ts, "wr-race")
		roles.mu.Lock()
		roles.race = func() {
			issue := roles.find("wr-race")
			issue.Labels = []string{"landed"}
			roles.put(issue)
		}
		roles.mu.Unlock()
		writePost(t, ts, "deleteBead", `{"bead":"wr-race","expectedRevision":"`+recordRevision(t, record)+`"}`).refused(t, bdpRevisionMismatch)
	})
	t.Run("IT-12 any incident Link refuses, and none is named", func(t *testing.T) {
		seedBead(t, ts, "wr-a")
		seedBead(t, ts, "wr-b")
		seedLink(t, ts, roles, "wr-a", "wr-b")
		for _, id := range []string{"wr-a", "wr-b"} {
			reply := writePost(t, ts, "deleteBead", `{"bead":"`+id+`"}`)
			reply.refused(t, bdpIncidentLinksExist)
			if strings.Contains(reply.raw, coreLinkIDPrefix) || strings.Contains(reply.raw, "wr-a") || strings.Contains(reply.raw, "wr-b") {
				t.Errorf("the refusal discloses a Link: %s", reply.raw)
			}
		}
		roles.store.mu.Lock()
		roles.store.edges = nil
		roles.store.mu.Unlock()
		writePost(t, ts, "deleteBead", `{"bead":"wr-a"}`).succeeded(t, writeOutcomeDeleted, true)
	})
	t.Run("an external target is an incident Link too", func(t *testing.T) {
		seedBead(t, ts, "wr-e")
		seedLink(t, ts, roles, "wr-e", "external:proj:cap")
		writePost(t, ts, "deleteBead", `{"bead":"wr-e"}`).refused(t, bdpIncidentLinksExist)
	})
	t.Run("Donna's ruling: revisionProvenance is admitted, and its actor recorded", func(t *testing.T) {
		seedBead(t, ts, "wr-p")
		writePost(t, ts, "deleteBead", `{"bead":"wr-p","revisionProvenance":{"agent":"x"}}`).refused(t, bdpValidationFailed)
		writePost(t, ts, "deleteBead", `{"bead":"wr-p","revisionProvenance":{"actor":{"id":"sam","basis":"writer-supplied"}}}`).succeeded(t, writeOutcomeDeleted, true)
		if roles.lastActor() != "sam" {
			t.Errorf("actor = %q", roles.lastActor())
		}
	})
	t.Run("delete modes are not admitted", func(t *testing.T) {
		seedBead(t, ts, "wr-m")
		for _, member := range []string{"blockedBy", "cascade", "dryRun", "beads"} {
			writePost(t, ts, "deleteBead", `{"bead":"wr-m","`+member+`":true}`).refused(t, bdpMalformedRequest)
		}
	})
}

// seedLink stores a live edge as `issue/dep/add` would, since Write makes none, and
// answers its record as a read serves it.
func seedLink(t *testing.T, ts *testServer, roles *writeRoles, source, target string) map[string]any {
	t.Helper()
	roles.store.mu.Lock()
	roles.store.edges = append(roles.store.edges, &types.Dependency{
		IssueID: source, DependsOnID: target, Type: types.DepBlocks, CreatedBy: "sam", CreatedAt: coreClock,
	})
	roles.store.mu.Unlock()
	_, list := coreGetJSON(t, ts, "/"+source+"?view=links&direction=outbound")
	items, _ := list["items"].([]any)
	for _, item := range items {
		record := item.(map[string]any)
		if record["target"] == ts.base+"/"+target || record["target"] == target {
			return record
		}
	}
	t.Fatalf("no served Link %s -> %s in %v", source, target, list)
	return nil
}

// TestLinkWritesAreExceptions: Donna's ruling, that all three wait until Links have
// identity and revisions comparable to Beads. Each refuses, always, with the words of
// the only unsupported-operation refusal Write has, reading neither body nor state.
func TestLinkWritesAreExceptions(t *testing.T) {
	ts, roles := newWriteServer(t, Config{})
	seedBead(t, ts, "wr-a")
	seedBead(t, ts, "wr-b")
	seedBead(t, ts, "wr-c")
	link := seedLink(t, ts, roles, "wr-a", "wr-b")
	id := link["id"].(string)
	bodies := map[string][]string{
		"createLink": {
			`{"type":"` + writeBlock + `","source":"wr-a","target":"wr-c"}`,
			`{"type":"` + writeBlock + `","source":"wr-a","target":"wr-b"}`,
			`{"type":"` + writeBlock + `","source":"wr-a","target":"wr-a"}`,
			`{"type":"` + writeBlock + `","source":"no-such","target":"wr-c"}`,
		},
		"updateLink": {`{"link":"` + id + `","metadata":{"m":1}}`, `{"link":"no-such"}`},
		"deleteLink": {`{"link":"` + id + `"}`, `{"link":"` + id + `","expectedRevision":"` + recordRevision(t, link) + `"}`, `{"link":"no-such"}`},
	}
	// A store that cannot answer shows the refusal asks it nothing.
	roles.store.mu.Lock()
	roles.store.getErr = errors.New("the store was read")
	roles.store.mu.Unlock()
	for member, sent := range bodies {
		t.Run(member, func(t *testing.T) {
			for _, body := range append(sent, `not json at all`, `{}`, strings.Repeat("x", writeBodyBytes+1)) {
				reply := writePost(t, ts, member, body)
				reply.refused(t, bdpOperationNotSupported)
				if reply.status != http.StatusUnprocessableEntity || reply.body["operation"] != member {
					t.Errorf("problem = %d %s", reply.status, reply.raw)
				}
				detail, _ := reply.body["detail"].(string)
				if !strings.Contains(detail, member) || strings.Contains(detail, "Delete the Link") {
					t.Errorf("the refusal's words are not true of %s: %s", member, reply.raw)
				}
			}
			writePostAs(t, ts, member, `{}`, http.Header{"Content-Type": {"application/json"}}).refused(t, bdpUnsupportedMedia)
		})
	}
	roles.store.mu.Lock()
	roles.store.getErr = nil
	edges := len(roles.store.edges)
	roles.store.mu.Unlock()
	if edges != 1 {
		t.Fatalf("edges = %d, want the one seeded", edges)
	}
	_, read := coreGetJSON(t, ts, strings.TrimPrefix(id, ts.base))
	if read["revision"] != link["revision"] {
		t.Error("the Link moved")
	}
}

// TestCoreNumberAdmission holds the rule to the examples Core gives.
func TestCoreNumberAdmission(t *testing.T) {
	for literal, want := range map[string]bool{
		"1e300": true, "0": true, "-0.0": true, "1.5": true, "1.50": true, "9007199254740992": true,
		"9007199254740993": false, "1.2345678901234567890": false, "1e400": false, "1e-400": false,
	} {
		if got := coreNumberAdmissible(literal); got != want {
			t.Errorf("%s: got %t, want %t", literal, got, want)
		}
	}
}
