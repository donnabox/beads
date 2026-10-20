package clioutput

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// CollectionDescriptor binds a native collection object to its delivery unit.
// ItemsMember names an existing top-level array; all other native members are
// preserved. The adapter owns membership, ordering, read context and continuation.
// This renderer never fetches, flattens a compound result, or invents a BDP shape.
type CollectionDescriptor struct {
	Item           Descriptor
	ResultKind     string
	ItemsMember    string
	DefaultColumns []string
	// ContextMembers names mandatory native top-level context (for example the
	// subject of a version collection). Human output displays it before rows.
	ContextMembers []string
}

// Completion is the CLI collection completion schema. A nil Exhausted
// is encoded as null: a failed read may not establish exhaustion. Command
// admission must qualify its native collection shape.
type Completion struct {
	Returned   int    `json:"returned"`
	Exhausted  *bool  `json:"exhausted"`
	StopReason string `json:"stop_reason"`
}

// CollectionRenderer applies one immutable item projection to a native collection.
// It is safe to reuse concurrently. JSONL streaming uses NewJSONL instead.
type CollectionRenderer struct {
	item           *Renderer
	kind           string
	itemsMember    string
	format         string
	raw            bool
	columns        []string
	contextMembers []string
}

// NewCollection validates presentation before any provider operation. The default
// is table. Only resource_collection, version_collection and rows are admitted;
// graphs, comparisons, mutations and aggregates have different delivery units.
func NewCollection(descriptor CollectionDescriptor, options Options) (*CollectionRenderer, error) {
	switch descriptor.ResultKind {
	case "resource_collection", "version_collection", "rows":
	default:
		return nil, fmt.Errorf("unsupported collection result kind %q", descriptor.ResultKind)
	}
	if descriptor.ItemsMember == "" {
		return nil, fmt.Errorf("a native item array member is required")
	}
	for _, member := range descriptor.ContextMembers {
		if member == descriptor.ItemsMember {
			return nil, fmt.Errorf("item array cannot be enclosing collection context")
		}
	}
	format, err := collectionFormat(options.Formats, options.JSON)
	if err != nil {
		return nil, err
	}
	if format == "jsonl" && (descriptor.ResultKind != "resource_collection" || len(descriptor.ContextMembers) != 0) {
		return nil, fmt.Errorf("JSONL requires an admitted resource collection stream contract")
	}
	raw := false
	for _, path := range options.Omit {
		raw = raw || path == "envelope"
	}
	if raw && format != "json" && format != "jsonl" {
		return nil, fmt.Errorf("omit envelope requires JSON or JSONL output")
	}
	itemOptions := options
	itemOptions.Formats = []string{"json"}
	itemOptions.JSON = false
	itemOptions.Omit = append(append([]string(nil), options.Omit...), "envelope")
	item, err := New(descriptor.Item, itemOptions)
	if err != nil {
		return nil, err
	}
	columns := descriptor.DefaultColumns
	if len(options.Include) != 0 {
		columns = options.Include
	}
	columns = append([]string(nil), columns...)
	for _, required := range item.protection.Required {
		columns = append(columns, pointerChild("", required))
	}
	for i, column := range columns {
		if _, err := pointerParts(column); err != nil {
			return nil, fmt.Errorf("table column: %w", err)
		}
		columns[i] = protectedColumn(column, item.protection)
	}
	return &CollectionRenderer{item: item, kind: descriptor.ResultKind, itemsMember: descriptor.ItemsMember, format: format, raw: raw, columns: canonicalPaths(columns), contextMembers: canonicalPaths(descriptor.ContextMembers)}, nil
}

func collectionFormat(formats []string, shorthand bool) (string, error) {
	selected := ""
	for _, format := range formats {
		if format == "default" {
			format = "table"
		}
		switch format {
		case "table", "detail", "json", "jsonl":
		default:
			return "", fmt.Errorf("unsupported collection output format %q", format)
		}
		if selected != "" && selected != format {
			return "", fmt.Errorf("conflicting output formats")
		}
		selected = format
	}
	if shorthand {
		if selected != "" && selected != "json" {
			return "", fmt.Errorf("JSON shorthand conflicts with output format")
		}
		selected = "json"
	}
	if selected == "" {
		selected = "table"
	}
	return selected, nil
}

// Render buffers a complete ordinary result. It projects only native items and
// retains other native members exactly as JSON values. Returned must equal the
// delivered item count; projection cannot change that count. No bytes escape on
// any input, projection or serialization failure. The caller handles write errors.
func (r *CollectionRenderer) Render(payload json.RawMessage, completion Completion) ([]byte, error) {
	if r.format == "jsonl" {
		return nil, fmt.Errorf("JSONL requires the streaming writer")
	}
	object, err := collectionObject(payload)
	if err != nil {
		return nil, err
	}
	context := make(map[string]json.RawMessage, len(r.contextMembers))
	for _, member := range r.contextMembers {
		value, exists := object[member]
		if !exists {
			return nil, fmt.Errorf("native collection is missing mandatory context %q", member)
		}
		context[member] = value
	}
	rawItems, ok := object[r.itemsMember]
	if !ok {
		return nil, fmt.Errorf("native collection is missing %q", r.itemsMember)
	}
	if value := bytes.TrimSpace(rawItems); len(value) == 0 || value[0] != '[' {
		return nil, fmt.Errorf("native collection member %q must be an array", r.itemsMember)
	}
	var items []json.RawMessage
	if err := json.Unmarshal(rawItems, &items); err != nil {
		return nil, err
	}
	if err := validateCompletion(completion, len(items), false); err != nil {
		return nil, err
	}
	projected := make([]json.RawMessage, len(items))
	for i, item := range items {
		projected[i], err = r.item.Render(item)
		if err != nil {
			return nil, fmt.Errorf("collection item %d: %w", i, err)
		}
	}
	switch r.format {
	case "table":
		body, err := r.renderTable(items)
		if err != nil {
			return nil, err
		}
		return collectionHumanContext(context, body)
	case "detail":
		var out bytes.Buffer
		for i, item := range projected {
			detail, err := renderDetail(item)
			if err != nil {
				return nil, err
			}
			if i != 0 {
				out.WriteByte('\n')
			}
			out.Write(detail)
		}
		return collectionHumanContext(context, out.Bytes())
	}
	object[r.itemsMember], err = json.Marshal(projected)
	if err != nil {
		return nil, err
	}
	if r.raw {
		return document(object)
	}
	return document(struct {
		SchemaVersion int                        `json:"schema_version"`
		BDPSpec       string                     `json:"bdp_spec"`
		ResultKind    string                     `json:"result_kind"`
		Outcome       string                     `json:"outcome"`
		Data          map[string]json.RawMessage `json:"data"`
		Projection    *projectionMetadata        `json:"projection,omitempty"`
		Completion    Completion                 `json:"completion"`
	}{2, r.item.bdpSpec, r.kind, "success", object, r.item.projection, completion})
}

// RenderFailure emits no data from an unsuccessful buffered read. Consequently
// returned is always zero, even if the provider delivered items before failing.
// Raw and human modes emit nothing; diagnostics/status remain the adapter's job.
func (r *CollectionRenderer) RenderFailure(origin string, problem json.RawMessage) ([]byte, error) {
	wrapper, err := collectionProblem(origin, problem)
	if err != nil {
		return nil, err
	}
	if r.format == "jsonl" {
		return nil, fmt.Errorf("JSONL failures require a terminal record")
	}
	if r.raw || r.format != "json" {
		return nil, nil
	}
	return document(struct {
		SchemaVersion int                 `json:"schema_version"`
		BDPSpec       string              `json:"bdp_spec"`
		ResultKind    string              `json:"result_kind"`
		Outcome       string              `json:"outcome"`
		Projection    *projectionMetadata `json:"projection,omitempty"`
		Completion    Completion          `json:"completion"`
		Error         *problemWrapper     `json:"error"`
	}{2, r.item.bdpSpec, r.kind, "failure", r.item.projection, Completion{StopReason: "error"}, wrapper})
}

func collectionObject(raw json.RawMessage) (map[string]json.RawMessage, error) {
	value := bytes.TrimSpace(raw)
	if !json.Valid(value) || value[0] != '{' {
		return nil, fmt.Errorf("native collection requires a JSON object")
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(value, &object); err != nil {
		return nil, err
	}
	return object, nil
}

func collectionProblem(origin string, raw json.RawMessage) (*problemWrapper, error) {
	if origin != "cli" && origin != "bdp" {
		return nil, fmt.Errorf("problem origin must be bdp or cli")
	}
	if _, err := collectionObject(raw); err != nil {
		return nil, fmt.Errorf("problem must be a JSON object")
	}
	return &problemWrapper{Origin: origin, Problem: append(json.RawMessage(nil), raw...)}, nil
}

func validateCompletion(c Completion, returned int, failed bool) error {
	if c.Returned != returned || returned < 0 {
		return fmt.Errorf("completion returned must match delivered item count")
	}
	if failed {
		if c.StopReason != "error" {
			return fmt.Errorf("failed collection must stop with error")
		}
		return nil
	}
	if c.Exhausted == nil {
		return fmt.Errorf("successful collection requires known exhaustion")
	}
	switch c.StopReason {
	case "exhausted":
		if !*c.Exhausted {
			return fmt.Errorf("exhausted stop requires exhausted collection")
		}
	case "page", "limit":
		if *c.Exhausted {
			return fmt.Errorf("exhausted collection must stop with exhausted reason")
		}
	default:
		return fmt.Errorf("invalid successful collection stop reason %q", c.StopReason)
	}
	return nil
}

func collectionHumanContext(context map[string]json.RawMessage, body []byte) ([]byte, error) {
	if len(context) == 0 {
		return body, nil
	}
	raw, err := json.Marshal(context)
	if err != nil {
		return nil, err
	}
	header, err := renderDetail(raw)
	if err != nil {
		return nil, err
	}
	header = append(header, '\n')
	return append(header, body...), nil
}
