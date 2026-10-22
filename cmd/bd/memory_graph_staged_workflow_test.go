//go:build cgo

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestMemoryConstraintReadCLIWorkflow(t *testing.T) {
	isolateBeadsDirForTest(t)
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			const scope = "https://example.invalid/staged-cli/"
			work, home := graphListWorkspace(t, bd, engine, scope)
			var fixture memoryAcceptanceCorpus
			if err := json.Unmarshal([]byte(`{"fixtures":{"recurrence":{"beads":[{"id":"s","type":"policy"},{"id":"c","type":"specification"},{"id":"d","type":"requirement"},{"id":"else","type":"policy"}],"links":[{"id":"sc","source":"s","target":"c","type":"subject-to"},{"id":"cs","source":"c","target":"s","type":"requires"},{"id":"sd","source":"s","target":"d","type":"requires"},{"id":"ce","source":"c","target":"else","type":"subject-to"}]}}}`), &fixture); err != nil {
				t.Fatal(err)
			}
			memoryAcceptanceSeed(t, work, scope, fixture)
			call := func(args ...string) string { t.Helper(); return graphPolicyCLI(t, bd, work, home, nil, "", args...) }
			decode := func(output string) memoryGraphResult {
				t.Helper()
				var env struct {
					Outcome string
					Data    memoryGraphResult
				}
				if err := json.Unmarshal([]byte(output), &env); err != nil || env.Outcome != "success" {
					t.Fatalf("bad graph: %s %v", output, err)
				}
				return env.Data
			}
			result := decode(call("memory", "constraints", "recurrence/s", "--json"))
			if len(result.Beads) != 3 || len(result.Links) != 3 || len(result.Traversal.Roots) != 1 || result.Traversal.Discovery != "explicit-subject-to-and-requires" {
				t.Fatalf("staged selection: %+v", result.Traversal)
			}
			root := result.Traversal.Roots[0]
			if root.Depth != nil || result.Traversal.Direction != "staged" || root.Stages == nil || root.Stages.Requires.Depth != 1000 || len(root.Stages.SubjectTo) != 1 || len(root.Stages.Requires.Beads) != 3 || len(root.Stages.Requires.Links) != 2 || root.Stages.SubjectTo[0].Link != scope+"links/recurrence/sc" {
				t.Fatalf("phase witnesses lost: %+v", root.Stages)
			}
			human := call("memory", "constraints", "recurrence/s")
			if strings.Count(human, "s [policy]") != 2 || !strings.Contains(human, "(recurrence/sd)") || strings.Contains(human, "(recurrence/ce)") || !strings.Contains(human, "Discovery: explicit Subject To declarations and their Requires dependencies") {
				t.Fatalf("phase recurrence/scope:\n%s", human)
			}
			human = call("memory", "constraints", "recurrence/s", "recurrence/c")
			if strings.Count(human, "Discovery: explicit Subject To declarations") != 1 || !strings.Contains(human, "Root recurrence/s:") || !strings.Contains(human, "Root recurrence/c:") {
				t.Fatalf("multi-root scope attribution:\n%s", human)
			}
			direct := decode(call("memory", "subjects", "recurrence/c", "--json"))
			if len(direct.Beads) != 2 || len(direct.Links) != 1 || direct.Traversal.Discovery != "direct-incoming-subject-to" {
				t.Fatalf("direct declaration scope: %+v", direct.Traversal)
			}
			human = call("memory", "subjects", "recurrence/c")
			if !strings.Contains(human, "Discovery: direct incoming Subject To declarations only") || strings.Contains(human, "depth:") || strings.Contains(human, "(recurrence/cs)") {
				t.Fatalf("direct scope display:\n%s", human)
			}
			for _, verb := range []string{"subjects", "constraints"} {
				ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
				command := graphMemoryReadCommand(ctx, bd, work, home, "memory", verb, "recurrence/c", "--json", "--depth=2")
				var stdout, stderr bytes.Buffer
				command.Stdout, command.Stderr = &stdout, &stderr
				err := command.Run()
				cancel()
				if err == nil || !strings.Contains(stderr.String(), "depth") {
					t.Fatalf("unselected public stage-depth admitted: %s %v %s", verb, err, &stderr)
				}
				if strings.Contains(stdout.String(), `"outcome":"success"`) {
					t.Fatalf("unsupported flag emitted success: %s", &stdout)
				}
			}
		})
	}
}
