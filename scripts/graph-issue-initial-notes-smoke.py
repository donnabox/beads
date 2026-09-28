#!/usr/bin/env python3
"""Compact installed graph Issue initial notes/authorship proof, with normal CLI-only authoring.

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
SCOPE = "https://example.invalid/disposable-initial-notes/"
LIMITATIONS = [
    "initial native Issue notes/creator/owner properties are not authenticated identity, common metadata or a public History contract",
    "Owner follows normal CLI git-email defaults; per-command isolated environment and gitconfig inputs are recorded separately",
    "internal lease absence, exact recorder/audit counts and cumulative workspace-budget rollback are separate storage-test gates",
    "claim is five-minute and nonrenewing; append preserves the observed lease, without expiry/heartbeat/reclaim support",
    "notes replacement/clear remains unavailable; no BDP Write or HTTP request is exercised by this harness",
    "ordinary Dolt provisioning is sequential; no concurrent writer or uncertain-COMMIT claim here",
]


def semantic_digest(value):
    return hashlib.sha256(json.dumps(value, ensure_ascii=False, sort_keys=True,
                                     separators=(",", ":")).encode()).hexdigest()


def exercise(capture):
    backend = "server" if capture.args.server_port else "embedded"
    saved, inputs = [], []
    gitconfig = capture.home / "gitconfig"
    git_text = "[user]\n\tname = Git Creator\n\temail = git-owner@example.invalid\n"
    gitconfig.write_text(git_text)

    def environment(label, argv):
        # Capture's original provenance describes its constructor environment.
        # Record the actual controlled identity inputs for EVERY launched command.
        inputs.append({"label": label, "argv": [str(capture.args.bd), *argv],
                       "environment": dict(capture.env), "gitconfig_path": str(gitconfig),
                       "gitconfig_text": gitconfig.read_text(), "gitconfig_sha256": c0.sha256(gitconfig)})
        c0.write_json(capture.output / "input-environments.json", inputs)

    def command(label, argv):
        args = [*argv, "--json"]
        environment(label, args)
        return c0.envelope(capture.success(label, args))

    def raw(label, argv):
        environment(label, argv)
        receipt, _, _ = capture.run(label, argv)
        folder = capture.output / capture.records[-1]["artifact"]
        return receipt, (folder / "stdout.log").read_bytes(), (folder / "stderr.log").read_bytes()

    def save(label, record):
        c0.require(record.get("id", "").startswith(SCOPE) and record.get("version") and record.get("revision") and
                   record.get("type") and record.get("attribution"), label + ": incomplete record")
        saved.append({"label": label, "record": copy.deepcopy(record), "semantic_sha256": semantic_digest(record)})
        c0.write_json(capture.output / "expected-records.json", saved)
        return record

    def current(label, record):
        c0.require(command(label, ["show", record["id"], "--readonly"]) == record, label + ": current full record differs")

    def exact(label, record):
        c0.require(command(label, ["show", record["id"], "--version", record["version"], "--readonly"]) == record,
                   label + ": saved full record differs")

    def listing(label, records):
        argv = ["list", "--all", "--limit", "0", "--format", "records-json", "--readonly"]
        environment(label, argv)
        result = c0.envelope(capture.success(label, argv))
        expected = {row["id"]: row for row in records}
        c0.require(result.get("hasMore") is False and len(result["items"]) == len(expected) and
                   {row["id"]: row for row in result["items"]} == expected, label + ": full list mismatch")

    def refuse(label, argv, code, exit_code):
        receipt, out, err = raw(label, [*argv, "--json"])
        c0.require(receipt["exit_code"] == exit_code and out == b"", label + ": refusal emitted success or wrong exit")
        problem = json.loads(err)
        c0.require(problem.get("code") == code and problem.get("retryable") is False and "result" not in problem,
                   label + ": wrong typed refusal")

    def transition(label, before, result, fields, actor):
        c0.require(set(result) == {"issue", "changed"} and result["changed"] is True, label + ": missing changed mutation")
        after = result["issue"]
        expected = copy.deepcopy(before)
        for key, value in fields.items():
            if value is None:
                expected["properties"].pop(key, None)
            else:
                expected["properties"][key] = value
        expected["properties"]["updated_at"] = after["properties"]["updated_at"]
        for key in ["version", "revision", "attribution"]:
            expected[key] = after[key]
        c0.require(after == expected and after["version"] != before["version"] and
                   after["attribution"]["actor"] == actor, label + ": unexpected sibling/owned/authorship mutation")
        return save(label, after)

    def set_identity(beads=None, deprecated=None, email=None, git_email=True):
        for key, value in [("BEADS_ACTOR", beads), ("BD_ACTOR", deprecated), ("GIT_AUTHOR_EMAIL", email)]:
            if value is None:
                capture.env.pop(key, None)
            else:
                capture.env[key] = value
        gitconfig.write_text(git_text if git_email else "[user]\n\tname = Git Creator\n")

    init = ["init", "--prefix", "ina", "--non-interactive", "--skip-hooks", "--skip-agents",
            "--graph-mode", "link", "--scope-url", SCOPE]
    if capture.args.server_port:
        init += ["--server", "--external", "--server-host", "127.0.0.1", "--server-port", str(capture.args.server_port),
                 "--database", "ina_" + capture.root.name.replace("-", "_"), "--server-user", "root"]
    initialized = command("normal-init", init)
    c0.require(initialized.get("scope") == SCOPE and initialized.get("backend") == backend, "wrong initialized authority")
    caps = command("authorship-capabilities", ["status", "--graph"])["capabilities"]
    c0.require(caps.get("issueInitialNotes") is True and caps.get("issueCreateAuthorship") is True and
               caps.get("issueWorkflows") is False, "capabilities overstate full Issue compatibility")
    notes = "  Initial notes — 雪\r\nsecond line\n  "
    set_identity("primary-env", "deprecated-env", "  author@example.invalid  ")
    fields = {"description": "First body — 雪", "design": "  Design\r\n ", "acceptance_criteria": "Accepted — café",
              "assignee": "assigned-worker", "estimated_minutes": 0, "external_ref": "tracker #42", "spec_id": "spec/雪",
              "notes": notes, "created_by": "explicit-creator", "owner": "  author@example.invalid  ", "status": "open"}
    source = save("first-source", command("create-initial-notes", ["create", "Authored first Issue", "--id", "beads/source",
                  "--description", fields["description"], "--design", fields["design"], "--acceptance", fields["acceptance_criteria"],
                  "-a", fields["assignee"], "-e", "0", "--external-ref", fields["external_ref"], "--spec-id", fields["spec_id"],
                  "--notes", notes, "--actor", "explicit-creator"]))
    first = copy.deepcopy(source)
    c0.require(all(source["properties"].get(key) == value for key, value in fields.items()) and source["owned"] == [] and
               source["attribution"]["actor"] == "explicit-creator" and
               all(key not in source["properties"] for key in ["lease_expires_at", "heartbeat_at", "lease_granted_node"]),
               "initial notes/identity/assignment or no-claim projection differs")
    current("first-current", source)
    exact("first-exact", source)
    listing("first-list", [source])
    capture.passed("one create retains initial notes, all six fields, independent assignee and explicit creator/owner defaults")

    identity_records = []
    cases = [("primary-env", "primary-env", "deprecated-env", True, "", "primary-env", "git-owner@example.invalid"),
             ("deprecated-env", None, "deprecated-env", True, "-", "deprecated-env", "git-owner@example.invalid"),
             ("git-fallback", None, None, True, "Git note", "Git Creator", "git-owner@example.invalid"),
             ("missing-owner", None, None, False, None, "Git Creator", None)]
    for label, primary, deprecated, git_email, text, creator, owner in cases:
        set_identity(primary, deprecated, git_email=git_email)
        argv = ["create", label, "--id", "beads/" + label, "--assignee", "other-assignee"]
        if text is not None:
            argv += ["--notes", text]
        record = save(label, command("create-" + label, argv))
        c0.require(record["properties"].get("created_by") == creator and record["attribution"]["actor"] == creator and
                   record["properties"].get("owner") == owner and record["properties"].get("notes", "") == (text or "") and
                   record["properties"].get("assignee") == "other-assignee" and record["properties"].get("status") == "open",
                   label + ": default precedence or literal notes changed")
        current(label + "-current", record)
        exact(label + "-exact", record)
        identity_records.append(record)
    set_identity()
    capture.passed("primary/deprecated actor environments, git creator, email fallback and missing owner preserve ordinary CLI precedence")

    target = save("target", command("create-target", ["create", "Prerequisite", "--id", "beads/target"]))
    memory = save("memory", command("create-memory", ["remember", "Related context", "--id", "beads/context", "--title", "Context"]))
    dep = command("add-dependency", ["dep", "add", source["id"], target["id"]])
    source, dependency = save("source-owned", dep["source"]), save("dependency", dep["link"])
    linked = command("link-memory", ["link", source["id"], memory["id"], "--id", "links/context",
                     "--resource-type", SCOPE + "types/preview-related-v2", "--properties", '{"note":"context"}'])
    link = save("context-link", linked["link"])
    c0.require(source["owned"] == [dependency] and linked["source"] == source and
               all(source["properties"].get(key) == value for key, value in fields.items()), "relationships changed initial fields")
    owned = copy.deepcopy(source)
    claimed_result = command("claim-assigned-issue", ["update", source["id"], "--claim", "--actor", "assigned-worker"])
    p = claimed_result["issue"]["properties"]
    c0.require(p.get("started_at") and p.get("heartbeat_at") and p.get("lease_expires_at"), "claim lacks lease projection")
    claim_fields = {"status": "in_progress", "assignee": "assigned-worker"}
    for key in ["started_at", "heartbeat_at", "lease_expires_at", "lease_granted_node"]:
        claim_fields[key] = p.get(key)
    source = transition("claimed", source, claimed_result, claim_fields, "assigned-worker")
    claimed = copy.deepcopy(source)
    appended = command("append-progress", ["update", source["id"], "--append-notes", "Progress — 雪\r\n",
                       "--if-revision", source["revision"], "--actor", "progress-author"])
    source = transition("appended", source, appended, {"notes": notes + "\nProgress — 雪\r\n"}, "progress-author")
    for label, record in [("source-current", source), ("target-unchanged", target), ("memory-unchanged", memory),
                          ("dependency-unchanged", dependency), ("context-unchanged", link)]:
        current(label, record)
    for label, record in [("first-after-progress", first), ("owned-before-claim", owned), ("claimed-before-append", claimed)]:
        exact(label, record)
    listing("after-progress-list", [source, target, *identity_records])
    capture.passed("Dependency/Memory links, native claim and guarded append retain original creator/owner and exact prior complete state")

    human_records = []
    for quiet in [False, True]:
        label = "quiet" if quiet else "human"
        path = "beads/" + label
        receipt, out, err = raw("create-" + label, ["create", label.title(), "--id", path, "--notes", "Human initial note",
                               "--actor", "human-creator", *(["--quiet"] if quiet else [])])
        c0.require(receipt["exit_code"] == 0 and err == b"" and
                   out == (b"" if quiet else ("Created " + path + "\n\n").encode()), label + ": incorrect human/quiet output")
        record = save(label, command("show-" + label, ["show", path, "--readonly"]))
        c0.require(record["properties"].get("notes") == "Human initial note" and
                   record["properties"].get("created_by") == "human-creator" and
                   record["properties"].get("owner") == "git-owner@example.invalid", label + ": human authoring lost fields")
        exact(label + "-exact", record)
        human_records.append(record)
    capture.passed("human and quiet create persist notes and defaults while preserving established exact output")

    for label, identity in [("creator", {"beads": "雪" * 256}), ("owner", {"email": "雪" * 256})]:
        set_identity(**identity)
        refuse("overlong-" + label, ["create", "Refused", "--id", "beads/refused", "--notes=No allocation"],
               "invalid_properties", 2)
        refuse("overlong-" + label + "-absent", ["show", "beads/refused"], "not_found", 3)
    set_identity()
    capture.passed("environment-derived overlong creator and owner refuse without allocating the canonical path")

    for label, text in [("replace-held", "Replacement"), ("clear-held", "")]:
        refuse(label, ["update", source["id"], "--notes", text, "--if-revision", source["revision"]], "capability_unavailable", 5)
    mayor, freeze = capture.work / "mayor", capture.work / "MIGRATION-FREEZE"
    c0.require(not mayor.exists() and not freeze.exists(), "unexpected policy files")
    for policy in ["readonly", "freeze"]:
        try:
            if policy == "freeze":
                mayor.mkdir()
                (mayor / "town.json").write_text("{}\n")
                freeze.write_text("initial-notes\t2026-09-28T00:00:00Z\twrite policy\n")
            before = c0.tree_digest(capture.work)
            refuse("policy-" + policy, ["create", "Refused", "--id", "beads/refused", "--notes", "No allocation",
                                       *(["--readonly"] if policy == "readonly" else [])], "permission_denied", 5)
            c0.require(c0.tree_digest(capture.work) == before, policy + ": refusal changed workspace files")
        finally:
            if policy == "freeze":
                freeze.unlink(missing_ok=True)
                (mayor / "town.json").unlink(missing_ok=True)
                mayor.rmdir()
    refuse("refused-path-absent", ["show", "beads/refused"], "not_found", 3)
    current("final-source-preserved", source)
    exact("final-first-retained", first)
    listing("final-list", [source, target, *identity_records, *human_records])
    c0.require(len(inputs) == len(capture.records), "per-command identity provenance incomplete")
    capture.passed("held replacement/clear and readonly/freeze refuse atomically; every invocation retains exact identity-input provenance")
    return {"saved_versions": len(saved), "http_exercised": False,
            "expected_records_sha256": c0.sha256(capture.output / "expected-records.json"),
            "input_environments_sha256": c0.sha256(capture.output / "input-environments.json")}


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
