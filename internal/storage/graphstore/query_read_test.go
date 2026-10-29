package graphstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	graph "github.com/steveyegge/beads/graphops"
)

// These tests use detached descriptors and database/sql fixtures only. They do
// not establish engine isolation, persisted mixed-graph hydration, or installed
// behavior; those belong to the separately reserved provider qualification.
func TestQuerySnapshotContextBeforeOpen(t *testing.T) {
	s := &Store{}
	got, err := s.ReadQuerySnapshot(nil)
	if err == nil || !reflect.DeepEqual(got, QuerySnapshot{}) {
		t.Fatalf("nil context opened store or returned state: %+v %v", got, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	got, err = s.ReadQuerySnapshot(ctx)
	if !errors.Is(err, context.Canceled) || !reflect.DeepEqual(got, QuerySnapshot{}) {
		t.Fatalf("canceled context opened store or returned state: %+v %v", got, err)
	}
}

func TestQuerySnapshotEmptyContextAndCleanup(t *testing.T) {
	rollbackFault := errors.New("Query rollback failed")
	afterAcquisitionFault := errors.New("Query inventory acquisition failed")
	for _, tc := range []struct {
		name          string
		rollbackFault error
		readFault     error
	}{
		{name: "complete-empty-context"},
		{name: "rollback-failure", rollbackFault: rollbackFault},
		{name: "read-and-rollback-failure", rollbackFault: rollbackFault, readFault: afterAcquisitionFault},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, driver := exportCleanupStore(t, exportCleanupFaults{rollback: tc.rollbackFault})
			acquired := false
			s.afterWrite = func(stage string) error {
				if stage == "query-inventory" {
					acquired = true
					return tc.readFault
				}
				return nil
			}
			got, err := s.ReadQuerySnapshot(context.Background())
			if !acquired {
				t.Fatalf("fixture never reached complete Query acquisition: %v", err)
			}
			if tc.rollbackFault != nil {
				if !errors.Is(err, tc.rollbackFault) || (tc.readFault != nil && !errors.Is(err, tc.readFault)) || !reflect.DeepEqual(got, QuerySnapshot{}) {
					t.Fatalf("failed Query leaked state or lost cleanup: %+v %v", got, err)
				}
			} else {
				if err != nil || len(got.Snapshot.Records) != 0 || got.Snapshot.WriterToken != exportCleanupToken ||
					len(got.TypeFamilies) != len(driver.descriptors) || len(got.TypeKinds) != len(driver.descriptors) {
					t.Fatalf("empty context lost complete Type facts: %+v %v", got, err)
				}
				constraint := s.ScopeURL() + "constraint"
				policy := s.ScopeURL() + "policy"
				if got.TypeKinds[constraint] != graph.KindBead || !containsQueryAncestor(got.TypeFamilies[policy], constraint) {
					t.Fatalf("uninstantiated Constraint family lost ancestry: %+v", got.TypeFamilies[policy])
				}
			}
			exportCleanupTransactionCount(t, s, driver)
		})
	}
}

func TestQueryAllocationBudgetBeforePayloads(t *testing.T) {
	for _, count := range []uint64{0, PreviewSnapshotLimit + 1, QueryMaxResources, QueryMaxResources + 1} {
		t.Run(fmt.Sprintf("count-%d", count), func(t *testing.T) {
			s, mock := transactionPolicyMock(t)
			mock.ExpectBegin()
			mock.ExpectQuery("SELECT COUNT(*) FROM graph_preview_catalog").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(count))
			mock.ExpectRollback()
			err := s.withTx(context.Background(), false, func(tx *sql.Tx) error {
				return checkQueryAllocationCount(context.Background(), tx, QueryMaxResources)
			})
			if errors.Is(err, ErrLimitExceeded) != (count > QueryMaxResources) || (count <= QueryMaxResources && err != nil) {
				t.Fatalf("allocation budget count%d: %v", count, err)
			}
		})
	}
}

func TestQueryAllocationRowsFailClosed(t *testing.T) {
	rowsFault := errors.New("Query allocation rows close failed")
	for _, tc := range []struct {
		name    string
		rows    *sqlmock.Rows
		want    error
		wantLen int
	}{
		{name: "valid-mixed", rows: sqlmock.NewRows([]string{"path", "backing", "kind"}).AddRow("links/z", "dependency", "link").AddRow("beads/a", "generic", "bead"), wantLen: 2},
		{name: "kind-mismatch", rows: sqlmock.NewRows([]string{"path", "backing", "kind"}).AddRow("links/z", "generic", "bead"), want: ErrInvalidStore},
		{name: "backing-mismatch", rows: sqlmock.NewRows([]string{"path", "backing", "kind"}).AddRow("beads/a", "dependency", "bead"), want: ErrInvalidStore},
		{name: "unknown-backing", rows: sqlmock.NewRows([]string{"path", "backing", "kind"}).AddRow("beads/a", "unsupported", "bead"), want: ErrInvalidStore},
		{name: "malformed-path", rows: sqlmock.NewRows([]string{"path", "backing", "kind"}).AddRow("beads/../a", "generic", "bead"), want: ErrInvalidStore},
		{name: "iteration-failure", rows: sqlmock.NewRows([]string{"path", "backing", "kind"}).AddRow("beads/a", "generic", "bead").RowError(0, rowsFault), want: rowsFault},
		{name: "validation-and-close-failure", rows: sqlmock.NewRows([]string{"path", "backing", "kind"}).AddRow("beads/a", "unsupported", "bead").CloseError(rowsFault), want: rowsFault},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, mock := transactionPolicyMock(t)
			mock.ExpectBegin()
			mock.ExpectQuery("SELECT path,backing,resource_kind FROM graph_preview_catalog WHERE allocation_state='live' LIMIT ?").WithArgs(QueryMaxResources + 1).WillReturnRows(tc.rows).RowsWillBeClosed()
			mock.ExpectRollback()
			var result []queryAllocation
			err := s.withTx(context.Background(), false, func(tx *sql.Tx) error {
				var err error
				result, err = queryAllocationsInTx(context.Background(), tx)
				return err
			})
			if tc.want != nil {
				if !errors.Is(err, tc.want) || result != nil {
					t.Fatalf("failed allocation feed leaked state: %+v %v", result, err)
				}
			} else if err != nil || len(result) != tc.wantLen || result[0].path != "beads/a" {
				t.Fatalf("complete allocation order: %+v %v", result, err)
			}
		})
	}
}

func TestQueryTypeFamiliesCompleteAndValidated(t *testing.T) {
	const scope = "https://example.test/query/"
	descriptors := []graph.TypeDescriptor{}
	for _, definition := range previewTypeDefinitions() {
		descriptor, err := definition.build(scope)
		if err != nil {
			t.Fatal(err)
		}
		descriptors = append(descriptors, descriptor)
	}
	families, kinds, err := queryTypeFamilies(context.Background(), descriptors)
	if err != nil || len(families) != len(descriptors) || len(kinds) != len(descriptors) {
		t.Fatalf("complete nominal facts: %v", err)
	}
	for _, pair := range [][2]string{{"policy", "constraint"}, {"policy", "memory"}, {"defines", "requires"}, {"task", "issue"}, {"blocks", "dependency"}} {
		if !containsQueryAncestor(families[scope+pair[0]], scope+pair[1]) || !containsQueryAncestor(families[scope+pair[0]], scope+pair[0]) {
			t.Fatalf("lost exact reflexive family fact %v: %+v", pair, families[scope+pair[0]])
		}
	}
	newDescriptor := func(name string, parents ...string) graph.TypeDescriptor {
		t.Helper()
		descriptor, err := graph.NewTypeDescriptor(graph.TypeDescriptorSpec{ID: scope + name, Name: name, Describes: graph.KindBead, ConformsTo: parents})
		if err != nil {
			t.Fatal(err)
		}
		return descriptor
	}
	for name, broken := range map[string][]graph.TypeDescriptor{
		"duplicate":         append(append([]graph.TypeDescriptor(nil), descriptors...), descriptors[0]),
		"wrong-kind-parent": append(append([]graph.TypeDescriptor(nil), descriptors...), newDescriptor("bad-child", scope+"requires")),
		"missing-parent":    {newDescriptor("child", scope+"absent")},
		"cycle":             {newDescriptor("a", scope+"b"), newDescriptor("b", scope+"a")},
		"zero-descriptor":   {{}},
	} {
		t.Run(name, func(t *testing.T) {
			families, kinds, err := queryTypeFamilies(context.Background(), broken)
			if !errors.Is(err, ErrInvalidStore) || families != nil || kinds != nil {
				t.Fatalf("invalid ancestry returned facts: %+v %+v %v", families, kinds, err)
			}
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if families, kinds, err := queryTypeFamilies(ctx, descriptors); !errors.Is(err, context.Canceled) || families != nil || kinds != nil {
		t.Fatalf("canceled ancestry returned facts: %+v %+v %v", families, kinds, err)
	}
}

func containsQueryAncestor(family []string, expected string) bool {
	for _, ancestor := range family {
		if ancestor == expected {
			return true
		}
	}
	return false
}
