package memoryapi_test

import (
	"errors"
	"testing"

	"github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/memoryapi"
)

// This pins the two actual domains without selecting how a future import must
// handle their difference. Neither a codec nor a new key restriction is enabled.
func TestLegacyKeyAndCanonicalIDDomainsAreDistinct(t *testing.T) {
	for _, key := range []string{"..", "a//b", "\x00"} {
		got, err := memoryapi.ValidateKey(key)
		if err != nil || got != key {
			t.Errorf("legacy key changed: %q -> %q %v", key, got, err)
			continue
		}
		if err := graphops.ValidateCanonicalResourceID(got); !errors.Is(err, graphops.ErrValidation) {
			t.Errorf("legacy key %q did not expose canonical grammar boundary: %v", key, err)
		}
	}
}
