#!/usr/bin/env python3
"""Installed canonical Issue label replacement and exact-version workflow proof.

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
SCOPE = "https://example.invalid/disposable-issue-labels/"
LIMITATIONS = [
    "experimental whole-label replacement through existing update; add/remove operations remain unavailable in graph mode",
    "CLI trims/deduplicates labels without the legacy whitespace warning; storage preserves literal whitespace; case/accent distinctions are verified only on pinned Dolt 2.1.8 binary collation utf8mb4_0900_bin; returned label ordering is inherited, not newly specified",
    "one accepted graph revision per command is visible; internal version/event counts, forced overlap and rollback need separate real-store tests",
    "ordinary RowVersion coverage and whole-table Dolt staging retain their existing review gates; this CLI capture does not settle them",
    "exact retained reads/comparisons are not public ordered History, full Memory, or HTTP writes",
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
                   all(row == wanted[row["id"]] for row in rows), label + ": label replacement changed readiness or complete ready records")

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

    args = ["init", "--prefix", "ilabels", "--non-interactive", "--skip-hooks", "--skip-agents", "--json",
            "--graph-mode", "link", "--scope-url", SCOPE]
    if capture.args.server_port:
        args += ["--server", "--external", "--server-host", "127.0.0.1", "--server-port", str(capture.args.server_port),
                 "--database", "ilabels_" + capture.root.name.replace("-", "_"), "--server-user", "root"]
    initialized = c0.envelope(capture.success("normal-init", args))
    c0.require(initialized.get("scope") == SCOPE and initialized.get("backend") == backend, "wrong authority/backend")
    caps = command("label-capabilities", ["status", "--graph"])["capabilities"]
    c0.require(caps.get("issueLabelReplacement") is True and caps.get("issueTextUpdate") is True and
               caps.get("issuePriorityUpdate") is True and caps.get("issueList") is True and
               caps.get("issueWorkflows") is False and caps.get("memory") is False and caps.get("historyExact") is False,
               "label capability overstates destination")
    source = save("source-created", command("create-source", ["create", "Label work", "--id", "beads/source",
        "--description", "Preserve description — 雪", "--labels", "old,retained", "--type", "bug", "--priority", "3"]))
    target = save("target", command("create-target", ["create", "Prerequisite", "--id", "beads/target",
        "--priority", "1", "--labels", "target-only"]))
    memory = save("memory", command("create-memory", ["remember", "Unchanged context — 雪", "--id", "beads/context", "--title", "Context"]))
    relation = command("create-dependency", ["dep", "add", source["id"], target["id"], "--actor", "dependency-author"])
    source, dependency = save("source-owned", relation["source"]), save("dependency", relation["link"])
    relation = command("create-context-link", ["link", source["id"], memory["id"], "--resource-type",
        SCOPE + "types/preview-related-v2", "--id", "links/context", "--properties", '{"note":"preserve context"}'])
    context = save("context-link", relation["link"])
    c0.require(relation["source"] == source and source["owned"] == [dependency], "informational Link changed Issue ownership")
    unrelated = [target, memory, dependency, context]
    current("authored-baseline", [source, *unrelated])
    ready("blocked-before-labels", [target])
    listing("initial-label-filter", [source], ["--label", "old,retained"])
    capture.passed("normal initialization and CLI-only mixed-graph authoring yield a blocked Issue with complete owned Dependency")

    def checked(label, before, after, expected_labels, actor, changed, fields=None):
        actual_labels = after["properties"].get("labels", [])
        c0.require(isinstance(actual_labels, list) and len(actual_labels) == len(set(actual_labels)) and
                   set(actual_labels) == set(expected_labels), label + ": label normalization lost case/accent, retained duplicates or wrong membership")
        if not changed:
            c0.require(after == before, label + ": no-op changed complete record/version/attribution")
            return before
        wanted = copy.deepcopy(before)
        if actual_labels:
            wanted["properties"]["labels"] = actual_labels
        else:
            wanted["properties"].pop("labels", None)
        wanted["properties"].update(fields or {})
        wanted["properties"]["updated_at"] = after["properties"]["updated_at"]
        for key in ["version", "revision", "attribution"]:
            wanted[key] = after[key]
        c0.require(after == wanted and after["version"] != before["version"] and after["version"] == after["revision"] and
                   after["attribution"]["actor"] == actor, label + ": mixed edit changed unrelated fields/owned state or lost one result revision")
        return save(label, after)

    def replace(label, before, flags, expected_labels, changed=True, fields=None, unconditional=False, quiet=False):
        guard = ["--unconditional"] if unconditional else ["--if-revision", before["revision"]]
        actor = label + "-actor"
        argv = ["update", before["id"], *flags, *guard, "--actor", actor]
        if quiet:
            argv += ["--quiet"]
        result = command(label, argv)
        c0.require(set(result) == {"issue", "changed"} and result["changed"] is changed,
                   label + ": wrong mutation envelope or invented Memory replacement disclosure")
        after = checked(label, before, result["issue"], expected_labels, actor, changed, fields)
        current(label + "-fresh", [after, *unrelated])
        exact(label + "-prior", before)
        compare(label + "-compare", before, after)
        ready(label + "-ready", [target])
        return after

    prior = source
    labels = ["A", "a", "e", "é", "β"]
    source = replace("normalized-label-replacement", source, ["--set-labels", " A ,a, é ,e,β,,A,é "], labels)
    listing("normalized-label-and", [source], ["--label", "A,a,é"])
    listing("label-any-target-and-source", [target, source], ["--label-any", "target-only,β"])
    listing("old-label-removed", [], ["--label", "old"])
    source = replace("reordered-duplicate-noop", source, ["--set-labels", "β,é,e,a,A,β"], labels, changed=False)
    refuse("stale-equal-labels", ["update", source["id"], "--set-labels", "A,a,e,é,β", "--if-revision", prior["revision"]], "revision_conflict")
    refuse("stale-different-labels", ["update", source["id"], "--set-labels", "different", "--if-revision", prior["revision"]], "revision_conflict")
    current("stale-refusals-preserve", [source, *unrelated])
    capture.passed("guarded replacement normalizes whitespace/duplicates without merging case/accent; no-op preserves the complete record and stale equal requests still refuse")

    source = replace("clear-labels", source, ["--set-labels="], [])
    listing("cleared-label-disappears", [], ["--label-any", "A,a,é,β"])
    source = replace("clear-already-empty-noop", source, ["--set-labels="], [], changed=False)
    source = replace("mixed-label-priority-title", source,
        ["--set-labels", "shipping,é", "--priority", "P0", "--title", "Selected labels — 雪"], ["shipping", "é"],
        fields={"priority": 0, "title": "Selected labels — 雪"})
    listing("mixed-edit-filter", [source], ["--label", "shipping,é", "--priority", "0", "--title", "Selected labels"])
    listing("mixed-edit-order", [source, target])
    source = replace("unconditional-label-replacement", source, ["--set-labels", " shipping ,é,e,shipping"],
                     ["shipping", "é", "e"], unconditional=True)
    source = replace("quiet-json-label-noop", source, ["--set-labels", "e,shipping,é"],
                     ["shipping", "é", "e"], changed=False, quiet=True)
    capture.passed("explicit clear and mixed title/priority/labels return one complete accepted revision; exact previous records and comparison preserve ownership")

    for label, flags, wanted_labels, quiet in [
        ("human-label-change", ["--set-labels", "human,é"], ["human", "é"], False),
        ("quiet-label-noop", ["--set-labels", "é,human,human"], ["human", "é"], True),
    ]:
        before = source
        actor = label + "-actor"
        argv = ["update", source["id"], *flags, "--if-revision", source["revision"], "--actor", actor]
        if quiet:
            argv += ["--quiet"]
        receipt, out, err = raw(label, argv)
        c0.require(receipt["exit_code"] == 0 and err == b"", label + ": human update failed")
        after = command(label + "-show", ["show", source["id"], "--readonly"])
        changed = set(before["properties"].get("labels", [])) != set(wanted_labels)
        source = checked(label, before, after, wanted_labels, actor, changed)
        c0.require(out == (b"" if quiet else (("Updated" if changed else "Unchanged") + " " + source["id"] + "\n").encode()),
                   label + ": wrong human/quiet output")
        exact(label + "-prior", before)
        compare(label + "-compare", before, source)
    current("human-preserves-context", [source, *unrelated])
    capture.passed("human and quiet label edits preserve the same current/exact record semantics as JSON")

    common = ["update", source["id"], "--set-labels", "refused"]
    cases = [
        ("missing-guard", common, "invalid_selector"),
        ("empty-guard", [*common, "--if-revision", ""], "invalid_selector"),
        ("both-guards", [*common, "--if-revision", source["revision"], "--unconditional"], "invalid_selector"),
        ("false-unconditional", [*common, "--unconditional=false"], "invalid_selector"),
        ("overlength-label", ["update", source["id"], "--set-labels", "x" * 256, "--unconditional"], "invalid_properties"),
        ("memory-kind", ["update", memory["id"], "--set-labels", "x", "--unconditional"], "capability_unavailable"),
        ("mixed-add-label", [*common, "--add-label", "extra", "--unconditional"], "capability_unavailable"),
        ("mixed-remove-label", [*common, "--remove-label", "human", "--unconditional"], "capability_unavailable"),
        ("add-label-only", ["update", source["id"], "--add-label", "extra", "--unconditional"], "capability_unavailable"),
        ("remove-label-only", ["update", source["id"], "--remove-label", "human", "--unconditional"], "capability_unavailable"),
        ("readonly-refusal", [*common, "--unconditional", "--readonly"], "permission_denied"),
    ]
    for label, argv, code in cases:
        refuse(label, argv, code)
    current("final-current", [source, *unrelated])
    ready("final-readiness", [target])
    listing("final-label-selection", [source], ["--label", "human,é"])
    for i, item in enumerate(saved):
        exact("final-retained-" + str(i), item["record"])
    capture.passed("invalid labels/guards, unsupported deltas/kinds and readonly writes refuse without altering graph records; all saved versions stay exact")
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
