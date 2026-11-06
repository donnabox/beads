//go:build cgo

package httpclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/steveyegge/beads/internal/httpapi"
	"github.com/steveyegge/beads/internal/httpapi/apigen"
	"github.com/steveyegge/beads/internal/httpclient/wire"
	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/issueops"
)

// The operations no native command reached in the parity run, end to end: the real
// client, on the wire as Callables, to a real server over real storage. The role
// contracts elsewhere in this tier cover what each one means. What this adds is the
// statement that nothing but POST /bdp/call/NAME went out to get those answers.

// wireLog is the client's transport, recording what it really sent.
type wireLog struct {
	mu       sync.Mutex
	requests []string
}

func (l *wireLog) RoundTrip(req *http.Request) (*http.Response, error) {
	l.mu.Lock()
	l.requests = append(l.requests, req.Method+" "+req.URL.EscapedPath())
	l.mu.Unlock()
	return http.DefaultTransport.RoundTrip(req)
}

func (l *wireLog) count() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.requests)
}

// wantInvocations says every request was an invocation and that each name was one.
func (l *wireLog) wantInvocations(t *testing.T, names ...string) {
	t.Helper()
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, request := range l.requests {
		if !strings.HasPrefix(request, "POST /bdp/call/") {
			t.Errorf("the client sent %q", request)
		}
	}
	for _, name := range names {
		if !slices.Contains(l.requests, "POST /bdp/call/"+name) {
			t.Errorf("no invocation of %s among %v", name, l.requests)
		}
	}
}

// recordedSubject is env's server behind a client whose every request is logged.
func recordedSubject(t *testing.T, env *servedEnv) (*Store, *wireLog) {
	t.Helper()
	log := &wireLog{}
	client, err := wire.New(env.base, nil, wire.Options{HTTPClient: &http.Client{Transport: log}})
	if err != nil {
		t.Fatalf("build the client: %v", err)
	}
	return New(Target{BaseURL: env.base}, client, nil), log
}

func servedCallablesIssue(env *servedEnv, suffix string) *types.Issue {
	return &types.Issue{
		ID: env.prefix + "-" + suffix, Title: "issue " + suffix,
		Status: types.StatusOpen, Priority: 2, IssueType: types.TypeTask,
	}
}

func TestServedCallablesClaimAndRelease(t *testing.T) {
	env := newServedEnv(t, "hclaim")
	ctx := t.Context()
	subject, log := recordedSubject(t, env)
	issue := servedCallablesIssue(env, "a")
	if err := env.createIssue(ctx, issue, "seed"); err != nil {
		t.Fatal(err)
	}

	claimer, err := subject.IssueClaimer()
	if err != nil {
		t.Fatal(err)
	}
	claimed, err := claimer.Claim(ctx, issueops.ClaimRequest{Actor: "kim", IssueID: issue.ID})
	if err != nil {
		t.Fatalf("Claim: %v", err)
	}
	if !claimed.Changed || claimed.Issue.Assignee != "kim" {
		t.Errorf("Claim = %+v", claimed)
	}
	if row, err := env.getIssue(ctx, issue.ID); err != nil || row.Assignee != "kim" || row.Status != types.StatusInProgress {
		t.Errorf("after the claim: %+v, %v", row, err)
	}

	// A refusal comes back through `error` and is rebuilt as the typed conflict.
	var conflict *issueops.ClaimConflictError
	if _, err := claimer.Claim(ctx, issueops.ClaimRequest{Actor: "sam", IssueID: issue.ID}); !errors.As(err, &conflict) || conflict.Assignee != "kim" {
		t.Errorf("a foreign claim: %v", err)
	}

	releaser, err := subject.Releaser()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := releaser.Release(ctx, issueops.ReleaseRequest{Actor: "kim", IssueID: issue.ID}); err != nil {
		t.Fatalf("Release: %v", err)
	}
	if row, err := env.getIssue(ctx, issue.ID); err != nil || row.Assignee != "" || row.Status != types.StatusOpen {
		t.Errorf("after the release: %+v, %v", row, err)
	}
	log.wantInvocations(t, "issue/context", "issue/claim", "issue/unclaim")
}

func TestServedCallablesCompareAndSetMetadata(t *testing.T) {
	env := newServedEnv(t, "hcas")
	ctx := t.Context()
	subject, log := recordedSubject(t, env)
	issue := servedCallablesIssue(env, "a")
	if err := env.createIssue(ctx, issue, "seed"); err != nil {
		t.Fatal(err)
	}
	cas, err := subject.MetadataCAS()
	if err != nil {
		t.Fatal(err)
	}
	raw := func(s string) *json.RawMessage {
		v := json.RawMessage(s)
		return &v
	}

	won, err := cas.CompareAndSetKey(ctx, issueops.CompareAndSetKeyRequest{Actor: "kim", IssueID: issue.ID, Key: "lock", Value: raw(`"kim"`)})
	if err != nil || !won.Swapped {
		t.Fatalf("first writer: %+v, %v", won, err)
	}
	lost, err := cas.CompareAndSetKey(ctx, issueops.CompareAndSetKeyRequest{Actor: "sam", IssueID: issue.ID, Key: "lock", Value: raw(`"sam"`)})
	if err != nil || lost.Swapped {
		t.Fatalf("second writer: %+v, %v", lost, err)
	}
	if lost.Current == nil || string(*lost.Current) != `"kim"` {
		t.Errorf("a lost race reports the holder: %+v", lost)
	}
	row, err := env.getIssue(ctx, issue.ID)
	if err != nil || !strings.Contains(string(row.Metadata), `"kim"`) {
		t.Errorf("after the race: %s, %v", row.Metadata, err)
	}
	log.wantInvocations(t, "issue/metadata/compare-and-set")
}

func TestServedCallablesBatchCreate(t *testing.T) {
	env := newServedEnv(t, "hbatch")
	ctx := t.Context()
	subject, log := recordedSubject(t, env)
	creator, err := subject.BatchCreator()
	if err != nil {
		t.Fatal(err)
	}
	item := func(title string) issueops.BatchCreateItem {
		return issueops.BatchCreateItem{Issue: &types.Issue{Title: title, Priority: 2, IssueType: types.TypeTask}}
	}
	result, err := creator.CreateBatch(ctx, issueops.CreateBatchRequest{Actor: "kim", Items: []issueops.BatchCreateItem{item("one"), item("two")}})
	if err != nil {
		t.Fatalf("CreateBatch: %v", err)
	}
	if len(result.Issues) != 2 || result.Issues[0].Title != "one" || result.Issues[1].Title != "two" {
		t.Fatalf("CreateBatch = %+v", result)
	}
	for _, issue := range result.Issues {
		if row, err := env.getIssue(ctx, issue.ID); err != nil || row.Title != issue.Title {
			t.Errorf("%s after the batch: %+v, %v", issue.ID, row, err)
		}
	}
	// All or nothing, on a refusal that came back through the mount.
	if _, err := creator.CreateBatch(ctx, issueops.CreateBatchRequest{Actor: "kim", Items: []issueops.BatchCreateItem{item("three"), item("")}}); !errors.Is(err, issueops.ErrValidation) {
		t.Errorf("an empty title: %v", err)
	}
	var rows int
	if err := env.queryScalar(ctx, "SELECT COUNT(*) FROM issues", nil, &rows); err != nil {
		t.Fatal(err)
	}
	if rows != 2 {
		t.Errorf("%d rows after a refused batch, want 2", rows)
	}
	log.wantInvocations(t, "issue/create-many")
}

func TestServedCallablesCountDependencyEdges(t *testing.T) {
	env := newServedEnv(t, "hcount")
	ctx := t.Context()
	subject, log := recordedSubject(t, env)
	a, b, c := servedCallablesIssue(env, "a"), servedCallablesIssue(env, "b"), servedCallablesIssue(env, "c")
	for _, issue := range []*types.Issue{a, b, c} {
		if err := env.createIssue(ctx, issue, "seed"); err != nil {
			t.Fatal(err)
		}
	}
	for _, target := range []*types.Issue{b, c} {
		if err := env.addDependency(ctx, &types.Dependency{IssueID: a.ID, DependsOnID: target.ID, Type: types.DepBlocks}, "seed"); err != nil {
			t.Fatal(err)
		}
	}
	counter, err := subject.GraphCounter()
	if err != nil {
		t.Fatal(err)
	}
	missing := env.prefix + "-none"
	result, err := counter.CountEdges(ctx, issueops.EdgeCountRequest{IDs: []string{a.ID, b.ID, missing}, Direction: issueops.EdgeDirectionOut})
	if err != nil {
		t.Fatalf("CountEdges: %v", err)
	}
	if len(result.Anchors) != 3 || result.Anchors[0].Count != 2 || result.Anchors[1].Count != 0 || !result.Anchors[2].Missing {
		t.Errorf("CountEdges = %+v", result.Anchors)
	}
	log.wantInvocations(t, "issue/dep/count")
}

func TestServedCallablesListEvents(t *testing.T) {
	t.Run("the journal off is the operation's own refusal", func(t *testing.T) {
		env := newServedEnv(t, "hjoff", func(cfg *httpapi.Config) {
			cfg.EventsJournal, cfg.EventsJournalEnabled = nil, false
		})
		subject, log := recordedSubject(t, env)
		_, err := subject.ReadEventsJournalPage(t.Context(), 0, 10)
		if !errors.Is(err, wire.ErrEventsJournalDisabled) {
			t.Errorf("err = %v", err)
		}
		log.wantInvocations(t, "events/list")
	})
	t.Run("the journal on reads what was written", func(t *testing.T) {
		env := newServedEnv(t, "hjon")
		ctx := t.Context()
		env.reference.SetEventsJournalEnabled(true)
		t.Cleanup(func() { env.reference.SetEventsJournalEnabled(false) })
		subject, log := recordedSubject(t, env)
		issue := servedCallablesIssue(env, "a")
		if err := env.createIssue(ctx, issue, "seed"); err != nil {
			t.Fatal(err)
		}
		page, err := subject.ReadEventsJournalPage(ctx, 0, 10)
		if err != nil {
			t.Fatalf("ReadEventsJournalPage: %v", err)
		}
		if len(page.Rows) == 0 || page.Head == 0 {
			t.Fatalf("page = %+v", page)
		}
		// `limit` and `since` arrive as the numbers they are, by way of JSON strings.
		capped, err := subject.ReadEventsJournalPage(ctx, page.Rows[0].Seq, 1)
		if err != nil {
			t.Fatalf("since the first row: %v", err)
		}
		if len(capped.Rows) > 0 && capped.Rows[0].Seq <= page.Rows[0].Seq {
			t.Errorf("since was not honored: %+v", capped.Rows)
		}
		log.wantInvocations(t, "events/list")
	})
}

// TestServedCallablesListIssuesSecondPage is the page walk the parity run never
// exercised: a cursor the server minted, handed back as an argument.
func TestServedCallablesListIssuesSecondPage(t *testing.T) {
	env := newServedEnv(t, "hpage")
	ctx := t.Context()
	subject, log := recordedSubject(t, env)
	var want []string
	for i := range 5 {
		issue := servedCallablesIssue(env, fmt.Sprint(i))
		if err := env.createIssue(ctx, issue, "seed"); err != nil {
			t.Fatal(err)
		}
		want = append(want, issue.ID)
	}

	var got []string
	pages := 0
	cursor := ""
	for {
		query := url.Values{"limit": {"2"}}
		if cursor != "" {
			query.Set("cursor", cursor)
		}
		var page apigen.IssuesPage
		if err := subject.dispatch(ctx, listRequest(query), &page); err != nil {
			t.Fatalf("page %d: %v", pages, err)
		}
		pages++
		for _, item := range page.Items {
			got = append(got, item.ID)
		}
		if !page.HasMore {
			break
		}
		if page.NextCursor == nil || *page.NextCursor == "" {
			t.Fatalf("page %d says has_more with no cursor", pages)
		}
		cursor = *page.NextCursor
	}
	if pages != 3 {
		t.Errorf("walked %d pages of 2 over 5 rows", pages)
	}
	slices.Sort(got)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Errorf("pages held %v, want %v", got, want)
	}
	log.wantInvocations(t, "issue/list")

	var page apigen.IssuesPage
	bad := listRequest(url.Values{"cursor": {"not-a-cursor"}})
	if err := subject.dispatch(ctx, bad, &page); !errors.Is(err, wire.ErrInvalidCursor) {
		t.Errorf("a forged cursor: %v", err)
	}

	// The role's own walk, which holds the cursor itself once a page is smaller than
	// what the caller asked for.
	saved := walkPageSize
	walkPageSize = 2
	t.Cleanup(func() { walkPageSize = saved })
	reader, err := subject.IssueReader()
	if err != nil {
		t.Fatal(err)
	}
	before := log.count()
	limit := 0
	walked, err := reader.List(ctx, issueops.ListRequest{Limit: &limit})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(walked.Items) != 5 || walked.HasMore {
		t.Errorf("List = %d items, has_more %v", len(walked.Items), walked.HasMore)
	}
	if sent := log.count() - before; sent < 3 {
		t.Errorf("the walk took %d requests over pages of 2", sent)
	}
}
