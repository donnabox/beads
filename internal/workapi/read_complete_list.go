package workapi

import (
	"context"
	"errors"
	"fmt"

	"github.com/steveyegge/beads/issueops"
)

// ErrIncompleteListSelection prevents a truncated or malformed reader answer
// from being reported as an exact cardinality. It is not an empty selection.
var ErrIncompleteListSelection = errors.New("incomplete List selection")

// ReadCompleteList applies the existing List population and ready policy, with
// delivery paging removed. It intentionally calls List rather than Count:
// Count has different Issue/wisp duplicate and readiness behavior. Reader
// access/routing and lifetime belong to the caller, just as for ordinary List.
func ReadCompleteList(ctx context.Context, reader interface {
	List(context.Context, issueops.ListRequest) (issueops.IssuePage, error)
}, request issueops.ListRequest) (issueops.IssuePage, error) {
	if err := ctx.Err(); err != nil {
		return issueops.IssuePage{}, err
	}
	if reader == nil {
		return issueops.IssuePage{}, &issueops.ErrUnsupported{Op: "ReadCompleteList", Backend: "nil"}
	}
	unlimited := 0
	request.Limit = &unlimited
	request.Offset = 0
	request.AfterCreatedAt = nil
	request.AfterID = ""
	request.AfterPriority = nil
	// MaxRows/MaxRowsSource, ReadyFlag, planes, visibility, parent, metadata,
	// order and hydration policy are deliberately unchanged.
	page, err := reader.List(ctx, request)
	if err != nil {
		return issueops.IssuePage{}, err
	}
	if page.HasMore {
		return issueops.IssuePage{}, fmt.Errorf("%w: reader truncated an unlimited List", ErrIncompleteListSelection)
	}
	seen := make(map[string]bool, len(page.Items))
	for _, item := range page.Items {
		if err := ctx.Err(); err != nil {
			return issueops.IssuePage{}, err
		}
		if item == nil || item.Issue == nil || item.ID == "" || seen[item.ID] {
			return issueops.IssuePage{}, fmt.Errorf("%w: reader returned an invalid or duplicate identity", ErrIncompleteListSelection)
		}
		seen[item.ID] = true
	}
	if err := ctx.Err(); err != nil {
		return issueops.IssuePage{}, err
	}
	return page, nil
}
