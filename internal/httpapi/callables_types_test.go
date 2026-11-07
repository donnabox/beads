package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/issueops"
)

// The Types rule of specs/types/issue/callables.md §2 on the mount: a Callable
// carries a Bead's Type or a dependency's kind as `type`, valued with a Type
// reference, and v0 sees its own names.

const (
	ruleBug     = "https://github.com/gastownhall/bdp/types/issue/bug"
	ruleTask    = "https://github.com/gastownhall/bdp/types/issue/task"
	ruleRelated = "https://github.com/gastownhall/bdp/types/dependency/related"
)

func TestCallableTypesRuleReadsArgumentsBackToNames(t *testing.T) {
	created := createdIssue("bd-1")
	created.IssueType = types.TypeBug
	lifecycle := &roleLifecycle{createResult: issueops.CreateResult{Issue: created}}
	ts := newBDPServer(t, rolesConfig(Config{Lifecycle: lifecycle}))

	result, _ := callResult(t, ts.callAs(t, "issue/create", "sam", `{"title":"t","type":"`+ruleBug+`"}`))
	reqs := lifecycle.createRequests()
	if len(reqs) != 1 || reqs[0].Issue.IssueType != types.TypeBug {
		t.Fatalf("the role saw %+v, want one create of a bug", reqs)
	}
	var got struct{ Issue map[string]any }
	if err := json.Unmarshal(result, &got.Issue); err != nil {
		t.Fatal(err)
	}
	if got.Issue["type"] != ruleBug {
		t.Errorf("result issue type = %v, want %s", got.Issue["type"], ruleBug)
	}
	if _, has := got.Issue["issue_type"]; has {
		t.Errorf("result issue still carries issue_type: %v", got.Issue)
	}

	// The service default is the task reference, and reaches v0 as task.
	callResult(t, ts.callAs(t, "issue/create", "sam", `{"title":"t"}`))
	if reqs := lifecycle.createRequests(); len(reqs) != 2 || reqs[1].Issue.IssueType != types.TypeTask {
		t.Errorf("an omitted type reached the role as %+v", reqs[len(reqs)-1].Issue.IssueType)
	}
}

func TestCallableTypesRuleRefusesWhatNamesNoType(t *testing.T) {
	cases := []struct {
		name, arguments, location, words string
	}{
		{"a short name", `{"title":"t","type":"bug"}`, "/arguments/type", "is not a Type reference"},
		{"an alias", `{"title":"t","type":"feat"}`, "/arguments/type", "is not a Type reference"},
		{"no Type of the package", `{"title":"t","type":"https://github.com/gastownhall/bdp/types/issue/bugs"}`, "/arguments/type", "has no Type `bugs`"},
		{"the abstract base", `{"title":"t","type":"https://github.com/gastownhall/bdp/types/issue/issue"}`, "/arguments/type", "abstract"},
		{"a Link Type", `{"title":"t","type":"https://github.com/gastownhall/bdp/types/dependency/blocks"}`, "/arguments/type", "Link Type"},
		{"an unconfigured local name", `{"title":"t","type":"SCOPE/types/issue/widget"}`, "/arguments/type", "not a Type this Scope configures"},
		{"another Scope's Type", `{"title":"t","type":"https://elsewhere.test/types/issue/widget"}`, "/arguments/type", "Issue and Dependency Types only"},
		{"not a string", `{"title":"t","type":7}`, "/arguments/type", "is a string"},
		{"a dependency's kind in an edge", `{"title":"t","dependencies":[{"target_id":"bd-2","type":"blocks"}]}`, "/arguments/dependencies/0/type", "is not a Type reference"},
		{"v0's spelling of the member", `{"title":"t","issue_type":"bug"}`, "/arguments/issue_type", "not declared parameters"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			lifecycle := &roleLifecycle{createResult: issueops.CreateResult{Issue: createdIssue("bd-1")}}
			ts := newBDPServer(t, rolesConfig(Config{Lifecycle: lifecycle}))
			arguments := strings.ReplaceAll(tc.arguments, "SCOPE", ts.base)
			body := wantProblem(t, ts.callAs(t, "issue/create", "sam", arguments), http.StatusUnprocessableEntity, bdpValidationFailed)
			if !strings.Contains(body["detail"].(string), tc.words) {
				t.Errorf("detail = %q, want it to say %q", body["detail"], tc.words)
			}
			diagnostics, _ := body["diagnostics"].([]any)
			if len(diagnostics) != 1 || diagnostics[0].(map[string]any)["instanceLocation"] != tc.location {
				t.Errorf("diagnostics = %v, want one at %s", diagnostics, tc.location)
			}
			if reqs := lifecycle.createRequests(); len(reqs) != 0 {
				t.Errorf("the role ran: %+v", reqs)
			}
		})
	}

	t.Run("issue_type in a patch", func(t *testing.T) {
		lifecycle := &roleLifecycle{updateResult: issueops.UpdateResult{Issue: updatedIssue("bd-1"), Changed: true}}
		ts := newBDPServer(t, rolesConfig(Config{Lifecycle: lifecycle}))
		body := wantProblem(t, ts.callAs(t, "issue/update", "sam", `{"id":"bd-1","patch":{"issue_type":"bug"}}`), http.StatusUnprocessableEntity, bdpValidationFailed)
		wantDiagnosticNaming(t, body, "issue_type")
		if reqs := lifecycle.updateRequests(); len(reqs) != 0 {
			t.Errorf("the old spelling reached the role: %+v", reqs)
		}
		callResult(t, ts.callAs(t, "issue/update", "sam", `{"id":"bd-1","patch":{"type":"`+ruleBug+`"}}`))
		reqs := lifecycle.updateRequests()
		if len(reqs) != 1 || !reqs[0].Patch.IssueType.Set || reqs[0].Patch.IssueType.Value != types.TypeBug {
			t.Errorf("the patch reached the role as %+v", reqs)
		}
	})
}

func TestCallableTypesRuleSpellsAConfiguredTypeUnderTheScope(t *testing.T) {
	created := createdIssue("bd-1")
	created.IssueType = "widget"
	lifecycle := &roleLifecycle{createResult: issueops.CreateResult{Issue: created}}
	ts := newBDPServer(t, rolesConfig(Config{Lifecycle: lifecycle, Settings: &roleSettings{value: "widget"}}))
	local := ts.base + "/types/issue/widget"

	result, _ := callResult(t, ts.callAs(t, "issue/create", "sam", `{"title":"t","type":"`+local+`"}`))
	if reqs := lifecycle.createRequests(); len(reqs) != 1 || reqs[0].Issue.IssueType != "widget" {
		t.Fatalf("the role saw %+v", reqs)
	}
	var got struct{ Issue map[string]any }
	if err := json.Unmarshal(result, &got.Issue); err != nil {
		t.Fatal(err)
	}
	if got.Issue["type"] != local {
		t.Errorf("result type = %v, want %s", got.Issue["type"], local)
	}
}

func TestCallableTypesRuleSpellsResults(t *testing.T) {
	t.Run("a dependency's kind is a Link Type", func(t *testing.T) {
		editor := &roleDependencyEditor{}
		ts := newBDPServer(t, rolesConfig(Config{DependencyEditor: editor}))
		result, _ := callResult(t, ts.callAs(t, "issue/dep/add", "a", `{"source":"bd-1","target":"bd-2","type":"`+ruleRelated+`"}`))
		var got struct {
			Added []map[string]any `json:"added"`
		}
		if err := json.Unmarshal(result, &got); err != nil {
			t.Fatal(err)
		}
		if len(got.Added) != 1 || got.Added[0]["type"] != ruleRelated {
			t.Errorf("added = %v, want the edge's type as %s", got.Added, ruleRelated)
		}
	})

	t.Run("a stored name with no identity is not served", func(t *testing.T) {
		issue := seededIssue("bd-1", "alice", types.StatusOpen)
		issue.IssueType = "zork"
		reader := &roleReader{details: &issueops.IssueDetails{Issue: *issue}}
		ts := newBDPServer(t, rolesConfig(Config{Reader: reader}))
		body := wantProblem(t, ts.call(t, "issue/show", `{"id":"bd-1"}`), http.StatusServiceUnavailable, bdpUnavailable)
		if !strings.Contains(body["detail"].(string), "zork") {
			t.Errorf("detail = %v, want it to name the classification", body["detail"])
		}
		// v0's own route is unchanged.
		if resp := ts.get(t, "/v0/beads/issues/bd-1"); resp.StatusCode != http.StatusOK {
			t.Errorf("v0 status = %d", resp.StatusCode)
		}
	})

	t.Run("v0's route keeps v0's spelling", func(t *testing.T) {
		reader := &roleReader{details: &issueops.IssueDetails{Issue: *seededIssue("bd-1", "alice", types.StatusOpen)}}
		ts := newBDPServer(t, rolesConfig(Config{Reader: reader}))
		direct := readAll(t, ts.get(t, "/v0/beads/issues/bd-1"))
		if !strings.Contains(direct, `"issue_type":"task"`) || strings.Contains(direct, ruleTask) {
			t.Errorf("v0's body changed: %s", direct)
		}
	})
}
