#!/usr/bin/env python3
"""Installed graph Issue text-edit proof, using normal CLI authoring only.

Each command is a fresh process. Embedded and caller-owned ordinary Dolt run
sequentially; no SQL, fixtures, mocks or server provisioning occur here.
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
SCOPE = "https://example.invalid/disposable-issue-edit/"
FIELDS = {"title": "title", "description": "description",
          "design": "design", "acceptance_criteria": "acceptance"}
LIMITATIONS = [
    "this suite exercises four inline Issue text fields only; notes edits are explicitly unavailable; no file/stdin or complete Issue workflows",
    "existing attribution is not native History commit-time context",
    "no concurrency, corruption, cancellation, crash or uncertain-commit qualification",
    "visible equality does not prove unchanged database bytes, event counts or coordination",
    "no HTTP/public-client qualification here",
    "ordinary Dolt 2.1.8 database provisioning must remain serialized",
]

def semantic_digest(value):
    return hashlib.sha256(json.dumps(value, ensure_ascii=False, sort_keys=True,
                                     separators=(",", ":")).encode()).hexdigest()

def exercise(capture):
    backend = "server" if capture.args.server_port else "embedded"
    init = ["init", "--graph-mode", "link", "--scope-url", SCOPE, "--prefix", "issueedit",
            "--non-interactive", "--skip-hooks", "--skip-agents", "--json"]
    if capture.args.server_port:
        init += ["--server", "--external", "--server-host", "127.0.0.1", "--server-port",
                 str(capture.args.server_port), "--database", "iedit_" + capture.root.name.replace("-", "_"),
                 "--server-user", "root"]
    initialized = c0.envelope(capture.success("normal-init", init))
    c0.require(initialized.get("scope") == SCOPE and initialized.get("backend") == backend,
               "wrong initialized authority/backend")
    status = c0.envelope(capture.success("status", ["status", "--graph", "--json"]))
    caps = status.get("capabilities", {})
    c0.require(caps.get("issueTextUpdate") is True and caps.get("issueTextFileInput") is False
               and caps.get("issueTextStdinInput") is False, "wrong text capability advertisement")
    c0.require(caps.get("issueWorkflows") is False and caps.get("historyExact") is False
               and caps.get("memory") is False, "bounded edit overstates complete capabilities")
    capture.passed("normal initialization and truthful bounded Issue text capabilities")
    saved = []

    def save(label, value):
        c0.require(isinstance(value, dict) and value.get("id", "").startswith(SCOPE), label + ": invalid record")
        for key in ["type", "revision", "version"]:
            c0.require(isinstance(value.get(key), str) and value[key], label + ": missing " + key)
        saved.append({"label": label, "record": copy.deepcopy(value), "semantic_sha256": semantic_digest(value)})
        c0.write_json(capture.output / "expected-records.json", saved)
        return value

    def current(label, records):
        for i, record in enumerate(records):
            actual = c0.envelope(capture.success(label + "-" + str(i),
                                ["show", record["id"], "--readonly", "--json"]))
            c0.require(actual == record, label + ": current or unrelated record changed")

    def exact(label, record):
        actual = c0.envelope(capture.success(label, ["show", record["id"], "--version",
                                                   record["version"], "--readonly", "--json"]))
        c0.require(actual == record, label + ": retained body/owned state differs")

    def ready(label):
        response = capture.success(label, ["ready", "--readonly", "--json"])
        c0.require(response.get("preview") is True and response.get("schemaVersion") == 1, label + ": wrong envelope")
        rows = response.get("result")
        c0.require(isinstance(rows, list), label + ": ready must be an array")
        ids = [row["id"] for row in rows]
        c0.require(len(ids) == len(set(ids)), label + ": duplicate ready ID")
        return sorted(ids)

    def compare(label, before, after):
        result = c0.envelope(capture.success(label, ["compare", before["id"], "--from", before["version"],
                                                    "--to", after["version"], "--readonly", "--json"]))
        c0.require(result.get("resource") == {"id": before["id"], "type": before["type"]}, label + ": wrong identity")
        for side, record in [("from", before), ("to", after)]:
            c0.require(result.get(side) == {"version": record["version"], "attribution": record["attribution"]},
                       label + ": wrong version context")
        c0.require(result.get("compared") == ["properties", "owned"] and
                   result.get("unsupported") == ["commonMetadata"], label + ": wrong coverage")
        old, new, changes = before["properties"], after["properties"], []
        for key in sorted(old.keys() | new.keys(), key=lambda item: item.encode("utf-16-be")):
            left, right = {"present": key in old}, {"present": key in new}
            if key in old:
                left["value"] = old[key]
            if key in new:
                right["value"] = new[key]
            if left != right:
                changes.append({"area": "properties", "member": key, "from": left, "to": right})
        c0.require(before["owned"] == after["owned"], label + ": oracle requires unchanged owned Links")
        c0.require(result.get("changes") == changes, label + ": incomplete properties or changed owned set")

    source = save("source-created", c0.envelope(capture.success("source-create",
        ["create", "Original source", "--id", "beads/source", "--description", "Original description", "--json"])))
    target = save("target-created", c0.envelope(capture.success("target-create",
        ["create", "Prerequisite", "--id", "beads/target", "--json"])))
    memory = save("context-memory", c0.envelope(capture.success("memory-create",
        c0.remember("beads/context", "Durable context — 雪", "Context"))))
    dependency = c0.envelope(capture.success("dependency-create",
        ["dep", "add", "beads/source", "beads/target", "--json"]))
    dep, source = save("blocking-link", dependency["link"]), save("source-owned", dependency["source"])
    relation = c0.envelope(capture.success("informational-link-create",
        ["link", "beads/source", "beads/context", "--resource-type", SCOPE + "types/preview-related-v2",
         "--id", "links/context", "--properties", '{"note":"Keep this relation"}', "--json"]))
    context_link = save("informational-link", relation["link"])
    c0.require(relation["source"] == source and source["owned"] == [dep],
               "informational Link changed Issue version or blocking ownership")
    unrelated = [target, memory, dep, context_link]
    current("baseline", [source, *unrelated])
    ready_before = ready("ready-before")
    c0.require(source["id"] not in ready_before and target["id"] in ready_before, "blocking readiness missing")

    def flags_for(fields):
        flags = []
        for key, value in fields.items():
            flags += ["--" + FIELDS[key], value]
        return flags

    def update(label, before, fields, unconditional=False, selector=None):
        guard = ["--unconditional"] if unconditional else ["--if-revision", before["revision"]]
        result = c0.envelope(capture.success(label, ["update", selector or before["id"], *flags_for(fields),
                                                   *guard, "--actor", "text-editor", "--json"]))
        after = save(label, result.get("issue"))
        c0.require(result.get("changed") is True and after["version"] != before["version"] and
                   after["revision"] == after["version"], label + ": wrong changed/version result")
        c0.require(after["id"] == before["id"] and after["type"] == before["type"] and
                   after["owned"] == before["owned"], label + ": changed identity or complete owned set")
        wanted = copy.deepcopy(before["properties"])
        for key, value in fields.items():
            value = value.strip() if key == "title" else value
            if value == "" and key != "title":
                wanted.pop(key, None)
            else:
                wanted[key] = value
        wanted["updated_at"] = after["properties"]["updated_at"]
        c0.require(after["properties"] == wanted, label + ": changed an omitted or non-text Issue field")
        c0.require(after["attribution"]["actor"] == "text-editor", label + ": missing recorded actor")
        current(label + "-reopened", [after, *unrelated])
        exact(label + "-old-version", before)
        compare(label + "-compare", before, after)
        return after

    all_fields = {"title": "  Edited source — 雪  ", "description": " Description\r\nbody 😀\n",
                  "design": "# Design\r\n  preserve spaces  ",
                  "acceptance_criteria": "Acceptance — e\u0301\n"}
    original = source
    source = update("all-four-text-fields", source, all_fields, selector="beads/source")
    c0.require(ready("ready-after-edit") == ready_before, "text edit changed readiness")
    noop_fields = {key: source["properties"][key] for key in FIELDS}
    noop = c0.envelope(capture.success("different-actor-noop", ["update", source["id"], *flags_for(noop_fields),
        "--if-revision", source["revision"], "--actor", "other-actor", "--json"]))
    c0.require(noop.get("changed") is False and noop.get("issue") == source, "no-op changed revision/actor/state")
    compare("same-version-compare", source, source)
    capture.passed("all four inline edits preserve identity, owned Dependencies, unrelated context and readiness; no-op preserves state")

    def refuse(label, argv, code):
        c0.refusal(capture.run(label, [*argv, "--json"]), {code}, label)

    for label, fields in [("stale-change", {"description": "must not persist"}), ("stale-noop", noop_fields)]:
        refuse(label, ["update", source["id"], *flags_for(fields), "--if-revision", original["revision"]],
               "revision_conflict")
    source = update("clear-optional-text", source, {key: "" for key in FIELDS if key != "title"}, unconditional=True)
    source = update("description-only", source, {"description": "Only this field changes"})
    source = update("design-literal-dash", source, {"design": "-"})
    aliases = c0.envelope(capture.success("identical-description-aliases-noop", ["update", source["id"],
        "--description", source["properties"]["description"], "--body", source["properties"]["description"],
        "--message", source["properties"]["description"], "--if-revision", source["revision"], "--json"]))
    c0.require(aliases.get("changed") is False and aliases.get("issue") == source,
               "identical description aliases changed the version or meaning")
    capture.passed("stale edit/no-op refuse; empty clears and omitted fields remain distinct")

    guard = ["--title", "Denied", "--if-revision", source["revision"]]
    failures = [
        ("missing-guard", ["update", source["id"], "--title", "Denied"], "invalid_selector"),
        ("both-guards", ["update", source["id"], *guard, "--unconditional"], "invalid_selector"),
        ("empty-guard", ["update", source["id"], "--title", "Denied", "--if-revision", ""], "invalid_selector"),
        ("false-unconditional", ["update", source["id"], "--title", "Denied", "--unconditional=false"], "invalid_selector"),
        ("no-field", ["update", source["id"], "--if-revision", source["revision"]], "invalid_properties"),
        ("empty-title", ["update", source["id"], "--title", " \t ", "--unconditional"], "invalid_properties"),
        ("long-title", ["update", source["id"], "--title", "x" * 501, "--unconditional"], "invalid_properties"),
        ("body-dash", ["update", source["id"], "--body", "-", "--unconditional"], "capability_unavailable"),
        ("message-dash", ["update", source["id"], "--message", "-", "--unconditional"], "capability_unavailable"),
        ("conflicting-aliases", ["update", source["id"], "--description", "one", "--body", "two", "--unconditional"], "invalid_properties"),
        ("description-dash", ["update", source["id"], "--description", "-", "--unconditional"], "capability_unavailable"),
        ("empty-notes", ["update", source["id"], "--notes", "", "--unconditional"], "capability_unavailable"),
        ("memory-kind", ["update", memory["id"], "--title", "Denied", "--unconditional"], "capability_unavailable"),
        ("link-kind", ["update", context_link["id"], "--title", "Denied", "--unconditional"], "capability_unavailable"),
        ("unknown", ["update", "beads/missing", "--title", "Denied", "--unconditional"], "not_found"),
        ("legacy-selector", ["update", "issueedit-legacy", "--title", "Denied", "--unconditional"], "invalid_selector"),
        ("mixed-properties", ["update", source["id"], *guard, "--properties", '{}'], "capability_unavailable"),
    ]
    for flag, value in [("status", "closed"), ("add-label", "new-label"), ("type", "bug"), ("assignee", "worker"),
                        ("notes", "replace"), ("append-notes", "append"), ("metadata", '{}'), ("if-source-revision", source["revision"]),
                        ("body-file", str(capture.root / "does-not-exist.md"))]:
        failures.append(("unsupported-" + flag, ["update", source["id"], *guard, "--" + flag, value],
                         "capability_unavailable"))
    for label, argv, code in failures:
        refuse(label, argv, code)
    current("after-refusals", [source, *unrelated])
    c0.require(ready("ready-after-refusals") == ready_before, "refusal changed readiness")
    capture.passed("bad guards/titles, unsupported sources, wrong kinds, unknown selectors and mixed flags refuse")

    yaml = capture.work / ".beads" / "config.yaml"
    old_yaml = yaml.read_bytes() if yaml.exists() else None
    old_readonly = capture.env.get("BD_READONLY")
    mayor, freeze = capture.work / "mayor", capture.work / "MIGRATION-FREEZE"
    c0.require(not mayor.exists() and not freeze.exists(), "unexpected town policy files")
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
                freeze.write_text("issue-edit\t2026-09-26T00:00:00Z\tIssue text policy\n")
            before = c0.tree_digest(capture.work)
            refuse("policy-" + policy, ["update", source["id"], "--description", "-", "--unconditional", *flags],
                   "permission_denied")
            c0.require(c0.tree_digest(capture.work) == before, policy + ": changed workspace files")
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
    current("after-policy", [source, *unrelated])
    capture.passed("readonly flag/environment/configuration and freeze refuse before unsupported source handling")

    target = save("target-closed", c0.envelope(capture.success("close-target", ["close", target["id"], "--json"]))["issue"])
    source = save("source-closed", c0.envelope(capture.success("close-source", ["close", source["id"], "--json"]))["issue"])
    unrelated = [target, memory, dep, context_link]
    closed = source
    c0.require(closed["properties"].get("status") == "closed" and closed["properties"].get("closed_at"), "source not closed")
    source = update("closed-text-edit", source, {"description": "Text after close — 雪", "design": "Post-close design"})
    c0.require(source["properties"]["closed_at"] == closed["properties"]["closed_at"] and
               source["properties"]["status"] == "closed", "text edit changed closed status/time")
    c0.require(ready("ready-after-closed-edit") == [], "closed Issue reentered ready")
    capture.passed("closed Issue permits text edit without reopening, changed closed time, Link or target")
    for entry in saved:
        exact(entry["label"] + "-final-retained", entry["record"])
    current("final-current", [source, *unrelated])
    capture.passed("all complete retained versions and current records survive new processes after all edits/refusals")
    return {"saved_versions": len(saved), "saved_records_sha256": c0.sha256(capture.output / "expected-records.json"),
            "final_issue_semantic_sha256": semantic_digest(source), "http_exercised": False}


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
