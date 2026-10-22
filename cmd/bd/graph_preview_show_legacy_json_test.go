//go:build cgo

package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
)

func TestGraphPreviewShowLegacyIssueJSON(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			initArgs := []string{"init", "--graph-mode", "link", "--prefix", "show", "--scope-url", "https://example.invalid/show/", "--skip-hooks", "--skip-agents", "--non-interactive", "--json"}
			if engine == "server" {
				port := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
				if port == "" {
					t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for ordinary shared-server show qualification")
				}
				initArgs = append(initArgs, "--server", "--external", "--server-host", "127.0.0.1", "--server-port", port, "--server-user", "root")
			}
			graphPolicyCLI(t, bd, work, home, nil, "", initArgs...)
			graphPolicyCLI(t, bd, work, home, nil, "not_found", "show", "--current", "--actor", "nobody", "--json")
			created := graphMixedResult[graphstore.IssueRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", "create", "Visible issue", "--id", "beads/work", "--description", "Body", "--labels", "one,two", "--estimate", "17", "--metadata", `{"review":true}`, "--json"))
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
			refsJSON := graphPolicyCLI(t, bd, work, home, nil, "", "show", "beads/needs", "--refs", "--json")
			var refs map[string]json.RawMessage
			if err := json.Unmarshal([]byte(refsJSON), &refs); err != nil {
				t.Fatalf("--refs lost the shipping reverse-reference JSON shape: %v\n%s", err, refsJSON)
			}
			var referencedBy []*types.IssueWithDependencyMetadata
			if err := json.Unmarshal(refs["beads/needs"], &referencedBy); err != nil || len(referencedBy) != 1 || referencedBy[0].ID != "beads/work" {
				t.Fatalf("--refs lost its Issue rows: %v\n%s", err, refsJSON)
			}
			refsHuman := graphPolicyCLI(t, bd, work, home, nil, "", "show", "beads/needs", "--refs")
			if !strings.Contains(refsHuman, "References to beads/needs") || !strings.Contains(refsHuman, "Visible issue") {
				t.Fatalf("--refs lost the shipping human view: %s", refsHuman)
			}
			noRefs := graphPolicyCLI(t, bd, work, home, nil, "", "show", "beads/work", "--refs", "--json")
			if !strings.Contains(noRefs, `"beads/work": []`) {
				t.Fatalf("--refs omitted an Issue with no incoming references: %s", noRefs)
			}
			parent := graphMixedResult[graphstore.IssueRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", "create", "Parent", "--id", "beads/parent", "--type", "epic", "--json"))
			child := graphMixedResult[graphstore.IssueRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", "create", "Child", "--parent", "beads/parent", "--json"))
			childrenJSON := graphPolicyCLI(t, bd, work, home, nil, "", "show", "beads/parent", "--children", "--json")
			var children map[string]json.RawMessage
			if err := json.Unmarshal([]byte(childrenJSON), &children); err != nil {
				t.Fatalf("--children lost the shipping child JSON shape: %v\n%s", err, childrenJSON)
			}
			var childRows []*types.IssueWithDependencyMetadata
			if err := json.Unmarshal(children["beads/parent"], &childRows); err != nil || len(childRows) != 1 || childRows[0].ID != "beads/"+child.Properties.ID || childRows[0].DependencyType != types.DepParentChild {
				t.Fatalf("--children lost its Issue rows: %v\nparent=%+v child=%+v\n%s", err, parent, child, childrenJSON)
			}
			childrenHuman := graphPolicyCLI(t, bd, work, home, nil, "", "show", "beads/parent", "--children", "--short")
			if !strings.Contains(childrenHuman, "Children of beads/parent") || !strings.Contains(childrenHuman, "Child") {
				t.Fatalf("--children lost the shipping human view: %s", childrenHuman)
			}
			graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "show", "beads/parent", "--children", "--format", "graph-json", "--json")
			current := graphMixedResult[graphstore.IssueRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", "show", "beads/work", "--format", "graph-json", "--json"))
			human := graphPolicyCLI(t, bd, work, home, nil, "", "show", "beads/work")
			for _, want := range []string{"Visible issue", "DESCRIPTION", "Body", "LABELS:", "one, two", "DEPENDS ON", "Prerequisite"} {
				if !strings.Contains(human, want) {
					t.Fatalf("human Issue detail omitted %q: %s", want, human)
				}
			}
			if strings.HasPrefix(strings.TrimSpace(human), "{") {
				t.Fatalf("human Issue detail rendered a JSON record: %s", human)
			}
			short := graphPolicyCLI(t, bd, work, home, nil, "", "show", "beads/work", "--short", "--json")
			if !strings.Contains(short, "beads/work") || !strings.Contains(short, "Visible issue") || strings.Contains(short, "DESCRIPTION") || strings.HasPrefix(strings.TrimSpace(short), "{") {
				t.Fatalf("--short did not retain shipping compact output precedence over --json: %s", short)
			}
			long := graphPolicyCLI(t, bd, work, home, nil, "", "show", "beads/work", "--long")
			if !strings.Contains(long, "EXTENDED DETAILS") || !strings.Contains(long, "Estimated: 17 minutes") {
				t.Fatalf("--long omitted Issue extended fields: %s", long)
			}

			shown := graphPolicyCLI(t, bd, work, home, nil, "", "show", "beads/work", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "", "show", "beads/needs", "--json")
			var fallback []types.IssueDetails
			fallbackJSON := graphPolicyCLI(t, bd, work, home, nil, "", "show", "--current", "--actor", "nobody", "--json")
			if err := json.Unmarshal([]byte(fallbackJSON), &fallback); err != nil || len(fallback) != 1 || fallback[0].ID != "beads/needs" {
				t.Fatalf("--current did not use the last-touched fallback after empty actor reads: %v\n%s", err, fallbackJSON)
			}
			graphPolicyCLI(t, bd, work, home, nil, "invalid_selector", "show", "beads/work", "--current", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "invalid_selector", "show", "--id=beads/work", "--current", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "", "create", "Active work", "--id", "beads/active", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "", "update", "beads/active", "--claim", "--actor", "holder", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "", "show", "beads/needs", "--json")
			var active []types.IssueDetails
			activeJSON := graphPolicyCLI(t, bd, work, home, nil, "", "show", "--current", "--actor", "holder", "--json")
			if err := json.Unmarshal([]byte(activeJSON), &active); err != nil || len(active) != 1 || active[0].ID != "beads/active" {
				t.Fatalf("--current did not prefer claimed Issue over last touched: %v\n%s", err, activeJSON)
			}
			byID := graphPolicyCLI(t, bd, work, home, nil, "", "show", "--id=beads/work", "--json")
			if byID != shown {
				t.Fatalf("--id must select the same Issue as the positional operand:\n%s\n%s", byID, shown)
			}
			manyJSON := graphPolicyCLI(t, bd, work, home, nil, "", "show", "beads/work", "--id=beads/needs", "--json")
			var many []types.IssueDetails
			if err := json.Unmarshal([]byte(manyJSON), &many); err != nil || len(many) != 2 || many[0].ID != "beads/work" || many[1].ID != "beads/needs" {
				t.Fatalf("multi-ID JSON lost shipping Issue order or detail shape: %v\n%s", err, manyJSON)
			}
			partialJSON := graphPolicyCLI(t, bd, work, home, nil, "", "show", "beads/missing", "--id=beads/needs", "--json")
			var partial []types.IssueDetails
			if err := json.Unmarshal([]byte(partialJSON), &partial); err != nil || len(partial) != 1 || partial[0].ID != "beads/needs" {
				t.Fatalf("missing multi-ID member prevented a successful Issue read: %v\n%s", err, partialJSON)
			}
			graphPolicyCLI(t, bd, work, home, nil, "not_found", "show", "beads/missing-one", "--id=beads/missing-two", "--json")
			manyHuman := graphPolicyCLI(t, bd, work, home, nil, "", "show", "beads/work", "beads/needs")
			if !strings.Contains(manyHuman, "Visible issue") || !strings.Contains(manyHuman, "Prerequisite") || !strings.Contains(manyHuman, "─") {
				t.Fatalf("multi-ID human output lost Issue detail separator: %s", manyHuman)
			}
			manyShort := graphPolicyCLI(t, bd, work, home, nil, "", "show", "beads/work", "--id=beads/needs", "--short", "--json")
			if strings.Count(strings.TrimSpace(manyShort), "\n") != 1 || strings.Contains(manyShort, "DESCRIPTION") || strings.HasPrefix(strings.TrimSpace(manyShort), "[") {
				t.Fatalf("multi-ID short output did not take precedence over JSON: %s", manyShort)
			}
			graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "show", "beads/work", "links/missing", "--json")
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
			var metadata map[string]any
			if err := json.Unmarshal(got.Metadata, &metadata); err != nil || len(metadata) != 1 || metadata["review"] != true || len(got.Labels) != 2 || got.Labels[0] != "one" || got.Labels[1] != "two" {
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
			graphPolicyCLI(t, bd, work, home, nil, "", "comment", "beads/work", "First note", "--actor", "alice", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "", "comment", "beads/work", "Second note", "--actor", "bob", "--json")
			tail := graphPolicyCLI(t, bd, work, home, nil, "", "show", "beads/work", "--comments-tail", "1")
			if !strings.Contains(tail, "1 older comment hidden") || !strings.Contains(tail, "Second note") || strings.Contains(tail, "First note") {
				t.Fatalf("--comments-tail did not preserve familiar elision: %s", tail)
			}
			graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "show", "beads/work", "--version", created.Revision, "--long", "--json")

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
			graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "show", "beads/work", "beads/memory", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "show", "beads/memory", "--refs", "--json")
			var memory struct{ Result graphstore.Record }
			memoryJSON := graphPolicyCLI(t, bd, work, home, nil, "", "show", "beads/memory", "--json")
			if err := json.Unmarshal([]byte(memoryJSON), &memory); err != nil || memory.Result.ID != "https://example.invalid/show/beads/memory" {
				t.Fatalf("Memory show JSON changed unexpectedly: %v\n%s", err, memoryJSON)
			}
			graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "show", "beads/memory", "--short", "--json")
		})
	}
}

func TestGraphPreviewShowNativeAbbreviation(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			initArgs := []string{"init", "--graph-mode", "link", "--prefix", "show", "--scope-url", "https://example.invalid/show/", "--skip-hooks", "--skip-agents", "--non-interactive", "--json"}
			if engine == "server" {
				port := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
				if port == "" {
					t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for shared-server show qualification")
				}
				initArgs = append(initArgs, "--server", "--external", "--server-host", "127.0.0.1", "--server-port", port, "--server-user", "root")
			}
			graphPolicyCLI(t, bd, work, home, nil, "", initArgs...)
			created := graphMixedResult[graphstore.IssueRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", "create", "Selected Issue", "--id", "beads/source-path", "--json"))
			nativeID := created.Properties.ID
			if !strings.HasPrefix(nativeID, "show-") || len(nativeID) < len("show-")+3 {
				t.Fatalf("fixture lacks a native Issue hash: %q", nativeID)
			}
			hash := strings.TrimPrefix(nativeID, "show-")
			abbreviated := hash[:len(hash)-1]
			var details []types.IssueDetails
			shown := graphPolicyCLI(t, bd, work, home, nil, "", "show", abbreviated, "--json")
			if err := json.Unmarshal([]byte(shown), &details); err != nil || len(details) != 1 || details[0].ID != "beads/source-path" || details[0].Title != "Selected Issue" {
				t.Fatalf("bare abbreviation did not select the Issue detail: %v\n%s", err, shown)
			}
			if record := graphMixedResult[graphstore.IssueRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", "show", abbreviated, "--format", "graph-json", "--json")); record.ID != created.ID {
				t.Fatalf("canonical show selected a different Issue: %+v", record)
			}
			short := graphPolicyCLI(t, bd, work, home, nil, "", "show", abbreviated, "--short")
			if !strings.Contains(short, "beads/source-path") || !strings.Contains(short, "Selected Issue") {
				t.Fatalf("short show lost the resolved graph identity: %s", short)
			}
			graphPolicyCLI(t, bd, work, home, nil, "", "remember", "Colliding Memory", "--id", "beads/"+abbreviated, "--json")
			graphPolicyCLI(t, bd, work, home, nil, "invalid_properties", "show", abbreviated, "--json")
			if memory := graphMixedResult[graphstore.Record](t, graphPolicyCLI(t, bd, work, home, nil, "", "show", "beads/"+abbreviated, "--json")); memory.Properties.Title != "Colliding Memory" {
				t.Fatalf("explicit graph Memory path was not exact: %+v", memory)
			}
		})
	}
}
