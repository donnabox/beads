package issueops

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"math"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/types"
	publicops "github.com/steveyegge/beads/issueops"
)

func TestCanonicalFinalContentValidation(t *testing.T) {
	negative, oversized := -1, int(math.MaxInt32)
	oversized++ // remains a refusal on 32-bit test hosts (negative after overflow)
	longRef, longSpec := strings.Repeat("r", 256), strings.Repeat("s", 1025)
	fractional := time.Date(2026, 10, 29, 12, 0, 0, 1, time.UTC)
	for _, tc := range []struct {
		name    string
		content CanonicalIssueContent
	}{
		{"blank title", CanonicalIssueContent{Title: publicops.Field[string]{Set: true}}},
		{"invalid UTF-8", CanonicalIssueContent{Description: publicops.Field[string]{Set: true, Value: string([]byte{255})}}},
		{"invalid priority", CanonicalIssueContent{Priority: publicops.Field[int]{Set: true, Value: 5}}},
		{"negative estimate", CanonicalIssueContent{EstimatedMinutes: publicops.Field[*int]{Set: true, Value: &negative}}},
		{"storage estimate", CanonicalIssueContent{EstimatedMinutes: publicops.Field[*int]{Set: true, Value: &oversized}}},
		{"external ref", CanonicalIssueContent{ExternalRef: publicops.Field[*string]{Set: true, Value: &longRef}}},
		{"spec id", CanonicalIssueContent{SpecID: publicops.Field[string]{Set: true, Value: longSpec}}},
		{"fractional date", CanonicalIssueContent{DueAt: publicops.Field[*time.Time]{Set: true, Value: &fractional}}},
		{"null metadata", CanonicalIssueContent{Metadata: publicops.Field[json.RawMessage]{Set: true, Value: json.RawMessage(`null`)}}},
		{"array metadata", CanonicalIssueContent{Metadata: publicops.Field[json.RawMessage]{Set: true, Value: json.RawMessage(`[]`)}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := ValidateCanonicalIssueContent(tc.content); !errors.Is(err, storage.ErrValidation) {
				t.Fatalf("expected validation refusal, got %v", err)
			}
		})
	}
	zero := 0
	valid := CanonicalIssueContent{
		Title:            publicops.Field[string]{Set: true, Value: "A valid native task"},
		Priority:         publicops.Field[int]{Set: true, Value: 0},
		EstimatedMinutes: publicops.Field[*int]{Set: true, Value: &zero},
		ExternalRef:      publicops.Field[*string]{Set: true}, DueAt: publicops.Field[*time.Time]{Set: true},
		Metadata: publicops.Field[json.RawMessage]{Set: true, Value: json.RawMessage(`{"nested":{"revision":"literal"}}`)},
	}
	if err := ValidateCanonicalIssueContent(valid); err != nil {
		t.Fatal(err)
	}
}

func TestCanonicalFinalDependencyClassification(t *testing.T) {
	for _, kind := range types.WellKnownDependencyTypes() {
		request := CanonicalDependencyMetadata{IssueID: "a", DependsOnID: "b", Type: kind, Metadata: json.RawMessage(`{"note":"final"}`)}
		err := ValidateCanonicalDependencyMetadata(request, "actor")
		if kind.AffectsReadyWork() {
			if !errors.Is(err, storage.ErrValidation) {
				t.Fatalf("readiness relationship %s accepted: %v", kind, err)
			}
		} else if err != nil {
			t.Fatalf("non-readiness relationship %s refused: %v", kind, err)
		}
	}
	for _, kind := range []types.DependencyType{"custom-relation", "", types.DepWaitsFor} {
		if err := ValidateCanonicalDependencyMetadata(CanonicalDependencyMetadata{IssueID: "a", DependsOnID: "b", Type: kind, Metadata: json.RawMessage(`{}`)}, "actor"); !errors.Is(err, storage.ErrValidation) {
			t.Fatalf("unadmitted %q: %v", kind, err)
		}
	}
}

func TestCanonicalFinalStoredIntent(t *testing.T) {
	minutes := 10
	after := &types.Issue{Title: "final", EstimatedMinutes: &minutes, Metadata: json.RawMessage(`{"b":2,"a":1}`)}
	if err := checkCanonicalStoredContent(after, map[string]interface{}{"title": "final", "estimated_minutes": &minutes, "metadata": json.RawMessage(`{"a":1,"b":2}`)}); err != nil {
		t.Fatal(err)
	}
	if err := checkCanonicalStoredContent(after, map[string]interface{}{"title": "silently truncated"}); !errors.Is(err, storage.ErrValidation) {
		t.Fatalf("lost text accepted: %v", err)
	}
	after.Metadata = nil // the native hydrator omits a stored empty object
	if err := checkCanonicalStoredContent(after, map[string]interface{}{"metadata": json.RawMessage(`{}`)}); err != nil {
		t.Fatal(err)
	}
	if err := checkCanonicalStoredContent(after, map[string]interface{}{"metadata": json.RawMessage(`{"lost":true}`)}); !errors.Is(err, storage.ErrValidation) {
		t.Fatalf("lost metadata accepted: %v", err)
	}
}

func canonicalFinalHydration(mock sqlmock.Sqlmock, id, title string) {
	mock.ExpectQuery(`SELECT .* FROM issues .* WHERE id = \?`).WithArgs(id).
		WillReturnRows(issueRows().AddRow(issueRowValues(id, title)...))
	mock.ExpectQuery(`SELECT label FROM labels WHERE issue_id = \?`).WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{"label"}))
}

func canonicalFinalNotWisp(mock sqlmock.Sqlmock, id string) {
	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM wisps WHERE id = ? LIMIT 1")).WithArgs(id).
		WillReturnError(sql.ErrNoRows)
}

// SQL mocks prove the composite does not accidentally invoke an intermediate
// audit, journal or native version writer. Real engine qualification belongs
// to the enclosing atomic publisher's final snapshot/rollback fixtures.
func TestCanonicalFinalIssueDefersPublication(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	cleanupJournal := ScopeEventsJournalTransaction(tx, true)
	defer cleanupJournal()
	cleanupVersions := ScopeVersionedHistoryTransaction(tx, true)
	defer cleanupVersions()
	canonicalFinalHydration(mock, "a", "before")
	canonicalFinalNotWisp(mock, "a")
	canonicalFinalNotWisp(mock, "a")
	canonicalFinalHydration(mock, "a", "before")
	mock.ExpectExec("UPDATE issues SET .*WHERE id = ").WillReturnResult(sqlmock.NewResult(0, 1))
	canonicalFinalHydration(mock, "a", "after")
	facts, err := ApplyCanonicalIssueContentInTx(context.Background(), tx, "a", CanonicalIssueContent{Title: publicops.Field[string]{Set: true, Value: "after"}}, "actor")
	if err != nil {
		t.Fatal(err)
	}
	if !facts.Changed() || facts.IssueID() != "a" {
		t.Fatalf("wrong final facts: %+v", facts)
	}
	if !journalEnabled(context.Background(), tx) || !VersionedHistoryEnabledFor(tx) {
		t.Fatal("composite changed transaction activation")
	}
	if facts.oldValue == "" || facts.newValue != `{"title":"after"}` {
		t.Fatalf("lost deferred audit: %+v", facts)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
	// Final facts use the original transaction activation and settled row.
	// The graph publisher will place its final version/snapshot writes before
	// this call; this test intentionally has no engine or fake version mint.
	mock.ExpectQuery("SELECT id FROM events").WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectExec("INSERT INTO events").WillReturnResult(sqlmock.NewResult(0, 1))
	canonicalFinalJournal(mock, "after", "update", nil)
	if err := EmitCanonicalFinalFactsInTx(context.Background(), tx, facts); err != nil {
		t.Fatal(err)
	}
	mock.ExpectRollback()
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func canonicalFinalJournal(mock sqlmock.Sqlmock, title, op string, dep any) {
	canonicalFinalHydration(mock, "a", title)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT is_blocked FROM issues WHERE id = ?")).WithArgs("a").
		WillReturnRows(sqlmock.NewRows([]string{"is_blocked"}).AddRow(0))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE bd_events_seq SET next_seq = next_seq + 1 WHERE id = 0")).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT next_seq FROM bd_events_seq WHERE id = 0")).WillReturnRows(sqlmock.NewRows([]string{"next_seq"}).AddRow(1))
	mock.ExpectExec("INSERT INTO bd_events_journal").WithArgs(1, sqlmock.AnyArg(), op, "a", "actor", sqlmock.AnyArg(), dep, nil).
		WillReturnResult(sqlmock.NewResult(0, 1))
}

func TestCanonicalFinalDependencyDefersPublication(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	cleanup := ScopeEventsJournalTransaction(tx, true)
	defer cleanup()
	cleanupVersions := ScopeVersionedHistoryTransaction(tx, true)
	defer cleanupVersions()
	for _, id := range []string{"a", "b"} {
		canonicalFinalHydration(mock, id, "existing")
		canonicalFinalNotWisp(mock, id)
	}
	mock.ExpectQuery(`SELECT type, metadata FROM dependencies WHERE issue_id=\? AND depends_on_issue_id=\? AND depends_on_wisp_id IS NULL AND depends_on_external IS NULL`).WithArgs("a", "b").
		WillReturnRows(sqlmock.NewRows([]string{"type", "metadata"}).AddRow("related", `{"note":"old"}`))
	for _, id := range []string{"a", "b"} {
		mock.ExpectQuery(`SELECT issue_type FROM issues WHERE id = \?`).WithArgs(id).
			WillReturnRows(sqlmock.NewRows([]string{"issue_type"}).AddRow("task"))
	}
	mock.ExpectQuery(`SELECT type, metadata, thread_id FROM dependencies WHERE issue_id = \? AND `).WithArgs("a", "b").
		WillReturnRows(sqlmock.NewRows([]string{"type", "metadata", "thread_id"}).AddRow("related", `{"note":"old"}`, "original-thread"))
	mock.ExpectExec(`UPDATE dependencies SET metadata = \?, thread_id = \? WHERE issue_id = \? AND `).
		WithArgs(`{"note":"final"}`, "original-thread", "a", "b").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT metadata FROM dependencies WHERE issue_id=\? AND depends_on_issue_id=\?`).WithArgs("a", "b").
		WillReturnRows(sqlmock.NewRows([]string{"metadata"}).AddRow(`{"note":"final"}`))
	facts, err := ApplyCanonicalDependencyMetadataInTx(context.Background(), tx, CanonicalDependencyMetadata{IssueID: "a", DependsOnID: "b", Type: types.DepRelated, Metadata: json.RawMessage(`{"note":"final"}`)}, "actor")
	if err != nil {
		t.Fatal(err)
	}
	if !facts.Changed() || facts.IssueID() != "a" {
		t.Fatalf("wrong facts: %+v", facts)
	}
	if !journalEnabled(context.Background(), tx) {
		t.Fatal("journal was globally disabled")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
	canonicalFinalJournal(mock, "existing", "dep_add", `{"kind":"related","target":"b","metadata":"{\"note\":\"final\"}"}`)
	if err := EmitCanonicalFinalFactsInTx(context.Background(), tx, facts); err != nil {
		t.Fatal(err)
	}
	mock.ExpectRollback()
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCanonicalFinalMissingDependencyDoesNotCreate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"a", "b"} {
		canonicalFinalHydration(mock, id, "existing")
		canonicalFinalNotWisp(mock, id)
	}
	mock.ExpectQuery(`SELECT type, metadata FROM dependencies WHERE issue_id=\? AND depends_on_issue_id=\? AND depends_on_wisp_id IS NULL AND depends_on_external IS NULL`).WithArgs("a", "b").WillReturnError(sql.ErrNoRows)
	facts, err := ApplyCanonicalDependencyMetadataInTx(context.Background(), tx, CanonicalDependencyMetadata{IssueID: "a", DependsOnID: "b", Type: types.DepRelated, Metadata: json.RawMessage(`{}`)}, "actor")
	if facts != nil || !errors.Is(err, storage.ErrNotFound) {
		t.Fatalf("missing pair created: %+v, %v", facts, err)
	}
	mock.ExpectRollback()
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCanonicalFinalDependencyTemplateOwnerRefused(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	values := issueRowValues("a", "template")
	for index, column := range issueColumns() {
		if column == "is_template" {
			values[index] = 1
		}
	}
	mock.ExpectQuery(`SELECT .* FROM issues .* WHERE id = \?`).WithArgs("a").WillReturnRows(issueRows().AddRow(values...))
	mock.ExpectQuery(`SELECT label FROM labels WHERE issue_id = \?`).WithArgs("a").WillReturnRows(sqlmock.NewRows([]string{"label"}))
	canonicalFinalNotWisp(mock, "a")
	facts, err := ApplyCanonicalDependencyMetadataInTx(context.Background(), tx, CanonicalDependencyMetadata{IssueID: "a", DependsOnID: "b", Type: types.DepRelated, Metadata: json.RawMessage(`{}`)}, "actor")
	var templateError *publicops.TemplateReadOnlyError
	if facts != nil || !errors.As(err, &templateError) {
		t.Fatalf("template owner accepted: %+v, %v", facts, err)
	}
	mock.ExpectRollback()
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCanonicalFinalIssueNoopAndFactScope(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	canonicalFinalHydration(mock, "a", "unchanged")
	canonicalFinalNotWisp(mock, "a")
	facts, err := ApplyCanonicalIssueContentInTx(context.Background(), tx, "a", CanonicalIssueContent{Title: publicops.Field[string]{Set: true, Value: "unchanged"}}, "actor")
	if err != nil {
		t.Fatal(err)
	}
	if facts.Changed() {
		t.Fatal("same final row changed")
	}
	if err := EmitCanonicalFinalFactsInTx(context.Background(), nil, facts); !errors.Is(err, storage.ErrValidation) {
		t.Fatalf("wrong transaction accepted: %v", err)
	}
	if err := EmitCanonicalFinalFactsInTx(context.Background(), tx, facts); err != nil {
		t.Fatal(err)
	}
	if err := EmitCanonicalFinalFactsInTx(context.Background(), tx, facts); !errors.Is(err, storage.ErrValidation) {
		t.Fatalf("duplicate emission accepted: %v", err)
	}
	mock.ExpectRollback()
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
