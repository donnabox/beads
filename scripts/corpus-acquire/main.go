// corpus-acquire performs a Dolt-native, filesystem-level acquisition of a
// point-in-time clone of a named production corpus (my_db | gascity) from
// the shared multi-database Dolt data_dir, plus depth measurement, for the
// downstream replay-with-oracle harness (be-hs42e.5.2) to consume.
//
// It never invokes git, never opens a write-capable (or any) MySQL-protocol
// connection to the shared Dolt server, and never creates a scratch
// database on that server — see be-hs42e.5.1's exit_contract and
// corpusacquire_test.go for the full acceptance criteria this satisfies.
//
// Usage:
//
//	go run ./scripts/corpus-acquire -data-dir=<dir> -db=my_db -dest=<dir>
package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// CorpusClone is the CORPUS_CLONE-shaped output record be-hs42e.5.2 (the
// downstream replay-with-oracle harness) consumes.
type CorpusClone struct {
	SourceDB           string    `json:"source_db"`
	SourceServer       string    `json:"source_server"`
	SnapshotCommitHash string    `json:"snapshot_commit_hash"`
	IssueCount         int       `json:"issue_count"`
	DoltLogCommitCount int       `json:"dolt_log_commit_count"`
	OldestCommitAt     time.Time `json:"oldest_commit_at"`
	NewestCommitAt     time.Time `json:"newest_commit_at"`
	Partial            bool      `json:"partial"`
	PartialReason      string    `json:"partial_reason,omitempty"`
}

func main() {
	dataDir := flag.String("data-dir", "", "Dolt shared server data_dir containing the source database (required)")
	db := flag.String("db", "", "source database name, e.g. my_db or gascity (required)")
	dest := flag.String("dest", "", "destination directory for the acquired clone (default: minted under -out-root)")
	outRoot := flag.String("out-root", "", "parent directory to mint a destination name under, when -dest is not given")
	sourceServer := flag.String("source-server", "127.0.0.1:28231", "address of the source Dolt server, recorded for provenance only — never connected to")
	baselineDepth := flag.Int("baseline-depth", 0, "prior dolt_log commit-count baseline; 0 disables the partial-coverage check")
	flag.Parse()

	if *dataDir == "" || *db == "" {
		fmt.Fprintln(os.Stderr, "usage: corpus-acquire -data-dir=<dir> -db=<name> [-dest=<dir>] [-out-root=<dir>] [-source-server=host:port] [-baseline-depth=N]")
		os.Exit(2)
	}

	destDir := *dest
	if destDir == "" {
		if *outRoot == "" {
			fmt.Fprintln(os.Stderr, "corpus-acquire: one of -dest or -out-root is required")
			os.Exit(2)
		}
		destDir = filepath.Join(*outRoot, GenerateCloneID(*db, time.Now()))
	}
	if err := ValidateMintedName(filepath.Base(destDir)); err != nil {
		fmt.Fprintln(os.Stderr, "corpus-acquire:", err)
		os.Exit(1)
	}

	ctx := context.Background()
	if err := AcquireCorpus(ctx, *dataDir, *db, destDir); err != nil {
		fmt.Fprintln(os.Stderr, "corpus-acquire:", err)
		os.Exit(1)
	}

	result, err := MeasureCorpus(ctx, destDir, *db, *sourceServer, *baselineDepth)
	if err != nil {
		fmt.Fprintln(os.Stderr, "corpus-acquire:", err)
		os.Exit(1)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(result); err != nil {
		fmt.Fprintln(os.Stderr, "corpus-acquire: encoding result:", err)
		os.Exit(1)
	}
}

// AcquireCorpus performs a Dolt-native, filesystem-level acquisition of
// dataDir/dbName into the isolated destDir, via a two-hop dolt-native
// backup (never `dolt clone`, which fails against chunk-journal-format
// repos — see be-hs42e.5.1's design-finding notes):
//
//  1. `dolt --data-dir=<sourceDir> backup sync-url file://<staging>`, run
//     from a neutral cwd so this process never chdir's into the live
//     source dir and never contends for its lock while a dolt sql-server
//     may still be serving it.
//  2. `dolt backup restore file://<staging> <name>`, cwd = destDir's
//     parent, materializing the actual usable working clone at destDir.
//
// Both hops shell out to the local dolt CLI only, with a sanitized child
// environment (sanitizedEnv) — never git, never a network or MySQL-protocol
// address, never the shared server.
func AcquireCorpus(ctx context.Context, dataDir, dbName, destDir string) error {
	sourceDir := filepath.Join(dataDir, dbName)
	if !isDoltRepo(sourceDir) {
		return fmt.Errorf("corpus-acquire: no dolt database %q found under %s", dbName, dataDir)
	}

	stagingParent, err := os.MkdirTemp("", "corpus-acquire-staging-*")
	if err != nil {
		return fmt.Errorf("corpus-acquire: creating staging parent dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(stagingParent) }()
	// dolt creates the staging dir itself; it must not already exist.
	stagingDir := filepath.Join(stagingParent, "backup")

	neutralDir, err := os.MkdirTemp("", "corpus-acquire-cwd-*")
	if err != nil {
		return fmt.Errorf("corpus-acquire: creating neutral cwd: %w", err)
	}
	defer func() { _ = os.RemoveAll(neutralDir) }()

	if _, err := doltRun(ctx, neutralDir, "--data-dir="+sourceDir, "backup", "sync-url", "file://"+stagingDir); err != nil {
		return fmt.Errorf("corpus-acquire: backup sync-url: %w", err)
	}

	destParent := filepath.Dir(destDir)
	if err := os.MkdirAll(destParent, 0o755); err != nil {
		return fmt.Errorf("corpus-acquire: creating destination parent: %w", err)
	}
	destName := filepath.Base(destDir)

	if _, err := doltRun(ctx, destParent, "backup", "restore", "file://"+stagingDir, destName); err != nil {
		return fmt.Errorf("corpus-acquire: backup restore: %w", err)
	}

	return nil
}

// MeasureCorpus re-measures the local clone at cloneDir on every call (never
// cached): its HEAD commit hash, issue count, and dolt_log depth and date
// range, all via read-only SQL queries against the local clone only — never
// the sourceServer, which is recorded on the result purely for provenance
// and is never connected to. If priorBaselineDepth is positive and the
// clone's actual depth falls below it, the result is flagged Partial rather
// than silently reporting narrower coverage as if it were complete.
func MeasureCorpus(ctx context.Context, cloneDir, sourceDB, sourceServer string, priorBaselineDepth int) (CorpusClone, error) {
	cc := CorpusClone{
		SourceDB:     sourceDB,
		SourceServer: sourceServer,
	}

	headRow, err := doltSQLRow(ctx, cloneDir, "SELECT commit_hash FROM dolt_log LIMIT 1")
	if err != nil {
		return cc, fmt.Errorf("corpus-acquire: measuring snapshot commit hash: %w", err)
	}
	cc.SnapshotCommitHash = strings.TrimSpace(headRow[0])

	issueRow, err := doltSQLRow(ctx, cloneDir, "SELECT COUNT(*) FROM issues")
	if err != nil {
		return cc, fmt.Errorf("corpus-acquire: measuring issue count: %w", err)
	}
	issueCount, err := strconv.Atoi(strings.TrimSpace(issueRow[0]))
	if err != nil {
		return cc, fmt.Errorf("corpus-acquire: parsing issue count %q: %w", issueRow[0], err)
	}
	cc.IssueCount = issueCount

	depthRow, err := doltSQLRow(ctx, cloneDir, "SELECT COUNT(*), MIN(UNIX_TIMESTAMP(date)), MAX(UNIX_TIMESTAMP(date)) FROM dolt_log")
	if err != nil {
		return cc, fmt.Errorf("corpus-acquire: measuring commit depth: %w", err)
	}
	if len(depthRow) < 3 {
		return cc, fmt.Errorf("corpus-acquire: unexpected dolt_log depth query result: %v", depthRow)
	}
	commitCount, err := strconv.Atoi(strings.TrimSpace(depthRow[0]))
	if err != nil {
		return cc, fmt.Errorf("corpus-acquire: parsing dolt_log commit count %q: %w", depthRow[0], err)
	}
	cc.DoltLogCommitCount = commitCount

	oldestSec, err := strconv.ParseFloat(strings.TrimSpace(depthRow[1]), 64)
	if err != nil {
		return cc, fmt.Errorf("corpus-acquire: parsing oldest commit timestamp %q: %w", depthRow[1], err)
	}
	newestSec, err := strconv.ParseFloat(strings.TrimSpace(depthRow[2]), 64)
	if err != nil {
		return cc, fmt.Errorf("corpus-acquire: parsing newest commit timestamp %q: %w", depthRow[2], err)
	}
	cc.OldestCommitAt = time.Unix(int64(oldestSec), 0).UTC()
	cc.NewestCommitAt = time.Unix(int64(newestSec), 0).UTC()

	if priorBaselineDepth > 0 && commitCount < priorBaselineDepth {
		cc.Partial = true
		cc.PartialReason = fmt.Sprintf("dolt_log_commit_count=%d is below the supplied baseline depth=%d", commitCount, priorBaselineDepth)
	}

	return cc, nil
}

var mintedNameDangerRE = regexp.MustCompile(`^[0-9a-v]{16,31}$`)

// ValidateMintedName rejects any name this tool would use for a directory
// or branch that matches the shape of a truncated Dolt/Noms content hash:
// 16-31 characters entirely within the base32-style [0-9a-v] alphabet. Real
// Dolt commit hashes are always exactly 32 characters in this alphabet, so
// a 32-character match is deliberately allowed — the danger zone is names
// that merely *look* hash-shaped by coincidence at a shorter length, which
// can confuse downstream ref resolution (NFR5/R2, the gap in upstream
// issueops.ValidateRef tracked on PR #6730). Ordinary English compound
// words can innocently fall in this trap: "productionmirror",
// "releasecandidate", and "integrationtests" are all real 16-character
// examples entirely within [0-9a-v].
func ValidateMintedName(name string) error {
	if mintedNameDangerRE.MatchString(name) {
		return fmt.Errorf("corpus-acquire: name %q matches the dangerous truncated-hash shape ^[0-9a-v]{16,31}$ (NFR5/R2) — choose a name with a hyphen, underscore, uppercase letter, or a length outside 16-31", name)
	}
	return nil
}

// GenerateCloneID mints a destination identifier for dbName that is safe
// under ValidateMintedName by construction: the literal hyphens in
// "corpus-<db>-<timestamp>" fall outside the [0-9a-v] character class, so
// the full string can never match the dangerous shape regardless of dbName
// or when it is called.
func GenerateCloneID(dbName string, now time.Time) string {
	return fmt.Sprintf("corpus-%s-%s", dbName, now.UTC().Format("20060102-150405"))
}

// sanitizedEnv returns base with every environment variable this tool must
// never blindly trust stripped out (mirrors scripts/ci/pr-core.sh's
// pattern): the ambient Dolt server address/port, actor identity, and city
// root, none of which this tool's own filesystem-level operations need, and
// all of which have caused a real incident when leaked into a child
// process (cairn beads-testdb-production-leak: ambient
// BEADS_DOLT_SERVER_PORT fail-open created a scratch database directly on
// the shared production server).
func sanitizedEnv(base []string) []string {
	strip := map[string]bool{
		"BEADS_DOLT_SERVER_PORT": true,
		"BEADS_DOLT_PORT":        true,
		"BEADS_ACTOR":            true,
		"BD_ACTOR":               true,
		"GT_ROOT":                true,
	}
	out := make([]string, 0, len(base))
	for _, kv := range base {
		name := kv
		if i := strings.IndexByte(kv, '='); i >= 0 {
			name = kv[:i]
		}
		if strip[name] {
			continue
		}
		out = append(out, kv)
	}
	return out
}

// isDoltRepo reports whether dir looks like a dolt database directory.
// AcquireCorpus uses this to fail cleanly on a missing source without ever
// creating one — this tool only ever reads.
func isDoltRepo(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, ".dolt"))
	return err == nil && info.IsDir()
}

// doltRun runs the dolt CLI (resolved via PATH at call time, like any other
// exec.Command — this is also what lets a test's PATH-prepended shim
// intercept it) with dir as its working directory and a sanitized child
// environment. It never invokes git and never receives a non-file:// URL or
// server address from any caller in this package.
func doltRun(ctx context.Context, dir string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "dolt", args...)
	cmd.Dir = dir
	cmd.Env = sanitizedEnv(os.Environ())
	out, err := cmd.CombinedOutput()
	if err != nil {
		return out, fmt.Errorf("dolt %s (dir=%s): %w\n%s", strings.Join(args, " "), dir, err, out)
	}
	return out, nil
}

// doltSQLRow runs a single-row, read-only SQL query against the dolt
// database at dir — the local clone only, never a network or server
// connection — and returns its columns as strings.
func doltSQLRow(ctx context.Context, dir, query string) ([]string, error) {
	out, err := doltRun(ctx, dir, "sql", "-q", query, "-r", "csv")
	if err != nil {
		return nil, err
	}
	r := csv.NewReader(bytes.NewReader(out))
	rows, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("corpus-acquire: parsing CSV from query %q: %w", query, err)
	}
	if len(rows) < 2 {
		return nil, fmt.Errorf("corpus-acquire: query %q returned no data row", query)
	}
	return rows[1], nil
}
