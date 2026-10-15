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
BASE_SHA = "f6758bd5b94232b1d2c6186835b5271d156ab665a0f2681269e550f9527b2ac8"
NAME = "omit-context-control"
SESSION = "2026-10-08T20:26:35Z-81e5"
PACKET_SHA = None  # Root pins actual post-build manifests in the separate gate.
CONTROL_SHA = "2e5e15ccfe1f4a7544cb02c086cfce2685bacc35bbb41e8688a5314ac5690cdd"
COMMAND_SHA = "9360fffc5539f852962960d824473c66932327509fd9b0a2fb41edf02a7e8ce4"

def require(ok, message):
    if not ok:
        raise RuntimeError(message)

def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()

def read(path):
    return json.loads(path.read_text())

def normalized_shim(data):
    # Only pnpm-generated NODE_PATH/comment spellings may differ. Executable
    # shell statements and all other dependency bytes retain their normal hash.
    require((C / "node_modules").resolve(strict=True) == (W / "node_modules").resolve(strict=True),
            "Root tooling node_modules targets differ")
    copied = str(C / "node_modules").encode()
    positive = str(W / "node_modules").encode()
    lines = []
    for line in data.splitlines(keepends=True):
        if copied in line:
            require(line.lstrip().startswith((b"export NODE_PATH=", b"# cmd-shim-target=")),
                    "Unexpected copy path in executable shim")
            line = line.replace(copied, positive)
        lines.append(line)
    return b"".join(lines)

def verify_graph():
    drift = []
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
    require(root.is_symlink() and str(root.resolve(strict=True)) == dependencies["rootLink"]["realpath"], "Root tooling graph changed")
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
                require(str(path.resolve(strict=True)) == item["realpath"], "Dependency effective target changed")
                if os.readlink(path) != item["link"]:
                    drift.append({"path": str(path), "beforeLink": item["link"], "afterLink": os.readlink(path), "realpath": item["realpath"]})
            else:
                if sha(path) != item["sha256"]:
                    require(".bin" in path.parts and "normalizedShimSha256" in item, "Non-shim dependency bytes changed")
                    require(hashlib.sha256(normalized_shim(path.read_bytes())).hexdigest() == item["normalizedShimSha256"], "Shim changed beyond bound path spelling")
                    drift.append({"path": str(path), "beforeSha256": item["sha256"], "afterSha256": sha(path), "normalFormSha256": item["normalizedShimSha256"]})
            seen.add(relative)
    require(seen == set(dependencies["members"]), "Shared dependency member missing")
    for name, item in dependencies["links"].items():
        positive, copied = W / name, C / name
        require(positive.is_symlink() and str(positive.resolve(strict=True)) == item["positiveRealpath"], "Positive package link changed")
        for path, oldtext in ((positive, item["positiveLink"]), (copied, item["copyLink"])):
            if path.is_symlink() and os.readlink(path) != oldtext:
                drift.append({"path": str(path), "beforeLink": oldtext, "afterLink": os.readlink(path), "realpath": str(path.resolve(strict=True))})
        require(copied.is_symlink() and str(copied.resolve(strict=True)) == item["copyRealpath"], "Copied package link changed")
    for name, item in dependencies["localShimFiles"].items():
        for root, key in ((W, "positiveSha256"), (C, "copySha256")):
            path = root / name
            require(path.is_file() and not path.is_symlink(), "Tool shim layout changed")
            if sha(path) != item[key]:
                require(hashlib.sha256(normalized_shim(path.read_bytes())).hexdigest() == item["normalizedShimSha256"],
                        "Tool shim changed beyond path spelling")
                drift.append({"path": str(path), "beforeSha256": item[key], "afterSha256": sha(path),
                              "normalFormSha256": item["normalizedShimSha256"]})
    for item in dependencies["workspaceFallbacks"]:
        require(sha(Path(item["realpath"]) / "package.json") == item["packageSha256"],
                "Shared fallback package changed")
    return {**packet, "allowedToolingPathSpellingChanges": drift}

def adjudicate(module):
    result = read(D / "run" / (NAME + ".exit.json"))
    require(result["exit"] == 1 and result["supervisorExit"] == 1, "Expected actual test failure absent")
    require(not result["timeout"] and not result["forcedCleanup"] and not result["errors"],
            "Control execution/cleanup error")
    for field in ("remainingAfterCleanup", "sampledPidsStillPresent", "runtimeDirectoriesStillPresent"):
        require(not result[field], "Control leaked owned resources: " + field)
    require(result["sourceUnchanged"], "Positive source changed")
    spec = read(P / "command.json")
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
    contract = read(D / "control-plan.json")
    failures = executed[0].get("failureMessages", [])
    require(len(failures) == 8, "Expected exactly eight assertion failures")
    labels_seen = []
    for failure in failures:
        labels = [label for label in contract["namedTupleFailures"] if label in failure]
        require(len(labels) == 1 and "AssertionError" in failure, "Unexpected assertion failure")
        labels_seen.extend(labels)
    require(sorted(labels_seen) == sorted(contract["namedTupleFailures"]), "Named failures mismatch")
    log = D / "run" / (NAME + ".log")
    observations = []
    for line in re.sub(r"\x1b\[[0-9;]*m", "", log.read_text()).splitlines():
        try:
            value = json.loads(line.strip())
        except ValueError:
            continue
        if isinstance(value, dict) and value.get("observation") == "source-receipt-missing-context":
            observations.append(value)
    require(len(observations) == 1, "Genuine receipt source report missing or duplicated")
    item = observations[0]
    require(item["attempted"] == 13 and item["tupleChecks"] == 8
            and item["registeredCleanup"] == 2 and item["completedCleanup"] == 2
            and item["deadlineExceeded"] is False and item["primaryOrCleanupFailureCount"] == 0,
            "Source prerequisites/cleanup failed")
    responses = item["responses"]
    require(len(responses) == 13 and [r["status"] for r in responses] ==
            [200,200,404,404,404,404,404,404,200,404,404,401,401], "Response count/status mismatch")
    seed_url = responses[0]["url"]
    require(seed_url.endswith("/operations/batch"), "Seed route mismatch")
    scope = seed_url[:-len("operations/batch")]
    root, page, unknown = responses[4]["url"], responses[1]["url"], responses[2]["url"]
    require(re.fullmatch(re.escape(scope + "receipts/") + r"[A-Za-z0-9_-]+", root) is not None,
            "Receipt root shape mismatch")
    require(re.fullmatch(re.escape(root + "/pages/") + r"[A-Za-z0-9_-]+", page) is not None,
            "Issued page shape mismatch")
    require(unknown == root.rsplit("/",1)[0] + "/unknown" and unknown != root, "Negative sibling mismatch")
    expected = [("POST",seed_url),("GET",page)]
    for url in (unknown,root,page):
        expected += [("GET",url),("HEAD",url)]
    expected += [("POST",scope+"operations/create-bead"),("GET",root),("GET",page),("GET",root),("HEAD",page)]
    require([(r["method"],r["url"]) for r in responses] == expected, "Actual route sequence mismatch")
    fields = ("bdp-scope-epoch","bdp-authorization-view","bdp-scope-position")
    first = [responses[0]["headers"][k] for k in fields]
    require(all(first) and first[2] == "pos-1", "Seed context missing")
    require([responses[1]["headers"][k] for k in fields] == first, "Page context mismatch")
    require([responses[8]["headers"][k] for k in fields] == first[:2]+["pos-2"], "Advance context mismatch")
    for i in [2,3,4,5,6,7,9,10,11,12]:
        require([responses[i]["headers"][k] for k in fields] == [None,None,None], "Omitted context mismatch")
    for i,r in enumerate(responses):
        require(r["headers"]["x-bdp-development"] == "non-attesting", "Marker missing")
        require(isinstance(r["bodyOctets"],int) and 0 <= r["bodyOctets"] <= 32768, "Body cap mismatch")
        if r["method"] == "HEAD":
            require(r["bodyOctets"] == 0, "HEAD body mismatch")
    for i in [2,4,6]:
        require(responses[i]["headers"]["content-length"] == responses[i+1]["headers"]["content-length"],
                "GET/HEAD length mismatch")
    observed_labels = labels_seen
    return {"verdict": "EXPECTED_NAMED_CONTROL_FAILURE", "actualTestExit": 1,
            "fullName": executed[0]["fullName"], "namedFailures": observed_labels,
            "httpResponses": 13, "productChildren": 0, "sourceObservation": item,
            "resultSha256": sha(output), "logSha256": sha(log),
            "limitation": "Source-consumer derivative, not installed qualification. Supervisor samples process groups; source report and final absence checks are separate evidence."}

def main():
    global PACKET_SHA
    require(sys.argv[1:] == [NAME], "Exactly the accepted control command is allowed")
    require(sha(BASE) == BASE_SHA, "Reviewed base supervisor changed")
    require(sha(D / "control-plan.json") == CONTROL_SHA, "Control contract changed")
    gate = read(P / "execution-gate.json")  # Intentionally absent until root release.
    require(gate["status"] == "released" and gate["session"] == SESSION, "Control gate closed")
    PACKET_SHA = gate["packetSha256"]
    require(isinstance(PACKET_SHA, str) and re.fullmatch("[0-9a-f]{64}", PACKET_SHA), "Actual build/copy binding absent")
    require(sha(P / "command.json") == COMMAND_SHA, "Control command changed")
    require(gate["authorizedStages"] == [NAME] and gate["installedGate"] == "closed", "Broader control release rejected")
    require(gate["supervisorSha256"] == sha(Path(__file__)) and gate["packetSha256"] == PACKET_SHA,
            "Control supervisor/packet binding changed")
    require(gate["baseSupervisorSha256"] == BASE_SHA, "Base supervisor gate mismatch")
    authority = Path(gate["rootAcceptance"]["path"])
    require(authority.is_relative_to(D.parent) and sha(authority) == gate["rootAcceptance"]["sha256"],
            "Root acceptance missing or changed")
    verify_graph()
    spec_module = importlib.util.spec_from_file_location("reviewed_receipt_source_supervisor", BASE)
    module = importlib.util.module_from_spec(spec_module)
    spec_module.loader.exec_module(module)
    module.BINDING_SHA = gate["sourceBindingSha256"]
    original_source_hashes = module.source_hashes
    def source_with_graph():
        verify_graph()
        return original_source_hashes()
    module.source_hashes = source_with_graph
    def load_control(name):
        require(name == NAME, "Control identity mismatch")
        require(sha(module.COMMANDS) == module.COMMANDS_SHA
                and sha(module.GUARD) == module.GUARD_SHA, "Commands/session guard changed")
        source = module.source_hashes()
        binding = read(module.BINDING)
        require(module.source_patch_hash() == binding["diffSha256"], "Positive draft changed")
        for previous in module.ORDER:
            result = read(D / "run" / (previous + ".exit.json"))
            require(result["exit"] == 0 and result["supervisorExit"] == 0 and result["sourceUnchanged"]
                    and result["sourceSha256After"] == source and not result["errors"]
                    and not result["timeout"] and not result["forcedCleanup"]
                    and not result["remainingAfterCleanup"] and not result["sampledPidsStillPresent"]
                    and not result["runtimeDirectoriesStillPresent"], "Positive prerequisite failed/unfinished")
            prior_spec = next(x for x in read(module.COMMANDS)["commands"] if x["id"] == previous)
            module.adjudicate_source(prior_spec)
        spec = read(P / "command.json")
        require(spec["cwd"] == str(C) and spec["supervisorSeconds"] == 120
                and spec["cases"] == 1 and spec["http"] == 13 and spec["productChildren"] == 0, "Control limits changed")
        for arg in spec["argv"]:
            if arg.startswith("--outputFile="):
                require(not Path(arg.split("=", 1)[1]).exists(), "Prior test result exists; no retry")
        spec["ownedTempPrefixes"] = ["bdp-tx-http-"]
        spec["workload"] = {key: spec[key] for key in ("cases", "http", "owners", "directories", "listeners", "productChildren")}
        return {**gate, "sourceSha256": source, "sourceDiffSha256": binding["diffSha256"],
                "preparationSha256": sha(Path(__file__)),
                "controls": {NAME: read(P / "mutation.json")}}, {**spec, "control": NAME}
    module.load_gate = load_control
    code = module.main()
    # Preserve actual exit1; do not convert it into a claim that the test passed.
    outcome = {"verdict": "CONTROL_NOT_ACCEPTED", "actualSupervisorExit": code}
    try:
        graph = verify_graph()
        outcome = {**adjudicate(module), "dependencyPathSpellingChanges": graph["allowedToolingPathSpellingChanges"]}
    except Exception as error:
        outcome["reason"] = str(error)
    destination = D / "run" / (NAME + ".adjudication.json")
    with destination.open("x") as handle:
        handle.write(json.dumps(outcome, indent=2) + "\n")
    print(json.dumps({"adjudication": str(destination), "verdict": outcome["verdict"]}))
    return 0 if outcome["verdict"] == "EXPECTED_NAMED_CONTROL_FAILURE" else 1

if __name__ == "__main__":
    sys.exit(main())
