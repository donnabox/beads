package typecatalog

import (
	"bytes"
	"encoding/json"
	"net/url"
	"strings"
	"unicode/utf8"

	graph "github.com/steveyegge/beads/graphops"
)

// DefinitionCapture preserves an explicitly supplied complete-definition lookup
// wrapper. It is a private preparatory value, not a registered Type, a wire
// advertisement, an interpreter, or evidence of authority/completeness. Its zero
// value has no captured record. No live catalog or storage is consulted.
type DefinitionCapture struct {
	raw, definition []byte
	reference       Reference
	origin          string
	hasOrigin       bool
}

// CaptureDefinition copies one {definition, originScope?} JSON wrapper. Structural
// checks require exact identity/affiliation pins and the supplied owned-Link plane;
// they do not establish actual category, complete source disclosure, ownership,
// schema admission, Refines semantics, or permission to use the described Type.
// Unknown record/descriptor values and every original JSON lexeme are retained.
func CaptureDefinition(raw []byte) (DefinitionCapture, error) {
	if !utf8.Valid(raw) || !json.Valid(raw) {
		return DefinitionCapture{}, invalid("definition capture needs one UTF-8 JSON document")
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := captureTokens(decoder, raw); err != nil {
		return DefinitionCapture{}, err
	}
	wrapper, err := captureObject(raw)
	if err != nil {
		return DefinitionCapture{}, err
	}
	for name := range wrapper {
		if name != "definition" && name != "originScope" {
			return DefinitionCapture{}, invalid("unknown definition wrapper member %q", name)
		}
	}
	record, err := captureObject(wrapper["definition"])
	if err != nil {
		return DefinitionCapture{}, err
	}
	if _, present := record["links"]; present {
		return DefinitionCapture{}, invalid("captured definition must not carry a links plane")
	}
	ref, err := captureRecord(record)
	if err != nil {
		return DefinitionCapture{}, err
	}
	if _, err := captureObject(record["properties"]); err != nil {
		return DefinitionCapture{}, err
	}
	owned, err := captureObject(record["ownedLinks"])
	if err != nil {
		return DefinitionCapture{}, err
	}
	for name, group := range owned {
		if err := captureHTTPURI(name); err != nil {
			return DefinitionCapture{}, err
		}
		if len(bytes.TrimSpace(group)) == 0 || bytes.TrimSpace(group)[0] != '[' {
			return DefinitionCapture{}, invalid("captured owned group must be an array")
		}
		var links []map[string]json.RawMessage
		if err := json.Unmarshal(group, &links); err != nil {
			return DefinitionCapture{}, invalid("captured owned Link must be an object")
		}
		for _, link := range links {
			if _, err := captureRecord(link); err != nil {
				return DefinitionCapture{}, err
			}
			if _, err := captureObject(link["properties"]); err != nil {
				return DefinitionCapture{}, err
			}
			for _, side := range []string{"source", "target"} {
				if err := captureEndpoint(link[side]); err != nil {
					return DefinitionCapture{}, err
				}
			}
		}
	}
	result := DefinitionCapture{raw: bytes.Clone(raw), definition: bytes.Clone(wrapper["definition"]), reference: ref}
	if origin, present := wrapper["originScope"]; present {
		result.origin, err = captureString(origin)
		if err != nil {
			return DefinitionCapture{}, err
		}
		if err := captureHTTPURI(result.origin); err != nil {
			return DefinitionCapture{}, err
		}
		result.hasOrigin = true
	}
	return result, nil
}

// Reference returns the supplied exact definition identity; it grants no admission.
func (c DefinitionCapture) Reference() Reference { return c.reference }

// Bytes returns an independent copy of the entire original wrapper, or nil for zero.
func (c DefinitionCapture) Bytes() []byte { return bytes.Clone(c.raw) }

// DefinitionBytes returns an independent copy of the original inner record.
func (c DefinitionCapture) DefinitionBytes() []byte { return bytes.Clone(c.definition) }

// OriginScope returns recorded source context, not verified mutation authority.
func (c DefinitionCapture) OriginScope() (string, bool) { return c.origin, c.hasOrigin }

// These extraction helpers run only after captureTokens has rejected duplicate
// decoded keys and Unicode replacement; they do not independently admit JSON.
func captureObject(raw []byte) (map[string]json.RawMessage, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return nil, invalid("captured object is missing or not an object")
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		return nil, invalid("invalid captured object: %v", err)
	}
	return object, nil
}

func captureString(raw []byte) (string, error) {
	var value string
	if err := json.Unmarshal(raw, &value); err != nil || value == "" {
		return "", invalid("captured identity/revision must be a nonempty string")
	}
	return value, nil
}

func captureRecord(record map[string]json.RawMessage) (Reference, error) {
	uri, err := captureString(record["id"])
	if err != nil {
		return Reference{}, err
	}
	revision, err := captureString(record["revision"])
	if err != nil {
		return Reference{}, err
	}
	ref := Reference{URI: uri, Revision: revision}
	if err := captureHTTPURI(ref.URI); err != nil {
		return Reference{}, err
	}
	if err := capturePin(record["type"], true); err != nil {
		return Reference{}, err
	}
	return ref, nil
}

func capturePin(raw []byte, typeURI bool) error {
	object, err := captureObject(raw)
	if err != nil {
		return err
	}
	if len(object) != 2 {
		return invalid("captured pin must have exactly uri and revision")
	}
	uri, err := captureString(object["uri"])
	if err != nil {
		return err
	}
	if _, err := captureString(object["revision"]); err != nil {
		return err
	}
	if typeURI {
		return captureHTTPURI(uri)
	}
	return captureURI(uri)
}

// The old fixed-catalog validReference also imposes canonical URL admission
// (for example, no fragments or credentials). A capture is not that boundary.
// Preserve HTTP(S) URI spellings admitted by the shared structural carrier;
// canonical identity and publication authority remain the owner's checks.
func captureHTTPURI(uri string) error {
	if !(strings.HasPrefix(uri, "http://") && len(uri) > len("http://")) &&
		!(strings.HasPrefix(uri, "https://") && len(uri) > len("https://")) {
		return invalid("captured identity must be an absolute HTTP(S) URI")
	}
	return captureURI(uri)
}

// Parse syntax without applying canonical URL admission or rewriting the source.
// The shared URI format requires more than a scheme/character scan: malformed
// authorities such as https://[ must refuse, including on endpoint references.
func captureURI(uri string) error {
	if err := graph.ValidateReferenceURI(uri); err != nil {
		return err
	}
	// net/url rejects percent escapes in reg-name hosts that RFC 3986 permits.
	// The shared character scan above already checked every escape. Substitute
	// ordinary characters only in this validation copy: escaped delimiters stay
	// non-delimiters, while bracket/authority/port syntax is still parsed. Never
	// decode, normalize, or retain this scratch spelling as the source identity.
	syntax := []byte(uri)
	for i := 0; i < len(syntax); i++ {
		if syntax[i] == '%' {
			syntax[i], syntax[i+1], syntax[i+2] = 'x', 'x', 'x'
			i += 2
		}
	}
	if _, err := url.Parse(string(syntax)); err != nil {
		return invalid("captured URI syntax: %v", err)
	}
	return nil
}

func captureEndpoint(raw []byte) error {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) != 0 && trimmed[0] == '"' {
		uri, err := captureString(raw)
		if err != nil {
			return err
		}
		return captureURI(uri)
	}
	return capturePin(raw, false)
}

// Token validation refuses duplicate decoded keys and encoding/json's surrogate
// replacement. Only string tokens pass through the existing graph JSON checker:
// arbitrary numbers stay json.Number tokens and never encounter binary64 rules.
// json.Valid already checked the complete document and standard nesting bound.
func captureTokens(decoder *json.Decoder, raw []byte) error {
	start := decoder.InputOffset()
	token, err := decoder.Token()
	if err != nil {
		return invalid("captured JSON token: %v", err)
	}
	if _, ok := token.(string); ok {
		text := bytes.TrimLeft(raw[start:decoder.InputOffset()], " \t\r\n,:")
		if _, err := graph.CanonicalizeJSON(text); err != nil {
			return invalid("captured string at byte %d: %v", start, err)
		}
		return nil
	}
	delim, composite := token.(json.Delim)
	if !composite {
		return nil
	}
	seen := map[string]bool{}
	for decoder.More() {
		if delim == '{' {
			keyStart := decoder.InputOffset()
			key, err := decoder.Token()
			if err != nil {
				return invalid("captured member: %v", err)
			}
			name, ok := key.(string)
			if !ok {
				return invalid("captured member name must be a string")
			}
			if seen[name] {
				return invalid("duplicate captured member %q", name)
			}
			seen[name] = true
			text := bytes.TrimLeft(raw[keyStart:decoder.InputOffset()], " \t\r\n,:")
			if _, err := graph.CanonicalizeJSON(text); err != nil {
				return invalid("captured member %q at byte %d: %v", name, keyStart, err)
			}
		}
		if err := captureTokens(decoder, raw); err != nil {
			return err
		}
	}
	if _, err := decoder.Token(); err != nil {
		return invalid("captured closing delimiter: %v", err)
	}
	return nil
}
