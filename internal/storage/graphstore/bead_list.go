package graphstore

import (
	"context"
	"database/sql"
	"fmt"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/internal/workapi"
)

// BeadListRequest selects one page of the current Bead inventory: the Memories
// and Issues the ordinary Issue list would show, across every installed Bead
// Type.
type BeadListRequest struct {
	// TypeURL keeps only Beads of this installed Bead Type; empty keeps every Type.
	TypeURL string
	// All also lists the Issues the ordinary list hides by default.
	All bool
	// Limit keeps the first Limit Beads; zero keeps every Bead.
	Limit int
	// MaxRows refuses a result of more than MaxRows Beads; zero disables the check.
	MaxRows int
}

// BeadListPage contains complete canonical Memory and Issue records. HasMore
// reports an intentional page limit, not a snapshot cursor.
type BeadListPage struct {
	Items   []any `json:"items"`
	HasMore bool  `json:"hasMore"`
}

// ListBeads reads the current inventory and, when hiding applies, the Issue
// list configuration in one read transaction, then hides, pages and caps the
// Beads with pageBeads. It retains no transaction between requests.
func (s *Store) ListBeads(ctx context.Context, request BeadListRequest) (BeadListPage, error) {
	var page BeadListPage
	err := s.withTx(ctx, false, func(tx *sql.Tx) error {
		snapshot, err := s.currentSnapshotInTx(ctx, tx)
		if err != nil {
			return err
		}
		if err := checkBeadListType(snapshot.Types, request.TypeURL); err != nil {
			return err
		}
		var cfg workapi.ListConfig
		if !request.All {
			if cfg, err = issueListConfigInTx(ctx, tx); err != nil {
				return err
			}
		}
		page, err = pageBeads(snapshot.Records, cfg, request)
		return err
	})
	if err != nil {
		return BeadListPage{}, err
	}
	return page, nil
}

func checkBeadListType(installed []graph.TypeDescriptor, typeURL string) error {
	if typeURL == "" {
		return nil
	}
	for _, descriptor := range installed {
		if descriptor.ID() == typeURL && descriptor.Describes() == graph.KindBead {
			return nil
		}
	}
	return fmt.Errorf("%w: --bead-type must name an installed Bead Type", ErrCapabilityUnavailable)
}

// pageBeads is the whole listing policy over already validated records, so it
// needs no database: it keeps the Beads of the requested Type, hides what the
// ordinary list hides unless request.All is set, applies the row cap and trims
// to the limit.
func pageBeads(records []any, cfg workapi.ListConfig, request BeadListRequest) (BeadListPage, error) {
	page := BeadListPage{Items: []any{}}
	for _, value := range records {
		switch record := value.(type) {
		case Record:
			if request.TypeURL == "" || record.Type == request.TypeURL {
				page.Items = append(page.Items, record)
			}
		case IssueRecord:
			if record.Properties == nil {
				return BeadListPage{}, fmt.Errorf("%w: Issue record without properties", ErrInvalidStore)
			}
			if request.TypeURL != "" && record.Type != request.TypeURL {
				continue
			}
			if request.All || !issueHiddenByDefault(record.Properties.Status, record.Properties.Pinned, cfg) {
				page.Items = append(page.Items, record)
			}
		case LinkRecord:
			// A Link is a Resource, but never a Bead.
		default:
			return BeadListPage{}, fmt.Errorf("%w: unsupported current Resource projection", ErrInvalidStore)
		}
	}
	if request.MaxRows > 0 && len(page.Items) > request.MaxRows {
		return BeadListPage{}, fmt.Errorf("%w: Bead list exceeds configured maximum of %d rows", ErrLimitExceeded, request.MaxRows)
	}
	if request.Limit > 0 && len(page.Items) > request.Limit {
		page.Items = page.Items[:request.Limit]
		page.HasMore = true
	}
	return page, nil
}

// issueHiddenByDefault is the one definition of what the Bead list leaves out
// unless asked: what the ordinary Issue list excludes by default, which is a
// closed or pinned status, a set pinned flag, or a custom status whose category
// is done or frozen. workapi.BuildListFilter owns that rule for the Issue
// query; a test pins this copy to it, so an upstream change to the default
// exclusion turns that test red instead of silently diverging.
func issueHiddenByDefault(status types.Status, pinned bool, cfg workapi.ListConfig) bool {
	if pinned || status == types.StatusClosed || status == types.StatusPinned {
		return true
	}
	for _, custom := range cfg.CustomStatuses {
		if types.Status(custom.Name) == status && (custom.Category == types.CategoryDone || custom.Category == types.CategoryFrozen) {
			return true
		}
	}
	return false
}
