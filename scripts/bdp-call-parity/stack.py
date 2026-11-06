"""Bring up the servers the parity harness compares, and take them down again.

A Stack is one `dolt sql-server` with its own data directory and one workspace
initialized against it. A served Stack also runs `bd serve` on that workspace.
A client workspace is a second git-initialized directory attached with
`bd connect` to a served Stack.

Everything lives under one temporary root, and reap() only kills processes whose
command line or working directory is under that root.
"""

import json
import os
import shutil
import signal
import socket
import subprocess
import tempfile
import time
import urllib.error
import urllib.request

CALLABLE_MEDIA = "application/vnd.bdp.v0.1+json"
TELEMETRY_ENV = {
    "BD_DISABLE_METRICS": "1",
    "DO_NOT_TRACK": "1",
    "BD_DISABLE_EVENT_FLUSH": "1",
    "DOLT_DISABLE_EVENT_FLUSH": "1",
}
DROPPED_ENV = ("BEADS_DIR", "BEADS_DB", "BD_ACTOR", "BEADS_ACTOR", "BEADS_DOLT_SERVER_PORT",
               "BEADS_DOLT_PASSWORD", "BEADS_HTTP_TOKEN", "BEADS_HTTP_TOKEN_COMMAND")


def free_port():
    with socket.socket() as sock:
        sock.bind(("127.0.0.1", 0))
        return sock.getsockname()[1]


class Root:
    """The temporary root: HOME, binaries on PATH, and every child process."""

    def __init__(self, bd, dolt):
        self.bd = bd
        self.dolt = dolt
        self.path = os.path.realpath(tempfile.mkdtemp(prefix="bdp-call-parity."))
        self.home = os.path.join(self.path, "home")
        os.makedirs(self.home)
        self.procs = []
        self.env = {k: v for k, v in os.environ.items() if k not in DROPPED_ENV}
        self.env.update(TELEMETRY_ENV)
        self.env.update({
            "HOME": self.home,
            "PATH": os.path.dirname(dolt) + os.pathsep + os.environ.get("PATH", ""),
            "BEADS_DOLT_AUTO_START": "0",
            "BD_NON_INTERACTIVE": "1",
            "GIT_AUTHOR_NAME": "parity",
            "GIT_AUTHOR_EMAIL": "parity@example.test",
            "GIT_COMMITTER_NAME": "parity",
            "GIT_COMMITTER_EMAIL": "parity@example.test",
            "NO_COLOR": "1",
            "TZ": "UTC",
        })
        subprocess.run([dolt, "config", "--global", "--add", "metrics.disabled", "true"],
                       env=self.env, check=True, capture_output=True)
        subprocess.run([dolt, "config", "--global", "--add", "user.name", "parity"],
                       env=self.env, check=True, capture_output=True)
        subprocess.run([dolt, "config", "--global", "--add", "user.email", "parity@example.test"],
                       env=self.env, check=True, capture_output=True)

    def spawn(self, args, cwd, log):
        handle = open(log, "ab")
        proc = subprocess.Popen(args, cwd=cwd, env=self.env, stdout=handle, stderr=handle,
                                stdin=subprocess.DEVNULL, start_new_session=True)
        self.procs.append(proc)
        return proc

    def stop(self, proc):
        if proc.poll() is None:
            proc.terminate()
            try:
                proc.wait(10)
            except subprocess.TimeoutExpired:
                proc.kill()
                proc.wait(5)

    def own_daemons(self):
        """Pids of daemons under this root, found by command line then by cwd."""
        listing = subprocess.run(["ps", "-axo", "pid=,args="], capture_output=True, text=True).stdout
        found = []
        for line in listing.splitlines():
            pid, _, args = line.strip().partition(" ")
            if int(pid) == os.getpid():
                continue
            if self.path in args:
                found.append(int(pid))
            elif any(word in args for word in ("db-proxy-child", "dolt sql-server", " serve ")):
                # A daemon that does not name the root on its command line.
                cwd = subprocess.run(["lsof", "-a", "-d", "cwd", "-p", pid, "-Fn"],
                                     capture_output=True, text=True).stdout
                if self.path in cwd:
                    found.append(int(pid))
        return found

    def reap(self, keep=False):
        for proc in reversed(self.procs):
            self.stop(proc)
        for _ in range(3):
            pids = self.own_daemons()
            if not pids:
                break
            for pid in pids:
                try:
                    os.kill(pid, signal.SIGTERM)
                except ProcessLookupError:
                    pass
            time.sleep(1)
        for pid in self.own_daemons():
            try:
                os.kill(pid, signal.SIGKILL)
            except ProcessLookupError:
                pass
        if not keep:
            shutil.rmtree(self.path, ignore_errors=True)


class Stack:
    """One Dolt server and the workspace initialized against it."""

    def __init__(self, root, name, actor="parity", journal=False, serve=False):
        self.root = root
        self.name = name
        self.actor = actor
        self.dir = os.path.join(root.path, name)
        self.data = os.path.join(self.dir, "dolt")
        self.work = os.path.join(self.dir, "work")
        os.makedirs(self.data)
        os.makedirs(self.work)
        self.dolt_port = free_port()
        self.serve_port = None
        self.serve_proc = None
        self.serve_args = []
        self.serve_log = os.path.join(self.dir, "serve.log")
        self.dolt_proc = root.spawn([root.dolt, "sql-server", "--host", "127.0.0.1", "--port", str(self.dolt_port)],
                   self.data, os.path.join(self.dir, "dolt.log"))
        self._wait_port(self.dolt_port, "dolt sql-server")
        self._git_init(self.work)
        self.bd("init", "--server", "--external", "--server-host", "127.0.0.1", "--server-port",
                str(self.dolt_port), "--server-user", "root", "--prefix", "sw", "--quiet",
                "--non-interactive", check=True)
        if journal:
            self.bd("config", "set", "events-journal", "true", check=True)
        if serve:
            self.start_serve()

    def _git_init(self, path):
        subprocess.run(["git", "init", "-q"], cwd=path, env=self.root.env, check=True)

    def _wait_port(self, port, what, timeout=60):
        deadline = time.time() + timeout
        while time.time() < deadline:
            with socket.socket() as sock:
                sock.settimeout(0.5)
                if sock.connect_ex(("127.0.0.1", port)) == 0:
                    return
            time.sleep(0.2)
        raise RuntimeError(f"{what} on port {port} did not come up in {self.dir}")

    def env(self, actor=None):
        env = dict(self.root.env)
        env["BD_ACTOR"] = actor or self.actor
        return env

    def bd(self, *args, check=False, cwd=None, actor=None, timeout=90, stdin=None):
        return run_bd(self.root.bd, args, cwd or self.work, self.env(actor), check, timeout, stdin)

    def start_serve(self):
        self.serve_port = free_port()
        args = [self.root.bd, "serve", "--addr", f"127.0.0.1:{self.serve_port}"] + self.serve_args
        proc = subprocess.Popen(args, cwd=self.work, env=self.env(),
                                stdout=open(self.serve_log, "ab"), stderr=subprocess.STDOUT,
                                stdin=subprocess.DEVNULL, start_new_session=True)
        self.root.procs.append(proc)
        self.serve_proc = proc
        self._wait_port(self.serve_port, "bd serve")
        return self.serve_port

    def restart_serve(self):
        self.root.stop(self.serve_proc)
        return self.start_serve()

    @property
    def url(self):
        return f"http://127.0.0.1:{self.serve_port}"

    def connect_client(self, name="client", actor=None, url=None):
        path = os.path.join(self.dir, name)
        os.makedirs(path)
        self._git_init(path)
        run_bd(self.root.bd, ["connect", url or self.url], path, self.env(actor), True, 60, None)
        return path

    def http(self, method, path, body=None, headers=None, timeout=60):
        data = None
        all_headers = dict(headers or {})
        if body is not None:
            data = json.dumps(body).encode()
            all_headers.setdefault("Content-Type", "application/json")
        request = urllib.request.Request(self.url + path, data=data, method=method, headers=all_headers)
        try:
            with urllib.request.urlopen(request, timeout=timeout) as response:
                return Reply(response.status, dict(response.headers), response.read())
        except urllib.error.HTTPError as err:
            return Reply(err.code, dict(err.headers), err.read())

    def call(self, name, arguments, provenance=None, headers=None):
        body = {"arguments": arguments}
        if provenance is not None:
            body["revisionProvenance"] = provenance
        all_headers = {"Content-Type": CALLABLE_MEDIA}
        all_headers.update(headers or {})
        return self.http("POST", "/bdp/call/" + name, body, all_headers)


class Reply:
    def __init__(self, status, headers, raw):
        self.status = status
        self.headers = headers
        self.raw = raw
        try:
            self.json = json.loads(raw) if raw else None
        except ValueError:
            self.json = {"unparsed": raw[:400].decode(errors="replace")}


class Ran:
    def __init__(self, code, out, err):
        self.code = code
        self.out = out
        self.err = err


def run_bd(bd, args, cwd, env, check, timeout, stdin):
    args = [str(a) for a in args]
    try:
        done = subprocess.run([bd] + args, cwd=cwd, env=env, capture_output=True, text=True,
                              timeout=timeout, input=stdin if stdin is not None else "")
        ran = Ran(done.returncode, done.stdout, done.stderr)
    except subprocess.TimeoutExpired as err:
        ran = Ran(-1, (err.stdout or b"").decode(errors="replace") if isinstance(err.stdout, bytes) else (err.stdout or ""),
                  "TIMEOUT after %ss" % timeout)
    if check and ran.code != 0:
        raise RuntimeError(f"bd {' '.join(args)} in {cwd} failed ({ran.code}): {ran.out} {ran.err}")
    return ran
