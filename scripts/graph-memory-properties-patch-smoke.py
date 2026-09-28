#!/usr/bin/env python3
"""Installed ordered Memory property-change proof against normal initialized storage.

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
SCOPE = "https://example.invalid/disposable-memory-properties-patch/"
LIMIT = 1 << 20
LIMITATIONS = [
    "private ordered Memory patch ends in title/body strings; not generic Resource or BDP HTTP Write",
    "current/exact records and saved comparisons are not native/public History",
    "internal transaction counts, forced concurrency and lost acknowledgments are qualified by separate real-engine tests",
    "ordinary Dolt2.1.8 different-database provisioning stays serialized",
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
    saved, inputs = [], []
    folder = capture.output / "inputs"
    folder.mkdir(mode=0o700)

    def file(name, data):
        path = folder / name
        path.write_bytes(data)
        inputs.append({"path": str(path), "bytes": len(data), "sha256": c0.sha256(path)})
        c0.write_json(capture.output / "inputs.json", inputs)
        return str(path)

    def command(label, argv, stdin=None):
        receipt, out, err = capture.run(label, argv if "--format" in argv else [*argv, "--json"], input_text=stdin)
        c0.require(receipt["exit_code"] == 0, label + ": " + err)
        return c0.envelope(json.loads(out))

    def refuse(label, argv, code):
        receipt, out, err = capture.run(label, [*argv, "--json"])
        c0.require(receipt["exit_code"] != 0 and out == "", label + ": refusal/output")
        problem = json.loads(err)
        c0.require(problem.get("code") == code and problem.get("retryable") is False and "result" not in problem,
                   label + ": typed error differs")

    def save(label, record):
        c0.require(record["version"] == record["revision"] and record["version"] and record["id"].startswith(SCOPE), label + ": identity")
        saved.append({"label": label, "record": copy.deepcopy(record), "semantic_sha256": semantic_digest(record)})
        c0.write_json(capture.output / "expected-records.json", saved)
        return record

    def current(label, record):
        c0.require(command(label, ["show", record["id"], "--readonly"]) == record, label + ": current differs")

    missing = str(folder / "missing.json")
    refuse("legacy-before-input", ["update", "beads/plan", "--patch", "@"+missing, "--unconditional"], "capability_unavailable")
    init = ["init", "--prefix", "mpatch", "--non-interactive", "--skip-hooks", "--skip-agents", "--graph-mode", "link", "--scope-url", SCOPE]
    if capture.args.server_port:
        init += ["--server", "--external", "--server-host", "127.0.0.1", "--server-port", str(capture.args.server_port),
                 "--database", "mpatch_"+capture.root.name.replace("-", "_"), "--server-user", "root"]
    initialized = command("normal-init", init)
    c0.require(initialized.get("scope") == SCOPE and initialized.get("backend") == backend, "initialization authority")
    status = command("patch-capabilities", ["status", "--graph"])
    caps = status["capabilities"]
    c0.require(caps.get("memoryPropertiesPatch") is True and caps.get("memory") is False and
               caps.get("historyExact") is False and caps.get("requestStatus") is False, "overstated capability")
    for name, expected in {"memoryPatchInputBytes":LIMIT, "memoryPatchDocumentBytes":LIMIT,
                           "memoryPatchOperations":256, "memoryPatchPointerBytes":4096,
                           "memoryPatchPointerSegments":64, "memoryPatchDepth":64, "memoryPatchEvaluationBytes":16<<20}.items():
        c0.require(status["limits"].get(name) == expected, "incorrect advertised patch limit: "+name)
    memory = save("initial", command("create-memory", ["remember", "Original body", "--id", "beads/plan", "--title", "Original title", "--actor", "creator"]))
    target = save("target", command("create-target", ["create", "Issue target", "--id", "beads/work"]))
    link_result = command("link-memory-to-issue", ["link", memory["id"], target["id"], "--id", "links/out", "--resource-type", SCOPE+"types/preview-related-v2", "--properties", '{"note":"keep"}', "--if-source-revision", memory["revision"]])
    memory, link = save("owned", link_result["source"]), save("link", link_result["link"])
    initial = memory

    def patch(label, before, operations, wanted, *, unconditional=False, source="literal", changed=True):
        text = json.dumps(operations, ensure_ascii=False)
        value, stdin = text, None
        if source == "file": value = "@"+file(label+".json", text.encode())
        if source == "stdin": value, stdin = "@-", text
        argv = ["update", before["id"], "--patch", value, "--actor", "patch-author"]
        argv += ["--unconditional"] if unconditional else ["--if-revision", before["revision"]]
        result = command(label, argv, stdin)
        if not changed:
            c0.require(result == {"memory": before, "changed": False}, label + ": no-op differs")
            return before
        after = result["memory"]
        expected = copy.deepcopy(before)
        expected["properties"] = wanted
        for key in ("revision", "version", "attribution"): expected[key] = after[key]
        keys = {"memory", "changed", "replaced"} if unconditional else {"memory", "changed"}
        c0.require(set(result) == keys and result["changed"] is True and after == expected and after["version"] != before["version"] and after["attribution"]["actor"] == "patch-author", label + ": complete mutation differs")
        if unconditional:
            c0.require(result["replaced"] == {"id": before["id"], "version": before["version"], "attribution": before["attribution"]}, label + ": wrong predecessor disclosure")
        save(label, after)
        current(label+"-fresh", after)
        return after

    props = {"title": "Patch title 雪", "body": "  patchtoken body\r\n😀  "}
    memory = patch("literal-two-fields", memory, [{"op":"replace", "path":"/title", "value":props["title"]}, {"op":"replace", "path":"/body", "value":props["body"]}], props)
    refuse("empty-operations-refused", ["update", memory["id"], "--patch", "[]", "--unconditional"], "invalid_properties")
    patch("same-value-noop", memory, [{"op":"replace","path":"/title","value":props["title"]}], props, changed=False)
    patch("unconditional-reversing-noop", memory, [{"op":"replace", "path":"/title", "value":"temporary"}, {"op":"replace", "path":"/title", "value":props["title"]}], props, unconditional=True, changed=False)
    refuse("stale-noop", ["update", memory["id"], "--patch", json.dumps([{"op":"replace","path":"/title","value":props["title"]}]), "--if-revision", initial["revision"]], "revision_conflict")
    compare = command("compare-two-properties", ["compare", memory["id"], "--from", initial["version"], "--to", memory["version"], "--readonly"])
    c0.require(compare["changes"] == [{"area":"properties", "member":key, "from":{"present":True,"value":initial["properties"][key]}, "to":{"present":True,"value":props[key]}} for key in ["body","title"]], "comparison differs")
    discovery = command("discover-patched-body", ["memories", "patchtoken", "--details", "--format", "records-json", "--readonly"])
    c0.require(discovery["complete"] and len(discovery["items"]) == 1 and discovery["items"][0]["version"] == memory["version"], "discovery differs")
    capture.passed("guarded atomic two-field patch, same-value/reversing no-op, stale refusal, compare and discovery")

    props = {**props, "body":""}
    memory = patch("file-remove-add-clear", memory, [{"op":"remove","path":"/body"},{"op":"add","path":"/body","value":""}], props, source="file")
    props = {"title":"From stdin", "body":"stdin 雪\r\n  "}
    memory = patch("stdin-root-replacement", memory, [{"op":"replace","path":"","value":props}], props, source="stdin", unconditional=True)
    patch("temporary-array-and-special-key", memory, [{"op":"add","path":"/__proto__","value":[1,2]},{"op":"add","path":"/__proto__/1","value":3},{"op":"remove","path":"/__proto__/0"},{"op":"add","path":"/__proto__/-","value":4},{"op":"remove","path":"/__proto__"}], props, changed=False)
    before_link = memory
    link_result = command("owned-link-change", ["update", link["id"], "--properties", '{"note":"updated"}', "--if-revision", link["revision"], "--if-source-revision", memory["revision"], "--actor", "link-author"])
    memory, link = save("owned-link-source", link_result["source"]), save("updated-link", link_result["link"])
    refuse("owned-link-stale-noop", ["update", memory["id"], "--patch", json.dumps([{"op":"replace","path":"/title","value":props["title"]}]), "--if-revision", before_link["revision"]], "revision_conflict")
    props = {**props, "title":"After owned change"}
    memory = patch("unconditional-actual-predecessor", memory, [{"op":"replace","path":"/title","value":props["title"]}], props, unconditional=True)
    capture.passed("file/stdin/root/temporary-array behavior and actual owned predecessor disclosure")

    for label, text in [
        ("remove-final-body", '[{"op":"remove","path":"/body"}]'),
        ("null-final-body", '[{"op":"replace","path":"/body","value":null}]'),
        ("extra-final-property", '[{"op":"add","path":"/extra","value":1}]'),
        ("missing-target", '[{"op":"replace","path":"/missing","value":1}]'),
        ("unsupported-op", '[{"op":"copy","from":"/title","path":"/body"}]'),
        ("invalid-pointer", '[{"op":"remove","path":"/bad~2"}]'),
        ("duplicate-member", '[{"op":"replace","op":"add","path":"/body","value":"x"}]'),
        ("unsafe-number", '[{"op":"add","path":"/temporary","value":9007199254740993}]'),
    ]:
        refuse(label, ["update", memory["id"], "--patch", text, "--unconditional"], "invalid_properties")
    refuse("issue-kind", ["update", target["id"], "--patch", '[{"op":"replace","path":"/title","value":"Issue target"}]', "--unconditional"], "capability_unavailable")
    refuse("link-before-input", ["update", link["id"], "--patch", "@"+missing, "--unconditional", "--properties", "{}"], "capability_unavailable")
    for label, flags in [("no-guard-before-input", []),("both-guards-before-input", ["--unconditional","--if-revision",memory["revision"]])]:
        refuse(label, ["update", memory["id"], "--patch", "@"+missing, *flags], "invalid_selector")
    for label, flags in [("mixed-properties", ["--properties", "{}"]), ("mixed-issue-field", ["--priority", "1"])]:
        refuse(label, ["update", memory["id"], "--patch", "@"+missing, "--unconditional", *flags], "capability_unavailable")
    refuse("readonly-before-input", ["update", memory["id"], "--patch", "@"+missing, "--unconditional", "--readonly"], "permission_denied")
    mayor, freeze = capture.work / "mayor", capture.work / "MIGRATION-FREEZE"
    try:
        mayor.mkdir()
        (mayor/"town.json").write_text("{}\n")
        freeze.write_text("property-patch\t2026-09-28T00:00:00Z\tinput policy\n")
        refuse("freeze-before-input", ["update", memory["id"], "--patch", "@"+missing, "--unconditional"], "permission_denied")
    finally:
        freeze.unlink(missing_ok=True)
        (mayor/"town.json").unlink(missing_ok=True)
        mayor.rmdir()
    oversized = file("over-input.json", b" "*(LIMIT+1))
    refuse("input-byte-bound", ["update", memory["id"], "--patch", "@"+oversized, "--unconditional"], "invalid_properties")
    for label, operations in [
        ("operation-bound", [{"op":"replace","path":"/title","value":"x"}]*257),
        ("pointer-byte-bound", [{"op":"remove","path":"/"+"x"*4096}]),
        ("pointer-depth-bound", [{"op":"remove","path":"/"*65}]),
    ]:
        refuse(label, ["update", memory["id"], "--patch", json.dumps(operations), "--unconditional"], "invalid_properties")
    # Parse-time depth and transaction-time working depth have distinct error
    # categories, consistent with input admission versus store capability limits.
    deep = "["*64 + "0" + "]"*64
    refuse("value-depth-bound", ["update", memory["id"], "--patch", '[{"op":"add","path":"/work","value":['+deep+']}]', "--unconditional"], "invalid_properties")
    refuse("working-depth-bound", ["update", memory["id"], "--patch", '[{"op":"add","path":"/work","value":'+deep+'}]', "--unconditional"], "capability_unavailable")
    expensive = [{"op":"add","path":"/work","value":[0]*50000}] + [{"op":"replace","path":"/work/0","value":0}]*220 + [{"op":"remove","path":"/work"}]
    work_file = file("work-budget.json", json.dumps(expensive).encode())
    refuse("evaluation-work-bound", ["update", memory["id"], "--patch", "@"+work_file, "--unconditional"], "capability_unavailable")
    large_file = file("large-body.txt", b"x"*(LIMIT-100))
    large = save("large-before", command("create-large-memory", ["remember", "--body-file", large_file, "--title", "Budget", "--id", "beads/large"]))
    c0.require(large["properties"] == {"title":"Budget", "body":"x"*(LIMIT-100)}, "large input changed")
    refuse("working-document-bound", ["update", large["id"], "--patch", json.dumps([{"op":"replace","path":"/title","value":"t"*200}]), "--if-revision", large["revision"]], "capability_unavailable")
    current("large-after-document-refusal",large)
    large = patch("shrink-large-subject",large,[{"op":"replace","path":"/body","value":"small"}],{"title":"Budget","body":"small"})
    # IO and flag admission precede store writes. All failures retain full state.
    for label, record in [("current-after-refusals",memory),("target-unchanged",target),("link-unchanged",link)]: current(label, record)
    capture.passed("malformed, invalid final representation, wrong kind and admission refusals preserve complete state")
    # Exercise changed human and quiet output, then independently compare the
    # complete postimage. Human unconditional output must identify the predecessor.
    for mode in ["human", "quiet"]:
        before = memory
        text = json.dumps([{"op":"replace","path":"/body","value":mode+" output body"}])
        argv = ["update",before["id"],"--patch",text,"--unconditional","--actor","display-author"]
        if mode == "quiet": argv.append("--quiet")
        receipt, out, err = capture.run(mode+"-changed", argv)
        c0.require(receipt["exit_code"] == 0 and not err, mode+": changed failed")
        if mode == "quiet": c0.require(out == "", "quiet changed leaked output")
        else:
            c0.require("Updated "+before["id"] in out and "Replaced Memory "+before["id"]+" version "+before["version"] in out and before["attribution"]["actor"] in out, "human changed omitted predecessor")
        memory = command(mode+"-changed-read", ["show",before["id"],"--readonly"])
        expected = copy.deepcopy(before)
        expected["properties"]["body"] = mode+" output body"
        for key in ("revision","version","attribution"): expected[key] = memory[key]
        c0.require(memory == expected and memory["version"] != before["version"] and memory["attribution"]["actor"] == "display-author", "display mode altered unexpected state")
        save(mode+"-changed",memory)
    props = memory["properties"]
    for i,item in enumerate(saved):
        record=item["record"]
        c0.require(command("retained-"+str(i), ["show",record["id"],"--version",record["version"],"--readonly"]) == record, "saved record changed")
    receipt, out, err = capture.run("quiet-noop", ["update",memory["id"],"--patch",json.dumps([{"op":"replace","path":"/title","value":props["title"]}]),"--unconditional","--quiet"])
    c0.require(receipt["exit_code"] == 0 and out == "" and err == "", "quiet no-op leaked output")
    receipt, out, err = capture.run("human-noop", ["update",memory["id"],"--patch",json.dumps([{"op":"replace","path":"/title","value":props["title"]}]),"--unconditional"])
    c0.require(receipt["exit_code"] == 0 and "Unchanged" in out and "Replaced" not in out and not err, "human no-op differs")
    capture.passed("all saved versions resolve from fresh processes, human and quiet output retain no-op behavior")
    return {"saved_versions":len(saved), "input_files":len(inputs), "http_exercised":False,
            "expected_records_sha256":c0.sha256(capture.output/"expected-records.json"),
            "input_manifest_sha256":c0.sha256(capture.output/"inputs.json")}


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
