package graphstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/utils"
)

// IssueSelector keeps the explicit graph path separate from a bare native
// Issue ID. NativeID is set only for bare CLI spelling, never for beads/PATH
// or a canonical URL.
type IssueSelector struct {
	Path     string
	NativeID string
}

// A non-Issue selection has a different user-visible state-read refusal from
// a bare name that ambiguously identifies an Issue and another graph Bead.
var errNonIssueSelector = errors.New("selector is not an Issue")

func nonIssueSelectorError(backing string) error {
	return fmt.Errorf("%w: selector names %s, not an Issue: %w", storage.ErrValidation, backing, errNonIssueSelector)
}

func validateIssueSelector(selector IssueSelector) error {
	if err := validatePath(selector.Path); err != nil {
		return err
	}
	if selector.NativeID != "" && selector.Path != "beads/"+selector.NativeID {
		return fmt.Errorf("%w: native Issue fallback requires its matching bare selector", storage.ErrValidation)
	}
	return nil
}

// Resolve against the checked live catalog in the caller's transaction. A
// bare spelling that also names a different graph path is ambiguous, even if
// that path is a Memory or a retained deletion. Explicit paths remain exact.
func (s *Store) resolveLiveIssueSelectorInTx(ctx context.Context, tx *sql.Tx, selector IssueSelector) (string, IssueRecord, error) {
	if selector.NativeID == "" {
		var backing string
		err := tx.QueryRowContext(ctx, `SELECT backing FROM graph_preview_catalog WHERE path=?`, selector.Path).Scan(&backing)
		if errors.Is(err, sql.ErrNoRows) {
			return "", IssueRecord{}, ErrNotFound
		}
		if err != nil {
			return "", IssueRecord{}, err
		}
		if backing != "issue" {
			return "", IssueRecord{}, nonIssueSelectorError(backing)
		}
		issue, err := s.requireIssueInTx(ctx, tx, selector.Path)
		return selector.Path, issue, err
	}
	var backing string
	err := tx.QueryRowContext(ctx, `SELECT backing FROM graph_preview_catalog WHERE path=?`, selector.Path).Scan(&backing)
	pathExists := err == nil
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", IssueRecord{}, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT path FROM graph_preview_catalog WHERE backing='issue' AND backing_key=? AND allocation_state='live' LIMIT 2`, selector.NativeID)
	if err != nil {
		return "", IssueRecord{}, err
	}
	var paths []string
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			_ = rows.Close()
			return "", IssueRecord{}, err
		}
		paths = append(paths, path)
	}
	if err := errors.Join(rows.Err(), rows.Close()); err != nil {
		return "", IssueRecord{}, err
	}
	if len(paths) > 1 {
		return "", IssueRecord{}, fmt.Errorf("%w: native Issue ID has multiple live graph allocations", ErrInvalidStore)
	}
	if len(paths) == 1 && pathExists && paths[0] != selector.Path {
		return "", IssueRecord{}, fmt.Errorf("%w: bare Issue ID %q matches both a graph path and a different native Issue; use an explicit beads/PATH", storage.ErrValidation, selector.NativeID)
	}
	path := selector.Path
	if !pathExists {
		if len(paths) == 0 {
			return "", IssueRecord{}, ErrNotFound
		}
		path = paths[0]
		if err := validatePath(path); err != nil {
			return "", IssueRecord{}, fmt.Errorf("%w: invalid native Issue graph allocation: %v", ErrInvalidStore, err)
		}
	}
	if pathExists && backing != "issue" {
		return "", IssueRecord{}, nonIssueSelectorError(backing)
	}
	issue, err := s.requireIssueInTx(ctx, tx, path)
	return path, issue, err
}

// The familiar positional bd link accepts shipping Issue abbreviations. Keep
// selection in the writer's transaction: a CLI lookup followed by AddDependency
// could link a different Issue if the live catalog changed between calls.
// Explicit beads/PATH selectors never reach this fallback.
func (s *Store) resolvePositionalIssueInTx(ctx context.Context, tx *sql.Tx, path, nativeID string) (string, IssueRecord, error) {
	if nativeID == "" {
		issue, err := s.requireIssueInTx(ctx, tx, path)
		return path, issue, err
	}
	switch strings.ToLower(strings.TrimSpace(nativeID)) {
	case "", "null", "undefined", "none", "nil":
		return "", IssueRecord{}, fmt.Errorf("%w: refusing missing-value token %q as an Issue ID", storage.ErrValidation, nativeID)
	}
	selected, issue, err := s.resolveLiveIssueSelectorInTx(ctx, tx, IssueSelector{Path: path, NativeID: nativeID})
	if err == nil && issue.Properties.ID == nativeID {
		return selected, issue, nil
	}
	if err != nil && !errors.Is(err, ErrNotFound) && !errors.Is(err, errNonIssueSelector) {
		return selected, issue, err
	}
	selectedErr := err
	// Shipping first probes the exact spelling and its configured-prefix form.
	// Querying the live catalog (rather than issues alone) excludes deleted and
	// otherwise unallocated native rows.
	var prefix string
	if err := tx.QueryRowContext(ctx, "SELECT value FROM config WHERE `key`='issue_prefix'").Scan(&prefix); err != nil {
		return "", IssueRecord{}, err
	}
	prefix = strings.TrimSuffix(prefix, "-")
	known := []string{prefix}
	var allowed string
	if err := tx.QueryRowContext(ctx, "SELECT value FROM config WHERE `key`='allowed_prefixes'").Scan(&allowed); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", IssueRecord{}, err
	}
	for _, value := range strings.Split(allowed, ",") {
		if value = strings.TrimSuffix(strings.TrimSpace(value), "-"); value != "" {
			known = append(known, value)
		}
	}
	search := nativeID
	if p := utils.ExtractIssuePrefixKnown(nativeID, known); p != "" {
		search = strings.TrimPrefix(nativeID, p+"-")
	}
	if !validPartialIssueHash(search) {
		return "", IssueRecord{}, ErrNotFound
	}
	type match struct{ path, id string }
	var normalizedExact []match
	seenPrefix := map[string]bool{}
	for _, candidatePrefix := range known {
		if candidatePrefix == "" || seenPrefix[candidatePrefix] {
			continue
		}
		seenPrefix[candidatePrefix] = true
		candidateID := candidatePrefix + "-" + search
		var candidatePath string
		err := tx.QueryRowContext(ctx, `SELECT path FROM graph_preview_catalog WHERE backing='issue' AND allocation_state='live' AND backing_key=?`, candidateID).Scan(&candidatePath)
		if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		if err != nil {
			return "", IssueRecord{}, err
		}
		normalizedExact = append(normalizedExact, match{candidatePath, candidateID})
	}
	if len(normalizedExact) > 0 {
		if len(normalizedExact) > 1 {
			return "", IssueRecord{}, fmt.Errorf("%w: Issue hash %q has multiple exact prefix matches; use a full ID", storage.ErrValidation, nativeID)
		}
		candidate := normalizedExact[0]
		if (selectedErr == nil && selected != candidate.path) || errors.Is(selectedErr, errNonIssueSelector) {
			return "", IssueRecord{}, fmt.Errorf("%w: bare Issue ID %q matches both a graph path and a different native Issue; use an explicit beads/PATH", storage.ErrValidation, nativeID)
		}
		resolved, err := s.requireIssueInTx(ctx, tx, candidate.path)
		return candidate.path, resolved, err
	}
	// SQL narrows the candidate set; Go applies the same leading-hash rule as
	// v1.3.1. The limit is a refusal bound, never a truncated answer.
	queryPart := search
	if i := strings.LastIndex(search, "-"); i >= 0 && i+1 < len(search) && validPartialIssueHash(search[i+1:]) {
		queryPart = search[i+1:]
	}
	rows, err := tx.QueryContext(ctx, `SELECT path,backing_key FROM graph_preview_catalog WHERE backing='issue' AND allocation_state='live' AND backing_key LIKE ? LIMIT 1001`, "%"+queryPart+"%")
	if err != nil {
		return "", IssueRecord{}, err
	}
	var exact, partial []match
	seen := 0
	for rows.Next() {
		seen++
		var candidate match
		if err := rows.Scan(&candidate.path, &candidate.id); err != nil {
			_ = rows.Close()
			return "", IssueRecord{}, err
		}
		if seen > 1000 {
			_ = rows.Close()
			return "", IssueRecord{}, fmt.Errorf("%w: too many Issue ID candidates for %q; use a longer ID", storage.ErrValidation, nativeID)
		}
		hash := candidate.id
		if p := utils.ExtractIssuePrefixKnown(candidate.id, known); p != "" {
			hash = strings.TrimPrefix(candidate.id, p+"-")
		}
		if candidate.id == nativeID || hash == search {
			exact = append(exact, candidate)
		} else if strings.HasPrefix(hash, search) {
			partial = append(partial, candidate)
		}
	}
	if err := errors.Join(rows.Err(), rows.Close()); err != nil {
		return "", IssueRecord{}, err
	}
	matches := partial
	if len(exact) > 0 {
		matches = exact
	}
	if len(matches) == 0 {
		return selected, issue, selectedErr
	}
	if selectedErr == nil && selected != matches[0].path {
		return "", IssueRecord{}, fmt.Errorf("%w: bare Issue ID %q matches both a graph path and a different native Issue; use an explicit beads/PATH", storage.ErrValidation, nativeID)
	}
	if errors.Is(selectedErr, errNonIssueSelector) {
		return "", IssueRecord{}, fmt.Errorf("%w: bare Issue ID %q matches a non-Issue path and a native Issue; use an explicit beads/PATH", storage.ErrValidation, nativeID)
	}
	if len(matches) > 1 {
		ids := make([]string, 0, len(matches))
		for _, candidate := range matches {
			ids = append(ids, candidate.id)
		}
		sort.Strings(ids)
		return "", IssueRecord{}, fmt.Errorf("%w: Issue ID %q matches %v; use more characters", storage.ErrValidation, nativeID, ids)
	}
	if err := validatePath(matches[0].path); err != nil {
		return "", IssueRecord{}, fmt.Errorf("%w: invalid native Issue allocation: %v", ErrInvalidStore, err)
	}
	issue, err = s.requireIssueInTx(ctx, tx, matches[0].path)
	return matches[0].path, issue, err
}

func validPartialIssueHash(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '.') {
			return false
		}
	}
	return true
}

func (s *Store) ShowIssueSelected(ctx context.Context, selector IssueSelector) (IssueRecord, error) {
	if err := validateIssueSelector(selector); err != nil {
		return IssueRecord{}, err
	}
	var result IssueRecord
	err := s.withTx(ctx, false, func(tx *sql.Tx) error {
		if err := checkBinding(ctx, tx, s.options); err != nil {
			return err
		}
		_, issue, err := s.resolveLiveIssueSelectorInTx(ctx, tx, selector)
		result = issue
		return err
	})
	if err != nil {
		if errors.Is(err, errNonIssueSelector) {
			return IssueRecord{}, ErrCapabilityUnavailable
		}
		return IssueRecord{}, err
	}
	return result, nil
}

// ResolveIssuePaths resolves count filters against the checked Issue catalog.
// Missing and non-Issue IDs retain their literal graph path so a count can
// return zero, as the ordinary command does. Ambiguous bare IDs fail.
func (s *Store) ResolveIssuePaths(ctx context.Context, selectors []IssueSelector) ([]string, error) {
	if len(selectors) == 0 {
		return []string{}, nil
	}
	for _, selector := range selectors {
		if err := validateIssueSelector(selector); err != nil {
			return nil, err
		}
	}
	paths := make([]string, len(selectors))
	err := s.withTx(ctx, false, func(tx *sql.Tx) error {
		if err := checkBinding(ctx, tx, s.options); err != nil {
			return err
		}
		for i, selector := range selectors {
			path, _, err := s.resolveLiveIssueSelectorInTx(ctx, tx, selector)
			switch {
			case errors.Is(err, ErrNotFound):
				paths[i] = selector.Path
			case errors.Is(err, errNonIssueSelector):
				paths[i] = selector.Path
			case err != nil:
				return err
			default:
				paths[i] = path
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return paths, nil
}
