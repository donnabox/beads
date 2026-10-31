package graphstore

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/typecatalog"
	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/resourceget"
)

// Private frames here are explicit SQLmock/control inputs, not a claim that a
// caller can authenticate a native profile or Type catalog with a constructor.
func issueDiscoveryControlContext(s *Store, tx *sql.Tx, budget *intrinsicMetadataBudget) context.Context {
	capture := checkedResourceCapture{installed: checkedInstalledTypes{catalog: &typecatalog.Catalog{}}}
	ctx := s.contextForResourceCapture(context.Background(), tx, capture, budget)
	return context.WithValue(ctx, issueDiscoveryReadKey{}, &issueDiscoveryRead{store: s, tx: tx, budget: budget, capture: capture})
}

func TestIssueDiscoveryCheckedDefinitionCorrespondence(t *testing.T) {
	const scope = "https://scope.example/"
	const revision = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	makeDoc := func(kind graph.ResourceKind, id, typ string) resourceget.ResourceDocument {
		raw, err := canonicalJSON(map[string]any{"id": scope + id, "type": map[string]string{"uri": scope + typ, "revision": "checked-type-edition"}, "revision": revision, "properties": map[string]any{}})
		if err != nil {
			t.Fatal(err)
		}
		if kind == graph.KindLink {
			raw, err = canonicalJSON(map[string]any{"id": scope + id, "type": map[string]string{"uri": scope + typ, "revision": "checked-type-edition"}, "revision": revision, "source": scope + "base", "target": scope + "derived", "properties": map[string]any{}})
			if err != nil {
				t.Fatal(err)
			}
		}
		document, err := resourceget.NewResourceDocument(scope, kind, raw)
		if err != nil {
			t.Fatal(err)
		}
		return document
	}
	bead := makeDoc(graph.KindBead, "opaque-type", "bead-type")
	link := makeDoc(graph.KindLink, "opaque-refinement", "refines")
	for _, mode := range []string{"intrinsic", "installed-type", "type-definition", "intrinsic-link", "installed-type-link", "wrong-kind", "wrong-pin-uri", "wrong-revision", "wrong-public-id", "duplicate-definition", "missing-definition", "unknown-backing"} {
		t.Run(mode, func(t *testing.T) {
			docs := []resourceget.ResourceDocument{bead}
			path, kind, typ, rev, backing, id := "beads/opaque-private", "bead", scope+"bead-type", revision, mode, "opaque-type"
			wantErr := false
			switch mode {
			case "intrinsic-link", "installed-type-link":
				docs = []resourceget.ResourceDocument{link}
				path = "links/opaque-private"
				kind = "link"
				typ = scope + "refines"
				id = "opaque-refinement"
			case "wrong-kind":
				backing = "installed-type"
				kind = "link"
				wantErr = true
			case "wrong-pin-uri":
				backing = "installed-type"
				typ = scope + "other-type"
				wantErr = true
			case "wrong-revision":
				backing = "installed-type"
				rev = strings.Repeat("b", 32)
				wantErr = true
			case "wrong-public-id":
				backing = "installed-type"
				id = "other"
				wantErr = true
			case "duplicate-definition":
				backing = "installed-type"
				docs = append(docs, bead)
				wantErr = true
			case "missing-definition":
				backing = "installed-type"
				docs = nil
				wantErr = true
			case "unknown-backing":
				wantErr = true
			}
			err := issueCatalogDefinitionMatches(docs, scope, path, kind, typ, rev, backing, id)
			if wantErr {
				if !errors.Is(err, ErrInvalidStore) {
					t.Fatalf("unchecked definition correspondence admitted: %v", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestIssueDiscoveryPolicySharesOperationBudget(t *testing.T) {
	queries := []string{
		"SELECT COUNT(*), CAST(LEAST(COALESCE(SUM(OCTET_LENGTH(name)+OCTET_LENGTH(category)),0),?) AS UNSIGNED) FROM custom_statuses",
		"SELECT COUNT(*), CAST(LEAST(COALESCE(SUM(OCTET_LENGTH(`key`)+OCTET_LENGTH(value)),0),?) AS UNSIGNED) FROM config WHERE `key`='status.custom'",
	}
	fault := errors.New("policy driver fault")
	for _, mode := range []string{"exact-boundary", "combined-row-over", "combined-byte-over", "shared-budget-exhausted", "query-error", "scan-error", "canceled"} {
		t.Run(mode, func(t *testing.T) {
			tx, mock := intrinsicProfileTransaction(t)
			budget := newIntrinsicMetadataBudget()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			var want error
			if mode == "canceled" {
				cancel()
				want = context.Canceled
			} else {
				rows, size := uint64(0), uint64(0)
				if mode == "exact-boundary" {
					rows = PreviewIssueListConfigRowLimit - 1
					size = PreviewIssueListConfigByteLimit - 1
				}
				if mode == "combined-row-over" {
					rows = PreviewIssueListConfigRowLimit
					want = ErrLimitExceeded
				}
				if mode == "combined-byte-over" {
					size = PreviewIssueListConfigByteLimit
					want = ErrLimitExceeded
				}
				if mode == "shared-budget-exhausted" {
					budget.remaining = 0
					size = 1
					want = ErrLimitExceeded
				}
				q := mock.ExpectQuery(queries[0]).WithArgs(PreviewIssueListConfigByteLimit + 1)
				if mode == "query-error" {
					q.WillReturnError(fault)
					want = fault
				} else if mode == "scan-error" {
					q.WillReturnRows(sqlmock.NewRows([]string{"count", "bytes"}).AddRow("bad", 0))
				} else {
					q.WillReturnRows(sqlmock.NewRows([]string{"count", "bytes"}).AddRow(rows, size)).RowsWillBeClosed()
				}
				if mode != "query-error" && mode != "scan-error" && mode != "shared-budget-exhausted" {
					mock.ExpectQuery(queries[1]).WithArgs(PreviewIssueListConfigByteLimit + 1).WillReturnRows(sqlmock.NewRows([]string{"count", "bytes"}).AddRow(1, 1)).RowsWillBeClosed()
				}
			}
			err := chargeIssueDiscoveryPolicyInTx(ctx, tx, budget)
			if mode == "scan-error" {
				if err == nil {
					t.Fatal("invalid policy scalar admitted")
				}
				return
			}
			if want == nil {
				if err != nil || budget.remaining >= PreviewCurrentReadByteLimit {
					t.Fatalf("policy not charged: %v, %+v", err, budget)
				}
			} else if !errors.Is(err, want) {
				t.Fatalf("policy refusal/cause: %v", err)
			}
		})
	}
}

func TestIssueDiscoveryRepeatedHydrationIsCharged(t *testing.T) {
	fault := errors.New("selected mapping driver fault")
	for _, mode := range []string{"repeated-charge", "missing-retained", "oversize", "second-exhausted", "driver-error", "wrong-store", "wrong-tx", "outside-discovery-no-frame"} {
		t.Run(mode, func(t *testing.T) {
			tx, mock := intrinsicProfileTransaction(t)
			s := &Store{options: Options{Binding: Binding{SchemaVersion: TypesFreshSchemaVersion}}}
			budget := newIntrinsicMetadataBudget()
			ctx := issueDiscoveryControlContext(s, tx, budget)
			var want error
			if mode == "outside-discovery-no-frame" {
				ctx = context.Background()
			} else if mode == "wrong-store" {
				s = &Store{}
				want = ErrInvalidStore
			} else if mode == "wrong-tx" {
				frame := ctx.Value(issueDiscoveryReadKey{}).(*issueDiscoveryRead)
				frame.tx = nil
				want = ErrInvalidStore
			} else {
				count, size := uint64(1), uint64(8192)
				if mode == "missing-retained" {
					count = 0
					want = ErrInvalidStore
				}
				if mode == "oversize" {
					size = PreviewCurrentReadByteLimit + 1
					want = ErrLimitExceeded
				}
				q := mock.ExpectQuery(issueDiscoveryHydrationSizeQuery).WithArgs(PreviewCurrentReadByteLimit+1, "beads/opaque")
				if mode == "driver-error" {
					q.WillReturnError(fault)
					want = fault
				} else {
					q.WillReturnRows(sqlmock.NewRows([]string{"count", "bytes"}).AddRow(count, size)).RowsWillBeClosed()
				}
				if mode == "second-exhausted" {
					budget.remaining = 8192
				}
			}
			err := s.chargeIssueDiscoveryHydrationInTx(ctx, tx, "beads/opaque")
			if want != nil {
				if !errors.Is(err, want) {
					t.Fatalf("lost selected read refusal: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if mode == "repeated-charge" || mode == "second-exhausted" {
				mock.ExpectQuery(issueDiscoveryHydrationSizeQuery).WithArgs(PreviewCurrentReadByteLimit+1, "beads/opaque").WillReturnRows(sqlmock.NewRows([]string{"count", "bytes"}).AddRow(1, 8192)).RowsWillBeClosed()
				err = s.chargeIssueDiscoveryHydrationInTx(ctx, tx, "beads/opaque")
				if mode == "second-exhausted" {
					if !errors.Is(err, ErrLimitExceeded) {
						t.Fatalf("repeat read reset allowance: %v", err)
					}
				} else if err != nil || budget.remaining != PreviewCurrentReadByteLimit-16384 {
					t.Fatalf("repeat read uncharged: %v %+v", err, budget)
				}
			}
		})
	}
}

func TestIssueDiscoverySelectorsPreserveOrderAndFailures(t *testing.T) {
	const predicate = `(backing='issue' AND backing_key=?) OR (backing IN ('generic','dependency','informational','intrinsic','intrinsic-link','type-definition','installed-type','installed-type-link') AND COALESCE(public_id,path)=?)`
	sizeQuery := `SELECT CAST(LEAST(COALESCE(SUM(
 4*(256+OCTET_LENGTH(path)+COALESCE(OCTET_LENGTH(backing_key),0)+COALESCE(OCTET_LENGTH(public_id),0))
 ),0),?) AS UNSIGNED) FROM graph_preview_catalog WHERE ` + predicate
	for _, mode := range []string{"ordered-duplicates-and-nonissue", "second-driver-fault", "ambiguous-public", "tombstone", "budget", "canceled"} {
		t.Run(mode, func(t *testing.T) {
			tx, mock := intrinsicProfileTransaction(t)
			s := &Store{options: Options{Binding: Binding{SchemaVersion: TypesFreshSchemaVersion, ScopeURL: "https://scope.example/"}}}
			budget := newIntrinsicMetadataBudget()
			ctx := issueDiscoveryControlContext(s, tx, budget)
			selectors := []IssueSelector{{Path: "beads/missing", NativeID: "missing", Exact: true}, {Path: "beads/type", NativeID: "type", Exact: true}, {Path: "beads/missing", NativeID: "missing", Exact: true}}
			fault := errors.New("second selector driver fault")
			var want error
			if mode == "budget" {
				budget.remaining = 0
				want = ErrLimitExceeded
			} else if mode == "canceled" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
				want = context.Canceled
			} else {
				for i, selector := range selectors {
					q := mock.ExpectQuery(sizeQuery).WithArgs(PreviewCurrentReadByteLimit+1, selector.NativeID, selector.NativeID)
					if mode == "second-driver-fault" && i == 1 {
						q.WillReturnError(fault)
						want = fault
						break
					}
					q.WillReturnRows(sqlmock.NewRows([]string{"bytes"}).AddRow(0)).RowsWillBeClosed()
					rows := sqlmock.NewRows([]string{"path", "kind", "state", "backing", "public_id"})
					if selector.NativeID == "type" {
						rows.AddRow("beads/installed-opaque", "bead", "live", "installed-type", "type")
					}
					if mode == "ambiguous-public" {
						rows.AddRow("beads/a", "bead", "live", "generic", selector.NativeID).AddRow("beads/b", "bead", "live", "generic", selector.NativeID)
						want = ErrInvalidStore
					}
					if mode == "tombstone" {
						rows.AddRow("beads/a", "bead", "deleted", "generic", selector.NativeID)
						want = ErrGone
					}
					mock.ExpectQuery(`SELECT `+allocatedIdentityColumns+` FROM graph_preview_catalog WHERE `+predicate+` LIMIT 2`).WithArgs(selector.NativeID, selector.NativeID).WillReturnRows(rows).RowsWillBeClosed()
					if mode == "ambiguous-public" || mode == "tombstone" {
						break
					}
					if selector.NativeID == "type" {
						mock.ExpectQuery(`SELECT backing FROM graph_preview_catalog WHERE path=?`).WithArgs("beads/installed-opaque").WillReturnRows(sqlmock.NewRows([]string{"backing"}).AddRow("installed-type"))
					}
				}
			}
			paths, err := s.resolveIssuePathsInTx(ctx, tx, selectors)
			if want != nil {
				if !errors.Is(err, want) || paths != nil {
					t.Fatalf("partial selector output/refusal: %v %v", paths, err)
				}
			} else if err != nil || !reflect.DeepEqual(paths, []string{"missing", "type", "missing"}) {
				t.Fatalf("reordered/deduped selector output: %v %v", paths, err)
			}
		})
	}
}

func TestIssueDiscoveryPublicRefusalSuppressesOutputs(t *testing.T) {
	for _, mode := range []string{"durable-required", "ephemeral-refused", "invalid-selector", "too-many-selectors", "canceled", "nil-context", "binding-refusal", "rollback-failure"} {
		t.Run(mode, func(t *testing.T) {
			s := &Store{options: Options{Binding: Binding{SchemaVersion: TypesFreshSchemaVersion}}}
			filter := types.IssueFilter{SkipWisps: true}
			selectors := []IssueSelector(nil)
			ctx := context.Background()
			var want error
			switch mode {
			case "durable-required":
				filter.SkipWisps = false
				want = ErrCapabilityUnavailable
			case "ephemeral-refused":
				value := true
				filter.Ephemeral = &value
				want = ErrCapabilityUnavailable
			case "invalid-selector":
				selectors = []IssueSelector{{Path: "../outside"}}
			case "too-many-selectors":
				for i := 0; i < PreviewSnapshotLimit+1; i++ {
					selectors = append(selectors, IssueSelector{Path: "beads/missing", NativeID: "missing", Exact: true})
				}
				want = ErrLimitExceeded
			case "canceled":
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
				want = context.Canceled
			case "nil-context":
				ctx = nil
			}
			var rollbackFault error
			if mode == "binding-refusal" || mode == "rollback-failure" {
				var mock sqlmock.Sqlmock
				s, mock = transactionPolicyMock(t)
				s.options.Binding.SchemaVersion = TypesFreshSchemaVersion
				mock.ExpectBegin()
				mock.ExpectQuery(`SELECT workspace, scope_url, authority_id,
        schema_version, writer_token FROM graph_preview_scope WHERE singleton = 1
 AND OCTET_LENGTH(workspace)=? AND OCTET_LENGTH(scope_url)=?`).WithArgs(0, 0).WillReturnError(errors.New("binding unavailable"))
				want = ErrInvalidStore
				if mode == "rollback-failure" {
					rollbackFault = errors.New("read rollback unavailable")
					mock.ExpectRollback().WillReturnError(rollbackFault)
				} else {
					mock.ExpectRollback()
				}
			}
			rows, paths, err := s.SearchIssueRowsWithSelectors(ctx, "", filter, selectors)
			if err == nil || want != nil && !errors.Is(err, want) || rows != nil || paths != nil {
				t.Fatalf("refusal leaked discovery result: %v %v %v", rows, paths, err)
			}
			if rollbackFault != nil && !errors.Is(err, rollbackFault) {
				t.Fatalf("lost rollback cause: %v", err)
			}
		})
	}
}

// Legacy callers retain the original exact native-ID SQL and literal fallback;
// no selected14 Type frame or whole-Resource lookup is introduced on11.
func TestIssueDiscoveryLegacyMissingSelectorCompatibility(t *testing.T) {
	tx, mock := intrinsicProfileTransaction(t)
	s := &Store{options: Options{Binding: Binding{SchemaVersion: SchemaVersion, ScopeURL: "https://scope.example/"}}}
	selectors := []IssueSelector{{Path: "beads/missing", NativeID: "missing", Exact: true}, {Path: "beads/missing", NativeID: "missing", Exact: true}}
	for range selectors {
		mock.ExpectQuery(`SELECT path FROM graph_preview_catalog WHERE backing='issue' AND backing_key=? AND allocation_state='live' LIMIT 2`).WithArgs("missing").WillReturnRows(sqlmock.NewRows([]string{"path"})).RowsWillBeClosed()
		mock.ExpectQuery(`SELECT backing FROM graph_preview_catalog
  WHERE COALESCE(public_id,path)=? AND backing IN ('generic','dependency','informational') AND allocation_state='live'`).WithArgs("missing").WillReturnError(sql.ErrNoRows)
		mock.ExpectQuery(`SELECT backing FROM graph_preview_catalog WHERE path=? AND allocation_state='live'`).WithArgs("beads/missing").WillReturnError(sql.ErrNoRows)
		mock.ExpectQuery(`SELECT backing FROM graph_preview_catalog
  WHERE backing='issue' AND backing_key=? AND allocation_state='deleted'`).WithArgs("missing").WillReturnError(sql.ErrNoRows)
	}
	paths, err := s.resolveIssuePathsInTx(context.Background(), tx, selectors)
	if err != nil || !reflect.DeepEqual(paths, []string{"missing", "missing"}) {
		t.Fatalf("legacy selector changed: %v, %v", paths, err)
	}
}

// These are strict pre-hydration public-API refusal and cleanup controls.
// They do not claim after-hydration connection-close or real engine proof.
func TestNativeIssueReadPublicRefusalSuppressesRecord(t *testing.T) {
	for _, entry := range []string{"direct", "selected"} {
		for _, mode := range []string{"nil-context", "canceled", "binding-refusal", "rollback-failure"} {
			t.Run(entry+"-"+mode, func(t *testing.T) {
				s := &Store{options: Options{Binding: Binding{SchemaVersion: TypesFreshSchemaVersion}}}
				ctx := context.Background()
				var want error
				var rollbackFault error
				switch mode {
				case "nil-context":
					ctx = nil
				case "canceled":
					var cancel context.CancelFunc
					ctx, cancel = context.WithCancel(ctx)
					cancel()
					want = context.Canceled
				case "binding-refusal", "rollback-failure":
					var mock sqlmock.Sqlmock
					s, mock = transactionPolicyMock(t)
					s.options.Binding.SchemaVersion = TypesFreshSchemaVersion
					mock.ExpectBegin()
					mock.ExpectQuery(`SELECT workspace, scope_url, authority_id,
        schema_version, writer_token FROM graph_preview_scope WHERE singleton = 1
 AND OCTET_LENGTH(workspace)=? AND OCTET_LENGTH(scope_url)=?`).WithArgs(0, 0).WillReturnError(errors.New("binding unavailable"))
					want = ErrInvalidStore
					if mode == "rollback-failure" {
						rollbackFault = errors.New("read rollback unavailable")
						mock.ExpectRollback().WillReturnError(rollbackFault)
					} else {
						mock.ExpectRollback()
					}
				}
				var result IssueRecord
				var err error
				if entry == "direct" {
					result, err = s.ShowIssue(ctx, "beads/item")
				} else {
					result, err = s.ShowIssueSelected(ctx, IssueSelector{Path: "beads/item", Exact: true})
				}
				if err == nil || want != nil && !errors.Is(err, want) || !reflect.ValueOf(result).IsZero() {
					t.Fatalf("public read refusal leaked result: %+v %v", result, err)
				}
				if rollbackFault != nil && !errors.Is(err, rollbackFault) {
					t.Fatalf("lost rollback cause: %v", err)
				}
				if s.db != nil && s.db.Stats().InUse != 0 {
					t.Fatal("read refusal leaked connection")
				}
			})
		}
	}
}
