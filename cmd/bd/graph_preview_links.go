package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// A disposable CLI input budget, advertised by status --graph. It is not a
// negotiated BDP property limit and does not establish the production contract.
const graphPreviewPropertiesLimit = 1 << 20

func graphPreviewProperties(input string, stdin io.Reader) (map[string]any, error) {
	var reader io.Reader = strings.NewReader(input)
	if strings.HasPrefix(input, "@") {
		if input == "@-" {
			reader = stdin
		} else {
			file, err := os.Open(strings.TrimPrefix(input, "@")) // #nosec G304 -- explicit operator-selected properties file
			if err != nil {
				return nil, err
			}
			defer func() { _ = file.Close() }()
			reader = file
		}
	}
	raw, err := io.ReadAll(io.LimitReader(reader, graphPreviewPropertiesLimit+1))
	if err != nil {
		return nil, err
	}
	if len(raw) > graphPreviewPropertiesLimit {
		return nil, fmt.Errorf("Link properties input exceeds the preview limit of %d bytes", graphPreviewPropertiesLimit)
	}
	// Validate before decoding: encoding/json by itself silently accepts duplicate
	// members and repairs invalid Unicode. The existing graph admission rejects both.
	canonical, err := graph.CanonicalizeJSON(raw)
	if err != nil {
		return nil, err
	}
	var properties map[string]any
	if err := json.Unmarshal(canonical, &properties); err != nil || properties == nil {
		return nil, fmt.Errorf("Link properties must be a JSON object")
	}
	return properties, nil
}

func graphPreviewRevisionGuard(cmd *cobra.Command, source, required bool) (string, bool, error) {
	revisionFlag, unconditionalFlag := "if-revision", "unconditional"
	if source {
		revisionFlag, unconditionalFlag = "if-source-revision", "unconditional-source"
	}
	hasRevision, hasUnconditional := cmd.Flags().Changed(revisionFlag), cmd.Flags().Changed(unconditionalFlag)
	if !hasRevision && !hasUnconditional && !required {
		return "", false, nil
	}
	revision, _ := cmd.Flags().GetString(revisionFlag)
	unconditional, _ := cmd.Flags().GetBool(unconditionalFlag)
	if hasRevision == hasUnconditional || (hasRevision && revision == "") || (hasUnconditional && !unconditional) {
		return "", false, graphFailure("invalid_selector", "choose either --"+revisionFlag+" REVISION or --"+unconditionalFlag, 2)
	}
	return revision, unconditional, nil
}

func runGraphPreviewLink(cmd *cobra.Command, args []string) error {
	typ, _ := cmd.Flags().GetString("resource-type")
	if typ != graphstore.RelatedTypeURL(graphPreviewConfig.GraphScopeURL) {
		return runGraphPreviewAddDependency(cmd, args)
	}
	if err := graphPreviewWritePolicy(); err != nil {
		return err
	}
	if err := graphPreviewFlags(cmd, "resource-type", "id", "properties", "if-source-revision", "unconditional-source"); err != nil {
		return err
	}
	if len(args) != 2 {
		return graphFailure("invalid_selector", "graph Link creation requires two canonical Bead selectors", 2)
	}
	paths := make([]string, 2)
	for i, selector := range args {
		path, err := graphPreviewResourcePath(graphPreviewConfig.GraphScopeURL, selector)
		if err != nil {
			return graphFailure("invalid_selector", err.Error(), 2)
		}
		if err := graph.ValidateBeadPath(path); err != nil {
			return graphFailure("invalid_selector", err.Error(), 2)
		}
		paths[i] = path
	}
	path, _ := cmd.Flags().GetString("id")
	if cmd.Flags().Changed("id") {
		if err := graph.ValidateLinkPath(path); err != nil {
			return graphFailure("invalid_selector", err.Error(), 2)
		}
	}
	revision, unconditional, err := graphPreviewRevisionGuard(cmd, true, false)
	if err != nil {
		return err
	}
	properties := map[string]any{}
	if cmd.Flags().Changed("properties") {
		input, _ := cmd.Flags().GetString("properties")
		properties, err = graphPreviewProperties(input, cmd.InOrStdin())
		if err != nil {
			return graphFailure("invalid_properties", err.Error(), 2)
		}
	}
	return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		result, err := store.AddInformationalLink(ctx, graphstore.LinkCreateRequest{
			Path: path, SourcePath: paths[0], TargetPath: paths[1], Properties: properties,
			Actor: getActorWithGit(), ExpectedSourceRevision: revision, UnconditionalSource: unconditional,
		})
		return result, fmt.Sprintf("Created %s: %s → %s", result.Link.ID, args[0], args[1]), err
	})
}

func runGraphPreviewUpdateLink(cmd *cobra.Command, args []string) error {
	if err := graphPreviewWritePolicy(); err != nil {
		return err
	}
	if err := graphPreviewFlags(cmd, "properties", "if-revision", "unconditional", "if-source-revision", "unconditional-source"); err != nil {
		return err
	}
	if len(args) != 1 {
		return graphFailure("invalid_selector", "graph update requires one canonical links/PATH", 2)
	}
	path, err := graphPreviewResourcePath(graphPreviewConfig.GraphScopeURL, args[0])
	if err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	if err := graph.ValidateLinkPath(path); err != nil {
		return graphFailure("capability_unavailable", "this preview updates informational Link properties only", 5)
	}
	if !cmd.Flags().Changed("properties") {
		return graphFailure("invalid_properties", "Link update requires --properties to explicitly replace the complete properties object", 2)
	}
	revision, unconditional, err := graphPreviewRevisionGuard(cmd, false, true)
	if err != nil {
		return err
	}
	sourceRevision, unconditionalSource, err := graphPreviewRevisionGuard(cmd, true, false)
	if err != nil {
		return err
	}
	input, _ := cmd.Flags().GetString("properties")
	properties, err := graphPreviewProperties(input, cmd.InOrStdin())
	if err != nil {
		return graphFailure("invalid_properties", err.Error(), 2)
	}
	return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		result, err := store.UpdateLink(ctx, graphstore.LinkUpdateRequest{
			Path: path, Properties: properties, Actor: getActorWithGit(), ExpectedRevision: revision,
			Unconditional: unconditional, ExpectedSourceRevision: sourceRevision, UnconditionalSource: unconditionalSource,
		})
		verb := "Updated"
		if !result.Changed {
			verb = "Unchanged"
		}
		return result, fmt.Sprintf("%s %s", verb, result.Link.ID), err
	})
}
