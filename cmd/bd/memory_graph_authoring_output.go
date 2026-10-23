package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/clioutput"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// Available effect facts are CLI problem context, never substitute Resources.
// No failed projection or delivery can turn a known applied write into rollback.
func memoryLinkEffect(state memoryLinkExecution) map[string]any {
	effect := map[string]any{
		"linkID":              state.result.Link.ID,
		"requestedSource":     state.request.SourcePath,
		"requestedTarget":     state.request.TargetPath,
		"requestedLink":       state.request.Path,
		"ifSourceRevision":    state.request.ExpectedSourceRevision,
		"unconditionalSource": state.request.UnconditionalSource,
	}
	if state.applied || state.result.Link.ID != "" {
		effect["changed"] = state.result.Changed
	}
	if state.applied && !errors.Is(state.err, graphstore.ErrOutcomeUnknown) {
		effect["applied"] = true
	}
	switch source := state.result.Source.(type) {
	case graphstore.Record:
		effect["sourceID"], effect["sourceRevision"] = source.ID, source.Revision
	case graphstore.IssueRecord:
		effect["sourceID"], effect["sourceRevision"] = source.ID, source.Revision
	}
	if replaced := state.result.ReplacedSource; replaced != nil {
		effect["replacedSourceID"], effect["replacedSourceRevision"] = replaced.ID, replaced.Version
	}
	return effect
}

func memoryLinkProblem(cmd *cobra.Command, renderer *clioutput.MutationRenderer, code, message string, status int, state memoryLinkExecution) error {
	problem := map[string]any{"code": code, "message": message, "retryable": false}
	if state.applied || code == "outcome_unknown" {
		problem["effect"] = memoryLinkEffect(state)
	}
	raw, err := json.Marshal(problem)
	if err != nil {
		return err
	}
	var data []byte
	switch {
	case code == "outcome_unknown":
		data, err = renderer.RenderUnknown("cli", raw)
	case state.applied && len(state.raw) > 0:
		data, err = renderer.RenderFailureWithResult("cli", raw, state.raw)
	default:
		data, err = renderer.RenderFailure("cli", raw)
	}
	if err != nil {
		// If an unavailable/invalid postimage cannot be framed, retain only the
		// truthful effect in the problem, never a malformed operation result.
		message += "; result framing failed: " + err.Error()
		problem["message"] = message
		raw, err = json.Marshal(problem)
		if err == nil {
			if code == "outcome_unknown" {
				data, err = renderer.RenderUnknown("cli", raw)
			} else {
				data, err = renderer.RenderFailure("cli", raw)
			}
		}
	}
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "output_error: %v; effect=%v\n", err, memoryLinkEffect(state))
		if code == "outcome_unknown" {
			return &exitError{Code: 6}
		}
		return &exitError{Code: 1}
	}
	fmt.Fprintf(cmd.ErrOrStderr(), "%s: %s", code, message) //nolint:gosec // G705: CLI stderr diagnostics are not HTML.
	if state.applied && code != "outcome_unknown" {
		fmt.Fprintf(cmd.ErrOrStderr(), "; operation completed; inspect before retry; effect=%v", memoryLinkEffect(state))
	}
	fmt.Fprintln(cmd.ErrOrStderr())
	// Keep the original refusal/effect diagnostic even if stdout delivery fails.
	if len(data) > 0 {
		if err := writeMemoryLinkDocument(cmd, data, state, true); err != nil {
			return err
		}
	}
	return &exitError{Code: status}
}

// Exactly one final publication attempt. A short/failed writer may already have
// delivered a prefix, so appending another envelope would corrupt the stream.
func writeMemoryLinkDocument(cmd *cobra.Command, data []byte, state memoryLinkExecution, diagnosisPrinted bool) error {
	if err := writeCanonicalDocument(cmd.OutOrStdout(), data); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "output_error: %v", err)
		if !diagnosisPrinted && state.applied && !errors.Is(state.err, graphstore.ErrOutcomeUnknown) {
			fmt.Fprintf(cmd.ErrOrStderr(), "; operation completed; do not replay automatically; inspect effect=%v", memoryLinkEffect(state))
		}
		fmt.Fprintln(cmd.ErrOrStderr())
		if errors.Is(state.err, graphstore.ErrOutcomeUnknown) {
			if !diagnosisPrinted {
				fmt.Fprintln(cmd.ErrOrStderr(), "outcome_unknown: inspect before retry; do not automatically replay")
			}
			return &exitError{Code: 6}
		}
		return &exitError{Code: 1}
	}
	return nil
}
