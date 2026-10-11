package main

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
)

func TestGraphPreviewHeartbeatReportsNativeIssueID(t *testing.T) {
	issue := graphstore.IssueRecord{ID: "https://example.invalid/scope/beads/real-id", Properties: &types.Issue{ID: "real-id"}}
	result, human, err := graphHeartbeatReport(issue, "holder")
	if err != nil || human != "Heartbeat real-id (lease refreshed)" {
		t.Fatalf("heartbeat report: %q %v", human, err)
	}
	var out bytes.Buffer
	if err := graphPrintTo(&out, result, human, false, true); err != nil {
		t.Fatal(err)
	}
	var envelope struct {
		Result struct{ ID, Status, Owner string }
	}
	if err := json.Unmarshal(out.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Result.ID != "real-id" || envelope.Result.Status != "heartbeat" || envelope.Result.Owner != "holder" {
		t.Fatalf("heartbeat JSON projected canonical URL: %s", out.String())
	}
}

func TestGraphPreviewReclaimReportKeepsOperatorScope(t *testing.T) {
	for _, tc := range []struct {
		name   string
		filter types.ReclaimFilter
		scoped bool
	}{
		{"bare", types.ReclaimFilter{}, false},
		{"any-replica", types.ReclaimFilter{AnyReplica: true}, false},
		{"id", types.ReclaimFilter{IDs: []string{"real-id"}}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// The graph storage call adds eligible native IDs for its own fence.
			// Its report must retain the selection made before that expansion.
			scoped := !tc.filter.IsEmpty()
			tc.filter.IDs = []string{"mapped-native-id"}
			result, human, err := graphReclaimReport([]types.ReclaimedLease{}, scoped, "No stale leases to reclaim")
			if err != nil || human != "No stale leases to reclaim" {
				t.Fatalf("reclaim report: %q %v", human, err)
			}
			var out bytes.Buffer
			if err := graphPrintTo(&out, result, human, false, true); err != nil {
				t.Fatal(err)
			}
			var envelope struct {
				Result struct {
					Scoped    bool
					Count     int
					Reclaimed []types.ReclaimedLease
				}
			}
			if err := json.Unmarshal(out.Bytes(), &envelope); err != nil {
				t.Fatal(err)
			}
			if envelope.Result.Scoped != tc.scoped || envelope.Result.Count != 0 || envelope.Result.Reclaimed == nil {
				t.Fatalf("wrong reclaim JSON scope: %s", out.String())
			}
		})
	}
}
