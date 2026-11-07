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

// importIssuesRemote imports over applyBatch, whose create item is the only bulk
// create that carries an explicit id (batchCreateIssues mints its own, so an
// imported row would lose its identity). Each chunk is one all-or-nothing request.
//
// The create item carries no stored timestamps, comments or close state, so a row
// holding one refuses, naming the member, rather than arriving without it. It
// creates and never upserts, so a row the server already holds refuses too: the
// direct route's upsert merges columns inside a transaction, which no operation
// on the wire does.
//
// Rows naming no id and nothing else the narrower bulk create leaves out go there
// instead, which is the operation that exists for them.
func importIssuesRemote(ctx context.Context, st storage.DoltStorage, rows []*types.Issue, actor string, opts ImportOptions) error {
	if opts.ConflictSkip {
		return fmt.Errorf("skipping conflicting rows is not available over bd connect: bd serve's bulk create has no such mode")
	}
	if done, err := importWithBatchCreate(ctx, st, rows, actor); done || err != nil {
		return err
	}
	var items []issueops.ApplyItem
	for _, row := range rows {
		if row == nil {
			continue
		}
		if len(row.Comments) > 0 {
			return fmt.Errorf("%s carries comments, which bd serve's bulk create cannot store with their timestamps; import it on the server's own workspace", row.ID)
		}
		issue := *row
		deps := issue.Dependencies
		issue.Dependencies = nil
		items = append(items, issueops.ApplyItem{Kind: issueops.ItemCreate, Create: &issueops.CreateItem{Issue: &issue}})
		for _, dep := range deps {
			items = append(items, issueops.ApplyItem{Kind: issueops.ItemDepAdd, DepAdd: &issueops.DepAddItem{
				Source: issueops.Ref{ID: row.ID},
				Target: issueops.Ref{ID: dep.DependsOnID},
				Type:   dep.Type,
			}})
		}
	}
	applier, err := st.BatchApplier()
	if err != nil {
		return err
	}
	for len(items) > 0 {
		n := min(len(items), issueops.MaxApplyBatchItems)
		_, err := applier.ApplyBatch(ctx, issueops.ApplyBatchRequest{
			Actor:         actor,
			Items:         items[:n],
			ForceIDPrefix: opts.SkipPrefixValidation,
		})
		if errors.Is(err, issueops.ErrAlreadyExists) {
			return fmt.Errorf("import over bd connect creates new issues only, and does not upsert a row the server holds: %w", err)
		}
		if err != nil {
			return err
		}
		items = items[n:]
	}
	return nil
}

// importWithBatchCreate imports id-less rows through batchCreateIssues, and
// reports false, with nothing sent, when a row is one that operation cannot carry.
func importWithBatchCreate(ctx context.Context, st storage.DoltStorage, rows []*types.Issue, actor string) (bool, error) {
	items := make([]issueops.BatchCreateItem, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		if row.ID != "" || len(row.Dependencies) > 0 || len(row.Comments) > 0 || (row.Status != "" && row.Status != types.StatusOpen) {
			return false, nil
		}
		issue := *row
		issue.Status = "" // open is the operation's own default, and it has no such member.
		items = append(items, issueops.BatchCreateItem{Issue: &issue})
	}
	creator, err := st.BatchCreator()
	if err != nil {
		return false, err
	}
	created, err := creator.CreateBatch(ctx, issueops.CreateBatchRequest{Actor: actor, Items: items})
	if isUnsupportedByStore(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	// The import reports each row's id, and the server minted them; the answer is
	// in request order (httpBatchCreator checks its length).
	next := 0
	for _, row := range rows {
		if row != nil {
			row.ID = created.Issues[next].ID
			next++
		}
	}
	return true, nil
}

// isUnsupportedByStore reports an act the store refuses for having no way to do it
// (nothing was sent), which a caller with another route takes as a reason to use
// that route rather than as a failure.
func isUnsupportedByStore(err error) bool {
	var unsupported *storage.ErrUnsupported
	return errors.As(err, &unsupported)
}

// createBatchOverEitherOperation is `bd create --file`. On a local store it is the
// role. On a remote one batchCreateIssues carries eight issue members; a request
// holding another (Owner, which `bd create` fills from the git identity, is the
// usual one) is refused before the dial, and goes as applyBatch's create items
// instead, which carry it in the same one transaction.
func createBatchOverEitherOperation(ctx context.Context, st storage.DoltStorage, request issueops.CreateBatchRequest) (issueops.CreateBatchResult, error) {
	creator, err := st.BatchCreator()
	if err != nil {
		return issueops.CreateBatchResult{}, err
	}
	result, err := creator.CreateBatch(ctx, request)
	if err == nil || !storeIsRemoteBackend(st) || !isUnsupportedByStore(err) {
		return result, err
	}

	items := make([]issueops.ApplyItem, 0, len(request.Items))
	for i, item := range request.Items {
		key := fmt.Sprintf("item-%d", i)
		items = append(items, issueops.ApplyItem{Kind: issueops.ItemCreate, Create: &issueops.CreateItem{
			Key: key, Issue: item.Issue, DefaultPriority: item.DefaultPriority,
		}})
		for _, dep := range item.Dependencies {
			items = append(items, issueops.ApplyItem{Kind: issueops.ItemDepAdd, DepAdd: &issueops.DepAddItem{
				Source: issueops.Ref{Key: key},
				Target: issueops.Ref{ID: dep.TargetID},
				Type:   dep.Type,
			}})
		}
	}
	applier, err := st.BatchApplier()
	if err != nil {
		return issueops.CreateBatchResult{}, err
	}
	applied, err := applier.ApplyBatch(ctx, issueops.ApplyBatchRequest{Actor: request.Actor, Items: items, ForceIDPrefix: request.ForceIDPrefix})
	if err != nil {
		return issueops.CreateBatchResult{}, err
	}
	// applyBatch answers ids and no rows, and the command reports the rows.
	ids := make([]string, 0, len(request.Items))
	for i := range request.Items {
		ids = append(ids, applied.Keys[fmt.Sprintf("item-%d", i)])
	}
	rows, err := st.GetIssuesByIDs(ctx, ids)
	if err != nil {
		return issueops.CreateBatchResult{}, fmt.Errorf("the issues were created, and reading them back failed: %w", err)
	}
	byID := make(map[string]*types.Issue, len(rows))
	for _, row := range rows {
		byID[row.ID] = row
	}
	issues := make([]*types.Issue, 0, len(ids))
	for _, id := range ids {
		if row, ok := byID[id]; ok {
			issues = append(issues, row)
		}
	}
	return issueops.CreateBatchResult{Issues: issues}, nil
}
