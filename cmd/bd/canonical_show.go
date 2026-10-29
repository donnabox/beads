package main

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/clioutput"
	"github.com/steveyegge/beads/internal/httpapi/bdpwire"
	"github.com/steveyegge/beads/internal/httpapi/graphread"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/resourceget"
)

func init() {
	registerCanonicalOutputFlags(showCmd)
	registerCanonicalOutputFlags(linkShowCmd)
}

func canonicalLinkProtection() *clioutput.Protection {
	return &clioutput.Protection{Required: []string{"id", "type", "revision", "source", "target"}}
}

// The fresh local definition adapter adopts PR316's exact local/global identity
// contract. Legacy records retain bdpwire.Pin and its legacy DTOs unchanged.
const intrinsicDefinitionContract = "c6dfb3c05fe707cf605030c5661123e8dda0d8f8"

func canonicalShowDescriptor(_ string, record any) clioutput.Descriptor {
	descriptor := canonicalResourceDescriptor()
	if doc, ok := record.(resourceget.DefinitionDocument); ok {
		descriptor.BDPSpec = intrinsicDefinitionContract
		if doc.Kind() == resourceget.DefinitionLink {
			descriptor.Protection = canonicalLinkProtection()
		} else {
			var body struct {
				OwnedLinks map[string]json.RawMessage `json:"ownedLinks"`
			}
			if json.Unmarshal(doc.NativeJSON(), &body) == nil {
				groups := make(map[string]*clioutput.Protection, len(body.OwnedLinks))
				for typ := range body.OwnedLinks {
					groups[typ] = &clioutput.Protection{Items: canonicalLinkProtection()}
				}
				descriptor.Protection.Members = map[string]*clioutput.Protection{"ownedLinks": {Members: groups}}
			}
		}
		return descriptor
	}

	if _, isLink := record.(bdpwire.LinkRecord); isLink {
		descriptor.Protection = canonicalLinkProtection()
	}
	if bead, ok := record.(bdpwire.BeadRecord); ok {
		groups := make(map[string]*clioutput.Protection, len(bead.OwnedLinks))
		for typ := range bead.OwnedLinks {
			groups[typ] = &clioutput.Protection{Items: canonicalLinkProtection()}
		}
		descriptor.Protection.Members = map[string]*clioutput.Protection{"ownedLinks": {Members: groups}}
	}
	return descriptor
}

func runCanonicalShow(cmd *cobra.Command, args []string, linkOnly bool) error {
	opts, selected, err := canonicalOutputOptions(cmd)
	definitionWorkspace := graphPreviewConfig != nil && graphPreviewConfig.GraphSchemaVersion == graphstore.TypesFreshSchemaVersion
	// This fresh layout adopts canonical Show as its ordinary read. Preserve the
	// familiar --json spelling here without changing legacy layout 11 output.
	if err == nil && definitionWorkspace && !selected {
		if cmd.Flags().Lookup("json") != nil {
			opts.JSON, _ = cmd.Flags().GetBool("json")
		} else if cmd.Root().PersistentFlags().Lookup("json") != nil {
			opts.JSON, _ = cmd.Root().PersistentFlags().GetBool("json")
		}
	}
	if err != nil {
		return canonicalFailure(cmd, canonicalFailureRendererFor(cmd), "invalid_selector", err.Error(), 2)
	}
	// Validate even when the selector subsequently fails, so JSON failures can
	// carry the chosen contract; ambiguous formats never guess an output mode.
	baseDescriptor := canonicalResourceDescriptor()
	if definitionWorkspace {
		baseDescriptor.BDPSpec = intrinsicDefinitionContract
	}
	renderer, err := clioutput.New(baseDescriptor, opts)
	if err != nil {
		return canonicalFailure(cmd, canonicalFailureRendererFor(cmd), "invalid_selector", err.Error(), 2)
	}
	literalFlag := false
	ids, _ := cmd.Flags().GetStringArray("id")
	if len(ids) != 0 {
		if len(args) != 0 || len(ids) != 1 {
			return canonicalFailure(cmd, renderer, "invalid_selector", "canonical show requires one positional selector or one literal --id", 2)
		}
		args = ids
		literalFlag = true
	}
	if len(args) != 1 {
		return canonicalFailure(cmd, renderer, "invalid_selector", "canonical show requires exactly one Resource identity", 2)
	}
	allowed := []string{"id", "revision"}
	if definitionWorkspace && !selected {
		allowed = append(allowed, "json")
	}
	if err := canonicalGraphFlags(cmd, renderer, allowed...); err != nil {
		return err
	}
	if err := canonicalScalarOccurrences(cmd, os.Args[1:], "revision"); err != nil {
		return canonicalFailure(cmd, renderer, "invalid_selector", err.Error(), 2)
	}
	var literal string
	if literalFlag {
		literal, err = graphPreviewResolveLiteralArgument(args[0])
	} else {
		literal, err = graphPreviewResolveMemoryArgument(args[0])
	}
	if err != nil {
		return err
	}
	ref, err := graphPreviewParseReference(literal)
	if err != nil {
		if graphPreviewReferenceStorageError(err) {
			return canonicalReadError(cmd, renderer, err, "route_unavailable")
		}
		return canonicalFailure(cmd, renderer, "invalid_selector", err.Error(), 2)
	}
	if ref.Explicit && cmd.Flags().Changed("revision") {
		return canonicalFailure(cmd, renderer, "invalid_selector", "choose ID@revision or --revision, not both", 2)
	}
	ref, err = graphPreviewResolveParsedReference(ref)
	if err != nil {
		if graphPreviewReferenceStorageError(err) {
			return canonicalReadError(cmd, renderer, err, "route_unavailable")
		}
		return canonicalFailure(cmd, renderer, "invalid_selector", err.Error(), 2)
	}
	path, localID, err := graphPreviewResourceReadSelector(ref.Identity)
	if err != nil {
		return canonicalFailure(cmd, renderer, "invalid_selector", err.Error(), 2)
	}
	if linkOnly && localID == "" && !strings.HasPrefix(path, graphPreviewConfig.GraphScopeURL) && !strings.HasPrefix(path, "links/") {
		return canonicalFailure(cmd, renderer, "invalid_selector", "link show requires a Link Resource", 2)
	}
	version, _ := cmd.Flags().GetString("revision")
	versioned := cmd.Flags().Changed("revision")
	if ref.Explicit && versioned {
		return canonicalFailure(cmd, renderer, "invalid_selector", "choose ID@revision or --revision, not both", 2)
	}
	if versioned {
		version, err = graphPreviewRevisionValue("revision", version, ref.Identity, false)
		if err != nil {
			if graphPreviewReferenceStorageError(err) {
				return canonicalReadError(cmd, renderer, err, "route_unavailable")
			}
			return canonicalFailure(cmd, renderer, "invalid_selector", err.Error(), 2)
		}
	}
	if ref.Explicit {
		if !ref.currentSelection() {
			version, versioned = ref.Revision, true
		}
	}
	if versioned && (version == "" || !utf8.ValidString(version) || len(version) > graphstore.PreviewVersionTokenLimit) {
		return canonicalFailure(cmd, renderer, "invalid_selector", "--revision requires a nonempty UTF-8 token of at most 4096 bytes", 2)
	}
	if versioned && localID != "" {
		return canonicalFailure(cmd, renderer, "capability_unavailable", "retained versions require an exact canonical Scope URL", 5)
	}
	renderer, err = clioutput.New(baseDescriptor, opts)
	if err != nil {
		return canonicalFailure(cmd, canonicalFailureRendererFor(cmd), "invalid_selector", err.Error(), 2)
	}
	return canonicalReadWithGraphStore(cmd, renderer, func(ctx context.Context, store *graphstore.Store) (any, error) {
		reader := graphread.New(store)
		var result any
		if versioned {
			result, err = reader.ResourceVersion(ctx, path, version)
		} else {
			path, result, err = reader.ResourceSelected(ctx, path, localID)
		}
		if err != nil {
			return nil, err
		}
		if linkOnly {
			if _, ok := result.(bdpwire.LinkRecord); !ok {
				doc, isDefinition := result.(resourceget.DefinitionDocument)
				if !isDefinition || doc.Kind() != resourceget.DefinitionLink {
					return nil, graphstore.ErrCapabilityUnavailable
				}
			}
		}
		return result, nil
	}, func(result any) error {
		descriptor := canonicalShowDescriptor(path, result)
		renderer, err := clioutput.New(descriptor, opts)
		if err != nil {
			return canonicalFailure(cmd, canonicalFailureRendererFor(cmd), "invalid_selector", err.Error(), 2)
		}
		return renderCanonicalResult(cmd, renderer, result)
	})
}
