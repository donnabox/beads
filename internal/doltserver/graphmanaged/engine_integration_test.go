//go:build cgo && graphmanaged_engine && (darwin || linux)

package graphmanaged

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/dolthub/dolt/go/cmd/dolt/commands/engine"
	"github.com/dolthub/dolt/go/libraries/doltcore/env"
	"github.com/dolthub/dolt/go/libraries/utils/config"
	"github.com/dolthub/dolt/go/libraries/utils/filesys"
	"github.com/dolthub/go-mysql-server/eventscheduler"
	"github.com/dolthub/go-mysql-server/server"
	gsql "github.com/dolthub/go-mysql-server/sql"
	vitess "github.com/dolthub/vitess/go/mysql"
	"github.com/go-sql-driver/mysql"
)

// This opt-in test couples the existing private controller to a real engine in
// two separate child processes. Its admission values are fixture metadata, not
// a production qualification, and its SQL table is not a graph or BDP resource.
func TestManagedEnginePersistence(t *testing.T) {
	root := testRoot(t)
	for _, name := range []string{"home", "tmp", "data", "security"} {
		if err := os.Mkdir(filepath.Join(root, name), 0700); err != nil {
			t.Fatal(err)
		}
	}
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		t.Fatal(err)
	}
	password := hex.EncodeToString(secret)
	passwordPath := filepath.Join(root, "security", "password")
	mustWrite(t, passwordPath, []byte(password), 0600)
	const value = "Monday 14:00\nReason: recovery check — exact retained bytes."
	for _, phase := range []string{"write", "read"} {
		t.Run(phase, func(t *testing.T) {
			a := fixtureAdmission(t, "unused")
			a.cwd = filepath.Join(root, "data")
			a.environment = []string{"HOME=" + filepath.Join(root, "home"), "TMPDIR=" + filepath.Join(root, "tmp")}
			a.argv = []string{"-test.run=^TestManagedEngineChild$", "managed-engine", "--root=" + root, "--profile=" + a.profile, "--config=" + a.config, "--registration=" + a.registration}
			// testing cancels t.Context before cleanup callbacks; the owned
			// generation must remain alive until its explicit close completes.
			ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
			t.Cleanup(cancel)
			controller := &controller{observe: recordOwner}
			g, err := controller.startWithTiming(ctx, a, timing{startup: 25 * time.Second, cleanup: 15 * time.Second, grace: 5 * time.Second, pipe: 100 * time.Millisecond})
			if g != nil {
				t.Cleanup(func() {
					closeCtx, stop := context.WithTimeout(context.Background(), 18*time.Second)
					defer stop()
					if err := g.close(closeCtx); !onlyProcessGone(err) {
						g.stderr.mu.Lock()
						stderr := string(g.stderr.bytes)
						g.stderr.mu.Unlock()
						t.Errorf("managed close: %v; stderr: %s", err, stderr)
					}
					select {
					case <-g.waitDone:
					default:
						t.Error("managed child has not been reaped")
					}
				})
			}
			if err != nil {
				if g != nil {
					g.stderr.mu.Lock()
					stderr := string(g.stderr.bytes)
					g.stderr.mu.Unlock()
					t.Logf("managed startup stderr: %s", stderr)
				}
				t.Fatal(err)
			}
			if !g.valid() {
				t.Fatal("managed generation was not published")
			}
			t.Logf("phase=%s child=%d endpoint=%s", phase, g.cmd.Process.Pid, g.expected.Socket)
			cfg := mysql.NewConfig()
			cfg.Net, cfg.Addr, cfg.User, cfg.Passwd = "tcp", g.expected.Socket, "fixture", password
			cfg.Timeout, cfg.ReadTimeout, cfg.WriteTimeout = 5*time.Second, 5*time.Second, 5*time.Second
			connector, err := mysql.NewConnector(cfg)
			if err != nil {
				t.Fatal(err)
			}
			db := sql.OpenDB(connector)
			db.SetMaxOpenConns(1)
			defer func() {
				if err := db.Close(); err != nil {
					t.Errorf("client close: %v", err)
				}
			}()
			if phase == "write" {
				for _, q := range []string{"CREATE DATABASE managed_smoke", "CREATE TABLE managed_smoke.records (id INT PRIMARY KEY, content TEXT NOT NULL)"} {
					if _, err := db.ExecContext(ctx, q); err != nil {
						t.Fatal(err)
					}
				}
				if _, err := db.ExecContext(ctx, "INSERT INTO managed_smoke.records VALUES (1, ?)", value); err != nil {
					t.Fatal(err)
				}
			}
			var id int
			var got string
			if err := db.QueryRowContext(ctx, "SELECT id, content FROM managed_smoke.records WHERE id = 1").Scan(&id, &got); err != nil {
				t.Fatal(err)
			}
			if id != 1 || got != value {
				t.Fatalf("persisted row = (%d, %q), want (1, %q)", id, got, value)
			}
		})
		if t.Failed() {
			t.FailNow()
		}
		if phase == "write" {
			if _, err := os.Stat(filepath.Join(root, "data", "managed_smoke", ".dolt")); err != nil {
				t.Fatalf("database was not persisted in managed data root: %v", err)
			}
		}
	}
}

// TestManagedEngineChild is entered only by the emitted test binary with the
// private positional marker. Ordinary test runs never initialize an engine.
func TestManagedEngineChild(t *testing.T) {
	options := map[string]string{}
	marked := false
	for _, arg := range os.Args {
		marked = marked || arg == "managed-engine"
		if key, value, ok := strings.Cut(arg, "="); ok {
			options[key] = value
		}
	}
	if !marked {
		t.Skip("child entrypoint; requires managed-engine marker")
	}
	if err := runManagedEngineChild(options); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	os.Exit(0)
}

func runManagedEngineChild(options map[string]string) (result error) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	control, err := inheritedPipe(4)
	if err != nil {
		return err
	}
	defer control.Close()
	report, err := inheritedPipe(5)
	if err != nil {
		return err
	}
	defer report.Close()
	f := os.NewFile(3, "managed engine listener")
	syscall.CloseOnExec(3)
	listener, err := net.FileListener(f)
	closeErr := f.Close()
	if err != nil || closeErr != nil {
		if listener != nil {
			return errors.Join(err, closeErr, listener.Close())
		}
		return errors.Join(err, closeErr)
	}
	defer listener.Close()
	addr, ok := listener.Addr().(*net.TCPAddr)
	if !ok || !addr.IP.Equal(net.IPv4(127, 0, 0, 1)) {
		return errInput
	}
	sequence, err := strconv.ParseUint(options["--generation"], 10, 64)
	if err != nil || sequence == 0 || options["--managed-protocol"] != "1" {
		return errInput
	}
	root := options["--root"]
	if !filepath.IsAbs(root) {
		return errInput
	}
	fs, err := filesys.LocalFS.WithWorkingDir(filepath.Join(root, "data"))
	if err != nil {
		return err
	}
	mr, err := env.MultiEnvForConfigAndDirectory(ctx, config.NewMapConfig(map[string]string{"user.name": "Managed test", "user.email": "managed@example.invalid", "init.defaultbranch": "main", "metrics.disabled": "true"}), fs, nil)
	if err != nil {
		return err
	}
	defer func() { result = errors.Join(result, mr.Close(context.Background())) }()
	se, err := engine.NewSqlEngine(ctx, mr, &engine.SqlEngineConfig{
		Autocommit: true, SkipRootUserInitialization: true,
		DoltCfgDirPath: filepath.Join(root, "security"), PrivFilePath: filepath.Join(root, "security", "privileges.db"), BranchCtrlFilePath: filepath.Join(root, "security", "branch_control.db"),
		EventSchedulerStatus: eventscheduler.SchedulerDisabled,
		SystemVariables:      engine.SystemVariables{"dolt_stats_enabled": int8(0), "local_infile": int8(0), "secure_file_priv": filepath.Join(root, "security", "no-files")},
	})
	if err != nil {
		return err
	}
	engineClosed := false
	closeEngine := func() error {
		if engineClosed {
			return nil
		}
		engineClosed = true
		err := se.Close()
		// GMS BackgroundThreads.Shutdown cancels its own context on normal
		// shutdown. Only that exact sentinel is tolerated, never joined errors.
		if err == context.Canceled {
			return nil
		}
		return err
	}
	defer func() { result = errors.Join(result, closeEngine()) }()
	password, err := os.ReadFile(filepath.Join(root, "security", "password"))
	if err != nil {
		return err
	}
	users := se.GetUnderlyingEngine().Analyzer.Catalog.MySQLDb
	ed := users.Editor()
	users.AddSuperUser(ed, "fixture", "%", string(password))
	ed.Close()
	srv, err := server.NewServer(server.Config{Listener: listener, Protocol: "tcp", Address: listener.Addr().String(), NoDefaults: true, DisableClientMultiStatements: true, MaxConnections: 4, ConnReadTimeout: 5 * time.Second, ConnWriteTimeout: 5 * time.Second, MaxLoggedQueryLen: -1}, se.GetUnderlyingEngine(), se.ContextFactory, func(ctx context.Context, conn *vitess.Conn, address string) (gsql.Session, error) {
		base, err := gsql.BaseSessionFromConnection(ctx, conn, address)
		if err != nil {
			return nil, err
		}
		return se.NewDoltSession(ctx, base)
	}, nil)
	if err != nil {
		return err
	}
	var serving chan error
	defer func() {
		result = errors.Join(result, srv.Close())
		sm := srv.SessionManager()
		var killErr error
		iterErr := sm.Iter(func(s gsql.Session) (bool, error) {
			killErr = errors.Join(killErr, sm.KillConnection(s.ID()))
			return false, nil
		})
		result = errors.Join(result, iterErr, killErr)
		sm.WaitForClosedConnections()
		if serving != nil {
			result = errors.Join(result, <-serving)
		}
		result = errors.Join(result, closeEngine())
	}()
	prepared := frame{Protocol: 1, Phase: "prepared", Generation: sequence, Profile: options["--profile"], Config: options["--config"], Registration: options["--registration"], Socket: listener.Addr().String()}
	bytes, err := frameBytes(prepared)
	if err != nil {
		return err
	}
	if err := writeFrame(report, bytes, time.Second); err != nil {
		return err
	}
	activation, err := readFrame(ctx, control, 100*time.Millisecond)
	if err != nil || activation != (frame{Protocol: 1, Phase: "activate", Generation: sequence}) {
		return errors.Join(errProtocol, err)
	}
	serving = make(chan error, 1)
	go func() { serving <- srv.Start() }()
	prepared.Phase = "activated"
	bytes, err = frameBytes(prepared)
	if err != nil {
		return err
	}
	if err := writeFrame(report, bytes, time.Second); err != nil {
		return err
	}
	readCtx, cancelRead := context.WithCancel(ctx)
	defer cancelRead()
	controlDone := make(chan struct{})
	var controlErr error
	go func() {
		defer close(controlDone)
		var extra [1]byte
		controlErr = readExact(readCtx, control, extra[:], 100*time.Millisecond)
		if controlErr == nil {
			controlErr = errProtocol
		}
	}()
	select {
	case <-ctx.Done():
	case <-controlDone:
	}
	cancelRead()
	<-controlDone
	if controlErr != nil && !errors.Is(controlErr, io.EOF) && !errors.Is(controlErr, context.Canceled) {
		return controlErr
	}
	return result
}
