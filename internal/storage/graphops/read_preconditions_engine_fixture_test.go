//go:build cgo

package graphops

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage/authority"
)

// This disposable composition uses a synthetic witness value, never a minted
// witness, provider claim or lawful ledger. It qualifies only observation ->
// comparison -> body wiring on one embedded transaction, not serving authority.
func readPreconditionComposition(t *testing.T, parent context.Context, conn *sql.Conn, data string) {
	t.Helper()
	seq := uint64(1)
	baseline := readPreconditionFixture(t, parent, conn, &seq)
	key := strings.Repeat("b", 64)
	if baseline.scope.presence != observationPresent || baseline.scope.url != fixtureScope || baseline.scope.authorityID != strings.Repeat("a", 32) || baseline.scope.epoch != 2 ||
		baseline.ledger.recorded != (ledgerPoint{observationPresent, 1, strings.Repeat("f", 64)}) || baseline.ledger.tip != baseline.ledger.recorded ||
		baseline.lease.presence != observationPresent || baseline.lease.holder != key || baseline.lease.epoch != 2 || baseline.lease.grantedAt != "2026-09-17 00:00:00.000000" ||
		baseline.state.database != "graph_precondition_fixture" || baseline.state.branch != "main" {
		t.Fatalf("composition seed mismatch: %+v", baseline)
	}
	w, err := authority.NewWitness(authority.WitnessFields{
		ScopeURL: fixtureScope, AuthorityID: strings.Repeat("a", 32), Epoch: 2,
		LedgerSeq: 1, LedgerHash: strings.Repeat("f", 64),
		StateVersion: baseline.state.version, StateCommit: baseline.ledger.head,
		GrantedAt: "2026-09-17T00:00:00Z",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := expectedReadFacts{witness: &w, installationKey: key, witnessInstallationKey: key, database: "graph_precondition_fixture", branch: "main"}
	receipt := compositionReceipt(t, data)
	for _, method := range []string{"bead", "link", "incident"} {
		t.Run("composition/"+method, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(parent, 10*time.Second)
			defer cancel()
			tx := preconditionTx(t, ctx, conn)
			q := &preconditionQueryHook{tx: tx}
			got, err := observePreconditionsInTx(ctx, q, &seq)
			if err != nil {
				t.Fatal(err)
			}
			if err := checkReadFacts(ctx, want, got); err != nil {
				t.Fatal(err)
			}
			calls := compositionBody(t, ctx, q, method, receipt)
			if q.calls != calls {
				t.Fatalf("%s SELECT count = %d, want %d", method, q.calls, calls)
			}
			if err := tx.Rollback(); err != nil {
				t.Fatal(err)
			}
			t.Logf("%s: exact body and %d SELECTs; transaction closed", method, calls)
		})
	}
	compositionUnchanged(t, parent, conn, baseline)
	for _, tc := range []struct {
		name, setup string
		want        error
	}{
		{"no Scope", "DELETE FROM graph_scope", graph.ErrNoScope},
		{"missing prefix", "DELETE FROM graph_ledger_events WHERE seq = 1", graph.ErrStateRewound},
		{"foreign holder", "UPDATE graph_authority_lease SET holder_installation_key = REPEAT('c',64)", graph.ErrNotAuthority},
		{"expired", "UPDATE graph_authority_lease SET expires_at = NOW(6)", errReadLeaseExpired},
		{"short budget", "UPDATE graph_authority_lease SET expires_at = NOW(6) + INTERVAL 30 SECOND", errReadLeaseBudget},
		{"changed state", `UPDATE graph_beads SET properties = '{"text":"refusal control"}' WHERE path = 'beads/plan'`, graph.ErrStateChanged},
	} {
		t.Run("composition/refuse/"+tc.name, func(t *testing.T) {
			limit := 10 * time.Second
			if errors.Is(tc.want, errReadLeaseBudget) {
				// Leave scheduling headroom between the distinct expired and
				// insufficient-budget diagnoses, within the parent fixture bound.
				limit = time.Minute
			}
			ctx, cancel := context.WithTimeout(parent, limit)
			defer cancel()
			tx := preconditionTx(t, ctx, conn)
			// Test-owned fault setup precedes the measured four-SELECT read.
			stateExec(t, ctx, tx, tc.setup)
			q := &preconditionQueryHook{tx: tx}
			got, err := observePreconditionsInTx(ctx, q, &seq)
			if err != nil {
				t.Fatalf("observation failed before intended refusal: %v", err)
			}
			err = checkReadFacts(ctx, want, got)
			bodyCalled := false
			var row graph.BeadRecord
			if err == nil {
				bodyCalled = true
				row, err = readBeadInTx(ctx, q, fixtureScope, "beads/plan", fixtureLimits)
			}
			if !errors.Is(err, tc.want) || bodyCalled || q.calls != 4 || !reflect.DeepEqual(row, graph.BeadRecord{}) {
				t.Fatalf("refusal: err=%v want=%v body=%v SELECTs=%d row=%+v", err, tc.want, bodyCalled, q.calls, row)
			}
			if err := tx.Rollback(); err != nil {
				t.Fatal(err)
			}
			compositionUnchanged(t, parent, conn, baseline)
			t.Logf("%s: four observations, no body; rollback restored baseline", tc.name)
		})
	}
	if t.Failed() {
		t.Fatal("composition controls failed")
	}
	t.Log("GRAPH_READ_COMPOSITION_OK: private value comparison, exact bodies, SELECT budgets 7/5/5, six refusals before body; transactions closed")
}

func compositionReceipt(t *testing.T, data string) fixtureReceipt {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(data, "fixture-receipt.json"))
	if err != nil {
		t.Fatal(err)
	}
	var receipt fixtureReceipt
	if err := json.Unmarshal(raw, &receipt); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256([]byte(engineFixtureSQL))
	if receipt.SchemaSHA256 != hex.EncodeToString(sum[:]) || len(receipt.Beads) != 3 || len(receipt.Links) != 3 || receipt.Beads[0].Path != "beads/plan" || receipt.Links[0].Path != "links/plan-decision" || receipt.Links[2].Path != "links/external-plan" {
		t.Fatal("composition seed receipt mismatch")
	}
	return receipt
}

func compositionBody(t *testing.T, ctx context.Context, q queryer, method string, receipt fixtureReceipt) int {
	t.Helper()
	switch method {
	case "bead":
		r, err := readBeadInTx(ctx, q, fixtureScope, "beads/plan", fixtureLimits)
		if err != nil {
			t.Fatal(err)
		}
		_, attributed := r.Bead.Attribution()
		if r.Bead.Path() != "beads/plan" || r.Bead.TypeURL() != memoryType || r.Bead.Revision().String() != receipt.Beads[0].Revision || r.Bead.Properties().String() != `{"text":"precondition successor"}` || attributed ||
			len(r.OwnedLinks) != 1 || r.OwnedLinks[0].TypeURL != relationType || len(r.OwnedLinks[0].Links) != 1 {
			t.Fatalf("unexpected composed Bead: %+v", r)
		}
		compositionLink(t, r.OwnedLinks[0].Links[0], receipt.Links[0])
		return 7
	case "link":
		link, err := readLinkInTx(ctx, q, fixtureScope, "links/plan-decision", fixtureLimits)
		if err != nil {
			t.Fatal(err)
		}
		compositionLink(t, link, receipt.Links[0])
		return 5
	case "incident":
		links, err := readIncidentLinksInTx(ctx, q, fixtureScope, "beads/plan", graph.DirectionBoth, fixtureLimits)
		if err != nil || len(links) != 2 {
			t.Fatalf("composed incident links: %v, %v", links, err)
		}
		compositionLink(t, links[0], receipt.Links[2])
		compositionLink(t, links[1], receipt.Links[0])
		return 5
	default:
		t.Fatalf("unknown composition method %q", method)
		return 0
	}
}

func compositionLink(t *testing.T, link graph.Link, want fixtureRecord) {
	t.Helper()
	_, attributed := link.Attribution()
	if link.Path() != want.Path || link.TypeURL() != want.TypeURL || link.Properties().String() != want.Properties || link.Revision().String() != want.Revision || attributed ||
		link.Source().InScope() != (want.SourceURI == "") || link.Source().Path() != want.Source || link.Source().URI() != want.SourceURI || link.Source().Pin() != want.SourcePin ||
		!link.Target().InScope() || link.Target().Path() != want.Target || link.Target().Pin() != want.TargetPin {
		t.Fatalf("composed Link differs from seed receipt: %s", want.Path)
	}
}

func compositionUnchanged(t *testing.T, ctx context.Context, conn *sql.Conn, baseline preconditionObservations) {
	t.Helper()
	seq := uint64(1)
	got := readPreconditionFixture(t, ctx, conn, &seq)
	if !reflect.DeepEqual(preconditionStableFacts(got), preconditionStableFacts(baseline)) {
		t.Fatalf("composition failed to preserve stable baseline: %+v", got)
	}
}
