//go:build unix

package authority

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/steveyegge/beads/internal/config"
	"github.com/steveyegge/beads/internal/lockfile"
	"github.com/steveyegge/beads/internal/utils"
	"golang.org/x/sys/unix"
)

// Per-call seams cover persistence failures without a simulated filesystem.
type installationIO struct {
	random    func([]byte) (int, error)
	syncFile  func(*os.File) error
	syncDir   func(string) error
	closeFile func(*os.File) error
	lock      func(*os.File) error
	unlock    func(*os.File) error
	device    func(string) (string, error)
}

func nativeInstallationIO() installationIO {
	return installationIO{rand.Read, (*os.File).Sync, syncInstallationDirectory,
		(*os.File).Close, lockfile.FlockExclusiveNonBlocking, lockfile.FlockUnlock, installationDevice}
}

func installationKey(ctx context.Context, beadsDir string) (string, error) {
	path := os.Getenv("BEADS_INSTALLATION_ID_FILE")
	if path == "" {
		configuration, err := config.UserConfigYamlPath()
		if err != nil {
			return "", fmt.Errorf("installation config path: %w", err)
		}
		path = filepath.Join(filepath.Dir(configuration), "installation-id")
	}
	return installationKeyAt(ctx, beadsDir, path, nativeInstallationIO())
}

func installationKeyAt(ctx context.Context, beadsDir, path string, ops installationIO) (key string, err error) {
	base := filepath.Base(path)
	if !filepath.IsAbs(path) || strings.HasSuffix(path, string(filepath.Separator)) || base == "." || base == ".." {
		return "", installationPathError("validate ID file path", path, ErrInvalidIdentity)
	}
	workspace, workspaceInfo, err := installationDirectory(beadsDir)
	if err != nil {
		return "", installationPathError("validate workspace", beadsDir, err)
	}
	if err = ctx.Err(); err != nil {
		return "", err
	}
	if err = makeInstallationParents(ctx, filepath.Dir(path), ops); err != nil {
		return "", err
	}
	parent, _, err := installationDirectory(filepath.Dir(path))
	if err != nil {
		return "", installationPathError("validate parent", filepath.Dir(path), err)
	}
	path = filepath.Join(parent, base)
	if filepath.Dir(path) != parent {
		return "", installationPathError("validate joined ID parent", path, ErrInvalidIdentity)
	}
	lockPath := path + ".lock"
	lock, err := openInstallationFile(lockPath, true)
	if err != nil {
		return "", err
	}
	locked := false
	defer func() {
		if locked {
			err = errors.Join(err, wrapInstallationError("unlock", lockPath, ops.unlock(lock)))
		}
		err = errors.Join(err, wrapInstallationError("close lock", lockPath, ops.closeFile(lock)))
		if err != nil {
			key = ""
		}
	}()
	if err = waitInstallationLock(ctx, lock, ops.lock); err != nil {
		return "", err
	}
	locked = true
	if err = checkInstallationFile(lockPath, lock); err != nil {
		return "", err
	}
	if err = createInstallationID(ctx, path, ops); err != nil {
		return "", err
	}
	file, err := openInstallationFile(path, false)
	if err != nil {
		return "", err
	}
	defer func() {
		err = errors.Join(err, wrapInstallationError("close ID", path, ops.closeFile(file)))
		if err != nil {
			key = ""
		}
	}()
	id, err := readInstallationID(file)
	if err != nil {
		return "", err
	}
	if err = ctx.Err(); err != nil {
		return "", err
	}
	if err = wrapInstallationError("sync ID", path, ops.syncFile(file)); err != nil {
		return "", err
	}
	if err = syncInstallationAncestors(ctx, parent, ops); err != nil {
		return "", err
	}
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		return "", installationPathError("seek ID for reread", path, err)
	}
	persisted, err := readInstallationID(file)
	if err != nil {
		return "", err
	}
	if !bytes.Equal(id, persisted) {
		return "", installationPathError("ID changed during use", path, ErrInvalidIdentity)
	}
	if err = checkInstallationFile(path, file); err != nil {
		return "", err
	}
	if err = checkInstallationFile(lockPath, lock); err != nil {
		return "", err
	}
	current, err := os.Lstat(workspace)
	if err != nil || !current.IsDir() || !os.SameFile(workspaceInfo, current) {
		return "", installationPathError("recheck workspace", workspace, errors.Join(ErrInvalidIdentity, err))
	}
	if err = ctx.Err(); err != nil {
		return "", err
	}
	digest := sha256.Sum256(append(append(persisted[:64:64], ':'), []byte(workspace)...))
	return hex.EncodeToString(digest[:]), nil
}

func installationDirectory(path string) (string, os.FileInfo, error) {
	before, err := os.Stat(path)
	if err != nil {
		return "", nil, err
	}
	if !before.IsDir() {
		return "", nil, ErrInvalidIdentity
	}
	resolved, err := utils.CanonicalizeExistingPath(path)
	if err != nil {
		return "", nil, err
	}
	after, err := os.Lstat(resolved)
	if err != nil {
		return "", nil, err
	}
	if !after.IsDir() || !os.SameFile(before, after) {
		return "", nil, ErrInvalidIdentity
	}
	return resolved, after, nil
}

func makeInstallationParents(ctx context.Context, path string, ops installationIO) (err error) {
	defer func() { err = installationPathError("prepare parent", path, err) }()
	if err := ctx.Err(); err != nil {
		return err
	}
	info, err := os.Stat(path)
	if err == nil {
		if !info.IsDir() {
			return ErrInvalidIdentity
		}
		return nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	parent := filepath.Dir(path)
	if parent == path {
		return err
	}
	if err := makeInstallationParents(ctx, parent, ops); err != nil {
		return err
	}
	if err := os.Mkdir(path, 0700); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}
	resolved, _, err := installationDirectory(parent)
	if err != nil {
		return err
	}
	return wrapInstallationError("sync new installation parent entry", resolved, ops.syncDir(resolved))
}

func privateInstallationFile(info os.FileInfo) bool {
	stat, ok := info.Sys().(*syscall.Stat_t)
	return ok && int64(stat.Uid) == int64(os.Geteuid()) && info.Mode().IsRegular() && info.Mode().Perm()&0077 == 0
}

func openInstallationFile(path string, create bool) (file *os.File, err error) {
	defer func() { err = installationPathError("open private file", path, err) }()
	before, err := os.Lstat(path)
	flags := unix.O_RDWR | unix.O_NOFOLLOW | unix.O_NONBLOCK | unix.O_CLOEXEC
	if errors.Is(err, os.ErrNotExist) && create {
		flags |= unix.O_CREAT | unix.O_EXCL
	} else if err != nil {
		return nil, err
	} else if !privateInstallationFile(before) {
		return nil, ErrInvalidIdentity
	}
	fd, err := unix.Open(path, flags, 0600)
	if errors.Is(err, unix.EEXIST) && create {
		return openInstallationFile(path, false)
	}
	if err != nil {
		return nil, err
	}
	file = os.NewFile(uintptr(fd), path)
	if err := checkInstallationFile(path, file); err != nil {
		return nil, errors.Join(err, file.Close())
	}
	if before != nil {
		after, err := file.Stat()
		if err != nil || !os.SameFile(before, after) {
			return nil, errors.Join(ErrInvalidIdentity, err, file.Close())
		}
	}
	return file, nil
}

func checkInstallationFile(path string, file *os.File) (err error) {
	defer func() { err = installationPathError("check private file identity", path, err) }()
	opened, err := file.Stat()
	if err != nil {
		return err
	}
	current, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !privateInstallationFile(opened) || !privateInstallationFile(current) || !os.SameFile(opened, current) {
		return ErrInvalidIdentity
	}
	return nil
}

func createInstallationID(ctx context.Context, path string, ops installationIO) (err error) {
	defer func() { err = installationPathError("create ID", path, err) }()
	if _, err := os.Lstat(path); err == nil {
		return nil // openInstallationFile validates the existing winner.
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	var random [32]byte
	if n, err := ops.random(random[:]); err != nil || n != len(random) {
		return errors.Join(err, io.ErrUnexpectedEOF)
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	fd, err := unix.Open(path, unix.O_WRONLY|unix.O_CREAT|unix.O_EXCL|unix.O_NOFOLLOW|unix.O_NONBLOCK|unix.O_CLOEXEC, 0600)
	if errors.Is(err, unix.EEXIST) {
		return nil // Reread the actual winner, never the generated candidate.
	}
	if err != nil {
		return err
	}
	file := os.NewFile(uintptr(fd), path)
	defer func() { err = errors.Join(err, wrapInstallationError("close created ID", path, ops.closeFile(file))) }()
	if err = checkInstallationFile(path, file); err != nil {
		return err
	}
	encoded := hex.EncodeToString(random[:]) + "\n"
	if n, writeErr := io.WriteString(file, encoded); writeErr != nil || n != len(encoded) {
		return errors.Join(writeErr, io.ErrShortWrite)
	}
	return wrapInstallationError("sync created ID", path, ops.syncFile(file))
}

func readInstallationID(file *os.File) (data []byte, err error) {
	defer func() { err = installationPathError("read ID", file.Name(), err) }()
	data, err = io.ReadAll(io.LimitReader(file, 66))
	if err != nil {
		return nil, err
	}
	if len(data) != 65 || data[64] != '\n' {
		return nil, ErrInvalidIdentity
	}
	for _, c := range data[:64] {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return nil, ErrInvalidIdentity
		}
	}
	return data, nil
}

func waitInstallationLock(ctx context.Context, file *os.File, lock func(*os.File) error) (err error) {
	defer func() { err = installationPathError("wait for lock", file.Name(), err) }()
	wait, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	waitError := func() error {
		if err := ctx.Err(); err != nil {
			return err
		}
		return errors.Join(ErrInstallationBusy, wait.Err())
	}
	for {
		if err := wait.Err(); err != nil {
			return waitError()
		}
		err := lock(file)
		if err == nil {
			return nil
		}
		if !errors.Is(err, lockfile.ErrLocked) && !errors.Is(err, lockfile.ErrLockBusy) {
			return wrapInstallationError("lock installation ID", file.Name(), err)
		}
		timer := time.NewTimer(25 * time.Millisecond)
		select {
		case <-wait.Done():
			timer.Stop()
			return waitError()
		case <-timer.C:
		}
	}
}

// Existing ancestors can be residue from a failed earlier call. Flush every
// same-device ancestor including its root; never stop merely because it exists.
func syncInstallationAncestors(ctx context.Context, parent string, ops installationIO) error {
	device, err := ops.device(parent)
	if err != nil {
		return installationPathError("stat parent device", parent, err)
	}
	for directory := parent; ; {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := wrapInstallationError("sync installation ancestor", directory, ops.syncDir(directory)); err != nil {
			return err
		}
		next := filepath.Dir(directory)
		if next == directory {
			return nil
		}
		nextDevice, err := ops.device(next)
		if err != nil {
			return installationPathError("stat ancestor device", next, err)
		}
		if nextDevice != device {
			return nil
		}
		directory = next
	}
}

func installationDevice(path string) (string, error) {
	var stat unix.Stat_t
	if err := unix.Lstat(path, &stat); err != nil {
		return "", err
	}
	if stat.Mode&unix.S_IFMT != unix.S_IFDIR {
		return "", ErrInvalidIdentity
	}
	return fmt.Sprint(stat.Dev), nil
}

func syncInstallationDirectory(path string) (err error) {
	defer func() { err = installationPathError("sync directory", path, err) }()
	before, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !before.IsDir() {
		return ErrInvalidIdentity
	}
	fd, err := unix.Open(path, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_NONBLOCK|unix.O_CLOEXEC, 0)
	if err != nil {
		return &os.PathError{Op: "open installation directory", Path: path, Err: err}
	}
	file := os.NewFile(uintptr(fd), path)
	defer func() { err = errors.Join(err, file.Close()) }()
	opened, err := file.Stat()
	if err != nil || !os.SameFile(before, opened) {
		return errors.Join(ErrInvalidIdentity, err)
	}
	if err := wrapInstallationError("sync directory", path, file.Sync()); err != nil {
		return err
	}
	after, err := os.Lstat(path)
	if err != nil || !after.IsDir() || !os.SameFile(opened, after) {
		return errors.Join(ErrInvalidIdentity, err)
	}
	return nil
}

func installationPathError(operation, path string, err error) error {
	if err == nil {
		return nil
	}
	return &os.PathError{Op: "installation " + operation, Path: path, Err: err}
}

func wrapInstallationError(operation, path string, err error) error {
	if err == nil {
		return nil
	}
	var pathError *os.PathError
	// A directory-sync operation also opens/checks/closes its descriptor. Only
	// the actual Sync error (or a direct injected sync errno) qualifies EINVAL.
	syncError := strings.HasPrefix(operation, "sync ") && (!errors.As(err, &pathError) || pathError.Op == "sync")
	if errors.Is(err, unix.ENOTSUP) || errors.Is(err, unix.EOPNOTSUPP) || errors.Is(err, unix.ENOSYS) || (syncError && errors.Is(err, unix.EINVAL)) {
		err = errors.Join(ErrInstallationUnsupported, err)
	}
	return installationPathError(operation, path, err)
}
