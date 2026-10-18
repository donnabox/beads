package main

import (
	"encoding/json"
	"fmt"
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
	object := assertGraphFamiliarJSON(t, result)
	if string(object["id"]) != `"real-id"` || string(object["status"]) != `"heartbeat"` || string(object["owner"]) != `"holder"` {
		t.Fatalf("heartbeat JSON projected wrong fields: %+v", object)
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
			object := assertGraphFamiliarJSON(t, result)
			if string(object["scoped"]) != fmt.Sprint(tc.scoped) || string(object["count"]) != "0" || string(object["reclaimed"]) != "[]" {
				t.Fatalf("wrong reclaim JSON scope: %+v", object)
			}
		})
	}
}

func TestGraphPreviewReclaimReportsNativeIssueID(t *testing.T) {
	result, _, err := graphReclaimReport([]types.ReclaimedLease{{ID: "real-id", PreviousOwner: "holder"}}, true, "reclaimed")
	if err != nil {
		t.Fatal(err)
	}
	object := assertGraphFamiliarJSON(t, result)
	var reclaimed []types.ReclaimedLease
	if err := json.Unmarshal(object["reclaimed"], &reclaimed); err != nil {
		t.Fatal(err)
	}
	if len(reclaimed) != 1 || reclaimed[0].ID != "real-id" || reclaimed[0].PreviousOwner != "holder" {
		t.Fatalf("reclaim local ID changed: %+v", reclaimed)
	}
}
