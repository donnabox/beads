package graphops

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	graph "github.com/steveyegge/beads/graphops"
)

// A deliberately narrow oracle for the pinned FORMAT=tree dialect. This proves
// planned hydration access and input placement, not total incident work or cost.
// Unknown equivalent optimizer rewrites must be inspected, not silently accepted.
type pagePlanNode struct {
	label              string
	depth, parent, end int
}
type pagePlanTree []pagePlanNode

func parsePagePlan(lines []string) (pagePlanTree, error) {
	if len(lines) == 0 || len(lines) > 256 {
		return nil, fmt.Errorf("invalid plan size")
	}
	var out pagePlanTree
	var stack []int
	for _, line := range lines {
		label := strings.TrimLeft(line, " │├└─")
		depth := utf8.RuneCountInString(line) - utf8.RuneCountInString(label)
		if label == "" || len(line) > 4096 || strings.ContainsAny(label, "\n\r\t") {
			return nil, fmt.Errorf("invalid plan line")
		}
		for len(stack) > 0 && out[stack[len(stack)-1]].depth >= depth {
			out[stack[len(stack)-1]].end = len(out)
			stack = stack[:len(stack)-1]
		}
		parent := -1
		if len(stack) > 0 {
			parent = stack[len(stack)-1]
			if depth != out[parent].depth+4 {
				return nil, fmt.Errorf("ambiguous plan depth")
			}
		} else if len(out) > 0 || depth != 0 {
			return nil, fmt.Errorf("multiple or indented roots")
		}
		out = append(out, pagePlanNode{label: label, depth: depth, parent: parent, end: len(lines)})
		stack = append(stack, len(out)-1)
	}
	for n := range out {
		for _, prefix := range []string{"name: ", "columns: ", "index: ", "keys: ", "static: ", "filters: ", "outerVisibility: ", "isLateral: ", "cacheable: ", "colSet: ", "tableId: "} {
			if strings.HasPrefix(out[n].label, prefix) && out[n].end != n+1 {
				return nil, fmt.Errorf("field has nested operators")
			}
		}
	}
	return out, nil
}
func (p pagePlanTree) below(n, ancestor int) bool { return n > ancestor && n < p[ancestor].end }
func (p pagePlanTree) children(n int) []int {
	var out []int
	for i := n + 1; i < p[n].end; i++ {
		if p[i].parent == n {
			out = append(out, i)
		}
	}
	return out
}
func (p pagePlanTree) field(n int, prefix string) (string, error) {
	value := ""
	count := 0
	for _, i := range p.children(n) {
		if strings.HasPrefix(p[i].label, prefix) {
			value = strings.TrimPrefix(p[i].label, prefix)
			count++
		}
	}
	if count != 1 {
		return "", fmt.Errorf("%s needs exactly one %s", p[n].label, prefix)
	}
	return value, nil
}
func (p pagePlanTree) one(label string) (int, error) {
	found := -1
	for i, n := range p {
		if n.label == label {
			if found != -1 {
				return 0, fmt.Errorf("duplicate %s", label)
			}
			found = i
		}
	}
	if found < 0 {
		return 0, fmt.Errorf("missing %s", label)
	}
	return found, nil
}
func (p pagePlanTree) alias(name string) (int, error) {
	found := -1
	for i, n := range p {
		if n.label != "SubqueryAlias" {
			continue
		}
		v, err := p.field(i, "name: ")
		if err != nil {
			return 0, err
		}
		if v == name {
			if found != -1 {
				return 0, fmt.Errorf("duplicate alias %s", name)
			}
			found = i
		}
	}
	if found < 0 {
		return 0, fmt.Errorf("missing alias %s", name)
	}
	return found, nil
}
func (p pagePlanTree) accessUnder(alias int, table, index string) (int, error) {
	found := -1
	for i := alias + 1; i < p[alias].end; i++ {
		if p[i].label == "IndexedTableAccess("+table+")" {
			if found != -1 {
				return 0, fmt.Errorf("duplicate %s access", table)
			}
			found = i
		}
	}
	if found < 0 {
		return 0, fmt.Errorf("missing %s access", table)
	}
	got, err := p.field(found, "index: ")
	if err != nil || got != index {
		return 0, fmt.Errorf("wrong %s index: %q", table, got)
	}
	return found, nil
}
func (p pagePlanTree) exactAccessFields(n int, field, value string) error {
	got, err := p.field(n, field)
	if err != nil || got != value {
		return fmt.Errorf("wrong %s: %q", field, got)
	}
	for _, i := range p.children(n) {
		s := p[i].label
		if strings.HasPrefix(s, "index: ") || strings.HasPrefix(s, "columns: ") || strings.HasPrefix(s, field) {
			continue
		}
		return fmt.Errorf("unexpected access field %q", s)
	}
	return nil
}

// Anchor probes use the qualified exact-read LIMIT2 -> PK shape. Unknown
// equivalent rewrites require actual plan review, not a broader oracle.
func (p pagePlanTree) incidentAnchorAccess(alias int, table, path string) (int, int, error) {
	access, err := p.accessUnder(alias, table, "["+table+".path]")
	if err != nil {
		return 0, 0, err
	}
	limit := -1
	for _, child := range p.children(alias) {
		label := p[child].label
		if label == "Limit(2)" && limit == -1 {
			limit = child
			continue
		}
		metadata := false
		for _, prefix := range []string{"name: ", "outerVisibility: ", "isLateral: ", "cacheable: ", "colSet: ", "tableId: "} {
			metadata = metadata || strings.HasPrefix(label, prefix)
		}
		if !metadata {
			return 0, 0, fmt.Errorf("unexpected anchor wrapper")
		}
	}
	if limit < 0 {
		return 0, 0, fmt.Errorf("missing anchor LIMIT2")
	}
	children := p.children(limit)
	if len(children) != 1 || children[0] != access {
		return 0, 0, fmt.Errorf("unexpected anchor wrapper")
	}
	if err := p.exactAccessFields(access, "filters: ", "[{["+path+", "+path+"]}]"); err != nil {
		return 0, 0, err
	}
	return access, limit, nil
}

func checkIncidentHydrationPlan(lines []string, path string, direction graph.Direction, window pageWindow) error {
	p, err := parsePagePlan(lines)
	if err != nil {
		return err
	}
	for _, node := range p {
		if node.label == "name: graph_links" || node.label == "name: graph_beads" || node.label == "name: graph_allocations" {
			return fmt.Errorf("graph table scan")
		}
	}
	l, err := p.alias("l")
	if err != nil {
		return err
	}
	i, err := p.alias("i")
	if err != nil {
		return err
	}
	src, err := p.one("TableAlias(src)")
	if err != nil {
		return err
	}
	b, err := p.alias("b")
	if err != nil {
		return err
	}
	a, err := p.alias("a")
	if err != nil {
		return err
	}
	req, err := p.alias("req")
	if err != nil {
		return err
	}
	if p.below(a, l) || p.below(b, l) || p.below(req, l) || p.below(a, b) || p.below(b, a) || p.below(req, a) || p.below(req, b) || p.below(a, req) || p.below(b, req) {
		return fmt.Errorf("wrong anchor roles")
	}
	join, err := p.one("LeftOuterLookupJoin")
	if err != nil {
		return err
	}
	if !p.below(i, l) || !p.below(src, l) || p.below(src, i) || p.below(b, l) || !p.below(join, l) {
		return fmt.Errorf("wrong hydration roles")
	}
	branches := p.children(join)
	if len(branches) != 2 || !(branches[0] == i || p.below(i, branches[0])) || !(branches[1] == src || p.below(src, branches[1])) {
		return fmt.Errorf("lookup operands do not own i then src")
	}
	hydration, err := p.accessUnder(src, "graph_links", "[graph_links.path]")
	if err != nil {
		return err
	}
	if err := p.exactAccessFields(hydration, "keys: ", "i.path"); err != nil {
		return err
	}
	// No wrappers may hide a hash, scan, filter or second join on the right.
	if children := p.children(src); len(children) != 1 || children[0] != hydration || branches[1] != src {
		return fmt.Errorf("unexpected hydration wrapper")
	}
	anchor, anchorLimit, err := p.incidentAnchorAccess(b, "graph_beads", path)
	if err != nil {
		return err
	}
	allocation, allocationLimit, err := p.incidentAnchorAccess(a, "graph_allocations", path)
	if err != nil {
		return err
	}
	var caps []int
	unions := []int{}
	accesses := []int{}
	for n, node := range p {
		if node.label == "name: graph_links" || node.label == "name: graph_beads" || node.label == "name: graph_allocations" {
			return fmt.Errorf("graph table scan")
		}
		if strings.HasPrefix(node.label, "Limit(") || strings.HasPrefix(node.label, "TopN(") {
			if n == anchorLimit || n == allocationLimit {
				continue
			}
			if !p.below(n, i) {
				return fmt.Errorf("candidate cap outside i")
			}
			caps = append(caps, n)
		}
		if strings.HasPrefix(node.label, "Union") {
			unions = append(unions, n)
		}
		if node.label == "IndexedTableAccess(graph_links)" {
			accesses = append(accesses, n)
		}
		if node.label == "IndexedTableAccess(graph_allocations)" && n != allocation {
			return fmt.Errorf("extra allocation access")
		}
		if node.label == "IndexedTableAccess(graph_beads)" && n != anchor {
			return fmt.Errorf("extra anchor access")
		}
	}
	capNode := -1
	if direction == graph.DirectionBoth {
		if len(caps) != 1 {
			return fmt.Errorf("Both candidate cap count")
		}
		capNode = caps[0]
		if p[capNode].parent != i || p[capNode].label != fmt.Sprintf("TopN(Limit: [%d]; candidates.path ASC)", window.limit+1) {
			return fmt.Errorf("Both candidate cap shape")
		}
	} else {
		if len(caps) != 2 {
			return fmt.Errorf("single-direction candidate cap count")
		}
		limit, top := caps[0], caps[1]
		capNode = top
		if p[limit].parent != i || p[limit].label != fmt.Sprintf("Limit(%d)", window.limit+1) || p[top].label != fmt.Sprintf("TopN(Limit: [%d]; graph_links.path ASC)", window.limit+1) {
			return fmt.Errorf("single-direction candidate cap shape")
		}
		children := p.children(limit)
		if len(children) != 1 || p[children[0]].label != "Project" || p[top].parent != children[0] {
			return fmt.Errorf("single-direction candidate cap chain")
		}
		projectChildren := p.children(children[0])
		if len(projectChildren) != 2 || p[projectChildren[0]].label != "columns: [graph_links.path]" || projectChildren[1] != top {
			return fmt.Errorf("single-direction candidate cap projection")
		}
	}
	expected := []string{"source"}
	if direction == graph.DirectionIn {
		expected = []string{"target"}
	} else if direction == graph.DirectionBoth {
		expected = []string{"source", "target"}
	} else if direction != graph.DirectionOut {
		return fmt.Errorf("invalid direction")
	}
	if len(accesses) != len(expected)+1 {
		return fmt.Errorf("unaccounted Link access")
	}
	if direction == graph.DirectionBoth {
		c, err := p.alias("candidates")
		if err != nil {
			return err
		}
		if len(unions) != 1 || p[unions[0]].label != "Union distinct" || p[c].parent != capNode || p[unions[0]].parent != c {
			return fmt.Errorf("Both global cap does not precede distinct union")
		}
	} else if len(unions) != 0 {
		return fmt.Errorf("unexpected union")
	}
	seen := map[string]bool{}
	for _, a := range accesses {
		if a == hydration {
			continue
		}
		if !p.below(a, capNode) {
			return fmt.Errorf("candidate access outside global cap")
		}
		if len(unions) == 1 && !p.below(a, unions[0]) {
			return fmt.Errorf("candidate outside union")
		}
		index, err := p.field(a, "index: ")
		if err != nil {
			return err
		}
		endpoint := ""
		for _, want := range expected {
			if index == "[graph_links."+want+"_path,graph_links.type_url,graph_links.path]" {
				endpoint = want
			}
		}
		if endpoint == "" || seen[endpoint] {
			return fmt.Errorf("wrong/repeated candidate index %q", index)
		}
		seen[endpoint] = true
		ranges := "[{[" + path + ", " + path + "], [NULL, ∞), (" + window.afterPath + ", ∞)}]"
		if err := p.exactAccessFields(a, "filters: ", ranges); err != nil {
			return err
		}
		parent := p[a].parent
		if parent < 0 || p[parent].label != "Filter" {
			return fmt.Errorf("missing candidate kind filter")
		}
		fields := p.children(parent)
		if len(fields) != 2 || p[fields[0]].label != "(graph_links."+endpoint+"_kind = 'in')" || fields[1] != a {
			return fmt.Errorf("wrong candidate kind filter")
		}
	}
	return nil
}

// Synthetic positive trees constrain the oracle; only emitted engine plans can
// qualify the proposed SQL. Baseline actual hash trees below are negative cases.
func pageHydrationPlanFixture(direction graph.Direction) []string {
	var lines []string
	add := func(depth int, s string) { lines = append(lines, strings.Repeat("    ", depth)+s) }
	add(0, "Project")
	add(1, "columns: [req.requested_path, a.path, convert(a.resource_kind, char), convert(a.state, char), (NOT(b.path IS NULL)), coalesce(b.path,''), (NOT(l.candidate_path IS NULL)), (NOT(l.path IS NULL)), l.path]")
	add(1, "Sort(l.candidate_path ASC)")
	add(2, "LeftOuterJoin")
	add(3, "LeftOuterJoin")
	add(4, "LeftOuterJoin")
	add(5, "SubqueryAlias")
	add(6, "name: req")
	add(6, "Project")
	add(7, "columns: ['beads/plan' as requested_path]")
	add(7, "Table")
	add(8, "name: ")
	addAnchor := func(depth int, alias, table string) {
		add(depth, "SubqueryAlias")
		add(depth+1, "name: "+alias)
		add(depth+1, "Limit(2)")
		add(depth+2, "IndexedTableAccess("+table+")")
		add(depth+3, "index: ["+table+".path]")
		add(depth+3, "filters: [{[beads/plan, beads/plan]}]")
	}
	addAnchor(5, "a", "graph_allocations")
	addAnchor(4, "b", "graph_beads")
	add(3, "SubqueryAlias")
	add(4, "name: l")
	add(4, "Project")
	add(5, "columns: [i.path as candidate_path, src.path]")
	add(5, "LeftOuterLookupJoin")
	add(6, "SubqueryAlias")
	add(7, "name: i")
	depth := 7
	key := "graph_links.path"
	ends := []string{"source"}
	if direction == graph.DirectionIn {
		ends = []string{"target"}
	}
	if direction == graph.DirectionBoth {
		key = "candidates.path"
		ends = []string{"source", "target"}
	}
	if direction != graph.DirectionBoth {
		add(depth, "Limit(3)")
		depth++
		add(depth, "Project")
		depth++
		add(depth, "columns: [graph_links.path]")
	}
	add(depth, "TopN(Limit: [3]; "+key+" ASC)")
	depth++
	if direction == graph.DirectionBoth {
		add(depth, "SubqueryAlias")
		depth++
		add(depth, "name: candidates")
		add(depth, "Union distinct")
		depth++
	}
	for _, endpoint := range ends {
		filterDepth := depth
		if direction == graph.DirectionBoth {
			add(depth, "Project")
			add(depth+1, "columns: [graph_links.path]")
			filterDepth++
		}
		add(filterDepth, "Filter")
		add(filterDepth+1, "(graph_links."+endpoint+"_kind = 'in')")
		add(filterDepth+1, "IndexedTableAccess(graph_links)")
		add(filterDepth+2, "index: [graph_links."+endpoint+"_path,graph_links.type_url,graph_links.path]")
		add(filterDepth+2, "filters: [{[beads/plan, beads/plan], [NULL, ∞), (links/a, ∞)}]")
	}
	add(6, "TableAlias(src)")
	add(7, "IndexedTableAccess(graph_links)")
	add(8, "index: [graph_links.path]")
	add(8, "keys: i.path")
	return lines
}

func TestIncidentHydrationPlanOracle(t *testing.T) {
	for _, direction := range []graph.Direction{graph.DirectionIn, graph.DirectionOut, graph.DirectionBoth} {
		t.Run(fmt.Sprint(direction), func(t *testing.T) {
			good := pageHydrationPlanFixture(direction)
			check := func(lines []string) error {
				return checkIncidentHydrationPlan(lines, "beads/plan", direction, pageWindow{afterPath: "links/a", limit: 2})
			}
			if err := check(good); err != nil {
				t.Fatal(err)
			}
			capShape, capCount := "single-direction candidate cap shape", "single-direction candidate cap count"
			if direction == graph.DirectionBoth {
				capShape, capCount = "Both candidate cap shape", "Both candidate cap count"
			}
			reasons := map[string]string{"hash": "missing LeftOuterLookupJoin", "plain-join": "missing LeftOuterLookupJoin", "wrong-key": "wrong keys:", "candidate-key": "wrong keys:", "literal-key": "wrong keys:", "full-range": "wrong keys:", "key-plus-range": "unexpected access field", "sibling-key": "wrong keys:", "composite-hydration": "wrong graph_links index:", "fake-access": "field has nested operators", "src-alias": "missing TableAlias(src)", "candidate-alias": "missing alias i", "hydration-alias": "missing alias l", "cap-value": capShape, "cap-missing": capCount, "candidate-range": "wrong filters:", "candidate-kind": "wrong candidate kind filter", "anchor-point": "wrong filters:", "anchor-index": "wrong graph_beads index:"}
			for _, tc := range []struct{ name, old, new string }{
				{"hash", "LeftOuterLookupJoin", "LeftOuterHashJoin"},
				{"plain-join", "LeftOuterLookupJoin", "LeftOuterJoin"},
				{"wrong-key", "keys: i.path", "keys: b.path"},
				{"candidate-key", "keys: i.path", "keys: candidates.path"},
				{"literal-key", "keys: i.path", "keys: 'i.path'"},
				{"full-range", "keys: i.path", "static: [{[NULL, ∞)}]"},
				{"key-plus-range", "keys: i.path", "keys: i.path\n                                static: [{[NULL, ∞)}]"},
				{"sibling-key", "                                keys: i.path", "                            keys: i.path"},
				{"composite-hydration", "index: [graph_links.path]", "index: [graph_links.path,graph_links.type_url]"},
				{"fake-access", "IndexedTableAccess(graph_links)", "columns: ['IndexedTableAccess(graph_links)']"},
				{"src-alias", "TableAlias(src)", "TableAlias(other)"},
				{"candidate-alias", "name: i", "name: src"},
				{"hydration-alias", "name: l", "name: i"},
				{"cap-value", "Limit: [3]", "Limit: [2]"},
				{"cap-missing", "TopN(Limit: [3];", "Sort("},
				{"candidate-range", "(links/a, ∞)", "(links/b, ∞)"},
				{"candidate-kind", "_kind = 'in'", "_kind = 'ext'"},
				{"anchor-point", "[beads/plan, beads/plan]}]", "[beads/other, beads/other]}]"},
				{"anchor-index", "[graph_beads.path]", "[graph_beads.type_url]"},
			} {
				t.Run(tc.name, func(t *testing.T) {
					raw := strings.Join(good, "\n")
					changed := strings.ReplaceAll(raw, tc.old, tc.new)
					if changed == raw {
						t.Fatal("mutation did not apply")
					}
					lines := strings.Split(changed, "\n")
					if tc.name != "fake-access" {
						if _, err := parsePagePlan(lines); err != nil {
							t.Fatalf("mutation must parse: %v", err)
						}
					}
					requirePagePlanReason(t, check(lines), reasons[tc.name])
				})
			}
			requirePagePlanReason(t, check(append(append([]string{}, good...), "    Table", "        name: graph_links")), "graph table scan")
			requirePagePlanReason(t, checkIncidentHydrationPlan(good, "beads/plan", direction, pageWindow{afterPath: "links/a", limit: 3}), capShape)
		})
	}
	good := strings.Join(pageHydrationPlanFixture(graph.DirectionBoth), "\n")
	for _, replacement := range []string{"Union all", "LeftOuterLookupJoin"} {
		lines := strings.Split(strings.Replace(good, "Union distinct", replacement, 1), "\n")
		if _, err := parsePagePlan(lines); err != nil {
			t.Fatal(err)
		}
		want := "Both global cap does not precede distinct union"
		if replacement == "LeftOuterLookupJoin" {
			want = "duplicate LeftOuterLookupJoin"
		}
		requirePagePlanReason(t, checkIncidentHydrationPlan(lines, "beads/plan", graph.DirectionBoth, pageWindow{afterPath: "links/a", limit: 2}), want)
	}
	for _, bad := range [][]string{nil, {"Project", " Project"}, {"Project", "Project"}} {
		if _, err := parsePagePlan(bad); err == nil {
			t.Fatal("invalid tree accepted")
		}
	}
}

// Exact first-page baseline FORMAT=tree output (embedded-r5 receipt), retained
// as negatives. Candidate indexes and textual i.path did not avoid hydration scans.
func pageHydrationHashBaseline(direction graph.Direction) []string {
	switch direction {
	case graph.DirectionOut:
		return strings.Split(`Project
 ├─ columns: [b.path, (NOT(i.path IS NULL)), (NOT(l.path IS NULL)), coalesce(l.path,''), coalesce(l.type_url,''), coalesce(l.revision,''), l.attribution_principal, convert(l.attribution_status, char), CASE  WHEN ((length(l.properties) < 0) OR (length(l.properties) > 8192)) THEN NULL ELSE l.properties END, length(l.properties), coalesce(convert(l.source_kind, char),''), l.source_path, l.source_url, l.source_pin, coalesce(convert(l.target_kind, char),''), l.target_path, l.target_url, l.target_pin]
 └─ Sort(i.path ASC)
     └─ LeftOuterHashJoin
         ├─ (l.path = i.path)
         ├─ LeftOuterJoin
         │   ├─ TableAlias(b)
         │   │   └─ IndexedTableAccess(graph_beads)
         │   │       ├─ index: [graph_beads.path]
         │   │       ├─ filters: [{[beads/page-anchor, beads/page-anchor]}]
         │   │       └─ columns: [path]
         │   └─ CachedResults
         │       └─ SubqueryAlias
         │           ├─ name: i
         │           ├─ outerVisibility: false
         │           ├─ isLateral: false
         │           ├─ cacheable: true
         │           └─ Limit(3)
         │               └─ Project
         │                   ├─ columns: [graph_links.path]
         │                   └─ TopN(Limit: [3]; graph_links.path ASC)
         │                       └─ Filter
         │                           ├─ (graph_links.source_kind = 'in')
         │                           └─ IndexedTableAccess(graph_links)
         │                               ├─ index: [graph_links.source_path,graph_links.type_url,graph_links.path]
         │                               ├─ filters: [{[beads/page-anchor, beads/page-anchor], [NULL, ∞), (, ∞)}]
         │                               └─ columns: [path source_kind source_path]
         └─ HashLookup
             ├─ left-key: (i.path)
             ├─ right-key: (l.path)
             └─ TableAlias(l)
                 └─ Table
                     ├─ name: graph_links
                     └─ columns: [path type_url revision attribution_principal attribution_status properties source_kind source_path source_url source_pin target_kind target_path target_url target_pin]`, "\n")
	case graph.DirectionIn:
		return strings.Split(`Project
 ├─ columns: [b.path, (NOT(i.path IS NULL)), (NOT(l.path IS NULL)), coalesce(l.path,''), coalesce(l.type_url,''), coalesce(l.revision,''), l.attribution_principal, convert(l.attribution_status, char), CASE  WHEN ((length(l.properties) < 0) OR (length(l.properties) > 8192)) THEN NULL ELSE l.properties END, length(l.properties), coalesce(convert(l.source_kind, char),''), l.source_path, l.source_url, l.source_pin, coalesce(convert(l.target_kind, char),''), l.target_path, l.target_url, l.target_pin]
 └─ Sort(i.path ASC)
     └─ LeftOuterHashJoin
         ├─ (l.path = i.path)
         ├─ LeftOuterJoin
         │   ├─ TableAlias(b)
         │   │   └─ IndexedTableAccess(graph_beads)
         │   │       ├─ index: [graph_beads.path]
         │   │       ├─ filters: [{[beads/page-anchor, beads/page-anchor]}]
         │   │       └─ columns: [path]
         │   └─ CachedResults
         │       └─ SubqueryAlias
         │           ├─ name: i
         │           ├─ outerVisibility: false
         │           ├─ isLateral: false
         │           ├─ cacheable: true
         │           └─ Limit(3)
         │               └─ Project
         │                   ├─ columns: [graph_links.path]
         │                   └─ TopN(Limit: [3]; graph_links.path ASC)
         │                       └─ Filter
         │                           ├─ (graph_links.target_kind = 'in')
         │                           └─ IndexedTableAccess(graph_links)
         │                               ├─ index: [graph_links.target_path,graph_links.type_url,graph_links.path]
         │                               ├─ filters: [{[beads/page-anchor, beads/page-anchor], [NULL, ∞), (, ∞)}]
         │                               └─ columns: [path target_kind target_path]
         └─ HashLookup
             ├─ left-key: (i.path)
             ├─ right-key: (l.path)
             └─ TableAlias(l)
                 └─ Table
                     ├─ name: graph_links
                     └─ columns: [path type_url revision attribution_principal attribution_status properties source_kind source_path source_url source_pin target_kind target_path target_url target_pin]`, "\n")
	case graph.DirectionBoth:
		return strings.Split(`Project
 ├─ columns: [b.path, (NOT(i.path IS NULL)), (NOT(l.path IS NULL)), coalesce(l.path,''), coalesce(l.type_url,''), coalesce(l.revision,''), l.attribution_principal, convert(l.attribution_status, char), CASE  WHEN ((length(l.properties) < 0) OR (length(l.properties) > 8192)) THEN NULL ELSE l.properties END, length(l.properties), coalesce(convert(l.source_kind, char),''), l.source_path, l.source_url, l.source_pin, coalesce(convert(l.target_kind, char),''), l.target_path, l.target_url, l.target_pin]
 └─ Sort(i.path ASC)
     └─ LeftOuterHashJoin
         ├─ (l.path = i.path)
         ├─ LeftOuterJoin
         │   ├─ TableAlias(b)
         │   │   └─ IndexedTableAccess(graph_beads)
         │   │       ├─ index: [graph_beads.path]
         │   │       ├─ filters: [{[beads/page-anchor, beads/page-anchor]}]
         │   │       └─ columns: [path]
         │   └─ CachedResults
         │       └─ SubqueryAlias
         │           ├─ name: i
         │           ├─ outerVisibility: false
         │           ├─ isLateral: false
         │           ├─ cacheable: true
         │           └─ TopN(Limit: [3]; candidates.path ASC)
         │               └─ SubqueryAlias
         │                   ├─ name: candidates
         │                   ├─ outerVisibility: false
         │                   ├─ isLateral: false
         │                   ├─ cacheable: true
         │                   └─ Union distinct
         │                       ├─ Project
         │                       │   ├─ columns: [graph_links.path]
         │                       │   └─ Filter
         │                       │       ├─ (graph_links.source_kind = 'in')
         │                       │       └─ IndexedTableAccess(graph_links)
         │                       │           ├─ index: [graph_links.source_path,graph_links.type_url,graph_links.path]
         │                       │           ├─ filters: [{[beads/page-anchor, beads/page-anchor], [NULL, ∞), (, ∞)}]
         │                       │           └─ columns: [path source_kind source_path]
         │                       └─ Project
         │                           ├─ columns: [graph_links.path]
         │                           └─ Filter
         │                               ├─ (graph_links.target_kind = 'in')
         │                               └─ IndexedTableAccess(graph_links)
         │                                   ├─ index: [graph_links.target_path,graph_links.type_url,graph_links.path]
         │                                   ├─ filters: [{[beads/page-anchor, beads/page-anchor], [NULL, ∞), (, ∞)}]
         │                                   └─ columns: [path target_kind target_path]
         └─ HashLookup
             ├─ left-key: (i.path)
             ├─ right-key: (l.path)
             └─ TableAlias(l)
                 └─ Table
                     ├─ name: graph_links
                     └─ columns: [path type_url revision attribution_principal attribution_status properties source_kind source_path source_url source_pin target_kind target_path target_url target_pin]`, "\n")
	default:
		return nil
	}
}
func TestIncidentHydrationRejectsActualHashBaselines(t *testing.T) {
	for _, direction := range []graph.Direction{graph.DirectionIn, graph.DirectionOut, graph.DirectionBoth} {
		lines := pageHydrationHashBaseline(direction)
		if _, err := parsePagePlan(lines); err != nil {
			t.Fatal(err)
		}
		requirePagePlanReason(t, checkIncidentHydrationPlan(lines, "beads/page-anchor", direction, pageWindow{limit: 2}), "graph table scan")
	}
}

// Substitute the original receipt's complete i subtree into the proposed role
// scaffold, changing only the fixture operand spelling and indentation. This
// exercises the actual cap dialect, including harmless engine metadata fields.
func TestIncidentHydrationMeasuredCandidateSubtrees(t *testing.T) {
	for _, direction := range []graph.Direction{graph.DirectionIn, graph.DirectionOut, graph.DirectionBoth} {
		baseline, err := parsePagePlan(pageHydrationHashBaseline(direction))
		if err != nil {
			t.Fatal(err)
		}
		oldI, err := baseline.alias("i")
		if err != nil {
			t.Fatal(err)
		}
		good := pageHydrationPlanFixture(direction)
		tree, err := parsePagePlan(good)
		if err != nil {
			t.Fatal(err)
		}
		newI, err := tree.alias("i")
		if err != nil {
			t.Fatal(err)
		}
		replacement := append([]string{}, good[:newI]...)
		for _, node := range baseline[oldI:baseline[oldI].end] {
			label := strings.ReplaceAll(node.label, "beads/page-anchor", "beads/plan")
			label = strings.ReplaceAll(label, "(, ∞)", "(links/a, ∞)")
			replacement = append(replacement, strings.Repeat(" ", node.depth-baseline[oldI].depth+tree[newI].depth)+label)
		}
		replacement = append(replacement, good[tree[newI].end:]...)
		if err := checkIncidentHydrationPlan(replacement, "beads/plan", direction, pageWindow{afterPath: "links/a", limit: 2}); err != nil {
			t.Fatalf("measured candidate dialect %v: %v", direction, err)
		}
	}
}

func TestIncidentHydrationPlanAncestryRefusals(t *testing.T) {
	good := pageHydrationPlanFixture(graph.DirectionBoth)
	parsed, err := parsePagePlan(good)
	if err != nil {
		t.Fatal(err)
	}
	find := func(label string) int {
		n, err := parsed.one(label)
		if err != nil {
			t.Fatal(err)
		}
		return n
	}
	capAt := find("TopN(Limit: [3]; candidates.path ASC)")
	src := find("TableAlias(src)")
	filter := -1
	for n, node := range parsed {
		if node.label == "Filter" {
			filter = n
			break
		}
	}
	for _, kind := range []string{"cap-above-hydration", "cap-in-one-branch", "cap-duplicate", "src-beneath-i", "extra-indexed-scan", "duplicate-src", "sibling-anchor-point", "hash-wrapper"} {
		t.Run(kind, func(t *testing.T) {
			nodes := append(pagePlanTree{}, parsed...)
			if kind == "cap-above-hydration" || kind == "cap-in-one-branch" {
				nodes[capAt].label = "Project"
			}
			var lines []string
			for n, node := range nodes {
				if kind == "cap-above-hydration" && n == 1 {
					lines = append(lines, "    TopN(Limit: [3]; candidates.path ASC)")
				}
				if kind == "cap-in-one-branch" && n == filter {
					lines = append(lines, strings.Repeat(" ", node.depth)+"TopN(Limit: [3]; candidates.path ASC)")
				}
				if kind == "cap-in-one-branch" && n >= filter && n < parsed[filter].end {
					node.depth += 4
				}
				if kind == "src-beneath-i" && n >= src && n < parsed[src].end {
					node.depth += 4
				}
				if kind == "sibling-anchor-point" && node.label == "filters: [{[beads/plan, beads/plan]}]" {
					node.depth -= 4
				}
				if kind == "hash-wrapper" && n == src+1 {
					lines = append(lines, strings.Repeat(" ", node.depth)+"HashLookup")
				}
				if kind == "hash-wrapper" && n > src && n < parsed[src].end {
					node.depth += 4
				}
				lines = append(lines, strings.Repeat(" ", node.depth)+node.label)
			}
			switch kind {
			case "cap-duplicate":
				lines = append(lines, "    Limit(3)")
			case "extra-indexed-scan":
				lines = append(lines, "    IndexedTableAccess(graph_links)", "        index: [graph_links.path]", "        keys: i.path")
			case "duplicate-src":
				lines = append(lines, "    TableAlias(src)")
			}
			if _, err := parsePagePlan(lines); err != nil {
				t.Fatalf("ancestry mutant must parse: %v", err)
			}
			reasons := map[string]string{"cap-above-hydration": "candidate cap outside i", "cap-in-one-branch": "Both candidate cap shape", "cap-duplicate": "candidate cap outside i", "src-beneath-i": "wrong hydration roles", "extra-indexed-scan": "unaccounted Link access", "duplicate-src": "duplicate TableAlias(src)", "sibling-anchor-point": "unexpected anchor wrapper", "hash-wrapper": "unexpected hydration wrapper"}
			requirePagePlanReason(t, checkIncidentHydrationPlan(lines, "beads/plan", graph.DirectionBoth, pageWindow{afterPath: "links/a", limit: 2}), reasons[kind])
		})
	}
	requirePagePlanReason(t, checkIncidentHydrationPlan(pageHydrationPlanFixture(graph.DirectionIn), "beads/plan", graph.DirectionOut, pageWindow{afterPath: "links/a", limit: 2}), "wrong/repeated candidate index")
}

// Stable rejection categories make the negative controls distinguish the gate
// they exercise; malformed syntax is not counted as an access/ancestry proof.
func requirePagePlanReason(t *testing.T, err error, prefix string) {
	t.Helper()
	if err == nil || prefix == "" || !strings.HasPrefix(err.Error(), prefix) {
		t.Fatalf("rejection=%v want reason %q", err, prefix)
	}
}

func TestIncidentHydrationMeasuredCapDialect(t *testing.T) {
	for _, direction := range []graph.Direction{graph.DirectionIn, graph.DirectionOut} {
		raw := strings.Join(pageHydrationPlanFixture(direction), "\n")
		for _, tc := range []struct{ old, new, want string }{{"Limit(3)", "Limit(2)", "single-direction candidate cap shape"}, {"TopN(Limit: [3]; graph_links.path ASC)", "Sort(graph_links.path ASC)", "single-direction candidate cap count"}, {"Limit(3)", "Project", "single-direction candidate cap count"}} {
			altered := strings.Replace(raw, tc.old, tc.new, 1)
			if altered == raw {
				t.Fatal("cap mutation did not apply")
			}
			lines := strings.Split(altered, "\n")
			if _, err := parsePagePlan(lines); err != nil {
				t.Fatal(err)
			}
			requirePagePlanReason(t, checkIncidentHydrationPlan(lines, "beads/plan", direction, pageWindow{afterPath: "links/a", limit: 2}), tc.want)
		}
	}
	// A second cap INSIDE i differs from the separately tested outside-i cap.
	raw := strings.Join(pageHydrationPlanFixture(graph.DirectionBoth), "\n")
	altered := strings.Replace(raw, "                            name: i", "                            name: i\n                            Limit(3)", 1)
	if altered == raw {
		t.Fatal("inside-i mutation did not apply")
	}
	lines := strings.Split(altered, "\n")
	if _, err := parsePagePlan(lines); err != nil {
		t.Fatal(err)
	}
	requirePagePlanReason(t, checkIncidentHydrationPlan(lines, "beads/plan", graph.DirectionBoth, pageWindow{afterPath: "links/a", limit: 2}), "Both candidate cap count")
	// New aliases remain valid, but actual baseline-style HashLookup/Table is
	// substituted for src PK access. Rejection must be the scan gate.
	altered = strings.Replace(raw, "LeftOuterLookupJoin", "LeftOuterHashJoin", 1)
	altered = strings.Replace(altered, "                            IndexedTableAccess(graph_links)\n                                index: [graph_links.path]\n                                keys: i.path", "                            HashLookup\n                                left-key: (i.path)\n                                right-key: (src.path)\n                                Table\n                                    name: graph_links", 1)
	if altered == strings.Replace(raw, "LeftOuterLookupJoin", "LeftOuterHashJoin", 1) {
		t.Fatal("hash hydration mutation did not apply")
	}
	lines = strings.Split(altered, "\n")
	tree, err := parsePagePlan(lines)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tree.alias("l"); err != nil {
		t.Fatal(err)
	}
	if _, err := tree.alias("i"); err != nil {
		t.Fatal(err)
	}
	if _, err := tree.one("TableAlias(src)"); err != nil {
		t.Fatal(err)
	}
	requirePagePlanReason(t, checkIncidentHydrationPlan(lines, "beads/plan", graph.DirectionBoth, pageWindow{afterPath: "links/a", limit: 2}), "graph table scan")
}

// Mutations isolate the newly introduced allocation/request probes while leaving
// the already qualified candidate and dynamic hydration subtrees unchanged.
func TestIncidentAllocationAnchorPlanOracle(t *testing.T) {
	for _, direction := range []graph.Direction{graph.DirectionIn, graph.DirectionOut, graph.DirectionBoth} {
		good := strings.Join(pageHydrationPlanFixture(direction), "\n")
		check := func(raw string) error {
			return checkIncidentHydrationPlan(strings.Split(raw, "\n"), "beads/plan", direction, pageWindow{afterPath: "links/a", limit: 2})
		}
		if err := check(good); err != nil {
			t.Fatal(err)
		}
		for _, tc := range []struct{ name, old, replacement, reason string }{
			{"allocation-index", "index: [graph_allocations.path]", "index: [graph_allocations.state]", "wrong graph_allocations index:"},
			{"allocation-point", "index: [graph_allocations.path]\n                                filters: [{[beads/plan, beads/plan]}]", "index: [graph_allocations.path]\n                                filters: [{[beads/other, beads/other]}]", "wrong filters:"},
			{"allocation-limit", "                        Limit(2)", "                        Limit(1)", "unexpected anchor wrapper"},
			{"bead-limit", "\n                    Limit(2)", "\n                    Limit(1)", "unexpected anchor wrapper"},
			{"allocation-project", "                            IndexedTableAccess(graph_allocations)\n                                index: [graph_allocations.path]\n                                filters: [{[beads/plan, beads/plan]}]", "                            Project\n                                columns: [graph_allocations.path]\n                                IndexedTableAccess(graph_allocations)\n                                    index: [graph_allocations.path]\n                                    filters: [{[beads/plan, beads/plan]}]", "unexpected anchor wrapper"},
			{"allocation-alias", "name: a", "name: other", "missing alias a"},
			{"request-alias", "name: req", "name: other", "missing alias req"},
		} {
			t.Run(fmt.Sprintf("%d/%s", direction, tc.name), func(t *testing.T) {
				bad := strings.Replace(good, tc.old, tc.replacement, 1)
				if bad == good {
					t.Fatal("mutation did not apply")
				}
				if _, err := parsePagePlan(strings.Split(bad, "\n")); err != nil {
					t.Fatal(err)
				}
				requirePagePlanReason(t, check(bad), tc.reason)
			})
		}
		requirePagePlanReason(t, check(good+"\n    IndexedTableAccess(graph_allocations)\n        index: [graph_allocations.path]\n        filters: [{[beads/plan, beads/plan]}]"), "extra allocation access")
	}
}
