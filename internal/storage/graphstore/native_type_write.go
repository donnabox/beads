package graphstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	graph "github.com/steveyegge/beads/graphops"
	publicops "github.com/steveyegge/beads/issueops"
	"github.com/steveyegge/beads/resourceget"
)

// selectedStoredTypeEditionInTx selects the actual stored definition edition
// at inception/reclassification. Callers are native writers whose same-Tx
// admission has already checked the installed record/compiler inventory. This
// helper cannot substitute a descriptor fingerprint for a Resource revision.
func (s *Store) selectedStoredTypeEditionInTx(ctx context.Context, tx *sql.Tx, uri string, budget *intrinsicMetadataBudget) (resourceget.TypeReference, error) {
	if s.LayoutVersion() != TypesFreshSchemaVersion {
		return resourceget.TypeReference{}, ErrCapabilityUnavailable
	}
	if err := budget.chargeBody(8192 + 4*len(uri)); err != nil {
		return resourceget.TypeReference{}, err
	}
	publicID, err := graph.ScopeLocalPublicResourceID(s.ScopeURL(), uri)
	if err != nil {
		return resourceget.TypeReference{}, err
	}
	var count int
	if err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM graph_preview_catalog WHERE public_id=? AND OCTET_LENGTH(public_id)=?`, publicID, len(publicID)).Scan(&count); err != nil {
		return resourceget.TypeReference{}, err
	}
	if count != 1 {
		return resourceget.TypeReference{}, fmt.Errorf("%w: installed Type identity is absent or ambiguous", ErrInvalidStore)
	}
	var revision string
	var size uint64
	if err = tx.QueryRowContext(ctx, `SELECT c.revision,OCTET_LENGTH(p.properties) FROM graph_preview_catalog c JOIN graph_preview_payloads p ON p.path=c.path WHERE c.public_id=? AND OCTET_LENGTH(c.public_id)=? AND c.backing='installed-type' AND c.resource_kind='bead' AND c.allocation_state='live'`, publicID, len(publicID)).Scan(&revision, &size); err != nil {
		return resourceget.TypeReference{}, fmt.Errorf("%w: installed Type definition allocation: %w", ErrInvalidStore, err)
	}
	if size == 0 || size > PreviewCurrentReadByteLimit || validateVersionToken(revision) != nil {
		return resourceget.TypeReference{}, ErrInvalidStore
	}
	if err = budget.chargeBody(4 * int(size)); err != nil {
		return resourceget.TypeReference{}, err
	}
	// This is a redundant identity guard on the actual definition body. Full
	// source/schema/contract correspondence remains the operation's admission.
	var body []byte
	if err = tx.QueryRowContext(ctx, `SELECT p.properties FROM graph_preview_payloads p JOIN graph_preview_catalog c ON c.path=p.path WHERE c.public_id=? AND OCTET_LENGTH(c.public_id)=? AND c.revision=? AND OCTET_LENGTH(p.properties)=?`, publicID, len(publicID), revision, size).Scan(&body); err != nil {
		return resourceget.TypeReference{}, err
	}
	var record struct {
		ID       string `json:"id"`
		Revision string `json:"revision"`
	}
	if json.Unmarshal(body, &record) != nil || record.ID != uri || record.Revision != revision {
		return resourceget.TypeReference{}, fmt.Errorf("%w: installed Type definition identity differs", ErrInvalidStore)
	}
	return resourceget.TypeReference{URI: uri, Revision: revision}, nil
}

// retainNativeResourceVersionInTx preserves the predecessor's exact Type pin
// on unrelated updates and owned-Link successors. A changed nominal Type is a
// new selection from the SAME transaction's checked installed definition. No
// first-use or retained read fills missing old evidence from current state.
func (s *Store) retainNativeResourceVersionInTx(ctx context.Context, tx *sql.Tx, path, version string, snapshot []byte, actor string, budget *intrinsicMetadataBudget) error {
	if s.LayoutVersion() != TypesFreshSchemaVersion {
		return insertPreviewVersionInTx(ctx, tx, path, version, snapshot, actor)
	}
	if err := budget.chargeBody(8*len(snapshot) + 8192); err != nil {
		return err
	}
	var header struct {
		Type  string `json:"type"`
		State string `json:"state"`
	}
	if json.Unmarshal(snapshot, &header) != nil {
		return fmt.Errorf("%w: malformed native version", ErrInvalidStore)
	}
	// Terminal markers are deliberately not Resource records. Their existing
	// predecessor/version checks remain authoritative, and they acquire no Type.
	if header.State == "deleted" {
		return insertPreviewVersionInTx(ctx, tx, path, version, snapshot, actor)
	}
	if graph.ValidateTypeURL(header.Type) != nil {
		return fmt.Errorf("%w: native version lacks nominal Type", ErrInvalidStore)
	}
	var previous []byte
	var size uint64
	err := tx.QueryRowContext(ctx, `SELECT OCTET_LENGTH(snapshot)+OCTET_LENGTH(actor) FROM graph_preview_versions WHERE path=? ORDER BY ordinal DESC LIMIT 1`, path).Scan(&size)
	var pin resourceget.TypeReference
	if err == nil {
		if size > PreviewCurrentReadByteLimit {
			return ErrLimitExceeded
		}
		if err = budget.chargeBody(4 * int(size)); err != nil {
			return err
		}
		var previousVersion, previousActor string
		var previousOrdinal int64
		if err = tx.QueryRowContext(ctx, `SELECT snapshot,version,actor,ordinal FROM graph_preview_versions WHERE path=? AND OCTET_LENGTH(snapshot)+OCTET_LENGTH(actor)=? ORDER BY ordinal DESC LIMIT 1`, path, size).Scan(&previous, &previousVersion, &previousActor, &previousOrdinal); err != nil {
			return err
		}
		prior, err := s.nativePredecessorTypeInTx(ctx, tx, path, header.Type, previousVersion, previousActor, previousOrdinal, previous, budget)
		if err != nil {
			return err
		}
		if prior.URI == header.Type {
			pin = prior
		}
	} else if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if pin.URI == "" {
		pin, err = s.selectedStoredTypeEditionInTx(ctx, tx, header.Type, budget)
		if err != nil {
			return err
		}
	}
	encoded, err := encodeNativeSnapshotEdition(snapshot, pin)
	if err != nil {
		return err
	}
	return insertPreviewVersionInTx(ctx, tx, path, version, encoded, actor)
}

// issueTypeContextForWriteInTx binds one exact native issue_versions body, not
// issues.current_revision (a batch can retain intermediate native revisions).
func (s *Store) issueTypeContextForWriteInTx(ctx context.Context, tx *sql.Tx, path, issueID string, ordinal int64, value issueRevisionContext, budget *intrinsicMetadataBudget) ([]byte, error) {
	if s.LayoutVersion() != TypesFreshSchemaVersion {
		return encodeIssueRevisionContext(value)
	}
	if err := budget.chargeBody(8192); err != nil {
		return nil, err
	}
	var size uint64
	if err := tx.QueryRowContext(ctx, `SELECT OCTET_LENGTH(durable_state) FROM issue_versions WHERE issue_id=? AND revision=?`, issueID, ordinal).Scan(&size); err != nil {
		return nil, err
	}
	if size == 0 || size > PreviewCurrentReadByteLimit {
		return nil, ErrLimitExceeded
	}
	if err := budget.chargeBody(4 * int(size)); err != nil {
		return nil, err
	}
	var raw []byte
	if err := tx.QueryRowContext(ctx, `SELECT durable_state FROM issue_versions WHERE issue_id=? AND revision=? AND OCTET_LENGTH(durable_state)=?`, issueID, ordinal, size).Scan(&raw); err != nil {
		return nil, err
	}
	var issue publicops.Issue
	if json.Unmarshal(raw, &issue) != nil || issue.ID != issueID || !sameVersionJSON(raw, &issue) {
		return nil, fmt.Errorf("%w: native Issue version identity", ErrInvalidStore)
	}
	uri, err := NativeIssueTypeURL(s.ScopeURL(), string(issue.IssueType))
	if err != nil {
		return nil, err
	}
	var pin resourceget.TypeReference
	var previousSize sql.NullInt64
	err = tx.QueryRowContext(ctx, `SELECT OCTET_LENGTH(revision_context) FROM graph_preview_issue_versions WHERE path=? ORDER BY history_ordinal DESC LIMIT 1`, path).Scan(&previousSize)
	if err == nil {
		if !previousSize.Valid || previousSize.Int64 <= 0 || previousSize.Int64 > PreviewCurrentReadByteLimit {
			return nil, fmt.Errorf("%w: native Issue predecessor Type evidence", ErrInvalidStore)
		}
		if err = budget.chargeBody(4 * int(previousSize.Int64)); err != nil {
			return nil, err
		}
		var previous []byte
		if err = tx.QueryRowContext(ctx, `SELECT revision_context FROM graph_preview_issue_versions WHERE path=? AND OCTET_LENGTH(revision_context)=? ORDER BY history_ordinal DESC LIMIT 1`, path, previousSize.Int64).Scan(&previous); err != nil {
			return nil, err
		}
		_, prior, err := decodeNativeIssueTypeEdition(previous)
		if err != nil {
			return nil, err
		}
		if prior.URI == uri {
			pin = prior
		}
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if pin.URI == "" {
		pin, err = s.selectedStoredTypeEditionInTx(ctx, tx, uri, budget)
		if err != nil {
			return nil, err
		}
	}
	return encodeNativeIssueTypeEdition(value, pin)
}

// installNativeDefinitionInTx is part of the existing concrete native writer's
// first-use transaction, not general Type authoring. Descriptor, ordinary Type
// Bead and real Refines capture either commit together or all roll back.
func (s *Store) installNativeDefinitionInTx(ctx context.Context, tx *sql.Tx, name string, descriptor graph.TypeDescriptor) error {
	state, err := s.nativeOperationInContext(ctx, tx)
	if err != nil {
		return err
	}
	if state.capture.plan == nil {
		return ErrInvalidStore
	}
	source, err := nativeInstalledTypeSource(s.ScopeURL(), name, descriptor)
	if err != nil {
		return err
	}
	refs := map[string]resourceget.TypeReference{}
	for _, fact := range state.capture.installed.facts {
		if !fact.Current || fact.Reference != fact.Definition {
			return ErrInvalidStore
		}
		refs[fact.Reference.URI] = fact.Reference
	}
	candidates, err := prepareInstalledTypeSources(ctx, tx, s.options, state.capture.plan, state.budget, []installedTypeSource{source}, refs)
	if err != nil {
		return err
	}
	if err = installTypeCandidatesInTx(ctx, tx, state.capture.plan, state.budget, candidates); err != nil {
		return err
	}
	// Qualification is from the actual stored records after all inserts, not the
	// prospective compiler. Shared allowance applies to this redundant census.
	installed, err := loadInstalledTypesInTx(ctx, tx, s.options, state.capture.plan, state.budget)
	if err != nil {
		return err
	}
	state.capture.installed = installed
	return nil
}
