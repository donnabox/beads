//go:build cgo

package graphops

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"
	"time"

	"github.com/dolthub/vitess/go/vt/sqlparser"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage/embeddeddolt"
)

//go:embed testdata/fixture.sql
var engineFixtureSQL string

// This crosses a different boundary from managed-server qualification: the
// supported embedded driver, in two nonoverlapping OS processes. A fixture,
// never an authoritative graph Scope or production graph schema.
func TestGraphReadEnginePersistenceAcrossProcesses(t *testing.T) {
	if os.Getenv("GRAPH_READ_FIXTURE_PHASE") != "" || os.Getenv("GRAPH_READ_FIXTURE_DATA") != "" {
		t.Fatal("parent phase/data environment must be empty")
	}
	if os.Getenv("BEADS_TEST_GRAPH_READ_FIXTURE") != "1" {
		t.Skip("opt-in embedded graph-row persistence fixture")
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		t.Logf("effective build info:\n%s", info.String())
	} else {
		t.Fatal("build info unavailable")
	}
	root := t.TempDir()
	home := filepath.Join(root, "home")
	data := filepath.Join(root, "data")
	for _, dir := range []string{home, data} {
		if err := os.Mkdir(dir, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(home, ".gitconfig"), []byte("[user]\n name = graph-row-fixture\n email = fixture@example.invalid\n[core]\n hooksPath = /dev/null\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	var seedDigest string
	for _, phase := range []string{"seed", "read"} {
		ctx, cancel := context.WithTimeout(t.Context(), 4*time.Minute)
		cmd := exec.CommandContext(ctx, executable, "-test.run=^TestGraphReadEnginePhase$", "-test.v", "-test.timeout=3m")
		cmd.Dir = root
		cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "HOME=" + home, "TMPDIR=" + root, "GIT_CONFIG_GLOBAL=" + filepath.Join(home, ".gitconfig"), "GIT_CONFIG_SYSTEM=/dev/null", "DOLT_DISABLE_EVENT_FLUSH=1", "DOLT_METRICS_DISABLED=1", "GRAPH_READ_FIXTURE_PHASE=" + phase, "GRAPH_READ_FIXTURE_DATA=" + data}
		var output boundedFixtureOutput
		cmd.Stdout = &output
		cmd.Stderr = &output
		cmd.WaitDelay = 2 * time.Second
		err := cmd.Run()
		cancel()
		pid := 0
		if cmd.Process != nil {
			pid = cmd.Process.Pid
		}
		t.Logf("phase=%s pid=%d exit=%v output=%s", phase, pid, err, output.String())
		if err != nil || output.overflow {
			t.Fatalf("embedded fixture %s failed: %v overflow=%t", phase, err, output.overflow)
		}
		if !strings.Contains(output.String(), "GRAPH_READ_FIXTURE_OK "+phase) {
			t.Fatal("missing completed phase evidence")
		}
		var digest string
		for _, line := range strings.Split(output.String(), "\n") {
			if strings.HasPrefix(line, "GRAPH_READ_TABLE_DIGEST ") {
				if digest != "" {
					t.Fatal("duplicate digest receipt")
				}
				digest = strings.TrimPrefix(line, "GRAPH_READ_TABLE_DIGEST ")
			}
		}
		decoded, err := hex.DecodeString(digest)
		if err != nil || len(decoded) != sha256.Size {
			t.Fatal("missing or invalid phase table digest")
		}
		if phase == "seed" {
			seedDigest = digest
		} else if digest != seedDigest {
			t.Fatal("persisted table digest changed across processes")
		}

	}
}

type boundedFixtureOutput struct {
	buffer   bytes.Buffer
	overflow bool
}

func (b *boundedFixtureOutput) Write(p []byte) (int, error) {
	const limit = 1 << 20
	n := len(p)
	remaining := limit - b.buffer.Len()
	if len(p) > remaining {
		p = p[:remaining]
		b.overflow = true
	}
	_, err := b.buffer.Write(p)
	return n, err
}

func (b *boundedFixtureOutput) String() string { return b.buffer.String() }

func TestFixtureOutputCopyBound(t *testing.T) {
	var output boundedFixtureOutput
	// Hide strings.Reader.WriteTo: os/exec's pipe copy may select a promoted
	// Buffer.ReadFrom instead of Write if Buffer is anonymously embedded.
	source := io.LimitReader(strings.NewReader(strings.Repeat("x", (1<<20)+1)), (1<<20)+1)
	n, err := io.Copy(&output, source)
	if err != nil || n != (1<<20)+1 || !output.overflow || output.buffer.Len() != 1<<20 {
		t.Fatalf("copy bound: read=%d kept=%d overflow=%t error=%v", n, output.buffer.Len(), output.overflow, err)
	}
}

func TestGraphReadEnginePhase(t *testing.T) {
	phase := os.Getenv("GRAPH_READ_FIXTURE_PHASE")
	if phase == "" {
		t.Skip("parent-owned subprocess phase")
	}
	if phase != "seed" && phase != "read" {
		t.Fatal("unknown phase")
	}
	data := os.Getenv("GRAPH_READ_FIXTURE_DATA")
	if !filepath.IsAbs(data) {
		t.Fatal("explicit absolute fixture data path required")
	}
	ctx, cancel := context.WithTimeout(t.Context(), 170*time.Second)
	defer cancel()
	if phase == "seed" {
		boot, cleanup, err := embeddeddolt.OpenSQL(ctx, data, "", "")
		if err != nil {
			t.Fatal(err)
		}
		_, createErr := boot.ExecContext(ctx, "CREATE DATABASE graph_read_fixture")
		closeErr := cleanup()
		if err := errors.Join(createErr, closeErr); err != nil {
			t.Fatal(err)
		}
	} else if _, err := os.Stat(filepath.Join(data, "graph_read_fixture")); err != nil {
		t.Fatalf("read phase refuses missing database: %v", err)
	}
	db, cleanup, err := embeddeddolt.OpenSQL(ctx, data, "graph_read_fixture", "main")
	if err != nil {
		t.Fatal(err)
	}
	closed := false
	defer func() {
		if !closed {
			if err := cleanup(); err != nil {
				t.Error(err)
			}
		}
	}()
	if phase == "seed" {
		seedEngineFixture(t, ctx, db, data)
	}
	before := fixtureTablesDigest(t, ctx, db)
	verifyEngineFixture(t, ctx, db, data)
	verifyEngineNegativeControls(t, ctx, db)
	verifyEngineByteControls(t, ctx, db)
	verifyEngineOwnedOverflow(t, ctx, db)
	verifyEngineAllocationControls(t, ctx, db)
	verifyAllocationSnapshot(t, ctx, db)
	logAllocationPlans(t, ctx, db)
	after := fixtureTablesDigest(t, ctx, db)
	if before != after {
		t.Fatal("read/rolled-back controls changed persisted tables")
	}
	fmt.Printf("GRAPH_READ_TABLE_DIGEST %s\n", after)
	closed = true
	if err := cleanup(); err != nil {
		t.Fatal(err)
	}
	// No engine remains at success emission; diagnostic failures stay failures.
	if !t.Failed() {
		fmt.Println("GRAPH_READ_FIXTURE_OK", phase)
	}
}

type fixtureRecord struct{ Path, Revision, Properties, TypeURL, Source, Target, SourceURI, SourcePin, TargetPin string }
type fixtureReceipt struct {
	Beads, Links []fixtureRecord
	SchemaSHA256 string
}

// Both *sql.DB and an explicitly bound *sql.Conn implement this fixture seam.
// It grants no public graph role and changes no production storage interface.
type fixtureDB interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	BeginTx(context.Context, *sql.TxOptions) (*sql.Tx, error)
}

func TestFixtureSchemaStatementParity(t *testing.T) {
	parts, err := sqlparser.SplitStatementToPieces(engineFixtureSQL)
	if err != nil || len(parts) != 5 {
		t.Fatalf("fixture statements: %d, %v", len(parts), err)
	}
	// The tokenizer only removes delimiters; all original DDL/comment bytes
	// and statement order must survive. This is the embedded seeder's schema.
	if strings.TrimSpace(strings.Join(parts, ";"))+";" != strings.TrimSpace(engineFixtureSQL) {
		t.Fatal("schema splitting changed source bytes")
	}
	const quoted = "SELECT ';'; SELECT 'second;value' /* ; */;"
	pieces, err := sqlparser.SplitStatementToPieces(quoted)
	if err != nil || len(pieces) != 2 || strings.Join(pieces, ";")+";" != quoted {
		t.Fatalf("quoted/comment semicolon split: %q %v", pieces, err)
	}
}

func seedEngineFixture(t *testing.T, ctx context.Context, db fixtureDB, data string) {
	t.Helper()
	// The same source-defined schema runs over embedded and managed MySQL;
	// the managed server deliberately refuses multi-statement requests.
	statements, err := sqlparser.SplitStatementToPieces(engineFixtureSQL)
	if err != nil {
		t.Fatal(err)
	}
	for _, statement := range statements {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			t.Fatalf("create fixture projection schema: %v", err)
		}
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, "INSERT INTO graph_scope VALUES (1, ?, ?, 1, '2026-09-17 00:00:00')", fixtureScope, strings.Repeat("a", 32)); err != nil {
		t.Fatalf("insert fixture Scope: %v", err)
	}
	owns, err := graph.NewOwnedLinkDecl(relationType, "Explains", 2)
	if err != nil {
		t.Fatal(err)
	}
	for _, descriptor := range []graph.TypeDescriptor{memoryDescriptor(t, owns), relationDescriptor(t)} {
		if _, err := tx.ExecContext(ctx, "INSERT INTO graph_type_descriptors VALUES (?, ?, ?, 1, '2026-09-17 00:00:00', ?, 1)", descriptor.ID(), descriptor.CanonicalJSON(), descriptor.Fingerprint(), strings.Repeat("a", 32)); err != nil {
			t.Fatalf("insert fixture descriptor %s: %v", descriptor.ID(), err)
		}
	}
	receipt := fixtureReceipt{}
	for _, input := range []struct{ path, text string }{{"beads/plan", "Launch Monday"}, {"beads/decision", "Moved launch after readiness review"}, {"beads/finding", "Friday is not ready"}} {
		raw, err := json.Marshal(map[string]string{"text": input.text})
		if err != nil {
			t.Fatal(err)
		}
		properties, err := graph.NewProperties(raw)
		if err != nil {
			t.Fatal(err)
		}
		revision := graph.MintRevision().String()
		if _, err := tx.ExecContext(ctx, "INSERT INTO graph_beads (path, type_url, revision, properties, last_authority_id, last_epoch, created_at, updated_at) VALUES (?, ?, ?, ?, ?, 1, '2026-09-17 00:00:00', '2026-09-17 00:00:00')", input.path, memoryType, revision, properties.Bytes(), strings.Repeat("a", 32)); err != nil {
			t.Fatalf("insert fixture Bead %s: %v", input.path, err)
		}
		seedFixtureAllocation(t, ctx, tx, input.path, graph.KindBead)
		receipt.Beads = append(receipt.Beads, fixtureRecord{Path: input.path, Revision: revision, Properties: properties.String(), TypeURL: memoryType})
	}
	for _, edge := range []struct{ path, source, target string }{{"links/plan-decision", "beads/plan", "beads/decision"}, {"links/decision-finding", "beads/decision", "beads/finding"}} {
		revision := graph.MintRevision().String()
		if _, err := tx.ExecContext(ctx, "INSERT INTO graph_links (path, type_url, revision, properties, source_kind, source_path, target_kind, target_path, last_authority_id, last_epoch, created_at, updated_at) VALUES (?, ?, ?, ?, 'in', ?, 'in', ?, ?, 1, '2026-09-17 00:00:00', '2026-09-17 00:00:00')", edge.path, relationType, revision, []byte("{}"), edge.source, edge.target, strings.Repeat("a", 32)); err != nil {
			t.Fatalf("insert fixture Link %s: %v", edge.path, err)
		}
		seedFixtureAllocation(t, ctx, tx, edge.path, graph.KindLink)
		receipt.Links = append(receipt.Links, fixtureRecord{Path: edge.path, Revision: revision, Properties: "{}", TypeURL: relationType, Source: edge.source, Target: edge.target})
	}
	// Exercise the external endpoint columns and opaque pins against actual rows.
	ext := fixtureRecord{Path: "links/external-plan", Revision: graph.MintRevision().String(), Properties: "{}", TypeURL: relationType, SourceURI: "https://external.example/memory/42", SourcePin: "source opaque pin", Target: "beads/plan", TargetPin: "target opaque pin"}
	if _, err := tx.ExecContext(ctx, "INSERT INTO graph_links (path, type_url, revision, properties, source_kind, source_url, source_pin, target_kind, target_path, target_pin, last_authority_id, last_epoch, created_at, updated_at) VALUES (?, ?, ?, ?, 'ext', ?, ?, 'in', ?, ?, ?, 1, '2026-09-17 00:00:00', '2026-09-17 00:00:00')", ext.Path, ext.TypeURL, ext.Revision, []byte(ext.Properties), ext.SourceURI, ext.SourcePin, ext.Target, ext.TargetPin, strings.Repeat("a", 32)); err != nil {
		t.Fatalf("insert external fixture Link: %v", err)
	}
	seedFixtureAllocation(t, ctx, tx, ext.Path, graph.KindLink)
	receipt.Links = append(receipt.Links, ext)
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256([]byte(engineFixtureSQL))
	receipt.SchemaSHA256 = hex.EncodeToString(sum[:])
	raw, err := json.Marshal(receipt)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(data, "fixture-receipt.json"), raw, 0o600); err != nil {
		t.Fatal(err)
	}
}

func verifyEngineFixture(t *testing.T, ctx context.Context, db fixtureDB, data string) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(data, "fixture-receipt.json"))
	if err != nil {
		t.Fatal(err)
	}
	var receipt fixtureReceipt
	if err := json.Unmarshal(raw, &receipt); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256([]byte(engineFixtureSQL))
	if receipt.SchemaSHA256 != hex.EncodeToString(sum[:]) {
		t.Fatal("fixture schema receipt mismatch")
	}
	if len(receipt.Beads) != 3 || len(receipt.Links) != 3 {
		t.Fatal("incomplete seed receipt")
	}
	// Embedded driver/v2 ignores ReadOnly; managed MySQL sends READ ONLY.
	// Before/after table digests check the no-mutation invariant on both legs.
	tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()
	scope := scopeFromFixtureTx(t, ctx, tx)
	for i, expected := range receipt.Beads {
		record, err := readBeadInTx(ctx, tx, scope, expected.Path, fixtureLimits)
		if err != nil {
			t.Fatalf("read fixture Bead %s: %v", expected.Path, err)
		}
		expectedOwned := 1
		if i == 2 {
			expectedOwned = 0
		}
		if _, present := record.Bead.Attribution(); present {
			t.Fatal("SQL NULL Bead attribution became present")
		}
		if len(record.OwnedLinks) != 1 || record.OwnedLinks[0].TypeURL != relationType || len(record.OwnedLinks[0].Links) != expectedOwned {
			t.Fatalf("wrong owned expansion for %s", expected.Path)
		}
		if record.Bead.Revision().String() != expected.Revision || record.Bead.Properties().String() != expected.Properties || record.Bead.TypeURL() != expected.TypeURL {
			t.Fatalf("changed bead %s", expected.Path)
		}
	}
	path := "beads/plan"
	for i, expected := range receipt.Links[:2] {
		links, err := readIncidentLinksInTx(ctx, tx, scope, path, graph.DirectionOut, fixtureLimits)
		if err != nil {
			t.Fatalf("read outbound fixture Links for %s: %v", path, err)
		}
		if len(links) != 1 || links[0].Path() != expected.Path || links[0].Revision().String() != expected.Revision {
			t.Fatalf("changed edge from %s", path)
		}
		exact, err := readLinkInTx(ctx, tx, scope, links[0].Path(), fixtureLimits)
		if err != nil {
			t.Fatalf("read exact fixture Link %s: %v", links[0].Path(), err)
		}
		if _, present := exact.Attribution(); present {
			t.Fatal("SQL NULL Link attribution became present")
		}
		if exact.Source().Path() != expected.Source || exact.Target().Path() != expected.Target || exact.TypeURL() != expected.TypeURL || exact.Properties().String() != expected.Properties || exact.Revision().String() != expected.Revision {
			t.Fatalf("changed exact Link %s", expected.Path)
		}
		path = exact.Target().Path()
		if path != receipt.Beads[i+1].Path {
			t.Fatalf("wrong persisted target %s", path)
		}
	}
	empty, err := readIncidentLinksInTx(ctx, tx, scope, "beads/finding", graph.DirectionOut, fixtureLimits)
	if err != nil || len(empty) != 0 {
		t.Fatalf("empty incident anchor: %v %v", empty, err)
	}
	incoming, err := readIncidentLinksInTx(ctx, tx, scope, "beads/decision", graph.DirectionIn, fixtureLimits)
	if err != nil || len(incoming) != 1 || incoming[0].Path() != "links/plan-decision" {
		t.Fatalf("actual inbound traversal: %v %v", incoming, err)
	}
	external := receipt.Links[2]
	checkExternal := func(link graph.Link) {
		if link.Path() != external.Path || link.Source().InScope() || link.Source().URI() != external.SourceURI || link.Source().Pin() != external.SourcePin || !link.Target().InScope() || link.Target().Path() != external.Target || link.Target().Pin() != external.TargetPin || link.Revision().String() != external.Revision || link.TypeURL() != external.TypeURL || link.Properties().String() != external.Properties {
			t.Fatalf("external/pinned endpoint row changed: %s", link.Path())
		}
	}
	link, err := readLinkInTx(ctx, tx, scope, external.Path, fixtureLimits)
	if err != nil {
		t.Fatal(err)
	}
	checkExternal(link)
	incoming, err = readIncidentLinksInTx(ctx, tx, scope, "beads/plan", graph.DirectionIn, fixtureLimits)
	if err != nil || len(incoming) != 1 {
		t.Fatalf("external inbound row: %v", err)
	}
	checkExternal(incoming[0])
	var absence *exactAbsence
	if _, err := readBeadInTx(ctx, tx, scope, "beads/absent", fixtureLimits); !errors.As(err, &absence) || absence.state != "" {
		t.Fatalf("absence=%v", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	verifyEnginePageControls(t, ctx, db)
	fmt.Printf("GRAPH_READ_RECORDS %s\n", raw)
}

func verifyEngineNegativeControls(t *testing.T, ctx context.Context, db fixtureDB) {
	t.Helper()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()
	scope := scopeFromFixtureTx(t, ctx, tx)
	// Both directions of each attribution pair must still be refused by SQL.
	for _, query := range []string{
		"UPDATE graph_beads SET attribution_principal = 'fixture' WHERE path = 'beads/plan'",
		"UPDATE graph_beads SET attribution_status = 'claimed' WHERE path = 'beads/plan'",
		"UPDATE graph_links SET attribution_principal = 'fixture' WHERE path = 'links/plan-decision'",
		"UPDATE graph_links SET attribution_status = 'claimed' WHERE path = 'links/plan-decision'",
	} {
		if _, err := tx.ExecContext(ctx, query); err == nil || !strings.Contains(err.Error(), "Check constraint") || !strings.Contains(err.Error(), "violated") {
			t.Fatalf("partial attribution must violate CHECK (%s): %v", query, err)
		}
	}
	// The equivalent CHECK must also admit both-present attribution.
	for _, query := range []string{
		"UPDATE graph_beads SET attribution_principal = 'fixture', attribution_status = 'claimed' WHERE path = 'beads/plan'",
		"UPDATE graph_links SET attribution_principal = 'fixture', attribution_status = 'claimed' WHERE path = 'links/plan-decision'",
	} {
		result, err := tx.ExecContext(ctx, query)
		if err != nil {
			t.Fatalf("paired attribution must pass CHECK: %v", err)
		}
		if n, err := result.RowsAffected(); err != nil || n != 1 {
			t.Fatalf("paired attribution affected %d rows: %v", n, err)
		}
	}
	for _, status := range []string{"claimed", "unknown"} {
		for _, query := range []string{"UPDATE graph_beads SET attribution_status = ? WHERE path = 'beads/plan'", "UPDATE graph_links SET attribution_status = ? WHERE path = 'links/plan-decision'"} {
			if _, err := tx.ExecContext(ctx, query, status); err != nil {
				t.Fatal(err)
			}
		}
		bead, err := readBeadInTx(ctx, tx, scope, "beads/plan", fixtureLimits)
		if err != nil {
			t.Fatal(err)
		}
		attribution, present := bead.Bead.Attribution()
		if !present || attribution.Principal() != "fixture" || string(attribution.Status()) != status {
			t.Fatal("Bead attribution CAST changed value")
		}
		link, err := readLinkInTx(ctx, tx, scope, "links/plan-decision", fixtureLimits)
		if err != nil {
			t.Fatal(err)
		}
		attribution, present = link.Attribution()
		if !present || attribution.Principal() != "fixture" || string(attribution.Status()) != status {
			t.Fatal("Link attribution CAST changed value")
		}
	}
	if _, err := tx.ExecContext(ctx, "UPDATE graph_beads SET properties = ? WHERE path = ?", []byte(`{ "text": "not canonical" }`), "beads/plan"); err != nil {
		t.Fatal(err)
	}
	if _, err := readBeadInTx(ctx, tx, scope, "beads/plan", fixtureLimits); !errors.Is(err, errCorrupt) {
		t.Fatalf("noncanonical persisted row: %v", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	// Embedded driver/v2 ignores ReadOnly; managed MySQL sends READ ONLY.
	// The before/after table digest is the mutation oracle on both legs.
	read, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = read.Rollback() }()
	scope = scopeFromFixtureTx(t, ctx, read)
	limits := fixtureLimits
	limits.rows = 1
	if _, err := readIncidentLinksInTx(ctx, read, scope, "beads/decision", graph.DirectionBoth, limits); !errors.Is(err, errBudget) {
		t.Fatalf("actual incident row bound: %v", err)
	}
	limits = fixtureLimits
	limits.valueBytes = 64
	limits.bytes = 64
	q := &preconditionQueryHook{tx: read}
	if _, err := readLinkInTx(ctx, q, scope, "links/plan-decision", limits); !errors.Is(err, errBudget) || q.calls != 1 {
		t.Fatalf("actual aggregate byte bound: calls=%d err=%v", q.calls, err)
	}
	if err := read.Rollback(); err != nil {
		t.Fatal(err)
	}
}

func fixtureTablesDigest(t *testing.T, ctx context.Context, db fixtureDB) string {
	t.Helper()
	hash := sha256.New()
	for _, entry := range []struct{ table, columns string }{
		{"graph_scope", "id, scope_url, authority_id, epoch, minted_at"},
		{"graph_allocations", "path, CAST(resource_kind AS CHAR), birth_seq, birth_authority_id, birth_authority_epoch, CAST(state AS CHAR), tombstone_seq, last_authority_id, last_authority_epoch"},
		{"graph_type_descriptors", "url, descriptor, fingerprint, installed_seq, installed_at, last_authority_id, last_epoch"},
		{"graph_beads", "path, type_url, revision, attribution_principal, CAST(attribution_status AS CHAR), properties, last_authority_id, last_epoch, created_at, updated_at"},
		{"graph_links", "path, type_url, revision, attribution_principal, CAST(attribution_status AS CHAR), properties, source_kind, source_path, source_url, source_pin, target_kind, target_path, target_url, target_pin, last_authority_id, last_epoch, created_at, updated_at"},
	} {
		// Closed complete column inventory. Driver/v2 panics scanning a NULL ENUM;
		// the CAST preserves its SQL NULL/value without changing stored columns.
		table := entry.table
		rows, err := db.QueryContext(ctx, "SELECT "+entry.columns+" FROM "+table+" ORDER BY 1")
		if err != nil {
			t.Fatal(err)
		}
		columns, err := rows.Columns()
		if err != nil {
			_ = rows.Close()
			t.Fatal(err)
		}
		_, _ = fmt.Fprintf(hash, "table:%s columns:%v\n", table, columns)
		for rows.Next() {
			values := make([]sql.RawBytes, len(columns))
			dest := make([]any, len(columns))
			for i := range values {
				dest[i] = &values[i]
			}
			if err := rows.Scan(dest...); err != nil {
				_ = rows.Close()
				t.Fatal(err)
			}
			for _, value := range values {
				if value == nil {
					_, _ = hash.Write([]byte("NULL;"))
				} else {
					_, _ = fmt.Fprintf(hash, "%d:", len(value))
					_, _ = hash.Write(value)
					_, _ = hash.Write([]byte(";"))
				}
			}
			_, _ = hash.Write([]byte("\n"))
		}
		if err := errors.Join(rows.Err(), rows.Close()); err != nil {
			t.Fatal(err)
		}
	}
	return hex.EncodeToString(hash.Sum(nil))
}

// Fixture row only: this is the future Scope assertion's read slot, not a witness
// or authority check. Reading it inside each transaction binds classification.
func scopeFromFixtureTx(t *testing.T, ctx context.Context, tx *sql.Tx) string {
	t.Helper()
	var scope string
	if err := tx.QueryRowContext(ctx, "SELECT scope_url FROM graph_scope WHERE id = 1").Scan(&scope); err != nil {
		t.Fatal(err)
	}
	if scope != fixtureScope {
		t.Fatal("fixture Scope changed")
	}
	return scope
}

func verifyEngineOwnedOverflow(t *testing.T, ctx context.Context, db fixtureDB) {
	t.Helper()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()
	scope := scopeFromFixtureTx(t, ctx, tx)
	for _, path := range []string{"links/overflow-a", "links/overflow-b"} {
		_, err := tx.ExecContext(ctx, "INSERT INTO graph_links (path, type_url, revision, properties, source_kind, source_path, target_kind, target_path, last_authority_id, last_epoch, created_at, updated_at) VALUES (?, ?, ?, ?, 'in', 'beads/plan', 'in', 'beads/finding', ?, 1, '2026-09-17 00:00:00', '2026-09-17 00:00:00')", path, relationType, graph.MintRevision().String(), []byte("{}"), strings.Repeat("a", 32))
		if err != nil {
			t.Fatal(err)
		}
	}
	if _, err := readBeadInTx(ctx, tx, scope, "beads/plan", fixtureLimits); !errors.Is(err, errCorrupt) || errors.Is(err, errBudget) {
		t.Fatalf("actual owned descriptor overflow: %v", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
}

func verifyEngineByteControls(t *testing.T, ctx context.Context, db fixtureDB) {
	t.Helper()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	var originalMode string
	if err := tx.QueryRowContext(ctx, "SELECT @@session.sql_mode").Scan(&originalMode); err != nil {
		_ = tx.Rollback()
		t.Fatal(err)
	}
	restored := false
	defer func() {
		if !restored {
			if err := setFixtureSQLMode(ctx, tx, originalMode); err != nil {
				t.Errorf("restore SQL mode: %v", err)
			}
		}
		_ = tx.Rollback()
	}()
	scope := scopeFromFixtureTx(t, ctx, tx)
	for _, mode := range []string{"STRICT_ALL_TABLES", ""} {
		if err := setFixtureSQLMode(ctx, tx, mode); err != nil {
			t.Fatal(err)
		}
		t.Logf("raw BLOB controls verified session sql_mode=%q", mode)
		for _, resource := range []struct {
			table, path string
			read        func() error
		}{
			{"graph_beads", "beads/plan", func() error { _, e := readBeadInTx(ctx, tx, scope, "beads/plan", fixtureLimits); return e }},
			{"graph_links", "links/plan-decision", func() error { _, e := readLinkInTx(ctx, tx, scope, "links/plan-decision", fixtureLimits); return e }},
		} {
			var original []byte
			if err := tx.QueryRowContext(ctx, "SELECT properties FROM "+resource.table+" WHERE path = ?", resource.path).Scan(&original); err != nil {
				t.Fatal(err)
			}
			invalid := append(append([]byte(nil), original...), 0xff)
			oversized := []byte(`{"text":"` + strings.Repeat("☃", fixtureLimits.valueBytes/3+1) + `"}`)
			for _, control := range []struct {
				name string
				raw  []byte
				want error
			}{{"invalid UTF8", invalid, errCorrupt}, {"oversized multibyte", oversized, errBudget}} {
				if _, err := tx.ExecContext(ctx, "UPDATE "+resource.table+" SET properties = ? WHERE path = ?", control.raw, resource.path); err != nil {
					t.Fatal(err)
				}
				verifyFixtureBlobProjection(t, ctx, tx, resource.table, "properties", "path", resource.path, control.raw)
				other := errBudget
				if control.want == errBudget {
					other = errCorrupt
				}
				if err := resource.read(); !errors.Is(err, control.want) || errors.Is(err, other) {
					t.Fatalf("%s %s mode=%q: %v", resource.path, control.name, mode, err)
				}
				if resource.table == "graph_links" {
					if _, err := readIncidentLinksInTx(ctx, tx, scope, "beads/plan", graph.DirectionOut, fixtureLimits); !errors.Is(err, control.want) || errors.Is(err, other) {
						t.Fatalf("incident %s mode=%q: %v", control.name, mode, err)
					}
				}
			}
			if _, err := tx.ExecContext(ctx, "UPDATE "+resource.table+" SET properties = ? WHERE path = ?", original, resource.path); err != nil {
				t.Fatal(err)
			}
		}
		for _, descriptor := range []struct {
			id   string
			read func() error
		}{
			{memoryType, func() error { _, e := readBeadInTx(ctx, tx, scope, "beads/plan", fixtureLimits); return e }},
			{relationType, func() error { _, e := readLinkInTx(ctx, tx, scope, "links/plan-decision", fixtureLimits); return e }},
		} {
			var original []byte
			if err := tx.QueryRowContext(ctx, "SELECT descriptor FROM graph_type_descriptors WHERE url = ?", descriptor.id).Scan(&original); err != nil {
				t.Fatal(err)
			}
			// Keep the original fingerprint: repairing the invalid suffix would falsely pass.
			invalid := append(append([]byte(nil), original...), 0xff)
			if _, err := tx.ExecContext(ctx, "UPDATE graph_type_descriptors SET descriptor = ? WHERE url = ?", invalid, descriptor.id); err != nil {
				t.Fatal(err)
			}
			verifyFixtureBlobProjection(t, ctx, tx, "graph_type_descriptors", "descriptor", "url", descriptor.id, invalid)
			if err := descriptor.read(); !errors.Is(err, errCorrupt) || errors.Is(err, errBudget) {
				t.Fatalf("descriptor raw-byte preservation mode=%q: %v", mode, err)
			}
			if _, err := tx.ExecContext(ctx, "UPDATE graph_type_descriptors SET descriptor = ? WHERE url = ?", original, descriptor.id); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := setFixtureSQLMode(ctx, tx, originalMode); err != nil {
		t.Fatal(err)
	}
	restored = true
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
}

// Session SET success alone does not establish that the requested mode took effect.
func setFixtureSQLMode(ctx context.Context, tx *sql.Tx, mode string) error {
	if _, err := tx.ExecContext(ctx, "SET SESSION sql_mode = ?", mode); err != nil {
		return err
	}
	var observed string
	if err := tx.QueryRowContext(ctx, "SELECT @@session.sql_mode").Scan(&observed); err != nil {
		return err
	}
	if observed != mode {
		return fmt.Errorf("session sql_mode = %q, want %q", observed, mode)
	}
	return nil
}

// Use the production projection expression, independently of the decoder's
// refusal: a Go-side budget/length check could otherwise hide SQL repair or
// transfer of an oversized value. Identifiers are fixed fixture call sites.
func verifyFixtureBlobProjection(t *testing.T, ctx context.Context, tx *sql.Tx, table, column, key, id string, persisted []byte) {
	t.Helper()
	_, projection, ok := strings.Cut(beadColumns, "CASE WHEN ")
	if !ok {
		t.Fatal("production raw BLOB projection missing")
	}
	projection = strings.ReplaceAll("CASE WHEN "+projection, "properties", column)
	var raw []byte
	var length sql.NullInt64
	if err := tx.QueryRowContext(ctx, "SELECT "+projection+" FROM "+table+" WHERE "+key+" = ?", fixtureLimits.valueBytes, id).Scan(&raw, &length); err != nil {
		t.Fatal(err)
	}
	if !length.Valid || length.Int64 != int64(len(persisted)) {
		t.Fatalf("%s %s projected length=%v, want %d bytes", table, id, length, len(persisted))
	}
	if len(persisted) > fixtureLimits.valueBytes {
		if raw != nil {
			t.Fatalf("%s %s transferred %d oversized bytes, want SQL NULL", table, id, len(raw))
		}
	} else if !bytes.Equal(raw, persisted) {
		t.Fatalf("%s %s projection changed raw bytes", table, id)
	}
}

// These are test projection rows, not evidence of applied allocation events.
func seedFixtureAllocation(t *testing.T, ctx context.Context, tx *sql.Tx, path string, kind graph.ResourceKind) {
	t.Helper()
	if _, err := tx.ExecContext(ctx, "INSERT INTO graph_allocations (path, resource_kind, birth_seq, birth_authority_id, birth_authority_epoch, state, tombstone_seq, last_authority_id, last_authority_epoch) VALUES (?, ?, 1, REPEAT('a',32), 1, 'live', NULL, REPEAT('a',32), 1)", path, string(kind)); err != nil {
		t.Fatal(err)
	}
}
