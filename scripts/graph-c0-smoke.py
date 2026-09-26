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

    def run(self, label, arguments, input_text=None):
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
        stdin_file = None
        if input_text is not None:
            stdin_path = directory / "stdin.txt"
            stdin_path.write_bytes(input_text.encode("utf-8"))
            receipt["stdin_sha256"] = sha256(stdin_path)
            stdin_file = stdin_path.open("rb")
        child = None
        failure = None
        try:
            with (directory / "stdout.log").open("wb") as out, (directory / "stderr.log").open("wb") as err:
                child = subprocess.Popen(argv, cwd=self.work, env=self.env,
                                         stdin=stdin_file if stdin_file is not None else subprocess.DEVNULL, stdout=subprocess.PIPE,
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
            if stdin_file is not None:
                stdin_file.close()
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

    def success(self, label, arguments, input_text=None):
        receipt, out, err = self.run(label, arguments, input_text=input_text)
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
    require(value["type"] == SCOPE + "types/preview-memory-v2", "expected installed non-Issue Memory descriptor")
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
    require(config.get("graph_schema_version") == 5 and config.get("graph_ready") is True,
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
    require(issue.get("type") != SCOPE + "types/preview-memory-v2", "Issue was represented as Memory")
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


def exercise_dependency_workflow(capture):
    """Installed-process wiring, ownership and workflow evidence, not History qualification."""
    release_id = "beads/release"
    prerequisite_id = "beads/prerequisite"
    dependency_type = SCOPE + "types/preview-blocks-v1"

    def show(label, path):
        return envelope(capture.success(label, ["show", path, "--json"]))

    def issue(label, path, title):
        value = envelope(capture.success(label, ["create", title, "--id", path, "--json"]))
        require(value.get("id") == SCOPE + path, "Issue canonical identity mismatch")
        require(value.get("type") == SCOPE + "types/preview-issue-v2", "wrong ownership-aware Issue descriptor")
        require(value.get("properties", {}).get("status") == "open", "new Issue not open")
        require(value.get("owned") == [], "new Issue already owns Links")
        require(value.get("revision") and value.get("version"), "Issue lacks revision/version")
        return value

    def ready(label, expected):
        value = capture.success(label, ["ready", "--json"])
        require(isinstance(value, dict) and value.get("schemaVersion") == 1 and value.get("preview") is True,
                "ready lacks preview envelope")
        items = value.get("result")
        require(isinstance(items, list), "ready result is not an array")
        ids = [item.get("id") for item in items]
        require(len(ids) == len(set(ids)), "ready returned duplicate Issues")
        require(set(ids) == {SCOPE + path for path in expected}, f"wrong ready set: {ids}")
        require(all(item.get("type") == SCOPE + "types/preview-issue-v2" for item in items),
                "ready returned non-Issue records")
        return items

    def dependency(label, argv, source, target, changed):
        result = envelope(capture.success(label, [*argv, "--json"]))
        require(result.get("changed") is changed, f"{label}: wrong changed indication")
        link = result.get("link", {})
        require(isinstance(link.get("id"), str) and link["id"].startswith(SCOPE + "links/") and
                len(link["id"]) > len(SCOPE + "links/"), "Dependency lacks canonical Link identity")
        require(link.get("type") == dependency_type, "Dependency has wrong registered Type")
        require(link.get("source") == SCOPE + source and link.get("target") == SCOPE + target,
                "Dependency endpoints changed")
        require(link.get("revision") and link.get("version"), "Link lacks revision/version")
        require(link.get("properties") == {}, "unexpected Dependency properties")
        owner = result.get("source", {})
        require(owner.get("id") == SCOPE + source and owner.get("owned") == [link],
                "source does not own exact canonical Link state")
        require(show(label + "-link-reopen", link["id"]) == link, "fresh process changed Link")
        require(show(label + "-source-reopen", source) == owner, "fresh process changed source")
        return result

    release = issue("workflow-release-create", release_id, "Release the deployment")
    prerequisite = issue("workflow-prerequisite-create", prerequisite_id, "Verify deployment")
    for path, body, title in [("beads/workflow-plan", "Deployment context", "Deployment plan"),
                              ("beads/workflow-decision", "Deploy after verification", "Decision")]:
        created = record(capture.success("workflow-memory-create", remember(path, body, title)), path, body, title)
        require(show("workflow-memory-reopen", path) == created, "workflow Memory changed after reopen")
    ready("workflow-ready-before", ["beads/work", release_id, prerequisite_id])
    capture.passed("normal creation of two workflow Issues and two Memories; ready includes Issues only")

    added = dependency("workflow-dep-add", ["dep", "add", release_id, prerequisite_id],
                       release_id, prerequisite_id, True)
    owner = added["source"]
    require(owner["revision"] != release["revision"] and owner["version"] != release["version"],
            "adding owned Dependency did not advance source revision and version")
    require(owner["properties"] == release["properties"], "adding Dependency changed Issue properties")
    require(show("workflow-target-unchanged", prerequisite_id) == prerequisite,
            "adding incoming Dependency changed target")
    ready("workflow-ready-blocked", ["beads/work", prerequisite_id])
    capture.passed("blocking Link creation advances owning source, leaves target unchanged and removes source from ready")

    for label, argv in [
        ("workflow-dep-repeat", ["dep", "add", release_id, prerequisite_id]),
        ("workflow-dep-alias-repeat", ["dep", "add", release_id, prerequisite_id, "--type", "blocked-by"]),
        ("workflow-generic-repeat", ["link", release_id, prerequisite_id, "--resource-type", dependency_type,
                                      "--if-source-revision", owner["revision"]]),
        ("workflow-generic-unconditional-repeat", ["link", release_id, prerequisite_id,
                                                    "--resource-type", dependency_type, "--unconditional-source"]),
    ]:
        repeated = dependency(label, argv, release_id, prerequisite_id, False)
        require(repeated["link"] == added["link"] and repeated["source"] == owner,
                "repeated relationship assertion minted or changed graph state")
    capture.passed("familiar, alias and generic repeated assertions preserve Link identity and source version")

    protected = {release_id: owner, prerequisite_id: prerequisite,
                 "beads/workflow-plan": show("workflow-memory-before-refusals", "beads/workflow-plan")}
    refusals = [
        ("workflow-required-source-guard", ["link", release_id, prerequisite_id,
                                            "--resource-type", dependency_type], {"invalid_selector"}),
        ("workflow-stale-source-guard", ["link", release_id, prerequisite_id,
                                         "--resource-type", dependency_type, "--if-source-revision", release["revision"]],
         {"revision_conflict"}),
        ("workflow-conflicting-source-guards", ["link", release_id, prerequisite_id,
                                               "--resource-type", dependency_type, "--if-source-revision", owner["revision"],
                                               "--unconditional-source"], {"invalid_selector"}),
        ("workflow-blocked-close", ["close", release_id], {"constraint_violation"}),
        ("workflow-cycle", ["dep", "add", prerequisite_id, release_id], {"invalid_properties"}),
        ("workflow-memory-source", ["dep", "add", "beads/workflow-plan", prerequisite_id], {"invalid_properties"}),
        ("workflow-memory-target", ["dep", "add", release_id, "beads/workflow-plan"], {"invalid_properties"}),
        ("workflow-memory-close", ["close", "beads/workflow-plan"], {"invalid_properties"}),
        ("workflow-unsupported-deptype", ["dep", "add", release_id, prerequisite_id, "--type", "related"],
         {"capability_unavailable"}),
        ("workflow-unregistered-type", ["link", release_id, prerequisite_id, "--resource-type", SCOPE + "types/unknown",
                                         "--unconditional-source"],
         {"capability_unavailable"}),
        ("workflow-force-close", ["close", release_id, "--force"], {"capability_unavailable"}),
    ]
    for label, argv, codes in refusals:
        refusal(capture.run(label, [*argv, "--json"]), codes, label)
        for path, expected in protected.items():
            require(show(label + "-unchanged", path) == expected, f"{label} changed {path}")
        require(show(label + "-link-unchanged", added["link"]["id"]) == added["link"],
                f"{label} changed existing Link")
    require(show("workflow-link-after-refusals", added["link"]["id"]) == added["link"],
            "refused operations changed existing Link")
    ready("workflow-ready-after-refusals", ["beads/work", prerequisite_id])
    capture.passed("missing/stale/conflicting source guards, blocked close, cycle, Memory endpoints and unsupported mutations refuse without observable record changes")

    for value, code in [("1", "capability_unavailable"), ("malformed", "invalid_properties")]:
        previous = capture.env.get("BEADS_MAX_ROWS")
        capture.env["BEADS_MAX_ROWS"] = value
        try:
            before = tree_digest(capture.work)
            refusal(capture.run("workflow-ready-env-limit-" + value, ["ready", "--json"]),
                    {code}, "ready environment row limit")
            require(tree_digest(capture.work) == before, "ready limit refusal changed workspace bytes")
        finally:
            if previous is None:
                capture.env.pop("BEADS_MAX_ROWS", None)
            else:
                capture.env["BEADS_MAX_ROWS"] = previous
        for path, expected in protected.items():
            require(show("workflow-ready-limit-unchanged", path) == expected,
                    f"ready environment limit {value} changed {path}")
        require(show("workflow-ready-limit-link-unchanged", added["link"]["id"]) == added["link"],
                "ready environment limit changed existing Link")
    ready("workflow-ready-after-env-limit", ["beads/work", prerequisite_id])
    capture.passed("configured nonzero and malformed ready row limits refuse without observable changes")

    closed = envelope(capture.success("workflow-prerequisite-close", ["close", prerequisite_id, "--json"]))
    require(closed.get("changed") is True, "first close was not a change")
    closed_issue = closed.get("issue", {})
    require(closed_issue.get("id") == prerequisite["id"] and closed_issue.get("owned") == [],
            "close changed target identity or ownership")
    require(closed_issue.get("properties", {}).get("status") == "closed", "close did not close Issue")
    require(closed_issue.get("revision") != prerequisite["revision"] and
            closed_issue.get("version") != prerequisite["version"], "close did not advance target revision/version")
    require(show("workflow-closed-reopen", prerequisite_id) == closed_issue, "closed Issue changed on reopen")
    require(show("workflow-owner-after-target-close", release_id) == owner,
            "derived readiness unexpectedly changed source version or owned Link")
    ready("workflow-ready-unblocked", ["beads/work", release_id])
    repeated_close = envelope(capture.success("workflow-close-repeat", ["close", prerequisite_id, "--json"]))
    require(repeated_close.get("changed") is False and repeated_close.get("issue") == closed_issue,
            "repeated close minted a new version")
    capture.passed("closing prerequisite restores source readiness; repeated close is a version-preserving no-op")

    # The baseline Issue proves the generic spelling can create a new edge,
    # rather than only recognizing the familiar spelling's existing edge.
    generic_source = show("workflow-generic-source-before", "beads/work")
    generic = dependency("workflow-generic-create", ["link", "beads/work", prerequisite_id,
                         "--resource-type", dependency_type, "--if-source-revision", generic_source["revision"]],
                         "beads/work", prerequisite_id, True)
    require(generic["link"]["id"] != added["link"]["id"], "independent relationship reused Link identity")
    require(show("workflow-closed-target-still-unchanged", prerequisite_id) == closed_issue,
            "generic creation changed its target")
    ready("workflow-ready-final", ["beads/work", release_id])
    capture.passed("generic spelling creates a distinct canonical Link through the same workflow adapter")

    metadata = capture.work / ".beads" / "metadata.json"
    original = metadata.read_bytes()
    try:
        config = json.loads(original)
        config["graph_schema_version"] = 2
        write_json(metadata, config)
        before = tree_digest(capture.work)
        refusal(capture.run("workflow-old-preview-refused", ["close", release_id, "--json"]),
                {"graph_not_initialized"}, "old preview schema")
        require(tree_digest(capture.work) == before, "old preview refusal changed workspace bytes")
    finally:
        metadata.write_bytes(original)
    require(show("workflow-current-preview-reopen", release_id) == owner, "schema refusal changed owner")
    capture.passed("previous preview metadata refuses without migration or current-state effects")
    write_json(capture.output / "workflow-versions.json", {
        "release_before_dependency": release, "release_after_dependency": owner,
        "prerequisite_before_close": prerequisite, "prerequisite_after_close": closed_issue,
        "blocking_link": added["link"], "generic_created_link": generic["link"],
        "limitation": "observed current records and opaque version transitions; historical retrieval and retained-state atomicity require separate evidence",
    })


def exercise_mixed_links(capture):
    """Prove CLI JSON/guard wiring and fresh-process current reads of mixed Links."""
    related_type = SCOPE + "types/preview-related-v2"
    issue_path, source_path, target_path = "beads/release", "beads/workflow-plan", "beads/workflow-decision"

    def show(label, path):
        return envelope(capture.success(label, ["show", path, "--json"]))

    def ready(label):
        value = capture.success(label, ["ready", "--json"])
        require(value.get("preview") is True and value.get("schemaVersion") == 1,
                "mixed-Link ready lacks preview envelope")
        require(isinstance(value.get("result"), list), "mixed-Link ready result is not an array")
        return value["result"]

    def mutation(label, argv, path, source, target, properties, changed=True, input_text=None):
        result = envelope(capture.success(label, [*argv, "--json"], input_text=input_text))
        require(result.get("changed") is changed, f"{label}: wrong changed indication")
        link = result.get("link", {})
        require(link.get("id") == SCOPE + path and link.get("type") == related_type,
                f"{label}: Link identity or registered Type changed")
        require(link.get("source") == SCOPE + source and link.get("target") == SCOPE + target,
                f"{label}: Link endpoints changed")
        require(link.get("properties") == properties, f"{label}: Link properties changed")
        require(link.get("revision") and link.get("version"), f"{label}: missing Link revision/version")
        require(result.get("source", {}).get("id") == SCOPE + source, f"{label}: missing affected source")
        require(show(label + "-link-reopen", path) == link, f"{label}: Link changed in fresh process")
        require(show(label + "-source-reopen", source) == result["source"],
                f"{label}: source changed in fresh process")
        return result

    issue = show("mixed-issue-before", issue_path)
    source = show("mixed-source-before", source_path)
    target = show("mixed-target-before", target_path)
    ready_before = ready("mixed-ready-before")
    require(source.get("type") == SCOPE + "types/preview-memory-v2" and source.get("owned") == [],
            "mixed source must start as an unlinked ownership-aware Memory")

    # Issue ownership currently admits blocking Dependencies only. This
    # informational edge must not silently broaden that provisional contract.
    issue_link = mutation("mixed-issue-memory-create", ["link", issue_path, target_path,
                          "--resource-type", related_type, "--id", "links/issue-context",
                          "--properties", '{"note":"Issue context"}'],
                          "links/issue-context", issue_path, target_path, {"note": "Issue context"})
    require(issue_link["source"] == issue, "unowned informational Link changed Issue revision/history state")
    issue_updated = mutation("mixed-issue-link-update", ["update", "links/issue-context", "--properties",
                             '{"note":"Updated Issue context"}', "--if-revision", issue_link["link"]["revision"]],
                             "links/issue-context", issue_path, target_path, {"note": "Updated Issue context"})
    require(issue_updated["source"] == issue, "unowned Link property update changed Issue")
    issue_noop = mutation("mixed-unowned-optional-source-guard", ["update", "links/issue-context", "--properties",
                          '{"note":"Updated Issue context"}', "--if-revision", issue_updated["link"]["revision"],
                          "--if-source-revision", issue["revision"]],
                          "links/issue-context", issue_path, target_path, {"note": "Updated Issue context"}, changed=False)
    require(issue_noop["source"] == issue and issue_noop["link"] == issue_updated["link"],
            "unowned guarded semantic no-op changed Issue or Link")
    require(issue_updated["link"]["revision"] != issue_link["link"]["revision"] and
            issue_updated["link"]["version"] != issue_link["link"]["version"],
            "real unowned Link property change did not advance its revision/version")
    require(show("mixed-target-after-issue-link", target_path) == target, "incoming Issue Link changed target")
    require(ready("mixed-ready-after-issue-link") == ready_before, "informational Issue Link affected ready")
    capture.passed("Issue→Memory informational Link create/read/update preserves Issue and target; ready unaffected")

    memory_link = mutation("mixed-memory-memory-create", ["link", SCOPE + source_path, SCOPE + target_path,
                           "--resource-type", related_type, "--id", "links/memory-context",
                           "--properties", '{"note":"Memory context"}', "--if-source-revision", source["revision"]],
                           "links/memory-context", source_path, target_path, {"note": "Memory context"})
    owner = memory_link["source"]
    require(owner.get("owned") == [memory_link["link"]], "Memory does not own full created Link state")
    require(owner["revision"] != source["revision"] and owner["version"] != source["version"],
            "owned Link creation did not advance source revision/version")
    require(owner["properties"] == source["properties"], "owned Link creation changed source properties")
    require(show("mixed-target-after-memory-link", target_path) == target, "incoming Memory Link changed target")
    capture.passed("Memory→Memory Link owns complete Link state and advances source only; canonical URL selectors work")

    properties_file = capture.output / "mixed-properties.json"
    write_json(properties_file, {"note": "Updated context — 雪"})
    updated = mutation("mixed-owned-property-file", ["update", SCOPE + "links/memory-context",
                       "--properties", "@" + str(properties_file), "--if-revision", memory_link["link"]["revision"],
                       "--if-source-revision", owner["revision"]],
                       "links/memory-context", source_path, target_path, {"note": "Updated context — 雪"})
    current = updated["source"]
    require(updated["link"]["revision"] != memory_link["link"]["revision"] and
            updated["link"]["version"] != memory_link["link"]["version"], "property update did not version Link")
    require(current["revision"] != owner["revision"] and current["version"] != owner["version"],
            "owned Link property update did not version source")
    require(current["properties"] == source["properties"] and current.get("owned") == [updated["link"]],
            "source did not retain exact new owned-Link content with unchanged properties")
    repeated = mutation("mixed-owned-noop-stdin", ["update", "links/memory-context", "--properties", "@-",
                        "--if-revision", updated["link"]["revision"], "--if-source-revision", current["revision"]],
                        "links/memory-context", source_path, target_path, {"note": "Updated context — 雪"},
                        changed=False, input_text=' { "note" : "Updated context — 雪" }\n')
    require(repeated["link"] == updated["link"] and repeated["source"] == current,
            "semantic no-op changed Link or owning Memory")
    require(show("mixed-target-after-property", target_path) == target, "property update changed target")
    capture.passed("descriptor-admitted property replacement via @file versions Link and owner; @- semantic no-op preserves both")

    link_base = ["link", source_path, target_path, "--resource-type", related_type,
                 "--id", "links/refused-mixed", "--properties", '{"note":"Refused"}']
    update_base = ["update", "links/memory-context", "--properties", '{"note":"Refused"}']
    link_guard = ["--if-revision", updated["link"]["revision"]]
    source_guard = ["--if-source-revision", current["revision"]]
    bad_guard = {"invalid_selector", "invalid_properties"}
    refusals = [
        ("mixed-create-missing-source-guard", link_base, bad_guard),
        ("mixed-create-stale-source-guard", [*link_base, "--if-source-revision", source["revision"]], {"revision_conflict"}),
        ("mixed-update-missing-link-guard", [*update_base, *source_guard], bad_guard),
        ("mixed-update-missing-source-guard", [*update_base, *link_guard], bad_guard),
        ("mixed-update-stale-link-guard", [*update_base, "--if-revision", memory_link["link"]["revision"], *source_guard],
         {"revision_conflict"}),
        ("mixed-update-stale-source-guard", [*update_base, *link_guard, "--if-source-revision", owner["revision"]],
         {"revision_conflict"}),
        ("mixed-noop-stale-link-guard", ["update", "links/memory-context", "--properties",
                                        json.dumps(updated["link"]["properties"]), "--if-revision",
                                        memory_link["link"]["revision"], *source_guard], {"revision_conflict"}),
        ("mixed-noop-stale-source-guard", ["update", "links/memory-context", "--properties",
                                          json.dumps(updated["link"]["properties"]), *link_guard,
                                          "--if-source-revision", owner["revision"]], {"revision_conflict"}),
        ("mixed-unowned-stale-optional-source-guard", ["update", "links/issue-context", "--properties",
                                                     json.dumps(issue_updated["link"]["properties"]),
                                                     "--if-revision", issue_updated["link"]["revision"],
                                                     "--if-source-revision", source["revision"]], {"revision_conflict"}),
        ("mixed-update-conflicting-link-guards", [*update_base, *link_guard, "--unconditional", *source_guard], bad_guard),
        ("mixed-update-conflicting-source-guards", [*update_base, *link_guard, *source_guard, "--unconditional-source"], bad_guard),
        ("mixed-explicit-ID-reuse", ["link", source_path, target_path, "--resource-type", related_type,
                                    "--id", "links/memory-context", "--properties", '{"note":"Duplicate"}', *source_guard],
         {"identity_reserved"}),
    ]
    for name, payload in [("malformed", '{"note":'), ("duplicate", '{"note":"a","note":"b"}'),
                          ("non-object", '["note"]'), ("unknown-property", '{"unexpected":"x"}'),
                          ("wrong-property-type", '{"note":7}')]:
        refusals.append(("mixed-json-" + name,
                         ["update", "links/memory-context", "--properties", payload, *link_guard, *source_guard],
                         {"invalid_properties"}))
    protected = {source_path: current, target_path: target, issue_path: issue,
                 "links/memory-context": updated["link"], "links/issue-context": issue_updated["link"]}
    for label, argv, codes in refusals:
        refusal(capture.run(label, [*argv, "--json"]), codes, label)
        # Check the touched current records through new CLI processes. The
        # storage suite separately proves no partial retained-history effects.
        for path in [source_path, "links/memory-context"]:
            require(show(label + "-unchanged", path) == protected[path], f"{label}: changed {path}")
    refusal(capture.run("mixed-refused-ID-absent", ["show", "links/refused-mixed", "--json"]),
            {"not_found"}, "refused mixed Link absent")
    for path, expected in protected.items():
        require(show("mixed-refusal-final-unchanged", path) == expected, f"refused operations changed {path}")
    capture.passed("mixed-Link missing/stale/conflicting guards, duplicate ID and invalid JSON refuse without observable record changes")

    # These are actual mutation routes, admitted by the installed CLI before
    # storage opens. Local byte equality supplements post-refusal reads; it
    # does not stand in for server-side retained-history atomicity evidence.
    yaml_path = capture.work / ".beads" / "config.yaml"
    original_yaml = yaml_path.read_bytes() if yaml_path.exists() else None
    mayor = capture.work / "mayor"
    freeze_path = capture.work / "MIGRATION-FREEZE"
    require(not mayor.exists() and not freeze_path.exists(), "unexpected existing town markers")
    for policy in ["readonly-flag", "readonly-env", "readonly-config", "migration-freeze"]:
        flags = ["--readonly"] if policy == "readonly-flag" else []
        try:
            if policy == "readonly-env":
                capture.env["BD_READONLY"] = "true"
            elif policy == "readonly-config":
                yaml_path.write_text("readonly: true\n")
            elif policy == "migration-freeze":
                mayor.mkdir()
                (mayor / "town.json").write_text("{}\n")
                freeze_path.write_text("mixed-smoke\t2026-09-25T00:00:00Z\tgraph write control\n")
            for operation, argv in [("link", [*link_base, *source_guard]),
                                    ("update", [*update_base, *link_guard, *source_guard])]:
                label = "mixed-policy-" + policy + "-" + operation
                before = tree_digest(capture.work)
                refusal(capture.run(label, [*argv, *flags, "--json"]), {"permission_denied"}, label)
                require(tree_digest(capture.work) == before, f"{label}: changed workspace bytes")
        finally:
            capture.env.pop("BD_READONLY", None)
            if original_yaml is None:
                yaml_path.unlink(missing_ok=True)
            else:
                yaml_path.write_bytes(original_yaml)
            if policy == "migration-freeze":
                freeze_path.unlink(missing_ok=True)
                (mayor / "town.json").unlink(missing_ok=True)
                mayor.rmdir()
    for path in [source_path, "links/memory-context"]:
        require(show("mixed-policy-unchanged", path) == protected[path], f"write policy changed {path}")
    refusal(capture.run("mixed-policy-ID-absent", ["show", "links/refused-mixed", "--json"]),
            {"not_found"}, "policy-refused mixed Link absent")
    capture.passed("readonly flag/environment/config and migration freeze block mixed Link creation/update before workspace changes")

    # Same endpoints/Type are deliberately another Link, unlike the legacy
    # blocking Dependency assertion. Explicit IDs are identity, not replay keys.
    parallel = mutation("mixed-same-endpoint-multiedge", ["link", source_path, target_path,
                        "--resource-type", related_type, "--id", "links/memory-context-second",
                        "--properties", '{"note":"Second intentional context"}', "--unconditional-source"],
                        "links/memory-context-second", source_path, target_path,
                        {"note": "Second intentional context"})
    require(parallel["link"]["id"] != updated["link"]["id"], "equal-endpoint Links were deduplicated")
    expected_owned = sorted([updated["link"], parallel["link"]], key=lambda item: item["id"])
    require(parallel["source"].get("owned") == expected_owned, "Memory lost an independently identified Link")
    require(parallel["source"]["revision"] != current["revision"] and
            parallel["source"]["version"] != current["version"], "second Link did not version owning Memory")
    unconditional = mutation("mixed-unconditional-noop", ["update", "links/memory-context-second",
                             "--properties", '{"note":"Second intentional context"}', "--unconditional", "--unconditional-source"],
                             "links/memory-context-second", source_path, target_path,
                             {"note": "Second intentional context"}, changed=False)
    require(unconditional["link"] == parallel["link"] and unconditional["source"] == parallel["source"],
            "unconditional no-op changed Link/source")
    require(show("mixed-first-Link-still-independent", "links/memory-context") == updated["link"],
            "second Link changed first Link")
    require(show("mixed-final-target", target_path) == target and show("mixed-final-issue", issue_path) == issue,
            "informational Memory changes affected target or Issue")
    require(ready("mixed-ready-final") == ready_before, "informational Links affected readiness")
    capture.passed("equal-endpoint informational Links remain independent; complete owned set and unconditional no-op survive fresh processes")
    cleared = mutation("mixed-whole-properties-empty", ["update", "links/memory-context-second", "--properties", "{}",
                       "--if-revision", parallel["link"]["revision"],
                       "--if-source-revision", parallel["source"]["revision"]],
                       "links/memory-context-second", source_path, target_path, {})
    require(cleared["link"]["revision"] != parallel["link"]["revision"] and
            cleared["link"]["version"] != parallel["link"]["version"], "whole replacement did not version changed Link")
    require(cleared["source"]["revision"] != parallel["source"]["revision"] and
            cleared["source"]["version"] != parallel["source"]["version"], "whole replacement did not version owning Memory")
    require(cleared["source"].get("owned") == sorted([updated["link"], cleared["link"]], key=lambda item: item["id"]),
            "whole properties replacement merged old properties or changed another owned Link")
    require(show("mixed-replacement-target-unchanged", target_path) == target,
            "whole properties replacement changed target")
    capture.passed("whole properties replacement with empty object removes previous note and versions only Link and owning Memory")
    write_json(capture.output / "mixed-link-versions.json", {
        "issue_before": issue, "issue_informational_link": issue_updated["link"],
        "memory_before": source, "memory_after_create": owner, "memory_after_property_update": current,
        "memory_after_second_link": parallel["source"], "memory_after_empty_properties": cleared["source"],
        "target_unchanged": target,
        "limitation": "current reads and opaque version transitions; exact retained-state atomicity requires storage evidence; public History reads remain unavailable",
    })


def exercise_link_lifecycle(capture):
    """Prove installed listing, guards, multiedge selection and durable deletion wiring."""
    related_type = SCOPE + "types/preview-related-v2"
    blocks_type = SCOPE + "types/preview-blocks-v1"
    source_path, target_path, issue_path = "beads/workflow-plan", "beads/workflow-decision", "beads/release"
    first_path, second_path = "links/memory-context", "links/memory-context-second"

    def show(label, path):
        return envelope(capture.success(label, ["show", path, "--json"]))

    def links(label, path, direction="both", resource_type=None):
        argv = ["links", path, "--direction", direction, "--json"]
        if resource_type is not None:
            argv += ["--resource-type", resource_type]
        result = capture.success(label, argv)
        require(result.get("schemaVersion") == 1 and result.get("preview") is True,
                f"{label}: missing preview envelope")
        items = result.get("result")
        require(isinstance(items, list), f"{label}: missing Link array")
        ids = [item.get("id") for item in items]
        require(ids == sorted(set(ids)), f"{label}: Links not unique and canonically sorted")
        return items

    def unlink(label, argv, old_link, before, owned):
        result = envelope(capture.success(label, ["unlink", *argv, "--json"]))
        tombstone = result.get("link", {})
        require(result.get("changed") is True and tombstone.get("state") == "deleted",
                f"{label}: no durable deletion result")
        require(tombstone.get("id") == old_link["id"] and tombstone.get("type") == old_link["type"],
                f"{label}: tombstone identity or Type changed")
        require(tombstone.get("previousVersion") == old_link["version"], f"{label}: lost previous Version")
        require(tombstone.get("revision") and tombstone["revision"] != old_link["revision"] and
                tombstone.get("version") and tombstone["version"] != old_link["version"],
                f"{label}: deletion failed to advance Link revision/version")
        after = result.get("source", {})
        if owned:
            require(after.get("revision") != before["revision"] and after.get("version") != before["version"],
                    f"{label}: unlink failed to version owning source")
            require(after.get("properties") == before["properties"] and
                    after.get("owned") == [item for item in before["owned"] if item["id"] != old_link["id"]],
                    f"{label}: source did not remove exactly the selected owned Link")
        else:
            require(after == before, f"{label}: unowned unlink changed source")
        require(show(label + "-source-reopen", after["id"]) == after,
                f"{label}: source changed across fresh processes")
        refusal(capture.run(label + "-deleted-read", ["show", old_link["id"], "--json"]), {"gone"}, label)
        return result

    source = show("lifecycle-source-before", source_path)
    target = show("lifecycle-target-before", target_path)
    issue = show("lifecycle-issue-before", issue_path)
    first = show("lifecycle-first-before", first_path)
    second = show("lifecycle-second-before", second_path)
    issue_link = show("lifecycle-issue-link-before", "links/issue-context")
    require(links("lifecycle-source-out", SCOPE + source_path, "out") == [first, second],
            "source outgoing listing lost multiedges")
    require(links("lifecycle-source-in", source_path, "in") == [], "source has unexpected incoming Links")
    require(links("lifecycle-source-both", source_path) == [first, second], "source both listing differs")
    expected_incoming = sorted([first, second, issue_link], key=lambda item: item["id"])
    require(links("lifecycle-target-in", target_path, "in", related_type) == expected_incoming,
            "incoming listing lost mixed Issue/Memory sources")
    require(links("lifecycle-target-out", target_path, "out") == [], "target outgoing listing is not empty")
    blocking = links("lifecycle-issue-blocks", issue_path, "out", blocks_type)
    require(len(blocking) == 1 and blocking[0]["type"] == blocks_type,
            "specialized blocking Dependency missing from generic listing")
    require(links("lifecycle-issue-all", issue_path) == sorted([blocking[0], issue_link], key=lambda item: item["id"]),
            "generic listing did not combine specialized and informational Links")
    require(links("lifecycle-issue-related", issue_path, "out", related_type) == [issue_link],
            "registered Type filter returned wrong Links")
    capture.passed("fresh-process incident listing combines generic and specialized Links; exact Type, in/out/both and canonical URL selectors preserve multiedges")

    guarded = [first_path, "--if-revision", first["revision"], "--if-source-revision", source["revision"]]
    bad_guard = {"invalid_selector", "invalid_properties"}
    refusals = [
        ("missing-link-guard", [first_path, "--if-source-revision", source["revision"]], bad_guard),
        ("missing-source-guard", [first_path, "--if-revision", first["revision"]], bad_guard),
        ("stale-link-guard", [first_path, "--if-revision", "stale", "--if-source-revision", source["revision"]], {"revision_conflict"}),
        ("stale-source-guard", [first_path, "--if-revision", first["revision"], "--if-source-revision", "stale"], {"revision_conflict"}),
        ("conflicting-link-guards", [*guarded, "--unconditional"], bad_guard),
        ("conflicting-source-guards", [*guarded, "--unconditional-source"], bad_guard),
        ("blocking-adapter-required", [blocking[0]["id"], "--if-revision", blocking[0]["revision"],
                                       "--if-source-revision", issue["revision"]], {"capability_unavailable"}),
        ("missing-pair", [target_path, source_path, "--resource-type", related_type,
                          "--unconditional", "--unconditional-source"], {"not_found"}),
    ]
    for name, argv, codes in refusals:
        label = "lifecycle-refuse-" + name
        refusal(capture.run(label, ["unlink", *argv, "--json"]), codes, label)
    ambiguity = refusal(capture.run("lifecycle-pair-ambiguous", ["unlink", source_path, target_path,
                         "--resource-type", related_type, "--unconditional", "--unconditional-source", "--json"]),
                         {"ambiguous_link"}, "ambiguous pair")
    require(ambiguity.get("candidateIDs") == [first["id"], second["id"]],
            "ambiguity did not return sorted exact candidate IDs")
    for path, before in [(source_path, source), (target_path, target), (issue_path, issue),
                         (first_path, first), (second_path, second), (blocking[0]["id"], blocking[0])]:
        require(show("lifecycle-refusals-unchanged", path) == before, f"unlink refusals changed {path}")
    capture.passed("unlink requires separate Link/source guards; stale/conflicting guards, blocking adapter and absent/ambiguous pair refuse without current-state changes")

    yaml_path = capture.work / ".beads" / "config.yaml"
    original_yaml = yaml_path.read_bytes() if yaml_path.exists() else None
    mayor, freeze_path = capture.work / "mayor", capture.work / "MIGRATION-FREEZE"
    require(not mayor.exists() and not freeze_path.exists(), "unexpected existing town markers")
    for policy in ["readonly-flag", "readonly-env", "readonly-config", "migration-freeze"]:
        flags = ["--readonly"] if policy == "readonly-flag" else []
        try:
            if policy == "readonly-env":
                capture.env["BD_READONLY"] = "true"
            elif policy == "readonly-config":
                yaml_path.write_text("readonly: true\n")
            elif policy == "migration-freeze":
                mayor.mkdir()
                (mayor / "town.json").write_text("{}\n")
                freeze_path.write_text("lifecycle-smoke\t2026-09-25T00:00:00Z\tgraph write control\n")
            before = tree_digest(capture.work)
            label = "lifecycle-policy-" + policy
            refusal(capture.run(label, ["unlink", *guarded, *flags, "--json"]), {"permission_denied"}, label)
            require(tree_digest(capture.work) == before, f"{label}: changed workspace bytes")
        finally:
            capture.env.pop("BD_READONLY", None)
            if original_yaml is None:
                yaml_path.unlink(missing_ok=True)
            else:
                yaml_path.write_bytes(original_yaml)
            if policy == "migration-freeze":
                freeze_path.unlink(missing_ok=True)
                (mayor / "town.json").unlink(missing_ok=True)
                mayor.rmdir()
    require(show("lifecycle-policy-source-unchanged", source_path) == source and
            show("lifecycle-policy-Link-unchanged", first_path) == first, "policy-refused unlink changed records")
    capture.passed("readonly flag/environment/config and migration freeze block unlink before workspace changes")

    deleted = unlink("lifecycle-ID-unlink", [SCOPE + first_path, *guarded[1:]], first, source, True)
    owner = deleted["source"]
    require(show("lifecycle-surviving-multiedge", second_path) == second, "ID unlink changed another same-endpoint Link")
    require(links("lifecycle-live-only-after-ID", source_path) == [second], "deleted Link remained in incident listing")
    require(show("lifecycle-unlink-target-unchanged", target_path) == target, "owned unlink changed distinct target")
    refusal(capture.run("lifecycle-repeat-unlink", ["unlink", first_path, "--unconditional", "--unconditional-source", "--json"]),
            {"gone"}, "repeated unlink")
    refusal(capture.run("lifecycle-deleted-ID-reserved", ["link", source_path, target_path,
                       "--resource-type", related_type, "--id", first_path, "--properties", "{}",
                       "--if-source-revision", owner["revision"], "--json"]), {"identity_reserved"}, "deleted identity reuse")
    require(show("lifecycle-repeat-source-unchanged", source_path) == owner,
            "repeated unlink or identity reuse advanced owning source again")
    capture.passed("ID unlink retains a versioned tombstone, advances owning Memory once, preserves target/other multiedge; repeat is gone and identity stays reserved")

    pair_deleted = unlink("lifecycle-single-pair-unlink", [SCOPE + source_path, target_path, "--resource-type", related_type,
                          "--if-revision", second["revision"], "--if-source-revision", owner["revision"]], second, owner, True)
    require(links("lifecycle-source-empty", source_path) == [], "single-pair unlink left a live Link")
    unowned_deleted = unlink("lifecycle-unowned-unlink", ["links/issue-context", "--unconditional"], issue_link, issue, False)
    require(links("lifecycle-issue-blocking-survives", issue_path) == blocking, "unowned unlink changed blocking Dependency")
    require(show("lifecycle-final-distinct-target", target_path) == target, "unlink sequence changed target")
    capture.passed("single exact pair unlinks after ambiguity is resolved; unowned Issue Link deletion preserves Issue and specialized Dependency")

    self_created = envelope(capture.success("lifecycle-self-create", ["link", target_path, target_path,
                            "--resource-type", related_type, "--id", "links/self-context", "--properties", "{}",
                            "--if-source-revision", target["revision"], "--json"]))
    self_link, self_owner = self_created["link"], self_created["source"]
    require(self_owner.get("owned") == [self_link] and self_owner["version"] != target["version"],
            "self-Link creation did not version its single owning Memory")
    for direction in ["in", "out", "both"]:
        require(links("lifecycle-self-" + direction, target_path, direction) == [self_link],
                f"self-Link was missing or duplicated in {direction} listing")
    self_deleted = unlink("lifecycle-self-unlink", ["links/self-context", "--unconditional", "--unconditional-source"],
                          self_link, self_owner, True)
    require(links("lifecycle-self-gone", target_path) == [], "deleted self-Link remained live")
    capture.passed("self-Link appears once in in/out/both listing and unlink removes it from its owning Memory")
    write_json(capture.output / "link-lifecycle-versions.json", {
        "memory_before": source, "first_unlink": deleted, "second_unlink": pair_deleted,
        "unowned_unlink": unowned_deleted, "self_unlink": self_deleted,
        "distinct_target_unchanged_during_unlink": target,
        "limitation": "current reads, tombstone receipts and opaque versions only; storage tests separately establish retained-state rollback/concurrency; no public History or restore claim",
    })


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--bd", required=True, type=Path, help="absolute installed bd executable")
    parser.add_argument("--output-dir", "--outputdir", required=True, type=Path, help="new artifact directory")
    parser.add_argument("--server-port", type=int, help="caller-owned ordinary loopback Dolt server")
    parser.add_argument("--server-root", type=Path, help="existing disposable server data root (provenance only)")
    parser.add_argument("--command-timeout", type=float, default=120)
    parser.add_argument("--total-timeout", type=float, default=900)
    parser.add_argument("--dependency-workflow", action="store_true",
                        help="also exercise installed blocking Dependency ownership, close and ready workflow")
    parser.add_argument("--mixed-links", action="store_true",
                        help="also exercise installed informational Links and property updates; implies --dependency-workflow")
    parser.add_argument("--link-lifecycle", action="store_true",
                        help="also exercise incident listing and guarded informational unlink; implies --mixed-links")
    args = parser.parse_args()
    args.mixed_links = args.mixed_links or args.link_lifecycle
    args.dependency_workflow = args.dependency_workflow or args.mixed_links
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
        if args.dependency_workflow:
            exercise_dependency_workflow(capture)
        if args.mixed_links:
            exercise_mixed_links(capture)
        if args.link_lifecycle:
            exercise_link_lifecycle(capture)
        require(sha256(args.bd) == capture.binary_hash, "binary changed during capture")
    except BaseException as exc:
        failure = f"{type(exc).__name__}: {exc}"
    finally:
        capture.stop()
        write_json(capture.output / "summary.json", {
            "passed": failure is None, "c0_qualified": False, "failure": failure, "checks": capture.passes,
            "qualification_gaps": capture.qualification_gaps,
            "commands": capture.records, "private_root": str(capture.root),
            "dependency_workflow_requested": args.dependency_workflow,
            "mixed_links_requested": args.mixed_links,
            "link_lifecycle_requested": args.link_lifecycle,
            "active_child_count": len(capture.active),
            "limits": ["this smoke harness alone cannot qualify C0",
                       "workspace digest and absent-ID reads do not prove unchanged remote database or atomic retained history/effects",
                       "process launch barrier does not prove internal transaction overlap",
                       "no cancellation, crash, network-loss or uncertain-commit injection",
                       "no Linux/macOS platform-matrix qualification",
                       "no historical-read, migration, or performance claim",
                       "Dependency and mixed-Link current-read checks do not establish complete retained-history or rollback atomicity",
                       "optional Link lifecycle covers bounded incident listing and informational unlink; no blocking unlink, pagination, target pin, metadata, restore or request replay claim"],
        })
    print(f"{'PASS' if failure is None else 'FAIL'} capture: {capture.output}", flush=True)
    if failure:
        print(failure, flush=True)
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
