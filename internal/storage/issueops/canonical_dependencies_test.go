package issueops

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/storage/depid"
	"github.com/steveyegge/beads/internal/storage/domain"
	"github.com/steveyegge/beads/internal/types"
)

func dependencyFinalRow(source, target string, kind types.DependencyType, metadata string) *types.Dependency {
	return &types.Dependency{ID: depid.New(source, target), IssueID: source, DependsOnID: target, Type: kind, Metadata: metadata, CreatedBy: "before", CreatedAt: time.Date(2026, 10, 29, 1, 0, 0, 0, time.UTC)}
}

func dependencyFinalInventory(mock sqlmock.Sqlmock, ids []string, deps ...*types.Dependency) {
	rows := sqlmock.NewRows([]string{"id", "ephemeral", "no_history", "wisp_type", "is_template"})
	for _, id := range ids {
		rows.AddRow(id, false, false, nil, false)
	}
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, ephemeral, no_history, wisp_type, is_template FROM issues LIMIT 1001`)).WillReturnRows(rows)
	for _, table := range []string{"wisps", "wisp_dependencies"} {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM ` + table)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	}
	rows = sqlmock.NewRows([]string{"id", "issue_id", "depends_on_issue_id", "depends_on_wisp_id", "depends_on_external", "type", "created_at", "created_by", "metadata", "thread_id"})
	for _, d := range deps {
		rows.AddRow(d.ID, d.IssueID, d.DependsOnID, nil, nil, string(d.Type), d.CreatedAt, d.CreatedBy, d.Metadata, d.ThreadID)
	}
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, issue_id, depends_on_issue_id, depends_on_wisp_id, depends_on_external, type, created_at, created_by, metadata, thread_id FROM dependencies LIMIT 1001`)).WillReturnRows(rows)
}

func dependencyFinalConsistent(mock sqlmock.Sqlmock, count int64) {
	for _, spec := range []struct{ table, dep, target string }{{"issues", "dependencies", "depends_on_issue_id"}, {"wisps", "wisp_dependencies", "depends_on_wisp_id"}} {
		mock.ExpectQuery(regexp.QuoteMeta(parentsExplainedBySubtreeSQL(spec.table, spec.dep, spec.target, unbatchedCandSources()))).WillReturnRows(sqlmock.NewRows([]string{"id"}))
	}
	mock.ExpectQuery(regexp.QuoteMeta(countStaleIsBlockedSQL("issues", "i", "dependencies", subtreeExplainedParents{}))).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(count))
	mock.ExpectQuery(regexp.QuoteMeta(countStaleIsBlockedSQL("wisps", "w", "wisp_dependencies", subtreeExplainedParents{}))).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
}

func dependencyFinalRecompute(mock sqlmock.Sqlmock, ids []string) {
	// A small complete population has one existing-kernel mark/unmark pass.
	mock.ExpectQuery(`SELECT COUNT\(d.depends_on_issue_id\), COUNT\(d.depends_on_wisp_id\)`).WillReturnRows(sqlmock.NewRows([]string{"issues", "wisps"}).AddRow(0, 0))
	placeholders, args := buildSQLInClause(ids)
	for _, template := range []string{markBlockedTemplate(issuesBlockedSpec, subtreeExplainedParents{}), unmarkBlockedTemplate(issuesBlockedSpec, subtreeExplainedParents{})} {
		statement, _, err := expandBatchTemplate(template, placeholders, args, nil)
		if err != nil {
			panic(err)
		}
		mock.ExpectExec(regexp.QuoteMeta(statement)).WillReturnResult(sqlmock.NewResult(0, 0))
	}
}

func TestCanonicalDependencyFinalMetadataPhases(t *testing.T) {
	_, mock, tx := beginMockTx(t)
	ctx := context.Background()
	cleanupJournal := ScopeEventsJournalTransaction(tx, true)
	defer cleanupJournal()
	cleanupVersions := ScopeVersionedHistoryTransaction(tx, true)
	defer cleanupVersions()
	ids := []string{"a", "b"}
	before := dependencyFinalRow("a", "b", types.DepRelated, `{"old":true}`)
	dependencyFinalInventory(mock, ids, before)
	dependencyFinalConsistent(mock, 0)
	mock.ExpectQuery(`SELECT id, is_blocked FROM issues WHERE id IN`).WithArgs("a", "b").WillReturnRows(sqlmock.NewRows([]string{"id", "is_blocked"}).AddRow("a", 0).AddRow("b", 0))
	for _, id := range ids {
		mock.ExpectQuery(`SELECT issue_type FROM issues WHERE id = \?`).WithArgs(id).WillReturnRows(sqlmock.NewRows([]string{"issue_type"}).AddRow("task"))
	}
	mock.ExpectQuery(`SELECT type, metadata, thread_id FROM dependencies WHERE`).WithArgs("a", "b").WillReturnRows(sqlmock.NewRows([]string{"type", "metadata", "thread_id"}).AddRow("related", before.Metadata, nil))
	mock.ExpectExec(`UPDATE dependencies SET metadata = \?, thread_id = \? WHERE`).WithArgs(`{"final":true}`, nil, "a", "b").WillReturnResult(sqlmock.NewResult(0, 1))
	token, err := ApplyCanonicalDependencyFinalRowsInTx(ctx, tx, CanonicalDependencyFinalRequest{Actor: "actor", Population: ids, Metadata: []CanonicalDependencyMetadataFinal{{Existing: CanonicalExistingDependency{Pair: CanonicalDependencyPair{"a", "b"}, NativeID: before.ID, Type: before.Type}, Metadata: json.RawMessage(`{"final":true}`)}}})
	if err != nil {
		t.Fatal(err)
	}
	if !journalEnabled(ctx, tx) || !VersionedHistoryEnabledFor(tx) {
		t.Fatal("composite disabled transaction publication")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal("intermediate facts or mints: ", err)
	}
	after := cloneCanonicalDependency(before)
	after.Metadata = `{"final":true}`
	dependencyFinalInventory(mock, ids, after)
	dependencyFinalRecompute(mock, ids)
	final, err := SettleCanonicalDependencyReadinessInTx(ctx, tx, token)
	if err != nil {
		t.Fatal(err)
	}
	if got := final.OwnerIDs(); !reflect.DeepEqual(got, []string{"a"}) {
		t.Fatalf("owners %v", got)
	}
	changes := final.RowChanges()
	if len(changes) != 1 || changes[0].Before.Metadata != before.Metadata || changes[0].After.Metadata != after.Metadata {
		t.Fatalf("wrong final actual rows: %+v", changes)
	}
	changes[0].After.Metadata = `{"tampered":true}`
	if final.RowChanges()[0].After.Metadata != after.Metadata {
		t.Fatal("caller modified private facts")
	}
	if again, err := SettleCanonicalDependencyReadinessInTx(ctx, tx, token); again != nil || !errors.Is(err, storage.ErrValidation) {
		t.Fatalf("phase reused: %+v %v", again, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal("settlement emitted facts or versions: ", err)
	}
	canonicalFinalJournal(mock, "existing", "dep_add", `{"kind":"related","target":"b","metadata":"{\"final\":true}"}`)
	mock.ExpectQuery(`SELECT id, is_blocked FROM issues WHERE id IN`).WithArgs("a", "b").WillReturnRows(sqlmock.NewRows([]string{"id", "is_blocked"}).AddRow("a", 0).AddRow("b", 0))
	if err := EmitCanonicalDependencyFinalFactsInTx(ctx, tx, final); err != nil {
		t.Fatal(err)
	}
	if err := EmitCanonicalDependencyFinalFactsInTx(ctx, tx, final); !errors.Is(err, storage.ErrValidation) {
		t.Fatalf("facts reused: %v", err)
	}
	mock.ExpectRollback()
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCanonicalDependencyFinalBaselineRefusalsBeforeEffects(t *testing.T) {
	for _, reason := range []string{"missing-pair", "existing-insert", "wrong-classification", "stale-cache", "template", "thread", "wisp", "population"} {
		t.Run(reason, func(t *testing.T) {
			_, mock, tx := beginMockTx(t)
			ids := []string{"a", "b"}
			row := dependencyFinalRow("a", "b", types.DepRelated, `{}`)
			request := CanonicalDependencyFinalRequest{Actor: "writer", Population: ids, Delete: []CanonicalExistingDependency{{Pair: CanonicalDependencyPair{"a", "b"}, NativeID: row.ID, Type: row.Type}}}
			switch reason {
			case "missing-pair":
				dependencyFinalInventory(mock, ids)
			case "existing-insert":
				request.Delete = nil
				request.Insert = []CanonicalDependencyInsertFinal{{Pair: CanonicalDependencyPair{"a", "b"}, Type: row.Type}}
				dependencyFinalInventory(mock, ids, row)
			case "wrong-classification":
				request.Delete[0].Type = types.DepBlocks
				dependencyFinalInventory(mock, ids, row)
			case "stale-cache":
				dependencyFinalInventory(mock, ids, row)
				dependencyFinalConsistent(mock, 1)
			case "template":
				mock.ExpectQuery(`SELECT id, ephemeral, no_history, wisp_type, is_template FROM issues`).WillReturnRows(sqlmock.NewRows([]string{"id", "ephemeral", "no_history", "wisp_type", "is_template"}).AddRow("a", false, false, nil, true).AddRow("b", false, false, nil, false))
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM wisps`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM wisp_dependencies`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				mock.ExpectQuery(`SELECT id, issue_id, depends_on_issue_id`).WillReturnRows(sqlmock.NewRows([]string{"id", "source", "target", "wisp", "external", "type", "created", "actor", "metadata", "thread"}).AddRow(row.ID, "a", "b", nil, nil, "related", row.CreatedAt, "before", "{}", nil))
			case "thread":
				row.ThreadID = "unsupported"
				dependencyFinalInventory(mock, ids, row)
			case "wisp":
				mock.ExpectQuery(`SELECT id, ephemeral, no_history, wisp_type, is_template FROM issues`).WillReturnRows(sqlmock.NewRows([]string{"id", "ephemeral", "no_history", "wisp_type", "is_template"}).AddRow("a", false, false, nil, false).AddRow("b", false, false, nil, false))
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM wisps`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			case "population":
				mock.ExpectQuery(`SELECT id, ephemeral, no_history, wisp_type, is_template FROM issues`).WillReturnRows(sqlmock.NewRows([]string{"id", "ephemeral", "no_history", "wisp_type", "is_template"}).AddRow("a", false, false, nil, false))
			}
			result, err := ApplyCanonicalDependencyFinalRowsInTx(context.Background(), tx, request)
			if result != nil || err == nil {
				t.Fatalf("baseline %s accepted: %+v %v", reason, result, err)
			}
			mock.ExpectRollback()
			if err := tx.Rollback(); err != nil {
				t.Fatal(err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestCanonicalDependencyFinalRowAssertions(t *testing.T) {
	a := dependencyFinalRow("a", "b", types.DepBlocks, `{"a":1}`)
	pair := CanonicalDependencyPair{"a", "b"}
	before := map[CanonicalDependencyPair]*types.Dependency{pair: a}
	plan := CanonicalDependencyIntentPlan{Actor: "writer", Changes: []CanonicalDependencyIntent{{Transition: CanonicalDependencyMetadataTransition, Pair: pair, Type: a.Type, Metadata: json.RawMessage(`{"a":1}`)}}}
	if got, err := checkCanonicalDependencyFinalRows(plan, before, before); err != nil || len(got) != 0 {
		t.Fatalf("semantic no-op: %+v %v", got, err)
	}
	after := cloneCanonicalDependency(a)
	after.CreatedBy = "writer"
	plan.Changes[0].Transition = CanonicalDependencyReassertTransition
	got, err := checkCanonicalDependencyFinalRows(plan, before, map[CanonicalDependencyPair]*types.Dependency{pair: after})
	if err != nil || len(got) != 1 {
		t.Fatalf("reassert collapsed: %+v %v", got, err)
	}
	for _, mutation := range []func(*types.Dependency){func(d *types.Dependency) { d.Type = types.DepRelated }, func(d *types.Dependency) { d.Metadata = `{"lost":true}` }, func(d *types.Dependency) { d.CreatedBy = "wrong" }, func(d *types.Dependency) { d.ThreadID = "old-thread" }} {
		broken := cloneCanonicalDependency(after)
		mutation(broken)
		if _, err := checkCanonicalDependencyFinalRows(plan, before, map[CanonicalDependencyPair]*types.Dependency{pair: broken}); err == nil {
			t.Fatalf("lost final assertion accepted: %+v", broken)
		}
	}
	plan.Changes[0].Transition = CanonicalDependencyDeleteTransition
	if got, err := checkCanonicalDependencyFinalRows(plan, before, nil); err != nil || len(got) != 1 || got[0].After != nil {
		t.Fatalf("delete: %+v %v", got, err)
	}
	if _, err := checkCanonicalDependencyFinalRows(plan, before, before); err == nil {
		t.Fatal("retained deletion accepted")
	}
	plan.Changes = nil
	if _, err := checkCanonicalDependencyFinalRows(plan, before, map[CanonicalDependencyPair]*types.Dependency{pair: after}); err == nil {
		t.Fatal("unplanned mutation accepted")
	}
}

func TestCanonicalDependencyFinalPhaseFailuresConsumeToken(t *testing.T) {
	_, mock, tx := beginMockTx(t)
	_, _, other := beginMockTx(t)
	token := &CanonicalDependencyRows{tx: tx, plan: CanonicalDependencyIntentPlan{Population: []string{"a"}}}
	if _, err := SettleCanonicalDependencyReadinessInTx(context.Background(), other, token); !errors.Is(err, storage.ErrValidation) || token.consumed {
		t.Fatalf("cross tx token: %v", err)
	}
	sentinel := errors.New("late native read failure")
	mock.ExpectQuery(`SELECT id, ephemeral, no_history, wisp_type, is_template FROM issues`).WillReturnError(sentinel)
	if final, err := SettleCanonicalDependencyReadinessInTx(context.Background(), tx, token); final != nil || !errors.Is(err, sentinel) || !token.consumed {
		t.Fatalf("failed phase: %+v %v", final, err)
	}
	if _, err := SettleCanonicalDependencyReadinessInTx(context.Background(), tx, token); !errors.Is(err, storage.ErrValidation) {
		t.Fatalf("failed token reused: %v", err)
	}
	if err := EmitCanonicalDependencyFinalFactsInTx(context.Background(), other, &CanonicalDependencyFinalization{tx: tx}); !errors.Is(err, storage.ErrValidation) {
		t.Fatalf("cross tx facts: %v", err)
	}
	mock.ExpectRollback()
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCanonicalDependencyPrivateFinalPolicyAndOrdinaryChecks(t *testing.T) {
	for _, final := range []bool{false, true} {
		t.Run(fmt.Sprintf("final=%v", final), func(t *testing.T) {
			_, mock, tx := beginMockTx(t)
			dep := dependencyFinalRow("a", "b", types.DepBlocks, `{}`)
			for _, id := range []string{"a", "b"} {
				mock.ExpectQuery(`SELECT issue_type FROM issues WHERE id = \?`).WithArgs(id).WillReturnRows(sqlmock.NewRows([]string{"type"}).AddRow("task"))
			}
			kind := DepTargetIssue
			opts := AddDependencyOpts{SourceTable: "issues", TargetTable: "issues", WriteTable: "dependencies", TargetKind: &kind}
			if !final {
				sentinel := errors.New("ordinary hierarchy probe")
				mock.ExpectQuery(`WITH RECURSIVE`).WillReturnError(sentinel)
				if _, err := addDependencyInTx(context.Background(), tx, dep, "writer", opts, nil, false); !errors.Is(err, sentinel) {
					t.Fatalf("ordinary skipped graph validation: %v", err)
				}
			} else {
				cleanup := ScopeEventsJournalTransaction(tx, true)
				defer cleanup()
				cleanupVersion := ScopeVersionedHistoryTransaction(tx, true)
				defer cleanupVersion()
				mock.ExpectQuery(`SELECT type, metadata, thread_id FROM dependencies`).WillReturnError(sql.ErrNoRows)
				mock.ExpectExec(`INSERT INTO dependencies`).WithArgs(dep.ID, "a", "b", types.DepBlocks, "writer", "{}", "").WillReturnResult(sqlmock.NewResult(0, 1))
				if _, err := addDependencyWithPolicyInTx(context.Background(), tx, dep, "writer", opts, nil, false, dependencyCanonicalFinalRows); err != nil {
					t.Fatal(err)
				}
				if !journalEnabled(context.Background(), tx) || !VersionedHistoryEnabledFor(tx) {
					t.Fatal("transaction settings changed")
				}
			}
			mock.ExpectRollback()
			if err := tx.Rollback(); err != nil {
				t.Fatal(err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestCanonicalDependencyFinalSchedulingGraphUsesNativeKinds(t *testing.T) {
	_, _, tx := beginMockTx(t)
	rows := map[CanonicalDependencyPair]*types.Dependency{
		{"a", "b"}: dependencyFinalRow("a", "b", types.DepWaitsFor, `{}`),
		{"b", "a"}: dependencyFinalRow("b", "a", types.DepWaitsFor, `{}`),
	}
	if err := validateCanonicalDependencyGraph(context.Background(), tx, rows); err != nil {
		t.Fatalf("waits-for confused with scheduling cycle: %v", err)
	}
	rows[CanonicalDependencyPair{"a", "b"}].Type = types.DepParentChild
	rows[CanonicalDependencyPair{"b", "a"}].Type = types.DepBlocks
	if err := validateCanonicalDependencyGraph(context.Background(), tx, rows); !errors.Is(err, storage.ErrValidation) || !errors.Is(err, domain.ErrDependencyCycle) {
		t.Fatalf("mixed scheduling cycle accepted: %v", err)
	}
}

func TestCanonicalDependencyFinalReassertDefersRemoveAndAddFacts(t *testing.T) {
	_, mock, tx := beginMockTx(t)
	ctx := context.Background()
	cleanupVersions := ScopeVersionedHistoryTransaction(tx, true)
	defer cleanupVersions()
	before := dependencyFinalRow("a", "b", types.DepRelated, `{"old":true}`)
	ids := []string{"a", "b"}
	dependencyFinalInventory(mock, ids, before)
	dependencyFinalConsistent(mock, 0)
	canonicalFinalNotWisp(mock, "a")
	mock.ExpectQuery(`SELECT type, metadata FROM dependencies WHERE`).WithArgs("a", "b").WillReturnRows(sqlmock.NewRows([]string{"type", "metadata"}).AddRow("related", before.Metadata))
	mock.ExpectExec(`DELETE FROM dependencies WHERE`).WithArgs("a", "b").WillReturnResult(sqlmock.NewResult(0, 1))
	for _, id := range ids {
		mock.ExpectQuery(`SELECT issue_type FROM issues WHERE id = \?`).WithArgs(id).WillReturnRows(sqlmock.NewRows([]string{"type"}).AddRow("task"))
	}
	mock.ExpectQuery(`SELECT type, metadata, thread_id FROM dependencies WHERE`).WithArgs("a", "b").WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(`INSERT INTO dependencies`).WithArgs(before.ID, "a", "b", types.DepRelated, "writer", `{"new":true}`, "").WillReturnResult(sqlmock.NewResult(0, 1))
	request := CanonicalDependencyFinalRequest{Actor: "writer", Population: ids, Reassert: []CanonicalDependencyReassertFinal{{Existing: CanonicalExistingDependency{Pair: CanonicalDependencyPair{"a", "b"}, NativeID: before.ID, Type: before.Type}, Metadata: json.RawMessage(`{"new":true}`)}}}
	token, err := ApplyCanonicalDependencyFinalRowsInTx(ctx, tx, request)
	if err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal("intermediate publication: ", err)
	}
	after := cloneCanonicalDependency(before)
	after.Metadata = `{"new":true}`
	after.CreatedBy = "writer"
	after.CreatedAt = before.CreatedAt.Add(time.Second)
	dependencyFinalInventory(mock, ids, after)
	dependencyFinalRecompute(mock, ids)
	final, err := SettleCanonicalDependencyReadinessInTx(ctx, tx, token)
	if err != nil {
		t.Fatal(err)
	}
	changes := final.RowChanges()
	if len(changes) != 1 || !changes[0].Before.CreatedAt.Equal(before.CreatedAt) || !changes[0].After.CreatedAt.Equal(after.CreatedAt) || changes[0].After.CreatedBy != "writer" || !reflect.DeepEqual(final.OwnerIDs(), []string{"a"}) {
		t.Fatalf("reassert actual facts lost: %+v", changes)
	}
	// Journal disabled for this tx: ordinary human audits still occur, once for
	// the actual removed assertion and once for the new assertion, after settle.
	for range 2 {
		mock.ExpectQuery(`SELECT id FROM events`).WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock.ExpectExec(`INSERT INTO events`).WillReturnResult(sqlmock.NewResult(0, 1))
	}
	if err := EmitCanonicalDependencyFinalFactsInTx(ctx, tx, final); err != nil {
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

func TestCanonicalDependencyFinalLocalOpaqueParentTarget(t *testing.T) {
	_, mock, tx := beginMockTx(t)
	target := "external:actually-local"
	dep := dependencyFinalRow("a", target, types.DepParentChild, `{}`)
	mock.ExpectQuery(`SELECT issue_type FROM issues WHERE id = \?`).WithArgs("a").WillReturnRows(sqlmock.NewRows([]string{"type"}).AddRow("task"))
	mock.ExpectQuery(`SELECT type, metadata, thread_id FROM dependencies WHERE`).WithArgs("a", target).WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(`INSERT INTO dependencies`).WithArgs(dep.ID, "a", target, types.DepParentChild, "writer", "{}", "").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta("REPLACE INTO local_metadata (`key`, value) VALUES (?, ?)")).WithArgs(dependencyCoordinationKey(target, "dependencies"), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 1))
	kind := DepTargetIssue
	if _, err := addDependencyWithPolicyInTx(context.Background(), tx, dep, "writer", AddDependencyOpts{SourceTable: "issues", TargetTable: "issues", WriteTable: "dependencies", TargetKind: &kind}, nil, false, dependencyCanonicalFinalRows); err != nil {
		t.Fatal("complete local proof rejected because of spelling: ", err)
	}
	// The ordinary caller retains its external-capability prohibition.
	if _, err := addDependencyInTx(context.Background(), tx, dep, "writer", AddDependencyOpts{}, nil, false); err == nil {
		t.Fatal("ordinary external hierarchy policy changed")
	}
	mock.ExpectRollback()
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
