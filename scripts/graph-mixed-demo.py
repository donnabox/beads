#!/usr/bin/env python3
"""Record a short installed-CLI story in a fresh disposable graph workspace.

Server mode also reads that same graph through the independently built BDP
client. The caller owns the disposable Dolt server; provision databases serially.
This presentation recording supplements qualification and does not replace it.
"""
import argparse
import importlib.util
import json
import os
import re
import shlex
from pathlib import Path
import signal
import socket
import time

HERE = Path(__file__).resolve().parent
spec = importlib.util.spec_from_file_location("http_capture", HERE / "graph-bdp-read-smoke.py")
http = importlib.util.module_from_spec(spec)
spec.loader.exec_module(http)
c0 = http.c0


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--bd", type=Path, required=True)
    parser.add_argument("--backend", choices=["embedded", "server"], required=True)
    parser.add_argument("--server-port", type=int)
    parser.add_argument("--bdp-checkout", type=Path)
    parser.add_argument("--node", type=Path)
    parser.add_argument("--output-dir", type=Path, required=True)
    parser.add_argument("--command-timeout", type=float, default=60)
    parser.add_argument("--total-timeout", type=float, default=300)
    args = parser.parse_args()
    c0.require(args.command_timeout > 0 and args.total_timeout > 0, "timeouts must be positive")
    for path in [args.bd, args.output_dir]:
        c0.require(path.is_absolute(), "binary and output paths must be absolute")
    c0.require(args.bd.is_file() and os.access(args.bd, os.X_OK), "installed binary required")
    if args.backend == "server":
        c0.require(args.server_port and 0 < args.server_port < 65536, "disposable server port required")
        c0.require(args.bdp_checkout and args.node, "server demo requires pinned BDP checkout and Node")
        for path in [args.bdp_checkout, args.node]:
            c0.require(path.is_absolute(), "client paths must be absolute")
        # Reuse the existing independently built archive and verify its source
        # manifest, never silently accept a different client revision/build.
        manifest = args.bdp_checkout.parent / "client-build.json"
        c0.require(manifest.is_file(), "missing sibling client-build.json; use the presenter guide build recipe")
        provenance = json.loads(manifest.read_text())
        c0.require(provenance.get("commit") == http.PIN and provenance.get("passed") is True,
                   "client build provenance must match the qualified public pin")
        required = {"pnpm-lock.yaml", "packages/client/src/index.ts", "packages/client/dist/index.js",
                    "packages/protocol/src/index.ts", "packages/protocol/dist/index.js", "schemas/bdp-v0.schema.json"}
        c0.require(required <= provenance.get("files", {}).keys(), "client manifest is missing required source/build files")
        for name, digest in provenance["files"].items():
            path = (args.bdp_checkout / name).resolve()
            c0.require(path.is_relative_to(args.bdp_checkout.resolve()), "unsafe manifest path")
            c0.require(c0.sha256(path) == digest, "public client build changed: " + name)
    args.server_root = None
    capture = c0.Capture(args)
    processes, chapters = [], []
    if args.backend == "server":
        c0.write_json(capture.output / "client-provenance.json", {
            "manifest": str(manifest), "manifest_sha256": c0.sha256(manifest),
            "validated_build": provenance,
        })
    summary = {"passed": False, "backend": args.backend, "qualification": False,
               "required_runtime_base": "b7bf5040beae863b0d4fafa8d4b838dd04b3cd20",
               "limitations": ["disposable preview", "saved versions are not public History",
                               "HTTP is Read-only and ordinary-server only"]}

    def run(label, argv):
        return c0.envelope(capture.success(label, [*argv, "--json"]))

    def show(label, path, version=None):
        return run(label, ["show", path, *(["--version", version] if version else [])])

    def chapter(title):
        chapters.append({"title": title, "firstCommand": capture.number + 1})
        print("CHAPTER " + title, flush=True)

    try:
        version = capture.success("binary-version", ["version", "--json"])
        build = version.get("build", "")
        c0.require(re.fullmatch(r"[0-9a-f]{7,40}", build) and summary["required_runtime_base"].startswith(build),
                   "this demo requires a build label from the pinned b7bf5040 source")
        if "commit" in version:
            c0.require(version["commit"] == summary["required_runtime_base"], "binary VCS commit differs from pin")
        summary["binary"] = version
        summary["binary_sha256"] = capture.binary_hash
        with socket.socket() as reservation:
            reservation.bind(("127.0.0.1", 0))
            port = reservation.getsockname()[1]
        scope = f"http://127.0.0.1:{port}/demo/" if args.backend == "server" else "https://example.invalid/mixed-demo/"
        chapter("1. Give the project a plan and a reason")
        init = ["init", "--graph-mode", "link", "--scope-url", scope, "--prefix", "demo",
                "--non-interactive", "--skip-hooks", "--skip-agents"]
        if args.backend == "server":
            init += ["--server", "--external", "--server-host", "127.0.0.1", "--server-user", "root",
                     "--server-port", str(args.server_port), "--database", "demo_" + capture.root.name.replace("-", "_")]
        initialized = run("init", init)
        c0.require(initialized["scope"] == scope and initialized["backend"] == args.backend and
                   initialized["memoryComplete"] is False, "wrong initialized identity/backend/capability")
        capabilities = run("graph-capabilities", ["status", "--graph"])
        c0.require(all(capabilities["capabilities"][name] is False for name in
                   ["memory", "historyExact", "issueWorkflows", "requestStatus", "backupContinuity"]),
                   "demo capability boundary changed")
        plan = run("remember-plan", ["remember", "Deploy after verification.\n", "--id", "beads/plan", "--title", "Deployment plan"])
        rationale = run("remember-rationale", ["remember", "Verification catches regressions.\n", "--id", "beads/rationale", "--title", "Why verify?"])
        run("create-release", ["create", "Release deployment", "--id", "beads/release"])
        run("create-verification", ["create", "Verify deployment", "--id", "beads/verification"])
        related = scope + "types/preview-related-v2"
        run("issue-context", ["link", "beads/release", "beads/plan", "--resource-type", related,
                              "--id", "links/context", "--properties", '{"note":"Read this before release"}'])
        relation = run("memory-rationale", ["link", "beads/plan", "beads/rationale", "--resource-type", related,
                         "--id", "links/rationale", "--properties", '{"note":"Original rationale"}',
                         "--if-source-revision", plan["revision"]])
        linked = show("show-linked-plan", "beads/plan")
        c0.require(linked == relation["source"] and linked["owned"] == [relation["link"]], "complete owned Link missing")
        c0.require(show("show-rationale-unchanged", "beads/rationale") == rationale, "incoming link rewrote target")
        capture.success("list-plan-links", ["links", "beads/plan", "--json"])
        chapter("2. A relationship can carry its own facts")
        edited = run("edit-link-note", ["update", "links/rationale", "--properties", '{"note":"Verification found a real regression"}',
                     "--if-revision", relation["link"]["revision"], "--if-source-revision", linked["revision"]])
        before = show("show-plan-after-link-edit", "beads/plan")
        c0.require(before == edited["source"] and before["owned"] == [edited["link"]], "owned edit not saved")
        c0.require(before["version"] != linked["version"], "owning Memory did not advance")
        c0.require(show("show-original-linked-plan", "beads/plan", linked["version"]) == linked, "old owned state lost")
        chapter("3. Change the plan without losing what it used to say")
        properties = {"title": "Deployment plan", "body": "Deploy after verification and a rollback rehearsal.\n"}
        changed = run("edit-memory", ["update", "beads/plan", "--properties", json.dumps(properties), "--if-revision", before["revision"]])
        after = show("show-current-plan", "beads/plan")
        c0.require(changed["changed"] and after == changed["memory"], "fresh read differs from write")
        c0.require(after["properties"] == properties and after["owned"] == before["owned"], "body edit damaged owned state")
        c0.refusal(capture.run("stale-memory-edit", ["update", "beads/plan", "--properties", json.dumps(properties),
                   "--if-revision", before["revision"], "--json"]), {"revision_conflict"}, "stale edit")
        c0.require(show("show-after-stale-refusal", "beads/plan") == after, "stale writer changed current state")
        old = show("show-saved-plan", "beads/plan", before["version"])
        c0.require(old == before, "saved complete record changed")
        receipt, body, _ = capture.run("recall-saved-body", ["recall", "beads/plan", "--version", before["version"]])
        c0.require(receipt["exit_code"] == 0 and body == before["properties"]["body"], "saved body changed")
        comparison = run("compare-plans", ["compare", "beads/plan", "--from", before["version"], "--to", after["version"]])
        c0.require(comparison["changes"] == [{"area": "properties", "member": "body",
                   "from": {"present": True, "value": before["properties"]["body"]},
                   "to": {"present": True, "value": properties["body"]}}], "unexpected comparison")
        chapter("4. Keep the ordinary Issue workflow connected")
        run("add-blocking-dependency", ["dep", "add", "beads/release", "beads/verification"])
        blocked = capture.success("ready-before-verification", ["ready", "--json"])["result"]
        c0.require([x["id"] for x in blocked] == [scope + "beads/verification"], "release was not blocked")
        run("close-verification", ["close", "beads/verification"])
        ready = capture.success("ready-after-verification", ["ready", "--json"])["result"]
        c0.require([x["id"] for x in ready] == [scope + "beads/release"], "release did not become ready")
        release = show("show-release", "beads/release")
        c0.write_json(capture.output / "expected.json", {"plan": after, "release": release, "rationale": rationale})
        if args.backend == "server":
            chapter("5. Another client reads the same graph through BDP")
            serve = http.Process(capture, "serve", [str(args.bd), "serve", "--readonly", "--addr", f"127.0.0.1:{port}"])
            processes.append(serve)
            end = min(capture.deadline, time.monotonic() + 30)
            while True:
                c0.require(serve.child.poll() is None, "serve exited before readiness")
                try:
                    with socket.create_connection(("127.0.0.1", port), timeout=0.2):
                        break
                except OSError:
                    c0.require(time.monotonic() < end, "serve readiness timed out")
                    time.sleep(0.1)
            capture.env.update(BDP_SCOPE=scope, BDP_CHECKOUT=str(args.bdp_checkout), BDP_EVIDENCE=str(capture.output))
            client = http.Process(capture, "public-client", [str(args.node), str(HERE / "graph-mixed-demo-client.mjs")])
            processes.append(client)
            client.wait(90)
            c0.require(serve.child.poll() is None, "serve exited during the client chapter")
            summary["http"] = json.loads((capture.output / "client-results.json").read_text())
            c0.require(summary["http"]["passed"], "public client failed")
        summary["passed"] = True
    except BaseException as exc:
        summary["failure"] = f"{type(exc).__name__}: {exc}"
    finally:
        for process in reversed(processes):
            try:
                process.close()
                c0.require(not process.failure and process.child.returncode == 0, "process cleanup failed")
            except BaseException as exc:
                capture.stop()
                summary["passed"] = False
                summary["failure"] = summary.get("failure") or f"{type(exc).__name__}: {exc}"
        capture.stop()
        for child in list(capture.active):
            try:
                child.wait(timeout=10)
                capture.active.discard(child)
            except BaseException as exc:
                summary["passed"] = False
                summary["failure"] = summary.get("failure") or f"unreaped child {child.pid}: {exc}"
        summary.update(active_children=len(capture.active), cli_commands=len(capture.records), workspace=str(capture.work), chapters=chapters)
        c0.write_json(capture.output / "summary.json", summary)
        c0.write_json(capture.output / "commands.json", capture.records)
        lines = ["# Recorded mixed graph demo", "", "Disposable preview; this recording supplements qualification and does not qualify a new runtime.", "",
                 f"Backend: {args.backend}. Required runtime source: `{summary['required_runtime_base']}`.",
                 f"Binary SHA-256: `{capture.binary_hash}`. Build label is checked; source/install evidence remains separate.",
                 f"Recording passed: {summary['passed']}. Qualification: false.", "",
                 *["- " + value for value in summary["limitations"]], "",
                 "Actual fresh-process results follow. Absolute paths and hashes are in the adjacent receipts.", ""]
        headings = {x["firstCommand"]: x["title"] for x in chapters}
        for index, command in enumerate(capture.records, 1):
            if index in headings:
                lines += ["## " + headings[index], ""]
            folder = capture.output / command["artifact"]
            lines += ["```sh", shlex.join(["bd", *command["argv"][1:]]), "```", "",
                      "<details><summary>Recorded output (exit " + str(command["exit_code"]) + ")</summary>", "",
                      "````text", (folder / "stdout.log").read_text(), (folder / "stderr.log").read_text(),
                      "````", "", "</details>", ""]
        if args.backend == "server" and (capture.output / "public-client/stdout.log").exists():
            lines += ["## 5. Another client reads the same graph through BDP", "", "```text",
                      (capture.output / "public-client/stdout.log").read_text(),
                      (capture.output / "public-client/stderr.log").read_text(), "```", "",
                      "Request URLs, statuses and body hashes: [client-results.json](client-results.json).", ""]
        (capture.output / "recording.md").write_text("\n".join(lines))
        c0.write_json(capture.output / "demo-source.json", {p.name: c0.sha256(p) for p in [Path(__file__), HERE / "graph-mixed-demo-client.mjs", HERE / "graph-bdp-read-smoke.py", HERE / "graph-c0-smoke.py"]})
    c0.require(not capture.active, "demo leaked an unreaped child; see summary.json")
    c0.require(summary["passed"], summary.get("failure", "demo failed"))
    print(json.dumps(summary, indent=2))


if __name__ == "__main__":
    def interrupted(_signal, _frame):
        raise KeyboardInterrupt("demo interrupted")
    signal.signal(signal.SIGTERM, interrupted)
    main()
