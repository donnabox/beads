//go:build cgo

package graphstore

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/types"
	publicops "github.com/steveyegge/beads/issueops"
)

// A fault after the first catalog row is stronger than rejecting input up
// front: the native batch writer and the first mapping have both run, yet the
// transaction must restore every Issue, retained version and coordination row.
func TestGraphIssueCreateBatchAtomicMappingAndInitialVersions(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			ctx, options := issueExperimentOptions(t, backend)
			s, err := OpenExisting(ctx, options)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := s.Close(); err != nil {
					t.Error(err)
				}
			})
			request := publicops.CreateBatchRequest{Actor: "test-author", Items: []publicops.BatchCreateItem{
				{Issue: plainIssue("First heading").Issue},
				{Issue: plainIssue("Second heading").Issue},
			}}
			before := workflowState(t, ctx, s)
			fault := errors.New("after first Issue catalog mapping")
			catalogWrites := 0
			s.afterWrite = func(stage string) error {
				if stage == "issue-catalog" {
					catalogWrites++
					if catalogWrites == 1 {
						return fault
					}
				}
				return nil
			}
			failed, err := s.CreateIssueBatch(ctx, request)
			s.afterWrite = nil
			if !errors.Is(err, fault) || failed != nil || catalogWrites != 1 {
				t.Fatalf("first-mapping fault: records=%+v err=%v hook calls=%d", failed, err, catalogWrites)
			}
			if after := workflowState(t, ctx, s); !reflect.DeepEqual(before, after) {
				t.Fatal("failed batch left native Issue, History, catalog, retained version, or coordination state")
			}
			for _, table := range []string{"issues", "issue_versions", "graph_preview_catalog", "graph_preview_issue_versions"} {
				if got := graphBatchRowCount(t, ctx, s, table); got != 0 {
					t.Fatalf("failed batch left %d rows in %s", got, table)
				}
			}

			records, err := s.CreateIssueBatch(ctx, request)
			if err != nil || len(records) != 2 {
				t.Fatalf("successful batch: records=%+v err=%v", records, err)
			}
			for _, record := range records {
				path := strings.TrimPrefix(record.ID, options.Binding.ScopeURL)
				if path == record.ID {
					t.Fatalf("batch ID is outside workspace Scope: %s", record.ID)
				}
				var nativeCount, mappingCount, ordinal, current int
				if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM issue_versions WHERE issue_id=?`, record.Properties.ID).Scan(&nativeCount); err != nil {
					t.Fatal(err)
				}
				if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM graph_preview_issue_versions WHERE path=?`, path).Scan(&mappingCount); err != nil {
					t.Fatal(err)
				}
				if err := s.db.QueryRowContext(ctx, `SELECT current_revision FROM issues WHERE id=?`, record.Properties.ID).Scan(&current); err != nil {
					t.Fatal(err)
				}
				if err := s.db.QueryRowContext(ctx, `SELECT issue_revision FROM graph_preview_issue_versions WHERE path=?`, path).Scan(&ordinal); err != nil {
					t.Fatal(err)
				}
				if nativeCount != 1 || mappingCount != 1 || ordinal != current {
					t.Fatalf("Issue %s has native versions=%d mappings=%d mapped ordinal=%d current=%d", record.ID, nativeCount, mappingCount, ordinal, current)
				}
			}
			if got := graphBatchRowCount(t, ctx, s, "issue_versions"); got != 2 {
				t.Fatalf("successful batch wrote %d native initial versions, want 2", got)
			}
		})
	}
}

func graphBatchRowCount(t *testing.T, ctx context.Context, s *Store, table string) int {
	t.Helper()
	var count int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}

func TestGraphIssueCreateBatchMapsLocalDependenciesAtomically(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			ctx, options := issueExperimentOptions(t, backend)
			s, err := OpenExisting(ctx, options)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = s.Close() })
			prereq, err := s.CreateIssue(ctx, "beads/prereq", plainIssue("Prerequisite"))
			if err != nil {
				t.Fatal(err)
			}
			consumer, err := s.CreateIssue(ctx, "beads/consumer", plainIssue("Consumer"))
			if err != nil {
				t.Fatal(err)
			}
			request := publicops.CreateBatchRequest{Actor: "test-author", Items: []publicops.BatchCreateItem{
				{Issue: plainIssue("First heading").Issue},
				{Issue: plainIssue("Second heading").Issue, Dependencies: []publicops.CreateDependency{
					{TargetID: "beads/prereq", Type: types.DepBlocks},
					{TargetID: consumer.Properties.ID, Type: types.DepBlocks, Reverse: true},
				}},
			}}
			before := workflowState(t, ctx, s)
			fault := errors.New("dependency Link retention failed")
			s.afterWrite = func(stage string) error {
				if stage == "create-dependency-retained" {
					return fault
				}
				return nil
			}
			_, err = s.CreateIssueBatch(ctx, request)
			s.afterWrite = nil
			if !errors.Is(err, fault) || !reflect.DeepEqual(before, workflowState(t, ctx, s)) {
				t.Fatalf("failed batch left native or graph state: %v", err)
			}
			records, err := s.CreateIssueBatch(ctx, request)
			if err != nil || len(records) != 2 {
				t.Fatalf("batch records=%+v err=%v", records, err)
			}
			updatedConsumer, err := s.ShowIssue(ctx, "beads/consumer")
			if err != nil {
				t.Fatal(err)
			}
			if len(records[0].Owned) != 0 || len(records[1].Owned) != 1 || len(updatedConsumer.Owned) != 1 {
				t.Fatalf("batch ownership first=%d second=%d consumer=%d", len(records[0].Owned), len(records[1].Owned), len(updatedConsumer.Owned))
			}
			for _, check := range []struct {
				owned  json.RawMessage
				source string
				target string
			}{
				{records[1].Owned[0], records[1].ID, prereq.ID},
				{updatedConsumer.Owned[0], consumer.ID, records[1].ID},
			} {
				var link LinkRecord
				if err := json.Unmarshal(check.owned, &link); err != nil {
					t.Fatal(err)
				}
				if link.Source != check.source || link.Target != check.target || link.Properties["dep_type"] != string(types.DepBlocks) {
					t.Fatalf("incorrect canonical dependency Link: %+v", link)
				}
			}
			before = workflowState(t, ctx, s)
			request.Items[1].Dependencies[0].TargetID = "missing-target"
			_, err = s.CreateIssueBatch(ctx, request)
			if !errors.Is(err, ErrNotFound) || !reflect.DeepEqual(before, workflowState(t, ctx, s)) {
				t.Fatalf("missing dependency target left partial batch: %v", err)
			}
		})
	}
}

func TestGraphIssueCreateBatchCanLinkEarlierExplicitItem(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			ctx, options := issueExperimentOptions(t, backend)
			s, err := OpenExisting(ctx, options)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = s.Close() })
			first := plainIssue("Prerequisite").Issue
			first.ID = "exp-prereq"
			request := publicops.CreateBatchRequest{Actor: "test-author", Items: []publicops.BatchCreateItem{
				{Issue: first},
				{Issue: plainIssue("Dependent").Issue, Dependencies: []publicops.CreateDependency{{TargetID: first.ID, Type: types.DepBlocks}}},
			}}
			before := workflowState(t, ctx, s)
			fault := errors.New("retaining same-batch Link failed")
			s.afterWrite = func(stage string) error {
				if stage == "create-dependency-retained" {
					return fault
				}
				return nil
			}
			_, err = s.CreateIssueBatch(ctx, request)
			s.afterWrite = nil
			if !errors.Is(err, fault) || !reflect.DeepEqual(before, workflowState(t, ctx, s)) {
				t.Fatalf("same-batch Link fault left native or graph state: %v", err)
			}
			records, err := s.CreateIssueBatch(ctx, request)
			if err != nil || len(records) != 2 {
				t.Fatalf("same-batch records=%+v err=%v", records, err)
			}
			if records[0].Properties.ID != first.ID || len(records[1].Owned) != 1 {
				t.Fatalf("same-batch Issue identity/owned Links: %+v", records)
			}
			var link LinkRecord
			if err := json.Unmarshal(records[1].Owned[0], &link); err != nil {
				t.Fatal(err)
			}
			if link.Source != records[1].ID || link.Target != records[0].ID || link.Properties["dep_type"] != string(types.DepBlocks) {
				t.Fatalf("same-batch Link = %+v", link)
			}
			before = workflowState(t, ctx, s)
			request.Items[0].Issue.ID = "exp-next"
			request.Items[1].Dependencies = []publicops.CreateDependency{{TargetID: "exp-next", Type: types.DepBlocks, Reverse: true}}
			_, err = s.CreateIssueBatch(ctx, request)
			if !errors.Is(err, ErrCapabilityUnavailable) || !reflect.DeepEqual(before, workflowState(t, ctx, s)) {
				t.Fatalf("reverse same-batch edge was not an atomic explicit refusal: %v", err)
			}
		})
	}
}
