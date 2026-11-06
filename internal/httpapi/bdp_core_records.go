package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/steveyegge/beads/internal/config"
	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/issueops"
)

// Issue and Dependency Types §2: the shared identities, one fixed table in the
// provider. A native name is a suffix of its namespace, and the tables below are
// what stops a suffix being derived from anything else.
const (
	coreBeadTypeNamespace = "https://github.com/gastownhall/bdp/types/issue/"
	coreLinkTypeNamespace = "https://github.com/gastownhall/bdp/types/dependency/"
	coreBeadFamily        = coreBeadTypeNamespace + "issue"
	coreLinkFamily        = coreLinkTypeNamespace + "dependency"

	coreLocalBeadTypes = "types/issue/"
	coreLocalLinkTypes = "types/dependency/"

	coreKindBead = "bead"
	coreKindLink = "link"

	// coreLinkIDPrefix opens every Link's one opaque segment. The native id
	// grammar is prefix-hash, so it cannot be mistaken for a Bead's.
	coreLinkIDPrefix = "dep~"

	// coreForeignScheme spells a stored target naming no row here as an opaque
	// absolute URI. Issue §4.1 leaves the spelling of a cross-repository
	// identifier to the provider (preview). A dangling row reads the same way,
	// because nothing stored tells the two apart.
	coreForeignScheme = "beads-id:"

	coreExternalPrefix = "external:"
)

var coreBeadTypeNames = map[string]bool{
	"task": true, "bug": true, "feature": true, "epic": true, "chore": true, "decision": true,
	"message": true, "molecule": true, "gate": true, "spike": true, "story": true,
	"milestone": true, "event": true,
}

var coreLinkTypeNames = map[string]bool{
	"blocks": true, "parent-child": true, "conditional-blocks": true, "waits-for": true,
	"related": true, "discovered-from": true, "replies-to": true, "relates-to": true,
	"duplicates": true, "supersedes": true, "authored-by": true, "assigned-to": true,
	"approved-by": true, "attests": true, "tracks": true, "until": true, "caused-by": true,
	"validates": true, "delegated-from": true,
}

// coreDefaultInfraTypes is what an unconfigured workspace calls infrastructure,
// the same fallback the HTTP client spells for itself: this package reaches
// storage through roles, and no role answers it.
var coreDefaultInfraTypes = []string{"agent", "role", "message"}

// coreInstantProperties are the Issue properties §3.2 types as instants. Core
// emits every instant in UTC with uppercase T and Z.
var coreInstantProperties = []string{
	"created_at", "updated_at", "started_at", "closed_at", "due_at", "defer_until", "compacted_at",
}

// coreBeadOmitted are the native members §3.4 keeps out of a Bead's properties:
// its identity, its classification, its common metadata, relation projections,
// and the operational observations that change without changing the record.
var coreBeadOmitted = []string{
	"id", "issue_type", "metadata", "dependencies", "comments",
	"lease_expires_at", "heartbeat_at", "lease_granted_node", "is_blocked",
}

// coreBead is a Bead record, and its members are in the schema's order.
type coreBead struct {
	ID         string         `json:"id"`
	Type       string         `json:"type"`
	Revision   string         `json:"revision"`
	Properties map[string]any `json:"properties"`
	Metadata   map[string]any `json:"metadata"`
}

// coreLink is a Link record.
type coreLink struct {
	ID         string         `json:"id"`
	Type       string         `json:"type"`
	Revision   string         `json:"revision"`
	Source     string         `json:"source"`
	Target     string         `json:"target"`
	Properties map[string]any `json:"properties"`
	Metadata   map[string]any `json:"metadata"`
}

// candidate is the value a Selector sees: Core's Selection fixes its members,
// so a revision is not among them.
func (b *coreBead) candidate() map[string]any {
	return map[string]any{"id": b.ID, "type": b.Type, "properties": b.Properties, "metadata": b.Metadata}
}

func (l *coreLink) candidate() map[string]any {
	return map[string]any{
		"id": l.ID, "type": l.Type, "source": l.Source, "target": l.Target,
		"properties": l.Properties, "metadata": l.Metadata,
	}
}

// errCoreUnservable is a stored value Core has no honest spelling for. §2 says
// such a record is not served, and refused rather than given an invented
// identity; leaving it out of a page would be a page that silently lies.
type errCoreUnservable struct{ reason string }

func (e errCoreUnservable) Error() string { return e.reason }

// coreProjector turns one request's rows into records.
type coreProjector struct {
	s     *Server
	r     *http.Request
	scope string
	// configured is the workspace's configured Issue classifications, read the
	// first time a row names one the fixed table does not hold.
	configured map[string]bool
}

func newCoreProjector(s *Server, r *http.Request) *coreProjector {
	return &coreProjector{s: s, r: r, scope: coreScope(r)}
}

// configuredTypes is what the native tracker admits by configuration: its custom
// types and its infra types, from the database and from config.yaml.
func (p *coreProjector) configuredTypes(ctx context.Context) (map[string]bool, error) {
	if p.configured != nil {
		return p.configured, nil
	}
	settings, err := p.s.workspaceConfig(p.r)
	if err != nil {
		return nil, err
	}
	read := func(key string) ([]string, error) {
		result, err := settings.GetSetting(ctx, issueops.GetSettingRequest{Key: key})
		if err != nil {
			return nil, err
		}
		var names []string
		for _, name := range strings.Split(result.Value, ",") {
			if name = strings.TrimSpace(name); name != "" {
				names = append(names, name)
			}
		}
		return names, nil
	}
	custom, err := read("types.custom")
	if err != nil {
		return nil, err
	}
	infra, err := read("types.infra")
	if err != nil {
		return nil, err
	}
	if len(infra) == 0 {
		infra = config.GetInfraTypesFromYAML()
	}
	if len(infra) == 0 {
		infra = coreDefaultInfraTypes
	}
	set := map[string]bool{}
	for _, names := range [][]string{custom, config.GetCustomTypesFromYAML(), infra} {
		for _, name := range names {
			set[name] = true
		}
	}
	p.configured = set
	return set, nil
}

// beadType is §2's mapping of a native classification. An empty stored value is
// the column default, task. A name in neither the fixed table nor the
// configuration has no identity, and §2 forbids inventing one.
func (p *coreProjector) beadType(ctx context.Context, name string) (string, error) {
	if name == "" {
		name = string(types.TypeTask)
	}
	if coreBeadTypeNames[name] {
		return coreBeadTypeNamespace + name, nil
	}
	configured, err := p.configuredTypes(ctx)
	if err != nil {
		return "", err
	}
	if !configured[name] {
		return "", errCoreUnservable{"a record's classification " + strconv.Quote(name) + " has no identity in this Scope's type table"}
	}
	return p.scope + coreLocalBeadTypes + url.PathEscape(name), nil
}

// linkType maps a native edge kind. The native kind vocabulary is open with no
// configuration behind it, so every kind outside §2's table is Scope-local: there
// is nothing to check it against, and refusing would make every read of a
// workspace with its own kind unservable.
func (p *coreProjector) linkType(name string) string {
	if coreLinkTypeNames[name] {
		return coreLinkTypeNamespace + name
	}
	return p.scope + coreLocalLinkTypes + url.PathEscape(name)
}

// coreEscapeLocalID spells a native id as its local Resource ID, or refuses one
// this provider cannot serve (Issue §3.2 binds its identifier policy to Core's
// grammar). The policy is one segment. Native ids are prefix-hash, and with it a
// path with a slash in it stays the unrouted path it has always been, rather than
// every mistyped URL becoming a credentialed database read.
func coreEscapeLocalID(id string) (string, bool) {
	if !coreSafeSegment(id) {
		return "", false
	}
	return url.PathEscape(id), true
}

// coreSafeSegment is one decoded segment of Core's local-ID grammar.
func coreSafeSegment(segment string) bool {
	if segment == "" || segment == "." || segment == ".." || !utf8.ValidString(segment) {
		return false
	}
	return !strings.ContainsFunc(segment, func(c rune) bool {
		return c == '/' || c == '\\' || unicode.IsControl(c)
	})
}

// beadURL is the canonical URL of a native id.
func (p *coreProjector) beadURL(id string) (string, error) {
	local, ok := coreEscapeLocalID(id)
	if !ok {
		return "", errCoreUnservable{"a record's native id " + strconv.Quote(id) + " is outside Core's identifier grammar"}
	}
	return p.scope + local, nil
}

// bead projects one row into its record.
func (p *coreProjector) bead(ctx context.Context, issue *types.Issue) (*coreBead, error) {
	id, err := p.beadURL(issue.ID)
	if err != nil {
		return nil, err
	}
	typeURI, err := p.beadType(ctx, string(issue.IssueType))
	if err != nil {
		return nil, err
	}
	properties, err := coreBeadProperties(issue)
	if err != nil {
		return nil, err
	}
	metadata := coreMetadataObject(issue.Metadata)
	digest, err := coreDigest(struct {
		Type       string         `json:"type"`
		Properties map[string]any `json:"properties"`
		Metadata   map[string]any `json:"metadata"`
	}{typeURI, properties, metadata})
	if err != nil {
		return nil, err
	}
	return &coreBead{
		ID:         id,
		Type:       typeURI,
		Revision:   types.RevisionToken(issue.RowVersion) + "-" + digest,
		Properties: properties,
		Metadata:   metadata,
	}, nil
}

// coreBeadProperties is §3.2 and §3.4: the row's own JSON, less what the record
// carries elsewhere or not at all. Unknown native members survive by being
// removed from nothing, which is how a new durable column reaches a client.
func coreBeadProperties(issue *types.Issue) (map[string]any, error) {
	raw, err := json.Marshal(issue)
	if err != nil {
		return nil, fmt.Errorf("marshal issue: %w", err)
	}
	properties, err := coreDecodeObject(raw)
	if err != nil {
		return nil, fmt.Errorf("decode issue: %w", err)
	}
	for _, name := range coreBeadOmitted {
		delete(properties, name)
	}
	coreNormalizeInstants(properties)
	// §3.2 makes labels a set. The revision depends on the array, so a store
	// that handed the same set back in another order must not move it.
	if labels, ok := properties["labels"].([]any); ok {
		slices.SortFunc(labels, func(a, b any) int { return strings.Compare(fmt.Sprint(a), fmt.Sprint(b)) })
	}
	return properties, nil
}

// coreNormalizeInstants rewrites the instants to UTC, keeping the exact time.
func coreNormalizeInstants(properties map[string]any) {
	for _, name := range coreInstantProperties {
		text, ok := properties[name].(string)
		if !ok {
			continue
		}
		if parsed, err := time.Parse(time.RFC3339Nano, text); err == nil {
			properties[name] = parsed.UTC().Format(time.RFC3339Nano)
		}
	}
}

// coreDecodeObject decodes a JSON object, keeping number literals exactly.
func coreDecodeObject(raw []byte) (map[string]any, error) {
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.UseNumber()
	var value map[string]any
	if err := dec.Decode(&value); err != nil {
		return nil, err
	}
	if value == nil {
		return nil, errors.New("not an object")
	}
	return value, nil
}

// coreMetadataObject is §4.2's rule, applied to both Resource kinds: a stored
// JSON object is served as that object, and anything else as {}, with the native
// bytes kept until a write replaces them.
func coreMetadataObject(raw []byte) map[string]any {
	if metadata, err := coreDecodeObject(raw); err == nil {
		return metadata
	}
	return map[string]any{}
}

// coreDigest names the content half of a revision: scheme
// sha256-gojson-literal, SHA-256 over this package's JSON spelling with object
// members sorted and every number kept as its stored literal. Core requires a
// content-derived token to declare its number model, and this is that
// declaration: it is not sha256-jcs, because no peer verifies it and the stored
// literal is the value Core's exact-decimal comparison would see.
func coreDigest(value any) (string, error) {
	body, err := marshalCoreJSON(value)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:16]), nil
}

// coreLinkID allocates a Link's one opaque segment from its ordered pair and a
// generation (Issue §5). The stores keep no generation counter, so the row's
// creation second stands in for one: removing a pair's Link and adding it again
// in a later second is a new identity, and the old URL answers 404. The segment
// is reversible on purpose, because the roles can read a pair's row given its
// source and have no way to look one up by a hash.
func coreLinkID(source, target string, createdAt time.Time) string {
	plain := source + "\x00" + target + "\x00" + strconv.FormatInt(createdAt.Unix(), 10)
	return coreLinkIDPrefix + base64.RawURLEncoding.EncodeToString([]byte(plain))
}

// coreParseLinkID is coreLinkID's inverse.
func coreParseLinkID(segment string) (source, target string, generation int64, ok bool) {
	encoded, found := strings.CutPrefix(segment, coreLinkIDPrefix)
	if !found {
		return "", "", 0, false
	}
	plain, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return "", "", 0, false
	}
	parts := strings.Split(string(plain), "\x00")
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" {
		return "", "", 0, false
	}
	generation, err = strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return "", "", 0, false
	}
	return parts[0], parts[1], generation, true
}

// linkTarget spells a stored target: an in-Scope Bead's canonical URL, the
// native external form (already an opaque URI), or the foreign spelling for an
// id this database holds no row for.
func (p *coreProjector) linkTarget(target string, live map[string]bool) (string, error) {
	if strings.HasPrefix(target, coreExternalPrefix) {
		return target, nil
	}
	if live[target] {
		return p.beadURL(target)
	}
	return coreForeignScheme + url.PathEscape(target), nil
}

// link projects one edge into its record, given the set of live ids its target
// is judged against.
func (p *coreProjector) link(edge *issueops.Dependency, live map[string]bool) (*coreLink, error) {
	source, err := p.beadURL(edge.IssueID)
	if err != nil {
		return nil, err
	}
	target, err := p.linkTarget(edge.DependsOnID, live)
	if err != nil {
		return nil, err
	}
	properties := map[string]any{"created_at": edge.CreatedAt.UTC().Format(time.RFC3339Nano)}
	if edge.CreatedBy != "" {
		properties["created_by"] = edge.CreatedBy
	}
	if edge.ThreadID != "" {
		properties["thread_id"] = edge.ThreadID
	}
	metadata := coreMetadataObject([]byte(edge.Metadata))
	typeURI := p.linkType(string(edge.Type))
	digest, err := coreDigest(struct {
		Type       string         `json:"type"`
		Source     string         `json:"source"`
		Target     string         `json:"target"`
		Properties map[string]any `json:"properties"`
		Metadata   map[string]any `json:"metadata"`
	}{typeURI, source, target, properties, metadata})
	if err != nil {
		return nil, err
	}
	return &coreLink{
		ID:         p.scope + coreLinkID(edge.IssueID, edge.DependsOnID, edge.CreatedAt),
		Type:       typeURI,
		Revision:   digest,
		Source:     source,
		Target:     target,
		Properties: properties,
		Metadata:   metadata,
	}, nil
}

// failCoreRead answers a projection or role failure: an unservable row is
// Core's temporary failure, as §2 directs, and the rest is failCoreErr's.
func (s *Server) failCoreRead(w http.ResponseWriter, r *http.Request, err error) {
	var unservable errCoreUnservable
	if errors.As(err, &unservable) {
		writeCoreProblem(w, r, coreCodeUnavailable, unservable.reason)
		return
	}
	s.failCoreErr(w, r, err)
}
