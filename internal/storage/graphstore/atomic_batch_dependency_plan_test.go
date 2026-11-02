package graphstore

import (
	"encoding/json"
	"errors"
	"testing"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/canonicalbatch"
	"github.com/steveyegge/beads/internal/canonicalbatchstate"
	"github.com/steveyegge/beads/internal/storage/issueops"
	"github.com/steveyegge/beads/internal/types"
	publicops "github.com/steveyegge/beads/issueops"
)

func nativeDependencyPlanningRun(t *testing.T, existing bool) (atomicBatchRun, map[string]canonicalbatchstate.Resource) {
	t.Helper()
	const scope = "https://example.invalid/"
	run := atomicBatchRun{store: &Store{options: Options{Binding: Binding{ScopeURL: scope}}}, actor: "actor", initial: map[string]any{}, backings: map[string]string{}, ownerIDs: map[string][]string{}, ownerCandidates: map[string]bool{}}
	current := map[string]canonicalbatchstate.Resource{}
	for _, id := range []string{"a", "b", "c"} {
		record := IssueRecord{ID: scope + id, Type: scope + "task", Revision: "original", Properties: &publicops.Issue{ID: "native-" + id, Title: id, IssueType: types.TypeTask, Priority: 2}, Metadata: json.RawMessage(`{}`)}
		run.initial[record.ID] = record
		run.backings[record.ID] = "issue"
		state, err := atomicBatchStateRecord(record)
		if err != nil {
			t.Fatal(err)
		}
		current[record.ID] = state
	}
	if existing {
		link := LinkRecord{ID: scope + "old", Type: scope + "blocks", Revision: "old-link", Source: scope + "a", Target: scope + "b", Properties: map[string]any{"dep_type": "blocks"}, Metadata: json.RawMessage(`{"value":1}`)}
		run.initial[link.ID] = link
		run.backings[link.ID] = "dependency"
		state, err := atomicBatchStateRecord(link)
		if err != nil {
			t.Fatal(err)
		}
		current[link.ID] = state
	}
	return run, current
}

func stageNativeDependencyMembers(t *testing.T, run *atomicBatchRun, current map[string]canonicalbatchstate.Resource, members ...canonicalbatch.Member) error {
	t.Helper()
	plan, err := canonicalbatchstate.Stage(canonicalbatchstate.Input{Scope: run.store.ScopeURL(), Current: current, Members: members, AdmitMutation: run.admitNativeMutation, AdmitResolvedMutation: run.admitResolvedNativeMutation})
	if err == nil {
		run.plan = plan
	}
	return err
}

func nativeDependencyCreate(id, source, target string) canonicalbatch.Member {
	const scope = "https://example.invalid/"
	return canonicalbatch.Member{Operation: "createLink", ID: id, Type: canonicalbatch.Reference{URI: scope + "blocks"}, Source: canonicalbatch.Operand{URI: scope + source}, Target: canonicalbatch.Operand{URI: scope + target}, Properties: json.RawMessage(`{"dep_type":"blocks"}`), Metadata: json.RawMessage(`{}`)}
}

func TestCanonicalDependencyPairLineageTransitions(t *testing.T) {
	const scope = "https://example.invalid/"
	for _, scenario := range []string{"insert", "metadata", "delete", "reassert", "metadata-noop", "allocation-only"} {
		t.Run(scenario, func(t *testing.T) {
			run, current := nativeDependencyPlanningRun(t, scenario != "insert" && scenario != "allocation-only")
			var members []canonicalbatch.Member
			switch scenario {
			case "insert":
				members = []canonicalbatch.Member{nativeDependencyCreate("new", "a", "b")}
			case "metadata":
				members = []canonicalbatch.Member{{Operation: "updateLink", Link: canonicalbatch.Operand{URI: scope + "old"}, Metadata: json.RawMessage(`{"value":2}`)}}
			case "delete":
				members = []canonicalbatch.Member{{Operation: "deleteLink", Link: canonicalbatch.Operand{URI: scope + "old"}}}
			case "reassert":
				members = []canonicalbatch.Member{nativeDependencyCreate("new", "a", "b"), {Operation: "deleteLink", Link: canonicalbatch.Operand{URI: scope + "old"}}}
			case "metadata-noop":
				members = []canonicalbatch.Member{{Operation: "updateLink", Link: canonicalbatch.Operand{URI: scope + "old"}, Metadata: json.RawMessage(`{"value":2}`)}, {Operation: "updateLink", Link: canonicalbatch.Operand{URI: scope + "old"}, Metadata: json.RawMessage(`{"value":1}`)}}
			case "allocation-only":
				members = []canonicalbatch.Member{nativeDependencyCreate("new", "a", "b"), {Operation: "deleteLink", Link: canonicalbatch.Operand{URI: scope + "new"}}}
			}
			if err := stageNativeDependencyMembers(t, &run, current, members...); err != nil {
				t.Fatal(err)
			}
			if !run.nativeDependencyAuthored {
				t.Fatal("authored native state was not observed")
			}
			if err := run.admit(); err != nil {
				t.Fatal(err)
			}
			plan, err := run.planNativeDependencies()
			if err != nil {
				t.Fatal(err)
			}
			normalized, err := issueops.NormalizeCanonicalDependencyFinalRequest(plan.request)
			if err != nil {
				t.Fatal(err)
			}
			if len(normalized.Population) != 3 {
				t.Fatalf("native population omitted untouched Issue: %v", normalized.Population)
			}
			if scenario == "metadata-noop" || scenario == "allocation-only" {
				if len(normalized.Changes) != 0 {
					t.Fatalf("final no-row decision has native effects: %+v", normalized.Changes)
				}
				if scenario == "allocation-only" && len(run.plan.ConsumedIDs) != 1 {
					t.Fatal("transient identity was not consumed")
				}
				return
			}
			if len(normalized.Changes) != 1 || string(normalized.Changes[0].Transition) != scenario {
				t.Fatalf("wrong final transition: %+v", normalized.Changes)
			}
			lineage := plan.lineage[issueops.CanonicalDependencyPair{SourceID: "native-a", TargetID: "native-b"}]
			if scenario == "reassert" && (lineage.BeforeID != scope+"old" || lineage.AfterID != scope+"new" || lineage.BeforeID == lineage.AfterID) {
				t.Fatalf("reassert identity collapsed: %+v", lineage)
			}
		})
	}
}

func TestCanonicalDependencyFinalPairConflictAndRetypeRefused(t *testing.T) {
	run, current := nativeDependencyPlanningRun(t, true)
	if err := stageNativeDependencyMembers(t, &run, current, nativeDependencyCreate("new", "a", "b")); err != nil {
		t.Fatal(err)
	}
	if _, err := run.planNativeDependencies(); !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("new public identity adopted occupied native pair: %v", err)
	}
	run, current = nativeDependencyPlanningRun(t, true)
	create := nativeDependencyCreate("new", "a", "b")
	create.Type.URI = run.store.ScopeURL() + "related"
	create.Properties = json.RawMessage(`{"dep_type":"related"}`)
	if err := stageNativeDependencyMembers(t, &run, current, create, canonicalbatch.Member{Operation: "deleteLink", Link: canonicalbatch.Operand{URI: run.store.ScopeURL() + "old"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := run.planNativeDependencies(); !errors.Is(err, ErrCapabilityUnavailable) {
		t.Fatalf("reassert reclassified native row: %v", err)
	}
}

func TestCanonicalDependencyCanceledCreateAdmission(t *testing.T) {
	const scope = "https://example.invalid/"
	for _, scenario := range []string{"actor", "self", "classification", "non-native-endpoint", "template", "pin", "later-retype", "later-property-edit"} {
		t.Run(scenario, func(t *testing.T) {
			run, current := nativeDependencyPlanningRun(t, false)
			create := nativeDependencyCreate("new", "a", "b")
			members := []canonicalbatch.Member{create}
			switch scenario {
			case "actor":
				run.actor = ""
			case "self":
				members[0].Target = members[0].Source
			case "classification":
				members[0].Properties = json.RawMessage(`{"dep_type":"related"}`)
			case "non-native-endpoint":
				delete(run.initial, scope+"b")
			case "template":
				record := run.initial[scope+"a"].(IssueRecord)
				record.Properties.IsTemplate = true
			case "pin":
				members[0].Source.Revision = "TOKEN"
			case "later-retype":
				members = append(members, canonicalbatch.Member{Operation: "updateLink", Link: canonicalbatch.Operand{URI: scope + "new"}, Type: canonicalbatch.Reference{URI: scope + "related"}})
			case "later-property-edit":
				members = append(members, canonicalbatch.Member{Operation: "updateLink", Link: canonicalbatch.Operand{URI: scope + "new"}, PropertiesChange: json.RawMessage(`[{"op":"replace","path":"/dep_type","value":"related"}]`)})
			}
			members = append(members, canonicalbatch.Member{Operation: "deleteLink", Link: canonicalbatch.Operand{URI: scope + "new"}})
			if err := stageNativeDependencyMembers(t, &run, current, members...); err == nil {
				t.Fatal("canceled invalid native intent was admitted")
			}
		})
	}
}

func TestCanonicalDependencyInformationalIssueLinkStaysInformational(t *testing.T) {
	run, current := nativeDependencyPlanningRun(t, false)
	member := nativeDependencyCreate("new", "a", "b")
	member.Type.URI = run.store.ScopeURL() + "references"
	member.Properties = json.RawMessage(`{}`)
	if err := stageNativeDependencyMembers(t, &run, current, member); err != nil {
		t.Fatal(err)
	}
	if run.nativeDependencyAuthored || run.backings[run.store.ScopeURL()+"new"] == "dependency" {
		t.Fatal("endpoint kinds were used to infer native Link authority")
	}
	if run.plan.Final[run.store.ScopeURL()+"new"].Kind != graph.KindLink {
		t.Fatal("informational Link disappeared")
	}
}
