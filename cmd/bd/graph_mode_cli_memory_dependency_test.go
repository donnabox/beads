//go:build cgo

package main

import (
	"reflect"
	"testing"
)

// Memory reads are additive in link mode. Deliberately pin the legacy query
// behavior, including its repeated-title filter overwrite, in dependency mode.
// Correcting that quirk requires a separate compatibility decision/change.
func TestGraphModeCLIMemoryDependencyQueryCompatibility(t *testing.T) {
	bd := buildEmbeddedBD(t)
	dir, _, _ := bdInit(t, bd, "--prefix", "cmp", "--graph-mode", "dependency")
	first := bdCreate(t, bd, dir, "alpha beta", "--type", "task")
	second := bdCreate(t, bd, dir, "beta", "--type", "task")
	for _, expression := range []string{"status=open", "title=alpha", "title=alpha AND title=beta"} {
		t.Run(expression, func(t *testing.T) {
			implicit := bdQueryJSON(t, bd, dir, expression)
			explicit := bdQueryJSON(t, bd, dir, expression, "--graph-mode=dependency")
			if !reflect.DeepEqual(implicit, explicit) {
				t.Fatalf("explicit mode changed result: %v vs %v", implicit, explicit)
			}
			if expression == "title=alpha AND title=beta" {
				ids := map[string]bool{}
				for _, row := range explicit {
					if id, ok := row["id"].(string); ok {
						ids[id] = true
					}
				}
				if !ids[first.ID] || !ids[second.ID] || len(ids) != 2 {
					t.Fatalf("legacy repeated-title behavior changed: %v", explicit)
				}
			}
		})
	}
}
