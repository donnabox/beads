"""One closed emitted-control invocation; reuses the reviewed source supervisor.
No pnpm/shim dispatch, setup, build, installed child, retry or positive-source edit.
"""
import hashlib, importlib.util, json, os, re, sys
from pathlib import Path
P=Path(__file__).resolve().parent
D=P.parent
C=P/"mirror"
W=Path("/Users/dbox/repos/wt/vickie/vickie-authorization-transition")
BASE=D/"source-validation-t11/run-source.py"
BASE_SHA='ea48c657f7ab814014f1e91abf9b49bf4490e82f0cce17795912c6e13521389e'
COMMAND_SHA='36151ef241463a99abe2cb466dd71ac47e95c2bd5f1082affcb8c099713bb858'
PACKET_SHA='8bfb392db3e7ab4267f33e378f4ef1be79051a84d8f32eab9c9f76a1626faa25'
NAME="suppress-policy-publication"
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
 require(len(failures)==1 and "authorization-transition:generation-and-lease" in failures[0] and "AssertionError" in failures[0],"Expected sole named assertion failure absent")
 observations=[]
 log=P/"run"/(NAME+".log")
 for line in re.sub(r"\x1b\[[0-9;]*m","",log.read_text()).splitlines():
  try:v=json.loads(line.strip())
  except ValueError:continue
  if isinstance(v,dict) and v.get("label") in ["authorization-transition-before-cleanup","authorization-transition-cleanup"]:observations.append(v)
 require(len(observations)==2,"Expected before/after cleanup observations absent or duplicated")
 before,cleanup=observations
 require(before["label"]=="authorization-transition-before-cleanup" and before["name"]=="visibility-handles" and before["attempts"]==7 and before["failures"]==1,"Exact seven-request failure prerequisites absent")
 wires=[v for v in before["observations"] if "index" in v]
 require([v["index"] for v in wires]==list(range(1,8)),"Seven actual wire observations missing")
 require([v["method"] for v in wires]==["POST","POST","GET","GET","GET","GET","GET"],"Wire methods differ")
 require([v["target"] for v in wires]==["/acme/operations/batch","/acme/operations/batch","/acme/beads/b00","/acme/beads/?limit=2","/acme/snapshot","/acme/events/","/acme/changes/?start=now"],"Wire targets differ")
 require(all(v["status"]==200 for v in wires),"Positive wire prerequisites failed")
 require([v["requestBodyBytes"] for v in wires[:2]]==[3728,132],"Seed bodies changed")
 require([v["requestBodySha256"] for v in wires[:2]]==["a778feb6f8559743228db36139ccc653370c43a30cdcf3d2017db2e85349b054","8f22a1dee8d5c16750f476c2c0f4883ca91c257014e07b9efef9d2e678a98220"],"Actual request fingerprints differ")
 require(all(0<v["bodyBytes"]<=32768 and re.fullmatch(r"[0-9a-f]{64}",v["bodySha256"]) for v in wires[:6]),"Finite response bounds/fingerprints missing")
 context=wires[0]["context"]
 require(len(context)==3 and all(isinstance(v,str) and v for v in context) and context[2]=="pos-1","Initial observed context missing")
 require(all(v["context"]==[context[0],context[1],"pos-2"] for v in wires[1:]),"Current wire contexts differ")
 require(wires[-1].get("stream") is True and wires[-1].get("closed") is not True and not wires[-1].get("failedBeforeHeaders"),"Stream not open before cleanup")
 policy=[v for v in before["observations"] if "publication" in v]
 require(len(policy)==1 and len(before["observations"])==8,"Publication observation absent/duplicated")
 require(policy[0]=={"publication":{"revision":1,"bindingKind":"scope","generation":0,"authorizationView":context[1],"readPolicy":"all-visible","canMutate":True},"oldStreamClosedAtPublication":False},"Coherent old policy was not observed")
 require(cleanup["label"]=="authorization-transition-cleanup" and cleanup["name"]=="visibility-handles" and cleanup["attempts"]==7 and cleanup["failures"]==1 and cleanup["runtimeReturned"] is True and cleanup["closed"] is True and cleanup["directoryRemoved"] is True and cleanup["claimEligible"] is False,"Source lifecycle cleanup failed")
 directory=Path(cleanup["directory"])
 require(directory.is_relative_to(P/"run"/(NAME+".tmp")) and directory.name.startswith("bdp-authorization-http-") and not directory.exists(),"Reported owned store remains or escaped")
 return {"verdict":"EXPECTED_NAMED_CONTROL_FAILURE","actualTestExit":1,"namedFailure":"authorization-transition:generation-and-lease","httpExchanges":7,"safeObservations":observations,"resultSha256":sha(P/"run/control.results.json"),"logSha256":sha(log),"stderrSha256":sha(P/"run"/(NAME+".stderr.log")),"limitation":"One emitted source-consumer sensitivity only. Six finite responses plus one open stream; raw body schema/full assertions execute before the named failure but are not reconstructed from fingerprints. No installed qualification or network-byte attestation."}

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
 module.BINDING=Path(gate["sourceBindingPath"])
 module.BINDING_SHA=gate["sourceBindingSha256"]
 module.OWNED_PATHS=list(read(module.BINDING)["files"])
 require(module.BINDING==D/"source-validation-t11/source-binding.json" and module.BINDING_SHA==read(P/"copy-manifest.json")["sourceBindingSha256"],"Final source binding differs")
 original_source=module.source_hashes
 def checked_source():verify_graph();return original_source()
 module.source_hashes=checked_source
 def load_control(name):
  require(name==NAME,"Control identity differs")
  require(sha(module.COMMANDS)==module.COMMANDS_SHA and sha(module.GUARD)==module.GUARD_SHA,"Source commands/session guard changed")
  source=module.source_hashes();binding=read(module.BINDING)
  require(module.source_patch_hash()==binding["diffSha256"],"Draft differs")
  module.verify_prior_evidence()
  # Historical v4 owner stage remains failed; its ten passed cases are bound by
  # prior-evidence, while the only repaired case must actually pass here.
  for prior in module.ORDER:
   r=read(D/"source-validation-t11/run"/(prior+".exit.json"))
   require(r["exit"]==0 and r["supervisorExit"]==0 and r["sourceUnchanged"] and r["sourceSha256After"]==source,"Targeted positive prerequisite failed")
   require(not any(r[k] for k in ["errors","timeout","forcedCleanup","remainingAfterCleanup","sampledPidsStillPresent","runtimeDirectoriesStillPresent"]),"Positive cleanup failed")
   module.adjudicate_source(next(x for x in read(module.COMMANDS)["commands"] if x["id"]==prior))
  command=read(P/"command.json")
  require(command["cwd"]==str(C) and command["cases"]==1 and command["expectedHttp"]==7 and command["maximumHttp"]==17 and command["productChildren"]==0,"Control bounds changed")
  require(command["argv"][0]=="/Users/dbox/.nvm/versions/node/v24.16.0/bin/node" and Path(command["argv"][1]).resolve().is_relative_to(C),"Copied runner absent")
  for arg in command["argv"]:
   if arg.startswith("--outputFile="):require(not Path(arg.split("=",1)[1]).exists(),"Prior result exists; no retry")
  command["workload"]={k:command[k] for k in ["cases","expectedHttp","maximumHttp","owners","directories","listeners","productChildren"]}
  return {**gate,"sourceSha256":source,"sourceDiffSha256":binding["diffSha256"],"preparationSha256":sha(Path(__file__)),"controls":{NAME:read(P/"mutation.json")}}, {**command,"control":NAME}
 module.load_gate=load_control;module.E=P/"run"
 code=module.main();outcome={"verdict":"CONTROL_NOT_ACCEPTED","actualSupervisorExit":code}
 try:graph=verify_graph();outcome={**adjudicate(),"graph":graph}
 except Exception as e:outcome["reason"]=str(e)
 with (P/"run/adjudication.json").open("x") as f:json.dump(outcome,f,indent=2);f.write("\n")
 print(json.dumps(outcome),flush=True)
 return 0 if outcome["verdict"]=="EXPECTED_NAMED_CONTROL_FAILURE" else 1
if __name__=="__main__":sys.exit(main())
