package graphstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	graph "github.com/steveyegge/beads/graphops"
)

// QueryMaxResources is the total allocation budget for the initial coherent
// Query reader, including deleted allocations inspected for integrity. It is
// independent of the older CurrentSnapshot limit and of any result page size.
// Exceeding it refuses the entire read before acquiring allocation payloads.
const QueryMaxResources = 10000

// QuerySnapshot captures authoritative current Bead/Link records, ownership
// facts and complete nominal Type ancestry in one checked transaction. The
// existing legacy Type descriptors are a separate category: they contribute
// Type facts but are not invented as revisioned Beads in Records. The HTTP/CLI
// adapter projects Snapshot with the existing Resource projector before Query
// evaluation. This internal structure is not a wire envelope or cursor.
type QuerySnapshot struct {
	Snapshot     Snapshot
	TypeFamilies map[string][]string
	TypeKinds    map[string]graph.ResourceKind
}

// ReadQuerySnapshot acquires a bounded, complete current read context. Callers
// must admit the Query request before invoking this method. It does not invoke
// legacy Issue query evaluation, resolve retained/external References, or open
// independently transacted point reads. Pins stay authored pins; evaluation
// must refuse an unavailable edition rather than substitute its current body.
// All acquisition, cancellation and rollback/connection failures return zero
// state. The caller still owns Store.Close before publishing output.
func (s *Store) ReadQuerySnapshot(ctx context.Context) (QuerySnapshot, error) {
	var result QuerySnapshot
	err := s.withTx(ctx, false, func(tx *sql.Tx) error {
		var err error
		result, err = s.readQuerySnapshotInTx(ctx, tx)
		return err
	})
	if ctx != nil {
		err = errors.Join(err, ctx.Err())
	}
	if err != nil {
		return QuerySnapshot{}, err
	}
	return result, nil
}

func (s *Store) readQuerySnapshotInTx(ctx context.Context, tx *sql.Tx) (QuerySnapshot, error) {
	if err := checkBinding(ctx, tx, s.options); err != nil {
		return QuerySnapshot{}, err
	}
	if err := checkQueryAllocationCount(ctx, tx, QueryMaxResources); err != nil {
		return QuerySnapshot{}, err
	}
	// Charge all persisted current acquisitions, including owner hydration,
	// native relations, Type/control data and tombstone validation. No payload
	// is decoded before the shared 16 MiB admission completes.
	if err := checkCurrentReadBytes(ctx, tx); err != nil {
		return QuerySnapshot{}, err
	}
	if err := s.checkCollectionMappingsInTx(ctx, tx); err != nil {
		return QuerySnapshot{}, err
	}
	var invalid int
	err := tx.QueryRowContext(ctx, `SELECT 1 FROM graph_preview_catalog WHERE allocation_state NOT IN ('live','deleted') LIMIT 1`).Scan(&invalid)
	if err == nil {
		return QuerySnapshot{}, fmt.Errorf("%w: unknown Query allocation state", ErrInvalidStore)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return QuerySnapshot{}, err
	}
	result := QuerySnapshot{Snapshot: Snapshot{
		Records: []any{}, Types: []graph.TypeDescriptor{},
		OwnedLinkTypes: map[string][]string{}, OwnedLinkTypesByResource: map[string][]string{},
	}}
	if err := tx.QueryRowContext(ctx, `SELECT writer_token FROM graph_preview_scope WHERE singleton=1`).Scan(&result.Snapshot.WriterToken); err != nil {
		return QuerySnapshot{}, err
	}
	result.Snapshot.Types, err = s.selectedDescriptorsInTx(ctx, tx)
	if err != nil {
		return QuerySnapshot{}, err
	}
	result.TypeFamilies, result.TypeKinds, err = queryTypeFamilies(ctx, result.Snapshot.Types)
	if err != nil {
		return QuerySnapshot{}, err
	}
	var linkTypes []string
	for _, descriptor := range result.Snapshot.Types {
		if descriptor.Describes() == graph.KindLink {
			linkTypes = append(linkTypes, descriptor.ID())
		}
	}
	for _, descriptor := range result.Snapshot.Types {
		if descriptor.Describes() != graph.KindBead {
			continue
		}
		owned, err := s.SelectedOwnedLinkTypesWithDescriptors(descriptor.ID(), result.Snapshot.Types, linkTypes...)
		if err != nil {
			return QuerySnapshot{}, err
		}
		result.Snapshot.OwnedLinkTypes[descriptor.ID()] = owned
	}
	allocations, err := queryAllocationsInTx(ctx, tx)
	if err != nil {
		return QuerySnapshot{}, err
	}
	for _, allocation := range allocations {
		if err := ctx.Err(); err != nil {
			return QuerySnapshot{}, err
		}
		var record any
		switch allocation.backing {
		case "generic":
			record, err = s.showMemoryInTx(ctx, tx, allocation.path)
		case "issue":
			record, err = s.showIssueInTx(ctx, tx, allocation.path)
		case "dependency", "informational":
			record, err = s.showLinkInTx(ctx, tx, allocation.path)
		default:
			return QuerySnapshot{}, fmt.Errorf("%w: unsupported Query allocation backing", ErrInvalidStore)
		}
		if err != nil {
			return QuerySnapshot{}, err
		}
		if err := s.captureQueryOwnership(record, &result.Snapshot); err != nil {
			return QuerySnapshot{}, err
		}
		result.Snapshot.Records = append(result.Snapshot.Records, record)
	}
	if err := s.afterStage("query-inventory"); err != nil {
		return QuerySnapshot{}, err
	}
	return result, nil
}

func checkQueryAllocationCount(ctx context.Context, tx *sql.Tx, limit uint64) error {
	var count uint64
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM graph_preview_catalog`).Scan(&count); err != nil {
		return err
	}
	if count > limit {
		return fmt.Errorf("%w: complete Query context exceeds %d allocations before payload acquisition", ErrLimitExceeded, limit)
	}
	return nil
}

type queryAllocation struct{ path, backing, kind string }

func queryAllocationsInTx(ctx context.Context, tx *sql.Tx) ([]queryAllocation, error) {
	rows, err := tx.QueryContext(ctx, `SELECT path,backing,resource_kind FROM graph_preview_catalog WHERE allocation_state='live' LIMIT ?`, QueryMaxResources+1)
	if err != nil {
		return nil, err
	}
	allocations := []queryAllocation{}
	for rows.Next() {
		var allocation queryAllocation
		if err := rows.Scan(&allocation.path, &allocation.backing, &allocation.kind); err != nil {
			return nil, errors.Join(err, rows.Close())
		}
		if err := validateResourcePath(allocation.path); err != nil {
			return nil, errors.Join(fmt.Errorf("%w: malformed Query allocation path: %v", ErrInvalidStore, err), rows.Close())
		}
		link := allocation.kind == "link"
		if (allocation.kind != "bead" && !link) || link != strings.HasPrefix(allocation.path, "links/") ||
			(link && allocation.backing != "dependency" && allocation.backing != "informational") ||
			(!link && allocation.backing != "generic" && allocation.backing != "issue") {
			return nil, errors.Join(fmt.Errorf("%w: Query allocation kind/backing/path mismatch", ErrInvalidStore), rows.Close())
		}
		allocations = append(allocations, allocation)
		if len(allocations) > QueryMaxResources {
			return nil, errors.Join(fmt.Errorf("%w: complete Query inventory exceeds %d live Resources", ErrLimitExceeded, QueryMaxResources), rows.Close())
		}
	}
	if err := errors.Join(rows.Err(), rows.Close()); err != nil {
		return nil, err
	}
	sort.Slice(allocations, func(i, j int) bool { return graph.CompareCodeUnits(allocations[i].path, allocations[j].path) < 0 })
	return allocations, nil
}

func (s *Store) captureQueryOwnership(record any, snapshot *Snapshot) error {
	var id, typ string
	var owned []json.RawMessage
	switch value := record.(type) {
	case Record:
		id, typ, owned = value.ID, value.Type, value.Owned
	case IssueRecord:
		id, typ, owned = value.ID, value.Type, value.Owned
	case LinkRecord:
		return nil
	default:
		return fmt.Errorf("%w: unknown Query record category", ErrInvalidStore)
	}
	var observed []string
	for _, raw := range owned {
		var link LinkRecord
		if err := json.Unmarshal(raw, &link); err != nil {
			return fmt.Errorf("%w: malformed captured Query owned Link", ErrInvalidStore)
		}
		observed = append(observed, link.Type)
	}
	ownedTypes, err := s.SelectedOwnedLinkTypesWithDescriptors(typ, snapshot.Types, observed...)
	if err != nil {
		return err
	}
	snapshot.OwnedLinkTypesByResource[id] = ownedTypes
	return nil
}

// queryTypeFamilies derives exact reflexive ancestry from the entire checked
// descriptor inventory. Empty/abstract families remain known even with no
// instances. Unknown parents, kind changes and cycles fail closed. No URI
// spelling, local label or current Resource population confers Type membership.
func queryTypeFamilies(ctx context.Context, descriptors []graph.TypeDescriptor) (map[string][]string, map[string]graph.ResourceKind, error) {
	byID := make(map[string]graph.TypeDescriptor, len(descriptors))
	kinds := make(map[string]graph.ResourceKind, len(descriptors))
	for _, descriptor := range descriptors {
		if descriptor.IsZero() || (descriptor.Describes() != graph.KindBead && descriptor.Describes() != graph.KindLink) {
			return nil, nil, fmt.Errorf("%w: invalid Query Type descriptor", ErrInvalidStore)
		}
		if _, duplicate := byID[descriptor.ID()]; duplicate {
			return nil, nil, fmt.Errorf("%w: duplicate Query Type descriptor", ErrInvalidStore)
		}
		byID[descriptor.ID()], kinds[descriptor.ID()] = descriptor, descriptor.Describes()
	}
	families := make(map[string][]string, len(descriptors))
	state := make(map[string]uint8, len(descriptors))
	var visit func(string) ([]string, error)
	visit = func(id string) ([]string, error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if state[id] == 1 {
			return nil, fmt.Errorf("%w: cyclic Query Type ancestry", ErrInvalidStore)
		}
		if state[id] == 2 {
			return families[id], nil
		}
		descriptor, exists := byID[id]
		if !exists {
			return nil, fmt.Errorf("%w: incomplete Query Type ancestry", ErrInvalidStore)
		}
		state[id] = 1
		closure := map[string]bool{id: true}
		for _, parent := range descriptor.ConformsTo() {
			if kinds[parent] != descriptor.Describes() {
				return nil, fmt.Errorf("%w: missing or wrong-kind Query Type parent", ErrInvalidStore)
			}
			inherited, err := visit(parent)
			if err != nil {
				return nil, err
			}
			for _, ancestor := range inherited {
				closure[ancestor] = true
			}
		}
		for ancestor := range closure {
			families[id] = append(families[id], ancestor)
		}
		sort.Slice(families[id], func(i, j int) bool { return graph.CompareCodeUnits(families[id][i], families[id][j]) < 0 })
		state[id] = 2
		return families[id], nil
	}
	for _, descriptor := range descriptors {
		if _, err := visit(descriptor.ID()); err != nil {
			return nil, nil, err
		}
	}
	return families, kinds, nil
}
