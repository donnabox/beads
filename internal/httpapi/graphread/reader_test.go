//go:build cgo

package graphread

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/httpapi/bdpwire"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/issueops"
)

// This is the real storage-to-wire boundary. Writers use normal initialization
// and public store operations; no test schema or SQL payload is seeded. It is
// not an installed CLI or HTTP interoperability demonstration.
func TestAuthoritativeRecordsProjectToPublicWire(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
			defer cancel()
			workspace, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			o := graphstore.Options{Backend: backend, Database: fmt.Sprintf("graph_read_%d", time.Now().UnixNano()), DataDir: filepath.Join(workspace, "dolt"), Branch: "main", Binding: graphstore.Binding{WorkspaceID: workspace, ScopeURL: "https://example.test/read/", AuthorityID: "0123456789abcdef0123456789abcdef", SchemaVersion: graphstore.SchemaVersion}}
			if backend == "server" {
				if os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT") == "" {
					t.Skip("ordinary shared-server Dolt port not configured")
				}
				o.ServerPort, err = strconv.Atoi(os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT"))
				if err != nil {
					t.Fatal(err)
				}
				o.ServerHost, o.ServerUser = "127.0.0.1", "root"
			}
			if err := graphstore.Init(ctx, o); err != nil {
				t.Fatal(err)
			}
			s, err := graphstore.OpenExisting(ctx, o)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := s.Close(); err != nil {
					t.Error(err)
				}
			})
			r := New(s)
			memory, err := s.Create(ctx, graphstore.CreateRequest{Path: "beads/plan", Title: "Memory — 記憶", Body: "", Actor: "human:Donna <unchanged>"})
			if err != nil {
				t.Fatal(err)
			}
			other, err := s.Create(ctx, graphstore.CreateRequest{Path: "beads/other", Title: "Target"})
			if err != nil {
				t.Fatal(err)
			}
			createIssue := func(path string) graphstore.IssueRecord {
				v, err := s.CreateIssue(ctx, path, issueops.CreateRequest{Issue: &issueops.Issue{Title: path, Status: "open", IssueType: "task", Priority: 2}, Actor: "agent:test"})
				if err != nil {
					t.Fatal(err)
				}
				return v
			}
			issue := createIssue("beads/task")
			_ = createIssue("beads/blocker")
			readBead := func(path string) bdpwire.BeadRecord {
				v, err := r.Resource(ctx, path)
				if err != nil {
					t.Fatal(err)
				}
				bead, ok := v.(bdpwire.BeadRecord)
				if !ok {
					t.Fatalf("not a Bead: %T", v)
				}
				checkWire(t, "beadRecord", bead)
				return bead
			}
			first := readBead("beads/plan")
			if first.ID != memory.ID || first.Revision != memory.Revision || first.OwnedLinks == nil || len(first.OwnedLinks) != 0 || string(first.Properties["body"]) != `""` || first.Attribution.Principal != memory.Attribution.Actor {
				t.Fatalf("Memory projection changed identity/content: %+v", first)
			}
			if readBead("beads/other").Attribution != nil {
				t.Fatal("fabricated actor")
			}
			firstIssue := readBead("beads/task")
			if group, ok := firstIssue.OwnedLinks[graphstore.DependencyTypeURL(o.Binding.ScopeURL)]; !ok || len(group) != 0 {
				t.Fatal("missing declared empty owned group")
			}
			wantProperties, _ := json.Marshal(issue.Properties)
			gotProperties, _ := json.Marshal(firstIssue.Properties)
			var a, b any
			_ = json.Unmarshal(wantProperties, &a)
			_ = json.Unmarshal(gotProperties, &b)
			if !reflect.DeepEqual(a, b) {
				t.Fatal("Issue properties were filtered or changed")
			}
			if _, err := s.AddDependency(ctx, graphstore.DependencyRequest{SourcePath: "beads/task", TargetPath: "beads/blocker", Path: "links/blocks", Actor: "agent:test"}); err != nil {
				t.Fatal(err)
			}
			var created graphstore.LinkRecord
			for _, path := range []string{"links/z", "links/a"} {
				v, err := s.AddInformationalLink(ctx, graphstore.LinkCreateRequest{Path: path, SourcePath: "beads/plan", TargetPath: "beads/other", Actor: "human:Donna", UnconditionalSource: true, Properties: map[string]any{"note": path}})
				if err != nil {
					t.Fatal(err)
				}
				created = v.Link
			}
			for _, path := range []string{"links/blocks", "links/a"} {
				v, err := r.Resource(ctx, path)
				if err != nil {
					t.Fatal(err)
				}
				checkWire(t, "linkRecord", v)
			}
			linked := readBead("beads/plan")
			group := linked.OwnedLinks[created.Type]
			if len(group) != 2 || !strings.HasSuffix(group[0].ID, "/links/a") || group[1].Source.URI != memory.ID || linked.Revision == first.Revision {
				t.Fatal("owned projection ordering or membership differs")
			}
			if readBead("beads/other").Revision != other.Revision {
				t.Fatal("incoming Link changed target")
			}
			withBlock := readBead("beads/task")
			if len(withBlock.OwnedLinks[graphstore.DependencyTypeURL(o.Binding.ScopeURL)]) != 1 || withBlock.Revision == firstIssue.Revision {
				t.Fatal("Issue owned state did not advance")
			}
			for _, typ := range []string{memory.Type, issue.Type, created.Type, graphstore.DependencyTypeURL(o.Binding.ScopeURL)} {
				d, err := r.Type(ctx, strings.TrimPrefix(typ, o.Binding.ScopeURL))
				if err != nil {
					t.Fatal(err)
				}
				if d.ID != typ {
					t.Fatal("descriptor identity changed")
				}
				checkWire(t, "typeDescriptor", d)
			}
			if _, err := s.UpdateLink(ctx, graphstore.LinkUpdateRequest{Path: "links/a", Actor: "human:Donna", Unconditional: true, UnconditionalSource: true, Properties: map[string]any{"note": "revised"}}); err != nil {
				t.Fatal(err)
			}
			updated := readBead("beads/plan")
			if updated.Revision == linked.Revision || string(updated.OwnedLinks[created.Type][0].Properties["note"]) != `"revised"` {
				t.Fatal("current owned update missing")
			}
			if _, err := s.Unlink(ctx, graphstore.LinkDeleteRequest{Path: "links/a", Actor: "human:Donna", Unconditional: true, UnconditionalSource: true}); err != nil {
				t.Fatal(err)
			}
			if _, err := r.Resource(ctx, "links/a"); !errors.Is(err, graphstore.ErrGone) {
				t.Fatalf("deleted state lost: %v", err)
			}
			if _, err := r.Resource(ctx, "links/missing"); !errors.Is(err, graphstore.ErrNotFound) {
				t.Fatalf("absence changed: %v", err)
			}
			remaining := readBead("beads/plan")
			if len(remaining.OwnedLinks[created.Type]) != 1 || remaining.Revision == updated.Revision {
				t.Fatal("deleted Link retained in current owned projection")
			}
			if err := s.Close(); err != nil {
				t.Fatal(err)
			}
			s, err = graphstore.OpenExisting(ctx, o)
			if err != nil {
				t.Fatal(err)
			}
			r = New(s)
			if !reflect.DeepEqual(remaining, readBead("beads/plan")) {
				t.Fatal("projection changed after reopen")
			}
		})
	}
}

func checkWire(t *testing.T, kind string, value any) {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var decoded any
	switch kind {
	case "beadRecord":
		decoded = &bdpwire.BeadRecord{}
	case "linkRecord":
		decoded = &bdpwire.LinkRecord{}
	case "typeDescriptor":
		decoded = &bdpwire.TypeDescriptor{}
	}
	if err := bdpwire.Unmarshal(raw, decoded); err != nil {
		t.Fatal(err)
	}
	var members map[string]json.RawMessage
	if err := json.Unmarshal(raw, &members); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"version", "owned", "actor", "recordedAt"} {
		if _, ok := members[name]; ok {
			t.Fatalf("private member leaked: %s", name)
		}
	}
	receipt, _ := json.Marshal(map[string]any{"kind": kind, "value": value})
	t.Logf("WIRE_RECEIPT %s", receipt)
}
