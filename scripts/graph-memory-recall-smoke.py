#!/usr/bin/env python3
"""Prove installed graph Memory recall returns exact current and saved body bytes.

Uses normal initialization and public CLI writes in disposable workspaces, with
fresh processes for every read. Embedded and caller-supplied ordinary Dolt run
sequentially. No schema seeding, SQL, fixtures, mocks or HTTP calls are used.
"""

import argparse
import copy
import hashlib
import importlib.util
import json
import os
from pathlib import Path
import secrets
import signal
import time

HERE = Path(__file__).resolve().parent
SPEC = importlib.util.spec_from_file_location("graph_c0_capture", HERE / "graph-c0-smoke.py")
c0 = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(c0)
SCOPE = "https://example.invalid/disposable-memory-recall/"
LIMITATIONS = [
    "body-only graph recall; complete structured Memory JSON remains unavailable",
    "saved exact versions do not establish History enumeration, lineage or restore",
    "legacy non-graph recall is not exercised by this graph-only harness",
    "no concurrent-writer, crash, corruption or HTTP/public-client qualification",
    "current-record equality does not prove unchanged internal database bytes",
    "ordinary Dolt 2.1.8 server provisioning must remain serialized",
]


def exercise(capture):
    backend = "server" if capture.args.server_port else "embedded"
    init = ["init", "--graph-mode", "link", "--scope-url", SCOPE, "--prefix", "recall",
            "--non-interactive", "--skip-hooks", "--skip-agents", "--json"]
    if capture.args.server_port:
        init += ["--server", "--external", "--server-host", "127.0.0.1", "--server-port",
                 str(capture.args.server_port), "--database", "recall_" + capture.root.name.replace("-", "_"),
                 "--server-user", "root"]
    initialized = c0.envelope(capture.success("normal-init", init))
    c0.require(initialized.get("scope") == SCOPE and initialized.get("backend") == backend,
               "normal initialization selected the wrong authority/backend")
    capture.passed("normal disposable graph initialization")
    status = c0.envelope(capture.success("capabilities", ["status", "--graph", "--json"]))
    caps = status.get("capabilities", {})
    c0.require(caps.get("exactVersionRead") is True and caps.get("memory") is False
               and caps.get("historyExact") is False, "status overstated Memory/History or omitted exact reads")
    c0.require(caps.get("memoryBodyRecall") is True and caps.get("memoryJSONRecall") is False,
               "status must distinguish body recall from unavailable structured Memory")
    c0.require(status.get("limits", {}).get("versionTokenBytes") == 4096,
               "status omitted the exact-version token byte limit")
    capture.passed("status distinguishes body recall, structured Memory and full History")
    saved, raw_receipts = [], []

    def save(label, record):
        c0.require(isinstance(record, dict) and record.get("type") == SCOPE + "types/preview-memory-v2",
                   label + ": missing Memory record")
        c0.require(isinstance(record.get("version"), str) and record["version"], label + ": missing version")
        saved.append({"label": label, "record": copy.deepcopy(record)})
        c0.write_json(capture.output / "expected-records.json", saved)
        return record

    def raw_run(label, arguments):
        receipt, _, _ = capture.run(label, arguments)
        artifact = capture.output / capture.records[-1]["artifact"]
        # Capture.run's text return normalizes CRLF. Inspect the original files.
        return receipt, (artifact / "stdout.log").read_bytes(), (artifact / "stderr.log").read_bytes()

    def recall(label, record, flags=None, historical=False, selector=None):
        arguments = ["recall", selector or record["id"]]
        if historical:
            arguments += ["--version", record["version"]]
        arguments += flags or []
        receipt, output, error = raw_run(label, arguments)
        expected = record["properties"]["body"].encode("utf-8")
        raw_receipts.append({"label": label, "command_artifact": capture.records[-1]["artifact"],
                             "expected_bytes": len(expected), "actual_bytes": len(output),
                             "expected_sha256": hashlib.sha256(expected).hexdigest(),
                             "actual_sha256": hashlib.sha256(output).hexdigest(),
                             "exit_code": receipt["exit_code"]})
        c0.write_json(capture.output / "raw-recall-receipts.json", raw_receipts)
        c0.require(receipt["exit_code"] == 0, label + ": recall failed: " + error.decode("utf-8", "replace"))
        c0.require(output == expected, label + ": body bytes changed or a newline/decoration was appended")
        c0.require(error == b"", label + ": successful recall emitted unexpected stderr")

    def refuse(label, arguments, code):
        receipt, output, error = raw_run(label, arguments)
        c0.require(receipt["exit_code"] != 0 and output == b"", label + ": refusal succeeded or leaked body")
        if "--json" in arguments:
            problem = json.loads(error)
            c0.require(problem.get("code") == code and problem.get("retryable") is False,
                       label + ": wrong structured refusal")
        else:
            c0.require(error.startswith((code + ": ").encode()), label + ": wrong plain refusal code")

    original_body = "  leading\t雪 😀 e\u0301\r\nline two\ntrailing  "
    original = save("original", c0.envelope(capture.success("memory-create", c0.remember(
        "beads/plan", original_body, "Recall original"))))
    c0.require(original["properties"]["body"] == original_body, "remember normalized original body")
    recall("initial-local-path", original, selector="beads/plan")
    recall("initial-canonical-url-quiet", original, flags=["--quiet"])
    target = c0.envelope(capture.success("issue-create", ["create", "Recall target", "--id", "beads/work", "--json"]))
    relation = c0.envelope(capture.success("owned-link-create", [
        "link", "beads/plan", "beads/work", "--resource-type", SCOPE + "types/preview-related-v2",
        "--id", "links/context", "--properties", '{"note":"first"}',
        "--if-source-revision", original["revision"], "--json"]))
    link_before = relation["link"]
    linked = save("linked-original", relation["source"])
    c0.require(linked.get("owned") == [link_before], "normal link did not populate Memory ownership")
    recall("after-owned-link-create", linked)
    capture.passed("current recall preserves Unicode, CRLF, whitespace and quiet output exactly")

    def update(label, current, body):
        result = c0.envelope(capture.success(label, [
            "update", "beads/plan", "--properties", json.dumps({"title": "Recall updated", "body": body}, ensure_ascii=False),
            "--if-revision", current["revision"], "--json"]))
        record = save(label, result.get("memory"))
        c0.require(result.get("changed") is True and record["version"] != current["version"],
                   label + ": semantic update did not create a new version")
        c0.require(record["properties"]["body"] == body and record.get("owned") == current.get("owned"),
                   label + ": update changed body or ownership unexpectedly")
        return record

    edited = update("content-edit", linked, "\n\n  second body\r\n終わり\n")
    changed_link = c0.envelope(capture.success("owned-link-edit", [
        "update", "links/context", "--properties", '{"note":"later"}',
        "--if-revision", link_before["revision"], "--if-source-revision", edited["revision"], "--json"]))
    link_after = changed_link["link"]
    linked_edit = save("edited-with-later-link", changed_link["source"])
    c0.require(linked_edit.get("owned") == [link_after] and
               linked_edit["properties"] == edited["properties"], "Link update changed Memory body")
    recall("current-after-link-edit", linked_edit)
    for entry in saved:
        recall(entry["label"] + "-historical", entry["record"], historical=True, flags=["--readonly"])
    capture.passed("saved body versions survive later content and owned-Link writes in fresh processes")

    emptied = update("empty-body-edit", linked_edit, "")
    recall("empty-current", emptied)
    recall("empty-current-quiet", emptied, flags=["--quiet", "--readonly"])
    final = update("final-body-edit", emptied, "final 雪\r\n  ")
    recall("empty-historical-quiet", emptied, historical=True, flags=["--quiet"])
    recall("final-readonly-quiet", final, flags=["--readonly", "--quiet"])
    old_readonly = capture.env.get("BD_READONLY")
    capture.env["BD_READONLY"] = "true"
    try:
        recall("environment-readonly-current", final)
        recall("environment-readonly-historical", original, historical=True)
    finally:
        if old_readonly is None:
            capture.env.pop("BD_READONLY", None)
        else:
            capture.env["BD_READONLY"] = old_readonly
    capture.passed("empty body is zero-byte success; readonly current and historical reads preserve exact bytes")

    unknown = secrets.token_hex(16)
    c0.require(unknown not in {entry["record"]["version"] for entry in saved}, "unknown token collision")
    cases = [
        ("json-not-yet-supported", ["recall", "beads/plan", "--json"], "capability_unavailable"),
        ("json-historical-not-yet-supported", ["recall", "beads/plan", "--version", original["version"], "--json"],
         "capability_unavailable"),
        ("healthy-issue-not-memory", ["recall", "beads/work"], "capability_unavailable"),
        ("healthy-link-not-memory", ["recall", "links/context"], "capability_unavailable"),
        ("unknown-version", ["recall", "beads/plan", "--version", unknown], "revision_unknown"),
        ("cross-subject-version", ["recall", "beads/plan", "--version", target["version"]], "revision_unknown"),
        ("empty-version", ["recall", "beads/plan", "--version", ""], "invalid_selector"),
        ("oversize-version", ["recall", "beads/plan", "--version", "雪" * 1366], "invalid_selector"),
        ("missing-resource", ["recall", "beads/missing"], "not_found"),
        ("foreign-scope", ["recall", "https://other.invalid/foreign/beads/plan"], "invalid_selector"),
    ]
    for label, arguments, code in cases:
        refuse(label, arguments, code)
    capture.passed("unsupported JSON/kinds, unknown versions and invalid selectors refuse without body output")
    for entry in saved:
        record = entry["record"]
        recall(entry["label"] + "-final-historical", record, historical=True, flags=["--readonly"])
        actual = c0.envelope(capture.success(entry["label"] + "-exact-record", [
            "show", record["id"], "--version", record["version"], "--readonly", "--json"]))
        c0.require(actual == record, entry["label"] + ": exact retained content or owned set changed")
    for label, record in [("memory", final), ("link", link_after), ("issue", target)]:
        actual = c0.envelope(capture.success("final-unchanged-" + label, ["show", record["id"], "--readonly", "--json"]))
        c0.require(actual == record, label + ": recall/refusal changed current record")
    capture.passed("all retained bodies and owned sets remain exact; all current records survive reads/refusals")
    return {"saved_versions": len(saved), "raw_recall_checks": len(raw_receipts),
            "expected_records_sha256": c0.sha256(capture.output / "expected-records.json"),
            "raw_recall_receipts_sha256": c0.sha256(capture.output / "raw-recall-receipts.json"),
            "http_exercised": False, "legacy_recall_exercised": False}


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
