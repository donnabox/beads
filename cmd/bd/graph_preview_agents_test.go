package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/templates/agents"
)

func TestGraphPreviewAgentInstructionsFileRouting(t *testing.T) {
	for _, filename := range []string{"AGENTS.md", "TEAM.md"} {
		t.Run(filename, func(t *testing.T) {
			work := t.TempDir()
			path := filepath.Join(work, filename)
			user := "# User instructions\nKeep my instructions.\n"
			if err := os.WriteFile(path, []byte(user), 0600); err != nil {
				t.Fatal(err)
			}
			claude := "@" + filename + "\n"
			if err := os.WriteFile(filepath.Join(work, "CLAUDE.md"), []byte(claude), 0600); err != nil {
				t.Fatal(err)
			}
			plan, err := prepareGraphPreviewAgentInstructions(work, filename, false)
			if err != nil {
				t.Fatal(err)
			}
			if err := plan.install(); err != nil {
				t.Fatal(err)
			}
			got, err := os.ReadFile(path)
			if err != nil || !strings.HasPrefix(string(got), user) || strings.Count(string(got), "profile:graph-preview") != 1 {
				t.Fatalf("managed graph guidance did not preserve user content: %q err=%v", got, err)
			}
			info, err := os.Stat(path)
			if err != nil || info.Mode().Perm() != 0600 {
				t.Fatalf("existing permissions changed: info=%v err=%v", info, err)
			}
			importFile, err := os.ReadFile(filepath.Join(work, "CLAUDE.md"))
			if err != nil || string(importFile) != claude {
				t.Fatalf("CLAUDE import changed: %q err=%v", importFile, err)
			}
			plan, err = prepareGraphPreviewAgentInstructions(work, filename, false)
			if err != nil {
				t.Fatal(err)
			}
			if err := plan.install(); err != nil {
				t.Fatal(err)
			}
			next, err := os.ReadFile(path)
			if err != nil || string(next) != string(got) {
				t.Fatalf("repeated managed install changed current content: err=%v", err)
			}
		})
	}
}

func TestGraphPreviewAgentInstructionsRefuseUnsafeFiles(t *testing.T) {
	minimal := agents.RenderSection(agents.ProfileMinimal)
	cases := map[string]string{
		"full":                           agents.RenderSection(agents.ProfileFull),
		"unknown":                        strings.Replace(minimal, "profile:minimal", "profile:future", 1),
		"legacy":                         "<!-- BEGIN BEADS INTEGRATION -->\nkeep\n<!-- END BEADS INTEGRATION -->\n",
		"missing-end":                    "<!-- BEGIN BEADS INTEGRATION v:1 profile:minimal hash:12345678 -->\nkeep\n",
		"missing-begin":                  "keep\n<!-- END BEADS INTEGRATION -->\n",
		"duplicate":                      minimal + minimal,
		"reversed":                       "<!-- END BEADS INTEGRATION -->\n<!-- BEGIN BEADS INTEGRATION v:1 profile:minimal hash:12345678 -->\n",
		"malformed-begin":                strings.Replace(minimal, " -->", "", 1),
		"duplicate-profile-full-minimal": strings.Replace(minimal, "profile:minimal", "profile:full profile:minimal", 1),
		"duplicate-profile-minimal-full": strings.Replace(minimal, "profile:minimal", "profile:minimal profile:full", 1),
		"missing-version":                strings.Replace(minimal, " v:1", "", 1),
		"invalid-version":                strings.Replace(minimal, "v:1", "v:not-a-version", 1),
		"unsupported-version":            strings.Replace(minimal, "v:1", "v:2", 1),
		"duplicate-version":              strings.Replace(minimal, "v:1", "v:2 v:1", 1),
		"missing-hash":                   strings.Replace(minimal, " hash:"+agents.CurrentHash(agents.ProfileMinimal), "", 1),
		"invalid-hash":                   strings.Replace(minimal, "hash:"+agents.CurrentHash(agents.ProfileMinimal), "hash:not-a-digest", 1),
		"duplicate-hash":                 strings.Replace(minimal, "hash:", "hash:12345678 hash:", 1),
		"unknown-marker-field":           strings.Replace(minimal, " -->", " extra:value -->", 1),
	}
	for name, content := range cases {
		t.Run(name, func(t *testing.T) {
			work := t.TempDir()
			path := filepath.Join(work, "AGENTS.md")
			if err := os.WriteFile(path, []byte(content), 0600); err != nil {
				t.Fatal(err)
			}
			if plan, err := prepareGraphPreviewAgentInstructions(work, "AGENTS.md", false); err == nil || plan != nil || !strings.Contains(err.Error(), "--skip-agents") {
				t.Fatalf("unsafe managed section accepted: plan=%+v err=%v", plan, err)
			}
			got, err := os.ReadFile(path)
			if err != nil || string(got) != content {
				t.Fatalf("refusal changed user file: %q err=%v", got, err)
			}
			plan, err := prepareGraphPreviewAgentInstructions(work, "AGENTS.md", true)
			if err != nil || plan != nil {
				t.Fatalf("explicit skip must bypass inspection: %+v %v", plan, err)
			}
			if err := plan.install(); err != nil {
				t.Fatal(err)
			}
		})
	}
	for _, kind := range []string{"directory", "symlink"} {
		t.Run(kind, func(t *testing.T) {
			work := t.TempDir()
			path := filepath.Join(work, "AGENTS.md")
			outside := filepath.Join(t.TempDir(), "outside.md")
			if err := os.WriteFile(outside, []byte("outside sentinel"), 0600); err != nil {
				t.Fatal(err)
			}
			var err error
			if kind == "directory" {
				err = os.Mkdir(path, 0700)
			} else {
				err = os.Symlink(outside, path)
			}
			if err != nil {
				t.Fatal(err)
			}
			if _, err := prepareGraphPreviewAgentInstructions(work, "AGENTS.md", false); err == nil {
				t.Fatal("nonregular agent target accepted")
			}
			got, err := os.ReadFile(outside)
			if err != nil || string(got) != "outside sentinel" {
				t.Fatalf("outside file changed: %q %v", got, err)
			}
		})
	}
}

func TestGraphPreviewAgentInstructionsFreshAndMinimal(t *testing.T) {
	for _, old := range []string{"", "user\n" + agents.RenderSection(agents.ProfileMinimal) + "tail\n"} {
		work := t.TempDir()
		path := filepath.Join(work, "AGENTS.md")
		if old != "" {
			if err := os.WriteFile(path, []byte(old), 0600); err != nil {
				t.Fatal(err)
			}
		}
		plan, err := prepareGraphPreviewAgentInstructions(work, "AGENTS.md", false)
		if err != nil {
			t.Fatal(err)
		}
		if err := plan.install(); err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(path)
		if err != nil || strings.Contains(string(got), "bd prime") || !strings.Contains(string(got), "bd status --graph") {
			t.Fatalf("graph install contains ordinary preamble or lacks graph guidance: %q %v", got, err)
		}
		if old != "" && (!strings.HasPrefix(string(got), "user\n") || !strings.HasSuffix(string(got), "tail\n")) {
			t.Fatal("minimal profile replacement lost surrounding user content")
		}
	}
}

func TestGraphPreviewAgentInstructionsDetectChangedTarget(t *testing.T) {
	work := t.TempDir()
	path := filepath.Join(work, "AGENTS.md")
	plan, err := prepareGraphPreviewAgentInstructions(work, "AGENTS.md", false)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("concurrent user edit"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := plan.install(); err == nil {
		t.Fatal("installation overwrote an intervening edit")
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != "concurrent user edit" {
		t.Fatalf("intervening edit lost: %q %v", got, err)
	}
	if _, err := prepareGraphPreviewAgentInstructions(work, "../outside.md", false); err == nil {
		t.Fatal("unsafe routing filename accepted")
	}
}

func TestGraphPreviewAgentInstructionsWriteFailure(t *testing.T) {
	work := filepath.Join(t.TempDir(), "workspace")
	if err := os.Mkdir(work, 0700); err != nil {
		t.Fatal(err)
	}
	plan, err := prepareGraphPreviewAgentInstructions(work, "AGENTS.md", false)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(work); err != nil {
		t.Fatal(err)
	}
	if err := plan.install(); err == nil || !strings.Contains(err.Error(), "write agent file") {
		t.Fatalf("file-write failure was not returned: %v", err)
	}
	if _, err := os.Lstat(work); !os.IsNotExist(err) {
		t.Fatalf("failed installation created unexpected files: %v", err)
	}
}
