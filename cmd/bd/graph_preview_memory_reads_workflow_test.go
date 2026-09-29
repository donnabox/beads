//go:build cgo

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/configfile"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// Match graphPolicyCLI's environment without inheriting operator routing or
// credentials. Raw recall assertions must preserve stdout bytes, not trim lines.
func graphMemoryReadCommand(ctx context.Context, bd, work, home string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, bd, args...)
	cmd.Dir = work
	cmd.Env = []string{
		"PATH=" + os.Getenv("PATH"), "HOME=" + home,
		"TMPDIR=" + os.TempDir(), "TMP=" + os.TempDir(), "TEMP=" + os.TempDir(),
		"XDG_CONFIG_HOME=" + filepath.Join(home, ".config"), "XDG_CACHE_HOME=" + filepath.Join(home, ".cache"),
		"GIT_CONFIG_GLOBAL=" + filepath.Join(home, "missing-gitconfig"), "GIT_CONFIG_NOSYSTEM=1",
		"BD_DISABLE_METRICS=1", "BD_DISABLE_EVENT_FLUSH=1", "BD_NON_INTERACTIVE=1",
		"BEADS_DOLT_AUTO_START=0", "DOLT_METRICS_DISABLED=1", "NO_COLOR=1",
	}
	if developerDir := os.Getenv("DEVELOPER_DIR"); developerDir != "" {
		cmd.Env = append(cmd.Env, "DEVELOPER_DIR="+developerDir)
	}
	cmd.WaitDelay = 5 * time.Second
	return cmd
}

func graphMemoryReadRaw(t *testing.T, bd, work, home, want, code string, args ...string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
	defer cancel()
	cmd := graphMemoryReadCommand(ctx, bd, work, home, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	if ctx.Err() != nil {
		t.Fatalf("read process exceeded deadline: %v stderr=%s", ctx.Err(), stderr.String())
	}
	if code == "" {
		if err != nil || stderr.Len() != 0 || !bytes.Equal(stdout.Bytes(), []byte(want)) {
			t.Fatalf("raw read %v: err=%v stdout=%q want=%q stderr=%q", args, err, stdout.Bytes(), want, stderr.Bytes())
		}
		return
	}
	var exit *exec.ExitError
	expectedExit := map[string]int{"invalid_selector": 2, "not_found": 3, "gone": 3, "revision_unknown": 3, "capability_unavailable": 5}[code]
	if !errors.As(err, &exit) || exit.ExitCode() != expectedExit || stdout.Len() != 0 || !strings.HasPrefix(stderr.String(), code+": ") {
		t.Fatalf("raw refusal %v: err=%v stdout=%q stderr=%q want=%s/%d", args, err, stdout.Bytes(), stderr.Bytes(), code, expectedExit)
	}
}

func graphMemoryReadBrokenPipe(t *testing.T, bd, work, home, selector string) {
	t.Helper()
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := read.Close(); err != nil {
		_ = write.Close()
		t.Fatal(err)
	}
	defer func() {
		if err := write.Close(); err != nil {
			t.Error(err)
		}
	}()
	ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
	defer cancel()
	cmd := graphMemoryReadCommand(ctx, bd, work, home, "recall", selector)
	var stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = write, &stderr
	err = cmd.Run()
	var exit *exec.ExitError
	if ctx.Err() != nil || !errors.As(err, &exit) || (!strings.Contains(stderr.String(), "broken pipe") && exit.ProcessState.String() != "signal: broken pipe") {
		t.Fatalf("closed output did not propagate EPIPE: err=%v deadline=%v stderr=%q", err, ctx.Err(), stderr.String())
	}
}

// Read-only inspection brackets the CLI read phase. Close before opening another
// installed process, including embedded mode. No fixture rows or schema are seeded.
func graphMemoryReadSnapshot(t *testing.T, work string) graphstore.Snapshot {
	t.Helper()
	cfg, err := configfile.LoadForDiscovery(filepath.Join(work, ".beads"))
	if err != nil || cfg == nil {
		t.Fatalf("read initialized binding: %v", err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
	defer cancel()
	s, err := graphstore.OpenExisting(ctx, graphstore.Options{
		Backend: cfg.DoltMode, DataDir: filepath.Join(cfg.GraphWorkspace, "embeddeddolt"), Database: cfg.DoltDatabase, Branch: "main",
		Binding:    graphstore.Binding{WorkspaceID: cfg.GraphWorkspace, ScopeURL: cfg.GraphScopeURL, AuthorityID: cfg.GraphAuthorityID, SchemaVersion: cfg.GraphSchemaVersion},
		ServerHost: cfg.DoltServerHost, ServerPort: cfg.DoltServerPort, ServerUser: cfg.DoltServerUser, ServerSocket: cfg.DoltServerSocket, ServerTLS: cfg.DoltServerTLS,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := s.Close(); err != nil {
			t.Error(err)
		}
	}()
	snapshot, err := s.CurrentSnapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	return snapshot
}

func graphMemoryReadValue(t *testing.T, value any) graphstore.VersionValue {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return graphstore.VersionValue{Present: true, Value: raw}
}

func graphMemoryReadComparison(t *testing.T, output string, want graphstore.VersionComparison) {
	t.Helper()
	got := graphMixedResult[graphstore.VersionComparison](t, output)
	// Normalize JSON member order inside retained owned-Link values; the oracle
	// still checks every value, presence bit, endpoint and unsupported field.
	semantic := func(value any) any {
		raw, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		var normalized any
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.UseNumber()
		if err := decoder.Decode(&normalized); err != nil {
			t.Fatal(err)
		}
		return normalized
	}
	if !reflect.DeepEqual(semantic(got), semantic(want)) {
		t.Fatalf("comparison=%+v want=%+v", got, want)
	}
}

func TestGraphPreviewMemoryReadsWorkflow(t *testing.T) {
	bd := buildBDUnderTest(t)
	for _, engine := range []string{"embedded", "server"} {
		t.Run(engine, func(t *testing.T) {
			work, home := t.TempDir(), t.TempDir()
			const scope = "https://example.invalid/memory-reads/"
			args := []string{"init", "--graph-mode", "link", "--scope-url", scope, "--skip-hooks", "--skip-agents", "--non-interactive"}
			if engine == "server" {
				port := os.Getenv("BEADS_GRAPH_TEST_SERVER_PORT")
				if port == "" {
					t.Skip("set BEADS_GRAPH_TEST_SERVER_PORT for ordinary shared-server read qualification")
				}
				args = append(args, "--server", "--external", "--server-host", "127.0.0.1", "--server-port", port, "--server-user", "root")
			}
			call := func(args ...string) string {
				t.Helper()
				return graphPolicyCLI(t, bd, work, home, nil, "", append(args, "--json")...)
			}
			discover := func(search string, extra ...string) graphMemoryDiscoveryResult {
				t.Helper()
				args := append([]string{"memories", search, "--format", "records-json", "--all", "--details"}, extra...)
				return graphMixedResult[graphMemoryDiscoveryResult](t, graphPolicyCLI(t, bd, work, home, nil, "", args...))
			}
			recall := func(want string, args ...string) {
				t.Helper()
				graphMemoryReadRaw(t, bd, work, home, want, "", append([]string{"recall"}, args...)...)
			}
			compare := func(from, to graphstore.Record, changes []graphstore.VersionChange, extra ...string) string {
				t.Helper()
				output := call(append([]string{"compare", from.ID, "--from", from.Version, "--to", to.Version}, extra...)...)
				graphMemoryReadComparison(t, output, graphstore.VersionComparison{
					Resource: graphstore.VersionComparisonResource{ID: from.ID, Type: from.Type},
					From:     graphstore.VersionComparisonEndpoint{Version: from.Version, Attribution: from.Attribution},
					To:       graphstore.VersionComparisonEndpoint{Version: to.Version, Attribution: to.Attribution},
					Compared: []string{"properties", "owned"}, Unsupported: []string{"commonMetadata", "inception", "derivation"}, Changes: changes,
				})
				return output
			}
			call(args...)
			body := "  leading\t雪 😀 e\u0301\r\nline two\ntrailing  "
			original := graphMixedResult[graphstore.Record](t, call("remember", body, "--id", "beads/plan", "--title", "Straße plan"))
			empty := graphMixedResult[graphstore.Record](t, call("remember", "", "--id", "beads/empty", "--title", "Empty"))
			contextMemory := graphMixedResult[graphstore.Record](t, call("remember", "Surviving context", "--id", "beads/context", "--title", "Context"))
			issue := graphMixedResult[graphstore.IssueRecord](t, call("create", "Surviving Issue", "--id", "beads/work"))
			selected := discover("STRASSE")
			if selected.Projection != "summary" || selected.Scope != scope || !selected.Complete || selected.Next != nil || len(selected.Items) != 1 || selected.Items[0].ID != original.ID || selected.Items[0].Version != original.Version || !reflect.DeepEqual(selected.Items[0].MatchedFields, []string{"title"}) || selected.Items[0].Excerpt != nil {
				t.Fatalf("literal case-folded discovery: %+v", selected)
			}
			recall(body, selected.Items[0].ID, "--version", selected.Items[0].Version)
			recall(body, "beads/plan", "--quiet")
			recall("", empty.ID)
			recall("", empty.ID, "--version", empty.Version, "--quiet")
			properties := graphstore.Properties{Title: "Revised plan", Body: "new-body-sentinel\n雪"}
			rawProperties, err := json.Marshal(properties)
			if err != nil {
				t.Fatal(err)
			}
			edited := graphMixedResult[graphstore.MemoryMutationResult](t, call("update", "beads/plan", "--properties", string(rawProperties), "--if-revision", original.Revision)).Memory
			related := scope + "types/preview-related-v2"
			firstLink := graphMixedResult[graphstore.LinkMutationResult](t, call("link", "beads/plan", "beads/context", "--id", "links/owned", "--resource-type", related, "--properties", `{"note":"first"}`, "--if-source-revision", edited.Revision)).Link
			linked := graphMixedResult[graphstore.Record](t, call("show", "beads/plan"))
			changedLink := graphMixedResult[graphstore.LinkMutationResult](t, call("update", "links/owned", "--properties", `{"note":"second"}`, "--if-revision", firstLink.Revision, "--if-source-revision", linked.Revision)).Link
			current := graphMixedResult[graphstore.Record](t, call("show", "beads/plan"))
			survivingLink := graphMixedResult[graphstore.LinkMutationResult](t, call("link", "beads/work", "beads/context", "--id", "links/survivor", "--resource-type", related)).Link
			issueCurrent := call("show", issue.ID)
			if current.Properties != properties || original.Properties.Body != body || empty.Properties.Body != "" {
				t.Fatal("fixture did not preserve authored body bytes")
			}

			// Every snapshot is fetched in its own installed process. Issue exact
			// projection intentionally need not invent native current-only fields.
			saved := []graphstore.Record{original, edited, linked, current, empty, contextMemory}
			exactBefore := map[string]string{}
			for _, memory := range saved {
				out := call("show", memory.ID, "--version", memory.Version)
				if got := graphMixedResult[graphstore.Record](t, out); !reflect.DeepEqual(got, memory) {
					t.Fatalf("exact Memory: %+v want=%+v", got, memory)
				}
				exactBefore[memory.ID+"@"+memory.Version] = out
			}
			issueExact := call("show", issue.ID, "--version", issue.Version)
			issueOld := graphMixedResult[graphstore.IssueRecord](t, issueExact)
			if issueOld.ID != issue.ID || issueOld.Version != issue.Version || issueOld.Properties == nil || issueOld.Properties.Title != "Surviving Issue" {
				t.Fatal("exact Issue lost saved identity or title")
			}
			for _, link := range []graphstore.LinkRecord{firstLink, changedLink, survivingLink} {
				if got := graphMixedResult[graphstore.LinkRecord](t, call("show", link.ID, "--version", link.Version)); !reflect.DeepEqual(got, link) {
					t.Fatalf("exact Link: %+v want=%+v", got, link)
				}
			}
			before := graphMemoryReadSnapshot(t, work)
			all := discover("")
			if !all.Complete || all.Next != nil || len(all.Items) != 3 || all.Items[0].ID != contextMemory.ID || all.Items[1].ID != empty.ID || all.Items[2].ID != current.ID || all.Items[2].Version != current.Version || all.Items[2].Details == nil || all.Items[2].Details.OwnedLinkCount != 1 {
				t.Fatalf("complete current discovery: %+v", all)
			}
			matched := discover("BODY-SENTINEL")
			if len(matched.Items) != 1 || matched.Items[0].ID != current.ID || !reflect.DeepEqual(matched.Items[0].MatchedFields, []string{"body"}) || matched.Items[0].Excerpt == nil || matched.Items[0].Excerpt.Text != properties.Body || matched.Items[0].Excerpt.Truncated {
				t.Fatalf("body excerpt: %+v", matched)
			}
			if got := discover("STRASSE"); len(got.Items) != 0 || !got.Complete || got.Next != nil {
				t.Fatalf("discovery used old state: %+v", got)
			}
			recall(body, selected.Items[0].ID, "--version", selected.Items[0].Version, "--readonly")
			recall(properties.Body, current.ID, "--quiet", "--readonly")
			contentChanges := []graphstore.VersionChange{
				{Area: "properties", Member: "body", From: graphMemoryReadValue(t, body), To: graphMemoryReadValue(t, properties.Body)},
				{Area: "properties", Member: "title", From: graphMemoryReadValue(t, original.Properties.Title), To: graphMemoryReadValue(t, properties.Title)},
			}
			forward := compare(original, edited, contentChanges)
			reverse := []graphstore.VersionChange{contentChanges[0], contentChanges[1]}
			for i := range reverse {
				reverse[i].From, reverse[i].To = reverse[i].To, reverse[i].From
			}
			compare(edited, original, reverse)
			compare(current, current, []graphstore.VersionChange{})
			compare(edited, linked, []graphstore.VersionChange{{Area: "owned", ID: firstLink.ID, From: graphstore.VersionValue{}, To: graphMemoryReadValue(t, firstLink)}})
			compare(linked, current, []graphstore.VersionChange{{Area: "owned", ID: firstLink.ID, From: graphMemoryReadValue(t, firstLink), To: graphMemoryReadValue(t, changedLink)}})
			linkComparison := call("compare", firstLink.ID, "--from", firstLink.Version, "--to", changedLink.Version)
			graphMemoryReadComparison(t, linkComparison, graphstore.VersionComparison{Resource: graphstore.VersionComparisonResource{ID: firstLink.ID, Type: firstLink.Type, Source: firstLink.Source, Target: firstLink.Target}, From: graphstore.VersionComparisonEndpoint{Version: firstLink.Version, Attribution: firstLink.Attribution}, To: graphstore.VersionComparisonEndpoint{Version: changedLink.Version, Attribution: changedLink.Attribution}, Compared: []string{"properties"}, Unsupported: []string{"commonMetadata"}, Changes: []graphstore.VersionChange{{Area: "properties", Member: "note", From: graphMemoryReadValue(t, "first"), To: graphMemoryReadValue(t, "second")}}})
			graphMemoryReadComparison(t, call("compare", issue.ID, "--from", issue.Version, "--to", issue.Version), graphstore.VersionComparison{Resource: graphstore.VersionComparisonResource{ID: issue.ID, Type: issue.Type}, From: graphstore.VersionComparisonEndpoint{Version: issueOld.Version, Attribution: issueOld.Attribution}, To: graphstore.VersionComparisonEndpoint{Version: issueOld.Version, Attribution: issueOld.Attribution}, Compared: []string{"properties", "owned"}, Unsupported: []string{"commonMetadata"}, Changes: []graphstore.VersionChange{}})
			for _, token := range []string{"unknown-token", contextMemory.Version, "", strings.Repeat("x", 4097), string([]byte{255})} {
				code := "revision_unknown"
				if token == "" || len(token) > 4096 || token == string([]byte{255}) {
					code = "invalid_selector"
				}
				graphPolicyCLI(t, bd, work, home, nil, code, "show", current.ID, "--version", token, "--json")
				graphMemoryReadRaw(t, bd, work, home, "", code, "recall", current.ID, "--version", token)
				graphPolicyCLI(t, bd, work, home, nil, code, "compare", current.ID, "--from", original.Version, "--to", token, "--json")
			}
			graphPolicyCLI(t, bd, work, home, nil, "revision_unknown", "compare", current.ID, "--from", "unknown-token", "--to", "unknown-token", "--json")
			graphPolicyCLI(t, bd, work, home, nil, "capability_unavailable", "recall", current.ID, "--json")
			graphMemoryReadRaw(t, bd, work, home, "", "capability_unavailable", "recall", issue.ID)
			graphMemoryReadRaw(t, bd, work, home, "", "capability_unavailable", "recall", changedLink.ID)
			graphMemoryReadRaw(t, bd, work, home, "", "not_found", "recall", "beads/missing")
			graphMemoryReadRaw(t, bd, work, home, "", "invalid_selector", "recall", "https://foreign.invalid/beads/plan")
			if got := discover("", "--readonly"); !reflect.DeepEqual(got, all) {
				t.Fatal("readonly discovery differs")
			}
			if got := compare(original, edited, contentChanges, "--readonly"); got != forward {
				t.Fatal("readonly compare differs")
			}
			writeFile(t, filepath.Join(work, "mayor", "town.json"), []byte("{}\n"))
			freeze := filepath.Join(work, "MIGRATION-FREEZE")
			writeFile(t, freeze, []byte("memory-reads\t2026-09-28T00:00:00Z\tread policy\n"))
			if got := discover(""); !reflect.DeepEqual(got, all) {
				t.Fatal("frozen discovery differs")
			}
			if got := compare(original, edited, contentChanges); got != forward {
				t.Fatal("frozen comparison differs")
			}
			recall(properties.Body, current.ID)
			recall(body, original.ID, "--version", original.Version)
			if got := call("show", original.ID, "--version", original.Version); got != exactBefore[original.ID+"@"+original.Version] {
				t.Fatal("frozen exact read differs")
			}
			if err := os.Remove(freeze); err != nil {
				t.Fatal(err)
			}
			graphMemoryReadBrokenPipe(t, bd, work, home, current.ID)
			recall(properties.Body, current.ID)
			if after := graphMemoryReadSnapshot(t, work); !reflect.DeepEqual(after, before) {
				t.Fatal("reads/refusals changed complete current snapshot or writer token")
			}
			for _, memory := range saved {
				if out := call("show", memory.ID, "--version", memory.Version); out != exactBefore[memory.ID+"@"+memory.Version] {
					t.Fatal("reads changed retained Memory")
				}
			}
			if out := call("show", issue.ID, "--version", issue.Version); out != issueExact {
				t.Fatal("reads changed retained Issue")
			}

			unlinked := graphMixedResult[graphstore.LinkDeleteResult](t, call("unlink", changedLink.ID, "--if-revision", changedLink.Revision, "--if-source-revision", current.Revision))
			if !unlinked.Changed || unlinked.Link.State != "deleted" || unlinked.Link.PreviousVersion != changedLink.Version {
				t.Fatalf("unlink receipt: %+v", unlinked)
			}
			graphPolicyCLI(t, bd, work, home, nil, "gone", "show", changedLink.ID, "--version", unlinked.Link.Version, "--json")
			final := graphMixedResult[graphstore.Record](t, call("show", current.ID))
			if len(final.Owned) != 0 || final.Properties != current.Properties {
				t.Fatal("unlink did not preserve final live body")
			}
			removed := []graphstore.VersionChange{{Area: "owned", ID: changedLink.ID, From: graphMemoryReadValue(t, changedLink), To: graphstore.VersionValue{}}}
			compare(current, final, removed)
			call("delete", final.ID, "--force", "--if-revision", final.Revision)
			graphPolicyCLI(t, bd, work, home, nil, "gone", "show", final.ID, "--json")
			graphMemoryReadRaw(t, bd, work, home, "", "gone", "recall", final.ID)
			remaining := discover("")
			if !remaining.Complete || remaining.Next != nil || len(remaining.Items) != 2 || remaining.Items[0].ID != contextMemory.ID || remaining.Items[1].ID != empty.ID {
				t.Fatalf("deleted discovery: %+v", remaining)
			}
			for _, memory := range append(saved, final) {
				if got := graphMixedResult[graphstore.Record](t, call("show", memory.ID, "--version", memory.Version)); !reflect.DeepEqual(got, memory) {
					t.Fatalf("retained after delete: %+v want=%+v", got, memory)
				}
				recall(memory.Properties.Body, memory.ID, "--version", memory.Version, "--readonly", "--quiet")
			}
			compare(current, final, removed)
			compare(final, current, []graphstore.VersionChange{{Area: "owned", ID: changedLink.ID, From: graphstore.VersionValue{}, To: graphMemoryReadValue(t, changedLink)}})
			compare(final, final, []graphstore.VersionChange{})
			for _, link := range []graphstore.LinkRecord{firstLink, changedLink, survivingLink} {
				if got := graphMixedResult[graphstore.LinkRecord](t, call("show", link.ID, "--version", link.Version)); !reflect.DeepEqual(got, link) {
					t.Fatal("delete changed retained Link")
				}
			}
			if out := call("show", issue.ID, "--version", issue.Version); out != issueExact {
				t.Fatal("delete changed retained Issue")
			}
			if out := call("show", issue.ID); out != issueCurrent {
				t.Fatal("delete changed surviving current Issue")
			}
			if got := graphMixedResult[graphstore.Record](t, call("show", contextMemory.ID)); !reflect.DeepEqual(got, contextMemory) {
				t.Fatal("delete changed surviving Memory")
			}
			if got := graphMixedResult[graphstore.LinkRecord](t, call("show", survivingLink.ID)); !reflect.DeepEqual(got, survivingLink) {
				t.Fatal("delete changed surviving unowned Link")
			}
			graphPolicyCLI(t, bd, work, home, nil, "identity_reserved", "remember", "cannot reuse", "--id", "beads/plan", "--title", "Reserved", "--json")
		})
	}
}
