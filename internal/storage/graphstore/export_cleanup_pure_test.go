package graphstore

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/types"
)

// These tests use database/sql's real cleanup paths with a test-owned connector,
// not an engine or network server. The fixture is an empty controlled workspace
// with the complete persisted Type installation. It proves error propagation, not
// engine consistency, SQL query correctness, or the publication contract.
func TestExportCurrentSnapshotRollbackCleanup(t *testing.T) {
	for _, failure := range []bool{false, true} {
		t.Run(fmt.Sprintf("failure=%t", failure), func(t *testing.T) {
			fault := errors.New("driver transaction rollback failed")
			faults := exportCleanupFaults{}
			if failure {
				faults.rollback = fault
			}
			s, d := exportCleanupStore(t, faults)
			acquired := false
			s.afterWrite = func(stage string) error {
				if stage == "export-comments" {
					acquired = true
				}
				return nil
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			state, err := s.ExportCurrentSnapshot(ctx)
			if !acquired {
				t.Fatalf("failure occurred before complete detached acquisition: %v", err)
			}
			if failure {
				if !errors.Is(err, fault) || !reflect.DeepEqual(state, ExportCurrentState{}) {
					t.Fatalf("rollback failure leaked state/lost error: %+v %v", state, err)
				}
			} else {
				want := ExportCurrentState{PinnedEndpoints: []any{}, IssuePrefix: "web-app-", CommentsByIssuePath: map[string][]types.Comment{}, Snapshot: Snapshot{WriterToken: exportCleanupToken, Records: []any{}, Types: []graph.TypeDescriptor{}, OwnedLinkTypes: map[string][]string{}, OwnedLinkTypesByResource: map[string][]string{}}}
				for _, descriptor := range d.descriptors {
					parsed, parseErr := graph.ParseTypeDescriptor(descriptor.CanonicalJSON())
					if parseErr != nil {
						t.Fatal(parseErr)
					}
					want.Snapshot.Types = append(want.Snapshot.Types, parsed)
					if parsed.Describes() == graph.KindBead {
						owned, ownedErr := s.SelectedOwnedLinkTypes(parsed.ID())
						if ownedErr != nil {
							t.Fatal(ownedErr)
						}
						want.Snapshot.OwnedLinkTypes[parsed.ID()] = owned
					}
				}
				sort.Slice(want.Snapshot.Types, func(i, j int) bool { return want.Snapshot.Types[i].ID() < want.Snapshot.Types[j].ID() })
				if err != nil || len(want.Snapshot.Types) != 52 || !reflect.DeepEqual(state, want) {
					t.Fatalf("successful complete acquisition differs: got=%+v want=%+v error=%v", state, want, err)
				}
			}
			exportCleanupTransactionCount(t, s, d)
		})
	}
}

func TestExportCurrentSnapshotCommentRowsCleanup(t *testing.T) {
	rowsFault := errors.New("driver comment rows close failed")
	rollbackFault := errors.New("driver transaction rollback also failed")
	s, d := exportCleanupStore(t, exportCleanupFaults{commentClose: rowsFault, rollback: rollbackFault})
	inventoryAcquired, commentsAcquired := false, false
	s.afterWrite = func(stage string) error {
		switch stage {
		case "export-inventory":
			inventoryAcquired = true
		case "export-comments":
			commentsAcquired = true
		}
		return nil
	}
	state, err := s.ExportCurrentSnapshot(context.Background())
	if !errors.Is(err, rowsFault) || !errors.Is(err, rollbackFault) || !reflect.DeepEqual(state, ExportCurrentState{}) {
		t.Fatalf("actual rows/rollback cleanup failure leaked state/lost chain: %+v %v", state, err)
	}
	if !inventoryAcquired || commentsAcquired || d.commentCloses.Load() != 1 {
		t.Fatalf("did not exercise comment EOF cleanup: inventory=%t comments=%t closes=%d", inventoryAcquired, commentsAcquired, d.commentCloses.Load())
	}
	exportCleanupTransactionCount(t, s, d)
}

func TestExportCurrentSnapshotBadConnectionCleanup(t *testing.T) {
	fault := errors.New("rollback lost its driver connection")
	s, d := exportCleanupStore(t, exportCleanupFaults{rollback: errors.Join(fault, driver.ErrBadConn)})
	state, err := s.ExportCurrentSnapshot(context.Background())
	// An ErrBadConn transaction result automatically closes the logical Conn.
	// withTx's subsequent checked Conn.Close therefore returns ErrConnDone.
	// This tests logical connection cleanup separately from driver.Conn.Close.
	if !errors.Is(err, fault) || !errors.Is(err, driver.ErrBadConn) || !errors.Is(err, sql.ErrConnDone) || !reflect.DeepEqual(state, ExportCurrentState{}) {
		t.Fatalf("bad connection cleanup lost chain/exposed state: %+v %v", state, err)
	}
	if d.connectionCloses.Load() != 1 {
		t.Fatalf("bad driver connection was not discarded: %d", d.connectionCloses.Load())
	}
	exportCleanupTransactionCount(t, s, d)
}

func TestExportCurrentSnapshotPhysicalCloseBoundary(t *testing.T) {
	fault := errors.New("physical driver connection close failed")
	s, d := exportCleanupStore(t, exportCleanupFaults{connectionClose: fault})
	state, err := s.ExportCurrentSnapshot(context.Background())
	// Logical Conn.Close returns a healthy connection to the pool. It does not
	// call driver.Conn.Close; the reader still leaves Store.Close to its caller.
	if err != nil || state.CommentsByIssuePath == nil || d.connectionCloses.Load() != 0 {
		t.Fatalf("logical connection return confused with physical close: %+v %v closes=%d", state, err, d.connectionCloses.Load())
	}
	exportCleanupTransactionCount(t, s, d)
	if err := s.Close(); !errors.Is(err, fault) {
		t.Fatalf("caller-owned Store.Close lost physical cleanup failure: %v", err)
	}
	if err := s.Close(); !errors.Is(err, fault) || d.connectionCloses.Load() != 1 {
		t.Fatalf("repeated cleanup lost failure or closed twice: %v closes=%d", err, d.connectionCloses.Load())
	}
}

func TestExportCommentsScanAndCloseCleanup(t *testing.T) {
	rowsFault := errors.New("driver rows close failed after scan error")
	rollbackFault := errors.New("driver rollback failed after scan error")
	s, d := exportCleanupStore(t, exportCleanupFaults{
		commentClose: rowsFault, rollback: rollbackFault,
		// A string cannot be scanned into the feed's time.Time destination.
		commentValues: [][]driver.Value{{"comment-a", "web-app-a1b", "author", "original", "invalid driver timestamp"}},
	})
	var leaked bool
	err := s.withTx(context.Background(), false, func(tx *sql.Tx) error {
		comments, err := exportCommentsInTx(context.Background(), tx, map[string]string{"web-app-a1b": "beads/work"}, 1)
		leaked = comments != nil
		return err
	})
	if err == nil || leaked || !strings.Contains(err.Error(), "Scan error") || !errors.Is(err, rowsFault) || !errors.Is(err, rollbackFault) || d.commentCloses.Load() != 1 {
		t.Fatalf("scan/rows-close/rollback error chain: leaked=%t closes=%d err=%v", leaked, d.commentCloses.Load(), err)
	}
	exportCleanupTransactionCount(t, s, d)
}

// Direct feeds exercise nonempty results without fabricating the native Issue
// hydration queried by the full reader. Real mixed-store acquisition remains an
// engine gate; database/sql still owns every rows/transaction cleanup here.
func TestExportCommentsFeedCleanup(t *testing.T) {
	stamp := time.Date(2026, time.October, 10, 12, 30, 45, 123000000, time.UTC)
	first := []driver.Value{"comment-a", "web-app-a1b", "author", "original", stamp}
	second := []driver.Value{"comment-b", "web-app-a1b", "another author", "next", stamp.Add(time.Second)}
	bad := func(field int, value string) []driver.Value {
		row := append([]driver.Value(nil), second...)
		row[field] = value
		return row
	}
	for _, tc := range []struct {
		name       string
		values     [][]driver.Value
		count      uint64
		message    string
		closeFault bool
	}{
		{name: "valid", values: [][]driver.Value{first, second}, count: 2},
		{name: "invalid-id-utf8", values: [][]driver.Value{first, bad(0, "comment-\xff")}, count: 2, message: "not UTF-8", closeFault: true},
		{name: "invalid-owner-utf8", values: [][]driver.Value{first, bad(1, "web-app-\xff")}, count: 2, message: "not UTF-8", closeFault: true},
		{name: "invalid-author-utf8", values: [][]driver.Value{first, bad(2, "author-\xff")}, count: 2, message: "not UTF-8", closeFault: true},
		{name: "invalid-text-utf8", values: [][]driver.Value{first, bad(3, "orig\xffinal")}, count: 2, message: "not UTF-8", closeFault: true},
		{name: "unknown-owner", values: [][]driver.Value{first, bad(1, "web-app-other")}, count: 2, message: "owner missing", closeFault: true},
		{name: "count-overflow", values: [][]driver.Value{first, second}, count: 1, message: "differs from its bounded inventory", closeFault: true},
		// A Close error at EOF takes precedence over incomplete-feed validation.
		// Keep Close healthy here to exercise that later refusal itself.
		{name: "incomplete-feed", values: [][]driver.Value{first}, count: 2, message: "incomplete comment feed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rowsFault := errors.New("direct feed rows close failed")
			rollbackFault := errors.New("direct feed rollback failed")
			faults := exportCleanupFaults{commentValues: tc.values}
			if tc.message != "" {
				faults.rollback = rollbackFault
			}
			if tc.closeFault {
				faults.commentClose = rowsFault
			}
			s, d := exportCleanupStore(t, faults)
			var feed map[string][]types.Comment
			err := s.withTx(context.Background(), false, func(tx *sql.Tx) error {
				var err error
				feed, err = exportCommentsInTx(context.Background(), tx, map[string]string{
					"web-app-a1b": "beads/work", "web-app-empty": "beads/empty",
				}, tc.count)
				return err
			})
			if tc.message == "" {
				want := map[string][]types.Comment{
					"beads/work": {
						{ID: "comment-a", IssueID: "web-app-a1b", Author: "author", Text: "original", CreatedAt: stamp},
						{ID: "comment-b", IssueID: "web-app-a1b", Author: "another author", Text: "next", CreatedAt: stamp.Add(time.Second)},
					},
					"beads/empty": {},
				}
				if err != nil || !reflect.DeepEqual(feed, want) {
					t.Fatalf("valid comment feed changed fields/ownership/order: %+v err=%v", feed, err)
				}
			} else if feed != nil || !errors.Is(err, ErrInvalidStore) || !errors.Is(err, rollbackFault) || !strings.Contains(err.Error(), tc.message) {
				t.Fatalf("failed feed leaked partial map/lost refusal/rollback: %+v err=%v", feed, err)
			}
			if tc.closeFault && !errors.Is(err, rowsFault) {
				t.Fatalf("lost explicit rows-close failure: %v", err)
			}
			if d.commentCloses.Load() != 1 {
				t.Fatalf("direct feed physically closed %d times", d.commentCloses.Load())
			}
			exportCleanupTransactionCount(t, s, d)
		})
	}
}

func TestExportCommentsNextCleanup(t *testing.T) {
	for _, failClose := range []bool{false, true} {
		t.Run(fmt.Sprintf("close-failure=%t", failClose), func(t *testing.T) {
			nextFault := errors.New("driver comment Next failed after partial feed")
			rowsFault := errors.New("driver comment Close also failed")
			rollbackFault := errors.New("driver rollback failed after Next error")
			faults := exportCleanupFaults{
				commentNext: nextFault, rollback: rollbackFault,
				commentValues: [][]driver.Value{{"comment-a", "web-app-a1b", "author", "original", time.Date(2026, time.October, 10, 12, 30, 45, 0, time.UTC)}},
			}
			if failClose {
				faults.commentClose = rowsFault
			}
			s, d := exportCleanupStore(t, faults)
			var feed map[string][]types.Comment
			err := s.withTx(context.Background(), false, func(tx *sql.Tx) error {
				var err error
				feed, err = exportCommentsInTx(context.Background(), tx, map[string]string{"web-app-a1b": "beads/work"}, 2)
				return err
			})
			if feed != nil || !errors.Is(err, nextFault) || !errors.Is(err, rollbackFault) || d.commentCloses.Load() != 1 {
				t.Fatalf("Next failure leaked partial map/lost cleanup chain: %+v closes=%d err=%v", feed, d.commentCloses.Load(), err)
			}
			// database/sql auto-closes on non-EOF Next errors. Rows.Err keeps the
			// Next cause, while the later explicit Rows.Close is idempotent. A
			// simultaneous driver Close fault is suppressed by that API boundary;
			// do not claim the reader can propagate a cause SQL never exposes.
			if errors.Is(err, rowsFault) {
				t.Fatalf("unexpected auto-close error propagation alongside Next: %v", err)
			}
			exportCleanupTransactionCount(t, s, d)
		})
	}
}

func exportCleanupTransactionCount(t *testing.T, s *Store, d *exportCleanupDriver) {
	t.Helper()
	if d.begins.Load() != 1 || d.rollbacks.Load() != 1 || d.commits.Load() != 0 || s.db.Stats().InUse != 0 {
		t.Fatalf("cleanup was not one rollback-only transaction: begin=%d rollback=%d commit=%d in-use=%d", d.begins.Load(), d.rollbacks.Load(), d.commits.Load(), s.db.Stats().InUse)
	}
}

const exportCleanupToken = "0123456789abcdef0123456789abcdef"

type exportCleanupFaults struct {
	rollback, commentClose, commentNext, connectionClose, pinnedFault error
	commentValues                                                     [][]driver.Value
	pinned, duplicateIdentity                                         bool
}

type exportCleanupDriver struct {
	faults                                                      exportCleanupFaults
	options                                                     Options
	descriptors                                                 map[string]graph.TypeDescriptor
	memory                                                      Record
	target, retained                                            Record
	link, citation                                              LinkRecord
	identityOpens, identityCloses                               atomic.Int64
	privateBounds, privateIdentities, publicBounds, publicURLs  map[string]int
	pinnedReads                                                 atomic.Int64
	begins, rollbacks, commits, commentCloses, connectionCloses atomic.Int64
	foreignKeyReads, foreignKeyRestores                         atomic.Int64
}

func exportCleanupStore(t *testing.T, faults exportCleanupFaults) (*Store, *exportCleanupDriver) {
	t.Helper()
	d := &exportCleanupDriver{
		faults: faults,
		options: Options{Binding: Binding{
			WorkspaceID: "/disposable/export-cleanup-fixture", ScopeURL: "https://example.test/cleanup/",
			AuthorityID: exportCleanupToken, SchemaVersion: SchemaVersion,
		}},
		descriptors:   make(map[string]graph.TypeDescriptor),
		privateBounds: make(map[string]int), privateIdentities: make(map[string]int), publicBounds: make(map[string]int), publicURLs: make(map[string]int),
	}
	for _, definition := range previewTypeDefinitions() {
		descriptor, err := definition.build(d.options.Binding.ScopeURL)
		if err != nil {
			t.Fatal(err)
		}
		d.descriptors[definition.name] = descriptor
	}
	if faults.pinned {
		scope := d.options.Binding.ScopeURL
		attribution := Attribution{Status: "unknown", RecordedAt: "2026-10-20T00:00:00Z"}
		d.memory = Record{PublicID: "links/literal-memory", ID: scope + "links/literal-memory", Type: MemoryTypeURL(scope), Revision: exportCleanupToken, Version: exportCleanupToken, Properties: Properties{Title: "Memory"}, Metadata: json.RawMessage(`{}`), Owned: []json.RawMessage{}, Attribution: attribution}
		d.target = d.memory
		d.target.ID, d.target.PublicID = scope+"types/literal-target", "types/literal-target"
		d.target.Properties = Properties{Title: "Target Memory"}
		d.retained = d.target
		d.retained.Revision = "11111111111111111111111111111111"
		d.retained.Version = d.retained.Revision
		d.retained.Properties = Properties{Title: "Retained Memory", Body: "Before the current edit"}
		d.retained.Metadata = json.RawMessage(`{"retained":true}`)
		d.link = LinkRecord{ID: scope + "links/reference", Type: scope + "types/references", Revision: exportCleanupToken, Version: exportCleanupToken, Source: d.memory.ID, Target: d.target.ID, TargetRevision: d.retained.Revision, Properties: map[string]any{}, Metadata: json.RawMessage(`{}`), Attribution: attribution}
		d.citation = d.link
		d.citation.ID = scope + "links/citation"
	}
	db := sql.OpenDB(d)
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	s := &Store{db: db, options: d.options, cleanup: db.Close}
	t.Cleanup(func() {
		if err := s.Close(); err != nil && !errors.Is(err, faults.connectionClose) {
			t.Error(err)
		}
	})
	return s, d
}

func (d *exportCleanupDriver) Driver() driver.Driver { return d }
func (d *exportCleanupDriver) Open(string) (driver.Conn, error) {
	return d.Connect(context.Background())
}
func (d *exportCleanupDriver) Connect(ctx context.Context) (driver.Conn, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return &exportCleanupConn{d}, nil
}

type exportCleanupConn struct{ owner *exportCleanupDriver }

func (c *exportCleanupConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("cleanup fixture does not prepare statements")
}
func (c *exportCleanupConn) Close() error {
	c.owner.connectionCloses.Add(1)
	return c.owner.faults.connectionClose
}
func (c *exportCleanupConn) Begin() (driver.Tx, error) {
	c.owner.begins.Add(1)
	return &exportCleanupTx{c.owner}, nil
}
func (c *exportCleanupConn) ExecContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Result, error) {
	switch strings.Join(strings.Fields(strings.ToLower(query)), " ") {
	case "set @@dolt_transaction_commit = 0", "set @@dolt_force_transaction_commit = 0", "set @@dolt_allow_commit_conflicts = 0", "set @@time_zone = '+00:00'":
		return driver.RowsAffected(0), nil
	case "set @@foreign_key_checks = 0":
		return driver.RowsAffected(0), nil
	case "set @@foreign_key_checks = 1":
		c.owner.foreignKeyRestores.Add(1)
		return driver.RowsAffected(0), nil
	default:
		return nil, fmt.Errorf("cleanup fixture refused non-session SQL: %s", query)
	}
}

func (c *exportCleanupConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	q := strings.Join(strings.Fields(strings.ToLower(query)), " ")
	d := c.owner
	if d.faults.pinned {
		if rows, handled, err := d.pinnedRows(q, args); handled {
			return rows, err
		}
	}
	row := func(columns []string, values ...driver.Value) driver.Rows {
		return &exportCleanupRows{columns: columns, values: [][]driver.Value{values}}
	}
	switch {
	case q == "select @@foreign_key_checks":
		d.foreignKeyReads.Add(1)
		return row([]string{"foreign_key_checks"}, int64(1)), nil
	case q == "select active_branch()":
		return row([]string{"branch"}, "main"), nil
	case q == "select count(*) from graph_preview_types":
		return row([]string{"count"}, int64(len(d.descriptors))), nil
	case q == "select name,case when octet_length(descriptor)<=? then descriptor else null end,fingerprint from graph_preview_types":
		if len(args) != 1 || args[0].Value != int64(len(d.options.Binding.ScopeURL)*4+8192) {
			return nil, fmt.Errorf("cleanup bounded descriptor arguments: %v", args)
		}
		rows := &exportCleanupRows{columns: []string{"name", "descriptor", "fingerprint"}}
		for name, descriptor := range d.descriptors {
			var raw driver.Value
			if int64(len(descriptor.CanonicalJSON())) <= args[0].Value.(int64) {
				raw = descriptor.CanonicalJSON()
			}
			rows.values = append(rows.values, []driver.Value{name, raw, descriptor.Fingerprint()})
		}
		return rows, nil
	case strings.HasPrefix(q, "select descriptor, fingerprint from graph_preview_types where name=?"), strings.HasPrefix(q, "select descriptor,fingerprint from graph_preview_types where name=?"):
		if len(args) != 2 {
			return nil, fmt.Errorf("cleanup descriptor lookup arguments: %v", args)
		}
		descriptor, ok := d.descriptors[fmt.Sprint(args[0].Value)]
		if !ok || args[1].Value != int64(len(descriptor.CanonicalJSON())) {
			return nil, errors.New("cleanup descriptor lookup did not match persisted Type")
		}
		return row([]string{"descriptor", "fingerprint"}, descriptor.CanonicalJSON(), descriptor.Fingerprint()), nil
	case strings.HasPrefix(q, "select workspace, scope_url, authority_id, schema_version, writer_token from graph_preview_scope"):
		b := d.options.Binding
		return row([]string{"workspace", "scope", "authority", "schema", "token"}, b.WorkspaceID, b.ScopeURL, b.AuthorityID, int64(b.SchemaVersion), exportCleanupToken), nil
	case q == "select writer_token from graph_preview_scope where singleton=1":
		return row([]string{"token"}, exportCleanupToken), nil
	case q == "select value from config where `key`='issue_prefix'":
		if len(args) != 0 {
			return nil, errors.New("literal prefix lookup had arguments")
		}
		return row([]string{"value"}, "web-app-"), nil
	case q == "select value from config where `key` = ?":
		if len(args) != 1 {
			return nil, fmt.Errorf("cleanup config lookup arguments: %v", args)
		}
		switch args[0].Value {
		case "issue_prefix":
			return row([]string{"value"}, "web-app-"), nil
		case "types.custom":
			return &exportCleanupRows{columns: []string{"value"}}, nil
		default:
			return nil, fmt.Errorf("cleanup fixture unknown config key: %v", args[0].Value)
		}
	case q == "select name from custom_types order by name":
		return &exportCleanupRows{columns: []string{"name"}}, nil
	case strings.HasPrefix(q, "select id,issue_id,author,text,created_at from comments order by"):
		return &exportCleanupRows{columns: []string{"id", "issue_id", "author", "text", "created_at"}, values: d.faults.commentValues, closeErr: d.faults.commentClose, nextErr: d.faults.commentNext, closes: &d.commentCloses}, nil
	case strings.HasSuffix(q, "limit 0"):
		return &exportCleanupRows{columns: []string{"schema-guard"}}, nil
	case strings.HasPrefix(q, "select path,backing,resource_kind from graph_preview_catalog"):
		return &exportCleanupRows{columns: []string{"path", "backing", "kind"}}, nil
	case strings.HasPrefix(q, "select path from graph_preview_catalog"):
		return &exportCleanupRows{columns: []string{"path"}}, nil
	case strings.HasPrefix(q, "select 1 from graph_preview_catalog"):
		return &exportCleanupRows{columns: []string{"invalid"}}, nil
	case strings.HasPrefix(q, "select count(*)"), strings.HasPrefix(q, "select exists("), strings.HasPrefix(q, "select cast(least("):
		// Empty current tables: no allocations, relations, comments, alternative
		// planes or mapping failures. Accounting values are irrelevant to this
		// cleanup test; real engine tests own acquisition-bound correctness.
		return row([]string{"count"}, int64(0)), nil
	default:
		return nil, fmt.Errorf("unsupported cleanup fixture query: %s", query)
	}
}

// The populated connector uses two distinct Beads and two distinct admissible
// informational Links selecting one older endpoint. It tests the database/sql
// boundary and explicit query/argument dispatch, not real-engine SQL semantics.
func (d *exportCleanupDriver) pinnedRows(q string, args []driver.NamedValue) (driver.Rows, bool, error) {
	row := func(columns string, values ...driver.Value) (driver.Rows, bool, error) {
		return &exportCleanupRows{columns: strings.Split(columns, ","), values: [][]driver.Value{values}}, true, nil
	}
	mismatch := func() (driver.Rows, bool, error) {
		return nil, true, fmt.Errorf("cleanup populated SQL/arguments mismatch: %s %v", q, args)
	}
	memories := map[string]Record{"beads/memory": d.memory, "beads/target": d.target}
	links := map[string]LinkRecord{"links/reference": d.link, "links/citation": d.citation}
	pathArg := func(index int) string {
		if index < len(args) {
			if path, ok := args[index].Value.(string); ok {
				return path
			}
		}
		return ""
	}
	identity := func(path string, duplicate bool) (driver.Rows, bool, error) {
		var kind, backing, publicID string
		if memory, ok := memories[path]; ok {
			kind, backing, publicID = "bead", "generic", memory.PublicID
		} else if _, ok := links[path]; ok {
			kind, backing, publicID = "link", "informational", path
		} else {
			return mismatch()
		}
		d.identityOpens.Add(1)
		rows := &exportCleanupRows{columns: []string{"path", "resource_kind", "allocation_state", "backing", "public_id"}, values: [][]driver.Value{{path, kind, "live", backing, publicID}}, closes: &d.identityCloses}
		if duplicate {
			rows.values = append(rows.values, []driver.Value{"beads/duplicate", kind, "live", backing, publicID})
		}
		return rows, true, nil
	}
	switch q {
	case "select path,resource_kind,allocation_state,backing, case when backing='issue' then backing_key else coalesce(public_id,path) end from graph_preview_catalog where path=? limit 2":
		if len(args) != 1 {
			return mismatch()
		}
		d.privateIdentities[pathArg(0)]++
		return identity(pathArg(0), false)
	case "select path,resource_kind,allocation_state,backing, case when backing='issue' then backing_key else coalesce(public_id,path) end from graph_preview_catalog where (backing='issue' and backing_key=?) or (backing in ('generic','dependency','informational') and coalesce(public_id,path)=?) limit 2":
		if len(args) != 2 || pathArg(0) != d.target.PublicID || pathArg(1) != d.target.PublicID {
			return mismatch()
		}
		return identity("beads/target", d.faults.duplicateIdentity)
	case "select cast(least(coalesce(sum( 4*(256+octet_length(path)+coalesce(octet_length(backing_key),0)+coalesce(octet_length(public_id),0)) ),0),?) as unsigned) from graph_preview_catalog where path=?":
		if len(args) != 2 || args[0].Value != int64(PreviewCurrentReadByteLimit+1) {
			return mismatch()
		}
		path := pathArg(1)
		if _, bead := memories[path]; !bead {
			if _, link := links[path]; !link {
				return mismatch()
			}
		}
		d.privateBounds[path]++
		return row("bytes", int64(1024))
	case "select cast(least(coalesce(sum( 4*(256+octet_length(path)+coalesce(octet_length(backing_key),0)+coalesce(octet_length(public_id),0)) ),0),?) as unsigned) from graph_preview_catalog where (backing='issue' and backing_key=?) or (backing in ('generic','dependency','informational') and coalesce(public_id,path)=?)":
		if len(args) != 3 || args[0].Value != int64(PreviewCurrentReadByteLimit+1) || pathArg(1) != d.target.PublicID || pathArg(2) != d.target.PublicID {
			return mismatch()
		}
		d.publicBounds[pathArg(1)]++
		return row("bytes", int64(1024))
	case "select path,backing,resource_kind from graph_preview_catalog where allocation_state='live' limit ?":
		if len(args) != 1 || args[0].Value != int64(PreviewSnapshotLimit+1) {
			return mismatch()
		}
		return &exportCleanupRows{columns: []string{"path", "backing", "kind"}, values: [][]driver.Value{{"beads/memory", "generic", "bead"}, {"beads/target", "generic", "bead"}, {"links/citation", "informational", "link"}, {"links/reference", "informational", "link"}}}, true, nil
	case "select resource_kind, type_url, revision, allocation_state, backing from graph_preview_catalog where path = ?":
		memory, ok := memories[pathArg(0)]
		if len(args) != 1 || !ok {
			return mismatch()
		}
		return row("kind,type,revision,state,backing", "bead", memory.Type, memory.Revision, "live", "generic")
	case "select coalesce(octet_length(public_id),0) from graph_preview_catalog where path=?", "select public_id from graph_preview_catalog where path=?":
		memory, ok := memories[pathArg(0)]
		if len(args) != 1 || !ok {
			return mismatch()
		}
		d.publicURLs[pathArg(0)]++
		if strings.Contains(q, "octet_length") {
			return row("bytes", int64(len(memory.PublicID)))
		}
		return row("public_id", memory.PublicID)
	case "select properties, metadata from graph_preview_payloads where path = ?":
		memory, ok := memories[pathArg(0)]
		if len(args) != 1 || !ok {
			return mismatch()
		}
		properties, err := canonicalJSON(memory.Properties)
		if err != nil {
			return nil, true, err
		}
		return row("properties,metadata", properties, []byte(memory.Metadata))
	case "select snapshot, actor from graph_preview_versions where path = ? and version = ?", "select snapshot,actor from graph_preview_versions where path=? and version=?":
		if len(args) != 2 {
			return mismatch()
		}
		var value any
		if pathArg(0) == "beads/target" && pathArg(1) == d.retained.Revision {
			d.pinnedReads.Add(1)
			if d.faults.pinnedFault != nil {
				return nil, true, d.faults.pinnedFault
			}
			value = d.retained
		} else if memory, ok := memories[pathArg(0)]; ok && pathArg(1) == memory.Revision {
			value = memory
		} else if link, ok := links[pathArg(0)]; ok && pathArg(1) == link.Revision {
			value = link
		} else {
			return mismatch()
		}
		raw, err := canonicalJSON(value)
		if err != nil {
			return nil, true, err
		}
		return row("snapshot,actor", raw, "")
	case "select type_url from graph_preview_catalog where path=? and resource_kind='bead' and allocation_state='live'":
		memory, ok := memories[pathArg(0)]
		if len(args) != 1 || !ok {
			return mismatch()
		}
		return row("type", memory.Type)
	case "select l.path,c.type_url from graph_preview_links l join graph_preview_catalog c on c.path=l.path and c.resource_kind='link' and c.allocation_state='live' where l.source_path=? order by l.path":
		if len(args) != 1 {
			return mismatch()
		}
		rows := &exportCleanupRows{columns: []string{"path", "type"}}
		switch pathArg(0) {
		case "beads/memory":
			rows.values = [][]driver.Value{{"links/citation", d.citation.Type}, {"links/reference", d.link.Type}}
		case "beads/target":
		default:
			return mismatch()
		}
		return rows, true, nil
	case "select backing from graph_preview_catalog where path=?", "select allocation_state from graph_preview_catalog where path=?":
		if len(args) != 1 {
			return mismatch()
		}
		if _, ok := links[pathArg(0)]; !ok {
			return mismatch()
		}
		if strings.Contains(q, "backing") {
			return row("backing", "informational")
		}
		return row("state", "live")
	case "select resource_kind,type_url,revision,allocation_state,backing,backing_key from graph_preview_catalog where path=?":
		link, ok := links[pathArg(0)]
		if len(args) != 1 || !ok {
			return mismatch()
		}
		return row("kind,type,revision,state,backing,key", "link", link.Type, link.Revision, "live", "informational", nil)
	case "select source_path,target_path,source_revision,target_revision,properties,metadata,attribution from graph_preview_links where path=?":
		link, ok := links[pathArg(0)]
		if len(args) != 1 || !ok {
			return mismatch()
		}
		attribution, err := canonicalJSON(link.Attribution)
		if err != nil {
			return nil, true, err
		}
		return row("source,target,source_pin,target_pin,properties,metadata,attribution", "beads/memory", "beads/target", link.SourceRevision, link.TargetRevision, []byte(`{}`), []byte(link.Metadata), attribution)
	case "select resource_kind,type_url,allocation_state,backing from graph_preview_catalog where path=?":
		memory, ok := memories[pathArg(0)]
		if len(args) != 1 || !ok {
			return mismatch()
		}
		return row("kind,type,state,backing", "bead", memory.Type, "live", "generic")
	case "select resource_kind,backing,allocation_state, case when octet_length(type_url)<=? then type_url else null end,backing_key from graph_preview_catalog where path=?", "select backing,resource_kind,allocation_state,revision, case when octet_length(type_url)<=? then type_url else null end,backing_key from graph_preview_catalog where path=?":
		if len(args) != 2 || args[0].Value != int64(len(d.options.Binding.ScopeURL)+512) || pathArg(1) != "beads/target" {
			return mismatch()
		}
		if strings.HasPrefix(q, "select resource_kind,") {
			return row("kind,backing,state,type,key", "bead", "generic", "live", d.target.Type, nil)
		}
		return row("backing,kind,state,revision,type,key", "generic", "bead", "live", d.target.Revision, d.target.Type, nil)
	case "select exists(select 1 from graph_preview_versions where path=? and version=? and not(case when json_valid(cast(snapshot as char)) then coalesce(json_unquote(json_extract(cast(snapshot as char),'$.state'))='deleted',false) else false end))":
		if len(args) != 2 || pathArg(0) != "beads/target" || pathArg(1) != d.retained.Revision {
			return mismatch()
		}
		return row("exists", true)
	case "select octet_length(snapshot)+octet_length(actor)+coalesce((select octet_length(public_id) from graph_preview_catalog where path=graph_preview_versions.path),0) from graph_preview_versions where path=? and version=?":
		if len(args) != 2 || pathArg(0) != "beads/target" || pathArg(1) != d.retained.Revision {
			return mismatch()
		}
		raw, err := canonicalJSON(d.retained)
		if err != nil {
			return nil, true, err
		}
		return row("bytes", int64(len(raw)+len(d.target.PublicID)))
	}
	// An unrecognized populated acquisition query must not silently fall through
	// to an empty fixture row. Global mapping/aggregate probes remain below.
	for _, prefix := range []string{"select path,resource_kind", "select resource_kind", "select backing,resource_kind", "select snapshot", "select source_path", "select l.path", "select type_url", "select properties", "select public_id", "select coalesce(octet_length(public_id)", "select octet_length(snapshot)", "select cast(least(coalesce(sum( 4*(256+octet_length(path)", "select exists(select 1 from graph_preview_versions where path=?"} {
		if strings.HasPrefix(q, prefix) {
			return mismatch()
		}
	}
	return nil, false, nil
}

func exportCleanupExpectedState(t *testing.T, s *Store, d *exportCleanupDriver) ExportCurrentState {
	t.Helper()
	want := ExportCurrentState{PinnedEndpoints: []any{d.retained}, IssuePrefix: "web-app-", CommentsByIssuePath: map[string][]types.Comment{}, Snapshot: Snapshot{WriterToken: exportCleanupToken, Records: []any{d.memory, d.target, d.citation, d.link}, Types: []graph.TypeDescriptor{}, OwnedLinkTypes: map[string][]string{}, OwnedLinkTypesByResource: map[string][]string{}}}
	var linkTypes []string
	for _, descriptor := range d.descriptors {
		parsed, err := graph.ParseTypeDescriptor(descriptor.CanonicalJSON())
		if err != nil {
			t.Fatal(err)
		}
		want.Snapshot.Types = append(want.Snapshot.Types, parsed)
		if parsed.Describes() == graph.KindLink {
			linkTypes = append(linkTypes, parsed.ID())
		}
	}
	sort.Slice(want.Snapshot.Types, func(i, j int) bool {
		return graph.CompareCodeUnits(want.Snapshot.Types[i].ID(), want.Snapshot.Types[j].ID()) < 0
	})
	for _, descriptor := range want.Snapshot.Types {
		if descriptor.Describes() != graph.KindBead {
			continue
		}
		owned, err := s.SelectedOwnedLinkTypesWithObserved(descriptor.ID(), linkTypes...)
		if err != nil {
			t.Fatal(err)
		}
		want.Snapshot.OwnedLinkTypes[descriptor.ID()] = owned
	}
	for _, memory := range []Record{d.memory, d.target} {
		owned, err := s.SelectedOwnedLinkTypesWithObserved(memory.Type)
		if err != nil {
			t.Fatal(err)
		}
		want.Snapshot.OwnedLinkTypesByResource[memory.ID] = owned
	}
	return want
}

func TestExportPinnedEndpointCleanup(t *testing.T) {
	for _, kind := range []string{"success", "pin-hydration", "duplicate-allocation", "comments", "rollback", "bad-connection", "cancel"} {
		t.Run(kind, func(t *testing.T) {
			fault := errors.New("cleanup failure while acquiring retained endpoint")
			physicalFault := errors.New("populated physical connection close failed")
			faults := exportCleanupFaults{pinned: true}
			switch kind {
			case "success":
				faults.connectionClose = physicalFault
			case "pin-hydration":
				faults.pinnedFault = fault
			case "duplicate-allocation":
				faults.duplicateIdentity = true
			case "comments":
				faults.commentClose = fault
			case "rollback":
				faults.rollback = fault
			case "bad-connection":
				faults.rollback = errors.Join(fault, driver.ErrBadConn)
			}
			s, d := exportCleanupStore(t, faults)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			endpointsAcquired := false
			s.afterWrite = func(stage string) error {
				// Sample before rollback/cancellation can hide a leaked Rows.
				// Normal one-row EOF may auto-close in database/sql: this proves
				// timely release, not whether the explicit Close caused it.
				if d.identityOpens.Load() != d.identityCloses.Load() {
					return errors.New("identity Rows remained open at acquisition boundary")
				}
				if stage == "export-endpoints" {
					endpointsAcquired = true
				}
				if kind == "cancel" && stage == "export-comments" {
					cancel()
				}
				return nil
			}
			state, err := s.ExportCurrentSnapshot(ctx)
			wantReads := int64(1)
			if kind == "duplicate-allocation" {
				wantReads = 0
			}
			if d.pinnedReads.Load() != wantReads {
				t.Fatalf("exact retained reads=%d want=%d err=%v", d.pinnedReads.Load(), wantReads, err)
			}
			if kind == "success" {
				want := exportCleanupExpectedState(t, s, d)
				if err != nil || !reflect.DeepEqual(state, want) {
					t.Fatalf("complete populated acquisition differs: got=%+v want=%+v err=%v", state, want, err)
				}
				if reflect.DeepEqual(state.PinnedEndpoints[0], d.target) {
					t.Fatal("current target substituted for older exact edition")
				}
				// Two distinct Links select the same endpoint; all identity byte
				// guards and corresponding private identity reads remain present.
				wantPrivate := map[string]int{"beads/memory": 3, "beads/target": 4, "links/citation": 1, "links/reference": 1}
				if !reflect.DeepEqual(d.privateBounds, wantPrivate) || !reflect.DeepEqual(d.privateIdentities, wantPrivate) || !reflect.DeepEqual(d.publicBounds, map[string]int{d.target.PublicID: 1}) || !reflect.DeepEqual(d.publicURLs, map[string]int{"beads/memory": 2, "beads/target": 4}) || d.identityOpens.Load() != 10 || d.identityCloses.Load() != 10 {
					t.Fatalf("identity acquisition/charge census: bounds=%v reads=%v public=%v publicURLs=%v opened=%d closed=%d", d.privateBounds, d.privateIdentities, d.publicBounds, d.publicURLs, d.identityOpens.Load(), d.identityCloses.Load())
				}
			} else {
				wantFault := fault
				if kind == "cancel" {
					wantFault = context.Canceled
				}
				if kind == "duplicate-allocation" {
					wantFault = ErrInvalidStore
				}
				if !errors.Is(err, wantFault) || !reflect.DeepEqual(state, ExportCurrentState{}) {
					t.Fatalf("retained state leaked/lost cause: %+v %v", state, err)
				}
				if kind == "duplicate-allocation" && !strings.Contains(err.Error(), "identity lookup has multiple allocations") {
					t.Fatalf("wrong duplicate-allocation refusal: %v", err)
				}
				if (kind == "pin-hydration" || kind == "duplicate-allocation") && endpointsAcquired {
					t.Fatal("endpoint stage reached after mid-acquisition failure")
				}
				if kind == "bad-connection" && (!errors.Is(err, driver.ErrBadConn) || !errors.Is(err, sql.ErrConnDone) || d.connectionCloses.Load() != 1) {
					t.Fatalf("pin bad-connection cleanup chain: %v closes=%d", err, d.connectionCloses.Load())
				}
			}
			if kind != "bad-connection" && d.connectionCloses.Load() != 0 {
				t.Fatalf("logical return physically closed populated connection: %d", d.connectionCloses.Load())
			}
			exportCleanupTransactionCount(t, s, d)
			if kind == "success" {
				if err := s.Close(); !errors.Is(err, physicalFault) || d.connectionCloses.Load() != 1 {
					t.Fatalf("caller physical cleanup lost cause: %v closes=%d", err, d.connectionCloses.Load())
				}
			}
		})
	}
}

func TestExportPinnedEndpointOwnedSourceAndIdentityCleanup(t *testing.T) {
	for _, duplicate := range []bool{false, true} {
		t.Run(fmt.Sprintf("duplicate=%t", duplicate), func(t *testing.T) {
			s, d := exportCleanupStore(t, exportCleanupFaults{pinned: true, duplicateIdentity: duplicate})
			// Synthetic unsupported observation: live writers refuse owned-source
			// pins. Targeting the older Memory as a source exercises the exact
			// helper's ownership refusal after real selected hydration.
			link := d.link
			link.Source, link.Target = d.target.ID, d.memory.ID
			link.SourceRevision, link.TargetRevision = d.retained.Revision, ""
			link.Type = s.ScopeURL() + "types/subject-to"
			err := s.withTx(context.Background(), false, func(tx *sql.Tx) error {
				values, err := s.exportEndpointPinsInTx(context.Background(), tx, Snapshot{Records: []any{link}})
				want := ErrCapabilityUnavailable
				message := "pinned owned source observation point is not selected (retained source)"
				if duplicate {
					want, message = ErrInvalidStore, "identity lookup has multiple allocations"
				}
				if values != nil || !errors.Is(err, want) || !strings.Contains(err.Error(), message) {
					return fmt.Errorf("unsupported exact selection leaked/wrong refusal: %+v %w", values, err)
				}
				// In the duplicate case the reader returned before EOF. This
				// sample is before callback return/rollback, so Tx auto-cleanup
				// cannot satisfy a missing explicit early Rows.Close.
				if d.identityOpens.Load() != d.identityCloses.Load() {
					return errors.New("identity Rows not released before transaction cleanup")
				}
				return err
			})
			wantReads := int64(1)
			wantCloses := int64(2)
			if duplicate {
				wantReads, wantCloses = 0, 1
			}
			if err == nil || d.pinnedReads.Load() != wantReads || d.identityCloses.Load() != wantCloses {
				t.Fatalf("selected source/identity cleanup: reads=%d closes=%d err=%v", d.pinnedReads.Load(), d.identityCloses.Load(), err)
			}
			exportCleanupTransactionCount(t, s, d)
		})
	}
}

func TestExportPinnedEndpointSQLDispatch(t *testing.T) {
	_, d := exportCleanupStore(t, exportCleanupFaults{pinned: true})
	for _, tc := range []struct {
		query string
		args  []driver.NamedValue
	}{
		{"select resource_kind,backing,allocation_state, case when octet_length(type_url)<=? then type_url else null end,backing_key from graph_preview_catalog where path=?", []driver.NamedValue{{Value: int64(1)}, {Value: "beads/target"}}},
		{"select backing,resource_kind,allocation_state,revision, case when octet_length(type_url)<=? then type_url else null end,backing_key from graph_preview_catalog where path=?", []driver.NamedValue{{Value: int64(len(d.options.Binding.ScopeURL) + 512)}, {Value: "beads/wrong"}}},
		{"select exists(select 1 from graph_preview_versions where path=? and version=? and not(case when json_valid(cast(snapshot as char)) then coalesce(json_unquote(json_extract(cast(snapshot as char),'$.state'))='deleted',false) else false end))", []driver.NamedValue{{Value: "beads/wrong"}, {Value: d.retained.Revision}}},
		{"select cast(least(coalesce(sum( 4*(256+octet_length(path)+coalesce(octet_length(backing_key),0)+coalesce(octet_length(public_id),0)) ),0),?) as unsigned) from graph_preview_catalog where path<>?", []driver.NamedValue{{Value: int64(PreviewCurrentReadByteLimit + 1)}, {Value: "beads/target"}}},
		{"select exists(select 1 from graph_preview_versions where path=? and version=?)", []driver.NamedValue{{Value: "beads/target"}, {Value: d.retained.Revision}}},
		{"select backing from graph_preview_catalog where path=?", []driver.NamedValue{{Value: "beads/memory"}}},
		{"select path,backing,resource_kind from graph_preview_catalog where allocation_state='live' limit ?", []driver.NamedValue{{Value: int64(PreviewSnapshotLimit)}}},
		{"select resource_kind,type_url,revision,allocation_state,backing,backing_key from graph_preview_catalog where path=?", []driver.NamedValue{{Value: "beads/target"}}},
	} {
		rows, handled, err := d.pinnedRows(tc.query, tc.args)
		if rows != nil || !handled || err == nil || !strings.Contains(err.Error(), "SQL/arguments mismatch") {
			t.Fatalf("query drift hidden by fixture: %s rows=%v handled=%t err=%v", tc.query, rows, handled, err)
		}
	}
}

type exportCleanupTx struct{ owner *exportCleanupDriver }

func (tx *exportCleanupTx) Commit() error {
	tx.owner.commits.Add(1)
	return errors.New("cleanup fixture forbids committing an export")
}
func (tx *exportCleanupTx) Rollback() error {
	tx.owner.rollbacks.Add(1)
	return tx.owner.faults.rollback
}

type exportCleanupRows struct {
	columns  []string
	values   [][]driver.Value
	position int
	closeErr error
	nextErr  error
	closes   *atomic.Int64
}

func (r *exportCleanupRows) Columns() []string { return r.columns }
func (r *exportCleanupRows) Close() error {
	if r.closes != nil {
		r.closes.Add(1)
	}
	return r.closeErr
}
func (r *exportCleanupRows) Next(dest []driver.Value) error {
	if r.position == len(r.values) {
		if r.nextErr != nil {
			return r.nextErr
		}
		return io.EOF
	}
	copy(dest, r.values[r.position])
	r.position++
	return nil
}

var _ driver.Connector = (*exportCleanupDriver)(nil)
var _ driver.QueryerContext = (*exportCleanupConn)(nil)
var _ driver.ExecerContext = (*exportCleanupConn)(nil)

// The builder seeds by-Type ownership with all installed Link descriptors;
// per-Resource metadata observes only its captured Owned set. We model that
// narrower map for snapshot fidelity, but it is deliberately not an input to
// the authoritative closure check. This control omits a custom native edge.
func TestExportSnapshotClosureCustomOwnedOmission(t *testing.T) {
	const scope = "https://example.test/graph/"
	snapshot := exportClosureFixture(t, scope)
	store := &Store{options: Options{Binding: Binding{ScopeURL: scope}}}
	customType, err := NativeDependencyTypeURL(scope, "custom-kind")
	if err != nil {
		t.Fatal(err)
	}
	descriptor, err := nativeConcreteDescriptor(scope, customType)
	if err != nil {
		t.Fatal(err)
	}
	snapshot.Types = append(snapshot.Types, descriptor)
	link := snapshot.Records[3].(LinkRecord)
	link.Type = customType
	link.Properties = map[string]any{"dep_type": "custom-kind"}
	snapshot.Records[3] = link
	issue := snapshot.Records[1].(IssueRecord)
	issue.Owned[0], err = canonicalJSON(link)
	if err != nil {
		t.Fatal(err)
	}
	snapshot.Records[1] = issue
	var allLinkTypes []string
	for _, descriptor := range snapshot.Types {
		if descriptor.Describes() == graph.KindLink {
			allLinkTypes = append(allLinkTypes, descriptor.ID())
		}
	}
	for typ := range snapshot.OwnedLinkTypes {
		owned, err := store.SelectedOwnedLinkTypesWithObserved(typ, allLinkTypes...)
		if err != nil {
			t.Fatal(err)
		}
		snapshot.OwnedLinkTypes[typ] = owned
	}
	snapshot.OwnedLinkTypesByResource = make(map[string][]string)
	capture := func() {
		for _, value := range snapshot.Records {
			var id, typ string
			var owned []json.RawMessage
			switch record := value.(type) {
			case Record:
				id, typ, owned = record.ID, record.Type, record.Owned
			case IssueRecord:
				id, typ, owned = record.ID, record.Type, record.Owned
			default:
				continue
			}
			var observed []string
			for _, raw := range owned {
				var link LinkRecord
				if err := json.Unmarshal(raw, &link); err != nil {
					t.Fatal(err)
				}
				observed = append(observed, link.Type)
			}
			captured, err := store.SelectedOwnedLinkTypesWithObserved(typ, observed...)
			if err != nil {
				t.Fatal(err)
			}
			snapshot.OwnedLinkTypesByResource[id] = captured
		}
	}
	capture()
	if _, err := exportSnapshotClosure(scope, snapshot); err != nil {
		t.Fatalf("complete custom native ownership: %v", err)
	}
	issue.Owned = []json.RawMessage{}
	snapshot.Records[1] = issue
	capture()
	paths, err := exportSnapshotClosure(scope, snapshot)
	if paths != nil || !errors.Is(err, ErrInvalidStore) || !strings.Contains(err.Error(), "incomplete acquired owned Link set") {
		t.Fatalf("omitted custom owned Link admitted: %+v %v", paths, err)
	}
}
