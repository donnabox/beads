package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"unicode/utf8"

	"github.com/spf13/cobra"
	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

type memoryLinkOperation struct{ name, family, description string }

var memoryLinkOperations = []memoryLinkOperation{
	{"reference", "references", "Record a reference to another Bead"},
	{"supersede", "superseded-by", "Retire an older Memory in favor of a successor"},
	{"requires", "requires", "Record a necessary Constraint"},
	{"defines", "defines", "Record an owned subordinate Constraint"},
	{"subject-to", "subject-to", "Declare a Constraint governing a Bead"},
}

// Input construction is separate from execution and shares generic Link parsing,
// metadata and guards. No facade allocates identities or bypasses Type checks.
func newMemoryLinkCommand(op memoryLinkOperation) *cobra.Command {
	cmd := &cobra.Command{Use: op.name + " [SOURCE TARGET]", Short: op.description,
		Args:        func(*cobra.Command, []string) error { return nil },
		RunE:        func(cmd *cobra.Command, args []string) error { return runMemoryLinkCommand(cmd, args, op) },
		Annotations: map[string]string{canonicalDefaultAnnotation: "1"},
	}
	cmd.Flags().String("from", "", "Source Bead identity")
	cmd.Flags().String("to", "", "Target Bead identity")
	for _, flag := range []string{"title", "body", "status"} {
		cmd.Flags().String(flag, "", "Optional Link "+flag+" property")
	}
	cmd.Flags().String("properties", "", "Link properties JSON object, @file or @-; cannot duplicate named properties")
	cmd.Flags().String("metadata", "", "Informational metadata JSON object")
	cmd.Flags().String("id", "", "Optional shared Link identity")
	cmd.Flags().String("source-revision", "", "Explicit source endpoint revision selector")
	cmd.Flags().String("target-revision", "", "Explicit target endpoint revision selector")
	cmd.Flags().String("if-source-revision", "", "Require this current owner revision atomically")
	registerCanonicalOutputFlags(cmd)
	cmd.Flags().Lookup("output-format").Usage = "Output: default, detail or json; --json is shorthand for json"
	canonicalFailureFactories[cmd] = newMemoryMutationFailureRenderer
	return cmd
}

func memoryLinkProperties(cmd *cobra.Command) (map[string]any, error) {
	properties := map[string]any{}
	if cmd.Flags().Changed("properties") {
		input, _ := cmd.Flags().GetString("properties")
		var err error
		raw, err := graphPreviewPropertiesJSON(input, cmd.InOrStdin())
		if err != nil {
			return nil, err
		}
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.UseNumber()
		if err := decoder.Decode(&properties); err != nil {
			return nil, err
		}
	}
	for _, name := range []string{"title", "body", "status"} {
		if !cmd.Flags().Changed(name) {
			continue
		}
		if _, exists := properties[name]; exists {
			return nil, fmt.Errorf("property %q supplied by both --%s and --properties", name, name)
		}
		value, _ := cmd.Flags().GetString(name)
		if !utf8.ValidString(value) {
			return nil, fmt.Errorf("--%s must be valid UTF-8", name)
		}
		properties[name] = value
	}
	return properties, nil
}

func memoryLinkRequest(cmd *cobra.Command, args []string, scope string, op memoryLinkOperation) (graphstore.LinkCreateRequest, error) {
	source, target, err := graphPreviewAuthoringEndpoints(cmd, args)
	if err != nil {
		return graphstore.LinkCreateRequest{}, fmt.Errorf("%s %w", op.name, err)
	}
	return memoryLinkRequestWithEndpoints(cmd, scope, op, source, target)
}

func memoryLinkRequestWithEndpoints(cmd *cobra.Command, scope string, op memoryLinkOperation, source, target string) (graphstore.LinkCreateRequest, error) {
	zero := graphstore.LinkCreateRequest{}
	result := graphstore.LinkCreateRequest{TypeURL: scope + op.family}
	endpoints := []string{source, target}
	paths := make([]string, 2)
	pins := make([]graphstore.EndpointPin, 2)
	localIDs := make([]string, 2)
	for i, name := range []string{"source-revision", "target-revision"} {
		flag, _ := cmd.Flags().GetString(name)
		ref, err := graphPreviewReferenceWithFlag(endpoints[i], flag, cmd.Flags().Changed(name))
		if err != nil {
			return zero, err
		}
		path, localID, err := graphPreviewResourceReadSelectorIn(scope, ref.Identity)
		if err != nil {
			return zero, err
		}
		if localID == "" && graph.ValidateBeadPath(path) != nil && !graphPreviewCanonicalResourceAddress(scope, path) {
			return zero, fmt.Errorf("relationship endpoint must be a Bead")
		}
		paths[i], localIDs[i] = path, localID
		pins[i] = graphPreviewEndpointPin(ref)
	}
	result.SourcePath, result.TargetPath = paths[0], paths[1]
	result.SourceLocalID, result.TargetLocalID = localIDs[0], localIDs[1]
	result.SourcePin, result.TargetPin = pins[0], pins[1]
	if cmd.Flags().Changed("id") {
		value, _ := cmd.Flags().GetString("id")
		if err := graph.ValidateCanonicalResourceID(value); err != nil {
			return zero, err
		}
		result.PublicID = value
	}
	revision, unconditional, err := graphPreviewRevisionGuard(cmd, true, endpoints[0], false)
	if err != nil {
		return zero, err
	}
	if !utf8.ValidString(revision) || len(revision) > graphstore.PreviewVersionTokenLimit {
		return zero, fmt.Errorf("--if-source-revision requires a UTF-8 token of at most %d bytes", graphstore.PreviewVersionTokenLimit)
	}
	result.ExpectedSourceRevision, result.UnconditionalSource = revision, unconditional
	return result, nil
}

func init() {
	for _, op := range memoryLinkOperations {
		memoryGraphCmd.AddCommand(newMemoryLinkCommand(op))
	}
}
