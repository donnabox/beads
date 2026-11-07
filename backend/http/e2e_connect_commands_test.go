//go:build cgo

package bdhttp_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/httpapi"
)

// TestE2E_ConnectCommandParity runs, through the real `bd` binary against an
// in-process `bd serve`, every command that used to reach the refusing shell:
// some refused, and some printed the refusal, wrote nothing and exited 0.
//
// It runs on BDP, which is where all of them are fixed. On v0 the ones that used to
// answer, however badly, stay as they were (cmd/bd's remoteOnV0 and wire_proof.py).
//
// A dial is named both ways, by its v0 path and by its Callable, because which of
// the two the client speaks is the wire's business and these are about the command.
func TestE2E_ConnectCommandParity(t *testing.T) {
	skipUnlessEmbeddedDolt(t)
	bin := buildBD(t)
	addr := startProtocolServer(t, httpapi.ProtocolBoth)
	proxyAddr, counter := startCountingProxy(t, addr)
	f := &remoteRolesFixture{t: t, bin: bin, dir: t.TempDir(), ids: map[string]string{}}
	f.mustRun(nil, "connect", "http://"+proxyAddr, "--protocol", "bdp", "--expect-project-id", e2eProjectID, "--json")

	dialed := func(t *testing.T, callable, v0 string) {
		t.Helper()
		for _, request := range counter.snapshot() {
			if strings.Contains(request, "/bdp/call/"+callable) || strings.Contains(request, v0) {
				return
			}
		}
		t.Errorf("no request reached %s (%s): %v", callable, v0, counter.snapshot())
	}
	show := func(t *testing.T, id string) map[string]any {
		t.Helper()
		out := f.mustRun(nil, "show", id, "--json").stdout
		var rows []map[string]any
		if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &rows); err != nil {
			t.Fatalf("parse bd show: %v\n%s", err, out)
		}
		if len(rows) != 1 {
			t.Fatalf("bd show %s answered %d rows: %s", id, len(rows), out)
		}
		return rows[0]
	}
	refused := func(t *testing.T, want string, args ...string) {
		t.Helper()
		r := f.run(nil, args...)
		if r.code == 0 {
			t.Fatalf("bd %v exited 0: stdout=%s stderr=%s", args, r.stdout, r.stderr)
		}
		if !strings.Contains(r.stdout+r.stderr, want) {
			t.Errorf("bd %v: want %q in stdout=%s stderr=%s", args, want, r.stdout, r.stderr)
		}
	}

	t.Run("defer and undefer write", func(t *testing.T) {
		id := f.create("defer me")
		counter.reset()
		f.mustRun(nil, "defer", id, "--until", "2099-01-01T00:00:00Z", "--reason", "later")
		dialed(t, "issue/update", "PATCH /v0/beads/issues/"+id)
		row := show(t, id)
		if row["status"] != "deferred" || !strings.Contains(row["notes"].(string), "later") || row["defer_until"] == nil {
			t.Fatalf("after bd defer: %v", row)
		}
		f.mustRun(nil, "undefer", id)
		if row := show(t, id); row["status"] != "open" || row["defer_until"] != nil {
			t.Fatalf("after bd undefer: %v", row)
		}
	})

	t.Run("a per-id failure makes the exit nonzero", func(t *testing.T) {
		id := f.create("not deferred")
		refused(t, "is not deferred", "undefer", id)
	})

	t.Run("graph reports dependents", func(t *testing.T) {
		parent := f.create("graph root")
		child := f.create("graph dependent")
		f.mustRun(nil, "dep", "add", child, parent)
		out := f.mustRun(nil, "graph", parent, "--json").stdout
		if !strings.Contains(out, child) {
			t.Fatalf("bd graph %s does not name its dependent %s: %s", parent, child, out)
		}
	})

	t.Run("ready --explain and the epic annotation", func(t *testing.T) {
		blocker := f.create("explain blocker")
		blocked := f.create("explain blocked")
		f.mustRun(nil, "dep", "add", blocked, blocker)
		out := f.mustRun(nil, "ready", "--explain", "--json").stdout
		if !strings.Contains(out, "explain blocker") {
			t.Fatalf("bd ready --explain lost the blocker's title: %s", out)
		}
		epic := strings.TrimSpace(f.mustRun(nil, "create", "annotated epic", "-t", "epic", "--silent").stdout)
		f.mustRun(nil, "create", "annotated child", "--parent", epic, "--silent")
		if text := f.mustRun(nil, "ready").stdout; !strings.Contains(text, "annotated epic") {
			t.Fatalf("bd ready lost the parent epic: %s", text)
		}
	})

	t.Run("show --refs and --children carry whole rows", func(t *testing.T) {
		epic := strings.TrimSpace(f.mustRun(nil, "create", "refs epic", "-t", "epic", "--silent").stdout)
		strings.TrimSpace(f.mustRun(nil, "create", "refs child", "--parent", epic, "-d", "the body", "--silent").stdout)
		for _, flag := range []string{"--refs", "--children"} {
			counter.reset()
			out := f.mustRun(nil, "show", epic, flag, "--json").stdout
			if strings.Contains(out, "0001-01-01") || !strings.Contains(out, "the body") {
				t.Errorf("bd show %s came back shallow: %s", flag, out)
			}
			dialed(t, "issue/get-many", ":batchGet")
		}
	})

	t.Run("the kv family refuses", func(t *testing.T) {
		for _, args := range [][]string{{"kv", "get", "k"}, {"kv", "list"}, {"kv", "set", "k", "v"}, {"kv", "clear", "k"}} {
			refused(t, "not available over bd connect", args...)
		}
	})

	t.Run("unclaim dials the release", func(t *testing.T) {
		id := f.create("unclaim me")
		f.mustRun(nil, "update", id, "--claim")
		counter.reset()
		f.mustRun(nil, "unclaim", id, "--reason", "changed my mind")
		dialed(t, "issue/unclaim", ":release")
		row := show(t, id)
		if row["status"] != "open" || row["assignee"] != nil {
			t.Fatalf("after bd unclaim: %v", row)
		}
		if comments := f.mustRun(nil, "comments", id).stdout; !strings.Contains(comments, "changed my mind") {
			t.Errorf("the reason did not land as a comment: %s", comments)
		}
	})

	t.Run("a bare --claim dials claimIssue", func(t *testing.T) {
		id := f.create("claim me")
		counter.reset()
		out := f.mustRun(nil, "update", id, "--claim", "--json").stdout
		dialed(t, "issue/claim", ":claim")
		if !strings.Contains(out, "in_progress") {
			t.Errorf("bd update --claim --json: %s", out)
		}
		// A claim beside a field edit stays one write.
		other := f.create("claim me with a title")
		counter.reset()
		f.mustRun(nil, "update", other, "--claim", "--title", "claimed and renamed")
		if row := show(t, other); row["title"] != "claimed and renamed" || row["status"] != "in_progress" {
			t.Fatalf("after a claim with an edit: %v", row)
		}
	})

	t.Run("priority, note and tag", func(t *testing.T) {
		id := f.create("edit me")
		f.mustRun(nil, "priority", id, "1")
		f.mustRun(nil, "note", id, "a note")
		f.mustRun(nil, "tag", id, "t1")
		row := show(t, id)
		if row["priority"] != float64(1) || !strings.Contains(row["notes"].(string), "a note") {
			t.Fatalf("after priority and note: %v", row)
		}
		if labels, _ := row["labels"].([]any); !slices.Contains(labels, any("t1")) {
			t.Fatalf("after tag: %v", row)
		}
	})

	t.Run("q, q --parent and todo add", func(t *testing.T) {
		id := strings.TrimSpace(f.mustRun(nil, "q", "quick capture", "-p", "1", "-l", "x").stdout)
		if row := show(t, id); row["title"] != "quick capture" || row["priority"] != float64(1) {
			t.Fatalf("after bd q: %v", row)
		}
		parent := strings.TrimSpace(f.mustRun(nil, "create", "q parent", "-t", "epic", "-l", "inherited", "--silent").stdout)
		child := strings.TrimSpace(f.mustRun(nil, "q", "quick child", "--parent", parent).stdout)
		if !strings.HasPrefix(child, parent+".") {
			t.Errorf("bd q --parent minted %s under %s", child, parent)
		}
		if labels, _ := show(t, child)["labels"].([]any); !slices.Contains(labels, any("inherited")) {
			t.Errorf("bd q --parent did not inherit labels: %v", labels)
		}
		f.mustRun(nil, "todo", "add", "a todo", "-d", "todo body")
		if out := f.mustRun(nil, "todo", "--json").stdout; !strings.Contains(out, "a todo") {
			t.Errorf("bd todo: %s", out)
		}
	})

	t.Run("duplicate, supersede, dep relate and unrelate", func(t *testing.T) {
		a, b := f.create("duplicate of"), f.create("canonical")
		f.mustRun(nil, "duplicate", a, "--of", b)
		if show(t, a)["status"] != "closed" {
			t.Errorf("bd duplicate did not close %s", a)
		}
		c, d := f.create("superseded"), f.create("replacement")
		f.mustRun(nil, "supersede", c, "--with", d)
		if show(t, c)["status"] != "closed" {
			t.Errorf("bd supersede did not close %s", c)
		}
		e, g := f.create("relate one"), f.create("relate two")
		f.mustRun(nil, "dep", "relate", e, g)
		if !depListHas(t, f.mustRun(nil, "dep", "list", e, "--json").stdout, g, "related") &&
			!depListHas(t, f.mustRun(nil, "dep", "list", e, "--json").stdout, g, "relates-to") {
			t.Errorf("bd dep relate left no edge: %s", f.mustRun(nil, "dep", "list", e, "--json").stdout)
		}
		f.mustRun(nil, "dep", "unrelate", e, g)
		if out := f.mustRun(nil, "dep", "list", e, "--json").stdout; strings.Contains(out, g) {
			t.Errorf("bd dep unrelate left the edge: %s", out)
		}
	})

	t.Run("text delete previews", func(t *testing.T) {
		leaf := f.create("delete preview leaf")
		if out := f.mustRun(nil, "delete", leaf).stdout; !strings.Contains(out, "delete preview leaf") {
			t.Errorf("bd delete's preview does not name the issue: %s", out)
		}
		show(t, leaf)

		id := f.create("delete me")
		target := f.create("delete target")
		f.mustRun(nil, "dep", "add", id, target)
		if out := f.mustRun(nil, "delete", id, "--dry-run").stdout; !strings.Contains(out, target) {
			t.Errorf("bd delete --dry-run does not name the edge's target: %s", out)
		}
		show(t, id)
	})

	t.Run("create --file", func(t *testing.T) {
		file := filepath.Join(f.dir, "many.md")
		if err := os.WriteFile(file, []byte("## first from md\n\n## second from md\n\n### Priority\n1\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		out := f.mustRun(nil, "create", "--file", file, "--json").stdout
		if !strings.Contains(out, "first from md") || !strings.Contains(out, "second from md") {
			t.Errorf("bd create --file: %s", out)
		}
	})

	t.Run("import creates, and refuses what it cannot carry", func(t *testing.T) {
		file := filepath.Join(f.dir, "ids.jsonl")
		line := `{"id":"e2e-imp1","title":"imported one","issue_type":"task","status":"open","priority":2}` + "\n"
		if err := os.WriteFile(file, []byte(line), 0o600); err != nil {
			t.Fatal(err)
		}
		counter.reset()
		f.mustRun(nil, "import", "--input", file, "--json")
		dialed(t, "issue/batch", ":batchApply")
		if row := show(t, "e2e-imp1"); row["title"] != "imported one" {
			t.Fatalf("after bd import: %v", row)
		}

		titles := filepath.Join(f.dir, "titles.jsonl")
		if err := os.WriteFile(titles, []byte(`{"title":"imported by title","issue_type":"task","priority":2}`+"\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		counter.reset()
		f.mustRun(nil, "import", "--input", titles, "--json")
		dialed(t, "issue/create-many", ":batchCreate")

		stamped := filepath.Join(f.dir, "stamped.jsonl")
		stampedLine := `{"id":"e2e-imp2","title":"stamped","issue_type":"task","status":"open","priority":2,"created_at":"2020-01-01T00:00:00Z","updated_at":"2020-01-01T00:00:00Z"}` + "\n"
		if err := os.WriteFile(stamped, []byte(stampedLine), 0o600); err != nil {
			t.Fatal(err)
		}
		refused(t, "CreatedAt", "import", "--input", stamped)
	})

	t.Run("batch applies as one request", func(t *testing.T) {
		a, b, c := f.create("batch one"), f.create("batch two"), f.create("batch three")
		script := filepath.Join(f.dir, "plan.txt")
		plan := "create task 2 \"batch created\"\nupdate " + a + " title=changed\nclose " + b + " because\ndep add " + c + " " + a + " blocks\n"
		if err := os.WriteFile(script, []byte(plan), 0o600); err != nil {
			t.Fatal(err)
		}
		counter.reset()
		f.mustRun(nil, "batch", "-f", script)
		dialed(t, "issue/batch", ":batchApply")
		if row := show(t, a); row["title"] != "changed" {
			t.Errorf("after bd batch: %v", row)
		}
		if row := show(t, b); row["status"] != "closed" || row["close_reason"] != "because" {
			t.Errorf("after bd batch: %v", row)
		}
	})

	t.Run("a batch dep remove refuses before anything is sent", func(t *testing.T) {
		a, b := f.create("no batch one"), f.create("no batch two")
		script := filepath.Join(f.dir, "remove.txt")
		if err := os.WriteFile(script, []byte("update "+a+" title=must-not-land\ndep remove "+a+" "+b+"\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		refused(t, "line 2", "batch", "-f", script)
		if row := show(t, a); row["title"] == "must-not-land" {
			t.Errorf("the batch wrote before refusing: %v", row)
		}
	})

	t.Run("listings through a raw search", func(t *testing.T) {
		for _, args := range [][]string{
			{"lint", "--json"}, {"orphans", "--json"}, {"duplicates", "--json"}, {"find-duplicates", "--json"},
			{"human", "list", "--json"}, {"label", "list-all", "--json"}, {"swarm", "list", "--json"}, {"graph", "--all", "--json"},
			{"ready", "--gated", "--json"}, {"mol", "ready", "--json"}, {"stale", "--json"}, {"epic", "status", "--json"}, {"mol", "stale", "--json"},
		} {
			r := f.run(nil, args...)
			if strings.Contains(r.stdout+r.stderr, "not supported") || strings.Contains(r.stdout+r.stderr, "not expressible") {
				t.Errorf("bd %v still refuses: stdout=%s stderr=%s", args, r.stdout, r.stderr)
			}
		}
	})

	t.Run("events read and pruning refuses", func(t *testing.T) {
		// This server does not journal, so the read reaches it and is told so. The
		// rows themselves are TestServedJournalCarriesTheActor's.
		counter.reset()
		refused(t, "events_journal_disabled", "events", "tail")
		dialed(t, "events/list", "/v0/beads/events")
		refused(t, "events_journal_disabled", "events", "export")
		refused(t, "not available over bd connect", "events", "prune", "--before", "1")
	})

	t.Run("a close reason keeps its newline", func(t *testing.T) {
		id := f.create("close with a reason file")
		file := filepath.Join(f.dir, "reason.txt")
		if err := os.WriteFile(file, []byte("two\nlines\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		f.mustRun(nil, "close", id, "--reason-file", file)
		if row := show(t, id); row["close_reason"] != "two\nlines\n" {
			t.Fatalf("after bd close --reason-file: %q", row["close_reason"])
		}
	})

	t.Run("close --claim-next", func(t *testing.T) {
		id := f.create("close then claim")
		f.create("next up")
		counter.reset()
		out := f.mustRun(nil, "close", id, "--claim-next", "--json").stdout
		dialed(t, "issue/claim-next", ":claimNext")
		if !strings.Contains(out, "claimed") {
			t.Errorf("bd close --claim-next --json: %s", out)
		}
	})

	t.Run("sql, gc and the version-control family say why", func(t *testing.T) {
		refused(t, "over bd connect", "sql", "select 1")
		refused(t, "over bd connect", "gc", "--dry-run")
		refused(t, "over bd connect", "sync")
		refused(t, "over bd connect", "show", "e2e-imp1", "--as-of", "HEAD", "--json")
		for _, args := range [][]string{{"history", "e2e-imp1"}, {"branch"}, {"vc", "status"}, {"dolt", "remote", "list"}} {
			r := f.run(nil, args...)
			if r.code == 0 {
				t.Errorf("bd %v exited 0: %s", args, r.stdout)
			}
		}
	})
}
