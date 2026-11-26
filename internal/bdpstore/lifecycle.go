package bdpstore

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/steveyegge/beads/internal/bdp"
	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/issueops"
)

// This file maps the beads issueops LIFECYCLE state machine onto BDP Core `call`
// of issue-pack operations — Lydia's ruling (2026-11-26): claim/close/reopen are
// issue-pack CALLABLES invoked via `call`, NOT patch (the lease/heartbeat fields
// are serviceManaged = read-only to patch, and claimIssue carries a declared
// domain refusal patch cannot). A declared failure rides Outcome.failure (409);
// the client reconstructs the issueops domain error from it.
//
// EVERYTHING HERE IS PROVISIONAL and NOT live-validated, flagged to Lydia/Britt:
//   - The operation element PATHS below are not pinned in core.lumen, the issue
//     pack, or the clean-room — the vendored issue-package catalog even uses
//     createBead/updateBead verbs, not these. Isolated so a flip is a one-liner.
//   - The call argument names and the failure-error schema (the `code`) are
//     likewise unspecified in-repo; derived from the Go request structs.
//   - cleanroom-v023 offers NO call at all, so none of this is live-validated —
//     only mock-server tests. Proving it needs a call-offering, issue-pack
//     server (cross-track dependency).
const (
	opClaimIssue = "issueops/claimIssue"
	opCloseIssue = "issueops/closeIssue"
	opReopen     = "issueops/reopenIssue"
)

// arg builds one named call argument, JSON-encoding v.
func arg(name string, v any) bdp.Argument {
	b, _ := json.Marshal(v)
	return bdp.Argument{Name: name, Value: b}
}

// callOp invokes op with args and returns the settled Outcome (a pre-dispatch
// fault or transport failure comes back as the error).
func (s *Store) callOp(ctx context.Context, op string, args ...bdp.Argument) (bdp.Outcome, error) {
	reply, err := s.scope.Call(ctx, bdp.ResourceID{ID: op}, bdp.Request{Arguments: args})
	if err != nil {
		return bdp.Outcome{}, err
	}
	return reply.Outcome, nil
}

// issueFromOutcome decodes a lifecycle call's success value as the post-op Issue.
func issueFromOutcome(o bdp.Outcome) (*types.Issue, error) {
	var iss types.Issue
	if len(o.Value) > 0 {
		if err := json.Unmarshal(o.Value, &iss); err != nil {
			return nil, fmt.Errorf("bdpstore: decode lifecycle result: %w", err)
		}
	}
	return &iss, nil
}

// failureCode pulls the provisional `code` discriminator from an Outcome.failure.
func failureCode(o bdp.Outcome) string {
	var e struct {
		Code string `json:"code"`
	}
	_ = json.Unmarshal(o.Error, &e)
	return e.Code
}

// Close retires an issue via `call issueops/closeIssue`.
func (l bdpLifecycle) Close(ctx context.Context, req issueops.CloseRequest) (issueops.CloseResult, error) {
	args := []bdp.Argument{arg("issue", req.IssueID), arg("actor", req.Actor)}
	if req.Reason != "" {
		args = append(args, arg("reason", req.Reason))
	}
	if req.Force {
		args = append(args, arg("force", true))
	}
	o, err := l.s.callOp(ctx, opCloseIssue, args...)
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

// Update maps ONLY the claim state-machine path (Claim) onto `call
// issueops/claimIssue`. Plain field edits (Patch) are not the state machine and
// are a separate question (Core put/patch of body fields), left unsupported here.
func (l bdpLifecycle) Update(ctx context.Context, req issueops.UpdateRequest) (issueops.UpdateResult, error) {
	if !req.Claim {
		return issueops.UpdateResult{}, l.s.unsupported("update (non-claim field edit)")
	}
	o, err := l.s.callOp(ctx, opClaimIssue, arg("issue", req.IssueID), arg("actor", req.Actor))
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

// Reopen moves a closed issue back to open via `call issueops/reopenIssue`.
func (l bdpLifecycle) Reopen(ctx context.Context, req issueops.ReopenRequest) (issueops.ReopenResult, error) {
	args := []bdp.Argument{arg("issue", req.IssueID), arg("actor", req.Actor)}
	if req.Reason != "" {
		args = append(args, arg("reason", req.Reason))
	}
	o, err := l.s.callOp(ctx, opReopen, args...)
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

// claimFailure reconstructs the issueops claim refusal from Outcome.failure. The
// ruling: ClaimError (already-claimed | not-claimable/closed) → 409. The failure
// `code` schema is provisional.
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

// closeFailure maps a close refusal `code` onto its issueops sentinel (so
// errors.Is classifies it); the typed-struct detail is dropped at this boundary.
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
