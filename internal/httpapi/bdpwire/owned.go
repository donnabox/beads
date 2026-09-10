package bdpwire

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
)

// OwnedWildcardDeclaration is the closed max-only wildcard envelope. It has
// no label because it names no single Link Type.
type OwnedWildcardDeclaration struct {
	Max int `json:"max"`
}

// OwnedOutgoingDeclarations separates the reserved wildcard from explicit
// Link Type declarations. Types must never contain the reserved "*" key.
// A nil TypeDescriptor.OwnsOutgoing omits the optional declaration altogether.
type OwnedOutgoingDeclarations struct {
	Wildcard *OwnedWildcardDeclaration       `json:"*,omitempty"`
	Types    map[string]OwnedLinkDeclaration `json:"-"`
}

var ownedTypeURLPattern = regexp.MustCompile(`^https?://.+`)

// Validate checks declaration bounds, including the cross-entry whole-set
// bound that JSON Schema cannot express. Record grouping remains graph law.
func (d OwnedOutgoingDeclarations) Validate() error {
	if d.Wildcard == nil && len(d.Types) == 0 {
		return fmt.Errorf("bdpwire: ownsOutgoing must not be empty")
	}
	if d.Wildcard != nil && d.Wildcard.Max < 1 {
		return fmt.Errorf("bdpwire: wildcard max must be positive")
	}
	for key, entry := range d.Types {
		// This is the pinned absoluteHttpUrl grammar, not URL normalization.
		if !ownedTypeURLPattern.MatchString(key) {
			return fmt.Errorf("bdpwire: explicit owned Link Type must be an absolute HTTP URL: %q", key)
		}
		if entry.Max < 1 {
			return fmt.Errorf("bdpwire: explicit max must be positive")
		}
		if d.Wildcard != nil && entry.Max > d.Wildcard.Max {
			return fmt.Errorf("bdpwire: explicit max exceeds wildcard max")
		}
	}
	return nil
}

func (d OwnedOutgoingDeclarations) MarshalJSON() ([]byte, error) {
	if err := d.Validate(); err != nil {
		return nil, err
	}
	members := make(map[string]any, len(d.Types)+1)
	for key, entry := range d.Types {
		members[key] = entry
	}
	if d.Wildcard != nil {
		members["*"] = d.Wildcard
	}
	return json.Marshal(members)
}

func (d *OwnedOutgoingDeclarations) UnmarshalJSON(data []byte) error {
	raw, err := oneDocument(data)
	if err != nil {
		return err
	}
	return decodeOwnedOutgoing(raw, d, "ownsOutgoing")
}

func decodeOwnedOutgoing(raw json.RawMessage, target *OwnedOutgoingDeclarations, path string) error {
	if jsonTypeOf(raw) != "object" {
		return typeError(path, "object", raw)
	}
	members, err := splitObject(raw)
	if err != nil {
		return err
	}
	result := OwnedOutgoingDeclarations{Types: map[string]OwnedLinkDeclaration{}}
	for _, m := range members {
		if m.name == "*" {
			var wildcard OwnedWildcardDeclaration
			if err := decodeValue(m.value, reflect.ValueOf(&wildcard).Elem(), path+"/*"); err != nil {
				return err
			}
			result.Wildcard = &wildcard
		} else {
			var entry OwnedLinkDeclaration
			if err := decodeValue(m.value, reflect.ValueOf(&entry).Elem(), path+"/"+m.name); err != nil {
				return err
			}
			// Optional absence is valid; an explicitly present empty label is not.
			fields, err := splitObject(m.value)
			if err != nil {
				return err
			}
			for _, field := range fields {
				if field.name == "label" && entry.Label == "" {
					return fmt.Errorf("bdpwire: explicit label must not be empty")
				}
			}
			result.Types[m.name] = entry
		}
	}
	if err := result.Validate(); err != nil {
		return err
	}
	*target = result
	return nil
}
