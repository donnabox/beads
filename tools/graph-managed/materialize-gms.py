#!/usr/bin/env python3
"""Materialize the one reviewed GMS source bundle; never build or execute it."""

import argparse
import contextlib
import hashlib
import json
import os
from pathlib import Path
import re
import stat
import sys
import time
import unicodedata
import zipfile

MANIFEST_SHA256 = "a3d7e07a020431fa4a15d03a10bb1bcbd07263d6c2a0656929c5b6c58cd884fc"
FIXED_SECTIONS = {
    'upstream': '59302f30aa932fd325332eee7f5daf29aef1c9aace63a5f6fe1da669633aaf3f',
    'consumer': '02e0a67e15faef78d06250f7594812a2e5bc8d39bfdc3b189e55f53d01dcb4cc',
    'patch': '6172e434a5bf97af82d4aa641bf3e45bca7f4d9757aed6491e3001fc17b9b232',
    'lineage': 'df455f9224f24a01f34f9f4866ab860d4ca4acd2e917918ab08580ad1aeaf8a2',
    'licenses': 'd0d80bc79b051fb74494b5a9c78bf872d106de7418e4722230e4fa8458e42045',
    'qualification': '8f2cd0a0ae70468815eb3839f60a61fc055cf5c0ac760a0f0e46041dc1ce8eb5',
}
MIB = 1024 * 1024
MANIFEST_LIMIT, PATCH_LIMIT, ARCHIVE_LIMIT = 2 * MIB, MIB, 64 * MIB
FILE_LIMIT, TOTAL_LIMIT, ENTRY_LIMIT = 8 * MIB, 128 * MIB, 2048
CHUNK = 64 * 1024
TARGETS = (
    "server/close_existing_session_test.go", "server/context.go",
    "server/handler.go", "server/terminal_connection_integration_test.go",
)
NEW_FILES = {TARGETS[0], TARGETS[3]}
TOP_KEYS = {"schema_version", "component_id", "upstream", "consumer", "patch",
            "original_files", "result_files", "lineage", "licenses", "qualification"}
NUMBER = rb"(0|[1-9][0-9]{0,9})"
HUNK = re.compile(rb"@@ -" + NUMBER + rb"," + NUMBER + rb" \+" + NUMBER + rb"," + NUMBER + rb" @@\n")


class Refusal(Exception):
    """A failed admission, with an owned location separate from bounded detail."""

    def __init__(self, detail, *, owned_path=None, cleanup=()):
        super().__init__(detail)
        self.owned_path = owned_path
        self.cleanup = list(cleanup)


def format_refusal(error):
    message = "materialization refused: " + str(error)[:512]
    if isinstance(error, Refusal):
        if error.owned_path is not None:
            message += "\nincomplete output retained at " + error.owned_path
        for detail in error.cleanup[:5]:
            message += "\ncleanup diagnostic: " + detail[:512]
    return message


def require(condition, message):
    if not condition:
        raise Refusal(message)


def digest(data):
    return hashlib.sha256(data).hexdigest()


def canonical(value):
    return (json.dumps(value, sort_keys=True, separators=(",", ":")) + "\n").encode()


class Deadline:
    def __init__(self):
        self.ends = time.monotonic() + 120

    def check(self):
        require(time.monotonic() < self.ends, "cooperative materialization deadline")


class Budget:
    def __init__(self, limit):
        self.limit, self.used = limit, 0

    def charge(self, size):
        require(type(size) is int and 0 <= size <= self.limit - self.used,
                "byte or occurrence budget exceeded")
        self.used += size


def unique_object(pairs):
    result = {}
    for key, value in pairs:
        require(key not in result, "duplicate JSON key")
        result[key] = value
    return result


def path_parts(path):
    require(type(path) is str and 0 < len(path.encode("utf-8")) <= 256, "invalid relative path size")
    require(not any(c in path for c in "\\\x00:\r\n"), "invalid relative path character")
    parts = path.split("/")
    require(len(parts) <= 16 and all(x not in ("", ".", "..") for x in parts), "invalid path component")
    require(unicodedata.normalize("NFC", path) == path, "noncanonical Unicode path")
    return parts


def validate_paths(paths):
    aliases, files, directories = {}, set(), set()
    for path in paths:
        parts = path_parts(path)
        require(path not in files, "duplicate file path")
        files.add(path)
        for i in range(1, len(parts) + 1):
            prefix = "/".join(parts[:i])
            folded = prefix.casefold()
            require(folded not in aliases or aliases[folded] == prefix, "path alias collision")
            aliases[folded] = prefix
            if i < len(parts):
                directories.add(prefix)
    require(not files.intersection(directories), "file/directory collision")


def rows_by_path(rows, count):
    require(type(rows) is list and len(rows) == count <= ENTRY_LIMIT, "wrong inventory count")
    total, result = Budget(TOTAL_LIMIT), {}
    for row in rows:
        require(type(row) is dict and set(row) == {"path", "bytes", "sha256"}, "invalid inventory row")
        size = row["bytes"]
        require(type(size) is int and 0 <= size <= FILE_LIMIT, "invalid member size")
        total.charge(size)
        require(type(row["sha256"]) is str and re.fullmatch(r"[0-9a-f]{64}", row["sha256"]), "invalid member hash")
        path_parts(row["path"])
        require(row["path"] not in result, "duplicate inventory path")
        result[row["path"]] = row
    require(list(result) == sorted(result), "inventory must be sorted")
    validate_paths(result)
    return result


def validate_manifest(m):
    require(type(m) is dict and set(m) == TOP_KEYS, "unknown or missing manifest field")
    require(type(m["schema_version"]) is int and m["schema_version"] == 1, "unsupported schema version")
    require(m["component_id"] == "gms-terminal-session", "wrong component")
    # This bundle has no configurable metadata schema: exact canonical section
    # hashes bind every nested key/type/value to the reviewed manifest.
    for key, expected in FIXED_SECTIONS.items():
        require(digest(canonical(m[key])) == expected, "changed fixed manifest section: " + key)
    original = rows_by_path(m["original_files"], 1688)
    final = rows_by_path(m["result_files"], 1690)
    require(set(final) == set(original) | NEW_FILES, "wrong final path set")
    changes = {p for p in final if p not in original or final[p] != original[p]}
    require(changes == set(TARGETS), "patch must change exactly four paths")
    return original, final


def identity(st):
    return st.st_dev, st.st_ino, st.st_size, st.st_mtime_ns, st.st_ctime_ns


@contextlib.contextmanager
def pinned_input(path, cap, expected, deadline, expected_size=None, recheck_on_exit=True):
    fd = os.open(path, os.O_RDONLY | os.O_NOFOLLOW | os.O_NONBLOCK | os.O_CLOEXEC)
    with os.fdopen(fd, "rb", buffering=0) as stream:
        before = os.fstat(stream.fileno())
        require(stat.S_ISREG(before.st_mode) and 0 <= before.st_size <= cap, "input is not a bounded regular file")
        if expected_size is not None:
            require(before.st_size == expected_size, "input byte size mismatch")
        h, remaining = hashlib.sha256(), before.st_size
        while remaining:
            deadline.check()
            data = stream.read(min(CHUNK, remaining))
            require(data, "truncated input")
            h.update(data)
            remaining -= len(data)
        require(not stream.read(1), "input grew")
        require(identity(os.fstat(stream.fileno())) == identity(before), "input changed while hashing")
        require(h.hexdigest() == expected, "input SHA256 mismatch")
        stream.seek(0)
        yield stream
        if recheck_on_exit:
            require(identity(os.fstat(stream.fileno())) == identity(before), "input identity changed")


def inventory(archive, prefix, expected, deadline):
    entries, occurrences, total = {}, Budget(ENTRY_LIMIT), Budget(TOTAL_LIMIT)
    for info in archive.infolist():
        deadline.check()
        occurrences.charge(1)
        require(info.orig_filename == info.filename and info.filename.startswith(prefix), "wrong ZIP member prefix/name")
        name = info.filename[len(prefix):]
        path_parts(name)
        require(name not in entries, "duplicate ZIP member")
        require(not info.is_dir() and info.external_attr >> 16 == 0, "ZIP links, modes or directories are unsupported")
        require(not info.flag_bits & 1 and info.compress_type == zipfile.ZIP_DEFLATED, "unsupported ZIP encoding")
        require(0 <= info.file_size <= FILE_LIMIT, "ZIP member size exceeds limit")
        total.charge(info.file_size)
        require(name in expected and info.file_size == expected[name]["bytes"], "unexpected ZIP member/size")
        entries[name] = info
    require(set(entries) == set(expected), "incomplete ZIP inventory")
    validate_paths(entries)
    return entries


def member_chunks(archive, info, row, budget, deadline):
    h, remaining = hashlib.sha256(), row["bytes"]
    with archive.open(info) as stream:
        while remaining:
            deadline.check()
            size = min(CHUNK, remaining)
            budget.charge(size)  # Reserve before the bounded read/allocation.
            data = stream.read(size)
            require(len(data) == size, "truncated ZIP member")
            remaining -= size
            h.update(data)
            yield data
        # Constant-size decoder probe; never retained or written as payload.
        require(not stream.read(1), "ZIP member exceeded declared size")
    require(h.hexdigest() == row["sha256"], "ZIP member hash mismatch")


def apply_patch(raw, before, expected, deadline):
    require(len(raw) <= PATCH_LIMIT and b"\r" not in raw and b"\x00" not in raw, "invalid patch encoding/size")
    raw.decode("utf-8", "strict")
    require(raw.endswith(b"\n"), "patch must end with LF")
    lines, pos, result = raw.splitlines(keepends=True), 0, {}
    for name in sorted(expected):
        deadline.check()
        old = before[name]
        header = b"--- /dev/null\n" if old is None else ("--- a/" + name + "\n").encode()
        require(old is not None or name in NEW_FILES, "unexpected new file")
        require(pos + 2 <= len(lines) and lines[pos] == header and
                lines[pos + 1] == ("+++ b/" + name + "\n").encode(), "patch file header mismatch")
        pos += 2
        source, emitted, cursor, hunks = (old or b"").splitlines(keepends=True), [], 0, 0
        output_budget = Budget(FILE_LIMIT)

        def emit(data):
            output_budget.charge(len(data))
            emitted.append(data)

        while pos < len(lines) and lines[pos].startswith(b"@@"):
            deadline.check()
            match = HUNK.fullmatch(lines[pos])
            require(match is not None, "invalid hunk header")
            start, removed, new_start, added = (int(x) for x in match.groups())
            require(max(start, removed, new_start, added) <= 2147483647, "hunk integer overflow")
            if old is None:
                require(hunks == 0 and start == removed == 0 and new_start == 1, "invalid creation hunk")
            else:
                require(start > 0 and removed > 0, "unsupported empty existing hunk")
            offset = start - 1 if removed else start
            require(cursor <= offset <= len(source) and offset + removed <= len(source), "hunk offset/overlap")
            for line in source[cursor:offset]:
                emit(line)
            require(new_start - 1 == len(emitted), "new hunk offset mismatch")
            cursor, used_old, used_new = offset, 0, 0
            pos += 1
            while used_old < removed or used_new < added:
                deadline.check()
                require(pos < len(lines), "missing hunk body")
                line = lines[pos]
                require(line.endswith(b"\n") and line[:1] in (b" ", b"-", b"+"), "invalid hunk record")
                tag, payload = line[:1], line[1:]
                if tag != b"+":
                    require(used_old < removed and cursor < len(source) and source[cursor] == payload, "patch old context mismatch")
                    cursor += 1
                    used_old += 1
                if tag != b"-":
                    require(used_new < added, "patch new count mismatch")
                    emit(payload)
                    used_new += 1
                pos += 1
            hunks += 1
        require(hunks > 0, "missing file hunk")
        for line in source[cursor:]:
            emit(line)
        data = b"".join(emitted)
        require(len(data) == expected[name]["bytes"] and digest(data) == expected[name]["sha256"], "patched output mismatch")
        result[name] = data
    require(pos == len(lines), "trailing patch data or extra section")
    return result


def directory_flags():
    require(os.name == "posix" and all(hasattr(os, x) for x in ("O_DIRECTORY", "O_NOFOLLOW", "O_CLOEXEC")), "POSIX descriptor APIs required")
    require(os.link in os.supports_dir_fd and os.link in os.supports_follow_symlinks, "exclusive receipt publication APIs required")
    return os.O_RDONLY | os.O_DIRECTORY | os.O_NOFOLLOW | os.O_CLOEXEC


def open_parent(path, deadline):
    require(type(path) is str and path.startswith("/") and len(os.fsencode(path)) <= 4096, "absolute bounded output parent required")
    parts = path.split("/")[1:]
    require(parts and all(x not in ("", ".", "..") for x in parts), "canonical output parent required")
    fd = os.open("/", directory_flags())
    try:
        for part in parts:
            deadline.check()
            next_fd = os.open(part, directory_flags(), dir_fd=fd)
            os.close(fd)
            fd = next_fd
        st = os.fstat(fd)
        require(st.st_uid == os.geteuid() and stat.S_IMODE(st.st_mode) & 0o077 == 0, "output parent must be private and owned")
        return fd
    except BaseException:
        os.close(fd)
        raise


@contextlib.contextmanager
def file_parent(root_fd, name, deadline):
    parts = path_parts(name)
    fd = os.dup(root_fd)
    try:
        for part in parts[:-1]:
            deadline.check()
            try:
                os.mkdir(part, 0o700, dir_fd=fd)
            except FileExistsError:
                pass
            next_fd = os.open(part, directory_flags(), dir_fd=fd)
            os.close(fd)
            fd = next_fd
            st = os.fstat(fd)
            require(st.st_uid == os.geteuid() and stat.S_IMODE(st.st_mode) == 0o700, "unexpected output directory")
        yield fd, parts[-1]
    finally:
        os.close(fd)


def write_file(root_fd, name, chunks, output_budget, deadline, mode=0o644):
    with file_parent(root_fd, name, deadline) as (parent, leaf):
        fd = os.open(leaf, os.O_WRONLY | os.O_CREAT | os.O_EXCL | os.O_NOFOLLOW | os.O_CLOEXEC, mode, dir_fd=parent)
        with os.fdopen(fd, "wb", buffering=0) as stream:
            os.fchmod(stream.fileno(), mode)
            for data in chunks:
                deadline.check()
                output_budget.charge(len(data))
                view = memoryview(data)
                while view:
                    deadline.check()
                    written = stream.write(view)
                    require(written is not None and written > 0, "incomplete file write")
                    view = view[written:]


def verify_tree(root_fd, final, deadline):
    observed, expected_dirs = {}, set()
    for name in final:
        parts = name.split("/")
        expected_dirs.update("/".join(parts[:i]) for i in range(1, len(parts)))
    seen_dirs = set()

    def visit(fd, prefix):
        for name in sorted(os.listdir(fd)):
            deadline.check()
            path = prefix + name
            st = os.stat(name, dir_fd=fd, follow_symlinks=False)
            if stat.S_ISDIR(st.st_mode):
                require(path in expected_dirs and st.st_uid == os.geteuid() and stat.S_IMODE(st.st_mode) == 0o700, "unexpected final directory")
                seen_dirs.add(path)
                child = os.open(name, directory_flags(), dir_fd=fd)
                try:
                    visit(child, path + "/")
                finally:
                    os.close(child)
            else:
                require(path in final and stat.S_ISREG(st.st_mode) and st.st_nlink == 1 and
                        st.st_uid == os.geteuid() and stat.S_IMODE(st.st_mode) == 0o644 and
                        st.st_size == final[path]["bytes"], "unexpected final file")
                child = os.open(name, os.O_RDONLY | os.O_NOFOLLOW | os.O_CLOEXEC, dir_fd=fd)
                with os.fdopen(child, "rb", buffering=0) as stream:
                    require(identity(os.fstat(stream.fileno())) == identity(st), "final file replaced")
                    h, remaining = hashlib.sha256(), st.st_size
                    while remaining:
                        deadline.check()
                        data = stream.read(min(CHUNK, remaining))
                        require(data, "final file truncated")
                        remaining -= len(data)
                        h.update(data)
                    require(not stream.read(1) and identity(os.fstat(stream.fileno())) == identity(st), "final file changed")
                require(h.hexdigest() == final[path]["sha256"], "final file hash mismatch")
                observed[path] = final[path]
    visit(root_fd, "")
    require(set(observed) == set(final) and seen_dirs == expected_dirs, "incomplete final tree")


def install(archive, entries, original, final, changed, parent_path, name, receipt, read_budget, deadline, input_pin=None):
    require(type(name) is str and re.fullmatch(r"[A-Za-z0-9][A-Za-z0-9._-]{0,63}", name), "invalid output name")
    validate_paths(final)
    container_path = parent_path + "/" + name
    committed, success, created = False, None, False
    cleanup_errors = []

    def close_owned(fd):
        try:
            os.close(fd)
        except BaseException as error:
            # All owned closes are attempted once. Their diagnostics must not
            # replace a primary prepublication refusal during stack unwinding.
            cleanup_errors.append(type(error).__name__ + ": " + str(error)[:512])

    try:
        with contextlib.ExitStack() as owned:
            parent = open_parent(parent_path, deadline)
            owned.callback(close_owned, parent)
            parent_stat = os.fstat(parent)
            os.mkdir(name, 0o700, dir_fd=parent)  # Existing destinations always refuse.
            created = True
            try:
                container = os.open(name, directory_flags(), dir_fd=parent)
            except OSError as error:
                raise Refusal(str(error)[:512], owned_path=container_path) from error
            owned.callback(close_owned, container)
            try:
                os.fchmod(container, 0o700)
                os.mkdir("source", 0o700, dir_fd=container)
                source = os.open("source", directory_flags(), dir_fd=container)
                owned.callback(close_owned, source)
                os.fchmod(source, 0o700)
                output_budget = Budget(TOTAL_LIMIT)
                for path in sorted(final):
                    deadline.check()
                    chunks = (changed[path],) if path in changed else member_chunks(archive, entries[path], original[path], read_budget, deadline)
                    write_file(source, path, chunks, output_budget, deadline)
                verify_tree(source, final, deadline)
                again = open_parent(parent_path, deadline)
                try:
                    require((os.fstat(again).st_dev, os.fstat(again).st_ino) == (parent_stat.st_dev, parent_stat.st_ino), "output parent replaced")
                finally:
                    os.close(again)
                for owner, leaf, fd in [(parent, name, container), (container, "source", source)]:
                    st, actual = os.stat(leaf, dir_fd=owner, follow_symlinks=False), os.fstat(fd)
                    require(stat.S_ISDIR(st.st_mode) and (st.st_dev, st.st_ino) == (actual.st_dev, actual.st_ino), "output directory replaced")
                if input_pin is not None:
                    stream, expected_identity = input_pin
                    require(identity(os.fstat(stream.fileno())) == expected_identity, "archive identity changed")
                success = dict(receipt, success=True, container=container_path,
                               output_map_sha256=digest(canonical([final[p] for p in sorted(final)])),
                               files=len(final), output_bytes=output_budget.used,
                               container_identity=[os.fstat(container).st_dev, os.fstat(container).st_ino])
                payload = canonical(success)
                write_file(container, "receipt.pending", (payload,), Budget(MIB), deadline, 0o600)
                # The nonauthoritative pending file must be fully written AND
                # closed before it is eligible for atomic exclusive publication.
                pending = os.open("receipt.pending", os.O_RDONLY | os.O_NOFOLLOW | os.O_CLOEXEC, dir_fd=container)
                with os.fdopen(pending, "rb", buffering=0) as stream:
                    st = os.fstat(stream.fileno())
                    require(stat.S_ISREG(st.st_mode) and st.st_nlink == 1 and st.st_uid == os.geteuid() and
                            stat.S_IMODE(st.st_mode) == 0o600 and st.st_size == len(payload), "invalid pending receipt")
                    require(stream.read(len(payload) + 1) == payload, "pending receipt content mismatch")
                    require(identity(os.fstat(stream.fileno())) == identity(st), "pending receipt changed")
                require(identity(os.stat("receipt.pending", dir_fd=container, follow_symlinks=False)) == identity(st), "pending receipt replaced")
                deadline.check()
                os.link("receipt.pending", "receipt.json", src_dir_fd=container, dst_dir_fd=container, follow_symlinks=False)
                committed = True  # The exclusive link is the success point.
            except BaseException as error:
                if committed:
                    raise
                # Retain the owned partial output. A receipt failure cannot erase the
                # original error or authorize overwrite/removal of a partial receipt.
                try:
                    failure = dict(receipt, success=False, container=container_path,
                                   failure=type(error).__name__, detail=str(error)[:512])
                    write_file(container, "receipt.failure.json", (canonical(failure),), Budget(MIB), deadline, 0o600)
                except BaseException as receipt_error:
                    cleanup_errors.append("failure receipt: " + type(receipt_error).__name__ + ": " + str(receipt_error)[:512])
                raise Refusal(str(error)[:512], owned_path=container_path) from error

    except BaseException as error:
        # Committed source stays successful even if later cleanup is interrupted;
        # KeyboardInterrupt/SystemExit are diagnostics here, not a rollback.
        if committed:
            return dict(success, postpublication_cleanup=cleanup_errors + [type(error).__name__ + ": " + str(error)[:512]])
        if created or cleanup_errors:
            failure = error if isinstance(error, Refusal) else Refusal(str(error)[:512])
            if created:
                failure.owned_path = container_path
            failure.cleanup.extend(cleanup_errors)
            if failure is error:
                raise
            raise failure from error
        raise
    if cleanup_errors:
        return dict(success, postpublication_cleanup=cleanup_errors)
    return success


def materialize(archive_path, consumer_path, parent_path, name):
    deadline = Deadline()
    directory_flags()
    bundle = Path(__file__).resolve().parent / "patches" / "gms"
    with pinned_input(bundle / "manifest.json", MANIFEST_LIMIT, MANIFEST_SHA256, deadline) as stream:
        raw_manifest = stream.read(MANIFEST_LIMIT + 1)
    manifest = json.loads(raw_manifest, object_pairs_hook=unique_object)
    require(canonical(manifest) == raw_manifest, "manifest serialization changed")
    original, final = validate_manifest(manifest)
    with pinned_input(bundle / "terminal-session.patch", PATCH_LIMIT, manifest["patch"]["sha256"], deadline, manifest["patch"]["bytes"]) as stream:
        patch = stream.read(PATCH_LIMIT + 1)
    consumer = manifest["consumer"]
    with pinned_input(consumer_path, CHUNK, consumer["go_mod_sha256"], deadline, consumer["go_mod_bytes"]):
        pass
    arc = manifest["upstream"]["archive"]
    committed_result, primary_failure = None, None
    try:
        with pinned_input(archive_path, ARCHIVE_LIMIT, arc["sha256"], deadline, arc["bytes"], recheck_on_exit=False) as stream:
            # Recheck this live descriptor inside install before publishing success;
            # an exit-time check would be too late to invalidate that receipt.
            input_pin = stream, identity(os.fstat(stream.fileno()))
            with zipfile.ZipFile(stream) as archive:
                entries = inventory(archive, arc["prefix"], original, deadline)
                read_budget = Budget(TOTAL_LIMIT)
                before = {p: b"".join(member_chunks(archive, entries[p], original[p], read_budget, deadline)) if p in original else None for p in TARGETS}
                changed = apply_patch(patch, before, {p: final[p] for p in TARGETS}, deadline)
                receipt = {"schema_version": 1, "component_id": manifest["component_id"],
                           "limits": {"manifest": MANIFEST_LIMIT, "patch": PATCH_LIMIT, "archive": ARCHIVE_LIMIT,
                                      "file": FILE_LIMIT, "total": TOTAL_LIMIT, "entries": ENTRY_LIMIT,
                                      "path_bytes": 256, "path_components": 16, "cooperative_seconds": 120},
                           "materializer_sha256": digest(Path(__file__).read_bytes()),
                           "manifest_sha256": MANIFEST_SHA256, "patch_sha256": manifest["patch"]["sha256"],
                           "archive_sha256": arc["sha256"], "consumer_go_mod_sha256": consumer["go_mod_sha256"],
                           "source_commit": manifest["lineage"]["final_local_commit"],
                           "qualification": "verified source materialization only; no build or execution"}
                try:
                    committed_result = install(archive, entries, original, final, changed, parent_path, name, receipt, read_budget, deadline, input_pin)
                except BaseException as error:
                    primary_failure = error
                    raise
    except BaseException as error:
        if committed_result is None:
            if primary_failure is not None and error is not primary_failure:
                failure = primary_failure if isinstance(primary_failure, Refusal) else Refusal(str(primary_failure)[:512])
                failure.cleanup.append(type(error).__name__ + ": " + str(error)[:512])
                raise failure from error
            raise
        committed_result.setdefault("postpublication_cleanup", []).append(type(error).__name__ + ": " + str(error)[:512])
    return committed_result


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--archive", required=True)
    parser.add_argument("--consumer-go-mod", required=True)
    parser.add_argument("--output-parent", required=True)
    parser.add_argument("--name", required=True)
    args = parser.parse_args()
    try:
        result = materialize(args.archive, args.consumer_go_mod, args.output_parent, args.name)
    except (Refusal, OSError, ValueError, zipfile.BadZipFile, RuntimeError, RecursionError) as error:
        print(format_refusal(error), file=sys.stderr)
        return 1
    try:
        print(json.dumps(result, sort_keys=True), flush=True)
    except (OSError, ValueError) as error:
        # receipt.json remains authoritative after publication. Reporting or
        # descriptor cleanup failure is not a reversal of materialization.
        print("source materialized at " + result["container"] + "; result reporting failed: " + str(error)[:512], file=sys.stderr)
        return 2
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
