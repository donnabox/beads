#!/usr/bin/env python3
"""Installed graph Issue assignee edit and exact-version workflow proof.

Normal CLI initialization and authoring; both engines execute sequentially.
No SQL fixtures, hidden schema bootstrap, mock or HTTP write demonstration.
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
SCOPE = "https://example.invalid/disposable-issue-assignee/"
LIMITATIONS = [
    "experimental ordinary assignee edits on canonical graph Issues; no claim, force, status edit or assignee-compare-and-set surface",
    "literal valid UTF-8 assignee values retain whitespace; this does not resolve actor identity/canonicalization contracts",
    "assignee-specific list filters remain unavailable; existing title/priority/full-list reads verify complete assigned records",
    "graph create only authors open Issues; active foreign-owner and lease preservation/refusal require separate real-store tests",
    "visible complete-record no-ops and one accepted version do not count internal events or prove concurrent/unknown-COMMIT outcomes",
    "exact saved reads/comparisons are not full Memory or public ordered History; no HTTP write demonstration",
    "ordinary Dolt 2.1.8 database provisioning remains serialized",
]


def semantic_digest(value):
    return hashlib.sha256(json.dumps(value, ensure_ascii=False, sort_keys=True,
                                     separators=(",", ":")).encode()).hexdigest()

def exercise(capture):
    backend = "server" if capture.args.server_port else "embedded"
    saved, pages = [], []

    def raw(label, args):
        receipt, _, _ = capture.run(label, args)
        directory = capture.output / capture.records[-1]["artifact"]
        return receipt, (directory / "stdout.log").read_bytes(), (directory / "stderr.log").read_bytes()

    def command(label, args):
        return c0.envelope(capture.success(label, [*args, "--json"]))

    def refuse(label, args, code):
        receipt, out, err = raw(label, [*args, "--json"])
        exit_code = {"invalid_selector": 2, "invalid_properties": 2, "revision_conflict": 4,
                     "capability_unavailable": 5, "permission_denied": 5, "not_found": 3}[code]
        c0.require(receipt["exit_code"] == exit_code and out == b"", label + ": wrong refusal exit or success output")
        problem = json.loads(err)
        c0.require(problem.get("code") == code and problem.get("retryable") is False and "result" not in problem,
                   label + ": wrong typed refusal")

    def save(label, record):
        c0.require(record.get("id", "").startswith(SCOPE) and record.get("type") and record.get("version") and
                   record.get("revision") and record.get("attribution"), label + ": incomplete accepted record")
        saved.append({"label": label, "record": copy.deepcopy(record), "semantic_sha256": semantic_digest(record)})
        c0.write_json(capture.output / "expected-records.json", saved)
        return record

    def current(label, records):
        for i, record in enumerate(records):
            c0.require(command(label + "-" + str(i), ["show", record["id"], "--readonly"]) == record,
                       label + ": current complete record changed")

    def exact(label, record):
        c0.require(command(label, ["show", record["id"], "--version", record["version"], "--readonly"]) == record,
                   label + ": complete retained record changed")

    def ready(label, records):
        response = capture.success(label, ["ready", "--readonly", "--json"])
        c0.require(set(response) == {"schemaVersion", "preview", "result"} and response["schemaVersion"] == 1 and
                   response["preview"] is True and isinstance(response["result"], list), label + ": invalid ready envelope")
        rows = response["result"]
        wanted = {record["id"]: record for record in records}
        c0.require(len(rows) == len(records) and {r["id"] for r in rows} == set(wanted) and
                   all(row == wanted[row["id"]] for row in rows), label + ": assignee edit changed readiness or complete ready records")

    def listing(label, records, flags=None):
        args = ["list", "--format", "records-json", "--all", "--limit", "0", "--sort", "priority", *(flags or [])]
        response = capture.success(label, args)
        c0.require(set(response) == {"schemaVersion", "preview", "result"} and response["schemaVersion"] == 1 and
                   response["preview"] is True, label + ": invalid list envelope")
        result = response["result"]
        c0.require(result == {"items": records, "hasMore": False}, label + ": list order/filter/complete-record mismatch")
        pages.append({"label": label, "argv": args, "result": result, "semantic_sha256": semantic_digest(result)})
        c0.write_json(capture.output / "list-results.json", pages)
        return result["items"]

    def compare(label, before, after):
        result = command(label, ["compare", before["id"], "--from", before["version"], "--to", after["version"], "--readonly"])
        c0.require(result["resource"] == {"id": before["id"], "type": before["type"]} and
                   result["from"] == {"version": before["version"], "attribution": before["attribution"]} and
                   result["to"] == {"version": after["version"], "attribution": after["attribution"]} and
                   result["compared"] == ["properties", "owned"] and result["unsupported"] == ["commonMetadata"],
                   label + ": wrong comparison identity/context/coverage")
        old, new, changes = before["properties"], after["properties"], []
        for key in sorted(old.keys() | new.keys(), key=lambda x: x.encode("utf-16-be")):
            left, right = {"present": key in old}, {"present": key in new}
            if key in old:
                left["value"] = old[key]
            if key in new:
                right["value"] = new[key]
            if left != right:
                changes.append({"area": "properties", "member": key, "from": left, "to": right})
        c0.require(before["owned"] == after["owned"] and result["changes"] == changes,
                   label + ": comparison changed ownership or lost a property")

    args = ["init", "--prefix", "iassignee", "--non-interactive", "--skip-hooks", "--skip-agents", "--json",
            "--graph-mode", "link", "--scope-url", SCOPE]
    if capture.args.server_port:
        args += ["--server", "--external", "--server-host", "127.0.0.1", "--server-port", str(capture.args.server_port),
                 "--database", "iassignee_" + capture.root.name.replace("-", "_"), "--server-user", "root"]
    initialized = c0.envelope(capture.success("normal-init", args))
    c0.require(initialized.get("scope") == SCOPE and initialized.get("backend") == backend, "wrong authority/backend")
    caps = command("assignee-capabilities", ["status", "--graph"])["capabilities"]
    c0.require(caps.get("issueAssigneeUpdate") is True and caps.get("issueTextUpdate") is True and
               caps.get("issuePriorityUpdate") is True and caps.get("issueList") is True and
               caps.get("issueWorkflows") is False and caps.get("memory") is False and caps.get("historyExact") is False,
               "assignee capability overstates destination")
    source = save("source-created", command("create-source", ["create", "Assignment work", "--id", "beads/source",
        "--description", "Preserve description — 雪", "--labels", "retained,work", "--type", "bug", "--priority", "3"]))
    target = save("target", command("create-target", ["create", "Prerequisite", "--id", "beads/target", "--priority", "1"]))
    memory = save("memory", command("create-memory", ["remember", "Unchanged context — 雪", "--id", "beads/context", "--title", "Context"]))
    relation = command("create-dependency", ["dep", "add", source["id"], target["id"], "--actor", "dependency-author"])
    source, dependency = save("source-owned", relation["source"]), save("dependency", relation["link"])
    relation = command("create-context-link", ["link", source["id"], memory["id"], "--resource-type",
        SCOPE + "types/preview-related-v2", "--id", "links/context", "--properties", '{"note":"preserve context"}'])
    context = save("context-link", relation["link"])
    c0.require(relation["source"] == source and source["owned"] == [dependency], "informational Link changed Issue ownership")
    unrelated = [target, memory, dependency, context]
    current("authored-baseline", [source, *unrelated])
    ready("blocked-before-assignment", [target])
    listing("initial-complete-list", [target, source])
    capture.passed("normal initialization and CLI-authored graph yield an open unassigned blocked Issue with a complete owned Dependency")

    def checked(label, before, after, fields, actor, changed):
        if not changed:
            c0.require(after == before, label + ": no-op changed complete record/version/attribution")
            return before
        wanted = copy.deepcopy(before)
        for key, value in fields.items():
            if key == "title":
                value = value.strip()
            if value == "" and key != "title":
                wanted["properties"].pop(key, None)
            else:
                wanted["properties"][key] = value
        wanted["properties"]["updated_at"] = after["properties"]["updated_at"]
        for key in ["version", "revision", "attribution"]:
            wanted[key] = after[key]
        c0.require(after == wanted and after["version"] != before["version"] and after["version"] == after["revision"] and
                   after["attribution"]["actor"] == actor, label + ": changed unrelated Issue fields/owned state or wrong accepted revision")
        return save(label, after)

    def update(label, before, flags, fields, changed=True, unconditional=False, quiet=False):
        guard = ["--unconditional"] if unconditional else ["--if-revision", before["revision"]]
        actor = label + "-actor"
        argv = ["update", before["id"], *flags, *guard, "--actor", actor]
        if quiet:
            argv += ["--quiet"]
        result = command(label, argv)
        c0.require(set(result) == {"issue", "changed"} and result["changed"] is changed,
                   label + ": wrong Issue mutation envelope")
        after = checked(label, before, result["issue"], fields, actor, changed)
        current(label + "-fresh", [after, *unrelated])
        exact(label + "-prior", before)
        compare(label + "-compare", before, after)
        ready(label + "-ready", [target])
        return after

    before_assignment = source
    source = update("set-assignee-short-alias", source, ["-a", "alice"], {"assignee": "alice"})
    listing("assigned-alice-record", [source], ["--title", "Assignment work"])
    listing("unassigned-target-unchanged", [target], ["--title", "Prerequisite"])
    source = update("different-actor-assignee-noop", source, ["--assignee", "alice"], {"assignee": "alice"}, changed=False)
    refuse("stale-equal-assignee", ["update", source["id"], "--assignee", "alice", "--if-revision", before_assignment["revision"]], "revision_conflict")
    refuse("stale-other-assignee", ["update", source["id"], "--assignee", "bob", "--if-revision", before_assignment["revision"]], "revision_conflict")
    source = update("replace-assignee-unconditional", source, ["--assignee", "bob"], {"assignee": "bob"}, unconditional=True)
    listing("replaced-bob-record", [source], ["--title", "Assignment work"])
    capture.passed("set and replace use existing assignee semantics; no-op preserves complete state and stale equal-value guards still refuse")

    literal = "  Zoë 雪  "
    source = update("literal-unicode-assignee", source, ["--assignee", literal], {"assignee": literal})
    listing("literal-assignee-record", [source], ["--title", "Assignment work"])
    source = update("omit-assignee-text-edit", source, ["--description", "Omission retains literal assignee"],
                    {"description": "Omission retains literal assignee"})
    c0.require(source["properties"]["assignee"] == literal, "omitted assignee was replaced or trimmed")
    source = update("mixed-assignee-title-priority", source,
        ["--assignee", "carol", "--title", "Assigned work — 雪", "--priority", "P0"],
        {"assignee": "carol", "title": "Assigned work — 雪", "priority": 0})
    listing("mixed-assignee-record", [source], ["--priority", "0", "--title", "Assigned work"])
    listing("mixed-priority-order", [source, target])
    source = update("clear-assignee", source, ["--assignee="], {"assignee": ""})
    c0.require("assignee" not in source["properties"], "clear should omit empty optional assignee property")
    listing("cleared-assignee-complete-list", [source, target])
    source = update("already-clear-noop", source, ["--assignee="], {"assignee": ""}, changed=False, unconditional=True)
    capture.passed("literal Unicode/whitespace values survive omitted-field edits; mixed title/priority/assignee returns one accepted revision and clear removes the optional property")

    for label, value, quiet, unconditional in [
        ("human-assignee-set", "human-owner", False, False),
        ("quiet-assignee-noop", "human-owner", True, True),
    ]:
        before = source
        actor = label + "-actor"
        guard = ["--unconditional"] if unconditional else ["--if-revision", source["revision"]]
        argv = ["update", source["id"], "--assignee", value, *guard, "--actor", actor]
        if quiet:
            argv += ["--quiet"]
        receipt, out, err = raw(label, argv)
        c0.require(receipt["exit_code"] == 0 and err == b"", label + ": human assignee update failed")
        after = command(label + "-show", ["show", source["id"], "--readonly"])
        changed = before["properties"].get("assignee", "") != value
        source = checked(label, before, after, {"assignee": value}, actor, changed)
        expected = (("Updated" if changed else "Unchanged") + " " + source["id"] + "\n").encode()
        c0.require(out == (b"" if quiet else expected), label + ": wrong human/quiet output")
        exact(label + "-prior", before)
        compare(label + "-compare", before, source)
    source = update("quiet-json-assignee-noop", source, ["--assignee", "human-owner"],
                    {"assignee": "human-owner"}, changed=False, quiet=True)
    capture.passed("human, quiet and JSON edits use the same complete current/exact Issue state")

    common = ["update", source["id"], "--assignee", "refused"]
    cases = [
        ("missing-guard", common, "invalid_selector"),
        ("empty-guard", [*common, "--if-revision", ""], "invalid_selector"),
        ("both-guards", [*common, "--if-revision", source["revision"], "--unconditional"], "invalid_selector"),
        ("false-unconditional", [*common, "--unconditional=false"], "invalid_selector"),
        ("overlength-assignee", ["update", source["id"], "--assignee", "雪" * 256, "--unconditional"], "invalid_properties"),
        ("memory-kind", ["update", memory["id"], "--assignee", "x", "--unconditional"], "capability_unavailable"),
        ("link-kind", ["update", context["id"], "--assignee", "x", "--unconditional"], "capability_unavailable"),
        ("claim-refused", [*common, "--unconditional", "--claim=false"], "capability_unavailable"),
        ("force-refused", [*common, "--unconditional", "--force=false"], "capability_unavailable"),
        ("status-refused", [*common, "--unconditional", "--status", "in_progress"], "capability_unavailable"),
        ("assignee-guard-refused", [*common, "--unconditional", "--if-assignee", "human-owner"], "capability_unavailable"),
        ("status-guard-refused", [*common, "--unconditional", "--if-status", "open"], "capability_unavailable"),
        ("readonly-refusal", [*common, "--unconditional", "--readonly"], "permission_denied"),
    ]
    for label, argv, code in cases:
        refuse(label, argv, code)
    current("final-current", [source, *unrelated])
    ready("final-readiness", [target])
    listing("final-assignee-record", [source], ["--title", "Assigned work"])
    for i, item in enumerate(saved):
        exact("final-retained-" + str(i), item["record"])
    capture.passed("invalid fields/guards, unsupported ownership commands/kinds and readonly writes refuse; dependencies, targets, Memory and informational Link remain unchanged")
    return {"saved_versions": len(saved), "list_calls": len(pages), "http_exercised": False,
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
