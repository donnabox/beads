package canonicalbatch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/graphpatch"
)

func parseMember(raw []byte, scope string, author bool) (Member, error) {
	canonical, err := canonicalObject(raw)
	if err != nil {
		return Member{}, err
	}
	var fields map[string]json.RawMessage
	if err = json.Unmarshal(canonical, &fields); err != nil {
		return Member{}, invalid("member object")
	}
	var m Member
	m.Operation, err = stringField(fields, "operation", true)
	if err != nil {
		return m, err
	}
	base := []string{"operation"}
	create, update, link := false, false, false
	switch m.Operation {
	case "createBead":
		create = true
		m.Kind = graphops.KindBead
	case "createLink":
		create = true
		link = true
		m.Kind = graphops.KindLink
	case "updateBead":
		update = true
		m.Kind = graphops.KindBead
	case "updateLink":
		update = true
		link = true
		m.Kind = graphops.KindLink
	case "deleteBead":
		m.Kind = graphops.KindBead
	case "deleteLink":
		link = true
		m.Kind = graphops.KindLink
	case "select":
		if err = closed(fields, "operation", "name", "query"); err != nil {
			return m, err
		}
		m.Name, err = stringField(fields, "name", true)
		if err != nil {
			return m, err
		}
		if !validName(m.Name) {
			return m, invalid("invalid name %q", m.Name)
		}
		m.Query = bytes.Clone(fields["query"])
		if m.Query == nil {
			return m, invalid("missing query")
		}
		if err = validateQuery(m.Query); err != nil {
			return m, err
		}
		m.Set = true
		return m, nil
	default:
		return m, invalid("unknown operation %q", m.Operation)
	}
	if create {
		base = append(base, "name", "id", "type", "properties", "metadata", "revisionProvenance")
	} else {
		if link {
			base = append(base, "link")
		} else {
			base = append(base, "bead")
		}
		base = append(base, "expectedRevision")
		if update {
			base = append(base, "type", "properties", "metadata", "propertiesChange", "metadataChange", "revisionProvenance")
		} else if link {
			base = append(base, "revisionProvenance")
		}
	}
	if link && (create || update) {
		base = append(base, "source", "target")
	}
	if err = closed(fields, base...); err != nil {
		return m, err
	}
	if create {
		m.Name, err = stringField(fields, "name", false)
		if err != nil {
			return m, err
		}
		if m.Name != "" && !validName(m.Name) {
			return m, invalid("invalid name %q", m.Name)
		}
		m.ID, err = stringField(fields, "id", false)
		if err != nil {
			return m, err
		}
		if m.ID != "" {
			if strings.Contains(m.ID, "@") {
				return m, invalid("new IDs reserve @")
			}
			if strings.Contains(m.ID, "://") {
				if _, err = resolveURI(m.ID, scope, false); err == nil && scope != "" {
					err = requireLocal(m.ID, scope)
				}
			} else {
				err = graphops.ValidateCanonicalResourceID(m.ID)
			}
			if err != nil {
				return m, fmt.Errorf("id: %w", err)
			}
		}
	}
	if rawType, exists := fields["type"]; exists {
		operand, parseErr := parseOperand(rawType, scope, author, false)
		if parseErr != nil {
			return m, invalid("type: %v", parseErr)
		}
		m.Type = Reference{URI: operand.URI, Revision: operand.Revision}
	} else if create {
		return m, invalid("missing type")
	}
	for _, entry := range []struct {
		name     string
		dest     *Operand
		required bool
		pin      bool
	}{{"bead", &m.Bead, !create && !link, false}, {"link", &m.Link, !create && link, false}, {"source", &m.Source, create && link, true}, {"target", &m.Target, create && link, true}} {
		value, exists := fields[entry.name]
		if !exists {
			if entry.required {
				return m, invalid("missing %s", entry.name)
			}
			continue
		}
		*entry.dest, err = parseOperand(value, scope, author, true)
		if err != nil {
			return m, invalid("%s: %v", entry.name, err)
		}
		if !entry.pin && entry.dest.Revision != "" {
			return m, invalid("%s must be unpinned", entry.name)
		}
	}
	m.ExpectedRevision, err = stringField(fields, "expectedRevision", false)
	if err != nil {
		return m, err
	}
	for _, entry := range []struct {
		name   string
		dest   *json.RawMessage
		object bool
	}{{"properties", &m.Properties, true}, {"metadata", &m.Metadata, true}, {"propertiesChange", &m.PropertiesChange, false}, {"metadataChange", &m.MetadataChange, false}} {
		value, exists := fields[entry.name]
		if !exists {
			continue
		}
		if entry.object {
			if len(value) == 0 || value[0] != '{' {
				return m, invalid("%s must be an object", entry.name)
			}
		} else {
			// Empty changes are admitted by ordinaryUpdateChange. They are a no-op,
			// not omission, and must survive lowering and idempotency freezing.
			if !bytes.Equal(value, []byte("[]")) {
				if _, err = graphpatch.Parse(value); err != nil {
					return m, invalid("%s: %v", entry.name, err)
				}
			}
		}
		*entry.dest = bytes.Clone(value)
	}
	if m.Properties != nil && m.PropertiesChange != nil {
		return m, invalid("properties and propertiesChange are exclusive")
	}
	if m.Metadata != nil && m.MetadataChange != nil {
		return m, invalid("metadata and metadataChange are exclusive")
	}
	if update && m.Type.URI == "" && m.Properties == nil && m.Metadata == nil && m.PropertiesChange == nil && m.MetadataChange == nil && !m.Source.Present() && !m.Target.Present() {
		return m, invalid("update requires at least one edit")
	}
	if value, exists := fields["revisionProvenance"]; exists {
		if err = validateProvenance(value); err != nil {
			return m, err
		}
		m.RevisionProvenance = bytes.Clone(value)
	}
	return m, nil
}

func parseOperand(raw []byte, scope string, author, bindings bool) (Operand, error) {
	var o Operand
	if len(raw) > 0 && raw[0] == '"' {
		if json.Unmarshal(raw, &o.URI) != nil {
			return o, invalid("reference string")
		}
	} else {
		if len(raw) == 0 || raw[0] != '{' {
			return o, invalid("expected Reference or binding")
		}
		var fields map[string]json.RawMessage
		if json.Unmarshal(raw, &fields) != nil {
			return o, invalid("reference object")
		}
		if _, exists := fields["binding"]; exists {
			if !bindings {
				return o, invalid("binding not allowed here")
			}
			if err := closed(fields, "binding"); err != nil {
				return o, err
			}
			var err error
			o.Binding, err = stringField(fields, "binding", true)
			if err != nil {
				return o, err
			}
			if !validName(o.Binding) {
				return o, invalid("invalid binding name")
			}
			return o, nil
		}
		if err := closed(fields, "uri", "revision"); err != nil {
			return o, err
		}
		var err error
		o.URI, err = stringField(fields, "uri", true)
		if err != nil {
			return o, err
		}
		o.Revision, err = stringField(fields, "revision", true)
		if err != nil {
			return o, err
		}
	}
	uri, err := resolveURI(o.URI, scope, author)
	if err != nil {
		return o, err
	}
	o.URI = uri
	return o, nil
}

func resolveURI(value, scope string, author bool) (string, error) {
	if value == "" {
		return "", invalid("empty Reference")
	}
	if strings.HasPrefix(value, "batch-ref::") {
		return "", invalid("batch-ref spelling must be lowered to a structural binding")
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return "", invalid("invalid Reference URI")
	}
	if strings.Contains(value, "://") {
		if parsed.Scheme != "http" && parsed.Scheme != "https" {
			return "", invalid("Reference must be HTTP(S)")
		}
		if parsed.Host == "" || parsed.User != nil || parsed.Fragment != "" {
			return "", invalid("invalid absolute Reference")
		}
		return value, nil
	}
	if !author {
		return "", invalid("native Reference must be absolute")
	}
	if scope == "" {
		return "", invalid("relative Reference needs a Scope")
	}
	if err = graphops.ValidateCanonicalResourceID(value); err != nil {
		return "", invalid("invalid local Reference: %v", err)
	}
	base, err := url.Parse(scope)
	if err != nil || base.Host == "" || (base.Scheme != "http" && base.Scheme != "https") || base.User != nil || base.RawQuery != "" || base.Fragment != "" {
		return "", invalid("invalid Scope")
	}
	return strings.TrimSuffix(scope, "/") + "/" + value, nil
}
func validateProvenance(raw []byte) error {
	if len(raw) == 0 || raw[0] != '{' {
		return invalid("revisionProvenance must be an object")
	}
	var f map[string]json.RawMessage
	if json.Unmarshal(raw, &f) != nil {
		return invalid("revisionProvenance object")
	}
	if err := closed(f, "actor", "agent", "message"); err != nil {
		return err
	}
	if actor, exists := f["actor"]; exists {
		if len(actor) == 0 || actor[0] != '{' {
			return invalid("actor must be an object")
		}
		var a map[string]json.RawMessage
		if json.Unmarshal(actor, &a) != nil {
			return invalid("actor object")
		}
		if err := closed(a, "id", "basis"); err != nil {
			return err
		}
		if _, err := stringField(a, "id", true); err != nil {
			return err
		}
		basis, err := stringField(a, "basis", true)
		if err != nil {
			return err
		}
		if basis != "writer-supplied" && basis != "unknown" {
			return invalid("unrecognized actor basis")
		}
	}
	for _, name := range []string{"agent", "message"} {
		if value, exists := f[name]; exists && !bytes.Equal(value, []byte("null")) {
			var s string
			if json.Unmarshal(value, &s) != nil || (name == "agent" && s == "") {
				return invalid("invalid %s", name)
			}
		}
	}
	return nil
}

// ParseOperand is the ordinary command adapter's one structural input parser.
// It validates exact JSON and freezes it before interpreting any fields.
func ParseWireOperand(raw []byte, scope string, allowBinding bool) (Operand, error) {
	canonical, err := graphops.CanonicalizeJSON(raw)
	if err != nil {
		return Operand{}, fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	return parseOperand(canonical, scope, true, allowBinding)
}

// requireLocal compares the exact binding; it neither normalizes another origin
// into this Scope nor infers kind from a path prefix.
func requireLocal(uri, scope string) error {
	prefix := strings.TrimSuffix(scope, "/") + "/"
	if !strings.HasPrefix(uri, prefix) {
		return fmt.Errorf("%w: cross-Scope acquisition or mutation", ErrUnsupported)
	}
	return graphops.ValidateCanonicalResourceID(strings.TrimPrefix(uri, prefix))
}
