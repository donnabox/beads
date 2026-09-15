package authority

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/steveyegge/beads/graphops"
)

// These limits bound local administrative metadata, not BDP protocol values.
const MaxWitnessBytes = 64 * 1024
const MaxWitnessTokenBytes = 256

var (
	ErrInvalidRecord      = errors.New("invalid witness record")
	ErrWrongInstallation  = errors.Join(errors.New("wrong witness installation"), graphops.ErrNotAuthority)
	ErrWitnessBusy        = errors.New("witness busy")
	ErrWitnessPending     = errors.Join(errors.New("witness transition pending"), graphops.ErrNotAuthority)
	ErrWitnessUnverified  = errors.Join(errors.New("witness unverified"), graphops.ErrNotAuthority)
	ErrEvidence           = errors.New("witness evidence required or conflicting")
	ErrWitnessUnsupported = errors.New("witness operation unsupported")
	ErrGuardMisuse        = errors.New("witness guard misuse")
	ErrPersistence        = errors.New("witness persistence failed")
)

// Effect describes the authority envelope only; Git hygiene is reported separately.
type Effect string

const (
	Unchanged       Effect = "no_durable_change"
	VisibleUnknown  Effect = "replacement_visible_durability_unknown"
	RetainedPending Effect = "durable_pending"
)

// PersistenceError never includes record contents in its message.
type PersistenceError struct {
	Effect                Effect
	cause                 error
	HygieneMayHaveChanged bool
}

func (e *PersistenceError) Error() string { return "witness persistence: " + string(e.Effect) }
func (e *PersistenceError) Unwrap() error { return errors.Join(ErrPersistence, e.cause) }

// WitnessFields is copied on construction and observation. It contains only values.
type WitnessFields struct {
	ScopeURL     string `json:"scope_url"`
	AuthorityID  string `json:"authority_id"`
	Epoch        uint64 `json:"epoch"`
	LedgerSeq    uint64 `json:"ledger_seq"`
	LedgerHash   string `json:"ledger_hash"`
	StateVersion string `json:"state_version"`
	StateCommit  string `json:"state_commit"`
	Unverified   bool   `json:"unverified"`
	GrantedAt    string `json:"granted_at"`
}
type Witness struct{ fields WitnessFields }

func NewWitness(v WitnessFields) (Witness, error) {
	if err := validateWitness(v); err != nil {
		return Witness{}, err
	}
	return Witness{v}, nil
}
func (w Witness) Fields() WitnessFields { return w.fields }
func validateWitness(w WitnessFields) error {
	if !validURL(w.ScopeURL) || !hexLength(w.AuthorityID, 32) || w.Epoch == 0 || w.LedgerSeq == 0 || w.LedgerSeq > graphops.MaxLedgerSeq || !hexLength(w.LedgerHash, 64) || !hexLength(w.StateVersion, 64) || !validToken(w.StateCommit) || !validTime(w.GrantedAt) {
		return ErrInvalidRecord
	}
	return nil
}
func hexLength(s string, n int) bool {
	if len(s) != n {
		return false
	}
	for _, c := range s {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}
func validToken(s string) bool {
	if s == "" || len(s) > MaxWitnessTokenBytes || !utf8.ValidString(s) {
		return false
	}
	for _, r := range s {
		if unicode.IsControl(r) {
			return false
		}
	}
	return true
}
func validURL(s string) bool {
	return len(s) <= MaxWitnessBytes && graphops.ValidatePersistedScopeURL(s) == nil
}
func validTime(s string) bool {
	t, e := time.Parse(time.RFC3339Nano, s)
	return e == nil && t.After(time.Unix(0, 0)) && t.UTC().Format(time.RFC3339Nano) == s
}
func digest(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func digestValue(v any) string {
	b, e := json.Marshal(v)
	if e != nil {
		return ""
	}
	return digest(b)
}

// Presence encodes absence explicitly; zero values are invalid.
type Presence string

const (
	Absent  Presence = "absent"
	Present Presence = "present"
)

type OptionalToken struct {
	Presence Presence `json:"presence"`
	Value    string   `json:"value"`
}
type ScopeState struct {
	Presence    Presence `json:"presence"`
	URL         string   `json:"url"`
	AuthorityID string   `json:"authority_id"`
	Epoch       uint64   `json:"epoch"`
}
type LedgerState struct {
	Presence Presence `json:"presence"`
	Seq      uint64   `json:"seq"`
	Hash     string   `json:"hash"`
}
type LeaseState struct {
	Renewer     string   `json:"renewer"`
	GrantedAt   string   `json:"granted_at"`
	HeartbeatAt string   `json:"heartbeat_at"`
	Presence    Presence `json:"presence"`
	Holder      string   `json:"holder"`
	ScopeURL    string   `json:"scope_url"`
	AuthorityID string   `json:"authority_id"`
	Epoch       uint64   `json:"epoch"`
	Fence       string   `json:"fence"`
	ExpiresAt   string   `json:"expires_at"`
}
type TableState struct {
	Presence Presence `json:"presence"`
	Schema   string   `json:"schema"`
	Content  string   `json:"content"`
}

// TableStates uses the owning specification's fixed eight-table order.
type TableStates [8]TableState

func validOptional(v OptionalToken) bool {
	return v.Presence == Absent && v.Value == "" || v.Presence == Present && validToken(v.Value)
}
func validScope(v ScopeState) bool {
	return v.Presence == Absent && v.URL == "" && v.AuthorityID == "" && v.Epoch == 0 || v.Presence == Present && validURL(v.URL) && hexLength(v.AuthorityID, 32) && v.Epoch > 0
}
func validLedger(v LedgerState) bool {
	return v.Presence == Absent && v.Seq == 0 && v.Hash == "" || v.Presence == Present && v.Seq > 0 && v.Seq <= graphops.MaxLedgerSeq && hexLength(v.Hash, 64)
}
func validLease(v LeaseState) bool {
	return v.Presence == Absent && v.Renewer == "" && v.GrantedAt == "" && v.HeartbeatAt == "" && v.Holder == "" && v.ScopeURL == "" && v.AuthorityID == "" && v.Epoch == 0 && v.Fence == "" && v.ExpiresAt == "" || v.Presence == Present && hexLength(v.Renewer, 32) && validTime(v.GrantedAt) && validTime(v.HeartbeatAt) && hexLength(v.Holder, 64) && validURL(v.ScopeURL) && hexLength(v.AuthorityID, 32) && v.Epoch > 0 && hexLength(v.Fence, 32) && validTime(v.ExpiresAt)
}
func validTables(v TableStates) bool {
	for _, t := range v {
		if !(t.Presence == Absent && t.Schema == "" && t.Content == "" || t.Presence == Present && validToken(t.Schema) && validToken(t.Content)) {
			return false
		}
	}
	return true
}
func completeTables(v TableStates) bool {
	if !validTables(v) {
		return false
	}
	for _, t := range v {
		if t.Presence != Present {
			return false
		}
	}
	return true
}

type MutationOperation string

const (
	UpdateMutation    MutationOperation = "update"
	AllocateMutation  MutationOperation = "allocate"
	TombstoneMutation MutationOperation = "tombstone"
)

func validMutationOperation(o MutationOperation) bool {
	return o == UpdateMutation || o == AllocateMutation || o == TombstoneMutation
}

type Kind string

const (
	Mint        Kind = "mint"
	Promote     Kind = "promote"
	Rotate      Kind = "rotate"
	Install     Kind = "install"
	Mutation    Kind = "mutation"
	LedgerApply Kind = "ledger_apply"
	Adopt       Kind = "adopt"
)

type PromoteMode string

const (
	SelfRegrant PromoteMode = "self_regrant"
	Steal       PromoteMode = "steal"
)

type RotateCause string

const (
	ExplicitRotate           RotateCause = "explicit_rotate"
	RestoreWithoutContinuity RotateCause = "restore_without_continuity"
)

type Lane string

const (
	Ordinary        Lane = "ordinary"
	RestoreRecovery Lane = "restore_recovery"
)

type Topology string

const (
	SharedDatabase Topology = "shared_database"
	Solo           Topology = "solo"
)

type WorkspaceOwnership string

const (
	SharedWorkspace    WorkspaceOwnership = "shared"
	ExclusiveWorkspace WorkspaceOwnership = "exclusive_stopped"
)

// ConfigurationIntent carries only the one permitted scoped key; no path or executable text.
type ConfigurationIntent struct {
	Previous OptionalToken `json:"previous"`
	Desired  string        `json:"desired"`
}

// IntentFields contains desired operands only. Provider pre-state belongs to Preparation.
type IntentFields struct {
	Kind           Kind                 `json:"kind"`
	ScopeURL       string               `json:"scope_url"`
	ArtifactDigest string               `json:"artifact_digest"`
	Operation      MutationOperation    `json:"operation"`
	PayloadDigest  string               `json:"payload_digest"`
	PromoteMode    PromoteMode          `json:"promote_mode"`
	RotateCause    RotateCause          `json:"rotate_cause"`
	Lane           Lane                 `json:"lane"`
	ExpectedLocal  string               `json:"expected_local"`
	ExpectedSource string               `json:"expected_source"`
	Configuration  *ConfigurationIntent `json:"configuration"`
}
type Intent struct{ encoded string }

func NewIntent(f IntentFields) (Intent, error) {
	if err := validateIntent(f); err != nil {
		return Intent{}, err
	}
	b, e := json.Marshal(f)
	return Intent{string(b)}, e
}
func (i Intent) Fields() IntentFields {
	var f IntentFields
	if json.Unmarshal([]byte(i.encoded), &f) != nil {
		return IntentFields{}
	}
	return f
}
func validateIntent(f IntentFields) error {
	if f.Configuration != nil {
		c := f.Configuration
		if !validURL(c.Desired) || !(c.Previous.Presence == Absent && c.Previous.Value == "" || c.Previous.Presence == Present && validURL(c.Previous.Value)) {
			return ErrInvalidRecord
		}
	}
	clean := f
	clean.Kind = ""
	switch f.Kind {
	case Mint:
		if !validURL(f.ScopeURL) || !hexLength(f.ArtifactDigest, 64) || f.Configuration != nil && f.Configuration.Desired != f.ScopeURL {
			return ErrInvalidRecord
		}
		clean.ScopeURL = ""
		clean.ArtifactDigest = ""
		clean.Configuration = nil
	case Promote:
		if f.PromoteMode != SelfRegrant && f.PromoteMode != Steal {
			return ErrInvalidRecord
		}
		clean.PromoteMode = ""
	case Rotate:
		if !validURL(f.ScopeURL) || (f.RotateCause != ExplicitRotate && f.RotateCause != RestoreWithoutContinuity) || f.Configuration == nil || f.Configuration.Desired != f.ScopeURL {
			return ErrInvalidRecord
		}
		clean.ScopeURL = ""
		clean.RotateCause = ""
		clean.Configuration = nil
	case Install:
		if !hexLength(f.ArtifactDigest, 64) {
			return ErrInvalidRecord
		}
		clean.ArtifactDigest = ""
	case Mutation:
		if !validMutationOperation(f.Operation) || !hexLength(f.PayloadDigest, 64) {
			return ErrInvalidRecord
		}
		clean.Operation = ""
		clean.PayloadDigest = ""
	case LedgerApply:
		if !hexLength(f.ArtifactDigest, 64) || (f.Lane != Ordinary && f.Lane != RestoreRecovery) {
			return ErrInvalidRecord
		}
		clean.ArtifactDigest = ""
		clean.Lane = ""
	case Adopt:
		if !validToken(f.ExpectedLocal) || !validToken(f.ExpectedSource) || f.ExpectedLocal == f.ExpectedSource {
			return ErrInvalidRecord
		}
		clean.ExpectedLocal = ""
		clean.ExpectedSource = ""
		clean.Configuration = nil
	default:
		return ErrWitnessUnsupported
	}
	if clean != (IntentFields{}) {
		return ErrInvalidRecord
	}
	return nil
}
