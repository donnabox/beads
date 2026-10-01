//go:build cgo

package graphstore

import (
	"context"
	"errors"
	"testing"
	"time"
)

func openVersionsStore(t *testing.T, ctx context.Context, o Options) *Store {
	t.Helper()
	s, err := OpenExisting(ctx, o)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := s.Close(); err != nil {
			t.Error(err)
		}
	})
	return s
}

// checkVersionList asserts the invariants every returned list must satisfy,
// independent of plane: newest first by ordinal, strictly descending, positive,
// a resolvable token on every row, and a populated change instant. marker names
// the one token allowed to refuse with ErrGone instead of resolving, which is a
// deleted Link's deletion marker; "" means every row must resolve.
func checkVersionList(t *testing.T, ctx context.Context, s *Store, path string, rows []VersionRow, marker string) {
	t.Helper()
	if len(rows) == 0 {
		t.Fatalf("%s: empty version list", path)
	}
	for i, row := range rows {
		if row.Ordinal < 1 {
			t.Errorf("%s row %d: non-positive ordinal %d", path, i, row.Ordinal)
		}
		if i > 0 && row.Ordinal >= rows[i-1].Ordinal {
			t.Errorf("%s row %d: ordinal %d does not descend below %d", path, i, row.Ordinal, rows[i-1].Ordinal)
		}
		if row.ChangeAt.IsZero() || row.ChangeAt.Location() != time.UTC {
			t.Errorf("%s row %d: change instant %v is not a populated UTC time", path, i, row.ChangeAt)
		}
		// Every token this reader hands out must address retained state; that is
		// the whole point of returning Version alongside the local ordinal.
		_, err := s.ReadVersion(ctx, path, row.Version)
		if row.Version == marker {
			if !errors.Is(err, ErrGone) {
				t.Errorf("%s row %d: deletion marker %q resolved as a Resource: %v", path, i, row.Version, err)
			}
			continue
		}
		if err != nil {
			t.Errorf("%s row %d: listed token %q does not resolve: %v", path, i, row.Version, err)
		}
	}
}

// TestVersionsIssuePlaneRecordsWithoutTheConfigFlag verifies the premise this
// reader rests on rather than assuming it: a graph workspace holds native
// issue_versions rows for every accepted Issue mutation even though no
// versioned-history setting is enabled anywhere in it. Each graph Issue write
// path scopes recording on for its own transaction
// (issueops.ScopeVersionedHistoryTransaction(tx, true)), so the
// config default of false never reaches those writes. If this premise ever
// breaks, the Issue plane silently loses its ordering authority, so it is
// asserted directly against the tables, not inferred from the reader's output.
func TestVersionsIssuePlaneRecordsWithoutTheConfigFlag(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			ctx, o := issueExperimentOptions(t, backend)
			s := openVersionsStore(t, ctx, o)
			work, err := s.CreateIssue(ctx, "beads/work", plainIssue("Work"))
			if err != nil {
				t.Fatal(err)
			}
			prereq, err := s.CreateIssue(ctx, "beads/prereq", plainIssue("Prerequisite"))
			if err != nil {
				t.Fatal(err)
			}
			title := "Work, retitled — 記憶"
			updated, err := s.UpdateIssue(ctx, UpdateIssueRequest{Path: "beads/work", Actor: "editor",
				ExpectedRevision: work.Revision, Title: &title})
			if err != nil {
				t.Fatal(err)
			}
			dep, err := s.AddDependency(ctx, DependencyRequest{Path: "links/block", SourcePath: "beads/work",
				TargetPath: "beads/prereq", ExpectedSourceRevision: updated.Issue.Revision, Actor: "author"})
			if err != nil {
				t.Fatal(err)
			}
			claimed, err := s.ClaimIssue(ctx, "beads/work", "claimant")
			if err != nil {
				t.Fatal(err)
			}
			if _, err := s.CloseIssue(ctx, "beads/prereq", "Done", "closer"); err != nil {
				t.Fatal(err)
			}
			reopened, err := s.ReopenIssue(ctx, "beads/prereq", "Not done after all", "reopener")
			if err != nil {
				t.Fatal(err)
			}
			// No versioned-history setting exists in this workspace at all: the
			// preview installs exactly one config row, issue_prefix.
			var settings int
			if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM config WHERE `key` LIKE 'versioned-history%'").Scan(&settings); err != nil {
				t.Fatal(err)
			}
			if settings != 0 {
				t.Fatalf("workspace enables versioned history through config in %d rows; the premise under test is that it does not", settings)
			}
			for _, subject := range []struct {
				path, issueID, head string
				mutations           int
			}{
				{"beads/work", work.Properties.ID, claimed.Issue.Revision, 4},
				{"beads/prereq", prereq.Properties.ID, reopened.Issue.Revision, 3},
			} {
				var native int
				if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM issue_versions WHERE issue_id=?`, subject.issueID).Scan(&native); err != nil {
					t.Fatal(err)
				}
				if native < subject.mutations {
					t.Fatalf("%s: native plane recorded %d versions, want at least %d; recording is not being forced on", subject.path, native, subject.mutations)
				}
				kind, rows, err := s.Versions(ctx, subject.path)
				if err != nil {
					t.Fatal(err)
				}
				if kind != KindIssue {
					t.Fatalf("%s: kind=%q, want %q", subject.path, kind, KindIssue)
				}
				if len(rows) != subject.mutations {
					t.Fatalf("%s: %d versions, want %d: %+v", subject.path, len(rows), subject.mutations, rows)
				}
				checkVersionList(t, ctx, s, subject.path, rows, "")
				// The newest listed version is the current head, and the oldest
				// is the creating mutation's native ordinal 1.
				if rows[0].Version != subject.head {
					t.Errorf("%s: newest version %q, want current head %q", subject.path, rows[0].Version, subject.head)
				}
				if rows[len(rows)-1].Ordinal != 1 {
					t.Errorf("%s: oldest ordinal %d, want 1", subject.path, rows[len(rows)-1].Ordinal)
				}
				// Issues carry the native attribution_status; Memory and Link do not.
				for i, row := range rows {
					if row.Attribution == "" {
						t.Errorf("%s row %d: empty attribution status on the Issue plane", subject.path, i)
					}
				}
			}
			// Actor comes from the mutation that minted the version, not from
			// the subject's creator: the claim is the newest version of work.
			kind, rows, err := s.Versions(ctx, "beads/work")
			if err != nil || kind != KindIssue {
				t.Fatalf("work versions: %v %q", err, kind)
			}
			if rows[0].Actor != "claimant" {
				t.Errorf("newest work version actor=%q, want the claiming actor", rows[0].Actor)
			}
			// The owned Dependency is a Link with its own independent history.
			kind, links, err := s.Versions(ctx, "links/block")
			if err != nil {
				t.Fatal(err)
			}
			if kind != KindLink || len(links) != 1 || links[0].Ordinal != 1 || links[0].Version != dep.Link.Revision {
				t.Fatalf("Dependency versions: kind=%q rows=%+v", kind, links)
			}
			if links[0].Attribution != "" {
				t.Errorf("Link version claims attribution status %q; that column exists only on the native plane", links[0].Attribution)
			}
			checkVersionList(t, ctx, s, "links/block", links, "")
		})
	}
}

// TestVersionsMemoryAndLinkPlanes covers the graph_preview_versions plane, whose
// ordinal is allocated by the preview writer rather than by the Issue domain.
func TestVersionsMemoryAndLinkPlanes(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			ctx, o := issueExperimentOptions(t, backend)
			s := openVersionsStore(t, ctx, o)
			plan, err := s.Create(ctx, CreateRequest{Path: "beads/plan", Title: "Plan", Body: "Original — 雪", Actor: "author"})
			if err != nil {
				t.Fatal(err)
			}
			notes, err := s.Create(ctx, CreateRequest{Path: "beads/notes", Title: "Notes", Actor: "author"})
			if err != nil {
				t.Fatal(err)
			}
			revised, err := s.UpdateMemory(ctx, MemoryUpdateRequest{Path: "beads/plan",
				Properties: Properties{Title: "Plan", Body: "Revised"}, Actor: "editor", ExpectedRevision: plan.Revision})
			if err != nil {
				t.Fatal(err)
			}
			// A Memory owns its outgoing Links, so linking mints a Memory version
			// as well as the Link's own first version.
			added, err := s.AddInformationalLink(ctx, LinkCreateRequest{Path: "links/context",
				SourcePath: "beads/plan", TargetPath: "beads/notes", ExpectedSourceRevision: revised.Memory.Revision,
				Properties: map[string]any{"note": "before"}, Actor: "linker"})
			if err != nil {
				t.Fatal(err)
			}
			relinked, err := s.UpdateLink(ctx, LinkUpdateRequest{Path: "links/context",
				ExpectedRevision: added.Link.Revision, ExpectedSourceRevision: added.Source.(Record).Revision,
				Properties: map[string]any{"note": "after"}, Actor: "relinker"})
			if err != nil {
				t.Fatal(err)
			}
			kind, rows, err := s.Versions(ctx, "beads/plan")
			if err != nil {
				t.Fatal(err)
			}
			if kind != KindMemory {
				t.Fatalf("Memory kind=%q", kind)
			}
			// create, update, link (owned set changed), relink (owned set changed).
			if len(rows) != 4 {
				t.Fatalf("Memory versions: %d rows %+v", len(rows), rows)
			}
			checkVersionList(t, ctx, s, "beads/plan", rows, "")
			if rows[0].Version != relinked.Source.(Record).Revision || rows[len(rows)-1].Version != plan.Revision {
				t.Errorf("Memory list is not newest-first: %+v", rows)
			}
			if rows[0].Actor != "relinker" || rows[len(rows)-1].Actor != "author" {
				t.Errorf("Memory actors: newest=%q oldest=%q", rows[0].Actor, rows[len(rows)-1].Actor)
			}
			for i, row := range rows {
				if row.Attribution != "" {
					t.Errorf("Memory row %d claims attribution status %q", i, row.Attribution)
				}
			}
			kind, links, err := s.Versions(ctx, "links/context")
			if err != nil {
				t.Fatal(err)
			}
			if kind != KindLink || len(links) != 2 {
				t.Fatalf("Link versions: kind=%q rows=%+v", kind, links)
			}
			checkVersionList(t, ctx, s, "links/context", links, "")
			if links[0].Version != relinked.Link.Revision || links[1].Version != added.Link.Revision {
				t.Errorf("Link list is not newest-first: %+v", links)
			}
			if links[0].Ordinal != 2 || links[1].Ordinal != 1 {
				t.Errorf("Link ordinals %d,%d, want 2,1", links[0].Ordinal, links[1].Ordinal)
			}
			// A Memory that was never linked still has its own single version.
			kind, untouched, err := s.Versions(ctx, "beads/notes")
			if err != nil || kind != KindMemory || len(untouched) != 1 || untouched[0].Version != notes.Revision {
				t.Fatalf("unlinked Memory: kind=%q rows=%+v err=%v", kind, untouched, err)
			}
		})
	}
}

// TestVersionsOrderByOrdinalNotChangeAt pins the ordering authority on both
// planes. change_at is an observed wall clock: a clock step, a replayed write,
// or two mutations inside the same microsecond must not reorder a history, so
// the stored instants are deliberately set to disagree with the ordinals here.
func TestVersionsOrderByOrdinalNotChangeAt(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			ctx, o := issueExperimentOptions(t, backend)
			s := openVersionsStore(t, ctx, o)
			plan, err := s.Create(ctx, CreateRequest{Path: "beads/plan", Title: "Plan", Actor: "author"})
			if err != nil {
				t.Fatal(err)
			}
			revised, err := s.UpdateMemory(ctx, MemoryUpdateRequest{Path: "beads/plan",
				Properties: Properties{Title: "Plan", Body: "Revised"}, Actor: "editor", ExpectedRevision: plan.Revision})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := s.UpdateMemory(ctx, MemoryUpdateRequest{Path: "beads/plan",
				Properties: Properties{Title: "Plan", Body: "Revised again"}, Actor: "editor",
				ExpectedRevision: revised.Memory.Revision}); err != nil {
				t.Fatal(err)
			}
			issue, err := s.CreateIssue(ctx, "beads/work", plainIssue("Work"))
			if err != nil {
				t.Fatal(err)
			}
			title := "Work, retitled"
			if _, err := s.UpdateIssue(ctx, UpdateIssueRequest{Path: "beads/work", Actor: "editor",
				ExpectedRevision: issue.Revision, Title: &title}); err != nil {
				t.Fatal(err)
			}
			// Invert both planes' instants against their ordinals: the oldest
			// version now carries the newest timestamp.
			base := time.Date(2031, 3, 1, 12, 0, 0, 0, time.UTC)
			for ordinal := 1; ordinal <= 3; ordinal++ {
				at := base.Add(time.Duration(-ordinal) * time.Hour)
				if _, err := s.db.ExecContext(ctx, `UPDATE graph_preview_versions SET change_at=? WHERE path='beads/plan' AND ordinal=?`, at, ordinal); err != nil {
					t.Fatal(err)
				}
				if _, err := s.db.ExecContext(ctx, `UPDATE issue_versions SET change_at=? WHERE issue_id=? AND revision=?`, at, issue.Properties.ID, ordinal); err != nil {
					t.Fatal(err)
				}
			}
			for _, path := range []string{"beads/plan", "beads/work"} {
				_, rows, err := s.Versions(ctx, path)
				if err != nil {
					t.Fatal(err)
				}
				if len(rows) < 2 {
					t.Fatalf("%s: %d versions", path, len(rows))
				}
				for i, row := range rows {
					if i > 0 && row.Ordinal >= rows[i-1].Ordinal {
						t.Fatalf("%s: ordinals stopped descending once change_at disagreed: %+v", path, rows)
					}
					// The inverted instants must come back ascending, which is
					// only possible if nothing sorted by them.
					if i > 0 && !row.ChangeAt.After(rows[i-1].ChangeAt) {
						t.Fatalf("%s: list follows change_at rather than ordinal: %+v", path, rows)
					}
				}
			}
		})
	}
}

// TestVersionsMissingSubjectIsNotAnEmptyList keeps the outcomes distinct: a path
// that was never allocated refuses with the package's existing ErrNotFound, and
// a real subject with nothing retained returns its kind with an empty list and
// no error. The Issue plane's "nothing retained" case is its mappings being
// gone; a mapping whose ordering authority is missing is a refusal instead, and
// TestVersionsRefusesUnorderableAndCorruptState covers it.
func TestVersionsMissingSubjectIsNotAnEmptyList(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			ctx, o := issueExperimentOptions(t, backend)
			s := openVersionsStore(t, ctx, o)
			for _, path := range []string{"beads/absent", "links/absent"} {
				kind, rows, err := s.Versions(ctx, path)
				if !errors.Is(err, ErrNotFound) || kind != "" || rows != nil {
					t.Fatalf("%s: kind=%q rows=%+v err=%v; want ErrNotFound", path, kind, rows, err)
				}
			}
			// A malformed path is a validation refusal, not an absence: the
			// caller asked an unanswerable question rather than a false one.
			if _, _, err := s.Versions(ctx, "not-a-resource"); err == nil || errors.Is(err, ErrNotFound) {
				t.Fatalf("malformed path: %v", err)
			}
			if _, err := s.Create(ctx, CreateRequest{Path: "beads/plan", Title: "Plan", Actor: "author"}); err != nil {
				t.Fatal(err)
			}
			if _, err := s.CreateIssue(ctx, "beads/work", plainIssue("Work")); err != nil {
				t.Fatal(err)
			}
			// Drop the retained rows out of band. Neither plane can reach this
			// state through the writer, which always retains its head, but the
			// contract distinguishes "allocated, nothing to show" from both a
			// refusal and corruption, so the distinction has to be exercised.
			if _, err := s.db.ExecContext(ctx, `DELETE FROM graph_preview_versions WHERE path='beads/plan'`); err != nil {
				t.Fatal(err)
			}
			if _, err := s.db.ExecContext(ctx, `DELETE FROM graph_preview_issue_versions WHERE path='beads/work'`); err != nil {
				t.Fatal(err)
			}
			for _, subject := range []struct {
				path string
				kind ResourceKind
			}{{"beads/plan", KindMemory}, {"beads/work", KindIssue}} {
				kind, rows, err := s.Versions(ctx, subject.path)
				if err != nil || kind != subject.kind || rows != nil {
					t.Fatalf("%s with nothing retained: kind=%q rows=%+v err=%v", subject.path, kind, rows, err)
				}
			}
		})
	}
}

// TestVersionsRefusesUnorderableAndCorruptState covers the two ways this reader
// could otherwise shorten a history without saying so. The two refusals are
// deliberately different: a plane with no ordering authority left is a missing
// capability, while retained rows with no allocation are corruption.
func TestVersionsRefusesUnorderableAndCorruptState(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			ctx, o := issueExperimentOptions(t, backend)
			s := openVersionsStore(t, ctx, o)
			issue, err := s.CreateIssue(ctx, "beads/work", plainIssue("Work"))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := s.Create(ctx, CreateRequest{Path: "beads/plan", Title: "Plan", Actor: "author"}); err != nil {
				t.Fatal(err)
			}
			// Mappings whose native rows are all gone leave the Issue plane with
			// nothing to order by. Reporting an empty list would let that pass
			// as "no history yet", which is a different and answerable fact.
			if _, err := s.db.ExecContext(ctx, `DELETE FROM issue_versions WHERE issue_id=?`, issue.Properties.ID); err != nil {
				t.Fatal(err)
			}
			_, rows, err := s.Versions(ctx, "beads/work")
			if !errors.Is(err, ErrCapabilityUnavailable) || rows != nil {
				t.Fatalf("mappings without native versions: rows=%+v err=%v", rows, err)
			}
			// Retained rows with no allocation are corruption, never an absence.
			if _, err := s.db.ExecContext(ctx, `DELETE FROM graph_preview_catalog WHERE path='beads/plan'`); err != nil {
				t.Fatal(err)
			}
			if _, _, err := s.Versions(ctx, "beads/plan"); !errors.Is(err, ErrInvalidStore) {
				t.Fatalf("orphan retained rows: %v", err)
			}
		})
	}
}

// TestVersionsOfDeletedSubjects pins what survives deletion: a deleted subject
// keeps its history rather than becoming an absence.
func TestVersionsOfDeletedSubjects(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			ctx, o := issueExperimentOptions(t, backend)
			s := openVersionsStore(t, ctx, o)
			plan, err := s.Create(ctx, CreateRequest{Path: "beads/plan", Title: "Plan", Actor: "author"})
			if err != nil {
				t.Fatal(err)
			}
			notes, err := s.Create(ctx, CreateRequest{Path: "beads/notes", Title: "Notes", Actor: "author"})
			if err != nil {
				t.Fatal(err)
			}
			added, err := s.AddInformationalLink(ctx, LinkCreateRequest{Path: "links/context",
				SourcePath: "beads/plan", TargetPath: "beads/notes", ExpectedSourceRevision: plan.Revision,
				Properties: map[string]any{"note": "before"}, Actor: "linker"})
			if err != nil {
				t.Fatal(err)
			}
			removed, err := s.Unlink(ctx, LinkDeleteRequest{Path: "links/context",
				ExpectedRevision: added.Link.Revision, ExpectedSourceRevision: added.Source.(Record).Revision,
				Actor: "remover"})
			if err != nil {
				t.Fatal(err)
			}
			kind, links, err := s.Versions(ctx, "links/context")
			if err != nil {
				t.Fatal(err)
			}
			if kind != KindLink {
				t.Fatalf("deleted Link kind=%q", kind)
			}
			// The reader reports what the table holds, so the deletion marker is
			// listed. It is the one listed token ReadVersion refuses, with
			// ErrGone: whether to withhold it is a presentation decision for the
			// caller, and this test pins today's behaviour so that decision is
			// made against a measured list rather than a guessed one.
			checkVersionList(t, ctx, s, "links/context", links, removed.Link.Revision)
			if len(links) != 2 || links[0].Version != removed.Link.Revision || links[1].Version != added.Link.Revision {
				t.Fatalf("deleted Link history: %+v", links)
			}
			// A deleted Memory mints no successor version, so its final live head
			// stays its newest retained version and stays readable.
			deleted, err := s.DeleteMemory(ctx, MemoryDeleteRequest{Path: "beads/notes",
				ExpectedRevision: notes.Revision, Actor: "remover"})
			if err != nil || !deleted.Deleted {
				t.Fatalf("delete Memory: %+v %v", deleted, err)
			}
			kind, rows, err := s.Versions(ctx, "beads/notes")
			if err != nil {
				t.Fatal(err)
			}
			if kind != KindMemory || len(rows) != 1 || rows[0].Version != notes.Revision {
				t.Fatalf("deleted Memory history: kind=%q rows=%+v", kind, rows)
			}
			checkVersionList(t, ctx, s, "beads/notes", rows, "")
		})
	}
}

// TestVersionsIsAReadOnlyOperation pins that listing history changes no writer
// coordination state, the same guarantee the exact retained reader gives.
func TestVersionsIsAReadOnlyOperation(t *testing.T) {
	ctx, o := issueExperimentOptions(t, "embedded")
	s := openVersionsStore(t, ctx, o)
	if _, err := s.Create(ctx, CreateRequest{Path: "beads/plan", Title: "Plan", Actor: "author"}); err != nil {
		t.Fatal(err)
	}
	var before, after string
	if err := s.db.QueryRowContext(ctx, `SELECT writer_token FROM graph_preview_scope WHERE singleton=1`).Scan(&before); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.Versions(ctx, "beads/plan"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.Versions(ctx, "beads/absent"); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
	if err := s.db.QueryRowContext(ctx, `SELECT writer_token FROM graph_preview_scope WHERE singleton=1`).Scan(&after); err != nil {
		t.Fatal(err)
	}
	if before != after {
		t.Fatal("listing versions changed writer coordination state")
	}
}
