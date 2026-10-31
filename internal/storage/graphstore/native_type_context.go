package graphstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

// retainResourceVersionInTx is the native-domain version seam. Fixed intrinsic
// and installed definition writers retain their complete exact-pin bodies via
// the original insert helper; ordinary11 remains byte-for-byte unchanged.
func (s *Store) retainResourceVersionInTx(ctx context.Context, tx *sql.Tx, path, version string, snapshot []byte, actor string) error {
	if s.LayoutVersion() != TypesFreshSchemaVersion {
		return insertPreviewVersionInTx(ctx, tx, path, version, snapshot, actor)
	}
	state, err := s.nativeOperationInContext(ctx, tx)
	if err != nil {
		return err
	}
	return s.retainNativeResourceVersionInTx(ctx, tx, path, version, snapshot, actor, state.budget)
}

func (s *Store) nativeDomainSnapshot(raw []byte) ([]byte, error) {
	if s.LayoutVersion() != TypesFreshSchemaVersion {
		return raw, nil
	}
	var header struct {
		State string `json:"state"`
	}
	if json.Unmarshal(raw, &header) != nil {
		return nil, fmt.Errorf("%w: native retained JSON", ErrInvalidStore)
	}
	if header.State == "deleted" {
		return raw, nil
	}
	domain, _, err := decodeNativeSnapshotEdition(raw)
	return domain, err
}

func (s *Store) readVersionDomainBytes(ctx context.Context, tx *sql.Tx, path, version, head string) ([]byte, string, error) {
	raw, actor, err := readVersionBytes(ctx, tx, path, version, head)
	if err != nil {
		return nil, "", err
	}
	domain, err := s.nativeDomainSnapshot(raw)
	if err != nil {
		return nil, "", err
	}
	return domain, actor, nil
}

func (s *Store) applyNativeIssueRevisionContext(attribution Attribution, raw []byte) (Attribution, error) {
	if s.LayoutVersion() != TypesFreshSchemaVersion {
		return applyIssueRevisionContext(attribution, raw)
	}
	value, _, err := decodeNativeIssueTypeEdition(raw)
	if err != nil {
		return Attribution{}, err
	}
	attribution.Inception, attribution.HistoricalUpdate = value.Inception, value.HistoricalUpdate
	return attribution, nil
}

func (s *Store) currentNativeIssueRevisionContextInTx(ctx context.Context, tx *sql.Tx, path string) (issueRevisionContext, error) {
	if s.LayoutVersion() != TypesFreshSchemaVersion {
		return currentIssueRevisionContextInTx(ctx, tx, path)
	}
	state, err := s.nativeOperationInContext(ctx, tx)
	if err != nil {
		return issueRevisionContext{}, err
	}
	var size sql.NullInt64
	if err = tx.QueryRowContext(ctx, `SELECT OCTET_LENGTH(m.revision_context) FROM graph_preview_catalog c JOIN graph_preview_issue_versions m ON m.path=c.path AND m.version=c.revision WHERE c.path=?`, path).Scan(&size); err != nil {
		return issueRevisionContext{}, err
	}
	if !size.Valid || size.Int64 <= 0 || size.Int64 > PreviewCurrentReadByteLimit {
		return issueRevisionContext{}, fmt.Errorf("%w: native Issue recorded context", ErrInvalidStore)
	}
	if err = state.budget.chargeBody(4*int(size.Int64) + 4096); err != nil {
		return issueRevisionContext{}, err
	}
	var raw []byte
	if err = tx.QueryRowContext(ctx, `SELECT m.revision_context FROM graph_preview_catalog c JOIN graph_preview_issue_versions m ON m.path=c.path AND m.version=c.revision WHERE c.path=? AND OCTET_LENGTH(m.revision_context)=?`, path, size.Int64).Scan(&raw); err != nil {
		return issueRevisionContext{}, err
	}
	value, _, err := decodeNativeIssueTypeEdition(raw)
	return value, err
}

// captureAppliedNativeResultInTx must run inside the admitted callback before
// commit. A later commit/close failure still returns the caller's zero result.
// No renderer needs a postcommit latest query to reconstruct the applied pin.
func (s *Store) captureAppliedNativeResultInTx(ctx context.Context, tx *sql.Tx, path string, result any) error {
	if s.LayoutVersion() != TypesFreshSchemaVersion {
		return nil
	}
	state, err := s.nativeOperationInContext(ctx, tx)
	if err != nil {
		return err
	}
	var revision string
	var value any
	switch record := result.(type) {
	case *Record:
		revision = record.Revision
		value = *record
	case *IssueRecord:
		revision = record.Revision
		value = *record
	case *LinkRecord:
		revision = record.Revision
		value = *record
	default:
		return ErrInvalidStore
	}
	document, err := s.projectCheckedNativeInTx(ctx, tx, path, revision, value, state.capture.installed, state.budget)
	if err != nil {
		return err
	}
	switch record := result.(type) {
	case *Record:
		record.resourceDocument = &document
	case *IssueRecord:
		record.resourceDocument = &document
	case *LinkRecord:
		record.resourceDocument = &document
	}
	return nil
}

func (s *Store) captureNativeReadbackResultInTx(ctx context.Context, tx *sql.Tx, path string, result any) error {
	if s.LayoutVersion() != TypesFreshSchemaVersion || ctx.Value(nativeOperationKey{}) == nil {
		return nil
	}
	return s.captureAppliedNativeResultInTx(ctx, tx, path, result)
}

func (s *Store) insertInitialNativeIssueMappingInTx(ctx context.Context, tx *sql.Tx, path, revision, issueID string, ordinal int64) error {
	state, err := s.nativeOperationInContext(ctx, tx)
	if err != nil {
		return err
	}
	encoded, err := s.issueTypeContextForWriteInTx(ctx, tx, path, issueID, ordinal, issueRevisionContext{Inception: "create"}, state.budget)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO graph_preview_issue_versions (path,version,issue_id,issue_revision,owned,revision_context,history_ordinal) VALUES (?,?,?,?,'[]',?,?)`, path, revision, issueID, ordinal, encoded, ordinal)
	return err
}
