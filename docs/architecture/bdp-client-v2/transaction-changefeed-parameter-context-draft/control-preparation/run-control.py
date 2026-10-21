"""One closed emitted-control invocation; reuses the reviewed source supervisor.
No pnpm/shim dispatch, setup, build, installed child, retry or positive-source edit.
"""
import hashlib, importlib.util, json, os, re, sys
from pathlib import Path
P=Path(__file__).resolve().parent
D=P.parent
C=P/"mirror"
W=Path("/Users/dbox/repos/wt/vickie/vickie-changefeed-parameter-context")
BASE=D/"execution-preparation/run-source.py"
BASE_SHA="72a0bb22fb72c059c17d833a8c4194d36560240e13402114d442a3ae8a4102a6"
COMMAND_SHA="4b0d0c33ef76fd8bf3251fb4b1ec82ff77d1ee0661e1c9d18e39a7e38b673200"
PACKET_SHA="538faf62f0dd43793f9f43d88015a019b1d768009d62eb021039bd59f3e3b260"
NAME="omit-context-control"
SESSION="2026-10-08T20:26:35Z-81e5"
def require(ok,message):
 if not ok: raise RuntimeError(message)
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def read(p):return json.loads(p.read_text())
def verify_graph():
 require(sha(P/"preparation-manifests.json")==PACKET_SHA,"Packet changed")
 for name,h in read(P/"preparation-manifests.json")["files"].items():require(sha(P/name)==h,"Manifest changed: "+name)
 manifest=read(P/"copy-manifest.json"); deps=read(P/"dependency-manifest.json")
 require(sha(D/"run/build.exit.json")==manifest["buildExitSha256"],"Build record changed")
 require(sha(D/"source-binding.json")==manifest["sourceBindingSha256"],"Source binding changed")
 for name,item in manifest["files"].items():
  require(sha(W/name)==item["positiveSha256"] and sha(C/name)==item["copySha256"],"Source/build member changed: "+name)
 for name,item in deps["files"].items():
  require(sha(W/name)==item["sha256"] and sha(C/name)==item["sha256"],"Dependency/tool bytes changed: "+name)
 for name,item in deps["links"].items():
  for root,key,real in [(W,"positiveLink","positiveRealpath"),(C,"copyLink","copyRealpath")]:
   p=root/name;require(p.is_symlink() and os.readlink(p)==item[key] and str(p.resolve(strict=True))==item[real],"Dependency link changed: "+name)
  require((C/name).resolve(strict=True).is_relative_to(C),"Positive dependency escape")
 expected=set(manifest["files"])|set(deps["files"])|set(deps["links"])
 actual=set()
 for root,dirs,names in os.walk(C,followlinks=False):
  dirs[:]=[n for n in dirs if n not in (".vite",".vite-temp")]
  for n in dirs+names:
   p=Path(root)/n
   if p.is_symlink() or p.is_file():actual.add(str(p.relative_to(C)))
 require(actual==expected,"Mirror file/link inventory changed")
 return {"packetSha256":PACKET_SHA,"sourceAndBuildMembers":len(manifest["files"]),"toolFiles":len(deps["files"]),"links":len(deps["links"])}
def adjudicate():
 result=read(P/"run"/(NAME+".exit.json"))
 require(result["exit"]==1 and result["supervisorExit"]==1,"Actual expected test failure absent")
 require(not result["timeout"] and not result["forcedCleanup"] and not result["errors"] and result["sourceUnchanged"],"Execution/source failure")
 for field in ("remainingAfterCleanup","sampledPidsStillPresent","runtimeDirectoriesStillPresent"):require(not result[field],"Owned cleanup failed: "+field)
 report=read(P/"run/control.results.json")
 cases=[c for f in report["testResults"] for c in f.get("assertionResults",[]) if c["status"] not in ("pending","skipped","todo")]
 command=read(P/"command.json")
 require(len(cases)==1 and cases[0]["status"]=="failed" and cases[0]["fullName"]==command["selectedFullNames"][0],"Selected failed case mismatch")
 require(report["success"] is False and report["numFailedTests"]==1 and report["numPassedTests"]==0,"Aggregate mismatch")
 require(not report.get("unhandledErrors") and not report.get("errors") and not report.get("numRuntimeErrorTestSuites",0),"Unhandled runtime failure")
 labels=["changefeed-context:"+str(n) for n in [3,4,5,6,7,8,12]]
 failures=cases[0].get("failureMessages",[]);require(len(failures)==7,"Expected exactly seven soft assertions")
 seen=[]
 for failure in failures:
  found=[x for x in labels if x in failure]
  require(len(found)==1 and "AssertionError" in failure,"Unexpected assertion failure")
  seen+=found
 require(sorted(seen)==sorted(labels),"Named failure inventory differs")
 observations=[]
 log=P/"run"/(NAME+".log")
 for line in re.sub(r"\x1b\[[0-9;]*m","",log.read_text()).splitlines():
  try:v=json.loads(line.strip())
  except ValueError:continue
  if isinstance(v,dict) and v.get("observation")=="private-source-changefeed-parameter-context":observations.append(v)
 require(len(observations)==1,"Actual observation missing/duplicated")
 o=observations[0]
 require(o["outcome"]=="failed" and o["attempted"]==12 and o["exchanges"]==12 and o["contextFailures"]==labels,"Observation count/oracles differ")
 require(o["hardFailureCount"]==0 and not o["deadlineExceeded"] and o["startupCompleted"] and o["closed"] and o["directoryAbsent"] and not o["claimEligible"],"Prerequisite/cleanup failure")
 require(o["directory"] and not Path(o["directory"]).exists(),"Owned directory remains")
 rows=o["completedResponses"]
 require(len(rows)==12 and [r["ordinal"] for r in rows]==list(range(1,13)),"Response inventory differs")
 require([r["status"] for r in rows]==[200,200,400,400,400,400,400,400,401,400,200,400],"Actual statuses differ")
 scope="https://example.test/feed/";future=rows[2]["url"]
 require(re.fullmatch(re.escape(scope+"changes/?after=")+r"cg_[A-Za-z0-9_-]{43}_[A-Za-z0-9_-]{43}_3",future) is not None,"Derived future input differs")
 initial=scope+"changes/?after=genesis";batch=scope+"operations/batch"
 routes=[("POST",batch),("GET",initial),("GET",future),("HEAD",future),("GET",future),("HEAD",future),("GET",initial),("HEAD",initial),("GET",future),("GET",scope+"changes/?after=genesis&start=now"),("POST",batch),("GET",future)]
 require([(r["method"],r["url"]) for r in rows]==routes,"Actual method/routes differ")
 require([r["accept"] for r in rows]==["application/json"]*4+["text/event-stream"]*4+["application/json"]*4,"Actual media differs")
 require([r["reconnect"] for r in rows]==[None]*6+["malformed"]*2+[None]*4,"Reconnect input differs")
 first=rows[0]["context"];require(all(first) and first[2]=="pos-1" and rows[1]["context"]==first,"Initial context absent")
 require(rows[10]["context"]==first[:2]+["pos-2"],"Real position advance absent")
 for i in [2,3,4,5,6,7,8,9,11]:require(rows[i]["context"]==[None,None,None],"Expected omitted/contextless tuple differs")
 for r in rows:
  require(type(r["bodyOctets"]) is int and 0<=r["bodyOctets"]<=32768,"Body bound failed")
  if r["method"]=="HEAD":require(r["bodyOctets"]==0,"HEAD body present")
 for i in [2,3,4,5,6,7,9,11]:require(rows[i]["contentType"]=="application/problem+json" and rows[i]["contentLength"]=="118","Problem metadata differs")
 return {"verdict":"EXPECTED_NAMED_CONTROL_FAILURE","actualTestExit":1,"namedFailures":seen,"httpResponses":12,"sourceObservation":o,"resultSha256":sha(P/"run/control.results.json"),"logSha256":sha(log),"limitation":"Emitted source-consumer sensitivity only; full body prerequisites assert in memory, response log does not retain full body/wire. Not installed qualification."}
def main():
 require(sys.argv[1:]==[NAME],"One exact control only")
 require(sha(BASE)==BASE_SHA and sha(P/"command.json")==COMMAND_SHA,"Reviewed runner/command changed")
 gate=read(P/"execution-gate.json")
 require(gate["status"]=="released" and gate["session"]==SESSION and gate["authorizedStages"]==[NAME] and gate["installedGate"]=="closed","Root gate closed/broader")
 require(gate["packetSha256"]==PACKET_SHA and gate["supervisorSha256"]==sha(Path(__file__)) and gate["baseSupervisorSha256"]==BASE_SHA,"Gate bindings differ")
 require(gate["commandsSha256"]==COMMAND_SHA,"Command gate differs")
 authority=Path(gate["rootAcceptance"]["path"]);require(authority.is_relative_to(D.parent) and sha(authority)==gate["rootAcceptance"]["sha256"],"Root acceptance differs")
 verify_graph()
 spec=importlib.util.spec_from_file_location("reviewed_changefeed_source_supervisor",BASE);module=importlib.util.module_from_spec(spec);spec.loader.exec_module(module)
 module.BINDING_SHA=gate["sourceBindingSha256"]
 original_source=module.source_hashes
 def checked_source():verify_graph();return original_source()
 module.source_hashes=checked_source
 def load_control(name):
  require(name==NAME,"Control identity differs")
  require(sha(module.COMMANDS)==module.COMMANDS_SHA and sha(module.GUARD)==module.GUARD_SHA,"Source commands/session guard changed")
  source=module.source_hashes();binding=read(module.BINDING)
  require(module.source_patch_hash()==binding["diffSha256"],"Draft differs")
  for prior in module.ORDER:
   r=read(D/"run"/(prior+".exit.json"))
   require(r["exit"]==0 and r["supervisorExit"]==0 and r["sourceUnchanged"] and r["sourceSha256After"]==source,"Positive prerequisite failed")
   require(not any(r[k] for k in ["errors","timeout","forcedCleanup","remainingAfterCleanup","sampledPidsStillPresent","runtimeDirectoriesStillPresent"]),"Positive cleanup failed")
   module.adjudicate_source(next(x for x in read(module.COMMANDS)["commands"] if x["id"]==prior))
  command=read(P/"command.json")
  require(command["cwd"]==str(C) and command["cases"]==1 and command["http"]==12 and command["productChildren"]==0,"Control bounds changed")
  require(command["argv"][0]=="/Users/dbox/.nvm/versions/node/v24.16.0/bin/node" and Path(command["argv"][1]).resolve().is_relative_to(C),"Copied runner absent")
  for arg in command["argv"]:
   if arg.startswith("--outputFile="):require(not Path(arg.split("=",1)[1]).exists(),"Prior result exists; no retry")
  command["workload"]={k:command[k] for k in ["cases","http","owners","directories","listeners","productChildren"]}
  return {**gate,"sourceSha256":source,"sourceDiffSha256":binding["diffSha256"],"preparationSha256":sha(Path(__file__)),"controls":{NAME:read(P/"mutation.json")}}, {**command,"control":NAME}
 module.load_gate=load_control;module.E=P/"run"
 code=module.main();outcome={"verdict":"CONTROL_NOT_ACCEPTED","actualSupervisorExit":code}
 try:graph=verify_graph();outcome={**adjudicate(),"graph":graph}
 except Exception as e:outcome["reason"]=str(e)
 with (P/"run/adjudication.json").open("x") as f:json.dump(outcome,f,indent=2);f.write("\n")
 print(json.dumps(outcome),flush=True)
 return 0 if outcome["verdict"]=="EXPECTED_NAMED_CONTROL_FAILURE" else 1
if __name__=="__main__":sys.exit(main())
