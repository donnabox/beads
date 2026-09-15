package authority

import "context"

// Advance updates only the state/ledger fields proven to descend from the
// retained witness. Equality of opaque tokens is never ancestry evidence.
func (g *Guard) Advance(ctx context.Context, candidate Witness, e Evidence) error {
	if err := g.enter(ctx); err != nil {
		return err
	}
	defer g.mu.Unlock()
	if err := validateWitness(candidate.fields); err != nil {
		return err
	}
	s, err := g.read(ctx)
	if err != nil {
		return err
	}
	current, err := activeForKey(s, g.manager.key)
	if err != nil {
		return err
	}
	if current.fields.Unverified {
		return ErrWitnessUnverified
	}
	want, have := candidate.fields, current.fields
	want.LedgerSeq = have.LedgerSeq
	want.LedgerHash = have.LedgerHash
	want.StateVersion = have.StateVersion
	want.StateCommit = have.StateCommit
	if want != have {
		return ErrEvidence
	}
	o, err := ownership(ctx, e)
	if err != nil {
		return err
	}
	r := CompareRequest{binding(s, o, "compare", candidate.fields), current, candidate}
	comparison, err := e.CompareAdvance(ctx, r)
	if err != nil {
		return err
	}
	if err = sameOwnership(ctx, e, o); err != nil {
		return err
	}
	if comparison.binding != r.binding {
		return ErrEvidence
	}
	switch comparison.relation {
	case Equal:
		if candidate.fields.LedgerSeq != have.LedgerSeq || candidate.fields.LedgerHash != have.LedgerHash || candidate.fields.StateCommit != have.StateCommit {
			return ErrEvidence
		}
		// The explicit state-equality proof also covers a changed selected-state
		// commitment. Preserve the already retained snapshot on equality.
		return g.flush(ctx, s)
	case CandidateContained:
		if candidate.fields.LedgerSeq > have.LedgerSeq || candidate.fields.LedgerSeq == have.LedgerSeq && candidate.fields.LedgerHash != have.LedgerHash {
			return ErrEvidence
		}
		return g.flush(ctx, s)
	case CurrentContained:
		if candidate.fields.LedgerSeq < have.LedgerSeq || candidate.fields.LedgerSeq == have.LedgerSeq && candidate.fields.LedgerHash != have.LedgerHash {
			return ErrEvidence
		}
		rec := s.record()
		rec.Generation, err = nextGeneration(s)
		if err != nil {
			return err
		}
		fields := candidate.fields
		rec.Witness = &fields
		_, err = g.persistOwned(ctx, s, &rec, func() error { return sameOwnership(ctx, e, o) })
		return err
	default:
		return ErrEvidence
	}
}

// ClearUnverified has no matching-HEAD shortcut. The provider must supply the
// independently qualified restore-continuity or completed-rotation proof.
func (g *Guard) ClearUnverified(ctx context.Context, e Evidence) error {
	if err := g.enter(ctx); err != nil {
		return err
	}
	defer g.mu.Unlock()
	s, err := g.read(ctx)
	if err != nil {
		return err
	}
	current, err := activeForKey(s, g.manager.key)
	if err != nil {
		return err
	}
	candidate := current
	candidate.fields.Unverified = false
	o, err := ownership(ctx, e)
	if err != nil {
		return err
	}
	r := ClearRequest{binding(s, o, "clear", candidate.fields), current, candidate}
	proof, err := e.VerifyClearUnverified(ctx, r)
	if err != nil {
		return err
	}
	if err = sameOwnership(ctx, e, o); err != nil {
		return err
	}
	if proof.binding != r.binding || (proof.reason != ContinuityProved && proof.reason != RotationProved) {
		return ErrEvidence
	}
	if !current.fields.Unverified {
		return g.flush(ctx, s)
	}
	rec := s.record()
	rec.Generation, err = nextGeneration(s)
	if err != nil {
		return err
	}
	fields := candidate.fields
	rec.Witness = &fields
	_, err = g.persistOwned(ctx, s, &rec, func() error { return sameOwnership(ctx, e, o) })
	return err
}
