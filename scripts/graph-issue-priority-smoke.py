#!/usr/bin/env python3
"""Installed select/reprioritize/list/exact-version Issue workflow proof.

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
SCOPE = "https://example.invalid/disposable-issue-priority/"
LIMITATIONS = [
    "experimental canonical Issue priority and existing inline text updates, not complete Issue workflows",
    "priority uses the existing permissive numeric/P-prefixed parser; no new strict lexical grammar is asserted",
    "exact retained records/comparisons are not public ordered History or HTTP writes",
    "visible no-op record equality does not prove internal event/version counts; separate tests cover priority rollback/overlap; uncertain-COMMIT behavior relies on existing shared transaction controls",
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
                   all(row == wanted[row["id"]] for row in rows), label + ": priority changed readiness or complete ready records")

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

    args = ["init", "--prefix", "ipriority", "--non-interactive", "--skip-hooks", "--skip-agents", "--json",
            "--graph-mode", "link", "--scope-url", SCOPE]
    if capture.args.server_port:
        args += ["--server", "--external", "--server-host", "127.0.0.1", "--server-port", str(capture.args.server_port),
                 "--database", "ipriority_" + capture.root.name.replace("-", "_"), "--server-user", "root"]
    initialized = c0.envelope(capture.success("normal-init", args))
    c0.require(initialized.get("scope") == SCOPE and initialized.get("backend") == backend, "wrong authority/backend")
    caps = command("priority-capabilities", ["status", "--graph"])["capabilities"]
    c0.require(caps.get("issuePriorityUpdate") is True and caps.get("issueTextUpdate") is True and
               caps.get("issueList") is True and caps.get("issueWorkflows") is False and
               caps.get("memory") is False and caps.get("historyExact") is False, "priority capability overstates destination")
    source = save("source-created", command("create-source", ["create", "Selected source", "--id", "beads/source",
        "--description", "Preserve this description", "--labels", "retained,work", "--type", "bug", "--priority", "3"]))
    target = save("target", command("create-target", ["create", "First prerequisite", "--id", "beads/target", "--priority", "1"]))
    peer = save("peer", command("create-peer", ["create", "Other work", "--id", "beads/peer", "--priority", "2"]))
    memory = save("memory", command("create-memory", ["remember", "Context body — 雪", "--id", "beads/context", "--title", "Context"]))
    relation = command("create-dependency", ["dep", "add", source["id"], target["id"], "--actor", "dependency-author"])
    source, dependency = save("source-owned", relation["source"]), save("dependency", relation["link"])
    relation = command("create-context-link", ["link", source["id"], memory["id"], "--resource-type",
        SCOPE + "types/preview-related-v2", "--id", "links/context", "--properties", '{"note":"preserve context"}'])
    context = save("context-link", relation["link"])
    c0.require(relation["source"] == source and source["owned"] == [dependency], "informational Link altered Issue ownership")
    unrelated = [target, peer, memory, dependency, context]
    selected = listing("select-current-work", [target, peer, source])[-1]
    c0.require(selected == source, "selected wrong complete Issue")
    ready("blocked-before-priority", [target, peer])
    current("authored-baseline", [source, *unrelated])
    capture.passed("normal init authors mixed graph records and selects complete blocked Issue through installed list")

    def checked_change(label, before, after, fields, actor, changed):
        if not changed:
            c0.require(after == before, label + ": no-op minted version/attribution or changed state")
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
        c0.require(after == wanted and after["version"] != before["version"] and after["revision"] == after["version"] and
                   after["attribution"]["actor"] == actor, label + ": changed unintended fields/ownership or wrong version/actor")
        return save(label, after)

    def update(label, before, priority, parsed, fields=None, unconditional=False, changed=True, quiet=False):
        guard = ["--unconditional"] if unconditional else ["--if-revision", before["revision"]]
        flags = ["--priority", priority]
        expected = {"priority": parsed, **(fields or {})}
        for field, value in (fields or {}).items():
            flags += ["--" + {"acceptance_criteria": "acceptance"}.get(field, field), value]
        if quiet:
            flags += ["--quiet"]
        actor = label + "-actor"
        result = command(label, ["update", before["id"], *flags, *guard, "--actor", actor])
        c0.require(set(result) == {"issue", "changed"} and result["changed"] is changed,
                   label + ": wrong Issue mutation or invented replacement disclosure")
        after = checked_change(label, before, result["issue"], expected, actor, changed)
        current(label + "-fresh", [after, *unrelated])
        exact(label + "-old-version", before)
        compare(label + "-compare", before, after)
        listing(label + "-priority-filter", [after], ["--priority", str(parsed), "--title", after["properties"]["title"]])
        return after

    prior = source
    source = update("selected-P0", selected, "P0", 0)
    listing("reprioritized-order", [source, target, peer])
    source = update("different-actor-noop", source, "0", 0, changed=False)
    for label, value in [("stale-priority-change", "4"), ("stale-priority-noop", "P0")]:
        refuse(label, ["update", source["id"], "--priority", value, "--if-revision", prior["revision"]], "revision_conflict")
    for label, value, parsed, fields, unconditional in [
        ("priority-P1-unconditional", "P1", 1, None, True),
        ("priority-P2-and-text", "2", 2, {"title": "Refined source", "description": "Reprioritized description — 雪",
             "design": "Design stays literal", "acceptance_criteria": "Ready for review"}, False),
        ("priority-lowercase-P4", "p4", 4, None, True),
        ("priority-whitespace-P3", " P3 ", 3, None, False),
    ]:
        source = update(label, source, value, parsed, fields, unconditional)
    ready("blocked-after-priorities", [target, peer])
    capture.passed("selected guarded priority update reorders/filter lists; all priorities and combined text preserve ownership/readiness, while no-op/stale guards are exact")

    before_owned = source
    relation = command("second-owned-dependency", ["dep", "add", source["id"], peer["id"], "--actor", "second-dependency-author"])
    source, second_dependency = save("source-two-dependencies", relation["source"]), save("second-dependency", relation["link"])
    c0.require({row["id"] for row in source["owned"]} == {dependency["id"], second_dependency["id"]}, "owned Dependency not installed")
    unrelated = [target, peer, memory, dependency, second_dependency, context]
    refuse("stale-after-owned-change", ["update", source["id"], "--priority", "P3", "--if-revision", before_owned["revision"]],
           "revision_conflict")
    source = update("unconditional-after-owned", source, "P0", 0, unconditional=True)
    ready("blocked-after-owned-priority", [target, peer])
    capture.passed("owned Dependency changes invalidate the selected revision; unconditional priority preserves the current complete owned set and targets")

    for label, value, unconditional, quiet in [
        ("human-changed", 1, False, False), ("human-unconditional-noop", 1, True, False),
        ("quiet-unconditional-changed", 2, True, True), ("quiet-guarded-noop", 2, False, True),
    ]:
        before = source
        guard = ["--unconditional"] if unconditional else ["--if-revision", source["revision"]]
        priority_flag = "-p" if label == "human-changed" else "--priority"
        argv = ["update", source["id"], priority_flag, str(value), *guard, "--actor", label + "-actor"]
        if quiet:
            argv += ["--quiet"]
        receipt, out, err = raw(label, argv)
        c0.require(receipt["exit_code"] == 0 and err == b"", label + ": human/quiet update failed")
        changed = value != before["properties"]["priority"]
        source = command(label + "-show", ["show", source["id"], "--readonly"])
        source = checked_change(label, before, source, {"priority": value}, label + "-actor", changed)
        expected = (("Updated" if changed else "Unchanged") + " " + source["id"] + "\n").encode()
        c0.require(out == (b"" if quiet else expected), label + ": wrong human/quiet result")
        exact(label + "-old", before)
        compare(label + "-compare", before, source)
    source = update("quiet-json-noop", source, "P2", 2, changed=False, quiet=True)
    current("after-human", [source, *unrelated])
    ready("blocked-after-human", [target, peer])
    capture.passed("JSON/human/quiet preserve exact Issue mutation semantics with no Memory predecessor disclosure")

    for label, value in [("priority-negative", "-1"), ("priority-too-large", "5"), ("priority-prefixed-range", "P9"),
                         ("priority-empty", ""), ("priority-word", "high")]:
        refuse(label, ["update", source["id"], "--priority", value, "--unconditional"], "invalid_properties")
    common = ["update", source["id"], "--priority", "P0"]
    for label, argv, code in [
        ("missing-guard", common, "invalid_selector"),
        ("empty-guard", [*common, "--if-revision", ""], "invalid_selector"),
        ("both-guards", [*common, "--if-revision", source["revision"], "--unconditional"], "invalid_selector"),
        ("false-unconditional", [*common, "--unconditional=false"], "invalid_selector"),
        ("mixed-properties", [*common, "--unconditional", "--properties", '{}'], "capability_unavailable"),
        ("memory-kind", ["update", memory["id"], "--priority", "0", "--unconditional"], "capability_unavailable"),
        ("link-kind", ["update", context["id"], "--priority", "0", "--unconditional"], "capability_unavailable"),
        ("unknown-issue", ["update", "beads/missing", "--priority", "0", "--unconditional"], "not_found"),
        ("legacy-selector", ["update", "ipriority-legacy", "--priority", "0", "--unconditional"], "invalid_selector"),
    ]:
        refuse(label, argv, code)
    for flag, value in [("status", "closed"), ("type", "task"), ("if-assignee", "worker"), ("notes", ""),
                        ("add-label", "new"), ("if-source-revision", source["revision"]),
                        ("body-file", str(capture.root / "missing.md"))]:
        refuse("unsupported-" + flag, [*common, "--unconditional", "--" + flag, value], "capability_unavailable")
    for flag in ["--claim=false", "--force=false"]:
        refuse("unsupported-" + flag[2:], [*common, "--unconditional", flag], "capability_unavailable")
    current("refusals-preserve-state", [source, *unrelated])
    capture.passed("invalid ranges/guards, wrong kinds/selectors and unsupported mixed/false flags refuse with exact typed codes and no partial success")

    yaml = capture.work / ".beads" / "config.yaml"
    old_yaml = yaml.read_bytes() if yaml.exists() else None
    old_readonly = capture.env.get("BD_READONLY")
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
                freeze.write_text("issue-priority\t2026-09-27T00:00:00Z\twrite policy\n")
            before = c0.tree_digest(capture.work)
            refuse("policy-" + policy, [*common, "--description", "-", "--unconditional", *flags], "permission_denied")
            c0.require(c0.tree_digest(capture.work) == before, policy + ": refusal changed workspace files")
        finally:
            if old_readonly is None:
                capture.env.pop("BD_READONLY", None)
            else:
                capture.env["BD_READONLY"] = old_readonly
            if old_yaml is None:
                yaml.unlink(missing_ok=True)
            else:
                yaml.write_bytes(old_yaml)
            if policy == "freeze":
                freeze.unlink(missing_ok=True)
                (mayor / "town.json").unlink(missing_ok=True)
                mayor.rmdir()
    current("after-policy", [source, *unrelated])
    capture.passed("readonly flag/environment/configuration and freeze refuse priority mutation before unsupported input acquisition")

    target = save("target-closed", command("close-target", ["close", target["id"], "--reason", "Complete prerequisite"])["issue"])
    peer = save("peer-closed", command("close-peer", ["close", peer["id"], "--reason", "Complete prerequisite"])["issue"])
    source = save("source-closed", command("close-source", ["close", source["id"], "--reason", "Complete selected work"])["issue"])
    closed = source
    unrelated = [target, peer, memory, dependency, second_dependency, context]
    source = update("closed-priority-and-description", source, "P4", 4,
                    {"description": "Post-close priority review — 雪"})
    c0.require(source["properties"]["status"] == "closed" and
               source["properties"]["closed_at"] == closed["properties"]["closed_at"], "priority edit reopened/reclosed Issue")
    ready("closed-priority-stays-not-ready", [])
    listing("closed-priority-visible-with-all", [source], ["--priority", "4", "--title", source["properties"]["title"]])
    current("final-current", [source, *unrelated])
    for i, item in enumerate(saved):
        exact("final-retained-" + str(i), item["record"])
    capture.passed("closed Issue reprioritization preserves lifecycle timestamps and owned Links; every saved predecessor remains exact")
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
