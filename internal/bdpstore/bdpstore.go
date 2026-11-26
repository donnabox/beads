// Package bdpstore adapts a BDP client ([bdp.Scope], e.g.
// bdpclient.RemoteScope) to the beads issueops role interfaces, mapping the
// Issue domain onto BDP's generic resources. It is the client-end of Track A's
// `bd connect` work: the CRUD slice (Reader.Get, Lifecycle.Create,
// Deleter.Delete) translated onto BDP get/create/delete.
//
// SCOPE BOUNDARY (Donna's ruling): this adapter deliberately does NOT satisfy
// storage.DoltStorage and is NOT registered as a backend — the cmd/bd-crossing
// wiring (store_factory registration + connect --protocol persist) is held. The
// adapter is proven end-to-end by exercising it directly against a live BDP
// server in a test, never by routing a live bd command through it. The
// remaining role methods (list/ready/update/close/reopen) fault unsupported
// until built out (M-A3+); the lifecycle→BDP mapping for those is the question
// with the spec owner.
package bdpstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/steveyegge/beads/beadserrors"
	"github.com/steveyegge/beads/internal/bdp"
	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/issueops"
)

// Store adapts a bdp.Scope to the beads issueops roles.
type Store struct {
	scope bdp.Scope
}

// New returns a Store backed by scope (typically a bdpclient.RemoteScope).
func New(scope bdp.Scope) *Store { return &Store{scope: scope} }

// IssueReader returns the read role (Get implemented; Ready/List unsupported).
func (s *Store) IssueReader() (issueops.Reader, error) { return bdpReader{s}, nil }

// IssueLifecycle returns the write role (Create implemented; the rest unsupported).
func (s *Store) IssueLifecycle() (issueops.Lifecycle, error) { return bdpLifecycle{s}, nil }

// Deleter returns the delete role.
func (s *Store) Deleter() (issueops.Deleter, error) { return bdpDeleter{s}, nil }

func (s *Store) unsupported(op string) error {
	return &beadserrors.ErrUnsupported{Op: op, Backend: "bdp adapter"}
}

// --- Issue <-> BDP domain mapping ---

// typePathOf renders a beads IssueType as a BDP bead type path ("issue/task").
func typePathOf(it types.IssueType) string {
	if it == "" {
		return "issue/task"
	}
	return "issue/" + string(it)
}

// issueToBody marshals an Issue as the opaque BDP bead body. The envelope
// (id/rev) rides the BDP ResourceID, not the body.
func issueToBody(iss *types.Issue) (json.RawMessage, error) {
	b, err := json.Marshal(iss)
	if err != nil {
		return nil, fmt.Errorf("bdpstore: marshal issue: %w", err)
	}
	return b, nil
}

// repToIssue decodes a BDP representation into an Issue, restoring the id from
// the envelope and stitching RowVersion from the BDP revision token when it is
// numeric (Issue.RowVersion is json:"-", so it never rides the body).
func repToIssue(rep bdp.Representation) (*types.Issue, error) {
	var iss types.Issue
	if len(rep.Body) > 0 {
		if err := json.Unmarshal(rep.Body, &iss); err != nil {
			return nil, fmt.Errorf("bdpstore: decode issue body: %w", err)
		}
	}
	iss.ID = rep.ID.ID
	if rv, err := strconv.ParseInt(string(rep.ID.Rev), 10, 64); err == nil {
		iss.RowVersion = rv
	}
	return &iss, nil
}

// --- Reader ---

type bdpReader struct{ s *Store }

var _ issueops.Reader = bdpReader{}

func (r bdpReader) Get(ctx context.Context, req issueops.GetRequest) (*issueops.IssueDetails, error) {
	rep, err := r.s.scope.Get(ctx, bdp.ResourceID{ID: req.ID}, nil)
	if err != nil {
		if errors.Is(err, bdp.ErrNotFound) {
			return nil, issueops.ErrNotFound
		}
		return nil, err
	}
	iss, err := repToIssue(rep)
	if err != nil {
		return nil, err
	}
	return &issueops.IssueDetails{Issue: *iss}, nil
}

// Ready is "unblocked open work in policy order" — it needs dependency-graph
// awareness (is_blocked) that BDP Core's flat list cannot express. That is a
// query-pack capability, not Core, so Ready stays unsupported until that surface
// exists.
func (r bdpReader) Ready(context.Context, issueops.ReadyRequest) (issueops.IssuePage, error) {
	return issueops.IssuePage{}, r.s.unsupported("ready")
}

// List lists the scope's current resources and hydrates each into an Issue. The
// clean-room's list returns ids only, so each id is resolved with Get.
//
// Slice limitations (documented, not silent): server-side filter push-down is
// not yet wired, so an explicit Status / IssueType filter is applied
// CLIENT-SIDE here, and beads' DEFAULT status exclusions are NOT replicated — an
// unset Status lists every current resource. Full filter parity waits on
// server-side push-down and the query/ready roles.
func (r bdpReader) List(ctx context.Context, req issueops.ListRequest) (issueops.IssuePage, error) {
	limit := 0
	if req.Limit != nil {
		limit = *req.Limit
	}
	page, err := r.s.scope.List(ctx, bdp.Namespace{}, nil, nil, nil)
	if err != nil {
		return issueops.IssuePage{}, err
	}
	items := make([]*types.IssueWithCounts, 0, len(page.Items))
	for _, rep := range page.Items {
		got, err := r.s.scope.Get(ctx, rep.ID, nil)
		if err != nil {
			if errors.Is(err, bdp.ErrNotFound) {
				continue // raced with a delete between list and hydrate
			}
			return issueops.IssuePage{}, err
		}
		iss, err := repToIssue(got)
		if err != nil {
			return issueops.IssuePage{}, err
		}
		if !matchesList(iss, req) {
			continue
		}
		if limit > 0 && len(items) >= limit {
			return issueops.IssuePage{Items: items, HasMore: true}, nil
		}
		items = append(items, &types.IssueWithCounts{Issue: iss})
	}
	return issueops.IssuePage{Items: items}, nil
}

// matchesList applies the explicit Status / IssueType filters client-side (each
// a comma-OR set); an unset filter matches everything.
func matchesList(iss *types.Issue, req issueops.ListRequest) bool {
	if req.Status != "" && !inCSV(string(iss.Status), req.Status) {
		return false
	}
	if req.IssueType != "" && !inCSV(string(iss.IssueType), req.IssueType) {
		return false
	}
	return true
}

func inCSV(val, csv string) bool {
	for _, p := range strings.Split(csv, ",") {
		if strings.EqualFold(strings.TrimSpace(p), val) {
			return true
		}
	}
	return false
}

// --- Lifecycle ---

type bdpLifecycle struct{ s *Store }

var _ issueops.Lifecycle = bdpLifecycle{}

func (l bdpLifecycle) Create(ctx context.Context, req issueops.CreateRequest) (issueops.CreateResult, error) {
	if req.Issue == nil {
		return issueops.CreateResult{}, fmt.Errorf("bdpstore: create requires an issue")
	}
	iss := *req.Issue
	if iss.CreatedBy == "" {
		iss.CreatedBy = req.Actor
	}
	body, err := issueToBody(&iss)
	if err != nil {
		return issueops.CreateResult{}, err
	}
	ref, err := l.s.scope.Create(ctx, bdp.CreateInput{
		ID:   bdp.IDProposal{Exact: iss.ID},
		Type: typePathOf(iss.IssueType),
		Body: body,
	})
	if err != nil {
		if errors.Is(err, bdp.ErrAlreadyExists) {
			return issueops.CreateResult{}, issueops.ErrAlreadyExists
		}
		return issueops.CreateResult{}, err
	}
	out := iss
	out.ID = ref.ID
	if rv, perr := strconv.ParseInt(string(ref.Rev), 10, 64); perr == nil {
		out.RowVersion = rv
	}
	return issueops.CreateResult{Issue: &out}, nil
}

func (l bdpLifecycle) Update(context.Context, issueops.UpdateRequest) (issueops.UpdateResult, error) {
	return issueops.UpdateResult{}, l.s.unsupported("update")
}

func (l bdpLifecycle) Close(context.Context, issueops.CloseRequest) (issueops.CloseResult, error) {
	return issueops.CloseResult{}, l.s.unsupported("close")
}

func (l bdpLifecycle) Reopen(context.Context, issueops.ReopenRequest) (issueops.ReopenResult, error) {
	return issueops.ReopenResult{}, l.s.unsupported("reopen")
}

// --- Deleter ---

type bdpDeleter struct{ s *Store }

var _ issueops.Deleter = bdpDeleter{}

func (d bdpDeleter) Delete(ctx context.Context, req issueops.DeleteRequest) (issueops.DeleteResult, error) {
	if len(req.IDs) == 0 {
		return issueops.DeleteResult{}, issueops.ErrValidation
	}
	var ifRev bdp.Rev
	if req.ExpectedVersion != nil {
		ifRev = bdp.Rev(strconv.FormatInt(*req.ExpectedVersion, 10))
	}
	deleted := make([]string, 0, len(req.IDs))
	for _, id := range req.IDs {
		if _, err := d.s.scope.Delete(ctx, bdp.ResourceID{ID: id}, ifRev, nil); err != nil {
			if errors.Is(err, bdp.ErrNotFound) {
				return issueops.DeleteResult{}, issueops.ErrNotFound
			}
			return issueops.DeleteResult{}, err
		}
		deleted = append(deleted, id)
	}
	return issueops.DeleteResult{Deleted: len(deleted), DeletedIDs: deleted}, nil
}
