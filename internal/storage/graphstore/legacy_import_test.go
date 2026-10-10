//go:build cgo

package graphstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/legacyimport"
	"github.com/steveyegge/beads/internal/types"
	publicops "github.com/steveyegge/beads/issueops"
)

const legacyImportFixture = `{"_type":"issue","id":"old-a","title":"Prerequisite","priority":0,"status":"closed","issue_type":"task","created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-02T00:00:00Z","closed_at":"2026-01-02T00:00:00Z","close_reason":"done","created_by":"original","owner":"owner@example.invalid"}
{"_type":"issue","id":"old-b","title":"Work","priority":2,"status":"open","issue_type":"bug","created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-02T00:00:00Z","description":"A body","design":"Design","acceptance_criteria":"Tests","notes":"Notes","labels":["two","one"],"metadata":{"custom":{"flag":true}},"dependencies":[{"issue_id":"old-b","depends_on_id":"old-a","type":"blocks","created_at":"2026-01-01T00:00:00Z","created_by":"original","metadata":"{ }"}],"comments":[{"id":"old-comment","issue_id":"old-b","author":"original","text":"A note","created_at":"2026-01-01T00:00:00Z"}]}
{"_type":"memory","key":"context","value":"A remembered fact"}
`

func parseLegacyFixture(t *testing.T, input string) legacyimport.Batch {
	t.Helper()
	batch, err := legacyimport.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	return batch
}

func legacyImportState(t *testing.T, ctx context.Context, s *Store) map[string]string {
	t.Helper()
	state := reopenState(t, ctx, s)
	for _, table := range []string{"labels", "comments", "child_counters"} {
		rows, err := s.db.QueryContext(ctx, "SELECT * FROM "+table)
		if err != nil {
			t.Fatal(err)
		}
		columns, err := rows.Columns()
		if err != nil {
			t.Fatal(err)
		}
		var values []string
		for rows.Next() {
			fields := make([]any, len(columns))
			pointers := make([]any, len(columns))
			for i := range fields {
				pointers[i] = &fields[i]
			}
			if err := rows.Scan(pointers...); err != nil {
				t.Fatal(err)
			}
			raw, err := json.Marshal(fields)
			if err != nil {
				t.Fatal(err)
			}
			values = append(values, string(raw))
		}
		if err := errors.Join(rows.Err(), rows.Close()); err != nil {
			t.Fatal(err)
		}
		sort.Strings(values)
		state[table] = strings.Join(values, "\n")
	}

	var token string
	if err := s.db.QueryRowContext(ctx, `SELECT writer_token FROM graph_preview_scope`).Scan(&token); err != nil {
		t.Fatal(err)
	}
	state["writer_token"] = token
	var prefix string
	if err := s.db.QueryRowContext(ctx, "SELECT value FROM config WHERE `key`='issue_prefix'").Scan(&prefix); err != nil {
		t.Fatal(err)
	}
	state["issue_prefix"] = prefix
	return state
}

func TestLegacyImportAtomicityAndPersistence(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			ctx, o := issueExperimentOptions(t, backend)
			s, err := OpenExisting(ctx, o)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := s.Close(); err != nil {
					t.Error(err)
				}
			})
			batch := parseLegacyFixture(t, legacyImportFixture)
			before := legacyImportState(t, ctx, s)
			for _, stage := range []string{"coordination", "import-prefix", "import-native", "import-issues", "import-links", "source-catalog", "source-retained", "allocation", "payload", "retained", "import-retained"} {
				t.Run(stage, func(t *testing.T) {
					fault := errors.New("injected late import failure")
					s.afterWrite = func(at string) error {
						if at == stage {
							return fault
						}
						return nil
					}
					_, err := s.ImportLegacy(ctx, batch, "importer", false)
					s.afterWrite = nil
					if !errors.Is(err, fault) {
						t.Fatalf("stage %s: %v", stage, err)
					}
					if !reflect.DeepEqual(before, legacyImportState(t, ctx, s)) {
						t.Fatal("failed import left database changes")
					}
				})
			}
			preview, err := s.ImportLegacy(ctx, batch, "importer", true)
			if err != nil || !preview.DryRun {
				t.Fatalf("dry run: %+v %v", preview, err)
			}
			if !reflect.DeepEqual(before, legacyImportState(t, ctx, s)) {
				t.Fatal("dry run changed graph")
			}
			result, err := s.ImportLegacy(ctx, batch, "importer", false)
			if err != nil {
				t.Fatal(err)
			}
			if result.Issues != 2 || result.Memories != 1 || result.Dependencies != 1 || result.Comments != 1 {
				t.Fatalf("result=%+v", result)
			}
			b, err := s.ShowIssue(ctx, "beads/old-b")
			if err != nil || len(b.Owned) != 1 || b.Properties.ID != "old-b" {
				t.Fatalf("issue=%+v err=%v", b, err)
			}
			comments, err := s.IssueComments(ctx, "beads/old-b")
			if err != nil || len(comments) != 1 || comments[0].ID != "old-comment" {
				t.Fatalf("comments=%+v err=%v", comments, err)
			}
			_, versions, err := s.Versions(ctx, "beads/old-b")
			if err != nil || len(versions) != 1 {
				t.Fatalf("initial versions=%+v %v", versions, err)
			}
			imported := legacyImportState(t, ctx, s)
			for _, dry := range []bool{false, true} {
				if _, err := s.ImportLegacy(ctx, batch, "importer", dry); err == nil {
					t.Fatal("repeat import accepted")
				}
				if !reflect.DeepEqual(imported, legacyImportState(t, ctx, s)) {
					t.Fatal("repeat import changed graph")
				}
			}
			if err := s.Close(); err != nil {
				t.Fatal(err)
			}
			s, err = OpenExisting(ctx, o)
			if err != nil {
				t.Fatal(err)
			}
			after, err := s.ShowIssue(ctx, "beads/old-b")
			if err != nil || !reflect.DeepEqual(b, after) {
				t.Fatalf("reopen lost data: %v", err)
			}
			fresh, err := s.CreateIssue(ctx, "beads/new-work", publicops.CreateRequest{Actor: "importer", Issue: &types.Issue{Title: "After import", Priority: 2, Status: types.StatusOpen, IssueType: types.TypeTask}})
			if err != nil || types.ExtractPrefix(fresh.Properties.ID) != "old-" {
				t.Fatalf("post-import ID prefix: %+v %v", fresh, err)
			}
			for _, pair := range [][2]string{{"beads/new-work", "beads/old-a"}, {"beads/old-b", "beads/new-work"}} {
				if _, err := s.AddDependency(ctx, DependencyRequest{SourcePath: pair[0], TargetPath: pair[1], Actor: "importer"}); err != nil {
					t.Fatalf("post-import dependency %v: %v", pair, err)
				}
			}

		})
	}
}

func TestLegacyImportValidationRollsBack(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			ctx, o := issueExperimentOptions(t, backend)
			s, err := OpenExisting(ctx, o)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := s.Close(); err != nil {
					t.Error(err)
				}
			})
			before := legacyImportState(t, ctx, s)
			for _, input := range []string{
				strings.Replace(legacyImportFixture, `"type":"blocks"`, `"type":"parent-child"`, 1),
				strings.Replace(legacyImportFixture, `"depends_on_id":"old-a"`, `"depends_on_id":"missing"`, 1),
				strings.Replace(legacyImportFixture, `"metadata":{"custom":{"flag":true}}`, `"metadata":[1]`, 1),
				strings.Replace(legacyImportFixture, `"closed_at":"2026-01-02T00:00:00Z"`, `"closed_at":"2026-01-02T00:00:00.123Z"`, 1),
				legacyImportFixture + `{"id":"old-a","title":"duplicate"}`,
				`{"id":"x","title":"x","ephemeral":true}`,
				`{"id":"x","title":"x","no_history":true}`,
				`{"id":"x","title":"x","lease_expires_at":"2099-01-01T00:00:00Z"}`,
				`{"id":"x","title":"x","status":"tombstone"}`,
				`{"id":"old-a","title":"a","dependencies":[{"depends_on_id":"old-b","type":"blocks"}]}
{"id":"old-b","title":"b","dependencies":[{"depends_on_id":"old-a","type":"blocks"}]}`,
				`{"id":"one-a","title":"a"}` + "\n" + `{"id":"two-b","title":"b"}`,
				`{"id":"old-a","title":"a","bonded_from":[{"proto_id":"p"}]}`,
				`{"id":"old-a","title":"a","source_formula":"formula"}`,
				`{"id":"old-a","title":"a","source_location":"steps[0]"}`,
				`{"id":"old-a","title":"a","storage_class":"versioned"}`,
				"",
				`{"_schema":"beads-jsonl/1"}`,
			} {
				if _, err := s.ImportLegacy(ctx, parseLegacyFixture(t, input), "importer", false); err == nil {
					t.Fatalf("accepted %s", input)
				}
				if !reflect.DeepEqual(before, legacyImportState(t, ctx, s)) {
					t.Fatal("invalid import changed graph")
				}
			}
		})
	}
}

// These fields appear in ordinary exports but are not exercised by the small
// producer fixture. Import itself verifies the hydrated native row and retained
// graph projection before returning, including during rollback-only dry-run.
func TestLegacyImportFieldFamilies(t *testing.T) {
	families := map[string]string{
		"assignment":     `"assignee":"iris","estimated_minutes":20,"external_ref":"gh-123","spec_id":"spec","source_system":"adapter"`,
		"schedule":       `"status":"deferred","started_at":"2026-01-01T00:00:00Z","defer_until":"2099-01-01T00:00:00Z","due_at":"2099-01-02T00:00:00Z"`,
		"closing":        `"status":"closed","closed_at":"2026-01-02T00:00:00Z","closed_by_session":"session","close_reason":"finished"`,
		"context":        `"pinned":true,"is_template":true`,
		"compaction":     `"compaction_level":1,"compacted_at":"2026-01-02T00:00:00Z","compacted_at_commit":"abc123","original_size":100`,
		"classification": `"mol_type":"swarm","work_type":"open_competition","wisp_type":"patrol","sender":"original"`,
		"gate":           `"await_type":"timer","await_id":"timer-1","timeout":60000000000,"waiters":["iris@example.invalid"]`,
		"event":          `"event_kind":"agent.started","actor":"actor","target":"target","payload":"{\"flag\":true}"`,
	}
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			ctx, o := issueExperimentOptions(t, backend)
			s, err := OpenExisting(ctx, o)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := s.Close(); err != nil {
					t.Error(err)
				}
			})
			before := legacyImportState(t, ctx, s)
			for name, fields := range families {
				t.Run(name, func(t *testing.T) {
					input := `{"id":"old-a","title":"Fields","created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-02T00:00:00Z",` + fields + `}`
					if _, err := s.ImportLegacy(ctx, parseLegacyFixture(t, input), "importer", true); err != nil {
						t.Fatal(err)
					}
					if !reflect.DeepEqual(before, legacyImportState(t, ctx, s)) {
						t.Fatal("field dry-run changed database")
					}
				})
			}
		})
	}
}

func TestLegacyImportCurrentReadLimitRollsBack(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			ctx, o := issueExperimentOptions(t, backend)
			s, err := OpenExisting(ctx, o)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := s.Close(); err != nil {
					t.Error(err)
				}
			})
			before := legacyImportState(t, ctx, s)
			batch := parseLegacyFixture(t, `{"id":"old-a","title":"Large","description":"`+strings.Repeat("x", 9*1024*1024)+`"}`)
			_, err = s.ImportLegacy(ctx, batch, "importer", false)
			if !errors.Is(err, ErrLimitExceeded) {
				t.Fatalf("read-budget error: %v", err)
			}
			if !reflect.DeepEqual(before, legacyImportState(t, ctx, s)) {
				t.Fatal("read-budget refusal left data or prefix")
			}
			many := legacyimport.Batch{}
			for i := 0; i <= PreviewSnapshotLimit; i++ {
				many.Memories = append(many.Memories, legacyimport.Memory{Key: fmt.Sprintf("key-%d", i), Value: "value"})
			}
			if _, err := s.ImportLegacy(ctx, many, "importer", false); !errors.Is(err, ErrLimitExceeded) {
				t.Fatalf("count-budget error: %v", err)
			}
		})
	}
}
