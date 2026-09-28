package main

import (
	"time"
	"unicode/utf8"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/timeparsing"
)

// Empty is absent on create and an explicit clear on update. The caller owns
// presence; storage owns the existing DATETIME representation, not this parser.
func graphPreviewIssueDueInput(cmd *cobra.Command) (*time.Time, error) {
	value, err := cmd.Flags().GetString("due")
	if err != nil {
		return nil, graphFailure("invalid_properties", err.Error(), 2)
	}
	if !utf8.ValidString(value) || len(value) > 4096 {
		return nil, graphFailure("invalid_properties", "--due must be UTF-8 and at most 4096 bytes", 2)
	}
	if value == "" {
		return nil, nil
	}
	parsed, err := timeparsing.ParseRelativeTime(value, time.Now())
	if err != nil {
		return nil, graphFailure("invalid_properties", err.Error(), 2)
	}
	return &parsed, nil
}
