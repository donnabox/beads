//go:build cgo

package main

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func TestGraphPreviewCreateParentNativeChildAndCanonicalLink(t *testing.T) {
	bd := buildBDUnderTest(t)
	work, home := t.TempDir(), t.TempDir()
	const scope = "https://example.invalid/create-parent/"
	call := func(code string, args ...string) string {
		return graphPolicyCLI(t, bd, work, home, nil, code, args...)
	}
	call("", "init", "--graph-mode", "link", "--prefix", "demo", "--scope-url", scope,
		"--skip-hooks", "--skip-agents", "--non-interactive", "--json")
	parent := graphMixedResult[graphstore.IssueRecord](t, call("", "create", "Parent", "--id", "beads/selected-parent",
		"--type", "epic", "--labels", "inherited", "--json"))
	child := graphMixedResult[graphstore.IssueRecord](t, call("", "create", "Child", "--parent", "selected-parent",
		"--labels", "own", "--json"))
	if !strings.HasPrefix(child.Properties.ID, parent.Properties.ID+".") ||
		child.ID != scope+"beads/"+child.Properties.ID || len(child.Owned) != 1 ||
		!reflect.DeepEqual(child.Properties.Labels, []string{"inherited", "own"}) {
		t.Fatalf("child ID, owner Link or inherited labels: %+v", child)
	}
	var link graphstore.LinkRecord
	if err := json.Unmarshal(child.Owned[0], &link); err != nil || link.Source != child.ID ||
		link.Target != parent.ID || link.Type != scope+"types/dependency" || link.Properties["dep_type"] != "parent-child" {
		t.Fatalf("child has no canonical parent-child Link: %+v %v", link, err)
	}
	var children []struct {
		ID     string  `json:"id"`
		Parent *string `json:"parent"`
	}
	if err := json.Unmarshal([]byte(call("", "children", "selected-parent", "--json")), &children); err != nil ||
		len(children) != 1 || children[0].ID != "beads/"+child.Properties.ID ||
		children[0].Parent == nil || *children[0].Parent != "beads/selected-parent" {
		t.Fatalf("native child edge missing from children: %+v %v", children, err)
	}
	call("", "remember", "Reserved Memory", "--id", "beads/"+parent.Properties.ID+".2", "--json")
	second := graphMixedResult[graphstore.IssueRecord](t, call("", "create", "Second child", "--parent", scope+"beads/selected-parent",
		"--no-inherit-labels", "--labels", "own", "--json"))
	if second.Properties.ID != parent.Properties.ID+".2" || second.ID == scope+"beads/"+second.Properties.ID ||
		!strings.HasPrefix(second.ID, scope+"beads/"+second.Properties.ID+"-") ||
		!reflect.DeepEqual(second.Properties.Labels, []string{"own"}) {
		t.Fatalf("second child ID or label opt-out: %+v", second.Properties)
	}
	before := call("", "children", "selected-parent", "--json")
	call("invalid_selector", "create", "Refused", "--parent", "selected-parent", "--id", "beads/explicit", "--json")
	call("invalid_selector", "create", "Refused", "--parent=", "--json")
	call("invalid_selector", "create", "Refused", "--no-inherit-labels", "--json")
	call("not_found", "create", "Refused", "--parent", "missing", "--json")
	if after := call("", "children", "selected-parent", "--json"); after != before {
		t.Fatalf("refused create changed parent children\nbefore: %s\nafter: %s", before, after)
	}
}
