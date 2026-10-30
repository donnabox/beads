//go:build cgo

package embeddeddolt

import (
	"context"
	storeops "github.com/steveyegge/beads/internal/storage/issueops"
	"github.com/steveyegge/beads/internal/storage/versioncontrolops"
	"github.com/steveyegge/beads/internal/testutil/epochfixture"
)

// SurvivalTestAdapter exists only in the test binary.
type SurvivalTestAdapter struct{ store *EmbeddedDoltStore }

func NewSurvivalTestAdapter(store *EmbeddedDoltStore) SurvivalTestAdapter {
	return SurvivalTestAdapter{store}
}
func PrepareSurvivalTestStore(ctx context.Context, store *EmbeddedDoltStore) error {
	return store.withPinnedDBConn(ctx, func(conn versioncontrolops.DBConn) error { return epochfixture.Prepare(ctx, conn) })
}
func (s SurvivalTestAdapter) CurrentEpoch(ctx context.Context, storeID string) (int, error) {
	return s.store.survivalCurrentEpoch(ctx, storeID)
}
func (s SurvivalTestAdapter) BumpEpoch(ctx context.Context, storeID, reason string) (int, error) {
	return s.store.survivalBumpEpoch(ctx, storeID, reason)
}
func (s SurvivalTestAdapter) BumpEpochCarrying(ctx context.Context, storeID, reason string) (int, error) {
	return s.store.survivalBumpEpochCarrying(ctx, storeID, reason)
}
func (s SurvivalTestAdapter) MintUnderEpoch(ctx context.Context, storeID, id string) (string, error) {
	return s.store.survivalMintUnderEpoch(ctx, storeID, id)
}
func (s SurvivalTestAdapter) StillServes(ctx context.Context, storeID, address string) (bool, error) {
	return s.store.survivalStillServes(ctx, storeID, address)
}
func (s SurvivalTestAdapter) Resolve(ctx context.Context, storeID, address string) (storeops.EpochResolveResult, error) {
	return s.store.survivalResolve(ctx, storeID, address)
}
func (s SurvivalTestAdapter) CurrentAddressFor(ctx context.Context, storeID, oldAddress string) (string, error) {
	return s.store.survivalCurrentAddressFor(ctx, storeID, oldAddress)
}
func (s SurvivalTestAdapter) LoseVersion(ctx context.Context, storeID, address string) error {
	return s.store.survivalLoseVersion(ctx, storeID, address)
}
