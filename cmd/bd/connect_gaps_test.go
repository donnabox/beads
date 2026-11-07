package main

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/types"
)

// gapStubStore answers the reads rows D3, D4 and D6 make and fails the one a test
// names. The embedded nil DoltStorage makes any other call panic, so a test that
// wanders off the path it pins fails loudly.
type gapStubStore struct {
	storage.DoltStorage
	failing string
	remote  bool
}

var errGapRead = errors.New("the read failed")

func (s *gapStubStore) fail(read string) error {
	if s.failing == read {
		return errGapRead
	}
	return nil
}

func (s *gapStubStore) IsRemoteBackendStore() bool { return s.remote }

func (s *gapStubStore) GetIssue(_ context.Context, id string) (*types.Issue, error) {
	return &types.Issue{ID: id, Title: id, Status: types.StatusOpen}, nil
}

func (s *gapStubStore) GetDependents(context.Context, string) ([]*types.Issue, error) {
	return nil, s.fail("GetDependents")
}

func (s *gapStubStore) GetDependencies(context.Context, string) ([]*types.Issue, error) {
	return nil, s.fail("GetDependencies")
}

func (s *gapStubStore) GetDependencyRecords(context.Context, string) ([]*types.Dependency, error) {
	return nil, s.fail("GetDependencyRecords")
}

func (s *gapStubStore) GetDependencyCounts(context.Context, []string) (map[string]*types.DependencyCounts, error) {
	return nil, s.fail("GetDependencyCounts")
}

func (s *gapStubStore) GetDependencyRecordsForIssues(context.Context, []string) (map[string][]*types.Dependency, error) {
	return nil, s.fail("GetDependencyRecordsForIssues")
}

func (s *gapStubStore) DetectCycles(context.Context) ([][]*types.Issue, error) {
	return nil, s.fail("DetectCycles")
}

func (s *gapStubStore) GetIssuesByIDs(context.Context, []string) ([]*types.Issue, error) {
	return nil, s.fail("GetIssuesByIDs")
}

// TestGraphAFailedReadIsTheCommandsFailure pins row D3, on every route that loads a
// subgraph this way: a read that fails ended the walk for that node, so bd graph
// exited 0 with only the root and its own edges, as if that were the graph.
func TestGraphAFailedReadIsTheCommandsFailure(t *testing.T) {
	for _, read := range []string{"GetDependents", "GetDependencies", "GetDependencyRecords"} {
		t.Run(read, func(t *testing.T) {
			_, err := loadGraphSubgraph(context.Background(), &gapStubStore{failing: read}, "t-root")
			if !errors.Is(err, errGapRead) {
				t.Fatalf("err = %v, want the failed read", err)
			}
		})
	}
	if _, err := loadGraphSubgraph(context.Background(), &gapStubStore{}, "t-root"); err != nil {
		t.Fatalf("with every read answering: %v", err)
	}
}

// TestReadyExplainAFailedReadIsTheCommandsFailure pins row D4, on every route: a
// read that fails went to a debug log, so bd ready --explain exited 0 with blank
// blockers and zero counts, as if those were the answer.
func TestReadyExplainAFailedReadIsTheCommandsFailure(t *testing.T) {
	origJSONOutput := jsonOutput
	jsonOutput = false
	t.Cleanup(func() { jsonOutput = origJSONOutput })
	ready := []*types.Issue{{ID: "t-ready", Dependencies: []*types.Dependency{{IssueID: "t-ready", DependsOnID: "t-other"}}}}

	for read, says := range map[string]string{
		"GetDependencyCounts":           "failed to get dependency counts",
		"GetDependencyRecordsForIssues": "failed to get dependency records",
		"DetectCycles":                  "failed to detect cycles",
		"GetIssuesByIDs":                "failed to get blocker issues",
	} {
		t.Run(read, func(t *testing.T) {
			var err error
			stderr := captureStderrDuring(t, func() {
				_, _, _, _, err = readReadyExplanationDetail(context.Background(), &gapStubStore{failing: read}, ready, nil)
			})
			if err == nil {
				t.Fatal("the read failed and the explanation was built")
			}
			if !strings.Contains(stderr+err.Error(), says) {
				t.Errorf("stderr = %q, err = %v, want %q", stderr, err, says)
			}
		})
	}

	if _, _, _, _, err := readReadyExplanationDetail(context.Background(), &gapStubStore{}, ready, nil); err != nil {
		t.Fatalf("with every read answering: %v", err)
	}
	if ready[0].Dependencies != nil {
		t.Error("the row's own edges stayed: the explanation reports edges from its own read")
	}
}
