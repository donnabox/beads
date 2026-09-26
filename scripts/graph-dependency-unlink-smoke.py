#!/usr/bin/env python3
"""Installed CLI proof for canonical blocking Dependency removal.

Fresh processes author a normally initialized disposable workspace. Embedded
and caller-owned ordinary Dolt are exercised sequentially; no SQL or fixtures.
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
SCOPE = "https://example.invalid/disposable-dependency-unlink/"
LIMITATIONS = [
    "canonical blocking Dependency ID removal only; typed blocking pair and legacy dep remove remain unavailable",
    "saved-version reads and comparison are not complete native or public HTTP History",
    "no concurrent writers, corruption, cancellation, crash or uncertain-commit qualification here",
    "visible equality does not prove unchanged internal bytes, version counts or event counts",
    "public BDP client qualification runs separately",
    "ordinary Dolt 2.1.8 database provisioning must remain serialized",
]

def semantic_digest(value):
    return hashlib.sha256(json.dumps(value, ensure_ascii=False, sort_keys=True,
                                     separators=(",", ":")).encode()).hexdigest()

def exercise(capture):
    backend = "server" if capture.args.server_port else "embedded"
    init = ["init", "--graph-mode", "link", "--scope-url", SCOPE, "--prefix", "depunlink",
            "--non-interactive", "--skip-hooks", "--skip-agents", "--json"]
    if capture.args.server_port:
        init += ["--server", "--external", "--server-host", "127.0.0.1", "--server-port",
                 str(capture.args.server_port), "--database", "unlink_" + capture.root.name.replace("-", "_"),
                 "--server-user", "root"]
    initialized = c0.envelope(capture.success("normal-init", init))
    c0.require(initialized.get("scope") == SCOPE and initialized.get("backend") == backend,
               "wrong initialized authority/backend")
    status = c0.envelope(capture.success("status", ["status", "--graph", "--json"]))
    caps = status.get("capabilities", {})
    c0.require(caps.get("blockingDependencyUnlink") is True and caps.get("blockingDependencyPairUnlink") is False
               and caps.get("issueWorkflows") is False and caps.get("historyExact") is False,
               "incorrect bounded unlink capability advertisement")
    capture.passed("normal initialization advertises canonical blocking unlink and keeps broader contracts unavailable")
    saved = []

    def save(label, value):
        c0.require(isinstance(value, dict) and value.get("id", "").startswith(SCOPE), label + ": missing identity")
        for key in ["type", "revision", "version"]:
            c0.require(isinstance(value.get(key), str) and value[key], label + ": missing " + key)
        saved.append({"label": label, "record": copy.deepcopy(value), "semantic_sha256": semantic_digest(value)})
        c0.write_json(capture.output / "expected-records.json", saved)
        return value

    def show(label, subject):
        return c0.envelope(capture.success(label, ["show", subject, "--readonly", "--json"]))

    def current(label, records):
        for index, record in enumerate(records):
            c0.require(show(label + "-" + str(index), record["id"]) == record, label + ": current record changed")

    def exact(label, record):
        value = c0.envelope(capture.success(label, ["show", record["id"], "--version", record["version"],
                                                   "--readonly", "--json"]))
        c0.require(value == record, label + ": retained complete record changed")

    def rows(label, argv):
        response = capture.success(label, [*argv, "--readonly", "--json"])
        c0.require(response.get("preview") is True and response.get("schemaVersion") == 1, label + ": wrong envelope")
        result = response.get("result")
        c0.require(isinstance(result, list), label + ": expected array")
        c0.require(len({item["id"] for item in result}) == len(result), label + ": duplicate IDs")
        return result

    def ready(label):
        return sorted(item["id"] for item in rows(label, ["ready"]))

    def incident(label, expected):
        actual = rows(label, ["links", "beads/source", "--direction", "both"])
        c0.require(actual == sorted(expected, key=lambda item: item["id"]), label + ": incident inventory differs")

    def refuse(label, argv, codes):
        c0.refusal(capture.run(label, [*argv, "--json"]), set(codes), label)

    source = save("source-empty", c0.envelope(capture.success("create-source",
                  ["create", "Source — 雪", "--id", "beads/source", "--json"])))
    first_target = save("first-target", c0.envelope(capture.success("create-first",
                  ["create", "First prerequisite", "--id", "beads/first", "--json"])))
    second_target = save("second-target", c0.envelope(capture.success("create-second",
                  ["create", "Second prerequisite", "--id", "beads/second", "--json"])))
    memory = save("memory", c0.envelope(capture.success("remember-context",
                  c0.remember("beads/context", "Keep this context", "Context"))))
    related = SCOPE + "types/preview-related-v2"
    info = c0.envelope(capture.success("informational-link", ["link", "beads/source", "beads/context",
            "--resource-type", related, "--id", "links/context", "--properties", '{"note":"survives"}', "--json"]))
    informational = save("informational", info["link"])
    c0.require(info["source"] == source, "unowned informational Link changed Issue")
    deps = []
    for label, target in [("first", first_target), ("second", second_target)]:
        result = c0.envelope(capture.success("add-" + label,
                            ["dep", "add", source["id"], target["id"], "--actor", "dependency-author", "--json"]))
        c0.require(result.get("changed") is True, "Dependency add did not change")
        deps.append(save(label + "-link", result["link"]))
        source = save("source-after-" + label, result["source"])
    first, second = deps
    c0.require(source["owned"] == sorted(deps, key=lambda item: item["id"]), "owned set lost a blocking Dependency")
    unchanged = [first_target, second_target, memory, informational]
    wanted_ready = sorted([first_target["id"], second_target["id"]])
    c0.require(ready("ready-two-blockers") == wanted_ready, "source ready with two open blockers")
    incident("incident-two-blockers", [*deps, informational])
    current("seed-reopen", [source, *deps, *unchanged])
    capture.passed("normal CLI creates two blocking prerequisites plus unchanged informational context")

    guard = ["--if-revision", first["revision"], "--if-source-revision", source["revision"]]
    bad_guard = {"invalid_selector", "invalid_properties"}
    failures = [
        ("missing-link-guard", [first["id"], "--if-source-revision", source["revision"]], bad_guard),
        ("missing-source-guard", [first["id"], "--if-revision", first["revision"]], bad_guard),
        ("both-link-guards", [first["id"], *guard, "--unconditional"], bad_guard),
        ("both-source-guards", [first["id"], *guard, "--unconditional-source"], bad_guard),
        ("false-link-unconditional", [first["id"], "--unconditional=false", "--unconditional-source"], bad_guard),
        ("false-source-unconditional", [first["id"], "--unconditional", "--unconditional-source=false"], bad_guard),
        ("empty-link-guard", [first["id"], "--if-revision", "", "--unconditional-source"], bad_guard),
        ("empty-source-guard", [first["id"], "--unconditional", "--if-source-revision", ""], bad_guard),
        ("stale-link", [first["id"], "--if-revision", "stale", "--if-source-revision", source["revision"]], {"revision_conflict"}),
        ("stale-source", [first["id"], "--if-revision", first["revision"], "--if-source-revision", "stale"], {"revision_conflict"}),
        ("blocking-pair", [source["id"], first_target["id"], "--resource-type", first["type"],
                          "--unconditional", "--unconditional-source"], {"capability_unavailable"}),
        ("wrong-kind", [source["id"], "--unconditional", "--unconditional-source"], {"invalid_selector"}),
        ("missing-link", ["links/missing", "--unconditional", "--unconditional-source"], {"not_found"}),
        ("legacy-selector", ["depunlink-legacy", "--unconditional", "--unconditional-source"], {"invalid_selector"}),
        ("mixed-type-selector", [first["id"], *guard, "--resource-type", first["type"]], {"invalid_selector"}),
    ]
    for label, argv, codes in failures:
        refuse(label, ["unlink", *argv], codes)
    current("refusals-unchanged", [source, *deps, *unchanged])
    c0.require(ready("ready-after-refusals") == wanted_ready, "refusals changed readiness")
    capture.passed("separate guards, stale revisions, selector/kind and typed-pair refusals preserve current state")

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
                freeze.write_text("dependency-unlink\t2026-09-26T00:00:00Z\tgraph write policy\n")
            before = c0.tree_digest(capture.work)
            refuse("policy-" + policy, ["unlink", first["id"], *guard, *flags], {"permission_denied"})
            c0.require(c0.tree_digest(capture.work) == before, policy + ": changed workspace bytes")
        finally:
            if old_readonly is None:
                capture.env.pop("BD_READONLY", None)
            else:
                capture.env["BD_READONLY"] = old_readonly
            if old_yaml is None:
                yaml.unlink(missing_ok=True)
            else:
                yaml.write_bytes(old_yaml)
            if policy == "freeze":
                freeze.unlink(missing_ok=True)
                (mayor / "town.json").unlink(missing_ok=True)
                mayor.rmdir()
    current("policy-unchanged", [source, *deps, *unchanged])
    capture.passed("readonly flag/environment/configuration and freeze refuse without visible changes")

    def unlink(label, before, link, remaining, unconditional=False):
        guards = ["--unconditional", "--unconditional-source"] if unconditional else [
            "--if-revision", link["revision"], "--if-source-revision", before["revision"]]
        result = c0.envelope(capture.success(label, ["unlink", link["id"], *guards,
                                                  "--actor", "dependency-remover", "--json"]))
        tombstone = result.get("link", {})
        c0.require(result.get("changed") is True and tombstone.get("id") == link["id"] and
                   tombstone.get("type") == link["type"] and tombstone.get("state") == "deleted" and
                   tombstone.get("previousVersion") == link["version"] and
                   tombstone.get("version") == tombstone.get("revision") and
                   tombstone.get("version") != link["version"], label + ": wrong tombstone")
        c0.require(tombstone.get("attribution", {}).get("actor") == "dependency-remover", label + ": lost actor")
        after = save(label + "-source", result.get("source"))
        c0.require(after.get("attribution", {}).get("actor") == "dependency-remover", label + ": lost source actor")
        c0.require(after["id"] == before["id"] and after["type"] == before["type"] and
                   after["version"] != before["version"] and after["revision"] == after["version"],
                   label + ": wrong source identity/version")
        c0.require(after["owned"] == sorted(remaining, key=lambda item: item["id"]), label + ": wrong remaining owned set")
        before_props, after_props = copy.deepcopy(before["properties"]), copy.deepcopy(after["properties"])
        before_props.pop("updated_at", None)
        after_props.pop("updated_at", None)
        c0.require(before_props == after_props, label + ": removal altered Issue content")
        current(label + "-current", [after, *remaining, *unchanged])
        incident(label + "-incident", [*remaining, informational])
        exact(label + "-old-source", before)
        exact(label + "-old-link", link)
        comparison = c0.envelope(capture.success(label + "-compare", ["compare", before["id"],
            "--from", before["version"], "--to", after["version"], "--readonly", "--json"]))
        owned_changes = [change for change in comparison["changes"] if change["area"] == "owned"]
        c0.require(owned_changes == [{"area": "owned", "id": link["id"],
            "from": {"present": True, "value": link}, "to": {"present": False}}],
            label + ": comparison does not report exactly removed owned Link")
        refuse(label + "-gone", ["show", link["id"], "--readonly"], {"gone"})
        refuse(label + "-deletion-token-gone", ["show", link["id"], "--version", tombstone["version"], "--readonly"], {"gone"})
        refuse(label + "-repeat", ["unlink", link["id"], *guards], {"gone"})
        current(label + "-repeat-unchanged", [after, *remaining, *unchanged])
        return after

    source = unlink("remove-first", source, first, [second])
    c0.require(ready("ready-one-blocker") == wanted_ready, "source ready while one blocker remains")
    source = unlink("remove-final", source, second, [], unconditional=True)
    c0.require(ready("ready-no-blockers") == sorted([source["id"], *wanted_ready]), "source not ready after final removal")
    capture.passed("partial unlink stays blocked; final unconditional unlink becomes ready; old source/Link remain readable")

    readded = c0.envelope(capture.success("readd-same-pair",
                          ["dep", "add", source["id"], first_target["id"], "--actor", "dependency-reauthor", "--json"]))
    replacement = save("replacement-link", readded["link"])
    source = save("source-readded", readded["source"])
    c0.require(readded.get("changed") is True and replacement["id"] != first["id"] and
               replacement["source"] == first["source"] and replacement["target"] == first["target"] and
               replacement["type"] == first["type"] and source["owned"] == [replacement],
               "same-pair re-add resurrected old identity or lost mapping")
    c0.require(ready("ready-readded") == wanted_ready, "same-pair re-add did not block source")
    incident("incident-readded", [replacement, informational])
    for label, old_link in [("first", first), ("second", second)]:
        refuse("old-" + label + "-still-gone", ["show", old_link["id"], "--readonly"], {"gone"})
        exact("old-" + label + "-retained-after-readd", old_link)
        refuse("reserved-" + label, ["link", source["id"], memory["id"], "--resource-type", related,
               "--id", old_link["id"].removeprefix(SCOPE), "--properties", "{}"], {"identity_reserved"})
    current("final-current", [source, replacement, *unchanged])
    for entry in saved:
        exact(entry["label"] + "-final-retained", entry["record"])
    incident("final-inventory", [replacement, informational])
    capture.passed("same pair gets a fresh canonical Link; old IDs stay gone/reserved; current and retained inventories remain healthy")
    return {"saved_versions": len(saved), "saved_records_sha256": c0.sha256(capture.output / "expected-records.json"),
            "final_source_semantic_sha256": semantic_digest(source), "http_exercised": False}


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
