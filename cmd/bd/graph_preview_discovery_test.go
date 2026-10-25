package main

import (
	"reflect"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/query"
	"github.com/steveyegge/beads/internal/types"
)

func TestGraphSearchMatchesCanonicalIDsAndNativeText(t *testing.T) {
	external := "gh-42"
	rows := []*types.IssueWithCounts{
		{Issue: &types.Issue{ID: "beads/work", Title: "Authentication work", ExternalRef: &external}},
		{Issue: &types.Issue{ID: "beads/other", Title: "Other"}},
	}
	for _, tc := range []struct {
		query string
		want  []string
	}{
		{"beads/work", []string{"beads/work"}},
		{"authentication", []string{"beads/work"}},
		{"gh-42", []string{"beads/work"}},
		{"other", []string{"beads/other"}},
		{"missing", nil},
		{"Auth%work", []string{"beads/work"}},
		{"beads/w_rk", []string{"beads/work"}},
	} {
		found, err := graphSearchMatches(rows, tc.query)
		if err != nil {
			t.Fatal(err)
		}
		var ids []string
		for _, row := range found {
			ids = append(ids, row.ID)
		}
		if !reflect.DeepEqual(ids, tc.want) {
			t.Errorf("search %q: got %v, want %v", tc.query, ids, tc.want)
		}
	}
}

func TestGraphQueryFieldAdmissionUsesAST(t *testing.T) {
	for _, tc := range []struct {
		expression  string
		unsupported bool
	}{
		{"title=parenting AND type=task", false},
		{"parent=beads/work", false},
		{"type=task OR ephemeral=true", true},
		{"NOT mol_type=work", true},
	} {
		if got := graphQueryUnsupportedField(tc.expression); got != tc.unsupported {
			t.Errorf("%q: unsupported=%t, want %t", tc.expression, got, tc.unsupported)
		}
	}
	match, err := query.PredicateAt("id=beads/work OR title=Other", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if !match(&types.Issue{ID: "beads/work"}) || !match(&types.Issue{Title: "Other"}) || match(&types.Issue{ID: "beads/missing"}) {
		t.Fatal("full predicate did not match canonical graph IDs and boolean alternatives")
	}
}

func TestGraphDiscoveryCanonicalCountSelectors(t *testing.T) {
	parent := "beads/work"
	other := "beads/other"
	rows := []*types.IssueWithCounts{
		{Issue: &types.Issue{ID: parent}},
		{Issue: &types.Issue{ID: "beads/work.child"}},
		{Issue: &types.Issue{ID: "beads/edge"}, Parent: &parent},
		{Issue: &types.Issue{ID: "beads/work.other-edge"}, Parent: &other},
	}
	ids := func(rows []*types.IssueWithCounts) []string {
		got := make([]string, 0, len(rows))
		for _, row := range rows {
			got = append(got, row.ID)
		}
		return got
	}
	if got := ids(graphDiscoverySelectRows(rows, map[string]bool{"beads/edge": true}, "", false)); !reflect.DeepEqual(got, []string{"beads/edge"}) {
		t.Fatalf("ID filter = %v", got)
	}
	if got := ids(graphDiscoverySelectRows(rows, nil, parent, false)); !reflect.DeepEqual(got, []string{"beads/work.child", "beads/edge"}) {
		t.Fatalf("parent filter = %v", got)
	}
	if got := ids(graphDiscoverySelectRows(rows, nil, "", true)); !reflect.DeepEqual(got, []string{"beads/work", "beads/work.child"}) {
		t.Fatalf("no-parent filter = %v", got)
	}
	parentNode, err := query.Parse("parent=beads/work AND type=task")
	if err != nil {
		t.Fatal(err)
	}
	match, err := graphQueryPredicate(parentNode, time.Now(), "https://example.invalid/discovery/")
	if err != nil {
		t.Fatal(err)
	}
	rows[1].IssueType, rows[2].IssueType = types.TypeTask, types.TypeBug
	if !match(rows[1]) || match(rows[2]) || match(rows[3]) {
		t.Fatal("query parent predicate lost canonical parent or other field")
	}
}

func TestGraphDiscoveryIssuePathWithExplicitScope(t *testing.T) {
	scope := "https://example.invalid/discovery/"
	for _, tc := range []struct {
		selector string
		want     string
	}{
		{"work", "work"},
		{scope + "work", "work"},
		{"links/work", "links/work"},
		{"types/work", "types/work"},
		{"beads/work", "beads/work"},
		{scope + "beads/work", "beads/work"},
	} {
		got, err := graphDiscoveryIssuePath(scope, tc.selector)
		if err != nil || got != tc.want {
			t.Errorf("selector %q: got %q, %v; want %q", tc.selector, got, err, tc.want)
		}
	}
	if _, err := graphDiscoveryIssuePath(scope, "a//b"); err == nil {
		t.Fatal("empty path segment was accepted")
	}

}
