"""Bounded source build/static/eight-CPU adaptation. No installed runtime release."""
import datetime
import hashlib
import json
import os
import signal
import subprocess
import sys
import time
from pathlib import Path

P = Path(__file__).resolve().parent
D = P
W = Path("/Users/dbox/repos/wt/vickie/vickie-update-cold-directory")
E = D / "source-run"
L = Path("/Users/dbox/repos/wt/vickie/bdp-integration-logs")
GUARD = L / "guarded-action.py"
BINDING = D / "source-binding.json"
COMMANDS = D / "commands.json"
BINDING_SHA = "cf670a46fd94495ea550878ef3e26d4d6e9e3487b04f058da2cdd2b58540114c"
COMMANDS_SHA = "9c8a3325a71f58f29fb178e274c035b1c5964f255f62aba6747ba9ef59cc2ac2"
ORDER = ['cumulative-callback']
def require(ok, message):
    if not ok: raise RuntimeError(message)
def sha(value): return hashlib.sha256(value).hexdigest()
def git(*args): return subprocess.check_output(["git", *args], cwd=W, timeout=10)
def source_hashes():
    require(sha(BINDING.read_bytes()) == BINDING_SHA, "Source binding changed")
    b=json.loads(BINDING.read_text())
    require(git("rev-parse","HEAD").decode().strip()==b["parent"], "HEAD changed")
    require(git("branch","--show-current").decode().strip()==b["branch"], "Branch changed")
    require(not git("diff","--cached","--name-only").strip(), "Staged source")
    require(not git("ls-files","--others","--exclude-standard").strip(), "Untracked source")
    require(set(git("diff","HEAD","--name-only").decode().splitlines())==set(b["sourceFiles"]), "Scope changed")
    actual={p:sha((W/p).read_bytes()) for p in b["sourceFiles"]}
    require(actual==b["sourceFiles"], "Frozen source changed")
    return actual
def source_patch_hash(): return sha(git("diff","--binary","HEAD"))
def load_gate(name):
    require(name in ORDER, "Only root-released source stages")
    require(sha(COMMANDS.read_bytes())==COMMANDS_SHA, "Commands changed")
    b=json.loads(BINDING.read_text()); source=source_hashes()
    require(source_patch_hash()==b["patchSha256"], "Patch changed")
    release=json.loads((D/"source-static-release.json").read_text())
    require(release["authorizedStages"]==ORDER and release["installedReleased"] is False, "Release scope changed")
    for previous in ORDER[:ORDER.index(name)]:
        r=json.loads((E/(previous+".exit.json")).read_text())
        require(r["supervisorExit"]==0 and r["sourceUnchanged"] and r["sourceSha256After"]==source, "Prior stage not clean")
    c=json.loads(COMMANDS.read_text()); spec=dict(next(x for x in c["stages"] if x["id"]==name))
    spec.update(cwd=str(W),ownedTempPrefixes=[])
    if name=="cumulative-callback":
        spec["selectedFullNames"]=[x["fullName"] for x in json.loads((D/"selection.json").read_text())["cpu"]]
    for arg in spec["argv"]:
        if arg.startswith("--outputFile="): require(not Path(arg.split("=",1)[1]).exists(),"No retry")
    return {"sourceSha256":source,"sourceDiffSha256":b["patchSha256"],"rootAcceptance":{"sha256":sha((D/"source-static-release.json").read_bytes())},"preparationSha256":sha(Path(__file__).read_bytes())},spec

def adjudicate_source(spec):
    if "selectedFullNames" not in spec:
        return {"kind": "command-only", "testVerdict": "not-applicable"}
    output = next(Path(arg.split("=", 1)[1]) for arg in spec["argv"] if arg.startswith("--outputFile="))
    report = json.loads(output.read_text())
    assertions = [case for file in report["testResults"] for case in file.get("assertionResults", [])]
    executed = [case for case in assertions if case["status"] not in ["pending", "skipped", "todo"]]
    passed = [case["fullName"] for case in executed if case["status"] == "passed"]
    require(sorted(passed) == sorted(spec["selectedFullNames"]) and len(executed) == len(passed), "Selected test names/count/status mismatch")
    require(report.get("success") is True and report.get("numFailedTests") == 0 and report.get("numPassedTests") == len(passed), "Test aggregate failed")
    require(not report.get("unhandledErrors") and not report.get("errors") and not report.get("numRuntimeErrorTestSuites", 0), "Reported unhandled errors")
    require(all(not file.get("message") and not file.get("failureMessage") for file in report["testResults"]), "Test file error")
    return {"kind": "exact-source-selection", "passedFullNames": passed,
            "skipped": len(assertions) - len(executed), "jsonSha256": sha(output.read_bytes()),
            "limitation": "Static HTTP/owner model is not an independently observed network quota; raw logs and reported cleanup remain separately reviewed."}

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
        if key.startswith("BDP_") or key in ("NODE_OPTIONS", "NODE_PATH", "PYTHONOPTIMIZE", "NODE_DEBUG", "NODE_DEBUG_NATIVE"):
            env.pop(key)
    env["PATH"] = "/Users/dbox/.nvm/versions/node/v24.16.0/bin:" + env.get("PATH", "")
    env["CI"] = "true"
    env.update(spec.get("validatedBdpOverrides", {}))
    tmp = Path(str(stem) + ".tmp")
    env["TMPDIR"] = str(tmp)
    started = datetime.datetime.now(datetime.timezone.utc).isoformat()
    start = {
        "name": name, "argv": spec["argv"], "cwd": spec["cwd"], "started": started,
        "head": git("rev-parse", "HEAD").decode().strip(),
        "workspaceDirty": bool(git("status", "--porcelain").strip()),
        "sourceSha256": gate["sourceSha256"], "sourceDiffSha256": gate["sourceDiffSha256"],
        "rootAcceptanceSha256": gate["rootAcceptance"]["sha256"],
        "preparationSha256": gate["preparationSha256"],
        "bdpOverrides": spec.get("validatedBdpOverrides", {}),
        "guardedArgv": [sys.executable, str(GUARD), *spec["argv"]],
        "nodePathPrefix": "/Users/dbox/.nvm/versions/node/v24.16.0/bin",
        "sourceAttribution": "Externally bound source snapshot or frozen package; not child-attested Git identity",
        "plannedWorkload": spec.get("workload", {}),
        "derivativeBinding": gate.get("controls", {}).get(spec.get("control")),
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
        end = min(time.monotonic() + 5, begin + 120)
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
            proc = subprocess.Popen([sys.executable, str(GUARD), *spec["argv"]], cwd=spec["cwd"], env=env,
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
                if elapsed >= 115:
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
                                source_patch_hash() == gate["sourceDiffSha256"] and
                                git("rev-parse", "HEAD").decode().strip() == start["head"])
            require(sha(COMMANDS.read_bytes()) == COMMANDS_SHA, "Commands changed during run")
        except Exception as exc:
            errors.append("source-postcheck:" + type(exc).__name__)
            post_hashes = {}
            source_unchanged = False
        adjudication = None
        if code == 0 and not timed and not forced and not errors:
            try:
                adjudication = adjudicate_source(spec)
            except Exception as exc:
                errors.append("result-adjudication:" + type(exc).__name__)
        supervisor_exit = 124 if timed else 1 if (
            forced or errors or after or still_pids or runtime_dirs or not source_unchanged
        ) else (code if code >= 0 else 1)
        result = {
            "name": name, "argv": spec["argv"], "guardedArgv": [sys.executable, str(GUARD), *spec["argv"]],
            "cwd": spec["cwd"], "started": started, "adjudication": adjudication,
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
            "commandWindowSeconds": 115, "groupHardDeadlineSeconds": 120,
            "cleanupBound": "TERM at115s, KILL no later than120s; wait/reap after kill adds no product execution allowance",
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
