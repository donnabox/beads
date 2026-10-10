package graphstore

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/legacyimport"
	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/storage/depid"
	"github.com/steveyegge/beads/internal/storage/domain"
	"github.com/steveyegge/beads/internal/storage/issueops"
	"github.com/steveyegge/beads/internal/types"
)

// LegacyImportResult makes each old identity's new graph address explicit.
type LegacyImportResult struct {
	Issues        int               `json:"issues"`
	Memories      int               `json:"memories"`
	Dependencies  int               `json:"dependencies"`
	Comments      int               `json:"comments"`
	IDs           map[string]string `json:"ids"`
	MemoryIDs     map[string]string `json:"memory_ids"`
	DependencyIDs map[string]string `json:"dependency_ids"`
	DryRun        bool              `json:"dry_run"`
	History       string            `json:"history"`
}

// ImportLegacy installs an ordinary export into an empty graph. Every native
// row, canonical allocation and initial retained snapshot shares one checked
// transaction. It never adopts existing rows or replays a failed transaction.
// Dry-run executes the same engine validation in a rollback-only transaction.
func (s *Store) ImportLegacy(ctx context.Context, input legacyimport.Batch, actor string, dryRun bool) (LegacyImportResult, error) {
	batch, err := prepareLegacyImport(input, actor)
	if err != nil {
		return LegacyImportResult{}, err
	}
	result := LegacyImportResult{Issues: len(batch.Issues), Memories: len(batch.Memories),
		IDs: map[string]string{}, MemoryIDs: map[string]string{}, DependencyIDs: map[string]string{},
		DryRun: dryRun, History: "current data only; new initial graph versions; source history is not present in legacy export"}
	err = s.withTx(ctx, !dryRun, func(tx *sql.Tx) error {
		if err := checkBinding(ctx, tx, s.options); err != nil {
			return err
		}
		if err := requireEmptyLegacyTarget(ctx, tx); err != nil {
			return err
		}
		if len(batch.Issues)+len(batch.Memories) == 0 {
			return nil
		}
		if err := s.touchCoordination(ctx, tx); err != nil {
			return err
		}
		infra := issueops.ResolveInfraTypesInTx(ctx, tx)
		for _, issue := range batch.Issues {
			if infra[string(issue.IssueType.Normalize())] {
				return fmt.Errorf("%w: infrastructure Issue %s is unsupported", storage.ErrValidation, issue.ID)
			}
		}
		unscope := issueops.ScopeVersionedHistoryTransaction(tx, true)
		defer unscope()
		if len(batch.Issues) > 0 {
			_, err := issueops.CreateIssuesInTxWithResult(ctx, tx, batch.Issues, actor, storage.BatchCreateOptions{CreateOnly: true, SkipPrefixValidation: true})
			if err != nil {
				if errors.Is(err, domain.ErrDependencyCycle) {
					return fmt.Errorf("%w: legacy dependencies contain a cycle", storage.ErrValidation)
				}
				return err
			}
			if err := s.afterStage("import-native"); err != nil {
				return err
			}
		}
		for _, issue := range batch.Issues {
			path := "beads/" + issue.ID
			if _, err := tx.ExecContext(ctx, `INSERT INTO graph_preview_catalog (path,resource_kind,type_url,revision,allocation_state,backing,backing_key) VALUES (?,'bead',?,?,'live','issue',?)`, path, IssueTypeURL(s.options.Binding.ScopeURL), strings.Repeat("0", 32), issue.ID); err != nil {
				return err
			}
			result.IDs[issue.ID] = graph.CanonicalURL(s.options.Binding.ScopeURL, path)
			result.Comments += len(issue.Comments)
		}
		if err := s.afterStage("import-issues"); err != nil {
			return err
		}
		for _, issue := range batch.Issues {
			for _, dep := range issue.Dependencies {
				key := depid.New(issue.ID, dep.DependsOnID)
				path := "links/" + key
				revision, err := freshToken()
				if err != nil {
					return err
				}
				if _, err := tx.ExecContext(ctx, `INSERT INTO graph_preview_catalog (path,resource_kind,type_url,revision,allocation_state,backing,backing_key) VALUES (?,'link',?,?,'live','dependency',?)`, path, DependencyTypeURL(s.options.Binding.ScopeURL), revision, key); err != nil {
					return err
				}
				link, err := s.currentLinkInTx(ctx, tx, path)
				if err != nil {
					return err
				}
				raw, err := canonicalJSON(link)
				if err != nil {
					return err
				}
				if err := insertPreviewVersionInTx(ctx, tx, path, revision, raw, link.Attribution.Actor); err != nil {
					return err
				}
				result.DependencyIDs[issue.ID+" -> "+dep.DependsOnID] = link.ID
				result.Dependencies++
			}
		}
		if err := s.afterStage("import-links"); err != nil {
			return err
		}
		for _, issue := range batch.Issues {
			path := "beads/" + issue.ID
			if err := s.recordIssueMappingInTx(ctx, tx, path, issue.ID); err != nil {
				return err
			}
			// Detect column coercion instead of silently changing source data.
			stored, err := issueops.HydrateIssueOperationResult(ctx, tx, issue.ID, true)
			if err != nil {
				return err
			}
			if err := verifyLegacyIssue(issue, stored); err != nil {
				return err
			}
			if _, err := s.showIssueInTx(ctx, tx, path); err != nil {
				return err
			}
		}
		for _, memory := range batch.Memories {
			path := "beads/" + memory.Key
			revision, err := freshToken()
			if err != nil {
				return err
			}
			r := Record{ID: graph.CanonicalURL(s.options.Binding.ScopeURL, path), Type: MemoryTypeURL(s.options.Binding.ScopeURL), Revision: revision, Version: revision,
				Properties: Properties{Title: memory.Key, Body: memory.Value}, Metadata: json.RawMessage(`{}`), Owned: []json.RawMessage{},
				Attribution: Attribution{Actor: actor, Status: "claimed", RecordedAt: time.Now().UTC().Format(time.RFC3339Nano)}}
			properties, err := canonicalJSON(r.Properties)
			if err != nil {
				return err
			}
			snapshot, err := canonicalJSON(r)
			if err != nil {
				return err
			}
			if err := s.createInTx(ctx, tx, CreateRequest{Path: path, Actor: actor}, r, properties, snapshot); err != nil {
				return err
			}
			if _, err := s.showMemoryInTx(ctx, tx, path); err != nil {
				return err
			}
			result.MemoryIDs[memory.Key] = r.ID
		}
		if err := s.afterStage("import-retained"); err != nil {
			return err
		}
		return checkCurrentReadBytes(ctx, tx)
	})
	if err != nil {
		return LegacyImportResult{}, err
	}
	return result, nil
}

func requireEmptyLegacyTarget(ctx context.Context, tx *sql.Tx) error {
	// Include reserved/deleted allocations and unmapped native rows: deleting
	// everything or bypassing graph writers never makes a workspace fresh again.
	for _, table := range []string{"graph_preview_catalog", "issues", "wisps", "dependencies", "comments", "issue_versions", "graph_preview_versions", "graph_preview_issue_versions"} {
		var count int
		//nolint:gosec // G201: table is from the fixed internal list above.
		if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table).Scan(&count); err != nil {
			return err
		}
		if count != 0 {
			return fmt.Errorf("%w: legacy import requires an empty fresh graph workspace; %s is occupied (repeat imports do not update or duplicate data)", storage.ErrValidation, table)
		}
	}
	return nil
}

func prepareLegacyImport(input legacyimport.Batch, actor string) (legacyimport.Batch, error) {
	invalid := func(format string, args ...any) (legacyimport.Batch, error) {
		return legacyimport.Batch{}, fmt.Errorf("%w: "+format, append([]any{storage.ErrValidation}, args...)...)
	}
	if actor == "" || !utf8.ValidString(actor) {
		return invalid("legacy import requires a UTF-8 actor")
	}
	// Isolate mutable native-import normalization from the caller's values.
	raw, err := json.Marshal(input)
	if err != nil {
		return invalid("legacy input: %v", err)
	}
	if len(raw) > legacyimport.MaxBytes {
		return invalid("legacy import exceeds %d bytes", legacyimport.MaxBytes)
	}
	var batch legacyimport.Batch
	if err := json.Unmarshal(raw, &batch); err != nil {
		return invalid("legacy input: %v", err)
	}
	ids := map[string]bool{}
	commentIDs := map[string]bool{}
	resources := len(batch.Issues) + len(batch.Memories)
	for _, issue := range batch.Issues {
		if issue == nil || issue.ID == "" {
			return invalid("each exported Issue requires an ID")
		}
		if err := validatePath("beads/" + issue.ID); err != nil {
			return invalid("Issue ID %q: %v", issue.ID, err)
		}
		if ids[issue.ID] {
			return invalid("duplicate Issue ID %q", issue.ID)
		}
		ids[issue.ID] = true
		if issue.Ephemeral || issue.NoHistory || issue.StorageClass.Normalize() == types.StorageClassEphemeral || issue.StorageClass.Normalize() == types.StorageClassUnversioned || issue.LeaseExpiresAt != nil || issue.HeartbeatAt != nil || issue.LeaseGrantedNode != "" {
			return invalid("Issue %s has unsupported ephemeral, no-history or live lease data", issue.ID)
		}
		if issue.Status == "tombstone" {
			return invalid("tombstone %s is unsupported", issue.ID)
		}
		issue.SetDefaults()
		now := time.Now().UTC().Truncate(time.Second)
		if issue.CreatedAt.IsZero() {
			issue.CreatedAt = now
		}
		if issue.UpdatedAt.IsZero() {
			issue.UpdatedAt = now
		}
		metadata, err := commonMetadata(issue.Metadata)
		if err != nil {
			return invalid("Issue %s metadata: %v", issue.ID, err)
		}
		issue.Metadata = metadata
		if err := validateIssueCreateFields(issue); err != nil {
			return invalid("Issue %s: %v", issue.ID, err)
		}
		if err := issueops.PrepareIssueForInsert(issue, nil, nil); err != nil {
			return invalid("Issue %s: %v", issue.ID, err)
		}
		pairs := map[string]bool{}
		for _, dep := range issue.Dependencies {
			if dep == nil {
				return invalid("Issue %s has a null dependency", issue.ID)
			}
			if dep.IssueID == "" {
				dep.IssueID = issue.ID
			}
			if types.ExtractPrefix(dep.IssueID) != types.ExtractPrefix(dep.DependsOnID) || issueops.IsExternalDepTarget(dep.IssueID, dep.DependsOnID) {
				return invalid("dependency %s -> %s uses unsupported external or cross-prefix routing", dep.IssueID, dep.DependsOnID)
			}
			if dep.Metadata != "" {
				canonical, err := commonMetadata(json.RawMessage(dep.Metadata))
				if err != nil || string(canonical) != "{}" {
					return invalid("dependency metadata for %s must be an empty object", issue.ID)
				}
				dep.Metadata = "{}"
			}
			if dep.IssueID != issue.ID || dep.Type != types.DepBlocks || dep.DependsOnID == issue.ID || dep.ThreadID != "" {
				return invalid("Issue %s supports only local blocks dependencies without edge metadata or threads", issue.ID)
			}
			if pairs[dep.DependsOnID] {
				return invalid("duplicate dependency %s -> %s", issue.ID, dep.DependsOnID)
			}
			pairs[dep.DependsOnID] = true
			resources++
		}
		if len(issue.Dependencies) > PreviewOwnedLinkLimit {
			return invalid("Issue %s exceeds owned Link limit", issue.ID)
		}
		for _, comment := range issue.Comments {
			if comment == nil {
				return invalid("Issue %s has a null comment", issue.ID)
			}
			if comment.IssueID != "" && comment.IssueID != issue.ID {
				return invalid("comment issue_id differs from %s", issue.ID)
			}
			comment.IssueID = issue.ID
			if comment.ID != "" && commentIDs[comment.ID] {
				return invalid("duplicate comment ID %q", comment.ID)
			}
			if comment.ID != "" {
				commentIDs[comment.ID] = true
			}
		}
	}
	for _, issue := range batch.Issues {
		for _, dep := range issue.Dependencies {
			if !ids[dep.DependsOnID] {
				return invalid("dependency %s -> %s does not name an Issue in this import", issue.ID, dep.DependsOnID)
			}
		}
	}
	for _, memory := range batch.Memories {
		if memory.Key == "" || !utf8.ValidString(memory.Key) || !utf8.ValidString(memory.Value) {
			return invalid("memory requires a nonempty UTF-8 key and UTF-8 value")
		}
		if err := validatePath("beads/" + memory.Key); err != nil {
			return invalid("memory key %q: %v", memory.Key, err)
		}
		if ids[memory.Key] {
			return invalid("memory key %q duplicates another Bead identity", memory.Key)
		}
		ids[memory.Key] = true
	}
	if resources > PreviewSnapshotLimit {
		return invalid("legacy import exceeds preview limit of %d Beads and Links", PreviewSnapshotLimit)
	}
	return batch, nil
}

func verifyLegacyIssue(want, got *types.Issue) error {
	// Native import intentionally regenerates dependency surrogate IDs and
	// derived readiness. Comments remain outside the retained Issue payload.
	w, g := *want, *got
	var err error
	w.Metadata, err = commonMetadata(w.Metadata)
	if err != nil {
		return err
	}
	g.Metadata, err = commonMetadata(g.Metadata)
	if err != nil {
		return err
	}
	w.Labels = append([]string(nil), w.Labels...)
	g.Labels = append([]string(nil), g.Labels...)
	sort.Strings(w.Labels)
	sort.Strings(g.Labels)
	w.Dependencies, g.Dependencies = nil, nil
	w.Comments, g.Comments = nil, nil
	w.IsBlocked, g.IsBlocked = false, false
	w.ContentHash, g.ContentHash = "", ""
	w.RowVersion, g.RowVersion = 0, 0
	wraw, err := canonicalJSON(w)
	if err != nil {
		return err
	}
	graw, err := canonicalJSON(g)
	if err != nil {
		return err
	}
	if !bytes.Equal(wraw, graw) {
		var wf, gf map[string]json.RawMessage
		if err := json.Unmarshal(wraw, &wf); err != nil {
			return err
		}
		if err := json.Unmarshal(graw, &gf); err != nil {
			return err
		}
		for key, value := range wf {
			if !bytes.Equal(value, gf[key]) {
				return fmt.Errorf("%w: Issue %s field %s cannot be represented exactly in graph storage", storage.ErrValidation, w.ID, key)
			}
		}
		return fmt.Errorf("%w: Issue %s acquired unsupported fields in graph storage", storage.ErrValidation, w.ID)
	}
	// Retain the native comment feed exactly, including imported IDs. Detect
	// duplicate-content collapse and timestamp coercion before committing.
	if len(want.Comments) != len(got.Comments) {
		return fmt.Errorf("%w: comments for %s cannot be preserved exactly", storage.ErrValidation, w.ID)
	}
	for _, comment := range want.Comments {
		found := false
		for _, stored := range got.Comments {
			if comment.ID == stored.ID && comment.IssueID == stored.IssueID && comment.Author == stored.Author && comment.Text == stored.Text && (comment.CreatedAt.IsZero() || comment.CreatedAt.Equal(stored.CreatedAt)) {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("%w: comment for %s cannot be preserved exactly", storage.ErrValidation, w.ID)
		}
	}
	// Verify edge provenance separately; the graph Link reader validates its
	// shape and the native retained snapshot validates the complete edge set.
	if len(want.Dependencies) != len(got.Dependencies) {
		return fmt.Errorf("%w: dependencies for %s cannot be preserved exactly", storage.ErrValidation, w.ID)
	}
	for _, dep := range want.Dependencies {
		found := false
		for _, stored := range got.Dependencies {
			if dep.IssueID == stored.IssueID && dep.DependsOnID == stored.DependsOnID && dep.Type == stored.Type && (dep.CreatedBy == "" || dep.CreatedBy == stored.CreatedBy) && (dep.CreatedAt.IsZero() || dep.CreatedAt.Equal(stored.CreatedAt)) {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("%w: dependency for %s cannot be preserved exactly", storage.ErrValidation, w.ID)
		}
	}
	return nil
}
