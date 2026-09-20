package graphops

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// This tests the actual linter rule, then deletes that rule and proves that the
// very same forbidden import becomes permitted. Opt-in because it invokes the
// separately installed/pinned golangci-lint executable, not a Go dependency.
func TestGraphBodyConsumerDepguardMutation(t *testing.T) {
	linter := os.Getenv("BEADS_GRAPH_DEP_GUARD_LINTER")
	if linter == "" {
		t.Skip("requires explicit absolute golangci-lint executable")
	}
	if !filepath.IsAbs(linter) {
		t.Fatal("absolute linter path required")
	}
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", ".golangci.yml"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	start := strings.Index(text, "        graphops-body-consumers:\n")
	if start < 0 {
		t.Fatal("missing consumer rule")
	}
	end := strings.Index(text[start:], "        # end graphops-body-consumers\n")
	if end < 0 {
		t.Fatal("missing rule boundary")
	}
	rule := text[start : start+end]
	directory := t.TempDir()
	files := map[string]string{
		"internal/storage/dolt/consumer.go": "package dolt\nimport _ \"github.com/steveyegge/beads/internal/storage/graphops\"\n",
		"go.mod":                            "module github.com/steveyegge/beads\n\ngo 1.26.5\n",
		"internal/storage/graphops/body.go": "package graphops\n",
		"cmd/bd/forbidden.go":               "package main\nimport _ \"github.com/steveyegge/beads/internal/storage/graphops\"\nfunc main() {}\n",
	}
	for name, contents := range files {
		p := filepath.Join(directory, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(contents), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	prefix := "version: '2'\nrun:\n  tests: false\nlinters:\n  default: none\n  enable: [depguard]\n  settings:\n    depguard:\n      rules:\n"
	for _, test := range []struct {
		enabled bool
		target  string
	}{{true, "./cmd/bd"}, {false, "./cmd/bd"}, {true, "./internal/storage/dolt"}} {
		enabled := test.enabled
		configuration := prefix + rule
		if !enabled {
			configuration = prefix + "        graphops-body-consumers:\n          list-mode: lax\n          files: ['$all']\n          allow: ['$gostd', 'github.com/steveyegge/beads']\n"
		}
		configPath := filepath.Join(directory, ".golangci.yml")
		if err := os.WriteFile(configPath, []byte(configuration), 0o600); err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
		cmd := exec.CommandContext(ctx, linter, "run", "--config", configPath, test.target)
		cmd.Dir = directory
		cmd.Env = append(os.Environ(), "GOWORK=off", "CGO_ENABLED=0", "GOPROXY=off", "GOSUMDB=off", "GOTOOLCHAIN=local")
		output, err := cmd.CombinedOutput()
		cancel()
		if enabled && test.target == "./cmd/bd" {
			if err == nil || !strings.Contains(string(output), "Graph transaction bodies are private to storage legs") {
				t.Fatalf("deny rule did not reject import: %v %s", err, output)
			}
		} else if err != nil {
			t.Fatalf("allowed import must pass: %v %s", err, output)
		}
		t.Logf("deny-enabled=%t target=%s exit=%v output=%s", enabled, test.target, err, output)
	}
}
