package httpclient

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/httpclient/wire"
	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/issueops"
)

// A Callable's service defaults are the native CLI's, and they are not v0's. This
// client was written against v0, and omits a parameter to mean what v0 means by its
// absence, so a default that reached one of its requests would change the answer with
// no error at all. These run each role down to the real bytes, and hold each request
// to saying what v0 would have done. The other half of the seam is v0's own reading of
// those values, which Part A of scripts/bdp-call-parity and the served tier hold.

// invocationSpy is a bd serve that records what each invocation really carried.
type invocationSpy struct {
	mu   sync.Mutex
	sent map[string]map[string]json.RawMessage
	raw  map[string]json.RawMessage
}

func (s *invocationSpy) store(t *testing.T, caps ...string) *Store {
	t.Helper()
	s.sent, s.raw = map[string]map[string]json.RawMessage{}, map[string]json.RawMessage{}
	// A current server's tokens for what these roles dial, so that none refuses first.
	caps = append(caps, "issues.list", wire.CapListSort, "issues.related", "issues.close", "issues.create",
		"issues.addComment", "issues.sweep", "issues.delete", "issues.reclaim")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/bdp/call/")
		var invocation struct {
			Arguments  map[string]json.RawMessage `json:"arguments"`
			Provenance json.RawMessage            `json:"revisionProvenance"`
		}
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &invocation); err != nil {
			t.Errorf("%s: %v: %s", name, err, body)
		}
		w.Header().Set("Content-Type", "application/vnd.bdp.v0.1+json")
		if name == "issue/context" {
			handshake := v0Context("proj-omissions")
			handshake.Capabilities = caps
			_ = json.NewEncoder(w).Encode(map[string]any{"result": handshake})
			return
		}
		s.mu.Lock()
		s.sent[name] = invocation.Arguments
		s.raw[name] = invocation.Provenance
		s.mu.Unlock()
		_, _ = io.WriteString(w, `{"changed":false,"result":{"items":[],"has_more":false}}`)
	}))
	t.Cleanup(srv.Close)
	base, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	client, err := wire.New(base, nil, wire.Options{HTTPClient: srv.Client()})
	if err != nil {
		t.Fatal(err)
	}
	return New(Target{BaseURL: base}, client, nil)
}

func TestNativeRequestsSayWhatV0WouldHaveDone(t *testing.T) {
	sweepCaps := []string{wire.CapSweepWispsPlane, wire.CapSweepLiveDependents, wire.CapSweepRetainedTypes}
	ptrInt := func(n int) *int { return &n }
	for _, tc := range []struct {
		name     string
		do       func(*testing.T, *Store) error
		callable string
		want     map[string]string
	}{
		{
			"a walk in the flagless order reads created pages, as v0 serves them",
			func(_ *testing.T, s *Store) error {
				reader, _ := s.IssueReader()
				_, err := reader.List(t.Context(), issueops.ListRequest{Limit: ptrInt(0)})
				return err
			},
			"issue/list", map[string]string{"sort": `"created"`},
		},
		{
			"a list that names created",
			func(_ *testing.T, s *Store) error {
				reader, _ := s.IssueReader()
				_, err := reader.List(t.Context(), issueops.ListRequest{SortBy: "created", Limit: ptrInt(2)})
				return err
			},
			"issue/list", map[string]string{"sort": `"created"`},
		},
		{
			"the neighbors of an issue, in",
			func(_ *testing.T, s *Store) error {
				relations, _ := s.IssueRelations()
				_, err := relations.Related(t.Context(), issueops.RelatedRequest{ID: "bd-1", Direction: issueops.RelationIn})
				return err
			},
			"issue/dep/list", map[string]string{"direction": `"in"`},
		},
		{
			"a close with no reason stores none",
			func(_ *testing.T, s *Store) error {
				lifecycle, _ := s.IssueLifecycle()
				_, err := lifecycle.Close(t.Context(), issueops.CloseRequest{Actor: "sam", IssueID: "bd-1"})
				return err
			},
			"issue/close", map[string]string{"reason": `""`},
		},
		{
			"the store's close, which bd duplicate and bd supersede call with no reason",
			func(_ *testing.T, s *Store) error {
				return s.CloseIssue(t.Context(), "bd-1", "", "sam", "")
			},
			"issue/close", map[string]string{"reason": `""`},
		},
		{
			"a create with no type stays v0's refusal, and is not a task",
			func(_ *testing.T, s *Store) error {
				lifecycle, _ := s.IssueLifecycle()
				_, err := lifecycle.Create(t.Context(), issueops.CreateRequest{Actor: "sam", Issue: &types.Issue{Title: "t"}})
				return err
			},
			"issue/create", map[string]string{"issue_type": `""`},
		},
		{
			"a purge with the protection off, which issue/purge defaults on",
			func(_ *testing.T, s *Store) error {
				sweeper, _ := s.Sweeper()
				_, err := sweeper.Sweep(t.Context(), issueops.SweepRequest{Tier: issueops.SweepEphemeral})
				return err
			},
			"issue/purge", map[string]string{"tier": `"ephemeral"`, "dry_run": `false`, "protect_referenced": `false`, "protect_live_dependents": `false`},
		},
		{
			"a prune that runs for real, which issue/prune defaults to preview",
			func(_ *testing.T, s *Store) error {
				sweeper, _ := s.Sweeper()
				_, err := sweeper.Sweep(t.Context(), issueops.SweepRequest{Tier: issueops.SweepDurable, IDPattern: "*"})
				return err
			},
			"issue/prune", map[string]string{"tier": `"durable"`, "dry_run": `false`, "protect_referenced": `false`, "protect_live_dependents": `false`},
		},
		{
			"a wisps-plane sweep, which is a purge",
			func(_ *testing.T, s *Store) error {
				sweeper, _ := s.Sweeper()
				_, err := sweeper.Sweep(t.Context(), issueops.SweepRequest{Tier: issueops.SweepWispsPlane, ProtectReferenced: true, ProtectLiveDependents: true, DryRun: true})
				return err
			},
			"issue/purge", map[string]string{"tier": `"wisps-plane"`, "dry_run": `true`, "protect_referenced": `true`, "protect_live_dependents": `true`},
		},
		{
			"a delete that runs for real",
			func(_ *testing.T, s *Store) error {
				deleter, _ := s.Deleter()
				_, err := deleter.Delete(t.Context(), issueops.DeleteRequest{IDs: []string{"bd-1"}})
				return err
			},
			"issue/delete", map[string]string{"dry_run": `false`},
		},
		{
			"a reclaim with no grace window waits for nothing, as bd reclaim --older-than 0s means",
			func(_ *testing.T, s *Store) error {
				reclaimer, _ := s.LeaseReclaimer()
				_, err := reclaimer.Reclaim(t.Context(), issueops.ReclaimRequest{Actor: "sam"})
				return err
			},
			"issue/reclaim", map[string]string{"older_than_seconds": `0`},
		},
		{
			"a reclaim that names its window",
			func(_ *testing.T, s *Store) error {
				reclaimer, _ := s.LeaseReclaimer()
				_, err := reclaimer.Reclaim(t.Context(), issueops.ReclaimRequest{Actor: "sam", OlderThan: 5 * time.Second})
				return err
			},
			"issue/reclaim", map[string]string{"older_than_seconds": `5`},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			spy := &invocationSpy{}
			store := spy.store(t, sweepCaps...)
			// What the role makes of the canned answer is not the question here. That it
			// went out is, and a refusal before the dial shows as nothing recorded.
			err := tc.do(t, store)
			spy.mu.Lock()
			defer spy.mu.Unlock()
			arguments, sent := spy.sent[tc.callable]
			if !sent {
				t.Fatalf("no %s went out (role said %v); sent %v", tc.callable, err, spy.sent)
			}
			for name, want := range tc.want {
				if got, present := arguments[name]; !present || string(got) != want {
					t.Errorf("%s = %s (present %v), want %s", name, got, present, want)
				}
			}
			for _, member := range []string{"actor", "author"} {
				if _, present := arguments[member]; present {
					t.Errorf("`%s` went out in arguments: %v", member, arguments)
				}
			}
		})
	}
}

// TestTheRoleActorGoesOutAsProvenance: the same name, in the only place a Callable
// reads one, for the two v0 spellings and for a role that may send none.
func TestTheRoleActorGoesOutAsProvenance(t *testing.T) {
	spy := &invocationSpy{}
	store := spy.store(t, wire.CapSweepRetainedTypes)
	provenance := func(callable string) string {
		spy.mu.Lock()
		defer spy.mu.Unlock()
		if _, sent := spy.sent[callable]; !sent {
			t.Fatalf("no %s went out", callable)
		}
		return string(spy.raw[callable])
	}
	const sam = `{"actor":{"basis":"writer-supplied","id":"sam"}}`

	commenter, _ := store.Commenter()
	if _, err := commenter.AddComment(t.Context(), issueops.AddCommentRequest{IssueID: "bd-1", Author: "sam", Text: "hi"}); err == nil {
		t.Log("the canned answer read as a comment")
	}
	if got := provenance("issue/comment"); got != sam {
		t.Errorf("issue/comment: %s, want %s", got, sam)
	}

	reclaimer, _ := store.LeaseReclaimer()
	_, _ = reclaimer.Reclaim(t.Context(), issueops.ReclaimRequest{Actor: "sam"})
	if got := provenance("issue/reclaim"); got != sam {
		t.Errorf("issue/reclaim: %s, want %s", got, sam)
	}

	deleter, _ := store.Deleter()
	if _, err := deleter.Delete(t.Context(), issueops.DeleteRequest{IDs: []string{"bd-1"}, Actor: "  "}); errors.Is(err, issueops.ErrValidation) {
		t.Errorf("the role omits a blank actor on a delete rather than refusing it: %v", err)
	}
	if got := provenance("issue/delete"); got != "" {
		t.Errorf("a delete with a blank actor sent provenance %s", got)
	}
}
