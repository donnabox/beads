package bdpstore

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/steveyegge/beads/internal/bdp"
	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/issueops"
)

// This file maps the beads issueops LIFECYCLE state machine onto BDP Core `call`
// of issue-pack callables — Lydia's ruling (2026-11-26), confirmed against the
// GENERATED catalog specs/generated/callables/issue.callables.json (vendored at
// third_party/bdp/spec/issue.callables.json). Dedicated callables are REQUIRED:
// generic update cannot express claim/heartbeat, because the lease state is
// serviceManaged (update-forbidden) and claim/heartbeat/close each carry a
// declared domain refusal a generic update cannot carry — the same reason patch
// was ruled out.
//
// Pinned paths (each a `call` callable returning rep<Issue> with a declared
// error). The acting principal rides in provenance, NEVER as a parameter.
//
// STILL MOCK-ONLY: cleanroom-v023 offers no `call`, so none of this is
// live-validated — proving it needs Track B's call-offering, issue-pack server.
const (
	opClaim     = "issue/claim"     // claimIssue   — ClaimError
	opClose     = "issue/close"     // closeIssue   — CloseError
	opReopen    = "issue/reopen"    // reopenIssue
	opHeartbeat = "issue/heartbeat" // heartbeatIssue — HeartbeatError
)

// arg builds one named call argument, JSON-encoding v.
func arg(name string, v any) bdp.Argument {
	b, _ := json.Marshal(v)
	return bdp.Argument{Name: name, Value: b}
}

// actorProvenance carries the acting principal in the request provenance, where
// the issue-pack callables read the actor from (never a parameter). The
// principal shape is provisional.
func actorProvenance(actor string) *bdp.Provenance {
	if actor == "" {
		return nil
	}
	p, _ := json.Marshal(map[string]string{"id": actor})
	return &bdp.Provenance{Principals: []bdp.Principal{bdp.Principal(p)}}
}

// callOp invokes op with the actor in provenance and args, returning the settled
// Outcome (a pre-dispatch fault / transport failure comes back as the error).
func (s *Store) callOp(ctx context.Context, op, actor string, args ...bdp.Argument) (bdp.Outcome, error) {
	reply, err := s.scope.Call(ctx, bdp.ResourceID{ID: op}, bdp.Request{
		Arguments:  args,
		Provenance: actorProvenance(actor),
	})
	if err != nil {
		return bdp.Outcome{}, err
	}
	return reply.Outcome, nil
}

// issueFromOutcome decodes a callable's rep<Issue> success value as an Issue.
func issueFromOutcome(o bdp.Outcome) (*types.Issue, error) {
	var iss types.Issue
	if len(o.Value) > 0 {
		if err := json.Unmarshal(o.Value, &iss); err != nil {
			return nil, fmt.Errorf("bdpstore: decode lifecycle result: %w", err)
		}
	}
	return &iss, nil
}

// failureCode pulls the provisional `code` discriminator from an Outcome.failure
// (the declared error's internal schema is not pinned in the catalog yet).
func failureCode(o bdp.Outcome) string {
	var e struct {
		Code string `json:"code"`
	}
	_ = json.Unmarshal(o.Error, &e)
	return e.Code
}

// Close closes an issue via `call issue/close` (params: issue, closeReason).
func (l bdpLifecycle) Close(ctx context.Context, req issueops.CloseRequest) (issueops.CloseResult, error) {
	args := []bdp.Argument{arg("issue", req.IssueID)}
	if req.Reason != "" {
		args = append(args, arg("closeReason", req.Reason))
	}
	o, err := l.s.callOp(ctx, opClose, req.Actor, args...)
	if err != nil {
		return issueops.CloseResult{}, err
	}
	switch o.Kind {
	case bdp.OutcomeSuccess:
		iss, err := issueFromOutcome(o)
		if err != nil {
			return issueops.CloseResult{}, err
		}
		return issueops.CloseResult{Issue: iss, Changed: true}, nil
	case bdp.OutcomeFailure:
		return issueops.CloseResult{}, closeFailure(req.IssueID, failureCode(o))
	default:
		return issueops.CloseResult{}, fmt.Errorf("bdpstore: closeIssue returned a %q outcome", o.Kind)
	}
}

// Update maps ONLY the claim state-machine path (Claim) onto `call issue/claim`
// (param: issue; actor in provenance). Plain field edits (Patch) are not the
// state machine — a separate Core put/patch question — and stay unsupported.
func (l bdpLifecycle) Update(ctx context.Context, req issueops.UpdateRequest) (issueops.UpdateResult, error) {
	if !req.Claim {
		return issueops.UpdateResult{}, l.s.unsupported("update (non-claim field edit)")
	}
	o, err := l.s.callOp(ctx, opClaim, req.Actor, arg("issue", req.IssueID))
	if err != nil {
		return issueops.UpdateResult{}, err
	}
	switch o.Kind {
	case bdp.OutcomeSuccess:
		iss, err := issueFromOutcome(o)
		if err != nil {
			return issueops.UpdateResult{}, err
		}
		return issueops.UpdateResult{Issue: iss, Changed: true}, nil
	case bdp.OutcomeFailure:
		return issueops.UpdateResult{}, claimFailure(req.IssueID, o)
	default:
		return issueops.UpdateResult{}, fmt.Errorf("bdpstore: claimIssue returned a %q outcome", o.Kind)
	}
}

// Reopen reopens a closed issue via `call issue/reopen` (params: issue, reason).
func (l bdpLifecycle) Reopen(ctx context.Context, req issueops.ReopenRequest) (issueops.ReopenResult, error) {
	args := []bdp.Argument{arg("issue", req.IssueID)}
	if req.Reason != "" {
		args = append(args, arg("reason", req.Reason))
	}
	o, err := l.s.callOp(ctx, opReopen, req.Actor, args...)
	if err != nil {
		return issueops.ReopenResult{}, err
	}
	switch o.Kind {
	case bdp.OutcomeSuccess:
		iss, err := issueFromOutcome(o)
		if err != nil {
			return issueops.ReopenResult{}, err
		}
		return issueops.ReopenResult{Issue: iss, Changed: true}, nil
	case bdp.OutcomeFailure:
		return issueops.ReopenResult{}, fmt.Errorf("bdpstore: reopenIssue refused: %s", failureCode(o))
	default:
		return issueops.ReopenResult{}, fmt.Errorf("bdpstore: reopenIssue returned a %q outcome", o.Kind)
	}
}

// --- Heartbeater (a role of its own, not Lifecycle) ---

type bdpHeartbeater struct{ s *Store }

var _ issueops.Heartbeater = bdpHeartbeater{}

// heartbeatValue is the provisional success shape of `call issue/heartbeat` (its
// return schema is not pinned in the catalog beyond rep<Issue>).
type heartbeatValue struct {
	IssueID        string    `json:"issueId"`
	Holder         string    `json:"holder"`
	HeartbeatAt    time.Time `json:"heartbeatAt"`
	LeaseExpiresAt time.Time `json:"leaseExpiresAt"`
	Rearmed        bool      `json:"rearmed"`
}

// Heartbeat proves lease liveness via `call issue/heartbeat` (param: issue).
func (h bdpHeartbeater) Heartbeat(ctx context.Context, req issueops.HeartbeatRequest) (issueops.HeartbeatResult, error) {
	o, err := h.s.callOp(ctx, opHeartbeat, req.Actor, arg("issue", req.IssueID))
	if err != nil {
		return issueops.HeartbeatResult{}, err
	}
	switch o.Kind {
	case bdp.OutcomeSuccess:
		var v heartbeatValue
		if len(o.Value) > 0 {
			if err := json.Unmarshal(o.Value, &v); err != nil {
				return issueops.HeartbeatResult{}, fmt.Errorf("bdpstore: decode heartbeat result: %w", err)
			}
		}
		if v.IssueID == "" {
			v.IssueID = req.IssueID
		}
		return issueops.HeartbeatResult{
			IssueID:        v.IssueID,
			Holder:         v.Holder,
			HeartbeatAt:    v.HeartbeatAt,
			LeaseExpiresAt: v.LeaseExpiresAt,
			Rearmed:        v.Rearmed,
		}, nil
	case bdp.OutcomeFailure:
		return issueops.HeartbeatResult{}, heartbeatFailure(req.IssueID, o)
	default:
		return issueops.HeartbeatResult{}, fmt.Errorf("bdpstore: heartbeatIssue returned a %q outcome", o.Kind)
	}
}

// --- declared-failure → issueops domain error reconstruction ---

// claimFailure reconstructs the issueops claim refusal (ClaimError → 409).
func claimFailure(issueID string, o bdp.Outcome) error {
	var e struct {
		Code     string `json:"code"`
		Assignee string `json:"assignee"`
	}
	_ = json.Unmarshal(o.Error, &e)
	sentinel := issueops.ErrNotClaimable
	if e.Code == "already-claimed" {
		sentinel = issueops.ErrAlreadyClaimed
	}
	return &issueops.ClaimConflictError{IssueID: issueID, Assignee: e.Assignee, Err: sentinel}
}

// heartbeatFailure reconstructs the issueops lease refusal (HeartbeatError → 409).
func heartbeatFailure(issueID string, o bdp.Outcome) error {
	var e struct {
		Code   string `json:"code"`
		Holder string `json:"holder"`
	}
	_ = json.Unmarshal(o.Error, &e)
	sentinel := issueops.ErrLeaseNotHeld
	if e.Code == "already-claimed" {
		sentinel = issueops.ErrAlreadyClaimed
	}
	return &issueops.LeaseConflictError{IssueID: issueID, Holder: e.Holder, Err: sentinel}
}

// closeFailure maps a close refusal `code` onto its issueops sentinel.
func closeFailure(issueID, code string) error {
	switch code {
	case "open-children":
		return fmt.Errorf("bdpstore: close %s refused: %w", issueID, issueops.ErrCloseOpenChildren)
	case "blocked":
		return fmt.Errorf("bdpstore: close %s refused: %w", issueID, issueops.ErrCloseBlocked)
	case "pinned":
		return fmt.Errorf("bdpstore: close %s refused: %w", issueID, issueops.ErrPinned)
	case "not-assignee":
		return fmt.Errorf("bdpstore: close %s refused: %w", issueID, issueops.ErrNotOwner)
	default:
		return fmt.Errorf("bdpstore: close %s refused: %s", issueID, code)
	}
}
