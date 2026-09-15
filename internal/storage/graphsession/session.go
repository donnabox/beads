package graphsession

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/go-sql-driver/mysql"
)

var (
	errBusy      = errors.New("graph session: acquisition busy")
	errState     = errors.New("graph session: invalid or overlapping lifecycle call")
	errResult    = errors.New("graph session: incomplete or invalid result")
	errReconnect = errors.New("graph session: reconnect forbidden")
	errUncertain = errors.New("graph session: operation outcome uncertain")
)

type phase uint8

const (
	waiting phase = iota
	held
	dispatched
	terminal
	aborted
	discarded
)

type session struct {
	active    atomic.Bool
	state     phase
	db        *sql.DB
	conn      *sql.Conn
	transport *transport
	target    endpoint
	name      string
	id        uint64
	closeErr  error
}

func open(ctx context.Context, e endpoint) (*session, error) {
	if err := e.validate(); err != nil {
		return nil, err
	}
	s := &session{target: e, name: lockName(e.base)}
	c, err := mysql.NewConnector(s.config(e))
	if err != nil {
		return nil, err
	}
	s.db = sql.OpenDB(&connector{Connector: c})
	s.db.SetMaxOpenConns(1)
	s.db.SetMaxIdleConns(0)
	bounded, cancel := context.WithTimeout(ctx, ioBound)
	defer cancel()
	s.conn, err = s.db.Conn(bounded)
	if err == nil {
		err = s.initialize(bounded)
	}
	if err != nil {
		return nil, s.dispose(err)
	}
	return s, nil
}

func (s *session) enter() error {
	if !s.active.CompareAndSwap(false, true) {
		return errState
	}
	return nil
}
func (s *session) leave() { s.active.Store(false) }

func (s *session) initialize(ctx context.Context) error {
	row, err := s.one(ctx, "SELECT @@autocommit, @@dolt_transactions_disabled, @@dolt_transaction_commit, CONNECTION_ID()")
	if err != nil {
		return err
	}
	if len(row) != 4 || row[0] != "1" || row[1] != "0" || row[2] != "0" {
		return errResult
	}
	s.id, err = strconv.ParseUint(row[3], 10, 64)
	if err != nil || s.id == 0 {
		return errResult
	}
	return nil
}

// acquire polls only definitive zero results. Any ambiguous attempt disposes
// the physical session; a subsequent GET_LOCK could otherwise be reentrant.
func (s *session) acquire(ctx context.Context, budget time.Duration) error {
	if err := s.enter(); err != nil {
		return err
	}
	defer s.leave()
	if s.state != waiting || budget <= 0 || budget > 30*time.Second {
		return errState
	}
	bounded, cancel := context.WithTimeoutCause(ctx, budget, errBusy)
	defer cancel()
	for {
		if err := bounded.Err(); err != nil {
			return s.dispose(context.Cause(bounded))
		}
		attempt, stop := context.WithTimeout(bounded, writeBound)
		row, err := s.one(attempt, "SELECT GET_LOCK(?, 0), CONNECTION_ID()", s.name)
		stop()
		if err != nil {
			return s.dispose(errors.Join(err, context.Cause(bounded)))
		}
		if len(row) != 2 || row[1] != strconv.FormatUint(s.id, 10) || (row[0] != "0" && row[0] != "1") {
			return s.dispose(errResult)
		}
		if row[0] == "1" {
			break
		}
		timer := time.NewTimer(20 * time.Millisecond)
		select {
		case <-bounded.Done():
			timer.Stop()
			return s.dispose(context.Cause(bounded))
		case <-timer.C:
		}
	}
	s.state = held
	// Only construction-owned read work exists here. START/SET repair is banned.
	if _, err := s.consume(bounded, "ROLLBACK"); err != nil {
		return s.dispose(err)
	}
	if err := s.identity(bounded); err != nil {
		return s.dispose(err)
	}
	return nil
}

func (s *session) identity(ctx context.Context) error {
	row, err := s.one(ctx, "SELECT DATABASE(), active_branch(), CONNECTION_ID(), IS_USED_LOCK(?)", s.name)
	if err != nil {
		return err
	}
	if len(row) != 4 {
		return errResult
	}
	base, branch := splitDatabase(row[0])
	if base != s.target.base || branch != s.target.branch || row[1] != s.target.branch || row[2] != strconv.FormatUint(s.id, 10) || row[3] != row[2] {
		return errIdentity
	}
	return nil
}

type operation uint8

const (
	commitOperation operation = iota
	mergeOperation
)

// execute supports only the closed candidate commands, never a callback or raw
// SQL runner. There is no production caller or qualified merge form yet. Future
// composition must supply lease/workload/evidence admission before this call.
func (s *session) execute(ctx context.Context, op operation, operand string) ([]resultSet, error) {
	if err := s.enter(); err != nil {
		return nil, err
	}
	defer s.leave()
	if s.state != held {
		return nil, errState
	}
	var query string
	var args []any
	switch op {
	case commitOperation:
		if operand != "" {
			return nil, errState
		}
		query = "COMMIT"
	case mergeOperation:
		if !validCommitOperand(operand) {
			return nil, errState
		}
		query = "CALL DOLT_MERGE(?)"
		args = []any{operand}
	default:
		return nil, errState
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	deadline, ok := ctx.Deadline()
	if !ok || time.Until(deadline) > 30*time.Second {
		return nil, errState
	}
	s.state = dispatched // Must precede QueryContext and any possible wire write.
	result, err := s.consume(ctx, query, args...)
	if err != nil {
		return result, s.dispose(errors.Join(errUncertain, err))
	}
	s.state = terminal
	return result, nil
}

// refresh is restricted to a completely consumed successful operation. It is
// not recovery or rollback of published effects; failed operations are gone.
func (s *session) refresh(ctx context.Context) error {
	if err := s.enter(); err != nil {
		return err
	}
	defer s.leave()
	if s.state != terminal {
		return errState
	}
	bounded, cancel := context.WithTimeout(ctx, ioBound)
	defer cancel()
	if _, err := s.consume(bounded, "ROLLBACK"); err != nil {
		return s.dispose(err)
	}
	if err := s.identity(bounded); err != nil {
		return s.dispose(err)
	}
	return nil
}

// finish may be called only after external required inspection/witness work.
// There is no success Boolean which could manufacture a terminal operation.
func (s *session) finish(ctx context.Context) error {
	if err := s.enter(); err != nil {
		return err
	}
	defer s.leave()
	if s.state != terminal {
		return errState
	}
	return s.release(ctx)
}
func (s *session) abort(ctx context.Context) error {
	if err := s.enter(); err != nil {
		return err
	}
	defer s.leave()
	if s.state != held {
		return errState
	}
	bounded, cancel := context.WithTimeout(ctx, ioBound)
	defer cancel()
	if _, err := s.consume(bounded, "ROLLBACK"); err != nil {
		return s.dispose(err)
	}
	s.state = aborted
	return s.release(bounded)
}
func (s *session) release(ctx context.Context) error {
	bounded, cancel := context.WithTimeout(ctx, ioBound)
	defer cancel()
	row, err := s.one(bounded, "SELECT RELEASE_LOCK(?)", s.name)
	if err == nil && (len(row) != 1 || row[0] != "1") {
		err = errResult
	}
	if err == nil {
		row, err = s.one(bounded, "SELECT IS_USED_LOCK(?)", s.name)
		// NULL or a successor owner is lawful. Do not require globally free.
		if err == nil {
			if len(row) != 1 {
				err = errResult
			} else if row[0] != "\x00" {
				owner, parseErr := strconv.ParseUint(row[0], 10, 64)
				if parseErr != nil || owner == 0 || owner == s.id {
					err = errResult
				}
			}
		}
	}
	return s.dispose(err)
}

// close never attempts release/recovery after an uncertain dispatch. A caller
// abandoning held ownership must use abort to obtain a clean release result.
func (s *session) close() error {
	if err := s.enter(); err != nil {
		return err
	}
	defer s.leave()
	if s.state == discarded {
		return s.closeErr
	}
	return s.dispose(nil)
}
func (s *session) dispose(primary error) error {
	if s.state == discarded {
		return errors.Join(primary, s.closeErr)
	}
	s.state = discarded
	var cleanup error
	if s.conn != nil {
		err := s.conn.Raw(func(value any) error {
			c, ok := value.(driver.Conn)
			if !ok {
				cleanup = errState
			} else {
				cleanup = c.Close()
			}
			return driver.ErrBadConn
		})
		if err != nil && !errors.Is(err, driver.ErrBadConn) && !errors.Is(err, sql.ErrConnDone) {
			cleanup = errors.Join(cleanup, err)
		}
		err = s.conn.Close()
		if err != nil && !errors.Is(err, sql.ErrConnDone) {
			cleanup = errors.Join(cleanup, err)
		}
	}
	if s.transport != nil {
		cleanup = errors.Join(cleanup, s.transport.Close(), s.transport.errors())
	}
	if s.db != nil {
		cleanup = errors.Join(cleanup, s.db.Close())
	}
	s.closeErr = errors.Join(primary, cleanup)
	return s.closeErr
}

type resultSet struct {
	columns []string
	rows    [][]sql.NullString
}

// consume owns every Rows and result set, copies data, checks Err and Close,
// and bounds returned evidence, preserving NULL separately from every string.
// This is protocol completion, not engine proof.
func (s *session) consume(ctx context.Context, query string, args ...any) (sets []resultSet, err error) {
	// mysql Rows.Close stops its own context watcher before draining packets.
	// Keep an independent physical-transport cancellation owner alive through
	// both explicit and database/sql-initiated Close. Every callback is joined.
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	ownedTransport := s.transport
	if ownedTransport == nil {
		return nil, errState
	}
	finished := make(chan struct{})
	stop := context.AfterFunc(ctx, func() {
		defer close(finished)
		_ = ownedTransport.Close()
	})
	var rows *sql.Rows
	defer func() {
		// A local refusal is not permission to drain an unbounded unread tail.
		// Invalidate before Rows.Close and retain only errors actually observed.
		if err != nil {
			_ = ownedTransport.Close()
		}
		if rows != nil {
			err = errors.Join(err, rows.Close())
		}
		if !stop() {
			<-finished
		}
		err = errors.Join(err, ctx.Err(), ownedTransport.errors())
	}()
	rows, err = s.conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	size, count := 0, 0
	for {
		columns, e := rows.Columns()
		if e != nil {
			return sets, e
		}
		if len(columns) > 64 || len(sets) >= 16 {
			return sets, errResult
		}
		for _, column := range columns {
			size += len(column)
		}
		if size > 1<<20 {
			return sets, errResult
		}
		sets = append(sets, resultSet{columns: append([]string(nil), columns...)})
		set := &sets[len(sets)-1]
		for rows.Next() {
			values := make([]sql.NullString, len(columns))
			dest := make([]any, len(columns))
			for i := range values {
				dest[i] = &values[i]
			}
			if e := rows.Scan(dest...); e != nil {
				return sets, e
			}
			for _, v := range values {
				size += len(v.String)
			}
			count++
			if count > 256 || size > 1<<20 {
				return sets, errResult
			}
			set.rows = append(set.rows, values)
		}
		if e := rows.Err(); e != nil {
			return sets, e
		}
		if !rows.NextResultSet() {
			return sets, rows.Err()
		}
	}
}
func (s *session) one(ctx context.Context, query string, args ...any) ([]string, error) {
	sets, err := s.consume(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	if len(sets) != 1 || len(sets[0].rows) != 1 {
		return nil, fmt.Errorf("%w: expected one row", errResult)
	}
	values := sets[0].rows[0]
	row := make([]string, len(values))
	for i, value := range values {
		if value.Valid {
			if strings.ContainsRune(value.String, '\x00') {
				return nil, errResult
			}
			row[i] = value.String
		} else {
			row[i] = "\x00"
		}
	}
	return row, nil
}
