package httpapi

import (
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/steveyegge/beads/internal/httpapi/apigen"
)

// Protocol is which surfaces a server answers on its one address.
//
//   - ProtocolV0 is v0 alone, exactly as this server answered before BDP existed: no
//     /bdp, no Scope root, and every path outside the route table is v0's own 404.
//   - ProtocolBDP is BDP alone: Core, Write for the Issue and Dependency Types, and the
//     Callables, which still run on their v0 rows but are no longer reachable as v0 routes.
//   - ProtocolBoth is both.
//
// The zero value is v0, so a Config that never heard of this field serves what it
// always served.
type Protocol string

const (
	ProtocolV0   Protocol = "v0"
	ProtocolBDP  Protocol = "bdp"
	ProtocolBoth Protocol = "both"
)

// ParseProtocol reads `bd serve --protocol`.
func ParseProtocol(name string) (Protocol, error) {
	switch p := Protocol(name); p {
	case ProtocolV0, ProtocolBDP, ProtocolBoth:
		return p, nil
	}
	return "", fmt.Errorf("--protocol must be v0, bdp or both, got %q", name)
}

func (p Protocol) servesV0() bool { return p != ProtocolBDP }

func (p Protocol) servesBDP() bool { return p == ProtocolBDP || p == ProtocolBoth }

// v0PathPrefix is what ProtocolBDP does not route. /healthz is outside it: liveness
// is not an operation on a Scope and belongs to neither surface, so an orchestrator's
// probe keeps working whichever one the operator picked.
const v0PathPrefix = "/v0/"

func isV0Route(rt route) bool { return strings.HasPrefix(rt.pattern, v0PathPrefix) }

// codeProtocolNotServed answers a /v0 path on a server with v0 off. It is
// deliberately not v0's not_found: a client maps that to "no such issue", so a
// workspace left on v0 after its server moved would be told an issue it can see in
// any other client does not exist. A client takes a code it does not know as a
// refusal and prints the detail, which is what this one needs and every release of
// the client already does. It is not in the document's vocabulary because a server
// answering it does not serve the document.
const codeProtocolNotServed = "protocol_not_served"

// serveProtocolNotServed answers every /v0 path under ProtocolBDP. Like an unrouted
// path it asks for no credential: it says nothing about the workspace.
func (s *Server) serveProtocolNotServed(w http.ResponseWriter, r *http.Request) {
	detail := "this bd serve runs with --protocol bdp and does not serve /v0: " +
		"connect with `bd connect --protocol bdp <url>`, or run the server with `bd serve --protocol both`"
	if r.URL.Path == v0PathPrefix+"beads/events:watch" {
		// The one thing a client on BDP still wants from /v0: a stream is not an
		// invocation, so there is no Callable to point at.
		detail = "this bd serve runs with --protocol bdp and does not serve /v0, and the event stream has no Callable: " +
			"run the server with `bd serve --protocol both`"
	}
	s.fail(w, r, Result{Problem: apigen.Problem{
		Status: http.StatusNotFound,
		Title:  http.StatusText(http.StatusNotFound),
		Code:   codeProtocolNotServed,
		Detail: &detail,
	}})
}

// capabilitiesFor drops what p takes off the wire. Every operation but the event
// stream is still served under ProtocolBDP, as a Callable, so its token stays true;
// the stream's would send a client to a route that is not there.
func capabilitiesFor(p Protocol, capabilities []string) []string {
	stream, ok := routeForOp(OpWatchEvents)
	if p.servesV0() || !ok {
		return capabilities
	}
	return slices.DeleteFunc(capabilities, func(token string) bool { return token == stream.capability })
}
