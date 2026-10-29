//go:build cgo

package graphstore

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/storage/issueops"
	"github.com/steveyegge/beads/internal/types"
)

func TestCanonicalBatchNativeIssueAndDependencyFinalPublication(t *testing.T) {
	ctx, _, s := atomicBatchFixture(t)
	ctx = issueops.WithEventsJournal(ctx, true)
	source, err := s.CreateIssue(ctx, "beads/native-source", plainIssue("Source"))
	if err != nil {
		t.Fatal(err)
	}
	target, err := s.CreateIssue(ctx, "beads/native-target", plainIssue("Target"))
	if err != nil {
		t.Fatal(err)
	}
	relation, err := s.AddDependency(ctx, DependencyRequest{Path: "links/native-related", SourcePath: fixtureResourcePath(t, ctx, s, source.ID), TargetPath: fixtureResourcePath(t, ctx, s, target.ID), Type: types.DepRelated, Actor: "test-author"})
	if err != nil {
		t.Fatal(err)
	}
	source = relation.Source
	count := func(table, where string, args ...any) int {
		t.Helper()
		var n int
		if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table+" WHERE "+where, args...).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	sourcePath := fixtureResourcePath(t, ctx, s, source.ID)
	linkPath := fixtureResourcePath(t, ctx, s, relation.Link.ID)
	nativeVersions := count("issue_versions", "issue_id=?", source.Properties.ID)
	ownerVersions := count("graph_preview_issue_versions", "path=?", sourcePath)
	linkVersions := count("graph_preview_versions", "path=?", linkPath)
	journalRows := count("bd_events_journal", "1=1")
	input := fmt.Sprintf(`{"operation":"updateBead","bead":%q,"propertiesChange":[{"op":"replace","path":"/title","value":"Temporary, never published"}]}
{"operation":"updateBead","bead":%q,"propertiesChange":[{"op":"replace","path":"/title","value":"Published"}],"metadata":{"reviewed":true}}
{"operation":"updateLink","link":%q,"metadata":{"review":"final"}}`, source.ID, source.ID, relation.Link.ID)
	result, err := s.ApplyCanonicalBatch(ctx, atomicBatchCompile(t, s, input), "native-final", "test-author")
	if err != nil {
		t.Fatal(err)
	}
	if !result.Changed || len(result.Results) != 3 || count("issue_versions", "issue_id=?", source.Properties.ID) != nativeVersions+1 || count("graph_preview_issue_versions", "path=?", sourcePath) != ownerVersions+1 || count("graph_preview_versions", "path=?", linkPath) != linkVersions+1 {
		t.Fatalf("native state was not published exactly once: %+v", result)
	}
	if count("bd_events_journal", "1=1") != journalRows+2 {
		t.Fatal("final Issue update and Dependency replacement did not journal exactly once")
	}
	var published IssueRecord
	if err := json.Unmarshal(atomicBatchOutcome(t, result, source.ID).Resource, &published); err != nil {
		t.Fatal(err)
	}
	if published.Type != source.Type || published.Properties.Title != "Published" || string(published.Metadata) != `{"reviewed":true}` || len(published.Owned) != 1 {
		t.Fatalf("native record lost classification/content/owned state: %+v", published)
	}
	var owned LinkRecord
	if err := json.Unmarshal(published.Owned[0], &owned); err != nil {
		t.Fatal(err)
	}
	link := atomicBatchOutcome(t, result, relation.Link.ID)
	var finalLink LinkRecord
	if err := json.Unmarshal(link.Resource, &finalLink); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(owned, finalLink) || string(finalLink.Metadata) != `{"review":"final"}` || len(link.OwnerRevisions) != 1 || link.OwnerRevisions[0].Revision != published.Revision {
		t.Fatalf("native owner retained an intermediate Link: %+v %+v", published, link)
	}
	shown, err := s.ShowIssue(ctx, sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	// Native in-process Issues also carry fields such as RowVersion tagged
	// json:"-". The retained batch result is a wire record, so compare the
	// complete canonical wire projections rather than private hydration state.
	shownJSON, err := canonicalJSON(shown)
	if err != nil {
		t.Fatal(err)
	}
	publishedJSON, err := canonicalJSON(published)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(shownJSON, publishedJSON) {
		t.Fatalf("ordinary native read differs from final result: shown=%s published=%s", shownJSON, publishedJSON)
	}
	// A later decision must not alter the original retained result.
	later := fmt.Sprintf(`{"operation":"updateBead","bead":%q,"propertiesChange":[{"op":"replace","path":"/title","value":"Later"}]}`, source.ID)
	if _, err := s.ApplyCanonicalBatch(ctx, atomicBatchCompile(t, s, later), "native-later", "test-author"); err != nil {
		t.Fatal(err)
	}
	replayed, err := s.ApplyCanonicalBatch(ctx, atomicBatchCompile(t, s, input), "native-final", "test-author")
	if err != nil || !reflect.DeepEqual(result, replayed) {
		t.Fatalf("native recovery reexecuted: %+v %v", replayed, err)
	}
}

func TestCanonicalBatchNativeIssueNoopPolicyAndRollback(t *testing.T) {
	ctx, _, s := atomicBatchFixture(t)
	ctx = issueops.WithEventsJournal(ctx, true)
	source, err := s.CreateIssue(ctx, "beads/native-source", plainIssue("Source"))
	if err != nil {
		t.Fatal(err)
	}
	input := fmt.Sprintf(`{"operation":"updateBead","bead":%q,"propertiesChange":[{"op":"replace","path":"/title","value":"Temporary"}],"metadataChange":[{"op":"add","path":"/temporary","value":true}]}
{"operation":"updateBead","bead":%q,"propertiesChange":[{"op":"replace","path":"/title","value":"Source"}],"metadataChange":[{"op":"remove","path":"/temporary"}]}`, source.ID, source.ID)
	target, err := s.CreateIssue(ctx, "beads/native-target", plainIssue("Target"))
	if err != nil {
		t.Fatal(err)
	}
	relation, err := s.AddDependency(ctx, DependencyRequest{Path: "links/native-related", SourcePath: fixtureResourcePath(t, ctx, s, source.ID), TargetPath: fixtureResourcePath(t, ctx, s, target.ID), Type: types.DepRelated, Actor: "test-author"})
	if err != nil {
		t.Fatal(err)
	}
	// Input admission still applies when publication would otherwise skip an
	// unchanged or canceled native mutation. A refusal retains no decision.
	before := workflowState(t, ctx, s)
	for index, actorless := range []string{
		input,
		fmt.Sprintf(`{"operation":"updateBead","bead":%q,"propertiesChange":[]}`, source.ID),
		fmt.Sprintf(`{"operation":"updateLink","link":%q,"metadata":{}}`, relation.Link.ID),
		fmt.Sprintf(`{"operation":"updateLink","link":%q,"metadata":{"temporary":true}}`+"\n"+`{"operation":"updateLink","link":%q,"metadata":{}}`, relation.Link.ID, relation.Link.ID),
	} {
		if _, err := s.ApplyCanonicalBatch(ctx, atomicBatchCompile(t, s, actorless), fmt.Sprintf("native-no-actor-%d", index), ""); !errors.Is(err, storage.ErrValidation) {
			t.Fatalf("actorless native no-op admitted: %v", err)
		}
		if after := workflowState(t, ctx, s); !reflect.DeepEqual(before, after) {
			t.Fatal("actorless native refusal retained effects or a decision")
		}
	}
	result, err := s.ApplyCanonicalBatch(ctx, atomicBatchCompile(t, s, input), "native-noop", "test-author")
	if err != nil || result.Changed || result.Transaction != "" {
		t.Fatalf("native canceled changes published: %+v %v", result, err)
	}
	after := workflowState(t, ctx, s)
	delete(before, "graph_preview_scope") // Private decision coordination only.
	delete(after, "graph_preview_scope")
	if !reflect.DeepEqual(before, after) {
		t.Fatal("native no-op changed row locks, attribution, audit, journal or history")
	}
	// Static policy sees the forbidden authored status before cancellation.
	forbidden := fmt.Sprintf(`{"operation":"updateBead","bead":%q,"propertiesChange":[{"op":"replace","path":"/status","value":"closed"}]}
{"operation":"updateBead","bead":%q,"propertiesChange":[{"op":"replace","path":"/status","value":"open"}]}`, source.ID, source.ID)
	before = workflowState(t, ctx, s)
	if _, err := s.ApplyCanonicalBatch(ctx, atomicBatchCompile(t, s, forbidden), "native-protected", "test-author"); !errors.Is(err, ErrCapabilityUnavailable) {
		t.Fatalf("canceled protected native input was admitted: %v", err)
	}
	if after := workflowState(t, ctx, s); !reflect.DeepEqual(before, after) {
		t.Fatal("refused native input left effects")
	}
	// A failure after native final versions and facts rolls the entire composite
	// back, including an earlier generic allocation and successful-result row.
	fault := errors.New("native final result must roll back")
	s.afterWrite = func(stage string) error {
		if stage == "batch-decision" {
			return fault
		}
		return nil
	}
	mixed := `{"operation":"createBead","id":"must-rollback","type":"memory","properties":{"title":"Memory","body":""}}` + "\n" + strings.ReplaceAll(input[:strings.Index(input, "\n")], "Temporary", "Changed")
	if _, err := s.ApplyCanonicalBatch(ctx, atomicBatchCompile(t, s, mixed), "native-fault", "test-author"); !errors.Is(err, fault) {
		t.Fatalf("expected final fault, got %v", err)
	}
	s.afterWrite = nil
	if after := workflowState(t, ctx, s); !reflect.DeepEqual(before, after) {
		t.Fatal("native final failure left rows, versions, journal, audit or generic state")
	}
	if _, err := s.ResolveExactAllocatedResourcePath(ctx, "must-rollback"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("rollback left a generic allocation: %v", err)
	}
}
