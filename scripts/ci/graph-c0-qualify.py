#!/usr/bin/env python3
"""Required C0 proof: real engines, source-matched CLI, no optional test skips.

This is deliberately the create/read checkpoint, not full Memory or BDP proof.
The caller supplies the ordinary released Dolt binary and CI Build Artifacts.
All databases and process groups belong to this run; no existing server is used.
"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import signal
import socket
import subprocess
import time


def require(condition, message):
    if not condition:
        raise RuntimeError(message)


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def group_exited(pid, timeout=5):
    # A command may exit just before its already-exiting child is reaped.
    # Allow a bounded grace period, but never count forced cleanup as success.
    deadline = time.monotonic() + timeout
    while True:
        try:
            os.killpg(pid, 0)
        except ProcessLookupError:
            return True
        if time.monotonic() >= deadline:
            return False
        time.sleep(0.05)


def verify_tests(events, expected):
    runs, passes = set(), set()
    for event in events:
        name, action = event.get("Test"), event.get("Action")
        require(action not in ("fail", "skip"), f"required test did not pass: {event}")
        if name:
            require(name.split("/", 1)[0] in expected, f"unexpected test or subtest: {name}")
            if action == "run":
                require(name not in runs, f"duplicate test run: {name}")
                runs.add(name)
            elif action == "pass":
                require(name in runs and name not in passes, f"unmatched test pass: {name}")
                passes.add(name)
    roots = {name for name in runs if "/" not in name}
    require(expected and roots == expected and runs == passes,
            f"test discovery/execution mismatch: expected roots={sorted(expected)}, runs={sorted(runs)}, passes={sorted(passes)}")
    require(any(e.get("Action") == "pass" and not e.get("Test") for e in events),
            "missing successful package completion")
    return len(roots)


class Qualification:
    def __init__(self, args):
        self.root = Path(__file__).resolve().parents[2]
        self.output = Path(args.output).resolve()
        self.output.mkdir(parents=True, exist_ok=False)
        self.bd = Path(args.bd).resolve()
        self.dolt = Path(args.dolt).resolve()
        self.artifacts = Path(args.artifacts).resolve()
        self.children = set()
        self.receipts = []
        self.counts = {}
        self.env = {k: os.environ[k] for k in (
            "PATH", "GOCACHE", "GOMODCACHE", "GOROOT", "DEVELOPER_DIR", "TMPDIR") if k in os.environ}
        home = self.output / "home"
        home.mkdir()
        for key in ("GOCACHE", "GOMODCACHE"):
            if key not in self.env:
                self.env[key] = subprocess.check_output(["go", "env", key], cwd=self.root, text=True).strip()
        self.env.update(HOME=str(home), XDG_CONFIG_HOME=str(home / ".config"),
                        GIT_CONFIG_NOSYSTEM="1", GIT_CONFIG_GLOBAL=str(home / "empty-gitconfig"),
                        CGO_ENABLED="1", BD_DISABLE_METRICS="1", BD_DISABLE_EVENT_FLUSH="1",
                        DOLT_METRICS_DISABLED="1", DOLT_DISABLE_EVENT_FLUSH="1",
                        BEADS_DOLT_AUTO_START="0", NO_COLOR="1",
                        BEADS_TEST_BD_BINARY=str(self.bd), BEADS_TEST_IGNORE_REPO_CONFIG="1")

    def run(self, args, label, cwd=None, timeout=120, stdin=None, expected=0):
        print(f"C0 {label}: {args}", flush=True)
        process = subprocess.Popen([str(a) for a in args], cwd=cwd or self.root,
                                   env=self.env, stdin=subprocess.PIPE, stdout=subprocess.PIPE,
                                   stderr=subprocess.PIPE, start_new_session=True)
        self.children.add(process.pid)
        try:
            out, err = process.communicate(stdin, timeout=timeout)
        except BaseException:
            os.killpg(process.pid, signal.SIGKILL)
            out, err = process.communicate()
            (self.output / f"{label}.stdout").write_bytes(out)
            (self.output / f"{label}.stderr").write_bytes(err)
            raise
        finally:
            self.children.discard(process.pid)
        (self.output / f"{label}.stdout").write_bytes(out)
        (self.output / f"{label}.stderr").write_bytes(err)
        if not group_exited(process.pid):
            os.killpg(process.pid, signal.SIGKILL)
            raise RuntimeError(f"{label} left a live process group after exit")
        self.receipts.append(dict(label=label, argv=[str(a) for a in args], pid=process.pid,
                                  exit=process.returncode, stdoutSHA256=hashlib.sha256(out).hexdigest()))
        require(process.returncode == expected,
                f"{label} exited {process.returncode}, expected {expected}; see saved stdout/stderr")
        return out

    def tests(self, package, selector, files, label):
        # Source discovery also sees accidentally build-excluded added tests.
        expected = set()
        for path in files:
            names = re.findall(r"^func (Test\w+)\(t \*testing.T\)", path.read_text(), re.MULTILINE)
            expected.update(name for name in names if re.search(selector, name))
        listing = self.run(["go", "test", "-tags", "gms_pure_go", "-list", selector, package],
                           label + "-discovery", timeout=600).decode()
        listed = set(re.findall(r"^Test\w+$", listing, re.MULTILINE))
        require(listed == expected and expected, f"{label}: source and compiled test discovery disagree")
        out = self.run(["go", "test", "-tags", "gms_pure_go", "-json", "-count=1", "-p=1",
                        "-parallel=1", "-timeout=15m", "-run", selector, package], label, timeout=960)
        self.counts[label] = verify_tests([json.loads(line) for line in out.splitlines()], expected)

    def cli(self, work, label, *args):
        return json.loads(self.run([self.bd, *args, "--json"], label, cwd=work))

    def capture(self, engine, port):
        work = self.output / (engine + "-workspace")
        work.mkdir()
        scope = f"https://example.invalid/ci-c0/{engine}/"
        self.run(["git", "init", "-q", str(work)], engine + "-git-init")
        self.run(["git", "-C", str(work), "config", "core.hooksPath", ".git/hooks"], engine + "-git-hooks")
        flags = ["--server", "--external", "--server-host", "127.0.0.1", "--server-port", str(port), "--server-user", "root"] if engine == "server" else []
        initialized = self.cli(work, engine + "-init", "init", "--graph-mode", "link", "--scope-url", scope,
                               "--skip-hooks", "--skip-agents", "--non-interactive", *flags)
        require(initialized["result"]["backend"] == engine and initialized["result"]["scope"] == scope,
                "initialization selected the wrong backend or Scope")
        created = self.cli(work, engine + "-create", "remember", "Durable C0 body — 記憶", "--id", "beads/plan", "--title", "Plan")
        record = created["result"]
        require(record["id"] == scope + "beads/plan" and record["type"] == scope + "types/preview-memory-v2"
                and record["properties"] == {"title": "Plan", "body": "Durable C0 body — 記憶"}
                and record["owned"] == [] and record["revision"] and record["version"], "incomplete/non-Memory creation")
        # Each call starts and reaps a distinct installed CLI process.
        for suffix, identity in [("show", "beads/plan"), ("reopen", record["id"])]:
            require(self.cli(work, engine + "-" + suffix, "show", identity) == created,
                    "fresh-process read differs from committed creation")
        status = self.cli(work, engine + "-status", "status", "--graph")
        require(status["result"]["capabilities"]["memoryRead"] is True, "Memory read not advertised")

    def execute(self):
        require(self.bd.is_file() and os.access(self.bd, os.X_OK), "missing installed CLI")
        require(self.dolt.is_file() and os.access(self.dolt, os.X_OK), "missing released Dolt")
        head = self.run(["git", "rev-parse", "HEAD"], "head").decode().strip()
        require(not self.run(["git", "status", "--porcelain", "--untracked-files=all"], "source-status"),
                "qualification requires clean source, including no untracked files")
        manifest = dict(line.split("=", 1) for line in (self.artifacts / "build-manifest.txt").read_text().splitlines())
        require(manifest["commit"] == head and manifest["build_tags"] == "gms_pure_go", "CLI artifact source/build flags mismatch")
        expected_hash = (self.artifacts / "SHA256SUMS").read_text().split()[0]
        require(digest(self.bd) == expected_hash, "installed CLI hash differs from build artifact")
        version = self.run([self.dolt, "version"], "dolt-version").decode().splitlines()
        require(version and version[0].strip() == "dolt version 2.1.8", "C0 requires released Dolt2.1.8")
        (self.output / "source.json").write_text(json.dumps(dict(commit=head, binarySHA256=digest(self.bd),
            doltSHA256=digest(self.dolt), runnerSHA256=digest(Path(__file__))), indent=2) + "\n")
        with socket.socket() as reservation:
            reservation.bind(("127.0.0.1", 0))
            port = reservation.getsockname()[1]
        server_root = self.output / "server-data"
        server_root.mkdir()
        log = (self.output / "server.log").open("wb")
        server = subprocess.Popen([str(self.dolt), "sql-server", "--host", "127.0.0.1", "--port", str(port), "--data-dir", str(server_root)],
                                  cwd=server_root, env=self.env, stdout=log, stderr=subprocess.STDOUT, start_new_session=True)
        self.children.add(server.pid)
        try:
            deadline = time.monotonic() + 30
            while True:
                require(server.poll() is None, "owned server exited before readiness")
                try:
                    with socket.create_connection(("127.0.0.1", port), timeout=1):
                        break
                except OSError:
                    require(time.monotonic() < deadline, "server readiness timed out")
                    time.sleep(0.2)
            self.env["BEADS_GRAPH_TEST_SERVER_PORT"] = str(port)
            # Sequential package runs/provisioning; concurrency inside same-store tests remains exercised.
            self.tests("./internal/storage/graphstore", "^Test", (self.root / "internal/storage/graphstore").glob("*_test.go"), "storage")
            self.tests("./internal/configfile", "^TestGraphMode", (self.root / "internal/configfile").glob("graph_mode_test.go"), "config")
            self.tests("./cmd/bd", "^Test(GraphModeCLI|GraphPreview)", (self.root / "cmd/bd").glob("graph*test.go"), "cli")
            self.capture("embedded", port)
            self.capture("server", port)
            require(server.poll() is None, "server exited during qualification")
        finally:
            server.terminate()
            forced = False
            try:
                code = server.wait(timeout=15)
            except subprocess.TimeoutExpired:
                forced = True
                os.killpg(server.pid, signal.SIGKILL)
                code = server.wait()
            finally:
                self.children.discard(server.pid)
                log.close()
            group_gone = group_exited(server.pid)
            if not group_gone:
                os.killpg(server.pid, signal.SIGKILL)
            port_closed = False
            try:
                with socket.create_connection(("127.0.0.1", port), timeout=1):
                    pass
            except ConnectionRefusedError:
                port_closed = True
            self.receipts.append(dict(label="server-shutdown", pid=server.pid, port=port,
                                      exit=code, forced=forced, groupGone=group_gone,
                                      portClosed=port_closed))
            require(code == 0 and not forced and group_gone and port_closed,
                    f"server cleanup failed: exit={code}, forced={forced}, groupGone={group_gone}, portClosed={port_closed}")
        require(not self.children, "owned child process remains")
        (self.output / "summary.json").write_text(json.dumps(dict(commit=head, passed=self.counts,
            installedCLICommands=10, engines=["embedded", "server"], childrenRemaining=0), indent=2) + "\n")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    for name in ("bd", "dolt", "artifacts", "output"):
        parser.add_argument("--" + name, required=True)
    def interrupted(signum, _frame):
        raise RuntimeError(f"qualification interrupted by signal {signum}")

    signal.signal(signal.SIGTERM, interrupted)
    signal.signal(signal.SIGINT, interrupted)
    qualification = Qualification(parser.parse_args())
    try:
        qualification.execute()
    finally:
        (qualification.output / "receipts.json").write_text(json.dumps(qualification.receipts, indent=2) + "\n")


if __name__ == "__main__":
    main()
