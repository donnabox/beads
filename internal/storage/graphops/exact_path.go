package graphops

import (
	"context"
	"database/sql"
	"errors"

	graph "github.com/steveyegge/beads/graphops"
)

// exactAbsence is private physical information, not a public Reader error or
// permission to disclose history. Empty state means never seen in the qualified
// projection; other values retain the three distinct non-live allocation states.
// It deliberately has no Is or Unwrap mapping to any public or private sentinel.
type exactAbsence struct{ state graph.AllocationState }

func (*exactAbsence) Error() string { return "exact graph resource is unavailable" }

type exactPathRow struct {
	requested         string
	path, kind, state sql.NullString
	present           bool
}

func (r *exactPathRow) targets() []any {
	return []any{&r.requested, &r.path, &r.kind, &r.state, &r.present}
}

// B4 allocations derive from allocate events; B3 requires provenance validation.
// An orphan is therefore a corrupt projection, not evidence of a new state.
// This check does not certify the ledger, birth metadata or anti-reuse history.
func (r exactPathRow) classify(path string, kind graph.ResourceKind, resourcePath string, b *readBudget) (*exactAbsence, error) {
	if err := b.charge([]byte(r.requested), []byte(r.path.String), []byte(r.kind.String), []byte(r.state.String)); err != nil {
		return nil, err
	}
	if r.requested != path || r.present != (resourcePath != "") || (r.present && resourcePath != path) {
		return nil, corrupt(errors.New("exact resource anchor mismatch"))
	}
	if r.path.Valid != r.kind.Valid || r.path.Valid != r.state.Valid {
		return nil, corrupt(errors.New("partial allocation projection"))
	}
	if !r.path.Valid {
		if r.present {
			return nil, corrupt(errors.New("resource lacks allocation"))
		}
		return &exactAbsence{}, nil
	}
	if r.path.String != path || r.kind.String != string(kind) {
		return nil, corrupt(errors.New("allocation identity mismatch"))
	}
	switch r.state.String {
	case graph.AllocationLive:
		if !r.present {
			return nil, corrupt(errors.New("allocation/resource presence disagreement"))
		}
		return nil, nil
	case graph.AllocationReserved, graph.AllocationPruned, graph.AllocationErased:
		if r.present {
			return nil, corrupt(errors.New("allocation/resource presence disagreement"))
		}
		return &exactAbsence{state: r.state.String}, nil
	default:
		return nil, corrupt(errors.New("invalid allocation state"))
	}
}

type exactResult[T any] struct {
	value   T
	absence *exactAbsence
}

// Consume the complete singleton result before returning absence: duplicate,
// iteration and close errors must override any apparent missing/gone outcome.
func readExact[T any](ctx context.Context, tx queryer, query string, args []any, scan func(*sql.Rows) (exactResult[T], error)) (zero T, err error) {
	rows, err := readRows(ctx, tx, query, args, 1, scan)
	if errors.Is(err, errRowOverflow) {
		err = errors.Join(err, corrupt(errors.New("duplicate singleton row")))
	}
	if err != nil {
		return zero, err
	}
	// errCorrupt denotes an inconsistent private projection, including an
	// impossible query shape; this does not diagnose persisted storage damage.
	if len(rows) != 1 {
		return zero, corrupt(errors.New("missing exact lookup anchor"))
	}
	if rows[0].absence != nil {
		return zero, rows[0].absence
	}
	return rows[0].value, nil
}

const allocationColumns = "req.requested_path, a.path, CAST(a.resource_kind AS CHAR), CAST(a.state AS CHAR), "
const joinedBeadColumns = "COALESCE(b.path, ''), COALESCE(b.type_url, ''), COALESCE(b.revision, ''), b.attribution_principal, CAST(b.attribution_status AS CHAR), CASE WHEN LENGTH(b.properties) < 0 OR LENGTH(b.properties) > ? THEN NULL ELSE b.properties END, LENGTH(b.properties)"

// Independently keyed subqueries preserve the anchor while exposing PK filters.
// LIMIT 2 retains duplicate detection; index use is qualified by actual plans.
const exactBeadQuery = "SELECT " + allocationColumns + "b.path IS NOT NULL, " + joinedBeadColumns + " FROM (SELECT ? AS requested_path) req LEFT JOIN (SELECT path, resource_kind, state FROM graph_allocations WHERE path = ? LIMIT 2) a ON TRUE LEFT JOIN (SELECT path, type_url, revision, attribution_principal, attribution_status, properties FROM graph_beads WHERE path = ? LIMIT 2) b ON TRUE LIMIT 2"

// Keep the descriptor join inside the path-keyed Link subquery so its hint
// addresses two physical relations. The pinned optimizer may ignore a hint;
// only actual point-lookup plans qualify it. Raw values may materialize inside
// the engine; the outer CASE bounds transfer and decoding, not engine memory.
const exactLinkQuery = "SELECT " + allocationColumns + "l.path IS NOT NULL, " + joinedLinkColumns + ", l.descriptor_url, CASE WHEN LENGTH(l.descriptor_payload) < 0 OR LENGTH(l.descriptor_payload) > ? THEN NULL ELSE l.descriptor_payload END, LENGTH(l.descriptor_payload), l.descriptor_fingerprint FROM (SELECT ? AS requested_path) req LEFT JOIN (SELECT path, resource_kind, state FROM graph_allocations WHERE path = ? LIMIT 2) a ON TRUE LEFT JOIN (SELECT /*+ LEFT_OUTER_LOOKUP_JOIN(src,d) */ src.path, src.type_url, src.revision, src.attribution_principal, src.attribution_status, src.properties, src.source_kind, src.source_path, src.source_url, src.source_pin, src.target_kind, src.target_path, src.target_url, src.target_pin, d.url AS descriptor_url, d.descriptor AS descriptor_payload, d.fingerprint AS descriptor_fingerprint FROM graph_links src LEFT JOIN graph_type_descriptors d ON d.url = src.type_url WHERE src.path = ? LIMIT 2) l ON TRUE LIMIT 2"
