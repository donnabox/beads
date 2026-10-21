"""Draft one-command supervisor. Usage: python3 supervise.py COMMAND_ID.
No retries, no control/pack/install support yet. Root gate is mandatory.
"""
import datetime
import json
import os
import signal
import subprocess
import sys
import time
from pathlib import Path
from gate import P, W, E, load_gate, require, source_hashes, sha, git

def main():
    require(len(sys.argv) == 2, "Exactly one command ID required")
    name = sys.argv[1]
    gate, spec = load_gate(name)
    E.mkdir(exist_ok=True)
    stem = E / name
    for suffix in (".started.json", ".exit.json", ".log", ".processes.jsonl", ".results.json", ".tmp"):
        require(not Path(str(stem) + suffix).exists(), "Attempt already exists; preserve it, do not retry")
    env = os.environ.copy()
    for key in list(env):
        if key.startswith("BDP_") or key in ("NODE_OPTIONS", "NODE_PATH"):
            env.pop(key)
    env["PATH"] = "/Users/dbox/.nvm/versions/node/v24.16.0/bin:" + env.get("PATH", "")
    env["CI"] = "true"
    tmp = Path(str(stem) + ".tmp")
    env["TMPDIR"] = str(tmp)
    started = datetime.datetime.now(datetime.timezone.utc).isoformat()
    start = {
        "name": name, "argv": spec["argv"], "cwd": spec["cwd"], "started": started,
        "head": git("rev-parse", "HEAD").decode().strip(),
        "workspaceDirty": bool(git("status", "--porcelain").strip()),
        "sourceSha256": gate["sourceSha256"], "sourceDiffSha256": gate["sourceDiffSha256"],
        "rootAcceptanceSha256": gate["rootAcceptanceSha256"],
        "preparationSha256": gate["preparationSha256"],
        "bdpOverrides": {}, "sourceAttribution": "Measured dirty draft bytes; no installed commit override",
        "plannedWorkload": spec.get("workload", {}),
    }
    # Exclusive start marker survives even spawn/setup failures.
    with Path(str(stem) + ".started.json").open("x") as f:
        f.write(json.dumps(start, indent=2) + "\n")
    begin = time.monotonic()
    proc = None
    code = 1
    timed = forced = False
    errors = []
    seen_pids = set()
    seen_paths = set()

    def snapshot():
        try:
            # comm contains only executable identity, not command arguments.
            lines = subprocess.check_output(
                ["ps", "-axo", "pid=,ppid=,pgid=,rss=,etime=,state=,comm="], text=True
            ).splitlines()
            return [line for line in lines if len(line.split(None, 6)) == 7]
        except Exception as exc:
            errors.append("process-inventory:" + type(exc).__name__)
            return []

    def group(rows):
        return [line for line in rows if proc is not None and line.split(None, 6)[2] == str(proc.pid)]

    def entries():
        try:
            return {str(item) for item in tmp.iterdir()} if tmp.exists() else set()
        except Exception as exc:
            errors.append("temp-inventory:" + type(exc).__name__)
            return set()

    def stop_group():
        nonlocal forced
        if proc is None:
            return
        forced = True
        try:
            os.killpg(proc.pid, signal.SIGTERM)
        except ProcessLookupError:
            return
        end = time.monotonic() + 5
        while time.monotonic() < end and group(snapshot()):
            time.sleep(0.05)
        if group(snapshot()):
            try:
                os.killpg(proc.pid, signal.SIGKILL)
            except ProcessLookupError:
                pass

    def interrupt(_signum, _frame):
        raise KeyboardInterrupt("Supervisor interrupted")

    old_signals = {sig: signal.signal(sig, interrupt) for sig in (signal.SIGTERM, signal.SIGINT)}
    try:
        tmp.mkdir(exist_ok=False)
        with Path(str(stem) + ".log").open("x") as log, Path(str(stem) + ".processes.jsonl").open("x") as samples:
            proc = subprocess.Popen(spec["argv"], cwd=spec["cwd"], env=env,
                                    stdout=log, stderr=subprocess.STDOUT, start_new_session=True)
            seen_pids.add(proc.pid)
            next_progress = 30
            while proc.poll() is None:
                elapsed = time.monotonic() - begin
                rows = group(snapshot())
                seen_pids.update(int(row.split(None, 6)[0]) for row in rows)
                seen_paths.update(entries())
                samples.write(json.dumps({"elapsedSeconds": elapsed, "rows": rows}) + "\n")
                samples.flush()
                if elapsed >= next_progress:
                    print(json.dumps({"command": name, "elapsedSeconds": elapsed,
                                      "logBytes": Path(str(stem) + ".log").stat().st_size,
                                      "sampledGroupMembers": len(rows)}), flush=True)
                    next_progress += 30
                if elapsed >= 120:
                    timed = True
                    stop_group()
                    break
                time.sleep(0.05)
            code = proc.wait(timeout=5)
    except BaseException as exc:
        errors.append("execution:" + type(exc).__name__)
        stop_group()
        if proc is not None:
            try:
                code = proc.wait(timeout=5)
            except Exception as wait_error:
                errors.append("reap:" + type(wait_error).__name__)
    finally:
        # Do not permit a second interrupt to skip retained cleanup evidence.
        for sig in old_signals:
            signal.signal(sig, signal.SIG_IGN)
        before = group(snapshot())
        if before:
            stop_group()
        after = group(snapshot())
        live = {int(row.split(None, 6)[0]): row for row in snapshot()}
        still_pids = {str(pid): live[pid] for pid in seen_pids if pid in live}
        final_entries = entries()
        seen_paths.update(final_entries)
        present = {item for item in seen_paths if Path(item).exists()} | final_entries
        prefixes = spec.get("ownedTempPrefixes", [])
        runtime_dirs = sorted(item for item in present if Path(item).name.startswith(tuple(prefixes))) if prefixes else []
        unknown = sorted(present - set(runtime_dirs))
        try:
            post_hashes = source_hashes()
            source_unchanged = (post_hashes == gate["sourceSha256"] and
                                sha(git("diff", "--binary", "HEAD")) == gate["sourceDiffSha256"] and
                                git("rev-parse", "HEAD").decode().strip() == start["head"] and
                                not git("ls-files", "--others", "--exclude-standard").strip())
        except Exception as exc:
            errors.append("source-postcheck:" + type(exc).__name__)
            post_hashes = {}
            source_unchanged = False
        supervisor_exit = 124 if timed else 1 if (
            forced or errors or after or still_pids or runtime_dirs or not source_unchanged
        ) else (code if code >= 0 else 1)
        result = {
            "name": name, "argv": spec["argv"], "cwd": spec["cwd"], "started": started,
            "elapsedSeconds": time.monotonic() - begin, "exit": code,
            "supervisorExit": supervisor_exit, "timeout": timed, "forcedCleanup": forced,
            "errors": errors, "processGroup": proc.pid if proc else None,
            "remainingBeforeCleanup": before, "remainingAfterCleanup": after,
            "sampledPids": sorted(seen_pids), "sampledPidsStillPresent": still_pids,
            "sampledTempEntries": sorted(seen_paths), "tempEntriesAfter": sorted(final_entries),
            "remainingTempEntriesForClassification": sorted(present),
            "runtimeDirectoriesStillPresent": runtime_dirs,
            "unclassifiedTempEntriesStillPresent": unknown,
            "sourceUnchanged": source_unchanged, "sourceSha256After": post_hashes,
            "samplingSeconds": 0.05,
            "samplingLimitation": "Short-lived or detached descendants may evade group sampling; helper counts and independent final checks remain separate.",
            "tempClassification": "Only the selected fixture prefixes under this attempt TMPDIR are owned runtime roots; all other entries are preserved for review, not deleted or called leaks.",
            "resultLimitation": "Exit0 is command completion, not independent oracle/claim/provenance adjudication.",
        }
        with Path(str(stem) + ".exit.json").open("x") as f:
            f.write(json.dumps(result, indent=2) + "\n")
        print(json.dumps({"command": name, "supervisorExit": supervisor_exit,
                          "elapsedSeconds": result["elapsedSeconds"], "result": str(stem) + ".exit.json"}), flush=True)
        for sig, previous in old_signals.items():
            signal.signal(sig, previous)
    return supervisor_exit

if __name__ == "__main__":
    sys.exit(main())
