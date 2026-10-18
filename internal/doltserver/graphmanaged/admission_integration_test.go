//go:build cgo && graphmanaged_engine && (darwin || linux)

package graphmanaged

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// Fixture metadata, not admission evidence. The companion test binary keeps
// graphsession's private API private. There is exactly one generation/worker.
type admissionInput struct {
	Protocol   int    `json:"protocol"`
	Root       string `json:"root"`
	Endpoint   string `json:"endpoint"`
	RunID      string `json:"runId"`
	Generation uint64 `json:"generation"`
}

func admissionReceipt(output string, input admissionInput) error {
	next, done := 1, false
	for _, line := range strings.Split(output, "\n") {
		if !strings.HasPrefix(line, "E1_RESULT ") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 5 || fields[1] != input.RunID || fields[2] != strconv.FormatUint(input.Generation, 10) || done {
			return errors.New("invalid admission receipt identity/order")
		}
		if fields[3] == "done" {
			if fields[4] != "6" || next != 7 {
				return errors.New("incomplete admission receipt")
			}
			done = true
			continue
		}
		if fields[3] != strconv.Itoa(next) || fields[4] != "pass" || next > 6 {
			return errors.New("invalid admission scenario receipt")
		}
		next++
	}
	if !done {
		return errors.New("missing admission completion receipt")
	}
	return nil
}

func runAdmissionWorker(ctx context.Context, executable string, input admissionInput, control string) (string, error) {
	if _, ok := ctx.Deadline(); !ok {
		return "", errors.New("worker requires deadline")
	}
	raw, err := json.Marshal(input)
	if err != nil {
		return "", err
	}
	testName := "TestAdmissionEngineWorker"
	if control != "" {
		testName = "TestAdmissionHarnessChild"
	}
	cmd := exec.CommandContext(ctx, executable, "-test.run=^"+testName+"$", "-test.v", "-test.timeout=5m") //nolint:gosec // G702: caller pins the companion binary or uses this test executable; argv is fixed, without a shell.
	cmd.Dir = input.Root
	cmd.Env = []string{"PATH=/usr/bin:/bin", "HOME=" + filepath.Join(input.Root, "home"), "TMPDIR=" + filepath.Join(input.Root, "tmp"), "TMP=" + filepath.Join(input.Root, "tmp"), "TEMP=" + filepath.Join(input.Root, "tmp"), "GRAPH_E1_WORKER=1", "DOLT_METRICS_DISABLED=1", "DOLT_DISABLE_EVENT_FLUSH=1", "TZ=UTC", "GOMAXPROCS=2", "GOMEMLIMIT=8GiB"}
	if control != "" {
		cmd.Env = append(cmd.Env, "GRAPH_E1_CONTROL="+control)
	}
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
		return output.String(), errors.Join(err, ctx.Err(), fmt.Errorf("admission worker failed: overflow=%t", output.overflow))
	}
	return output.String(), admissionReceipt(output.String(), input)
}

func TestManagedAdmissionExperiment(t *testing.T) {
	if os.Getenv("BEADS_TEST_E1_ADMISSION_FIXTURE") != "1" {
		t.Skip("opt-in fixture-only admission experiment")
	}
	for _, key := range []string{"GRAPH_E1_WORKER", "GRAPH_E1_CONTROL", "BEADS_TEST_MANAGED_GRAPH_FIXTURE", "GRAPH_MANAGED_WORKER", "GRAPH_WORKER_CONTROL", "GRAPH_READ_FIXTURE_PHASE", "GRAPH_READ_FIXTURE_DATA"} {
		if os.Getenv(key) != "" {
			t.Fatalf("polluted parent: %s", key)
		}
	}
	if os.Getenv("DOLT_METRICS_DISABLED") != "1" || os.Getenv("DOLT_DISABLE_EVENT_FLUSH") != "1" {
		t.Fatal("metrics must be disabled before startup")
	}
	worker, err := pinnedGraphWorker(os.Getenv("GRAPH_E1_WORKER_BINARY"), os.Getenv("GRAPH_E1_WORKER_SHA256"))
	if err != nil {
		t.Fatal(err)
	}
	root, err := os.MkdirTemp("/private/tmp", "janet-e1-fixture-")
	if err != nil {
		t.Fatal(err)
	}
	// Register first: every later cleanup runs before deciding to retain/remove.
	t.Cleanup(func() {
		if t.Failed() {
			t.Logf("retained failed private fixture: %s", root)
			return
		}
		if err := os.RemoveAll(root); err != nil {
			t.Error(err)
		}
	})
	for ancestor := root; ; ancestor = filepath.Dir(ancestor) {
		if _, err := os.Lstat(filepath.Join(ancestor, ".beads")); !os.IsNotExist(err) {
			t.Fatal("refusing .beads ancestor")
		}
		if filepath.Dir(ancestor) == ancestor {
			break
		}
	}
	for _, dir := range []string{"home", "tmp", "data", "security"} {
		if err := os.Mkdir(filepath.Join(root, dir), 0700); err != nil {
			t.Fatal(err)
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
	a := fixtureAdmission(t, "unused")
	a.cwd = filepath.Join(root, "data")
	a.environment = []string{"HOME=" + filepath.Join(root, "home"), "TMPDIR=" + filepath.Join(root, "tmp"), "TMP=" + filepath.Join(root, "tmp"), "TEMP=" + filepath.Join(root, "tmp"), "DOLT_METRICS_DISABLED=1", "DOLT_DISABLE_EVENT_FLUSH=1", "TZ=UTC", "GOMAXPROCS=2", "GOMEMLIMIT=8GiB"}
	a.argv = []string{"-test.run=^TestManagedEngineChild$", "managed-engine", "--root=" + root, "--profile=" + a.profile, "--config=" + a.config, "--registration=" + a.registration}
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Minute)
	defer cancel()
	owner := &controller{observe: recordOwner}
	g, err := owner.startWithTiming(ctx, a, timing{startup: 25 * time.Second, cleanup: 15 * time.Second, grace: 5 * time.Second, pipe: 100 * time.Millisecond})
	if g != nil {
		t.Cleanup(func() {
			closeCtx, stop := context.WithTimeout(context.Background(), 18*time.Second)
			defer stop()
			if err := g.close(closeCtx); !onlyProcessGone(err) {
				t.Errorf("engine close: %v", err)
			}
			select {
			case <-g.waitDone:
			default:
				t.Error("engine not reaped")
			}
			if g.expected.Socket != "" {
				conn, err := net.DialTimeout("tcp", g.expected.Socket, 250*time.Millisecond)
				if err == nil {
					_ = conn.Close()
					t.Error("owned endpoint remains reachable")
				}
			}
		})
	}
	if err != nil {
		t.Fatal(err)
	}
	if !g.valid() {
		t.Fatal("unpublished generation")
	}
	input := admissionInput{Protocol: 1, Root: root, Endpoint: g.expected.Socket, RunID: hex.EncodeToString(id), Generation: g.sequence}
	t.Logf("root=%s generation=%d pid=%d endpoint=%s workerSHA=%s", root, g.sequence, g.cmd.Process.Pid, input.Endpoint, os.Getenv("GRAPH_E1_WORKER_SHA256"))
	worker, err = pinnedGraphWorker(worker, os.Getenv("GRAPH_E1_WORKER_SHA256"))
	if err != nil {
		t.Fatal(err)
	}
	workerCtx, stop := context.WithTimeout(ctx, 240*time.Second)
	defer stop()
	output, err := runAdmissionWorker(workerCtx, worker, input, "")
	t.Logf("admission worker observations:\n%s", output)
	if err != nil {
		t.Fatal(err)
	}
	// Close/reap and root cleanup run before the Go test can report PASS.
}

func TestAdmissionReceipt(t *testing.T) {
	input := admissionInput{RunID: strings.Repeat("a", 32), Generation: 1}
	var b strings.Builder
	for i := 1; i <= 6; i++ {
		fmt.Fprintf(&b, "E1_RESULT %s 1 %d pass\n", input.RunID, i)
	}
	complete := b.String() + "E1_RESULT " + input.RunID + " 1 done 6\n"
	if err := admissionReceipt(complete, input); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{"", b.String(), complete + complete, strings.Replace(complete, " 1 3 pass", " 1 2 pass", 1), strings.Replace(complete, input.RunID, strings.Repeat("b", 32), 1), strings.Replace(complete, " done 6", " done 5", 1), strings.Replace(complete, " 1 4 pass", " 1 4 fail", 1)} {
		if admissionReceipt(bad, input) == nil {
			t.Fatalf("accepted malformed receipt %q", bad)
		}
	}
}

func TestAdmissionHarness(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	input := admissionInput{Protocol: 1, Root: t.TempDir(), Endpoint: "127.0.0.1:1234", RunID: strings.Repeat("a", 32), Generation: 1}
	for _, mode := range []string{"success", "fail", "missing", "duplicate", "wrong-run", "overflow", "timeout"} {
		t.Run(mode, func(t *testing.T) {
			bound := 10 * time.Second
			if mode == "timeout" {
				bound = 100 * time.Millisecond
			}
			ctx, cancel := context.WithTimeout(t.Context(), bound)
			defer cancel()
			_, err := runAdmissionWorker(ctx, executable, input, mode)
			if (mode == "success") != (err == nil) {
				t.Fatalf("mode=%s error=%v", mode, err)
			}
		})
	}
}

func TestAdmissionHarnessChild(t *testing.T) {
	mode := os.Getenv("GRAPH_E1_CONTROL")
	if mode == "" {
		t.Skip("inert parent-owned subprocess")
	}
	var input admissionInput
	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	for i := 1; i <= 6; i++ {
		fmt.Fprintf(&b, "E1_RESULT %s %d %d pass\n", input.RunID, input.Generation, i)
	}
	fmt.Fprintf(&b, "E1_RESULT %s %d done 6\n", input.RunID, input.Generation)
	out := b.String()
	switch mode {
	case "success":
		fmt.Print(out)
	case "fail":
		fmt.Print(out)
		t.Fatal("intentional nonzero exit")
	case "missing":
	case "duplicate":
		fmt.Print(out + out)
	case "wrong-run":
		fmt.Print(strings.ReplaceAll(out, input.RunID, "wrong"))
	case "overflow":
		fmt.Print(strings.Repeat("x", (1<<20)+1))
		fmt.Print("\n" + out)
	case "timeout":
		time.Sleep(time.Minute)
	default:
		t.Fatal("unknown control")
	}
}
