#!/usr/bin/env python3
"""Installed dependency-blocked inspection proof; normal CLI-only authoring.

Both backends run sequentially. The caller owns the ordinary Dolt server.
No SQL, schema fixtures, hidden bootstrap, or concurrent-provisioning proof.
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
SCOPE = "https://example.invalid/disposable-issue-blocked/"
LIMITATIONS = [
    "dependency-blocked is native query membership, not the complement of ready or every manually blocked/deferred/leased Issue",
    "complete inherited Issue properties retain native properties.id; only wrapper blockedBy and human identity presentation are canonical",
    "read equality proves complete visible record preservation, not internal event/version counts; storage tests own read immutability, budget, corruption and forced overlap",
    "current and exact saved CLI records do not establish full public History, HTTP writes, authentication, or complete Issue workflows",
    "ordinary Dolt 2.1.8 database provisioning is sequential; no automatic replay or server lifecycle management",
    "owned child processes are terminated; disposable workspace and caller-owned server database are retained for evidence and caller cleanup",
]


def semantic_digest(value):
    return hashlib.sha256(json.dumps(value, ensure_ascii=False, sort_keys=True,
                                     separators=(",", ":")).encode()).hexdigest()


def exercise(capture):
    backend = "server" if capture.args.server_port else "embedded"
    saved, queries = [], []
    exact_count = 0
    exact_versions = set()

    def command(label, argv):
        return c0.envelope(capture.success(label, [*argv, "--json"]))

    def raw(label, argv):
        receipt, _, _ = capture.run(label, argv)
        folder = capture.output / capture.records[-1]["artifact"]
        return receipt, (folder / "stdout.log").read_bytes(), (folder / "stderr.log").read_bytes()

    def save(label, record):
        c0.require(isinstance(record, dict) and record.get("id", "").startswith(SCOPE) and
                   record.get("version") == record.get("revision") and record.get("version") and
                   record.get("type") and record.get("attribution"), label + ": incomplete record")
        saved.append({"label": label, "record": copy.deepcopy(record), "semantic_sha256": semantic_digest(record)})
        c0.write_json(capture.output / "expected-records.json", saved)
        return record

    def current(label, records):
        for index, record in enumerate(records):
            c0.require(command(label + "-" + str(index), ["show", record["id"], "--readonly"]) == record,
                       label + ": fresh complete current record changed")

    def exact(label, record):
        nonlocal exact_count
        c0.require(command(label, ["show", record["id"], "--version", record["version"], "--readonly"]) == record,
                   label + ": complete saved version changed")
        exact_count += 1
        exact_versions.add((record["id"], record["version"]))

    def record_query(label, argv, expected):
        queries.append({"label": label, "argv": argv, "expected": copy.deepcopy(expected),
                        "semantic_sha256": semantic_digest(expected)})
        c0.write_json(capture.output / "expected-queries.json", queries)
        return expected

    def array_result(value):
        c0.require(isinstance(value, dict) and value.get("schemaVersion") == 1 and value.get("preview") is True and
                   isinstance(value.get("result"), list), "missing complete graph array envelope")
        return value["result"]

    def blocked(label, source=None, targets=(), flags=()):
        argv = ["blocked", "--readonly", *flags, "--json"]
        wanted = [] if source is None else [{"issue": source, "blockedBy": sorted(row["id"] for row in targets)}]
        got = array_result(capture.success(label, argv))
        c0.require(got == wanted, label + ": native blocking result or complete canonical record differs")
        record_query(label, argv, wanted)

    def ready(label, records):
        argv = ["ready", "--readonly", "--json"]
        got = array_result(capture.success(label, argv))
        wanted = {row["id"]: row for row in records}
        c0.require(isinstance(got, list) and len(got) == len(wanted) and
                   {row["id"] for row in got} == set(wanted) and all(row == wanted[row["id"]] for row in got),
                   label + ": complete ready records differ")
        record_query(label, argv, got)

    def listing(label, records):
        argv = ["list", "--format", "records-json", "--all", "--readonly"]
        got = c0.envelope(capture.success(label, argv))
        wanted = {row["id"]: row for row in records}
        c0.require(set(got) == {"items", "hasMore"} and got["hasMore"] is False and
                   len(got["items"]) == len(wanted) and {row["id"] for row in got["items"]} == set(wanted) and
                   all(row == wanted[row["id"]] for row in got["items"]), label + ": complete Issue inventory differs")
        record_query(label, argv, got)

    def refusal(label, argv, code, status):
        receipt, out, err = raw(label, [*argv, "--json"])
        c0.require(receipt["exit_code"] == status and out == b"", label + ": wrong exit or partial stdout")
        problem = json.loads(err)
        c0.require(problem.get("code") == code and problem.get("retryable") is False and "result" not in problem,
                   label + ": wrong typed refusal")

    def owned_transition(label, before, after, links, actor):
        expected = copy.deepcopy(before)
        expected["owned"] = sorted(links, key=lambda row: row["id"])
        expected["properties"]["updated_at"] = after["properties"]["updated_at"]
        for key in ["version", "revision", "attribution"]:
            expected[key] = after[key]
        c0.require(after == expected and after["version"] == after["revision"] and after["version"] != before["version"] and
                   after["attribution"]["actor"] == actor, label + ": ownership mutation changed unrelated source state")

    def transition(label, verb, before):
        actor, reason = label + "-actor", "Verified " + label
        result = command(label, [verb, before["id"], "--reason", reason, "--actor", actor])
        c0.require(set(result) == {"issue", "changed"} and result["changed"] is True, label + ": expected actual transition")
        after = result["issue"]
        expected = copy.deepcopy(before)
        properties = expected["properties"]
        properties["status"] = "closed" if verb == "close" else "open"
        if verb == "close":
            c0.require(after["properties"].get("closed_at"), label + ": missing closed time")
            properties["closed_at"] = after["properties"]["closed_at"]
            properties["close_reason"] = reason
        else:
            for key in ["closed_at", "close_reason", "closed_by_session", "defer_until"]:
                properties.pop(key, None)
                c0.require(not after["properties"].get(key), label + ": stale closure state")
        properties["updated_at"] = after["properties"]["updated_at"]
        for key in ["version", "revision", "attribution"]:
            expected[key] = after[key]
        c0.require(after == expected and after["version"] != before["version"] and after["attribution"]["actor"] == actor,
                   label + ": transition changed unrelated properties, ownership or identity")
        exact(label + "-prior", before)
        return save(label, after)

    init = ["init", "--prefix", "iblock", "--non-interactive", "--skip-hooks", "--skip-agents",
            "--graph-mode", "link", "--scope-url", SCOPE]
    if capture.args.server_port:
        init += ["--server", "--external", "--server-host", "127.0.0.1", "--server-port", str(capture.args.server_port),
                 "--database", "iblock_" + capture.root.name.replace("-", "_"), "--server-user", "root"]
    initialized = command("normal-init", init)
    c0.require(initialized.get("scope") == SCOPE and initialized.get("backend") == backend, "wrong graph authority")
    status = command("blocked-capabilities", ["status", "--graph"])
    c0.require(status["capabilities"].get("issueBlocked") is True and status["capabilities"].get("issueWorkflows") is False and
               status["capabilities"].get("historyExact") is False and status["limits"].get("issueBlockedInventoryResources") == 1000 and
               status["limits"].get("issueBlockedOutputBytes") == 16 << 20, "incorrect bounded capability advertisement")
    blocked("empty-initial")
    source = save("source-created", command("create-source", ["create", "Dependent work — 雪", "--id", "beads/work", "--priority", "2"]))
    first = save("first-created", command("create-first", ["create", "First prerequisite", "--id", "beads/first", "--priority", "0"]))
    second = save("second-created", command("create-second", ["create", "Second prerequisite", "--id", "beads/second", "--priority", "1"]))
    memory = save("memory-created", command("create-memory", ["remember", "Context does not block work", "--id", "beads/context", "--title", "Plan"]))
    relation = command("add-context", ["link", source["id"], memory["id"], "--id", "links/context",
                                      "--resource-type", SCOPE + "types/preview-related-v2", "--properties", '{"note":"informational only"}'])
    context = save("informational-link", relation["link"])
    c0.require(relation["source"] == source and source["owned"] == [], "informational Link became an Issue-owned blocker")
    blocked("informational-is-not-blocking")
    ready("ready-before-dependencies", [source, first, second])
    listing("list-before-dependencies", [source, first, second])
    initial_source = copy.deepcopy(source)
    added = command("add-first-dependency", ["dep", "add", source["id"], first["id"], "--actor", "dependency-author"])
    owned_transition("first-dependency-source", source, added["source"], [added["link"]], "dependency-author")
    source, first_link = save("source-one-dependency", added["source"]), save("first-dependency", added["link"])
    one_owned_source = copy.deepcopy(source)
    added = command("add-second-dependency", ["dep", "add", source["id"], second["id"], "--actor", "dependency-author"])
    owned_transition("second-dependency-source", source, added["source"], [first_link, added["link"]], "dependency-author")
    source, second_link = save("source-two-dependencies", added["source"]), save("second-dependency", added["link"])
    for link, target in [(first_link, first), (second_link, second)]:
        c0.require(link["type"] == SCOPE + "types/preview-blocks-v1" and link["source"] == source["id"] and
                   link["target"] == target["id"] and link["properties"] == {}, "wrong complete Dependency fixture")
    c0.require(source["owned"] == sorted([first_link, second_link], key=lambda row: row["id"]), "wrong complete owned dependency set")
    owned_source = copy.deepcopy(source)
    blocked("two-canonical-blockers", source, [first, second])
    ready("ready-excludes-dependent", [first, second])
    listing("list-includes-dependent", [source, first, second])
    capture.passed("normal initialization and informational context leave work ready; two Dependencies produce canonical blockers while list retains complete source")

    expected_human = ('Dependency-blocked Issues (1; graph preview)\n' + json.dumps(source["id"]) +
                      ' P2 ' + json.dumps(source["properties"]["title"], ensure_ascii=False) + '\n' +
                      ''.join('  blocked by ' + json.dumps(row["id"]) + '\n' for row in sorted([first, second], key=lambda row: row["id"]))).encode()
    receipt, out, err = raw("human-two-blockers", ["blocked", "--readonly"])
    c0.require(receipt["exit_code"] == 0 and out == expected_human and err == b"", "human view has wrong canonical IDs/content")
    receipt, out, err = raw("quiet-two-blockers", ["blocked", "--quiet", "--readonly"])
    c0.require(receipt["exit_code"] == 0 and out == b"" and err == b"", "quiet human view emitted output")
    blocked("quiet-json-two-blockers", source, [first, second], ["--quiet"])
    current("reads-preserve-all-records", [source, first, second, memory, context, first_link, second_link])
    exact("initial-source-still-exact", initial_source)
    capture.passed("human, quiet, JSON and readonly views preserve complete source/targets/Memory/Links and the initial saved version")

    first = transition("close-first", "close", first)
    blocked("one-blocker-after-close", source, [second])
    ready("second-still-ready", [second])
    first = transition("reopen-first", "reopen", first)
    blocked("two-after-reopen", source, [first, second])
    ready("both-prerequisites-ready-again", [first, second])
    first = transition("reclose-first", "close", first)
    blocked("one-before-unlink", source, [second])
    current("close-reopen-preserved-records", [source, first, second, memory, context, first_link, second_link])
    capture.passed("close removes one blocker, reopen restores it, and reclose leaves the other blocker without altering source ownership")

    removed = command("unlink-last-active-blocker", ["unlink", second_link["id"], "--if-revision", second_link["revision"],
                      "--if-source-revision", source["revision"], "--actor", "dependency-remover"])
    tombstone = removed.get("link", {})
    c0.require(set(removed) == {"source", "link", "changed"} and removed["changed"] is True and
               tombstone.get("id") == second_link["id"] and tombstone.get("type") == second_link["type"] and
               tombstone.get("state") == "deleted" and tombstone.get("previousVersion") == second_link["version"] and
               tombstone.get("revision") == tombstone.get("version") and tombstone.get("version") != second_link["version"] and
               tombstone.get("attribution", {}).get("actor") == "dependency-remover", "wrong deleted Dependency result")
    after = removed["source"]
    owned_transition("unlink-source", source, after, [first_link], "dependency-remover")
    source = save("source-after-unlink", after)
    save("deleted-dependency", tombstone)
    blocked("no-blockers-after-unlink")
    ready("source-ready-after-unlink", [source, second])
    listing("list-after-resolution", [source, first, second])
    refusal("deleted-link-current-gone", ["show", second_link["id"]], "gone", 3)
    exact("deleted-link-old-version", second_link)
    exact("old-owned-set-after-unlink", owned_source)
    receipt, out, err = raw("human-empty-after-resolution", ["blocked"])
    c0.require(receipt["exit_code"] == 0 and out == b"Dependency-blocked Issues (0; graph preview)\n" and err == b"", "empty human view differs")
    capture.passed("canonical guarded unlink removes the last active blocker, source becomes ready, and deleted Link/old owned set remain exactly readable")

    for label, flags, code, status in [
        ("positional", ["beads/work"], "invalid_selector", 2),
        ("parent", ["--parent=beads/work"], "capability_unavailable", 5),
        ("empty-parent", ["--parent="], "capability_unavailable", 5),
        ("label", ["--label=demo"], "capability_unavailable", 5),
        ("empty-label", ["--label="], "capability_unavailable", 5),
        ("label-any", ["--label-any=demo"], "capability_unavailable", 5),
        ("exclude-label", ["--exclude-label=demo"], "capability_unavailable", 5),
    ]:
        refusal("unsupported-" + label, ["blocked", *flags], code, status)
    controls = []
    for value, code, status in [("1", "capability_unavailable", 5), ("bad", "invalid_properties", 2), ("-1", "invalid_properties", 2)]:
        capture.env["BEADS_MAX_ROWS"] = value
        try:
            controls.append({"label": "rowcap-" + value, "environment": {"BEADS_MAX_ROWS": value}})
            c0.write_json(capture.output / "admission-controls.json", controls)
            refusal("rowcap-" + value, ["blocked"], code, status)
        finally:
            capture.env.pop("BEADS_MAX_ROWS", None)
    capture.env["BEADS_MAX_ROWS"] = "0"
    try:
        controls.append({"label": "zero-rowcap-complete", "environment": {"BEADS_MAX_ROWS": "0"}})
        c0.write_json(capture.output / "admission-controls.json", controls)
        blocked("zero-rowcap-complete")
    finally:
        capture.env.pop("BEADS_MAX_ROWS", None)
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
        refusal("wrong-workspace-authority", ["blocked"], "not_authority", 5)
        c0.require(metadata.read_bytes() == (control / "wrong.json").read_bytes(), "authority refusal rewrote configuration")
    finally:
        metadata.write_bytes(original)
    c0.write_json(control / "hashes.json", {"before_sha256": c0.sha256(control / "before.json"),
                  "wrong_sha256": c0.sha256(control / "wrong.json"), "restored_sha256": c0.sha256(metadata)})
    current("final-current-preserved", [source, first, second, memory, context, first_link])
    for index, record in enumerate([initial_source, one_owned_source, owned_source, source, first, second, memory, context, first_link, second_link]):
        exact("final-exact-" + str(index), record)
    blocked("final-empty")
    ready("final-ready", [source, second])
    live_versions = {(entry["record"]["id"], entry["record"]["version"]) for entry in saved if entry["record"].get("state") != "deleted"}
    c0.require(exact_versions == live_versions, "not every distinct saved live record was replayed exactly")
    capture.passed("unsupported selectors/filters/caps and wrong workspace authority refuse; restored authority and every surviving/current or saved record remain intact")
    return {"saved_versions": len(saved), "saved_live_versions": len(live_versions),
            "saved_tombstones": sum(entry["record"].get("state") == "deleted" for entry in saved),
            "distinct_exact_versions": len(exact_versions), "exact_reads": exact_count, "query_count": len(queries), "http_exercised": False,
            "expected_records_sha256": c0.sha256(capture.output / "expected-records.json"),
            "expected_queries_sha256": c0.sha256(capture.output / "expected-queries.json"),
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
