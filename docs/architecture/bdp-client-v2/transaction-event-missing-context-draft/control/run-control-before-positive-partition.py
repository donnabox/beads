"""One accepted emitted-control invocation. Gate absent/closed by default.
Reuses the reviewed source supervisor; preserves actual exit1 separately from
negative-control adjudication. No setup, build, package, extra test or retries.
"""
import hashlib
import importlib.util
import json
import os
from pathlib import Path
import re
import sys

P = Path(__file__).resolve().parent
D = P.parent
C = P / "copy"
W = Path("/Users/dbox/repos/wt/vickie/vickie-sequence-member-evidence")
BASE = D / "execution-preparation/run-source.py"
BASE_SHA = "4d765e2408033fb42c828cdebfdf59b2e2b1c16d154522b3e4e018d3e8c8689f"
NAME = "omit-context-control"
SESSION = "2026-10-08T20:26:35Z-81e5"
PACKET_SHA = "a003ed6fb2568097f9977b50be78279ba4603891689980d57cf7b877ec8892d7"
CONTROL_SHA = "ec640209f149619156e85cd860d58e9d336450ca433fa6be3009d2d433b929d4"

def require(ok, message):
    if not ok:
        raise RuntimeError(message)

def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()

def read(path):
    return json.loads(path.read_text())

def verify_graph():
    require(sha(P / "preparation-manifests.json") == PACKET_SHA, "Preparation binding changed")
    packet = read(P / "preparation-manifests.json")
    for name, digest in packet["files"].items():
        require(sha(P / name) == digest, "Prepared manifest changed: " + name)
    copy_manifest = read(P / "copy-manifest.json")
    require(sha(D / "run/build.exit.json") == copy_manifest["buildExitSha256"], "Positive build record changed")
    files = copy_manifest["files"]
    actual = set()
    for root, dirs, names in os.walk(C, followlinks=False):
        dirs[:] = [name for name in dirs if name != "node_modules"]
        for name in names:
            path = Path(root) / name
            relative = str(path.relative_to(C))
            require(not path.is_symlink() and relative in files, "Unexpected copied member")
            require(sha(path) == files[relative]["copySha256"], "Copied member changed: " + relative)
            actual.add(relative)
    require(actual == set(files), "Copied member inventory changed")
    for name, item in files.items():
        require(sha(W / name) == item["positiveSha256"], "Positive source/build changed: " + name)
    dependencies = read(P / "dependency-manifest.json")
    root = C / "node_modules"
    require(root.is_symlink() and os.readlink(root) == dependencies["rootLink"]["link"]
            and str(root.resolve(strict=True)) == dependencies["rootLink"]["realpath"], "Root tooling graph changed")
    store = Path(dependencies["sharedStore"])
    seen = set()
    for parent, dirs, names in os.walk(store, followlinks=False):
        for name in dirs + names:
            path = Path(parent) / name
            relative = str(path.relative_to(W))
            if not path.is_symlink() and not path.is_file():
                continue
            require(relative in dependencies["members"], "Shared dependency inventory changed")
            item = dependencies["members"][relative]
            if path.is_symlink():
                require(os.readlink(path) == item["link"]
                        and str(path.resolve(strict=True)) == item["realpath"], "Dependency link changed")
            else:
                require(sha(path) == item["sha256"], "Dependency bytes changed")
            seen.add(relative)
    require(seen == set(dependencies["members"]), "Shared dependency member missing")
    for name, item in dependencies["links"].items():
        positive, copied = W / name, C / name
        require(positive.is_symlink() and os.readlink(positive) == item["positiveLink"]
                and str(positive.resolve(strict=True)) == item["positiveRealpath"], "Positive package link changed")
        require(copied.is_symlink() and os.readlink(copied) == item["copyLink"]
                and str(copied.resolve(strict=True)) == item["copyRealpath"], "Copied package link changed")
    for item in dependencies["workspaceFallbacks"]:
        require(sha(Path(item["realpath"]) / "package.json") == item["packageSha256"],
                "Shared fallback package changed")
    return packet

def adjudicate(module):
    result = read(D / "run" / (NAME + ".exit.json"))
    require(result["exit"] == 1 and result["supervisorExit"] == 1, "Expected actual test failure absent")
    require(not result["timeout"] and not result["forcedCleanup"] and not result["errors"],
            "Control execution/cleanup error")
    for field in ("remainingAfterCleanup", "sampledPidsStillPresent", "runtimeDirectoriesStillPresent"):
        require(not result[field], "Control leaked owned resources: " + field)
    require(result["sourceUnchanged"], "Positive source changed")
    spec = next(x for x in read(D / "commands.json")["commands"] if x["id"] == NAME)
    output = next(Path(x.split("=", 1)[1]) for x in spec["argv"] if x.startswith("--outputFile="))
    report = read(output)
    executed = [case for file in report["testResults"] for case in file.get("assertionResults", [])
                if case["status"] not in ("pending", "skipped", "todo")]
    require(len(executed) == 1 and executed[0]["status"] == "failed"
            and executed[0]["fullName"] == spec["selectedFullNames"][0], "Negative case selection changed")
    require(report["success"] is False and report["numFailedTests"] == 1
            and report["numPassedTests"] == 0, "Negative aggregate mismatch")
    require(not report.get("unhandledErrors") and not report.get("errors")
            and not report.get("numRuntimeErrorTestSuites", 0), "Unhandled control error")
    contract = read(D / "control.json")
    failures = executed[0].get("failureMessages", [])
    require(len(failures) == 8, "Expected exactly eight assertion failures")
    observed_labels = []
    for failure in failures:
        labels = [label for label in contract["expectedNamedFailures"] if label in failure]
        require(len(labels) == 1 and "AssertionError" in failure, "Unexpected assertion failure")
        observed_labels.extend(labels)
    require(sorted(observed_labels) == sorted(contract["expectedNamedFailures"]), "Named failure set mismatch")
    log = D / "run" / (NAME + ".log")
    observations = []
    for line in re.sub(r"\x1b\[[0-9;]*m", "", log.read_text()).splitlines():
        try:
            item = json.loads(line.strip())
        except ValueError:
            continue
        if isinstance(item, dict) and item.get("observation") == contract["reportLabel"]:
            observations.append(item)
    require(len(observations) == 1, "Genuine source observation missing or duplicated")
    item = observations[0]
    require(item["verdict"] == "see-final-vitest-result" and item["attempts"] == 11
            and item["closed"] is True and item["directoryAbsent"] is True
            and item["startupCompleted"] is True and item["deadlineExceeded"] is False
            and item["failureCount"] == 0, "Source prerequisite/cleanup failed")
    require(not Path(item["directory"]).exists(), "Owned source directory remains")
    responses = item["completedResponses"]
    require(len(responses) == 11
            and [x["status"] for x in responses] == contract["expectedObservedStatuses"], "Actual HTTP statuses/count changed")
    # Fixed Scope independently read from the unchanged source fixture.
    scope = "https://example.test/events/"
    source_text = (C / contract["consumerSourceFile"]).read_text()
    require('const scope = "' + scope + '"' in source_text, "Reported Scope not in fixed source fixture")
    expected = [("POST", scope + "operations/batch", "application/json"),
                ("GET", scope + "events/", "application/json")]
    for collection in ("beads", "links"):
        for accept in ("application/json", "text/event-stream"):
            for method in ("GET", "HEAD"):
                expected.append((method, scope + collection + "/missing?view=events", accept))
    expected.append(("GET", scope + "beads/missing?view=events", "text/event-stream"))
    require([(r["method"], r["url"], r["accept"]) for r in responses] == expected, "Actual route/method/Accept ledger mismatch")
    fields = ("bdp-scope-epoch", "bdp-authorization-view", "bdp-scope-position")
    baseline = [responses[1]["headers"][name] for name in fields]
    require(all(baseline) and baseline[2] == "pos-1", "Independent positive context missing")
    require(all([r["headers"][name] for name in fields] == [None, None, None]
                for r in responses[2:]), "Negative/anonymous context tuple mismatch")
    require(all(re.fullmatch("[0-9a-f]{64}", r["bodySha256"]) for r in responses), "Body observation hashes malformed")
    return {"verdict": "EXPECTED_NAMED_CONTROL_FAILURE", "actualTestExit": 1,
            "fullName": executed[0]["fullName"], "namedFailures": observed_labels,
            "httpResponses": 11, "productChildren": 0, "sourceObservation": item,
            "resultSha256": sha(output), "logSha256": sha(log),
            "limitation": "Source-consumer derivative, not installed qualification. Supervisor samples process groups; source report and final absence checks are separate evidence."}

def main():
    require(sys.argv[1:] == [NAME], "Exactly the accepted control command is allowed")
    require(sha(BASE) == BASE_SHA, "Reviewed base supervisor changed")
    require(sha(D / "control.json") == CONTROL_SHA, "Control contract changed")
    gate = read(P / "execution-gate.json")  # Intentionally absent until root release.
    require(gate["status"] == "released" and gate["session"] == SESSION, "Control gate closed")
    require(gate["authorizedStages"] == [NAME] and gate["installedGate"] == "closed", "Broader control release rejected")
    require(gate["supervisorSha256"] == sha(Path(__file__)) and gate["packetSha256"] == PACKET_SHA,
            "Control supervisor/packet binding changed")
    require(gate["baseSupervisorSha256"] == BASE_SHA, "Base supervisor gate mismatch")
    authority = Path(gate["rootAcceptance"]["path"])
    require(authority.is_relative_to(D.parent) and sha(authority) == gate["rootAcceptance"]["sha256"],
            "Root acceptance missing or changed")
    verify_graph()
    spec_module = importlib.util.spec_from_file_location("reviewed_event_source_supervisor", BASE)
    module = importlib.util.module_from_spec(spec_module)
    spec_module.loader.exec_module(module)
    original_source_hashes = module.source_hashes
    def source_with_graph():
        verify_graph()
        return original_source_hashes()
    module.source_hashes = source_with_graph
    def load_control(name):
        require(name == NAME, "Control identity mismatch")
        require(sha(D / "commands.json") == module.COMMANDS_SHA
                and sha(module.GUARD) == module.GUARD_SHA, "Commands/session guard changed")
        source = module.source_hashes()
        binding = read(module.BINDING)
        require(module.source_patch_hash() == binding["diffSha256"], "Positive draft changed")
        for previous in module.ORDER:
            previous_result = read(D / "run" / (previous + ".exit.json"))
            require(previous_result["supervisorExit"] == 0 and previous_result["sourceUnchanged"]
                    and previous_result["sourceSha256After"] == source, "Positive prerequisite incomplete/failed")
        spec = next(dict(x) for x in read(D / "commands.json")["commands"] if x["id"] == NAME)
        require(spec["cwd"] == str(C) and spec["supervisorSeconds"] == 120
                and spec["cases"] == 1 and spec["http"] == 11 and spec["productChildren"] == 0, "Control limits changed")
        for arg in spec["argv"]:
            if arg.startswith("--outputFile="):
                require(not Path(arg.split("=", 1)[1]).exists(), "Prior test result exists; no retry")
        spec["ownedTempPrefixes"] = ["bdp-events-http-"]
        spec["workload"] = {key: spec[key] for key in ("cases", "http", "owners", "directories", "listeners", "productChildren")}
        return {**gate, "sourceSha256": source, "sourceDiffSha256": binding["diffSha256"],
                "preparationSha256": sha(Path(__file__)),
                "controls": {NAME: read(P / "mutation.json")}}, {**spec, "control": NAME}
    module.load_gate = load_control
    code = module.main()
    # Preserve actual exit1; do not convert it into a claim that the test passed.
    outcome = {"verdict": "CONTROL_NOT_ACCEPTED", "actualSupervisorExit": code}
    try:
        verify_graph()
        outcome = adjudicate(module)
    except Exception as error:
        outcome["reason"] = str(error)
    destination = D / "run" / (NAME + ".adjudication.json")
    with destination.open("x") as handle:
        handle.write(json.dumps(outcome, indent=2) + "\n")
    print(json.dumps({"adjudication": str(destination), "verdict": outcome["verdict"]}))
    return 0 if outcome["verdict"] == "EXPECTED_NAMED_CONTROL_FAILURE" else 1

if __name__ == "__main__":
    sys.exit(main())
