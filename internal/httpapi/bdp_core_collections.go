package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/issueops"
)

const (
	coreCursorVersion = 1

	// coreEdgeAnchorChunk is how many sources one edge read asks about.
	coreEdgeAnchorChunk = 500

	// coreSelectorScanBudget bounds how many rows one Selector page reads. A
	// sparse Selector then answers a short or empty page with a next URL that
	// has advanced, which Core allows, rather than holding a request open.
	coreSelectorScanBudget = 5000
)

// coreCollectionParams are the once-only query members of a fixed inventory or
// an incident page, parsed and checked.
type coreCollectionParams struct {
	// fixed are the members a continuation repeats and a cursor is bound to.
	fixed url.Values
	// navigation are the members a continuation repeats that select the
	// population rather than filter it; the cursor's binding names them.
	navigation url.Values
	limit      int
	// limitText is the request's own limit, repeated on next when it had one.
	limitText string
	// after is the canonical local id the walk resumes after.
	after    string
	selector *coreSelector
	types    coreTypeSelection
	source   string
	target   string
	endpoint string
}

// coreQueryParams parses a query string strictly: url.Values alone would drop a
// malformed pair, and Core refuses repeated and unknown members.
func coreQueryParams(raw string) (url.Values, error) {
	values, err := url.ParseQuery(raw)
	if err != nil {
		return nil, errors.New("the query string is malformed")
	}
	for name, list := range values {
		if len(list) != 1 {
			return nil, errors.New("`" + name + "` may appear once")
		}
	}
	return values, nil
}

// coreRefusal is a query refusal: its code and its true message.
type coreRefusal struct{ code, detail string }

func coreInvalid(detail string) *coreRefusal {
	return &coreRefusal{coreCodeInvalidParameter, detail}
}

// parseCoreCollection reads the members in allowed out of values, which must
// hold nothing else. kind fixes which Type family conformsTo answers for.
func (p *coreProjector) parseCoreCollection(ctx context.Context, values url.Values, allowed []string, kind, binding string) (*coreCollectionParams, *coreRefusal) {
	for name := range values {
		if !slices.Contains(allowed, name) {
			return nil, coreInvalid("`" + name + "` is not a parameter of this read")
		}
	}
	params := &coreCollectionParams{fixed: url.Values{}, limit: corePageDefaultItems}
	for _, name := range allowed {
		if name == "limit" || name == "cursor" {
			continue
		}
		if value, present := values[name]; present {
			if value[0] == "" {
				return nil, coreInvalid("`" + name + "` is empty")
			}
			params.fixed.Set(name, value[0])
		}
	}
	if text, present := values["limit"]; present {
		limit, err := strconv.Atoi(text[0])
		if err != nil || limit < 1 {
			return nil, coreInvalid("`limit` is a positive integer")
		}
		if limit > corePageMaximumItems {
			return nil, &coreRefusal{coreCodeLimitExceeded, "`limit` is above this Scope's advertised maximum of " + strconv.Itoa(corePageMaximumItems)}
		}
		params.limit, params.limitText = limit, text[0]
	}
	if text := params.fixed.Get("selector"); text != "" {
		selector, err := parseCoreSelector(text)
		if errors.Is(err, errCoreSelectorLimit) {
			return nil, &coreRefusal{coreCodeLimitExceeded, err.Error()}
		}
		if err != nil {
			return nil, coreInvalid("`selector`: " + err.Error())
		}
		params.selector = selector
	}
	selection, refusal := p.parseCoreTypes(ctx, params.fixed, kind)
	if refusal != nil {
		return nil, refusal
	}
	params.types = selection
	for _, name := range []struct {
		param string
		field *string
	}{{"source", &params.source}, {"target", &params.target}, {"endpoint", &params.endpoint}} {
		if text := params.fixed.Get(name.param); text != "" {
			uri, ok := p.endpointOperand(text)
			if !ok {
				return nil, coreInvalid("`" + name.param + "` is neither a local Bead id nor an absolute URI")
			}
			*name.field = uri
		}
	}
	if text, present := values["cursor"]; present {
		after, ok := parseCoreCursor(text[0], coreCursorBinding(binding, params.fixed))
		if !ok {
			return nil, coreInvalid("`cursor` is not a continuation issued for this read")
		}
		params.after = after
	}
	return params, nil
}

// endpointOperand normalizes an endpoint filter: a local id becomes its
// canonical URL, an absolute URI stays exactly as sent.
func (p *coreProjector) endpointOperand(text string) (string, bool) {
	if coreHasURIScheme(text) {
		return text, true
	}
	local, ok := coreEscapeLocalID(text)
	if !ok {
		return "", false
	}
	return p.scope + local, true
}

func coreHasURIScheme(text string) bool {
	scheme, _, found := strings.Cut(text, ":")
	if !found || scheme == "" || !(scheme[0] >= 'a' && scheme[0] <= 'z' || scheme[0] >= 'A' && scheme[0] <= 'Z') {
		return false
	}
	return !strings.ContainsFunc(scheme, func(c rune) bool {
		return !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '+' || c == '-' || c == '.')
	})
}

// coreTypeSelection is the `type` and `conformsTo` predicates.
type coreTypeSelection struct {
	// exact is `type`: one nominal identity, compared to the record's URI.
	exact string
	// family is `conformsTo`, resolved: "" when absent or the whole family,
	// otherwise the one concrete identity it selects by reflexivity.
	family string
	// none says conformsTo names a family no record of this kind can be in.
	none bool
}

func (t coreTypeSelection) matches(typeURI string) bool {
	if t.none || (t.exact != "" && typeURI != t.exact) || (t.family != "" && typeURI != t.family) {
		return false
	}
	return true
}

// nativeBeadType is the built-in classification the predicates pin, for pushing
// into the list role, or "" when they pin none it can validate.
func (t coreTypeSelection) nativeBeadType() string {
	for _, uri := range []string{t.exact, t.family} {
		if name, found := strings.CutPrefix(uri, coreBeadTypeNamespace); found && types.IssueType(name).IsValid() {
			return name
		}
	}
	return ""
}

// parseCoreTypes reads `type` and `conformsTo`.
//
// `type` is an exact comparison needing no interpretation, so a well-formed URI
// nothing here holds selects nothing: an empty page, because that is the true
// answer. conformsTo is different. Core says an interpretation that is not
// available refuses rather than falling back to nominal equality, and Issue §8
// makes the fixed catalog the only interpretation this Scope has, so a URI
// outside it is invalid-parameter.
func (p *coreProjector) parseCoreTypes(ctx context.Context, fixed url.Values, kind string) (coreTypeSelection, *coreRefusal) {
	var selection coreTypeSelection
	if text := fixed.Get("type"); text != "" {
		if !coreAbsoluteHTTPURL(text) {
			return selection, coreInvalid("`type` is an unpinned absolute HTTP(S) identity URI")
		}
		selection.exact = text
		selection.none = !p.holdableType(text, kind)
	}
	text := fixed.Get("conformsTo")
	if text == "" {
		return selection, nil
	}
	if !coreAbsoluteHTTPURL(text) {
		return selection, coreInvalid("`conformsTo` is an unpinned absolute HTTP(S) identity URI")
	}
	refused := coreInvalid("`conformsTo` names a Type outside this Scope's fixed catalog, whose conformance it cannot interpret")

	family, namespace, local, names := coreBeadFamily, coreBeadTypeNamespace, coreLocalBeadTypes, coreBeadTypeNames
	otherFamily, otherNamespace, otherLocal, otherNames := coreLinkFamily, coreLinkTypeNamespace, coreLocalLinkTypes, coreLinkTypeNames
	if kind == coreKindLink {
		family, otherFamily = otherFamily, family
		namespace, otherNamespace = otherNamespace, namespace
		local, otherLocal = otherLocal, local
		names, otherNames = otherNames, names
	}
	switch {
	case text == family:
		return selection, nil
	case strings.HasPrefix(text, namespace) && names[strings.TrimPrefix(text, namespace)]:
		selection.family = text
		return selection, nil
	case strings.HasPrefix(text, p.scope+local):
		name, err := url.PathUnescape(strings.TrimPrefix(text, p.scope+local))
		if err != nil || name == "" {
			return selection, refused
		}
		// A Link's kinds are an open vocabulary, so any name is one a row could
		// hold. An Issue's are configured, and an unconfigured one is not an
		// identity this Scope has.
		if kind == coreKindBead {
			configured, err := p.configuredTypes(ctx)
			if err != nil || !configured[name] {
				return selection, refused
			}
		}
		selection.family = text
		return selection, nil
	}
	// The other kind's family is in the catalog too, so the answer is known:
	// no record of this kind conforms to it.
	if text == otherFamily || strings.HasPrefix(text, p.scope+otherLocal) ||
		(strings.HasPrefix(text, otherNamespace) && otherNames[strings.TrimPrefix(text, otherNamespace)]) {
		selection.none = true
		return selection, nil
	}
	return selection, refused
}

// holdableType reports whether a record of this kind could be affiliated with
// uri: a concrete identity of its table, or a Scope-local one. The abstract bases
// admit no direct affiliation. Anything else selects nothing without a scan.
func (p *coreProjector) holdableType(uri, kind string) bool {
	namespace, local, names := coreBeadTypeNamespace, coreLocalBeadTypes, coreBeadTypeNames
	if kind == coreKindLink {
		namespace, local, names = coreLinkTypeNamespace, coreLocalLinkTypes, coreLinkTypeNames
	}
	if name, found := strings.CutPrefix(uri, namespace); found {
		return names[name]
	}
	return strings.HasPrefix(uri, p.scope+local)
}

func coreAbsoluteHTTPURL(text string) bool {
	parsed, err := url.Parse(text)
	return err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != "" && parsed.Fragment == ""
}

// coreCursorBinding names what a continuation may be used for: this read with
// these predicates. limit stays out of it, because Core lets a client resize one.
func coreCursorBinding(binding string, fixed url.Values) string {
	sum := sha256.Sum256([]byte(binding + "?" + fixed.Encode()))
	return hex.EncodeToString(sum[:8])
}

type coreCursor struct {
	Version int    `json:"v"`
	After   string `json:"a"`
	Binding string `json:"b"`
}

func encodeCoreCursor(after, binding string) string {
	body, err := json.Marshal(coreCursor{Version: coreCursorVersion, After: after, Binding: binding})
	if err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(body)
}

func parseCoreCursor(token, binding string) (string, bool) {
	body, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return "", false
	}
	var cursor coreCursor
	if err := json.Unmarshal(body, &cursor); err != nil {
		return "", false
	}
	if cursor.Version != coreCursorVersion || cursor.After == "" || cursor.Binding != binding {
		return "", false
	}
	return cursor.After, true
}

// coreNextURL is the server-generated absolute continuation.
func (p *coreProjector) coreNextURL(path string, params *coreCollectionParams, binding, after string) string {
	query := url.Values{}
	for name, values := range params.fixed {
		query[name] = values
	}
	for name, values := range params.navigation {
		query[name] = values
	}
	if params.limitText != "" {
		query.Set("limit", params.limitText)
	}
	query.Set("cursor", encodeCoreCursor(after, coreCursorBinding(binding, params.fixed)))
	return p.scope + strings.TrimPrefix(path, "/") + "?" + query.Encode()
}

// coreBeadPage is the closed beadCollection schema.
type coreBeadPage struct {
	Items []*coreBead `json:"items"`
	Next  *string     `json:"next"`
}

// coreLinkPage is the closed linkCollection schema.
type coreLinkPage struct {
	Items []*coreLink `json:"items"`
	Next  *string     `json:"next"`
}

// beadCollectionMembers are the Bead inventory's parameters.
var beadCollectionMembers = []string{"type", "conformsTo", "selector", "limit", "cursor"}

// linkCollectionMembers are the Link inventory's and the incident view's.
var linkCollectionMembers = []string{"type", "conformsTo", "source", "target", "endpoint", "selector", "limit", "cursor"}

// serveCoreBeads answers GET <scope>/bdp/beads.
func (s *Server) serveCoreBeads(w http.ResponseWriter, r *http.Request) {
	values, err := coreQueryParams(r.URL.RawQuery)
	if err != nil {
		writeCoreProblem(w, r, coreCodeInvalidParameter, err.Error())
		return
	}
	release, ok := s.coreAcquire(w, r)
	if !ok {
		return
	}
	defer release()

	p := newCoreProjector(s, r)
	params, refusal := p.parseCoreCollection(r.Context(), values, beadCollectionMembers, coreKindBead, coreBeadsPath)
	if refusal != nil {
		writeCoreProblem(w, r, refusal.code, refusal.detail)
		return
	}
	if !coreAcceptable(w, r) {
		return
	}
	page, err := p.beadPage(r.Context(), params)
	if err != nil {
		s.failCoreRead(w, r, err)
		return
	}
	writeCoreCollection(w, r, p.scope, coreKindBead, "", page)
}

// coreAcceptable answers the bodyless 406 and reports whether to go on.
func coreAcceptable(w http.ResponseWriter, r *http.Request) bool {
	if acceptsMedia(r.Header.Values("Accept")) {
		return true
	}
	w.Header().Set("Cache-Control", "private, no-store")
	w.WriteHeader(http.StatusNotAcceptable)
	return false
}

// writeCoreCollection answers one page with Core's collection metadata. A page
// has no entity tag, so only `*` conditions can apply to it.
func writeCoreCollection(w http.ResponseWriter, r *http.Request, scope, kind, subjectKind string, page any) {
	h := w.Header()
	setCoreNavigation(h, scope)
	h.Set("BDP-Collection-Kind", kind)
	if subjectKind != "" {
		h.Set("BDP-Resource-Kind", subjectKind)
	}
	if coreNotModified(w, r, "") {
		return
	}
	writeCoreJSON(w, page)
}

// coreAcquire buys a database slot the way s.route does, for a read that goes on
// to use roles.
func (s *Server) coreAcquire(w http.ResponseWriter, r *http.Request) (func(), bool) {
	release, err := s.acquire(r.Context(), requestInfo(r.Context()))
	if err != nil {
		s.failCoreErr(w, r, err)
		return nil, false
	}
	return release, true
}

// beadIDs lists every id in the Scope in canonical-uri order: every status and
// every kind of row, both planes, because the inventory hides nothing (Issue
// HD-17). The store has no id keyset, so this is a whole-table projection on
// every page, and only the page's rows are then hydrated.
func (p *coreProjector) beadIDs(ctx context.Context, nativeType string) ([]coreBeadKey, error) {
	reader, err := p.s.reader(p.r)
	if err != nil {
		return nil, err
	}
	unlimited := 0
	page, err := reader.List(ctx, issueops.ListRequest{
		IssueType:        nativeType,
		AllFlag:          true,
		IncludeTemplates: true,
		IncludeGates:     true,
		IncludeInfra:     true,
		IncludeEphemeral: true,
		Brief:            true,
		SkipLabels:       true,
		SkipCounts:       true,
		SortBy:           "id",
		Limit:            &unlimited,
	})
	if err != nil {
		return nil, err
	}
	keys := make([]coreBeadKey, 0, len(page.Items))
	for _, item := range page.Items {
		if item == nil || item.Issue == nil {
			continue
		}
		local, ok := coreEscapeLocalID(item.ID)
		if !ok {
			return nil, errCoreUnservable{"a record's native id " + strconv.Quote(item.ID) + " is outside Core's identifier grammar"}
		}
		keys = append(keys, coreBeadKey{id: item.ID, local: local})
	}
	// Escaped local ids are ASCII, so byte order is Core's code-unit order, and
	// a shared Scope prefix leaves the order of the absolute URLs unchanged.
	slices.SortFunc(keys, func(a, b coreBeadKey) int { return strings.Compare(a.local, b.local) })
	return keys, nil
}

type coreBeadKey struct{ id, local string }

// beadPage walks the ids after the cursor, hydrating a batch at a time, until
// the page is full or the Selector's scan budget is spent. The cursor is a
// keyset over live rows, so a row written between two pages is seen or missed
// according to where it falls: Core's frozen snapshot is not met (Issue HD-22).
func (p *coreProjector) beadPage(ctx context.Context, params *coreCollectionParams) (*coreBeadPage, error) {
	page := &coreBeadPage{Items: []*coreBead{}}
	if params.types.none {
		return page, nil
	}
	keys, err := p.beadIDs(ctx, params.types.nativeBeadType())
	if err != nil {
		return nil, err
	}
	start := 0
	if params.after != "" {
		start, _ = slices.BinarySearchFunc(keys, params.after, func(key coreBeadKey, after string) int {
			if key.local <= after {
				return -1
			}
			return 1
		})
	}
	getter, err := p.s.batchGetter(p.r)
	if err != nil {
		return nil, err
	}
	scanned, position := 0, start
	for position < len(keys) && len(page.Items) < params.limit && scanned < coreSelectorScanBudget {
		batch := keys[position:min(position+coreBatchSize(params), len(keys))]
		ids := make([]string, len(batch))
		for i, key := range batch {
			ids[i] = key.id
		}
		result, err := getter.GetMany(ctx, issueops.GetManyRequest{IDs: ids})
		if err != nil {
			return nil, err
		}
		rows := make(map[string]*types.Issue, len(result.Issues))
		for _, issue := range result.Issues {
			if issue != nil {
				rows[issue.ID] = issue
			}
		}
		for _, key := range batch {
			position++
			scanned++
			issue := rows[key.id]
			if issue == nil {
				continue // Deleted since the id list: a keyset walk just passes it.
			}
			record, err := p.bead(ctx, issue)
			if err != nil {
				return nil, err
			}
			if !params.types.matches(record.Type) ||
				(params.selector != nil && !params.selector.matches(record.candidate())) {
				continue
			}
			page.Items = append(page.Items, record)
			if len(page.Items) == params.limit {
				break
			}
		}
	}
	if position < len(keys) && position > 0 {
		next := p.coreNextURL(coreBeadsPath, params, coreBeadsPath, keys[position-1].local)
		page.Next = &next
	}
	return page, nil
}

// coreBatchSize is how many rows one hydration asks for. Without a predicate
// that needs the row, the page itself; with one, more, since most may not match.
func coreBatchSize(params *coreCollectionParams) int {
	if params.selector == nil && params.types.exact == "" && params.types.family == "" {
		return params.limit
	}
	return max(params.limit, corePageDefaultItems)
}

// serveCoreLinks answers GET <scope>/bdp/links.
func (s *Server) serveCoreLinks(w http.ResponseWriter, r *http.Request) {
	values, err := coreQueryParams(r.URL.RawQuery)
	if err != nil {
		writeCoreProblem(w, r, coreCodeInvalidParameter, err.Error())
		return
	}
	release, ok := s.coreAcquire(w, r)
	if !ok {
		return
	}
	defer release()

	p := newCoreProjector(s, r)
	params, refusal := p.parseCoreCollection(r.Context(), values, linkCollectionMembers, coreKindLink, coreLinksPath)
	if refusal != nil {
		writeCoreProblem(w, r, refusal.code, refusal.detail)
		return
	}
	if !coreAcceptable(w, r) {
		return
	}
	edges, live, err := p.inventoryEdges(r.Context(), params)
	if err != nil {
		s.failCoreRead(w, r, err)
		return
	}
	page, err := p.linkPage(edges, live, params, coreLinksPath, coreLinksPath)
	if err != nil {
		s.failCoreRead(w, r, err)
		return
	}
	writeCoreCollection(w, r, p.scope, coreKindLink, "", page)
}

// inventoryEdges reads the edges the inventory's predicates could select, and
// the live ids their targets are judged against. A `source` naming a Bead here
// anchors the read on that one Bead. Otherwise every source is read, because the
// edge role is keyed by source and the store offers no whole-table edge read.
func (p *coreProjector) inventoryEdges(ctx context.Context, params *coreCollectionParams) ([]*issueops.Dependency, map[string]bool, error) {
	if params.source != "" {
		if id, ok := p.localBeadID(params.source); ok {
			edges, err := p.outgoingEdges(ctx, []string{id})
			if err != nil {
				return nil, nil, err
			}
			live, err := p.liveTargets(ctx, edges)
			return edges, live, err
		}
		return nil, map[string]bool{}, nil
	}
	keys, err := p.beadIDs(ctx, "")
	if err != nil {
		return nil, nil, err
	}
	ids := make([]string, len(keys))
	live := make(map[string]bool, len(keys))
	for i, key := range keys {
		ids[i] = key.id
		live[key.id] = true
	}
	edges, err := p.outgoingEdges(ctx, ids)
	return edges, live, err
}

// localBeadID undoes beadURL for a URL in this Scope.
func (p *coreProjector) localBeadID(uri string) (string, bool) {
	local, found := strings.CutPrefix(uri, p.scope)
	if !found || local == "" {
		return "", false
	}
	id, err := url.PathUnescape(local)
	if err != nil {
		return "", false
	}
	return id, true
}

// outgoingEdges reads the stored outgoing edges of ids, a chunk at a time.
func (p *coreProjector) outgoingEdges(ctx context.Context, ids []string) ([]*issueops.Dependency, error) {
	reader, err := p.s.edgeReader(p.r)
	if err != nil {
		return nil, err
	}
	var edges []*issueops.Dependency
	for chunk := range slices.Chunk(ids, coreEdgeAnchorChunk) {
		result, err := reader.ReadEdges(ctx, issueops.EdgeReadRequest{IDs: chunk})
		if err != nil {
			return nil, err
		}
		for _, anchor := range result.Anchors {
			for _, edge := range anchor.Edges {
				if edge != nil {
					edges = append(edges, edge)
				}
			}
		}
	}
	return edges, nil
}

// liveTargets reports which of the edges' non-external targets hold a row here.
func (p *coreProjector) liveTargets(ctx context.Context, edges []*issueops.Dependency) (map[string]bool, error) {
	var ids []string
	for _, edge := range edges {
		if !strings.HasPrefix(edge.DependsOnID, coreExternalPrefix) {
			ids = append(ids, edge.DependsOnID)
		}
	}
	slices.Sort(ids)
	ids = slices.Compact(ids)
	getter, err := p.s.batchGetter(p.r)
	if err != nil {
		return nil, err
	}
	live := make(map[string]bool, len(ids))
	for chunk := range slices.Chunk(ids, issueops.MaxGetManyIDs) {
		result, err := getter.GetMany(ctx, issueops.GetManyRequest{IDs: chunk})
		if err != nil {
			return nil, err
		}
		for _, issue := range result.Issues {
			if issue != nil {
				live[issue.ID] = true
			}
		}
	}
	return live, nil
}

// linkPage selects, orders and pages a set of edges. It is all in memory, because
// a Link's identity and every predicate are derived from the whole edge, so the
// set is projected before it is ordered; only the page's Selector stops early.
func (p *coreProjector) linkPage(edges []*issueops.Dependency, live map[string]bool, params *coreCollectionParams, path, binding string) (*coreLinkPage, error) {
	page := &coreLinkPage{Items: []*coreLink{}}
	if params.types.none {
		return page, nil
	}
	records := make([]*coreLink, 0, len(edges))
	for _, edge := range edges {
		record, err := p.link(edge, live)
		if err != nil {
			return nil, err
		}
		if params.after != "" && strings.TrimPrefix(record.ID, p.scope) <= params.after {
			continue
		}
		if !params.types.matches(record.Type) ||
			(params.source != "" && record.Source != params.source) ||
			(params.target != "" && record.Target != params.target) ||
			(params.endpoint != "" && record.Source != params.endpoint && record.Target != params.endpoint) {
			continue
		}
		records = append(records, record)
	}
	slices.SortFunc(records, func(a, b *coreLink) int { return strings.Compare(a.ID, b.ID) })
	records = slices.CompactFunc(records, func(a, b *coreLink) bool { return a.ID == b.ID })
	for i, record := range records {
		if params.selector != nil && !params.selector.matches(record.candidate()) {
			continue
		}
		page.Items = append(page.Items, record)
		if len(page.Items) == params.limit {
			if i+1 < len(records) {
				next := p.coreNextURL(path, params, binding, strings.TrimPrefix(record.ID, p.scope))
				page.Next = &next
			}
			break
		}
	}
	return page, nil
}
