#!/usr/bin/env python3
"""Exercise installed-CLI Memory replacement and exact retained versions.

Private workspaces use normal init/remember/create/link/update commands only.
Each command runs in a fresh process. Embedded and supplied ordinary Dolt modes
run sequentially. Caller owns the server and must serialize unrelated database
provisioning. No SQL, seeded schema, mocks or HTTP fixtures are used.
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
SCOPE = "https://example.invalid/disposable-memory-update/"
INPUT_LIMIT = 1 << 20
LIMITATIONS = [
    "Memory title/body replacement only; no complete Memory or public History claim",
    "exact saved-version reads do not establish enumeration, lineage or historical BDP HTTP",
    "no crash, uncertain-commit, corruption, restore or concurrent-writer qualification",
    "visible current-record equality is not proof of unchanged internal database bytes",
    "no HTTP/public-client check here; the parent runs that independently",
    "ordinary Dolt 2.1.8 server provisioning must remain serialized",
]


def semantic_digest(value):
    raw = json.dumps(value, ensure_ascii=False, sort_keys=True, separators=(",", ":")).encode("utf-8")
    return hashlib.sha256(raw).hexdigest()


def exercise(capture):
    backend = "server" if capture.args.server_port else "embedded"
    init = ["init", "--graph-mode", "link", "--scope-url", SCOPE, "--prefix", "memupdate",
            "--non-interactive", "--skip-hooks", "--skip-agents", "--json"]
    if capture.args.server_port:
        init += ["--server", "--external", "--server-host", "127.0.0.1", "--server-port",
                 str(capture.args.server_port), "--database", "memory_" + capture.root.name.replace("-", "_"),
                 "--server-user", "root"]
    initialized = c0.envelope(capture.success("normal-init", init))
    c0.require(initialized.get("scope") == SCOPE and initialized.get("backend") == backend,
               "normal initialization selected the wrong authority/backend")
    capture.passed("normal disposable graph initialization")
    status = c0.envelope(capture.success("capabilities", ["status", "--graph", "--json"]))
    capabilities = status.get("capabilities", {})
    c0.require(capabilities.get("memoryPropertiesUpdate") is True and capabilities.get("exactVersionRead") is True,
               "status omitted implemented update/exact-read capabilities")
    c0.require(capabilities.get("memory") is False and capabilities.get("historyExact") is False,
               "partial Memory update claimed full Memory or public History")
    c0.require(status.get("limits", {}).get("memoryPropertiesInputBytes") == INPUT_LIMIT,
               "status omitted the 1 MiB Memory input limit")
    capture.passed("status separates Memory property updates from complete Memory/History")
    expected, inputs = [], []
    input_dir = capture.output / "inputs"
    input_dir.mkdir(mode=0o700)

    def payload_file(name, raw):
        path = input_dir / name
        path.write_bytes(raw)
        inputs.append({"path": str(path), "bytes": len(raw), "sha256": c0.sha256(path)})
        c0.write_json(capture.output / "inputs.json", inputs)
        return "@" + str(path)

    def save(label, value):
        c0.require(isinstance(value, dict) and value.get("id", "").startswith(SCOPE),
                   f"{label}: missing canonical record")
        for field in ["type", "revision", "version"]:
            c0.require(isinstance(value.get(field), str) and value[field], f"{label}: missing {field}")
        expected.append({"label": label, "record": copy.deepcopy(value), "semantic_sha256": semantic_digest(value)})
        c0.write_json(capture.output / "expected-records.json", expected)
        return value

    def show(label, value):
        subject = value if isinstance(value, str) else value["id"]
        return c0.envelope(capture.success(label, ["show", subject, "--json"]))

    def assert_current(label, records):
        for index, value in enumerate(records):
            c0.require(show(f"{label}-{index}", value) == value, f"{label}: unrelated/current state changed")

    def exact(label, value):
        actual = c0.envelope(capture.success(label, ["show", value["id"], "--version", value["version"],
                                                          "--readonly", "--json"]))
        c0.require(actual == value and semantic_digest(actual) == semantic_digest(value),
                   f"{label}: retained content or owned set changed")

    original = save("memory-original", c0.envelope(capture.success("memory-create", c0.remember(
        "beads/plan", "Original body — 雪", "Original title"))))
    observer = save("incoming-owner-original", c0.envelope(capture.success("observer-create", c0.remember(
        "beads/observer", "Observes plan", "Observer"))))
    target = save("issue-target", c0.envelope(capture.success("issue-create", [
        "create", "Apply the plan", "--id", "beads/work", "--json"])))
    c0.require(original["type"] == observer["type"] == SCOPE + "types/preview-memory-v2",
               "remember did not create Memory Beads")
    related = SCOPE + "types/preview-related-v2"
    outgoing = c0.envelope(capture.success("outgoing-owned-link", [
        "link", "beads/plan", "beads/work", "--resource-type", related, "--id", "links/outgoing",
        "--properties", '{"note":"original outgoing context"}', "--if-source-revision", original["revision"], "--json"]))
    link_before = save("outgoing-link-original", outgoing["link"])
    before = save("memory-original-with-owned-link", outgoing["source"])
    incoming = c0.envelope(capture.success("incoming-owned-link", [
        "link", "beads/observer", "beads/plan", "--resource-type", related, "--id", "links/incoming",
        "--properties", '{"note":"incoming relationship"}', "--if-source-revision", observer["revision"], "--json"]))
    incoming_link = save("incoming-link", incoming["link"])
    incoming_owner = save("incoming-owner-with-link", incoming["source"])
    c0.require(before.get("owned") == [link_before], "Memory did not retain its complete outgoing Link")
    unchanged = [link_before, target, incoming_link, incoming_owner]
    assert_current("baseline", [before, *unchanged])
    capture.passed("normal Memory/Issue and incoming/outgoing ownership setup")

    def update(label, current, properties, source=None, stdin=None, unconditional=False):
        text = json.dumps(properties, ensure_ascii=False) if source is None else source
        guard = ["--unconditional"] if unconditional else ["--if-revision", current["revision"]]
        result = c0.envelope(capture.success(label, ["update", current["id"], "--properties", text,
                                                   *guard, "--actor", "memory-writer", "--json"], input_text=stdin))
        memory = save(label, result.get("memory"))
        c0.require(result.get("changed") is True, f"{label}: semantic change was reported as a no-op")
        c0.require(memory["id"] == current["id"] and memory["type"] == current["type"],
                   f"{label}: update changed identity or installed Type")
        c0.require(memory["revision"] != current["revision"] and memory["version"] != current["version"],
                   f"{label}: changed content failed to version Memory")
        c0.require(memory.get("properties") == properties and memory.get("owned") == current.get("owned"),
                   f"{label}: replacement lost properties or changed owned Links")
        c0.require(memory.get("attribution", {}).get("actor") == "memory-writer",
                   f"{label}: changed Memory lost explicit actor attribution")
        assert_current(label + "-fresh-process", [memory, *unchanged])
        exact(label + "-old-version", current)
        return memory

    edited = update("inline-memory-update", before, {"title": "Edited — 雪", "body": "Body two 😀 e\u0301\nline two"})
    changed_link = c0.envelope(capture.success("owned-link-update", [
        "update", "links/outgoing", "--properties", '{"note":"later Link context"}',
        "--if-revision", link_before["revision"], "--if-source-revision", edited["revision"], "--json"]))
    link_after = save("outgoing-link-updated", changed_link["link"])
    owner_after_link = save("edited-memory-after-link-update", changed_link["source"])
    c0.require(owner_after_link.get("properties") == edited["properties"] and
               owner_after_link.get("owned") == [link_after], "owned-Link update changed edited Memory content")
    unchanged = [link_after, target, incoming_link, incoming_owner]
    exact("old-body-and-old-link", before)
    exact("edited-body-and-old-link", edited)
    exact("original-memory-still-empty-owned", original)
    capture.passed("exact old Memory versions preserve body and owned-Link revisions after later Link updates")

    empty_properties = {"title": "", "body": ""}
    file_input = payload_file("empty-properties.json", json.dumps(empty_properties).encode())
    emptied = update("file-empty-memory-update", owner_after_link, empty_properties, source=file_input)
    final_properties = {"title": "Final mémoire", "body": "Final body — 雪\n😀"}
    final = update("stdin-unconditional-update", emptied, final_properties, source="@-",
                   stdin=json.dumps(final_properties, ensure_ascii=False), unconditional=True)
    noop = c0.envelope(capture.success("different-actor-noop", [
        "update", final["id"], "--properties", json.dumps(final_properties, ensure_ascii=False),
        "--if-revision", final["revision"], "--actor", "different-noop-actor", "--json"]))
    c0.require(noop.get("changed") is False and noop.get("memory") == final,
               "different-actor semantic no-op changed revision, attribution or content")
    capture.passed("inline/file/stdin support Unicode and empty strings; unconditional update and different-actor no-op work")

    good = json.dumps(final_properties, ensure_ascii=False)
    guard = ["--if-revision", final["revision"]]
    for label, props in [("stale-write", {"title": "must not persist", "body": "bad"}),
                         ("stale-noop", final_properties)]:
        c0.refusal(capture.run(label, ["update", final["id"], "--properties", json.dumps(props),
                                      "--if-revision", before["revision"], "--json"]), {"revision_conflict"}, label)
    capture.passed("stale revision refuses changes and semantic no-ops alike")

    bad_utf8 = payload_file("invalid-utf8.json", b'{"title":"bad\xff","body":""}')
    oversized = payload_file("oversized.json", json.dumps({"title": "", "body": "x" * INPUT_LIMIT}).encode())
    invalid = [
        ("missing-body", '{"title":"x"}'), ("missing-title", '{"body":"x"}'),
        ("null-body", '{"title":"x","body":null}'), ("null-title", '{"title":null,"body":"x"}'),
        ("extra-key", '{"title":"x","body":"y","gunk":true}'),
        ("number-title", '{"title":1,"body":"x"}'), ("array-body", '{"title":"x","body":[]}'),
        ("null-object", 'null'), ("array-object", '[]'), ("empty-object", '{}'),
        ("duplicate-key", '{"title":"one","title":"two","body":""}'),
        ("unpaired-surrogate", r'{"title":"\ud800","body":""}'),
        ("trailing-json", '{"title":"x","body":"y"}{}'),
        ("invalid-utf8", bad_utf8), ("oversized-input", oversized),
    ]
    for label, text in invalid:
        c0.refusal(capture.run(label, ["update", final["id"], "--properties", text, *guard, "--json"]),
                   {"invalid_properties"}, label)
    c0.refusal(capture.run("missing-properties", ["update", final["id"], *guard, "--json"]),
               {"invalid_properties"}, "missing properties")
    for label, flags in [
        ("missing-guard", []), ("empty-guard", ["--if-revision", ""]),
        ("conflicting-guard", [*guard, "--unconditional"]),
        ("false-unconditional", ["--unconditional=false"]),
    ]:
        c0.refusal(capture.run(label, ["update", final["id"], "--properties", good, *flags, "--json"]),
                   {"invalid_selector"}, label)
    for label, flags in [
        ("source-guard", ["--if-source-revision", incoming_owner["revision"]]),
        ("unconditional-source", ["--unconditional-source"]),
        ("legacy-title", ["--title", "must not persist"]),
        ("legacy-status", ["--status", "closed"]),
    ]:
        c0.refusal(capture.run(label, ["update", final["id"], "--properties", good, *guard, *flags, "--json"]),
                   {"capability_unavailable"}, label)
    c0.refusal(capture.run("healthy-issue-refusal", ["update", target["id"], "--properties", good,
                                                    "--unconditional", "--json"]),
               {"capability_unavailable"}, "healthy Issue kind")
    assert_current("after-input-and-guard-refusals", [final, *unchanged])
    capture.passed("invalid properties, Unicode, oversize, guards, legacy flags and Issue kind refuse without visible changes")

    yaml = capture.work / ".beads" / "config.yaml"
    original_yaml = yaml.read_bytes() if yaml.exists() else None
    denied = ["update", final["id"], "--properties", '{"title":"denied","body":"must not persist"}', *guard]
    for policy in ["flag", "environment", "configuration"]:
        try:
            flags = ["--readonly"] if policy == "flag" else []
            if policy == "environment":
                capture.env["BD_READONLY"] = "true"
            elif policy == "configuration":
                yaml.write_text("readonly: true\n")
            before_bytes = c0.tree_digest(capture.work)
            c0.refusal(capture.run("readonly-" + policy, [*denied, *flags, "--json"]),
                       {"permission_denied"}, "readonly " + policy)
            c0.require(c0.tree_digest(capture.work) == before_bytes, "readonly refusal changed workspace bytes")
        finally:
            capture.env.pop("BD_READONLY", None)
            if original_yaml is None:
                yaml.unlink(missing_ok=True)
            else:
                yaml.write_bytes(original_yaml)
    assert_current("after-readonly-refusals", [final, *unchanged])
    capture.passed("readonly flag, environment and configuration refuse before visible workspace changes")

    for entry in expected:
        exact(entry["label"] + "-final-exact-reopen", entry["record"])
    assert_current("after-all-exact-reopens", [final, *unchanged])
    capture.passed("all saved versions reopen exactly after every content/Link update and refusal")
    return {"saved_versions": len(expected), "saved_records_sha256": c0.sha256(capture.output / "expected-records.json"),
            "input_manifest_sha256": c0.sha256(capture.output / "inputs.json"),
            "final_memory_semantic_sha256": semantic_digest(final), "http_exercised": False}

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
