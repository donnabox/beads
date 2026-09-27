#!/usr/bin/env python3
"""Prove installed-CLI disclosure of actual overwritten Memory versions.

Each command starts a new process in a normally initialized disposable workspace.
Both backends run sequentially. The caller owns the ordinary Dolt server.
No SQL, fixtures, mocks or schema/bootstrap helpers are used.
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
SCOPE = "https://example.invalid/disposable-overwrite-disclosure/"
LIMITATIONS = [
    "provisional CLI result fields only; no BDP wire expansion or HTTP writes",
    "recorded attribution is copied exactly, not asserted to be a native commit instant",
    "normal CLI actors are claimed; unknown attribution needs separate storage evidence",
    "no concurrent writes, injected rollback, authority corruption, cancellation or uncertain-commit qualification",
    "current/retained equality does not prove internal database/event-count equality",
    "ordinary Dolt 2.1.8 provisioning must remain serialized",
]

def semantic_digest(value):
    return hashlib.sha256(json.dumps(value, ensure_ascii=False, sort_keys=True,
                                     separators=(",", ":")).encode()).hexdigest()

def exercise(capture):
    backend = "server" if capture.args.server_port else "embedded"
    init = ["init", "--graph-mode", "link", "--scope-url", SCOPE, "--prefix", "overwrite",
            "--non-interactive", "--skip-hooks", "--skip-agents", "--json"]
    if capture.args.server_port:
        init += ["--server", "--external", "--server-host", "127.0.0.1", "--server-port",
                 str(capture.args.server_port), "--database", "replace_" + capture.root.name.replace("-", "_"),
                 "--server-user", "root"]
    initialized = c0.envelope(capture.success("normal-init", init))
    c0.require(initialized.get("scope") == SCOPE and initialized.get("backend") == backend, "wrong initialization")
    status = c0.envelope(capture.success("status", ["status", "--graph", "--json"]))
    c0.require(status.get("capabilities", {}).get("memoryOverwriteDisclosure") is True,
               "status omitted Memory overwrite disclosure capability")
    c0.require(status.get("capabilities", {}).get("memory") is False and
               status.get("capabilities", {}).get("historyExact") is False, "partial disclosure overstated Memory/History")
    saved, disclosures, human_disclosures = [], [], []

    def save(label, record):
        c0.require(isinstance(record, dict) and record.get("id", "").startswith(SCOPE), label + ": no record")
        for field in ["type", "version", "revision", "attribution"]:
            c0.require(record.get(field), label + ": missing " + field)
        saved.append({"label": label, "record": copy.deepcopy(record), "semantic_sha256": semantic_digest(record)})
        c0.write_json(capture.output / "expected-records.json", saved)
        return record

    def command(label, args):
        return c0.envelope(capture.success(label, [*args, "--json"]))

    def exact(label, record):
        actual = command(label, ["show", record["id"], "--version", record["version"], "--readonly"])
        c0.require(actual == record, label + ": complete retained state/attribution changed")

    def current(label, records):
        for index, record in enumerate(records):
            actual = command(label + "-" + str(index), ["show", record["id"], "--readonly"])
            c0.require(actual == record, label + ": target, source or current Link changed")

    def disclosure(label, result, field=None, before=None):
        other = {"replaced", "replacedSource"}
        if field is not None:
            wanted = {key: copy.deepcopy(before[key]) for key in ["id", "version", "attribution"]}
            c0.require(result.get(field) == wanted, label + ": wrong replaced identity/version/attribution")
            other.remove(field)
            disclosures.append({"label": label, "field": field, "expected": wanted,
                                "actual": copy.deepcopy(result[field])})
            c0.write_json(capture.output / "disclosures.json", disclosures)
        c0.require(not any(key in result for key in other), label + ": unexpected replacement field (including null)")

    def refusal(label, args, codes):
        problem = c0.refusal(capture.run(label, [*args, "--json"]), set(codes), label)
        c0.require(not any(key in problem for key in ["result", "replaced", "replacedSource"]),
                   label + ": refusal leaked a success/disclosure")

    def human_disclosure(label, args, before):
        receipt, out, err = capture.run(label, args)
        c0.require(receipt["exit_code"] == 0 and not err, label + ": human command failed")
        attribution = before["attribution"]
        quote = lambda value: json.dumps(value, ensure_ascii=False)
        expected = (f"Replaced Memory {before['id']} version {before['version']}; recorded attribution: "
                    f"actor={quote(attribution['actor'])} status={quote(attribution['status'])} "
                    f"recordedAt={quote(attribution['recordedAt'])}")
        c0.require(out.splitlines().count(expected) == 1, label + ": missing/extraneous human prior-state disclosure")
        human_disclosures.append({"label": label, "expected_line": expected, "stdout": out,
                                  "prior": {key: copy.deepcopy(before[key]) for key in ["id", "version", "attribution"]}})
        c0.write_json(capture.output / "human-disclosures.json", human_disclosures)

    memory = save("memory-created", command("create-memory",
        ["remember", "Old body — 雪\r\n  keep whitespace  ", "--id", "beads/plan", "--title", "Original",
         "--actor", "memory-original-author"]))
    target = save("memory-target", command("create-target",
        ["remember", "Unchanged neighboring Memory", "--id", "beads/target", "--title", "Neighbor",
         "--actor", "target-author"]))
    issue = save("issue", command("create-issue", ["create", "Unchanged Issue", "--id", "beads/work",
                                                "--actor", "issue-author"]))
    related = SCOPE + "types/preview-related-v2"

    def create_link(label, path, before, target_record, unconditional):
        guard = ["--unconditional-source"] if unconditional else ["--if-source-revision", before["revision"]]
        result = command(label, ["link", before["id"], target_record["id"], "--resource-type", related,
            "--id", path, "--properties", '{"note":"initial"}', *guard, "--actor", label + "-author"])
        disclosure(label, result, "replacedSource" if unconditional else None, before)
        c0.require(result.get("changed") is True, label + ": not changed")
        after, link = save(label + "-source", result["source"]), save(label + "-link", result["link"])
        c0.require(after["properties"] == before["properties"] and after["version"] != before["version"] and
                   after["owned"] == sorted([*before["owned"], link], key=lambda item: item["id"]),
                   label + ": wrong source transition")
        exact(label + "-retained-source", before)
        return after, link

    memory, first = create_link("guarded-link-create", "links/first", memory, target, False)
    before_body = memory
    new_properties = {"title": "Edited", "body": "New body — 雪\nStill owns the same context."}
    direct = command("unconditional-memory-update", ["update", memory["id"], "--properties",
        json.dumps(new_properties, ensure_ascii=False), "--unconditional", "--actor", "body-replacer"])
    disclosure("unconditional-memory-update", direct, "replaced", memory)
    memory = save("memory-unconditionally-edited", direct["memory"])
    c0.require(direct.get("changed") is True and memory["properties"] == new_properties and
               memory["owned"] == before_body["owned"] and memory["version"] != before_body["version"],
               "direct edit changed ownership or failed to version")
    exact("replaced-memory-complete-state", before_body)
    receipt, _, _ = capture.run("replaced-memory-body-recall",
        ["recall", before_body["id"], "--version", direct["replaced"]["version"], "--readonly"])
    raw = capture.output / capture.records[-1]["artifact"]
    c0.require(receipt["exit_code"] == 0 and (raw / "stderr.log").read_bytes() == b"" and
               (raw / "stdout.log").read_bytes() == before_body["properties"]["body"].encode(),
               "replaced version cannot be recalled byte-exactly")
    for label, guards in [("unconditional-memory-noop", ["--unconditional"]),
                          ("guarded-memory-noop", ["--if-revision", memory["revision"]])]:
        noop = command(label, ["update", memory["id"], "--properties", json.dumps(memory["properties"]),
                              *guards, "--actor", "different-noop-author"])
        disclosure(label, noop)
        c0.require(noop.get("changed") is False and noop["memory"] == memory, label + ": no-op changed state")
    guarded_properties = {"title": "Guarded edit", "body": "Explicitly based on the observed revision."}
    guarded_before = memory
    guarded = command("guarded-memory-update", ["update", memory["id"], "--properties",
        json.dumps(guarded_properties), "--if-revision", memory["revision"], "--actor", "guarded-body-author"])
    disclosure("guarded-memory-update", guarded)
    memory = save("memory-guarded-edit", guarded["memory"])
    c0.require(guarded.get("changed") is True and memory["properties"] == guarded_properties and
               memory["owned"] == guarded_before["owned"], "guarded edit failed")
    exact("guarded-memory-old-version", guarded_before)
    current("after-memory-edits", [memory, first, target, issue])
    capture.passed("direct unconditional edit discloses actual old attribution/version; guarded/no-op omit; exact old body and ownership reopen")

    human_before = memory
    human_properties = {"title": "Human-mode change", "body": "Unconditional human output explains the replaced version."}
    human_disclosure("human-memory-update", ["update", memory["id"], "--properties", json.dumps(human_properties),
                     "--unconditional", "--actor", "human-body-author"], memory)
    memory = save("human-memory-updated", command("human-memory-result", ["show", memory["id"], "--readonly"]))
    c0.require(memory["properties"] == human_properties and memory["owned"] == human_before["owned"] and
               memory["version"] != human_before["version"], "human Memory edit failed")
    exact("human-memory-old-version", human_before)
    capture.passed("human unconditional Memory output identifies the exact replaced version and recorded attribution")

    memory, second = create_link("unconditional-link-create", "links/second", memory, issue, True)

    def update_link(label, before, link, link_unconditional, source_unconditional, properties):
        guard = ["--unconditional"] if link_unconditional else ["--if-revision", link["revision"]]
        source_guard = ["--unconditional-source"] if source_unconditional else ["--if-source-revision", before["revision"]]
        result = command(label, ["update", link["id"], "--properties", json.dumps(properties),
                                *guard, *source_guard, "--actor", label + "-author"])
        changed = properties != link["properties"]
        disclosure(label, result, "replacedSource" if source_unconditional and changed else None, before)
        c0.require(result.get("changed") is changed, label + ": wrong changed classification")
        if not changed:
            c0.require(result["source"] == before and result["link"] == link, label + ": no-op changed state")
            return before, link
        after, updated = save(label + "-source", result["source"]), save(label + "-link", result["link"])
        c0.require(after["version"] != before["version"] and updated["version"] != link["version"] and
                   updated["properties"] == properties and after["properties"] == before["properties"] and
                   after["owned"] == [updated if item["id"] == link["id"] else item for item in before["owned"]],
                   label + ": wrong complete owned transition")
        exact(label + "-old-source", before)
        exact(label + "-old-link", link)
        return after, updated

    for label, link_unconditional, source_unconditional in [
        ("guarded-link-unconditional-source", False, True),
        ("unconditional-link-guarded-source", True, False),
        ("both-unconditional", True, True),
        ("both-guarded", False, False),
    ]:
        memory, first = update_link(label, memory, first, link_unconditional, source_unconditional, {"note": label})
    for label, link_unconditional, source_unconditional in [
        ("both-unconditional-link-noop", True, True),
        ("guarded-link-unconditional-source-noop", False, True),
        ("unconditional-link-guarded-source-noop", True, False),
        ("both-guarded-link-noop", False, False),
    ]:
        memory, first = update_link(label, memory, first, link_unconditional, source_unconditional, first["properties"])
    human_source, human_link = memory, first
    human_disclosure("human-owned-link-update", ["update", first["id"], "--properties", '{"note":"human context"}',
                     "--if-revision", first["revision"], "--unconditional-source", "--actor", "human-link-author"], memory)
    first = save("human-link-updated", command("human-link-result", ["show", first["id"], "--readonly"]))
    memory = save("human-link-source-updated", command("human-link-source-result", ["show", memory["id"], "--readonly"]))
    c0.require(first["properties"] == {"note": "human context"} and first["version"] != human_link["version"] and
               memory["version"] != human_source["version"] and memory["properties"] == human_source["properties"] and
               memory["owned"] == [first if item["id"] == first["id"] else item for item in human_source["owned"]],
               "human owned Link edit changed wrong source state")
    exact("human-link-old-source", human_source)
    exact("human-link-old-link", human_link)
    capture.passed("human owned-Link output discloses prior source independently of its guarded Link revision")
    current("after-link-edits", [memory, first, second, target, issue])
    capture.passed("owned Link creation/update disclosure follows source guard independently of Link guard; no-op omits both fields")

    good_properties = json.dumps(first["properties"])
    bad = [
        ("stale-memory-change", ["update", memory["id"], "--properties", '{"title":"bad","body":"bad"}',
                                "--if-revision", before_body["revision"]], {"revision_conflict"}),
        ("stale-memory-noop", ["update", memory["id"], "--properties", json.dumps(memory["properties"]),
                              "--if-revision", before_body["revision"]], {"revision_conflict"}),
        ("invalid-memory", ["update", memory["id"], "--properties", '{"body":"missing title"}',
                            "--unconditional"], {"invalid_properties"}),
        ("stale-link-unconditional-source", ["update", first["id"], "--properties", '{"note":"bad"}',
                  "--if-revision", "stale", "--unconditional-source"], {"revision_conflict"}),
        ("stale-link-noop-unconditional-source", ["update", first["id"], "--properties", good_properties,
                  "--if-revision", "stale", "--unconditional-source"], {"revision_conflict"}),
        ("unconditional-link-stale-source", ["update", first["id"], "--properties", '{"note":"bad"}',
                  "--unconditional", "--if-source-revision", "stale"], {"revision_conflict"}),
        ("unconditional-link-noop-stale-source", ["update", first["id"], "--properties", good_properties,
                  "--unconditional", "--if-source-revision", "stale"], {"revision_conflict"}),
        ("invalid-link", ["update", first["id"], "--properties", '{"unknown":true}',
                         "--unconditional", "--unconditional-source"], {"invalid_properties"}),
        ("refused-create", ["link", memory["id"], target["id"], "--resource-type", related,
                           "--id", "links/first", "--unconditional-source"], {"identity_reserved"}),
        ("stale-unlink", ["unlink", first["id"], "--unconditional", "--if-source-revision", "stale"],
                         {"revision_conflict"}),
        ("readonly-memory", ["update", memory["id"], "--properties", '{"title":"bad","body":"bad"}',
                              "--unconditional", "--readonly"], {"permission_denied"}),
        ("readonly-link", ["update", first["id"], "--properties", '{"note":"bad"}',
                           "--unconditional", "--unconditional-source", "--readonly"], {"permission_denied"}),
    ]
    for label, args, codes in bad:
        refusal(label, args, codes)
    current("after-refusals", [memory, first, second, target, issue])
    capture.passed("stale changes/no-ops, invalid payload, identity collision and readonly failures emit no successful disclosure")

    def unlink(label, before, link, link_unconditional, source_unconditional):
        guard = ["--unconditional"] if link_unconditional else ["--if-revision", link["revision"]]
        source_guard = ["--unconditional-source"] if source_unconditional else ["--if-source-revision", before["revision"]]
        result = command(label, ["unlink", link["id"], *guard, *source_guard, "--actor", label + "-author"])
        disclosure(label, result, "replacedSource" if source_unconditional else None, before)
        after = save(label + "-source", result["source"])
        c0.require(result.get("changed") is True and result["link"]["state"] == "deleted" and
                   result["link"]["previousVersion"] == link["version"] and
                   after["version"] != before["version"] and after["properties"] == before["properties"] and
                   after["owned"] == [item for item in before["owned"] if item["id"] != link["id"]],
                   label + ": wrong deletion transition")
        exact(label + "-old-source", before)
        exact(label + "-old-link", link)
        refusal(label + "-repeat", ["unlink", link["id"], "--unconditional", "--unconditional-source"], {"gone"})
        return after

    memory = unlink("unconditional-link-guarded-source-unlink", memory, second, True, False)
    memory = unlink("guarded-link-unconditional-source-unlink", memory, first, False, True)
    memory, third = create_link("third-guarded-create", "links/third", memory, target, False)
    memory = unlink("both-unconditional-unlink", memory, third, True, True)
    current("after-unlinks", [memory, target, issue])
    capture.passed("owned unlink discloses only unconditional source replacement; deleted Link and complete prior owner remain retained")

    unowned = command("issue-unowned-create", ["link", issue["id"], target["id"], "--resource-type", related,
         "--id", "links/unowned", "--properties", '{"note":"first"}', "--unconditional-source", "--actor", "unowned-author"])
    disclosure("issue-unowned-create", unowned)
    c0.require(unowned["source"] == issue and unowned.get("changed") is True, "unowned create changed Issue")
    unowned = command("issue-unowned-update", ["update", unowned["link"]["id"], "--properties", '{"note":"second"}',
        "--unconditional", "--unconditional-source", "--actor", "unowned-editor"])
    disclosure("issue-unowned-update", unowned)
    c0.require(unowned["source"] == issue and unowned.get("changed") is True, "unowned update changed Issue")
    deleted = command("issue-unowned-unlink", ["unlink", unowned["link"]["id"], "--unconditional",
                       "--unconditional-source", "--actor", "unowned-remover"])
    disclosure("issue-unowned-unlink", deleted)
    c0.require(deleted["source"] == issue and deleted.get("changed") is True, "unowned unlink changed Issue")
    current("final-current", [memory, target, issue])
    capture.passed("unowned Issue Link create/update/unlink does not invent a replaced source despite unconditional-source")
    for entry in saved:
        exact(entry["label"] + "-final-retained", entry["record"])
    capture.passed("all saved complete versions and recorded attribution remain independently readable after later writes")
    return {"saved_versions": len(saved), "disclosures": len(disclosures), "human_disclosures": len(human_disclosures),
            "refusal_cases": len(bad) + 3, "http_exercised": False,
            "saved_records_sha256": c0.sha256(capture.output / "expected-records.json"),
            "disclosures_sha256": c0.sha256(capture.output / "disclosures.json"),
            "human_disclosures_sha256": c0.sha256(capture.output / "human-disclosures.json")}


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
