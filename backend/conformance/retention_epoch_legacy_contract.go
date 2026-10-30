package conformance

import (
	"context"
	"testing"
)

// Released epoch contract retained alongside the incoming survival contract.
type LegacyEpochFixture struct {
	IssuePrefix string

	// CurrentEpoch reports storeID's current epoch generation. A nil hook
	// means this backend has no epoch concept yet, and cases that need one
	// skip with that reason.
	CurrentEpoch func(ctx context.Context, storeID string) (int, error)

	// BumpEpoch advances storeID's epoch for the given trigger, returning
	// the new epoch. A nil hook means this backend has no epoch concept
	// yet, and cases that need one skip with that reason.
	BumpEpoch func(ctx context.Context, storeID string, trigger EpochBumpTrigger) (int, error)

	// MintUnderEpoch mints a fresh Version for id under storeID's current
	// epoch, returning its Address. A nil hook means this backend has no
	// way to mint a Version under a known epoch, and cases that need one
	// skip with that reason.
	MintUnderEpoch func(ctx context.Context, storeID, id string) (Address, error)

	// StillServes reports whether storeID still serves the Version named by
	// a prior-epoch address, despite the epoch having moved on. A nil hook
	// means this backend cannot report that, and the case that needs it
	// skips with that reason.
	StillServes func(ctx context.Context, storeID string, address Address) (bool, error)

	// Resolve is the SAME RetentionResolve type RetentionFixture uses — an
	// epoch bump's effect on an Address is observed through the identical
	// read path as an ordinary removal. A nil hook means this backend does
	// not yet report retention state, and cases that need one skip with
	// that reason.
	Resolve RetentionResolve

	// CurrentAddressFor reports the Address a still-served, prior-epoch
	// Version now resolves to under the new epoch. A nil hook means this
	// backend cannot report that, and the case that needs it skips with
	// that reason.
	CurrentAddressFor func(ctx context.Context, storeID string, oldAddress Address) (Address, error)
}

func RunLegacyAnEpochBumpIsTriggeredOnlyByRestoreReinitOrSchemeChange(t *testing.T, ctx context.Context, fixture LegacyEpochFixture) {
	t.Helper()
	if fixture.BumpEpoch == nil {
		t.Skip("this backend has no epoch concept yet (BumpEpoch is nil)")
	}
	if fixture.CurrentEpoch == nil {
		t.Skip("this backend cannot report its current epoch (CurrentEpoch is nil)")
	}
	store := legacyEpochStore(fixture, "triggers")

	for _, trigger := range []EpochBumpTrigger{
		EpochBumpTriggerRestore,
		EpochBumpTriggerDestructiveReinit,
		EpochBumpTriggerTokenSchemeChange,
	} {
		before, err := fixture.CurrentEpoch(ctx, store)
		if err != nil {
			t.Fatalf("CurrentEpoch before %s: %v", trigger, err)
		}
		after, err := fixture.BumpEpoch(ctx, store, trigger)
		if err != nil {
			t.Fatalf("BumpEpoch(%s): %v", trigger, err)
		}
		if after <= before {
			t.Errorf("BumpEpoch(%s) = %d, want an epoch greater than the pre-bump value %d", trigger, after, before)
		}
		observed, err := fixture.CurrentEpoch(ctx, store)
		if err != nil {
			t.Fatalf("CurrentEpoch after %s: %v", trigger, err)
		}
		if observed != after {
			t.Errorf("CurrentEpoch after BumpEpoch(%s) = %d, want the bumped value %d", trigger, observed, after)
		}
	}
}

func RunLegacyEpochBumpVoidsOnlyAddressesOfVersionsNoLongerServed(t *testing.T, ctx context.Context, fixture LegacyEpochFixture) {
	t.Helper()
	if fixture.MintUnderEpoch == nil {
		t.Skip("this backend has no way to mint a Version under a known epoch (MintUnderEpoch is nil)")
	}
	if fixture.BumpEpoch == nil {
		t.Skip("this backend has no epoch concept yet (BumpEpoch is nil)")
	}
	if fixture.StillServes == nil {
		t.Skip("this backend cannot report whether it still serves a prior-epoch Version (StillServes is nil)")
	}
	if fixture.Resolve == nil {
		t.Skip("this backend does not yet report retention state (Resolve is nil)")
	}
	if fixture.CurrentAddressFor == nil {
		t.Skip("this backend cannot report a still-served Version's new Address (CurrentAddressFor is nil)")
	}
	store := legacyEpochStore(fixture, "voids")

	addrA, err := fixture.MintUnderEpoch(ctx, store, "record-a")
	if err != nil {
		t.Fatalf("MintUnderEpoch(record-a): %v", err)
	}
	addrB, err := fixture.MintUnderEpoch(ctx, store, "record-b")
	if err != nil {
		t.Fatalf("MintUnderEpoch(record-b): %v", err)
	}

	newEpoch, err := fixture.BumpEpoch(ctx, store, EpochBumpTriggerRestore)
	if err != nil {
		t.Fatalf("BumpEpoch: %v", err)
	}

	// record-a's Version survives the transition by being minted again under
	// the new epoch — MintUnderEpoch's own contract is "mints a fresh
	// Version for id under storeID's CURRENT epoch", so a second call here
	// carries record-a's id forward. record-b is never minted again, so it
	// does not survive. This makes both halves of R20-n deterministically in
	// play instead of something to be discovered after the fact: addrA must
	// still be served through the retained mapping, addrB must not, and a
	// backend that gets the direction backwards must not be able to pass by
	// having this case discover-then-swap around whatever it observes.
	if _, err := fixture.MintUnderEpoch(ctx, store, "record-a"); err != nil {
		t.Fatalf("MintUnderEpoch(record-a) [carries it into the new epoch]: %v", err)
	}

	servesA, err := fixture.StillServes(ctx, store, addrA)
	if err != nil {
		t.Fatalf("StillServes(%s): %v", addrA, err)
	}
	if !servesA {
		t.Errorf("StillServes(record-a's prior-epoch address) = false, want true: record-a was minted again under the new epoch, so its prior address must still resolve through R20-n's retained mapping")
	}
	servesB, err := fixture.StillServes(ctx, store, addrB)
	if err != nil {
		t.Fatalf("StillServes(%s): %v", addrB, err)
	}
	if servesB {
		t.Errorf("StillServes(record-b's prior-epoch address) = true, want false: record-b was never minted again, so it did not survive the epoch transition")
	}

	// stillServed/noLongerServed are pinned to addrA/addrB directly, not
	// discovered from servesA/servesB and swapped to match
	// (gastownhall/beads#6664, bee-ghosttrack review 5268699223, item B6):
	// the assertions immediately above already require servesA true and
	// servesB false, so a backend that gets the direction backwards fails
	// there and must not be able to recover by having this line quietly
	// swap stillServed/noLongerServed to whatever it actually observed —
	// that would let the Resolve/CurrentAddressFor assertions below pass
	// against the wrong address pair instead of catching the reversal too.
	stillServed, noLongerServed := addrA, addrB

	voidedAnswer, err := fixture.Resolve(ctx, store, noLongerServed)
	if err != nil {
		t.Fatalf("Resolve(no-longer-served): %v", err)
	}
	if voidedAnswer.Restriction != RestrictionGoneReorganization {
		t.Errorf("Resolve(a prior-epoch Address no longer served) = %s, want RestrictionGoneReorganization (R20-n)", voidedAnswer.Restriction)
	}
	if voidedAnswer.Epoch == nil {
		t.Fatal("Resolve(a prior-epoch Address no longer served).Epoch = nil, want the current epoch populated (R20-n: the answer must name the current epoch, not just Restriction alone)")
	} else if *voidedAnswer.Epoch != newEpoch {
		t.Errorf("Resolve(a prior-epoch Address no longer served).Epoch = %d, want the current epoch %d", *voidedAnswer.Epoch, newEpoch)
	}

	keptAnswer, err := fixture.Resolve(ctx, store, stillServed)
	if err != nil {
		t.Fatalf("Resolve(still-served): %v", err)
	}
	if keptAnswer.Restriction != RestrictionLive {
		t.Errorf("Resolve(a prior-epoch Address the backend still serves) = %s, want RestrictionLive (R20-n's exception)", keptAnswer.Restriction)
	}

	newAddr, err := fixture.CurrentAddressFor(ctx, store, stillServed)
	if err != nil {
		t.Fatalf("CurrentAddressFor(%s): %v", stillServed, err)
	}
	if newAddr == "" {
		t.Error("CurrentAddressFor(a still-served prior-epoch Address) returned an empty Address, want a real one under the new epoch")
	}

	// Second phase (B1): a later, non-contiguous re-mint must not
	// retroactively retain an address across a gap epoch. Its own store —
	// legacyEpochStore(fixture, "voids-gap") — keeps it from seeing or being
	// satisfied by the addrA/addrB state the first phase set up above.
	// Epoch numbers are captured from BumpEpoch's own return values rather
	// than assumed: store_epoch is one physical counter per database
	// (epoch_cas.go's package doc), not per-store, so this store's epoch
	// numbering continues from wherever the shared counter already stood.
	gapStore := legacyEpochStore(fixture, "voids-gap")

	gapAddr1, err := fixture.MintUnderEpoch(ctx, gapStore, "record-gap")
	if err != nil {
		t.Fatalf("MintUnderEpoch(record-gap) [gap phase, first epoch]: %v", err)
	}

	// Bump once and deliberately do NOT re-mint record-gap here — this is
	// the gap. The point under test is what a query sees after the SECOND
	// bump, not this one, so this phase does not assert anything yet.
	if _, err := fixture.BumpEpoch(ctx, gapStore, EpochBumpTriggerRestore); err != nil {
		t.Fatalf("BumpEpoch [gap phase, leaving the gap]: %v", err)
	}

	// Bump again, THEN re-mint record-gap — reusing the same id after the
	// gap, not carrying it forward through the epoch in between.
	gapNewEpoch, err := fixture.BumpEpoch(ctx, gapStore, EpochBumpTriggerRestore)
	if err != nil {
		t.Fatalf("BumpEpoch [gap phase, after the gap]: %v", err)
	}
	if _, err := fixture.MintUnderEpoch(ctx, gapStore, "record-gap"); err != nil {
		t.Fatalf("MintUnderEpoch(record-gap) [gap phase, re-mint after the gap]: %v", err)
	}

	gapServes, err := fixture.StillServes(ctx, gapStore, gapAddr1)
	if err != nil {
		t.Fatalf("StillServes(%s) [gap phase]: %v", gapAddr1, err)
	}
	if gapServes {
		t.Errorf("StillServes(record-gap's first-epoch address) = true, want false: record-gap's mapping lapsed at the gap epoch (never re-minted there) before being reused later — a later, non-contiguous re-mint must not retroactively retain the first-epoch address (R20-n requires the mapping to be CARRIED FORWARD, not merely reused after a gap)")
	}

	gapAnswer, err := fixture.Resolve(ctx, gapStore, gapAddr1)
	if err != nil {
		t.Fatalf("Resolve(%s) [gap phase]: %v", gapAddr1, err)
	}
	if gapAnswer.Restriction != RestrictionGoneReorganization {
		t.Errorf("Resolve(record-gap's first-epoch address, after a lapsed-then-reused mapping) = %s, want RestrictionGoneReorganization", gapAnswer.Restriction)
	}
	if gapAnswer.Epoch == nil {
		t.Fatal("Resolve(record-gap's first-epoch address).Epoch = nil, want the current epoch populated")
	} else if *gapAnswer.Epoch != gapNewEpoch {
		t.Errorf("Resolve(record-gap's first-epoch address).Epoch = %d, want the current epoch %d", *gapAnswer.Epoch, gapNewEpoch)
	}

	// Third phase (continuous multi-hop): a mapping carried forward through
	// EVERY intervening epoch — not just one hop — must still resolve Live.
	// The first phase above only bumps once, so (per this function's own
	// doc comment) it cannot structurally distinguish a backend that checks
	// an unbroken chain from one that merely checks bare existence of a row
	// at the current epoch; the second (gap) phase proves a GAP voids
	// correctly, but proves nothing about a continuous chain longer than
	// one hop — a single-hop-only arithmetic guard (epoch-priorEpoch == 1)
	// would pass both existing phases and still be wrong. This phase bumps
	// two more epochs past the minting epoch and re-mints record-continuous
	// at EVERY intervening epoch, leaving no gap anywhere, then asserts the
	// ORIGINAL, first-epoch address still resolves Live.
	continuousStore := legacyEpochStore(fixture, "voids-continuous")

	continuousAddr1, err := fixture.MintUnderEpoch(ctx, continuousStore, "record-continuous")
	if err != nil {
		t.Fatalf("MintUnderEpoch(record-continuous) [continuous phase, first epoch]: %v", err)
	}

	if _, err := fixture.BumpEpoch(ctx, continuousStore, EpochBumpTriggerRestore); err != nil {
		t.Fatalf("BumpEpoch [continuous phase, second epoch]: %v", err)
	}
	if _, err := fixture.MintUnderEpoch(ctx, continuousStore, "record-continuous"); err != nil {
		t.Fatalf("MintUnderEpoch(record-continuous) [continuous phase, carry into second epoch]: %v", err)
	}

	continuousFinalEpoch, err := fixture.BumpEpoch(ctx, continuousStore, EpochBumpTriggerRestore)
	if err != nil {
		t.Fatalf("BumpEpoch [continuous phase, third epoch]: %v", err)
	}
	if _, err := fixture.MintUnderEpoch(ctx, continuousStore, "record-continuous"); err != nil {
		t.Fatalf("MintUnderEpoch(record-continuous) [continuous phase, carry into third epoch]: %v", err)
	}

	continuousServes, err := fixture.StillServes(ctx, continuousStore, continuousAddr1)
	if err != nil {
		t.Fatalf("StillServes(%s) [continuous phase]: %v", continuousAddr1, err)
	}
	if !continuousServes {
		t.Errorf("StillServes(record-continuous's first-epoch address) = false, want true: record-continuous was re-minted at EVERY intervening epoch with no gap, so its first-epoch address must still resolve through R20-n's retained mapping across the full chain, not just one hop")
	}

	continuousAnswer, err := fixture.Resolve(ctx, continuousStore, continuousAddr1)
	if err != nil {
		t.Fatalf("Resolve(%s) [continuous phase]: %v", continuousAddr1, err)
	}
	if continuousAnswer.Restriction != RestrictionLive {
		t.Errorf("Resolve(record-continuous's first-epoch address, carried forward through 2 intervening bumps with no gap) = %s, want RestrictionLive: a backend that only checks one hop (epoch-priorEpoch == 1) wrongly reports this GoneReorganization", continuousAnswer.Restriction)
	}
	if continuousAnswer.Epoch == nil {
		t.Fatal("Resolve(record-continuous's first-epoch address).Epoch = nil, want the current epoch populated")
	} else if *continuousAnswer.Epoch != continuousFinalEpoch {
		t.Errorf("Resolve(record-continuous's first-epoch address).Epoch = %d, want the current epoch %d", *continuousAnswer.Epoch, continuousFinalEpoch)
	}

	continuousNewAddr, err := fixture.CurrentAddressFor(ctx, continuousStore, continuousAddr1)
	if err != nil {
		t.Fatalf("CurrentAddressFor(%s) [continuous phase]: %v", continuousAddr1, err)
	}
	if continuousNewAddr == "" {
		t.Error("CurrentAddressFor(record-continuous's first-epoch address) [continuous phase] returned an empty Address, want a real one under the current epoch")
	}
}

func legacyEpochStore(fixture LegacyEpochFixture, tag string) string {
	return fixture.IssuePrefix + "-epoch-store-" + tag
}
