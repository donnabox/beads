package graphstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/resourceget"
)

// Private, read-only and owned by exactly this checked transaction. No Store
// cache, writer flag or caller-authenticated Type record is introduced.
type issueDiscoveryReadKey struct{}
type issueDiscoveryRead struct {
	store   *Store
	tx      *sql.Tx
	budget  *intrinsicMetadataBudget
	capture checkedResourceCapture
}

func (s *Store) checkedIssueDiscoveryRead(ctx context.Context, tx *sql.Tx) (*issueDiscoveryRead, error) {
	frame, ok := ctx.Value(issueDiscoveryReadKey{}).(*issueDiscoveryRead)
	if !ok || frame == nil || frame.store != s || frame.tx != tx || frame.budget == nil || !actualCheckedOperationBinding(ctx, tx, s.options, frame.budget) {
		return nil, ErrInvalidStore
	}
	return frame, nil
}

// withCheckedIssueReadTx admits only the two selected14 native Issue readers.
// It reuses the complete same-transaction Type capture and hydration allowance;
// it does not open txOrdinary, grant writes, or convert Resource JSON to Issues.
func (s *Store) withCheckedIssueReadTx(ctx context.Context, fn func(context.Context, *sql.Tx) error) error {
	if s.LayoutVersion() != TypesFreshSchemaVersion {
		return ErrCapabilityUnavailable
	}
	err := s.withDefinitionReadTx(ctx, txDefinitionRead, func(tx *sql.Tx, budget *intrinsicMetadataBudget) error {
		if err := checkResourceQueryReadBytesInTx(ctx, tx); err != nil {
			return err
		}
		if err := chargeIssueDiscoveryPolicyInTx(ctx, tx, budget); err != nil {
			return err
		}
		capture, err := s.captureTypeRecordsInTx(ctx, tx, budget)
		if err != nil {
			return err
		}
		readctx := s.contextForResourceCapture(ctx, tx, capture, budget)
		readctx = context.WithValue(readctx, issueDiscoveryReadKey{}, &issueDiscoveryRead{store: s, tx: tx, budget: budget, capture: capture})
		return fn(readctx, tx)
	})
	if ctx != nil {
		err = errors.Join(err, ctx.Err())
	}
	return err
}

// This scalar gate precedes native IssueWithCounts/selected record hydration.
// The policy's SQL-backed names/categories and config fallback share the same
// operation allowance. This does not admit ListCustomStatuses' separate API or
// claim to bound the frontend's earlier YAML acquisition.
func chargeIssueDiscoveryPolicyInTx(ctx context.Context, tx *sql.Tx, budget *intrinsicMetadataBudget) error {
	var rowsUsed, bytesUsed uint64
	for _, query := range []string{
		"SELECT COUNT(*), CAST(LEAST(COALESCE(SUM(OCTET_LENGTH(name)+OCTET_LENGTH(category)),0),?) AS UNSIGNED) FROM custom_statuses",
		"SELECT COUNT(*), CAST(LEAST(COALESCE(SUM(OCTET_LENGTH(`key`)+OCTET_LENGTH(value)),0),?) AS UNSIGNED) FROM config WHERE `key`='status.custom'",
	} {
		if err := ctx.Err(); err != nil {
			return err
		}
		var rows, size uint64
		if err := tx.QueryRowContext(ctx, query, PreviewIssueListConfigByteLimit+1).Scan(&rows, &size); err != nil {
			return err
		}
		if rows > PreviewIssueListConfigRowLimit-rowsUsed || size > PreviewIssueListConfigByteLimit-bytesUsed {
			return ErrLimitExceeded
		}
		if err := budget.chargeBody(8 * int(size+256*rows)); err != nil {
			return err
		}
		rowsUsed += rows
		bytesUsed += size
	}
	return nil
}

const issueDiscoveryHydrationSizeQuery = `SELECT COUNT(*), CAST(LEAST(COALESCE(SUM(8192+12*(OCTET_LENGTH(v.durable_state)+OCTET_LENGTH(m.owned)+COALESCE(OCTET_LENGTH(m.revision_context),0))),0),?) AS UNSIGNED) FROM graph_preview_catalog c JOIN graph_preview_issue_versions m ON m.path=c.path AND m.version=c.revision JOIN issue_versions v ON v.issue_id=m.issue_id AND v.revision=m.issue_revision WHERE c.path=? AND c.backing='issue' AND c.allocation_state='live'`

// Invoked only for this discovery frame, before the unchanged native hydrator.
// Charge each actual read, including repeated selectors and parent readbacks;
// the aggregate native search debit is separate and is not reset per row.
func (s *Store) chargeIssueDiscoveryHydrationInTx(ctx context.Context, tx *sql.Tx, path string) error {
	if ctx.Value(issueDiscoveryReadKey{}) == nil {
		return nil
	}
	frame, err := s.checkedIssueDiscoveryRead(ctx, tx)
	if err != nil {
		return err
	}
	var count, size uint64
	if err = tx.QueryRowContext(ctx, issueDiscoveryHydrationSizeQuery, PreviewCurrentReadByteLimit+1, path).Scan(&count, &size); err != nil {
		return err
	}
	if count != 1 {
		return fmt.Errorf("%w: selected Issue retained mapping", ErrInvalidStore)
	}
	if size > PreviewCurrentReadByteLimit {
		return ErrLimitExceeded
	}
	return frame.budget.chargeBody(int(size))
}

// A selector/readback also proves its retained exact affiliation/captured Links,
// even when its native row is later excluded by a Count filter. Rows themselves
// still come from native SearchIssuesWithCountsInTx, not this document.
func (s *Store) verifyIssueDiscoveryRecordInTx(ctx context.Context, tx *sql.Tx, path string, record IssueRecord) error {
	if ctx.Value(issueDiscoveryReadKey{}) == nil {
		return nil
	}
	frame, err := s.checkedIssueDiscoveryRead(ctx, tx)
	if err != nil {
		return err
	}
	document, err := s.projectCheckedNativeInTx(ctx, tx, path, record.Revision, record, frame.capture.installed, frame.budget)
	if err != nil {
		return err
	}
	if document.ID() != record.ID || document.Revision() != record.Revision || document.Kind() != graph.KindBead {
		return ErrInvalidStore
	}
	return nil
}

func (s *Store) resolveIssueDiscoverySelectorInTx(ctx context.Context, tx *sql.Tx, selector IssueSelector) (string, IssueRecord, error) {
	if s.LayoutVersion() != TypesFreshSchemaVersion {
		return s.resolveLiveIssueSelectorInTx(ctx, tx, selector)
	}
	frame, err := s.checkedIssueDiscoveryRead(ctx, tx)
	if err != nil {
		return "", IssueRecord{}, err
	}
	if err = frame.budget.chargeBody(128 + 4*(len(selector.Path)+len(selector.NativeID))); err != nil {
		return "", IssueRecord{}, err
	}
	if selector.NativeID != "" {
		identity, lookupErr := s.resolveCommonResourceSelectorInTx(ctx, tx, "", selector.NativeID, frame.budget)
		if lookupErr == nil {
			if identity.State == "deleted" {
				return "", IssueRecord{}, ErrGone
			}
			// Exact installed definitions, Memory or Links do not fall through into
			// native abbreviation. Public-ID meaning/backing keys are unchanged.
			return s.resolveExactLiveIssueSelectorInTx(ctx, tx, IssueSelector{Path: identity.PrivatePath, Exact: true})
		}
		if !errors.Is(lookupErr, ErrNotFound) {
			return "", IssueRecord{}, lookupErr
		}
		if selector.Exact {
			return "", IssueRecord{}, ErrNotFound
		}
	}
	return s.resolveLiveIssueSelectorInTx(ctx, tx, selector)
}

const issueListCatalog14SizeQuery = `SELECT CAST(LEAST(COALESCE(SUM(4*(256+OCTET_LENGTH(path)+OCTET_LENGTH(resource_kind)+COALESCE(OCTET_LENGTH(type_url),0)+OCTET_LENGTH(revision)+OCTET_LENGTH(backing)+COALESCE(OCTET_LENGTH(backing_key),0)+COALESCE(OCTET_LENGTH(public_id),0))),0),?) AS UNSIGNED) FROM graph_preview_catalog WHERE allocation_state='live'`

func chargeIssueListCatalog14InTx(ctx context.Context, tx *sql.Tx, budget *intrinsicMetadataBudget) error {
	var size uint64
	if err := tx.QueryRowContext(ctx, issueListCatalog14SizeQuery, PreviewCurrentReadByteLimit+1).Scan(&size); err != nil {
		return err
	}
	if size > PreviewCurrentReadByteLimit {
		return ErrLimitExceeded
	}
	return budget.chargeBody(int(size))
}

// Pure exact correspondence; call only after the caller checks actual same-Tx
// read-frame authority. It performs no SQL and cannot promote arbitrary records
// into a checked definition, so it is safe while catalog rows are open.
func issueCatalogDefinitionMatches(definitions []resourceget.ResourceDocument, scope, path, kind, typ, revision, backing, publicID string) error {
	switch backing {
	case "intrinsic", "type-definition", "installed-type":
		if kind != "bead" {
			return ErrInvalidStore
		}
	case "intrinsic-link", "installed-type-link":
		if kind != "link" {
			return ErrInvalidStore
		}
	default:
		return ErrInvalidStore
	}
	if validateResourcePath(path) != nil || validatePublicID(publicID) != nil {
		return ErrInvalidStore
	}
	uri, err := graph.CanonicalPublicResourceURL(scope, publicID)
	if err != nil {
		return ErrInvalidStore
	}
	matches := 0
	for _, document := range definitions {
		if document.ID() != uri {
			continue
		}
		matches++
		if string(document.Kind()) != kind || document.Type().URI != typ || document.Revision() != revision {
			return ErrInvalidStore
		}
	}
	if matches != 1 {
		return fmt.Errorf("%w: definition allocation differs from checked stored record", ErrInvalidStore)
	}
	return nil
}
