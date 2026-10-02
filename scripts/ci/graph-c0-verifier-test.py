#!/usr/bin/env python3
"""Negative controls for the required Go-test receipt gate; no engine needed."""
import importlib.util
import json
import os
import re
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


class RequiredEngineControls(unittest.TestCase):
    def exercise(self, engines):
        # Drive the complete discovery/execution gate with a source root and
        # valid complete package receipts. A missing engine never emits a skip,
        # so ordinary run/pass equality alone cannot catch the omission.
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            source = root / "workflow_test.go"
            source.write_text("func TestWorkflow(t *testing.T) {}\n")
            events = [event("run", "TestWorkflow")]
            for engine in engines:
                events += [event("run", "TestWorkflow/" + engine), event("pass", "TestWorkflow/" + engine)]
            events += [event("pass", "TestWorkflow"), event("pass")]
            q = module.Qualification.__new__(module.Qualification)
            q.counts = {}
            listing = b"TestWorkflow\n"
            execution = "\n".join(json.dumps(item) for item in events).encode()
            with mock.patch.object(q, "run", side_effect=[listing, execution]):
                q.tests("./fixture", "^TestWorkflow", [source], "cli",
                        ("TestWorkflow/embedded", "TestWorkflow/server"))
            self.assertEqual(q.counts, {"cli": 1})

    def test_both_installed_engines_complete(self):
        self.exercise(("embedded", "server"))

    def test_missing_installed_engine_refused(self):
        for engines in [("embedded",), ("server",), ()]:
            with self.subTest(engines=engines), self.assertRaisesRegex(RuntimeError, "required engine proof absent"):
                self.exercise(engines)


class RequiredGroupedExecution(unittest.TestCase):
    def exercise(self, mode):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            names = ["Test" + letter for letter in "ABCDEFGH"]
            source = root / "storage_test.go"
            source.write_text("\n".join("func " + name + "(t *testing.T) {}" for name in names))
            q = module.Qualification.__new__(module.Qualification)
            q.output, q.counts = root, {}
            calls = []
            def run(args, label, **kwargs):
                calls.append((args, label, kwargs))
                if "-list" in args:
                    listed = names + ([names[0]] if mode == "duplicate-discovery" else [])
                    return ("\n".join(listed) + "\n").encode()
                selected = args[args.index("-run") + 1][2:-2].split("|")
                if mode == "missing": selected = selected[1:]
                if mode == "duplicate": selected += selected[:1]
                events = []
                for name in selected:
                    events += [event("run", name), event("run", name + "/embedded"), event("pass", name + "/embedded"), event("pass", name)]
                events += [event("fail" if mode == "failure" else "pass")]
                raw = ("\n".join(json.dumps(item) for item in events) + "\n").encode()
                (root / (label + ".stdout")).write_bytes(raw)
                (root / (label + ".stderr")).write_bytes(b"")
                if mode == "timeout": raise subprocess.TimeoutExpired(args, 960)
                return raw
            with mock.patch.object(q, "run", run):
                if mode == "good":
                    q.tests("./fixture", "^Test", [source], "storage", groups=4)
                else:
                    with self.assertRaises((RuntimeError, subprocess.TimeoutExpired)):
                        q.tests("./fixture", "^Test", [source], "storage", groups=4)
            if mode == "duplicate-discovery":
                self.assertEqual(len(calls), 1)
                return
            plan = json.loads((root / "storage-groups.json").read_text())
            self.assertEqual(plan["complete"], mode == "good")
            self.assertEqual(plan["sourceRoots"], sorted(names))
            self.assertEqual(plan["compiledRoots"], sorted(names))
            self.assertEqual(len(plan["groups"]), 4)
            parts = [root / (group["label"] + ".stdout") for group in plan["groups"]]
            self.assertEqual((root / "storage.stdout").read_bytes(), b"".join(path.read_bytes() for path in parts if path.exists()))
            for args, label, kwargs in calls[1:]:
                self.assertIn("-count=1", args)
                self.assertIn("-timeout=15m", args)
                self.assertIn("-p=1", args)
                self.assertIn("-parallel=1", args)
                self.assertEqual(kwargs["timeout"], 960)
            if mode == "good":
                self.assertEqual(q.counts, {"storage": 8})
                self.assertTrue(all(group["passed"] for group in plan["groups"]))
                self.assertEqual(len(calls), 5)
                self.assertEqual(module.verify_tests([json.loads(line) for line in (root / "storage.stdout").read_bytes().splitlines()], set(names)), 8)
                self.assertEqual(plan["stdoutSHA256"], module.digest(root / "storage.stdout"))
            else:
                self.assertEqual(q.counts, {})
                self.assertFalse(any(group["passed"] for group in plan["groups"]))

    def test_exhaustive_groups_and_actual_concatenation(self): self.exercise("good")

    def test_missing_duplicate_failed_or_timed_out_group_refused(self):
        for mode in ["missing", "duplicate", "failure", "timeout", "duplicate-discovery"]:
            with self.subTest(mode=mode): self.exercise(mode)

    def test_zero_duplicate_and_empty_group_selection_refused(self):
        for names, groups in [([], 4), (["TestA", "TestA"], 2), (["TestA"], 4)]:
            with self.subTest(names=names), self.assertRaises(RuntimeError):
                module.partition_required_tests(names, groups)
        names = ["Test" + letter for letter in "ABCDEFGH"]
        self.assertEqual(module.partition_required_tests(names, 4), module.partition_required_tests(list(reversed(names)), 4))


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


class SmokeDeadlines(unittest.TestCase):
    # The HTTP/BDP smoke starts with a server-backed `bd init`. Hosted runners differ by
    # about 2x: that init took 34 s on a fast runner and 58 s and 65 s on slow ones, and the
    # smoke's 60 s per-command default killed it on both slow attempts. Pin the deadlines the
    # qualification really hands the smoke, parsed the way the smoke parses them (so a dropped
    # flag falls back to the smoke's own default and is judged as such), and the job budget
    # around them, so none can slide back under a slow runner's measured times.
    SLOWEST_INIT_SECONDS = 65
    INIT_HEADROOM = 2.5
    # The wrapper must outlast the smoke's whole run by the time the smoke needs to reap its
    # children and report its own labelled failure; the original margin was this much.
    WRAPPER_MARGIN_SECONDS = 60
    JOB_NAME = "Graph core / real engines, installed CLI and BDP"
    # Whole-job durations on hosted runners: 27 min on a fast one, 43 to 44 min on five slow
    # ones (each stopped at the smoke's first command), and one slow runner whose job was cut
    # off by the 50 min cap 11 s into the smoke. That last one is a floor for what a complete
    # slow run needs; the headroom covers the smoke and upload it never reached and a runner
    # slower still.
    SLOWEST_JOB_MINUTES = 51
    JOB_HEADROOM = 1.75

    @staticmethod
    def job_timeout_minutes(name):
        # Plain text on purpose: the lane that runs this file has no YAML dependency.
        lines = (Path(__file__).resolve().parents[2] / ".github/workflows/pr.yml").read_text().splitlines()
        headers = [i for i, line in enumerate(lines) if re.fullmatch(r"  [\w-]+:\s*", line)]
        named = [i for i, line in enumerate(lines) if line == "    name: " + name]
        if len(named) != 1:
            raise AssertionError(f"expected exactly one job named {name!r}, found {len(named)}")
        start = max(i for i in headers if i < named[0])
        end = min([i for i in headers if i > named[0]] + [len(lines)])
        found = [m for m in (re.fullmatch(r"    timeout-minutes:\s*(\d+)\s*", line) for line in lines[start:end]) if m]
        if len(found) != 1:
            raise AssertionError(f"expected exactly one timeout-minutes on job {name!r}, found {len(found)}")
        return int(found[0].group(1))

    def test_slow_runner_deadlines_nest_inside_the_job_budget(self):
        smoke_path = Path(__file__).resolve().parents[1] / "graph-bdp-read-smoke.py"
        spec = importlib.util.spec_from_file_location("graph_bdp_read_smoke", smoke_path)
        smoke = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(smoke)
        q = module.Qualification.__new__(module.Qualification)
        q.root, q.output = Path("/repo"), Path("/out")
        q.bd, q.node = Path("/bin/bd"), Path("/bin/node")
        q.client_checkout, q.client_manifest = Path("/client"), Path("/client.json")
        argv, wrapper = q.smoke_command(0)
        self.assertEqual(Path(argv[1]), Path("/repo/scripts/graph-bdp-read-smoke.py"))
        args = smoke.build_parser().parse_args(argv[2:])
        command, total = args.command_timeout, args.total_timeout
        with self.subTest("one command outlasts a slow runner's server-backed init"):
            floor = self.INIT_HEADROOM * self.SLOWEST_INIT_SECONDS
            self.assertGreaterEqual(command, floor,
                                    f"per-command deadline {command}s is under {self.INIT_HEADROOM}x the slowest measured init ({self.SLOWEST_INIT_SECONDS}s)")
        with self.subTest("per-command < total < wrapper, with the wrapper's cleanup margin"):
            self.assertLess(command, total, "a single command may not outlast the smoke's whole run")
            self.assertGreaterEqual(wrapper - total, self.WRAPPER_MARGIN_SECONDS,
                                    f"wrapper {wrapper}s must outlast the smoke total {total}s by {self.WRAPPER_MARGIN_SECONDS}s so the smoke reports its own failure")
        with self.subTest("the job's own cap outlasts a slow runner's whole job"):
            minutes = self.job_timeout_minutes(self.JOB_NAME)
            floor = self.SLOWEST_JOB_MINUTES * self.JOB_HEADROOM
            self.assertGreaterEqual(minutes, floor,
                                    f"job cap {minutes} min is under {self.JOB_HEADROOM}x the slowest measured whole job ({self.SLOWEST_JOB_MINUTES} min)")
            self.assertGreater(minutes * 60, wrapper, "the job cap must outlast the smoke wrapper")


if __name__ == "__main__":
    unittest.main()
