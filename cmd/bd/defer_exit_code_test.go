package main

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/issueops"
)

// deferExitStubStore fails every write to one id, which no real store does on
// demand. The embedded nil DoltStorage makes any other call panic, so a test that
// wanders off the direct route fails loudly.
type deferExitStubStore struct {
	storage.DoltStorage
	issues  map[string]*types.Issue
	failID  string
	written []string
}

func (s *deferExitStubStore) SearchIssues(_ context.Context, _ string, filter types.IssueFilter) ([]*types.Issue, error) {
	var out []*types.Issue
	for _, id := range filter.IDs {
		if issue, ok := s.issues[id]; ok {
			out = append(out, issue)
		}
	}
	return out, nil
}

func (s *deferExitStubStore) GetIssue(_ context.Context, id string) (*types.Issue, error) {
	return s.issues[id], nil
}

func (s *deferExitStubStore) UpdateIssue(_ context.Context, id string, _ map[string]interface{}, _ string) error {
	if id == s.failID {
		return errors.New("the write failed")
	}
	s.written = append(s.written, id)
	return nil
}

// IssueLifecycle answers for the same write, so this pin holds whichever way the
// commands reach the store: row B8, on its own branch, moves them onto the role.
func (s *deferExitStubStore) IssueLifecycle() (issueops.Lifecycle, error) {
	return deferExitLifecycle{store: s}, nil
}

type deferExitLifecycle struct {
	issueops.Lifecycle
	store *deferExitStubStore
}

func (l deferExitLifecycle) Update(ctx context.Context, req issueops.UpdateRequest) (issueops.UpdateResult, error) {
	if err := l.store.UpdateIssue(ctx, req.IssueID, nil, req.Actor); err != nil {
		return issueops.UpdateResult{}, err
	}
	return issueops.UpdateResult{Issue: l.store.issues[req.IssueID]}, nil
}

// TestDeferAndUndeferExitNonzeroOnAPerIDFailure pins row C1 on the direct route: a
// failure for one id reaches the exit code, and the other ids still run. Exit 0
// used to tell a script the issue was deferred when nothing was written.
func TestDeferAndUndeferExitNonzeroOnAPerIDFailure(t *testing.T) {
	for _, tc := range []struct {
		name   string
		run    func() error
		status types.Status
		says   string
	}{
		{
			name:   "defer",
			run:    func() error { return deferCmd.RunE(deferCmd, []string{"t-bad", "t-good"}) },
			status: types.StatusOpen,
			says:   "Error deferring t-bad: the write failed",
		},
		{
			name:   "undefer",
			run:    func() error { return undeferCmd.RunE(undeferCmd, []string{"t-bad", "t-good"}) },
			status: types.StatusDeferred,
			says:   "Error undeferring t-bad: the write failed",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stub := &deferExitStubStore{failID: "t-bad", issues: map[string]*types.Issue{
				"t-bad":  {ID: "t-bad", Status: tc.status},
				"t-good": {ID: "t-good", Status: tc.status},
			}}
			oldStore := store
			store = stub
			t.Cleanup(func() { store = oldStore })

			var err error
			stderr := captureStderrDuring(t, func() { err = tc.run() })

			var exit *exitError
			if !errors.As(err, &exit) || exit.Code != 1 {
				t.Errorf("err = %v, want a nonzero exit", err)
			}
			if !strings.Contains(stderr, tc.says) {
				t.Errorf("stderr = %q, want %q", stderr, tc.says)
			}
			if !slices.Equal(stub.written, []string{"t-good"}) {
				t.Errorf("wrote %v, want the other id to have run", stub.written)
			}
		})
	}
}
