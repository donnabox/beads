package graphstore

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	graph "github.com/steveyegge/beads/graphops"
)

func TestCustomBeadTypeGeneratedRecord(t *testing.T) {
	_, plan := intrinsicCollectionFixture(t)
	ownType, err := customDefinitionAffiliation(plan)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := generatedBeadTypeDocument(plan.Scope().URI, "links/Foo", strings.Repeat("d", 32), ownType)
	if err != nil {
		t.Fatal(err)
	}
	var body struct {
		ID, Revision string
		Type         struct{ URI, Revision string }
		Properties   struct {
			Schema map[string]any `json:"propertiesSchema"`
		}
		Owned map[string]json.RawMessage `json:"ownedLinks"`
	}
	if err = json.Unmarshal(doc.NativeJSON(), &body); err != nil {
		t.Fatal(err)
	}
	local, ok := plan.LocalFor(ownType)
	if !ok || body.ID != plan.Scope().URI+"links/Foo" || doc.Kind() != "bead" || body.Type.URI != ownType.URI || body.Type.URI == local.URI || body.Type.Revision != ownType.Revision {
		t.Fatal("lost whole ID, actual kind or shared exact affiliation")
	}
	if body.Properties.Schema["$schema"] != "https://json-schema.org/draft/2020-12/schema" || body.Properties.Schema["type"] != "object" || len(body.Owned) != 0 {
		t.Fatal("changed supported schema or invented first Refines group")
	}
	for _, id := range []string{"type", "BeadType", "link-type", "refines"} {
		if _, err := (&Store{options: intrinsicProfileOptions()}).CreateBeadType(context.Background(), id, "actor"); !errors.Is(err, graph.ErrValidation) {
			t.Fatalf("reserved %q: %v", id, err)
		}
	}
	if _, err := (&Store{options: Options{Binding: Binding{SchemaVersion: SchemaVersion}}}).CreateBeadType(context.Background(), "foo", "actor"); !errors.Is(err, ErrCapabilityUnavailable) {
		t.Fatalf("old layout gained writer: %v", err)
	}
}

func TestCustomDefinitionContextCodec(t *testing.T) {
	_, plan := intrinsicCollectionFixture(t)
	ownType, err := customDefinitionAffiliation(plan)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := encodeCustomDefinitionContext(plan.Scope().ContextID, ownType)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := decodeCustomDefinitionContext(raw)
	if err != nil || decoded.ContextID != plan.Scope().ContextID || decoded.Type != ownType {
		t.Fatalf("codec roundtrip: %+v %v", decoded, err)
	}
	for _, tc := range []struct {
		name   string
		mutate func([]byte) []byte
	}{
		{"unknown", func(b []byte) []byte { return append([]byte(`{"extra":1,`), b[1:]...) }},
		{"wrong-encoding", func(b []byte) []byte {
			return bytes.Replace(b, []byte(customDefinitionEncoding), []byte("type-definition-context/2"), 1)
		}},
		{"wrong-case", func(b []byte) []byte { return bytes.Replace(b, []byte(`"contextID"`), []byte(`"ContextID"`), 1) }},
		{"invalid-context", func(b []byte) []byte { return bytes.Replace(b, []byte(plan.Scope().ContextID), []byte("short"), 1) }},
		{"missing-pin", func(b []byte) []byte { return bytes.Replace(b, []byte(`"revision":"bead-type-1",`), nil, 1) }},
		{"duplicate", func(b []byte) []byte { return append([]byte(`{"encoding":"bad",`), b[1:]...) }},
		{"null", func([]byte) []byte { return []byte("null") }},
		{"oversize", func(b []byte) []byte { return append(b, bytes.Repeat([]byte(" "), 513)...) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out, err := decodeCustomDefinitionContext(tc.mutate(raw))
			if err == nil || out != (customDefinitionContext{}) {
				t.Fatalf("codec mutation escaped: %+v %v", out, err)
			}
		})
	}
}

func TestCustomDefinitionReadCorrespondence(t *testing.T) {
	const aggregate = `SELECT COUNT(*),COUNT(p.path),CAST(LEAST(COALESCE(SUM(512+COALESCE(OCTET_LENGTH(c.public_id),0)+COALESCE(OCTET_LENGTH(c.type_url),0)+COALESCE(OCTET_LENGTH(p.properties),0)+COALESCE(OCTET_LENGTH(p.metadata),0)),0),?) AS UNSIGNED) FROM graph_preview_catalog c LEFT JOIN graph_preview_payloads p ON p.path=c.path WHERE c.path=? AND c.backing='type-definition'`
	const header = `SELECT c.public_id,c.backing_key,c.revision,c.type_url,c.allocation_state,p.metadata,OCTET_LENGTH(p.properties) FROM graph_preview_catalog c JOIN graph_preview_payloads p ON p.path=c.path WHERE c.path=? AND c.backing='type-definition' AND c.resource_kind='bead' AND OCTET_LENGTH(p.metadata)<=512`
	_, plan := intrinsicCollectionFixture(t)
	ownType, err := customDefinitionAffiliation(plan)
	if err != nil {
		t.Fatal(err)
	}
	o := intrinsicProfileOptions()
	o.Binding.ScopeURL = plan.Scope().URI
	o.Binding.AuthorityID = plan.Scope().AuthorityID
	key, revision := strings.Repeat("c", 32), strings.Repeat("d", 32)
	path := "beads/" + key
	doc, err := generatedBeadTypeDocument(o.Binding.ScopeURL, "links/Foo", revision, ownType)
	if err != nil {
		t.Fatal(err)
	}
	raw := doc.NativeJSON()
	codec, err := encodeCustomDefinitionContext(plan.Scope().ContextID, ownType)
	if err != nil {
		t.Fatal(err)
	}
	fault := errors.New("custom definition driver fault")
	for _, mode := range []string{"success", "missing-payload", "over-budget", "budget", "wrong-type", "wrong-context", "wrong-pin", "wrong-key", "deleted", "unknown-edition", "extra-edition", "wrong-body", "body-query-error", "extra-history", "wrong-snapshot", "missing-snapshot"} {
		t.Run(mode, func(t *testing.T) {
			tx, mock := intrinsicProfileTransaction(t)
			limit := PreviewCurrentReadByteLimit
			if mode == "budget" {
				limit = 0
			} else {
				pcount, total := int64(1), int64(512+len(raw)+len(codec)+len("links/Foo")+len(ownType.URI))
				if mode == "missing-payload" {
					pcount = 0
				}
				if mode == "over-budget" {
					total = PreviewCurrentReadByteLimit
				}
				mock.ExpectQuery(aggregate).WithArgs(PreviewCurrentReadByteLimit-4096+1, path).WillReturnRows(sqlmock.NewRows([]string{"count", "payload_count", "total"}).AddRow(int64(1), pcount, total)).RowsWillBeClosed()
				if mode != "missing-payload" && mode != "over-budget" {
					typ, k, state, meta := ownType.URI, key, "live", codec
					if mode == "wrong-type" {
						typ = typ[:len(typ)-1] + "X"
					}
					if mode == "wrong-key" {
						k = strings.Repeat("e", 32)
					}
					if mode == "deleted" {
						state = "deleted"
					}
					if mode == "wrong-context" {
						meta, _ = encodeCustomDefinitionContext(strings.Repeat("f", 32), ownType)
					}
					if mode == "wrong-pin" {
						wrong := ownType
						wrong.Revision = "bead-type-0"
						meta, _ = encodeCustomDefinitionContext(plan.Scope().ContextID, wrong)
					}
					mock.ExpectQuery(header).WithArgs(path).WillReturnRows(sqlmock.NewRows([]string{"id", "key", "revision", "type", "state", "metadata", "length"}).AddRow("links/Foo", k, revision, typ, state, meta, int64(len(raw)))).RowsWillBeClosed()
					switch mode {
					case "unknown-edition", "extra-edition":
						n := 0
						if mode == "extra-edition" {
							n = 1
						}
						mock.ExpectQuery(`SELECT COUNT(*) FROM graph_preview_versions WHERE path=? AND version=?`).WithArgs(path, "unavailable").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(n)).RowsWillBeClosed()
					case "wrong-type", "wrong-context", "wrong-pin", "wrong-key", "deleted": // strict no full-body query on these failures
					default:
						body := raw
						if mode == "wrong-body" {
							body = bytes.Replace(raw, []byte(`"type":"object"`), []byte(`"type":"number"`), 1)
						}
						expected := mock.ExpectQuery(`SELECT properties FROM graph_preview_payloads WHERE path=? AND OCTET_LENGTH(properties)=?`).WithArgs(path, len(raw))
						if mode == "body-query-error" {
							expected.WillReturnError(fault)
						} else {
							expected.WillReturnRows(sqlmock.NewRows([]string{"body"}).AddRow(body)).RowsWillBeClosed()
							if mode != "wrong-body" {
								count := 1
								if mode == "extra-history" {
									count = 2
								}
								mock.ExpectQuery(`SELECT COUNT(*) FROM graph_preview_versions WHERE path=?`).WithArgs(path).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(count)).RowsWillBeClosed()
								if mode != "extra-history" {
									q := mock.ExpectQuery(`SELECT snapshot FROM graph_preview_versions WHERE path=? AND version=? AND ordinal=1 AND OCTET_LENGTH(snapshot)=?`).WithArgs(path, revision, len(raw))
									if mode == "missing-snapshot" {
										q.WillReturnError(sql.ErrNoRows)
									} else {
										snapshot := raw
										if mode == "wrong-snapshot" {
											snapshot = append([]byte(nil), raw...)
											snapshot[0] = '!'
										}
										q.WillReturnRows(sqlmock.NewRows([]string{"snapshot"}).AddRow(snapshot)).RowsWillBeClosed()
									}
								}
							}
						}
					}
				}
			}
			requested := ""
			if mode == "unknown-edition" || mode == "extra-edition" {
				requested = "unavailable"
			}
			budget, _ := newIntrinsicMetadataBudget(limit)
			actual, err := readCustomDefinitionInTx(context.Background(), tx, o, path, requested, plan, budget)
			if mode == "success" {
				if err != nil || !bytes.Equal(actual.NativeJSON(), raw) {
					t.Fatalf("readback: %v", err)
				}
			} else {
				if err == nil || actual.ID() != "" {
					t.Fatalf("failed read returned success: %s %v", actual.ID(), err)
				}
				if mode == "unknown-edition" && !errors.Is(err, ErrVersionUnknown) {
					t.Fatalf("unknown edition: %v", err)
				}
				if mode == "body-query-error" && !errors.Is(err, fault) {
					t.Fatalf("driver cause lost: %v", err)
				}
				if (mode == "budget" || mode == "over-budget") && !errors.Is(err, ErrLimitExceeded) {
					t.Fatalf("wrong budget diagnosis: %v", err)
				}
			}
		})
	}
}

func TestCustomDefinitionInventoryRefusesOmissions(t *testing.T) {
	const aggregate = `SELECT COUNT(*),COUNT(p.path),CAST(LEAST(COALESCE(SUM(512+OCTET_LENGTH(c.path)+COALESCE(OCTET_LENGTH(c.public_id),0)+COALESCE(OCTET_LENGTH(p.properties),0)+COALESCE(OCTET_LENGTH(p.metadata),0)),0),?) AS UNSIGNED) FROM graph_preview_catalog c LEFT JOIN graph_preview_payloads p ON p.path=c.path WHERE c.backing='type-definition'`
	_, plan := intrinsicCollectionFixture(t)
	o := intrinsicProfileOptions()
	o.Binding.ScopeURL = plan.Scope().URI
	o.Binding.AuthorityID = plan.Scope().AuthorityID
	fault := errors.New("inventory row fault")
	for _, mode := range []string{"empty", "missing-payload", "overflow", "extra-path", "missing-path", "row-error", "close-error", "extra-result", "budget"} {
		t.Run(mode, func(t *testing.T) {
			tx, mock := intrinsicProfileTransaction(t)
			limit := PreviewCurrentReadByteLimit
			if mode == "budget" {
				limit = 0
			} else {
				count, payload, total := int64(0), int64(0), int64(0)
				switch mode {
				case "missing-payload":
					count = 1
				case "overflow":
					count = PreviewCurrentReadByteLimit
					payload = count
				case "missing-path", "row-error", "close-error", "extra-result":
					count = 1
					payload = 1
					total = 512
				}
				mock.ExpectQuery(aggregate).WithArgs(PreviewCurrentReadByteLimit - 4096 + 1).WillReturnRows(sqlmock.NewRows([]string{"count", "payload_count", "total"}).AddRow(count, payload, total)).RowsWillBeClosed()
				if mode != "missing-payload" && mode != "overflow" {
					rows := sqlmock.NewRows([]string{"path"})
					if mode == "extra-path" || mode == "row-error" || mode == "close-error" || mode == "extra-result" {
						rows.AddRow("beads/" + strings.Repeat("c", 32))
					}
					if mode == "row-error" {
						rows.RowError(0, fault)
					}
					if mode == "close-error" {
						rows.CloseError(fault)
					}
					expected := mock.ExpectQuery(`SELECT path FROM graph_preview_catalog WHERE backing='type-definition' ORDER BY BINARY path LIMIT ?`).WithArgs(uint64(count + 1))
					if mode == "extra-result" {
						expected.WillReturnRows(rows, sqlmock.NewRows([]string{"path"}).AddRow("beads/other")).RowsWillBeClosed()
					} else {
						expected.WillReturnRows(rows).RowsWillBeClosed()
					}
				}
			}
			budget, _ := newIntrinsicMetadataBudget(limit)
			docs, err := (&Store{options: o}).listCustomDefinitionsInTx(context.Background(), tx, plan, budget)
			if mode == "empty" {
				if err != nil || len(docs) != 0 {
					t.Fatalf("empty: %v %v", docs, err)
				}
			} else {
				if err == nil || docs != nil {
					t.Fatalf("inventory failure returned records: %v %v", docs, err)
				}
				if (mode == "row-error" || mode == "close-error") && !errors.Is(err, fault) {
					t.Fatalf("lost inventory driver cause: %v", err)
				}
			}
		})
	}
}

// The same allowance pays intrinsic selection and subsequent custom acquisition.
// Existing strict mocks would fail on any body/query after this exhausted debit.
func TestCustomDefinitionSharedCollectionAllowance(t *testing.T) {
	catalog, plan := intrinsicCollectionFixture(t)
	tx, mock := intrinsicProfileTransaction(t)
	budget, err := newIntrinsicMetadataBudget(PreviewCurrentReadByteLimit)
	if err != nil {
		t.Fatal(err)
	}
	intrinsic, err := selectIntrinsicDefinitionBeads(plan.Scope().URI, catalog, plan, "type", "", budget)
	if err != nil || len(intrinsic) != 4 {
		t.Fatalf("intrinsic selection: %d %v", len(intrinsic), err)
	}
	// Retain a valid already-used operation allowance with only one byte left;
	// custom bookkeeping must refuse before any SQL or retaining another body.
	if err := budget.chargeBody(budget.remaining - 1); err != nil {
		t.Fatal(err)
	}
	docs, err := (&Store{options: intrinsicProfileOptions()}).listCustomDefinitionsInTx(context.Background(), tx, plan, budget)
	if !errors.Is(err, ErrLimitExceeded) || docs != nil {
		t.Fatalf("shared allowance escaped: %d %v", len(docs), err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
