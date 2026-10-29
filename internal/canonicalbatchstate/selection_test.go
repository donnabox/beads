package canonicalbatchstate

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/canonicalbatch"
	"github.com/steveyegge/beads/internal/resourcequery"
)

func selectMember(name, source, predicate string) canonicalbatch.Member {
	if predicate != "" {
		source = `{"op":"filter","input":` + source + `,"bind":"b","predicate":` + predicate + `}`
	}
	return canonicalbatch.Member{Operation: "select", Name: name, Set: true, Query: raw(`{"ir":"bdp-query-design-1","profile":"list","selection":` + source + `}`)}
}

const beadScan = `{"op":"scan","collection":"beads"}`
const linkScan = `{"op":"scan","collection":"links"}`
const titleIsNew = `{"op":"compare","operator":"eq","left":{"op":"field","of":{"op":"binding","name":"b"},"path":["properties","title"]},"right":{"op":"literal","value":"new"}}`

func TestSelectionSeesStagedValuesAndFreezesMembership(t *testing.T) {
	base := existing("one")
	change := canonicalbatch.Member{Operation: "updateBead", Bead: operand("one"), Properties: raw(`{"title":"new"}`)}
	selectIDs := selectMember("picked", beadScan, titleIsNew)
	updateSet := canonicalbatch.Member{Operation: "updateBead", Bead: canonicalbatch.Operand{Binding: "picked"}, Properties: raw(`{"title":"changed after capture"}`)}
	deleteSet := canonicalbatch.Member{Operation: "deleteBead", Bead: canonicalbatch.Operand{Binding: "picked"}}
	plan, err := Stage(Input{Scope: testScope, Current: map[string]Resource{base.ID: base}, Members: []canonicalbatch.Member{change, createBead("new", ""), selectIDs, updateSet, createBead("later", ""), deleteSet}})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{testScope + "new", testScope + "one"}
	if len(plan.Selections) != 1 || !reflect.DeepEqual(plan.Selections[0].IDs, want) {
		t.Fatalf("selection did not see staged creates/edits: %+v", plan.Selections)
	}
	if len(plan.Final) != 1 || plan.Final[testScope+"later"].ID == "" {
		t.Fatalf("predicate was rerun or late creation joined fixed set: %+v", plan.Final)
	}
	if len(plan.Outcomes) != 3 || plan.Outcomes[1].Outcome != "allocated" || plan.Outcomes[2].Outcome != "deleted" {
		t.Fatalf("incorrect final publication: %+v", plan.Outcomes)
	}
	for _, mapping := range plan.Mappings {
		if mapping.OperationIndex == 3 || mapping.OperationIndex == 5 {
			if mapping.ID == testScope+"later" {
				t.Fatal("late identity entered fixed set")
			}
		}
	}
}

func TestLaterSelectionReadsUpdatedBodiesOfPriorSet(t *testing.T) {
	one, two := existing("one"), existing("two")
	first := selectMember("all", beadScan, "")
	change := canonicalbatch.Member{Operation: "updateBead", Bead: operand("one"), Properties: raw(`{"title":"ready","ready":true}`)}
	second := selectMember("ready", `{"op":"refs","collection":"beads","references":{"binding":"all"}}`, `{"op":"present","value":{"op":"field","of":{"op":"binding","name":"b"},"path":["properties","ready"]}}`)
	plan, err := Stage(Input{Scope: testScope, Current: map[string]Resource{one.ID: one, two.ID: two}, Members: []canonicalbatch.Member{first, change, second}})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(plan.Selections[0].IDs, []string{one.ID, two.ID}) || !reflect.DeepEqual(plan.Selections[1].IDs, []string{one.ID}) {
		t.Fatalf("capture froze bodies instead of identity membership: %+v", plan.Selections)
	}
}

func TestSelectionCanReadPriorScalarCreation(t *testing.T) {
	query := selectMember("fromScalar", `{"op":"refs","collection":"beads","references":{"binding":"created"}}`, titleIsNew)
	plan, err := Stage(Input{Scope: testScope, Members: []canonicalbatch.Member{createBead("one", "created"), query}})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(plan.Selections[0].IDs, []string{testScope + "one"}) {
		t.Fatalf("scalar source not captured: %+v", plan.Selections)
	}
}

func TestEmptySetWritesAreNoopsButCannotSatisfyGuard(t *testing.T) {
	query := selectMember("empty", beadScan, "false")
	update := canonicalbatch.Member{Operation: "updateBead", Bead: canonicalbatch.Operand{Binding: "empty"}, Properties: raw(`{}`)}
	remove := canonicalbatch.Member{Operation: "deleteBead", Bead: canonicalbatch.Operand{Binding: "empty"}}
	plan, err := Stage(Input{Scope: testScope, Members: []canonicalbatch.Member{query, update, remove}})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Outcomes) != 0 || len(plan.Mappings) != 0 || len(plan.Selections) != 1 || plan.Selections[0].IDs == nil {
		t.Fatalf("empty set created writes or missing capture: %+v", plan)
	}
	update.ExpectedRevision = "token"
	if _, err := Stage(Input{Scope: testScope, Members: []canonicalbatch.Member{query, update}}); !errors.Is(err, ErrRevision) {
		t.Fatalf("empty set accepted guard: %v", err)
	}
}

func TestSetGuardsRequireOneCommittedBaseline(t *testing.T) {
	one, two := existing("one"), existing("two")
	query := selectMember("picked", beadScan, "")
	update := canonicalbatch.Member{Operation: "updateBead", Bead: canonicalbatch.Operand{Binding: "picked"}, ExpectedRevision: one.Revision, Properties: raw(`{"title":"guarded"}`)}
	if _, err := Stage(Input{Scope: testScope, Current: map[string]Resource{one.ID: one}, Members: []canonicalbatch.Member{query, update}}); err != nil {
		t.Fatalf("singleton baseline guard refused: %v", err)
	}
	if _, err := Stage(Input{Scope: testScope, Current: map[string]Resource{one.ID: one, two.ID: two}, Members: []canonicalbatch.Member{query, update}}); !errors.Is(err, ErrRevision) {
		t.Fatalf("multi-target guard accepted: %v", err)
	}
	if _, err := Stage(Input{Scope: testScope, Members: []canonicalbatch.Member{createBead("one", ""), query, update}}); !errors.Is(err, ErrRevision) {
		t.Fatalf("staged creation satisfied baseline guard: %v", err)
	}
}

func TestCapturedDeletedIdentityIsNotSilentlyDropped(t *testing.T) {
	one := existing("one")
	query := selectMember("picked", beadScan, "")
	remove := canonicalbatch.Member{Operation: "deleteBead", Bead: operand("one")}
	update := canonicalbatch.Member{Operation: "updateBead", Bead: canonicalbatch.Operand{Binding: "picked"}, Properties: raw(`{}`)}
	plan, err := Stage(Input{Scope: testScope, Current: map[string]Resource{one.ID: one}, Members: []canonicalbatch.Member{query, remove, update}})
	if !errors.Is(err, ErrNotFound) || !reflect.DeepEqual(plan, Plan{}) {
		t.Fatalf("captured deletion became successful empty selection: %+v %v", plan, err)
	}
	refs := selectMember("again", `{"op":"refs","collection":"beads","references":{"binding":"picked"}}`, "")
	if _, err := Stage(Input{Scope: testScope, Current: map[string]Resource{one.ID: one}, Members: []canonicalbatch.Member{query, remove, refs}}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("later selection omitted unavailable captured identity: %v", err)
	}
}

func TestSetKindsAndScalarEndpointSlotsRemainDistinct(t *testing.T) {
	one := existing("one")
	query := selectMember("picked", beadScan, "")
	link := createLink("link", "one", "one")
	link.Source = canonicalbatch.Operand{Binding: "picked"}
	if _, err := Stage(Input{Scope: testScope, Current: map[string]Resource{one.ID: one}, Members: []canonicalbatch.Member{query, link}}); !errors.Is(err, ErrKind) {
		t.Fatalf("a set filled a scalar endpoint: %v", err)
	}
	wrong := canonicalbatch.Member{Operation: "deleteLink", Link: canonicalbatch.Operand{Binding: "picked"}}
	if _, err := Stage(Input{Scope: testScope, Current: map[string]Resource{one.ID: one}, Members: []canonicalbatch.Member{query, wrong}}); !errors.Is(err, ErrKind) {
		t.Fatalf("Bead set filled a Link target: %v", err)
	}
}

func TestLinkSelectionAndSetDelete(t *testing.T) {
	one, two := existing("one"), existing("two")
	link := Resource{ID: testScope + "citation", Kind: graphops.KindLink, Revision: "link-revision", Type: reference("references"), Properties: raw(`{}`), Metadata: raw(`{}`), Source: reference("one"), Target: reference("two")}
	query := selectMember("links", linkScan, "")
	remove := canonicalbatch.Member{Operation: "deleteLink", Link: canonicalbatch.Operand{Binding: "links"}}
	plan, err := Stage(Input{Scope: testScope, Current: map[string]Resource{one.ID: one, two.ID: two, link.ID: link}, Members: []canonicalbatch.Member{query, remove}})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Final) != 2 || plan.Selections[0].Kind != graphops.KindLink || plan.Outcomes[0].Outcome != "deleted" {
		t.Fatalf("Link set mutation broken: %+v", plan)
	}
}

func TestLocalListValueSemantics(t *testing.T) {
	one := existing("one")
	one.Properties = raw(`{"title":"alpha%beta","null":null,"flag":false,"count":2,"array":[null,false,2],"object":{"x":1}}`)
	field := func(path string) string {
		return `{"op":"field","of":{"op":"binding","name":"b"},"path":["properties","` + path + `"]}`
	}
	compare := func(path, op, value string) string {
		return `{"op":"compare","operator":"` + op + `","left":` + field(path) + `,"right":{"op":"literal","value":` + value + `}}`
	}
	for _, tc := range []struct {
		name, predicate string
		want            bool
	}{
		{"absent-null equality", compare("absent", "eq", "null"), false},
		{"absent-null inequality", compare("absent", "ne", "null"), true},
		{"null physically present", `{"op":"present","value":` + field("null") + `}`, true},
		{"false physically present", `{"op":"present","value":` + field("flag") + `}`, true},
		{"no numeric string coercion", compare("count", "eq", `"2"`), false},
		{"numeric ordering", compare("count", "gt", "1.0"), true},
		{"object structural equality", compare("object", "eq", `{"x":1.0}`), true},
		{"literal contains", `{"op":"contains","text":` + field("title") + `,"substring":{"op":"literal","value":"%"}}`, true},
		{"case sensitive", `{"op":"contains","text":` + field("title") + `,"substring":{"op":"literal","value":"ALPHA"}}`, false},
		{"literal in local array", `{"op":"member","value":{"op":"literal","value":false},"array":` + field("array") + `}`, true},
		{"local scalar in literal array", `{"op":"member","value":` + field("count") + `,"array":{"op":"literal","value":[1,2,3]}}`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			plan, err := Stage(Input{Scope: testScope, Current: map[string]Resource{one.ID: one}, Members: []canonicalbatch.Member{selectMember("picked", beadScan, tc.predicate)}})
			if err != nil {
				t.Fatal(err)
			}
			if (len(plan.Selections[0].IDs) == 1) != tc.want {
				t.Fatalf("incorrect answer: %+v", plan.Selections)
			}
		})
	}
}

func TestSelectionTypeAuthorityAndPinnedAffiliation(t *testing.T) {
	one := existing("one")
	one.Type = canonicalbatch.Reference{URI: testScope + "task", Revision: "exact-definition"}
	predicate := `{"op":"typeIs","of":{"op":"binding","name":"b"},"type":"` + testScope + `memory","match":"family"}`
	query := selectMember("picked", beadScan, predicate)
	input := Input{Scope: testScope, Current: map[string]Resource{one.ID: one}, Members: []canonicalbatch.Member{query}, TypeKinds: map[string]graphops.ResourceKind{testScope + "memory": graphops.KindBead, testScope + "task": graphops.KindBead}, TypeFamilies: map[string][]string{testScope + "memory": {testScope + "memory"}, testScope + "task": {testScope + "task", testScope + "memory"}}}
	plan, err := Stage(input)
	if err != nil || len(plan.Selections[0].IDs) != 1 {
		t.Fatalf("family did not use checked ancestry: %+v %v", plan, err)
	}
	input.TypeFamilies = nil
	input.Current = nil
	if _, err := Stage(input); !errors.Is(err, ErrUnsupported) {
		t.Fatalf("empty scan silently substituted exact matching: %v", err)
	}
	input.TypeFamilies = map[string][]string{testScope + "memory": {testScope + "memory"}}
	if _, err := Stage(input); !errors.Is(err, ErrInvalid) {
		t.Fatalf("incomplete ancestry accepted: %v", err)
	}
	exact := strings.ReplaceAll(strings.ReplaceAll(predicate, `"family"`, `"exact"`), testScope+"memory", testScope+"task")
	input.Members = []canonicalbatch.Member{selectMember("exact", beadScan, exact)}
	input.Current = map[string]Resource{one.ID: one}
	input.TypeKinds = nil
	input.TypeFamilies = nil
	plan, err = Stage(input)
	if err != nil || len(plan.Selections[0].IDs) != 1 {
		t.Fatalf("exact Type compared definition revision: %+v %v", plan, err)
	}
}

func TestUnsupportedPredicateBranchesRefuseBeforeAllocator(t *testing.T) {
	for _, predicate := range []string{
		`{"op":"or","operands":[true,{"op":"any","input":{"op":"scan","collection":"beads"},"bind":"other","predicate":true}]}`,
		`{"op":"compare","operator":"eq","left":{"op":"field","of":{"op":"binding","name":"b"},"path":["properties","a"]},"right":{"op":"field","of":{"op":"binding","name":"b"},"path":["properties","b"]}}`,
		`{"op":"present","value":{"op":"field","of":{"op":"binding","name":"b"},"path":["derived","ownedLinks"]}}`,
	} {
		create := createBead("unused", "")
		create.ID = ""
		called := false
		_, err := Stage(Input{Scope: testScope, Members: []canonicalbatch.Member{create, selectMember("unsupported", beadScan, predicate)}, Allocate: func(int, graphops.ResourceKind) (string, error) { called = true; return testScope + "allocated", nil }})
		if !errors.Is(err, ErrUnsupported) || called {
			t.Fatalf("unsupported branch was hidden by empty input/short circuit: called=%v err=%v", called, err)
		}
	}
}

func TestSelectionLiteralObjectsAreNeverBindingOrCode(t *testing.T) {
	one := existing("one")
	one.Properties = raw(`{"title":"initial","object":{"binding":"not-a-batch-name","op":"walk"}}`)
	predicate := `{"op":"compare","operator":"eq","left":{"op":"field","of":{"op":"binding","name":"b"},"path":["properties","object"]},"right":{"op":"literal","value":{"binding":"not-a-batch-name","op":"walk"}}}`
	plan, err := Stage(Input{Scope: testScope, Current: map[string]Resource{one.ID: one}, Members: []canonicalbatch.Member{selectMember("picked", beadScan, predicate)}})
	if err != nil || len(plan.Selections[0].IDs) != 1 {
		t.Fatalf("literal was interpreted as structure: %+v %v", plan, err)
	}
}

func TestSelectionResultAndPriorCapturesAreDetached(t *testing.T) {
	one := existing("one")
	first := selectMember("first", beadScan, "")
	second := selectMember("second", `{"op":"refs","collection":"beads","references":{"binding":"first"}}`, "")
	plan, err := Stage(Input{Scope: testScope, Current: map[string]Resource{one.ID: one}, Members: []canonicalbatch.Member{first, second}})
	if err != nil {
		t.Fatal(err)
	}
	plan.Selections[0].IDs[0] = "changed"
	if plan.Selections[1].IDs[0] != one.ID || one.ID != testScope+"one" {
		t.Fatal("captures share mutable slices")
	}
}

func TestStagedRefAdapterParityWithSharedListKernel(t *testing.T) {
	one, two, excluded := existing("one"), existing("two"), existing("excluded")
	one.Type.Revision = "definition-pin"
	one.Metadata = raw(`{"visible":null}`)
	two.Metadata = nil
	excluded.Metadata = raw(`{"visible":null}`)
	one.Properties = raw(`{"title":"one","literal":{"op":"refs","binding":"not-code","input":{"binding":"not-input"}}}`)
	prior := selectMember("prior", `{"op":"refs","collection":"beads","references":["`+one.ID+`","`+two.ID+`"]}`, "")
	for _, tc := range []struct{ name, predicate string }{
		{"metadata omission", `{"op":"present","value":{"op":"field","of":{"op":"binding","name":"b"},"path":["metadata","visible"]}}`},
		{"pinned Type representation", `{"op":"compare","operator":"eq","left":{"op":"field","of":{"op":"binding","name":"b"},"path":["type","revision"]},"right":{"op":"literal","value":"definition-pin"}}`},
		{"opaque literal object", `{"op":"compare","operator":"eq","left":{"op":"field","of":{"op":"binding","name":"b"},"path":["properties","literal"]},"right":{"op":"literal","value":{"op":"refs","binding":"not-code","input":{"binding":"not-input"}}}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			staged := selectMember("selected", `{"op":"refs","collection":"beads","references":{"binding":"prior"}}`, tc.predicate)
			plan, err := Stage(Input{Scope: testScope, Current: map[string]Resource{one.ID: one, two.ID: two, excluded.ID: excluded}, Members: []canonicalbatch.Member{prior, staged}})
			if err != nil {
				t.Fatal(err)
			}
			fragment := raw(`{"op":"filter","input":` + beadScan + `,"bind":"b","predicate":` + tc.predicate + `}`)
			selector, err := resourcequery.CompileListIDs(fragment, resourcequery.Limits{})
			if err != nil {
				t.Fatal(err)
			}
			inventory := resourcequery.ListInventory{Kind: graphops.KindBead, Candidates: []resourcequery.ListCandidate{}}
			for _, record := range []Resource{one, two} {
				// Independent closed input construction catches adapter omissions
				// and broadening of a refs input into the entire Scope inventory.
				body := map[string]any{"id": record.ID, "type": record.Type, "properties": record.Properties}
				if record.Metadata != nil {
					body["metadata"] = record.Metadata
				}
				data, err := json.Marshal(body)
				if err != nil {
					t.Fatal(err)
				}
				inventory.Candidates = append(inventory.Candidates, resourcequery.ListCandidate{Data: data})
			}
			want, err := selector.Select(context.Background(), inventory)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(plan.Selections[1].IDs, want) || !reflect.DeepEqual(want, []string{one.ID}) {
				t.Fatalf("staged/source kernel disagreement: staged=%v direct=%v", plan.Selections[1].IDs, want)
			}
		})
	}
}

func TestSharedSelectorCoreOnlyExactAndOptionalMetadata(t *testing.T) {
	one, two := existing("one"), existing("two")
	one.Metadata = nil
	two.Metadata = raw(`{}`)
	exact := `{"op":"typeIs","of":{"op":"binding","name":"b"},"type":"` + one.Type.URI + `","match":"exact"}`
	metadata := `{"op":"present","value":{"op":"field","of":{"op":"binding","name":"b"},"path":["metadata"]}}`
	plan, err := Stage(Input{Scope: testScope, Current: map[string]Resource{one.ID: one, two.ID: two}, Members: []canonicalbatch.Member{selectMember("exact", beadScan, exact), selectMember("present", beadScan, metadata)}})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(plan.Selections[0].IDs, []string{one.ID, two.ID}) || !reflect.DeepEqual(plan.Selections[1].IDs, []string{two.ID}) {
		t.Fatalf("Core nominal or metadata presence changed: %+v", plan.Selections)
	}
}

func TestExplicitMissingRefRefusesEvenWithFalsePredicate(t *testing.T) {
	one := existing("one")
	query := selectMember("missing", `{"op":"refs","collection":"beads","references":["`+one.ID+`","`+testScope+`missing"]}`, "false")
	plan, err := Stage(Input{Scope: testScope, Current: map[string]Resource{one.ID: one}, Members: []canonicalbatch.Member{query}})
	if !errors.Is(err, ErrNotFound) || !reflect.DeepEqual(plan, Plan{}) {
		t.Fatalf("missing refs became empty success: %+v %v", plan, err)
	}
}

func TestSharedSelectionLimitsReturnNoPartialPlan(t *testing.T) {
	one, two := existing("one"), existing("two")
	for _, tc := range []struct {
		name   string
		change func(*resourcequery.Limits)
		filter bool
	}{
		{"result bound", func(l *resourcequery.Limits) { l.MaxResults = 1 }, false},
		{"work bound", func(l *resourcequery.Limits) { l.MaxWork = 1 }, false},
		{"request bound", func(l *resourcequery.Limits) { l.RequestBytes = 64 }, false},
		{"node bound", func(l *resourcequery.Limits) { l.Nodes = 1 }, true},
		{"nesting bound", func(l *resourcequery.Limits) { l.Nesting = 1 }, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			limits := resourcequery.DefaultLimits()
			tc.change(&limits)
			predicate := ""
			if tc.filter {
				predicate = titleIsNew
			}
			plan, err := Stage(Input{Scope: testScope, SelectionLimits: limits, Current: map[string]Resource{one.ID: one, two.ID: two}, Members: []canonicalbatch.Member{selectMember("limited", beadScan, predicate)}})
			if !errors.Is(err, ErrLimit) || !reflect.DeepEqual(plan, Plan{}) {
				t.Fatalf("selector limit produced partial success: %+v %v", plan, err)
			}
		})
	}
	limits := resourcequery.DefaultLimits()
	limits.MaxResults = 1
	plan, err := Stage(Input{Scope: testScope, SelectionLimits: limits, Current: map[string]Resource{one.ID: one, two.ID: two}, Members: []canonicalbatch.Member{selectMember("first", beadScan, "false"), selectMember("later", beadScan, "")}})
	if !errors.Is(err, ErrLimit) || !reflect.DeepEqual(plan, Plan{}) {
		t.Fatalf("earlier capture escaped later failure: %+v %v", plan, err)
	}
}

func TestStagedContextCancellationBeforeAndDuringWork(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	called := false
	create := createBead("unused", "")
	create.ID = ""
	plan, err := Stage(Input{Context: ctx, Scope: testScope, Members: []canonicalbatch.Member{create}, Allocate: func(int, graphops.ResourceKind) (string, error) { called = true; return testScope + "allocated", nil }})
	if !errors.Is(err, context.Canceled) || called || !reflect.DeepEqual(plan, Plan{}) {
		t.Fatalf("cancelled input executed: %+v %v", plan, err)
	}
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	plan, err = Stage(Input{Context: ctx, Scope: testScope, Members: []canonicalbatch.Member{create}, Allocate: func(int, graphops.ResourceKind) (string, error) { cancel(); return testScope + "allocated", nil }})
	if !errors.Is(err, context.Canceled) || !reflect.DeepEqual(plan, Plan{}) {
		t.Fatalf("cancellation in last member returned a plan: %+v %v", plan, err)
	}
}

type selectionCancellingContext struct {
	context.Context
	cancel context.CancelFunc
	checks atomic.Int64
}

func (c *selectionCancellingContext) Err() error {
	if c.checks.Add(1) == 100 {
		c.cancel()
	}
	return c.Context.Err()
}

func TestContextReachesSharedPredicateKernel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	checked := &selectionCancellingContext{Context: ctx, cancel: cancel}
	one := existing("one")
	one.Properties = raw(`{"title":"initial","values":[` + strings.Repeat(`"item",`, 199) + `"item"]}`)
	predicate := `{"op":"member","value":{"op":"literal","value":"absent"},"array":{"op":"field","of":{"op":"binding","name":"b"},"path":["properties","values"]}}`
	plan, err := Stage(Input{Context: checked, Scope: testScope, Current: map[string]Resource{one.ID: one}, Members: []canonicalbatch.Member{selectMember("cancelled", beadScan, predicate)}})
	if !errors.Is(err, context.Canceled) || !reflect.DeepEqual(plan, Plan{}) {
		t.Fatalf("shared predicate ignored transaction cancellation: checks=%d plan=%+v err=%v", checked.checks.Load(), plan, err)
	}
}

func TestSharedRevisionlessListRejectsUnsupportedFieldsInAllBranches(t *testing.T) {
	for _, field := range []string{"source", "target", "revision", "revisionProvenance", "derived"} {
		t.Run(field, func(t *testing.T) {
			query := selectMember("unsupported", beadScan, `{"op":"or","operands":[true,{"op":"present","value":{"op":"field","of":{"op":"binding","name":"b"},"path":["`+field+`"]}}]}`)
			called := false
			create := createBead("unused", "")
			create.ID = ""
			plan, err := Stage(Input{Scope: testScope, Members: []canonicalbatch.Member{create, query}, Allocate: func(int, graphops.ResourceKind) (string, error) { called = true; return testScope + "allocated", nil }})
			if !errors.Is(err, ErrUnsupported) || called || !reflect.DeepEqual(plan, Plan{}) {
				t.Fatalf("unsupported field hidden by true branch/empty inventory: %+v %v", plan, err)
			}
		})
	}
}

func TestCandidateByteBudgetIncludesIdentityAndType(t *testing.T) {
	one := existing(strings.Repeat("a", 512))
	one.Type.URI = testScope + strings.Repeat("t", 512)
	if len(one.Properties)+len(one.Metadata) >= 100 {
		t.Fatal("fixture must fit obsolete payload-only accounting")
	}
	member, err := admit(selectMember("later", beadScan, ""), resourcequery.Limits{})
	if err != nil {
		t.Fatal(err)
	}
	state := staging{input: Input{Context: context.Background()}, live: map[string]Resource{one.ID: one}, bindings: map[string]capturedBinding{}, selectionBytes: MaxSelectionBytes - 100}
	if err := state.selectIDs(0, member); !errors.Is(err, ErrLimit) {
		t.Fatalf("large identity/Type view bypassed cumulative byte bound: %v", err)
	}
	if len(state.selections) != 0 || len(state.bindings) != 0 {
		t.Fatal("over-budget candidate returned a partial capture")
	}
}
