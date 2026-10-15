"""Adjudicate one genuine installed cohort report; partial evidence, never admission."""
from pathlib import Path
import json,re,sys
from gate import P,W,E,load_gate,completed,document,require,sha
name=sys.argv[1]
g,_=load_gate(name+"-review")
completed(name)
config=json.loads((P/"commands.json").read_text());spec=config["commands"][name]
results=json.loads((E/(name+".results.json")).read_text())
rows=[(str(Path(suite["name"]).resolve().relative_to(W)),a["fullName"],a["status"]) for suite in results["testResults"] for a in suite["assertionResults"] if a["status"] not in ("pending","skipped","todo")]
require(rows==[(a["path"],a["fullName"],"passed") for a in spec["selectedCases"]],"Unexpected selected installed callback")
require(results["success"] and results["numPassedTests"]==1 and results["numFailedTests"]==0 and not results.get("unhandledErrors"),"Installed callback failed")
control=None if name=="installed-positive" else "minimum-batch-body-conflict" if name=="installed-batch-control" else "minimum-readonly-credential"
report=E/(name+".observation.json" if control is None else name+".observation-"+control+".json")
o=json.loads(report.read_text());r=o["result"];p=document(g,"installedProvenance")
require(o["sourceCommit"]==g["finalHead"] and Path(o["entry"]).resolve()==Path(p["entryRealpath"]),"Wrong source or installed entry")
s=o["baseline"]["served"]
require(s["sourceCommit"]==g["finalHead"] and s["installedPayloadDigest"]==p["payloadIdentity"],"Wrong payload/source binding")
require(s["scope"]=="https://reference.invalid/acme/" and s["profile"]=="transactional" and s["history"] is False and o["baseline"]["authorityMode"]=="canonical","Wrong serving configuration")
for key in ("installedContractsDigest","policyDigest","authenticationDigest","fixtureDigest"):
    require(re.fullmatch("[0-9a-f]{64}",s[key]) is not None,"Missing serving digest")
require(s["fixtureDigest"]==sha((W/"packages/conformance/test-support/transactional/fixture-v1.json").read_bytes()),"Wrong fixture")
for ready in o["sessions"]:
    require(ready["payloadIdentity"]==p["payloadIdentity"] and ready["runtimeConfigurationIdentity"]==o["baseline"]["runtimeConfigurationIdentity"],"Session payload/runtime changed")
    require(ready["authorityMode"]=="canonical" and ready["profile"]=="transactional" and ready["history"] is False and ready["claimEligible"] is False,"Bad session declaration")
require(r["claimEligible"] is False and r["profile"]=="transactional","False qualification")
require(r["declarations"]["exactConfigurationDigest"]=="02dd7d9f35e7482c07ce1550a6572232cc48ba6858cd66f8dfc8f5384f3628be","Static route digest differs")
newid="read-update.http.command-minimum-fields";row=next(x for x in r["scenarios"] if x["id"]==newid)
ids=["m-create-a","m-replay-a","m-create-b","m-replay-b","m-create-c","m-replay-c","m-body-conflict","m-stale-update","m-retained-a","m-current","m-events"]
rawids=[x for x in ids if x!="m-retained-a"];pairs=[["m-replay-a","m-create-a","whole-json"],["m-replay-b","m-create-b","whole-json"],["m-replay-c","m-create-c","whole-json"]]
supp=o["supplemental"];count=10 if control is None else 4 if control=="minimum-batch-body-conflict" else 1
require(supp["kind"]=="command-minimum-wire-supplement" and supp["active"] is True and supp["control"]==(control or "positive"),"Wrong supplement identity")
require(supp["observedIds"]==rawids[:count] and supp["performedComparisons"]==pairs[:3 if control is None else 1 if count==4 else 0],"Incomplete actual hook/performed ledger")
require(row["requiredProfile"]=="read-update" and not row.get("cleanupError") and not row.get("prerequisiteFailure"),"Wrong inherited row or cleanup")
for exchange in row["exchanges"]:
    require(not exchange.get("transportError") and not exchange.get("harnessError") and exchange["request"]["exact"]["writeState"]=="complete","Unexpected transport/harness error")
if control is None:
    require(row["state"]=="pass" and [e["request"]["id"] for e in row["exchanges"]]==ids,"Positive incomplete")
    require(sum(x["state"]=="pass" for x in r["scenarios"])==14 and len(r["scenarios"])==256,"Wrong full positive outcomes")
    require([a["action"]["id"] for a in row["actions"]]==["m-retained-a"] and all(a["passed"] for a in row["actions"][0]["assertions"]),"Missing client receipt proof")
    require(all(a["passed"] for ex in row["exchanges"] for a in ex["assertions"]),"Positive assertion failed")
    source=row["exchanges"][8]["sourceReceiptObservation"]
    require(source["action"]=="m-create-a" and source["receipt"]==supp["originalRoot"] and source["responseDigest"]==supp["originalResponseDigest"] and source["pointer"]=="/id" and source["credentialRef"]=="operator","Wrong retained source binding")
    require(source["session"]==row["exchanges"][8]["servedSession"],"Unexpected reopen")
    require(r["artifacts"]["manifestDigest"]==sha((W/"packages/conformance/test-support/transactional/manifest-v1.json").read_bytes()),"Wrong positive manifest")
else:
    d=config["controlOracle"];require(sha(Path(d["path"]).read_bytes())==d["sha256"],"Control oracle changed")
    oracle=next(x for x in json.loads(Path(d["path"]).read_text()) if x["control"]==control)
    require(r["selectedScenarioIds"]==[newid] and len(r["scenarios"])==1 and row["state"]=="fail","Wrong selected negative row")
    require([e["request"]["id"] for e in row["exchanges"]]==ids[:count] and not row.get("actions"),"Wrong negative prefix")
    require(all(a["passed"] for ex in row["exchanges"][:-1] for a in ex["assertions"]),"Earlier negative prefix failure")
    last=row["exchanges"][-1]
    require(last["response"]["status"]==(409 if count==4 else 403),"Wrong actual refusal")
    require([a["id"] for a in last["assertions"]]==oracle["allAssertionIds"] and [a["id"] for a in last["assertions"] if not a["passed"]]==oracle["expectedFailedIds"],"Wrong intended failure IDs")
    require(r["artifacts"]["manifestDigest"]==oracle["compactSha256"],"Wrong selected derivative")
actual=sum(len(x["exchanges"]) for x in r["scenarios"])
require(actual==spec["workload"]["httpExpected"] and len(o["sessions"])==spec["workload"]["children"]-1,"Actual cohort counts differ")
logs=[]
for line in (E/(name+".log")).read_text().splitlines():
    line=re.sub(r"\x1b\[[0-9;]*m","",line).strip()
    if line.startswith("{"):
        try:value=json.loads(line)
        except json.JSONDecodeError:continue
        if value.get("kind")==("installed-tx-fourteen-partial-summary" if control is None else "installed-tx-command-minimum-control"):logs.append(value)
require(len(logs)==1 and logs[0]["sourceCommit"]==g["finalHead"] and logs[0]["childLaunches"]==spec["workload"]["children"],"Missing child summary")
if control is None:require(logs[0]["closedChildren"]==22,"Positive children not all closed")
# Negative callback asserts each child closed before its summary; supervisor
# independently checks process-group/sample absence and owned temporary roots.
review={"candidateMatches":True,"rootReviewRequired":True,"command":name,"actualHttp":actual,"reportedChildren":logs[0]["childLaunches"],"payloadIdentity":p["payloadIdentity"],"reportSha256":sha(report.read_bytes()),"resultsSha256":sha((E/(name+".results.json")).read_bytes()),"logSha256":sha((E/(name+".log")).read_bytes()),"limitations":"Partial inherited evidence only. Actual raw equality is asserted in frozen app/helper before report projection; report does not contain raw bodies. Supervisor/build/tar attribution is not child-attested Git. Detached/short-lived process sampling limits remain explicit."}
with (E/(name+".review-candidate.json")).open("x") as f:f.write(json.dumps(review,indent=2)+"\n")
print(json.dumps(review))
