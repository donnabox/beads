//go:build cgo

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// This is a deterministic interleaving across the CLI's new composition seam,
// not a substitute for the writer's existing forced transaction-overlap tests.
// The server is caller owned and disposable; provisioning is deliberately serial.
func TestGraphRememberSelectedUpdateInterveningWrites(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
			defer cancel()
			workspace, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			o := graphstore.Options{Backend: backend, DataDir: filepath.Join(workspace, "dolt"),
				Database: fmt.Sprintf("selected_update_%d", time.Now().UnixNano()), Branch: "main", IssuePrefix: "edit",
				Binding: graphstore.Binding{WorkspaceID: workspace, ScopeURL: "https://example.invalid/selected-update/",
					AuthorityID: "0123456789abcdef0123456789abcdef", SchemaVersion: graphstore.SchemaVersion}}
			if backend == "server" {
				port := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
				if port == "" {
					t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for caller-owned disposable server checks")
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
			defer func() {
				if err := s.Close(); err != nil {
					t.Error(err)
				}
			}()
			target, err := s.Create(ctx, graphstore.CreateRequest{Path: "beads/target", Title: "Target", Body: "untouched"})
			if err != nil {
				t.Fatal(err)
			}
			for _, competing := range []string{"title", "owned-link"} {
				t.Run(competing, func(t *testing.T) {
					path := "beads/" + competing
					original, err := s.Create(ctx, graphstore.CreateRequest{Path: path, Title: "Original title — 雪", Body: "original"})
					if err != nil {
						t.Fatal(err)
					}
					observed, err := s.Read(ctx, path)
					if err != nil {
						t.Fatal(err)
					}
					properties, err := graphPreviewRememberUpdateProperties(observed, "my new body", nil, original.Revision)
					if err != nil || properties.Title != original.Properties.Title || properties.Body != "my new body" {
						t.Fatalf("composition: %+v %v", properties, err)
					}
					var winner graphstore.Record
					var link graphstore.LinkRecord
					if competing == "title" {
						changed, err := s.UpdateMemory(ctx, graphstore.MemoryUpdateRequest{Path: path,
							Properties:       graphstore.Properties{Title: "Concurrent title", Body: original.Properties.Body},
							ExpectedRevision: original.Revision, Actor: "winner"})
						if err != nil || !changed.Changed {
							t.Fatalf("competing title: %+v %v", changed, err)
						}
						winner = changed.Memory
					} else {
						changed, err := s.AddInformationalLink(ctx, graphstore.LinkCreateRequest{Path: "links/owned",
							SourcePath: path, TargetPath: "beads/target", ExpectedSourceRevision: original.Revision,
							Properties: map[string]any{"note": "concurrent context"}, Actor: "winner"})
						if err != nil || !changed.Changed {
							t.Fatalf("competing Link: %+v %v", changed, err)
						}
						winner, link = changed.Source.(graphstore.Record), changed.Link
					}
					result, err := s.UpdateMemory(ctx, graphstore.MemoryUpdateRequest{Path: path, Properties: properties,
						ExpectedRevision: original.Revision, Actor: "loser"})
					if !errors.Is(err, graphstore.ErrConflict) || !reflect.DeepEqual(result, graphstore.MemoryMutationResult{}) {
						t.Fatalf("intervening change was not refused atomically: %+v %v", result, err)
					}
					if got, err := s.Read(ctx, path); err != nil || !reflect.DeepEqual(got, winner) {
						t.Fatalf("winner changed: %+v %v", got, err)
					}
					if got, err := s.ReadVersion(ctx, path, original.Version); err != nil || !reflect.DeepEqual(got, original) {
						t.Fatalf("original retained state changed: %+v %v", got, err)
					}
					if got, err := s.ReadVersion(ctx, path, winner.Version); err != nil || !reflect.DeepEqual(got, winner) {
						t.Fatalf("winner retained state changed: %+v %v", got, err)
					}
					if got, err := s.Read(ctx, "beads/target"); err != nil || !reflect.DeepEqual(got, target) {
						t.Fatalf("target changed: %+v %v", got, err)
					}
					if competing == "owned-link" {
						if got, err := s.Read(ctx, "links/owned"); err != nil || !reflect.DeepEqual(got, link) {
							t.Fatalf("competing Link changed: %+v %v", got, err)
						}
					}
				})
			}
		})
	}
}
