//go:build cgo

package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/httpapi/bdpwire"
	"github.com/steveyegge/beads/internal/httpapi/graphread"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func TestGraphReadHTTPPreservesEndpointPins(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
			defer cancel()
			workspace, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			const scope = "https://authority.example/read/"
			options := graphstore.Options{Backend: backend, Database: fmt.Sprintf("graph_pin_%d", time.Now().UnixNano()), Branch: "main", DataDir: filepath.Join(workspace, "dolt"), Binding: graphstore.Binding{WorkspaceID: workspace, ScopeURL: scope, AuthorityID: "0123456789abcdef0123456789abcdef", SchemaVersion: graphstore.SchemaVersion}}
			if backend == "server" {
				port, err := strconv.Atoi(os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT"))
				if err != nil || port == 0 {
					t.Skip("ordinary released-server port not supplied")
				}
				options.ServerHost, options.ServerUser, options.ServerPort = "127.0.0.1", "root", port
			}
			if err := graphstore.Init(ctx, options); err != nil {
				t.Fatal(err)
			}
			store, err := graphstore.OpenExisting(ctx, options)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := store.Close(); err != nil {
					t.Error(err)
				}
			})
			source, err := store.Create(ctx, graphstore.CreateRequest{Path: "beads/source", Title: "Source", Body: "Source"})
			if err != nil {
				t.Fatal(err)
			}
			target, err := store.Create(ctx, graphstore.CreateRequest{Path: "beads/target", Title: "Target", Body: "Target"})
			if err != nil {
				t.Fatal(err)
			}
			created, err := store.AddInformationalLink(ctx, graphstore.LinkCreateRequest{Path: "links/citation", SourcePath: "beads/source", TargetPath: "beads/target", SourcePin: graphstore.EndpointPin{Current: true}, TargetPin: graphstore.EndpointPin{Revision: target.Revision}})
			if err != nil {
				t.Fatal(err)
			}
			constraint, err := store.Create(ctx, graphstore.CreateRequest{Path: "beads/rule", TypeURL: scope + "constraint", Title: "Rule"})
			if err != nil {
				t.Fatal(err)
			}
			owned, err := store.AddInformationalLink(ctx, graphstore.LinkCreateRequest{Path: "links/obligation", TypeURL: scope + "subject-to", SourcePath: "beads/source", TargetPath: "beads/rule", TargetPin: graphstore.EndpointPin{Current: true}, ExpectedSourceRevision: source.Revision})
			if err != nil {
				t.Fatal(err)
			}
			graph, err := NewGraphRead(graphread.New(store), scope)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(graph.Close)
			tokenPath := filepath.Join(workspace, "token")
			if err := os.WriteFile(tokenPath, []byte("pin-test-token\n"), 0600); err != nil {
				t.Fatal(err)
			}
			auth, err := NewTokenFileAuth(tokenPath)
			if err != nil {
				t.Fatal(err)
			}
			server := newTestServer(t, Config{GraphRead: graph, Auth: auth, AllowedHosts: []string{"transport.example"}})
			fetch := func(path string) []byte {
				t.Helper()
				request, err := http.NewRequestWithContext(ctx, "GET", server.base+path, nil)
				if err != nil {
					t.Fatal(err)
				}
				request.Host = "transport.example"
				request.Header.Set("Authorization", "Bearer pin-test-token")
				request.Header.Set("Accept", "application/json")
				response, err := http.DefaultClient.Do(request)
				if err != nil {
					t.Fatal(err)
				}
				raw, err := io.ReadAll(response.Body)
				_ = response.Body.Close()
				if err != nil {
					t.Fatal(err)
				}
				if response.StatusCode != http.StatusOK {
					t.Fatalf("HTTP %d: %s", response.StatusCode, raw)
				}
				return raw
			}
			raw := fetch("/read/links/citation")
			var wire bdpwire.LinkRecord
			if err := json.Unmarshal(raw, &wire); err != nil {
				t.Fatalf("invalid structured wire Reference: %s %v", raw, err)
			}
			if wire.ID != created.Link.ID || wire.Source.URI != source.ID || wire.Source.Revision != source.Revision || wire.Target.URI != target.ID || wire.Target.Revision != target.Revision {
				t.Fatalf("HTTP lost separate pin state: %+v", wire)
			}
			assertOwned := func(record bdpwire.BeadRecord) {
				t.Helper()
				links := record.OwnedLinks[scope+"subject-to"]
				if len(links) != 1 || links[0].ID != owned.Link.ID || links[0].Source.URI != source.ID || links[0].Source.Revision != "" || links[0].Target.URI != constraint.ID || links[0].Target.Revision != constraint.Revision {
					t.Fatalf("effective ownership projection lost pinned owned target: %+v", record)
				}
			}
			var owner bdpwire.BeadRecord
			if err := json.Unmarshal(fetch("/read/beads/source"), &owner); err != nil {
				t.Fatal(err)
			}
			assertOwned(owner)
			inventory, err := graphread.New(store).Inventory(ctx)
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, record := range inventory.Beads {
				if record.ID == source.ID {
					assertOwned(record)
					found = true
				}
			}
			if !found {
				t.Fatal("source absent from same-transaction inventory")
			}

		})
	}
}
