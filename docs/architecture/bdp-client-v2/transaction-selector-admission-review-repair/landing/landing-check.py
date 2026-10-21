#!/usr/bin/env python3
"""Read-only PR124 pre-merge / actual-merge verification. Prints JSON; never merges."""
import argparse, datetime, hashlib, json, re, subprocess
from pathlib import Path
P = Path(__file__).resolve().parent
E = P.parent
W = Path("/Users/dbox/repos/wt/vickie/vickie-selector-review-repair")
REPO = "gastownhall/bdp"
PR, RUN = 124, 38677561010
BRANCH = "codex/vickie-selector-review-repair"
B = "757aedc8008386745d59d5005498a9256d6ea76e"
O = "22777eccc347b9c926de1da92409de9a924ba692"
H = "00d7fff3320eb10d4165718d23d86519cd3b7c5e"
OT = "fe1867a19de19a824b3569a3b901269dddec221f"
T = "062c32778217aeb702b2f7798dc82b0d985e9c67"
BINDING = "03a861851a195d5da505f311291bee6d3cd98f63866d4b6df756e2c3ae289afd"
PATCH = "2c841e647af2067abb5d50705be60489eb8c9c868e6346e922a0f2d5ee85298a"
TEST = "apps/bdptest/src/transactional-selector-admission-installed.test.ts"
TITLE = "installed Selector grammar and fresh capacity preserve keys receipts and complete effects"
ROLES = {
    "nativeSource", "claudeSourceReview", "claudeSourceExit", "sourceDisposition",
    "staticResults", "staticExit", "sourcePositives", "specBoundary", "controls",
    "nativeControls", "repairBinding", "repairPatch", "nativeRepair",
    "claudeRepairReview", "claudeRepairExit", "repairDisposition",
    "installedProvenance", "installedExit", "installedResults", "installedLog",
    "installedCandidate", "nativeInstalled", "ciLog",
}
def check(value, message):
    if not value: raise RuntimeError(message)
def digest(data): return hashlib.sha256(data).hexdigest()
def raw(*args): return subprocess.check_output(args, cwd=W, timeout=60)
def command(*args): return raw(*args).decode().strip()
def gh(*args): return json.loads(raw("gh", *args))
def git(*args): return command("git", *args)
def api(path): return gh("api", "repos/" + REPO + "/" + path)
def current_main(): return api("git/ref/heads/main")["object"]["sha"]
def parse(data): return json.loads(data)
def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--verify-merged", action="store_true")
    args = parser.parse_args()
    acceptance_path = P / "root-landing-acceptance.json"
    acceptance_bytes = acceptance_path.read_bytes()  # absent until actual root acceptance
    a = parse(acceptance_bytes)
    check(a["accepted"] is True and a["actionableFindingsResolved"] is True,
          "root review/disposition incomplete")
    check((a["pr"], a["ciRun"], a["base"], a["originalHead"], a["head"], a["tree"])
          == (PR, RUN, B, O, H, T), "wrong acceptance identity")
    check(a["publicQualification"] is False and a["releaseAuthorized"] is False,
          "out-of-scope qualification/release")
    check(a["checkerSHA256"] == digest(Path(__file__).read_bytes()), "checker changed")
    check(set(a["evidence"]) == ROLES, "complete exact evidence-role set required")
    captured = {}
    for role, item in a["evidence"].items():
        path = Path(item["path"])
        check(path.is_absolute() and path.resolve().is_relative_to(E.resolve()), "evidence outside task: " + role)
        data = path.read_bytes()
        check(digest(data) == item["sha256"], "changed evidence: " + role)
        captured[role] = (path, data)
    j = lambda role: parse(captured[role][1])
    # Preserve the original draft's evidence attribution. Never relabel its runs as H.
    binding_bytes = (E / "source-binding.json").read_bytes()
    check(digest(binding_bytes) == BINDING, "historical source binding changed")
    old_files = {x["path"]: x["sha256"] for x in parse(binding_bytes)["files"]}
    check(len(old_files) == 12, "historical twelve-file footprint")
    check(git("rev-parse", "HEAD") == H and git("rev-parse", "HEAD^{tree}") == T
          and not git("status", "--porcelain"), "clean exact final source required")
    check(git("show", "-s", "--format=%P", H) == O and git("show", "-s", "--format=%P", O) == B,
          "ordinary chain changed")
    check(git("rev-parse", O + "^{tree}") == OT, "original tree changed")
    check(set(git("diff", "--name-only", B, H).splitlines()) == set(old_files), "twelve-file scope changed")
    check(git("diff", "--name-only", O, H) == TEST, "assertion repair scope changed")
    delta = raw("git", "diff", "--binary", O, H)
    check(digest(delta) == PATCH and delta == captured["repairPatch"][1], "one-line repair patch changed")
    repair = j("repairBinding")
    check(repair["head"] == O and repair["tree"] == OT and repair["file"] == TEST
          and repair["beforeSha256"] == old_files[TEST] and repair["patchSha256"] == PATCH, "repair binding changed")
    final_files = {**old_files, TEST: repair["afterSha256"]}
    for path, before in old_files.items():
        check(digest(raw("git", "show", O + ":" + path)) == before, "historical source mismatch: " + path)
        check(digest((W / path).read_bytes()) == final_files[path], "final source mismatch: " + path)
    statics = j("staticResults")
    check(j("staticExit") == {"completed": 7, "allPassed": True, "sourceUnchanged": True}, "static outcome")
    check(len(statics) == 7 and all(x["exitCode"] == 0 and not x["timeout"] and x["sourceUnchanged"] for x in statics), "static checks incomplete")
    positive = j("sourcePositives")
    check(positive["accepted"] and positive["actualPassedCases"] == 90 and positive["completedCommands"] == 4
          and not positive["notRun"] and positive["sourceBefore"]["files"] == old_files, "historical positive evidence")
    boundary = j("specBoundary")
    check(boundary["accepted"] and boundary["exit"] == 0 and len(boundary["passedTitles"]) == 9
          and not boundary["failed"] and not boundary["processesRemaining"], "nine spec boundary cases incomplete")
    controls = j("controls")
    check(controls["accepted"] and not controls["notRun"] and len(controls["results"]) == 2, "historical controls incomplete")
    check([x["id"] for x in controls["results"]] == ["grammar-development", "omit-fresh-capacity"], "control identity")
    for x in controls["results"]:
        check(x["accepted"] and x["exitCode"] == 1 and not x["timeout"] and not x["errors"]
              and x["allBindingsPreservedAfter"] and not x["processesAfterCleanup"] and not x["tempAfter"], "unclean control")
    original_review, delta_review = j("claudeSourceExit"), j("claudeRepairExit")
    check(original_review["exitCode"] == 0 and not original_review["timeout"] and original_review["bothSourcesUnchanged"]
          and original_review["outputSHA256"] == digest(captured["claudeSourceReview"][1]), "original actual council missing")
    check(delta_review["exitCode"] == 0 and not delta_review["timeout"] and delta_review["sourceUnchanged"]
          and delta_review["stdoutSha256"] == digest(captured["claudeRepairReview"][1]), "repair actual council missing")
    disposition = j("repairDisposition")
    check(disposition["accepted"] is True and disposition["source"] == H, "repair disposition incomplete")
    for path, expected in disposition["artifacts"].items():
        target = (E / path).resolve()
        check(target.is_relative_to(E.resolve()) and digest(target.read_bytes()) == expected, "disposition artifact changed: " + path)
    installed_exit = j("installedExit")
    check(installed_exit["exit"] == installed_exit["supervisorExit"] == 0 and installed_exit["sourceUnchanged"], "installed command failed")
    for key in ("timeout", "forcedCleanup", "errors", "remainingAfterCleanup", "sampledPidsStillPresent", "runtimeDirectoriesStillPresent"):
        check(not installed_exit[key], "installed cleanup failed: " + key)
    tests = j("installedResults")
    selected = [x for suite in tests["testResults"] for x in suite["assertionResults"] if x["status"] not in ("pending", "skipped", "todo", "disabled")]
    check(tests["success"] and len(selected) == 1 and selected[0]["fullName"] == TITLE and selected[0]["status"] == "passed", "installed title/result")
    installed = j("installedCandidate")
    check(installed["candidateMatches"] and installed["actualCompletedResponses"] == 25 and installed["reportedChildren"] == 1
          and installed["logSha256"] == digest(captured["installedLog"][1])
          and installed["resultsSha256"] == digest(captured["installedResults"][1]), "installed candidate not bound")
    observation = installed["observation"]
    check(observation["sourceAttribution"] == H and not observation["workspaceDirty"] and not observation["claimEligible"]
          and observation["failureCount"] == 0 and len(observation["observations"]) == 25, "installed actual attribution")
    cleanup = observation["cleanup"]
    check(cleanup["exitCode"] == 0 and cleanup["signalCode"] is None and cleanup["childAbsent"] and cleanup["directoryAbsent"], "installed child cleanup")
    provenance = j("installedProvenance")
    check(provenance["sourceCommit"] == H and provenance["sourceTree"] == T
          and observation["entry"] == provenance["entryRealpath"] and observation["entrySha256"] == provenance["entrySha256"]
          and observation["binding"]["served"]["installedPayloadDigest"] == provenance["payloadIdentity"], "installed payload binding")
    # Read the actual PR/run, not a branch-name assertion or an old green run.
    fields = "state,headRefOid,headRefName,baseRefOid,baseRefName,isDraft,statusCheckRollup,mergeCommit"
    pr = gh("pr", "view", str(PR), "--repo", REPO, "--json", fields)
    check(pr["headRefOid"] == H and pr["headRefName"] == BRANCH and pr["baseRefName"] == "main", "PR identity changed")
    ci = gh("run", "view", str(RUN), "--repo", REPO, "--json", "databaseId,status,conclusion,headSha,jobs,event,workflowName")
    check((ci["databaseId"], ci["headSha"], ci["status"], ci["conclusion"], ci["event"], ci["workflowName"])
          == (RUN, H, "completed", "success", "pull_request", "CI"), "exact completed green CI missing")
    check(len(ci["jobs"]) == 1, "unexpected CI job inventory")
    job = ci["jobs"][0]
    check(job["databaseId"] == a["ciJob"] and job["name"] == "verify" and job["status"] == "completed"
          and job["conclusion"] == "success" and job["steps"], "CI job not complete")
    check(all(x["status"] == "completed" and x["conclusion"] == "success" for x in job["steps"]), "CI has non-success step")
    for name in ("Run pnpm install --frozen-lockfile", "Run pnpm format:check", "Run pnpm lint", "Run pnpm check:boundaries", "Run pnpm typecheck", "Run pnpm build", "Run pnpm test", "Run pnpm e2e:ready:preflight", "Run pnpm smoke", "Read-cohort evidence gate"):
        check(sum(x["name"] == name for x in job["steps"]) == 1, "missing/duplicate CI step: " + name)
    log = raw("gh", "api", f"repos/{REPO}/actions/jobs/{a['ciJob']}/logs")
    check(log == captured["ciLog"][1], "live CI log differs from accepted actual log")
    lines = re.sub(r"\x1b\[[0-9;]*m", "", log.decode()).splitlines()
    tested = []
    for index, line in enumerate(lines[:-1]):
        if "/usr/bin/git log -1 --format=%H" in line:
            match = re.search(r"\b[0-9a-f]{40}\b", lines[index + 1])
            check(match is not None, "missing actual checkout commit")
            tested.append(match.group())
    check(tested and set(tested) == {a["testedCommit"]}, "actual tested checkout mismatch")
    synthetic = api("git/commits/" + a["testedCommit"])
    check(synthetic["tree"]["sha"] == T and [x["sha"] for x in synthetic["parents"]] == [B, H], "CI synthetic tree/ordered parents differ")
    rollup = pr["statusCheckRollup"]
    check(rollup and all((x.get("status") == "COMPLETED" and x.get("conclusion") == "SUCCESS") or x.get("state") == "SUCCESS" for x in rollup), "PR has pending/non-green checks")
    check(any(f"/runs/{RUN}/job/{a['ciJob']}" in x.get("detailsUrl", "") for x in rollup), "PR rollup lacks exact CI job")
    main_head = current_main()
    actual = None
    if args.verify_merged:
        check(pr["state"] == "MERGED" and pr["mergeCommit"]["oid"] == main_head, "actual merge is not current main")
        actual = api("git/commits/" + main_head)
        check(actual["tree"]["sha"] == T and [x["sha"] for x in actual["parents"]] == [B, H], "actual merge tree/ordered parents differ")
    else:
        check(pr["state"] == "OPEN" and pr["baseRefOid"] == B and main_head == B, "base moved or PR closed")
        check(api("git/ref/heads/" + BRANCH)["object"]["sha"] == H, "published branch moved")
    check(gh("pr", "view", str(PR), "--repo", REPO, "--json", fields) == pr and current_main() == main_head, "remote state changed during verification")
    check(acceptance_path.read_bytes() == acceptance_bytes and all(path.read_bytes() == data for path, data in captured.values()), "accepted evidence changed during verification")
    check(git("rev-parse", "HEAD") == H and not git("status", "--porcelain")
          and all(digest((W / p).read_bytes()) == v for p, v in final_files.items()), "source changed during verification")
    print(json.dumps({"verifiedAt": datetime.datetime.now(datetime.timezone.utc).isoformat(), "mode": "merged" if args.verify_merged else "pre-merge",
        "pr": PR, "head": H, "base": B, "tree": T, "ciRun": RUN, "ciJob": a["ciJob"], "testedCommit": a["testedCommit"],
        "historicalSourceBinding": BINDING, "historicalOriginalHead": O, "installedEvidenceHead": H,
        "currentMain": main_head, "actualMerge": None if actual is None else {"sha": main_head, "tree": actual["tree"]["sha"], "parents": [x["sha"] for x in actual["parents"]]},
        "draft": pr["isDraft"], "acceptanceSHA256": digest(acceptance_bytes), "publicQualification": False, "releaseAuthorized": False}, indent=2))
if __name__ == "__main__": main()
