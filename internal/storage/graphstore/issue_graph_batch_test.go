//go:build cgo

package graphstore

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/types"
	publicops "github.com/steveyegge/beads/issueops"
)

func TestApplyNewIssueGraphBindsForwardKeyAndRollsBackAfterNativeEdge(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			ctx, options := issueExperimentOptions(t, backend)
			s, err := OpenExisting(ctx, options)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = s.Close() })
			request := publicops.ApplyBatchRequest{Actor: "test-author", Items: []publicops.ApplyItem{
				{Kind: publicops.ItemCreate, Create: &publicops.CreateItem{Key: "first", Issue: plainIssue("First").Issue}},
				{Kind: publicops.ItemCreate, Create: &publicops.CreateItem{Key: "second", Issue: plainIssue("Second").Issue}},
				{Kind: publicops.ItemDepAdd, DepAdd: &publicops.DepAddItem{
					Source: publicops.Ref{Key: "first"}, Target: publicops.Ref{Key: "second"}, Type: types.DepBlocks,
				}},
			}}
			before := workflowState(t, ctx, s)
			fault := errors.New("retaining graph batch Link failed")
			s.afterWrite = func(stage string) error {
				if stage == "create-dependency-retained" {
					return fault
				}
				return nil
			}
			failed, err := s.ApplyNewIssueGraph(ctx, request)
			s.afterWrite = nil
			if !errors.Is(err, fault) || failed != nil {
				t.Fatalf("Link-retention fault: result=%v err=%v", failed, err)
			}
			if after := workflowState(t, ctx, s); !reflect.DeepEqual(before, after) {
				t.Fatal("failed graph batch left native Issue, edge, catalog, retained version, History or coordination state")
			}
			for _, table := range []string{"issues", "dependencies", "issue_versions", "graph_preview_catalog", "graph_preview_issue_versions"} {
				if got := graphBatchRowCount(t, ctx, s, table); got != 0 {
					t.Fatalf("failed graph batch left %d rows in %s", got, table)
				}
			}

			records, err := s.ApplyNewIssueGraph(ctx, request)
			if err != nil {
				t.Fatal(err)
			}
			first, second := records["first"], records["second"]
			if first.Properties.ID == "" || second.Properties.ID == "" || first.Properties.ID == second.Properties.ID {
				t.Fatalf("symbolic IDs did not bind to two Issues: %+v", records)
			}
			if len(first.Owned) != 1 || len(second.Owned) != 0 {
				t.Fatalf("graph-owned Links: first=%d second=%d", len(first.Owned), len(second.Owned))
			}
			var link LinkRecord
			if err := json.Unmarshal(first.Owned[0], &link); err != nil {
				t.Fatal(err)
			}
			if link.Source != first.ID || link.Target != second.ID || link.Properties["dep_type"] != string(types.DepBlocks) {
				t.Fatalf("symbolic edge mapped incorrectly: %+v", link)
			}
			if got := graphBatchRowCount(t, ctx, s, "dependencies"); got != 1 {
				t.Fatalf("native edges=%d, want 1", got)
			}
			for _, record := range records {
				path := strings.TrimPrefix(record.ID, options.Binding.ScopeURL)
				var retained int
				if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM graph_preview_issue_versions WHERE path=?`, path).Scan(&retained); err != nil {
					t.Fatal(err)
				}
				if retained != 1 {
					t.Fatalf("Issue %s retained versions=%d, want 1", record.ID, retained)
				}
			}

			before = workflowState(t, ctx, s)
			request.Items[2].DepAdd.Target = publicops.Ref{ID: second.Properties.ID}
			refused, err := s.ApplyNewIssueGraph(ctx, request)
			if !errors.Is(err, ErrCapabilityUnavailable) || refused != nil || !reflect.DeepEqual(before, workflowState(t, ctx, s)) {
				t.Fatalf("existing-ID edge was not a no-write refusal: result=%v err=%v", refused, err)
			}
		})
	}
}
