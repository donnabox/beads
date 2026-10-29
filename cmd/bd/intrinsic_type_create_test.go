package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/configfile"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func TestIntrinsicPreviewPreservesLegacyTitleAdmission(t *testing.T) {
	isolateBeadsDirForTest(t)
	dir := t.TempDir()
	t.Setenv("BEADS_DIR", dir)
	cmd := &cobra.Command{Use: "create"}
	for _, flag := range []string{"type", "title", "file", "graph"} {
		cmd.Flags().String(flag, "", "")
	}
	if err := cmd.Flags().Set("type", "foo"); err != nil {
		t.Fatal(err)
	}
	cfg := &configfile.Config{GraphMode: "link", GraphReady: true, GraphSchemaVersion: graphstore.TypesFreshSchemaVersion}
	for _, tc := range []struct {
		name    string
		version int
		marker  string
		preview bool
	}{
		{"fresh-explicit", graphstore.TypesFreshSchemaVersion, graphPreviewGeneration, true},
		{"legacy", graphstore.SchemaVersion, graphPreviewGeneration, false},
		{"bad-generation", graphstore.TypesFreshSchemaVersion, "unsupported", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg.GraphSchemaVersion = tc.version
			if err := cfg.Save(dir); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, graphPreviewMarker), []byte(tc.marker), 0o600); err != nil {
				t.Fatal(err)
			}
			err := validateCreateArgs(cmd, nil)
			if (err == nil) != tc.preview {
				t.Fatalf("title admission changed for %s: %v", tc.name, err)
			}
		})
	}
}

func TestIntrinsicTypeCreateRefusesOtherWorkspaces(t *testing.T) {
	oldActive, oldConfig, oldOutput := graphPreviewActive, graphPreviewConfig, graphFailureOutput
	t.Cleanup(func() { graphPreviewActive, graphPreviewConfig, graphFailureOutput = oldActive, oldConfig, oldOutput })
	graphPreviewActive, graphPreviewConfig = false, nil
	var diagnostics bytes.Buffer
	graphFailureOutput = &diagnostics
	err := runIntrinsicTypeCreate(intrinsicTypeCreateCmd, nil)
	var exit *exitError
	if !errors.As(err, &exit) || exit.Code != 5 || diagnostics.Len() == 0 {
		t.Fatalf("nonfresh writer admission: %v %s", err, diagnostics.String())
	}
}
