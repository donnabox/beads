package graphinput

import (
	"bytes"
	"context"
	"errors"
	"io"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"testing/iotest"
)

func noResidue(t *testing.T, dir string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 0 {
		t.Fatalf("staging residue: %v (%v)", entries, err)
	}
}

type readerFunc func([]byte) (int, error)

func (f readerFunc) Read(p []byte) (int, error) { return f(p) }

type countedReader struct {
	io.Reader
	read   int
	closed bool
}

func (r *countedReader) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)
	r.read += n
	return n, err
}

func (r *countedReader) Close() error { r.closed = true; return nil }

func TestStageCompleteBytes(t *testing.T) {
	for _, data := range [][]byte{nil, []byte("\x00\xff\n\r\nunchanged"), bytes.Repeat([]byte("binary\x00"), chunkBytes)} {
		dir := t.TempDir()
		r := &countedReader{Reader: bytes.NewReader(data)}
		s, err := Stage(context.Background(), r, Options{Dir: dir, MaxBytes: int64(max(1, len(data)))})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = s.Close() })
		got, err := io.ReadAll(s)
		if err != nil || !bytes.Equal(got, data) || s.Size() != int64(len(data)) || r.closed {
			t.Fatalf("staged bytes/size/source ownership: size=%d err=%v closed=%t", s.Size(), err, r.closed)
		}
		if _, writable := any(s).(io.Writer); writable {
			t.Fatal("published input exposes a writer")
		}
		if _, err := s.file.Write([]byte("forbidden")); err == nil {
			t.Fatal("published file descriptor is writable")
		}
		if _, err := s.Seek(0, io.SeekStart); err != nil {
			t.Fatal(err)
		}
		again, err := io.ReadAll(s)
		if err != nil || !bytes.Equal(again, data) {
			t.Fatal("second validation pass changed bytes", err)
		}
		if runtime.GOOS != "windows" {
			for _, path := range []string{s.dir, s.path} {
				info, err := os.Stat(path)
				if err != nil || info.Mode().Perm()&0o077 != 0 {
					t.Fatal("staging permissions expose group/other bits", err)
				}
			}
		}
		if err := s.Close(); err != nil {
			t.Fatal(err)
		}
		if err := s.Close(); err != nil {
			t.Fatal("repeated successful close", err)
		}
		if _, err := s.Read(make([]byte, 1)); !errors.Is(err, os.ErrClosed) {
			t.Fatal("read after close", err)
		}
		if _, err := s.Seek(0, io.SeekStart); !errors.Is(err, os.ErrClosed) {
			t.Fatal("seek after close", err)
		}
		noResidue(t, dir)
	}
}

func TestStageAdmissionBeforeIO(t *testing.T) {
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	for _, tc := range []struct {
		name  string
		ctx   context.Context
		input io.Reader
		limit int64
		want  error
	}{
		{"nil-context", nil, strings.NewReader("x"), 1, ErrNilContext},
		{"canceled", canceled, nil, 0, context.Canceled},
		{"nil-reader", context.Background(), nil, 1, ErrNilReader},
		{"zero", context.Background(), strings.NewReader("x"), 0, ErrInvalidLimit},
		{"negative", context.Background(), strings.NewReader("x"), -1, ErrInvalidLimit},
		{"overflow", context.Background(), strings.NewReader("x"), math.MaxInt64, ErrInvalidLimit},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			s, err := Stage(tc.ctx, tc.input, Options{Dir: dir, MaxBytes: tc.limit})
			if s != nil || !errors.Is(err, tc.want) {
				t.Fatalf("result=%v error=%v want=%v", s, err, tc.want)
			}
			noResidue(t, dir)
		})
	}
	if s, err := Stage(context.Background(), strings.NewReader(""), Options{MaxBytes: 1}); s != nil || !errors.Is(err, ErrEmptyDir) {
		t.Fatalf("empty parent: %v %v", s, err)
	}
	dir := t.TempDir()
	s, err := Stage(context.Background(), strings.NewReader("x"), Options{Dir: dir, MaxBytes: math.MaxInt64 - 1})
	if err != nil || s.Size() != 1 {
		t.Fatalf("large legal bound: %v %v", s, err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	noResidue(t, dir)
}

func TestStageReaderBoundaries(t *testing.T) {
	fault := errors.New("source failed after data")
	for _, tc := range []struct {
		name  string
		input io.Reader
		want  error
	}{
		{"overflow", strings.NewReader("123456789"), ErrLimitExceeded},
		{"no-progress", readerFunc(func([]byte) (int, error) { return 0, nil }), io.ErrNoProgress},
		{"negative-count", readerFunc(func([]byte) (int, error) { return -1, nil }), nil},
		{"excess-count", readerFunc(func(p []byte) (int, error) { return len(p) + 1, nil }), nil},
		{"partial-error", readerFunc(func(p []byte) (int, error) { p[0] = 'x'; return 1, fault }), fault},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			r := &countedReader{Reader: tc.input}
			s, err := Stage(context.Background(), r, Options{Dir: dir, MaxBytes: 5})
			if s != nil || err == nil || (tc.want != nil && !errors.Is(err, tc.want)) {
				t.Fatalf("result=%v error=%v", s, err)
			}
			if tc.name == "overflow" && r.read != 6 {
				t.Fatalf("source consumed %d bytes, want exactly limit+1", r.read)
			}
			if r.closed {
				t.Fatal("closed caller's source")
			}
			noResidue(t, dir)
		})
	}
	dir := t.TempDir()
	r := readerFunc(func(p []byte) (int, error) { return copy(p, "end"), io.EOF })
	s, err := Stage(context.Background(), r, Options{Dir: dir, MaxBytes: 3})
	if err != nil {
		t.Fatal("data with EOF", err)
	}
	got, err := io.ReadAll(s)
	if err != nil || string(got) != "end" {
		t.Fatal("EOF bytes lost", err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	noResidue(t, dir)
	chunked, err := Stage(context.Background(), iotest.OneByteReader(strings.NewReader("chunks")), Options{Dir: dir, MaxBytes: 6})
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(chunked)
	if err != nil || string(data) != "chunks" {
		t.Fatal("short successful reads changed input", err)
	}
	if err := chunked.Close(); err != nil {
		t.Fatal(err)
	}
	noResidue(t, dir)
}

type faultFile struct {
	stageFile
	write func([]byte) (int, error)
	sync  func() error
	close func() error
	stat  func() (os.FileInfo, error)
}

func (f *faultFile) Write(p []byte) (int, error) {
	if f.write != nil {
		return f.write(p)
	}
	return f.stageFile.Write(p)
}
func (f *faultFile) Sync() error {
	if f.sync != nil {
		return f.sync()
	}
	return f.stageFile.Sync()
}
func (f *faultFile) Close() error {
	if f.close != nil {
		return f.close()
	}
	return f.stageFile.Close()
}
func (f *faultFile) Stat() (os.FileInfo, error) {
	if f.stat != nil {
		return f.stat()
	}
	return f.stageFile.Stat()
}

func TestStageFilesystemFailures(t *testing.T) {
	fault := errors.New("injected filesystem failure")
	for _, point := range []string{"create", "write", "short-write", "sync", "writer-close", "open", "stat", "size"} {
		t.Run(point, func(t *testing.T) {
			dir := t.TempDir()
			ops := diskOperations()
			create, open := ops.create, ops.open
			ops.create = func(dir string) (stageFile, error) {
				if point == "create" {
					return nil, fault
				}
				file, err := create(dir)
				if err != nil {
					return nil, err
				}
				f := &faultFile{stageFile: file}
				switch point {
				case "write":
					f.write = func([]byte) (int, error) { return 0, fault }
				case "short-write":
					f.write = func(p []byte) (int, error) { return len(p) - 1, nil }
				case "sync":
					f.sync = func() error { return fault }
				case "writer-close":
					f.close = func() error { return errors.Join(file.Close(), fault) }
				}
				return f, nil
			}
			ops.open = func(path string) (stageFile, error) {
				if point == "open" {
					return nil, fault
				}
				f, err := open(path)
				if err != nil {
					return nil, err
				}
				if point == "size" {
					if err := os.Truncate(path, 0); err != nil {
						_ = f.Close()
						return nil, err
					}
				}
				if point == "stat" {
					return &faultFile{stageFile: f, stat: func() (os.FileInfo, error) { return nil, fault }}, nil
				}
				return f, nil
			}
			s, err := stage(context.Background(), strings.NewReader("data"), Options{Dir: dir, MaxBytes: 4}, ops)
			if s != nil || err == nil {
				t.Fatalf("fault returned success: %v %v", s, err)
			}
			if point == "short-write" && !errors.Is(err, io.ErrShortWrite) {
				t.Fatal("short write identity lost", err)
			}
			if point != "short-write" && point != "size" && !errors.Is(err, fault) {
				t.Fatal("injected error identity lost", err)
			}
			noResidue(t, dir)
		})
	}
}

func TestStageCancellationAndCleanup(t *testing.T) {
	for _, late := range []bool{false, true} {
		dir := t.TempDir()
		ctx, cancel := context.WithCancel(context.Background())
		ops := diskOperations()
		input := io.Reader(strings.NewReader("data"))
		if late {
			open := ops.open
			ops.open = func(path string) (stageFile, error) { f, err := open(path); cancel(); return f, err }
		} else {
			input = readerFunc(func(p []byte) (int, error) { cancel(); return copy(p, "data"), io.EOF })
		}
		s, err := stage(ctx, input, Options{Dir: dir, MaxBytes: 4}, ops)
		cancel()
		if s != nil || !errors.Is(err, context.Canceled) {
			t.Fatalf("cancellation late=%t result=%v error=%v", late, s, err)
		}
		noResidue(t, dir)
	}
	dir := t.TempDir()
	fault := errors.New("remove failure")
	ops := diskOperations()
	remove := ops.remove
	ops.remove = func(path string) error {
		if filepath.Base(path) != filepath.Base(dir) && strings.HasPrefix(filepath.Base(path), "payload-") {
			return fault
		}
		return remove(path)
	}
	s, err := stage(context.Background(), strings.NewReader("x"), Options{Dir: dir, MaxBytes: 1}, ops)
	if err != nil {
		t.Fatal(err)
	}
	first := s.Close()
	if !errors.Is(first, fault) || s.Close() != first {
		t.Fatal("cleanup failure lost or repeated close claims success", first)
	}
	if _, err := os.Stat(s.path); err != nil {
		t.Fatal("failed cleanup falsely removed payload", err)
	}
	if err := os.Remove(s.path); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(s.dir); err != nil {
		t.Fatal(err)
	}
	noResidue(t, dir)
}

func TestStageNonrecursiveCleanupAndConcurrency(t *testing.T) {
	dir := t.TempDir()
	type result struct {
		input *StagedInput
		err   error
	}
	results := make(chan result, 8)
	for range cap(results) {
		go func() {
			s, err := Stage(context.Background(), strings.NewReader("same-parent"), Options{Dir: dir, MaxBytes: 11})
			results <- result{s, err}
		}()
	}
	seen := make(map[string]bool)
	for range cap(results) {
		r := <-results
		if r.err != nil {
			t.Fatal(r.err)
		}
		s := r.input
		if seen[s.dir] {
			t.Fatal("stages reused a directory")
		}
		seen[s.dir] = true
		data, err := io.ReadAll(s)
		if err != nil || string(data) != "same-parent" {
			t.Fatal("concurrent stage mixed bytes", err)
		}
		if err := s.Close(); err != nil {
			t.Fatal(err)
		}
	}
	noResidue(t, dir)
	s, err := Stage(context.Background(), strings.NewReader(""), Options{Dir: dir, MaxBytes: 1})
	if err != nil {
		t.Fatal(err)
	}
	foreign := filepath.Join(s.dir, "not-owned")
	if err := os.WriteFile(foreign, []byte("preserve"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err == nil {
		t.Fatal("directory removal ignored unexpected content")
	}
	if got, err := os.ReadFile(foreign); err != nil || string(got) != "preserve" {
		t.Fatal("cleanup recursively removed unrelated content", err)
	}
}

func TestStagePreservesOperationAndCleanupErrors(t *testing.T) {
	dir := t.TempDir()
	readErr, removeErr, closeErr := errors.New("source"), errors.New("remove"), errors.New("close")
	ops := diskOperations()
	create, remove := ops.create, ops.remove
	var path, child string
	ops.create = func(d string) (stageFile, error) {
		child = d
		f, err := create(d)
		if err != nil {
			return nil, err
		}
		path = f.Name()
		return &faultFile{stageFile: f, close: func() error { return errors.Join(f.Close(), closeErr) }}, nil
	}
	ops.remove = func(p string) error {
		if p == path {
			return removeErr
		}
		return remove(p)
	}
	r := readerFunc(func([]byte) (int, error) { return 0, readErr })
	s, err := stage(context.Background(), r, Options{Dir: dir, MaxBytes: 1}, ops)
	if s != nil || !errors.Is(err, readErr) || !errors.Is(err, closeErr) || !errors.Is(err, removeErr) {
		t.Fatalf("failure identities lost: %v %v", s, err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(child); err != nil {
		t.Fatal(err)
	}
	noResidue(t, dir)
}
