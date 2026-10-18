//go:build cgo && graphmanaged_engine && (darwin || linux)

package graphsession

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
)

// Deliberately duplicated fixture protocol across two emitted test binaries;
// neither private package exports an authority constructor or SQL callback.
type admissionDescriptor struct {
	Protocol   int    `json:"protocol"`
	Root       string `json:"root"`
	Endpoint   string `json:"endpoint"`
	RunID      string `json:"runId"`
	Generation uint64 `json:"generation"`
}

func parseAdmissionDescriptor(raw []byte) (admissionDescriptor, error) {
	var d admissionDescriptor
	if len(raw) > 4096 {
		return d, errResult
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return d, errResult
	}
	seen := map[string]bool{}
	for decoder.More() {
		token, err := decoder.Token()
		key, ok := token.(string)
		if err != nil || !ok || seen[key] {
			return d, errResult
		}
		seen[key] = true
		var value json.RawMessage
		if decoder.Decode(&value) != nil || bytes.Equal(value, []byte("null")) {
			return d, errResult
		}
	}
	if _, err := decoder.Token(); err != nil {
		return d, err
	}
	if decoder.Decode(new(any)) != io.EOF || len(seen) != 5 {
		return d, errResult
	}
	decoder = json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&d); err != nil {
		return d, err
	}
	host, port, err := net.SplitHostPort(d.Endpoint)
	n, portErr := strconv.Atoi(port)
	id, idErr := hex.DecodeString(d.RunID)
	if d.Protocol != 1 || d.Generation == 0 || !filepath.IsAbs(d.Root) || filepath.Clean(d.Root) != d.Root || err != nil || host != "127.0.0.1" || portErr != nil || n < 1 || n > 65535 || strconv.Itoa(n) != port || idErr != nil || len(id) != 16 || hex.EncodeToString(id) != d.RunID {
		return d, errResult
	}
	return d, nil
}

func admissionPassword(root string) (string, error) {
	canonical, err := filepath.EvalSymlinks(root)
	if err != nil || canonical != root {
		return "", errIdentity
	}
	for _, name := range []string{"", "home", "tmp", "data", "security"} {
		info, err := os.Lstat(filepath.Join(root, name))
		if err != nil {
			return "", err
		}
		st, ok := info.Sys().(*syscall.Stat_t)
		if !ok || !info.IsDir() || info.Mode().Perm() != 0700 || st.Uid != uint32(os.Getuid()) {
			return "", errIdentity
		}
	}
	fd, err := syscall.Open(filepath.Join(root, "security", "password"), syscall.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_CLOEXEC|syscall.O_NONBLOCK, 0)
	if err != nil {
		return "", err
	}
	f := os.NewFile(uintptr(fd), "fixture password")
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return "", err
	}
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok || !info.Mode().IsRegular() || info.Mode().Perm() != 0600 || info.Size() != 64 || st.Uid != uint32(os.Getuid()) || st.Nlink != 1 {
		return "", errIdentity
	}
	raw, err := io.ReadAll(io.LimitReader(f, 65))
	decoded, decodeErr := hex.DecodeString(string(raw))
	if err != nil || decodeErr != nil || len(decoded) != 32 || hex.EncodeToString(decoded) != string(raw) {
		return "", errIdentity
	}
	return string(raw), nil
}

type admissionActor struct {
	s         *session
	role      string
	branch    string
	last      time.Time
	held      bool
	closed    bool
	mutations int
}

type admissionScenario struct {
	t           *testing.T
	ctx         context.Context
	input       admissionDescriptor
	target      endpoint
	actors      []*admissionActor
	previous    uint64
	sequence    int
	cleanupLeft *time.Duration
}

func (f *admissionScenario) event(a *admissionActor, operation string) {
	f.t.Helper()
	f.sequence++
	fmt.Printf("E1_OBSERVE run=%s generation=%d database=%s sequence=%d role=%s connection=%d operation=%q\n", f.input.RunID, f.input.Generation, f.target.base, f.sequence, a.role, a.s.id, operation)
}

func (f *admissionScenario) identity(a *admissionActor) {
	f.t.Helper()
	if a.closed {
		return
	}
	if time.Since(a.last) > 2*time.Second {
		f.t.Fatalf("holder/session liveness interval exceeded for %s; no continuity credit", a.role)
	}
	ctx, cancel := context.WithTimeout(f.ctx, 2*time.Second)
	defer cancel()
	row, err := a.s.one(ctx, "SELECT DATABASE(), active_branch(), CONNECTION_ID(), IS_USED_LOCK(?)", a.s.name)
	if err != nil || len(row) != 4 {
		f.t.Fatalf("identity %s: %v %v", a.role, row, err)
	}
	base, branch := splitDatabase(row[0])
	if base != f.target.base || (branch != "" && branch != a.branch) || row[1] != a.branch || row[2] != strconv.FormatUint(a.s.id, 10) {
		f.t.Fatalf("candidate binding %s: %v", a.role, row)
	}
	if a.held && row[3] != row[2] {
		f.t.Fatalf("lost holder %s: %v; no idle-drop success credit", a.role, row)
	}
	a.last = time.Now()
	f.event(a, "identity")
	f.t.Logf("identity database=%s role=%s values=%q", f.target.base, a.role, row)
}

// Identity samples use the same session, never race an in-flight command, and
// fail after a scheduler gap. They do not prove an unobserved time interval.
func (f *admissionScenario) before() {
	f.t.Helper()
	for _, a := range f.actors {
		f.identity(a)
	}
}
func (f *admissionScenario) one(a *admissionActor, q string, args ...any) []string {
	f.t.Helper()
	f.before()
	ctx, cancel := context.WithTimeout(f.ctx, 2*time.Second)
	defer cancel()
	row, err := a.s.one(ctx, q, args...)
	a.last = time.Now()
	if err != nil {
		f.t.Fatalf("%s %s: %v", a.role, q, err)
	}
	f.event(a, q)
	f.t.Logf("result database=%s role=%s query=%q values=%q", f.target.base, a.role, q, row)
	return row
}
func (f *admissionScenario) sql(a *admissionActor, q string, args ...any) {
	f.t.Helper()
	f.before()
	ctx, cancel := context.WithTimeout(f.ctx, 2*time.Second)
	defer cancel()
	_, err := a.s.consume(ctx, q, args...)
	a.last = time.Now()
	if err != nil {
		f.t.Fatalf("%s %s: %v", a.role, q, err)
	}
	f.event(a, q)
}
func (f *admissionScenario) scalar(a *admissionActor, q, want string, args ...any) {
	f.t.Helper()
	row := f.one(a, q, args...)
	if len(row) != 1 || row[0] != want {
		f.t.Fatalf("%s: got %v want %q", q, row, want)
	}
}

func (f *admissionScenario) add(role string, create bool) *admissionActor {
	f.t.Helper()
	f.before()
	live := 0
	for _, a := range f.actors {
		if !a.closed {
			live++
		}
	}
	if live >= 3 {
		f.t.Fatal("fixture exceeds three live roles")
	}
	ctx, cancel := context.WithTimeout(f.ctx, 2*time.Second)
	defer cancel()
	var s *session
	var err error
	if !create {
		s, err = open(ctx, f.target)
	} else {
		s = &session{target: f.target, name: lockName(f.target.base)}
		cfg, e := s.config(f.target)
		if e != nil {
			f.t.Fatal(e)
		}
		cfg.DBName = "" // Test-only CREATE DATABASE setup, never a production initializer.
		c, e := mysql.NewConnector(cfg)
		if e != nil {
			f.t.Fatal(e)
		}
		s.db = sql.OpenDB(&connector{Connector: c})
		s.db.SetMaxOpenConns(1)
		s.db.SetMaxIdleConns(0)
		s.conn, err = s.db.Conn(ctx)
		if err == nil {
			err = s.initialize(ctx)
		}
		if err == nil {
			_, err = s.consume(ctx, "CREATE DATABASE "+f.target.base)
		}
		if err == nil {
			_, err = s.consume(ctx, "USE "+f.target.base)
		}
	}
	if err != nil {
		if s != nil {
			f.t.Logf("failed open cleanup: %v", s.close())
		}
		f.t.Fatalf("candidate connection/initialization %s: %v", role, err)
	}
	a := &admissionActor{s: s, role: role, branch: "main", last: time.Now()}
	f.actors = append(f.actors, a)
	for _, other := range f.actors {
		if other != a && other.s.id == a.s.id {
			f.t.Fatal("connection ID reused")
		}
	}
	f.identity(a)
	if f.previous != 0 {
		f.waitAbsent(a, f.previous)
		f.previous = 0
	}
	return a
}
func (f *admissionScenario) acquire(a *admissionActor) {
	f.t.Helper()
	f.before()
	if err := a.s.acquire(f.ctx, 500*time.Millisecond); err != nil {
		f.t.Fatalf("acquire %s: %v", a.role, err)
	}
	a.last = time.Now()
	a.held = true
	f.identity(a)
	f.event(a, "acquire")
}
func (f *admissionScenario) busy(observer, holder *admissionActor) {
	f.t.Helper()
	f.scalar(observer, "SELECT GET_LOCK(?,0)", "0", holder.s.name)
	f.scalar(observer, "SELECT IS_USED_LOCK(?)", strconv.FormatUint(holder.s.id, 10), holder.s.name)
	f.identity(holder)
}
func (f *admissionScenario) release(a *admissionActor) {
	f.t.Helper()
	f.scalar(a, "SELECT RELEASE_LOCK(?)", "1", a.s.name)
	a.held = false
	f.scalar(a, "SELECT IS_USED_LOCK(?)", "\x00", a.s.name)
	// No actor polls/acquires concurrently with this checked readback. This raw
	// fixture helper retains the connection; production finish disposes instead.
	a.s.state = waiting
	f.event(a, "release-complete")
}
func (f *admissionScenario) tryMutate(a *admissionActor, q string) error {
	if !a.held || a.closed || a.s.state != held {
		return errState
	}
	a.mutations++
	f.sql(a, q)
	return nil
}
func (f *admissionScenario) mutate(a *admissionActor, q string) {
	f.t.Helper()
	if err := f.tryMutate(a, q); err != nil {
		f.t.Fatal(err)
	}
}
func (f *admissionScenario) waitAbsent(observer *admissionActor, id uint64) {
	f.t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for {
		row := f.one(observer, "SELECT COUNT(*) FROM information_schema.processlist WHERE ID=?", id)
		if len(row) != 1 {
			f.t.Fatal(row)
		}
		if row[0] == "0" {
			return
		}
		if time.Now().After(deadline) {
			f.t.Fatalf("connection %d still present", id)
		}
		select {
		case <-f.ctx.Done():
			f.t.Fatal(f.ctx.Err())
		case <-time.After(10 * time.Millisecond):
		}
	}
}
func (f *admissionScenario) closeActor(a *admissionActor) {
	f.t.Helper()
	if a.closed {
		return
	}
	err := a.s.close()
	a.closed = true
	if err != nil {
		f.t.Errorf("close %s: %v", a.role, err)
	}
}

var errAdmissionCleanupBudget = errors.New("admission fixture cleanup budget exhausted")

// Emergency closure never needs polling time. First break all owned transports,
// then dispose database/sql handles; errors remain visible on the failed test.
func (f *admissionScenario) closeRemainingActors() error {
	var result error
	for _, a := range f.actors {
		if !a.closed && a.s.transport != nil {
			result = errors.Join(result, a.s.transport.Close())
		}
	}
	for _, a := range f.actors {
		if !a.closed {
			result = errors.Join(result, a.s.close())
			a.closed = true
		}
	}
	return result
}
func (f *admissionScenario) exhaustedCleanup() error {
	if *f.cleanupLeft > 0 {
		return nil
	}
	return errors.Join(errAdmissionCleanupBudget, f.closeRemainingActors())
}
func (f *admissionScenario) cleanup() uint64 {
	f.t.Helper()
	start := time.Now()
	defer func() {
		if err := f.closeRemainingActors(); err != nil {
			f.t.Errorf("emergency owned cleanup: %v", err)
		}
	}()
	if err := f.exhaustedCleanup(); err != nil {
		f.t.Error(err)
		return 0
	}
	ctx, cancel := context.WithTimeout(context.Background(), *f.cleanupLeft)
	defer cancel()
	defer func() { *f.cleanupLeft -= time.Since(start) }()
	f.ctx = ctx
	var observer *admissionActor
	for _, a := range f.actors {
		if !a.closed && a.s.state != discarded {
			observer = a
			break
		}
	}
	for _, a := range f.actors {
		if a != observer {
			f.closeActor(a)
		}
	}
	if observer == nil {
		return 0
	}
	// Cleanup polling must not borrow the scenario's expired context. Unlike an
	// admission assertion it need not preserve the expired liveness interval.
	defer f.closeActor(observer)
	observer.last = time.Now()
	for _, a := range f.actors {
		if a != observer {
			f.waitAbsent(observer, a.s.id)
		}
	}
	id := observer.s.id
	f.closeActor(observer)
	return id
}
func (f *admissionScenario) setup() *admissionActor {
	f.t.Helper()
	a := f.add("setup", true)
	for _, q := range []string{"CREATE TABLE events (id INT PRIMARY KEY, value INT NOT NULL)", "CREATE TABLE fence (id INT PRIMARY KEY, value INT NOT NULL)", "CREATE TABLE renewal_probe (id INT PRIMARY KEY, value INT NOT NULL)", "INSERT INTO events VALUES (1,0),(2,0),(3,0)", "INSERT INTO fence VALUES (1,0)", "INSERT INTO renewal_probe VALUES (1,0)"} {
		f.sql(a, q)
	}
	return a
}

func admissionDrain(f *admissionScenario, w *admissionActor) {
	r := f.add("R", false)
	for _, commit := range []bool{true, false} {
		f.acquire(w)
		f.mutate(w, "START TRANSACTION")
		f.mutate(w, "UPDATE events SET value=value+1 WHERE id=1")
		f.busy(r, w)
		if commit {
			f.sql(w, "COMMIT")
		} else {
			f.sql(w, "ROLLBACK")
		}
		f.busy(r, w)
		f.release(w)
		f.acquire(r)
		f.sql(r, "ROLLBACK")
		f.scalar(r, "SELECT value FROM events WHERE id=1", "1")
		f.release(r)
	}
}
func admissionHold(f *admissionScenario, r *admissionActor) {
	w, n := f.add("W", false), f.add("renewal-prototype", false)
	f.acquire(r)
	f.sql(r, "START TRANSACTION")
	f.scalar(r, "SELECT value FROM fence WHERE id=1", "0")
	f.scalar(r, "SELECT value FROM events WHERE id=2", "0")
	f.busy(w, r)
	f.busy(n, r)
	if f.tryMutate(w, "START TRANSACTION") != errState || f.tryMutate(n, "START TRANSACTION") != errState {
		f.t.Fatal("fixture allowed BEGIN before enrollment")
	}
	if w.mutations != 0 || n.mutations != 0 {
		f.t.Fatal("mutation sent before enrollment")
	}
	f.sql(r, "ROLLBACK")
	f.release(r)
	for _, a := range []*admissionActor{w, n} {
		f.acquire(a)
		f.mutate(a, "START TRANSACTION")
		f.scalar(a, "SELECT value FROM fence WHERE id=1", "0")
		query := "UPDATE events SET value=1 WHERE id=2"
		if a == n {
			query = "UPDATE renewal_probe SET value=1 WHERE id=1"
		}
		f.mutate(a, query)
		f.sql(a, "COMMIT")
		f.sql(a, "ROLLBACK")
		read := "SELECT value FROM events WHERE id=2"
		if a == n {
			read = "SELECT value FROM renewal_probe WHERE id=1"
		}
		f.scalar(a, read, "1")
		f.release(a)
	}
	f.t.Log("renewal_probe is a fixture prototype, not a graph renewer or TTL qualification")
}
func admissionBypass(f *admissionScenario, r *admissionActor) {
	b := f.add("unenrolled-B", false)
	f.acquire(r)
	f.sql(r, "START TRANSACTION")
	f.scalar(r, "SELECT value FROM events WHERE id=1", "0")
	f.sql(b, "START TRANSACTION")
	f.sql(b, "UPDATE events SET value=9 WHERE id=1")
	f.sql(b, "COMMIT")
	f.identity(r)
	f.sql(b, "ROLLBACK")
	f.scalar(b, "SELECT value FROM events WHERE id=1", "9")
	f.sql(r, "ROLLBACK")
	f.release(r)
}
func admissionFence(f *admissionScenario, o *admissionActor) {
	writer, r := f.add("fence-F", false), f.add("R", false)
	f.sql(o, "START TRANSACTION")
	f.sql(o, "UPDATE events SET value=9 WHERE id=1")
	f.sql(writer, "START TRANSACTION")
	f.sql(writer, "UPDATE fence SET value=1 WHERE id=1")
	f.sql(writer, "COMMIT")
	f.acquire(r)
	f.sql(r, "START TRANSACTION")
	f.scalar(r, "SELECT value FROM fence WHERE id=1", "1")
	f.scalar(r, "SELECT value FROM events WHERE id=1", "0")
	f.sql(o, "COMMIT")
	f.identity(r)
	f.sql(writer, "ROLLBACK")
	f.scalar(writer, "SELECT value FROM events WHERE id=1", "9")
	f.sql(r, "ROLLBACK")
	f.release(r)
}

func (f *admissionScenario) checkout(a *admissionActor, branch string, create bool) {
	f.t.Helper()
	switch branch {
	case "main", "warm", "incoming":
	default:
		f.t.Fatal("unrecognized fixture branch")
	}
	if create {
		f.sql(a, "CALL DOLT_CHECKOUT('-b',?)", branch)
	} else {
		f.sql(a, "CALL DOLT_CHECKOUT(?)", branch)
	}
	a.branch = branch
	f.identity(a)
}

func (a *admissionActor) executeMerge(ctx context.Context, operand string) ([]resultSet, error) {
	sets, err := a.s.execute(ctx, mergeOperation, operand)
	a.last = time.Now()
	if err != nil {
		// A canceled pre-dispatch call still owns its transport. Only the
		// session's actual disposal state permits cleanup to skip this actor.
		a.closed = a.s.state == discarded
	}
	return sets, err
}

func admissionMerge(f *admissionScenario, setup *admissionActor) {
	// Setup touches only the disposable database. Scratch fast-forward warms
	// engine work without changing candidate transport or timeout semantics.
	f.sql(setup, "CALL DOLT_ADD('-A')")
	f.sql(setup, "CALL DOLT_COMMIT('-m','admission base','--author','Fixture <fixture@example.invalid>')")
	f.checkout(setup, "warm", true)
	f.sql(setup, "UPDATE events SET value=1 WHERE id=3")
	f.sql(setup, "CALL DOLT_COMMIT('-am','warm','--author','Fixture <fixture@example.invalid>')")
	warm := f.one(setup, "SELECT HASHOF('HEAD')")
	if len(warm) != 1 || !validCommitOperand(warm[0]) {
		f.t.Fatal("invalid warm hash")
	}
	f.checkout(setup, "main", false)
	f.sql(setup, "CALL DOLT_MERGE(?)", warm[0])
	f.sql(setup, "ROLLBACK")
	base := f.one(setup, "SELECT HASHOF('HEAD')")
	f.checkout(setup, "incoming", true)
	f.sql(setup, "UPDATE events SET value=7 WHERE id=1")
	f.sql(setup, "CALL DOLT_COMMIT('-am','incoming','--author','Fixture <fixture@example.invalid>')")
	incoming := f.one(setup, "SELECT HASHOF('HEAD')")
	if len(base) != 1 || len(incoming) != 1 || !validCommitOperand(incoming[0]) || base[0] == incoming[0] {
		f.t.Fatal("invalid measured merge ancestry setup")
	}
	f.checkout(setup, "main", false)
	f.sql(setup, "ROLLBACK")
	r := f.add("replication-candidate", false)
	f.acquire(r)
	f.scalar(r, "SELECT HASHOF('HEAD')", base[0])
	f.busy(setup, r)
	f.before()
	ctx, cancel := context.WithTimeout(f.ctx, 2*time.Second)
	sets, err := r.executeMerge(ctx, incoming[0])
	cancel()
	if err != nil {
		if errors.Is(err, errUncertain) {
			f.t.Fatalf("uncertain merge; no Phase A terminality classification or replay: %v", err)
		}
		f.t.Fatalf("merge candidate refused: state=%v error=%v", r.s.state, err)
	}
	row := []string{incoming[0], "1", "0", "merge successful"}
	found := false
	for _, set := range sets {
		if len(set.columns) == 0 && len(set.rows) == 0 {
			continue
		}
		if found || strings.Join(set.columns, ",") != "hash,fast_forward,conflicts,message" || len(set.rows) != 1 || len(set.rows[0]) != 4 {
			f.t.Fatalf("unexpected merge result %#v", sets)
		}
		for i, v := range set.rows[0] {
			if !v.Valid || v.String != row[i] {
				f.t.Fatalf("unexpected merge result %#v", sets)
			}
		}
		found = true
	}
	if !found {
		f.t.Fatal("missing merge result")
	}
	f.t.Logf("exact fast-forward result: %#v", sets)
	f.scalar(r, "SELECT HASHOF('HEAD')", incoming[0])
	f.scalar(r, "SELECT value FROM events WHERE id=1", "7")
	f.sql(setup, "ROLLBACK")
	f.scalar(setup, "SELECT HASHOF('HEAD')", incoming[0])
	f.scalar(setup, "SELECT value FROM events WHERE id=1", "7")
	f.busy(setup, r)
	f.before()
	if err := r.s.refresh(f.ctx); err != nil {
		f.t.Fatal(err)
	}
	r.last = time.Now()
	f.identity(r)
	if err := r.s.finish(f.ctx); err != nil {
		f.t.Fatal(err)
	}
	r.closed = true
	r.held = false
	f.acquire(setup)
	f.release(setup)
	f.t.Logf("boundary continuity only: base=%s incoming=%s connection=%d; no in-flight terminality claim", base[0], incoming[0], r.s.id)
}

// Every terminal cause must be the exact expected sentinel. errors.Is alone
// would accept a single wrapper around a join containing an unexpected cause.
func admissionOnlyError(err, target error) bool {
	if err == nil {
		return false
	}
	if list, ok := err.(interface{ Unwrap() []error }); ok {
		parts := list.Unwrap()
		if len(parts) == 0 {
			return false
		}
		for _, part := range parts {
			if !admissionOnlyError(part, target) {
				return false
			}
		}
		return true
	}
	if single, ok := err.(interface{ Unwrap() error }); ok {
		return admissionOnlyError(single.Unwrap(), target)
	}
	return err == target
}
func admissionDisconnect(f *admissionScenario, w *admissionActor) {
	f.acquire(w)
	busy := f.add("bounded-refusal", false)
	f.before()
	err := busy.s.acquire(f.ctx, 500*time.Millisecond)
	busy.closed = busy.s.state == discarded
	if !busy.closed || !admissionOnlyError(err, errBusy) {
		f.t.Fatalf("busy result/cleanup must be only errBusy: %v", err)
	}
	f.identity(w)
	f.waitAbsent(w, busy.s.id)
	c := f.add("EOF-observer-C", false)
	f.busy(c, w)
	f.identity(w)
	if err := w.s.transport.Close(); err != nil {
		f.t.Fatal(err)
	}
	// Raw transport close is intentional. Driver cleanup may attempt COM_QUIT
	// on this already-closed connection; only net.ErrClosed is expected here.
	err = w.s.close()
	w.closed = true
	if err != nil && !admissionExpectedClosed(err) {
		f.t.Fatalf("unexpected raw EOF cleanup error: %v", err)
	}
	f.t.Logf("intentional raw transport EOF close result: %v", err)
	ctx, cancel := context.WithTimeout(f.ctx, 2*time.Second)
	defer cancel()
	if err := c.s.acquire(ctx, 2*time.Second); err != nil {
		f.t.Fatal(err)
	}
	c.last = time.Now()
	c.held = true
	f.identity(c)
	f.waitAbsent(c, w.s.id)
	f.release(c)
	f.t.Log("idle cleanup only; absence/acquisition is not an in-flight merge outcome")
}
func admissionExpectedClosed(err error) bool {
	return err == nil || admissionOnlyError(err, net.ErrClosed)
}

func TestAdmissionEngineWorker(t *testing.T) {
	if os.Getenv("GRAPH_E1_WORKER") != "1" {
		t.Skip("parent-owned admission worker")
	}
	if os.Getenv("GRAPH_E1_CONTROL") != "" || os.Getenv("BEADS_TEST_E1_ADMISSION_FIXTURE") != "" {
		t.Fatal("polluted worker environment")
	}
	raw, err := io.ReadAll(io.LimitReader(os.Stdin, 4097))
	if err != nil {
		t.Fatal(err)
	}
	input, err := parseAdmissionDescriptor(raw)
	if err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]string{"HOME": filepath.Join(input.Root, "home"), "TMPDIR": filepath.Join(input.Root, "tmp"), "TMP": filepath.Join(input.Root, "tmp"), "TEMP": filepath.Join(input.Root, "tmp"), "DOLT_METRICS_DISABLED": "1", "DOLT_DISABLE_EVENT_FLUSH": "1", "TZ": "UTC"} {
		if os.Getenv(key) != want {
			t.Fatalf("worker isolation %s", key)
		}
	}
	password, err := admissionPassword(input.Root)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 210*time.Second)
	defer cancel()
	scenarios := []func(*admissionScenario, *admissionActor){admissionDrain, admissionHold, admissionBypass, admissionFence, admissionMerge, admissionDisconnect}
	cleanupLeft := 20 * time.Second
	setupLeft := 60 * time.Second
	var previous uint64
	for i, scenario := range scenarios {
		passed := t.Run(strconv.Itoa(i+1), func(t *testing.T) {
			bound := 20 * time.Second
			if i == 4 {
				bound = 30 * time.Second
			}
			scenarioCtx, stop := context.WithTimeout(ctx, bound)
			defer stop()
			f := &admissionScenario{t: t, ctx: scenarioCtx, input: input, target: endpoint{address: input.Endpoint, user: "fixture", password: password, base: fmt.Sprintf("e1_fixture_%d", i+1), branch: "main"}, previous: previous, cleanupLeft: &cleanupLeft}
			t.Cleanup(func() { previous = f.cleanup() })
			start := time.Now()
			a := f.setup()
			setupLeft -= time.Since(start)
			if setupLeft <= 0 {
				t.Fatal("setup total budget exhausted")
			}
			scenario(f, a)
		})
		if !passed {
			t.Fatal("admission scenario failed; stopping without aggregate receipt")
		}
		fmt.Printf("E1_RESULT %s %d %d pass\n", input.RunID, input.Generation, i+1)
	}
	fmt.Printf("E1_RESULT %s %d done 6\n", input.RunID, input.Generation)
}

func TestAdmissionDescriptor(t *testing.T) {
	good := `{"protocol":1,"root":"/private/tmp/fixture","endpoint":"127.0.0.1:1234","runId":"` + strings.Repeat("a", 32) + `","generation":1}`
	if _, err := parseAdmissionDescriptor([]byte(good)); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{good + "{}", strings.Replace(good, `"protocol":1`, `"protocol":1,"protocol":1`, 1), strings.Replace(good, `"generation":1`, `"generation":0`, 1), strings.Replace(good, `"generation":1`, `"generation":null`, 1), strings.Replace(good, "127.0.0.1", "localhost", 1), strings.Replace(good, "1234", "01234", 1), strings.Replace(good, `"root"`, `"unknown"`, 1)} {
		if _, err := parseAdmissionDescriptor([]byte(bad)); err == nil {
			t.Fatalf("accepted %s", bad)
		}
	}
}

// These controls exercise refusal before any SQL boundary is present.
func TestAdmissionWriterGuard(t *testing.T) {
	for _, actor := range []*admissionActor{{s: &session{state: waiting}}, {s: &session{state: held}, closed: true, held: true}, {s: &session{state: discarded}, held: true}} {
		f := &admissionScenario{}
		if f.tryMutate(actor, "START TRANSACTION") != errState || actor.mutations != 0 {
			t.Fatal("unadmitted write reached dispatch")
		}
	}
}
func TestAdmissionCloseErrorClassification(t *testing.T) {
	unexpected := errors.New("unexpected cleanup failure")
	nested := fmt.Errorf("wrapped: %w", errors.Join(net.ErrClosed, unexpected))
	if admissionExpectedClosed(nested) {
		t.Fatal("wrapped join swallowed unexpected cleanup cause")
	}
	if !admissionExpectedClosed(fmt.Errorf("wrapped: %w", errors.Join(net.ErrClosed, fmt.Errorf("nested: %w", net.ErrClosed)))) {
		t.Fatal("expected wrapped close causes refused")
	}
	if !admissionOnlyError(errors.Join(errBusy), errBusy) || admissionOnlyError(errors.Join(errBusy, unexpected), errBusy) {
		t.Fatal("busy cleanup classification")
	}
	if !admissionExpectedClosed(errors.Join(net.ErrClosed)) || admissionExpectedClosed(errors.Join(net.ErrClosed, unexpected)) {
		t.Fatal("closed cleanup classification")
	}
}

type admissionCloseFailure struct {
	net.Conn
	failure error
}

func (c admissionCloseFailure) Close() error { return errors.Join(c.Conn.Close(), c.failure) }

func TestAdmissionExhaustedCleanup(t *testing.T) {
	for _, withFailure := range []bool{false, true} {
		t.Run(strconv.FormatBool(withFailure), func(t *testing.T) {
			budget := time.Duration(0)
			f := &admissionScenario{cleanupLeft: &budget}
			unexpected := errors.New("owned transport close failure")
			var peers []net.Conn
			for i := 0; i < 3; i++ {
				owned, peer := net.Pipe()
				peers = append(peers, peer)
				t.Cleanup(func() { _ = owned.Close(); _ = peer.Close() })
				var conn net.Conn = owned
				if withFailure && i == 0 {
					conn = admissionCloseFailure{Conn: owned, failure: unexpected}
				}
				f.actors = append(f.actors, &admissionActor{s: &session{transport: &transport{Conn: conn}}})
			}
			err := f.exhaustedCleanup()
			if !errors.Is(err, errAdmissionCleanupBudget) || errors.Is(err, unexpected) != withFailure {
				t.Fatalf("cleanup lost failure: %v", err)
			}
			for i, peer := range peers {
				if !f.actors[i].closed || f.actors[i].s.state != discarded {
					t.Fatal("actor not disposed after budget exhaustion")
				}
				if err := peer.SetReadDeadline(time.Now().Add(time.Second)); err != nil && !errors.Is(err, io.ErrClosedPipe) {
					t.Fatal(err)
				}
				var b [1]byte
				if _, err := peer.Read(b[:]); err != io.EOF {
					t.Fatalf("owned transport survived: %v", err)
				}
			}
		})
	}
}

// There is deliberately no database/sql Conn: reaching dispatch would panic.
// A canceled admission must remain owned until the fixture closes transport.
func TestAdmissionCanceledMergeCleanup(t *testing.T) {
	owned, peer := net.Pipe()
	t.Cleanup(func() { _ = owned.Close(); _ = peer.Close() })
	a := &admissionActor{s: &session{state: held, transport: &transport{Conn: owned}}, held: true}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	sets, err := a.executeMerge(ctx, strings.Repeat("a", 32))
	if err != context.Canceled || len(sets) != 0 || a.s.state != held {
		t.Fatalf("expected refusal before dispatch, state=%v err=%v", a.s.state, err)
	}
	budget := time.Duration(0)
	f := &admissionScenario{actors: []*admissionActor{a}, cleanupLeft: &budget}
	if err := f.exhaustedCleanup(); !errors.Is(err, errAdmissionCleanupBudget) {
		t.Fatal(err)
	}
	if err := peer.SetReadDeadline(time.Now().Add(100 * time.Millisecond)); err != nil && !errors.Is(err, io.ErrClosedPipe) {
		t.Fatal(err)
	}
	var b [1]byte
	if _, err := peer.Read(b[:]); err != io.EOF {
		t.Fatalf("transport survived canceled no-dispatch merge: %v", err)
	}
	if !a.closed || a.s.state != discarded {
		t.Fatal("owned candidate not disposed")
	}
}
