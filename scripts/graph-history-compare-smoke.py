#!/usr/bin/env python3
"""Capture installed experimental graph comparison against saved CLI records.

Normal init and public writes author every record. Each command uses a fresh
process; embedded and caller-owned ordinary Dolt modes run sequentially.
Explicit from/to tokens supply direction, never inferred time or ancestry.
No SQL seeds, schema fixtures, mocks, HTTP, or internal APIs are used.
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
SCOPE = "https://example.invalid/disposable-history-compare/"
LIMITATIONS = [
    "experimental retained-preview comparison shape, not a settled public output contract",
    "common metadata unavailable; Memory Inception/derivation unavailable",
    "no History enumeration, chronology, ancestry, restore or full Memory claim",
    "no concurrency, corruption, crash, authority-transition or HTTP qualification",
    "current-record equality does not prove unchanged internal database bytes",
    "ordinary Dolt 2.1.8 database provisioning must remain serialized",
]


def semantic_digest(value):
    return hashlib.sha256(json.dumps(value, ensure_ascii=False, sort_keys=True,
                                     separators=(",", ":")).encode()).hexdigest()


def expected_comparison(before, after):
    """Build the observable difference solely from recorded public CLI values."""
    c0.require(before["id"] == after["id"] and before["type"] == after["type"],
               "comparison oracle received unlike subjects")
    identity = {name: before[name] for name in ["id", "type", "source", "target"] if name in before}
    c0.require(identity == {name: after[name] for name in identity}, "immutable identity changed")
    is_bead = "source" not in identity
    memory = before["type"] == SCOPE + "types/preview-memory-v2"
    result = {
        "resource": identity,
        "from": {name: before[name] for name in ["version", "attribution"]},
        "to": {name: after[name] for name in ["version", "attribution"]},
        "compared": ["properties", "owned"] if is_bead else ["properties"],
        "unsupported": ["commonMetadata", "inception", "derivation"] if memory else ["commonMetadata"],
        "changes": [],
    }

    def add_changes(area, key_name, old, new):
        # BDP canonical order uses UTF-16 code units; bytes preserve that order.
        for key in sorted(old.keys() | new.keys(), key=lambda item: item.encode("utf-16-be")):
            left = {"present": key in old}
            right = {"present": key in new}
            if key in old:
                left["value"] = old[key]
            if key in new:
                right["value"] = new[key]
            if left != right:
                result["changes"].append({"area": area, key_name: key, "from": left, "to": right})

    add_changes("properties", "member", before["properties"], after["properties"])
    if is_bead:
        old = {record["id"]: record for record in before["owned"]}
        new = {record["id"]: record for record in after["owned"]}
        c0.require(len(old) == len(before["owned"]) and len(new) == len(after["owned"]),
                   "saved owned set has duplicate IDs")
        add_changes("owned", "id", old, new)
    return result


def exercise(capture):
    backend = "server" if capture.args.server_port else "embedded"
    init = ["init", "--graph-mode", "link", "--scope-url", SCOPE, "--prefix", "compare",
            "--non-interactive", "--skip-hooks", "--skip-agents", "--json"]
    if capture.args.server_port:
        init += ["--server", "--external", "--server-host", "127.0.0.1", "--server-port",
                 str(capture.args.server_port), "--database", "compare_" + capture.root.name.replace("-", "_"),
                 "--server-user", "root"]
    initialized = c0.envelope(capture.success("normal-init", init))
    c0.require(initialized.get("scope") == SCOPE and initialized.get("backend") == backend,
               "normal init selected wrong authority/backend")
    status = c0.envelope(capture.success("status-capabilities", ["status", "--graph", "--json"]))
    c0.require(status.get("capabilities", {}).get("historyExact") is False and
               status.get("capabilities", {}).get("memory") is False,
               "comparison must not advertise complete History or Memory")
    c0.require(status.get("capabilities", {}).get("exactVersionCompare") is True and
               status.get("capabilities", {}).get("exactVersionRead") is True,
               "status omitted qualified exact-read/comparison capabilities")
    capture.passed("normal graph initialization and truthful incomplete History/Memory capabilities")
    saved, comparisons = [], []

    def save(label, record):
        c0.require(isinstance(record, dict) and record.get("id", "").startswith(SCOPE),
                   label + ": missing canonical record")
        for key in ["version", "revision", "type"]:
            c0.require(isinstance(record.get(key), str) and record[key], label + ": missing " + key)
        saved.append({"label": label, "record": copy.deepcopy(record)})
        c0.write_json(capture.output / "expected-records.json", saved)
        return record

    def raw_run(label, argv):
        receipt, _, _ = capture.run(label, argv)
        directory = capture.output / capture.records[-1]["artifact"]
        return receipt, (directory / "stdout.log").read_bytes(), (directory / "stderr.log").read_bytes()

    def compare(label, before, after, mode="json", selector=None):
        argv = ["compare", selector or before["id"], "--from", before["version"],
                "--to", after["version"], "--readonly"]
        if mode in {"json", "quiet-json"}:
            argv.append("--json")
            if mode == "quiet-json":
                argv.append("--quiet")
        elif mode == "quiet":
            argv.append("--quiet")
        receipt, stdout, stderr = raw_run(label, argv)
        c0.require(receipt["exit_code"] == 0 and stderr == b"", label + ": comparison failed")
        expected = expected_comparison(before, after)
        if mode == "quiet":
            c0.require(stdout == b"", label + ": quiet emitted output")
        else:
            actual = json.loads(stdout)
            if mode in {"json", "quiet-json"}:
                actual = c0.envelope(actual)
            c0.require(actual == expected, label + ": result differs from full saved properties/owned records")
            if mode == "human":
                c0.require(b"\n  " in stdout, label + ": human output is not indented JSON")
        comparisons.append({"label": label, "mode": mode, "from": before["version"], "to": after["version"],
                            "expected": expected, "expected_sha256": semantic_digest(expected),
                            "command_artifact": capture.records[-1]["artifact"]})
        c0.write_json(capture.output / "expected-comparisons.json", comparisons)
        return expected

    def refuse(label, argv, code):
        receipt, stdout, stderr = raw_run(label, ["compare", *argv, "--json"])
        c0.require(receipt["exit_code"] != 0 and stdout == b"", label + ": error emitted partial output")
        problem = json.loads(stderr)
        c0.require(problem.get("code") == code and problem.get("retryable") is False,
                   label + ": wrong refusal code")

    memory = save("memory-original", c0.envelope(capture.success("remember-plan", c0.remember(
        "beads/plan", " original 雪\r\nbody 😀  ", "Original title"))))
    issue = save("issue-original", c0.envelope(capture.success("create-work", [
        "create", "Deliver plan", "--id", "beads/work", "--json"])))
    prerequisite = save("prerequisite-original", c0.envelope(capture.success("create-prerequisite", [
        "create", "Approve plan", "--id", "beads/prerequisite", "--json"])))
    updated = c0.envelope(capture.success("memory-content-update", [
        "update", "beads/plan", "--properties", json.dumps({"title": "", "body": "\nnew mémoire e\u0301\n"}),
        "--if-revision", memory["revision"], "--json"]))
    edited = save("memory-edited", updated["memory"])
    body_diff = compare("memory-content-forward", memory, edited, selector="beads/plan")
    c0.require({change.get("member") for change in body_diff["changes"]} == {"title", "body"},
               "Memory edit did not exercise both durable properties")
    compare("memory-content-reverse", edited, memory)
    compare("memory-content-human", memory, edited, mode="human")
    compare("memory-content-quiet", memory, edited, mode="quiet")
    compare("memory-content-quiet-json", memory, edited, mode="quiet-json")
    c0.require(compare("memory-same-token", edited, edited)["changes"] == [], "same version had differences")
    compare("memory-same-token-human", edited, edited, mode="human")
    capture.passed("complete Memory properties, explicit reverse direction and same-token human/JSON output")

    def link(label, path, owner):
        result = c0.envelope(capture.success(label, [
            "link", "beads/plan", "beads/work", "--resource-type", SCOPE + "types/preview-related-v2",
            "--id", path, "--properties", '{"note":"first"}', "--if-source-revision", owner["revision"], "--json"]))
        return save(label + "-link", result["link"]), save(label + "-owner", result["source"])

    first, one = link("first-owned-link", "links/context-a", edited)
    second, two = link("second-owned-link", "links/context-b", one)
    c0.require(first["id"] != second["id"] and first["source"] == second["source"]
               and first["target"] == second["target"], "did not exercise distinct equal-endpoint Links")
    compare("owned-first-added", edited, one)
    added = compare("owned-equal-endpoint-added", one, two)
    c0.require(len(added["changes"]) == 1 and added["changes"][0].get("id") == second["id"],
               "equal-endpoint addition lost distinct Link identity")

    def edit_link(label, record, owner, properties):
        result = c0.envelope(capture.success(label, [
            "update", record["id"], "--properties", json.dumps(properties),
            "--if-revision", record["revision"], "--if-source-revision", owner["revision"], "--json"]))
        return save(label + "-link", result["link"]), save(label + "-owner", result["source"])

    first_changed, owner_changed = edit_link("first-link-note-edit", first, two, {"note": "later 雪"})
    compare("link-properties-edited", first, first_changed)
    compare("owned-link-note-edited", two, owner_changed)
    first_reverted, owner_reverted = edit_link("first-link-note-revert", first_changed, owner_changed, {"note": "first"})
    c0.require(first_reverted["properties"] == first["properties"] and first_reverted["version"] != first["version"],
               "revert did not produce equal properties with distinct saved versions")
    c0.require(compare("link-equal-properties-distinct-version", first, first_reverted)["changes"] == [],
               "Link attribution/version context became a durable property diff")
    version_diff = compare("owned-version-substitution", two, owner_reverted)
    c0.require(len(version_diff["changes"]) == 1 and version_diff["changes"][0].get("id") == first["id"],
               "owned Link version substitution disappeared")

    first_without_note, owner_without_note = edit_link("first-link-note-remove", first_reverted, owner_reverted, {})
    absence = compare("link-property-absence", first_reverted, first_without_note)
    c0.require(absence["changes"] == [{"area": "properties", "member": "note",
               "from": {"present": True, "value": "first"}, "to": {"present": False}}],
               "absent note was conflated with empty or null")
    deleted = c0.envelope(capture.success("unlink-first", [
        "unlink", first_without_note["id"], "--if-revision", first_without_note["revision"],
        "--if-source-revision", owner_without_note["revision"], "--json"]))
    final_memory = save("memory-after-unlink", deleted["source"])
    tombstone = deleted["link"]
    c0.require(final_memory["owned"] == [second], "unlink changed another equal-endpoint Link")
    compare("owned-link-removed", owner_without_note, final_memory)
    compare("deleted-link-retained-comparison", first, first_changed)
    c0.refusal(capture.run("current-deleted-link", ["show", first["id"], "--json"]), {"gone"},
               "current deleted Link")
    refuse("tombstone-is-not-live-version", [
        first["id"], "--from", first["version"], "--to", tombstone["version"]], "gone")
    capture.passed("owned identity/add/update/remove, version-only changes and retained deleted-Link comparisons")

    dependency_result = c0.envelope(capture.success("blocking-dependency-create", [
        "dep", "add", "beads/work", "beads/prerequisite", "--json"]))
    dependency = save("blocking-dependency", dependency_result["link"])
    issue_owned = save("issue-with-dependency", dependency_result["source"])
    compare("issue-owned-dependency-added", issue, issue_owned)
    c0.require(compare("dependency-same-version", dependency, dependency)["changes"] == [],
               "same Dependency version had changes")
    prerequisite_closed = save("prerequisite-closed", c0.envelope(capture.success(
        "close-prerequisite", ["close", "beads/prerequisite", "--json"]))["issue"])
    issue_closed = save("issue-closed", c0.envelope(capture.success(
        "close-work", ["close", "beads/work", "--json"]))["issue"])
    close_diff = compare("issue-close-all-properties", issue_owned, issue_closed)
    c0.require(any(change.get("member") == "status" for change in close_diff["changes"]),
               "Issue close did not exercise status comparison")
    compare("issue-create-to-close-including-owned", issue, issue_closed)
    compare("prerequisite-close", prerequisite, prerequisite_closed)
    capture.passed("complete Issue property projection and owned Dependency records survive comparison")

    unknown = secrets.token_hex(16)
    c0.require(unknown not in {item["record"]["version"] for item in saved}, "unknown token collision")
    invalid_cases = [
        ("missing-resource", ["beads/missing", "--from", memory["version"], "--to", edited["version"]], "not_found"),
        ("unknown-first", ["beads/plan", "--from", unknown, "--to", memory["version"]], "revision_unknown"),
        ("unknown-second", ["beads/plan", "--from", memory["version"], "--to", unknown], "revision_unknown"),
        ("unknown-same", ["beads/plan", "--from", unknown, "--to", unknown], "revision_unknown"),
        ("cross-subject", ["beads/plan", "--from", memory["version"], "--to", issue["version"]], "revision_unknown"),
        ("missing-from", ["beads/plan", "--to", memory["version"]], "invalid_selector"),
        ("missing-to", ["beads/plan", "--from", memory["version"]], "invalid_selector"),
        ("empty-from", ["beads/plan", "--from", "", "--to", memory["version"]], "invalid_selector"),
        ("empty-to", ["beads/plan", "--from", memory["version"], "--to", ""], "invalid_selector"),
        ("oversize-from", ["beads/plan", "--from", "雪" * 1366, "--to", memory["version"]], "invalid_selector"),
        ("oversize-to", ["beads/plan", "--from", memory["version"], "--to", "x" * 4097], "invalid_selector"),
        ("foreign-selector", ["https://other.invalid/beads/plan", "--from", memory["version"],
                              "--to", edited["version"]], "invalid_selector"),
    ]
    for label, arguments, code in invalid_cases:
        refuse(label, arguments, code)
    receipt, output, error = raw_run("human-unknown-second", [
        "compare", "beads/plan", "--from", memory["version"], "--to", unknown])
    c0.require(receipt["exit_code"] != 0 and output == b"" and error.startswith(b"revision_unknown: "),
               "human error emitted partial output or lost its refusal code")
    capture.passed("unavailable operands and invalid selectors fail atomically with typed diagnostics")
    old_readonly = capture.env.get("BD_READONLY")
    capture.env["BD_READONLY"] = "true"
    try:
        compare("environment-readonly-comparison", memory, final_memory)
    finally:
        if old_readonly is None:
            capture.env.pop("BD_READONLY", None)
        else:
            capture.env["BD_READONLY"] = old_readonly
    for label, record in [("memory", final_memory), ("live-link", second), ("issue", issue_closed),
                          ("prerequisite", prerequisite_closed), ("dependency", dependency)]:
        actual = c0.envelope(capture.success("unchanged-current-" + label, [
            "show", record["id"], "--readonly", "--json"]))
        c0.require(actual == record, label + ": current record changed during comparison")
    for item in saved:
        record = item["record"]
        actual = c0.envelope(capture.success("unchanged-exact-" + item["label"], [
            "show", record["id"], "--version", record["version"], "--readonly", "--json"]))
        c0.require(actual == record, item["label"] + ": retained record changed during comparison")
    capture.passed("readonly comparisons and refusals preserve every current and saved exact record")
    return {"saved_versions": len(saved), "comparison_checks": len(comparisons),
            "expected_records_sha256": c0.sha256(capture.output / "expected-records.json"),
            "expected_comparisons_sha256": c0.sha256(capture.output / "expected-comparisons.json"),
            "public_history_exercised": False, "http_exercised": False}


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
