//go:build cgo

package embeddeddolt_test

import (
	"context"

	"testing"

	"github.com/steveyegge/beads/backend/conformance"
	"github.com/steveyegge/beads/internal/storage/embeddeddolt"
	storeops "github.com/steveyegge/beads/internal/storage/issueops"
)

// The production/default store retains the released epoch contract.
func TestLegacyEpochContract(t *testing.T) {
	skipUnlessEmbeddedDolt(t)
	te := newTestEnv(t, "epch")
	ctx := t.Context()
	fixture := conformance.LegacyEpochFixture{
		IssuePrefix:       "epch",
		CurrentEpoch:      legacyEpochEmbeddedCurrentEpoch(te.store),
		BumpEpoch:         legacyEpochEmbeddedBumpEpoch(te.store),
		MintUnderEpoch:    legacyEpochEmbeddedMintUnderEpoch(te.store),
		StillServes:       legacyEpochEmbeddedStillServes(te.store),
		Resolve:           legacyEpochEmbeddedResolve(te.store),
		CurrentAddressFor: legacyEpochEmbeddedCurrentAddressFor(te.store),
	}

	t.Run("AnEpochBumpIsTriggeredOnlyByRestoreReinitOrSchemeChange", func(t *testing.T) {
		conformance.RunLegacyAnEpochBumpIsTriggeredOnlyByRestoreReinitOrSchemeChange(t, ctx, fixture)
	})
	t.Run("EpochBumpVoidsOnlyAddressesOfVersionsNoLongerServed", func(t *testing.T) {
		conformance.RunLegacyEpochBumpVoidsOnlyAddressesOfVersionsNoLongerServed(t, ctx, fixture)
	})
}

func legacyEpochEmbeddedCurrentEpoch(store *embeddeddolt.EmbeddedDoltStore) func(ctx context.Context, storeID string) (int, error) {
	return func(ctx context.Context, storeID string) (int, error) {
		return store.CurrentEpoch(ctx, storeID)
	}
}

func legacyEpochEmbeddedBumpEpoch(store *embeddeddolt.EmbeddedDoltStore) func(ctx context.Context, storeID string, trigger conformance.EpochBumpTrigger) (int, error) {
	return func(ctx context.Context, storeID string, trigger conformance.EpochBumpTrigger) (int, error) {
		return store.BumpEpoch(ctx, storeID, trigger.String())
	}
}

func legacyEpochEmbeddedMintUnderEpoch(store *embeddeddolt.EmbeddedDoltStore) func(ctx context.Context, storeID, id string) (conformance.Address, error) {
	return func(ctx context.Context, storeID, id string) (conformance.Address, error) {
		address, err := store.MintUnderEpoch(ctx, storeID, id)
		if err != nil {
			return "", err
		}
		return conformance.Address(address), nil
	}
}

func legacyEpochEmbeddedStillServes(store *embeddeddolt.EmbeddedDoltStore) func(ctx context.Context, storeID string, address conformance.Address) (bool, error) {
	return func(ctx context.Context, storeID string, address conformance.Address) (bool, error) {
		return store.StillServes(ctx, storeID, string(address))
	}
}

func legacyEpochEmbeddedResolve(store *embeddeddolt.EmbeddedDoltStore) func(ctx context.Context, storeID string, address conformance.Address) (conformance.RetentionAnswer, error) {
	return func(ctx context.Context, storeID string, address conformance.Address) (conformance.RetentionAnswer, error) {
		result, err := store.Resolve(ctx, storeID, string(address))
		if err != nil {
			return conformance.RetentionAnswer{}, err
		}
		return legacyEpochResolveResultToConformance(result), nil
	}
}

func legacyEpochEmbeddedCurrentAddressFor(store *embeddeddolt.EmbeddedDoltStore) func(ctx context.Context, storeID string, oldAddress conformance.Address) (conformance.Address, error) {
	return func(ctx context.Context, storeID string, oldAddress conformance.Address) (conformance.Address, error) {
		address, err := store.CurrentAddressFor(ctx, storeID, string(oldAddress))
		if err != nil {
			return "", err
		}
		return conformance.Address(address), nil
	}
}

// legacyEpochRestrictionToConformance translates this leg's local
// storeops.EpochRestriction into backend/conformance's own Restriction
// vocabulary — the contract-test file is the translation boundary, per
// epoch_cas.go's package doc.
func legacyEpochRestrictionToConformance(r storeops.EpochRestriction) conformance.Restriction {
	switch r {
	case storeops.EpochRestrictionLive:
		return conformance.RestrictionLive
	case storeops.EpochRestrictionGoneReorganization:
		return conformance.RestrictionGoneReorganization
	default:
		return conformance.RestrictionUnknown
	}
}

func legacyEpochResolveResultToConformance(r storeops.EpochResolveResult) conformance.RetentionAnswer {
	return conformance.RetentionAnswer{
		Restriction:    legacyEpochRestrictionToConformance(r.Restriction),
		ProducingStore: r.ProducingStore,
		Epoch:          r.Epoch,
	}
}
