//go:build cgo

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/configfile"
)

func TestGraphPreviewInformationalCommandsWithoutEngine(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := graphListWorkspace(t, bd, engine, "https://example.invalid/information/")
			empty := t.TempDir()
			// Preserve an initialized graph workspace but make its engine unusable.
			// Informational commands must still complete through the no-store path.
			beadsDir := filepath.Join(work, ".beads")
			if engine == "embedded" {
				data := filepath.Join(beadsDir, "embeddeddolt")
				if err := os.Rename(data, data+".saved"); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(data, []byte("engine must not be opened\n"), 0600); err != nil {
					t.Fatal(err)
				}
			} else {
				cfg, err := configfile.LoadForDiscovery(beadsDir)
				if err != nil || cfg == nil {
					t.Fatalf("read initialized graph metadata: %v", err)
				}
				cfg.DoltServerPort = 1
				if err := cfg.Save(beadsDir); err != nil {
					t.Fatal(err)
				}
			}
			before := legacyUpgradeTreeDigest(t, work)
			var completeReference string
			for _, args := range [][]string{
				{}, {"--version"}, {"-V"}, {"--help"},
				{"version"}, {"version", "--json"}, {"version", "--format", "json"},
				{"version", "--readonly", "--no-color"}, {"version", "--graph-mode", "link"},
				{"help", "--all"}, {"help", "--list"}, {"help", "version"},
				{"help", "--doc", "version"}, {"help", "--doc", "link", "create"},
			} {
				wantOut, wantErr, wantExit := graphGenerationRun(t, bd, empty, home, args...)
				out, stderr, exit := graphGenerationRun(t, bd, work, home, args...)
				if wantExit != 0 || exit != 0 || out != wantOut || stderr != wantErr {
					t.Fatalf("%v differs inside graph workspace: exit=%d/%d stdout=%q/%q stderr=%q/%q", args, exit, wantExit, out, wantOut, stderr, wantErr)
				}
				if len(args) > 0 && args[0] == "version" && strings.Contains(strings.Join(args, " "), "json") {
					var version map[string]any
					if err := json.Unmarshal([]byte(out), &version); err != nil || version["version"] == nil || version["build"] == nil {
						t.Fatalf("version output lost ordinary JSON shape: %s %v", out, err)
					}
				}
				if len(args) == 2 && args[0] == "help" && args[1] == "--all" {
					if !strings.Contains(out, "# bd — Complete Command Reference") || !strings.Contains(out, "### bd create") {
						t.Fatalf("complete reference missing commands: %s", out)
					}
					completeReference = out
				}
			}
			generated := t.TempDir()
			out, stderr, exit := graphGenerationRun(t, bd, work, home, "help", "--docs-root", generated)
			reference, err := os.ReadFile(filepath.Join(generated, "docs", "CLI_REFERENCE.md"))
			// The ordinary Markdown writer stores exactly one trailing newline.
			wantReference := strings.TrimRight(completeReference, "\n") + "\n"
			if exit != 0 || out != "" || stderr != "" || err != nil || string(reference) != wantReference {
				t.Fatalf("help docs generation lost its no-store semantics: exit=%d stdout=%q stderr=%q read=%v", exit, out, stderr, err)
			}
			// Existing argument errors and graph-mode syntax still apply.
			for _, args := range [][]string{{"version", "not-a-version-selector"}, {"version", "--graph-mode", "invalid"}, {"--version", "--graph-mode", "invalid"}, {"--graph-mode", "invalid"}} {
				wantOut, wantErr, wantExit := graphGenerationRun(t, bd, empty, home, args...)
				out, stderr, exit := graphGenerationRun(t, bd, work, home, args...)
				if wantExit == 0 || exit != wantExit || out != wantOut || stderr != wantErr {
					t.Fatalf("%v lost ordinary refusal: exit=%d/%d stdout=%q/%q stderr=%q/%q", args, exit, wantExit, out, wantOut, stderr, wantErr)
				}
			}
			graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "duplicate", "arbitrary", "--json")
			if after := legacyUpgradeTreeDigest(t, work); after != before {
				t.Fatal("informational commands or unsupported operation changed workspace files")
			}
		})
	}
}
