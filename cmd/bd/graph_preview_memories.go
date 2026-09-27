package main

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// Discovery projects one checked current snapshot, then releases storage before
// emitting any output. It never expands neighboring bodies or invents a cursor.
func runGraphPreviewMemories(cmd *cobra.Command, args []string) error {
	if err := graphPreviewFlags(cmd, "all", "details", "format"); err != nil {
		return err
	}
	if cmd.Flags().Changed("json") || jsonOutput {
		return graphFailure("capability_unavailable", "graph memories --json has no settled compatibility mapping; use --format records-json for experimental summaries", 5)
	}
	format, _ := cmd.Flags().GetString("format")
	if format != "table" && format != "records-json" {
		return graphFailure("capability_unavailable", "graph memories supports --format table or records-json; full records and legacy-json are unavailable", 5)
	}
	if len(args) > 1 {
		return graphFailure("invalid_selector", "memories accepts at most one literal search string", 2)
	}
	search := ""
	if len(args) == 1 {
		search = args[0]
	}
	if !utf8.ValidString(search) || len(search) > graphMemoryDiscoveryQueryLimit {
		return graphFailure("invalid_selector", "Memory search must be UTF-8 and at most 4096 bytes", 2)
	}
	all, _ := cmd.Flags().GetBool("all")
	details, _ := cmd.Flags().GetBool("details")
	return withGraphStoreOutput(func(ctx context.Context, s *graphstore.Store) (any, string, error) {
		snapshot, err := s.CurrentSnapshot(ctx)
		if err != nil {
			return nil, "", err
		}
		result, err := graphMemoryDiscovery(s.ScopeURL(), snapshot, search, all, details)
		if err != nil {
			return nil, "", err
		}
		output, err := renderGraphMemoryDiscovery(result, format == "records-json", quietFlag)
		return nil, output, err
	}, func(_ any, output string) error {
		_, err := fmt.Fprint(cmd.OutOrStdout(), output)
		return err
	})
}

func renderGraphMemoryDiscovery(result graphMemoryDiscoveryResult, structured, quiet bool) (string, error) {
	var human strings.Builder
	if !structured && !quiet {
		fmt.Fprintf(&human, "Memories (%d; complete summaries)\n", len(result.Items))
		for _, item := range result.Items {
			fmt.Fprintf(&human, "\n%q  %q\n", item.ID, item.Title)
			if len(item.MatchedFields) > 0 {
				fmt.Fprintf(&human, "  Matched: %s\n", strings.Join(item.MatchedFields, ", "))
			}
			if item.Excerpt != nil {
				fmt.Fprintf(&human, "  %s excerpt: %q (shortened: %t)\n", item.Excerpt.Field, item.Excerpt.Text, item.Excerpt.Truncated)
			}
			fmt.Fprintf(&human, "  Attribution: actor=%q status=%q recordedAt=%q\n", item.Attribution.Actor, item.Attribution.Status, item.Attribution.RecordedAt)
			if item.Details != nil {
				fmt.Fprintf(&human, "  Owned Links: %d\n", item.Details.OwnedLinkCount)
			}
			fmt.Fprintf(&human, "  Recall: bd recall %s --version %s\n", graphMemoryShellArg(item.ID), graphMemoryShellArg(item.Version))
		}
		human.WriteString("\nExperimental title/body search; keys and complete Memory fields are unavailable.\n")
	}
	var output bytes.Buffer
	if err := graphPrintTo(&output, result, strings.TrimSuffix(human.String(), "\n"), quiet, structured); err != nil {
		return "", err
	}
	if output.Len() > graphMemoryDiscoveryOutputLimit {
		return "", fmt.Errorf("%w: Memory discovery output exceeds %d bytes; narrow the search", graphstore.ErrLimitExceeded, graphMemoryDiscoveryOutputLimit)
	}
	return output.String(), nil
}

// Values are data, even when displayed as an explicit follow-up shell command.
func graphMemoryShellArg(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}
