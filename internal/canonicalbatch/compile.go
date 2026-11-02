package canonicalbatch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/steveyegge/beads/graphops"
)

// Compile freezes the complete input before returning any executable members.
// Version 2 is an explicit experimental encoder. It never falls back to the
// legacy runner or reinterprets retained version-1 input.
func Compile(raw []byte, opts Options) (*Batch, error) {
	if len(raw) > MaxInputBytes {
		return nil, fmt.Errorf("%w: bytes", ErrLimit)
	}
	format := opts.Format
	if format == "" {
		format = "jsonl"
	}
	var records []json.RawMessage
	native := format == "json"
	if native {
		if opts.Version != "" || len(opts.Args) != 0 {
			return nil, invalid("native JSON refuses version overrides and arguments")
		}
		canonical, err := canonicalObject(raw)
		if err != nil {
			return nil, err
		}
		var fields map[string]json.RawMessage
		if err = json.Unmarshal(canonical, &fields); err != nil {
			return nil, invalid("request object")
		}
		version, err := stringField(fields, "capabilityVersion", true)
		if err != nil {
			return nil, err
		}
		if version != Version {
			return nil, fmt.Errorf("%w: version %q", ErrUnsupported, version)
		}
		if err = closed(fields, "capabilityVersion", "operations"); err != nil {
			return nil, err
		}
		operations, ok := fields["operations"]
		if !ok || len(operations) == 0 || operations[0] != '[' {
			return nil, invalid("operations must be an array")
		}
		if err = json.Unmarshal(operations, &records); err != nil {
			return nil, invalid("operations array")
		}
	} else if format == "jsonl" {
		if opts.Version != "" && opts.Version != Version {
			return nil, fmt.Errorf("%w: version %q", ErrUnsupported, opts.Version)
		}
		for i, line := range bytes.Split(raw, []byte{'\n'}) {
			if len(bytes.TrimSpace(line)) == 0 {
				continue
			}
			canonical, err := canonicalObject(line)
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", i+1, err)
			}
			records = append(records, canonical)
			if len(records) > MaxOperations {
				return nil, fmt.Errorf("%w: members", ErrLimit)
			}
		}
	} else {
		return nil, invalid("input format must be jsonl or json")
	}
	if len(records) == 0 {
		return nil, invalid("at least one operation is required")
	}
	if len(records) > MaxOperations {
		return nil, fmt.Errorf("%w: members", ErrLimit)
	}
	members := make([]Member, len(records))
	definitions := make(map[string]int)
	for i, record := range records {
		m, err := parseMember(record, opts.Scope, !native)
		if err != nil {
			return nil, fmt.Errorf("operation %d: %w", i, err)
		}
		if m.Name != "" {
			if _, exists := definitions[m.Name]; exists {
				return nil, invalid("duplicate definition %q", m.Name)
			}
			definitions[m.Name] = i
		}
		members[i] = m
	}
	args, err := parseArgs(opts.Args, opts.Scope)
	if err != nil {
		return nil, err
	}
	for name := range args {
		if _, exists := definitions[name]; exists {
			return nil, invalid("argument %q overrides an internal definition", name)
		}
	}
	used := make(map[string]bool)
	// Infer each selector's output kind in member order after structural parsing.
	for i := range members {
		m := &members[i]
		resolve := func(operand *Operand, want graphops.ResourceKind, allowPin bool) error {
			if !operand.Present() {
				return nil
			}
			if !allowPin && operand.Revision != "" {
				return invalid("pinned mutation target is not admitted")
			}
			if operand.Binding == "" {
				if !allowPin && opts.Scope != "" {
					return requireLocal(operand.URI, opts.Scope)
				}
				return nil
			}
			name := operand.Binding
			if def, exists := definitions[name]; exists {
				if def >= i {
					return invalid("binding %q must refer to an earlier member", name)
				}
				if members[def].Kind != want {
					return invalid("binding %q has kind %s, expected %s", name, members[def].Kind, want)
				}
				return nil
			}
			if native {
				return invalid("unknown internal binding %q", name)
			}
			value, exists := args[name]
			if !exists {
				return invalid("missing argument %q", name)
			}
			if !allowPin && value.Revision != "" {
				return invalid("argument %q pins a mutation target", name)
			}
			*operand = Operand{URI: value.URI, Revision: value.Revision}
			if !allowPin && opts.Scope != "" {
				if localErr := requireLocal(operand.URI, opts.Scope); localErr != nil {
					return localErr
				}
			}
			used[name] = true
			return nil
		}
		for _, entry := range []struct {
			op   *Operand
			kind graphops.ResourceKind
			pin  bool
		}{{&m.Bead, graphops.KindBead, false}, {&m.Link, graphops.KindLink, false}, {&m.Source, graphops.KindBead, true}, {&m.Target, graphops.KindBead, true}} {
			if err = resolve(entry.op, entry.kind, entry.pin); err != nil {
				return nil, fmt.Errorf("operation %d: %w", i, err)
			}
		}
		if m.Operation == "select" {
			m.Query, m.Kind, err = resolveQuery(m.Query, opts.Scope, !native, resolve)
			if err != nil {
				return nil, fmt.Errorf("operation %d: %w", i, err)
			}
		}
	}
	for name := range args {
		if !used[name] {
			return nil, invalid("unused argument %q", name)
		}
	}
	encoded := make([]json.RawMessage, len(members))
	for i, m := range members {
		encoded[i], err = marshalMember(m)
		if err != nil {
			return nil, err
		}
	}
	request, err := json.Marshal(struct {
		Version    string            `json:"capabilityVersion"`
		Operations []json.RawMessage `json:"operations"`
	}{Version, encoded})
	if err != nil {
		return nil, err
	}
	canonical, err := graphops.CanonicalizeJSON(request)
	if err != nil {
		return nil, invalid("compiled request: %v", err)
	}
	if len(canonical) > MaxInputBytes {
		return nil, fmt.Errorf("%w: compiled bytes", ErrLimit)
	}
	return &Batch{native: canonical, members: members}, nil
}

// EmitMember emits one canonical JSONL record, validating ordinary member syntax
// without resolving batch bindings. It is useful to CRUD dry-run adapters; the
// eventual Compile call establishes definitions, arguments and ordering.
func EmitMember(m Member) ([]byte, error) {
	raw, err := marshalMember(m)
	if err != nil {
		return nil, err
	}
	parsed, err := parseMember(raw, "", false)
	if err != nil {
		return nil, err
	}
	raw, err = marshalMember(parsed)
	if err != nil {
		return nil, err
	}
	canonical, err := graphops.CanonicalizeJSON(raw)
	if err != nil {
		return nil, invalid("member: %v", err)
	}
	return append(canonical, '\n'), nil
}

func marshalMember(m Member) ([]byte, error) {
	fields := map[string]any{"operation": m.Operation}
	if m.Name != "" {
		fields["name"] = m.Name
	}
	if m.ID != "" {
		fields["id"] = m.ID
	}
	if m.Type.URI != "" {
		fields["type"] = m.Type
	}
	for _, entry := range []struct {
		name  string
		value Operand
	}{{"bead", m.Bead}, {"link", m.Link}, {"source", m.Source}, {"target", m.Target}} {
		if entry.value.Present() {
			fields[entry.name] = entry.value
		}
	}
	for _, entry := range []struct {
		name  string
		value json.RawMessage
	}{{"properties", m.Properties}, {"metadata", m.Metadata}, {"propertiesChange", m.PropertiesChange}, {"metadataChange", m.MetadataChange}, {"revisionProvenance", m.RevisionProvenance}, {"query", m.Query}} {
		if entry.value != nil {
			fields[entry.name] = entry.value
		}
	}
	if m.ExpectedRevision != "" {
		fields["expectedRevision"] = m.ExpectedRevision
	}
	return json.Marshal(fields)
}
func invalid(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalid, fmt.Sprintf(format, args...))
}
func canonicalObject(raw []byte) ([]byte, error) {
	canonical, err := graphops.CanonicalizeJSON(raw)
	if err != nil {
		return nil, invalid("JSON: %v", err)
	}
	if len(canonical) == 0 || canonical[0] != '{' {
		return nil, invalid("expected one JSON object")
	}
	return canonical, nil
}
func closed(fields map[string]json.RawMessage, allowed ...string) error {
	for key := range fields {
		found := false
		for _, name := range allowed {
			if key == name {
				found = true
				break
			}
		}
		if !found {
			return invalid("unknown member %q", key)
		}
	}
	return nil
}
func stringField(fields map[string]json.RawMessage, name string, required bool) (string, error) {
	raw, exists := fields[name]
	if !exists {
		if required {
			return "", invalid("missing %s", name)
		}
		return "", nil
	}
	var value string
	if json.Unmarshal(raw, &value) != nil || value == "" {
		return "", invalid("%s must be a nonempty string", name)
	}
	return value, nil
}
func validName(name string) bool {
	if name == "" || !asciiLetter(name[0]) {
		return false
	}
	for i := 1; i < len(name); i++ {
		c := name[i]
		if !asciiLetter(c) && (c < '0' || c > '9') && c != '_' && c != '-' {
			return false
		}
	}
	return true
}
func asciiLetter(c byte) bool { return c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' }
func parseArgs(values []string, scope string) (map[string]Reference, error) {
	args := make(map[string]Reference, len(values))
	for _, arg := range values {
		name, value, ok := strings.Cut(arg, "=")
		if !ok || !validName(name) {
			return nil, invalid("argument must be NAME=REFERENCE")
		}
		if _, exists := args[name]; exists {
			return nil, invalid("duplicate argument %q", name)
		}
		ref, err := ParseReference(value, scope)
		if err != nil {
			return nil, fmt.Errorf("argument %q: %w", name, err)
		}
		args[name] = ref
	}
	return args, nil
}
