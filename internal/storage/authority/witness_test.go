//go:build unix

package authority

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/beadsignore"
	"github.com/steveyegge/beads/internal/lockfile"
	"golang.org/x/sys/unix"
)

const testURL = "https://beads.example/acme/"
const testTime = "2026-09-14T01:02:03Z"

func repeated(c string, n int) string { return strings.Repeat(c, n) }
func testWitness() WitnessFields {
	return WitnessFields{testURL, repeated("a", 32), 1, 2, repeated("b", 64), repeated("c", 64), "before", false, testTime}
}
func testTables() TableStates {
	var v TableStates
	for n := range v {
		v[n] = TableState{Present, "schema", "root"}
	}
	return v
}
func testLease(key string, w WitnessFields, fence string) LeaseState {
	return LeaseState{Presence: Present, Holder: key, ScopeURL: w.ScopeURL, AuthorityID: w.AuthorityID, Epoch: w.Epoch, Fence: repeated(fence, 32), ExpiresAt: "2026-09-14T01:12:03Z", Renewer: repeated("7", 32), GrantedAt: w.GrantedAt, HeartbeatAt: testTime}
}
func testManager(t *testing.T) *Manager {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, ".beads")
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("BEADS_INSTALLATION_ID_FILE", filepath.Join(root, "user", "installation-id"))
	t.Setenv("BDP_SCOPE_URL", "")
	t.Setenv("BD_BDP_SCOPE_URL", "")
	m, err := New(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	return m
}
func saveActive(t *testing.T, m *Manager, foreign bool, unverified bool) Snapshot {
	t.Helper()
	w := testWitness()
	w.Unverified = unverified
	key := m.key
	if foreign {
		key = repeated("d", 64)
	}
	s, err := encodeEnvelope(envelope{1, key, 1, &w, nil})
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(m.path(), s.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	return s
}
func testGuard(t *testing.T, m *Manager) *Guard {
	t.Helper()
	g, err := m.Acquire(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if !g.closed {
			if err := g.Close(); err != nil {
				t.Error(err)
			}
		}
	})
	return g
}

// recordingEvidence describes disposable fixture facts. It is not a storage
// provider, an engine observation, or production evidence qualification.
type recordingEvidence struct {
	owner        Ownership
	observations int
	prepareCount int
	inspectCount int
	relation     Relation
	prepare      func(PrepareRequest) (Preparation, error)
	inspect      func(InspectRequest) (Inspection, error)
	observe      func(int, *Ownership) error
	compare      func(CompareRequest) (Comparison, error)
}

func recording() *recordingEvidence {
	return &recordingEvidence{owner: Ownership{"fixture-database", "main", "session-one", 1, Solo, ExclusiveWorkspace}, relation: CurrentContained}
}
func (e *recordingEvidence) ObserveOwnership(ctx context.Context) (Ownership, error) {
	e.observations++
	if e.observe != nil {
		if err := e.observe(e.observations, &e.owner); err != nil {
			return Ownership{}, err
		}
	}
	return e.owner, ctx.Err()
}
func (e *recordingEvidence) PrepareTransition(ctx context.Context, r PrepareRequest) (Preparation, error) {
	e.prepareCount++
	if e.prepare != nil {
		return e.prepare(r)
	}
	return NewPreparation(r, preparationFor(r))
}
func (e *recordingEvidence) CompareAdvance(ctx context.Context, r CompareRequest) (Comparison, error) {
	if e.compare != nil {
		return e.compare(r)
	}
	return NewComparison(r, e.relation)
}
func (e *recordingEvidence) InspectTransition(ctx context.Context, r InspectRequest) (Inspection, error) {
	e.inspectCount++
	if e.inspect != nil {
		return e.inspect(r)
	}
	return NewInspection(r, completionFor(r.Snapshot()))
}
func (e *recordingEvidence) VerifyClearUnverified(ctx context.Context, r ClearRequest) (ClearProof, error) {
	return NewClearProof(r, ContinuityProved)
}
func preparationFor(r PrepareRequest) PreparationFields {
	i := r.Intent().Fields()
	w := testWitness()
	if old, ok := r.Prior().Witness(); ok {
		w = old.Fields()
	}
	w.Unverified = false
	preLease := testLease(r.InstallationKey(), w, "1")
	if i.Kind == Promote && i.PromoteMode == Steal {
		preLease.Holder = repeated("e", 64)
	}
	p := PreparationFields{PreStateVersion: OptionalToken{Present, w.StateVersion}, PreLineage: repeated("4", 64), PreHead: OptionalToken{Present, w.StateCommit}, PreLedger: LedgerState{Present, w.LedgerSeq, w.LedgerHash}, PreScope: ScopeState{Present, w.ScopeURL, w.AuthorityID, w.Epoch}, PreLease: preLease, WorkingTables: testTables(), ResultTables: testTables()}
	result := w
	result.StateCommit = ""
	result.StateVersion = repeated("f", 64)
	result.LedgerSeq++
	result.LedgerHash = repeated("9", 64)
	switch i.Kind {
	case Mint:
		p.PreLineage = ""
		p.PreStateVersion = OptionalToken{Presence: Absent}
		p.PreScope = ScopeState{Presence: Absent}
		p.PreLedger = LedgerState{Presence: Absent}
		p.PreLease = LeaseState{Presence: Absent}
		p.PreHead = OptionalToken{Presence: Absent}
		for n := range p.WorkingTables {
			p.WorkingTables[n] = TableState{Presence: Absent}
		}
		result.ScopeURL = i.ScopeURL
		p.Mint = &MintPreparation{i.ArtifactDigest, Verified}
	case Promote:
		if i.PromoteMode == SelfRegrant {
			result = w
		} else {
			result.Epoch++
		}
		p.Promote = &PromotePreparation{Mode: i.PromoteMode}
		if i.PromoteMode == Steal {
			p.Promote.SameDatabase = Verified
		}
	case Rotate:
		result.ScopeURL = i.ScopeURL
		result.AuthorityID = repeated("2", 32)
		p.Rotate = &RotatePreparation{Cause: i.RotateCause, RefusedURL: w.ScopeURL, PreservationDigest: repeated("3", 64)}
		if i.RotateCause == RestoreWithoutContinuity {
			p.Rotate.RestoreDecision = "none"
		}
	case Install:
		p.Install = &InstallPreparation{i.ArtifactDigest, Verified}
	case Mutation:
		p.Mutation = &MutationPreparation{i.Operation, i.PayloadDigest, Verified, Verified}
	case LedgerApply:
		p.LedgerApply = &LedgerPreparation{ScopeURL: p.PreScope.URL, ReplayedScope: ScopeState{Present, result.ScopeURL, result.AuthorityID, result.Epoch}, ManifestDigest: i.ArtifactDigest, Lineage: repeated("4", 64), FirstSeq: w.LedgerSeq + 1, LastSeq: result.LedgerSeq, Predecessor: w.LedgerHash, Head: result.LedgerHash, Chain: Verified, AntiReuseDigest: repeated("5", 64), Lane: i.Lane}
	case Adopt:
		source := p.PreScope
		source.AuthorityID = repeated("6", 32)
		p.Adopt = &AdoptPreparation{i.ExpectedLocal, i.ExpectedSource, source, p.ResultTables, Verified, Verified, repeated("7", 64), Verified}
		p.ResultLease = LeaseState{Presence: Absent}
		return p
	}
	p.ResultWitness = &result
	p.ResultLease = testLease(r.InstallationKey(), result, "8")
	if p.Promote != nil {
		p.Promote.RenewedLease = p.ResultLease
	}
	if p.Rotate != nil {
		p.Rotate.NewLease = p.ResultLease
	}
	if p.LedgerApply != nil {
		p.LedgerApply.Regrant = p.ResultLease
	}
	return p
}
func completionFor(s Snapshot) InspectionFields {
	p := s.record().Pending
	q := p.Prepared
	f := InspectionFields{State: CompleteEvent, Tables: q.ResultTables, Lease: q.ResultLease, Complete: Verified, Head: OptionalToken{Present, "after"}, OperationID: p.OperationID, OperationCommit: "after"}
	if q.ResultWitness != nil {
		w := q.ResultWitness
		f.Scope = ScopeState{Present, w.ScopeURL, w.AuthorityID, w.Epoch}
		f.Ledger = LedgerState{Present, w.LedgerSeq, w.LedgerHash}
		f.StateVersion = w.StateVersion
	}
	switch {
	case p.Intent.Kind == Promote && p.Intent.PromoteMode == SelfRegrant:
		f.State = CompleteSelfRegrant
		f.Head = q.PreHead
		f.OperationID = ""
		f.OperationCommit = ""
	case p.Intent.Kind == LedgerApply:
		f.State = CompleteLedgerApply
		f.OperationID = ""
		f.OperationCommit = ""
		f.ManifestDigest = q.LedgerApply.ManifestDigest
		f.AntiReuseDigest = q.LedgerApply.AntiReuseDigest
	case p.Intent.Kind == Adopt:
		f.State = CompleteAdopt
		f.Head.Value = q.Adopt.SourceCommit
		f.Scope = q.Adopt.SourceScope
		f.Ledger = q.PreLedger
		f.OperationID = ""
		f.OperationCommit = ""
		f.AntiReuseDigest = q.Adopt.AntiReuseDigest
	}
	return f
}
func preFor(s Snapshot) InspectionFields {
	q := s.Preparation()
	return InspectionFields{State: ExactPre, StateVersion: q.PreStateVersion.Value, Tables: q.WorkingTables, Scope: q.PreScope, Ledger: q.PreLedger, Lease: q.PreLease, Head: q.PreHead, Complete: Verified, NoConfigurationEffects: Verified}
}
func intentFor(kind Kind, mode PromoteMode, lane Lane) Intent {
	f := IntentFields{Kind: kind}
	switch kind {
	case Mint:
		f.ScopeURL = testURL
		f.ArtifactDigest = repeated("a", 64)
	case Promote:
		f.PromoteMode = mode
	case Rotate:
		f.ScopeURL = "https://beads.example/rotated/"
		f.RotateCause = ExplicitRotate
		f.Configuration = &ConfigurationIntent{OptionalToken{Present, testURL}, f.ScopeURL}
	case Install:
		f.ArtifactDigest = repeated("a", 64)
	case Mutation:
		f.Operation = UpdateMutation
		f.PayloadDigest = repeated("a", 64)
	case LedgerApply:
		f.ArtifactDigest = repeated("a", 64)
		f.Lane = lane
	case Adopt:
		f.ExpectedLocal = "before"
		f.ExpectedSource = "source"
	}
	i, _ := NewIntent(f)
	return i
}
func TestWitnessMintRoundTrip(t *testing.T) {
	m := testManager(t)
	g := testGuard(t, m)
	e := recording()
	tr, err := g.Begin(context.Background(), intentFor(Mint, "", ""), e)
	if err != nil {
		t.Fatal(err)
	}
	s, err := m.Load(context.Background())
	if err != nil || !s.Pending() {
		t.Fatalf("pending %v %v", s, err)
	}
	if e.prepareCount != 1 || e.inspectCount != 0 {
		t.Fatal("preparation substituted")
	}
	for _, phase := range []Phase{LocalCommitted, Published, ConfigWritten} {
		if err = tr.SetPhase(context.Background(), phase, e, nil); err != nil {
			t.Fatal(err)
		}
	}
	if err = tr.Finalize(context.Background(), e, nil); err != nil {
		t.Fatal(err)
	}
	s, err = m.Load(context.Background())
	w, ok := s.Witness()
	if err != nil || !ok || s.Pending() || w.fields.StateCommit != "after" {
		t.Fatal("final witness", err)
	}
	if err = tr.Finalize(context.Background(), e, nil); !errors.Is(err, ErrGuardMisuse) {
		t.Fatal("spent handle", err)
	}
}
func TestWitnessPassiveAndImmutable(t *testing.T) {
	m := testManager(t)
	s, err := m.Load(context.Background())
	if err != nil || !s.Absent() {
		t.Fatal(err)
	}
	if _, err = os.Lstat(filepath.Join(m.dir, beadsignore.LockName)); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("Load created lock")
	}
	old := saveActive(t, m, false, false)
	s, err = m.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	b := s.Bytes()
	b[0] = 'x'
	w, _ := s.Witness()
	w.fields.ScopeURL = "changed"
	again, err := m.Load(context.Background())
	if err != nil || again.Token() != old.Token() {
		t.Fatal("mutable snapshot", err)
	}
	saveActive(t, m, true, false)
	if _, err = m.Load(context.Background()); !errors.Is(err, ErrWrongInstallation) {
		t.Fatal(err)
	}
}
func TestWitnessCodecAdmission(t *testing.T) {
	m := testManager(t)
	s := saveActive(t, m, false, false)
	b := s.Bytes()
	cases := map[string][]byte{
		"duplicate": bytes.Replace(b, []byte(`"format":1`), []byte(`"format":1,"\u0066ormat":1`), 1), "unknown": bytes.Replace(b, []byte(`"format":1`), []byte(`"format":1,"extra":false`), 1), "null flag": bytes.Replace(b, []byte(`"unverified":false`), []byte(`"unverified":null`), 1), "missing": bytes.Replace(b, []byte(`"unverified":false,`), nil, 1), "fraction": bytes.Replace(b, []byte(`"epoch":1`), []byte(`"epoch":1.0`), 1), "overflow": bytes.Replace(b, []byte(`"epoch":1`), []byte(`"epoch":18446744073709551616`), 1), "bad time": bytes.Replace(b, []byte(testTime), []byte("2026-09-14T01:02:03+00:00"), 1), "surrogate": bytes.Replace(b, []byte(`"before"`), []byte(`"\ud800"`), 1), "utf8": append([]byte{255}, b...), "trailing": append(append([]byte{}, b...), []byte(` {}`)...), "partial": b[:len(b)/2], "version": bytes.Replace(b, []byte(`"format":1`), []byte(`"format":2`), 1), "uppercase hash": bytes.Replace(b, []byte(repeated("a", 32)), []byte(repeated("A", 32)), 1)}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := decodeEnvelope(body); err == nil {
				t.Fatal("admitted malformed bytes")
			}
		})
	}
	exact := append(bytes.Clone(b), bytes.Repeat([]byte(" "), MaxWitnessBytes-len(b))...)
	if _, err := decodeEnvelope(exact); err != nil {
		t.Fatal("exact boundary", err)
	}
	if _, err := decodeEnvelope(append(exact, ' ')); err == nil {
		t.Fatal("one over")
	}
	fields := testWitness()
	fields.StateCommit = repeated("x", MaxWitnessTokenBytes)
	if _, err := NewWitness(fields); err != nil {
		t.Fatal(err)
	}
	fields.StateCommit += "x"
	if _, err := NewWitness(fields); err == nil {
		t.Fatal("token over")
	}
}
func TestWitnessUnsafeFiles(t *testing.T) {
	for _, kind := range []string{"symlink", "fifo", "directory", "permissions", "hardlink"} {
		t.Run(kind, func(t *testing.T) {
			m := testManager(t)
			s := saveActive(t, m, false, false)
			if err := os.Remove(m.path()); err != nil {
				t.Fatal(err)
			}
			other := filepath.Join(m.dir, "other")
			if err := os.WriteFile(other, s.Bytes(), 0600); err != nil {
				t.Fatal(err)
			}
			var err error
			switch kind {
			case "symlink":
				err = os.Symlink(other, m.path())
			case "fifo":
				err = unix.Mkfifo(m.path(), 0600)
			case "directory":
				err = os.Mkdir(m.path(), 0700)
			case "permissions":
				err = os.WriteFile(m.path(), s.Bytes(), 0644)
			case "hardlink":
				err = os.Link(other, m.path())
			}
			if err != nil {
				t.Fatal(err)
			}
			if _, err = m.Load(context.Background()); err == nil {
				t.Fatal("unsafe Load")
			}
			body, err := os.ReadFile(other)
			if err != nil || !bytes.Equal(body, s.Bytes()) {
				t.Fatal("unsafe input changed")
			}
		})
	}
}
func TestWitnessLockAndGuard(t *testing.T) {
	m := testManager(t)
	g := testGuard(t, m)
	info, err := g.file.Stat()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	if _, err = m.Acquire(ctx); !errors.Is(err, context.DeadlineExceeded) || errors.Is(err, ErrWitnessBusy) {
		t.Fatal("caller cap", err)
	}
	if err = g.Close(); err != nil {
		t.Fatal(err)
	}
	g2 := testGuard(t, m)
	now, err := g2.file.Stat()
	if err != nil || !os.SameFile(info, now) {
		t.Fatal("lock inode replaced")
	}
	if err = g.MarkUnverified(context.Background()); !errors.Is(err, ErrGuardMisuse) {
		t.Fatal(err)
	}
	for _, sentinel := range []error{lockfile.ErrLocked, lockfile.ErrLockBusy} {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if err := waitWitnessLock(ctx, g2.file, func(*os.File) error { return sentinel }); !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	}
}
