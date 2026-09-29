#!/usr/bin/env python3
"""Required graph/BDP Read proof: real engines, exact CLI and public clients.

Includes the C0 captures and every discovered graph test, including the mixed
Issue/Memory/Link installed workflow, HTTP security and authenticated independent
Node/Python BDP reads. This is not full Memory, History or HTTP Write proof.
The caller supplies the ordinary released Dolt binary and CI Build Artifacts.
All databases and process groups belong to this run; no existing server is used.
"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import signal
import socket
import subprocess
import sys
import time

OWNER_CLEANUP_GRACE = 30


def require(condition, message):
    if not condition:
        raise RuntimeError(message)


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def group_exited(pid, timeout=5):
    # A command may exit just before its already-exiting child is reaped.
    # Allow a bounded grace period, but never count forced cleanup as success.
    deadline = time.monotonic() + timeout
    while True:
        try:
            os.killpg(pid, 0)
        except ProcessLookupError:
            return True
        if time.monotonic() >= deadline:
            return False
        time.sleep(0.05)


def verify_tests(events, expected):
    runs, passes = set(), set()
    for event in events:
        name, action = event.get("Test"), event.get("Action")
        require(action not in ("fail", "skip"), f"required test did not pass: {event}")
        if name:
            require(name.split("/", 1)[0] in expected, f"unexpected test or subtest: {name}")
            if action == "run":
                require(name not in runs, f"duplicate test run: {name}")
                runs.add(name)
            elif action == "pass":
                require(name in runs and name not in passes, f"unmatched test pass: {name}")
                passes.add(name)
    roots = {name for name in runs if "/" not in name}
    require(expected and roots == expected and runs == passes,
            f"test discovery/execution mismatch: expected roots={sorted(expected)}, runs={sorted(runs)}, passes={sorted(passes)}")
    require(any(e.get("Action") == "pass" and not e.get("Test") for e in events),
            "missing successful package completion")
    return len(roots)


def cleanup_registered_groups(path):
    """Fallback for an interrupted/failed owner, including dead group leaders."""
    active = set()
    if path is not None and path.exists():
        for line in path.read_text().splitlines():
            event = json.loads(line)
            require(type(event.get("pid")) is int and event["pid"] > 1, "invalid owned process identity")
            require(event.get("action") in ("start", "end"), "invalid process registry event")
            if event["action"] == "start":
                active.add(event["pid"])
            else:
                active.discard(event["pid"])
    forced = []
    for pid in active:
        try:
            os.killpg(pid, signal.SIGTERM)
        except ProcessLookupError:
            pass
    for pid in active:
        if not group_exited(pid, timeout=5):
            forced.append(pid)
            try:
                os.killpg(pid, signal.SIGKILL)
            except ProcessLookupError:
                pass
        require(group_exited(pid, timeout=5), f"owned process group survived cleanup: {pid}")
    return dict(registeredGroups=len(active), forcedGroups=forced, groupsGone=True)


def verify_clean_receipt(receipt, listener=False):
    require(receipt.get("exit_code") == 0 and receipt.get("failure") is None
            and receipt.get("forced") is False and receipt.get("group_gone") is True,
            "HTTP capture process did not finish cleanly")
    if listener:
        require(receipt.get("port_closed") is True, "HTTP listener remained open")


def verify_http_capture(root, binary_hash):
    summary = json.loads((root / "summary.json").read_text())
    require(summary.get("passed") is True and summary.get("failure") is None
            and summary.get("active_children") == 0 and summary.get("cli_commands") == 16,
            "incomplete installed HTTP capture")
    require(summary.get("client_pin") == "53bdbd03136875f952af184fce7b3c7af8f74e96"
            and summary.get("installed_binary_sha256") == binary_hash, "HTTP source/binary provenance mismatch")
    require(summary.get("python") == {"passed": True, "beads": 4, "limit": 1,
            "authenticated": True, "mechanism": "unchanged standard-library example over BDP HTTP"},
            "Python BDP consumer proof missing")
    require(summary.get("python_pages") == 4, "Python continuation observations missing")
    processes = list(root.glob("*/receipt.json"))
    require(len(processes) == 21, "missing seed/delete/serve/client/Python process receipt")
    for path in processes:
        receipt = json.loads(path.read_text())
        verify_clean_receipt(receipt, listener=path.parent.name == "serve")
        for stream in ("stdout", "stderr"):
            require(digest(path.parent / (stream + ".log")) == receipt[stream + "_sha256"],
                    "HTTP process output hash differs")
    client = json.loads((root / "client-results.json").read_text())
    require(client.get("passed") is True and client.get("failure") is None
            and len(client.get("checks", [])) == 12, "incomplete public-client checks")
    network = json.loads((root / "client-network.json").read_text())
    require(network and client["requests"] == len(network), "HTTP network receipts incomplete")
    for observation in network:
        require(observation["url"].startswith(summary["scope"])
                and re.fullmatch(r"[0-9a-f]{64}", observation["bodySha256"]), "invalid HTTP observation")
    verify_issue_append_capture(root, summary, binary_hash, network)
    verify_memory_deletion_capture(root, summary)


def verify_issue_append_capture(root, summary, binary_hash, network):
    """Keep the real writer receipts, public records and raw HTTP observations linked."""
    def data(relative):
        return json.loads((root / relative).read_text())
    artifacts = data("client-artifacts.json")
    notes = artifacts["issueNotes"]
    scope, actor = summary["scope"], "bdp-read-note-holder"
    work_id = scope + "beads/work"
    before, claimed, after = notes["before"], notes["claimed"], notes["after"]
    require(before == artifacts["issueEdit"]["after"], "append preimage lost preceding Issue-edit proof")
    require(all(item["id"] == work_id and item["type"] == before["type"] for item in (claimed, after)),
            "append changed canonical Issue identity/Type")
    expected_claim = dict(before["properties"])
    for key in ("status", "assignee", "started_at", "updated_at", "lease_expires_at", "heartbeat_at", "lease_granted_node"):
        if key in claimed["properties"]:
            expected_claim[key] = claimed["properties"][key]
        else:
            expected_claim.pop(key, None)
    require(claimed["properties"] == expected_claim and claimed["properties"]["status"] == "in_progress"
            and claimed["properties"]["assignee"] == actor and claimed["properties"].get("notes", "") == ""
            and claimed["revision"] != before["revision"], "incomplete installed claim precondition")
    require(notes["noopRecord"] == claimed and notes["noopMutation"]["result"]["changed"] is False
            and notes["noopMutation"]["result"]["issue"] == notes["claimMutation"]["result"]["issue"],
            "empty append was not a complete semantic no-op")
    text = "  Progress — 雪\r\nsecond line\t  "
    require(notes["appendText"] == text and after["properties"] == dict(claimed["properties"],
            notes=text, updated_at=after["properties"]["updated_at"]), "appended properties/lease changed or bytes lost")
    require(after["revision"] != claimed["revision"] and after["attribution"]["principal"] == actor
            and after["ownedLinks"] == claimed["ownedLinks"] == before["ownedLinks"], "append ownership/revision proof missing")
    require(notes["properties"] == after["properties"] and notes["inventory"]["next"] is None
            and len(notes["inventory"]["items"]) == 4
            and [item for item in notes["inventory"]["items"] if item["id"] == work_id] == [after]
            and notes["incident"] == artifacts["issueEdit"]["incidentAfter"], "append BDP views disagree")
    require(notes["etagBefore"] and notes["etagAfter"] and notes["etagBefore"] != notes["etagAfter"],
            "append did not change ETag")
    binary = data("02-memory-create/receipt.json")["argv"][0]
    commands = (
        ("client-cli-issue-claim", "claimMutation", ["update", work_id, "--claim", "--actor", actor, "--json"], True, claimed),
        ("client-cli-issue-append-noop", "noopMutation", ["update", work_id, "--append-notes=", "--if-revision", claimed["revision"], "--actor", actor, "--json"], False, claimed),
        ("client-cli-issue-append", "appendMutation", ["update", work_id, "--append-notes", text, "--if-revision", claimed["revision"], "--actor", actor, "--json"], True, after),
    )
    for name, key, args, changed, record in commands:
        receipt, mutation = data(name + ".json"), data(name + ".stdout.log")
        require(receipt["argv"] == [binary, *args] and receipt["binarySha256"] == binary_hash
                and receipt["exitCode"] == 0 and receipt["signal"] is None, "append CLI source/exit/guard differs")
        for stream in ("stdout", "stderr"):
            require(digest(root / (name + "." + stream + ".log")) == receipt[stream + "Sha256"],
                    "append CLI output hash differs")
        require(mutation == notes[key] and mutation["schemaVersion"] == 1 and mutation["preview"] is True
                and mutation["result"]["changed"] is changed and mutation["result"]["issue"]["id"] == work_id
                and mutation["result"]["issue"]["revision"] == record["revision"], "append CLI/public record mismatch")
    for label, index_key, status, etag, record in (
        ("before", "claimNetworkIndex", 200, notes["etagBefore"], claimed),
        ("noop", "noopNetworkIndex", 304, notes["etagBefore"], None),
        ("after", "appendNetworkIndex", 200, notes["etagAfter"], after),
    ):
        index = notes[index_key]
        require(type(index) is int and 0 <= index < len(network), "append HTTP observation index missing")
        observation = network[index]
        body = root / ("client-issue-notes-" + label + ".body")
        require(observation["url"] == work_id and observation["method"] == "GET" and observation["status"] == status
                and observation["headers"].get("etag") == etag and observation["bodyBytes"] == body.stat().st_size
                and observation["bodySha256"] == digest(body), "append HTTP status/ETag/raw bytes disagree")
        require((body.read_bytes() == b"") if record is None else json.loads(body.read_bytes()) == record,
                "append raw HTTP body differs from public record")
    require(notes["claimNetworkIndex"] < notes["noopNetworkIndex"] < notes["appendNetworkIndex"],
            "append HTTP observations are not ordered")


def verify_memory_deletion_capture(root, summary):
    """Check the new proof without overwriting any pre-deletion evidence."""
    def data(relative):
        return json.loads((root / relative).read_text())
    alpha = data("02-memory-create/stdout.log")["result"]
    scope = summary["scope"]
    require(alpha["id"] == scope + "beads/alpha", "wrong deletion fixture identity")
    deletion = data("16-memory-delete/stdout.log")
    require(deletion.get("preview") is True and deletion.get("schemaVersion") == 1
            and deletion.get("result") == {"memory": alpha, "preview": False, "deleted": True},
            "deletion did not retain/disclose exact final live state")
    argv = data("16-memory-delete/receipt.json")["argv"]
    require(argv == [data("02-memory-create/receipt.json")["argv"][0], "delete", "beads/alpha", "--force",
                     "--if-revision", alpha["revision"], "--json"], "deletion did not use original observed guard")
    require(summary.get("memory_delete") == {"passed": True, "id": alpha["id"], "final_live_revision": alpha["revision"]},
            "Memory deletion summary missing")
    require(summary.get("python_after_delete") == {"passed": True, "beads": 3, "limit": 1,
            "authenticated": True, "mechanism": "unchanged standard-library example over BDP HTTP"}
            and summary.get("python_after_delete_pages") == 3, "post-delete Python summary missing")
    observations = data("python-after-delete-network.json")
    require(len(observations) == 4 and observations[0]["url"] == scope + "bdp.json"
            and observations[0]["document"]["scope"] == scope
            and observations[0]["document"]["profile"] == "read", "post-delete discovery missing")
    next_url, records = scope + "beads/?limit=1", []
    for observation in observations:
        require(observation["method"] == "GET" and observation["status"] == 200
                and observation["url"].startswith(scope) and observation["body_bytes"] > 0
                and re.fullmatch(r"[0-9a-f]{64}", observation["body_sha256"]), "invalid post-delete HTTP observation")
    for observation in observations[1:]:
        page = observation["document"]
        require(observation["url"] == next_url and len(page["items"]) == 1,
                "post-delete Python did not follow actual one-record next page")
        records.extend(page["items"])
        next_url = page["next"]
    require(next_url is None and len(records) == 3 and {item["id"] for item in records} ==
            {scope + "beads/" + name for name in ["plan", "work", "prereq"]}
            and data("python-after-delete/stdout.log") == records,
            "post-delete Python enumeration retained deleted state or omitted survivors")


class Qualification:
    def __init__(self, args):
        self.root = Path(__file__).resolve().parents[2]
        self.output = Path(args.output).resolve()
        self.output.mkdir(parents=True, exist_ok=False)
        self.bd = Path(args.bd).resolve()
        self.dolt = Path(args.dolt).resolve()
        self.artifacts = Path(args.artifacts).resolve()
        self.children = set()
        self.receipts = []
        self.counts = {}
        self.client_checkout = Path(args.bdp_checkout).resolve()
        self.client_manifest = Path(args.client_manifest).resolve()
        self.node = Path(args.node).resolve()
        self.env = {k: os.environ[k] for k in (
            "PATH", "GOCACHE", "GOMODCACHE", "GOROOT", "DEVELOPER_DIR", "TMPDIR") if k in os.environ}
        home = self.output / "home"
        home.mkdir()
        for key in ("GOCACHE", "GOMODCACHE"):
            if key not in self.env:
                self.env[key] = subprocess.check_output(["go", "env", key], cwd=self.root, text=True).strip()
        self.env.update(HOME=str(home), XDG_CONFIG_HOME=str(home / ".config"),
                        GIT_CONFIG_NOSYSTEM="1", GIT_CONFIG_GLOBAL=str(home / "empty-gitconfig"),
                        CGO_ENABLED="1", PYTHONDONTWRITEBYTECODE="1", BD_DISABLE_METRICS="1", BD_DISABLE_EVENT_FLUSH="1",
                        DOLT_METRICS_DISABLED="1", DOLT_DISABLE_EVENT_FLUSH="1",
                        BEADS_DOLT_AUTO_START="0", NO_COLOR="1",
                        BEADS_TEST_BD_BINARY=str(self.bd), BEADS_TEST_IGNORE_REPO_CONFIG="1")

    def run(self, args, label, cwd=None, timeout=120, stdin=None, expected=0, owned_groups=None):
        print(f"C0 {label}: {args}", flush=True)
        process = subprocess.Popen([str(a) for a in args], cwd=cwd or self.root,
                                   env=self.env, stdin=subprocess.PIPE, stdout=subprocess.PIPE,
                                   stderr=subprocess.PIPE, start_new_session=True)
        self.children.add(process.pid)
        try:
            out, err = process.communicate(stdin, timeout=timeout)
        except BaseException as original:
            # The smoke process owns independent groups. Give its signal
            # handler a bounded chance to drain them before killing the owner.
            try:
                process.send_signal(signal.SIGTERM)
            except ProcessLookupError:
                pass
            forced_owner = False
            cleanup_errors = []
            out, err = b"", b""
            nested = None
            try:
                out, err = process.communicate(timeout=OWNER_CLEANUP_GRACE)
            except BaseException as drain_error:
                forced_owner = True
                if not isinstance(drain_error, subprocess.TimeoutExpired):
                    cleanup_errors.append(f"owner drain: {drain_error!r}")
                # Keep the owner alive while its children stop, so it can reap
                # them. A registry failure must never leave the owner alive.
                try:
                    cleanup_registered_groups(owned_groups)
                except BaseException as cleanup_error:
                    cleanup_errors.append(f"nested before owner kill: {cleanup_error!r}")
                finally:
                    try:
                        os.killpg(process.pid, signal.SIGKILL)
                    except ProcessLookupError:
                        pass
                try:
                    out, err = process.communicate(timeout=10)
                except BaseException as reap_error:
                    cleanup_errors.append(f"owner reap: {reap_error!r}")
                    if isinstance(reap_error, subprocess.TimeoutExpired):
                        out, err = reap_error.output or b"", reap_error.stderr or b""
            if not group_exited(process.pid):
                forced_owner = True
                try:
                    os.killpg(process.pid, signal.SIGKILL)
                except ProcessLookupError:
                    pass
                if not group_exited(process.pid):
                    cleanup_errors.append("owner process group survived forced cleanup")
            # Retry after the owner has retired, even if the first registry
            # pass failed. Preserve the triggering failure and every cleanup
            # failure in the receipt instead of replacing the original error.
            try:
                nested = cleanup_registered_groups(owned_groups)
            except BaseException as cleanup_error:
                cleanup_errors.append(f"nested after owner retirement: {cleanup_error!r}")
            (self.output / f"{label}.stdout").write_bytes(out)
            (self.output / f"{label}.stderr").write_bytes(err)
            self.receipts.append(dict(label=label, argv=[str(a) for a in args], pid=process.pid,
                                      exit=process.returncode, failure=type(original).__name__,
                                      forcedOwner=forced_owner, nestedCleanup=nested,
                                      cleanupErrors=cleanup_errors,
                                      stdoutSHA256=hashlib.sha256(out).hexdigest()))
            raise
        finally:
            self.children.discard(process.pid)
        (self.output / f"{label}.stdout").write_bytes(out)
        (self.output / f"{label}.stderr").write_bytes(err)
        if not group_exited(process.pid):
            os.killpg(process.pid, signal.SIGKILL)
            raise RuntimeError(f"{label} left a live process group after exit")
        self.receipts.append(dict(label=label, argv=[str(a) for a in args], pid=process.pid,
                                  exit=process.returncode, stdoutSHA256=hashlib.sha256(out).hexdigest()))
        if owned_groups is not None:
            nested = cleanup_registered_groups(owned_groups)
            require(nested["registeredGroups"] == 0, "completed owner left nested groups registered")
        require(process.returncode == expected,
                f"{label} exited {process.returncode}, expected {expected}; see saved stdout/stderr")
        return out

    def tests(self, package, selector, files, label, required_subtests=()):
        # Source discovery also sees accidentally build-excluded added tests.
        expected = set()
        for path in files:
            names = re.findall(r"^func (Test\w+)\(t \*testing.T\)", path.read_text(), re.MULTILINE)
            expected.update(name for name in names if re.search(selector, name))
        listing = self.run(["go", "test", "-tags", "gms_pure_go", "-list", selector, package],
                           label + "-discovery", timeout=600).decode()
        listed = set(re.findall(r"^Test\w+$", listing, re.MULTILINE))
        require(listed == expected and expected, f"{label}: source and compiled test discovery disagree")
        out = self.run(["go", "test", "-tags", "gms_pure_go", "-json", "-count=1", "-p=1",
                        "-parallel=1", "-timeout=15m", "-run", selector, package], label, timeout=960)
        events = [json.loads(line) for line in out.splitlines()]
        self.counts[label] = verify_tests(events, expected)
        completed = {e.get("Test") for e in events if e.get("Action") == "pass"}
        require(set(required_subtests) <= completed, f"{label}: required engine proof absent")

    def cli(self, work, label, *args):
        return json.loads(self.run([self.bd, *args, "--json"], label, cwd=work))

    def capture(self, engine, port):
        work = self.output / (engine + "-workspace")
        work.mkdir()
        scope = f"https://example.invalid/ci-c0/{engine}/"
        self.run(["git", "init", "-q", str(work)], engine + "-git-init")
        self.run(["git", "-C", str(work), "config", "core.hooksPath", ".git/hooks"], engine + "-git-hooks")
        flags = ["--server", "--external", "--server-host", "127.0.0.1", "--server-port", str(port), "--server-user", "root"] if engine == "server" else []
        initialized = self.cli(work, engine + "-init", "init", "--graph-mode", "link", "--scope-url", scope,
                               "--skip-hooks", "--skip-agents", "--non-interactive", *flags)
        require(initialized["result"]["backend"] == engine and initialized["result"]["scope"] == scope,
                "initialization selected the wrong backend or Scope")
        created = self.cli(work, engine + "-create", "remember", "Durable C0 body — 記憶", "--id", "beads/plan", "--title", "Plan")
        record = created["result"]
        require(record["id"] == scope + "beads/plan" and record["type"] == scope + "types/preview-memory-v2"
                and record["properties"] == {"title": "Plan", "body": "Durable C0 body — 記憶"}
                and record["owned"] == [] and record["revision"] and record["version"], "incomplete/non-Memory creation")
        # Each call starts and reaps a distinct installed CLI process.
        for suffix, identity in [("show", "beads/plan"), ("reopen", record["id"])]:
            require(self.cli(work, engine + "-" + suffix, "show", identity) == created,
                    "fresh-process read differs from committed creation")
        status = self.cli(work, engine + "-status", "status", "--graph")
        require(status["result"]["capabilities"]["memoryRead"] is True, "Memory read not advertised")

    def execute(self):
        require(self.bd.is_file() and os.access(self.bd, os.X_OK), "missing installed CLI")
        require(self.dolt.is_file() and os.access(self.dolt, os.X_OK), "missing released Dolt")
        head = self.run(["git", "rev-parse", "HEAD"], "head").decode().strip()
        require(not self.run(["git", "status", "--porcelain", "--untracked-files=all"], "source-status"),
                "qualification requires clean source, including no untracked files")
        manifest = dict(line.split("=", 1) for line in (self.artifacts / "build-manifest.txt").read_text().splitlines())
        require(manifest["commit"] == head and manifest["build_tags"] == "gms_pure_go", "CLI artifact source/build flags mismatch")
        expected_hash = (self.artifacts / "SHA256SUMS").read_text().split()[0]
        require(digest(self.bd) == expected_hash, "installed CLI hash differs from build artifact")
        version = self.run([self.dolt, "version"], "dolt-version").decode().splitlines()
        require(version and version[0].strip() == "dolt version 2.1.8", "C0 requires released Dolt2.1.8")
        (self.output / "source.json").write_text(json.dumps(dict(commit=head, binarySHA256=digest(self.bd),
            doltSHA256=digest(self.dolt), runnerSHA256=digest(Path(__file__))), indent=2) + "\n")
        with socket.socket() as reservation:
            reservation.bind(("127.0.0.1", 0))
            port = reservation.getsockname()[1]
        server_root = self.output / "server-data"
        server_root.mkdir()
        log = (self.output / "server.log").open("wb")
        server = subprocess.Popen([str(self.dolt), "sql-server", "--host", "127.0.0.1", "--port", str(port), "--data-dir", str(server_root)],
                                  cwd=server_root, env=self.env, stdout=log, stderr=subprocess.STDOUT, start_new_session=True)
        self.children.add(server.pid)
        try:
            deadline = time.monotonic() + 30
            while True:
                require(server.poll() is None, "owned server exited before readiness")
                try:
                    with socket.create_connection(("127.0.0.1", port), timeout=1):
                        break
                except OSError:
                    require(time.monotonic() < deadline, "server readiness timed out")
                    time.sleep(0.2)
            self.env["BEADS_GRAPH_TEST_SERVER_PORT"] = str(port)
            # Sequential package runs/provisioning; concurrency inside same-store tests remains exercised.
            self.tests("./internal/graphpatch", "^Test", (self.root / "internal/graphpatch").glob("*_test.go"), "graphpatch")
            self.tests("./internal/storage/graphstore", "^Test", (self.root / "internal/storage/graphstore").glob("*_test.go"), "storage")
            self.tests("./internal/configfile", "^TestGraphMode", (self.root / "internal/configfile").glob("graph_mode_test.go"), "config")
            self.tests("./internal/storage/issueops", "^TestResolve(CustomConfigStrict|InfraTypesStrict|ConfigLegacy)",
                       (self.root / "internal/storage/issueops").glob("config_strict_test.go"), "query-config")
            self.tests("./cmd/bd", "^Test(GraphModeCLI|GraphPreview)", (self.root / "cmd/bd").glob("graph*test.go"), "cli",
                       ("TestGraphPreviewIssueAppendWorkflow/embedded", "TestGraphPreviewIssueAppendWorkflow/server",
                        "TestGraphPreviewIssueClaimWorkflow/embedded", "TestGraphPreviewIssueClaimWorkflow/server",
                        "TestGraphPreviewIssueAssignmentWorkflow/embedded", "TestGraphPreviewIssueAssignmentWorkflow/server",
                        "TestGraphPreviewPropertiesPatchWorkflow/embedded", "TestGraphPreviewPropertiesPatchWorkflow/server"))
            self.env["BDP_SPEC_AT_PIN"] = str(self.client_checkout / "docs/specs/bdp.md")
            self.tests("./internal/httpapi/bdpwire", "^Test", (self.root / "internal/httpapi/bdpwire").glob("*_test.go"), "bdpwire")
            self.tests("./internal/httpapi/graphread", "^Test", (self.root / "internal/httpapi/graphread").glob("*_test.go"), "graphread",
                       ("TestAuthoritativeRecordsProjectToPublicWire/embedded", "TestAuthoritativeRecordsProjectToPublicWire/server"))
            self.tests("./internal/httpapi", "^TestGraphRead", (self.root / "internal/httpapi").glob("graph_read*test.go"), "graph-http",
                       ("TestGraphReadHTTPAuthorityAndSecurity/issue-priority-assignment",
                        "TestGraphReadHTTPAuthorityAndSecurity/issue-append-notes",))
            self.run([sys.executable, "-m", "unittest", "discover", "-s", "examples/bdp-read", "-p", "test_*.py"], "python-example-tests")
            python_roots = sum(len(re.findall(r"^    def test_\w+\(", path.read_text(), re.MULTILINE))
                               for path in (self.root / "examples/bdp-read").glob("test_*.py"))
            python_result = (self.output / "python-example-tests.stderr").read_text()
            require(python_roots > 0 and re.search(rf"Ran {python_roots} tests? in", python_result)
                    and "skipped=" not in python_result, "Python example tests missing or skipped")
            self.capture("embedded", port)
            self.capture("server", port)
            self.run([sys.executable, str(self.root / "scripts/graph-bdp-read-smoke.py"),
                      "--bd", str(self.bd), "--server-port", str(port),
                      "--bdp-checkout", str(self.client_checkout), "--client-manifest", str(self.client_manifest),
                      "--node", str(self.node), "--output-dir", str(self.output / "http"),
                      "--owned-groups", str(self.output / "http-owned-groups.jsonl")], "bdp-http-capture", timeout=360,
                     owned_groups=self.output / "http-owned-groups.jsonl")
            verify_http_capture(self.output / "http", digest(self.bd))
            http_summary = json.loads((self.output / "http/summary.json").read_text())
            for field, relative in [("harness_sha256", "scripts/graph-bdp-read-smoke.py"),
                                    ("client_harness_sha256", "scripts/graph-bdp-read-client.mjs"),
                                    ("python_example_sha256", "examples/bdp-read/read_beads.py")]:
                require(http_summary[field] == digest(self.root / relative), "HTTP capture source changed")
            require(server.poll() is None, "server exited during qualification")
        finally:
            server.terminate()
            forced = False
            try:
                code = server.wait(timeout=15)
            except subprocess.TimeoutExpired:
                forced = True
                os.killpg(server.pid, signal.SIGKILL)
                code = server.wait()
            finally:
                self.children.discard(server.pid)
                log.close()
            group_gone = group_exited(server.pid)
            if not group_gone:
                os.killpg(server.pid, signal.SIGKILL)
            port_closed = False
            try:
                with socket.create_connection(("127.0.0.1", port), timeout=1):
                    pass
            except ConnectionRefusedError:
                port_closed = True
            self.receipts.append(dict(label="server-shutdown", pid=server.pid, port=port,
                                      exit=code, forced=forced, groupGone=group_gone,
                                      portClosed=port_closed))
            require(code == 0 and not forced and group_gone and port_closed,
                    f"server cleanup failed: exit={code}, forced={forced}, groupGone={group_gone}, portClosed={port_closed}")
        require(not self.children, "owned child process remains")
        (self.output / "summary.json").write_text(json.dumps(dict(commit=head, passed=self.counts,
            installedCLICommands=10, engines=["embedded", "server"], childrenRemaining=0, bdpHTTP=True, pythonBDP=True), indent=2) + "\n")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    for name in ("bd", "dolt", "artifacts", "output", "bdp-checkout", "client-manifest", "node"):
        parser.add_argument("--" + name, required=True)
    def interrupted(signum, _frame):
        raise RuntimeError(f"qualification interrupted by signal {signum}")

    signal.signal(signal.SIGTERM, interrupted)
    signal.signal(signal.SIGINT, interrupted)
    qualification = Qualification(parser.parse_args())
    try:
        qualification.execute()
    finally:
        (qualification.output / "receipts.json").write_text(json.dumps(qualification.receipts, indent=2) + "\n")


if __name__ == "__main__":
    main()
