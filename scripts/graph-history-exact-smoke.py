#!/usr/bin/env python3
"""Capture installed-CLI exact-version reads in fresh disposable graph workspaces.

Normal bd init/create/remember/link/update/unlink/dep/close commands author every
record. No SQL, seeded schema, mock or hidden bootstrap is used. Each command
runs in a new process. Embedded and a caller-owned ordinary Dolt server run
sequentially; reserve that server against unrelated concurrent provisioning
(Dolt 2.1.8 restriction). This proves an exact-version CLI slice, not BDP History.
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
SCOPE = "https://example.invalid/disposable-exact-history/"
LIMITATIONS = [
    "exact saved versions only; no version enumeration, ordering, lineage or full Memory History",
    "no BDP HTTP historicalResolution or public History capability claim",
    "no restore, corruption, erasure, crash, cancellation or uncertain-commit qualification",
    "ordinary shared-server provisioning must be serialized against unrelated database creation",
    "read-only current-state comparisons are not proof of unchanged internal database bytes",
]


def semantic_digest(value):
    raw = json.dumps(value, ensure_ascii=False, sort_keys=True, separators=(",", ":")).encode("utf-8")
    return hashlib.sha256(raw).hexdigest()


def exercise(capture):
    backend = "server" if capture.args.server_port else "embedded"
    init = ["init", "--graph-mode", "link", "--scope-url", SCOPE, "--prefix", "history",
            "--non-interactive", "--skip-hooks", "--skip-agents", "--json"]
    if capture.args.server_port:
        init += ["--server", "--external", "--server-host", "127.0.0.1", "--server-port",
                 str(capture.args.server_port), "--database", "history_" + capture.root.name.replace("-", "_"),
                 "--server-user", "root"]
    initialized = c0.envelope(capture.success("normal-init", init))
    c0.require(initialized.get("scope") == SCOPE and initialized.get("backend") == backend,
               "normal initialization selected the wrong authority/backend")
    c0.require(initialized.get("memoryComplete") is False, "init claimed complete Memory")
    capture.passed("normal graph initialization, without schema fixtures or SQL")

    status = c0.envelope(capture.success("status-capabilities", ["status", "--graph", "--json"]))
    c0.require(status.get("capabilities", {}).get("exactVersionRead") is True,
               "status does not advertise the implemented exact-version CLI slice")
    c0.require(status.get("capabilities", {}).get("historyExact") is False,
               "private exact-version CLI slice must not claim public History")
    c0.require(status.get("limits", {}).get("versionTokenBytes") == 4096,
               "missing explicit version-token operational bound")
    capture.passed("status distinguishes exactVersionRead from complete public History")

    expected = []

    def save(label, value):
        c0.require(isinstance(value, dict), f"{label}: expected a record")
        c0.require(value.get("id", "").startswith(SCOPE), f"{label}: wrong canonical identity")
        for field in ["type", "revision", "version"]:
            c0.require(isinstance(value.get(field), str) and value[field], f"{label}: missing {field}")
        entry = {"label": label, "record": copy.deepcopy(value), "semantic_sha256": semantic_digest(value)}
        expected.append(entry)
        c0.write_json(capture.output / "expected-records.json", expected)
        return value

    def show(label, subject):
        return c0.envelope(capture.success(label, ["show", subject, "--json"]))

    def exact(label, value, readonly=False):
        argv = ["show", value["id"], "--version", value["version"], "--json"]
        if readonly:
            argv.append("--readonly")
        actual = c0.envelope(capture.success(label, argv))
        c0.require(actual == value, f"{label}: saved version differs from the original complete record")
        c0.require(semantic_digest(actual) == semantic_digest(value), f"{label}: semantic digest changed")
        return actual

    plan = save("memory-initial", c0.envelope(capture.success("memory-plan-create", c0.remember(
        "beads/plan", "Remember before edits — 雪 😀 e\u0301\nsecond line", "Deployment plan"))))
    target = save("memory-target", c0.envelope(capture.success("memory-target-create", c0.remember(
        "beads/decision", "A separate subject for token isolation", "Decision"))))
    work = save("issue-initial", c0.envelope(capture.success("issue-work-create", [
        "create", "Deliver the plan", "--id", "beads/work", "--json"])))
    prerequisite = save("prerequisite-initial", c0.envelope(capture.success("issue-prerequisite-create", [
        "create", "Review the plan", "--id", "beads/prerequisite", "--json"])))
    c0.require(plan["type"] == target["type"] == SCOPE + "types/preview-memory-v2",
               "remember did not create non-Issue Memory Beads")
    c0.require(work["type"] == prerequisite["type"] == SCOPE + "types/preview-issue-v2",
               "create did not use the installed Issue adapter")
    c0.require(plan["properties"] == {"title": "Deployment plan", "body": "Remember before edits — 雪 😀 e\u0301\nsecond line"},
               "Memory creation changed supplied properties")
    c0.require(plan.get("owned") == [] and work.get("owned") == [], "fresh owners already have Links")
    for entry in list(expected):
        c0.require(show(entry["label"] + "-new-process", entry["record"]["id"]) == entry["record"],
                   "fresh-process current read changed a newly created record")
    capture.passed("Memory and Issue records survive new processes with exact complete contents")

    related = SCOPE + "types/preview-related-v2"
    added = c0.envelope(capture.success("owned-link-create", [
        "link", "beads/plan", "beads/work", "--resource-type", related,
        "--id", "links/context", "--properties", '{"note":"original context — 雪"}',
        "--if-source-revision", plan["revision"], "--json"]))
    link_before = save("link-before-update", added["link"])
    owner_before = save("memory-with-original-link", added["source"])
    c0.require(link_before["type"] == related and link_before.get("source") == plan["id"] and
               link_before.get("target") == work["id"], "mixed Memory/Issue Link identity changed")
    c0.require(owner_before.get("owned") == [link_before], "Memory did not own the complete original Link")
    c0.require(owner_before["version"] != plan["version"], "owned-Link creation did not version Memory")

    updated = c0.envelope(capture.success("owned-link-update", [
        "update", "links/context", "--properties", '{"note":"new context"}',
        "--if-revision", link_before["revision"],
        "--if-source-revision", owner_before["revision"], "--json"]))
    link_after = save("link-after-update", updated["link"])
    owner_after = save("memory-with-updated-link", updated["source"])
    c0.require(link_after["version"] != link_before["version"] and
               owner_after["version"] != owner_before["version"], "update failed to version Link and owner")
    c0.require(owner_after.get("owned") == [link_after], "new owner does not contain the updated Link")
    exact("initial-memory-after-link-writes", plan)
    exact("old-memory-after-link-update", owner_before)
    exact("old-link-after-link-update", link_before)
    capture.passed("old Memory owns the historical Link contents, not today's updated Link")

    deleted = c0.envelope(capture.success("owned-link-unlink", [
        "unlink", "links/context", "--if-revision", link_after["revision"],
        "--if-source-revision", owner_after["revision"], "--json"]))
    tombstone = deleted.get("link", {})
    c0.require(deleted.get("changed") is True and tombstone.get("state") == "deleted",
               "unlink did not report a durable tombstone")
    c0.require(tombstone.get("id") == link_after["id"] and
               tombstone.get("previousVersion") == link_after["version"], "tombstone lost its previous version")
    c0.require(tombstone.get("version") and tombstone["version"] != link_after["version"],
               "tombstone did not allocate a new private token")
    owner_deleted = save("memory-after-unlink", deleted["source"])
    c0.require(owner_deleted.get("owned") == [], "current Memory retained a deleted owned Link")
    c0.write_json(capture.output / "deletion-receipt.json", deleted)
    c0.refusal(capture.run("current-deleted-link", ["show", "links/context", "--json"]), {"gone"},
               "current deleted Link")
    exact("deleted-link-original-version", link_before)
    exact("deleted-link-last-live-version", link_after)
    exact("old-owner-before-link-deletion", owner_before)
    exact("old-owner-after-update-before-deletion", owner_after)
    c0.refusal(capture.run("tombstone-is-not-live-version", [
        "show", "links/context", "--version", tombstone["version"], "--json"]), {"gone"},
        "private tombstone token")
    capture.passed("deleted Link's saved live versions remain readable; current and tombstone reads refuse")

    dependency = c0.envelope(capture.success("blocking-dependency-create", [
        "dep", "add", "beads/work", "beads/prerequisite", "--json"]))
    dependency_link = save("blocking-dependency", dependency["link"])
    work_owned = save("issue-with-dependency", dependency["source"])
    c0.require(dependency_link["type"] == SCOPE + "types/preview-blocks-v1" and
               dependency_link.get("source") == work["id"] and dependency_link.get("target") == prerequisite["id"],
               "Dependency did not project the expected Issue relationship")
    c0.require(work_owned.get("owned") == [dependency_link], "Issue did not retain its owned Dependency Link")
    exact("old-issue-before-dependency", work)
    prereq_close = c0.envelope(capture.success("prerequisite-close", ["close", "beads/prerequisite", "--json"]))
    prerequisite_closed = save("prerequisite-closed", prereq_close["issue"])
    c0.require(prerequisite_closed.get("properties", {}).get("status") == "closed",
               "prerequisite close did not change status")
    work_close = c0.envelope(capture.success("work-close", ["close", "beads/work", "--json"]))
    work_closed = save("issue-closed-with-dependency", work_close["issue"])
    c0.require(work_closed.get("properties", {}).get("status") == "closed" and
               work_closed.get("owned") == [dependency_link], "Issue close lost its retained owned Link")
    exact("old-prerequisite-before-close", prerequisite)
    exact("old-issue-with-dependency-before-close", work_owned)
    exact("dependency-after-issue-closes", dependency_link)
    capture.passed("saved Issue and Dependency versions preserve old status and owned Link state after close")

    unknown = secrets.token_hex(16)
    while unknown in {entry["record"]["version"] for entry in expected} | {tombstone["version"]}:
        unknown = secrets.token_hex(16)
    for label, subject, token in [
        ("unknown-version", plan["id"], unknown),
        ("cross-subject-version", target["id"], plan["version"]),
        ("cross-kind-version", dependency_link["id"], plan["version"]),
    ]:
        c0.refusal(capture.run(label, ["show", subject, "--version", token, "--json"]),
                   {"revision_unknown"}, label)
    capture.passed("unknown, cross-subject and cross-kind tokens return typed revision_unknown")
    for label, token in [("empty-version", ""), ("oversized-version", "x" * 4097)]:
        c0.refusal(capture.run(label, ["show", plan["id"], "--version", token, "--json"]),
                   {"invalid_selector"}, label)
    capture.passed("invalid version selectors refuse rather than falling back to current state")


    current = [owner_deleted, target, work_closed, prerequisite_closed, dependency_link]
    for value in current:
        c0.require(show("current-before-readonly-replay", value["id"]) == value,
                   "current state disagrees with the latest successful write")
    # Every command here is another process after all later writes/deletion.
    # Repeat the full retained set, including exact current versions, in readonly mode.
    for entry in expected:
        exact(entry["label"] + "-final-readonly-version", entry["record"], readonly=True)
    for value in current:
        c0.require(show("current-after-readonly-replay", value["id"]) == value,
                   "historical reads changed current visible state")
    capture.passed("every captured saved version reopens exactly after all writes; reads preserve visible current state")
    return {"saved_versions": len(expected), "saved_records_sha256": c0.sha256(capture.output / "expected-records.json"),
            "tombstone": tombstone, "current_records_semantic_sha256": semantic_digest(current)}


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
