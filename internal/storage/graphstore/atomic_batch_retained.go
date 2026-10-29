package graphstore

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/canonicalbatch"
)

// validateAtomicBatchRetained checks only the frozen local result carrier. It
// deliberately does not re-admit historical data against today's graph, Type
// catalog, configuration, actor, or ownership policy. Properties stay raw so
// native optional fields and user JSON are not normalized by a second schema.
// Historical Type/endpoint References remain syntactic carriers; only subject
// identities and owner result mappings must belong to the frozen local Scope.
func validateAtomicBatchRetained(raw []byte, scope string, members []canonicalbatch.Member) (BatchResult, error) {
	invalid := func(reason string) (BatchResult, error) {
		return BatchResult{}, fmt.Errorf("%w: retained batch %s", ErrInvalidStore, reason)
	}
	if len(raw) == 0 || len(raw) > AtomicBatchResultLimit || len(members) == 0 || len(members) > canonicalbatch.MaxOperations {
		return invalid("size or member count")
	}
	canonical, err := graph.CanonicalizeJSON(raw)
	if err != nil || !bytes.Equal(canonical, raw) {
		return invalid("noncanonical result")
	}
	var result BatchResult
	if json.Unmarshal(raw, &result) != nil || !sameVersionJSON(raw, result) || result.CapabilityVersion != canonicalbatch.Version || len(result.Results) != len(members) || result.Resources == nil {
		return invalid("closed result shape")
	}
	if result.Changed != (result.Transaction != "") || result.Changed && !authorityID.MatchString(result.Transaction) {
		return invalid("transaction/change disagreement")
	}
	setMembers := atomicBatchSetMembers(members)
	mutations := map[string]graph.ResourceKind{}
	created := map[string]bool{}
	lastOperation := map[string]string{}
	bindings := map[string][]string{}
	// Scalar IDs select only the historical carrier shape, never weaker validation.
	scalar := result.Results[0].ID != ""
	for i, row := range result.Results {
		member := members[i]
		if row.OperationIndex != i || row.Operation != member.Operation || (row.ID != "") != scalar {
			return invalid("operation identity or mixed carriers")
		}
		ids := row.IDs
		if scalar {
			if row.IDs != nil || member.Operation == "select" || setMembers[i] {
				return invalid("legacy scalar mapping")
			}
			ids = []string{row.ID}
		} else if ids == nil || member.Operation != "select" && !setMembers[i] && len(ids) != 1 {
			return invalid("identity array cardinality")
		}
		for j, id := range ids {
			if !retainedBatchLocal(scope, id) || j > 0 && graph.CompareCodeUnits(ids[j-1], id) >= 0 {
				return invalid("identity array ordering or Scope")
			}
		}
		if member.Operation != "select" {
			kind := graph.KindBead
			switch member.Operation {
			case "createBead", "updateBead", "deleteBead":
			case "createLink", "updateLink", "deleteLink":
				kind = graph.KindLink
			default:
				return invalid("unknown mutation")
			}
			operand := member.Bead
			if kind == graph.KindLink {
				operand = member.Link
			}
			if operand.Binding != "" {
				bound, exists := bindings[operand.Binding]
				if !exists || !retainedBatchEqualIDs(ids, bound) {
					return invalid("bound mutation identities")
				}
			} else {
				expected := operand.URI
				if strings.HasPrefix(member.Operation, "create") && member.ID != "" {
					expected = member.ID
					if graph.ValidateCanonicalResourceID(expected) == nil {
						expected, err = graph.CanonicalPublicResourceURL(scope, expected)
						if err != nil {
							return invalid("explicit create identity")
						}
					}
				}
				if expected != "" && (len(ids) != 1 || ids[0] != expected) {
					return invalid("literal mutation identity")
				}
			}
			for _, id := range ids {
				if previous := mutations[id]; previous != "" && previous != kind {
					return invalid("mutation identity kind")
				}
				if strings.HasPrefix(member.Operation, "create") {
					if mutations[id] != "" {
						return invalid("repeated creation identity")
					}
					created[id] = true
				}
				mutations[id] = kind
				lastOperation[id] = member.Operation
			}
		}
		if member.Name != "" {
			bindings[member.Name] = ids
		}
	}

	outcomes := map[string]BatchResourceOutcome{}
	owners := map[string]bool{}
	changed := false
	for i, row := range result.Resources {
		if !retainedBatchLocal(scope, row.ID) || !row.ResourceKind.Valid() || i > 0 && graph.CompareCodeUnits(result.Resources[i-1].ID, row.ID) >= 0 {
			return invalid("outcome identity, kind or order")
		}
		if kind := mutations[row.ID]; kind != "" && kind != row.ResourceKind {
			return invalid("outcome/mutation kind")
		}
		last := lastOperation[row.ID]
		switch {
		case last == "":
			if row.Outcome != "updated" && row.Outcome != "unchanged" {
				return invalid("owner-only outcome variant")
			}
		case strings.HasPrefix(last, "delete"):
			if created[row.ID] && row.Outcome != "allocated" || !created[row.ID] && row.Outcome != "deleted" {
				return invalid("deleted mutation outcome")
			}
		case created[row.ID]:
			if row.Outcome != "created" {
				return invalid("creation outcome")
			}
		default:
			if row.Outcome != "updated" && row.Outcome != "unchanged" {
				return invalid("existing mutation outcome")
			}
		}
		switch row.Outcome {
		case "created", "updated", "unchanged":
			if row.Deleted != nil || len(row.Resource) == 0 || row.Changed != (row.Outcome != "unchanged") {
				return invalid("live outcome variant")
			}
			if err := validateRetainedBatchEnvelope(row.Resource, scope, row.ID, row.ResourceKind); err != nil {
				return invalid("live envelope: " + err.Error())
			}
		case "deleted":
			if !row.Changed || len(row.Resource) != 0 || row.Deleted == nil || row.Deleted.ID != row.ID || row.Deleted.ResourceKind != row.ResourceKind || graph.ValidateReferenceURI(row.Deleted.Type) != nil || !authorityID.MatchString(row.Deleted.Revision) {
				return invalid("deleted identity variant")
			}
		case "allocated":
			if !row.Changed || len(row.Resource) != 0 || row.Deleted != nil || len(row.OwnerRevisions) != 0 {
				return invalid("allocation-only variant")
			}
		default:
			return invalid("unknown outcome")
		}
		for j, owner := range row.OwnerRevisions {
			if row.ResourceKind != graph.KindLink || !retainedBatchLocal(scope, owner.ID) || !authorityID.MatchString(owner.Revision) || j > 0 && graph.CompareCodeUnits(row.OwnerRevisions[j-1].ID, owner.ID) >= 0 {
				return invalid("owner revision mapping")
			}
			owners[owner.ID] = true
		}
		changed = changed || row.Changed
		outcomes[row.ID] = row
	}
	if changed != result.Changed {
		return invalid("aggregate change flag")
	}
	for _, row := range result.Resources {
		if mutations[row.ID] == "" && (row.ResourceKind != graph.KindBead || !owners[row.ID]) {
			return invalid("unrelated extra outcome")
		}
		if row.ResourceKind == graph.KindBead && len(row.Resource) != 0 {
			var bead retainedBatchBeadEnvelope
			if json.Unmarshal(row.Resource, &bead) != nil {
				return invalid("owned capture envelope")
			}
			for _, capture := range bead.Owned {
				var link retainedBatchLinkEnvelope
				if json.Unmarshal(capture, &link) != nil {
					return invalid("owned Link envelope")
				}
				if target, exists := outcomes[link.ID]; exists {
					if target.ResourceKind != graph.KindLink || len(target.Resource) == 0 || !bytes.Equal(capture, target.Resource) {
						return invalid("owned Link/final outcome disagreement")
					}
					found := false
					for _, owner := range target.OwnerRevisions {
						found = found || owner.ID == bead.ID && owner.Revision == bead.Revision
					}
					if !found {
						return invalid("missing captured owner revision")
					}
				}
			}
		}
		for _, owner := range row.OwnerRevisions {
			target, exists := outcomes[owner.ID]
			if !exists || target.ResourceKind != graph.KindBead || target.Outcome == "allocated" {
				return invalid("missing owner outcome")
			}
			revision := ""
			if target.Deleted != nil {
				revision = target.Deleted.Revision // Last live; never a terminal token.
			} else {
				var envelope retainedBatchEnvelope
				if json.Unmarshal(target.Resource, &envelope) != nil {
					return invalid("owner envelope")
				}
				revision = envelope.Revision
			}
			if owner.Revision != revision {
				return invalid("owner final revision")
			}
		}
	}
	for id := range mutations {
		if _, exists := outcomes[id]; !exists {
			return invalid("missing mutation outcome")
		}
	}
	return result, nil
}

func retainedBatchEqualIDs(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func retainedBatchLocal(scope, id string) bool {
	_, err := graph.ScopeLocalPublicResourceID(scope, id)
	return err == nil
}

type retainedBatchEnvelope struct {
	ID          string          `json:"id"`
	Type        string          `json:"type"`
	Revision    string          `json:"revision"`
	Version     string          `json:"version"`
	Properties  json.RawMessage `json:"properties"`
	Metadata    json.RawMessage `json:"metadata,omitempty"`
	Attribution Attribution     `json:"attribution"`
}

type retainedBatchBeadEnvelope struct {
	retainedBatchEnvelope
	Owned []json.RawMessage `json:"owned"`
}

type retainedBatchLinkEnvelope struct {
	retainedBatchEnvelope
	Source EndpointReference `json:"source"`
	Target EndpointReference `json:"target"`
}

func validateRetainedBatchEnvelope(raw []byte, scope, id string, kind graph.ResourceKind) error {
	var envelope retainedBatchEnvelope
	invalid := func() error {
		return fmt.Errorf("%w: invalid historical Resource carrier", ErrInvalidStore)
	}
	if kind == graph.KindLink {
		var link retainedBatchLinkEnvelope
		if json.Unmarshal(raw, &link) != nil || !sameVersionJSON(raw, link) || graph.ValidateReferenceURI(link.Source.URI) != nil || graph.ValidateReferenceURI(link.Target.URI) != nil {
			return invalid()
		}
		envelope = link.retainedBatchEnvelope
	} else {
		var bead retainedBatchBeadEnvelope
		if json.Unmarshal(raw, &bead) != nil || !sameVersionJSON(raw, bead) || bead.Owned == nil || len(bead.Owned) > PreviewOwnedLinkLimit {
			return invalid()
		}
		envelope = bead.retainedBatchEnvelope
		// Existing carriers order owned captures by their private storage path.
		// Public identities can have a different order; preserve that array.
		seen := map[string]bool{}
		for _, rawLink := range bead.Owned {
			var link retainedBatchLinkEnvelope
			if json.Unmarshal(rawLink, &link) != nil || link.Source.URI != id || seen[link.ID] {
				return invalid()
			}
			if err := validateRetainedBatchEnvelope(rawLink, scope, link.ID, graph.KindLink); err != nil {
				return invalid()
			}
			seen[link.ID] = true
		}
	}
	if envelope.ID != id || !retainedBatchLocal(scope, envelope.ID) || graph.ValidateReferenceURI(envelope.Type) != nil || !authorityID.MatchString(envelope.Revision) || envelope.Version != envelope.Revision || !validVersionAttribution(envelope.Attribution, envelope.Attribution.Actor) {
		return invalid()
	}
	if len(envelope.Properties) == 0 || envelope.Properties[0] != '{' {
		return invalid()
	}
	if envelope.Metadata != nil {
		metadata, err := commonMetadata(envelope.Metadata)
		if err != nil || !bytes.Equal(metadata, envelope.Metadata) {
			return invalid()
		}
	}
	return nil
}
