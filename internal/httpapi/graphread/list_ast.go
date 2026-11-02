package graphread

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/resourcequery"
)

// AdmitResourceList validates the complete public List AST before acquisition.
// Public List retains real editions and its own projection/delivery controls;
// it is distinct from staged Batch's deliberately revisionless List-ID view.
func AdmitResourceList(raw []byte, kind graph.ResourceKind) error {
	request, err := resourcequery.Decode(raw)
	if err != nil {
		return err
	}
	if err := resourcequery.Admit(request, resourcequery.DefaultLimits()); err != nil {
		return err
	}
	if !kind.Valid() || request.Profile() != "list" || request.ResourceKind() != kind {
		return fmt.Errorf("%w: use a List AST scanning this command's Resource kind", resourcequery.ErrInvalid)
	}
	if request.ResultForm() != "resources" {
		return fmt.Errorf("%w: List AST returns resources; use query --ast for an aggregate", resourcequery.ErrUnsupported)
	}
	// The retained List codec still uses attribution/changeContext, while Query
	// maps those fields (including nested owned Links) to revisionProvenance.
	// Refuse only these known mismatches instead of silently testing one shape
	// and returning another, changing a pin, or redefining missing properties.
	for _, root := range request.FieldRoots() {
		switch root {
		case "revisionProvenance", "attribution", "changeContext", "ownedLinks":
			return fmt.Errorf("%w: List AST field %q requires a shared predicate/result representation; use query --ast for the Query representation", resourcequery.ErrUnsupported, root)
		}
	}
	return nil
}

// InventoryMatchingList returns the same checked inventory with only its
// candidate side filtered. It retains endpoint Beads when listing Links, so the
// caller applies Type and endpoint selectors without acquiring another current
// inventory. It opens no provider or remote Type authority during evaluation.
func (r *Reader) InventoryMatchingList(ctx context.Context, raw []byte, kind graph.ResourceKind) (Inventory, error) {
	if err := AdmitResourceList(raw, kind); err != nil {
		return Inventory{}, err
	}
	inventory, data, err := r.queryInventory(ctx)
	if err != nil {
		return Inventory{}, err
	}
	return matchListInventory(ctx, raw, kind, inventory, data)
}

func matchListInventory(ctx context.Context, raw []byte, kind graph.ResourceKind, inventory Inventory, data resourcequery.Dataset) (Inventory, error) {
	response, err := resourcequery.Execute(ctx, raw, data, resourcequery.DefaultLimits())
	if err != nil {
		return Inventory{}, err
	}
	result, ok := response.Result.(resourcequery.ResourcesResult)
	if !ok || !response.Evaluation.Complete || result.ResourceKind != kind || result.Next != nil {
		return Inventory{}, fmt.Errorf("%w: List requires a complete current Resource selection", resourcequery.ErrDataset)
	}
	selected := make(map[resourcequery.Edition]bool, len(result.Items))
	for _, raw := range result.Items {
		var edition resourcequery.Edition
		if err := json.Unmarshal(raw, &edition); err != nil {
			return Inventory{}, fmt.Errorf("%w: List result edition", resourcequery.ErrDataset)
		}
		selected[edition] = true
	}
	matched := 0
	if kind == graph.KindBead {
		beads := inventory.Beads[:0:0]
		for _, bead := range inventory.Beads {
			if selected[resourcequery.Edition{ID: bead.ID, Revision: bead.Revision}] {
				beads = append(beads, bead)
				matched++
			}
		}
		inventory.Beads = beads
	} else {
		links := inventory.Links[:0:0]
		for _, link := range inventory.Links {
			if selected[resourcequery.Edition{ID: link.ID, Revision: link.Revision}] {
				links = append(links, link)
				matched++
			}
		}
		inventory.Links = links
	}
	if matched != len(selected) {
		return Inventory{}, fmt.Errorf("%w: List selection differs from captured inventory", resourcequery.ErrDataset)
	}
	if err := ctx.Err(); err != nil {
		return Inventory{}, err
	}
	return inventory, nil
}

// LegacyListField is one fixed CLI field condition, not an AST operator. The
// compatibility CLI follows direct List's JSON_UNQUOTE(JSON_EXTRACT(...)) text
// comparison. Keep that interpretation out of typed AST equality. Namespace
// and Key address one literal top-level bag member.
type LegacyListField struct {
	Namespace string
	Key       string
	Value     string
	Presence  bool
}

// LegacyListFields is an immutable, bounded compatibility selection. It never
// acquires Resources, follows endpoints, or evaluates caller-supplied syntax.
type LegacyListFields struct{ fields []LegacyListField }

func CompileLegacyListFields(fields []LegacyListField) (*LegacyListFields, error) {
	if len(fields) > 256 {
		return nil, fmt.Errorf("%w: too many List field conditions", resourcequery.ErrLimit)
	}
	size := 0
	for _, field := range fields {
		if (field.Namespace != "properties" && field.Namespace != "metadata") || field.Key == "" {
			return nil, fmt.Errorf("%w: List field namespace or key", resourcequery.ErrInvalid)
		}
		size += len(field.Namespace) + len(field.Key) + len(field.Value)
		if size > 16384 {
			return nil, fmt.Errorf("%w: List field conditions exceed byte budget", resourcequery.ErrLimit)
		}
	}
	return &LegacyListFields{fields: append([]LegacyListField(nil), fields...)}, nil
}

// Select applies all conditions before delivery or counting. Repeated keys
// remain separate conditions; they are never collapsed into a last-wins map.
// Missing members fail equality/presence; an explicitly null member is present.
// Only the candidate's two bags are inspected, including for Link candidates.
func (f *LegacyListFields) Select(ctx context.Context, records []json.RawMessage) ([]json.RawMessage, error) {
	if f == nil {
		return nil, fmt.Errorf("%w: uncompiled List fields", resourcequery.ErrInvalid)
	}
	if len(f.fields) == 0 {
		return records, ctx.Err()
	}
	selected := make([]json.RawMessage, 0, len(records))
	remaining := QueryResourceByteLimit * 4
	for _, record := range records {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		var candidate struct {
			Properties map[string]json.RawMessage `json:"properties"`
			Metadata   map[string]json.RawMessage `json:"metadata"`
		}
		if err := json.Unmarshal(record, &candidate); err != nil {
			return nil, fmt.Errorf("%w: List field candidate JSON: %v", resourcequery.ErrDataset, err)
		}
		textCache := make(map[string]string)
		matches := true
		for _, field := range f.fields {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			bag := candidate.Properties
			if field.Namespace == "metadata" {
				bag = candidate.Metadata
			}
			raw, exists := bag[field.Key]
			if !exists {
				matches = false
				break
			}
			if field.Presence {
				continue
			}
			cacheKey := field.Namespace + "\x00" + field.Key
			text, cached := textCache[cacheKey]
			if !cached {
				var err error
				text, err = legacyListSQLText(ctx, raw, &remaining)
				if err != nil {
					return nil, err
				}
				textCache[cacheKey] = text
			}
			if text != field.Value {
				matches = false
				break
			}
		}
		if matches {
			selected = append(selected, record)
		}
	}
	return selected, ctx.Err()
}

// Donna selected direct List flags as the compatibility authority. This helper
// follows the pinned go-mysql-server JSON formatter + JSON_UNQUOTE behavior:
// length/byte-ordered object keys, spaced separators, fixed decimal numbers,
// null as "null", and decoded strings. JSON_UNQUOTE also decodes strings inside
// containers; writing their decoded contents within quotes directly yields the
// same comparison text. This text need not itself be valid JSON.
//
// Source: go-mysql-server e5eafe0da809 sql/types/json_encode.go and
// sql/expression/function/json/json_unquote.go. No SQL engine is imported here.
func legacyListSQLText(ctx context.Context, raw json.RawMessage, remaining *int) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if len(raw) > *remaining {
		return "", fmt.Errorf("%w: List field comparison byte budget", resourcequery.ErrLimit)
	}
	*remaining -= len(raw)
	canonical, err := graph.CanonicalizeJSON(raw)
	if err != nil {
		return "", fmt.Errorf("%w: List field JSON: %v", resourcequery.ErrDataset, err)
	}
	var value any
	if err = json.Unmarshal(canonical, &value); err != nil {
		return "", fmt.Errorf("%w: List field JSON: %v", resourcequery.ErrDataset, err)
	}
	var text strings.Builder
	write := func(part string) error {
		if len(part) > *remaining {
			return fmt.Errorf("%w: List field comparison byte budget", resourcequery.ErrLimit)
		}
		*remaining -= len(part)
		text.WriteString(part)
		return nil
	}
	var appendValue func(any, bool) error
	appendValue = func(value any, nested bool) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		switch value := value.(type) {
		case nil:
			return write("null")
		case bool:
			return write(strconv.FormatBool(value))
		case float64:
			if value == float64(int64(value)) {
				return write(strconv.FormatInt(int64(value), 10))
			}
			return write(strconv.FormatFloat(value, 'f', -1, 64))
		case string:
			if nested {
				return write("\"" + value + "\"")
			}
			return write(value)
		case []any:
			if err := write("["); err != nil {
				return err
			}
			for i, item := range value {
				if i != 0 {
					if err := write(", "); err != nil {
						return err
					}
				}
				if err := appendValue(item, true); err != nil {
					return err
				}
			}
			return write("]")
		case map[string]any:
			keys := make([]string, 0, len(value))
			for key := range value {
				keys = append(keys, key)
			}
			sort.Slice(keys, func(i, j int) bool {
				if len(keys[i]) != len(keys[j]) {
					return len(keys[i]) < len(keys[j])
				}
				return keys[i] < keys[j]
			})
			if err := write("{"); err != nil {
				return err
			}
			for i, key := range keys {
				if i != 0 {
					if err := write(", "); err != nil {
						return err
					}
				}
				if err := appendValue(key, true); err != nil {
					return err
				}
				if err := write(": "); err != nil {
					return err
				}
				if err := appendValue(value[key], true); err != nil {
					return err
				}
			}
			return write("}")
		default:
			return fmt.Errorf("%w: List field JSON kind", resourcequery.ErrDataset)
		}
	}
	if err := appendValue(value, false); err != nil {
		return "", err
	}
	return text.String(), nil
}
