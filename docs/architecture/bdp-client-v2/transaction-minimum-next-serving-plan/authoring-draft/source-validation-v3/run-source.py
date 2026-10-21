"""Closed one-command minimum-current-resource supervisor, adapted from reviewed receipt source supervisor.
Root supplies execution-gate.json; this file does not create/release that gate.
Usage: python3 run-source.py COMMAND_ID. Never retries or invokes controls/packages.
"""
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
D = P.parent
M = D.parent
W = Path("/Users/dbox/repos/wt/vickie/vickie-minimum-current-resource")
E = P / "run"
L = Path("/Users/dbox/repos/wt/vickie/bdp-integration-logs")
GUARD = L / "guarded-action.py"
BINDING = None  # Root must bind the final combined author manifest.
COMMANDS = P / "commands.json"
BINDING_SHA = None  # Actual frozen manifest must be pinned by root gate, never guessed.
COMMANDS_SHA = 'b6635fd16dd899bee6de96b49d7ba85b759c7c779c7c7f315f4a752ea06ff0c9'
GUARD_SHA = 'fd81ba029cd4e7c7e60623df89e8292af91141d311d9ce789ed19e2152375e84'
HEAD = '1647c5df37a82e1364d281ac46beb60a94ab6c76'
TREE = 'a070420a4c17aa67a821c742d363ff9e82e872a6'
SESSION = "2026-10-08T20:26:35Z-81e5"
ORDER = ['setup', 'typecheck', 'build', 'format-check', 'lint', 'boundaries', 'read-seal', 'consistency', 'owner-read', 'http', 'build-inventory']
OWNED_PATHS = None  # Root must freeze the exact combined changed-path inventory.

def require(ok, message):
    if not ok:
        raise RuntimeError(message)

def sha(value):
    return hashlib.sha256(value).hexdigest()

def git(*args):
    return subprocess.check_output(["git", *args], cwd=W, timeout=10)

def source_hashes():
    require(sha(BINDING.read_bytes()) == BINDING_SHA, "Successor source binding changed")
    binding = json.loads(BINDING.read_text())
    require(set(binding["files"]) == set(OWNED_PATHS), "Combined frozen path inventory changed")
    require(binding["head"] == HEAD and binding["tree"] == TREE, "Manifest base mismatch")
    require(binding["branch"] == 'codex/vickie-minimum-current-resource', "Unexpected draft branch")
    require(binding["workspace"] == str(W), "Manifest workspace mismatch")
    require(git("rev-parse", "HEAD").decode().strip() == HEAD, "HEAD changed")
    require(git("rev-parse", "HEAD^{tree}").decode().strip() == TREE, "Base tree changed")
    require(git("branch", "--show-current").decode().strip() == binding["branch"], "Branch changed")
    require(not git("diff", "--cached", "--name-only").strip(), "Unexpected staged changes")
    expected_new = {name for name, item in binding["files"].items() if item["new"]}
    actual_new = set(git("ls-files", "--others", "--exclude-standard").decode().splitlines())
    require(actual_new == expected_new, "Unexpected untracked source")
    changed = set(git("diff", "HEAD", "--name-only").decode().splitlines()) | actual_new
    require(changed == set(binding["files"]), "Unexpected source scope")
    actual = {name: sha((W / name).read_bytes()) for name in binding["files"]}
    require(actual == {name: item["sha256"] for name, item in binding["files"].items()}, "Frozen source changed")
    pinned = json.loads((P / "input-binding.json").read_text())
    require({p: sha((W/p).read_bytes()) for p in pinned} == pinned, "Tracked build/source dependency changed")
    return actual

def source_patch_hash():
    binding = json.loads(BINDING.read_text())
    patch = git("diff", "--binary", "HEAD")
    for name, item in binding["files"].items():
        if item["new"]:
            result = subprocess.run(["git", "diff", "--no-index", "--", "/dev/null", name],
                                    cwd=W, capture_output=True, timeout=10)
            require(result.returncode == 1, "New source diff unavailable")
            patch += result.stdout
    return sha(patch)

def load_gate(name):
    global BINDING_SHA, BINDING, OWNED_PATHS
    require(name in ORDER, "Control, package and installed stages are separately closed")
    gate = json.loads((P / "execution-gate.json").read_text())
    require(gate["status"] == "released" and gate["session"] == SESSION, "Root source gate closed")
    BINDING_SHA = gate["sourceBindingSha256"]
    BINDING = Path(gate["sourceBindingPath"])
    require(BINDING.is_absolute() and BINDING.resolve().is_relative_to(D.resolve()), "Combined binding must be an actual plan artifact")
    OWNED_PATHS = gate["ownedPaths"]
    allowed = json.loads((M / "review-successor/root-authoring-release.json").read_text())
    require(isinstance(OWNED_PATHS, list) and len(OWNED_PATHS) == len(set(OWNED_PATHS)) and bool(OWNED_PATHS), "Exact combined paths unavailable")
    require(set(OWNED_PATHS) <= {'packages/server/src/transaction-read.test.ts', 'docs/specs/bdp.md', 'apps/bdptest/src/transactional-minimum-current-resource-installed.test.ts', 'packages/server/src/transaction-http.ts', 'packages/server/src/transaction-consistency.ts', 'packages/server/src/transaction-http.test.ts', 'packages/server/src/transaction-consistency.test.ts', 'packages/server/src/transaction-service.ts', 'docs/design/transaction-implementation-progress.md', 'packages/server/src/transaction-read.ts'}, "Source scope exceeds authoring release")
    require(sha((M / "review-successor/root-authoring-release.json").read_bytes()) == gate["authoringReleaseSha256"], "Authoring scope changed")
    require(isinstance(BINDING_SHA, str) and len(BINDING_SHA) == 64 and all(c in "0123456789abcdef" for c in BINDING_SHA), "Actual frozen source manifest is not bound")
    require(gate["controlGate"] == "closed" and gate["installedGate"] == "closed", "Unexpected broader release")
    require(gate["supervisorSha256"] == sha(Path(__file__).read_bytes()), "Supervisor binding changed")
    require(gate["sourceBindingSha256"] == BINDING_SHA and gate["commandsSha256"] == COMMANDS_SHA, "Gate binding mismatch")
    require(gate["expectedHead"] == HEAD and gate["expectedTree"] == TREE, "Gate source identity mismatch")
    require(sha(COMMANDS.read_bytes()) == COMMANDS_SHA and sha(GUARD.read_bytes()) == GUARD_SHA, "Commands or session guard changed")
    authorized = gate["authorizedStages"]
    require(authorized == ORDER[:len(authorized)] and name in authorized, "Stage not authorized")
    acceptance = gate["rootAcceptance"]
    authority = Path(acceptance["path"])
    require(authority.is_relative_to(L) and sha(authority.read_bytes()) == acceptance["sha256"], "Root acceptance missing or changed")
    require(gate["inputBindingSha256"] == sha((P/"input-binding.json").read_bytes()), "Input inventory changed")
    for item, digest in gate["preparationFiles"].items():
        require(sha((P/item).read_bytes()) == digest, "Preparation changed: " + item)
    require(gate.get("unmergedCandidateValidation") is True, "Unmerged source attribution must be explicit")
    source = source_hashes()
    binding = json.loads(BINDING.read_text())
    require(source_patch_hash() == binding["diffSha256"], "Full draft diff changed")
    for previous in ORDER[:ORDER.index(name)]:
        result = json.loads((E / (previous + ".exit.json")).read_text())
        require(result["supervisorExit"] == 0 and result["sourceUnchanged"], "Prior command failed or unfinished")
        require(result["sourceSha256After"] == source, "Prior command has different source")
    matches = [item for item in json.loads(COMMANDS.read_text())["commands"] if item["id"] == name]
    require(len(matches) == 1, "Command identity ambiguous")
    spec = dict(matches[0])
    require(spec["cwd"] == str(W) and spec["supervisorSeconds"] == 120, "Unexpected cwd/bound")
    if name == "build-inventory":
        require(spec["argv"] == ["/usr/bin/python3", str(P/"inventory.py")], "Inventory command changed")
    else:
        expected_tool = str(W / "node_modules/vitest/vitest.mjs") if "selectedFullNames" in spec else "/Users/dbox/.cache/node/corepack/v1/pnpm/11.20.0/bin/pnpm.cjs"
        require(spec["argv"][:2] == ["/Users/dbox/.nvm/versions/node/v24.16.0/bin/node", expected_tool], "Toolchain changed")
    require(not spec.get("validatedBdpOverrides"), "Source phase must not inherit installed/control overrides")
    spec.setdefault("ownedTempPrefixes", [])
    spec["workload"] = {key: spec[key] for key in ["cases", "http", "owners", "directories", "listeners", "productChildren", "readonlyInventoryHandles", "otherInventoryHandles", "rejectedOpenDirectories"] if key in spec}
    for argument in spec["argv"]:
        if argument.startswith("--outputFile="):
            require(not Path(argument.split("=", 1)[1]).exists(), "Test result already exists; no retry")
    gate = {**gate, "sourceSha256": source, "sourceDiffSha256": binding["diffSha256"],
            "preparationSha256": sha(Path(__file__).read_bytes())}
    return gate, spec

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
    for suffix in (".started.json", ".exit.json", ".log", ".stderr.log", ".processes.jsonl", ".results.json", ".tmp"):
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
        "buildInputFiles": json.loads((P/"input-binding.json").read_text()) if name == "build" else None,
        "inputBindingSha256": sha((P/"input-binding.json").read_bytes()),
        "toolSha256": sha(Path(spec["argv"][0]).read_bytes()),
        "toolEntrySha256": sha(Path(spec["argv"][1]).read_bytes()),
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
        with Path(str(stem) + ".log").open("x") as log, Path(str(stem) + ".stderr.log").open("x") as stderr_log, Path(str(stem) + ".processes.jsonl").open("x") as samples:
            proc = subprocess.Popen([sys.executable, str(GUARD), *spec["argv"]], cwd=spec["cwd"], env=env,
                                    stdout=log, stderr=stderr_log, start_new_session=True)
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
                                      "stderrBytes": Path(str(stem) + ".stderr.log").stat().st_size,
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
            errors.append("source-postcheck:" + type(exc).__name__ + ":" + str(exc))
            post_hashes = {}
            source_unchanged = False
        adjudication = None
        if code == 0 and not timed and not forced and not errors:
            try:
                adjudication = adjudicate_source(spec)
            except Exception as exc:
                errors.append("result-adjudication:" + type(exc).__name__ + ":" + str(exc))
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
