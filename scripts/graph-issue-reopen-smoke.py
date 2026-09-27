#!/usr/bin/env python3
"""Installed graph Issue close/reopen/readiness and retained-state proof.

Normal authoring only, fresh CLI processes and serialized embedded/server modes.
The caller owns any ordinary Dolt server; no SQL, fixtures or schema bootstrap.
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
SCOPE = "https://example.invalid/disposable-issue-reopen/"
LIMITATIONS = [
    "single explicit canonical durable Issue reopen only; no complete Issue workflow claim",
    "reopen reason is existing event/audit text, not an Issue property or exposed graph history feed",
    "exact saved versions are not native/public History or native commit-time attribution",
    "no forced concurrency, fault injection, event-count, corruption or uncertain-commit qualification here",
    "visible equality does not establish internal row/mapping/event counts",
    "ordinary Dolt 2.1.8 provisioning must remain serialized",
]

def semantic_digest(value):
    return hashlib.sha256(json.dumps(value, ensure_ascii=False, sort_keys=True,
                                     separators=(",", ":")).encode()).hexdigest()

def exercise(capture):
    backend = "server" if capture.args.server_port else "embedded"
    saved = []

    def init_args(graph, prefix):
        args = ["init", "--prefix", prefix, "--non-interactive", "--skip-hooks", "--skip-agents", "--json"]
        if graph:
            args += ["--graph-mode", "link", "--scope-url", SCOPE]
        if capture.args.server_port:
            args += ["--server", "--external", "--server-host", "127.0.0.1", "--server-port",
                     str(capture.args.server_port), "--database", prefix + "_" + capture.root.name.replace("-", "_"),
                     "--server-user", "root"]
        return args

    def command(label, args):
        return c0.envelope(capture.success(label, [*args, "--json"]))

    initialized = c0.envelope(capture.success("normal-graph-init", init_args(True, "reopen")))
    c0.require(initialized.get("scope") == SCOPE and initialized.get("backend") == backend, "wrong graph initialization")
    status = command("status", ["status", "--graph"])
    caps = status.get("capabilities", {})
    c0.require(caps.get("issueReopen") is True and caps.get("issueWorkflows") is False and
               caps.get("historyExact") is False, "reopen capability overstates full workflow/history")

    def save(label, record):
        c0.require(isinstance(record, dict) and record.get("id", "").startswith(SCOPE) and
                   record.get("type") and record.get("version") and record.get("revision"), label + ": missing identity")
        saved.append({"label": label, "record": copy.deepcopy(record), "semantic_sha256": semantic_digest(record)})
        c0.write_json(capture.output / "expected-records.json", saved)
        return record

    def current(label, records):
        for index, record in enumerate(records):
            c0.require(command(label + "-" + str(index), ["show", record["id"], "--readonly"]) == record,
                       label + ": unexpected current state change")

    def exact(label, record):
        c0.require(command(label, ["show", record["id"], "--version", record["version"], "--readonly"]) == record,
                   label + ": complete retained record changed")

    def ready(label, ids):
        response = capture.success(label, ["ready", "--readonly", "--json"])
        c0.require(response.get("preview") is True and response.get("schemaVersion") == 1 and
                   isinstance(response.get("result"), list), label + ": invalid ready envelope")
        c0.require(sorted(item["id"] for item in response["result"]) == sorted(ids), label + ": wrong readiness")

    def compare(label, before, after):
        result = command(label, ["compare", before["id"], "--from", before["version"], "--to", after["version"], "--readonly"])
        c0.require(result["resource"] == {"id": before["id"], "type": before["type"]} and
                   result["from"] == {"version": before["version"], "attribution": before["attribution"]} and
                   result["to"] == {"version": after["version"], "attribution": after["attribution"]} and
                   result["compared"] == ["properties", "owned"] and result["unsupported"] == ["commonMetadata"],
                   label + ": wrong comparison context")
        changes = []
        old, new = before["properties"], after["properties"]
        for key in sorted(old.keys() | new.keys(), key=lambda item: item.encode("utf-16-be")):
            left, right = {"present": key in old}, {"present": key in new}
            if key in old:
                left["value"] = old[key]
            if key in new:
                right["value"] = new[key]
            if left != right:
                changes.append({"area": "properties", "member": key, "from": left, "to": right})
        c0.require(before["owned"] == after["owned"] and result["changes"] == changes,
                   label + ": comparison lost properties or changed owned Links")

    prerequisite = save("prerequisite-created", command("create-prerequisite",
        ["create", "Prerequisite — 雪", "--id", "beads/prerequisite", "--actor", "creator"]))
    dependent = save("dependent-created", command("create-dependent",
        ["create", "Dependent work", "--id", "beads/dependent", "--description", "Preserve this content", "--actor", "creator"]))
    memory = save("context-memory", c0.envelope(capture.success("remember-context",
                  c0.remember("beads/context", "Durable contextual body", "Context"))))
    added = command("add-blocker", ["dep", "add", dependent["id"], prerequisite["id"], "--actor", "dependency-author"])
    dependent, dependency = save("dependent-owned", added["source"]), save("blocking-link", added["link"])
    relation = command("add-context", ["link", dependent["id"], memory["id"], "--resource-type",
        SCOPE + "types/preview-related-v2", "--id", "links/context", "--properties", '{"note":"preserve"}'])
    context = save("context-link", relation["link"])
    c0.require(relation["source"] == dependent and dependent["owned"] == [dependency], "wrong initial ownership")
    ready("initial-ready", [prerequisite["id"]])
    capture.passed("normal graph initialization authors prerequisite, dependent, blocking Link and separate Memory context")

    def transition(label, verb, before, reason, selector=None):
        result = command(label, [verb, selector or before["id"], "--reason", reason, "--actor", label + "-actor"])
        after = save(label, result.get("issue"))
        c0.require(set(result) == {"issue", "changed"} and result["changed"] is True and
                   after["id"] == before["id"] and after["type"] == before["type"] and
                   after["version"] != before["version"] and after["revision"] == after["version"] and
                   after["owned"] == before["owned"], label + ": wrong changed identity/version/ownership")
        c0.require(after["attribution"]["actor"] == label + "-actor", label + ": missing accepted actor")
        if verb == "reopen":
            c0.require(after["properties"]["status"] == "open", label + ": Issue not open")
            for field in ["closed_at", "close_reason", "closed_by_session", "defer_until"]:
                c0.require(not after["properties"].get(field), label + ": closure state was retained: " + field)
            ignored = {"status", "updated_at", "closed_at", "close_reason", "closed_by_session", "defer_until"}
            c0.require({key: value for key, value in before["properties"].items() if key not in ignored} ==
                       {key: value for key, value in after["properties"].items() if key not in ignored},
                       label + ": reopen changed unrelated content or stored its reason as a property")
        else:
            c0.require(after["properties"]["status"] == "closed" and after["properties"].get("closed_at") and
                       after["properties"].get("close_reason") == reason, label + ": wrong closed state/reason")
        exact(label + "-prior-retained", before)
        compare(label + "-comparison", before, after)
        return after

    prerequisite = transition("close-prerequisite-A", "close", prerequisite, "Reason A")
    closed_a = prerequisite
    ready("ready-after-close-A", [dependent["id"]])
    prerequisite = transition("reopen-prerequisite", "reopen", prerequisite, "Revisit the prerequisite — 雪",
                              selector="beads/prerequisite")
    reopened = prerequisite
    ready("blocked-after-reopen", [prerequisite["id"]])
    current("reopen-preserved-dependencies", [dependent, dependency, context, memory])
    noop = command("reopen-already-open-different-reason",
                   ["reopen", prerequisite["id"], "--reason", "Another reason", "--actor", "different-noop-actor"])
    c0.require(noop == {"issue": prerequisite, "changed": False}, "repeat reopen changed version/attribution/content")
    compare("noop-same-version", prerequisite, prerequisite)
    prerequisite = transition("close-prerequisite-B", "close", prerequisite, "Reason B")
    ready("ready-after-amended-close", [dependent["id"]])
    exact("old-close-A-still-exact", closed_a)
    exact("old-reopened-still-exact", reopened)
    capture.passed("close makes dependent ready; reopen blocks it; no-op stays identical; explicit reclose amends reason without rewriting history")

    # Close the dependent while its prerequisite is closed, then reopen the
    # prerequisite first so reopening the dependent must restore a blocked Issue.
    dependent = transition("close-dependent", "close", dependent, "Pause this dependent")
    ready("both-closed", [])
    prerequisite = transition("reopen-prerequisite-again", "reopen", prerequisite, "Dependency still matters")
    ready("prerequisite-only-ready", [prerequisite["id"]])
    dependent = transition("reopen-dependent", "reopen", dependent, "Resume with its blocker")
    c0.require(dependent["owned"] == [dependency], "reopen lost complete owned Dependency")
    ready("reopened-dependent-remains-blocked", [prerequisite["id"]])
    current("owned-reopen-current", [dependent, prerequisite, dependency, context, memory])
    noop = command("dependent-reopen-noop", ["reopen", dependent["id"], "--reason", "", "--actor", "noop-actor"])
    c0.require(noop == {"issue": dependent, "changed": False}, "dependent no-op changed graph state")
    receipt, out, err = capture.run("human-reopen-noop", ["reopen", dependent["id"], "--reason", "Human no-op"])
    c0.require(receipt["exit_code"] == 0 and not err and dependent["id"] in out and
               "Reopened" not in out and "Unchanged" in out, "human no-op implies a new accepted reopen")
    capture.passed("reopened dependent preserves its owned Link and stays blocked; JSON/human no-ops do not claim a change")

    def refuse(label, args, codes=None):
        result = capture.run(label, [*args, "--json"])
        if codes is not None:
            c0.refusal(result, set(codes), label)
        else:
            receipt, out, err = result
            c0.require(receipt["exit_code"] != 0 and out == "" and err, label + ": unsupported flag did not refuse atomically")

    failures = [
        ("missing-selector", ["reopen"], None),
        ("empty-selector", ["reopen", ""], {"invalid_selector"}),
        ("multiple-selectors", ["reopen", prerequisite["id"], dependent["id"]], {"invalid_selector"}),
        ("unknown-selector", ["reopen", "beads/unknown"], {"not_found"}),
        ("foreign-selector", ["reopen", "https://other.invalid/beads/prerequisite"], {"invalid_selector"}),
        ("malformed-selector", ["reopen", "beads/../prerequisite"], {"invalid_selector"}),
        ("legacy-selector", ["reopen", "reopen-legacy"], {"invalid_selector"}),
        ("memory-kind", ["reopen", memory["id"]], {"invalid_properties"}),
        ("link-kind", ["reopen", dependency["id"]], {"invalid_selector"}),
    ]
    for label, args, codes in failures:
        refuse(label, args, codes)
    unsupported = [
        ["--if-revision", dependent["revision"]], ["--unconditional"], ["--force"],
        ["--if-source-revision", dependent["revision"]], ["--session", "unsupported"],
    ]
    for index, flags in enumerate(unsupported):
        refuse("unsupported-flag-" + str(index), ["reopen", dependent["id"], *flags])
    current("after-refusals", [dependent, prerequisite, dependency, context, memory])
    ready("ready-after-refusals", [prerequisite["id"]])
    capture.passed("missing/multiple/foreign/malformed/unknown/wrong-kind selectors and unsupported flags refuse with empty stdout")

    yaml = capture.work / ".beads" / "config.yaml"
    prior_yaml = yaml.read_bytes() if yaml.exists() else None
    prior_readonly = capture.env.get("BD_READONLY")
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
                freeze.write_text("reopen-smoke\t2026-09-27T00:00:00Z\tgraph write policy\n")
            before = c0.tree_digest(capture.work)
            refuse("policy-" + policy, ["reopen", dependent["id"], "--reason", "Must refuse", *flags], {"permission_denied"})
            c0.require(c0.tree_digest(capture.work) == before, policy + ": refusal changed workspace files")
        finally:
            if prior_readonly is None:
                capture.env.pop("BD_READONLY", None)
            else:
                capture.env["BD_READONLY"] = prior_readonly
            if prior_yaml is None:
                yaml.unlink(missing_ok=True)
            else:
                yaml.write_bytes(prior_yaml)
            if policy == "freeze":
                freeze.unlink(missing_ok=True)
                (mayor / "town.json").unlink(missing_ok=True)
                mayor.rmdir()
    current("after-policy", [dependent, prerequisite, dependency, context, memory])
    for entry in saved:
        exact(entry["label"] + "-final-retained", entry["record"])
    capture.passed("readonly/freeze refuse even no-op requests; all complete saved versions survive fresh processes")

    graph_work = capture.work
    capture.work = capture.root / "legacy-workspace"
    capture.work.mkdir(mode=0o700)
    try:
        receipt, _, err = capture.run("normal-legacy-init", init_args(False, "lreopen"))
        c0.require(receipt["exit_code"] == 0, "legacy init failed: " + err)
        legacy = capture.success("legacy-create", ["create", "Legacy reopen control", "--json"])
        legacy_id = legacy["id"]
        capture.success("legacy-close", ["close", legacy_id, "--reason", "Legacy close", "--json"])
        opened = capture.success("legacy-reopen", ["reopen", legacy_id, "--reason", "Legacy reason", "--json"])
        c0.require(isinstance(opened, list) and len(opened) == 1 and opened[0]["id"] == legacy_id and
                   opened[0]["status"] == "open" and not opened[0].get("closed_at") and
                   not opened[0].get("close_reason"), "legacy reopen output changed")
        before = capture.success("legacy-current", ["show", legacy_id, "--json"])
        receipt, out, err = capture.run("legacy-noop", ["reopen", legacy_id, "--reason", "Another legacy reason", "--json"])
        c0.require(receipt["exit_code"] == 0 and out == "" and "already open" in err.lower(),
                   "legacy no-op output changed")
        c0.require(capture.success("legacy-current-after-noop", ["show", legacy_id, "--json"]) == before,
                   "legacy no-op changed current Issue")
    finally:
        capture.work = graph_work
    capture.passed("ordinary legacy reopen keeps its changed-array and already-open no-op behavior")
    # Isolate output-mode coverage after all earlier readiness assertions.
    # This Issue has no blockers and cannot change the story's dependency state.
    output_versions_start = len(saved)
    standalone = save("output-control-created", command("create-output-control",
        ["create", "Human and quiet reopen control", "--id", "beads/output-control", "--actor", "output-author"]))
    for mode in ["human", "quiet"]:
        previous = standalone
        closed_result = command(mode + "-control-close", ["close", standalone["id"],
            "--reason", mode + " output preparation", "--actor", "output-closer"])
        closed = save(mode + "-control-closed", closed_result["issue"])
        c0.require(closed_result["changed"] is True and closed["id"] == previous["id"] and
                   closed["type"] == previous["type"] and closed["owned"] == previous["owned"] == [] and
                   closed["version"] != previous["version"] and closed["properties"]["status"] == "closed" and
                   closed["properties"].get("closed_at") and
                   closed["properties"].get("close_reason") == mode + " output preparation",
                   mode + ": standalone close failed")
        args = ["reopen", closed["id"], "--reason", mode + " output check", "--actor", mode + "-reopener"]
        if mode == "quiet":
            args.append("--quiet")
        receipt, _, _ = capture.run(mode + "-changed-reopen", args)
        artifact = capture.output / capture.records[-1]["artifact"]
        stdout, stderr = (artifact / "stdout.log").read_bytes(), (artifact / "stderr.log").read_bytes()
        expected_stdout = ("Reopened " + closed["id"] + "\n\n").encode() if mode == "human" else b""
        c0.require(receipt["exit_code"] == 0 and stdout == expected_stdout and stderr == b"",
                   mode + ": changed reopen output is not exact")
        standalone = save(mode + "-control-reopened", command(mode + "-reopened-current",
                           ["show", closed["id"], "--readonly"]))
        c0.require(standalone["id"] == closed["id"] and standalone["type"] == closed["type"] and
                   standalone["owned"] == closed["owned"] and standalone["version"] != closed["version"] and
                   standalone["revision"] == standalone["version"] and standalone["properties"]["status"] == "open" and
                   standalone["attribution"]["actor"] == mode + "-reopener",
                   mode + ": output succeeded without the expected versioned reopen")
        for field in ["closed_at", "close_reason", "closed_by_session", "defer_until"]:
            c0.require(not standalone["properties"].get(field), mode + ": reopen retained closure state")
        ignored = {"status", "updated_at", "closed_at", "close_reason", "closed_by_session", "defer_until"}
        c0.require({key: value for key, value in closed["properties"].items() if key not in ignored} ==
                   {key: value for key, value in standalone["properties"].items() if key not in ignored},
                   mode + ": reopen changed unrelated content")
    for entry in saved[output_versions_start:]:
        exact(entry["label"] + "-output-retained", entry["record"])
    capture.passed("changed human reopen preserves canonical result and existing renderer blank line; quiet prints nothing; both commit readable versioned state")

    return {"saved_versions": len(saved), "refusal_cases": len(failures) + len(unsupported) + 4,
            "http_exercised": False, "saved_records_sha256": c0.sha256(capture.output / "expected-records.json")}


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
