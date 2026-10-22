package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/clioutput"
	"github.com/steveyegge/beads/internal/memoryquery"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/typecatalog"
)

var memoryGraphCmd = &cobra.Command{
	Use:         "memory",
	Short:       "Explore Memory relationships and explicit obligations",
	Long:        "Read Memory relationships in a graph-mode link workspace. Existing remember, recall, memories and forget retain their own contracts.",
	Annotations: map[string]string{skipStoreAnnotation: "1", canonicalDefaultAnnotation: "1"},
}

type memoryReadOperation struct {
	name, family, description string
	direction                 graph.Direction
	depth                     int
}

var memoryReadOperations = []memoryReadOperation{
	{"references", "references", "Show referenced knowledge, including refining Link Types", graph.DirectionOut, 1},
	{"referenced-by", "references", "Show Beads that reference the input", graph.DirectionIn, 1},
	{"successor", "superseded-by", "Follow the replacement chain, retaining the starting Memory", graph.DirectionOut, 1000},
	{"supersedes", "superseded-by", "Show Memories directly replaced by the input", graph.DirectionIn, 1},
}

func init() {
	for _, op := range memoryReadOperations {
		memoryGraphCmd.AddCommand(newMemoryReadCommand(op))
	}
	rootCmd.AddCommand(memoryGraphCmd)
}

func newMemoryReadCommand(op memoryReadOperation) *cobra.Command {
	cmd := &cobra.Command{Use: op.name + " ID...", Short: op.description,
		Args: func(*cobra.Command, []string) error { return nil },
		RunE: func(cmd *cobra.Command, args []string) error { return runMemoryCurrentRead(cmd, args, op) },
	}
	cmd.Flags().Int("depth", op.depth, "Maximum relationship hops for each input root (0..1000)")
	cmd.Flags().Int("max-nodes", 100, "Maximum unique selected Beads across roots (1..1000); excess refuses")
	cmd.Flags().Int("max-links", 200, "Maximum unique selected Links across roots (1..1000); excess refuses")
	registerCanonicalOutputFlags(cmd)
	cmd.Flags().Lookup("output-format").Usage = "Output: default, tree, detail or json; --json is shorthand for json"
	canonicalFailureFactories[cmd] = func(options clioutput.Options) (canonicalFailureRenderer, error) {
		renderer, _, err := newMemoryGraphRenderer(options)
		return renderer, err
	}
	return cmd
}

func graphPreviewMemoryNamespaceCommand(cmd *cobra.Command) bool {
	for c := cmd; c != nil; c = c.Parent() {
		if c == memoryGraphCmd {
			return true
		}
	}
	return false
}

func memoryCurrentRequest(cmd *cobra.Command, args []string, scope string, op memoryReadOperation) (graphstore.CurrentFamilyRequest, error) {
	zero := graphstore.CurrentFamilyRequest{}
	if len(args) == 0 || len(args) > graphstore.PreviewSnapshotLimit {
		return zero, fmt.Errorf("one to %d Bead roots are required", graphstore.PreviewSnapshotLimit)
	}
	if raw := os.Getenv(maxRowsEnvVar); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 0 {
			return zero, fmt.Errorf("BEADS_MAX_ROWS must be a non-negative integer")
		}
		if value > 0 {
			return zero, fmt.Errorf("%w: Memory graph reads use explicit node/Link bounds; unset BEADS_MAX_ROWS", graphstore.ErrCapabilityUnavailable)
		}
	}
	depth, _ := cmd.Flags().GetInt("depth")
	nodes, _ := cmd.Flags().GetInt("max-nodes")
	links, _ := cmd.Flags().GetInt("max-links")
	if depth < 0 || depth > 1000 || nodes < 1 || nodes > 1000 || links < 1 || links > 1000 {
		return zero, fmt.Errorf("depth must be 0..1000 and node/Link bounds 1..1000")
	}
	request := graphstore.CurrentFamilyRequest{Family: typecatalog.Reference{URI: scope + "types/" + op.family},
		Direction: op.direction, Limits: memoryquery.Limits{MaxNodes: nodes, MaxLinks: links, MaxInspected: 100000},
		MaxBytes: graphstore.PreviewCurrentReadByteLimit,
	}
	for _, arg := range args {
		ref, err := graphPreviewParseReference(arg)
		if err != nil {
			return zero, err
		}
		if ref.Explicit && ref.Revision != "current" && ref.Revision != "unpinned" {
			return zero, fmt.Errorf("%w: Memory relationship reads select current roots only", graphstore.ErrCapabilityUnavailable)
		}
		path, err := graphPreviewResourcePath(scope, ref.Identity)
		if err != nil {
			return zero, err
		}
		if err := graph.ValidateBeadPath(path); err != nil {
			return zero, err
		}
		request.Roots = append(request.Roots, memoryquery.CurrentRoot{Path: path, Depth: depth})
	}
	return request, nil
}

func runMemoryCurrentRead(cmd *cobra.Command, args []string, op memoryReadOperation) error {
	options, _, err := canonicalOutputOptions(cmd)
	if err != nil {
		return canonicalFailure(cmd, nil, "invalid_selector", err.Error(), 2)
	}
	renderer, format, err := newMemoryGraphRenderer(options)
	if err != nil {
		return canonicalFailure(cmd, canonicalFailureRendererFor(cmd), "invalid_selector", err.Error(), 2)
	}
	if !graphPreviewActive {
		return canonicalFailure(cmd, renderer, "capability_unavailable", "Memory relationships require a graph-mode link workspace", 5)
	}
	if err := canonicalGraphFlags(cmd, renderer, "depth", "max-nodes", "max-links", "json"); err != nil {
		return err
	}
	if err := canonicalScalarOccurrences(cmd, os.Args[1:], "depth", "max-nodes", "max-links", "json"); err != nil {
		return canonicalFailure(cmd, renderer, "invalid_selector", err.Error(), 2)
	}
	// A graph is one compound item. Full Resource masks need a graph-specific
	// common adapter; until then refuse them rather than dropping membership.
	if len(options.Include) != 0 {
		return canonicalFailure(cmd, renderer, "capability_unavailable", "Memory graph field masks are not available yet", 5)
	}
	for _, omit := range options.Omit {
		if omit != "envelope" {
			return canonicalFailure(cmd, renderer, "capability_unavailable", "Memory graph field masks are not available yet", 5)
		}
	}
	request, err := memoryCurrentRequest(cmd, args, graphPreviewConfig.GraphScopeURL, op)
	if err != nil {
		if errors.Is(err, graphstore.ErrCapabilityUnavailable) {
			return canonicalReadError(cmd, renderer, err, "invalid_selector")
		}
		return canonicalFailure(cmd, renderer, "invalid_selector", err.Error(), 2)
	}
	previous := graphFailureReporter
	graphFailureReporter = func(code, message string, status int) error {
		return canonicalFailure(cmd, renderer, code, message, status)
	}
	defer func() { graphFailureReporter = previous }()
	return withGraphStoreStageErrorMappers(30*time.Second, func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		result, err := store.ReadCurrentFamily(ctx, request)
		return result, "", err
	}, func(value any, _ string) error {
		read, ok := value.(graphstore.CurrentFamilyRead)
		if !ok {
			return canonicalReadError(cmd, renderer, graphstore.ErrInvalidStore, "invalid_store")
		}
		result, err := memoryCurrentResult(read, graphPreviewConfig.GraphScopeURL, op, request)
		if err != nil {
			return canonicalReadError(cmd, renderer, err, "invalid_store")
		}
		if format == "json" || format == "detail" {
			return renderMemoryCanonicalResult(cmd, renderer, result)
		}
		// Bind presentation to the exact direction used for acquisition. The
		// private capture is not an independently reusable query description.
		text, err := renderMemoryCurrentGraph(read, graphPreviewConfig.GraphScopeURL, request.Direction, graphstore.PreviewCurrentReadByteLimit)
		if err != nil {
			return canonicalReadError(cmd, renderer, err, "output_error")
		}
		if quietFlag {
			return nil
		}
		if err := writeCanonicalDocument(cmd.OutOrStdout(), []byte(text)); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "output_error: %v\n", err)
			return &exitError{Code: 1}
		}
		return nil
	}, func(err error) error { return graphStorageErrorWithFallback(err, "graph_not_initialized") }, func(err error) error { return graphStorageErrorWithFallback(err, "route_unavailable") })
}
