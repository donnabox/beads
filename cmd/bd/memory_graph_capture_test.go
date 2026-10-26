package main

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/memoryquery"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func memoryCaptureFixture(t *testing.T) (graphstore.CurrentFamilyRead, []memoryquery.CurrentRoot) {
	t.Helper()
	roots := []memoryquery.CurrentRoot{{Path: "beads/a", Depth: 1}}
	r := memoryPresentationFixture(t, roots, graph.DirectionOut, genericTestEdge{"ab", "a", "b"})
	for i, value := range r.Records {
		switch v := value.(type) {
		case graphstore.Record:
			v.Revision = v.Version
			r.Records[i] = v
		case graphstore.LinkRecord:
			v.Revision = v.Version
			r.Records[i] = v
		}
	}
	return r, roots
}

func TestMemoryCaptureContext(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*graphstore.CurrentFamilyRead, []memoryquery.CurrentRoot)
	}{
		{"root-budget", func(r *graphstore.CurrentFamilyRead, _ []memoryquery.CurrentRoot) { r.Graph.Views[0].Root.Depth++ }},
		{"record-summary-version", func(r *graphstore.CurrentFamilyRead, _ []memoryquery.CurrentRoot) {
			r.Graph.Nodes[0].Version = "different"
		}},
		{"record-summary-type", func(r *graphstore.CurrentFamilyRead, _ []memoryquery.CurrentRoot) {
			r.Graph.Nodes[0].TypeURL = genericTestScope + "types/memory"
		}},
		{"duplicate-record", func(r *graphstore.CurrentFamilyRead, _ []memoryquery.CurrentRoot) {
			r.Records = append(r.Records, r.Records[0])
		}},
		{"missing-record", func(r *graphstore.CurrentFamilyRead, _ []memoryquery.CurrentRoot) { r.Records = r.Records[1:] }},
		{"duplicate-membership", func(r *graphstore.CurrentFamilyRead, _ []memoryquery.CurrentRoot) {
			r.Graph.Views[0].Nodes = append(r.Graph.Views[0].Nodes, "beads/a")
		}},
		{"wrong-kind", func(r *graphstore.CurrentFamilyRead, _ []memoryquery.CurrentRoot) {
			r.Graph.Views[0].Nodes[0] = "links/ab"
		}},
		{"missing-endpoint", func(r *graphstore.CurrentFamilyRead, _ []memoryquery.CurrentRoot) {
			r.Graph.Views[0].Nodes = []string{"beads/a"}
		}},
		{"missing-aggregate-edge", func(r *graphstore.CurrentFamilyRead, _ []memoryquery.CurrentRoot) { r.Graph.Views[0].Links = nil }},
		{"unknown-frontier", func(r *graphstore.CurrentFamilyRead, _ []memoryquery.CurrentRoot) {
			r.Graph.Views[0].Frontier = []string{"beads/foreign"}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, roots := memoryCaptureFixture(t)
			if err := validateMemoryCurrentCapture(r, genericTestScope, roots); err != nil {
				t.Fatalf("baseline: %v", err)
			}
			tc.mutate(&r, roots)
			if err := validateMemoryCurrentCapture(r, genericTestScope, roots); !errors.Is(err, graphstore.ErrInvalidStore) {
				t.Fatalf("accepted malformed capture: %v", err)
			}
		})
	}
	r, roots := memoryCaptureFixture(t)
	before := r.Graph.Views[0].Root
	if err := validateMemoryCurrentCapture(r, genericTestScope, roots); err != nil || !reflect.DeepEqual(before, r.Graph.Views[0].Root) {
		t.Fatalf("valid captured graph changed: %v", err)
	}
	// The producer fixture deliberately carries a ghost Owned copy: it must not
	// become a selected edge or cause topology validation to re-traverse Owned.
	if len(r.Graph.Links) != 1 {
		t.Fatal("owned copy entered selection")
	}
}

func TestMemoryGraphEncodingAndErrors(t *testing.T) {
	raw, err := encodeMemoryGraphJSON(struct {
		Text string `json:"text"`
	}{"A & B <tag>"})
	if err != nil || string(raw) != `{"text":"A & B <tag>"}` {
		t.Fatalf("not common CLI escaping: %s %v", raw, err)
	}
	c := &cobra.Command{}
	var stderr bytes.Buffer
	c.SetErr(&stderr)
	err = memoryGraphOutputError(c, nil, errors.New("encoding failed"))
	var exit *exitError
	if !errors.As(err, &exit) || exit.Code != 1 || !bytes.Contains(stderr.Bytes(), []byte("output_error:")) {
		t.Fatalf("output error status: %v %s", err, &stderr)
	}
	for _, member := range []string{"beads", "links", "traversal", "scope"} {
		if !containsMemoryProtected(memoryGraphDescriptor().Protection.Required, member) {
			t.Fatalf("graph membership descriptor missing %s", member)
		}
	}
}
func containsMemoryProtected(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func TestMemoryCaptureAndTreePreserveCommonIDs(t *testing.T) {
	roots := []memoryquery.CurrentRoot{{Path: "rootless", Depth: 1}}
	reader := memoryPresentationReader{nodes: map[string]memoryquery.CurrentNode{
		"rootless":     {Path: "rootless", TypeURL: genericTestScope + "types/policy", Version: "current"},
		"links/memory": {Path: "links/memory", TypeURL: genericTestScope + "types/policy", Version: "current"},
	}}
	source, err := graph.NewInScopeRef("rootless", "current")
	if err != nil {
		t.Fatal(err)
	}
	target, err := graph.NewInScopeRef("links/memory", "")
	if err != nil {
		t.Fatal(err)
	}
	revision, err := graph.NewRevision("link-current")
	if err != nil {
		t.Fatal(err)
	}
	link, err := graph.NewLink(graph.LinkSpec{Path: "beads/edge", TypeURL: genericTestScope + "types/references", Revision: revision, Source: source, Target: target})
	if err != nil {
		t.Fatal(err)
	}
	reader.links = []graph.Link{link}
	walk, err := memoryquery.CurrentFamily(context.Background(), reader, func(context.Context, string) (bool, error) { return true, nil }, roots, graph.DirectionOut, memoryquery.Limits{MaxNodes: 10, MaxLinks: 10, MaxInspected: 100})
	if err != nil {
		t.Fatal(err)
	}
	read := graphstore.CurrentFamilyRead{Graph: walk}
	for _, node := range walk.Nodes {
		read.Records = append(read.Records, graphstore.Record{ID: genericTestScope + node.Path, Type: node.TypeURL, Revision: node.Version, Version: node.Version, Properties: graphstore.Properties{Title: node.Path}})
	}
	read.Records = append(read.Records, graphstore.LinkRecord{ID: genericTestScope + "beads/edge", Type: link.TypeURL(), Revision: "link-current", Version: "link-current", Source: source.URL(genericTestScope), Target: target.URL(genericTestScope), SourceRevision: source.Pin()})
	if err := validateMemoryCurrentCapture(read, genericTestScope, roots); err != nil {
		t.Fatal(err)
	}
	text, err := renderMemoryCurrentGraph(read, genericTestScope, graph.DirectionOut, graphstore.PreviewCurrentReadByteLimit)
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"rootless", "links/memory", "beads/edge", "types/policy", "types/references"} {
		if !strings.Contains(text, id) {
			t.Fatalf("tree elided literal identity %q: %s", id, text)
		}
	}
}
