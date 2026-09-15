package graphsession

import (
	"context"
	"database/sql"
	"errors"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
)

func TestFreshLifecycle(t *testing.T) {
	s, p := openPeer(t, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := s.acquire(ctx, ioBound); err != nil {
		t.Fatal(err)
	}
	result, err := s.execute(ctx, mergeOperation, testCommit)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || len(result[0].rows) != 1 {
		t.Fatalf("result lost: %#v", result)
	}
	if err := s.refresh(ctx); err != nil {
		t.Fatal(err)
	}
	if err := s.finish(ctx); err != nil {
		t.Fatal(err)
	}
	if err := s.close(); err != nil {
		t.Fatal(err)
	}
	if s.db.Stats().OpenConnections != 0 {
		t.Fatal("private pool not closed")
	}
	commands := p.commands()
	want := []string{"SELECT @@autocommit", "SELECT GET_LOCK", "ROLLBACK", "SELECT DATABASE", "CALL DOLT_MERGE", "ROLLBACK", "SELECT DATABASE", "SELECT RELEASE_LOCK", "SELECT IS_USED_LOCK"}
	if len(commands) != len(want) {
		t.Fatalf("commands: %q", commands)
	}
	for i, prefix := range want {
		if !strings.HasPrefix(commands[i], prefix) {
			t.Fatalf("command %d: %q", i, commands[i])
		}
	}
	if commands[4] != "CALL DOLT_MERGE('"+testCommit+"')" {
		t.Fatalf("immutable operand changed: %q", commands[4])
	}
}
func TestPredispatchAbortReleasesOnce(t *testing.T) {
	for range 3 {
		s, p := openPeer(t, nil)
		ctx := context.Background()
		if err := s.acquire(ctx, ioBound); err != nil {
			t.Fatal(err)
		}
		if err := s.abort(ctx); err != nil {
			t.Fatal(err)
		}
		if err := s.abort(ctx); !errors.Is(err, errState) {
			t.Fatalf("second abort: %v", err)
		}
		count := 0
		for _, q := range p.commands() {
			if strings.HasPrefix(q, "SELECT RELEASE_LOCK") {
				count++
			}
			if strings.HasPrefix(q, "CALL") {
				t.Fatal("abort dispatched")
			}
		}
		if count != 1 {
			t.Fatalf("release count %d", count)
		}
	}
}
func TestAcquireRefusalAndCancellation(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value string
		want  error
	}{{"malformed", "2", errResult}, {"ambiguous", "", errResult}, {"busy", "0", errBusy}} {
		t.Run(tc.name, func(t *testing.T) {
			s, p := openPeer(t, func(q string) reply {
				if strings.HasPrefix(q, "SELECT GET_LOCK") {
					return reply{rows: [][]string{{tc.value, "41"}}}
				}
				return defaultReply(q)
			})
			if err := s.acquire(context.Background(), 70*time.Millisecond); !errors.Is(err, tc.want) {
				t.Fatalf("got %v", err)
			}
			for _, q := range p.commands() {
				if strings.Contains(q, "RELEASE") {
					t.Fatal("ambiguous/busy release")
				}
			}
			if s.db.Stats().OpenConnections != 0 {
				t.Fatal("pool leaked")
			}
		})
	}
	s, _ := openPeer(t, nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := s.acquire(ctx, ioBound); !errors.Is(err, context.Canceled) || errors.Is(err, errBusy) {
		t.Fatalf("caller cause: %v", err)
	}
}
func TestFreshSettingsAndIdentityRefuseWithoutRepair(t *testing.T) {
	for _, bad := range []string{"autocommit", "disabled", "branch", "session", "owner"} {
		t.Run(bad, func(t *testing.T) {
			e, p := startPeer(t, func(q string) reply {
				r := defaultReply(q)
				if strings.HasPrefix(q, "SELECT @@autocommit") {
					if bad == "autocommit" {
						r.rows[0][0] = "0"
					}
					if bad == "disabled" {
						r.rows[0][1] = "1"
					}
				}
				if strings.HasPrefix(q, "SELECT DATABASE") {
					switch bad {
					case "branch":
						r.rows[0][1] = "other"
					case "session":
						r.rows[0][2] = "42"
					case "owner":
						r.rows[0][3] = "42"
					}
				}
				return r
			})
			s, err := open(context.Background(), e)
			if err == nil {
				err = s.acquire(context.Background(), ioBound)
				_ = s.close()
			}
			if err == nil {
				t.Fatal("bad identity/settings admitted")
			}
			for _, q := range p.commands() {
				if strings.HasPrefix(q, "SET ") || strings.HasPrefix(q, "START ") || strings.Contains(q, "CHECKOUT") {
					t.Fatalf("session repair %q", q)
				}
			}
		})
	}
}
func TestDispatchedFailureNeverReleasesOrReplays(t *testing.T) {
	for _, drop := range []bool{false, true} {
		t.Run(map[bool]string{false: "server error", true: "lost reply"}[drop], func(t *testing.T) {
			s, p := openPeer(t, func(q string) reply {
				if q == "COMMIT" {
					return reply{err: !drop, drop: drop}
				}
				return defaultReply(q)
			})
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			if err := s.acquire(ctx, ioBound); err != nil {
				t.Fatal(err)
			}
			_, err := s.execute(ctx, commitOperation, "")
			if !errors.Is(err, errUncertain) {
				t.Fatalf("expected uncertainty: %v", err)
			}
			if !drop {
				var mysqlErr *mysql.MySQLError
				if !errors.As(err, &mysqlErr) {
					t.Fatalf("lost original error: %v", err)
				}
			}
			if _, err := s.execute(ctx, commitOperation, ""); !errors.Is(err, errState) {
				t.Fatal("replay admitted")
			}
			commits := 0
			for _, q := range p.commands() {
				if q == "COMMIT" {
					commits++
				}
				if strings.Contains(q, "RELEASE") {
					t.Fatal("uncertain release")
				}
			}
			if commits != 1 {
				t.Fatalf("commits %d", commits)
			}
		})
	}
}
func TestReleaseFailureIsNotRetried(t *testing.T) {
	s, p := openPeer(t, func(q string) reply {
		if strings.HasPrefix(q, "SELECT RELEASE_LOCK") {
			return reply{err: true}
		}
		return defaultReply(q)
	})
	if err := s.acquire(context.Background(), ioBound); err != nil {
		t.Fatal(err)
	}
	err := s.abort(context.Background())
	var mysqlErr *mysql.MySQLError
	if !errors.As(err, &mysqlErr) {
		t.Fatalf("lost release failure: %v", err)
	}
	if s.db.Stats().OpenConnections != 0 {
		t.Fatal("not discarded")
	}
	n := 0
	for _, q := range p.commands() {
		if strings.Contains(q, "RELEASE_LOCK") {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("release attempts %d", n)
	}
}

type closeObserved struct {
	onCommit func()
	net.Conn
	once   sync.Once
	closed chan struct{}
}

func (c *closeObserved) Write(b []byte) (int, error) {
	if len(b) > 5 && b[4] == 3 && string(b[5:]) == "COMMIT" && c.onCommit != nil {
		c.onCommit()
	}
	return c.Conn.Write(b)
}

func (c *closeObserved) Close() error {
	err := c.Conn.Close()
	c.once.Do(func() { close(c.closed) })
	return err
}

// net.Pipe is a real deadline-aware transport. The peer intentionally stops
// reading after a successful COMMIT, making the actual driver's COM_QUIT write
// wait for WriteTimeout. The production config/cleanup paths remain in use.
func TestQuitWriteBoundAndOwnedCleanup(t *testing.T) {
	for _, canceled := range []bool{false, true} {
		t.Run(map[bool]string{false: "successful command", true: "canceled neighbor"}[canceled], func(t *testing.T) {
			client, server := net.Pipe()
			closed := make(chan struct{})
			observed := &closeObserved{Conn: client, closed: closed}
			p := &peer{}
			done := make(chan struct{})
			go func() { defer close(done); defer server.Close(); _ = p.serve(server, closed) }()
			s := &session{target: endpoint{base: "sales", branch: "main"}, name: lockName("sales")}
			dispatchedBeforeWrite := false
			observed.onCommit = func() { dispatchedBeforeWrite = s.state == dispatched }

			cfg := s.config(s.target)
			// Test-only transport substitution; no injectable factory exists in open.
			cfg.DialFunc = func(context.Context, string, string) (net.Conn, error) {
				s.transport = &transport{Conn: observed}
				return s.transport, nil
			}
			c, err := mysql.NewConnector(cfg)
			if err != nil {
				t.Fatal(err)
			}
			s.db = sql.OpenDB(&connector{Connector: c})
			s.db.SetMaxOpenConns(1)
			s.db.SetMaxIdleConns(0)
			t.Cleanup(func() {
				_ = s.close()
				_ = client.Close()
				_ = server.Close()
				select {
				case <-done:
				case <-time.After(time.Second):
					t.Error("peer not joined")
				}
			})
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			s.conn, err = s.db.Conn(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if err = s.initialize(ctx); err != nil {
				t.Fatal(err)
			}
			if err = s.acquire(ctx, ioBound); err != nil {
				t.Fatal(err)
			}
			if _, err = s.execute(ctx, commitOperation, ""); err != nil {
				t.Fatal(err)
			}
			if !dispatchedBeforeWrite {
				t.Fatal("mutation write preceded dispatched state")
			}
			var primary error
			if canceled {
				primary = context.Canceled
			}
			before := time.Now()
			err = s.dispose(primary)
			elapsed := time.Since(before)
			var ne net.Error
			if !errors.As(err, &ne) || !ne.Timeout() {
				t.Fatalf("missing close timeout: %v", err)
			}
			if canceled && !errors.Is(err, context.Canceled) {
				t.Fatalf("lost primary: %v", err)
			}
			if elapsed < writeBound/2 || elapsed > ioBound {
				t.Fatalf("quit write duration %v", elapsed)
			}
			select {
			case <-closed:
			default:
				t.Fatal("owned transport open")
			}
			select {
			case <-done:
			case <-time.After(time.Second):
				t.Fatal("peer still running")
			}
			if s.db.Stats().OpenConnections != 0 {
				t.Fatal("private DB open")
			}
			if second := s.close(); !errors.As(second, &ne) {
				t.Fatalf("repeat lost cleanup result: %v", second)
			}
			t.Logf("COM_QUIT bound=%v observed=%v; transport closed, fixture joined, pool=0", writeBound, elapsed)
		})
	}
}
func TestIdentityAndClosedConfiguration(t *testing.T) {
	e := endpoint{address: "127.0.0.1:1", base: "sales", branch: "main"}
	if err := e.validate(); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"Sales/main", "SALES@main", "sales/main"} {
		b, r := splitDatabase(name)
		if b != "sales" || r != "main" {
			t.Fatal(name)
		}
	}
	if lockName("sales") != "bd.graph.e1:GymvA02rglnq-whYCUnQW3ODtznlKQIZzNGzShJFVHY" || len(lockName("sales")) != 55 || lockName("sales") == lockName("other") {
		t.Fatal("lock identity")
	}
	for _, address := range []string{"example.com:1", "192.0.2.1:1", "127.0.0.1:0"} {
		bad := e
		bad.address = address
		if bad.validate() == nil {
			t.Fatal(address)
		}
	}
	s := &session{}
	cfg := s.config(e)
	if cfg.WriteTimeout != writeBound || cfg.WriteTimeout <= 0 || cfg.Params != nil || cfg.MultiStatements || cfg.AllowAllFiles || cfg.MaxAllowedPacket <= 0 || cfg.DialFunc == nil {
		t.Fatal("unclosed config")
	}
}

func TestCompleteResultsAndDrainErrors(t *testing.T) {
	for _, scenario := range []string{"null", "multiple sets", "iteration failure", "limit and close failure"} {
		t.Run(scenario, func(t *testing.T) {
			s, _ := openPeer(t, func(q string) reply {
				if q != "CALL DOLT_MERGE('"+testCommit+"')" {
					return defaultReply(q)
				}
				switch scenario {
				case "null":
					return reply{rows: [][]string{{""}}, null: true}
				case "multiple sets":
					return reply{rows: [][]string{{"first"}}, multi: true}
				case "iteration failure":
					return reply{rows: [][]string{{"retained"}}, tailError: true}
				default:
					r := reply{tailError: true}
					for range 257 {
						r.rows = append(r.rows, []string{"bounded"})
					}
					return r
				}
			})
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			if err := s.acquire(ctx, ioBound); err != nil {
				t.Fatal(err)
			}
			sets, err := s.execute(ctx, mergeOperation, testCommit)
			switch scenario {
			case "null":
				if err != nil || sets[0].rows[0][0].Valid {
					t.Fatalf("NULL lost: %#v %v", sets, err)
				}
			case "multiple sets":
				if err != nil || len(sets) != 2 {
					t.Fatalf("sets not drained: %#v %v", sets, err)
				}
			default:
				if !errors.Is(err, errUncertain) {
					t.Fatalf("lost uncertainty: %v", err)
				}
				var mysqlErr *mysql.MySQLError
				if scenario == "iteration failure" && !errors.As(err, &mysqlErr) {
					t.Fatalf("lost observed iteration error: %v", err)
				}
				if len(sets) == 0 || len(sets[0].rows) == 0 {
					t.Fatal("complete partial rows not retained")
				}
				if scenario == "limit and close failure" && !errors.Is(err, errResult) {
					t.Fatalf("lost limit error: %v", err)
				}
			}
		})
	}
}

func TestNoReconnectOrGlobalDialer(t *testing.T) {
	mysql.RegisterDialContext("tcp", func(context.Context, string) (net.Conn, error) { return nil, errors.New("global dialer must not run") })
	t.Cleanup(func() { mysql.DeregisterDialContext("tcp") })
	e, _ := startPeer(t, nil)
	s := &session{}
	cfg := s.config(e)
	mc, err := mysql.NewConnector(cfg)
	if err != nil {
		t.Fatal(err)
	}
	c := &connector{Connector: mc}
	ctx, cancel := context.WithTimeout(context.Background(), ioBound)
	defer cancel()
	conn, err := c.Connect(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := conn.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Connect(ctx); !errors.Is(err, errReconnect) {
		t.Fatalf("second Connect: %v", err)
	}
	// Each new owner still constructs its own new physical transport.
	a, err := open(ctx, e)
	if err != nil {
		t.Fatal(err)
	}
	defer a.close()
	b, err := open(ctx, e)
	if err != nil {
		t.Fatal(err)
	}
	defer b.close()
	if a.transport == b.transport || a.transport.Conn == b.transport.Conn {
		t.Fatal("borrowed/reused transport")
	}
}

func TestOverlappingUseAndCanceledOwner(t *testing.T) {
	entered := make(chan struct{})
	resume := make(chan struct{})
	var resumeOnce sync.Once
	releasePeer := func() { resumeOnce.Do(func() { close(resume) }) }
	defer releasePeer()
	var once sync.Once
	s, _ := openPeer(t, func(q string) reply {
		if strings.HasPrefix(q, "SELECT GET_LOCK") {
			once.Do(func() { close(entered) })
			<-resume
		}
		return defaultReply(q)
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- s.acquire(ctx, ioBound) }()
	select {
	case <-entered:
	case <-time.After(time.Second):
		releasePeer()
		t.Fatal("query did not start")
	}
	if err := s.close(); !errors.Is(err, errState) {
		t.Fatalf("overlap: %v", err)
	}
	cancel()
	releasePeer()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("canceled owner: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("owned method not joined")
	}
	if s.db.Stats().OpenConnections != 0 {
		t.Fatal("canceled owner leaked transport")
	}
}

func TestReleaseReadbackMustBeValidSuccessor(t *testing.T) {
	for _, value := range []string{"41", "nonsense", "0", "\x00"} {
		t.Run(value, func(t *testing.T) {
			s, _ := openPeer(t, func(q string) reply {
				if strings.HasPrefix(q, "SELECT IS_USED_LOCK") {
					return reply{rows: [][]string{{value}}}
				}
				return defaultReply(q)
			})
			if err := s.acquire(context.Background(), ioBound); err != nil {
				t.Fatal(err)
			}
			if err := s.abort(context.Background()); !errors.Is(err, errResult) {
				t.Fatalf("readback accepted: %v", err)
			}
		})
	}
}

func TestReleaseNullOwner(t *testing.T) {
	s, _ := openPeer(t, func(q string) reply {
		if strings.HasPrefix(q, "SELECT IS_USED_LOCK") {
			return reply{rows: [][]string{{""}}, null: true}
		}
		return defaultReply(q)
	})
	if err := s.acquire(context.Background(), ioBound); err != nil {
		t.Fatal(err)
	}
	if err := s.abort(context.Background()); err != nil {
		t.Fatal(err)
	}
}
