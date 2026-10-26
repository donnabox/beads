package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/legacyimport"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func runGraphPreviewImport(cmd *cobra.Command, args []string) error {
	if err := graphPreviewWritePolicy(); err != nil {
		return err
	}
	if err := graphPreviewFlags(cmd, "input", "dry-run", "source-prefix", "id-root", "ignore-unknown-fields"); err != nil {
		return err
	}
	if cmd.Flags().Changed("id-root") && importIDRoot == "" {
		return graphFailure("invalid_properties", "--id-root requires a nonempty Scope-relative root", 2)
	}
	if cmd.Flags().Changed("source-prefix") && importSourcePrefix == "" {
		return graphFailure("invalid_properties", "--source-prefix requires a nonempty prefix", 2)
	}
	if len(args) > 1 {
		return graphFailure("invalid_selector", "import accepts one file or '-'", 2)
	}
	if err := runImportInner(args); err != nil {
		if _, ok := err.(*exitError); ok {
			return err
		}
		return graphFailure("invalid_properties", err.Error(), 2)
	}
	return nil
}

func runGraphPreviewImportReader(reader io.Reader, source string) error {
	batch, err := parseGraphLegacyInput(reader, importIgnoreUnknownFields, os.Stderr)
	if err != nil {
		var warningError *graphLegacyImportWarningError
		if errors.As(err, &warningError) {
			return graphFailure("output_error", warningError.Error(), 5)
		}
		if errors.Is(err, legacyimport.ErrLimitExceeded) {
			return graphStorageError(fmt.Errorf("%w: %w", graphstore.ErrLimitExceeded, err))
		}
		return graphFailure("invalid_properties", err.Error(), 2)
	}
	return withGraphStoreStageErrorMappers(5*time.Minute, func(ctx context.Context, s *graphstore.Store) (any, string, error) {
		result, err := s.ImportLegacyWithOptions(ctx, batch, getActorWithGit(), importDryRun, graphstore.LegacyImportOptions{SourcePrefix: importSourcePrefix, IDRoot: importIDRoot})
		if err != nil {
			return nil, "", err
		}
		action := "Imported"
		if importDryRun {
			action = "Would import"
		}
		return result, graphPreviewLegacyImportSummary(action, source, result), nil
	}, func(result any, human string) error {
		return graphPrint(result, human, quietFlag)
	}, graphLegacyImportOpenError, graphLegacyImportError)
}

// Warnings describe discarded source fields, not a successful storage operation.
// This seam stays independent of providers, so diagnostics are written before
// admission and cannot contaminate graphPrint's structured stdout.
func parseGraphLegacyInput(reader io.Reader, ignoreUnknown bool, warnings io.Writer) (legacyimport.Batch, error) {
	batch, report, err := legacyimport.ParseWithOptions(reader, legacyimport.Options{IgnoreUnknownFields: ignoreUnknown})
	if err != nil {
		return legacyimport.Batch{}, err
	}
	if report.Total == 0 {
		return batch, nil
	}
	if _, err := fmt.Fprintf(warnings, "Warning: legacy parsing discarded %s across %s; storage admission has not occurred.\n", pluralCount(report.Total, "unknown field", "unknown fields"), pluralCount(report.Distinct, "distinct path", "distinct paths")); err != nil { //nolint:gosec // G705: stderr diagnostics, not a browser; source values are excluded.
		return legacyimport.Batch{}, &graphLegacyImportWarningError{cause: err}
	}
	for _, field := range report.Fields {
		suffix := ""
		if field.Truncated {
			suffix = " (path component truncated; full-path digest " + field.PathDigest + ")"
		}
		if _, err := fmt.Fprintf(warnings, "  %s: %s%s\n", field.Path, pluralCount(field.Count, "occurrence", "occurrences"), suffix); err != nil { //nolint:gosec // G705: stderr only; path components are bounded and JSON-escaped.
			return legacyimport.Batch{}, &graphLegacyImportWarningError{cause: err}
		}
	}
	if report.OmittedDistinct != 0 {
		if _, err := fmt.Fprintf(warnings, "  %s omitted (%s).\n", pluralCount(report.OmittedDistinct, "additional distinct path", "additional distinct paths"), pluralCount(report.OmittedOccurrences, "occurrence", "occurrences")); err != nil { //nolint:gosec // G705: stderr-only numeric counts, not a browser context.
			return legacyimport.Batch{}, &graphLegacyImportWarningError{cause: err}
		}
	}
	return batch, nil
}

// A diagnostic write fails before provider admission. Preserve its cause while
// keeping it distinct from malformed input and storage outcome uncertainty.
type graphLegacyImportWarningError struct {
	cause error
}

func (e *graphLegacyImportWarningError) Error() string {
	return "cannot report ignored legacy fields; storage admission has not occurred: " + e.cause.Error()
}

func (e *graphLegacyImportWarningError) Unwrap() error { return e.cause }

// Import inspections use a bounded-operation refusal for cancellation. Keep
// uncertainty and already-reported errors ahead of this route-local mapping.
func graphLegacyImportError(err error) error {
	return graphLegacyImportStageError(err, "route_unavailable")
}

func graphLegacyImportOpenError(err error) error {
	return graphLegacyImportStageError(err, "graph_not_initialized")
}

func graphLegacyImportStageError(err error, fallback string) error {
	if errors.Is(err, graphstore.ErrOutcomeUnknown) {
		return graphStorageError(err)
	}
	var reported *exitError
	if errors.As(err, &reported) {
		return reported
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return graphFailure("capability_unavailable", err.Error(), 5)
	}
	return graphStorageErrorWithFallback(err, fallback)
}

func graphPreviewLegacyImportSummary(action, source string, result graphstore.LegacyImportResult) string {
	placement := ""
	if result.IDRoot != "" {
		placement = fmt.Sprintf(" ID root: %s. Workspace Issue prefix remains %s.", result.IDRoot, result.IssuePrefix)
		if result.SourceIssuePrefix != "" {
			placement += fmt.Sprintf(" Source Issue prefix: %s.", result.SourceIssuePrefix)
		}
	} else if result.SourceIssuePrefix != "" {
		if result.DryRun {
			placement = fmt.Sprintf(" Would adopt Issue prefix: %s.", result.IssuePrefix)
		} else {
			placement = fmt.Sprintf(" Workspace Issue prefix: %s.", result.IssuePrefix)
		}
	}
	provisional := ""
	if result.GeneratedIDsProvisional {
		provisional = " Generated Link IDs are provisional; dry-run reserves no IDs."
	}
	return fmt.Sprintf("%s %d Issues, %d Memories, %d Dependencies and %d comments from %s.%s%s Source history is not present in legacy exports.", action, result.Issues, result.Memories, result.Dependencies, result.Comments, source, placement, provisional)
}
