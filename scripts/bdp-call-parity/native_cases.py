"""Part B's cases: the native command behind each Callable, with its flag variants.

The mapping of Callable to command was written from the spec's cliMapping and from
cmd/bd, then checked against the serve log: each result records the v0 operations
the command actually dialed, so `ops` in results.json is the evidence for the row.

Phase "read" runs on one fixture no write ever touches. Phase "write" gives each
case its own issues. Phase "global" commands act on the whole database, so each runs
on a fresh pair of workspaces.
"""

import shlex

R1, R2, R3, R4, R5, R6, R7, R8, R9 = [f"sw-fix-r{i}" for i in range(1, 10)]
RA = "sw-fix-ra"

SEP = "\x00"


def create(id_, *flags):
    """A seed: `bd create` in direct mode, on both sides, so both hold the same rows."""
    title = [] if "--title" in flags else [id_]
    return ["create"] + title + ["--id", id_, "-t", "task", "-p", "2", "--silent"] + list(flags)


READ_SEED = [
    create(R1, "--title", "alpha one", "-p", "1", "-a", "alice", "-l", "red", "--metadata", '{"k":"v"}',
           "-d", "needle in description", "--notes", "needle in notes", "--external-ref", "ext-1",
           "--due", "2030-01-01T00:00:00Z"),
    create(R2, "--title", "alpha two", "-l", "blue", "--deps", f"blocks:{R1}"),
    create(R3, "--title", "beta closed", "-p", "3"),
    create(R4, "--title", "the epic", "-t", "epic", "-l", "red,blue"),
    create(R5, "--title", "child of epic", "--parent", R4),
    create(R6, "--title", "claimed by bob", "-t", "bug", "-p", "0"),
    create(R7, "--title", "commented"),
    create(R8, "--title", "deferred", "--defer", "2099-01-01T00:00:00Z"),
    create(R9, "--title", "ephemeral", "--ephemeral"),
    create(RA, "--title", "related out", "--deps", f"related:{R7}"),
    create("sw-fix-rb", "--title", "pinned one", "--status", "pinned"),
    create("sw-fix-rc", "--title", "overdue one", "--due", "2001-01-01T00:00:00Z"),
    create("sw-fix-rd", "--title", "swarm one", "--mol-type", "swarm"),
    create("sw-fix-re", "--title", "no labels no description", "--allow-empty-description"),
]
READ_BEFORE = [
    ["close", R3, "--reason", "done"],
    ["update", R6, "--claim", "--actor", "bob"],
    ["comment", R7, "first comment", "--actor", "alice"],
    ["config", "set", "fix.read", "v1"],
    ["kv", "set", "fix-kv", "kv-value"],
    ["remember", "remembered thing", "--key", "fix-read"],
]


def N(callables, command, seed=(), before=(), files=None, stdin=None, label=None, phase="write", actor=None, plain=False, wait=0):
    return {"callables": [callables] if isinstance(callables, str) else list(callables),
            "argv": shlex.split(command), "seed": list(seed), "before": list(before), "files": files or {},
            "stdin": stdin, "label": label or command, "phase": phase, "actor": actor, "plain": plain, "wait": wait,
            "mode": "text" if plain else "json"}


def reads():
    c = []

    def r(callables, *commands):
        for command in commands:
            c.append(N(callables, command, phase="read"))

    r("issue/context", "context", "where", "info", "ping", "version")
    r("issue/ready", "ready", "ready -n 2", "ready -p 1", "ready -a alice", "ready -u", "ready -s priority", "ready -s oldest",
      "ready -s hybrid", "ready -l red", "ready -t bug", "ready --label-any red,blue", "ready --exclude-label red",
      "ready --exclude-type epic", "ready --label-pattern 're*'", "ready --label-regex '^r'", f"ready --parent {R4}",
      "ready --metadata-field k=v", "ready --has-metadata-key k", "ready --include-ephemeral", "ready --include-deferred",
      "ready --brief", "ready --explain", "ready --gated", f"ready --mol {R4}", "ready --mol-type swarm", "ready --plain",
      "ready --pretty", "ready --flat", "ready --offset 1", "ready --max-rows 2", "ready --max-rows 99")
    # The text listing asks the ReadyCounter for "Showing N of M" when the page is full.
    c.append(N("issue/ready-count", "ready -n 1", phase="read", plain=True))
    c.append(N("issue/ready-count", "ready -n 1 -l red", phase="read", plain=True))
    r("issue/stats", "stats", "status", "status --assigned", "status --no-activity", "status --no-blocked", "status --all", "stats --no-blocked")
    r("issue/blocked", "blocked", f"blocked --parent {R4}", "blocked --label blue", "blocked --label-any blue,red",
      "blocked --exclude-label blue")
    # --flags reads its ids from stdin.
    for command in ("blocked --flags", "blocked --flags --raw"):
        c.append(N("issue/blocking-annotations", command, stdin=f"{R1}\n{R2}\n{R7}\n", phase="read"))
    r("issue/dep/cycles", "dep cycles", "dep cycles --include-tracks", "dep cycles --limit 1")
    r("issue/dep/tree", f"dep tree {R2}", f"dep tree {R1} --direction up", f"dep tree {R2} --direction down",
      f"dep tree {R2} --direction both", f"dep tree {R2} --max-depth 1", f"dep tree {R2} --status open",
      f"dep tree {R2} --reverse", f"dep tree {R2} --show-all-paths", f"dep tree {R2} --format mermaid",
      f"dep tree {R2} --max-rows 5", "dep tree sw-fix-none", f"graph {R2}", f"graph {R2} --compact", "graph --all", f"graph {R2} --dot")
    r("issue/list", "list", "list --status open", "list --status open,closed", "list -t bug", "list -a alice", "list -l red",
      "list --label-any red,blue", "list --exclude-label red", f"list --parent {R4}", "list --all", "list --include-templates",
      "list --include-gates", "list --include-infra", "list --include-ephemeral", "list --created-before 2099-01-01",
      "list --created-after 2001-01-01", "list --metadata-field k=v", "list --has-metadata-key k", "list --no-assignee",
      "list --title 'alpha one'", f"list --id {R1},{R2}", "list --title-contains alpha", "list --desc-contains needle",
      "list --notes-contains needle", "list --external-contains ext", "list --external-ref ext-1", "list --spec x",
      "list --label-pattern 're*'", "list --label-regex '^r'", "list --updated-after 2001-01-01", "list --updated-before 2099-01-01",
      "list --closed-after 2001-01-01", "list --closed-before 2099-01-01", "list --defer-after 2001-01-01", "list --defer-before 2099-12-31",
      "list --due-after 2001-01-01", "list --due-before 2099-01-01", "list -p 1", "list --priority-min 0", "list --priority-max 1",
      "list --empty-description", "list --no-labels", "list --no-parent", "list --exclude-type epic", "list --pinned", "list --no-pinned",
      "list --deferred", "list --overdue", "list --mol-type swarm", "list --wisp-type patrol --include-ephemeral", "list --sort priority", "list --sort created",
      "list --sort updated", "list --sort title", "list --sort id", "list --reverse", "list -n 2", "list --offset 1", "list --brief", "list --long",
      "list --flat", "list --tree", "list --pretty", "list --ready", "list --include-comments", "list --skip-labels",
      "list --format digraph", "list --format dot", "list --max-rows 3", "list --max-rows 99", "search alpha", "search --query alpha", "search alpha --status open",
      "search alpha -a alice", "search alpha -l red", "search alpha --limit 1", "search alpha --sort priority", "search alpha --long",
      f"children {R4}", f"children {R4} --pretty", "gate list", "gate list --all", "mol stale", "mol ready", "epic status", "epic status --eligible-only",
      "stale", "stale --days 1", "stale --status open", "stale -l red", "stale --limit 1", "orphans", "lint", "todo", "todo list", "duplicates", "find-duplicates",
      "human list", "swarm list", "label list-all", "export", "export --all", "export --include-infra", "export --include-memories")
    r("issue/query", "query status=open", "query 'priority<2' -a", "query status=open --sort priority", "query status=open --sort priority -r",
      "query status=open -n 1", "query status=open --offset 1", "query status=open --long", "query status=open --parse-only", "query 'this is = not ( valid'")
    r("issue/count", "count", "count --status open", "count -t bug", "count -a alice", "count -p 1", "count --priority-min 0", "count --priority-max 1",
      "count -l red", "count --label-any red,blue", "count --title 'alpha one'", f"count --id {R1}", "count --title-contains alpha",
      "count --desc-contains needle", "count --notes-contains needle", "count --created-after 2001-01-01", "count --created-before 2099-01-01",
      "count --updated-after 2001-01-01", "count --updated-before 2099-01-01", "count --closed-after 2001-01-01", "count --closed-before 2099-01-01",
      "count --empty-description", "count --no-assignee", "count --no-labels", "count --metadata-field k=v", "count --include-ephemeral",
      "count --has-metadata-key k", "count --include-infra", f"count --parent {R4}", "count --no-parent", "count --exclude-type epic",
      "count --exclude-status closed", "count --by-status", "count --by-priority", "count --by-type", "count --by-assignee", "count --by-label")
    r("issue/show", f"show {R1}", f"show {R1} {R2}", f"show {R1} --long", f"show {R2} --refs", f"show {R4} --children",
      f"show {R7} --include-comments", f"show {R7} --comments-tail 1", f"show {R1} --include-dependents", f"show {R2} --brief-deps",
      f"show {R7} --thread", f"show --id {R1}", f"show {R1} --local-time", f"show {R1} --as-of HEAD", "show --current", "show sw-fix-none", f"show {R1} sw-fix-none")
    for command in (f"show {R1} --short", f"show {R1} {R2} --short", "list --deps"):
        c.append(N("issue/show", command, phase="read", plain=True))
    r("issue/get-many", f"mol show {R4}", f"ready --mol {R4}", f"mol progress {R4}")
    r("issue/dep/list", f"dep list {R2}", f"dep list {R1} --direction up", f"dep list {R2} --direction down", f"dep list {RA} -t related",
      f"dep list {R1} {R2}", "dep list sw-fix-none", f"label list {R1}", f"comments {R7}", f"comments {R7} --local-time")
    r("issue/dep/edges", f"dep list {R2} -t blocks", f"show {R2}")
    r("config/list", "config list", "statuses", "types")
    r("config/get", "config get fix.read", "config get fix.absent", "config get issue_prefix")
    r("memory/list", "memories", "memories remembered", "memories nothing-matches")
    r("memory/recall", "recall fix-read", "recall fix-absent")
    r("events/list", "events tail", "events tail --since 0", "events tail --limit 2", "events export")
    # Reads with no Callable of their own, kept for the separate list.
    r("(none)", "kv list", "kv get fix-kv", "kv get fix-absent", f"state {R1} phase", f"history {R1}", f"versions {R1}", "diff HEAD~1 HEAD", "vc status", "branch", "dolt status",
      "dolt remote list", "sql 'select 1'", "backup status", "doctor", "prime", "schema", "conflicts", f"provenance {R1}", "merge-slot check",
      "federation list", "rules audit", "preflight", "mol progress " + R4, "mol current", f"mol last-activity {R4}", "compact --help")
    return c


def writes():
    c = []
    n = 0

    def fresh():
        nonlocal n
        n += 1
        return f"sw-fix-w{n:03d}"

    def one(callables, template, *flags, before=(), extra_seed=(), **kw):
        """template names its issue as {i}; the case gets a fresh one."""
        i = fresh()
        seed = list(extra_seed)
        seed = [[arg.replace("{i}", i) for arg in row] for row in seed] + [create(i, *flags)]
        c.append(N(callables, template.replace("{i}", i), seed=seed,
                   before=[[arg.replace("{i}", i) for arg in row] for row in before], **kw))

    # issue/create
    def mk(command, **kw):
        c.append(N("issue/create", command, **kw))
    mk("create 'plain minted'")
    mk("create 'with type' -t bug -p 1")
    mk("create 'p string' -p P1")
    mk("create --title 'title flag'")
    mk("create 'with desc' -d text")
    mk("create 'with design' --design ds --acceptance ac --notes nn")
    mk("create 'with append notes' --append-notes nn")
    mk("create 'assigned' -a alice")
    mk("create 'labels' -l a,b")
    mk("create 'with estimate' --estimate 30")
    mk("create 'with external' --external-ref ext-x")
    mk("create 'with due' --due 2030-01-01T00:00:00Z")
    mk("create 'with relative due' --due +6h")
    mk("create 'with defer' --defer 2030-01-01T00:00:00Z")
    mk("create 'metadata' --metadata '{\"a\":1}'")
    mk("create 'ephemeral' --ephemeral")
    mk("create 'no history' --no-history")
    mk("create 'mol type' --mol-type swarm")
    mk("create 'wisp type' --ephemeral --wisp-type patrol")
    mk("create 'spec' --spec-id spec-1")
    mk("create 'status' --status in_progress")
    mk("create 'event' -t event --event-category cat --event-actor a --event-target t --event-payload p")
    mk("create 'skills' --skills s1 --context c1")
    mk("create 'storage ephemeral' --storage-class ephemeral")
    mk("create 'storage unversioned' --storage-class unversioned")
    mk("create 'storage versioned' --storage-class versioned")
    mk("create 'storage bad' --storage-class hot")
    mk("create 'validate' -t epic --validate -d 'has none of the sections'")
    mk("create 'empty ok' --allow-empty-description")
    mk("create 'silent' --silent", plain=True)
    mk("create 'dry' --dry-run")
    mk("create --id sw-fix-explicit 'explicit id'")
    mk("create --id zz-other 'wrong prefix'")
    mk("create --id zz-forced 'forced prefix' --force")
    mk("create 'parented' --parent sw-fix-cp", seed=[create("sw-fix-cp", "-t", "epic", "-l", "inh")])
    mk("create 'not inheriting' --parent sw-fix-cp --no-inherit-labels")
    mk("create 'deps' --deps blocks:sw-fix-cp")
    mk("create 'deps bare' --deps sw-fix-cp")
    mk("create 'deps related' --deps related:sw-fix-cp")
    mk("create 'dep missing' --deps blocks:sw-fix-none")
    mk("create 'parent missing' --parent sw-fix-none")
    mk("create 'waits' --waits-for sw-fix-cp --waits-for-gate all-children")
    mk("create 'body file' --body-file body.txt", files={"body.txt": "from a file\n"})
    mk("create 'design file' --design-file design.txt", files={"design.txt": "design from a file\n"})
    mk("create 'body stdin' --stdin", stdin="from stdin\n")
    mk("create --file many.md", files={"many.md": "## first from md\n\n### Type\ntask\n\n## second from md\n\n### Priority\n1\n"})
    mk("create --graph plan.json", files={"plan.json": '{"nodes":[{"key":"a","title":"graph a"},{"key":"b","title":"graph b"}],"edges":[{"from_key":"b","to_key":"a","type":"blocks"}]}'})
    mk("create")
    mk("create 'bad type' -t nonsense")
    mk("create 'bad priority' -p 99")
    mk("create 'gate type' -t gate")
    c.append(N("issue/create", "q 'quick capture'", plain=True))
    c.append(N("issue/create", "q 'quick typed' -t bug -p 1 -l x", plain=True))
    c.append(N("issue/create", "q 'quick parented' --parent sw-fix-qp", plain=True, seed=[create("sw-fix-qp", "-t", "epic", "-l", "inh")]))
    c.append(N("issue/create", "todo add 'a todo'"))
    c.append(N("issue/create", "todo add 'a todo with p' -p 1 -d text"))
    c.append(N("issue/create", "gate create --help", plain=True))
    # issue/update and the commands that are spellings of it
    u = "issue/update"
    one(u, "update {i} --title new")
    one(u, "update {i} --description d")
    one(u, "update {i} -p 1")
    one(u, "update {i} -p P0")
    one(u, "update {i} --status in_progress")
    one(u, "update {i} --status closed")
    one(u, "update {i} -a bob")
    one(u, "update {i} -a ''", "-a", "bob", label="clear the assignee")
    one(u, "update {i} -t bug")
    one(u, "update {i} --notes n")
    one(u, "update {i} --notes over", "--notes", "existing", label="overwrite notes")
    one(u, "update {i} --notes over --force", "--notes", "existing", label="overwrite notes with force")
    one(u, "update {i} --append-notes more", "--notes", "existing")
    one(u, "update {i} --clear-notes", "--notes", "existing")
    one(u, "update {i} --design d --acceptance a")
    one(u, "update {i} --design-file d.txt", files={"d.txt": "d\n"})
    one(u, "update {i} --body-file d.txt", files={"d.txt": "d\n"})
    one(u, "update {i} --stdin", stdin="from stdin\n")
    one(u, "update {i} --add-label x")
    one(u, "update {i} --remove-label y", "-l", "y")
    one(u, "update {i} --set-labels only", "-l", "y")
    one(u, "update {i} --metadata '{\"a\":1}'")
    one(u, "update {i} --set-metadata a=1")
    one(u, "update {i} --unset-metadata a", "--metadata", '{"a":1}')
    one(u, "update {i} --estimate 5")
    one(u, "update {i} --external-ref e-{i}")
    one(u, "update {i} --due 2030-01-01T00:00:00Z")
    one(u, "update {i} --due +6h")
    one(u, "update {i} --due ''", "--due", "2030-01-01T00:00:00Z", label="clear due")
    one(u, "update {i} --defer 2030-01-01T00:00:00Z")
    one(u, "update {i} --defer ''", "--defer", "2030-01-01T00:00:00Z", label="clear defer")
    one(u, "update {i} --parent sw-fix-up", extra_seed=[create("sw-fix-up-{i}", "-t", "epic")], label="set parent to a missing id")
    one(u, "update {i} --parent sw-fix-up-{i}", extra_seed=[create("sw-fix-up-{i}", "-t", "epic")], label="set parent")
    one(u, "update {i} --parent ''", "--parent", "sw-fix-cp", label="clear parent")
    one(u, "update {i} --await-id x")
    one(u, "update {i} --spec-id s")
    one(u, "update {i} --session sess --title t")
    one(u, "update {i} --ephemeral")
    one(u, "update {i} --persistent", "--ephemeral")
    one(u, "update {i} --no-history")
    one(u, "update {i} --history", "--no-history")
    one(u, "update {i} --if-status open --title t")
    one(u, "update {i} --if-status closed --title t")
    one(u, "update {i} --if-assignee nobody --title t")
    one(u, "update {i} --if-revision 1 --title t")
    one(u, "update {i} --if-revision $rev:{i} --title t")
    one(u, "update {i} --title same", "--title", "same", label="a no-op update")
    one(u, "update {i}", label="no flags")
    one(u, "update {i} sw-fix-none --title t", label="one good id, one missing")
    c.append(N(u, "update sw-fix-none --title t"))
    i1, i2 = fresh(), fresh()
    c.append(N(u, f"update {i1} {i2} --title both", seed=[create(i1), create(i2)]))
    one(u, "assign {i} alice")
    one(u, "assign {i} alice --force", before=[["update", "{i}", "--claim", "--actor", "bob"]])
    one(u, "assign {i} alice", before=[["update", "{i}", "--claim", "--actor", "bob"]], label="assign over a claim")
    one(u, "assign {i} alice --if-revision 1")
    one(u, "assign {i} alice --if-revision $rev:{i}")
    one(u, "priority {i} 1")
    one(u, "priority {i} P1")
    one(u, "tag {i} t1")
    one(u, "note {i} hello")
    one(u, "note {i} --file n.txt", files={"n.txt": "hello\n"})

    one(u, "label add {i} l1")
    one(u, "label remove {i} l1", "-l", "l1")
    one(u, "defer {i}")
    one(u, "defer {i} --until 2030-01-01T00:00:00Z")
    one(u, "defer {i} --reason later")
    one(u, "undefer {i}", "--defer", "2030-01-01T00:00:00Z")
    # The exit-zero family: a per-id failure is a nonzero exit, on both sides.
    one(u, "undefer {i}", label="undefer an issue that is not deferred")
    c.append(N(u, "defer sw-fix-none", label="defer a missing id"))
    c.append(N(u, "undefer sw-fix-none", label="undefer a missing id"))
    d_ok, d_closed = fresh(), fresh()
    c.append(N(u, f"defer {d_ok} {d_closed}", seed=[create(d_ok), create(d_closed)], before=[["update", d_closed, "--status", "pinned"]], label="defer two ids"))

    one(u, "set-state {i} phase=x --reason y")
    one(u, "promote {i}", "--ephemeral")
    # issue/claim
    one("issue/claim", "update {i} --claim")
    one("issue/claim", "update {i} --claim --title t")
    one("issue/claim", "update {i} --claim", before=[["update", "{i}", "--claim"]], label="claim already held")
    one("issue/claim", "update {i} --claim", before=[["update", "{i}", "--claim", "--actor", "bob"]], label="claimed by another")
    one("issue/claim", "update {i} --claim", before=[["close", "{i}"]], label="claim a closed issue")
    c.append(N("issue/claim", "update sw-fix-none --claim"))
    # issue/unclaim
    claimed = [["update", "{i}", "--claim"]]
    bobs = [["update", "{i}", "--claim", "--actor", "bob"]]
    one("issue/unclaim", "unclaim {i}", before=claimed)
    one("issue/unclaim", "unclaim {i} --reason r", before=claimed)
    one("issue/unclaim", "unclaim {i}", before=bobs, label="unclaim another's")
    one("issue/unclaim", "unclaim {i} --force", before=bobs)
    one("issue/unclaim", "unclaim {i} --if-assignee bob", before=bobs)
    one("issue/unclaim", "unclaim {i} --if-assignee alice", before=bobs)
    one("issue/unclaim", "unclaim {i}", label="not claimed")
    c.append(N("issue/unclaim", "unclaim sw-fix-none"))
    # issue/heartbeat
    # A renewal inside the claim's own second moves nothing a read can see.
    one("issue/heartbeat", "heartbeat {i}", before=claimed, wait=1.2)
    one("issue/heartbeat", "heartbeat {i}", before=bobs, label="another's")
    one("issue/heartbeat", "heartbeat {i}", label="not held")
    c.append(N("issue/heartbeat", "heartbeat sw-fix-none"))
    # issue/close, issue/close-many
    one("issue/close", "close {i}")
    one("issue/close", "close {i} -r done")
    one("issue/close", "close {i} --reason-file r.txt", files={"r.txt": "done\n"}, label="a reason file ending in a newline")
    one("issue/close", "close {i} --reason-file r.txt", files={"r.txt": "done"}, label="a reason file with no newline")
    one("issue/close", "close {i} --reason-file r.txt", files={"r.txt": "x" * 300}, label="a reason longer than 255 characters")
    one("issue/close", "close {i} -r 'two\nlines'", label="a reason with a newline")
    one("issue/close", "close {i} --session s")
    one("issue/close", "close {i}", before=[["close", "{i}"]], label="re-close")
    one("issue/close", "close {i} -f", "--status", "pinned", label="force a pinned issue")
    one("issue/close", "close {i}", "--status", "pinned", label="close a pinned issue")
    one("issue/close", "close {i}", before=bobs, label="close another's claim")
    one("issue/close", "close {i} --if-revision 1")
    one("issue/close", "close {i} --if-revision $rev:{i}")
    one("issue/close", "close {i} --no-auto")
    one("issue/close", "close {i} --suggest-next")
    one("issue/close", "close {i} --continue")
    one("issue/close", "close {i} --claim-next")
    b1, b2 = fresh(), fresh()
    c.append(N("issue/close", f"close {b1}", seed=[create(b2), create(b1, "--deps", f"blocks:{b2}")], label="close a blocked issue"))
    b3, b4 = fresh(), fresh()
    c.append(N("issue/close", f"close {b3} --force", seed=[create(b4), create(b3, "--deps", f"blocks:{b4}")], label="force a blocked issue"))
    c.append(N("issue/close", "close sw-fix-none"))
    m1, m2 = fresh(), fresh()
    c.append(N("issue/close-many", f"close {m1} {m2} -r done", seed=[create(m1), create(m2)]))
    m3 = fresh()
    c.append(N("issue/close-many", f"close {m3} sw-fix-none", seed=[create(m3)], label="one good, one missing"))
    # issue/advance
    a0, a1, a2 = fresh(), fresh(), fresh()
    molecule = [create(a0, "-t", "epic", "--mol-type", "swarm"), create(a1, "--parent", a0), create(a2, "--parent", a0)]
    c.append(N("issue/advance", f"close {a1} --continue", seed=molecule, label="close and continue the molecule"))
    a3, a4, a5 = fresh(), fresh(), fresh()
    c.append(N("issue/advance", f"close {a4} --continue --no-auto", seed=[create(a3, "-t", "epic", "--mol-type", "swarm"), create(a4, "--parent", a3), create(a5, "--parent", a3)]))
    # issue/reopen
    closed = [["close", "{i}"]]
    one("issue/reopen", "reopen {i}", before=closed)
    one("issue/reopen", "reopen {i} -r again", before=closed)
    one("issue/reopen", "reopen {i}", label="reopen an open issue")
    one("issue/reopen", "reopen {i} --if-revision 1", before=closed)
    one("issue/reopen", "reopen {i} --if-revision $rev:{i}", before=closed)
    c.append(N("issue/reopen", "reopen sw-fix-none"))
    # issue/metadata/compare-and-set
    cas = "issue/metadata/compare-and-set"
    one(cas, "update {i} --set-metadata a=1", label="no native command takes expected/value; the nearest write")
    # issue/delete
    one("issue/delete", "delete {i} --dry-run")
    one("issue/delete", "delete {i} --force")
    one("issue/delete", "delete {i}", label="no flags: preview")
    one("issue/delete", "delete {i} --force --if-revision 1")
    one("issue/delete", "delete {i} --force --if-revision $rev:{i}")
    one("issue/delete", "delete {i} --dry-run --if-revision 1")
    one("issue/delete", "delete {i} --from-file ids.txt --force", files={"ids.txt": "sw-fix-w-none\n"}, label="--from-file with a missing id")
    d0, d1 = fresh(), fresh()
    c.append(N("issue/delete", f"delete {d0} --dry-run", seed=[create(d0), create(d1, "--deps", f"blocks:{d0}")], label="preview with dependents"))
    d4, d5 = fresh(), fresh()
    c.append(N("issue/delete", f"delete {d4} --dry-run", seed=[create(d4), create(d5)], before=[["dep", "add", d5, d4]], label="preview of an issue others depend on"))
    d2, d3 = fresh(), fresh()
    c.append(N("issue/delete", f"delete {d2} --force", seed=[create(d2), create(d3, "--deps", f"blocks:{d2}")], label="force with dependents"))
    d4, d5 = fresh(), fresh()
    c.append(N("issue/delete", f"delete {d4} --cascade --force", seed=[create(d4, "-t", "epic"), create(d5, "--parent", d4)], label="cascade"))
    d6, d7 = fresh(), fresh()
    c.append(N("issue/delete", f"delete {d6} {d7} --force", seed=[create(d6), create(d7)]))
    c.append(N("issue/delete", "delete sw-fix-none --force"))
    d8 = fresh()
    c.append(N("issue/delete", f"delete --from-file ids.txt --force", seed=[create(d8)], files={"ids.txt": d8 + "\n"}))
    # issue/create-many and issue/batch
    c.append(N("issue/create-many", "import --input many.jsonl", files={"many.jsonl": '{"id":"sw-fix-imp1","title":"imported one","issue_type":"task","status":"open","priority":2}\n{"id":"sw-fix-imp2","title":"imported two","issue_type":"task","status":"open","priority":2}\n'}))
    c.append(N("issue/create-many", "import --input titles.jsonl", files={"titles.jsonl": '{"title":"imported by title","issue_type":"task","priority":2}\n'}, label="import rows naming no id"))
    c.append(N("issue/create-many", "import --input stamped.jsonl", files={"stamped.jsonl": '{"id":"sw-fix-imp4","title":"stamped","issue_type":"task","status":"open","priority":2,"created_at":"2020-01-01T00:00:00Z","updated_at":"2020-01-01T00:00:00Z"}\n'}, label="import a row with stored timestamps"))
    c.append(N("issue/create-many", "import --input many.jsonl --dry-run", files={"many.jsonl": '{"id":"sw-fix-imp3","title":"imported three","issue_type":"task","status":"open","priority":2}\n'}))
    c.append(N("issue/create-many", "create --file many.md --dry-run", files={"many.md": "## dry md\n"}))
    e1, e2, e3 = fresh(), fresh(), fresh()
    plan = f'create task 2 "batch created"\nupdate {e1} title=changed\nclose {e2} because\ndep add {e3} {e1} blocks\ndep remove {e3} {e1}\n'
    c.append(N("issue/batch", "batch -f plan.txt", seed=[create(e1), create(e2), create(e3)], files={"plan.txt": plan}))
    c.append(N("issue/batch", "batch --dry-run -f plan.txt", files={"plan.txt": 'create task 2 "batch dry"\n'}))
    c.append(N("issue/batch", "batch -m message", stdin='create task 2 "batch stdin"\n'))
    c.append(N("issue/batch", "batch", stdin="close sw-fix-none\n", label="batch with a miss"))
    e4, e5 = fresh(), fresh()
    plan2 = f'create task 2 "batch created two"\nupdate {e4} title=changed priority=1\nclose {e5} because\n'
    c.append(N("issue/batch", "batch -f plan.txt", seed=[create(e4), create(e5)], files={"plan.txt": plan2}, label="batch with no dep remove"))
    e6 = fresh()
    c.append(N("issue/batch", "batch -f plan.txt", seed=[create(e6)], files={"plan.txt": f"update {e6} title=must-not-land\ndep remove {e6} sw-fix-none\n"}, label="batch refusing a dep remove up front"))
    # issue/dep/add
    def pair(command, *flags, **kw):
        x, y = fresh(), fresh()
        c.append(N("issue/dep/add", command.replace("{x}", x).replace("{y}", y), seed=[create(x, *flags), create(y)], **kw))
    pair("dep add {x} {y}")
    pair("dep add {x} {y} -t related")
    pair("dep add {x} --blocked-by {y}")
    pair("dep add {x} --depends-on {y}")
    pair("dep add {x} {y} --no-cycle-check")
    pair("dep {x} --blocks {y}")
    pair("dep add {x} {y} -t discovered-from")
    pair("dep add {x} sw-fix-none")
    pair("dep add sw-fix-none {x}")
    pair("dep relate {x} {y}")
    pair("link {x} {y}")
    pair("link {x} {y} --type related")
    pair("duplicate {x} --of {y}")
    pair("supersede {x} --with {y}")
    x1, y1 = fresh(), fresh()
    c.append(N("issue/dep/add", f"dep add {x1} {y1}", seed=[create(y1), create(x1, "--deps", f"blocks:{y1}")], label="same edge again"))
    x2, y2 = fresh(), fresh()
    c.append(N("issue/dep/add", f"dep add {y2} {x2}", seed=[create(y2), create(x2, "--deps", f"blocks:{y2}")], label="a cycle"))
    x3, y3 = fresh(), fresh()
    c.append(N("issue/dep/add", f"dep add {x3} {y3} -t related", seed=[create(y3), create(x3, "--deps", f"blocks:{y3}")], label="other type on a held edge"))
    x4, y4 = fresh(), fresh()
    c.append(N("issue/dep/add", "dep add --file edges.jsonl", seed=[create(x4), create(y4)], files={"edges.jsonl": f'{{"issue_id":"{x4}","depends_on_id":"{y4}","type":"blocks"}}\n'}))
    # issue/dep/remove
    x5, y5 = fresh(), fresh()
    c.append(N("issue/dep/remove", f"dep remove {x5} {y5}", seed=[create(y5), create(x5, "--deps", f"blocks:{y5}")]))
    x6, y6 = fresh(), fresh()
    c.append(N("issue/dep/remove", f"dep remove {x6} {y6}", seed=[create(y6), create(x6)], label="no such edge"))
    x7, y7 = fresh(), fresh()
    c.append(N("issue/dep/remove", f"dep unrelate {x7} {y7}", seed=[create(x7), create(y7)], before=[["dep", "relate", x7, y7]]))
    c.append(N("issue/dep/remove", "dep remove sw-fix-none sw-fix-none2"))
    # issue/comment
    one("issue/comment", "comment {i} hello")
    one("issue/comment", "comment {i} --file c.txt", files={"c.txt": "from file\n"})
    one("issue/comment", "comment {i} --stdin", stdin="from stdin\n")
    one("issue/comment", "comments add {i} hello")
    one("issue/comment", "comments add {i} hello --author kim", label="--author names someone else")
    one("issue/comment", "comments add {i} hello --actor kim --author kim", label="--actor and --author, one name")
    one("issue/comment", "comments add {i} hello --actor sam --author kim", label="--actor and --author, two names")
    one("issue/comment", "comments add {i} -a alice hello")
    one("issue/comment", "comments add {i} -f c.txt", files={"c.txt": "from file\n"})
    c.append(N("issue/comment", "comment sw-fix-none hello"))
    # config, kv
    c.append(N("config/set", "config set fix.w1 value"))
    c.append(N("config/set", "config set fix.w1 value", label="same value again"))
    c.append(N("config/set", "config set issue_prefix zz", label="a protected key"))
    c.append(N("config/unset", "config unset fix.w1"))
    c.append(N("config/unset", "config unset fix.w1", label="already gone"))
    # memory
    c.append(N("memory/remember", "remember keyed --key fix-w1"))
    c.append(N("memory/remember", "remember replaced --key fix-w1", label="replace a memory"))
    c.append(N("memory/remember", "remember replaced --key fix-w1 --force", label="replace with force"))
    c.append(N("memory/remember", "remember 'the key is derived from this'"))
    c.append(N("memory/remember", "remember ''", label="empty content"))
    c.append(N("memory/forget", "forget fix-w1"))
    c.append(N("memory/forget", "forget fix-w1", label="already gone"))
    # Writes with no Callable of their own, kept for the separate list.
    c.append(N("(none)", "kv set fix-w1 value"))
    c.append(N("(none)", "kv clear fix-w1"))
    one("(none)", "rename {i} sw-fix-renamed-{i}")
    one("(none)", "mol burn {i}")
    one("(none)", "mol squash {i}")
    one("(none)", "restore {i}")
    c.append(N("(none)", "recompute-blocked"))
    c.append(N("(none)", "dolt commit -m x"))
    return c


def as_text(case):
    """The same command with no --json: the text path often reads through different calls."""
    if case["plain"]:
        return None
    clone = dict(case)
    clone["plain"] = True
    clone["mode"] = "text"
    return clone


def global_cases():
    """Each runs on its own pair of fresh workspaces, seeded the same way."""
    closed_e = [create("sw-fix-p1", "--ephemeral"), create("sw-fix-p2", "--ephemeral")]
    close_e = [["close", "sw-fix-p1"], ["close", "sw-fix-p2"]]
    closed_d = [create("sw-fix-q1"), create("sw-fix-q2")]
    close_d = [["close", "sw-fix-q1"], ["close", "sw-fix-q2"]]
    held = [create("sw-fix-e1", "-l", "el")]
    hold = [["update", "sw-fix-e1", "--claim", "--actor", "bob"]]
    ready = [create("sw-fix-g1", "-l", "next", "-p", "1"), create("sw-fix-g2", "-p", "2", "-t", "bug")]
    c = []

    def g(callables, command, seed, before, **kw):
        c.append(N(callables, command, seed=seed, before=before, phase="global", **kw))

    for command in ("purge --dry-run", "purge --force", "purge", "purge --older-than 1s --dry-run", "purge --older-than 1s --force",
                    "purge --older-than 3650d --force", "purge --pattern sw-fix-p1 --force", "purge --pattern 'sw-fix-p*' --dry-run",
                    "purge --limit 1 --force", "purge --retain-ids --force", "purge --wisps-plane --pattern '*' --dry-run",
                    "purge --wisps-plane --pattern '*' --force"):
        g("issue/purge", command, closed_e, close_e, wait=3)
    for command in ("prune --pattern '*' --dry-run", "prune --older-than 1s --dry-run", "prune --older-than 1s --force", "prune --older-than 3650d --force",
                    "prune --pattern sw-fix-q1 --dry-run", "prune --pattern sw-fix-q1 --force", "prune --pattern 'sw-fix-q*' --force",
                    "prune --older-than 1s --ignore-references --force", "prune --older-than 1s --retain-ids --force", "prune --older-than 1s", "prune"):
        g("issue/prune", command, closed_d, close_d, wait=3)
    for command in ("reclaim", "reclaim --older-than 0s", "reclaim --older-than 24h", "reclaim --older-than 0s --id sw-fix-e1",
                    "reclaim --older-than 0s --assignee bob", "reclaim --older-than 0s --assignee nobody", "reclaim --older-than 0s -l el",
                    "reclaim --older-than 0s --label-any el,x", "reclaim --older-than 0s --exclude-label el", "reclaim --older-than 0s --any-replica"):
        g("issue/reclaim", command, held, hold)
    for command in ("ready --claim", "ready --claim -p 1", "ready --claim -t bug", "ready --claim -l next", "ready --claim -l nothing",
                    "ready --claim -u", "ready --claim -s priority", "ready --claim --label-any next,x",
                    "ready --claim --exclude-label next", "ready --claim --exclude-type bug", "ready --claim --include-deferred"):
        g("issue/claim-next", command, ready, [])
    g("(none)", "rename-prefix zz --dry-run", ready, [])
    g("(none)", "rename-prefix zz", ready, [])
    g("(none)", "gc --dry-run", closed_d, close_d)
    g("(none)", "flatten --dry-run", ready, [])
    g("(none)", "compact --dry-run", ready, [])
    g("(none)", "sync", ready, [])
    g("(none)", "dolt push", ready, [])
    g("(none)", "dolt pull", ready, [])
    return c
