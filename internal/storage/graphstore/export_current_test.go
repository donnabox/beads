//go:build cgo

package graphstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/types"
)

func TestExportCurrentSnapshotLifecycle(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			ctx, options, s := memoryDeleteStore(t, backend)
			empty, err := s.ExportCurrentSnapshot(ctx)
			if err != nil || len(empty.Snapshot.Records) != 0 || len(empty.Snapshot.Types) != 6 || empty.IssuePrefix != "exp" || empty.DeletedAllocations != 0 || len(empty.CommentsByIssuePath) != 0 {
				t.Fatalf("empty export: %+v %v", empty, err)
			}
			if err := exportSetPrefix(ctx, s, "web-app-"); err != nil {
				t.Fatal(err)
			}
			memory, err := s.Create(ctx, CreateRequest{Path: "beads/plan", Title: " Plan 記憶 ", Body: "\t🌱\n", Metadata: json.RawMessage(`{"source":"disposable"}`)})
			if err != nil {
				t.Fatal(err)
			}
			work, err := s.CreateIssue(ctx, "beads/work", plainIssue("Work"))
			if err != nil || work.Properties == nil || !strings.HasPrefix(work.Properties.ID, "web-app-") {
				t.Fatalf("hyphenated native prefix: %+v %v", work, err)
			}
			if _, err := s.CreateIssue(ctx, "beads/prereq", plainIssue("Prerequisite")); err != nil {
				t.Fatal(err)
			}
			if _, err := s.AddDependency(ctx, DependencyRequest{Path: "links/block", SourcePath: "beads/work", TargetPath: "beads/prereq"}); err != nil {
				t.Fatal(err)
			}
			if _, err := s.AddInformationalLink(ctx, LinkCreateRequest{Path: "links/context", SourcePath: "beads/plan", TargetPath: "beads/work", ExpectedSourceRevision: memory.Revision}); err != nil {
				t.Fatal(err)
			}
			if _, err := s.AddInformationalLink(ctx, LinkCreateRequest{Path: "links/issue-context", SourcePath: "beads/prereq", TargetPath: "beads/plan"}); err != nil {
				t.Fatal(err)
			}
			for _, text := range []string{"\toriginal 記憶 🌱\n", ""} {
				if err := s.AddIssueComment(ctx, "beads/work", " author \n", text); err != nil {
					t.Fatal(err)
				}
			}
			// Equal timestamps exercise the original comment-ID tie breaker.
			if err := s.withTx(ctx, true, func(tx *sql.Tx) error {
				_, err := tx.ExecContext(ctx, "UPDATE comments SET created_at='2026-10-10 12:34:56' WHERE issue_id=?", work.Properties.ID)
				return err
			}); err != nil {
				t.Fatal(err)
			}
			native, err := s.IssueComments(ctx, "beads/work")
			if err != nil || len(native) != 2 {
				t.Fatalf("native comments: %+v %v", native, err)
			}
			wantComments := []types.Comment{*native[0], *native[1]}
			before := exportSourceState(t, ctx, s)
			current, err := s.ExportCurrentSnapshot(ctx)
			if err != nil || current.IssuePrefix != "web-app-" || current.DeletedAllocations != 0 || len(current.Snapshot.Records) != 6 || !reflect.DeepEqual(current.CommentsByIssuePath["beads/work"], wantComments) {
				t.Fatalf("complete current export: %+v %v", current, err)
			}
			if comments, present := current.CommentsByIssuePath["beads/prereq"]; !present || comments == nil || len(comments) != 0 {
				t.Fatal("Issue with no comments was absent or null")
			}
			ordinary, err := s.CurrentSnapshot(ctx)
			if err != nil || !reflect.DeepEqual(ordinary, current.Snapshot) {
				t.Fatalf("existing projection changed: %v", err)
			}
			wantSnapshot, err := canonicalJSON(ordinary)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(before, exportSourceState(t, ctx, s)) {
				t.Fatal("read changed source rows")
			}
			// Returned feeds are detached: caller mutations cannot affect records
			// or the next acquisition and survive neither close nor reopen.
			current.CommentsByIssuePath["beads/work"][0].Text = "caller mutation"
			delete(current.CommentsByIssuePath, "beads/prereq")
			mutatedIssue, mutatedOwned := false, false
			for _, raw := range current.Snapshot.Records {
				if issue, ok := raw.(IssueRecord); ok {
					issue.Properties.Title = "caller mutation"
					mutatedIssue = true
					if len(issue.Owned) > 0 && len(issue.Owned[0]) > 0 {
						issue.Owned[0][0] = '!'
						mutatedOwned = true
					}
				}
			}
			if !mutatedIssue || !mutatedOwned {
				t.Fatal("fixture lacks nested Issue properties or owned Link bytes")
			}
			fresh, err := s.ExportCurrentSnapshot(ctx)
			freshSnapshot, snapshotErr := canonicalJSON(fresh.Snapshot)
			if err != nil || snapshotErr != nil || !reflect.DeepEqual(fresh.CommentsByIssuePath["beads/work"], wantComments) || string(freshSnapshot) != string(wantSnapshot) {
				t.Fatalf("same-Store detachment: %+v %v", fresh, err)
			}
			if !reflect.DeepEqual(before, exportSourceState(t, ctx, s)) {
				t.Fatal("same-Store reacquisition changed source rows")
			}
			if comments, present := fresh.CommentsByIssuePath["beads/prereq"]; !present || comments == nil || len(comments) != 0 {
				t.Fatal("caller deletion affected the next acquisition")
			}
			if err := s.Close(); err != nil {
				t.Fatal(err)
			}
			s, err = OpenExisting(ctx, options)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := s.Close(); err != nil {
					t.Error(err)
				}
			})
			reopened, err := s.ExportCurrentSnapshot(ctx)
			reopenedSnapshot, snapshotErr := canonicalJSON(reopened.Snapshot)
			if err != nil || snapshotErr != nil || !reflect.DeepEqual(reopened.CommentsByIssuePath["beads/work"], wantComments) || string(reopenedSnapshot) != string(wantSnapshot) || reopened.IssuePrefix != "web-app-" {
				t.Fatalf("reopen preservation: %+v %v", reopened, err)
			}
			if !reflect.DeepEqual(before, exportSourceState(t, ctx, s)) {
				t.Fatal("reopened acquisition changed source rows")
			}
			// Native permanent comments cascade on Issue deletion. The retained
			// Issue History remains outside this current-only acquisition.
			removed, err := s.CreateIssue(ctx, "beads/removed", plainIssue("Removed"))
			if err != nil {
				t.Fatal(err)
			}
			if err := s.AddIssueComment(ctx, "beads/removed", "author", "current feed before deletion"); err != nil {
				t.Fatal(err)
			}
			if _, err := s.DeleteIssue(ctx, IssueDeleteRequest{Path: "beads/removed", ExpectedRevision: removed.Revision}); err != nil {
				t.Fatal(err)
			}
			deleted, err := s.ExportCurrentSnapshot(ctx)
			if err != nil || deleted.DeletedAllocations != 1 || len(deleted.Snapshot.Records) != 6 {
				t.Fatalf("deleted allocation disclosure: %+v %v", deleted, err)
			}
			if _, present := deleted.CommentsByIssuePath["beads/removed"]; present {
				t.Fatal("deleted Issue acquired as a current comment owner")
			}
			var surviving uint64
			if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM comments WHERE issue_id=?", removed.Properties.ID).Scan(&surviving); err != nil || surviving != 0 {
				t.Fatalf("native comment deletion boundary: %d %v", surviving, err)
			}
		})
	}
}

func TestExportCurrentSnapshotFailures(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			ctx, _, s := memoryDeleteStore(t, backend)
			work, err := s.CreateIssue(ctx, "beads/work", plainIssue("Work"))
			if err != nil {
				t.Fatal(err)
			}
			if err := s.AddIssueComment(ctx, "beads/work", "author", "original"); err != nil {
				t.Fatal(err)
			}
			before := exportSourceState(t, ctx, s)
			// These are callback faults, not injected driver cleanup failures.
			for _, stage := range []string{"export-preflight", "export-inventory", "export-comments"} {
				t.Run(stage, func(t *testing.T) {
					fault := errors.New("injected export callback failure")
					s.afterWrite = func(at string) error {
						if at == stage {
							return fault
						}
						return nil
					}
					got, err := s.ExportCurrentSnapshot(ctx)
					s.afterWrite = nil
					if !errors.Is(err, fault) || !reflect.DeepEqual(got, ExportCurrentState{}) || !reflect.DeepEqual(before, exportSourceState(t, ctx, s)) {
						t.Fatalf("callback failure changed/exposed state: %+v %v", got, err)
					}
				})
			}
			canceled, cancel := context.WithCancel(ctx)
			fault := errors.New("hydration error text without a cancellation chain")
			s.afterWrite = func(stage string) error {
				if stage == "export-inventory" {
					cancel()
					return fault
				}
				return nil
			}
			got, err := s.ExportCurrentSnapshot(canceled)
			s.afterWrite = nil
			cancel()
			if !errors.Is(err, context.Canceled) || !errors.Is(err, fault) || !reflect.DeepEqual(got, ExportCurrentState{}) {
				t.Fatalf("lost cancellation cause: %+v %v", got, err)
			}
			cases := []struct{ name, query, message string }{
				{"unmapped-comment-owner", "DELETE FROM graph_preview_catalog WHERE path='beads/work'", "comment without a live graph Issue"},
				{"deleted-comment-owner", "UPDATE graph_preview_catalog SET allocation_state='deleted' WHERE path='beads/work'", "comment without a live graph Issue"},
				{"legacy-memory-even-empty", "INSERT INTO config (`key`,value) VALUES ('kv.memory.old','')", "legacy KV Memories"},
				{"wisps", "INSERT INTO wisps (id,title) VALUES ('exp-wisp','disposable')", "wisps"},
				{"wisp-comments", "INSERT INTO wisp_comments (id,issue_id,text) VALUES ('wisp-comment','orphan','disposable')", "wisp_comments"},
				{"wisp-labels", "INSERT INTO wisp_labels (issue_id,label) VALUES ('orphan','disposable')", "wisp_labels"},
			}
			for _, test := range cases {
				t.Run(test.name, func(t *testing.T) {
					// Temporary corrupt/alternative rows never leave this rollback.
					err := s.withTx(ctx, false, func(tx *sql.Tx) error {
						if _, err := tx.ExecContext(ctx, test.query); err != nil {
							return err
						}
						got, err := s.exportCurrentSnapshotInTx(ctx, tx)
						if !errors.Is(err, ErrInvalidStore) || !strings.Contains(err.Error(), test.message) || !reflect.DeepEqual(got, ExportCurrentState{}) {
							return fmt.Errorf("refusal exposed state: %+v %v", got, err)
						}
						return nil
					})
					if err != nil || !reflect.DeepEqual(before, exportSourceState(t, ctx, s)) {
						t.Fatalf("refusal/rollback: %v", err)
					}
				})
			}
			// The scan path has a second owner check, independently of preflight.
			if err := s.withTx(ctx, false, func(tx *sql.Tx) error {
				comments, err := exportCommentsInTx(ctx, tx, map[string]string{}, 1)
				if !errors.Is(err, ErrInvalidStore) || comments != nil {
					return fmt.Errorf("scan owner refusal: %+v %v", comments, err)
				}
				comments, err = exportCommentsInTx(ctx, tx, map[string]string{work.Properties.ID: "beads/work"}, 0)
				if !errors.Is(err, ErrInvalidStore) || comments != nil {
					return fmt.Errorf("scan count refusal: %+v %v", comments, err)
				}
				return nil
			}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestExportCurrentSnapshotAcquisitionBounds(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			ctx, _, s := memoryDeleteStore(t, backend)
			work, err := s.CreateIssue(ctx, "beads/work", plainIssue("Work"))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := s.Create(ctx, CreateRequest{Path: "beads/filler", Body: strings.Repeat("m", 12<<10)}); err != nil {
				t.Fatal(err)
			}
			const byteLimit = 64 << 10
			if err := s.withTx(ctx, false, func(tx *sql.Tx) error {
				_, err := checkExportReadBoundsAtLimit(ctx, tx, exportCommentLimit, byteLimit)
				return err
			}); err != nil {
				t.Fatalf("baseline must fit small test budget: %v", err)
			}
			// Each comment fits the small budget by itself. Its aggregate with
			// the complete record acquisition does not; the feed is never read.
			if err := s.AddIssueComment(ctx, "beads/work", "author", strings.Repeat("c", 40<<10)); err != nil {
				t.Fatal(err)
			}
			before := exportSourceState(t, ctx, s)
			if err := s.withTx(ctx, false, func(tx *sql.Tx) error {
				if _, err := checkExportReadBoundsAtLimit(ctx, tx, exportCommentLimit, byteLimit); !errors.Is(err, ErrLimitExceeded) {
					return fmt.Errorf("aggregate comment bytes not bounded: %v", err)
				}
				if _, err := checkExportReadBoundsAtLimit(ctx, tx, 0, PreviewCurrentReadByteLimit); !errors.Is(err, ErrLimitExceeded) {
					return fmt.Errorf("comment count not bounded before acquisition: %v", err)
				}
				if _, err := tx.ExecContext(ctx, "DELETE FROM graph_preview_catalog WHERE backing_key=?", work.Properties.ID); err != nil {
					return err
				}
				if _, err := checkExportReadBoundsAtLimit(ctx, tx, exportCommentLimit, byteLimit); !errors.Is(err, ErrLimitExceeded) {
					return fmt.Errorf("orphan comment escaped byte budget: %v", err)
				}
				return nil
			}); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(before, exportSourceState(t, ctx, s)) {
				t.Fatal("bounded refusal changed state")
			}
			const deletedHeadByteLimit = 40 << 10
			if err := s.withTx(ctx, false, func(tx *sql.Tx) error {
				if _, err := tx.ExecContext(ctx, "DELETE FROM comments"); err != nil {
					return err
				}
				_, err := checkExportReadBoundsAtLimit(ctx, tx, exportCommentLimit, deletedHeadByteLimit)
				return err
			}); err != nil {
				t.Fatalf("baseline must fit deleted-head test budget: %v", err)
			}
			// Deleted Issue validation decodes its final retained head, despite
			// excluding it from live output; its acquisition also has a budget.
			request := plainIssue("Final retained head")
			request.Issue.Description = strings.Repeat("d", 40<<10)
			deleted, err := s.CreateIssue(ctx, "beads/removed", request)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := s.DeleteIssue(ctx, IssueDeleteRequest{Path: "beads/removed", ExpectedRevision: deleted.Revision}); err != nil {
				t.Fatal(err)
			}
			if err := s.withTx(ctx, false, func(tx *sql.Tx) error {
				// Isolate this admission check from unrelated live filler/feed.
				if _, err := tx.ExecContext(ctx, "DELETE FROM comments"); err != nil {
					return err
				}
				if _, err := tx.ExecContext(ctx, "DELETE FROM graph_preview_payloads WHERE path='beads/filler'"); err != nil {
					return err
				}
				if _, err := tx.ExecContext(ctx, "UPDATE graph_preview_catalog SET allocation_state='deleted' WHERE path='beads/filler'"); err != nil {
					return err
				}
				if _, err := checkExportReadBoundsAtLimit(ctx, tx, exportCommentLimit, deletedHeadByteLimit); !errors.Is(err, ErrLimitExceeded) {
					return fmt.Errorf("deleted final head escaped aggregate budget: %v", err)
				}
				return nil
			}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestExportCurrentSnapshotConcurrentWriter(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			base, options, reader := memoryDeleteStore(t, backend)
			ctx, cancel := context.WithTimeout(base, time.Minute)
			defer cancel()
			memory, err := reader.Create(ctx, CreateRequest{Path: "beads/source", Body: "original"})
			if err != nil {
				t.Fatal(err)
			}
			drop, err := reader.Create(ctx, CreateRequest{Path: "beads/drop"})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := reader.CreateIssue(ctx, "beads/work", plainIssue("Work")); err != nil {
				t.Fatal(err)
			}
			before, err := reader.ExportCurrentSnapshot(ctx)
			if err != nil {
				t.Fatal(err)
			}
			// Embedded uses its supported serialized connector. Server must
			// permit the controlled writer to commit while this read is paused.
			writer := &Store{db: reader.db, options: options}
			if backend == "server" {
				writer, err = OpenExisting(ctx, options)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					if err := writer.Close(); err != nil {
						t.Error(err)
					}
				})
			}
			reached, release := make(chan struct{}, 1), make(chan struct{})
			var releaseOnce sync.Once
			reader.afterWrite = func(stage string) error {
				if stage != "export-inventory" {
					return nil
				}
				reached <- struct{}{}
				select {
				case <-release:
					return nil
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			type outcome struct {
				state ExportCurrentState
				err   error
			}
			readDone := make(chan outcome, 1)
			writeDone := make(chan error, 1)
			started := make(chan struct{})
			var workers sync.WaitGroup
			defer func() {
				cancel()
				releaseOnce.Do(func() { close(release) })
				workers.Wait()
				reader.afterWrite = nil
			}()
			workers.Add(1)
			go func() {
				defer workers.Done()
				state, err := reader.ExportCurrentSnapshot(ctx)
				readDone <- outcome{state, err}
			}()
			select {
			case <-reached:
			case result := <-readDone:
				t.Fatalf("reader exited before overlap: %+v", result)
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			}
			workers.Add(1)
			go func() {
				defer workers.Done()
				close(started)
				if err := writer.AddIssueComment(ctx, "beads/work", "writer", "new comment"); err != nil {
					writeDone <- err
					return
				}
				if _, err := writer.AddInformationalLink(ctx, LinkCreateRequest{Path: "links/new", SourcePath: "beads/source", TargetPath: "beads/work", ExpectedSourceRevision: memory.Revision}); err != nil {
					writeDone <- err
					return
				}
				if _, err := writer.DeleteMemory(ctx, MemoryDeleteRequest{Path: "beads/drop", ExpectedRevision: drop.Revision}); err != nil {
					writeDone <- err
					return
				}
				writeDone <- exportSetPrefix(ctx, writer, "web-app-")
			}()
			<-started
			if backend == "server" {
				select {
				case err := <-writeDone:
					if err != nil {
						t.Fatalf("writer failed before reader release: %v", err)
					}
				case <-ctx.Done():
					t.Fatalf("writer did not commit during forced overlap: %v", ctx.Err())
				}
			}
			releaseOnce.Do(func() { close(release) })
			select {
			case result := <-readDone:
				if result.err != nil || !reflect.DeepEqual(result.state, before) {
					t.Fatalf("torn current-state acquisition: %+v %v", result.state, result.err)
				}
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			}
			if backend == "embedded" {
				select {
				case err := <-writeDone:
					if err != nil {
						t.Fatal(err)
					}
				case <-ctx.Done():
					t.Fatal(ctx.Err())
				}
			}
			workers.Wait()
			reader.afterWrite = nil
			after, err := reader.ExportCurrentSnapshot(ctx)
			if err != nil || after.Snapshot.WriterToken == before.Snapshot.WriterToken || after.IssuePrefix != "web-app-" || after.DeletedAllocations != 1 || len(after.CommentsByIssuePath["beads/work"]) != 1 {
				t.Fatalf("next acquisition did not observe complete writer state: %+v %v", after, err)
			}
			var owned int
			for _, record := range after.Snapshot.Records {
				if memory, ok := record.(Record); ok && memory.ID == options.Binding.ScopeURL+"beads/source" {
					owned = len(memory.Owned)
				}
			}
			if owned != 1 {
				t.Fatalf("new owner/Link state missing: %d", owned)
			}
		})
	}
}

func exportSetPrefix(ctx context.Context, s *Store, prefix string) error {
	return s.withTx(ctx, true, func(tx *sql.Tx) error {
		if err := s.touchCoordination(ctx, tx); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, "UPDATE config SET value=? WHERE `key`='issue_prefix'", prefix)
		return err
	})
}

func exportSourceState(t *testing.T, ctx context.Context, s *Store) map[string]string {
	t.Helper()
	state := workflowState(t, ctx, s)
	for _, table := range []string{"comments", "config", "graph_preview_types", "labels", "wisps", "wisp_comments", "wisp_labels", "wisp_dependencies"} {
		rows, err := s.db.QueryContext(ctx, "SELECT * FROM "+table)
		if err != nil {
			t.Fatal(err)
		}
		columns, err := rows.Columns()
		if err != nil {
			t.Fatal(errors.Join(err, rows.Close()))
		}
		var values []string
		for rows.Next() {
			fields := make([]any, len(columns))
			pointers := make([]any, len(columns))
			for i := range fields {
				pointers[i] = &fields[i]
			}
			if err := rows.Scan(pointers...); err != nil {
				t.Fatal(errors.Join(err, rows.Close()))
			}
			raw, err := json.Marshal(fields)
			if err != nil {
				t.Fatal(errors.Join(err, rows.Close()))
			}
			values = append(values, string(raw))
		}
		if err := errors.Join(rows.Err(), rows.Close()); err != nil {
			t.Fatal(err)
		}
		sort.Strings(values)
		state[table] = strings.Join(values, "\n")
	}
	return state
}
