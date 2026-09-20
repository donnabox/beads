//go:build cgo && graphmanaged_engine && (darwin || linux)

package graphmanaged

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

// Separate emitted test binaries preserve both packages' private APIs. This
// protocol is fixture metadata, never production admission or graph authority.
type graphWorkerInput struct {
	Protocol int    `json:"protocol"`
	Phase    string `json:"phase"`
	Root     string `json:"root"`
	Endpoint string `json:"endpoint"`
	RunID    string `json:"runId"`
}

type graphWorkerOutput struct {
	mu       sync.Mutex
	buffer   bytes.Buffer
	overflow bool
}

func (o *graphWorkerOutput) Write(raw []byte) (int, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	n := len(raw)
	if remaining := (1 << 20) - o.buffer.Len(); n > remaining {
		o.overflow = true
		raw = raw[:remaining]
	}
	_, err := o.buffer.Write(raw)
	return n, err
}

func (o *graphWorkerOutput) String() string {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.buffer.String()
}

func graphWorkerDigest(output string, input graphWorkerInput) (string, error) {
	var digest string
	for _, line := range strings.Split(output, "\n") {
		if !strings.HasPrefix(line, "MANAGED_GRAPH_WORKER ") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 4 || fields[1] != input.RunID || fields[2] != input.Phase || !digestValid(fields[3]) || digest != "" {
			return "", errors.New("invalid or duplicate graph worker receipt")
		}
		digest = fields[3]
	}
	if digest == "" {
		return "", errors.New("missing graph worker receipt")
	}
	return digest, nil
}

// The caller supplies a deadline. Command cancellation owns the worker's process
// group, and Run waits for exit and I/O drain before returning any receipt.
func runGraphWorker(ctx context.Context, executable string, input graphWorkerInput, testName string, extraEnv ...string) (string, string, error) {
	if _, ok := ctx.Deadline(); !ok {
		return "", "", errors.New("graph worker requires deadline")
	}
	raw, err := json.Marshal(input)
	if err != nil {
		return "", "", err
	}
	cmd := exec.CommandContext(ctx, executable, "-test.run=^"+testName+"$", "-test.v", "-test.timeout=3m")
	cmd.Dir = input.Root
	cmd.Env = append([]string{"PATH=/usr/bin:/bin", "HOME=" + filepath.Join(input.Root, "home"), "TMPDIR=" + filepath.Join(input.Root, "tmp"), "TMP=" + filepath.Join(input.Root, "tmp"), "TEMP=" + filepath.Join(input.Root, "tmp"), "GRAPH_MANAGED_WORKER=1", "DOLT_METRICS_DISABLED=1", "DOLT_DISABLE_EVENT_FLUSH=1", "TZ=UTC"}, extraEnv...)
	cmd.Stdin = bytes.NewReader(raw)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return os.ErrProcessDone
		}
		err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
	cmd.WaitDelay = time.Second
	var output graphWorkerOutput
	cmd.Stdout, cmd.Stderr = &output, &output
	err = cmd.Run()
	if err != nil || output.overflow {
		return "", output.String(), errors.Join(err, ctx.Err(), fmt.Errorf("graph worker failed: output overflow=%t", output.overflow))
	}
	digest, err := graphWorkerDigest(output.String(), input)
	return digest, output.String(), err
}

func pinnedGraphWorker(path, expected string) (string, error) {
	canonical, err := filepath.EvalSymlinks(path)
	if err != nil || !filepath.IsAbs(path) || canonical != path || !digestValid(expected) {
		return "", errors.New("explicit canonical graph worker and SHA256 required")
	}
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() || info.Mode().Perm()&0111 == 0 || info.Mode().Perm()&0022 != 0 || info.Size() > 1<<30 {
		return "", errors.New("invalid graph worker executable")
	}
	hash := sha256.New()
	if _, err := io.Copy(hash, io.LimitReader(f, (1<<30)+1)); err != nil {
		return "", err
	}
	if hex.EncodeToString(hash.Sum(nil)) != expected {
		return "", errors.New("graph worker SHA256 mismatch")
	}
	return canonical, nil
}

func TestManagedGraphReadPersistence(t *testing.T) {
	if os.Getenv("BEADS_TEST_MANAGED_GRAPH_FIXTURE") != "1" {
		t.Skip("opt-in two-binary managed graph fixture")
	}
	for _, key := range []string{"GRAPH_MANAGED_WORKER", "GRAPH_WORKER_CONTROL", "GRAPH_READ_FIXTURE_PHASE", "GRAPH_READ_FIXTURE_DATA"} {
		if os.Getenv(key) != "" {
			t.Fatalf("parent fixture env pollution: %s", key)
		}
	}
	if os.Getenv("DOLT_METRICS_DISABLED") != "1" || os.Getenv("DOLT_DISABLE_EVENT_FLUSH") != "1" {
		t.Fatal("metrics/event disablement required before process startup")
	}
	worker, err := pinnedGraphWorker(os.Getenv("GRAPH_MANAGED_WORKER_BINARY"), os.Getenv("GRAPH_MANAGED_WORKER_SHA256"))
	if err != nil {
		t.Fatal(err)
	}
	root := testRoot(t)
	if err := os.Chmod(root, 0700); err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{"home", "tmp", "data", "security"} {
		if err := os.Mkdir(filepath.Join(root, dir), 0700); err != nil {
			t.Fatal(err)
		}
	}
	for ancestor := root; ; ancestor = filepath.Dir(ancestor) {
		if _, err := os.Lstat(filepath.Join(ancestor, ".beads")); !os.IsNotExist(err) {
			t.Fatal("fixture refuses .beads ancestor")
		}
		if filepath.Dir(ancestor) == ancestor {
			break
		}
	}
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		t.Fatal(err)
	}
	mustWrite(t, filepath.Join(root, "security", "password"), []byte(hex.EncodeToString(secret)), 0600)
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		t.Fatal(err)
	}
	input := graphWorkerInput{Protocol: 1, Root: root, RunID: hex.EncodeToString(id)}
	var seedDigest string
	for _, phase := range []string{"seed", "read"} {
		var digest string
		passed := t.Run(phase, func(t *testing.T) {
			a := fixtureAdmission(t, "unused")
			a.cwd = filepath.Join(root, "data")
			a.environment = []string{"HOME=" + filepath.Join(root, "home"), "TMPDIR=" + filepath.Join(root, "tmp"), "TMP=" + filepath.Join(root, "tmp"), "TEMP=" + filepath.Join(root, "tmp"), "DOLT_METRICS_DISABLED=1", "DOLT_DISABLE_EVENT_FLUSH=1", "TZ=UTC"}
			a.argv = []string{"-test.run=^TestManagedEngineChild$", "managed-engine", "--root=" + root, "--profile=" + a.profile, "--config=" + a.config, "--registration=" + a.registration}
			ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
			t.Cleanup(cancel)
			controller := &controller{observe: recordOwner}
			generation, err := controller.startWithTiming(ctx, a, timing{startup: 25 * time.Second, cleanup: 15 * time.Second, grace: 5 * time.Second, pipe: 100 * time.Millisecond})
			// Cleanup is registered before checking startup failure and runs before
			// t.Run returns, so the next generation never overlaps this owner.
			if generation != nil {
				t.Cleanup(func() {
					closeCtx, cancel := context.WithTimeout(context.Background(), 18*time.Second)
					defer cancel()
					if err := generation.close(closeCtx); !onlyProcessGone(err) {
						generation.stderr.mu.Lock()
						stderr := string(generation.stderr.bytes)
						generation.stderr.mu.Unlock()
						t.Errorf("managed graph generation close: %v; stderr: %s", err, stderr)
					}
					select {
					case <-generation.waitDone:
					default:
						t.Error("managed graph engine not reaped")
					}
				})
			}
			if err != nil {
				if generation != nil {
					generation.stderr.mu.Lock()
					stderr := string(generation.stderr.bytes)
					generation.stderr.mu.Unlock()
					t.Logf("managed graph startup stderr: %s", stderr)
				}
				t.Fatal(err)
			}
			if !generation.valid() {
				t.Fatal("unpublished managed generation")
			}
			input.Phase, input.Endpoint = phase, generation.expected.Socket
			t.Logf("phase=%s child=%d endpoint=%s workerSHA256=%s privateRoot=%s metricsDisabled=1", phase, generation.cmd.Process.Pid, input.Endpoint, os.Getenv("GRAPH_MANAGED_WORKER_SHA256"), root)
			workerCtx, stop := context.WithTimeout(ctx, 180*time.Second)
			defer stop()
			// Recheck the companion executable for every generation, immediately
			// before launch, matching the controller's per-start engine recheck.
			worker, err = pinnedGraphWorker(worker, os.Getenv("GRAPH_MANAGED_WORKER_SHA256"))
			if err != nil {
				t.Fatal(err)
			}
			var output string
			digest, output, err = runGraphWorker(workerCtx, worker, input, "TestManagedGraphWorker")
			t.Logf("worker phase=%s output:\n%s", phase, output)
			if err != nil {
				t.Fatal(err)
			}
		})
		if !passed {
			t.FailNow()
		}
		if phase == "seed" {
			seedDigest = digest
			if _, err := os.Stat(filepath.Join(root, "data", "graph_read_fixture", ".dolt")); err != nil {
				t.Fatal(err)
			}
		} else if digest != seedDigest {
			t.Fatal("managed graph digest changed across engine generations")
		}
	}
	if t.Failed() {
		t.Fatal("managed graph controls failed")
	}
	fmt.Printf("MANAGED_GRAPH_FIXTURE_OK %s %s\n", input.RunID, seedDigest)
}

// These subprocess controls execute no SQL/engine and require no companion
// binary. They pin parent failure handling independently of the happy path.
func TestGraphWorkerHarness(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	input := graphWorkerInput{Protocol: 1, Root: t.TempDir(), Endpoint: "127.0.0.1:1234", Phase: "seed", RunID: strings.Repeat("a", 32)}
	for _, mode := range []string{"success", "fail", "missing", "duplicate", "wrong-phase", "overflow", "timeout"} {
		t.Run(mode, func(t *testing.T) {
			bound := 10 * time.Second
			if mode == "timeout" {
				bound = 100 * time.Millisecond
			}
			ctx, cancel := context.WithTimeout(t.Context(), bound)
			defer cancel()
			digest, _, err := runGraphWorker(ctx, executable, input, "TestGraphWorkerHarnessChild", "GRAPH_WORKER_CONTROL="+mode)
			if mode == "success" {
				if err != nil || digest != strings.Repeat("b", 64) {
					t.Fatalf("success: %s %v", digest, err)
				}
			} else if err == nil || digest != "" {
				t.Fatalf("accepted %s worker", mode)
			}
		})
	}
}

func TestGraphWorkerHarnessChild(t *testing.T) {
	mode := os.Getenv("GRAPH_WORKER_CONTROL")
	if mode == "" {
		t.Skip("parent-owned inert worker")
	}
	var input graphWorkerInput
	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		t.Fatal(err)
	}
	line := fmt.Sprintf("MANAGED_GRAPH_WORKER %s %s %s\n", input.RunID, input.Phase, strings.Repeat("b", 64))
	switch mode {
	case "success":
		fmt.Print(line)
	case "fail":
		fmt.Print(line)
		t.Fatal("deliberate exit failure despite receipt")
	case "missing":
	case "duplicate":
		fmt.Print(line + line)
	case "wrong-phase":
		fmt.Print(strings.Replace(line, " seed ", " read ", 1))
	case "overflow":
		fmt.Print(strings.Repeat("x", (1<<20)+1))
		fmt.Print("\n" + line)
	case "timeout":
		time.Sleep(time.Minute)
	default:
		t.Fatal("unknown control")
	}
}
