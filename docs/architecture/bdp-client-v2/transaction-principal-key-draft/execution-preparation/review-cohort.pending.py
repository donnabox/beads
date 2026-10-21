"""Adjudicate a genuine principal cohort. No product import or fabricated response.
Actual negative test/supervisor exit1 remains unchanged; this separate review may exit0."""
from pathlib import Path
import json,os,re,sys,math
from gate import P,D,W,E,load_gate,completed,document,require,sha

require(len(sys.argv)==2 and sys.argv[1] in ("positive","principal-collapse"),"Exactly one cohort required")
cohort=sys.argv[1];negative=cohort=="principal-collapse";name="installed-"+cohort
g,_=load_gate("review-"+cohort);terminal=completed(name)
require(Path(g["reportContract"]["path"]).resolve()==(P/"report-adapter-binding.json").resolve(),"Wrong adapter binding path")
binding=document(g,"reportContract")
for key in ("reportContract","ledger","selection","integrationSource"):
 pin=binding[key];require(sha(Path(pin["path"]).read_bytes())==pin["sha256"],"Adapter input changed: "+key)
contract=json.loads(Path(binding["reportContract"]["path"]).read_text())
ledger=json.loads(Path(binding["ledger"]["path"]).read_text())
selection=json.loads(Path(binding["selection"]["path"]).read_text())
test=selection["file"]
require(sha((W/test).read_bytes())==binding["testSha256"],"Reviewed test changed")
spec=json.loads((P/"commands.json").read_text())["commands"][name]
require(spec["expectedExit"]==(1 if negative else 0),"Wrong expected process result")
require(terminal["exit"]==terminal["supervisorExit"]==spec["expectedExit"],"Actual process exit differs")
start=json.loads((E/(name+".started.json")).read_text())
require(start["head"]==g["finalHead"] and start["workspaceDirty"] is False,"Caller attribution differs")
require(start["argv"]==spec["argv"] and start["cwd"]==str(W),"Wrong actual invocation")

results=E/(name+".results.json");r=json.loads(results.read_text())
require(len(r["testResults"])==1,"Unexpected test file fanout")
f=r["testResults"][0]
require(Path(f["name"]).resolve()==(W/test).resolve(),"Wrong executed file")
cases=f["assertionResults"]
require(len(cases)==1 and cases[0]["fullName"]==selection["fullName"],"Wrong case selection")
case=cases[0];require(case["status"]==("failed" if negative else "passed"),"Wrong actual case outcome")
require(r["success"] is (not negative),"Wrong suite result")
require(r["numTotalTests"]==1 and r["numPassedTests"]==(0 if negative else 1) and r["numFailedTests"]==(1 if negative else 0),"Wrong test totals")
for key in ("numPendingTests","numTodoTests","numRuntimeErrorTestSuites","unhandledErrors","errors"):
 require(not r.get(key),"Unexpected runtime/pending error")
require(not f.get("message") and not f.get("failureMessage"),"Test-file failure outside selected assertion")
messages=case.get("failureMessages",[])
if negative:
 require(len(messages)==1,"Exactly one assertion failure required")
 # Pinned Vitest reporter shape. The real failure must be expected200 vs actual409.
 first=re.sub(r"\x1b\[[0-9;]*m","",messages[0]).splitlines()[0]
 require(first=="AssertionError: principal-key:b-independent-create:completed: expected 409 to be 200 // Object.is equality","Not the intended named assertion")
else:require(messages==[],"Positive assertion failure")

out=E/(name+".observation.json");o=json.loads(out.read_text())
stdout=[]
for line in (E/(name+".log")).read_text().splitlines():
 line=re.sub(r"\x1b\[[0-9;]*m","",line).strip()
 if not line.startswith("{"):continue
 try:value=json.loads(line)
 except json.JSONDecodeError:continue
 if isinstance(value,dict) and value.get("observation")==contract["label"]:stdout.append(value)
require(len(stdout)==1,"Exactly one genuine stdout report required")
logged=stdout[0]
require(set(o)==set(logged)==set(contract["topLevelFields"]),"Unexpected top-level report fields")
require({k:v for k,v in o.items() if k!="elapsedMs"}=={k:v for k,v in logged.items() if k!="elapsedMs"},"File/stdout observation disagreement")
require(type(o["elapsedMs"]) in (int,float) and math.isfinite(o["elapsedMs"]) and 0<=o["elapsedMs"]<=logged["elapsedMs"]<=60000,"Invalid elapsed record")
require(o["observation"]==contract["label"] and o["mode"]==("collapse-b-to-a" if negative else "positive"),"Wrong report identity")
require(o["outcome"]==("failed" if negative else "passed") and o["failureCount"]==(1 if negative else 0),"Unexpected failure count/outcome")
require(o["deadlineExceeded"] is False and o["cleanupExceeded"] is False and o["claimEligible"] is False and o["workspaceDirty"] is False,"Invalid deadline/claim/source state")
require(o["retention"]=="response fingerprints/allowlisted fields and validated negative Problem; full receipt/Resource/Event equality in memory; no request wire or credential-bearing objects","Unexpected retention text")
require(o["startupLimitation"]=="Pre-ready PID unknown; independent supervisor owned-process evidence required","Unexpected startup metadata")

p=document(g,"installedProvenance")
require(p["sourceCommit"]==g["finalHead"] and p["sourceTree"]==g["finalTree"],"Installed source differs")
entry=Path(p["entryRealpath"]).resolve(strict=True)
require(Path(o["entry"]).is_absolute() and Path(o["entry"]).resolve(strict=True)==entry and entry.is_relative_to((E/"consumer").resolve()),"Wrong external entry")
require(o["declaredSourceCommit"]==g["finalHead"] and o["entrySha256"]==p["entrySha256"]==sha(entry.read_bytes()),"Wrong source/entry bytes")
overrides=start["bdpOverrides"]
require(overrides=={"BDP_TRANSACTIONAL_INSTALLED_ENTRY":str(entry),"BDP_TRANSACTIONAL_SOURCE_COMMIT":g["finalHead"],"BDP_TRANSACTIONAL_PRINCIPAL_KEY_MODE":o["mode"],"BDP_TRANSACTIONAL_PRINCIPAL_KEY_OUTPUT":str(out)},"Wrong safe environment")
b=o["binding"];require(isinstance(b,dict) and set(b)=={"served","operationTargets","runtimeConfigurationIdentity","authorityMode"},"Unexpected binding fields")
s=b["served"];keys={"scope","profile","history","sourceCommit","installedPayloadDigest","installedContractsDigest","policyDigest","authenticationDigest","fixtureDigest","fixtureId","capabilities"}
require(isinstance(s,dict) and set(s)==keys,"Unexpected served binding fields")
hex64=lambda v:isinstance(v,str) and re.fullmatch("[0-9a-f]{64}",v) is not None
require(b["authorityMode"]=="canonical" and hex64(b["runtimeConfigurationIdentity"]),"Invalid runtime identity")
require(s["scope"]==ledger["scope"] and s["profile"]=="transactional" and s["history"] is False and s["sourceCommit"]==g["finalHead"] and s["installedPayloadDigest"]==p["payloadIdentity"],"Wrong served identity")
for k in ("installedPayloadDigest","installedContractsDigest","policyDigest","authenticationDigest","fixtureDigest"):require(hex64(s[k]),"Invalid served digest")
fixture_path=W/"packages/conformance/test-support/transactional/fixture-v1.json"
fixture=json.loads(fixture_path.read_text())
require(s["fixtureDigest"]==sha(fixture_path.read_bytes()) and s["fixtureId"]==fixture["id"] and s["capabilities"]==sorted(fixture["capabilities"]),"Fixture identity differs")
targets={"createBead":"create-bead","updateBead":"update-bead","deleteBead":"delete-bead","createLink":"create-link","updateLink":"update-link","deleteLink":"delete-link","putAlias":"put-alias","deleteAlias":"delete-alias","updateWhere":"update-where","deleteWhere":"delete-where","batch":"batch","sequence":"sequence"}
require(b["operationTargets"]=={k:ledger["scope"]+"operations/"+v for k,v in targets.items()},"Wrong operation target inventory")
require(o["descriptorSchemaUtf8"]==ledger["configuration"]["artifacts"],"Unexpected artifact bytes/fields")
require(o["routeDigest"]==ledger["routeSha256"],"Wrong route binding")

responses=o["completedResponses"];count=2 if negative else 7
require(type(o["attempted"]) is int and type(o["exchanges"]) is int and o["attempted"]==o["exchanges"]==len(responses)==count,"Wrong actual attempts/responses")
expected=ledger["requests"][:count];problem=ledger["cohorts"][1]["actualExpectedProblem"]
for x,row in zip(responses,expected):
 require(isinstance(x,dict) and set(x)==set(contract["completedResponseFields"]),"Unexpected response fields")
 for actual,declared in [("index","index"),("label","label"),("method","method"),("url","url"),("credentialRef","credentialRef"),("requestBodyBytes","bodyBytes"),("requestBodySha256","bodySha256")]:
  require(x[actual]==row[declared],"Request ledger fingerprint mismatch")
 require(x["writeState"]=="complete","Incomplete request write")
 expected_status=409 if negative and row["index"]==2 else row["status"]
 require(type(x["status"]) is int and x["status"]==expected_status,"Wrong actual status")
 require(type(x["bodyOctets"]) is int and 0<x["bodyOctets"]<=32768 and x["contentLength"]==str(x["bodyOctets"]),"Invalid actual response length")
 require(hex64(x["bodySha256"]) and hex64(x["responseWireSha256"]),"Missing response fingerprints")
 require(x["contentType"]==("application/problem+json" if expected_status==409 else "application/json"),"Wrong media")
 require(x["validatedNegativeProblem"]==(problem if expected_status==409 else None),"Wrong safe validated Problem")
 for k in ("scopeEpoch","authorizationView","scopePosition"):require(isinstance(x[k],str) and x[k] and len(x[k].encode())<=65536,"Missing bounded observed context")
 require(x["scopeEpoch"]==responses[0]["scopeEpoch"],"Epoch changed")
 require(x["scopePosition"]==("pos-1" if negative else row["currentPosition"]),"Wrong serving position")
if negative:
 require(responses[1]["authorizationView"]==responses[0]["authorizationView"],"Collapsed principal context differs")
 require(o["passedAssertions"]==["principal-key:a-create:completed"] and o["currentAssertion"]==selection["negative"]["expectedCurrentAssertion"],"Unexpected failure progression")
else:
 expected_pass=["principal-key:a-create:completed","principal-key:b-independent-create:completed","principal-key:a-rotated-replay:completed","principal-key:b-replay:completed","principal-key:a-conflict:refused","principal-key:current-complete","principal-key:events-complete"]
 require(o["passedAssertions"]==expected_pass and o["currentAssertion"]=="principal-key:complete","Incomplete positive progression")
 require(responses[2]["authorizationView"]==responses[4]["authorizationView"]==responses[0]["authorizationView"] and responses[3]["authorizationView"]==responses[1]["authorizationView"],"Principal-specific replay context differs")
 require(responses[2]["bodySha256"]==responses[0]["bodySha256"] and responses[3]["bodySha256"]==responses[1]["bodySha256"],"Retained receipt fingerprints differ")

require(o["launchAttempted"] is True and o["startupCompleted"] is True and o["closed"] is True and o["stopState"]=="stopped" and o["childExitCode"]==0 and o["childSignalCode"] is None,"Child was not proven normally stopped")
pid=o["childPid"];require(type(pid) is int and pid>0,"Missing child PID")
require(terminal["remainingAfterCleanup"]==[] and terminal["sampledPidsStillPresent"]=={} and terminal["runtimeDirectoriesStillPresent"]==[],"Supervisor cleanup incomplete")
sampled=terminal["sampledPids"];require(isinstance(sampled,list) and sampled and all(type(n) is int and n>0 for n in sampled),"Missing actual process samples")
require(terminal["processGroup"] in sampled,"Supervisor group leader not sampled")
absent=[]
for n in sorted(set(sampled+[pid])):
 try:os.kill(n,0)
 except ProcessLookupError:absent.append(n)
 else:raise RuntimeError("Owned child/sampled PID remains present")
directory=Path(o["directory"]);tmp=(E/(name+".tmp")).resolve()
require(directory.is_absolute() and directory.resolve().parent==tmp and directory.name.startswith("bdp-principal-key-installed-") and o["directoryAbsent"] is True and not directory.exists(),"Owned store remains or is outside expected root")
record={"candidateMatches":True,"command":name,"actualTestExit":terminal["exit"],"actualSupervisorExit":terminal["supervisorExit"],"adjudication":"intended named failure" if negative else "positive passed","actualHttp":count,"actualChildren":1,"reportedClosedChildren":1,"sourceCommit":g["finalHead"],"sourceTree":g["finalTree"],"payloadIdentity":p["payloadIdentity"],"entryRealpath":str(entry),"entrySha256":p["entrySha256"],"runtimeConfigurationIdentity":b["runtimeConfigurationIdentity"],"authenticationIdentity":s["authenticationDigest"],"childPid":pid,"childPidWasSampled":pid in sampled,"independentlyAbsentPids":absent,"storeDirectoryAbsent":str(directory),"rootReviewRequired":True,"claimEligible":False,"observationSha256":sha(out.read_bytes()),"resultsSha256":sha(results.read_bytes()),"rawLogSha256":sha((E/(name+".log")).read_bytes()),"terminalSha256":sha((E/(name+".exit.json")).read_bytes()),"limitations":"Source-bound serializer/full schema/body equality are in-process assertions. Retained request role/body fields are the checked authored ledger, not packet capture; response fingerprints are not retained bodies. Caller Git attribution is externally bound, not child-attested. Samples can miss short-lived or detached descendants; actual returned PID is independently checked regardless of sampling. No public qualification."}
with (E/("review-"+cohort+".review-candidate.json")).open("x") as f:f.write(json.dumps(record,indent=2)+"\n")
print(json.dumps(record))
