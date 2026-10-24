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
	"github.com/steveyegge/beads/internal/types"
)

// A disposable CLI input budget, advertised by status --graph. It is not a
// negotiated BDP property limit and does not establish the production contract.
const graphPreviewPropertiesLimit = 1 << 20

func graphPreviewProperties(input string, stdin io.Reader) (map[string]any, error) {
	canonical, err := graphPreviewPropertiesJSON(input, stdin)
	if err != nil {
		return nil, err
	}
	var properties map[string]any
	if err := json.Unmarshal(canonical, &properties); err != nil {
		return nil, err
	}
	return properties, nil
}

func graphPreviewPropertiesJSON(input string, stdin io.Reader) (json.RawMessage, error) {
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
		return nil, fmt.Errorf("Resource properties input exceeds the preview limit of %d bytes", graphPreviewPropertiesLimit)
	}
	// Validate before decoding: encoding/json by itself silently accepts duplicate
	// members and repairs invalid Unicode. The existing graph admission rejects both.
	canonical, err := graph.CanonicalizeJSON(raw)
	if err != nil {
		return nil, err
	}
	if len(canonical) == 0 || canonical[0] != '{' {
		return nil, fmt.Errorf("Resource properties must be a JSON object")
	}
	return canonical, nil
}

func graphPreviewRevisionGuard(cmd *cobra.Command, source bool) (string, bool, error) {
	revisionFlag := "if-revision"
	if source {
		revisionFlag = "if-source-revision"
	}
	if !cmd.Flags().Changed(revisionFlag) {
		// A missing comparison accepts the current predecessor. Keep the
		// writer's explicit choice internal to the CLI adapter.
		return "", true, nil
	}
	revision, _ := cmd.Flags().GetString(revisionFlag)
	if revision == "" {
		return "", false, graphFailure("invalid_selector", "--"+revisionFlag+" requires a nonempty revision token", 2)
	}
	return revision, false, nil
}

// Ordinary edits use the current predecessor when no comparison was asked for.
// The writer still checks a supplied token atomically and records the actual
// replaced version. Destructive commands retain their separate explicit guard.
func graphPreviewEditRevisionGuard(cmd *cobra.Command) (string, bool, error) {
	if !cmd.Flags().Changed("if-revision") {
		return "", true, nil
	}
	revision, _ := cmd.Flags().GetString("if-revision")
	if revision == "" {
		return "", false, graphFailure("invalid_selector", "--if-revision requires a nonempty revision token", 2)
	}
	return revision, false, nil
}

func registerGraphLinkTypeFlag(cmd *cobra.Command) {
	cmd.Flags().String("type", "", "Installed Link Type or built-in Dependency type (graph preview only)")
}

func graphPreviewLinkTypeChanged(cmd *cobra.Command) bool {
	return cmd.Flags().Changed("type")
}

func graphPreviewNominalTypeFlagChanged(cmd *cobra.Command) bool {
	if !cmd.Flags().Changed("type") {
		return false
	}
	selector, _ := cmd.Flags().GetString("type")
	return graphPreviewNominalSelector(selector)
}

func graphPreviewNominalSelector(selector string) bool {
	// Native Issue/Dependency classifications keep their bare names. Nominal
	// Types use types/NAME or a full URI, including the reserved-name refusal.
	return strings.HasPrefix(selector, "types/") || strings.Contains(selector, "://") ||
		graphPreviewReservedTypeSelector(selector)
}

func graphPreviewTypeURL(scope, selector string) (string, error) {
	if graphPreviewReservedTypeSelector(selector) {
		return "", fmt.Errorf("intrinsic Type is not a public selector")
	}
	if selector != "" && !strings.Contains(selector, "/") {
		selector = scope + "types/" + selector
	} else if strings.HasPrefix(selector, "types/") {
		selector = scope + selector
	}
	prefix := scope + "types/"
	if !strings.HasPrefix(selector, prefix) {
		return "", fmt.Errorf("Type must be NAME, types/NAME, or a canonical Type URL in this Scope")
	}
	for _, segment := range strings.Split(strings.TrimPrefix(selector, prefix), "/") {
		if err := graph.ValidateCanonicalSegment(segment); err != nil {
			return "", fmt.Errorf("invalid local Type path: %w", err)
		}
	}
	if err := graph.ValidateTypeURL(selector); err != nil {
		return "", err
	}
	return selector, nil
}

func graphPreviewLinkType(cmd *cobra.Command) (string, error) {
	if !graphPreviewLinkTypeChanged(cmd) {
		return "", nil
	}
	selector, _ := cmd.Flags().GetString("type")
	scope := graphPreviewConfig.GraphScopeURL
	if !graphPreviewNominalSelector(selector) {
		if kind := canonicalDependencyType(types.DependencyType(selector)); kind.IsWellKnown() {
			return graphstore.DependencyTypeURL(scope), nil
		}
		return "", graphFailure("capability_unavailable", "Link type is not registered in this preview", 5)
	}
	typ, err := graphPreviewTypeURL(scope, selector)
	if err != nil {
		return "", graphFailure("invalid_selector", err.Error(), 2)
	}
	if !graphPreviewSelectedTypeCandidate(scope, typ) {
		return "", graphFailure("capability_unavailable", "Link Type is not installed in this workspace", 5)
	}
	return typ, nil
}

func runGraphPreviewLink(cmd *cobra.Command, args []string) error {
	typ, err := graphPreviewLinkType(cmd)
	if err != nil {
		return err
	}
	if len(args) != 2 {
		return graphFailure("invalid_selector", "graph Link creation requires two public Bead IDs or canonical Bead URLs", 2)
	}
	resolved := make([]string, 2)
	pins := make([]graphstore.EndpointPin, 2)
	for i, endpoint := range []struct{ name, value string }{{"source-revision", args[0]}, {"target-revision", args[1]}} {
		flag, _ := cmd.Flags().GetString(endpoint.name)
		identity, err := graphPreviewResolveMemoryArgument(endpoint.value)
		if err != nil {
			return err
		}
		ref, err := graphPreviewReferenceWithFlag(identity, flag, cmd.Flags().Changed(endpoint.name))
		if err != nil {
			return graphFailure("invalid_selector", err.Error(), 2)
		}
		pins[i] = graphPreviewEndpointPin(ref)
		resolved[i] = ref.Identity
	}
	args = resolved
	if _, native := graphPreviewNativeDependencyType(graphPreviewConfig.GraphScopeURL, typ); typ == "" || native {
		if pins[0] != (graphstore.EndpointPin{}) || pins[1] != (graphstore.EndpointPin{}) {
			return graphFailure("capability_unavailable", "native Dependency endpoints must remain floating", 5)
		}
		return runGraphPreviewAddDependency(cmd, args)
	}
	if err := graphPreviewWritePolicy(); err != nil {
		return err
	}
	if err := graphPreviewFlags(cmd, "type", "source", "target", "source-revision", "target-revision", "id", "properties", "metadata", "if-source-revision"); err != nil {
		return err
	}
	if len(args) != 2 {
		return graphFailure("invalid_selector", "graph Link creation requires two public Bead IDs or canonical Bead URLs", 2)
	}
	paths := make([]string, 2)
	localIDs := make([]string, 2)
	for i, selector := range args {
		path, localID, err := graphPreviewResourceReadSelector(selector)
		if err != nil {
			return graphFailure("invalid_selector", err.Error(), 2)
		}
		paths[i], localIDs[i] = path, localID
	}
	publicID, _ := cmd.Flags().GetString("id")
	if cmd.Flags().Changed("id") {
		if err := graph.ValidateCanonicalResourceID(publicID); err != nil {
			return graphFailure("invalid_selector", err.Error(), 2)
		}
	}
	revision, unconditional, err := graphPreviewRevisionGuard(cmd, true)
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
	metadata, err := graphPreviewMetadataCreate(cmd)
	if err != nil {
		return err
	}
	return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		if err := graphPreviewRequireInstalledType(ctx, store, typ, graph.KindLink, false); err != nil {
			return nil, "", err
		}
		result, err := store.AddInformationalLink(ctx, graphstore.LinkCreateRequest{
			PublicID: publicID, TypeURL: typ, SourcePath: paths[0], TargetPath: paths[1], SourceLocalID: localIDs[0], TargetLocalID: localIDs[1], SourcePin: pins[0], TargetPin: pins[1], Properties: properties, Metadata: metadata,
			Actor: getActorWithGit(), ExpectedSourceRevision: revision, UnconditionalSource: unconditional,
		})
		return graphPreviewReplacementResult(result, fmt.Sprintf("Created %s: %s → %s", result.Link.ID, args[0], args[1]), result.ReplacedSource, err)
	})
}

func runGraphPreviewUpdateLink(cmd *cobra.Command, args []string) error {
	if err := graphPreviewWritePolicy(); err != nil {
		return err
	}
	if err := graphPreviewFlags(cmd, "properties", "source-revision", "target-revision", "metadata", "replace-metadata", "set-metadata", "unset-metadata", "if-revision", "if-source-revision"); err != nil {
		return err
	}
	if len(args) != 1 {
		return graphFailure("invalid_selector", "graph update requires one canonical links/PATH", 2)
	}
	path, err := graphPreviewResolvedHistoricalPath(args[0])
	if err != nil {
		return err
	}
	if err := graph.ValidateLinkPath(path); err != nil {
		return graphFailure("capability_unavailable", "Link updates require a canonical links/PATH identity", 5)
	}
	sourcePin, targetPin, err := graphPreviewLinkPinIntents(cmd)
	if err != nil {
		return err
	}
	if !cmd.Flags().Changed("properties") && sourcePin == nil && targetPin == nil {
		return graphFailure("invalid_properties", "Link update requires --properties or an endpoint revision flag", 2)
	}
	revision, unconditional, err := graphPreviewEditRevisionGuard(cmd)
	if err != nil {
		return err
	}
	sourceRevision, unconditionalSource, err := graphPreviewRevisionGuard(cmd, true)
	if err != nil {
		return err
	}
	var properties map[string]any
	if cmd.Flags().Changed("properties") {
		input, _ := cmd.Flags().GetString("properties")
		properties, err = graphPreviewProperties(input, cmd.InOrStdin())
		if err != nil {
			return graphFailure("invalid_properties", err.Error(), 2)
		}
	}
	metadata, err := graphPreviewMetadataPatch(cmd)
	if err != nil {
		return err
	}
	return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		result, err := store.UpdateLink(ctx, graphstore.LinkUpdateRequest{
			Path: path, Properties: properties, MergeProperties: cmd.Flags().Changed("properties"), SourcePin: sourcePin, TargetPin: targetPin, Metadata: metadata, Actor: getActorWithGit(), ExpectedRevision: revision,
			Unconditional: unconditional, ExpectedSourceRevision: sourceRevision, UnconditionalSource: unconditionalSource,
		})
		verb := "Updated"
		if !result.Changed {
			verb = "Unchanged"
		}
		return graphPreviewReplacementResult(result, fmt.Sprintf("%s %s", verb, result.Link.ID), result.ReplacedSource, err)
	})
}

func graphPreviewEndpointPin(ref graphPreviewReference) graphstore.EndpointPin {
	if !ref.Explicit || ref.Revision == "unpinned" {
		return graphstore.EndpointPin{}
	}
	if ref.Revision == "current" {
		return graphstore.EndpointPin{Current: true}
	}
	return graphstore.EndpointPin{Revision: ref.Revision}
}

func graphPreviewLinkPinFlagsChanged(cmd *cobra.Command) bool {
	return cmd.Flags().Changed("source-revision") || cmd.Flags().Changed("target-revision")
}

func graphPreviewLinkPinIntents(cmd *cobra.Command) (*graphstore.EndpointPin, *graphstore.EndpointPin, error) {
	var pins [2]*graphstore.EndpointPin
	for i, name := range []string{"source-revision", "target-revision"} {
		if !cmd.Flags().Changed(name) {
			continue
		}
		value, _ := cmd.Flags().GetString(name)
		if err := graphPreviewReferenceRevision(value); err != nil {
			return nil, nil, graphFailure("invalid_selector", err.Error(), 2)
		}
		pin := graphPreviewEndpointPin(graphPreviewReference{Revision: value, Explicit: true})
		pins[i] = &pin
	}
	return pins[0], pins[1], nil
}
