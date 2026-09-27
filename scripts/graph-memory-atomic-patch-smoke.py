#!/usr/bin/env python3
"""Installed atomic selected Memory patch proof with exact omitted-field preservation.

Normal authoring and fresh processes only. Embedded and caller-owned ordinary
Dolt run sequentially; no SQL, schema fixtures, mocks or hidden bootstrap.
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
SCOPE = "https://example.invalid/disposable-memory-atomic-patch/"
LIMIT = 1 << 20
LIMITATIONS = [
    "this suite exercises guarded and unconditional title-only, body-only and combined selected edits; omitted fields are preserved and title derivation is unavailable",
    "provisional CLI convenience, not complete Memory R4, aliases/upsert or full portable Memory fields",
    "recorded attribution and exact saved versions are not native/public History",
    "sequential stale guards are exercised; no forced transaction overlap, rollback or uncertain-commit injection here",
    "visible record equality does not prove internal database/event-count equality",
    "ordinary Dolt 2.1.8 database provisioning must remain serialized",
]

class RawStdin:
    def __init__(self, data):
        self.data = data

    def encode(self, encoding):
        c0.require(encoding == "utf-8", "unexpected Capture stdin encoding")
        return self.data

def semantic_digest(value):
    return hashlib.sha256(json.dumps(value, ensure_ascii=False, sort_keys=True,
                                     separators=(",", ":")).encode()).hexdigest()

def exercise(capture):
    backend = "server" if capture.args.server_port else "embedded"
    inputs = capture.output / "inputs"
    inputs.mkdir(mode=0o700)
    manifest, saved = [], []

    def file(name, data):
        path = inputs / name
        path.write_bytes(data)
        manifest.append({"path": str(path), "bytes": len(data), "sha256": c0.sha256(path)})
        c0.write_json(capture.output / "inputs.json", manifest)
        return str(path)

    body_bytes = "---\r\nnot: metadata\r\n---\r\n  Snow 雪 😀 e\u0301\r\ntrailing  ".encode()
    body_file = file("body.md", body_bytes)
    empty_file = file("empty.md", b"")
    bad_file = file("invalid.md", b"invalid\xff")
    large_file = file("oversize.md", b"x" * (LIMIT + 1))
    missing_file = str(inputs / "missing.md")

    def raw(label, args, stdin=None):
        receipt, _, _ = capture.run(label, args, input_text=None if stdin is None else RawStdin(stdin))
        artifact = capture.output / capture.records[-1]["artifact"]
        return receipt, (artifact / "stdout.log").read_bytes(), (artifact / "stderr.log").read_bytes()

    def command(label, args, stdin=None):
        receipt, out, err = raw(label, [*args, "--json"], stdin)
        c0.require(receipt["exit_code"] == 0, label + ": failed: " + err.decode("utf-8", "replace"))
        return c0.envelope(json.loads(out))

    def refuse(label, args, code, stdin=None):
        receipt, out, err = raw(label, [*args, "--json"], stdin)
        c0.require(receipt["exit_code"] != 0 and out == b"", label + ": refusal succeeded or leaked stdout")
        problem = json.loads(err)
        c0.require(problem.get("code") == code and problem.get("retryable") is False and "result" not in problem,
                   label + ": incorrect typed refusal: " + err.decode("utf-8", "replace"))

    def init_args(graph, prefix):
        args = ["init", "--prefix", prefix, "--non-interactive", "--skip-hooks", "--skip-agents", "--json"]
        if graph:
            args += ["--graph-mode", "link", "--scope-url", SCOPE]
        if capture.args.server_port:
            args += ["--server", "--external", "--server-host", "127.0.0.1", "--server-port",
                     str(capture.args.server_port), "--database", prefix + "_" + capture.root.name.replace("-", "_"),
                     "--server-user", "root"]
        return args

    before = c0.tree_digest(capture.work)
    refuse("legacy-before-init", ["remember", "--update", "beads/plan", "--unconditional", "--body-file", missing_file],
           "capability_unavailable")
    c0.require(c0.tree_digest(capture.work) == before, "legacy refusal opened storage or changed input")
    initialized = c0.envelope(capture.success("normal-graph-init", init_args(True, "apatch")))
    c0.require(initialized.get("scope") == SCOPE and initialized.get("backend") == backend, "wrong authority/backend")
    status = command("status", ["status", "--graph"])
    caps = status["capabilities"]
    c0.require(caps.get("memorySelectedUpdate") is True and caps.get("memorySelectedUpdateUnconditional") is True and
               caps.get("memory") is False and caps.get("historyExact") is False and
               status["limits"].get("memoryBodyInputBytes") == LIMIT, "incorrect patch/Memory/History capability boundary")

    def save(label, value):
        c0.require(value.get("id", "").startswith(SCOPE) and value.get("version") and value.get("attribution"),
                   label + ": missing record identity/version/attribution")
        saved.append({"label": label, "record": copy.deepcopy(value), "semantic_sha256": semantic_digest(value)})
        c0.write_json(capture.output / "expected-records.json", saved)
        return value

    def current(label, records):
        for index, value in enumerate(records):
            c0.require(command(label + "-" + str(index), ["show", value["id"], "--readonly"]) == value,
                       label + ": current/unrelated state changed")

    def exact(label, value):
        c0.require(command(label, ["show", value["id"], "--version", value["version"], "--readonly"]) == value,
                   label + ": complete retained state changed")
        if value["type"] == SCOPE + "types/preview-memory-v2":
            receipt, out, err = raw(label + "-recall", ["recall", value["id"], "--version", value["version"], "--readonly", "--quiet"])
            c0.require(receipt["exit_code"] == 0 and err == b"" and out == value["properties"]["body"].encode(),
                       label + ": exact recall changed body bytes")

    def compare(label, before, after):
        result = command(label, ["compare", before["id"], "--from", before["version"], "--to", after["version"], "--readonly"])
        c0.require(result["resource"] == {"id": before["id"], "type": before["type"]} and
                   result["from"] == {"version": before["version"], "attribution": before["attribution"]} and
                   result["to"] == {"version": after["version"], "attribution": after["attribution"]},
                   label + ": wrong comparison context")
        changes = []
        for member in ["body", "title"]:
            if before["properties"][member] != after["properties"][member]:
                changes.append({"area": "properties", "member": member,
                    "from": {"present": True, "value": before["properties"][member]},
                    "to": {"present": True, "value": after["properties"][member]}})
        c0.require(before["owned"] == after["owned"] and result["changes"] == changes,
                   label + ": complete property/owned comparison differs")

    memory = save("created-memory", command("create-memory", ["remember", "Original\r\n  雪 😀  ",
        "--id", "beads/plan", "--title", "  Original title 雪  ", "--actor", "original-author"]))
    target = save("target", command("create-target", ["create", "Untouched target", "--id", "beads/work"]))
    relation = command("create-owned-link", ["link", memory["id"], target["id"], "--resource-type",
        SCOPE + "types/preview-related-v2", "--id", "links/context", "--properties", '{"note":"context"}',
        "--if-source-revision", memory["revision"], "--actor", "context-author"])
    memory, link = save("memory-owned", relation["source"]), save("link-original", relation["link"])

    def patched(label, before, flags, wanted, unconditional=False, stdin=None, changed=True):
        guard = ["--unconditional"] if unconditional else ["--if-revision", before["revision"]]
        result = command(label, ["remember", "--update", before["id"], *guard, "--actor", label + "-actor", *flags], stdin)
        expected_keys = {"memory", "changed", "replaced"} if unconditional and changed else {"memory", "changed"}
        c0.require(set(result) == expected_keys and result["changed"] is changed, label + ": wrong result/disclosure shape")
        if unconditional and changed:
            c0.require(result["replaced"] == {key: before[key] for key in ["id", "version", "attribution"]},
                       label + ": disclosure is not the actual checked predecessor")
        after = result["memory"]
        if not changed:
            c0.require(after == before, label + ": no-op minted state or attribution")
            return before
        expected = copy.deepcopy(before)
        expected["properties"] = wanted
        expected["revision"] = after["revision"]
        expected["version"] = after["version"]
        expected["attribution"] = after["attribution"]
        c0.require(after == expected and after["revision"] == after["version"] and after["version"] != before["version"] and
                   after["attribution"]["actor"] == label + "-actor", label + ": unexpected complete accepted record")
        after = save(label, after)
        current(label + "-new-process", [after, link, target])
        compare(label + "-compare", before, after)
        return after

    def props(title=None, body=None):
        result = copy.deepcopy(memory["properties"])
        if title is not None:
            result["title"] = title
        if body is not None:
            result["body"] = body
        return result

    memory = patched("guarded-title-only", memory, ["--title", "Guarded title 雪"], props(title="Guarded title 雪"),
                     stdin=b"must not become implicit body\xff")
    memory = patched("unconditional-title-only", memory, ["--title", "Unconditional title"], props(title="Unconditional title"),
                     unconditional=True, stdin=b"ignored unrequested input\xff")
    memory = patched("guarded-body-only", memory, ["--body-file", body_file], props(body=body_bytes.decode()))
    stdin_bytes = b"  stdin\r\nkeeps spaces  \n"
    memory = patched("unconditional-body-only", memory, ["--stdin"], props(body=stdin_bytes.decode()),
                     unconditional=True, stdin=stdin_bytes)
    memory = patched("explicit-both", memory, ["--title", "Both fields", "singleword"],
                     props(title="Both fields", body="singleword"))
    memory = patched("clear-title-only", memory, ["--title", ""], props(title=""), unconditional=True)
    memory = patched("clear-body-file", memory, ["--body-file", empty_file], props(body=""))
    memory = patched("empty-stdin-noop", memory, ["--stdin"], props(body=""), unconditional=True, stdin=b"", changed=False)
    memory = patched("empty-positional-noop", memory, [""], props(body=""), changed=False)
    memory = patched("empty-title-noop", memory, ["--title", ""], props(title=""), unconditional=True, changed=False)
    capture.passed("guarded/unconditional title-only and body-only patches preserve omitted bytes and owned Links; explicit empties and different-actor no-ops are exact")

    # Independent authoring precedes the patch: this proves current predecessor
    # disclosure without claiming forced overlapping transaction execution.
    stale = memory
    prior = command("separate-full-title-write", ["update", memory["id"], "--properties",
        json.dumps({"title": "Independent title", "body": memory["properties"]["body"]}),
        "--if-revision", memory["revision"], "--actor", "separate-title-writer"])
    memory = save("independent-title", prior["memory"])
    refuse("stale-title-patch", ["remember", "--update", memory["id"], "--title", "rejected",
        "--if-revision", stale["revision"]], "revision_conflict")
    memory = patched("unconditional-after-title", memory, ["--body-file", body_file], props(body=body_bytes.decode()),
                     unconditional=True)
    stale = memory
    relation = command("separate-link-write", ["update", link["id"], "--properties", '{"note":"new context"}',
        "--if-revision", link["revision"], "--if-source-revision", memory["revision"], "--actor", "separate-link-writer"])
    memory, link = save("independent-owned", relation["source"]), save("link-updated", relation["link"])
    refuse("stale-owned-noop", ["remember", "--update", memory["id"], "--title", memory["properties"]["title"],
        "--if-revision", stale["revision"]], "revision_conflict")
    memory = patched("unconditional-after-owned", memory, ["--title", "Human title"], props(title="Human title"),
                     unconditional=True)
    capture.passed("unconditional disclosure names actual current predecessor after independent title/owned-Link changes; guarded stale edits and stale no-ops refuse")

    for label, unconditional, quiet, wanted in [
        ("human-unconditional-change", True, False, "Human body"),
        ("human-unconditional-noop", True, False, "Human body"),
        ("human-guarded-change", False, False, "Guarded body"),
        ("quiet-unconditional-change", True, True, "Quiet body"),
        ("quiet-unconditional-noop", True, True, "Quiet body"),
    ]:
        before = memory
        guard = ["--unconditional"] if unconditional else ["--if-revision", before["revision"]]
        args = ["remember", wanted, "--update", before["id"], *guard, "--actor", label + "-actor"]
        if quiet:
            args += ["--quiet"]
        receipt, out, err = raw(label, args)
        c0.require(receipt["exit_code"] == 0 and err == b"", label + ": human/quiet patch failed")
        memory = command(label + "-show", ["show", before["id"], "--readonly"])
        changed = before["properties"]["body"] != wanted
        if changed:
            expected = copy.deepcopy(before)
            expected["properties"]["body"] = wanted
            for key in ["version", "revision", "attribution"]:
                expected[key] = memory[key]
            c0.require(memory == expected and memory["version"] != before["version"] and
                       memory["revision"] == memory["version"] and memory["attribution"]["actor"] == label + "-actor",
                       label + ": wrong complete state")
            save(label, memory)
        else:
            c0.require(memory == before, label + ": no-op minted a record")
        expected = ("Updated" if changed else "Unchanged") + " " + memory["id"] + ': "Human title"'
        if unconditional and changed:
            attribution = before["attribution"]
            expected += ("\nReplaced Memory " + before["id"] + " version " + before["version"] +
                "; recorded attribution: actor=" + json.dumps(attribution.get("actor", "")) +
                " status=" + json.dumps(attribution["status"]) + " recordedAt=" + json.dumps(attribution.get("recordedAt", "")))
        c0.require(out == (b"" if quiet else (expected + "\n").encode()), label + ": wrong human/quiet disclosure")
        current(label + "-unrelated", [link, target])
    capture.passed("human changed unconditional writes disclose actual predecessor; guarded/no-op omit it and quiet suppresses output")

    good = ["--update", memory["id"], "--if-revision", memory["revision"]]
    invalid = [
        ("missing-guard", ["--update", memory["id"], "--title", "x"], None, "invalid_selector"),
        ("empty-guard", ["--update", memory["id"], "--if-revision", "", "--title", "x"], None, "invalid_selector"),
        ("both-guards", [*good, "--unconditional", "--title", "x"], None, "invalid_selector"),
        ("false-guard", ["--update", memory["id"], "--unconditional=false", "--title", "x"], None, "invalid_selector"),
        ("missing-fields", good, None, "invalid_properties"),
        ("unrequested-pipe", good, b"not an implicit body", "invalid_properties"),
        ("two-positionals", [*good, "one", "two"], None, "invalid_properties"),
        ("file-positional", [*good, "--body-file", body_file, "body"], None, "invalid_properties"),
        ("file-stdin", [*good, "--body-file", body_file, "--stdin"], b"ignored", "invalid_properties"),
        ("false-stdin", [*good, "--stdin=false", "--title", "x"], None, "invalid_properties"),
        ("invalid-file", [*good, "--body-file", bad_file], None, "invalid_properties"),
        ("invalid-stdin", [*good, "--stdin"], b"invalid\xff", "invalid_properties"),
        ("oversize-file", [*good, "--body-file", large_file], None, "capability_unavailable"),
        ("missing-file", [*good, "--body-file", missing_file], None, "invalid_properties"),
        ("key", [*good, "--key", "legacy", "--body-file", missing_file], None, "capability_unavailable"),
        ("create-guard", ["body", "--id", "beads/refused", "--title", "x", "--if-revision", "x"], None, "capability_unavailable"),
        ("create-unconditional", ["body", "--id", "beads/refused", "--title", "x", "--unconditional"], None, "capability_unavailable"),
        ("issue-kind", ["--update", target["id"], "--if-revision", target["revision"], "--title", "x"], None, "capability_unavailable"),
    ]
    for label, flags, stdin, code in invalid:
        refuse(label, ["remember", *flags], code, stdin)
    refuse("refused-identity-absent", ["show", "beads/refused"], "not_found")
    current("after-refusals", [memory, link, target])
    capture.passed("invalid guards, omitted fields, conflicting/invalid inputs, key, creation guards and Issue kind refuse with zero success output")

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
                freeze.write_text("atomic-patch\t2026-09-27T00:00:00Z\tinput policy\n")
            before = c0.tree_digest(capture.work)
            refuse("policy-" + policy, ["remember", "--update", memory["id"], "--unconditional",
                   "--body-file", missing_file, *flags], "permission_denied")
            c0.require(c0.tree_digest(capture.work) == before, policy + ": refused input changed workspace")
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
    current("after-policy", [memory, link, target])
    for entry in saved:
        exact(entry["label"] + "-retained", entry["record"])
    capture.passed("read-only/freeze admission precedes file acquisition; all saved complete records remain exact in fresh processes")

    graph_work = capture.work
    capture.work = capture.root / "legacy-workspace"
    capture.work.mkdir(mode=0o700)
    try:
        receipt, _, err = capture.run("normal-legacy-init", init_args(False, "latomic"))
        c0.require(receipt["exit_code"] == 0, "legacy init failed: " + err)
        legacy_body = "Legacy keyed Memory remains intact"
        legacy = capture.success("legacy-keyed-write", ["remember", legacy_body, "--key", "legacy-plan", "--json"])
        c0.require(legacy["key"] == "legacy-plan" and legacy["value"] == legacy_body, "legacy authoring changed")
        refuse("legacy-unconditional", ["remember", "--update", "beads/plan", "--unconditional",
               "--body-file", missing_file], "capability_unavailable")
        c0.require(capture.success("legacy-unchanged", ["memories", "--json"]) ==
                   {"legacy-plan": legacy_body, "schema_version": 1}, "new graph flags changed legacy state")
    finally:
        capture.work = graph_work
    capture.passed("legacy authoring remains available and refuses graph unconditional partial update before missing-file input")
    return {"saved_versions": len(saved), "input_files": len(manifest), "http_exercised": False,
            "expected_records_sha256": c0.sha256(capture.output / "expected-records.json"),
            "input_manifest_sha256": c0.sha256(capture.output / "inputs.json")}


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
