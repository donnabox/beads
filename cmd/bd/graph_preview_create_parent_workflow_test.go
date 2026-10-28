//go:build cgo

package main

import (
	"encoding/json"
	"os"
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
	parent := graphMixedResult[graphstore.IssueRecord](t, call("", "create", "Parent", "--graph-path", "beads/selected-parent",
		"--type", "epic", "--labels", "inherited", "--json"))
	child := graphMixedResult[graphstore.IssueRecord](t, call("", "create", "Child", "--parent", parent.ID,
		"--labels", "own", "--json"))
	if !strings.HasPrefix(child.Properties.ID, parent.Properties.ID+".") ||
		child.ID != scope+child.Properties.ID || len(child.Owned) != 1 ||
		!reflect.DeepEqual(child.Properties.Labels, []string{"inherited", "own"}) {
		t.Fatalf("child ID, owner Link or inherited labels: %+v", child)
	}
	nativeParentType, err := graphstore.NativeDependencyTypeURL(scope, "parent-child")
	if err != nil {
		t.Fatal(err)
	}
	var link graphstore.LinkRecord
	if err := json.Unmarshal(child.Owned[0], &link); err != nil || link.Source != child.ID ||
		link.Target != parent.ID || link.Type != nativeParentType || link.Properties["dep_type"] != "parent-child" {
		t.Fatalf("child has no canonical parent-child Link: %+v %v", link, err)
	}
	var children []struct {
		ID     string  `json:"id"`
		Parent *string `json:"parent"`
	}
	if err := json.Unmarshal([]byte(call("", "children", parent.ID, "--json")), &children); err != nil ||
		len(children) != 1 || children[0].ID != child.Properties.ID ||
		children[0].Parent == nil || *children[0].Parent != parent.Properties.ID {
		t.Fatalf("native child edge missing from children: %+v %v", children, err)
	}
	second := graphMixedResult[graphstore.IssueRecord](t, call("", "create", "Second child", "--parent", parent.ID,
		"--no-inherit-labels", "--labels", "own", "--json"))
	if second.Properties.ID != parent.Properties.ID+".2" || second.ID != scope+second.Properties.ID ||
		!reflect.DeepEqual(second.Properties.Labels, []string{"own"}) {
		t.Fatalf("second child ID or label opt-out: %+v", second)
	}
	before := call("", "children", parent.ID, "--json")
	call("", "remember", "Reserved Memory", "--id", parent.Properties.ID+".3", "--json")
	// A deterministic native child ID cannot be silently renamed when a
	// Memory has reserved that spelling. The writer must roll back the child
	// counter, native Issue and parent Link together.
	call("identity_reserved", "create", "Third child", "--parent", parent.ID,
		"--no-inherit-labels", "--labels", "own", "--json")
	if after := call("", "children", parent.ID, "--json"); after != before {
		t.Fatalf("reserved child ID changed parent children\nbefore: %s\nafter: %s", before, after)
	}
	call("invalid_selector", "create", "Refused", "--parent", parent.ID, "--id", "beads/explicit", "--json")
	call("invalid_selector", "create", "Refused", "--parent=", "--json")
	call("invalid_selector", "create", "Refused", "--no-inherit-labels", "--json")
	call("not_found", "create", "Refused", "--parent", "missing", "--json")
	if after := call("", "children", parent.ID, "--json"); after != before {
		t.Fatalf("refused create changed parent children\nbefore: %s\nafter: %s", before, after)
	}
}

func TestGraphPreviewCreateSilentTwoEngines(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			args := []string{"init", "--graph-mode", "link", "--prefix", "silent", "--scope-url", "https://example.invalid/silent/", "--skip-hooks", "--skip-agents", "--non-interactive"}
			if engine == "server" {
				port := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
				if port == "" {
					t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for shared-server create qualification")
				}
				args = append(args, "--server", "--external", "--server-host", "127.0.0.1", "--server-port", port, "--server-user", "root")
			}
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", args...)
			}
			call(args...)
			if got := call("create", "Explicit", "--id", "beads/fixed", "--force", "--silent"); got != "beads/fixed\n" {
				t.Fatalf("silent explicit ID output = %q", got)
			}
			if got := call("create", "Generated", "--silent", "--quiet"); !strings.HasPrefix(got, "silent-") || strings.Count(got, "\n") != 1 {
				t.Fatalf("silent generated ID output = %q", got)
			}
			created := graphMixedResult[graphstore.IssueRecord](t, call("create", "JSON", "--id", "beads/json", "--force", "--silent", "--json"))
			if created.Properties.Title != "JSON" || created.ID != "https://example.invalid/silent/beads/json" {
				t.Fatalf("--json did not retain graph create result: %+v", created)
			}
			if got := call("create", "Normal", "--silent=false"); !strings.Contains(got, "Normal") {
				t.Fatalf("explicit false changed normal output: %q", got)
			}
			graphPolicyCLI(t, bd, work, home, nil, "identity_reserved", "create", "Duplicate", "--id", "beads/fixed", "--force", "--silent", "--json")
		})
	}
}
