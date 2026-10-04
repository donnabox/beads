//go:build cgo

package graphstore

import (
	"errors"
	"reflect"
	"testing"

	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/types"
	publicops "github.com/steveyegge/beads/issueops"
)

func TestIssueHolderUnclaim(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			ctx, _, s, before, target, dependency := reopenFixture(t, backend)
			claimed, err := s.ClaimIssue(ctx, "beads/work", "rig.agent")
			if err != nil || !claimed.Changed {
				t.Fatalf("claim: %+v %v", claimed, err)
			}
			assertClaimLease(t, ctx, s, before.Properties.ID, "rig.agent")
			state := reopenState(t, ctx, s)
			foreign, err := s.UnclaimIssue(ctx, "beads/work", "foreign")
			if !errors.Is(err, storage.ErrNotOwner) || !reflect.DeepEqual(foreign, IssueMutationResult{}) || !reflect.DeepEqual(state, reopenState(t, ctx, s)) {
				t.Fatalf("foreign unclaim changed claim: %+v %v", foreign, err)
			}
			fault := errors.New("release must roll back")
			s.afterWrite = func(stage string) error {
				if stage == "issue-unclaim" {
					return fault
				}
				return nil
			}
			failed, err := s.UnclaimIssue(ctx, "beads/work", "rig.agent")
			s.afterWrite = nil
			if !errors.Is(err, fault) || !reflect.DeepEqual(failed, IssueMutationResult{}) || !reflect.DeepEqual(state, reopenState(t, ctx, s)) {
				t.Fatalf("unclaim rollback: %+v %v", failed, err)
			}
			released, err := s.UnclaimIssue(ctx, "beads/work", "rig_agent")
			if err != nil || !released.Changed {
				t.Fatalf("holder alias release: %+v %v", released, err)
			}
			p := released.Issue.Properties
			if p == nil || p.Status != types.StatusOpen || p.Assignee != "" || p.StartedAt != nil || p.LeaseExpiresAt != nil || p.HeartbeatAt != nil || released.Issue.Revision == claimed.Issue.Revision || released.Issue.Version != released.Issue.Revision || released.Issue.Attribution.Actor != "rig_agent" {
				t.Fatalf("incomplete release: %+v", released.Issue)
			}
			assertAssigneeLeaseCount(t, ctx, s, before.Properties.ID, 0)
			assertIssueEditCounts(t, ctx, s, before.Properties.ID, 4)
			for _, record := range []IssueRecord{before, claimed.Issue, released.Issue} {
				assertIssueEditVersion(t, ctx, s, "beads/work", record)
			}
			state = reopenState(t, ctx, s)
			again, err := s.UnclaimIssue(ctx, "beads/work", "rig.agent")
			if !errors.Is(err, publicops.ErrNotReleasable) || !reflect.DeepEqual(again, IssueMutationResult{}) || !reflect.DeepEqual(state, reopenState(t, ctx, s)) {
				t.Fatalf("repeat release changed state: %+v %v", again, err)
			}
			if got, err := s.ShowIssue(ctx, "beads/prereq"); err != nil || !reflect.DeepEqual(got, target) {
				t.Fatalf("target changed: %+v %v", got, err)
			}
			if got, err := s.ShowLink(ctx, "links/block"); err != nil || !reflect.DeepEqual(got, dependency) {
				t.Fatalf("dependency changed: %+v %v", got, err)
			}
			assigned, err := s.UpdateIssue(ctx, UpdateIssueRequest{Path: "beads/work", Actor: "assigner", ExpectedRevision: released.Issue.Revision, Assignee: issueEditString("rig.agent")})
			if err != nil || !assigned.Changed || assigned.Issue.Properties.Status != types.StatusOpen {
				t.Fatalf("open assignment: %+v %v", assigned, err)
			}
			state = reopenState(t, ctx, s)
			openRelease, err := s.UnclaimIssue(ctx, "beads/work", "rig.agent")
			if !errors.Is(err, publicops.ErrNotReleasable) || !reflect.DeepEqual(openRelease, IssueMutationResult{}) || !reflect.DeepEqual(state, reopenState(t, ctx, s)) {
				t.Fatalf("unclaim erased open assignment: %+v %v", openRelease, err)
			}
			cleared, err := s.UpdateIssue(ctx, UpdateIssueRequest{Path: "beads/work", Actor: "assigner", ExpectedRevision: assigned.Issue.Revision, Assignee: issueEditString("")})
			if err != nil || !cleared.Changed {
				t.Fatalf("clear assignment before reclaim: %+v %v", cleared, err)
			}
			reclaimed, err := s.ClaimIssue(ctx, "beads/work", "next.agent")
			if err != nil || !reclaimed.Changed || reclaimed.Issue.Properties.Assignee != "next.agent" {
				t.Fatalf("reclaim after release: %+v %v", reclaimed, err)
			}
			assertClaimLease(t, ctx, s, before.Properties.ID, "next.agent")
		})
	}
}
