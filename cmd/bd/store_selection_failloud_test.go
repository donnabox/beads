//go:build cgo

package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// writeCorruptMetadata creates a .beads dir whose metadata.json exists but
// cannot be parsed — the state a reader sees when the file is caught
// mid-rewrite (os.WriteFile truncate window) or hit by a transient read
// failure under load.
func writeCorruptMetadata(t *testing.T) string {
	t.Helper()
	beadsDir := filepath.Join(t.TempDir(), ".beads")
	if err := os.MkdirAll(beadsDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(beadsDir, "metadata.json"), []byte(`{"dolt_mode":"serv`), 0o600); err != nil {
		t.Fatalf("write corrupt metadata.json: %v", err)
	}
	return beadsDir
}

// A present-but-unloadable metadata.json must be a hard error, never a
// silent fall-through to the embedded store. In managed server-mode
// deployments the embedded directory is an empty relic, so the silent
// fallback answers every query with an empty result set and exit 0 —
// callers read "no work" where the real store has rows (false-empty).
func TestNewDoltStoreFromConfigCorruptMetadataFailsLoud(t *testing.T) {
	beadsDir := writeCorruptMetadata(t)
	store, err := newDoltStoreFromConfig(context.Background(), beadsDir)
	if err == nil {
		if store != nil {
			_ = store.Close()
		}
		t.Fatal("newDoltStoreFromConfig: want error for corrupt metadata.json, got nil (silent embedded fallback)")
	}
}

func TestNewReadOnlyStoreFromConfigCorruptMetadataFailsLoud(t *testing.T) {
	beadsDir := writeCorruptMetadata(t)
	store, err := newReadOnlyStoreFromConfig(context.Background(), beadsDir)
	if err == nil {
		if store != nil {
			_ = store.Close()
		}
		t.Fatal("newReadOnlyStoreFromConfig: want error for corrupt metadata.json, got nil (silent embedded fallback)")
	}
}

// loadServerModeFromBeadsDir feeds the serverMode globals that the primary
// store-init path consults; a swallowed load failure leaves serverMode=false
// and routes data commands to the embedded store. The error must surface.
func TestLoadServerModeFromBeadsDirCorruptMetadataReturnsError(t *testing.T) {
	beadsDir := writeCorruptMetadata(t)
	if err := loadServerModeFromBeadsDir(beadsDir); err == nil {
		t.Fatal("loadServerModeFromBeadsDir: want error for corrupt metadata.json, got nil")
	}
}

// End-to-end contract for a corrupt metadata.json, exercised through the
// real binary: store-free information remains available, but commands that
// inspect or initialize storage refuse an unknown storage mode. In particular,
// doctor has direct store paths, and init must not infer a storage mode from
// unparseable metadata. Neither is an automatic corrupt-metadata repair.
func TestCorruptMetadataRefusesUnsafeCommandsWithoutMutation(t *testing.T) {
	if os.Getenv("BEADS_TEST_EMBEDDED_DOLT") != "1" {
		t.Skip("set BEADS_TEST_EMBEDDED_DOLT=1 to run embedded dolt tests")
	}
	t.Parallel()

	bd := buildEmbeddedBD(t)
	beadsDir := writeCorruptMetadata(t)
	dir := filepath.Dir(beadsDir)
	metadataPath := filepath.Join(beadsDir, "metadata.json")
	metadata, err := os.ReadFile(metadataPath)
	if err != nil {
		t.Fatal(err)
	}

	run := func(args ...string) (string, string, error) {
		cmd := exec.CommandContext(t.Context(), bd, args...)
		cmd.Dir = dir
		cmd.Env = bdEnv(dir)
		stdout, stderr, err := runCommandBuffers(t, cmd)
		return stdout.String(), stderr.String(), err
	}
	assertPreserved := func() {
		t.Helper()
		after, err := os.ReadFile(metadataPath)
		if err != nil || !bytes.Equal(after, metadata) {
			t.Fatalf("corrupt metadata changed: %q, %v", after, err)
		}
		entries, err := os.ReadDir(beadsDir)
		if err != nil || len(entries) != 1 || entries[0].Name() != "metadata.json" || !entries[0].Type().IsRegular() {
			t.Fatalf("unexpected entries while storage mode was unknown: %v, %v", entries, err)
		}
	}

	// These commands can report useful information without selecting a store.
	for _, args := range [][]string{{"version"}, {"help"}} {
		stdout, stderr, err := run(args...)
		if err != nil {
			t.Errorf("bd %s with corrupt metadata.json: want success, got %v\nstdout: %s\nstderr: %s", strings.Join(args, " "), err, stdout, stderr)
		}
		for _, line := range strings.Split(stderr, "\n") {
			if strings.HasPrefix(line, "Error:") {
				t.Errorf("successful bd %s emitted an error: %s", strings.Join(args, " "), line)
			}
		}
		assertPreserved()
	}

	// Refuse diagnostics with direct store paths, initialization and data reads.
	// A successful empty result or inferred embedded initialization is unsafe.
	for _, args := range [][]string{{"doctor"}, {"init", "--prefix", "cm"}, {"init", "--reinit-local", "--prefix", "cm"}, {"list", "--json"}} {
		stdout, stderr, err := run(args...)
		if err == nil {
			t.Errorf("bd %s with corrupt metadata: want refusal, got success:\n%s", strings.Join(args, " "), stdout)
		}
		// The pre-run warning also appears on successful commands. Require
		// the command's actual parse refusal, not an unrelated no-DB exit.
		parseRefusal := false
		for _, line := range strings.Split(stderr, "\n") {
			if strings.HasPrefix(line, "Error:") && strings.Contains(line, "parsing config") {
				parseRefusal = true
			}
		}
		if !parseRefusal || stdout != "" || strings.Contains(stderr, "no beads database found") {
			t.Errorf("bd %s did not refuse metadata parsing specifically:\nstdout: %s\nstderr: %s", strings.Join(args, " "), stdout, stderr)
		}
		// Preserve the warning's file identification and no-storage guarantee
		// separately from the command error above.
		for _, want := range []string{"metadata.json", "parsing config", "no storage database was opened or modified"} {
			if !strings.Contains(stderr, want) {
				t.Errorf("bd %s warning missing %q:\n%s", strings.Join(args, " "), want, stderr)
			}
		}
		assertPreserved()
	}

	// Environment sanity control: healthy initialization and reads work in a
	// separate clean workspace, without repairing the negative input above.
	healthyDir, _, _ := bdInit(t, bd, "--prefix", "cm")
	cmd := exec.CommandContext(t.Context(), bd, "list", "--json")
	cmd.Dir = healthyDir
	cmd.Env = bdEnv(healthyDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("bd list in separately initialized workspace: %v\n%s", err, out)
	}
}

// Absent metadata.json stays a legitimate fresh-repo default: no error.
func TestLoadServerModeFromBeadsDirAbsentMetadataIsFine(t *testing.T) {
	beadsDir := filepath.Join(t.TempDir(), ".beads")
	if err := os.MkdirAll(beadsDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := loadServerModeFromBeadsDir(beadsDir); err != nil {
		t.Fatalf("loadServerModeFromBeadsDir: want nil for absent metadata.json, got %v", err)
	}
}
