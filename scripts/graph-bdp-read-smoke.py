#!/usr/bin/env python3
"""Installed CLI / real public-client BDP Read capture; disposable server workspace.

The caller owns the ordinary Dolt server. Its port must be reserved for this
run's sequential provisioning (Dolt 2.1.8 cannot safely provision unrelated
databases concurrently). No production workspace, SQL, or mock is used.
"""

import argparse
import importlib.util
import hashlib
import json
import os
import runpy
import urllib.request
from pathlib import Path
import shutil
import secrets
import sys
import signal
import socket
import subprocess
import threading
import time


PIN = "53bdbd03136875f952af184fce7b3c7af8f74e96"
HERE = Path(__file__).resolve().parent
sys.dont_write_bytecode = True
spec = importlib.util.spec_from_file_location("graph_c0_capture", HERE / "graph-c0-smoke.py")
c0 = importlib.util.module_from_spec(spec)
spec.loader.exec_module(c0)


class Process:
    """Bounded output, owned process group, and unconditional reap."""

    def __init__(self, capture, name, argv):
        self.capture, self.name = capture, name
        self.path = capture.output / name
        self.path.mkdir()
        self.failure = None
        self.forced = False
        self.port = None
        self.started = time.monotonic()
        self.receipt = {"argv": argv, "cwd": str(capture.work), "started_unix": time.time()}
        self.child = capture.spawn(argv, cwd=capture.work, env=capture.env,
                                      stdin=subprocess.DEVNULL, stdout=subprocess.PIPE,
                                      stderr=subprocess.PIPE)
        self.receipt["pid"] = self.child.pid
        self.threads = []
        for name, pipe in [("stdout", self.child.stdout), ("stderr", self.child.stderr)]:
            thread = threading.Thread(target=self.drain, args=(name, pipe), daemon=True)
            thread.start()
            self.threads.append(thread)

    def drain(self, name, pipe):
        size = 0
        try:
            with (self.path / (name + ".log")).open("wb") as output:
                while True:
                    data = os.read(pipe.fileno(), 65536)
                    if not data:
                        break
                    output.write(data[:max(0, c0.OUTPUT_CAP - size)])
                    size += len(data)
                    if size > c0.OUTPUT_CAP:
                        self.failure = "process output cap exceeded"
                        self.forced = True
                        self.capture.kill_group(self.child)
        except BaseException as exc:
            self.failure = str(exc)
            self.forced = True
            self.capture.kill_group(self.child)
        finally:
            pipe.close()

    def wait(self, timeout):
        self.child.wait(timeout=min(timeout, max(0.1, self.capture.deadline - time.monotonic())))
        for thread in self.threads:
            thread.join(timeout=5)
        c0.require(not any(thread.is_alive() for thread in self.threads), "output reader did not stop")
        c0.require(not self.failure, self.failure)
        c0.require(self.child.returncode == 0, f"{self.name} failed; see {self.path}")

    def close(self):
        if self.child.poll() is None:
            try:
                os.killpg(self.child.pid, signal.SIGINT)
            except ProcessLookupError:
                pass
            try:
                self.child.wait(timeout=10)
            except subprocess.TimeoutExpired:
                self.failure = self.failure or "graceful shutdown timed out"
                self.forced = True
                self.capture.kill_group(self.child)
                self.child.wait(timeout=10)
        group_gone = c0.group_exited(self.child.pid)
        if not group_gone:
            self.failure = self.failure or "live process group after leader exit"
            self.forced = True
            self.capture.kill_group(self.child)
        for thread in self.threads:
            thread.join(timeout=5)
        if any(thread.is_alive() for thread in self.threads):
            self.failure = self.failure or "output readers did not stop"
        port_closed = None
        if self.port is not None:
            try:
                with socket.create_connection(("127.0.0.1", self.port), timeout=1):
                    port_closed = False
            except ConnectionRefusedError:
                port_closed = True
            if not port_closed:
                self.failure = self.failure or "listener remained open after shutdown"
        self.capture.retire(self.child)
        self.receipt.update(exit_code=self.child.returncode, failure=self.failure,
                            forced=self.forced, group_gone=group_gone, port_closed=port_closed,
                            elapsed_seconds=time.monotonic() - self.started)
        for stream in ["stdout", "stderr"]:
            self.receipt[stream + "_sha256"] = c0.sha256(self.path / (stream + ".log"))
        c0.write_json(self.path / "receipt.json", self.receipt)


def observe_python(example, transcript, arguments):
    """Run unchanged example; record actual reads without changing HTTP data."""
    observations = []
    original_open = urllib.request.OpenerDirector.open

    def observed_open(opener, request, *args, **kwargs):
        response = original_open(opener, request, *args, **kwargs)
        original_read, original_close = response.read, response.close
        chunks, recorded = [], False

        def observed_read(*read_args, **read_kwargs):
            data = original_read(*read_args, **read_kwargs)
            chunks.append(data)
            return data

        def observed_close():
            nonlocal recorded
            try:
                return original_close()
            finally:
                if not recorded:
                    recorded = True
                    raw = b"".join(chunks)
                    try:
                        document = json.loads(raw)
                    except (ValueError, UnicodeError):
                        document = None
                    observations.append({"url": request.full_url, "method": request.get_method(),
                        "status": response.status, "body_bytes": len(raw),
                        "body_sha256": hashlib.sha256(raw).hexdigest(), "document": document})

        response.read, response.close = observed_read, observed_close
        return response

    urllib.request.OpenerDirector.open = observed_open
    sys.argv = [str(example), *arguments]
    try:
        runpy.run_path(str(example), run_name="__main__")
    finally:
        urllib.request.OpenerDirector.open = original_open
        c0.write_json(Path(transcript), observations)


def verify_python_pages(observations, scope, expected_ids):
    c0.require(observations and observations[0]["url"] == scope + "bdp.json", "Python discovery observation missing")
    pages = observations[1:]
    c0.require(len(pages) == len(expected_ids) > 1, "Python did not traverse one-record pages")
    next_url, found = scope + "beads/?limit=1", []
    for observation in pages:
        c0.require(observation["url"] == next_url and observation["method"] == "GET"
                   and observation["status"] == 200, "Python did not follow actual returned next URL")
        page = observation["document"]
        c0.require(isinstance(page, dict) and len(page.get("items", [])) == 1, "Python response was not a one-record page")
        found.append(page["items"][0]["id"])
        next_url = page["next"]
    c0.require(next_url is None and len(set(found)) == len(found) and set(found) == set(expected_ids),
               "Python pagination did not exhaust complete inventory")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--bd", type=Path, required=True)
    parser.add_argument("--server-port", type=int, required=True)
    parser.add_argument("--bdp-checkout", type=Path, required=True)
    parser.add_argument("--client-manifest", type=Path, required=True)
    parser.add_argument("--owned-groups", type=Path)
    parser.add_argument("--node", type=Path, default=Path(shutil.which("node") or "/missing/node"))
    parser.add_argument("--output-dir", type=Path, required=True)
    parser.add_argument("--command-timeout", type=float, default=60)
    parser.add_argument("--total-timeout", type=float, default=300)
    args = parser.parse_args()
    for name in ["bd", "bdp_checkout", "node", "output_dir", "client_manifest"]:
        value = getattr(args, name)
        c0.require(value.is_absolute(), f"--{name.replace('_', '-')} must be absolute")
    c0.require(args.bd.is_file() and os.access(args.bd, os.X_OK), "installed bd is not executable")
    c0.require(args.node.is_file() and os.access(args.node, os.X_OK), "Node is not executable")
    c0.require(0 < args.server_port < 65536, "invalid ordinary Dolt port")
    c0.require(args.command_timeout > 0 and args.total_timeout > 0, "timeouts must be positive")
    args.server_root = None
    capture = c0.Capture(args)
    processes = []
    failure = None
    summary = {"passed": False, "client_pin": PIN, "aggregate_public_client_api": False,
               "limitations": ["aggregate uses real fetch plus separate public parsers; pinned client has no aggregate API",
                               "ordinary shared-server only; no History or hot-restore proof"]}
    try:
        # A separate exact-source checkout and successful pinned build are both
        # required; a plausible dist directory alone is not provenance.
        head = subprocess.check_output(["git", "-C", str(args.bdp_checkout), "rev-parse", "HEAD"],
                                       env=capture.env, timeout=10, text=True).strip()
        c0.require(head == PIN, "independent client checkout is not the pinned revision")
        dirty = subprocess.check_output(["git", "-C", str(args.bdp_checkout), "status", "--porcelain", "--untracked-files=all"],
                                        env=capture.env, timeout=10)
        c0.require(not dirty, "independent client source is dirty")
        provenance = json.loads(args.client_manifest.read_text())
        c0.require(provenance.get("commit") == PIN and provenance.get("passed") is True
                   and provenance.get("node") == "v24.16.0" and provenance.get("pnpm") == "11.20.0",
                   "independent client build provenance is incomplete or mismatched")
        c0.require(provenance.get("install") == "pnpm install --frozen-lockfile --ignore-scripts"
                   and provenance.get("build") == "pnpm exec tsc -b packages/client", "wrong client build recipe")
        tracked = subprocess.check_output(["git", "-C", str(args.bdp_checkout), "ls-files", "-z"],
                                          env=capture.env, timeout=10).decode().split("\0")
        required = {path for path in tracked if path}
        for package in ["client", "protocol"]:
            required.update(str(path.relative_to(args.bdp_checkout)) for path in
                            (args.bdp_checkout / "packages" / package / "dist").rglob("*") if path.is_file())
        c0.require(set(provenance["files"]) == required, "client manifest does not cover complete source/dist set")
        for relative, expected in provenance["files"].items():
            path = args.bdp_checkout / relative
            c0.require(path.resolve().is_relative_to(args.bdp_checkout.resolve()), "unsafe manifest path")
            c0.require(c0.sha256(path) == expected, f"pinned client artifact changed: {relative}")
        node_version = subprocess.check_output([str(args.node), "--version"], env=capture.env, timeout=10, text=True).strip()
        c0.require(node_version == provenance["node"], "runtime Node differs from built client")
        summary["client_revision_evidence"] = {"git_head": head, "manifest_sha256": c0.sha256(args.client_manifest)}
        shutil.copyfile(args.client_manifest, capture.output / "client-build.json")
        for package in ["client", "protocol"]:
            path = args.bdp_checkout / "packages" / package / "dist" / "index.js"
            c0.require(path.is_file(), f"build missing: {path}")
            summary[package + "_dist_sha256"] = c0.sha256(path)
        with socket.socket() as reservation:
            reservation.bind(("127.0.0.1", 0))
            port = reservation.getsockname()[1]
        scope = f"http://127.0.0.1:{port}/demo/"
        summary.update(scope=scope, installed_binary_sha256=capture.binary_hash,
                       harness_sha256=c0.sha256(Path(__file__)),
                       client_harness_sha256=c0.sha256(HERE / "graph-bdp-read-client.mjs"))
        init = c0.envelope(capture.success("init", ["init", "--graph-mode", "link", "--scope-url", scope,
            "--prefix", "httpdemo", "--non-interactive", "--skip-hooks", "--skip-agents", "--json",
            "--server", "--external", "--server-host", "127.0.0.1", "--server-port", str(args.server_port),
            "--database", "http_" + capture.root.name.replace("-", "_"), "--server-user", "root"]))
        c0.require(init["scope"] == scope and init["backend"] == "server", "wrong initialized authority")
        records = {}
        for path, title in [("beads/alpha", "First memory"), ("beads/plan", "Plan — 雪")]:
            records[path] = c0.envelope(capture.success("memory-create", c0.remember(path, "Deployment context — 雪", title)))
        for path in ["beads/work", "beads/prereq"]:
            fields = [] if path == "beads/work" else [
                "--design", "Initial design — 雪", "--acceptance", "Ready\r\n",
                "--assignee", "author", "--estimate=0", "--external-ref", " tracker #1 ",
                "--spec-id", " spec ", "--notes", " Initial\r\n雪 "]
            records[path] = c0.envelope(capture.success("issue-create", ["create", "Deployment task", "--id", path, *fields, "--json"]))
        related = scope + "types/preview-related-v2"
        capture.success("memory-issue-link", ["link", "beads/plan", "beads/work", "--resource-type", related,
            "--id", "links/context", "--properties", '{"note":"before page"}',
            "--if-source-revision", records["beads/plan"]["revision"], "--json"])
        capture.success("issue-memory-link", ["link", "beads/work", "beads/plan", "--resource-type", related,
            "--id", "links/back", "--properties", '{"note":"Issue context"}', "--json"])
        dependency = c0.envelope(capture.success("blocking-dependency", ["dep", "add", "beads/work", "beads/prereq", "--json"]))
        for path in [*records, "links/context", "links/back", dependency["link"]["id"]]:
            capture.success("fresh-process-show", ["show", path, "--json"])
        capture.passed("normal initialization and fresh-process Memory/Issue/mixed-Link/Dependency creation and reads")
        token = secrets.token_urlsafe(32)
        token_file = capture.root / "http-token"
        token_file.write_text(token + "\n")
        token_file.chmod(0o600)
        capture.env["BDP_TOKEN"] = token
        serve = Process(capture, "serve", [str(args.bd), "serve", "--readonly", "--graph-mode", "link", "--addr", f"127.0.0.1:{port}", "--auth-token-file", str(token_file)])
        serve.port = port
        processes.append(serve)
        end = min(capture.deadline, time.monotonic() + 30)
        while True:
            c0.require(serve.child.poll() is None, "serve exited before listening (no port/identity retry)")
            try:
                with socket.create_connection(("127.0.0.1", port), timeout=0.2):
                    break
            except OSError:
                c0.require(time.monotonic() < end, "serve readiness deadline exceeded")
                time.sleep(0.1)
        capture.env.update(BDP_SCOPE=scope, BDP_CHECKOUT=str(args.bdp_checkout), BDP_BD=str(args.bd),
                           BDP_EVIDENCE=str(capture.output), BDP_DEPENDENCY_ID=dependency["link"]["id"])
        # Run the exact documented Python HTTP consumer, with one-record pages.
        example = HERE.parent / "examples/bdp-read/read_beads.py"
        summary["python_example_sha256"] = c0.sha256(example)
        for name, extra in [("python-inventory", ["--limit", "1"]),
                            ("python-current", ["--id", scope + "beads/plan"])]:
            consumer = Process(capture, name, [sys.executable, str(Path(__file__)), "--observe-python",
                               str(example), str(capture.output / (name + "-network.json")), "--scope", scope, *extra])
            processes.append(consumer)
            consumer.wait(60)
        inventory = json.loads((capture.output / "python-inventory/stdout.log").read_text())
        current = json.loads((capture.output / "python-current/stdout.log").read_text())
        c0.require(isinstance(inventory, list) and len(inventory) == 4
                   and {item["id"] for item in inventory} == {scope + path for path in records},
                   "Python consumer did not enumerate every Memory and Issue")
        c0.require(current == next(item for item in inventory if item["id"] == scope + "beads/plan")
                   and current["properties"]["title"] == "Plan — 雪", "Python current read differs from inventory")
        observations = json.loads((capture.output / "python-inventory-network.json").read_text())
        verify_python_pages(observations, scope, {scope + path for path in records})
        summary["python_pages"] = len(observations) - 1
        summary["python"] = {"passed": True, "beads": len(inventory), "limit": 1,
                             "authenticated": True, "mechanism": "unchanged standard-library example over BDP HTTP"}
        client = Process(capture, "public-client", [str(args.node), str(HERE / "graph-bdp-read-client.mjs")])
        processes.append(client)
        client.wait(150)
        c0.require(serve.child.poll() is None, "server exited during client demonstration")
        summary["client"] = json.loads((capture.output / "client-results.json").read_text())
        c0.require(summary["client"]["passed"] is True, "public client checks failed before deletion")
        # Alpha remains unreferenced and unchanged throughout the public-client
        # checks. Use its original observed revision, not a fresh pre-read.
        alpha = records["beads/alpha"]
        deleted = c0.envelope(capture.success("memory-delete", ["delete", "beads/alpha", "--force",
            "--if-revision", alpha["revision"], "--json"]))
        c0.require(deleted == {"memory": alpha, "preview": False, "deleted": True},
                   "deletion did not disclose the exact final live predecessor")
        name = "python-after-delete"
        consumer = Process(capture, name, [sys.executable, str(Path(__file__)), "--observe-python",
                           str(example), str(capture.output / (name + "-network.json")), "--scope", scope, "--limit", "1"])
        processes.append(consumer)
        consumer.wait(60)
        remaining = {scope + path for path in records if path != "beads/alpha"}
        after_delete = json.loads((capture.output / name / "stdout.log").read_text())
        observations = json.loads((capture.output / (name + "-network.json")).read_text())
        verify_python_pages(observations, scope, remaining)
        observed_records = [item for observation in observations[1:] for item in observation["document"]["items"]]
        c0.require(isinstance(after_delete, list) and len(after_delete) == 3
                   and {item["id"] for item in after_delete} == remaining and after_delete == observed_records,
                   "Python post-delete inventory differs from actual complete BDP pages")
        authoring = json.loads((capture.output / "client-artifacts.json").read_text())["issueAuthoring"]["after"]
        c0.require([item for item in after_delete if item["id"] == scope + "beads/prereq"] == [authoring],
                   "unchanged Python consumer lost final Issue authoring fields")
        c0.require(serve.child.poll() is None, "server exited during post-delete Python enumeration")
        summary["memory_delete"] = {"passed": True, "id": alpha["id"], "final_live_revision": alpha["revision"]}
        summary["python_after_delete"] = {"passed": True, "beads": 3, "limit": 1, "authenticated": True,
                                           "mechanism": "unchanged standard-library example over BDP HTTP"}
        summary["python_after_delete_pages"] = len(observations) - 1
        summary["passed"] = True
    except BaseException as exc:
        failure = f"{type(exc).__name__}: {exc}"
        summary["passed"] = False
    finally:
        for process in reversed(processes):
            try:
                process.close()
                if process.failure or process.forced or process.child.returncode != 0:
                    failure = failure or f"{process.name} cleanup/exit failure; see receipt"
                    summary["passed"] = False
            except BaseException as exc:
                capture.stop()
                failure = failure or f"cleanup: {exc}"
                summary["passed"] = False
        if capture.active:
            capture.stop()
            failure = failure or "unclosed capture children required emergency cleanup"
            summary["passed"] = False
        if capture.forced_cleanup:
            summary["passed"] = False
        if 'port' in locals():
            summary["http_port"] = port
        summary.update(failure=failure, active_children=len(capture.active), cli_commands=len(capture.records),
                       workspace=str(capture.work), server_ownership="caller-owned; not stopped")
        c0.write_json(capture.output / "summary.json", summary)
    c0.require(summary["passed"] and summary.get("python", {}).get("passed") is True
               and summary.get("python_after_delete", {}).get("passed") is True
               and summary["active_children"] == 0, failure or "smoke failed")
    print(json.dumps(summary, indent=2))


if __name__ == "__main__":
    def interrupted(_signal, _frame):
        raise KeyboardInterrupt("capture interrupted")
    signal.signal(signal.SIGTERM, interrupted)
    if len(sys.argv) > 1 and sys.argv[1] == "--observe-python":
        observe_python(sys.argv[2], sys.argv[3], sys.argv[4:])
    else:
        main()
