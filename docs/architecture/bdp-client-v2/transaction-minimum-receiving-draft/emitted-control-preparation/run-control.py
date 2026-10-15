"""One closed emitted-control invocation; reuses the reviewed source supervisor.
No pnpm/shim dispatch, setup, build, installed child, retry or positive-source edit.
"""
import hashlib, importlib.util, json, os, re, sys
from pathlib import Path
P=Path(__file__).resolve().parent
D=P.parent
C=P/"mirror"
W=Path("/Users/dbox/repos/wt/vickie/vickie-minimum-receiving")
BASE=D/"execution-preparation-v3/run-source.py"
BASE_SHA='468e6f32d9f7c2aa70ff1e769ae313d61782b83ee43443baed268b91d0b2c2b3'
COMMAND_SHA='6643f5a616966ac943ac9c77fbe3d90ba17c4dc218582dd7dbd5c1766cace5d5'
PACKET_SHA='73284d0875358b3b40f78249dc1a91b9e8acb30fb9a1a0a5df5509ab9fca6dcb'
NAME="omit-receiving-gate"
SESSION="2026-10-08T20:26:35Z-81e5"
def require(ok,message):
 if not ok: raise RuntimeError(message)
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def read(p):return json.loads(p.read_text())
def verify_graph():
 require(sha(P/"preparation-manifests.json")==PACKET_SHA,"Packet changed")
 for name,h in read(P/"preparation-manifests.json")["files"].items():require(sha(P/name)==h,"Manifest changed: "+name)
 manifest=read(P/"copy-manifest.json"); deps=read(P/"dependency-manifest.json")
 require(sha(Path(manifest["buildExitPath"]))==manifest["buildExitSha256"],"Build record changed")
 require(sha(Path(manifest["sourceBindingPath"]))==manifest["sourceBindingSha256"],"Source binding changed")
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
 failures=cases[0].get("failureMessages",[])
 require(len(failures)==1 and "minimum-receiving-malformed-request" in failures[0] and "AssertionError" in failures[0],"Expected sole named assertion failure absent")
 observations=[]
 log=P/"run"/(NAME+".log")
 for line in re.sub(r"\x1b\[[0-9;]*m","",log.read_text()).splitlines():
  try:v=json.loads(line.strip())
  except ValueError:continue
  if isinstance(v,dict) and v.get("observation")=="minimum-receiving-control":observations.append(v)
 require(observations==[{"observation":"minimum-receiving-control","requests":1,"status":500,"marker":"minimum-position","body":""}],"Actual old response prerequisites absent/duplicated")
 return {"verdict":"EXPECTED_NAMED_CONTROL_FAILURE","actualTestExit":1,"namedFailure":"minimum-receiving-malformed-request","httpResponses":1,"sourceObservation":observations[0],"resultSha256":sha(P/"run/control.results.json"),"logSha256":sha(log),"limitation":"One emitted source-consumer gate sensitivity only. Full positive Problem/schema checks belong to source positives. Control500 body is genuinely empty. Runtime owner/listener budget is source-derived; process/temp cleanup independently checked. Not installed qualification."}
def main():
 require(sys.argv[1:]==[NAME],"One exact control only")
 require(sha(BASE)==BASE_SHA and sha(P/"command.json")==COMMAND_SHA,"Reviewed runner/command changed")
 gate=read(P/"execution-gate.json")
 require(gate["status"]=="released" and gate["session"]==SESSION and gate["authorizedStages"]==[NAME] and gate["installedGate"]=="closed","Root gate closed/broader")
 require(gate["packetSha256"]==PACKET_SHA and gate["supervisorSha256"]==sha(Path(__file__)) and gate["baseSupervisorSha256"]==BASE_SHA,"Gate bindings differ")
 require(gate["commandsSha256"]==COMMAND_SHA,"Command gate differs")
 authority=Path(gate["rootAcceptance"]["path"]);require(authority.is_relative_to(D.parent) and sha(authority)==gate["rootAcceptance"]["sha256"],"Root acceptance differs")
 verify_graph()
 spec=importlib.util.spec_from_file_location("reviewed_minimum_source_supervisor",BASE);module=importlib.util.module_from_spec(spec);spec.loader.exec_module(module)
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
   r=read(D/"run-v3"/(prior+".exit.json"))
   require(r["exit"]==0 and r["supervisorExit"]==0 and r["sourceUnchanged"] and r["sourceSha256After"]==source,"Positive prerequisite failed")
   require(not any(r[k] for k in ["errors","timeout","forcedCleanup","remainingAfterCleanup","sampledPidsStillPresent","runtimeDirectoriesStillPresent"]),"Positive cleanup failed")
   module.adjudicate_source(next(x for x in read(module.COMMANDS)["commands"] if x["id"]==prior))
  command=read(P/"command.json")
  require(command["cwd"]==str(C) and command["cases"]==1 and command["http"]==1 and command["productChildren"]==0,"Control bounds changed")
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
