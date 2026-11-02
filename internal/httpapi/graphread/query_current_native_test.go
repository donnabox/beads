//go:build cgo

package graphread

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/resourcequery"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/issueops"
)

// A normal fresh embedded fixture, not the public CLI or a SQL-seeded schema.
// Each fixture retains ONE90-second context, with no per-stage reset/skip/server.
func currentProducerNativeStore(t *testing.T, schema int) (context.Context, graphstore.Options, *graphstore.Store) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	t.Cleanup(cancel)
	workspace, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	options := graphstore.Options{Backend: "embedded", DataDir: filepath.Join(workspace, "dolt"), Database: "current_producer", Branch: "main", TypeGeneration: graphstore.TypeGeneration, Binding: graphstore.Binding{WorkspaceID: workspace, ScopeURL: "https://producer-native.example/", AuthorityID: "0123456789abcdef0123456789abcdef", SchemaVersion: schema}}
	if err = graphstore.Init(ctx, options); err != nil {
		t.Fatal(err)
	}
	store, err := graphstore.OpenExisting(ctx, options)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Error(err)
		}
	})
	return ctx, options, store
}

func currentProducerRequest(t *testing.T, form string, selection map[string]any) []byte {
	t.Helper()
	raw, err := json.Marshal(map[string]any{"ir": resourcequery.IR, "profile": "query", "selection": selection, "result": map[string]any{"form": form}})
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func currentProducerRefs(collection, id, revision string) map[string]any {
	var reference any = id
	if revision != "" {
		reference = map[string]any{"uri": id, "revision": revision}
	}
	return map[string]any{"op": "refs", "collection": collection, "references": []any{reference}}
}

func currentProducerOne(t *testing.T, ctx context.Context, data resourcequery.Dataset, selection map[string]any, want []byte) {
	t.Helper()
	response, err := resourcequery.Execute(ctx, currentProducerRequest(t, "resources", selection), data, resourcequery.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	result, ok := response.Result.(resourcequery.ResourcesResult)
	if !ok || !response.Evaluation.Complete || len(result.Items) != 1 || !bytes.Equal(result.Items[0], want) {
		t.Fatalf("exact captured complete body changed: %+v", response)
	}
}

func currentProducerUnsupported(t *testing.T, ctx context.Context, data resourcequery.Dataset, selection map[string]any) {
	t.Helper()
	raw := currentProducerRequest(t, "resources", selection)
	if _, err := resourcequery.Decode(raw); err != nil {
		t.Fatalf("dynamic control failed before Dataset acquisition: %v", err)
	}
	response, err := resourcequery.Execute(ctx, raw, data, resourcequery.DefaultLimits())
	if !errors.Is(err, resourcequery.ErrUnsupported) || response.Result != nil || response.Evaluation.Complete {
		t.Fatalf("unavailable exact edition fell back: %+v %v", response, err)
	}
}

// Static pinned request refs are intentionally unsupported during Decode,
// before any Dataset/authority is inspected. They prove admission only.
func currentProducerStaticPinRefused(t *testing.T, selection map[string]any) {
	t.Helper()
	_, err := resourcequery.Decode(currentProducerRequest(t, "resources", selection))
	if !errors.Is(err, resourcequery.ErrUnsupported) {
		t.Fatalf("static pin admission changed: %v", err)
	}
}

func TestNativeQueryProducerCurrentCaptureCoversAllKinds(t *testing.T) {
	ctx, options, store := currentProducerNativeStore(t, graphstore.TypesFreshSchemaVersion)
	makeTask := func(path, id string) graphstore.IssueRecord {
		record, err := store.CreateIssue(ctx, path, issueops.CreateRequest{Actor: "native-producer", Issue: &issueops.Issue{ID: id, Title: id, Status: "open", IssueType: "task", Priority: 2}})
		if err != nil {
			t.Fatal(err)
		}
		return record
	}
	taskA := makeTask("beads/task-a", "bd-producer-a")
	taskB := makeTask("beads/task-b", "bd-producer-b")
	dependency, err := store.AddDependency(ctx, graphstore.DependencyRequest{PublicID: "producer-blocks", Path: "links/producer-blocks", SourcePath: "beads/task-a", TargetPath: "beads/task-b", Type: "blocks", Actor: "native-producer", ExpectedSourceRevision: taskA.Revision})
	if err != nil {
		t.Fatal(err)
	}
	if dependency.Source.Revision == taskA.Revision {
		t.Fatal("owned Dependency did not create the expected actual source successor")
	}
	source, err := store.Create(ctx, graphstore.CreateRequest{PublicID: "producer-source", Title: "Source", Body: "Keep native Memory", Actor: "native-producer"})
	if err != nil {
		t.Fatal(err)
	}
	target, err := store.Create(ctx, graphstore.CreateRequest{PublicID: "producer-target", Title: "Before", Body: "Old native body", Actor: "native-producer"})
	if err != nil {
		t.Fatal(err)
	}
	floating, err := store.AddInformationalLink(ctx, graphstore.LinkCreateRequest{PublicID: "producer-floating", SourcePath: "beads/producer-source", TargetPath: "beads/producer-target", SourcePin: graphstore.EndpointPin{Current: true}, Actor: "native-producer"})
	if err != nil {
		t.Fatal(err)
	}
	pinned, err := store.AddInformationalLink(ctx, graphstore.LinkCreateRequest{PublicID: "producer-pinned", SourcePath: "beads/producer-source", TargetPath: "beads/producer-target", SourcePin: graphstore.EndpointPin{Current: true}, TargetPin: graphstore.EndpointPin{Revision: target.Revision}, Actor: "native-producer"})
	if err != nil {
		t.Fatal(err)
	}
	if floating.Link.TargetRevision != "" || pinned.Link.TargetRevision != target.Revision {
		t.Fatal("fixture lost explicit floating versus pinned endpoint intent")
	}
	updated, err := store.UpdateMemory(ctx, graphstore.MemoryUpdateRequest{Path: "beads/producer-target", ExpectedRevision: target.Revision, Properties: graphstore.Properties{Title: "After", Body: "Current native body"}, Actor: "native-producer"})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Memory.Revision == target.Revision {
		t.Fatal("fixture needs a real retained Memory edition")
	}
	// The same fixture is reopened normally after native writes; no second
	// snapshot supplies evidence. Keep the SAME original90-second context.
	if err = store.Close(); err != nil {
		t.Fatal(err)
	}
	store, err = graphstore.OpenExisting(ctx, options)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Error(err)
		}
	})
	reader := New(store)
	// THE owning successful native read. Everything below evaluates only this
	// returned snapshot/data; no callback, supplied snapshot or second read proves it.
	snapshot, data, err := reader.resourceQueryCapture(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if data.CurrentEditions == nil || data.TypeEditions == nil || data.TypeFamilies != nil || data.TypeKinds != nil || len(data.Resources) != len(snapshot.Documents) || len(data.TypeEditions) != len(snapshot.Types) {
		t.Fatal("native bridge omitted exact authority or used legacy fallback")
	}
	bodies := map[string]json.RawMessage{}
	editions := map[string]resourcequery.Edition{}
	kinds := map[string]graph.ResourceKind{}
	for i, document := range snapshot.Documents {
		if _, duplicate := bodies[document.ID()]; duplicate || !bytes.Equal(document.NativeJSON(), data.Resources[i].Data) || document.Kind() != data.Resources[i].Kind {
			t.Fatal("top-level native capture lost unique full record correspondence")
		}
		bodies[document.ID()] = document.NativeJSON()
		editions[document.ID()] = resourcequery.Edition{ID: document.ID(), Revision: document.Revision()}
		kinds[document.ID()] = document.Kind()
	}
	currentDefinitions := map[string]bool{}
	for i, fact := range snapshot.Types {
		mapped := data.TypeEditions[i]
		if mapped.Reference != (resourcequery.Edition{ID: fact.Reference.URI, Revision: fact.Reference.Revision}) || mapped.Kind != fact.Describes || mapped.Current != fact.Current || len(mapped.Ancestors) != len(fact.Ancestors) {
			t.Fatal("captured Type facts changed")
		}
		for j, ancestor := range fact.Ancestors {
			if mapped.Ancestors[j] != (resourcequery.Edition{ID: ancestor.URI, Revision: ancestor.Revision}) {
				t.Fatal("captured ancestry substituted")
			}
		}
		if fact.Current {
			if editions[fact.Definition.URI] != (resourcequery.Edition{ID: fact.Definition.URI, Revision: fact.Definition.Revision}) || kinds[fact.Definition.URI] != graph.KindBead {
				t.Fatal("installed definition fact has no same-identity current Type Bead")
			}
			currentDefinitions[fact.Reference.URI] = true
		}
	}
	taskType, err := graphstore.NativeIssueTypeURL(store.ScopeURL(), "task")
	if err != nil {
		t.Fatal(err)
	}
	blocksType, err := graphstore.NativeDependencyTypeURL(store.ScopeURL(), "blocks")
	if err != nil {
		t.Fatal(err)
	}
	for _, uri := range []string{taskType, graphstore.MemoryTypeURL(store.ScopeURL()), blocksType, graphstore.RelatedTypeURL(store.ScopeURL())} {
		if !currentDefinitions[uri] {
			t.Fatalf("actual native installed Type missing: %s", uri)
		}
	}
	// All checked current definition facts above are joined, not just this native
	// subset. Every top-level Bead/Link below must be exact-selectable. Retained
	// ancestry/owned bodies do not become extra independent CURRENT Resources.
	for _, id := range []string{dependency.Source.ID, taskB.ID, source.ID, updated.Memory.ID, dependency.Link.ID, floating.Link.ID, pinned.Link.ID} {
		if _, present := bodies[id]; !present {
			t.Fatalf("actual native Resource omitted: %s", id)
		}
	}
	if !bytes.Equal(bodies[updated.Memory.ID], mustCurrentProducerMemory(t, updated.Memory)) {
		t.Fatal("current Memory complete applied body changed")
	}
	t.Run("canceled-read", func(t *testing.T) {
		canceled, cancel := context.WithCancel(ctx)
		cancel()
		escapedSnapshot, escapedData, e := reader.resourceQueryCapture(canceled)
		if !errors.Is(e, context.Canceled) || !reflect.DeepEqual(escapedSnapshot, graphstore.ResourceSnapshot{}) || !reflect.DeepEqual(escapedData, resourcequery.Dataset{}) {
			t.Fatal("canceled capture returned partial authority")
		}
	})
	if err = store.Close(); err != nil {
		t.Fatal(err)
	}
	t.Run("closed-read", func(t *testing.T) {
		escapedSnapshot, escapedData, e := reader.resourceQueryCapture(ctx)
		if e == nil || ctx.Err() != nil || !reflect.DeepEqual(escapedSnapshot, graphstore.ResourceSnapshot{}) || !reflect.DeepEqual(escapedData, resourcequery.Dataset{}) {
			t.Fatal("closed native read did not fail without partial authority under a live context")
		}
	})
	// The store is now CLOSED. Whole-ID selection and dynamic stored-pin
	// acquisitions use only this Dataset, never a hidden mutable endpoint lookup.
	for _, document := range snapshot.Documents {
		collection := "beads"
		if document.Kind() == graph.KindLink {
			collection = "links"
		}
		currentProducerOne(t, ctx, data, currentProducerRefs(collection, document.ID(), ""), document.NativeJSON())
		if document.Kind() == graph.KindLink {
			// Dynamic stored pins, not forbidden static request pins. This includes
			// actual installed/intrinsic Refines Links and native Dependency/Info Links.
			var endpoints struct {
				Source graphstore.EndpointReference `json:"source"`
				Target graphstore.EndpointReference `json:"target"`
			}
			if err := json.Unmarshal(document.NativeJSON(), &endpoints); err != nil {
				t.Fatal(err)
			}
			for _, side := range []struct {
				name     string
				endpoint graphstore.EndpointReference
			}{{"source", endpoints.Source}, {"target", endpoints.Target}} {
				current, known := editions[side.endpoint.URI]
				if !known || kinds[side.endpoint.URI] != graph.KindBead {
					t.Fatalf("fixture endpoint is not an actual captured Bead: %s", side.endpoint.URI)
				}
				selection := map[string]any{"op": "endpoints", "input": currentProducerRefs("links", document.ID(), ""), "side": side.name}
				if side.endpoint.Revision == "" || side.endpoint.Revision == current.Revision {
					currentProducerOne(t, ctx, data, selection, bodies[side.endpoint.URI])
				} else {
					currentProducerUnsupported(t, ctx, data, selection)
				}
			}
		}
	}
	for _, tc := range []struct{ name, collection, id, revision string }{
		{"static-pin-old-memory-admission", "beads", target.ID, target.Revision},
		{"static-pin-noncurrent-source-admission", "beads", taskA.ID, taskA.Revision},
		{"static-pin-missing-admission", "beads", store.ScopeURL() + "never-created", updated.Memory.Revision},
		{"static-pin-wrong-kind-bead-admission", "beads", floating.Link.ID, floating.Link.Revision},
		{"static-pin-wrong-kind-link-admission", "links", updated.Memory.ID, updated.Memory.Revision},
	} {
		t.Run(tc.name, func(t *testing.T) {
			currentProducerStaticPinRefused(t, currentProducerRefs(tc.collection, tc.id, tc.revision))
		})
	}
	t.Run("current-info-source-pin", func(t *testing.T) {
		if floating.Link.SourceRevision != editions[source.ID].Revision {
			t.Fatal("fixture lacks an actual current stored source pin")
		}
		selection := map[string]any{"op": "endpoints", "input": currentProducerRefs("links", floating.Link.ID, ""), "side": "source"}
		currentProducerOne(t, ctx, data, selection, bodies[source.ID])
	})
	t.Run("current-info-endpoint", func(t *testing.T) {
		selection := map[string]any{"op": "endpoints", "input": currentProducerRefs("links", floating.Link.ID, ""), "side": "target"}
		currentProducerOne(t, ctx, data, selection, bodies[updated.Memory.ID])
	})
	t.Run("retained-info-endpoint-refused", func(t *testing.T) {
		selection := map[string]any{"op": "endpoints", "input": currentProducerRefs("links", pinned.Link.ID, ""), "side": "target"}
		currentProducerUnsupported(t, ctx, data, selection)
	})
	t.Run("current-graph", func(t *testing.T) {
		selection := map[string]any{"op": "walk", "roots": currentProducerRefs("beads", source.ID, ""), "direction": "outgoing", "depth": 1, "bind": "l", "through": map[string]any{"op": "compare", "operator": "eq", "left": map[string]any{"op": "field", "of": map[string]any{"op": "binding", "name": "l"}, "path": []any{"id"}}, "right": map[string]any{"op": "literal", "value": floating.Link.ID}}}
		response, e := resourcequery.Execute(ctx, currentProducerRequest(t, "graph", selection), data, resourcequery.DefaultLimits())
		if e != nil {
			t.Fatal(e)
		}
		graphResult, ok := response.Result.(resourcequery.GraphResult)
		if !ok || !response.Evaluation.Complete || len(graphResult.Beads) != 2 || len(graphResult.Links) != 1 || len(graphResult.Endpoints) != 1 {
			t.Fatalf("current native graph incomplete: %+v", response)
		}
		mapping := graphResult.Endpoints[0]
		if mapping.Link != editions[floating.Link.ID] || mapping.Source != editions[source.ID] || mapping.Target != editions[updated.Memory.ID] || !bytes.Equal(graphResult.Links[0], bodies[floating.Link.ID]) {
			t.Fatal("graph endpoint editions/body escaped the captured context")
		}
		seen := map[string]bool{}
		for _, raw := range graphResult.Beads {
			var header struct {
				ID string `json:"id"`
			}
			if json.Unmarshal(raw, &header) != nil || !bytes.Equal(raw, bodies[header.ID]) {
				t.Fatal("graph Bead body rewritten")
			}
			seen[header.ID] = true
		}
		if !seen[source.ID] || !seen[updated.Memory.ID] {
			t.Fatal("graph endpoint membership changed")
		}
	})
}

func mustCurrentProducerMemory(t *testing.T, record graphstore.Record) []byte {
	t.Helper()
	document, present := record.ResourceDocument()
	if !present {
		t.Fatal("normal native Memory writer lost its same-Tx applied facet")
	}
	return document.NativeJSON()
}

func TestLegacyQueryProducerKeepsCurrentAuthorityAbsent(t *testing.T) {
	ctx, _, store := currentProducerNativeStore(t, graphstore.SchemaVersion)
	memory, err := store.Create(ctx, graphstore.CreateRequest{Path: "beads/legacy", Title: "Legacy", Body: "No opt-in"})
	if err != nil {
		t.Fatal(err)
	}
	reader := New(store)
	_, data, err := reader.queryInventory(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if data.CurrentEditions != nil || data.TypeEditions != nil || len(data.Resources) != 1 {
		t.Fatal("legacy projection gained exact current authority")
	}
	// Acquire once, then close. Pure representation/evaluator checks below do
	// not fabricate trusted current facts from a matching body/sentinel token.
	if err = store.Close(); err != nil {
		t.Fatal(err)
	}
	t.Run("static-pin-matching-body-admission", func(t *testing.T) {
		currentProducerStaticPinRefused(t, currentProducerRefs("beads", memory.ID, memory.Revision))
		currentProducerOne(t, ctx, data, currentProducerRefs("beads", memory.ID, ""), data.Resources[0].Data)
	})
	t.Run("static-pin-current-token-admission", func(t *testing.T) {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(data.Resources[0].Data, &fields); err != nil {
			t.Fatal(err)
		}
		fields["revision"] = json.RawMessage(`"current"`)
		raw, e := json.Marshal(fields)
		if e != nil {
			t.Fatal(e)
		}
		resource, e := resourcequery.NewResource(graph.KindBead, raw)
		if e != nil {
			t.Fatal(e)
		}
		detached := data
		detached.Resources = []resourcequery.Resource{resource}
		if detached.CurrentEditions != nil {
			t.Fatal("pure legacy fixture acquired authority")
		}
		currentProducerStaticPinRefused(t, currentProducerRefs("beads", memory.ID, "current"))
		currentProducerOne(t, ctx, detached, currentProducerRefs("beads", memory.ID, ""), resource.Data)
	})
	t.Run("dynamic-matching-pin-without-authority", func(t *testing.T) {
		// Detached representation-only Link fixture; it grants no native provenance
		// and never calls the trusted-current constructor. Its target pin matches
		// the actual legacy captured Memory body, so nil capability is the boundary.
		fields := map[string]any{"id": store.ScopeURL() + "links/legacy-pin-fixture", "revision": "legacy-link-fixture", "type": graphstore.RelatedTypeURL(store.ScopeURL()), "source": memory.ID, "target": map[string]any{"uri": memory.ID, "revision": memory.Revision}, "properties": map[string]any{}, "metadata": map[string]any{}}
		raw, e := json.Marshal(fields)
		if e != nil {
			t.Fatal(e)
		}
		link, e := resourcequery.NewResource(graph.KindLink, raw)
		if e != nil {
			t.Fatal(e)
		}
		detached := data
		detached.Resources = append(append([]resourcequery.Resource(nil), data.Resources...), link)
		if detached.CurrentEditions != nil {
			t.Fatal("legacy representation fixture gained authority")
		}
		selection := map[string]any{"op": "endpoints", "input": currentProducerRefs("links", store.ScopeURL()+"links/legacy-pin-fixture", ""), "side": "target"}
		currentProducerUnsupported(t, ctx, detached, selection)
		// Control: same valid fixture with floating target works ordinarily. This
		// prevents a malformed Link or missing nominal Type from masking the pin.
		fields["target"] = memory.ID
		raw, e = json.Marshal(fields)
		if e != nil {
			t.Fatal(e)
		}
		link, e = resourcequery.NewResource(graph.KindLink, raw)
		if e != nil {
			t.Fatal(e)
		}
		detached.Resources = append(append([]resourcequery.Resource(nil), data.Resources...), link)
		currentProducerOne(t, ctx, detached, selection, data.Resources[0].Data)
	})
	t.Run("native14-capability-refused", func(t *testing.T) {
		snapshot, escaped, e := reader.resourceQueryCapture(ctx)
		if !errors.Is(e, graphstore.ErrCapabilityUnavailable) || !reflect.DeepEqual(snapshot, graphstore.ResourceSnapshot{}) || !reflect.DeepEqual(escaped, resourcequery.Dataset{}) {
			t.Fatal("legacy native acquisition returned usable exact authority")
		}
	})
}
