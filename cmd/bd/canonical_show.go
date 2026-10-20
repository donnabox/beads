package main

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/clioutput"
	"github.com/steveyegge/beads/internal/httpapi/bdpwire"
	"github.com/steveyegge/beads/internal/httpapi/graphread"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func init() {
	registerCanonicalOutputFlags(showCmd)
	registerCanonicalOutputFlags(linkShowCmd)
}

func canonicalLinkProtection() *clioutput.Protection {
	return &clioutput.Protection{Required: []string{"id", "type", "revision", "source", "target"}}
}

func canonicalShowDescriptor(path string, record any) clioutput.Descriptor {
	descriptor := canonicalResourceDescriptor()
	if strings.HasPrefix(path, "links/") {
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
	if len(args) != 1 {
		return canonicalFailure(cmd, renderer, "invalid_selector", "canonical show requires exactly one Resource identity", 2)
	}
	if err := canonicalGraphFlags(cmd, renderer, "version"); err != nil {
		return err
	}
	ref, err := graphPreviewParseReference(args[0])
	if err != nil {
		return canonicalFailure(cmd, renderer, "invalid_selector", err.Error(), 2)
	}
	var path string
	if linkOnly {
		path, err = graphPreviewLinkPath(graphPreviewConfig.GraphScopeURL, ref.Identity)
	} else {
		path, err = graphPreviewResourcePath(graphPreviewConfig.GraphScopeURL, ref.Identity)
	}
	if err != nil {
		return canonicalFailure(cmd, renderer, "invalid_selector", err.Error(), 2)
	}
	version, _ := cmd.Flags().GetString("version")
	versioned := cmd.Flags().Changed("version")
	if ref.Explicit {
		if versioned {
			return canonicalFailure(cmd, renderer, "invalid_selector", "choose ID@revision or --version, not both", 2)
		}
		if ref.Revision != "current" && ref.Revision != "unpinned" {
			version, versioned = ref.Revision, true
		}
	}
	if versioned && (version == "" || !utf8.ValidString(version) || len(version) > graphstore.PreviewVersionTokenLimit) {
		return canonicalFailure(cmd, renderer, "invalid_selector", "--version requires a nonempty UTF-8 token of at most 4096 bytes", 2)
	}
	renderer, err = clioutput.New(canonicalShowDescriptor(path, nil), opts)
	if err != nil {
		return canonicalFailure(cmd, canonicalFailureRendererFor(cmd), "invalid_selector", err.Error(), 2)
	}
	return canonicalReadWithGraphStore(cmd, renderer, func(ctx context.Context, store *graphstore.Store) (any, error) {
		reader := graphread.New(store)
		if versioned {
			return reader.ResourceVersion(ctx, path, version)
		}
		return reader.Resource(ctx, path)
	}, func(result any) error {
		descriptor := canonicalShowDescriptor(path, result)
		renderer, err := clioutput.New(descriptor, opts)
		if err != nil {
			return canonicalFailure(cmd, canonicalFailureRendererFor(cmd), "invalid_selector", err.Error(), 2)
		}
		return renderCanonicalResult(cmd, renderer, result)
	})
}
