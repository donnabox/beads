package graphstore

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestIntrinsicExpectationFixedNonreadySource(t *testing.T) {
	m, err := loadIntrinsicExpected()
	if err != nil {
		t.Fatal(err)
	}
	if m.Ready || len(m.Unfilled) != 4 || len(m.Objects) != 44 || len(m.Sources) != 65 {
		t.Fatal("source boundary changed")
	}
	// The ignored-lane migration appends granted_node after heartbeat_at even
	// on fresh main-created leases; this is exact source order, not an allowance
	// for arbitrary native extra columns or a runtime-canonical assertion.
	const leaseMigration = "internal/storage/schema/migrations/ignored/0016_add_lease_granted_node.up.sql"
	leaseSource, ok := m.Sources[leaseMigration]
	if !ok || leaseSource.GitBlob != "018f15701fcd47a769c405d7e54b2932e648a2c8" || leaseSource.SHA256 != "31a55421b78d8c90db6cfba5bf8e99e87d6123271575282f761aa077e5d36386" {
		t.Fatal("leases final migration source missing")
	}
	leaseChecked := false
	for _, object := range m.Objects {
		if object.Name != "leases" {
			continue
		}
		if !strings.Contains(object.Create, "heartbeat_at DATETIME NOT NULL,\n    granted_node VARCHAR(255) NOT NULL DEFAULT '',\n    INDEX idx_leases_expires") {
			t.Fatal("leases final column/order/default differs from selected source")
		}
		if len(object.Sources) != 4 || object.Sources[3] != leaseMigration {
			t.Fatal("leases construction correspondence differs")
		}
		leaseChecked = true
	}
	if !leaseChecked {
		t.Fatal("leases source expectation missing")
	}
	first := m.Objects[0].Name
	m.Objects[0].Name = "caller mutation"
	m.Sources["caller"] = intrinsicSourceIdentity{}
	again, err := loadIntrinsicExpected()
	if err != nil || again.Objects[0].Name != first || len(again.Sources) != 65 {
		t.Fatalf("shared mutable expectation: %v", err)
	}
	for _, tc := range []struct {
		name   string
		mutate func([]byte) []byte
	}{
		{"ready", func(b []byte) []byte {
			return bytes.Replace(b, []byte(`"completeRuntimeComparisonReady": false`), []byte(`"completeRuntimeComparisonReady": true`), 1)
		}},
		{"table", func(b []byte) []byte {
			return bytes.Replace(b, []byte(`"bd_events_journal"`), []byte(`"other_table"`), 1)
		}},
		{"leases", func(b []byte) []byte {
			return bytes.Replace(b, []byte("granted_node VARCHAR(255)"), []byte("unreviewed_node VARCHAR(255)"), 1)
		}},
		{"unknown", func(b []byte) []byte { return append(b, []byte(` {"unexpected":true}`)...) }},
		{"whitespace", func(b []byte) []byte { return append(b, ' ') }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := tc.mutate(append([]byte(nil), intrinsicExpectedBytes...))
			if bytes.Equal(input, intrinsicExpectedBytes) {
				t.Fatal("mutation did not change bytes")
			}
			got, err := decodeIntrinsicExpected(input)
			if !errors.Is(err, ErrInvalidStore) || !reflect.DeepEqual(got, intrinsicExpectedManifest{}) {
				t.Fatalf("untrusted source accepted: %v", err)
			}
		})
	}
}

func TestIntrinsicMetadataBudgetAndScalars(t *testing.T) {
	b := &intrinsicMetadataBudget{remaining: 300}
	if err := b.charge(256); err != nil {
		t.Fatal(err)
	}
	if err := b.charge(45); !errors.Is(err, ErrLimitExceeded) || b.remaining != 44 {
		t.Fatalf("shared budget: %+v %v", b, err)
	}
	if fixed := newIntrinsicMetadataBudget(); fixed.remaining != PreviewCurrentReadByteLimit || fixed.metadata != 0 {
		t.Fatalf("production allowance changed: %+v", fixed)
	}
	// Invalid-state refusal belongs at the actual debit boundary now that the
	// production factory has no variable input. Both planes must refuse before
	// acquisition, without changing the corrupted private test state.
	for _, remaining := range []int{-1, PreviewCurrentReadByteLimit + 1} {
		for _, plane := range []string{"metadata", "body"} {
			invalid := &intrinsicMetadataBudget{remaining: remaining}
			before := *invalid
			var err error
			if plane == "metadata" {
				err = invalid.charge(0)
			} else {
				err = invalid.chargeBody(0)
			}
			if !errors.Is(err, ErrLimitExceeded) || *invalid != before {
				t.Fatalf("invalid %s allowance changed/admitted: %+v %v", plane, invalid, err)
			}
		}
	}
	metadata := newIntrinsicMetadataBudget()
	if err := metadata.charge(4 << 20); err != nil {
		t.Fatal(err)
	}
	if err := metadata.charge(1); !errors.Is(err, ErrLimitExceeded) || metadata.remaining != PreviewCurrentReadByteLimit-(4<<20) {
		t.Fatalf("metadata sublimit escaped: %+v %v", metadata, err)
	}
	for _, value := range []any{nil, 1, false} {
		if _, err := intrinsicCellString(value); !errors.Is(err, ErrInvalidStore) {
			t.Fatalf("non-text/nil metadata accepted: %#v %v", value, err)
		}
	}
	for _, value := range []any{nil, -1, int8(-1), int16(-1), int32(-1), int64(-1), float64(1), "+1", " 1", "1.0", "18446744073709551616"} {
		if _, err := intrinsicUnsigned(value); !errors.Is(err, ErrInvalidStore) {
			t.Fatalf("scalar accepted: %#v", value)
		}
	}
	for _, value := range []any{int8(1), int16(1), int32(1), int64(1), uint8(1), uint16(1), uint32(1), uint64(1), "1", []byte("1")} {
		if got, err := intrinsicUnsigned(value); err != nil || got != 1 {
			t.Fatalf("scalar refused: %#v %v", value, err)
		}
	}
	if intrinsicQuote("a`b") != "`a``b`" {
		t.Fatal("identifier not quoted literally")
	}
}

func TestIntrinsicSQLModeClosedSet(t *testing.T) {
	if err := intrinsicSQLModes("strict_trans_tables, ONLY_FULL_GROUP_BY,no_engine_substitution"); err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"", "ONLY_FULL_GROUP_BY", "STRICT_TRANS_TABLES,ONLY_FULL_GROUP_BY,NO_ENGINE_SUBSTITUTION,ANSI_QUOTES", "STRICT_TRANS_TABLES,ONLY_FULL_GROUP_BY,NO_ENGINE_SUBSTITUTION,STRICT_TRANS_TABLES"} {
		if err := intrinsicSQLModes(mode); !errors.Is(err, ErrInvalidStore) {
			t.Fatalf("mode accepted %q: %v", mode, err)
		}
	}
}

func TestIntrinsicNativeIgnoreBooleanScalars(t *testing.T) {
	for _, v := range []any{int8(0), int64(0)} {
		got, err := intrinsicIgnoreBoolean(v)
		if err != nil || got {
			t.Fatalf("zero %T: %v %v", v, got, err)
		}
	}
	for _, v := range []any{int8(1), int64(1)} {
		got, err := intrinsicIgnoreBoolean(v)
		if err != nil || !got {
			t.Fatalf("one %T: %v %v", v, got, err)
		}
	}
	for _, v := range []any{int8(-1), int8(2), int64(-1), float64(1), nil, "true"} {
		if _, err := intrinsicIgnoreBoolean(v); !errors.Is(err, ErrInvalidStore) {
			t.Fatalf("invalid %T(%v): %v", v, v, err)
		}
	}
}

func TestIntrinsicCheckedProfileBinding(t *testing.T) {
	for _, tc := range []struct {
		name          string
		fact, binding int
		valid         bool
	}{
		{"selected", TypesFreshSchemaVersion, TypesFreshSchemaVersion, true},
		{"zero-fact", 0, TypesFreshSchemaVersion, false},
		{"wrong-fact", SchemaVersion, TypesFreshSchemaVersion, false},
		{"wrong-binding", TypesFreshSchemaVersion, SchemaVersion, false},
		{"matching-legacy-is-not-checked14", SchemaVersion, SchemaVersion, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := (checkedStoreProfile{schemaVersion: tc.fact}).requireBinding(Binding{SchemaVersion: tc.binding})
			if tc.valid {
				if err != nil {
					t.Fatal(err)
				}
			} else if !errors.Is(err, ErrInvalidStore) {
				t.Fatalf("unqualified/mismatched profile fact admitted: %v", err)
			}
		})
	}
}
