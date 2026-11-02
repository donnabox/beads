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

func TestTypesReadAdmissionRetainsWriterBoundary(t *testing.T) {
	for _, command := range []*cobra.Command{showCmd, listCmd, linkListCmd, countCmd, graphCmd, queryCmd, intrinsicTypeCreateCmd, createCmd} {
		if !graphPreviewTypesCommandSupported(command) {
			t.Fatalf("ordinary receiver blocked: %s", command.CommandPath())
		}
	}
	// A matching spelling is not the actual receiver, and adding common reads
	// must not expose unrelated writers/import or arbitrary future commands.
	for _, command := range []*cobra.Command{nil, updateCmd, rememberCmd, linkCreateCmd, linkUpdateCmd, linkDeleteCmd, graphUnlinkCmd, &cobra.Command{Use: "show"}, &cobra.Command{Use: "create"}} {
		if graphPreviewTypesCommandSupported(command) {
			t.Fatalf("unreviewed receiver admitted: %v", command)
		}
	}
}

// The receiving merge must preserve each existing initializer and refuse an
// unsupported combination before any workspace or database is created.
func TestFreshGraphProfileComposition(t *testing.T) {
	for _, tc := range []struct {
		name, mode, profile, atomic string
		want                        int
	}{
		{"default", "", "", "", graphstore.SchemaVersion},
		{"mixed", "link", "", "", graphstore.SchemaVersion},
		{"types", "link", "types", "", graphstore.TypesFreshSchemaVersion},
		{"batch", "link", "", "true", graphstore.AtomicBatchSchemaVersion},
		{"batch-disabled", "link", "", "false", graphstore.SchemaVersion},
		{"types-batch-disabled", "link", "types", "false", graphstore.TypesFreshSchemaVersion},
		{"separate-layouts", "link", "types", "true", 0},
		{"batch-no-link", "", "", "true", 0},
		{"batch-dependency", "dependency", "", "true", 0},
		{"unknown-types", "link", "atomic-batch", "true", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "init"}
			cmd.Flags().String("graph-profile", "", "")
			cmd.Flags().Bool("atomic-batch", false, "")
			for name, value := range map[string]string{"graph-profile": tc.profile, "atomic-batch": tc.atomic} {
				if value != "" {
					if err := cmd.Flags().Set(name, value); err != nil {
						t.Fatal(err)
					}
				}
			}
			got, err := selectedGraphPreviewInitSchemaVersion(cmd, tc.mode)
			if got != tc.want || (err != nil) != (tc.want == 0) {
				t.Fatalf("schema=%d error=%v; want schema=%d", got, err, tc.want)
			}
		})
	}
	for _, version := range []int{graphstore.SchemaVersion, graphstore.TypesFreshSchemaVersion, graphstore.AtomicBatchSchemaVersion} {
		if !canonicalGraphSchemaSupported(version) {
			t.Fatalf("lost supported layout %d", version)
		}
	}
	for _, version := range []int{0, -1, 12, 13, 16} {
		if canonicalGraphSchemaSupported(version) {
			t.Fatalf("invented layout %d", version)
		}
	}
}
