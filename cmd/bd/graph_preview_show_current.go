package main

import (
	"context"
	"strings"

	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
	publicops "github.com/steveyegge/beads/issueops"
)

// Keep shipping --current precedence while reading only the selected graph
// workspace. A failed list is never a miss: only two successful empty reads
// may use the last-touched fallback.
func graphPreviewCurrentIssueID(actor string) (string, error) {
	if actor == "" || strings.EqualFold(actor, "none") || strings.EqualFold(actor, "null") {
		return GetLastTouchedID(), nil
	}
	var selected string
	err := withGraphStoreOutput(func(ctx context.Context, s *graphstore.Store) (any, string, error) {
		one := 1
		for _, status := range []types.Status{types.StatusInProgress, types.StatusHooked} {
			page, err := s.ListIssues(ctx, publicops.ListRequest{
				Status: string(status), Assignee: actor, AllFlag: true, Limit: &one,
			})
			if err != nil {
				return nil, "", err
			}
			if len(page.Items) != 0 {
				selected, err = graphPreviewResourcePath(graphPreviewConfig.GraphScopeURL, page.Items[0].ID)
				return nil, "", err
			}
		}
		return nil, "", nil
	}, func(any, string) error { return nil })
	if err != nil {
		return "", err
	}
	if selected != "" {
		return selected, nil
	}
	return GetLastTouchedID(), nil
}
