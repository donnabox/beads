//go:build unix

package authority

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/steveyegge/beads/internal/lockfile"
	"golang.org/x/sys/unix"
)

func newWitnessManager(ctx context.Context, dir string) (*Manager, error) {
	key, err := InstallationKey(ctx, dir)
	if err != nil {
		return nil, err
	}
	canonical, info, err := installationDirectory(dir)
	if err != nil {
		return nil, witnessFSError(err)
	}
	return &Manager{canonical, key, info, nativeWitnessIO()}, nil
}
func privateWitnessFile(info os.FileInfo) bool {
	if info == nil {
		return false
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	return ok && stat.Nlink == 1 && privateInstallationFile(info)
}
func openWitnessFile(path string, create bool) (*os.File, error) {
	before, err := os.Lstat(path)
	flags := unix.O_RDONLY | unix.O_NOFOLLOW | unix.O_NONBLOCK | unix.O_CLOEXEC
	if create {
		flags = unix.O_RDWR | unix.O_NOFOLLOW | unix.O_NONBLOCK | unix.O_CLOEXEC
	}
	if errors.Is(err, os.ErrNotExist) && create {
		flags |= unix.O_CREAT | unix.O_EXCL
	} else if err != nil {
		return nil, err
	} else if !privateWitnessFile(before) {
		return nil, ErrInvalidRecord
	}
	fd, err := unix.Open(path, flags, 0600)
	if create && errors.Is(err, unix.EEXIST) {
		return openWitnessFile(path, true)
	}
	if err != nil {
		return nil, err
	}
	f := os.NewFile(uintptr(fd), path)
	if err = checkWitnessFile(path, f); err != nil {
		return nil, errors.Join(err, f.Close())
	}
	opened, err := f.Stat()
	if err != nil || before != nil && !os.SameFile(before, opened) {
		return nil, errors.Join(ErrInvalidRecord, err, f.Close())
	}
	return f, nil
}
func checkWitnessFile(path string, f *os.File) error {
	opened, err := f.Stat()
	if err != nil {
		return err
	}
	named, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !privateWitnessFile(opened) || !privateWitnessFile(named) || !os.SameFile(opened, named) {
		return ErrInvalidRecord
	}
	return nil
}
func readWitnessFile(path string) (s Snapshot, err error) {
	f, err := openWitnessFile(path, false)
	if errors.Is(err, os.ErrNotExist) {
		return Snapshot{}, nil
	}
	if err != nil {
		return Snapshot{}, err
	}
	defer func() {
		err = errors.Join(err, f.Close())
		if err != nil {
			s = Snapshot{}
		}
	}()
	b, err := io.ReadAll(io.LimitReader(f, MaxWitnessBytes+1))
	if err != nil {
		return Snapshot{}, err
	}
	if err = checkWitnessFile(path, f); err != nil {
		return Snapshot{}, err
	}
	return decodeEnvelope(b)
}
func waitWitnessLock(ctx context.Context, f *os.File, lock func(*os.File) error) error {
	wait, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	for {
		if wait.Err() != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return errors.Join(ErrWitnessBusy, wait.Err())
		}
		err := lock(f)
		if err == nil {
			return nil
		}
		if !errors.Is(err, lockfile.ErrLocked) && !errors.Is(err, lockfile.ErrLockBusy) {
			return err
		}
		timer := time.NewTimer(25 * time.Millisecond)
		select {
		case <-wait.Done():
			timer.Stop()
		case <-timer.C:
		}
	}
}

// Reuse the identity implementation's full same-device ancestry and safe opens.
// Translate its private diagnostic labels without changing its behavior or bytes.
func syncWitnessAncestors(ctx context.Context, dir string) error {
	return witnessFSError(syncInstallationAncestors(ctx, dir, nativeInstallationIO()))
}
func witnessFSError(err error) error {
	if err == nil {
		return nil
	}
	if err == ErrInvalidIdentity {
		return ErrInvalidRecord
	}
	if err == ErrInstallationUnsupported {
		return ErrWitnessUnsupported
	}
	if p, ok := err.(*os.PathError); ok {
		return &os.PathError{Op: strings.ReplaceAll(p.Op, "installation", "witness"), Path: p.Path, Err: witnessFSError(p.Err)}
	}
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		var translated []error
		for _, child := range joined.Unwrap() {
			translated = append(translated, witnessFSError(child))
		}
		return errors.Join(translated...)
	}
	return err
}
