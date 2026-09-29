#!/usr/bin/env python3
"""Negative controls for the required Go-test receipt gate; no engine needed."""
import importlib.util
import json
import os
import signal
import socket
import tempfile
import time
import subprocess
import sys
import threading
from pathlib import Path
import unittest
from unittest import mock

sys.dont_write_bytecode = True

spec = importlib.util.spec_from_file_location("graph_c0", Path(__file__).with_name("graph-c0-qualify.py"))
module = importlib.util.module_from_spec(spec)
spec.loader.exec_module(module)


def event(action, test=None):
    return dict(Action=action, **({"Test": test} if test else {}))


class RequiredReceipts(unittest.TestCase):
    def setUp(self):
        self.good = [event("run", "TestRequired"), event("run", "TestRequired/child"),
                     event("pass", "TestRequired/child"), event("pass", "TestRequired"), event("pass")]

    def test_complete(self):
        self.assertEqual(module.verify_tests(self.good, {"TestRequired"}), 1)

    def test_no_tests(self):
        with self.assertRaises(RuntimeError):
            module.verify_tests([event("pass")], set())

    def test_skipped_subtest(self):
        with self.assertRaises(RuntimeError):
            module.verify_tests(self.good + [event("skip", "TestRequired/optional")], {"TestRequired"})

    def test_missing_subtest_pass(self):
        with self.assertRaises(RuntimeError):
            module.verify_tests(self.good + [event("run", "TestRequired/unfinished")], {"TestRequired"})

    def test_extra_subtest_pass(self):
        with self.assertRaises(RuntimeError):
            module.verify_tests(self.good + [event("pass", "TestRequired/never-started")], {"TestRequired"})

    def test_extra_subtest_root(self):
        with self.assertRaises(RuntimeError):
            module.verify_tests(self.good + [event("run", "TestOther/child"), event("pass", "TestOther/child")], {"TestRequired"})

    def test_duplicate_subtest_pass(self):
        with self.assertRaises(RuntimeError):
            module.verify_tests(self.good + [event("pass", "TestRequired/child")], {"TestRequired"})

    def test_missing_new_test(self):
        with self.assertRaises(RuntimeError):
            module.verify_tests(self.good, {"TestRequired", "TestNew"})

    def test_truncated_package(self):
        with self.assertRaises(RuntimeError):
            module.verify_tests(self.good[:-1], {"TestRequired"})

    def test_failed_package(self):
        with self.assertRaises(RuntimeError):
            module.verify_tests(self.good + [event("fail")], {"TestRequired"})

    def test_cached_without_execution(self):
        with self.assertRaises(RuntimeError):
            module.verify_tests([event("pass", "TestRequired"), event("pass")], {"TestRequired"})


class RequiredDiscovery(unittest.TestCase):
    # Exercise the actual source/compiled/execution boundary with temporary Go
    # source. Only subprocess output is supplied; no engine or Go build is needed.
    def exercise(self, source, compiled, events):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "patch_test.go").write_text(source)
            q = module.Qualification.__new__(module.Qualification)
            q.counts = {}
            calls = []

            def run(argv, label, **kwargs):
                calls.append((argv, label))
                if label == "graphpatch-discovery":
                    return compiled.encode()
                self.assertEqual(label, "graphpatch")
                return "\n".join(json.dumps(e) for e in events).encode()

            q.run = run
            q.tests("./internal/graphpatch", "^Test", root.glob("*_test.go"), "graphpatch")
            self.assertEqual(q.counts, {"graphpatch": 1})
            self.assertEqual([label for _, label in calls], ["graphpatch-discovery", "graphpatch"])
            self.assertTrue(all(argv[-1] == "./internal/graphpatch" for argv, _ in calls))
            self.assertIn("-count=1", calls[1][0])

    def test_required_package_discovery_and_execution(self):
        self.exercise("package graphpatch\nfunc TestPatch(t *testing.T) {}\n", "TestPatch\n",
                      [event("run", "TestPatch"), event("run", "TestPatch/child"),
                       event("pass", "TestPatch/child"), event("pass", "TestPatch"), event("pass")])

    def test_omitted_or_unexecuted_required_source_refuses(self):
        source = "package graphpatch\nfunc TestPatch(t *testing.T) {}\n"
        good = [event("run", "TestPatch"), event("pass", "TestPatch"), event("pass")]
        for name, text, compiled, events in [
            ("empty-source", "package graphpatch\n", "", [event("pass")]),
            ("compiled-excludes-source", source, "", good),
            ("source-glob-omits-compiled-root", "package graphpatch\n", "TestPatch\n", good),
            ("compiled-adds-unregistered-root", source, "TestPatch\nTestOther\n", good),
            ("discovered-but-not-executed", source, "TestPatch\n", [event("pass")]),
            ("discovered-but-skipped", source, "TestPatch\n", [event("skip", "TestPatch"), event("pass")]),
        ]:
            with self.subTest(name=name), self.assertRaises(RuntimeError):
                self.exercise(text, compiled, events)


class HTTPReceiptControls(unittest.TestCase):
    def test_clean_listener(self):
        module.verify_clean_receipt(dict(exit_code=0, failure=None, forced=False,
                                         group_gone=True, port_closed=True), listener=True)

    def test_forced_or_incomplete_cleanup_refused(self):
        good = dict(exit_code=0, failure=None, forced=False, group_gone=True, port_closed=True)
        for change in [dict(forced=True), dict(group_gone=False), dict(port_closed=False),
                       dict(exit_code=1), dict(failure="timeout")]:
            with self.subTest(change=change), self.assertRaises(RuntimeError):
                module.verify_clean_receipt({**good, **change}, listener=True)

    def test_missing_cleanup_field_refused(self):
        with self.assertRaises(RuntimeError):
            module.verify_clean_receipt(dict(exit_code=0, failure=None), listener=True)


class ProcessCleanup(unittest.TestCase):
    def test_live_group_is_not_clean(self):
        child = subprocess.Popen([sys.executable, "-c", "import time; time.sleep(30)"], start_new_session=True)
        try:
            self.assertFalse(module.group_exited(child.pid, timeout=0.01))
        finally:
            child.kill()
            child.wait(timeout=5)
        self.assertTrue(module.group_exited(child.pid, timeout=0.01))

    def test_exiting_group_gets_bounded_grace(self):
        child = subprocess.Popen([sys.executable, "-c", "import time; time.sleep(0.15)"], start_new_session=True)
        reaper = threading.Thread(target=child.wait)
        reaper.start()
        try:
            self.assertTrue(module.group_exited(child.pid, timeout=5))
        finally:
            if child.poll() is None:
                child.kill()
            reaper.join(timeout=5)
        self.assertFalse(reaper.is_alive())


class NestedOwnershipControls(unittest.TestCase):
    # Real independent process group + bound socket, using the transferred
    # capture's Process and registry. No database or mocked process receipt.
    OWNER = r'''
import argparse, importlib.util, json, signal, sys, time
from pathlib import Path
sys.dont_write_bytecode = True
spec = importlib.util.spec_from_file_location("smoke", sys.argv[1])
smoke = importlib.util.module_from_spec(spec)
spec.loader.exec_module(smoke)
registry, ready, output = map(Path, sys.argv[2:5])
args = argparse.Namespace(output_dir=output, bd=Path(sys.executable), server_port=0,
    server_root=None, total_timeout=120, command_timeout=60, owned_groups=registry)
capture = smoke.c0.Capture(args)
def interrupted(*_):
    raise KeyboardInterrupt("test outer interruption")
signal.signal(signal.SIGTERM, interrupted)
child_code = """import json, os, signal, socket, sys
from pathlib import Path
sock=socket.socket(); sock.bind(('127.0.0.1',0)); sock.listen()
def stop(*_):
    sock.close(); sys.exit(0)
signal.signal(signal.SIGINT, stop); signal.signal(signal.SIGTERM, stop)
Path(sys.argv[1]).write_text(json.dumps({'pid':os.getpid(),'port':sock.getsockname()[1]}))
signal.pause()
"""
process = None
try:
    process = smoke.Process(capture, "listener", [sys.executable, "-c", child_code, str(ready)])
    while not ready.exists():
        time.sleep(0.01)
    process.port = json.loads(ready.read_text())["port"]
    ready.with_suffix(".owner").write_text("ready")
    time.sleep(60)
except KeyboardInterrupt:
    pass
finally:
    if process:
        process.close()
    if capture.active:
        capture.stop()
'''

    def exercise(self, interrupt):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            owner, registry, ready = root / "owner.py", root / "groups.jsonl", root / "ready.json"
            owner.write_text(self.OWNER)
            q = module.Qualification.__new__(module.Qualification)
            q.root, q.output, q.env = root, root, dict(os.environ, PYTHONDONTWRITEBYTECODE="1")
            q.children, q.receipts = set(), []
            argv = [sys.executable, str(owner), str(Path(__file__).resolve().parents[1] / "graph-bdp-read-smoke.py"),
                    str(registry), str(ready), str(root / "capture")]
            thread = None
            previous = None
            if interrupt:
                def receive(*_):
                    raise RuntimeError("real signal interruption control")
                previous = signal.signal(signal.SIGUSR1, receive)
                def send_when_listening():
                    deadline = time.monotonic() + 5
                    while not ready.with_suffix(".owner").exists() and time.monotonic() < deadline:
                        time.sleep(0.01)
                    if ready.with_suffix(".owner").exists():
                        os.kill(os.getpid(), signal.SIGUSR1)
                thread = threading.Thread(target=send_when_listening)
                thread.start()
            try:
                with self.assertRaises(RuntimeError if interrupt else subprocess.TimeoutExpired):
                    q.run(argv, "nested", timeout=10 if interrupt else 3, owned_groups=registry)
                self.assertTrue(ready.exists(), "real listener never started")
                info = json.loads(ready.read_text())
                self.assertTrue(module.group_exited(info["pid"], timeout=0.1))
                with self.assertRaises(ConnectionRefusedError):
                    socket.create_connection(("127.0.0.1", info["port"]), timeout=1)
                self.assertEqual(q.children, set())
                self.assertEqual(q.receipts[-1]["failure"], "RuntimeError" if interrupt else "TimeoutExpired")
                self.assertFalse(q.receipts[-1]["forcedOwner"])
                receipt = json.loads((root / "capture/listener/receipt.json").read_text())
                module.verify_clean_receipt(receipt, listener=True)
            finally:
                if thread:
                    thread.join(timeout=6)
                if previous is not None:
                    signal.signal(signal.SIGUSR1, previous)
                # Failures must not leave the negative control itself alive.
                module.cleanup_registered_groups(registry)

    def test_outer_timeout_drains_independent_listener(self):
        self.exercise(interrupt=False)

    def test_outer_interruption_drains_independent_listener(self):
        self.exercise(interrupt=True)

    def test_registry_failure_cannot_leave_unresponsive_owner_alive(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            owner, registry, ready = root / "owner.py", root / "groups.jsonl", root / "ready.json"
            # Ignore outer termination but reap real children when they stop.
            # This isolates an unresponsive owner from orphan-zombie behavior.
            source = self.OWNER.replace("import argparse, importlib.util, json, signal, sys, time",
                                        "import argparse, importlib.util, json, os, signal, sys, time")
            source = source.replace("signal.signal(signal.SIGTERM, interrupted)",
                "signal.signal(signal.SIGTERM, signal.SIG_IGN)\n"
                "def reap(*_):\n"
                "    try:\n"
                "        while os.waitpid(-1, os.WNOHANG)[0]: pass\n"
                "    except ChildProcessError: pass\n"
                "signal.signal(signal.SIGCHLD, reap)")
            owner.write_text(source)
            q = module.Qualification.__new__(module.Qualification)
            q.root, q.output, q.env = root, root, dict(os.environ, PYTHONDONTWRITEBYTECODE="1")
            q.children, q.receipts = set(), []
            argv = [sys.executable, str(owner), str(Path(__file__).resolve().parents[1] / "graph-bdp-read-smoke.py"),
                    str(registry), str(ready), str(root / "capture")]
            actual_cleanup = module.cleanup_registered_groups
            calls = []
            def fail_first_registry_pass(path):
                result = actual_cleanup(path)
                calls.append(result)
                if len(calls) == 1:
                    raise RuntimeError("injected registry failure after real child cleanup")
                return result
            try:
                # Only the grace duration and injected registry error vary;
                # owner, independent listener, sockets and signals are real.
                with mock.patch.object(module, "OWNER_CLEANUP_GRACE", 0.2), \
                     mock.patch.object(module, "cleanup_registered_groups", fail_first_registry_pass), \
                     self.assertRaises(subprocess.TimeoutExpired):
                    q.run(argv, "forced-owner", timeout=3, owned_groups=registry)
                info = json.loads(ready.read_text())
                receipt = q.receipts[-1]
                self.assertTrue(module.group_exited(receipt["pid"], timeout=0.1))
                self.assertTrue(module.group_exited(info["pid"], timeout=0.1))
                with self.assertRaises(ConnectionRefusedError):
                    socket.create_connection(("127.0.0.1", info["port"]), timeout=1)
                self.assertEqual(receipt["failure"], "TimeoutExpired")
                self.assertTrue(receipt["forcedOwner"])
                self.assertEqual(receipt["exit"], -signal.SIGKILL)
                self.assertTrue(any("injected registry failure" in error for error in receipt["cleanupErrors"]))
                self.assertGreaterEqual(len(calls), 2)
                self.assertTrue(receipt["nestedCleanup"]["groupsGone"])
                self.assertEqual(q.children, set())
            finally:
                actual_cleanup(registry)


if __name__ == "__main__":
    unittest.main()
