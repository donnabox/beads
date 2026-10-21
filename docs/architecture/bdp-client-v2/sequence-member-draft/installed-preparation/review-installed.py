"""Read genuine cohort reports and independently reconcile bindings/cleanup. No imports of product."""
from pathlib import Path
import json,os,re,sys
from gate import P,W,E,load_gate,completed,document,require,sha
name=sys.argv[1]
g,_=load_gate(name+"-review");outcome=completed(name)
cfg=json.loads((P/"commands.json").read_text());spec=cfg["commands"][name]
results=json.loads((E/(name+".results.json")).read_text())
executed=[(str(Path(f["name"]).resolve().relative_to(W)),a["fullName"],a["status"]) for f in results["testResults"] for a in f.get("assertionResults",[]) if a["status"] not in ("pending","skipped","todo")]
require(executed==[("apps/bdptest/src/transactional-matrix.test.ts",spec["selectedFullNames"][0],"passed")],"Wrong actual selected case")
require(results["success"] and results["numPassedTests"]==1 and results["numFailedTests"]==0 and not results.get("unhandledErrors") and not results.get("errors") and not results.get("numRuntimeErrorTestSuites",0),"Test result failed")
report=Path(spec["observationPath"]);o=json.loads(report.read_text());r=o["result"];p=document(g,"installedProvenance")
require(o["sourceCommit"]==g["finalHead"] and Path(o["entry"]).resolve()==Path(p["entryRealpath"]),"Wrong source/entry")
s=o["baseline"]["served"]
require(s["sourceCommit"]==g["finalHead"] and s["installedPayloadDigest"]==p["payloadIdentity"],"Wrong payload/source binding")
require(s["scope"]=="https://reference.invalid/acme/" and s["profile"]=="transactional" and s["history"] is False and o["baseline"]["authorityMode"]=="canonical","Wrong serving configuration")
for key in ("installedContractsDigest","policyDigest","authenticationDigest","fixtureDigest"):require(re.fullmatch("[0-9a-f]{64}",s[key]) is not None,"Missing serving digest")
require(s["fixtureDigest"]==sha((W/"packages/conformance/test-support/transactional/fixture-v1.json").read_bytes())==r["artifacts"]["fixtureDigest"],"Fixture differs")
for ready in o["sessions"]:
    require(ready["payloadIdentity"]==p["payloadIdentity"] and ready["runtimeConfigurationIdentity"]==o["baseline"]["runtimeConfigurationIdentity"],"Session payload/runtime differs")
    require(ready["authorityMode"]=="canonical" and ready["profile"]=="transactional" and ready["history"] is False and ready["claimEligible"] is False,"Session declaration differs")
require(r["claimEligible"] is False and r["profile"]=="transactional","False qualification")
require(r["declarations"]["exactConfigurationDigest"]=="02dd7d9f35e7482c07ce1550a6572232cc48ba6858cd66f8dfc8f5384f3628be","Route configuration differs")
member="transactional.sequence.member-transactions";memberids=["s-sequence","s-first-receipt","s-first-facts","s-failed-receipt","s-failed-facts","s-last-receipt","s-last-facts","s-replay","s-current","s-events"]
manifest=json.loads((W/"packages/conformance/test-support/transactional/manifest-v1.json").read_bytes());authored={x["id"] for x in manifest["scenarios"]}
selected=name=="selected-sequence-delete-control";negative=name!="positive-fifteen";oracle=cfg["oracles"].get(name)
if negative:
    artifact=Path(oracle["specimenPath"]);require(sha(artifact.read_bytes())==oracle["specimenSha256"],"Derivative specimen changed")
    require(r["artifacts"]["manifestDigest"]==oracle["compactSha256"],"Actual derivative differs")
    changed=next(x for x in r["scenarios"] if x["id"]==oracle["scenario"])
    require(changed["state"]=="fail" and [e["request"]["id"] for e in changed["exchanges"]]==oracle["prefix"],"Negative prefix differs")
    last=changed["exchanges"][-1]
    require(last["request"]["id"]==oracle["action"] and last["response"]["status"]==oracle["status"],"Negative response differs")
    require([a["id"] for a in last["assertions"]]==oracle["allAssertionIds"] and [a["id"] for a in last["assertions"] if not a["passed"]]==oracle["failedIds"],"Wrong exact negative failures")
    require(all(a["passed"] for ex in changed["exchanges"][:-1] for a in ex["assertions"]),"Earlier negative failure")
else:require(r["artifacts"]["manifestDigest"]==sha((W/"packages/conformance/test-support/transactional/manifest-v1.json").read_bytes()),"Positive source artifact differs")
if selected:require(r["selectedScenarioIds"]==[member] and len(r["scenarios"])==1,"Wrong selected row")
else:
    require(len(r["scenarios"])==256 and len(r["selectedScenarioIds"])==256 and member in r["selectedScenarioIds"],"Wrong full catalog selection")
    require(sum(x["state"]=="pass" for x in r["scenarios"])==(14 if negative else 15),"Wrong authored pass count")
for row in r["scenarios"]:
    if row["id"] not in authored:continue
    require(not row.get("cleanupError") and not row.get("prerequisiteFailure"),"Row cleanup/prerequisite error")
    if not negative or row["id"]!=oracle["scenario"]:require(row["state"]=="pass","Unexpected authored failure")
    for action in row.get("actions",[]):require(not action.get("executionError") and all(a["passed"] for a in action["assertions"]),"Client action failure")
    for ex in row["exchanges"]:
        require(not ex.get("transportError") and not ex.get("harnessError") and ex["request"]["exact"]["writeState"]=="complete","Transport/harness/write failure")
        if not negative or row["id"]!=oracle["scenario"] or ex["request"]["id"]!=oracle["action"]:require(all(a["passed"] for a in ex["assertions"]),"Unexpected assertion failure")
row=next(x for x in r["scenarios"] if x["id"]==member)
require([e["request"]["id"] for e in row["exchanges"]]==(["s-sequence"] if selected else memberids),"Member observations incomplete")
if selected:require(not row.get("actions"),"Negative unexpectedly reached client")
else:
    require([a["action"]["id"] for a in row["actions"]]==["s-first-facts","s-failed-facts","s-last-facts"],"Wrong helper inventory")
    require([e["request"]["method"] for e in row["exchanges"]]==["POST","POST","GET","POST","GET","POST","GET","POST","GET","GET"],"Wrong actual method inventory")
    roots=[]
    for index,source in [(2,"s-first-receipt"),(4,"s-failed-receipt"),(6,"s-last-receipt")]:
        ex=row["exchanges"][index];ref=ex["sourceReceiptObservation"];roots.append(ref["receipt"])
        require(ref["action"]==source and ref["pointer"]=="/id" and ref["credentialRef"]=="operator" and re.fullmatch("[0-9a-f]{64}",ref["responseDigest"]),"Wrong original-wire helper binding")
        require(ex["request"]["url"]==ref["receipt"] and ex["servedSession"]==ref["session"] and ex["request"]["exact"]["bodyPresent"] is False,"Wrong helper actual request/session")
        require([(a["id"],a["passed"]) for a in ex["assertions"]]==[("actual-status",True),("actual-request",True),("actual-body",True)],"Internal helper equality absent")
    require(len(set(roots))==3,"Receipt roots not distinct")
if name in ("positive-fifteen","selected-sequence-delete-control"):
    supp=o["supplemental"] if selected else o["sequenceMemberSupplemental"]
    require(supp["kind"]=="sequence-member-wire-supplement" and supp["negative"] is selected,"Wrong supplement")
    require(supp["observedIds"]==(["s-sequence"] if selected else [x for x in memberids if not x.endswith("-facts")]),"Missing raw observations")
    require(supp["performedComparisons"]==([] if selected else [["s-replay","s-sequence","whole-json"]]),"Missing actual replay comparison")
actual=sum(len(x["exchanges"]) for x in r["scenarios"])
require(actual==spec["workload"]["httpExpected"] and actual<=spec["workload"]["httpMaximum"] and len(o["sessions"])==spec["workload"]["children"]-1,"Actual workload differs")
kind="installed-tx-fifteen-partial-summary" if name=="positive-fifteen" else "installed-tx-sequence-member-control" if selected else "installed-tx-cross-carrier-control"
logs=[]
for line in (E/(name+".log")).read_text().splitlines():
    line=re.sub(r"\x1b\[[0-9;]*m","",line).strip()
    if not line.startswith("{"):continue
    try:v=json.loads(line)
    except json.JSONDecodeError:continue
    if v.get("kind")==kind:logs.append(v)
require(len(logs)==1 and logs[0]["sourceCommit"]==g["finalHead"] and logs[0]["childLaunches"]==logs[0]["closedChildren"]==spec["workload"]["children"],"Child inventory/closure differs")
# Check observed process IDs and owned temporary roots again after supervisor.
for pid in outcome["sampledPids"]:
    try:os.kill(pid,0)
    except ProcessLookupError:continue
    raise RuntimeError("Sampled process still present: "+str(pid))
tmp=E/(name+".tmp")
owned=[p for p in tmp.iterdir() if p.name.startswith("bdp-tx-installed-")] if tmp.exists() else []
require(not owned,"Owned fixture directory remains")
review={"candidateMatches":True,"rootReviewRequired":True,"command":name,"actualHttp":actual,"reportedChildren":logs[0]["childLaunches"],"reportedClosedChildren":logs[0]["closedChildren"],"sampledPidsIndependentlyAbsent":outcome["sampledPids"],"ownedRootsAbsent":True,"payloadIdentity":p["payloadIdentity"],"sourceCommit":g["finalHead"],"sourceTree":g["finalTree"],"reportSha256":sha(report.read_bytes()),"resultsSha256":sha((E/(name+".results.json")).read_bytes()),"logSha256":sha((E/(name+".log")).read_bytes()),"claimEligible":False,"limitations":"Full raw bodies/Resource/Event/replay checks occur in frozen app and original-wire helper before redaction; this review cannot recover report body values. Cross reports deliberately retain original shape without new supplement; passing exact callback is the supplement evidence. Source attribution is supervisor verified, not child-attested Git. Process sampling cannot prove every detached/short-lived PID was seen."}
with (E/(name+".review-candidate.json")).open("x") as f:f.write(json.dumps(review,indent=2)+"\n")
if name=="cross-sequence-control":
    allrows=[json.loads((E/(n+".review-candidate.json")).read_text()) for n in ["positive-fifteen","selected-sequence-delete-control","cross-batch-control","cross-sequence-control"]]
    require(sum(x["actualHttp"] for x in allrows)==623<=632 and sum(x["reportedChildren"] for x in allrows)==71,"Aggregate workload differs")
print(json.dumps(review))
