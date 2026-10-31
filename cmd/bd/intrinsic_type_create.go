package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/clioutput"
	"github.com/steveyegge/beads/internal/configfile"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/resourceget"
)

var intrinsicTypeCmd = &cobra.Command{
	Use: "type", GroupID: "issues", Short: "Work with local Type definition Beads",
}

var intrinsicTypeCreateCmd = &cobra.Command{
	Use: "create --id ID", Short: "Create a local BeadType definition",
	Long: "Create a real locally owned BeadType definition in a fresh workspace initialized with --graph-mode link --graph-profile types. New publication2 workspaces omit propertiesSchema when no local constraints are supplied; inherited constraints still apply. Retained publication1 workspaces preserve their explicit open-object schema. Custom instance creation is not yet supported.",
	Args: cobra.NoArgs, SilenceErrors: true, SilenceUsage: true,
	Annotations: map[string]string{skipStoreAnnotation: "1", canonicalDefaultAnnotation: "1"},
	RunE:        runIntrinsicTypeCreate,
}

func init() {
	intrinsicTypeCreateCmd.Flags().String("id", "", "Whole Scope-local identity for the new Type definition")
	registerCanonicalOutputFlags(intrinsicTypeCreateCmd)
	canonicalFailureFactories[intrinsicTypeCreateCmd] = func(options clioutput.Options) (canonicalFailureRenderer, error) {
		descriptor := canonicalResourceDescriptor()
		descriptor.BDPSpec = intrinsicDefinitionContract
		return clioutput.New(descriptor, options)
	}
	intrinsicTypeCmd.AddCommand(intrinsicTypeCreateCmd)
	rootCmd.AddCommand(intrinsicTypeCmd)
}

func runIntrinsicTypeCreate(cmd *cobra.Command, _ []string) error {
	if !graphPreviewActive || graphPreviewConfig == nil || graphPreviewConfig.GraphSchemaVersion != graphstore.TypesFreshSchemaVersion {
		return graphFailure("capability_unavailable", "type create requires the fresh Types development workspace", 5)
	}
	opts, _, err := canonicalOutputOptions(cmd)
	if err != nil {
		return canonicalFailure(cmd, canonicalFailureRendererFor(cmd), "invalid_selector", err.Error(), 2)
	}
	descriptor := canonicalResourceDescriptor()
	descriptor.BDPSpec = intrinsicDefinitionContract
	renderer, err := clioutput.New(descriptor, opts)
	if err != nil {
		return canonicalFailure(cmd, nil, "invalid_selector", err.Error(), 2)
	}
	if err := canonicalGraphFlags(cmd, renderer, "id", "json"); err != nil {
		return err
	}
	if err := canonicalScalarOccurrences(cmd, os.Args[1:], "id"); err != nil {
		return canonicalFailure(cmd, renderer, "invalid_selector", err.Error(), 2)
	}
	id, _ := cmd.Flags().GetString("id")
	if id == "" {
		return canonicalFailure(cmd, renderer, "invalid_selector", "type create requires --id", 2)
	}
	if _, err := graph.CanonicalPublicResourceURL(graphPreviewConfig.GraphScopeURL, id); err != nil {
		return canonicalFailure(cmd, renderer, "invalid_selector", err.Error(), 2)
	}
	previous := graphFailureReporter
	graphFailureReporter = func(code, message string, status int) error {
		return canonicalFailure(cmd, renderer, code, message, status)
	}
	defer func() { graphFailureReporter = previous }()
	if err := graphPreviewWritePolicy(); err != nil {
		return err
	}
	mapError := func(err error) error {
		switch {
		case errors.Is(err, graphstore.ErrOutcomeUnknown):
			return canonicalFailure(cmd, renderer, "outcome_unknown", err.Error(), 5)
		case errors.Is(err, graphstore.ErrConflict), errors.Is(err, graphstore.ErrAlreadyExists):
			return canonicalFailure(cmd, renderer, "conflict", err.Error(), 4)
		case errors.Is(err, graph.ErrValidation):
			return canonicalFailure(cmd, renderer, "invalid_selector", err.Error(), 2)
		default:
			return canonicalReadError(cmd, renderer, err, "route_unavailable")
		}
	}
	return withGraphStoreErrorMapper(30*time.Second, func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		doc, err := store.CreateBeadType(ctx, id, getActorWithGit())
		return doc, "", err
	}, func(result any, _ string) error {
		doc := result.(resourceget.DefinitionDocument)
		r, err := clioutput.New(canonicalShowDescriptor(doc.ID(), doc), opts)
		if err != nil {
			return canonicalFailure(cmd, renderer, "output_error", err.Error(), 1)
		}
		data, err := r.Render(doc.NativeJSON())
		if err != nil {
			return canonicalFailure(cmd, renderer, "output_error", err.Error(), 1)
		}
		return writeCanonicalDocument(cmd.OutOrStdout(), data)
	}, mapError)
}

// Argument validation runs before graph admission. Only this explicit fresh
// layout may omit the legacy Issue title for the unsaved prototype. This is
// metadata discovery, not authority: normal admission still checks the marker,
// workspace binding and route before the real Type is read.
func intrinsicCreatePreviewInvocation(cmd *cobra.Command) bool {
	if !cmd.Flags().Changed("type") {
		return false
	}
	dir, err := graphCandidateDir(cmd)
	if err != nil || dir == "" {
		return false
	}
	cfg, err := configfile.LoadForDiscovery(dir)
	if err != nil || cfg == nil || cfg.GraphMode != "link" || !cfg.GraphReady || cfg.GraphSchemaVersion != graphstore.TypesFreshSchemaVersion {
		return false
	}
	marker, err := os.ReadFile(filepath.Join(dir, graphPreviewMarker)) // #nosec G304 -- fixed marker in the selected workspace
	return err == nil && graphPreviewGenerationSupported(marker)
}

// Donna explicitly selected a disposable stand-in for instance creation while
// packs and admission are redesigned. This preview reads the real exact Type;
// it neither allocates a Bead nor claims a persisted Resource or revision.
func runIntrinsicTypedBeadPreview(cmd *cobra.Command, args []string) error {
	if err := graphPreviewFlags(cmd, "type", "id", "title", "properties"); err != nil {
		return err
	}
	if err := canonicalScalarOccurrences(cmd, os.Args[1:], "type", "id", "title", "properties"); err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	if !cmd.Flags().Changed("type") {
		return graphFailure("invalid_selector", "the provisional create preview requires --type", 2)
	}
	selector, _ := cmd.Flags().GetString("type")
	if selector == "" {
		return graphFailure("invalid_selector", "--type requires a local definition identity", 2)
	}
	properties, _ := cmd.Flags().GetString("properties")
	if properties == "" {
		properties = "{}"
	}
	canonical, err := graph.CanonicalizeJSON([]byte(properties))
	if err != nil {
		return graphFailure("invalid_selector", err.Error(), 2)
	}
	var values map[string]json.RawMessage
	if json.Unmarshal(canonical, &values) != nil || values == nil {
		return graphFailure("invalid_selector", "prototype properties must be a JSON object", 2)
	}
	title, _ := cmd.Flags().GetString("title")
	if len(args) != 0 {
		if title != "" {
			return graphFailure("invalid_selector", "choose a positional title or --title", 2)
		}
		title = args[0]
	}
	if title != "" {
		if _, exists := values["title"]; exists {
			return graphFailure("invalid_selector", "title is already supplied in --properties", 2)
		}
		values["title"], _ = json.Marshal(title)
	}
	id, _ := cmd.Flags().GetString("id")
	proposedID := ""
	if id != "" {
		proposedID, err = graph.CanonicalPublicResourceURL(graphPreviewConfig.GraphScopeURL, id)
		if err != nil {
			return graphFailure("invalid_selector", err.Error(), 2)
		}
	}
	return withGraphStoreOutput(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		doc, err := store.GetDefinition(ctx, selector, "")
		if err != nil {
			return nil, "", err
		}
		if !intrinsicPreviewTypeSupported(doc.NativeJSON()) {
			return nil, "", fmt.Errorf("%w: provisional create accepts only the generated unconstrained BeadType for its recorded publication", graphstore.ErrCapabilityUnavailable)
		}
		preview := map[string]any{"prototype": true, "persisted": false, "type": map[string]string{"uri": doc.ID(), "revision": doc.Revision()}, "properties": values}
		if id != "" {
			preview["proposedId"] = proposedID
		}
		return preview, "UNSAVED PREVIEW — no Bead was created.\nIt will not appear in bd list.\nType: " + doc.ID() + "@" + doc.Revision(), nil
	}, func(value any, human string) error { return graphPrint(value, human, quietFlag) })
}

// GetDefinition has already checked the complete generated body against its
// recorded native context. This deliberately narrow preview check is not a
// generic Type/schema validator: unknown editions and authored constraints refuse.
// Publication1 keeps its old literal schema; publication2 keeps genuine absence.
func intrinsicPreviewTypeSupported(raw []byte) bool {
	var definition struct {
		Type       struct{ URI, Revision string } `json:"type"`
		Properties map[string]json.RawMessage     `json:"properties"`
	}
	if json.Unmarshal(raw, &definition) != nil || definition.Properties == nil || definition.Type.URI != "https://github.com/gastownhall/bdp/types/type-system/bead-type" {
		return false
	}
	switch definition.Type.Revision {
	case "bead-type-1":
		schema, err := graph.CanonicalizeJSON(definition.Properties["propertiesSchema"])
		return err == nil && string(schema) == `{"$schema":"https://json-schema.org/draft/2020-12/schema","type":"object"}`
	case "bead-type-2":
		return len(definition.Properties) == 0
	default:
		return false
	}
}
