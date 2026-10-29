"""Read-only witness for this ONE canonical local-server fixture on Linux.

No process learned here is ever signalled or reaped. Directory and proc-net
file descriptors retain the observed filesystem/network views for postchecks.
"""
import hashlib
import os
from pathlib import Path
import re
import time


def birth(pid):
    tail = Path(f"/proc/{pid}/stat").read_text().rsplit(") ", 1)[1].split()
    return int(tail[1]), tail[19]  # ppid, Linux stat field22/starttime


def table(fd):
    os.lseek(fd, 0, os.SEEK_SET)
    chunks = []
    while True:
        chunk = os.read(fd, 65536)
        if not chunk:
            break
        chunks.append(chunk)
    return b"".join(chunks).decode().splitlines()[1:]


class RoutingFixtureObserver:
    def __init__(self, namespace, test_hash, engine_hash):
        self.namespace = str(namespace).encode()
        self.test_hash = test_hash
        self.engine_hash = engine_hash
        self.test = None
        self.server = None
        self.parent_fd = None
        self.net_fd = None
        self.listener = False
        self.error = None
        self.last_observed = 0.0
        self.engine_hashes = {}
        self.engine_hash_attempts = 0
        self.engine_hash_completions = 0

    def observe(self):
        # A read-only witness must never terminate the live test it observes.
        # Missing/failed witness still prevents a green diagnostic afterward.
        now = time.monotonic()
        if self.error or now - self.last_observed < 1.0:
            return
        self.last_observed = now
        try:
            self._observe()
        except Exception as error:
            self.error = {"type": type(error).__name__, "message": str(error)}

    def _observe(self):
        if self.listener:
            return
        if self.test is None:
            for entry in Path("/proc").glob("[0-9]*"):
                try:
                    if entry.stat().st_uid != os.geteuid():
                        continue
                    fields = (entry / "environ").read_bytes().split(b"\0")
                    variables = dict(field.split(b"=", 1) for field in fields if b"=" in field)
                    # rules_go reexecs the same bd_test executable for XML.
                    # Its inner test owner explicitly receives GO_TEST_WRAP=0;
                    # the outer XML wrapper must never be latched as the owner.
                    if variables.get(b"GO_TEST_WRAP") != b"0":
                        continue
                    if self.namespace not in variables.get(b"TEST_TMPDIR", b""):
                        continue
                    if not variables.get(b"TMPDIR"):
                        continue
                    if not os.readlink(entry / "exe").endswith("/bd_test"):
                        continue
                    if hashlib.sha256((entry / "exe").read_bytes()).hexdigest() != self.test_hash:
                        raise RuntimeError("scoped test executable changed")
                    pid = int(entry.name)
                    _, started = birth(pid)
                    self.test = {"pid": pid, "birth": started,
                                 "initialTMPDIR": os.fsdecode(variables[b"TMPDIR"])}
                    break
                except (FileNotFoundError, ProcessLookupError):
                    continue
        if self.test is None:
            return
        pid = self.test["pid"]
        if birth(pid)[1] != self.test["birth"]:
            raise RuntimeError("test identity disappeared before fixture witness")
        children = set()
        for task in Path(f"/proc/{pid}/task").iterdir():
            try:
                children.update((task / "children").read_text().split())
            except FileNotFoundError:
                pass
        for child in children:
            if self._witness_child(pid, int(child)):
                return

    def _engine_digest(self, proc):
        # Share a completed hash across processes using the exact same file.
        # Process parent/birth identities are checked separately at commitment.
        exe = proc / "exe"
        stat = exe.stat()
        identity = (stat.st_dev, stat.st_ino, stat.st_size, stat.st_mtime_ns, stat.st_ctime_ns)
        if identity not in self.engine_hashes:
            self.engine_hash_attempts += 1  # Includes partial/failed reads.
            value = hashlib.sha256(exe.read_bytes()).hexdigest()
            self.engine_hash_completions += 1
            after = exe.stat()
            if identity != (after.st_dev, after.st_ino, after.st_size, after.st_mtime_ns, after.st_ctime_ns):
                raise RuntimeError("engine executable identity changed during hashing")
            self.engine_hashes[identity] = value
        return self.engine_hashes[identity]

    def _witness_child(self, test_pid, child):
        # A fixture may lose the free-port/bind race and legitimately retry.
        # Keep each candidate and its FDs provisional until a live listener
        # belongs to that exact child. Never retain an unvalidated candidate.
        parent_fd = net_fd = None
        try:
            proc = Path("/proc") / str(child)
            argv = (proc / "cmdline").read_bytes().rstrip(b"\0").split(b"\0")
            if len(argv) != 4 or argv[1:3] != [b"sql-server", b"--config"]:
                return False
            parent, started = birth(child)
            if parent != test_pid:
                return False
            if self._engine_digest(proc) != self.engine_hash:
                raise RuntimeError("fixture engine executable differs from pinned input")
            config = Path(os.fsdecode(argv[3]))
            base = Path(self.test["initialTMPDIR"])
            if config.name != "server.yaml" or config.parent.parent != base:
                raise RuntimeError("unexpected canonical server root")
            if not re.fullmatch(r"bdt-dolt-[A-Za-z0-9]+", config.parent.name):
                raise RuntimeError("unexpected fixture root name")
            view = proc / "root" / str(config).lstrip("/")
            if (view.parent / "owner.pid").read_text().strip() != str(test_pid):
                raise RuntimeError("fixture owner marker differs from live exact test")
            text = view.read_text()
            match = re.search(r"(?m)^  host: 127\.0\.0\.1\n  port: ([0-9]+)$", text)
            if not match:
                raise RuntimeError("unexpected canonical listener config")
            parent_fd = os.open(view.parent.parent, os.O_RDONLY | os.O_DIRECTORY | os.O_CLOEXEC)
            net_fd = os.open(proc / "net/tcp", os.O_RDONLY | os.O_CLOEXEC)
            candidate = {"pid": child, "birth": started,
                         "rootName": config.parent.name, "port": int(match[1]),
                         "configSHA256": hashlib.sha256(text.encode()).hexdigest()}
            sockets = set()
            for fd in (proc / "fd").iterdir():
                try:
                    socket = re.fullmatch(r"socket:\[([0-9]+)\]", os.readlink(fd))
                    if socket:
                        sockets.add(socket[1])
                except FileNotFoundError:
                    pass
            address = f"0100007F:{candidate['port']:04X}"
            listening = any(row.split()[1] == address and row.split()[3] == "0A"
                            and row.split()[9] in sockets for row in table(net_fd))
            if not listening:
                return False
            # Revalidate identities at commitment, after reading the listener.
            if birth(test_pid)[1] != self.test["birth"]:
                raise RuntimeError("test identity changed before fixture witness")
            if birth(child) != (test_pid, started):
                return False
            if self._engine_digest(proc) != self.engine_hash:
                raise RuntimeError("fixture engine changed before witness commitment")
            if birth(child) != (test_pid, started):
                return False
            self.server = candidate
            self.parent_fd, self.net_fd = parent_fd, net_fd
            self.listener = True
            parent_fd = net_fd = None  # Ownership transfers only at commitment.
            return True
        except (FileNotFoundError, ProcessLookupError):
            return False
        finally:
            for fd in [parent_fd, net_fd]:
                if fd is not None:
                    os.close(fd)

    def finish(self):
        report = {"test": self.test, "server": self.server, "ownedListenerWitness": self.listener,
                  "borrowedPIDSignals": 0, "borrowedPIDReaps": 0,
                  "observerError": self.error,
                  "engineDigestScope": "completed content digest memoized by device/inode/size/mtime_ns/ctime_ns across live processes; commitment re-stats file identity and compares cached digest, not a second content read; completions count complete reads even when post-read identity rejects caching; parent/birth identity checked separately",
                  "engineHashAttempts": self.engine_hash_attempts,
                  "engineHashCompletions": self.engine_hash_completions,
                  "engineHashCacheEntries": len(self.engine_hashes),
                  "normalFixtureCleanupProven": False,
                  "cleanupObservationScope": "post canonical run_under kill/remove trap and owned Bazel shutdown; not a proof of normal fixture cleanup"}
        try:
            if self.error:
                raise RuntimeError("fixture witness failed; underlying test verdict remains separate")
            if not self.server or not self.listener:
                raise RuntimeError("no exact canonical fixture/listener witness")
            try:
                gone = birth(self.server["pid"])[1] != self.server["birth"]
            except (FileNotFoundError, ProcessLookupError):
                gone = True
            try:
                os.stat(self.server["rootName"], dir_fd=self.parent_fd, follow_symlinks=False)
                removed = False
            except FileNotFoundError:
                removed = True
            address = f"0100007F:{self.server['port']:04X}"
            closed = not any(row.split()[1] == address and row.split()[3] == "0A"
                             for row in table(self.net_fd))
            report.update(sameBirthServerGone=gone, ownedRootRemoved=removed, listenerClosed=closed)
            if not (gone and removed and closed):
                raise RuntimeError("owned fixture/listener postconditions failed")
        except Exception as error:
            report["error"] = str(error)
        finally:
            for fd in [self.parent_fd, self.net_fd]:
                if fd is not None:
                    os.close(fd)
        return report
