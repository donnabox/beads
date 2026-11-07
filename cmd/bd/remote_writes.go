package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/steveyegge/beads/internal/storage"
	storageops "github.com/steveyegge/beads/internal/storage/issueops"
	"github.com/steveyegge/beads/internal/types"
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

// runBatchRemote applies a `bd batch` script as one applyBatch request: the same
// whole-batch rollback on the first failing line, with the failing line reported
// from the item the server names.
//
// A line the operation has no item kind for refuses here, naming the line, before
// anything is sent: sending the rest and dropping it would be a batch that did
// less than the script said.
func runBatchRemote(ctx context.Context, st storage.DoltStorage, ops []batchOp, commitMsg string) ([]batchOpResult, error) {
	items := make([]issueops.ApplyItem, 0, len(ops))
	results := make([]batchOpResult, 0, len(ops))
	for _, op := range ops {
		item, target, err := batchApplyItem(op)
		if err != nil {
			return nil, fmt.Errorf("line %d (%s): %w", op.line, op.raw, err)
		}
		items = append(items, item)
		results = append(results, batchOpResult{Line: op.line, Op: op.cmd, Target: target})
	}

	applier, err := st.BatchApplier()
	if err != nil {
		return nil, err
	}
	applied, err := applier.ApplyBatch(ctx, issueops.ApplyBatchRequest{
		Actor:      getActor(),
		Items:      items,
		Provenance: commitMsg,
	})
	if err != nil {
		var itemErr *issueops.ItemError
		if errors.As(err, &itemErr) && itemErr.Index >= 0 && itemErr.Index < len(ops) {
			op := ops[itemErr.Index]
			return nil, fmt.Errorf("line %d (%s): %w", op.line, op.raw, itemErr.Err)
		}
		return nil, err
	}
	for i := range results {
		if results[i].Op == "create" {
			results[i].Target = applied.Items[i].IssueID
		}
	}
	return results, nil
}

// batchApplyItem projects one parsed line onto applyBatch's item, and answers the
// target it reports. The argument rules are runBatchOp's own.
func batchApplyItem(op batchOp) (issueops.ApplyItem, string, error) {
	switch op.cmd {
	case "close":
		if len(op.args) < 1 {
			return issueops.ApplyItem{}, "", fmt.Errorf("close requires <id>")
		}
		id := op.args[0]
		reason := "Closed"
		if len(op.args) > 1 {
			reason = strings.Join(op.args[1:], " ")
		}
		// Force because the direct route's transaction close runs no close
		// policy: without it the same line would refuse here and close there.
		return issueops.ApplyItem{Kind: issueops.ItemClose, Close: &issueops.CloseItem{
			Target: issueops.Ref{ID: id}, Reason: reason, Force: true,
		}}, id, nil

	case "update":
		if len(op.args) < 2 {
			return issueops.ApplyItem{}, "", fmt.Errorf("update requires <id> and at least one key=value")
		}
		id := op.args[0]
		updates, err := parseUpdateKVs(op.args[1:])
		if err != nil {
			return issueops.ApplyItem{}, "", err
		}
		item := &issueops.UpdateItem{Target: issueops.Ref{ID: id}}
		for key, value := range updates {
			switch key {
			case "status":
				item.Patch.Status = issueops.Field[types.Status]{Set: true, Value: types.Status(value.(string))}
			case "priority":
				item.Patch.Priority = issueops.Field[int]{Set: true, Value: value.(int)}
			case "title":
				item.Patch.Title = issueops.Field[string]{Set: true, Value: value.(string)}
			case "assignee":
				item.Patch.Assignee = issueops.Field[string]{Set: true, Value: value.(string)}
				// The transaction's raw update has no live-claim fence; the
				// fence is `bd update`'s and `bd assign`'s.
				item.ForceAssigneeTransfer = true
			case storageops.OpForceClosePolicy:
				item.ForceClosePolicy = value.(bool)
			default:
				return issueops.ApplyItem{}, "", fmt.Errorf("update: unsupported key %q", key)
			}
		}
		return issueops.ApplyItem{Kind: issueops.ItemUpdate, Update: item}, id, nil

	case "create":
		if len(op.args) < 3 {
			return issueops.ApplyItem{}, "", fmt.Errorf("create requires <type> <priority> <title>")
		}
		if strings.TrimSpace(op.args[0]) == "" {
			return issueops.ApplyItem{}, "", fmt.Errorf("create: type cannot be empty")
		}
		priority, err := strconv.Atoi(op.args[1])
		if err != nil {
			return issueops.ApplyItem{}, "", fmt.Errorf("create: invalid priority %q: %w", op.args[1], err)
		}
		title := strings.Join(op.args[2:], " ")
		if strings.TrimSpace(title) == "" {
			return issueops.ApplyItem{}, "", fmt.Errorf("create: title cannot be empty")
		}
		return issueops.ApplyItem{Kind: issueops.ItemCreate, Create: &issueops.CreateItem{Issue: &types.Issue{
			Title:     title,
			IssueType: types.IssueType(op.args[0]),
			Status:    types.StatusOpen,
			Priority:  priority,
		}}}, "", nil

	case "dep.add":
		dep, err := batchDepAddEdge(op.args)
		if err != nil {
			return issueops.ApplyItem{}, "", err
		}
		return issueops.ApplyItem{Kind: issueops.ItemDepAdd, DepAdd: &issueops.DepAddItem{
			Source: issueops.Ref{ID: dep.IssueID},
			Target: issueops.Ref{ID: dep.DependsOnID},
			Type:   dep.Type,
		}}, fmt.Sprintf("%s->%s", dep.IssueID, dep.DependsOnID), nil

	case "dep.remove":
		return issueops.ApplyItem{}, "", fmt.Errorf("bd batch over bd connect cannot remove a dependency: bd serve's batch operation has no such item, and nothing was sent; run the removal as `bd dep remove` on its own")
	}
	return issueops.ApplyItem{}, "", fmt.Errorf("internal: unhandled batch op %q", op.cmd)
}
