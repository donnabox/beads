"""Adjudicate actual guarded-batch cohort records; no product imports or execution.
Negative exit1 is retained; adapter exit0 means only intended failure with prerequisites."""
from pathlib import Path
import json,os,re,sys,math
from gate import P,D,W,E,load_gate,completed,document,require,sha
require(len(sys.argv)==2 and sys.argv[1] in ("positive","unguarded-contenders"),"Exactly one cohort required")
cohort=sys.argv[1];negative=cohort=="unguarded-contenders";name="installed-"+cohort;count=3 if negative else 7
g,_=load_gate("review-"+cohort);terminal=completed(name)
require(Path(g["reportContract"]["path"]).resolve()==(P/"report-adapter-binding.json").resolve(),"Wrong adapter binding")
binding=document(g,"reportContract")
for key in ("reportContract","ledger","selection","requests","sourceBinding"):
 pin=binding[key];require(sha(Path(pin["path"]).read_bytes())==pin["sha256"],"Adapter input changed: "+key)
contract=json.loads(Path(binding["reportContract"]["path"]).read_text())
ledger=json.loads(Path(binding["ledger"]["path"]).read_text())
selection=json.loads(Path(binding["selection"]["path"]).read_text())
request_ledger=json.loads(Path(binding["requests"]["path"]).read_text())
test=selection["path"];scope=ledger["runtimeConfiguration"]["scope"]
require(sha((W/test).read_bytes())==binding["testSha256"],"Reviewed test changed")
for path,h in binding["referenceSourceSha256"].items():require(sha((W/path).read_bytes())==h,"Reference source changed")
spec=json.loads((P/"commands.json").read_text())["commands"][name]
require(terminal["exit"]==terminal["supervisorExit"]==spec["expectedExit"]==(1 if negative else 0),"Wrong actual process exit")
start=json.loads((E/(name+".started.json")).read_text())
require(start["head"]==g["finalHead"] and start["workspaceDirty"] is False and start["argv"]==spec["argv"] and start["cwd"]==str(W),"Wrong caller/invocation")
results=E/(name+".results.json");r=json.loads(results.read_text())
require(len(r["testResults"])==1,"Unexpected test file fanout")
f=r["testResults"][0];cases=f["assertionResults"]
require(Path(f["name"]).resolve()==(W/test).resolve() and len(cases)==1 and cases[0]["fullName"]==selection["title"],"Wrong exact selection")
case=cases[0]
require(case["status"]==("failed" if negative else "passed") and r["success"] is (not negative),"Wrong case result")
require(r["numTotalTests"]==1 and r["numPassedTests"]==(0 if negative else 1) and r["numFailedTests"]==(1 if negative else 0),"Wrong totals")
for k in ("numPendingTests","numTodoTests","numRuntimeErrorTestSuites","unhandledErrors","errors"):require(not r.get(k),"Unexpected runtime/pending error")
require(not f.get("message") and not f.get("failureMessage"),"Failure outside selected assertion")
messages=case.get("failureMessages",[])
require(len(messages)==(1 if negative else 0),"Unexpected failure count")
if negative:
 first=re.sub(r"\x1b\[[0-9;]*m","",messages[0]).splitlines()[0]
 require(first=="AssertionError: guarded-batch:one-retained-revision-failure: expected +0 to be 1 // Object.is equality" or first=="AssertionError: guarded-batch:one-retained-revision-failure: expected 0 to be 1 // Object.is equality","Wrong named negative failure")

out=E/(name+".observation.json")
require(0<out.stat().st_size<=65536,"Report size bound")
o=json.loads(out.read_text());logs=[]
for line in (E/(name+".log")).read_text().splitlines():
 line=re.sub(r"\x1b\[[0-9;]*m","",line).strip()
 if not line.startswith("{"):continue
 try:v=json.loads(line)
 except json.JSONDecodeError:continue
 if isinstance(v,dict) and v.get("label")==contract["label"]:logs.append(v)
require(len(logs)==1,"One genuine stdout report required");logged=logs[0]
require(set(o)==set(logged)==set(binding["topLevelFields"]),"Unexpected report fields")
require({k:v for k,v in o.items() if k!="elapsedMs"}=={k:v for k,v in logged.items() if k!="elapsedMs"},"File/stdout disagree")
def number(v):return type(v) in (int,float) and math.isfinite(v)
require(number(o["elapsedMs"]) and number(logged["elapsedMs"]) and 0<=o["elapsedMs"]<=logged["elapsedMs"]<=60000,"Elapsed bound")
require(o["label"]==contract["label"] and o["mode"]==cohort and o["executionMode"]=="external-installed","Wrong report identity")
require(o["sourceAttribution"]=="caller label; supervisor must verify source/package equality" and o["workspaceDirty"] is False,"Wrong source attribution")
require(o["retention"]=="Full parsed receipt/Resource/Event comparisons occur in memory; response fingerprints are not retained full-wire evidence.","Unexpected retention declaration")
require(o["outcome"]==("failed" if negative else "passed") and o["primaryFailureCount"]==o["failureCount"]==(1 if negative else 0) and o["cleanupFailureCount"]==0,"Unexpected failure or inconclusive outcome")
require(o["failedAssertion"]==("guarded-batch:one-retained-revision-failure" if negative else None) and o["failedPhase"]==("oracle" if negative else None),"Wrong failure phase")
require(o["namedAssertion"]==o["failedAssertion"],"Named assertion projection differs")
for k in ("deadlineExceeded","cleanupExceeded","drainExceeded","milestoneOverflow","reportWriteFailed","reportOverflow","claimEligible","inconclusivePending"):require(o[k] is False,"Unacceptable state: "+k)
require(o["bounds"]=={"workMs":45000,"requestMs":5000,"cleanupMs":10000,"requestDrainMs":5000,"testMs":60000,"maximumRequests":count,"responseBodyBytes":32768,"responseHeaderBytes":16384,"reportBytes":65536},"Runtime bounds changed")
hex64=lambda v:isinstance(v,str) and re.fullmatch("[0-9a-f]{64}",v) is not None
def token(v):return isinstance(v,str) and 0<len(v.encode())<=8192
def owned(v,kind):
 require(isinstance(v,str) and len(v)<=8192 and len(v.encode())<=8192 and re.fullmatch(re.escape(scope+kind+"/")+r"[A-Za-z0-9_-]+",v) is not None,"Invalid issued URL")
 return v
def pos(v):
 require(isinstance(v,str) and re.fullmatch(r"pos-(0|[1-9][0-9]*)",v) is not None and int(v[4:])<=9007199254740991,"Invalid position")
 return int(v[4:])

p=document(g,"installedProvenance");entry=Path(p["entryRealpath"]).resolve(strict=True)
require(p["sourceCommit"]==g["finalHead"] and p["sourceTree"]==g["finalTree"],"Installed attribution")
require(Path(o["entry"]).is_absolute() and Path(o["entry"]).resolve(strict=True)==entry and entry.is_relative_to((E/"consumer").resolve()),"Wrong external entry")
require(o["declaredSourceCommit"]==g["finalHead"] and o["entrySha256"]==p["entrySha256"]==sha(entry.read_bytes()),"Entry/source bytes differ")
require(start["bdpOverrides"]=={"BDP_TRANSACTIONAL_INSTALLED_ENTRY":str(entry),"BDP_TRANSACTIONAL_SOURCE_COMMIT":g["finalHead"],"BDP_TRANSACTIONAL_GUARDED_BATCH_MODE":cohort,"BDP_TRANSACTIONAL_GUARDED_BATCH_OUTPUT":str(out)},"Unexpected BDP overrides")
b=o["binding"];require(isinstance(b,dict) and set(b)=={"served","operationTargets","runtimeConfigurationIdentity","authorityMode"},"Unexpected binding fields")
s=b["served"];require(isinstance(s,dict) and set(s)=={"scope","profile","history","sourceCommit","installedPayloadDigest","installedContractsDigest","policyDigest","authenticationDigest","fixtureDigest","fixtureId","capabilities"},"Unexpected served fields")
require(b["authorityMode"]=="canonical" and hex64(b["runtimeConfigurationIdentity"]),"Runtime identity")
require(s["scope"]==scope and s["profile"]=="transactional" and s["history"] is False and s["sourceCommit"]==g["finalHead"] and s["installedPayloadDigest"]==p["payloadIdentity"],"Served identity differs")
for k in ("installedPayloadDigest","installedContractsDigest","policyDigest","authenticationDigest","fixtureDigest"):require(hex64(s[k]),"Invalid served digest")
fp=W/"packages/conformance/test-support/transactional/fixture-v1.json";fixture=json.loads(fp.read_text())
require(s["fixtureDigest"]==sha(fp.read_bytes()) and s["fixtureId"]==fixture["id"] and s["capabilities"]==sorted(fixture["capabilities"]),"Fixture differs")
targets={"createBead":"create-bead","updateBead":"update-bead","deleteBead":"delete-bead","createLink":"create-link","updateLink":"update-link","deleteLink":"delete-link","putAlias":"put-alias","deleteAlias":"delete-alias","updateWhere":"update-where","deleteWhere":"delete-where","batch":"batch","sequence":"sequence"}
require(b["operationTargets"]=={k:scope+"operations/"+v for k,v in targets.items()},"Target inventory differs")
ids=o["servedIdentities"];require(isinstance(ids,dict) and set(ids)=={"runtimeConfigurationIdentity","artifactIdentity","policyIdentity","authenticationIdentity","payloadIdentity"},"Unexpected identity projection")
require(ids=={"runtimeConfigurationIdentity":b["runtimeConfigurationIdentity"],"artifactIdentity":"development-schema5-"+s["installedContractsDigest"],"policyIdentity":s["policyDigest"],"authenticationIdentity":s["authenticationDigest"],"payloadIdentity":p["payloadIdentity"]},"Identity projections disagree")
require(o["artifacts"]==ledger["runtimeConfiguration"]["artifacts"],"Artifact bytes differ")
projection={"scope":scope,"directory":o["directory"],"create":True,"artifacts":o["artifacts"],"credentials":[{"token":{"length":32,"charset":"base64url"},"principal":"writer","canMutate":True}]}
require(o["credentialFreeConfigurationProjectionSha256"]==sha(json.dumps(projection,separators=(",",":"),ensure_ascii=False).encode()),"Safe configuration projection differs")


responses=o["responses"]
require(type(o["attempts"]) is int and type(o["completedResponses"]) is int and o["attempts"]==o["completedResponses"]==len(responses)==count,"Wrong actual request count")
require(all(isinstance(x,dict) and set(x)==set(contract["responsesFields"]) for x in responses),"Unexpected response fields")
require(sorted(x["index"] for x in responses)==list(range(1,count+1)),"Response indices")
by={x["index"]:x for x in responses};receipts=o["receiptObservations"]
require([x.get("index") for x in receipts]==([1,2,3] if negative else [1,2,3,4,5]) and all(set(x)==set(contract["receiptFacts"]) for x in receipts),"Receipt record shape/order")
rr={x["index"]:x for x in receipts}
for x in receipts:
 i=x["index"];owned(x["id"],"receipts");require(token(x["transaction"]),"Invalid transaction identity")
 require(x["key"]==({1:"guarded-seed",2:"guarded-a",3:"guarded-b"}.get(i,o["selectedLoser"])),"Wrong receipt key")
 require(x["status"] in ("completed","failed") and x["requiredPosition"]==("pos-1" if i==1 else (x["effectPosition"] if negative else "pos-2")),"Wrong terminal receipt state")
 require(isinstance(x["resources"],list) and all(isinstance(v,dict) and set(v)=={"id","revision"} and token(v["revision"]) for v in x["resources"]),"Unexpected resource projection")
 if x["status"]=="failed":
  require(not negative and i!=1 and x["effectPosition"] is None and x["resources"]==[] and x["problem"]==ledger["failedProblem"],"Unexpected failed receipt")
 else:
  require(x["problem"] is None and len(x["resources"])==(1 if i==1 else 2),"Unexpected completed receipt")
  expected_ids=[scope+"beads/guarded-seed"] if i==1 else [scope+"beads/marker-"+("a" if i==2 else "b"),scope+"beads/guarded-seed"]
  require([v["id"] for v in x["resources"]]==expected_ids,"Wrong complete result identities")
  require(pos(x["effectPosition"])==(1 if i==1 else pos(x["requiredPosition"])),"Effect/required mismatch")
  if i>1:require(x["resources"][1]["revision"]!=rr[1]["resources"][0]["revision"],"Guarded update did not change seed revision")
 require(x["servingPosition"]==by[i]["headers"]["bdp-scope-position"],"Receipt serving projection mismatch")
require(rr[1]["status"]=="completed" and rr[1]["effectPosition"]=="pos-1","Seed prerequisite")
require(rr[2]["id"]!=rr[3]["id"] and rr[2]["transaction"]!=rr[3]["transaction"],"Pair identities not distinct")
completed_pair=[x for i,x in rr.items() if i in (2,3) and x["status"]=="completed"]
failed_pair=[x for i,x in rr.items() if i in (2,3) and x["status"]=="failed"]
require(o["completedContenderCount"]==len(completed_pair)==(2 if negative else 1) and o["failedContenderCount"]==len(failed_pair)==(0 if negative else 1),"Wrong independently counted contender outcomes")
require(set(o["prerequisites"])==set(contract["prerequisites"]) and all(o["prerequisites"][k] is True for k in ("seedSchemaAndLiterals","pairSchemasAndLiterals","dispatchDiscipline")) and o["prerequisites"]["controlTwoCompletedEffects"] is negative,"Prerequisite flags differ")
revision=rr[1]["resources"][0]["revision"]
compact=lambda v:json.dumps(v,separators=(",",":"),ensure_ascii=False).encode()
encoded=len(compact(revision))
require(type(o["revisionEncodingBytes"]) is int and o["revisionEncodingBytes"]==encoded and 0<encoded<=128,"Captured revision encoding")
def body(side):
 value=json.loads(ledger["inputs"][side+"ReferenceGuard"]["body"])
 if negative:del value["operations"][1]["expectedRevision"]
 else:value["operations"][1]["expectedRevision"]=revision
 return compact(value)
a_body,b_body=body("a"),body("b")
require(len(a_body)==len(b_body)==(323 if negative else 343+encoded),"Dynamic contender bound")
if negative:
 require(o["selectedWinner"] is None and o["selectedLoser"] is None,"Control advanced beyond intended oracle")
 require(sorted(x["effectPosition"] for x in completed_pair)==["pos-2","pos-3"] and rr[2]["resources"][1]["revision"]!=rr[3]["resources"][1]["revision"],"Control independent effects absent")
else:
 loser=failed_pair[0];winner=completed_pair[0]
 require(o["selectedLoser"]==loser["key"] and o["selectedWinner"]==winner["key"] and winner["effectPosition"]=="pos-2","Wrong winner/loser selection")
 for i in (4,5):
  require({k:v for k,v in rr[i].items() if k!="index"}=={k:v for k,v in loser.items() if k!="index"},"Retained receipt facts differ")
  require(by[i]["bodySha256"]==by[loser["index"]]["bodySha256"] and by[i]["bodyBytes"]==by[loser["index"]]["bodyBytes"],"Retained complete receipt fingerprints differ")
bodies={1:ledger["inputs"]["seed"]["body"].encode(),2:a_body,3:b_body}
if not negative:bodies[4]=bodies[loser["index"]]
for i,row in enumerate(request_ledger["requests"][:count],1):
 x=by[i];expected_url=rr[5]["id"] if i==5 else row["url"]
 require(x["method"]==row["method"] and x["url"]==expected_url,"Request route differs")
 bdy=bodies.get(i);key={1:"guarded-seed",2:"guarded-a",3:"guarded-b",4:o["selectedLoser"]}.get(i)
 require(x["sentKey"]==key and type(x["requestBodyBytes"]) is int and x["requestBodyBytes"]==(0 if bdy is None else len(bdy)) and x["requestBodySha256"]==(None if bdy is None else sha(bdy)),"Actual request fingerprint differs")
 require(type(x["bodyBytes"]) is int and 0<x["bodyBytes"]<=32768 and hex64(x["bodySha256"]) and type(x["status"]) is int and x["status"]==200,"Response status/fingerprint invalid")
 h=x["headers"];require(isinstance(h,dict) and set(h)==set(contract["safeHeaderNames"]) and all(v is None or (isinstance(v,str) and len(v.encode())<=16384) for v in h.values()),"Unexpected response header projection")
 require(h["content-length"]==str(x["bodyBytes"]) and h["content-type"]=="application/json" and h["cache-control"]=="private, no-store" and h["x-bdp-development"]=="non-attesting","Wrong response representation")
 require(token(h["bdp-scope-epoch"]) and token(h["bdp-authorization-view"]) and h["bdp-scope-epoch"]==by[1]["headers"]["bdp-scope-epoch"] and h["bdp-authorization-view"]==by[1]["headers"]["bdp-authorization-view"],"Changed context")
 serving=pos(h["bdp-scope-position"])
 require(serving==1 if i==1 else (pos(rr[i]["effectPosition"])<=serving<=3 if negative else serving==2),"Wrong current context")
 if i<=4:require(isinstance(h["vary"],str) and h["vary"].lower()=="accept","Mutation Vary")
require(sum(x["requestBodyBytes"] for x in responses)==(818 if negative else 172+3*(343+encoded)),"Body sum differs")
expected_performed=["literal-inputs","ready-targets","receipt-1-schema-and-literals","captured-revision-and-bodies","dispatch-discipline","receipt-2-schema-and-literals","receipt-3-schema-and-literals","pair-identities"]
if negative:expected_performed+=["control-prerequisites"]
else:expected_performed+=["one-retained-revision-failure","frozen-loser-replay-input","receipt-4-schema-and-literals","retained-failure-full-equality","receipt-5-schema-and-literals","issued-failure-full-equality","complete-rollback-resource-collection","complete-rollback-event-linkage","complete-fixed-schedule"]
require(o["performed"]==expected_performed,"Wrong exact oracle progression")

ms=o["milestones"];require(isinstance(ms,list) and 1<=len(ms)<=32,"Milestone bound")
allowed={"invoked","end-called","socket-assigned","connected","finish","response-start","response-end","settled","request-close","socket-close","response-close"}
seen=set()
for order,m in enumerate(ms):
 require(isinstance(m,dict) and set(m)==set(binding["milestoneFields"]) and m["order"]==order and m["request"] in (2,3) and m["kind"] in allowed and number(m["elapsedMs"]) and 0<=m["elapsedMs"]<=o["elapsedMs"],"Invalid milestone")
 require((m["request"],m["kind"]) not in seen,"Duplicate milestone");seen.add((m["request"],m["kind"]))
 if order:require(m["elapsedMs"]>=ms[order-1]["elapsedMs"],"Nonmonotonic milestone time")
 require(m["socket"] is None or type(m["socket"]) is int,"Invalid milestone socket")
sockets=o["requestSockets"]
require(isinstance(sockets,list) and len(sockets)==count and all(set(x)=={"index","socket"} and type(x["index"]) is int and type(x["socket"]) is int for x in sockets),"Invalid request socket records")
require(sorted(x["index"] for x in sockets)==list(range(1,count+1)) and sorted(x["socket"] for x in sockets)==list(range(1,count+1)),"Socket reuse or missing request")
socket_by={x["index"]:x["socket"] for x in sockets}
for m in ms:
 require(m["socket"]==(None if m["kind"] in ("invoked","end-called") else socket_by[m["request"]]),"Milestone socket binding")
def order(i,kind):
 values=[m["order"] for m in ms if m["request"]==i and m["kind"]==kind]
 require(len(values)==1,"Required milestone missing: "+kind);return values[0]
first_settled=min(order(i,"settled") for i in (2,3));first_response=min(order(i,"response-start") for i in (2,3))
for i in (2,3):
 require(order(i,"invoked")<order(i,"end-called")<first_settled,"Dispatch discipline not observed")
 require(order(i,"socket-assigned")<order(i,"connected")<order(i,"response-start")<order(i,"response-end")<order(i,"settled"),"Response lifecycle ordering")
 require(order(i,"finish")<order(i,"response-end") and order(i,"response-end")<order(i,"request-close") and order(i,"connected")<order(i,"socket-close"),"Complete writes/cleanup not observed")
require(o["clientDispatchDisciplinePassed"] is True and o["bothConnectedBeforeFirstResponse"] is all(order(i,"connected")<first_response for i in (2,3)) and o["bothWritesPrecededFirstResponse"] is all(order(i,"finish")<first_response for i in (2,3)),"Milestone summary disagrees")
require(o["socketCount"]==count and type(o["maximumLiveSockets"]) is int and 1<=o["maximumLiveSockets"]<=2 and o["liveSocketsAfterCleanup"]==0,"Socket bounds")

require(o["launchAttempted"] is True and o["closed"] is True and o["stopKnowledge"]=="known-stopped-or-never-launched" and o["childExitCode"]==0 and o["childSignalCode"] is None,"Child not known stopped cleanly")
pid=o["childPid"];require(type(pid) is int and pid>0,"Missing child PID")
require(terminal["remainingAfterCleanup"]==[] and terminal["sampledPidsStillPresent"]=={} and terminal["runtimeDirectoriesStillPresent"]==[],"Supervisor cleanup incomplete")
sampled=terminal["sampledPids"];require(isinstance(sampled,list) and sampled and all(type(n) is int and n>0 for n in sampled) and terminal["processGroup"] in sampled,"Missing owned process samples")
absent=[]
for n in sorted(set(sampled+[pid])):
 try:os.kill(n,0)
 except ProcessLookupError:absent.append(n)
 else:raise RuntimeError("Owned child/sampled PID remains present")
directory=Path(o["directory"]);tmp=(E/(name+".tmp")).resolve()
require(Path(o["temporaryRoot"]).resolve()==tmp and directory.is_absolute() and directory.resolve().parent==tmp and directory.name.startswith("bdp-guarded-batch-installed-") and o["directoryAbsent"] is True and not directory.exists(),"Store remains or wrong owned root")
record={"candidateMatches":True,"command":name,"actualTestExit":terminal["exit"],"actualSupervisorExit":terminal["supervisorExit"],"adjudication":"intended named failure" if negative else "positive passed","actualHttp":count,"actualChildren":1,"reportedClosedChildren":1,"sourceCommit":g["finalHead"],"sourceTree":g["finalTree"],"payloadIdentity":p["payloadIdentity"],"entryRealpath":str(entry),"entrySha256":p["entrySha256"],"runtimeConfigurationIdentity":b["runtimeConfigurationIdentity"],"childPid":pid,"childPidWasSampled":pid in sampled,"independentlyAbsentPids":absent,"storeDirectoryAbsent":str(directory),"controlPrerequisitesIndependentlyChecked":negative,"clientDispatchDisciplineChecked":True,"bothConnectedBeforeFirstResponse":o["bothConnectedBeforeFirstResponse"],"bothWritesPrecededFirstResponse":o["bothWritesPrecededFirstResponse"],"inconclusivePending":o["inconclusivePending"],"rootReviewRequired":True,"claimEligible":False,"observationSha256":sha(out.read_bytes()),"resultsSha256":sha(results.read_bytes()),"rawLogSha256":sha((E/(name+".log")).read_bytes()),"terminalSha256":sha((E/(name+".exit.json")).read_bytes()),"limitations":"Full schema/body/Resource/Event equality are source-bound in-memory assertions; retained fingerprints are not full-wire evidence. No physical/server admission overlap claim. Caller source attribution is externally verified, not child-attested. Process samples can miss short-lived descendants; returned PID independently checked. No qualification."}
with (E/("review-"+cohort+".review-candidate.json")).open("x") as f:f.write(json.dumps(record,indent=2)+"\n")
print(json.dumps(record))
