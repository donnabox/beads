#!/usr/bin/env python3
"""Compact installed graph Issue due-date proof, with normal CLI-only authoring.

Run embedded and caller-owned ordinary Dolt sequentially. No SQL setup, mocks,
hidden bootstrap, internal count claims or concurrent-process qualification.
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
SCOPE = "https://example.invalid/disposable-issue-due/"
LIMITATIONS = [
    "due writes use the existing nullable datetime: nearest second with half-second ties forward; filters use existing UTC whole-second truncation",
    "CLI parsing and current/exact saved records do not prove independent API offset handling or internal recorder/event counts; separate storage tests own those gates",
    "no public ordered History, HTTP Write, scheduler, automatic wake-up, concurrency or uncertain-COMMIT claim",
    "overdue uses the native current clock and excludes closed Issues; fixtures bound that clock well away from due instants",
    "ordinary Dolt 2.1.8 provisioning is sequential through normal init, with owned child cleanup and no SQL fixtures",
]


def semantic_digest(value):
    return hashlib.sha256(json.dumps(value, ensure_ascii=False, sort_keys=True,
                                     separators=(",", ":")).encode()).hexdigest()


def exercise(capture):
    backend = "server" if capture.args.server_port else "embedded"
    saved, lists = [], []
    c0.require(2001 <= time.gmtime().tm_year < 2099, "overdue fixture requires clock safely between past2000 and future2100")

    def command(label, argv):
        return c0.envelope(capture.success(label, [*argv, "--json"]))

    def save(label, record):
        c0.require(record.get("id", "").startswith(SCOPE) and record.get("version") == record.get("revision") and
                   record.get("version") and record.get("type") and record.get("attribution"), label + ": incomplete record")
        saved.append({"label": label, "record": copy.deepcopy(record), "semantic_sha256": semantic_digest(record)})
        c0.write_json(capture.output / "expected-records.json", saved)
        return record

    def current(label, record):
        c0.require(command(label, ["show", record["id"], "--readonly"]) == record, label + ": complete current record changed")

    def exact(label, record):
        c0.require(command(label, ["show", record["id"], "--version", record["version"], "--readonly"]) == record,
                   label + ": complete saved record changed")

    def refuse(label, argv, code, exit_code):
        receipt, _, _ = capture.run(label, [*argv, "--json"])
        folder = capture.output / capture.records[-1]["artifact"]
        c0.require(receipt["exit_code"] == exit_code and (folder / "stdout.log").read_bytes() == b"", label + ": wrong refusal status/output")
        error = json.loads((folder / "stderr.log").read_bytes())
        c0.require(error.get("code") == code and error.get("retryable") is False and "result" not in error, label + ": wrong structured refusal")

    def accepted(label, before, result, fields, actor):
        c0.require(set(result) == {"issue", "changed"} and result["changed"] is True, label + ": missing changed mutation")
        after = result["issue"]
        expected = copy.deepcopy(before)
        for key, value in fields.items():
            if value is None:
                expected["properties"].pop(key, None)
            else:
                expected["properties"][key] = value
        expected["properties"]["updated_at"] = after["properties"]["updated_at"]
        for key in ("version", "revision", "attribution"):
            expected[key] = after[key]
        c0.require(after == expected and after["version"] == after["revision"] != before["version"] and
                   after["attribution"]["actor"] == actor, label + ": wrong due or unexpected sibling/owned/author change")
        return save(label, after)

    def edit(label, before, flags, fields, *, unconditional=False, changed=True):
        argv = ["update", before["id"], *flags, "--actor", "due-author"]
        argv += ["--unconditional"] if unconditional else ["--if-revision", before["revision"]]
        result = command(label, argv)
        if changed:
            after = accepted(label, before, result, fields, "due-author")
        else:
            c0.require(result == {"issue": before, "changed": False}, label + ": complete no-op record changed")
            after = before
        current(label + "-fresh", after)
        return after

    def listing(label, records, flags=(), *, all_states=True, limit=0):
        argv = ["list", "--format", "records-json", "--readonly", "--sort", "title", "--limit", str(limit)]
        if all_states:
            argv.append("--all")
        argv += list(flags)
        result = c0.envelope(capture.success(label, argv))
        ordered = sorted(records, key=lambda row: row["properties"]["title"])
        expected = ordered[:limit] if limit else ordered
        c0.require(result["items"] == expected and result["hasMore"] is bool(limit and len(ordered) > limit),
                   label + ": full filtered records/order/hasMore differ")
        lists.append({"label": label, "argv": argv, "result": result, "semantic_sha256": semantic_digest(result)})
        c0.write_json(capture.output / "list-results.json", lists)

    init = ["init", "--prefix", "idue", "--non-interactive", "--skip-hooks", "--skip-agents",
            "--graph-mode", "link", "--scope-url", SCOPE]
    if capture.args.server_port:
        init += ["--server", "--external", "--server-host", "127.0.0.1", "--server-port", str(capture.args.server_port),
                 "--database", "idue_" + capture.root.name.replace("-", "_"), "--server-user", "root"]
    initialized = command("normal-init", init)
    c0.require(initialized.get("scope") == SCOPE and initialized.get("backend") == backend, "wrong initialized authority")
    caps = command("due-capabilities", ["status", "--graph"])["capabilities"]
    c0.require(caps.get("issueDueDate") is True and caps.get("issueDueFilter") is True and
               caps.get("issueWorkflows") is False and caps.get("historyExact") is False, "due capability overstates scope")
    original = save("source-initial", command("create-fractional-offset", ["create", "A work", "--id", "beads/work",
                     "--due", "2000-01-02T03:04:05.5-07:00", "--notes", "Keep original notes", "--actor", "initial-author"]))
    c0.require(original["properties"].get("due_at") == "2000-01-02T10:04:06Z" and
               original["properties"].get("notes") == "Keep original notes", "fractional offset did not normalize half-forward to UTC")
    target = save("target", command("create-omitted-due", ["create", "B target", "--id", "beads/target"]))
    empty = save("empty", command("create-empty-due", ["create", "C empty", "--id", "beads/empty", "--due="]))
    future = save("future", command("create-future", ["create", "D future", "--id", "beads/future", "--due", "2100-01-01T00:00:00Z"]))
    past = save("past-open", command("create-past", ["create", "E closed", "--id", "beads/closed", "--due", "1999-01-01T00:00:00Z"]))
    c0.require("due_at" not in target["properties"] and "due_at" not in empty["properties"], "omitted/empty create must remain absent")
    for record, title, fields, actor in [
        (original, "A work", {"due_at": "2000-01-02T10:04:06Z", "notes": "Keep original notes"}, "initial-author"),
        (target, "B target", {}, "C0 Smoke"), (empty, "C empty", {}, "C0 Smoke"),
        (future, "D future", {"due_at": "2100-01-01T00:00:00Z"}, "C0 Smoke"),
        (past, "E closed", {"due_at": "1999-01-01T00:00:00Z"}, "C0 Smoke"),
    ]:
        props = record["properties"]
        expected = {"id": props["id"], "created_at": props["created_at"], "updated_at": props["updated_at"],
                    "title": title, "status": "open", "priority": 2, "issue_type": "task",
                    "created_by": actor, "owner": "c0@example.invalid", **fields}
        c0.require(props == expected and record["owned"] == [] and record["attribution"]["actor"] == actor,
                   "initial complete properties, due presence or native author defaults differ")
    memory = save("memory", command("create-memory", ["remember", "Keep this body", "--id", "beads/context", "--title", "Context"]))
    dep_result = command("add-dependency", ["dep", "add", original["id"], target["id"]])
    source, dependency = save("source-owned", dep_result["source"]), save("dependency", dep_result["link"])
    c0.require(source["properties"] == original["properties"] and source["owned"] == [dependency], "dependency changed source properties")
    context_result = command("add-context", ["link", source["id"], memory["id"], "--id", "links/context",
                             "--resource-type", SCOPE + "types/preview-related-v2", "--properties", '{"note":"unchanged"}'])
    context = save("context-link", context_result["link"])
    c0.require(context_result["source"] == source, "informational link unexpectedly owns source")
    current("created-source-current", source)
    exact("initial-source-exact", original)
    close_result = command("close-past", ["close", past["id"], "--reason", "Verified", "--actor", "closer"])
    props = close_result["issue"]["properties"]
    c0.require(props.get("closed_at"), "closed fixture missing actual close time")
    closed = accepted("past-closed", past, close_result, {"status": "closed", "closed_at": props["closed_at"], "close_reason": "Verified"}, "closer")
    capture.passed("normal CLI authoring preserves initial due UTC seconds, nullable presence, owned Dependency and mixed Memory context")

    old = source
    source = edit("guarded-set", source, ["--due", "2030-01-02T03:04:05.499999-07:00"], {"due_at": "2030-01-02T10:04:05Z"})
    source = edit("equivalent-offset-noop", source, ["--due", "2030-01-02T11:04:05+01:00"], {}, changed=False)
    refuse("stale-same-due", ["update", source["id"], "--due", "2030-01-02T10:04:05Z", "--if-revision", old["revision"]], "revision_conflict", 4)
    current("stale-source-preserved", source)
    source = edit("unconditional-clear", source, ["--due="], {"due_at": None}, unconditional=True)
    source = edit("empty-clear-noop", source, ["--due="], {}, changed=False)
    source = edit("unconditional-set", source, ["--due", "2000-01-02T10:04:06.4Z"], {"due_at": "2000-01-02T10:04:06Z"}, unconditional=True)
    source = edit("guarded-clear", source, ["--due="], {"due_at": None})
    source = edit("mixed-set", source, ["--due", "2000-01-02T10:04:05.5Z", "--title", "A revised", "--priority", "0"],
                  {"due_at": "2000-01-02T10:04:06Z", "title": "A revised", "priority": 0})
    source = edit("omission-preserves", source, ["--description", "Only description changes"], {"description": "Only description changes"})
    capture.passed("guarded/unconditional set and clear, equal-instant no-op, stale same-value refusal and mixed/omitted fields preserve full records")

    all_rows = [source, target, empty, future, closed]
    listing("all-current", all_rows)
    listing("default-open", [source, target, empty, future], all_states=False)
    listing("strict-before", [closed], ["--due-before", "2000-01-02T10:04:06Z"])
    listing("fractional-before-truncates", [closed], ["--due-before", "2000-01-02T10:04:06.9Z"])
    listing("strict-after", [future], ["--due-after", "2000-01-02T03:04:06-07:00"])
    listing("fractional-after-truncates", [source, future], ["--due-after", "2000-01-02T10:04:05.9Z"])
    listing("range-intersection", [source], ["--due-after", "1999-01-01T00:00:00Z", "--due-before", "2001-01-01T00:00:00Z"])
    listing("closed-before", [closed], ["--status", "closed", "--due-before", "2001-01-01T00:00:00Z"])
    listing("overdue-excludes-closed", [source], ["--overdue"])
    listing("closed-overdue-empty", [], ["--status", "closed", "--overdue"])
    listing("false-overdue-empty-bounds", all_rows, ["--overdue=false", "--due-before=", "--due-after="])
    listing("limited-page", all_rows, limit=2)
    capture.passed("strict UTC-truncated before/after and native overdue intersections return complete current records with honest hasMore")

    for label, flags, code, status in [
        ("missing-guard", ["--due="], "invalid_selector", 2),
        ("both-guards", ["--due=", "--if-revision", source["revision"], "--unconditional"], "invalid_selector", 2),
        ("invalid-date", ["--due=not-a-date", "--unconditional"], "invalid_properties", 2),
        ("upper-rounding-overflow", ["--due=9999-12-31T23:59:59.5Z", "--title=Must not change", "--unconditional"], "invalid_properties", 2),
        ("readonly", ["--due=", "--unconditional", "--readonly"], "permission_denied", 5),
        ("held-defer", ["--due=", "--defer=tomorrow", "--unconditional"], "capability_unavailable", 5),
    ]:
        refuse(label, ["update", source["id"], *flags], code, status)
    current("all-refusals-preserved", source)
    for label, record in [("target-unchanged", target), ("empty-unchanged", empty), ("future-unchanged", future),
                          ("memory-unchanged", memory), ("dependency-unchanged", dependency), ("context-unchanged", context)]:
        current(label, record)
    capture.passed("invalid/held/readonly/stale writes publish no partial state and leave all unrelated records unchanged")

    for index, entry in enumerate(saved):
        exact("saved-version-" + str(index), entry["record"])
    capture.passed("every saved Issue, Memory and Link version remains exactly readable from a fresh process")
    return {"saved_versions": len(saved), "exact_reads": len(saved) + 1, "list_calls": len(lists), "http_exercised": False,
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
