package main

import (
	"context"
	"encoding/json"
	"strings"
	"unicode/utf8"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func graphPreviewRememberStatus(cmd *cobra.Command) (*string, error) {
	if !cmd.Flags().Changed("status") {
		return nil, nil
	}
	value, _ := cmd.Flags().GetString("status")
	if !utf8.ValidString(value) || strings.TrimSpace(value) == "" {
		return nil, graphFailure("invalid_properties", "Memory --status must be nonblank UTF-8", 2)
	}
	return &value, nil
}

// Typed creation is deliberately limited to an explicit new identity. A
// complete properties object is passed to the existing checked Memory writer;
// this command does not pre-read a predecessor or implement another upsert.
func runGraphPreviewRememberCreateOnly(cmd *cobra.Command, args []string, publicID string) error {
	// Reject independent scalar flag errors before opening --stdin or @-.
	typeURL, err := graphPreviewMemoryTypeFlag(cmd)
	if err != nil {
		return err
	}
	status, err := graphPreviewRememberStatus(cmd)
	if err != nil {
		return err
	}
	var explicitTitle string
	if cmd.Flags().Changed("title") {
		explicitTitle, _ = cmd.Flags().GetString("title")
		if !utf8.ValidString(explicitTitle) || strings.TrimSpace(explicitTitle) == "" {
			return graphFailure("invalid_properties", "an explicit creation title must be nonempty UTF-8", 2)
		}
	}
	properties := map[string]json.RawMessage{}
	if cmd.Flags().Changed("properties") {
		input, _ := cmd.Flags().GetString("properties")
		raw, err := graphPreviewPropertiesJSON(input, cmd.InOrStdin())
		if err != nil {
			return graphFailure("invalid_properties", err.Error(), 2)
		}
		if err := json.Unmarshal(raw, &properties); err != nil {
			return graphFailure("invalid_properties", err.Error(), 2)
		}
	}
	for _, name := range []string{"title", "body", "status"} {
		if _, present := properties[name]; present && cmd.Flags().Changed(name) {
			return graphFailure("invalid_properties", "Memory "+name+" was supplied by both --properties and --"+name, 2)
		}
	}
	if _, present := properties["body"]; present && (len(args) != 0 || cmd.Flags().Changed("body-file") || cmd.Flags().Changed("stdin")) {
		return graphFailure("invalid_properties", "Memory body was supplied by both --properties and another body source", 2)
	}
	var body string
	if raw, present := properties["body"]; present {
		var supplied *string
		if err := json.Unmarshal(raw, &supplied); err != nil || supplied == nil {
			return graphFailure("invalid_properties", "Memory body property must be a string", 2)
		}
		body = *supplied
		var err error
		body, err = graphPreviewReadMemoryBody(strings.NewReader(body))
		if err != nil {
			return err
		}
	} else {
		var err error
		body, err = graphPreviewRememberBody(cmd, args)
		if err != nil {
			return err
		}
	}
	title := graphPreviewMemoryTitle(body)
	if raw, present := properties["title"]; present {
		var supplied *string
		if err := json.Unmarshal(raw, &supplied); err != nil || supplied == nil {
			return graphFailure("invalid_properties", "Memory title property must be a string", 2)
		}
		title = *supplied
	} else if cmd.Flags().Changed("title") {
		title = explicitTitle
	}
	if !utf8.ValidString(title) || strings.TrimSpace(title) == "" {
		return graphFailure("invalid_properties", "an explicit creation title must be nonempty UTF-8", 2)
	}
	if raw, present := properties["status"]; present {
		var propertyStatus string
		if err := json.Unmarshal(raw, &propertyStatus); err != nil || !utf8.ValidString(propertyStatus) || strings.TrimSpace(propertyStatus) == "" {
			return graphFailure("invalid_properties", "Memory status property must be nonblank UTF-8", 2)
		}
	}
	properties["title"], _ = json.Marshal(title)
	properties["body"], _ = json.Marshal(body)
	if status != nil {
		properties["status"], _ = json.Marshal(*status)
	}
	complete, err := json.Marshal(properties)
	if err != nil {
		return graphFailure("invalid_properties", err.Error(), 2)
	}
	metadata, err := graphPreviewMetadataCreate(cmd)
	if err != nil {
		return err
	}
	return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		record, err := store.Create(ctx, graphstore.CreateRequest{
			PublicID: publicID, TypeURL: typeURL, PropertiesJSON: complete,
			Actor: getActorWithGit(), Metadata: metadata,
		})
		return record, graphHumanRow(store.ScopeURL(), record), err
	})
}
