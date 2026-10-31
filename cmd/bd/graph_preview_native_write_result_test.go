package main

import (
	"errors"
	"strings"
	"testing"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/resourceget"
)

type nativeWriteResultFixture struct {
	document resourceget.ResourceDocument
	present  bool
	calls    int
}

func (f *nativeWriteResultFixture) ResourceDocument() (resourceget.ResourceDocument, bool) {
	f.calls++
	return f.document, f.present
}

func TestTypesNativeWriteResultUsesAppliedCapture(t *testing.T) {
	const scope = "https://example.test/applied/"
	document, err := resourceget.NewResourceDocument(scope, graph.KindBead, []byte(`{"id":"https://example.test/applied/333","revision":"created-1","type":{"uri":"https://example.test/applied/task","revision":"selected-before-commit"},"properties":{"title":"Hello world"},"ownedLinks":{}}`))
	if err != nil {
		t.Fatal(err)
	}
	applied := &nativeWriteResultFixture{document: document, present: true}
	value, err := graphPreviewNativeWriteResult(graphstore.TypesFreshSchemaVersion, applied)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := value.(resourceget.ResourceDocument)
	if !ok || applied.calls != 1 || string(got.NativeJSON()) != string(document.NativeJSON()) || got.Type().Revision != "selected-before-commit" {
		t.Fatal("applied snapshot or exact selected Type was replaced")
	}
	legacy := &graphstore.IssueRecord{ID: scope + "333"}
	old, err := graphPreviewNativeWriteResult(graphstore.SchemaVersion, legacy)
	if err != nil || old != legacy {
		t.Fatal("legacy-layout creation representation changed")
	}
}

func TestTypesNativeWriteResultRefusesMissingAppliedCapture(t *testing.T) {
	for _, value := range []any{
		&graphstore.IssueRecord{},
		&nativeWriteResultFixture{},
		&nativeWriteResultFixture{present: true},
		nil,
	} {
		result, err := graphPreviewNativeWriteResult(graphstore.TypesFreshSchemaVersion, value)
		if result != nil || !errors.Is(err, graphstore.ErrOutcomeUnknown) {
			t.Fatalf("missing committed capture became a safe retry/result: %#v %v", result, err)
		}
	}
	const allocated = "https://example.test/applied/generated-identity"
	_, err := graphPreviewNativeWriteResult(graphstore.TypesFreshSchemaVersion, graphstore.IssueRecord{ID: allocated})
	if !errors.Is(err, graphstore.ErrOutcomeUnknown) || !strings.Contains(err.Error(), allocated) {
		t.Fatal("missing capture hid the authority-allocated identity", err)
	}
}
