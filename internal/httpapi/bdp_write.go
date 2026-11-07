package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"mime"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/storage/uow"
	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/issueops"
)

// The Write surface of the Beads Protocol, for the Issue and Dependency Types: the
// operation directory and the six targets it lists. See
// specs/profiles/write/write.md and specs/types/issue/issue.md.
//
// Like Core it holds no storage. It runs on the roles the Callables run on, one role
// call and so one transaction per request, and what it adds is the translation:
// Write's patch form onto native fields, native refusals onto Write's table, and a
// row back into a record. Each check that depends on the state a write lands on runs
// inside that write's transaction, through the role's InTx hook.
//
// THE LINK WRITES ARE NOT IMPLEMENTED, by Donna's ruling: `createLink`, `updateLink` and
// `deleteLink` wait until Links have identity and revisions comparable to Beads. The
// directory must list them (Write: "the six Resource members always"), so they are
// listed, answer operation-not-supported, and discovery says so.
const (
	coreOperationsPath = "/bdp/operations"

	opWriteDirectory  = "writeDirectory"
	opWriteCreateBead = "writeCreateBead"
	opWriteUpdateBead = "writeUpdateBead"
	opWriteDeleteBead = "writeDeleteBead"
	opWriteCreateLink = "writeCreateLink"
	opWriteUpdateLink = "writeUpdateLink"
	opWriteDeleteLink = "writeDeleteLink"

	writeOutcomeCreated = "created"
	writeOutcomeUpdated = "updated"
	writeOutcomeDeleted = "deleted"

	// writeBodyBytes bounds a request, as v0's create and update bound theirs: one
	// record can carry a description, a design and a metadata document at once.
	writeBodyBytes = maxUpdateBodyBytes

	// writeNoActor is who the native row records when a request names no actor.
	// Write makes the actor optional and the native roles require a name, and an
	// omission must not be recorded as anyone's. It is a native attribution string
	// and no BDP attribution: no record carries revisionProvenance here.
	writeNoActor = "bdp-write"

	// writeExceptionsCapability says what this Scope lists but does not serve.
	// Core has base clients ignore a capability they do not know, so it changes
	// nothing for them; a client that reads it can skip the refused call.
	writeExceptionsCapability        = "https://github.com/steveyegge/beads/bdp/capabilities/write-exceptions"
	writeExceptionsCapabilityVersion = "1"
)

// writeLinkExceptions are the directory members this Scope lists and does not serve.
var writeLinkExceptions = []string{"createLink", "updateLink", "deleteLink"}

// writeLinkExceptionDetail says why, and what changes a Link in the meantime.
func writeLinkExceptionDetail(member string) string {
	return "`" + member + "` is not implemented on this Scope: Link writes wait until Links have identity and revisions comparable to Beads. The `issue/dep/add` and `issue/dep/remove` Callables change Links."
}

// writeResult is the closed mutation result. There is no ownerRevisions: preview-1
// of the Issue package declares no ownership (Issue §6).
type writeResult struct {
	Outcome  string        `json:"outcome"`
	Changed  bool          `json:"changed"`
	Resource any           `json:"resource,omitempty"`
	Deleted  *writeDeleted `json:"deleted,omitempty"`
}

type writeDeleted struct {
	ResourceKind string           `json:"resourceKind"`
	Resource     writeDeletedName `json:"resource"`
}

type writeDeletedName struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Revision string `json:"revision"`
}

// writeRefusal carries a Problem out of a role's transaction, which rolls back on
// any error, to the handler that answers it.
type writeRefusal struct{ problem *callableProblem }

func (e *writeRefusal) Error() string { return "write refused: " + e.problem.Code }

// writeCall is one admitted request.
type writeCall struct {
	s       *Server
	w       http.ResponseWriter
	r       *http.Request
	ctx     context.Context
	p       *coreProjector
	members map[string]json.RawMessage
}

// writeOperationTargets are the directory's six members, in the order Write lists
// them, with the local name each is served at.
func (s *Server) writeOperationTargets() []struct {
	member, op string
	serve      http.HandlerFunc
} {
	return []struct {
		member, op string
		serve      http.HandlerFunc
	}{
		{"createBead", opWriteCreateBead, s.writeTarget(createBeadMembers, serveCreateBead)},
		{"updateBead", opWriteUpdateBead, s.writeTarget(updateBeadMembers, serveUpdateBead)},
		{"deleteBead", opWriteDeleteBead, s.writeTarget(deleteBeadMembers, serveDeleteBead)},
		{"createLink", opWriteCreateLink, serveLinkException("createLink")},
		{"updateLink", opWriteUpdateLink, serveLinkException("updateLink")},
		{"deleteLink", opWriteDeleteLink, serveLinkException("deleteLink")},
	}
}

// mountWrite registers the directory and its targets beside Core's routes. The rest
// of bdp/operations/ stays the catch-all's uniform 404.
func (s *Server) mountWrite(mux *http.ServeMux) {
	mux.Handle("GET "+coreOperationsPath, s.coreGate(opWriteDirectory, true, s.serveWriteDirectory))
	mux.HandleFunc(coreOperationsPath, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Allow", "GET, HEAD")
		w.WriteHeader(http.StatusMethodNotAllowed)
	})
	for _, target := range s.writeOperationTargets() {
		path := coreOperationsPath + "/" + target.member
		mux.Handle(http.MethodPost+" "+path, s.coreGate(target.op, false, target.serve))
		mux.Handle(path, methodNotAllowed(http.MethodPost))
	}
}

// serveWriteDirectory answers GET of the directory. Like discovery it is the
// addresses the caller used and constants of this build, so it asks no credential.
func (s *Server) serveWriteDirectory(w http.ResponseWriter, r *http.Request) {
	if len(r.URL.RawQuery) != 0 {
		writeCoreProblem(w, r, coreCodeInvalidParameter, "the operation directory takes no query parameters")
		return
	}
	if !writeAccepts(r) {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}
	scope := coreScope(r)
	setCoreNavigation(w.Header(), scope)
	if coreNotModified(w, r, "") {
		return
	}
	directory := map[string]string{}
	for _, target := range s.writeOperationTargets() {
		directory[target.member] = scope + strings.TrimPrefix(coreOperationsPath, "/") + "/" + target.member
	}
	writeCoreJSON(w, directory)
}

// serveLinkException is Donna's exception, for each Link write. It reads no body and
// no state. Write prescribes `operation-not-supported` for a recognized operation a
// provider does not serve only inside atomic batch, so that code, its family, status and
// retry are what is answered, with no batch pointer: there is no /operations/I here.
func serveLinkException(member string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !admitWriteMedia(w, r) {
			return
		}
		problem := newCallableProblem(bdpOperationNotSupported, writeLinkExceptionDetail(member))
		problem.Operation = member
		writeWriteProblem(w, r, problem)
	}
}

// writeAccepts is Core's Accept rule for the one media type every answer here offers.
func writeAccepts(r *http.Request) bool {
	return acceptsMedia(r.Header.Values("Accept"))
}

// admitWriteMedia is the part of admission that reads nothing: Write's 415 before the
// body is interpreted, and its bodyless 406.
func admitWriteMedia(w http.ResponseWriter, r *http.Request) bool {
	if media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || media != bdpMediaType {
		writeWriteProblem(w, r, newCallableProblem(bdpUnsupportedMedia, "a mutation's Content-Type is "+bdpMediaType))
		return false
	}
	if !writeAccepts(r) {
		w.Header().Set("Cache-Control", "private, no-store")
		w.WriteHeader(http.StatusNotAcceptable)
		return false
	}
	return true
}

// writeWriteProblem answers one Problem. Its shape is the shared table's, and Core's
// after-delay rows carry Retry-After.
func writeWriteProblem(w http.ResponseWriter, r *http.Request, problem *callableProblem) {
	requestInfo(r.Context()).code = Code(problem.Code)
	h := w.Header()
	h.Set("Content-Type", problemMediaType)
	h.Set("Cache-Control", "private, no-store")
	if problem.Retry == retryAfterDelay {
		h.Set("Retry-After", strconv.Itoa(coreRetryAfterSeconds))
	}
	w.WriteHeader(problem.Status)
	_ = json.NewEncoder(w).Encode(problem)
}

// writeTarget admits a request up to its operation: media, an empty query, the
// database slot, and one closed JSON object with no duplicate member at any depth.
func (s *Server) writeTarget(closed writeMembers, serve func(*writeCall) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !admitWriteMedia(w, r) {
			return
		}
		if r.URL.RawQuery != "" {
			writeWriteProblem(w, r, newCallableProblem(bdpInvalidParameter, "a mutation target takes no query parameters"))
			return
		}
		release, ok := s.coreAcquire(w, r)
		if !ok {
			return
		}
		defer release()

		members, problem := readWriteBody(w, r, closed)
		if problem != nil {
			writeWriteProblem(w, r, problem)
			return
		}
		call := &writeCall{s: s, w: w, r: r, ctx: r.Context(), p: newCoreProjector(s, r), members: members}
		if err := serve(call); err != nil {
			s.failWrite(w, r, err)
		}
	}
}

// writeMembers is one target's closed vocabulary.
type writeMembers struct {
	required, optional []string
}

var (
	createBeadMembers = writeMembers{[]string{"type"}, []string{"id", "properties", "metadata", "revisionProvenance"}}
	updateBeadMembers = writeMembers{[]string{"bead"}, []string{
		"type", "properties", "propertiesChange", "metadata", "metadataChange", "expectedRevision", "revisionProvenance",
	}}
	// revisionProvenance on deleteBead is Donna's 2026-11-06 ruling; write.md's table
	// and schema have not caught up.
	deleteBeadMembers = writeMembers{[]string{"bead"}, []string{"expectedRevision", "revisionProvenance"}}
)

// readWriteBody reads and checks the carrier: scanInvocation refuses what
// encoding/json lets through, at every depth, in one place for both surfaces.
func readWriteBody(w http.ResponseWriter, r *http.Request, closed writeMembers) (map[string]json.RawMessage, *callableProblem) {
	body, err := readLimitedBody(w, r, writeBodyBytes)
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			return nil, newCallableProblem(bdpRequestTooLarge, "the request is larger than "+strconv.Itoa(writeBodyBytes)+" bytes")
		}
		return nil, newCallableProblem(bdpMalformedRequest, "the request body could not be read")
	}
	members, _, detail := scanInvocation(body)
	if detail != "" {
		return nil, newCallableProblem(bdpMalformedRequest, strings.Replace(detail, "the invocation", "the request", 1))
	}
	if members == nil {
		return nil, newCallableProblem(bdpMalformedRequest, "the request is a JSON object")
	}
	for name := range members {
		if !slices.Contains(closed.required, name) && !slices.Contains(closed.optional, name) {
			return nil, newCallableProblem(bdpMalformedRequest, "`"+name+"` is not a member of this request")
		}
	}
	for _, name := range closed.required {
		if _, ok := members[name]; !ok {
			return nil, newCallableProblem(bdpMalformedRequest, "`"+name+"` is required")
		}
	}
	return members, nil
}

// respond answers 200 with the complete result: no ETag and no Location, because the
// target is a command URL and not the Resource.
func (c *writeCall) respond(result writeResult) {
	c.w.Header().Set("Cache-Control", "private, no-store")
	writeCoreJSON(c.w, result)
}

// actor reads revisionProvenance: the id of its actor, or writeNoActor. It shares the
// closed-input check with the Callables, so an agent or message, which native storage
// has nowhere to put, is refused the same way rather than dropped.
func (c *writeCall) actor() (string, *callableProblem) {
	raw, sent := c.members["revisionProvenance"]
	if !sent {
		return writeNoActor, nil
	}
	var provenance map[string]json.RawMessage
	if err := json.Unmarshal(raw, &provenance); err != nil || provenance == nil {
		return "", newCallableProblem(bdpMalformedRequest, "`revisionProvenance` is an object")
	}
	id, problem := checkProvenance(&callableRow{actorMember: "actor"}, provenance)
	if problem != nil {
		// A Callable's pointer is into its arguments, and Write's are into a document.
		for i := range problem.Diagnostics {
			problem.Diagnostics[i].InstanceLocation = ""
		}
		return "", problem
	}
	if id == nil {
		return writeNoActor, nil
	}
	var name string
	if err := json.Unmarshal(id, &name); err != nil {
		return "", newCallableProblem(bdpMalformedRequest, "`revisionProvenance.actor.id` is a string")
	}
	trimmed, res := validateActor(name)
	if res != nil {
		return "", writeInvalid(writeRequestDiagnostic("revisionProvenance.actor.id", "this is not a name this store can record"))
	}
	return trimmed, nil
}

// expectedRevision reads the guard, or "" for none.
func (c *writeCall) expectedRevision() (string, *callableProblem) {
	raw, sent := c.members["expectedRevision"]
	if !sent {
		return "", nil
	}
	var token string
	if err := json.Unmarshal(raw, &token); err != nil || token == "" || isJSONNull(raw) {
		return "", newCallableProblem(bdpMalformedRequest, "`expectedRevision` is a nonempty string")
	}
	return token, nil
}

// warmConfiguredTypes reads the type table before a transaction opens, so a
// projection inside the transaction never takes a second connection for it.
func (c *writeCall) warmConfiguredTypes() error {
	_, err := c.p.configuredTypes(c.ctx)
	return err
}

func writeMismatch(member string) error {
	return &writeRefusal{newCallableProblem(bdpRevisionMismatch, "`"+member+"` is a revision this Resource does not carry")}
}

// serveCreateBead is createBead. Whatever the role refuses leaves nothing behind; the
// id, the prefix, the retired ids and the create-only guard are the role's.
func serveCreateBead(c *writeCall) error {
	nativeType, problem := c.beadType(c.members["type"])
	if problem != nil {
		return &writeRefusal{problem}
	}
	local, problem := c.createdBeadID()
	if problem != nil {
		return &writeRefusal{problem}
	}
	properties, problem := writeObjectMember(c.members, "properties")
	if problem != nil {
		return &writeRefusal{problem}
	}
	metadata, problem := writeObjectMember(c.members, "metadata")
	if problem != nil {
		return &writeRefusal{problem}
	}
	if problem := checkAdmissibleNumbers(properties, metadata); problem != nil {
		return &writeRefusal{problem}
	}
	actor, problem := c.actor()
	if problem != nil {
		return &writeRefusal{problem}
	}
	issue, problem := issueFromProperties(properties)
	if problem != nil {
		return &writeRefusal{problem}
	}
	issue.ID = local
	issue.IssueType = types.IssueType(nativeType)
	if raw, sent := c.members["metadata"]; sent {
		issue.Metadata = raw
	}
	lifecycle, err := c.s.lifecycle(c.r)
	if err != nil {
		return err
	}
	result, err := lifecycle.Create(c.ctx, issueops.CreateRequest{Actor: actor, Issue: issue, StoredTimestamps: true})
	if err != nil {
		return err
	}
	record, err := c.p.bead(c.ctx, result.Issue)
	if err != nil {
		return err
	}
	c.respond(writeResult{Outcome: writeOutcomeCreated, Changed: true, Resource: record})
	return nil
}

// serveUpdateBead is updateBead. What it will write is decided inside the transaction,
// from the row that transaction reads, and Write's order is kept there: the complete
// result is validated, then the guard is compared, then an equal state is a no-op.
func serveUpdateBead(c *writeCall) error {
	subject, problem := c.subject("bead")
	if problem != nil {
		return &writeRefusal{problem}
	}
	edit, problem := c.beadEdit()
	if problem != nil {
		return &writeRefusal{problem}
	}
	expected, problem := c.expectedRevision()
	if problem != nil {
		return &writeRefusal{problem}
	}
	actor, problem := c.actor()
	if problem != nil {
		return &writeRefusal{problem}
	}
	local, ok := subject.bead()
	if !ok {
		return &writeRefusal{newCallableProblem(bdpResourceNotFound, "")}
	}
	if err := c.warmConfiguredTypes(); err != nil {
		return err
	}
	lifecycle, err := c.s.lifecycle(c.r)
	if err != nil {
		return err
	}
	result, err := lifecycle.Update(c.ctx, issueops.UpdateRequest{
		Actor:      actor,
		IssueID:    local,
		Provenance: "bd serve: BDP updateBead",
		InTx: func(current *types.Issue, attempt *issueops.UpdateRequest) error {
			patch, problem := edit.patchFor(current)
			if problem != nil {
				return &writeRefusal{problem}
			}
			record, err := c.p.bead(c.ctx, current)
			if err != nil {
				return err
			}
			if expected != "" && record.Revision != expected {
				return writeMismatch("expectedRevision")
			}
			attempt.Patch = patch
			// Write's patch form is an explicit replacement, which is what the
			// native notes fence asks a caller to say. The assignee fence and close
			// policy stay: Write's text gives no reason to step around them.
			attempt.ForceNotesOverwrite = patch.Notes.Set
			return nil
		},
	})
	if err != nil {
		return err
	}
	record, err := c.p.bead(c.ctx, result.Issue)
	if err != nil {
		return err
	}
	c.respond(writeResult{Outcome: writeOutcomeUpdated, Changed: result.Changed, Resource: record})
	return nil
}

// serveDeleteBead is deleteBead: the guard and the incident-Link refusal are in the
// deleting transaction. Unextended deletion mints no edition, so the identity reported
// is the last live one. The refusal names no Link: Write says hidden Links are not
// disclosed by it, and one rule for every caller keeps it from saying which are hidden.
func serveDeleteBead(c *writeCall) error {
	subject, problem := c.subject("bead")
	if problem != nil {
		return &writeRefusal{problem}
	}
	expected, problem := c.expectedRevision()
	if problem != nil {
		return &writeRefusal{problem}
	}
	actor, problem := c.actor()
	if problem != nil {
		return &writeRefusal{problem}
	}
	local, ok := subject.bead()
	if !ok {
		return &writeRefusal{newCallableProblem(bdpResourceNotFound, "")}
	}
	if err := c.warmConfiguredTypes(); err != nil {
		return err
	}
	deleter, err := c.s.deleter(c.r)
	if err != nil {
		return err
	}
	var name writeDeletedName
	_, err = deleter.Delete(c.ctx, issueops.DeleteRequest{
		Actor:               actor,
		IDs:                 []string{local},
		RefuseIncidentEdges: true,
		InTx: func(current *types.Issue) error {
			record, err := c.p.bead(c.ctx, current)
			if err != nil {
				return err
			}
			if expected != "" && record.Revision != expected {
				return writeMismatch("expectedRevision")
			}
			name = writeDeletedName{ID: record.ID, Type: record.Type, Revision: record.Revision}
			return nil
		},
	})
	if err != nil {
		return err
	}
	c.respond(writeResult{
		Outcome: writeOutcomeDeleted, Changed: true,
		Deleted: &writeDeleted{ResourceKind: coreKindBead, Resource: name},
	})
	return nil
}

// failWrite answers a failure from the role or the projection, in Write's table.
//
// A native refusal whose words name tables or configuration is not quoted, as v0's
// handlers do not quote it: the Problem says what kind of thing was refused, and the
// log carries the rest.
func (s *Server) failWrite(w http.ResponseWriter, r *http.Request, err error) {
	var (
		refusal      *writeRefusal
		unservable   errCoreUnservable
		incident     *issueops.IncidentEdgesError
		expired      *issueops.IdempotencyExpiredError
		notFound     *issueops.NotFoundError
		openChildren *issueops.CloseOpenChildrenError
	)
	refused := func() {
		s.event("request_refused", "request_id", requestInfo(r.Context()).id, "error", err.Error())
	}
	invalid := func(message string) {
		refused()
		writeWriteProblem(w, r, writeInvalid(callableDiagnostic{Message: message}))
	}
	switch {
	case errors.As(err, &refusal):
		writeWriteProblem(w, r, refusal.problem)
	case errors.As(err, &unservable):
		writeCoreProblem(w, r, coreCodeUnavailable, unservable.reason)
	case errors.Is(err, issueops.ErrVersionMismatch):
		writeWriteProblem(w, r, newCallableProblem(bdpRevisionMismatch, "`expectedRevision` is a revision this Resource does not carry"))
	case errors.As(err, &incident):
		writeWriteProblem(w, r, newCallableProblem(bdpIncidentLinksExist, "this Bead has live incident Links; remove them first"))
	case errors.As(err, &notFound), errors.Is(err, issueops.ErrNotFound), errors.Is(err, storage.ErrNotFound):
		writeWriteProblem(w, r, newCallableProblem(bdpResourceNotFound, ""))
	case errors.As(err, &expired), errors.Is(err, issueops.ErrIdempotencyExpired):
		writeWriteProblem(w, r, newCallableProblem(bdpIdentityTaken, "this identity was retired and cannot be created again"))
	case errors.Is(err, issueops.ErrAlreadyExists):
		writeWriteProblem(w, r, newCallableProblem(bdpIdentityTaken, "a live Resource already has this identity"))
	case errors.Is(err, uow.ErrDeleteFenceUnavailable):
		writeCoreProblem(w, r, coreCodeUnavailable, "another write to this Resource is still in progress")
	case errors.As(err, &openChildren), errors.Is(err, issueops.ErrCloseOpenChildren):
		invalid("this status closes an issue with open children; close them first, through `issue/close`")
	case errors.Is(err, issueops.ErrCloseBlocked):
		invalid("this status closes a blocked issue; clear the blocker first")
	case errors.Is(err, storage.ErrAlreadyClaimed):
		invalid("`assignee` transfers work away from a live foreign owner; `issue/update` carries the override")
	case errors.Is(err, types.ErrFieldTooLong):
		invalid("a value is longer than storage holds; nothing was written")
	case isTemplateRefusal(err):
		invalid("templates are read-only; nothing was written")
	case errors.Is(err, issueops.ErrValidation), errors.Is(err, storage.ErrValidation):
		invalid("a value was refused by this workspace's own validation; nothing was written")
	default:
		s.failCoreErr(w, r, err)
	}
}

func isTemplateRefusal(err error) bool {
	var template *issueops.TemplateReadOnlyError
	return errors.As(err, &template)
}
