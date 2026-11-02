package httpclient

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/steveyegge/beads/internal/httpapi/apigen"
	"github.com/steveyegge/beads/internal/httpclient/encode"
	"github.com/steveyegge/beads/internal/httpclient/wire"
	"github.com/steveyegge/beads/issueops"
)

// Count every read/handshake seam as well as the operation-shaped writes the
// embedded stub records. A refusal must not depend on a server handshake.
type metadataOperationsWire struct {
	*stubWire
	preflights int
	reads      int
	contexts   int
}

func (w *metadataOperationsWire) Preflight(context.Context, string) error {
	w.preflights++
	return errors.New("unexpected metadata preflight")
}

func (w *metadataOperationsWire) Do(context.Context, wire.Request, any) error {
	w.reads++
	return errors.New("unexpected metadata read")
}

func (w *metadataOperationsWire) ServerContext(context.Context) (*apigen.ContextResponse, error) {
	w.contexts++
	return nil, errors.New("unexpected metadata handshake")
}

func TestMetadataOperationsRefuseBeforeWire(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value json.RawMessage
	}{
		{"ordered operations", json.RawMessage(`[{"op":"add","path":"/reviewed","value":true}]`)},
		{"empty list", json.RawMessage(`[]`)},
		{"nil", nil},
		{"empty bytes", json.RawMessage{}},
		{"null", json.RawMessage(`null`)},
		{"malformed", json.RawMessage(`[{`)},
	} {
		for _, op := range []encode.Op{encode.OpUpdateIssue, encode.OpApplyBatch} {
			t.Run(tc.name+"/"+string(op), func(t *testing.T) {
				w := &metadataOperationsWire{stubWire: &stubWire{}}
				s := New(testTarget(t), w, nil)
				patch := issueops.IssuePatch{
					Metadata: issueops.MetadataPatch{
						Operations: set(tc.value),
					},
				}
				if tc.name == "ordered operations" {
					patch.Title = set("must not land")
					// Even contradictory/malformed legacy metadata must not
					// hide the unsupported-operation refusal. Other cases use
					// Operations alone to exercise the metadata presence gate.
					patch.Metadata.Merge = set(json.RawMessage(`{"broken":`))
				}
				var err error
				if op == encode.OpUpdateIssue {
					var role issueops.Lifecycle
					role, err = s.IssueLifecycle()
					if err == nil {
						_, err = role.Update(t.Context(), issueops.UpdateRequest{
							Actor: "writer", IssueID: "bd-1", Patch: patch, Claim: true,
						})
					}
				} else {
					var role issueops.BatchApplier
					role, err = s.BatchApplier()
					if err == nil {
						// The valid write prefix must not land, and the lineage
						// item must not trigger a capability handshake first.
						req := applyOneCreate()
						req.Items = append(req.Items,
							issueops.ApplyItem{Kind: issueops.ItemDepAdd, DepAdd: &issueops.DepAddItem{
								Source: issueops.Ref{ID: "bd-1"}, Target: issueops.Ref{ID: "bd-2"},
								Type: "waits-for", ThreadID: "thread-1",
							}},
							issueops.ApplyItem{Kind: issueops.ItemUpdate, Update: &issueops.UpdateItem{
								Target: issueops.Ref{ID: "bd-1"}, Patch: patch,
							}},
						)
						_, err = role.ApplyBatch(t.Context(), req)
					}
				}
				var unsupported *ErrHTTPUnsupported
				if !errors.As(err, &unsupported) {
					t.Fatalf("error = %v, want ErrHTTPUnsupported", err)
				}
				var refused *encode.RefusedError
				if !errors.As(err, &refused) || refused.Row.ID != "W-MetadataPatch.Operations" || refused.Op != op {
					t.Fatalf("error = %v, want Operations refusal for %s", err, op)
				}
				if w.preflights != 0 || w.reads != 0 || w.contexts != 0 || len(w.calls) != 0 {
					t.Fatalf("refused request touched transport: preflights=%d reads=%d contexts=%d writes=%v",
						w.preflights, w.reads, w.contexts, w.calls)
				}
			})
		}
	}
}

func TestMetadataLegacyModesRemainSupported(t *testing.T) {
	for _, tc := range []struct {
		name  string
		patch issueops.MetadataPatch
		want  map[string]any
	}{
		{"omitted", issueops.MetadataPatch{}, nil},
		// Field.Value alone is not presence, even when it contains bytes.
		{"unset operations value", issueops.MetadataPatch{Operations: issueops.Field[json.RawMessage]{Value: json.RawMessage(`[]`)}}, nil},
		{"replace clear", issueops.MetadataPatch{Replace: set(json.RawMessage(nil))}, map[string]any{"replace": json.RawMessage(`{}`)}},
		{"replace", issueops.MetadataPatch{Replace: set(json.RawMessage(`{"n":1}`))}, map[string]any{"replace": json.RawMessage(`{"n":1}`)}},
		{"merge", issueops.MetadataPatch{Merge: set(json.RawMessage(`{"n":2}`))}, map[string]any{"merge": json.RawMessage(`{"n":2}`)}},
		{"set", issueops.MetadataPatch{Set: map[string]json.RawMessage{"n": json.RawMessage(`null`)}}, map[string]any{"set": map[string]json.RawMessage{"n": json.RawMessage(`null`)}}},
		{"unset", issueops.MetadataPatch{Unset: []string{"n"}}, map[string]any{"unset": []string{"n"}}},
	} {
		for _, op := range []encode.Op{encode.OpUpdateIssue, encode.OpApplyBatch} {
			t.Run(tc.name+"/"+string(op), func(t *testing.T) {
				w := &stubWire{update: &apigen.UpdateIssueResponse{Changed: true, Revision: "0"}}
				patch := issueops.IssuePatch{Title: set("supported edit"), Metadata: tc.patch}
				var document map[string]any
				if op == encode.OpUpdateIssue {
					role, err := stubStore(t, w).IssueLifecycle()
					if err != nil {
						t.Fatal(err)
					}
					if _, err := role.Update(t.Context(), issueops.UpdateRequest{Actor: "writer", IssueID: "bd-1", Patch: patch}); err != nil {
						t.Fatal(err)
					}
					document = w.lastPatch
				} else {
					if _, err := applyRole(t, w).ApplyBatch(t.Context(), issueops.ApplyBatchRequest{
						Actor: "writer", Items: []issueops.ApplyItem{{Kind: issueops.ItemUpdate, Update: &issueops.UpdateItem{
							Target: issueops.Ref{ID: "bd-1"}, Patch: patch,
						}}},
					}); err != nil {
						t.Fatal(err)
					}
					document = w.lastApply.Items[0].Update.Patch
				}
				if len(w.calls) != 1 || document["title"] != "supported edit" {
					t.Fatalf("write = %v, document = %#v", w.calls, document)
				}
				got, present := document["metadata"]
				if tc.want == nil {
					if present {
						t.Fatalf("omitted metadata encoded as %#v", got)
					}
				} else if !reflect.DeepEqual(got, tc.want) {
					t.Fatalf("metadata = %#v, want %#v", got, tc.want)
				}
			})
		}
	}
}
