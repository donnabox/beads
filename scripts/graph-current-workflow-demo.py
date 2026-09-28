#!/usr/bin/env python3
"""Prepare a compact, auditable current mixed-workflow recording.

Installed CLI writes only; ordinary-server BDP Read uses the independently built
pinned client. This recording supplements qualification and does not qualify a
runtime. Caller owns server/database/workspace cleanup. No hidden client build.
"""
import argparse
import copy
import importlib.util
import json
import os
from pathlib import Path
import re
import shlex
import shutil
import signal
import socket
import subprocess
import time

HERE = Path(__file__).resolve().parent


def load_helpers(directory):
    spec = importlib.util.spec_from_file_location("current_demo_http", directory / "graph-bdp-read-smoke.py")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def runtime_evidence(args, c0):
    c0.require(re.fullmatch(r"[0-9a-f]{40}", args.runtime_commit), "runtime commit must be a full lowercase SHA")
    evidence = json.loads(args.runtime_evidence.read_text())
    c0.require(evidence.get("commit") == args.runtime_commit and evidence.get("source_clean") is True and
               evidence.get("install_passed") is True and evidence.get("binary_sha256") == c0.sha256(args.bd),
               "supplied clean-source/install evidence does not match the runtime pin/binary")
    artifacts = evidence.get("artifacts", {})
    c0.require({"git-head.txt", "git-status.txt", "install.log"} <= artifacts.keys(),
               "runtime evidence requires git-head.txt, git-status.txt and install.log hashes")
    root = args.runtime_evidence.parent.resolve()
    for relative, digest in artifacts.items():
        c0.require(not Path(relative).is_absolute() and ".." not in Path(relative).parts, "unsafe runtime evidence path")
        path = (root / relative).resolve()
        c0.require(path.is_relative_to(root) and path.is_file() and c0.sha256(path) == digest,
                   "runtime evidence artifact mismatch: " + relative)
    c0.require((root / "git-head.txt").read_text().strip() == args.runtime_commit,
               "captured runtime git HEAD differs")
    c0.require((root / "git-status.txt").read_bytes() == b"", "captured runtime source was not clean")
    c0.require((root / "install.log").stat().st_size > 0, "missing install receipt content")
    return evidence


def client_evidence(args, http):
    c0 = http.c0
    manifest = args.bdp_checkout.parent / "client-build.json"
    c0.require(manifest.is_file(), "missing sibling client-build.json; supply the successful exact-pin build manifest")
    evidence = json.loads(manifest.read_text())
    required = {"pnpm-lock.yaml", "packages/client/src/index.ts", "packages/client/dist/index.js",
                "packages/protocol/src/index.ts", "packages/protocol/dist/index.js", "schemas/bdp-v0.schema.json"}
    c0.require(evidence.get("commit") == http.PIN and evidence.get("passed") is True and
               required <= evidence.get("files", {}).keys(), "missing successful pinned client build evidence")
    for relative, digest in evidence["files"].items():
        path = (args.bdp_checkout / relative).resolve()
        c0.require(path.is_relative_to(args.bdp_checkout.resolve()) and path.is_file(),
                   "missing or unsafe client source/build artifact: " + relative)
        c0.require(c0.sha256(path) == digest, "client source/build artifact differs: " + relative)
    return {"manifest_sha256": c0.sha256(manifest), "validated_build": evidence}


def record(args, http, supplied):
    c0 = http.c0
    capture = c0.Capture(args)
    processes, chapters, saved = [], [], {}
    summary = {"passed": False, "qualification": False, "backend": args.backend,
               "runtime_commit": args.runtime_commit, "binary_sha256": capture.binary_hash,
               "provenance_boundary": "Caller-supplied clean-source/install evidence; not independent source attestation.",
               "limitations": ["disposable preview; full Memory and Issue workflows remain incomplete",
                               "exact saved versions are not chronological or public History",
                               "HTTP is Read-only and ordinary-server only; all writers are installed CLI",
                               "no adoption, aliases, metadata/Inception, custom Types or remote traversal",
                               "caller retains and later cleans disposable workspace/database; server remains caller-owned"]}
    evidence_dir = capture.output / "runtime-evidence"
    evidence_dir.mkdir()
    c0.write_json(evidence_dir / "manifest.json", supplied)
    for relative in supplied["artifacts"]:
        target = evidence_dir / relative
        target.parent.mkdir(parents=True, exist_ok=True)
        shutil.copyfile(args.runtime_evidence.parent / relative, target)
    c0.write_json(capture.output / "demo-source.json", {str(p.name): c0.sha256(p) for p in
                  [Path(__file__), HERE / "graph-current-workflow-demo-client.mjs",
                   args.helpers_dir / "graph-bdp-read-smoke.py", args.helpers_dir / "graph-c0-smoke.py"]})

    def command(label, argv, structured=True):
        actor = ["--actor", "demo"] if argv[0] in {"create", "remember", "dep", "link", "update", "close"} else []
        value = capture.success(label, [*argv, *actor, *(["--json"] if structured else [])])
        c0.require(value.get("schemaVersion") == 1 and value.get("preview") is True and "result" in value,
                   label + ": missing preview envelope")
        return value["result"]

    def keep(name, value):
        c0.require(value.get("revision") == value.get("version") and value.get("version") and value.get("id"),
                   name + ": incomplete saved record")
        saved[name] = copy.deepcopy(value)
        c0.write_json(capture.output / "expected-records.json", saved)
        return value

    def timestamp(value, label):
        c0.require(isinstance(value, str) and re.fullmatch(r"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d{1,9})?Z", value),
                   label + ": missing RFC3339Nano-shaped UTC timestamp")

    def transition(label, before, after, properties, owned=None):
        # Build the oracle from the accepted predecessor. Only opaque fresh
        # version and observed write attribution are learned from the postimage.
        c0.require(isinstance(after.get("version"), str) and after["version"] and
                   after["version"] != before["version"] and after.get("revision") == after["version"],
                   label + ": changed write did not mint a fresh consistent version")
        attribution = after.get("attribution", {})
        c0.require(set(attribution) == {"actor", "status", "recordedAt"} and attribution["actor"] == "demo" and
                   attribution["status"] == "claimed", label + ": wrong explicit actor/attribution")
        timestamp(attribution["recordedAt"], label)
        wanted = copy.deepcopy(before)
        wanted.update(version=after["version"], revision=after["revision"], attribution=copy.deepcopy(attribution))
        wanted["properties"] = copy.deepcopy(properties)
        if owned is not None:
            wanted["owned"] = copy.deepcopy(owned)
        c0.require(after == wanted, label + ": unexpected complete-state transition")
        return after

    def show(label, value, exact=False):
        got = command(label, ["show", value["id"], *(["--version", value["version"]] if exact else []), "--readonly"])
        c0.require(got == value, label + ": complete record changed")
        return got

    def recall(label, value, exact=False):
        receipt, _, _ = capture.run(label, ["recall", value["id"], *(["--version", value["version"]] if exact else []), "--readonly"])
        raw = (capture.output / capture.records[-1]["artifact"] / "stdout.log").read_bytes()
        c0.require(receipt["exit_code"] == 0 and raw == value["properties"]["body"].encode(), label + ": body bytes differ")

    def chapter(title):
        chapters.append({"title": title, "firstCommand": capture.number + 1})

    def traverse(label, depth, records, links, frontier):
        got = command(label, ["graph", "beads/work", "--view", "generic", "--direction", "out", "--depth", str(depth), "--readonly"])
        wanted = {"projection": "summary", "scope": scope, "root": work["id"], "direction": "out", "depth": depth,
                  "maxNodes": 100, "maxLinks": 200,
                  "nodes": sorted([{**{key: row[key] for key in ["id", "type", "version", "attribution"]},
                                    "title": row["properties"]["title"]} for row in records], key=lambda row: row["id"]),
                  "links": sorted([{key: row[key] for key in ["id", "type", "source", "target", "version", "attribution"]}
                                   for row in links], key=lambda row: row["id"]),
                  "frontier": sorted(frontier), "complete": not frontier}
        c0.require(got == wanted, label + ": summary topology/allowlist/frontier differs")

    try:
        version = capture.success("binary-version", ["version", "--json"])
        build = version.get("build", "")
        c0.require(re.fullmatch(r"[0-9a-f]{7,40}", build) and args.runtime_commit.startswith(build), "binary build label differs from supplied pin")
        if "commit" in version:
            c0.require(version["commit"] == args.runtime_commit, "binary commit differs from pin")
        summary["binary"] = version
        if args.backend == "server":
            c0.write_json(capture.output / "client-provenance.json", client_evidence(args, http))
        with socket.socket() as reservation:
            reservation.bind(("127.0.0.1", 0))
            port = reservation.getsockname()[1]
        scope = f"http://127.0.0.1:{port}/demo/" if args.backend == "server" else "https://example.invalid/current-workflow/"
        summary["scope"] = scope
        chapter("1. Give work a plan, a reason, and a prerequisite")
        init = ["init", "--graph-mode", "link", "--scope-url", scope, "--prefix", "demo", "--non-interactive", "--skip-hooks", "--skip-agents"]
        if args.backend == "server":
            database = "demo_" + capture.root.name.replace("-", "_")
            summary["database"] = database
            init += ["--server", "--external", "--server-host", "127.0.0.1", "--server-user", "root", "--server-port", str(args.server_port), "--database", database]
        initialized = command("init", init)
        c0.require(initialized["scope"] == scope and initialized["backend"] == args.backend and initialized["memoryComplete"] is False, "initialization boundary differs")
        status = command("capabilities", ["status", "--graph"])
        c0.require(all(status["capabilities"][key] for key in ["genericTraversal", "issueBlocked", "issueDueFilter", "memoryPropertiesPatch", "linkPropertiesPatch"]), "required preview capability missing")
        c0.require(all(status["capabilities"][key] is False for key in ["memory", "historyExact", "issueWorkflows", "requestStatus", "backupContinuity"]), "incomplete capability disclosure changed")
        work = keep("work-created", command("create-work", ["create", "Release to a small group", "--id", "beads/work", "--due", "2030-01-02T12:00:00Z"]))
        c0.require(work["properties"].get("due_at") == "2030-01-02T12:00:00Z", "initial due timestamp differs")
        verify = keep("verify-open", command("create-verification", ["create", "Verify the deployment", "--id", "beads/verify"]))
        dep = command("add-prerequisite", ["dep", "add", "beads/work", "beads/verify"])
        work, dependency = keep("work-blocked", dep["source"]), keep("dependency", dep["link"])
        plan = keep("plan-created", command("remember-plan", ["remember", "Ship after verification. Keep the rollout reversible.\n", "--id", "beads/plan", "--title", "Deployment plan"]))
        rationale = keep("rationale", command("remember-rationale", ["remember", "Small releases make rollback practical.\n", "--id", "beads/rationale", "--title", "Why start small?"]))
        related = scope + "types/preview-related-v2"
        context = command("link-work-plan", ["link", "beads/work", "beads/plan", "--resource-type", related, "--id", "links/context", "--properties", '{"note":"Read the plan before release"}'])
        c0.require(context["source"] == work, "unowned Issue context changed source")
        context = keep("context", context["link"])
        relation = command("link-plan-rationale", ["link", "beads/plan", "beads/rationale", "--resource-type", related, "--id", "links/rationale", "--properties", '{"note":"Original rationale"}', "--if-source-revision", plan["revision"]])
        before, old_link = keep("plan-linked", relation["source"]), keep("rationale-link-before", relation["link"])
        c0.require(before["properties"] == plan["properties"] and before["owned"] == [old_link], "owned set missing")
        chapter("2. Find the work and follow its context")
        listed = command("list-due-work", ["list", "--format", "records-json", "--due-before", "2030-01-03T00:00:00Z"], structured=False)
        c0.require(listed == {"items": [work], "hasMore": False}, "due selection/complete record differs")
        c0.require(command("blocked-before", ["blocked"]) == [{"issue": work, "blockedBy": [verify["id"]]}], "canonical blockers differ")
        c0.require(command("ready-before", ["ready"]) == [verify], "blocked work unexpectedly ready")
        traverse("graph-depth-one", 1, [work, verify, before], [dependency, context], [before["id"]])
        traverse("graph-depth-two", 2, [work, verify, before, rationale], [dependency, context, old_link], [])
        recall("recall-selected-plan", before)
        chapter("3. Change the plan and preserve the old relationship")
        new_body = "Verification first; then ship to a small group.\n"
        memory_ops = [{"op": "remove", "path": "/body"}, {"op": "add", "path": "/body", "value": new_body}]
        changed = command("patch-plan", ["update", before["id"], "--patch", json.dumps(memory_ops), "--if-revision", before["revision"]])
        c0.require(changed["changed"] is True, "Memory patch unexpectedly no-op")
        mid = keep("plan-body-edited", transition("Memory patch", before, changed["memory"],
                   {**before["properties"], "body": new_body}))
        c0.require(changed["changed"] and mid["properties"] == {**before["properties"], "body": new_body} and mid["owned"] == before["owned"] and mid["version"] != before["version"], "Memory patch did not preserve complete owned state")
        link_ops = [{"op": "remove", "path": "/note"}, {"op": "add", "path": "/note", "value": "Verification supports a small reversible rollout"}]
        edited = command("patch-owned-link", ["update", old_link["id"], "--patch", json.dumps(link_ops), "--if-revision", old_link["revision"], "--if-source-revision", mid["revision"]])
        c0.require(edited["changed"] is True, "owned Link patch unexpectedly no-op")
        link = keep("rationale-link-final", transition("Link patch", old_link, edited["link"], {"note": link_ops[1]["value"]}))
        after = keep("plan-final", transition("owned Memory patch", mid, edited["source"], mid["properties"], [link]))
        c0.require(edited["changed"] and after["properties"] == mid["properties"] and after["owned"] == [link] and after["version"] != mid["version"], "owned patch did not advance complete source")
        c0.require({key: link[key] for key in ["id", "type", "source", "target"]} == {key: old_link[key] for key in ["id", "type", "source", "target"]} and link["properties"] == {"note": link_ops[1]["value"]}, "Link patch changed immutable fields or final properties")
        stale = capture.run("stale-plan-patch", ["update", before["id"], "--patch", json.dumps(memory_ops), "--if-revision", before["revision"], "--actor", "demo", "--json"])
        c0.refusal(stale, {"revision_conflict"}, "stale saved plan")
        stale_stdout = capture.output / capture.records[-1]["artifact"] / "stdout.log"
        c0.require(stale[0]["exit_code"] == 4 and stale_stdout.read_bytes() == b"", "stale refusal must exit4 with zero stdout bytes")
        show("show-after-stale-refusal", after)
        show("show-saved-plan", before, exact=True)
        recall("recall-saved-body", before, exact=True)
        comparison = command("compare-plan-versions", ["compare", before["id"], "--from", before["version"], "--to", after["version"]])
        wanted = {"resource": {key: before[key] for key in ["id", "type"]},
                  "from": {key: before[key] for key in ["version", "attribution"]}, "to": {key: after[key] for key in ["version", "attribution"]},
                  "compared": ["properties", "owned"], "unsupported": ["commonMetadata", "inception", "derivation"],
                  "changes": [{"area": "properties", "member": "body", "from": {"present": True, "value": before["properties"]["body"]}, "to": {"present": True, "value": new_body}},
                              {"area": "owned", "id": link["id"], "from": {"present": True, "value": old_link}, "to": {"present": True, "value": link}}]}
        c0.require(comparison == wanted, "complete saved comparison differs")
        show("show-saved-link", old_link, exact=True)
        chapter("4. Resolve the prerequisite without erasing the graph")
        reason = "Verification passed"
        closed = command("close-verification", ["close", "beads/verify", "--reason", reason])
        c0.require(closed["changed"] is True, "prerequisite close unexpectedly no-op")
        close_properties = copy.deepcopy(verify["properties"])
        for field in ["updated_at", "closed_at"]:
            timestamp(closed["issue"]["properties"].get(field), "close " + field)
            close_properties[field] = closed["issue"]["properties"][field]
        close_properties.update(status="closed", close_reason=reason)
        # This fixture has never held a lease, defer or session marker. Close
        # must preserve all other properties/owned state, including their absence.
        verify_closed = keep("verify-closed", transition("prerequisite close", verify, closed["issue"], close_properties))
        c0.require(command("blocked-after", ["blocked"]) == [], "resolved work still dependency-blocked")
        c0.require(command("ready-after", ["ready"]) == [work], "work did not become ready unchanged")
        traverse("graph-after-close", 2, [work, verify_closed, after, rationale], [dependency, context, link], [])
        show("show-work-unchanged", work)
        show("show-current-link", link)
        show("show-context-unchanged", context)
        show("show-rationale-unchanged", rationale)
        c0.write_json(capture.output / "expected.json", {"beads": [work, verify_closed, after, rationale], "links": [dependency, context, link], "plan": after})
        if args.backend == "server":
            chapters.append({"title": "5. An independent BDP client reads the same state", "processes": ["serve", "public-client"]})
            c0.require(time.monotonic() < capture.deadline, "total capture deadline expired before listener startup")
            # Scope was fixed at init. Refuse this race honestly; never silently
            # select another port/Scope after authoring the database.
            try:
                with socket.socket() as check:
                    check.bind(("127.0.0.1", port))
            except OSError as exc:
                raise RuntimeError("reserved Scope port was taken between init and serve") from exc
            serve = http.Process(capture, "serve", [str(args.bd), "serve", "--readonly", "--addr", f"127.0.0.1:{port}"])
            processes.append(serve)
            end = min(capture.deadline, time.monotonic() + 30)
            while True:
                c0.require(serve.child.poll() is None, "serve exited before listening")
                try:
                    with socket.create_connection(("127.0.0.1", port), timeout=0.2):
                        break
                except OSError:
                    c0.require(time.monotonic() < end, "serve readiness timed out")
                    time.sleep(0.1)
            capture.env.update(BDP_SCOPE=scope, BDP_CHECKOUT=str(args.bdp_checkout), BDP_EVIDENCE=str(capture.output))
            remaining = capture.deadline - time.monotonic()
            c0.require(remaining > 0, "total capture deadline expired before public-client startup")
            summary["public_client_remaining_seconds"] = remaining
            client = http.Process(capture, "public-client", [str(args.node), str(HERE / "graph-current-workflow-demo-client.mjs")])
            processes.append(client)
            remaining = capture.deadline - time.monotonic()
            c0.require(remaining > 0, "total capture deadline expired before public-client wait")
            try:
                client.wait(min(90, remaining))
            except subprocess.TimeoutExpired as exc:
                raise RuntimeError(f"public-client wait expired: {remaining:.3f}s of the {args.total_timeout:g}s backend budget remained at wait start; client wait cap90s") from exc
            c0.require(serve.child.poll() is None, "serve exited during public Read")
            summary["http"] = json.loads((capture.output / "client-results.json").read_text())
            c0.require(summary["http"]["passed"] and summary["http"]["httpRequests"] == 10, "public proof differs")
        summary["passed"] = True
    except BaseException as exc:
        summary["failure"] = f"{type(exc).__name__}: {exc}"
        summary["interrupted"] = isinstance(exc, KeyboardInterrupt)
    finally:
        for process in reversed(processes):
            try:
                process.close()
                c0.require(not process.failure and process.child.returncode == 0, "process cleanup failed")
            except BaseException as exc:
                capture.stop()
                summary["interrupted"] = summary.get("interrupted", False) or isinstance(exc, KeyboardInterrupt)
                summary["passed"] = False
                summary["failure"] = summary.get("failure") or f"cleanup: {exc}"
        capture.stop()
        for child in list(capture.active):
            try:
                child.wait(timeout=10)
                capture.active.discard(child)
            except BaseException as exc:
                summary["interrupted"] = summary.get("interrupted", False) or isinstance(exc, KeyboardInterrupt)
                summary["passed"] = False
                summary["failure"] = summary.get("failure") or f"unreaped child {child.pid}: {exc}"
        summary.update(active_children=len(capture.active), cli_commands=len(capture.records), workspace=str(capture.work), chapters=chapters)
        c0.write_json(capture.output / "summary.json", summary)
        c0.write_json(capture.output / "commands.json", capture.records)
        artifacts = [p for p in capture.output.rglob("*") if p.is_file() and p.name not in {"artifact-hashes.json", "recording.md"}]
        c0.write_json(capture.output / "artifact-hashes.json", {str(p.relative_to(capture.output)): c0.sha256(p) for p in artifacts})
        lines = ["# Current mixed-workflow recording", "", f"Passed: {summary['passed']}. Qualification: false. Backend: {args.backend}.",
                 f"Runtime pin: `{args.runtime_commit}`. Binary SHA-256: `{capture.binary_hash}`.", "", summary["provenance_boundary"], "",
                 *["- " + item for item in summary["limitations"]], "", "[Complete receipt index](commands.json) · [Saved complete records](expected-records.json) · [Source hashes](demo-source.json) · [Artifact hashes](artifact-hashes.json)", ""]
        headings = {entry["firstCommand"]: entry["title"] for entry in chapters if "firstCommand" in entry}
        for index, receipt in enumerate(capture.records, 1):
            if index in headings:
                lines += ["## " + headings[index], ""]
            folder = capture.output / receipt["artifact"]
            lines += ["```sh", shlex.join(["bd", *receipt["argv"][1:]]), "```", "", f"Exit {receipt.get('exit_code')}. [Receipt]({receipt['artifact']}/receipt.json)", ""]
            for stream in ["stdout", "stderr"]:
                log = folder / (stream + ".log")
                raw = log.read_bytes() if log.exists() else b""
                lines += [f"{stream} ({len(raw)} bytes):", "", "````text", raw.decode("utf-8", errors="replace"), "````", ""]
        for entry in chapters:
            if "processes" not in entry:
                continue
            lines += ["## " + entry["title"], ""]
            if (capture.output / "client-results.json").exists():
                lines += ["[Complete HTTP evidence](client-results.json)", ""]
            for name in entry["processes"]:
                folder = capture.output / name
                lines += ["### " + name, ""]
                if (folder / "receipt.json").exists():
                    lines += [f"[Process receipt]({name}/receipt.json)", ""]
                elif not folder.exists():
                    lines += ["Process was not started; see the retained failure in summary.json.", ""]
                else:
                    lines += ["No process receipt was completed; see summary.json and available streams.", ""]
                for stream in ["stdout", "stderr"]:
                    log = folder / (stream + ".log")
                    if not log.exists():
                        lines += [f"{stream}: not captured.", ""]
                        continue
                    raw = log.read_bytes()
                    lines += [f"{stream} ({len(raw)} bytes): [raw log]({name}/{stream}.log)", "", "````text", raw.decode("utf-8", errors="replace"), "````", ""]
        (capture.output / "recording.md").write_text("\n".join(lines))
    return summary


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--bd", type=Path, required=True)
    parser.add_argument("--runtime-commit", required=True)
    parser.add_argument("--runtime-evidence", type=Path, required=True)
    parser.add_argument("--helpers-dir", type=Path, required=True)
    parser.add_argument("--backend", choices=["embedded", "server", "both"], default="both")
    parser.add_argument("--server-port", type=int)
    parser.add_argument("--server-root", type=Path)
    parser.add_argument("--bdp-checkout", type=Path)
    parser.add_argument("--node", type=Path)
    parser.add_argument("--output-dir", type=Path, required=True)
    parser.add_argument("--command-timeout", type=float, default=60)
    parser.add_argument("--total-timeout", type=float, default=300, help="per-backend wall limit")
    args = parser.parse_args()
    if not args.helpers_dir.is_absolute() or not args.helpers_dir.is_dir():
        parser.error("--helpers-dir must name an absolute existing scripts directory")
    for name in ["graph-bdp-read-smoke.py", "graph-c0-smoke.py"]:
        if not (args.helpers_dir / name).is_file():
            parser.error("--helpers-dir is missing required helper " + name)
    http = load_helpers(args.helpers_dir)
    c0 = http.c0
    for path in [args.bd, args.runtime_evidence, args.helpers_dir, args.output_dir]:
        c0.require(path.is_absolute(), "all filesystem arguments must be absolute")
    c0.require(args.command_timeout > 0 and args.total_timeout > 0, "positive timeouts required")
    c0.require(args.bd.is_file() and os.access(args.bd, os.X_OK), "installed executable required")
    if args.backend in ["both", "server"]:
        c0.require(args.server_port and 1024 < args.server_port < 65536, "explicit disposable ordinary-server high port required")
        c0.require(args.bdp_checkout and args.bdp_checkout.is_absolute() and args.node and args.node.is_absolute() and os.access(args.node, os.X_OK), "absolute pinned client checkout and executable Node required")
    evidence = runtime_evidence(args, c0)
    modes = ["embedded", "server"] if args.backend == "both" else [args.backend]
    args.output_dir.mkdir(parents=True, exist_ok=False)
    outcomes = []
    for mode in modes:
        child = copy.copy(args)
        child.backend = mode
        if mode == "embedded":
            child.server_port = None
            child.server_root = None
        child.output_dir = args.output_dir / mode
        outcomes.append(record(child, http, evidence))
        if outcomes[-1].get("interrupted"):
            break  # An operator stop must never initialize another database.
    summary = {"passed": all(row["passed"] and row["active_children"] == 0 for row in outcomes), "qualification": False,
               "runtime_commit": args.runtime_commit, "cli_commands": sum(row["cli_commands"] for row in outcomes), "backends": outcomes}
    c0.write_json(args.output_dir / "summary.json", summary)
    if any(row.get("interrupted") for row in outcomes):
        raise KeyboardInterrupt("recording interrupted; remaining backends were not started")
    c0.require(summary["passed"], "recording failed; complete evidence retained")
    print(json.dumps(summary, indent=2))


if __name__ == "__main__":
    signal.signal(signal.SIGTERM, lambda *_: (_ for _ in ()).throw(KeyboardInterrupt("recording interrupted")))
    main()
