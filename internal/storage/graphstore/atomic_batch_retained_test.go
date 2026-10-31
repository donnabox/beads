package graphstore

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/canonicalbatch"
)

const retainedTestScope = "https://example.invalid/"
const retainedTestRevision = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
const retainedTestTransaction = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"

func retainedTestJSON(t *testing.T, value any) json.RawMessage {
	t.Helper()
	raw, err := canonicalJSON(value)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func retainedTestEnvelope(t *testing.T, id string, native bool) json.RawMessage {
	t.Helper()
	properties := json.RawMessage(`{"body":"kept","title":"Original","user":{"revision":3}}`)
	typ := retainedTestScope + "retired-memory-type"
	if native {
		// Native optional properties and metadata are retained verbatim, without
		// rerunning current domain classification/configuration or JSON aliases.
		properties = json.RawMessage(`{"created_at":"2026-10-29T00:00:00Z","estimated_minutes":0,"id":"native-a","issue_type":"historical-class","metadata":{"x":true},"priority":2,"status":"open","title":"Native","updated_at":"2026-10-29T00:00:00Z"}`)
		typ = retainedTestScope + "historical-class"
	}
	return retainedTestJSON(t, retainedBatchBeadEnvelope{retainedBatchEnvelope: retainedBatchEnvelope{ID: retainedTestScope + id, Type: typ, Revision: retainedTestRevision, Version: retainedTestRevision, Properties: properties, Metadata: json.RawMessage(`{}`), Attribution: Attribution{Actor: "prior-author", Status: "claimed", RecordedAt: "2026-10-29T00:00:00Z"}}, Owned: []json.RawMessage{}})
}

func retainedTestFixture(t *testing.T) (BatchResult, []canonicalbatch.Member) {
	t.Helper()
	return BatchResult{CapabilityVersion: canonicalbatch.Version, Changed: true, Transaction: retainedTestTransaction, Results: []BatchOperationResult{{OperationIndex: 0, Operation: "createBead", IDs: []string{retainedTestScope + "a"}}}, Resources: []BatchResourceOutcome{{ID: retainedTestScope + "a", ResourceKind: graph.KindBead, Changed: true, Outcome: "created", Resource: retainedTestEnvelope(t, "a", false)}}}, []canonicalbatch.Member{{Operation: "createBead", ID: "a"}}
}

func TestAtomicBatchRetainedPositiveCarriers(t *testing.T) {
	for _, scenario := range []string{"generic", "native", "noop", "allocated", "empty-set", "metadata-omitted", "bindings", "remote-type"} {
		t.Run(scenario, func(t *testing.T) {
			result, members := retainedTestFixture(t)
			switch scenario {
			case "native":
				result.Resources[0].Resource = retainedTestEnvelope(t, "a", true)
				result.Resources[0].Outcome = "updated"
				result.Results[0].Operation = "updateBead"
				members = []canonicalbatch.Member{{Operation: "updateBead", Bead: canonicalbatch.Operand{URI: retainedTestScope + "a"}}}
			case "noop":
				result.Changed, result.Transaction = false, ""
				result.Resources[0].Changed, result.Resources[0].Outcome = false, "unchanged"
				result.Results[0].Operation = "updateBead"
				members = []canonicalbatch.Member{{Operation: "updateBead", Bead: canonicalbatch.Operand{URI: retainedTestScope + "a"}}}
			case "allocated":
				result.Resources[0].Resource, result.Resources[0].Outcome = nil, "allocated"
				result.Results = append(result.Results, BatchOperationResult{OperationIndex: 1, Operation: "deleteBead", IDs: []string{retainedTestScope + "a"}})
				members = append(members, canonicalbatch.Member{Operation: "deleteBead", Bead: canonicalbatch.Operand{URI: retainedTestScope + "a"}})
			case "empty-set":
				result.Changed, result.Transaction, result.Resources = false, "", []BatchResourceOutcome{}
				result.Results = []BatchOperationResult{{Operation: "select", IDs: []string{}}, {OperationIndex: 1, Operation: "deleteBead", IDs: []string{}}}
				members = []canonicalbatch.Member{{Operation: "select", Name: "empty"}, {Operation: "deleteBead", Bead: canonicalbatch.Operand{Binding: "empty"}}}
			case "bindings":
				members[0].Name = "renamed"
				result.Results[0].OperationName = "original-capture"
				members = append(members, canonicalbatch.Member{Operation: "updateBead", Bead: canonicalbatch.Operand{Binding: "renamed"}})
				result.Results = append(result.Results, BatchOperationResult{OperationIndex: 1, Operation: "updateBead", IDs: []string{retainedTestScope + "a"}})
			case "remote-type":
				var envelope retainedBatchBeadEnvelope
				if err := json.Unmarshal(result.Resources[0].Resource, &envelope); err != nil {
					t.Fatal(err)
				}
				envelope.Type = "https://old-authority.invalid/retired-definition"
				result.Resources[0].Resource = retainedTestJSON(t, envelope)
			case "metadata-omitted":
				var envelope retainedBatchBeadEnvelope
				if err := json.Unmarshal(result.Resources[0].Resource, &envelope); err != nil {
					t.Fatal(err)
				}
				envelope.Metadata = nil
				result.Resources[0].Resource = retainedTestJSON(t, envelope)
			}
			raw := retainedTestJSON(t, result)
			decoded, err := validateAtomicBatchRetained(raw, retainedTestScope, members)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(raw, retainedTestJSON(t, decoded)) {
				t.Fatal("validation rewrote historical carrier")
			}
		})
	}
}

func retainedTestOwnedFixture(t *testing.T) (BatchResult, []canonicalbatch.Member) {
	t.Helper()
	result, _ := retainedTestFixture(t)
	var source retainedBatchBeadEnvelope
	if err := json.Unmarshal(result.Resources[0].Resource, &source); err != nil {
		t.Fatal(err)
	}
	link := retainedBatchLinkEnvelope{retainedBatchEnvelope: retainedBatchEnvelope{ID: retainedTestScope + "z-link", Type: retainedTestScope + "retired-link-type", Revision: retainedTestRevision, Version: retainedTestRevision, Properties: json.RawMessage(`{"note":"kept"}`), Metadata: json.RawMessage(`{}`), Attribution: source.Attribution}, Source: EndpointReference{URI: source.ID}, Target: EndpointReference{URI: "https://remote.invalid/unreturned-target", Revision: "historical-exact-token"}}
	linkRaw := retainedTestJSON(t, link)
	source.Owned = []json.RawMessage{linkRaw}
	result.Results[0] = BatchOperationResult{Operation: "createLink", IDs: []string{link.ID}}
	result.Resources = []BatchResourceOutcome{{ID: source.ID, ResourceKind: graph.KindBead, Changed: true, Outcome: "updated", Resource: retainedTestJSON(t, source)}, {ID: link.ID, ResourceKind: graph.KindLink, Changed: true, Outcome: "created", Resource: linkRaw, OwnerRevisions: []BatchOwnerRevision{{ID: source.ID, Revision: source.Revision}}}}
	return result, []canonicalbatch.Member{{Operation: "createLink", ID: "z-link"}}
}

func retainedTestScalar(t *testing.T, result *BatchResult) {
	t.Helper()
	for i := range result.Results {
		row := &result.Results[i]
		if len(row.IDs) != 1 {
			t.Fatal("scalar fixture requires singleton operation results")
		}
		row.ID, row.IDs = row.IDs[0], nil
	}
}

func TestAtomicBatchRetainedScalarCompleteCarriers(t *testing.T) {
	for _, scenario := range []string{"created", "native", "noop", "allocated", "owned-pin", "deleted-owner"} {
		t.Run(scenario, func(t *testing.T) {
			result, members := retainedTestFixture(t)
			switch scenario {
			case "native", "noop":
				result.Results[0].Operation = "updateBead"
				members = []canonicalbatch.Member{{Operation: "updateBead", Bead: canonicalbatch.Operand{URI: retainedTestScope + "a"}}}
				result.Resources[0].Outcome = "updated"
				if scenario == "native" {
					result.Resources[0].Resource = retainedTestEnvelope(t, "a", true)
				} else {
					result.Changed, result.Transaction = false, ""
					result.Resources[0].Changed, result.Resources[0].Outcome = false, "unchanged"
				}
			case "allocated":
				result.Resources[0].Resource, result.Resources[0].Outcome = nil, "allocated"
				result.Results = append(result.Results, BatchOperationResult{OperationIndex: 1, Operation: "deleteBead", IDs: []string{retainedTestScope + "a"}})
				members = append(members, canonicalbatch.Member{Operation: "deleteBead", Bead: canonicalbatch.Operand{URI: retainedTestScope + "a"}})
			case "owned-pin", "deleted-owner":
				result, members = retainedTestOwnedFixture(t)
				if scenario == "deleted-owner" {
					for i := range result.Resources {
						row := &result.Resources[i]
						row.Outcome, row.Resource = "deleted", nil
						row.Deleted = &BatchDeletedIdentity{ID: row.ID, ResourceKind: row.ResourceKind, Type: retainedTestScope + "retired-type", Revision: retainedTestRevision}
					}
					result.Results = []BatchOperationResult{{Operation: "deleteLink", IDs: []string{retainedTestScope + "z-link"}}, {OperationIndex: 1, Operation: "deleteBead", IDs: []string{retainedTestScope + "a"}}}
					members = []canonicalbatch.Member{{Operation: "deleteLink", Link: canonicalbatch.Operand{URI: retainedTestScope + "z-link"}}, {Operation: "deleteBead", Bead: canonicalbatch.Operand{URI: retainedTestScope + "a"}}}
				}
			}
			retainedTestScalar(t, &result)
			raw := retainedTestJSON(t, result)
			decoded, err := validateAtomicBatchRetained(raw, retainedTestScope, members)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(raw, retainedTestJSON(t, decoded)) {
				t.Fatal("validation rewrote complete historical scalar carrier or pins")
			}
		})
	}
}

func TestAtomicBatchRetainedRejectsScalarDisguisedCorruption(t *testing.T) {
	for _, scenario := range []string{"missing-all", "missing-one", "wrong-variant", "unrelated-extra", "owner-only-variant", "missing-owner", "owner-revision", "missing-captured-owner"} {
		t.Run(scenario, func(t *testing.T) {
			result, members := retainedTestFixture(t)
			switch scenario {
			case "missing-all":
				// This was the fictional compatibility fixture: no historical
				// producer emitted an update without its Resource outcome.
				result.Changed, result.Transaction, result.Resources = false, "", []BatchResourceOutcome{}
				result.Results[0].Operation = "updateBead"
				members = []canonicalbatch.Member{{Operation: "updateBead", Bead: canonicalbatch.Operand{URI: retainedTestScope + "a"}}}
			case "missing-one":
				// Keep another changed outcome so completeness, rather than the
				// aggregate Changed check, must reject this missing identity.
				result.Results = append(result.Results, BatchOperationResult{OperationIndex: 1, Operation: "createBead", IDs: []string{retainedTestScope + "b"}})
				members = append(members, canonicalbatch.Member{Operation: "createBead", ID: "b"})
			case "wrong-variant":
				result.Results[0].Operation = "updateBead"
				members = []canonicalbatch.Member{{Operation: "updateBead", Bead: canonicalbatch.Operand{URI: retainedTestScope + "a"}}}
			case "unrelated-extra":
				result.Resources = append(result.Resources, BatchResourceOutcome{ID: retainedTestScope + "z", ResourceKind: graph.KindBead, Changed: true, Outcome: "updated", Resource: retainedTestEnvelope(t, "z", false)})
			default:
				result, members = retainedTestOwnedFixture(t)
				switch scenario {
				case "owner-only-variant":
					result.Resources[0].Outcome = "created"
				case "missing-owner":
					result.Resources = result.Resources[1:]
				case "owner-revision":
					result.Resources[1].OwnerRevisions[0].Revision = retainedTestTransaction
				case "missing-captured-owner":
					// A direct owner mutation must not disguise a missing
					// correlation between its capture and the Link outcome.
					result.Results = append(result.Results, BatchOperationResult{OperationIndex: 1, Operation: "updateBead", IDs: []string{retainedTestScope + "a"}})
					members = append(members, canonicalbatch.Member{Operation: "updateBead", Bead: canonicalbatch.Operand{URI: retainedTestScope + "a"}})
					result.Resources[1].OwnerRevisions = nil
				}
			}
			retainedTestScalar(t, &result)
			if _, err := validateAtomicBatchRetained(retainedTestJSON(t, result), retainedTestScope, members); !errors.Is(err, ErrInvalidStore) {
				t.Fatalf("scalar carrier disguised retained corruption: %v", err)
			}
		})
	}
}

func TestAtomicBatchRetainedOwnersPinsAndDeletedLastLive(t *testing.T) {
	result, members := retainedTestOwnedFixture(t)
	raw := retainedTestJSON(t, result)
	decoded, err := validateAtomicBatchRetained(raw, retainedTestScope, members)
	if err != nil || !bytes.Equal(raw, retainedTestJSON(t, decoded)) {
		t.Fatalf("owner-only row/pinned historical endpoint rejected: %v", err)
	}
	for i := range result.Resources {
		row := &result.Resources[i]
		row.Outcome, row.Resource = "deleted", nil
		row.Deleted = &BatchDeletedIdentity{ID: row.ID, ResourceKind: row.ResourceKind, Type: retainedTestScope + "retired-type", Revision: retainedTestRevision}
	}
	result.Results = []BatchOperationResult{{Operation: "deleteLink", IDs: []string{retainedTestScope + "z-link"}}, {OperationIndex: 1, Operation: "deleteBead", IDs: []string{retainedTestScope + "a"}}}
	members = []canonicalbatch.Member{{Operation: "deleteLink", Link: canonicalbatch.Operand{URI: retainedTestScope + "z-link"}}, {Operation: "deleteBead", Bead: canonicalbatch.Operand{URI: retainedTestScope + "a"}}}
	if _, err := validateAtomicBatchRetained(retainedTestJSON(t, result), retainedTestScope, members); err != nil {
		t.Fatalf("deleted owner last-live revision rejected: %v", err)
	}
}

func TestAtomicBatchRetainedPreservesOwnedCarrierOrder(t *testing.T) {
	result, members := retainedTestOwnedFixture(t)
	var bead retainedBatchBeadEnvelope
	if err := json.Unmarshal(result.Resources[0].Resource, &bead); err != nil {
		t.Fatal(err)
	}
	var second retainedBatchLinkEnvelope
	if err := json.Unmarshal(bead.Owned[0], &second); err != nil {
		t.Fatal(err)
	}
	second.ID = retainedTestScope + "b-link"
	bead.Owned = append(bead.Owned, retainedTestJSON(t, second))
	result.Resources[0].Resource = retainedTestJSON(t, bead)
	raw := retainedTestJSON(t, result)
	got, err := validateAtomicBatchRetained(raw, retainedTestScope, members)
	if err != nil || !bytes.Equal(raw, retainedTestJSON(t, got)) {
		t.Fatalf("historical private-path order was rewritten or refused: %v", err)
	}
	bead.Owned = append(bead.Owned, bead.Owned[0])
	result.Resources[0].Resource = retainedTestJSON(t, bead)
	if _, err := validateAtomicBatchRetained(retainedTestJSON(t, result), retainedTestScope, members); !errors.Is(err, ErrInvalidStore) {
		t.Fatalf("duplicate owned capture accepted: %v", err)
	}
}

func TestAtomicBatchRetainedRequiresCapturedOwnerCorrelation(t *testing.T) {
	result, members := retainedTestOwnedFixture(t)
	// A directly mutated owner must not hide a missing Link-to-owner mapping.
	result.Results = append(result.Results, BatchOperationResult{OperationIndex: 1, Operation: "updateBead", IDs: []string{retainedTestScope + "a"}})
	members = append(members, canonicalbatch.Member{Operation: "updateBead", Bead: canonicalbatch.Operand{URI: retainedTestScope + "a"}})
	if _, err := validateAtomicBatchRetained(retainedTestJSON(t, result), retainedTestScope, members); err != nil {
		t.Fatalf("valid correlated owner fixture refused: %v", err)
	}
	result.Resources[1].OwnerRevisions = nil
	if _, err := validateAtomicBatchRetained(retainedTestJSON(t, result), retainedTestScope, members); !errors.Is(err, ErrInvalidStore) {
		t.Fatalf("missing owner correlation accepted: %v", err)
	}
}

func TestAtomicBatchRetainedRejectsCorruptSuccess(t *testing.T) {
	for _, scenario := range []string{"missing-outcome", "unrelated-outcome", "unknown-outcome", "kind", "missing-body", "allocated-body", "changed", "transaction-missing", "transaction-spurious", "revision", "outer-unknown", "outcome-unknown", "envelope-unknown", "properties-null", "metadata-null", "cross-scope", "wrong-literal", "resources-null", "mapping-duplicate", "mixed-carriers"} {
		t.Run(scenario, func(t *testing.T) {
			result, members := retainedTestFixture(t)
			switch scenario {
			case "missing-outcome":
				result.Resources = []BatchResourceOutcome{}
			case "unrelated-outcome":
				result.Resources = append(result.Resources, BatchResourceOutcome{ID: retainedTestScope + "z", ResourceKind: graph.KindBead, Changed: true, Outcome: "created", Resource: retainedTestEnvelope(t, "z", false)})
			case "unknown-outcome":
				result.Resources[0].Outcome = "future"
			case "kind":
				result.Resources[0].ResourceKind = graph.KindLink
			case "missing-body":
				result.Resources[0].Resource = nil
			case "allocated-body":
				result.Resources[0].Outcome = "allocated"
			case "changed":
				result.Changed, result.Transaction = false, ""
			case "transaction-missing":
				result.Transaction = ""
			case "transaction-spurious":
				result.Changed = false
			case "cross-scope":
				result.Results[0].IDs[0] = "https://other.invalid/a"
			case "wrong-literal":
				members[0].ID = "another"
			case "resources-null":
				result.Resources = nil
			case "mapping-duplicate":
				result.Results[0].IDs = append(result.Results[0].IDs, result.Results[0].IDs[0])
			case "mixed-carriers":
				result.Results = append(result.Results, BatchOperationResult{OperationIndex: 1, Operation: "updateBead", ID: retainedTestScope + "a"})
				members = append(members, canonicalbatch.Member{Operation: "updateBead", Bead: canonicalbatch.Operand{URI: retainedTestScope + "a"}})
			}
			raw := retainedTestJSON(t, result)
			var object map[string]json.RawMessage
			if err := json.Unmarshal(raw, &object); err != nil {
				t.Fatal(err)
			}
			if scenario == "outer-unknown" {
				object["future"] = json.RawMessage(`true`)
				raw = retainedTestJSON(t, object)
			}
			if scenario == "outcome-unknown" {
				var rows []map[string]json.RawMessage
				if err := json.Unmarshal(object["resources"], &rows); err != nil {
					t.Fatal(err)
				}
				rows[0]["future"] = json.RawMessage(`true`)
				object["resources"] = retainedTestJSON(t, rows)
				raw = retainedTestJSON(t, object)
			}
			if scenario == "revision" || scenario == "envelope-unknown" || scenario == "properties-null" || scenario == "metadata-null" {
				var envelope map[string]json.RawMessage
				if err := json.Unmarshal(result.Resources[0].Resource, &envelope); err != nil {
					t.Fatal(err)
				}
				switch scenario {
				case "revision":
					envelope["revision"] = json.RawMessage(`"not-an-issued-token"`)
				case "envelope-unknown":
					envelope["future"] = json.RawMessage(`true`)
				case "properties-null":
					envelope["properties"] = json.RawMessage(`null`)
				case "metadata-null":
					envelope["metadata"] = json.RawMessage(`null`)
				}
				result.Resources[0].Resource = retainedTestJSON(t, envelope)
				raw = retainedTestJSON(t, result)
			}
			if _, err := validateAtomicBatchRetained(raw, retainedTestScope, members); !errors.Is(err, ErrInvalidStore) {
				t.Fatalf("corrupted retained success accepted: %s: %v", raw, err)
			}
		})
	}
}

func TestAtomicBatchRetainedRejectsOwnerAndCaptureCorruption(t *testing.T) {
	for _, scenario := range []string{"owner-missing", "owner-revision", "owner-duplicate", "capture-source", "capture-disagrees", "pin-empty", "pin-extra"} {
		t.Run(scenario, func(t *testing.T) {
			result, members := retainedTestOwnedFixture(t)
			switch scenario {
			case "owner-missing":
				result.Resources = result.Resources[1:]
			case "owner-revision":
				result.Resources[1].OwnerRevisions[0].Revision = retainedTestTransaction
			case "owner-duplicate":
				result.Resources[1].OwnerRevisions = append(result.Resources[1].OwnerRevisions, result.Resources[1].OwnerRevisions[0])
			default:
				var source retainedBatchBeadEnvelope
				if err := json.Unmarshal(result.Resources[0].Resource, &source); err != nil {
					t.Fatal(err)
				}
				var link map[string]json.RawMessage
				if err := json.Unmarshal(source.Owned[0], &link); err != nil {
					t.Fatal(err)
				}
				switch scenario {
				case "capture-source":
					link["source"] = retainedTestJSON(t, retainedTestScope+"another")
				case "capture-disagrees":
					link["metadata"] = json.RawMessage(`{"different":true}`)
				case "pin-empty":
					link["target"] = retainedTestJSON(t, map[string]string{"uri": retainedTestScope + "unreturned-target", "revision": ""})
				case "pin-extra":
					link["target"] = retainedTestJSON(t, map[string]string{"uri": retainedTestScope + "unreturned-target", "revision": "pin", "extra": "lost"})
				}
				source.Owned[0] = retainedTestJSON(t, link)
				result.Resources[0].Resource = retainedTestJSON(t, source)
			}
			if _, err := validateAtomicBatchRetained(retainedTestJSON(t, result), retainedTestScope, members); !errors.Is(err, ErrInvalidStore) {
				t.Fatalf("owner/capture corruption accepted: %v", err)
			}
		})
	}
}

func TestAtomicBatchRetainedBoundAndDetachedResult(t *testing.T) {
	result, members := retainedTestFixture(t)
	raw := retainedTestJSON(t, result)
	decoded, err := validateAtomicBatchRetained(raw, retainedTestScope, members)
	if err != nil {
		t.Fatal(err)
	}
	original := bytes.Clone(raw)
	decoded.Resources[0].Resource[0] = '['
	decoded.Results[0].IDs[0] = "changed"
	if !bytes.Equal(raw, original) {
		t.Fatal("decoded result aliases retained input")
	}
	if _, err := validateAtomicBatchRetained([]byte(strings.Repeat(" ", AtomicBatchResultLimit+1)), retainedTestScope, members); !errors.Is(err, ErrInvalidStore) {
		t.Fatalf("oversized retained result was decoded: %v", err)
	}
}
