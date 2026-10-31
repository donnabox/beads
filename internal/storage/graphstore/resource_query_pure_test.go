package graphstore

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// These strict SQLmock controls qualify no engine, native schema, full capture,
// or evaluator. They exercise the production preflight and header paths only.
func expectResourceQuerySizes(mock sqlmock.Sqlmock, sizes map[int]uint64, stop int, fault error) {
	for i, part := range resourceQueryReadSizeQueries() {
		fields := []string{"256"}
		for _, column := range strings.Split(part.columns, ",") {
			fields = append(fields, "COALESCE(OCTET_LENGTH("+strings.TrimSpace(column)+"),0)")
		}
		query := "SELECT CAST(LEAST(COALESCE(SUM((" + strings.Join(fields, "+") + ")*?),0),?) AS UNSIGNED) FROM " + part.from
		expected := mock.ExpectQuery(query).WithArgs(part.copies, PreviewCurrentReadByteLimit+1)
		if i == stop && fault != nil {
			expected.WillReturnError(fault)
			return
		}
		expected.WillReturnRows(sqlmock.NewRows([]string{"bytes"}).AddRow(sizes[i])).RowsWillBeClosed()
		if i == stop {
			return
		}
	}
}

func TestResourceQueryPreflightBeforePayload(t *testing.T) {
	fault := errors.New("preflight driver fault")
	oldCount := len(currentReadSizeQueries())
	for _, mode := range []string{"count-over-including-deleted", "count-query-error", "count-scan-error", "exact-limit", "base-one-over", "context-cumulative-over", "member-cumulative-over", "context-query-error", "member-query-error", "wisp-refusal", "wisp-query-error", "canceled"} {
		t.Run(mode, func(t *testing.T) {
			tx, mock := intrinsicProfileTransaction(t)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			var want error
			if mode == "canceled" {
				cancel()
				want = context.Canceled
			} else {
				expected := mock.ExpectQuery(`SELECT COUNT(*) FROM graph_preview_catalog`)
				switch mode {
				case "count-over-including-deleted":
					expected.WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(resourceQueryAllocationLimit + 1))
					want = ErrLimitExceeded
				case "count-query-error":
					expected.WillReturnError(fault)
					want = fault
				case "count-scan-error":
					expected.WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow("invalid count"))
				default:
					// The boundary counts all10000 allocations, not merely returned/live rows.
					expected.WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(resourceQueryAllocationLimit))
					sizes := map[int]uint64{}
					stop := -1
					var sizeFault error
					switch mode {
					case "exact-limit":
						sizes[0] = PreviewCurrentReadByteLimit - 2
						sizes[oldCount] = 1
						sizes[oldCount+1] = 1
					case "base-one-over":
						sizes[0] = PreviewCurrentReadByteLimit + 1
						stop = 0
						want = ErrLimitExceeded
					case "context-cumulative-over":
						sizes[0] = PreviewCurrentReadByteLimit
						sizes[oldCount] = 1
						stop = oldCount
						want = ErrLimitExceeded
					case "member-cumulative-over":
						sizes[0] = PreviewCurrentReadByteLimit - 1
						sizes[oldCount] = 1
						sizes[oldCount+1] = 1
						stop = oldCount + 1
						want = ErrLimitExceeded
					case "context-query-error":
						stop = oldCount
						sizeFault = fault
						want = fault
					case "member-query-error":
						stop = oldCount + 1
						sizeFault = fault
						want = fault
					}
					expectResourceQuerySizes(mock, sizes, stop, sizeFault)
					if stop < 0 {
						wisp := mock.ExpectQuery(`SELECT COUNT(*) FROM wisps w JOIN graph_preview_catalog c ON c.backing='issue' AND c.backing_key=w.id WHERE c.allocation_state='live'`)
						switch mode {
						case "wisp-refusal":
							wisp.WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
							want = ErrInvalidStore
						case "wisp-query-error":
							wisp.WillReturnError(fault)
							want = fault
						default:
							wisp.WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
						}
					}
				}
			}
			err := checkResourceQueryPreflightInTx(ctx, tx)
			if mode == "count-scan-error" {
				if err == nil {
					t.Fatal("invalid count admitted")
				}
				return
			}
			if want == nil {
				if err != nil {
					t.Fatal(err)
				}
			} else if !errors.Is(err, want) {
				t.Fatalf("lost preflight refusal/cause: %v", err)
			}
			// No Type/native payload expectation exists: any premature acquisition
			// fails this strict ordered mock rather than being silently accepted.
		})
	}
}

func TestResourceQueryCountAndBytesPrecedeTypeCapture(t *testing.T) {
	for _, mode := range []string{"count-over", "bytes-over", "context-driver-error", "logical-allowance-exhausted"} {
		t.Run(mode, func(t *testing.T) {
			tx, mock := intrinsicProfileTransaction(t)
			want := ErrLimitExceeded
			count := resourceQueryAllocationLimit
			if mode == "count-over" {
				count++
			}
			mock.ExpectQuery(`SELECT COUNT(*) FROM graph_preview_catalog`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(count))
			if mode == "bytes-over" {
				expectResourceQuerySizes(mock, map[int]uint64{0: PreviewCurrentReadByteLimit + 1}, 0, nil)
			}
			if mode == "context-driver-error" || mode == "logical-allowance-exhausted" {
				expectResourceQuerySizes(mock, nil, -1, nil)
				mock.ExpectQuery(`SELECT COUNT(*) FROM wisps w JOIN graph_preview_catalog c ON c.backing='issue' AND c.backing_key=w.id WHERE c.allocation_state='live'`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				if mode == "context-driver-error" {
					want = errors.New("first exact context query failed")
					mock.ExpectQuery(`SELECT m.context_id FROM graph_preview_intrinsic_members m JOIN graph_preview_catalog c ON c.path=m.path AND c.revision=m.revision WHERE m.role=? AND c.backing=? AND c.allocation_state=? LIMIT 2`).WithArgs("type", "intrinsic", "live").WillReturnError(want)
				}
			}
			budget := newIntrinsicMetadataBudget()
			if mode == "logical-allowance-exhausted" {
				budget.remaining = 0
			}
			snapshot, err := (&Store{}).readResourceQuerySnapshotInTx(context.Background(), tx, budget)
			if !errors.Is(err, want) || !reflect.DeepEqual(snapshot, ResourceSnapshot{}) {
				t.Fatalf("partial snapshot/order/cause: %+v, %v", snapshot, err)
			}
		})
	}
}

func TestResourceQueryLiveMappingCleanup(t *testing.T) {
	fault := errors.New("mapping driver failure")
	for _, mode := range []string{"ordinary-boundary", "ordinary-over", "query-boundary", "query-over", "invalid-limit", "invalid-path", "invalid-public-id", "invalid-state", "invalid-kind", "duplicate-public-id", "query-error", "scan-error", "row-error", "canceled-row", "close-error", "extra-result", "pre-canceled"} {
		t.Run(mode, func(t *testing.T) {
			tx, mock := intrinsicProfileTransaction(t)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			limit, count := PreviewSnapshotLimit, 1
			var want error
			switch mode {
			case "ordinary-boundary":
				count = PreviewSnapshotLimit
			case "ordinary-over":
				count = PreviewSnapshotLimit + 1
				want = ErrLimitExceeded
			case "query-boundary":
				limit = resourceQueryAllocationLimit
				count = resourceQueryAllocationLimit
			case "query-over":
				limit = resourceQueryAllocationLimit
				count = resourceQueryAllocationLimit + 1
				want = ErrLimitExceeded
			case "invalid-limit":
				limit = resourceQueryAllocationLimit + 1
				want = ErrInvalidStore
			case "pre-canceled":
				cancel()
				want = context.Canceled
			case "query-error", "row-error", "close-error":
				want = fault
			case "canceled-row":
				want = context.Canceled
			case "invalid-path", "invalid-public-id", "invalid-state", "invalid-kind", "duplicate-public-id", "extra-result":
				want = ErrInvalidStore
			}
			if mode != "invalid-limit" && mode != "pre-canceled" {
				expected := mock.ExpectQuery(`SELECT ` + allocatedIdentityColumns + ` FROM graph_preview_catalog WHERE allocation_state='live' ORDER BY path LIMIT ?`).WithArgs(limit + 1)
				rows := sqlmock.NewRows([]string{"path", "kind", "state", "backing", "public_id"})
				for i := 0; i < count; i++ {
					path, id, kind, state := fmt.Sprintf("beads/r%d", i), fmt.Sprintf("r%d", i), "bead", "live"
					switch mode {
					case "invalid-path":
						path = "../outside"
					case "invalid-public-id":
						id = ""
					case "invalid-state":
						state = "deleted"
					case "invalid-kind":
						kind = "unknown"
					}
					if mode == "scan-error" {
						rows.AddRow(nil, kind, state, "generic", id)
					} else {
						rows.AddRow(path, kind, state, "generic", id)
					}
				}
				switch mode {
				case "duplicate-public-id":
					rows.AddRow("beads/other", "bead", "live", "generic", "r0")
				case "row-error":
					rows.RowError(0, fault)
				case "canceled-row":
					rows.RowError(0, context.Canceled)
				case "close-error":
					rows.CloseError(fault)
				}
				if mode == "query-error" {
					expected.WillReturnError(fault)
				} else if mode == "extra-result" {
					expected.WillReturnRows(rows, sqlmock.NewRows([]string{"path", "kind", "state", "backing", "public_id"})).RowsWillBeClosed()
				} else {
					expected.WillReturnRows(rows).RowsWillBeClosed()
				}
			}
			identities, err := readCommonLiveIdentitiesInTx(ctx, tx, limit)
			if mode == "scan-error" {
				if err == nil || identities != nil {
					t.Fatalf("invalid scan returned identities: %v, %v", identities, err)
				}
				return
			}
			if want != nil {
				if !errors.Is(err, want) || identities != nil {
					t.Fatalf("partial header mapping/cause: %v, %v", identities, err)
				}
				return
			}
			if err != nil || len(identities) != count {
				t.Fatalf("boundary mapping: %d != %d, %v", len(identities), count, err)
			}
		})
	}
}

func TestResourceQueryRejectsBeforeDatabaseAccess(t *testing.T) {
	for _, mode := range []string{"nil-context", "canceled", "layout11"} {
		t.Run(mode, func(t *testing.T) {
			s := &Store{options: Options{Binding: Binding{SchemaVersion: TypesFreshSchemaVersion}}}
			ctx := context.Background()
			var want error
			switch mode {
			case "nil-context":
				ctx = nil
			case "canceled":
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
				want = context.Canceled
			case "layout11":
				s.options.Binding.SchemaVersion = SchemaVersion
				want = ErrCapabilityUnavailable
			}
			snapshot, err := s.ReadResourceQuerySnapshot(ctx)
			if err == nil || want != nil && !errors.Is(err, want) || !reflect.DeepEqual(snapshot, ResourceSnapshot{}) {
				t.Fatalf("pre-database refusal: %+v, %v", snapshot, err)
			}
		})
	}
}

// This invokes the PUBLIC entry and existing transaction/session policy. A
// refused actual binding must release its transaction/connection, preserve the
// rollback cause, and never enter preflight or return any captured result.
func TestResourceQueryBindingFailureRollsBack(t *testing.T) {
	for _, mode := range []string{"binding-refusal", "rollback-error"} {
		t.Run(mode, func(t *testing.T) {
			s, mock := transactionPolicyMock(t)
			s.options.Binding.SchemaVersion = TypesFreshSchemaVersion
			mock.ExpectBegin()
			mock.ExpectQuery(`SELECT workspace, scope_url, authority_id,
        schema_version, writer_token FROM graph_preview_scope WHERE singleton = 1
 AND OCTET_LENGTH(workspace)=? AND OCTET_LENGTH(scope_url)=?`).WithArgs(0, 0).WillReturnError(errors.New("binding unavailable"))
			rollbackFault := errors.New("read rollback unavailable")
			if mode == "rollback-error" {
				mock.ExpectRollback().WillReturnError(rollbackFault)
			} else {
				mock.ExpectRollback()
			}
			snapshot, err := s.ReadResourceQuerySnapshot(context.Background())
			if !errors.Is(err, ErrInvalidStore) || !reflect.DeepEqual(snapshot, ResourceSnapshot{}) {
				t.Fatalf("binding failure exposed a result: %+v, %v", snapshot, err)
			}
			if mode == "rollback-error" && !errors.Is(err, rollbackFault) {
				t.Fatalf("lost rollback cause: %v", err)
			}
			if s.db.Stats().InUse != 0 {
				t.Fatal("Resource query leaked connection")
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
