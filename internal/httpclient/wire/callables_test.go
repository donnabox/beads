package wire

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/steveyegge/beads/internal/httpapi"
	"github.com/steveyegge/beads/internal/httpapi/apigen"
	"github.com/steveyegge/beads/internal/httpclient/wire/wiretest"
	"github.com/steveyegge/beads/issueops"
)

// The move to /bdp/call: what goes on the wire, and what comes back from it. Every
// other test in this package sees the v0 vocabulary through wiretest.V0; these look
// at the real bytes.

// wireClient is a client aimed at a server that answers with h, unwrapped.
func wireClient(t *testing.T, opts Options, h http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	base, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	c, err := New(base, nil, opts)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func answer(status int, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(status)
		_, _ = io.WriteString(w, body)
	}
}

// TestEveryRequestIsAnInvocation pins the wire form: POST /bdp/call/NAME, the two
// media types, nothing in the URL, and the operation's parameters as one object.
func TestEveryRequestIsAnInvocation(t *testing.T) {
	cases := []struct {
		name       string
		r          Request
		want       string // the Callable
		args       string // the invocation's arguments, as JSON
		provenance string // the invocation's revisionProvenance; "" is none
	}{
		{"no parameters", Request{Op: OpGetContext, Method: http.MethodGet, Path: PathContext}, "issue/context", `{}`, ""},
		{
			"a path argument takes its name, unescaped",
			Request{Op: OpGetIssue, Method: http.MethodGet, Path: "/v0/beads/issues/a%2Fb%3Ac", PathArg: "a/b:c"},
			"issue/show", `{"id":"a/b:c"}`, "",
		},
		{
			"a setting's key", Request{Op: OpGetSetting, Method: http.MethodGet, Path: "/v0/beads/config/k", PathArg: "k"},
			"config/get", `{"key":"k"}`, "",
		},
		{
			"one query value is a string and a repeated one an array",
			Request{
				Op: OpListIssues, Method: http.MethodGet, Path: PathIssues,
				Query: url.Values{"status": {"open", "closed"}, "limit": {"10"}, "assignee": {""}, "none": {}},
			},
			"issue/list", `{"status":["open","closed"],"limit":"10","assignee":"","sort":"created"}`, "",
		},
		{
			"a body's members", Request{Op: OpCreateIssue, Method: http.MethodPost, Path: PathIssues, Body: map[string]any{"title": "t", "priority": 2}},
			"issue/create", `{"title":"t","priority":2,"issue_type":""}`, "",
		},
		{
			"path and body", Request{Op: OpClaimIssue, Method: http.MethodPost, Path: "/x", PathArg: "ga-1", Body: map[string]string{"actor": "a"}},
			"issue/claim", `{"id":"ga-1"}`, `{"actor":{"id":"a","basis":"writer-supplied"}}`,
		},
		{
			"query and body", Request{Op: OpClaimNextIssue, Method: http.MethodPost, Path: PathIssuesClaimNext, Query: url.Values{"label": {"x"}}, Body: map[string]string{"actor": "a"}},
			"issue/claim-next", `{"label":"x"}`, `{"actor":{"id":"a","basis":"writer-supplied"}}`,
		},
		{
			"a PATCH", Request{Op: OpUpdateIssue, Method: http.MethodPatch, Path: "/x", PathArg: "ga-1", Body: map[string]string{"actor": "a"}},
			"issue/update", `{"id":"ga-1"}`, `{"actor":{"id":"a","basis":"writer-supplied"}}`,
		},
		{
			"a DELETE with a path argument alone", Request{Op: OpForgetMemory, Method: http.MethodDelete, Path: "/x", PathArg: "k"},
			"memory/forget", `{"key":"k"}`, "",
		},
		{
			"sweep, ephemeral", Request{Op: OpSweepIssues, Method: http.MethodPost, Path: PathIssuesSweep, Body: map[string]string{"tier": "ephemeral"}},
			"issue/purge", `{"tier":"ephemeral","dry_run":false,"protect_referenced":true,"protect_live_dependents":false}`, "",
		},
		{
			"sweep, durable", Request{Op: OpSweepIssues, Method: http.MethodPost, Path: PathIssuesSweep, Body: map[string]string{"tier": "durable"}},
			"issue/prune", `{"tier":"durable","dry_run":false,"protect_referenced":true,"protect_live_dependents":false}`, "",
		},
		{
			"sweep, wisps-plane", Request{Op: OpSweepIssues, Method: http.MethodPost, Path: PathIssuesSweep, Body: map[string]string{"tier": "wisps-plane"}},
			"issue/purge", `{"tier":"wisps-plane","dry_run":false,"protect_referenced":true,"protect_live_dependents":false}`, "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var seen recorded
			c := wireClient(t, Options{}, func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				seen = recorded{method: r.Method, path: r.URL.Path, rawQuery: r.URL.RawQuery, header: r.Header.Clone(), body: body}
				_, _ = io.WriteString(w, `{"result":{}}`)
			})
			if err := c.Do(ctx(t), tc.r, &map[string]any{}); err != nil {
				t.Fatalf("Do: %v", err)
			}
			if seen.method != http.MethodPost || seen.path != "/bdp/call/"+tc.want || seen.rawQuery != "" {
				t.Errorf("sent %s %s?%s, want POST /bdp/call/%s", seen.method, seen.path, seen.rawQuery, tc.want)
			}
			if got := seen.header.Get("Content-Type"); got != "application/vnd.bdp.v0.1+json" {
				t.Errorf("Content-Type = %q", got)
			}
			if got := seen.header.Get("Accept"); !strings.Contains(got, "application/vnd.bdp.v0.1+json") {
				t.Errorf("Accept = %q", got)
			}
			if _, present := seen.header["Idempotency-Key"]; present {
				t.Errorf("an Idempotency-Key went on the wire: %v", seen.header)
			}
			var got struct {
				Arguments  map[string]any `json:"arguments"`
				Provenance any            `json:"revisionProvenance"`
			}
			if err := json.Unmarshal(seen.body, &got); err != nil {
				t.Fatalf("body %s: %v", seen.body, err)
			}
			var want map[string]any
			if err := json.Unmarshal([]byte(tc.args), &want); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got.Arguments, want) {
				t.Errorf("arguments = %s, want %s", seen.body, tc.args)
			}
			var wantProvenance any
			if tc.provenance != "" {
				if err := json.Unmarshal([]byte(tc.provenance), &wantProvenance); err != nil {
					t.Fatal(err)
				}
			}
			if !reflect.DeepEqual(got.Provenance, wantProvenance) {
				t.Errorf("revisionProvenance in %s, want %s", seen.body, tc.provenance)
			}
		})
	}
}

// TestOnlyHealthGoesPlain: nothing reaches v0 by accident. health stays a plain
// route; an operation with no Callable, or a Request that lost its path argument, is
// refused before anything is sent.
func TestOnlyHealthGoesPlain(t *testing.T) {
	var seen []string
	c := wireClient(t, Options{}, func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.Method+" "+r.URL.Path)
		_, _ = io.WriteString(w, `{"status":"ok"}`)
	})
	if err := c.Health(ctx(t)); err != nil {
		t.Fatalf("Health: %v", err)
	}
	if !slices.Equal(seen, []string{"GET /healthz"}) {
		t.Errorf("Health sent %v", seen)
	}

	seen = nil
	for name, r := range map[string]Request{
		"an unmapped operation":  {Op: "madeUp", Method: http.MethodGet, Path: "/v0/x"},
		"a lost path argument":   {Op: OpGetIssue, Method: http.MethodGet, Path: "/v0/beads/issues/a"},
		"a stray path argument":  {Op: OpListIssues, Method: http.MethodGet, Path: PathIssues, PathArg: "a"},
		"a body that collides":   {Op: OpClaimIssue, Method: http.MethodPost, Path: "/x", PathArg: "a", Body: map[string]string{"id": "b"}},
		"a body that is no JSON": {Op: OpClaimIssue, Method: http.MethodPost, Path: "/x", PathArg: "a", Body: []string{"b"}},
	} {
		if err := c.Do(ctx(t), r, nil); err == nil {
			t.Errorf("%s was not refused", name)
		}
	}
	if len(seen) != 0 {
		t.Errorf("a refused request was sent: %v", seen)
	}
}

// TestResultEnvelope: v0's own body is what gets decoded, and an answer without one
// is an error rather than an empty success.
func TestResultEnvelope(t *testing.T) {
	var out apigen.ContextResponse
	read := Request{Op: OpGetContext, Method: http.MethodGet, Path: PathContext}

	c := wireClient(t, Options{}, answer(200, `{"result":{"api_version":"v0","project_id":"p"}}`))
	if err := c.Do(ctx(t), read, &out); err != nil || out.ProjectId != "p" {
		t.Errorf("Do = %v, out = %+v", err, out)
	}
	c = wireClient(t, Options{}, answer(200, `{"changed":false,"result":{"project_id":"q"}}`))
	if err := c.Do(ctx(t), read, &out); err != nil || out.ProjectId != "q" {
		t.Errorf("Do = %v, out = %+v", err, out)
	}
	for _, body := range []string{`{"api_version":"v0"}`, `{"changed":true}`, `not json`} {
		c = wireClient(t, Options{}, answer(200, body))
		if err := c.Do(ctx(t), read, &out); err == nil {
			t.Errorf("%s decoded as a success", body)
		}
	}
}

// TestResponseCapIsOnV0sBody: the envelope is not charged against the cap. A body
// exactly at it still reads, and one a byte over is refused, as they were on v0.
func TestResponseCapIsOnV0sBody(t *testing.T) {
	const limit = 1024
	bodyOf := func(size int) string { return `"` + strings.Repeat("x", size-2) + `"` }
	read := Request{Op: OpGetIssue, Method: http.MethodGet, Path: "/x", PathArg: "a"}

	c := wireClient(t, Options{MaxResponseBytes: limit}, answer(200, `{"result":`+bodyOf(limit)+"}\n"))
	var out string
	if err := c.Do(ctx(t), read, &out); err != nil {
		t.Errorf("a body at the cap: %v", err)
	}
	c = wireClient(t, Options{MaxResponseBytes: limit}, answer(200, `{"result":`+bodyOf(limit+1)+"}\n"))
	if err := c.Do(ctx(t), read, &out); !errors.Is(err, ErrResponseTooLarge) {
		t.Errorf("a body over the cap: %v", err)
	}
}

// TestRefusalsAreReadFromV0sOwnBody is the decision that matters most in the mapping:
// the status above the mapper is v0's, so §5's 503 for `internal` does not newly
// retry, and every typed conflict is rebuilt as it was.
func TestRefusalsAreReadFromV0sOwnBody(t *testing.T) {
	read := Request{Op: OpGetIssue, Method: http.MethodGet, Path: "/x", PathArg: "ga-1", IssueID: "ga-1"}
	problem := func(outer int, code, member, inner string) string {
		return fmt.Sprintf(`{"type":"https://github.com/gastownhall/bdp/problems/x","status":%d,"code":%q,"retry":"never",%q:%s}`, outer, code, member, inner)
	}

	t.Run("an outer 503 for v0's internal does not retry", func(t *testing.T) {
		body := problem(503, "temporarily-unavailable", "v0", `{"status":500,"code":"internal","detail":"d","request_id":"req-1"}`)
		c := wireClient(t, Options{}, answer(503, body))
		var prob *ProblemError
		if err := c.Do(ctx(t), read, nil); !errors.As(err, &prob) {
			t.Fatalf("err = %v", err)
		}
		if prob.Status != 500 || prob.Code != "internal" || prob.RequestID != "req-1" || prob.Retryable() || !errors.Is(prob, ErrServerFault) {
			t.Errorf("got %+v", prob)
		}
	})
	t.Run("v0's busy still retries, and keeps its Retry-After", func(t *testing.T) {
		body := problem(503, "temporarily-unavailable", "v0", `{"status":503,"code":"busy","detail":"d"}`)
		c := wireClient(t, Options{}, func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Retry-After", "2")
			w.WriteHeader(503)
			_, _ = io.WriteString(w, body)
		})
		var prob *ProblemError
		if err := c.Do(ctx(t), read, nil); !errors.As(err, &prob) {
			t.Fatalf("err = %v", err)
		}
		if !prob.Retryable() || prob.RetryAfter.Seconds() != 2 || !errors.Is(prob, ErrBusy) {
			t.Errorf("got %+v", prob)
		}
	})
	t.Run("a resource-not-found is v0's not_found", func(t *testing.T) {
		body := problem(404, "resource-not-found", "v0", `{"status":404,"code":"not_found","detail":"d"}`)
		c := wireClient(t, Options{}, answer(404, body))
		if err := c.Do(ctx(t), read, nil); !errors.Is(err, issueops.ErrNotFound) {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("callable-error carries v0 under `error`, with v0's own status", func(t *testing.T) {
		body := problem(409, "callable-error", "error", `{"status":410,"code":"events_journal_truncated","floor":3,"head":9}`)
		c := wireClient(t, Options{}, answer(409, body))
		var prob *ProblemError
		if err := c.Do(ctx(t), read, nil); !errors.As(err, &prob) {
			t.Fatalf("err = %v", err)
		}
		if prob.Status != 410 || !errors.Is(prob, ErrEventsJournalTruncated) || prob.Floor == nil || *prob.Floor != 3 {
			t.Errorf("got %+v", prob)
		}
	})
	t.Run("a typed conflict is rebuilt from the carried members", func(t *testing.T) {
		body := problem(409, "callable-error", "error", `{"status":409,"code":"already_claimed","assignee":"kim"}`)
		c := wireClient(t, Options{}, answer(409, body))
		err := c.Do(ctx(t), read, nil)
		var conflict *issueops.ClaimConflictError
		if !errors.As(err, &conflict) || conflict.Assignee != "kim" || conflict.IssueID != "ga-1" {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("a project mismatch stays one", func(t *testing.T) {
		body := problem(422, "validation-failed", "v0", `{"status":400,"code":"invalid_argument","reason":"project_mismatch","server_project_id":"other"}`)
		c := wireClient(t, Options{ExpectProjectID: "mine"}, answer(422, body))
		var mismatch *ProjectMismatchError
		if err := c.Do(ctx(t), read, nil); !errors.As(err, &mismatch) || mismatch.Got != "other" || mismatch.Expected != "mine" {
			t.Errorf("err = %v", err)
		}
	})
}

// TestTheMountsOwnProblems: answers about the invocation, with no v0 body to read.
func TestTheMountsOwnProblems(t *testing.T) {
	read := Request{Op: OpGetIssue, Method: http.MethodGet, Path: "/x", PathArg: "ga-1"}
	own := func(status int, code, detail string) string {
		return fmt.Sprintf(`{"type":"https://github.com/gastownhall/bdp/problems/x","status":%d,"code":%q,"retry":"never","detail":%q}`, status, code, detail)
	}
	cases := []struct {
		name      string
		status    int
		body      string
		want      error
		retryable bool
		says      string
	}{
		{"unsupported media type is this client's bug", 415, own(415, "unsupported-media-type", "the type"), ErrBadRequest, false, "the type"},
		{"a malformed invocation is this client's bug", 400, own(400, "malformed-request", "the shape"), ErrBadRequest, false, "the shape"},
		{"too large is v0's refusal of an oversized body", 413, own(413, "request-too-large", "too big"), issueops.ErrValidation, false, "too big"},
		{
			"an undeclared argument names it", 422,
			`{"status":422,"code":"validation-failed","detail":"not declared","diagnostics":[{"message":"` + "`nope`" + ` is not a declared parameter","instanceLocation":"/arguments/nope"}]}`,
			issueops.ErrValidation, false, "`nope` is not a declared parameter",
		},
		{"the mount's own 503 says come back", 503, own(503, "temporarily-unavailable", ""), ErrBusy, true, "503"},
		{"a bodyless 405", 405, "", ErrBadRequest, false, "does not take a POST"},
		{"a bodyless 406", 406, "", ErrBadRequest, false, "application/vnd.bdp.v0.1+json"},
		{"something in front of bd serve reads as it always has", 502, "<html>bad gateway</html>", ErrServerFault, false, "502"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := wireClient(t, Options{}, answer(tc.status, tc.body))
			err := c.Do(ctx(t), read, nil)
			var prob *ProblemError
			if !errors.As(err, &prob) {
				t.Fatalf("err = %v", err)
			}
			if !errors.Is(err, tc.want) || errors.Is(err, issueops.ErrNotFound) {
				t.Errorf("err = %v, want %v", err, tc.want)
			}
			if prob.Retryable() != tc.retryable {
				t.Errorf("Retryable = %v", prob.Retryable())
			}
			if !strings.Contains(err.Error(), tc.says) {
				t.Errorf("err = %q, want it to say %q", err, tc.says)
			}
		})
	}

	t.Run("the undeclared argument is the Param", func(t *testing.T) {
		c := wireClient(t, Options{}, answer(422, `{"status":422,"code":"validation-failed","diagnostics":[{"message":"m","instanceLocation":"/arguments/nope"}]}`))
		var prob *ProblemError
		if err := c.Do(ctx(t), read, nil); !errors.As(err, &prob) || prob.Param != "nope" {
			t.Errorf("err = %v, prob = %+v", err, prob)
		}
	})
}

// TestAServerFromAnotherReleaseIsToldApart: lockstep, in words. An older server
// answers /bdp/call with v0's catch-all 404, which must not read as "no such entity".
// That is the handshake's first call, and the first call of a command that makes none.
func TestAServerFromAnotherReleaseIsToldApart(t *testing.T) {
	noRoute := `{"type":"about:blank","title":"Not Found","status":404,"code":"not_found","detail":"no such route on this server","request_id":"r"}`

	t.Run("no mount at all, on the handshake", func(t *testing.T) {
		c := wireClient(t, Options{}, answer(404, noRoute))
		_, err := c.Handshake(ctx(t))
		var absent *CallablesAbsentError
		if !errors.As(err, &absent) || !absent.NoMount || !errors.Is(err, ErrCallablesAbsent) {
			t.Fatalf("err = %v", err)
		}
		if errors.Is(err, issueops.ErrNotFound) {
			t.Errorf("reads as no such entity: %v", err)
		}
		for _, want := range []string{"predates Callables", "upgrade"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("err = %q, want it to say %q", err, want)
			}
		}
	})
	t.Run("no mount at all, on a read", func(t *testing.T) {
		c := wireClient(t, Options{}, answer(404, noRoute))
		err := c.Do(ctx(t), Request{Op: OpGetIssue, Method: http.MethodGet, Path: "/x", PathArg: "ga-1"}, nil)
		if !errors.Is(err, ErrCallablesAbsent) || errors.Is(err, issueops.ErrNotFound) {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("a mount without this Callable", func(t *testing.T) {
		c := wireClient(t, Options{}, answer(404, `{"status":404,"code":"callable-not-found","detail":"d"}`))
		err := c.Do(ctx(t), Request{Op: OpGetIssue, Method: http.MethodGet, Path: "/x", PathArg: "ga-1"}, nil)
		var absent *CallablesAbsentError
		if !errors.As(err, &absent) || absent.NoMount || errors.Is(err, issueops.ErrNotFound) {
			t.Fatalf("err = %v", err)
		}
		if !strings.Contains(err.Error(), "issue/show") {
			t.Errorf("err = %q does not name the Callable", err)
		}
	})
	t.Run("a real miss is still a miss", func(t *testing.T) {
		c := wireClient(t, Options{}, answer(404, `{"status":404,"code":"resource-not-found","v0":{"status":404,"code":"not_found"}}`))
		err := c.Do(ctx(t), Request{Op: OpGetIssue, Method: http.MethodGet, Path: "/x", PathArg: "ga-1"}, nil)
		if !errors.Is(err, issueops.ErrNotFound) || errors.Is(err, ErrCallablesAbsent) {
			t.Errorf("err = %v", err)
		}
	})
}

// TestReplayAndMemoSeeTheLogicalRequest: POST on the wire, still a read to the client.
func TestReplayAndMemoSeeTheLogicalRequest(t *testing.T) {
	var sends atomic.Int32
	c := wireClient(t, Options{}, func(w http.ResponseWriter, _ *http.Request) {
		sends.Add(1)
		_, _ = io.WriteString(w, `{"result":{"id":"bd-1"}}`)
	})
	c.EnableReadMemo()
	read := Request{Op: OpGetIssue, Method: http.MethodGet, Path: "/v0/beads/issues/bd-1", PathArg: "bd-1"}
	other := Request{Op: OpGetIssue, Method: http.MethodGet, Path: "/v0/beads/issues/bd-2", PathArg: "bd-2"}
	withComments := read
	withComments.Query = url.Values{"include_comments": {"true"}}
	write := Request{Op: OpClaimIssue, Method: http.MethodPost, Path: "/x", PathArg: "bd-1", Body: map[string]string{"actor": "a"}}
	do := func(r Request) {
		t.Helper()
		if err := c.Do(ctx(t), r, &map[string]any{}); err != nil {
			t.Fatalf("Do: %v", err)
		}
	}

	if !read.Replayable() || write.Replayable() {
		t.Errorf("Replayable: read %v, write %v", read.Replayable(), write.Replayable())
	}
	do(read)
	do(read)
	if got := sends.Load(); got != 1 {
		t.Errorf("a repeated read went out %d times; the memo did not answer it", got)
	}
	do(other)
	do(withComments)
	if got := sends.Load(); got != 3 {
		t.Errorf("sends = %d: reads of different things shared a memo entry", got)
	}
	do(read)
	if got := sends.Load(); got != 3 {
		t.Errorf("sends = %d: a read cleared the memo", got)
	}
	do(write)
	do(read)
	if got := sends.Load(); got != 5 {
		t.Errorf("sends = %d: a write did not clear the memo", got)
	}
}

// markerSeenByTransport says what net/http's transport was given, which a server
// cannot see: it is what decides the transport's own resend.
type markerSeenByTransport struct {
	marker map[string]bool
}

func (m *markerSeenByTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	values, present := req.Header["Idempotency-Key"]
	if present && len(values) != 0 {
		panic("an Idempotency-Key value would reach the wire")
	}
	m.marker[req.URL.Path] = present
	return &http.Response{
		StatusCode: 200,
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader(`{"result":{}}`)),
		Request:    req,
	}, nil
}

// TestReplayableRequestsAskTheTransportToResend: net/http resends a request on a
// stale idle connection only if it counts as idempotent, and a GET did where a POST
// does not. The empty header is how a replayable POST says so; a write must not.
func TestReplayableRequestsAskTheTransportToResend(t *testing.T) {
	seen := &markerSeenByTransport{marker: map[string]bool{}}
	base, _ := url.Parse("http://127.0.0.1:1")
	c, err := New(base, nil, Options{HTTPClient: &http.Client{Transport: seen}})
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range []Request{
		{Op: OpGetContext, Method: http.MethodGet, Path: PathContext},
		{Op: OpBatchGetIssues, Method: http.MethodPost, Path: PathIssuesBatchGet, Body: map[string]any{"ids": []string{"a"}}, Idempotent: true},
		{Op: OpCreateIssue, Method: http.MethodPost, Path: PathIssues, Body: map[string]any{"title": "t"}},
	} {
		if err := c.Do(ctx(t), r, &map[string]any{}); err != nil {
			t.Fatalf("%s: %v", r.Op, err)
		}
	}
	want := map[string]bool{"/bdp/call/issue/context": true, "/bdp/call/issue/get-many": true, "/bdp/call/issue/create": false}
	if !reflect.DeepEqual(seen.marker, want) {
		t.Errorf("marked = %v, want %v", seen.marker, want)
	}
}

// TestPostShapedReadsSurviveStaleIdleConnections is the behavior itself. The pool
// holds several connections that are dead but do not look it: each takes the next
// request's bytes and closes. net/http walks through them for a request it counts as
// idempotent, which a GET always was. Left to the client's one retry, a POST-shaped
// read lands on a second dead connection and fails.
func TestPostShapedReadsSurviveStaleIdleConnections(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	const (
		pooled   = 4
		response = "HTTP/1.1 200 OK\r\nContent-Type: application/vnd.bdp.v0.1+json\r\nContent-Length: 13\r\n\r\n{\"result\":{}}"
	)
	// The first pooled requests are answered together, so each holds its own
	// connection; every connection after them is a healthy one.
	var arrived sync.WaitGroup
	arrived.Add(pooled)
	var accepted atomic.Int32
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			first := accepted.Add(1) <= pooled
			go func() {
				defer func() { _ = conn.Close() }()
				reader := bufio.NewReader(conn)
				req, err := http.ReadRequest(reader)
				if err != nil {
					return
				}
				_, _ = io.Copy(io.Discard, req.Body)
				if first {
					arrived.Done()
					arrived.Wait()
				}
				_, _ = io.WriteString(conn, response)
				if !first {
					return
				}
				// Dead from here on: it reads what it is sent and answers nothing.
				if req, err = http.ReadRequest(reader); err == nil {
					_, _ = io.Copy(io.Discard, req.Body)
				}
			}()
		}
	}()

	base, _ := url.Parse("http://" + ln.Addr().String())
	var retried atomic.Int32
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConnsPerHost = pooled
	t.Cleanup(transport.CloseIdleConnections)
	c, err := New(base, nil, Options{
		HTTPClient:   &http.Client{Transport: transport},
		OnStallRetry: func(StallRetry) { retried.Add(1) },
	})
	if err != nil {
		t.Fatal(err)
	}
	read := Request{Op: OpGetIssue, Method: http.MethodGet, Path: "/x", PathArg: "a"}
	var filling sync.WaitGroup
	for range pooled {
		filling.Go(func() {
			if err := c.Do(ctx(t), read, &map[string]any{}); err != nil {
				t.Errorf("filling the pool: %v", err)
			}
		})
	}
	filling.Wait()

	if err := c.Do(ctx(t), read, &map[string]any{}); err != nil {
		t.Fatalf("a read after the pool went stale, with %d stall retries: %v", retried.Load(), err)
	}
	if got := retried.Load(); got != 0 {
		t.Errorf("%d stall retries: net/http did not resend this itself", got)
	}
}

// sentInvocation is what one Request put on the wire, or the error that kept it home
// with nothing dialed.
func sentInvocation(t *testing.T, r Request) (arguments map[string]json.RawMessage, provenance json.RawMessage, err error) {
	t.Helper()
	dials := 0
	c := wireClient(t, Options{}, func(w http.ResponseWriter, req *http.Request) {
		dials++
		var invocation struct {
			Arguments  map[string]json.RawMessage `json:"arguments"`
			Provenance json.RawMessage            `json:"revisionProvenance"`
		}
		body, _ := io.ReadAll(req.Body)
		if err := json.Unmarshal(body, &invocation); err != nil {
			t.Errorf("%s: %v", body, err)
		}
		arguments, provenance = invocation.Arguments, invocation.Provenance
		_, _ = io.WriteString(w, `{"result":{}}`)
	})
	err = c.Do(ctx(t), r, &map[string]any{})
	if err != nil && dials != 0 {
		t.Fatalf("Do dialed and then failed: %v", err)
	}
	return arguments, provenance, err
}

// TestTheActorGoesInRevisionProvenance: no declaration has `actor`, or `author` on a
// comment, so neither is ever an argument. Any other member named author stays where
// it is, and the server refuses it for what it is.
func TestTheActorGoesInRevisionProvenance(t *testing.T) {
	for _, tc := range []struct {
		name string
		r    Request
		id   string
	}{
		{"an operation's actor", Request{Op: OpClaimIssue, Method: http.MethodPost, Path: "/x", PathArg: "a", Body: map[string]string{"actor": "sam"}}, `"sam"`},
		{"a comment's author", Request{Op: OpAddComment, Method: http.MethodPost, Path: "/x", PathArg: "a", Body: map[string]string{"author": "kim", "text": "hi"}}, `"kim"`},
		{"the optional actor of a delete", Request{Op: OpDeleteIssues, Method: http.MethodPost, Path: PathIssues, Body: map[string]any{"actor": "sam", "ids": []string{"a"}, "dry_run": true}}, `"sam"`},
		{"sent as written: v0 trims it, and refuses what it refuses", Request{Op: OpCloseIssue, Method: http.MethodPost, Path: "/x", PathArg: "a", Body: map[string]string{"actor": "  a\tb  "}}, `"  a\tb  "`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			arguments, provenance, err := sentInvocation(t, tc.r)
			if err != nil {
				t.Fatal(err)
			}
			for _, member := range []string{"actor", "author"} {
				if _, present := arguments[member]; present {
					t.Errorf("`%s` went out in arguments: %v", member, arguments)
				}
			}
			var got struct {
				Actor map[string]json.RawMessage `json:"actor"`
			}
			if err := json.Unmarshal(provenance, &got); err != nil {
				t.Fatalf("revisionProvenance %s: %v", provenance, err)
			}
			if string(got.Actor["id"]) != tc.id || string(got.Actor["basis"]) != `"writer-supplied"` || len(got.Actor) != 2 {
				t.Errorf("revisionProvenance = %s, want the id %s", provenance, tc.id)
			}
		})
	}
}

// TestABlankActorIsStillRefused: a blank actor sends no provenance, and v0 used to
// refuse it wherever it found one. Where the Callable requires an actor it still does,
// on the server, as it does for no actor at all. Where it does not, nothing on the
// server could tell the blank from an omission, so it is refused here, undialed.
func TestABlankActorIsStillRefused(t *testing.T) {
	for _, blank := range []any{"", "   ", " \t\n", nil} {
		for _, tc := range []struct {
			name   string
			r      Request
			member string
		}{
			{"a claim", Request{Op: OpClaimIssue, Method: http.MethodPost, Path: "/x", PathArg: "a"}, "actor"},
			{"a comment", Request{Op: OpAddComment, Method: http.MethodPost, Path: "/x", PathArg: "a"}, "author"},
		} {
			tc.r.Body = map[string]any{tc.member: blank}
			arguments, provenance, err := sentInvocation(t, tc.r)
			if err != nil {
				t.Fatalf("%s, %q: %v", tc.name, blank, err)
			}
			if _, present := arguments[tc.member]; present || provenance != nil {
				t.Errorf("%s, %q: arguments %v, provenance %s: the server should see no actor, which it refuses", tc.name, blank, arguments, provenance)
			}
		}
		for _, r := range []Request{
			{Op: OpDeleteIssues, Method: http.MethodPost, Path: PathIssues, Body: map[string]any{"actor": blank, "ids": []string{"a"}}},
			{Op: OpSweepIssues, Method: http.MethodPost, Path: PathIssues, Body: map[string]any{"actor": blank, "tier": "durable", "pattern": "*"}},
		} {
			arguments, _, err := sentInvocation(t, r)
			if !errors.Is(err, issueops.ErrValidation) || arguments != nil {
				t.Errorf("%s with the actor %q: error %v, sent %v: a blank must not become a run with no actor", r.Op, blank, err, arguments)
			}
		}
	}
	if _, _, err := sentInvocation(t, Request{Op: OpClaimIssue, Method: http.MethodPost, Path: "/x", PathArg: "a", Body: map[string]any{"actor": 7}}); !errors.Is(err, issueops.ErrValidation) {
		t.Errorf("an actor that is no string: %v", err)
	}
}

// TestAnOmissionGoesOutAsV0sOwnValue is the trap in the service defaults. A request
// that omits a parameter goes out saying what v0 would have done with its absence, and
// one that names a value is left alone, so no native request ever reaches a default.
func TestAnOmissionGoesOutAsV0sOwnValue(t *testing.T) {
	for _, tc := range []struct {
		op      string
		pathArg string
		query   url.Values
		body    map[string]any
		want    map[string]string
	}{
		{op: OpListIssues, want: map[string]string{"sort": `"created"`}},
		{op: OpListIssues, query: url.Values{"sort": {"priority"}}, want: map[string]string{"sort": `"priority"`}},
		{op: OpListRelatedIssues, pathArg: "a", want: map[string]string{"direction": `""`}},
		{op: OpListRelatedIssues, pathArg: "a", query: url.Values{"direction": {"in"}}, want: map[string]string{"direction": `"in"`}},
		{op: OpCloseIssue, pathArg: "a", body: map[string]any{}, want: map[string]string{"reason": `""`}},
		{op: OpCloseIssue, pathArg: "a", body: map[string]any{"reason": "done"}, want: map[string]string{"reason": `"done"`}},
		{op: OpCreateIssue, body: map[string]any{"title": "t"}, want: map[string]string{"issue_type": `""`}},
		{op: OpCreateIssue, body: map[string]any{"title": "t", "issue_type": "bug"}, want: map[string]string{"issue_type": `"bug"`}},
		{op: OpSweepIssues, body: map[string]any{}, want: map[string]string{"tier": `""`, "dry_run": `false`, "protect_referenced": `true`, "protect_live_dependents": `false`}},
		{op: OpSweepIssues, body: map[string]any{"tier": "durable", "dry_run": true, "protect_referenced": false, "protect_live_dependents": true},
			want: map[string]string{"tier": `"durable"`, "dry_run": `true`, "protect_referenced": `false`, "protect_live_dependents": `true`}},
		{op: OpDeleteIssues, body: map[string]any{"ids": []string{"a"}}, want: map[string]string{"dry_run": `false`}},
		{op: OpDeleteIssues, body: map[string]any{"ids": []string{"a"}, "dry_run": true}, want: map[string]string{"dry_run": `true`}},
		{op: OpReclaimIssues, body: map[string]any{}, want: map[string]string{"older_than_seconds": `0`}},
		{op: OpReclaimIssues, body: map[string]any{"older_than_seconds": 60}, want: map[string]string{"older_than_seconds": `60`}},
	} {
		r := Request{Op: tc.op, Method: http.MethodPost, Path: "/x", PathArg: tc.pathArg, Query: tc.query}
		if tc.body != nil {
			r.Body = tc.body
		}
		arguments, _, err := sentInvocation(t, r)
		if err != nil {
			t.Fatalf("%s: %v", tc.op, err)
		}
		for name, want := range tc.want {
			if got := string(arguments[name]); got != want {
				t.Errorf("%s %v: %s = %s, want %s", tc.op, tc.body, name, got, want)
			}
		}
	}
}

// TestEveryDefaultHasAV0Spelling keeps the table above the seam honest as the server's
// grows: a default with no row here is a native request that would silently change.
func TestEveryDefaultHasAV0Spelling(t *testing.T) {
	covered := 0
	for name, row := range httpapi.CallableRoutes() {
		for param := range row.Defaults {
			covered++
			if _, ok := v0Omissions[row.Op][param]; !ok {
				t.Errorf("%s defaults %s, and v0Omissions has no value for %s's absence", name, param, row.Op)
			}
		}
		if actorlessOps[row.Op] != row.ActorOptional {
			t.Errorf("%s: actorlessOps says %v, the server %v", name, actorlessOps[row.Op], row.ActorOptional)
		}
	}
	if covered == 0 {
		t.Fatal("the server declares no defaults, so this asserted nothing")
	}
	for op, spelled := range v0Omissions {
		for param := range spelled {
			found := false
			for _, row := range httpapi.CallableRoutes() {
				if _, has := row.Defaults[param]; has && row.Op == op {
					found = true
				}
			}
			if !found {
				t.Errorf("v0Omissions spells %s on %s, and no Callable defaults it", param, op)
			}
		}
	}
}

// TestTheClientCallableMapMatchesTheServerTable is the seam the client's table and the
// server's would otherwise drift across: every row on the server is reachable from an
// operation here under the same name and path argument, and the test double's routes
// are the server's.
func TestTheClientCallableMapMatchesTheServerTable(t *testing.T) {
	server := httpapi.CallableRoutes()
	if len(wiretest.Routes) != len(server) {
		t.Errorf("wiretest.Routes has %d rows, the server %d", len(wiretest.Routes), len(server))
	}
	seen := map[string]bool{}
	for op, call := range callables {
		seen[call.name] = true
		row, ok := server[call.name]
		if !ok {
			t.Errorf("%s: the server has no Callable %s", op, call.name)
			continue
		}
		if row.Op != op {
			t.Errorf("%s routes to %s on the server, not %s", call.name, row.Op, op)
		}
		if got := strings.Join(row.PathParams, ","); got != call.pathName {
			t.Errorf("%s: path parameters %q on the server, %q here", call.name, got, call.pathName)
		}
	}
	for name, row := range server {
		if name == "issue/prune" {
			continue // Reached through sweepIssues' tier; see TestEveryRequestIsAnInvocation.
		}
		if !seen[name] {
			t.Errorf("no operation here reaches %s", name)
		}
		double, ok := wiretest.Routes[name]
		if !ok {
			t.Errorf("wiretest has no route for %s", name)
			continue
		}
		if double.Method != row.Method || double.Path != row.Path {
			t.Errorf("%s: wiretest says %s %s, the server %s %s", name, double.Method, double.Path, row.Method, row.Path)
		}
		if len(row.QueryParams) > 0 && (len(row.BodyParams) > 0 || row.ActorMember != "") {
			if want := append(slices.Clone(row.BodyParams), row.ActorMember); !slices.Equal(double.Body, want) {
				t.Errorf("%s: wiretest's body members %v, the server's %v", name, double.Body, want)
			}
		}
		if double.Actor != "" && double.Actor != row.ActorMember || double.Actor == "" && row.ActorMember == "author" {
			t.Errorf("%s: wiretest sends the actor as %q, the server as %q", name, double.Actor, row.ActorMember)
		}
	}
}
