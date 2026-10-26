package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/spf13/cobra"
	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/clioutput"
	"github.com/steveyegge/beads/internal/httpapi/bdpwire"
	"github.com/steveyegge/beads/internal/httpapi/graphread"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func init() {
	for _, cmd := range []*cobra.Command{listCmd, linkListCmd} {
		registerCanonicalCollectionFlags(cmd)
		links := cmd == linkListCmd
		canonicalFailureFactories[cmd] = func(options clioutput.Options) (canonicalFailureRenderer, error) {
			descriptor, err := canonicalCollectionDescriptor(links, nil)
			if err != nil {
				return nil, err
			}
			renderer, err := clioutput.NewCollection(descriptor, options)
			if err != nil {
				return nil, err
			}
			jsonl := len(options.Formats) > 0 && options.Formats[0] == "jsonl"
			return canonicalCollectionFailure{renderer: renderer, jsonl: jsonl}, nil
		}
	}
	linkListCmd.Flags().Int("limit", 0, "Requested total delivery bound; 0 means unlimited; canonical local reads refuse larger selections without emitting a subset")
	linkListCmd.Flags().Lookup("limit").Annotations = map[string][]string{"canonical-output-only": {"true"}}
}

func registerCanonicalCollectionFlags(cmd *cobra.Command) {
	// Legacy positional diagnostics stay on the existing surface. Canonical
	// admission waits for a renderer so JSON/JSONL refusals share its contract.
	legacyArgs := cmd.Args
	cmd.Args = func(cmd *cobra.Command, args []string) error {
		if canonicalOutputSelected(cmd) || legacyArgs == nil {
			return nil
		}
		return legacyArgs(cmd, args)
	}
	registerCanonicalOutputFlags(cmd)
	cmd.Flags().Lookup("output-format").Usage = "Canonical collection output: default, table, detail, json or jsonl; cannot mix legacy --json/--format"
	cmd.Flags().String("exact-type", "", "Select only the exact declared Type; --type includes refinements")
	cmd.Flags().String("conforms-to", "", "Select a canonical absolute HTTP(S) Type URI and its conformance family; compatibility alias for --type")
	cmd.Flags().Bool("paged", false, "Return one page in canonical output; current local provider refuses truncation")
	cmd.Flags().Bool("unpaged", false, "Return the complete bounded current selection in canonical output")
	cmd.Flags().Int("page-size", 50, "Canonical page size; use --unpaged if the complete local selection exceeds this boundary")
	cmd.Flags().String("continue", "", "Resume a provider continuation (not available from the current local collection provider)")
	for _, name := range []string{"conforms-to", "paged", "unpaged", "page-size", "continue"} {
		cmd.Flags().Lookup(name).Annotations = map[string][]string{"canonical-output-only": {"true"}}
	}
}

// The local Reader returns one complete transactional selection, not a retained
// cursor authority. Do not construct an ephemeral pager and publish dead cursors,
// or set native next:null after discarding undisplayed rows.
type canonicalCollectionPlan struct {
	selection   *graphread.Selection
	linkRequest graphstore.LinkCollectionRequest
	options     clioutput.Options
	paging      clioutput.Paging
	links       bool
	jsonl       bool
}

type canonicalCollectionFailure struct {
	renderer *clioutput.CollectionRenderer
	jsonl    bool
}

// Distinguish store-bound endpoint/pin failures from pure invocation errors.
// The latter have already been checked before opening the graph.
type canonicalCollectionEndpointReadError struct{ error }

func (e *canonicalCollectionEndpointReadError) Unwrap() error { return e.error }

func (f canonicalCollectionFailure) RenderFailure(origin string, problem json.RawMessage) ([]byte, error) {
	if !f.jsonl {
		return f.renderer.RenderFailure(origin, problem)
	}
	var out bytes.Buffer
	stream, err := f.renderer.NewJSONL(&out)
	if err != nil {
		return nil, err
	}
	if err := stream.Failure(origin, problem); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func canonicalCollectionDescriptor(links bool, records []json.RawMessage) (clioutput.CollectionDescriptor, error) {
	item := canonicalResourceDescriptor()
	if links {
		item.Protection = &clioutput.Protection{Required: []string{"id", "type", "revision", "source", "target"}}
	} else {
		groups := make(map[string]*clioutput.Protection)
		for _, raw := range records {
			var bead bdpwire.BeadRecord
			if err := bdpwire.Unmarshal(raw, &bead); err != nil {
				return clioutput.CollectionDescriptor{}, err
			}
			for typ := range bead.OwnedLinks {
				groups[typ] = &clioutput.Protection{Items: &clioutput.Protection{Required: []string{"id", "type", "revision", "source", "target"}}}
			}
		}
		item.Protection.Members = map[string]*clioutput.Protection{"ownedLinks": {Members: groups}}
	}
	return clioutput.CollectionDescriptor{Item: item, ResultKind: "resource_collection", ItemsMember: "items"}, nil
}

func canonicalCollectionRequest(cmd *cobra.Command, links bool, options clioutput.Options) (canonicalCollectionPlan, error) {
	return canonicalCollectionRequestIn(cmd, links, options, nil, nil)
}

func canonicalCollectionRequestIn(cmd *cobra.Command, links bool, options clioutput.Options, ctx context.Context, store *graphstore.Store) (canonicalCollectionPlan, error) {
	plan, exact, family, err := canonicalCollectionPureRequest(cmd, links, options)
	if err != nil {
		return plan, err
	}
	parameters := url.Values{}
	if exact != "" {
		parameters.Set("type", exact)
	}
	if family != "" {
		parameters.Set("conformsTo", family)
	}
	collection := "beads"
	if links {
		collection = "links"
		if store == nil {
			plan.linkRequest, err = graphPreviewLinkCollectionRequest(cmd, cmd.Flags().Args())
		} else {
			plan.linkRequest, err = graphPreviewLinkCollectionRequestWith(cmd, cmd.Flags().Args(),
				func(selector string) (string, string, graphstore.LinkPinSelection, error) {
					return graphPreviewLinkCollectionEndpointRaw(ctx, store, selector)
				}, func(message string) error { return errors.New(message) })
		}
		if err != nil {
			if store != nil {
				return plan, &canonicalCollectionEndpointReadError{err}
			}
			return plan, err
		}
		// The BDP selector has no pair-union parameter. For a two-Bead
		// bidirectional CLI selection, capture the complete typed inventory
		// and apply the shared Link predicate before paging or output.
		if plan.linkRequest.PairDirection != "both" {
			source, target := plan.linkRequest.SourceURL, plan.linkRequest.TargetURL
			if plan.linkRequest.PairDirection == "in" {
				source, target = target, source
			}
			for key, uri := range map[string]string{"source": source, "target": target, "endpoint": plan.linkRequest.IncidentURL} {
				if uri != "" {
					parameters.Set(key, uri)
				}
			}
		}
	}
	plan.selection, err = graphread.CompileCollection(graphPreviewConfig.GraphScopeURL, collection, parameters, graphread.SelectorLimits{Bytes: 16384, Depth: 256, Nodes: 2048})
	var parameter *graphread.ParameterError
	if errors.As(err, &parameter) && (parameter.Name == "type" || parameter.Name == "conformsTo") {
		flag := "type"
		if parameter.Name == "conformsTo" {
			flag = "conforms-to"
		}
		return plan, fmt.Errorf("--%s requires a canonical absolute HTTP(S) Type URI (for a local Type, %sNAME); use an installed Type or native classification shorthand", flag, graphPreviewConfig.GraphScopeURL)
	}
	return plan, err
}

// Reject deterministic option and selector syntax before opening a graph.
// Endpoint allocation and pin checks stay inside the canonical read owner.
func canonicalCollectionPureRequest(cmd *cobra.Command, links bool, options clioutput.Options) (canonicalCollectionPlan, string, string, error) {
	plan := canonicalCollectionPlan{options: options, links: links}
	format := "default"
	if len(options.Formats) != 0 {
		format = options.Formats[0]
	}
	plan.jsonl = format == "jsonl"
	paged, _ := cmd.Flags().GetBool("paged")
	unpaged, _ := cmd.Flags().GetBool("unpaged")
	pageSize, _ := cmd.Flags().GetInt("page-size")
	limit := 0
	if cmd.Flags().Changed("limit") {
		limit, _ = cmd.Flags().GetInt("limit")
	}
	if pageSize <= 0 {
		return plan, "", "", fmt.Errorf("--page-size must be positive")
	}
	paging, err := clioutput.ResolvePaging(clioutput.PagingOptions{Paged: paged, Unpaged: unpaged, PageSize: pageSize, Limit: limit}, format)
	if err != nil {
		return plan, "", "", err
	}
	plan.paging = paging
	if cmd.Flags().Changed("continue") {
		return plan, "", "", fmt.Errorf("%w: local collection reads cannot resume retained continuations; use a complete --unpaged selection", graphstore.ErrCapabilityUnavailable)
	}
	exact, family, err := graphPreviewCollectionTypes(cmd, graphPreviewConfig.GraphScopeURL, links)
	return plan, exact, family, err
}

func canonicalCollectionPreflight(cmd *cobra.Command, links bool, options clioutput.Options) error {
	if _, _, _, err := canonicalCollectionPureRequest(cmd, links, options); err != nil {
		return err
	}
	if !links {
		return nil
	}
	_, err := graphPreviewLinkCollectionRequestWith(cmd, cmd.Flags().Args(),
		func(selector string) (string, string, graphstore.LinkPinSelection, error) {
			ref, err := graphPreviewParseReference(selector)
			if err != nil {
				return "", "", graphstore.LinkPinSelection{}, err
			}
			path, _, err := graphPreviewResourceReadSelectorIn(graphPreviewConfig.GraphScopeURL, ref.Identity)
			return path, "", graphstore.LinkPinSelection{}, err
		}, func(message string) error { return errors.New(message) })
	return err
}

func canonicalCollectionRequestFailure(cmd *cobra.Command, failure canonicalCollectionFailure, err error) error {
	var endpoint *canonicalCollectionEndpointReadError
	if errors.As(err, &endpoint) {
		if errors.Is(endpoint.error, graph.ErrValidation) {
			return canonicalFailure(cmd, failure, "invalid_selector", endpoint.error.Error(), 2)
		}
		return canonicalReadError(cmd, failure, endpoint.error, "route_unavailable")
	}
	if errors.Is(err, graphstore.ErrCapabilityUnavailable) {
		return canonicalFailure(cmd, failure, "capability_unavailable", err.Error(), 5)
	}
	if errors.Is(err, graphstore.ErrNotFound) || errors.Is(err, graphstore.ErrGone) || errors.Is(err, graphstore.ErrVersionUnknown) || errors.Is(err, graphstore.ErrInvalidStore) {
		return canonicalReadError(cmd, failure, err, "invalid_selector")
	}
	return canonicalFailure(cmd, failure, "invalid_selector", err.Error(), 2)
}

func runCanonicalCollection(cmd *cobra.Command, links bool) error {
	options, _, err := canonicalOutputOptions(cmd)
	if err != nil {
		return canonicalFailure(cmd, canonicalFailureRendererFor(cmd), "invalid_selector", err.Error(), 2)
	}
	descriptor, err := canonicalCollectionDescriptor(links, nil)
	if err != nil {
		return err
	}
	renderer, err := clioutput.NewCollection(descriptor, options)
	if err != nil {
		return canonicalFailure(cmd, canonicalFailureRendererFor(cmd), "invalid_selector", err.Error(), 2)
	}
	format := "default"
	if len(options.Formats) > 0 {
		format = options.Formats[0]
	}
	failure := canonicalCollectionFailure{renderer: renderer, jsonl: format == "jsonl"}
	if len(cmd.Flags().Args()) != 0 && !links {
		return canonicalFailure(cmd, failure, "invalid_selector", "canonical collection listing does not accept positional arguments; use selection flags", 2)
	}
	allowed := []string{"type", "exact-type", "conforms-to", "paged", "unpaged", "page-size", "limit", "continue"}
	if links {
		allowed = append(allowed, "from", "to", "incident-to", "direction", "all-revisions")
	}
	if err := canonicalGraphFlags(cmd, failure, allowed...); err != nil {
		return err
	}
	if err := canonicalScalarOccurrences(cmd, os.Args[1:], "type", "exact-type", "conforms-to", "from", "to", "incident-to", "direction", "all-revisions", "paged", "unpaged", "page-size", "limit", "continue"); err != nil {
		return canonicalFailure(cmd, failure, "invalid_selector", err.Error(), 2)
	}
	if err := canonicalCollectionPreflight(cmd, links, options); err != nil {
		return canonicalCollectionRequestFailure(cmd, failure, err)
	}
	ctx, cancel := context.WithTimeout(getRootContext(), 30*time.Second)
	defer cancel()
	store, err := graphstore.OpenExisting(ctx, graphOptions(graphPreviewConfig))
	if err != nil {
		return canonicalReadError(cmd, failure, err, "graph_not_initialized")
	}
	closed := false
	defer func() {
		if !closed {
			_ = store.Close()
		}
	}()
	plan, err := canonicalCollectionRequestIn(cmd, links, options, ctx, store)
	if err != nil {
		return canonicalCollectionRequestFailure(cmd, failure, err)
	}
	inventory, err := graphread.New(store).Inventory(ctx)
	if err != nil {
		return canonicalReadError(cmd, failure, err, "route_unavailable")
	}
	records, err := selectCanonicalCollection(plan, inventory)
	if err != nil {
		return canonicalReadError(cmd, failure, err, "route_unavailable")
	}
	closeErr := store.Close()
	closed = true
	if closeErr != nil {
		return canonicalReadError(cmd, failure, closeErr, "route_unavailable")
	}
	return renderCanonicalCollection(cmd, plan, records, failure)
}

func selectCanonicalCollection(plan canonicalCollectionPlan, inventory graphread.Inventory) ([]json.RawMessage, error) {
	records, err := plan.selection.Select(inventory)
	if err != nil || !plan.links {
		return records, err
	}
	request := plan.linkRequest
	current := make(map[string]string)
	for _, bead := range inventory.Beads {
		for _, selected := range []struct{ path, url string }{
			{request.SourcePath, request.SourceURL}, {request.TargetPath, request.TargetURL}, {request.IncidentPath, request.IncidentURL},
		} {
			if selected.path != "" && bead.ID == selected.url {
				current[selected.path] = bead.Revision
			}
		}
	}
	for _, path := range []string{request.SourcePath, request.TargetPath, request.IncidentPath} {
		if path != "" && current[path] == "" {
			return nil, graphstore.ErrNotFound
		}
	}
	filtered := make([]json.RawMessage, 0, len(records))
	for _, raw := range records {
		var link bdpwire.LinkRecord
		if err := bdpwire.Unmarshal(raw, &link); err != nil {
			return nil, fmt.Errorf("%w: current Link collection item: %v", graphstore.ErrInvalidStore, err)
		}
		captured := graphstore.LinkRecord{Source: link.Source.URI, Target: link.Target.URI,
			SourceRevision: link.Source.Revision, TargetRevision: link.Target.Revision}
		if graphstore.LinkCollectionMatches(captured, request, current) {
			filtered = append(filtered, raw)
		}
	}
	return filtered, nil
}

func renderCanonicalCollection(cmd *cobra.Command, plan canonicalCollectionPlan, records []json.RawMessage, failure canonicalCollectionFailure) error {
	if records == nil {
		records = []json.RawMessage{}
	}
	boundary, err := plan.paging.PlanPage(0, len(records), true, nil, nil)
	if err != nil {
		return canonicalFailure(cmd, failure, "invalid_selector", err.Error(), 2)
	}
	// PlanPage's page size is a per-fetch request, not a truncate instruction.
	// This provider cannot page, so explicitly check the requested first boundary.
	size, err := plan.paging.RequestSize(0)
	if err != nil {
		return canonicalFailure(cmd, failure, "invalid_selector", err.Error(), 2)
	}
	paged, _ := cmd.Flags().GetBool("paged")
	unpaged, _ := cmd.Flags().GetBool("unpaged")
	onePage := paged || (!unpaged && !plan.jsonl)
	if boundary.Take != len(records) || (onePage && len(records) > size) {
		bound := fmt.Sprintf("requested page boundary of %d", size)
		if boundary.Take != len(records) {
			bound = fmt.Sprintf("requested total delivery limit of %d", boundary.Take)
		}
		return canonicalFailure(cmd, failure, "capability_unavailable", fmt.Sprintf("complete selection of %d items exceeds the %s; local collection reads cannot satisfy this bound with a safe continuation, so no items were emitted; use --unpaged --limit 0 or a narrower selection", len(records), bound), 5)
	}
	descriptor, err := canonicalCollectionDescriptor(plan.links, records)
	if err != nil {
		return canonicalFailure(cmd, failure, "output_error", err.Error(), 1)
	}
	renderer, err := clioutput.NewCollection(descriptor, plan.options)
	if err != nil {
		return canonicalFailure(cmd, failure, "invalid_selector", err.Error(), 2)
	}
	// Qualify the same current public DTOs used by BDP Read before projecting.
	for _, record := range records {
		if plan.links {
			var link bdpwire.LinkRecord
			err = bdpwire.Unmarshal(record, &link)
		} else {
			var bead bdpwire.BeadRecord
			err = bdpwire.Unmarshal(record, &bead)
		}
		if err != nil {
			return canonicalFailure(cmd, failure, "output_error", fmt.Sprintf("BDP collection item: %v", err), 1)
		}
	}
	if plan.jsonl {
		stream, err := renderer.NewJSONL(cmd.OutOrStdout())
		if err != nil {
			return canonicalCollectionWriteError(cmd, err)
		}
		for _, record := range records {
			if err := stream.Item(record); err != nil {
				code, status := "output_error", 1
				if errors.Is(err, clioutput.ErrInvalidProjection) {
					code, status = "invalid_selector", 2
				}
				problem, _ := json.Marshal(map[string]any{"code": code, "message": err.Error(), "retryable": false})
				terminalErr := stream.Failure("cli", problem)
				fmt.Fprintf(cmd.ErrOrStderr(), "%s: %v\n", code, errors.Join(err, terminalErr)) //nolint:gosec // G705: CLI stderr diagnostics are not HTML.
				return &exitError{Code: status}
			}
		}
		if err := stream.Finish(true, "exhausted", nil); err != nil {
			return canonicalCollectionWriteError(cmd, err)
		}
		return nil
	}
	payload, err := json.Marshal(graphread.Page{Items: records, Next: nil})
	if err != nil {
		return canonicalFailure(cmd, failure, "output_error", err.Error(), 1)
	}
	exhausted := true
	data, err := renderer.Render(payload, clioutput.Completion{Returned: len(records), Exhausted: &exhausted, StopReason: "exhausted"})
	if err != nil {
		if errors.Is(err, clioutput.ErrInvalidProjection) {
			return canonicalFailure(cmd, failure, "invalid_selector", err.Error(), 2)
		}
		return canonicalFailure(cmd, failure, "output_error", err.Error(), 1)
	}
	if err := writeCanonicalDocument(cmd.OutOrStdout(), data); err != nil {
		return canonicalCollectionWriteError(cmd, err)
	}
	return nil
}

func canonicalCollectionWriteError(cmd *cobra.Command, err error) error {
	fmt.Fprintf(cmd.ErrOrStderr(), "output_error: %v\n", err) //nolint:gosec // G705: CLI stderr diagnostics are not HTML.
	return &exitError{Code: 1}
}
