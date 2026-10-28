package main

import (
	"context"
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

func canonicalShowDescriptor(_ string, record any) (clioutput.Descriptor, error) {
	descriptor := canonicalResourceDescriptor()
	// The public document owns its immutable native bytes. Decode only for the
	// existing renderer's endpoint/owned-Link protection, never to read again.
	if doc, ok := record.(resourceget.NativeDocument); ok {
		if doc.Kind() == resourceget.Link {
			descriptor.Protection = canonicalLinkProtection()
			return descriptor, nil
		}
		var bead bdpwire.CurrentBeadRecord
		if err := bdpwire.Unmarshal(doc.NativeJSON(), &bead); err != nil {
			return descriptor, err
		}
		groups := make(map[string]*clioutput.Protection, len(bead.OwnedLinks))
		for typ := range bead.OwnedLinks {
			groups[typ] = &clioutput.Protection{Items: canonicalLinkProtection()}
		}
		descriptor.Protection.Members = map[string]*clioutput.Protection{"ownedLinks": {Members: groups}}
		return descriptor, nil
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
	return descriptor, nil
}

func runCanonicalShow(cmd *cobra.Command, args []string, linkOnly bool) error {
	opts, _, err := canonicalOutputOptions(cmd)
	if err != nil {
		return canonicalFailure(cmd, canonicalFailureRendererFor(cmd), "invalid_selector", err.Error(), 2)
	}
	// Validate even when the selector subsequently fails, so JSON failures can
	// carry the chosen contract; ambiguous formats never guess an output mode.
	renderer, err := clioutput.New(canonicalResourceDescriptor(), opts)
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
	if err := canonicalGraphFlags(cmd, renderer, "id", "revision"); err != nil {
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
	renderer, err = clioutput.New(canonicalResourceDescriptor(), opts)
	if err != nil {
		return canonicalFailure(cmd, canonicalFailureRendererFor(cmd), "invalid_selector", err.Error(), 2)
	}
	return canonicalReadWithGraphStore(cmd, renderer, func(ctx context.Context, store *graphstore.Store) (any, error) {
		reader := graphread.New(store)
		if !versioned {
			kind := resourceget.Any
			if linkOnly {
				kind = resourceget.Link
			}
			return reader.Get(ctx, resourceget.Request{ID: ref.Identity, AssertKind: kind})
		}
		result, err := reader.ResourceVersion(ctx, path, version)
		if err != nil {
			return nil, err
		}
		if linkOnly {
			if _, ok := result.(bdpwire.LinkRecord); !ok {
				return nil, graphstore.ErrCapabilityUnavailable
			}
		}
		return result, nil
	}, func(result any) error {
		descriptor, err := canonicalShowDescriptor(path, result)
		if err != nil {
			return canonicalReadError(cmd, renderer, err, "route_unavailable")
		}
		renderer, err := clioutput.New(descriptor, opts)
		if err != nil {
			return canonicalFailure(cmd, canonicalFailureRendererFor(cmd), "invalid_selector", err.Error(), 2)
		}
		return renderCanonicalResult(cmd, renderer, result)
	})
}
