package main

import (
	"context"
	"encoding/json"
	"os"
	"unicode/utf8"

	"github.com/spf13/cobra"
	graph "github.com/steveyegge/beads/graphops"
	"github.com/steveyegge/beads/internal/clioutput"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

var graphCompareCmd = &cobra.Command{
	Use:   "compare RESOURCE --from TOKEN --to TOKEN",
	Short: "Compare two exact retained graph versions (experimental)",
	Long: `Compare complete retained preview properties and owned Links for one Resource.
An unprefixed selector means a Bead under beads/; use links/PATH for a Link.
The explicit from/to tokens select direction, not a chronology. Both versions must
be available. Output is experimental JSON (indented for humans); common metadata,
full Memory fields and public History are not implemented. Use bd diff for Dolt refs.`,
	Args:          cobra.ExactArgs(1),
	GroupID:       "views",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          runGraphPreviewCompare,
}

func init() {
	graphCompareCmd.Flags().String("from", "", "Required exact retained source version token")
	graphCompareCmd.Flags().String("to", "", "Required exact retained destination version token")
	rootCmd.AddCommand(graphCompareCmd)
}

func runGraphPreviewCompare(cmd *cobra.Command, args []string) error {
	options, canonical, err := canonicalOutputOptions(cmd)
	if err != nil {
		return canonicalFailure(cmd, nil, "invalid_selector", err.Error(), 2)
	}
	var renderer *clioutput.Renderer
	if canonical {
		renderer, err = newGraphComparisonOutput(options)
		if err != nil {
			return canonicalFailure(cmd, canonicalFailureRendererFor(cmd), "invalid_selector", err.Error(), 2)
		}
		if len(args) != 1 {
			return canonicalFailure(cmd, renderer, "invalid_selector", "canonical compare requires exactly one Resource identity", 2)
		}
		if err := canonicalScalarOccurrences(cmd, os.Args[1:], "from", "to"); err != nil {
			return canonicalFailure(cmd, renderer, "invalid_selector", err.Error(), 2)
		}
		err = canonicalGraphFlags(cmd, renderer, "from", "to")
	} else {
		err = graphPreviewFlags(cmd, "from", "to")
	}
	if err != nil {
		return err
	}
	invalidSelector := func(message string) error {
		if canonical {
			return canonicalFailure(cmd, renderer, "invalid_selector", message, 2)
		}
		return graphFailure("invalid_selector", message, 2)
	}
	path, err := graphPreviewResolvedHistoricalPath(args[0])
	if canonical && cmd.Parent() != nil && cmd.Parent().Name() == "link" {
		path, err = graphPreviewResolvedHistoricalPath(args[0])
		if err == nil {
			err = graph.ValidateLinkPath(path)
		}
	}
	if err != nil {
		return invalidSelector(err.Error())
	}
	from, _ := cmd.Flags().GetString("from")
	to, _ := cmd.Flags().GetString("to")
	for _, token := range []string{from, to} {
		if token == "" || !utf8.ValidString(token) || len(token) > graphstore.PreviewVersionTokenLimit {
			return invalidSelector("--from and --to require nonempty UTF-8 tokens of at most 4096 bytes each")
		}
	}
	if canonical {
		return canonicalReadWithGraphStore(cmd, renderer, func(ctx context.Context, store *graphstore.Store) (any, error) {
			result, err := store.CompareVersions(ctx, path, from, to)
			if err != nil {
				return nil, err
			}
			return graphProjectCompleteRecords(result)
		}, func(result any) error {
			return renderCanonicalResult(cmd, renderer, result)
		})
	}
	return withGraphStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		result, err := store.CompareVersions(ctx, path, from, to)
		if err != nil {
			return nil, "", err
		}
		projected, err := graphProjectCompleteRecords(result)
		if err != nil {
			return nil, "", err
		}
		human, err := json.MarshalIndent(projected, "", "  ")
		return result, string(human), err
	})
}
