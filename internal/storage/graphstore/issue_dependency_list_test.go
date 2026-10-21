//go:build cgo

package graphstore

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/steveyegge/beads/internal/types"
)

// A small query cap exercises the same SQL predicate as the public 1,000-Link
// cap without creating thousands of Issues. Filtering must precede the cap.
func TestIssueDependencyListFiltersBeforeIncidentLimit(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			ctx, options, store := issueListFixture(t, backend)
			for _, path := range []string{"beads/target", "beads/tracker", "beads/blocker-a", "beads/blocker-b"} {
				if _, err := store.CreateIssue(ctx, path, plainIssue(path)); err != nil {
					t.Fatal(err)
				}
			}
			dependencyType := DependencyTypeURL(options.Binding.ScopeURL)
			for _, edge := range []struct {
				source string
				kind   types.DependencyType
			}{
				{"beads/tracker", types.DepTracks},
				{"beads/blocker-a", types.DepBlocks},
				{"beads/blocker-b", types.DepBlocks},
			} {
				if _, err := store.AddDependency(ctx, DependencyRequest{SourcePath: edge.source, TargetPath: "beads/target", Type: edge.kind, TypeURL: dependencyType, Actor: "author"}); err != nil {
					t.Fatal(err)
				}
			}
			read := func(kind string) ([]LinkRecord, error) {
				t.Helper()
				var links []LinkRecord
				err := store.withTx(ctx, false, func(tx *sql.Tx) error {
					var err error
					links, err = store.listIssueDependenciesInTx(ctx, tx, "beads/target", "in", kind, 1)
					return err
				})
				return links, err
			}
			got, err := read(string(types.DepTracks))
			if err != nil || len(got) != 1 || got[0].Properties["dep_type"] != string(types.DepTracks) || got[0].Source != options.Binding.ScopeURL+"beads/tracker" {
				t.Fatalf("selective inbound kind must fit below cap: %+v, %v", got, err)
			}
			for _, kind := range []string{"", string(types.DepBlocks)} {
				if _, err := read(kind); !errors.Is(err, ErrLimitExceeded) {
					t.Fatalf("kind %q with two or more matches must refuse at cap: %v", kind, err)
				}
			}
		})
	}
}
