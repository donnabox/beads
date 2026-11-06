package httpclient

import (
	"context"
	"fmt"
	"net/url"
	"slices"
	"sort"
	"time"

	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/issueops"
)

// The raw storage reads a composition of v0 operations can answer exactly.
//
// completeness_test.go's rule puts a method on the refusing shell only when no
// v0 operation can answer it, and these were on the shell although several
// operations together can: a command reading through one of them got a refusal
// it often logged at debug and carried on from, which is how `bd graph` lost
// its dependents and `bd ready --explain` its blockers with exit 0.
//
// Each read below asks the server for the fields it needs rather than deciding
// what the caller needs, because the caller is a command written against the
// local store and reads whatever the local row carries. So a shallow wire row
// (getIssue's dependents arm: id, status, type, priority, title, edge type) is
// always hydrated before it is handed back. A row that vanishes between the two
// reads is absent from the answer, as a row deleted a moment earlier would be
// locally.

// GetIssuesByIDs answers batchGetIssues, in chunks of its published cap. An id
// that names nothing is absent from the answer, as it is locally.
func (s *Store) GetIssuesByIDs(ctx context.Context, ids []string) ([]*types.Issue, error) {
	wanted := distinct(ids)
	if len(wanted) == 0 {
		return nil, nil
	}
	getter, err := s.BatchGetter()
	if err != nil {
		return nil, err
	}
	out := make([]*types.Issue, 0, len(wanted))
	for _, chunk := range chunkStrings(wanted, issueops.MaxGetManyIDs) {
		res, err := getter.GetMany(ctx, issueops.GetManyRequest{IDs: chunk})
		if err != nil {
			return nil, err
		}
		out = append(out, res.Issues...)
	}
	return out, nil
}

// GetDependents answers getIssue's dependents arm with the rows hydrated.
func (s *Store) GetDependents(ctx context.Context, id string) ([]*types.Issue, error) {
	edges, err := s.GetDependentsWithMetadata(ctx, id)
	if err != nil || len(edges) == 0 {
		return nil, err
	}
	ids := make([]string, 0, len(edges))
	for _, edge := range edges {
		ids = append(ids, edge.ID)
	}
	return s.GetIssuesByIDs(ctx, ids)
}

// GetDependencies answers listDependencies with the targets hydrated. An edge
// whose target names no row (an external reference, a dangling id) has no issue
// to return, which is also the local answer.
func (s *Store) GetDependencies(ctx context.Context, id string) ([]*types.Issue, error) {
	records, err := s.GetDependencyRecords(ctx, id)
	if err != nil || len(records) == 0 {
		return nil, err
	}
	ids := make([]string, 0, len(records))
	for _, record := range records {
		ids = append(ids, record.DependsOnID)
	}
	return s.GetIssuesByIDs(ctx, ids)
}

// GetDependencyRecordsForIssues answers listDependencies in chunks of its
// anchor bound. An id with no outgoing edges, or naming nothing, has no key in
// the answer, as it has none locally.
func (s *Store) GetDependencyRecordsForIssues(ctx context.Context, ids []string) (map[string][]*types.Dependency, error) {
	result := make(map[string][]*types.Dependency)
	reader := httpEdgeReader{store: s}
	for _, chunk := range chunkStrings(distinct(ids), maxEdgeReadAnchors) {
		res, err := reader.readEdgeChunk(ctx, chunk, nil)
		if err != nil {
			return nil, err
		}
		for _, anchor := range res.Anchors {
			if len(anchor.Edges) > 0 {
				result[anchor.ID] = anchor.Edges
			}
		}
	}
	return result, nil
}

// GetDependencyCounts answers the blocks-only pair from countDependencyEdges,
// one call for each direction per chunk of its anchor bound. Every requested id
// gets an entry, the local shape.
func (s *Store) GetDependencyCounts(ctx context.Context, ids []string) (map[string]*types.DependencyCounts, error) {
	result := make(map[string]*types.DependencyCounts, len(ids))
	for _, id := range ids {
		result[id] = &types.DependencyCounts{}
	}
	counter, err := s.GraphCounter()
	if err != nil {
		return nil, err
	}
	blocks := []types.DependencyType{types.DepBlocks}
	for _, chunk := range chunkStrings(distinct(ids), maxEdgeCountAnchors) {
		out, err := counter.CountEdges(ctx, issueops.EdgeCountRequest{IDs: chunk, Direction: issueops.EdgeDirectionOut, Types: blocks})
		if err != nil {
			return nil, err
		}
		for _, anchor := range out.Anchors {
			result[anchor.ID].DependencyCount = int(anchor.Count)
		}
		in, err := counter.CountEdges(ctx, issueops.EdgeCountRequest{IDs: chunk, Direction: issueops.EdgeDirectionIn, Types: blocks})
		if err != nil {
			return nil, err
		}
		for _, anchor := range in.Anchors {
			result[anchor.ID].DependentCount = int(anchor.Count)
		}
	}
	return result, nil
}

// DetectCycles answers listDependencyCycles in the raw method's shape. A cycle
// member the server could not hydrate has no issue to put in the path, and a
// path silently missing a member would read as a shorter cycle, so it refuses.
func (s *Store) DetectCycles(ctx context.Context) ([][]*types.Issue, error) {
	report, err := httpCycleDetector{store: s}.DetectCycles(ctx, issueops.DetectCyclesRequest{})
	if err != nil {
		return nil, err
	}
	if len(report.Cycles) == 0 {
		return nil, nil
	}
	cycles := make([][]*types.Issue, 0, len(report.Cycles))
	for _, cycle := range report.Cycles {
		path := make([]*types.Issue, 0, len(cycle.Members))
		for _, member := range cycle.Members {
			if member.Issue == nil {
				return nil, fmt.Errorf("the server at %s reported a dependency cycle through %q without the issue itself",
					s.target.String(), member.ID)
			}
			path = append(path, member.Issue)
		}
		cycles = append(cycles, path)
	}
	return cycles, nil
}

// GetStaleIssues composes the staleness question from listIssues: the same
// status set, labels and updated_at bound the local query reads. The listing has
// no predicate for the durable plane under an infra-inclusive read, nor for the
// live-heartbeat exemption, so both are applied to the rows, along with the
// oldest-first order and the limit, which have to follow them.
func (s *Store) GetStaleIssues(ctx context.Context, filter types.StaleFilter) ([]*types.Issue, error) {
	cutoff := time.Now().UTC().AddDate(0, 0, -filter.Days)
	search := types.IssueFilter{
		UpdatedBefore: &cutoff,
		Labels:        filter.Labels,
		LabelsAny:     filter.LabelsAny,
		ExcludeLabels: filter.ExcludeLabels,
	}
	if filter.Status != "" {
		status := types.Status(filter.Status)
		search.Status = &status
	} else {
		search.Statuses = []types.Status{types.StatusOpen, types.StatusInProgress}
	}
	rows, err := s.SearchIssues(ctx, "", search)
	if err != nil {
		return nil, err
	}
	stale := make([]*types.Issue, 0, len(rows))
	for _, row := range rows {
		if row.Ephemeral || row.NoHistory {
			continue
		}
		if row.HeartbeatAt != nil && !row.HeartbeatAt.Before(cutoff) {
			continue
		}
		stale = append(stale, row)
	}
	sort.SliceStable(stale, func(i, j int) bool { return stale[i].UpdatedAt.Before(stale[j].UpdatedAt) })
	if filter.Limit > 0 && len(stale) > filter.Limit {
		stale = stale[:filter.Limit]
	}
	return stale, nil
}

// GetEpicsEligibleForClosure composes the closure question: the unclosed durable
// epics from listIssues, each one's children from getIssue's dependents arm, and
// the status and close reason of those children from batchGetIssues, because a
// close reason decides eligibility (GH#5026) and the dependents arm omits it.
func (s *Store) GetEpicsEligibleForClosure(ctx context.Context) ([]*types.EpicStatus, error) {
	epicType := types.IssueType("epic")
	epics, err := s.SearchIssues(ctx, "", types.IssueFilter{
		IssueType:     &epicType,
		ExcludeStatus: []types.Status{types.StatusClosed},
	})
	if err != nil {
		return nil, err
	}
	slices.SortFunc(epics, func(a, b *types.Issue) int {
		switch {
		case a.ID < b.ID:
			return -1
		case a.ID > b.ID:
			return 1
		}
		return 0
	})

	childrenOf := make(map[string][]string, len(epics))
	var allChildren []string
	for _, epic := range epics {
		if epic.Ephemeral || epic.NoHistory {
			continue
		}
		q := url.Values{}
		q.Set("include_dependents", "true")
		details, err := s.getIssueDetails(ctx, epic.ID, q)
		if err != nil {
			return nil, err
		}
		if details == nil {
			continue
		}
		for _, edge := range details.Dependents {
			if edge.DependencyType == types.DepParentChild {
				childrenOf[epic.ID] = append(childrenOf[epic.ID], edge.ID)
				allChildren = append(allChildren, edge.ID)
			}
		}
	}
	rows, err := s.GetIssuesByIDs(ctx, allChildren)
	if err != nil {
		return nil, err
	}
	childByID := make(map[string]*types.Issue, len(rows))
	for _, row := range rows {
		childByID[row.ID] = row
	}

	var results []*types.EpicStatus
	for _, epic := range epics {
		children := childrenOf[epic.ID]
		if len(children) == 0 {
			continue
		}
		closed, completing := 0, 0
		for _, id := range children {
			child, ok := childByID[id]
			if !ok || child.Status != types.StatusClosed {
				continue
			}
			closed++
			if !types.IsNonCompletingClose(child.CloseReason) {
				completing++
			}
		}
		results = append(results, &types.EpicStatus{
			Epic:             epic,
			TotalChildren:    len(children),
			ClosedChildren:   closed,
			EligibleForClose: len(children) == completing,
		})
	}
	return results, nil
}
