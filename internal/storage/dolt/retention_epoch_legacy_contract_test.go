package dolt

import (
	"context"
	"testing"

	"github.com/steveyegge/beads/backend/conformance"
	storeops "github.com/steveyegge/beads/internal/storage/issueops"
)

// The production/default store retains the released epoch contract.
func TestLegacyEpochContract(t *testing.T) {
	fixture, ctx, cleanup := legacyNewEpochDoltFixture(t, "epch")
	defer cleanup()

	t.Run("AnEpochBumpIsTriggeredOnlyByRestoreReinitOrSchemeChange", func(t *testing.T) {
		conformance.RunLegacyAnEpochBumpIsTriggeredOnlyByRestoreReinitOrSchemeChange(t, ctx, fixture)
	})
	t.Run("EpochBumpVoidsOnlyAddressesOfVersionsNoLongerServed", func(t *testing.T) {
		conformance.RunLegacyEpochBumpVoidsOnlyAddressesOfVersionsNoLongerServed(t, ctx, fixture)
	})
}

// legacyNewEpochDoltFixture wires this leg's *DoltStore into the R20 epoch
// contract. Like newExpectedRevisionDoltFixture, this fixture needs none of
// roleFixtureKit's CreateIssue/CreateWisp/QueryScalar/CountHistory hooks —
// it has no fields for them — so it does not route through
// newDoltRoleFixtureKit at all; IssuePrefix is set directly, the same value
// the kit would have set verbatim.
func legacyNewEpochDoltFixture(t *testing.T, prefix string) (conformance.LegacyEpochFixture, context.Context, func()) {
	t.Helper()
	store, storeCleanup := setupTestStore(t)
	ctx, cancel := testContext(t)
	stop := func() {
		cancel()
		storeCleanup()
	}
	return conformance.LegacyEpochFixture{
		IssuePrefix:       prefix,
		CurrentEpoch:      legacyEpochDoltCurrentEpoch(store),
		BumpEpoch:         legacyEpochDoltBumpEpoch(store),
		MintUnderEpoch:    legacyEpochDoltMintUnderEpoch(store),
		StillServes:       legacyEpochDoltStillServes(store),
		Resolve:           legacyEpochDoltResolve(store),
		CurrentAddressFor: legacyEpochDoltCurrentAddressFor(store),
	}, ctx, stop
}

func legacyEpochDoltCurrentEpoch(store *DoltStore) func(ctx context.Context, storeID string) (int, error) {
	return func(ctx context.Context, storeID string) (int, error) {
		return store.CurrentEpoch(ctx, storeID)
	}
}

func legacyEpochDoltBumpEpoch(store *DoltStore) func(ctx context.Context, storeID string, trigger conformance.EpochBumpTrigger) (int, error) {
	return func(ctx context.Context, storeID string, trigger conformance.EpochBumpTrigger) (int, error) {
		return store.BumpEpoch(ctx, storeID, trigger.String())
	}
}

func legacyEpochDoltMintUnderEpoch(store *DoltStore) func(ctx context.Context, storeID, id string) (conformance.Address, error) {
	return func(ctx context.Context, storeID, id string) (conformance.Address, error) {
		address, err := store.MintUnderEpoch(ctx, storeID, id)
		if err != nil {
			return "", err
		}
		return conformance.Address(address), nil
	}
}

func legacyEpochDoltStillServes(store *DoltStore) func(ctx context.Context, storeID string, address conformance.Address) (bool, error) {
	return func(ctx context.Context, storeID string, address conformance.Address) (bool, error) {
		return store.StillServes(ctx, storeID, string(address))
	}
}

func legacyEpochDoltResolve(store *DoltStore) func(ctx context.Context, storeID string, address conformance.Address) (conformance.RetentionAnswer, error) {
	return func(ctx context.Context, storeID string, address conformance.Address) (conformance.RetentionAnswer, error) {
		result, err := store.Resolve(ctx, storeID, string(address))
		if err != nil {
			return conformance.RetentionAnswer{}, err
		}
		return legacyEpochResolveResultToConformance(result), nil
	}
}

func legacyEpochDoltCurrentAddressFor(store *DoltStore) func(ctx context.Context, storeID string, oldAddress conformance.Address) (conformance.Address, error) {
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
