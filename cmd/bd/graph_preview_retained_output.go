package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/clioutput"
	"github.com/steveyegge/beads/internal/httpapi/bdpwire"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func init() {
	for _, cmd := range []*cobra.Command{versionsCmd, historyCmd, linkVersionsCmd, graphCompareCmd, linkCompareCmd} {
		registerCanonicalOutputFlags(cmd)
		legacyArgs := cmd.Args
		cmd.Args = func(cmd *cobra.Command, args []string) error {
			if canonicalOutputSelected(cmd) {
				return nil // The result adapter emits structured argument failures.
			}
			return legacyArgs(cmd, args)
		}
	}
	for _, cmd := range []*cobra.Command{versionsCmd, historyCmd, linkVersionsCmd} {
		cmd.Flags().Lookup("output-format").Usage = "Canonical output: default, table, detail or json (cannot combine with legacy --json or --format)"
		canonicalFailureFactories[cmd] = func(options clioutput.Options) (canonicalFailureRenderer, error) {
			renderer, err := newGraphVersionOutput(bdpwire.Pin, options)
			if err != nil {
				return nil, err
			}
			return renderer, nil
		}
	}
	for _, cmd := range []*cobra.Command{graphCompareCmd, linkCompareCmd} {
		canonicalFailureFactories[cmd] = func(options clioutput.Options) (canonicalFailureRenderer, error) {
			renderer, err := newGraphComparisonOutput(options)
			if err != nil {
				return nil, err
			}
			return renderer, nil
		}
	}
}

// graphVersionItemDescriptor describes the actual CLI version summary. These
// rows are neither Resource snapshots nor BDP HistoryVersionRows. In particular,
// version is the citable token and local_revision is only a store-local ordinal.
// The containing collection retains resource and kind outside this projection.
func graphVersionItemDescriptor(bdpSpec string) clioutput.Descriptor {
	return clioutput.Descriptor{
		BDPSpec:    bdpSpec,
		Protection: &clioutput.Protection{Required: []string{"version"}},
	}
}

func newGraphVersionOutput(bdpSpec string, options clioutput.Options) (*clioutput.CollectionRenderer, error) {
	return clioutput.NewCollection(clioutput.CollectionDescriptor{
		Item: graphVersionItemDescriptor(bdpSpec), ResultKind: "version_collection",
		ItemsMember: "versions", ContextMembers: []string{"resource", "kind"},
		DefaultColumns: []string{"/version", "/local_revision", "/change_at", "/actor", "/attribution"},
	}, options)
}

// graphComparisonDescriptor describes the local comparison result. Its from/to
// members are references, not complete Resource operands. Keep differences,
// direction, supported areas and coverage intact; masking changes could turn a
// partial display into the false assertion that two revisions are equal.
func graphComparisonDescriptor(bdpSpec string) clioutput.Descriptor {
	return clioutput.Descriptor{
		BDPSpec: bdpSpec, ResultKind: "comparison",
		Protection: &clioutput.Protection{Required: []string{
			"resource", "from", "to", "compared", "unsupported", "changes",
		}},
	}
}

// graphComparisonOptions admits envelope removal but no data masks. The current
// provider has no declared Resource projection roots. Refusing before acquisition
// is safer than interpreting /properties against the difference document or
// fetching complete snapshots that the operation did not return.
func graphComparisonOptions(options clioutput.Options) (clioutput.Options, error) {
	if len(options.Include) != 0 {
		return clioutput.Options{}, fmt.Errorf("comparison output does not expose Resource projection roots; --include is unsupported")
	}
	for _, path := range options.Omit {
		if path != "envelope" {
			return clioutput.Options{}, fmt.Errorf("comparison output does not expose Resource projection roots; only --omit envelope is supported")
		}
	}
	return options, nil
}

func runCanonicalGraphVersions(cmd *cobra.Command, args []string, options clioutput.Options) error {
	renderer, err := newGraphVersionOutput(bdpwire.Pin, options)
	if err != nil {
		return canonicalFailure(cmd, nil, "invalid_selector", err.Error(), 2)
	}
	if len(args) != 1 {
		return canonicalFailure(cmd, renderer, "invalid_selector", "canonical versions requires exactly one Resource identity", 2)
	}
	if err := canonicalGraphFlags(cmd, renderer); err != nil {
		return err
	}
	path, err := graphPreviewResourcePath(graphPreviewConfig.GraphScopeURL, args[0])
	if cmd.Parent() != nil && cmd.Parent().Name() == "link" {
		path, err = graphPreviewLinkPath(graphPreviewConfig.GraphScopeURL, args[0])
	}
	if err != nil {
		return canonicalFailure(cmd, renderer, "invalid_selector", err.Error(), 2)
	}
	var count int
	return canonicalReadWithGraphStore(cmd, renderer, func(ctx context.Context, store *graphstore.Store) (any, error) {
		kind, rows, err := store.Versions(ctx, path)
		if errors.Is(err, graphstore.ErrNotFound) {
			return nil, graphVersionsNotFound(args[0])
		}
		if err != nil {
			return nil, err
		}
		citable, _, err := graphCitableVersionRows(kind, rows)
		if err != nil {
			return nil, err
		}
		count = len(citable)
		return map[string]any{"resource": path, "kind": string(kind), "versions": graphVersionRowsJSON(citable)}, nil
	}, func(result any) error {
		raw, err := graphProjectCompleteRecords(result)
		if err == nil {
			// Versions is a bounded complete-result provider. It refuses above
			// its acquisition limit and does not expose a continuation cursor.
			exhausted := true
			raw, err = renderer.Render(raw, clioutput.Completion{Returned: count, Exhausted: &exhausted, StopReason: "exhausted"})
		}
		if err != nil {
			return canonicalFailure(cmd, renderer, "output_error", err.Error(), 1)
		}
		if err := writeCanonicalDocument(cmd.OutOrStdout(), raw); err != nil {
			// A write may have emitted a prefix. Do not append another document.
			return canonicalFailure(cmd, nil, "output_error", err.Error(), 1)
		}
		return nil
	})
}

func newGraphComparisonOutput(options clioutput.Options) (*clioutput.Renderer, error) {
	options, err := graphComparisonOptions(options)
	if err != nil {
		return nil, err
	}
	return clioutput.New(graphComparisonDescriptor(bdpwire.Pin), options)
}
