// Package local acquires the current Get role from an explicitly configured,
// existing local workspace. Both backends require its real filesystem identity;
// this factory is not a hosted-engine constructor or an initialization API.
package local

import (
	"context"
	"errors"
	"io"
	"sync"

	"github.com/steveyegge/beads/internal/httpapi/graphread"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/resourceget"
)

type Backend uint8

const (
	Embedded Backend = iota + 1
	Server
)

// ServerOptions supplies explicit connection settings without ambient lookup.
// Password must be treated as a secret by the caller.
type ServerOptions struct {
	Host     string
	Port     int
	Socket   string
	User     string
	Password string `json:"-"`
	TLS      bool
}

// Options describes the exact existing workspace and persisted authority. It
// never creates, relocates or migrates one. DataDir applies to Embedded only.
// The branch is main, the only branch supported by the existing graph store.
type Options struct {
	Backend     Backend
	WorkspaceID string
	ScopeURL    string
	AuthorityID string
	Database    string
	DataDir     string
	Server      ServerOptions
}

// Handle owns one checked source. Close waits for in-flight Get calls. There
// are no background processes or retries; callers control request cancellation.
type Handle struct {
	mu     sync.RWMutex
	store  io.Closer
	reader resourceget.Reader
}

var _ resourceget.Reader = (*Handle)(nil)

// OpenExisting performs the existing store's real WorkspaceID, Scope, authority,
// schema and installed-contract checks. It never calls Init or performs DDL.
// Access is the trusted local workspace view, not per-principal authorization.
func OpenExisting(ctx context.Context, options Options) (*Handle, error) {
	if ctx == nil {
		return nil, resourceget.ErrInvalidRequest
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	opts, err := storeOptions(options)
	if err != nil {
		return nil, err
	}
	store, err := graphstore.OpenExisting(ctx, opts)
	if err != nil {
		return nil, errors.Join(resourceget.ErrInvalidSource, err)
	}
	reader := graphread.New(store)
	if err := reader.ValidateAuthority(ctx); err != nil {
		return nil, errors.Join(resourceget.ErrInvalidSource, err, store.Close())
	}
	return &Handle{store: store, reader: reader}, nil
}

func storeOptions(o Options) (graphstore.Options, error) {
	backend := ""
	switch o.Backend {
	case Embedded:
		backend = "embedded"
	case Server:
		backend = "server"
	default:
		return graphstore.Options{}, resourceget.ErrInvalidRequest
	}
	return graphstore.Options{
		Backend: backend, Database: o.Database, DataDir: o.DataDir, Branch: "main",
		Binding:        graphstore.Binding{WorkspaceID: o.WorkspaceID, ScopeURL: o.ScopeURL, AuthorityID: o.AuthorityID, SchemaVersion: graphstore.SchemaVersion},
		TypeGeneration: graphstore.TypeGeneration,
		ServerHost:     o.Server.Host, ServerPort: o.Server.Port, ServerSocket: o.Server.Socket,
		ServerUser: o.Server.User, ServerPassword: o.Server.Password, ServerTLS: o.Server.TLS,
	}, nil
}

func (h *Handle) Get(ctx context.Context, request resourceget.Request) (resourceget.NativeDocument, error) {
	if ctx == nil {
		return resourceget.NativeDocument{}, resourceget.ErrInvalidRequest
	}
	if err := ctx.Err(); err != nil {
		return resourceget.NativeDocument{}, err
	}
	if h == nil {
		return resourceget.NativeDocument{}, resourceget.ErrClosed
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.store == nil || h.reader == nil {
		return resourceget.NativeDocument{}, resourceget.ErrClosed
	}
	return h.reader.Get(ctx, request)
}

// Close is idempotent. The source cannot be reused after Close, even if its
// underlying cleanup reports an error.
func (h *Handle) Close() error {
	if h == nil {
		return nil
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.store == nil {
		return nil
	}
	store := h.store
	h.store, h.reader = nil, nil
	return store.Close()
}
