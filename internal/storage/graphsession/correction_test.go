package graphsession

import (
	"context"
	"database/sql"
	"errors"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
)

const testCommit = "0123456789abcdefghijklmnopqrstuv"

// Real local protocol I/O, with a controlled continuing stream and no engine.
func pipeSession(t *testing.T, response reply) (*session, *peer, <-chan struct{}) {
	t.Helper()
	client, server := net.Pipe()
	p := &peer{fn: func(q string) reply {
		if q == "COMMIT" {
			return response
		}
		return defaultReply(q)
	}}
	done := make(chan struct{})
	go func() { defer close(done); defer server.Close(); _ = p.serve(server, nil) }()
	s := &session{target: endpoint{base: "sales", branch: "main"}, name: lockName("sales")}
	cfg := s.config(s.target)
	cfg.DialFunc = func(context.Context, string, string) (net.Conn, error) {
		s.transport = &transport{Conn: client}
		return s.transport, nil
	}
	connectorValue, err := mysql.NewConnector(cfg)
	if err != nil {
		t.Fatal(err)
	}
	s.db = sql.OpenDB(&connector{Connector: connectorValue})
	s.db.SetMaxOpenConns(1)
	s.db.SetMaxIdleConns(0)
	t.Cleanup(func() {
		_ = client.Close()
		_ = server.Close()
		_ = s.close()
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Error("pipe peer not joined")
		}
	})
	ctx, cancel := context.WithTimeout(context.Background(), ioBound)
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
	return s, p, done
}

func TestContinuingDrain(t *testing.T) {
	for _, scenario := range []string{"local overflow", "implicit close cancellation", "implicit close deadline"} {
		t.Run(scenario, func(t *testing.T) {
			entered := make(chan struct{})
			response := reply{stream: true, streamEntered: entered, rows: [][]string{{"1"}}}
			if scenario == "local overflow" {
				response.rows = nil
				for range 257 {
					response.rows = append(response.rows, []string{"1"})
				}
			} else {
				response.integer = true
				response.malformedRow = true
			}
			s, p, peerDone := pipeSession(t, response)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			if scenario == "implicit close deadline" {
				cancel()
				ctx, cancel = context.WithTimeout(context.Background(), 250*time.Millisecond)
			}
			defer cancel()
			completed := make(chan error, 1)
			go func() { _, err := s.execute(ctx, commitOperation, ""); completed <- err }()
			// Always close the actual transport and join the caller on an assertion failure.
			joined := false
			defer func() {
				_ = s.transport.Close()
				if !joined {
					select {
					case <-completed:
					case <-time.After(time.Second):
						t.Error("call not joined during cleanup")
					}
				}
			}()
			if scenario != "local overflow" {
				select {
				case <-entered:
				case <-time.After(time.Second):
					t.Fatal("implicit Close did not consume the continuing stream")
				}
				if scenario == "implicit close cancellation" {
					cancel()
				}
			}
			var err error
			select {
			case err = <-completed:
				joined = true
			case <-time.After(time.Second):
				t.Fatal("result drain outlived cancellation/cleanup bound")
			}
			if !errors.Is(err, errUncertain) {
				t.Fatalf("uncertainty lost: %v", err)
			}
			if scenario == "local overflow" {
				if !errors.Is(err, errResult) {
					t.Fatalf("local refusal lost: %v", err)
				}
			} else {
				cause := context.Canceled
				if scenario == "implicit close deadline" {
					cause = context.DeadlineExceeded
				}
				if !errors.Is(err, cause) {
					t.Fatalf("context lost: %v", err)
				}
				var conversion *strconv.NumError
				if !errors.As(err, &conversion) {
					t.Fatalf("observed conversion failure lost: %v", err)
				}
			}
			if s.db.Stats().OpenConnections != 0 {
				t.Fatal("pool not disposed")
			}
			select {
			case <-peerDone:
			case <-time.After(time.Second):
				t.Fatal("peer not joined")
			}
			for _, q := range p.commands() {
				if q == "SELECT RELEASE_LOCK('"+s.name+"')" {
					t.Fatal("uncertain release")
				}
			}
			if _, err = s.execute(ctx, commitOperation, ""); !errors.Is(err, errState) {
				t.Fatal("replay admitted")
			}
		})
	}
}

func TestImmutableMergeOperand(t *testing.T) {
	for _, operand := range []string{"main", "HEAD~1", "--abort", "--squash", "", strings.Repeat("a", 31), strings.Repeat("a", 33), strings.Repeat("w", 32), strings.Repeat("A", 32), testCommit[:31] + "\x00", "immutable'operand"} {
		t.Run(operand, func(t *testing.T) {
			s, p := openPeer(t, nil)
			ctx, cancel := context.WithTimeout(context.Background(), ioBound)
			defer cancel()
			if err := s.acquire(ctx, ioBound); err != nil {
				t.Fatal(err)
			}
			if _, err := s.execute(ctx, mergeOperation, operand); !errors.Is(err, errState) {
				t.Fatalf("invalid operand accepted: %v", err)
			}
			if s.state != held {
				t.Fatal("invalid operand reached dispatch")
			}
			if err := s.abort(ctx); err != nil {
				t.Fatal(err)
			}
			for _, q := range p.commands() {
				if strings.HasPrefix(q, "CALL ") {
					t.Fatal("invalid operand reached wire")
				}
			}
		})
	}
}

func TestDriverParameterEscapingStaysBelowImmutableOperandBoundary(t *testing.T) {
	s, p := openPeer(t, nil)
	ctx, cancel := context.WithTimeout(context.Background(), ioBound)
	defer cancel()
	if _, err := s.consume(ctx, "SELECT ?", "escaped'parameter"); err != nil {
		t.Fatal(err)
	}
	queries := p.commands()
	if queries[len(queries)-1] != "SELECT 'escaped\\'parameter'" {
		t.Fatalf("driver parameter changed: %q", queries)
	}
}

func TestExactReturnedResultCaps(t *testing.T) {
	for _, kind := range []string{"sets", "columns", "rows", "combined bytes"} {
		for _, over := range []bool{false, true} {
			name := kind + " exact"
			if over {
				name = kind + " one over"
			}
			t.Run(name, func(t *testing.T) {
				r := reply{rows: [][]string{{"1"}}}
				n := 0
				if over {
					n = 1
				}
				switch kind {
				case "sets":
					for range 15 + n {
						r.nextSets = append(r.nextSets, reply{rows: [][]string{{"1"}}})
					}
				case "columns":
					r.rows = [][]string{make([]string, 64+n)}
				case "rows":
					r.rows = nil
					for range 256 + n {
						r.rows = append(r.rows, []string{"1"})
					}
				case "combined bytes":
					// Four column-name bytes and four values spread across two sets/columns.
					value := strings.Repeat("x", ((1<<20)-4)/4)
					r.columnNames = []string{"v", "v"}
					r.rows = [][]string{{value, value}}
					r.nextSets = []reply{{columnNames: []string{"v", "v"}, rows: [][]string{{value, value + strings.Repeat("x", n)}}}}
				}
				s, p, done := pipeSession(t, r)
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				sets, err := s.execute(ctx, commitOperation, "")
				if over {
					if !errors.Is(err, errResult) || !errors.Is(err, errUncertain) {
						t.Fatalf("cap not refused with uncertainty: %v", err)
					}
					if s.db.Stats().OpenConnections != 0 {
						t.Fatal("overflow pool open")
					}
					select {
					case <-done:
					case <-time.After(time.Second):
						t.Fatal("overflow peer unjoined")
					}
					for _, q := range p.commands() {
						if strings.Contains(q, "RELEASE") {
							t.Fatal("overflow released")
						}
					}
				} else {
					if err != nil {
						t.Fatal(err)
					}
					rows, columns, bytes := 0, 0, 0
					for _, set := range sets {
						columns += len(set.columns)
						for _, c := range set.columns {
							bytes += len(c)
						}
						rows += len(set.rows)
						for _, row := range set.rows {
							for _, v := range row {
								bytes += len(v.String)
							}
						}
					}
					switch kind {
					case "sets":
						if len(sets) != 16 {
							t.Fatal("set boundary changed")
						}
					case "columns":
						if columns != 64 {
							t.Fatal("column boundary changed")
						}
					case "rows":
						if rows != 256 {
							t.Fatal("row boundary changed")
						}
					case "combined bytes":
						if bytes != 1<<20 || len(sets) != 2 || columns != 4 {
							t.Fatalf("byte accounting changed: %d", bytes)
						}
					}
				}
			})
		}
	}
}
