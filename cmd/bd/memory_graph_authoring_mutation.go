package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/clioutput"
	"github.com/steveyegge/beads/internal/httpapi/bdpwire"
	"github.com/steveyegge/beads/internal/httpapi/graphread"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// This interface deliberately has no Resource reader. Installed fixed Type and
// ownership context is acquired before the write; only the writer's postimages
// may be projected afterward. The command never retries a mutation.
type memoryLinkStore interface {
	ListInstalledTypes(context.Context) ([]graph.TypeDescriptor, error)
	SelectedOwnedLinkTypesWithDescriptors(string, []graph.TypeDescriptor, ...string) ([]string, error)
	AddInformationalLink(context.Context, graphstore.LinkCreateRequest) (graphstore.LinkMutationResult, error)
	Close() error
}

func newMemoryMutationFailureRenderer(options clioutput.Options) (canonicalFailureRenderer, error) {
	// A rejected format/mask must not suppress a separately requested JSON error
	// frame. Preserve raw-envelope intent and JSON intent, not invalid selectors.
	for _, format := range options.Formats {
		options.JSON = options.JSON || format == "json"
	}
	options.Formats = nil
	options.Include = nil
	var omit []string
	for _, path := range options.Omit {
		if path == "envelope" {
			omit = append(omit, path)
		}
	}
	options.Omit = omit
	return newMemoryMutationRenderer(options)
}
func newMemoryMutationRenderer(options clioutput.Options) (*clioutput.MutationRenderer, error) {
	return clioutput.NewMutation(clioutput.MutationDescriptor{BDPSpec: bdpwire.Pin, Required: []string{"link", "source", "changed"}}, options)
}

func runMemoryLinkCommand(cmd *cobra.Command, args []string, op memoryLinkOperation) error {
	options, _, err := canonicalOutputOptions(cmd)
	if err != nil {
		return canonicalFailure(cmd, nil, "invalid_selector", err.Error(), 2)
	}
	renderer, err := newMemoryMutationRenderer(options)
	if err != nil {
		return canonicalFailure(cmd, canonicalFailureRendererFor(cmd), "invalid_selector", err.Error(), 2)
	}
	state := memoryLinkExecution{}
	previous := graphFailureReporter
	graphFailureReporter = func(code, message string, status int) error {
		return memoryLinkProblem(cmd, renderer, code, message, status, state)
	}
	defer func() { graphFailureReporter = previous }()
	if !graphPreviewActive {
		return graphFailure("capability_unavailable", "Memory authoring requires a graph-mode link workspace", 5)
	}
	if err := graphPreviewWritePolicy(); err != nil {
		return err
	}
	flags := []string{"from", "to", "title", "body", "status", "properties", "metadata", "id", "source-revision", "target-revision", "if-source-revision", "json"}
	if err := canonicalGraphFlags(cmd, renderer, flags...); err != nil {
		return err
	}
	if err := canonicalScalarOccurrences(cmd, os.Args[1:], flags...); err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	if err := canonicalExactOccurrences(cmd, os.Args[1:], "from", "to"); err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	source, target, err := graphPreviewAuthoringEndpoints(cmd, args)
	if err != nil {
		err = fmt.Errorf("%s %w", op.name, err)
	}
	if err == nil {
		source, err = graphPreviewResolveMemoryArgument(source)
	}
	if err == nil {
		target, err = graphPreviewResolveMemoryArgument(target)
	}
	var request graphstore.LinkCreateRequest
	if err == nil {
		request, err = memoryLinkRequestWithEndpoints(cmd, graphPreviewConfig.GraphScopeURL, op, source, target)
	}
	if err != nil {
		var exit *exitError
		if errors.As(err, &exit) {
			return err
		}
		return graphPreviewReferenceFailure(err)
	}
	request.Properties, err = memoryLinkProperties(cmd)
	if err != nil {
		return graphFailure("invalid_properties", err.Error(), 2)
	}
	request.Metadata, err = graphPreviewMetadataCreate(cmd)
	if err != nil {
		return err
	}
	request.Actor = getActorWithGit()
	ctx, cancel := context.WithTimeout(getRootContext(), 30*time.Second)
	defer cancel()
	store, err := graphstore.OpenExisting(ctx, graphOptions(graphPreviewConfig))
	if err != nil {
		return graphStorageErrorWithFallback(err, "graph_not_initialized")
	}
	state = executeMemoryLinkCreate(ctx, store, request, encodeMemoryGraphJSON)
	return publishMemoryLinkExecution(cmd, renderer, options, state)
}

type memoryLinkExecution struct {
	request   graphstore.LinkCreateRequest
	result    graphstore.LinkMutationResult
	applied   bool
	raw       json.RawMessage
	err       error
	outputErr bool
}

// Close is checked after capturing the returned operation result. In particular,
// a cleanup error must not discard an already applied write's actual effect.
func executeMemoryLinkCreate(ctx context.Context, store memoryLinkStore, request graphstore.LinkCreateRequest, encode func(any) ([]byte, error)) (state memoryLinkExecution) {
	state.request = request
	closed := false
	defer func() {
		if !closed {
			_ = store.Close()
		}
	}()
	selection, err := memoryLinkProjectionContext(ctx, store)
	if err == nil {
		state.result, err = store.AddInformationalLink(ctx, request)
		state.applied = err == nil
		if state.applied {
			state.raw, err = projectMemoryLinkMutation(selection, state.result, encode)
			state.outputErr = err != nil
		}
	}
	closed = true
	state.err = errors.Join(err, store.Close())
	return state
}

func memoryLinkProjectionContext(ctx context.Context, store memoryLinkStore) (graphstore.Snapshot, error) {
	types, err := store.ListInstalledTypes(ctx)
	if err != nil {
		return graphstore.Snapshot{}, err
	}
	if len(types) == 0 {
		return graphstore.Snapshot{}, fmt.Errorf("%w: missing installed Type context", graphstore.ErrInvalidStore)
	}
	selection := graphstore.Snapshot{Types: types, OwnedLinkTypes: map[string][]string{}, Records: []any{}}
	for _, typ := range types {
		if typ.Describes() != graph.KindBead {
			continue
		}
		owned, err := store.SelectedOwnedLinkTypesWithDescriptors(typ.ID(), types)
		if err != nil {
			return graphstore.Snapshot{}, err
		}
		if owned == nil {
			return graphstore.Snapshot{}, fmt.Errorf("%w: missing ownership context", graphstore.ErrInvalidStore)
		}
		selection.OwnedLinkTypes[typ.ID()] = owned
	}
	if _, err := graphread.ProjectCurrentRecords(selection); err != nil {
		return graphstore.Snapshot{}, err
	}
	return selection, nil
}

func projectMemoryLinkMutation(selection graphstore.Snapshot, result graphstore.LinkMutationResult, encode func(any) ([]byte, error)) (json.RawMessage, error) {
	selection.Records = []any{result.Source, result.Link}
	projected, err := graphread.ProjectCurrentRecords(selection)
	if err != nil {
		return nil, err
	}
	if len(projected.Beads) != 1 || len(projected.Links) != 1 {
		return nil, fmt.Errorf("%w: invalid Link mutation postimages", graphstore.ErrInvalidStore)
	}
	var replaced json.RawMessage
	if result.ReplacedSource != nil {
		replaced, err = memoryLinkPredecessor(result.ReplacedSource, encode)
		if err != nil {
			return nil, err
		}
	}
	// This is the existing native Link result vocabulary, with actual Resources
	// substituted for storage DTOs. No new envelope, schema version or inference.
	return encode(struct {
		Link           bdpwire.LinkRecord `json:"link"`
		Source         bdpwire.BeadRecord `json:"source"`
		Changed        bool               `json:"changed"`
		ReplacedSource json.RawMessage    `json:"replacedSource,omitempty"`
	}{projected.Links[0], projected.Beads[0], result.Changed, replaced})
}

func publishMemoryLinkExecution(cmd *cobra.Command, renderer *clioutput.MutationRenderer, options clioutput.Options, state memoryLinkExecution) error {
	if state.err != nil {
		// Uncertainty always outranks any joined validation, projection or Close error.
		if errors.Is(state.err, graphstore.ErrOutcomeUnknown) {
			return memoryLinkProblem(cmd, renderer, "outcome_unknown", state.err.Error()+"; do not automatically replay; inspect the source and Link identity before retry", 6, state)
		}
		if state.outputErr {
			return memoryLinkProblem(cmd, renderer, "output_error", state.err.Error(), 1, state)
		}
		previous := graphFailureReporter
		graphFailureReporter = func(code, message string, status int) error {
			return memoryLinkProblem(cmd, renderer, code, message, status, state)
		}
		defer func() { graphFailureReporter = previous }()
		return graphStorageErrorWithFallback(state.err, "route_unavailable")
	}
	jsonFormat := options.JSON
	for _, format := range options.Formats {
		jsonFormat = jsonFormat || format == "json"
	}
	if quietFlag && !jsonFormat {
		return nil
	}
	if !jsonFormat && len(options.Include) == 0 && len(options.Omit) == 0 {
		return writeMemoryLinkDocument(cmd, []byte(graphHumanRow(graphPreviewConfig.GraphScopeURL, state.result.Link)+"\n"), state, false)
	}
	data, err := renderer.RenderSuccess(state.raw)
	if err != nil {
		return memoryLinkProblem(cmd, renderer, "output_error", err.Error(), 1, state)
	}
	return writeMemoryLinkDocument(cmd, data, state, false)
}

// This is predecessor identity context, not an invented complete Resource.
// Preserve only facts returned by the writer and use public attribution names.
func memoryLinkPredecessor(previous *graphstore.ReplacedMemory, encode func(any) ([]byte, error)) (json.RawMessage, error) {
	var attribution *bdpwire.Attribution
	if previous.Attribution.Actor != "" {
		var basis bdpwire.AttributionBasis
		switch previous.Attribution.Status {
		case "claimed":
			basis = bdpwire.AttributionWriterSupplied
		case "unknown":
			basis = bdpwire.AttributionUnknown
		default:
			return nil, fmt.Errorf("%w: unsupported predecessor attribution", graphstore.ErrInvalidStore)
		}
		attribution = &bdpwire.Attribution{Principal: previous.Attribution.Actor, Basis: basis}
	} else if previous.Attribution.Status != "" && previous.Attribution.Status != "unknown" {
		return nil, fmt.Errorf("%w: unsupported predecessor attribution", graphstore.ErrInvalidStore)
	}
	return encode(struct {
		ID          string               `json:"id"`
		Revision    string               `json:"revision"`
		Attribution *bdpwire.Attribution `json:"attribution,omitempty"`
	}{previous.ID, previous.Version, attribution})
}
