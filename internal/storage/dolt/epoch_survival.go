package dolt

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	storeops "github.com/steveyegge/beads/internal/storage/issueops"
)

// Private complete-format adapters retain upstream behavior for explicit
// disposable Survival fixtures. No normal open or schema probe selects them.
func (s *DoltStore) survivalCurrentEpoch(ctx context.Context, storeID string) (int, error) {
	var epoch int
	if err := s.withReadTx(ctx, func(tx *sql.Tx) error {
		var err error
		epoch, err = storeops.SurvivalCurrentEpochInTx(ctx, tx, storeID)
		return err
	}); err != nil {
		return 0, err
	}
	return epoch, nil
}

func (s *DoltStore) survivalBumpEpoch(ctx context.Context, storeID, reason string) (int, error) {
	var epoch int
	if err := s.withRetryTx(ctx, func(tx *sql.Tx) error {
		var err error
		epoch, err = storeops.SurvivalBumpEpochInTx(ctx, tx, storeID, reason)
		return err
	}); err != nil {
		return failedBump(storeID, epoch, err)
	}
	if err := s.doltAddAndCommitPostTx(ctx, []string{"store_epoch"},
		fmt.Sprintf("bd: bump epoch for %s (%s)", storeID, reason)); err != nil {
		return epoch, unpublishedBump(storeID, epoch, err)
	}
	return epoch, nil
}

func (s *DoltStore) survivalBumpEpochCarrying(ctx context.Context, storeID, reason string) (int, error) {
	var epoch int
	if err := s.withRetryTx(ctx, func(tx *sql.Tx) error {
		var err error
		epoch, err = storeops.SurvivalBumpEpochCarryingInTx(ctx, tx, storeID, reason)
		return err
	}); err != nil {
		return failedBump(storeID, epoch, err)
	}
	if err := s.doltAddAndCommitPostTx(ctx, []string{"store_epoch", "epoch_minted_addresses"},
		fmt.Sprintf("bd: carry addresses to epoch %d for %s (%s)", epoch, storeID, reason)); err != nil {
		return epoch, unpublishedBump(storeID, epoch, err)
	}
	return epoch, nil
}

func (s *DoltStore) survivalMintUnderEpoch(ctx context.Context, storeID, id string) (string, error) {
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
		if errors.Is(err, ErrCommitIndeterminate) {
			return address, err
		}
		return "", err
	}
	return address, nil
}

func (s *DoltStore) survivalStillServes(ctx context.Context, storeID, address string) (bool, error) {
	var serves bool
	if err := s.withReadTx(ctx, func(tx *sql.Tx) error {
		var err error
		serves, err = storeops.SurvivalStillServesInTx(ctx, tx, storeID, address)
		return err
	}); err != nil {
		return false, err
	}
	return serves, nil
}

func (s *DoltStore) survivalResolve(ctx context.Context, storeID, address string) (storeops.EpochResolveResult, error) {
	var result storeops.EpochResolveResult
	if err := s.withReadTx(ctx, func(tx *sql.Tx) error {
		var err error
		result, err = storeops.SurvivalResolveEpochInTx(ctx, tx, storeID, address)
		return err
	}); err != nil {
		return storeops.EpochResolveResult{}, err
	}
	return result, nil
}

func (s *DoltStore) survivalCurrentAddressFor(ctx context.Context, storeID, oldAddress string) (string, error) {
	var address string
	if err := s.withReadTx(ctx, func(tx *sql.Tx) error {
		var err error
		address, err = storeops.SurvivalCurrentAddressForInTx(ctx, tx, storeID, oldAddress)
		return err
	}); err != nil {
		return "", err
	}
	return address, nil
}

func (s *DoltStore) survivalLoseVersion(ctx context.Context, storeID, address string) error {
	if err := s.withRetryTx(ctx, func(tx *sql.Tx) error {
		return storeops.SurvivalLoseVersionInTx(ctx, tx, storeID, address)
	}); err != nil {
		return err
	}
	if err := s.doltAddAndCommitPostTx(ctx, []string{"epoch_minted_addresses"},
		fmt.Sprintf("bd: lose version %s for %s", address, storeID)); err != nil {
		return fmt.Errorf("loss of %s for %s recorded but its Dolt commit failed: %w: %w", address, storeID, err, ErrCommitIndeterminate)
	}
	return nil
}
