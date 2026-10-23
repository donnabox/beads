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
	"github.com/steveyegge/beads/internal/storage"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/typecatalog"
)

var memoryGraphCmd = &cobra.Command{
	Use:         "memory",
	Short:       "Explore and record Memory relationships",
	Long:        "Explore and record Memory relationships in a graph-mode link workspace. Existing remember, recall, memories and forget retain their own contracts.",
	Annotations: map[string]string{skipStoreAnnotation: "1", canonicalDefaultAnnotation: "1"},
}

// Direct declaration captures have a fixed one-hop root context, not a CLI depth.
const memoryDirectSubjectsDepth = 1

// Family/direction select ordinary reads; constraints/subjects select their own
// fixed operations. Their depth is an internal capture bound only.
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
	{"requirements", "requires", "Show necessary Constraints, including defining children", graph.DirectionOut, 1000},
	{"required-by", "requires", "Show Constraints that directly require the input", graph.DirectionIn, 1},
	{"definition", "defines", "Show the defining Constraint subtree", graph.DirectionOut, 1000},
	{"defined-by", "defines", "Show the input Constraint's defining parent", graph.DirectionIn, 1},
	{"constraints", "subject-to", "Discover explicitly governing Constraints and their requirements", graph.DirectionOut, 1000},
	{"subjects", "subject-to", "Show direct explicit subjects of a Constraint", graph.DirectionIn, memoryDirectSubjectsDepth},
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
	if op.name != "constraints" && op.name != "subjects" {
		cmd.Flags().Int("depth", op.depth, "Maximum relationship hops for each input root (0..1000)")
	}
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
			return zero, fmt.Errorf("%w: BEADS_MAX_ROWS must be a non-negative integer", storage.ErrValidation)
		}
		if value > 0 {
			return zero, fmt.Errorf("%w: Memory graph reads use explicit node/Link bounds; unset BEADS_MAX_ROWS", graphstore.ErrCapabilityUnavailable)
		}
	}
	depth := op.depth
	if cmd.Flags().Lookup("depth") != nil {
		depth, _ = cmd.Flags().GetInt("depth")
	}
	nodes, _ := cmd.Flags().GetInt("max-nodes")
	links, _ := cmd.Flags().GetInt("max-links")
	if depth < 0 || depth > 1000 || nodes < 1 || nodes > 1000 || links < 1 || links > 1000 {
		return zero, fmt.Errorf("depth must be 0..1000 and node/Link bounds 1..1000")
	}
	request := graphstore.CurrentFamilyRequest{Limits: memoryquery.Limits{MaxNodes: nodes, MaxLinks: links, MaxInspected: 100000},
		MaxBytes: graphstore.PreviewCurrentReadByteLimit,
	}
	if op.name != "constraints" && op.name != "subjects" {
		request.Family = typecatalog.Reference{URI: scope + "types/" + op.family}
		request.Direction = op.direction
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
	resolvedArgs := append([]string(nil), args...)
	for i, arg := range resolvedArgs {
		resolved, err := graphPreviewResolveMemoryArgument(arg)
		if err != nil {
			return err
		}
		resolvedArgs[i] = resolved
	}
	request, err := memoryCurrentRequest(cmd, resolvedArgs, graphPreviewConfig.GraphScopeURL, op)
	if err != nil {
		if errors.Is(err, storage.ErrValidation) {
			return canonicalFailure(cmd, renderer, "invalid_properties", err.Error(), 2)
		}
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
		switch op.name {
		case "subjects":
			roots := make([]string, 0, len(request.Roots))
			for _, root := range request.Roots {
				roots = append(roots, root.Path)
			}
			result, err := store.ReadCurrentSubjects(ctx, graphstore.CurrentSubjectsRequest{Roots: roots, Limits: request.Limits, MaxBytes: request.MaxBytes})
			return result, "", err
		case "constraints":
			roots := make([]graphstore.CurrentConstraintRoot, 0, len(request.Roots))
			for _, root := range request.Roots {
				roots = append(roots, graphstore.CurrentConstraintRoot{Path: root.Path, RequiresDepth: root.Depth})
			}
			result, err := store.ReadCurrentConstraints(ctx, graphstore.CurrentConstraintsRequest{Roots: roots, Limits: request.Limits, MaxBytes: request.MaxBytes})
			return result, "", err
		default:
			result, err := store.ReadCurrentFamily(ctx, request)
			return result, "", err
		}
	}, func(value any, _ string) error {
		var read graphstore.CurrentFamilyRead
		var staged *graphstore.CurrentConstraintsRead
		switch captured := value.(type) {
		case graphstore.CurrentFamilyRead:
			read = captured
		case graphstore.CurrentConstraintsRead:
			read = captured.CurrentFamilyRead
			staged = &captured
		default:
			return canonicalReadError(cmd, renderer, graphstore.ErrInvalidStore, "invalid_store")
		}
		if err := validateMemoryCurrentCapture(read, graphPreviewConfig.GraphScopeURL, request.Roots); err != nil {
			return canonicalReadError(cmd, renderer, err, "invalid_store")
		}
		if err := validateMemoryDiscoveryCapture(read, staged, graphPreviewConfig.GraphScopeURL, op.name); err != nil {
			return canonicalReadError(cmd, renderer, err, "invalid_store")
		}
		if quietFlag && format != "json" {
			return nil
		}
		if format == "json" || format == "detail" {
			result, err := memoryCurrentResult(read, graphPreviewConfig.GraphScopeURL, op, request)
			if err != nil {
				return canonicalReadError(cmd, renderer, err, "invalid_store")
			}
			if op.name == "subjects" {
				result.Traversal.Discovery = "direct-incoming-subject-to"
			}
			if staged != nil {
				result.Traversal.Discovery = "explicit-subject-to-and-requires"
				if err := memoryConstraintStages(&result, *staged); err != nil {
					return canonicalReadError(cmd, renderer, err, "invalid_store")
				}
			}
			return renderMemoryCanonicalResult(cmd, renderer, result, format)
		}
		// Bind presentation to the exact direction used for acquisition. The
		// private capture is not an independently reusable query description.
		disclosure := memoryDiscoveryLine(op.name)
		presentationBudget := graphstore.PreviewCurrentReadByteLimit - len(disclosure)
		var text string
		var err error
		switch {
		case staged != nil:
			text, err = renderMemoryConstraintsGraph(*staged, graphPreviewConfig.GraphScopeURL, presentationBudget)
		case op.name == "subjects":
			text, err = renderMemorySubjectsGraph(read, graphPreviewConfig.GraphScopeURL, presentationBudget)
		default:
			text, err = renderMemoryCurrentGraph(read, graphPreviewConfig.GraphScopeURL, request.Direction, presentationBudget)
		}
		if err != nil {
			return memoryGraphOutputError(cmd, renderer, err)
		}
		if err := writeCanonicalDocument(cmd.OutOrStdout(), []byte(text+disclosure)); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "output_error: %v\n", err)
			return &exitError{Code: 1}
		}
		return nil
	}, func(err error) error { return graphStorageErrorWithFallback(err, "graph_not_initialized") }, func(err error) error { return graphStorageErrorWithFallback(err, "route_unavailable") })
}
