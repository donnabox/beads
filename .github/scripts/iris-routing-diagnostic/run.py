"""ONE branch-only routing observation. Not a shipping CI gate or runner API."""
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import signal
import subprocess
import time
import xml.etree.ElementTree as ET
from urllib.parse import unquote, urlsplit

from held_group import OwnedGroup, cleanup_owned, supervise_owned
from fixture_observer import RoutingFixtureObserver

BRANCH = "refs/heads/codex/iris-routing-deadline-diag-20261029"
EXPECTED = [
    "TestContributorRoutingTracer",
    "TestContributorRoutingTracer/ExpandPath_tilde_expansion",
    "TestContributorRoutingTracer/DetermineTargetRepo_contributor_routes_to_planning",
    "TestContributorRoutingTracer/DetermineTargetRepo_maintainer_stays_local",
    "TestContributorRoutingTracer/E2E_routing_decision_with_store",
]
SOURCES = {
    'tools/bazel/go_test_pinned_shard.sh': '027580fd9c779275f1f8c70ae1ca5e218a139d9572777b78056af946b2076792',
    'cmd/bd/BUILD.bazel': '01d554667ade3e350fccf96823dae1467c12f6231a9d249e077ae4a2b81790c2',
    "cmd/bd/contributor_routing_e2e_test.go": "3ceefafd4e30957dc87fd51efdc4972e1401f2df30c152235c327de6c1bae1e5",
    "internal/storage/dolt/issues.go": "7a6dc6631cf2fa4327a8e25bcc5b6c587f416f441c32f4ce3d89b589548be177",
    "internal/storage/dolt/store.go": "c6bdd23d0292af0f2ac2f8ef97a2652f68c2da60868eb6b2a846c677c8d61a06",
    "internal/testutil/localdoltserver.go": "2515e188b89e1676921d3a7dd215b8fa824f1206a72aa81da9aa85e491f6f3df",
    "tools/bazel/test_env.sh": "8bb3e5f57e2d8816753a4e0e9f08be5d64e51b89bde61d0d799f3bf04e1dd3d0",
    "tools/bazel/BUILD.bazel": "f363b5f75da05a33646fc1deb2f615eeaf19ee248c4cbb7c206a34bd5c51651c",
    "tools/bazel/dolt.bzl": "699d49317af8c669b92ee2cc03fdaed75aeac2f873d7c1221dcd751cde20f49f",
    ".bazelrc": "19b90656806bf2c28ad2bdd7aac03911ba7ac5c27b801b37f384f6c9fbcbe4bb",
    ".github/workflows/bazel.yml": "b6b86fd56ff6f3d59bf8476c0cf9263e4218f7c2b638cd65b3b7f24969f4f8a3",
    ".github/scripts/iris-routing-diagnostic/fixture_observer.py": "918f1ed7e4f4a28fdf96ca2d2839773722e2bf3c3faa731e8ab8c05976f958fa"
}


def digest(path):
    return hashlib.sha256(Path(path).read_bytes()).hexdigest()


def raw_go_events(lines):
    """Derived log evidence only; original raw stream is retained independently."""
    events = []
    for number, line in enumerate(lines, 1):
        line = line.lstrip("\x16\t ")
        run = re.match(r"^\s*=== RUN\s+(\S+)\s*$", line)
        terminal = re.match(r"^\s*--- (PASS|FAIL|SKIP):\s+(\S+)", line)
        if run:
            events.append({"line": number, "event": "RUN", "test": run[1]})
        if terminal:
            events.append({"line": number, "event": terminal[1], "test": terminal[2]})
    return events



def distinct_binary_path(text):
    """Configured data deps may repeat a file; distinct paths remain ambiguous."""
    paths = sorted({line for line in text.splitlines()
                    if line.startswith("bazel-out/") and line.endswith("/bin/cmd/bd/bd_test_/bd_test")})
    assert len(paths) == 1, "ambiguous/missing distinct binary path; full merged raw retained"
    return paths[0]


def preserve_test_outputs(logdir, evidence, result, phase="configured-info"):
    """Copy raw test files; preservation failure cannot suppress derived planes."""
    source_path = None if logdir is None else str(logdir)
    receipt = {"phase": phase, "resolvedSourcePath": source_path, "copied": []}
    result.setdefault("rawTestlogPreservations", []).append(receipt)
    def refused(error):
        result["errors"].append({"evidencePlane": "raw-testlogs", "phase": phase, "resolvedSourcePath": source_path,
                                 "type": type(error).__name__, "message": str(error)})
    try:
        source_path = None if logdir is None else str(logdir.resolve())
        receipt["resolvedSourcePath"] = source_path
        if logdir is None or not logdir.exists():
            raise FileNotFoundError("raw testlogs path unset or missing")
        out = evidence / "raw-testlogs"
        out.mkdir(exist_ok=True)
        for source in logdir.rglob("*"):
            try:
                if source.name not in ["test.log", "test.xml"] or not source.is_file():
                    continue
                dest = out / source.relative_to(logdir)
                dest.parent.mkdir(parents=True, exist_ok=True)
                reused = dest.exists()
                if reused:
                    if digest(source) != digest(dest):
                        raise RuntimeError("raw file differs from already preserved exact-action output: " + str(source))
                else:
                    shutil.copyfile(source, dest)
                receipt["copied"].append({"source": str(source), "file": str(dest.relative_to(evidence)),
                                          "sha256": digest(dest), "reused": reused})
            except Exception as error:
                result["errors"].append({"evidencePlane": "raw-testlog-file", "phase": phase,
                                         "source": str(source), "type": type(error).__name__, "message": str(error)})
        if not receipt["copied"]:
            raise FileNotFoundError("raw testlogs path yielded zero copied log/XML files")
    except Exception as error:
        refused(error)



def read_bep_events(path):
    """Strict explicit UTF-8 per physical line; retain earlier complete events."""
    events, parse_errors, blanks = [], [], 0
    with path.open("rb") as stream:
        for number, raw in enumerate(stream, 1):
            try:
                line = raw.decode("utf-8", errors="strict")
                if not line.strip():
                    blanks += 1
                    continue
                event = json.loads(line)
                if not isinstance(event, dict):
                    raise ValueError("BEP event is not an object")
                events.append(event)
            except Exception as error:
                parse_errors.append({"line": number, "type": type(error).__name__, "message": str(error)})
    return events, parse_errors, blanks


def preserve_bep_outputs(evidence, result, owned_root):
    """Recover only exact sole-attempt output URIs within this owned Bazel root."""
    receipt = {"phase": "bep-exact-action", "ownedRoot": str(owned_root.resolve()), "copied": []}
    result.setdefault("rawTestlogPreservations", []).append(receipt)
    def refused(error, source=None):
        result["errors"].append({"evidencePlane": "bep-output-preservation", "source": source,
                                 "type": type(error).__name__, "message": str(error)})
    try:
        events, errors, blanks = read_bep_events(evidence / "bep.json")
        receipt.update(bepParseErrors=errors, blankBepLines=blanks)
        if errors:
            refused(ValueError("BEP has undecodable/malformed lines; complete exact-action outputs may still be retained"))
        actions = [event for event in events if "testResult" in event]
        if len(actions) != 1:
            raise ValueError("BEP raw recovery requires exactly one testResult")
        action = actions[0]
        ident = action["id"]["testResult"]
        if ident["label"] != "//cmd/bd:bd_dolt_server_test" or any(ident.get(key) != 1 for key in ["run", "shard", "attempt"]):
            raise ValueError("BEP raw recovery refuses other target/run/shard/attempt")
        receipt["testResultIdentity"] = ident
        outputs = [output for output in action["testResult"].get("testActionOutput", [])
                   if output.get("name") in ["test.log", "test.xml"]]
        names = [output["name"] for output in outputs]
        if sorted(names) != ["test.log", "test.xml"]:
            raise ValueError("BEP raw recovery requires exactly one named test.log and test.xml")
        for output in outputs:
            uri = output.get("uri")
            try:
                if not isinstance(uri, str) or re.search(r"%(?![0-9a-fA-F]{2})", uri):
                    raise ValueError("invalid local output URI encoding")
                parsed = urlsplit(uri)
                if parsed.scheme != "file" or parsed.netloc or parsed.query or parsed.fragment:
                    raise ValueError("BEP raw recovery accepts local file URI only")
                source = Path(unquote(parsed.path, encoding="utf-8", errors="strict"))
                expected = "/testlogs/cmd/bd/bd_dolt_server_test/" + output["name"]
                if not source.is_absolute() or not str(source).endswith(expected):
                    raise ValueError("BEP output path does not name the exact unsharded action output")
                if ".." in source.parts or not source.is_relative_to(owned_root.resolve()):
                    raise ValueError("BEP output escapes this owned Bazel root")
                source = source.resolve(strict=True)
                if not source.is_relative_to(owned_root.resolve()):
                    raise ValueError("BEP output escapes this owned Bazel root")
                dest = evidence / "raw-testlogs" / output["name"]
                dest.parent.mkdir(exist_ok=True)
                reused = dest.exists()
                if reused:
                    if digest(source) != digest(dest):
                        raise RuntimeError("BEP output differs from already preserved raw file")
                else:
                    shutil.copyfile(source, dest)
                receipt["copied"].append({"uri": uri, "source": str(source),
                                          "file": str(dest.relative_to(evidence)),
                                          "sha256": digest(dest), "reused": reused})
            except Exception as error:
                refused(error, uri)
    except Exception as error:
        refused(error)


def collect_test_evidence(evidence, result, save):
    """Keep independent evidence planes, including partial/corrupt inputs.

    Raw inputs have already been preserved. Failed result persistence remains a
    sticky refusal, but cannot suppress collection of later evidence planes.
    """
    def persist(plane):
        try:
            save()
        except Exception as error:
            result["errors"].append({"evidencePlane": "result-save", "sourcePlane": plane,
                                     "type": type(error).__name__, "message": str(error)})

    def refused(plane, error):
        result["errors"].append({"evidencePlane": plane,
                                 "type": type(error).__name__, "message": str(error)})
        persist(plane)

    try:
        events, parse_errors, blanks = read_bep_events(evidence / "bep.json")
        for record in parse_errors:
            result["errors"].append({"evidencePlane": "bep-line", **record})
        result["blankBepLines"] = blanks
        result["bepParseErrors"] = parse_errors
        result["completeBepEvents"] = len(events)
        actions = [event for event in events if "testResult" in event]
        result["actualTestResults"] = actions
        persist("bep")
        assert len(actions) == 1, "one actual test attempt only"
        action = actions[0]
        result["actualTestResult"] = action
        persist("bep")
        ident = action["id"]["testResult"]
        assert ident["label"] == "//cmd/bd:bd_dolt_server_test"
        assert all(ident.get(key) == 1 for key in ["run", "shard", "attempt"])
        assert not action["testResult"].get("cachedLocally", False)
        info = action["testResult"].get("executionInfo", {})
        assert not info.get("cachedRemotely", False)
        assert info.get("strategy") in ["linux-sandbox", "processwrapper-sandbox", "local"]
    except Exception as error:
        refused("bep", error)

    try:
        files = sorted((evidence / "raw-testlogs").rglob("test.xml"))
        records = []
        for path in files:
            record = {"file": str(path.relative_to(evidence))}
            try:
                record["sha256"] = digest(path)
                cases = list(ET.parse(path).getroot().iter("testcase"))
                record.update(names=sorted(case.attrib["name"] for case in cases),
                              failures=sum(len(case.findall("failure")) for case in cases),
                              errors=sum(len(case.findall("error")) for case in cases),
                              skips=sum(len(case.findall("skipped")) for case in cases))
            except Exception as error:
                record["parseError"] = {"type": type(error).__name__, "message": str(error)}
            records.append(record)
        result["xmlCensuses"] = records
        persist("xml-census")
        assert len(records) == 1, "expected one unsharded action XML"
        assert "parseError" not in records[0], "XML parse refused; raw and error retained"
        result["census"] = records[0]
        persist("xml-census")
        assert records[0]["names"] == sorted(EXPECTED) and not records[0]["skips"]
    except Exception as error:
        refused("xml-census", error)

    try:
        rawlogs = sorted((evidence / "raw-testlogs").rglob("test.log"))
        events, parse_errors = [], []
        for path in rawlogs:
            try:
                for event in raw_go_events(path.read_text(encoding="utf-8", errors="replace").splitlines()):
                    event["file"] = str(path.relative_to(evidence))
                    events.append(event)
            except Exception as error:
                record = {"file": str(path.relative_to(evidence)),
                          "type": type(error).__name__, "message": str(error)}
                parse_errors.append(record)
                result["errors"].append({"evidencePlane": "raw-log-file", **record})
        result["rawLogParseErrors"] = parse_errors
        (evidence / "DERIVED-RAW-GO-EVENTS.json").write_text(json.dumps(events, indent=2) + "\n")
        persist("derived-raw-go-events")
        assert len(rawlogs) == 1, "expected one unsharded raw test log"
        assert sorted(event["test"] for event in events if event["event"] == "RUN") == sorted(EXPECTED)
        assert sorted(event["test"] for event in events if event["event"] in ["PASS", "FAIL", "SKIP"]) == sorted(EXPECTED)
    except Exception as error:
        refused("derived-raw-go-events", error)


def main():
    assert os.environ["GITHUB_WORKFLOW"] == "Bazel"
    assert os.environ["GITHUB_EVENT_NAME"] == "workflow_dispatch"
    assert os.environ["GITHUB_REF"] == BRANCH
    assert os.environ["GITHUB_RUN_ATTEMPT"] == "1", "no job rerun"
    assert os.uname().sysname == "Linux" and os.uname().machine == "x86_64"
    workspace = Path(os.environ["GITHUB_WORKSPACE"]).resolve()
    root = Path(os.environ["RUNNER_TEMP"]) / "iris-routing-diagnostic"
    root.mkdir(mode=0o700)  # Refuse reuse; never remove an earlier run.
    for name in ["home", "tmp"]:
        (root / name).mkdir(mode=0o700)
    evidence = root / "evidence"
    evidence.mkdir(mode=0o700)
    started = time.monotonic()
    end = started + 3300
    cancelled = []
    for number in [signal.SIGINT, signal.SIGTERM, signal.SIGHUP]:
        signal.signal(number, lambda n, frame: cancelled.append(n))
    result = {"scope": "diagnostic only", "source": SOURCES, "stages": [],
              "errors": [], "normalArbitraryDescendantCleanupProven": False}
    result["evidenceLimits"] = {
        "poolStats": "DoltStore.db cumulative stats; selected CreateIssue uses it, not a general caller-chosen beginner measurement",
        "instrumentation": "isolated diagnostic branch only; getenv/stderr logging and 1Hz proc/environ/fd/socket reads perturb timing; test executable hashing and memoized engine hashing add observer I/O/page-cache/CPU load inside the window; not shipping code or original-load equivalence",
        "timeout": "original1200s inclusive custody retained;1190s observation may truncate late alarm/writeout; incomplete evidence never green",
        "cleanup": "post run_under wrapper trap/Bazel shutdown only; normal fixture/arbitrary descendant cleanup not proven",
        "postCensus": "scoped remaining processes are not automatically classified as fixture leaks",
        "configuredTestlogsInfo": "same config/race argv retained as cross-check; info does not analyze Starlark options in pinned Bazel; exact-action owned BEP URIs provide raw recovery; info failures/missing paths remain sticky",
        "resultPersistence": "best-effort saves within evidence planes retain sticky failures; command/final save failures still exit nonzero and may leave partial RESULT.json",
        "enginePath": "prebuilt pinned release content is configuration-independent; actual path/version/content hash are recorded, not inferred from an unconfigured info path",
        "preflightFailure": "workflow/controller preflight refusal may have job logs only, no evidence artifact; never a test verdict",
    }
    env = os.environ.copy()
    for key in list(env):
        if key.startswith("RBE_") or key.endswith("_TOKEN") or key in [
            "BAZEL_REMOTE_EXECUTOR", "BAZEL_FORK_REMOTE", "IRIS_ROUTE_DIAG01_START"
        ]:
            env.pop(key, None)
    env.update(HOME=str(root / "home"), TMPDIR=str(root / "tmp"))
    env["DOLT_DISABLE_EVENT_FLUSH"] = "1"
    bazel = ["/usr/local/bin/bazel", "--output_user_root=" + str(root / "bazel")]
    bazel_started = False
    logdir = None
    groups = []
    fixture = None

    def save():
        (evidence / "RESULT.json").write_text(json.dumps(result, indent=2) + "\n")

    def call(label, argv, bound, *, check=True, cleanup=False, stdin=None, observer=None):
        nonlocal bazel_started
        if not cleanup and cancelled:
            raise RuntimeError("cancelled before " + label)
        remaining = end - time.monotonic() - (0 if cleanup else 60)
        if remaining <= 0:
            raise RuntimeError("outer deadline before " + label)
        began = time.monotonic()
        with (evidence / (label + ".log")).open("xb") as output:
            child = subprocess.Popen(argv, cwd=workspace, env=env, stdin=stdin,
                                     stdout=output, stderr=subprocess.STDOUT,
                                     start_new_session=True)
            if argv[:1] == bazel[:1]:
                bazel_started = True
            group = OwnedGroup(child)
            groups.append((label, group))
            def observed_sleep(seconds):
                if observer is not None:
                    observer.observe()
                time.sleep(seconds)
            receipt = supervise_owned(
                group, time.monotonic, observed_sleep,
                (lambda: False) if cleanup else (lambda: bool(cancelled)),
                min(bound, remaining), lambda seconds: print(label, round(seconds), flush=True),
                began, cleanup=lambda owned, now, sleep: cleanup_owned(owned, now=now, sleep=time.sleep))
        receipt.update(label=label, argv=argv, logSHA256=digest(evidence / (label + ".log")),
                       allInclusiveSeconds=time.monotonic() - began)
        result["stages"].append(receipt)
        save()  # Preserve raw and custody result before deciding it passed.
        c = receipt["cleanup"]
        okay = (not receipt["operationError"] and not receipt["observation"]["timedOut"]
                and c.get("parentReaped") and c.get("groupGone")
                and not c.get("extraCleanupRequired") and c.get("childExit") == 0
                and (cleanup or not cancelled))
        if label == "one-routing-root" and receipt["allInclusiveSeconds"] >= 1200:
            okay = False
        if check and not okay:
            raise RuntimeError(label + " failed; see preserved receipt")
        return okay, (evidence / (label + ".log")).read_text(encoding="utf-8", errors="replace").strip()

    def unique_line(text, predicate):
        rows = [line for line in text.splitlines() if predicate(line)]
        assert len(rows) == 1, "ambiguous/missing metadata row; full merged raw retained"
        return rows[0]


    try:
        for path, expected in SOURCES.items():
            assert digest(workspace / path) == expected, path
        assert digest(Path(__file__).with_name("held_group.py")) == "4eab6b89a75c979c8c22172c9f482a9cc0081af39be58124c9dc86982e296f4c"
        _, head = call("git-head", ["/usr/bin/git", "rev-parse", "HEAD"], 10)
        head = unique_line(head, lambda line: bool(re.fullmatch(r"[0-9a-f]{40}", line)))
        assert head == os.environ["GITHUB_SHA"]
        result["head"] = head
        call("engine-materialize", bazel + ["build", "//tools/bazel:dolt"], 900)
        _, execution = call("execution-root", bazel + ["info", "execution_root"], 30)
        execution = Path(unique_line(execution, lambda line: line.startswith(str(root / "bazel") + "/")))
        assert execution.is_relative_to(root / "bazel")
        _, engine_file = call("engine-file", bazel + ["cquery", "//tools/bazel:dolt", "--output=files"], 60)
        engine_file = unique_line(engine_file, lambda line: bool(re.fullmatch(r"(?:bazel-out/|external/|/)[^\s]+/dolt", line)))
        engine = (execution / engine_file).resolve()
        result["engine"] = {"file": str(engine), "sha256": digest(engine),
                            "bytes": engine.stat().st_size, "archiveVersion": "2.2.0",
                            "archiveSHA256": "1f7ad8c2622995789420a3fb0f2d16b4aa7430000a825dd91d5938f36480cbf6"}
        _, version = call("engine-version", [str(engine), "version"], 10)
        assert re.search(r"\b2\.2\.0\b", version), version
        result["engine"]["actualVersion"] = version
        # Build the same race-enabled test target once before the separately
        # bounded1200s test custody window; this does not execute any test.
        call("test-build", bazel + ["build", "//cmd/bd:bd_dolt_server_test",
             "--config=doltserver-cmd", "--@rules_go//go/config:race"], 1800)
        _, binary_file = call("test-binary-file", bazel + ["cquery", "labels(data, //cmd/bd:bd_dolt_server_test)",
                              "--config=doltserver-cmd", "--@rules_go//go/config:race", "--output=files"], 60)
        binary_file = distinct_binary_path(binary_file)
        result["testBinaryRelation"] = {
            "configuredDataQuery": "labels(data, //cmd/bd:bd_dolt_server_test)",
            "selectionScope": "one distinct configured data-output path with pinned rules_go/package suffix /bin/cmd/bd/bd_test_/bd_test; not an independent configured-label query",
            "file": binary_file,
            "sourceContract": "pinned cmd/bd/BUILD.bazel passes bd_test as second wrapper argument; pinned go_test_pinned_shard.sh execs that argument unsharded",
        }
        result["testBinarySHA256"] = digest(execution / binary_file)
        fixture = RoutingFixtureObserver(root / "bazel", result["testBinarySHA256"], result["engine"]["sha256"])
        logs_okay, logs = call("testlogs-root", bazel + ["info", "bazel-testlogs",
                              "--config=doltserver-cmd", "--@rules_go//go/config:race"], 30, check=False)
        info_custody = result["stages"][-1]["cleanup"]
        if not info_custody.get("parentReaped") or not info_custody.get("groupGone"):
            raise RuntimeError("configured testlogs info lost custody; refuse test launch")
        try:
            if not logs_okay:
                raise RuntimeError("configured testlogs info command/custody failed; no unconfigured command fallback")
            logs = unique_line(logs, lambda line: line.startswith(str(root / "bazel") + "/"))
            logdir = Path(logs) / "cmd/bd/bd_dolt_server_test"
            if not logdir.resolve().is_relative_to(root / "bazel"):
                raise ValueError("configured testlogs info path escapes owned Bazel root")
        except Exception as error:
            logdir = None
            result["errors"].append({"evidencePlane": "configured-testlogs-info", "type": type(error).__name__, "message": str(error)})
        argv = bazel + ["test", "//cmd/bd:bd_dolt_server_test", "--config=doltserver-cmd",
                        "--@rules_go//go/config:race",
                        "--test_sharding_strategy=disabled", "--test_arg=-test.run=^TestContributorRoutingTracer$",
                        "--test_arg=-test.count=1", "--nocache_test_results",
                        "--flaky_test_attempts=1", "--experimental_remote_cache_eviction_retries=0",
                        "--build_event_json_file=" + str(evidence / "bep.json"),
                        "--execution_log_compact_file=" + str(evidence / "exec.log.zst"),
                        "--experimental_build_event_upload_strategy=local", "--test_output=all"]
        #1190s observation + bounded retained-child cleanup/reap leaves room
        # under1200s inclusive; reaching1200 is a failed diagnostic outcome.
        okay, _ = call("one-routing-root", argv, 1190, check=False, observer=fixture)
        # Read only completed BEP metadata to recover exact owned raw output
        # locations before general evidence interpretation or later Bazel calls.
        preserve_bep_outputs(evidence, result, root / "bazel")
        preserve_test_outputs(logdir, evidence, result, "post-test-configured-info")
        result["diagnosticCommandSucceeded"] = okay
        collect_test_evidence(evidence, result, save)
        assert digest(execution / binary_file) == result["testBinarySHA256"]
        for path, expected in SOURCES.items():
            assert digest(workspace / path) == expected
        if not okay:
            raise RuntimeError("diagnostic command/custody failed; underlying test result is separately retained; no retry")
        if result["errors"]:
            raise RuntimeError("diagnostic evidence refused; independent planes retained; no retry")
        if result["census"]["failures"] or result["census"]["errors"]:
            raise RuntimeError("actual routing test assertions failed; no retry")
    except BaseException as error:
        result["errors"].append({"type": type(error).__name__, "message": str(error)})
    finally:
        try:
            preserve_test_outputs(logdir, evidence, result, "final-configured-info")
        except BaseException as error:
            result["errors"].append({"rawPreservation": str(error)})
        for label, group in groups:
            if not group.reaped:
                result["errors"].append({"lateCleanup": label, "receipt": cleanup_owned(group)})
        if bazel_started:
            try:
                call("owned-bazel-shutdown", bazel + ["shutdown"], 30, cleanup=True)
            except BaseException as error:
                result["errors"].append({"shutdown": str(error)})
        if fixture is not None:
            result["fixturePostconditions"] = fixture.finish()
            if result["fixturePostconditions"].get("error"):
                result["errors"].append({"fixture": result["fixturePostconditions"]["error"]})
        # Read-only scoped observer. Never signal/reap a PID learned here.
        remaining = []
        observation_errors = []
        for proc in Path("/proc").glob("[0-9]*"):
            try:
                if int(proc.name) == os.getpid() or proc.stat().st_uid != os.geteuid():
                    continue
                cmd = (proc / "cmdline").read_bytes()
                environ = (proc / "environ").read_bytes()
                if str(root).encode() in cmd or str(root).encode() in environ:
                    remaining.append({"pid": int(proc.name), "commandSHA256": hashlib.sha256(cmd).hexdigest()})
            except (FileNotFoundError, ProcessLookupError):
                pass
            except OSError as error:
                observation_errors.append({"pid": proc.name, "error": str(error)})
        result["scopedPostObservation"] = {"remaining": remaining, "errors": observation_errors,
            "scope": "same-uid processes retaining unique outputroot/HOME/TMPDIR/TEST_TMPDIR; not arbitrary descendants"}
        if remaining or observation_errors or cancelled or time.monotonic() >= end:
            result["errors"].append({"postconditions": "leftovers/incomplete observation/cancellation/deadline"})
        result["cancelled"] = cancelled
        result["elapsedSeconds"] = time.monotonic() - started
        result["greenDiagnosticOnly"] = not result["errors"]
        save()
    return 0 if result["greenDiagnosticOnly"] else 1


if __name__ == "__main__":
    raise SystemExit(main())
