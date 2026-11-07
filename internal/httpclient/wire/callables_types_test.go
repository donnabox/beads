package wire

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

// The Types rule (specs/types/issue/callables.md §2) is the third thing this layer
// cancels out: the layers above send and read v0's `issue_type` and short names, and
// the wire carries `type` and Type references.

const (
	wireIssueTypes = "https://github.com/gastownhall/bdp/types/issue/"
	wireLinkTypes  = "https://github.com/gastownhall/bdp/types/dependency/"
)

func TestTheClientSpellsArgumentTypesAsReferences(t *testing.T) {
	cases := []struct {
		name string
		r    Request
		want map[string]string
	}{
		{"a create's type and its edges' kinds",
			Request{Op: OpCreateIssue, Method: http.MethodPost, Path: PathIssues, Body: map[string]any{
				"title": "t", "issue_type": "bug", "dependencies": []map[string]any{{"target_id": "bd-2", "type": "tracks"}},
			}},
			map[string]string{"type": `"` + wireIssueTypes + `bug"`, "dependencies": `[{"target_id":"bd-2","type":"` + wireLinkTypes + `tracks"}]`}},
		{"a patch's type, under its own name",
			Request{Op: OpUpdateIssue, Method: http.MethodPatch, Path: "/x", PathArg: "bd-1", Body: map[string]any{
				"actor": "sam", "patch": map[string]any{"issue_type": "epic"},
			}},
			map[string]string{"patch": `{"type":"` + wireIssueTypes + `epic"}`}},
		{"a ready filter expands an alias, as v0 did, and splits a list",
			Request{Op: OpListReadyWork, Method: http.MethodGet, Path: PathReady, Query: url.Values{"type": {"feat"}, "exclude_type": {"bug, mr", "chore"}}},
			map[string]string{"type": `"` + wireIssueTypes + `feature"`, "exclude_type": `["` + wireIssueTypes + `bug","SCOPEtypes/issue/merge-request","` + wireIssueTypes + `chore"]`}},
		{"a count's type is exact, as v0's is",
			Request{Op: OpCountIssues, Method: http.MethodGet, Path: "/x", Query: url.Values{"type": {"feat"}}},
			map[string]string{"type": `"SCOPEtypes/issue/feat"`}},
		{"an edge filter is a Link Type, and an unknown kind is the Scope's",
			Request{Op: OpListDependencies, Method: http.MethodGet, Path: PathDependencies, Query: url.Values{"issue_id": {"bd-1"}, "type": {"blocks", "my-kind"}}},
			map[string]string{"type": `["` + wireLinkTypes + `blocks","SCOPEtypes/dependency/my-kind"]`}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var sent map[string]json.RawMessage
			var scope string
			c := wireClient(t, Options{}, func(w http.ResponseWriter, req *http.Request) {
				scope = "http://" + req.Host + "/"
				var invocation struct {
					Arguments map[string]json.RawMessage `json:"arguments"`
				}
				body, _ := io.ReadAll(req.Body)
				if err := json.Unmarshal(body, &invocation); err != nil {
					t.Errorf("%s: %v", body, err)
				}
				sent = invocation.Arguments
				_, _ = io.WriteString(w, `{"result":{}}`)
			})
			if err := c.Do(ctx(t), tc.r, &map[string]any{}); err != nil {
				t.Fatalf("Do: %v", err)
			}
			if _, has := sent["issue_type"]; has {
				t.Errorf("issue_type went on the wire: %v", sent)
			}
			for name, want := range tc.want {
				want = strings.ReplaceAll(want, "SCOPE", scope)
				var got, wanted any
				_ = json.Unmarshal(sent[name], &got)
				_ = json.Unmarshal([]byte(want), &wanted)
				if !reflect.DeepEqual(got, wanted) {
					t.Errorf("%s = %s, want %s", name, sent[name], want)
				}
			}
		})
	}
}

func TestTheClientReadsResultTypesBackToNames(t *testing.T) {
	var scope string
	answer := ""
	c := wireClient(t, Options{}, func(w http.ResponseWriter, req *http.Request) {
		scope = "http://" + req.Host + "/"
		_, _ = io.WriteString(w, strings.ReplaceAll(answer, "SCOPE", scope))
	})

	// An issue's dependencies are the related issues, each with its own Type and the
	// edge's dependency_type, which the rule leaves as v0 spells it.
	answer = `{"result":{"id":"bd-1","title":"t","type":"` + wireIssueTypes + `bug","dependencies":[{"id":"bd-2","type":"SCOPEtypes/issue/my-kind","dependency_type":"blocks"}]}}`
	var issue map[string]any
	if err := c.Do(ctx(t), Request{Op: OpGetIssue, Method: http.MethodGet, Path: "/x", PathArg: "bd-1"}, &issue); err != nil {
		t.Fatalf("Do: %v", err)
	}
	if issue["issue_type"] != "bug" || issue["type"] != nil {
		t.Errorf("issue = %v, want issue_type bug and no type", issue)
	}
	deps, _ := issue["dependencies"].([]any)
	if len(deps) != 1 || deps[0].(map[string]any)["issue_type"] != "my-kind" || deps[0].(map[string]any)["dependency_type"] != "blocks" {
		t.Errorf("dependencies = %v, want the Scope's Type as its name and dependency_type as it was", deps)
	}

	answer = `{"result":{"items":[{"issue_id":"bd-1","depends_on_id":"bd-2","type":"SCOPEtypes/dependency/my-kind"}],"missing":[]}}`
	var edges map[string]any
	if err := c.Do(ctx(t), Request{Op: OpListDependencies, Method: http.MethodGet, Path: PathDependencies, Query: url.Values{"issue_id": {"bd-1"}}}, &edges); err != nil {
		t.Fatalf("Do: %v", err)
	}
	if items, _ := edges["items"].([]any); len(items) != 1 || items[0].(map[string]any)["type"] != "my-kind" {
		t.Errorf("edges = %v, want the Scope's kind as its name", edges)
	}

	answer = `{"result":{"id":"bd-1","title":"t","type":"SCOPEtypes/issue/widget"}}`
	if err := c.Do(ctx(t), Request{Op: OpGetIssue, Method: http.MethodGet, Path: "/x", PathArg: "bd-1"}, &issue); err != nil {
		t.Fatalf("Do: %v", err)
	}
	if issue["issue_type"] != "widget" {
		t.Errorf("a configured Type read back as %v", issue["issue_type"])
	}

	// A reference this client cannot spell is an error, never a guess.
	for _, foreign := range []string{"https://elsewhere.test/types/issue/widget", wireIssueTypes + "issue", wireLinkTypes + "blocks"} {
		answer = `{"result":{"id":"bd-1","title":"t","type":"` + foreign + `"}}`
		err := c.Do(ctx(t), Request{Op: OpGetIssue, Method: http.MethodGet, Path: "/x", PathArg: "bd-1"}, &issue)
		if err == nil || !strings.Contains(err.Error(), foreign) {
			t.Errorf("%s: err = %v, want a refusal naming it", foreign, err)
		}
	}
}

func TestTheClientReadsAProblemsTypesBackToNames(t *testing.T) {
	body := `{"status":409,"title":"Conflict","code":"not_closable","blockers":[{"id":"bd-2","kind":"local","type":"` + wireLinkTypes + `waits-for"}]}`
	got, err := typeNames(OpCloseIssue, "http://127.0.0.1:1/", []byte(body), false)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"status":409,"title":"Conflict","code":"not_closable","blockers":[{"id":"bd-2","kind":"local","type":"waits-for"}]}`
	if string(got) != want {
		t.Errorf("problem = %s, want %s", got, want)
	}
}
