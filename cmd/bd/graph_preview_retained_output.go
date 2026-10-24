package main

import (
	"context"
	"encoding/json"
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
			if legacyArgs != nil {
				return legacyArgs(cmd, args)
			}
			return nil
		}
	}
	for _, cmd := range []*cobra.Command{versionsCmd, historyCmd, linkVersionsCmd} {
		cmd.Flags().Lookup("output-format").Usage = "Canonical output: default, table, detail, json or jsonl (cannot combine with legacy --json or --format)"
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
	format := ""
	if len(options.Formats) != 0 {
		format = options.Formats[0]
	}
	item := graphVersionItemDescriptor(bdpSpec)
	standalone := format == "jsonl"
	if standalone {
		item.Protection.Required = []string{"resource", "kind", "version"}
	}
	return clioutput.NewCollection(clioutput.CollectionDescriptor{
		Item: item, ResultKind: "version_collection",
		ItemsMember: "versions", ContextMembers: []string{"resource", "kind"},
		StandaloneJSONLItems: standalone,
		DefaultColumns:       []string{"/version", "/local_revision", "/change_at", "/actor", "/attribution"},
	}, options)
}

// A JSONL record is independently interpretable even after --omit envelope.
// The ordinary JSON collection keeps its established {resource,kind,versions}
// structure; only the stream's delivery unit repeats the subject context.
func graphVersionStreamItem(resource, kind string, row map[string]any) (json.RawMessage, error) {
	item := make(map[string]any, len(row)+2)
	for key, value := range row {
		item[key] = value
	}
	item["resource"], item["kind"] = resource, kind
	return json.Marshal(item)
}

// Every native collection member outside versions is context for interpreting
// its rows. Until the stream has a declared way to carry that member in each
// raw item, refuse the stream rather than silently losing it.
func graphVersionJSONLContextSupported(result map[string]any) error {
	for member := range result {
		switch member {
		case "resource", "kind", "versions":
		default:
			return fmt.Errorf("version JSONL cannot preserve native collection context %q", member)
		}
	}
	return nil
}

type graphVersionJSONLFailure struct{ stream *clioutput.JSONLWriter }

func (f graphVersionJSONLFailure) RenderFailure(origin string, problem json.RawMessage) ([]byte, error) {
	return nil, f.stream.Failure(origin, problem)
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
		return canonicalFailure(cmd, canonicalFailureRendererFor(cmd), "invalid_selector", err.Error(), 2)
	}
	if len(args) != 1 {
		return canonicalFailure(cmd, renderer, "invalid_selector", "canonical versions requires exactly one Resource identity", 2)
	}
	if err := canonicalGraphFlags(cmd, renderer); err != nil {
		return err
	}
	jsonl := len(options.Formats) != 0 && options.Formats[0] == "jsonl"
	path, err := graphPreviewResourcePath(graphPreviewConfig.GraphScopeURL, args[0])
	if cmd.Parent() != nil && cmd.Parent().Name() == "link" {
		path, err = graphPreviewLinkPath(graphPreviewConfig.GraphScopeURL, args[0])
	}
	if err != nil {
		return canonicalFailure(cmd, renderer, "invalid_selector", err.Error(), 2)
	}
	var stream *clioutput.JSONLWriter
	var failure canonicalFailureRenderer = renderer
	if jsonl {
		stream, err = renderer.NewJSONL(cmd.OutOrStdout())
		if err != nil {
			return canonicalFailure(cmd, nil, "output_error", err.Error(), 1)
		}
		failure = graphVersionJSONLFailure{stream}
	}
	var count int
	return canonicalReadWithGraphStore(cmd, failure, func(ctx context.Context, store *graphstore.Store) (any, error) {
		history, err := store.History(ctx, path)
		kind, rows := history.Kind, history.Versions
		if errors.Is(err, graphstore.ErrNotFound) {
			return nil, graphVersionsNotFound(args[0])
		}
		if err != nil {
			return nil, err
		}
		citable, terminals, _, err := graphCitableVersionRows(kind, rows)
		if err != nil {
			return nil, err
		}
		count = len(citable)
		result := map[string]any{"resource": path, "kind": string(kind), "versions": graphVersionRowsJSON(citable)}
		if history.CurrentState != nil {
			result["currentState"] = history.CurrentState
		}
		if (kind == graphstore.KindMemory || kind == graphstore.KindIssue || kind == graphstore.KindLink) && len(terminals) != 0 {
			result["terminal_versions"] = graphVersionRowsJSON(terminals)
		}
		return result, nil
	}, func(result any) error {
		if jsonl {
			data := result.(map[string]any)
			if err := graphVersionJSONLContextSupported(data); err != nil {
				return canonicalFailure(cmd, failure, "capability_unavailable", err.Error(), 5)
			}
			rows := data["versions"].([]map[string]any)
			for _, row := range rows {
				item, err := graphVersionStreamItem(path, data["kind"].(string), row)
				if err != nil {
					return canonicalFailure(cmd, failure, "output_error", err.Error(), 1)
				}
				if err := stream.Item(item); err != nil {
					return canonicalFailure(cmd, nil, "output_error", err.Error(), 1)
				}
			}
			if err := stream.Finish(true, "exhausted", nil); err != nil {
				return canonicalFailure(cmd, nil, "output_error", err.Error(), 1)
			}
			return nil
		}
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
