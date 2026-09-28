#!/usr/bin/env python3
"""Installed current Issue-list preview proof using normal CLI authoring.

Normal authoring and fresh processes only. Embedded and caller-owned ordinary
Dolt run sequentially; no SQL, schema fixtures, mocks or hidden bootstrap.
"""
import argparse
import copy
import hashlib
import importlib.util
import json
import os
from pathlib import Path
import signal
import time

HERE = Path(__file__).resolve().parent
SPEC = importlib.util.spec_from_file_location("graph_c0_capture", HERE / "graph-c0-smoke.py")
c0 = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(c0)
SCOPE = "https://example.invalid/disposable-issue-list/"
LIMITATIONS = [
    "explicit Issue records-json and flat preview only; not legacy JSON/tree compatibility or generic collection listing",
    "hasMore reports a limited Issue page, not a BDP continuation or complete History",
    "custom statuses, infrastructure visibility, corruption/authority transaction failures and large acquisition/output bounds require separate real-store tests",
    "order assertions use unique priorities/titles; equal-key backing-ID ties are not claimed to be canonical-ID ordering",
    "visible record equality cannot prove database/event-count equality; no forced overlap or rollback injection here",
    "ordinary Dolt 2.1.8 database provisioning must remain serialized",
]


def semantic_digest(value):
    return hashlib.sha256(json.dumps(value, ensure_ascii=False, sort_keys=True,
                                     separators=(",", ":")).encode()).hexdigest()

def exercise(capture):
    backend = "server" if capture.args.server_port else "embedded"
    saved, listings = [], []

    def raw(label, args):
        receipt, _, _ = capture.run(label, args)
        d = capture.output / capture.records[-1]["artifact"]
        return receipt, (d / "stdout.log").read_bytes(), (d / "stderr.log").read_bytes()

    def command(label, args):
        response = capture.success(label, [*args, "--json"])
        if args[0] == "ready":
            # Ready's established preview result is an array, unlike mutation
            # and single-record results accepted by the shared c0.envelope.
            c0.require(isinstance(response, dict) and set(response) == {"schemaVersion", "preview", "result"} and
                       response["schemaVersion"] == 1 and response["preview"] is True and
                       isinstance(response["result"], list) and all(isinstance(row, dict) for row in response["result"]),
                       label + ": invalid ready array envelope")
            return response["result"]
        return c0.envelope(response)

    def refuse(label, flags, code):
        receipt, out, err = raw(label, ["list", *flags])
        expected_exit = 2 if code == "invalid_selector" else 5
        c0.require(receipt["exit_code"] == expected_exit and out == b"",
                   label + ": wrong refusal exit or partial records")
        structured = "records-json" in flags or "--json" in flags
        if structured:
            problem = json.loads(err)
            c0.require(problem.get("code") == code and problem.get("retryable") is False and "result" not in problem,
                       label + ": incorrect typed refusal")
        else:
            # --format json may make legacy global JSON active before admission;
            # both error encodings must retain the exact typed failure code.
            if err.lstrip().startswith(b"{"):
                c0.require(json.loads(err).get("code") == code, label + ": wrong typed JSON failure")
            else:
                c0.require(err.decode().startswith(code + ":"), label + ": wrong typed text failure: " + err.decode())

    def init_args(graph, prefix):
        args = ["init", "--prefix", prefix, "--non-interactive", "--skip-hooks", "--skip-agents", "--json"]
        if graph:
            args += ["--graph-mode", "link", "--scope-url", SCOPE]
        if capture.args.server_port:
            args += ["--server", "--external", "--server-host", "127.0.0.1", "--server-port",
                     str(capture.args.server_port), "--database", prefix + "_" + capture.root.name.replace("-", "_"),
                     "--server-user", "root"]
        return args

    initialized = c0.envelope(capture.success("normal-graph-init", init_args(True, "ilist")))
    c0.require(initialized.get("scope") == SCOPE and initialized.get("backend") == backend, "wrong authority/backend")
    status = command("list-capabilities", ["status", "--graph"])
    caps = status["capabilities"]
    c0.require(caps.get("issueList") is True and caps.get("issueListLegacyJSON") is False and
               caps.get("issueListTree") is False and caps.get("issueWorkflows") is False and
               caps.get("memory") is False and caps.get("historyExact") is False and
               status["limits"].get("issueListOutputBytes") == 16 << 20, "list capabilities overstate compatibility or profiles")

    def save(label, value):
        c0.require(value.get("id", "").startswith(SCOPE) and value.get("type") and value.get("version") and
                   value.get("revision") and value.get("attribution"), label + ": incomplete authored record")
        saved.append({"label": label, "record": copy.deepcopy(value), "semantic_sha256": semantic_digest(value)})
        c0.write_json(capture.output / "expected-records.json", saved)
        return value

    def current(label, values):
        for i, value in enumerate(values):
            c0.require(command(label + "-" + str(i), ["show", value["id"], "--readonly"]) == value,
                       label + ": current record changed during read-only listing")

    def listing(label, values, flags=None, more=False, ordered=True):
        args = ["list", "--format", "records-json", *(flags or [])]
        receipt, out, err = raw(label, args)
        c0.require(receipt["exit_code"] == 0 and err == b"", label + ": list failed: " + err.decode("utf-8", "replace"))
        c0.require(len(out) <= 16 << 20, label + ": output exceeds disclosed byte bound")
        envelope = json.loads(out)
        c0.require(set(envelope) == {"schemaVersion", "preview", "result"} and
                   envelope["schemaVersion"] == 1 and envelope["preview"] is True, label + ": wrong envelope")
        result = envelope["result"]
        c0.require(set(result) == {"items", "hasMore"} and result["hasMore"] is more and isinstance(result["items"], list),
                   label + ": wrong page completeness/shape")
        wanted = {value["id"]: value for value in values}
        items = result["items"]
        c0.require(len(items) == len(values) and {value["id"] for value in items} == set(wanted) and
                   all(value == wanted[value["id"]] for value in items), label + ": incomplete/foreign/duplicate/mutated Issue records")
        if ordered:
            c0.require(items == values, label + ": wrong Issue ordering")
        listings.append({"label": label, "argv": args, "result": result, "semantic_sha256": semantic_digest(result)})
        c0.write_json(capture.output / "list-results.json", listings)
        return items

    listing("empty-current-issues", [])
    memory = save("memory", command("create-memory", ["remember", "This body is not an Issue", "--id", "beads/memory",
                  "--title", "Alpha memory should not list"]))
    issues = []
    for name, title, priority, kind, labels in [
        ("alpha", "Alpha roadmap", 0, "feature", "red,blue"),
        ("beta", "Beta repair", 1, "bug", "red"),
        ("gamma", "Gamma cleanup", 2, "task", "green"),
        ("delta", "Delta review", 3, "task", ""),
    ]:
        flags = ["create", title, "--id", "beads/" + name, "--priority", str(priority), "--type", kind,
                 "--description", "Preserved description for " + name, "--actor", "list-author"]
        if labels:
            flags += ["--labels", labels]
        issues.append(save(name + "-created", command("create-" + name, flags)))
    a, b, g, d = issues
    added = command("add-blocker", ["dep", "add", b["id"], a["id"], "--actor", "dependency-author"])
    b, dependency = save("beta-with-dependency", added["source"]), save("dependency", added["link"])
    related = command("add-memory-context", ["link", memory["id"], g["id"], "--resource-type",
        SCOPE + "types/preview-related-v2", "--id", "links/context", "--properties", '{"note":"not an Issue"}',
        "--if-source-revision", memory["revision"]])
    memory, context = save("memory-owned", related["source"]), save("context", related["link"])
    all_open = [a, b, g, d]
    listing("mixed-resources-issues-only", all_open)
    listing("stable-repeat", all_open)
    listing("records-json-quiet", all_open, ["--quiet"])
    listing("records-json-over-flat", all_open, ["--flat"])
    ready = command("ready-is-not-list", ["ready", "--readonly"])
    c0.require({row["id"] for row in ready} == {a["id"], g["id"], d["id"]}, "blocked open Issue was treated as ready")
    listing("blocked-open-is-still-listed", [b], ["--title", "Beta"])
    current("new-process-complete-records", all_open + [memory, dependency, context])
    capture.passed("normal init and mixed authoring produce complete canonical Issue records; blocked open Issue appears in list but not ready")

    for label, values, flags in [
        ("status-open", all_open, ["--status", "open"]),
        ("state-alias", all_open, ["--state", "open"]),
        ("status-comma-union", all_open, ["--status", "open,closed"]),
        ("title-substring", [a], ["--title", "roadmap"]),
        ("title-contains", [b], ["--title-contains", "REPAIR"]),
        ("type", [b], ["--type", "bug"]),
        ("exact-priority", [g], ["--priority", "P2"]),
        ("priority-range", [b, g], ["--priority-min", "1", "--priority-max", "P2"]),
        ("labels-and", [a], ["--label", "red,blue"]),
        ("labels-any", [a, b, g], ["--label-any", "red,green"]),
        ("exclude-label", [g, d], ["--exclude-label", "red"]),
        ("intersect-filters", [b], ["--label", "red", "--type", "bug", "--priority-max", "1"]),
        ("pinned-empty", [], ["--pinned"]),
        ("not-pinned", all_open, ["--no-pinned"]),
        ("no-match", [], ["--title", "nothing-matches-this-title"]),
        ("limited-page", [a, b], ["--limit", "2"]),
        ("exact-size-page", all_open, ["--limit", "4"]),
        ("explicit-unlimited", all_open, ["--limit", "0"]),
        ("title-order", [a, b, d, g], ["--sort", "title"]),
        ("title-reverse", [g, d, b, a], ["--sort", "title", "--reverse"]),
        ("priority-reverse", [d, g, b, a], ["--sort", "priority", "--reverse"]),
    ]:
        listing(label, values, flags, more=label == "limited-page")
    # Stable membership and monotonic keys avoid making a new equal-key ID
    # ordering contract for existing Issue sort semantics.
    for field, key, descending in [("created", "created_at", True), ("updated", "updated_at", True),
                                   ("type", "issue_type", False), ("status", "status", False)]:
        items = listing("sort-" + field, all_open, ["--sort", field], ordered=False)
        keys = [row["properties"][key] for row in items]
        c0.require(keys == sorted(keys, reverse=descending), "sort-" + field + ": keys out of order")
        listing("repeat-sort-" + field, items, ["--sort", field])
    capture.passed("supported status/title/type/priority/label filters intersect; sorting and exact hasMore distinguish bounded and unlimited pages")

    a_open = a
    a = save("alpha-closed", command("close-alpha", ["close", a["id"], "--reason", "Review complete"])["issue"])
    listing("closed-default-excluded", [b, g, d])
    listing("all-includes-closed", [a, b, g, d], ["--all"])
    listing("explicit-closed", [a], ["--status", "closed"])
    listing("status-all", [a, b, g, d], ["--status", "all"])
    listing("state-comma", [a, b, g, d], ["--state", "open,closed"])
    listing("all-explicit-limit", [a], ["--all", "--limit", "1"], more=True)
    ready = command("ready-after-close", ["ready", "--readonly"])
    c0.require({row["id"] for row in ready} == {b["id"], g["id"], d["id"]}, "close did not unblock dependent")

    yaml = capture.work / ".beads" / "config.yaml"
    prior_yaml = yaml.read_bytes() if yaml.exists() else None
    try:
        yaml.write_text("list:\n  limit: 2\n")
        listing("configured-limit", [b, g], more=True)
        listing("all-overrides-configured-limit", [a, b, g, d], ["--all"])
        listing("explicit-limit-overrides-all-config", [a], ["--all", "--limit", "1"], more=True)
        listing("explicit-zero-overrides-config", [b, g, d], ["--limit", "0"])
        yaml.write_text("json: true\n")
        listing("explicit-records-over-ambient-json", [b, g, d])
    finally:
        if prior_yaml is None:
            yaml.unlink(missing_ok=True)
        else:
            yaml.write_bytes(prior_yaml)
    a = save("alpha-reopened", command("reopen-alpha", ["reopen", a["id"], "--reason", "Needs follow-up"])["issue"])
    all_open = [a, b, g, d]
    listing("reopened-default-included", all_open)
    c0.require(command("old-open-exact", ["show", a_open["id"], "--version", a_open["version"], "--readonly"]) == a_open,
               "list lifecycle changed old retained Issue")
    capture.passed("close/reopen changes Issue default visibility and readiness; --all retains legacy visibility and limit precedence")

    def human(label, values, flags=None, more=False, quiet=False):
        args = ["list", "--flat", *(flags or [])]
        if quiet:
            args += ["--quiet"]
        receipt, out, err = raw(label, args)
        c0.require(receipt["exit_code"] == 0 and err == b"", label + ": human list failed")
        expected = "Issues (%d; more: %s; graph preview)\n" % (len(values), str(more).lower())
        for row in values:
            p = row["properties"]
            expected += '%s %s P%d %s\n' % (json.dumps(row["id"]), json.dumps(p["status"]), p["priority"], json.dumps(p["title"]))
        if more:
            expected += "More matching Issues exist; increase --limit or use --all within preview bounds.\n"
        c0.require(out == (b"" if quiet else expected.encode()), label + ": human list differs from complete promised output")
    human("human-flat", all_open)
    human("human-truncated", [a], ["--limit", "1"], more=True)
    human("human-empty", [], ["--title", "nothing-matches-this-title"])
    human("human-quiet", all_open, quiet=True)
    capture.passed("explicit flat output names canonical IDs and hasMore; empty and quiet forms are exact")

    for label, flags in [
        ("bare-tree", []), ("legacy-json", ["--json"]), ("legacy-format-json", ["--format", "json"]),
        ("false-json", ["--format", "records-json", "--json=false"]),
        ("false-tree", ["--format", "records-json", "--tree=false"]),
        ("false-flat", ["--flat=false"]),
        ("repeated-status", ["--format", "records-json", "--status", "open", "--status", "closed"]),
        ("repeated-state", ["--format", "records-json", "--state", "open", "--state", "closed"]),
        ("repeated-type", ["--format", "records-json", "--type", "bug", "--type", "task"]),
        ("false-ready", ["--format", "records-json", "--ready=false"]),
        ("unsupported-id", ["--format", "records-json", "--sort", "id"]),
        ("unsupported-offset", ["--format", "records-json", "--offset", "0"]),
        ("unsupported-assignee-sort", ["--format", "records-json", "--sort", "assignee"]),
        ("unsupported-parent", ["--format", "records-json", "--parent", a["id"]]),
    ]:
        refuse(label, flags, "capability_unavailable")
    for label, flags in [
        ("negative-limit", ["--limit", "-1"]), ("invalid-priority", ["--priority", "P99"]),
        ("unknown-status", ["--status", "unknown-status"]), ("unknown-type", ["--type", "unknown-type"]),
        ("empty-label", ["--label", ""]), ("conflicting-pinned", ["--pinned", "--no-pinned"]),
        ("status-and-state", ["--status", "open", "--state", "closed"]),
    ]:
        refuse(label, ["--format", "records-json", *flags], "invalid_selector")
    current("after-refusals", all_open + [memory, dependency, context])
    capture.passed("unsupported output modes and even false unsupported flags refuse explicitly with no partial stdout")

    prior_readonly = capture.env.get("BD_READONLY")
    mayor, freeze = capture.work / "mayor", capture.work / "MIGRATION-FREEZE"
    c0.require(not mayor.exists() and not freeze.exists(), "unexpected policy files")
    for policy in ["flag", "environment", "configuration", "freeze"]:
        try:
            flags = ["--readonly"] if policy == "flag" else []
            if policy == "environment":
                capture.env["BD_READONLY"] = "true"
            elif policy == "configuration":
                yaml.write_text("readonly: true\n")
            elif policy == "freeze":
                mayor.mkdir()
                (mayor / "town.json").write_text("{}\n")
                freeze.write_text("issue-list\t2026-09-27T00:00:00Z\tread policy\n")
            listing("read-policy-" + policy, all_open, flags)
        finally:
            if prior_readonly is None:
                capture.env.pop("BD_READONLY", None)
            else:
                capture.env["BD_READONLY"] = prior_readonly
            if prior_yaml is None:
                yaml.unlink(missing_ok=True)
            else:
                yaml.write_bytes(prior_yaml)
            if policy == "freeze":
                freeze.unlink(missing_ok=True)
                (mayor / "town.json").unlink(missing_ok=True)
                mayor.rmdir()
    metadata = capture.work / ".beads" / "metadata.json"
    original_metadata = metadata.read_bytes()
    try:
        invalid = json.loads(original_metadata)
        invalid["graph_workspace"] = str(capture.root)
        metadata.write_text(json.dumps(invalid))
        refuse("authority-mismatch", ["--format", "records-json"], "not_authority")
    finally:
        metadata.write_bytes(original_metadata)
    current("final-current-records", all_open + [memory, dependency, context])
    for i, item in enumerate(saved):
        row = item["record"]
        c0.require(command("final-exact-" + str(i), ["show", row["id"], "--version", row["version"], "--readonly"]) == row,
                   "retained Issue/Memory/Link changed during listing")
    capture.passed("read-only and freeze permit reads, mismatched authority refuses, and all current/retained records remain unchanged")

    graph_work = capture.work
    capture.work = capture.root / "legacy-workspace"
    capture.work.mkdir(mode=0o700)
    try:
        receipt, _, err = capture.run("normal-legacy-init", init_args(False, "lilist"))
        c0.require(receipt["exit_code"] == 0, "legacy init failed: " + err)
        legacy = capture.success("legacy-create", ["create", "Legacy list remains available", "--json"])
        for label, flags in [("legacy-json-list", ["--json"]), ("legacy-format-alias", ["--format", "json"])]:
            result = capture.success(label, ["list", *flags])
            c0.require(isinstance(result, list) and len(result) == 1 and result[0]["id"] == legacy["id"] and
                       result[0]["title"] == legacy["title"], label + ": legacy array projection changed")
        receipt, out, err = raw("legacy-flat-list", ["list", "--flat"])
        c0.require(receipt["exit_code"] == 0 and legacy["id"].encode() in out and legacy["title"].encode() in out,
                   "legacy flat listing no longer works")
    finally:
        capture.work = graph_work
    capture.passed("ordinary legacy JSON array, format alias and flat Issue listing remain available")
    return {"saved_versions": len(saved), "list_calls": len(listings), "http_exercised": False,
            "expected_records_sha256": c0.sha256(capture.output / "expected-records.json"),
            "list_results_sha256": c0.sha256(capture.output / "list-results.json")}


def run_backend(args, backend, deadline):
    selected = argparse.Namespace(**vars(args))
    selected.output_dir = args.output_dir / backend
    selected.server_port = args.server_port if backend == "server" else None
    selected.server_root = args.server_root if backend == "server" else None
    selected.total_timeout = max(0.01, deadline - time.monotonic())
    capture = c0.Capture(selected)
    began = time.monotonic()
    failure, interrupted, result = None, False, {}
    previous_signals = {}

    def interrupted_run(signum, _frame):
        capture.stop()
        raise KeyboardInterrupt(signum)

    try:
        for signum in [signal.SIGINT, signal.SIGTERM]:
            previous_signals[signum] = signal.signal(signum, interrupted_run)
        result = exercise(capture)
        c0.require(c0.sha256(args.bd) == capture.binary_hash, "installed binary changed during capture")
    except BaseException as exc:
        interrupted = isinstance(exc, KeyboardInterrupt)
        failure = f"{type(exc).__name__}: {exc}"
    finally:
        capture.stop()
        for signum, handler in previous_signals.items():
            signal.signal(signum, handler)
        if capture.active:
            failure = failure or "owned child processes remain after capture"
        summary = {"passed": failure is None, "backend": backend, "failure": failure,
                   "interrupted": interrupted, "elapsed_seconds": time.monotonic() - began,
                   "checks": capture.passes, "commands": capture.records,
                   "active_child_count": len(capture.active), "private_root": str(capture.root),
                   "installed_binary_sha256": capture.binary_hash,
                   "harness_sha256": c0.sha256(Path(__file__)),
                   "capture_helper_sha256": c0.sha256(HERE / "graph-c0-smoke.py"),
                   "limitations": LIMITATIONS, **result}
        c0.write_json(capture.output / "summary.json", summary)
    print(f"{'PASS' if summary['passed'] else 'FAIL'} {backend}: {capture.output}", flush=True)
    return summary


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--bd", type=Path, required=True, help="absolute installed bd executable")
    parser.add_argument("--output-dir", type=Path, required=True, help="new capture directory")
    parser.add_argument("--backend", choices=["embedded", "server", "both"], default="both")
    parser.add_argument("--server-port", type=int, help="caller-owned ordinary loopback Dolt server")
    parser.add_argument("--server-root", type=Path, help="disposable caller-owned server root, provenance only")
    parser.add_argument("--command-timeout", type=float, default=120)
    parser.add_argument("--total-timeout", type=float, default=900, help="whole sequential capture budget")
    args = parser.parse_args()
    c0.require(args.bd.is_absolute() and args.bd.is_file() and os.access(args.bd, os.X_OK),
               "--bd must name an absolute executable installed binary")
    c0.require(args.output_dir.is_absolute(), "--output-dir must be absolute")
    c0.require(args.command_timeout > 0 and args.total_timeout > 0, "timeouts must be positive")
    if args.backend in {"server", "both"}:
        c0.require(args.server_port is not None and 0 < args.server_port < 65536,
                   "server/both require --server-port for an ordinary caller-owned server")
    if args.server_root is not None:
        c0.require(args.server_root.is_absolute() and args.server_root.is_dir(),
                   "--server-root must name an existing absolute disposable server directory")
    args.output_dir.mkdir(mode=0o700, parents=True, exist_ok=False)
    deadline = time.monotonic() + args.total_timeout
    modes = ["embedded", "server"] if args.backend == "both" else [args.backend]
    results = []
    failure = None
    try:
        for backend in modes:
            c0.require(time.monotonic() < deadline, "total capture deadline expired before next backend")
            result = run_backend(args, backend, deadline)
            results.append(result)
            if result["interrupted"]:
                break
    except BaseException as exc:
        failure = f"{type(exc).__name__}: {exc}"
    passed = failure is None and len(results) == len(modes) and all(item["passed"] for item in results)
    c0.write_json(args.output_dir / "summary.json", {
        "passed": passed, "failure": failure, "requested_backends": modes,
        "backends": [{key: value for key, value in result.items() if key != "commands"} for result in results],
        "harness_sha256": c0.sha256(Path(__file__)), "installed_binary_sha256": c0.sha256(args.bd),
        "active_child_count": sum(result["active_child_count"] for result in results),
        "command_count": sum(len(result["commands"]) for result in results),
        "check_count": sum(len(result["checks"]) for result in results),
        "limitations": LIMITATIONS,
    })
    return 0 if passed else 1


if __name__ == "__main__":
    raise SystemExit(main())
