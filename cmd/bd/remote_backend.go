package main

import (
	"context"
	"fmt"

	"github.com/steveyegge/beads/internal/beads"
	"github.com/steveyegge/beads/internal/configfile"
	"github.com/steveyegge/beads/internal/httpclient"
	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/issueops"
)

// THE ATTRIBUTION OMISSIONS (S6d).
//
// A store that reports storage.CallerAttributionLimitedStore (today the http
// backend) refuses two caller-authored members its wire cannot carry, and
// would otherwise refuse two everyday commands outright. The CLI omits them
// for such a store only where the omission loses nothing a reader relies on:
//
//   - `bd reopen`'s Provenance. The CLI's label is the fixed "bd: reopen <id>",
//     fully derivable from the operation and the id; the server records its
//     own reopen entry for the same issue and the same actor. Nothing is lost.
//
//   - `bd update -s closed`'s closed_by_session. It is attribution only: no
//     bd or gc behavior (ready, close guards, claims) reads it, and the status
//     change, its close-policy guards and the atomicity with every other field
//     in the same update are all kept. It is also usually AMBIENT — every agent
//     under Claude Code exports CLAUDE_SESSION_ID — so refusing the close
//     would break the command for an input the caller never typed. The close
//     still happens, and a one-line notice on stderr says the session was not
//     recorded — for an explicit --session and for the ambient
//     CLAUDE_SESSION_ID alike, so the omission is never silent — pointing at
//     `bd close`, whose wire operation does carry the session.
//
// No capability token is spent on either: the wire publishes neither member,
// so there is nothing to negotiate yet. When a wire revision publishes one,
// the store's CallerAttributionLimited answer becomes the handshake check and
// these call sites need no change.

// storeCarriesCallerAttribution reports whether s can record the caller's own
// history label and closed_by_session.
func storeCarriesCallerAttribution(s storage.DoltStorage) bool {
	if s == nil {
		return true
	}
	limited, ok := storage.UnwrapStore(s).(storage.CallerAttributionLimitedStore)
	return !ok || !limited.CallerAttributionLimited()
}

// reopenProvenance is `bd reopen`'s history label for store, or "" (the
// backend's own label) where the store cannot carry one.
func reopenProvenance(store storage.DoltStorage, id string) string {
	if !storeCarriesCallerAttribution(store) {
		return ""
	}
	return "bd: reopen " + id
}

// storeIsRemoteBackend reports whether s is a pure network client of a remote
// bd serve (storage.RemoteBackendStore), seen through the decorator chain.
func storeIsRemoteBackend(s storage.DoltStorage) bool {
	if s == nil {
		return false
	}
	remote, ok := storage.UnwrapStore(s).(storage.RemoteBackendStore)
	return ok && remote.IsRemoteBackendStore()
}

// remoteOnV0 reports a store attached by a bd connect speaking v0, the default.
//
// A v0 connection sends what it sent before BDP, request for request, except for a
// command that used to refuse and now works. So a change to a command that already
// answered, for the better or not, is for connections that asked for BDP, and every
// use of this names what the old code did instead. That includes a command that read
// through a refused raw method, swallowed the refusal and printed less: the read
// stays unsent here, and the gap stays as it was (scripts/bdp-call-parity's
// wire_proof.py holds all of this to a bd built before the change).
func remoteOnV0(s storage.DoltStorage) bool {
	remote, ok := storage.UnwrapStore(s).(interface{ Protocol() httpclient.Protocol })
	return ok && remote.Protocol() == httpclient.ProtocolV0
}

// workspaceOnV0Connection is remoteOnV0 for a check that runs before any store is
// open. A sidecar it cannot read is left for the store's open to report.
func workspaceOnV0Connection() bool {
	if !workspaceIsRemoteBackend() {
		return false
	}
	target, err := httpclient.LoadTarget(beads.FindBeadsDir())
	return err == nil && target.Protocol == httpclient.ProtocolV0
}

// refuseKVOverRemoteBackend is the answer every `bd kv` subcommand gives on a
// bd connect workspace speaking BDP. bd serve deliberately hides the kv.* config
// namespace (issueops/workspaceconfig.go), so a read would answer "not
// found" for a key that exists and a write would be refused by key. Saying so
// up front is the only true answer the family has. On v0 the read goes out as it
// always did, and gets that answer: see remoteOnV0.
func refuseKVOverRemoteBackend(s storage.DoltStorage, verb string) error {
	if !storeIsRemoteBackend(s) || remoteOnV0(s) {
		return nil
	}
	return HandleErrorRespectJSON("bd kv %s is not available over bd connect: bd serve does not expose the kv.* namespace to clients; run it on the server's own workspace", verb)
}

// hydrateRemoteDependents replaces the shallow rows a remote store's dependents
// read answers with whole ones. bd serve projects that read onto id, status,
// type, priority, title and edge type, which is enough for `bd show`'s
// dependents section and not for a caller rendering the row whole: every other
// field would print blank and every timestamp as 0001-01-01. A row the follow-up
// read cannot supply refuses rather than printing that. On v0 the shallow rows are
// the answer, as they always were: see remoteOnV0.
func hydrateRemoteDependents(ctx context.Context, s storage.DoltStorage, refs []*types.IssueWithDependencyMetadata) ([]*types.IssueWithDependencyMetadata, error) {
	if !storeIsRemoteBackend(s) || remoteOnV0(s) || len(refs) == 0 {
		return refs, nil
	}
	ids := make([]string, 0, len(refs))
	for _, ref := range refs {
		ids = append(ids, ref.ID)
	}
	rows, err := s.GetIssuesByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	byID := make(map[string]*types.Issue, len(rows))
	for _, row := range rows {
		byID[row.ID] = row
	}
	hydrated := make([]*types.IssueWithDependencyMetadata, 0, len(refs))
	for _, ref := range refs {
		row, ok := byID[ref.ID]
		if !ok {
			return nil, fmt.Errorf("the server listed %q as a dependent but did not return its row", ref.ID)
		}
		hydrated = append(hydrated, &types.IssueWithDependencyMetadata{Issue: *row, DependencyType: ref.DependencyType})
	}
	return hydrated, nil
}

// releaseThroughRole is `bd unclaim` on a remote store: the two raw unclaims have
// no operation of their own, and the Releaser role is what dials releaseIssue.
func releaseThroughRole(ctx context.Context, st storage.DoltStorage, id string, conditional bool, expected string, force bool) error {
	releaser, err := st.Releaser()
	if err != nil {
		return err
	}
	req := issueops.ReleaseRequest{Actor: currentActor(), IssueID: id, Force: force}
	if conditional {
		req.ExpectedAssignee = &expected
	}
	_, err = releaser.Release(ctx, req)
	return err
}

// addReasonComment appends a `--reason` comment. A remote store has no raw
// comment append; the Commenter role is what dials addComment.
func addReasonComment(ctx context.Context, st storage.DoltStorage, id, reason string) error {
	if !storeIsRemoteBackend(st) {
		_, err := st.AddIssueComment(ctx, id, currentActor(), reason)
		return err
	}
	commenter, err := st.Commenter()
	if err != nil {
		return err
	}
	_, err = commenter.AddComment(ctx, issueops.AddCommentRequest{Author: currentActor(), IssueID: id, Text: reason})
	return err
}

// workspaceIsRemoteBackend reports whether this workspace is attached with `bd
// connect`, for a command that decides before any store is open.
func workspaceIsRemoteBackend() bool {
	beadsDir := beads.FindBeadsDir()
	if beadsDir == "" {
		return false
	}
	cfg, err := configfile.Load(beadsDir)
	return err == nil && cfg != nil && cfg.GetBackend() == httpclient.Backend
}

// remoteClaimNext is `bd close --claim-next` on a remote store. batchCloseIssues
// publishes no claim-next member, so the claim is a second request after the
// closes rather than part of their transaction; ready.claim-next is the same
// selection the direct route's batch makes.
func remoteClaimNext(ctx context.Context, st storage.DoltStorage, filter issueops.ReadyRequest) (*types.IssueWithCounts, error) {
	claimer, err := st.ReadyClaimer()
	if err != nil {
		return nil, err
	}
	result, err := claimer.ClaimNext(ctx, issueops.ClaimNextRequest{Actor: currentActor(), Filter: filter})
	if err != nil {
		return nil, err
	}
	return result.Claimed, nil
}
