"""Read retained outcome only; no product import. Negative must fail its named conjunction."""
from pathlib import Path
import json,os,re,sys
from gate import E,P,W,sha,require,document,source_hashes

def review(name,preliminary=False):
 g=json.loads((P/"execution-gate.json").read_text());spec=json.loads((P/"commands.json").read_text())["commands"][name]
 report=json.loads((E/(name+".json")).read_text());items=[x for f in report["testResults"] for x in f["assertionResults"] if x["status"] not in ("pending","skipped","todo")]
 require(len(items)==1 and [x["fullName"] for x in items]==spec["selectedFullNames"],"Exact one selected case required")
 require(not report.get("unhandledErrors") and not report.get("errors") and not report.get("numRuntimeErrorTestSuites",0),"Unhandled runtime failure")
 raw=E/(name+"-observation.json");o=json.loads(raw.read_text());require(o["claimEligible"] is False and o["sourceCommit"]==g["finalHead"] and o["workspaceDirty"] is False,"Source/claim differs")
 require(o["mode"]==spec["mode"] and o["attempted"]==len(o["completedResponses"])==spec["workload"]["httpExpected"],"Count/mode differs")
 require(o["overlapObserved"] is True and o["pairSchemasValid"] and not o["eventOverflow"] and not o["droppedEvents"] and not o["workDeadline"],"Incomplete witness")
 require(not o["cleanupFailures"] and o["directoryAbsent"] and not Path(o["directory"]).exists() and o["childExitCode"]==0 and o["childSignalCode"] is None,"Unclean child/store")
 try:os.kill(o["childPid"],0)
 except ProcessLookupError:pass
 else:raise RuntimeError("Product child still present")
 trace=o["milestones"];require(len(trace)<=32 and [x["ordinal"] for x in trace]==list(range(1,len(trace)+1)),"Milestone sequence differs")
 slots={(x["request"],x["kind"]):x for x in trace};require(len(slots)==len(trace),"Duplicate milestone slot")
 for pair in ("A","B"):
  for kind in ("invoked","end-called","socket-assigned","settled"):require((pair,kind) in slots,"Required milestone absent")
 first=min(slots[(pair,"settled")]["ordinal"] for pair in ("A","B"))
 require(all(slots[(pair,kind)]["ordinal"]<first for pair in ("A","B") for kind in ("invoked","end-called")),"No outstanding-call overlap")
 require(slots[("A","socket-assigned")]["socket"]!=slots[("B","socket-assigned")]["socket"],"Socket identities shared")
 for key in ("configurationIdentity","payloadIdentity","policyIdentity","authenticationIdentity"):
  require(re.fullmatch("[a-f0-9]{64}",o["ready"][key]) is not None,"Ready digest invalid")
 require(re.fullmatch("development-schema5-[a-f0-9]{64}",o["ready"]["artifactIdentity"]) is not None,"Artifact identity invalid")
 rows=o["completedResponses"];require([x["number"] for x in rows]==list(range(1,len(rows)+1)),"Missing/reordered request association")
 require([x["method"] for x in rows]==(["GET","GET","GET","POST","POST","POST"] if name=="negative" else ["GET","GET","GET","POST","POST","POST","GET","POST","POST","GET"]),"Method ledger differs")
 require([x["status"] for x in rows[:4]]==[204,200,200,200],"Bootstrap/seed differs")
 require(rows[4]["url"]==rows[5]["url"],"Pair targets differ")
 if name=="positive":require(rows[7]["url"]==rows[8]["url"]==rows[4]["url"] and rows[6]["url"]==rows[9]["url"],"Replay/read routes differ")
 require(sum(x["requestBodyOctets"] for x in rows)==o["bodyBytes"],"Recorded body total differs")
 pair=rows[4:6];require(all(x["bodyOctets"]<=32768 and re.fullmatch("[a-f0-9]{64}",x["bodySha256"]) for x in rows),"Fingerprint/bound differs")
 if name=="negative":
  require(items[0]["status"]=="failed" and report["numFailedTests"]==1 and report["numPassedTests"]==0 and report["success"] is False,"Negative did not fail")
  require(o["outcome"]=="failed" and o["failureCount"]==1 and o["terminationClass"]=="oracle" and o["namedOracle"]=="overlap:exactly-one-winner","Wrong negative failure")
  require([x["status"] for x in pair]==[200,200] and o["winner"] is None,"Not two schema-valid updates")
  binding=document(g,"derivativeBinding");payload=document(g,"derivativePayload")["payloadIdentity"]
  require(o["bodyBytes"]==364+2*o["revisionJsonBytes"],"Negative byte ledger differs")
 else:
  require(items[0]["status"]=="passed" and report["success"] and report["numPassedTests"]==1 and report["numFailedTests"]==0,"Positive failed")
  require(o["outcome"]=="passed" and o["failureCount"]==0 and o["namedOracle"] is None,"Positive outcome differs")
  require(sorted(x["status"] for x in pair)==[200,409] and rows[6]["status"]==rows[9]["status"]==200,"Wrong winner/read statuses")
  require(rows[7]["status"]==rows[4]["status"] and rows[8]["status"]==rows[5]["status"],"Wrong replay statuses")
  require(rows[7]["bodySha256"]==rows[4]["bodySha256"] and rows[8]["bodySha256"]==rows[5]["bodySha256"] and rows[6]["bodySha256"]==rows[9]["bodySha256"],"Retained fingerprint equality failed")
  binding=document(g,"installedProvenance");payload=binding["payloadIdentity"]
  require(o["bodyBytes"]==634+4*o["revisionJsonBytes"],"Positive byte ledger differs")
 entry=Path(binding["entry"] if name=="negative" else binding["entryRealpath"]).resolve(strict=True)
 require(Path(o["entry"]).resolve()==entry and o["entrySha256"]==sha(entry.read_bytes())==binding["entrySha256"],"Entry differs")
 require(o["ready"]["payloadIdentity"]==payload and o["ready"]["profile"]=="read-update" and o["ready"]["history"] is False and o["ready"]["claimEligible"] is False,"Ready binding differs")
 result={"accepted":True,"name":name,"http":len(rows),"productChildren":1,"childPidAbsent":o["childPid"],"storeAbsent":o["directory"],"sourceCommit":g["finalHead"],"sourceHashes":source_hashes(),"observationSha256":sha(raw.read_bytes()),"resultSha256":sha((E/(name+".json")).read_bytes()),"payloadIdentity":payload,"claimEligible":False,"limit":"Full schemas/body/replay equality ran in memory; hashes/headers and named failed test record are retained. Not independent non-reexecution proof for loser or source attestation."}
 if not preliminary:
  with (E/(name+"-review.json")).open("x") as f:f.write(json.dumps(result,indent=2)+"\n")
 return result
if __name__=="__main__":
 from gate import load_gate
 load_gate(sys.argv[1]+"-review");print(json.dumps(review(sys.argv[1])))
