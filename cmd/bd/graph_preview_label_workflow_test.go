//go:build cgo

package main

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/types"
	publicops "github.com/steveyegge/beads/issueops"
)

func TestGraphPreviewLabelListSelectedIssue(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			const scope = "https://example.invalid/labels/"
			work, home := graphListWorkspace(t, bd, engine, scope, "label")
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", args...)
			}
			labels := func(selector string) []string {
				t.Helper()
				var got []string
				if err := json.Unmarshal([]byte(call("label", "list", selector, "--json")), &got); err != nil {
					t.Fatal(err)
				}
				return got
			}
			id := graphListSeedIssue(t, work, "beads/labeled", publicops.CreateRequest{Actor: "author", ForceIDPrefix: true, Issue: &publicops.Issue{
				ID: "label-a3f8e9", Title: "Labeled", IssueType: types.TypeTask, Status: types.StatusOpen,
				Labels: []string{"red", "blue"},
			}})
			hash := id[strings.LastIndex(id, "-")+1:]
			abbreviation := hash
			if len(abbreviation) > 4 {
				abbreviation = abbreviation[:4]
			}
			for _, selector := range []string{id, abbreviation, scope + id} {
				if got := labels(selector); !reflect.DeepEqual(got, []string{"blue", "red"}) {
					t.Fatalf("label list %q: %v", selector, got)
				}
			}
			call("create", "Empty", "--id", "beads/empty", "--force", "--json")
			if got := labels("beads/empty"); got == nil || len(got) != 0 {
				t.Fatalf("empty labels must be a JSON array: %v", got)
			}
			if out := call("label", "list", "beads/empty"); !strings.Contains(out, "beads/empty has no labels") {
				t.Fatalf("empty human label list: %q", out)
			}
			call("remember", "Memo", "--id", "beads/memo", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "label", "list", "beads/memo", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "label", "add", id, "new", "--json")
		})
	}
}
