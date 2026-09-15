//go:build unix

package authority

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/steveyegge/beads/internal/atomicfile"
	"github.com/steveyegge/beads/internal/beadsignore"
)

func TestWitnessPersistenceFailureEffects(t *testing.T) {
	for _, point := range []string{"before replacement", "visible replacement", "file sync", "directory sync"} {
		t.Run(point, func(t *testing.T) {
			m := testManager(t)
			before := saveActive(t, m, false, false)
			g := testGuard(t, m)
			cause := errors.New("injected persistence failure")
			realIO := m.io
			switch point {
			case "before replacement":
				m.io.replace = func(string, []byte, os.FileMode) error { return cause }
			case "visible replacement":
				m.io.replace = func(p string, b []byte, mode os.FileMode) error {
					if err := atomicfile.WriteFile(p, b, mode); err != nil {
						return err
					}
					return cause
				}
			case "file sync":
				m.io.syncFile = func(*os.File) error { return cause }
			case "directory sync":
				m.io.syncDirectory = func(ctx context.Context, dir string) error {
					s, err := readWitnessFile(m.path())
					if err != nil {
						return err
					}
					if s.Token() != before.Token() {
						return cause
					}
					return realIO.syncDirectory(ctx, dir)
				}
			}
			restoreCalls := 0
			err := g.MarkUnverified(context.Background())
			if err == nil {
				restoreCalls++
			}
			var persistence *PersistenceError
			if !errors.As(err, &persistence) || !errors.Is(err, cause) || restoreCalls != 0 {
				t.Fatal("failure swallowed", err)
			}
			expected := VisibleUnknown
			if point == "before replacement" {
				expected = Unchanged
			}
			if persistence.Effect != expected {
				t.Fatalf("effect %s", persistence.Effect)
			}
			visible, err := g.read(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if (visible.Token() == before.Token()) != (point == "before replacement") {
				t.Fatal("wrong visible state")
			}
			m.io = realIO
			fileSyncs, dirSyncs := 0, 0
			m.io.syncFile = func(f *os.File) error { fileSyncs++; return realIO.syncFile(f) }
			m.io.syncDirectory = func(ctx context.Context, dir string) error { dirSyncs++; return realIO.syncDirectory(ctx, dir) }
			if err = g.MarkUnverified(context.Background()); err != nil {
				t.Fatal(err)
			}
			if fileSyncs == 0 || dirSyncs == 0 {
				t.Fatal("retry did not flush retained state")
			}
			if err = g.MarkUnverified(context.Background()); err != nil {
				t.Fatal(err)
			}
			s, err := g.read(context.Background())
			w, ok := s.Witness()
			if err != nil || !ok || !w.fields.Unverified {
				t.Fatal("marker lost")
			}
		})
	}
}
func TestWitnessRetainedNoopAndCleanup(t *testing.T) {
	m := testManager(t)
	saveActive(t, m, false, true)
	g := testGuard(t, m)
	cause := errors.New("sync failure")
	original := m.io.syncDirectory
	m.io.syncDirectory = func(context.Context, string) error { return cause }
	if err := g.MarkUnverified(context.Background()); !errors.Is(err, cause) {
		t.Fatal("repeated mark skipped sync", err)
	}
	m.io.syncDirectory = original
	unlockErr, closeErr := errors.New("unlock"), errors.New("close")
	unlock, closeFile := m.io.unlock, m.io.close
	m.io.unlock = func(f *os.File) error { return errors.Join(unlock(f), unlockErr) }
	m.io.close = func(f *os.File) error { return errors.Join(closeFile(f), closeErr) }
	err := g.Close()
	if !errors.Is(err, unlockErr) || !errors.Is(err, closeErr) {
		t.Fatal("cleanup errors", err)
	}
	m.io.unlock = unlock
	m.io.close = closeFile
	g2 := testGuard(t, m)
	if err = g2.MarkUnverified(context.Background()); err != nil {
		t.Fatal("actual descriptors not released", err)
	}
}
func TestWitnessRemovalFailureAndResidue(t *testing.T) {
	m := testManager(t)
	saveActive(t, m, false, false)
	orphan := filepath.Join(m.dir, ".~graph-authority.local.json.orphan")
	if err := os.WriteFile(orphan, []byte("not a witness"), 0600); err != nil {
		t.Fatal(err)
	}
	g := testGuard(t, m)
	e := recording()
	tr, err := g.Begin(context.Background(), intentFor(Adopt, "", ""), e)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range []Phase{LocalCommitted, Published, ConfigWritten} {
		if err = tr.SetPhase(context.Background(), p, e, nil); err != nil {
			t.Fatal(err)
		}
	}
	cause := errors.New("unlink failure")
	remove := m.io.remove
	m.io.remove = func(string) error { return cause }
	if err = tr.Finalize(context.Background(), e, nil); !errors.Is(err, cause) {
		t.Fatal(err)
	}
	s, err := g.read(context.Background())
	if err != nil || !s.Pending() {
		t.Fatal("failed removal lost pending")
	}
	m.io.remove = remove
	if err = g.Recover(context.Background(), e, nil); err != nil {
		t.Fatal(err)
	}
	s, err = g.read(context.Background())
	if err != nil || !s.Absent() {
		t.Fatal("removal recovery")
	}
	b, err := os.ReadFile(orphan)
	if err != nil || string(b) != "not a witness" {
		t.Fatal("orphan used or removed")
	}
	if !beadsignore.Sensitive(orphan) {
		t.Fatal("orphan not sensitive")
	}
}
func TestWitnessReplacedFilesAndDirectory(t *testing.T) {
	for _, which := range []string{"witness", "lock", "directory"} {
		t.Run(which, func(t *testing.T) {
			m := testManager(t)
			saveActive(t, m, false, false)
			g := testGuard(t, m)
			e := recording()
			e.prepare = func(r PrepareRequest) (Preparation, error) {
				p := preparationFor(r)
				switch which {
				case "witness":
					s := r.Prior()
					rec := s.record()
					rec.Generation++
					newer, err := encodeEnvelope(rec)
					if err != nil {
						return Preparation{}, err
					}
					if err = atomicfile.WriteFile(m.path(), newer.Bytes(), 0600); err != nil {
						return Preparation{}, err
					}
				case "lock":
					if err := os.Rename(g.file.Name(), g.file.Name()+".old"); err != nil {
						return Preparation{}, err
					}
					if err := os.WriteFile(g.file.Name(), nil, 0600); err != nil {
						return Preparation{}, err
					}
				case "directory":
					if err := os.Rename(m.dir, m.dir+"-moved"); err != nil {
						return Preparation{}, err
					}
					if err := os.Mkdir(m.dir, 0700); err != nil {
						return Preparation{}, err
					}
				}
				return NewPreparation(r, p)
			}
			if tr, err := g.Begin(context.Background(), intentFor(Mutation, "", ""), e); tr != nil || err == nil {
				t.Fatal("replacement admitted")
			}
		})
	}
}
func TestWitnessActualMoveAndForeignPending(t *testing.T) {
	m := testManager(t)
	before := saveActive(t, m, false, true)
	alias := m.dir + "-alias"
	if err := os.Symlink(m.dir, alias); err != nil {
		t.Fatal(err)
	}
	same, err := New(context.Background(), alias)
	if err != nil || same.key != m.key {
		t.Fatal("alias binding", err)
	}
	moved := m.dir + "-moved"
	if err = os.Rename(m.dir, moved); err != nil {
		t.Fatal(err)
	}
	current, err := New(context.Background(), moved)
	if err != nil {
		t.Fatal(err)
	}
	if current.key == m.key {
		t.Fatal("move did not rebind identity")
	}
	if _, err = current.Load(context.Background()); !errors.Is(err, ErrWrongInstallation) {
		t.Fatal("moved became authority", err)
	}
	g := testGuard(t, current)
	for _, kind := range []Kind{Mint, Install, Mutation, LedgerApply, Adopt} {
		e := recording()
		if tr, err := g.Begin(context.Background(), intentFor(kind, "", Ordinary), e); tr != nil || err == nil || e.prepareCount != 0 {
			t.Fatal("foreign ordinary admission", kind, err)
		}
	}
	e := recording()
	tr, err := g.Begin(context.Background(), intentFor(Promote, Steal, ""), e)
	if err != nil {
		t.Fatal(err)
	}
	pending, err := g.read(context.Background())
	if err != nil || pending.InstallationKey() != before.InstallationKey() || pending.OperationInstallationKey() != current.key {
		t.Fatal("pending provenance", err)
	}
	if err = g.Close(); err != nil {
		t.Fatal(err)
	}
	t.Setenv("BEADS_INSTALLATION_ID_FILE", filepath.Join(filepath.Dir(moved), "different-user", "installation-id"))
	other, err := New(context.Background(), moved)
	if err != nil {
		t.Fatal(err)
	}
	foreign := testGuard(t, other)
	for _, call := range []func() error{func() error { return foreign.MarkUnverified(context.Background()) }, func() error { return foreign.Recover(context.Background(), recording(), nil) }} {
		if err = call(); !errors.Is(err, ErrWrongInstallation) {
			t.Fatal("foreign operation resumed", err)
		}
	}
	raw, err := readWitnessFile(other.path())
	if err != nil || raw.Token() != pending.Token() {
		t.Fatal("foreign pending changed")
	}
	if err = tr.Finalize(context.Background(), e, nil); !errors.Is(err, ErrGuardMisuse) {
		t.Fatal("old guard reused")
	}
}
func TestWitnessConfigurationFailureRetry(t *testing.T) {
	m := testManager(t)
	saveActive(t, m, false, false)
	g := testGuard(t, m)
	e := recording()
	c := configFixture(t, m)
	_, err := g.Begin(context.Background(), intentFor(Rotate, "", ""), e)
	if err != nil {
		t.Fatal(err)
	}
	if err = g.Recover(context.Background(), e, nil); !errors.Is(err, ErrEvidence) {
		t.Fatal("missing adapter", err)
	}
	before, err := g.read(context.Background())
	if err != nil || before.Phase() != Published {
		t.Fatal("advanced without config")
	}
	cause := errors.New("response lost after fixture config sync")
	c.afterWrite = func() error { return cause }
	if err = g.Recover(context.Background(), e, c); !errors.Is(err, cause) {
		t.Fatal(err)
	}
	s, err := g.read(context.Background())
	if err != nil || s.Phase() != Published || s.OperationID() != before.OperationID() {
		t.Fatal("config failure changed phase")
	}
	c.afterWrite = nil
	if err = g.Recover(context.Background(), e, c); err != nil {
		t.Fatal(err)
	}
}
