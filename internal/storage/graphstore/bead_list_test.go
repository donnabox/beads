package graphstore

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/internal/workapi"
	publicops "github.com/steveyegge/beads/issueops"
)

const (
	beadListScope     = "https://example.test/"
	beadListMemoryTyp = beadListScope + "types/memory"
	beadListIssueTyp  = beadListScope + "types/issue"
	beadListRecorded  = "2026-09-26T12:00:00Z"
)

func beadListMemory(path string) Record {
	return Record{ID: beadListScope + path, Type: beadListMemoryTyp, Properties: Properties{Title: path},
		Attribution: Attribution{Status: "unknown", RecordedAt: beadListRecorded}}
}

func beadListIssue(path string, status types.Status, pinned bool) IssueRecord {
	return IssueRecord{ID: beadListScope + path, Type: beadListIssueTyp, Properties: &publicops.Issue{Title: path, Status: status, Pinned: pinned},
		Attribution: Attribution{Status: "unknown", RecordedAt: beadListRecorded}}
}

// beadListPaths names the listed Beads without their scope, in page order.
func beadListPaths(t *testing.T, page BeadListPage, scope string) []string {
	t.Helper()
	paths := []string{}
	for _, item := range page.Items {
		switch record := item.(type) {
		case Record:
			paths = append(paths, strings.TrimPrefix(record.ID, scope))
		case IssueRecord:
			paths = append(paths, strings.TrimPrefix(record.ID, scope))
		default:
			t.Fatalf("a Bead page holds %T", item)
		}
	}
	return paths
}

// The Bead list hides what the ordinary Issue list hides. The rule is not
// restated here: it is read back from workapi.BuildListFilter, so a change to
// the upstream default exclusion fails this test instead of silently diverging.
func TestIssueHiddenByDefaultMatchesTheOrdinaryIssueList(t *testing.T) {
	cfg := workapi.ListConfig{CustomStatuses: []types.CustomStatus{
		{Name: "triaged", Category: types.CategoryActive},
		{Name: "reviewing", Category: types.CategoryWIP},
		{Name: "archived", Category: types.CategoryDone},
		{Name: "on-ice", Category: types.CategoryFrozen},
	}}
	statuses := append(slices.Clone(types.AllStatuses), "triaged", "reviewing", "archived", "on-ice")

	filter, err := workapi.BuildListFilter(publicops.ListRequest{}, cfg)
	if err != nil {
		t.Fatal(err)
	}
	ordinaryHides := func(status types.Status, pinned bool) bool {
		return slices.Contains(filter.ExcludeStatus, status) || (pinned && filter.Pinned != nil && !*filter.Pinned)
	}
	hidden, shown := 0, 0
	for _, status := range statuses {
		for _, pinned := range []bool{false, true} {
			want, got := ordinaryHides(status, pinned), issueHiddenByDefault(status, pinned, cfg)
			kind := "Issue"
			if pinned {
				kind = "pinned Issue"
			}
			switch {
			case want && !got:
				t.Errorf("%s visible: status %q pinned=%t is hidden by the ordinary list", kind, status, pinned)
			case !want && got:
				t.Errorf("%s hidden: status %q pinned=%t is shown by the ordinary list", kind, status, pinned)
			}
			if want {
				hidden++
			} else {
				shown++
			}
		}
	}
	// The table must hold both outcomes, or a predicate that hides everything
	// (or nothing) would agree with a filter that does too.
	if hidden == 0 || shown == 0 {
		t.Fatalf("parity table is one-sided: hidden=%d shown=%d", hidden, shown)
	}

	all, err := workapi.BuildListFilter(publicops.ListRequest{AllFlag: true}, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(all.ExcludeStatus) != 0 || all.Pinned != nil {
		t.Fatalf("the ordinary list with --all still excludes: %+v pinned=%v", all.ExcludeStatus, all.Pinned)
	}
	records := []any{beadListMemory("beads/plan")}
	for _, status := range statuses {
		records = append(records, beadListIssue("beads/"+string(status), status, false), beadListIssue("beads/flag-"+string(status), status, true))
	}
	page, err := pageBeads(records, cfg, BeadListRequest{All: true})
	if err != nil || len(page.Items) != len(records) {
		t.Fatalf("--all must hide nothing: %d of %d Beads, err=%v", len(page.Items), len(records), err)
	}
}

func TestPageBeadsHidesWhatTheOrdinaryListHides(t *testing.T) {
	cfg := workapi.ListConfig{CustomStatuses: []types.CustomStatus{
		{Name: "triaged", Category: types.CategoryActive},
		{Name: "archived", Category: types.CategoryDone},
		{Name: "on-ice", Category: types.CategoryFrozen},
	}}
	records := []any{
		beadListMemory("beads/plan"),
		beadListIssue("beads/open", types.StatusOpen, false),
		beadListIssue("beads/closed", types.StatusClosed, false),
		beadListIssue("beads/pinned", types.StatusPinned, false),
		beadListIssue("beads/flagged", types.StatusOpen, true),
		beadListIssue("beads/archived", "archived", false),
		beadListIssue("beads/on-ice", "on-ice", false),
		beadListIssue("beads/triaged", "triaged", false),
		LinkRecord{ID: beadListScope + "links/context"},
	}
	visibleIssues := []string{"beads/open", "beads/triaged"}
	everyIssue := []string{"beads/open", "beads/closed", "beads/pinned", "beads/flagged", "beads/archived", "beads/on-ice", "beads/triaged"}
	for _, tc := range []struct {
		name    string
		request BeadListRequest
		want    []string
	}{
		{"default hides closed, pinned and done or frozen", BeadListRequest{}, append([]string{"beads/plan"}, visibleIssues...)},
		{"all shows every Memory and Issue", BeadListRequest{All: true}, append([]string{"beads/plan"}, everyIssue...)},
		{"Issue Type hides too", BeadListRequest{TypeURL: beadListIssueTyp}, visibleIssues},
		{"Issue Type with all", BeadListRequest{TypeURL: beadListIssueTyp, All: true}, everyIssue},
		{"Memory Type is never hidden", BeadListRequest{TypeURL: beadListMemoryTyp}, []string{"beads/plan"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			page, err := pageBeads(records, cfg, tc.request)
			if err != nil || page.HasMore {
				t.Fatalf("page=%+v err=%v", page, err)
			}
			got := beadListPaths(t, page, beadListScope)
			slices.Sort(got)
			want := slices.Clone(tc.want)
			slices.Sort(want)
			if !slices.Equal(got, want) {
				t.Fatalf("listed %v, want %v", got, want)
			}
		})
	}
}

func TestPageBeadsLimitAndCapCountOnlyTheBeadsThatAreShown(t *testing.T) {
	records := []any{
		beadListMemory("beads/plan"),
		beadListIssue("beads/open", types.StatusOpen, false),
		beadListIssue("beads/closed", types.StatusClosed, false),
		beadListIssue("beads/pinned", types.StatusPinned, false),
	}
	page, err := pageBeads(records, workapi.ListConfig{}, BeadListRequest{Limit: 1})
	if err != nil || len(page.Items) != 1 || !page.HasMore {
		t.Fatalf("a limit under the shown Beads must report more: %+v %v", page, err)
	}
	page, err = pageBeads(records, workapi.ListConfig{}, BeadListRequest{Limit: 2})
	if err != nil || len(page.Items) != 2 || page.HasMore {
		t.Fatalf("hidden Beads must not count as more: %+v %v", page, err)
	}
	page, err = pageBeads(records, workapi.ListConfig{}, BeadListRequest{All: true, Limit: 2})
	if err != nil || len(page.Items) != 2 || !page.HasMore {
		t.Fatalf("an explicit limit still trims --all: %+v %v", page, err)
	}
	// Two Beads are shown and two are hidden: a cap of two admits the listing
	// that a cap counted over the hidden Beads would have refused.
	if page, err = pageBeads(records, workapi.ListConfig{}, BeadListRequest{MaxRows: 2}); err != nil || len(page.Items) != 2 {
		t.Fatalf("hiding must bring the shown Beads under the cap: %+v %v", page, err)
	}
	if _, err = pageBeads(records, workapi.ListConfig{}, BeadListRequest{MaxRows: 1}); !errors.Is(err, ErrLimitExceeded) {
		t.Fatalf("two shown Beads must trip a cap of one: %v", err)
	}
	if _, err = pageBeads(records, workapi.ListConfig{}, BeadListRequest{All: true, MaxRows: 3}); !errors.Is(err, ErrLimitExceeded) {
		t.Fatalf("--all shows four Beads, which must trip a cap of three: %v", err)
	}
}

func TestPageBeadsRefusesRecordsItCannotProject(t *testing.T) {
	for name, record := range map[string]any{
		"an Issue without properties": IssueRecord{ID: beadListScope + "beads/bare", Type: beadListIssueTyp},
		"an unknown projection":       fmt.Sprint("not a record"),
	} {
		if _, err := pageBeads([]any{record}, workapi.ListConfig{}, BeadListRequest{}); !errors.Is(err, ErrInvalidStore) {
			t.Fatalf("%s must be an invalid store, got %v", name, err)
		}
	}
}
