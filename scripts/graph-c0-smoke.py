#!/usr/bin/env python3
"""Capture support extracted from source9c86 for the BDP HTTP smoke harness.

No SQL setup, model, or mock is used. A supplied ordinary Dolt server is owned
by the caller and is never started or stopped here. All subprocess output and
workspaces are retained, including on failure. This is a POSIX capture harness.
"""

import hashlib
import json
import os
from pathlib import Path
import selectors
import signal
import subprocess
import tempfile
import threading
import time


OUTPUT_CAP = 2 * 1024 * 1024


def require(condition, message):
    if not condition:
        raise RuntimeError(message)


def write_json(path, value):
    path.write_text(json.dumps(value, ensure_ascii=False, indent=2) + "\n")


def sha256(path):
    digest = hashlib.sha256()
    with path.open("rb") as source:
        for block in iter(lambda: source.read(1024 * 1024), b""):
            digest.update(block)
    return digest.hexdigest()



class Capture:
    def __init__(self, args):
        self.args = args
        self.output = args.output_dir
        self.output.mkdir(mode=0o700, parents=True, exist_ok=False)
        base = Path("/private/tmp") if Path("/private/tmp").is_dir() else Path("/tmp")
        for ancestor in [base, *base.parents]:
            require(not (ancestor / ".beads").exists(), f"ancestor .beads: {ancestor}")
        self.root = Path(tempfile.mkdtemp(prefix="bd-c0-", dir=base)).resolve()
        self.work = self.root / "workspace"
        self.work.mkdir(mode=0o700)
        self.home = self.root / "home"
        self.home.mkdir(mode=0o700)
        self.tmp = self.root / "tmp"
        self.tmp.mkdir(mode=0o700)
        gitconfig = self.home / "gitconfig"
        gitconfig.write_text("[user]\n\tname = C0 Smoke\n\temail = c0@example.invalid\n")
        self.env = {
            "PATH": os.environ.get("PATH", "/usr/bin:/bin"),
            "HOME": str(self.home), "TMPDIR": str(self.tmp),
            "TMP": str(self.tmp), "TEMP": str(self.tmp),
            "XDG_CONFIG_HOME": str(self.home / "config"),
            "XDG_CACHE_HOME": str(self.home / "cache"),
            "XDG_DATA_HOME": str(self.home / "data"),
            "GIT_CONFIG_GLOBAL": str(gitconfig), "GIT_CONFIG_NOSYSTEM": "1",
            "GIT_TERMINAL_PROMPT": "0", "BD_NON_INTERACTIVE": "1",
            "BD_DISABLE_METRICS": "1", "BD_DISABLE_EVENT_FLUSH": "1",
            "DOLT_METRICS_DISABLED": "1", "DOLT_DISABLE_EVENT_FLUSH": "1",
            "PYTHONDONTWRITEBYTECODE": "1", "DO_NOT_TRACK": "1", "BEADS_DOLT_AUTO_START": "0",
            "NO_COLOR": "1", "TZ": "UTC", "LANG": "en_US.UTF-8",
        }
        if os.environ.get("DEVELOPER_DIR"):
            self.env["DEVELOPER_DIR"] = os.environ["DEVELOPER_DIR"]
        self.deadline = time.monotonic() + args.total_timeout
        self.active = set()
        self.lock = threading.Lock()
        self.stopped = threading.Event()
        self.number = 0
        self.records = []
        self.registry = getattr(args, "owned_groups", None) or self.output / "owned-groups.jsonl"
        self.forced_cleanup = False
        self.passes = []
        self.qualification_gaps = []
        self.binary_hash = sha256(args.bd)
        write_json(self.output / "provenance.json", {
            "binary": str(args.bd), "binary_sha256": self.binary_hash,
            "harness_sha256": sha256(Path(__file__).resolve()),
            "private_root": str(self.root), "environment": self.env,
            "server_port": args.server_port,
            "server_root": str(args.server_root) if args.server_root else None,
            "server_ownership": "caller-owned; not started or stopped by harness",
            "command_timeout_seconds": args.command_timeout,
            "total_timeout_seconds": args.total_timeout,
        })

    def registry_event(self, action, child):
        # Each small append is durable before the owner can accept interruption.
        with self.registry.open("a") as output:
            output.write(json.dumps({"action": action, "pid": child.pid}) + "\n")
            output.flush()
            os.fsync(output.fileno())

    def spawn(self, argv, **kwargs):
        # Defer interruption until ownership is published. Do not block the
        # signal mask: exec children would inherit blocked shutdown signals.
        pending = []
        prior = {sig: signal.getsignal(sig) for sig in (signal.SIGTERM, signal.SIGINT)}
        for sig in prior:
            signal.signal(sig, lambda number, _frame: pending.append(number))
        child = None
        try:
            child = subprocess.Popen(argv, start_new_session=True, **kwargs)
            with self.lock:
                self.active.add(child)
                self.registry_event("start", child)
            return child
        except BaseException:
            if child is not None:
                self.kill_group(child)
                child.wait(timeout=10)
            raise
        finally:
            for sig, handler in prior.items():
                signal.signal(sig, handler)
            if pending:
                handler = prior[pending[0]]
                if callable(handler):
                    handler(pending[0], None)
                elif handler != signal.SIG_IGN:
                    raise KeyboardInterrupt("capture interrupted while publishing child")

    def retire(self, child):
        # A dead leader is insufficient: all members of its group must be gone.
        if group_exited(child.pid):
            with self.lock:
                self.registry_event("end", child)
                self.active.discard(child)

    def stop(self):
        self.stopped.set()
        with self.lock:
            children = list(self.active)
        for child in children:
            self.forced_cleanup = True
            self.kill_group(child)
            child.wait(timeout=10)
            self.retire(child)

    @staticmethod
    def kill_group(child):
        try:
            os.killpg(child.pid, signal.SIGKILL)
        except ProcessLookupError:
            pass

    def run(self, label, arguments, input_text=None):
        require(not self.stopped.is_set(), "capture was cancelled")
        require(time.monotonic() < self.deadline, "total capture deadline expired")
        require(sha256(self.args.bd) == self.binary_hash, "installed binary changed")
        with self.lock:
            self.number += 1
            stem = f"{self.number:02d}-{label}"
        directory = self.output / stem
        directory.mkdir()
        started = time.monotonic()
        argv = [str(self.args.bd), *arguments]
        receipt = {"argv": argv, "cwd": str(self.work), "started_unix": time.time()}
        print(f"START {stem}: {arguments[0]}", flush=True)
        stdin_file = None
        if input_text is not None:
            stdin_path = directory / "stdin.txt"
            stdin_path.write_bytes(input_text.encode("utf-8"))
            receipt["stdin_sha256"] = sha256(stdin_path)
            stdin_file = stdin_path.open("rb")
        child = None
        failure = None
        try:
            with (directory / "stdout.log").open("wb") as out, (directory / "stderr.log").open("wb") as err:
                child = self.spawn(argv, cwd=self.work, env=self.env,
                                         stdin=stdin_file if stdin_file is not None else subprocess.DEVNULL, stdout=subprocess.PIPE,
                                         stderr=subprocess.PIPE)
                receipt["pid"] = child.pid
                selector = selectors.DefaultSelector()
                selector.register(child.stdout, selectors.EVENT_READ, out)
                selector.register(child.stderr, selectors.EVENT_READ, err)
                sizes = {out: 0, err: 0}
                try:
                    end = min(self.deadline, started + self.args.command_timeout)
                    while selector.get_map():
                        if self.stopped.is_set() or time.monotonic() >= end:
                            raise RuntimeError("command cancelled or deadline exceeded")
                        for key, _ in selector.select(timeout=0.2):
                            data = os.read(key.fileobj.fileno(), 65536)
                            if not data:
                                selector.unregister(key.fileobj)
                                key.fileobj.close()
                                continue
                            available = OUTPUT_CAP - sizes[key.data]
                            key.data.write(data[:available])
                            sizes[key.data] += len(data)
                            require(sizes[key.data] <= OUTPUT_CAP, "output cap exceeded")
                    child.wait(timeout=max(0.01, end - time.monotonic()))
                finally:
                    selector.close()
        except BaseException as exc:
            failure = f"{type(exc).__name__}: {exc}"
            raise
        finally:
            if stdin_file is not None:
                stdin_file.close()
            if child is not None:
                forced = child.poll() is None
                if forced:
                    failure = failure or "command did not finish naturally"
                    self.kill_group(child)
                child.wait(timeout=10)
                group_gone = group_exited(child.pid)
                if not group_gone:
                    forced = True
                    failure = failure or "command left a live process group"
                    self.kill_group(child)
                receipt.update(forced=forced, group_gone=group_gone)
                for pipe in [child.stdout, child.stderr]:
                    if pipe:
                        pipe.close()
                self.retire(child)
                receipt["exit_code"] = child.returncode
            receipt.update(elapsed_seconds=time.monotonic() - started, failure=failure)
            for stream in ["stdout", "stderr"]:
                path = directory / f"{stream}.log"
                if path.exists():
                    receipt[f"{stream}_sha256"] = sha256(path)
            write_json(directory / "receipt.json", receipt)
            with self.lock:
                self.records.append({"artifact": stem, **receipt})
            print(f"END {stem}: exit={receipt.get('exit_code')} failure={failure}", flush=True)
        require(not failure, failure or "command failed")
        return receipt, (directory / "stdout.log").read_text(), (directory / "stderr.log").read_text()

    def success(self, label, arguments, input_text=None):
        receipt, out, err = self.run(label, arguments, input_text=input_text)
        require(receipt["exit_code"] == 0, f"{label} failed: {err}")
        return json.loads(out)

    def passed(self, name):
        self.passes.append(name)
        print(f"PASS {name}", flush=True)


def envelope(value):
    require(isinstance(value, dict), "expected JSON preview envelope")
    require(value.get("schemaVersion") == 1, "unexpected envelope schemaVersion")
    require(value.get("preview") is True, "missing preview disclosure")
    require(isinstance(value.get("result"), dict), "missing result object")
    return value["result"]



def remember(path, body, title):
    return ["remember", body, "--id", path, "--title", title, "--json"]


def group_exited(pid, timeout=5):
    """Bounded grace for exiting children; forced cleanup is never proof."""
    deadline = time.monotonic() + timeout
    while True:
        try:
            os.killpg(pid, 0)
        except ProcessLookupError:
            return True
        if time.monotonic() >= deadline:
            return False
        time.sleep(0.05)
