package main

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func TestGraphProfileSelectionPreservesDefaults(t *testing.T) {
	for _, tc := range []struct {
		name, mode, profile string
		explicit            bool
		want                int
	}{
		{"ordinary-default", "", "", false, graphstore.SchemaVersion},
		{"mixed-default", "link", "", false, graphstore.SchemaVersion},
		{"types", "link", "types", true, graphstore.TypesFreshSchemaVersion},
		{"no-link-mode", "", "types", true, 0},
		{"dependency-mode", "dependency", "types", true, 0},
		{"empty-explicit", "link", "", true, 0},
		{"unknown", "link", "atomic-batch", true, 0},
		{"numeric-layout", "link", "14", true, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "init"}
			cmd.Flags().String("graph-profile", "", "")
			if tc.explicit {
				if err := cmd.Flags().Set("graph-profile", tc.profile); err != nil {
					t.Fatal(err)
				}
			}
			actual, err := graphPreviewInitSchemaVersion(cmd, tc.mode)
			if actual != tc.want || (err != nil) != (tc.want == 0) {
				t.Fatalf("schema=%d error=%v; want schema=%d", actual, err, tc.want)
			}
		})
	}
}

func TestIntrinsicPreviewRejectsRepeatedScalars(t *testing.T) {
	for _, flag := range []string{"type", "id", "title", "properties"} {
		t.Run(flag, func(t *testing.T) {
			cmd := &cobra.Command{Use: "create"}
			for _, name := range []string{"type", "id", "title", "properties"} {
				cmd.Flags().String(name, "", "")
			}
			args := []string{"--type", "foo", "--" + flag, "first", "--" + flag + "=second"}
			if err := cmd.ParseFlags(args); err != nil {
				t.Fatal(err)
			}
			oldArgs, oldOutput, oldReporter := os.Args, graphFailureOutput, graphFailureReporter
			t.Cleanup(func() { os.Args, graphFailureOutput, graphFailureReporter = oldArgs, oldOutput, oldReporter })
			var diagnostics bytes.Buffer
			os.Args = append([]string{"bd", "create"}, args...)
			graphFailureOutput, graphFailureReporter = &diagnostics, nil
			err := runIntrinsicTypedBeadPreview(cmd, nil)
			var exit *exitError
			if !errors.As(err, &exit) || exit.Code != 2 || !strings.Contains(diagnostics.String(), flag) {
				t.Fatalf("ambiguous --%s reached storage or escaped refusal: %v %s", flag, err, diagnostics.String())
			}
		})
	}
}
