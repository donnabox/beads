#!/usr/bin/env python3
"""Installed current generic traversal proof using only normal CLI authoring.

Both engines run sequentially; the caller owns the ordinary Dolt server.
Private summary DTO only: no external, historical, custom-Type or BDP claim.
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
SCOPE = "https://example.invalid/disposable-generic-traversal/"
BODY = "PRIVATE_MEMORY_BODY_7b12 — 雪\r\nDo not expose in summaries."
NOTE = "PRIVATE_LINK_NOTE_39e4"
LONG = "PRIVATE_ISSUE_DESCRIPTION_f248"
LIMITATIONS = [
    "Private current local summary projection; defaults and frontier shape are not a durable CLI or BDP contract",
    "No external/pinned/custom-Type traversal, native History, public HTTP traversal or write profile",
    "Depth frontier is observable; node/Link/output cap exhaustion refuses rather than returning partial results",
    "Record equality proves visible read preservation; storage suites own internal counts, snapshot races and corruption controls",
    "No forced concurrent engine overlap or large 1MiB output fixture in this harness; pure projection tests own output boundary",
    "Ordinary Dolt provisioning is serial; caller owns server lifecycle and retained disposable databases",
]


def digest(value):
    return hashlib.sha256(json.dumps(value, ensure_ascii=False, sort_keys=True,
                                     separators=(",", ":")).encode()).hexdigest()


def exercise(capture):
    backend = "server" if capture.args.server_port else "embedded"
    records, queries, controls = {}, [], []
    visited_versions = set()

    def command(label, argv):
        return c0.envelope(capture.success(label, [*argv, "--json"]))

    def raw(label, argv):
        receipt, _, _ = capture.run(label, argv)
        path = capture.output / capture.records[-1]["artifact"]
        return receipt, (path / "stdout.log").read_bytes(), (path / "stderr.log").read_bytes()

    def refuse(label, argv, code=None, status=None):
        receipt, out, err = raw(label, [*argv, "--json"])
        c0.require(receipt["exit_code"] != 0 and out == b"" and err, label + ": missing refusal or partial output")
        if status is not None:
            c0.require(receipt["exit_code"] == status, label + ": wrong exit")
        if code is not None:
            problem = json.loads(err)
            c0.require(problem.get("code") == code and problem.get("retryable") is False,
                       label + ": wrong classified refusal")

    def init_args(prefix, graph=True):
        args = ["init", "--prefix", prefix, "--non-interactive", "--skip-hooks", "--skip-agents"]
        if graph:
            args += ["--graph-mode", "link", "--scope-url", SCOPE]
        if capture.args.server_port:
            args += ["--server", "--external", "--server-host", "127.0.0.1", "--server-port", str(capture.args.server_port),
                     "--database", prefix + "_" + capture.root.name.replace("-", "_"), "--server-user", "root"]
        return args

    def save(key, record):
        c0.require(record.get("id", "").startswith(SCOPE) and record.get("version") and
                   record["version"] == record.get("revision") and record.get("attribution"), "invalid saved " + key)
        records[key] = copy.deepcopy(record)
        return record

    initialized = command("normal-init", init_args("gwalk"))
    c0.require(initialized.get("scope") == SCOPE and initialized.get("backend") == backend, "incorrect checked authority")
    status = command("traversal-capabilities", ["status", "--graph"])
    c0.require(status["capabilities"]["issueWorkflows"] is False and status["capabilities"]["historyExact"] is False,
               "traversal must not imply full workflows or History")
    c0.require(status["capabilities"].get("genericTraversal") is True, "missing private traversal capability")
    for key, wanted in {"genericTraversalOutputBytes": 1 << 20, "genericTraversalDepth": 1000,
                        "genericTraversalNodes": 1000, "genericTraversalLinks": 1000,
                        "genericTraversalInventoryResources": 1000}.items():
        c0.require(status["limits"].get(key) == wanted, "wrong advertised bound: " + key)
    save("W", command("create-work", ["create", "Work — 雪", "--id", "beads/work", "--description", LONG]))
    save("P", command("create-prerequisite", ["create", "Prerequisite", "--id", "beads/prerequisite"]))
    for key, path, title in [("A", "a", "Plan"), ("B", "b", "Rationale")]:
        save(key, command("create-memory-" + path, ["remember", BODY, "--id", "beads/" + path, "--title", title]))
    added = command("create-blocking-dependency", ["dep", "add", records["W"]["id"], records["P"]["id"], "--actor", "graph-author"])
    save("W", added["source"])
    save("D", added["link"])
    # Independent topology: W -> P; W => A (two Link identities), W -> B,
    # A -> B, B -> A, A -> A. W/A/B form a diamond plus back/self cycles.
    for key, source, target in [("WA1", "W", "A"), ("WA2", "W", "A"), ("WB", "W", "B"),
                                ("AB", "A", "B"), ("BA", "B", "A"), ("AA", "A", "A")]:
        argv = ["link", records[source]["id"], records[target]["id"], "--id", "links/" + key.lower(),
                "--resource-type", SCOPE + "types/preview-related-v2", "--properties", json.dumps({"note": NOTE}),
                "--actor", "graph-author"]
        if source in {"A", "B"}:
            argv += ["--if-source-revision", records[source]["revision"]]
        result = command("create-" + key.lower(), argv)
        save(source, result["source"])
        save(key, result["link"])
    c0.require(records["WA1"]["id"] != records["WA2"]["id"] and records["WA1"]["source"] == records["WA2"]["source"] and
               records["WA1"]["target"] == records["WA2"]["target"], "parallel identity fixture missing")
    c0.write_json(capture.output / "expected-records.json", [
        {"key": key, "record": record, "semantic_sha256": digest(record)} for key, record in sorted(records.items())])

    def workflow(label):
        expected = {"blocked": [{"issue": records["W"], "blockedBy": [records["P"]["id"]]}], "ready": [records["P"]]}
        for verb in ["blocked", "ready"]:
            value = capture.success(label + "-" + verb, [verb, "--readonly", "--json"])
            c0.require(value.get("preview") is True and value.get("schemaVersion") == 1 and value.get("result") == expected[verb],
                       label + ": informational traversal fixture changed native workflow")

    workflow("before-traversal")
    capture.passed("normal init and CLI writes create independently identified mixed graph, diamond, parallel Links, self/back cycles and one native blocker")

    def projected(key, node):
        fields = ["id", "type", "version", "attribution"] + (["source", "target"] if not node else [])
        result = {field: records[key][field] for field in fields}
        if node:
            result["title"] = records[key]["properties"]["title"]
        return result

    def traversal(label, root, direction, depth, nodes, links, frontier, extra=(), defaults=False, *, max_nodes=100, max_links=200):
        argv = ["graph", "--view", "generic", records[root]["id"], "--readonly", *extra]
        if not defaults:
            argv += ["--direction", direction, "--depth", str(depth)]
        if not defaults or (max_nodes, max_links) != (100, 200):
            argv += ["--max-nodes", str(max_nodes), "--max-links", str(max_links)]
        wanted = {"projection": "summary", "scope": SCOPE, "root": records[root]["id"], "direction": direction,
                  "depth": depth, "maxNodes": max_nodes, "maxLinks": max_links,
                  "nodes": sorted([projected(k, True) for k in nodes], key=lambda x: x["id"]),
                  "links": sorted([projected(k, False) for k in links], key=lambda x: x["id"]),
                  "frontier": sorted(records[k]["id"] for k in frontier), "complete": not frontier}
        actual = command(label, argv)
        c0.require(actual == wanted, label + ": topology, summary fields, order, frontier or completeness differs")
        for sentinel in [BODY, NOTE, LONG]:
            c0.require(sentinel not in json.dumps(actual, ensure_ascii=False), label + ": private payload leaked")
        for key in [*nodes, *links]:
            visited_versions.add((records[key]["id"], records[key]["version"]))
        queries.append({"label": label, "argv": [*argv, "--json"], "expected": wanted, "semantic_sha256": digest(wanted)})
        c0.write_json(capture.output / "expected-traversals.json", queries)

    all_nodes, all_links = ["W", "P", "A", "B"], ["D", "WA1", "WA2", "WB", "AB", "BA", "AA"]
    # Tables are explicit expectations, not a second implementation of BFS.
    for direction in ["out", "both"]:
        traversal(direction + "-depth0", "W", direction, 0, ["W"], [], ["W"])
        traversal(direction + "-depth1", "W", direction, 1, all_nodes, ["D", "WA1", "WA2", "WB"], ["A", "B"])
        traversal(direction + "-depth2", "W", direction, 2, all_nodes, all_links, [])
    for depth in [0, 1, 2]:
        traversal("in-work-depth" + str(depth), "W", "in", depth, ["W"], [], [])
    traversal("in-memory-depth0", "A", "in", 0, ["A"], [], ["A"])
    traversal("in-memory-depth1", "A", "in", 1, ["A", "W", "B"], ["WA1", "WA2", "BA", "AA"], ["B"])
    traversal("in-memory-depth2", "A", "in", 2, ["A", "W", "B"], ["WA1", "WA2", "WB", "AB", "BA", "AA"], [])
    traversal("default-both-depth1", "W", "both", 1, all_nodes, ["D", "WA1", "WA2", "WB"], ["A", "B"], defaults=True)
    traversal("quiet-json-complete", "W", "out", 2, all_nodes, all_links, [], extra=["--quiet"])
    traversal("exact-caps", "W", "both", 1, all_nodes, ["D", "WA1", "WA2", "WB"], ["A", "B"],
              max_nodes=4, max_links=4)
    capture.passed("explicit topology tables prove direction/depth/defaults, unique nodes and Link identities, boundary frontier and complete cycles with stable summary-only output")

    for key, record in sorted(records.items()):
        c0.require(command("exact-returned-" + key, ["show", record["id"], "--version", record["version"], "--readonly"]) == record,
                   "returned summary did not resolve to saved exact full record: " + key)
    c0.require(visited_versions == {(r["id"], r["version"]) for r in records.values()}, "not every returned version was replayed")
    # Python JSON quoting is only an exact oracle for this deliberately narrow
    # fixture alphabet; it is not a general implementation of Go's %q.
    for key in all_nodes:
        title = records[key]["properties"]["title"]
        c0.require(all(0x20 <= ord(ch) <= 0x7e or ch in "—雪" for ch in title),
                   "human title fixture exceeds the shared quote alphabet")
    for row in records.values():
        for field in ["id", *(["source", "target"] if "source" in row else [])]:
            c0.require(all(0x21 <= ord(ch) <= 0x7e for ch in row[field]),
                       "canonical identity fixture must use printable ASCII")

    def human_expected(depth, link_keys, frontier):
        quote = lambda value: json.dumps(value, ensure_ascii=False)
        lines = ["Generic graph (summary; graph preview)",
                 "Root " + quote(records["W"]["id"]) + " direction both depth " + str(depth)]
        for key in sorted(all_nodes, key=lambda k: records[k]["id"]):
            lines.append("Bead " + quote(records[key]["id"]) + " " + quote(records[key]["properties"]["title"]))
        for key in sorted(link_keys, key=lambda k: records[k]["id"]):
            row = records[key]
            lines.append("Link " + quote(row["id"]) + " " + quote(row["source"]) + " -> " + quote(row["target"]))
        lines.append("Complete: " + ("false" if frontier else "true") + "; frontier: " + str(len(frontier)))
        for key in sorted(frontier, key=lambda k: records[k]["id"]):
            lines.append("Frontier " + quote(records[key]["id"]))
        return ("\n".join(lines) + "\n").encode()

    for label, depth, link_keys, frontier in [("human-traversal", 2, all_links, []),
                                               ("human-frontier", 1, ["D", "WA1", "WA2", "WB"], ["A", "B"])]:
        receipt, out, err = raw(label, ["graph", "--view", "generic", "beads/work", "--depth", str(depth), "--readonly"])
        c0.require(receipt["exit_code"] == 0 and not err and out == human_expected(depth, link_keys, frontier),
                   label + ": human identity/summary/frontier bytes differ")
        c0.require(all(value.encode() not in out for value in [BODY, NOTE, LONG]), label + ": human payload leak")
    receipt, out, err = raw("quiet-human", ["graph", "--view", "generic", "beads/work", "--quiet"])
    c0.require(receipt["exit_code"] == 0 and out == b"" and err == b"", "quiet human emitted output")
    capture.passed("every returned saved version resolves exactly in a fresh process; human and quiet views preserve payload boundaries")

    base = ["graph", "--view", "generic", "beads/work"]
    for label, args, code, status in [
        ("node-cap", [*base, "--max-nodes=1"], "capability_unavailable", 5),
        ("link-cap", [*base, "--max-links=1"], "capability_unavailable", 5),
        ("negative-depth", [*base, "--depth=-1"], "invalid_selector", 2),
        ("excess-depth", [*base, "--depth=1001"], "invalid_selector", 2),
        ("zero-nodes", [*base, "--max-nodes=0"], "invalid_selector", 2),
        ("excess-links", [*base, "--max-links=1001"], "invalid_selector", 2),
        ("bad-direction", [*base, "--direction=sideways"], "invalid_selector", 2),
        ("missing-root", ["graph", "--view", "generic"], "invalid_selector", 2),
        ("unknown-root", ["graph", "--view", "generic", "beads/missing"], "not_found", 3),
        ("link-root", ["graph", "--view", "generic", "links/wa1"], "invalid_selector", 2),
        ("foreign-root", ["graph", "--view", "generic", "https://foreign.invalid/beads/a"], "invalid_selector", 2),
        ("alias-root", ["graph", "--view", "generic", "alias/plan"], "invalid_selector", 2),
        ("legacy-all-false", [*base, "--all=false"], "capability_unavailable", 5),
        ("legacy-dot-false", [*base, "--dot=false"], "capability_unavailable", 5),
        ("omitted-view", ["graph", "beads/work"], "capability_unavailable", 5),
    ]:
        refuse(label, args, code, status)
    refuse("excess-arguments", [*base, "beads/a"])
    for value, code, status in [("1", "capability_unavailable", 5), ("bad", "invalid_properties", 2)]:
        capture.env["BEADS_MAX_ROWS"] = value
        controls.append({"label": "rowcap-" + value, "environment": {"BEADS_MAX_ROWS": value}})
        try:
            refuse("rowcap-" + value, base, code, status)
        finally:
            capture.env.pop("BEADS_MAX_ROWS", None)
    c0.write_json(capture.output / "admission-controls.json", controls)
    capture.passed("node/Link caps, invalid parameters, unsupported roots and changed ordinary flags refuse atomically with no partial stdout")

    metadata = capture.work / ".beads" / "metadata.json"
    original = metadata.read_bytes()
    wrong = json.loads(original)
    wrong["graph_workspace"] = str(capture.root)
    control = capture.output / "wrong-authority-input"
    control.mkdir()
    (control / "before.json").write_bytes(original)
    c0.write_json(control / "wrong.json", wrong)
    metadata.write_bytes((control / "wrong.json").read_bytes())
    try:
        refuse("wrong-workspace-authority", base, "not_authority", 5)
        c0.require(metadata.read_bytes() == (control / "wrong.json").read_bytes(), "authority refusal rewrote config")
    finally:
        metadata.write_bytes(original)
    c0.write_json(control / "hashes.json", {"before_sha256": c0.sha256(control / "before.json"),
                  "wrong_sha256": c0.sha256(control / "wrong.json"), "restored_sha256": c0.sha256(metadata)})
    for key, record in sorted(records.items()):
        c0.require(command("current-after-refusals-" + key, ["show", record["id"], "--readonly"]) == record,
                   "read/refusal changed complete record " + key)
    workflow("after-traversal")
    capture.passed("authority mismatch refuses; restored current records and native blocked/ready views remain exactly unchanged")

    graph_work = capture.work
    capture.work = capture.root / "legacy"
    capture.work.mkdir()
    try:
        receipt, _, _ = raw("legacy-normal-init", [*init_args("gwalklegacy", graph=False), "--json"])
        c0.require(receipt["exit_code"] == 0, "normal legacy initialization failed")
        before = c0.tree_digest(capture.work)
        for label, flags in [("view", ["--view", "generic"]), ("depth", ["--depth=0"]), ("direction", ["--direction=both"])]:
            refuse("legacy-rejects-" + label, ["graph", "beads/work", *flags], "capability_unavailable", 5)
        after = c0.tree_digest(capture.work)
        c0.require(after == before, "legacy generic refusal changed workspace")
        c0.write_json(capture.output / "legacy-tree.json", {"before": before, "after": after})
    finally:
        capture.work = graph_work
    capture.passed("normal legacy workspace rejects generic syntax without changing its workspace")
    return {"saved_versions": len(records), "distinct_exact_versions": len(visited_versions),
            "traversal_count": len(queries), "http_exercised": False,
            "expected_records_sha256": c0.sha256(capture.output / "expected-records.json"),
            "expected_traversals_sha256": c0.sha256(capture.output / "expected-traversals.json"),
            "admission_controls_sha256": c0.sha256(capture.output / "admission-controls.json")}


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
