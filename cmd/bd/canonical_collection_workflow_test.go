//go:build cgo

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
			call("link", "create", scope+"beads/first", scope+"beads/second", "--type", scope+"superseded-by", "--id", "links/related", "--json")
			// Allocation succeeds for this selector, but the selected resource is
			// a Link rather than a Bead endpoint. Keep that refusal a selector
			// error after the store-bound identity lookup.
			ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
			wrongKind := exec.CommandContext(ctx, bd, "link", "list", "--from", "links/related", "--output-format=json")
			wrongKind.Dir = work
			wrongKind.Env = []string{"PATH=" + os.Getenv("PATH"), "HOME=" + home, "BD_DISABLE_METRICS=1", "BD_DISABLE_EVENT_FLUSH=1", "BD_NON_INTERACTIVE=1", "BEADS_DOLT_AUTO_START=0", "DOLT_METRICS_DISABLED=1"}
			wrongOutput, wrongErr := wrongKind.Output()
			cancel()
			var wrongExit *exec.ExitError
			if !errors.As(wrongErr, &wrongExit) || wrongExit.ExitCode() != 2 || !strings.Contains(string(wrongOutput), `"code":"invalid_selector"`) {
				t.Fatalf("Link as Bead endpoint: exit=%v output=%s", wrongErr, wrongOutput)
			}
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
			owned := page.Items[0].OwnedLinks[scope+"superseded-by"]
			if page.Items[0].ID != scope+"beads/first" || len(owned) != 1 || owned[0].ID != scope+"links/related" {
				t.Fatalf("public current projection lost owned state: %s", raw)
			}
			links := call("link", "list", "--output-format", "json", "--unpaged", "--from", scope+"beads/first", "--to", scope+"beads/second", "--omit", "envelope")
			var linkPage bdpwire.LinkCollection
			if err := bdpwire.Unmarshal([]byte(links), &linkPage); err != nil {
				t.Fatal(err)
			}
			if len(linkPage.Items) != 1 || linkPage.Items[0].Source.URI != scope+"beads/first" || linkPage.Items[0].Target.URI != scope+"beads/second" || linkPage.Next != nil {
				t.Fatalf("wrong Link projection: %s", links)
			}
			stream := call("link", "list", "--output-format", "jsonl", "--incident-to", scope+"beads/second", "--direction", "in")
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
					for _, invalid := range [][]string{{"list", "--omit=/revision"}, {"list", "--include=/properties/~bad"}, {"list", "--continue=token"}, {"list", "--type=any", "--paged", "--page-size=1"}, {"list", "unexpected"}, {"link", "list", "unexpected"}, {"list", "--limit=010", "--limit=10"}, {"list", "--all"}, {"list", "--type=bug", "--conforms-to=" + graphstore.IssueTypeURL(scope)}, {"link", "list", "--conforms-to=preview-memory"}} {
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

// One installed population drives legacy list/count and canonical output.
// It catches a count/list drift that pure endpoint predicate tests cannot.
func TestGraphLinkCollectionPositionalRevisionWorkflow(t *testing.T) {
	isolateBeadsDirForTest(t)
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			const scope = "https://example.invalid/link-collection-selection/"
			work, home := graphListWorkspace(t, bd, engine, scope)
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", args...)
			}
			first := graphMixedResult[graphstore.Record](t, call("remember", "A", "--id", "a", "--json"))
			call("remember", "B", "--id", "b", "--json")
			create := func(id, from, to string) {
				t.Helper()
				call("link", "create", from, to, "--type", "references", "--id", id, "--json")
			}
			create("floating", "a", "b")
			create("old", "a@current", "b")
			call("update", "a", "--properties", `{"body":"A revised"}`, "--json")
			current := graphMixedResult[graphstore.Record](t, call("show", "a", "--json"))
			if current.Revision == first.Revision {
				t.Fatal("source revision did not advance")
			}
			create("current", "a@current", "b")
			create("reverse", "b", "a")
			list := func(args ...string) []graphstore.LinkRecord {
				t.Helper()
				return graphMixedResult[[]graphstore.LinkRecord](t, call(append([]string{"link", "list"}, append(args, "--json")...)...))
			}
			count := func(args ...string) string {
				t.Helper()
				return strings.TrimSpace(call(append([]string{"link", "count"}, args...)...))
			}
			for _, tc := range []struct {
				name string
				args []string
				want int
			}{
				{"all current", nil, 4},
				{"bare incident", []string{"a"}, 4},
				{"current incident", []string{"a@current"}, 3},
				{"stable ordinal", []string{"a@1"}, 1},
				{"exact old token", []string{"a@=" + first.Revision}, 1},
				{"all revisions still current Links", []string{"a@1", "--all-revisions"}, 4},
				{"flag historical source", []string{"--from", "a@1"}, 1},
				{"flag current outgoing", []string{"--incident-to", "a@current", "--direction", "out"}, 2},
				{"flag current incoming", []string{"--to", "a@current"}, 1},
				{"incident and source intersection", []string{"a", "--from", "b"}, 1},
				{"incident and target intersection", []string{"a", "--to", "b"}, 3},
				{"out pair", []string{"a", "b"}, 3},
				{"pair and incident intersection", []string{"a", "b", "--incident-to", "a@1"}, 1},
				{"in pair", []string{"a", "b", "--direction", "in"}, 1},
				{"both pair", []string{"a", "b", "--direction", "both"}, 4},
				{"self pair empty", []string{"a", "a", "--direction", "both"}, 0},
				{"filtered family and nominal", []string{"a", "--type", "references", "--exact-type", "references"}, 4},
			} {
				t.Run(tc.name, func(t *testing.T) {
					if got := list(tc.args...); len(got) != tc.want {
						t.Fatalf("list returned %d want %d: %+v", len(got), tc.want, got)
					}
					if got := count(tc.args...); got != fmt.Sprint(tc.want) {
						t.Fatalf("count=%q want %d", got, tc.want)
					}
				})
			}
			var canonical bdpwire.LinkCollection
			if raw := call("link", "list", "a@current", "--output-format", "json", "--unpaged", "--omit", "envelope"); bdpwire.Unmarshal([]byte(raw), &canonical) != nil || len(canonical.Items) != 3 {
				t.Fatalf("canonical Link list lost pin applicability: %s", raw)
			}
			for _, refusal := range []struct {
				selector string
				code     string
			}{
				{"missing", "not_found"},
				{"a@=no-such-revision", "revision_unknown"},
				{"a@", "invalid_selector"},
			} {
				ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
				command := graphMemoryReadCommand(ctx, bd, work, home, "link", "list", refusal.selector,
					"--output-format", "json", "--unpaged", "--omit", "envelope")
				var stdout, stderr bytes.Buffer
				command.Stdout, command.Stderr = &stdout, &stderr
				err := command.Run()
				cancel()
				var exit *exec.ExitError
				wantExit := map[string]int{"not_found": 3, "revision_unknown": 3, "invalid_selector": 2}[refusal.code]
				if !errors.As(err, &exit) || exit.ExitCode() != wantExit || stdout.Len() != 0 ||
					strings.Count(stderr.String(), refusal.code+": ") != 1 || strings.Count(stderr.String(), "\n") != 1 {
					t.Fatalf("canonical Link refusal %q: err=%v stdout=%q stderr=%q", refusal.selector, err, &stdout, &stderr)
				}
			}
			for _, refusal := range []struct {
				code string
				args []string
			}{
				{"invalid_selector", []string{"link", "list", "a", "--incident-to", "b"}},
				{"invalid_selector", []string{"link", "list", "a", "b", "--from", "a"}},
				{"invalid_selector", []string{"link", "list", "--direction", "in"}},
				{"revision_unknown", []string{"link", "list", "a@=no-such-revision"}},
			} {
				graphMemoryReadRaw(t, bd, work, home, "", refusal.code, refusal.args...)
			}
			call("remember", "literal", "--id", "beads/a", "--json")
			if got := list("beads/a"); len(got) != 0 {
				t.Fatalf("path-looking public ID selected a different Bead's Links: %+v", got)
			}
		})
	}
}
