//go:build cgo

package graphstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/steveyegge/beads/internal/storage/domain"
	"github.com/steveyegge/beads/internal/storage/issueops"
	"github.com/steveyegge/beads/internal/types"
)

func nativeBatchCount(t *testing.T, ctx context.Context, s *Store, table, where string, args ...any) int {
	t.Helper()
	var count int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table+" WHERE "+where, args...).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}

func TestCanonicalBatchDependencyFinalReadinessAndOwners(t *testing.T) {
	ctx, _, s := atomicBatchFixture(t)
	ctx = issueops.WithEventsJournal(ctx, true)
	issues := map[string]IssueRecord{}
	for _, name := range []string{"source", "target", "neighbor", "other"} {
		record, err := s.CreateIssue(ctx, "beads/"+name, plainIssue(name))
		if err != nil {
			t.Fatal(err)
		}
		issues[name] = record
	}
	parent, err := s.AddDependency(ctx, DependencyRequest{SourcePath: fixtureResourcePath(t, ctx, s, issues["neighbor"].ID), TargetPath: fixtureResourcePath(t, ctx, s, issues["source"].ID), Type: types.DepParentChild, Actor: "setup"})
	if err != nil {
		t.Fatal(err)
	}
	issues["neighbor"] = parent.Source
	source, neighbor := issues["source"], issues["neighbor"]
	sourceVersions := nativeBatchCount(t, ctx, s, "issue_versions", "issue_id=?", source.Properties.ID)
	neighborVersions := nativeBatchCount(t, ctx, s, "issue_versions", "issue_id=?", neighbor.Properties.ID)
	journal := nativeBatchCount(t, ctx, s, "bd_events_journal", "1=1")
	input := fmt.Sprintf(`{"operation":"createLink","id":"new-block","type":"blocks","source":%q,"target":%q,"properties":{"dep_type":"blocks"}}
{"operation":"createLink","id":"new-related","type":"related","source":%q,"target":%q,"properties":{"dep_type":"related"}}
{"operation":"updateBead","bead":%q,"propertiesChange":[{"op":"replace","path":"/title","value":"final source"}]}`, source.ID, issues["target"].ID, source.ID, issues["other"].ID, source.ID)
	result, err := s.ApplyCanonicalBatch(ctx, atomicBatchCompile(t, s, input), "dependency-final", "batch-author")
	if err != nil {
		t.Fatal(err)
	}
	if !result.Changed || nativeBatchCount(t, ctx, s, "issue_versions", "issue_id=?", source.Properties.ID) != sourceVersions+1 || nativeBatchCount(t, ctx, s, "issue_versions", "issue_id=?", neighbor.Properties.ID) != neighborVersions {
		t.Fatal("direct owner did not mint once or derived-only neighbor minted")
	}
	assertDependencyReadiness(t, ctx, s, source, false)
	assertDependencyReadiness(t, ctx, s, neighbor, false)
	if nativeBatchCount(t, ctx, s, "bd_events_journal", "1=1") != journal+5 {
		t.Fatal("expected content, two final dep facts and two derived readiness facts")
	}
	var published IssueRecord
	if err := json.Unmarshal(atomicBatchOutcome(t, result, source.ID).Resource, &published); err != nil {
		t.Fatal(err)
	}
	if published.Properties.Title != "final source" || len(published.Owned) != 2 {
		t.Fatalf("owner did not retain complete final state: %+v", published)
	}
	for _, id := range []string{"new-block", "new-related"} {
		public := s.ScopeURL() + id
		var link LinkRecord
		if err := json.Unmarshal(atomicBatchOutcome(t, result, public).Resource, &link); err != nil {
			t.Fatal(err)
		}
		shown, err := s.ShowLink(ctx, fixtureResourcePath(t, ctx, s, public))
		if err != nil || !reflect.DeepEqual(shown, link) {
			t.Fatalf("ordinary native read differs: %+v %v", shown, err)
		}
		var found bool
		for _, raw := range published.Owned {
			var owned LinkRecord
			if err := json.Unmarshal(raw, &owned); err != nil {
				t.Fatal(err)
			}
			found = found || reflect.DeepEqual(owned, link)
		}
		if !found {
			t.Fatalf("owner captured an intermediate Link: %+v", link)
		}
	}
	for _, row := range result.Resources {
		if row.ID == neighbor.ID {
			t.Fatal("cache-only neighbor gained a Resource outcome")
		}
	}
	replay, err := s.ApplyCanonicalBatch(ctx, atomicBatchCompile(t, s, input), "dependency-final", "batch-author")
	if err != nil || !reflect.DeepEqual(replay, result) {
		t.Fatalf("native result replay changed: %+v %v", replay, err)
	}
	if nativeBatchCount(t, ctx, s, "issue_versions", "issue_id=?", source.Properties.ID) != sourceVersions+1 {
		t.Fatal("retained replay re-executed native writes")
	}
}

func TestCanonicalBatchDependencyReassertAndAllocationOnly(t *testing.T) {
	ctx, _, s := atomicBatchFixture(t)
	ctx = issueops.WithEventsJournal(ctx, true)
	a, err := s.CreateIssue(ctx, "beads/a", plainIssue("a"))
	if err != nil {
		t.Fatal(err)
	}
	b, err := s.CreateIssue(ctx, "beads/b", plainIssue("b"))
	if err != nil {
		t.Fatal(err)
	}
	old, err := s.AddDependency(ctx, DependencyRequest{Path: "links/old-native", SourcePath: fixtureResourcePath(t, ctx, s, a.ID), TargetPath: fixtureResourcePath(t, ctx, s, b.ID), Type: types.DepBlocks, Actor: "old-author"})
	if err != nil {
		t.Fatal(err)
	}
	a = old.Source
	oldPath := fixtureResourcePath(t, ctx, s, old.Link.ID)
	var nativeKey string
	if err := s.db.QueryRowContext(ctx, `SELECT backing_key FROM graph_preview_catalog WHERE path=?`, oldPath).Scan(&nativeKey); err != nil {
		t.Fatal(err)
	}
	versions := nativeBatchCount(t, ctx, s, "issue_versions", "issue_id=?", a.Properties.ID)
	input := fmt.Sprintf(`{"operation":"createLink","id":"new-native","type":"blocks","source":%q,"target":%q,"properties":{"dep_type":"blocks"},"metadata":{"fresh":true}}
{"operation":"deleteLink","link":%q}`, a.ID, b.ID, old.Link.ID)
	result, err := s.ApplyCanonicalBatch(ctx, atomicBatchCompile(t, s, input), "reassert-native", "new-author")
	if err != nil {
		t.Fatal(err)
	}
	if nativeBatchCount(t, ctx, s, "issue_versions", "issue_id=?", a.Properties.ID) != versions+1 {
		t.Fatal("reassert did not mint exactly one final owner")
	}
	var detached sql.NullString
	if err := s.db.QueryRowContext(ctx, `SELECT backing_key FROM graph_preview_catalog WHERE path=?`, oldPath).Scan(&detached); err != nil || detached.Valid {
		t.Fatalf("deleted allocation kept reused backing: %v %v", detached, err)
	}
	var retained any
	retained, err = s.ReadVersion(ctx, oldPath, old.Link.Revision)
	if err != nil || !reflect.DeepEqual(retained, old.Link) {
		t.Fatalf("old Link evidence lost: %+v %v", retained, err)
	}
	var newLink LinkRecord
	if err := json.Unmarshal(atomicBatchOutcome(t, result, s.ScopeURL()+"new-native").Resource, &newLink); err != nil {
		t.Fatal(err)
	}
	if newLink.ID == old.Link.ID || newLink.Attribution.Actor != "new-author" || string(newLink.Metadata) != `{"fresh":true}` {
		t.Fatalf("new assertion inherited old identity/facts: %+v", newLink)
	}
	var key, actor string
	if err := s.db.QueryRowContext(ctx, `SELECT id,created_by FROM dependencies WHERE issue_id=? AND depends_on_issue_id=?`, a.Properties.ID, b.Properties.ID).Scan(&key, &actor); err != nil || key != nativeKey || actor != "new-author" {
		t.Fatalf("native assertion not fresh at deterministic pair: %s %s %v", key, actor, err)
	}
	// Metadata A->B->A has no final row effect. A transient extra public identity
	// on the occupied pair is consumed, with no native assertion or version.
	versions = nativeBatchCount(t, ctx, s, "issue_versions", "issue_id=?", a.Properties.ID)
	journal := nativeBatchCount(t, ctx, s, "bd_events_journal", "1=1")
	allocated := fmt.Sprintf(`{"operation":"updateLink","link":%q,"metadata":{"temporary":true}}
{"operation":"updateLink","link":%q,"metadata":{"fresh":true}}
{"operation":"createLink","id":"transient-native","type":"blocks","source":%q,"target":%q,"properties":{"dep_type":"blocks"}}
{"operation":"deleteLink","link":"transient-native"}`, newLink.ID, newLink.ID, a.ID, b.ID)
	only, err := s.ApplyCanonicalBatch(ctx, atomicBatchCompile(t, s, allocated), "native-allocation-only", "new-author")
	if err != nil {
		t.Fatal(err)
	}
	if !only.Changed || atomicBatchOutcome(t, only, s.ScopeURL()+"transient-native").Outcome != "allocated" || nativeBatchCount(t, ctx, s, "issue_versions", "issue_id=?", a.Properties.ID) != versions || nativeBatchCount(t, ctx, s, "bd_events_journal", "1=1") != journal {
		t.Fatal("allocation-only identity wrote native row/history/facts")
	}
	if nativeBatchCount(t, ctx, s, "graph_preview_batch_allocations", "public_id=?", "transient-native") != 1 || nativeBatchCount(t, ctx, s, "dependencies", "issue_id=? AND depends_on_issue_id=?", a.Properties.ID, b.Properties.ID) != 1 {
		t.Fatal("allocated-only identity was lost or duplicate native pair created")
	}
}

func TestCanonicalBatchDependencyCycleAndLateRollback(t *testing.T) {
	ctx, _, s := atomicBatchFixture(t)
	ctx = issueops.WithEventsJournal(ctx, true)
	a, err := s.CreateIssue(ctx, "beads/a", plainIssue("a"))
	if err != nil {
		t.Fatal(err)
	}
	b, err := s.CreateIssue(ctx, "beads/b", plainIssue("b"))
	if err != nil {
		t.Fatal(err)
	}
	c, err := s.CreateIssue(ctx, "beads/c", plainIssue("c"))
	if err != nil {
		t.Fatal(err)
	}
	old, err := s.AddDependency(ctx, DependencyRequest{SourcePath: fixtureResourcePath(t, ctx, s, a.ID), TargetPath: fixtureResourcePath(t, ctx, s, b.ID), Type: types.DepBlocks, Actor: "setup"})
	if err != nil {
		t.Fatal(err)
	}
	before := workflowState(t, ctx, s)
	cycle := fmt.Sprintf(`{"operation":"createLink","id":"cycle","type":"blocks","source":%q,"target":%q,"properties":{"dep_type":"blocks"}}`, b.ID, a.ID)
	if _, err := s.ApplyCanonicalBatch(ctx, atomicBatchCompile(t, s, cycle), "native-cycle", "actor"); !errors.Is(err, domain.ErrDependencyCycle) {
		t.Fatalf("final scheduling cycle accepted/wrong diagnosis: %v", err)
	}
	if !reflect.DeepEqual(before, workflowState(t, ctx, s)) || nativeBatchCount(t, ctx, s, "graph_preview_batch_decisions", "decision_key=?", "native-cycle") != 0 {
		t.Fatal("cycle refusal left native/public effects")
	}
	fault := errors.New("final native Dependency decision fault")
	s.afterWrite = func(stage string) error {
		if stage == "batch-decision" {
			return fault
		}
		return nil
	}
	input := fmt.Sprintf(`{"operation":"createBead","id":"generic-rollback","type":"memory","properties":{"title":"rollback","body":""}}
{"operation":"createLink","id":"native-rollback","type":"blocks","source":%q,"target":%q,"properties":{"dep_type":"blocks"}}`, b.ID, c.ID)
	if _, err := s.ApplyCanonicalBatch(ctx, atomicBatchCompile(t, s, input), "native-final-fault", "actor"); !errors.Is(err, fault) {
		t.Fatalf("expected late native fault: %v", err)
	}
	s.afterWrite = nil
	if !reflect.DeepEqual(before, workflowState(t, ctx, s)) || nativeBatchCount(t, ctx, s, "graph_preview_batch_decisions", "decision_key=?", "native-final-fault") != 0 {
		t.Fatal("late fault left rows/cache/versions/allocations/facts/decision")
	}
	// The head reader alone can read a legacy row with a valid retained mapping;
	// batch native closure must still reject NULL participation before minting.
	var generation int64
	if err := s.db.QueryRowContext(ctx, `SELECT participation_generation FROM issues WHERE id=?`, b.Properties.ID).Scan(&generation); err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE issues SET participation_generation=NULL,updated_at=updated_at WHERE id=?`, b.Properties.ID); err != nil {
		t.Fatal(err)
	}
	before = workflowState(t, ctx, s)
	input = fmt.Sprintf(`{"operation":"createLink","id":"legacy-refusal","type":"related","source":%q,"target":%q,"properties":{"dep_type":"related"}}`, b.ID, c.ID)
	if _, err := s.ApplyCanonicalBatch(ctx, atomicBatchCompile(t, s, input), "native-legacy-refusal", "actor"); !errors.Is(err, ErrInvalidStore) {
		t.Fatalf("legacy mint-skip row admitted: %v", err)
	}
	if !reflect.DeepEqual(before, workflowState(t, ctx, s)) {
		t.Fatal("legacy admission failure left effects")
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE issues SET participation_generation=?,updated_at=updated_at WHERE id=?`, generation, b.Properties.ID); err != nil {
		t.Fatal(err)
	}
	// Reversing a scheduling edge is valid as one final graph, even though the
	// first authored create would cycle if singleton writers ran in command order.
	aVersions := nativeBatchCount(t, ctx, s, "issue_versions", "issue_id=?", a.Properties.ID)
	bVersions := nativeBatchCount(t, ctx, s, "issue_versions", "issue_id=?", b.Properties.ID)
	repair := fmt.Sprintf(`{"operation":"createLink","id":"reverse","type":"blocks","source":%q,"target":%q,"properties":{"dep_type":"blocks"}}
{"operation":"deleteLink","link":%q}`, b.ID, a.ID, old.Link.ID)
	if _, err := s.ApplyCanonicalBatch(ctx, atomicBatchCompile(t, s, repair), "native-final-repair", "actor"); err != nil {
		t.Fatal(err)
	}
	assertDependencyReadiness(t, ctx, s, a, true)
	assertDependencyReadiness(t, ctx, s, b, false)
	if nativeBatchCount(t, ctx, s, "issue_versions", "issue_id=?", a.Properties.ID) != aVersions+1 || nativeBatchCount(t, ctx, s, "issue_versions", "issue_id=?", b.Properties.ID) != bVersions+1 {
		t.Fatal("final graph repair did not publish each direct owner once")
	}
}

func TestCanonicalBatchDependencyMetadataSettlesWaitsFor(t *testing.T) {
	ctx, _, s := atomicBatchFixture(t)
	ctx = issueops.WithEventsJournal(ctx, true)
	issues := map[string]IssueRecord{}
	for _, name := range []string{"waiter", "spawner", "closed-child", "open-child"} {
		record, err := s.CreateIssue(ctx, "beads/"+name, plainIssue(name))
		if err != nil {
			t.Fatal(err)
		}
		issues[name] = record
	}
	if _, err := s.CloseIssue(ctx, fixtureResourcePath(t, ctx, s, issues["closed-child"].ID), "setup closed child", "setup"); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"closed-child", "open-child"} {
		if _, err := s.AddDependency(ctx, DependencyRequest{SourcePath: fixtureResourcePath(t, ctx, s, issues[name].ID), TargetPath: fixtureResourcePath(t, ctx, s, issues["spawner"].ID), Type: types.DepParentChild, Actor: "setup"}); err != nil {
			t.Fatal(err)
		}
	}
	relation, err := s.AddDependency(ctx, DependencyRequest{SourcePath: fixtureResourcePath(t, ctx, s, issues["waiter"].ID), TargetPath: fixtureResourcePath(t, ctx, s, issues["spawner"].ID), Type: types.DepWaitsFor, Actor: "setup"})
	if err != nil {
		t.Fatal(err)
	}
	assertDependencyReadiness(t, ctx, s, relation.Source, false)
	versions := nativeBatchCount(t, ctx, s, "issue_versions", "issue_id=?", relation.Source.Properties.ID)
	audits := nativeBatchCount(t, ctx, s, "events", "issue_id=? AND event_type=?", relation.Source.Properties.ID, types.EventDependencyAdded)
	input := fmt.Sprintf(`{"operation":"updateLink","link":%q,"metadata":{"gate":"any-children"}}`, relation.Link.ID)
	result, err := s.ApplyCanonicalBatch(ctx, atomicBatchCompile(t, s, input), "waits-for-gate", "actor")
	if err != nil {
		t.Fatal(err)
	}
	if !result.Changed || nativeBatchCount(t, ctx, s, "issue_versions", "issue_id=?", relation.Source.Properties.ID) != versions+1 {
		t.Fatal("readiness metadata was not final-published once")
	}
	assertDependencyReadiness(t, ctx, s, relation.Source, true)
	if nativeBatchCount(t, ctx, s, "events", "issue_id=? AND event_type=?", relation.Source.Properties.ID, types.EventDependencyAdded) != audits {
		t.Fatal("metadata replacement emitted a new-assertion human audit")
	}
}
