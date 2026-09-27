#!/usr/bin/env python3
"""Installed selected, guarded remember update proof with exact body input.

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
SCOPE = "https://example.invalid/disposable-memory-selected-update/"
LIMIT = 1 << 20
LIMITATIONS = [
    "selected remember update requires an observed revision and explicit body; no unconditional partial update or title derivation",
    "provisional CLI convenience, not complete Memory R4, aliases/upsert or full portable Memory fields",
    "recorded attribution and exact saved versions are not native/public History",
    "sequential stale guards are exercised; no forced composition/write interleave, rollback or uncertain-commit injection here",
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

    file_body = ("---\r\ntitle: not a title directive\r\n---\r\n  newtoken 雪 😀 e\u0301\r\ntrailing  ").encode()
    stdin_body = ("---\r\nsource: stdin, not metadata\r\n---\r\n  stdin 雪 😀  \n").encode()
    body_file = file("body.md", file_body)
    empty_file = file("empty.md", b"")
    bad_file = file("bad-utf8.md", b"before\xffafter")
    boundary_file = file("boundary.md", b"b" * LIMIT)
    large_file = file("oversize.md", b"x" * (LIMIT + 1))
    missing = str(inputs / "missing.md")

    def raw(label, args, stdin=None):
        receipt, _, _ = capture.run(label, args, input_text=None if stdin is None else RawStdin(stdin))
        artifact = capture.output / capture.records[-1]["artifact"]
        return receipt, (artifact / "stdout.log").read_bytes(), (artifact / "stderr.log").read_bytes()

    def command(label, args, stdin=None):
        receipt, out, err = raw(label, [*args, "--json"], stdin)
        c0.require(receipt["exit_code"] == 0, label + ": command failed: " + err.decode("utf-8", "replace"))
        return c0.envelope(json.loads(out))

    def refuse(label, args, code, stdin=None):
        receipt, out, err = raw(label, [*args, "--json"], stdin)
        c0.require(receipt["exit_code"] != 0 and out == b"", label + ": refusal succeeded or leaked stdout")
        if code is None:
            c0.require(bool(err), label + ": missing diagnostic")
            return
        problem = json.loads(err)
        c0.require(problem.get("code") == code and problem.get("retryable") is False and "result" not in problem,
                   label + ": wrong typed refusal")

    # The new graph selection/guard flags must fail before opening an input
    # file or any legacy store, even in a directory without a workspace.
    before = c0.tree_digest(capture.work)
    for label, flags in [
        ("legacy-update-before-init", ["--update", "beads/plan"]),
        ("legacy-guard-before-init", ["--if-revision", "observed"]),
        ("legacy-selected-before-init", ["--update", "beads/plan", "--if-revision", "observed"]),
    ]:
        refuse(label, ["remember", *flags, "--body-file", missing], "capability_unavailable")
    c0.require(c0.tree_digest(capture.work) == before, "legacy admission opened storage or input")
    capture.passed("selected-update flags refuse in an uninitialized legacy route before missing-file input")

    def init_args(graph, prefix):
        args = ["init", "--prefix", prefix, "--non-interactive", "--skip-hooks", "--skip-agents", "--json"]
        if graph:
            args += ["--graph-mode", "link", "--scope-url", SCOPE]
        if capture.args.server_port:
            args += ["--server", "--external", "--server-host", "127.0.0.1", "--server-port",
                     str(capture.args.server_port), "--database", prefix + "_" + capture.root.name.replace("-", "_"),
                     "--server-user", "root"]
        return args

    initialized = c0.envelope(capture.success("normal-graph-init", init_args(True, "selected")))
    c0.require(initialized.get("scope") == SCOPE and initialized.get("backend") == backend, "wrong authority/backend")
    status = command("status", ["status", "--graph"])
    c0.require(status["capabilities"].get("memorySelectedUpdate") is True and
               status["capabilities"].get("memorySelectedUpdateUnconditional") is False and
               status["capabilities"].get("memory") is False and status["capabilities"].get("historyExact") is False and
               status["limits"].get("memoryBodyInputBytes") == LIMIT, "authoring overstates selected update/Memory/History or input limit")

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

    def discover(label, query, memory):
        result = c0.envelope(capture.success(label, ["memories", query, "--details", "--format", "records-json", "--readonly"]))
        c0.require(result.get("projection") == "summary" and result.get("complete") is True and
                   "next" in result and result["next"] is None and len(result["items"]) == 1, label + ": incomplete discovery")
        item = result["items"][0]
        c0.require(all(item[key] == memory[key] for key in ["id", "type", "version", "attribution"]) and
                   item["title"] == memory["properties"]["title"] and
                   item["details"] == {"ownedLinkCount": len(memory["owned"])}, label + ": discovery selected wrong state")
        return item

    title = "  Distinctive original title — 雪  "
    memory = save("created-memory", command("create-memory",
        ["remember", "Original body\r\n  雪 😀  ", "--id", "beads/plan", "--title", title, "--actor", "original-author"]))
    c0.require(memory["properties"]["title"] == title, "creation changed distinctive title")
    target = save("issue-target", command("create-target", ["create", "Target stays unchanged", "--id", "beads/work"]))
    relation = command("owned-link", ["link", memory["id"], target["id"], "--resource-type",
        SCOPE + "types/preview-related-v2", "--id", "links/context", "--properties", '{"note":"first context"}',
        "--if-source-revision", memory["revision"], "--actor", "context-author"])
    memory, link = save("memory-with-link", relation["source"]), save("context-original", relation["link"])
    selected = discover("discover-before-edit", "Distinctive", memory)
    exact("selected-before-edit", memory)
    c0.require(selected["id"] == memory["id"] and selected["version"] == memory["revision"], "discovery guard differs")

    def selected_update(label, before, flags, wanted, new_title=None, stdin=None, selector=None, changed=True):
        args = ["remember", "--update", selector or before["id"], "--if-revision", before["revision"],
                "--actor", label + "-actor", *flags]
        if new_title is not None:
            args += ["--title", new_title]
        result = command(label, args, stdin)
        c0.require(set(result) == {"memory", "changed"} and result["changed"] is changed,
                   label + ": wrong mutation result or invented replacement disclosure")
        after = result["memory"]
        if not changed:
            c0.require(after == before, label + ": no-op minted state/attribution")
            return before
        after = save(label, after)
        c0.require(after["id"] == before["id"] and after["type"] == before["type"] and
                   after["revision"] == after["version"] and after["version"] != before["version"] and
                   after["owned"] == before["owned"] and after["properties"] == {
                       "title": before["properties"]["title"] if new_title is None else new_title,
                       "body": wanted.decode("utf-8")}, label + ": changed omitted title, body bytes or complete owned state")
        c0.require(after["attribution"]["actor"] == label + "-actor", label + ": wrong accepted actor")
        current(label + "-fresh", [after])
        compare(label + "-compare", before, after)
        return after

    before_file = memory
    memory = selected_update("selected-file-update", memory, ["--body-file", body_file], file_body,
                             selector="beads/plan")
    discover("discover-new-body", "newtoken", memory)
    memory = selected_update("selected-noop-different-actor", memory, ["--body-file", body_file], file_body, changed=False)
    for label, source in [("stale-change", ["must refuse"]), ("stale-noop", ["--body-file", body_file])]:
        refuse(label, ["remember", *source, "--update", memory["id"], "--if-revision", before_file["revision"]], "revision_conflict")
    current("after-file-and-stale", [memory, link, target])
    exact("file-update-old-selected", before_file)
    capture.passed("discovery-selected file update preserves exact title/body/owned state; no-op and stale edit/no-op have correct guards")

    before_title = memory
    title_properties = {"title": "Changed by another writer — 雪", "body": memory["properties"]["body"]}
    title_edit = command("independent-title-update", ["update", memory["id"], "--properties", json.dumps(title_properties),
        "--if-revision", memory["revision"], "--actor", "title-writer"])
    memory = save("title-winner", title_edit["memory"])
    refuse("stale-after-title-update", ["remember", "--body-file", body_file, "--update", memory["id"],
                                     "--if-revision", before_title["revision"]], "revision_conflict")
    before_link = memory
    link_edit = command("independent-owned-link-update", ["update", link["id"], "--properties", '{"note":"new context"}',
        "--if-revision", link["revision"], "--if-source-revision", memory["revision"], "--actor", "link-writer"])
    memory, link = save("owned-winner", link_edit["source"]), save("context-updated", link_edit["link"])
    refuse("stale-after-owned-update", ["remember", "--body-file", body_file, "--update", memory["id"],
                                     "--if-revision", before_link["revision"]], "revision_conflict")
    current("winner-states-preserved", [memory, link, target])
    capture.passed("selected update retains caller revision and refuses after independent title or owned-Link changes")

    memory = selected_update("one-token-body-explicit-title", memory, ["singleword"], b"singleword",
                             new_title="  Explicit replacement title  ")
    memory = selected_update("stdin-exact-body", memory, ["--stdin"], stdin_body, stdin=stdin_body)
    memory = selected_update("exact-limit-body", memory, ["--body-file", boundary_file], b"b" * LIMIT)
    memory = selected_update("empty-file-body", memory, ["--body-file", empty_file], b"")
    memory = selected_update("empty-stdin-and-title", memory, ["--stdin"], b"", new_title="", stdin=b"")
    memory = selected_update("empty-positional-noop", memory, [""], b"", changed=False)
    current("after-body-sources", [memory, link, target])
    capture.passed("positional one-token writes and file/stdin/explicit empties preserve source bytes and omitted fields")

    for label, quiet, wanted, new_title in [
        ("human-changed", False, b"Human body", "Human-visible title"),
        ("human-noop", False, b"Human body", None),
        ("quiet-changed", True, b"Quiet changed body", None),
        ("quiet-noop", True, b"Quiet changed body", None),
    ]:
        before = memory
        args = ["remember", wanted.decode(), "--update", memory["id"], "--if-revision", memory["revision"],
                "--actor", label + "-actor"]
        if new_title is not None:
            args += ["--title", new_title]
        if quiet:
            args.append("--quiet")
        receipt, out, err = raw(label, args)
        c0.require(receipt["exit_code"] == 0 and err == b"", label + ": human/quiet update failed")
        changed = label.endswith("-changed")
        memory = command(label + "-current", ["show", before["id"], "--readonly"])
        if changed:
            memory = save(label, memory)
            c0.require(memory["version"] != before["version"] and memory["owned"] == before["owned"] and
                       memory["properties"] == {"title": before["properties"]["title"] if new_title is None else new_title,
                                                "body": wanted.decode()}, label + ": wrong accepted state")
        else:
            c0.require(memory == before, label + ": no-op changed state")
        if quiet:
            c0.require(out == b"", label + ": quiet output was not empty")
        else:
            text = out.decode()
            expected = ("Updated" if changed else "Unchanged") + " " + memory["id"] + ': "Human-visible title"\n'
            c0.require(text == expected, label + ": human output differs from the complete expected line")
            c0.require(memory["id"] in text and memory["properties"]["title"] in text and
                       ("Updated" if changed else "Unchanged") in text and "Replaced Memory" not in text,
                       label + ": human result omitted identity/title or misrepresented guard/disclosure")
    capture.passed("human output distinguishes changed/unchanged and names identity/title; quiet retains successful mutation semantics")

    good_guard = ["--update", memory["id"], "--if-revision", memory["revision"]]
    invalid = [
        ("missing-source", [*good_guard], None, "invalid_properties"),
        ("pipe-without-stdin", [*good_guard], b"must not silently become a body", "invalid_properties"),
        ("two-positional", ["one", "two", *good_guard], None, "invalid_properties"),
        ("file-plus-positional", ["body", "--body-file", body_file, *good_guard], None, "invalid_properties"),
        ("stdin-plus-positional", ["body", "--stdin", *good_guard], b"ignored", "invalid_properties"),
        ("file-plus-stdin", ["--body-file", body_file, "--stdin", *good_guard], b"ignored", "invalid_properties"),
        ("false-stdin", ["--stdin=false", *good_guard], None, "invalid_properties"),
        ("empty-file-name", ["--body-file", "", *good_guard], None, "invalid_properties"),
        ("missing-file", ["--body-file", missing, *good_guard], None, "invalid_properties"),
        ("invalid-file-utf8", ["--body-file", bad_file, *good_guard], None, "invalid_properties"),
        ("invalid-stdin-utf8", ["--stdin", *good_guard], b"bad\xff", "invalid_properties"),
        ("oversize-file", ["--body-file", large_file, *good_guard], None, "capability_unavailable"),
        ("oversize-stdin", ["--stdin", *good_guard], b"x" * (LIMIT + 1), "capability_unavailable"),
        ("id-update-conflict", ["--body-file", missing, "--id", "beads/other", *good_guard], None, "invalid_selector"),
        ("empty-update", ["body", "--update", "", "--if-revision", memory["revision"]], None, "invalid_selector"),
        ("malformed-update", ["body", "--update", "beads/../plan", "--if-revision", memory["revision"]], None, "invalid_selector"),
        ("foreign-update", ["body", "--update", "https://other.invalid/beads/plan", "--if-revision", memory["revision"]], None, "invalid_selector"),
        ("missing-memory", ["body", "--update", "beads/missing", "--if-revision", memory["revision"]], None, "not_found"),
        ("issue-kind", ["body", "--update", target["id"], "--if-revision", target["revision"]], None, "capability_unavailable"),
        ("link-kind", ["body", "--update", link["id"], "--if-revision", link["revision"]], None, "invalid_selector"),
        ("missing-guard", ["--body-file", missing, "--update", memory["id"]], None, "invalid_selector"),
        ("empty-guard", ["--body-file", missing, "--update", memory["id"], "--if-revision", ""], None, "invalid_selector"),
        ("oversized-guard-before-input", ["--body-file", missing, "--update", memory["id"], "--if-revision", "x" * 4097], None, "invalid_selector"),
        ("guard-without-selection", ["--body-file", missing, "--id", "beads/other", "--title", "Denied",
                                    "--if-revision", memory["revision"]], None, "capability_unavailable"),
        ("unsupported-unconditional", ["--body-file", missing, "--update", memory["id"], "--unconditional"], None, None),
        ("unsupported-key", ["--body-file", missing, *good_guard, "--key", "legacy"], None, "capability_unavailable"),
    ]
    for label, flags, stdin, code in invalid:
        refuse(label, ["remember", *flags], code, stdin)
    for label, flags in [("unsupported-model", ["--model", "none"]), ("unsupported-metadata", ["--metadata", "{}"])]:
        refuse(label, ["remember", "body", *good_guard, *flags], None)
    refuse("missing-remains-absent", ["show", "beads/missing"], "not_found")
    refuse("conflicting-id-not-created", ["show", "beads/other"], "not_found")
    current("after-input-refusals", [memory, link, target])
    capture.passed("missing/conflicting/invalid/oversize sources, selectors and unsupported contracts refuse without output or creation")

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
                freeze.write_text("selected-update\t2026-09-27T00:00:00Z\tinput policy\n")
            before = c0.tree_digest(capture.work)
            refuse("policy-" + policy, ["remember", "--body-file", missing, *good_guard, *flags], "permission_denied")
            c0.require(c0.tree_digest(capture.work) == before, policy + ": input refusal changed workspace")
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
        exact(entry["label"] + "-final-retained", entry["record"])
    capture.passed("readonly/freeze refuse before missing-file acquisition; every saved body and complete owned state remains exact")

    graph_work = capture.work
    capture.work = capture.root / "legacy-workspace"
    capture.work.mkdir(mode=0o700)
    try:
        receipt, _, err = capture.run("normal-legacy-init", init_args(False, "lselected"))
        c0.require(receipt["exit_code"] == 0, "legacy init failed: " + err)
        legacy_body = "Existing keyed legacy behavior stays intact."
        remembered = capture.success("legacy-keyed-write", ["remember", legacy_body, "--key", "legacy-plan", "--json"])
        c0.require(remembered["key"] == "legacy-plan" and remembered["value"] == legacy_body, "legacy write changed")
        recalled = capture.success("legacy-bare-key-read", ["remember", "legacy-plan", "--json"])
        c0.require(recalled["key"] == "legacy-plan" and recalled["value"] == legacy_body and
                   recalled["action"] == "recalled", "bare legacy key no longer recalls")
        refuse("legacy-selected-after-init", ["remember", "--update", "beads/plan", "--if-revision", "observed",
                                             "--body-file", missing], "capability_unavailable")
        c0.require(capture.success("legacy-map-unchanged", ["memories", "--json"]) ==
                   {"legacy-plan": legacy_body, "schema_version": 1}, "selected flags altered legacy state")
    finally:
        capture.work = graph_work
    capture.passed("ordinary legacy keyed write and bare-key recall remain unchanged; selected update stays graph-only")
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
