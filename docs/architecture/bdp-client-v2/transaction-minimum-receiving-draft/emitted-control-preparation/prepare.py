"""Copy/hash only; no product import, test, build or install. Invoke under guarded-action.py.
Requires the actual successful source-v3 build; refuses an existing mirror.
"""
import difflib, hashlib, json, os, re, shutil, subprocess
from pathlib import Path
P=Path(__file__).resolve().parent
D=P.parent
W=Path("/Users/dbox/repos/wt/vickie/vickie-minimum-receiving")
C=P/"mirror"
B=D/"receiving-test-repair/source-binding.json"
B_SHA="b7082c68869b1cecd715ed8d6e0ef88c6ba064a2e698beddf02844ed4dc252fb"
BUILD=D/"run-v3/build.exit.json"
def require(ok,message):
 if not ok: raise RuntimeError(message)
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def write(p,v):p.write_text(json.dumps(v,indent=2)+"\n")
def main():
 require(sha(B)==B_SHA,"Source binding differs")
 b=json.loads(B.read_text());build=json.loads(BUILD.read_text())
 require(build["exit"]==0 and build["supervisorExit"]==0 and build["sourceUnchanged"],"Actual build prerequisite failed")
 require(build["sourceSha256After"]=={n:i["sha256"] for n,i in b["files"].items()},"Build/source mismatch")
 require(not C.exists(),"Mirror exists; preserve prior attempt")
 for n,i in b["files"].items():require(sha(W/n)==i["sha256"],"Frozen source differs: "+n)
 require(subprocess.check_output(["git","rev-parse","HEAD"],cwd=W,text=True).strip()==b["head"],"Base changed")
 tracked=set(subprocess.check_output(["git","ls-files"],cwd=W,text=True).splitlines())
 tracked.update(n for n,i in b["files"].items() if i["new"])
 emitted={str(p.relative_to(W)) for top in ["packages","apps"] for dist in (W/top).glob("*/dist") for p in dist.rglob("*") if p.is_file()}
 require("packages/server/dist/transaction-http.js" in emitted,"Actual emitted HTTP missing")
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
 target="packages/server/dist/transaction-http.js";before=(C/target).read_bytes()
 needle=b'        if (malformedTransactionMinimumHeaders(request.headers))\n            return transactionHttpProblem("malformed-request");\n'
 require(before.count(needle)==1,"Sole emitted gate bytes not found")
 after=before.replace(needle,b'');(C/target).write_bytes(after);files[target]["copySha256"]=sha(C/target)
 require([n for n,i in files.items() if i["copySha256"]!=i["positiveSha256"]]==[target],"Unexpected product delta")
 (P/"mutation.patch").write_text("".join(difflib.unified_diff(before.decode().splitlines(True),after.decode().splitlines(True),fromfile="positive/"+target,tofile="negative/"+target)))
 write(P/"mutation.json",{"target":target,"beforeSha256":hashlib.sha256(before).hexdigest(),"afterSha256":sha(C/target),"removedUtf8":needle.decode(),"changedProductFiles":1,"sourceChanged":False,"mapsChanged":False})
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
 source_test=C/"packages/server/src/transaction-http.test.ts"
 require('new URL("../dist/transaction-runtime.js", import.meta.url).href' in source_test.read_text(),"Dynamic emitted fixture changed")
 write(P/"resolution.json",{"runner":str(runner),"test":str(source_test),"dynamicRuntime":str(C/"packages/server/dist/transaction-runtime.js"),"runtimeHttp":str(C/target),"relativeEmittedEdges":edges,"limitation":"Static relative-import inventory; actual named negative response remains required. Every package/tool symlink resolves within mirror; all eight workspace source aliases copied."})
 title="enforces the minimum-receiving-malformed-request boundary over the emitted listener"
 write(P/"command.json",{"id":"omit-receiving-gate","cwd":str(C),"argv":["/Users/dbox/.nvm/versions/node/v24.16.0/bin/node",str(runner),"run","packages/server/src/transaction-http.test.ts","--maxWorkers=1","--reporter=verbose","--reporter=json","--outputFile="+str(P/"run/control.results.json"),"-t","^"+re.escape(title)+"$"],"selectedFullNames":[title],"cases":1,"http":1,"owners":1,"listeners":1,"directories":1,"productChildren":0,"ownedTempPrefixes":["bdp-tx-http-"],"supervisorSeconds":120,"expectedRawTestExit":1})
 write(P/"preparation-manifests.json",{"sourceBindingSha256":B_SHA,"files":{n:sha(P/n) for n in ["copy-manifest.json","dependency-manifest.json","resolution.json","mutation.json","mutation.patch"]}})
 print(json.dumps({"productMembers":len(files),"emittedMembers":len(emitted),"toolFiles":len(deps),"links":len(links),"relativeEmittedEdges":len(edges),"packetSha256":sha(P/"preparation-manifests.json"),"commandSha256":sha(P/"command.json")}))
if __name__=="__main__":main()
