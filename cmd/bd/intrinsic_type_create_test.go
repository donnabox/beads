package main

import (
	"bytes"
	"encoding/json"
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

func TestIntrinsicPreviewRequiresExactGeneratedEdition(t *testing.T) {
	const beadType = "https://github.com/gastownhall/bdp/types/type-system/bead-type"
	const openObject = `{"propertiesSchema":{"type":"object","$schema":"https://json-schema.org/draft/2020-12/schema"}}`
	for _, tc := range []struct {
		name, uri, revision, properties string
		accepted                        bool
	}{
		{"retained-explicit", beadType, "bead-type-1", openObject, true},
		{"current-omitted", beadType, "bead-type-2", `{}`, true},
		{"old-cannot-omit", beadType, "bead-type-1", `{}`, false},
		{"old-null-is-not-schema", beadType, "bead-type-1", `{"propertiesSchema":null}`, false},
		{"current-null-is-not-omission", beadType, "bead-type-2", `{"propertiesSchema":null}`, false},
		{"current-explicit-not-generated", beadType, "bead-type-2", openObject, false},
		{"current-authored-not-generated", beadType, "bead-type-2", `{"abstract":true}`, false},
		{"current-resources-not-generated", beadType, "bead-type-2", `{"schemaResources":{}}`, false},
		{"unknown-edition", beadType, "bead-type-3", `{}`, false},
		{"foreign-same-basename", "https://foreign.example.test/bead-type", "bead-type-2", `{}`, false},
		{"wrong-kind", "https://github.com/gastownhall/bdp/types/type-system/link-type", "link-type-2", `{}`, false},
		{"missing-properties", beadType, "bead-type-2", `null`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := json.Marshal(map[string]any{
				"type":       map[string]string{"uri": tc.uri, "revision": tc.revision},
				"properties": json.RawMessage(tc.properties),
			})
			if err != nil {
				t.Fatal(err)
			}
			before := bytes.Clone(raw)
			if got := intrinsicPreviewTypeSupported(raw); got != tc.accepted {
				t.Fatalf("preview admission = %v, want %v", got, tc.accepted)
			}
			if !bytes.Equal(raw, before) {
				t.Fatal("preview inserted or rewrote descriptor values")
			}
		})
	}
	if intrinsicPreviewTypeSupported([]byte(`{"properties":`)) {
		t.Fatal("malformed descriptor admitted")
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
