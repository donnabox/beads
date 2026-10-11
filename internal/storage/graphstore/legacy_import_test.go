//go:build cgo

package graphstore

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/legacyimport"
	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/types"
	publicops "github.com/steveyegge/beads/issueops"
)

func TestLegacyImportPrefixAdmission(t *testing.T) {
	for _, id := range []string{"web-app-123", "web-app-a1b", "web-app-abc", "web-app-123.1.2"} {
		t.Run(id, func(t *testing.T) {
			batch := legacyimport.Batch{Issues: []*types.Issue{{ID: id, Title: "Imported"}}}
			if _, err := prepareLegacyImport(batch, "importer"); err != nil {
				t.Fatal(err)
			}
			if got := legacyImportPrefix(id); got != "web-app" {
				t.Fatalf("prefix=%q", got)
			}
		})
	}
	for _, ids := range [][]string{{"web-app-123", "web-api-456"}, {"web-app-"}, {"missing"}, {"-123"}} {
		t.Run(strings.Join(ids, ","), func(t *testing.T) {
			var batch legacyimport.Batch
			for _, id := range ids {
				batch.Issues = append(batch.Issues, &types.Issue{ID: id, Title: "Imported"})
			}
			if _, err := prepareLegacyImport(batch, "importer"); err == nil {
				t.Fatal("invalid or mixed inferred prefixes admitted")
			}
		})
	}
	for _, ids := range [][]string{{"web-app-123", "web-app-abcdef"}, {"web-app-abcdef", "web-app-123"}} {
		var batch legacyimport.Batch
		for _, id := range ids {
			batch.Issues = append(batch.Issues, &types.Issue{ID: id, Title: "Imported"})
		}
		prepared, err := prepareLegacyImportWithPrefix(batch, "importer", "")
		if err != nil || prepared.prefix != "web-app" {
			t.Fatalf("whole-batch prefix %v = %q, %v", ids, prepared.prefix, err)
		}
	}
	ambiguous := legacyimport.Batch{Issues: []*types.Issue{{ID: "web-app-abcdef", Title: "Imported"}}}
	if _, err := prepareLegacyImport(ambiguous, "importer"); err == nil || !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("ambiguous prefix not refused: %v", err)
	}
	for _, prefix := range []string{"web-app", "web"} {
		prepared, err := prepareLegacyImportWithPrefix(ambiguous, "importer", prefix)
		if err != nil || prepared.prefix != prefix {
			t.Fatalf("explicit prefix %q = %q, %v", prefix, prepared.prefix, err)
		}
	}
	for _, prefix := range []string{"web-", " web-app", "web-app ", "other"} {
		if _, err := prepareLegacyImportWithPrefix(ambiguous, "importer", prefix); err == nil {
			t.Fatalf("invalid override %q admitted", prefix)
		}
	}
	if _, err := prepareLegacyImportWithPrefix(legacyimport.Batch{Memories: []legacyimport.Memory{{Key: "key"}}}, "importer", "web-app"); err == nil {
		t.Fatal("memory-only source prefix admitted")
	}
}

func TestLegacyImportLabelSetAndInternalState(t *testing.T) {
	issue := &types.Issue{ID: "old-a", Title: "Imported", Labels: []string{"one", "two", "one", "three"}}
	prepared, err := prepareLegacyImport(legacyimport.Batch{Issues: []*types.Issue{issue}}, "importer")
	if err != nil || !reflect.DeepEqual(prepared.Issues[0].Labels, []string{"one", "two", "three"}) || !reflect.DeepEqual(issue.Labels, []string{"one", "two", "one", "three"}) {
		t.Fatalf("label set or caller changed: %+v %+v %v", prepared, issue, err)
	}
	plane := false
	for field, internal := range map[string]*types.Issue{
		"WispPlaneOverride": {WispPlaneOverride: &plane}, "IDPrefix": {IDPrefix: "custom"}, "PrefixOverride": {PrefixOverride: "other"}, "SourceRepo": {SourceRepo: "another"}, "IsLitePartial": {IsLitePartial: true},
	} {
		tag, ok := reflect.TypeOf(types.Issue{}).FieldByName(field)
		if !ok || tag.Tag.Get("json") != "-" {
			t.Fatalf("internal field %s entered legacy JSON: %v", field, tag)
		}
		internal.ID, internal.Title = "old-a", "Imported"
		if _, err := prepareLegacyImport(legacyimport.Batch{Issues: []*types.Issue{internal}}, "importer"); !errors.Is(err, storage.ErrValidation) || !strings.Contains(err.Error(), field) {
			t.Fatalf("nonserialized state silently ignored: %+v", internal)
		}
	}
}

func TestLegacyImportCommentBudgetAndOriginalUTF8(t *testing.T) {
	issue := &types.Issue{ID: "old-a", Title: "Imported"}
	for i := range LegacyImportCommentLimit {
		issue.Comments = append(issue.Comments, &types.Comment{ID: fmt.Sprint(i), Text: "Comment"})
	}
	batch := legacyimport.Batch{Issues: []*types.Issue{issue}}
	if _, err := prepareLegacyImport(batch, "importer"); err != nil {
		t.Fatal(err)
	}
	issue.Comments = append(issue.Comments, &types.Comment{ID: "one-too-many"})
	if _, err := prepareLegacyImport(batch, "importer"); !errors.Is(err, ErrLimitExceeded) {
		t.Fatalf("comment ceiling: %v", err)
	}
	for _, invalid := range []*types.Issue{
		{ID: "old-a", Title: "bad\xff"},
		{ID: "old-a", Title: "Imported", Description: "bad\xff"},
		{ID: "old-a", Title: "Imported", Labels: []string{"bad\xff"}},
		{ID: "old-a", Title: "Imported", Comments: []*types.Comment{{Text: "bad\xff"}}},
		{ID: "old-a", Title: "Imported", Dependencies: []*types.Dependency{{DependsOnID: "bad\xff"}}},
	} {
		if _, err := prepareLegacyImport(legacyimport.Batch{Issues: []*types.Issue{invalid}}, "importer"); err == nil {
			t.Fatal("invalid original UTF-8 was normalized instead of refused")
		}
	}
}

func TestLegacyImportMemoryPaths(t *testing.T) {
	keys := []string{"safe", "old-a", "has space", "has%20space", "雪", "%", ".", "..", "a/b", "a//b", "back\\slash", "control\x00", " leading ", "Case", "case", strings.Repeat("long", 300)}
	// A direct key must retain its address even if it occupies a remap's
	// preferred hash path. Remap allocation must not depend on input order.
	digest := sha256.Sum256([]byte("."))
	reserved := fmt.Sprintf("imported-memory-%x", digest)
	keys = append(keys, reserved)
	batch := legacyimport.Batch{Issues: []*types.Issue{{ID: "old-a", Title: "Issue"}}}
	for _, key := range keys {
		batch.Memories = append(batch.Memories, legacyimport.Memory{Key: key, Value: ""})
	}
	prepared, err := prepareLegacyImport(batch, "importer")
	if err != nil {
		t.Fatal(err)
	}
	paths := legacyImportMemoryPaths(prepared)
	seen := map[string]bool{"beads/old-a": true}
	for _, key := range keys {
		path := paths[key]
		if err := validatePath(path); err != nil || seen[path] {
			t.Fatalf("key=%q path=%q duplicate=%v err=%v", key, path, seen[path], err)
		}
		seen[path] = true
	}
	for _, key := range []string{"safe", "has%20space", "a/b", "Case", "case", reserved} {
		if paths[key] != "beads/"+key {
			t.Fatalf("safe direct mapping changed: %q -> %q", key, paths[key])
		}
	}
	if paths["雪"] != "beads/"+graph.CanonicalSegment("雪") || paths["has space"] == paths["has%20space"] {
		t.Fatal("encoded keys collide or Unicode path is not canonical")
	}
	if paths["."] != "beads/"+reserved+"-1" {
		t.Fatalf("fallback displaced direct key: %q", paths["."])
	}
	for i, j := 0, len(prepared.Memories)-1; i < j; i, j = i+1, j-1 {
		prepared.Memories[i], prepared.Memories[j] = prepared.Memories[j], prepared.Memories[i]
	}
	if got := legacyImportMemoryPaths(prepared); !reflect.DeepEqual(paths, got) {
		t.Fatalf("order changed mappings: %v -> %v", paths, got)
	}
	batch.Memories = append(batch.Memories, legacyimport.Memory{Key: "safe", Value: "different"})
	if _, err := prepareLegacyImport(batch, "importer"); err == nil {
		t.Fatal("duplicate memory key admitted")
	}
	for _, memory := range []legacyimport.Memory{{Key: ""}, {Key: "bad\xff"}, {Key: "valid", Value: "bad\xff"}} {
		if _, err := prepareLegacyImport(legacyimport.Batch{Memories: []legacyimport.Memory{memory}}, "importer"); err == nil {
			t.Fatalf("invalid memory admitted: %+v", memory)
		}
	}
}

func TestLegacyImportCommentVerification(t *testing.T) {
	want := &types.Issue{ID: "old-a", Title: "Issue"}
	got := &types.Issue{ID: "old-a", Title: "Issue"}
	stamp := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := range 10000 {
		c := types.Comment{ID: fmt.Sprintf("comment-%d", i), IssueID: want.ID, Author: "author", Text: "text", CreatedAt: stamp}
		want.Comments = append(want.Comments, &c)
		copy := c
		got.Comments = append(got.Comments, &copy)
	}
	// Missing timestamps stay wildcards; native creation has already assigned
	// IDs before verification. Stored feed ordering is irrelevant to identity.
	want.Comments[0].CreatedAt = time.Time{}
	for i, j := 0, len(got.Comments)-1; i < j; i, j = i+1, j-1 {
		got.Comments[i], got.Comments[j] = got.Comments[j], got.Comments[i]
	}
	if err := verifyLegacyIssue(want, got); err != nil {
		t.Fatal(err)
	}
	got.Comments[0].Text = "changed"
	if err := verifyLegacyIssue(want, got); err == nil {
		t.Fatal("changed comment text admitted")
	}
}

// Benchmark helper verification separately from native persistence. Sizes up
// to the admitted ceiling make verification cost observable without timing
// assertions that depend on the test machine or database engine.
func BenchmarkLegacyImportCommentVerification(b *testing.B) {
	for _, count := range []int{10, 100, LegacyImportCommentLimit} {
		b.Run(fmt.Sprint(count), func(b *testing.B) {
			want := &types.Issue{ID: "old-a", Title: "Issue"}
			got := &types.Issue{ID: "old-a", Title: "Issue"}
			stamp := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
			for i := range count {
				c := types.Comment{ID: fmt.Sprint(i), IssueID: want.ID, Author: "author", Text: "text", CreatedAt: stamp}
				want.Comments = append(want.Comments, &c)
				copy := c
				got.Comments = append(got.Comments, &copy)
			}
			for i, j := 0, len(got.Comments)-1; i < j; i, j = i+1, j-1 {
				got.Comments[i], got.Comments[j] = got.Comments[j], got.Comments[i]
			}
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				if err := verifyLegacyIssue(want, got); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

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
	for _, table := range []string{"labels", "comments", "child_counters", "wisps", "wisp_dependencies", "wisp_comments", "wisp_labels", "config"} {
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

// Controlled contamination is confined to disposable test stores. Foreign-key
// checks are disabled on one borrowed connection only for orphan fixtures and
// restored before that connection returns to the pool.
func legacyImportContaminate(ctx context.Context, s *Store, query string, orphan bool) (err error) {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, conn.Close()) }()
	if orphan {
		if _, err := conn.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS=0"); err != nil {
			return err
		}
		defer func() {
			restore, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_, resetErr := conn.ExecContext(restore, "SET FOREIGN_KEY_CHECKS=1")
			if resetErr != nil {
				// A failed reset must not return a permissive session to the pool.
				discardErr := conn.Raw(func(any) error { return driver.ErrBadConn })
				resetErr = errors.Join(resetErr, discardErr)
			}
			err = errors.Join(err, resetErr)
		}()
	}
	_, err = conn.ExecContext(ctx, query)
	return err
}

func TestLegacyImportFreshDataPlanes(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			ctx, options := issueExperimentOptions(t, backend)
			s, err := OpenExisting(ctx, options)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := s.Close(); err != nil {
					t.Error(err)
				}
			})
			batch := parseLegacyFixture(t, legacyImportFixture)
			cases := []struct {
				name, insert, remove, message string
				orphan                        bool
			}{
				{"payload", "INSERT INTO graph_preview_payloads(path,properties) VALUES('beads/orphan','{}')", "DELETE FROM graph_preview_payloads WHERE path='beads/orphan'", "graph_preview_payloads", false},
				{"link", "INSERT INTO graph_preview_links(path,source_path,target_path,properties,attribution) VALUES('links/orphan','beads/absent','beads/absent','{}','{}')", "DELETE FROM graph_preview_links WHERE path='links/orphan'", "graph_preview_links", false},
				{"label", "INSERT INTO labels(issue_id,label) VALUES('absent','retained')", "DELETE FROM labels WHERE issue_id='absent'", "labels", true},
				{"wisp-dependency", "INSERT INTO wisp_dependencies(id,issue_id,depends_on_external) VALUES('orphan-dep','absent','external:target')", "DELETE FROM wisp_dependencies WHERE id='orphan-dep'", "wisp_dependencies", true},
				{"wisp-comment", "INSERT INTO wisp_comments(id,issue_id,text) VALUES('orphan-comment','absent','retained')", "DELETE FROM wisp_comments WHERE id='orphan-comment'", "wisp_comments", true},
				{"wisp-label", "INSERT INTO wisp_labels(issue_id,label) VALUES('absent','retained')", "DELETE FROM wisp_labels WHERE issue_id='absent'", "wisp_labels", true},
				{"kv-memory", "INSERT INTO config (`key`,value) VALUES('kv.memory.retained','original')", "DELETE FROM config WHERE `key`='kv.memory.retained'", "legacy KV Memories", false},
				{"empty-kv-memory", "INSERT INTO config (`key`,value) VALUES('kv.memory.retained','')", "DELETE FROM config WHERE `key`='kv.memory.retained'", "legacy KV Memories", false},
			}
			for _, test := range cases {
				t.Run(test.name, func(t *testing.T) {
					if err := legacyImportContaminate(ctx, s, test.insert, test.orphan); err != nil {
						t.Fatal(err)
					}
					before := legacyImportState(t, ctx, s)
					for _, dry := range []bool{true, false} {
						t.Run(fmt.Sprintf("dry=%t", dry), func(t *testing.T) {
							got, err := s.ImportLegacy(ctx, batch, "importer", dry)
							if !errors.Is(err, storage.ErrValidation) || !strings.Contains(err.Error(), test.message) || !reflect.DeepEqual(got, LegacyImportResult{}) {
								t.Fatalf("occupied plane admitted or result leaked: %+v %v", got, err)
							}
							if !reflect.DeepEqual(before, legacyImportState(t, ctx, s)) {
								t.Fatal("refusal changed existing rows, prefix or writer token")
							}
						})
					}
					if _, err := s.db.ExecContext(ctx, test.remove); err != nil {
						t.Fatal(err)
					}
				})
			}
			// Similar configuration keys are not Memories. Ordinary config does
			// not disqualify the fresh destination or get overwritten by dry-run.
			if _, err := s.db.ExecContext(ctx, "INSERT INTO config (`key`,value) VALUES('kv.memoryx.retained','ordinary config'),('iris.test.setting','kept')"); err != nil {
				t.Fatal(err)
			}
			before := legacyImportState(t, ctx, s)
			if _, err := s.ImportLegacy(ctx, batch, "importer", true); err != nil {
				t.Fatalf("ordinary configuration disqualified fresh destination: %v", err)
			}
			if !reflect.DeepEqual(before, legacyImportState(t, ctx, s)) {
				t.Fatal("dry-run changed ordinary configuration")
			}
		})
	}
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
			// A matching prefix update may report zero changed rows. It still
			// validates and dry-run must preserve the configuration.
			samePrefix := strings.ReplaceAll(legacyImportFixture, "old-", before["issue_prefix"]+"-")
			if _, err := s.ImportLegacy(ctx, parseLegacyFixture(t, samePrefix), "importer", true); err != nil {
				t.Fatalf("same-prefix dry-run: %v", err)
			}
			if !reflect.DeepEqual(before, legacyImportState(t, ctx, s)) {
				t.Fatal("same-prefix dry-run changed configuration")
			}
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
			for _, deadline := range []bool{false, true} {
				t.Run(fmt.Sprintf("late-context-deadline=%t", deadline), func(t *testing.T) {
					var operation context.Context
					var cancel context.CancelFunc
					var expected error
					if deadline {
						operation, cancel = context.WithTimeout(ctx, 15*time.Second)
						expected = context.DeadlineExceeded
					} else {
						operation, cancel = context.WithCancel(ctx)
						expected = context.Canceled
					}
					defer cancel()
					reached := false
					s.afterWrite = func(stage string) error {
						if stage == "import-retained" {
							reached = true
							if deadline {
								<-operation.Done()
							} else {
								cancel()
							}
						}
						return nil
					}
					_, err := s.ImportLegacy(operation, batch, "importer", false)
					s.afterWrite = nil
					// Inspect through the independent live context, after driver
					// cancellation and transaction cleanup have completed.
					// A driver watcher may already have rolled back the transaction.
					// ErrTxDone cleanup must keep outcome_unknown precedence;
					// other driver cleanup errors may also be uncertain.
					unknown := errors.Is(err, ErrOutcomeUnknown)
					alreadyDone := errors.Is(err, sql.ErrTxDone)
					if !reached || !errors.Is(err, expected) || (alreadyDone && !unknown) || !reflect.DeepEqual(before, legacyImportState(t, ctx, s)) {
						t.Fatalf("late context classification or rollback failed: reached=%v unknown=%v alreadyDone=%v err=%v", reached, unknown, alreadyDone, err)
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
// graph projection before returning. Exercise each family's rollback first,
// then apply all families together and compare current/retained reads on reopen.
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
			var names []string
			for name := range families {
				names = append(names, name)
			}
			sort.Strings(names)
			var input strings.Builder
			for i, name := range names {
				fmt.Fprintf(&input, `{"id":"fields-%d","title":"%s","created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-02T00:00:00Z",%s}`+"\n", i+1, name, families[name])
			}
			if _, err := s.ImportLegacy(ctx, parseLegacyFixture(t, input.String()), "importer", false); err != nil {
				t.Fatal(err)
			}
			if err := s.Close(); err != nil {
				t.Fatal(err)
			}
			s, err = OpenExisting(ctx, o)
			if err != nil {
				t.Fatal(err)
			}
			for i, name := range names {
				path := fmt.Sprintf("beads/fields-%d", i+1)
				current, err := s.ShowIssue(ctx, path)
				if err != nil {
					t.Fatal(err)
				}
				payload, err := json.Marshal(current.Properties)
				if err != nil {
					t.Fatal(err)
				}
				var got, want map[string]json.RawMessage
				if err := json.Unmarshal(payload, &got); err != nil {
					t.Fatal(err)
				}
				if err := json.Unmarshal([]byte("{"+families[name]+"}"), &want); err != nil {
					t.Fatal(err)
				}
				for field, value := range want {
					if string(got[field]) != string(value) {
						t.Errorf("%s field %s=%s, want %s", name, field, got[field], value)
					}
				}
				_, versions, err := s.Versions(ctx, path)
				if err != nil || len(versions) != 1 {
					t.Fatalf("%s initial versions=%v err=%v", name, versions, err)
				}
				retained, err := s.ReadVersion(ctx, path, versions[0].Version)
				if err != nil {
					t.Fatal(err)
				}
				initial, ok := retained.(IssueRecord)
				if !ok {
					t.Fatalf("%s retained type=%T", name, retained)
				}
				// Native row locks and content digests are deliberately excluded
				// from retained Issue history; compare only the durable payload.
				current.Properties.RowVersion, current.Properties.ContentHash = 0, ""
				if !reflect.DeepEqual(current, initial) {
					t.Fatalf("%s current/initial retained mismatch: %v", name, err)
				}
			}
		})
	}
}

// Exercise native comment normalization and persistence, not just the verifier.
func TestLegacyImportCommentFeedPersistence(t *testing.T) {
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
			stamp := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
			before := legacyImportState(t, ctx, s)
			collapsed := &types.Issue{ID: "comments-a1b", Title: "Distinct source comments", Comments: []*types.Comment{
				{ID: "first", Author: "author", Text: "same", CreatedAt: stamp},
				{ID: "second", Author: "author", Text: "same", CreatedAt: stamp},
			}}
			if _, err := s.ImportLegacy(ctx, legacyimport.Batch{Issues: []*types.Issue{collapsed}}, "importer", false); err == nil {
				t.Fatal("native duplicate-content collapse was silently accepted")
			}
			if !reflect.DeepEqual(before, legacyImportState(t, ctx, s)) {
				t.Fatal("late comment refusal changed the database")
			}
			issue := &types.Issue{ID: "comments-a1b", Title: "Comment-heavy import"}
			for i := range 80 {
				comment := &types.Comment{Author: "author", Text: fmt.Sprintf("Comment %d", i)}
				if i%2 == 0 {
					comment.ID = fmt.Sprintf("source-%d", i)
					comment.CreatedAt = stamp.Add(time.Duration(i) * time.Second)
				}
				issue.Comments = append(issue.Comments, comment)
			}
			result, err := s.ImportLegacy(ctx, legacyimport.Batch{Issues: []*types.Issue{issue}}, "importer", false)
			if err != nil || result.Comments != 80 {
				t.Fatalf("comment-heavy import: %+v %v", result, err)
			}
			if err := s.Close(); err != nil {
				t.Fatal(err)
			}
			s, err = OpenExisting(ctx, o)
			if err != nil {
				t.Fatal(err)
			}
			feed, err := s.IssueComments(ctx, "beads/comments-a1b")
			if err != nil || len(feed) != 80 {
				t.Fatalf("reopened comment feed: %d %v", len(feed), err)
			}
			byText, ids := map[string]*types.Comment{}, map[string]bool{}
			for _, comment := range feed {
				if comment.ID == "" || ids[comment.ID] || comment.IssueID != issue.ID || comment.Author != "author" || comment.CreatedAt.IsZero() {
					t.Fatalf("invalid normalized comment: %+v", comment)
				}
				ids[comment.ID] = true
				byText[comment.Text] = comment
			}
			for _, original := range issue.Comments {
				stored := byText[original.Text]
				if stored == nil || (original.ID != "" && stored.ID != original.ID) || (!original.CreatedAt.IsZero() && !original.CreatedAt.Equal(stored.CreatedAt)) {
					t.Fatalf("source comment changed: %+v -> %+v", original, stored)
				}
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
