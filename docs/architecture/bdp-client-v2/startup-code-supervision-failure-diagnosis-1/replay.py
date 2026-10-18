"""Pure frozen-data diagnostics. Never imports or executes any supervisor/runner.
Input and output paths are fixed beneath this new diagnosis packet.
"""
from pathlib import Path
import collections
import datetime
import hashlib
import json
from zoneinfo import ZoneInfo

ROOT = Path(__file__).resolve().parent
MAXIMUM_BYTES = 64 * 1024 * 1024

def read(relative):
    p = ROOT / relative
    data = p.read_bytes()
    if len(data) > MAXIMUM_BYTES:
        raise ValueError("bounded frozen input exceeded")
    return data

def document(relative):
    return json.loads(read(relative))

def records(relative):
    return [json.loads(line) for line in read(relative).splitlines() if line.strip()]

origins = document("input-map.json")
for entry in origins:
    assert hashlib.sha256(read(entry["copy"])).hexdigest() == entry["sha256"]

summaries = {}
all_rows = []
replays = []
for name, case in document("cases.json").items():
    prefix = "inputs/" + name
    terminal = document(prefix + "/supervision/attempt/outer-exit.json")
    events = records(prefix + "/supervision/attempt/outer-progress.jsonl")
    process = document(prefix + "/runner/process.json")
    control = document(prefix + "/supervision/control.json")
    pid = process["pid"]
    started = datetime.datetime.fromisoformat(process["started"]).timestamp()
    epoch = datetime.datetime.fromisoformat(terminal["started"]).timestamp()
    static = {
        "pidTypeAndRecordPgid": type(pid) is int and process["pgid"] == pid,
        "argvMatchesControl": process["argv"] == control["reviewerArgv"],
        "recordFreshRelativeToOuterStart": started >= epoch - 1,
    }
    previous_auth = {}
    periodic = []
    for ordinal, event in enumerate(events, 1):
        for row in event.get("owned", []):
            all_rows.append({"case": name, "source": "outer-periodic", "eventOrdinal": ordinal,
                             "elapsedSeconds": event["elapsedSeconds"], "row": row})
            if row["pid"] != pid:
                continue
            # lstart has no zone: use the operator-supplied America/Los_Angeles
            # context. Original environment TZ is not separately logged.
            epoch_row = datetime.datetime.strptime(row["start"], "%a %b %d %H:%M:%S %Y").replace(tzinfo=ZoneInfo("America/Los_Angeles")).timestamp()
            auth = {"pid": pid, "pgid": row["pgid"], "start": row["start"],
                    "recordStarted": process["started"], "argv": process["argv"]}
            tuple_ok = row["pgid"] == pid and abs(epoch_row - started) < 2
            zombie = row.get("state", "").startswith("Z")
            branch = (previous_auth.get(pid) == auth if zombie and case["helperVersion"] >= 3
                      else Path(row["comm"]).name == "claude")
            ok = all(static.values()) and tuple_ok and branch
            result = {"case": name, "source": "outer-periodic", "eventOrdinal": ordinal,
                      "elapsedSeconds": event["elapsedSeconds"], "row": row,
                      "tupleMatches": tuple_ok, "startDeltaSeconds": epoch_row - started,
                      "branch": "bound-zombie" if zombie and case["helperVersion"] >= 3 else "executable",
                      "branchPass": branch, "passesSavedInputs": ok,
                      "runtimeEnvironmentTimezoneRecorded": False}
            periodic.append(result)
            replays.append(result)
            if ok and not zombie and case["helperVersion"] >= 3:
                previous_auth[pid] = auth
    progress = records(prefix + "/runner/progress.jsonl")
    raw = records(prefix + "/runner/stream.jsonl")
    results = [r for r in raw if r.get("type") == "result"]
    progress_records = [r for r in progress if r.get("kind") == "record"]
    assert [r["recordOrdinal"] for r in progress_records] == list(range(1, len(raw) + 1))
    assert len(results) == 1 and results[0].get("subtype") == "success" and results[0].get("is_error") is False
    runner_review_rows = []
    for ordinal, event in enumerate(progress, 1):
        for row in event.get("processes", []):
            all_rows.append({"case": name, "source": "runner-periodic", "eventOrdinal": ordinal,
                             "elapsedSeconds": event["elapsedSeconds"], "row": row})
            if row["pid"] == pid:
                runner_review_rows.append({"elapsedSeconds": event["elapsedSeconds"], "row": row,
                                          "executableNameMatches": Path(row["comm"]).name == "claude",
                                          "fullIdentityReplay": "UNKNOWN: runner sample omits state and lstart"})
    for row in terminal["observedOwned"]:
        all_rows.append({"case": name, "source": "outer-latest-per-pid-aggregate", "row": row})
    latest = next(r for r in terminal["observedOwned"] if r["pid"] == pid)
    latest_auth = {"pid": pid, "pgid": latest["pgid"], "start": latest["start"],
                   "recordStarted": process["started"], "argv": process["argv"]}
    stored_auth = {r["pid"]: r for r in terminal.get("authenticatedReviewers", [])}
    latest_branch = (stored_auth.get(pid) == latest_auth if case["helperVersion"] >= 3 and latest.get("state", "").startswith("Z")
                     else Path(latest["comm"]).name == "claude")
    latest_replay = {"row": latest, "recordStaticChecks": static,
                     "terminalStoredAuthenticationEqual": stored_auth.get(pid) == latest_auth,
                     "helperSpecificBranchPass": latest_branch,
                     "chronology": "Aggregate latest row, not identified as failure-triggering inventory"}
    summaries[name] = {
        "helperVersion": case["helperVersion"], "reviewerPid": pid,
        "recordStaticChecks": static, "outerSamples": len(periodic),
        "outerSampleFailuresAgainstFinalRecord": sum(not r["passesSavedInputs"] for r in periodic),
        "runnerReviewerRows": len(runner_review_rows),
        "runnerNonmatchingExecutableNames": sum(not r["executableNameMatches"] for r in runner_review_rows),
        "runnerRows": runner_review_rows, "latestAggregateReplay": latest_replay,
        "eventCounts": dict(collections.Counter(e["kind"] for e in events)),
        "lastPeriodicElapsed": max(e["elapsedSeconds"] for e in events if e["kind"] == "process_sample"),
        "signals": [e for e in events if e["kind"] in ["signal", "signal_pass"]],
        "rawRecords": len(raw), "resultCount": len(results), "resultTextSha256": hashlib.sha256(results[0]["result"].encode()).hexdigest(),
        "lastResultProgressElapsed": next(e["elapsedSeconds"] for e in reversed(progress_records) if e.get("type") == "result"),
        "allRecordedRoleChecks": len(terminal["roleChecks"]),
        "allRecordedRoleChecksZero": all(r.get("exitCode") == 0 for r in terminal["roleChecks"]),
        "classification": terminal["classification"], "wrapperExitCode": terminal["wrapperExitCode"],
        "elapsedSeconds": terminal["elapsedSeconds"], "errors": terminal["errors"],
        "ownershipUnknown": terminal["ownershipUnknown"], "finalRoleVerified": terminal["finalRoleVerified"],
        "absentRunnerOutputs": case["absentRunnerOutputs"],
        "firstFailureInventoryAndPredicateRecorded": False,
    }
assert summaries["startup"]["outerSampleFailuresAgainstFinalRecord"] == 0
assert summaries["d1-code"]["outerSampleFailuresAgainstFinalRecord"] == 0
assert summaries["d1-plan"]["outerSampleFailuresAgainstFinalRecord"] == 0
assert summaries["startup"]["latestAggregateReplay"]["helperSpecificBranchPass"] is True
assert summaries["d1-code"]["latestAggregateReplay"]["helperSpecificBranchPass"] is True
assert summaries["d1-plan"]["latestAggregateReplay"]["helperSpecificBranchPass"] is False
for target, value in [("offline-replay.json", summaries), ("all-recorded-ps-rows.json", all_rows),
                      ("outer-reviewer-predicate-table.json", replays)]:
    (ROOT / target).write_text(json.dumps(value, indent=2) + "\n")
print(json.dumps({"classification": "REPLAY_COMPLETE_CAUSE_UNPROVEN", "cases": len(summaries),
                  "allSavedOuterReviewerSamplesPassFinalRecord": True,
                  "allRecordedPsRows": len(all_rows), "outerReviewerRows": len(replays),
                  "currentHelpersExecuted": False}))
