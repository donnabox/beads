package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/clioutput"
	"github.com/steveyegge/beads/internal/httpapi/graphread"
	"github.com/steveyegge/beads/internal/resourcequery"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// Canonical identity of the selected values. Exact source/schema revisions are
// qualification metadata, not alternate bdp_spec identities. This additive path
// does not upgrade the retained GET/Selector binding or advertise a Core service.
const queryASTContract = "https://github.com/gastownhall/bdp/schemas/bdp-v0.1.schema.json"

func init() {
	queryCmd.Flags().StringArray("ast", nil, "Execute one JSON Query AST in graph mode; use @FILE or - for stdin")
	registerCanonicalOutputFlags(queryCmd)
	canonicalFailureFactories[queryCmd] = func(options clioutput.Options) (canonicalFailureRenderer, error) {
		if !queryCmd.Flags().Changed("ast") {
			return clioutput.New(canonicalResourceDescriptor(), options)
		}
		return clioutput.New(queryASTDescriptor("resources"), options)
	}
}

func readQueryAST(cmd *cobra.Command, args []string) ([]byte, error) {
	values, err := cmd.Flags().GetStringArray("ast")
	if err != nil {
		return nil, err
	}
	if len(values) != 1 || values[0] == "" || len(args) != 0 {
		return nil, errors.New("use exactly one --ast JSON, --ast @FILE or --ast -, without a legacy expression")
	}
	var input io.Reader = strings.NewReader(values[0])
	var file *os.File
	if values[0] == "-" {
		input = cmd.InOrStdin()
	} else if strings.HasPrefix(values[0], "@") {
		if len(values[0]) == 1 {
			return nil, errors.New("--ast @ requires a filename")
		}
		file, err = os.Open(values[0][1:])
		if err != nil {
			return nil, fmt.Errorf("read Query AST: %w", err)
		}
		input = file
	}
	bound := resourcequery.DefaultLimits().RequestBytes
	raw, err := io.ReadAll(io.LimitReader(input, int64(bound)+1))
	if file != nil {
		err = errors.Join(err, file.Close())
	}
	if err != nil {
		return nil, err
	}
	if len(raw) > bound {
		return nil, errors.New("Query AST exceeds request byte budget")
	}
	return raw, nil
}

func queryASTOutputOptions(cmd *cobra.Command) (clioutput.Options, error) {
	formats, _ := cmd.Flags().GetStringArray("output-format")
	jsonRequested, _ := cmd.Flags().GetBool("json")
	if cmd.Root().PersistentFlags().Lookup("json") != nil {
		jsonRequested, _ = cmd.Root().PersistentFlags().GetBool("json")
	}
	if len(formats) == 0 {
		formats = []string{"json"}
	}
	omit, _ := cmd.Flags().GetStringArray("omit")
	include, _ := cmd.Flags().GetStringArray("include")
	if len(include) != 0 {
		return clioutput.Options{}, errors.New("Query field projection belongs in the AST, not --include")
	}
	for _, path := range omit {
		if path != "envelope" {
			return clioutput.Options{}, errors.New("Query field projection belongs in the AST; --omit accepts only envelope")
		}
	}
	return clioutput.Options{Formats: formats, JSON: jsonRequested, Omit: omit}, nil
}

func queryASTDescriptor(form string) clioutput.Descriptor {
	kind := "resource_collection"
	if form == "count" {
		kind = "aggregate"
	} else if form == "graph" {
		kind = "graph"
	}
	return clioutput.Descriptor{BDPSpec: queryASTContract, ResultKind: kind,
		Protection: &clioutput.Protection{Required: []string{"result", "evaluation"}}}
}

func queryASTRenderer(cmd *cobra.Command, form string) (*clioutput.Renderer, error) {
	options, err := queryASTOutputOptions(cmd)
	if err != nil {
		return nil, err
	}
	return clioutput.New(queryASTDescriptor(form), options)
}

func runQueryAST(cmd *cobra.Command, args []string) error {
	renderer, err := queryASTRenderer(cmd, "resources")
	if err != nil {
		return canonicalFailure(cmd, nil, "invalid_selector", err.Error(), 2)
	}
	if !graphPreviewActive {
		return canonicalFailure(cmd, renderer, "capability_unavailable", "JSON Query AST execution requires an existing graph-mode link workspace", 5)
	}
	if err := canonicalGraphFlags(cmd, renderer, "ast", "json"); err != nil {
		return err
	}
	raw, err := readQueryAST(cmd, args)
	if err != nil {
		return canonicalFailure(cmd, renderer, "invalid_selector", err.Error(), 2)
	}
	request, err := resourcequery.Decode(raw)
	if err == nil {
		err = resourcequery.Admit(request, resourcequery.DefaultLimits())
	}
	if err != nil {
		if errors.Is(err, resourcequery.ErrUnsupported) || errors.Is(err, resourcequery.ErrLimit) {
			return canonicalFailure(cmd, renderer, "capability_unavailable", err.Error(), 5)
		}
		return canonicalFailure(cmd, renderer, "invalid_selector", err.Error(), 2)
	}
	renderer, err = queryASTRenderer(cmd, request.ResultForm())
	if err != nil {
		return canonicalFailure(cmd, nil, "invalid_selector", err.Error(), 2)
	}
	return canonicalReadWithGraphStore(cmd, renderer, func(ctx context.Context, store *graphstore.Store) (any, error) {
		result, err := graphread.New(store).ExecuteResourceQuery(ctx, raw)
		switch {
		case errors.Is(err, resourcequery.ErrUnavailable):
			err = errors.Join(graphstore.ErrNotFound, err)
		case errors.Is(err, resourcequery.ErrUnsupported), errors.Is(err, resourcequery.ErrLimit):
			err = errors.Join(graphstore.ErrCapabilityUnavailable, err)
		case errors.Is(err, resourcequery.ErrDataset):
			err = errors.Join(graphstore.ErrInvalidStore, err)
		}
		return result, err
	}, func(value any) error {
		return renderCanonicalResult(cmd, renderer, value)
	})
}
