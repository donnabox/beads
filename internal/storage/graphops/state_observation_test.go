package graphops

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

var stateColumns = []string{"database", "branch", "scope", "history", "descriptors", "beads", "links", "seq", "events", "allocations"}

func stateLabels() [8]string {
	var labels [8]string
	for i, c := range "012345uv" {
		labels[i] = strings.Repeat(string(c), 32)
	}
	return labels
}

func stateValues() []driver.Value {
	values := []driver.Value{"graph_fixture", "main"}
	for _, label := range stateLabels() {
		values = append(values, label)
	}
	return values
}

func stateContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	t.Cleanup(cancel)
	return ctx
}

func TestStateVersionGoldenAndOrder(t *testing.T) {
	// Independently computed with Python hashlib.sha256 over 256 ASCII bytes.
	const golden = "b002b7191c2bef3e461f08d05c7cd61959e3463edc999a3987eb390bcac35006"
	labels := stateLabels()
	got, err := composeStateVersion(labels)
	if err != nil || got != golden {
		t.Fatalf("digest=%s err=%v", got, err)
	}
	for i := 0; i < len(labels); i++ {
		for j := i + 1; j < len(labels); j++ {
			swapped := labels
			swapped[i], swapped[j] = swapped[j], swapped[i]
			got, err := composeStateVersion(swapped)
			if err != nil || got == golden {
				t.Fatalf("swap %d,%d digest=%s err=%v", i, j, got, err)
			}
		}
	}
	// Hex-only labels are a valid subset of the engine's base32 alphabet.
	labels[7] = strings.Repeat("abcdef01", 4)
	if _, err := composeStateVersion(labels); err != nil {
		t.Fatal(err)
	}
}

func TestStateVersionRejectsMalformedLabels(t *testing.T) {
	for _, bad := range []string{"", strings.Repeat("0", 31), strings.Repeat("0", 33), strings.Repeat("A", 32), strings.Repeat("w", 32), strings.Repeat("z", 32), strings.Repeat("é", 16), strings.Repeat(" ", 32)} {
		for slot := range 8 {
			labels := stateLabels()
			labels[slot] = bad
			got, err := composeStateVersion(labels)
			if got != "" || !errors.Is(err, errCorrupt) {
				t.Fatalf("slot=%d bad=%q digest=%q err=%v", slot, bad, got, err)
			}
		}
	}
}

// Exposes only QueryContext, so the observation cannot obtain transaction/write
// capabilities through its input. The counter pins the one-statement budget.
type stateQueryCounter struct {
	tx    *sql.Tx
	calls int
}

func (q *stateQueryCounter) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	q.calls++
	return q.tx.QueryContext(ctx, query, args...)
}

func TestStateObservationPreservesBoundFacts(t *testing.T) {
	for _, bytes := range []bool{false, true} {
		t.Run(map[bool]string{false: "string", true: "driver bytes"}[bytes], func(t *testing.T) {
			tx, m := mockTx(t)
			values := stateValues()
			values[0], values[1] = "graph_fixture/feature", "feature/Case-Sensitive"
			if bytes {
				for i, value := range values {
					values[i] = []byte(value.(string))
				}
			}
			m.ExpectQuery(regexp.QuoteMeta(stateObservationQuery)).WithoutArgs().WillReturnRows(sqlmock.NewRows(stateColumns).AddRow(values...)).RowsWillBeClosed()
			q := &stateQueryCounter{tx: tx}
			got, err := observeStateInTx(stateContext(t), q)
			if err != nil || got.database != "graph_fixture/feature" || got.branch != "feature/Case-Sensitive" || got.hashes != stateLabels() || got.descriptorHash() != strings.Repeat("2", 32) || got.version != "b002b7191c2bef3e461f08d05c7cd61959e3463edc999a3987eb390bcac35006" || q.calls != 1 {
				t.Fatalf("observation=%+v calls=%d err=%v", got, q.calls, err)
			}
		})
	}
	// One fixed SELECT is essential with a multi-statement-capable driver.
	if strings.Contains(stateObservationQuery, ";") || strings.Count(stateObservationQuery, "SELECT") != 1 {
		t.Fatal("observation lost its single-statement boundary")
	}
	// Independent B4 table inventory pins the mapping, not only the digest order.
	tables := []string{"graph_scope", "graph_scope_history", "graph_type_descriptors", "graph_beads", "graph_links", "graph_ledger_seq", "graph_ledger_events", "graph_allocations"}
	matches := regexp.MustCompile(`DOLT_HASHOF_TABLE\('([^']+)'\)`).FindAllStringSubmatch(stateObservationQuery, -1)
	if len(matches) != len(tables) {
		t.Fatalf("table expressions=%v", matches)
	}
	for i, table := range tables {
		if matches[i][1] != table {
			t.Fatalf("slot %d = %q, want %q", i, matches[i][1], table)
		}
	}
}

func TestStateObservationFailuresReturnNoPartialResult(t *testing.T) {
	boom := errors.New("state engine failure")
	for _, mode := range []string{"query", "scan", "iteration", "close", "empty", "duplicate", "null hash", "malformed hash", "null database", "empty branch", "database overflow", "branch overflow"} {
		t.Run(mode, func(t *testing.T) {
			tx, m := mockTx(t)
			values := stateValues()
			q := m.ExpectQuery(regexp.QuoteMeta(stateObservationQuery)).WithoutArgs()
			rows := sqlmock.NewRows(stateColumns)
			switch mode {
			case "query":
				q.WillReturnError(boom)
			case "scan":
				rows = sqlmock.NewRows([]string{"one"}).AddRow("bad")
			case "iteration":
				rows.AddRow(values...).AddRow(values...).RowError(1, boom)
			case "close":
				rows.AddRow(values...).CloseError(boom)
			case "empty":
			case "duplicate":
				rows.AddRow(values...).AddRow(values...)
			default:
				switch mode {
				case "null hash":
					values[9] = nil
				case "malformed hash":
					values[9] = strings.Repeat("z", 32)
				case "null database":
					values[0] = nil
				case "empty branch":
					values[1] = ""
				case "database overflow":
					values[0] = strings.Repeat("d", 1025)
				case "branch overflow":
					values[1] = strings.Repeat("é", 513)
				}
				rows.AddRow(values...)
			}
			if mode != "query" {
				q.WillReturnRows(rows).RowsWillBeClosed()
			}
			got, err := observeStateInTx(stateContext(t), tx)
			if err == nil || got != (stateObservation{}) {
				t.Fatalf("partial observation=%+v err=%v", got, err)
			}
			if (mode == "query" || mode == "iteration" || mode == "close") && !errors.Is(err, boom) {
				t.Fatalf("lost engine error: %v", err)
			}
			switch mode {
			case "empty", "duplicate", "null hash", "malformed hash", "null database", "empty branch":
				if !errors.Is(err, errCorrupt) || errors.Is(err, errBudget) {
					t.Fatalf("corruption classification: %v", err)
				}
			case "database overflow", "branch overflow":
				if !errors.Is(err, errBudget) || errors.Is(err, errCorrupt) {
					t.Fatalf("budget classification: %v", err)
				}
			}
		})
	}
}

func TestStateObservationRequiresLiveDeadline(t *testing.T) {
	tx, _ := mockTx(t)
	q := &stateQueryCounter{tx: tx}
	if got, err := observeStateInTx(context.Background(), q); !errors.Is(err, errObservationDeadline) || got != (stateObservation{}) {
		t.Fatalf("no deadline: %+v %v", got, err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	cancel()
	if got, err := observeStateInTx(ctx, q); !errors.Is(err, context.Canceled) || got != (stateObservation{}) {
		t.Fatalf("canceled: %+v %v", got, err)
	}
	if q.calls != 0 {
		t.Fatal("invalid context issued SQL")
	}
}

func TestStateObservationCancellationDuringQuery(t *testing.T) {
	tx, m := mockTx(t)
	m.ExpectQuery(regexp.QuoteMeta(stateObservationQuery)).WillDelayFor(time.Second).WillReturnRows(sqlmock.NewRows(stateColumns).AddRow(stateValues()...))
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Millisecond)
	defer cancel()
	got, err := observeStateInTx(ctx, tx)
	if err == nil || ctx.Err() != context.DeadlineExceeded || got != (stateObservation{}) {
		t.Fatalf("query deadline: %+v %v context=%v", got, err, ctx.Err())
	}
}
