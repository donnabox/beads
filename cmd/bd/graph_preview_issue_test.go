//go:build cgo

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
)

func TestGraphPreviewIssueCreateReopen(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			graphPolicyCLI(t, bd, work, home, nil, "", graphSetupInitArgs(t, engine, "https://example.invalid/issues/", "--prefix", "demo", "--skip-hooks", "--skip-agents")...)
			created := graphPolicyCLI(t, bd, work, home, nil, "", "create", "A real Issue", "--graph-path", "beads/work", "--description", "Preserve this body — 雪", "--type", "enhancement", "--priority", "1", "--labels", " demo , ,demo", "--label", "demo, other ", "--json")
			var envelope struct {
				Result struct {
					ID, Type, Revision, Version string
					Properties                  types.Issue
					Owned                       []json.RawMessage
				}
			}
			if err := json.Unmarshal([]byte(created), &envelope); err != nil {
				t.Fatal(err)
			}
			r := envelope.Result
			wantType, err := graphstore.NativeIssueTypeURL("https://example.invalid/issues/", "feature")
			if err != nil {
				t.Fatal(err)
			}
			if r.Type != wantType {
				t.Fatalf("concrete feature Type: got %q want %q", r.Type, wantType)
			}
			if r.ID != "https://example.invalid/issues/"+r.Properties.ID || r.ID == "https://example.invalid/issues/beads/work" || r.Type == "" || r.Revision == "" || r.Version != "" || strings.Contains(created, `"version":`) || r.Owned == nil {
				t.Fatalf("incomplete Issue graph record: %s", created)
			}
			if r.Properties.Title != "A real Issue" || r.Properties.Description != "Preserve this body — 雪" || r.Properties.IssueType != types.TypeFeature || r.Properties.Priority != 1 || !strings.HasPrefix(r.Properties.ID, "demo-") || !reflect.DeepEqual(r.Properties.Labels, []string{"demo", "other"}) {
				t.Fatalf("Issue content or configured ID prefix was lost: %+v", r.Properties)
			}
			for range 2 {
				shown := graphPolicyCLI(t, bd, work, home, nil, "", "show", r.ID, "--format", "graph-json", "--json")
				if shown != created {
					t.Fatalf("fresh process changed graph record\ncreate: %s\nshow: %s", created, shown)
				}
			}
			graphPolicyCLI(t, bd, work, home, nil, "identity_reserved", "remember", "collision", "--id", r.Properties.ID, "--title", "Memory", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "", "remember", "Coexisting Memory", "--id", "beads/plan", "--title", "Plan", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "identity_reserved", "create", "collision", "--id", "beads/plan", "--force", "--json")
			beforeRefusals := graphMemoryReadSnapshot(t, work)
			for _, flag := range []string{"--ephemeral", "--no-history"} {
				graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "create", "Unsupported", "--graph-path", "beads/refused", flag, "--json")
				graphPolicyCLI(t, bd, work, home, nil, "not_found", "show", "beads/refused", "--json")
			}
			graphPolicyCLI(t, bd, work, home, nil, "invalid_properties", "create", "Unsupported", "--graph-path", "beads/refused", "--deps=blocks:beads/plan", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "not_found", "show", "beads/refused", "--json")
			if !reflect.DeepEqual(beforeRefusals, graphMemoryReadSnapshot(t, work)) {
				t.Fatal("unsupported create flags changed captured state")
			}
			graphPolicyCLI(t, bd, work, home, []string{"BD_READONLY=true"}, "permission_denied", "create", "Read-only", "--graph-path", "beads/refused", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "not_found", "show", "beads/refused", "--json")
			if !reflect.DeepEqual(beforeRefusals, graphMemoryReadSnapshot(t, work)) {
				t.Fatal("read-only create changed captured state")
			}
			for _, class := range []string{"ephemeral", "unversioned", "invalid"} {
				code := "capability_unavailable"
				if class == "invalid" {
					code = "invalid_properties"
				}
				configPath := filepath.Join(work, ".beads", "config.yaml")
				writeFile(t, configPath, []byte("storage-class.task: "+class+"\n"))
				before := legacyUpgradeTreeDigest(t, work)
				graphPolicyCLI(t, bd, work, home, nil, code, "create", "Configured class", "--graph-path", "beads/refused", "--json")
				if after := legacyUpgradeTreeDigest(t, work); after != before {
					t.Fatal("unsupported storage class refusal changed workspace")
				}
				if err := os.Remove(configPath); err != nil {
					t.Fatal(err)
				}
				graphPolicyCLI(t, bd, work, home, nil, "not_found", "show", "beads/refused", "--json")
			}
			// Version 2 is explicitly disposable; this binary must not silently
			// adopt its marker or run a migration during an ordinary read.
			metadataPath := filepath.Join(work, ".beads", "metadata.json")
			metadata, err := os.ReadFile(metadataPath)
			if err != nil {
				t.Fatal(err)
			}
			var oldFormat map[string]any
			if err := json.Unmarshal(metadata, &oldFormat); err != nil {
				t.Fatal(err)
			}
			oldFormat["graph_schema_version"] = 2
			metadata, err = json.Marshal(oldFormat)
			if err != nil {
				t.Fatal(err)
			}
			writeFile(t, metadataPath, metadata)
			writeFile(t, filepath.Join(work, ".beads", graphPreviewMarker), []byte("link-preview-v2\n"))
			before := legacyUpgradeTreeDigest(t, work)
			graphPolicyCLI(t, bd, work, home, nil, "graph_not_initialized", "show", "beads/work", "--json")
			if after := legacyUpgradeTreeDigest(t, work); after != before {
				t.Fatal("old preview refusal changed workspace")
			}
		})
	}
}
