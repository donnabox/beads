package graphstore

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/typecatalog"
	"github.com/steveyegge/beads/resourceget"
)

// projectNativeResourceInTx receives an already domain-validated complete
// record. It replaces only the legacy representation's URI-only affiliation
// with its separately retained exact evidence. Native properties remain native
// properties; no Issue is converted to Memory and no omitted schema is filled.
func projectNativeResourceInTx(ctx context.Context, tx *sql.Tx, scope string, kind graph.ResourceKind, domain []byte, pin resourceget.TypeReference, installed checkedInstalledTypes, budget *intrinsicMetadataBudget) (resourceget.ResourceDocument, error) {
	if err := budget.chargeBody(12*len(domain) + 8192); err != nil {
		return resourceget.ResourceDocument{}, err
	}
	var fields map[string]json.RawMessage
	if json.Unmarshal(domain, &fields) != nil || fields == nil {
		return resourceget.ResourceDocument{}, ErrInvalidStore
	}
	var id, nominal, revision, version string
	if json.Unmarshal(fields["id"], &id) != nil || json.Unmarshal(fields["type"], &nominal) != nil || json.Unmarshal(fields["revision"], &revision) != nil || json.Unmarshal(fields["version"], &version) != nil || nominal != pin.URI || revision != version || validateVersionToken(revision) != nil {
		return resourceget.ResourceDocument{}, fmt.Errorf("%w: native Resource identity/affiliation", ErrInvalidStore)
	}
	ref := typecatalog.Reference{URI: pin.URI, Revision: pin.Revision}
	if installed.catalog == nil {
		return resourceget.ResourceDocument{}, ErrInvalidStore
	}
	if err := installed.catalog.ValidateDeclaredType(ref, kind); err != nil {
		return resourceget.ResourceDocument{}, fmt.Errorf("%w: native exact Type: %w", ErrInvalidStore, err)
	}
	definition, ok := installed.catalog.Definition(pin.URI)
	if !ok || definition.Identity != ref {
		return resourceget.ResourceDocument{}, fmt.Errorf("%w: native Type edition is not captured", ErrInvalidStore)
	}
	// NominalOnly is the native domain boundary, not permission to bypass the
	// domain validator that supplied domain. Memory schemas retain all parents.
	if !definition.NominalOnly {
		if err := installed.catalog.ValidateProperties(ref, fields["properties"]); err != nil {
			return resourceget.ResourceDocument{}, fmt.Errorf("%w: native properties/Type correspondence: %w", ErrInvalidStore, err)
		}
	}
	body := map[string]any{"id": id, "type": pin, "revision": revision, "properties": fields["properties"]}
	if metadata, present := fields["metadata"]; present {
		body["metadata"] = metadata
	}
	var attribution Attribution
	if err := json.Unmarshal(fields["attribution"], &attribution); err != nil {
		return resourceget.ResourceDocument{}, fmt.Errorf("%w: native attribution: %w", ErrInvalidStore, err)
	}
	if attribution.Actor != "" {
		basis := "unknown"
		if attribution.Status == "claimed" {
			basis = "writer-supplied"
		} else if attribution.Status != "unknown" {
			return resourceget.ResourceDocument{}, ErrInvalidStore
		}
		body["revisionProvenance"] = map[string]any{"actor": map[string]any{"id": attribution.Actor, "basis": basis}}
	} else if attribution.Status != "unknown" {
		return resourceget.ResourceDocument{}, ErrInvalidStore
	}
	if kind == graph.KindLink {
		body["source"], body["target"] = fields["source"], fields["target"]
	} else if kind == graph.KindBead {
		groups := map[string][]json.RawMessage{}
		eligible := map[string]bool{}
		// Ownership is declared/inherited by the checked exact compiler. Eligible
		// selectors never create public output buckets without actual captures.
		owner := false
		closure, err := installed.catalog.Closure(ref)
		if err != nil {
			return resourceget.ResourceDocument{}, err
		}
		for _, ancestor := range closure {
			declaration, known := installed.catalog.Definition(ancestor.URI)
			if !known || declaration.Identity != ancestor {
				return resourceget.ResourceDocument{}, ErrInvalidStore
			}
			if len(declaration.OwnsOutgoing) != 0 {
				owner = true
			}
		}
		for _, candidate := range installed.facts {
			if candidate.Describes != graph.KindLink {
				continue
			}
			owns, err := installed.catalog.Owns(ref, typecatalog.Reference{URI: candidate.Reference.URI, Revision: candidate.Reference.Revision})
			if err != nil {
				return resourceget.ResourceDocument{}, err
			}
			if owns {
				eligible[candidate.Reference.URI] = true
			}
		}
		var owned []json.RawMessage
		if err := json.Unmarshal(fields["owned"], &owned); err != nil || owned == nil {
			return resourceget.ResourceDocument{}, fmt.Errorf("%w: native owned collection", ErrInvalidStore)
		}
		seen := map[string]bool{}
		for _, captured := range owned {
			var header struct {
				ID       string          `json:"id"`
				Type     string          `json:"type"`
				Revision string          `json:"revision"`
				Source   json.RawMessage `json:"source"`
			}
			if json.Unmarshal(captured, &header) != nil || seen[header.ID] {
				return resourceget.ResourceDocument{}, fmt.Errorf("%w: native captured identity", ErrInvalidStore)
			}
			seen[header.ID] = true
			if !eligible[header.Type] {
				return resourceget.ResourceDocument{}, fmt.Errorf("%w: native capture not owned by exact Type", ErrInvalidStore)
			}
			publicID, err := graph.ScopeLocalPublicResourceID(scope, header.ID)
			if err != nil {
				return resourceget.ResourceDocument{}, err
			}
			// Exact retained ID/revision only. No current Link or current Type can
			// manufacture a pin for a previously captured URI-only Link.
			var path string
			var count int
			if err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM graph_preview_catalog WHERE COALESCE(public_id,path)=? AND resource_kind='link'`, publicID).Scan(&count); err != nil {
				return resourceget.ResourceDocument{}, err
			}
			if count != 1 {
				return resourceget.ResourceDocument{}, fmt.Errorf("%w: ambiguous captured Link identity", ErrInvalidStore)
			}
			if err = tx.QueryRowContext(ctx, `SELECT path FROM graph_preview_catalog WHERE COALESCE(public_id,path)=? AND resource_kind='link'`, publicID).Scan(&path); err != nil {
				return resourceget.ResourceDocument{}, err
			}
			var size uint64
			if err = tx.QueryRowContext(ctx, `SELECT OCTET_LENGTH(snapshot) FROM graph_preview_versions WHERE path=? AND version=?`, path, header.Revision).Scan(&size); err != nil {
				return resourceget.ResourceDocument{}, fmt.Errorf("%w: captured Link edition: %w", ErrInvalidStore, err)
			}
			if size > PreviewCurrentReadByteLimit {
				return resourceget.ResourceDocument{}, ErrLimitExceeded
			}
			if err = budget.chargeBody(4*int(size) + 4096); err != nil {
				return resourceget.ResourceDocument{}, err
			}
			var retained []byte
			if err = tx.QueryRowContext(ctx, `SELECT snapshot FROM graph_preview_versions WHERE path=? AND version=? AND OCTET_LENGTH(snapshot)=?`, path, header.Revision, size).Scan(&retained); err != nil {
				return resourceget.ResourceDocument{}, err
			}
			raw, affiliation, err := decodeNativeSnapshotEdition(retained)
			if err != nil || !bytes.Equal(raw, captured) {
				return resourceget.ResourceDocument{}, fmt.Errorf("%w: native retained capture correspondence: %w", ErrInvalidStore, err)
			}
			document, err := projectNativeResourceInTx(ctx, tx, scope, graph.KindLink, raw, affiliation, installed, budget)
			if err != nil {
				return resourceget.ResourceDocument{}, err
			}
			// Copies are separately charged before the immutable carrier getter.
			if err = budget.chargeBody(4*len(raw) + 4096); err != nil {
				return resourceget.ResourceDocument{}, err
			}
			groups[header.Type] = append(groups[header.Type], document.NativeJSON())
		}
		if owner {
			body["ownedLinks"] = groups
		}
	} else {
		return resourceget.ResourceDocument{}, ErrInvalidStore
	}
	raw, err := canonicalJSON(body)
	if err != nil {
		return resourceget.ResourceDocument{}, err
	}
	return resourceget.NewResourceDocument(scope, kind, raw)
}
