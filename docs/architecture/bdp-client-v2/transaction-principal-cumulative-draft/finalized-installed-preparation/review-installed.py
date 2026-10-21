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
require(r["declarations"]["exactConfigurationDigest"]=="9a9404e74dd41fcb39b9c3bb5fc86917ff972dfdbc95faa5f53d13b35bb60bec","Route configuration differs")
principal="read-update.idempotency.principal-isolation"
member="transactional.sequence.member-transactions"
memberids=["s-sequence","s-first-receipt","s-first-facts","s-failed-receipt","s-failed-facts","s-last-receipt","s-last-facts","s-replay","s-current","s-events"]
manifest=json.loads((W/"packages/conformance/test-support/transactional/manifest-v1.json").read_bytes())
authored={x["id"] for x in manifest["scenarios"]}
selected=name in ("principal-positive","principal-collapse")
collapse=name=="principal-collapse"
cross=name in ("cross-batch","cross-sequence")
negative=collapse or cross
oracle=cfg["oracles"].get(name)
if cross:
    artifact=Path(oracle["specimenPath"])
    require(sha(artifact.read_bytes())==oracle["specimenSha256"],"Derivative specimen changed")
    require(r["artifacts"]["manifestDigest"]==oracle["compactSha256"],"Actual cross derivative differs")
    changed_id=oracle["scenario"];action_id=oracle["action"]
    changed=next(x for x in r["scenarios"] if x["id"]==changed_id)
    require(changed["state"]=="fail" and [e["request"]["id"] for e in changed["exchanges"]]==oracle["prefix"],"Cross prefix differs")
    last=changed["exchanges"][-1]
    require(last["request"]["id"]==action_id and last["response"]["status"]==oracle["status"],"Cross actual response differs")
    require([a["id"] for a in last["assertions"]]==oracle["allAssertionIds"] and [a["id"] for a in last["assertions"] if not a["passed"]]==oracle["failedIds"],"Wrong exact cross failures")
else:
    require(r["artifacts"]["manifestDigest"]==sha((W/"packages/conformance/test-support/transactional/manifest-v1.json").read_bytes()),"Canonical source artifact differs")
    changed_id=principal if collapse else None;action_id="pk-b-create" if collapse else None
if selected:
    require(r["selectedScenarioIds"]==[principal] and len(r["scenarios"])==1,"Wrong selected principal row")
else:
    require(len(r["scenarios"])==256 and len(r["selectedScenarioIds"])==256 and authored<=set(r["selectedScenarioIds"]),"Wrong full catalog selection")
    require(sum(x["state"]=="pass" for x in r["scenarios"])==(15 if cross else 16),"Wrong authored pass count")
for row in r["scenarios"]:
    if row["id"] not in authored:continue
    require(not row.get("cleanupError") and not row.get("prerequisiteFailure"),"Row cleanup/prerequisite error")
    require(row["state"]==("fail" if negative and row["id"]==changed_id else "pass"),"Unexpected authored row state")
    for action in row.get("actions",[]):require(not action.get("executionError") and all(a["passed"] for a in action["assertions"]),"Client action failure")
    for ex in row["exchanges"]:
        require(not ex.get("transportError") and not ex.get("harnessError") and ex["request"]["exact"]["writeState"]=="complete","Transport/harness/write failure")
        if not negative or row["id"]!=changed_id or ex["request"]["id"]!=action_id:require(all(a["passed"] for a in ex["assertions"]),"Unexpected assertion failure")
# Principal safe projection is corroboration of the frozen app's in-memory full-body checks.
# It cannot recover or independently recreate receipt/Resource/Event JSON from fingerprints.
contract=json.loads((P/"report-contract.json").read_text())
requests=json.loads((P/"requests.json").read_text())["requests"]
control=json.loads((P/"control.json").read_text())
pr=next(x for x in r["scenarios"] if x["id"]==principal)
count=2 if collapse else 7
require(not pr.get("actions") and len(pr["exchanges"])==count,"Principal raw-only inventory differs")
require([x["request"]["id"] for x in pr["exchanges"]]==[x["action"] for x in requests[:count]],"Principal request order differs")
if collapse:
    last=pr["exchanges"][1]
    expected_ids=[a["id"] for a in manifest["scenarios"][0]["actions"][1]["assertions"]]
    require([a["id"] for a in last["assertions"]]==expected_ids,"Collapse complete assertion inventory differs")
    require([a["id"] for a in last["assertions"] if not a["passed"]]==control["failingAssertionIds"] and len(control["failingAssertionIds"])==18,"Collapse expected18 failures absent")
    require([a["id"] for a in last["assertions"] if a["passed"]]==control["passingAssertionIds"] and len(control["passingAssertionIds"])==7,"Collapse expected7 passes absent")
else:
    for i,ex in enumerate(pr["exchanges"]):
        require([a["id"] for a in ex["assertions"]]==[a["id"] for a in manifest["scenarios"][0]["actions"][i]["assertions"]],"Principal assertion inventory differs")
supp=o["principalSupplemental"]
require(set(supp)==set(contract["supplementalKeys"]),"Supplement field inventory differs")
require(supp["mode"]==("collapse-principal-b-to-a" if collapse else "positive"),"Supplement mode differs")
require(supp["provenance"]=="input method/url/credentialRef; local zero-based ordinal; report id/reportedUrl; actual response metadata and completed request fingerprints; no wire credential attestation","Wrong safe projection attribution")
require(supp["performed"]==contract["controlPerformed" if collapse else "positivePerformed"],"Missing actual performed checks")
require(len(supp["dispatches"])==len(supp["responses"])==len(supp["joined"])==count,"Incomplete principal observations")
for i,(literal,dispatch,response,joined,exchange) in enumerate(zip(requests[:count],supp["dispatches"],supp["responses"],supp["joined"],pr["exchanges"])):
    require(dispatch=={"ordinal":i,"method":literal["method"],"url":literal["dispatchUrl"],"credentialRef":literal["credentialRef"]},"Actual dispatch projection differs")
    require(joined=={"id":literal["action"],"reportedUrl":literal["reportedUrl"],"method":literal["method"],"dispatchUrl":literal["dispatchUrl"],"credentialRef":literal["credentialRef"],"ordinal":i},"Actual report join differs")
    require(exchange["request"]["id"]==joined["id"] and exchange["request"]["url"]==joined["reportedUrl"] and exchange["request"]["method"]==joined["method"],"Join not bound to actual report")
    allowed=set(contract["responseFields"])
    require(set(response)==allowed-({"requestBodyDigest"} if literal["body"] is None else set()),"Unexpected/missing safe response field")
    status=409 if collapse and i==1 else literal["status"]
    position="pos-1" if i==0 or collapse else "pos-2"
    require(response["ordinal"]==i and response["url"]==literal["dispatchUrl"] and response["status"]==status and response["position"]==position,"Actual principal response differs")
    require(response["writeState"]==exchange["request"]["exact"]["writeState"]=="complete" and exchange["response"]["status"]==status,"Actual complete response mismatch")
    require(response["requestBodyBytes"]==literal["bodyBytes"] and response.get("requestBodyDigest")==literal["bodySha256"],"Principal request fingerprint differs")
    require(exchange["request"]["exact"]["bodyOctets"]==literal["bodyBytes"] and exchange["request"]["exact"]["bodyPresent"]==(literal["body"] is not None),"Reported actual body framing differs")
    require(re.fullmatch("[0-9a-f]{64}",response["bodyDigest"]) is not None,"Response fingerprint absent")
    require(all(isinstance(response[k],str) and response[k] for k in ("epoch","view")),"Empty context")
require(len({x["epoch"] for x in supp["responses"]})==1,"Principal epoch drift")
if collapse:require(supp["responses"][1]["view"]==supp["responses"][0]["view"],"Collapsed conflict context differs")
else:
    require(supp["responses"][2]["view"]==supp["responses"][4]["view"]==supp["responses"][0]["view"] and supp["responses"][3]["view"]==supp["responses"][1]["view"],"Replay/conflict serving view differs")
if not selected:
    row=next(x for x in r["scenarios"] if x["id"]==member)
    require([e["request"]["id"] for e in row["exchanges"]]==memberids,"Member observations incomplete")
    require([a["action"]["id"] for a in row["actions"]]==["s-first-facts","s-failed-facts","s-last-facts"],"Wrong helper inventory")
    require([e["request"]["method"] for e in row["exchanges"]]==["POST","POST","GET","POST","GET","POST","GET","POST","GET","GET"],"Wrong actual method inventory")
    roots=[]
    for index,source in [(2,"s-first-receipt"),(4,"s-failed-receipt"),(6,"s-last-receipt")]:
        ex=row["exchanges"][index];ref=ex["sourceReceiptObservation"];roots.append(ref["receipt"])
        require(ref["action"]==source and ref["pointer"]=="/id" and ref["credentialRef"]=="operator" and re.fullmatch("[0-9a-f]{64}",ref["responseDigest"]),"Wrong original-wire helper binding")
        require(ex["request"]["url"]==ref["receipt"] and ex["servedSession"]==ref["session"] and ex["request"]["exact"]["bodyPresent"] is False,"Wrong helper actual request/session")
        require([(a["id"],a["passed"]) for a in ex["assertions"]]==[("actual-status",True),("actual-request",True),("actual-body",True)],"Internal helper equality absent")
    require(len(set(roots))==3,"Receipt roots not distinct")
    if name=="full-positive":
        supp_member=o["sequenceMemberSupplemental"]
        require(supp_member["kind"]=="sequence-member-wire-supplement" and supp_member["negative"] is False,"Wrong member supplement")
        require(supp_member["observedIds"]==[x for x in memberids if not x.endswith("-facts")],"Missing member observations")
        require(supp_member["performedComparisons"]==[["s-replay","s-sequence","whole-json"]],"Missing member replay")
actual=sum(len(x["exchanges"]) for x in r["scenarios"])
require(actual==spec["workload"]["httpExpected"] and actual<=spec["workload"]["httpMaximum"] and len(o["sessions"])==spec["workload"]["children"]-1,"Actual workload differs")
body_bytes=sum(ex["request"]["exact"]["bodyOctets"] for row in r["scenarios"] for ex in row["exchanges"])
require(body_bytes==spec["workload"]["requestBodyBytes"] and body_bytes<=spec["workload"]["maximumRequestBodyBytes"],"Actual request body byte sum differs")
kind="installed-tx-sixteen-partial-summary" if name=="full-positive" else "installed-tx-principal-collapse" if collapse else "installed-tx-principal-positive" if selected else "installed-tx-cross-carrier-control"
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
if selected:require(logs[0]["principalSupplemental"]==o["principalSupplemental"] and logs[0]["exchanges"]==actual,"Log/report principal projection differs")
if name=="full-positive":require(logs[0]["actions"]==227 and logs[0]["sessions"]==23 and logs[0]["elapsedMs"]<=63000,"Full cohort bound differs")
if cross:require(logs[0]["failures"]==oracle["failedIds"] and logs[0]["observedStatus"]==oracle["status"] and logs[0]["exchanges"]==actual,"Cross log outcome differs")
review={"candidateMatches":True,"rootReviewRequired":True,"command":name,"actualHttp":actual,"actualRequestBodyBytes":body_bytes,"reportedChildren":logs[0]["childLaunches"],"reportedClosedChildren":logs[0]["closedChildren"],"sampledPidsIndependentlyAbsent":outcome["sampledPids"],"ownedRootsAbsent":True,"payloadIdentity":p["payloadIdentity"],"sourceCommit":g["finalHead"],"sourceTree":g["finalTree"],"reportSha256":sha(report.read_bytes()),"resultsSha256":sha((E/(name+".results.json")).read_bytes()),"logSha256":sha((E/(name+".log")).read_bytes()),"principalPerformed":o["principalSupplemental"]["performed"],"claimEligible":False,"limitations":"Full receipt/Resource/Event/replay comparisons occur in frozen app before redaction; fingerprints cannot reconstruct bodies or independently prove equality. Credential refs are actual input handles, not wire credential attestation. Real ready payload and source/archive verification are distinct; no child Git attestation. Process sampling may miss short-lived/detached descendants; root independent cleanup review remains required."}
with (E/(name+".review-candidate.json")).open("x") as f:f.write(json.dumps(review,indent=2)+"\n")
if name=="cross-sequence":
    allrows=[json.loads((E/(n+".review-candidate.json")).read_text()) for n in ["principal-positive","principal-collapse","full-positive","cross-batch","cross-sequence"]]
    require(sum(x["actualHttp"] for x in allrows)==652<=657 and sum(x["reportedChildren"] for x in allrows)==76 and sum(x["actualRequestBodyBytes"] for x in allrows)==54960<=55410,"Aggregate workload differs")
print(json.dumps(review))
