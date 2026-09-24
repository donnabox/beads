#!/usr/bin/env python3
"""Exercise an installed bd through disposable, fresh-process graph commands.

No SQL setup, model, or mock is used. A supplied ordinary Dolt server is owned
by the caller and is never started or stopped here. All subprocess output and
workspaces are retained, including on failure. This is a POSIX capture harness.
"""

import argparse
import concurrent.futures
import hashlib
import json
import os
from pathlib import Path
import selectors
import signal
import subprocess
import tempfile
import threading
import time


SCOPE_INPUT = "https://example.invalid/disposable-c0"
SCOPE = SCOPE_INPUT + "/"
OUTPUT_CAP = 2 * 1024 * 1024


def require(condition, message):
    if not condition:
        raise RuntimeError(message)


def write_json(path, value):
    path.write_text(json.dumps(value, ensure_ascii=False, indent=2) + "\n")


def sha256(path):
    digest = hashlib.sha256()
    with path.open("rb") as source:
        for block in iter(lambda: source.read(1024 * 1024), b""):
            digest.update(block)
    return digest.hexdigest()


def tree_digest(root):
    """Content, names, modes and symlink targets; ignore access-time noise."""
    entries = []
    for path in sorted(root.rglob("*")):
        stat = path.lstat()
        relative = str(path.relative_to(root))
        if path.is_symlink():
            value = [relative, stat.st_mode, "link", os.readlink(path)]
        elif path.is_file():
            value = [relative, stat.st_mode, "file", sha256(path)]
        elif path.is_dir():
            value = [relative, stat.st_mode, "directory"]
        else:
            raise RuntimeError(f"unexpected workspace entry: {path}")
        entries.append(value)
    raw = json.dumps(entries, separators=(",", ":")).encode()
    return {"sha256": hashlib.sha256(raw).hexdigest(), "entries": entries}


class Capture:
    def __init__(self, args):
        self.args = args
        self.output = args.output_dir
        self.output.mkdir(mode=0o700, parents=True, exist_ok=False)
        base = Path("/private/tmp") if Path("/private/tmp").is_dir() else Path("/tmp")
        for ancestor in [base, *base.parents]:
            require(not (ancestor / ".beads").exists(), f"ancestor .beads: {ancestor}")
        self.root = Path(tempfile.mkdtemp(prefix="bd-c0-", dir=base)).resolve()
        self.work = self.root / "workspace"
        self.work.mkdir(mode=0o700)
        self.home = self.root / "home"
        self.home.mkdir(mode=0o700)
        self.tmp = self.root / "tmp"
        self.tmp.mkdir(mode=0o700)
        gitconfig = self.home / "gitconfig"
        gitconfig.write_text("[user]\n\tname = C0 Smoke\n\temail = c0@example.invalid\n")
        self.env = {
            "PATH": os.environ.get("PATH", "/usr/bin:/bin"),
            "HOME": str(self.home), "TMPDIR": str(self.tmp),
            "TMP": str(self.tmp), "TEMP": str(self.tmp),
            "XDG_CONFIG_HOME": str(self.home / "config"),
            "XDG_CACHE_HOME": str(self.home / "cache"),
            "XDG_DATA_HOME": str(self.home / "data"),
            "GIT_CONFIG_GLOBAL": str(gitconfig), "GIT_CONFIG_NOSYSTEM": "1",
            "GIT_TERMINAL_PROMPT": "0", "BD_NON_INTERACTIVE": "1",
            "BD_DISABLE_METRICS": "1", "BD_DISABLE_EVENT_FLUSH": "1",
            "DOLT_METRICS_DISABLED": "1", "DOLT_DISABLE_EVENT_FLUSH": "1",
            "DO_NOT_TRACK": "1", "BEADS_DOLT_AUTO_START": "0",
            "NO_COLOR": "1", "TZ": "UTC", "LANG": "en_US.UTF-8",
        }
        if os.environ.get("DEVELOPER_DIR"):
            self.env["DEVELOPER_DIR"] = os.environ["DEVELOPER_DIR"]
        self.deadline = time.monotonic() + args.total_timeout
        self.active = set()
        self.lock = threading.Lock()
        self.stopped = threading.Event()
        self.number = 0
        self.records = []
        self.passes = []
        self.qualification_gaps = []
        self.binary_hash = sha256(args.bd)
        write_json(self.output / "provenance.json", {
            "binary": str(args.bd), "binary_sha256": self.binary_hash,
            "harness_sha256": sha256(Path(__file__).resolve()),
            "private_root": str(self.root), "environment": self.env,
            "server_port": args.server_port,
            "server_root": str(args.server_root) if args.server_root else None,
            "server_ownership": "caller-owned; not started or stopped by harness",
            "command_timeout_seconds": args.command_timeout,
            "total_timeout_seconds": args.total_timeout,
        })

    def stop(self):
        self.stopped.set()
        with self.lock:
            children = list(self.active)
        for child in children:
            self.kill_group(child)

    @staticmethod
    def kill_group(child):
        try:
            os.killpg(child.pid, signal.SIGKILL)
        except ProcessLookupError:
            pass

    def run(self, label, arguments):
        require(not self.stopped.is_set(), "capture was cancelled")
        require(time.monotonic() < self.deadline, "total capture deadline expired")
        require(sha256(self.args.bd) == self.binary_hash, "installed binary changed")
        with self.lock:
            self.number += 1
            stem = f"{self.number:02d}-{label}"
        directory = self.output / stem
        directory.mkdir()
        started = time.monotonic()
        argv = [str(self.args.bd), *arguments]
        receipt = {"argv": argv, "cwd": str(self.work), "started_unix": time.time()}
        print(f"START {stem}: {arguments[0]}", flush=True)
        child = None
        failure = None
        try:
            with (directory / "stdout.log").open("wb") as out, (directory / "stderr.log").open("wb") as err:
                child = subprocess.Popen(argv, cwd=self.work, env=self.env,
                                         stdin=subprocess.DEVNULL, stdout=subprocess.PIPE,
                                         stderr=subprocess.PIPE, start_new_session=True)
                receipt["pid"] = child.pid
                with self.lock:
                    self.active.add(child)
                selector = selectors.DefaultSelector()
                selector.register(child.stdout, selectors.EVENT_READ, out)
                selector.register(child.stderr, selectors.EVENT_READ, err)
                sizes = {out: 0, err: 0}
                try:
                    end = min(self.deadline, started + self.args.command_timeout)
                    while selector.get_map():
                        if self.stopped.is_set() or time.monotonic() >= end:
                            raise RuntimeError("command cancelled or deadline exceeded")
                        for key, _ in selector.select(timeout=0.2):
                            data = os.read(key.fileobj.fileno(), 65536)
                            if not data:
                                selector.unregister(key.fileobj)
                                key.fileobj.close()
                                continue
                            available = OUTPUT_CAP - sizes[key.data]
                            key.data.write(data[:available])
                            sizes[key.data] += len(data)
                            require(sizes[key.data] <= OUTPUT_CAP, "output cap exceeded")
                    child.wait(timeout=max(0.01, end - time.monotonic()))
                finally:
                    selector.close()
        except BaseException as exc:
            failure = f"{type(exc).__name__}: {exc}"
            raise
        finally:
            if child is not None:
                # A command owns its process group; kill leftovers even after
                # its leader exits, and reap that leader before returning.
                self.kill_group(child)
                child.wait(timeout=10)
                for pipe in [child.stdout, child.stderr]:
                    if pipe:
                        pipe.close()
                with self.lock:
                    self.active.discard(child)
                receipt["exit_code"] = child.returncode
            receipt.update(elapsed_seconds=time.monotonic() - started, failure=failure)
            for stream in ["stdout", "stderr"]:
                path = directory / f"{stream}.log"
                if path.exists():
                    receipt[f"{stream}_sha256"] = sha256(path)
            write_json(directory / "receipt.json", receipt)
            with self.lock:
                self.records.append({"artifact": stem, **receipt})
            print(f"END {stem}: exit={receipt.get('exit_code')} failure={failure}", flush=True)
        return receipt, (directory / "stdout.log").read_text(), (directory / "stderr.log").read_text()

    def success(self, label, arguments):
        receipt, out, err = self.run(label, arguments)
        require(receipt["exit_code"] == 0, f"{label} failed: {err}")
        return json.loads(out)

    def passed(self, name):
        self.passes.append(name)
        print(f"PASS {name}", flush=True)


def envelope(value):
    require(isinstance(value, dict), "expected JSON preview envelope")
    require(value.get("schemaVersion") == 1, "unexpected envelope schemaVersion")
    require(value.get("preview") is True, "missing preview disclosure")
    require(isinstance(value.get("result"), dict), "missing result object")
    return value["result"]


def record(value, path, body, title):
    value = envelope(value)
    require(value.get("id") == SCOPE + path, "canonical full ID mismatch")
    for field in ["type", "revision", "version"]:
        require(isinstance(value.get(field), str) and value[field], f"missing {field}")
    require(value["type"] == SCOPE + "types/preview-memory-v1", "expected installed non-Issue Memory descriptor")
    require(value.get("properties") == {"body": body, "title": title}, "Memory properties changed")
    require(value.get("owned") == [], "expected explicit empty owned-Link state")
    return value


def refusal(result, allowed, label):
    receipt, out, err = result
    require(receipt["exit_code"] != 0, f"{label} unexpectedly succeeded")
    require(not out.strip(), f"{label} emitted success output on failure")
    try:
        value = json.loads(err)
    except json.JSONDecodeError as exc:
        raise RuntimeError(f"{label} lacked typed JSON stderr: {err}") from exc
    require(isinstance(value, dict), f"{label} malformed refusal")
    code = value.get("code")
    require(code != "outcome_unknown", f"{label}: outcome_unknown is a qualification gap, never a safe refusal")
    require(code in allowed, f"{label} unexpected refusal code: {code}")
    require(value.get("retryable") is False, f"{label} unexpectedly permits automatic replay")
    require(isinstance(value.get("message"), str) and value["message"], f"{label} missing diagnostic")
    return value


def remember(path, body, title):
    return ["remember", body, "--id", path, "--title", title, "--json"]


def exercise(capture):
    args = ["init", "--graph-mode", "link", "--scope-url", SCOPE_INPUT, "--prefix", "demo",
            "--non-interactive", "--skip-hooks", "--skip-agents", "--json"]
    if capture.args.server_port:
        args += ["--server", "--external", "--server-host", "127.0.0.1", "--server-port",
                 str(capture.args.server_port), "--database", "c0_" + capture.root.name.replace("-", "_"),
                 "--server-user", "root"]
        capture.env["BEADS_DOLT_SERVER_USER"] = "c0_unselected_user"
        write_json(capture.output / "init-user-precedence.json",
                   {"environment_user": "c0_unselected_user", "explicit_user": "root"})
    try:
        initialized = envelope(capture.success("init", args))
    finally:
        capture.env.pop("BEADS_DOLT_SERVER_USER", None)
    require(initialized.get("scope") == SCOPE, "initialized Scope mismatch")
    require(initialized.get("backend") == ("server" if capture.args.server_port else "embedded"),
            "initialized backend mismatch")
    require(initialized.get("memoryComplete") is False, "preview must not claim complete Memory")
    capture.passed("normal graph initialization")
    metadata = capture.work / ".beads" / "metadata.json"
    original = metadata.read_bytes()
    config = json.loads(original)
    require(config.get("graph_mode") == "link", "missing graph mode marker")
    require(config.get("graph_scope_url") == SCOPE, "missing Scope binding")
    require(config.get("graph_workspace") == str(metadata.parent.resolve()), "missing workspace binding")
    require(config.get("graph_schema_version") == 2 and config.get("graph_ready") is True,
            "graph schema/readiness was not published")
    require(bool(config.get("graph_authority_id")), "missing authority binding")
    for name, body, title in [("plan", "Remember the deployment plan", "Plan"),
                              ("unicode", "雪 😀 e\u0301\nline two", "Mémoire 🧭"),
                              ("empty", "", "Empty body")]:
        path = "beads/" + name
        created = record(capture.success(name + "-create", remember(path, body, title)), path, body, title)
        shown = record(capture.success(name + "-read", ["show", path, "--json"]), path, body, title)
        reopened = record(capture.success(name + "-reopen", ["show", path, "--json"]), path, body, title)
        require(shown == created == reopened, f"fresh-process record changed: {name}")
        capture.passed(name + " create, read and second fresh-process exact read")

    issue_args = ["create", "Fix deployment", "--id", "beads/work", "--description",
                  "Issue body — 雪", "--type", "enhancement", "--priority", "1",
                  "--labels", " demo , ,demo", "--label", "demo", "--json"]
    issue = envelope(capture.success("issue-create", issue_args))
    require(issue.get("id") == SCOPE + "beads/work", "Issue canonical identity mismatch")
    require(issue.get("type") != SCOPE + "types/preview-memory-v1", "Issue was represented as Memory")
    require(all(issue.get(key) for key in ["type", "revision", "version"]), "missing Issue graph identity/version")
    require(issue.get("owned") == [], "Issue unexpectedly has owned Links")
    properties = issue.get("properties", {})
    require(properties.get("title") == "Fix deployment" and properties.get("description") == "Issue body — 雪",
            "Issue content changed")
    require(properties.get("issue_type") == "feature" and properties.get("priority") == 1 and
            properties.get("labels") == ["demo"] and properties.get("id", "").startswith("demo-"),
            "Issue classification, labels or configured backing prefix changed")
    for name in ["issue-read", "issue-reopen"]:
        require(envelope(capture.success(name, ["show", "beads/work", "--json"])) == issue,
                "fresh process changed Issue graph record")
    refusal(capture.run("memory-collides-with-issue", remember("beads/work", "collision", "Collision")),
            {"identity_reserved"}, "shared Issue/Memory allocation")
    refusal(capture.run("issue-collides-with-memory", ["create", "collision", "--id", "beads/plan", "--json"]),
            {"identity_reserved"}, "shared Memory/Issue allocation")
    for flag in ["--ephemeral", "--no-history", "--deps=blocks:beads/plan"]:
        refusal(capture.run("issue-unsupported-" + flag.split("=")[0].lstrip("-"),
                            ["create", "Unsupported", "--id", "beads/refused-issue", flag, "--json"]),
                {"capability_unavailable"}, "unsupported Issue effects")
        refusal(capture.run("issue-unsupported-absent", ["show", "beads/refused-issue", "--json"]),
                {"not_found"}, "refused Issue absent")
    capture.passed("Issue create, read and reopen with shared allocation and explicit unsupported-effect refusal")

    # A barrier aligns launches. Actual process intervals are retained; this
    # is competing independent CLI processes, not a proof of internal overlap.
    for same_id in [True, False]:
        barrier = threading.Barrier(2)
        def create(index):
            path = "beads/race" if same_id else f"beads/parallel-{index}"
            body = f"racer {index}"
            barrier.wait(timeout=10)
            result = capture.run(f"race-{same_id}-{index}", remember(path, body, "Concurrent"))
            return path, body, result
        with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:
            futures = [pool.submit(create, i) for i in range(2)]
            results = [future.result() for future in futures]
        winners = []
        losers = []
        unknowns = []
        for path, body, result in results:
            receipt, out, err = result
            if receipt["exit_code"] == 0:
                winners.append((path, body, record(json.loads(out), path, body, "Concurrent")))
            else:
                try:
                    code = json.loads(err).get("code")
                except (json.JSONDecodeError, AttributeError):
                    code = None
                if code == "outcome_unknown":
                    # Observe only: never replay, infer rollback, or count this
                    # as conflict/success even if a later current read exists.
                    observation = capture.run("unknown-outcome-observe", ["show", path, "--json"])
                    gap = {"check": "concurrent create", "path": path, "code": code,
                           "observation_exit_code": observation[0]["exit_code"],
                           "limitation": "current read does not account for allocation, history and effects"}
                    capture.qualification_gaps.append(gap)
                    unknowns.append(path)
                else:
                    refusal(result, {"identity_reserved", "revision_conflict"} if same_id else {"revision_conflict"},
                            "concurrent create")
                    losers.append(path)
        require(not unknowns, "concurrent outcome_unknown leaves C0 unqualified; observations retained, no replay")
        require(len(winners) == 1 if same_id else len(winners) >= 1,
                "unexpected concurrent winner count")
        for path, body, created in winners:
            shown = record(capture.success("race-read", ["show", path, "--json"]), path, body, "Concurrent")
            require(created == shown, "concurrent winner changed after reopen")
        if not same_id:
            for path in losers:
                refusal(capture.run("conflicted-ID-absent", ["show", path, "--json"]), {"not_found"},
                        "conflicted ID read")
        capture.passed("duplicate-ID single winner and typed refusal" if same_id else "different-ID coherent success/conflict")

    controls = [
        ("missing-metadata", lambda _: None, {"graph_not_initialized"}),
        ("missing-mode", lambda value: {k: v for k, v in value.items() if k != "graph_mode"}, {"graph_not_initialized"}),
        ("wrong-mode", lambda value: dict(value, graph_mode="dependency"), {"graph_not_initialized"}),
        ("unsupported-mode", lambda value: dict(value, graph_mode="future-c0"), {"capability_unavailable"}),
        ("wrong-scope", lambda value: dict(value, graph_scope_url=SCOPE_INPUT + "-wrong/"), {"graph_not_initialized"}),
        ("wrong-authority", lambda value: dict(value, graph_authority_id="0" * 32), {"graph_not_initialized"}),
        ("wrong-workspace", lambda value: dict(value, graph_workspace=str(capture.root)), {"not_authority"}),
        ("wrong-schema", lambda value: dict(value, graph_schema_version=999), {"graph_not_initialized"}),
        ("not-ready", lambda value: dict(value, graph_ready=False), {"graph_not_initialized"}),
        ("wrong-database", lambda value: dict(value, dolt_database="missing_c0_" + capture.root.name.replace("-", "_")), {"graph_not_initialized"}),
    ]
    for backend in ["postgres", "mysql", "sqlite", "unsupported-c0"]:
        controls.append(("wrong-backend-" + backend, lambda value, backend=backend: dict(value, backend=backend),
                         {"graph_not_initialized"}))
    for field in ["graph_scope_url", "graph_authority_id", "graph_workspace", "graph_schema_version", "graph_ready", "dolt_database"]:
        controls.append(("missing-" + field, lambda value, field=field: {k: v for k, v in value.items() if k != field},
                         {"graph_not_initialized"}))
    for name, transform, codes in controls:
        value = transform(config)
        path = "beads/refused-" + name
        if value is None:
            metadata.unlink()
        else:
            write_json(metadata, value)
        try:
            before = tree_digest(capture.work)
            refusal(capture.run(name, remember(path, "refuse", "Refuse")), codes, name)
            after = tree_digest(capture.work)
            write_json(capture.output / f"{name}-trees.json", {"before": before, "after": after})
            require(before == after, f"{name} changed workspace bytes")
        finally:
            metadata.write_bytes(original)
        refusal(capture.run(name + "-absent", ["show", path, "--json"]), {"not_found"}, name + " post-refusal read")
        capture.passed(name + " typed refusal, unchanged workspace and absent canonical ID")
    capture.passed("metadata restored after refusal controls")

    env_controls = [
        ("BEADS_DB", str(capture.root / "unselected" / ".beads" / "dolt"), {"capability_unavailable"}),
        ("BEADS_DOLT_SERVER_HOST", "unselected.invalid", {"not_authority"}),
        ("BEADS_DOLT_SERVER_PORT", "invalid", {"not_authority"}),
        ("BEADS_DOLT_SERVER_TLS", "true", {"not_authority"}),
        ("BEADS_DOLT_PROXIED_SERVER", "1", {"not_authority"}),
    ]
    if capture.args.server_port:
        env_controls.append(("BEADS_DOLT_CREDENTIAL_COMMAND", "false", {"capability_unavailable"}))
    for key, value, codes in env_controls:
        path = "beads/refused-env-" + key.lower()
        capture.env[key] = value
        try:
            before = tree_digest(capture.work)
            refusal(capture.run(key, remember(path, "refuse", "Refuse")), codes, key)
            require(tree_digest(capture.work) == before, f"{key} changed workspace bytes")
        finally:
            capture.env.pop(key, None)
        refusal(capture.run(key + "-absent", ["show", path, "--json"]), {"not_found"}, key + " absent")
        capture.passed(key + " conflicting route refused without effects")

    yaml_path = capture.work / ".beads" / "config.yaml"
    original_yaml = yaml_path.read_bytes() if yaml_path.exists() else None
    try:
        if capture.args.server_port:
            for value in ["1", "invalid"]:
                yaml_path.write_text("dolt.port: " + value + "\n")
                path = "beads/refused-yaml-port-" + value
                before = tree_digest(capture.work)
                refusal(capture.run("yaml-port-" + value, remember(path, "refuse", "Refuse")),
                        {"not_authority"}, "YAML port assertion")
                require(tree_digest(capture.work) == before, "YAML port refusal changed workspace")
                yaml_path.unlink()
                refusal(capture.run("yaml-port-absent-" + value, ["show", path, "--json"]),
                        {"not_found"}, "YAML port absent")
            yaml_path.write_text("dolt.port: " + str(capture.args.server_port) + "\n")
            capture.success("yaml-port-matching", ["show", "beads/plan", "--json"])
            yaml_path.write_text("dolt.port: 1\n")
            capture.env["BEADS_DOLT_SERVER_PORT"] = str(capture.args.server_port)
            try:
                capture.success("yaml-port-environment-override", ["show", "beads/plan", "--json"])
            finally:
                capture.env.pop("BEADS_DOLT_SERVER_PORT", None)
            capture.passed("YAML port assertion refuses conflicts and honors environment precedence")
        else:
            yaml_path.write_text("dolt.port: 1\n")
            capture.success("embedded-ambient-yaml-port", ["show", "beads/plan", "--json"])
            capture.passed("ambient YAML port does not change embedded route")
        for storage_class in ["ephemeral", "unversioned", "invalid"]:
            yaml_path.write_text("storage-class.task: " + storage_class + "\n")
            before = tree_digest(capture.work)
            code = "invalid_properties" if storage_class == "invalid" else "capability_unavailable"
            refusal(capture.run("issue-configured-" + storage_class,
                                ["create", "Unsupported class", "--id", "beads/refused-class", "--json"]),
                    {code}, "configured Issue storage class")
            require(tree_digest(capture.work) == before, "configured class refusal changed workspace")
            yaml_path.unlink()
            refusal(capture.run("issue-class-absent", ["show", "beads/refused-class", "--json"]),
                    {"not_found"}, "configured class refusal absent")
        capture.passed("configured unsupported Issue retention classes refuse without publication")
    finally:
        if original_yaml is None:
            yaml_path.unlink(missing_ok=True)
        else:
            yaml_path.write_bytes(original_yaml)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--bd", required=True, type=Path, help="absolute installed bd executable")
    parser.add_argument("--output-dir", "--outputdir", required=True, type=Path, help="new artifact directory")
    parser.add_argument("--server-port", type=int, help="caller-owned ordinary loopback Dolt server")
    parser.add_argument("--server-root", type=Path, help="existing disposable server data root (provenance only)")
    parser.add_argument("--command-timeout", type=float, default=120)
    parser.add_argument("--total-timeout", type=float, default=900)
    args = parser.parse_args()
    require(os.name == "posix", "POSIX process-group capture is required")
    require(args.bd.is_absolute() and args.bd.is_file() and os.access(args.bd, os.X_OK), "--bd must be an absolute executable")
    args.bd = args.bd.resolve()
    args.output_dir = args.output_dir.absolute()
    require(args.command_timeout > 0 and args.total_timeout > 0, "timeouts must be positive")
    require(bool(args.server_port) == bool(args.server_root), "supply both --server-port and --server-root")
    if args.server_port:
        require(1 <= args.server_port <= 65535, "invalid server port")
        args.server_root = args.server_root.resolve(strict=True)
        require(args.server_root.is_dir(), "server root must be a directory")
    capture = Capture(args)
    for sig in [signal.SIGINT, signal.SIGTERM]:
        signal.signal(sig, lambda signum, _frame: (capture.stop(), (_ for _ in ()).throw(KeyboardInterrupt(signum))))
    failure = None
    try:
        exercise(capture)
        require(sha256(args.bd) == capture.binary_hash, "binary changed during capture")
    except BaseException as exc:
        failure = f"{type(exc).__name__}: {exc}"
    finally:
        capture.stop()
        write_json(capture.output / "summary.json", {
            "passed": failure is None, "c0_qualified": False, "failure": failure, "checks": capture.passes,
            "qualification_gaps": capture.qualification_gaps,
            "commands": capture.records, "private_root": str(capture.root),
            "active_child_count": len(capture.active),
            "limits": ["this smoke harness alone cannot qualify C0",
                       "workspace digest and absent-ID reads do not prove unchanged remote database or atomic retained history/effects",
                       "process launch barrier does not prove internal transaction overlap",
                       "no cancellation, crash, network-loss or uncertain-commit injection",
                       "no Linux/macOS platform-matrix qualification",
                       "no historical-read, graph-Link, migration, or performance claim"],
        })
    print(f"{'PASS' if failure is None else 'FAIL'} capture: {capture.output}", flush=True)
    if failure:
        print(failure, flush=True)
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
