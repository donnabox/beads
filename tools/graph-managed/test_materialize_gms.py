"""Synthetic controls by default; the real archive gate requires explicit inputs."""

import argparse
import copy
import contextlib
import hashlib
import importlib.util
import io
import json
import os
from pathlib import Path
import stat
import struct
import sys
import tempfile
import unittest
from unittest import mock
import zipfile

SPEC = importlib.util.spec_from_file_location("materialize_gms", Path(__file__).with_name("materialize-gms.py"))
m = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(m)


def row(path, data):
    return {"path": path, "bytes": len(data), "sha256": hashlib.sha256(data).hexdigest()}


def fixture_zip(pairs, modes=None):
    stream = io.BytesIO()
    with zipfile.ZipFile(stream, "w") as archive:
        for i, (name, data) in enumerate(pairs):
            info = zipfile.ZipInfo(name)
            info.compress_type = zipfile.ZIP_DEFLATED
            # Nonzero DOS bit avoids ZipFile's default POSIX permissions. The
            # pinned Go archive has no high-word POSIX mode bits either.
            info.external_attr = 1 if modes is None else modes[i]
            archive.writestr(info, data)
    stream.seek(0)
    return stream


class MaterializerTests(unittest.TestCase):
    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory(prefix="gms-synthetic-")
        self.addCleanup(self.temporary.cleanup)
        self.parent = Path(self.temporary.name).resolve()
        self.deadline = m.Deadline()

    def refuse(self, function, *args):
        with self.assertRaises((m.Refusal, OSError, ValueError, zipfile.BadZipFile)):
            function(*args)

    def patch_case(self):
        name, old, new = "server/context.go", b"one\ntwo\nthree\n", b"one\nchanged\nthree\n"
        patch = (b"--- a/server/context.go\n+++ b/server/context.go\n"
                 b"@@ -1,3 +1,3 @@\n one\n-two\n+changed\n three\n")
        return patch, {name: old}, {name: row(name, new)}, new

    def test_budget_exact_then_over_and_noninteger(self):
        budget = m.Budget(3)
        budget.charge(3)
        self.refuse(budget.charge, 1)
        self.assertEqual(budget.used, 3)
        for value in (-1, True, 1.0):
            self.refuse(m.Budget(3).charge, value)

    def test_path_and_directory_aliases(self):
        m.validate_paths(["a/b", "a/c", "é/file"])
        cases = [
            (["../x"], "invalid path component"), (["/x"], "invalid path component"),
            (["a//b"], "invalid path component"), (["a/./b"], "invalid path component"),
            (["a\\b"], "invalid relative path character"), (["C:x"], "invalid relative path character"),
            (["a\x00b"], "invalid relative path character"), (["a\nb"], "invalid relative path character"),
            (["e\u0301/file"], "noncanonical Unicode path"), (["a", "a/b"], "file/directory collision"),
            (["a/b", "A/c"], "path alias collision"), (["x", "x"], "duplicate file path"),
            (["ß/a", "ss/b"], "path alias collision"), (["a" * 257], "invalid relative path size"),
            (["/".join(["a"] * 17)], "invalid path component"),
        ]
        for paths, message in cases:
            with self.subTest(paths=paths), self.assertRaisesRegex(m.Refusal, message):
                m.validate_paths(paths)
        m.validate_paths(["a" * 256, "/".join(["b"] * 16)])

    def test_manifest_closed_schema_and_nested_fields(self):
        path = Path(m.__file__).parent / "patches/gms/manifest.json"
        data = path.read_bytes()
        self.assertEqual(m.digest(data), m.MANIFEST_SHA256)
        manifest = json.loads(data, object_pairs_hook=m.unique_object)
        original, final = m.validate_manifest(manifest)
        self.assertEqual((len(original), len(final)), (1688, 1690))
        for edit in (lambda x: x.update(schema_version=True), lambda x: x.update(extra=1),
                     lambda x: x["consumer"].update(extra=1),
                     lambda x: x["result_files"][0].update(bytes=True)):
            changed = copy.deepcopy(manifest)
            edit(changed)
            self.refuse(m.validate_manifest, changed)
        with self.assertRaises(m.Refusal):
            json.loads('{"a":1,"a":2}', object_pairs_hook=m.unique_object)

    def test_inventory_row_and_file_size_neighbors(self):
        exact = row("a", b"abc")
        with mock.patch.object(m, "FILE_LIMIT", 3):
            self.assertEqual(m.rows_by_path([exact], 1), {"a": exact})
            self.refuse(m.rows_by_path, [row("a", b"abcd")], 1)
        self.refuse(m.rows_by_path, [exact, exact], 2)
        self.refuse(m.rows_by_path, [dict(exact, extra=1)], 1)
        self.refuse(m.rows_by_path, [dict(exact, bytes=True)], 1)
        self.refuse(m.rows_by_path, [dict(exact, sha256="0" * 63)], 1)

    def test_pinned_input_hash_size_symlink_and_change(self):
        path = self.parent / "input"
        path.write_bytes(b"abc")
        with m.pinned_input(path, 3, m.digest(b"abc"), self.deadline, 3) as source:
            self.assertEqual(source.read(), b"abc")
        for label in ("archive", "consumer", "patch", "manifest"):
            with self.subTest(label=label), self.assertRaises(m.Refusal):
                with m.pinned_input(path, 3, m.digest(label.encode()), self.deadline):
                    self.fail("mismatching input admitted")
        with self.assertRaises(m.Refusal):
            with m.pinned_input(path, 2, m.digest(b"abc"), self.deadline):
                self.fail("over-limit input admitted")
        link = self.parent / "link"
        link.symlink_to(path)
        with self.assertRaises(OSError):
            with m.pinned_input(link, 3, m.digest(b"abc"), self.deadline):
                self.fail("symlink admitted")
        with self.assertRaisesRegex(m.Refusal, "input identity changed"):
            with m.pinned_input(path, 3, m.digest(b"abc"), self.deadline):
                path.write_bytes(b"abcd")  # Size changes even on coarse-timestamp filesystems.

    def test_inventory_positive_and_rejections(self):
        expected = {"a": row("a", b"abc")}
        with zipfile.ZipFile(fixture_zip([("p/a", b"abc")])) as archive:
            admitted = m.inventory(archive, "p/", expected, self.deadline)
            self.assertEqual(b"".join(m.member_chunks(archive, admitted["a"], expected["a"], m.Budget(3), self.deadline)), b"abc")
            self.refuse(lambda: b"".join(m.member_chunks(archive, admitted["a"], expected["a"], m.Budget(2), self.deadline)))
            wrong = dict(expected["a"], sha256=m.digest(b"bad"))
            self.refuse(lambda: b"".join(m.member_chunks(archive, admitted["a"], wrong, m.Budget(3), self.deadline)))
        # Some ZIP variants reject at the expected inventory before alias checks;
        # direct validate_paths controls above discriminate those path rules.
        cases = [([("p/../a", b"abc")], None), ([("q/a", b"abc")], None),
                 ([("p/a", b"ab")], None), ([("p/a/", b"abc")], None),
                 ([("p/a", b"abc")], [(stat.S_IFLNK | 0o777) << 16]),
                 ([("p/a", b"abc")], [0o100755 << 16]),
                 ([("p/a", b"abc"), ("p/a", b"abc")], None),
                 ([("p/a", b"abc"), ("p/A", b"abc")], None), ([], None)]
        for pairs, modes in cases:
            with self.subTest(pairs=pairs), zipfile.ZipFile(fixture_zip(pairs, modes)) as archive:
                self.refuse(m.inventory, archive, "p/", expected, self.deadline)
        with zipfile.ZipFile(fixture_zip([("p/a", b"abc")])) as archive:
            with mock.patch.object(m, "ENTRY_LIMIT", 0):
                self.refuse(m.inventory, archive, "p/", expected, self.deadline)
            with mock.patch.object(m, "TOTAL_LIMIT", 2):
                self.refuse(m.inventory, archive, "p/", expected, self.deadline)

    def test_crc_and_truncation(self):
        data = bytearray(fixture_zip([("p/a", b"abc")]).getvalue())
        index = data.index(b"PK\x01\x02")
        # Corrupt central-directory CRC while retaining valid compressed bytes.
        struct.pack_into("<I", data, index + 16, 0)
        with zipfile.ZipFile(io.BytesIO(data)) as archive:
            self.refuse(lambda: b"".join(m.member_chunks(archive, archive.infolist()[0], row("a", b"abc"), m.Budget(3), self.deadline)))
        self.refuse(zipfile.ZipFile, io.BytesIO(bytes(data[:20])))

    def test_patch_exact_and_new_file(self):
        patch, before, after, new = self.patch_case()
        self.assertEqual(m.apply_patch(patch, before, after, self.deadline), {"server/context.go": new})
        name = "server/close_existing_session_test.go"
        patch = ("--- /dev/null\n+++ b/" + name + "\n@@ -0,0 +1,1 @@\n+new\n").encode()
        self.assertEqual(m.apply_patch(patch, {name: None}, {name: row(name, b"new\n")}, self.deadline), {name: b"new\n"})

    def test_patch_rejects_context_positions_counts_and_extra_syntax(self):
        patch, before, after, _ = self.patch_case()
        variants = [patch.replace(b"-two", b"-wrong"), patch.replace(b"-1,3", b"-2,3"),
                    patch.replace(b"+1,3", b"+2,3"), patch.replace(b"-1,3", b"-1,4"),
                    patch.replace(b"+1,3", b"+1,2"), patch.replace(b"-1,3", b"-01,3"),
                    patch.replace(b"-1,3", b"-2147483648,3"), patch.replace(b"@@\n", b"@@ suffix\n"),
                    patch.replace(b"--- a/", b"--- /dev/null\n--- a/"),
                    patch.replace(b"context.go", b"../escape"), patch[:patch.index(b"@@")],
                    b"diff --git a/x b/x\n" + patch, b"old mode 100644\n" + patch,
                    patch + b"\\ No newline at end of file\n", patch + b"extra\n", patch + patch,
                    patch.replace(b"\n", b"\r\n"), patch[:-1], patch + b"\x00\n"]
        for value in variants:
            with self.subTest(value=value):
                self.refuse(m.apply_patch, value, before, after, self.deadline)
        for value in (patch.replace(b"context.go", b"../escape"), patch.replace(b"--- a/", b"--- /dev/null\n--- a/")):
            with self.assertRaisesRegex(m.Refusal, "patch file header mismatch"):
                m.apply_patch(value, before, after, self.deadline)
        with self.assertRaisesRegex(m.Refusal, "patch old context mismatch"):
            m.apply_patch(patch.replace(b"-two", b"-wrong"), before, after, self.deadline)
        # A repeated hunk reaches the ascending/nonoverlap boundary directly.
        repeated = patch + b"@@ -1,3 +4,3 @@\n one\n-two\n+changed\n three\n"
        with self.assertRaisesRegex(m.Refusal, "hunk offset/overlap"):
            m.apply_patch(repeated, before, after, self.deadline)
        wrong_after = {p: dict(r, sha256="0" * 64) for p, r in after.items()}
        self.refuse(m.apply_patch, patch, before, wrong_after, self.deadline)
        unknown = "server/unknown.go"
        created = ("--- /dev/null\n+++ b/" + unknown + "\n@@ -0,0 +1,1 @@\n+x\n").encode()
        self.refuse(m.apply_patch, created, {unknown: None}, {unknown: row(unknown, b"x\n")}, self.deadline)

    def install_fixture(self, name="output", changed=None, final=None, input_pin=None):
        original = {"dir/a": row("dir/a", b"abc")}
        final = original if final is None else final
        with zipfile.ZipFile(fixture_zip([("p/dir/a", b"abc")])) as archive:
            entries = m.inventory(archive, "p/", original, self.deadline)
            return m.install(archive, entries, original, final, changed or {}, str(self.parent), name,
                             {"fixture": True}, m.Budget(m.TOTAL_LIMIT), self.deadline, input_pin)

    def test_install_write_once_modes_receipt_and_existing_refusal(self):
        original_write = m.write_file
        calls = []

        def recording(*args, **kwargs):
            calls.append(args[1])
            return original_write(*args, **kwargs)

        with mock.patch.object(m, "write_file", side_effect=recording):
            receipt = self.install_fixture()
        self.assertTrue(receipt["success"])
        self.assertEqual(calls, ["dir/a", "receipt.pending"])
        source = self.parent / "output/source/dir/a"
        self.assertEqual(source.read_bytes(), b"abc")
        self.assertEqual(stat.S_IMODE(source.stat().st_mode), 0o644)
        saved = (self.parent / "output/receipt.json").read_bytes()
        self.refuse(self.install_fixture)
        self.assertEqual((self.parent / "output/receipt.json").read_bytes(), saved)
        self.assertEqual(source.read_bytes(), b"abc")

    def test_output_symlink_parent_and_existing_leaf(self):
        link = self.parent / "link"
        link.symlink_to(self.parent, target_is_directory=True)
        self.refuse(m.open_parent, str(link), self.deadline)
        (self.parent / "output").symlink_to(self.parent, target_is_directory=True)
        # Exclusive mkdir refuses any existing destination, including a link;
        # this is not evidence of reaching the later descriptor no-follow open.
        with self.assertRaises(FileExistsError):
            self.install_fixture()
        fd = m.open_parent(str(self.parent), self.deadline)
        try:
            (self.parent / "leaf").write_bytes(b"old")
            self.refuse(m.write_file, fd, "leaf", (b"new",), m.Budget(3), self.deadline)
            self.assertEqual((self.parent / "leaf").read_bytes(), b"old")
        finally:
            os.close(fd)

    def test_partial_output_never_success_on_write_hash_or_identity_failure(self):
        def broken_chunks(*args):
            yield b"a"
            raise OSError("injected write-input failure")

        with mock.patch.object(m, "member_chunks", side_effect=broken_chunks):
            self.refuse(self.install_fixture, "write-failure")
        bad_final = {"dir/a": row("dir/a", b"bad")}
        self.refuse(self.install_fixture, "hash-failure", {}, bad_final)
        path = self.parent / "archive-input"
        path.write_bytes(b"old")
        with path.open("rb") as stream:
            old = m.identity(os.fstat(stream.fileno()))
            path.write_bytes(b"changed")
            self.refuse(self.install_fixture, "identity-failure", None, None, (stream, old))
        for name in ("write-failure", "hash-failure", "identity-failure"):
            self.assertTrue((self.parent / name / "source").is_dir())
            receipt = json.loads((self.parent / name / "receipt.failure.json").read_bytes())
            self.assertIs(receipt["success"], False)
            self.assertFalse((self.parent / name / "receipt.json").exists())

    def test_pending_partial_write_and_full_close_failure_never_publish(self):
        original_write, original_fdopen = m.write_file, os.fdopen
        for kind in ("partial", "close"):
            class FaultyFile:
                def __init__(self, real):
                    self.real = real

                def __enter__(self):
                    return self

                def __exit__(self, *ignored):
                    self.real.close()
                    if kind == "close":
                        raise OSError("injected pending close failure")

                def fileno(self):
                    return self.real.fileno()

                def write(self, data):
                    if kind == "partial":
                        self.real.write(data[:len(data) // 2])
                        raise OSError("injected partial pending write")
                    return self.real.write(data)

            def faulting_write(*args, **kwargs):
                if args[1] != "receipt.pending":
                    return original_write(*args, **kwargs)
                with mock.patch.object(m.os, "fdopen", side_effect=lambda *a, **k: FaultyFile(original_fdopen(*a, **k))):
                    return original_write(*args, **kwargs)

            with self.subTest(kind=kind), mock.patch.object(m, "write_file", side_effect=faulting_write):
                self.refuse(self.install_fixture, kind)
            directory = self.parent / kind
            self.assertFalse((directory / "receipt.json").exists())
            self.assertFalse(json.loads((directory / "receipt.failure.json").read_bytes())["success"])
            if kind == "close":
                # Fully parseable pending success is deliberately nonauthoritative.
                self.assertTrue(json.loads((directory / "receipt.pending").read_bytes())["success"])

    def test_exclusive_link_failure_and_postpublication_cleanup(self):
        original_link, original_close = os.link, os.close
        published = False

        def link_existing(src, dst, **kwargs):
            fd = os.open(dst, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600, dir_fd=kwargs["dst_dir_fd"])
            with os.fdopen(fd, "wb") as stream:
                stream.write(b"existing sentinel")
            return original_link(src, dst, **kwargs)

        def link_success(*args, **kwargs):
            nonlocal published
            result = original_link(*args, **kwargs)
            published = True
            return result

        def close_fault(fd):
            original_close(fd)
            if published:
                raise OSError("injected postpublication directory close")

        for name, behavior in (("link-failure", link_existing), ("cleanup", link_success)):
            with mock.patch.object(m.os, "link", side_effect=behavior) as link:
                with mock.patch.object(m.os, "supports_dir_fd", os.supports_dir_fd | {link}), mock.patch.object(m.os, "supports_follow_symlinks", os.supports_follow_symlinks | {link}):
                    if name == "link-failure":
                        self.refuse(self.install_fixture, name)
                    else:
                        with mock.patch.object(m.os, "close", side_effect=close_fault):
                            result = self.install_fixture(name)
                        self.assertTrue(result["success"])
                        self.assertIn("injected postpublication", result["postpublication_cleanup"][0])
                        receipt = json.loads((self.parent / name / "receipt.json").read_bytes())
                        self.assertTrue(receipt["success"])
        self.assertEqual((self.parent / "link-failure/receipt.json").read_bytes(), b"existing sentinel")
        self.assertFalse(json.loads((self.parent / "link-failure/receipt.failure.json").read_bytes())["success"])

    def test_archive_close_after_commit_is_diagnostic(self):
        bundle = Path(m.__file__).parent / "patches/gms"
        manifest = json.loads((bundle / "manifest.json").read_bytes())
        original = {r["path"]: r for r in manifest["original_files"]}
        archive_path = self.parent / "fixture.zip"
        archive_path.write_bytes(fixture_zip([]).getvalue())

        @contextlib.contextmanager
        def admitted(path, *args, **kwargs):
            if str(path) == "archive":
                with archive_path.open("rb") as source:
                    yield source
                raise OSError("injected archive close after publication")
            elif str(path) == "consumer":
                yield io.BytesIO(b"")
            else:
                with Path(path).open("rb") as source:
                    yield source

        success = {"success": True, "container": "fixture-owned"}
        with mock.patch.object(m, "pinned_input", side_effect=admitted), mock.patch.object(m, "inventory", return_value={p: None for p in original}), mock.patch.object(m, "member_chunks", side_effect=lambda *args: iter([b"fixture"])), mock.patch.object(m, "apply_patch", return_value={}), mock.patch.object(m, "install", return_value=success):
            result = m.materialize("archive", "consumer", str(self.parent), "result")
        self.assertTrue(result["success"])
        self.assertIn("archive close after publication", result["postpublication_cleanup"][0])

    def test_primary_write_failure_survives_owned_directory_close(self):
        class ReceiptAbort(BaseException):
            pass

        original_close, original_write = os.close, m.write_file
        for fail_receipt in (False, True):
            primary_seen = False
            source_closes = []
            name = "combined-receipt-failure" if fail_receipt else "combined-failure"
            container = self.parent / name

            def failed_chunks(*args):
                nonlocal primary_seen
                yield b"a"
                primary_seen = True
                raise OSError("primary-member-sentinel")

            def close_fault(fd):
                st = os.fstat(fd)
                source = container / "source"
                target = primary_seen and source.exists() and st.st_ino == source.stat().st_ino
                original_close(fd)
                if target:
                    source_closes.append(fd)
                    raise OSError("directory-cleanup-sentinel")

            def receipt_fault(*args, **kwargs):
                if fail_receipt and args[1] == "receipt.failure.json":
                    raise ReceiptAbort("failure-receipt-sentinel")
                return original_write(*args, **kwargs)

            with self.subTest(fail_receipt=fail_receipt), mock.patch.object(m, "member_chunks", side_effect=failed_chunks), mock.patch.object(m.os, "close", side_effect=close_fault), mock.patch.object(m, "write_file", side_effect=receipt_fault):
                with self.assertRaises(m.Refusal) as caught:
                    self.install_fixture(name)
            self.assertEqual(len(source_closes), 1)
            self.assertEqual(caught.exception.owned_path, str(container))
            message = m.format_refusal(caught.exception)
            self.assertIn("primary-member-sentinel", message)
            self.assertIn("directory-cleanup-sentinel", message)
            self.assertIn(str(container), message)
            self.assertTrue((container / "source").is_dir())
            self.assertFalse((container / "receipt.json").exists())
            if fail_receipt:
                self.assertIn("failure-receipt-sentinel", message)
                self.assertFalse((container / "receipt.failure.json").exists())
            else:
                self.assertFalse(json.loads((container / "receipt.failure.json").read_bytes())["success"])

    def test_primary_install_refusal_survives_archive_and_input_unwind(self):
        bundle = Path(m.__file__).parent / "patches/gms"
        manifest = json.loads((bundle / "manifest.json").read_bytes())
        original = {r["path"]: r for r in manifest["original_files"]}
        archive_path = self.parent / "unwind.zip"
        archive_path.write_bytes(fixture_zip([]).getvalue())
        real_zip = zipfile.ZipFile
        for kind in ("archive", "input"):
            # This is an exception-propagation seam, not an actual allocation of
            # the modeled owned output or a second archive-admission test.
            owned_path = str(self.parent / ("modeled-" + kind))
            primary = m.Refusal("primary-install-sentinel", owned_path=owned_path)

            @contextlib.contextmanager
            def admitted(path, *args, **kwargs):
                if str(path) == "archive":
                    with archive_path.open("rb") as source:
                        try:
                            yield source
                        finally:
                            if kind == "input":
                                raise OSError("input-cleanup-sentinel")
                elif str(path) == "consumer":
                    yield io.BytesIO(b"")
                else:
                    with Path(path).open("rb") as source:
                        yield source

            @contextlib.contextmanager
            def archive_context(stream):
                with real_zip(stream) as archive:
                    try:
                        yield archive
                    finally:
                        if kind == "archive":
                            raise OSError("archive-cleanup-sentinel")

            with self.subTest(kind=kind), mock.patch.object(m, "pinned_input", side_effect=admitted), mock.patch.object(m.zipfile, "ZipFile", side_effect=archive_context), mock.patch.object(m, "inventory", return_value={p: None for p in original}), mock.patch.object(m, "member_chunks", side_effect=lambda *args: iter([b"fixture"])), mock.patch.object(m, "apply_patch", return_value={}), mock.patch.object(m, "install", side_effect=primary):
                with self.assertRaises(m.Refusal) as caught:
                    m.materialize("archive", "consumer", str(self.parent), "result")
            message = m.format_refusal(caught.exception)
            self.assertEqual(caught.exception.owned_path, owned_path)
            self.assertIn("primary-install-sentinel", message)
            self.assertIn(kind + "-cleanup-sentinel", message)
            self.assertIn(owned_path, message)
            self.assertNotIn("source materialized", message)

    def test_long_owned_location_and_unowned_existing_refusal(self):
        parent = "/" + "/".join(["x" * 255] * 15 + ["y" * 239])
        self.assertLessEqual(len(os.fsencode(parent)), 4096)
        owned_path = parent + "/fresh-name"
        error = m.Refusal("primary-" + "e" * 2000, owned_path=owned_path)
        stderr = io.StringIO()
        argv = ["materialize-gms.py", "--archive", "fixture", "--consumer-go-mod", "fixture", "--output-parent", parent, "--name", "fresh-name"]
        with mock.patch.object(sys, "argv", argv), mock.patch.object(m, "materialize", side_effect=error), contextlib.redirect_stderr(stderr):
            self.assertEqual(m.main(), 1)
        self.assertIn(owned_path, stderr.getvalue())
        self.assertLess(len(stderr.getvalue()), len(owned_path) + 650)
        self.assertNotIn("e" * 513, stderr.getvalue())

        existing = self.parent / "existing"
        existing.mkdir()
        (existing / "sentinel").write_bytes(b"unowned")
        original_close = os.close
        parent_inode = self.parent.stat().st_ino

        def close_fault(fd):
            target = os.fstat(fd).st_ino == parent_inode
            original_close(fd)
            if target:
                raise OSError("unowned-parent-cleanup-sentinel")

        with mock.patch.object(m.os, "close", side_effect=close_fault):
            with self.assertRaises(m.Refusal) as caught:
                self.install_fixture("existing")
        self.assertIsNone(caught.exception.owned_path)
        self.assertIn("File exists", str(caught.exception))
        self.assertNotIn("incomplete output retained", m.format_refusal(caught.exception))
        self.assertEqual((existing / "sentinel").read_bytes(), b"unowned")
        self.assertFalse((existing / "source").exists())

    def test_private_parent_modes_uid_and_canonical_paths(self):
        parent = self.parent / "parent-admission"
        parent.mkdir(mode=0o700)
        fd = m.open_parent(str(parent), self.deadline)
        os.close(fd)
        try:
            for mode in (0o750, 0o705, 0o770, 0o701):
                parent.chmod(mode)
                with self.subTest(mode=oct(mode)), self.assertRaisesRegex(m.Refusal, "output parent must be private and owned"):
                    m.open_parent(str(parent), self.deadline)
        finally:
            parent.chmod(0o700)
        uid = os.geteuid()
        with mock.patch.object(m.os, "geteuid", return_value=uid + 1):
            with self.assertRaisesRegex(m.Refusal, "output parent must be private and owned"):
                m.open_parent(str(parent), self.deadline)
        for path, message in (("relative", "absolute bounded output parent required"),
                              (str(parent) + "/", "canonical output parent required"),
                              (str(parent) + "/.", "canonical output parent required"),
                              (str(parent) + "/..", "canonical output parent required"),
                              (str(parent) + "//child", "canonical output parent required"),
                              ("/" + "x" * 4096, "absolute bounded output parent required")):
            with self.subTest(path=path), self.assertRaisesRegex(m.Refusal, message):
                m.open_parent(path, self.deadline)
        self.assertEqual(list(parent.iterdir()), [])

    def test_main_success_reporting_failure_and_unsupported_platform(self):
        # The N1 long-location control already exercises main's exit1 branch.
        argv = ["materialize-gms.py", "--archive", "fixture", "--consumer-go-mod", "fixture",
                "--output-parent", str(self.parent), "--name", "not-created"]
        result = {"success": True, "container": str(self.parent / "modeled-committed")}
        stdout, stderr = io.StringIO(), io.StringIO()
        with mock.patch.object(sys, "argv", argv), mock.patch.object(m, "materialize", return_value=result), contextlib.redirect_stdout(stdout), contextlib.redirect_stderr(stderr):
            self.assertEqual(m.main(), 0)
        self.assertEqual(json.loads(stdout.getvalue()), result)
        self.assertEqual(stderr.getvalue(), "")

        class BrokenOutput(io.StringIO):
            def write(self, data):
                raise OSError("stdout-reporting-sentinel")

        stderr = io.StringIO()
        with mock.patch.object(sys, "argv", argv), mock.patch.object(m, "materialize", return_value=result), contextlib.redirect_stdout(BrokenOutput()), contextlib.redirect_stderr(stderr):
            self.assertEqual(m.main(), 2)
        self.assertIn("source materialized at " + result["container"], stderr.getvalue())
        self.assertIn("stdout-reporting-sentinel", stderr.getvalue())
        self.assertNotIn("materialization refused", stderr.getvalue())
        for capability in ("supports_dir_fd", "supports_follow_symlinks"):
            reduced = getattr(os, capability) - {os.link}
            with self.subTest(capability=capability), mock.patch.object(m.os, capability, reduced), mock.patch.object(m, "pinned_input", side_effect=AssertionError("input opened before capability admission")) as input_open:
                with self.assertRaisesRegex(m.Refusal, "exclusive receipt publication APIs required"):
                    m.materialize("fixture", "fixture", str(self.parent), "not-created")
                input_open.assert_not_called()
        self.assertEqual(list(self.parent.iterdir()), [])

    def test_real_gate_rejects_cleanup_diagnostics(self):
        argv = ["test_materialize_gms.py", "--real-archive", "fixture", "--consumer-go-mod", "fixture", "--output-parent", str(self.parent)]
        # This modeled result exercises the validation gate's extra clean-teardown
        # requirement; committed source remains success in the materializer.
        result = {"success": True, "files": 1690, "postpublication_cleanup": ["cleanup-sentinel"]}
        stdout = io.StringIO()
        with mock.patch.object(sys, "argv", argv), mock.patch.object(m, "materialize", return_value=result), contextlib.redirect_stdout(stdout):
            with self.assertRaisesRegex(AssertionError, "postpublication cleanup diagnostics"):
                real_gate()
        self.assertNotIn('"PASS"', stdout.getvalue())
        self.assertEqual(list(self.parent.iterdir()), [])

    def test_patch_multiple_sections_and_ascending_hunks(self):
        before = {"server/context.go": b"a\nb\nc\nd\ne\nf\ng\nh\n",
                  "server/handler.go": b"u\nv\nw\nx\ny\n"}
        outputs = {"server/context.go": b"a\nB\nc\nd\nE\nextra\nf\ng\nh\n",
                   "server/handler.go": b"u\nV\nw\nX\ny\n"}
        patch = (b"--- a/server/context.go\n+++ b/server/context.go\n"
                 b"@@ -2,1 +2,1 @@\n-b\n+B\n"
                 b"@@ -5,1 +5,2 @@\n-e\n+E\n+extra\n"
                 b"--- a/server/handler.go\n+++ b/server/handler.go\n"
                 b"@@ -2,1 +2,1 @@\n-v\n+V\n"
                 b"@@ -4,1 +4,1 @@\n-x\n+X\n")
        expected = {p: row(p, data) for p, data in outputs.items()}
        self.assertEqual(m.apply_patch(patch, before, expected, self.deadline), outputs)

    def test_expired_deadline_preserves_primary_owned_path_without_restart(self):
        def expiring_chunks(*args):
            yield b"a"
            self.deadline.ends = 0
            self.deadline.check()

        with mock.patch.object(m, "member_chunks", side_effect=expiring_chunks), mock.patch.object(m, "Deadline", side_effect=AssertionError("deadline restarted")) as new_deadline:
            with self.assertRaises(m.Refusal) as caught:
                self.install_fixture("expired")
            new_deadline.assert_not_called()
        container = self.parent / "expired"
        self.assertEqual(caught.exception.owned_path, str(container))
        self.assertEqual(str(caught.exception), "cooperative materialization deadline")
        self.assertIn(str(container), m.format_refusal(caught.exception))
        self.assertIn("failure receipt:", m.format_refusal(caught.exception))
        self.assertFalse((container / "receipt.json").exists())
        failure = container / "receipt.failure.json"
        if failure.exists():
            # Exclusive open can precede the expired chunk-write checkpoint;
            # an empty incomplete diagnostic file is never an authoritative receipt.
            self.assertEqual(failure.read_bytes(), b"")

    def test_output_limit_and_final_symlink_refuse(self):
        with mock.patch.object(m, "TOTAL_LIMIT", 3):
            self.refuse(self.install_fixture, "limit-failure", {"dir/a": b"abcd"}, {"dir/a": row("dir/a", b"abcd")})
        receipt = json.loads((self.parent / "limit-failure/receipt.failure.json").read_bytes())
        self.assertFalse(receipt["success"])
        self.install_fixture("linked")
        source = self.parent / "linked/source"
        (source / "dir/a").unlink()
        (source / "dir/a").symlink_to(self.parent / "unrelated")
        fd = os.open(source, m.directory_flags())
        try:
            self.refuse(m.verify_tree, fd, {"dir/a": row("dir/a", b"abc")}, self.deadline)
        finally:
            os.close(fd)


def real_gate():
    parser = argparse.ArgumentParser(description="Explicit pinned real-archive round trip; never downloads or executes GMS")
    parser.add_argument("--real-archive", required=True)
    parser.add_argument("--consumer-go-mod", required=True)
    parser.add_argument("--output-parent", required=True)
    args = parser.parse_args()
    # Cleanup belongs to this test's newly created temporary child only.
    fd = m.open_parent(args.output_parent, m.Deadline())
    os.close(fd)
    with tempfile.TemporaryDirectory(prefix="gms-real-", dir=args.output_parent) as temporary:
        parent = str(Path(temporary).resolve())
        receipt = m.materialize(args.real_archive, args.consumer_go_mod, parent, "result")
        if not receipt["success"] or receipt["files"] != 1690:
            raise AssertionError("real archive did not produce complete final map")
        if receipt.get("postpublication_cleanup"):
            raise AssertionError("source materialized with postpublication cleanup diagnostics; clean gate refused")
        saved = Path(parent, "result/receipt.json").read_bytes()
        try:
            m.materialize(args.real_archive, args.consumer_go_mod, parent, "result")
        except FileExistsError:
            pass
        else:
            raise AssertionError("existing real destination was accepted")
        if Path(parent, "result/receipt.json").read_bytes() != saved:
            raise AssertionError("existing receipt changed")
        print(json.dumps({"real_archive_gate": "PASS", "receipt": receipt}, sort_keys=True))


if __name__ == "__main__":
    if "--real-archive" in sys.argv:
        real_gate()
    else:
        unittest.main()
