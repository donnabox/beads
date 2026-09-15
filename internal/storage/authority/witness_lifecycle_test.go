//go:build unix

package authority

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/steveyegge/beads/internal/atomicfile"
)

// fileConfiguration is a real disposable file adapter for the test's closed
// fixture format. It does not implement production YAML or protected-key policy.
type fileConfiguration struct {
	calls      int
	afterWrite func() error
}

func (c *fileConfiguration) ApplyScopeURLIntent(ctx context.Context, r ConfigurationRequest) (ConfigurationReceipt, error) {
	c.calls++
	path := filepath.Join(r.Directory(), "fixture-config.json")
	values := map[string]string{}
	b, err := os.ReadFile(path)
	if err == nil {
		if err = json.Unmarshal(b, &values); err != nil {
			return ConfigurationReceipt{}, err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return ConfigurationReceipt{}, err
	}
	old, exists := values["bdp.scope_url"]
	i := r.Intent()
	if old != i.Desired && (i.Previous.Presence == Absent && exists || i.Previous.Presence == Present && (!exists || old != i.Previous.Value)) {
		return ConfigurationReceipt{}, ErrEvidence
	}
	values["bdp.scope_url"] = i.Desired
	b, err = json.Marshal(values)
	if err != nil {
		return ConfigurationReceipt{}, err
	}
	if err = atomicfile.WriteFile(path, b, 0600); err != nil {
		return ConfigurationReceipt{}, err
	}
	if err = syncWitnessAncestors(ctx, r.Directory()); err != nil {
		return ConfigurationReceipt{}, err
	}
	if c.afterWrite != nil {
		if err = c.afterWrite(); err != nil {
			return ConfigurationReceipt{}, err
		}
	}
	b, err = os.ReadFile(path)
	if err != nil {
		return ConfigurationReceipt{}, err
	}
	if err = json.Unmarshal(b, &values); err != nil {
		return ConfigurationReceipt{}, err
	}
	return NewConfigurationReceipt(r, values["bdp.scope_url"])
}
func configFixture(t *testing.T, m *Manager) *fileConfiguration {
	t.Helper()
	if err := os.WriteFile(filepath.Join(m.dir, "fixture-config.json"), []byte(`{"bdp.scope_url":"https://beads.example/acme/","unrelated":"retained"}`), 0600); err != nil {
		t.Fatal(err)
	}
	return &fileConfiguration{}
}
func TestWitnessKindsAndRecovery(t *testing.T) {
	cases := []struct {
		name                             string
		kind                             Kind
		mode                             PromoteMode
		lane                             Lane
		absent, foreign, marked, restore bool
	}{
		{name: "mint", kind: Mint, absent: true}, {name: "self regrant", kind: Promote, mode: SelfRegrant}, {name: "self regrant without local witness", kind: Promote, mode: SelfRegrant, absent: true}, {name: "steal", kind: Promote, mode: Steal}, {name: "foreign steal", kind: Promote, mode: Steal, foreign: true, marked: true}, {name: "rotate", kind: Rotate}, {name: "foreign rotate", kind: Rotate, foreign: true, marked: true}, {name: "restore rotate", kind: Rotate, marked: true, restore: true}, {name: "install", kind: Install}, {name: "mutation", kind: Mutation}, {name: "ledger", kind: LedgerApply, lane: Ordinary}, {name: "restore ledger", kind: LedgerApply, lane: RestoreRecovery, marked: true}, {name: "absent restore ledger", kind: LedgerApply, lane: RestoreRecovery, absent: true}, {name: "adopt", kind: Adopt}}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := testManager(t)
			var old Snapshot
			if !tc.absent {
				old = saveActive(t, m, tc.foreign, tc.marked)
			}
			g := testGuard(t, m)
			e := recording()
			intent := intentFor(tc.kind, tc.mode, tc.lane)
			if tc.restore {
				fields := intent.Fields()
				fields.RotateCause = RestoreWithoutContinuity
				var err error
				intent, err = NewIntent(fields)
				if err != nil {
					t.Fatal(err)
				}
			}
			c := configFixture(t, m)
			tr, err := g.Begin(context.Background(), intent, e)
			if err != nil {
				t.Fatal(err)
			}
			s, err := g.read(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			id := s.OperationID()
			if tc.foreign {
				if s.InstallationKey() != old.InstallationKey() || s.OperationInstallationKey() != m.key {
					t.Fatal("rebound saved provenance")
				}
				if _, err = m.Load(context.Background()); !errors.Is(err, ErrWrongInstallation) {
					t.Fatal("foreign became active", err)
				}
			}
			prepared := s.Preparation()
			if prepared.ResultWitness != nil {
				prepared.ResultWitness.AuthorityID = "tampered"
			}
			if _, err = decodeEnvelope(s.Bytes()); err != nil {
				t.Fatal("mutable preparation", err)
			}
			if err = g.Close(); err != nil {
				t.Fatal(err)
			}
			if err = tr.Abandon(context.Background(), e); !errors.Is(err, ErrGuardMisuse) {
				t.Fatal("closed handle", err)
			}
			g = testGuard(t, m)
			if err = g.Recover(context.Background(), e, c); err != nil {
				t.Fatal(err)
			}
			s, err = m.Load(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if tc.kind == Adopt {
				if !s.Absent() {
					t.Fatal("adopt created witness")
				}
			} else {
				w, ok := s.Witness()
				if !ok || s.Pending() || w.fields.Unverified != tc.marked || s.InstallationKey() != m.key {
					t.Fatal("final state")
				}
				if tc.mode == SelfRegrant {
					if w.fields.StateCommit != "before" || w.fields.Epoch != 1 || w.fields.LedgerSeq != 2 {
						t.Fatal("invented event")
					}
				}
			}
			if id == "" || e.prepareCount != 1 {
				t.Fatal("operation replaced")
			}
			if tc.kind == Rotate {
				b, err := os.ReadFile(filepath.Join(m.dir, "fixture-config.json"))
				if err != nil {
					t.Fatal(err)
				}
				var config map[string]string
				if err = json.Unmarshal(b, &config); err != nil || config["unrelated"] != "retained" || config["bdp.scope_url"] != intent.Fields().ScopeURL {
					t.Fatal("config corruption")
				}
			}
		})
	}
}
func TestWitnessAbandonAndContradictions(t *testing.T) {
	for _, kind := range []Kind{Mint, Promote, LedgerApply, Adopt, Rotate} {
		t.Run(string(kind), func(t *testing.T) {
			m := testManager(t)
			if kind != Mint {
				saveActive(t, m, kind == Rotate, true)
			}
			g := testGuard(t, m)
			e := recording()
			intent := intentFor(kind, SelfRegrant, RestoreRecovery)
			if kind == Adopt {
				saveActive(t, m, false, false)
			}
			tr, err := g.Begin(context.Background(), intent, e)
			if err != nil {
				t.Fatal(err)
			}
			e.inspect = func(r InspectRequest) (Inspection, error) { return NewInspection(r, preFor(r.Snapshot())) }
			if err = g.MarkUnverified(context.Background()); err != nil {
				t.Fatal(err)
			}
			if kind == Adopt {
				if err = tr.Abandon(context.Background(), e); !errors.Is(err, ErrGuardMisuse) {
					t.Fatal("stale handle", err)
				}
			}
			if err = g.Recover(context.Background(), e, nil); err != nil {
				t.Fatal(err)
			}
			s, err := g.read(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if kind == Mint {
				if !s.Absent() {
					t.Fatal("pre-mint retained")
				}
			} else {
				w, ok := s.Witness()
				if !ok || !w.fields.Unverified || s.Pending() {
					t.Fatal("lost sticky flag")
				}
			}
		})
	}
	for _, phase := range []Phase{LocalCommitted, Published, ConfigWritten} {
		t.Run("late "+string(phase), func(t *testing.T) {
			m := testManager(t)
			saveActive(t, m, false, false)
			g := testGuard(t, m)
			e := recording()
			tr, err := g.Begin(context.Background(), intentFor(Mutation, "", ""), e)
			if err != nil {
				t.Fatal(err)
			}
			for _, step := range []Phase{LocalCommitted, Published, ConfigWritten} {
				if err = tr.SetPhase(context.Background(), step, e, nil); err != nil {
					t.Fatal(err)
				}
				if step == phase {
					break
				}
			}
			before, err := g.read(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			e.inspect = func(r InspectRequest) (Inspection, error) { return NewInspection(r, preFor(r.Snapshot())) }
			if err = g.Recover(context.Background(), e, nil); !errors.Is(err, ErrEvidence) {
				t.Fatal(err)
			}
			after, err := g.read(context.Background())
			if err != nil || after.Token() != before.Token() {
				t.Fatal("late contradiction changed state")
			}
		})
	}
}
func TestWitnessPartialResultsStayPending(t *testing.T) {
	for _, name := range []string{"self lease effect", "adopt wrong lease", "adopt missing anti reuse", "adopt mixed tables", "event wrong op", "event wrong commit", "event wrong state", "ledger manifest", "unknown"} {
		t.Run(name, func(t *testing.T) {
			m := testManager(t)
			saveActive(t, m, false, false)
			g := testGuard(t, m)
			e := recording()
			kind := Mutation
			mode := PromoteMode("")
			if name == "self lease effect" {
				kind = Promote
				mode = SelfRegrant
			}
			if name[:5] == "adopt" {
				kind = Adopt
			}
			if name == "ledger manifest" {
				kind = LedgerApply
			}
			tr, err := g.Begin(context.Background(), intentFor(kind, mode, Ordinary), e)
			if err != nil {
				t.Fatal(err)
			}
			e.inspect = func(r InspectRequest) (Inspection, error) {
				f := completionFor(r.Snapshot())
				switch name {
				case "self lease effect":
					f = preFor(r.Snapshot())
					f.Lease = r.Snapshot().Preparation().ResultLease
				case "adopt wrong lease":
					f.Lease = testLease(m.key, testWitness(), "1")
				case "adopt missing anti reuse":
					f.AntiReuseDigest = ""
				case "adopt mixed tables":
					f.Tables[7].Content = "mixed"
				case "event wrong op":
					f.OperationID = repeated("0", 32)
				case "event wrong commit":
					f.OperationCommit = "other"
				case "event wrong state":
					f.StateVersion = repeated("0", 64)
				case "ledger manifest":
					f.ManifestDigest = repeated("0", 64)
				case "unknown":
					f.State = UnknownState
				}
				return NewInspection(r, f)
			}
			if err = tr.Abandon(context.Background(), e); !errors.Is(err, ErrEvidence) {
				t.Fatal("unsafe abandon", err)
			}
			if err = g.Recover(context.Background(), e, nil); !errors.Is(err, ErrEvidence) {
				t.Fatal(err)
			}
			s, err := g.read(context.Background())
			w, ok := s.Witness()
			if err != nil || !s.Pending() || !ok || !w.fields.Unverified {
				t.Fatal("partial result not retained")
			}
		})
	}
}
func TestWitnessAdvanceAndMarker(t *testing.T) {
	m := testManager(t)
	old := saveActive(t, m, false, false)
	g := testGuard(t, m)
	e := recording()
	base, _ := old.Witness()
	candidate := base
	candidate.fields.StateCommit = "later"
	candidate.fields.LedgerSeq++
	candidate.fields.LedgerHash = repeated("3", 64)
	candidate.fields.StateVersion = repeated("4", 64)
	if err := g.Advance(context.Background(), candidate, e); err != nil {
		t.Fatal(err)
	}
	newer, err := m.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	e.relation = CandidateContained
	if err = g.Advance(context.Background(), base, e); err != nil {
		t.Fatal(err)
	}
	after, err := m.Load(context.Background())
	if err != nil || after.Token() != newer.Token() {
		t.Fatal("older overwrote")
	}
	for _, relation := range []Relation{Fork, Unprovable} {
		e.relation = relation
		if err = g.Advance(context.Background(), candidate, e); !errors.Is(err, ErrEvidence) {
			t.Fatal(err)
		}
	}
	candidate.fields.Epoch++
	if err = g.Advance(context.Background(), candidate, e); !errors.Is(err, ErrEvidence) {
		t.Fatal("epoch smuggled")
	}
	if err = g.MarkUnverified(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err = g.Advance(context.Background(), base, e); !errors.Is(err, ErrWitnessUnverified) {
		t.Fatal(err)
	}
	if err = g.ClearUnverified(context.Background(), nil); !errors.Is(err, ErrEvidence) {
		t.Fatal("missing clear proof")
	}
	if err = g.ClearUnverified(context.Background(), e); err != nil {
		t.Fatal(err)
	}
	s, err := m.Load(context.Background())
	w, ok := s.Witness()
	if err != nil || !ok || w.fields.Unverified || w.fields.StateCommit != "later" {
		t.Fatal("clear altered state")
	}
}
func TestWitnessPreparationOwnership(t *testing.T) {
	for _, at := range []int{1, 2, 3, 4} {
		t.Run(string(rune('0'+at)), func(t *testing.T) {
			m := testManager(t)
			before := saveActive(t, m, false, false)
			g := testGuard(t, m)
			e := recording()
			e.observe = func(n int, o *Ownership) error {
				if n == at {
					return context.Canceled
				}
				return nil
			}
			tr, err := g.Begin(context.Background(), intentFor(Mutation, "", ""), e)
			if tr != nil || !errors.Is(err, context.Canceled) {
				t.Fatal("dispatch after loss", err)
			}
			s, readErr := g.read(context.Background())
			if readErr != nil {
				t.Fatal(readErr)
			}
			if at < 4 {
				if s.Token() != before.Token() {
					t.Fatal("pre-write change")
				}
			} else {
				var persistence *PersistenceError
				if !errors.As(err, &persistence) || persistence.Effect != RetainedPending || !s.Pending() {
					t.Fatal("missing retained classification", err)
				}
			}
		})
	}
	for _, name := range []string{"token", "intent", "branch", "key", "extra arm", "missing table", "missing lease", "unclassified operation", "recursive"} {
		t.Run(name, func(t *testing.T) {
			m := testManager(t)
			before := saveActive(t, m, false, false)
			g := testGuard(t, m)
			e := recording()
			e.prepare = func(r PrepareRequest) (Preparation, error) {
				p := preparationFor(r)
				switch name {
				case "token":
					r.binding.Token = repeated("0", 64)
				case "intent":
					r.binding.Digest = repeated("0", 64)
				case "branch":
					r.binding.Ownership.Branch = "other"
				case "key":
					r.currentKey = repeated("0", 64)
				case "extra arm":
					p.Install = &InstallPreparation{repeated("0", 64), Verified}
				case "missing table":
					p.ResultTables[3] = TableState{}
				case "missing lease":
					p.PreLease = LeaseState{}
				case "unclassified operation":
					p.Mutation.Operation = "shell"
				case "recursive":
					if err := g.MarkUnverified(context.Background()); !errors.Is(err, ErrGuardMisuse) {
						t.Fatal("recursive entry", err)
					}
					return Preparation{}, ErrEvidence
				}
				return NewPreparation(r, p)
			}
			if tr, err := g.Begin(context.Background(), intentFor(Mutation, "", ""), e); tr != nil || err == nil {
				t.Fatal("bad preparation admitted")
			}
			s, err := g.read(context.Background())
			if err != nil || s.Token() != before.Token() {
				t.Fatal("bad prep mutated")
			}
		})
	}
}

func TestWitnessRequiredPreparationVariants(t *testing.T) {
	cases := []struct {
		name   string
		kind   Kind
		mode   PromoteMode
		change func(*PreparationFields)
	}{
		{"mint existing scope", Mint, "", func(p *PreparationFields) {
			w := testWitness()
			p.PreScope = ScopeState{Present, w.ScopeURL, w.AuthorityID, w.Epoch}
		}},
		{"mint shared workspace", Mint, "", func(p *PreparationFields) {}},
		{"self changed version", Promote, SelfRegrant, func(p *PreparationFields) { p.ResultWitness.StateVersion = repeated("0", 64) }},
		{"self missing old lease", Promote, SelfRegrant, func(p *PreparationFields) { p.PreLease = LeaseState{Presence: Absent} }},
		{"steal missing same database", Promote, Steal, func(p *PreparationFields) { p.Promote.SameDatabase = "" }},
		{"steal stale fence", Promote, Steal, func(p *PreparationFields) { p.Promote.RenewedLease.Fence = p.PreLease.Fence }},
		{"rotate missing preservation", Rotate, "", func(p *PreparationFields) { p.Rotate.PreservationDigest = "" }},
		{"install missing validation", Install, "", func(p *PreparationFields) { p.Install.Validated = "" }},
		{"mutation no new event", Mutation, "", func(p *PreparationFields) { p.ResultWitness.LedgerSeq = p.PreLedger.Seq }},
		{"ledger wrong lineage", LedgerApply, "", func(p *PreparationFields) { p.LedgerApply.Lineage = repeated("0", 64) }},
		{"ledger wrong scope", LedgerApply, "", func(p *PreparationFields) { p.LedgerApply.ScopeURL = "https://beads.example/other/" }},
		{"ledger predecessor mismatch", LedgerApply, "", func(p *PreparationFields) { p.LedgerApply.Predecessor = repeated("0", 64) }},
		{"adopt missing source", Adopt, "", func(p *PreparationFields) { p.Adopt.SourceTables[2] = TableState{Presence: Absent} }},
		{"adopt dirty prestate", Adopt, "", func(p *PreparationFields) { p.Adopt.CleanWorkingState = "" }},
		{"adopt missing cleanup", Adopt, "", func(p *PreparationFields) { p.Adopt.OldLeaseCleanup = "" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := testManager(t)
			if tc.kind != Mint {
				saveActive(t, m, false, false)
			}
			g := testGuard(t, m)
			before, err := g.read(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			e := recording()
			if tc.name == "mint shared workspace" {
				e.owner.Topology = SharedDatabase
				e.owner.Workspace = SharedWorkspace
			}
			e.prepare = func(r PrepareRequest) (Preparation, error) {
				p := preparationFor(r)
				tc.change(&p)
				return NewPreparation(r, p)
			}
			if tr, err := g.Begin(context.Background(), intentFor(tc.kind, tc.mode, Ordinary), e); tr != nil || err == nil {
				t.Fatal("incomplete preparation admitted")
			}
			after, err := g.read(context.Background())
			if err != nil || after.Token() != before.Token() {
				t.Fatal("changed before valid preparation")
			}
		})
	}
}
func TestWitnessLedgerReplayScopeIsExplicit(t *testing.T) {
	m := testManager(t)
	saveActive(t, m, false, true)
	g := testGuard(t, m)
	e := recording()
	e.prepare = func(r PrepareRequest) (Preparation, error) {
		p := preparationFor(r)
		p.ResultWitness.Epoch++
		p.LedgerApply.ReplayedScope.Epoch++
		p.ResultLease.Epoch++
		p.LedgerApply.Regrant = p.ResultLease
		return NewPreparation(r, p)
	}
	if _, err := g.Begin(context.Background(), intentFor(LedgerApply, "", RestoreRecovery), e); err != nil {
		t.Fatal(err)
	}
	if err := g.Recover(context.Background(), e, nil); err != nil {
		t.Fatal(err)
	}
	s, err := m.Load(context.Background())
	w, ok := s.Witness()
	if err != nil || !ok || w.fields.Epoch != 2 || !w.fields.Unverified {
		t.Fatal("replay lost explicit scope or cleared flag")
	}
}
func TestWitnessPreparationBoundsAndGeneration(t *testing.T) {
	m := testManager(t)
	g := testGuard(t, m)
	longURL := "https://beads.example/" + repeated("x", MaxWitnessBytes-23) + "/"
	i, err := NewIntent(IntentFields{Kind: Mint, ScopeURL: longURL, ArtifactDigest: repeated("0", 64)})
	if err != nil {
		t.Fatal("bounded URL should fit intent", err)
	}
	if tr, err := g.Begin(context.Background(), i, recording()); tr != nil || err == nil {
		t.Fatal("oversized envelope admitted")
	}
	if s, err := g.read(context.Background()); err != nil || !s.Absent() {
		t.Fatal("oversize wrote envelope")
	}
	s := saveActive(t, m, false, false)
	rec := s.record()
	rec.Generation = ^uint64(0)
	s, err = encodeEnvelope(rec)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(m.path(), s.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	if err = g.MarkUnverified(context.Background()); !errors.Is(err, ErrInvalidRecord) {
		t.Fatal("generation wrapped", err)
	}
}
func TestWitnessZeroGuardsAndCanceledPreparation(t *testing.T) {
	if err := (&Guard{}).Close(); !errors.Is(err, ErrGuardMisuse) {
		t.Fatal(err)
	}
	if err := (&Transition{}).Finalize(context.Background(), nil, nil); !errors.Is(err, ErrGuardMisuse) {
		t.Fatal(err)
	}
	m := testManager(t)
	saveActive(t, m, false, false)
	g := testGuard(t, m)
	ctx, cancel := context.WithCancel(context.Background())
	e := recording()
	e.prepare = func(r PrepareRequest) (Preparation, error) {
		p, err := NewPreparation(r, preparationFor(r))
		cancel()
		return p, err
	}
	if tr, err := g.Begin(ctx, intentFor(Mutation, "", ""), e); tr != nil || !errors.Is(err, context.Canceled) {
		t.Fatal("canceled preparation dispatched", err)
	}
}
