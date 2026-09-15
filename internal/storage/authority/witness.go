package authority

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/steveyegge/beads/internal/atomicfile"
	"github.com/steveyegge/beads/internal/beadsignore"
	"github.com/steveyegge/beads/internal/lockfile"
)

// Manager is a private, SQL-free durable witness store. Its snapshots and provider
// facts do not qualify an engine, establish authority, or authorize dispatch.
// Callers retain their workspace gate, then Guard, then qualified engine session.
type Manager struct {
	dir, key  string
	directory os.FileInfo
	io        witnessIO
}
type witnessIO struct {
	replace             func(string, []byte, os.FileMode) error
	remove              func(string) error
	syncFile            func(*os.File) error
	syncDirectory       func(context.Context, string) error
	lock, unlock, close func(*os.File) error
}

func nativeWitnessIO() witnessIO {
	return witnessIO{atomicfile.WriteFile, os.Remove, (*os.File).Sync, syncWitnessAncestors, lockfile.FlockExclusiveNonBlocking, lockfile.FlockUnlock, (*os.File).Close}
}

// New captures the actual InstallationKey before any witness guard is acquired.
func New(ctx context.Context, beadsDir string) (*Manager, error) {
	if ctx == nil {
		return nil, ErrInvalidRecord
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return newWitnessManager(ctx, beadsDir)
}
func (m *Manager) path() string { return filepath.Join(m.dir, beadsignore.WitnessName) }
func (m *Manager) checkDirectory() error {
	if m == nil {
		return ErrGuardMisuse
	}
	now, err := os.Lstat(m.dir)
	if err != nil {
		return err
	}
	if !now.IsDir() || !os.SameFile(m.directory, now) {
		return ErrInvalidRecord
	}
	return nil
}

// Load is passive: it does not create a lock, flush, recover, or invoke evidence.
func (m *Manager) Load(ctx context.Context) (Snapshot, error) {
	s, err := m.read(ctx)
	if err != nil {
		return Snapshot{}, err
	}
	if !s.Absent() && (s.InstallationKey() != m.key || s.Pending() && s.OperationInstallationKey() != m.key) {
		return Snapshot{}, ErrWrongInstallation
	}
	return s, nil
}
func (m *Manager) read(ctx context.Context) (Snapshot, error) {
	if ctx == nil {
		return Snapshot{}, ErrInvalidRecord
	}
	if err := ctx.Err(); err != nil {
		return Snapshot{}, err
	}
	if err := m.checkDirectory(); err != nil {
		return Snapshot{}, err
	}
	if err := beadsignore.Check(ctx, m.dir, beadsignore.WitnessName, false); err != nil {
		return Snapshot{}, err
	}
	s, err := readWitnessFile(m.path())
	if err != nil {
		return Snapshot{}, err
	}
	if err := m.checkDirectory(); err != nil {
		return Snapshot{}, err
	}
	if err := ctx.Err(); err != nil {
		return Snapshot{}, err
	}
	return s, nil
}

// Guard owns the permanent lock inode until Close. It is nonreentrant.
type Guard struct {
	manager *Manager
	file    *os.File
	mu      sync.Mutex
	closed  bool
}

func (m *Manager) Acquire(ctx context.Context) (*Guard, error) {
	if ctx == nil {
		return nil, ErrInvalidRecord
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := m.checkDirectory(); err != nil {
		return nil, err
	}
	f, err := openWitnessFile(filepath.Join(m.dir, beadsignore.LockName), true)
	if err != nil {
		return nil, err
	}
	if err = waitWitnessLock(ctx, f, m.io.lock); err != nil {
		return nil, errors.Join(err, m.io.close(f))
	}
	g := &Guard{manager: m, file: f}
	if err = g.check(ctx); err != nil {
		return nil, errors.Join(err, m.io.unlock(f), m.io.close(f))
	}
	return g, nil
}
func (g *Guard) enter(ctx context.Context) error {
	if g == nil || !g.mu.TryLock() {
		return ErrGuardMisuse
	}
	if err := g.check(ctx); err != nil {
		g.mu.Unlock()
		return err
	}
	return nil
}
func (g *Guard) check(ctx context.Context) error {
	if g == nil || g.closed || g.manager == nil || g.file == nil || ctx == nil {
		return ErrGuardMisuse
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := g.manager.checkDirectory(); err != nil {
		return err
	}
	return checkWitnessFile(g.file.Name(), g.file)
}
func (g *Guard) Close() error {
	if g == nil || !g.mu.TryLock() {
		return ErrGuardMisuse
	}
	defer g.mu.Unlock()
	if g.closed || g.manager == nil || g.file == nil {
		return ErrGuardMisuse
	}
	g.closed = true
	return errors.Join(g.manager.io.unlock(g.file), g.manager.io.close(g.file))
}
func (g *Guard) read(ctx context.Context) (Snapshot, error) {
	if err := g.check(ctx); err != nil {
		return Snapshot{}, err
	}
	s, err := g.manager.read(ctx)
	if err != nil {
		return Snapshot{}, err
	}
	if s.Pending() && s.OperationInstallationKey() != g.manager.key {
		return Snapshot{}, ErrWrongInstallation
	}
	return s, nil
}
func (g *Guard) fresh(ctx context.Context, previous Snapshot) error {
	s, err := g.read(ctx)
	if err != nil {
		return err
	}
	if s.Token() != previous.Token() {
		return ErrGuardMisuse
	}
	return nil
}
func (g *Guard) flush(ctx context.Context, s Snapshot) (err error) {
	if err = g.fresh(ctx, s); err != nil {
		return err
	}
	if !s.Absent() {
		f, e := openWitnessFile(g.manager.path(), false)
		if e != nil {
			return e
		}
		err = errors.Join(g.manager.io.syncFile(f), checkWitnessFile(g.manager.path(), f), g.manager.io.close(f))
		if err != nil {
			return err
		}
	}
	if err = g.manager.io.syncDirectory(ctx, g.manager.dir); err != nil {
		return err
	}
	return g.fresh(ctx, s)
}
func (g *Guard) persist(ctx context.Context, previous Snapshot, e *envelope) (Snapshot, error) {
	return g.persistOwned(ctx, previous, e, nil)
}
func (g *Guard) persistOwned(ctx context.Context, previous Snapshot, e *envelope, beforeWrite func() error) (Snapshot, error) {
	var next Snapshot
	var err error
	if e != nil {
		next, err = encodeEnvelope(*e)
		if err != nil {
			return Snapshot{}, err
		}
	}
	if err = g.fresh(ctx, previous); err != nil {
		return Snapshot{}, err
	}
	if err = beadsignore.Prepare(ctx, g.manager.dir, func() error { return g.manager.io.syncDirectory(ctx, g.manager.dir) }); err != nil {
		return Snapshot{}, &PersistenceError{Unchanged, err, true}
	}
	if err = g.fresh(ctx, previous); err != nil {
		return Snapshot{}, &PersistenceError{Unchanged, err, true}
	}
	if beforeWrite != nil {
		if err = beforeWrite(); err != nil {
			return Snapshot{}, &PersistenceError{Unchanged, err, true}
		}
	}
	if e == nil {
		if !previous.Absent() {
			err = g.manager.io.remove(g.manager.path())
		}
	} else {
		err = g.manager.io.replace(g.manager.path(), next.Bytes(), 0600)
	}
	if err != nil {
		effect := Unchanged
		observed, readErr := g.manager.read(ctx)
		if readErr != nil || observed.Token() != previous.Token() {
			effect = VisibleUnknown
		}
		return Snapshot{}, &PersistenceError{effect, errors.Join(err, readErr), true}
	}
	if err = g.flush(ctx, next); err != nil {
		return Snapshot{}, &PersistenceError{VisibleUnknown, err, true}
	}
	return next, nil
}
func nextGeneration(s Snapshot) (uint64, error) {
	if s.Generation() == ^uint64(0) {
		return 0, ErrInvalidRecord
	}
	return s.Generation() + 1, nil
}
func activeForKey(s Snapshot, key string) (Witness, error) {
	if s.Absent() {
		return Witness{}, ErrEvidence
	}
	if s.InstallationKey() != key {
		return Witness{}, ErrWrongInstallation
	}
	if s.Pending() {
		return Witness{}, ErrWitnessPending
	}
	w, ok := s.Witness()
	if !ok {
		return Witness{}, ErrInvalidRecord
	}
	return w, nil
}

// MarkUnverified is durable before a caller may invoke its separately qualified restore.
func (g *Guard) MarkUnverified(ctx context.Context) error {
	if err := g.enter(ctx); err != nil {
		return err
	}
	defer g.mu.Unlock()
	s, err := g.read(ctx)
	if err != nil {
		return err
	}
	return g.mark(ctx, s)
}
func (g *Guard) mark(ctx context.Context, s Snapshot) error {
	e := s.record()
	if e.Witness == nil || e.Witness.Unverified {
		return g.flush(ctx, s)
	}
	generation, err := nextGeneration(s)
	if err != nil {
		return err
	}
	e.Generation = generation
	e.Witness.Unverified = true
	_, err = g.persist(ctx, s, &e)
	return err
}
