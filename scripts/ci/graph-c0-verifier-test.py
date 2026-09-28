#!/usr/bin/env python3
"""Negative controls for the required Go-test receipt gate; no engine needed."""
import importlib.util
import sys
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


if __name__ == "__main__":
    unittest.main()
