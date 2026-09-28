#!/usr/bin/env python3
"""Negative controls for the required Go-test receipt gate; no engine needed."""
import importlib.util
import subprocess
import sys
import threading
from pathlib import Path
import unittest

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


if __name__ == "__main__":
    unittest.main()
