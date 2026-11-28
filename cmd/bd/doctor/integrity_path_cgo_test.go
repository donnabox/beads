//go:build cgo

package doctor

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/steveyegge/beads/internal/beads"
	"github.com/steveyegge/beads/internal/storage/dolt"
)

func TestCheckRepoFingerprint_UsesTargetRepoOutsideCWD(t *testing.T) {
	port := doctorTestServerPort()
	if port == 0 {
		t.Skip("Dolt test server not available")
	}

	outerRepo := t.TempDir()
	targetRepo := t.TempDir()

	setupGitRepoInDir(t, outerRepo)
	setupGitRepoInDir(t, targetRepo)

	targetRepoID, err := beads.ComputeRepoIDForPath(targetRepo)
	if err != nil {
		t.Fatalf("ComputeRepoIDForPath(targetRepo) failed: %v", err)
	}

	beadsDir := filepath.Join(targetRepo, ".beads")
	// CheckRepoFingerprint looks for the local dolt directory before it opens
	// the store.
	if err := os.MkdirAll(filepath.Join(beadsDir, "dolt"), 0o755); err != nil {
		t.Fatalf("failed to create dolt directory: %v", err)
	}
	dbName := newDoctorTestDatabase(t, beadsDir, port)

	ctx := context.Background()
	store, err := dolt.New(ctx, &dolt.Config{
		Path:       filepath.Join(beadsDir, "dolt"),
		Database:   dbName,
		ServerHost: "127.0.0.1",
		ServerPort: port,
	})
	if err != nil {
		t.Fatalf("failed to open Dolt store: %v", err)
	}
	defer func() { _ = store.Close() }()

	if err := store.SetMetadata(ctx, "repo_id", targetRepoID); err != nil {
		t.Fatalf("failed to set repo_id metadata: %v", err)
	}

	runInDir(t, outerRepo, func() {
		check := CheckRepoFingerprint(targetRepo)

		if check.Status != StatusOK {
			t.Fatalf("Status = %q, want %q (message=%q detail=%q)", check.Status, StatusOK, check.Message, check.Detail)
		}
		if check.Message != "Verified ("+targetRepoID[:8]+")" {
			t.Fatalf("Message = %q, want %q", check.Message, "Verified ("+targetRepoID[:8]+")")
		}
	})
}
