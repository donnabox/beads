"""Reconcile one actual installed changefeed result; no product import or response fabrication."""
from pathlib import Path
import json,os,re
from gate import P,W,E,load_gate,completed,document,require,sha
name="installed-positive"
g,_=load_gate("installed-review");completed(name)
spec=json.loads((P/"commands.json").read_text())["commands"][name]
r=json.loads((E/(name+".json")).read_text())
actual=[a for f in r["testResults"] for a in f["assertionResults"] if a["status"] not in ("pending","skipped","todo")]
require(len(actual)==1 and actual[0]["status"]=="passed" and [a["fullName"] for a in actual]==spec["selectedFullNames"],"Wrong case/result")
require(r["success"] and r["numPassedTests"]==1 and r["numFailedTests"]==0 and not r.get("unhandledErrors") and not r.get("errors"),"Test failure")
records=[]
for line in (E/(name+".log")).read_text().splitlines():
    line=re.sub(r"\x1b\[[0-9;]*m","",line).strip()
    if not line.startswith("{"):continue
    try:value=json.loads(line)
    except json.JSONDecodeError:continue
    if value.get("observation")=="private-installed-changefeed-parameter-context":records.append((value,line))
require(len(records)==1,"Exactly one observation required")
o,line=records[0];p=document(g,"installedProvenance")
require(o["outcome"]=="passed" and not o["deadlineExceeded"] and o["failureCount"]==0 and o["claimEligible"] is False and o["workspaceDirty"] is False,"Invalid outcome")
require(o["declaredSourceCommit"]==g["finalHead"] and Path(o["entry"]).resolve()==Path(p["entryRealpath"]) and o["entrySha256"]==p["entrySha256"],"Wrong source/entry")
b=o["binding"];s=b["served"]
require(b["authorityMode"]=="canonical" and re.fullmatch("[0-9a-f]{64}",b["runtimeConfigurationIdentity"]) is not None,"Wrong runtime binding")
require(s["sourceCommit"]==g["finalHead"] and s["installedPayloadDigest"]==p["payloadIdentity"] and s["scope"]=="https://reference.invalid/acme/" and s["profile"]=="transactional" and s["history"] is False,"Wrong served binding")
for key in ("installedContractsDigest","policyDigest","authenticationDigest","fixtureDigest"):
    require(re.fullmatch("[0-9a-f]{64}",s[key]) is not None,"Invalid configuration digest")
require(s["fixtureDigest"]==sha((W/"packages/conformance/test-support/transactional/fixture-v1.json").read_bytes()),"Wrong fixture")
require(o["artifactBytes"]==336,"Wrong artifacts")
requestpin=json.loads((P/"commands.json").read_text())["requests"]
require(sha(Path(requestpin["path"]).read_bytes())==requestpin["sha256"],"Request ledger changed")
requests=json.loads(Path(requestpin["path"]).read_text())["installed"]["requests"]
responses=o["completedResponses"]
require(o["attempted"]==o["exchanges"]==len(responses)==len(requests)==10,"Wrong dispatch/response count")
# Match the literal issued-token shape, without inventing a durable checkpoint.
# Full page/checkpoint derivation and schema/body equality are in-memory test assertions.
future=responses[2]["url"]
require(re.fullmatch(r"https://reference\.invalid/acme/changes/\?after=cg_[A-Za-z0-9_-]{43}_[A-Za-z0-9_-]{43}_2",future) is not None,"Wrong private future target")
expected=[]
for request in requests:
    target=request["target"]
    if "<derived-future-checkpoint>" in target:target=future
    expected.append((request["ordinal"],request["method"],target,request["accept"],request.get("lastEventId"),request["expectedStatus"]))
require([(x["ordinal"],x["method"],x["url"],x["accept"],x["lastEventId"],x["status"]) for x in responses]==expected,"Wrong actual request map")
context=["bdp-scope-epoch","bdp-authorization-view","bdp-scope-position"]
known=[responses[1]["headers"][k] for k in context]
require(all(known) and known[2]=="pos-1","Missing comparator context")
problem={"type":"https://github.com/gastownhall/bdp/problems/request","code":"invalid-parameter","retry":"never","status":400}
problem_bytes=json.dumps(problem,separators=(",",":")).encode()
for i,x in enumerate(responses):
    require(type(x["bodyOctets"]) is int and 0<=x["bodyOctets"]<=32768,"Invalid response size")
    for key in ("wireSha256","bodySha256"):require(re.fullmatch("[0-9a-f]{64}",x[key]) is not None,"Missing fingerprint")
    require(x["headers"]["x-bdp-development"]=="non-attesting","Missing private marker")
    require([x["headers"][k] for k in context]==(known if i<8 else [None,None,None]),"Wrong context tuple")
    if x["status"]==400:
        require(x["headers"]["content-type"]=="application/problem+json","Wrong problem media")
        if x["method"]=="HEAD":require(x["bodySha256"]==sha(b""),"Wrong HEAD body fingerprint")
        if 2<=i<8:
            require(x["headers"]["cache-control"]=="private, no-store","Wrong private cache policy")
            require(all(x["headers"][k] is None for k in ["retry-after","www-authenticate","x-bdp-development-limit"]),"Unexpected refusal metadata")
            require(x["headers"]["content-length"]==str(len(problem_bytes)),"Wrong representation length")
            require(x["bodyOctets"]==(0 if x["method"]=="HEAD" else len(problem_bytes)),"Wrong received length")
    if x["status"]==401:
        require(x["headers"]["www-authenticate"]=='Bearer realm="bdp-transaction-development"',"Wrong anonymous challenge")
require(o["closed"] is True and o["startupCompleted"] is True and o["childExitCode"]==0 and o["childSignalCode"] is None,"Child did not close normally")
pid=o["childPid"];require(type(pid) is int and pid>0,"Missing child PID")
try:os.kill(pid,0)
except ProcessLookupError:pass
else:raise RuntimeError("Child remains alive")
directory=Path(o["directory"])
require(directory.is_relative_to(E/(name+".tmp")) and directory.name.startswith("bdp-changefeed-context-installed-") and o["directoryAbsent"] is True and not directory.exists(),"Owned store remains")
with (E/"installed-observation.json").open("x") as f:f.write(line+"\n")
result={"candidateMatches":True,"rootReviewRequired":True,"actualHttp":10,"actualChildren":1,"childPidIndependentlyAbsent":pid,"storeDirectoryAbsent":str(directory),"sourceCommit":g["finalHead"],"sourceTree":g["finalTree"],"payloadIdentity":p["payloadIdentity"],"runtimeConfigurationIdentity":b["runtimeConfigurationIdentity"],"servedConfigurationSha256":sha(json.dumps(s,separators=(",",":"),ensure_ascii=False).encode()),"observationSha256":sha((E/"installed-observation.json").read_bytes()),"resultsSha256":sha((E/(name+".json")).read_bytes()),"claimEligible":False,"limitations":"Full schemas/bodies and exact-request serializer fields asserted in process; log retains safe headers/fingerprints only. Method is authored ledger plus returned response mapping, not invented exactRequest.method. Git attribution is supervisor verified, not child attested; no qualification or public admission."}
with (E/"installed-review.json").open("x") as f:f.write(json.dumps(result,indent=2)+"\n")
print(json.dumps(result))
