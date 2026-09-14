//go:build !unix

package authority

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestInstallationUnsupportedHasNoEffects(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "uncreated", "installation-id")
	t.Setenv("BEADS_INSTALLATION_ID_FILE", path)
	key, err := InstallationKey(context.Background(), root)
	if key != "" || !errors.Is(err, ErrInstallationUnsupported) {
		t.Fatalf("unsupported: %q %v", key, err)
	}
	if _, err := os.Stat(filepath.Dir(path)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("filesystem effect: %v", err)
	}
}
