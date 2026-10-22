package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/cobra"

	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/clioutput"
	"github.com/steveyegge/beads/internal/httpapi/bdpwire"
	"github.com/steveyegge/beads/internal/httpapi/graphread"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// This is the new CLI graph adapter's compound result, not a BDP operation or
// public query representation. Membership/context remains outside Resource masks.
type memoryGraphResult struct {
	Scope     string               `json:"scope"`
	Beads     []bdpwire.BeadRecord `json:"beads"`
	Links     []bdpwire.LinkRecord `json:"links"`
	Traversal memoryGraphTraversal `json:"traversal"`
}
type memoryGraphTraversal struct {
	Operation string            `json:"operation"`
	Current   bool              `json:"current"`
	Direction string            `json:"direction"`
	MaxNodes  int               `json:"maxNodes"`
	MaxLinks  int               `json:"maxLinks"`
	Complete  bool              `json:"complete"`
	Roots     []memoryGraphRoot `json:"roots"`
}
type memoryGraphRoot struct {
	ID       string   `json:"id"`
	Revision string   `json:"revision"`
	Depth    int      `json:"depth"`
	Beads    []string `json:"beads"`
	Links    []string `json:"links"`
	Frontier []string `json:"frontier"`
}

func memoryGraphDescriptor() clioutput.Descriptor {
	return clioutput.Descriptor{BDPSpec: bdpwire.Pin, ResultKind: "graph",
		Protection: &clioutput.Protection{Required: []string{"scope", "traversal"}},
	}
}

func memoryCurrentResult(read graphstore.CurrentFamilyRead, scope string, op memoryReadOperation, req graphstore.CurrentFamilyRequest) (memoryGraphResult, error) {
	zero := memoryGraphResult{}
	if err := graph.ValidateScopeURL(scope); err != nil {
		return zero, err
	}
	if len(read.Graph.Views) != len(req.Roots) {
		return zero, fmt.Errorf("%w: missing root context", graphstore.ErrInvalidStore)
	}
	inventory, err := graphread.ProjectCurrentRecords(graphstore.Snapshot{WriterToken: read.WriterToken, Records: read.Records, Types: read.Types, OwnedLinkTypes: read.OwnedLinkTypes})
	if err != nil {
		return zero, err
	}
	result := memoryGraphResult{Scope: scope, Beads: inventory.Beads, Links: inventory.Links,
		Traversal: memoryGraphTraversal{Operation: op.name, Current: true, Direction: req.Direction.String(), MaxNodes: req.Limits.MaxNodes, MaxLinks: req.Limits.MaxLinks, Complete: true, Roots: []memoryGraphRoot{}},
	}
	versions := map[string]string{}
	identities := map[string]string{}
	remaining := graphstore.PreviewCurrentReadByteLimit
	charge := func(n int) error {
		if n > remaining {
			return fmt.Errorf("%w: Memory graph output budget", graphstore.ErrLimitExceeded)
		}
		remaining -= n
		return nil
	}
	// Charge canonical Resource JSON once and root-context references at every
	// appearance before encoding the compound result. This is a conservative
	// output budget, not a new serializer or a successful truncation policy.
	for _, bead := range inventory.Beads {
		raw, err := json.Marshal(bead)
		if err != nil {
			return zero, err
		}
		if err := charge(len(raw) + 2); err != nil {
			return zero, err
		}
	}
	for _, link := range inventory.Links {
		raw, err := json.Marshal(link)
		if err != nil {
			return zero, err
		}
		if err := charge(len(raw) + 2); err != nil {
			return zero, err
		}
	}
	if err := charge(512); err != nil {
		return zero, err
	}
	for _, n := range read.Graph.Nodes {
		versions[n.Path] = n.Version
	}
	canonical := func(ids []string) []string {
		out := make([]string, 0, len(ids))
		for _, id := range ids {
			value, ok := identities[id]
			if !ok {
				value = graph.CanonicalURL(scope, id)
				identities[id] = value
			}
			out = append(out, value)
		}
		return out
	}
	for i, view := range read.Graph.Views {
		if view.Root != req.Roots[i] || versions[view.Root.Path] == "" {
			return zero, fmt.Errorf("%w: mismatched root context", graphstore.ErrInvalidStore)
		}
		root := memoryGraphRoot{ID: graph.CanonicalURL(scope, view.Root.Path), Revision: versions[view.Root.Path], Depth: view.Root.Depth, Beads: canonical(view.Nodes), Links: canonical(view.Links), Frontier: canonical(view.Frontier)}
		if err := charge(6*len(root.ID) + 6*len(root.Revision) + 128); err != nil {
			return zero, err
		}
		for _, group := range [][]string{root.Beads, root.Links, root.Frontier} {
			for _, id := range group {
				if err := charge(6*len(id) + 3); err != nil {
					return zero, err
				}
			}
		}
		result.Traversal.Roots = append(result.Traversal.Roots, root)
		if len(root.Frontier) > 0 {
			result.Traversal.Complete = false
		}
	}
	return result, nil
}

// Reuse the common renderer/document writer. Buffer and bound the final framed
// bytes before publishing; a policy refusal never emits a success prefix.
func renderMemoryCanonicalResult(cmd *cobra.Command, renderer *clioutput.Renderer, result memoryGraphResult) error {
	raw, err := json.Marshal(result)
	if err != nil {
		return canonicalReadError(cmd, renderer, err, "output_error")
	}
	if len(raw) > graphstore.PreviewCurrentReadByteLimit {
		return canonicalReadError(cmd, renderer, graphstore.ErrLimitExceeded, "output_error")
	}
	output, err := renderer.Render(raw)
	if err != nil {
		if errors.Is(err, clioutput.ErrInvalidProjection) {
			return canonicalFailure(cmd, renderer, "invalid_selector", err.Error(), 2)
		}
		return canonicalReadError(cmd, renderer, err, "output_error")
	}
	if len(output) > graphstore.PreviewCurrentReadByteLimit {
		return canonicalReadError(cmd, renderer, graphstore.ErrLimitExceeded, "output_error")
	}
	if err := writeCanonicalDocument(cmd.OutOrStdout(), output); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "output_error: %v\n", err)
		return &exitError{Code: 1}
	}
	return nil
}

// Graph formats are adapter policy; the foundation still owns JSON framing,
// projection validation and failure documents. Default is this domain's tree.
func newMemoryGraphRenderer(options clioutput.Options) (*clioutput.Renderer, string, error) {
	format := ""
	for _, value := range options.Formats {
		if value == "default" {
			value = "tree"
		}
		if value != "tree" && value != "detail" && value != "json" {
			return nil, "", fmt.Errorf("unsupported Memory graph output format %q", value)
		}
		if format != "" && format != value {
			return nil, "", fmt.Errorf("conflicting output formats")
		}
		format = value
	}
	if options.JSON {
		if format != "" && format != "json" {
			return nil, "", fmt.Errorf("JSON shorthand conflicts with output format")
		}
		format = "json"
	}
	if format == "" {
		format = "tree"
	}
	base := options
	base.JSON = false
	base.Formats = []string{"detail"}
	if format == "json" {
		base.Formats = []string{"json"}
	}
	renderer, err := clioutput.New(memoryGraphDescriptor(), base)
	return renderer, format, err
}
