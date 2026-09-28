#!/usr/bin/env python3
"""Installed graph Issue standalone claim, lifecycle and retained-state proof.

Normal CLI initialization and authoring; both engines execute sequentially.
No SQL fixtures, hidden schema bootstrap, mock or HTTP write demonstration.
"""
import argparse
import copy
from datetime import datetime
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
SCOPE = "https://example.invalid/disposable-issue-claim/"
LIMITATIONS = [
    "standalone canonical Issue claim only; no mixed edits, forced assignment, TTL override, heartbeat or reclaim surface",
    "ordinary native five-minute lease; a same-owner claim is a no-op and does not renew it",
    "actor equivalence reuses existing native identity behavior; no new actor naming contract",
    "sequential fresh CLI processes only; forced writer overlap, rollback, lease-row counts and uncertain COMMIT belong to separate real-store tests",
    "visible complete-record equality does not independently establish internal event or native version row counts",
    "no application retry after an uncertain result; no public HTTP write or full Memory/History claim",
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
                     "capability_unavailable": 5, "permission_denied": 5, "not_found": 3, "constraint_violation": 4}[code]
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
                   all(row == wanted[row["id"]] for row in rows), label + ": claim changed readiness or complete ready records")

    def listing(label, records, flags=None, has_more=False):
        args = ["list", "--format", "records-json", "--all", "--limit", "0", "--sort", "priority", *(flags or [])]
        response = capture.success(label, args)
        c0.require(set(response) == {"schemaVersion", "preview", "result"} and response["schemaVersion"] == 1 and
                   response["preview"] is True, label + ": invalid list envelope")
        result = response["result"]
        c0.require(result == {"items": records, "hasMore": has_more}, label + ": list order/filter/complete-record mismatch")
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

    args = ["init", "--prefix", "iclaim", "--non-interactive", "--skip-hooks", "--skip-agents", "--json",
            "--graph-mode", "link", "--scope-url", SCOPE]
    if capture.args.server_port:
        args += ["--server", "--external", "--server-host", "127.0.0.1", "--server-port", str(capture.args.server_port),
                 "--database", "iclaim_" + capture.root.name.replace("-", "_"), "--server-user", "root"]
    initialized = c0.envelope(capture.success("normal-init", args))
    c0.require(initialized.get("scope") == SCOPE and initialized.get("backend") == backend, "wrong authority/backend")
    caps = command("claim-capabilities", ["status", "--graph"])["capabilities"]
    c0.require(caps.get("issueClaim") is True and caps.get("issueAssigneeFilter") is True and
               caps.get("issueWorkflows") is False and caps.get("memory") is False and caps.get("historyExact") is False,
               "claim capability overstates destination")
    actor, equivalent = "crew.alice", "crew_alice"
    source = save("source-created", command("create-source", ["create", "Claimable work — 雪", "--id", "beads/source",
        "--description", "Preserve claim context", "--labels", "retained,work", "--priority", "3"]))
    target = save("target-created", command("create-target", ["create", "Prerequisite", "--id", "beads/target", "--priority", "1"]))
    memory = save("memory", command("create-memory", ["remember", "Unchanged body — 雪", "--id", "beads/context", "--title", "Context"]))
    added = command("create-dependency", ["dep", "add", source["id"], target["id"], "--actor", "dependency-author"])
    source, dependency = save("source-owned", added["source"]), save("dependency", added["link"])
    related = command("create-context-link", ["link", source["id"], memory["id"], "--resource-type",
        SCOPE + "types/preview-related-v2", "--id", "links/context", "--properties", '{"note":"preserve"}'])
    context = save("context-link", related["link"])
    c0.require(related["source"] == source and source["owned"] == [dependency], "informational Link changed Issue ownership")

    def unchanged(label):
        current(label, [source, target, memory, dependency, context])

    unchanged("authored-baseline")
    ready("initial-blocked", [target])
    listing("initial-unassigned", [target, source], ["--no-assignee"])
    capture.passed("normal initialization authors blocked open Issue, prerequisite, complete owned Dependency and separate Memory/context Link")

    def accepted(label, before, result, fields, claimed_actor):
        c0.require(set(result) == {"issue", "changed"} and result["changed"] is True, label + ": invalid changed result")
        after = result["issue"]
        wanted = copy.deepcopy(before)
        for key, value in fields.items():
            if value is None:
                wanted["properties"].pop(key, None)
            else:
                wanted["properties"][key] = value
        wanted["properties"]["updated_at"] = after["properties"]["updated_at"]
        for key in ("version", "revision", "attribution"):
            wanted[key] = after[key]
        c0.require(after == wanted and after["version"] == after["revision"] and
                   after["version"] != before["version"] and after["attribution"]["actor"] == claimed_actor,
                   label + ": changed unrelated complete state or wrong version/actor")
        return save(label, after)

    def claim(label, before, who, changed, human=False, quiet=False):
        argv = ["update", before["id"], "--claim", "--actor", who]
        if quiet:
            argv += ["--quiet"]
        if human:
            receipt, out, err = raw(label, argv)
            c0.require(receipt["exit_code"] == 0 and err == b"", label + ": human claim failed")
            after = command(label + "-show", ["show", before["id"], "--readonly"])
            expected = b"" if quiet else (("Claimed" if changed else "Unchanged") + " " + before["id"] + "\n").encode()
            c0.require(out == expected, label + ": human/quiet output mismatch")
            result = {"issue": after, "changed": changed}
        else:
            result = command(label, argv)
        c0.require(set(result) == {"issue", "changed"} and result["changed"] is changed, label + ": claim envelope")
        if changed:
            after = result["issue"]
            props = after["properties"]
            c0.require(props.get("assignee") == who and props.get("status") == "in_progress" and
                       props.get("started_at") and props.get("heartbeat_at") and props.get("lease_expires_at"),
                       label + ": missing native claim state")
            stamp = lambda value: datetime.fromisoformat(value.replace("Z", "+00:00"))
            ttl = (stamp(props["lease_expires_at"]) - stamp(props["heartbeat_at"])).total_seconds()
            c0.require(299 <= ttl <= 301, label + ": default lease is not five minutes")
            fields = {"assignee": who, "status": "in_progress"}
            for key in ("started_at", "heartbeat_at", "lease_expires_at", "lease_granted_node"):
                fields[key] = props.get(key)
            if before["properties"].get("started_at"):
                c0.require(props["started_at"] == before["properties"]["started_at"], label + ": reclaimed Issue lost original start time")
            result["issue"] = accepted(label, before, result, fields, who)
        else:
            c0.require(result["issue"] == before, label + ": no-op changed version, attribution, content or lease")
        exact(label + "-prior", before)
        exact(label + "-accepted", result["issue"])
        compare(label + "-compare", before, result["issue"])
        return result["issue"]

    open_source = source
    source = claim("claim-blocked", source, actor, True)
    unchanged("claimed-fresh")
    ready("claimed-still-blocked", [target])
    listing("claimed-by-actor", [source], ["--status", "in_progress", "--assignee", actor])
    listing("remaining-unassigned", [target], ["--no-assignee"])
    claimed_source = source
    source = claim("same-owner-noop", source, actor, False)
    source = claim("equivalent-owner-noop", source, equivalent, False)
    source = claim("quiet-json-noop", source, actor, False, quiet=True)
    source = claim("human-noop", source, actor, False, human=True)
    source = claim("quiet-human-noop", source, actor, False, human=True, quiet=True)
    c0.require(source == claimed_source, "repeat claim renewed the lease or changed exact state")
    unchanged("after-idempotent-claims")
    capture.passed("blocked Issue remains claimable; standalone claim has a five-minute lease and same/equivalent actor claims preserve the complete accepted record without renewal")

    common = ["update", source["id"], "--claim", "--actor", actor]
    refuse("foreign-holder", ["update", source["id"], "--claim", "--actor", "crew.bob"], "constraint_violation")
    refuse("claim-false", ["update", source["id"], "--claim=false", "--actor", actor], "invalid_properties")
    refuse("claim-memory", ["update", memory["id"], "--claim", "--actor", actor], "invalid_properties")
    refuse("claim-link", ["update", context["id"], "--claim", "--actor", actor], "invalid_selector")
    refuse("claim-missing", ["update", "beads/missing", "--claim", "--actor", actor], "not_found")
    refuse("claim-legacy-selector", ["update", "iclaim-legacy", "--claim", "--actor", actor], "invalid_selector")
    refuse("claim-overlength-actor", ["update", source["id"], "--claim", "--actor", "雪" * 256], "invalid_properties")
    mixed = [
        ("title", ["--title", "Denied"]), ("description", ["--description", ""]),
        ("priority", ["--priority", "0"]), ("assignee", ["--assignee", actor]),
        ("status", ["--status", "in_progress"]), ("labels", ["--set-labels="]),
        ("force", ["--force"]), ("force-false", ["--force=false"]),
        ("revision", ["--if-revision", source["revision"]]), ("empty-revision", ["--if-revision", ""]),
        ("unconditional", ["--unconditional"]), ("unconditional-false", ["--unconditional=false"]),
        ("assignee-guard", ["--if-assignee", ""]), ("status-guard", ["--if-status", "in_progress"]),
        ("body-file", ["--body-file", str(capture.root / "does-not-exist.md")]),
        ("stdin-false", ["--stdin=false"]),
    ]
    for label, flags in mixed:
        refuse("mixed-" + label, [*common, *flags], "capability_unavailable")
    refuse("readonly-claim", [*common, "--readonly"], "permission_denied")
    original_readonly = capture.env.get("BD_READONLY")
    try:
        capture.env["BD_READONLY"] = "1"
        refuse("environment-readonly-claim", common, "permission_denied")
    finally:
        if original_readonly is None:
            capture.env.pop("BD_READONLY", None)
        else:
            capture.env["BD_READONLY"] = original_readonly
    mayor, freeze = capture.work / "mayor", capture.work / "MIGRATION-FREEZE"
    c0.require(not mayor.exists() and not freeze.exists(), "unexpected migration markers")
    try:
        mayor.mkdir()
        (mayor / "town.json").write_text("{}\n")
        freeze.write_text("disposable claim refusal proof\n")
        refuse("migration-freeze-claim", common, "permission_denied")
    finally:
        freeze.unlink(missing_ok=True)
        (mayor / "town.json").unlink(missing_ok=True)
        mayor.rmdir()
    # update has no TTL override flag; this is ordinary parser admission,
    # explicitly not a graph error-envelope claim.
    receipt, out, err = raw("unregistered-ttl", [*common, "--ttl", "1m", "--json"])
    c0.require(receipt["exit_code"] != 0 and out == b"" and b"unknown flag: --ttl" in err,
               "unregistered TTL unexpectedly accepted or lost parser diagnostic")
    unchanged("all-refusals-unchanged")
    exact("claim-prior-open-still-exact", open_source)
    exact("claim-accepted-still-exact", claimed_source)
    capture.passed("foreign holder, unsupported kinds/selectors, mixed fields/guards, readonly/freeze and TTL override refuse without mutating complete graph records")

    def transition(label, verb, before, who):
        reason = label + " — 雪"
        result = command(label, [verb, before["id"], "--reason", reason, "--actor", who])
        after = result["issue"]
        fields = {"status": "closed" if verb == "close" else "open"}
        if verb == "close":
            c0.require(after["properties"].get("closed_at") and after["properties"].get("close_reason") == reason,
                       label + ": close timestamp/reason missing")
            fields.update({"closed_at": after["properties"]["closed_at"], "close_reason": reason})
        else:
            for key in ("closed_at", "close_reason", "closed_by_session", "defer_until"):
                c0.require(not after["properties"].get(key), label + ": closure property retained")
                fields[key] = None
        for key in ("lease_expires_at", "heartbeat_at", "lease_granted_node"):
            c0.require(not after["properties"].get(key), label + ": ended claim retained live lease projection")
            fields[key] = None
        after = accepted(label, before, result, fields, who)
        exact(label + "-prior", before)
        compare(label + "-compare", before, after)
        return after

    refuse("blocked-close-refused", ["close", source["id"], "--actor", actor], "constraint_violation")
    unchanged("blocked-close-unchanged")
    target = transition("close-prerequisite", "close", target, "prerequisite-author")
    unchanged("closed-target-current")
    source = transition("close-claimed-source", "close", source, actor)
    ready("both-closed", [])
    refuse("closed-claim-refused", ["update", source["id"], "--claim", "--actor", actor], "constraint_violation")
    unchanged("closed-refusal-unchanged")
    source = transition("reopen-claimed-source", "reopen", source, actor)
    ready("reopened-unblocked", [source])
    listing("reopened-retains-assignee", [source], ["--status", "open", "--assignee", actor])
    source = claim("human-reclaim-open", source, actor, True, human=True)
    listing("reclaimed-in-progress", [source], ["--status", "in_progress", "--assignee", actor])
    unchanged("reclaimed-current")
    exact("original-claim-after-close-reopen", claimed_source)
    capture.passed("closing the prerequisite permits close of the claimed Issue; close removes the lease, reopen preserves assignee and first start time, and explicit claim accepts the reopened Issue")
    for index, item in enumerate(saved):
        exact("final-saved-" + str(index), item["record"])
    unchanged("final-complete-current")
    capture.passed("every saved version remains byte-semantically exact in a new process while owned Dependency, Memory and informational Link retain their complete state")
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
