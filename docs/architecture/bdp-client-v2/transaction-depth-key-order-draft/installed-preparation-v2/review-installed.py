"""Reconcile one actual installed depth/key result; no product import or response fabrication."""
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
    if value.get("observation")=="private-installed-depth-key-order":records.append((value,line))
require(len(records)==1,"Exactly one observation required")
o,line=records[0];p=document(g,"installedProvenance")
require(o["outcome"]=="passed" and not o["deadlineExceeded"] and not o["cleanupExceeded"] and o["failureCount"]==0 and o["claimEligible"] is False and o["workspaceDirty"] is False,"Invalid outcome")
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
requests=json.loads(Path(requestpin["path"]).read_text())["rows"]
responses=o["completedResponses"]
require(o["attempted"]==o["exchanges"]==len(responses)==len(requests)==13,"Wrong dispatch/response count")
require([(x["ordinal"],x["label"],x["method"],x["url"],x["accept"],x["key"],x["requestBodyBytes"],x["requestBodySha256"],x["status"]) for x in responses]==[(x["index"],x["label"],x["method"],x["url"],x["accept"],x["key"],x["bodyBytes"],x["bodySha256"],x["status"]) for x in requests],"Wrong actual request map")
require(o["routeBindings"]==[{"method":m,"url":u} for m,u in dict.fromkeys((x["method"],x["url"]) for x in requests)],"Wrong exact route bindings")
context=["bdp-scope-epoch","bdp-authorization-view","bdp-scope-position"]
known=[responses[0]["headers"][k] for k in context[:2]]
require(all(isinstance(x,str) and x for x in known),"Missing seed context")
for x,r in zip(responses,requests):
    require(type(x["bodyOctets"]) is int and 0<=x["bodyOctets"]<=131072,"Invalid response size")
    for key in ("wireSha256","bodySha256"):require(re.fullmatch("[0-9a-f]{64}",x[key]) is not None,"Missing fingerprint")
    h=x["headers"]
    require(h["x-bdp-development"]=="non-attesting" and h["cache-control"]=="private, no-store","Wrong private metadata")
    require([h[k] for k in context]==([*known,r["contextPosition"]] if r["contextPosition"] else [None,None,None]),"Wrong observed context")
    if r["method"]=="POST":require(h["vary"]=="Accept","Wrong negotiation metadata")
    if r["status"]==500:
        require(x["bodyOctets"]==0 and x["bodySha256"]==sha(b"") and h["content-length"]=="0","Fresh depth not bodyless")
        require(h["x-bdp-development-limit"]=="unsupported-transaction-composition" and h["content-type"] is None and h["retry-after"] is None,"Wrong fresh depth metadata")
    else:
        require(h["x-bdp-development-limit"] is None,"Unexpected depth marker")
        require(h["content-type"]==("application/problem+json" if r["status"]==409 else "application/json"),"Wrong media")
        if r["status"]==409:require(h["retry-after"] is None and h["www-authenticate"] is None,"Unexpected conflict headers")
# Fingerprint replay equality supplements (does not replace) in-process full receipt equality.
for seed,replay in [(0,2),(6,8)]:
    require(responses[seed]["bodySha256"]==responses[replay]["bodySha256"],"Retained receipt fingerprint differs")
require(o["closed"] is True and o["startupCompleted"] is True and o["childExitCode"]==0 and o["childSignalCode"] is None,"Child did not close normally")
pid=o["childPid"];require(type(pid) is int and pid>0,"Missing child PID")
try:os.kill(pid,0)
except ProcessLookupError:pass
else:raise RuntimeError("Child remains alive")
directory=Path(o["directory"])
require(directory.is_relative_to(E/(name+".tmp")) and directory.name.startswith("bdp-depth-key-installed-") and o["directoryAbsent"] is True and not directory.exists(),"Owned store remains")
with (E/"installed-observation.json").open("x") as f:f.write(line+"\n")
result={"candidateMatches":True,"rootReviewRequired":True,"actualHttp":13,"actualChildren":1,"childPidIndependentlyAbsent":pid,"storeDirectoryAbsent":str(directory),"sourceCommit":g["finalHead"],"sourceTree":g["finalTree"],"payloadIdentity":p["payloadIdentity"],"runtimeConfigurationIdentity":b["runtimeConfigurationIdentity"],"servedConfigurationSha256":sha(json.dumps(s,separators=(",",":"),ensure_ascii=False).encode()),"observationSha256":sha((E/"installed-observation.json").read_bytes()),"resultsSha256":sha((E/(name+".json")).read_bytes()),"claimEligible":False,"limitations":"Full schemas/bodies and exact-request serializer fields asserted in process; log retains safe headers/fingerprints only. Method is authored ledger plus returned response mapping, not invented exactRequest.method. Git attribution is supervisor verified, not child attested; no qualification or public admission."}
with (E/"installed-review.json").open("x") as f:f.write(json.dumps(result,indent=2)+"\n")
print(json.dumps(result))
