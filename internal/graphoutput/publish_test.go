package graphoutput

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/atomicfile"
)

type testWriter struct {
	write func([]byte) (int, error)
	close func() error
	abort func() error
}

func (w *testWriter) Write(p []byte) (int, error) { return w.write(p) }
func (w *testWriter) Close() error                { return w.close() }
func (w *testWriter) Abort() error                { return w.abort() }

func requireFile(t *testing.T, path string, want []byte) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("file bytes differ: got %d bytes, want %d", len(got), len(want))
	}
}

func requireNoTemps(t *testing.T, path string) {
	t.Helper()
	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != filepath.Base(path) {
		t.Fatalf("unexpected publication residue: %v", entries)
	}
}

func TestPublishFileReplacement(t *testing.T) {
	for _, size := range []int{0, 19, 2*writeChunkBytes + 7} {
		t.Run(strconv.Itoa(size), func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "export")
			if err := os.WriteFile(path, []byte("original"), 0o644); err != nil {
				t.Fatal(err)
			}
			data := bytes.Repeat([]byte{0xff}, size) // The sink imposes no JSON/UTF-8 format.
			originalData := bytes.Clone(data)
			if err := PublishFile(context.Background(), path, data, 0o640); err != nil {
				t.Fatal(err)
			}
			requireFile(t, path, data)
			if !bytes.Equal(data, originalData) {
				t.Fatal("publication changed caller data")
			}
			requireNoTemps(t, path)
			info, err := os.Stat(path)
			if err != nil {
				t.Fatal(err)
			}
			if got := info.Mode().Perm(); runtime.GOOS != "windows" && got != 0o640 {
				t.Fatalf("permissions = %o, want 0640", got)
			}
		})
	}
}

func TestInvalidContextBeforeIO(t *testing.T) {
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	expired, expire := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer expire()
	for _, tc := range []struct {
		name string
		ctx  context.Context
		want error
	}{
		{"nil", nil, ErrNilContext},
		{"canceled", canceled, context.Canceled},
		{"expired", expired, context.DeadlineExceeded},
	} {
		t.Run(tc.name, func(t *testing.T) {
			factory := func(string, os.FileMode) (publicationWriter, error) {
				t.Fatal("created a writer with an invalid context")
				return nil, nil
			}
			if err := publishFile(tc.ctx, "unused", nil, 0o600, factory); !errors.Is(err, tc.want) {
				t.Fatalf("file error = %v, want %v", err, tc.want)
			}
			writer := &testWriter{write: func([]byte) (int, error) {
				t.Fatal("wrote with an invalid context")
				return 0, nil
			}}
			if err := WriteStream(tc.ctx, writer, []byte("data")); !errors.Is(err, tc.want) {
				t.Fatalf("stream error = %v, want %v", err, tc.want)
			}
			path := filepath.Join(t.TempDir(), "output")
			if err := PublishFile(tc.ctx, path, []byte("data"), 0o600); !errors.Is(err, tc.want) {
				t.Fatalf("real file error = %v, want %v", err, tc.want)
			}
			entries, err := os.ReadDir(filepath.Dir(path))
			if err != nil || len(entries) != 0 {
				t.Fatalf("invalid-context IO: entries=%v err=%v", entries, err)
			}
		})
	}
}

func TestPublishFileCreateFailure(t *testing.T) {
	want := errors.New("create failure")
	if err := publishFile(context.Background(), "unused", nil, 0o600,
		func(string, os.FileMode) (publicationWriter, error) { return nil, want }); !errors.Is(err, want) {
		t.Fatalf("error = %v, want create cause", err)
	}
	path := filepath.Join(t.TempDir(), "absent", "output")
	if err := PublishFile(context.Background(), path, nil, 0o600); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("error = %v, want missing parent", err)
	}
}

// Wrapping a real atomic writer proves failure leaves the existing destination
// unchanged and cleans the owned temporary file, not merely fake call order.
func TestPublishFileFailurePreservesDestination(t *testing.T) {
	writeErr, closeErr, abortErr := errors.New("write failure"), errors.New("close failure"), errors.New("abort failure")
	for _, tc := range []struct {
		name                                                                string
		writeFailure, closeFailure, abortFailure, cancelCreate, cancelWrite bool
		dataSize                                                            int
		want                                                                error
	}{
		{name: "write", writeFailure: true, dataSize: 11, want: writeErr},
		{name: "write-and-abort", writeFailure: true, abortFailure: true, dataSize: 11, want: writeErr},
		{name: "close", closeFailure: true, dataSize: 11, want: closeErr},
		{name: "close-and-abort", closeFailure: true, abortFailure: true, dataSize: 11, want: closeErr},
		{name: "cancel-after-create", cancelCreate: true, dataSize: 11, want: context.Canceled},
		{name: "cancel-between-writes", cancelWrite: true, dataSize: writeChunkBytes + 1, want: context.Canceled},
		{name: "cancel-before-finalize", cancelWrite: true, dataSize: 11, want: context.Canceled},
		{name: "cancel-and-abort", cancelWrite: true, abortFailure: true, dataSize: 11, want: context.Canceled},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "export")
			original := []byte("original")
			if err := os.WriteFile(path, original, 0o644); err != nil {
				t.Fatal(err)
			}
			beforeInfo, err := os.Stat(path)
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			writes, closes, aborts := 0, 0, 0
			factory := func(gotPath string, gotPerm os.FileMode) (publicationWriter, error) {
				if gotPath != path || gotPerm != 0o600 {
					t.Fatal("factory inputs changed")
				}
				w, err := atomicfile.Create(gotPath, gotPerm)
				if err != nil {
					return nil, err
				}
				if tc.cancelCreate {
					cancel()
				}
				return &testWriter{
					write: func(p []byte) (int, error) {
						writes++
						if tc.writeFailure {
							n, err := w.Write(p[:1])
							return n, errors.Join(writeErr, err)
						}
						n, err := w.Write(p)
						if tc.cancelWrite {
							cancel()
						}
						return n, err
					},
					close: func() error {
						closes++
						if tc.closeFailure {
							return closeErr
						}
						return w.Close()
					},
					abort: func() error {
						aborts++
						err := w.Abort()
						if tc.abortFailure {
							return errors.Join(err, abortErr)
						}
						return err
					},
				}, nil
			}
			err = publishFile(ctx, path, bytes.Repeat([]byte("x"), tc.dataSize), 0o600, factory)
			if !errors.Is(err, tc.want) {
				t.Fatalf("error = %v, want %v", err, tc.want)
			}
			if tc.abortFailure && !errors.Is(err, abortErr) {
				t.Fatalf("lost abort cause: %v", err)
			}
			if tc.writeFailure && !errors.Is(err, io.ErrShortWrite) {
				t.Fatalf("lost short-write cause: %v", err)
			}
			wantWrites := 1
			if tc.cancelCreate {
				wantWrites = 0
			}
			wantCloses := 0
			if tc.closeFailure {
				wantCloses = 1
			}
			if writes != wantWrites || closes != wantCloses || aborts != 1 {
				t.Fatalf("writes/closes/aborts = %d/%d/%d, want %d/%d/1", writes, closes, aborts, wantWrites, wantCloses)
			}
			requireFile(t, path, original)
			requireNoTemps(t, path)
			info, err := os.Stat(path)
			if err != nil || info.Mode().Perm() != beforeInfo.Mode().Perm() {
				t.Fatalf("original permissions changed: %v %v", info, err)
			}
		})
	}
}

func TestPublishFileRealRenameFailure(t *testing.T) {
	path := filepath.Join(t.TempDir(), "destination")
	if err := os.Mkdir(path, 0o700); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(path, "retained")
	if err := os.WriteFile(marker, []byte("original"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := PublishFile(context.Background(), path, []byte("replacement"), 0o600); err == nil {
		t.Fatal("replacing a nonempty directory succeeded")
	}
	requireFile(t, marker, []byte("original"))
	requireNoTemps(t, path)
}

func TestPublishFileSuccessBoundary(t *testing.T) {
	for _, when := range []string{"during-finalization", "after-rename"} {
		t.Run(when, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "export")
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			data := []byte("complete")
			factory := func(path string, perm os.FileMode) (publicationWriter, error) {
				w, err := atomicfile.Create(path, perm)
				if err != nil {
					return nil, err
				}
				return &testWriter{
					write: w.Write,
					close: func() error {
						// Neither cancellation inside the final operation nor after
						// successful rename can make complete output a failure.
						if when == "during-finalization" {
							cancel()
						}
						err := w.Close()
						if when == "after-rename" {
							cancel()
						}
						return err
					},
					abort: func() error { t.Fatal("aborted successful publication"); return nil },
				}, nil
			}
			if err := publishFile(ctx, path, data, 0o600, factory); err != nil {
				t.Fatal(err)
			}
			if !errors.Is(ctx.Err(), context.Canceled) {
				t.Fatal("fixture did not cancel")
			}
			requireFile(t, path, data)
			requireNoTemps(t, path)
		})
	}
}

type shortStream struct {
	bytes.Buffer
	err   error
	calls int
}

func (w *shortStream) Write(p []byte) (int, error) {
	w.calls++
	n, _ := w.Buffer.Write(p[:len(p)-1])
	return n, w.err
}

func TestWriteStreamShortWrite(t *testing.T) {
	want := errors.New("pipe failure")
	for _, cause := range []error{nil, want} {
		w := &shortStream{err: cause}
		err := WriteStream(context.Background(), w, []byte("complete"))
		if !errors.Is(err, io.ErrShortWrite) || (cause != nil && !errors.Is(err, cause)) {
			t.Fatalf("lost short-write/pipe cause: %v", err)
		}
		if w.String() != "complet" || w.calls != 1 {
			t.Fatalf("partial delivery = %q in %d calls", w.String(), w.calls)
		}
	}
}

func TestWriteStreamCancellationAndCompletion(t *testing.T) {
	for _, size := range []int{17, writeChunkBytes + 1} {
		t.Run(strconv.Itoa(size), func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			var got bytes.Buffer
			w := &testWriter{write: func(p []byte) (int, error) {
				cancel()
				return got.Write(p)
			}}
			data := bytes.Repeat([]byte("x"), size)
			err := WriteStream(ctx, w, data)
			if size <= writeChunkBytes {
				if err != nil || !bytes.Equal(got.Bytes(), data) {
					t.Fatalf("complete output: len=%d err=%v", got.Len(), err)
				}
			} else if !errors.Is(err, context.Canceled) || got.Len() != writeChunkBytes {
				t.Fatalf("canceled partial output: len=%d err=%v", got.Len(), err)
			}
			if !errors.Is(ctx.Err(), context.Canceled) {
				t.Fatal("fixture did not cancel")
			}
		})
	}
}

func TestWriteStreamEmptyAndNilWriter(t *testing.T) {
	if err := WriteStream(context.Background(), nil, nil); !errors.Is(err, ErrNilWriter) {
		t.Fatalf("nil writer: %v", err)
	}
	w := &testWriter{write: func([]byte) (int, error) { t.Fatal("empty output wrote bytes"); return 0, nil }}
	if err := WriteStream(context.Background(), w, nil); err != nil {
		t.Fatal(err)
	}
	var got bytes.Buffer
	data := bytes.Repeat([]byte("data"), writeChunkBytes)
	originalData := bytes.Clone(data)
	if err := WriteStream(context.Background(), &got, data); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got.Bytes(), data) {
		t.Fatal("multi-write stream changed bytes")
	}
	if !bytes.Equal(data, originalData) {
		t.Fatal("stream changed caller data")
	}
}

func TestPublishFileEmptyPathBeforeIO(t *testing.T) {
	called := false
	err := publishFile(context.Background(), "", []byte("data"), 0o600,
		func(string, os.FileMode) (publicationWriter, error) {
			called = true
			return nil, errors.New("unexpected create")
		})
	if called || !errors.Is(err, ErrEmptyPath) {
		t.Fatalf("empty path: create=%t error=%v", called, err)
	}
}
