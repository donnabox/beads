// Package bdpclient is bd's BDP v0.2 client — a [bdp.Scope] implemented over
// HTTP. [RemoteScope] is the Go analog of the clean-room's RemoteScope: a remote
// store substitutes for a local one behind the same contract, so a caller reads
// a remote scope exactly as it would a local trait object.
//
// The spike maps the single-resource core — get, create, put, patch, delete.
// Every other primitive inherits the refusing default from [bdp.UnimplementedScope]
// (the clean-room server answers them 501), and [RemoteScope.Capabilities]
// advertises exactly the mapped set.
//
// This package deliberately does NOT touch bd's command wiring; it is a
// standalone library (Phase 0 ruling: no `bd connect` surface yet).
//
// EXPERIMENTAL and pre-release; see package bdp.
package bdpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/steveyegge/beads/internal/bdp"
)

// RemoteScope is a [bdp.Scope] backed by a remote BDP store over HTTP.
type RemoteScope struct {
	bdp.UnimplementedScope
	base   string
	client *http.Client
}

// New returns a RemoteScope talking to the store at base, an origin such as
// "http://127.0.0.1:8799" (no trailing slash required). It uses http.DefaultClient.
func New(base string) *RemoteScope {
	return NewWithClient(base, http.DefaultClient)
}

// NewWithClient is New with a caller-supplied http.Client (timeouts, transport).
func NewWithClient(base string, client *http.Client) *RemoteScope {
	if client == nil {
		client = http.DefaultClient
	}
	return &RemoteScope{
		UnimplementedScope: bdp.UnimplementedScope{Backend: "bdp remote client"},
		base:               strings.TrimRight(base, "/"),
		client:             client,
	}
}

var _ bdp.Scope = (*RemoteScope)(nil)

// Capabilities advertises the primitives this spike client maps and the
// scope-wide traits it knows. On the wire the spec reads discovery as a get of
// the well-known /sys/capabilities resource; this spike returns a STATIC
// advertisement instead, because the clean-room server does not yet serve that
// resource in a merged shape. The clean-room store is versioned — writes mint
// pinned revs — so RevisionModel is versioned. Fetching /sys/capabilities live
// is the follow-up once the server grows it (see package bdp SPELUNK.md).
func (c *RemoteScope) Capabilities(context.Context) (bdp.Capabilities, error) {
	return bdp.Capabilities{
		Core: bdp.CoreCapabilities{
			Get:    bdp.GetCaps{},
			Create: &bdp.CreateCaps{ClientProposedID: true, ServerAllocatesID: true},
			Put:    &bdp.PutCaps{IfRevision: true},
			Patch:  &bdp.PatchCaps{IfRevision: true},
			Delete: &bdp.DeleteCaps{IfRevision: true},
		},
		Scope: bdp.ScopeCapabilities{
			StorageClasses: []bdp.StorageClass{bdp.StorageClassDurable},
			RevisionModel:  bdp.RevisionModelVersioned,
			IdModel:        bdp.IdModelPermanent,
		},
	}, nil
}

// Get resolves ref to its representation. GET /{id} (+ ?revision=). The selector
// is accepted but not sent — the spike server does not project.
func (c *RemoteScope) Get(ctx context.Context, ref bdp.ResourceID, _ *bdp.Selector) (bdp.Representation, error) {
	resp, body, err := c.do(ctx, http.MethodGet, c.resourcePath(ref), nil, nil)
	if err != nil {
		return bdp.Representation{}, err
	}
	if resp.StatusCode/100 != 2 {
		return bdp.Representation{}, errorFrom(resp.StatusCode, body)
	}
	var wr wireRepresentation
	if err := json.Unmarshal(body, &wr); err != nil {
		return bdp.Representation{}, fmt.Errorf("bdpclient: bad representation: %w", err)
	}
	rep := wr.toBDP()
	if rev := revFromETag(resp); rev != "" && rep.ID.Rev == "" {
		rep.ID.Rev = rev
	}
	return rep, nil
}

// Create admits a new bead (POST /) and returns its pinned reference.
func (c *RemoteScope) Create(ctx context.Context, in bdp.CreateInput) (bdp.ResourceID, error) {
	req := wireCreateRequest{
		Kind:       kindOf(in.Type),
		Type:       typeRef(in.Type, in.Authority),
		Properties: nonEmpty(in.Body),
		Metadata:   nonEmpty(in.Metadata),
	}
	if id := proposedID(in.ID); id != "" {
		req.ID = &wireResourceID{ID: id}
	}
	return c.write(ctx, http.MethodPost, "/", req, "", bdp.ResourceID{ID: proposedID(in.ID)})
}

// Put is the idempotent full-state write (PUT /{id}) and returns the written
// pinned reference.
func (c *RemoteScope) Put(ctx context.Context, in bdp.PutInput) (bdp.ResourceID, error) {
	req := wireCreateRequest{
		Kind:       kindOf(in.Type),
		Type:       typeRef(in.Type, in.Authority),
		Properties: nonEmpty(in.Body),
		Metadata:   nonEmpty(in.Metadata),
	}
	return c.write(ctx, http.MethodPut, c.resourcePath(in.Target), req, string(in.IfRevision), in.Target)
}

// Patch applies a JSON Patch to /body and /metadata (PATCH /{id}) and returns
// the patched pinned reference. The precondition state guard is accepted but not
// yet sent — the spike server does not parse ?precondition= (wiring it is Track
// A), matching how the selector is accepted but not projected.
func (c *RemoteScope) Patch(ctx context.Context, ref bdp.ResourceID, patch bdp.JSONPatch, ifRevision bdp.Rev, _ bdp.ListFilter) (bdp.ResourceID, error) {
	resp, body, err := c.do(ctx, http.MethodPatch, c.resourcePath(ref), patch, func(h http.Header) {
		h.Set("Content-Type", "application/json-patch+json")
		if ifRevision != "" {
			h.Set("If-Match", string(ifRevision))
		}
	})
	if err != nil {
		return bdp.ResourceID{}, err
	}
	if resp.StatusCode/100 != 2 {
		return bdp.ResourceID{}, errorFrom(resp.StatusCode, body)
	}
	return writeResult(resp, body, ref)
}

// Delete retires a resource (DELETE /{id}), returning the tombstone revision.
// The precondition state guard is accepted but not yet sent (see Patch).
func (c *RemoteScope) Delete(ctx context.Context, ref bdp.ResourceID, ifRevision bdp.Rev, _ bdp.ListFilter) (bdp.Rev, error) {
	resp, body, err := c.do(ctx, http.MethodDelete, c.resourcePath(ref), nil, func(h http.Header) {
		if ifRevision != "" {
			h.Set("If-Match", string(ifRevision))
		}
	})
	if err != nil {
		return "", err
	}
	if resp.StatusCode/100 != 2 {
		return "", errorFrom(resp.StatusCode, body)
	}
	tombstone, err := writeResult(resp, body, ref)
	if err != nil {
		return "", err
	}
	return tombstone.Rev, nil
}

// --- plumbing ---

// write encodes req as a create/put body, sends it, and decodes the pinned ref.
func (c *RemoteScope) write(ctx context.Context, method, path string, req wireCreateRequest, ifRevision string, fallback bdp.ResourceID) (bdp.ResourceID, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return bdp.ResourceID{}, fmt.Errorf("bdpclient: marshal request: %w", err)
	}
	resp, body, err := c.do(ctx, method, path, payload, func(h http.Header) {
		h.Set("Content-Type", "application/json")
		if ifRevision != "" {
			h.Set("If-Match", ifRevision)
		}
	})
	if err != nil {
		return bdp.ResourceID{}, err
	}
	if resp.StatusCode/100 != 2 {
		return bdp.ResourceID{}, errorFrom(resp.StatusCode, body)
	}
	return writeResult(resp, body, fallback)
}

// writeResult decodes a write response: the pinned ref rides in the body (a
// wireResourceID), and the minted rev rides in the quoted ETag header, preferred
// over the body's own rev.
func writeResult(resp *http.Response, body []byte, fallback bdp.ResourceID) (bdp.ResourceID, error) {
	var wref wireResourceID
	if len(bytes.TrimSpace(body)) > 0 {
		if err := json.Unmarshal(body, &wref); err != nil {
			return bdp.ResourceID{}, fmt.Errorf("bdpclient: bad write response: %w", err)
		}
	}
	ref := fallback
	if wref.ID != "" {
		ref = wref.toBDP()
	}
	if rev := revFromETag(resp); rev != "" {
		ref.Rev = rev
	}
	return ref, nil
}

// do issues one request and reads the whole body.
func (c *RemoteScope) do(ctx context.Context, method, path string, body []byte, setHeaders func(http.Header)) (*http.Response, []byte, error) {
	var rdr io.Reader
	if body != nil {
		rdr = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.base+path, rdr)
	if err != nil {
		return nil, nil, fmt.Errorf("bdpclient: build request: %w", err)
	}
	if setHeaders != nil {
		setHeaders(req.Header)
	}
	resp, err := c.client.Do(req) //nolint:gosec // G704: base is the caller-configured store origin, not user-tainted input
	if err != nil {
		return nil, nil, fmt.Errorf("bdpclient: transport: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("bdpclient: read body: %w", err)
	}
	return resp, data, nil
}

func (c *RemoteScope) resourcePath(ref bdp.ResourceID) string {
	p := "/" + url.PathEscape(ref.ID)
	if ref.Rev != "" {
		p += "?revision=" + url.QueryEscape(string(ref.Rev))
	}
	return p
}

// errorFrom maps a non-2xx response to a classifiable error: a FaultError
// carrying the server's Problem, chained to the matching shared sentinel.
func errorFrom(status int, body []byte) error {
	var wp wireProblem
	_ = json.Unmarshal(body, &wp)
	problem := wp.toProblem()
	if problem.Status == 0 {
		problem.Status = status
	}
	var sentinel error
	switch status {
	case http.StatusNotFound, http.StatusGone: // 404, 410 (absent or erased)
		sentinel = bdp.ErrNotFound
	case http.StatusPreconditionFailed: // 412
		sentinel = bdp.ErrRevisionMismatch
	case http.StatusConflict: // 409
		sentinel = bdp.ErrAlreadyExists
	case http.StatusBadRequest, http.StatusUnprocessableEntity: // 400, 422
		sentinel = bdp.ErrValidation
	}
	return bdp.NewFaultError(problem, sentinel)
}

func revFromETag(resp *http.Response) bdp.Rev {
	return bdp.Rev(strings.Trim(resp.Header.Get("ETag"), `"`))
}

// kindOf derives the wire `kind` (the type's leaf) from a spec type path:
// "issue/task" → "task", "task" → "task".
func kindOf(typePath string) string {
	if i := strings.LastIndex(typePath, "/"); i >= 0 {
		return typePath[i+1:]
	}
	return typePath
}

// typeRef builds the wire type reference. The clean-room models a bead's type as
// a ref to a type bead (e.g. {"id":"BeadType/task"}); prefer the caller's
// Authority when given, else synthesize "BeadType/<kind>".
func typeRef(typePath string, authority *bdp.Authority) wireResourceID {
	if authority != nil {
		return toWireID(*authority)
	}
	return wireResourceID{ID: "BeadType/" + kindOf(typePath)}
}

func proposedID(p bdp.IDProposal) string {
	if p.Exact != "" {
		return p.Exact
	}
	return p.Prefix
}
