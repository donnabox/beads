"""One closed emitted-control invocation; reuses the reviewed source supervisor.
No pnpm/shim dispatch, setup, build, installed child, retry or positive-source edit.
"""
import hashlib, importlib.util, json, os, re, sys
from pathlib import Path
P=Path(__file__).resolve().parent
D=P.parent
C=P/"mirror"
W=Path("/Users/dbox/repos/wt/vickie/vickie-sequence-member-evidence")
BASE=D/"source-selection-repair/run-source.py"
BASE_SHA='0aa4431d69531603aafa0a02b9857a8d6f89e9b12220c9a21e27e5e0583a847a'
COMMAND_SHA='59cac9bc5ca3b4035517e5f40a1dc43a5ba8e0d8e1b04e9401ee332de8612650'
PACKET_SHA='2a5d92822153b08235f2dc8f246d6db96c942a1fed6b6f20373a29f16267a70c'
NAME="early-depth-control"
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
 require(sha(Path(manifest["buildManifestPath"]))==manifest["buildManifestSha256"],"Build manifest changed")
 require([n for n,v in manifest["files"].items() if v["positiveSha256"]!=v["copySha256"]]==["packages/server/dist/transaction-store.js"],"Exactly one emitted store mutation required")
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
 require(result["exit"]==1 and result["supervisorExit"]==1,"Expected actual test exit1 absent")
 require(not result["timeout"] and not result["forcedCleanup"] and not result["errors"] and result["sourceUnchanged"],"Execution/source failure")
 for field in ("remainingAfterCleanup","sampledPidsStillPresent","runtimeDirectoriesStillPresent"):require(not result[field],"Owned cleanup failed: "+field)
 report=read(P/"run/control.results.json")
 cases=[c for f in report["testResults"] for c in f.get("assertionResults",[]) if c["status"] not in ("pending","skipped","todo")]
 command=read(P/"command.json")
 require(len(cases)==1 and cases[0]["status"]=="failed" and cases[0]["fullName"]==command["selectedFullNames"][0],"Selected failed case mismatch")
 require(report["success"] is False and report["numFailedTests"]==1 and report["numPassedTests"]==0,"Aggregate mismatch")
 require(not report.get("unhandledErrors") and not report.get("errors") and not report.get("numRuntimeErrorTestSuites",0),"Unhandled runtime failure")
 failures=cases[0].get("failureMessages",[])
 require(len(failures)==1 and "depth-key:batch:conflict-before-depth" in failures[0] and "AssertionError" in failures[0],"Exactly one named status assertion required")
 log=(P/"run"/(NAME+".log")).read_text()
 def record(label):
  rows=[json.loads(line.split(label,1)[1]) for line in log.splitlines() if line.startswith(label)]
  require(len(rows)==1,"Expected one genuine "+label+" record")
  return rows[0]
 observation=record("depth-key-source-observation ")
 cleanup=record("depth-key-source-cleanup ")
 require(observation["attempts"]==2 and observation["primaryFailureCount"]==1,"Wrong observed control progress")
 responses=observation["responses"];require(len(responses)==2,"Expected two actual responses")
 expected=[
  ("batch-seed","application/json",200,150,"3db7ee64eb95248fc9f50347c1b7c482f00ad64c411aa81c15ff65703b1b8236"),
  ("batch-deep-conflict","text/plain",500,605,"7f559c520455c5a1ba45c73cf27cee9e58076b9b033ac7d8259408c03f2d0b8b"),
 ]
 for response,(label,accept,status,size,digest) in zip(responses,expected,strict=True):
  require(response["label"]==label and response["method"]=="POST" and response["accept"]==accept and response["status"]==status,"Wrong observed route/status")
  require(re.fullmatch(r"http://127\.0\.0\.1:[1-9][0-9]*/acme/operations/batch",response["url"]) is not None,"Wrong actual loopback target")
  require(response["requestBodyBytes"]==size and response["requestBodySha256"]==digest,"Request body changed")
  require(response["headers"]["x-bdp-development"]=="non-attesting","Private marker absent")
 require(responses[0]["url"]==responses[1]["url"],"Different listener target")
 seed=responses[0];refusal=responses[1]
 require(seed["headers"]["bdp-scope-position"]=="pos-1" and all(isinstance(seed["headers"][h],str) and seed["headers"][h] for h in ["bdp-scope-epoch","bdp-authorization-view"]),"Genuine seed context absent")
 require(seed["responseBodyBytes"]>0 and seed["headers"]["content-type"]=="application/json","Seed response absent")
 require(refusal["headers"]["x-bdp-development-limit"]=="unsupported-transaction-composition" and refusal["responseBodyBytes"]==0 and refusal["responseBodySha256"]==hashlib.sha256(b"").hexdigest(),"Private refusal differs")
 require(all(refusal["headers"][h] is None for h in ["bdp-scope-epoch","bdp-authorization-view","bdp-scope-position"]),"Refusal context changed")
 require(cleanup=={"cleanupActions":2,"cleanupActionsCompleted":2,"cleanupCompleted":True,"primaryOrCleanupFailureCount":1},"Primary/cleanup accounting differs")
 # Supervisor already checks sampled processes/runtime roots; recheck PIDs now.
 import subprocess
 present=subprocess.run(["ps","-p",",".join(map(str,result["sampledPids"])),"-o","pid="],capture_output=True,text=True)
 require(present.returncode in (0,1) and not present.stdout.strip(),"Sampled process remains")
 return {"verdict":"EXPECTED_NAMED_CONTROL_FAILURE","actualTestExit":1,"actualSupervisorExit":1,"requests":2,"responses":responses,"cleanup":cleanup,"failureMessages":failures,"resultsSha256":sha(P/"run/control.results.json"),"logSha256":sha(P/"run"/(NAME+".log")),"sourceAttribution":"External copied source/emitted/tool binding; no child attestation or public qualification"}

def main():
 require(sys.argv[1:]==[NAME],"One exact control only")
 require(sha(BASE)==BASE_SHA and sha(P/"command.json")==COMMAND_SHA,"Reviewed runner/command changed")
 gate=read(P/"execution-gate.json")
 require(gate["status"]=="released" and gate["session"]==SESSION and gate["authorizedStages"]==[NAME] and gate["installedGate"]=="closed","Root gate closed/broader")
 require(gate["packetSha256"]==PACKET_SHA and gate["supervisorSha256"]==sha(Path(__file__)) and gate["baseSupervisorSha256"]==BASE_SHA,"Gate bindings differ")
 require(gate["commandsSha256"]==COMMAND_SHA,"Command gate differs")
 authority=Path(gate["rootAcceptance"]["path"]);require(authority.is_relative_to(D.parent) and sha(authority)==gate["rootAcceptance"]["sha256"],"Root acceptance differs")
 verify_graph()
 spec=importlib.util.spec_from_file_location("reviewed_depth_source_supervisor",BASE);module=importlib.util.module_from_spec(spec);spec.loader.exec_module(module)
 module.BINDING_SHA=gate["sourceBindingSha256"]
 original_source=module.source_hashes
 def checked_source():verify_graph();return original_source()
 module.source_hashes=checked_source
 def load_control(name):
  require(name==NAME,"Control identity differs")
  require(sha(module.COMMANDS)==module.COMMANDS_SHA and sha(module.GUARD)==module.GUARD_SHA,"Source commands/session guard changed")
  source=module.source_hashes();binding=read(module.BINDING)
  require(module.source_patch_hash()==binding["diffSha256"],"Draft differs")
  for prior in read(P/"prerequisites.json")["records"]:
   p=Path(prior["path"]);require(sha(p)==prior["sha256"],"Prerequisite bytes changed")
   if prior["kind"] not in ("static","corrected17source"):continue
   r=read(p)
   require(r["exit"]==0 and r["supervisorExit"]==0 and r["sourceUnchanged"] and r["sourceSha256After"]==source,"Positive prerequisite failed")
   require(not any(r[k] for k in ["errors","timeout","forcedCleanup","remainingAfterCleanup","sampledPidsStillPresent","runtimeDirectoriesStillPresent"]),"Positive cleanup failed")
  module.adjudicate_source(read(module.COMMANDS)["commands"][0])
  command=read(P/"command.json")
  require(command["cwd"]==str(C) and command["cases"]==1 and command["http"]==13 and command["httpExpected"]==2 and command["httpHardMaximum"]==13 and command["productChildren"]==0,"Control bounds changed")
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
