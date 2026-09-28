#!/usr/bin/env python3
"""Compact installed graph Issue initial-field proof, with normal CLI-only authoring.

Run embedded and caller-owned ordinary Dolt sequentially. No SQL setup, mocks,
hidden bootstrap, internal count claims or concurrent-process qualification.
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
SCOPE = "https://example.invalid/disposable-create-fields/"
LIMITATIONS = [
    "one CLI create publishes all six initial fields; SQL NULL representation, exact recorder/audit counts and no-lease proof belong to separate storage tests",
    "complete current/exact saved records and owned Links are not public ordered History or BDP Write proof",
    "no concurrent writer or uncertain-COMMIT proof in this harness; those are separate storage qualification gates",
    "initial notes and ordinary creator/owner have a separate proof; status/defer, metadata and composite create remain outside this preview",
    "ordinary Dolt database provisioning is sequential, through normal init only",
]


def semantic_digest(value):
    return hashlib.sha256(json.dumps(value, ensure_ascii=False, sort_keys=True,
                                     separators=(",", ":")).encode()).hexdigest()


def exercise(capture):
    backend = "server" if capture.args.server_port else "embedded"
    saved = []

    def command(label, argv):
        return c0.envelope(capture.success(label, [*argv, "--json"]))

    def raw(label, argv):
        receipt, _, _ = capture.run(label, argv)
        folder = capture.output / capture.records[-1]["artifact"]
        return receipt, (folder / "stdout.log").read_bytes(), (folder / "stderr.log").read_bytes()

    def save(label, record):
        c0.require(record.get("id", "").startswith(SCOPE) and record.get("version") and
                   record.get("revision") and record.get("type") and record.get("attribution"),
                   label + ": incomplete record")
        saved.append({"label": label, "record": copy.deepcopy(record), "semantic_sha256": semantic_digest(record)})
        c0.write_json(capture.output / "expected-records.json", saved)
        return record

    def current(label, record):
        c0.require(command(label, ["show", record["id"], "--readonly"]) == record,
                   label + ": current complete record differs")

    def exact(label, record):
        c0.require(command(label, ["show", record["id"], "--version", record["version"], "--readonly"]) == record,
                   label + ": first/saved complete record differs")

    def listing(label, records):
        result = c0.envelope(capture.success(label, ["list", "--all", "--limit", "0", "--format", "records-json", "--readonly"]))
        wanted = {record["id"]: record for record in records}
        c0.require(result.get("hasMore") is False and len(result["items"]) == len(wanted) and
                   {item["id"]: item for item in result["items"]} == wanted,
                   label + ": list differs from complete expected records")

    def refuse(label, argv, code="invalid_properties", exit_code=2, structured=True):
        receipt, out, err = raw(label, [*argv, "--json"])
        c0.require(out == b"" and (receipt["exit_code"] == exit_code if structured else receipt["exit_code"] != 0),
                   label + ": wrong failure status/output")
        if structured:
            problem = json.loads(err)
            c0.require(problem.get("code") == code and problem.get("retryable") is False and
                       "result" not in problem, label + ": wrong typed refusal")
        else:
            c0.require(err != b"", label + ": missing parser diagnostic")

    init = ["init", "--prefix", "icf", "--non-interactive", "--skip-hooks", "--skip-agents",
            "--graph-mode", "link", "--scope-url", SCOPE]
    if capture.args.server_port:
        init += ["--server", "--external", "--server-host", "127.0.0.1", "--server-port", str(capture.args.server_port),
                 "--database", "icf_" + capture.root.name.replace("-", "_"), "--server-user", "root"]
    initialized = command("normal-init", init)
    c0.require(initialized.get("scope") == SCOPE and initialized.get("backend") == backend, "wrong initialized authority")
    capabilities = command("create-capabilities", ["status", "--graph"])["capabilities"]
    c0.require(capabilities.get("issueCreateFields") is True and capabilities.get("issueWorkflows") is False,
               "creation capability overstates complete Issue workflows")
    design, acceptance = "  Design — 雪\r\nnext  ", "  Accepted — café e\u0301\r\n  "
    owner, external, spec_id = "  rig/crew/雪  ", "  tracker #42 — 雪  ", "  specs/雪\r\nsection  "
    fields = {"design": design, "acceptance_criteria": acceptance, "assignee": owner,
              "estimated_minutes": 0, "external_ref": external, "spec_id": spec_id}
    source = save("initial-source", command("create-six-fields", ["create", "Complete first Issue", "--id", "beads/source",
                  "--description", "Original body — 雪\r\n", "--design", design, "--acceptance", acceptance,
                  "-a", owner, "-e", "0", "--external-ref", external, "--spec-id", spec_id, "--actor", "initial-author"]))
    first = copy.deepcopy(source)
    c0.require(all(source["properties"].get(key) == value for key, value in fields.items()) and
               source["properties"]["status"] == "open" and source["properties"].get("title") == "Complete first Issue" and
               source["properties"].get("description") == "Original body — 雪\r\n" and source["owned"] == [] and
               source["attribution"]["actor"] == "initial-author", "initial create lost fields or claimed the Issue")
    current("first-fresh-process", first)
    exact("first-exact-process", first)
    listing("first-complete-list", [first])
    capture.passed("all six literal fields appear in the initial create, fresh show, exact version and complete list")

    target = save("target", command("create-omitted", ["create", "Prerequisite", "--id", "beads/target"]))
    c0.require("estimated_minutes" not in target["properties"] and "external_ref" not in target["properties"],
               "omitted estimate/reference gained a value")
    empty = save("empty-fields", command("create-empty", ["create", "Empty fields", "--id", "beads/empty", "--design", "-",
                 "--acceptance=", "--assignee=", "--estimate", "7", "--external-ref=", "--spec-id="]))
    c0.require(empty["properties"].get("design") == "-" and empty["properties"].get("estimated_minutes") == 7 and
               all(key not in empty["properties"] for key in ["external_ref", "spec_id", "assignee", "acceptance_criteria"]),
               "empty/omitted native CLI semantics changed")
    current("empty-fresh", empty)
    exact("empty-exact", empty)
    capture.passed("omitted nullable fields remain absent; explicit zero, positive estimate, empty fields and literal design hyphen preserve CLI semantics")

    memory = save("memory", command("create-memory", ["remember", "Linked context — 雪", "--id", "beads/context", "--title", "Context"]))
    dep_result = command("add-blocker", ["dep", "add", source["id"], target["id"]])
    source, dependency = save("with-blocker", dep_result["source"]), save("dependency", dep_result["link"])
    link_result = command("link-context", ["link", source["id"], memory["id"], "--id", "links/context",
                          "--resource-type", SCOPE + "types/preview-related-v2", "--properties", '{"note":"context"}'])
    link = save("context-link", link_result["link"])
    c0.require(link_result["source"] == source and source["owned"] == [dependency] and
               all(source["properties"].get(key) == value for key, value in fields.items()), "link creation changed initial fields")
    before_edit = copy.deepcopy(source)
    edited = command("guarded-edit", ["update", source["id"], "--design", "Revised", "--estimate", "9",
                     "--if-revision", source["revision"], "--actor", "later-author"])
    c0.require(set(edited) == {"issue", "changed"} and edited["changed"] is True, "edit did not report one change")
    source = save("edited", edited["issue"])
    expected = copy.deepcopy(before_edit)
    expected["properties"].update(design="Revised", estimated_minutes=9, updated_at=source["properties"]["updated_at"])
    for key in ["version", "revision", "attribution"]:
        expected[key] = source[key]
    c0.require(source == expected and source["version"] != before_edit["version"] and
               source["attribution"]["actor"] == "later-author", "edit changed sibling fields/owned state")
    for label, record in [("current-source", source), ("target-preserved", target), ("memory-preserved", memory),
                          ("dependency-preserved", dependency), ("context-preserved", link)]:
        current(label, record)
    exact("initial-after-links-and-edit", first)
    exact("owned-before-edit", before_edit)
    listing("mixed-complete-list", [source, target, empty])
    capture.passed("later owned Dependency, informational Memory Link and guarded edit preserve initial exact record and unrelated complete state")

    human_records = []
    for quiet in [False, True]:
        label = "quiet" if quiet else "human"
        path = "beads/" + label
        receipt, out, err = raw("create-" + label, ["create", label.title(), "--id", path, "--design", "human design",
                               "--acceptance", "human acceptance", "--assignee", "human-owner", "--estimate", "4",
                               "--external-ref", "human-reference", "--spec-id", "human-spec", *(["--quiet"] if quiet else [])])
        c0.require(receipt["exit_code"] == 0 and err == b"" and
                   out == (b"" if quiet else ("Created " + path + "\n\n").encode()), label + ": human output changed")
        record = save(label, command("show-" + label, ["show", path, "--readonly"]))
        c0.require(all(record["properties"].get(key) == value for key, value in {
            "design": "human design", "acceptance_criteria": "human acceptance", "assignee": "human-owner",
            "estimated_minutes": 4, "external_ref": "human-reference", "spec_id": "human-spec"}.items()), label + ": lost fields")
        exact(label + "-exact", record)
        human_records.append(record)
    capture.passed("ordinary human and quiet create both persist all six fields with exact established output")

    common = ["create", "Refused", "--id", "beads/refused", "--design", "must not allocate"]
    for label, flags in [("negative", ["--estimate=-1"]), ("overflow", ["--estimate=2147483648"]),
                         ("assignee-long", ["--assignee", "雪" * 256]), ("external-long", ["--external-ref", "雪" * 256]),
                         ("spec-long", ["--spec-id", "雪" * 1025])]:
        refuse(label, [*common, *flags])
    for label, flags in [("status-with-notes", ["--notes=", "--status=open"]), ("status", ["--status=open"]), ("defer", ["--defer=tomorrow"]),
                         ("metadata", ["--metadata={}"]), ("design-file", ["--design-file", str(capture.root / "missing-design")])]:
        # Do not combine mutually exclusive design/design-file: test graph admission itself.
        argv = common[:-2] if label == "design-file" else common
        refuse(label, [*argv, *flags], "capability_unavailable", 5)
    refuse("claim-unknown", [*common, "--claim"], structured=False)
    refuse("refused-path-absent", ["show", "beads/refused"], "not_found", 3)
    capture.passed("invalid estimates/lengths and unsupported status/defer/metadata/file/claim (including status with admitted notes) refuse without allocating the canonical path")

    mayor, freeze = capture.work / "mayor", capture.work / "MIGRATION-FREEZE"
    c0.require(not mayor.exists() and not freeze.exists(), "unexpected policy files")
    for policy in ["readonly", "freeze"]:
        try:
            if policy == "freeze":
                mayor.mkdir()
                (mayor / "town.json").write_text("{}\n")
                freeze.write_text("issue-create-fields\t2026-09-27T00:00:00Z\twrite policy\n")
            before = c0.tree_digest(capture.work)
            refuse("policy-" + policy, [*common, *(["--readonly"] if policy == "readonly" else [])], "permission_denied", 5)
            c0.require(c0.tree_digest(capture.work) == before, policy + ": changed workspace files")
        finally:
            if policy == "freeze":
                freeze.unlink(missing_ok=True)
                (mayor / "town.json").unlink(missing_ok=True)
                mayor.rmdir()
    listing("final-complete-list", [source, target, empty, *human_records])
    current("final-source-unchanged", source)
    exact("final-first-version", first)
    refuse("final-refused-path-absent", ["show", "beads/refused"], "not_found", 3)
    capture.passed("read-only and migration freeze refuse creation; complete visible state and initial retained record remain unchanged")
    return {"saved_versions": len(saved), "http_exercised": False,
            "expected_records_sha256": c0.sha256(capture.output / "expected-records.json")}


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
