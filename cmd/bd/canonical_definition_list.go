package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/clioutput"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/resourceget"
)

// Definitions retain their complete Type References and owned Link captures.
// They do not pass through the legacy string-Type collection codec.
func intrinsicDefinitionCollectionDescriptor(scope string, records []json.RawMessage) (clioutput.CollectionDescriptor, error) {
	item := canonicalResourceDescriptor()
	item.BDPSpec = intrinsicDefinitionContract
	groups := make(map[string]*clioutput.Protection)
	for _, raw := range records {
		doc, err := resourceget.NewDefinitionDocument(scope, resourceget.DefinitionBead, raw)
		if err != nil {
			return clioutput.CollectionDescriptor{}, err
		}
		protection := canonicalShowDescriptor("", doc).Protection
		for typ, linkProtection := range protection.Members["ownedLinks"].Members {
			groups[typ] = linkProtection
		}
	}
	item.Protection.Members = map[string]*clioutput.Protection{"ownedLinks": {Members: groups}}
	return clioutput.CollectionDescriptor{Item: item, ResultKind: "resource_collection", ItemsMember: "items"}, nil
}

// This adapter is confined to the explicit fresh definition layout. Selector
// interpretation belongs to its checked native catalog, never Issue aliases.
func runIntrinsicDefinitionList(cmd *cobra.Command) error {
	opts, selected, err := canonicalOutputOptions(cmd)
	if err != nil {
		return canonicalFailure(cmd, canonicalFailureRendererFor(cmd), "invalid_selector", err.Error(), 2)
	}
	if !selected {
		opts.JSON, _ = cmd.Flags().GetBool("json")
	}
	descriptor, err := intrinsicDefinitionCollectionDescriptor(graphPreviewConfig.GraphScopeURL, nil)
	if err != nil {
		return err
	}
	renderer, err := clioutput.NewCollection(descriptor, opts)
	if err != nil {
		return canonicalFailure(cmd, canonicalFailureRendererFor(cmd), "invalid_selector", err.Error(), 2)
	}
	jsonl := len(opts.Formats) > 0 && opts.Formats[0] == "jsonl"
	failure := canonicalCollectionFailure{renderer: renderer, jsonl: jsonl}
	invalid := func(err error) error {
		return canonicalFailure(cmd, failure, "invalid_selector", err.Error(), 2)
	}
	if len(cmd.Flags().Args()) != 0 {
		return invalid(fmt.Errorf("definition listing does not accept positional arguments; use --type"))
	}
	allowed := []string{"type", "exact-type", "conforms-to", "paged", "unpaged", "page-size", "limit", "continue"}
	if !selected {
		allowed = append(allowed, "json")
	}
	if err := canonicalGraphFlags(cmd, failure, allowed...); err != nil {
		return err
	}
	if err := canonicalScalarOccurrences(cmd, os.Args[1:], "type", "exact-type", "conforms-to", "paged", "unpaged", "page-size", "limit", "continue"); err != nil {
		return invalid(err)
	}
	if cmd.Flags().Changed("continue") {
		return canonicalReadError(cmd, failure, graphstore.ErrCapabilityUnavailable, "route_unavailable")
	}
	family, exact, err := intrinsicDefinitionListTypes(cmd)
	if err != nil {
		return invalid(err)
	}
	paged, _ := cmd.Flags().GetBool("paged")
	unpaged, _ := cmd.Flags().GetBool("unpaged")
	pageSize, _ := cmd.Flags().GetInt("page-size")
	limit := 0
	if cmd.Flags().Changed("limit") {
		limit, _ = cmd.Flags().GetInt("limit")
	}
	format := "default"
	if len(opts.Formats) > 0 {
		format = opts.Formats[0]
	}
	paging, err := clioutput.ResolvePaging(clioutput.PagingOptions{Paged: paged, Unpaged: unpaged, PageSize: pageSize, Limit: limit}, format)
	if err != nil {
		return invalid(err)
	}
	plan := canonicalCollectionPlan{options: opts, paging: paging, jsonl: jsonl, definitionScope: graphPreviewConfig.GraphScopeURL}
	return canonicalReadWithGraphStore(cmd, failure, func(ctx context.Context, store *graphstore.Store) (any, error) {
		return store.ListDefinitions(ctx, family, exact)
	}, func(value any) error {
		documents := value.([]resourceget.DefinitionDocument)
		plan.definitionDocuments = append([]resourceget.DefinitionDocument{}, documents...)
		records := make([]json.RawMessage, len(documents))
		for i, document := range documents {
			records[i] = document.NativeJSON()
		}
		return renderCanonicalCollection(cmd, plan, records, failure)
	})
}

func intrinsicDefinitionListTypes(cmd *cobra.Command) (family, exact string, err error) {
	if cmd.Flags().Changed("type") && cmd.Flags().Changed("conforms-to") {
		return "", "", fmt.Errorf("--type and --conforms-to both select the Type family; supply only one")
	}
	for _, name := range []string{"type", "conforms-to", "exact-type"} {
		if !cmd.Flags().Changed(name) {
			continue
		}
		raw, _ := cmd.Flags().GetString(name)
		if raw == "" {
			return "", "", fmt.Errorf("--%s requires a Type", name)
		}
		if name == "conforms-to" {
			if err := graph.ValidateTypeURL(raw); err != nil {
				return "", "", fmt.Errorf("--conforms-to requires a canonical absolute HTTP(S) Type URI")
			}
		}
		if name == "exact-type" {
			exact = raw
		} else {
			family = raw
		}
	}
	return family, exact, nil
}
