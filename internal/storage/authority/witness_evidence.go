package authority

import (
	"context"
	"encoding/json"
	"github.com/steveyegge/beads/graphops"
)

// Ownership is a provider-owned observation, not a grant manufactured by this manager.
// Every field must remain equal across a call; the provider separately enforces lifetime.
type Ownership struct {
	Database           string             `json:"database"`
	Branch             string             `json:"branch"`
	SessionGeneration  string             `json:"session_generation"`
	ObservationVersion uint64             `json:"observation_version"`
	Topology           Topology           `json:"topology"`
	Workspace          WorkspaceOwnership `json:"workspace"`
}

func validOwnership(o Ownership) bool {
	return validToken(o.Database) && validToken(o.Branch) && validToken(o.SessionGeneration) && o.ObservationVersion > 0 && (o.Topology == SharedDatabase || o.Topology == Solo) && (o.Workspace == SharedWorkspace || o.Workspace == ExclusiveWorkspace) && (o.Topology != Solo || o.Workspace == ExclusiveWorkspace)
}

// RequestBinding ties provider output to this exact request and physical observation.
type RequestBinding struct {
	Digest    string    `json:"digest"`
	Token     string    `json:"token"`
	Ownership Ownership `json:"ownership"`
}
type PrepareRequest struct {
	binding                    RequestBinding
	intent                     Intent
	opID, currentKey, priorKey string
	prior                      Snapshot
}

func (r PrepareRequest) Binding() RequestBinding { return r.binding }
func (r PrepareRequest) Intent() Intent          { return r.intent }
func (r PrepareRequest) OperationID() string     { return r.opID }
func (r PrepareRequest) InstallationKey() string { return r.currentKey }
func (r PrepareRequest) PriorKey() string        { return r.priorKey }
func (r PrepareRequest) Prior() Snapshot         { return r.prior }

// VerifiedFact names a provider verification, not permission for the file manager to infer it.
type VerifiedFact string

const Verified VerifiedFact = "verified"

type MintPreparation struct {
	CatalogDigest string       `json:"catalog_digest"`
	LawfulTarget  VerifiedFact `json:"lawful_target"`
}
type PromotePreparation struct {
	Mode         PromoteMode  `json:"mode"`
	SameDatabase VerifiedFact `json:"same_database"`
	RenewedLease LeaseState   `json:"renewed_lease"`
}
type RotatePreparation struct {
	Cause              RotateCause `json:"cause"`
	RefusedURL         string      `json:"refused_url"`
	PreservationDigest string      `json:"preservation_digest"`
	RestoreDecision    string      `json:"restore_decision"`
	NewLease           LeaseState  `json:"new_lease"`
}
type InstallPreparation struct {
	ArtifactDigest string       `json:"artifact_digest"`
	Validated      VerifiedFact `json:"validated"`
}
type MutationPreparation struct {
	Operation     MutationOperation `json:"operation"`
	PayloadDigest string            `json:"payload_digest"`
	Authorized    VerifiedFact      `json:"authorized"`
	LeaseWindow   VerifiedFact      `json:"lease_window"`
}
type LedgerPreparation struct {
	ScopeURL        string       `json:"scope_url"`
	ReplayedScope   ScopeState   `json:"replayed_scope"`
	ManifestDigest  string       `json:"manifest_digest"`
	Lineage         string       `json:"lineage"`
	FirstSeq        uint64       `json:"first_seq"`
	LastSeq         uint64       `json:"last_seq"`
	Predecessor     string       `json:"predecessor"`
	Head            string       `json:"head"`
	Chain           VerifiedFact `json:"chain"`
	AntiReuseDigest string       `json:"anti_reuse_digest"`
	Lane            Lane         `json:"lane"`
	Regrant         LeaseState   `json:"regrant"`
}
type AdoptPreparation struct {
	LocalCommit       string       `json:"local_commit"`
	SourceCommit      string       `json:"source_commit"`
	SourceScope       ScopeState   `json:"source_scope"`
	SourceTables      TableStates  `json:"source_tables"`
	SourceValid       VerifiedFact `json:"source_valid"`
	CleanWorkingState VerifiedFact `json:"clean_working_state"`
	AntiReuseDigest   string       `json:"anti_reuse_digest"`
	OldLeaseCleanup   VerifiedFact `json:"old_lease_cleanup"`
}

// PreparationFields is a closed tagged union. Constructors and disk admission share validation.
// ResultWitness.StateCommit is empty until terminal inspection supplies the actual
// committed HEAD. Event-free SelfRegrant instead retains the known pre-HEAD.
// No new op_commit is predicted or accepted before dispatch.
type PreparationFields struct {
	PreStateVersion OptionalToken        `json:"pre_state_version"`
	PreLineage      string               `json:"pre_lineage"`
	PreHead         OptionalToken        `json:"pre_head"`
	PreLedger       LedgerState          `json:"pre_ledger"`
	PreScope        ScopeState           `json:"pre_scope"`
	PreLease        LeaseState           `json:"pre_lease"`
	WorkingTables   TableStates          `json:"working_tables"`
	ResultTables    TableStates          `json:"result_tables"`
	ResultLease     LeaseState           `json:"result_lease"`
	ResultWitness   *WitnessFields       `json:"result_witness"`
	Mint            *MintPreparation     `json:"mint"`
	Promote         *PromotePreparation  `json:"promote"`
	Rotate          *RotatePreparation   `json:"rotate"`
	Install         *InstallPreparation  `json:"install"`
	Mutation        *MutationPreparation `json:"mutation"`
	LedgerApply     *LedgerPreparation   `json:"ledger_apply"`
	Adopt           *AdoptPreparation    `json:"adopt"`
}
type Preparation struct {
	binding RequestBinding
	encoded string
}

func NewPreparation(r PrepareRequest, f PreparationFields) (Preparation, error) {
	if err := validatePreparation(r.intent.Fields(), f, r.currentKey, r.binding.Ownership); err != nil {
		return Preparation{}, err
	}
	b, e := json.Marshal(f)
	return Preparation{r.binding, string(b)}, e
}
func (p Preparation) Fields() PreparationFields {
	var f PreparationFields
	if json.Unmarshal([]byte(p.encoded), &f) != nil {
		return PreparationFields{}
	}
	return f
}

func validatePreparation(i IntentFields, p PreparationFields, key string, o Ownership) error {
	if validateIntent(i) != nil || !validOwnership(o) || !hexLength(key, 64) || !validOptional(p.PreHead) || !validOptional(p.PreStateVersion) || p.PreStateVersion.Presence == Present && !hexLength(p.PreStateVersion.Value, 64) || !validLedger(p.PreLedger) || !validScope(p.PreScope) || !validLease(p.PreLease) || !validLease(p.ResultLease) || !validTables(p.WorkingTables) || !completeTables(p.ResultTables) {
		return ErrEvidence
	}
	arms := 0
	for _, v := range []bool{p.Mint != nil, p.Promote != nil, p.Rotate != nil, p.Install != nil, p.Mutation != nil, p.LedgerApply != nil, p.Adopt != nil} {
		if v {
			arms++
		}
	}
	if arms != 1 {
		return ErrEvidence
	}
	if i.Kind == Adopt {
		if p.ResultWitness != nil || p.ResultLease.Presence != Absent {
			return ErrEvidence
		}
	} else {
		if p.ResultWitness == nil {
			return ErrEvidence
		}
		w := *p.ResultWitness
		w.StateCommit = "prepared"
		if validateWitness(w) != nil || w.Unverified || p.ResultLease.Holder != key || !leaseMatchesWitness(p.ResultLease, w) {
			return ErrEvidence
		}
		if !(i.Kind == Promote && i.PromoteMode == SelfRegrant) && p.ResultWitness.StateCommit != "" {
			return ErrEvidence
		}
	}
	if p.PreLease.Presence == Present && (p.PreScope.Presence != Present || p.PreLease.ScopeURL != p.PreScope.URL || p.PreLease.AuthorityID != p.PreScope.AuthorityID || p.PreLease.Epoch != p.PreScope.Epoch) {
		return ErrEvidence
	}
	if i.Kind == Mint {
		if p.PreLineage != "" {
			return ErrEvidence
		}
	} else if !hexLength(p.PreLineage, 64) || p.PreStateVersion.Presence != Present {
		return ErrEvidence
	}
	if i.Kind != Mint && p.PreScope.Presence != Present {
		return ErrEvidence
	}
	if i.Kind != Mint && p.PreLedger.Presence != Present {
		return ErrEvidence
	}
	if i.Kind == Mint || i.Kind == Install || i.Kind == LedgerApply || i.Kind == Adopt {
		if o.Workspace != ExclusiveWorkspace {
			return ErrEvidence
		}
	}
	switch i.Kind {
	case Mint:
		if p.Mint == nil || p.Mint.CatalogDigest != i.ArtifactDigest || p.Mint.LawfulTarget != Verified || p.PreScope.Presence != Absent || p.PreLedger.Presence != Absent || p.PreLease.Presence != Absent || p.ResultWitness.ScopeURL != i.ScopeURL {
			return ErrEvidence
		}
	case Promote:
		q := p.Promote
		if q == nil || q.Mode != i.PromoteMode || p.PreLease.Presence != Present || !validLease(q.RenewedLease) || q.RenewedLease.Presence != Present || q.RenewedLease.Holder != key || q.RenewedLease.Fence == p.PreLease.Fence {
			return ErrEvidence
		}
		if p.ResultWitness.ScopeURL != p.PreScope.URL || p.ResultWitness.AuthorityID != p.PreScope.AuthorityID {
			return ErrEvidence
		}
		if q.Mode == SelfRegrant {
			if q.SameDatabase != "" || p.ResultWitness.StateVersion != p.PreStateVersion.Value || p.PreLease.Holder != key || p.ResultWitness.Epoch != p.PreScope.Epoch || p.ResultWitness.LedgerSeq != p.PreLedger.Seq || p.ResultWitness.LedgerHash != p.PreLedger.Hash || p.ResultTables != p.WorkingTables || p.PreHead.Presence != Present || p.ResultWitness.StateCommit != p.PreHead.Value {
				return ErrEvidence
			}
		} else if q.SameDatabase != Verified || p.PreLease.Holder == key || p.PreScope.Epoch == ^uint64(0) || p.ResultWitness.Epoch != p.PreScope.Epoch+1 {
			return ErrEvidence
		}
		if q.RenewedLease != p.ResultLease || !leaseMatchesWitness(q.RenewedLease, *p.ResultWitness) {
			return ErrEvidence
		}
	case Rotate:
		q := p.Rotate
		if q == nil || q.Cause != i.RotateCause || q.RefusedURL != p.PreScope.URL || !hexLength(q.PreservationDigest, 64) || p.ResultWitness.ScopeURL != i.ScopeURL || p.ResultWitness.ScopeURL == p.PreScope.URL || p.ResultWitness.AuthorityID == p.PreScope.AuthorityID || !validLease(q.NewLease) || q.NewLease != p.ResultLease || q.NewLease.Holder != key || !leaseMatchesWitness(q.NewLease, *p.ResultWitness) {
			return ErrEvidence
		}
		if q.Cause == ExplicitRotate && q.RestoreDecision != "" || q.Cause == RestoreWithoutContinuity && q.RestoreDecision != "in_state_without_continuity" && q.RestoreDecision != "none" {
			return ErrEvidence
		}
	case Install:
		if p.Install == nil || p.Install.ArtifactDigest != i.ArtifactDigest || p.Install.Validated != Verified || p.PreLease.Holder != key {
			return ErrEvidence
		}
	case Mutation:
		if p.Mutation == nil || p.Mutation.Operation != i.Operation || p.Mutation.PayloadDigest != i.PayloadDigest || p.Mutation.Authorized != Verified || p.Mutation.LeaseWindow != Verified || p.PreLease.Holder != key {
			return ErrEvidence
		}
	case LedgerApply:
		q := p.LedgerApply
		if q == nil || q.ScopeURL != p.PreScope.URL || !validScope(q.ReplayedScope) || q.ReplayedScope.Presence != Present || q.ReplayedScope != (ScopeState{Present, p.ResultWitness.ScopeURL, p.ResultWitness.AuthorityID, p.ResultWitness.Epoch}) || q.ManifestDigest != i.ArtifactDigest || q.Lineage != p.PreLineage || !hexLength(q.Lineage, 64) || q.FirstSeq == 0 || q.LastSeq < q.FirstSeq || q.LastSeq > graphops.MaxLedgerSeq || q.Predecessor != p.PreLedger.Hash || q.FirstSeq != p.PreLedger.Seq+1 || !hexLength(q.Head, 64) || q.Chain != Verified || !hexLength(q.AntiReuseDigest, 64) || q.Lane != i.Lane || !validLease(q.Regrant) || q.Regrant != p.ResultLease || q.Regrant.Holder != key || !leaseMatchesWitness(q.Regrant, *p.ResultWitness) || p.ResultWitness.LedgerSeq != q.LastSeq || p.ResultWitness.LedgerHash != q.Head {
			return ErrEvidence
		}
	case Adopt:
		q := p.Adopt
		if q == nil || q.LocalCommit != i.ExpectedLocal || q.SourceCommit != i.ExpectedSource || p.PreHead.Presence != Present || q.LocalCommit != p.PreHead.Value || !validScope(q.SourceScope) || q.SourceScope.Presence != Present || q.SourceScope.URL != p.PreScope.URL || q.SourceScope.AuthorityID == p.PreScope.AuthorityID || !completeTables(q.SourceTables) || q.SourceTables != p.ResultTables || q.SourceValid != Verified || q.CleanWorkingState != Verified || !hexLength(q.AntiReuseDigest, 64) || q.OldLeaseCleanup != Verified {
			return ErrEvidence
		}
	default:
		return ErrWitnessUnsupported
	}
	if i.Kind == Install || i.Kind == Mutation {
		if p.ResultWitness.ScopeURL != p.PreScope.URL || p.ResultWitness.AuthorityID != p.PreScope.AuthorityID || p.ResultWitness.Epoch != p.PreScope.Epoch {
			return ErrEvidence
		}
	}
	if i.Kind != Adopt && p.PreLease.Presence == Present && p.ResultLease.Fence == p.PreLease.Fence {
		return ErrEvidence
	}
	if i.Kind == Mutation || i.Kind == Install || i.Kind == Rotate || i.Kind == Promote && i.PromoteMode == Steal {
		if p.ResultWitness.LedgerSeq <= p.PreLedger.Seq {
			return ErrEvidence
		}
	}
	return nil
}
func leaseMatchesWitness(l LeaseState, w WitnessFields) bool {
	return l.Presence == Present && l.ScopeURL == w.ScopeURL && l.AuthorityID == w.AuthorityID && l.Epoch == w.Epoch && l.GrantedAt == w.GrantedAt
}

type CompareRequest struct {
	binding            RequestBinding
	current, candidate Witness
}

func (r CompareRequest) Binding() RequestBinding { return r.binding }
func (r CompareRequest) Current() Witness        { return r.current }
func (r CompareRequest) Candidate() Witness      { return r.candidate }

type Relation string

const (
	Equal              Relation = "equal"
	CandidateContained Relation = "candidate_contained"
	CurrentContained   Relation = "current_contained"
	Fork               Relation = "fork"
	Unprovable         Relation = "unprovable"
)

type Comparison struct {
	binding  RequestBinding
	relation Relation
}

func NewComparison(r CompareRequest, v Relation) (Comparison, error) {
	switch v {
	case Equal, CandidateContained, CurrentContained, Fork, Unprovable:
		return Comparison{r.binding, v}, nil
	}
	return Comparison{}, ErrEvidence
}

type InspectRequest struct {
	binding  RequestBinding
	snapshot Snapshot
}

func (r InspectRequest) Binding() RequestBinding { return r.binding }
func (r InspectRequest) Snapshot() Snapshot      { return r.snapshot }

type InspectionState string

const (
	ExactPre            InspectionState = "exact_pre_no_effects"
	CompleteEvent       InspectionState = "complete_event"
	CompleteSelfRegrant InspectionState = "complete_self_regrant"
	CompleteLedgerApply InspectionState = "complete_ledger_apply"
	CompleteAdopt       InspectionState = "complete_adopt"
	UnknownState        InspectionState = "unknown"
)

// InspectionFields contains actual observed states; the manager compares them with recorded operands.
type InspectionFields struct {
	State                  InspectionState `json:"state"`
	StateVersion           string          `json:"state_version"`
	Tables                 TableStates     `json:"tables"`
	Scope                  ScopeState      `json:"scope"`
	Ledger                 LedgerState     `json:"ledger"`
	Lease                  LeaseState      `json:"lease"`
	Head                   OptionalToken   `json:"head"`
	OperationID            string          `json:"operation_id"`
	OperationCommit        string          `json:"operation_commit"`
	ManifestDigest         string          `json:"manifest_digest"`
	AntiReuseDigest        string          `json:"anti_reuse_digest"`
	NoConfigurationEffects VerifiedFact    `json:"no_configuration_effects"`
	Complete               VerifiedFact    `json:"complete"`
}
type Inspection struct {
	binding RequestBinding
	fields  InspectionFields
}

func NewInspection(r InspectRequest, f InspectionFields) (Inspection, error) {
	if !validTables(f.Tables) || !validScope(f.Scope) || !validLedger(f.Ledger) || !validLease(f.Lease) || !validOptional(f.Head) || f.Complete != Verified {
		return Inspection{}, ErrEvidence
	}
	switch f.State {
	case ExactPre, CompleteEvent, CompleteSelfRegrant, CompleteLedgerApply, CompleteAdopt, UnknownState:
	default:
		return Inspection{}, ErrEvidence
	}
	return Inspection{r.binding, f}, nil
}

type ClearRequest struct {
	binding            RequestBinding
	current, candidate Witness
}

func (r ClearRequest) Binding() RequestBinding { return r.binding }
func (r ClearRequest) Current() Witness        { return r.current }
func (r ClearRequest) Candidate() Witness      { return r.candidate }

type ClearReason string

const (
	ContinuityProved ClearReason = "restore_continuity"
	RotationProved   ClearReason = "completed_rotation"
)

type ClearProof struct {
	binding RequestBinding
	reason  ClearReason
}

func NewClearProof(r ClearRequest, reason ClearReason) (ClearProof, error) {
	if reason != ContinuityProved && reason != RotationProved {
		return ClearProof{}, ErrEvidence
	}
	return ClearProof{r.binding, reason}, nil
}

// Evidence has no implementation in this slice. Implementers must own the actual
// qualified provider observation; returning a valid value does not qualify one.
type Evidence interface {
	ObserveOwnership(context.Context) (Ownership, error)
	PrepareTransition(context.Context, PrepareRequest) (Preparation, error)
	CompareAdvance(context.Context, CompareRequest) (Comparison, error)
	InspectTransition(context.Context, InspectRequest) (Inspection, error)
	VerifyClearUnverified(context.Context, ClearRequest) (ClearProof, error)
}

// ConfigurationRequest binds the sole scoped key to an already verified operation.
type ConfigurationRequest struct {
	operationID, directory, intentDigest string
	intent                               ConfigurationIntent
}

func (r ConfigurationRequest) OperationID() string         { return r.operationID }
func (r ConfigurationRequest) Directory() string           { return r.directory }
func (r ConfigurationRequest) Intent() ConfigurationIntent { return r.intent }
func (r ConfigurationRequest) IntentDigest() string        { return r.intentDigest }

type ConfigurationReceipt struct{ operationID, intentDigest, desired string }

func NewConfigurationReceipt(r ConfigurationRequest, observed string) (ConfigurationReceipt, error) {
	if observed != r.intent.Desired {
		return ConfigurationReceipt{}, ErrEvidence
	}
	return ConfigurationReceipt{r.operationID, r.intentDigest, observed}, nil
}

type ConfigurationApplier interface {
	ApplyScopeURLIntent(context.Context, ConfigurationRequest) (ConfigurationReceipt, error)
}
