"""Copy/hash only; no product import, test, build or install. Invoke under guarded-action.py.
Requires root generation gate, repaired source freeze and actual unchanged positive build; refuses an existing mirror.
"""
import difflib, hashlib, json, os, re, shutil, subprocess
from pathlib import Path
P=Path(__file__).resolve().parent
D=P.parent
W=Path("/Users/dbox/repos/wt/vickie/vickie-authorization-transition")
C=P/"mirror"
B=None
B_SHA=None
BUILD=D/"source-validation-v4/run/build.exit.json"
GRAPH=P/"positive-graph.json"
GRAPH_SHA='6b461f78fb1fa5c0834a7ba94225dff8ad8f2cd0aca43d87bdd17d2ea1b06a6f'
MUTATION=P/"mutation-specification.json"
MUTATION_SHA='cdcaab6b55731c27efa6090297f6d50ceb29506a56c3bde69f05be38019ad38c'
def require(ok,message):
 if not ok: raise RuntimeError(message)
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def write(p,v):p.write_text(json.dumps(v,indent=2)+"\n")
def main():
 global B,B_SHA
 gate=json.loads((P/"generation-gate.json").read_text())
 require(gate["status"]=="released-for-copy-only" and gate["session"]=="2026-10-08T20:26:35Z-81e5","Generation gate closed")
 require(gate["executionGate"]=="closed","Control execution cannot be released by generator")
 require(sha(Path(__file__))==gate["generatorSha256"],"Generator changed")
 require(sha(GRAPH)==GRAPH_SHA==gate["provisionalGraphSha256"] and sha(MUTATION)==MUTATION_SHA,"Graph or mutation drift")
 approval=Path(gate["rootAcceptance"]["path"])
 require(approval.resolve().is_relative_to(D.parent.resolve()) and sha(approval)==gate["rootAcceptance"]["sha256"],"Root acceptance unavailable")
 B=Path(gate["sourceBindingPath"]);B_SHA=gate["sourceBindingSha256"]
 require(B.resolve().is_relative_to(D.parent.resolve()) and sha(B)==B_SHA,"Final repaired source binding unavailable")
 b=json.loads(B.read_text());build=json.loads(BUILD.read_text());graph=json.loads(GRAPH.read_text())
 require(sha(BUILD)==graph["buildExit"]["sha256"],"Historical build record changed")
 require(build["exit"]==0 and build["supervisorExit"]==0 and build["sourceUnchanged"],"Actual build prerequisite failed")
 require(not C.exists(),"Mirror exists; preserve prior attempt")
 for n,i in b["files"].items():require(sha(W/n)==i["sha256"],"Frozen source differs: "+n)
 require(subprocess.check_output(["git","rev-parse","HEAD"],cwd=W,text=True).strip()==b["head"]==graph["head"],"Base changed")
 actual_new=set(subprocess.check_output(["git","ls-files","--others","--exclude-standard"],cwd=W,text=True).splitlines())
 require(actual_new=={n for n,i in b["files"].items() if i["new"]},"Untracked scope changed")
 changed=set(subprocess.check_output(["git","diff","HEAD","--name-only"],cwd=W,text=True).splitlines())|actual_new
 require(changed==set(b["files"]),"Changed source scope differs")
 require(not subprocess.check_output(["git","diff","--cached","--name-only"],cwd=W).strip(),"Unexpected staged source")
 patch=subprocess.check_output(["git","diff","--binary","HEAD"],cwd=W)
 for n,i in b["files"].items():
  if i["new"]:
   r=subprocess.run(["git","diff","--no-index","--","/dev/null",n],cwd=W,capture_output=True)
   require(r.returncode==1,"New-file diff failed");patch+=r.stdout
 require(hashlib.sha256(patch).hexdigest()==b["diffSha256"],"Combined full patch changed")
 # Only the declared T11 test differs from the actual v4 build; current capture is exact.
 require(sha(B)==graph["currentSourceBinding"]["sha256"],"Current source binding differs from inventory")
 actual_input_changes=[n for n,i in b["files"].items() if i["sha256"]!=build["sourceSha256After"][n]]
 require(actual_input_changes==graph["testOnlyBuildBridge"]==gate["repairedTestPaths"]==["packages/server/src/transaction-authorization.test.ts"],"Actual build test-only bridge differs")
 for n,i in graph["sourceInputs"].items():require(sha(W/n)==i["sha256"],"Current captured input changed: "+n)
 for n,i in graph["emitted"].items():require(sha(W/n)==i["sha256"],"Positive emitted bytes changed: "+n)
 for n,i in graph["dependencyFiles"].items():require(sha(W/n)==i["sha256"],"Tool dependency changed: "+n)
 for n,i in graph["dependencyLinks"].items():require((W/n).is_symlink() and os.readlink(W/n)==i["link"] and str((W/n).resolve(strict=True).relative_to(W))==i["resolvedRelative"],"Dependency link changed: "+n)
 tracked=set(subprocess.check_output(["git","ls-files"],cwd=W,text=True).splitlines())
 tracked.update(n for n,i in b["files"].items() if i["new"])
 emitted={str(p.relative_to(W)) for top in ["packages","apps"] for dist in (W/top).glob("*/dist") for p in dist.rglob("*") if p.is_file()}
 require(emitted==set(graph["emitted"]),"Actual emitted inventory changed")
 require(tracked==set(graph["sourceInputs"]),"Tracked input inventory changed")
 files={}
 C.mkdir()
 for n in sorted(tracked|emitted):
  source=W/n;target=C/n
  require(source.is_file() and not source.is_symlink(),"Unexpected product symlink/nonfile: "+n)
  target.parent.mkdir(parents=True,exist_ok=True);shutil.copy2(source,target)
  require(not os.path.samefile(source,target) and sha(source)==sha(target),"Copy alias/drift: "+n)
  files[n]={"positiveSha256":sha(source),"copySha256":sha(target),"bytes":source.stat().st_size,"emitted":n in emitted}
 roots=[W/"node_modules"]+[p/"node_modules" for top in ["packages","apps"] for p in (W/top).iterdir() if p.is_dir() and (p/"node_modules").exists()]
 deps={};links={}
 for root in roots:
  for base,dirs,names in os.walk(root,followlinks=False):
   dirs[:]=[n for n in dirs if n not in (".vite",".vite-temp")]
   for name in dirs+names:
    source=Path(base)/name;n=str(source.relative_to(W));target=C/n
    if source.is_symlink():
     real=source.resolve(strict=True);require(real.is_relative_to(W),"External dependency escape: "+n)
     destination=C/real.relative_to(W);target.parent.mkdir(parents=True,exist_ok=True)
     text=os.path.relpath(destination,target.parent);target.symlink_to(text,target_is_directory=real.is_dir())
     links[n]={"positiveLink":os.readlink(source),"positiveRealpath":str(real),"copyLink":text,"copyRealpath":str(destination)}
    elif source.is_file():
     target.parent.mkdir(parents=True,exist_ok=True);shutil.copy2(source,target)
     require(not os.path.samefile(source,target) and sha(source)==sha(target),"Tool copy drift: "+n)
     deps[n]={"sha256":sha(source),"bytes":source.stat().st_size}
 for n,i in links.items():require(str((C/n).resolve(strict=True))==i["copyRealpath"] and (C/n).resolve().is_relative_to(C),"Unresolved copied dependency: "+n)
 mutation=json.loads(MUTATION.read_text());target=mutation["target"];before=(C/target).read_bytes()
 needle=mutation["needle"].encode();replacement=mutation["replacement"].encode()
 require(hashlib.sha256(before).hexdigest()==mutation["beforeSha256"] and before.count(needle)==1,"Exact emitted method missing or changed")
 after=before.replace(needle,replacement)
 require(hashlib.sha256(after).hexdigest()==mutation["afterSha256"],"Proposed mutation differs")
 (C/target).write_bytes(after);files[target]["copySha256"]=sha(C/target)
 require([n for n,i in files.items() if i["copySha256"]!=i["positiveSha256"]]==[target],"Unexpected product delta")
 (P/"mutation.patch").write_text("".join(difflib.unified_diff(before.decode().splitlines(True),after.decode().splitlines(True),fromfile="positive/"+target,tofile="negative/"+target)))
 write(P/"mutation.json",{"target":target,"beforeSha256":hashlib.sha256(before).hexdigest(),"afterSha256":sha(C/target),"removedUtf8":needle.decode(),"replacementUtf8":replacement.decode(),"changedProductFiles":1,"sourceChanged":False,"mapsChanged":False})
 write(P/"copy-manifest.json",{"sourceBindingPath":str(B),"sourceBindingSha256":B_SHA,"buildExitPath":str(BUILD),"buildExitSha256":sha(BUILD),"files":files})
 write(P/"dependency-manifest.json",{"files":deps,"links":links,"excludedCaches":[".vite",".vite-temp"]})
 edges=[]
 for n in sorted(emitted):
  if not n.endswith(".js"):continue
  for rel in re.findall(r'(?:from\s+|import\s*\()\s*["\'](\.[^"\']+)["\']',(C/n).read_text()):
   dest=(C/n).parent/rel
   require(dest.exists() and dest.resolve().is_relative_to(C),"Emitted edge escape/missing: "+n+" -> "+rel)
   edges.append({"from":n,"specifier":rel,"to":str(dest.resolve().relative_to(C))})
 runner=(C/"node_modules/vitest/vitest.mjs").resolve(strict=True)
 require(runner.is_relative_to(C),"Runner escaped")
 source_test=C/"packages/server/src/transaction-runtime.test.ts"
 require('new URL("../dist/transaction-runtime.js", import.meta.url).href' in source_test.read_text(),"Dynamic emitted fixture changed")
 write(P/"resolution.json",{"runner":str(runner),"test":str(source_test),"dynamicRuntime":str(C/"packages/server/dist/transaction-runtime.js"),"runtimeStore":str(C/target),"relativeEmittedEdges":edges,"limitation":"Static relative-import inventory; actual named negative response remains required. Every package/tool symlink resolves within mirror; all eight workspace source aliases copied."})
 title="narrows current Transactional visibility and fences issued handles after policy publication"
 write(P/"command.json",{"id":"suppress-policy-publication","cwd":str(C),"argv":["/Users/dbox/.nvm/versions/node/v24.16.0/bin/node",str(runner),"run","packages/server/src/transaction-runtime.test.ts","--maxWorkers=1","--reporter=verbose","--reporter=json","--outputFile="+str(P/"run/control.results.json"),"-t","^"+re.escape(title)+"$"],"selectedFullNames":[title],"cases":1,"expectedHttp":7,"maximumHttp":17,"owners":1,"listeners":1,"directories":1,"productChildren":0,"ownedTempPrefixes":["bdp-authorization-http-"],"supervisorSeconds":120,"expectedRawTestExit":1})
 require(set(deps)==set(graph["dependencyFiles"]) and set(links)==set(graph["dependencyLinks"]),"Tool inventory changed during copy")
 for n,i in deps.items():require(i["sha256"]==graph["dependencyFiles"][n]["sha256"] and sha(C/n)==i["sha256"] and sha(W/n)==i["sha256"],"Tool bytes drifted during copy")
 for n,i in files.items():require(sha(W/n)==i["positiveSha256"] and sha(C/n)==i["copySha256"],"Product bytes drifted during copy")
 write(P/"build-source-bridge.json",{"historicalBuild":graph["buildExit"],"historicalSource":graph["historicalSourceBinding"],"currentSource":{"path":str(B),"sha256":B_SHA},"actualChangedInputPaths":actual_input_changes,"allEmittedBytesEqual":True,"notAFreshBuild":True})
 write(P/"preparation-manifests.json",{"sourceBindingSha256":B_SHA,"files":{n:sha(P/n) for n in ["copy-manifest.json","dependency-manifest.json","resolution.json","mutation.json","mutation.patch","build-source-bridge.json"]}})
 print(json.dumps({"productMembers":len(files),"emittedMembers":len(emitted),"toolFiles":len(deps),"links":len(links),"relativeEmittedEdges":len(edges),"packetSha256":sha(P/"preparation-manifests.json"),"commandSha256":sha(P/"command.json")}))
if __name__=="__main__":main()
