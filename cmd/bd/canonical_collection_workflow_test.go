//go:build cgo

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/httpapi/bdpwire"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// Exercises real Cobra registration, persisted public-record projection, and
// independent CLI processes. Lower-seam tests own masks and paging refusals.
func TestCanonicalCollectionInstalledWorkflow(t *testing.T) {
	isolateBeadsDirForTest(t)
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			scope := "https://example.invalid/canonical-collection/"
			work, home := graphListWorkspace(t, bd, engine, scope)
			call := func(args ...string) string { return graphPolicyCLI(t, bd, work, home, nil, "", args...) }
			call("remember", "First body", "--id", "beads/first", "--json")
			call("remember", "Second body", "--id", "beads/second", "--json")
			call("link", "create", "first", "second", "--type", scope+"types/superseded-by", "--id", "links/related", "--json")
			var beads struct {
				SchemaVersion int             `json:"schema_version"`
				BDPSpec       string          `json:"bdp_spec"`
				Data          json.RawMessage `json:"data"`
			}
			raw := call("list", "--output-format", "json", "--unpaged", "--type", graphstore.MemoryTypeURL(scope))
			if err := json.Unmarshal([]byte(raw), &beads); err != nil {
				t.Fatal(err)
			}
			var page bdpwire.BeadCollection
			if err := bdpwire.Unmarshal(beads.Data, &page); err != nil {
				t.Fatal(err)
			}
			if beads.SchemaVersion != 2 || beads.BDPSpec != bdpwire.Pin || len(page.Items) != 2 || page.Next != nil {
				t.Fatalf("wrong native collection: %s", raw)
			}
			owned := page.Items[0].OwnedLinks[scope+"types/superseded-by"]
			if page.Items[0].ID != scope+"beads/first" || len(owned) != 1 || owned[0].ID != scope+"links/related" {
				t.Fatalf("public current projection lost owned state: %s", raw)
			}
			links := call("link", "list", "--output-format", "json", "--unpaged", "--source", "first", "--target", "second", "--omit", "envelope")
			var linkPage bdpwire.LinkCollection
			if err := bdpwire.Unmarshal([]byte(links), &linkPage); err != nil {
				t.Fatal(err)
			}
			if len(linkPage.Items) != 1 || linkPage.Items[0].Source.URI != scope+"beads/first" || linkPage.Items[0].Target.URI != scope+"beads/second" || linkPage.Next != nil {
				t.Fatalf("wrong Link projection: %s", links)
			}
			stream := call("link", "list", "--output-format", "jsonl", "--incident-to", "second", "--direction", "in")
			lines := strings.Split(strings.TrimSpace(stream), "\n")
			if len(lines) != 3 || !strings.Contains(lines[0], `"record_type":"header"`) || !strings.Contains(lines[2], `"stop_reason":"exhausted"`) {
				t.Fatalf("wrong finite stream: %s", stream)
			}
			// Compatibility stays on the legacy surface when canonical opt-in is absent.
			legacy := call("link", "list", "--json")
			if strings.Contains(legacy, `"schema_version":2`) {
				t.Fatal("legacy Link list switched contract")
			}

			for _, format := range []string{"json", "jsonl"} {
				for _, raw := range []bool{false, true} {
					for _, invalid := range [][]string{{"list", "--omit=/revision"}, {"list", "--include=/properties/~bad"}, {"list", "--continue=token"}, {"list", "--paged", "--page-size=1"}, {"list", "unexpected"}, {"link", "list", "unexpected"}, {"list", "--limit=010", "--limit=10"}, {"list", "--all"}, {"list", "--type=bug"}, {"link", "list", "--conforms-to=preview-memory"}} {
						args := append(append([]string{}, invalid...), "--output-format="+format)
						if raw {
							args = append(args, "--omit=envelope")
						}
						ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
						command := exec.CommandContext(ctx, bd, args...)
						command.Dir = work
						command.Env = []string{"PATH=" + os.Getenv("PATH"), "HOME=" + home, "BD_DISABLE_METRICS=1", "BD_DISABLE_EVENT_FLUSH=1", "BD_NON_INTERACTIVE=1", "BEADS_DOLT_AUTO_START=0", "DOLT_METRICS_DISABLED=1"}
						var stdout, stderr bytes.Buffer
						command.Stdout, command.Stderr = &stdout, &stderr
						err := command.Run()
						cancel()
						if err == nil || stderr.Len() == 0 {
							t.Fatalf("refusal missing for %v: %v %s", args, err, &stderr)
						}
						if raw {
							if stdout.Len() != 0 {
								t.Fatalf("raw failure leaked output: %s", &stdout)
							}
							continue
						}
						result := stdout.Bytes()
						if format == "jsonl" {
							lines := bytes.Split(bytes.TrimSpace(result), []byte("\n"))
							if len(lines) != 2 {
								t.Fatalf("pre-read failure must emit header and terminal: %s", result)
							}
							result = lines[1]
						}
						var frame struct {
							Outcome    string          `json:"outcome"`
							Error      json.RawMessage `json:"error"`
							Completion struct {
								Returned   int    `json:"returned"`
								Exhausted  *bool  `json:"exhausted"`
								StopReason string `json:"stop_reason"`
							} `json:"completion"`
						}
						if err := json.Unmarshal(result, &frame); err != nil || frame.Outcome != "failure" || len(frame.Error) == 0 || frame.Completion.Returned != 0 || frame.Completion.Exhausted != nil || frame.Completion.StopReason != "error" {
							t.Fatalf("bad collection failure %s: %v", result, err)
						}
					}
				}
			}
		})
	}
}
