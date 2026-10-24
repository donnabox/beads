package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/query"
	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	storageissueops "github.com/steveyegge/beads/internal/storage/issueops"
	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/internal/utils"
	"github.com/steveyegge/beads/internal/validation"
	"github.com/steveyegge/beads/internal/workapi"
	"github.com/steveyegge/beads/issueops"
)

// Graph discovery deliberately uses the native Issue filter and JSON row,
// including counts, after complete graph record checks. Unsupported options
// are refused before opening the store; they never silently select a list.
func graphDiscoveryRows(query string, filter types.IssueFilter) ([]*types.IssueWithCounts, error) {
	var rows []*types.IssueWithCounts
	err := withGraphStoreOutput(func(ctx context.Context, s *graphstore.Store) (any, string, error) {
		var err error
		rows, err = s.SearchIssueRows(ctx, query, filter)
		return nil, "", err
	}, func(any, string) error { return nil })
	return rows, err
}

func graphSearchDiscoveryRows(filter types.IssueFilter, status string) ([]*types.IssueWithCounts, error) {
	var rows []*types.IssueWithCounts
	err := withGraphStoreOutput(func(ctx context.Context, s *graphstore.Store) (any, string, error) {
		if status != "" && status != "all" {
			custom, err := s.ListCustomStatuses(ctx)
			if err != nil {
				return nil, "", err
			}
			names := make([]string, 0, len(custom))
			for _, item := range custom {
				names = append(names, item.Name)
			}
			if err := workapi.ApplyStatusFilter(&filter, status, names); err != nil {
				return nil, "", graphFailure("invalid_selector", err.Error(), 2)
			}
		}
		var err error
		rows, err = s.SearchIssueRows(ctx, "", filter)
		return nil, "", err
	}, func(any, string) error { return nil })
	return rows, err
}

func graphCountDiscoveryRows(filter types.IssueFilter, selectors []graphstore.IssueSelector) ([]*types.IssueWithCounts, []string, error) {
	var rows []*types.IssueWithCounts
	var paths []string
	err := withGraphStoreOutput(func(ctx context.Context, s *graphstore.Store) (any, string, error) {
		var err error
		paths, err = s.ResolveIssuePaths(ctx, selectors)
		if err != nil {
			return nil, "", err
		}
		rows, err = s.SearchIssueRows(ctx, "", filter)
		return nil, "", err
	}, func(any, string) error { return nil })
	return rows, paths, err
}

func graphDiscoveryJSON(w io.Writer, value any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(wrapWithSchemaVersion(value)); err != nil {
		return err
	}
	if !jsonEnvelopeEnabled() {
		emitEnvelopeDeprecation()
	}
	return nil
}

func graphDiscoveryFilter() types.IssueFilter {
	return types.IssueFilter{SkipWisps: true}
}

func graphDiscoveryFlags(cmd *cobra.Command, allowed ...string) error {
	if cmd.Flags().Changed("format") {
		format, _ := cmd.Flags().GetString("format")
		if !strings.EqualFold(format, "json") {
			return graphFailure("capability_unavailable", "graph "+cmd.Name()+" supports --format json only", 5)
		}
	}
	return graphPreviewFlags(cmd, append(allowed, "format")...)
}

func graphDiscoveryWantsJSON(cmd *cobra.Command) bool {
	format, _ := cmd.Flags().GetString("format")
	return jsonOutput || strings.EqualFold(format, "json")
}

func graphDiscoveryText(raw, name string) error {
	if !utf8.ValidString(raw) || len(raw) > graphstore.PreviewIssueListRequestByteLimit {
		return graphFailure("invalid_selector", name+" must be UTF-8 and within the graph preview request bound", 2)
	}
	return nil
}

func runGraphPreviewSearch(cmd *cobra.Command, args []string) error {
	if err := graphDiscoveryFlags(cmd, "query", "status", "assignee", "type", "label", "label-any", "limit", "long", "sort", "reverse", "priority-min", "priority-max", "desc-contains", "notes-contains", "external-contains", "empty-description", "no-assignee", "no-labels", "created-after", "created-before", "updated-after", "updated-before", "closed-after", "closed-before", "metadata-field", "has-metadata-key"); err != nil {
		return err
	}
	query := strings.Join(args, " ")
	if query == "" {
		query, _ = cmd.Flags().GetString("query")
	}
	if query == "" {
		return graphFailure("invalid_selector", "search query is required", 2)
	}
	if err := graphDiscoveryText(query, "search query"); err != nil {
		return err
	}
	if len(query) > 4096 {
		return graphFailure("invalid_selector", "search query must be at most 4096 bytes", 2)
	}
	filter := graphDiscoveryFilter()
	status, _ := cmd.Flags().GetString("status")
	assignee, _ := cmd.Flags().GetString("assignee")
	if assignee != "" {
		filter.Assignee = &assignee
	}
	issueType, _ := cmd.Flags().GetString("type")
	if issueType != "" {
		value := types.IssueType(issueType)
		filter.IssueType = &value
	}
	filter.Labels, _ = cmd.Flags().GetStringSlice("label")
	filter.LabelsAny, _ = cmd.Flags().GetStringSlice("label-any")
	filter.Labels = utils.NormalizeLabels(filter.Labels)
	filter.LabelsAny = utils.NormalizeLabels(filter.LabelsAny)
	if len(filter.Labels)+len(filter.LabelsAny) > graphstore.PreviewIssueListLabelLimit {
		return graphFailure("invalid_selector", "too many search labels", 2)
	}
	for _, label := range append(append([]string{}, filter.Labels...), filter.LabelsAny...) {
		if err := graphDiscoveryText(label, "search label"); err != nil {
			return err
		}
	}
	filter.DescriptionContains, _ = cmd.Flags().GetString("desc-contains")
	filter.NotesContains, _ = cmd.Flags().GetString("notes-contains")
	filter.ExternalRefContains, _ = cmd.Flags().GetString("external-contains")
	filter.EmptyDescription, _ = cmd.Flags().GetBool("empty-description")
	filter.NoAssignee, _ = cmd.Flags().GetBool("no-assignee")
	filter.NoLabels, _ = cmd.Flags().GetBool("no-labels")
	for _, bound := range []struct {
		flag string
		dest **time.Time
	}{
		{"created-after", &filter.CreatedAfter},
		{"created-before", &filter.CreatedBefore},
		{"updated-after", &filter.UpdatedAfter},
		{"updated-before", &filter.UpdatedBefore},
		{"closed-after", &filter.ClosedAfter},
		{"closed-before", &filter.ClosedBefore},
	} {
		raw, _ := cmd.Flags().GetString(bound.flag)
		if raw == "" {
			continue
		}
		parsed, err := parseTimeFlag(raw)
		if err != nil {
			return graphFailure("invalid_selector", "parsing --"+bound.flag+": "+err.Error(), 2)
		}
		*bound.dest = &parsed
	}
	metadataFields, _ := cmd.Flags().GetStringArray("metadata-field")
	for _, field := range metadataFields {
		key, value, ok := strings.Cut(field, "=")
		if !ok || key == "" {
			return graphFailure("invalid_selector", fmt.Sprintf("invalid --metadata-field: expected key=value, got %q", field), 2)
		}
		if err := storage.ValidateMetadataKey(key); err != nil {
			return graphFailure("invalid_selector", "invalid --metadata-field key: "+err.Error(), 2)
		}
		if err := graphDiscoveryText(value, "metadata value"); err != nil {
			return err
		}
		if filter.MetadataFields == nil {
			filter.MetadataFields = make(map[string]string, len(metadataFields))
		}
		filter.MetadataFields[key] = value
	}
	filter.HasMetadataKey, _ = cmd.Flags().GetString("has-metadata-key")
	if filter.HasMetadataKey != "" {
		if err := storage.ValidateMetadataKey(filter.HasMetadataKey); err != nil {
			return graphFailure("invalid_selector", "invalid --has-metadata-key: "+err.Error(), 2)
		}
	}
	if cmd.Flags().Changed("priority-min") {
		raw, _ := cmd.Flags().GetString("priority-min")
		value, err := validation.ValidatePriority(raw)
		if err != nil {
			return graphFailure("invalid_selector", err.Error(), 2)
		}
		filter.PriorityMin = &value
	}
	if cmd.Flags().Changed("priority-max") {
		raw, _ := cmd.Flags().GetString("priority-max")
		value, err := validation.ValidatePriority(raw)
		if err != nil {
			return graphFailure("invalid_selector", err.Error(), 2)
		}
		filter.PriorityMax = &value
	}
	for _, value := range []string{status, assignee, issueType, filter.DescriptionContains, filter.NotesContains, filter.ExternalRefContains} {
		if err := graphDiscoveryText(value, "search filter"); err != nil {
			return err
		}
	}
	limit, _ := cmd.Flags().GetInt("limit")
	if limit < 0 {
		return graphFailure("invalid_selector", "search limit must be non-negative", 2)
	}
	sortBy, _ := cmd.Flags().GetString("sort")
	if !graphDiscoverySortSupported(sortBy) {
		return graphFailure("capability_unavailable", "graph search does not support --sort "+sortBy, 5)
	}
	reverse, _ := cmd.Flags().GetBool("reverse")
	longFormat, _ := cmd.Flags().GetBool("long")
	rows, err := graphSearchDiscoveryRows(filter, status)
	if err != nil {
		return err
	}
	rows, err = graphSearchMatches(rows, query)
	if err != nil {
		return graphFailure("invalid_selector", "invalid search pattern: "+err.Error(), 2)
	}
	// Native search bounds the storage result before applying its display sort.
	if limit > 0 && len(rows) > limit {
		rows = rows[:limit]
	}
	issues := queryPageIssues(rows)
	workapi.SortIssues(issues, sortBy, reverse)
	if graphDiscoveryWantsJSON(cmd) {
		ordered := make([]*types.IssueWithCounts, 0, len(issues))
		byID := make(map[string]*types.IssueWithCounts, len(rows))
		for _, row := range rows {
			byID[row.ID] = row
		}
		for _, issue := range issues {
			ordered = append(ordered, byID[issue.ID])
		}
		return graphDiscoveryJSON(cmd.OutOrStdout(), ordered)
	}
	if quietFlag {
		return nil
	}
	outputSearchResults(issues, query, longFormat)
	return nil
}

func graphSearchMatches(rows []*types.IssueWithCounts, queryText string) ([]*types.IssueWithCounts, error) {
	queryText = strings.ToLower(queryText)
	idLike := storageissueops.LooksLikeIssueID(queryText)
	like, err := graphSearchMatcher(queryText)
	if err != nil {
		return nil, err
	}
	result := make([]*types.IssueWithCounts, 0, len(rows))
	for _, row := range rows {
		id := strings.ToLower(row.ID)
		matched := like(row.Title)
		if idLike {
			matched = matched || strings.HasPrefix(id, queryText)
		} else {
			matched = matched || like(id)
		}
		if idLike && !matched && row.ExternalRef != nil {
			matched = like(*row.ExternalRef)
		}
		if matched {
			result = append(result, row)
		}
	}
	return result, nil
}

// Native text search uses SQL LIKE. Preserve its wildcard behavior on graph
// canonical IDs after the checked read has replaced backing IDs.
func graphSearchMatcher(pattern string) (func(string) bool, error) {
	pattern = strings.ToLower(pattern)
	if !strings.ContainsAny(pattern, "%_") {
		return func(value string) bool { return strings.Contains(strings.ToLower(value), pattern) }, nil
	}
	var expression strings.Builder
	expression.WriteString("(?s)")
	for _, char := range pattern {
		switch char {
		case '%':
			expression.WriteString(".*")
		case '_':
			expression.WriteByte('.')
		default:
			expression.WriteString(regexp.QuoteMeta(string(char)))
		}
	}
	compiled, err := regexp.Compile(expression.String())
	if err != nil {
		return nil, err
	}
	return func(value string) bool { return compiled.MatchString(strings.ToLower(value)) }, nil
}

func graphDiscoverySortSupported(sortBy string) bool {
	switch sortBy {
	case "", "priority", "created", "updated", "closed", "status", "id", "title", "type", "assignee":
		return true
	}
	return false
}

func runGraphPreviewCount(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		return graphFailure("invalid_selector", "count does not accept positional arguments", 2)
	}
	if err := graphDiscoveryFlags(cmd, "status", "priority", "assignee", "type", "label", "label-any", "title", "id", "parent", "no-parent", "exclude-type", "exclude-status", "title-contains", "desc-contains", "notes-contains", "created-after", "created-before", "updated-after", "updated-before", "closed-after", "closed-before", "empty-description", "no-assignee", "no-labels", "priority-min", "priority-max", "by-status", "by-priority", "by-type", "by-assignee", "by-label"); err != nil {
		return err
	}
	selectedGroups := 0
	for _, name := range []string{"by-status", "by-priority", "by-type", "by-assignee", "by-label"} {
		selected, _ := cmd.Flags().GetBool(name)
		if selected {
			selectedGroups++
		}
	}
	if selectedGroups > 1 {
		return graphFailure("invalid_selector", "only one --by-* flag can be specified", 2)
	}
	for _, name := range []string{"created-after", "created-before", "updated-after", "updated-before", "closed-after", "closed-before"} {
		raw, _ := cmd.Flags().GetString(name)
		if raw == "" {
			continue
		}
		if _, err := parseTimeFlag(raw); err != nil {
			return graphFailure("invalid_selector", "parsing --"+name+": "+err.Error(), 2)
		}
	}
	parentSelector, _ := cmd.Flags().GetString("parent")
	noParentRequested, _ := cmd.Flags().GetBool("no-parent")
	if parentSelector != "" && noParentRequested {
		return graphFailure("invalid_selector", "--parent and --no-parent are mutually exclusive", 2)
	}
	request, group, err := parseCountRequest(cmd)
	if err != nil {
		return err // ordinary parser already emitted this diagnostic
	}
	for _, value := range []string{request.Status, request.IssueType, request.Assignee, request.TitleSearch, request.TitleContains, request.DescContains, request.NotesContains} {
		if err := graphDiscoveryText(value, "count filter"); err != nil {
			return err
		}
	}
	if len(request.Labels)+len(request.LabelsAny) > graphstore.PreviewIssueListLabelLimit {
		return graphFailure("invalid_selector", "too many count labels", 2)
	}
	for _, label := range append(append([]string{}, request.Labels...), request.LabelsAny...) {
		if err := graphDiscoveryText(label, "count label"); err != nil {
			return err
		}
	}
	filter, err := workapi.BuildCountFilter(request, workapi.ListConfig{})
	if err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	filter.SkipWisps = true
	// Resolve bare native IDs through the checked catalog. Explicit graph paths
	// remain exact, and an ambiguous bare ID fails before the count is printed.
	selectors := make([]graphstore.IssueSelector, 0, len(filter.IDs)+1)
	for _, id := range filter.IDs {
		selected, err := graphPreviewIssueSelector(id)
		if err != nil {
			return graphFailure("invalid_selector", err.Error(), 2)
		}
		selectors = append(selectors, selected)
	}
	idCount := len(selectors)
	filter.IDs = nil
	if filter.ParentID != nil {
		selected, err := graphPreviewIssueSelector(*filter.ParentID)
		if err != nil {
			return graphFailure("invalid_selector", err.Error(), 2)
		}
		selectors = append(selectors, selected)
		filter.ParentID = nil
	}
	noParent := filter.NoParent
	filter.NoParent = false
	rows, paths, err := graphCountDiscoveryRows(filter, selectors)
	if err != nil {
		return err
	}
	ids := make(map[string]bool, idCount)
	for _, path := range paths[:idCount] {
		ids[path] = true
	}
	parent := ""
	if len(paths) > idCount {
		parent = paths[idCount]
	}
	rows = graphDiscoverySelectRows(rows, ids, parent, noParent)
	if group == "" {
		if graphDiscoveryWantsJSON(cmd) {
			return graphDiscoveryJSON(cmd.OutOrStdout(), struct {
				Count int64 `json:"count"`
			}{int64(len(rows))})
		}
		_, err = fmt.Fprintln(cmd.OutOrStdout(), len(rows))
		return err
	}
	type groupCount struct {
		Group string `json:"group"`
		Count int    `json:"count"`
	}
	counts := map[string]int{}
	for _, row := range rows {
		switch group {
		case issueops.CountGroupStatus:
			counts[string(row.Status)]++
		case issueops.CountGroupPriority:
			counts[fmt.Sprintf("P%d", row.Priority)]++
		case issueops.CountGroupType:
			counts[string(row.IssueType)]++
		case issueops.CountGroupAssignee:
			assignee := row.Assignee
			if assignee == "" {
				assignee = "(unassigned)"
			}
			counts[assignee]++
		case issueops.CountGroupLabel:
			for _, label := range row.Labels {
				counts[label]++
			}
		}
	}
	groups := make([]groupCount, 0, len(counts))
	for label, count := range counts {
		groups = append(groups, groupCount{label, count})
	}
	slices.SortFunc(groups, func(a, b groupCount) int { return strings.Compare(a.Group, b.Group) })
	if graphDiscoveryWantsJSON(cmd) {
		return graphDiscoveryJSON(cmd.OutOrStdout(), struct {
			Total  int64        `json:"total"`
			Groups []groupCount `json:"groups"`
		}{int64(len(rows)), groups})
	}
	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Total: %d\n\n", len(rows)); err != nil {
		return err
	}
	for _, group := range groups {
		if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s: %d\n", group.Group, group.Count); err != nil {
			return err
		}
	}
	return nil
}

// Explicit scope keeps selector parsing independent of command admission state.
func graphDiscoveryIssuePath(scopeURL, selector string) (string, error) {
	path, err := graphPreviewResourcePath(scopeURL, selector)
	if err != nil || !strings.HasPrefix(path, "beads/") {
		return "", graphFailure("invalid_selector", fmt.Sprintf("invalid graph Issue selector %q", selector), 2)
	}
	return path, nil
}

// The edge wins over dotted-ID inference, just as native ParentID does.
func graphDiscoveryParentMatch(row *types.IssueWithCounts, parent string) bool {
	if row.Parent != nil {
		return *row.Parent == parent
	}
	return strings.HasPrefix(row.ID, parent+".")
}

func graphDiscoverySelectRows(rows []*types.IssueWithCounts, ids map[string]bool, parent string, noParent bool) []*types.IssueWithCounts {
	selected := make([]*types.IssueWithCounts, 0, len(rows))
	for _, row := range rows {
		if len(ids) > 0 && !ids[row.ID] || parent != "" && !graphDiscoveryParentMatch(row, parent) || noParent && row.Parent != nil {
			continue
		}
		selected = append(selected, row)
	}
	return selected
}

func runGraphPreviewQuery(cmd *cobra.Command, args []string) error {
	if err := graphDiscoveryFlags(cmd, "limit", "offset", "all", "long", "sort", "reverse", "parse-only"); err != nil {
		return err
	}
	if len(args) == 0 {
		return graphFailure("invalid_selector", "query expression is required", 2)
	}
	if offset, _ := cmd.Flags().GetInt("offset"); offset < 0 {
		return graphFailure("invalid_selector", "--offset must be non-negative", 2)
	}
	input, err := gatherQueryInput(cmd, args)
	if err != nil {
		return err
	}
	if err := graphDiscoveryText(input.expression, "query expression"); err != nil {
		return err
	}
	if input.parseOnly {
		return printParsedQuery(input.expression)
	}
	if !graphDiscoverySortSupported(input.request.SortBy) {
		return graphFailure("capability_unavailable", "graph query does not support --sort "+input.request.SortBy, 5)
	}
	if graphQueryUnsupportedField(input.expression) {
		return graphFailure("capability_unavailable", "graph query does not support mol_type or ephemeral fields", 5)
	}
	plan, err := workapi.BuildQueryPlan(input.request)
	if err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	node, err := query.Parse(input.expression)
	if err != nil {
		return graphFailure("invalid_selector", "invalid query expression: "+err.Error(), 2)
	}
	predicate, err := graphQueryPredicate(node, time.Now(), graphPreviewConfig.GraphScopeURL)
	if err != nil {
		return graphFailure("invalid_selector", "invalid query expression: "+err.Error(), 2)
	}
	filter := graphDiscoveryFilter()
	filter.ExcludeStatus = plan.Filter.ExcludeStatus
	rows, err := graphDiscoveryRows("", filter)
	if err != nil {
		return err
	}
	selected := make([]*types.IssueWithCounts, 0, len(rows))
	for _, row := range rows {
		if predicate(row) {
			selected = append(selected, row)
		}
	}
	rows = selected
	items, more := workapi.FinishPageAt(rows, plan.SortBy, plan.Reverse, plan.Offset, plan.Limit, false)
	if graphDiscoveryWantsJSON(cmd) {
		if err := graphDiscoveryJSON(cmd.OutOrStdout(), items); err != nil {
			return err
		}
	} else if !quietFlag {
		outputQueryResults(queryPageIssues(items), input.expression, input.longFormat)
	}
	printTruncationHint(more, plan.Limit)
	return nil
}

func graphQueryUnsupportedField(expression string) bool {
	node, err := query.Parse(expression)
	if err != nil {
		return false // BuildQueryPlan reports the syntax error.
	}
	var unsupported func(query.Node) bool
	unsupported = func(node query.Node) bool {
		switch n := node.(type) {
		case *query.ComparisonNode:
			return n.Field == "mol_type" || n.Field == "ephemeral"
		case *query.AndNode:
			return unsupported(n.Left) || unsupported(n.Right)
		case *query.OrNode:
			return unsupported(n.Left) || unsupported(n.Right)
		case *query.NotNode:
			return unsupported(n.Operand)
		default:
			return false
		}
	}
	return unsupported(node)
}

func graphQueryPredicate(node query.Node, now time.Time, scopeURL string) (func(*types.IssueWithCounts) bool, error) {
	switch n := node.(type) {
	case *query.ComparisonNode:
		if n.Field == "parent" {
			if n.Op != query.OpEquals {
				return nil, fmt.Errorf("parent only supports = operator")
			}
			parent, err := graphDiscoveryIssuePath(scopeURL, n.Value)
			if err != nil {
				return nil, err
			}
			return func(row *types.IssueWithCounts) bool { return graphDiscoveryParentMatch(row, parent) }, nil
		}
	case *query.AndNode:
		left, err := graphQueryPredicate(n.Left, now, scopeURL)
		if err != nil {
			return nil, err
		}
		right, err := graphQueryPredicate(n.Right, now, scopeURL)
		if err != nil {
			return nil, err
		}
		return func(row *types.IssueWithCounts) bool { return left(row) && right(row) }, nil
	}
	// Native BuildQueryPlan has already refused OR/NOT forms that require an
	// unsupported parent predicate. Its remaining subtrees use native rules.
	native, err := query.PredicateNodeAt(node, now)
	if err != nil {
		return nil, err
	}
	return func(row *types.IssueWithCounts) bool { return native(row.Issue) }, nil
}
