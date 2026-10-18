"""Adjudicate actual same-key cohort records; no product imports or execution.
Negative exit1 is retained; adapter exit0 means only intended failure with prerequisites."""
from pathlib import Path
import json,os,re,sys,math
from gate import P,D,W,E,load_gate,completed,document,require,sha
require(len(sys.argv)==2 and sys.argv[1] in ("positive","second-key-distinct"),"Exactly one cohort required")
cohort=sys.argv[1];negative=cohort=="second-key-distinct";name="installed-"+cohort;count=2 if negative else 6
g,_=load_gate("review-"+cohort);terminal=completed(name)
require(Path(g["reportContract"]["path"]).resolve()==(P/"report-adapter-binding.json").resolve(),"Wrong adapter binding")
binding=document(g,"reportContract")
for key in ("reportContract","ledger","selection","requests","sourceBinding"):
 pin=binding[key];require(sha(Path(pin["path"]).read_bytes())==pin["sha256"],"Adapter input changed: "+key)
contract=json.loads(Path(binding["reportContract"]["path"]).read_text())
ledger=json.loads(Path(binding["ledger"]["path"]).read_text())
selection=json.loads(Path(binding["selection"]["path"]).read_text())
request_ledger=json.loads(Path(binding["requests"]["path"]).read_text())
test=selection["path"];scope=ledger["wireConfiguration"]["scope"]
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
 require(first.startswith("AssertionError: same-key-overlap:shared-receipt-id: expected ") and first.endswith(" // Object.is equality") and " to be " in first,"Wrong named negative failure")

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
require(set(o)==set(logged)==set(contract["topLevelFields"]),"Unexpected report fields")
require({k:v for k,v in o.items() if k!="elapsedMs"}=={k:v for k,v in logged.items() if k!="elapsedMs"},"File/stdout disagree")
def number(v):return type(v) in (int,float) and math.isfinite(v)
require(number(o["elapsedMs"]) and number(logged["elapsedMs"]) and 0<=o["elapsedMs"]<=logged["elapsedMs"]<=60000,"Elapsed bound")
require(o["label"]==contract["label"] and o["mode"]==cohort and o["executionMode"]=="external-installed","Wrong report identity")
require(o["sourceAttribution"]=="caller label; supervisor must verify source/package equality" and o["workspaceDirty"] is False,"Wrong source attribution")
require(o["retention"]=="Full parsed receipt/Resource/Event comparisons occur in memory; response fingerprints are not retained full-wire evidence.","Unexpected retention declaration")
require(o["outcome"]==("failed" if negative else "passed") and o["primaryFailureCount"]==o["failureCount"]==(1 if negative else 0) and o["cleanupFailureCount"]==0,"Unexpected failure or inconclusive outcome")
require(o["failedAssertion"]==("same-key-overlap:shared-receipt-id" if negative else None) and o["failedPhase"]==("oracle" if negative else None),"Wrong failure phase")
for k in ("deadlineExceeded","cleanupExceeded","drainExceeded","milestoneOverflow","reportWriteFailed","reportOverflow","claimEligible","inconclusiveBothPending"):require(o[k] is False,"Unacceptable state: "+k)
require(o["bounds"]=={"workMs":45000,"requestMs":5000,"cleanupMs":10000,"requestDrainMs":5000,"testMs":60000,"maximumRequests":6,"responseBodyBytes":32768,"responseHeaderBytes":16384,"reportBytes":65536},"Runtime bounds changed")
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
require(start["bdpOverrides"]=={"BDP_TRANSACTIONAL_INSTALLED_ENTRY":str(entry),"BDP_TRANSACTIONAL_SOURCE_COMMIT":g["finalHead"],"BDP_TRANSACTIONAL_SAME_KEY_OVERLAP_MODE":cohort,"BDP_TRANSACTIONAL_SAME_KEY_OVERLAP_OUTPUT":str(out)},"Unexpected BDP overrides")
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
require(o["artifacts"]==ledger["wireConfiguration"]["artifacts"],"Artifact bytes differ")
projection={"scope":scope,"directory":o["directory"],"create":True,"artifacts":o["artifacts"],"credentials":[{"token":{"length":32,"charset":"base64url"},"principal":"writer","canMutate":True}]}
require(o["credentialFreeConfigurationProjectionSha256"]==sha(json.dumps(projection,separators=(",",":"),ensure_ascii=False).encode()),"Safe configuration projection differs")

responses=o["responses"]
require(type(o["attempts"]) is int and type(o["completedResponses"]) is int and o["attempts"]==o["completedResponses"]==len(responses)==count,"Wrong actual request count")
require(all(isinstance(x,dict) and set(x)==set(contract["responseFields"]) for x in responses),"Unexpected response fields")
require(sorted(x["index"] for x in responses)==list(range(1,count+1)),"Response indices")
by={x["index"]:x for x in responses};receipts=o["receiptObservations"];expected_receipts=[1,2] if negative else [1,2,3,6]
require([x.get("index") for x in receipts]==expected_receipts and all(set(x)==set(contract["receiptFields"]) for x in receipts),"Receipt record shape/order")
rr={x["index"]:x for x in receipts}
for x in receipts:
 i=x["index"];owned(x["id"],"receipts");require(token(x["transaction"]) and type(x["pending"]) is bool,"Invalid receipt identity")
 expected_key="same-key-overlap-other" if negative and i==2 else "same-key-overlap"
 require(x["key"]==expected_key,"Wrong receipt key")
 if x["pending"]:
  require(not negative and i<=2 and all(x[k] is None for k in ("resourceId","revision","effectPosition","requiredPosition")),"Unexpected pending shape")
 else:
  owned(x["resourceId"],"beads");require(token(x["revision"]),"Missing revision")
  require(x["effectPosition"]==x["requiredPosition"] and pos(x["effectPosition"]) in ([1,2] if negative else [1]),"Wrong effect/required position")
for i,row in enumerate(request_ledger["requests"][:count],1):
 x=by[i];expected_url=rr[1]["id"] if i==3 else row["url"]
 require(x["method"]==row["method"] and x["url"]==expected_url and x["requestBodyBytes"]==row["bodyBytes"] and x["requestBodySha256"]==row["bodySha256"],"Request fingerprint differs")
 sent="same-key-overlap-other" if negative and i==2 else ("same-key-overlap" if row["method"]=="POST" else None)
 require(x["sentKey"]==sent,"Wrong sent-key projection")
 require(type(x["bodyBytes"]) is int and 0<x["bodyBytes"]<=32768 and hex64(x["bodySha256"]),"Response fingerprint missing")
 expected_status=202 if i in rr and rr[i]["pending"] else 200
 require(type(x["status"]) is int and x["status"]==expected_status,"Wrong actual status")
 h=x["headers"];require(isinstance(h,dict) and set(h)==set(contract["safeResponseHeaders"]) and all(v is None or (isinstance(v,str) and len(v.encode())<=16384) for v in h.values()),"Unexpected response header projection")
 require(h["content-length"]==str(x["bodyBytes"]) and h["content-type"]=="application/json" and h["cache-control"]=="private, no-store" and h["x-bdp-development"]=="non-attesting","Wrong response representation")
 require(token(h["bdp-scope-epoch"]) and token(h["bdp-authorization-view"]) and h["bdp-scope-epoch"]==by[1]["headers"]["bdp-scope-epoch"] and h["bdp-authorization-view"]==by[1]["headers"]["bdp-authorization-view"],"Changed context")
 serving=pos(h["bdp-scope-position"])
 if expected_status==202:
  require(serving in (0,1) and h["location"]==rr[i]["id"] and h["retry-after"]=="1","Pending context/location/retry")
 else:
  require(serving in ([1,2] if negative else [1]),"Wrong current position")
  if i in rr:require(serving>=pos(rr[i]["effectPosition"]),"Current position behind effect")
 if i in (1,2,6):require(isinstance(h["vary"],str) and h["vary"].lower()=="accept","Mutation Vary")
require(o["pendingObserved"] is any(x["pending"] for x in receipts),"Pending observation mismatch")
require(o["receiptIdsDiffer"] is (rr[1]["id"]!=rr[2]["id"]),"Receipt distinction mismatch")
if negative:
 require(o["controlPrerequisites"] is True and o["receiptIdsDiffer"] is True,"Control prerequisite flags missing")
 require(sorted(x["effectPosition"] for x in receipts)==["pos-1","pos-2"] and len({x["transaction"] for x in receipts})==2 and len({x["resourceId"] for x in receipts})==2,"Independent control effects not established")
 expected_performed=["literal-inputs","ready-targets","dispatch-discipline","receipt-1-schema-and-literals","receipt-2-schema-and-literals","control-prerequisites"]
else:
 require(o["controlPrerequisites"] is False and o["receiptIdsDiffer"] is False and not all(rr[i]["pending"] for i in (1,2)),"Positive pair invalid")
 for k in ("id","transaction","key"):require(len({x[k] for x in receipts})==1,"Positive receipt identity differs")
 terminal_receipts=[x for x in receipts if not x["pending"]]
 for k in ("resourceId","revision","effectPosition","requiredPosition"):require(len({x[k] for x in terminal_receipts})==1,"Positive effect differs")
 require(len({by[x["index"]]["bodySha256"] for x in terminal_receipts})==1 and len({by[x["index"]]["bodyBytes"] for x in terminal_receipts})==1,"Full terminal receipt fingerprint differs")
 expected_performed=list(contract["positive"]["requiredPerformed"])
 if not any(rr[i]["pending"] for i in (1,2)):expected_performed.insert(expected_performed.index("receipt-3-schema-and-literals"),"pair-full-terminal-equality")
require(o["performed"]==expected_performed,"Wrong exact oracle progression")

ms=o["milestones"];require(isinstance(ms,list) and 1<=len(ms)<=32,"Milestone bound")
allowed={"invoked","end-called","socket-assigned","connected","finish","response-start","response-end","settled","request-close","socket-close","response-close"}
seen=set()
for order,m in enumerate(ms):
 require(isinstance(m,dict) and set(m)==set(contract["milestoneFields"]) and m["order"]==order and m["request"] in (1,2) and m["kind"] in allowed and number(m["elapsedMs"]) and 0<=m["elapsedMs"]<=o["elapsedMs"],"Invalid milestone")
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
first_settled=min(order(i,"settled") for i in (1,2));first_response=min(order(i,"response-start") for i in (1,2))
for i in (1,2):
 require(order(i,"invoked")<order(i,"end-called")<first_settled,"Dispatch discipline not observed")
 require(order(i,"socket-assigned")<order(i,"connected")<order(i,"response-start")<order(i,"response-end")<order(i,"settled"),"Response lifecycle ordering")
 require(order(i,"finish")<order(i,"response-end") and order(i,"response-end")<order(i,"request-close") and order(i,"connected")<order(i,"socket-close"),"Complete writes/cleanup not observed")
require(o["clientDispatchDisciplinePassed"] is True and o["bothConnectedBeforeFirstResponse"] is all(order(i,"connected")<first_response for i in (1,2)) and o["bothWritesPrecededFirstResponse"] is all(order(i,"finish")<first_response for i in (1,2)),"Milestone summary disagrees")
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
require(Path(o["temporaryRoot"]).resolve()==tmp and directory.is_absolute() and directory.resolve().parent==tmp and directory.name.startswith("bdp-same-key-overlap-installed-") and o["directoryAbsent"] is True and not directory.exists(),"Store remains or wrong owned root")
record={"candidateMatches":True,"command":name,"actualTestExit":terminal["exit"],"actualSupervisorExit":terminal["supervisorExit"],"adjudication":"intended named failure" if negative else "positive passed","actualHttp":count,"actualChildren":1,"reportedClosedChildren":1,"sourceCommit":g["finalHead"],"sourceTree":g["finalTree"],"payloadIdentity":p["payloadIdentity"],"entryRealpath":str(entry),"entrySha256":p["entrySha256"],"runtimeConfigurationIdentity":b["runtimeConfigurationIdentity"],"childPid":pid,"childPidWasSampled":pid in sampled,"independentlyAbsentPids":absent,"storeDirectoryAbsent":str(directory),"controlPrerequisitesIndependentlyChecked":negative,"clientDispatchDisciplineChecked":True,"bothConnectedBeforeFirstResponse":o["bothConnectedBeforeFirstResponse"],"bothWritesPrecededFirstResponse":o["bothWritesPrecededFirstResponse"],"pendingObserved":o["pendingObserved"],"rootReviewRequired":True,"claimEligible":False,"observationSha256":sha(out.read_bytes()),"resultsSha256":sha(results.read_bytes()),"rawLogSha256":sha((E/(name+".log")).read_bytes()),"terminalSha256":sha((E/(name+".exit.json")).read_bytes()),"limitations":"Full schema/body/Resource/Event equality are source-bound in-memory assertions; retained fingerprints are not full-wire evidence. No physical/server pending-join claim. Caller source attribution is externally verified, not child-attested. Process samples can miss short-lived descendants; returned PID independently checked. No qualification."}
with (E/("review-"+cohort+".review-candidate.json")).open("x") as f:f.write(json.dumps(record,indent=2)+"\n")
print(json.dumps(record))
