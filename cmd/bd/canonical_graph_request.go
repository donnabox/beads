package main

import (
	"encoding/json"
	"fmt"
	"unicode/utf8"

	"github.com/steveyegge/beads/internal/httpapi/bdpwire"
	"github.com/steveyegge/beads/internal/resourcequery"
)

// These are already-resolved operands, not CLI selectors or another wire type.
// Depth is the existing AST JSON literal (nonnegative integer or "unbounded").
// Direction/depth and optional Type match are explicit: this seam chooses no CLI
// defaults, expands no aliases, resolves no Type, and acquires no Resources.
// Empty roots and duplicate/pinned references retain Decode/Admit's own rules.
type canonicalGraphRequestOptions struct {
	Roots                  []bdpwire.Reference
	Direction              string
	Depth                  json.RawMessage
	LinkTypeURL, TypeMatch string
}

// buildCanonicalGraphRequest lowers only the existing graph-profile subset:
// direct walk over explicit Bead refs, optional Link typeIs through, graph result.
// It never rewrites the older generic graph's induced-Link/partial-cap behavior.
func buildCanonicalGraphRequest(options canonicalGraphRequestOptions) (*resourcequery.Request, error) {
	// encoding/json replaces invalid UTF-8. Refuse that representation loss
	// before serialization rather than silently changing an opaque identity.
	stringsToEncode := []string{options.Direction, options.LinkTypeURL, options.TypeMatch}
	roots := make([]any, 0, len(options.Roots))
	for _, root := range options.Roots {
		stringsToEncode = append(stringsToEncode, root.URI, root.Revision)
		if root.Revision == "" {
			roots = append(roots, root.URI)
		} else {
			// Preserve the exact pin; the current-only admission refuses it.
			roots = append(roots, map[string]string{"uri": root.URI, "revision": root.Revision})
		}
	}
	for _, value := range stringsToEncode {
		if !utf8.ValidString(value) {
			return nil, fmt.Errorf("%w: Graph request strings must be valid UTF-8", resourcequery.ErrInvalid)
		}
	}
	walk := map[string]any{
		"op": "walk", "roots": map[string]any{"op": "refs", "collection": "beads", "references": roots},
		"direction": options.Direction, "depth": options.Depth,
	}
	if options.LinkTypeURL != "" || options.TypeMatch != "" {
		// Do not silently ignore an incomplete Type option pair: its existing
		// typeIs schema/admission validates both values, including empty ones.
		walk["bind"] = "link"
		walk["through"] = map[string]any{
			"op": "typeIs", "of": map[string]string{"op": "binding", "name": "link"},
			"type": options.LinkTypeURL, "match": options.TypeMatch,
		}
	}
	raw, err := json.Marshal(map[string]any{
		"ir": resourcequery.IR, "profile": "graph", "selection": walk,
		"result": map[string]string{"form": "graph"},
	})
	if err != nil {
		return nil, fmt.Errorf("%w: encode Graph request: %v", resourcequery.ErrInvalid, err)
	}
	request, err := resourcequery.Decode(raw)
	if err != nil {
		return nil, err
	}
	if err := resourcequery.Admit(request, resourcequery.DefaultLimits()); err != nil {
		return nil, err
	}
	return request, nil
}
