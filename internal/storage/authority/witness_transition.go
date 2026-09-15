package authority

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
)

// Transition borrows its Guard and is bound to the exact persisted operation.
// It must not outlive that guard; successful Begin is not an engine permit.
type Transition struct {
	guard       *Guard
	opID, token string
}

func ownership(ctx context.Context, e Evidence) (Ownership, error) {
	if e == nil {
		return Ownership{}, ErrEvidence
	}
	if err := ctx.Err(); err != nil {
		return Ownership{}, err
	}
	o, err := e.ObserveOwnership(ctx)
	if err != nil {
		return Ownership{}, err
	}
	if !validOwnership(o) {
		return Ownership{}, ErrEvidence
	}
	if err = ctx.Err(); err != nil {
		return Ownership{}, err
	}
	return o, nil
}
func sameOwnership(ctx context.Context, e Evidence, o Ownership) error {
	now, err := ownership(ctx, e)
	if err != nil {
		return err
	}
	if now != o {
		return ErrEvidence
	}
	return nil
}
func binding(s Snapshot, o Ownership, kind string, operands any) RequestBinding {
	return RequestBinding{digestValue(struct {
		Kind, Token string
		Ownership   Ownership
		Operands    any
	}{kind, s.Token(), o, operands}), s.Token(), o}
}
func admitBegin(s Snapshot, key string, i IntentFields) error {
	if s.Pending() {
		return ErrWitnessPending
	}
	if s.Absent() {
		if i.Kind == Mint || i.Kind == Promote || i.Kind == Rotate || i.Kind == LedgerApply && i.Lane == RestoreRecovery {
			return nil
		}
		return ErrEvidence
	}
	if s.InstallationKey() != key {
		if i.Kind == Rotate || i.Kind == Promote && i.PromoteMode == Steal {
			return nil
		}
		return ErrWrongInstallation
	}
	if i.Kind == Mint {
		return ErrEvidence
	}
	w, ok := s.Witness()
	if !ok {
		return ErrInvalidRecord
	}
	if w.fields.Unverified && !(i.Kind == Promote || i.Kind == Rotate || i.Kind == LedgerApply && i.Lane == RestoreRecovery) {
		return ErrWitnessUnverified
	}
	return nil
}
func checkURLOverrides(i IntentFields) error {
	if i.Configuration != nil && (os.Getenv("BDP_SCOPE_URL") != "" || os.Getenv("BD_BDP_SCOPE_URL") != "") {
		return ErrEvidence
	}
	return nil
}
func (g *Guard) Begin(ctx context.Context, intent Intent, e Evidence) (*Transition, error) {
	if err := g.enter(ctx); err != nil {
		return nil, err
	}
	defer g.mu.Unlock()
	i := intent.Fields()
	if err := validateIntent(i); err != nil {
		return nil, err
	}
	if err := checkURLOverrides(i); err != nil {
		return nil, err
	}
	s, err := g.read(ctx)
	if err != nil {
		return nil, err
	}
	if err = admitBegin(s, g.manager.key, i); err != nil {
		return nil, err
	}
	generation, err := nextGeneration(s)
	if err != nil {
		return nil, err
	}
	var random [16]byte
	if _, err = rand.Read(random[:]); err != nil {
		return nil, err
	}
	opID := hex.EncodeToString(random[:])
	o, err := ownership(ctx, e)
	if err != nil {
		return nil, err
	}
	r := PrepareRequest{intent: intent, opID: opID, currentKey: g.manager.key, priorKey: s.InstallationKey(), prior: s}
	r.binding = binding(s, o, "prepare", struct {
		Intent                            IntentFields
		OperationID, CurrentKey, PriorKey string
	}{i, opID, g.manager.key, s.InstallationKey()})
	p, err := e.PrepareTransition(ctx, r)
	if err != nil {
		return nil, err
	}
	if err = sameOwnership(ctx, e, o); err != nil {
		return nil, err
	}
	if p.binding != r.binding {
		return nil, ErrEvidence
	}
	prepared := p.Fields()
	if err = validatePreparation(i, prepared, g.manager.key, o); err != nil {
		return nil, err
	}
	if err = prepareMatchesPrior(s, i, prepared); err != nil {
		return nil, err
	}
	rec := s.record()
	rec.Format = 1
	rec.Generation = generation
	prior := priorBinding{Absent, "", s.Token()}
	if s.Absent() {
		rec.InstallationKey = g.manager.key
	} else {
		prior = priorBinding{Present, s.InstallationKey(), s.Token()}
	}
	rec.Pending = &pendingRecord{g.manager.key, prior, opID, Begun, i, prepared, o, ""}
	next, err := g.persistOwned(ctx, s, &rec, func() error {
		if err := admitBegin(s, g.manager.key, i); err != nil {
			return err
		}
		return sameOwnership(ctx, e, o)
	})
	if err != nil {
		return nil, err
	}
	if err = sameOwnership(ctx, e, o); err != nil {
		return nil, &PersistenceError{RetainedPending, err, true}
	}
	if err = g.fresh(ctx, next); err != nil {
		return nil, &PersistenceError{RetainedPending, err, true}
	}
	return &Transition{g, opID, next.Token()}, nil
}
func prepareMatchesPrior(s Snapshot, i IntentFields, p PreparationFields) error {
	return prepareMatchesPriorFields(s.record().Witness, i, p)
}
func prepareMatchesPriorFields(w *WitnessFields, i IntentFields, p PreparationFields) error {
	if w != nil {
		// Restore recovery may observe rewound selected state; its manifest proof owns
		// continuity. Scope/authority lineage still cannot be substituted.
		if p.PreScope.URL != w.ScopeURL || p.PreScope.AuthorityID != w.AuthorityID {
			return ErrEvidence
		}
		if i.Kind != Rotate && i.Kind != Promote && !(i.Kind == LedgerApply && i.Lane == RestoreRecovery) {
			if p.PreScope.Epoch != w.Epoch || p.PreLedger.Seq != w.LedgerSeq || p.PreLedger.Hash != w.LedgerHash || p.PreHead.Value != w.StateCommit {
				return ErrEvidence
			}
		}
	}
	return nil
}
func (t *Transition) read(ctx context.Context) (Snapshot, error) {
	if t == nil || t.guard == nil {
		return Snapshot{}, ErrGuardMisuse
	}
	s, err := t.guard.read(ctx)
	if err != nil {
		return Snapshot{}, err
	}
	if !s.Pending() || s.OperationID() != t.opID || s.Token() != t.token {
		return Snapshot{}, ErrGuardMisuse
	}
	return s, nil
}
func inspect(ctx context.Context, s Snapshot, e Evidence) (InspectionFields, Ownership, error) {
	o, err := ownership(ctx, e)
	if err != nil {
		return InspectionFields{}, o, err
	}
	p := s.record().Pending
	if p == nil {
		return InspectionFields{}, o, ErrEvidence
	}
	if o.Database != p.Observation.Database || o.Branch != p.Observation.Branch || o.Topology != p.Observation.Topology || validatePreparation(p.Intent, p.Prepared, p.OperationKey, o) != nil {
		return InspectionFields{}, o, ErrEvidence
	}
	r := InspectRequest{binding(s, o, "inspect", s.OperationID()), s}
	f, err := e.InspectTransition(ctx, r)
	if err != nil {
		return InspectionFields{}, o, err
	}
	if err = sameOwnership(ctx, e, o); err != nil {
		return InspectionFields{}, o, err
	}
	if f.binding != r.binding {
		return InspectionFields{}, o, ErrEvidence
	}
	if _, err = NewInspection(r, f.fields); err != nil {
		return InspectionFields{}, o, err
	}
	return f.fields, o, nil
}
func exactPre(p *pendingRecord, f InspectionFields) bool {
	q := p.Prepared
	return f.State == ExactPre && f.Complete == Verified && f.StateVersion == q.PreStateVersion.Value && f.Tables == q.WorkingTables && f.Scope == q.PreScope && f.Ledger == q.PreLedger && f.Lease == q.PreLease && f.Head == q.PreHead && f.OperationID == "" && f.OperationCommit == "" && f.ManifestDigest == "" && f.AntiReuseDigest == "" && f.NoConfigurationEffects == Verified
}
func completeResult(p *pendingRecord, f InspectionFields) bool {
	q := p.Prepared
	if f.Complete != Verified || f.Tables != q.ResultTables || f.Lease != q.ResultLease || f.Head.Presence != Present || f.NoConfigurationEffects != "" {
		return false
	}
	if p.Intent.Kind == Adopt {
		return f.State == CompleteAdopt && f.Scope == q.Adopt.SourceScope && f.Head.Value == q.Adopt.SourceCommit && f.OperationID == "" && f.OperationCommit == "" && f.ManifestDigest == "" && f.AntiReuseDigest == q.Adopt.AntiReuseDigest
	}
	w := q.ResultWitness
	if w == nil || f.StateVersion != w.StateVersion || f.Scope != (ScopeState{Present, w.ScopeURL, w.AuthorityID, w.Epoch}) || f.Ledger != (LedgerState{Present, w.LedgerSeq, w.LedgerHash}) {
		return false
	}
	switch {
	case p.Intent.Kind == Promote && p.Intent.PromoteMode == SelfRegrant:
		return f.State == CompleteSelfRegrant && f.Head == q.PreHead && f.OperationID == "" && f.OperationCommit == "" && f.ManifestDigest == "" && f.AntiReuseDigest == ""
	case p.Intent.Kind == LedgerApply:
		return f.State == CompleteLedgerApply && f.OperationID == "" && f.OperationCommit == "" && f.ManifestDigest == q.LedgerApply.ManifestDigest && f.AntiReuseDigest == q.LedgerApply.AntiReuseDigest
	default:
		return f.State == CompleteEvent && f.OperationID == p.OperationID && validToken(f.OperationCommit) && f.Head.Value == f.OperationCommit && (p.OperationCommit == "" || p.OperationCommit == f.OperationCommit) && f.ManifestDigest == "" && f.AntiReuseDigest == ""
	}
}
func phaseRank(p Phase) int {
	switch p {
	case Begun:
		return 0
	case LocalCommitted:
		return 1
	case Published:
		return 2
	case ConfigWritten:
		return 3
	}
	return -1
}
func applyConfiguration(ctx context.Context, g *Guard, p *pendingRecord, c ConfigurationApplier) error {
	if p.Intent.Configuration == nil {
		return nil
	}
	if c == nil {
		return ErrEvidence
	}
	if err := checkURLOverrides(p.Intent); err != nil {
		return err
	}
	r := ConfigurationRequest{p.OperationID, g.manager.dir, digestValue(p.Intent.Configuration), *p.Intent.Configuration}
	receipt, err := c.ApplyScopeURLIntent(ctx, r)
	if err != nil {
		return err
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	if receipt.operationID != r.operationID || receipt.intentDigest != r.intentDigest || receipt.desired != r.intent.Desired {
		return ErrEvidence
	}
	return nil
}
func (t *Transition) SetPhase(ctx context.Context, phase Phase, e Evidence, c ConfigurationApplier) error {
	if t == nil {
		return ErrGuardMisuse
	}
	if err := t.guard.enter(ctx); err != nil {
		return err
	}
	defer t.guard.mu.Unlock()
	return t.setPhase(ctx, phase, e, c)
}
func (t *Transition) setPhase(ctx context.Context, phase Phase, e Evidence, c ConfigurationApplier) error {
	s, err := t.read(ctx)
	if err != nil {
		return err
	}
	rank := phaseRank(phase)
	current := phaseRank(s.Phase())
	if rank < 0 || rank < current || rank > current+1 {
		return ErrGuardMisuse
	}
	f, o, err := inspect(ctx, s, e)
	if err != nil {
		return err
	}
	rec := s.record()
	if !completeResult(rec.Pending, f) && !(phase == Begun && exactPre(rec.Pending, f)) {
		return ErrEvidence
	}
	if phase == ConfigWritten {
		if err = sameOwnership(ctx, e, o); err != nil {
			return err
		}
		if err = applyConfiguration(ctx, t.guard, rec.Pending, c); err != nil {
			return err
		}
		if err = sameOwnership(ctx, e, o); err != nil {
			return err
		}
	}
	if phase == Begun {
		return t.guard.flush(ctx, s)
	}
	if rec.Pending.Phase != phase {
		rec.Generation, err = nextGeneration(s)
		if err != nil {
			return err
		}
		rec.Pending.Phase = phase
		rec.Pending.OperationCommit = f.OperationCommit
		next, err := t.guard.persistOwned(ctx, s, &rec, func() error { return sameOwnership(ctx, e, o) })
		if err != nil {
			return err
		}
		t.token = next.Token()
		return nil
	}
	if err = sameOwnership(ctx, e, o); err != nil {
		return err
	}
	return t.guard.flush(ctx, s)
}
func (t *Transition) Finalize(ctx context.Context, e Evidence, c ConfigurationApplier) error {
	if t == nil {
		return ErrGuardMisuse
	}
	if err := t.guard.enter(ctx); err != nil {
		return err
	}
	defer t.guard.mu.Unlock()
	return t.finalize(ctx, e, c)
}
func (t *Transition) finalize(ctx context.Context, e Evidence, c ConfigurationApplier) error {
	s, err := t.read(ctx)
	if err != nil {
		return err
	}
	if s.Phase() != ConfigWritten {
		return ErrGuardMisuse
	}
	f, o, err := inspect(ctx, s, e)
	if err != nil {
		return err
	}
	rec := s.record()
	if !completeResult(rec.Pending, f) {
		return ErrEvidence
	}
	if err = sameOwnership(ctx, e, o); err != nil {
		return err
	}
	if err = applyConfiguration(ctx, t.guard, rec.Pending, c); err != nil {
		return err
	}
	if err = sameOwnership(ctx, e, o); err != nil {
		return err
	}
	var result *envelope
	if rec.Pending.Intent.Kind != Adopt {
		w := *rec.Pending.Prepared.ResultWitness
		w.StateCommit = f.Head.Value
		w.Unverified = rec.Witness != nil && rec.Witness.Unverified
		rec.Witness = &w
		rec.InstallationKey = t.guard.manager.key
		rec.Pending = nil
		rec.Generation, err = nextGeneration(s)
		if err != nil {
			return err
		}
		result = &rec
	}
	_, err = t.guard.persistOwned(ctx, s, result, func() error { return sameOwnership(ctx, e, o) })
	if err == nil {
		t.token = ""
	}
	return err
}
func (t *Transition) Abandon(ctx context.Context, e Evidence) error {
	if t == nil {
		return ErrGuardMisuse
	}
	if err := t.guard.enter(ctx); err != nil {
		return err
	}
	defer t.guard.mu.Unlock()
	return t.abandon(ctx, e)
}
func (t *Transition) abandon(ctx context.Context, e Evidence) error {
	s, err := t.read(ctx)
	if err != nil {
		return err
	}
	if s.Phase() != Begun {
		return ErrEvidence
	}
	f, o, err := inspect(ctx, s, e)
	if err != nil {
		return err
	}
	rec := s.record()
	if !exactPre(rec.Pending, f) {
		return ErrEvidence
	}
	var result *envelope
	if rec.Witness != nil {
		rec.Pending = nil
		rec.Generation, err = nextGeneration(s)
		if err != nil {
			return err
		}
		result = &rec
	}
	_, err = t.guard.persistOwned(ctx, s, result, func() error { return sameOwnership(ctx, e, o) })
	if err == nil {
		t.token = ""
	}
	return err
}

// Recover makes only local file/config decisions from fresh facts. It never
// invokes or retries a graph operation, restores a lease, or chooses a temp file.
func (g *Guard) Recover(ctx context.Context, e Evidence, c ConfigurationApplier) error {
	if err := g.enter(ctx); err != nil {
		return err
	}
	defer g.mu.Unlock()
	s, err := g.read(ctx)
	if err != nil {
		return err
	}
	if !s.Pending() {
		if !s.Absent() && s.InstallationKey() != g.manager.key {
			return ErrWrongInstallation
		}
		return g.flush(ctx, s)
	}
	f, _, err := inspect(ctx, s, e)
	if err != nil {
		return err
	}
	t := &Transition{g, s.OperationID(), s.Token()}
	p := s.record().Pending
	if exactPre(p, f) {
		return t.abandon(ctx, e)
	}
	if !completeResult(p, f) {
		return errors.Join(ErrEvidence, g.mark(ctx, s))
	}
	for _, phase := range []Phase{LocalCommitted, Published, ConfigWritten} {
		current, err := t.read(ctx)
		if err != nil {
			return err
		}
		if phaseRank(current.Phase()) < phaseRank(phase) {
			if err = t.setPhase(ctx, phase, e, c); err != nil {
				return err
			}
		}
	}
	return t.finalize(ctx, e, c)
}
