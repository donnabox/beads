//go:build cgo

package embeddeddolt

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/steveyegge/beads/internal/storage"
	storeops "github.com/steveyegge/beads/internal/storage/issueops"
)

// Private complete-format adapters retain upstream behavior for explicit
// disposable Survival fixtures. No normal open or schema probe selects them.
func (s *EmbeddedDoltStore) survivalCurrentEpoch(ctx context.Context, storeID string) (int, error) {
	var epoch int
	if err := s.withConn(ctx, false, func(tx *sql.Tx) error {
		var err error
		epoch, err = storeops.SurvivalCurrentEpochInTx(ctx, tx, storeID)
		return err
	}); err != nil {
		return 0, err
	}
	return epoch, nil
}

func (s *EmbeddedDoltStore) survivalBumpEpoch(ctx context.Context, storeID, reason string) (int, error) {
	var epoch int
	if err := s.runIssueOperationTxWithMessage(ctx, func(tx *sql.Tx) (storeops.ChangedTables, string, error) {
		var err error
		epoch, err = storeops.SurvivalBumpEpochInTx(ctx, tx, storeID, reason)
		if err != nil {
			return nil, "", err
		}
		return storeops.ChangedTables{"store_epoch": true},
			fmt.Sprintf("bd: bump epoch for %s (%s)", storeID, reason), nil
	}); err != nil {
		return failedBump(storeID, epoch, err)
	}
	return epoch, nil
}

func (s *EmbeddedDoltStore) survivalBumpEpochCarrying(ctx context.Context, storeID, reason string) (int, error) {
	var epoch int
	if err := s.runIssueOperationTxWithMessage(ctx, func(tx *sql.Tx) (storeops.ChangedTables, string, error) {
		var err error
		epoch, err = storeops.SurvivalBumpEpochCarryingInTx(ctx, tx, storeID, reason)
		if err != nil {
			return nil, "", err
		}
		return storeops.ChangedTables{"store_epoch": true, "epoch_minted_addresses": true},
			fmt.Sprintf("bd: carry addresses to epoch %d for %s (%s)", epoch, storeID, reason), nil
	}); err != nil {
		return failedBump(storeID, epoch, err)
	}
	return epoch, nil
}

func (s *EmbeddedDoltStore) survivalMintUnderEpoch(ctx context.Context, storeID, id string) (string, error) {
	var address string
	if err := s.runIssueOperationTxWithMessage(ctx, func(tx *sql.Tx) (storeops.ChangedTables, string, error) {
		var err error
		address, err = storeops.SurvivalMintUnderEpochInTx(ctx, tx, storeID, id)
		if err != nil {
			return nil, "", err
		}
		return storeops.SurvivalEpochMintDirtyTables(),
			fmt.Sprintf("bd: mint %s under epoch for %s", id, storeID), nil
	}); err != nil {
		if errors.Is(err, storage.ErrCommitIndeterminate) {
			return address, err
		}
		return "", err
	}
	return address, nil
}

func (s *EmbeddedDoltStore) survivalStillServes(ctx context.Context, storeID, address string) (bool, error) {
	var serves bool
	if err := s.withConn(ctx, false, func(tx *sql.Tx) error {
		var err error
		serves, err = storeops.SurvivalStillServesInTx(ctx, tx, storeID, address)
		return err
	}); err != nil {
		return false, err
	}
	return serves, nil
}

func (s *EmbeddedDoltStore) survivalResolve(ctx context.Context, storeID, address string) (storeops.EpochResolveResult, error) {
	var result storeops.EpochResolveResult
	if err := s.withConn(ctx, false, func(tx *sql.Tx) error {
		var err error
		result, err = storeops.SurvivalResolveEpochInTx(ctx, tx, storeID, address)
		return err
	}); err != nil {
		return storeops.EpochResolveResult{}, err
	}
	return result, nil
}

func (s *EmbeddedDoltStore) survivalCurrentAddressFor(ctx context.Context, storeID, oldAddress string) (string, error) {
	var address string
	if err := s.withConn(ctx, false, func(tx *sql.Tx) error {
		var err error
		address, err = storeops.SurvivalCurrentAddressForInTx(ctx, tx, storeID, oldAddress)
		return err
	}); err != nil {
		return "", err
	}
	return address, nil
}

func (s *EmbeddedDoltStore) survivalLoseVersion(ctx context.Context, storeID, address string) error {
	return s.runIssueOperationTxWithMessage(ctx, func(tx *sql.Tx) (storeops.ChangedTables, string, error) {
		if err := storeops.SurvivalLoseVersionInTx(ctx, tx, storeID, address); err != nil {
			return nil, "", err
		}
		return storeops.ChangedTables{"epoch_minted_addresses": true},
			fmt.Sprintf("bd: lose version %s for %s", address, storeID), nil
	})
}
