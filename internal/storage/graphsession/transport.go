package graphsession

import (
	"context"
	"database/sql/driver"
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-sql-driver/mysql"
)

const writeBound = 250 * time.Millisecond
const ioBound = 2 * time.Second

// transport retains errors the MySQL/database/sql cleanup paths may discard.
// The mutex protects observation only; never hold it over network I/O.
type transport struct {
	net.Conn
	mu       sync.Mutex
	failure  error
	once     sync.Once
	closeErr error
}

func (t *transport) record(err error) {
	if err != nil {
		t.mu.Lock()
		t.failure = errors.Join(t.failure, err)
		t.mu.Unlock()
	}
}
func (t *transport) Write(b []byte) (int, error) { n, e := t.Conn.Write(b); t.record(e); return n, e }
func (t *transport) SetWriteDeadline(d time.Time) error {
	e := t.Conn.SetWriteDeadline(d)
	t.record(e)
	return e
}
func (t *transport) Close() error {
	t.once.Do(func() { t.closeErr = t.Conn.Close(); t.record(t.closeErr) })
	return t.closeErr
}
func (t *transport) errors() error { t.mu.Lock(); defer t.mu.Unlock(); return t.failure }

type connector struct {
	driver.Connector
	attempted atomic.Bool
}

func (c *connector) Connect(ctx context.Context) (driver.Conn, error) {
	if !c.attempted.CompareAndSwap(false, true) {
		return nil, errReconnect
	}
	return c.Connector.Connect(ctx)
}

// config is always made from NewConfig, never parsed from an arbitrary DSN.
// The dialer bypasses the driver's process-global registered dialer map.
func (s *session) config(e endpoint) *mysql.Config {
	c := mysql.NewConfig()
	c.Net = "tcp"
	c.Addr = e.address
	c.User = e.user
	c.Passwd = e.password
	c.DBName = e.base + "/" + e.branch
	c.ParseTime = true
	c.Loc = time.UTC
	c.Timeout = ioBound
	c.ReadTimeout = ioBound
	c.WriteTimeout = writeBound
	// Use the driver's parameter escaping for the closed commands below. This
	// does not enable multiStatements, arbitrary parameters or SQL initializers.
	c.InterpolateParams = true
	c.DialFunc = func(ctx context.Context, network, address string) (net.Conn, error) {
		var d net.Dialer
		conn, err := d.DialContext(ctx, network, address)
		if err != nil {
			return nil, err
		}
		s.transport = &transport{Conn: conn}
		return s.transport, nil
	}
	return c
}
