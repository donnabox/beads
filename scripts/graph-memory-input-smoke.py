#!/usr/bin/env python3
"""Prove installed graph remember file/stdin input preserves exact UTF-8 bytes.

Normal disposable initialization and CLI authoring only; both engines run
sequentially. Caller owns the ordinary server. No SQL, fixtures or mocks.
"""

import argparse
import hashlib
import importlib.util
import json
import os
from pathlib import Path
import signal
import time

HERE = Path(__file__).resolve().parent
SPEC = importlib.util.spec_from_file_location("graph_c0_capture", HERE / "graph-c0-smoke.py")
c0 = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(c0)
SCOPE = "https://example.invalid/disposable-memory-input/"
LIMIT = 1 << 20
LIMITATIONS = [
    "explicit identity/title remain required; no title derivation or alias claims",
    "new source flags are graph-only; legacy gate checked without opening legacy storage",
    "no concurrent writer, crash, corruption or HTTP qualification",
    "visible record/workspace equality does not prove unchanged server-internal bytes",
    "ordinary Dolt 2.1.8 database provisioning must remain serialized",
]


class RawStdin:
    """Use Capture's binary stdin file and receipt without Unicode transcoding."""
    def __init__(self, value):
        self.value = value

    def encode(self, encoding):
        c0.require(encoding == "utf-8", "unexpected Capture stdin encoding")
        return self.value


def exercise(capture):
    backend = "server" if capture.args.server_port else "embedded"
    inputs = capture.output / "inputs"
    inputs.mkdir(mode=0o700)
    manifest, expected = [], []

    def file(name, body):
        path = inputs / name
        path.write_bytes(body)
        manifest.append({"path": str(path), "bytes": len(body), "sha256": c0.sha256(path)})
        c0.write_json(capture.output / "inputs.json", manifest)
        return str(path)

    body = b"---\r\ntitle: not a directive\r\n---\r\n" + "  雪 😀 e\u0301\nend  ".encode()
    body_file = file("body.md", body)
    empty_file = file("empty.md", b"")
    bad_file = file("invalid-utf8.md", b"before\xffafter")
    large_file = file("oversize.md", b"x" * (LIMIT + 1))
    missing = str(inputs / "does-not-exist.md")

    def raw(label, argv, stdin=None):
        receipt, _, _ = capture.run(label, argv, input_text=None if stdin is None else RawStdin(stdin))
        directory = capture.output / capture.records[-1]["artifact"]
        return receipt, (directory / "stdout.log").read_bytes(), (directory / "stderr.log").read_bytes()

    def refusal(label, argv, code, stdin=None):
        receipt, out, err = raw(label, argv, stdin)
        c0.require(receipt["exit_code"] != 0 and out == b"", label + ": refusal emitted stdout or succeeded")
        problem = json.loads(err)
        c0.require(problem.get("code") == code and problem.get("retryable") is False,
                   label + ": wrong refusal")

    # No workspace exists: flags must be refused before legacy store opening.
    legacy_before = c0.tree_digest(capture.work)
    for label, flags in [("legacy-file", ["--body-file", missing]), ("legacy-stdin", ["--stdin"]),
                         ("legacy-false-stdin", ["--stdin=false"])]:
        refusal(label, ["remember", *flags, "--json"], "capability_unavailable")
    c0.require(c0.tree_digest(capture.work) == legacy_before, "legacy source gate created workspace data")
    capture.passed("new source flags refuse before opening legacy storage")

    init = ["init", "--graph-mode", "link", "--scope-url", SCOPE, "--prefix", "input",
            "--non-interactive", "--skip-hooks", "--skip-agents", "--json"]
    if capture.args.server_port:
        init += ["--server", "--external", "--server-host", "127.0.0.1", "--server-port",
                 str(capture.args.server_port), "--database", "input_" + capture.root.name.replace("-", "_"),
                 "--server-user", "root"]
    initialized = c0.envelope(capture.success("normal-init", init))
    c0.require(initialized.get("scope") == SCOPE and initialized.get("backend") == backend,
               "wrong initialized authority/backend")

    status = c0.envelope(capture.success("capabilities", ["status", "--graph", "--json"]))
    caps = status.get("capabilities", {})
    c0.require(caps.get("memoryBodyFileInput") is True and caps.get("memoryBodyStdinInput") is True,
               "status omitted file/stdin input capability")
    c0.require(caps.get("memory") is False and caps.get("historyExact") is False and
               status.get("limits", {}).get("memoryBodyInputBytes") == LIMIT,
               "status overstated Memory/History or omitted body input limit")

    def create(label, flags, wanted, stdin=None):
        receipt, out, err = raw(label, ["remember", "--id", "beads/" + label,
                                       "--title", "Input " + label, "--json", *flags], stdin)
        c0.require(receipt["exit_code"] == 0, label + ": remember failed: " + err.decode("utf-8", "replace"))
        record = c0.envelope(json.loads(out))
        c0.require(record["properties"]["body"].encode("utf-8") == wanted, label + ": body bytes changed")
        expected.append({"label": label, "record": record, "body_bytes": len(wanted),
                         "body_sha256": hashlib.sha256(wanted).hexdigest()})
        c0.write_json(capture.output / "expected-records.json", expected)
        return record

    create("file-body", ["--body-file", body_file], body)
    create("stdin-body", ["--stdin"], body, stdin=body)
    create("empty-file", ["--body-file", empty_file], b"")
    create("empty-stdin", ["--stdin"], b"", stdin=b"")
    create("positional-body", ["--", body.decode()], body)
    capture.passed("file/stdin/positional bodies preserve Unicode, CRLF, frontmatter and whitespace; empty sources succeed")

    cases = [
        ("missing-file", ["--body-file", missing], None, "invalid_properties"),
        ("invalid-file-utf8", ["--body-file", bad_file], None, "invalid_properties"),
        ("invalid-stdin-utf8", ["--stdin"], b"bad\xff", "invalid_properties"),
        ("oversize-file", ["--body-file", large_file], None, "capability_unavailable"),
        ("oversize-stdin", ["--stdin"], b"x" * (LIMIT + 1), "capability_unavailable"),
        ("missing-source", [], None, "invalid_properties"),
        ("two-positional", ["one", "two"], None, "invalid_properties"),
        ("file-and-positional", ["body", "--body-file", body_file], None, "invalid_properties"),
        ("stdin-and-positional", ["body", "--stdin"], body, "invalid_properties"),
        ("file-and-stdin", ["--body-file", body_file, "--stdin"], body, "invalid_properties"),
        ("false-stdin", ["--stdin=false"], None, "invalid_properties"),
        ("empty-file-name", ["--body-file", ""], None, "invalid_properties"),
    ]
    for label, flags, stdin, code in cases:
        refusal(label, ["remember", *flags, "--id", "beads/" + label, "--title", "Denied", "--json"], code, stdin)
        refusal(label + "-not-created", ["show", "beads/" + label, "--json"], "not_found")
    capture.passed("invalid UTF-8, size, absent/conflicting sources and missing files refuse without creating Resources")

    yaml = capture.work / ".beads" / "config.yaml"
    old_yaml = yaml.read_bytes() if yaml.exists() else None
    mayor, freeze = capture.work / "mayor", capture.work / "MIGRATION-FREEZE"
    c0.require(not mayor.exists() and not freeze.exists(), "unexpected town markers")
    for policy in ["flag", "environment", "configuration", "freeze"]:
        try:
            flags = ["--readonly"] if policy == "flag" else []
            if policy == "environment":
                capture.env["BD_READONLY"] = "true"
            elif policy == "configuration":
                yaml.write_text("readonly: true\n")
            elif policy == "freeze":
                mayor.mkdir()
                (mayor / "town.json").write_text("{}\n")
                freeze.write_text("input-smoke\t2026-09-26T00:00:00Z\tinput policy\n")
            before = c0.tree_digest(capture.work)
            refusal("policy-" + policy, ["remember", "--body-file", missing, "--id", "beads/denied-" + policy,
                                        "--title", "Denied", *flags, "--json"], "permission_denied")
            c0.require(c0.tree_digest(capture.work) == before, policy + ": refusal changed workspace")
        finally:
            capture.env.pop("BD_READONLY", None)
            if old_yaml is None:
                yaml.unlink(missing_ok=True)
            else:
                yaml.write_bytes(old_yaml)
            if policy == "freeze":
                freeze.unlink(missing_ok=True)
                (mayor / "town.json").unlink(missing_ok=True)
                mayor.rmdir()
        refusal("policy-not-created-" + policy, ["show", "beads/denied-" + policy, "--json"], "not_found")
    capture.passed("readonly and migration-freeze refusal precedes missing-file input reading")

    for item in expected:
        record, label = item["record"], item["label"]
        for historical in [False, True]:
            suffix = "-exact" if historical else "-current"
            version = ["--version", record["version"]] if historical else []
            actual = c0.envelope(capture.success(label + "-show" + suffix, [
                "show", record["id"], *version, "--readonly", "--json"]))
            c0.require(actual == record, label + ": saved/current record changed")
            receipt, stdout, stderr = raw(label + "-recall" + suffix, [
                "recall", record["id"], *version, "--readonly", "--quiet"])
            wanted = record["properties"]["body"].encode()
            c0.require(receipt["exit_code"] == 0 and stdout == wanted and stderr == b"",
                       label + ": recall failed exact byte comparison")
    capture.passed("fresh-process show and current/exact recall preserve every authored byte after all refusals")
    return {"saved_records": len(expected), "input_manifest_sha256": c0.sha256(capture.output / "inputs.json"),
            "expected_records_sha256": c0.sha256(capture.output / "expected-records.json"),
            "http_exercised": False}


def run_backend(args, backend, deadline):
    selected = argparse.Namespace(**vars(args))
    selected.output_dir = args.output_dir / backend
    selected.server_port = args.server_port if backend == "server" else None
    selected.server_root = args.server_root if backend == "server" else None
    selected.total_timeout = max(0.01, deadline - time.monotonic())
    capture = c0.Capture(selected)
    began = time.monotonic()
    failure, interrupted, result = None, False, {}
    previous_signals = {}

    def interrupted_run(signum, _frame):
        capture.stop()
        raise KeyboardInterrupt(signum)

    try:
        for signum in [signal.SIGINT, signal.SIGTERM]:
            previous_signals[signum] = signal.signal(signum, interrupted_run)
        result = exercise(capture)
        c0.require(c0.sha256(args.bd) == capture.binary_hash, "installed binary changed during capture")
    except BaseException as exc:
        interrupted = isinstance(exc, KeyboardInterrupt)
        failure = f"{type(exc).__name__}: {exc}"
    finally:
        capture.stop()
        for signum, handler in previous_signals.items():
            signal.signal(signum, handler)
        if capture.active:
            failure = failure or "owned child processes remain after capture"
        summary = {"passed": failure is None, "backend": backend, "failure": failure,
                   "interrupted": interrupted, "elapsed_seconds": time.monotonic() - began,
                   "checks": capture.passes, "commands": capture.records,
                   "active_child_count": len(capture.active), "private_root": str(capture.root),
                   "installed_binary_sha256": capture.binary_hash,
                   "harness_sha256": c0.sha256(Path(__file__)),
                   "capture_helper_sha256": c0.sha256(HERE / "graph-c0-smoke.py"),
                   "limitations": LIMITATIONS, **result}
        c0.write_json(capture.output / "summary.json", summary)
    print(f"{'PASS' if summary['passed'] else 'FAIL'} {backend}: {capture.output}", flush=True)
    return summary


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--bd", type=Path, required=True, help="absolute installed bd executable")
    parser.add_argument("--output-dir", type=Path, required=True, help="new capture directory")
    parser.add_argument("--backend", choices=["embedded", "server", "both"], default="both")
    parser.add_argument("--server-port", type=int, help="caller-owned ordinary loopback Dolt server")
    parser.add_argument("--server-root", type=Path, help="disposable caller-owned server root, provenance only")
    parser.add_argument("--command-timeout", type=float, default=120)
    parser.add_argument("--total-timeout", type=float, default=900, help="whole sequential capture budget")
    args = parser.parse_args()
    c0.require(args.bd.is_absolute() and args.bd.is_file() and os.access(args.bd, os.X_OK),
               "--bd must name an absolute executable installed binary")
    c0.require(args.output_dir.is_absolute(), "--output-dir must be absolute")
    c0.require(args.command_timeout > 0 and args.total_timeout > 0, "timeouts must be positive")
    if args.backend in {"server", "both"}:
        c0.require(args.server_port is not None and 0 < args.server_port < 65536,
                   "server/both require --server-port for an ordinary caller-owned server")
    if args.server_root is not None:
        c0.require(args.server_root.is_absolute() and args.server_root.is_dir(),
                   "--server-root must name an existing absolute disposable server directory")
    args.output_dir.mkdir(mode=0o700, parents=True, exist_ok=False)
    deadline = time.monotonic() + args.total_timeout
    modes = ["embedded", "server"] if args.backend == "both" else [args.backend]
    results = []
    failure = None
    try:
        for backend in modes:
            c0.require(time.monotonic() < deadline, "total capture deadline expired before next backend")
            result = run_backend(args, backend, deadline)
            results.append(result)
            if result["interrupted"]:
                break
    except BaseException as exc:
        failure = f"{type(exc).__name__}: {exc}"
    passed = failure is None and len(results) == len(modes) and all(item["passed"] for item in results)
    c0.write_json(args.output_dir / "summary.json", {
        "passed": passed, "failure": failure, "requested_backends": modes,
        "backends": [{key: value for key, value in result.items() if key != "commands"} for result in results],
        "harness_sha256": c0.sha256(Path(__file__)), "installed_binary_sha256": c0.sha256(args.bd),
        "active_child_count": sum(result["active_child_count"] for result in results),
        "command_count": sum(len(result["commands"]) for result in results),
        "check_count": sum(len(result["checks"]) for result in results),
        "limitations": LIMITATIONS,
    })
    return 0 if passed else 1


if __name__ == "__main__":
    raise SystemExit(main())
