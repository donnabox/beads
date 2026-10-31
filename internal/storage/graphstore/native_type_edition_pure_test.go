package graphstore

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/steveyegge/beads/resourceget"
)

func TestNativeRetainedTypeEditionPreservesExactEvidence(t *testing.T) {
	pin := resourceget.TypeReference{URI: "https://scope.example/native-issue/task", Revision: "definition-inception"}
	domain := []byte(`{"id":"https://scope.example/333","owned":[],"properties":{"title":"Hello world"},"revision":"record-edition","type":"https://scope.example/native-issue/task"}`)
	encoded, err := encodeNativeSnapshotEdition(domain, pin)
	if err != nil {
		t.Fatal(err)
	}
	restored, actual, err := decodeNativeSnapshotEdition(encoded)
	if err != nil || actual != pin || !bytes.Equal(restored, domain) {
		t.Fatalf("exact retained correspondence = %s, %#v, %v", restored, actual, err)
	}
	var header map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &header); err != nil || !bytes.Equal(header["owned"], []byte(`[]`)) {
		t.Fatalf("private evidence displaced existing root query fields: %s, %v", encoded, err)
	}
	if bytes.Contains(restored, []byte(nativeTypeEditionMember)) {
		t.Fatal("private evidence escaped into native domain")
	}
	for name, raw := range map[string][]byte{
		"missing":       domain,
		"null":          bytes.Replace(encoded, header[nativeTypeEditionMember], []byte(`null`), 1),
		"wrong-nominal": bytes.Replace(encoded, []byte(`"type":"https://scope.example/native-issue/task"`), []byte(`"type":"https://scope.example/native-issue/bug"`), 1),
		"unknown-codec": bytes.Replace(encoded, []byte(nativeTypeEditionCodec), []byte("native-type-edition/future"), 1),
		"duplicate":     append([]byte(`{"$nativeTypeEdition":null,`), encoded[1:]...),
	} {
		t.Run(name, func(t *testing.T) {
			body, reference, err := decodeNativeSnapshotEdition(raw)
			if !errors.Is(err, ErrInvalidStore) || len(body) != 0 || reference != (resourceget.TypeReference{}) {
				t.Fatalf("ambiguous evidence produced a partial record: %s, %#v, %v", body, reference, err)
			}
		})
	}
	legacy := &Store{}
	unchanged, err := legacy.nativeDomainSnapshot(domain)
	if err != nil || !bytes.Equal(unchanged, domain) {
		t.Fatalf("ordinary11 codec changed: %s, %v", unchanged, err)
	}
}

func TestNativeIssueTypeEditionKeepsAbsentContextDistinct(t *testing.T) {
	pin := resourceget.TypeReference{URI: "https://scope.example/native-issue/task", Revision: "definition-inception"}
	encoded, err := encodeNativeIssueTypeEdition(issueRevisionContext{}, pin)
	if err != nil {
		t.Fatal(err)
	}
	value, actual, err := decodeNativeIssueTypeEdition(encoded)
	if err != nil || actual != pin {
		t.Fatalf("exact Issue evidence = %#v, %v", actual, err)
	}
	restored, err := encodeIssueRevisionContext(value)
	if err != nil || len(restored) != 0 {
		t.Fatalf("absent private attribution became an authored value: %s, %v", restored, err)
	}
	for name, raw := range map[string][]byte{"missing": []byte(`{}`), "null": []byte(`null`), "wrong-codec": bytes.Replace(encoded, []byte(nativeTypeEditionCodec), []byte("unknown"), 1)} {
		t.Run(name, func(t *testing.T) {
			_, reference, err := decodeNativeIssueTypeEdition(raw)
			if !errors.Is(err, ErrInvalidStore) || reference != (resourceget.TypeReference{}) {
				t.Fatalf("invalid Issue evidence selected a Type: %#v, %v", reference, err)
			}
		})
	}
}
