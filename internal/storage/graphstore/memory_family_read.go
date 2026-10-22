package graphstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/memorycatalog"
	"github.com/steveyegge/beads/internal/memoryquery"
	"github.com/steveyegge/beads/internal/typecatalog"
)

// CurrentFamilyRequest selects current local roots with independent hop depths.
// The first slice admits References and Superseded By, including refinements,
// in either direction. It is not a historical selector or a public CLI schema.
// MaxBytes and Limits accumulate across all roots, including validation feeds.
// MaxInspected independently caps accumulated acquisition rows and accumulated
// pure candidate visits; those are two counters, not one combined work budget.
// Fixed workspace/Type control reads are separately checked and may repeat in
// existing helpers; their row/CPU cost is outside these persisted-data budgets.
// No zero limit means unlimited. Exact caps succeed; excess refuses completely.
type CurrentFamilyRequest struct {
	Roots     []memoryquery.CurrentRoot
	Family    typecatalog.Reference
	Direction graph.Direction
	Limits    memoryquery.Limits
	MaxBytes  uint64
}

// CurrentFamilyRead retains canonical Records separately from selected graph
// membership: embedded Owned copies must never be interpreted as selected edges.
// Graph uses local paths; canonical Records retain their original absolute IDs,
// properties, metadata, pins, ownership and attribution. WriterToken and current
// revisions describe this one observed view, not resolution of historical pins.
// Types/OwnedLinkTypes use the existing checked Snapshot context vocabulary.
type CurrentFamilyRead struct {
	WriterToken    string
	Graph          memoryquery.CurrentGraph
	Records        []any
	Types          []graph.TypeDescriptor
	OwnedLinkTypes map[string][]string
}

// ReadCurrentFamily reads all roots in one rollback-only transaction. It does
// not collect the global graph, open a transaction per root, follow historical
// pins, or invoke the legacy query evaluator. Selected old pins fail closed;
// equal-current pins bind to that same checked current version. Errors,
// cancellation and cleanup failures return zero state. Caller checks Store.Close
// before output. No CLI dispatch, output writer or public result is added here.
func (s *Store) ReadCurrentFamily(ctx context.Context, req CurrentFamilyRequest) (CurrentFamilyRead, error) {
	if ctx == nil {
		return CurrentFamilyRead{}, errors.New("graph operation requires a context")
	}
	if len(req.Roots) == 0 || len(req.Roots) > PreviewSnapshotLimit || !req.Direction.Valid() || req.MaxBytes == 0 || req.MaxBytes > PreviewCurrentReadByteLimit || req.Limits.MaxNodes < 1 || req.Limits.MaxNodes > PreviewSnapshotLimit || req.Limits.MaxLinks < 0 || req.Limits.MaxLinks > PreviewSnapshotLimit || req.Limits.MaxInspected < 0 || req.Limits.MaxInspected > 1000000 || req.Family.Revision != "" {
		return CurrentFamilyRead{}, fmt.Errorf("%w: invalid current family read", graph.ErrValidation)
	}
	for _, root := range req.Roots {
		if validatePath(root.Path) != nil || root.Depth < 0 || root.Depth > PreviewSnapshotLimit {
			return CurrentFamilyRead{}, fmt.Errorf("%w: invalid current family root", graph.ErrValidation)
		}
	}
	var result CurrentFamilyRead
	err := s.withTx(ctx, false, func(tx *sql.Tx) error {
		if err := checkBinding(ctx, tx, s.options); err != nil {
			return err
		}
		catalog, err := s.selectedTypes()
		if err != nil {
			return err
		}
		refs := catalog.projectionBinding(memorycatalog.References)
		successors := catalog.projectionBinding(memorycatalog.SupersededBy)
		if req.Family.URI != refs && req.Family.URI != successors {
			return fmt.Errorf("%w: current family slice admits References/succession", ErrCapabilityUnavailable)
		}
		r := &currentFamilyReader{store: s, tx: tx, catalog: catalog, request: req, allocations: map[string]currentFamilyAllocation{}, links: map[string]LinkRecord{}, feeds: map[string][]graph.Link{}}
		// Reuse the established SQL-side cumulative byte/row charger, not the exact
		// Requires resolver or its ownership-only Outgoing interface.
		r.budget = &memoryQueryReader{tx: tx, request: MemoryRequiresRequest{MaxBytes: req.MaxBytes, Limits: req.Limits}}
		for _, root := range req.Roots {
			n, err := r.Current(ctx, root.Path)
			if err != nil {
				return err
			}
			if req.Family.URI == successors && !catalog.memoryType(n.TypeURL) {
				return fmt.Errorf("%w: succession root must be Memory", graph.ErrValidation)
			}
		}
		result.Graph, err = memoryquery.CurrentFamily(ctx, r, r.matches, req.Roots, req.Direction, req.Limits)
		if err != nil {
			return err
		}
		result.Records = make([]any, 0, len(result.Graph.Nodes)+len(result.Graph.Links))
		for _, node := range result.Graph.Nodes {
			record, err := r.hydrate(ctx, node.Path)
			if err != nil {
				return err
			}
			result.Records = append(result.Records, record)
		}
		for _, link := range result.Graph.Links {
			record, ok := r.links[link.Path()]
			if !ok {
				return fmt.Errorf("%w: missing selected current Link", ErrInvalidStore)
			}
			result.Records = append(result.Records, record)
		}
		result.Types, err = installedPreviewTypes(ctx, tx, s.ScopeURL())
		if err != nil {
			return err
		}
		result.OwnedLinkTypes = map[string][]string{}
		for _, descriptor := range result.Types {
			if descriptor.Describes() == graph.KindBead {
				owned, err := s.SelectedOwnedLinkTypes(descriptor.ID())
				if err != nil {
					return err
				}
				result.OwnedLinkTypes[descriptor.ID()] = owned
			}
		}
		return tx.QueryRowContext(ctx, `SELECT writer_token FROM graph_preview_scope WHERE singleton=1`).Scan(&result.WriterToken)
	})
	if errors.Is(err, memoryquery.ErrLimitExceeded) {
		err = errors.Join(ErrLimitExceeded, err)
	}
	if errors.Is(err, memoryquery.ErrCurrentContextUnavailable) {
		err = errors.Join(ErrCapabilityUnavailable, err)
	}
	err = errors.Join(err, ctx.Err())
	if err != nil {
		return CurrentFamilyRead{}, err
	}
	return result, nil
}

type currentFamilyAllocation struct {
	node         memoryquery.CurrentNode
	backing, key string
}
type currentFamilyReader struct {
	store       *Store
	tx          *sql.Tx
	catalog     *selectedTypeCatalog
	request     CurrentFamilyRequest
	budget      *memoryQueryReader
	allocations map[string]currentFamilyAllocation
	links       map[string]LinkRecord
	feeds       map[string][]graph.Link
}

func (r *currentFamilyReader) matches(_ context.Context, id string) (bool, error) {
	d, ok := r.catalog.types.Definition(id)
	if !ok || d.Kind != graph.KindLink {
		return false, fmt.Errorf("%w: unknown current Link Type", ErrInvalidStore)
	}
	return r.catalog.types.Conforms(d.Identity, r.request.Family), nil
}

func (r *currentFamilyReader) Current(ctx context.Context, path string) (memoryquery.CurrentNode, error) {
	if err := ctx.Err(); err != nil {
		return memoryquery.CurrentNode{}, err
	}
	if a, ok := r.allocations[path]; ok {
		return a.node, nil
	}
	if err := validatePath(path); err != nil {
		return memoryquery.CurrentNode{}, err
	}
	count, err := r.chargeRows(ctx, "graph_preview_catalog c WHERE c.path=?", "c.path,c.resource_kind,c.type_url,c.revision,c.allocation_state,c.backing,c.backing_key", 4, path)
	if err != nil {
		return memoryquery.CurrentNode{}, err
	}
	if count != 1 {
		return memoryquery.CurrentNode{}, ErrNotFound
	}
	var kind, typ, version, state, backing string
	var key sql.NullString
	if err := r.tx.QueryRowContext(ctx, `SELECT resource_kind,type_url,revision,allocation_state,backing,backing_key FROM graph_preview_catalog WHERE path=?`, path).Scan(&kind, &typ, &version, &state, &backing, &key); err != nil {
		return memoryquery.CurrentNode{}, err
	}
	if state == "deleted" {
		return memoryquery.CurrentNode{}, ErrGone
	}
	if state != "live" || kind != "bead" || !authorityID.MatchString(version) {
		return memoryquery.CurrentNode{}, fmt.Errorf("%w: unsupported current allocation", ErrInvalidStore)
	}
	if !((backing == "generic" && !key.Valid && r.catalog.memoryType(typ)) || (backing == "issue" && key.Valid && key.String != "" && typ == IssueTypeURL(r.store.ScopeURL()))) {
		return memoryquery.CurrentNode{}, fmt.Errorf("%w: current read supports live Memory/native Issue", ErrCapabilityUnavailable)
	}
	n := memoryquery.CurrentNode{Path: path, TypeURL: typ, Version: version}
	r.allocations[path] = currentFamilyAllocation{node: n, backing: backing, key: key.String}
	return n, nil
}

func (r *currentFamilyReader) Incident(ctx context.Context, path string, allowance int) ([]graph.Link, error) {
	if v, ok := r.feeds[path]; ok {
		return v, nil
	}
	// Both directions are necessary even for an outward walk: the induced graph
	// includes reverse/cross Links between reached nodes. This does not expand
	// reachability or frontier in the opposite direction.
	if err := r.chargeInformational(ctx, "(l.source_path=? OR l.target_path=?)", path, path); err != nil {
		return nil, err
	}
	if err := r.chargeNative(ctx, "(src.path=? OR dst.path=?)", path, path); err != nil {
		return nil, err
	}
	records, err := r.store.incidentLinksInTx(ctx, r.tx, LinksRequest{BeadPath: path, Direction: "both"}, "")
	if err != nil {
		return nil, err
	}
	if len(records) > allowance {
		return nil, fmt.Errorf("%w: current incidence allowance", ErrLimitExceeded)
	}
	out := make([]graph.Link, 0, len(records))
	for _, record := range records {
		src, err := graph.ParseRef(r.store.ScopeURL(), record.Source, record.SourceRevision)
		if err != nil {
			return nil, err
		}
		dst, err := graph.ParseRef(r.store.ScopeURL(), record.Target, record.TargetRevision)
		if err != nil {
			return nil, err
		}
		revision, err := graph.NewRevision(record.Revision)
		if err != nil {
			return nil, err
		}
		raw, err := canonicalJSON(record.Properties)
		if err != nil {
			return nil, err
		}
		properties, err := graph.NewProperties(raw)
		if err != nil {
			return nil, err
		}
		linkPath := strings.TrimPrefix(record.ID, r.store.ScopeURL())
		l, err := graph.NewLink(graph.LinkSpec{Path: linkPath, TypeURL: record.Type, Revision: revision, Source: src, Target: dst, Properties: properties})
		if err != nil {
			return nil, err
		}
		r.links[linkPath] = record
		out = append(out, l)
	}
	r.feeds[path] = out
	return out, nil
}

func (r *currentFamilyReader) hydrate(ctx context.Context, path string) (any, error) {
	a, ok := r.allocations[path]
	if !ok || (a.backing != "generic" && a.backing != "issue") {
		return nil, fmt.Errorf("%w: unselected hydration target", ErrInvalidStore)
	}
	if err := r.chargeInformational(ctx, "l.source_path=?", path); err != nil {
		return nil, err
	}
	if a.backing == "generic" {
		if _, err := r.chargeRows(ctx, "graph_preview_payloads p WHERE p.path=?", "p.properties,p.metadata", 3, path); err != nil {
			return nil, err
		}
		if _, err := r.chargeRows(ctx, "graph_preview_versions v WHERE v.path=? AND v.version=?", "v.snapshot,v.actor", 3, path, a.node.Version); err != nil {
			return nil, err
		}
		record, err := r.store.showMemoryInTx(ctx, r.tx, path)
		if err != nil {
			return nil, err
		}
		if record.Version != a.node.Version {
			return nil, ErrInvalidStore
		}
		return record, nil
	}
	if err := r.chargeIssue(ctx, path, a.key, a.node.Version); err != nil {
		return nil, err
	}
	record, err := r.store.showIssueInTx(ctx, r.tx, path)
	if err != nil {
		return nil, err
	}
	if record.Version != a.node.Version {
		return nil, ErrInvalidStore
	}
	return record, nil
}

// Compile-time guarantee: this is current incidence, never Requires Outgoing.
var _ memoryquery.CurrentContext = (*currentFamilyReader)(nil)
