package wire

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/steveyegge/beads/internal/httpapi/typeref"
	"github.com/steveyegge/beads/internal/utils"
)

// The Types rule of specs/types/issue/callables.md §2, the third thing this layer
// cancels out: a Callable carries a Bead's Type or a dependency's kind as `type`,
// valued with a Type reference, where everything above this file speaks v0's
// `issue_type` and short names. Where the rule reaches, and the shared identities,
// are typeref's, which bd serve reads too. So the layers above send and receive
// exactly what they did over v0.

// aliasingTypeFilters are the operations whose v0 `type` filter expands the native
// shorthand aliases (`feat`, `mr`, ...). A Type reference has no aliases, so the
// expansion v0 did is done here, and only where v0 did it: countIssues' `type` is
// exact. Every `exclude_type` is expanded, as v0 expands each one.
var aliasingTypeFilters = map[string]bool{
	OpListReadyWork: true, OpCountReadyWork: true, OpClaimNextIssue: true, OpListIssues: true,
}

// scopeFunc answers the Scope URL a Scope-local Type is spelled under. It is asked
// only for a name outside Issue §2's shared tables, which is rare, so the common
// request costs no extra round trip.
type scopeFunc func() (string, error)

// bdpScope is the Scope URL bd serve names in its Core discovery document, read once.
// It is the server's, not this client's base URL: a proxy in front of bd serve can
// change the Host the server sees, and with it the Scope it spells and accepts a
// Scope-local Type under. Nothing is guessed when it cannot be read.
func (c *Client) bdpScope(ctx context.Context) (string, error) {
	c.scopeMu.Lock()
	defer c.scopeMu.Unlock()
	if c.scope != "" {
		return c.scope, nil
	}
	r := Request{Op: opCoreDiscovery, Method: http.MethodGet, Path: coreDiscoveryPath}
	send := &outbound{method: http.MethodGet, url: c.base.JoinPath(coreDiscoveryPath)}
	req, err := c.newRequest(ctx, r, send)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", callableMediaType)
	res, err := c.roundTripOnce(ctx, r, send, req, c.headerTimeout, time.Time{})
	if err != nil {
		return "", fmt.Errorf("reading bd serve's Scope to spell a Scope-local Type: %w", err)
	}
	var discovery struct {
		Scope string `json:"scope"`
	}
	if res.status != http.StatusOK || json.Unmarshal(res.body, &discovery) != nil {
		return "", fmt.Errorf("reading bd serve's Scope to spell a Scope-local Type: %s answered %d", coreDiscoveryPath, res.status)
	}
	scope, err := url.Parse(discovery.Scope)
	if err != nil || (scope.Scheme != "http" && scope.Scheme != "https") || scope.Host == "" || !strings.HasSuffix(scope.Path, "/") {
		return "", fmt.Errorf("bd serve's discovery names no usable Scope: %q", discovery.Scope)
	}
	c.scope = discovery.Scope
	return c.scope, nil
}

const (
	opCoreDiscovery   = "coreDiscovery"
	coreDiscoveryPath = "/bdp"
)

// typeArguments spells the arguments' Types as the Callable declares them. An empty
// name stays empty, so the server refuses it as v0 refused it.
func typeArguments(op string, scope scopeFunc, arguments map[string]json.RawMessage) error {
	plans, err := typeref.Load()
	if err != nil {
		return err
	}
	shape := plans.Arguments(op, typeref.ToCallable)
	if shape == nil {
		return nil
	}
	if raw, ok := arguments["exclude_type"]; ok && typeref.ParameterKind(op, "exclude_type") == typeref.Bead {
		split, err := splitTypeList(raw)
		if err != nil {
			return fmt.Errorf("%s: exclude_type: %w", op, err)
		}
		arguments["exclude_type"] = split
	}
	_, err = shape.ApplyMembers(arguments, func(kind typeref.Kind, value, at string) (string, error) {
		if value == "" {
			return "", nil
		}
		namespace, local, table := typeref.BeadNamespace, typeref.LocalBeadTypes, typeref.BeadTypeNames
		if kind == typeref.Link {
			namespace, local, table = typeref.LinkNamespace, typeref.LocalLinkTypes, typeref.LinkTypeNames
		} else if at == "/exclude_type" || strings.HasPrefix(at, "/exclude_type/") || (at == "/type" && aliasingTypeFilters[op]) {
			value = utils.NormalizeIssueType(value)
		}
		if table[value] {
			return namespace + value, nil
		}
		prefix, err := scope()
		if err != nil {
			return "", err
		}
		return prefix + local + url.PathEscape(value), nil
	})
	if err != nil {
		return fmt.Errorf("%s: spelling the arguments' Types: %w", op, err)
	}
	return nil
}

// splitTypeList is v0's reading of `exclude_type`: repeated, and each value may be a
// comma-separated list, trimmed, with empty entries dropped. A reference is never
// split, so the list is split before it is spelled.
func splitTypeList(raw json.RawMessage) (json.RawMessage, error) {
	var values []string
	var one string
	if err := json.Unmarshal(raw, &one); err == nil {
		values = []string{one}
	} else if err := json.Unmarshal(raw, &values); err != nil {
		return nil, fmt.Errorf("a list of type names is a string or an array of strings")
	}
	split := []string{}
	for _, value := range values {
		for _, name := range strings.Split(value, ",") {
			if name = strings.TrimSpace(name); name != "" {
				split = append(split, name)
			}
		}
	}
	return json.Marshal(split)
}

// typeNames reads bd serve's Type references back to v0's names: a result when
// result is set, and otherwise a Problem body. A reference this client cannot spell
// as a name is an error rather than a guess.
func typeNames(op string, scope scopeFunc, body []byte, result bool) ([]byte, error) {
	plans, err := typeref.Load()
	if err != nil {
		return nil, err
	}
	shape := plans.Problem(typeref.ToV0)
	if result {
		shape = plans.Result(op, typeref.ToV0)
	}
	out, _, err := shape.Apply(body, func(kind typeref.Kind, value, _ string) (string, error) {
		namespace, local, table := typeref.BeadNamespace, typeref.LocalBeadTypes, typeref.BeadTypeNames
		if kind == typeref.Link {
			namespace, local, table = typeref.LinkNamespace, typeref.LocalLinkTypes, typeref.LinkTypeNames
		}
		if name, ok := strings.CutPrefix(value, namespace); ok && table[name] {
			return name, nil
		}
		prefix, err := scope()
		if err != nil {
			return "", err
		}
		if name, ok := strings.CutPrefix(value, prefix+local); ok {
			if plain, err := url.PathUnescape(name); err == nil && plain != "" {
				return plain, nil
			}
		}
		return "", fmt.Errorf("bd serve answered the %s Type %q, which names no type this client knows", kind, value)
	})
	return out, err
}
