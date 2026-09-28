#!/usr/bin/env python3
"""Installed ordered informational Link property-change proof against normal initialized storage.

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
SCOPE = "https://example.invalid/disposable-link-properties-patch/"
LIMIT = 1 << 20
LIMITATIONS = [
    "private ordered Link patch ends in {} or a string note; not arbitrary Link Types or BDP HTTP Write",
    "current/exact records and saved comparisons are not native/public History",
    "internal transaction counts, forced concurrency and lost acknowledgments are qualified by separate real-engine tests",
    "ordinary Dolt2.1.8 different-database provisioning stays serialized",
]


def semantic_digest(value):
    return hashlib.sha256(json.dumps(value, ensure_ascii=False, sort_keys=True,
                                     separators=(",", ":")).encode()).hexdigest()

def exercise(capture):
    backend = "server" if capture.args.server_port else "embedded"
    saved, inputs = [], []
    folder = capture.output / "inputs"
    folder.mkdir(mode=0o700)
    c0.write_json(capture.output / "inputs.json", inputs)

    def file(name, data):
        path = folder / name
        path.write_bytes(data)
        inputs.append({"path": str(path), "bytes": len(data), "sha256": c0.sha256(path)})
        c0.write_json(capture.output / "inputs.json", inputs)
        return str(path)

    def command(label, argv, stdin=None):
        receipt, out, err = capture.run(label, [*argv, "--json"], input_text=stdin)
        c0.require(receipt["exit_code"] == 0, label+": "+err)
        return c0.envelope(json.loads(out))

    def refuse(label, argv, code, exit_code=2):
        receipt, out, err = capture.run(label, [*argv, "--json"])
        c0.require(receipt["exit_code"] == exit_code and out == "", label+": exit/stdout")
        problem = json.loads(err)
        c0.require(problem.get("code") == code and problem.get("retryable") is False and "result" not in problem,
                   label+": typed refusal differs")
        return problem

    def save(label, record):
        c0.require(record["version"] == record["revision"] and record["version"] and record["id"].startswith(SCOPE), label+": identity")
        saved.append({"label":label, "record":copy.deepcopy(record), "semantic_sha256":semantic_digest(record)})
        c0.write_json(capture.output / "expected-records.json", saved)
        return record

    def current(label, record):
        c0.require(command(label, ["show", record["id"], "--readonly"]) == record, label+": complete current differs")

    missing = str(folder / "missing.json")
    refuse("legacy-before-input", ["update", "links/context", "--patch", "@"+missing, "--unconditional"], "capability_unavailable", 5)
    init = ["init", "--prefix", "lpatch", "--non-interactive", "--skip-hooks", "--skip-agents", "--graph-mode", "link", "--scope-url", SCOPE]
    if capture.args.server_port:
        init += ["--server", "--external", "--server-host", "127.0.0.1", "--server-port", str(capture.args.server_port),
                 "--database", "lpatch_"+capture.root.name.replace("-", "_"), "--server-user", "root"]
    initialized = command("normal-init", init)
    c0.require(initialized.get("scope") == SCOPE and initialized.get("backend") == backend, "initialization authority")
    status = command("patch-capabilities", ["status", "--graph"])
    for key in ["linkPropertiesPatch", "memoryPropertiesPatch"]:
        c0.require(status["capabilities"].get(key) is True, "missing capability: "+key)
    for key in ["memory", "historyExact", "requestStatus"]:
        c0.require(status["capabilities"].get(key) is False, "overstated capability: "+key)
    for suffix, value in {"InputBytes":LIMIT, "DocumentBytes":LIMIT, "Operations":256, "PointerBytes":4096,
                          "PointerSegments":64, "Depth":64, "EvaluationBytes":16<<20}.items():
        c0.require(status["limits"].get("linkPatch"+suffix) == value, "incorrect patch limit: "+suffix)
    memory = save("Memory-initial", command("create-memory", ["remember", "Keep body 雪", "--id", "beads/plan", "--title", "Keep title", "--actor", "creator"]))
    target = save("target", command("create-target", ["create", "Issue target", "--id", "beads/target"]))
    issue = save("Issue-source", command("create-Issue-source", ["create", "Unowned source", "--id", "beads/work"]))

    def add(label, source, target_record, path, properties):
        argv = ["link", source["id"], target_record["id"], "--id", path, "--resource-type", SCOPE+"types/preview-related-v2",
                "--properties", json.dumps(properties), "--if-source-revision", source["revision"], "--actor", "link-author"]
        result = command(label, argv)
        save(label+"-link", result["link"])
        save(label+"-source", result["source"])
        return result["link"], result["source"]

    link, memory = add("owned-Link", memory, target, "links/context", {})
    twin, memory = add("equal-endpoints-distinct-ID", memory, target, "links/twin", {"note":"untouched twin"})
    c0.require(link["id"] != twin["id"] and link["source"] == twin["source"] and link["target"] == twin["target"], "equal endpoint identities collapsed")
    unowned, unchanged_issue = add("unowned-Link", issue, target, "links/unowned", {})
    c0.require(unchanged_issue == issue, "informational Link creation changed Issue source")
    original_link, original_memory = copy.deepcopy(link), copy.deepcopy(memory)

    def check_transition(label, before_link, before_source, result, wanted, owned, changed, source_unconditional):
        if not changed:
            c0.require(result == {"link":before_link,"source":before_source,"changed":False}, label+": no-op changed complete records/disclosure")
            return before_link, before_source
        after, source = result["link"], result["source"]
        expected = copy.deepcopy(before_link)
        expected["properties"] = wanted
        for key in ["revision", "version", "attribution"]: expected[key] = after[key]
        c0.require(after == expected and after["version"] != before_link["version"] and after["version"] == after["revision"], label+": complete Link changed outside properties/version")
        c0.require(after["attribution"]["actor"] == "patch-author", label+": Link attribution")
        expected_keys = {"link","source","changed"}
        if owned:
            expected_source = copy.deepcopy(before_source)
            expected_source["owned"] = [after if member["id"] == after["id"] else member for member in before_source["owned"]]
            for key in ["revision", "version", "attribution"]: expected_source[key] = source[key]
            c0.require(source == expected_source and source["version"] != before_source["version"] and source["version"] == source["revision"], label+": complete owned Memory differs")
            c0.require(source["attribution"]["actor"] == "patch-author", label+": Memory attribution")
            if source_unconditional:
                expected_keys.add("replacedSource")
                c0.require(result["replacedSource"] == {"id":before_source["id"],"version":before_source["version"],"attribution":before_source["attribution"]}, label+": wrong actual predecessor disclosure")
        else:
            c0.require(source == before_source, label+": unowned Issue source changed")
        c0.require(set(result) == expected_keys and result["changed"] is True, label+": result/disclosure members differ")
        save(label+"-link", after)
        if owned: save(label+"-source", source)
        return after, source

    def patch(label, before_link, before_source, operations, wanted, *, owned=True, unconditional=False, source_unconditional=False, source_guard=True, changed=True, input_kind="literal", actor="patch-author"):
        text = json.dumps(operations, ensure_ascii=False)
        value, stdin = text, None
        if input_kind == "file": value = "@"+file(label+".json", text.encode())
        if input_kind == "stdin": value, stdin = "@-", text
        argv = ["update", before_link["id"], "--patch", value, "--actor", actor]
        argv += ["--unconditional"] if unconditional else ["--if-revision", before_link["revision"]]
        if source_guard:
            argv += ["--unconditional-source"] if source_unconditional else ["--if-source-revision", before_source["revision"]]
        result = command(label, argv, stdin)
        after, source = check_transition(label, before_link, before_source, result, wanted, owned, changed, source_unconditional)
        current(label+"-fresh-Link", after)
        current(label+"-fresh-source", source)
        return after, source

    link, memory = patch("absent-to-empty", link, memory, [{"op":"add","path":"/note","value":""}], {"note":""})
    link, memory = patch("file-Unicode-note", link, memory, [{"op":"replace","path":"/note","value":"  雪\r\n😀  "}], {"note":"  雪\r\n😀  "}, input_kind="file", unconditional=True)
    patch("same-value-noop", link, memory, [{"op":"replace","path":"/note","value":link["properties"]["note"]}], link["properties"], changed=False, unconditional=True, source_unconditional=True, actor="different-noop-author")
    patch("reversing-noop", link, memory, [{"op":"add","path":"/temporary","value":[1,None]},{"op":"remove","path":"/temporary"}], link["properties"], changed=False)
    link, memory = patch("stdin-remove-note", link, memory, [{"op":"remove","path":"/note"}], {}, input_kind="stdin", source_unconditional=True)
    link, memory = patch("root-add-note", link, memory, [{"op":"replace","path":"","value":{"note":"after-root"}}], {"note":"after-root"}, unconditional=True, source_unconditional=True)
    current("equal-endpoint-twin-unchanged", twin)
    current("target-unchanged-after-owned", target)
    capture.passed("ordered owned patches preserve complete endpoints/twin/target, distinguish absent and empty, and independently disclose source-unconditional replacement")

    no_change = [{"op":"replace","path":"/note","value":"after-root"}]
    refused = [
        ("stale-Link-noop", ["--if-revision",original_link["revision"],"--if-source-revision",memory["revision"]], no_change, "revision_conflict", 4),
        ("stale-source-noop", ["--if-revision",link["revision"],"--if-source-revision",original_memory["revision"]], no_change, "revision_conflict", 4),
        ("owned-missing-source-guard", ["--unconditional"], no_change, "invalid_properties", 2),
        ("remove-absent", ["--unconditional","--unconditional-source"], [{"op":"remove","path":"/absent"}], "invalid_properties", 2),
        ("null-note", ["--unconditional","--unconditional-source"], [{"op":"replace","path":"/note","value":None}], "invalid_properties", 2),
        ("extra-final-member", ["--unconditional","--unconditional-source"], [{"op":"add","path":"/extra","value":True}], "invalid_properties", 2),
        ("empty-operations", ["--unconditional","--unconditional-source"], [], "invalid_properties", 2),
    ]
    for label, guards, operations, code, exit_code in refused:
        refuse(label, ["update",link["id"],"--patch",json.dumps(operations),*guards], code, exit_code)
    for label, flags, code, exit_code in [
        ("mixed-properties-before-input", ["--unconditional","--properties","{}"], "capability_unavailable",5),
        ("missing-Link-guard-before-input", [], "invalid_selector",2),
        ("both-Link-guards-before-input", ["--unconditional","--if-revision","old"], "invalid_selector",2),
        ("false-source-before-input", ["--unconditional","--unconditional-source=false"], "invalid_selector",2),
        ("both-source-guards-before-input", ["--unconditional","--if-source-revision","old","--unconditional-source"], "invalid_selector",2),
        ("readonly-before-input", ["--unconditional","--readonly"], "permission_denied",5),
    ]:
        refuse(label,["update",link["id"],"--patch","@"+missing,*flags],code,exit_code)
    mayor, freeze = capture.work / "mayor", capture.work / "MIGRATION-FREEZE"
    mayor.mkdir(exist_ok=True)
    (mayor / "town.json").write_text("{}\n")
    try:
        freeze.write_text("link-patch\t2026-09-28T00:00:00Z\tinput policy\n")
        refuse("freeze-before-input", ["update",link["id"],"--patch","@"+missing,"--unconditional"], "permission_denied",5)
    finally:
        freeze.unlink(missing_ok=True)
        (mayor / "town.json").unlink(missing_ok=True)
        mayor.rmdir()
    invalid = file("invalid-UTF8.json", b'[{' + bytes([255]) + b'}]')
    refuse("invalid-UTF8", ["update",link["id"],"--patch","@"+invalid,"--unconditional","--unconditional-source"], "invalid_properties",2)
    expensive = [{"op":"add","path":"/work","value":[0]*50000}] + [{"op":"replace","path":"/work/0","value":0}]*220 + [{"op":"remove","path":"/work"}]
    work_file = file("evaluation-budget.json", json.dumps(expensive).encode())
    work_problem = refuse("evaluation-work-bound", ["update",link["id"],"--patch","@"+work_file,"--unconditional","--unconditional-source"], "capability_unavailable",5)
    c0.require("cumulative evaluation bytes" in work_problem.get("message", ""), "evaluation work receipt refused for a different reason")
    refuse("malformed-Link-before-input", ["update","links/","--patch","@"+missing,"--unconditional"], "invalid_selector",2)
    current("owned-Link-after-refusals", link)
    current("Memory-after-refusals", memory)
    capture.passed("stale guards precede no-op; invalid final properties and early admission refusals leave complete records unchanged")

    unowned, _ = patch("Issue-source-no-guard", unowned, issue, [{"op":"add","path":"/note","value":"Issue context"}], {"note":"Issue context"}, owned=False, source_guard=False)
    unowned, _ = patch("Issue-source-explicit-guard", unowned, issue, [{"op":"replace","path":"/note","value":"changed context"}], {"note":"changed context"}, owned=False, unconditional=True)
    unowned, _ = patch("Issue-source-unconditional", unowned, issue, [{"op":"remove","path":"/note"}], {}, owned=False, source_unconditional=True)
    refuse("Issue-source-stale-guard", ["update",unowned["id"],"--patch",'[{"op":"add","path":"/note","value":"x"}]',"--unconditional","--if-source-revision","0"*32], "revision_conflict",4)
    current("Issue-remains-unmodified", issue)
    current("Issue-Link-after-refusal", unowned)
    capture.passed("Issue informational source needs no guard, validates supplied guard and never gains owned state/version/disclosure")

    self_link, memory = add("self-Link", memory, memory, "links/self", {})
    self_link, memory = patch("self-Link-patch", self_link, memory, [{"op":"add","path":"/note","value":"self"}], {"note":"self"})
    current("primary-Link-after-self", link)
    current("target-after-self", target)
    before = memory
    receipt,out,err = capture.run("human-changed",["update",link["id"],"--patch",'[{"op":"replace","path":"/note","value":"human"}]',"--unconditional","--unconditional-source","--actor","patch-author"])
    c0.require(receipt["exit_code"] == 0 and not err and "Updated "+link["id"] in out and "Replaced Memory "+memory["id"]+" version "+memory["version"] in out,"human actual predecessor disclosure")
    after_link = command("human-changed-Link",["show",link["id"]])
    after_source = command("human-changed-source",["show",memory["id"]])
    synthetic = {"link":after_link,"source":after_source,"changed":True,"replacedSource":{"id":before["id"],"version":before["version"],"attribution":before["attribution"]}}
    link,memory = check_transition("human",link,memory,synthetic,{"note":"human"},True,True,True)
    for label,extra in [("human-noop",[]),("quiet-noop",["--quiet"])]:
        receipt,out,err = capture.run(label,["update",link["id"],"--patch",'[{"op":"replace","path":"/note","value":"human"}]',"--unconditional","--unconditional-source",*extra])
        c0.require(receipt["exit_code"] == 0 and not err and (out == "" if extra else "Unchanged "+link["id"] in out and "Replaced" not in out),label+": output")
    relation = command("create-blocking-Dependency", ["dep","add",issue["id"],target["id"],"--actor","dependency-author"])
    dependency = save("blocking-Dependency", relation["link"])
    issue = save("Issue-with-Dependency", relation["source"])
    c0.require(dependency["source"] == issue["id"] and dependency["target"] == target["id"] and
               issue["owned"] == [dependency], "normal Dependency authoring lost complete owned record")
    current("Dependency-before-refusal",dependency)
    current("Dependency-source-before-refusal",issue)
    refuse("dependency-Type", ["update",dependency["id"],"--patch",'[{"op":"add","path":"/note","value":"unsupported"}]',"--unconditional"], "invalid_properties",2)
    current("Dependency-after-refusal",dependency)
    current("Dependency-source-after-refusal",issue)
    current("informational-Link-after-Dependency",unowned)
    current("final-Memory",memory)
    current("final-Link",link)
    current("final-Issue",issue)
    current("final-target",target)
    for i,entry in enumerate(saved):
        record=entry["record"]
        c0.require(command(f"exact-{i:02d}-"+entry["label"],["show",record["id"],"--version",record["version"],"--readonly"]) == record, "exact saved record differs: "+entry["label"])
    capture.passed("self-Link mutation and fresh-process complete retained versions preserve prior owned sets; human and quiet output remain truthful")
    return {"saved_versions":len(saved),"input_files":len(inputs),"http_exercised":False,
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
