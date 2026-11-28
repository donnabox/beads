"""Pure generator mechanics tests; every subprocess is mocked, never executed."""
import contextlib
import hashlib
import io
import json
import os
from pathlib import Path
import runpy
import subprocess
import sys
import tempfile
import unittest
from unittest import mock


GENERATOR = Path(__file__).resolve().parents[3] / "scripts" / "legacy-import-fixture.py"
SOURCE = "0123456789abcdef0123456789abcdef01234567"
TREE = "abcdef0123456789abcdef0123456789abcdef0123"
EXPORTED = '{"_type":"issue","id":"legacy-a","title":"記憶"}\n{"_type":"memory","key":"fixture-key","value":"synthetic"}\n'


class GeneratorTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory(prefix="iris-mocked-generator-")
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name).resolve()
        self.source = self.root / "fake-checkout"
        self.source.mkdir()
        self.binary = self.root / "fake-bd"
        # This is deliberately not an executable producer. Missing subprocess
        # mocks cannot accidentally invoke a real bd binary or create a store.
        self.binary.write_bytes(b"disposable mocked producer bytes")
        self.output_parent = self.root / "fixtures"
        self.output_parent.mkdir()
        self.output = self.output_parent / "fixture"
        self.commands = []
        self.git_reads = []
        self.version = {"version": "1.3.0", "build": SOURCE[:9], "commit": SOURCE}
        self.export_failure = False
        self.binary_change = False
        self.source_change = False
        self.create_existing = False
        self.did_export = False

    def fake_check_output(self, args, **kwargs):
        self.assertEqual(Path(kwargs["cwd"]), self.source)
        self.assertTrue(kwargs["text"])
        self.assertEqual(kwargs["timeout"], 30)
        self.assertFalse(any(key.startswith("GIT_") for key in kwargs["env"]))
        self.git_reads.append(args)
        if args == ["git", "rev-parse", "HEAD"]:
            return ("f" * 40 if self.source_change and self.did_export else SOURCE) + "\n"
        if args == ["git", "rev-parse", "HEAD^{tree}"]:
            return TREE + "\n"
        if args == ["git", "status", "--porcelain"]:
            return ""
        raise AssertionError(f"unmocked source invocation: {args}")

    def fake_run(self, args, **kwargs):
        self.commands.append(args)
        if args[0] == "git":
            self.assertIn(args, [["git", "init", "-q"], ["git", "config", "core.hooksPath", "/dev/null"]])
            self.assertTrue(kwargs["check"])
            return subprocess.CompletedProcess(args, 0)
        self.assertEqual(args[0], str(self.binary))
        self.assertEqual(kwargs["timeout"], 90)
        self.assertTrue(kwargs["capture_output"] and kwargs["text"])
        self.assertNotEqual(Path(kwargs["cwd"]), self.source)
        command = args[1:]
        if command == ["version", "--json"]:
            return subprocess.CompletedProcess(args, 0, json.dumps(self.version) + "\n", "")
        if command == ["version"]:
            return subprocess.CompletedProcess(args, 0, f"bd version 1.3.0 ({SOURCE[:9]})\n", "")
        if command == ["export", "--include-memories"]:
            self.did_export = True
            if self.binary_change:
                self.binary.write_bytes(b"changed producer bytes")
            if self.create_existing:
                self.output.mkdir()
                (self.output / "producer.json").write_bytes(b"preexisting receipt")
            return subprocess.CompletedProcess(args, 1 if self.export_failure else 0, EXPORTED, "simulated export failure" if self.export_failure else "")
        self.assertIn(command[0], {"init", "create", "dep", "comments", "close", "remember", "config"})
        return subprocess.CompletedProcess(args, 0, "synthetic command receipt\n", "")

    def invoke(self, profile="basic"):
        stdout, stderr = io.StringIO(), io.StringIO()
        argv = [str(GENERATOR), "--bd", str(self.binary), "--source-dir", str(self.source), "--output", str(self.output), "--profile", profile]
        error = None
        with mock.patch.object(sys, "argv", argv), \
                mock.patch("subprocess.check_output", side_effect=self.fake_check_output), \
                mock.patch("subprocess.run", side_effect=self.fake_run), \
                mock.patch("subprocess.Popen", side_effect=AssertionError("real subprocess forbidden")), \
                contextlib.redirect_stdout(stdout), contextlib.redirect_stderr(stderr):
            try:
                runpy.run_path(str(GENERATOR), run_name="__main__")
            except (RuntimeError, SystemExit, OSError, ValueError) as caught:
                error = caught
        return error, stdout.getvalue(), stderr.getvalue()

    def require_no_output(self):
        self.assertFalse(os.path.lexists(self.output))
        self.assertEqual(list(self.output_parent.iterdir()), [])

    def require_only_version(self):
        self.assertEqual(self.commands, [[str(self.binary), "version", "--json"]])
        self.require_no_output()

    def test_mismatch_never_initializes_or_exports(self):
        # Matching text elsewhere (branch/build) cannot rescue an explicit wrong
        # commit; the former human-output substring check could admit this.
        self.version.update(commit="f" * 40, branch=SOURCE[:9])
        error, stdout, _ = self.invoke()
        self.assertIsInstance(error, RuntimeError)
        self.assertIn("does not match", str(error))
        self.assertEqual(stdout, "")
        self.require_only_version()

    def test_missing_or_short_build_identity_refused(self):
        for info in [{"build": "dev"}, {"build": SOURCE[:8]}, {"commit": "", "build": SOURCE[:9]}, []]:
            with self.subTest(version=info):
                self.commands.clear()
                self.version = info
                error, _, _ = self.invoke()
                self.assertIsInstance(error, RuntimeError)
                self.require_only_version()

    def test_conflicting_build_refused_despite_matching_vcs_commit(self):
        self.version["build"] = "f" * 9
        error, _, _ = self.invoke()
        self.assertIsInstance(error, RuntimeError)
        self.require_only_version()

    def test_export_failure_publishes_nothing(self):
        self.export_failure = True
        error, stdout, _ = self.invoke()
        self.assertIsInstance(error, RuntimeError)
        self.assertIn("simulated export failure", str(error))
        self.assertEqual(stdout, "")
        self.assertTrue(any(command[1] == "init" for command in self.commands if command[0] != "git"))
        self.require_no_output()

    def test_success_pair_hashes_and_receipts(self):
        error, stdout, stderr = self.invoke("completeness")
        self.assertIsNone(error, (error, stderr))
        self.assertEqual(stdout, str(self.output / "ordinary.jsonl") + "\n")
        self.assertEqual(sorted(p.name for p in self.output.iterdir()), ["ordinary.jsonl", "producer.json"])
        self.assertEqual(list(self.output_parent.iterdir()), [self.output])
        fixture = (self.output / "ordinary.jsonl").read_bytes()
        receipt = json.loads((self.output / "producer.json").read_bytes())
        self.assertEqual(fixture, EXPORTED.encode("utf-8"))
        self.assertEqual(receipt["fixtureSHA256"], hashlib.sha256(fixture).hexdigest())
        self.assertEqual(receipt["binarySHA256"], hashlib.sha256(self.binary.read_bytes()).hexdigest())
        self.assertEqual(receipt["producerSource"], SOURCE)
        self.assertEqual(receipt["producerSourceTree"], TREE)
        self.assertEqual(receipt["profile"], "completeness")
        self.assertEqual(receipt["producerVersion"], f"bd version 1.3.0 ({SOURCE[:9]})\n")
        binary_commands = [command[1:] for command in self.commands if command[0] != "git"]
        self.assertEqual([entry["args"] for entry in receipt["receipts"]], binary_commands)
        self.assertEqual(binary_commands[:2], [["version", "--json"], ["version"]])
        self.assertEqual(binary_commands[-1], ["export", "--include-memories"])
        self.assertTrue(all(entry["code"] == 0 for entry in receipt["receipts"]))

    def test_old_canonical_build_identity_supported(self):
        self.version = {"version": "1.3.0", "build": SOURCE[:9]}
        error, _, stderr = self.invoke()
        self.assertIsNone(error, (error, stderr))
        self.assertTrue((self.output / "ordinary.jsonl").exists())
        self.assertTrue((self.output / "producer.json").exists())

    def test_existing_output_refused_before_any_invocation(self):
        self.output.mkdir()
        fixture = self.output / "ordinary.jsonl"
        receipt = self.output / "producer.json"
        fixture.write_bytes(b"old fixture")
        receipt.write_bytes(b"old receipt")
        error, stdout, stderr = self.invoke()
        self.assertIsInstance(error, SystemExit)
        self.assertEqual(error.code, 2)
        self.assertIn("fresh directory", stderr)
        self.assertEqual(stdout, "")
        self.assertEqual(self.commands, [])
        self.assertEqual(self.git_reads, [])
        self.assertEqual(fixture.read_bytes(), b"old fixture")
        self.assertEqual(receipt.read_bytes(), b"old receipt")

    def test_empty_directory_file_and_dangling_symlink_refused(self):
        for kind in ["empty-directory", "file", "dangling-symlink"]:
            with self.subTest(kind=kind):
                if kind == "empty-directory":
                    self.output.mkdir()
                elif kind == "file":
                    self.output.write_bytes(b"untouched")
                else:
                    self.output.symlink_to(self.root / "absent")
                error, _, _ = self.invoke()
                self.assertIsInstance(error, SystemExit)
                self.assertEqual(self.commands, [])
                self.assertEqual(self.git_reads, [])
                self.assertTrue(os.path.lexists(self.output))
                if kind == "empty-directory":
                    self.assertEqual(list(self.output.iterdir()), [])
                    self.output.rmdir()
                else:
                    if kind == "file":
                        self.assertEqual(self.output.read_bytes(), b"untouched")
                    else:
                        self.assertTrue(self.output.is_symlink())
                    self.output.unlink()

    def test_changed_binary_refused_without_publication(self):
        self.binary_change = True
        error, _, _ = self.invoke()
        self.assertIsInstance(error, RuntimeError)
        self.assertIn("binary changed", str(error))
        self.require_no_output()

    def test_changed_source_refused_without_publication(self):
        self.source_change = True
        error, _, _ = self.invoke()
        self.assertIsInstance(error, RuntimeError)
        self.assertIn("source changed", str(error))
        self.require_no_output()

    def test_receipt_write_failure_cleans_staged_fixture(self):
        original = Path.write_bytes
        def fail_receipt(path, data):
            if path.name == "producer.json":
                raise OSError("simulated receipt write failure")
            return original(path, data)
        with mock.patch.object(Path, "write_bytes", fail_receipt):
            error, _, _ = self.invoke()
        self.assertIsInstance(error, OSError)
        self.assertIn("receipt write failure", str(error))
        self.require_no_output()

    def test_rename_failure_cleans_staged_pair(self):
        with mock.patch.object(Path, "rename", side_effect=OSError("simulated rename failure")):
            error, _, _ = self.invoke()
        self.assertIsInstance(error, OSError)
        self.assertIn("rename failure", str(error))
        self.require_no_output()

    def test_output_appearing_during_generation_is_preserved(self):
        self.create_existing = True
        error, stdout, stderr = self.invoke()
        self.assertIsInstance(error, SystemExit)
        self.assertIn("appeared during generation", stderr)
        self.assertEqual(stdout, "")
        self.assertEqual(list(self.output_parent.iterdir()), [self.output])
        self.assertEqual(list(self.output.iterdir()), [self.output / "producer.json"])
        self.assertEqual((self.output / "producer.json").read_bytes(), b"preexisting receipt")


if __name__ == "__main__":
    unittest.main(verbosity=2)
