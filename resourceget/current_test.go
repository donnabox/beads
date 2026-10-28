package resourceget_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/resourceget"
)

const scope = "https://example.test/read/"
const bead = `{"id":"https://example.test/read/task-1","type":"https://example.test/read/memory","revision":"r1","properties":{"empty":"","number":9007199254740993,"large":1e400},"metadata":{},"ownedLinks":{"https://example.test/read/subject-to":[]}}`
const link = `{"id":"https://example.test/read/team/link-1","type":"https://example.test/read/related","revision":"r2","properties":{},"metadata":{"reviewed":true},"source":{"uri":"https://example.test/read/task-1","revision":"opaque@=pin"},"target":"urn:example:target"}`

func TestNativeDocumentContractAndCopies(t *testing.T) {
	for _, tc := range []struct {
		kind resourceget.Kind
		raw  string
	}{{resourceget.Bead, bead}, {resourceget.Link, link}} {
		input := []byte(tc.raw)
		doc, err := resourceget.ParseLegacyCurrent(scope, tc.kind, input)
		if err != nil {
			t.Fatal(err)
		}
		if doc.Kind() != tc.kind || doc.ScopeURL() != scope || doc.ID() == "" || doc.Revision() == "" {
			t.Fatal("lost identity")
		}
		contract := doc.Contract()
		if contract.SpecCommit != "bad780f19761e5714873aa3a85adeb75bc885928" || contract.SchemaID != "https://github.com/gastownhall/bdp/schemas/bdp-v0.schema.json" || contract.BDPVersion != "0" {
			t.Fatalf("wrong contract: %+v", contract)
		}
		input[0] = '!'
		output := doc.NativeJSON()
		if !bytes.Equal(output, []byte(tc.raw)) {
			t.Fatal("input mutation changed document")
		}
		output[0] = '!'
		marshaled, err := doc.MarshalJSON()
		if err != nil || string(marshaled) != tc.raw {
			t.Fatalf("output mutation changed document: %s %v", marshaled, err)
		}
		marshaled[0] = '!'
		if string(doc.NativeJSON()) != tc.raw {
			t.Fatal("MarshalJSON exposed document bytes")
		}
		encoded, err := json.Marshal(doc)
		if err != nil || !json.Valid(encoded) || bytes.Contains(encoded, []byte("SpecCommit")) {
			t.Fatalf("not a native record: %s %v", encoded, err)
		}
		if err := json.Unmarshal(encoded, &doc); !errors.Is(err, resourceget.ErrInvalidDocument) || string(doc.NativeJSON()) != tc.raw {
			t.Fatalf("generic decode silently replaced a checked document: %v", err)
		}
	}
	var zero resourceget.NativeDocument
	if _, err := json.Marshal(zero); !errors.Is(err, resourceget.ErrInvalidDocument) {
		t.Fatalf("zero document marshaled: %v", err)
	}
	if zero.Contract() != (resourceget.Contract{}) {
		t.Fatal("zero value claims a contract")
	}
}

func TestCurrentDocumentRefusesIncompleteOrAmbiguousRecords(t *testing.T) {
	cases := map[string]string{
		"missing metadata":        strings.Replace(bead, `,"metadata":{}`, "", 1),
		"null metadata":           strings.Replace(bead, `"metadata":{}`, `"metadata":null`, 1),
		"duplicate":               strings.Replace(bead, `"revision":"r1"`, `"revision":"r1","revision":"r2"`, 1),
		"case variant":            strings.Replace(bead, `"revision"`, `"Revision"`, 1),
		"empty revision":          strings.Replace(bead, `"revision":"r1"`, `"revision":""`, 1),
		"foreign identity":        strings.Replace(bead, scope+"task-1", "https://foreign.test/task-1", 1),
		"query identity":          strings.Replace(bead, scope+"task-1", scope+"task-1?q=1", 1),
		"bad type":                strings.Replace(bead, scope+"memory", "urn:example:type", 1),
		"bad ownership key":       strings.Replace(bead, scope+"subject-to", "urn:example:type", 1),
		"aggregate":               strings.TrimSuffix(bead, "}") + `,"links":{"items":[],"next":null}}`,
		"duplicate open property": strings.Replace(bead, `"empty":""`, `"empty":"","empty":"changed"`, 1),
		"invalid Unicode":         strings.Replace(bead, `"empty":""`, `"empty":"\ud800"`, 1),
		"not JSON number":         strings.Replace(bead, "1e400", "NaN", 1),
	}
	for name, raw := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := resourceget.ParseLegacyCurrent(scope, resourceget.Bead, []byte(raw)); !errors.Is(err, resourceget.ErrInvalidDocument) {
				t.Fatalf("accepted malformed record: %v", err)
			}
		})
	}
	for _, kind := range []resourceget.Kind{resourceget.Any, resourceget.Kind(255)} {
		if _, err := resourceget.ParseLegacyCurrent(scope, kind, []byte(bead)); !errors.Is(err, resourceget.ErrInvalidDocument) {
			t.Fatalf("accepted kind %v", kind)
		}
	}
	if _, err := resourceget.ParseLegacyCurrent("https://EXAMPLE.test/read/", resourceget.Bead, []byte(bead)); !errors.Is(err, resourceget.ErrInvalidDocument) {
		t.Fatalf("accepted noncanonical Scope: %v", err)
	}
	if _, err := resourceget.ParseLegacyCurrent("https://EXAMPLE.test/read/", resourceget.Bead, []byte(bead)); !errors.Is(err, graphops.ErrValidation) {
		t.Fatalf("parser lost the public validation cause: %v", err)
	}
}

func TestCurrentDocumentOptionalOwnershipAndProvenance(t *testing.T) {
	plain := strings.Replace(bead, `,"ownedLinks":{"https://example.test/read/subject-to":[]}`, "", 1)
	context := `"changeContext":{"committedAt":{"state":"undetermined"},"agent":{"state":"absent"},"message":{"state":"present","value":""}}`
	for _, extra := range []string{"", `"attribution":{"principal":"agent:test","basis":"writer-supplied"}`, `"attribution":{"principal":"imported principal","basis":"unknown"}`, context} {
		raw := plain
		if extra != "" {
			raw = strings.TrimSuffix(raw, "}") + "," + extra + "}"
		}
		if _, err := resourceget.ParseLegacyCurrent(scope, resourceget.Bead, []byte(raw)); err != nil {
			t.Fatalf("rejected admitted optional provenance/ownership: %v", err)
		}
	}
	for _, extra := range []string{`"attribution":{"principal":"","basis":"writer-supplied"}`, `"attribution":{"principal":"agent:test","basis":"invented"}`, strings.Replace(context, `"state":"undetermined"`, `"state":"absent"`, 1), strings.Replace(context, `"state":"absent"`, `"state":"present"`, 1)} {
		raw := strings.TrimSuffix(plain, "}") + "," + extra + "}"
		if _, err := resourceget.ParseLegacyCurrent(scope, resourceget.Bead, []byte(raw)); !errors.Is(err, resourceget.ErrInvalidDocument) {
			t.Fatalf("accepted malformed provenance: %v", err)
		}
	}
}

func TestWholeReferencesAndOwnedCurrentRecords(t *testing.T) {
	for name, raw := range map[string]string{
		"relative target":  strings.Replace(link, `"urn:example:target"`, `"task-2"`, 1),
		"empty pin":        strings.Replace(link, `"revision":"opaque@=pin"`, `"revision":""`, 1),
		"partial pin":      strings.Replace(link, `,"revision":"opaque@=pin"`, "", 1),
		"extra pin member": strings.Replace(link, `"revision":"opaque@=pin"`, `"revision":"opaque@=pin","extra":true`, 1),
		"missing metadata": strings.Replace(link, `,"metadata":{"reviewed":true}`, "", 1),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := resourceget.ParseLegacyCurrent(scope, resourceget.Link, []byte(raw)); !errors.Is(err, resourceget.ErrInvalidDocument) {
				t.Fatalf("accepted bad Reference/current record: %v", err)
			}
		})
	}
	owned := strings.Replace(bead, `"https://example.test/read/subject-to":[]`, `"https://example.test/read/related":[`+link+`]`, 1)
	if _, err := resourceget.ParseLegacyCurrent(scope, resourceget.Bead, []byte(owned)); err != nil {
		t.Fatal(err)
	}
	for name, raw := range map[string]string{
		"wrong source":            strings.Replace(owned, `"uri":"https://example.test/read/task-1"`, `"uri":"https://example.test/read/other"`, 1),
		"wrong bucket":            strings.Replace(owned, `"https://example.test/read/related":[`, `"https://example.test/read/subject-to":[`, 1),
		"nested metadata missing": strings.Replace(owned, `,"metadata":{"reviewed":true}`, "", 1),
		"duplicate owned ID":      strings.Replace(owned, `[`+link+`]`, `[`+link+`,`+link+`]`, 1),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := resourceget.ParseLegacyCurrent(scope, resourceget.Bead, []byte(raw)); !errors.Is(err, resourceget.ErrInvalidDocument) {
				t.Fatalf("accepted inconsistent owned record: %v", err)
			}
		})
	}
}
