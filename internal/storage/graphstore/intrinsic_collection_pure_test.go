package graphstore

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/intrinsicmaterialization"
	"github.com/steveyegge/beads/internal/intrinsictypecatalog"
)

func intrinsicCollectionFixture(t *testing.T) (*intrinsictypecatalog.Catalog, *intrinsicmaterialization.Plan) {
	t.Helper()
	catalog, err := intrinsictypecatalog.LoadSelected()
	if err != nil {
		t.Fatal(err)
	}
	scope := intrinsicmaterialization.Scope{URI: "https://example.test/intrinsic-list/", AuthorityID: strings.Repeat("a", 32), ContextID: strings.Repeat("b", 32)}
	var allocations []intrinsicmaterialization.Allocation
	roles := []intrinsicmaterialization.Role{intrinsicmaterialization.Type, intrinsicmaterialization.BeadType, intrinsicmaterialization.LinkType, intrinsicmaterialization.Refines, intrinsicmaterialization.BeadTypeRefinesType, intrinsicmaterialization.LinkTypeRefinesType}
	for i, role := range roles {
		kind, prefix := graph.KindBead, "beads/"
		if i >= 4 {
			kind, prefix = graph.KindLink, "links/"
		}
		id := "opaque-" + string(role)
		allocations = append(allocations, intrinsicmaterialization.Allocation{Role: role, PublicID: id, PrivatePath: prefix + id, Kind: kind, Local: intrinsicmaterialization.Reference{URI: scope.URI + id, Revision: fmt.Sprintf("edition-%d", i)}, AuthorityID: scope.AuthorityID})
	}
	plan, err := intrinsicmaterialization.Build(catalog, scope, allocations)
	if err != nil {
		t.Fatal(err)
	}
	return catalog, plan
}

func TestIntrinsicDefinitionCollectionFamilies(t *testing.T) {
	catalog, plan := intrinsicCollectionFixture(t)
	const scope = "https://example.test/intrinsic-list/"
	const namespace = "https://github.com/gastownhall/bdp/types/type-system/"
	for _, tc := range []struct {
		name, family, exact string
		want                int
	}{
		{"all", "", "", 4},
		{"Type family", "type", "", 4},
		{"BeadType family", "bead-type", "", 3},
		{"LinkType family", "link-type", "", 1},
		{"global family", namespace + "type", "", 4},
		{"Type exact", "", "type", 0},
		{"BeadType exact", "", "bead-type", 3},
		{"LinkType exact", "", namespace + "link-type", 1},
		{"intersect", "type", "link-type", 1},
		{"disjoint", "bead-type", "link-type", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			budget, _ := newIntrinsicMetadataBudget(PreviewCurrentReadByteLimit)
			docs, err := selectIntrinsicDefinitionBeads(scope, catalog, plan, tc.family, tc.exact, budget)
			if err != nil || len(docs) != tc.want {
				t.Fatalf("want%d got%d: %v", tc.want, len(docs), err)
			}
			byID := map[string][]byte{}
			for _, binding := range plan.Bindings() {
				if binding.Allocation.Kind == graph.KindBead {
					// The carrier's documented canonical encoding can remove JSON
					// Marshal HTML escaping; compare the ENTIRE canonical record.
					expected, err := graph.CanonicalizeJSON(binding.Record)
					if err != nil {
						t.Fatal(err)
					}
					byID[binding.Allocation.Local.URI] = expected
				}
			}
			previous := ""
			for _, doc := range docs {
				if doc.Kind() != "bead" || !bytes.Equal(doc.NativeJSON(), byID[doc.ID()]) {
					t.Fatalf("changed full canonical native definition %s", doc.ID())
				}
				if previous != "" && graph.CompareCodeUnits(previous, doc.ID()) >= 0 {
					t.Fatal("unstable or duplicate ordering")
				}
				previous = doc.ID()
				value := doc.NativeJSON()
				value[0] = '!'
				if !bytes.Equal(doc.NativeJSON(), byID[doc.ID()]) {
					t.Fatal("mutable result")
				}
			}
			if tc.family == "link-type" && docs[0].ID() != scope+"opaque-refines" {
				t.Fatal("category came from definition identity instead of own affiliation")
			}
		})
	}
	for _, selector := range []string{"Type", "refines", "unknown", "https://other.test/type", scope + "opaque-type", "opaque-type", "x/type"} {
		t.Run("refuse-"+selector, func(t *testing.T) {
			budget, _ := newIntrinsicMetadataBudget(PreviewCurrentReadByteLimit)
			docs, err := selectIntrinsicDefinitionBeads(scope, catalog, plan, selector, "", budget)
			if !errors.Is(err, ErrCapabilityUnavailable) || docs != nil {
				t.Fatalf("unsupported selector escaped: %v %v", docs, err)
			}
		})
	}
	budget, _ := newIntrinsicMetadataBudget(0)
	if docs, err := selectIntrinsicDefinitionBeads(scope, catalog, plan, "type", "", budget); !errors.Is(err, ErrLimitExceeded) || docs != nil {
		t.Fatalf("exhausted shared budget: %v %v", docs, err)
	}
}

func TestIntrinsicDefinitionCurrentContextAcquisition(t *testing.T) {
	const query = `SELECT m.context_id FROM graph_preview_intrinsic_members m JOIN graph_preview_catalog c ON c.path=m.path AND c.revision=m.revision WHERE m.role=? AND c.backing=? AND c.allocation_state=? LIMIT 2`
	fault := errors.New("current-context driver fault")
	for _, mode := range []string{"one", "missing", "two", "invalid-id", "query-error", "row-error", "close-error", "extra-result", "budget"} {
		t.Run(mode, func(t *testing.T) {
			tx, mock := intrinsicProfileTransaction(t)
			limit := PreviewCurrentReadByteLimit
			if mode == "budget" {
				limit = 0
			} else {
				expected := mock.ExpectQuery(query).WithArgs("type", "intrinsic", "live")
				rows := sqlmock.NewRows([]string{"context_id"}).AddRow(strings.Repeat("c", 32))
				switch mode {
				case "missing":
					rows = sqlmock.NewRows([]string{"context_id"})
				case "two":
					rows.AddRow(strings.Repeat("d", 32))
				case "invalid-id":
					rows = sqlmock.NewRows([]string{"context_id"}).AddRow("not-an-exact-context")
				case "row-error":
					rows.RowError(0, fault)
				case "close-error":
					rows.CloseError(fault)
				}
				if mode == "query-error" {
					expected.WillReturnError(fault)
				} else if mode == "extra-result" {
					expected.WillReturnRows(rows, sqlmock.NewRows([]string{"context_id"}).AddRow(strings.Repeat("d", 32))).RowsWillBeClosed()
				} else {
					expected.WillReturnRows(rows).RowsWillBeClosed()
				}
			}
			budget, _ := newIntrinsicMetadataBudget(limit)
			id, err := currentIntrinsicContextInTx(context.Background(), tx, budget)
			if mode == "one" {
				if err != nil || id != strings.Repeat("c", 32) {
					t.Fatalf("one: %q %v", id, err)
				}
			} else {
				if err == nil || id != "" {
					t.Fatalf("failed acquisition leaked context: %q %v", id, err)
				}
				if strings.HasSuffix(mode, "error") && !errors.Is(err, fault) {
					t.Fatalf("lost driver cause: %v", err)
				}
				if mode == "budget" && !errors.Is(err, ErrLimitExceeded) {
					t.Fatalf("wrong budget diagnosis: %v", err)
				}
			}
		})
	}
}
