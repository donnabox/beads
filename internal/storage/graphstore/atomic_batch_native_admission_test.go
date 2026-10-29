package graphstore

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/canonicalbatch"
	"github.com/steveyegge/beads/internal/canonicalbatchstate"
	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/types"
	publicops "github.com/steveyegge/beads/issueops"
)

func TestCanonicalNativeAuthoredPolicy(t *testing.T) {
	const id = "https://example.invalid/source"
	before := IssueRecord{ID: id, Type: "https://example.invalid/task", Revision: "original", Properties: &publicops.Issue{ID: "native-source", Title: "Old", IssueType: types.TypeTask, Priority: 2}, Metadata: json.RawMessage(`{}`)}
	state, err := atomicBatchStateRecord(before)
	if err != nil {
		t.Fatal(err)
	}
	run := atomicBatchRun{actor: "test-author", initial: map[string]any{id: before}}
	allowed := canonicalbatch.Member{Operation: "updateBead", PropertiesChange: json.RawMessage(`[{"op":"replace","path":"/title","value":"New"}]`)}
	if err := run.admitNativeMutation(0, allowed, &state); err != nil {
		t.Fatal(err)
	}
	for _, member := range []canonicalbatch.Member{
		{Operation: "deleteBead"},
		{Operation: "updateBead", Properties: json.RawMessage(`{}`)},
		{Operation: "updateBead", PropertiesChange: json.RawMessage(`[{"op":"add","path":"/status","value":"closed"}]`)},
		{Operation: "updateBead", PropertiesChange: json.RawMessage(`[{"op":"add","path":"/title/part","value":"x"}]`)},
		{Operation: "updateBead", Type: canonicalbatch.Reference{URI: "https://example.invalid/bug"}},
	} {
		if err := run.admitNativeMutation(0, member, &state); !errors.Is(err, ErrCapabilityUnavailable) {
			t.Fatalf("unsupported authored native intent admitted: %+v %v", member, err)
		}
	}
	before.Properties.IsTemplate = true
	if err := run.admitNativeMutation(0, allowed, &state); err == nil {
		t.Fatal("template edit was admitted before later cancellation")
	}
}

func TestCanonicalNativeNewStagedRetypeRefused(t *testing.T) {
	const scope = "https://example.invalid/"
	run := atomicBatchRun{store: &Store{options: Options{Binding: Binding{ScopeURL: scope}}}, initial: map[string]any{}}
	memory := canonicalbatch.Reference{URI: scope + "memory"}
	baseline := func(id string) canonicalbatchstate.Resource {
		return canonicalbatchstate.Resource{ID: scope + id, Kind: graph.KindBead, Revision: "old", Type: memory, Properties: json.RawMessage(`{"title":"Old","body":""}`)}
	}
	for _, kind := range []graph.ResourceKind{graph.KindBead, graph.KindLink} {
		t.Run(string(kind), func(t *testing.T) {
			create, change, restore := canonicalbatch.Member{}, canonicalbatch.Member{}, canonicalbatch.Member{}
			if kind == graph.KindBead {
				create = canonicalbatch.Member{Operation: "createBead", ID: "new", Type: memory, Properties: json.RawMessage(`{"title":"New","body":""}`)}
				change = canonicalbatch.Member{Operation: "updateBead", Bead: canonicalbatch.Operand{URI: scope + "new"}, Type: canonicalbatch.Reference{URI: scope + "task"}}
				restore = change
				restore.Type = memory
			} else {
				create = canonicalbatch.Member{Operation: "createLink", ID: "new", Type: canonicalbatch.Reference{URI: scope + "references"}, Source: canonicalbatch.Operand{URI: scope + "a"}, Target: canonicalbatch.Operand{URI: scope + "b"}}
				change = canonicalbatch.Member{Operation: "updateLink", Link: canonicalbatch.Operand{URI: scope + "new"}, Type: canonicalbatch.Reference{URI: scope + "blocks"}}
				restore = change
				restore.Type = create.Type
			}
			plan, err := canonicalbatchstate.Stage(canonicalbatchstate.Input{Scope: scope, Current: map[string]canonicalbatchstate.Resource{scope + "a": baseline("a"), scope + "b": baseline("b")}, Members: []canonicalbatch.Member{create, change, restore}, AdmitMutation: run.admitNativeMutation})
			if !errors.Is(err, ErrCapabilityUnavailable) || len(plan.Final) != 0 {
				t.Fatalf("new identity canceled unsupported native retype: %+v %v", plan, err)
			}
		})
	}
}

func TestCanonicalNativePresenceRefusedBeforeSelection(t *testing.T) {
	const scope = "https://example.invalid/"
	before := IssueRecord{ID: scope + "source", Type: scope + "task", Revision: "original", Properties: &publicops.Issue{ID: "native-source", Title: "Old", IssueType: types.TypeTask, Priority: 2}, Metadata: json.RawMessage(`{}`)}
	state, err := atomicBatchStateRecord(before)
	if err != nil {
		t.Fatal(err)
	}
	run := atomicBatchRun{actor: "test-author", initial: map[string]any{before.ID: before}}
	for _, patch := range []string{
		`[{"op":"add","path":"/estimated_minutes","value":null}]`,
		`[{"op":"add","path":"/external_ref","value":null}]`,
		`[{"op":"add","path":"/due_at","value":null}]`,
		`[{"op":"add","path":"/description","value":""}]`,
		`[{"op":"add","path":"/external_ref","value":""}]`,
		`[{"op":"add","path":"/due_at","value":"2026-10-29T01:00:00+01:00"}]`,
	} {
		members := []canonicalbatch.Member{
			{Operation: "updateBead", Bead: canonicalbatch.Operand{URI: before.ID}, PropertiesChange: json.RawMessage(patch)},
			{Operation: "select", Name: "observed", Set: true, Query: json.RawMessage(`{"ir":"bdp-query-design-1","profile":"list","selection":{"op":"scan","collection":"beads"}}`)},
		}
		if plan, err := canonicalbatchstate.Stage(canonicalbatchstate.Input{Scope: scope, Current: map[string]canonicalbatchstate.Resource{before.ID: state}, Members: members, AdmitMutation: run.admitNativeMutation}); !errors.Is(err, ErrCapabilityUnavailable) || len(plan.Selections) != 0 {
			t.Fatalf("unrepresentable presence reached selection: %s %+v %v", patch, plan, err)
		}
	}
}

func TestCanonicalNativeFinalZeroValuesAndClears(t *testing.T) {
	const scope = "https://example.invalid/"
	before := IssueRecord{ID: scope + "source", Type: scope + "task", Revision: "original", Properties: &publicops.Issue{ID: "native-source", Title: "Old", IssueType: types.TypeTask, Priority: 2}, Metadata: json.RawMessage(`{}`)}
	for _, tc := range []struct {
		key   string
		value string
	}{
		{"priority", `0`}, {"estimated_minutes", `0`}, {"external_ref", `"REF-17"`}, {"due_at", `"2026-10-29T00:00:00Z"`},
	} {
		state, err := atomicBatchStateRecord(before)
		if err != nil {
			t.Fatal(err)
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(state.Properties, &fields); err != nil {
			t.Fatal(err)
		}
		fields[tc.key] = json.RawMessage(tc.value)
		state.Properties, err = canonicalJSON(fields)
		if err != nil {
			t.Fatal(err)
		}
		content, err := canonicalNativeIssueContent(before, state)
		if err != nil {
			t.Fatalf("representable %s refused: %v", tc.key, err)
		}
		if tc.key == "priority" && (!content.Priority.Set || content.Priority.Value != 0) || tc.key == "estimated_minutes" && (!content.EstimatedMinutes.Set || content.EstimatedMinutes.Value == nil || *content.EstimatedMinutes.Value != 0) || tc.key == "external_ref" && (!content.ExternalRef.Set || content.ExternalRef.Value == nil || *content.ExternalRef.Value != "REF-17") || tc.key == "due_at" && (!content.DueAt.Set || content.DueAt.Value == nil) {
			t.Fatalf("zero/nullable presence lost: %s %+v", tc.key, content)
		}
	}
	minutes, external := 10, "REF-17"
	date := time.Date(2026, 10, 29, 0, 0, 0, 0, time.UTC)
	before.Properties.EstimatedMinutes, before.Properties.ExternalRef, before.Properties.DueAt, before.Properties.Description = &minutes, &external, &date, "Clear me"
	state, err := atomicBatchStateRecord(before)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(state.Properties, &fields); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"estimated_minutes", "external_ref", "due_at", "description"} {
		delete(fields, key)
	}
	state.Properties, err = canonicalJSON(fields)
	if err != nil {
		t.Fatal(err)
	}
	content, err := canonicalNativeIssueContent(before, state)
	if err != nil || !content.EstimatedMinutes.Set || content.EstimatedMinutes.Value != nil || !content.ExternalRef.Set || content.ExternalRef.Value != nil || !content.DueAt.Set || content.DueAt.Value != nil || !content.Description.Set || content.Description.Value != "" {
		t.Fatalf("explicit removals did not clear through native setters: %+v %v", content, err)
	}
}

func TestCanonicalNativeFinalContentExactness(t *testing.T) {
	before := IssueRecord{ID: "https://example.invalid/source", Type: "https://example.invalid/task", Revision: "original", Properties: &publicops.Issue{ID: "native-source", Title: "Old", IssueType: types.TypeTask, Priority: 2}, Metadata: json.RawMessage(`{"keep":true}`)}
	state, err := atomicBatchStateRecord(before)
	if err != nil {
		t.Fatal(err)
	}
	with := func(key string, value json.RawMessage) canonicalbatchstate.Resource {
		t.Helper()
		after := state
		var properties map[string]json.RawMessage
		if err := json.Unmarshal(state.Properties, &properties); err != nil {
			t.Fatal(err)
		}
		properties[key] = value
		after.Properties, err = canonicalJSON(properties)
		if err != nil {
			t.Fatal(err)
		}
		return after
	}
	content, err := canonicalNativeIssueContent(before, with("title", json.RawMessage(`"New"`)))
	if err != nil || !content.Title.Set || content.Title.Value != "New" || content.Priority.Set || content.Metadata.Set {
		t.Fatalf("final content widened intent: %+v %v", content, err)
	}
	for key, value := range map[string]json.RawMessage{
		"status": json.RawMessage(`"closed"`), "title": json.RawMessage(`null`),
		"priority": json.RawMessage(`null`), "estimated_minutes": json.RawMessage(`2147483648`),
		"due_at": json.RawMessage(`"2026-10-29T00:00:00.123Z"`),
	} {
		if _, err := canonicalNativeIssueContent(before, with(key, value)); err == nil {
			t.Fatalf("protected or unrepresentable final %s admitted: %s", key, value)
		}
	}
	if _, err := canonicalNativeIssueContent(before, with("title", json.RawMessage(`42`))); !errors.Is(err, graph.ErrValidation) {
		t.Fatalf("native string coercion admitted: %v", err)
	}
}

func TestCanonicalNativeActorAdmissionBeforeNoop(t *testing.T) {
	const scope = "https://example.invalid/"
	issue := IssueRecord{ID: scope + "source", Type: scope + "task", Properties: &publicops.Issue{ID: "source", Title: "Old", IssueType: types.TypeTask}}
	target := IssueRecord{ID: scope + "target", Type: scope + "task", Properties: &publicops.Issue{ID: "target", Title: "Target", IssueType: types.TypeTask}}
	link := LinkRecord{ID: scope + "link", Type: scope + "related", Source: issue.ID, Target: target.ID, Properties: map[string]any{"dep_type": "related"}, Metadata: json.RawMessage(`{}`)}
	run := atomicBatchRun{store: &Store{options: Options{Binding: Binding{ScopeURL: scope}}}, initial: map[string]any{issue.ID: issue, target.ID: target, link.ID: link}, backings: map[string]string{link.ID: "dependency"}}
	for _, tc := range []struct {
		record any
		member canonicalbatch.Member
	}{
		{issue, canonicalbatch.Member{Operation: "updateBead", PropertiesChange: json.RawMessage(`[]`)}},
		{issue, canonicalbatch.Member{Operation: "updateBead", PropertiesChange: json.RawMessage(`[{"op":"replace","path":"/title","value":"Old"}]`)}},
		{link, canonicalbatch.Member{Operation: "updateLink", Metadata: json.RawMessage(`{}`)}},
	} {
		state, err := atomicBatchStateRecord(tc.record)
		if err != nil {
			t.Fatal(err)
		}
		if err := run.admitNativeMutation(0, tc.member, &state); !errors.Is(err, storage.ErrValidation) {
			t.Fatalf("actorless native no-op admitted: %v", err)
		}
		run.actor = "test-author"
		if err := run.admitNativeMutation(0, tc.member, &state); err != nil {
			t.Fatal(err)
		}
		run.actor = ""
	}
}

func TestCanonicalNativeScalarAdmissionBeforeCancellation(t *testing.T) {
	const scope = "https://example.invalid/"
	before := IssueRecord{ID: scope + "source", Type: scope + "task", Revision: "original", Properties: &publicops.Issue{ID: "native-source", Title: "Old", IssueType: types.TypeTask, Priority: 2}, Metadata: json.RawMessage(`{}`)}
	state, err := atomicBatchStateRecord(before)
	if err != nil {
		t.Fatal(err)
	}
	run := atomicBatchRun{actor: "test-author", initial: map[string]any{before.ID: before}}
	for _, patch := range []string{
		`[{"op":"replace","path":"/priority","value":5},{"op":"replace","path":"/priority","value":2}]`,
		`[{"op":"replace","path":"/priority","value":"2"},{"op":"replace","path":"/priority","value":2}]`,
		`[{"op":"replace","path":"/priority","value":2.5},{"op":"replace","path":"/priority","value":2}]`,
		`[{"op":"remove","path":"/priority"},{"op":"add","path":"/priority","value":2}]`,
		`[{"op":"remove","path":"/title"},{"op":"add","path":"/title","value":"Old"}]`,
		`[{"op":"replace","path":"/title","value":null},{"op":"replace","path":"/title","value":"Old"}]`,
		`[{"op":"replace","path":"/title","value":42},{"op":"replace","path":"/title","value":"Old"}]`,
		`[{"op":"replace","path":"/title","value":""},{"op":"replace","path":"/title","value":"Old"}]`,
		`[{"op":"add","path":"/estimated_minutes","value":-1},{"op":"remove","path":"/estimated_minutes"}]`,
		`[{"op":"add","path":"/estimated_minutes","value":2147483648},{"op":"remove","path":"/estimated_minutes"}]`,
		`[{"op":"add","path":"/description","value":12},{"op":"remove","path":"/description"}]`,
	} {
		member := canonicalbatch.Member{Operation: "updateBead", Bead: canonicalbatch.Operand{URI: before.ID}, PropertiesChange: json.RawMessage(patch)}
		selection := canonicalbatch.Member{Operation: "select", Name: "observed", Set: true, Query: json.RawMessage(`{"ir":"bdp-query-design-1","profile":"list","selection":{"op":"scan","collection":"beads"}}`)}
		plan, err := canonicalbatchstate.Stage(canonicalbatchstate.Input{Scope: scope, Current: map[string]canonicalbatchstate.Resource{before.ID: state}, Members: []canonicalbatch.Member{member, selection}, AdmitMutation: run.admitNativeMutation})
		if !errors.Is(err, storage.ErrValidation) || len(plan.Final) != 0 || len(plan.Selections) != 0 {
			t.Fatalf("invalid authored scalar reached canceled state/selection: %s %+v %v", patch, plan, err)
		}
	}
	// The native title validator permits whitespace; do not tighten its policy.
	if err := canonicalNativeAuthoredValue("/title", json.RawMessage(`" "`)); err != nil {
		t.Fatal(err)
	}
	// Nullable removal remains the admitted clear form after shared decoding.
	for _, path := range []string{"/description", "/design", "/acceptance_criteria", "/spec_id", "/estimated_minutes", "/external_ref", "/due_at"} {
		if err := canonicalNativeAuthoredValue(path, nil); err != nil {
			t.Fatalf("optional clear %s refused: %v", path, err)
		}
	}
}
