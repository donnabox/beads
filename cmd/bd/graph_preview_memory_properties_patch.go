package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/spf13/cobra"
	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/graphpatch"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// The update dispatcher checks write policy and canonical Scope selection first.
// Read the selected Type before dispatch. Issue patches pin and retry a
// checked predecessor; Memory patches evaluate inside their write transaction.
func runGraphPreviewMemoryPropertiesPatch(cmd *cobra.Command, path string) error {
	request, err := graphPreviewMemoryPropertiesPatchRequest(cmd, path)
	if err != nil {
		return err
	}
	issueUpdated := false
	err = withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		current, err := store.Read(ctx, path)
		if err != nil {
			return nil, "", err
		}
		if issue, ok := current.(graphstore.IssueRecord); ok {
			result, summary, err := graphPreviewPatchIssueProperties(ctx, store, cmd, request, issue)
			if err == nil {
				issueUpdated = true
			}
			return result, summary, err
		}
		if cmd.Flags().Changed("force") {
			return nil, "", graphFailure("invalid_properties", "--force applies only to an Issue notes patch", 2)
		}
		result, err := store.PatchMemoryProperties(ctx, request)
		if err != nil {
			return nil, "", err
		}
		verb := "Updated"
		if !result.Changed {
			verb = "Unchanged"
		}
		return graphPreviewReplacementResult(result, fmt.Sprintf("%s %s", verb, result.Memory.ID), result.Replaced, nil)
	})
	if err == nil && issueUpdated {
		SetLastTouchedID(path)
	}
	return err
}

func graphPreviewMemoryPropertiesPatchRequest(cmd *cobra.Command, path string) (graphstore.MemoryPropertiesPatchRequest, error) {
	var request graphstore.MemoryPropertiesPatchRequest
	if err := graphPreviewFlags(cmd, "patch", "replace-properties", "metadata", "replace-metadata", "set-metadata", "unset-metadata", "if-revision", "force"); err != nil {
		return request, err
	}
	if err := graph.ValidateBeadPath(path); err != nil {
		return request, graphFailure("invalid_selector", err.Error(), 2)
	}
	if !cmd.Flags().Changed("patch") && !cmd.Flags().Changed("replace-properties") {
		return request, graphFailure("invalid_properties", "property update requires --patch or --replace-properties", 2)
	}
	revision, unconditional, err := graphPreviewEditRevisionGuard(cmd)
	if err != nil {
		return request, err
	}
	if !utf8.ValidString(revision) || len(revision) > graphstore.PreviewVersionTokenLimit {
		return request, graphFailure("invalid_selector", fmt.Sprintf("--if-revision requires a UTF-8 token of at most %d bytes", graphstore.PreviewVersionTokenLimit), 2)
	}
	raw, err := graphPreviewPropertyOperations(cmd)
	if err != nil {
		return request, graphFailure("invalid_properties", err.Error(), 2)
	}
	metadata, err := graphPreviewMetadataPatch(cmd)
	if err != nil {
		return request, err
	}
	return graphstore.MemoryPropertiesPatchRequest{Path: path, Patch: raw, Actor: getActorWithGit(), ExpectedRevision: revision, Unconditional: unconditional, Metadata: metadata}, nil
}

// A full replacement is a single root operation evaluated against the checked
// predecessor by the existing Type writer. That writer retains validation,
// no-op suppression and Issue History ownership.
func graphPreviewPropertyOperations(cmd *cobra.Command) ([]byte, error) {
	if cmd.Flags().Changed("replace-properties") {
		input, _ := cmd.Flags().GetString("replace-properties")
		values, err := graphPreviewPropertiesJSON(input, cmd.InOrStdin())
		if err != nil {
			return nil, err
		}
		payload, err := json.Marshal([]struct {
			Op    string          `json:"op"`
			Path  string          `json:"path"`
			Value json.RawMessage `json:"value"`
		}{{Op: "replace", Path: "", Value: values}})
		if err != nil {
			return nil, err
		}
		if _, err := graphpatch.Parse(payload); err != nil {
			return nil, err
		}
		return payload, nil
	}
	input, _ := cmd.Flags().GetString("patch")
	return graphPreviewMemoryPropertiesPatchInput(input, cmd.InOrStdin())
}

// Keep operation-list bytes intact. The shared parser owns strict duplicate,
// Unicode, numeric and operation admission; encoding/json alone would repair or
// round rejected values. File/stdin acquisition is bounded before parsing.
func graphPreviewMemoryPropertiesPatchInput(input string, stdin io.Reader) ([]byte, error) {
	var reader io.Reader = strings.NewReader(input)
	var file *os.File
	if strings.HasPrefix(input, "@") {
		if input == "@-" {
			reader = stdin
		} else {
			var err error
			file, err = os.Open(strings.TrimPrefix(input, "@")) // #nosec G304 -- explicit operator-selected patch file
			if err != nil {
				return nil, fmt.Errorf("cannot open Resource patch file: %w", err)
			}
			reader = file
		}
	}
	raw, readErr := io.ReadAll(io.LimitReader(reader, graphpatch.MaxInputBytes+1))
	var closeErr error
	if file != nil {
		closeErr = file.Close()
	}
	if readErr != nil {
		return nil, fmt.Errorf("cannot read Resource patch: %w", readErr)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("cannot close Resource patch file: %w", closeErr)
	}
	if len(raw) > graphpatch.MaxInputBytes {
		return nil, fmt.Errorf("Resource patch input exceeds the preview limit of %d bytes", graphpatch.MaxInputBytes)
	}
	// Deliberate validation at both boundaries: syntax must refuse before the
	// CLI opens storage, while the raw-byte storage API independently validates
	// its own callers. Passing a trusted parsed object would weaken that boundary.
	if _, err := graphpatch.Parse(raw); err != nil {
		return nil, err
	}
	return raw, nil
}
