//go:build cgo

package graphstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	graph "github.com/steveyegge/beads/graphops"
)

func TestReadVersionRetainedLifecycle(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			ctx, o := issueExperimentOptions(t, backend)
			s, err := OpenExisting(ctx, o)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := s.Close(); err != nil {
					t.Error(err)
				}
			}()
			memory, err := s.Create(ctx, CreateRequest{Path: "beads/plan", Body: "Original — 雪", Actor: "author"})
			if err != nil {
				t.Fatal(err)
			}
			work, err := s.CreateIssue(ctx, "beads/work", plainIssue("Work"))
			if err != nil {
				t.Fatal(err)
			}
			prereq, err := s.CreateIssue(ctx, "beads/prereq", plainIssue("Prerequisite"))
			if err != nil {
				t.Fatal(err)
			}
			dep, err := s.AddDependency(ctx, DependencyRequest{Path: "links/block", SourcePath: "beads/work", TargetPath: "beads/prereq", Actor: "author"})
			if err != nil {
				t.Fatal(err)
			}
			closed, err := s.CloseIssue(ctx, "beads/prereq", "Done", "author")
			if err != nil {
				t.Fatal(err)
			}
			added, err := s.AddInformationalLink(ctx, LinkCreateRequest{Path: "links/context", SourcePath: "beads/plan", TargetPath: "beads/work", ExpectedSourceRevision: memory.Revision, Properties: map[string]any{"note": "before"}, Actor: "author"})
			if err != nil {
				t.Fatal(err)
			}
			updated, err := s.UpdateLink(ctx, LinkUpdateRequest{Path: "links/context", ExpectedRevision: added.Link.Revision, ExpectedSourceRevision: added.Source.(Record).Revision, Properties: map[string]any{"note": "after"}, Actor: "editor"})
			if err != nil {
				t.Fatal(err)
			}
			removed, err := s.Unlink(ctx, LinkDeleteRequest{Path: "links/context", ExpectedRevision: updated.Link.Revision, ExpectedSourceRevision: updated.Source.(Record).Revision, Actor: "remover"})
			if err != nil {
				t.Fatal(err)
			}
			cases := []struct {
				path, version string
				want          any
			}{
				{"beads/plan", memory.Revision, memory},
				{"beads/plan", added.Source.(Record).Revision, added.Source},
				{"beads/plan", updated.Source.(Record).Revision, updated.Source},
				{"beads/plan", removed.Source.(Record).Revision, removed.Source},
				{"links/context", added.Link.Revision, added.Link},
				{"links/context", updated.Link.Revision, updated.Link},
				{"beads/work", work.Revision, work},
				{"beads/work", dep.Source.Revision, dep.Source},
				{"beads/prereq", prereq.Revision, prereq},
				{"beads/prereq", closed.Issue.Revision, closed.Issue},
				{"links/block", dep.Link.Revision, dep.Link},
			}
			var tokenBefore, tokenAfter string
			if err := s.db.QueryRowContext(ctx, `SELECT writer_token FROM graph_preview_scope WHERE singleton=1`).Scan(&tokenBefore); err != nil {
				t.Fatal(err)
			}
			for _, pass := range []string{"before-reopen", "after-reopen"} {
				for _, tc := range cases {
					want := tc.want
					if issue, ok := want.(IssueRecord); ok {
						// Current hydration includes these two json:"-" fields.
						// Jim's durable_state intentionally retains neither the
						// computed content hash nor the current row-lock token.
						properties := *issue.Properties
						properties.ContentHash, properties.RowVersion = "", 0
						issue.Properties = &properties
						want = issue
					}
					got, err := s.ReadVersion(ctx, tc.path, tc.version)
					if err != nil || !reflect.DeepEqual(got, want) {
						t.Fatalf("%s %s/%s: %v\ngot=%+v\nwant=%+v", pass, tc.path, tc.version, err, got, want)
					}
				}
				if got, err := s.ReadVersion(ctx, "links/context", removed.Link.Revision); !errors.Is(err, ErrGone) || got != nil {
					t.Fatalf("private tombstone treated as Resource: %+v %v", got, err)
				}
				if got, err := s.ReadVersion(ctx, "beads/plan", added.Link.Revision); !errors.Is(err, ErrVersionUnknown) || got != nil {
					t.Fatalf("cross-subject token selected a record: %+v %v", got, err)
				}
				if err := s.Close(); err != nil {
					t.Fatal(err)
				}
				s, err = OpenExisting(ctx, o)
				if err != nil {
					t.Fatal(err)
				}
			}
			if err := s.db.QueryRowContext(ctx, `SELECT writer_token FROM graph_preview_scope WHERE singleton=1`).Scan(&tokenAfter); err != nil {
				t.Fatal(err)
			}
			if tokenBefore != tokenAfter {
				t.Fatal("exact retained reads changed writer state")
			}
			// Historic reads do not hydrate current payload or current owned sets.
			if _, err := s.db.ExecContext(ctx, `UPDATE graph_preview_payloads SET properties='{}' WHERE path='beads/plan'`); err != nil {
				t.Fatal(err)
			}
			got, err := s.ReadVersion(ctx, "beads/plan", added.Source.(Record).Revision)
			if err != nil || !reflect.DeepEqual(got, added.Source) {
				t.Fatalf("current payload contaminated exact history: %+v %v", got, err)
			}
		})
	}
}

func TestReadVersionRefusals(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			ctx, o := issueExperimentOptions(t, backend)
			s, err := OpenExisting(ctx, o)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := s.Close(); err != nil {
					t.Error(err)
				}
			}()
			memory, err := s.Create(ctx, CreateRequest{Path: "beads/plan", Body: "Plan"})
			if err != nil {
				t.Fatal(err)
			}
			for _, token := range []string{"unknown", "a+b / 雪?%", strings.Repeat("x", PreviewVersionTokenLimit)} {
				if got, err := s.ReadVersion(ctx, "beads/plan", token); got != nil || !errors.Is(err, ErrVersionUnknown) {
					t.Fatalf("opaque token %q: %+v %v", token, got, err)
				}
			}
			for _, token := range []string{"", string([]byte{0xff}), strings.Repeat("x", PreviewVersionTokenLimit+1)} {
				if got, err := s.ReadVersion(ctx, "beads/plan", token); got != nil || !errors.Is(err, graph.ErrValidation) {
					t.Fatalf("invalid token accepted: %+v %v", got, err)
				}
			}
			if got, err := s.ReadVersion(ctx, "beads/absent", memory.Version); got != nil || !errors.Is(err, ErrNotFound) {
				t.Fatalf("absent subject: %+v %v", got, err)
			}
			canceled, cancel := context.WithCancel(ctx)
			cancel()
			if got, err := s.ReadVersion(canceled, "beads/plan", memory.Version); got != nil || !errors.Is(err, context.Canceled) {
				t.Fatalf("canceled read: %+v %v", got, err)
			}
			for name, statement := range map[string]string{
				"binding":          `UPDATE graph_preview_scope SET authority_id='00000000000000000000000000000000'`,
				"snapshot":         `UPDATE graph_preview_versions SET snapshot='{}'`,
				"positive-missing": `DELETE FROM graph_preview_versions`,
				"actor":            `UPDATE graph_preview_versions SET actor='different'`,
			} {
				t.Run(name, func(t *testing.T) {
					var saved []byte
					if err := s.db.QueryRowContext(ctx, `SELECT snapshot FROM graph_preview_versions WHERE path='beads/plan'`).Scan(&saved); err != nil {
						t.Fatal(err)
					}
					if _, err := s.db.ExecContext(ctx, statement); err != nil {
						t.Fatal(err)
					}
					got, readErr := s.ReadVersion(ctx, "beads/plan", memory.Version)
					if _, err := s.db.ExecContext(ctx, `UPDATE graph_preview_scope SET authority_id=?`, o.Binding.AuthorityID); err != nil {
						t.Fatal(err)
					}
					if _, err := s.db.ExecContext(ctx, `REPLACE INTO graph_preview_versions(path,version,snapshot,actor) VALUES(?,?,?,'')`, "beads/plan", memory.Version, saved); err != nil {
						t.Fatal(err)
					}
					if !errors.Is(readErr, ErrInvalidStore) || got != nil {
						t.Fatalf("corruption leaked state: %+v %v", got, readErr)
					}
				})
			}
			// SQL-side sizing must reject before acquiring or decoding this blob.
			if _, err := s.db.ExecContext(ctx, `UPDATE graph_preview_versions SET snapshot=REPEAT('x',?)`, PreviewCurrentReadByteLimit+1); err != nil {
				t.Fatal(err)
			}
			if got, err := s.ReadVersion(ctx, "beads/plan", memory.Version); got != nil || !errors.Is(err, ErrLimitExceeded) {
				t.Fatalf("unbounded retained acquisition: %+v %v", got, err)
			}
		})
	}
}

// The exact Issue mapping fault is tested without altering Jim's recorder.
func TestReadVersionMissingIssueBody(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			ctx, o := issueExperimentOptions(t, backend)
			s, err := OpenExisting(ctx, o)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := s.Close(); err != nil {
					t.Error(err)
				}
			}()
			issue, err := s.CreateIssue(ctx, "beads/work", plainIssue("Work"))
			if err != nil {
				t.Fatal(err)
			}
			err = s.withTx(ctx, false, func(tx *sql.Tx) error {
				if _, err := tx.ExecContext(ctx, `DELETE FROM issue_versions`); err != nil {
					return err
				}
				got, err := s.readIssueVersionInTx(ctx, tx, "beads/work", issue.Version, issue.Version, issue.Properties.ID)
				if !errors.Is(err, ErrInvalidStore) || !reflect.DeepEqual(got, IssueRecord{}) {
					return fmt.Errorf("missing mapped Issue body leaked state: %+v %v", got, err)
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
