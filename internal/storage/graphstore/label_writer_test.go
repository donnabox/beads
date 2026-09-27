//go:build cgo

package graphstore

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/storage/issueops"
	"github.com/steveyegge/beads/internal/types"
	publicops "github.com/steveyegge/beads/issueops"
)

// These are shared native-writer prerequisite tests, not a graph label adapter.
// ExecuteCreate intentionally does not install a canonical graph mapping. All
// subsequent reads use native issueops; no graph current/retention claim follows.
func labelWriterFixture(t *testing.T, backend string, labels []string) (context.Context, *Store, *types.Issue) {
	t.Helper()
	ctx, _, s := issueListFixture(t, backend)
	// All mutations below run over real *sql.Tx values; enable the existing
	// test-only journal context so no-op and rollback checks cover that surface.
	ctx = issueops.WithEventsJournal(ctx, true)
	request := plainIssue("Native label writer")
	request.Issue.CreatedAt = time.Date(2001, 2, 3, 4, 5, 6, 0, time.UTC)
	request.Issue.UpdatedAt = time.Date(2002, 3, 4, 5, 6, 7, 0, time.UTC)
	request.Issue.Labels = slices.Clone(labels)
	var created publicops.CreateResult
	if err := s.withTx(ctx, true, func(tx *sql.Tx) error {
		var err error
		created, _, err = issueops.ExecuteCreate(ctx, tx, request)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if created.Issue == nil || !created.Issue.UpdatedAt.Equal(request.Issue.UpdatedAt) || !created.Issue.CreatedAt.Equal(request.Issue.CreatedAt) {
		t.Fatalf("constituent creation labels changed explicit timestamps: %+v", created.Issue)
	}
	assertNativeLabels(t, created.Issue, labels)
	got := readLabelWriterIssue(t, ctx, s, created.Issue.ID)
	if !reflect.DeepEqual(got, created.Issue) {
		t.Fatalf("returned creation differs from accepted state: got=%+v want=%+v", got, created.Issue)
	}
	return ctx, s, got
}

func assertNativeLabels(t *testing.T, issue *types.Issue, expected []string) {
	t.Helper()
	if issue == nil {
		t.Fatal("nil native Issue")
	}
	got, want := slices.Clone(issue.Labels), slices.Clone(expected)
	slices.Sort(got)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Fatalf("labels=%q want=%q", got, want)
	}
}

func readLabelWriterIssue(t *testing.T, ctx context.Context, s *Store, id string) *types.Issue {
	t.Helper()
	var got *types.Issue
	if err := s.withTx(ctx, false, func(tx *sql.Tx) error {
		var err error
		got, err = issueops.HydrateIssueOperationResult(ctx, tx, id, true)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	return got
}

func labelWriterCounts(t *testing.T, ctx context.Context, s *Store) map[string]int {
	t.Helper()
	counts := map[string]int{}
	for _, table := range []string{"issues", "labels", "events", "bd_events_journal", "issue_versions", "graph_preview_catalog", "graph_preview_issue_versions"} {
		var n int
		if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table).Scan(&n); err != nil {
			t.Fatal(err)
		}
		counts[table] = n
	}
	return counts
}

func TestNativeLabelWriterMutationAndNoop(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			ctx, s, current := labelWriterFixture(t, backend, []string{"seed", "A", "a", "e", "é"})
			// Case and accent variants are actual distinct labels under the
			// initialized schema; no casefolded set or fixture collation is assumed.
			for _, tc := range []struct {
				name    string
				patch   publicops.LabelPatch
				want    []string
				changed bool
				events  int
			}{
				{"add", publicops.LabelPatch{Add: []string{"new"}}, []string{"seed", "A", "a", "e", "é", "new"}, true, 1},
				{"remove only uppercase", publicops.LabelPatch{Remove: []string{"A"}}, []string{"seed", "a", "e", "é", "new"}, true, 1},
				{"replacement", publicops.LabelPatch{Replace: publicops.Field[[]string]{Set: true, Value: []string{"a", "e", "é", "replacement"}}}, []string{"a", "e", "é", "replacement"}, true, 3},
				{"duplicate add", publicops.LabelPatch{Add: []string{"é", "é"}}, []string{"a", "e", "é", "replacement"}, false, 0},
				{"missing remove", publicops.LabelPatch{Remove: []string{"missing"}}, []string{"a", "e", "é", "replacement"}, false, 0},
				{"reordered duplicate replacement", publicops.LabelPatch{Replace: publicops.Field[[]string]{Set: true, Value: []string{"replacement", "é", "e", "a", "é"}}}, []string{"a", "e", "é", "replacement"}, false, 0},
			} {
				t.Run(tc.name, func(t *testing.T) {
					if tc.changed && tc.name != "add" {
						// issues.updated_at is DATETIME(0). Separate consecutive
						// actual mutations so strict advancement is observable;
						// the first edit replaces an explicit 2002 timestamp.
						// No-op checks deliberately run without this delay.
						time.Sleep(1100 * time.Millisecond)
					}
					before, counts := current, labelWriterCounts(t, ctx, s)
					var result publicops.UpdateResult
					if err := s.withTx(ctx, true, func(tx *sql.Tx) error {
						var err error
						result, _, err = issueops.ExecuteUpdate(ctx, tx, publicops.UpdateRequest{IssueID: before.ID, Actor: "label-editor", IssuePlaneOnly: true, Patch: publicops.IssuePatch{Labels: tc.patch}})
						return err
					}); err != nil {
						t.Fatal(err)
					}
					if result.Changed != tc.changed {
						t.Fatalf("changed=%v want%v", result.Changed, tc.changed)
					}
					assertNativeLabels(t, result.Issue, tc.want)
					current = readLabelWriterIssue(t, ctx, s, before.ID)
					if !reflect.DeepEqual(result.Issue, current) {
						t.Fatalf("returned labels/state incomplete: returned=%+v stored=%+v", result.Issue, current)
					}
					afterCounts := labelWriterCounts(t, ctx, s)
					if tc.changed {
						if !current.UpdatedAt.After(before.UpdatedAt) || current.RowVersion == before.RowVersion {
							t.Fatalf("real labels did not advance timestamp/row token: before=%+v after=%+v", before, current)
						}
						if afterCounts["events"] != counts["events"]+tc.events || afterCounts["labels"] != len(tc.want) {
							t.Fatalf("wrong event/label delta: before=%v after=%v", counts, afterCounts)
						}
						if afterCounts["bd_events_journal"] <= counts["bd_events_journal"] {
							t.Fatal("actual label changes missing native journal entries")
						}
						preserved := *current
						preserved.Labels = before.Labels
						preserved.UpdatedAt = before.UpdatedAt
						preserved.RowVersion = before.RowVersion
						preserved.ContentHash = before.ContentHash
						if !reflect.DeepEqual(preserved, *before) {
							t.Fatal("label-only edit changed unrelated Issue fields")
						}
					} else if !reflect.DeepEqual(current, before) || !reflect.DeepEqual(counts, afterCounts) {
						t.Fatal("no-op changed record or event/version counts")
					}
					if afterCounts["graph_preview_catalog"] != 0 || afterCounts["graph_preview_issue_versions"] != 0 {
						t.Fatal("native-only prerequisite unexpectedly installed graph state")
					}
				})
			}
			// The aggregate skips identical sets before touching SQL. Exercise the
			// contributor's lower-level RowsAffected no-op guards independently.
			before, counts := current, labelWriterCounts(t, ctx, s)
			if err := s.withTx(ctx, true, func(tx *sql.Tx) error {
				if err := issueops.AddLabelInTx(ctx, tx, "labels", "events", current.ID, "é", "other actor"); err != nil {
					return err
				}
				return issueops.RemoveLabelInTx(ctx, tx, "labels", "events", current.ID, "missing", "other actor")
			}); err != nil {
				t.Fatal(err)
			}
			if got := readLabelWriterIssue(t, ctx, s, current.ID); !reflect.DeepEqual(got, before) || !reflect.DeepEqual(counts, labelWriterCounts(t, ctx, s)) {
				t.Fatal("direct native no-op touched snapshot or event/version counts")
			}
		})
	}
}

func TestNativeLabelWriterRollback(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			ctx, s, before := labelWriterFixture(t, backend, []string{"keep", "remove"})
			counts := labelWriterCounts(t, ctx, s)
			fault := errors.New("rollback after complete label replacement")
			err := s.withTx(ctx, true, func(tx *sql.Tx) error {
				result, _, err := issueops.ExecuteUpdate(ctx, tx, publicops.UpdateRequest{IssueID: before.ID, Actor: "rollback-editor", IssuePlaneOnly: true, Patch: publicops.IssuePatch{Labels: publicops.LabelPatch{Replace: publicops.Field[[]string]{Set: true, Value: []string{"keep", "added"}}}}})
				if err != nil {
					return err
				}
				if !result.Changed || !result.Issue.UpdatedAt.After(before.UpdatedAt) || result.Issue.RowVersion == before.RowVersion {
					t.Fatal("rollback probe did not exercise changed labels/timestamp/token")
				}
				assertNativeLabels(t, result.Issue, []string{"keep", "added"})
				return fault
			})
			if !errors.Is(err, fault) {
				t.Fatalf("wrong rollback result: %v", err)
			}
			if got := readLabelWriterIssue(t, ctx, s, before.ID); !reflect.DeepEqual(got, before) {
				t.Fatalf("rollback leaked Issue/labels: %+v", got)
			}
			if !reflect.DeepEqual(counts, labelWriterCounts(t, ctx, s)) {
				t.Fatal("rollback leaked event/version/label rows")
			}
		})
	}
}
