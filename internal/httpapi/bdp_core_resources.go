package httpapi

import (
	"context"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/issueops"
)

// coreResourceLookup is the Resource a canonical URL names.
type coreResourceLookup struct {
	bead *types.Issue
	edge *issueops.Dependency
	live map[string]bool
}

// coreLocalPath decodes a Resource path, which coreResourcesOr has already held
// to one segment: decoded once and held to Core's grammar. It reports false for
// anything that cannot be a Resource address.
func coreLocalPath(r *http.Request) (string, bool) {
	decoded, err := url.PathUnescape(strings.TrimPrefix(r.URL.EscapedPath(), "/"))
	if err != nil || !coreSafeSegment(decoded) {
		return "", false
	}
	return decoded, true
}

// lookup finds the Resource a local id names, or nil for the uniform miss.
func (p *coreProjector) lookup(ctx context.Context, local string) (*coreResourceLookup, error) {
	if source, target, generation, ok := coreParseLinkID(local); ok {
		return p.lookupLink(ctx, source, target, generation)
	}
	// The same bound handleGetIssue holds an id to before spending a slot on it.
	if types.CheckFieldLen("id", local) != nil {
		return nil, nil
	}
	getter, err := p.s.batchGetter(p.r)
	if err != nil {
		return nil, err
	}
	result, err := getter.GetMany(ctx, issueops.GetManyRequest{IDs: []string{local}})
	if err != nil {
		return nil, err
	}
	for _, issue := range result.Issues {
		if issue != nil && issue.ID == local {
			return &coreResourceLookup{bead: issue}, nil
		}
	}
	return nil, nil
}

// lookupLink reads the pair's row from its source. A row of a later generation
// is a different Link, and the old address is the same miss as an unknown one.
func (p *coreProjector) lookupLink(ctx context.Context, source, target string, generation int64) (*coreResourceLookup, error) {
	edges, err := p.outgoingEdges(ctx, []string{source})
	if err != nil {
		return nil, err
	}
	for _, edge := range edges {
		if edge.DependsOnID == target && edge.CreatedAt.Unix() == generation {
			live, err := p.liveTargets(ctx, []*issueops.Dependency{edge})
			if err != nil {
				return nil, err
			}
			return &coreResourceLookup{edge: edge, live: live}, nil
		}
	}
	return nil, nil
}

// coreIncludeMembers are the query members include=links admits.
var coreIncludeMembers = []string{"include", "direction", "limit"}

// serveCoreResource answers GET and HEAD of a canonical Resource URL.
func (s *Server) serveCoreResource(w http.ResponseWriter, r *http.Request) {
	values, err := coreQueryParams(r.URL.RawQuery)
	if err != nil {
		writeCoreProblem(w, r, coreCodeInvalidParameter, err.Error())
		return
	}
	view, include := values.Get("view"), values.Get("include")
	refuse := func(detail string) { writeCoreProblem(w, r, coreCodeInvalidParameter, detail) }
	switch {
	case values.Has("view") && values.Has("include"):
		refuse("`include` cannot be combined with `view`")
		return
	case values.Has("view") && view != "properties" && view != "links":
		refuse("this Scope serves `view=properties` and `view=links`; it advertises nothing else")
		return
	case values.Has("include") && include != "links":
		refuse("this Scope serves `include=links` and nothing else")
		return
	case view == "properties" && len(values) != 1:
		refuse("`view=properties` takes no other parameter")
		return
	case !values.Has("view") && !values.Has("include") && len(values) != 0:
		for name := range values {
			refuse("`" + name + "` is not a parameter of this read")
			return
		}
	}
	direction := values.Get("direction")
	if values.Has("direction") && direction != "inbound" && direction != "outbound" && direction != "both" {
		refuse("`direction` is inbound, outbound or both")
		return
	}
	if include != "" {
		for name := range values {
			if !slices.Contains(coreIncludeMembers, name) {
				refuse("`" + name + "` is not a parameter of `include=links`")
				return
			}
		}
	}

	local, ok := coreLocalPath(r)
	if !ok {
		writeCoreProblem(w, r, coreCodeNotFound, "")
		return
	}
	release, ok := s.coreAcquire(w, r)
	if !ok {
		return
	}
	defer release()

	p := newCoreProjector(s, r)
	ctx := r.Context()

	var params *coreCollectionParams
	path := r.URL.EscapedPath()
	binding := path + "?view=links&direction=" + direction
	if view == "links" || include == "links" {
		members := linkCollectionMembers
		if include == "links" {
			members = []string{"limit"}
		}
		rest := url.Values{}
		for name, list := range values {
			if name != "view" && name != "include" && name != "direction" {
				rest[name] = list
			}
		}
		var refusal *coreRefusal
		params, refusal = p.parseCoreCollection(ctx, rest, members, coreKindLink, binding)
		if refusal != nil {
			writeCoreProblem(w, r, refusal.code, refusal.detail)
			return
		}
		params.navigation = url.Values{"view": {"links"}}
		if direction != "" {
			params.navigation.Set("direction", direction)
		}
	}

	found, err := p.lookup(ctx, local)
	if err != nil {
		s.failCoreRead(w, r, err)
		return
	}
	if found == nil {
		writeCoreProblem(w, r, coreCodeNotFound, "")
		return
	}
	if found.edge != nil && (view == "links" || include == "links") {
		refuse("`view=links` and `include=links` are valid only for a Bead")
		return
	}
	if !coreAcceptable(w, r) {
		return
	}

	if found.edge != nil {
		record, err := p.link(found.edge, found.live)
		if err != nil {
			s.failCoreRead(w, r, err)
			return
		}
		writeCorePoint(w, r, p.scope, coreKindLink, record.Revision, view, record, record.Properties)
		return
	}
	record, err := p.bead(ctx, found.bead)
	if err != nil {
		s.failCoreRead(w, r, err)
		return
	}
	switch {
	case view == "links":
		page, err := p.incidentPage(ctx, found.bead.ID, direction, params, path, binding)
		if err != nil {
			s.failCoreRead(w, r, err)
			return
		}
		writeCoreCollection(w, r, p.scope, coreKindLink, coreKindBead, page)
	case include == "links":
		page, err := p.incidentPage(ctx, found.bead.ID, direction, params, path, binding)
		if err != nil {
			s.failCoreRead(w, r, err)
			return
		}
		s.writeCoreAggregate(w, r, p.scope, record, page)
	default:
		writeCorePoint(w, r, p.scope, coreKindBead, record.Revision, view, record, record.Properties)
	}
}

// writeCorePoint answers a full record or its properties view. Both carry the
// Resource's revision as their entity tag.
func writeCorePoint(w http.ResponseWriter, r *http.Request, scope, kind, revision, view string, record any, properties map[string]any) {
	h := w.Header()
	setCoreNavigation(h, scope)
	h.Set("BDP-Resource-Kind", kind)
	h.Set("ETag", `"`+revision+`"`)
	if coreNotModified(w, r, `"`+revision+`"`) {
		return
	}
	if view == "properties" {
		writeCoreJSON(w, properties)
		return
	}
	writeCoreJSON(w, record)
}

// coreAggregate is the ordinary Bead record with its first incident page.
type coreAggregate struct {
	*coreBead
	Links *coreLinkPage `json:"links"`
}

// writeCoreAggregate answers include=links. Its entity tag covers the whole
// response, so it moves when the embedded page does even if the Bead did not.
func (s *Server) writeCoreAggregate(w http.ResponseWriter, r *http.Request, scope string, record *coreBead, page *coreLinkPage) {
	aggregate := coreAggregate{coreBead: record, Links: page}
	digest, err := coreDigest(aggregate)
	if err != nil {
		s.failCoreRead(w, r, err)
		return
	}
	h := w.Header()
	setCoreNavigation(h, scope)
	h.Set("BDP-Resource-Kind", coreKindBead)
	etag := `"` + digest + `"`
	h.Set("ETag", etag)
	if coreNotModified(w, r, etag) {
		return
	}
	writeCoreJSON(w, aggregate)
}

// incidentPage is view=links. Direction establishes the population, then the
// Link predicates apply before paging. A self-Link appears once.
func (p *coreProjector) incidentPage(ctx context.Context, id, direction string, params *coreCollectionParams, path, binding string) (*coreLinkPage, error) {
	if direction == "" {
		direction = "both"
	}
	var edges []*issueops.Dependency
	if direction != "inbound" {
		outgoing, err := p.outgoingEdges(ctx, []string{id})
		if err != nil {
			return nil, err
		}
		edges = append(edges, outgoing...)
	}
	if direction != "outbound" {
		incoming, err := p.incomingEdges(ctx, id)
		if err != nil {
			return nil, err
		}
		edges = append(edges, incoming...)
	}
	live, err := p.liveTargets(ctx, edges)
	if err != nil {
		return nil, err
	}
	return p.linkPage(edges, live, params, path, binding)
}

// incomingEdges finds the edges whose target is id. The neighbor role names the
// sources, and the edge role has the rows; there is no target-keyed edge read.
func (p *coreProjector) incomingEdges(ctx context.Context, id string) ([]*issueops.Dependency, error) {
	relations, err := p.s.relations(p.r)
	if err != nil {
		return nil, err
	}
	neighbors, err := relations.Related(ctx, issueops.RelatedRequest{ID: id, Direction: issueops.RelationIn})
	if err != nil {
		return nil, err
	}
	var sources []string
	for _, neighbor := range neighbors {
		if neighbor != nil {
			sources = append(sources, neighbor.ID)
		}
	}
	slices.Sort(sources)
	sources = slices.Compact(sources)
	edges, err := p.outgoingEdges(ctx, sources)
	if err != nil {
		return nil, err
	}
	return slices.DeleteFunc(edges, func(edge *issueops.Dependency) bool { return edge.DependsOnID != id }), nil
}
