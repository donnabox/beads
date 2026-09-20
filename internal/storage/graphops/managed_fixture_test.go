//go:build cgo && graphmanaged_engine && (darwin || linux)

package graphops

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
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
)

// This is a test-binary protocol, not a public adapter or authority credential.
// Its parent is the graphmanaged fixture; all SQL uses that parent's private
// managed generation over MySQL. No embedded OpenSQL call occurs in this worker.
type managedGraphInput struct {
	Protocol int    `json:"protocol"`
	Phase    string `json:"phase"`
	Root     string `json:"root"`
	Endpoint string `json:"endpoint"`
	RunID    string `json:"runId"`
}

func parseManagedGraphInput(raw []byte) (managedGraphInput, error) {
	var input managedGraphInput
	if len(raw) > 4096 {
		return input, errors.New("fixture descriptor exceeds bound")
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return input, errors.New("fixture descriptor must be an object")
	}
	fields := map[string]json.RawMessage{}
	for decoder.More() {
		token, err := decoder.Token()
		key, ok := token.(string)
		if err != nil || !ok || fields[key] != nil {
			return input, errors.New("duplicate or invalid fixture descriptor key")
		}
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil || bytes.Equal(value, []byte("null")) {
			return input, errors.New("invalid fixture descriptor value")
		}
		fields[key] = value
	}
	if _, err := decoder.Token(); err != nil {
		return input, err
	}
	if decoder.Decode(new(any)) != io.EOF || len(fields) != 5 {
		return input, errors.New("fixture descriptor shape/trailing data")
	}
	decoder = json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		return managedGraphInput{}, err
	}
	host, port, err := net.SplitHostPort(input.Endpoint)
	p, portErr := strconv.Atoi(port)
	id, idErr := hex.DecodeString(input.RunID)
	if input.Protocol != 1 || (input.Phase != "seed" && input.Phase != "read") || !filepath.IsAbs(input.Root) || filepath.Clean(input.Root) != input.Root || host != "127.0.0.1" || err != nil || portErr != nil || p < 1 || p > 65535 || strconv.Itoa(p) != port || idErr != nil || len(id) != 16 || hex.EncodeToString(id) != input.RunID {
		return managedGraphInput{}, errors.New("invalid fixture descriptor")
	}
	return input, nil
}

func managedGraphPassword(root string) ([]byte, error) {
	canonical, err := filepath.EvalSymlinks(root)
	if err != nil || canonical != root {
		return nil, errors.New("fixture root must be canonical")
	}
	for _, name := range []string{"", "home", "tmp", "data", "security"} {
		info, err := os.Lstat(filepath.Join(root, name))
		if err != nil {
			return nil, err
		}
		stat, ok := info.Sys().(*syscall.Stat_t)
		if !ok || !info.IsDir() || info.Mode().Perm() != 0700 || stat.Uid != uint32(os.Getuid()) {
			return nil, errors.New("fixture directories require owned mode 0700")
		}
	}
	fd, err := syscall.Open(filepath.Join(root, "security", "password"), syscall.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_CLOEXEC|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, err
	}
	file := os.NewFile(uintptr(fd), "fixture password")
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || !info.Mode().IsRegular() || info.Mode().Perm() != 0600 || info.Size() != 64 || stat.Uid != uint32(os.Getuid()) || stat.Nlink != 1 {
		return nil, errors.New("fixture password requires private owned regular file")
	}
	password, err := io.ReadAll(io.LimitReader(file, 65))
	decoded, decodeErr := hex.DecodeString(string(password))
	if err != nil || decodeErr != nil || len(decoded) != 32 || hex.EncodeToString(decoded) != string(password) {
		return nil, errors.New("invalid fixture password bytes")
	}
	return password, nil
}

func TestManagedGraphWorker(t *testing.T) {
	if os.Getenv("GRAPH_MANAGED_WORKER") != "1" {
		t.Skip("parent-owned managed graph worker")
	}
	raw, err := io.ReadAll(io.LimitReader(os.Stdin, 4097))
	if err != nil {
		t.Fatal(err)
	}
	input, err := parseManagedGraphInput(raw)
	if err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]string{"HOME": filepath.Join(input.Root, "home"), "TMPDIR": filepath.Join(input.Root, "tmp"), "DOLT_METRICS_DISABLED": "1", "DOLT_DISABLE_EVENT_FLUSH": "1"} {
		if os.Getenv(key) != want {
			t.Fatalf("worker isolation mismatch for %s", key)
		}
	}
	password, err := managedGraphPassword(input.Root)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 150*time.Second)
	defer cancel()
	cfg := mysql.NewConfig()
	cfg.Net, cfg.Addr, cfg.User, cfg.Passwd = "tcp", input.Endpoint, "fixture", string(password)
	cfg.Timeout, cfg.ReadTimeout, cfg.WriteTimeout = 5*time.Second, 10*time.Second, 10*time.Second
	connector, err := mysql.NewConnector(cfg)
	if err != nil {
		t.Fatal(err)
	}
	db := sql.OpenDB(connector)
	defer db.Close()
	db.SetMaxOpenConns(1)
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if input.Phase == "seed" {
		if _, err := conn.ExecContext(ctx, "CREATE DATABASE graph_read_fixture"); err != nil {
			t.Fatal(err)
		}
	}
	for _, q := range []string{"USE graph_read_fixture", "SET SESSION time_zone = '+00:00'"} {
		if _, err := conn.ExecContext(ctx, q); err != nil {
			t.Fatal(err)
		}
	}
	var database, branch, zone string
	if err := conn.QueryRowContext(ctx, "SELECT DATABASE(), ACTIVE_BRANCH(), @@session.time_zone").Scan(&database, &branch, &zone); err != nil {
		t.Fatal(err)
	}
	if database != "graph_read_fixture" || branch != "main" || zone != "+00:00" {
		t.Fatalf("fixture connection binding = %q %q %q", database, branch, zone)
	}
	t.Logf("managed MySQL fixture database=%s branch=%s zone=%s phase=%s", database, branch, zone, input.Phase)
	data := filepath.Join(input.Root, "data")
	if input.Phase == "seed" {
		seedEngineFixture(t, ctx, conn, data)
	}
	before := fixtureTablesDigest(t, ctx, conn)
	verifyEngineFixture(t, ctx, conn, data)
	verifyEngineNegativeControls(t, ctx, conn)
	verifyEngineByteControls(t, ctx, conn)
	verifyEngineOwnedOverflow(t, ctx, conn)
	after := fixtureTablesDigest(t, ctx, conn)
	if before != after {
		t.Fatal("managed reads/rolled-back controls changed fixture tables")
	}
	if err := errors.Join(conn.Close(), db.Close()); err != nil {
		t.Fatal(err)
	}
	if t.Failed() {
		t.Fatal("managed graph controls failed")
	}
	// Parent verifies process exit, receipt identity and generation cleanup before
	// it emits the aggregate success marker. No credentials enter the receipt.
	fmt.Printf("MANAGED_GRAPH_WORKER %s %s %s\n", input.RunID, input.Phase, after)
}

func TestManagedGraphDescriptor(t *testing.T) {
	good := `{"protocol":1,"phase":"seed","root":"/private/tmp/fixture","endpoint":"127.0.0.1:1234","runId":"` + strings.Repeat("a", 32) + `"}`
	if _, err := parseManagedGraphInput([]byte(good)); err != nil {
		t.Fatal(err)
	}
	for _, raw := range []string{good + "{}", strings.Replace(good, `"protocol":1`, `"protocol":1,"protocol":1`, 1), strings.Replace(good, `"phase":"seed"`, `"phase":null`, 1), strings.Replace(good, `"root":`, `"unknown":`, 1), strings.Replace(good, "127.0.0.1", "example.com", 1), strings.Replace(good, "1234", "0", 1), strings.Replace(good, "seed", "write", 1), strings.Repeat(" ", 4097)} {
		if _, err := parseManagedGraphInput([]byte(raw)); err == nil {
			t.Fatalf("accepted malformed fixture descriptor: %s", raw)
		}
	}
}

func TestManagedGraphCredentialFile(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(root, 0700); err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{"home", "tmp", "data", "security"} {
		if err := os.Mkdir(filepath.Join(root, dir), 0700); err != nil {
			t.Fatal(err)
		}
	}
	path := filepath.Join(root, "security", "password")
	secret := []byte(strings.Repeat("a", 64))
	if err := os.WriteFile(path, secret, 0600); err != nil {
		t.Fatal(err)
	}
	if got, err := managedGraphPassword(root); err != nil || !bytes.Equal(got, secret) {
		t.Fatalf("valid credential: %v", err)
	}
	if err := os.Chmod(path, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := managedGraphPassword(root); err == nil {
		t.Fatal("accepted public credential")
	}
	if err := os.Chmod(path, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(path, path+".real"); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(path+".real", path); err != nil {
		t.Fatal(err)
	}
	if _, err := managedGraphPassword(root); err == nil {
		t.Fatal("accepted credential symlink")
	}
}

// A separate bounded process makes a blocking-open regression observable without
// leaving a goroutine stuck in open(2) in the parent test process.
func TestManagedGraphCredentialFIFO(t *testing.T) {
	if root := os.Getenv("GRAPH_CREDENTIAL_FIFO_ROOT"); root != "" {
		if _, err := managedGraphPassword(root); err == nil {
			t.Fatal("accepted FIFO credential")
		}
		return
	}
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(root, 0700); err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{"home", "tmp", "data", "security"} {
		if err := os.Mkdir(filepath.Join(root, dir), 0700); err != nil {
			t.Fatal(err)
		}
	}
	if err := syscall.Mkfifo(filepath.Join(root, "security", "password"), 0600); err != nil {
		t.Fatal(err)
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, executable, "-test.run=^TestManagedGraphCredentialFIFO$", "-test.timeout=4s")
	cmd.Env = []string{"PATH=/usr/bin:/bin", "HOME=" + filepath.Join(root, "home"), "TMPDIR=" + filepath.Join(root, "tmp"), "DOLT_METRICS_DISABLED=1", "DOLT_DISABLE_EVENT_FLUSH=1", "GRAPH_CREDENTIAL_FIFO_ROOT=" + root}
	cmd.WaitDelay = time.Second
	var output boundedFixtureOutput
	cmd.Stdout, cmd.Stderr = &output, &output
	if err := cmd.Run(); err != nil || output.overflow {
		t.Fatalf("FIFO refusal did not complete successfully: %v; context=%v; output=%s", err, ctx.Err(), output.String())
	}
}
