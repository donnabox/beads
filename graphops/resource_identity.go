package graphops

import (
	"fmt"
	"strings"
)

// ValidateCanonicalResourceID validates a Scope-local canonical URI path without
// inferring Resource kind from its spelling. It preserves the existing canonical
// segment grammar. It is not a validator for arbitrary legacy Memory keys.
//
// This additive preparation API does not change the pinned rooted readers or
// writers. Their common identity projection must be reconciled separately.
func ValidateCanonicalResourceID(id string) error {
	if strings.ContainsAny(id, "?#") {
		return fmt.Errorf("%w: local Resource ID must not contain query or fragment", ErrValidation)
	}
	for _, segment := range strings.Split(id, "/") {
		if err := ValidateCanonicalSegment(segment); err != nil {
			return fmt.Errorf("%w: local Resource ID %q: %s", ErrValidation, id, reason(err))
		}
	}
	return nil
}

// CanonicalResourceURL composes a canonical Scope and already-canonical local
// Resource ID. It never inserts a kind prefix, cleans a path, or encodes data.
func CanonicalResourceURL(scope, id string) (string, error) {
	if err := ValidateScopeURL(scope); err != nil {
		return "", err
	}
	if err := ValidateCanonicalResourceID(id); err != nil {
		return "", err
	}
	return scope + id, nil
}

// ScopeLocalResourceID is the exact inverse of CanonicalResourceURL for this
// Scope. It does not derive kind from a prefix or percent-decode the public ID.
func ScopeLocalResourceID(scope, uri string) (string, error) {
	if err := ValidateScopeURL(scope); err != nil {
		return "", err
	}
	if !strings.HasPrefix(uri, scope) {
		return "", fmt.Errorf("%w: Resource URL is outside the selected Scope", ErrValidation)
	}
	id := strings.TrimPrefix(uri, scope)
	if err := ValidateCanonicalResourceID(id); err != nil {
		return "", err
	}
	return id, nil
}

// MountCanonicalResourceID applies an explicit canonical ID root. An omitted
// root preserves id exactly. Neither operand is trimmed, normalized or stripped
// of literal beads/, links/ or types/ segments. This validates canonical paths;
// it does not select a disposition for nonrepresentable legacy Memory keys.
func MountCanonicalResourceID(root, id string) (string, error) {
	if err := ValidateCanonicalResourceID(id); err != nil {
		return "", err
	}
	if root == "" {
		return id, nil
	}
	if err := ValidateCanonicalResourceID(root); err != nil {
		return "", err
	}
	return root + "/" + id, nil
}
