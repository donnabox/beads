#!/usr/bin/env python3
"""Installed CLI / real public-client BDP Read capture; disposable server workspace.

The caller owns the ordinary Dolt server. Its port must be reserved for this
run's sequential provisioning (Dolt 2.1.8 cannot safely provision unrelated
databases concurrently). No production workspace, SQL, or mock is used.
"""

import argparse
import importlib.util
import json
import os
from pathlib import Path
import shutil
import signal
import socket
import subprocess
import threading
import time


PIN = "53bdbd03136875f952af184fce7b3c7af8f74e96"
HERE = Path(__file__).resolve().parent
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
        self.started = time.monotonic()
        self.receipt = {"argv": argv, "cwd": str(capture.work), "started_unix": time.time()}
        self.child = subprocess.Popen(argv, cwd=capture.work, env=capture.env,
                                      stdin=subprocess.DEVNULL, stdout=subprocess.PIPE,
                                      stderr=subprocess.PIPE, start_new_session=True)
        self.receipt["pid"] = self.child.pid
        with capture.lock:
            capture.active.add(self.child)
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
                        self.capture.kill_group(self.child)
        except BaseException as exc:
            self.failure = str(exc)
            self.capture.kill_group(self.child)
        finally:
            pipe.close()

    def wait(self, timeout):
        self.child.wait(timeout=min(timeout, max(0.1, self.capture.deadline - time.monotonic())))
        for thread in self.threads:
            thread.join(timeout=5)
        c0.require(not self.failure, self.failure)
        c0.require(self.child.returncode == 0, f"{self.name} failed; see {self.path}")

    def close(self):
        if self.child.poll() is None:
            os.killpg(self.child.pid, signal.SIGINT)
            try:
                self.child.wait(timeout=10)
            except subprocess.TimeoutExpired:
                self.failure = self.failure or "graceful shutdown timed out"
        self.capture.kill_group(self.child)
        self.child.wait(timeout=10)
        for thread in self.threads:
            thread.join(timeout=5)
        with self.capture.lock:
            self.capture.active.discard(self.child)
        self.receipt.update(exit_code=self.child.returncode, failure=self.failure,
                            elapsed_seconds=time.monotonic() - self.started)
        for stream in ["stdout", "stderr"]:
            self.receipt[stream + "_sha256"] = c0.sha256(self.path / (stream + ".log"))
        c0.write_json(self.path / "receipt.json", self.receipt)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--bd", type=Path, required=True)
    parser.add_argument("--server-port", type=int, required=True)
    parser.add_argument("--bdp-checkout", type=Path, required=True)
    parser.add_argument("--node", type=Path, default=Path(shutil.which("node") or "/missing/node"))
    parser.add_argument("--output-dir", type=Path, required=True)
    parser.add_argument("--command-timeout", type=float, default=60)
    parser.add_argument("--total-timeout", type=float, default=300)
    args = parser.parse_args()
    for name in ["bd", "bdp_checkout", "node", "output_dir"]:
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
        # An archive checkout may lack .git; provenance is validated against
        # root's exact-pin source manifest before use, or directly through git.
        if (args.bdp_checkout / ".git").exists():
            head = subprocess.check_output(["git", "-C", str(args.bdp_checkout), "rev-parse", "HEAD"],
                                           env=capture.env, timeout=10, text=True).strip()
            c0.require(head == PIN, "independent client checkout is not the pinned revision")
            summary["client_revision_evidence"] = {"git_head": head}
        else:
            manifest = args.bdp_checkout.parent / "client-build.json"
            c0.require(manifest.is_file(), "archive checkout requires sibling client-build.json provenance")
            summary["client_revision_evidence"] = {"manifest": str(manifest), "sha256": c0.sha256(manifest)}
            provenance = json.loads(manifest.read_text())
            c0.require(provenance.get("commit") == PIN and provenance.get("passed") is True,
                       "client provenance does not name a successful required-pin build")
            for relative, expected in provenance["files"].items():
                path = args.bdp_checkout / relative
                c0.require(path.resolve().is_relative_to(args.bdp_checkout.resolve()), "unsafe manifest path")
                c0.require(c0.sha256(path) == expected, f"pinned client artifact changed: {relative}")
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
            records[path] = c0.envelope(capture.success("issue-create", ["create", "Deployment task", "--id", path, "--json"]))
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
        serve = Process(capture, "serve", [str(args.bd), "serve", "--readonly", "--graph-mode", "link", "--addr", f"127.0.0.1:{port}"])
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
        client = Process(capture, "public-client", [str(args.node), str(HERE / "graph-bdp-read-client.mjs")])
        processes.append(client)
        client.wait(150)
        c0.require(serve.child.poll() is None, "server exited during client demonstration")
        summary["client"] = json.loads((capture.output / "client-results.json").read_text())
        summary["passed"] = summary["client"]["passed"] is True
    except BaseException as exc:
        failure = f"{type(exc).__name__}: {exc}"
        summary["passed"] = False
    finally:
        for process in reversed(processes):
            try:
                process.close()
                if process.failure or process.child.returncode != 0:
                    failure = failure or f"{process.name} cleanup/exit failure; see receipt"
                    summary["passed"] = False
            except BaseException as exc:
                capture.stop()
                failure = failure or f"cleanup: {exc}"
                summary["passed"] = False
        summary.update(failure=failure, active_children=len(capture.active), cli_commands=len(capture.records),
                       workspace=str(capture.work), server_ownership="caller-owned; not stopped")
        c0.write_json(capture.output / "summary.json", summary)
    c0.require(summary["passed"] and summary["active_children"] == 0, failure or "smoke failed")
    print(json.dumps(summary, indent=2))


if __name__ == "__main__":
    def interrupted(_signal, _frame):
        raise KeyboardInterrupt("capture interrupted")
    signal.signal(signal.SIGTERM, interrupted)
    main()
