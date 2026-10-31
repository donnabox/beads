package main

import (
	"encoding/json"
	"fmt"
	"sort"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/resourceget"
)

const graphGenericBound = 1000
const graphGenericOutputLimit = 1 << 20

type graphGenericInput struct {
	Root                      string
	All                       bool
	LinkTypes                 []string // selected installed Type URLs, OR union including refinements
	ExactLinkTypes            []string // selected installed Type URLs, without refinements
	AllLinkTypes              bool
	ImplicitDependency        bool // disclose omitted incidence only for the implicit default
	OpenIssuesOnly            bool // preserve the ordinary Dependency --all population
	Direction                 string
	Depth, MaxNodes, MaxLinks int
}

// These allowlists are a private CLI summary, not new BDP record schemas.
type graphGenericNode struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	Version     string                 `json:"version"`
	Attribution graphstore.Attribution `json:"attribution"`
	Status      string                 `json:"status,omitempty"`
}
type graphGenericLink struct {
	ID             string                 `json:"id"`
	Type           string                 `json:"type"`
	Source         string                 `json:"source"`
	Target         string                 `json:"target"`
	Version        string                 `json:"version"`
	Attribution    graphstore.Attribution `json:"attribution"`
	SourceRevision string                 `json:"sourceRevision,omitempty"`
	TargetRevision string                 `json:"targetRevision,omitempty"`
}
type graphGenericResult struct {
	Projection       string             `json:"projection"`
	Scope            string             `json:"scope"`
	Root             string             `json:"root"`
	Direction        string             `json:"direction"`
	Depth            int                `json:"depth"`
	MaxNodes         int                `json:"maxNodes"`
	MaxLinks         int                `json:"maxLinks"`
	Nodes            []graphGenericNode `json:"nodes"`
	Links            []graphGenericLink `json:"links"`
	Frontier         []string           `json:"frontier"`
	Complete         bool               `json:"complete"`
	All              bool               `json:"all,omitempty"`
	LinkTypes        []string           `json:"linkTypes"`
	ExactLinkTypes   []string           `json:"exactLinkTypes,omitempty"`
	AllLinkTypes     bool               `json:"allLinkTypes,omitempty"`
	OmittedLinkTypes []string           `json:"-"`
}

func validateGraphGenericBounds(in graphGenericInput) error {
	if in.Direction != "in" && in.Direction != "out" && in.Direction != "both" {
		return fmt.Errorf("direction must be in, out or both")
	}
	if in.Depth < 0 || in.Depth > graphGenericBound || in.MaxNodes < 1 || in.MaxNodes > graphGenericBound || in.MaxLinks < 1 || in.MaxLinks > graphGenericBound {
		return fmt.Errorf("depth must be 0..%d and node/Link bounds 1..%d", graphGenericBound, graphGenericBound)
	}
	return nil
}

// projectGraphGeneric uses exactly the supplied checked snapshot. It never
// follows Owned copies, fetches an endpoint, writes, or infers history ordering.
// #5283 motivates retaining diamond Links independently of node visitation;
// #6148 motivates direction/permutation tests. Their native algorithms remain
// contributor-owned and are neither imported nor replaced here.
func projectGraphGeneric(snapshot graphstore.Snapshot, scope string, in graphGenericInput) (graphGenericResult, error) {
	zero := graphGenericResult{}
	invalid := func(message string) (graphGenericResult, error) {
		return zero, fmt.Errorf("%w: %s", graphstore.ErrInvalidStore, message)
	}
	if err := validateGraphGenericBounds(in); err != nil {
		return zero, fmt.Errorf("%w: %v", graph.ErrValidation, err)
	}
	if err := graph.ValidateScopeURL(scope); err != nil {
		return invalid("invalid snapshot Scope")
	}
	if len(snapshot.Records) > graphstore.PreviewSnapshotLimit {
		return zero, fmt.Errorf("%w: current inventory exceeds %d Resources", graphstore.ErrLimitExceeded, graphstore.PreviewSnapshotLimit)
	}
	issueType := graphstore.IssueTypeURL(scope)
	dependencyTypes := map[string]bool{graphstore.DependencyTypeURL(scope): true}
	if len(snapshot.Types) != 0 {
		if len(snapshot.Types) < len(graphstore.FreshTypeNames()) || len(snapshot.Types) > len(graphstore.FreshTypeNames())+graphstore.PreviewSnapshotLimit {
			return invalid("unsupported Type installation")
		}
		installed := map[string]bool{}
		for _, descriptor := range snapshot.Types {
			if installed[descriptor.ID()] {
				return invalid("duplicate Type installation")
			}
			known := false
			for _, name := range graphstore.FreshTypeNames() {
				known = known || descriptor.ID() == scope+name
			}
			if !known && !graphstore.IsNativeConcreteTypeDescriptor(scope, descriptor) {
				return invalid("unsupported extra Type installation")
			}
			installed[descriptor.ID()] = true
		}
		for _, name := range graphstore.FreshTypeNames() {
			if !installed[scope+name] {
				return invalid("incomplete Type installation")
			}
		}
		for id := range dependencyTypes {
			if !installed[id] {
				return invalid("incomplete native Dependency Type installation")
			}
		}
	}
	for _, selected := range append(append([]string{}, in.LinkTypes...), in.ExactLinkTypes...) {
		if err := graphPreviewValidateSelectedType(scope, snapshot.Types, selected, graph.KindLink); err != nil {
			return zero, err
		}
	}
	matches := func(actual string) bool {
		if in.AllLinkTypes || (len(in.LinkTypes) == 0 && len(in.ExactLinkTypes) == 0) { // legacy private projection callers select all explicitly
			return true
		}
		for _, selected := range in.ExactLinkTypes {
			if actual == selected {
				return true
			}
		}
		for _, selected := range in.LinkTypes {
			if graphPreviewTypeRefines(snapshot.Types, actual, selected) {
				return true
			}
		}
		return false
	}
	nodes := map[string]graphGenericNode{}
	links := map[string]graphGenericLink{}
	openIssues := map[string]bool{}
	issues := map[string]bool{}
	for _, value := range snapshot.Records {
		var node graphGenericNode
		switch record := value.(type) {
		case graphstore.Record:
			if len(snapshot.Types) == 0 {
				if record.Type != graphstore.MemoryTypeURL(scope) {
					return invalid("a refined Memory Type requires its installed descriptor inventory")
				}
			} else {
				if err := graphPreviewValidateSelectedType(scope, snapshot.Types, record.Type, graph.KindBead); err != nil || !graphPreviewTypeRefines(snapshot.Types, record.Type, graphstore.MemoryTypeURL(scope)) {
					return invalid("unsupported Memory Type")
				}
			}
			node = graphGenericNode{ID: record.ID, Type: record.Type, Title: record.Properties.Title, Version: record.Version, Attribution: record.Attribution}
		case graphstore.IssueRecord:
			issues[record.ID] = true
			if record.Properties == nil {
				return invalid("incomplete Issue")
			}
			if len(snapshot.Types) == 0 {
				// Private historical pure fixtures retain the pre-concrete base.
				if record.Type != issueType {
					return invalid("unsupported legacy Issue fixture")
				}
			} else if expected, err := graphstore.NativeIssueTypeURL(scope, string(record.Properties.IssueType)); err != nil || record.Type != expected {
				return invalid("Issue classification differs from declared concrete Type")
			}
			node = graphGenericNode{ID: record.ID, Type: record.Type, Title: record.Properties.Title, Version: record.Version, Attribution: record.Attribution, Status: string(record.Properties.Status)}
			switch record.Properties.Status {
			case "open", "in_progress", "blocked":
				openIssues[record.ID] = true
			}
		case graphstore.LinkRecord:
			if len(snapshot.Types) != 0 {
				if err := graphPreviewValidateSelectedType(scope, snapshot.Types, record.Type, graph.KindLink); err != nil {
					return invalid("unsupported Link Type")
				}
			} else if !graphstore.IsInformationalTypeURL(scope, record.Type) && !dependencyTypes[record.Type] {
				return invalid("unsupported Link Type")
			}
			if _, ok := graphPreviewPublicResourceID(scope, record.ID); !ok || record.Version == "" {
				return invalid("invalid Link identity/version")
			}
			if _, duplicate := links[record.ID]; duplicate {
				return invalid("duplicate Link identity")
			}
			links[record.ID] = graphGenericLink{ID: record.ID, Type: record.Type, Source: record.Source, Target: record.Target, Version: record.Version, Attribution: record.Attribution, SourceRevision: record.SourceRevision, TargetRevision: record.TargetRevision}
			continue
		default:
			return invalid("unsupported snapshot record kind")
		}
		if _, ok := graphPreviewPublicResourceID(scope, node.ID); !ok || node.Version == "" {
			return invalid("invalid Bead identity/version")
		}
		if _, duplicate := nodes[node.ID]; duplicate {
			return invalid("duplicate Bead identity")
		}
		nodes[node.ID] = node
	}
	population := graphGenericPopulation{Nodes: nodes, Links: links, Issues: issues, OpenIssues: openIssues,
		MatchingLinks: map[string]bool{}, DependencyLinks: map[string]bool{}}
	for id, link := range links {
		population.MatchingLinks[id] = matches(link.Type)
		population.DependencyLinks[id] = graphPreviewTypeRefines(snapshot.Types, link.Type, graphstore.DependencyTypeURL(scope)) || link.Type == graphstore.DependencyTypeURL(scope)
	}
	return selectGraphGenericPopulation(scope, in, population)
}

// The legacy adapter and complete Resource adapter share traversal, induced
// edge selection and bounds. Acquisition supplies checked per-Link membership;
// the traversal does not reinterpret a descriptor or fetch mutable state.
type graphGenericPopulation struct {
	Nodes           map[string]graphGenericNode
	Links           map[string]graphGenericLink
	Issues          map[string]bool
	OpenIssues      map[string]bool
	MatchingLinks   map[string]bool
	DependencyLinks map[string]bool
	IncludeIsolated bool
}

func selectGraphGenericPopulation(scope string, in graphGenericInput, population graphGenericPopulation) (graphGenericResult, error) {
	zero := graphGenericResult{}
	invalid := func(message string) (graphGenericResult, error) {
		return zero, fmt.Errorf("%w: %s", graphstore.ErrInvalidStore, message)
	}
	nodes, links, openIssues := population.Nodes, population.Links, population.OpenIssues
	adjacency := map[string][]graphGenericLink{}
	excludedIncidence := map[string][]string{}
	eligible := func(link graphGenericLink) bool {
		return population.MatchingLinks[link.ID] && (!in.OpenIssuesOnly || (openIssues[link.Source] && openIssues[link.Target]))
	}
	for _, link := range links {
		if _, ok := nodes[link.Source]; !ok {
			return invalid("unsupported or missing Link source")
		}
		if _, ok := nodes[link.Target]; !ok {
			return invalid("unsupported or missing Link target")
		}
		if in.ImplicitDependency && !population.DependencyLinks[link.ID] {
			// This evidence comes only from the already-authorized snapshot.
			// Consult it only for reached nodes, never unrelated components.
			excludedIncidence[link.Source] = append(excludedIncidence[link.Source], link.Type)
			if link.Target != link.Source {
				excludedIncidence[link.Target] = append(excludedIncidence[link.Target], link.Type)
			}
		}
		if !eligible(link) {
			continue
		}
		if in.Direction != "in" {
			adjacency[link.Source] = append(adjacency[link.Source], link)
		}
		if in.Direction != "out" && (in.Direction == "in" || link.Target != link.Source) {
			adjacency[link.Target] = append(adjacency[link.Target], link)
		}
	}
	for id := range adjacency {
		sort.Slice(adjacency[id], func(i, j int) bool { return graph.CompareCodeUnits(adjacency[id][i].ID, adjacency[id][j].ID) < 0 })
	}
	if !in.All {
		if _, ok := nodes[in.Root]; !ok {
			return zero, graphstore.ErrNotFound
		}
	}
	distance := map[string]int{}
	queue := []string{}
	if in.All {
		if population.IncludeIsolated {
			for id := range nodes {
				distance[id] = 0
			}
		}
		for _, link := range links {
			if eligible(link) {
				distance[link.Source] = 0
				distance[link.Target] = 0
			}
		}
		if !in.AllLinkTypes && len(in.ExactLinkTypes) == 0 && len(in.LinkTypes) == 1 && in.LinkTypes[0] == graphstore.DependencyTypeURL(scope) {
			for id := range population.Issues {
				if !in.OpenIssuesOnly || openIssues[id] {
					distance[id] = 0
				}
			}
		}
		for id := range distance {
			queue = append(queue, id)
		}
		sort.Strings(queue)
	} else {
		distance[in.Root] = 0
		queue = append(queue, in.Root)
	}
	if len(distance) > in.MaxNodes {
		return zero, fmt.Errorf("%w: traversal exceeds %d Beads", graphstore.ErrLimitExceeded, in.MaxNodes)
	}
	for cursor := 0; cursor < len(queue); cursor++ {
		id := queue[cursor]
		if distance[id] >= in.Depth {
			continue
		}
		for _, link := range adjacency[id] {
			next := link.Target
			if in.Direction == "in" || (in.Direction == "both" && id != link.Source) {
				next = link.Source
			}
			if _, seen := distance[next]; !seen {
				if len(distance) >= in.MaxNodes {
					return zero, fmt.Errorf("%w: traversal exceeds %d Beads", graphstore.ErrLimitExceeded, in.MaxNodes)
				}
				distance[next] = distance[id] + 1
				queue = append(queue, next)
			}
		}
	}
	// Depth chooses nodes. Emit the induced graph on those nodes, rather than
	// only the discovery edges: boundary cross-links, cycles and parallels count.
	emitted := map[string]graphGenericLink{}
	for _, link := range links {
		_, source := distance[link.Source]
		_, target := distance[link.Target]
		if eligible(link) && source && target {
			if len(emitted) >= in.MaxLinks {
				return zero, fmt.Errorf("%w: traversal exceeds %d Links", graphstore.ErrLimitExceeded, in.MaxLinks)
			}
			emitted[link.ID] = link
		}
	}
	result := graphGenericResult{Projection: "summary", All: in.All, LinkTypes: append([]string{}, in.LinkTypes...), ExactLinkTypes: append([]string{}, in.ExactLinkTypes...), AllLinkTypes: in.AllLinkTypes, Scope: scope, Root: in.Root, Direction: in.Direction, Depth: in.Depth, MaxNodes: in.MaxNodes, MaxLinks: in.MaxLinks, Nodes: []graphGenericNode{}, Links: []graphGenericLink{}, Frontier: []string{}}
	omittedTypes := map[string]bool{}
	for id := range distance {
		result.Nodes = append(result.Nodes, nodes[id])
		for _, omitted := range excludedIncidence[id] {
			omittedTypes[omitted] = true
		}
		// Frontier identifies reached nodes with eligible directional neighbors
		// beyond the depth boundary. Induced cross-links do not imply truncation.
		for _, link := range adjacency[id] {
			if _, seen := emitted[link.ID]; !seen {
				result.Frontier = append(result.Frontier, id)
				break
			}
		}
	}
	for _, link := range emitted {
		result.Links = append(result.Links, link)
	}
	for omitted := range omittedTypes {
		result.OmittedLinkTypes = append(result.OmittedLinkTypes, omitted)
	}
	sort.Strings(result.OmittedLinkTypes)
	sort.Slice(result.Nodes, func(i, j int) bool { return graph.CompareCodeUnits(result.Nodes[i].ID, result.Nodes[j].ID) < 0 })
	sort.Slice(result.Links, func(i, j int) bool { return graph.CompareCodeUnits(result.Links[i].ID, result.Links[j].ID) < 0 })
	sort.Slice(result.Frontier, func(i, j int) bool { return graph.CompareCodeUnits(result.Frontier[i], result.Frontier[j]) < 0 })
	result.Complete = len(result.Frontier) == 0
	return result, nil
}

// Complete Resources use the same traversal as legacy native projections.
// Type membership comes only from facts captured with those exact records;
// a partial inventory descriptor or a current re-read cannot supply ancestry.
func projectResourceGraphGeneric(snapshot graphstore.ResourceSnapshot, scope string, in graphGenericInput) (graphGenericResult, error) {
	zero := graphGenericResult{}
	invalid := func(message string) (graphGenericResult, error) {
		return zero, fmt.Errorf("%w: %s", graphstore.ErrInvalidStore, message)
	}
	if err := validateGraphGenericBounds(in); err != nil {
		return zero, fmt.Errorf("%w: %v", graph.ErrValidation, err)
	}
	if err := graph.ValidateScopeURL(scope); err != nil {
		return invalid("invalid snapshot Scope")
	}
	if len(snapshot.Documents) > graphstore.PreviewSnapshotLimit {
		return zero, fmt.Errorf("%w: current inventory exceeds %d Resources", graphstore.ErrLimitExceeded, graphstore.PreviewSnapshotLimit)
	}
	facts := make(map[resourceget.TypeReference]graphstore.ResourceTypeFact, len(snapshot.Types))
	for _, fact := range snapshot.Types {
		if fact.Reference.URI == "" || fact.Reference.Revision == "" || fact.Definition.URI == "" || fact.Definition.Revision == "" {
			return invalid("incomplete captured Type fact")
		}
		if _, duplicate := facts[fact.Reference]; duplicate {
			return invalid("duplicate exact captured Type fact")
		}
		facts[fact.Reference] = fact
	}
	// Resolve product selectors on this capture without changing the request or
	// the literal semantic predicates used for retained affiliation matching.
	resolvedSelectors := make([]string, 0, len(in.LinkTypes)+len(in.ExactLinkTypes))
	for _, selector := range append(append([]string{}, in.LinkTypes...), in.ExactLinkTypes...) {
		selected, err := snapshot.ResolveType(selector)
		if err != nil {
			return zero, err
		}
		if facts[selected].Describes != graph.KindLink {
			return zero, fmt.Errorf("%w: graph Link selection requires a Link Type", graph.ErrValidation)
		}
		resolvedSelectors = append(resolvedSelectors, selected.URI)
	}
	linkTypes, exactLinkTypes := resolvedSelectors[:len(in.LinkTypes)], resolvedSelectors[len(in.LinkTypes):]
	population := graphGenericPopulation{Nodes: map[string]graphGenericNode{}, Links: map[string]graphGenericLink{},
		Issues: map[string]bool{}, OpenIssues: map[string]bool{}, MatchingLinks: map[string]bool{}, DependencyLinks: map[string]bool{},
		IncludeIsolated: in.All && in.AllLinkTypes}
	seen := map[string]bool{}
	endpoint := func(raw json.RawMessage) (resourceget.TypeReference, error) {
		var ref resourceget.TypeReference
		if len(raw) != 0 && raw[0] == '"' {
			if err := json.Unmarshal(raw, &ref.URI); err != nil {
				return ref, err
			}
		} else if err := json.Unmarshal(raw, &ref); err != nil {
			return ref, err
		}
		if ref.URI == "" {
			return ref, fmt.Errorf("%w: incomplete captured endpoint", graphstore.ErrInvalidStore)
		}
		return ref, nil
	}
	for _, doc := range snapshot.Documents {
		if doc.ScopeURL() != scope || doc.ID() == "" || doc.Revision() == "" || seen[doc.ID()] {
			return invalid("invalid or duplicate captured Resource")
		}
		seen[doc.ID()] = true
		fact, ok := facts[doc.Type()]
		if !ok || fact.Describes != doc.Kind() {
			return invalid("Resource affiliation differs from its captured Type kind")
		}
		if _, err := snapshot.MatchesType(doc.Type(), "", false); err != nil {
			return zero, err
		}
		var body struct {
			Properties map[string]json.RawMessage `json:"properties"`
			Source     json.RawMessage            `json:"source"`
			Target     json.RawMessage            `json:"target"`
		}
		if err := json.Unmarshal(doc.NativeJSON(), &body); err != nil || body.Properties == nil {
			return invalid("incomplete captured Resource body")
		}
		// These are optional human-summary strings, not a validation schema.
		// A legal nonstring user property is retained in complete JSON output.
		var title, status string
		_ = json.Unmarshal(body.Properties["title"], &title)
		_ = json.Unmarshal(body.Properties["status"], &status)
		switch doc.Kind() {
		case graph.KindBead:
			population.Nodes[doc.ID()] = graphGenericNode{ID: doc.ID(), Type: doc.Type().URI, Title: title, Version: doc.Revision(), Status: status}
			issue, err := snapshot.MatchesType(doc.Type(), graphstore.IssueTypeURL(scope), false)
			if err != nil {
				return zero, err
			}
			if issue {
				population.Issues[doc.ID()] = true
				population.OpenIssues[doc.ID()] = status == "open" || status == "in_progress" || status == "blocked"
			}
		case graph.KindLink:
			source, err := endpoint(body.Source)
			if err != nil {
				return zero, err
			}
			target, err := endpoint(body.Target)
			if err != nil {
				return zero, err
			}
			if _, err := graph.ScopeLocalPublicResourceID(scope, source.URI); err != nil {
				return zero, fmt.Errorf("%w: this local graph view cannot traverse a foreign endpoint", graphstore.ErrCapabilityUnavailable)
			}
			if _, err := graph.ScopeLocalPublicResourceID(scope, target.URI); err != nil {
				return zero, fmt.Errorf("%w: this local graph view cannot traverse a foreign endpoint", graphstore.ErrCapabilityUnavailable)
			}
			population.Links[doc.ID()] = graphGenericLink{ID: doc.ID(), Type: doc.Type().URI, Source: source.URI, Target: target.URI,
				Version: doc.Revision(), SourceRevision: source.Revision, TargetRevision: target.Revision}
			match := in.AllLinkTypes || len(in.LinkTypes) == 0 && len(in.ExactLinkTypes) == 0
			for _, selection := range []struct {
				values []string
				exact  bool
			}{{linkTypes, false}, {exactLinkTypes, true}} {
				for _, selector := range selection.values {
					matched, err := snapshot.MatchesType(doc.Type(), selector, selection.exact)
					if err != nil {
						return zero, err
					}
					match = match || matched
				}
			}
			population.MatchingLinks[doc.ID()] = match
			dependency, err := snapshot.MatchesType(doc.Type(), graphstore.DependencyTypeURL(scope), false)
			if err != nil {
				return zero, err
			}
			population.DependencyLinks[doc.ID()] = dependency
		default:
			return invalid("unsupported captured Resource kind")
		}
	}
	return selectGraphGenericPopulation(scope, in, population)
}
