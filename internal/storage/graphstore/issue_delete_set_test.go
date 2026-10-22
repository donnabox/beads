//go:build cgo

package graphstore

import (
	"database/sql"
	"errors"
	"reflect"
	"testing"
)

func TestIssueDeleteSetPlanUsesOneBoundedGraphSnapshot(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			f := setupDependencyUnlink(t, backend, 2)
			before := workflowState(t, f.ctx, f.store)
			plan := func(paths []string, cascade bool) (issueDeleteSetPlan, error) {
				var got issueDeleteSetPlan
				err := f.store.withTx(f.ctx, false, func(tx *sql.Tx) error {
					var err error
					got, err = f.store.planIssueDeleteSetInTx(f.ctx, tx, paths, cascade)
					return err
				})
				return got, err
			}
			flat, err := plan([]string{"beads/first", "beads/first"}, false)
			if err != nil || len(flat.Issues) != 1 || flat.Issues[0].ID != f.targets[0].ID ||
				len(flat.Links) != 1 || flat.Links[0].ID != f.links[0].ID || len(flat.Native.All) != 1 {
				t.Fatalf("flat plan: %+v %v", flat, err)
			}
			cascade, err := plan([]string{"beads/first"}, true)
			if err != nil || len(cascade.Issues) != 2 || cascade.Issues[0].ID != f.targets[0].ID ||
				cascade.Issues[1].ID != f.source.ID || len(cascade.Links) != 2 ||
				cascade.Links[0].ID != f.links[0].ID || cascade.Links[1].ID != f.links[1].ID ||
				len(cascade.Native.All) != 2 {
				t.Fatalf("cascade plan: %+v %v", cascade, err)
			}
			if _, err := plan([]string{"beads/first", "beads/missing"}, true); !errors.Is(err, ErrNotFound) {
				t.Fatalf("missing named Issue: %v", err)
			}
			if !reflect.DeepEqual(before, workflowState(t, f.ctx, f.store)) {
				t.Fatal("read-only plan changed graph or native state")
			}
		})
	}
}
