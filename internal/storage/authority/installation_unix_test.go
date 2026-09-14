//go:build unix

package authority

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/config"
	"github.com/steveyegge/beads/internal/lockfile"
	"github.com/steveyegge/beads/internal/utils"
	"golang.org/x/sys/unix"
)

func installationFixture(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	workspace := filepath.Join(root, "workspace")
	if err := os.Mkdir(workspace, 0700); err != nil {
		t.Fatal(err)
	}
	return workspace, filepath.Join(root, "config", "bd", "installation-id")
}

func mustInstallationKey(t *testing.T, workspace, path string, ops installationIO) string {
	t.Helper()
	key, err := installationKeyAt(context.Background(), workspace, path, ops)
	if err != nil || key == "" {
		t.Fatalf("key %q, error %v", key, err)
	}
	return key
}

func TestInstallationStableBinding(t *testing.T) {
	workspace, path := installationFixture(t)
	key := mustInstallationKey(t, workspace, path, nativeInstallationIO())
	id, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(id) != 65 || id[64] != '\n' {
		t.Fatalf("invalid representation length %d", len(id))
	}
	if _, err := hex.DecodeString(string(id[:64])); err != nil {
		t.Fatal(err)
	}
	canonical, err := utils.CanonicalizeExistingPath(workspace)
	if err != nil {
		t.Fatal(err)
	}
	want := sha256.Sum256([]byte(string(id[:64]) + ":" + canonical))
	if key != hex.EncodeToString(want[:]) {
		t.Fatal("key does not bind persisted ID and canonical directory")
	}
	if again := mustInstallationKey(t, workspace, path, nativeInstallationIO()); again != key {
		t.Fatal("key changed")
	}
	alias := filepath.Join(filepath.Dir(workspace), "alias")
	if err := os.Symlink(workspace, alias); err != nil {
		t.Fatal(err)
	}
	if got := mustInstallationKey(t, alias, path, nativeInstallationIO()); got != key {
		t.Fatal("symlink changed key")
	}
	other := filepath.Join(filepath.Dir(workspace), "other")
	if err := os.Mkdir(other, 0700); err != nil {
		t.Fatal(err)
	}
	if got := mustInstallationKey(t, other, path, nativeInstallationIO()); got == key {
		t.Fatal("distinct workspace shared key")
	}
	for _, name := range []string{path, path + ".lock"} {
		info, err := os.Stat(name)
		if err != nil || info.Mode().Perm() != 0600 {
			t.Fatalf("mode for %s: %v, %v", name, info, err)
		}
	}
	for parent := filepath.Dir(path); parent != filepath.Dir(workspace); parent = filepath.Dir(parent) {
		info, err := os.Stat(parent)
		if err != nil || info.Mode().Perm() != 0700 {
			t.Fatalf("parent mode: %v, %v", info, err)
		}
	}
}

func TestInstallationDefaultAndOverride(t *testing.T) {
	workspace, override := installationFixture(t)
	t.Setenv("HOME", filepath.Dir(workspace))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(filepath.Dir(workspace), "native"))
	t.Setenv("BEADS_INSTALLATION_ID_FILE", "")
	configuration, err := config.UserConfigYamlPath()
	if err != nil {
		t.Fatal(err)
	}
	if key, err := InstallationKey(context.Background(), workspace); err != nil || key == "" {
		t.Fatalf("default: %v", err)
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(configuration), "installation-id")); err != nil {
		t.Fatal(err)
	}
	t.Setenv("BEADS_INSTALLATION_ID_FILE", override)
	if _, err := InstallationKey(context.Background(), workspace); err != nil {
		t.Fatal(err)
	}
	parentAlias := filepath.Join(filepath.Dir(workspace), "config-alias")
	if err := os.Symlink(filepath.Dir(override), parentAlias); err != nil {
		t.Fatal(err)
	}
	first, err := InstallationKey(context.Background(), workspace)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("BEADS_INSTALLATION_ID_FILE", filepath.Join(parentAlias, "installation-id"))
	second, err := InstallationKey(context.Background(), workspace)
	if err != nil || first != second {
		t.Fatalf("parent alias differs: %v", err)
	}
	t.Setenv("BEADS_INSTALLATION_ID_FILE", "relative")
	if key, err := InstallationKey(context.Background(), workspace); key != "" || !errors.Is(err, ErrInvalidIdentity) {
		t.Fatalf("relative override: %q %v", key, err)
	}
	if key, err := InstallationKey(nil, workspace); key != "" || err == nil {
		t.Fatal("nil context accepted")
	}
}

func TestInstallationInvalidWorkspaceHasNoEffects(t *testing.T) {
	workspace, path := installationFixture(t)
	for _, missing := range []string{workspace + "-missing", path} {
		if key, err := installationKeyAt(context.Background(), missing, path, nativeInstallationIO()); key != "" || err == nil {
			t.Fatal("invalid workspace accepted")
		}
	}
	if _, err := os.Stat(filepath.Dir(path)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("created config parent: %v", err)
	}
}

func TestInstallationInvalidResiduePreserved(t *testing.T) {
	for _, value := range []string{"", "abc", strings.Repeat("A", 64) + "\n", strings.Repeat("a", 64), strings.Repeat("a", 64) + "\n\n", strings.Repeat("a", 1000)} {
		t.Run(fmt.Sprintf("bytes%d-%s", len(value), value[:min(1, len(value))]), func(t *testing.T) {
			workspace, path := installationFixture(t)
			if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(value), 0600); err != nil {
				t.Fatal(err)
			}
			if key, err := installationKeyAt(context.Background(), workspace, path, nativeInstallationIO()); key != "" || !errors.Is(err, ErrInvalidIdentity) {
				t.Fatalf("residue: %q %v", key, err)
			}
			got, err := os.ReadFile(path)
			if err != nil || string(got) != value {
				t.Fatal("invalid ID was changed")
			}
		})
	}
}

func TestInstallationSpecialFilesAndModes(t *testing.T) {
	for _, suffix := range []string{"", ".lock"} {
		for _, kind := range []string{"symlink", "fifo", "directory", "permissions"} {
			t.Run(kind+suffix, func(t *testing.T) {
				workspace, path := installationFixture(t)
				if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
					t.Fatal(err)
				}
				target := filepath.Join(filepath.Dir(workspace), "target")
				if err := os.WriteFile(target, []byte("untouched"), 0600); err != nil {
					t.Fatal(err)
				}
				var err error
				switch kind {
				case "symlink":
					err = os.Symlink(target, path+suffix)
				case "fifo":
					err = unix.Mkfifo(path+suffix, 0600)
				case "directory":
					err = os.Mkdir(path+suffix, 0700)
				case "permissions":
					err = os.WriteFile(path+suffix, []byte(strings.Repeat("a", 64)+"\n"), 0600)
					if err == nil {
						err = os.Chmod(path+suffix, 0644)
					}
				}
				if err != nil {
					t.Fatal(err)
				}
				if key, err := installationKeyAt(context.Background(), workspace, path, nativeInstallationIO()); key != "" || !errors.Is(err, ErrInvalidIdentity) {
					t.Fatalf("special file: %q %v", key, err)
				}
				got, err := os.ReadFile(target)
				if err != nil || string(got) != "untouched" {
					t.Fatal("symlink target changed")
				}
			})
		}
	}
}

func TestInstallationDurabilityRetryFlushesExistingAncestors(t *testing.T) {
	workspace, path := installationFixture(t)
	ancestor, err := utils.CanonicalizeExistingPath(filepath.Dir(workspace))
	if err != nil {
		t.Fatal(err)
	}
	fault := errors.New("upper ancestor flush failed")
	ops := nativeInstallationIO()
	ops.syncDir = func(p string) error {
		_, idErr := os.Stat(path)
		if p == ancestor && idErr == nil {
			return fault
		}
		return syncInstallationDirectory(p)
	}
	if key, err := installationKeyAt(context.Background(), workspace, path, ops); key != "" || !errors.Is(err, fault) {
		t.Fatalf("flush failure: %q %v", key, err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var flushed []string
	ops.syncDir = func(p string) error { flushed = append(flushed, p); return syncInstallationDirectory(p) }
	mustInstallationKey(t, workspace, path, ops)
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("retry regenerated ID")
	}
	want, err := utils.CanonicalizeExistingPath(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	for _, got := range flushed {
		if got != want {
			t.Fatalf("flushed %q, want %q", got, want)
		}
		want = filepath.Dir(want)
	}
	if len(flushed) == 0 {
		t.Fatal("did not flush any ancestor")
	}
	last := flushed[len(flushed)-1]
	if next := filepath.Dir(last); next != last {
		lastDevice, err := installationDevice(last)
		if err != nil {
			t.Fatal(err)
		}
		nextDevice, err := installationDevice(next)
		if err != nil {
			t.Fatal(err)
		}
		if lastDevice == nextDevice {
			t.Fatalf("stopped before device/root anchor: %v", flushed)
		}
	}
}

func TestInstallationFaultsReturnNoKey(t *testing.T) {
	for _, stage := range []string{"entropy", "sync", "close", "unlock", "unsupported", "replacement"} {
		t.Run(stage, func(t *testing.T) {
			workspace, path := installationFixture(t)
			ops := nativeInstallationIO()
			fault := errors.New(stage)
			switch stage {
			case "entropy":
				ops.random = func([]byte) (int, error) { return 0, fault }
			case "sync":
				ops.syncFile = func(*os.File) error { return fault }
			case "close":
				ops.closeFile = func(f *os.File) error { return errors.Join(f.Close(), fault) }
			case "unlock":
				ops.unlock = func(f *os.File) error { return errors.Join(lockfile.FlockUnlock(f), fault) }
			case "unsupported":
				fault = ErrInstallationUnsupported
				ops.syncFile = func(*os.File) error { return unix.ENOTSUP }
			case "replacement":
				fault = ErrInvalidIdentity
				ops.syncFile = func(f *os.File) error {
					if err := f.Sync(); err != nil {
						return err
					}
					if err := os.Rename(path, path+".old"); err != nil {
						return err
					}
					return os.WriteFile(path, []byte(strings.Repeat("b", 64)+"\n"), 0600)
				}
			}
			key, err := installationKeyAt(context.Background(), workspace, path, ops)
			if key != "" || !errors.Is(err, fault) {
				t.Fatalf("fault returned %q %v", key, err)
			}
			if stage == "entropy" {
				if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
					t.Fatal("entropy failure created ID")
				}
			}
		})
	}
}

func TestInstallationLockBound(t *testing.T) {
	workspace, path := installationFixture(t)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	lock, err := openInstallationFile(path+".lock", true)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := lock.Close(); err != nil {
			t.Error(err)
		}
	}()
	if err := lockfile.FlockExclusiveNonBlocking(lock); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := lockfile.FlockUnlock(lock); err != nil {
			t.Error(err)
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 75*time.Millisecond)
	defer cancel()
	key, err := installationKeyAt(ctx, workspace, path, nativeInstallationIO())
	if key != "" || errors.Is(err, ErrInstallationBusy) || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("contention: %q %v", key, err)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("contender created ID")
	}
	ops := nativeInstallationIO()
	ops.lock = func(*os.File) error { return lockfile.ErrLockBusy }
	start := time.Now()
	key, err = installationKeyAt(context.Background(), workspace, path, ops)
	if key != "" || !errors.Is(err, ErrInstallationBusy) || time.Since(start) > 4*time.Second {
		t.Fatalf("wait cap: %q %v %v", key, err, time.Since(start))
	}
}

// This child is the test binary only. The parent bounds and reaps every child.
func TestInstallationProcess(t *testing.T) {
	workspace := os.Getenv("BEADS_INSTALLATION_TEST_WORKSPACE")
	if workspace == "" {
		t.Skip("process helper; driven by the parent test")
	}
	if os.Getenv("BEADS_INSTALLATION_TEST_CRASH") == "1" {
		path := os.Getenv("BEADS_INSTALLATION_ID_FILE")
		lock, err := openInstallationFile(path+".lock", true)
		if err != nil {
			t.Fatal(err)
		}
		if err := lockfile.FlockExclusiveNonBlocking(lock); err != nil {
			t.Fatal(err)
		}
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := file.WriteString("partial"); err != nil {
			t.Fatal(err)
		}
		if err := file.Sync(); err != nil {
			t.Fatal(err)
		}
		fmt.Println("partial-ready")
		time.Sleep(30 * time.Second) // Parent kills and reaps this test-only child.
		_ = file.Close()
		_ = lockfile.FlockUnlock(lock)
		_ = lock.Close()
		t.Fatal("parent did not terminate crash fixture")
	}
	key, err := InstallationKey(context.Background(), workspace)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("installation-key=" + key)
}

func TestInstallationCreatorDeathPreservesPartialID(t *testing.T) {
	workspace, path := installationFixture(t)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, executable, "-test.run=^TestInstallationProcess$", "-test.timeout=15s")
	cmd.Env = append(os.Environ(), "BEADS_INSTALLATION_TEST_WORKSPACE="+workspace, "BEADS_INSTALLATION_ID_FILE="+path, "BEADS_INSTALLATION_TEST_CRASH=1")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	waited := false
	defer func() {
		if !waited {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
	}()
	ready := make(chan bool, 1)
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			if scanner.Text() == "partial-ready" {
				ready <- true
				return
			}
		}
		ready <- false
	}()
	select {
	case ok := <-ready:
		if !ok {
			t.Fatal("child exited before partial-write barrier")
		}
	case <-ctx.Done():
		t.Fatal("child did not reach partial-write barrier")
	}
	short, stop := context.WithTimeout(ctx, 75*time.Millisecond)
	key, err := installationKeyAt(short, workspace, path, nativeInstallationIO())
	stop()
	if key != "" || !errors.Is(err, context.DeadlineExceeded) || errors.Is(err, ErrInstallationBusy) {
		t.Fatalf("live partial creator was not excluded: %q %v", key, err)
	}
	if err := cmd.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	err = cmd.Wait()
	waited = true
	if err == nil {
		t.Fatal("crash fixture unexpectedly exited successfully")
	}
	key, err = installationKeyAt(context.Background(), workspace, path, nativeInstallationIO())
	if key != "" || !errors.Is(err, ErrInvalidIdentity) {
		t.Fatalf("dead creator residue: %q %v", key, err)
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "partial" {
		t.Fatal("partial identity was replaced")
	}
}

func TestInstallationNativePathBytes(t *testing.T) {
	workspace, path := installationFixture(t)
	native := filepath.Join(workspace, string([]byte{'x', 0xff}))
	if err := os.Mkdir(native, 0700); err != nil {
		if errors.Is(err, unix.EILSEQ) {
			t.Skip("fixture filesystem refuses non-UTF-8 filenames (EILSEQ); native arbitrary-byte path behavior is unqualified here")
		}
		t.Fatalf("native fixture filename: %v", err)
	}
	mustInstallationKey(t, native, path, nativeInstallationIO())
}

func TestInstallationExclusiveCreateRereadsWinner(t *testing.T) {
	workspace, path := installationFixture(t)
	winner := strings.Repeat("a", 64) + "\n"
	ops := nativeInstallationIO()
	ops.random = func(p []byte) (int, error) {
		if err := os.WriteFile(path, []byte(winner), 0600); err != nil {
			return 0, err
		}
		for i := range p {
			p[i] = 0xbb
		}
		return len(p), nil
	}
	key := mustInstallationKey(t, workspace, path, ops)
	canonical, err := utils.CanonicalizeExistingPath(workspace)
	if err != nil {
		t.Fatal(err)
	}
	want := sha256.Sum256([]byte(winner[:64] + ":" + canonical))
	if key != hex.EncodeToString(want[:]) {
		t.Fatal("returned generated loser instead of O_EXCL winner")
	}
}

func TestInstallationCleanupPreservesPrimaryError(t *testing.T) {
	workspace, path := installationFixture(t)
	primary, cleanup := errors.New("sync failure"), errors.New("close failure")
	ops := nativeInstallationIO()
	ops.syncFile = func(*os.File) error { return primary }
	ops.closeFile = func(f *os.File) error { return errors.Join(f.Close(), cleanup) }
	key, err := installationKeyAt(context.Background(), workspace, path, ops)
	if key != "" || !errors.Is(err, primary) || !errors.Is(err, cleanup) {
		t.Fatalf("lost error: %q %v", key, err)
	}
}

func TestInstallationConcurrentProcesses(t *testing.T) {
	workspace, path := installationFixture(t)
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	var commands []*exec.Cmd
	var outputs []*bytes.Buffer
	for range 6 {
		cmd := exec.CommandContext(ctx, executable, "-test.run=^TestInstallationProcess$", "-test.timeout=10s")
		cmd.Env = append(os.Environ(), "BEADS_INSTALLATION_TEST_WORKSPACE="+workspace, "BEADS_INSTALLATION_ID_FILE="+path)
		output := new(bytes.Buffer)
		cmd.Stdout, cmd.Stderr = output, output
		if err := cmd.Start(); err != nil {
			cancel()
			for _, started := range commands {
				_ = started.Wait()
			}
			t.Fatal(err)
		}
		commands = append(commands, cmd)
		outputs = append(outputs, output)
	}
	var keys []string
	for i, cmd := range commands {
		if err := cmd.Wait(); err != nil {
			t.Errorf("child: %v: %s", err, outputs[i])
			continue
		}
		for _, line := range strings.Split(outputs[i].String(), "\n") {
			if key, ok := strings.CutPrefix(line, "installation-key="); ok {
				keys = append(keys, key)
			}
		}
	}
	if len(keys) != len(commands) {
		t.Fatalf("only %d child keys", len(keys))
	}
	for _, key := range keys {
		if key != keys[0] {
			t.Fatal("creators disagreed")
		}
	}
	if key := mustInstallationKey(t, workspace, path, nativeInstallationIO()); key != keys[0] {
		t.Fatal("reopen differs")
	}
}

func TestInstallationInvalidFilePathsHaveNoEffects(t *testing.T) {
	for _, suffix := range []string{"/.", "/..", "/", "//", "root", "relative"} {
		t.Run(suffix, func(t *testing.T) {
			workspace, _ := installationFixture(t)
			parent := filepath.Join(filepath.Dir(workspace), "uncreated")
			path := parent + suffix
			if suffix == "root" {
				path = string(filepath.Separator)
			}
			if suffix == "relative" {
				path = "relative"
			}
			key, err := installationKeyAt(context.Background(), workspace, path, nativeInstallationIO())
			if key != "" || !errors.Is(err, ErrInvalidIdentity) || !strings.Contains(err.Error(), path) {
				t.Fatalf("invalid path: %q %v", key, err)
			}
			entries, err := os.ReadDir(filepath.Dir(workspace))
			if err != nil || len(entries) != 1 || entries[0].Name() != "workspace" {
				t.Fatalf("invalid path created entries: %v %v", entries, err)
			}
		})
	}
}

func TestInstallationErrorsIdentifyPathWithoutContents(t *testing.T) {
	for _, suffix := range []string{"", ".lock"} {
		t.Run("path"+suffix, func(t *testing.T) {
			workspace, path := installationFixture(t)
			if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
				t.Fatal(err)
			}
			contents := "private invalid identity contents"
			if err := os.WriteFile(path+suffix, []byte(contents), 0600); err != nil {
				t.Fatal(err)
			}
			if suffix != "" {
				if err := os.Chmod(path+suffix, 0644); err != nil {
					t.Fatal(err)
				}
			}
			_, err := installationKeyAt(context.Background(), workspace, path, nativeInstallationIO())
			canonical, canonicalErr := utils.CanonicalizeExistingPath(path + suffix)
			if canonicalErr != nil {
				t.Fatal(canonicalErr)
			}
			var pathErr *os.PathError
			if !errors.Is(err, ErrInvalidIdentity) || !errors.As(err, &pathErr) || pathErr.Op == "" || !strings.Contains(err.Error(), canonical) || strings.Contains(err.Error(), contents) {
				t.Fatalf("error lost context or exposed contents: %v", err)
			}
		})
	}
}

func TestInstallationSyncStopsAtDeviceBoundary(t *testing.T) {
	workspace, path := installationFixture(t)
	mustInstallationKey(t, workspace, path, nativeInstallationIO())
	anchor, err := utils.CanonicalizeExistingPath(filepath.Dir(workspace))
	if err != nil {
		t.Fatal(err)
	}
	foreign := filepath.Dir(anchor)
	ops := nativeInstallationIO()
	nativeDevice := ops.device
	ops.device = func(p string) (string, error) {
		device, err := nativeDevice(p)
		if p == foreign {
			return device + "-foreign", err
		}
		return device, err
	}
	var flushed []string
	ops.syncDir = func(p string) error {
		if p == foreign {
			t.Fatal("opened foreign-device ancestor for sync")
		}
		flushed = append(flushed, p)
		return syncInstallationDirectory(p)
	}
	mustInstallationKey(t, workspace, path, ops)
	if len(flushed) == 0 || flushed[len(flushed)-1] != anchor {
		t.Fatalf("anchor not flushed: %v", flushed)
	}
	fault := errors.New("ancestor stat fault")
	ops.device = func(p string) (string, error) {
		if p == foreign {
			return "", fault
		}
		return nativeDevice(p)
	}
	key, err := installationKeyAt(context.Background(), workspace, path, ops)
	if key != "" || !errors.Is(err, fault) || !strings.Contains(err.Error(), foreign) {
		t.Fatalf("stat failure ignored: %q %v", key, err)
	}
}

func TestInstallationEINVALClassification(t *testing.T) {
	for _, stage := range []string{"sync ID", "sync directory", "directory close", "lock", "unlock", "close"} {
		t.Run(stage, func(t *testing.T) {
			workspace, path := installationFixture(t)
			ops := nativeInstallationIO()
			switch stage {
			case "sync ID":
				ops.syncFile = func(*os.File) error { return unix.EINVAL }
			case "sync directory":
				ops.syncDir = func(string) error { return unix.EINVAL }
			case "directory close":
				ops.syncDir = func(p string) error { return &os.PathError{Op: "close", Path: p, Err: unix.EINVAL} }
			case "lock":
				ops.lock = func(*os.File) error { return unix.EINVAL }
			case "unlock":
				ops.unlock = func(f *os.File) error { return errors.Join(lockfile.FlockUnlock(f), unix.EINVAL) }
			case "close":
				ops.closeFile = func(f *os.File) error { return errors.Join(f.Close(), unix.EINVAL) }
			}
			key, err := installationKeyAt(context.Background(), workspace, path, ops)
			if key != "" || !errors.Is(err, unix.EINVAL) || errors.Is(err, ErrInstallationUnsupported) != strings.HasPrefix(stage, "sync ") {
				t.Fatalf("classification: %q %v", key, err)
			}
		})
	}
}

func TestInstallationCallerCancellationDuringLock(t *testing.T) {
	workspace, path := installationFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ops := nativeInstallationIO()
	ops.lock = func(*os.File) error { cancel(); return lockfile.ErrLocked }
	key, err := installationKeyAt(ctx, workspace, path, ops)
	if key != "" || !errors.Is(err, context.Canceled) || errors.Is(err, ErrInstallationBusy) {
		t.Fatalf("cancellation: %q %v", key, err)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("canceled contender created ID")
	}
}

func TestInstallationCreationAndSyncHoldSidecar(t *testing.T) {
	workspace, path := installationFixture(t)
	ops := nativeInstallationIO()
	random, sync := ops.random, ops.syncFile
	assertLocked := func() {
		probe, err := openInstallationFile(path+".lock", false)
		if err != nil {
			t.Fatal(err)
		}
		err = lockfile.FlockExclusiveNonBlocking(probe)
		if err == nil {
			_ = lockfile.FlockUnlock(probe)
		}
		closeErr := probe.Close()
		if !errors.Is(err, lockfile.ErrLocked) && !errors.Is(err, lockfile.ErrLockBusy) {
			t.Fatalf("ID persistence outside sidecar lock: %v", err)
		}
		if closeErr != nil {
			t.Fatal(closeErr)
		}
	}
	ops.random = func(p []byte) (int, error) { assertLocked(); return random(p) }
	ops.syncFile = func(f *os.File) error { assertLocked(); return sync(f) }
	mustInstallationKey(t, workspace, path, ops)
}

type installationOwnerInfo struct {
	os.FileInfo
	stat syscall.Stat_t
}

func (i installationOwnerInfo) Sys() any { return &i.stat }

func TestInstallationFileRequiresCurrentOwner(t *testing.T) {
	workspace, path := installationFixture(t)
	mustInstallationKey(t, workspace, path, nativeInstallationIO())
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		t.Fatal("native file stat unavailable")
	}
	other := installationOwnerInfo{FileInfo: info, stat: *stat}
	other.stat.Uid++
	if !privateInstallationFile(info) || privateInstallationFile(other) {
		t.Fatal("private file owner not enforced")
	}
}
