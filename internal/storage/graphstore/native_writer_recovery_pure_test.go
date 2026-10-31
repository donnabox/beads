package graphstore

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/steveyegge/beads/resourceget"
)

func TestNativeRecoveryUsesOnlyTerminalLastLivePin(t *testing.T) {
	for _, name := range []string{"memory", "informational", "missing-pin", "wrong-nominal", "wrong-id", "wrong-revision", "wrong-inception", "nonadjacent-predecessor", "over-budget", "driver-error"} {
		t.Run(name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			mock.ExpectBegin()
			tx, err := db.BeginTx(context.Background(), nil)
			if err != nil {
				t.Fatal(err)
			}
			scope := "https://example.test/work/"
			s := &Store{options: Options{Binding: Binding{SchemaVersion: TypesFreshSchemaVersion, ScopeURL: scope}}}
			path, kind, backing := "beads/subject", "bead", "generic"
			if name == "informational" {
				path, kind, backing = "links/subject", "link", "informational"
			}
			liveVersion, terminalVersion := strings.Repeat("1", 32), strings.Repeat("2", 32)
			pin := resourceget.TypeReference{URI: scope + "memory", Revision: strings.Repeat("a", 32)}
			attribution := Attribution{Inception: "create", Actor: "writer", Status: "claimed", RecordedAt: "2026-10-30T00:00:00Z"}
			terminal := ResourceTerminalState{ID: scope + "subject", Revision: terminalVersion, Version: terminalVersion, State: "deleted", PreviousVersion: liveVersion, Attribution: attribution}
			if kind == "link" {
				terminal.Type = pin.URI
			}
			raw, err := canonicalJSON(terminal)
			if err != nil {
				t.Fatal(err)
			}
			live := map[string]any{"id": terminal.ID, "type": pin.URI, "revision": liveVersion, "version": liveVersion, "properties": map[string]any{}, "attribution": attribution}
			switch name {
			case "wrong-id":
				live["id"] = scope + "other"
			case "wrong-revision":
				live["revision"] = strings.Repeat("3", 32)
			case "wrong-inception":
				other := attribution
				other.Inception = "revert"
				live["attribution"] = other
			}
			domain, err := canonicalJSON(live)
			if err != nil {
				t.Fatal(err)
			}
			retained, err := encodeNativeSnapshotEdition(domain, pin)
			if err != nil {
				t.Fatal(err)
			}
			if name == "missing-pin" {
				retained = domain
			}
			mock.ExpectQuery("SELECT CAST").WillReturnRows(sqlmock.NewRows([]string{"bytes"}).AddRow(256))
			mock.ExpectQuery("SELECT path,resource_kind,allocation_state,backing").WithArgs(path).WillReturnRows(sqlmock.NewRows([]string{"path", "kind", "state", "backing", "id"}).AddRow(path, kind, "live", backing, "subject"))
			query := mock.ExpectQuery("SELECT OCTET_LENGTH\\(snapshot\\)\\+OCTET_LENGTH\\(actor\\).*version=\\? AND ordinal=\\?").WithArgs(path, liveVersion, int64(1))
			fault := errors.New("retained acquisition fault")
			if name == "nonadjacent-predecessor" {
				query.WillReturnRows(sqlmock.NewRows([]string{"size"}))
			} else if name == "driver-error" {
				query.WillReturnError(fault)
			} else {
				query.WillReturnRows(sqlmock.NewRows([]string{"size"}).AddRow(len(retained) + 6))
			}
			budget := newIntrinsicMetadataBudget()
			if name == "over-budget" {
				if err := budget.chargeBody(PreviewCurrentReadByteLimit); err != nil {
					t.Fatal(err)
				}
			}
			if name != "nonadjacent-predecessor" && name != "driver-error" && name != "over-budget" {
				mock.ExpectQuery("SELECT snapshot FROM graph_preview_versions").WithArgs(path, liveVersion, int64(1), len(retained)+6).WillReturnRows(sqlmock.NewRows([]string{"snapshot"}).AddRow(retained))
			}
			nominal := pin.URI
			if name == "wrong-nominal" {
				nominal = scope + "other-type"
			}
			got, err := s.nativePredecessorTypeInTx(context.Background(), tx, path, nominal, terminalVersion, "writer", 2, raw, budget)
			if name == "memory" || name == "informational" {
				if err != nil || got != pin {
					t.Fatalf("last-live pin lost: %+v %v", got, err)
				}
			} else if err == nil || got != (resourceget.TypeReference{}) {
				t.Fatalf("partial or accepted corrupt predecessor: %+v %v", got, err)
			}
			if name == "driver-error" && !errors.Is(err, fault) {
				t.Fatalf("lost driver cause: %v", err)
			}
			if name == "over-budget" && !errors.Is(err, ErrLimitExceeded) {
				t.Fatalf("lost budget refusal: %v", err)
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
