#!/usr/bin/env python3
"""Required graph/BDP Read proof: real engines, exact CLI and public clients.

Includes the C0 captures and every discovered graph test, including the mixed
Issue/Memory/Link installed workflow, HTTP security and authenticated independent
Node/Python BDP reads. This is not full Memory, History or HTTP Write proof.
The caller supplies the ordinary released Dolt binary and CI Build Artifacts.
All databases and process groups belong to this run; no existing server is used.
"""
import argparse
import contextlib
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
GO_TEST_TIMEOUT = 900  # seconds; the -timeout=15m every go test group is given
SLOW_GROUP_SECONDS = GO_TEST_TIMEOUT * 6 // 10
TAIL_LINES = 40
SERVER_READY_TIMEOUT = 30
SERVER_STOP_TIMEOUT = 15
# Deadlines for the HTTP/BDP smoke, in seconds. Each outlasts the one before it: one
# command < the smoke's whole run < the wrapper that kills it from outside, so a slow runner
# ends in the smoke's own labelled failure and never in an unlabelled outer kill. The
# smoke's first command is a server-backed bd init: 35 s of a 45 s smoke on a fast runner,
# and 34 s, 58 s and 65 s for the qualification's own init on fast and slow runners. One
# command gets 180 s, about 2.8x the slowest of those. The total covers a full-length init
# plus the rest of the smoke at the same slowdown (about 230 s) with room to spare.
SMOKE_COMMAND_TIMEOUT = 180
SMOKE_TOTAL_TIMEOUT = 600
SMOKE_WRAPPER_TIMEOUT = 660  # total + 60 s for the smoke to reap its children and report


def require(condition, message):
    if not condition:
        raise RuntimeError(message)


def require_memory_creation(record, scope):
    require(record.get("id") == scope + "beads/plan"
            and record.get("type") == scope + "types/preview-memory-v2"
            and record.get("properties") == {"title": "Plan", "body": "Durable C0 body — 記憶"}
            and record.get("owned") == []
            and isinstance(record.get("revision"), str) and bool(record["revision"])
            and "version" not in record, "incomplete/non-Memory creation")


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


def partition_required_tests(names, groups):
    require(names and len(names) == len(set(names)), "empty or duplicate required test roots")
    require(type(groups) is int and 1 < groups <= len(names), "required groups must be nonempty")
    ordered = sorted(names)
    selections = [ordered[index::groups] for index in range(groups)]
    union = [name for group in selections for name in group]
    require(len(union) == len(set(union)) and set(union) == set(names), "incomplete required test partition")
    return selections


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
    require(summary.get("client_pin") == "de99030d13a57e77f4f2660b91ea3c39b82823cb"
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
            and len(client.get("checks", [])) == 13, "incomplete public-client checks")
    network = json.loads((root / "client-network.json").read_text())
    require(network and client["requests"] == len(network), "HTTP network receipts incomplete")
    for observation in network:
        require(observation["url"].startswith(summary["scope"])
                and re.fullmatch(r"[0-9a-f]{64}", observation["bodySha256"]), "invalid HTTP observation")
    verify_issue_append_capture(root, summary, binary_hash, network)
    verify_issue_authoring_capture(root, summary, binary_hash, network)
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
            and after["attribution"]["basis"] == "writer-supplied"
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



def verify_issue_authoring_capture(root, summary, binary_hash, network):
    """Independently connect initial CLI fields, guarded edits and BDP observations."""
    def data(name):
        return json.loads((root / name).read_text())
    proof = data("client-artifacts.json")["issueAuthoring"]
    before, edited, after = (proof[key] for key in ("before", "edited", "after"))
    resource = summary["scope"] + "beads/prereq"
    initial = {"design": "Initial design — 雪", "acceptance_criteria": "Ready\r\n",
               "assignee": "author", "estimated_minutes": 0, "external_ref": " tracker #1 ",
               "spec_id": " spec ", "notes": " Initial\r\n雪 ", "due_at": "2000-01-01T00:00:00Z"}
    require(proof["initialFields"] == initial and all(before["properties"].get(k) == v for k, v in initial.items()),
            "initial authoring fields/nullable zero/literal notes missing")
    require(not before["properties"].get("lease_expires_at"), "initial assignment unexpectedly claimed")
    seed = [path for path in root.glob("*-issue-create/stdout.log") if data(str(path.relative_to(root)))["result"]["id"] == resource]
    require(len(seed) == 1, "missing unique initial authoring CLI seed")
    seeded = json.loads(seed[0].read_text())["result"]
    require(seeded["properties"] == before["properties"] and seeded["revision"] == before["revision"],
            "initial CLI/BDP authoring records differ")
    expected = dict(before["properties"], design="Revised", acceptance_criteria="Accepted", estimated_minutes=45,
                    external_ref="revised", spec_id="revised spec", due_at="2100-01-01T00:00:00Z", notes=initial["notes"] + "\nProgress",
                    updated_at=edited["properties"]["updated_at"])
    require(edited["properties"] == expected and edited["revision"] != before["revision"], "combined edit changed unrelated authoring properties")
    expected = dict(edited["properties"], estimated_minutes=0, updated_at=after["properties"]["updated_at"])
    expected.pop("external_ref")
    expected.pop("spec_id")
    expected.pop("due_at")
    require(after["properties"] == expected and after["revision"] != edited["revision"], "nullable clear lost notes or unrelated fields")
    require(all(item["id"] == resource and item["type"] == before["type"] and item["ownedLinks"] == before["ownedLinks"] for item in (edited, after)),
            "authoring changed identity/type/ownership")
    actor = "bdp-read-author"
    require(all(item["attribution"]["principal"] == actor
                and item["attribution"]["basis"] == "writer-supplied" for item in (edited, after)),
            "authoring attribution missing or has wrong basis")
    require(proof["properties"] == after["properties"] and proof["inventory"]["next"] is None
            and len(proof["inventory"]["items"]) == 4
            and [item for item in proof["inventory"]["items"] if item["id"] == resource] == [after], "authoring BDP current/properties/inventory disagree")
    require([item for item in data("python-after-delete/stdout.log") if item["id"] == resource] == [after],
            "unchanged Python consumer lost final authored Issue")
    require(proof["noop"]["result"]["issue"] == proof["edit"]["result"]["issue"], "scalar noop changed Issue")
    binary = data("02-memory-create/receipt.json")["argv"][0]
    edits = ["--estimate=45", "--external-ref=revised", "--spec-id=revised spec", "--due=2100-01-01T00:00:00Z"]
    for name, key, flags, predecessor, changed, record in (
        ("edit", "edit", edits + ["--design=Revised", "--acceptance=Accepted", "--append-notes=Progress"], before, True, edited),
        ("noop", "noop", edits, edited, False, edited),
        ("clear", "clear", ["--estimate=0", "--external-ref=", "--spec-id=", "--due="], edited, True, after),
    ):
        prefix = "client-cli-issue-authoring-" + name
        receipt, mutation = data(prefix + ".json"), data(prefix + ".stdout.log")
        require(receipt["argv"] == [binary, "update", resource, *flags, "--if-revision", predecessor["revision"], "--actor", actor, "--json"]
                and receipt["binarySha256"] == binary_hash and receipt["exitCode"] == 0 and receipt["signal"] is None,
                "authoring installed command provenance/guard/exit differs")
        for stream in ("stdout", "stderr"):
            require(digest(root / (prefix + "." + stream + ".log")) == receipt[stream + "Sha256"], "authoring output hash differs")
        require(mutation == proof[key] and mutation["schemaVersion"] == 1 and mutation["preview"] is True
                and mutation["result"]["changed"] is changed and mutation["result"]["issue"]["id"] == resource
                and mutation["result"]["issue"]["revision"] == record["revision"]
                and mutation["result"]["issue"]["properties"] == record["properties"], "authoring CLI/BDP mismatch")
    require(proof["beforeETag"] and proof["editedETag"] and proof["beforeETag"] != proof["editedETag"], "authoring edit ETag unchanged")
    for label, index_key, status, etag, record in (
        ("before", "beforeIndex", 200, proof["beforeETag"], before),
        ("edited", "editedIndex", 200, proof["editedETag"], edited),
        ("noop", "noopIndex", 304, proof["editedETag"], None),
    ):
        index = proof[index_key]
        require(type(index) is int and 0 <= index < len(network), "authoring HTTP index missing")
        observation = network[index]
        body = root / ("client-issue-authoring-" + label + ".body")
        require(observation["url"] == resource and observation["method"] == "GET" and observation["status"] == status
                and observation["headers"].get("etag") == etag and observation["bodyBytes"] == body.stat().st_size
                and observation["bodySha256"] == digest(body), "authoring HTTP status/ETag/body receipt differs")
        require(body.read_bytes() == b"" if record is None else json.loads(body.read_bytes()) == record, "authoring raw body differs")
    require(proof["beforeIndex"] < proof["editedIndex"] < proof["noopIndex"], "authoring HTTP observations unordered")


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


def output_tail(stdout, stderr, limit=TAIL_LINES):
    """Last lines of a failed command's test output and of its stderr.

    go test -json carries the test text in each event's Output field; any
    other stdout line is kept as it is.
    """
    text = []
    for line in stdout.decode(errors="replace").splitlines():
        try:
            item = json.loads(line)
        except ValueError:
            text.append(line)
            continue
        text.extend(item.get("Output", "").splitlines() if isinstance(item, dict) and "Action" in item else [line])
    return text[-limit:], stderr.decode(errors="replace").splitlines()[-limit:]


def report_failure_tail(label, stdout, stderr):
    output, errors = output_tail(stdout, stderr)
    for title, lines in (("test output", output), ("stderr", errors)):
        if lines:
            print(f"--- {label}: last {len(lines)} lines of {title} ---\n" + "\n".join(lines), flush=True)


def phase_table(receipts):
    """One row per receipt; go test groups also show their share of the timeout."""
    rows = ["### Graph core C0 phases", "", "| Phase | Roots | Seconds | % of go test timeout |", "|---|---:|---:|---:|"]
    for item in receipts:
        seconds, roots = item.get("elapsedSeconds"), item.get("roots")
        rows.append("| {} | {} | {} | {} |".format(
            item["label"], "" if roots is None else roots, "" if seconds is None else f"{seconds:.1f}",
            "" if roots is None or seconds is None else f"{seconds / GO_TEST_TIMEOUT:.0%}"))
    return "\n".join(rows) + "\n"


def slow_group_warnings(receipts):
    """Advisory annotations for go test groups that used 60% or more of their timeout."""
    return [f"::warning title=Slow go test group::{item['label']} took {item['elapsedSeconds']:.0f} s, "
            f"{item['elapsedSeconds'] / GO_TEST_TIMEOUT:.0%} of its {GO_TEST_TIMEOUT} s go test timeout"
            for item in receipts
            if item.get("roots") is not None and item.get("elapsedSeconds", 0) >= SLOW_GROUP_SECONDS]


def publish_phase_report(receipts, summary_path):
    for warning in slow_group_warnings(receipts):
        print(warning, flush=True)
    if summary_path:
        try:
            with open(summary_path, "a") as summary:
                summary.write(phase_table(receipts))
        except OSError as error:
            print(f"::warning title=Phase table not written::{error}", flush=True)


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

    def run(self, args, label, cwd=None, timeout=120, stdin=None, expected=0, owned_groups=None, roots=None):
        print(f"C0 {label}: {args}", flush=True)
        started = time.monotonic()

        def timing():
            return dict(elapsedSeconds=round(time.monotonic() - started, 3), **({} if roots is None else dict(roots=roots)))

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
                                      stdoutSHA256=hashlib.sha256(out).hexdigest(), **timing()))
            report_failure_tail(label, out, err)
            raise
        finally:
            self.children.discard(process.pid)
        (self.output / f"{label}.stdout").write_bytes(out)
        (self.output / f"{label}.stderr").write_bytes(err)
        if not group_exited(process.pid):
            os.killpg(process.pid, signal.SIGKILL)
            raise RuntimeError(f"{label} left a live process group after exit")
        self.receipts.append(dict(label=label, argv=[str(a) for a in args], pid=process.pid,
                                  exit=process.returncode, stdoutSHA256=hashlib.sha256(out).hexdigest(), **timing()))
        if owned_groups is not None:
            nested = cleanup_registered_groups(owned_groups)
            require(nested["registeredGroups"] == 0, "completed owner left nested groups registered")
        if process.returncode != expected:
            report_failure_tail(label, out, err)
        require(process.returncode == expected,
                f"{label} exited {process.returncode}, expected {expected}; see saved stdout/stderr")
        return out

    @contextlib.contextmanager
    def server(self, label):
        """One owned Dolt server for one command: fresh data directory and port, own log and shutdown receipt.

        The port is exported as BEADS_GRAPH_TEST_SERVER_PORT while the server is up and only then. Every
        server must stop cleanly (exit 0, not forced, group gone, port closed) or the step fails.
        """
        with socket.socket() as reservation:
            reservation.bind(("127.0.0.1", 0))
            port = reservation.getsockname()[1]
        server_root = self.output / (label + "-server-data")
        server_root.mkdir()
        log = (self.output / (label + "-server.log")).open("wb")
        started = time.monotonic()
        server = subprocess.Popen([str(self.dolt), "sql-server", "--host", "127.0.0.1", "--port", str(port), "--data-dir", str(server_root)],
                                  cwd=server_root, env=self.env, stdout=log, stderr=subprocess.STDOUT, start_new_session=True)
        self.children.add(server.pid)
        try:
            deadline = time.monotonic() + SERVER_READY_TIMEOUT
            while True:
                require(server.poll() is None, "owned server exited before readiness")
                try:
                    with socket.create_connection(("127.0.0.1", port), timeout=1):
                        break
                except OSError:
                    require(time.monotonic() < deadline, "server readiness timed out")
                    time.sleep(0.2)
            self.env["BEADS_GRAPH_TEST_SERVER_PORT"] = str(port)
            yield port
            require(server.poll() is None, "server exited during qualification")
        finally:
            self.env.pop("BEADS_GRAPH_TEST_SERVER_PORT", None)
            server.terminate()
            forced = False
            try:
                code = server.wait(timeout=SERVER_STOP_TIMEOUT)
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
            self.receipts.append(dict(label=label + "-server-shutdown", pid=server.pid, port=port,
                                      exit=code, forced=forced, groupGone=group_gone,
                                      portClosed=port_closed, elapsedSeconds=round(time.monotonic() - started, 3)))
            require(code == 0 and not forced and group_gone and port_closed,
                    f"server cleanup failed: exit={code}, forced={forced}, groupGone={group_gone}, portClosed={port_closed}")

    def tests(self, package, selector, files, label, required_subtests=(), groups=1):
        # Source discovery also sees accidentally build-excluded added tests.
        source_names = []
        for path in files:
            names = re.findall(r"^func (Test\w+)\(t \*testing.T\)", path.read_text(), re.MULTILINE)
            source_names.extend(name for name in names if re.search(selector, name))
        expected = set(source_names)
        require(len(source_names) == len(expected), f"{label}: duplicate source test roots")
        listing = self.run(["go", "test", "-tags", "gms_pure_go", "-list", selector, package],
                           label + "-discovery", timeout=600).decode()
        compiled_names = re.findall(r"^Test\w+$", listing, re.MULTILINE)
        listed = set(compiled_names)
        require(len(compiled_names) == len(listed), f"{label}: duplicate compiled test roots")
        require(listed == expected and expected, f"{label}: source and compiled test discovery disagree")
        flags = ["go", "test", "-tags", "gms_pure_go", "-json", "-count=1", "-p=1",
                 "-parallel=1", "-timeout=15m", "-run"]
        if groups == 1:
            with self.server(label):
                out = self.run([*flags, selector, package], label, timeout=960, roots=len(expected))
        else:
            selections = partition_required_tests(source_names, groups)
            plan = dict(package=package, selector=selector, sourceRoots=sorted(expected),
                        compiledRoots=sorted(listed), complete=False, groups=[
                            dict(label=f"{label}-group-{index}", roots=names,
                                 selector="^(" + "|".join(names) + ")$", passed=False)
                            for index, names in enumerate(selections, 1)])
            aggregate = self.output / (label + ".stdout")
            aggregate.write_bytes(b"")
            aggregate_error = self.output / (label + ".stderr")
            aggregate_error.write_bytes(b"")
            try:
                for group in plan["groups"]:
                    group_label, names, selected = group["label"], group["roots"], group["selector"]
                    (self.output / (label + "-groups.json")).write_text(json.dumps(plan, indent=2) + "\n")
                    try:
                        with self.server(group_label):
                            result = self.run([*flags, selected, package], group_label, timeout=960, roots=len(names))
                        verify_tests([json.loads(line) for line in result.splitlines()], set(names))
                        group["passed"] = True
                    finally:
                        # Keep actual partial output even if the group fails or times out.
                        for suffix, combined in (("stdout", aggregate), ("stderr", aggregate_error)):
                            part = self.output / (group_label + "." + suffix)
                            if part.exists():
                                with combined.open("ab") as target:
                                    target.write(part.read_bytes())
                                group[suffix + "SHA256"] = digest(part)
                executed = [name for group in plan["groups"] for name in group["roots"]]
                require(len(executed) == len(set(executed)) and set(executed) == expected,
                        f"{label}: grouped test union differs from complete discovery")
                out = aggregate.read_bytes()
                verify_tests([json.loads(line) for line in out.splitlines()], expected)
                plan["complete"] = True
            finally:
                plan["stdoutSHA256"] = digest(aggregate)
                plan["stderrSHA256"] = digest(aggregate_error)
                (self.output / (label + "-groups.json")).write_text(json.dumps(plan, indent=2) + "\n")
        events = [json.loads(line) for line in out.splitlines()]
        self.counts[label] = verify_tests(events, expected)
        completed = {e.get("Test") for e in events if e.get("Action") == "pass"}
        require(set(required_subtests) <= completed, f"{label}: required engine proof absent")

    def cli(self, work, label, *args):
        return json.loads(self.run([self.bd, *args, "--json"], label, cwd=work))

    def smoke_command(self, port):
        argv = [sys.executable, str(self.root / "scripts/graph-bdp-read-smoke.py"),
                "--bd", str(self.bd), "--server-port", str(port),
                "--bdp-checkout", str(self.client_checkout), "--client-manifest", str(self.client_manifest),
                "--node", str(self.node), "--output-dir", str(self.output / "http"),
                "--owned-groups", str(self.output / "http-owned-groups.jsonl"),
                "--command-timeout", str(SMOKE_COMMAND_TIMEOUT), "--total-timeout", str(SMOKE_TOTAL_TIMEOUT)]
        return argv, SMOKE_WRAPPER_TIMEOUT

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
        require_memory_creation(record, scope)
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
        with contextlib.ExitStack() as servers:
            # Sequential package runs/provisioning; concurrency inside same-store tests remains exercised.
            self.tests("./internal/graphpatch", "^Test", (self.root / "internal/graphpatch").glob("*_test.go"), "graphpatch")
            self.tests("./internal/storage/graphstore", "^Test", (self.root / "internal/storage/graphstore").glob("*_test.go"), "storage",
                       tuple("TestLinkUnlinkDefaultSourceReadBudget/" + engine + "/" + case
                             for engine in ("embedded", "server")
                             for case in ("id-default", "pair-default", "id-explicit")), groups=4)
            self.tests("./internal/configfile", "^TestGraphMode", (self.root / "internal/configfile").glob("graph_mode_test.go"), "config")
            self.tests("./internal/storage/issueops", "^TestResolve(CustomConfigStrict|InfraTypesStrict|ConfigLegacy)",
                       (self.root / "internal/storage/issueops").glob("config_strict_test.go"), "query-config")
            self.tests("./internal/storage/issueops", "^TestPrepareIssueForInsertNormalizesOptionalTimestampsToUTC$",
                       (self.root / "internal/storage/issueops").glob("prepare_timestamps_utc_test.go"), "utc-issueops")
            self.tests("./internal/storage/domain/db", "^TestNormalizeIssueTimestampsConvertsOptionalTimestampsToUTC$",
                       (self.root / "internal/storage/domain/db").glob("normalize_timestamps_utc_test.go"), "utc-db")
            self.tests("./internal/types", "^TestNormalizeOptionalTimestampsToUTCCoversEveryPointerTimestamp$",
                       (self.root / "internal/types").glob("types_test.go"), "utc-types")
            self.tests("./internal/templates/agents", "^TestGraphPreview",
                       (self.root / "internal/templates/agents").glob("*_test.go"), "agent-template")
            self.tests("./cmd/bd/setup", "^TestGraphPreview", (self.root / "cmd/bd/setup").glob("graph_claude_test.go"), "graph-claude-setup")
            self.tests("./cmd/bd", "^Test(GraphModeCLI|GraphPreview)", (self.root / "cmd/bd").glob("graph*test.go"), "cli",
                       ("TestGraphPreviewAgentInstructionsWorkflow/embedded/fresh",
                        "TestGraphPreviewAgentInstructionsWorkflow/embedded/shared-file",
                        "TestGraphPreviewAgentInstructionsWorkflow/embedded/skip-agents",
                        "TestGraphPreviewAgentInstructionsWorkflow/embedded/full-profile-refusal",
                        "TestGraphPreviewAgentInstructionsWorkflow/server/fresh",
                        "TestGraphPreviewAgentInstructionsWorkflow/server/shared-file",
                        "TestGraphPreviewAgentInstructionsWorkflow/server/skip-agents",
                        "TestGraphPreviewAgentInstructionsWorkflow/server/full-profile-refusal",
                        "TestGraphPreviewUsabilityWorkflow/embedded", "TestGraphPreviewUsabilityWorkflow/server",
                        "TestGraphPreviewClaudeStopWorkflow/embedded", "TestGraphPreviewClaudeStopWorkflow/server",
                        "TestGraphPreviewClaudeStopAutoInit/embedded", "TestGraphPreviewClaudeStopAutoInit/server",
                        "TestGraphPreviewClaudeStopAutoInitPreflight",
                        "TestGraphPreviewClaudeStopAutoInitRefusalMessages/global-plugin", "TestGraphPreviewClaudeStopAutoInitRefusalMessages/claude-symlink",
                        "TestGraphPreviewClaudeStopInstallFailure/embedded", "TestGraphPreviewClaudeStopInstallFailure/server",
                        "TestGraphPreviewClaudeStopSkipAgentsAlone/embedded", "TestGraphPreviewClaudeStopSkipAgentsAlone/server",
                        "TestGraphPreviewClaudeStopHookAdmissionNonBlocking/embedded", "TestGraphPreviewClaudeStopHookAdmissionNonBlocking/server",
                        "TestGraphPreviewClaudeStopProfileRefusal/embedded/missing", "TestGraphPreviewClaudeStopProfileRefusal/embedded/minimal", "TestGraphPreviewClaudeStopProfileRefusal/embedded/stale",
                        "TestGraphPreviewClaudeStopProfileRefusal/server/missing", "TestGraphPreviewClaudeStopProfileRefusal/server/minimal", "TestGraphPreviewClaudeStopProfileRefusal/server/stale",
                        "TestGraphPreviewCompatibilityDefaultsWorkflow/embedded", "TestGraphPreviewCompatibilityDefaultsWorkflow/server",
                        "TestGraphPreviewIssueAuthoringWorkflow/embedded", "TestGraphPreviewIssueAuthoringWorkflow/server",
                        "TestGraphPreviewIssueAppendWorkflow/embedded", "TestGraphPreviewIssueAppendWorkflow/server",
                        "TestGraphPreviewIssueClaimWorkflow/embedded", "TestGraphPreviewIssueClaimWorkflow/server",
                        "TestGraphPreviewIssueAssignmentWorkflow/embedded", "TestGraphPreviewIssueAssignmentWorkflow/server",
                        "TestGraphPreviewMixedCoreWorkflow/embedded", "TestGraphPreviewMixedCoreWorkflow/server",
                        "TestGraphPreviewMemoryDeleteWorkflow/embedded", "TestGraphPreviewMemoryDeleteWorkflow/server",
                        "TestGraphPreviewQueryWorkflow/embedded", "TestGraphPreviewQueryWorkflow/server",
                        "TestGraphPreviewMemoryReadsWorkflow/embedded", "TestGraphPreviewMemoryReadsWorkflow/server",
                        "TestGraphPreviewPropertiesPatchWorkflow/embedded", "TestGraphPreviewPropertiesPatchWorkflow/server"),
                       groups=3)
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
            port = servers.enter_context(self.server("capture"))
            self.capture("embedded", port)
            self.capture("server", port)
            # Per command < smoke total < this wrapper (SMOKE_*): the wrapper outlasts the smoke,
            # so a command that outruns its deadline surfaces as the smoke's labelled failure.
            argv, timeout = self.smoke_command(port)
            self.run(argv, "bdp-http-capture", timeout=timeout,
                     owned_groups=self.output / "http-owned-groups.jsonl")
            verify_http_capture(self.output / "http", digest(self.bd))
            http_summary = json.loads((self.output / "http/summary.json").read_text())
            for field, relative in [("harness_sha256", "scripts/graph-bdp-read-smoke.py"),
                                    ("client_harness_sha256", "scripts/graph-bdp-read-client.mjs"),
                                    ("python_example_sha256", "examples/bdp-read/read_beads.py")]:
                require(http_summary[field] == digest(self.root / relative), "HTTP capture source changed")
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
        publish_phase_report(qualification.receipts, os.environ.get("GITHUB_STEP_SUMMARY"))


if __name__ == "__main__":
    main()
