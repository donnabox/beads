package main

import (
	"encoding/json"
	"testing"

	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
)

func assertGraphFamiliarJSON(t *testing.T, payload any) map[string]json.RawMessage {
	t.Helper()
	t.Setenv("BD_JSON_ENVELOPE", "")
	oldJSON := jsonOutput
	jsonOutput = true
	defer func() { jsonOutput = oldJSON }()
	out := captureStdout(t, func() error { return graphPrintFamiliar(payload, "human output") })
	var object map[string]json.RawMessage
	if err := json.Unmarshal([]byte(out), &object); err != nil {
		t.Fatalf("familiar JSON: %s: %v", out, err)
	}
	if _, ok := object["result"]; ok {
		t.Fatalf("graph envelope leaked into familiar command: %s", out)
	}
	if string(object["schema_version"]) != "1" {
		t.Fatalf("native schema version missing: %s", out)
	}
	return object
}

func TestGraphPreviewFamiliarStateStatusAndMergeSlotJSON(t *testing.T) {
	issue := graphstore.IssueRecord{ID: "https://example.invalid/scope/beads/local-id", Properties: &types.Issue{ID: "local-id", Labels: []string{"mode:active"}}}
	for _, list := range []bool{false, true} {
		payload, _, err := graphStateReport(issue, "mode", list)
		if err != nil {
			t.Fatal(err)
		}
		object := assertGraphFamiliarJSON(t, payload)
		if string(object["issue_id"]) != `"local-id"` {
			t.Fatalf("state reported canonical URL: %s", object["issue_id"])
		}
		if list {
			if _, ok := object["states"]; !ok {
				t.Fatal("state list missing states")
			}
		} else if string(object["value"]) != `"active"` {
			t.Fatalf("state value: %s", object["value"])
		}
	}
	status := assertGraphFamiliarJSON(t, &StatusOutput{Summary: &types.Statistics{TotalIssues: 2}})
	if _, ok := status["summary"]; !ok {
		t.Fatal("status summary missing")
	}
	if _, ok := status["blocked_count_skipped"]; ok {
		t.Fatal("status emitted non-native false skip flag")
	}
	for _, payload := range []map[string]any{
		{"id": "local-id", "status": "open"},
		{"id": "local-id", "available": true, "holder": nil, "waiters": []string{}},
		{"id": "local-id", "acquired": true, "holder": "worker"},
		{"id": "local-id", "released": true},
	} {
		object := assertGraphFamiliarJSON(t, payload)
		if string(object["id"]) != `"local-id"` {
			t.Fatalf("merge-slot local ID missing: %s", object["id"])
		}
	}
}

func TestGraphPreviewCoordinationActorMatchesDefaultClaim(t *testing.T) {
	oldActor := actor
	actor = ""
	t.Cleanup(func() { actor = oldActor })
	t.Setenv("BEADS_ACTOR", "")
	t.Setenv("BD_ACTOR", "")
	// Claim and the three coordination paths must resolve the same fallback.
	claimActor := getActorWithGit()
	if claimActor == "" || graphPreviewCoordinationActor() != claimActor {
		t.Fatalf("default claim actor %q does not match coordination actor", claimActor)
	}
	actor = "explicit-worker"
	if graphPreviewCoordinationActor() != "explicit-worker" {
		t.Fatal("explicit actor was ignored")
	}
	actor = ""
	t.Setenv("BEADS_ACTOR", "env-worker")
	if graphPreviewCoordinationActor() != "env-worker" {
		t.Fatal("BEADS_ACTOR fallback was ignored")
	}
}
