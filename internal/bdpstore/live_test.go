package bdpstore_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/bdp/bdpclient"
	"github.com/steveyegge/beads/internal/bdpstore"
	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/issueops"
)

// TestBDPStoreCRUDAgainstCleanroom drives the adapter's role interfaces against
// a RUNNING clean-room BDP server, proving the Issue↔BDP-body domain mapping
// round-trips end to end. Skipped unless BDP_SMOKE_ADDR names a server (e.g.
// "http://127.0.0.1:8799"), so it is inert in CI. It constructs the adapter
// DIRECTLY — no cmd/bd, no store_factory, no backend registration (the held
// wiring). Stand up the server with `lumen --store inmem: serve --bind
// 127.0.0.1:8799` (bdp-cleanroom-v023); the server is stateful, so each run uses
// a fresh id.
func TestBDPStoreCRUDAgainstCleanroom(t *testing.T) {
	base := os.Getenv("BDP_SMOKE_ADDR")
	if base == "" {
		t.Skip("set BDP_SMOKE_ADDR to a running clean-room server to run this test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	store := bdpstore.New(bdpclient.New(base))
	reader, err := store.IssueReader()
	if err != nil {
		t.Fatalf("IssueReader: %v", err)
	}
	lifecycle, err := store.IssueLifecycle()
	if err != nil {
		t.Fatalf("IssueLifecycle: %v", err)
	}
	deleter, err := store.Deleter()
	if err != nil {
		t.Fatalf("Deleter: %v", err)
	}

	id := fmt.Sprintf("bdpstore-%d", time.Now().UnixNano())

	// CREATE an Issue via the Lifecycle role → maps onto BDP create.
	cres, err := lifecycle.Create(ctx, issueops.CreateRequest{
		Actor: "sheila",
		Issue: &types.Issue{
			ID:        id,
			Title:     "adapter round-trip",
			Status:    types.Status("open"),
			Priority:  2,
			IssueType: types.IssueType("task"),
		},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if cres.Issue == nil || cres.Issue.ID != id {
		t.Fatalf("Create result = %+v, want issue id %s", cres.Issue, id)
	}

	// GET via the Reader role → the Issue fields round-trip through the BDP body.
	got, err := reader.Get(ctx, issueops.GetRequest{ID: id})
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Title != "adapter round-trip" {
		t.Errorf("Title = %q, want 'adapter round-trip'", got.Title)
	}
	if got.Priority != 2 {
		t.Errorf("Priority = %d, want 2", got.Priority)
	}
	if string(got.Status) != "open" {
		t.Errorf("Status = %q, want open", got.Status)
	}

	// LIST via the Reader role → the created issue appears (ids hydrated to Issues).
	listed, err := reader.List(ctx, issueops.ListRequest{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	found := false
	for _, it := range listed.Items {
		if it.ID == id {
			found = true
			if it.Title != "adapter round-trip" {
				t.Errorf("listed issue Title = %q, want 'adapter round-trip'", it.Title)
			}
		}
	}
	if !found {
		t.Errorf("List omitted the created issue %s (got %d items)", id, len(listed.Items))
	}

	// A Status filter that does not match drops it (client-side).
	filtered, err := reader.List(ctx, issueops.ListRequest{Status: "closed"})
	if err != nil {
		t.Fatalf("List(status=closed): %v", err)
	}
	for _, it := range filtered.Items {
		if it.ID == id {
			t.Errorf("status=closed filter should have excluded the open issue %s", id)
		}
	}

	// A missing issue maps the BDP fault onto issueops.ErrNotFound.
	if _, err := reader.Get(ctx, issueops.GetRequest{ID: "absent-" + id}); !errors.Is(err, issueops.ErrNotFound) {
		t.Errorf("Get missing: err = %v, want ErrNotFound", err)
	}

	// DELETE via the Deleter role, then confirm it is gone.
	dres, err := deleter.Delete(ctx, issueops.DeleteRequest{Actor: "sheila", IDs: []string{id}})
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if dres.Deleted != 1 {
		t.Errorf("Deleted = %d, want 1", dres.Deleted)
	}
	if _, err := reader.Get(ctx, issueops.GetRequest{ID: id}); !errors.Is(err, issueops.ErrNotFound) {
		t.Errorf("Get after delete: err = %v, want ErrNotFound", err)
	}
}
