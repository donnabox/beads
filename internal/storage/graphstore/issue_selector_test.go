//go:build cgo

package graphstore

import (
	"database/sql"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/types"
)

func TestIssueNativeSelectorHeartbeatStateAndReclaim(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			ctx, options := issueExperimentOptions(t, backend)
			s, err := OpenExisting(ctx, options)
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()
			created, err := s.CreateIssue(ctx, "beads/custom-path", plainIssue("Custom"))
			if err != nil {
				t.Fatal(err)
			}
			if created.Properties.ID == "custom-path" {
				t.Fatal("fixture needs a native ID distinct from canonical path")
			}
			claimed, err := s.ClaimIssue(ctx, "beads/custom-path", "holder")
			if err != nil || !claimed.Changed {
				t.Fatalf("claim: %+v %v", claimed, err)
			}
			selector := IssueSelector{Path: "beads/" + created.Properties.ID, NativeID: created.Properties.ID}
			shown, err := s.ShowIssueSelected(ctx, selector)
			if err != nil || !reflect.DeepEqual(shown, claimed.Issue) {
				t.Fatalf("selected read: %+v %v", shown, err)
			}
			renewed, err := s.HeartbeatIssueSelected(ctx, selector, "holder")
			if err != nil || renewed.Revision != claimed.Issue.Revision || renewed.Properties.LeaseExpiresAt == nil {
				t.Fatalf("selected heartbeat changed retained History: %+v %v", renewed, err)
			}
			assertIssueEditCounts(t, ctx, s, created.Properties.ID, 2)
			assertIssueEditVersion(t, ctx, s, "beads/custom-path", claimed.Issue)
			if err := s.withTx(ctx, true, func(tx *sql.Tx) error {
				_, err := tx.ExecContext(ctx, `UPDATE leases SET lease_expires_at=? WHERE issue_id=?`, time.Now().UTC().Add(-time.Hour), created.Properties.ID)
				return err
			}); err != nil {
				t.Fatal(err)
			}
			reclaimed, err := s.ReclaimExpiredIssueLeasesSelected(ctx, 0, types.ReclaimFilter{}, []IssueSelector{selector}, "reaper")
			if err != nil || len(reclaimed) != 1 || reclaimed[0].ID != created.Properties.ID {
				t.Fatalf("selected recovery: %+v %v", reclaimed, err)
			}
			after, err := s.ShowIssue(ctx, "beads/custom-path")
			if err != nil || after.Revision == renewed.Revision || after.Properties.Status != types.StatusOpen {
				t.Fatalf("selected recovery lost Issue writer: %+v %v", after, err)
			}
			assertIssueEditCounts(t, ctx, s, created.Properties.ID, 3)
			stable := workflowState(t, ctx, s)
			if _, err := s.HeartbeatIssueSelected(ctx, selector, "holder"); !errors.Is(err, storage.ErrNotClaimable) || !reflect.DeepEqual(stable, workflowState(t, ctx, s)) {
				t.Fatalf("released holder heartbeat changed state: %v", err)
			}
		})
	}
}
