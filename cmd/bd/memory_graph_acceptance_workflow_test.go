//go:build cgo

package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/configfile"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"github.com/steveyegge/beads/internal/types"
	publicops "github.com/steveyegge/beads/issueops"
)

type memoryAcceptanceCorpus struct {
	Fixtures map[string]struct {
		Beads []struct{ ID, Type string }
		Links []struct{ ID, Type, Source, Target string }
	}
	ReadCases []struct {
		ID, Fixture string
		Command     []string
		Operation   *struct {
			Root, Family, Direction string
			Depth                   int
		}
		Expected struct {
			SelectedBeads  []string `json:"selectedBeads"`
			SelectedLinks  []string `json:"selectedLinks"`
			Frontier       []string
			RootMembership map[string][]string `json:"rootMembership"`
			FrontierByRoot map[string][]string `json:"frontierByRoot"`
			LinksByRoot    map[string][]string `json:"linksByRoot"`
		}
	} `json:"readCases"`
}

// Turn the accepted authored scenarios into actual installed-command assertions.
// Fixtures are created by checked production writers, never direct SQL. Each
// fixture's identities have their own prefix in one disposable workspace.
func TestMemoryCLIReadAcceptanceWorkflow(t *testing.T) {
	isolateBeadsDirForTest(t)
	data, err := os.ReadFile(filepath.Join("testdata", "memory-cli-preview3-cases.json"))
	if err != nil {
		t.Fatal(err)
	}
	var corpus memoryAcceptanceCorpus
	if err := json.Unmarshal(data, &corpus); err != nil {
		t.Fatal(err)
	}
	if len(corpus.ReadCases) != 22 {
		t.Fatalf("expected accepted22 read cases, got%d", len(corpus.ReadCases))
	}
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			const scope = "https://example.invalid/memory-acceptance/"
			work, home := graphListWorkspace(t, bd, engine, scope)
			memoryAcceptanceSeed(t, work, scope, corpus)
			for _, tc := range corpus.ReadCases {
				t.Run(tc.ID, func(t *testing.T) {
					command := tc.Command
					if len(command) == 0 && tc.Operation != nil {
						command = []string{"bd", "graph", tc.Operation.Root, "--link-type", tc.Operation.Family, "--direction", tc.Operation.Direction, "--depth", strconv.Itoa(tc.Operation.Depth)}
					}
					if len(command) < 3 {
						t.Fatalf("missing concrete invocation for %s", tc.ID)
					}
					args := slices.Clone(command[1:])
					prefix := "beads/" + tc.Fixture + "/"
					start := 1
					if args[0] == "memory" {
						start = 2
					}
					for i := start; i < len(args) && !strings.HasPrefix(args[i], "-"); i++ {
						args[i] = prefix + args[i]
					}
					output := graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
					var beads, links, frontier []string
					var roots []memoryGraphRoot
					if args[0] == "graph" {
						result := graphSelectedTestSummary(t, output)
						if text := graphPolicyCLI(t, bd, work, home, nil, "", args...); strings.Contains(text, "(cycle)") {
							t.Fatalf("acyclic both-direction case labelled cycle:\n%s", text)
						}
						for _, b := range result.Nodes {
							beads = append(beads, b.ID)
						}
						for _, l := range result.Links {
							links = append(links, l.ID)
						}
						frontier = result.Frontier
					} else {
						var envelope struct {
							Outcome    string
							ResultKind string `json:"result_kind"`
							Data       memoryGraphResult
						}
						if err := json.Unmarshal([]byte(output), &envelope); err != nil || envelope.Outcome != "success" || envelope.ResultKind != "graph" {
							t.Fatalf("bad canonical graph: %s (%v)", output, err)
						}
						for _, b := range envelope.Data.Beads {
							beads = append(beads, b.ID)
						}
						for _, l := range envelope.Data.Links {
							links = append(links, l.ID)
						}
						roots = envelope.Data.Traversal.Roots
						for _, r := range roots {
							frontier = append(frontier, r.Frontier...)
						}
					}
					expect := func(what string, actual, want []string, kind string) {
						t.Helper()
						short := make([]string, 0, len(actual))
						for _, id := range actual {
							base := scope + kind + "/" + tc.Fixture + "/"
							if !strings.HasPrefix(id, base) {
								t.Fatalf("%s has foreign/unrelated identity %q", what, id)
							}
							short = append(short, strings.TrimPrefix(id, base))
						}
						sorted := slices.Clone(want)
						slices.Sort(short)
						slices.Sort(sorted)
						if !slices.Equal(short, sorted) {
							t.Fatalf("%s: got%v want%v", what, short, sorted)
						}
					}
					expect("Beads", beads, tc.Expected.SelectedBeads, "beads")
					expect("Links", links, tc.Expected.SelectedLinks, "links")
					if tc.Expected.RootMembership == nil {
						expect("frontier", frontier, tc.Expected.Frontier, "beads")
					} else {
						if len(roots) != len(tc.Expected.RootMembership) {
							t.Fatalf("root slots lost: %+v", roots)
						}
						for _, r := range roots {
							name := strings.TrimPrefix(r.ID, scope+prefix)
							membership, ok := tc.Expected.RootMembership[name]
							if !ok {
								t.Fatalf("unexpected root %s", r.ID)
							}
							expect("root "+name, r.Beads, membership, "beads")
							expect("root links "+name, r.Links, tc.Expected.LinksByRoot[name], "links")
							expect("root frontier "+name, r.Frontier, tc.Expected.FrontierByRoot[name], "beads")
						}
					}
				})
			}
		})
	}
}

func memoryAcceptanceSeed(t *testing.T, work, scope string, corpus memoryAcceptanceCorpus) {
	t.Helper()
	cfg, err := configfile.LoadForDiscovery(filepath.Join(work, ".beads"))
	if err != nil || cfg == nil {
		t.Fatalf("binding: %v", err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 3*time.Minute)
	defer cancel()
	options := graphOptions(cfg)
	options.DataDir = filepath.Join(cfg.GraphWorkspace, "embeddeddolt")
	store, err := graphstore.OpenExisting(ctx, options)
	if err != nil {
		t.Fatal(err)
	}
	closed := false
	defer func() {
		if !closed {
			_ = store.Close()
		}
	}()
	fixtures := make([]string, 0, len(corpus.Fixtures))
	for name := range corpus.Fixtures {
		fixtures = append(fixtures, name)
	}
	slices.Sort(fixtures)
	for _, name := range fixtures {
		fixture := corpus.Fixtures[name]
		for _, b := range fixture.Beads {
			path := "beads/" + name + "/" + b.ID
			if b.Type == "issue" {
				// Keep the corpus's literal prefixed public identity explicit;
				// specifying only its private path would generate another ID.
				_, err = store.CreateIssue(ctx, path, publicops.CreateRequest{Actor: "memory-cli-test", ForceIDPrefix: true, Issue: &publicops.Issue{ID: path, Title: b.ID, Status: types.StatusOpen, IssueType: types.TypeTask, Priority: 2}})
			} else {
				_, err = store.Create(ctx, graphstore.CreateRequest{Path: path, TypeURL: scope + "types/" + b.Type, Title: b.ID, Body: "Acceptance fixture", Actor: "memory-cli-test"})
			}
			if err != nil {
				t.Fatalf("seed Bead %s: %v", path, err)
			}
		}
		for _, l := range fixture.Links {
			_, err = store.AddInformationalLink(ctx, graphstore.LinkCreateRequest{Path: "links/" + name + "/" + l.ID, SourcePath: "beads/" + name + "/" + l.Source, TargetPath: "beads/" + name + "/" + l.Target, TypeURL: scope + "types/" + l.Type, Actor: "memory-cli-test", UnconditionalSource: true})
			if err != nil {
				t.Fatalf("seed Link %s/%s: %v", name, l.ID, err)
			}
		}
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	closed = true
}
