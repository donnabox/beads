package main

import (
	"errors"
	"reflect"
	"testing"

	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
)

func TestGraphDependencyProjectionUsesMappedBlockingLinks(t *testing.T) {
	const scope = "https://example.invalid/structure/"
	taskType, err := graphstore.NativeIssueTypeURL(scope, "task")
	if err != nil {
		t.Fatal(err)
	}
	blocksType, err := graphstore.NativeDependencyTypeURL(scope, "blocks")
	if err != nil {
		t.Fatal(err)
	}
	tracksType, err := graphstore.NativeDependencyTypeURL(scope, "tracks")
	if err != nil {
		t.Fatal(err)
	}
	issue := func(id string) graphstore.IssueRecord {
		return graphstore.IssueRecord{ID: scope + "native-" + id, Type: taskType, Properties: &types.Issue{ID: "native-" + id, IssueType: types.TypeTask}}
	}
	link := func(id, source, target string) graphstore.LinkRecord {
		return graphstore.LinkRecord{ID: scope + "links/" + id, Type: blocksType, Source: scope + "native-" + source, Target: scope + "native-" + target, Properties: map[string]any{"dep_type": "blocks"}}
	}
	nonblocking := link("custom", "a", "d")
	nonblocking.Type = tracksType
	nonblocking.Properties = map[string]any{"dep_type": "tracks"}
	snapshot := graphstore.Snapshot{Records: []any{
		issue("d"), issue("c"), issue("a"), issue("b"),
		link("a-b", "a", "b"), link("b-c", "b", "c"),
		nonblocking,
		graphstore.LinkRecord{ID: scope + "links/related", Type: graphstore.RelatedTypeURL(scope), Source: scope + "native-a", Target: scope + "native-d"},
	}}
	g, err := projectGraphDependencies(snapshot, scope)
	if err != nil {
		t.Fatal(err)
	}
	issues, links := g.component(scope + "native-b")
	if len(issues) != 3 || len(links) != 2 || issues[0].ID != scope+"native-a" || issues[2].ID != scope+"native-c" {
		t.Fatalf("component included a nonblocking link or lost a blocker: issues=%+v links=%+v", issues, links)
	}
	if down := g.tree(scope+"native-a", "down", 2); len(down) != 2 || down[0].Depth != 0 || down[1].ParentID != scope+"native-a" {
		t.Fatalf("bounded downward tree: %+v", down)
	}
	if up := g.tree(scope+"native-c", "up", 3); len(up) != 3 || up[2].Issue.ID != scope+"native-a" {
		t.Fatalf("upward tree: %+v", up)
	}
	if cycles := g.cycles(); len(cycles) != 0 {
		t.Fatalf("acyclic graph reported cycles: %+v", cycles)
	}
	snapshot.Records = append(snapshot.Records, link("c-a", "c", "a"))
	g, err = projectGraphDependencies(snapshot, scope)
	if err != nil {
		t.Fatal(err)
	}
	if cycles := g.cycles(); len(cycles) != 1 || len(cycles[0]) != 3 {
		t.Fatalf("blocking cycle omitted: %+v", cycles)
	}
	if native, err := g.nativeCycleIDs(); err != nil || !reflect.DeepEqual(native, [][]string{{"native-a", "native-b", "native-c"}}) {
		t.Fatalf("familiar check must use native Issue IDs: cycles=%v err=%v", native, err)
	}
	if canonical := g.cycles(); !reflect.DeepEqual(canonical, [][]string{{scope + "native-a", scope + "native-b", scope + "native-c"}}) {
		t.Fatalf("graph-specific cycles lost canonical IDs: %v", canonical)
	}
	snapshot.Records = append(snapshot.Records, link("unmapped", "a", "missing"))
	if _, err := projectGraphDependencies(snapshot, scope); !errors.Is(err, graphstore.ErrInvalidStore) {
		t.Fatalf("unmapped blocking endpoint accepted: %v", err)
	}
}
