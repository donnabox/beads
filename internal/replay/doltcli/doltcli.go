// Package doltcli is the only place the replay harness starts dolt. Every
// query goes through one runner with one environment deny-list, refuses to
// query a directory a dolt sql-server is serving, and treats stdout as data
// and stderr as diagnostics only.
package doltcli

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// deniedEnvVars are ambient environment variables that must never reach a
// dolt or bd process the harness starts: letting one through risks silently
// routing a replay at a shared or production store, or attaching the wrong
// actor identity to a mutation the harness makes on its own behalf.
var deniedEnvVars = map[string]bool{
	"BEADS_DOLT_SERVER_PORT":      true,
	"BEADS_DOLT_PORT":             true,
	"BEADS_ACTOR":                 true,
	"BD_ACTOR":                    true,
	"GT_ROOT":                     true,
	"BEADS_DIR":                   true,
	"BEADS_HOLDER_TOKEN":          true,
	"GC_BEADS_SCOPE_ROOT":         true,
	"BEADS_DOLT_AUTO_START":       true,
	"BEADS_DOLT_SYNC_CLI_REMOTES": true,
	"BEADS_BACKUP_ENABLED":        true,
}

// SanitizedEnv returns base with every denied ambient variable removed, for
// use as the environment of any dolt or bd process the harness starts.
func SanitizedEnv(base []string) []string {
	out := make([]string, 0, len(base))
	for _, kv := range base {
		name, _, _ := strings.Cut(kv, "=")
		if deniedEnvVars[name] {
			continue
		}
		out = append(out, kv)
	}
	return out
}

// Path returns the dolt binary the harness runs, looked up on PATH at call
// time so a test can put a stand-in ahead of it. It is the one place dolt is
// located: a caller that only needs to know whether dolt is available asks the
// same question the runner does.
func Path() (string, error) {
	return exec.LookPath("dolt")
}

// command prepares a dolt invocation in dir, started with the sanitized
// environment.
func command(ctx context.Context, dir string, args ...string) (*exec.Cmd, error) {
	doltPath, err := Path()
	if err != nil {
		return nil, fmt.Errorf("doltcli: %w", err)
	}
	cmd := exec.CommandContext(ctx, doltPath, args...)
	cmd.Dir = dir
	cmd.Env = SanitizedEnv(os.Environ())
	return cmd, nil
}

// Query runs one read-only SQL statement in the dolt database at dir and
// returns the CSV header and the data rows. It issues only the statement it
// is given, through `dolt sql -q ... -r csv`; dir must be a local clone, never
// a directory a dolt sql-server is serving. Only stdout is parsed: stderr is
// reported in the error and never mistaken for data.
func Query(ctx context.Context, dir, sql string) (header []string, rows [][]string, err error) {
	if served := servedDataDir(dir); served != "" {
		return nil, nil, fmt.Errorf("doltcli: %s is under %s, which a dolt sql-server is serving; querying it would route through that server instead of the local clone", dir, served)
	}
	cmd, err := command(ctx, dir, "sql", "-q", sql, "-r", "csv")
	if err != nil {
		return nil, nil, err
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, nil, fmt.Errorf("dolt sql: %w: %s", err, stderr.String())
	}
	all, err := csv.NewReader(bytes.NewReader(out)).ReadAll()
	if err != nil {
		return nil, nil, fmt.Errorf("parsing dolt sql csv output: %w", err)
	}
	if len(all) == 0 {
		return nil, nil, nil
	}
	return all[0], all[1:], nil
}

// Run runs an arbitrary dolt subcommand in dir and returns its combined
// output. It is for fixtures and other setup that must change a database; it
// does not apply the served-directory guard.
func Run(ctx context.Context, dir string, args ...string) ([]byte, error) {
	cmd, err := command(ctx, dir, args...)
	if err != nil {
		return nil, err
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return out, fmt.Errorf("dolt %s (dir=%s): %w\n%s", strings.Join(args, " "), dir, err, out)
	}
	return out, nil
}

// RowMap builds a column-name to value map from one CSV row and its header.
func RowMap(header, row []string) map[string]string {
	m := make(map[string]string, len(header))
	for i, h := range header {
		if i < len(row) {
			m[h] = row[i]
		}
	}
	return m
}

// SQLQuote wraps a value in single quotes for embedding in a dolt sql -q
// statement, doubling any quote inside it. dolt sql -q has no bind-parameter
// API, so callers validate refs and ids before they get here.
func SQLQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

// servedDataDir returns dir or its nearest ancestor that a dolt sql-server is
// serving, identified by the .dolt/sql-server.info record the server writes at
// startup, or "" when there is none. The dolt CLI routes commands for any
// database under such a directory through that server.
func servedDataDir(dir string) string {
	dir = filepath.Clean(dir)
	for {
		if _, err := os.Stat(filepath.Join(dir, ".dolt", "sql-server.info")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
