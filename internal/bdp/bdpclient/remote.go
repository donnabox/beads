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
	"strconv"
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

// Capabilities reads discovery LIVE — a get of the well-known /sys/capabilities
// resource (core.lumen §Capabilities), decoded through the wire-DTO layer
// (discovery.go), which absorbs the clean-room's documented shape divergences
// from the ratified document. A caller reads it once and relies on it: a
// primitive is offered iff its capability is present. See [RemoteScope.Discover]
// for the full connect-time conformance handshake (descriptor + negotiation +
// capabilities).
func (c *RemoteScope) Capabilities(ctx context.Context) (bdp.Capabilities, error) {
	return c.fetchCapabilities(ctx)
}

// List reads a window of a namespace's current resources — GET /-/list, with the
// filter pushed to the server as the `where=` query and paging as after/limit.
// NOTE: the clean-room's list returns a Page of IDS, not full representations, so
// the returned Representations carry only their ID (Body empty); a caller that
// needs the bodies resolves each with Get. The filter is the flat-AND low bar:
// each condition is a root-addressed Compare (see toWireConditions); a server
// that cannot evaluate a condition faults, which surfaces here rather than being
// silently widened.
func (c *RemoteScope) List(ctx context.Context, _ bdp.Namespace, where bdp.ListFilter, _ *bdp.Selector, req *bdp.ReadRequest) (bdp.Page[bdp.Representation], error) {
	var params []string
	if len(where) > 0 {
		wj, err := json.Marshal(toWireConditions(where))
		if err != nil {
			return bdp.Page[bdp.Representation]{}, fmt.Errorf("bdpclient: encode list filter: %w", err)
		}
		params = append(params, "where="+queryEscape(string(wj)))
	}
	if req != nil {
		if req.After != "" {
			params = append(params, "after="+queryEscape(string(req.After)))
		}
		if req.Limit > 0 {
			params = append(params, "limit="+strconv.Itoa(req.Limit))
		}
	}
	path := "/-/list"
	if len(params) > 0 {
		path += "?" + strings.Join(params, "&")
	}
	resp, body, err := c.do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return bdp.Page[bdp.Representation]{}, err
	}
	if resp.StatusCode/100 != 2 {
		return bdp.Page[bdp.Representation]{}, errorFrom(resp.StatusCode, body)
	}
	var wp wirePage[wireResourceID]
	if err := json.Unmarshal(body, &wp); err != nil {
		return bdp.Page[bdp.Representation]{}, fmt.Errorf("bdpclient: bad list page: %w", err)
	}
	page := bdp.Page[bdp.Representation]{Head: wp.Head}
	for _, it := range wp.Items {
		page.Items = append(page.Items, bdp.Resource{ID: it.toBDP()})
	}
	if wp.Cursor != nil {
		page.Cursor = bdp.Cursor(*wp.Cursor)
	}
	return page, nil
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
	var wr wireRep
	if err := json.Unmarshal(body, &wr); err != nil {
		return bdp.Representation{}, fmt.Errorf("bdpclient: bad representation: %w", err)
	}
	rep := wr.toBDP()
	if rev := revFromETag(resp); rev != "" && rep.ID.Rev == "" {
		rep.ID.Rev = rev
	}
	return rep, nil
}

// Create admits a new bead and returns its pinned reference. The rc2 create
// route is POST /-/create under the action fence (bare GET / is the
// ServiceDescriptor); the proposed id rides the query string (?id=), not the
// body (core.lumen §create).
func (c *RemoteScope) Create(ctx context.Context, in bdp.CreateInput) (bdp.ResourceID, error) {
	path := "/-/create"
	proposed := proposedID(in.ID)
	if proposed != "" {
		path += "?id=" + url.QueryEscape(proposed)
	}
	body := wireCreateBody{
		Authority: authorityRef(in.Authority),
		Type:      in.Type,
		Body:      nonEmpty(in.Body),
		Metadata:  nonEmpty(in.Metadata),
	}
	return c.write(ctx, http.MethodPost, path, body, "", bdp.ResourceID{ID: proposed})
}

// Put is the idempotent full-state write (PUT /{id}) and returns the written
// pinned reference. The body is exactly the representation; guards ride the
// request line (core.lumen §put).
func (c *RemoteScope) Put(ctx context.Context, in bdp.PutInput) (bdp.ResourceID, error) {
	body := wireCreateBody{
		Authority: authorityRef(in.Authority),
		Type:      in.Type,
		Body:      nonEmpty(in.Body),
		Metadata:  nonEmpty(in.Metadata),
	}
	return c.write(ctx, http.MethodPut, c.resourcePath(in.Target), body, string(in.IfRevision), in.Target)
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

// Call invokes an Operation — POST /{operation}/-/call under the action fence.
// The operation is addressed by its element path (an intrinsic op resolves to
// that path; the client does not materialize an Operation bead first). A sync
// 200/201 carries the Outcome; a 409 carries the declared domain failure
// (Outcome.Failure). A PRE-dispatch fault (unresolvable op, malformed request,
// unauthorized) is a bare Problem with NO Outcome and surfaces here as an error
// (core.lumen §call @ bdp-core-0.2.3-rc2).
//
// NOTE: no clean-room oracle — cleanroom-v023 answers every call "not offered",
// so this is built from the spec and exercised only by mock-server tests. It is
// NOT live-validated; proving it needs a call-offering, issue-pack-serving BDP
// server (a cross-track dependency).
func (c *RemoteScope) Call(ctx context.Context, operation bdp.ResourceID, req bdp.Request) (bdp.Reply, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return bdp.Reply{}, fmt.Errorf("bdpclient: marshal call request: %w", err)
	}
	path := "/" + strings.Trim(operation.ID, "/") + "/-/call"
	resp, body, err := c.do(ctx, http.MethodPost, path, payload, func(h http.Header) {
		h.Set("Content-Type", "application/json")
	})
	if err != nil {
		return bdp.Reply{}, err
	}
	// 2xx (success) and 409 (declared domain failure) both carry a Reply body;
	// any other non-2xx is a pre-dispatch Problem with no Outcome.
	if resp.StatusCode/100 == 2 || resp.StatusCode == http.StatusConflict {
		var reply bdp.Reply
		if err := json.Unmarshal(body, &reply); err != nil {
			return bdp.Reply{}, fmt.Errorf("bdpclient: bad call reply: %w", err)
		}
		return reply, nil
	}
	return bdp.Reply{}, errorFrom(resp.StatusCode, body)
}

// --- plumbing ---

// write encodes req as a create/put body, sends it, and decodes the pinned ref.
func (c *RemoteScope) write(ctx context.Context, method, path string, req wireCreateBody, ifRevision string, fallback bdp.ResourceID) (bdp.ResourceID, error) {
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

// queryEscape percent-encodes a query VALUE with %20 for spaces, per RFC 3986.
// url.Values.Encode()/url.QueryEscape emit "+" for a space, but the BDP server
// percent-decodes only (a "+" stays a literal "+"), so a filter value with a
// space must be %20-encoded or it silently mismatches (or breaks JSON parsing).
func queryEscape(s string) string {
	return strings.ReplaceAll(url.QueryEscape(s), "+", "%20")
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

// authorityRef maps an optional bdp.Authority (a ResourceID) to the wire ref.
func authorityRef(a *bdp.Authority) *wireResourceID {
	if a == nil {
		return nil
	}
	w := toWireID(*a)
	return &w
}

func proposedID(p bdp.IDProposal) string {
	if p.Exact != "" {
		return p.Exact
	}
	return p.Prefix
}
