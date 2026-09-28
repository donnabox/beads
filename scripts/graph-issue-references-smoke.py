#!/usr/bin/env python3
"""Compact installed graph Issue reference-pair proof, with normal CLI-only authoring.

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
SCOPE = "https://example.invalid/disposable-issue-references/"
LIMITATIONS = [
    "reference fields are literal Issue properties, not identity, aliases, document fetching or tracker synchronization",
    "empty clears are verified as JSON omission; SQL NULL versus empty-string storage and exact column bounds require separate storage tests",
    "complete current and exact saved CLI records are not public ordered History or HTTP write proof",
    "visible no-op equality does not establish internal event/version counts; separate storage tests own rollback, overlap and uncertain-COMMIT gates",
    "claim and closed-state checks preserve observed fields, without heartbeat, renewal, reclaim or expiry recovery claims",
    "ordinary Dolt 2.1.8 database provisioning is sequential; no automatic replay after an uncertain result",
]


def semantic_digest(value):
    return hashlib.sha256(json.dumps(value, ensure_ascii=False, sort_keys=True,
                                     separators=(",", ":")).encode()).hexdigest()


def exercise(capture):
    backend = "server" if capture.args.server_port else "embedded"
    saved = []

    def command(label, argv):
        return c0.envelope(capture.success(label, [*argv, "--json"]))

    def raw(label, argv):
        receipt, _, _ = capture.run(label, argv)
        folder = capture.output / capture.records[-1]["artifact"]
        return receipt, (folder / "stdout.log").read_bytes(), (folder / "stderr.log").read_bytes()

    def save(label, record):
        c0.require(record.get("id", "").startswith(SCOPE) and record.get("version") and
                   record.get("revision") and record.get("type") and record.get("attribution"),
                   label + ": incomplete accepted record")
        saved.append({"label": label, "record": copy.deepcopy(record), "semantic_sha256": semantic_digest(record)})
        c0.write_json(capture.output / "expected-records.json", saved)
        return record

    def current(label, record):
        c0.require(command(label, ["show", record["id"], "--readonly"]) == record,
                   label + ": fresh current record differs")

    def exact(label, record):
        c0.require(command(label, ["show", record["id"], "--version", record["version"], "--readonly"]) == record,
                   label + ": complete retained record differs")

    def ready(label, records):
        response = capture.success(label, ["ready", "--readonly", "--json"])
        c0.require(response.get("schemaVersion") == 1 and response.get("preview") is True and
                   isinstance(response.get("result"), list), label + ": invalid ready envelope")
        rows = response["result"]
        wanted = {row["id"]: row for row in records}
        c0.require(len(rows) == len(wanted) and {row["id"] for row in rows} == set(wanted) and
                   all(row == wanted[row["id"]] for row in rows), label + ": complete readiness result differs")

    def accepted(label, before, result, fields, actor):
        c0.require(set(result) == {"issue", "changed"} and result["changed"] is True,
                   label + ": missing changed Issue mutation")
        after = result["issue"]
        expected = copy.deepcopy(before)
        for key, value in fields.items():
            if value is None:
                expected["properties"].pop(key, None)
            else:
                expected["properties"][key] = value
        expected["properties"]["updated_at"] = after["properties"]["updated_at"]
        for key in ["version", "revision", "attribution"]:
            expected[key] = after[key]
        c0.require(after == expected and after["version"] == after["revision"] and
                   after["version"] != before["version"] and after["attribution"]["actor"] == actor,
                   label + ": wrong reference, extra changed property, ownership or attribution")
        return save(label, after)

    def edit(label, before, flags, fields, *, unconditional=False, human=False, quiet=False, actor="reference-author"):
        changed = any((key in before["properties"] if value is None else
                       key not in before["properties"] or before["properties"][key] != value)
                      for key, value in fields.items())
        argv = ["update", before["id"], *flags, "--actor", actor]
        argv += ["--unconditional"] if unconditional else ["--if-revision", before["revision"]]
        if quiet:
            argv.append("--quiet")
        if human:
            receipt, out, err = raw(label, argv)
            want = b"" if quiet else (("Updated" if changed else "Unchanged") + " " + before["id"] + "\n").encode()
            c0.require(receipt["exit_code"] == 0 and out == want and err == b"", label + ": misleading human/quiet output")
            after = command(label + "-fresh", ["show", before["id"], "--readonly"])
            result = {"issue": after, "changed": changed}
        else:
            result = command(label, argv)
            c0.require(set(result) == {"issue", "changed"} and result["changed"] is changed,
                       label + ": changed flag differs from requested fields")
            current(label + "-fresh", result["issue"])
        if changed:
            return accepted(label, before, result, fields, actor)
        c0.require(result == {"issue": before, "changed": False}, label + ": no-op changed complete record")
        return before

    def refuse(label, argv, code, exit_code):
        receipt, out, err = raw(label, [*argv, "--json"])
        c0.require(receipt["exit_code"] == exit_code and out == b"", label + ": wrong failure status/output")
        problem = json.loads(err)
        c0.require(problem.get("code") == code and problem.get("retryable") is False and
                   "result" not in problem, label + ": wrong typed refusal")

    init = ["init", "--prefix", "iref", "--non-interactive", "--skip-hooks", "--skip-agents",
            "--graph-mode", "link", "--scope-url", SCOPE]
    if capture.args.server_port:
        init += ["--server", "--external", "--server-host", "127.0.0.1", "--server-port", str(capture.args.server_port),
                 "--database", "iref_" + capture.root.name.replace("-", "_"), "--server-user", "root"]
    initialized = command("normal-init", init)
    c0.require(initialized.get("scope") == SCOPE and initialized.get("backend") == backend, "wrong initialized authority")
    caps = command("reference-capabilities", ["status", "--graph"])["capabilities"]
    c0.require(caps.get("issueReferenceUpdate") is True and caps.get("issueWorkflows") is False and
               caps.get("historyExact") is False, "reference capability overstates destination")
    source = save("source-created", command("create-source", ["create", "Reference work — 雪", "--id", "beads/source", "--priority", "3"]))
    target = save("target-created", command("create-target", ["create", "Prerequisite", "--id", "beads/target", "--priority", "1"]))
    memory = save("memory", command("create-memory", ["remember", "Keep this body", "--id", "beads/context", "--title", "Context"]))
    added = command("add-dependency", ["dep", "add", source["id"], target["id"]])
    source, dependency = save("source-owned", added["source"]), save("dependency", added["link"])
    linked = command("add-context", ["link", source["id"], memory["id"], "--resource-type", SCOPE + "types/preview-related-v2",
                          "--id", "links/context", "--properties", '{"note":"unchanged"}'])
    context = save("context-link", linked["link"])
    c0.require(linked["source"] == source and source["owned"] == [dependency], "unexpected informational ownership")
    c0.require("external_ref" not in source["properties"] and "spec_id" not in source["properties"], "new Issue has unexpected references")
    ready("initial-blocked", [target])
    initial = copy.deepcopy(source)
    capture.passed("normal initialization creates a reference-free Issue with owned Dependency and unrelated Memory/context")

    external = "  GH — 雪  "
    spec_id = "  specification — café e\u0301\r\nsection  "
    source = edit("set-external-literal", source, ["--external-ref", external], {"external_ref": external})
    external_only = copy.deepcopy(source)
    source = edit("set-spec-literal", source, ["--spec-id", spec_id], {"spec_id": spec_id})
    unicode_pair = copy.deepcopy(source)
    source = edit("replace-external-human", source, ["--external-ref", "-"], {"external_ref": "-"}, unconditional=True, human=True)
    source = edit("replace-spec-literal-file-marker", source, ["--spec-id", "@missing-spec.md"], {"spec_id": "@missing-spec.md"})
    replaced = copy.deepcopy(source)
    source = edit("omitted-pair-preserves", source, ["--title", "References remain literal"], {"title": "References remain literal"})
    source = edit("same-pair-other-actor-noop", source, ["--external-ref", "-", "--spec-id", "@missing-spec.md"],
                  {"external_ref": "-", "spec_id": "@missing-spec.md"}, actor="another-author")
    source = edit("clear-external-only", source, ["--external-ref="], {"external_ref": None})
    source = edit("clear-spec-only", source, ["--spec-id="], {"spec_id": None})
    cleared = copy.deepcopy(source)
    capture.passed("each reference sets, replaces and clears independently; Unicode/whitespace/dash/file-marker stay literal and omitted fields/no-ops preserve complete records")

    shared_fields = {"external_ref": "shared ticket — 雪", "spec_id": "shared specification"}
    shared_flags = ["--external-ref", shared_fields["external_ref"], "--spec-id", shared_fields["spec_id"]]
    source = edit("set-pair-together", source, shared_flags, shared_fields)
    target = edit("same-pair-on-second-issue", target, shared_flags, shared_fields)
    source = edit("clear-pair-together", source, ["--external-ref=", "--spec-id="], {"external_ref": None, "spec_id": None})
    target = edit("second-issue-clear-external", target, ["--external-ref="], {"external_ref": None})
    target = edit("second-issue-clear-spec", target, ["--spec-id="], {"spec_id": None})
    source = edit("repeat-clear-quiet-noop", source, ["--external-ref=", "--spec-id="],
                  {"external_ref": None, "spec_id": None}, human=True, quiet=True)
    capture.passed("two distinct Issues may share both reference values, then clear independently; repeated clear is a quiet no-op with JSON omission")

    claim_result = command("claim-before-reference", ["update", source["id"], "--claim", "--actor", "lease-holder"])
    props = claim_result["issue"]["properties"]
    c0.require(props.get("status") == "in_progress" and props.get("assignee") == "lease-holder" and
               props.get("started_at") and props.get("heartbeat_at") and props.get("lease_expires_at"), "claim fixture missing lease")
    fields = {"status": "in_progress", "assignee": "lease-holder"}
    fields.update({key: props.get(key) for key in ["started_at", "heartbeat_at", "lease_expires_at", "lease_granted_node"]})
    source = accepted("claimed-source", source, claim_result, fields, "lease-holder")
    source = edit("pair-while-claimed", source, shared_flags, shared_fields)
    claimed_after = copy.deepcopy(source)
    ready("still-blocked-after-reference", [target])

    def close(label, before):
        result = command(label, ["close", before["id"], "--reason", "Verified", "--actor", "closer"])
        props = result["issue"]["properties"]
        c0.require(props.get("closed_at"), label + ": missing closed time")
        fields = {"status": "closed", "closed_at": props["closed_at"], "close_reason": "Verified"}
        for key in ["lease_expires_at", "heartbeat_at", "lease_granted_node"]:
            c0.require(not props.get(key), label + ": closed Issue retained lease")
            fields[key] = None
        return accepted(label, before, result, fields, "closer")

    target = close("close-prerequisite", target)
    source = close("close-source", source)
    source = edit("replace-pair-while-closed", source, ["--external-ref", external, "--spec-id", spec_id],
                  {"external_ref": external, "spec_id": spec_id})
    closed_pair = copy.deepcopy(source)
    source = edit("clear-pair-while-closed", source, ["--external-ref=", "--spec-id="], {"external_ref": None, "spec_id": None})
    capture.passed("reference edits preserve actual claim lease/start fields and closed time/reason, owned Dependencies and canonical identity")

    same_clear = ["update", source["id"], "--external-ref=", "--spec-id="]
    for label, argv, code, status in [
        ("stale-equal-clear", [*same_clear, "--if-revision", closed_pair["revision"]], "revision_conflict", 4),
        ("external-overlong", ["update", source["id"], "--external-ref", "雪" * 256, "--title=Must not change", "--unconditional"], "invalid_properties", 2),
        ("spec-overlong", ["update", source["id"], "--spec-id", "雪" * 1025, "--external-ref=sibling-must-not-change", "--unconditional"], "invalid_properties", 2),
        ("claim-with-reference", ["update", source["id"], "--claim", "--external-ref=ignored"], "capability_unavailable", 5),
        ("false-claim-with-spec", ["update", source["id"], "--claim=false", "--spec-id=ignored"], "invalid_properties", 2),
        ("missing-guard", same_clear, "invalid_selector", 2),
        ("readonly", [*same_clear, "--unconditional", "--readonly"], "permission_denied", 5),
        ("create-with-external", ["create", "Extra", "--id=beads/extra", "--external-ref=x", "--notes=admitted", "--status=open"], "capability_unavailable", 5),
        ("create-with-spec", ["create", "Extra", "--id=beads/extra", "--spec-id=x", "--notes=admitted", "--status=open"], "capability_unavailable", 5),
        ("memory-kind", ["update", memory["id"], "--external-ref=invalid", "--unconditional"], "capability_unavailable", 5),
    ]:
        refuse(label, argv, code, status)
    for label, record in [("source", source), ("target", target), ("memory", memory), ("dependency", dependency), ("context", context)]:
        current("final-current-" + label, record)
    capture.passed("stale equal clears, column overlength with sibling edits, mixed claim, missing guard, readonly and Memory-kind writes refuse without state changes")

    exact_records = [initial, external_only, unicode_pair, replaced, cleared, claimed_after, closed_pair, source]
    for index, record in enumerate(exact_records):
        exact("saved-version-" + str(index), record)
    comparisons = []
    for label, before, after in [("compare-absent-to-set", initial, unicode_pair), ("compare-pair-to-cleared", closed_pair, source)]:
        result = command(label, ["compare", source["id"], "--from", before["version"], "--to", after["version"], "--readonly"])
        changes = []
        for key in sorted(before["properties"].keys() | after["properties"].keys(), key=lambda value: value.encode("utf-16-be")):
            left, right = {"present": key in before["properties"]}, {"present": key in after["properties"]}
            if left["present"]:
                left["value"] = before["properties"][key]
            if right["present"]:
                right["value"] = after["properties"][key]
            if left != right:
                changes.append({"area": "properties", "member": key, "from": left, "to": right})
        expected = {"resource": {"id": source["id"], "type": source["type"]},
                    "from": {"version": before["version"], "attribution": before["attribution"]},
                    "to": {"version": after["version"], "attribution": after["attribution"]},
                    "compared": ["properties", "owned"], "unsupported": ["commonMetadata"], "changes": changes}
        c0.require(before["owned"] == after["owned"] and result == expected, label + ": wrong complete comparison")
        comparisons.append({"label": label, "result": result, "semantic_sha256": semantic_digest(result)})
    c0.write_json(capture.output / "comparisons.json", comparisons)
    capture.passed("eight complete saved records retain literal/set/cleared/claimed/closed state; comparisons distinguish absent, present and cleared properties")
    return {"saved_versions": len(saved), "exact_reads": len(exact_records), "comparison_count": len(comparisons), "http_exercised": False,
            "expected_records_sha256": c0.sha256(capture.output / "expected-records.json"),
            "comparisons_sha256": c0.sha256(capture.output / "comparisons.json")}


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
