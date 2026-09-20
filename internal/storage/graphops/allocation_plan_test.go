package graphops

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"
)

// This is a qualification oracle for the pinned engine's tree-format plans,
// not a production SQL parser or a guarantee about another engine version.
func checkAllocationLookupPlan(lines, tables []string, path string) error {
	for _, table := range tables {
		found := false
		key := "path"
		if table == "graph_type_descriptors" {
			key = "url"
		}
		for i, line := range lines {
			node, depth := allocationPlanNode(line)
			if node == "name: "+table {
				return fmt.Errorf("exact lookup plan scans %s", table)
			}
			if node != "IndexedTableAccess("+table+")" {
				continue
			}
			found = true
			keyed, point := false, false
			for _, childLine := range lines[i+1:] {
				child, childDepth := allocationPlanNode(childLine)
				if childDepth <= depth {
					break
				}
				keyed = keyed || child == "index: ["+table+"."+key+"]"
				value, closedPoint := allocationPlanPoint(child)
				if strings.HasPrefix(child, "filters:") || strings.HasPrefix(child, "static:") {
					if !closedPoint || table != "graph_type_descriptors" && value != path {
						return fmt.Errorf("exact lookup plan has nonpoint or mismatched %s range", table)
					}
					point = true
				}
				// String(), used by FORMAT=tree, prints a scalar key without
				// brackets. This exact key links one Link to its descriptor.
				point = point || table == "graph_type_descriptors" && child == "keys: src.type_url"
			}
			if !keyed || !point {
				return fmt.Errorf("exact lookup plan lacks qualified %s primary-key point access", table)
			}
		}
		if !found {
			return fmt.Errorf("exact lookup plan lacks indexed %s access", table)
		}
	}
	return nil
}

func allocationPlanNode(line string) (string, int) {
	node := strings.TrimLeft(line, " \t│├└─")
	return node, utf8.RuneCountInString(line) - utf8.RuneCountInString(node)
}

func allocationPlanPoint(node string) (string, bool) {
	for _, prefix := range []string{"filters: [{[", "static: [{["} {
		if !strings.HasPrefix(node, prefix) || !strings.HasSuffix(node, "]}]") {
			continue
		}
		bounds := strings.Split(strings.TrimSuffix(strings.TrimPrefix(node, prefix), "]}]"), ", ")
		if len(bounds) == 2 && bounds[0] != "" && bounds[0] != "NULL" && !strings.Contains(bounds[0], "∞") && bounds[0] == bounds[1] {
			return bounds[0], true
		}
	}
	return "", false
}

func TestAllocationLookupPlanQualification(t *testing.T) {
	tables := []string{"graph_allocations", "graph_links", "graph_type_descriptors"}
	const path = "links/plan-decision"
	good := []string{
		" ├─ IndexedTableAccess(graph_allocations)", " │   ├─ index: [graph_allocations.path]", " │   └─ filters: [{[" + path + ", " + path + "]}]",
		" ├─ IndexedTableAccess(graph_links)", " │   ├─ index: [graph_links.path]", " │   └─ filters: [{[" + path + ", " + path + "]}]",
		" └─ IndexedTableAccess(graph_type_descriptors)", "     ├─ index: [graph_type_descriptors.url]", "     └─ keys: src.type_url",
	}
	if err := checkAllocationLookupPlan(good, tables, path); err != nil {
		t.Fatal(err)
	}
	bead := strings.Split(strings.ReplaceAll(strings.ReplaceAll(strings.Join(good[:6], "\n"), "graph_links", "graph_beads"), path, "beads/plan"), "\n")
	if err := checkAllocationLookupPlan(bead, []string{"graph_allocations", "graph_beads"}, "beads/plan"); err != nil {
		t.Fatal(err)
	}
	withReplacement := func(old, replacement string) []string {
		return strings.Split(strings.ReplaceAll(strings.Join(good, "\n"), old, replacement), "\n")
	}
	for _, tc := range []struct {
		name  string
		lines []string
	}{
		{"empty", nil},
		{"descriptor-absent", good[:6]},
		{"descriptor-scanned", append(append([]string{}, good[:6]...), " └─ Table", "     └─ name: graph_type_descriptors")},
		{"additional-resource-scan", append(append([]string{}, good...), " └─ Table", "     └─ name: graph_links")},
		{"wrong-index", withReplacement("graph_links.path", "graph_links.type_url")},
		{"unrelated-text", withReplacement("IndexedTableAccess(graph_links)", "columns: ['IndexedTableAccess(graph_links)']")},
		{"full-index-range", withReplacement("keys: src.type_url", "static: [{[NULL, ∞)}]")},
		{"point-and-full-range", withReplacement("keys: src.type_url", "keys: src.type_url\n     └─ static: [{[NULL, ∞)}]")},
		{"nonpoint-index-range", withReplacement("keys: src.type_url", "filters: [{[a, z]}]")},
		{"wrong-path-point", withReplacement("filters: [{["+path+", "+path+"]}]", "filters: [{[links/other, links/other]}]")},
		{"sibling-key", withReplacement("     ├─ index: [graph_type_descriptors.url]", " ├─ index: [graph_type_descriptors.url]")},
		{"sibling-point", withReplacement("     └─ keys: src.type_url", " └─ keys: src.type_url")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := checkAllocationLookupPlan(tc.lines, tables, path); err == nil {
				t.Fatal("unqualified lookup plan accepted")
			}
		})
	}
}
