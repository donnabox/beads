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
		{"parent=beads/work", true},
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
