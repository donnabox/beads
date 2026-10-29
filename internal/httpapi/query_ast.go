package httpapi

// This adapter selects the resource-selection version 1 proposal and its
// bdp-query-design-1 AST. It deliberately does not update bdpwire's older Read
// pin or ordinary GET/discovery representations. The enclosing server supplies
// authentication, Host checks, admission, deadlines and caller-owned storage.

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/httpapi/bdpwire"
	"github.com/steveyegge/beads/internal/httpapi/graphread"
	"github.com/steveyegge/beads/internal/resourcequery"
)

const queryASTMedia = "application/vnd.bdp.v0.1+json"
const queryASTCapability = "https://github.com/gastownhall/bdp/capabilities/resource-selection"
const queryASTRequestLimit = 262144

// QueryAST is an opt-in read-only adapter, never a listener or store owner.
// Both callbacks are bound to the same caller-owned checked Reader at creation.
type QueryAST struct {
	validateAuthority func(context.Context) error
	execute           func(context.Context, []byte) (resourcequery.Response, error)
}

// NewGraphReadWithQuery retains ordinary Read routing and adds the explicitly
// selected JSON AST operation at the persisted Scope root. Closing the handler
// does not close the caller-owned Reader/store.
func NewGraphReadWithQuery(reader *graphread.Reader, scope string) (*GraphRead, error) {
	g, err := NewGraphRead(reader, scope)
	if err != nil {
		return nil, err
	}
	g.queryAST = &QueryAST{validateAuthority: reader.ValidateAuthority, execute: reader.ExecuteResourceQuery}
	return g, nil
}

// Match only this operation's actual Scope root. In particular, a Resource whose
// local ID resembles an operation name retains its ordinary Read interpretation.
// POST root requests with missing selectors belong here so they fail closed.
func (q *QueryAST) matches(g *GraphRead, r *http.Request) bool {
	if r.URL == nil || r.URL.EscapedPath() != g.path {
		return false
	}
	if r.Method == http.MethodPost {
		return true
	}
	values, _ := url.ParseQuery(r.URL.RawQuery)
	for _, key := range []string{"readVersion", "operation", "cursor"} {
		if _, exists := values[key]; exists {
			return true
		}
	}
	return false
}

func (q *QueryAST) serve(g *GraphRead, w http.ResponseWriter, r *http.Request) {
	path, parameters, code := g.requestTarget(r)
	if code != "" {
		queryASTProblem(w, r, string(code), "")
		return
	}
	if path != "" {
		queryASTProblem(w, r, "resource-not-found", "")
		return
	}
	if q.validateAuthority == nil || q.execute == nil {
		queryASTProblem(w, r, "temporarily-unavailable", "")
		return
	}
	if err := q.validateAuthority(r.Context()); err != nil {
		queryASTFailure(w, r, err)
		return
	}
	if r.Method != http.MethodPost {
		queryASTResponse(w, r, http.StatusMethodNotAllowed, nil, "", http.Header{"Allow": {"POST"}})
		return
	}
	operation, problem := queryASTParameters(parameters)
	if problem != "" {
		queryASTProblem(w, r, problem, "")
		return
	}
	for _, header := range []string{"If-Match", "If-None-Match", "If-Modified-Since", "If-Unmodified-Since", "If-Range", "Range"} {
		if _, exists := graphReadField(r.Header, header); exists {
			queryASTProblem(w, r, "invalid-parameter", "")
			return
		}
	}
	media, present := graphReadField(r.Header, "Content-Type")
	parsed, params, err := mime.ParseMediaType(media)
	if !present || err != nil || parsed != queryASTMedia || len(params) != 0 || len(r.Header.Values("Content-Type")) != 1 {
		queryASTResponse(w, r, http.StatusUnsupportedMediaType, nil, "", nil)
		return
	}
	if encoding, exists := graphReadField(r.Header, "Content-Encoding"); exists && encoding != "identity" {
		queryASTResponse(w, r, http.StatusUnsupportedMediaType, nil, "", nil)
		return
	}
	// Refuse an unacceptable selected representation before reading/parsing its
	// request or executing a provider query. Authority and transport admission
	// above still retain their established precedence.
	if accept, exists := graphReadField(r.Header, "Accept"); exists && !queryASTAccepts(accept) {
		queryASTResponse(w, r, http.StatusNotAcceptable, nil, "", nil)
		return
	}
	if r.ContentLength > queryASTRequestLimit {
		queryASTProblem(w, r, "request-too-large", "")
		return
	}
	if r.Body == nil {
		queryASTProblem(w, r, "malformed-request", "")
		return
	}
	raw, err := io.ReadAll(io.LimitReader(r.Body, queryASTRequestLimit+1))
	if err != nil {
		if contextErr := r.Context().Err(); contextErr != nil {
			queryASTFailure(w, r, contextErr)
		} else {
			queryASTProblem(w, r, "malformed-request", "")
		}
		return
	}
	if len(raw) > queryASTRequestLimit {
		queryASTProblem(w, r, "request-too-large", "")
		return
	}
	ast, err := decodeQueryASTBody(raw, operation)
	if err != nil {
		queryASTFailure(w, r, err)
		return
	}
	if err = r.Context().Err(); err != nil {
		queryASTFailure(w, r, err)
		return
	}
	response, err := q.execute(r.Context(), ast)
	if err != nil {
		queryASTFailure(w, r, err)
		return
	}
	if err = r.Context().Err(); err != nil {
		queryASTFailure(w, r, err)
		return
	}
	body, err := json.Marshal(response)
	if err != nil {
		queryASTResponse(w, r, http.StatusInternalServerError, nil, "", nil)
		return
	}
	if len(body) > graphReadRepresentationLimit {
		queryASTProblem(w, r, "limit-exceeded", "")
		return
	}
	headers := http.Header{"Link": {"<" + g.scope + "bdp.json>; rel=\"service-desc\"; type=\"application/json\""}}
	if resources, ok := response.Result.(resourcequery.ResourcesResult); ok {
		headers.Set("BDP-Collection-Kind", string(resources.ResourceKind))
	}
	queryASTResponse(w, r, http.StatusOK, body, queryASTMedia, headers)
}

func queryASTParameters(parameters url.Values) (string, string) {
	for key := range parameters {
		if key != "readVersion" && key != "operation" {
			return "", "invalid-parameter"
		}
	}
	versions, operations := parameters["readVersion"], parameters["operation"]
	if len(versions) != 1 || versions[0] == "" || len(operations) != 1 || operations[0] == "" {
		return "", "invalid-parameter"
	}
	for _, c := range versions[0] {
		if c < '0' || c > '9' {
			return "", "invalid-parameter"
		}
	}
	if versions[0] != "1" {
		return "", "capability-not-supported"
	}
	if operations[0] != "list" && operations[0] != "query" {
		return "", "invalid-parameter"
	}
	return operations[0], ""
}

type queryASTInputFailure struct{ code, at string }

func (e *queryASTInputFailure) Error() string { return e.code }

// Decode the complete closed native request and admit its AST before storage.
// Valid but unavailable sort/page-size controls refuse instead of being ignored.
func decodeQueryASTBody(raw []byte, operation string) ([]byte, error) {
	canonical, err := graph.CanonicalizeJSON(raw)
	if err != nil {
		return nil, &queryASTInputFailure{code: "malformed-request"}
	}
	var members map[string]json.RawMessage
	if err = json.Unmarshal(canonical, &members); err != nil || members == nil {
		return nil, &queryASTInputFailure{code: "invalid-parameter"}
	}
	for key := range members {
		if key != "query" && key != "sort" && key != "limit" {
			return nil, &queryASTInputFailure{code: "invalid-parameter"}
		}
	}
	ast, exists := members["query"]
	if !exists {
		return nil, &queryASTInputFailure{code: "invalid-parameter", at: "/query"}
	}
	request, err := resourcequery.Decode(ast)
	if err != nil {
		return nil, err
	}
	if err = resourcequery.Admit(request, resourcequery.DefaultLimits()); err != nil {
		return nil, err
	}
	if operation == "list" && request.Profile() != "list" || operation == "query" && request.Profile() == "list" {
		return nil, &queryASTInputFailure{code: "invalid-parameter", at: "/query/profile"}
	}
	unsupportedAt := ""
	if sort, exists := members["sort"]; exists {
		var fields map[string]json.RawMessage
		if err = json.Unmarshal(sort, &fields); err != nil || len(fields) != 2 {
			return nil, &queryASTInputFailure{code: "invalid-parameter", at: "/sort"}
		}
		var path, direction string
		if string(fields["path"]) == "null" || string(fields["direction"]) == "null" || json.Unmarshal(fields["path"], &path) != nil || json.Unmarshal(fields["direction"], &direction) != nil || !queryASTPointer(path) || direction != "asc" && direction != "desc" || request.ResultForm() == "graph" {
			return nil, &queryASTInputFailure{code: "invalid-parameter", at: "/sort"}
		}
		unsupportedAt = "/sort"
	}
	if limit, exists := members["limit"]; exists {
		var value int
		if json.Unmarshal(limit, &value) != nil || value < 1 || request.ResultForm() != "resources" {
			return nil, &queryASTInputFailure{code: "invalid-parameter", at: "/limit"}
		}
		unsupportedAt = "/limit"
	}
	if unsupportedAt != "" {
		return nil, &queryASTInputFailure{code: "capability-not-supported", at: unsupportedAt}
	}
	return ast, nil
}

func queryASTPointer(path string) bool {
	if path != "" && !strings.HasPrefix(path, "/") {
		return false
	}
	for i := 0; i < len(path); i++ {
		if path[i] == '~' {
			i++
			if i >= len(path) || path[i] != '0' && path[i] != '1' {
				return false
			}
		}
	}
	return true
}

// A selected-source Problem extension, kept separate from the older pinned
// bdpwire table. No internal evaluator message, identity or storage detail leaks.
func queryASTFailure(w http.ResponseWriter, r *http.Request, err error) {
	var input *queryASTInputFailure
	var failure *resourcequery.Failure
	switch {
	case errors.As(err, &input):
		queryASTProblem(w, r, input.code, input.at)
	case errors.Is(err, resourcequery.ErrUnavailable):
		queryASTProblem(w, r, "resource-not-found", "")
	case errors.Is(err, resourcequery.ErrUnsupported):
		at := ""
		if errors.As(err, &failure) {
			at = failure.At
		}
		queryASTProblem(w, r, "capability-not-supported", at)
	case errors.Is(err, resourcequery.ErrInvalid):
		at := ""
		if errors.As(err, &failure) {
			at = failure.At
		}
		queryASTProblem(w, r, "invalid-parameter", at)
	case errors.Is(err, resourcequery.ErrLimit):
		queryASTProblem(w, r, "limit-exceeded", "")
	case errors.Is(err, resourcequery.ErrDataset):
		// A deterministic invalid checked dataset is an implementation fault,
		// not a temporary provider condition or a new public Problem code.
		queryASTResponse(w, r, http.StatusInternalServerError, nil, "", nil)
	default:
		code := graphReadFailureCode(err)
		if code == "" {
			queryASTResponse(w, r, http.StatusInternalServerError, nil, "", nil)
			return
		}
		queryASTProblem(w, r, string(code), "")
	}
}

func queryASTProblem(w http.ResponseWriter, r *http.Request, code, at string) {
	status, family, retry := 0, "", ""
	if code == "capability-not-supported" {
		status, family, retry = 422, "request", "after-state-change"
	} else {
		native := bdpwire.ReadProblemCode(code)
		status, family, retry = native.Status(), string(native.Family()), string(native.Retry())
	}
	if status == 0 {
		queryASTResponse(w, r, http.StatusInternalServerError, nil, "", nil)
		return
	}
	problem := map[string]any{"type": bdpwire.ProblemTypePrefix + family, "status": status, "code": code, "retry": retry}
	if code == "capability-not-supported" {
		problem["capability"] = queryASTCapability
		problem["capabilityVersion"] = "1"
	}
	if at != "" && code != "resource-not-found" {
		problem["at"] = at
	}
	body, err := json.Marshal(problem)
	if err != nil {
		queryASTResponse(w, r, http.StatusInternalServerError, nil, "", nil)
		return
	}
	queryASTResponse(w, r, status, body, "application/problem+json", nil)
}

func queryASTResponse(w http.ResponseWriter, r *http.Request, status int, body []byte, media string, headers http.Header) {
	for key, values := range headers {
		w.Header()[key] = append([]string(nil), values...)
	}
	w.Header().Set("Cache-Control", "private, no-store")
	graphReadVaryAccept(w.Header())
	w.Header().Del("ETag")
	w.Header().Del("Content-Type")
	if media != "" {
		w.Header().Set("Content-Type", media)
	}
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.WriteHeader(status)
	if r.Method != http.MethodHead && len(body) > 0 {
		//nolint:gosec // G705: fixed JSON/Problem media and serialized JSON, never HTML.
		_, _ = w.Write(body)
	}
}

// Keep Accept quality/specificity rules identical to Read while substituting the
// one selected media type. application/json alone does not admit this version.
func queryASTAccepts(value string) bool {
	entries, ok := graphReadSplitQuoted(value, ',')
	if !ok {
		return false
	}
	best, quality := -1, float64(0)
	for _, entry := range entries {
		parts, valid := graphReadSplitQuoted(entry, ';')
		if !valid || len(parts) == 0 {
			continue
		}
		name := strings.ToLower(strings.TrimSpace(parts[0]))
		if name == "application/json" {
			continue
		}
		if name == queryASTMedia {
			parts[0] = "application/json"
		} else if name != "application/*" && name != "*/*" {
			continue
		}
		specificity, q, valid := graphReadJSONRange(strings.Join(parts, ";"))
		if !valid {
			continue
		}
		if specificity > best {
			best, quality = specificity, q
		} else if specificity == best && q > quality {
			quality = q
		}
	}
	return quality > 0
}
