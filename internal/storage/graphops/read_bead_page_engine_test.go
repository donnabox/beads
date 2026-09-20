//go:build cgo

package graphops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"

	graph "github.com/steveyegge/beads/graphops"
)

// Private projection-only setup; this neither installs lawful graph state nor
// grants authority. All added rows and corruptions roll back before the digest
// check. Embedded and managed workers run the same bound statements.
func verifyEngineBeadPageControls(t *testing.T, ctx context.Context, db fixtureDB) {
	t.Helper()
	before := fixtureTablesDigest(t, ctx, db)
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()
	scope := scopeFromFixtureTx(t, ctx, tx)
	types := []string{memoryType + "/bead-page/A", memoryType + "/bead-page/a", memoryType + "/bead-page/none"}
	other := relationType + "/bead-page/other"
	explicit := beadPageDecl(t, relationType, 2)
	wild := beadPageDecl(t, "*", 3)
	descriptors := []graph.TypeDescriptor{beadPageDescriptor(t, types[0], explicit), beadPageDescriptor(t, types[1], explicit, wild), beadPageDescriptor(t, types[2])}
	endpoint, err := graph.NewEndpointConstraint(nil, graph.ExternalOpaque)
	if err != nil {
		t.Fatal(err)
	}
	relation, err := graph.NewTypeDescriptor(graph.TypeDescriptorSpec{ID: other, Name: "Other", Describes: graph.KindLink, Source: &endpoint, Target: &endpoint})
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range append(descriptors, relation) {
		if _, err := tx.ExecContext(ctx, "INSERT INTO graph_type_descriptors (url,descriptor,fingerprint,installed_seq,installed_at,last_authority_id,last_epoch) VALUES (?,?,?,1,'2026-09-20 00:00:00',?,1)", d.ID(), d.CanonicalJSON(), d.Fingerprint(), strings.Repeat("a", 32)); err != nil {
			t.Fatal(err)
		}
	}
	paths := []string{"beads/zz-page/A", "beads/zz-page/a", "beads/zz-page/caf%C3%A9", "beads/zz-page/nested/x"}
	typeFor := []string{types[0], types[1], types[2], types[0]}
	expected := map[string]graph.BeadRecord{}
	for i, path := range paths {
		revision := graph.MintRevision()
		properties, err := graph.NewProperties([]byte(`{"page":true}`))
		if err != nil {
			t.Fatal(err)
		}
		bead, err := graph.NewBead(graph.BeadSpec{Path: path, TypeURL: typeFor[i], Revision: revision, Properties: properties})
		if err != nil {
			t.Fatal(err)
		}
		record := graph.BeadRecord{Bead: bead}
		if typeFor[i] != types[2] {
			record.OwnedLinks = []graph.OwnedLinkGroup{{TypeURL: relationType}}
		}
		expected[path] = record
		if _, err := tx.ExecContext(ctx, "INSERT INTO graph_beads (path,type_url,revision,properties,last_authority_id,last_epoch,created_at,updated_at) VALUES (?,?,?,?,?,1,'2026-09-20 00:00:00','2026-09-20 00:00:00')", path, typeFor[i], revision.String(), properties.Bytes(), strings.Repeat("a", 32)); err != nil {
			t.Fatal(err)
		}
		beadPageFixtureAllocation(t, ctx, tx, path, "bead")
	}
	for _, f := range []struct{ path, source, typ, target string }{{"links/zz-page/01", paths[1], relationType, "urn:page:external"}, {"links/zz-page/02", paths[1], other, paths[0]}, {"links/zz-page/03", paths[3], relationType, paths[1]}} {
		source, err := graph.NewInScopeRef(f.source, "source opaque")
		if err != nil {
			t.Fatal(err)
		}
		target, err := graph.ParseRef(scope, f.target, "target opaque")
		if err != nil {
			t.Fatal(err)
		}
		properties, err := graph.NewProperties([]byte("{}"))
		if err != nil {
			t.Fatal(err)
		}
		link, err := graph.NewLink(graph.LinkSpec{Path: f.path, TypeURL: f.typ, Revision: graph.MintRevision(), Properties: properties, Source: source, Target: target})
		if err != nil {
			t.Fatal(err)
		}
		sk, sp, su := pageFixtureEndpoint(source)
		tk, tp, tu := pageFixtureEndpoint(target)
		if _, err := tx.ExecContext(ctx, "INSERT INTO graph_links (path,type_url,revision,properties,source_kind,source_path,source_url,source_pin,target_kind,target_path,target_url,target_pin,last_authority_id,last_epoch,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,1,'2026-09-20 00:00:00','2026-09-20 00:00:00')", f.path, f.typ, link.Revision().String(), properties.Bytes(), sk, sp, su, source.Pin(), tk, tp, tu, target.Pin(), strings.Repeat("a", 32)); err != nil {
			t.Fatal(err)
		}
		beadPageFixtureAllocation(t, ctx, tx, f.path, "link")
		record := expected[f.source]
		if f.typ == relationType {
			record.OwnedLinks[0].Links = append(record.OwnedLinks[0].Links, link)
		} else {
			record.OwnedLinks = append(record.OwnedLinks, graph.OwnedLinkGroup{TypeURL: f.typ, Links: []graph.Link{link}})
		}
		expected[f.source] = record
	}
	// Inputs and expected groups are independently enumerated, not obtained by
	// calling singleton bodies or reconstructing the batch query result.
	for _, tc := range []struct {
		name, typ string
		paths     []string
	}{{"all", "", paths}, {"explicit", types[0], []string{paths[0], paths[3]}}, {"wildcard", types[1], []string{paths[1]}}, {"none", types[2], []string{paths[2]}}, {"missing", memoryType + "/not-installed", nil}, {"link-kind", other, nil}} {
		after := "beads/zz-page"
		offset := 0
		for step := 0; step <= len(paths)+1; step++ {
			q := &preconditionQueryHook{tx: tx}
			page, err := readBeadPageInTx(ctx, q, scope, tc.typ, pageWindow{afterPath: after, limit: 1}, fixtureLimits)
			if err != nil {
				t.Fatalf("Bead page %s: %v", tc.name, err)
			}
			count := min(1, len(tc.paths)-offset)
			if len(page.items) != count || page.hasMore != (offset+count < len(tc.paths)) || q.calls > 3 {
				t.Fatalf("Bead page %s offset=%d calls=%d result=%+v", tc.name, offset, q.calls, page)
			}
			if count == 0 {
				requireZeroBeadPage(t, page)
				if q.calls != 1 {
					t.Fatal("empty page queried descriptors")
				}
				break
			}
			wanted := expected[tc.paths[offset]]
			beadPageFixtureEqual(t, page.items[0], wanted)
			if page.lastPath != wanted.Bead.Path() {
				t.Fatal("wrong continuation")
			}
			expectedCalls := 2
			if len(wanted.OwnedLinks) > 0 {
				expectedCalls = 3
			}
			if q.calls != expectedCalls {
				t.Fatalf("Bead page %s calls=%d want=%d", tc.name, q.calls, expectedCalls)
			}
			offset += count
			after = page.lastPath
			if step == len(paths)+1 {
				t.Fatal("unbounded Bead traversal")
			}
		}
	}
	// Execute a nonempty batch spanning explicit, wildcard, and no-owns
	// descriptors; limit-one traversal alone cannot qualify multi-owner SQL.
	batchQuery := &preconditionQueryHook{tx: tx}
	batch, err := readBeadPageInTx(ctx, batchQuery, scope, "", pageWindow{afterPath: "beads/zz-page", limit: 4}, fixtureLimits)
	if err != nil || len(batch.items) != len(paths) || batch.hasMore || batch.lastPath != paths[len(paths)-1] || batchQuery.calls != 3 {
		t.Fatalf("mixed-owner Bead page calls=%d result=%+v error=%v", batchQuery.calls, batch, err)
	}
	for i, path := range paths {
		beadPageFixtureEqual(t, batch.items[i], expected[path])
	}
	// Actual same-transaction malformed lookahead descriptor refusal. Restore
	// exact bytes before further checks; rollback remains the final cleanup.
	bad := descriptors[1]
	if _, err := tx.ExecContext(ctx, "UPDATE graph_type_descriptors SET descriptor=? WHERE url=?", []byte("null"), bad.ID()); err != nil {
		t.Fatal(err)
	}
	refused, err := readBeadPageInTx(ctx, tx, scope, "", pageWindow{afterPath: "beads/zz-page", limit: 1}, fixtureLimits)
	requireZeroBeadPage(t, refused)
	if !errors.Is(err, errCorrupt) {
		t.Fatalf("lookahead descriptor error=%v", err)
	}
	if _, err := tx.ExecContext(ctx, "UPDATE graph_type_descriptors SET descriptor=? WHERE url=?", bad.CanonicalJSON(), bad.ID()); err != nil {
		t.Fatal(err)
	}
	tight := fixtureLimits
	tight.rows = 4 // two Beads + two descriptors leaves no owned capacity.
	refused, err = readBeadPageInTx(ctx, tx, scope, "", pageWindow{afterPath: paths[0], limit: 1}, tight)
	requireZeroBeadPage(t, refused)
	if !errors.Is(err, errBudget) {
		t.Fatalf("complete expansion budget error=%v", err)
	}
	beadPageMaximumPreparedShapes(t, ctx, tx)
	if _, ok := db.(*sql.DB); ok {
		for _, typ := range []string{"", types[0]} {
			for _, after := range []string{"", paths[0]} {
				query, args := beadPageQuery(typ, pageWindow{afterPath: after, limit: 1}, fixtureLimits.valueBytes)
				capturePageExplain(t, ctx, tx, "Beads/"+typ, after, query, args)
			}
		}
		query, args := beadDescriptorBatchQuery(types, fixtureLimits.valueBytes)
		capturePageExplain(t, ctx, tx, "Beads/descriptors", "", query, args)
		budget, err := budgetFor(scope, fixtureLimits)
		if err != nil {
			t.Fatal(err)
		}
		owners, err := makeBeadPageOwners([]graph.Bead{expected[paths[0]].Bead, expected[paths[1]].Bead}, map[string]graph.TypeDescriptor{types[0]: descriptors[0], types[1]: descriptors[1]}, &beadPageBudget{readBudget: budget, rowsLeft: fixtureLimits.rows, groupsLeft: fixtureLimits.rows})
		if err != nil {
			t.Fatal(err)
		}
		for _, subset := range [][]*beadPageOwner{owners[:1], owners[1:], owners} {
			query, args := beadOwnedBatchQuery(subset, fixtureLimits.valueBytes, 10)
			capturePageExplain(t, ctx, tx, "Beads/owned", "", query, args)
		}
	} else {
		fmt.Println("GRAPH_BEAD_PAGE_PLAN_DIAGNOSTIC_OMITTED managed prepared EXPLAIN unsupported; bound semantic/max-shape queries remain mandatory")
	}
	// Persist two additional rows for the existing nested owner: all three
	// rows fit the transfer budget, but exceed its explicit descriptor Max=2.
	// They remain projection-only negative controls and roll back below.
	for _, path := range []string{"links/zz-page/04", "links/zz-page/05"} {
		if _, err := tx.ExecContext(ctx, "INSERT INTO graph_links (path,type_url,revision,properties,source_kind,source_path,source_pin,target_kind,target_path,target_pin,last_authority_id,last_epoch,created_at,updated_at) VALUES (?,?,?,'{}','in',?,'source opaque','in',?,'target opaque',?,1,'2026-09-20 00:00:00','2026-09-20 00:00:00')", path, relationType, graph.MintRevision().String(), paths[3], paths[1], strings.Repeat("a", 32)); err != nil {
			t.Fatal(err)
		}
		beadPageFixtureAllocation(t, ctx, tx, path, "link")
	}
	refused, err = readBeadPageInTx(ctx, tx, scope, "", pageWindow{afterPath: paths[2], limit: 1}, fixtureLimits)
	requireZeroBeadPage(t, refused)
	if !errors.Is(err, errCorrupt) || errors.Is(err, errBudget) {
		t.Fatalf("persisted descriptor Max error=%v", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if after := fixtureTablesDigest(t, ctx, db); after != before {
		t.Fatal("Bead page controls changed table digest")
	}
	if t.Failed() {
		return
	}
	fmt.Println("GRAPH_BEAD_PAGE_CONTROLS_OK: batched complete records, lookahead, exact Type, max IN/OR preparation, rollback digest")
}

func beadPageFixtureAllocation(t *testing.T, ctx context.Context, tx *sql.Tx, path, kind string) {
	t.Helper()
	if _, err := tx.ExecContext(ctx, "INSERT INTO graph_allocations (path,resource_kind,state,birth_seq,birth_authority_id,birth_authority_epoch,last_authority_id,last_authority_epoch) VALUES (?,?,'live',1,?,1,?,1)", path, kind, strings.Repeat("a", 32), strings.Repeat("a", 32)); err != nil {
		t.Fatal(err)
	}
}
func beadPageFixtureEqual(t *testing.T, got, want graph.BeadRecord) {
	t.Helper()
	if got.Bead.Path() != want.Bead.Path() || got.Bead.TypeURL() != want.Bead.TypeURL() || got.Bead.Revision() != want.Bead.Revision() || got.Bead.Properties().String() != want.Bead.Properties().String() || len(got.OwnedLinks) != len(want.OwnedLinks) {
		t.Fatalf("altered Bead record: got=%+v want=%+v", got, want)
	}
	for i, g := range got.OwnedLinks {
		w := want.OwnedLinks[i]
		if g.TypeURL != w.TypeURL || len(g.Links) != len(w.Links) {
			t.Fatal("owned groups differ")
		}
		for j, l := range g.Links {
			v := w.Links[j]
			if l.Path() != v.Path() || l.TypeURL() != v.TypeURL() || l.Revision() != v.Revision() || l.Properties().String() != v.Properties().String() || !l.Source().Equal(v.Source()) || !l.Target().Equal(v.Target()) {
				t.Fatal("owned Link differs")
			}
		}
	}
}

// Exercise the conservative syntactic ceiling on the actual prepared path,
// independently of body capacity. These absent identities create no rows. This
// qualifies parser/protocol acceptance, not a 1024-record public page capacity.
func beadPageMaximumPreparedShapes(t *testing.T, ctx context.Context, tx *sql.Tx) {
	t.Helper()
	ids := make([]string, 1024)
	owners := make([]*beadPageOwner, 1024)
	for i := range ids {
		ids[i] = fmt.Sprintf("https://graph.example/absent-page-type/%04d", i)
		bead, err := decodeBead(beadPageRow(fmt.Sprintf("beads/absent-page/%04d", i), memoryType))
		if err != nil {
			t.Fatal(err)
		}
		owners[i] = &beadPageOwner{record: graph.BeadRecord{Bead: bead}, owns: []graph.OwnedLinkDecl{beadPageDecl(t, ids[i], 1)}}
	}
	descriptorQuery, descriptorArgs := beadDescriptorBatchQuery(ids, fixtureLimits.valueBytes)
	ownedQuery, ownedArgs := beadOwnedBatchQuery(owners, fixtureLimits.valueBytes, 1)
	if len(descriptorArgs) != 1026 || len(ownedArgs) != 2050 {
		t.Fatal("maximum query shape changed")
	}
	for _, tc := range []struct {
		name, query string
		args        []any
	}{{"IN", descriptorQuery, descriptorArgs}, {"OR", ownedQuery, ownedArgs}} {
		rows, err := tx.QueryContext(ctx, tc.query, tc.args...)
		if err != nil {
			t.Fatalf("maximum %s %d binds: %v", tc.name, len(tc.args), err)
		}
		present := rows.Next()
		rowErr := rows.Err()
		closeErr := rows.Close()
		if present || errors.Join(rowErr, closeErr, ctx.Err()) != nil {
			t.Fatalf("maximum %s unexpected row/error: %t %v %v", tc.name, present, rowErr, closeErr)
		}
		t.Logf("Bead page maximum prepared %s shape accepted %d bind operands", tc.name, len(tc.args))
	}
}
