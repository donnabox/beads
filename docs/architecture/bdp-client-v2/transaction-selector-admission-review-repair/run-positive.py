#!/usr/bin/env python3
"""Four frozen source-positive commands only. Preparation is not execution approval."""
import argparse
from collections import Counter
from datetime import datetime, timezone
import hashlib
import json
import os
from pathlib import Path
import signal
import subprocess
import sys
import time

E = Path(__file__).resolve().parent
W = Path("/Users/dbox/repos/wt/vickie/vickie-selector-review-repair")
COMMANDS_SHA = "39997323fe8314348aababd29199b55371e3204b6e137ac010f1b972836c0683"
BINDING_SHA = "03a861851a195d5da505f311291bee6d3cd98f63866d4b6df756e2c3ae289afd"
SELECTION_SHA = "fb5ea49170f8cfe585e80f7a52ac86ade6d5f15533872bae416832fa250c2a06"
IDS = ["read-selector.test", "transaction-set.test", "transaction-submission.test", "transaction-http.test"]
COUNTS = [67, 6, 13, 4]
FIXTURE_PREFIXES = ("bdp-set-core-", "bdp-tx-submission-", "bdp-tx-http-")
CACHE_PREFIXES = ("node-compile-cache",)
RUN = E / "run"
SUMMARY = RUN / "positive-summary.json"


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def now():
    return datetime.now(timezone.utc).isoformat()


def save(path, data):
    path.write_text(json.dumps(data, indent=2) + "\n")


def verify():
    assert sha(E / "commands.json") == COMMANDS_SHA, "commands binding changed"
    assert sha(E / "source-binding.json") == BINDING_SHA, "source binding changed"
    assert sha(E / "selection.json") == SELECTION_SHA, "selection binding changed"
    bound = json.loads((E / "source-binding.json").read_text())
    assert Path(bound["workspace"]) == W and len(bound["files"]) == 12
    head = subprocess.check_output(["git", "rev-parse", "HEAD"], cwd=W, text=True).strip()
    assert head == bound["parent"], "draft parent changed"
    actual = {}
    for row in bound["files"]:
        actual[row["path"]] = sha(W / row["path"])
        assert actual[row["path"]] == row["sha256"], "source changed: " + row["path"]
    return {"head": head, "files": actual, "commandsSha256": COMMANDS_SHA,
            "bindingSha256": BINDING_SHA, "selectionSha256": SELECTION_SHA}


def group_snapshot(pgid):
    # No global process arguments/environment are retained. Only this owned group.
    output = subprocess.check_output(
        ["ps", "-axo", "pid=,ppid=,pgid=,rss=,etime=,state=,comm="], text=True, timeout=5)
    rows = []
    for line in output.splitlines():
        parts = line.strip().split(None, 6)
        if len(parts) == 7 and int(parts[2]) == pgid:
            rows.append({"pid": int(parts[0]), "ppid": int(parts[1]), "pgid": pgid,
                         "rssKiB": int(parts[3]), "elapsed": parts[4], "state": parts[5],
                         "executable": parts[6]})
    return rows


def entries(directory):
    # Do not recurse into or remove unknown paths. Root-level names suffice for owned fixtures.
    return sorted(p.name for p in directory.iterdir())


def stop_owned(proc, record):
    before = group_snapshot(proc.pid)
    record["beforeStop"] = before
    if before:
        try:
            os.killpg(proc.pid, signal.SIGTERM)
            record["signals"].append("SIGTERM")
        except ProcessLookupError:
            pass
        until = time.monotonic() + 3
        while time.monotonic() < until:
            proc.poll()  # Reap the leader while observing its owned group.
            if not group_snapshot(proc.pid):
                break
            time.sleep(0.1)
        remaining = group_snapshot(proc.pid)
        if remaining:
            try:
                os.killpg(proc.pid, signal.SIGKILL)
                record["signals"].append("SIGKILL")
            except ProcessLookupError:
                pass
    try:
        proc.wait(timeout=2)
    except subprocess.TimeoutExpired:
        record["stopWaitTimedOut"] = True
    record["afterStop"] = group_snapshot(proc.pid)


def classify_result(path, expected):
    data = json.loads(path.read_text())
    assertions = [a for f in data["testResults"] for a in f["assertionResults"]]
    passed = [a["fullName"] for a in assertions if a["status"] == "passed"]
    failed = [{"title": a["fullName"], "status": a["status"],
               "failureMessages": a.get("failureMessages", [])}
              for a in assertions if a["status"] not in ("passed", "pending", "skipped", "todo", "disabled")]
    skipped = [a["fullName"] for a in assertions
               if a["status"] in ("pending", "skipped", "todo", "disabled")]
    return {"jsonSha256": sha(path), "passedTitles": passed, "passedCount": len(passed),
            "failed": failed, "unselectedSkippedTitles": skipped, "skippedCount": len(skipped),
            "selectedSkipped": [t for t in skipped if t in expected],
            "expectedTitles": expected,
            "accepted": bool(data.get("success")) and not failed and Counter(passed) == Counter(expected)
                        and not any(t in expected for t in skipped)}


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--release", type=Path, required=True)
    parser.add_argument("--release-sha256", required=True)
    args = parser.parse_args()
    assert sha(args.release) == args.release_sha256, "release file hash mismatch"
    release = json.loads(args.release.read_text())
    assert release["scope"] == "selector-source-positives-only"
    assert release["scriptSha256"] == sha(Path(__file__))
    assert release["commandsSha256"] == COMMANDS_SHA
    assert release["sourceBindingSha256"] == BINDING_SHA
    initial = verify()
    all_commands = json.loads((E / "commands.json").read_text())["commands"]
    selections = json.loads((E / "selection.json").read_text())["sourceSelection"]
    commands = []
    for ident, count, selection in zip(IDS, COUNTS, selections, strict=True):
        matches = [c for c in all_commands if c["id"] == ident]
        assert len(matches) == 1
        c = matches[0]
        assert c["cwd"] == str(W) and c["expectedCases"] == count and c["outerTimeoutMs"] == 120000
        assert c["argv"].count("--maxWorkers=1") == 1
        assert c["argv"][3] == selection["file"]
        assert len(selection["expandedTitles"]) == count
        outputs = [x.split("=", 1)[1] for x in c["argv"] if x.startswith("--outputFile.json=")]
        assert outputs == [str(RUN / (ident + ".json"))]
        for p in [RUN / (ident + ".json"), RUN / (ident + ".log"), RUN / (ident + ".processes.jsonl"),
                  RUN / (ident + ".result.json"), RUN / "tmp" / ident]:
            assert not p.exists(), "refusing to overwrite prior attempt: " + str(p)
        commands.append((c, selection["expandedTitles"]))
    assert not SUMMARY.exists(), "prior positive attempt exists; no automatic retry"
    RUN.mkdir(exist_ok=True)
    (RUN / "tmp").mkdir(exist_ok=True)
    summary = {"started": now(), "release": str(args.release), "releaseSha256": args.release_sha256,
               "sourceBefore": initial, "results": [], "completedCommands": 0,
               "plannedOnly": {"cases": 90, "http": 94, "ownerOpens": 39, "listeners": 19,
                               "directories": 20, "readonlyInventoryHandles": 17},
               "actualHttpOrOwnerCountersInstrumented": False,
               "note": "JSON cases are actual; planned resource counts are static source model, not observations.",
               "accepted": False, "notRun": list(IDS)}
    save(SUMMARY, summary)
    for index, (command, expected) in enumerate(commands):
        ident = command["id"]
        record = {"id": ident, "argv": command["argv"], "cwd": str(W), "started": now(),
                  "expectedCases": len(expected), "timeout": False, "errors": [], "signals": []}
        proc = None
        temp = RUN / "tmp" / ident
        log = RUN / (ident + ".log")
        processlog = RUN / (ident + ".processes.jsonl")
        started = time.monotonic()
        try:
            record["sourceBefore"] = verify()
            temp.mkdir()
            record["tmpdir"] = str(temp)
            record["tempBefore"] = entries(temp)
            env = os.environ.copy()
            env["TMPDIR"] = str(temp)
            env["PATH"] = str(Path(command["argv"][0]).parent) + os.pathsep + env.get("PATH", "")
            with log.open("xb") as stream, processlog.open("x") as snapshots:
                proc = subprocess.Popen(command["argv"], cwd=W, env=env,
                                        stdout=stream, stderr=subprocess.STDOUT, start_new_session=True)
                record["pid"] = record["pgid"] = proc.pid
                deadline = started + 120
                progress = 0
                while proc.poll() is None:
                    elapsed = time.monotonic() - started
                    snapshot = {"at": now(), "elapsedSeconds": elapsed,
                                "processes": group_snapshot(proc.pid), "tempEntries": entries(temp),
                                "logBytes": log.stat().st_size, "logMtime": log.stat().st_mtime}
                    snapshots.write(json.dumps(snapshot) + "\n")
                    snapshots.flush()
                    if elapsed >= progress:
                        print(json.dumps({"id": ident, "completedCommands": index, "totalCommands": 4,
                                          "elapsedSeconds": elapsed, "logBytes": snapshot["logBytes"],
                                          "ownedProcesses": len(snapshot["processes"])}), flush=True)
                        progress += 30
                    if time.monotonic() >= deadline:
                        record["timeout"] = True
                        record["errors"].append("command exceeded 120-second execution deadline")
                        break
                    time.sleep(min(1, max(0, deadline - time.monotonic())))
            if proc.poll() is not None:
                record["exitCode"] = proc.returncode
            else:
                record["exitCode"] = None
            if record["exitCode"] != 0:
                record["errors"].append("command exit was not zero")
        except BaseException as error:
            record["errors"].append(type(error).__name__ + ": " + str(error))
        finally:
            # Never mask original execution/runner errors with cleanup or binding errors.
            if proc is not None:
                try:
                    record["processesAfterCommand"] = group_snapshot(proc.pid)
                    if record["processesAfterCommand"]:
                        record["errors"].append("owned process group remained after command/timeout")
                        stop_owned(proc, record)
                    else:
                        proc.wait(timeout=2)
                    record["finalExitCode"] = proc.returncode
                    record["processesAfterCleanup"] = group_snapshot(proc.pid)
                    if record["processesAfterCleanup"]:
                        record["errors"].append("owned process group still present after cleanup")
                except BaseException as error:
                    record["errors"].append("cleanup: " + type(error).__name__ + ": " + str(error))
            try:
                record["sourceAfter"] = verify()
            except BaseException as error:
                record["errors"].append("binding: " + type(error).__name__ + ": " + str(error))
            if temp.exists():
                record["tempAfter"] = entries(temp)
                record["fixtureResiduals"] = [x for x in record["tempAfter"] if x.startswith(FIXTURE_PREFIXES)]
                record["cacheResidualsRetained"] = [x for x in record["tempAfter"] if x.startswith(CACHE_PREFIXES)]
                record["unknownResidualsRetained"] = [x for x in record["tempAfter"]
                    if x not in record["fixtureResiduals"] + record["cacheResidualsRetained"]]
                if record["fixtureResiduals"] or record["unknownResidualsRetained"]:
                    record["errors"].append("owned fixture or unexplained temporary entries remain; preserved, not deleted")
            result = RUN / (ident + ".json")
            if result.exists():
                try:
                    record["testResult"] = classify_result(result, expected)
                    if not record["testResult"]["accepted"]:
                        record["errors"].append("actual passed titles/counts differ from exact selected inventory")
                except BaseException as error:
                    record["errors"].append("result: " + type(error).__name__ + ": " + str(error))
            else:
                record["errors"].append("runner did not produce JSON result")
            record["ended"] = now()
            record["elapsedSeconds"] = time.monotonic() - started
            if log.exists():
                record["logSha256"] = sha(log)
            record["accepted"] = not record["errors"] and record.get("exitCode") == 0 and not record["timeout"]
            save(RUN / (ident + ".result.json"), record)
            summary["results"].append(record)
            summary["notRun"] = IDS[index + 1:]
            summary["completedCommands"] = len(summary["results"])
            save(SUMMARY, summary)
        if not record["accepted"]:
            summary["stoppedAt"] = ident
            break
    summary["ended"] = now()
    summary["accepted"] = len(summary["results"]) == 4 and all(r["accepted"] for r in summary["results"])
    summary["actualPassedCases"] = sum(r.get("testResult", {}).get("passedCount", 0) for r in summary["results"])
    save(SUMMARY, summary)
    print(json.dumps({"summary": str(SUMMARY), "accepted": summary["accepted"],
                      "actualPassedCases": summary["actualPassedCases"], "notRun": summary["notRun"]}), flush=True)
    return 0 if summary["accepted"] else 1


if __name__ == "__main__":
    raise SystemExit(main())
