//go:build cgo

package graphstore

import (
	"bytes"
	"errors"
	"reflect"
	"testing"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/resourceget"
)

func TestInstalledTypesAreOrdinaryCompleteBeads(t *testing.T) {
	ctx, _, s := intrinsicHardeningStore(t)
	snapshot, err := s.CurrentResourceDocuments(ctx)
	if err != nil {
		t.Fatal(err)
	}
	byID := map[string]resourceget.ResourceDocument{}
	beads := 0
	for _, document := range snapshot.Documents {
		if _, duplicate := byID[document.ID()]; duplicate {
			t.Fatal("duplicate ordinary identity")
		}
		byID[document.ID()] = document
		if document.Kind() == graph.KindBead {
			beads++
			if _, definition := document.Definition(); !definition {
				t.Fatalf("installed definition %s is not an actual checked ordinary Bead", document.ID())
			}
		}
	}
	if beads != len(previewTypeDefinitions())+4 {
		t.Fatalf("ordinary installed definitions = %d, want %d", beads, len(previewTypeDefinitions())+4)
	}
	for _, fact := range snapshot.Types {
		document, present := byID[fact.Definition.URI]
		if !present || document.Kind() != graph.KindBead || document.Revision() != fact.Definition.Revision {
			t.Fatalf("Type index has a competing/missing ordinary record: %#v", fact)
		}
		shown, err := s.GetResourceDocument(ctx, document.ID(), document.Revision())
		if err != nil || !bytes.Equal(shown.NativeJSON(), document.NativeJSON()) {
			t.Fatalf("ordinary Show disagrees with captured Type Bead %s: %v", document.ID(), err)
		}
	}
	page, err := s.ListBeadDocuments(ctx, BeadListRequest{})
	if err != nil || page.HasMore || len(page.Items) != beads {
		t.Fatalf("ordinary List omitted Type Beads: %d/%d %v", len(page.Items), beads, err)
	}
	count, err := s.CountBeadDocuments(ctx, BeadListRequest{Limit: 1})
	if err != nil || count != beads {
		t.Fatalf("Count used paged length: %d/%d %v", count, beads, err)
	}
	for _, id := range []string{"type", "bead-type", "link-type", "refines"} {
		document, err := s.GetResourceDocument(ctx, id, "")
		if err != nil || document.ID() != s.ScopeURL()+id || document.Kind() != graph.KindBead {
			t.Fatalf("fresh ordinary intrinsic ID %s: %s %v", id, document.ID(), err)
		}
	}
}

func TestOrdinaryNativeWritesRetainAppliedExactTypeEditions(t *testing.T) {
	ctx, o, s := intrinsicHardeningStore(t)
	// Donna retained the existing Issue ID rule. Bind this positive to the
	// actual configured prefix, rather than silently forcing an arbitrary ID.
	var prefix string
	if err := s.db.QueryRowContext(ctx, "SELECT value FROM config WHERE `key`='issue_prefix'").Scan(&prefix); err != nil || prefix != "bd" {
		t.Fatalf("expected configured bd prefix, got %q: %v", prefix, err)
	}
	request := plainIssue("Hello world")
	request.Issue.ID = "bd-333"
	issue, err := s.CreateIssue(ctx, "", request)
	if err != nil {
		t.Fatal(err)
	}
	applied, present := issue.ResourceDocument()
	if !present || applied.ID() != s.ScopeURL()+"bd-333" || applied.Kind() != graph.KindBead {
		t.Fatal("ordinary Issue create did not return same-Tx full applied document")
	}
	expected, err := NativeIssueTypeURL(s.ScopeURL(), string(request.Issue.IssueType))
	if err != nil || applied.Type().URI != expected {
		t.Fatalf("native classification was renamed: %#v %v", applied.Type(), err)
	}
	definition, err := s.GetResourceDocument(ctx, expected, applied.Type().Revision)
	if err != nil || definition.ID() != expected || definition.Revision() != applied.Type().Revision {
		t.Fatalf("Issue affiliation has no same-identity ordinary Type Bead: %v", err)
	}
	shown, err := s.GetResourceDocument(ctx, "bd-333", issue.Revision)
	if err != nil || !bytes.Equal(shown.NativeJSON(), applied.NativeJSON()) {
		t.Fatalf("exact applied Issue differs from Show: %v", err)
	}
	memory, err := s.Create(ctx, CreateRequest{PublicID: "note", Title: "Before", Body: "Body", Actor: "native"})
	if err != nil {
		t.Fatal(err)
	}
	first, present := memory.ResourceDocument()
	if !present {
		t.Fatal("ordinary Memory create lost exact applied document")
	}
	updated, err := s.UpdateMemory(ctx, MemoryUpdateRequest{Path: "beads/note", ExpectedRevision: memory.Revision, Properties: Properties{Title: "After", Body: "Body"}, Actor: "native"})
	if err != nil {
		t.Fatal(err)
	}
	successor, present := updated.Memory.ResourceDocument()
	if !present || successor.Type() != first.Type() || successor.Revision() == first.Revision() {
		t.Fatal("unrelated Memory update selected a new Type edition or omitted applied result")
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := OpenExisting(ctx, o)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := reopened.Close(); err != nil {
			t.Error(err)
		}
	}()
	retained, err := reopened.GetResourceDocument(ctx, "note", first.Revision())
	if err != nil || !bytes.Equal(retained.NativeJSON(), first.NativeJSON()) {
		t.Fatalf("retained Memory exact pin/body changed after reopen: %v", err)
	}
	current, err := reopened.GetResourceDocument(ctx, "note", "")
	if err != nil || !bytes.Equal(current.NativeJSON(), successor.NativeJSON()) {
		t.Fatalf("reopened Memory lost accepted successor: %v", err)
	}
}

// The bare333 acceptance trace remains historically red. This ruling adds no
// public/backing-ID adapter, automatic ForceIDPrefix, or alias exemption.
func TestOrdinaryNativeExplicitIssueIDRefusesWrongPrefix(t *testing.T) {
	ctx, _, s := intrinsicHardeningStore(t)
	var prefix string
	if err := s.db.QueryRowContext(ctx, "SELECT value FROM config WHERE `key`='issue_prefix'").Scan(&prefix); err != nil || prefix != "bd" {
		t.Fatalf("expected configured bd prefix, got %q: %v", prefix, err)
	}
	before := workflowState(t, ctx, s)
	request := plainIssue("Hello world")
	request.Issue.ID = "333"
	if request.ForceIDPrefix {
		t.Fatal("fixture must exercise the ordinary prefix rule")
	}
	result, err := s.CreateIssue(ctx, "", request)
	if !errors.Is(err, storage.ErrPrefixMismatch) || !reflect.ValueOf(result).IsZero() {
		t.Fatalf("wrong-prefix create returned a partial result or wrong error: %+v %v", result, err)
	}
	if !reflect.DeepEqual(before, workflowState(t, ctx, s)) {
		t.Fatal("wrong-prefix create changed native/domain/retained/Type/reservation state")
	}
	shown, err := s.GetResourceDocument(ctx, "333", "")
	if !errors.Is(err, ErrNotFound) || shown.ID() != "" {
		t.Fatalf("refused ID acquired a public Resource: %s %v", shown.ID(), err)
	}
	if !reflect.DeepEqual(before, workflowState(t, ctx, s)) {
		t.Fatal("refused lookup changed state")
	}
}

// The public typed readers consume the same checked stored Type set as the
// common reader, while keeping native Issue fields and exact private locators.
func TestNativeIssueReadsUseCheckedTypeCapture(t *testing.T) {
	ctx, _, s := intrinsicHardeningStore(t)
	request := plainIssue("Checked native read")
	request.Issue.ID = "bd-read"
	created, err := s.CreateIssue(ctx, "beads/private-native-read", request)
	if err != nil {
		t.Fatal(err)
	}
	applied, present := created.ResourceDocument()
	if !present {
		t.Fatal("creation did not retain its exact applied Type edition")
	}
	before, err := canonicalJSON(created)
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"direct", "selected-path", "selected-native"} {
		t.Run(mode, func(t *testing.T) {
			var actual IssueRecord
			var err error
			switch mode {
			case "direct":
				actual, err = s.ShowIssue(ctx, "beads/private-native-read")
			case "selected-path":
				actual, err = s.ShowIssueSelected(ctx, IssueSelector{Path: "beads/private-native-read", Exact: true})
			case "selected-native":
				actual, err = s.ShowIssueSelected(ctx, IssueSelector{Path: "beads/bd-read", NativeID: "bd-read", Exact: true})
			}
			if err != nil {
				t.Fatal(err)
			}
			encoded, err := canonicalJSON(actual)
			if err != nil || !bytes.Equal(before, encoded) || actual.ID != s.ScopeURL()+"bd-read" {
				t.Fatalf("typed read changed native identity/domain state: %s %v", encoded, err)
			}
		})
	}
	current, err := s.GetResourceDocument(ctx, created.ID, created.Revision)
	if err != nil || !bytes.Equal(current.NativeJSON(), applied.NativeJSON()) || current.Type() != applied.Type() {
		t.Fatalf("checked common and native read state diverged: %v", err)
	}
	t.Run("missing", func(t *testing.T) {
		direct, err := s.ShowIssue(ctx, "beads/missing-native-read")
		if !errors.Is(err, ErrNotFound) || !reflect.ValueOf(direct).IsZero() {
			t.Fatalf("missing path produced a record: %+v %v", direct, err)
		}
		selected, err := s.ShowIssueSelected(ctx, IssueSelector{Path: "beads/bd-missing-read", NativeID: "bd-missing-read", Exact: true})
		if !errors.Is(err, ErrNotFound) || !reflect.ValueOf(selected).IsZero() {
			t.Fatalf("missing exact native ID produced a record: %+v %v", selected, err)
		}
	})
	t.Run("non-issue", func(t *testing.T) {
		selected, err := s.ShowIssueSelected(ctx, IssueSelector{Path: "beads/bead-type", NativeID: "bead-type", Exact: true})
		if !errors.Is(err, ErrCapabilityUnavailable) || !reflect.ValueOf(selected).IsZero() {
			t.Fatalf("installed Type Bead became a native Issue: %+v %v", selected, err)
		}
	})
}

func TestNativeIssueReadRefusesCorruptRecordedPin(t *testing.T) {
	for _, mode := range []string{"unknown-edition", "owned-capture"} {
		t.Run(mode, func(t *testing.T) {
			ctx, _, s := intrinsicHardeningStore(t)
			created, err := s.CreateIssue(ctx, "beads/corrupt-native-read", plainIssue("Corrupt read"))
			if err != nil {
				t.Fatal(err)
			}
			if mode == "unknown-edition" {
				var raw []byte
				if err = s.db.QueryRowContext(ctx, "SELECT revision_context FROM graph_preview_issue_versions WHERE path=? AND version=?", "beads/corrupt-native-read", created.Revision).Scan(&raw); err != nil {
					t.Fatal(err)
				}
				value, pin, err := decodeNativeIssueTypeEdition(raw)
				if err != nil {
					t.Fatal(err)
				}
				if pin.Revision == "ffffffffffffffffffffffffffffffff" {
					t.Fatal("fixture needs a distinct unknown edition")
				}
				pin.Revision = "ffffffffffffffffffffffffffffffff"
				if _, err = s.GetResourceDocument(ctx, pin.URI, pin.Revision); !errors.Is(err, ErrVersionUnknown) {
					t.Fatalf("corruption target was not an unknown stored Type edition: %v", err)
				}
				encoded, err := encodeNativeIssueTypeEdition(value, pin)
				if err != nil {
					t.Fatal(err)
				}
				if _, err = s.db.ExecContext(ctx, "UPDATE graph_preview_issue_versions SET revision_context=? WHERE path=? AND version=?", encoded, "beads/corrupt-native-read", created.Revision); err != nil {
					t.Fatal(err)
				}
			} else {
				if _, err = s.db.ExecContext(ctx, "UPDATE graph_preview_issue_versions SET owned=? WHERE path=? AND version=?", []byte(`[null]`), "beads/corrupt-native-read", created.Revision); err != nil {
					t.Fatal(err)
				}
			}
			direct, err := s.ShowIssue(ctx, "beads/corrupt-native-read")
			if !errors.Is(err, ErrInvalidStore) || !reflect.ValueOf(direct).IsZero() {
				t.Fatalf("corrupt retained state produced native record: %+v %v", direct, err)
			}
			selected, err := s.ShowIssueSelected(ctx, IssueSelector{Path: "beads/corrupt-native-read", Exact: true})
			if !errors.Is(err, ErrInvalidStore) || !reflect.ValueOf(selected).IsZero() {
				t.Fatalf("corrupt retained state produced selected record: %+v %v", selected, err)
			}
		})
	}
}
