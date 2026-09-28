package graphstore

// These read-only helpers preserve complete Memory owned-state validation from
// the qualified mixed implementation. C0 installs no Link or Issue writer.
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
	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/types"
)

func informationalProperties(properties map[string]any) ([]byte, error) {
	if properties == nil {
		properties = map[string]any{}
	}
	for key, value := range properties {
		note, ok := value.(string)
		if key != "note" || !ok || !utf8.ValidString(note) {
			return nil, fmt.Errorf("%w: informational Link properties admit only optional UTF-8 string note", storage.ErrValidation)
		}
	}
	return canonicalJSON(properties)
}

func (s *Store) currentInformationalLinkInTx(ctx context.Context, tx *sql.Tx, path string) (LinkRecord, error) {
	var kind, typ, revision, state, backing string
	var key sql.NullString
	if err := tx.QueryRowContext(ctx, `SELECT resource_kind,type_url,revision,allocation_state,backing,backing_key FROM graph_preview_catalog WHERE path=?`, path).Scan(&kind, &typ, &revision, &state, &backing, &key); err != nil {
		return LinkRecord{}, err
	}
	if kind != "link" || typ != RelatedTypeURL(s.options.Binding.ScopeURL) || !authorityID.MatchString(revision) || state != "live" || backing != "informational" || key.Valid {
		return LinkRecord{}, fmt.Errorf("%w: invalid informational Link allocation", ErrInvalidStore)
	}
	var source, target string
	var properties, attribution []byte
	if err := tx.QueryRowContext(ctx, `SELECT source_path,target_path,properties,attribution FROM graph_preview_links WHERE path=?`, path).Scan(&source, &target, &properties, &attribution); err != nil {
		return LinkRecord{}, fmt.Errorf("%w: missing informational backing: %v", ErrInvalidStore, err)
	}
	for _, endpoint := range []string{source, target} {
		var kind, typ, state, backing string
		if err := tx.QueryRowContext(ctx, `SELECT resource_kind,type_url,allocation_state,backing FROM graph_preview_catalog WHERE path=?`, endpoint).Scan(&kind, &typ, &state, &backing); err != nil {
			return LinkRecord{}, fmt.Errorf("%w: missing endpoint: %v", ErrInvalidStore, err)
		}
		if validatePath(endpoint) != nil || kind != "bead" || state != "live" || !((backing == "generic" && typ == MemoryTypeURL(s.options.Binding.ScopeURL)) || (backing == "issue" && typ == IssueTypeURL(s.options.Binding.ScopeURL))) {
			return LinkRecord{}, fmt.Errorf("%w: invalid informational endpoint", ErrInvalidStore)
		}
	}
	r := LinkRecord{ID: graph.CanonicalURL(s.options.Binding.ScopeURL, path), Type: typ, Revision: revision, Version: revision, Source: graph.CanonicalURL(s.options.Binding.ScopeURL, source), Target: graph.CanonicalURL(s.options.Binding.ScopeURL, target)}
	if json.Unmarshal(properties, &r.Properties) != nil || json.Unmarshal(attribution, &r.Attribution) != nil {
		return LinkRecord{}, fmt.Errorf("%w: malformed informational payload", ErrInvalidStore)
	}
	p, err := informationalProperties(r.Properties)
	if err != nil || !bytes.Equal(p, properties) {
		return LinkRecord{}, fmt.Errorf("%w: invalid informational properties", ErrInvalidStore)
	}
	a, err := canonicalJSON(r.Attribution)
	if err != nil || !bytes.Equal(a, attribution) {
		return LinkRecord{}, fmt.Errorf("%w: invalid informational attribution", ErrInvalidStore)
	}
	at, err := time.Parse(time.RFC3339Nano, r.Attribution.RecordedAt)
	if err != nil || at.UTC().Format(time.RFC3339Nano) != r.Attribution.RecordedAt || !utf8.ValidString(r.Attribution.Actor) || (r.Attribution.Actor == "" && r.Attribution.Status != "unknown") || (r.Attribution.Actor != "" && r.Attribution.Status != "claimed") {
		return LinkRecord{}, fmt.Errorf("%w: invalid informational attribution", ErrInvalidStore)
	}
	return r, nil
}

func (s *Store) memoryOwnedLinksInTx(ctx context.Context, tx *sql.Tx, sourcePath string) ([]json.RawMessage, error) {
	rows, err := tx.QueryContext(ctx, `SELECT l.path FROM graph_preview_links l WHERE l.source_path=?`, sourcePath)
	if err != nil {
		return nil, err
	}
	paths := []string{}
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			_ = rows.Close()
			return nil, err
		}
		paths = append(paths, path)
	}
	if err := errors.Join(rows.Err(), rows.Close()); err != nil {
		return nil, err
	}
	if len(paths) > PreviewOwnedLinkLimit {
		return nil, fmt.Errorf("%w: owned Link limit exceeded", ErrInvalidStore)
	}
	sort.Slice(paths, func(i, j int) bool { return graph.CompareCodeUnits(paths[i], paths[j]) < 0 })
	owned := make([]json.RawMessage, 0, len(paths))
	for _, path := range paths {
		link, err := s.showLinkInTx(ctx, tx, path)
		if err != nil {
			return nil, err
		}
		raw, err := canonicalJSON(link)
		if err != nil {
			return nil, err
		}
		owned = append(owned, raw)
	}
	return owned, nil
}

func (s *Store) currentLinkInTx(ctx context.Context, tx *sql.Tx, path string) (LinkRecord, error) {
	var selectedBacking string
	if err := tx.QueryRowContext(ctx, `SELECT backing FROM graph_preview_catalog WHERE path=?`, path).Scan(&selectedBacking); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return LinkRecord{}, ErrNotFound
		}
		return LinkRecord{}, err
	}
	var selectedState string
	if err := tx.QueryRowContext(ctx, `SELECT allocation_state FROM graph_preview_catalog WHERE path=?`, path).Scan(&selectedState); err != nil {
		return LinkRecord{}, err
	}
	if selectedState == "deleted" {
		return LinkRecord{}, s.deletedLinkErrorInTx(ctx, tx, path)
	}
	if selectedBacking == "informational" {
		return s.currentInformationalLinkInTx(ctx, tx, path)
	}

	var kind, typ, revision, state, backing, key string
	err := tx.QueryRowContext(ctx, `SELECT resource_kind,type_url,revision,allocation_state,backing,backing_key FROM graph_preview_catalog WHERE path=?`, path).Scan(&kind, &typ, &revision, &state, &backing, &key)
	if errors.Is(err, sql.ErrNoRows) {
		return LinkRecord{}, ErrNotFound
	}
	if err != nil {
		return LinkRecord{}, err
	}
	if kind != "link" || typ != DependencyTypeURL(s.options.Binding.ScopeURL) || !authorityID.MatchString(revision) || state != "live" || backing != "dependency" || key == "" {
		return LinkRecord{}, fmt.Errorf("%w: invalid Dependency allocation", ErrInvalidStore)
	}
	var sourcePath, targetPath, depType, metadata, thread, actor string
	var at time.Time
	err = tx.QueryRowContext(ctx, `SELECT src.path,dst.path,d.type,d.metadata,d.thread_id,d.created_by,d.created_at
 FROM dependencies d JOIN graph_preview_catalog src ON src.backing='issue' AND src.backing_key=d.issue_id AND src.allocation_state='live' AND src.resource_kind='bead' AND src.type_url=?
 JOIN graph_preview_catalog dst ON dst.backing='issue' AND dst.backing_key=d.depends_on_issue_id AND dst.allocation_state='live' AND dst.resource_kind='bead' AND dst.type_url=?
 WHERE d.id=? AND d.depends_on_wisp_id IS NULL AND d.depends_on_external IS NULL`, IssueTypeURL(s.options.Binding.ScopeURL), IssueTypeURL(s.options.Binding.ScopeURL), key).Scan(&sourcePath, &targetPath, &depType, &metadata, &thread, &actor, &at)
	if err != nil {
		return LinkRecord{}, fmt.Errorf("%w: Dependency backing: %v", ErrInvalidStore, err)
	}
	if depType != string(types.DepBlocks) || metadata != "{}" || thread != "" {
		return LinkRecord{}, fmt.Errorf("%w: Dependency exceeds preview", ErrInvalidStore)
	}
	status := "unknown"
	if actor != "" {
		status = "claimed"
	}
	return LinkRecord{ID: graph.CanonicalURL(s.options.Binding.ScopeURL, path), Type: typ, Revision: revision, Version: revision,
		Source: graph.CanonicalURL(s.options.Binding.ScopeURL, sourcePath), Target: graph.CanonicalURL(s.options.Binding.ScopeURL, targetPath), Properties: map[string]any{},
		Attribution: Attribution{Actor: actor, Status: status, RecordedAt: at.UTC().Format(time.RFC3339Nano)}}, nil
}

func (s *Store) showLinkInTx(ctx context.Context, tx *sql.Tx, path string) (LinkRecord, error) {
	link, err := s.currentLinkInTx(ctx, tx, path)
	if err != nil {
		return LinkRecord{}, err
	}
	var retained []byte
	var actor string
	if err := tx.QueryRowContext(ctx, `SELECT snapshot,actor FROM graph_preview_versions WHERE path=? AND version=?`, path, link.Version).Scan(&retained, &actor); err != nil {
		return LinkRecord{}, fmt.Errorf("%w: missing Link retained state: %v", ErrInvalidStore, err)
	}
	canonical, err := canonicalJSON(link)
	if err != nil {
		return LinkRecord{}, err
	}
	if !bytes.Equal(canonical, retained) || actor != link.Attribution.Actor {
		return LinkRecord{}, fmt.Errorf("%w: current and retained Link state differ", ErrInvalidStore)
	}
	return link, nil
}

func (s *Store) deletedLinkErrorInTx(ctx context.Context, tx *sql.Tx, path string) error {
	var kind, typ, revision, backing string
	var key sql.NullString
	if err := tx.QueryRowContext(ctx, `SELECT resource_kind,type_url,revision,backing,backing_key FROM graph_preview_catalog WHERE path=?`, path).Scan(&kind, &typ, &revision, &backing, &key); err != nil {
		return err
	}
	if !s.validDeletedLinkAllocation(kind, typ, backing, key) || !authorityID.MatchString(revision) {
		return fmt.Errorf("%w: invalid deleted allocation", ErrInvalidStore)
	}
	var snapshot []byte
	var actor string
	if err := tx.QueryRowContext(ctx, `SELECT snapshot,actor FROM graph_preview_versions WHERE path=? AND version=?`, path, revision).Scan(&snapshot, &actor); err != nil {
		return fmt.Errorf("%w: missing deletion state: %v", ErrInvalidStore, err)
	}
	var tombstone LinkTombstone
	if err := json.Unmarshal(snapshot, &tombstone); err != nil {
		return fmt.Errorf("%w: malformed deletion state", ErrInvalidStore)
	}
	canonical, err := canonicalJSON(tombstone)
	at, timeErr := time.Parse(time.RFC3339Nano, tombstone.Attribution.RecordedAt)
	if err != nil || !bytes.Equal(snapshot, canonical) || tombstone.ID != graph.CanonicalURL(s.options.Binding.ScopeURL, path) || tombstone.Type != typ || tombstone.State != "deleted" || tombstone.Revision != revision || tombstone.Version != revision || !authorityID.MatchString(tombstone.PreviousVersion) || tombstone.PreviousVersion == revision || actor != tombstone.Attribution.Actor || !utf8.ValidString(actor) || timeErr != nil || at.UTC().Format(time.RFC3339Nano) != tombstone.Attribution.RecordedAt || (actor == "" && tombstone.Attribution.Status != "unknown") || (actor != "" && tombstone.Attribution.Status != "claimed") {
		return fmt.Errorf("%w: inconsistent deletion state", ErrInvalidStore)
	}
	var previous []byte
	var previousActor string
	if err := tx.QueryRowContext(ctx, `SELECT snapshot,actor FROM graph_preview_versions WHERE path=? AND version=?`, path, tombstone.PreviousVersion).Scan(&previous, &previousActor); err != nil {
		return fmt.Errorf("%w: missing previous Link state: %v", ErrInvalidStore, err)
	}
	var link LinkRecord
	if json.Unmarshal(previous, &link) != nil || link.ID != tombstone.ID || link.Type != typ || link.Revision != tombstone.PreviousVersion || link.Version != tombstone.PreviousVersion {
		return fmt.Errorf("%w: invalid previous Link state", ErrInvalidStore)
	}
	if err := s.validateVersionLink(link, previous, previousActor); err != nil {
		return err
	}
	var count int
	if err := tx.QueryRowContext(ctx, `SELECT (SELECT COUNT(*) FROM graph_preview_links WHERE path=?) + (SELECT COUNT(*) FROM graph_preview_payloads WHERE path=?)`, path, path).Scan(&count); err != nil {
		return err
	}
	if count != 0 {
		return fmt.Errorf("%w: deleted Link has current payload", ErrInvalidStore)
	}
	return ErrGone
}

func (s *Store) validateRetainedInformational(link LinkRecord, snapshot []byte, actor string) error {
	canonical, err := canonicalJSON(link)
	if err != nil || !bytes.Equal(canonical, snapshot) || link.Properties == nil {
		return fmt.Errorf("%w: noncanonical retained Link", ErrInvalidStore)
	}
	for _, endpoint := range []string{link.Source, link.Target} {
		path := strings.TrimPrefix(endpoint, s.options.Binding.ScopeURL)
		if validatePath(path) != nil || graph.CanonicalURL(s.options.Binding.ScopeURL, path) != endpoint {
			return fmt.Errorf("%w: invalid retained Link endpoint", ErrInvalidStore)
		}
	}
	if _, err := informationalProperties(link.Properties); err != nil {
		return fmt.Errorf("%w: invalid retained Link properties", ErrInvalidStore)
	}
	at, err := time.Parse(time.RFC3339Nano, link.Attribution.RecordedAt)
	if err != nil || at.UTC().Format(time.RFC3339Nano) != link.Attribution.RecordedAt || actor != link.Attribution.Actor || !utf8.ValidString(actor) || (actor == "" && link.Attribution.Status != "unknown") || (actor != "" && link.Attribution.Status != "claimed") {
		return fmt.Errorf("%w: invalid retained Link attribution", ErrInvalidStore)
	}
	return nil
}

func (s *Store) validDeletedLinkAllocation(kind, typ, backing string, key sql.NullString) bool {
	return kind == "link" && !key.Valid &&
		((backing == "informational" && typ == RelatedTypeURL(s.ScopeURL())) ||
			(backing == "dependency" && typ == DependencyTypeURL(s.ScopeURL())))
}

func (s *Store) ScopeURL() string { return s.options.Binding.ScopeURL }

func sameVersionJSON(raw []byte, value any) bool {
	canonical, err := canonicalJSON(value)
	return err == nil && bytes.Equal(raw, canonical)
}

func validVersionAttribution(value Attribution, actor string) bool {
	at, err := time.Parse(time.RFC3339Nano, value.RecordedAt)
	return err == nil && at.UTC().Format(time.RFC3339Nano) == value.RecordedAt && utf8.ValidString(actor) &&
		value.Actor == actor && ((actor == "" && value.Status == "unknown") || (actor != "" && value.Status == "claimed"))
}

func (s *Store) validateVersionLink(link LinkRecord, raw []byte, actor string) error {
	path := strings.TrimPrefix(link.ID, s.ScopeURL())
	if validateLinkPath(path) != nil || s.ScopeURL()+path != link.ID || !authorityID.MatchString(link.Revision) ||
		link.Version != link.Revision || !sameVersionJSON(raw, link) || !validVersionAttribution(link.Attribution, actor) {
		return fmt.Errorf("%w: invalid retained Link", ErrInvalidStore)
	}
	if link.Type == RelatedTypeURL(s.ScopeURL()) {
		return s.validateRetainedInformational(link, raw, actor)
	}
	if link.Type != DependencyTypeURL(s.ScopeURL()) || link.Properties == nil || len(link.Properties) != 0 {
		return fmt.Errorf("%w: invalid retained Dependency", ErrInvalidStore)
	}
	for _, endpoint := range []string{link.Source, link.Target} {
		path := strings.TrimPrefix(endpoint, s.ScopeURL())
		if validatePath(path) != nil || s.ScopeURL()+path != endpoint {
			return fmt.Errorf("%w: invalid retained endpoint", ErrInvalidStore)
		}
	}
	return nil
}

func validateLinkPath(path string) error {
	if err := graph.ValidateLinkPath(path); err != nil {
		return err
	}
	if len(path) > 1024 {
		return fmt.Errorf("%w: preview Link path exceeds 1024 bytes", storage.ErrValidation)
	}
	return nil
}
