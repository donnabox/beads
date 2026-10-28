package local

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/steveyegge/beads/resourceget"
)

func TestExplicitOptionsPreserveWorkspaceAndAuthorityChecks(t *testing.T) {
	for _, backend := range []Backend{Embedded, Server} {
		o := Options{Backend: backend, WorkspaceID: "/real/workspace", ScopeURL: "https://example.test/get/", AuthorityID: "0123456789abcdef0123456789abcdef", Database: "existing", DataDir: "/real/workspace/dolt", Server: ServerOptions{Host: "localhost", Port: 3307, User: "reader", Password: "secret", Socket: "/tmp/dolt.sock", TLS: true}}
		got, err := storeOptions(o)
		if err != nil {
			t.Fatal(err)
		}
		if got.Binding.WorkspaceID != o.WorkspaceID || got.Binding.ScopeURL != o.ScopeURL || got.Binding.AuthorityID != o.AuthorityID || got.TypeGeneration != 13 || got.Binding.SchemaVersion != 11 || got.Branch != "main" || got.ServerPassword != o.Server.Password {
			t.Fatal("lost exact workspace/source configuration")
		}
		encoded, err := json.Marshal(o)
		if err != nil || bytes.Contains(encoded, []byte("secret")) || bytes.Contains(encoded, []byte("Password")) {
			t.Fatal("connection options exposed a password in JSON")
		}
	}
	if _, err := OpenExisting(context.Background(), Options{}); !errors.Is(err, resourceget.ErrInvalidRequest) {
		t.Fatalf("unknown backend: %v", err)
	}
	// Both backends refuse before attempting an engine/connection open. This
	// exercises the existing real filesystem WorkspaceID requirement unchanged.
	for _, backend := range []Backend{Embedded, Server} {
		_, err := OpenExisting(context.Background(), Options{Backend: backend, WorkspaceID: "relative", ScopeURL: "https://example.test/get/", AuthorityID: "0123456789abcdef0123456789abcdef", Database: "existing", DataDir: "/tmp/existing", Server: ServerOptions{Host: "127.0.0.1", Port: 3307}})
		if !errors.Is(err, resourceget.ErrInvalidSource) {
			t.Fatalf("unchecked WorkspaceID admitted: %v", err)
		}
	}
}

func TestClosedHandleAndCanceledOpen(t *testing.T) {
	for _, h := range []*Handle{nil, {}} {
		if _, err := h.Get(context.Background(), resourceget.Request{ID: "task-1"}); !errors.Is(err, resourceget.ErrClosed) {
			t.Fatalf("closed handle: %v", err)
		}
		if err := h.Close(); err != nil {
			t.Fatal(err)
		}
		if err := h.Close(); err != nil {
			t.Fatal(err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := OpenExisting(ctx, Options{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled open: %v", err)
	}
}

// These fakes replace only the private ownership carrier and Reader role. The
// returned document is admitted by the same public parser as a real read.
type blockedCurrentReader struct {
	doc                        resourceget.NativeDocument
	entered, release, finished chan struct{}
}

func (r *blockedCurrentReader) Get(ctx context.Context, _ resourceget.Request) (resourceget.NativeDocument, error) {
	close(r.entered)
	defer close(r.finished)
	select {
	case <-r.release:
		return r.doc, nil
	case <-ctx.Done():
		return resourceget.NativeDocument{}, ctx.Err()
	}
}

type observedCurrentCloser struct {
	calls            atomic.Int32
	closedDuringRead atomic.Bool
	readerFinished   <-chan struct{}
	err              error
}

func (c *observedCurrentCloser) Close() error {
	c.calls.Add(1)
	select {
	case <-c.readerFinished:
	default:
		c.closedDuringRead.Store(true)
	}
	return c.err
}

func TestHandleCloseWaitsForInflightGet(t *testing.T) {
	cleanupFailure := errors.New("cleanup failed")
	for _, tc := range []struct {
		name       string
		cleanupErr error
	}{{name: "success"}, {name: "cleanup error", cleanupErr: cleanupFailure}} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			doc, err := resourceget.ParseLegacyCurrent("https://example.test/get/", resourceget.Bead, []byte(`{"id":"https://example.test/get/task-1","type":"https://example.test/get/memory","revision":"edition-1","properties":{"body":"kept"},"metadata":{}}`))
			if err != nil {
				t.Fatal(err)
			}
			reader := &blockedCurrentReader{doc: doc, entered: make(chan struct{}), release: make(chan struct{}), finished: make(chan struct{})}
			closer := &observedCurrentCloser{readerFinished: reader.finished, err: tc.cleanupErr}
			h := &Handle{store: closer, reader: reader}
			type getResult struct {
				doc resourceget.NativeDocument
				err error
			}
			got := make(chan getResult, 1)
			go func() { doc, err := h.Get(ctx, resourceget.Request{ID: "task-1"}); got <- getResult{doc, err} }()
			select {
			case <-reader.entered:
			case <-ctx.Done():
				t.Fatal("Get did not enter the reader")
			}
			closed := make(chan error, 1)
			go func() { closed <- h.Close() }()
			// Wait until Close is actually queued for ownership, rather than
			// inferring blocking from a scheduler-dependent sleep. An RWMutex
			// refuses a new read lock once its writer is pending.
			for h.mu.TryRLock() {
				h.mu.RUnlock()
				select {
				case err := <-closed:
					t.Fatalf("Close returned before in-flight Get: %v", err)
				case <-ctx.Done():
					t.Fatal("Close did not reach the in-flight read")
				default:
					runtime.Gosched()
				}
			}
			if closer.calls.Load() != 0 {
				t.Fatal("underlying source closed during Get")
			}
			close(reader.release)
			select {
			case result := <-got:
				if result.err != nil || result.doc.ID() != doc.ID() || !bytes.Equal(result.doc.NativeJSON(), doc.NativeJSON()) {
					t.Fatalf("in-flight Get did not finish intact: %+v", result)
				}
			case <-ctx.Done():
				t.Fatal("in-flight Get did not finish")
			}
			select {
			case err := <-closed:
				if !errors.Is(err, tc.cleanupErr) {
					t.Fatalf("Close lost cleanup error: %v", err)
				}
			case <-ctx.Done():
				t.Fatal("Close did not finish")
			}
			if closer.calls.Load() != 1 || closer.closedDuringRead.Load() {
				t.Fatal("cleanup ran more than once or before Get finished")
			}
			// Concurrent repeated Close/Get calls exercise the sealed state under
			// -race, including when the first underlying cleanup returned an error.
			var callers sync.WaitGroup
			for range 8 {
				callers.Add(1)
				go func() {
					defer callers.Done()
					if _, err := h.Get(ctx, resourceget.Request{ID: "task-1"}); !errors.Is(err, resourceget.ErrClosed) {
						t.Errorf("post-Close Get: %v", err)
					}
					if err := h.Close(); err != nil {
						t.Errorf("repeated Close reran cleanup: %v", err)
					}
				}()
			}
			callers.Wait()
			if closer.calls.Load() != 1 {
				t.Fatal("repeated Close invoked cleanup again")
			}
		})
	}
}
