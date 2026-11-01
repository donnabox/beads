// Package resourceget carries checked native Resource documents at the read
// boundary. It does not acquire Resources or establish their authorization.
package resourceget

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/steveyegge/beads/graphops"
)

// DefinitionKind identifies the actual Resource kind, independently of its ID.
type DefinitionKind string

const (
	DefinitionBead DefinitionKind = "bead"
	DefinitionLink DefinitionKind = "link"
)

// DefinitionDocument is an immutable complete local definition Resource. Its
// constructor checks the selected v0.1 carrier shape only. The native getter
// must establish actual definition category, completeness, authorization and
// the checked persisted interpretation context before constructing this value.
// This is neither the legacy v0 Resource codec nor an exact-lookup wrapper.
type DefinitionDocument struct {
	scope    string
	kind     DefinitionKind
	id       string
	revision string
	native   []byte
	origin   *DefinitionOrigin
}

// DefinitionOrigin carries display provenance supplied by the trusted native
// reader after checking its persisted exact local/global correspondence. It is
// not part of the Resource representation and grants no authority or admission.
type DefinitionOrigin struct {
	URI      string
	Revision string
}

// NewDefinitionDocument admits the selected complete definition carrier. Exact
// Type affiliations and whole endpoint References are retained without alias
// resolution, current-definition substitution or global/local ID rewriting.
func NewDefinitionDocument(scope string, kind DefinitionKind, native []byte) (DefinitionDocument, error) {
	if err := graphops.ValidateScopeURL(scope); err != nil {
		return DefinitionDocument{}, fmt.Errorf("definition Scope: %w", err)
	}
	if kind != DefinitionBead && kind != DefinitionLink {
		return DefinitionDocument{}, fmt.Errorf("invalid definition Resource kind %q", kind)
	}
	canonical, err := graphops.CanonicalizeJSON(native)
	if err != nil {
		return DefinitionDocument{}, fmt.Errorf("definition JSON: %w", err)
	}
	fields, err := definitionObject(canonical)
	if err != nil {
		return DefinitionDocument{}, err
	}
	if err := validateDefinitionRecord(scope, kind, fields); err != nil {
		return DefinitionDocument{}, err
	}
	id, _ := definitionString(fields["id"])
	revision, _ := definitionString(fields["revision"])
	return DefinitionDocument{scope: scope, kind: kind, id: id, revision: revision, native: bytes.Clone(canonical)}, nil
}

func (d DefinitionDocument) ID() string           { return d.id }
func (d DefinitionDocument) Revision() string     { return d.revision }
func (d DefinitionDocument) ScopeURL() string     { return d.scope }
func (d DefinitionDocument) Kind() DefinitionKind { return d.kind }

// WithOrigin returns a new document with the reader's checked provenance. The
// caller must establish correspondence; URL syntax alone cannot establish it.
// A document's immutable local identity, revision and bytes remain unchanged.
func (d DefinitionDocument) WithOrigin(uri, revision string) (DefinitionDocument, error) {
	if len(d.native) == 0 {
		return DefinitionDocument{}, fmt.Errorf("uninitialized definition document")
	}
	if err := graphops.ValidateReferenceURI(uri); err != nil {
		return DefinitionDocument{}, fmt.Errorf("definition origin: %w", err)
	}
	if revision == "" || !utf8.ValidString(revision) {
		return DefinitionDocument{}, fmt.Errorf("definition origin requires an exact revision")
	}
	if d.origin != nil && (d.origin.URI != uri || d.origin.Revision != revision) {
		return DefinitionDocument{}, fmt.Errorf("definition origin cannot be replaced")
	}
	d.origin = &DefinitionOrigin{URI: uri, Revision: revision}
	return d, nil
}

// Origin returns an independent provenance value, never a mutable carrier.
func (d DefinitionDocument) Origin() (DefinitionOrigin, bool) {
	if d.origin == nil {
		return DefinitionOrigin{}, false
	}
	return *d.origin, true
}

// NativeJSON returns an independent copy of the complete native representation.
func (d DefinitionDocument) NativeJSON() []byte { return bytes.Clone(d.native) }

// MarshalJSON preserves the selected Resource shape without a second envelope.
func (d DefinitionDocument) MarshalJSON() ([]byte, error) {
	if len(d.native) == 0 {
		return nil, fmt.Errorf("uninitialized definition document")
	}
	return d.NativeJSON(), nil
}

func validateDefinitionRecord(scope string, kind DefinitionKind, fields map[string]json.RawMessage) error {
	allowed := []string{"id", "type", "revision", "properties", "metadata", "revisionProvenance"}
	if kind == DefinitionBead {
		allowed = append(allowed, "ownedLinks")
	} else {
		allowed = append(allowed, "source", "target")
	}
	if err := definitionClosed(fields, allowed...); err != nil {
		return err
	}
	id, err := definitionString(fields["id"])
	if err != nil {
		return fmt.Errorf("definition id: %w", err)
	}
	if _, err := graphops.ScopeLocalPublicResourceID(scope, id); err != nil {
		return fmt.Errorf("definition id: %w", err)
	}
	if _, err := definitionString(fields["revision"]); err != nil {
		return fmt.Errorf("definition revision: %w", err)
	}
	if _, err := definitionReference(fields["type"], true, true); err != nil {
		return fmt.Errorf("definition Type affiliation: %w", err)
	}
	for _, name := range []string{"properties", "metadata"} {
		raw, present := fields[name]
		// The selected Core record requires properties; metadata is optional.
		// Preserve its absence rather than inventing an empty metadata object.
		if name == "metadata" && !present {
			continue
		}
		if _, err := definitionObject(raw); err != nil {
			return fmt.Errorf("definition %s: %w", name, err)
		}
	}
	if raw, present := fields["revisionProvenance"]; present {
		if err := validateDefinitionProvenance(raw); err != nil {
			return err
		}
	}
	if kind == DefinitionLink {
		localEndpoint := false
		for _, name := range []string{"source", "target"} {
			uri, err := definitionReference(fields[name], false, false)
			if err != nil {
				return fmt.Errorf("definition Link %s: %w", name, err)
			}
			if strings.HasPrefix(uri, scope) {
				if _, err := graphops.ScopeLocalPublicResourceID(scope, uri); err != nil {
					return fmt.Errorf("definition Link local %s: %w", name, err)
				}
				localEndpoint = true
			}
		}
		if !localEndpoint {
			return fmt.Errorf("definition Link requires an endpoint in its Scope")
		}
		return nil
	}
	owned, err := definitionObject(fields["ownedLinks"])
	if err != nil {
		return fmt.Errorf("definition owned Links: %w", err)
	}
	seen := make(map[string]bool)
	for bucket, raw := range owned {
		if err := graphops.ValidateTypeURL(bucket); err != nil {
			return fmt.Errorf("owned Link bucket: %w", err)
		}
		var links []json.RawMessage
		if len(raw) == 0 || raw[0] != '[' || json.Unmarshal(raw, &links) != nil {
			return fmt.Errorf("owned Link bucket must contain an array")
		}
		for _, rawLink := range links {
			link, err := definitionObject(rawLink)
			if err != nil {
				return err
			}
			if err := validateDefinitionRecord(scope, DefinitionLink, link); err != nil {
				return fmt.Errorf("owned Link: %w", err)
			}
			linkID, _ := definitionString(link["id"])
			if seen[linkID] {
				return fmt.Errorf("owned Link id %q occurs more than once", linkID)
			}
			seen[linkID] = true
			typeURI, _ := definitionReference(link["type"], true, true)
			if typeURI != bucket {
				return fmt.Errorf("owned Link Type does not match its nominal bucket")
			}
			source, _ := definitionReference(link["source"], false, false)
			if source != id {
				return fmt.Errorf("owned Link source does not identify its containing Bead")
			}
		}
	}
	return nil
}

func definitionObject(raw []byte) (map[string]json.RawMessage, error) {
	var fields map[string]json.RawMessage
	if len(raw) == 0 || raw[0] != '{' || json.Unmarshal(raw, &fields) != nil || fields == nil {
		return nil, fmt.Errorf("expected a JSON object")
	}
	return fields, nil
}

func definitionString(raw []byte) (string, error) {
	var value string
	if len(raw) == 0 || raw[0] != '"' || json.Unmarshal(raw, &value) != nil || value == "" {
		return "", fmt.Errorf("expected a nonempty string")
	}
	return value, nil
}

func definitionClosed(fields map[string]json.RawMessage, allowed ...string) error {
	for name := range fields {
		found := false
		for _, candidate := range allowed {
			if name == candidate {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("unsupported definition member %q", name)
		}
	}
	return nil
}

func definitionReference(raw []byte, requirePin, typeReference bool) (string, error) {
	var uri string
	if len(raw) > 0 && raw[0] == '"' {
		if requirePin {
			return "", fmt.Errorf("definition affiliation requires an exact revision pin")
		}
		var err error
		uri, err = definitionString(raw)
		if err != nil {
			return "", err
		}
	} else {
		fields, err := definitionObject(raw)
		if err != nil {
			return "", err
		}
		if err := definitionClosed(fields, "uri", "revision"); err != nil {
			return "", err
		}
		uri, err = definitionString(fields["uri"])
		if err != nil {
			return "", err
		}
		if _, err := definitionString(fields["revision"]); err != nil {
			return "", err
		}
	}
	if typeReference {
		if err := graphops.ValidateTypeURL(uri); err != nil {
			return "", err
		}
	} else if err := graphops.ValidateReferenceURI(uri); err != nil {
		return "", err
	}
	return uri, nil
}

func validateDefinitionProvenance(raw []byte) error {
	fields, err := definitionObject(raw)
	if err != nil || len(fields) == 0 {
		return fmt.Errorf("revision provenance requires a nonempty object")
	}
	if err := definitionClosed(fields, "actor", "committedAt", "agent", "message"); err != nil {
		return err
	}
	if rawActor, present := fields["actor"]; present {
		actor, err := definitionObject(rawActor)
		if err != nil {
			return err
		}
		if err := definitionClosed(actor, "id", "basis"); err != nil {
			return err
		}
		if _, err := definitionString(actor["id"]); err != nil {
			return err
		}
		basis, err := definitionString(actor["basis"])
		if err != nil || basis != "unknown" && basis != "writer-supplied" {
			return fmt.Errorf("revision actor requires a carried basis")
		}
	}
	context := 0
	for _, name := range []string{"committedAt", "agent", "message"} {
		if rawValue, present := fields[name]; present {
			context++
			if err := validateDefinitionContext(name, rawValue); err != nil {
				return err
			}
		}
	}
	if context != 0 && context != 3 {
		return fmt.Errorf("revision change context must carry all three value states")
	}
	return nil
}

func validateDefinitionContext(name string, raw []byte) error {
	fields, err := definitionObject(raw)
	if err != nil {
		return err
	}
	if err := definitionClosed(fields, "state", "value"); err != nil {
		return err
	}
	state, err := definitionString(fields["state"])
	if err != nil {
		return err
	}
	value, present := fields["value"]
	if state == "undetermined" || state == "absent" && name != "committedAt" {
		if present {
			return fmt.Errorf("nonpresent provenance state cannot carry a value")
		}
		return nil
	}
	if state != "present" || !present {
		return fmt.Errorf("invalid %s provenance value state", name)
	}
	var text string
	if len(value) == 0 || value[0] != '"' || json.Unmarshal(value, &text) != nil || text == "" && name != "message" {
		return fmt.Errorf("invalid %s provenance value", name)
	}
	if name == "committedAt" && !definitionDateTime(text) {
		return fmt.Errorf("invalid committedAt date-time")
	}
	return nil
}

var definitionDateTimePattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})$`)

// Validate native emission, including legitimate UTC leap-second positions.
// Do not normalize stored context or inherit permissive offset/clock coercion.
func definitionDateTime(value string) bool {
	if !definitionDateTimePattern.MatchString(value) || value[11:13] > "23" || value[14:16] > "59" || value[17:19] > "60" {
		return false
	}
	if value[len(value)-1] != 'Z' {
		offset := value[len(value)-5:]
		if offset[:2] > "23" || offset[3:] > "59" {
			return false
		}
	}
	leap := value[17:19] == "60"
	parsedValue := value
	if leap {
		parsedValue = value[:17] + "59" + value[19:]
	}
	parsed, err := time.Parse(time.RFC3339Nano, parsedValue)
	if err != nil {
		return false
	}
	if leap {
		hour, minute, _ := parsed.UTC().Clock()
		return hour == 23 && minute == 59
	}
	return true
}
