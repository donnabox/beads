#!/usr/bin/env python3
"""Installed graph Issue append-only notes and retained-state proof.

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
SCOPE = "https://example.invalid/disposable-issue-append-notes/"
LIMITATIONS = [
    "append-only Issue notes through the existing transaction operation; replacement, clear, status and mixed claim edits remain unavailable",
    "literal existing append semantics: newline separator on nonempty notes, including empty payload; repeated text is appended again",
    "complete current/retained CLI records and comparisons are not public ordered History or HTTP write demonstrations",
    "sequential installed processes do not force overlap, rollback, cancellation or lost COMMIT acknowledgments; separate real-store tests own those gates",
    "visible no-op equality does not establish internal event, lease or native version row counts",
    "claimed Issue append preserves the observed lease; no heartbeat, renewal, reclaim, migration or recovery support is implied",
    "ordinary Dolt 2.1.8 database provisioning remains serialized; no automatic replay after an uncertain result",
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
                   all(row == wanted[row["id"]] for row in rows), label + ": append changed readiness or complete ready records")

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

    args = ["init", "--prefix", "inotes", "--non-interactive", "--skip-hooks", "--skip-agents", "--json",
            "--graph-mode", "link", "--scope-url", SCOPE]
    if capture.args.server_port:
        args += ["--server", "--external", "--server-host", "127.0.0.1", "--server-port", str(capture.args.server_port),
                 "--database", "inotes_" + capture.root.name.replace("-", "_"), "--server-user", "root"]
    initialized = c0.envelope(capture.success("normal-init", args))
    c0.require(initialized.get("scope") == SCOPE and initialized.get("backend") == backend, "wrong authority/backend")
    caps = command("append-capabilities", ["status", "--graph"])["capabilities"]
    c0.require(caps.get("issueNotesAppend") is True and caps.get("issueNotesReplace") is False and
               caps.get("issueWorkflows") is False and caps.get("historyExact") is False,
               "append capability overstates destination")
    source = save("source-created", command("create-source", ["create", "Notes work — 雪", "--id", "beads/source",
        "--description", "Preserve context", "--labels", "retained,work", "--priority", "3"]))
    target = save("target-created", command("create-target", ["create", "Prerequisite", "--id", "beads/target", "--priority", "1"]))
    memory = save("memory", command("create-memory", ["remember", "Unchanged body — 雪", "--id", "beads/context", "--title", "Context"]))
    added = command("create-dependency", ["dep", "add", source["id"], target["id"], "--actor", "dependency-author"])
    source, dependency = save("source-owned", added["source"]), save("dependency", added["link"])
    related = command("create-context-link", ["link", source["id"], memory["id"], "--resource-type",
        SCOPE + "types/preview-related-v2", "--id", "links/context", "--properties", '{"note":"preserve"}'])
    context = save("context-link", related["link"])
    c0.require(related["source"] == source and source["owned"] == [dependency], "informational Link changed Issue ownership")
    current("authored-baseline", [source, target, memory, dependency, context])
    ready("initial-blocked", [target])
    capture.passed("normal initialization authors an Issue with owned Dependency, prerequisite and unrelated Memory/context Link")

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

    def edit(label, before, text=None, extra=None, fields=None, unconditional=False, human=False, quiet=False):
        # None means omitted. An empty string remains an explicit append intent.
        changes = dict(fields or {})
        argv = ["update", before["id"], "--actor", "notes-author"]
        if text is not None:
            argv += ["--append-notes", text]
            old = before["properties"].get("notes", "")
            combined = old + ("\n" if old else "") + text
            if combined != old:
                changes["notes"] = combined
        argv += extra or []
        argv += ["--unconditional"] if unconditional else ["--if-revision", before["revision"]]
        if quiet:
            argv += ["--quiet"]
        changed = any(before["properties"].get(key) != value for key, value in changes.items())
        if human:
            receipt, out, err = raw(label, argv)
            c0.require(receipt["exit_code"] == 0 and err == b"", label + ": human edit failed")
            expected = b"" if quiet else (("Updated" if changed else "Unchanged") + " " + before["id"] + "\n").encode()
            c0.require(out == expected, label + ": human/quiet output mismatch")
            after = command(label + "-fresh", ["show", before["id"], "--readonly"])
            result = {"issue": after, "changed": changed}
        else:
            result = command(label, argv)
            after = result["issue"]
            current(label + "-fresh", [after])
        c0.require(set(result) == {"issue", "changed"} and result["changed"] is changed, label + ": wrong mutation result")
        if changed:
            after = accepted(label, before, result, changes, "notes-author")
        else:
            c0.require(after == before, label + ": no-op changed complete record")
        exact(label + "-prior", before)
        compare(label + "-compare", before, after)
        return after

    original = source
    source = edit("empty-on-empty-noop", source, "")
    literal = "  First — 雪 café e\u0301\r\nsecond line\t  "
    source = edit("literal-unicode-crlf", source, literal)
    source = edit("literal-dash", source, "-")
    source = edit("duplicate-literal", source, literal)
    source = edit("empty-on-nonempty-newline", source, "")
    c0.require(source["properties"]["notes"] == literal + "\n-\n" + literal + "\n",
               "literal/newline/duplicate append semantics differ")
    capture.passed("empty-on-empty is unchanged; literal Unicode, CRLF, whitespace, dash, duplicates and empty-on-nonempty follow the native newline operation")

    source = edit("mixed-title-priority-append", source, "Mixed edit", extra=["--title", "Renamed notes work", "--priority", "0"],
                  fields={"title": "Renamed notes work", "priority": 0})
    source = edit("omitted-append-preserved", source, extra=["--description", "Description only"],
                  fields={"description": "Description only"})
    source = edit("explicit-unconditional-append", source, "Unconditional append", unconditional=True)
    capture.passed("mixed title/priority/append and an omitted-append scalar edit preserve every other field and complete owned state")

    source = edit("human-append", source, "Human append", human=True)
    source = edit("quiet-human-append", source, "Quiet append", human=True, quiet=True)
    target = edit("quiet-json-empty-noop", target, "", quiet=True)
    capture.passed("human and quiet writes and quiet JSON no-op report the actual accepted complete state")

    before_claim = source
    result = command("claim-before-append", ["update", source["id"], "--claim", "--actor", "lease-holder"])
    props = result["issue"]["properties"]
    c0.require(props.get("status") == "in_progress" and props.get("assignee") == "lease-holder" and
               props.get("started_at") and props.get("heartbeat_at") and props.get("lease_expires_at"),
               "claim did not establish the actual lease-preservation fixture")
    fields = {"status": "in_progress", "assignee": "lease-holder"}
    for key in ("started_at", "heartbeat_at", "lease_expires_at", "lease_granted_node"):
        fields[key] = props.get(key)
    source = accepted("claimed-source", source, result, fields, "lease-holder")
    current("claim-fresh", [source])
    exact("claim-prior", before_claim)
    compare("claim-compare", before_claim, source)
    claimed = copy.deepcopy(source)
    source = edit("append-while-claimed", source, "Claimed handoff")
    for key in ("status", "assignee", "started_at", "heartbeat_at", "lease_expires_at", "lease_granted_node"):
        c0.require(source["properties"].get(key) == claimed["properties"].get(key), "append changed native claim/lease field " + key)
    capture.passed("a real CLI claim followed by notes append preserves assignee, start time and the complete observed lease without renewal")

    def close(label, before):
        reason = label + " — 雪"
        result = command(label, ["close", before["id"], "--reason", reason, "--actor", "closer"])
        props = result["issue"]["properties"]
        c0.require(props.get("closed_at") and props.get("close_reason") == reason, label + ": missing close state")
        fields = {"status": "closed", "closed_at": props["closed_at"], "close_reason": reason}
        for key in ("lease_expires_at", "heartbeat_at", "lease_granted_node"):
            c0.require(not props.get(key), label + ": close retained lease projection")
            fields[key] = None
        after = accepted(label, before, result, fields, "closer")
        current(label + "-fresh", [after])
        exact(label + "-prior", before)
        compare(label + "-compare", before, after)
        return after

    target = close("close-prerequisite", target)
    source = close("close-source", source)
    source = edit("append-closed-issue", source, "Post-close note")
    c0.require(source["properties"]["status"] == "closed", "notes append reopened a closed Issue")
    listing("closed-complete-list", [source, target])
    ready("both-closed", [])
    capture.passed("notes append to a closed Issue retains closure, ownership and empty ready result; explicit complete listing reflects the accepted records")

    common = ["update", source["id"], "--append-notes", "Refused append"]
    for label, argv, code in [
        ("stale-append", [*common, "--if-revision", original["revision"]], "revision_conflict"),
        ("stale-empty-append", ["update", target["id"], "--append-notes", "", "--if-revision", saved[1]["record"]["revision"]], "revision_conflict"),
        ("missing-guard", common, "invalid_selector"),
        ("empty-guard", [*common, "--if-revision", ""], "invalid_selector"),
        ("both-guards", [*common, "--if-revision", source["revision"], "--unconditional"], "invalid_selector"),
        ("false-unconditional", [*common, "--unconditional=false"], "invalid_selector"),
        ("memory-kind", ["update", memory["id"], "--append-notes", "Refused", "--unconditional"], "capability_unavailable"),
        ("link-kind", ["update", context["id"], "--append-notes", "Refused", "--unconditional"], "capability_unavailable"),
        ("missing-issue", ["update", "beads/missing", "--append-notes", "Refused", "--unconditional"], "not_found"),
        ("foreign-selector", ["update", "https://other.invalid/beads/source", "--append-notes", "Refused", "--unconditional"], "invalid_selector"),
        ("legacy-selector", ["update", "inotes-legacy", "--append-notes", "Refused", "--unconditional"], "invalid_selector"),
        ("mixed-properties", [*common, "--unconditional", "--properties", '{}'], "capability_unavailable"),
    ]:
        refuse(label, argv, code)
    for flag, value in [("notes", "Replacement"), ("status", "open"), ("claim", "true"), ("claim", "false"),
                        ("force", "false"), ("if-assignee", ""), ("if-source-revision", source["revision"]),
                        ("body-file", str(capture.root / "missing.md"))]:
        refuse("unsupported-" + flag + ("-" + value if flag == "claim" else ""),
               [*common, "--unconditional", "--" + flag + "=" + value],
               "invalid_properties" if flag == "claim" and value == "false" else "capability_unavailable")

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
                freeze.write_text("append-notes\t2026-09-27T00:00:00Z\twrite policy\n")
            refuse("policy-" + policy, [*common, "--unconditional", *flags], "permission_denied")
        finally:
            if old_readonly is None:
                capture.env.pop("BD_READONLY", None)
            else:
                capture.env["BD_READONLY"] = old_readonly
            if old_yaml is None:
                yaml.unlink(missing_ok=True)
            else:
                yaml.write_bytes(old_yaml)
            freeze.unlink(missing_ok=True)
            if mayor.exists():
                (mayor / "town.json").unlink(missing_ok=True)
                mayor.rmdir()
    current("refusals-preserve-complete-state", [source, target, memory, dependency, context])
    capture.passed("stale/no-op guards, invalid guards, unsupported mixed operations, selectors and policy refusals produce no successful output or complete-record mutation")
    for index, item in enumerate(saved):
        exact("final-saved-" + str(index), item["record"])
    capture.passed("every accepted saved version remains exact in a new process, including notes, complete owned Dependency and claimed lease observation")
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
    if args.backend in {"server", "both"}:
        c0.require(args.server_root is not None, "server/both require --server-root provenance")
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
