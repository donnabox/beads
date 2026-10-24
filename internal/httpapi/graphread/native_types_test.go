//go:build cgo

package graphread

import (
	"context"
	"fmt"
	"net/url"
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

func TestNativeConcreteTypesWireAndRetainedProjection(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 3*time.Minute)
			defer cancel()
			workspace, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			const scope = "https://example.test/native-wire/"
			o := graphstore.Options{Backend: backend, Database: fmt.Sprintf("native_wire_%d", time.Now().UnixNano()), DataDir: filepath.Join(workspace, "dolt"), Branch: "main", Binding: graphstore.Binding{WorkspaceID: workspace, ScopeURL: scope, AuthorityID: "0123456789abcdef0123456789abcdef", SchemaVersion: graphstore.SchemaVersion}}
			if backend == "server" {
				port := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
				if port == "" {
					t.Skip("private released-Dolt server port required")
				}
				o.ServerPort, err = strconv.Atoi(port)
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
			var original graphstore.IssueRecord
			for _, path := range []string{"beads/a", "beads/b", "beads/c"} {
				created, err := s.CreateIssue(ctx, path, issueops.CreateRequest{Actor: "tester", Issue: &issueops.Issue{Title: path, IssueType: "task", Status: "open", Priority: 2}})
				if err != nil {
					t.Fatal(err)
				}
				if path == "beads/a" {
					original = created
				}
			}
			if _, err := s.Create(ctx, graphstore.CreateRequest{Path: "beads/memory", Title: "Memory"}); err != nil {
				t.Fatal(err)
			}
			unknown, _ := graphstore.NativeDependencyTypeURL(scope, "imported-unfamiliar")
			if _, err := s.AddDependency(ctx, graphstore.DependencyRequest{Path: "links/owned", SourcePath: "beads/a", TargetPath: "beads/b", Type: "imported-unfamiliar", Actor: "tester"}); err != nil {
				t.Fatal(err)
			}
			current, err := s.ShowIssue(ctx, "beads/a")
			if err != nil {
				t.Fatal(err)
			}
			bug, _ := graphstore.NativeIssueTypeURL(scope, "bug")
			if _, err := s.UpdateIssue(ctx, graphstore.UpdateIssueRequest{Path: "beads/a", Actor: "tester", ExpectedRevision: current.Revision, TypeURL: bug}); err != nil {
				t.Fatal(err)
			}
			r := New(s)
			inventory, err := r.Inventory(ctx)
			if err != nil {
				t.Fatal(err)
			}
			for _, path := range []string{"beads/a", "beads/b", "beads/c"} {
				value, err := r.Resource(ctx, path)
				if err != nil {
					t.Fatal(err)
				}
				got := value.(bdpwire.BeadRecord)
				var listed *bdpwire.BeadRecord
				for i := range inventory.Beads {
					if inventory.Beads[i].ID == scope+path {
						listed = &inventory.Beads[i]
						break
					}
				}
				if listed == nil || !reflect.DeepEqual(got, *listed) {
					t.Fatalf("GET/inventory disagreement for %s: %+v / %+v", path, got, listed)
				}
				_, hasUnknown := got.OwnedLinks[unknown]
				if hasUnknown != (path == "beads/a") {
					t.Fatalf("unrelated native ownership bucket on %s: %+v", path, got.OwnedLinks)
				}
				if path == "beads/a" && (got.Type != bug || string(got.Properties["issue_type"]) != `"bug"`) {
					t.Fatalf("contradictory native declaration: %+v", got)
				}
			}
			task, _ := graphstore.NativeIssueTypeURL(scope, "task")
			retained, err := r.ResourceVersion(ctx, "beads/a", original.Revision)
			if err != nil {
				t.Fatal(err)
			}
			if got := retained.(bdpwire.BeadRecord); got.Type != task || string(got.Properties["issue_type"]) != `"task"` {
				t.Fatalf("old affiliation rewritten: %+v", got)
			}
			descriptor, err := r.Type(ctx, strings.TrimPrefix(task, scope))
			if err != nil || len(descriptor.ConformsTo) != 1 || descriptor.ConformsTo[0] != graphstore.IssueTypeURL(scope) {
				t.Fatalf("concrete parent missing: %+v %v", descriptor, err)
			}
			base, err := r.Type(ctx, "types/issue")
			if err != nil || !strings.Contains(strings.ToLower(base.Description), "abstract") {
				t.Fatalf("abstract status hidden: %+v %v", base, err)
			}
			for _, tc := range []struct {
				parameter, value string
				want             int
			}{{"", "", 4}, {"type", graphstore.IssueTypeURL(scope), 0}, {"conformsTo", graphstore.IssueTypeURL(scope), 3}, {"type", task, 2}, {"conformsTo", task, 2}} {
				parameters := url.Values{}
				if tc.parameter != "" {
					parameters.Set(tc.parameter, tc.value)
				}
				selection, err := CompileCollection(scope, "beads", parameters, SelectorLimits{})
				if err != nil {
					t.Fatal(err)
				}
				records, err := selection.Select(inventory)
				if err != nil || len(records) != tc.want {
					t.Fatalf("unchanged BDP wire selector %+v: count=%d err=%v", tc, len(records), err)
				}
			}
		})
	}
}
