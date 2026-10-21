"""Read-only reconciliation of one genuine helper log/result; no matrix report assumptions."""
from pathlib import Path
import json,os,re,sys
from gate import P,W,E,load_gate,completed,document,require,sha
name=sys.argv[1]
require(name in ("installed-compatible-minimum","installed-snapshot-wait"),"Unknown helper")
g,_=load_gate(name+"-review");completed(name)
spec=json.loads((P/"commands.json").read_text())["commands"][name]
r=json.loads((E/(name+".results.json")).read_text())
actual=[(str(Path(s["name"]).resolve().relative_to(W)),a["fullName"],a["status"]) for s in r["testResults"] for a in s["assertionResults"] if a["status"] not in ("pending","skipped","todo")]
require(actual==[(x["path"],x["fullName"],"passed") for x in spec["selectedCases"]],"Wrong helper selection")
require(r["success"] and r["numPassedTests"]==1 and r["numFailedTests"]==0 and not r.get("unhandledErrors"),"Helper failed")
records=[]
for line in (E/(name+".log")).read_text().splitlines():
    line=re.sub(r"\x1b\[[0-9;]*m","",line).strip()
    if not line.startswith("{"):continue
    try:value=json.loads(line)
    except json.JSONDecodeError:continue
    if value.get("observation")==spec["helperObservation"]:records.append((value,line))
require(len(records)==1,"Missing/duplicate helper observation")
o,line=records[0];p=document(g,"installedProvenance")
require(o["outcome"]=="passed" and o["aborted"] is False and o["claimEligible"] is False and o["workspaceDirty"] is False,"Invalid helper outcome/declaration")
require(o["declaredSourceCommit"]==g["finalHead"] and Path(o["entry"]).resolve()==Path(p["entryRealpath"]) and o["entrySha256"]==p["entrySha256"],"Wrong helper source/entry")
b=o["binding"];s=b["served"]
require(b["authorityMode"]=="canonical" and re.fullmatch("[0-9a-f]{64}",b["runtimeConfigurationIdentity"]) is not None,"Wrong runtime binding")
require(s["sourceCommit"]==g["finalHead"] and s["installedPayloadDigest"]==p["payloadIdentity"] and s["scope"]=="https://reference.invalid/acme/" and s["profile"]=="transactional" and s["history"] is False,"Wrong served binding")
for key in ("installedContractsDigest","policyDigest","authenticationDigest","fixtureDigest"):
    require(re.fullmatch("[0-9a-f]{64}",s[key]) is not None,"Missing helper digest")
require(s["fixtureDigest"]==sha((W/"packages/conformance/test-support/transactional/fixture-v1.json").read_bytes()),"Wrong fixture bytes")
scope="https://reference.invalid/acme/"
if name=="installed-compatible-minimum":
    require(o["closed"] is True and o["childExitCode"]==0,"Collection child not closed normally")
    children=[o["childPid"]]
    expected=[("POST","operations/batch",200),("POST","operations/batch",200),("GET","beads/never",404),("HEAD","beads/never?view=properties",404),("GET","links/never",404),("HEAD","links/never?view=properties",404),("GET","beads/a",404),("GET","beads/a?view=properties",404),("GET","links/l",404),("GET","links/l?view=properties",500)]
else:
    require(len(o["children"])==1 and o["children"][0]["exitCode"]==0 and o["children"][0]["signalCode"] is None,"Snapshot child not closed normally")
    children=[o["children"][0]["pid"]];expected=[("GET","snapshot",200),("GET","snapshot",503)]
responses=o["completedResponses"]
require(o["exchanges"]==len(responses)==spec["workload"]["httpExpected"],"Wrong completed-response count")
require([(x["method"],x["url"],x["status"]) for x in responses]==[(m,scope+url,status) for m,url,status in expected],"Wrong helper method/URL/status inventory")
for x in responses:
    require(type(x["responseBytes"]) is int and x["responseBytes"]>=0 and re.fullmatch("[0-9a-f]{64}",x["wireSha256"]) is not None,"Invalid response fingerprint")
    if name=="installed-compatible-minimum":require(re.fullmatch("[0-9a-f]{64}",x["bodySha256"]) is not None,"Invalid body fingerprint")
    if x["method"]=="HEAD":require(x["responseBytes"]==0,"Unexpected HEAD body")
for pid in children:
    require(type(pid) is int and pid>0,"Missing child PID")
    try:os.kill(pid,0)
    except ProcessLookupError:pass
    else:raise RuntimeError("Reported helper child PID still present")
# completed() separately requires the supervised owned temp prefix/group/sampled PIDs absent.
observation=E/(name+".helper-observation.json")
with observation.open("x") as f:f.write(line+"\n")
result={"candidateMatches":True,"rootReviewRequired":True,"command":name,"actualHttp":len(responses),"reportedChildren":len(children),"childPidsIndependentlyAbsent":children,"sourceCommit":g["finalHead"],"sourceTree":g["finalTree"],"payloadIdentity":p["payloadIdentity"],"observationSha256":sha(observation.read_bytes()),"resultsSha256":sha((E/(name+".results.json")).read_bytes()),"logSha256":sha((E/(name+".log")).read_bytes()),"claimEligible":False,"limitations":"Operator-only helper regression. Actual logged response fingerprints retained; no raw bodies or exact headers reconstructed. Supervisor source attribution is not child-attested Git. No reader/anonymous, whole-profile or installed fault qualification."}
with (E/(name+".review-candidate.json")).open("x") as f:f.write(json.dumps(result,indent=2)+"\n")
print(json.dumps(result))
