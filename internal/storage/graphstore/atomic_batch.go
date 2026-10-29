package graphstore

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"unicode/utf8"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/canonicalbatch"
	"github.com/steveyegge/beads/internal/canonicalbatchstate"
	"github.com/steveyegge/beads/internal/storage/issueops"
	publicops "github.com/steveyegge/beads/issueops"
)

const AtomicBatchResultLimit = 4 << 20

var ErrBatchKeyConflict = errors.New("batch key belongs to a different frozen request")

type BatchOperationResult struct {
	OperationIndex int      `json:"operationIndex"`
	Operation      string   `json:"operation"`
	IDs            []string `json:"ids"`
	OperationName  string   `json:"operationName,omitempty"`
	// ID preserves the first local candidate's retained scalar result. New
	// decisions always use IDs; recovery must not rewrite an old frozen result.
	ID string `json:"id,omitempty"`
}

func (r BatchOperationResult) MarshalJSON() ([]byte, error) {
	if r.ID != "" {
		if r.IDs != nil {
			return nil, fmt.Errorf("%w: mixed batch result mappings", ErrInvalidStore)
		}
		return json.Marshal(struct {
			OperationIndex int    `json:"operationIndex"`
			Operation      string `json:"operation"`
			ID             string `json:"id"`
			OperationName  string `json:"operationName,omitempty"`
		}{r.OperationIndex, r.Operation, r.ID, r.OperationName})
	}
	type current BatchOperationResult
	return json.Marshal(current(r))
}

func (r *BatchOperationResult) UnmarshalJSON(raw []byte) error {
	canonical, err := graph.CanonicalizeJSON(raw)
	if err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(canonical, &fields); err != nil {
		return err
	}
	for key := range fields {
		switch key {
		case "operationIndex", "operation", "operationName", "id", "ids":
		default:
			return fmt.Errorf("%w: unknown batch result member", ErrInvalidStore)
		}
	}
	_, scalar := fields["id"]
	_, array := fields["ids"]
	if scalar == array || fields["operationIndex"] == nil || fields["operation"] == nil || bytes.Equal(fields["operationIndex"], []byte("null")) {
		return fmt.Errorf("%w: incomplete or mixed batch result mappings", ErrInvalidStore)
	}
	if name, present := fields["operationName"]; present && (len(name) == 0 || name[0] != '"') {
		return fmt.Errorf("%w: invalid batch operation name", ErrInvalidStore)
	}
	type current BatchOperationResult
	var decoded current
	if err := json.Unmarshal(canonical, &decoded); err != nil {
		return err
	}
	_, named := fields["operationName"]
	if decoded.OperationIndex < 0 || decoded.Operation == "" || named && decoded.OperationName == "" || scalar && decoded.ID == "" || array && decoded.IDs == nil {
		return fmt.Errorf("%w: invalid batch result mappings", ErrInvalidStore)
	}
	*r = BatchOperationResult(decoded)
	return nil
}

type BatchDeletedIdentity struct {
	ID           string             `json:"id"`
	ResourceKind graph.ResourceKind `json:"resourceKind"`
	Type         string             `json:"type"`
	Revision     string             `json:"revision"`
}

type BatchOwnerRevision struct {
	ID       string `json:"id"`
	Revision string `json:"revision"`
}

type BatchResourceOutcome struct {
	ID             string                `json:"id"`
	ResourceKind   graph.ResourceKind    `json:"resourceKind"`
	Changed        bool                  `json:"changed"`
	Outcome        string                `json:"outcome"`
	Resource       json.RawMessage       `json:"resource,omitempty"`
	Deleted        *BatchDeletedIdentity `json:"deleted,omitempty"`
	OwnerRevisions []BatchOwnerRevision  `json:"ownerRevisions,omitempty"`
}

// BatchResult is the local layout-15 candidate result. Resource bodies use this
// store's existing checked local representation, not an advertised new BDP wire
// profile. No public revision or Resource exists for an allocation-only row.
type BatchResult struct {
	CapabilityVersion string                 `json:"capabilityVersion"`
	Changed           bool                   `json:"changed"`
	Transaction       string                 `json:"transaction,omitempty"`
	Results           []BatchOperationResult `json:"results"`
	Resources         []BatchResourceOutcome `json:"resources"`
}

type atomicBatchRun struct {
	store                    *Store
	tx                       *sql.Tx
	ctx                      context.Context
	actor                    string
	plan                     canonicalbatchstate.Plan
	members                  []canonicalbatch.Member
	nativeContents           map[string]issueops.CanonicalIssueContent
	nativeChanged            map[string]bool
	nativeFacts              []*issueops.CanonicalFinalFacts
	nativeDescriptors        []graph.TypeDescriptor
	nativeDependencyAuthored bool
	nativeDependencyPlan     *atomicNativeDependencyPlan
	nativeDependencyFinal    *issueops.CanonicalDependencyFinalization
	initial                  map[string]any
	paths                    map[string]string
	backings                 map[string]string
	final                    map[string]any
	ownerIDs                 map[string][]string
	ownerCandidates          map[string]bool
}

// ApplyCanonicalBatch commits one bounded supported decision and its complete
// recovery result together. One key namespace belongs to the checked local Scope
// and authority. Actor is frozen input, not authentication or a separate key
// namespace. Retained successful results are read before current execution
// support, Type/default resolution, graph acquisition, or revision guards.
func (s *Store) ApplyCanonicalBatch(ctx context.Context, batch *canonicalbatch.Batch, key, actor string) (BatchResult, error) {
	if s.options.Binding.SchemaVersion != AtomicBatchSchemaVersion {
		return BatchResult{}, fmt.Errorf("%w: canonical batch requires fresh local profile %d", ErrCapabilityUnavailable, AtomicBatchSchemaVersion)
	}
	if batch == nil || len(batch.NativeJSON()) == 0 {
		return BatchResult{}, fmt.Errorf("%w: missing compiled batch", graph.ErrValidation)
	}
	if len(key) == 0 || len(key) > 256 || !utf8.ValidString(key) || !utf8.ValidString(actor) || len(actor) > 4096 {
		return BatchResult{}, fmt.Errorf("%w: invalid batch key or actor", graph.ErrValidation)
	}
	request, err := batch.DecisionJSON()
	if err != nil {
		return BatchResult{}, err
	}
	contextJSON, err := canonicalJSON(struct{ Scope, Authority, Carrier, Version string }{s.ScopeURL(), s.options.Binding.AuthorityID, "local-canonical-batch", batch.Version()})
	if err != nil {
		return BatchResult{}, err
	}
	var result BatchResult
	// Validation is called explicitly only for a fresh decision. Using the
	// unconditional final-Type-validation wrapper would re-admit a retained key
	// against today's graph before returning yesterday's immutable result.
	err = s.withTx(ctx, true, func(tx *sql.Tx) error {
		if err := checkBinding(ctx, tx, s.options); err != nil {
			return err
		}
		retained, found, err := readAtomicBatchDecision(ctx, tx, key, contextJSON, request, actor, s.ScopeURL(), batch.Members())
		if err != nil {
			return err
		}
		if found {
			result, err = projectAtomicBatchLabels(retained, batch.Members())
			return err
		}
		members := batch.Members()
		catalog, err := s.selectedTypes()
		if err != nil {
			return err
		}
		for i, m := range members {
			if m.Type.Revision != "" || m.RevisionProvenance != nil {
				return &canonicalbatchstate.MemberError{OperationIndex: i, Err: fmt.Errorf("%w: exact Type pins and supplied provenance have no layout-15 publisher", ErrCapabilityUnavailable)}
			}
			if m.Type.URI != "" {
				if err := s.requireAssessmentAuthoring(m.Type.URI); err != nil {
					return err
				}
				if _, err := s.readTypeInTx(ctx, tx, m.Type.URI); err != nil {
					return err
				}
				// A later deletion/cancellation does not grant this generic
				// publisher authority to stage a native domain creation/retype.
				if (m.Operation == "createBead" && !catalog.memoryType(m.Type.URI)) || (m.Operation == "createLink" && !IsInformationalTypeURL(s.ScopeURL(), m.Type.URI) && !isConcreteNativeDependency(s.ScopeURL(), m.Type.URI)) {
					return &canonicalbatchstate.MemberError{OperationIndex: i, Err: fmt.Errorf("%w: authored Type requires its native domain publisher", ErrCapabilityUnavailable)}
				}
			}
		}
		snapshot, err := s.currentSnapshotInTx(ctx, tx)
		if err != nil {
			return err
		}
		typeKinds, typeFamilies, err := atomicBatchTypeAuthority(ctx, snapshot.Types)
		if err != nil {
			return err
		}
		run := atomicBatchRun{nativeDescriptors: snapshot.Types, nativeContents: map[string]issueops.CanonicalIssueContent{}, nativeChanged: map[string]bool{}, members: members, store: s, tx: tx, ctx: ctx, actor: actor, initial: map[string]any{}, paths: map[string]string{}, backings: map[string]string{}, final: map[string]any{}, ownerIDs: map[string][]string{}, ownerCandidates: map[string]bool{}}
		current := map[string]canonicalbatchstate.Resource{}
		for _, record := range snapshot.Records {
			r, err := atomicBatchStateRecord(record)
			if err != nil {
				return err
			}
			current[r.ID] = r
			run.initial[r.ID] = record
			path, err := s.publicResourcePathInTx(ctx, tx, r.ID, r.Kind)
			if err != nil {
				return err
			}
			run.paths[r.ID] = path
			var backing string
			if err := tx.QueryRowContext(ctx, `SELECT backing FROM graph_preview_catalog WHERE path=?`, path).Scan(&backing); err != nil {
				return err
			}
			run.backings[r.ID] = backing
		}
		reserved := []string{}
		for _, m := range members {
			if m.ID != "" {
				id, err := atomicBatchLocalID(s.ScopeURL(), m.ID)
				if err != nil {
					return err
				}
				exists, err := publicResourceIDExistsInTx(ctx, tx, id)
				if err != nil {
					return err
				}
				if exists {
					reserved = append(reserved, s.ScopeURL()+id)
				}
			}
		}
		chosen := map[string]bool{}
		allocate := func(_ int, _ graph.ResourceKind) (string, error) {
			for attempts := 0; attempts < 32; attempts++ {
				id, err := selectGeneratedPublicIDInTx(ctx, tx, "")
				if err != nil {
					return "", err
				}
				if !chosen[id] {
					chosen[id] = true
					return s.ScopeURL() + id, nil
				}
			}
			return "", fmt.Errorf("%w: batch allocation collision budget", ErrAlreadyExists)
		}
		var pinBytes uint64
		checkPin := func(ref canonicalbatch.Reference) error {
			path, err := s.publicResourcePathInTx(ctx, tx, ref.URI, graph.KindBead)
			if err != nil {
				return err
			}
			_, err = s.endpointPinInTx(ctx, tx, path, EndpointPin{Revision: ref.Revision}, &pinBytes)
			return err
		}
		plan, err := canonicalbatchstate.Stage(canonicalbatchstate.Input{AdmitMutation: run.admitNativeMutation, AdmitResolvedMutation: run.admitResolvedNativeMutation, Context: ctx, TypeKinds: typeKinds, TypeFamilies: typeFamilies, Scope: s.ScopeURL(), Current: current, ReservedIDs: reserved, Members: members, Allocate: allocate, CheckPin: checkPin})
		if err != nil {
			return err
		}
		// A committed graph must remain acquirable by the next checked batch,
		// including a deletion intended to reduce its size. Byte admission alone
		// does not enforce currentSnapshotInTx's complete-inventory bound.
		if len(plan.Final) > PreviewSnapshotLimit {
			return fmt.Errorf("%w: final batch graph exceeds %d live Resources", ErrLimitExceeded, PreviewSnapshotLimit)
		}
		run.plan = plan
		if err := run.admit(); err != nil {
			return err
		}
		if err := run.prepareNativeDependencies(); err != nil {
			return err
		}
		if err := s.touchCoordination(ctx, tx); err != nil {
			return err
		}
		if err := run.publish(); err != nil {
			return err
		}
		if err := s.validateTypeStateInTx(ctx, tx); err != nil {
			return err
		}
		if err := checkCurrentReadBytes(ctx, tx); err != nil {
			return err
		}
		result, err = run.result()
		if err != nil {
			return err
		}
		encoded, err := canonicalJSON(result)
		if err != nil {
			return err
		}
		if len(encoded) > AtomicBatchResultLimit {
			return fmt.Errorf("%w: complete batch result exceeds local limit", graph.ErrValidation)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO graph_preview_batch_decisions(decision_key,context,request,actor,result) VALUES(?,?,?,?,?)`, key, contextJSON, request, actor, encoded); err != nil {
			return err
		}
		return s.afterStage("batch-decision")
	})
	if err != nil {
		return BatchResult{}, err
	}
	return result, nil
}

// Capture names are output presentation. The retained identity compares creator
// slots, so an equivalent retry may rename (or add/remove) unused names without
// changing any published identity, body, revision or transaction.
func projectAtomicBatchLabels(result BatchResult, members []canonicalbatch.Member) (BatchResult, error) {
	if len(result.Results) != len(members) {
		return BatchResult{}, fmt.Errorf("%w: retained member count", ErrInvalidStore)
	}
	setMembers := atomicBatchSetMembers(members)
	for i := range result.Results {
		row := &result.Results[i]
		if row.OperationIndex != i || members[i].Operation != row.Operation {
			return BatchResult{}, fmt.Errorf("%w: invalid retained operation mapping", ErrInvalidStore)
		}
		if row.ID != "" {
			if row.IDs != nil || members[i].Operation == "select" || setMembers[i] || graph.ValidateReferenceURI(row.ID) != nil {
				return BatchResult{}, fmt.Errorf("%w: invalid retained scalar mapping", ErrInvalidStore)
			}
			row.OperationName = members[i].Name
			continue
		}
		if row.IDs == nil {
			return BatchResult{}, fmt.Errorf("%w: missing retained identity array", ErrInvalidStore)
		}
		for j, id := range row.IDs {
			if graph.ValidateReferenceURI(id) != nil || j > 0 && graph.CompareCodeUnits(row.IDs[j-1], id) >= 0 {
				return BatchResult{}, fmt.Errorf("%w: retained identity array", ErrInvalidStore)
			}
		}
		if members[i].Operation != "select" && !setMembers[i] && len(row.IDs) != 1 {
			return BatchResult{}, fmt.Errorf("%w: retained singleton mapping", ErrInvalidStore)
		}
		row.OperationName = members[i].Name
	}
	raw, err := canonicalJSON(result)
	if err != nil {
		return BatchResult{}, err
	}
	if len(raw) > AtomicBatchResultLimit {
		return BatchResult{}, fmt.Errorf("%w: renamed batch result exceeds original local result budget", ErrLimitExceeded)
	}
	return result, nil
}

func readAtomicBatchDecision(ctx context.Context, tx *sql.Tx, key string, contextJSON, request []byte, actor, scope string, members []canonicalbatch.Member) (BatchResult, bool, error) {
	var storedContext, storedRequest, storedActor, result []byte
	var size uint64
	err := tx.QueryRowContext(ctx, `SELECT OCTET_LENGTH(context)+OCTET_LENGTH(request)+OCTET_LENGTH(actor)+OCTET_LENGTH(result) FROM graph_preview_batch_decisions WHERE decision_key=?`, key).Scan(&size)
	if errors.Is(err, sql.ErrNoRows) {
		return BatchResult{}, false, nil
	}
	if err != nil {
		return BatchResult{}, false, err
	}
	if size > uint64(2*AtomicBatchResultLimit+8192) {
		return BatchResult{}, false, fmt.Errorf("%w: oversized retained batch decision", ErrInvalidStore)
	}
	if err := tx.QueryRowContext(ctx, `SELECT context,request,actor,result FROM graph_preview_batch_decisions WHERE decision_key=?`, key).Scan(&storedContext, &storedRequest, &storedActor, &result); err != nil {
		return BatchResult{}, false, err
	}
	if !bytes.Equal(storedContext, contextJSON) || !bytes.Equal(storedRequest, request) || string(storedActor) != actor {
		return BatchResult{}, false, ErrBatchKeyConflict
	}
	decoded, err := validateAtomicBatchRetained(result, scope, members)
	if err != nil {
		return BatchResult{}, false, err
	}
	return decoded, true, nil
}

func atomicBatchLocalID(scope, id string) (string, error) {
	if validatePublicID(id) == nil {
		return id, nil
	}
	return graph.ScopeLocalPublicResourceID(scope, id)
}

func atomicBatchStateRecord(value any) (canonicalbatchstate.Resource, error) {
	var r canonicalbatchstate.Resource
	var properties any
	switch v := value.(type) {
	case Record:
		r = canonicalbatchstate.Resource{ID: v.ID, Kind: graph.KindBead, Revision: v.Revision, Type: canonicalbatch.Reference{URI: v.Type}, Metadata: v.Metadata}
		properties = v.Properties
	case IssueRecord:
		r = canonicalbatchstate.Resource{ID: v.ID, Kind: graph.KindBead, Revision: v.Revision, Type: canonicalbatch.Reference{URI: v.Type}, Metadata: v.Metadata}
		properties = v.Properties
	case LinkRecord:
		r = canonicalbatchstate.Resource{ID: v.ID, Kind: graph.KindLink, Revision: v.Revision, Type: canonicalbatch.Reference{URI: v.Type}, Metadata: v.Metadata, Source: canonicalbatch.Reference{URI: v.Source, Revision: v.SourceRevision}, Target: canonicalbatch.Reference{URI: v.Target, Revision: v.TargetRevision}}
		properties = v.Properties
	default:
		return r, fmt.Errorf("%w: unsupported current Resource record", ErrInvalidStore)
	}
	var err error
	r.Properties, err = canonicalJSON(properties)
	if err != nil {
		return r, err
	}
	r.Envelope, err = canonicalJSON(value)
	return r, err
}

func (run *atomicBatchRun) admit() error {
	catalog, err := run.store.selectedTypes()
	if err != nil {
		return err
	}
	for _, id := range run.plan.ConsumedIDs {
		local, err := atomicBatchLocalID(run.store.ScopeURL(), id)
		if err != nil {
			return err
		}
		if len(local) > 1024 {
			return fmt.Errorf("%w: batch allocated ID exceeds local 1024-byte limit", graph.ErrValidation)
		}
	}
	for _, o := range run.plan.Outcomes {
		if handled, err := run.admitNativeOutcome(o); handled || err != nil {
			if err != nil {
				return err
			}
			continue
		}
		if o.Before != nil && o.Changed {
			backing := run.backings[o.ID]
			if backing != "generic" && backing != "informational" {
				return fmt.Errorf("%w: changed native Resource %s has no atomic publisher", ErrCapabilityUnavailable, o.ID)
			}
		}
		if o.After != nil && o.Changed {
			r := *o.After
			if err := run.store.requireAssessmentAuthoring(r.Type.URI); err != nil {
				return err
			}
			if r.Kind == graph.KindBead {
				if !catalog.memoryType(r.Type.URI) {
					return fmt.Errorf("%w: atomic Bead requires its native domain writer", ErrCapabilityUnavailable)
				}
				var properties Properties
				if err := json.Unmarshal(r.Properties, &properties); err != nil {
					return err
				}
				if before, ok := run.initial[r.ID].(Record); ok && catalog.evaluationType(r.Type.URI) {
					if catalog.evaluationType(before.Type) {
						if err := preserveEvaluationSubjectGroup(before.Properties, properties); err != nil {
							return err
						}
					} else if err := run.store.preserveRetainedEvaluationGroup(run.ctx, run.tx, run.paths[r.ID], before.Revision, properties); err != nil {
						return err
					}
				}
			} else if !IsInformationalTypeURL(run.store.ScopeURL(), r.Type.URI) {
				return fmt.Errorf("%w: atomic Link requires its native domain writer", ErrCapabilityUnavailable)
			}
			if err := catalog.validateProperties(r.Type.URI, r.Properties, r.Kind); err != nil {
				return err
			}
			if _, err := commonMetadata(r.Metadata); err != nil {
				return err
			}
		}
		if o.Kind == graph.KindLink && o.Outcome != "allocated" {
			owners := map[string]bool{}
			for _, boundary := range []struct {
				link  *canonicalbatchstate.Resource
				graph map[string]canonicalbatchstate.Resource
			}{{o.Before, nil}, {o.After, run.plan.Final}} {
				if boundary.link == nil {
					continue
				}
				id := boundary.link.Source.URI
				var source any
				if boundary.graph == nil {
					source = run.initial[id]
				} else {
					r, ok := boundary.graph[id]
					if !ok {
						return ErrIncidentLinkConstraint
					}
					source = Record{Type: r.Type.URI}
				}
				owned, err := run.store.ownsLink(run.ctx, run.tx, source, boundary.link.Type.URI)
				if err != nil {
					return err
				}
				if owned {
					owners[id] = true
					run.ownerCandidates[id] = true
				}
			}
			for id := range owners {
				run.ownerIDs[o.ID] = append(run.ownerIDs[o.ID], id)
			}
			sort.Slice(run.ownerIDs[o.ID], func(i, j int) bool { return graph.CompareCodeUnits(run.ownerIDs[o.ID][i], run.ownerIDs[o.ID][j]) < 0 })
		}
		if o.Kind == graph.KindBead && o.After != nil && (o.Before == nil || run.backings[o.ID] == "generic") {
			run.ownerCandidates[o.ID] = true
		}
	}
	for _, r := range run.plan.Final {
		if r.Kind != graph.KindLink {
			continue
		}
		for _, endpoint := range []canonicalbatch.Reference{r.Source, r.Target} {
			bead, ok := run.plan.Final[endpoint.URI]
			if !ok || bead.Kind != graph.KindBead {
				return fmt.Errorf("%w: final Link endpoint %s", ErrIncidentLinkConstraint, endpoint.URI)
			}
		}
	}
	for id := range run.ownerCandidates {
		if value, ok := run.initial[id]; ok {
			switch owner := value.(type) {
			case Record:
			case IssueRecord:
				if owner.Properties == nil || owner.Properties.Ephemeral || owner.Properties.NoHistory {
					return fmt.Errorf("%w: native owner must be a durable Issue", ErrCapabilityUnavailable)
				}
				if err := issueops.AuthorizeTemplateUpdate(owner.Properties, publicops.UpdateRequest{}); err != nil {
					return err
				}
			default:
				return fmt.Errorf("%w: unknown owning Bead", ErrCapabilityUnavailable)
			}
		}
	}
	return nil
}
