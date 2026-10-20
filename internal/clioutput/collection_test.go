package clioutput_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/clioutput"
)

func collection(t *testing.T, format string, include, omit []string) *clioutput.CollectionRenderer {
	t.Helper()
	renderer, err := clioutput.NewCollection(clioutput.CollectionDescriptor{
		Item:       clioutput.Descriptor{BDPSpec: "fixture-only:collection-v1", Protection: &clioutput.Protection{Required: []string{"id", "revision"}}, Expandable: []string{"/properties", "/metadata"}},
		ResultKind: "resource_collection", ItemsMember: "items", DefaultColumns: []string{"/properties/title"},
	}, clioutput.Options{Formats: []string{format}, Include: include, Omit: omit})
	if err != nil {
		t.Fatal(err)
	}
	return renderer
}

func complete(count int) clioutput.Completion {
	exhausted := true
	return clioutput.Completion{Returned: count, Exhausted: &exhausted, StopReason: "exhausted"}
}

func decode(t *testing.T, raw []byte) map[string]json.RawMessage {
	t.Helper()
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		t.Fatalf("invalid object %s: %v", raw, err)
	}
	return object
}

func TestCollectionPreservesNativeShapeAndProjection(t *testing.T) {
	renderer := collection(t, "json", nil, []string{"/properties/body"})
	payload := json.RawMessage(`{"items":[{"id":"b:1","revision":"r1","properties":{"title":"T","body":"large","n":9007199254740993}},{"id":"b:2","revision":"r2","metadata":null}],"next":{"opaque":9007199254740993},"readContext":{"snapshot":"s1"}}`)
	exhausted := false
	output, err := renderer.Render(payload, clioutput.Completion{Returned: 2, Exhausted: &exhausted, StopReason: "page"})
	if err != nil {
		t.Fatal(err)
	}
	envelope := decode(t, output)
	if string(envelope["schema_version"]) != "2" || string(envelope["result_kind"]) != `"resource_collection"` {
		t.Fatalf("wrong framing: %s", output)
	}
	native := decode(t, envelope["data"])
	if string(native["next"]) != `{"opaque":9007199254740993}` || string(native["readContext"]) != `{"snapshot":"s1"}` {
		t.Fatalf("changed native metadata: %s", output)
	}
	if strings.Contains(string(native["items"]), "large") || !strings.Contains(string(native["items"]), "9007199254740993") || !strings.Contains(string(native["items"]), `"metadata":null`) {
		t.Fatalf("projection changed data: %s", output)
	}
	if string(envelope["projection"]) != `{"omit":["/properties/body"]}` {
		t.Fatalf("wrong projection metadata: %s", output)
	}
}

func TestCollectionRawAndEmpty(t *testing.T) {
	renderer := collection(t, "json", nil, []string{"envelope"})
	out, err := renderer.Render(json.RawMessage(`{"items":[],"next":null,"scope":"all"}`), complete(0))
	if err != nil {
		t.Fatal(err)
	}
	object := decode(t, out)
	if string(object["items"]) != "[]" || string(object["next"]) != "null" || object["schema_version"] != nil {
		t.Fatalf("raw changed native contract: %s", out)
	}
	for _, raw := range []string{`{"items":null}`, `{"items":{}}`, `{}`, `[]`} {
		if out, err := renderer.Render(json.RawMessage(raw), complete(0)); err == nil || len(out) != 0 {
			t.Fatalf("invalid shape emitted: %s %v", out, err)
		}
	}
}

func TestCollectionFailureDoesNotLeakBufferedSuccess(t *testing.T) {
	renderer := collection(t, "json", nil, nil)
	if out, err := renderer.Render(json.RawMessage(`{"items":[{"id":"first","revision":"r"},{"id":"missing-revision"}]}`), complete(2)); err == nil || len(out) != 0 {
		t.Fatalf("partial success: %s %v", out, err)
	}
	problem := json.RawMessage(`{"code":"provider-gone","retry":{"token":9007199254740993}}`)
	out, err := renderer.RenderFailure("bdp", problem)
	if err != nil {
		t.Fatal(err)
	}
	object := decode(t, out)
	if object["data"] != nil || string(object["completion"]) != `{"returned":0,"exhausted":null,"stop_reason":"error"}` {
		t.Fatalf("failure claims partial output: %s", out)
	}
	if string(decode(t, object["error"])["problem"]) != string(problem) {
		t.Fatalf("lost BDP problem: %s", out)
	}
	for _, format := range []string{"detail", "table", "json"} {
		var omit []string
		if format == "json" {
			omit = []string{"envelope"}
		}
		if out, err := collection(t, format, nil, omit).RenderFailure("cli", json.RawMessage(`{"code":"read"}`)); err != nil || len(out) != 0 {
			t.Fatalf("unexpected failure stdout: %s %v", out, err)
		}
	}
}

func TestTableUsesOriginalArrayIndexAndEscapesControls(t *testing.T) {
	renderer := collection(t, "table", []string{"/properties/a/2", "/properties/a/0", "/metadata/x~1y", "/properties/object"}, []string{"/properties/a/0", "/properties/object/secret"})
	out, err := renderer.Render(json.RawMessage(`{"items":[{"id":"b\u001b[31m","revision":"r","properties":{"a":["zero","one","two"],"object":{"secret":"hidden","ok":null}},"metadata":{"x/y":"line\nnext"}}]}`), complete(1))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(out, []byte(`"two"`)) || bytes.Contains(out, []byte(`"zero"`)) || bytes.Contains(out, []byte("\x1b")) || bytes.Contains(out, []byte("hidden")) {
		t.Fatalf("incorrect table: %s", out)
	}
	if !bytes.Contains(out, []byte(`{"ok":null}`)) || !bytes.Contains(out, []byte(`"line\nnext"`)) {
		t.Fatalf("lost structured cell/escaping: %s", out)
	}
}

func TestJSONLFramingAndRawFailure(t *testing.T) {
	for _, raw := range []bool{false, true} {
		t.Run(map[bool]string{false: "framed", true: "raw"}[raw], func(t *testing.T) {
			var omit []string
			if raw {
				omit = []string{"envelope"}
			}
			renderer := collection(t, "jsonl", nil, omit)
			var out bytes.Buffer
			stream, err := renderer.NewJSONL(&out)
			if err != nil {
				t.Fatal(err)
			}
			if err := stream.Item(json.RawMessage(`{"id":"b","revision":"r"}`)); err != nil {
				t.Fatal(err)
			}
			if stream.Returned() != 1 {
				t.Fatal("wrong emitted count")
			}
			if err := stream.Item(json.RawMessage(`{"id":"bad"}`)); err == nil {
				t.Fatal("invalid item accepted")
			}
			if err := stream.Failure("bdp", json.RawMessage(`{"code":"denied","extra":9007199254740993}`)); err != nil {
				t.Fatal(err)
			}
			lines := bytes.Split(bytes.TrimSpace(out.Bytes()), []byte("\n"))
			if raw {
				if len(lines) != 1 || decode(t, lines[0])["record_type"] != nil {
					t.Fatalf("raw acquired exceptional row: %s", out.Bytes())
				}
			} else {
				if len(lines) != 3 {
					t.Fatalf("wrong records: %s", out.Bytes())
				}
				header := decode(t, lines[0])
				terminal := decode(t, lines[2])
				if header["outcome"] != nil || string(header["record_type"]) != `"header"` {
					t.Fatal("header claims outcome")
				}
				if string(terminal["outcome"]) != `"failure"` || string(terminal["completion"]) != `{"returned":1,"exhausted":null,"stop_reason":"error"}` {
					t.Fatalf("wrong terminal: %s", lines[2])
				}
			}
			before := out.Len()
			if err := stream.Finish(true, "exhausted", nil); err == nil || out.Len() != before {
				t.Fatal("duplicate terminal accepted")
			}
		})
	}
}

func TestJSONLEmptyAndOpaqueContinuation(t *testing.T) {
	renderer := collection(t, "jsonl", nil, nil)
	var out bytes.Buffer
	stream, err := renderer.NewJSONL(&out)
	if err != nil {
		t.Fatal(err)
	}
	token := json.RawMessage(`{"token":[null,9007199254740993,"opaque"]}`)
	if err := stream.Finish(false, "page", token); err != nil {
		t.Fatal(err)
	}
	lines := bytes.Split(bytes.TrimSpace(out.Bytes()), []byte("\n"))
	if len(lines) != 2 {
		t.Fatalf("empty missing framing: %s", out.Bytes())
	}
	if string(decode(t, lines[1])["continuation"]) != string(token) {
		t.Fatal("changed opaque token")
	}
}

type failAfterHeader struct{ calls int }

func (w *failAfterHeader) Write(p []byte) (int, error) {
	w.calls++
	if w.calls > 1 {
		return len(p) / 2, nil
	}
	return len(p), nil
}

func TestJSONLFailedWritePoisonsStreamAndDoesNotCount(t *testing.T) {
	writer := &failAfterHeader{}
	stream, err := collection(t, "jsonl", nil, nil).NewJSONL(writer)
	if err != nil {
		t.Fatal(err)
	}
	if err := stream.Item(json.RawMessage(`{"id":"b","revision":"r"}`)); !errors.Is(err, io.ErrShortWrite) {
		t.Fatalf("short write not reported: %v", err)
	}
	if stream.Returned() != 0 {
		t.Fatal("partial item counted")
	}
	if err := stream.Failure("cli", json.RawMessage(`{"code":"write"}`)); !errors.Is(err, io.ErrShortWrite) {
		t.Fatalf("failed stream resumed: %v", err)
	}
	if writer.calls != 2 {
		t.Fatal("wrote terminal after partial line")
	}
}

func TestCollectionCompletionValidation(t *testing.T) {
	renderer := collection(t, "json", nil, nil)
	payload := json.RawMessage(`{"items":[]}`)
	yes, no := true, false
	for _, c := range []clioutput.Completion{
		{Returned: 1, Exhausted: &yes, StopReason: "exhausted"},
		{Exhausted: nil, StopReason: "page"},
		{Exhausted: &no, StopReason: "exhausted"},
		{Exhausted: &yes, StopReason: "limit"},
		{Exhausted: &no, StopReason: "error"},
	} {
		if out, err := renderer.Render(payload, c); err == nil || len(out) != 0 {
			t.Fatalf("invalid completion emitted %s %v", out, err)
		}
	}
}

func TestPagingHonorsProviderAndDeliveredBoundary(t *testing.T) {
	paged, err := clioutput.ResolvePaging(clioutput.PagingOptions{}, "json")
	if err != nil {
		t.Fatal(err)
	}
	if size, _ := paged.RequestSize(0); size != 50 {
		t.Fatal("wrong default page size")
	}
	next := json.RawMessage(`{"cursor":"after-page"}`)
	plan, err := paged.PlanPage(0, 2, false, next, nil)
	if err != nil || plan.StopReason != "page" || plan.Exhausted {
		t.Fatalf("short page invented exhaustion: %+v %v", plan, err)
	}
	unpaged, err := clioutput.ResolvePaging(clioutput.PagingOptions{Limit: 3}, "jsonl")
	if err != nil {
		t.Fatal(err)
	}
	if size, _ := unpaged.RequestSize(1); size != 2 {
		t.Fatal("limit not applied to request")
	}
	plan, err = unpaged.PlanPage(1, 5, false, next, nil)
	if err != nil || plan.Take != 2 || plan.StopReason != "limit" || plan.Continuation != nil || !plan.ContinuationUnavailable {
		t.Fatalf("unsafe truncated-page token: %+v %v", plan, err)
	}
	boundary := json.RawMessage(`{"cursor":"after-delivered"}`)
	plan, err = unpaged.PlanPage(1, 5, false, next, boundary)
	if err != nil || string(plan.Continuation) != string(boundary) || plan.ContinuationUnavailable {
		t.Fatalf("lost exact boundary: %+v %v", plan, err)
	}
	plan, err = unpaged.PlanPage(1, 2, true, nil, nil)
	if err != nil || plan.StopReason != "exhausted" || !plan.Exhausted {
		t.Fatalf("exhaustion at cap: %+v %v", plan, err)
	}
	all, _ := clioutput.ResolvePaging(clioutput.PagingOptions{}, "jsonl")
	plan, err = all.PlanPage(0, 1, false, next, nil)
	if err != nil || plan.StopReason != "" {
		t.Fatal("JSONL did not default to all")
	}
	for _, options := range []clioutput.PagingOptions{{Paged: true, Unpaged: true}, {Limit: -1}, {PageSize: -1}} {
		if _, err := clioutput.ResolvePaging(options, "json"); err == nil {
			t.Fatal("invalid paging flags accepted")
		}
	}
}

func TestVersionCollectionKeepsSubjectAndExactVersion(t *testing.T) {
	descriptor := clioutput.CollectionDescriptor{
		Item:       clioutput.Descriptor{BDPSpec: "fixture-only:versions-v1", Protection: &clioutput.Protection{Required: []string{"version"}}},
		ResultKind: "version_collection", ItemsMember: "versions", DefaultColumns: []string{"/actor"},
	}
	renderer, err := clioutput.NewCollection(descriptor, clioutput.Options{Formats: []string{"json"}, Include: []string{"/actor"}})
	if err != nil {
		t.Fatal(err)
	}
	out, err := renderer.Render(json.RawMessage(`{"resource":"link:1","kind":"link","versions":[{"version":"exact:1","local_revision":4,"actor":"alice"}]}`), complete(1))
	if err != nil {
		t.Fatal(err)
	}
	native := decode(t, decode(t, out)["data"])
	if string(native["resource"]) != `"link:1"` || string(native["kind"]) != `"link"` || string(native["versions"]) != `[{"actor":"alice","version":"exact:1"}]` {
		t.Fatalf("changed version semantics: %s", out)
	}
	if _, err := clioutput.NewCollection(descriptor, clioutput.Options{Formats: []string{"jsonl"}}); err == nil {
		t.Fatal("unadmitted subject-losing stream accepted")
	}
}

func TestCollectionFormatConflictsBeforeExecution(t *testing.T) {
	descriptor := clioutput.CollectionDescriptor{Item: clioutput.Descriptor{BDPSpec: "fixture-only:collection-v1", Protection: &clioutput.Protection{Required: []string{"id"}}}, ResultKind: "resource_collection", ItemsMember: "items"}
	for _, options := range []clioutput.Options{
		{Formats: []string{"json", "table"}},
		{Formats: []string{"default"}, JSON: true},
		{Formats: []string{"jsonl"}, JSON: true},
		{Formats: []string{"tree"}},
		{Formats: []string{"table"}, Omit: []string{"envelope"}},
		{Formats: []string{"json"}, Omit: []string{"/id"}},
	} {
		if _, err := clioutput.NewCollection(descriptor, options); err == nil {
			t.Fatalf("invalid options accepted: %+v", options)
		}
	}
	for _, options := range []clioutput.Options{
		{Formats: []string{"json", "json"}, JSON: true},
		{Formats: []string{"default", "table"}},
		{Formats: []string{"jsonl"}, Omit: []string{"envelope"}},
	} {
		if _, err := clioutput.NewCollection(descriptor, options); err != nil {
			t.Fatalf("equivalent options refused: %v", err)
		}
	}
}

func TestTableKeepsEmbeddedContextInProjectedObjectCell(t *testing.T) {
	descriptor := clioutput.CollectionDescriptor{
		Item: clioutput.Descriptor{
			BDPSpec: "fixture-only:embedded-v1",
			Protection: &clioutput.Protection{Required: []string{"id"}, Members: map[string]*clioutput.Protection{
				"ownedLinks": {Items: &clioutput.Protection{Required: []string{"id", "source", "target", "revision"}}},
			}},
			Expandable: []string{"/ownedLinks", "/properties"},
		},
		ResultKind: "resource_collection", ItemsMember: "items",
	}
	renderer, err := clioutput.NewCollection(descriptor, clioutput.Options{
		Formats: []string{"table"},
		Include: []string{"/ownedLinks/1/properties/title", "/ownedLinks/1/properties/color", "/properties/title"},
	})
	if err != nil {
		t.Fatal(err)
	}
	out, err := renderer.Render(json.RawMessage(`{"items":[{"id":"b","properties":{"title":"Bead title","body":"hidden"},"ownedLinks":[{"id":"skip","source":"b","target":"x","revision":"l0"},{"id":"keep","source":"b","target":"y","revision":"l1","properties":{"title":"Link title","color":"blue","body":"hidden"}}]}]}`), complete(1))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Count(out, []byte(`"/ownedLinks/1"`)) != 1 || bytes.Contains(out, []byte("/ownedLinks/1/properties")) || !bytes.Contains(out, []byte(`"/properties/title"`)) {
		t.Fatalf("wrong protected column roots: %s", out)
	}
	for _, fragment := range []string{`"id":"keep"`, `"source":"b"`, `"target":"y"`, `"revision":"l1"`, `"title":"Link title"`, `"color":"blue"`, `"Bead title"`} {
		if !bytes.Contains(out, []byte(fragment)) {
			t.Fatalf("missing %s: %s", fragment, out)
		}
	}
	if bytes.Contains(out, []byte("hidden")) || bytes.Contains(out, []byte(`"skip"`)) {
		t.Fatalf("table widened selection: %s", out)
	}
}

func TestCollectionHumanKeepsDeclaredSubjectContext(t *testing.T) {
	descriptor := clioutput.CollectionDescriptor{Item: clioutput.Descriptor{BDPSpec: "fixture-only:versions-v1", Protection: &clioutput.Protection{Required: []string{"version"}}}, ResultKind: "version_collection", ItemsMember: "versions", ContextMembers: []string{"resource", "kind"}}
	for _, format := range []string{"json", "table", "detail"} {
		renderer, err := clioutput.NewCollection(descriptor, clioutput.Options{Formats: []string{format}})
		if err != nil {
			t.Fatal(err)
		}
		out, err := renderer.Render(json.RawMessage(`{"resource":"link:1","kind":"link","versions":[{"version":"v1"}]}`), complete(1))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Contains(out, []byte(`"link:1"`)) || !bytes.Contains(out, []byte(`"v1"`)) {
			t.Fatalf("%s lost subject context: %s", format, out)
		}
		if out, err := renderer.Render(json.RawMessage(`{"versions":[]}`), complete(0)); err == nil || len(out) != 0 {
			t.Fatal("missing enclosing context accepted")
		}
	}
}
