package graphstore

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/typecatalog"
	"github.com/steveyegge/beads/resourceget"
)

// These are projector controls with checked finite compiler inputs and exact
// retained SQLmock evidence; they do not qualify a native store or admission.
func TestNativeProjectionEmitsOnlyActualNominalOwnedBuckets(t *testing.T) {
	const scope = "https://scope.example/"
	owner := typecatalog.Reference{URI: scope + "owner-type", Revision: "owner-type-edition"}
	nonowner := typecatalog.Reference{URI: scope + "nonowner-type", Revision: "nonowner-type-edition"}
	family := typecatalog.Reference{URI: scope + "link-family", Revision: "family-edition"}
	specialized := typecatalog.Reference{URI: scope + "specialized-link", Revision: "specialized-edition"}
	endpoint := typecatalog.EndpointConstraint{ResourceKind: "bead", Scope: "same", Types: typecatalog.Selector{Any: true}}
	pair := []typecatalog.EndpointPair{{Source: endpoint, Target: endpoint}}
	catalog, err := typecatalog.New([]typecatalog.Definition{
		{Identity: owner, Name: "owner", Kind: graph.KindBead, NominalOnly: true, OwnsOutgoing: []typecatalog.Selector{{AnyOf: []typecatalog.Reference{family}}}},
		{Identity: nonowner, Name: "nonowner", Kind: graph.KindBead, NominalOnly: true},
		{Identity: family, Name: "link family", Kind: graph.KindLink, NominalOnly: true, Abstract: true, EndpointPairs: pair},
		{Identity: specialized, Name: "specialized", Kind: graph.KindLink, NominalOnly: true, Parents: []typecatalog.Reference{family}, EndpointPairs: pair},
	})
	if err != nil {
		t.Fatal(err)
	}
	checked := checkedInstalledTypes{catalog: catalog, facts: []ResourceTypeFact{
		{Reference: resourceget.TypeReference{URI: family.URI, Revision: family.Revision}, Describes: graph.KindLink},
		{Reference: resourceget.TypeReference{URI: specialized.URI, Revision: specialized.Revision}, Describes: graph.KindLink},
	}}
	link, err := canonicalJSON(map[string]any{"id": scope + "actual-link", "type": specialized.URI, "revision": "link-edition", "version": "link-edition", "source": scope + "owner", "target": scope + "target", "properties": map[string]any{}, "attribution": map[string]any{"actor": "", "status": "unknown"}})
	if err != nil {
		t.Fatal(err)
	}
	retained, err := encodeNativeSnapshotEdition(link, resourceget.TypeReference{URI: specialized.URI, Revision: specialized.Revision})
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name        string
		reference   typecatalog.Reference
		captures    []json.RawMessage
		wantPresent bool
		wantBuckets int
		refuse      bool
	}{
		{name: "owner-zero", reference: owner, captures: []json.RawMessage{}, wantPresent: true},
		{name: "one-specialized", reference: owner, captures: []json.RawMessage{link}, wantPresent: true, wantBuckets: 1},
		{name: "true-nonowner", reference: nonowner, captures: []json.RawMessage{}},
		{name: "nonowner-capture-refused", reference: nonowner, captures: []json.RawMessage{link}, refuse: true},
	} {
		t.Run(test.name, func(t *testing.T) {
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
			if test.wantBuckets != 0 {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM graph_preview_catalog WHERE COALESCE(public_id,path)=? AND resource_kind='link'`)).WithArgs("actual-link").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT path FROM graph_preview_catalog WHERE COALESCE(public_id,path)=? AND resource_kind='link'`)).WithArgs("actual-link").WillReturnRows(sqlmock.NewRows([]string{"path"}).AddRow("links/actual-link"))
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT OCTET_LENGTH(snapshot) FROM graph_preview_versions WHERE path=? AND version=?`)).WithArgs("links/actual-link", "link-edition").WillReturnRows(sqlmock.NewRows([]string{"bytes"}).AddRow(len(retained)))
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT snapshot FROM graph_preview_versions WHERE path=? AND version=? AND OCTET_LENGTH(snapshot)=?`)).WithArgs("links/actual-link", "link-edition", len(retained)).WillReturnRows(sqlmock.NewRows([]string{"snapshot"}).AddRow(retained))
			}
			raw, err := canonicalJSON(map[string]any{"id": scope + "owner", "type": test.reference.URI, "revision": "owner-edition", "version": "owner-edition", "properties": map[string]any{}, "owned": test.captures, "attribution": map[string]any{"actor": "", "status": "unknown"}})
			if err != nil {
				t.Fatal(err)
			}
			document, err := projectNativeResourceInTx(context.Background(), tx, scope, graph.KindBead, raw, resourceget.TypeReference{URI: test.reference.URI, Revision: test.reference.Revision}, checked, newIntrinsicMetadataBudget())
			if test.refuse {
				if !errors.Is(err, ErrInvalidStore) || document.ID() != "" {
					t.Fatalf("unowned capture yielded a partial Resource: %s, %v", document.ID(), err)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				var fields map[string]json.RawMessage
				if err := json.Unmarshal(document.NativeJSON(), &fields); err != nil {
					t.Fatal(err)
				}
				owned, present := fields["ownedLinks"]
				if present != test.wantPresent {
					t.Fatalf("ownedLinks presence = %v, want %v: %s", present, test.wantPresent, document.NativeJSON())
				}
				if present {
					var groups map[string][]json.RawMessage
					if err := json.Unmarshal(owned, &groups); err != nil || groups == nil || len(groups) != test.wantBuckets {
						t.Fatalf("manufactured selector buckets: %s, %v", owned, err)
					}
					if test.wantBuckets != 0 {
						if len(groups[specialized.URI]) != 1 {
							t.Fatalf("lost actual nominal bucket: %s", owned)
						}
						var captured struct {
							Type resourceget.TypeReference `json:"type"`
						}
						if err := json.Unmarshal(groups[specialized.URI][0], &captured); err != nil || captured.Type != (resourceget.TypeReference{URI: specialized.URI, Revision: specialized.Revision}) {
							t.Fatalf("lost captured exact affiliation: %s, %v", groups[specialized.URI][0], err)
						}
					}
				}
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
