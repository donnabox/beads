package driver

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/steveyegge/beads/internal/replay/doltcli"
)

// execBd runs the bd binary at bdPath in dir with args, under the sanitized
// environment, and returns what it printed, standard output and standard error
// together. bdPath is the caller's choice and is never looked up on PATH: the
// code under test is the binary the caller names. A relative path is taken
// relative to the caller's directory, because bd runs with dir as its working
// directory, which would otherwise change what the path means.
func execBd(ctx context.Context, bdPath, dir string, args ...string) ([]byte, error) {
	absBd, err := filepath.Abs(bdPath)
	if err != nil {
		return nil, fmt.Errorf("resolving %s: %w", bdPath, err)
	}
	cmd := exec.CommandContext(ctx, absBd, args...)
	cmd.Dir = dir
	cmd.Env = doltcli.SanitizedEnv(os.Environ())
	return cmd.CombinedOutput()
}
