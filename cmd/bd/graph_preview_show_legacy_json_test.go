//go:build cgo

package main

import (
	"encoding/json"
	"testing"

	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
)

func TestGraphPreviewShowLegacyIssueJSON(t *testing.T) {
	bd := buildBDUnderTest(t)
	work, home := t.TempDir(), t.TempDir()
	graphPolicyCLI(t, bd, work, home, nil, "", "init", "--graph-mode", "link", "--prefix", "show", "--scope-url", "https://example.invalid/show/", "--skip-hooks", "--skip-agents", "--non-interactive", "--json")
	created := graphMixedResult[graphstore.IssueRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", "create", "Visible issue", "--id", "beads/work", "--description", "Body", "--labels", "one,two", "--metadata", `{"review":true}`, "--json"))
	prerequisite := graphMixedResult[graphstore.IssueRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", "create", "Prerequisite", "--id", "beads/needs", "--json"))
	// A native backing ID can also name a different canonical Bead. No ID in
	// the Issue detail response may accidentally route to that Memory.
	for _, nativeID := range []string{created.Properties.ID, prerequisite.Properties.ID} {
		graphPolicyCLI(t, bd, work, home, nil, "", "remember", "Colliding Memory", "--id", "beads/"+nativeID, "--json")
		memory := graphMixedResult[graphstore.Record](t, graphPolicyCLI(t, bd, work, home, nil, "", "show", "beads/"+nativeID, "--json"))
		if memory.Properties.Title == "" || memory.ID != "https://example.invalid/show/beads/"+nativeID {
			t.Fatalf("native backing ID did not select colliding Memory: %+v", memory)
		}
	}
	graphPolicyCLI(t, bd, work, home, nil, "", "dep", "add", "beads/work", "beads/needs", "--json")
	current := graphMixedResult[graphstore.IssueRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", "show", "beads/work", "--format", "graph-json", "--json"))

	shown := graphPolicyCLI(t, bd, work, home, nil, "", "show", "beads/work", "--json")
	var details []types.IssueDetails
	if err := json.Unmarshal([]byte(shown), &details); err != nil || len(details) != 1 {
		t.Fatalf("graph Issue show must use ordinary flat detail array: %v\n%s", err, shown)
	}
	got := details[0]
	if got.ID != "beads/work" || got.Title != "Visible issue" || got.Description != "Body" || got.Revision != current.Revision {
		t.Fatalf("legacy Issue identity, content or graph guard token changed: %+v", got)
	}
	if got.DependencyCount == nil || *got.DependencyCount != 1 || len(got.Dependencies) != 1 || got.Dependencies[0].ID != "beads/needs" || got.Dependencies[0].Title != "Prerequisite" || got.DependentCount == nil || *got.DependentCount != 0 || got.CommentCount == nil || *got.CommentCount != 0 {
		t.Fatalf("ordinary detail rows or counts were lost: %+v", got)
	}
	if len(got.Labels) != 2 || got.Labels[0] != "one" || got.Labels[1] != "two" || string(got.Metadata) != `{"review":true}` {
		t.Fatalf("ordinary Issue labels or metadata were lost: %+v", got)
	}
	for _, id := range []string{got.ID, got.Dependencies[0].ID} {
		var roundTrip []types.IssueDetails
		out := graphPolicyCLI(t, bd, work, home, nil, "", "show", id, "--json")
		if err := json.Unmarshal([]byte(out), &roundTrip); err != nil || len(roundTrip) != 1 || roundTrip[0].ID != id {
			t.Fatalf("Issue detail ID %q did not round-trip: %v\n%s", id, err, out)
		}
	}
	var needs []types.IssueDetails
	dependents := graphPolicyCLI(t, bd, work, home, nil, "", "show", "beads/needs", "--include-dependents", "--json")
	if err := json.Unmarshal([]byte(dependents), &needs); err != nil || len(needs) != 1 || len(needs[0].Dependents) != 1 || needs[0].Dependents[0].ID != "beads/work" {
		t.Fatalf("dependent ID was not canonical: %v\n%s", err, dependents)
	}

	// The old experimental complete-record envelope remains explicitly
	// addressable, including for a retained version.
	for _, args := range [][]string{{"show", "beads/work", "--format", "graph-json", "--json"}, {"show", "beads/work", "--version", created.Revision, "--format", "graph-json", "--json"}} {
		var envelope struct{ Result graphstore.IssueRecord }
		out := graphPolicyCLI(t, bd, work, home, nil, "", args...)
		if err := json.Unmarshal([]byte(out), &envelope); err != nil || envelope.Result.ID != created.ID || envelope.Result.Revision == "" {
			t.Fatalf("explicit graph record read lost its envelope: %v\n%s", err, out)
		}
	}
	graphPolicyCLI(t, bd, work, home, nil, "", "remember", "Memory body", "--id", "beads/memory", "--title", "Memo", "--json")
	var memory struct{ Result graphstore.Record }
	memoryJSON := graphPolicyCLI(t, bd, work, home, nil, "", "show", "beads/memory", "--json")
	if err := json.Unmarshal([]byte(memoryJSON), &memory); err != nil || memory.Result.ID != "https://example.invalid/show/beads/memory" {
		t.Fatalf("Memory show JSON changed unexpectedly: %v\n%s", err, memoryJSON)
	}
}
