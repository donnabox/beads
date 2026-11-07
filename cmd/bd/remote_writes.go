package main

import (
	"context"

	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/issueops"
)

// The writes a bd connect workspace sends through a role where the direct route
// runs a storage transaction. A transaction cannot span requests; each of these is
// one request (or, past the wire's cap, one per chunk) to the role that does the
// same act in one transaction on the server.

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
