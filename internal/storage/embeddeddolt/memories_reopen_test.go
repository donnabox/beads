//go:build cgo

package embeddeddolt_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/storage/embeddeddolt"
	"github.com/steveyegge/beads/memoryops"
)

// TestMemoriesPersistAfterFinalClose covers the keyed memory plane across a
// real store close and reopen. This is not graph memory, revision history, or a
// process restart: both store lifetimes run in the same test process.
func TestMemoriesPersistAfterFinalClose(t *testing.T) {
	skipUnlessEmbeddedDolt(t)
	ctx := t.Context()
	beadsDir := filepath.Join(t.TempDir(), ".beads")
	const key = "launch/reopen-proof"
	const content = "Launch moves to Monday at 14:00.\nReason: recovery check — preserve this exact text.\n"

	open := func() *embeddeddolt.EmbeddedDoltStore {
		t.Helper()
		store, err := embeddeddolt.Open(ctx, beadsDir, "memreopen", "main")
		if err != nil {
			t.Fatalf("Open: %v", err)
		}
		t.Cleanup(func() {
			if !store.IsClosed() {
				if err := store.Close(); err != nil {
					t.Errorf("cleanup Close: %v", err)
				}
			}
		})
		return store
	}
	closeStore := func(store *embeddeddolt.EmbeddedDoltStore) {
		t.Helper()
		if err := store.Close(); err != nil {
			t.Fatalf("final Close: %v", err)
		}
		if !store.IsClosed() {
			t.Fatal("final Close left the store open")
		}
	}

	store := open()
	memories, err := store.Memories()
	if err != nil {
		t.Fatalf("Memories before close: %v", err)
	}
	remembered, err := memories.Remember(ctx, memoryops.RememberRequest{Key: key, Content: content})
	if err != nil {
		t.Fatalf("Remember: %v", err)
	}
	if remembered.Key != key || remembered.Value != content || remembered.Replaced {
		t.Fatalf("Remember = %#v, want exact new key and content", remembered)
	}
	if err := store.Commit(ctx, "persist memory before final close"); err != nil {
		t.Fatalf("Commit: %v", err)
	}
	recalled, err := memories.Recall(ctx, memoryops.RecallRequest{Key: key})
	if err != nil || !recalled.Found || recalled.Key != key || recalled.Value != content {
		t.Fatalf("Recall before close = %#v, %v; want exact persisted key and content", recalled, err)
	}
	closeStore(store)

	reopened := open()
	if reopened == store {
		t.Fatal("Open returned the closed store")
	}
	memories, err = reopened.Memories()
	if err != nil {
		t.Fatalf("Memories after reopen: %v", err)
	}
	recalled, err = memories.Recall(ctx, memoryops.RecallRequest{Key: key})
	if err != nil {
		t.Fatalf("Recall after reopen: %v", err)
	}
	if !recalled.Found || recalled.Key != key || recalled.Value != content {
		t.Fatalf("Recall after reopen = %#v, want exact persisted key and content", recalled)
	}
	closeStore(reopened)
}

// TestMemoriesPersistAcrossProcesses additionally crosses an OS-process
// boundary, so no store or engine cache from the writer can satisfy the read.
// It exercises the legacy keyed memory plane, not graph or BDP history.
func TestMemoriesPersistAcrossProcesses(t *testing.T) {
	skipUnlessEmbeddedDolt(t)
	const phaseEnv = "BEADS_TEST_MEMORY_REOPEN_PHASE"
	const dirEnv = "BEADS_TEST_MEMORY_REOPEN_DIR"
	phase := os.Getenv(phaseEnv)
	if phase == "" {
		binary, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		beadsDir := filepath.Join(t.TempDir(), ".beads")
		// Keep the phases inline: a subtest filter must not skip the write.
		for _, childPhase := range []string{"write", "read"} {
			childTemp := t.TempDir()
			ctx, cancel := context.WithTimeout(t.Context(), 4*time.Minute)
			// Preserve a fresh database to cover the first write after schema
			// initialization; allow cold migrations under the race detector.
			cmd := exec.CommandContext(ctx, binary, "-test.v", "-test.run=^TestMemoriesPersistAcrossProcesses$", "-test.timeout=3m", "-test.count=1")
			cmd.Env = append(os.Environ(), phaseEnv+"="+childPhase, dirEnv+"="+beadsDir,
				"TMPDIR="+childTemp, "TMP="+childTemp, "TEMP="+childTemp)
			cmd.WaitDelay = 5 * time.Second
			output, err := cmd.CombinedOutput()
			cancel()
			// Go uses the first coverage summary in the output, so distinguish
			// child summaries while preserving their numbers for diagnostics.
			loggedOutput := strings.ReplaceAll("\n"+string(output), "\ncoverage: ", "\nchild coverage = ")
			t.Logf("%s child output:%s", childPhase, loggedOutput)
			if err != nil {
				t.Fatalf("%s process: %v\n%s", childPhase, err, output)
			}
			if !strings.Contains(string(output), "memory-reopen-phase="+childPhase+" completed") {
				t.Fatalf("%s process did not complete the memory operation:\n%s", childPhase, output)
			}
		}
		return
	}
	if phase != "write" && phase != "read" {
		t.Fatalf("unknown memory reopen phase %q", phase)
	}
	beadsDir := os.Getenv(dirEnv)
	if !filepath.IsAbs(beadsDir) {
		t.Fatal("memory reopen child requires its parent's absolute fixture path")
	}
	store, err := embeddeddolt.Open(t.Context(), beadsDir, "memprocess", "main")
	if err != nil {
		t.Fatalf("%s Open: %v", phase, err)
	}
	t.Cleanup(func() {
		if !store.IsClosed() {
			if err := store.Close(); err != nil {
				t.Errorf("%s cleanup Close: %v", phase, err)
			}
		}
	})
	memories, err := store.Memories()
	if err != nil {
		t.Fatalf("%s Memories: %v", phase, err)
	}
	const key = "launch/process-proof"
	const content = "Monday 14:00\nKeep the recovery decision — exactly.\n"
	if phase == "write" {
		remembered, err := memories.Remember(t.Context(), memoryops.RememberRequest{Key: key, Content: content})
		if err != nil {
			t.Fatalf("Remember: %v", err)
		}
		if remembered.Replaced || remembered.Key != key || remembered.Value != content {
			t.Fatalf("Remember = %#v, want exact new key and content", remembered)
		}
		if err := store.Commit(t.Context(), "persist memory before process exit"); err != nil {
			t.Fatalf("Commit: %v", err)
		}
	} else {
		recalled, err := memories.Recall(t.Context(), memoryops.RecallRequest{Key: key})
		if err != nil || !recalled.Found || recalled.Key != key || recalled.Value != content {
			t.Fatalf("Recall in new process = %#v, %v; want exact persisted key and content", recalled, err)
		}
	}
	if err := store.Close(); err != nil {
		t.Fatalf("%s final Close: %v", phase, err)
	}
	if !store.IsClosed() {
		t.Fatalf("%s final Close left the store open", phase)
	}
	t.Logf("memory-reopen-phase=%s completed (pid=%d)", phase, os.Getpid())
}
