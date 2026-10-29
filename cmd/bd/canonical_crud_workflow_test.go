//go:build cgo

package main

import (
	"encoding/json"
	"reflect"
	"testing"
)

// This journey uses the candidate binary and a fresh embedded workspace. It
// proves the ordinary CLI edits reach checked storage transactions; pure
// lowering tests alone cannot establish rollback, revision or owner behavior.
func TestCanonicalCRUDObjectEditsEmbeddedWorkflow(t *testing.T) {
	isolateBeadsDirForTest(t)
	bd := buildBDUnderTest(t)
	const scope = "https://example.invalid/canonical-crud/"
	work, home := graphListWorkspace(t, bd, "embedded", scope)
	call := func(args ...string) string {
		t.Helper()
		return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
	}
	refuse := func(code string, args ...string) {
		t.Helper()
		graphPolicyCLI(t, bd, work, home, nil, code, append(args, "--json")...)
	}
	type resource struct {
		ID         string                     `json:"id"`
		Revision   string                     `json:"revision"`
		Type       json.RawMessage            `json:"type"`
		Source     json.RawMessage            `json:"source,omitempty"`
		Target     json.RawMessage            `json:"target,omitempty"`
		OwnedLinks json.RawMessage            `json:"ownedLinks,omitempty"`
		Properties map[string]json.RawMessage `json:"properties"`
		Metadata   map[string]json.RawMessage `json:"metadata"`
	}
	type mutation struct {
		Memory  resource `json:"memory"`
		Link    resource `json:"link"`
		Changed bool     `json:"changed"`
	}
	show := func(id string) resource {
		t.Helper()
		return graphMixedResult[resource](t, call("show", id))
	}
	showLink := func(id string) resource {
		t.Helper()
		return graphMixedResult[resource](t, call("link", "show", id))
	}
	equal := func(got, want resource) {
		t.Helper()
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("Resource changed unexpectedly:\n got %+v\nwant %+v", got, want)
		}
	}
	value := func(fields map[string]json.RawMessage, key, want string) {
		t.Helper()
		if string(fields[key]) != want {
			t.Fatalf("%q = %s, want %s", key, fields[key], want)
		}
	}
	changed := func(before, after resource, didChange bool) {
		t.Helper()
		if !didChange || after.ID != before.ID || after.Revision == "" || after.Revision == before.Revision {
			t.Fatalf("missing changed Resource postimage: before=%+v after=%+v changed=%t", before, after, didChange)
		}
	}

	// Initialize both object planes through canonical create; show returns the
	// same native record, rather than a client reconstruction of authored input.
	source := graphMixedResult[resource](t, call("create", "--type", "memory", "--id", "source",
		"--properties", `{"title":"Original","body":"Original body"}`,
		"--metadata", `{"keep":true,"nested":{"flag":false}}`))
	target := graphMixedResult[resource](t, call("create", "--type", "memory", "--id", "target",
		"--properties", `{"title":"Target","body":"Untouched"}`))
	if source.ID != scope+"source" || target.ID != scope+"target" || source.Revision == "" {
		t.Fatalf("create lost canonical identities: source=%+v target=%+v", source, target)
	}
	equal(show(source.ID), source)
	equal(show(target.ID), target)

	// Properties choose keyed editing while metadata chooses strict nested
	// patching. JSON Pointer escaping addresses one literal top-level key.
	first := graphMixedResult[mutation](t, call("update", source.ID,
		"--set-property", "title=Reviewed",
		"--patch-metadata", `[{"op":"replace","path":"/nested/flag","value":true},{"op":"add","path":"/a~1b~0c","value":7}]`,
		"--if-revision", "="+source.Revision))
	changed(source, first.Memory, first.Changed)
	value(first.Memory.Properties, "title", `"Reviewed"`)
	value(first.Memory.Properties, "body", `"Original body"`)
	value(first.Memory.Metadata, "keep", "true")
	value(first.Memory.Metadata, "nested", `{"flag":true}`)
	value(first.Memory.Metadata, "a/b~c", "7")
	equal(show(source.ID), first.Memory)

	// Explicit empty patches are admitted and mint no Resource revision. A
	// stale guard and a failed strict operation roll back both object planes.
	noop := graphMixedResult[mutation](t, call("update", source.ID, "--patch-properties", "[]", "--patch-metadata", "[]", "--if-revision", "="+first.Memory.Revision))
	if noop.Changed {
		t.Fatal("empty object patches reported a change")
	}
	equal(noop.Memory, first.Memory)
	refuse("revision_conflict", "update", source.ID, "--set-property", "body=must not publish", "--replace-metadata", `{"bad":true}`, "--if-revision", "="+source.Revision)
	equal(show(source.ID), first.Memory)
	refuse("invalid_properties", "update", source.ID, "--set-property", "title=must not publish", "--patch-metadata", `[{"op":"remove","path":"/missing"}]`, "--if-revision", "="+first.Memory.Revision)
	equal(show(source.ID), first.Memory)
	refuse("invalid_properties", "update", source.ID, "--properties", "{}", "--set-property", "title=conflicting mode")
	equal(show(source.ID), first.Memory)

	replaced := graphMixedResult[mutation](t, call("update", source.ID,
		"--patch-properties", `[{"op":"replace","path":"/body","value":"New body"}]`,
		"--replace-metadata", `{"only":true}`, "--if-revision", "="+first.Memory.Revision))
	changed(first.Memory, replaced.Memory, replaced.Changed)
	value(replaced.Memory.Properties, "title", `"Reviewed"`)
	value(replaced.Memory.Properties, "body", `"New body"`)
	if len(replaced.Memory.Metadata) != 1 {
		t.Fatalf("replacement merged old metadata: %+v", replaced.Memory.Metadata)
	}
	value(replaced.Memory.Metadata, "only", "true")

	// Informational Link properties admit extension names. Repeated sets use
	// the final value; absent unsets are idempotent, including escaped keys.
	createdLink := graphMixedResult[mutation](t, call("link", "create", "--from", "source", "--to", "target", "--type", "references", "--id", "citation",
		"--properties", `{"note":"Original note"}`, "--metadata", `{"keep":true,"nested":{"old":1}}`))
	if createdLink.Link.ID != scope+"citation" || createdLink.Link.Revision == "" {
		t.Fatalf("Link create lost identity: %+v", createdLink)
	}
	equal(showLink(createdLink.Link.ID), createdLink.Link)
	editedLink := graphMixedResult[mutation](t, call("link", "update", "citation",
		"--set-property", "a/b~c=first", "--set-property", "a/b~c=last=value", "--set-property", "note=Reviewed note",
		"--unset-property", "absent", "--unset-property", "absent", "--metadata", `{"nested":{"new":2}}`,
		"--if-revision", "="+createdLink.Link.Revision))
	changed(createdLink.Link, editedLink.Link, editedLink.Changed)
	value(editedLink.Link.Properties, "a/b~c", `"last=value"`)
	value(editedLink.Link.Properties, "note", `"Reviewed note"`)
	if _, present := editedLink.Link.Properties["absent"]; present {
		t.Fatal("unset absent property materialized a value")
	}
	value(editedLink.Link.Metadata, "keep", "true")
	value(editedLink.Link.Metadata, "nested", `{"new":2}`)

	removedKey := graphMixedResult[mutation](t, call("link", "update", "citation",
		"--unset-property", "a/b~c", "--set-property", "a/b~c=discarded",
		"--patch-metadata", `[{"op":"replace","path":"/nested/new","value":3}]`, "--if-revision", "="+editedLink.Link.Revision))
	changed(editedLink.Link, removedKey.Link, removedKey.Changed)
	if _, present := removedKey.Link.Properties["a/b~c"]; present {
		t.Fatal("flag order changed set-then-unset semantics")
	}
	value(removedKey.Link.Metadata, "nested", `{"new":3}`)
	ownerBeforeNoop := show(source.ID)
	linkNoop := graphMixedResult[mutation](t, call("link", "update", "citation", "--patch-properties", "[]", "--patch-metadata", "[]", "--if-revision", "="+removedKey.Link.Revision))
	if linkNoop.Changed {
		t.Fatal("empty Link patches minted a version")
	}
	equal(linkNoop.Link, removedKey.Link)
	equal(show(source.ID), ownerBeforeNoop)
	refuse("revision_conflict", "link", "update", "citation", "--set-property", "note=must not publish", "--replace-metadata", "{}", "--if-revision", "="+createdLink.Link.Revision)
	equal(showLink(createdLink.Link.ID), removedKey.Link)
	equal(show(source.ID), ownerBeforeNoop)

	// Deletion previews are reads. Stale deletion guards preserve the live
	// records, and forced deletion retains a distinct terminal revision.
	linkPreview := graphMixedResult[struct {
		Preview bool     `json:"preview"`
		Link    resource `json:"link"`
	}](t, call("link", "delete", "citation", "--if-revision", "="+removedKey.Link.Revision))
	if !linkPreview.Preview {
		t.Fatal("Link delete did not return its observational preview")
	}
	equal(linkPreview.Link, removedKey.Link)
	equal(showLink(createdLink.Link.ID), removedKey.Link)
	refuse("revision_conflict", "link", "delete", "citation", "--force", "--if-revision", "="+createdLink.Link.Revision)
	equal(showLink(createdLink.Link.ID), removedKey.Link)
	deletedLink := graphMixedResult[mutation](t, call("link", "delete", "citation", "--force", "--if-revision", "="+removedKey.Link.Revision))
	changed(removedKey.Link, deletedLink.Link, deletedLink.Changed)
	refuse("gone", "link", "show", "citation")

	currentSource := show(source.ID)
	type memoryDelete struct {
		Memory   resource `json:"memory"`
		Terminal *struct {
			Revision string `json:"revision"`
		} `json:"terminal"`
		Preview bool `json:"preview"`
		Deleted bool `json:"deleted"`
	}
	preview := graphMixedResult[memoryDelete](t, call("delete", source.ID, "--if-revision", "="+currentSource.Revision))
	if !preview.Preview || preview.Deleted || preview.Terminal != nil {
		t.Fatalf("delete preview mutated state: %+v", preview)
	}
	equal(preview.Memory, currentSource)
	equal(show(source.ID), currentSource)
	refuse("revision_conflict", "delete", source.ID, "--force", "--if-revision", "="+source.Revision)
	equal(show(source.ID), currentSource)
	deletedSource := graphMixedResult[memoryDelete](t, call("delete", source.ID, "--force", "--if-revision", "="+currentSource.Revision))
	if !deletedSource.Deleted || deletedSource.Preview || deletedSource.Terminal == nil || deletedSource.Terminal.Revision == "" || deletedSource.Terminal.Revision == currentSource.Revision {
		t.Fatalf("missing terminal deletion outcome: %+v", deletedSource)
	}
	equal(deletedSource.Memory, currentSource)
	refuse("gone", "show", source.ID)
	equal(show(target.ID), target)
}
