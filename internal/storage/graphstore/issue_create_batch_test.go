//go:build cgo

package graphstore

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/steveyegge/beads/internal/types"
	publicops "github.com/steveyegge/beads/issueops"
)

// A fault after the first catalog row is stronger than rejecting input up
// front: the native batch writer and the first mapping have both run, yet the
// transaction must restore every Issue, retained version and coordination row.
func TestGraphIssueCreateBatchAtomicMappingAndInitialVersions(t *testing.T) {
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
			request := publicops.CreateBatchRequest{Actor: "test-author", Items: []publicops.BatchCreateItem{
				{Issue: plainIssue("First heading").Issue},
				{Issue: plainIssue("Second heading").Issue},
			}}
			before := workflowState(t, ctx, s)
			fault := errors.New("after first Issue catalog mapping")
			catalogWrites := 0
			s.afterWrite = func(stage string) error {
				if stage == "issue-catalog" {
					catalogWrites++
					if catalogWrites == 1 {
						return fault
					}
				}
				return nil
			}
			failed, err := s.CreateIssueBatch(ctx, request)
			s.afterWrite = nil
			if !errors.Is(err, fault) || failed != nil || catalogWrites != 1 {
				t.Fatalf("first-mapping fault: records=%+v err=%v hook calls=%d", failed, err, catalogWrites)
			}
			if after := workflowState(t, ctx, s); !reflect.DeepEqual(before, after) {
				t.Fatal("failed batch left native Issue, History, catalog, retained version, or coordination state")
			}
			for _, table := range []string{"issues", "issue_versions", "graph_preview_catalog", "graph_preview_issue_versions"} {
				if got := graphBatchRowCount(t, ctx, s, table); got != 0 {
					t.Fatalf("failed batch left %d rows in %s", got, table)
				}
			}

			records, err := s.CreateIssueBatch(ctx, request)
			if err != nil || len(records) != 2 {
				t.Fatalf("successful batch: records=%+v err=%v", records, err)
			}
			for _, record := range records {
				path := fixtureResourcePath(t, ctx, s, record.ID)
				var nativeCount, mappingCount, ordinal, current int
				if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM issue_versions WHERE issue_id=?`, record.Properties.ID).Scan(&nativeCount); err != nil {
					t.Fatal(err)
				}
				if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM graph_preview_issue_versions WHERE path=?`, path).Scan(&mappingCount); err != nil {
					t.Fatal(err)
				}
				if err := s.db.QueryRowContext(ctx, `SELECT current_revision FROM issues WHERE id=?`, record.Properties.ID).Scan(&current); err != nil {
					t.Fatal(err)
				}
				if err := s.db.QueryRowContext(ctx, `SELECT issue_revision FROM graph_preview_issue_versions WHERE path=?`, path).Scan(&ordinal); err != nil {
					t.Fatal(err)
				}
				if nativeCount != 1 || mappingCount != 1 || ordinal != current {
					t.Fatalf("Issue %s has native versions=%d mappings=%d mapped ordinal=%d current=%d", record.ID, nativeCount, mappingCount, ordinal, current)
				}
			}
			if got := graphBatchRowCount(t, ctx, s, "issue_versions"); got != 2 {
				t.Fatalf("successful batch wrote %d native initial versions, want 2", got)
			}
		})
	}
}

// A Memory's legacy graph path is also its public ID. A later Issue batch
// cannot claim that spelling, even when an earlier item in the same batch is
// otherwise valid; refusal must leave the whole transaction unchanged.
func TestGraphIssueCreateBatchRefusesReservedMemoryIDAtomically(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		for _, state := range []string{"live", "deleted"} {
			t.Run(backend+"/"+state, func(t *testing.T) {
				ctx, _, store := memoryDeleteStore(t, backend)
				const reservedID = "beads/exp-batch-reservation"
				if _, err := store.Create(ctx, CreateRequest{Path: reservedID, Title: "Reserved Memory identity"}); err != nil {
					t.Fatal(err)
				}
				if state == "deleted" {
					if _, err := store.DeleteMemory(ctx, MemoryDeleteRequest{Path: reservedID, Unconditional: true}); err != nil {
						t.Fatal(err)
					}
				}
				before := workflowState(t, ctx, store)
				first := plainIssue("Valid first Issue")
				collision := plainIssue("Must refuse cross-kind identity")
				collision.Issue.ID = reservedID
				records, err := store.CreateIssueBatch(ctx, publicops.CreateBatchRequest{
					Actor: "test-author", ForceIDPrefix: true,
					Items: []publicops.BatchCreateItem{{Issue: first.Issue}, {Issue: collision.Issue}},
				})
				if !errors.Is(err, ErrAlreadyExists) || records != nil {
					t.Fatalf("batch claimed reserved Memory ID: records=%+v err=%v", records, err)
				}
				if after := workflowState(t, ctx, store); !reflect.DeepEqual(before, after) {
					t.Fatal("refused batch left native Issue, retained History, or catalog effects")
				}
			})
		}
	}
}

func graphBatchRowCount(t *testing.T, ctx context.Context, s *Store, table string) int {
	t.Helper()
	var count int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}

func TestGraphIssueCreateBatchMapsLocalDependenciesAtomically(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			ctx, options := issueExperimentOptions(t, backend)
			s, err := OpenExisting(ctx, options)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = s.Close() })
			prereq, err := s.CreateIssue(ctx, "beads/prereq", plainIssue("Prerequisite"))
			if err != nil {
				t.Fatal(err)
			}
			consumer, err := s.CreateIssue(ctx, "beads/consumer", plainIssue("Consumer"))
			if err != nil {
				t.Fatal(err)
			}
			request := publicops.CreateBatchRequest{Actor: "test-author", Items: []publicops.BatchCreateItem{
				{Issue: plainIssue("First heading").Issue},
				{Issue: plainIssue("Second heading").Issue, Dependencies: []publicops.CreateDependency{
					{TargetID: prereq.ID, Type: types.DepBlocks},
					{TargetID: consumer.Properties.ID, Type: types.DepBlocks, Reverse: true},
				}},
			}}
			before := workflowState(t, ctx, s)
			fault := errors.New("dependency Link retention failed")
			s.afterWrite = func(stage string) error {
				if stage == "create-dependency-retained" {
					return fault
				}
				return nil
			}
			_, err = s.CreateIssueBatch(ctx, request)
			s.afterWrite = nil
			if !errors.Is(err, fault) || !reflect.DeepEqual(before, workflowState(t, ctx, s)) {
				t.Fatalf("failed batch left native or graph state: %v", err)
			}
			records, err := s.CreateIssueBatch(ctx, request)
			if err != nil || len(records) != 2 {
				t.Fatalf("batch records=%+v err=%v", records, err)
			}
			updatedConsumer, err := s.ShowIssue(ctx, "beads/consumer")
			if err != nil {
				t.Fatal(err)
			}
			if len(records[0].Owned) != 0 || len(records[1].Owned) != 1 || len(updatedConsumer.Owned) != 1 {
				t.Fatalf("batch ownership first=%d second=%d consumer=%d", len(records[0].Owned), len(records[1].Owned), len(updatedConsumer.Owned))
			}
			for _, check := range []struct {
				owned  json.RawMessage
				source string
				target string
			}{
				{records[1].Owned[0], records[1].ID, prereq.ID},
				{updatedConsumer.Owned[0], consumer.ID, records[1].ID},
			} {
				var link LinkRecord
				if err := json.Unmarshal(check.owned, &link); err != nil {
					t.Fatal(err)
				}
				if link.Source != check.source || link.Target != check.target || link.Properties["dep_type"] != string(types.DepBlocks) {
					t.Fatalf("incorrect canonical dependency Link: %+v", link)
				}
			}
			before = workflowState(t, ctx, s)
			request.Items[1].Dependencies[0].TargetID = "missing-target"
			_, err = s.CreateIssueBatch(ctx, request)
			if !errors.Is(err, ErrNotFound) || !reflect.DeepEqual(before, workflowState(t, ctx, s)) {
				t.Fatalf("missing dependency target left partial batch: %v", err)
			}
		})
	}
}

func TestGraphIssueCreateBatchCanLinkEarlierExplicitItem(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			ctx, options := issueExperimentOptions(t, backend)
			s, err := OpenExisting(ctx, options)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = s.Close() })
			first := plainIssue("Prerequisite").Issue
			first.ID = "exp-prereq"
			request := publicops.CreateBatchRequest{Actor: "test-author", Items: []publicops.BatchCreateItem{
				{Issue: first},
				{Issue: plainIssue("Dependent").Issue, Dependencies: []publicops.CreateDependency{{TargetID: first.ID, Type: types.DepBlocks}}},
			}}
			before := workflowState(t, ctx, s)
			fault := errors.New("retaining same-batch Link failed")
			s.afterWrite = func(stage string) error {
				if stage == "create-dependency-retained" {
					return fault
				}
				return nil
			}
			_, err = s.CreateIssueBatch(ctx, request)
			s.afterWrite = nil
			if !errors.Is(err, fault) || !reflect.DeepEqual(before, workflowState(t, ctx, s)) {
				t.Fatalf("same-batch Link fault left native or graph state: %v", err)
			}
			records, err := s.CreateIssueBatch(ctx, request)
			if err != nil || len(records) != 2 {
				t.Fatalf("same-batch records=%+v err=%v", records, err)
			}
			if records[0].Properties.ID != first.ID || len(records[1].Owned) != 1 {
				t.Fatalf("same-batch Issue identity/owned Links: %+v", records)
			}
			var link LinkRecord
			if err := json.Unmarshal(records[1].Owned[0], &link); err != nil {
				t.Fatal(err)
			}
			if link.Source != records[1].ID || link.Target != records[0].ID || link.Properties["dep_type"] != string(types.DepBlocks) {
				t.Fatalf("same-batch Link = %+v", link)
			}
			before = workflowState(t, ctx, s)
			request.Items[0].Issue.ID = "exp-next"
			request.Items[1].Dependencies = []publicops.CreateDependency{{TargetID: "exp-next", Type: types.DepBlocks, Reverse: true}}
			_, err = s.CreateIssueBatch(ctx, request)
			if !errors.Is(err, ErrCapabilityUnavailable) || !reflect.DeepEqual(before, workflowState(t, ctx, s)) {
				t.Fatalf("reverse same-batch edge was not an atomic explicit refusal: %v", err)
			}
		})
	}
}

// These selected14 fixtures complement the existing11 adapter roots. They
// retain the real native writers, one original90 context per family, complete
// applied facets and source-owned state; no CLI/backend admission is changed.
func TestNativeCompositeCreateFamiliesKeepExactTypes(t *testing.T) {
	for _, family := range []string{"dependencies", "rooted", "child", "batch", "graph-plan"} {
		t.Run(family, func(t *testing.T) {
			ctx, _, s := intrinsicHardeningStore(t)
			checkIssue := func(record IssueRecord) {
				t.Helper()
				applied, present := record.ResourceDocument()
				if !present || applied.ID() != record.ID || applied.Revision() != record.Revision || applied.Type().URI != record.Type || applied.Type().Revision == "" {
					t.Fatalf("native create lost same-Tx applied identity/edition: %+v", record)
				}
				shown, err := s.GetResourceDocument(ctx, record.ID, record.Revision)
				if err != nil || !reflect.DeepEqual(shown.NativeJSON(), applied.NativeJSON()) {
					t.Fatalf("committed native exact body/pin differs: %v", err)
				}
			}
			checkLink := func(record LinkRecord) {
				t.Helper()
				applied, present := record.ResourceDocument()
				if !present || applied.ID() != record.ID || applied.Revision() != record.Revision || applied.Type().URI != record.Type || applied.Type().Revision == "" {
					t.Fatalf("native Link lost same-Tx applied edition: %+v", record)
				}
				shown, err := s.GetResourceDocument(ctx, record.ID, record.Revision)
				if err != nil || !reflect.DeepEqual(shown.NativeJSON(), applied.NativeJSON()) {
					t.Fatalf("committed Link endpoint/capture/pin differs: %v", err)
				}
			}
			switch family {
			case "dependencies":
				bug := plainIssue("Prerequisite bug")
				bug.Issue.IssueType = types.TypeBug
				prereq, err := s.CreateIssue(ctx, "beads/prereq", bug)
				if err != nil {
					t.Fatal(err)
				}
				consumer, err := s.CreateIssue(ctx, "beads/consumer", plainIssue("Consumer"))
				if err != nil {
					t.Fatal(err)
				}
				oldConsumer, present := consumer.ResourceDocument()
				if !present {
					t.Fatal("missing initial consumer exact pin")
				}
				deps := []IssueCreateDependency{
					{Target: IssueSelector{Path: "beads/prereq", Exact: true}, Type: types.DepBlocks},
					{Target: IssueSelector{Path: "beads/consumer", Exact: true}, Type: types.DepBlocks, Reverse: true},
				}
				before := workflowState(t, ctx, s)
				fault := errors.New("dependency create retained fault")
				s.afterWrite = func(stage string) error {
					if stage == "create-dependency-retained" {
						return fault
					}
					return nil
				}
				failed, err := s.CreateIssueWithDependencies(ctx, "beads/new", plainIssue("New"), deps)
				s.afterWrite = nil
				if !errors.Is(err, fault) || !reflect.ValueOf(failed).IsZero() || !reflect.DeepEqual(before, workflowState(t, ctx, s)) {
					t.Fatalf("dependency create escaped rollback: %+v %v", failed, err)
				}
				created, err := s.CreateIssueWithDependencies(ctx, "beads/new", plainIssue("New"), deps)
				if err != nil {
					t.Fatal(err)
				}
				checkIssue(created)
				checkIssue(prereq)
				changed, err := s.GetResourceDocument(ctx, consumer.ID, "")
				if err != nil || changed.Revision() == oldConsumer.Revision() || changed.Type() != oldConsumer.Type() {
					t.Fatalf("reverse source lost successor or preserved Type pin: %v", err)
				}
			case "rooted":
				one, err := s.CreateIssueWithIDRoot(ctx, "work", plainIssue("Rooted"))
				if err != nil {
					t.Fatal(err)
				}
				two, err := s.CreateIssueWithIDRootAndDependencies(ctx, "work", plainIssue("Rooted dependent"), []IssueCreateDependency{{Target: IssueSelector{Path: "beads/" + one.Properties.ID, NativeID: one.Properties.ID, Exact: true}, Type: types.DepBlocks}})
				if err != nil {
					t.Fatal(err)
				}
				for _, record := range []IssueRecord{one, two} {
					id := record.Properties.ID
					if len(id) <= len("work/") || id[:len("work/")] != "work/" || record.ID != s.ScopeURL()+id {
						t.Fatalf("rooted native identity changed: %+v", record)
					}
					checkIssue(record)
				}
			case "child":
				parent, err := s.CreateIssue(ctx, "beads/parent", plainIssue("Parent"))
				if err != nil {
					t.Fatal(err)
				}
				if _, err = s.CreateIssue(ctx, "beads/prereq", plainIssue("Prerequisite")); err != nil {
					t.Fatal(err)
				}
				request := IssueChildCreateRequest{ParentPath: "beads/parent", ExpectedParentRevision: parent.Revision, InheritLabels: true, Create: plainIssue("Child"), Dependencies: []IssueCreateDependency{{Target: IssueSelector{Path: "beads/prereq", Exact: true}, Type: types.DepBlocks}}}
				before := workflowState(t, ctx, s)
				fault := errors.New("child Link retention fault")
				s.afterWrite = func(stage string) error {
					if stage == "child-link-retained" {
						return fault
					}
					return nil
				}
				failed, err := s.CreateChildIssue(ctx, request)
				s.afterWrite = nil
				if !errors.Is(err, fault) || !reflect.ValueOf(failed).IsZero() || !reflect.DeepEqual(before, workflowState(t, ctx, s)) {
					t.Fatalf("child create escaped rollback: %+v %v", failed, err)
				}
				created, err := s.CreateChildIssue(ctx, request)
				if err != nil {
					t.Fatal(err)
				}
				checkIssue(created.Child)
				checkLink(created.Link)
				if created.Child.Properties.ID == parent.Properties.ID || created.Link.Source != created.Child.ID || created.Link.Target != parent.ID {
					t.Fatal("child allocation or native parent direction changed")
				}
			case "batch":
				first := plainIssue("First").Issue
				first.ID = "bd-batch-first"
				second := plainIssue("Second bug").Issue
				second.ID = "bd-batch-second"
				second.IssueType = types.TypeBug
				request := publicops.CreateBatchRequest{Actor: "test-author", Items: []publicops.BatchCreateItem{
					{Issue: first},
					{Issue: second, Dependencies: []publicops.CreateDependency{{TargetID: first.ID, Type: types.DepBlocks}}},
				}}
				before := workflowState(t, ctx, s)
				fault := errors.New("batch Link retention fault")
				s.afterWrite = func(stage string) error {
					if stage == "create-dependency-retained" {
						return fault
					}
					return nil
				}
				failed, err := s.CreateIssueBatch(ctx, request)
				s.afterWrite = nil
				if !errors.Is(err, fault) || failed != nil || !reflect.DeepEqual(before, workflowState(t, ctx, s)) {
					t.Fatalf("batch create escaped rollback: %+v %v", failed, err)
				}
				created, err := s.CreateIssueBatch(ctx, request)
				if err != nil || len(created) != 2 {
					t.Fatalf("batch result: %+v %v", created, err)
				}
				for _, record := range created {
					checkIssue(record)
				}
			case "graph-plan":
				existing, err := s.CreateIssue(ctx, "beads/existing-source", plainIssue("Existing source"))
				if err != nil {
					t.Fatal(err)
				}
				oldSource, present := existing.ResourceDocument()
				if !present {
					t.Fatal("missing existing source pin")
				}
				request := publicops.ApplyBatchRequest{Actor: "test-author", Items: []publicops.ApplyItem{
					{Kind: publicops.ItemCreate, Create: &publicops.CreateItem{Key: "first", Issue: plainIssue("First").Issue}},
					{Kind: publicops.ItemCreate, Create: &publicops.CreateItem{Key: "second", Issue: plainIssue("Second").Issue}},
					{Kind: publicops.ItemDepAdd, DepAdd: &publicops.DepAddItem{Source: publicops.Ref{Key: "first"}, Target: publicops.Ref{Key: "second"}, Type: types.DepBlocks}},
					{Kind: publicops.ItemDepAdd, DepAdd: &publicops.DepAddItem{Source: publicops.Ref{ID: existing.Properties.ID}, Target: publicops.Ref{Key: "second"}, Type: types.DepBlocks}},
					{Kind: publicops.ItemUpdate, Update: &publicops.UpdateItem{Target: publicops.Ref{Key: "first"}, Patch: publicops.IssuePatch{Assignee: publicops.Field[string]{Set: true, Value: "agent"}}}},
				}}
				before := workflowState(t, ctx, s)
				fault := errors.New("graph-plan Link retention fault")
				s.afterWrite = func(stage string) error {
					if stage == "create-dependency-retained" {
						return fault
					}
					return nil
				}
				failed, err := s.ApplyNewIssueGraph(ctx, request)
				s.afterWrite = nil
				if !errors.Is(err, fault) || failed != nil || !reflect.DeepEqual(before, workflowState(t, ctx, s)) {
					t.Fatalf("graph plan escaped rollback: %+v %v", failed, err)
				}
				created, err := s.ApplyNewIssueGraph(ctx, request)
				if err != nil || len(created) != 2 {
					t.Fatalf("graph plan result: %+v %v", created, err)
				}
				for _, record := range created {
					checkIssue(record)
				}
				if created["first"].Properties.Assignee != "agent" {
					t.Fatal("supported deferred assignment was lost")
				}
				changed, err := s.GetResourceDocument(ctx, existing.ID, "")
				if err != nil || changed.Revision() == oldSource.Revision() || changed.Type() != oldSource.Type() {
					t.Fatalf("existing graph source lost successor or pin: %v", err)
				}
			}
		})
	}
}

func TestNativeMemoryCompleteRootedConstructorKeepsExactType(t *testing.T) {
	ctx, _, s := intrinsicHardeningStore(t)
	created, err := s.Create(ctx, CreateRequest{IDRoot: "notes/team", TypeURL: MemoryTypeURL(s.ScopeURL()), PropertiesJSON: []byte(`{"title":"Complete native Memory","body":"Body"}`), Actor: "test-author"})
	if err != nil {
		t.Fatal(err)
	}
	applied, present := created.ResourceDocument()
	if !present || applied.Type().URI != MemoryTypeURL(s.ScopeURL()) || applied.Type().Revision == "" || applied.ID() != created.ID || applied.Revision() != created.Revision {
		t.Fatal("complete native constructor lost exact applied Type")
	}
	if len(created.PublicID) <= len("notes/team/") || created.PublicID[:len("notes/team/")] != "notes/team/" {
		t.Fatal("Memory root was changed")
	}
	shown, err := s.GetResourceDocument(ctx, created.ID, created.Revision)
	if err != nil || !reflect.DeepEqual(shown.NativeJSON(), applied.NativeJSON()) {
		t.Fatalf("complete Memory committed different body/pin: %v", err)
	}
	before := workflowState(t, ctx, s)
	failed, err := s.Create(ctx, CreateRequest{PublicID: "invalid-complete", Title: "Ambiguous", PropertiesJSON: []byte(`{"title":"Other","body":"Body"}`), Actor: "test-author"})
	if err == nil || !reflect.ValueOf(failed).IsZero() || !reflect.DeepEqual(before, workflowState(t, ctx, s)) {
		t.Fatalf("mixed Memory constructor left state: %+v %v", failed, err)
	}
}
