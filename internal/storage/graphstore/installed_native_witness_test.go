//go:build cgo

package graphstore

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/resourceget"
)

// The core has no production API or arbitrary-store discovery. The private
// test request branch below must supply normally resolved exact Options and
// controller-bound exclusive fixture custody before calling the opener.
// It never initializes, migrates, repairs, reseeds or changes engine settings.
type installedNativeWitness struct {
	Origin                 string                           `json:"origin"`
	Scope                  string                           `json:"scope"`
	Documents              []installedWitnessDocument       `json:"documents"`
	Types                  []installedWitnessType           `json:"types"`
	InstalledDefinitionIDs []string                         `json:"installedDefinitionIDs"`
	QueryEndpointMappings  []installedWitnessEndpoints      `json:"queryEndpointMappings"`
	NativeState            map[string]installedWitnessTable `json:"nativeState"`
	RetainedPredecessors   []installedWitnessDocument       `json:"retainedPredecessors"`
	Bare333Lookup          string                           `json:"bare333Lookup"`
	SourceClosure          json.RawMessage                  `json:"sourceClosure"`
}

type installedWitnessDocument struct {
	Kind graph.ResourceKind `json:"kind"`
	Body json.RawMessage    `json:"body"`
}

type installedWitnessType struct {
	Reference  resourceget.TypeReference   `json:"reference"`
	Definition resourceget.TypeReference   `json:"definition"`
	Describes  graph.ResourceKind          `json:"describes"`
	Current    bool                        `json:"current"`
	Abstract   bool                        `json:"abstract"`
	Ancestors  []resourceget.TypeReference `json:"ancestors"`
}

type installedWitnessReference struct {
	ID       string `json:"id"`
	Revision string `json:"revision"`
}

type installedWitnessEndpoints struct {
	Link   installedWitnessReference `json:"link"`
	Source installedWitnessReference `json:"source"`
	Target installedWitnessReference `json:"target"`
}

type installedWitnessTable struct {
	Columns []string                  `json:"columns"`
	Rows    [][]*installedWitnessCell `json:"rows"`
}

// installedWitnessCell retains one lossless SQL value representation. Strings
// and binary cells keep their exact underlying bytes, not a second retained
// JSON-escaped/base64 expansion. Other values keep their original standard JSON.
// The digest remains a bucket index ONLY; equality always compares complete JSON.
// The existing256-byte entry reservation covers this private cell/header/index;
// every row/reference keeps its existing debit. No budget/cap/credit is added.
type installedWitnessCell struct {
	kind   byte
	raw    json.RawMessage
	text   string
	binary []byte
}

func (cell *installedWitnessCell) MarshalJSON() ([]byte, error) {
	switch cell.kind {
	case 1:
		return json.Marshal(cell.text)
	case 2:
		return json.Marshal(cell.binary)
	default:
		return cell.raw, nil
	}
}

// The only deferred encodings are string and []byte, for which encoding/json
// cannot fail. All other cells were successfully encoded before retention.
// Temporary encoding/comparator scratch is NOT an engine peak-heap guarantee.
func installedWitnessCellJSON(cell *installedWitnessCell) []byte {
	raw, err := cell.MarshalJSON()
	if err != nil {
		panic(err)
	}
	return raw
}

type installedWitnessAllocation struct {
	Path     string
	PublicID string
	Kind     graph.ResourceKind
	Revision string
	Backing  string
	State    string
}

// openInstalledNativeWitness is an explicit package-test entry, not a selected
// executable. All CLI and other native handles MUST already be closed. The
// caller owns the original complete predecessor bodies/applied result evidence;
// these requested exact references only verify retained persistence, not that a
// legacy batch response carried a same-transaction complete Resource facet.
func openInstalledNativeWitness(ctx context.Context, options Options, predecessors []installedWitnessReference, require333Absent bool) (installedNativeWitness, error) {
	s, err := OpenExisting(ctx, options)
	if err != nil {
		return installedNativeWitness{}, err
	}
	result, readErr := readInstalledNativeWitness(ctx, s, predecessors, require333Absent)
	err = errors.Join(readErr, s.Close())
	if err != nil {
		return installedNativeWitness{}, err
	}
	return result, nil
}

func readInstalledNativeWitness(ctx context.Context, s *Store, predecessors []installedWitnessReference, require333Absent bool) (installedNativeWitness, error) {
	if ctx == nil || s.LayoutVersion() != TypesFreshSchemaVersion {
		return installedNativeWitness{}, ErrCapabilityUnavailable
	}
	var result installedNativeWitness
	stage := "transaction-admission"
	remaining, metadata := -1, -1
	err := s.withDefinitionReadTx(ctx, txDefinitionRead, func(tx *sql.Tx, budget *intrinsicMetadataBudget) error {
		defer func() { remaining, metadata = budget.remaining, budget.metadata }()
		stage = "independent-allocation-census"
		// Independent allocation population: no List, CurrentResourceDocuments,
		// Query, common live-identity population helper or four-name inventory.
		allocations, err := installedWitnessAllocations(ctx, tx, budget)
		if err != nil {
			return err
		}
		stage = "checked-type-capture"
		capture, err := s.captureTypeRecordsInTx(ctx, tx, budget)
		if err != nil {
			return err
		}
		readctx := s.contextForResourceCapture(ctx, tx, capture, budget)
		stage = "checked-collection-mappings"
		if err = s.checkCollectionMappingsInTx(readctx, tx); err != nil {
			return err
		}
		stage = "common-native-row-reservation"
		if err = chargeCommonNativeRowsInTx(readctx, tx, budget); err != nil {
			return err
		}
		stage = "witness-selected-installed-sources"
		sources, err := selectedInstalledTypeSourcesInTx(readctx, tx, s.ScopeURL())
		if err != nil {
			return err
		}
		sort.Slice(sources, func(i, j int) bool { return sources[i].Descriptor.ID() < sources[j].Descriptor.ID() })
		stage = "witness-source-serialization"
		sourceBytes, err := json.Marshal(sources)
		if err != nil {
			return err
		}
		stage = "witness-source-copy-debit"
		if err = budget.chargeBody(len(sourceBytes)); err != nil {
			return err
		}
		result = installedNativeWitness{Origin: "native-installed-and-domain-witness", Scope: s.ScopeURL(), Documents: []installedWitnessDocument{}, Types: []installedWitnessType{}, InstalledDefinitionIDs: []string{}, QueryEndpointMappings: []installedWitnessEndpoints{}, RetainedPredecessors: []installedWitnessDocument{}, Bare333Lookup: "not-requested", SourceClosure: sourceBytes}
		documents := map[string]resourceget.ResourceDocument{}
		for _, document := range capture.snapshot.Documents {
			if _, duplicate := documents[document.ID()]; duplicate {
				return ErrInvalidStore
			}
			documents[document.ID()] = document
		}
		stage = "witness-live-document-census"
		installedCount := 0
		for _, allocation := range allocations {
			if allocation.Backing == "installed-type" {
				installedCount++
			}
			if allocation.State != "live" {
				continue
			}
			uri, err := graph.CanonicalPublicResourceURL(s.ScopeURL(), allocation.PublicID)
			if err != nil {
				return err
			}
			if document, known := documents[uri]; known {
				if document.Kind() != allocation.Kind || document.Revision() != allocation.Revision {
					return ErrInvalidStore
				}
			} else {
				switch allocation.Backing {
				case "issue", "generic", "dependency", "informational":
				default:
					return fmt.Errorf("%w: allocation omitted from checked definition census: %s", ErrInvalidStore, allocation.Path)
				}
				record, err := s.readCurrentVerifiedInTx(readctx, tx, allocation.Path)
				if err != nil {
					return err
				}
				document, err := s.projectCheckedNativeInTx(readctx, tx, allocation.Path, allocation.Revision, record, capture.installed, budget)
				if err != nil {
					return err
				}
				if document.ID() != uri || document.Kind() != allocation.Kind || document.Revision() != allocation.Revision {
					return ErrInvalidStore
				}
				documents[uri] = document
			}
		}
		// The expected source population includes actual dynamic native bindings.
		// Configured-but-unused classifications do not become installed records.
		if installedCount != len(sources) {
			return fmt.Errorf("%w: installed source/allocation census", ErrInvalidStore)
		}
		stage = "witness-population-correspondence"
		if err = checkInstalledWitnessPopulation(s.ScopeURL(), allocations, documents); err != nil {
			return err
		}
		for _, allocation := range allocations {
			if allocation.State == "live" && allocation.Kind == graph.KindBead {
				switch allocation.Backing {
				case "intrinsic", "installed-type", "type-definition":
					uri, err := graph.CanonicalPublicResourceURL(s.ScopeURL(), allocation.PublicID)
					if err != nil {
						return err
					}
					result.InstalledDefinitionIDs = append(result.InstalledDefinitionIDs, uri)
				}
			}
		}
		stage = "witness-type-fact-correspondence"
		for _, fact := range capture.snapshot.Types {
			definition, present := documents[fact.Definition.URI]
			if !present || definition.Revision() != fact.Definition.Revision || definition.Kind() != graph.KindBead {
				return ErrInvalidStore
			}
			result.Types = append(result.Types, installedWitnessType{fact.Reference, fact.Definition, fact.Describes, fact.Current, fact.Abstract, append([]resourceget.TypeReference(nil), fact.Ancestors...)})
		}
		// Current bodies come from checked stored snapshots/domain hydration. The
		// wrapper never rewrites their JSON, exact pins, nulls or owned captures.
		ids := make([]string, 0, len(documents))
		for id := range documents {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		stage = "witness-document-output"
		for _, id := range ids {
			document := documents[id]
			result.Documents = append(result.Documents, installedWitnessDocument{document.Kind(), document.NativeJSON()})
		}
		// Exact endpoint pins are preserved. A floating endpoint maps to its actual
		// same-transaction live allocation revision, explicitly recorded here; an
		// authored historical pin is read/validated at that exact revision instead.
		// No URI basename or current revision substitutes for a pinned endpoint.
		stage = "witness-endpoint-correspondence"
		for _, id := range ids {
			document := documents[id]
			if document.Kind() != graph.KindLink {
				continue
			}
			references, err := installedWitnessLinkEndpoints(document.NativeJSON())
			if err != nil {
				return err
			}
			endpoints := []installedWitnessReference{}
			for _, endpoint := range references {
				reference, err := installedWitnessEndpoint(readctx, tx, s, allocations, documents, endpoint, budget)
				if err != nil {
					return err
				}
				endpoints = append(endpoints, reference)
			}
			result.QueryEndpointMappings = append(result.QueryEndpointMappings, installedWitnessEndpoints{installedWitnessReference{document.ID(), document.Revision()}, endpoints[0], endpoints[1]})
		}
		stage = "witness-retained-predecessors"
		seenPredecessors := map[installedWitnessReference]bool{}
		for _, reference := range predecessors {
			if reference.ID == "" || validateVersionToken(reference.Revision) != nil || seenPredecessors[reference] {
				return ErrInvalidStore
			}
			seenPredecessors[reference] = true
			identity, err := s.resolveCommonResourceSelectorInTx(readctx, tx, reference.ID, "", budget)
			if err != nil {
				return err
			}
			document, err := s.readCommonResourceInTx(readctx, tx, identity, reference.Revision, budget)
			if err != nil {
				return err
			}
			if document.ID() != reference.ID || document.Revision() != reference.Revision {
				return ErrInvalidStore
			}
			result.RetainedPredecessors = append(result.RetainedPredecessors, installedWitnessDocument{document.Kind(), document.NativeJSON()})
		}
		sort.Strings(result.InstalledDefinitionIDs)
		sort.Slice(result.Types, func(i, j int) bool {
			a, b := result.Types[i], result.Types[j]
			if a.Reference.URI != b.Reference.URI {
				return a.Reference.URI < b.Reference.URI
			}
			return a.Reference.Revision < b.Reference.Revision
		})
		sort.Slice(result.RetainedPredecessors, func(i, j int) bool {
			return bytes.Compare(result.RetainedPredecessors[i].Body, result.RetainedPredecessors[j].Body) < 0
		})
		// Capture every exposed user table, not the narrower historical hardcoded
		// workflowState table list. No native reservations, history, deleted rows,
		// fence, coordination, binding or retained source/member tables are excluded.
		stage = "witness-whole-user-table-state"
		result.NativeState, err = installedWitnessWholeState(readctx, tx, budget)
		return err
	})
	if err == nil {
		err = ctx.Err()
	}
	if err != nil {
		return installedNativeWitness{}, fmt.Errorf("installed native witness %s (remaining=%d metadata=%d): %w", stage, remaining, metadata, err)
	}
	if require333Absent {
		// A separate public point-read crosscheck is allowed only under exclusive
		// fixture custody. It is not used to populate the expected inventory.
		document, lookupErr := s.GetResourceDocument(ctx, "333", "")
		if !errors.Is(lookupErr, ErrNotFound) || document.ID() != "" {
			return installedNativeWitness{}, fmt.Errorf("%w: refused 333 lookup: %v", ErrInvalidStore, lookupErr)
		}
		result.Bare333Lookup = "not_found"
	}
	if err = ctx.Err(); err != nil {
		return installedNativeWitness{}, err
	}
	return result, nil
}

func installedWitnessAllocations(ctx context.Context, tx *sql.Tx, budget *intrinsicMetadataBudget) ([]installedWitnessAllocation, error) {
	var count uint64
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM graph_preview_catalog`).Scan(&count); err != nil {
		return nil, err
	}
	if count > 10000 {
		return nil, ErrLimitExceeded
	}
	if err := budget.chargeBody(int(count) * 8192); err != nil {
		return nil, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT path,CASE WHEN backing='issue' THEN backing_key ELSE COALESCE(public_id,path) END,resource_kind,revision,backing,allocation_state FROM graph_preview_catalog ORDER BY path LIMIT 10001`)
	if err != nil {
		return nil, err
	}
	result := []installedWitnessAllocation{}
	seen := map[string]bool{}
	for rows.Next() {
		var a installedWitnessAllocation
		if err = rows.Scan(&a.Path, &a.PublicID, &a.Kind, &a.Revision, &a.Backing, &a.State); err != nil {
			break
		}
		if uint64(len(result)) >= count || validateResourcePath(a.Path) != nil || validatePublicID(a.PublicID) != nil || validateVersionToken(a.Revision) != nil || (a.Kind != graph.KindBead && a.Kind != graph.KindLink) || (a.State != "live" && a.State != "deleted") || seen[a.PublicID] {
			err = ErrInvalidStore
			break
		}
		seen[a.PublicID] = true
		result = append(result, a)
	}
	if err == nil && rows.NextResultSet() {
		err = ErrInvalidStore
	}
	err = errors.Join(err, ctx.Err(), rows.Err(), rows.Close())
	if err != nil {
		return nil, err
	}
	if uint64(len(result)) != count {
		return nil, ErrInvalidStore
	}
	return result, nil
}

func checkInstalledWitnessPopulation(scope string, allocations []installedWitnessAllocation, documents map[string]resourceget.ResourceDocument) error {
	expected := map[string]bool{}
	for _, a := range allocations {
		if a.State != "live" {
			continue
		}
		id, err := graph.CanonicalPublicResourceURL(scope, a.PublicID)
		if err != nil {
			return err
		}
		document, ok := documents[id]
		if !ok || expected[id] || document.ID() != id || document.Revision() != a.Revision || document.Kind() != a.Kind {
			return fmt.Errorf("%w: independent allocation/document correspondence", ErrInvalidStore)
		}
		expected[id] = true
	}
	if len(expected) != len(documents) {
		return fmt.Errorf("%w: document without live allocation", ErrInvalidStore)
	}
	return nil
}

// Reuse the native closed endpoint carrier: floating is a URI string, pinned
// is exactly {uri,revision}. Never decode absence/null/unknown members as an
// unpinned endpoint or rewrite the original complete Link body.
func installedWitnessLinkEndpoints(raw []byte) ([2]EndpointReference, error) {
	var fields struct {
		Source EndpointReference `json:"source"`
		Target EndpointReference `json:"target"`
	}
	if err := json.Unmarshal(raw, &fields); err != nil {
		return [2]EndpointReference{}, errors.Join(ErrInvalidStore, err)
	}
	if fields.Source.URI == "" || fields.Target.URI == "" {
		return [2]EndpointReference{}, fmt.Errorf("%w: witness Link endpoint absent", ErrInvalidStore)
	}
	return [2]EndpointReference{fields.Source, fields.Target}, nil
}

func installedWitnessEndpoint(ctx context.Context, tx *sql.Tx, s *Store, allocations []installedWitnessAllocation, documents map[string]resourceget.ResourceDocument, endpoint EndpointReference, budget *intrinsicMetadataBudget) (installedWitnessReference, error) {
	if endpoint.URI == "" {
		return installedWitnessReference{}, ErrInvalidStore
	}
	if endpoint.Revision == "" {
		document, present := documents[endpoint.URI]
		if !present || document.Kind() != graph.KindBead {
			return installedWitnessReference{}, fmt.Errorf("%w: unavailable floating endpoint", ErrInvalidStore)
		}
		return installedWitnessReference{document.ID(), document.Revision()}, nil
	}
	if validateVersionToken(endpoint.Revision) != nil {
		return installedWitnessReference{}, ErrInvalidStore
	}
	for _, a := range allocations {
		uri, err := graph.CanonicalPublicResourceURL(s.ScopeURL(), a.PublicID)
		if err != nil {
			return installedWitnessReference{}, err
		}
		if uri != endpoint.URI {
			continue
		}
		if a.Kind != graph.KindBead {
			return installedWitnessReference{}, ErrInvalidStore
		}
		if document, ok := documents[uri]; ok && document.Revision() == endpoint.Revision {
			return installedWitnessReference{uri, endpoint.Revision}, nil
		}
		identity := AllocatedResourceIdentity{PublicID: a.PublicID, PrivatePath: a.Path, Kind: a.Kind, State: a.State}
		document, err := s.readCommonResourceInTx(ctx, tx, identity, endpoint.Revision, budget)
		if err != nil {
			return installedWitnessReference{}, err
		}
		if document.ID() != uri || document.Revision() != endpoint.Revision || document.Kind() != graph.KindBead {
			return installedWitnessReference{}, ErrInvalidStore
		}
		return installedWitnessReference{uri, endpoint.Revision}, nil
	}
	return installedWitnessReference{}, fmt.Errorf("%w: endpoint has no allocation", ErrInvalidStore)
}

func installedWitnessWholeState(ctx context.Context, tx *sql.Tx, budget *intrinsicMetadataBudget) (map[string]installedWitnessTable, error) {
	rows, err := tx.QueryContext(ctx, "SHOW TABLES")
	if err != nil {
		return nil, err
	}
	tables := []string{}
	for rows.Next() {
		var table string
		if err = rows.Scan(&table); err != nil {
			break
		}
		if len(tables) >= 256 || table == "" {
			err = ErrLimitExceeded
			break
		}
		tables = append(tables, table)
	}
	if err == nil && rows.NextResultSet() {
		err = ErrInvalidStore
	}
	err = errors.Join(err, ctx.Err(), rows.Err(), rows.Close())
	if err != nil {
		return nil, err
	}
	sort.Strings(tables)
	result := map[string]installedWitnessTable{}
	// Reuse complete encoded cells across rows/tables, never discard a cell or row.
	// The shared operation allowance debits every retained unique body/index and
	// every row/reference below; no separate allowance or credit is introduced.
	cells := map[[sha256.Size]byte][]*installedWitnessCell{}
	for _, table := range tables {
		if _, duplicate := result[table]; duplicate {
			return nil, ErrInvalidStore
		}
		// Identifier is a server-returned table name, quoted as one identifier.
		query := "SELECT * FROM `" + strings.ReplaceAll(table, "`", "``") + "` LIMIT 10001"
		rows, err = tx.QueryContext(ctx, query)
		if err != nil {
			return nil, err
		}
		columns, e := rows.Columns()
		if e != nil {
			return nil, errors.Join(e, rows.Close())
		}
		state := installedWitnessTable{Columns: append([]string(nil), columns...), Rows: [][]*installedWitnessCell{}}
		for rows.Next() {
			if len(state.Rows) >= 10000 {
				err = ErrLimitExceeded
				break
			}
			fields := make([]any, len(columns))
			pointers := make([]any, len(columns))
			for i := range fields {
				pointers[i] = &fields[i]
			}
			if err = rows.Scan(pointers...); err != nil {
				break
			}
			row, e := installedWitnessInternRow(fields, cells, budget)
			if e != nil {
				err = fmt.Errorf("whole-state table %s row %d remaining=%d metadata=%d: %w", table, len(state.Rows), budget.remaining, budget.metadata, e)
				break
			}
			state.Rows = append(state.Rows, row)
		}
		if err == nil && rows.NextResultSet() {
			err = ErrInvalidStore
		}
		err = errors.Join(err, ctx.Err(), rows.Err(), rows.Close())
		if err != nil {
			return nil, err
		}
		sort.Slice(state.Rows, func(i, j int) bool { return installedWitnessRowLess(state.Rows[i], state.Rows[j]) })
		result[table] = state
	}
	// These are required semantic tables, not an inclusion whitelist. Every
	// additional user table remains captured above, including future reservations.
	for _, required := range []string{"issues", "dependencies", "events", "leases", "issue_versions", "graph_preview_scope", "graph_preview_catalog", "graph_preview_versions", "graph_preview_issue_versions", "graph_preview_intrinsic_contexts", "graph_preview_intrinsic_members", "graph_preview_types", "graph_preview_payloads"} {
		if _, present := result[required]; !present {
			return nil, fmt.Errorf("%w: missing workflow-state table %s", ErrInvalidStore, required)
		}
	}
	return result, nil
}

// installedWitnessInternRow preserves the exact json.Marshal(fields) cell
// encoding. Only byte-identical complete JSON cells share retained storage.
// Lossless owned SQL cells marshal back to the same row JSON array, including
// repeated/empty/null cells, base64 SQL []byte cells and row multiplicity.
func installedWitnessInternRow(fields []any, cells map[[sha256.Size]byte][]*installedWitnessCell, budget *intrinsicMetadataBudget) ([]*installedWitnessCell, error) {
	// Retain the previous 256-byte row custody debit plus explicit slice slots.
	if err := budget.chargeBody(256 + 32*len(fields)); err != nil {
		return nil, err
	}
	row := make([]*installedWitnessCell, len(fields))
	for i, field := range fields {
		raw, err := json.Marshal(field)
		if err != nil {
			return nil, err
		}
		// The fixed-size digest is only a bucket index, never state evidence.
		// Compare the complete retained bytes even inside a colliding bucket.
		key := sha256.Sum256(raw)
		reused := false
		for _, existing := range cells[key] {
			if bytes.Equal(installedWitnessCellJSON(existing), raw) {
				row[i] = existing
				reused = true
				break
			}
		}
		if reused {
			continue
		}
		// Retain exactly ONE lossless body, debit its actual representation
		// BEFORE copying. JSON expansion is transient and reproduced exactly
		// by standard encoding/json at comparison/output; never a hash surrogate.
		retained := len(raw)
		switch value := field.(type) {
		case string:
			retained = len(value)
		case []byte:
			retained = len(value)
		}
		if err = budget.chargeBody(retained + 256); err != nil {
			return nil, err
		}
		stored := &installedWitnessCell{raw: raw}
		switch value := field.(type) {
		case string:
			stored.kind, stored.raw, stored.text = 1, nil, strings.Clone(value)
		case []byte:
			stored.kind, stored.raw = 2, nil
			// Preserve nil versus nonnil empty bytes. Never retain mutable driver
			// scratch: the owned complete byte copy is made AFTER its debit.
			if value != nil {
				stored.binary = make([]byte, len(value))
				copy(stored.binary, value)
			}
		}
		cells[key] = append(cells[key], stored)
		row[i] = stored
	}
	return row, nil
}

// Compare the exact virtual JSON row bytes without retaining another expanded
// row for a sort key. This is the previous bytes.Compare(json.Marshal(fields))
// ordering, including prefix cases inside numeric/structured cell encodings.
func installedWitnessRowLess(a, b []*installedWitnessCell) bool {
	next := func(row []*installedWitnessCell) func() (byte, bool) {
		cell, offset := -1, 0
		var encoded []byte
		closed := false
		return func() (byte, bool) {
			if cell < 0 {
				cell = 0
				return '[', true
			}
			if cell == len(row) {
				if closed {
					return 0, false
				}
				closed = true
				return ']', true
			}
			if encoded == nil {
				encoded = installedWitnessCellJSON(row[cell])
			}
			if offset < len(encoded) {
				value := encoded[offset]
				offset++
				return value, true
			}
			cell++
			offset = 0
			encoded = nil
			if cell < len(row) {
				return ',', true
			}
			closed = true
			return ']', true
		}
	}
	left, right := next(a), next(b)
	for {
		x, okX := left()
		y, okY := right()
		if !okX || !okY {
			return !okX && okY
		}
		if x != y {
			return x < y
		}
	}
}

// Standard encoding/json and expanded raw-byte sorting are the independent
// oracle for lossless interning and the virtual comparator used by the census.
func checkInstalledWitnessCellReuse(t *testing.T) {
	t.Helper()
	rows := [][]any{
		{nil, "", []byte{}, []byte{0, 1, 127, 255}, "中文<&>\u2028", 1, json.Number("1.0"), 10},
		{1, 10}, {10, 1}, {1}, {1, 0}, {json.Number("1.0"), 1},
		{"a", "ab"}, {"ab", "a"}, {"a"}, {"a", nil},
		{[]any{1, "<&>", nil}, map[string]any{"nested": []any{nil, true}, "a": ""}},
		{}, {nil, "", []byte{}, []byte{0, 1, 127, 255}, "中文<&>\u2028", 1, json.Number("1.0"), 10},
		{1, 10},
	}
	var expanded []json.RawMessage
	var interned [][]*installedWitnessCell
	cells := map[[sha256.Size]byte][]*installedWitnessCell{}
	budget := newIntrinsicMetadataBudget()
	encodedCells := map[string]bool{}
	for _, fields := range rows {
		raw, err := json.Marshal(fields)
		if err != nil {
			t.Fatal(err)
		}
		expanded = append(expanded, raw)
		row, err := installedWitnessInternRow(fields, cells, budget)
		if err != nil {
			t.Fatal(err)
		}
		actual, err := json.Marshal(row)
		if err != nil || !bytes.Equal(actual, raw) {
			t.Fatalf("interning changed complete row JSON: want %s got %s err %v", raw, actual, err)
		}
		for i, field := range fields {
			encoded, err := json.Marshal(field)
			if err != nil || !bytes.Equal(encoded, installedWitnessCellJSON(row[i])) {
				t.Fatalf("distinct/null/empty cell changed: %s %s %v", encoded, installedWitnessCellJSON(row[i]), err)
			}
			encodedCells[string(encoded)] = true
		}
		interned = append(interned, row)
	}
	retainedCells := 0
	for _, bucket := range cells {
		retainedCells += len(bucket)
	}
	if retainedCells != len(encodedCells) {
		t.Fatal("unequal encoded cells merged or repeated cells not reused")
	}
	for _, row := range interned {
		for _, cell := range row {
			found := false
			for _, stored := range cells[sha256.Sum256(installedWitnessCellJSON(cell))] {
				if bytes.Equal(installedWitnessCellJSON(stored), installedWitnessCellJSON(cell)) {
					found = true
					if stored != cell {
						t.Fatal("equal complete cell bytes were retained twice")
					}
				}
			}
			if !found {
				t.Fatal("complete retained cell absent from its bucket")
			}
		}
	}
	// Exercise ordering independently in both directions for every pair,
	// including cell-prefix numbers and row-end versus next-cell delimiters.
	for i := range rows {
		for j := range rows {
			if installedWitnessRowLess(interned[i], interned[j]) != (bytes.Compare(expanded[i], expanded[j]) < 0) {
				t.Fatalf("virtual row order differs from complete raw JSON at %d/%d", i, j)
			}
		}
	}
	sort.Slice(expanded, func(i, j int) bool { return bytes.Compare(expanded[i], expanded[j]) < 0 })
	sort.Slice(interned, func(i, j int) bool { return installedWitnessRowLess(interned[i], interned[j]) })
	want, err := json.Marshal(expanded)
	if err != nil {
		t.Fatal(err)
	}
	actual, err := json.Marshal(interned)
	if err != nil || len(interned) != len(rows) || !bytes.Equal(want, actual) {
		t.Fatalf("whole sorted state/multiplicity changed: %s %s %v", want, actual, err)
	}
	// The first cell is retained, the second refuses. Failure returns no row
	// and cannot refund the already charged row and first unique-cell custody.
	tiny := &intrinsicMetadataBudget{remaining: 600}
	partial := map[[sha256.Size]byte][]*installedWitnessCell{}
	refused, err := installedWitnessInternRow([]any{"a", "b"}, partial, tiny)
	kept := partial[sha256.Sum256([]byte(`"a"`))]
	if !errors.Is(err, ErrLimitExceeded) || refused != nil || tiny.remaining != 23 || tiny.metadata != 0 || len(partial) != 1 || len(kept) != 1 || string(installedWitnessCellJSON(kept[0])) != `"a"` {
		t.Fatalf("cell custody refusal returned partial row or refunded budget: %v %v %+v", refused, err, tiny)
	}
}

// Exercise the real full-table reader with a previously debited SAME allowance.
// Standard JSON of the original SQL values is the independent complete-state
// oracle; no table, column, duplicate row, byte or output encoding may disappear.
func TestInstalledWitnessWholeStateRetainsSQLCellBytesWithinSharedBudget(t *testing.T) {
	tables := []string{"issues", "dependencies", "events", "leases", "issue_versions", "graph_preview_scope", "graph_preview_catalog", "graph_preview_versions", "graph_preview_issue_versions", "graph_preview_intrinsic_contexts", "graph_preview_intrinsic_members", "graph_preview_types", "graph_preview_payloads", "future_user_state"}
	sort.Strings(tables)
	payload := bytes.Repeat([]byte{0xff}, 4096)
	original := []any{"beads/a", payload, nil}
	expectedRow, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	expectedRows, err := json.Marshal([]json.RawMessage{expectedRow, expectedRow})
	if err != nil {
		t.Fatal(err)
	}
	read := func(initial int) (map[string]installedWitnessTable, *intrinsicMetadataBudget, error) {
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		if err != nil {
			t.Fatal(err)
		}
		mock.ExpectBegin()
		tx, err := db.Begin()
		if err != nil {
			t.Fatal(err)
		}
		budget := &intrinsicMetadataBudget{remaining: initial}
		// Retain an earlier witness-stage body, with its actual existing debit.
		previousBody := bytes.Repeat([]byte{'p'}, 1024)
		if err := budget.chargeBody(len(previousBody)); err != nil {
			t.Fatal(err)
		}
		names := sqlmock.NewRows([]string{"Tables"})
		for _, table := range tables {
			names.AddRow(table)
		}
		mock.ExpectQuery("SHOW TABLES").WillReturnRows(names).RowsWillBeClosed()
		for _, table := range tables {
			rows := sqlmock.NewRows([]string{"path", "properties", "metadata"})
			if table == "graph_preview_payloads" {
				rows.AddRow("beads/a", payload, nil).AddRow("beads/a", payload, nil)
			}
			mock.ExpectQuery("SELECT * FROM `" + table + "` LIMIT 10001").WillReturnRows(rows).RowsWillBeClosed()
			if initial < 7000 && table == "graph_preview_payloads" {
				break
			}
		}
		mock.ExpectRollback()
		mock.ExpectClose()
		result, readErr := installedWitnessWholeState(context.Background(), tx, budget)
		if len(previousBody) != 1024 || previousBody[0] != 'p' {
			t.Fatal("earlier retained body changed")
		}
		readErr = errors.Join(readErr, tx.Rollback(), db.Close())
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
		return result, budget, readErr
	}
	state, budget, err := read(7000)
	if err != nil {
		t.Fatal(err)
	}
	if len(state) != len(tables) {
		t.Fatal("full user-table membership lost")
	}
	for _, table := range tables {
		snapshot, ok := state[table]
		if !ok || !reflect.DeepEqual(snapshot.Columns, []string{"path", "properties", "metadata"}) {
			t.Fatal("table/column bytes changed", table)
		}
		if table != "graph_preview_payloads" && len(snapshot.Rows) != 0 {
			t.Fatal("empty table changed", table)
		}
	}
	got, err := json.Marshal(state["graph_preview_payloads"].Rows)
	if err != nil || !bytes.Equal(got, expectedRows) {
		t.Fatalf("full rows/SQL bytes/base64/null/multiplicity changed: %v", err)
	}
	// Caller/driver buffers must not remain authority over retained state.
	payload[0] = 0
	afterMutation, err := json.Marshal(state["graph_preview_payloads"].Rows)
	if err != nil || !bytes.Equal(afterMutation, expectedRows) {
		t.Fatal("retained table aliases mutable SQL bytes", err)
	}
	payload[0] = 0xff
	// 7000 - earlier1024 - two row352 - path(7+256) - binary(4096+256)
	// - null(4+256) =397. The second row reuses cells but still pays ALL slots.
	if budget.remaining != 397 || budget.metadata != 0 {
		t.Fatalf("cumulative retained custody changed: %+v", budget)
	}
	// One byte below the unchanged first row reservation: return ZERO state,
	// close rows/transaction, preserve the previous-stage debit, and never refund.
	refused, tiny, err := read(1024 + 351)
	if !errors.Is(err, ErrLimitExceeded) || refused != nil || tiny.remaining != 351 || tiny.metadata != 0 {
		t.Fatalf("partial state/refund on shared refusal: %v %v %+v", refused, err, tiny)
	}
}

// Compare the COMPLETE witness evidence, not its private SQL cell carrier.
// All non-state fields retain the original exact reflect.DeepEqual semantics.
// The shallow struct copies below copy headers only and do not mutate inputs.
// NativeState is compared in full, table by table / column by column / row by
// row / cell by cell, with the same complete JSON equivalence as the old raw
// cells. No whole-witness/state JSON copy or hash-only surrogate is created.
func installedNativeWitnessEqual(a, b installedNativeWitness) bool {
	left, right := a.NativeState, b.NativeState
	a.NativeState, b.NativeState = nil, nil
	if !reflect.DeepEqual(a, b) || (left == nil) != (right == nil) || len(left) != len(right) {
		return false
	}
	for name, table := range left {
		other, present := right[name]
		if !present || !reflect.DeepEqual(table.Columns, other.Columns) || (table.Rows == nil) != (other.Rows == nil) || len(table.Rows) != len(other.Rows) {
			return false
		}
		for i, row := range table.Rows {
			next := other.Rows[i]
			if (row == nil) != (next == nil) || len(row) != len(next) {
				return false
			}
			for j, cell := range row {
				candidate := next[j]
				if (cell == nil) != (candidate == nil) {
					return false
				}
				if cell != nil && !bytes.Equal(installedWitnessCellJSON(cell), installedWitnessCellJSON(candidate)) {
					return false
				}
			}
		}
	}
	return true
}

// The complete witness comparison must ignore the first SQL representative
// of a complete JSON cell class, while retaining EVERY other witness/state fact.
func TestInstalledNativeWitnessEqualityIgnoresCellRepresentation(t *testing.T) {
	rows := [][]any{
		{"", []byte{1}}, {"x", []byte{}},
		{nil, "nil"}, {[]byte(nil), "typednil"},
		{string([]byte{0xff}), "invalid-a"}, {string([]byte{0xfe}), "invalid-b"},
	}
	expected := make([]json.RawMessage, 0, len(rows))
	for _, fields := range rows {
		raw, err := json.Marshal(fields)
		if err != nil {
			t.Fatal(err)
		}
		expected = append(expected, raw)
	}
	sort.Slice(expected, func(i, j int) bool { return bytes.Compare(expected[i], expected[j]) < 0 })
	want, err := json.Marshal(expected)
	if err != nil {
		t.Fatal(err)
	}
	capture := func(reverse bool) installedNativeWitness {
		cells := map[[sha256.Size]byte][]*installedWitnessCell{}
		budget := newIntrinsicMetadataBudget()
		state := installedWitnessTable{Columns: []string{"first", "second"}, Rows: [][]*installedWitnessCell{}}
		for n := range rows {
			i := n
			if reverse {
				i = len(rows) - 1 - n
			}
			row, err := installedWitnessInternRow(rows[i], cells, budget)
			if err != nil {
				t.Fatal(err)
			}
			state.Rows = append(state.Rows, row)
		}
		sort.Slice(state.Rows, func(i, j int) bool { return installedWitnessRowLess(state.Rows[i], state.Rows[j]) })
		actual, err := json.Marshal(state.Rows)
		if err != nil || !bytes.Equal(actual, want) {
			t.Fatal("opposite scan changed complete sorted row JSON/multiplicity", err)
		}
		return installedNativeWitness{
			Origin: "native-installed-and-domain-witness", Scope: "https://example.test/scope/",
			Documents:              []installedWitnessDocument{{Kind: graph.KindBead, Body: json.RawMessage(`{"id":"whole-bead","revision":"exact"}`)}},
			Types:                  []installedWitnessType{{Reference: resourceget.TypeReference{URI: "whole-type", Revision: "exact"}, Definition: resourceget.TypeReference{URI: "whole-definition", Revision: "exact"}, Describes: graph.KindBead, Current: true, Abstract: false, Ancestors: []resourceget.TypeReference{{URI: "whole-ancestor", Revision: "exact"}}}},
			InstalledDefinitionIDs: []string{"whole-definition"},
			QueryEndpointMappings:  []installedWitnessEndpoints{{Link: installedWitnessReference{"whole-link", "exact"}, Source: installedWitnessReference{"whole-source", "exact"}, Target: installedWitnessReference{"whole-target", "exact"}}},
			NativeState:            map[string]installedWitnessTable{"all-state": state},
			RetainedPredecessors:   []installedWitnessDocument{{Kind: graph.KindBead, Body: json.RawMessage(`{"revision":"previous"}`)}},
			Bare333Lookup:          "not_found", SourceClosure: json.RawMessage(`{"source":"exact"}`),
		}
	}
	left, right := capture(false), capture(true)
	if reflect.DeepEqual(left.NativeState, right.NativeState) {
		t.Fatal("fixture failed to choose different private equivalent-cell representatives")
	}
	if !installedNativeWitnessEqual(left, right) || !installedNativeWitnessEqual(right, left) {
		t.Fatal("opposite-order JSON-equivalent witness falsely differs")
	}
	// Mutate every top-level field independently. NativeState remains present in
	// the SAME comparison used for before/after/refused333 and normal reopen.
	mutations := []func(*installedNativeWitness){
		func(w *installedNativeWitness) { w.Origin = "changed" },
		func(w *installedNativeWitness) { w.Scope = "changed" },
		func(w *installedNativeWitness) { w.Documents = nil },
		func(w *installedNativeWitness) { w.Types = nil },
		func(w *installedNativeWitness) { w.InstalledDefinitionIDs = nil },
		func(w *installedNativeWitness) { w.QueryEndpointMappings = nil },
		func(w *installedNativeWitness) { w.NativeState = nil },
		func(w *installedNativeWitness) { w.RetainedPredecessors = nil },
		func(w *installedNativeWitness) { w.Bare333Lookup = "changed" },
		func(w *installedNativeWitness) { w.SourceClosure = json.RawMessage(`{"source":"changed"}`) },
	}
	for i, mutate := range mutations {
		changed := left
		mutate(&changed)
		if installedNativeWitnessEqual(left, changed) || installedNativeWitnessEqual(changed, left) {
			t.Fatal("complete witness field omitted from equality", i)
		}
	}
	// Table membership, columns, row order, multiplicity and complete cell bytes
	// must independently refuse a difference; no normalization drops evidence.
	state := left.NativeState["all-state"]
	states := []map[string]installedWitnessTable{
		{"other-state": state},
		{"all-state": {Columns: []string{"changed", "second"}, Rows: state.Rows}},
		{"all-state": {Columns: state.Columns, Rows: append(append([][]*installedWitnessCell(nil), state.Rows...), state.Rows[0])}},
	}
	swapped := append([][]*installedWitnessCell(nil), state.Rows...)
	swapped[0], swapped[1] = swapped[1], swapped[0]
	states = append(states, map[string]installedWitnessTable{"all-state": {Columns: state.Columns, Rows: swapped}})
	changedRows := append([][]*installedWitnessCell(nil), state.Rows...)
	changedRows[0] = append([]*installedWitnessCell(nil), changedRows[0]...)
	changedRows[0][0] = &installedWitnessCell{kind: 1, text: "different complete cell"}
	states = append(states, map[string]installedWitnessTable{"all-state": {Columns: state.Columns, Rows: changedRows}})
	for i, changedState := range states {
		changed := left
		changed.NativeState = changedState
		if installedNativeWitnessEqual(left, changed) || installedNativeWitnessEqual(changed, left) {
			t.Fatal("complete table/column/row/cell evidence omitted", i)
		}
	}
	// Preserve prior nil-versus-empty table-row/column structure semantics too.
	nilRows := installedNativeWitness{NativeState: map[string]installedWitnessTable{"empty": {Columns: []string{}, Rows: nil}}}
	emptyRows := installedNativeWitness{NativeState: map[string]installedWitnessTable{"empty": {Columns: []string{}, Rows: [][]*installedWitnessCell{}}}}
	if installedNativeWitnessEqual(nilRows, emptyRows) {
		t.Fatal("nil/empty row distinction lost")
	}
	nilColumns := emptyRows
	nilColumns.NativeState = map[string]installedWitnessTable{"empty": {Columns: nil, Rows: [][]*installedWitnessCell{}}}
	if installedNativeWitnessEqual(nilColumns, emptyRows) {
		t.Fatal("nil/empty column distinction lost")
	}
	nilMap := installedNativeWitness{}
	emptyMap := installedNativeWitness{NativeState: map[string]installedWitnessTable{}}
	if installedNativeWitnessEqual(nilMap, emptyMap) {
		t.Fatal("nil/empty NativeState distinction lost")
	}
}

func TestInstalledWitnessCellReuse(t *testing.T) {
	checkInstalledWitnessCellReuse(t)
}

func TestInstalledNativeWitnessCensusStateAndReopen(t *testing.T) {
	requestPath, requestPresent := os.LookupEnv(installedWitnessRequestEnv)
	requestSHA, shaPresent := os.LookupEnv(installedWitnessRequestSHAEnv)
	if requestPresent || shaPresent {
		if !requestPresent || !shaPresent || requestPath == "" || requestSHA == "" {
			t.Fatal("explicit witness request requires pointer AND exact digest")
		}
		runInstalledNativeWitnessRequest(t, requestPath, requestSHA)
		return
	}
	checkInstalledWitnessCellReuse(t)
	ctx, options, s := intrinsicHardeningStore(t)
	initial, err := readInstalledNativeWitness(ctx, s, nil, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(initial.InstalledDefinitionIDs) == 0 || len(initial.Types) == 0 || len(initial.QueryEndpointMappings) == 0 {
		t.Fatal("missing actual installed definition/fact/Refines census")
	}
	// Deliberately remove an unreferenced installed record from the candidate
	// output. The independent allocation population must still require it.
	err = s.withDefinitionReadTx(ctx, txDefinitionRead, func(tx *sql.Tx, b *intrinsicMetadataBudget) error {
		allocations, e := installedWitnessAllocations(ctx, tx, b)
		if e != nil {
			return e
		}
		capture, e := s.captureTypeRecordsInTx(ctx, tx, b)
		if e != nil {
			return e
		}
		documents := map[string]resourceget.ResourceDocument{}
		for _, d := range capture.snapshot.Documents {
			documents[d.ID()] = d
		}
		if e = checkInstalledWitnessPopulation(s.ScopeURL(), allocations, documents); e != nil {
			return e
		}
		removed := false
		for _, a := range allocations {
			if a.Backing == "installed-type" && a.State == "live" {
				id, e := graph.CanonicalPublicResourceURL(s.ScopeURL(), a.PublicID)
				if e != nil {
					return e
				}
				delete(documents, id)
				removed = true
				break
			}
		}
		if !removed {
			return errors.New("fixture has no actual unused installed native Type")
		}
		if !errors.Is(checkInstalledWitnessPopulation(s.ScopeURL(), allocations, documents), ErrInvalidStore) {
			return errors.New("oracle accepted an omitted installed record")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	request := plainIssue("Witness Issue")
	request.Issue.ID = "bd-witness"
	issue, err := s.CreateIssue(ctx, "", request)
	if err != nil {
		t.Fatal(err)
	}
	if _, present := issue.ResourceDocument(); !present {
		t.Fatal("Issue missing original applied facet")
	}
	first, err := s.Create(ctx, CreateRequest{PublicID: "witness-one", Title: "Before", Body: "one", Actor: "witness"})
	if err != nil {
		t.Fatal(err)
	}
	if _, present := first.ResourceDocument(); !present {
		t.Fatal("Memory missing original applied facet")
	}
	second, err := s.Create(ctx, CreateRequest{PublicID: "witness-two", Title: "Two", Body: "two", Actor: "witness"})
	if err != nil {
		t.Fatal(err)
	}
	predecessor, present := second.ResourceDocument()
	if !present {
		t.Fatal("Memory predecessor missing original applied facet")
	}
	link, err := s.AddInformationalLink(ctx, LinkCreateRequest{PublicID: "witness-link", SourcePath: "beads/witness-one", TargetPath: "beads/witness-two", Actor: "witness"})
	if err != nil {
		t.Fatal(err)
	}
	updated, err := s.UpdateMemory(ctx, MemoryUpdateRequest{Path: "beads/witness-two", ExpectedRevision: second.Revision, Properties: Properties{Title: "After", Body: "two"}, Actor: "witness"})
	if err != nil {
		t.Fatal(err)
	}
	if _, present := updated.Memory.ResourceDocument(); !present {
		t.Fatal("Memory successor missing original applied facet")
	}
	retained := []installedWitnessReference{{predecessor.ID(), predecessor.Revision()}}
	before, err := readInstalledNativeWitness(ctx, s, retained, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(before.Documents) <= len(initial.Documents) || len(before.QueryEndpointMappings) <= len(initial.QueryEndpointMappings) {
		t.Fatal("native Issue/Memory/Link additions omitted")
	}
	for _, carrier := range []interface {
		ResourceDocument() (resourceget.ResourceDocument, bool)
	}{issue, updated.Memory, link.Link} {
		applied, present := carrier.ResourceDocument()
		if !present {
			t.Fatal("native applied carrier absent")
		}
		found := false
		for _, document := range before.Documents {
			if bytes.Equal(document.Body, applied.NativeJSON()) {
				found = true
			}
		}
		if !found {
			t.Fatalf("independent census lost exact original applied body %s", applied.ID())
		}
	}
	mapped := false
	for _, endpoints := range before.QueryEndpointMappings {
		if endpoints.Link.ID == link.Link.ID {
			mapped = true
			if endpoints.Source.ID != first.ID || endpoints.Target.ID != second.ID || endpoints.Source.Revision == "" || endpoints.Target.Revision == "" {
				t.Fatal("native endpoint mapping lost exact admitted editions")
			}
		}
	}
	if !mapped {
		t.Fatal("native Link endpoint mapping omitted")
	}
	bad := plainIssue("Refused")
	bad.Issue.ID = "333"
	refused, err := s.CreateIssue(ctx, "", bad)
	if !errors.Is(err, storage.ErrPrefixMismatch) || !reflect.ValueOf(refused).IsZero() {
		t.Fatalf("bare333 result: %+v %v", refused, err)
	}
	after, err := readInstalledNativeWitness(ctx, s, retained, true)
	if err != nil {
		t.Fatal(err)
	}
	if !installedNativeWitnessEqual(before, after) {
		t.Fatal("refused333 changed whole workflow evidence or checked bodies")
	}
	if err = s.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := openInstalledNativeWitness(ctx, options, retained, true)
	if err != nil {
		t.Fatal(err)
	}
	if !installedNativeWitnessEqual(after, reopened) {
		t.Fatal("normal existing reopen changed complete witness")
	}
	if len(reopened.RetainedPredecessors) != 1 || !bytes.Equal(reopened.RetainedPredecessors[0].Body, predecessor.NativeJSON()) {
		t.Fatal("retained exact complete predecessor differs from original applied evidence")
	}
}

// Pure carrier-boundary controls. The unchanged default native fixture also
// writes a real floating informational Link; this is not its runtime result.
func TestInstalledNativeWitnessEndpointCarrierArms(t *testing.T) {
	const source = "https://witness-endpoint.example/source"
	const target = "https://witness-endpoint.example/target"
	sourcePin := strings.Repeat("a", 64)
	targetPin := strings.Repeat("b", 64)
	for _, tc := range []struct {
		name, raw string
		want      [2]EndpointReference
	}{
		{"floating", fmt.Sprintf(`{"source":%q,"target":%q}`, source, target), [2]EndpointReference{{URI: source}, {URI: target}}},
		{"pinned", fmt.Sprintf(`{"source":{"uri":%q,"revision":%q},"target":{"uri":%q,"revision":%q}}`, source, sourcePin, target, targetPin), [2]EndpointReference{{URI: source, Revision: sourcePin}, {URI: target, Revision: targetPin}}},
		{"mixed", fmt.Sprintf(`{"source":%q,"target":{"uri":%q,"revision":%q}}`, source, target, targetPin), [2]EndpointReference{{URI: source}, {URI: target, Revision: targetPin}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := installedWitnessLinkEndpoints([]byte(tc.raw))
			if err != nil || got != tc.want {
				t.Fatalf("native endpoint arm/pin changed: %+v %v", got, err)
			}
		})
	}
	for _, tc := range []struct{ name, raw string }{
		{"pinned-missing-revision", fmt.Sprintf(`{"source":%q,"target":{"uri":%q}}`, source, target)},
		{"pinned-empty-revision", fmt.Sprintf(`{"source":%q,"target":{"uri":%q,"revision":""}}`, source, target)},
		{"pinned-null-revision", fmt.Sprintf(`{"source":%q,"target":{"uri":%q,"revision":null}}`, source, target)},
		{"pinned-extra-member", fmt.Sprintf(`{"source":%q,"target":{"uri":%q,"revision":%q,"extra":true}}`, source, target, targetPin)},
		{"missing-endpoint", fmt.Sprintf(`{"source":%q}`, source)},
		{"null-arm", fmt.Sprintf(`{"source":null,"target":%q}`, target)},
		{"array-arm", fmt.Sprintf(`{"source":[],"target":%q}`, target)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := installedWitnessLinkEndpoints([]byte(tc.raw))
			if !errors.Is(err, ErrInvalidStore) || got != ([2]EndpointReference{}) {
				t.Fatalf("malformed arm became usable floating/pinned reference: %+v %v", got, err)
			}
		})
	}
}

// The two private variables are test-binary entry controls only. Neither is
// read by production code. A hash without its pointer is an error, not fallback.
const installedWitnessRequestEnv = "BEADS_TEST_INSTALLED_NATIVE_WITNESS_REQUEST"
const installedWitnessRequestSHAEnv = "BEADS_TEST_INSTALLED_NATIVE_WITNESS_REQUEST_SHA256"

// Options contains only JSON-safe primitive fields. Embedded-only admission
// below refuses a nonempty password, and output never contains these options.
type installedWitnessRequest struct {
	Protocol              string                      `json:"protocol"`
	ReadMode              string                      `json:"readMode"`
	Phase                 string                      `json:"phase"`
	OwnedRoot             string                      `json:"ownedRoot"`
	FixtureRoot           string                      `json:"fixtureRoot"`
	OutputPath            string                      `json:"outputPath"`
	GrantSHA256           string                      `json:"grantSHA256"`
	SourceSHA256          string                      `json:"sourceSHA256"`
	CustodySHA256         string                      `json:"custodySHA256"`
	MetadataSHA256        string                      `json:"metadataSHA256"`
	MarkerSHA256          string                      `json:"markerSHA256"`
	AllOtherHandlesClosed bool                        `json:"allOtherHandlesClosed"`
	NormalOptions         Options                     `json:"normalOptions"`
	Predecessors          []installedWitnessReference `json:"predecessors"`
	Require333Absent      bool                        `json:"require333Absent"`
}

type installedWitnessCustody struct {
	Protocol              string `json:"protocol"`
	OwnedRoot             string `json:"ownedRoot"`
	FixtureRoot           string `json:"fixtureRoot"`
	Workspace             string `json:"workspace"`
	DataDir               string `json:"dataDir"`
	Scope                 string `json:"scope"`
	Authority             string `json:"authority"`
	GrantSHA256           string `json:"grantSHA256"`
	SourceSHA256          string `json:"sourceSHA256"`
	AllOtherHandlesClosed bool   `json:"allOtherHandlesClosed"`
}

// This is the exact primitive field subset consumed by frozen graphOptions;
// unknown metadata fields are tolerated exactly as current configfile.Load's
// json.Unmarshal into Config. The entire file hash remains bound. We do NOT
// call Load: its absent-metadata legacy branch can migrate/write config.json.
type installedWitnessMetadata struct {
	Backend                 string `json:"backend"`
	GraphMode               string `json:"graph_mode"`
	GraphAssessmentsEnabled bool   `json:"graph_assessments_enabled"`
	GraphScopeURL           string `json:"graph_scope_url"`
	GraphAuthorityID        string `json:"graph_authority_id"`
	GraphWorkspace          string `json:"graph_workspace"`
	GraphSchemaVersion      int    `json:"graph_schema_version"`
	GraphReady              bool   `json:"graph_ready"`
	DoltMode                string `json:"dolt_mode"`
	DoltDatabase            string `json:"dolt_database"`
	DoltServerHost          string `json:"dolt_server_host"`
	DoltServerPort          int    `json:"dolt_server_port"`
	DoltServerUser          string `json:"dolt_server_user"`
	DoltServerSocket        string `json:"dolt_server_socket"`
	DoltServerTLS           bool   `json:"dolt_server_tls"`
}

type installedWitnessCaptureEnvelope struct {
	Protocol      string                  `json:"protocol"`
	Phase         string                  `json:"phase"`
	RequestSHA256 string                  `json:"requestSHA256"`
	GrantSHA256   string                  `json:"grantSHA256"`
	SourceSHA256  string                  `json:"sourceSHA256"`
	Success       bool                    `json:"success"`
	Witness       *installedNativeWitness `json:"witness,omitempty"`
	Error         string                  `json:"error,omitempty"`
}

func runInstalledNativeWitnessRequest(t *testing.T, path, expectedSHA string) {
	t.Helper()
	requestBytes, err := installedWitnessReadBoundFile(path, expectedSHA, 64*1024)
	if err != nil {
		t.Fatal(err)
	}
	var request installedWitnessRequest
	if err = installedWitnessDecodeRequest(requestBytes, &request); err != nil {
		t.Fatal(err)
	}
	if request.Protocol != "installed-native-witness/v1" || request.ReadMode != "checked-definition-read" || !request.AllOtherHandlesClosed {
		t.Fatal("witness request protocol/read/custody refusal")
	}
	switch request.Phase {
	case "P0", "P1", "P2", "P3", "reopen", "bare333-before", "bare333-after":
	default:
		t.Fatal("unsupported witness phase")
	}
	if strings.HasPrefix(request.Phase, "bare333-") && !request.Require333Absent {
		t.Fatal("bare333 phase requires actual not_found crosscheck")
	}
	for _, value := range []string{request.GrantSHA256, request.SourceSHA256, request.CustodySHA256, request.MetadataSHA256, request.MarkerSHA256} {
		if !installedWitnessSHA(value) {
			t.Fatal("missing exact witness binding digest")
		}
	}
	if _, present := os.LookupEnv("BEADS_DOLT_PASSWORD"); present {
		t.Fatal("embedded witness refuses password environment")
	}
	if !filepath.IsAbs(request.OwnedRoot) || filepath.Dir(request.OwnedRoot) != "/private/var/tmp" {
		t.Fatal("witness requires a controller-owned outside-home root")
	}
	if !installedWitnessInside(request.OwnedRoot, request.FixtureRoot) || request.FixtureRoot == request.OwnedRoot || filepath.Dir(path) != request.OwnedRoot || filepath.Dir(request.OutputPath) != request.OwnedRoot || request.OutputPath == path {
		t.Fatal("witness request/output/fixture containment refusal")
	}
	workspace := filepath.Join(request.FixtureRoot, ".beads")
	dataDir := filepath.Join(workspace, "embeddeddolt")
	custodyPath := filepath.Join(request.OwnedRoot, "installed-native-witness-custody.json")
	if request.OutputPath == custodyPath {
		t.Fatal("witness output collides with custody record")
	}
	physical := map[string]os.FileInfo{}
	for _, directory := range []string{request.OwnedRoot, request.FixtureRoot, workspace, dataDir} {
		info, e := os.Lstat(directory)
		actual, realErr := filepath.EvalSymlinks(directory)
		if e != nil || realErr != nil || actual != directory || filepath.Clean(directory) != directory || !info.IsDir() {
			t.Fatal("witness directory identity refusal")
		}
		if directory == request.OwnedRoot && info.Mode().Perm() != 0700 {
			t.Fatal("witness owned root must be private")
		}
		physical[directory] = info
	}
	custodyBytes, err := installedWitnessReadBoundFile(custodyPath, request.CustodySHA256, 64*1024)
	if err != nil {
		t.Fatal(err)
	}
	var custody installedWitnessCustody
	if err = installedWitnessDecodeRequest(custodyBytes, &custody); err != nil {
		t.Fatal(err)
	}
	expectedCustody := installedWitnessCustody{"installed-native-witness-custody/v1", request.OwnedRoot, request.FixtureRoot, workspace, dataDir, request.NormalOptions.Binding.ScopeURL, request.NormalOptions.Binding.AuthorityID, request.GrantSHA256, request.SourceSHA256, true}
	if custody != expectedCustody {
		t.Fatal("witness supplied fixture/custody mismatch")
	}
	metadataPath := filepath.Join(workspace, "metadata.json")
	metadataBytes, err := installedWitnessReadBoundFile(metadataPath, request.MetadataSHA256, 1024*1024)
	if err != nil {
		t.Fatal(err)
	}
	var metadata installedWitnessMetadata
	if err = json.Unmarshal(metadataBytes, &metadata); err != nil {
		t.Fatal("witness metadata is not current JSON")
	}
	markerPath := filepath.Join(workspace, "graph-preview-format")
	markerBytes, err := installedWitnessReadBoundFile(markerPath, request.MarkerSHA256, 1024)
	if err != nil {
		t.Fatal(err)
	}
	if string(markerBytes) != "link-preview-v13\n" || TypeGeneration != 13 || metadata.Backend != "dolt" || metadata.GraphMode != "link" || !metadata.GraphReady || metadata.DoltMode != "embedded" || metadata.GraphSchemaVersion != TypesFreshSchemaVersion || metadata.GraphWorkspace != workspace {
		t.Fatal("witness requires exact normally initialized current embedded Types fixture")
	}
	// Field-for-field frozen graph_preview.go:graphOptions constructor; no
	// guessed defaults, environment credentials, init-only IssuePrefix, server
	// fallback or new storage option is introduced. Actual metadata supplies all
	// fields; the request must also equal the controller's normal resolved value.
	expectedOptions := Options{Backend: metadata.DoltMode, DataDir: dataDir, Database: metadata.DoltDatabase, Branch: "main", Binding: Binding{WorkspaceID: metadata.GraphWorkspace, ScopeURL: metadata.GraphScopeURL, AuthorityID: metadata.GraphAuthorityID, SchemaVersion: metadata.GraphSchemaVersion}, TypeGeneration: TypeGeneration, AssessmentsEnabled: metadata.GraphAssessmentsEnabled, ServerHost: metadata.DoltServerHost, ServerPort: metadata.DoltServerPort, ServerUser: metadata.DoltServerUser, ServerSocket: metadata.DoltServerSocket, ServerTLS: metadata.DoltServerTLS}
	if request.NormalOptions != expectedOptions || request.NormalOptions.ServerPassword != "" {
		t.Fatal("witness Options differ from exact normal constructor")
	}
	if err = validateOptions(expectedOptions); err != nil {
		t.Fatal(err)
	}
	// Reserve exactly one new0600 regular output before opening the fixture.
	// O_CREATE|O_EXCL refuses existing files and final symlinks. Canonical private
	// parent plus SameFile rechecks bind the descriptor without a portable
	// platform-specific syscall dependency or path-discovery mechanism.
	if _, e := os.Lstat(request.OutputPath + ".failure.json"); !errors.Is(e, os.ErrNotExist) {
		t.Fatal("witness failure sidecar must be absent")
	}
	output, err := os.OpenFile(request.OutputPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		t.Fatal(err)
	}
	outputInfo, infoErr := output.Stat()
	pathInfo, pathErr := os.Lstat(request.OutputPath)
	if infoErr != nil || pathErr != nil || !outputInfo.Mode().IsRegular() || outputInfo.Mode().Perm() != 0600 || !os.SameFile(outputInfo, pathInfo) {
		_ = output.Close()
		t.Fatal("exclusive witness output identity failure")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	witness, captureErr := openInstalledNativeWitness(ctx, expectedOptions, request.Predecessors, request.Require333Absent)
	for directory, before := range physical {
		after, e := os.Lstat(directory)
		actual, realErr := filepath.EvalSymlinks(directory)
		if e != nil || realErr != nil || actual != directory || !os.SameFile(before, after) || before.Mode() != after.Mode() {
			captureErr = errors.Join(captureErr, errors.New("witness directory identity changed"))
		}
	}
	// All stores are closed before postguards and output. Re-read the exact
	// request, custody, normal metadata and marker; no live options are adopted.
	for _, bound := range []struct {
		path, sha string
		limit     int64
	}{{path, expectedSHA, 64 * 1024}, {custodyPath, request.CustodySHA256, 64 * 1024}, {metadataPath, request.MetadataSHA256, 1024 * 1024}, {markerPath, request.MarkerSHA256, 1024}} {
		_, e := installedWitnessReadBoundFile(bound.path, bound.sha, bound.limit)
		captureErr = errors.Join(captureErr, e)
	}
	captureErr = errors.Join(captureErr, ctx.Err())
	envelope := installedWitnessCaptureEnvelope{Protocol: "installed-native-witness-result/v1", Phase: request.Phase, RequestSHA256: expectedSHA, GrantSHA256: request.GrantSHA256, SourceSHA256: request.SourceSHA256, Success: captureErr == nil}
	if captureErr == nil {
		envelope.Witness = &witness
	} else {
		envelope.Error = captureErr.Error()
	}
	publicationErr := installedWitnessWriteCapture(output, envelope)
	// Sticky deadline through durable publication: rewrite only our still-open
	// exclusively created descriptor to false if it crossed while fsyncing.
	if late := ctx.Err(); late != nil {
		captureErr = errors.Join(captureErr, late)
		envelope.Success = false
		envelope.Witness = nil
		envelope.Error = captureErr.Error()
		publicationErr = errors.Join(publicationErr, installedWitnessWriteCapture(output, envelope))
	}
	publicationErr = errors.Join(publicationErr, output.Close())
	directory, dirErr := os.Open(request.OwnedRoot)
	if dirErr == nil {
		dirErr = errors.Join(directory.Sync(), directory.Close())
	}
	publicationErr = errors.Join(publicationErr, dirErr)
	// A JSON success field alone is never authority: require the selected test's
	// natural PASS, unchanged external physical guards and controller custody.
	// Any write/fsync/close/deadline failure preserves the partial output and
	// FAILs this root; there is no retry, rollback or later capture substitution.
	if err = errors.Join(captureErr, publicationErr, ctx.Err()); err != nil {
		// As in the accepted controller finalization pattern, a late failure cannot
		// leave the main JSON alone looking authoritative. This exact derived
		// sidecar is separately exclusive0600; failure preserves all partials.
		envelope.Success = false
		envelope.Witness = nil
		envelope.Error = err.Error()
		sideErr := errors.New("witness owned root drift prevents failure-sidecar publication")
		rootNow, rootErr := os.Lstat(request.OwnedRoot)
		rootReal, rootRealErr := filepath.EvalSymlinks(request.OwnedRoot)
		if rootErr == nil && rootRealErr == nil && rootReal == request.OwnedRoot && os.SameFile(physical[request.OwnedRoot], rootNow) && rootNow.Mode() == physical[request.OwnedRoot].Mode() {
			var failed *os.File
			failed, sideErr = os.OpenFile(request.OutputPath+".failure.json", os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
			if sideErr == nil {
				sideErr = errors.Join(installedWitnessWriteCapture(failed, envelope), failed.Close())
			}
		}
		directory, syncErr := os.Open(request.OwnedRoot)
		if syncErr == nil {
			syncErr = errors.Join(directory.Sync(), directory.Close())
		}
		t.Fatal(errors.Join(err, sideErr, syncErr))
	}
}

func installedWitnessWriteCapture(output *os.File, envelope installedWitnessCaptureEnvelope) error {
	raw, err := json.Marshal(envelope)
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	if _, err = output.Seek(0, io.SeekStart); err != nil {
		return err
	}
	if err = output.Truncate(0); err != nil {
		return err
	}
	n, err := output.Write(raw)
	if err == nil && n != len(raw) {
		err = io.ErrShortWrite
	}
	return errors.Join(err, output.Sync())
}

func installedWitnessInside(root, path string) bool {
	if !filepath.IsAbs(path) || filepath.Clean(path) != path {
		return false
	}
	relative, err := filepath.Rel(root, path)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

func installedWitnessSHA(value string) bool {
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == sha256.Size && value == strings.ToLower(value)
}

func installedWitnessReadBoundFile(path, expectedSHA string, limit int64) ([]byte, error) {
	if !installedWitnessSHA(expectedSHA) || !filepath.IsAbs(path) || filepath.Clean(path) != path {
		return nil, errors.New("invalid witness file binding")
	}
	actual, err := filepath.EvalSymlinks(path)
	if err != nil || actual != path {
		return nil, errors.New("witness input realpath refusal")
	}
	before, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !before.Mode().IsRegular() || before.Mode().Perm() != 0600 || before.Size() > limit {
		return nil, errors.New("witness input mode/size refusal")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	opened, err := file.Stat()
	if err != nil || !os.SameFile(before, opened) {
		_ = file.Close()
		return nil, errors.New("witness input descriptor mismatch")
	}
	raw, readErr := io.ReadAll(io.LimitReader(file, limit+1))
	after, statErr := file.Stat()
	closeErr := file.Close()
	if err = errors.Join(readErr, statErr, closeErr); err != nil {
		return nil, err
	}
	if int64(len(raw)) > limit || !os.SameFile(before, after) || before.Mode() != after.Mode() || before.Size() != after.Size() {
		return nil, errors.New("witness input changed")
	}
	digest := sha256.Sum256(raw)
	if hex.EncodeToString(digest[:]) != expectedSHA {
		return nil, errors.New("witness input digest mismatch")
	}
	return raw, nil
}

func installedWitnessDecodeRequest(raw []byte, value any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return errors.New("invalid witness request/custody JSON")
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return errors.New("witness request/custody trailing JSON")
	}
	return nil
}
