//go:build cgo

package graphstore

import (
	"bytes"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/types"
)

func TestNativeMemoryDeleteRecoveryKeepsExactPin(t *testing.T) {
	ctx, _, s := intrinsicHardeningStore(t)
	before, err := s.Create(ctx, CreateRequest{Path: "beads/recovery", Title: "Original", Actor: "writer"})
	if err != nil {
		t.Fatal(err)
	}
	first, ok := before.ResourceDocument()
	if !ok {
		t.Fatal("missing inception facet")
	}
	state := workflowState(t, ctx, s)
	preview, err := s.DeleteMemory(ctx, MemoryDeleteRequest{Path: "beads/recovery", Preview: true})
	if err != nil || !preview.Preview || preview.Deleted || preview.Terminal != nil || !reflect.DeepEqual(state, workflowState(t, ctx, s)) {
		t.Fatalf("preview staged effects: %+v %v", preview, err)
	}
	if d, ok := preview.Memory.ResourceDocument(); !ok || !bytes.Equal(d.NativeJSON(), first.NativeJSON()) {
		t.Fatal("preview lost checked full record")
	}
	fault := errors.New("delete retained fault")
	s.afterWrite = func(stage string) error {
		if stage == "memory-delete-retained" {
			return fault
		}
		return nil
	}
	failed, err := s.DeleteMemory(ctx, MemoryDeleteRequest{Path: "beads/recovery", ExpectedRevision: before.Revision, Actor: "writer"})
	s.afterWrite = nil
	if !errors.Is(err, fault) || !reflect.ValueOf(failed).IsZero() || !reflect.DeepEqual(state, workflowState(t, ctx, s)) {
		t.Fatalf("failed deletion leaked: %+v %v", failed, err)
	}
	deleted, err := s.DeleteMemory(ctx, MemoryDeleteRequest{Path: "beads/recovery", ExpectedRevision: before.Revision, Actor: "writer"})
	if err != nil || !deleted.Deleted || deleted.Terminal == nil {
		t.Fatalf("delete: %+v %v", deleted, err)
	}
	recovered, err := s.PatchMemory(ctx, MemoryPatchRequest{Path: "beads/recovery", FromRevision: before.Revision, ExpectedRevision: deleted.Terminal.Revision, Actor: "writer"})
	if err != nil {
		t.Fatal(err)
	}
	document, ok := recovered.Memory.ResourceDocument()
	if !ok || document.Type() != first.Type() || document.ID() != first.ID() || document.Revision() == first.Revision() {
		t.Fatalf("recovery selected a replacement Type pin: %+v", document)
	}
}

func TestNativeReparentReturnsAppliedLinkFacet(t *testing.T) {
	ctx, _, s := intrinsicHardeningStore(t)
	for _, path := range []string{"beads/child", "beads/parent"} {
		if _, err := s.CreateIssue(ctx, path, plainIssue(path)); err != nil {
			t.Fatal(err)
		}
	}
	child, err := s.ShowIssue(ctx, "beads/child")
	if err != nil {
		t.Fatal(err)
	}
	state := workflowState(t, ctx, s)
	fault := errors.New("final validation fault")
	s.afterWrite = func(stage string) error {
		if stage == "native-final-validation" {
			return fault
		}
		return nil
	}
	failed, err := s.ReparentIssue(ctx, IssueReparentRequest{Path: "beads/child", ParentPath: "beads/parent", ExpectedRevision: child.Revision, Actor: "writer"})
	s.afterWrite = nil
	if !errors.Is(err, fault) || !reflect.ValueOf(failed).IsZero() || !reflect.DeepEqual(state, workflowState(t, ctx, s)) {
		t.Fatalf("failed reparent leaked: %+v %v", failed, err)
	}
	result, err := s.ReparentIssue(ctx, IssueReparentRequest{Path: "beads/child", ParentPath: "beads/parent", ExpectedRevision: child.Revision, Actor: "writer"})
	if err != nil || result.Added == nil {
		t.Fatalf("reparent: %+v %v", result, err)
	}
	link, ok := result.Added.ResourceDocument()
	if !ok {
		t.Fatal("missing same-Tx Added facet")
	}
	issue, ok := result.Issue.ResourceDocument()
	if !ok {
		t.Fatal("missing same-Tx Issue facet")
	}
	if link.ID() != result.Added.ID || link.Revision() != result.Added.Revision || issue.Revision() != result.Issue.Revision {
		t.Fatal("applied identities differ")
	}
	current, err := s.GetResourceDocument(ctx, link.ID(), "")
	if err != nil || !bytes.Equal(current.NativeJSON(), link.NativeJSON()) {
		t.Fatalf("applied Link differs from committed full body: %v", err)
	}
}

func TestNativeInformationalRecoveryKeepsExactPin(t *testing.T) {
	ctx, _, s := intrinsicHardeningStore(t)
	for _, path := range []string{"beads/source", "beads/target"} {
		if _, err := s.Create(ctx, CreateRequest{Path: path, Title: path, Actor: "writer"}); err != nil {
			t.Fatal(err)
		}
	}
	created, err := s.AddInformationalLink(ctx, LinkCreateRequest{Path: "links/recovery", SourcePath: "beads/source", TargetPath: "beads/target", Actor: "writer"})
	if err != nil {
		t.Fatal(err)
	}
	initial, ok := created.Link.ResourceDocument()
	if !ok {
		t.Fatal("initial Link lacks applied facet")
	}
	deleted, err := s.Unlink(ctx, LinkDeleteRequest{Path: "links/recovery", ExpectedRevision: created.Link.Revision, Actor: "writer"})
	if err != nil {
		t.Fatal(err)
	}
	state := workflowState(t, ctx, s)
	fault := errors.New("recovery finalization")
	s.afterWrite = func(stage string) error {
		if stage == "native-final-validation" {
			return fault
		}
		return nil
	}
	failed, err := s.UpdateLink(ctx, LinkUpdateRequest{Path: "links/recovery", FromRevision: created.Link.Revision, ExpectedRevision: deleted.Link.Revision, Actor: "writer"})
	s.afterWrite = nil
	if !errors.Is(err, fault) || !reflect.ValueOf(failed).IsZero() || !reflect.DeepEqual(state, workflowState(t, ctx, s)) {
		t.Fatalf("recovery failed without rollback: %+v %v", failed, err)
	}
	recovered, err := s.UpdateLink(ctx, LinkUpdateRequest{Path: "links/recovery", FromRevision: created.Link.Revision, ExpectedRevision: deleted.Link.Revision, Actor: "writer"})
	if err != nil {
		t.Fatal(err)
	}
	document, ok := recovered.Link.ResourceDocument()
	if !ok || document.Type() != initial.Type() || document.ID() != initial.ID() || document.Revision() == initial.Revision() {
		t.Fatal("recovery failed exact pin/identity correspondence")
	}
}

func TestNativeLeaseFinalizationSuppressesResults(t *testing.T) {
	for _, name := range []string{"heartbeat-fault", "heartbeat-unknown", "reclaim-fault", "reclaim-unknown"} {
		t.Run(name, func(t *testing.T) {
			ctx, _, s := intrinsicHardeningStore(t)
			created, err := s.CreateIssue(ctx, "beads/leased", plainIssue("Lease"))
			if err != nil {
				t.Fatal(err)
			}
			if _, err = s.ClaimIssue(ctx, "beads/leased", "holder"); err != nil {
				t.Fatal(err)
			}
			if name == "reclaim-fault" || name == "reclaim-unknown" {
				if _, err = s.db.ExecContext(ctx, `UPDATE leases SET lease_expires_at=? WHERE issue_id=?`, time.Now().UTC().Add(-time.Hour), created.Properties.ID); err != nil {
					t.Fatal(err)
				}
			}
			state := workflowState(t, ctx, s)
			fault := errors.New("postcallback failure")
			var injected error = fault
			if name == "heartbeat-unknown" || name == "reclaim-unknown" {
				injected = errors.Join(ErrOutcomeUnknown, fault)
			}
			reached := false
			s.afterWrite = func(stage string) error {
				if stage == "native-final-validation" {
					reached = true
					return injected
				}
				return nil
			}
			if name == "heartbeat-fault" || name == "heartbeat-unknown" {
				result, e := s.HeartbeatIssue(ctx, "beads/leased", "holder")
				err = e
				if !reflect.ValueOf(result).IsZero() {
					t.Fatalf("failed heartbeat exposed applied result: %+v", result)
				}
			} else {
				result, e := s.ReclaimExpiredIssueLeases(ctx, 0, types.ReclaimFilter{IDs: []string{created.Properties.ID}}, "reaper")
				err = e
				if result != nil {
					t.Fatalf("failed reclaim exposed provisional list: %+v", result)
				}
			}
			s.afterWrite = nil
			if !reached || !errors.Is(err, fault) || ((name == "heartbeat-unknown" || name == "reclaim-unknown") && !errors.Is(err, ErrOutcomeUnknown)) || !reflect.DeepEqual(state, workflowState(t, ctx, s)) {
				t.Fatalf("failure/cause/rollback mismatch: %v", err)
			}
		})
	}
}

func TestNativeClaimReadyPreservesLazyWake(t *testing.T) {
	for _, name := range []string{"due", "future", "wake-rollback"} {
		t.Run(name, func(t *testing.T) {
			ctx, _, s := intrinsicHardeningStore(t)
			due := time.Now().UTC().Add(-time.Hour)
			if name == "future" {
				due = time.Now().UTC().Add(time.Hour)
			}
			request := plainIssue("Deferred")
			request.Issue.Status = types.StatusDeferred
			request.Issue.DeferUntil = &due
			created, err := s.CreateIssue(ctx, "beads/deferred", request)
			if err != nil {
				t.Fatal(err)
			}
			state := workflowState(t, ctx, s)
			reached := false
			if name == "wake-rollback" {
				s.afterWrite = func(stage string) error {
					if stage == "native-final-validation" && !reached {
						reached = true
						return errors.New("wake rollback")
					}
					return nil
				}
			}
			claimed, err := s.ClaimReadyIssue(ctx, "holder", types.WorkFilter{})
			s.afterWrite = nil
			if err != nil {
				t.Fatal(err)
			}
			if name == "due" {
				if claimed == nil || claimed.Issue.Properties.Status != types.StatusInProgress || claimed.Issue.Properties.DeferUntil != nil || claimed.Issue.Revision == created.Revision {
					t.Fatalf("due defer not claimed: %+v", claimed)
				}
				if _, ok := claimed.Issue.ResourceDocument(); !ok {
					t.Fatal("claim lost full owned capture")
				}
				var events int
				if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM events WHERE issue_id=? AND actor='bd-defer-wake'`, created.Properties.ID).Scan(&events); err != nil || events != 1 {
					t.Fatalf("wake history count %d: %v", events, err)
				}
			} else if claimed != nil || (name == "wake-rollback" && !reached) || !reflect.DeepEqual(state, workflowState(t, ctx, s)) {
				t.Fatalf("future/rolledback wake changed state: %+v", claimed)
			}
		})
	}
}
