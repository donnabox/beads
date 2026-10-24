//go:build cgo

package main

import (
	"encoding/json"
	"os"
	"reflect"
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
			const scope = "https://example.invalid/show/"
			initArgs := []string{"init", "--graph-mode", "link", "--prefix", "parent", "--scope-url", scope, "--skip-hooks", "--skip-agents", "--non-interactive", "--json"}
			if engine == "server" {
				port := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
				if port == "" {
					t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for ordinary shared-server show qualification")
				}
				initArgs = append(initArgs, "--server", "--external", "--server-host", "127.0.0.1", "--server-port", port, "--server-user", "root")
			}
			graphPolicyCLI(t, bd, work, home, nil, "", initArgs...)
			graphPolicyCLI(t, bd, work, home, nil, "not_found", "show", "--current", "--actor", "nobody", "--json")
			created := graphMixedResult[graphstore.IssueRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", "create", "Visible issue", "--id", "work", "--force", "--description", "Body", "--labels", "one,two", "--estimate", "17", "--metadata", `{"review":true}`, "--json"))
			prerequisite := graphMixedResult[graphstore.IssueRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", "create", "Prerequisite", "--id", "needs", "--force", "--json"))
			// A Memory cannot reuse an allocated public Issue identity. Private
			// storage locators do not create a second public alias.
			for _, nativeID := range []string{created.Properties.ID, prerequisite.Properties.ID} {
				graphPolicyCLI(t, bd, work, home, nil, "identity_reserved", "remember", "Colliding Memory", "--id", nativeID, "--json")
			}
			dependency := graphMixedResult[graphstore.DependencyResult](t, graphPolicyCLI(t, bd, work, home, nil, "", "dep", "add", created.ID, prerequisite.ID, "--json"))
			refsJSON := graphPolicyCLI(t, bd, work, home, nil, "", "show", prerequisite.ID, "--refs", "--json")
			var refs map[string]json.RawMessage
			if err := json.Unmarshal([]byte(refsJSON), &refs); err != nil {
				t.Fatalf("--refs lost the shipping reverse-reference JSON shape: %v\n%s", err, refsJSON)
			}
			var referencedBy []*types.IssueWithDependencyMetadata
			if err := json.Unmarshal(refs[prerequisite.Properties.ID], &referencedBy); err != nil || len(referencedBy) != 1 || referencedBy[0].ID != created.Properties.ID {
				t.Fatalf("--refs lost its Issue rows: %v\n%s", err, refsJSON)
			}
			refsHuman := graphPolicyCLI(t, bd, work, home, nil, "", "show", prerequisite.ID, "--refs")
			if !strings.Contains(refsHuman, "References to needs") || !strings.Contains(refsHuman, "Visible issue") {
				t.Fatalf("--refs lost the shipping human view: %s", refsHuman)
			}
			noRefs := graphPolicyCLI(t, bd, work, home, nil, "", "show", created.ID, "--refs", "--json")
			if !strings.Contains(noRefs, `"work": []`) {
				t.Fatalf("--refs omitted an Issue with no incoming references: %s", noRefs)
			}
			parent := graphMixedResult[graphstore.IssueRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", "create", "Parent", "--id", "parent-1", "--type", "epic", "--json"))
			child := graphMixedResult[graphstore.IssueRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", "create", "Child", "--parent", parent.ID, "--json"))
			childrenJSON := graphPolicyCLI(t, bd, work, home, nil, "", "show", parent.ID, "--children", "--json")
			var children map[string]json.RawMessage
			if err := json.Unmarshal([]byte(childrenJSON), &children); err != nil {
				t.Fatalf("--children lost the shipping child JSON shape: %v\n%s", err, childrenJSON)
			}
			var childRows []*types.IssueWithDependencyMetadata
			if err := json.Unmarshal(children[parent.Properties.ID], &childRows); err != nil || len(childRows) != 1 || childRows[0].ID != child.Properties.ID || childRows[0].DependencyType != types.DepParentChild {
				t.Fatalf("--children lost its Issue rows: %v\nparent=%+v child=%+v\n%s", err, parent, child, childrenJSON)
			}
			childrenHuman := graphPolicyCLI(t, bd, work, home, nil, "", "show", parent.ID, "--children", "--short")
			if !strings.Contains(childrenHuman, "Children of parent-1") || !strings.Contains(childrenHuman, "Child") {
				t.Fatalf("--children lost the shipping human view: %s", childrenHuman)
			}
			graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "show", parent.ID, "--children", "--format", "graph-json", "--json")
			current := graphMixedResult[graphstore.IssueRecord](t, graphPolicyCLI(t, bd, work, home, nil, "", "show", created.ID, "--format", "graph-json", "--json"))
			human := graphPolicyCLI(t, bd, work, home, nil, "", "show", created.ID)
			for _, want := range []string{"Visible issue", "Body", `"labels": [`, `"one"`, `"two"`, "Current dependencies", "Prerequisite"} {
				if !strings.Contains(human, want) {
					t.Fatalf("human Issue detail omitted %q: %s", want, human)
				}
			}
			if strings.HasPrefix(strings.TrimSpace(human), "{") {
				t.Fatalf("human Issue detail rendered a JSON record: %s", human)
			}
			short := graphPolicyCLI(t, bd, work, home, nil, "", "show", created.ID, "--short", "--json")
			if !strings.Contains(short, created.Properties.ID) || !strings.Contains(short, "Visible issue") || strings.Contains(short, "DESCRIPTION") || strings.HasPrefix(strings.TrimSpace(short), "{") {
				t.Fatalf("--short did not retain shipping compact output precedence over --json: %s", short)
			}
			long := graphPolicyCLI(t, bd, work, home, nil, "", "show", created.ID, "--long")
			if !strings.Contains(long, "Type identity\n"+current.Type) || !strings.Contains(long, "Exact revision token\n"+current.Revision) || !strings.Contains(long, `"estimated_minutes": 17`) {
				t.Fatalf("--long omitted Issue extended fields: %s", long)
			}

			shown := graphPolicyCLI(t, bd, work, home, nil, "", "show", created.ID, "--json")
			graphPolicyCLI(t, bd, work, home, nil, "", "show", prerequisite.ID, "--json")
			var fallback []types.IssueDetails
			fallbackJSON := graphPolicyCLI(t, bd, work, home, nil, "", "show", "--current", "--actor", "nobody", "--json")
			if err := json.Unmarshal([]byte(fallbackJSON), &fallback); err != nil || len(fallback) != 1 || fallback[0].ID != prerequisite.Properties.ID {
				t.Fatalf("--current did not use the last-touched fallback after empty actor reads: %v\n%s", err, fallbackJSON)
			}
			graphPolicyCLI(t, bd, work, home, nil, "invalid_selector", "show", created.ID, "--current", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "invalid_selector", "show", "--id="+created.Properties.ID, "--current", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "", "create", "Active work", "--id", "active", "--force", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "", "update", scope+"active", "--claim", "--actor", "holder", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "", "show", prerequisite.ID, "--json")
			var active []types.IssueDetails
			activeJSON := graphPolicyCLI(t, bd, work, home, nil, "", "show", "--current", "--actor", "holder", "--json")
			if err := json.Unmarshal([]byte(activeJSON), &active); err != nil || len(active) != 1 || active[0].ID != "active" {
				t.Fatalf("--current did not prefer claimed Issue over last touched: %v\n%s", err, activeJSON)
			}
			byID := graphPolicyCLI(t, bd, work, home, nil, "", "show", "--id="+created.Properties.ID, "--json")
			if byID != shown {
				t.Fatalf("--id must select the same Issue as the positional operand:\n%s\n%s", byID, shown)
			}
			manyJSON := graphPolicyCLI(t, bd, work, home, nil, "", "show", created.ID, "--id="+prerequisite.Properties.ID, "--json")
			var many []types.IssueDetails
			if err := json.Unmarshal([]byte(manyJSON), &many); err != nil || len(many) != 2 || many[0].ID != created.Properties.ID || many[1].ID != prerequisite.Properties.ID {
				t.Fatalf("multi-ID JSON lost shipping Issue order or detail shape: %v\n%s", err, manyJSON)
			}
			partialJSON := graphPolicyCLI(t, bd, work, home, nil, "", "show", scope+"beads/missing", "--id="+prerequisite.Properties.ID, "--json")
			var partial []types.IssueDetails
			if err := json.Unmarshal([]byte(partialJSON), &partial); err != nil || len(partial) != 1 || partial[0].ID != prerequisite.Properties.ID {
				t.Fatalf("missing multi-ID member prevented a successful Issue read: %v\n%s", err, partialJSON)
			}
			graphPolicyCLI(t, bd, work, home, nil, "not_found", "show", scope+"beads/missing-one", "--id=beads/missing-two", "--json")
			manyHuman := graphPolicyCLI(t, bd, work, home, nil, "", "show", created.ID, prerequisite.ID)
			if !strings.Contains(manyHuman, "Visible issue") || !strings.Contains(manyHuman, "Prerequisite") || !strings.Contains(manyHuman, "─") {
				t.Fatalf("multi-ID human output lost Issue detail separator: %s", manyHuman)
			}
			manyShort := graphPolicyCLI(t, bd, work, home, nil, "", "show", created.ID, "--id="+prerequisite.Properties.ID, "--short", "--json")
			if strings.Count(strings.TrimSpace(manyShort), "\n") != 1 || strings.Contains(manyShort, "DESCRIPTION") || strings.HasPrefix(strings.TrimSpace(manyShort), "[") {
				t.Fatalf("multi-ID short output did not take precedence over JSON: %s", manyShort)
			}
			graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "show", created.ID, dependency.Link.ID, "--json")
			var details []types.IssueDetails
			if err := json.Unmarshal([]byte(shown), &details); err != nil || len(details) != 1 {
				t.Fatalf("graph Issue show must use ordinary flat detail array: %v\n%s", err, shown)
			}
			got := details[0]
			if got.ID != created.Properties.ID || got.Title != "Visible issue" || got.Description != "Body" || got.Revision != current.Revision {
				t.Fatalf("legacy Issue identity, content or graph guard token changed: %+v", got)
			}
			if got.DependencyCount == nil || *got.DependencyCount != 1 || len(got.Dependencies) != 1 || got.Dependencies[0].ID != prerequisite.Properties.ID || got.Dependencies[0].Title != "Prerequisite" || got.DependentCount == nil || *got.DependentCount != 0 || got.CommentCount == nil || *got.CommentCount != 0 {
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
					t.Fatalf("public Issue ID %q did not round-trip through its local selector: %v\n%s", id, err, out)
				}
			}
			var needs []types.IssueDetails
			dependents := graphPolicyCLI(t, bd, work, home, nil, "", "show", prerequisite.ID, "--include-dependents", "--json")
			if err := json.Unmarshal([]byte(dependents), &needs); err != nil || len(needs) != 1 || len(needs[0].Dependents) != 1 || needs[0].Dependents[0].ID != created.Properties.ID {
				t.Fatalf("dependent ID was not canonical: %v\n%s", err, dependents)
			}
			graphPolicyCLI(t, bd, work, home, nil, "", "comment", created.ID, "First note", "--actor", "alice", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "", "comment", created.ID, "Second note", "--actor", "bob", "--json")
			tail := graphPolicyCLI(t, bd, work, home, nil, "", "show", created.ID, "--comments-tail", "1")
			if !strings.Contains(tail, "1 older comments hidden") || !strings.Contains(tail, "Second note") || strings.Contains(tail, "First note") {
				t.Fatalf("--comments-tail did not preserve familiar elision: %s", tail)
			}
			retainedJSON := graphPolicyCLI(t, bd, work, home, nil, "", "show", created.ID, "--revision", "="+created.Revision, "--json")
			retainedLongJSON := graphPolicyCLI(t, bd, work, home, nil, "", "show", created.ID, "--revision", "="+created.Revision, "--long", "--json")
			retained := graphMixedResult[graphstore.IssueRecord](t, retainedLongJSON)
			if retainedLongJSON != retainedJSON || !strings.Contains(retainedJSON, `"schemaVersion":1`) || retained.ID != created.ID || retained.Revision != created.Revision || retained.Type != created.Type || !reflect.DeepEqual(retained.Properties, created.Properties) || !reflect.DeepEqual(retained.Metadata, created.Metadata) || !reflect.DeepEqual(retained.Owned, created.Owned) {
				t.Fatalf("retained --long changed complete stored record: %s", retainedLongJSON)
			}
			if strings.Contains(retainedLongJSON, "First note") || strings.Contains(retainedLongJSON, "Second note") || strings.Contains(retainedLongJSON, dependency.Link.ID) || (current.Revision != created.Revision && strings.Contains(retainedLongJSON, current.Revision)) {
				t.Fatalf("retained --long hydrated current comments, relations or revision: %s", retainedLongJSON)
			}

			// The old experimental complete-record envelope remains explicitly
			// addressable, including for a retained version.
			for _, args := range [][]string{{"show", created.ID, "--format", "graph-json", "--json"}, {"show", created.ID, "--revision", "=" + created.Revision, "--format", "graph-json", "--json"}} {
				var envelope struct{ Result graphstore.IssueRecord }
				out := graphPolicyCLI(t, bd, work, home, nil, "", args...)
				if err := json.Unmarshal([]byte(out), &envelope); err != nil || envelope.Result.ID != created.ID || envelope.Result.Revision == "" {
					t.Fatalf("explicit graph record read lost its envelope: %v\n%s", err, out)
				}
			}
			graphPolicyCLI(t, bd, work, home, nil, "", "remember", "Memory body", "--id", "beads/memory", "--title", "Memo", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "show", created.ID, scope+"beads/memory", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "show", scope+"beads/memory", "--refs", "--json")
			var memory struct{ Result graphstore.Record }
			memoryJSON := graphPolicyCLI(t, bd, work, home, nil, "", "show", scope+"beads/memory", "--json")
			if err := json.Unmarshal([]byte(memoryJSON), &memory); err != nil || memory.Result.ID != "https://example.invalid/show/beads/memory" {
				t.Fatalf("Memory show JSON changed unexpectedly: %v\n%s", err, memoryJSON)
			}
			graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "show", scope+"beads/memory", "--short", "--json")
		})
	}
}
