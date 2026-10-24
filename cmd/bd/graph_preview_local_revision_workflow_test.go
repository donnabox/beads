//go:build cgo

package main

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func TestGraphPreviewLocalRevisionAndGeneratedRootWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			const scope = "https://example.invalid/local-ordinal/"
			work, home := graphListWorkspace(t, bd, engine, scope)
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			refuse := func(code string, args ...string) {
				t.Helper()
				graphPatchProcess(t, bd, work, home, nil, code, 90*time.Second, append(args, "--json")...)
			}
			one := graphMixedResult[graphstore.Record](t, call("create", "One", "--type", "types/memory", "--id-root", "notes"))
			two := graphMixedResult[graphstore.Record](t, call("create", "Two", "--type", "types/memory"))
			oneID, twoID := strings.TrimPrefix(one.ID, scope), strings.TrimPrefix(two.ID, scope)
			if !strings.HasPrefix(oneID, "notes/") || len(strings.TrimPrefix(oneID, "notes/")) != 6 || len(twoID) != 6 {
				t.Fatalf("canonical generated IDs: %q %q", oneID, twoID)
			}
			linked := graphMixedResult[graphstore.LinkMutationResult](t, call("link", "create", oneID, twoID+"@1", "--type", "types/references", "--id-root", "refs"))
			if !strings.HasPrefix(linked.Link.ID, scope+"refs/") || len(strings.TrimPrefix(linked.Link.ID, scope+"refs/")) != 6 || linked.Link.TargetRevision != two.Revision {
				t.Fatalf("generated Link/full pin: %+v", linked.Link)
			}
			call("update", twoID, "--properties", `{"body":"Changed"}`)
			byNumber := call("show", twoID+"@1")
			if byNumber != call("show", twoID+"@="+two.Revision) || byNumber != call("show", two.ID, "--version", two.Revision) || byNumber == call("show", twoID+"@current") {
				t.Fatal("ordinal, exact token and current selection disagree")
			}
			refuse("revision_unknown", "show", twoID+"@999")
			refuse("revision_unknown", "link", "create", oneID, twoID+"@999", "--type", "types/references")
			canonical := func(selector string) json.RawMessage {
				t.Helper()
				out := graphPolicyCLI(t, bd, work, home, nil, "", "show", selector, "--output-format", "json")
				var envelope struct {
					Data json.RawMessage `json:"data"`
				}
				if err := json.Unmarshal([]byte(out), &envelope); err != nil || len(envelope.Data) == 0 {
					t.Fatalf("canonical read: %s %v", out, err)
				}
				return envelope.Data
			}
			if string(canonical(twoID+"@1")) != string(canonical(twoID+"@="+two.Revision)) {
				t.Fatal("canonical output changed selected revision")
			}
			issue := graphMixedResult[graphstore.IssueRecord](t, call("create", "Native rooted", "--id-root", "work"))
			if !strings.HasPrefix(issue.Properties.ID, "work/") || len(strings.TrimPrefix(issue.Properties.ID, "work/")) != 6 {
				t.Fatalf("native root ID: %+v", issue)
			}
			if call("show", issue.Properties.ID+"@1") != call("show", issue.ID, "--version", issue.Revision) {
				t.Fatal("native ordinal did not resolve authoritative history")
			}
			refuse("invalid_selector", "create", "Bad", "--type", "types/memory", "--id", "explicit", "--id-root", "root")
			refuse("invalid_selector", "link", "create", oneID, twoID, "--type", "types/references", "--id", "explicit", "--id-root", "root")
			refuse("invalid_selector", "show", twoID+"@unpinned")
			base := graphMixedResult[graphstore.Record](t, call("remember", "Base", "--id", "shadow"))
			before := call("show", "shadow@1")
			for _, suffix := range []string{"@1", "@=" + base.Revision, "@current"} {
				literal := graphMixedResult[graphstore.Record](t, call("remember", "Literal", "--id", "shadow"+suffix))
				if got := call("show", "shadow"+suffix); got != before {
					t.Fatalf("literal sibling changed reference meaning: %s", suffix)
				}
				shown := graphMixedResult[graphstore.Record](t, call("show", "--id", "shadow"+suffix))
				if shown.ID != literal.ID || call("show", literal.ID) != call("show", "--id", "shadow"+suffix) {
					t.Fatal("explicit literal escape lost @ identity")
				}
			}
			shadowPin := graphMixedResult[graphstore.LinkMutationResult](t, call("link", "create", oneID, "shadow@1", "--type", "types/references"))
			if shadowPin.Link.Target != base.ID || shadowPin.Link.TargetRevision != base.Revision {
				t.Fatal("literal sibling shadowed endpoint reference")
			}
			gone := graphMixedResult[graphstore.Record](t, call("remember", "Recoverable", "--id", "recovery"))
			call("delete", "recovery", "--force", "--if-revision", gone.Revision)
			history := graphMixedResult[struct {
				CurrentState *graphstore.ResourceTerminalState `json:"currentState"`
			}](t, call("versions", "recovery"))
			if history.CurrentState == nil {
				t.Fatalf("deletion history: %+v", history)
			}
			refuse("gone", "show", "recovery@2")
			call("update", "recovery", "--from-revision", gone.Revision, "--if-revision", history.CurrentState.Revision)
			refuse("gone", "show", "recovery@2")
			if call("show", "recovery@1") != call("show", gone.ID, "--version", gone.Revision) {
				t.Fatal("recovery renumbered old state")
			}
			if call("show", "recovery@3") != call("show", "recovery") {
				t.Fatal("new same-ID revision did not receive ordinal3")
			}
		})
	}
}
