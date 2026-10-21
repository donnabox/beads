#!/usr/bin/env python3
"""Two exact emitted-only controls. Requires separate root release; no build or packaging."""
import argparse, importlib.util, json, os, re, subprocess, time
from pathlib import Path
C = Path(__file__).resolve().parent
E = C.parent
COMMANDS_SHA = "cda1c6825cbfd5a7bf14a2ec3f8df92d8197c659e67ec2b2bf94f8e5b8468c7d"
POSITIVE_MANIFEST_SHA = "0f681290210f871871dcc62c0834377665110ad96f39665d27919bde27d9c523"
HELPER_SHA = "1036f3dfccfebfdd896b47f21054f96cccf2d941a1b20dc730011a57fccf39c6"
# Reuse only the reviewed supervisor's pure hashing/process/cleanup functions.
spec = importlib.util.spec_from_file_location("selector_positive_helpers", E / "run-positive.py")
base = importlib.util.module_from_spec(spec)
assert __import__("hashlib").sha256((E / "run-positive.py").read_bytes()).hexdigest() == HELPER_SHA
spec.loader.exec_module(base)
RUN = C / "run"

def verify():
    base.verify()
    assert base.sha(C / "commands.json") == COMMANDS_SHA
    assert base.sha(C / "positive-files.json") == POSITIVE_MANIFEST_SHA
    manifest = json.loads((C / "positive-files.json").read_text())
    config = json.loads((C / "commands.json").read_text())
    assert [r["id"] for r in config["controls"]] == ["grammar-development", "omit-fresh-capacity"]
    for relative, expected in manifest["files"].items():
        assert base.sha(base.W / relative) == expected, "positive graph drift: " + relative
    for row in config["controls"]:
        root = Path(row["workspace"])
        assert root == C / row["id"] / "workspace"
        for relative, expected in manifest["files"].items():
            if relative == row["emittedPath"]:
                expected = row["afterSha256"]
            assert base.sha(root / relative) == expected, "derivative drift: " + relative
        assert base.sha(C / row["id"] / "emitted.patch") == row["patchSha256"]
        assert base.sha(C / row["id"] / "module-binding.json") == row["moduleBindingSha256"]
        assert (root / "packages/server/node_modules/@bdp/protocol").resolve() == root / "packages/protocol"
        assert row["argv"].count("--maxWorkers=1") == 1
        assert row["cwd"] == str(root) and row["timeoutMs"] == 120000
    return config

def adjudicate(path, row):
    data = json.loads(path.read_text())
    assertions = [a for result in data["testResults"] for a in result["assertionResults"]]
    failed = [a for a in assertions if a["status"] == "failed"]
    passed = [a for a in assertions if a["status"] == "passed"]
    chosen = [a for a in assertions if a["fullName"] == row["selectedTitle"]]
    messages = "\n".join(str(m) for a in failed for m in a.get("failureMessages", []))
    plain = re.sub(r"\x1b\[[0-9;]*m", "", messages)
    observations = []
    decoder = json.JSONDecoder()
    for match in re.finditer(r"\{", plain):
        try:
            value, _ = decoder.raw_decode(plain[match.start():])
            if isinstance(value, dict) and value.get("oracle") == "Selector receiving status":
                observations.append(value)
        except ValueError:
            pass
    expected = {"oracle": "Selector receiving status", "status": 500,
                "developmentLimit": row["expectedDevelopmentLimit"], "contentLength": "0"}
    expected_assertion = "expected 500 to be " + str(row["expectedAssertionStatus"])
    accepted = (len(failed) == len(chosen) == 1 and not passed
                and failed[0]["fullName"] == row["selectedTitle"]
                and chosen[0]["status"] == "failed" and expected in observations
                and expected_assertion in plain and data.get("success") is False)
    return {"accepted": accepted, "runnerJsonSha256": base.sha(path),
            "failedTitles": [a["fullName"] for a in failed], "passedTitles": [a["fullName"] for a in passed],
            "unselectedSkippedCount": sum(a["status"] in ("pending", "skipped", "todo", "disabled") for a in assertions),
            "actualFirstStatusDiagnostics": observations, "expectedDiagnostic": expected,
            "expectedAssertionText": expected_assertion, "failureMessages": messages,
            "actualHTTPDispatchCounter": None,
            "requestCountInference": "Intended named first assertion follows the first sequential request; expected1, not an instrumented dispatch count.",
            "staticMaximumHTTP": row["maxHttp"]}

def main():
    p = argparse.ArgumentParser()
    p.add_argument("--release", type=Path, required=True)
    p.add_argument("--release-sha256", required=True)
    args = p.parse_args()
    assert base.sha(args.release) == args.release_sha256
    release = json.loads(args.release.read_text())
    assert release["scope"] == "selector-emitted-controls-only"
    assert release["scriptSha256"] == base.sha(Path(__file__))
    assert release["commandsSha256"] == COMMANDS_SHA
    assert release["positiveFilesSha256"] == POSITIVE_MANIFEST_SHA
    config = verify()
    assert not RUN.exists(), "a control attempt already exists; no automatic retry/overwrite"
    RUN.mkdir()
    summary = {"started": base.now(), "results": [], "accepted": False,
               "releaseSha256": args.release_sha256, "commandsSha256": COMMANDS_SHA,
               "notRun": [r["id"] for r in config["controls"]],
               "staticBounds": {"expectedHTTP": 2, "maxHTTP": 52, "owners": 2, "listeners": 2,
                                "directories": 2, "expectedReadonly": 0, "maxReadonly": 2},
               "actualResourceCountersInstrumented": False}
    base.save(RUN / "summary.json", summary)
    for index, row in enumerate(config["controls"]):
        ident = row["id"]
        result = {"id": ident, "started": base.now(), "argv": row["argv"], "cwd": row["cwd"],
                  "timeout": False, "errors": [], "signals": []}
        proc = None
        tmp = RUN / (ident + "-tmp")
        log = RUN / (ident + ".log")
        started = time.monotonic()
        try:
            verify()
            tmp.mkdir()
            result["tempBefore"] = base.entries(tmp)
            env = os.environ.copy()
            env["TMPDIR"] = str(tmp)
            env["PATH"] = str(Path(row["argv"][0]).parent) + os.pathsep + env.get("PATH", "")
            with log.open("xb") as output, (RUN / (ident + ".processes.jsonl")).open("x") as monitor:
                proc = subprocess.Popen(row["argv"], cwd=row["cwd"], env=env, stdout=output,
                                        stderr=subprocess.STDOUT, start_new_session=True)
                result["pid"] = result["pgid"] = proc.pid
                deadline = started + 120
                next_progress = 0
                while proc.poll() is None:
                    elapsed = time.monotonic() - started
                    snapshot = {"at": base.now(), "elapsedSeconds": elapsed,
                                "processes": base.group_snapshot(proc.pid), "tempEntries": base.entries(tmp),
                                "logBytes": log.stat().st_size, "logMtime": log.stat().st_mtime}
                    monitor.write(json.dumps(snapshot) + "\n"); monitor.flush()
                    if elapsed >= next_progress:
                        print(json.dumps({"id": ident, "elapsedSeconds": elapsed,
                                          "logBytes": snapshot["logBytes"], "completedControls": index}), flush=True)
                        next_progress += 30
                    if time.monotonic() >= deadline:
                        result["timeout"] = True
                        result["errors"].append("control exceeded120-second execution deadline")
                        break
                    time.sleep(min(1, max(0, deadline - time.monotonic())))
            result["exitCode"] = proc.poll()
            if result["exitCode"] != 1:
                result["errors"].append("expected test assertion exit1, not pass/setup/timeout outcome")
        except BaseException as error:
            result["errors"].append(type(error).__name__ + ": " + str(error))
        finally:
            if proc is not None:
                try:
                    result["processesAfterCommand"] = base.group_snapshot(proc.pid)
                    if result["processesAfterCommand"]:
                        result["errors"].append("owned process group remained; cleanup required")
                        base.stop_owned(proc, result)
                    else:
                        proc.wait(timeout=2)
                    result["processesAfterCleanup"] = base.group_snapshot(proc.pid)
                    if result["processesAfterCleanup"]:
                        result["errors"].append("owned group survives cleanup")
                except BaseException as error:
                    result["errors"].append("cleanup: " + type(error).__name__ + ": " + str(error))
            try:
                verify()
                result["allBindingsPreservedAfter"] = True
            except BaseException as error:
                result["errors"].append("binding: " + type(error).__name__ + ": " + str(error))
            if tmp.exists():
                result["tempAfter"] = base.entries(tmp)
                result["cacheResidualsRetained"] = [n for n in result["tempAfter"] if n.startswith(base.CACHE_PREFIXES)]
                result["otherResidualsRetained"] = [n for n in result["tempAfter"] if n not in result["cacheResidualsRetained"]]
                if result["otherResidualsRetained"]:
                    result["errors"].append("fixture/unknown temp entries remain; retained without deletion")
            json_path = RUN / (ident + ".json")
            if json_path.exists():
                try:
                    result["adjudication"] = adjudicate(json_path, row)
                    if not result["adjudication"]["accepted"]:
                        result["errors"].append("missing intended named first-status failure/actual500 diagnostic")
                except BaseException as error:
                    result["errors"].append("adjudication: " + type(error).__name__ + ": " + str(error))
            else:
                result["errors"].append("runner JSON absent")
            result["ended"] = base.now()
            result["elapsedSeconds"] = time.monotonic() - started
            result["accepted"] = not result["errors"] and result.get("exitCode") == 1 and not result["timeout"]
            if log.exists(): result["logSha256"] = base.sha(log)
            base.save(RUN / (ident + ".result.json"), result)
            summary["results"].append(result)
            summary["notRun"] = [r["id"] for r in config["controls"][index + 1:]]
            base.save(RUN / "summary.json", summary)
        if not result["accepted"]:
            summary["stoppedAt"] = ident
            break
    summary["ended"] = base.now()
    summary["accepted"] = len(summary["results"]) == 2 and all(r["accepted"] for r in summary["results"])
    base.save(RUN / "summary.json", summary)
    print(json.dumps({"accepted": summary["accepted"], "notRun": summary["notRun"], "summary": str(RUN / "summary.json")}), flush=True)
    return 0 if summary["accepted"] else 1

if __name__ == "__main__":
    raise SystemExit(main())
