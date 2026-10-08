package graphstore

import (
	"encoding/json"
	"fmt"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage"
	publicops "github.com/steveyegge/beads/issueops"
)

// commonMetadata admits one complete JSON object. The empty object is the
// public representation even when the native Issue plane stores no bytes.
func commonMetadata(raw json.RawMessage) (json.RawMessage, error) {
	if len(raw) == 0 {
		return json.RawMessage(`{}`), nil
	}
	canonical, err := graph.CanonicalizeJSON(raw)
	if err != nil {
		return nil, fmt.Errorf("%w: metadata: %v", storage.ErrValidation, err)
	}
	if len(canonical) == 0 || canonical[0] != '{' {
		return nil, fmt.Errorf("%w: metadata must be a JSON object", storage.ErrValidation)
	}
	return canonical, nil
}

func hasCommonMetadataPatch(p publicops.MetadataPatch) bool {
	return p.Replace.Set || p.Merge.Set || len(p.Set) > 0 || len(p.Unset) > 0
}
