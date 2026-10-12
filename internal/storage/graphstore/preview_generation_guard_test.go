//go:build cgo

package graphstore

import (
	"errors"
	"slices"
	"testing"

	"github.com/steveyegge/beads/internal/types"
)

// The workspace format generation is checked against the Types a fresh init
// installs, so the list it is checked against must be the installer's own list
// and not a second copy that can drift from it.
func TestFreshTypeNamesFollowPreviewTypeDefinitions(t *testing.T) {
	definitions := previewTypeDefinitions()
	want := make([]string, len(definitions))
	for i, definition := range definitions {
		want[i] = definition.name
	}
	got := FreshTypeNames()
	if !slices.Equal(got, want) {
		t.Fatalf("FreshTypeNames() = %v, want the installer's own order %v", got, want)
	}
	// The slice is the caller's: changing it must not reach the installer.
	got[0] = "changed"
	if again := FreshTypeNames(); !slices.Equal(again, want) {
		t.Fatalf("changing the returned slice changed a later call: %v", again)
	}
}

// Direct Store callers may omit TypeGeneration. That zero value can describe a
// fresh v7 installation or an older pinned one; only the installed, checked
// descriptor set decides whether relationship journeys are available.
func TestUnspecifiedStoreGenerationRequiresFullV7Relationships(t *testing.T) {
	ctx, options := issueExperimentOptions(t, "embedded")
	if options.TypeGeneration != 0 {
		t.Fatalf("direct Store fixture should omit TypeGeneration, got %d", options.TypeGeneration)
	}
	store, err := OpenExisting(ctx, options)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Error(err)
		}
	})
	epicRequest := plainIssue("Release epic")
	epicRequest.Issue.IssueType = types.TypeEpic
	epic, err := store.CreateIssue(ctx, "beads/epic", epicRequest)
	if err != nil {
		t.Fatal(err)
	}
	request := EpicCloseRequest{Path: "beads/epic", ExpectedRevision: epic.Revision, Reason: "Done", Actor: "author"}
	if _, err := store.ListChildren(ctx, "beads/epic"); err != nil {
		t.Fatalf("fresh v7 children rejected with unspecified generation: %v", err)
	}
	if _, err := store.ListChildrenLegacy(ctx, "beads/epic"); err != nil {
		t.Fatalf("fresh v7 counted children rejected with unspecified generation: %v", err)
	}
	if _, err := store.EpicStatuses(ctx); err != nil {
		t.Fatalf("fresh v7 epic progress rejected with unspecified generation: %v", err)
	}
	if _, err := store.CloseEligibleEpic(ctx, request); !errors.Is(err, ErrConflict) {
		t.Fatalf("empty epic should reach eligibility check, got %v", err)
	}

	var fingerprint string
	if err := store.db.QueryRowContext(ctx, `SELECT fingerprint FROM graph_preview_types WHERE name='parent-child'`).Scan(&fingerprint); err != nil {
		t.Fatal(err)
	}
	if _, err := store.db.ExecContext(ctx, `UPDATE graph_preview_types SET fingerprint='wrong' WHERE name='parent-child'`); err != nil {
		t.Fatal(err)
	}
	if _, err := store.EpicStatuses(ctx); !errors.Is(err, ErrInvalidStore) {
		t.Fatalf("altered v7 fingerprint admitted: %v", err)
	}
	if _, err := store.db.ExecContext(ctx, `UPDATE graph_preview_types SET fingerprint=? WHERE name='parent-child'`, fingerprint); err != nil {
		t.Fatal(err)
	}

	checkRefused := func(generation string) {
		t.Helper()
		if _, err := store.ListChildren(ctx, "beads/epic"); !errors.Is(err, ErrCapabilityUnavailable) {
			t.Fatalf("%s children admitted: %v", generation, err)
		}
		if _, err := store.ListChildrenLegacy(ctx, "beads/epic"); !errors.Is(err, ErrCapabilityUnavailable) {
			t.Fatalf("%s counted children admitted: %v", generation, err)
		}
		if _, err := store.EpicStatuses(ctx); !errors.Is(err, ErrCapabilityUnavailable) {
			t.Fatalf("%s epic progress admitted: %v", generation, err)
		}
		if _, err := store.CloseEligibleEpic(ctx, request); !errors.Is(err, ErrCapabilityUnavailable) {
			t.Fatalf("%s epic close admitted: %v", generation, err)
		}
	}
	if _, err := store.db.ExecContext(ctx, `DELETE FROM graph_preview_types WHERE name IN ('issue-v3','blocks-v2','parent-child','duplicates','supersedes')`); err != nil {
		t.Fatal(err)
	}
	checkRefused("six-Type")
	if _, err := store.db.ExecContext(ctx, `DELETE FROM graph_preview_types WHERE name IN ('example-follows','example-cites')`); err != nil {
		t.Fatal(err)
	}
	checkRefused("four-Type")
}

// The common-metadata columns change the preview storage layout and require
// SchemaVersion 7. Older workspaces remain pinned to their original binary;
// this binary refuses them rather than silently upgrading their data. The
// workspace marker generation still describes the installed Type set.
//
// Counting Types is plain SQL over graph_preview_types and identical on both
// engines, so only the embedded engine is exercised here.
func TestPreviewSchemaVersionAndInstalledTypeCounts(t *testing.T) {
	if SchemaVersion != 7 {
		t.Fatalf("SchemaVersion is %d, want 7 for the common-metadata storage layout", SchemaVersion)
	}
	ctx, o := issueExperimentOptions(t, "embedded")
	s, err := OpenExisting(ctx, o)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := s.Close(); err != nil {
			t.Error(err)
		}
	})
	installed := func() (int, error) {
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := tx.Rollback(); err != nil {
				t.Error(err)
			}
		}()
		descriptors, err := installedPreviewTypes(ctx, tx, s.ScopeURL())
		return len(descriptors), err
	}
	if count, err := installed(); err != nil || count != 11 {
		t.Fatalf("fresh installation admitted %d Types with %v, want all 11", count, err)
	}
	// Ten Types is a partial relationship generation and must be refused.
	if _, err := s.db.ExecContext(ctx, `DELETE FROM graph_preview_types WHERE name='supersedes'`); err != nil {
		t.Fatal(err)
	}
	if count, err := installed(); !errors.Is(err, ErrInvalidStore) {
		t.Fatalf("ten Types admitted: %d Types with %v, want ErrInvalidStore", count, err)
	}
	// Remove the rest of the v7 suffix to recover the exact six-Type set.
	if _, err := s.db.ExecContext(ctx, `DELETE FROM graph_preview_types WHERE name IN ('issue-v3','blocks-v2','parent-child','duplicates')`); err != nil {
		t.Fatal(err)
	}
	if count, err := installed(); err != nil || count != 6 {
		t.Fatalf("prior installation admitted %d Types with %v, want 6", count, err)
	}
	// Removing both example rows leaves precisely the original four.
	if _, err := s.db.ExecContext(ctx, `DELETE FROM graph_preview_types WHERE name IN ('example-follows','example-cites')`); err != nil {
		t.Fatal(err)
	}
	if count, err := installed(); err != nil || count != 4 {
		t.Fatalf("legacy installation admitted %d Types with %v, want the original 4", count, err)
	}
}
