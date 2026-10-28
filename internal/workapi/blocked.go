package workapi

import (
	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/internal/utils"
	"github.com/steveyegge/beads/issueops"
)

// BuildGraphBlockedFilter projects the blocked request's label clauses into
// the graph-preview store query. The CLI resolves --parent against the graph
// catalog in the blocked read transaction; ordinary listings use BlockedLister.
func BuildGraphBlockedFilter(in issueops.BlockedRequest) types.WorkFilter {
	return types.WorkFilter{
		Labels:        utils.NormalizeLabels(in.Labels),
		LabelsAny:     utils.NormalizeLabels(in.LabelsAny),
		ExcludeLabels: utils.NormalizeLabels(in.ExcludeLabels),
	}
}
