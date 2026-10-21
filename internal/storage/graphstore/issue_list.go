package graphstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/storage/issueops"
	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/internal/utils"
	"github.com/steveyegge/beads/internal/workapi"
	publicops "github.com/steveyegge/beads/issueops"
)

// These are acquisition limits for the preview's database-backed list policy,
// not new configuration or schema constraints. Workspace YAML is loaded by the
// existing frontend configuration machinery before this adapter runs.
const (
	PreviewIssueListRequestByteLimit = 64 << 10
	PreviewIssueListLabelLimit       = 256
	PreviewIssueListConfigByteLimit  = 64 << 10
	PreviewIssueListConfigRowLimit   = 256
)

// IssueListPage contains complete canonical Issue records. HasMore reports an
// intentional page limit, not a snapshot cursor or a complete BDP collection.
// It does not promise the legacy IssueWithCounts JSON representation.
type IssueListPage struct {
	Items   []IssueRecord `json:"items"`
	HasMore bool          `json:"hasMore"`
}

// ListIssues uses the existing Issue query policy inside the same read
// transaction as authority checks and complete canonical record validation.
// No legacy store, advisory wakeup, mutation or cross-request cursor is opened.
func (s *Store) ListIssues(ctx context.Context, request publicops.ListRequest) (IssueListPage, error) {
	return s.listIssues(ctx, request, nil)
}

// ListIssuesSelected resolves a graph parent in the same checked snapshot as
// the native Issue query. The request must not also name a native parent.
func (s *Store) ListIssuesSelected(ctx context.Context, request publicops.ListRequest, parent IssueSelector) (IssueListPage, error) {
	if err := validateIssueSelector(parent); err != nil {
		return IssueListPage{}, err
	}
	return s.listIssues(ctx, request, &parent)
}

func (s *Store) listIssues(ctx context.Context, request publicops.ListRequest, parent *IssueSelector) (IssueListPage, error) {
	request, err := prepareIssueListRequest(request)
	if err != nil {
		return IssueListPage{}, err
	}
	if request.ReadyFlag {
		s.wakeExpiredDefersAdvisory(ctx)
	}
	var result IssueListPage
	err = s.withTx(ctx, false, func(tx *sql.Tx) error {
		var err error
		result, err = s.listIssuesWithParentInTx(ctx, tx, &request, parent)
		return err
	})
	if err != nil {
		return IssueListPage{}, err
	}
	return result, nil
}

func prepareIssueListRequest(in publicops.ListRequest) (publicops.ListRequest, error) {
	// Reject dropped ready filters before a graph selector can be normalized
	// to an empty native ID set. Otherwise --ready --id missing would appear
	// to succeed with an empty page without checking the unsupported pairing.
	if err := publicops.ValidateReadyFlagScope(in); err != nil {
		return publicops.ListRequest{}, fmt.Errorf("%w: %v", storage.ErrValidation, err)
	}
	allowed := publicops.ListRequest{
		Status: in.Status, IssueType: in.IssueType, IDFilter: in.IDFilter, TitleSearch: in.TitleSearch, TitleContains: in.TitleContains,
		Assignee: in.Assignee, NoAssignee: in.NoAssignee,
		DueBefore: in.DueBefore, DueAfter: in.DueAfter, OverdueFlag: in.OverdueFlag,
		Labels: in.Labels, LabelsAny: in.LabelsAny, ExcludeLabels: in.ExcludeLabels,
		Priority: in.Priority, PriorityMin: in.PriorityMin, PriorityMax: in.PriorityMax,
		PinnedFlag: in.PinnedFlag, NoPinnedFlag: in.NoPinnedFlag, NoParent: in.NoParent, AllFlag: in.AllFlag,
		ReadyFlag: in.ReadyFlag,
		SortBy:    in.SortBy, Reverse: in.Reverse, Limit: in.Limit, MaxRows: in.MaxRows, MaxRowsSource: in.MaxRowsSource,
	}
	if !reflect.DeepEqual(in, allowed) {
		return publicops.ListRequest{}, fmt.Errorf("%w: unsupported graph Issue list option", ErrCapabilityUnavailable)
	}
	switch in.SortBy {
	case "", "priority", "created", "updated", "closed", "title", "status", "type", "id", "assignee":
	default:
		return publicops.ListRequest{}, fmt.Errorf("%w: unsupported graph Issue list sort", ErrCapabilityUnavailable)
	}
	if in.PinnedFlag && in.NoPinnedFlag {
		return publicops.ListRequest{}, fmt.Errorf("%w: pinned and no-pinned are mutually exclusive", storage.ErrValidation)
	}
	if (in.Limit != nil && *in.Limit < 0) || in.MaxRows < 0 {
		return publicops.ListRequest{}, fmt.Errorf("%w: list limits cannot be negative", storage.ErrValidation)
	}
	for _, p := range []*int{in.Priority, in.PriorityMin, in.PriorityMax} {
		if p != nil && (*p < 0 || *p > 4) {
			return publicops.ListRequest{}, fmt.Errorf("%w: priority must be between 0 and 4", storage.ErrValidation)
		}
	}
	requestBytes := 0
	for _, value := range []string{in.Status, in.IssueType, in.IDFilter, in.TitleSearch, in.TitleContains, in.MaxRowsSource, in.Assignee} {
		if len(value) > PreviewIssueListRequestByteLimit-requestBytes {
			return publicops.ListRequest{}, fmt.Errorf("%w: Issue list request text exceeds %d bytes", storage.ErrValidation, PreviewIssueListRequestByteLimit)
		}
		requestBytes += len(value)
		if !utf8.ValidString(value) {
			return publicops.ListRequest{}, fmt.Errorf("%w: Issue list text must be UTF-8", storage.ErrValidation)
		}
	}
	labelCount := 0
	for _, labels := range [][]string{in.Labels, in.LabelsAny, in.ExcludeLabels} {
		if len(labels) > PreviewIssueListLabelLimit-labelCount {
			return publicops.ListRequest{}, fmt.Errorf("%w: Issue list request exceeds %d labels", storage.ErrValidation, PreviewIssueListLabelLimit)
		}
		labelCount += len(labels)
		for _, label := range labels {
			if len(label) > PreviewIssueListRequestByteLimit-requestBytes {
				return publicops.ListRequest{}, fmt.Errorf("%w: Issue list request text exceeds %d bytes", storage.ErrValidation, PreviewIssueListRequestByteLimit)
			}
			requestBytes += len(label)
			if !utf8.ValidString(label) {
				return publicops.ListRequest{}, fmt.Errorf("%w: Issue list labels must be UTF-8", storage.ErrValidation)
			}
		}
	}
	copyInt := func(value *int) *int {
		if value == nil {
			return nil
		}
		copied := *value
		return &copied
	}
	allowed.Priority, allowed.PriorityMin, allowed.PriorityMax = copyInt(in.Priority), copyInt(in.PriorityMin), copyInt(in.PriorityMax)
	allowed.Limit = copyInt(in.Limit)
	allowed.Labels, allowed.LabelsAny, allowed.ExcludeLabels = slices.Clone(in.Labels), slices.Clone(in.LabelsAny), slices.Clone(in.ExcludeLabels)
	// Native query binds strict bounds as whole-second RFC3339 strings. Preserve
	// that truncation (rather than write rounding), but normalize API offsets
	// before formatting so comparisons use the same UTC instant as stored dates.
	for _, field := range []**time.Time{&allowed.DueBefore, &allowed.DueAfter} {
		if *field != nil {
			value := (*field).UTC().Truncate(time.Second)
			if value.Year() < 1 || value.Year() > 9999 {
				return publicops.ListRequest{}, fmt.Errorf("%w: Issue due filter year must be between 1 and 9999", storage.ErrValidation)
			}
			*field = &value
		}
	}
	return allowed, nil
}

func (s *Store) listIssuesInTx(ctx context.Context, tx *sql.Tx, request publicops.ListRequest) (IssueListPage, error) {
	return s.listIssuesWithParentInTx(ctx, tx, &request, nil)
}

func (s *Store) listIssuesWithParentInTx(ctx context.Context, tx *sql.Tx, request *publicops.ListRequest, parent *IssueSelector) (IssueListPage, error) {
	if err := checkCurrentReadBytes(ctx, tx); err != nil {
		return IssueListPage{}, err
	}
	if err := checkBinding(ctx, tx, s.options); err != nil {
		return IssueListPage{}, err
	}
	if err := s.checkCollectionMappingsInTx(ctx, tx); err != nil {
		return IssueListPage{}, err
	}
	var count int
	if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM graph_preview_catalog WHERE allocation_state='live'").Scan(&count); err != nil {
		return IssueListPage{}, err
	}
	if count > PreviewSnapshotLimit {
		return IssueListPage{}, fmt.Errorf("%w: graph Issue list exceeds %d live Resources before filtering", ErrLimitExceeded, PreviewSnapshotLimit)
	}
	if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM graph_preview_catalog WHERE allocation_state NOT IN ('live','deleted')").Scan(&count); err != nil {
		return IssueListPage{}, err
	}
	if count != 0 {
		return IssueListPage{}, fmt.Errorf("%w: unknown allocation state", ErrInvalidStore)
	}
	if err := s.checkIssueListCatalogInTx(ctx, tx); err != nil {
		return IssueListPage{}, err
	}
	// Resolve visible graph paths and shipping native IDs against the same
	// checked snapshot that runs the Issue query. A bare name that collides
	// with another Bead's path remains ambiguous, as it is for show.
	noMatchingIDs := false
	if request.IDFilter != "" && len(utils.NormalizeLabels(strings.Split(request.IDFilter, ","))) > 0 {
		ids, err := s.resolveIssueListIDsInTx(ctx, tx, request.IDFilter)
		if err != nil {
			return IssueListPage{}, err
		}
		noMatchingIDs = len(ids) == 0
		if !noMatchingIDs {
			request.IDFilter = strings.Join(ids, ",")
		}
	} else {
		request.IDFilter = ""
	}
	var parentNativeID, parentPath string
	if parent != nil {
		var err error
		parentNativeID, parentPath, err = s.resolveIssueListParentInTx(ctx, tx, *parent)
		if err != nil {
			return IssueListPage{}, err
		}
	}
	cfg, err := issueListConfigInTx(ctx, tx)
	if err != nil {
		return IssueListPage{}, err
	}
	if request.IssueType != "" && cfg.IsInfra(request.IssueType) {
		return IssueListPage{}, fmt.Errorf("%w: graph Issue list does not admit infrastructure or wisp scope", ErrCapabilityUnavailable)
	}
	filter, err := workapi.BuildListFilter(*request, cfg)
	if err != nil {
		return IssueListPage{}, fmt.Errorf("%w: %v", storage.ErrValidation, err)
	}
	if !filter.SkipWisps || (filter.Ephemeral != nil && *filter.Ephemeral) {
		return IssueListPage{}, fmt.Errorf("%w: graph Issue list requires durable scope", ErrCapabilityUnavailable)
	}
	if noMatchingIDs {
		return IssueListPage{Items: []IssueRecord{}}, nil
	}
	// Counts are not in this result's projection. Labels and complete Issue
	// properties remain present; callers cannot request partial graph records.
	filter.SkipCounts = true
	if parent != nil && request.ReadyFlag && parentNativeID != "" {
		// Shipping --ready --parent selects transitive parent-child descendants.
		// The ready query computes that set before applying blocker/status filters,
		// so a closed intermediary does not hide a ready grandchild. Keep the
		// graph selector resolution and the native query in this transaction.
		filter.ParentID = &parentNativeID
	}
	// Every admitted workspace has <=1000 live Resources. Bounding an enormous
	// requested SQL limit here cannot discard an admitted row, and prevents the
	// shared probe-row increment from overflowing an arbitrary caller's int.
	if parent != nil || filter.Limit > PreviewSnapshotLimit {
		filter.Limit = PreviewSnapshotLimit
	}
	if parent == nil {
		filter = workapi.WithFetchOneExtra(workapi.WithRowsBeforeThePage(filter, request.Offset))
	} else {
		// The parent membership check uses graph-visible IDs after the
		// bounded native scan. Apply the caller's row cap to the selected page,
		// not the larger candidate set.
		filter.MaxRows = 0
	}
	rows, err := queryIssueListRowsInTx(ctx, tx, filter, request.ReadyFlag)
	if err != nil {
		var capError *issueops.ErrTooManyRows
		if errors.As(err, &capError) {
			return IssueListPage{}, fmt.Errorf("%w: %w", ErrLimitExceeded, err)
		}
		return IssueListPage{}, err
	}
	records := make(map[string]IssueRecord, len(rows))
	selectedRows := make([]*types.IssueWithCounts, 0, len(rows))
	for _, row := range rows {
		if row == nil || row.Issue == nil {
			return IssueListPage{}, fmt.Errorf("%w: nil Issue query row", ErrInvalidStore)
		}
		if _, duplicate := records[row.ID]; duplicate {
			return IssueListPage{}, fmt.Errorf("%w: duplicate Issue query row", ErrInvalidStore)
		}
		var path string
		if err := tx.QueryRowContext(ctx, "SELECT path FROM graph_preview_catalog WHERE backing='issue' AND backing_key=? AND allocation_state='live'", row.ID).Scan(&path); err != nil {
			return IssueListPage{}, fmt.Errorf("%w: unmapped listed Issue: %v", ErrInvalidStore, err)
		}
		if err := validatePath(path); err != nil {
			return IssueListPage{}, fmt.Errorf("%w: malformed Issue allocation path: %v", ErrInvalidStore, err)
		}
		record, err := s.showIssueInTx(ctx, tx, path)
		if err != nil {
			return IssueListPage{}, err
		}
		records[row.ID] = record
		if parent == nil || (request.ReadyFlag && parentNativeID != "") || (!request.ReadyFlag && issueListParentMatches(row, path, parentNativeID, parentPath)) {
			selectedRows = append(selectedRows, row)
		}
	}
	rows = selectedRows
	// Validate the over-fetched probe before trimming it: HasMore must never
	// conceal a malformed queried record. Preserve the shared query/page order.
	sortBy, reverse := request.SortBy, request.Reverse
	if sortBy == "id" {
		// Native Issue IDs are backing keys, not the graph IDs users selected.
		// Sort the complete checked set by canonical paths before page trimming.
		slices.SortStableFunc(rows, func(a, b *types.IssueWithCounts) int {
			left := strings.TrimPrefix(records[a.ID].ID, s.ScopeURL())
			right := strings.TrimPrefix(records[b.ID].ID, s.ScopeURL())
			order := utils.NaturalCompareIDs(left, right)
			if reverse {
				return -order
			}
			return order
		})
		sortBy, reverse = "", false
	}
	rows, more := workapi.FinishPageAt(rows, sortBy, reverse, request.Offset, workapi.PageLimit(*request), false)
	result := IssueListPage{Items: make([]IssueRecord, 0, len(rows)), HasMore: more}
	for _, row := range rows {
		result.Items = append(result.Items, records[row.ID])
	}
	if parent != nil {
		if err := issueops.EnforceMaxRowsCap(len(result.Items), request.MaxRows, request.MaxRowsSource); err != nil {
			return IssueListPage{}, fmt.Errorf("%w: %w", ErrLimitExceeded, err)
		}
	}
	return result, nil
}

// queryIssueListRowsInTx keeps the blocker-aware ready predicate in the same
// checked snapshot as graph mapping and page validation. The ordinary list
// path and its legacy counted projection use this same query choice.
func queryIssueListRowsInTx(ctx context.Context, tx *sql.Tx, filter types.IssueFilter, ready bool) ([]*types.IssueWithCounts, error) {
	if !ready {
		return issueops.SearchIssuesWithCountsInTx(ctx, tx, "", filter)
	}
	work := workapi.ReadyFilterFromIssueFilter(filter)
	if work.Limit == 0 {
		// Natural ID sorting needs the complete candidate set before the
		// visible graph IDs can be ordered. Bound acquisition by the graph
		// snapshot limit plus one so overage is detected, not truncated.
		work.Limit = PreviewSnapshotLimit + 1
	}
	rows, err := issueops.GetReadyWorkWithCountsInTx(ctx, tx, work)
	if err != nil {
		return nil, err
	}
	if len(rows) > PreviewSnapshotLimit {
		return nil, fmt.Errorf("%w: graph Issue ready list exceeds %d rows", ErrLimitExceeded, PreviewSnapshotLimit)
	}
	return rows, nil
}

// resolveIssueListIDsInTx lowers the shipping --id set to native Issue keys.
// Explicit graph paths remain exact; bare IDs use the same collision rule as
// show. Missing and non-Issue selections contribute no matching Issue.
func (s *Store) resolveIssueListIDsInTx(ctx context.Context, tx *sql.Tx, raw string) ([]string, error) {
	selected := utils.NormalizeLabels(strings.Split(raw, ","))
	ids := make([]string, 0, len(selected))
	seen := make(map[string]bool, len(selected))
	for _, value := range selected {
		selector := IssueSelector{}
		switch {
		case strings.HasPrefix(value, s.ScopeURL()):
			selector.Path = strings.TrimPrefix(value, s.ScopeURL())
		case strings.HasPrefix(value, "beads/"):
			selector.Path = value
		default:
			selector.Path, selector.NativeID = "beads/"+value, value
		}
		if err := validateIssueSelector(selector); err != nil {
			return nil, fmt.Errorf("%w: invalid Issue ID %q: %v", storage.ErrValidation, value, err)
		}
		_, issue, err := s.resolveLiveIssueSelectorInTx(ctx, tx, selector)
		switch {
		case errors.Is(err, ErrNotFound), errors.Is(err, errNonIssueSelector):
			continue
		case err != nil:
			return nil, err
		}
		if issue.Properties == nil || issue.Properties.ID == "" {
			return nil, fmt.Errorf("%w: incomplete selected Issue mapping", ErrInvalidStore)
		}
		if !seen[issue.Properties.ID] {
			seen[issue.Properties.ID] = true
			ids = append(ids, issue.Properties.ID)
		}
	}
	return ids, nil
}

func (s *Store) resolveIssueListParentInTx(ctx context.Context, tx *sql.Tx, parent IssueSelector) (string, string, error) {
	path, issue, err := s.resolveLiveIssueSelectorInTx(ctx, tx, parent)
	switch {
	case err == nil:
		if issue.Properties == nil || issue.Properties.ID == "" {
			return "", "", fmt.Errorf("%w: incomplete parent Issue mapping", ErrInvalidStore)
		}
		return issue.Properties.ID, path, nil
	case errors.Is(err, ErrNotFound):
		// The ordinary filter also matches edge-less dotted descendants of a
		// bare ID, even when that parent has no current Issue allocation.
		if parent.NativeID != "" {
			return parent.NativeID, parent.Path, nil
		}
		return "", parent.Path, nil
	case errors.Is(err, errNonIssueSelector):
		return "", parent.Path, nil
	default:
		return "", "", err
	}
}

func issueListParentMatches(row *types.IssueWithCounts, path, parentNativeID, parentPath string) bool {
	if row.Parent != nil {
		return parentNativeID != "" && *row.Parent == parentNativeID
	}
	return strings.HasPrefix(path, parentPath+".")
}

type issueListConfigSource struct{ config workapi.ListConfig }

// ListCustomStatuses reads the same resolved Issue status configuration as
// graph Issue list. It does not open the legacy store or modify the workspace.
func (s *Store) ListCustomStatuses(ctx context.Context) ([]types.CustomStatus, error) {
	var statuses []types.CustomStatus
	err := s.withTx(ctx, false, func(tx *sql.Tx) error {
		if err := checkBinding(ctx, tx, s.options); err != nil {
			return err
		}
		cfg, err := issueListConfigInTx(ctx, tx)
		if err != nil {
			return err
		}
		statuses = cfg.CustomStatuses
		return nil
	})
	if err != nil {
		return nil, err
	}
	return statuses, nil
}

func (s issueListConfigSource) GetCustomStatuses(context.Context) ([]types.CustomStatus, error) {
	return s.config.CustomStatuses, nil
}
func (s issueListConfigSource) GetCustomTypes(context.Context) ([]string, error) {
	return s.config.CustomTypes, nil
}
func (s issueListConfigSource) GetInfraTypes(context.Context) (map[string]bool, error) {
	return s.config.InfraSet, nil
}

func issueListConfigInTx(ctx context.Context, tx *sql.Tx) (workapi.ListConfig, error) {
	var totalRows, totalBytes uint64
	// Package-owned SQL only. Clamp lengths in SQL before reading scalar counts;
	// do not acquire a corrupt huge config value and check its size afterwards.
	for _, query := range []string{
		"SELECT COUNT(*), CAST(LEAST(COALESCE(SUM(OCTET_LENGTH(name)+OCTET_LENGTH(category)),0),?) AS UNSIGNED) FROM custom_statuses",
		"SELECT COUNT(*), CAST(LEAST(COALESCE(SUM(OCTET_LENGTH(name)),0),?) AS UNSIGNED) FROM custom_types",
		"SELECT COUNT(*), CAST(LEAST(COALESCE(SUM(OCTET_LENGTH(`key`)+OCTET_LENGTH(value)),0),?) AS UNSIGNED) FROM config WHERE `key` IN ('status.custom','types.custom','types.infra')",
	} {
		var rows, size uint64
		if err := tx.QueryRowContext(ctx, query, PreviewIssueListConfigByteLimit+1).Scan(&rows, &size); err != nil {
			return workapi.ListConfig{}, err
		}
		if rows > PreviewIssueListConfigRowLimit-totalRows || size > PreviewIssueListConfigByteLimit-totalBytes {
			return workapi.ListConfig{}, fmt.Errorf("%w: graph Issue list configuration exceeds %d rows or %d bytes", ErrLimitExceeded, PreviewIssueListConfigRowLimit, PreviewIssueListConfigByteLimit)
		}
		totalRows, totalBytes = totalRows+rows, totalBytes+size
	}
	statuses, customTypes, err := issueops.ResolveCustomConfigStrictInTx(ctx, tx)
	if err != nil {
		return workapi.ListConfig{}, err
	}
	infra, err := issueops.ResolveInfraTypesStrictInTx(ctx, tx)
	if err != nil {
		return workapi.ListConfig{}, err
	}
	cfg, err := workapi.LoadListConfig(ctx, issueListConfigSource{workapi.ListConfig{CustomStatuses: statuses, CustomTypes: customTypes, InfraSet: infra}})
	if err != nil {
		return workapi.ListConfig{}, err
	}
	if err := checkResolvedIssueListConfig(cfg); err != nil {
		return workapi.ListConfig{}, err
	}
	return cfg, nil
}

// Catalog admission covers every live allocation, including filtered-out Issues
// and non-Issue records. It reads metadata/existence only: complete payload and
// retained-state verification remains scoped to queried Issues and their owned
// Links, including the query's probe row.
func (s *Store) checkIssueListCatalogInTx(ctx context.Context, tx *sql.Tx) error {
	issueType, err := installedIssueTypeURLInTx(ctx, tx, s.ScopeURL())
	if err != nil {
		return err
	}
	dependencyTypes := map[string]bool{}
	for _, id := range dependencyLinkTypesForIssueType(s.ScopeURL(), issueType) {
		dependencyTypes[id] = true
	}
	ids, err := s.installedInformationalTypes(ctx, tx)
	if err != nil {
		return err
	}
	informational := make(map[string]bool, len(ids))
	for _, id := range ids {
		informational[id] = true
	}
	rows, err := tx.QueryContext(ctx, `SELECT c.path,c.resource_kind,c.type_url,c.revision,c.backing,c.backing_key,
 p.path IS NOT NULL,i.id IS NOT NULL,d.id IS NOT NULL,l.path IS NOT NULL
 FROM graph_preview_catalog c
 LEFT JOIN graph_preview_payloads p ON p.path=c.path
 LEFT JOIN issues i ON c.backing='issue' AND i.id=c.backing_key
 LEFT JOIN dependencies d ON c.backing='dependency' AND d.id=c.backing_key
 LEFT JOIN graph_preview_links l ON l.path=c.path
 WHERE c.allocation_state='live' LIMIT ?`, PreviewSnapshotLimit+1)
	if err != nil {
		return err
	}
	var readErr error
	count := 0
	for rows.Next() {
		var path, kind, typ, revision, backing string
		var key sql.NullString
		var memoryExists, issueExists, dependencyExists, informationalExists bool
		if err := rows.Scan(&path, &kind, &typ, &revision, &backing, &key, &memoryExists, &issueExists, &dependencyExists, &informationalExists); err != nil {
			readErr = err
			break
		}
		count++
		if count > PreviewSnapshotLimit {
			readErr = fmt.Errorf("%w: graph Issue list live Resource limit exceeded", ErrLimitExceeded)
			break
		}
		valid := validateResourcePath(path) == nil && authorityID.MatchString(revision)
		switch backing {
		case "generic":
			valid = valid && kind == "bead" && strings.HasPrefix(path, "beads/") && s.isMemoryType(typ) && !key.Valid && memoryExists
		case "issue":
			valid = valid && kind == "bead" && strings.HasPrefix(path, "beads/") && typ == issueType && key.Valid && key.String != "" && issueExists
		case "dependency":
			valid = valid && kind == "link" && strings.HasPrefix(path, "links/") && dependencyTypes[typ] && key.Valid && key.String != "" && dependencyExists
		case "informational":
			valid = valid && kind == "link" && strings.HasPrefix(path, "links/") && informational[typ] && !key.Valid && informationalExists
		default:
			valid = false
		}
		if !valid {
			readErr = fmt.Errorf("%w: invalid live graph Issue list allocation", ErrInvalidStore)
			break
		}
	}
	return errors.Join(readErr, rows.Err(), rows.Close())
}

// checkResolvedIssueListConfig also charges frontend YAML fallbacks. This is a
// resolved-policy bound, not a limit on the frontend's earlier YAML acquisition.
// Each stored entry is charged, including duplicates retained in slices.
func checkResolvedIssueListConfig(cfg workapi.ListConfig) error {
	entries, bytes := 0, 0
	charge := func(parts ...string) error {
		if entries == PreviewIssueListConfigRowLimit {
			return fmt.Errorf("%w: resolved Issue list configuration exceeds %d entries", ErrLimitExceeded, PreviewIssueListConfigRowLimit)
		}
		entries++
		for _, part := range parts {
			if len(part) > PreviewIssueListConfigByteLimit-bytes {
				return fmt.Errorf("%w: resolved Issue list configuration exceeds %d bytes", ErrLimitExceeded, PreviewIssueListConfigByteLimit)
			}
			bytes += len(part)
		}
		return nil
	}
	for _, status := range cfg.CustomStatuses {
		if err := charge(status.Name, string(status.Category)); err != nil {
			return err
		}
	}
	for _, name := range cfg.CustomTypes {
		if err := charge(name); err != nil {
			return err
		}
	}
	for name := range cfg.InfraSet {
		if err := charge(name); err != nil {
			return err
		}
	}
	return nil
}
