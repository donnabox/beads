package uow

import (
	"context"
	"fmt"
	"testing"

	"github.com/steveyegge/beads/backend/conformance"
	storeops "github.com/steveyegge/beads/internal/storage/issueops"
)

// The production/default store retains the released epoch contract.
func TestLegacyEpochContract(t *testing.T) {
	ctx := context.Background()
	fixture := legacyNewUOWEpochFixture(t, ctx, "epch")

	t.Run("AnEpochBumpIsTriggeredOnlyByRestoreReinitOrSchemeChange", func(t *testing.T) {
		conformance.RunLegacyAnEpochBumpIsTriggeredOnlyByRestoreReinitOrSchemeChange(t, ctx, fixture)
	})
	t.Run("EpochBumpVoidsOnlyAddressesOfVersionsNoLongerServed", func(t *testing.T) {
		conformance.RunLegacyEpochBumpVoidsOnlyAddressesOfVersionsNoLongerServed(t, ctx, fixture)
	})
}

// legacyNewUOWEpochFixture wires this leg's UnitOfWorkProvider into the R20 epoch
// contract. Like newUOWExpectedRevisionFixture, this fixture needs none of
// roleFixtureKit's CreateIssue/CreateWisp/QueryScalar/CountHistory hooks —
// it has no fields for them — so it does not route through
// newUOWRoleFixtureKit at all.
func legacyNewUOWEpochFixture(t *testing.T, ctx context.Context, prefix string) conformance.LegacyEpochFixture {
	t.Helper()
	provider := newUOWRoleFixtureProvider(t, ctx, prefix)
	return conformance.LegacyEpochFixture{
		IssuePrefix:       prefix,
		CurrentEpoch:      legacyEpochUOWCurrentEpoch(provider),
		BumpEpoch:         legacyEpochUOWBumpEpoch(provider),
		MintUnderEpoch:    legacyEpochUOWMintUnderEpoch(provider),
		StillServes:       legacyEpochUOWStillServes(provider),
		Resolve:           legacyEpochUOWResolve(provider),
		CurrentAddressFor: legacyEpochUOWCurrentAddressFor(provider),
	}
}

// legacyEpochUOWRunner recovers the *baseUOW a given attempt's UnitOfWork actually
// is, the same type assertion expectedRevisionUOWMutateOutside makes: Tx.
// Runner() is unexported outside this package and none of the six epoch
// operations has a domain/use-case method to reach it through instead.
func legacyEpochUOWRunner(uw UnitOfWork) (storeops.DBTX, error) {
	base, ok := uw.(*baseUOW)
	if !ok {
		return nil, fmt.Errorf("epoch CAS: unit of work %T does not expose the runner the epoch Tx functions need", uw)
	}
	return base.tx.Runner(), nil
}

func legacyEpochUOWCurrentEpoch(provider UnitOfWorkProvider) func(ctx context.Context, storeID string) (int, error) {
	return func(ctx context.Context, storeID string) (int, error) {
		return RunTxRead(ctx, provider, func(ctx context.Context, uw UnitOfWork) (int, error) {
			runner, err := legacyEpochUOWRunner(uw)
			if err != nil {
				return 0, err
			}
			return storeops.CurrentEpochInTx(ctx, runner, storeID)
		})
	}
}

func legacyEpochUOWBumpEpoch(provider UnitOfWorkProvider) func(ctx context.Context, storeID string, trigger conformance.EpochBumpTrigger) (int, error) {
	return func(ctx context.Context, storeID string, trigger conformance.EpochBumpTrigger) (int, error) {
		return RunTxResult(ctx, provider, func(ctx context.Context, uw UnitOfWork) (int, string, error) {
			runner, err := legacyEpochUOWRunner(uw)
			if err != nil {
				return 0, "", err
			}
			epoch, err := storeops.BumpEpochInTx(ctx, runner, storeID, trigger.String())
			if err != nil {
				return 0, "", err
			}
			return epoch, fmt.Sprintf("bd: bump epoch for %s (%s)", storeID, trigger.String()), nil
		})
	}
}

func legacyEpochUOWMintUnderEpoch(provider UnitOfWorkProvider) func(ctx context.Context, storeID, id string) (conformance.Address, error) {
	return func(ctx context.Context, storeID, id string) (conformance.Address, error) {
		address, err := RunTxResult(ctx, provider, func(ctx context.Context, uw UnitOfWork) (string, string, error) {
			runner, err := legacyEpochUOWRunner(uw)
			if err != nil {
				return "", "", err
			}
			address, err := storeops.MintUnderEpochInTx(ctx, runner, storeID, id)
			if err != nil {
				return "", "", err
			}
			return address, fmt.Sprintf("bd: mint %s under epoch for %s", id, storeID), nil
		})
		if err != nil {
			return "", err
		}
		return conformance.Address(address), nil
	}
}

func legacyEpochUOWStillServes(provider UnitOfWorkProvider) func(ctx context.Context, storeID string, address conformance.Address) (bool, error) {
	return func(ctx context.Context, storeID string, address conformance.Address) (bool, error) {
		return RunTxRead(ctx, provider, func(ctx context.Context, uw UnitOfWork) (bool, error) {
			runner, err := legacyEpochUOWRunner(uw)
			if err != nil {
				return false, err
			}
			return storeops.StillServesInTx(ctx, runner, storeID, string(address))
		})
	}
}

func legacyEpochUOWResolve(provider UnitOfWorkProvider) func(ctx context.Context, storeID string, address conformance.Address) (conformance.RetentionAnswer, error) {
	return func(ctx context.Context, storeID string, address conformance.Address) (conformance.RetentionAnswer, error) {
		result, err := RunTxRead(ctx, provider, func(ctx context.Context, uw UnitOfWork) (storeops.EpochResolveResult, error) {
			runner, err := legacyEpochUOWRunner(uw)
			if err != nil {
				return storeops.EpochResolveResult{}, err
			}
			return storeops.ResolveEpochInTx(ctx, runner, storeID, string(address))
		})
		if err != nil {
			return conformance.RetentionAnswer{}, err
		}
		return legacyEpochResolveResultToConformance(result), nil
	}
}

func legacyEpochUOWCurrentAddressFor(provider UnitOfWorkProvider) func(ctx context.Context, storeID string, oldAddress conformance.Address) (conformance.Address, error) {
	return func(ctx context.Context, storeID string, oldAddress conformance.Address) (conformance.Address, error) {
		address, err := RunTxResult(ctx, provider, func(ctx context.Context, uw UnitOfWork) (string, string, error) {
			runner, err := legacyEpochUOWRunner(uw)
			if err != nil {
				return "", "", err
			}
			address, err := storeops.CurrentAddressForInTx(ctx, runner, storeID, string(oldAddress))
			if err != nil {
				return "", "", err
			}
			return address, fmt.Sprintf("bd: current address for %s (%s)", oldAddress, storeID), nil
		})
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
