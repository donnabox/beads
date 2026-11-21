package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
	"github.com/steveyegge/beads/internal/ui"
)

// Edit keeps the ordinary editor/tempfile interaction. The graph Issue read is
// closed before launching the editor, and its opaque revision protects the
// subsequent native Issue writer from replacing someone else's intervening edit.
func runGraphPreviewEdit(cmd *cobra.Command, args []string) error {
	if err := graphPreviewWritePolicy(); err != nil {
		return err
	}
	if err := graphPreviewFlags(cmd, "title", "description", "design", "notes", "acceptance"); err != nil {
		return err
	}
	field := "description"
	changed := 0
	for _, candidate := range []string{"title", "description", "design", "notes", "acceptance"} {
		if cmd.Flags().Changed(candidate) {
			field = candidate
			changed++
		}
	}
	if changed > 1 {
		return graphFailure("invalid_properties", "edit accepts one Issue field at a time", 2)
	}
	path, err := graphPreviewResourcePath(graphPreviewConfig.GraphScopeURL, args[0])
	if err != nil || !strings.HasPrefix(path, "beads/") {
		return graphFailure("invalid_selector", "edit requires an Issue Bead ID or beads/PATH", 2)
	}
	if field == "acceptance" {
		field = "acceptance_criteria"
	}
	var issue graphstore.IssueRecord
	err = withGraphStoreOutput(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		record, err := store.Read(ctx, path)
		if err != nil {
			return nil, "", err
		}
		var ok bool
		issue, ok = record.(graphstore.IssueRecord)
		if !ok {
			return nil, "", graphFailure("capability_unavailable", "edit requires an Issue", 5)
		}
		return nil, "", nil
	}, func(any, string) error { return nil })
	if err != nil {
		return err
	}
	current := map[string]string{
		"title": issue.Properties.Title, "description": issue.Properties.Description,
		"design": issue.Properties.Design, "notes": issue.Properties.Notes,
		"acceptance_criteria": issue.Properties.AcceptanceCriteria,
	}[field]

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		for _, fallback := range []string{"vim", "vi", "nano", "emacs"} {
			if _, err := exec.LookPath(fallback); err == nil {
				editor = fallback
				break
			}
		}
	}
	if editor == "" {
		return graphFailure("capability_unavailable", "no editor found; set $EDITOR or $VISUAL", 5)
	}
	parts := strings.Fields(editor)
	if len(parts) == 0 {
		return graphFailure("invalid_properties", "$EDITOR must name a program", 2)
	}
	tmp, err := os.CreateTemp("", fmt.Sprintf("bd-edit-%s-*.txt", field))
	if err != nil {
		return graphFailure("route_unavailable", "creating edit file: "+err.Error(), 5)
	}
	tmpPath := tmp.Name()
	saved := false
	defer func() {
		if saved {
			_ = os.Remove(tmpPath)
		}
	}()
	if _, err := tmp.WriteString(current); err != nil {
		_ = tmp.Close()
		return graphFailure("route_unavailable", "writing edit file: "+err.Error(), 5)
	}
	if err := tmp.Close(); err != nil {
		return graphFailure("route_unavailable", "closing edit file: "+err.Error(), 5)
	}
	process := exec.Command(parts[0], append(parts[1:], tmpPath)...) //nolint:gosec // editor is the user's selected command
	process.Env = envWithoutSharedHoldMarker()
	process.Stdin, process.Stdout, process.Stderr = cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr()
	if err := process.Run(); err != nil {
		return graphFailure("route_unavailable", fmt.Sprintf("running editor: %v; edit file is preserved in %s", err, tmpPath), 5)
	}
	// #nosec G304 -- tmpPath was created in this function.
	content, err := os.ReadFile(tmpPath)
	if err != nil {
		return graphFailure("route_unavailable", fmt.Sprintf("reading edited file: %v; edit file is preserved in %s", err, tmpPath), 5)
	}
	value := strings.TrimSpace(string(content))
	// Even an untouched file must pass the graph revision check. The writer
	// returns the current Issue without a History entry when the field is equal.
	unedited := value == current
	if field == "title" {
		if err := types.ValidateIssueTitle(value); err != nil {
			return graphFailure("invalid_properties", fmt.Sprintf("%v; edit file is preserved in %s", err, tmpPath), 2)
		}
	}
	request := graphstore.UpdateIssueRequest{Path: path, Actor: getActorWithGit(), ExpectedRevision: issue.Revision}
	switch field {
	case "title":
		request.Title = &value
	case "description":
		request.Description = &value
	case "design":
		request.Design = &value
	case "notes":
		request.Notes = &value
		request.ForceNotesOverwrite = true // choosing notes in an editor is explicit replacement intent
	case "acceptance_criteria":
		request.AcceptanceCriteria = &value
	}
	err = withGraphStoreOutput(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		result, err := store.UpdateIssue(ctx, request)
		if err != nil {
			return nil, "", fmt.Errorf("edit file is preserved in %s: %w", tmpPath, err)
		}
		if unedited && !result.Changed {
			return result, "No changes made", nil
		}
		verb := "Updated"
		if !result.Changed {
			verb = "Unchanged"
		}
		name := strings.ReplaceAll(field, "_", " ")
		return result, fmt.Sprintf("%s %s %s for issue: %s", ui.RenderPass("✓"), verb, name, formatFeedbackID(path, result.Issue.Properties.Title)), nil
	}, func(result any, human string) error { return graphPrint(result, human, quietFlag) })
	if err != nil {
		return err
	}
	saved = true
	SetLastTouchedID(path)
	return nil
}
