package graphstore

import (
	"errors"
	"testing"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/resourceget"
)

// A whole URI filter selects nominal identity, while each returned Resource
// retains the exact ancestor editions under which it was admitted. In
// particular, keeping an old fact must not make it a second current edition.
func TestResourceSnapshotKeepsCurrentAndRetainedTypeFactsDistinct(t *testing.T) {
	old := resourceget.TypeReference{URI: "https://types.example/child", Revision: "old"}
	current := resourceget.TypeReference{URI: old.URI, Revision: "current"}
	baseOld := resourceget.TypeReference{URI: "https://types.example/base", Revision: "base-old"}
	baseCurrent := resourceget.TypeReference{URI: baseOld.URI, Revision: "base-current"}
	newParent := resourceget.TypeReference{URI: "https://types.example/new-parent", Revision: "parent-current"}
	snapshot := ResourceSnapshot{Types: []ResourceTypeFact{
		{Reference: old, Definition: old, Describes: graph.KindBead, Ancestors: []resourceget.TypeReference{old, baseOld}},
		{Reference: current, Definition: current, Describes: graph.KindBead, Current: true, Ancestors: []resourceget.TypeReference{current, newParent}},
		{Reference: baseCurrent, Definition: baseCurrent, Describes: graph.KindBead, Current: true, Ancestors: []resourceget.TypeReference{baseCurrent}},
		{Reference: newParent, Definition: newParent, Describes: graph.KindBead, Current: true, Ancestors: []resourceget.TypeReference{newParent}},
	}}
	selected, err := snapshot.ResolveType(old.URI)
	if err != nil || selected != current {
		t.Fatalf("current selection = %#v, %v", selected, err)
	}
	matched, err := snapshot.MatchesType(old, baseOld.URI, false)
	if err != nil || !matched {
		t.Fatalf("retained ancestry lost: %v, %v", matched, err)
	}
	matched, err = snapshot.MatchesType(old, newParent.URI, false)
	if err != nil || matched {
		t.Fatalf("current ancestry substituted into retained affiliation: %v, %v", matched, err)
	}
	matched, err = snapshot.MatchesType(old, old.URI, true)
	if err != nil || !matched {
		t.Fatalf("whole nominal identity changed to latest-edition filter: %v, %v", matched, err)
	}
	snapshot.Types[0].Current = true
	if _, err = snapshot.ResolveType(old.URI); !errors.Is(err, ErrInvalidStore) {
		t.Fatalf("two current editions accepted: %v", err)
	}
}

func TestResourceSnapshotRequiresExactCapturedAffiliation(t *testing.T) {
	global := resourceget.TypeReference{URI: "https://types.example/intrinsic", Revision: "meaning-2"}
	local := resourceget.TypeReference{URI: "https://scope.example/type", Revision: "local-edition"}
	snapshot := ResourceSnapshot{Types: []ResourceTypeFact{{Reference: global, Definition: local, Current: true, Describes: graph.KindBead, Ancestors: []resourceget.TypeReference{global}}}}
	if selected, err := snapshot.ResolveType(local.URI); err != nil || selected != global {
		t.Fatalf("checked global/local correspondence lost: %#v, %v", selected, err)
	}
	if _, err := snapshot.MatchesType(global, local.URI, false); !errors.Is(err, ErrCapabilityUnavailable) {
		t.Fatalf("product lookup leaked into conformance: %v", err)
	}
	if _, err := snapshot.ResolveType("intrinsic"); !errors.Is(err, ErrCapabilityUnavailable) {
		t.Fatalf("basename alias accepted: %v", err)
	}
	if _, err := snapshot.MatchesType(resourceget.TypeReference{URI: global.URI, Revision: "meaning-1"}, "", false); !errors.Is(err, ErrInvalidStore) {
		t.Fatalf("uncaptured exact affiliation treated as absence: %v", err)
	}
	snapshot.Types = append(snapshot.Types, snapshot.Types[0])
	if _, err := snapshot.MatchesType(global, "", false); !errors.Is(err, ErrInvalidStore) {
		t.Fatalf("duplicate exact fact accepted: %v", err)
	}
}
