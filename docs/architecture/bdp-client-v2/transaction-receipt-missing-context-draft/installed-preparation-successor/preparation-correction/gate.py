"""Closed clean-commit/package gate. Does not release itself or import product."""
from pathlib import Path
import hashlib,json,re,subprocess
P=Path(__file__).resolve().parent
D=P.parent
W=Path("/Users/dbox/repos/wt/vickie/vickie-sequence-member-evidence")
E=P/"run"
L=D.parent
GUARD=L/"guarded-action.py"
PARENT="015cac1aa4fa827cc4a2c647dff08454c6c8e6f8"
BINDING=D/"stale-assertion-repair/source-binding.json"
BINDING_SHA="e96f012c8502285454eb3867f8b7b35c82c483ae3ceb3f7fc21eaa31127d0cf3"
COMMANDS=P/"commands.json"
COMMANDS_SHA="a302416a95191a4849ee784a8a50bd347e63a11296d1437b737eef43014a4245"
GUARD_SHA="fd81ba029cd4e7c7e60623df89e8292af91141d311d9ce789ed19e2152375e84"
sha=lambda b:hashlib.sha256(b).hexdigest()
def require(ok,message):
    if not ok:raise RuntimeError(message)
def git(*args):return subprocess.check_output(["git",*args],cwd=W,timeout=10)
def document(g,key):
    d=g[key];p=Path(d["path"])
    require(p.is_relative_to(L) and sha(p.read_bytes())==d["sha256"],"Changed bound document: "+key)
    return json.loads(p.read_text())
def source_hashes():
    require(sha(BINDING.read_bytes())==BINDING_SHA,"Source binding changed")
    expected={k:v["sha256"] for k,v in json.loads(BINDING.read_text())["files"].items()}
    actual={n:sha((W/n).read_bytes()) for n in expected}
    require(actual==expected,"Frozen source changed")
    require(not git("status","--porcelain").strip(),"Final clean source required")
    return actual
def source_patch_hash():return sha(git("diff","--binary","HEAD"))
def completed(name):
    x=json.loads((E/(name+".exit.json")).read_text())
    require(x["exit"]==x["supervisorExit"]==0 and x["sourceUnchanged"],"Prior stage failed: "+name)
    for key in ("timeout","forcedCleanup","errors","remainingAfterCleanup","sampledPidsStillPresent","runtimeDirectoriesStillPresent"):
        require(not x[key],"Unclean stage: "+name+" "+key)
    require(x["sourceSha256After"]==source_hashes(),"Prior source differs")
    return x
def build_outputs():
    paths=[W/"LICENSE"]
    for name in ("protocol","server"):
        root=W/"packages"/name;paths.append(root/"package.json")
        for sub in (("dist","schemas") if name=="protocol" else ("dist",)):
            paths.extend(p for p in (root/sub).rglob("*") if p.is_file())
    require(all(not p.is_symlink() for p in paths),"Package output symlink")
    return {p.relative_to(W).as_posix():sha(p.read_bytes()) for p in sorted(paths)}
def verify_build(g):
    inventory=document(g,"buildOutputInventory")
    reuse=document(g,"buildReuse")
    historical=reuse["historicalSourceBinding"]
    require(sha(Path(historical["path"]).read_bytes())==historical["sha256"]==inventory["historicalSourceBindingSha256"],"Historical source binding changed")
    old={p:x["sha256"] for p,x in json.loads(Path(historical["path"]).read_text())["files"].items()}
    require(inventory["sourceSha256"]==old,"Historical build attribution changed")
    require(reuse["successorSourceBinding"]=={"path":str(BINDING),"sha256":BINDING_SHA},"Wrong successor build reuse")
    require(sha((P/"build-output-inventory.json").read_bytes())==reuse["historicalBuildInventorySha256"],"Historical inventory changed")
    delta=reuse["allowedInputDelta"];path=delta["path"];current=source_hashes()
    require(path=="packages/server/src/transaction-http.test.ts" and {p for p in old if old[p]!=current[p]}=={path},"Unexpected build source delta")
    require(old[path]==delta["beforeSha256"] and current[path]==delta["afterSha256"],"Test delta hashes differ")
    data=(W/path).read_bytes();new=delta["newLiteral"].encode();previous=delta["oldLiteral"].encode()
    require(data.count(new)==1 and sha(data.replace(new,previous,1))==old[path],"Not exact one-assertion repair")
    correction=reuse["correction"]
    require(sha(Path(correction["path"]).read_bytes())==correction["sha256"],"Correction evidence changed")
    for k in ("buildExit","buildStart"):
        x=inventory[k];require(sha(Path(x["path"]).read_bytes())==x["sha256"],"Historical build evidence changed")
    result=json.loads(Path(inventory["buildExit"]["path"]).read_text())
    require(result["exit"]==result["supervisorExit"]==0 and result["sourceUnchanged"],"Historical build did not pass")
    start=json.loads(Path(inventory["buildStart"]["path"]).read_text())
    require(start["sourceSha256"]==old and start["argv"]==[g["node"],g["pnpm"],"build"],"Historical build command/source differs")
    inputs=dict(inventory["inputFiles"]);inputs[path]=current[path]
    names=git("ls-files").decode().splitlines()
    require({n:sha((W/n).read_bytes()) for n in names}==inputs,"Build inputs differ beyond exact test assertion")
    require(len(inventory["files"])==312 and inventory["files"]==build_outputs(),"Build output bytes/footprint changed")
    emitted={p.relative_to(W).as_posix():sha(p.read_bytes()) for scope in ("apps","packages") for root in (W/scope).iterdir() if root.is_dir() for p in (root/"dist").rglob("*") if p.is_file()}
    require(len(emitted)==440 and emitted==reuse["emittedFiles"],"Emitted graph changed")
def load_gate(name):
    require(__debug__,"Python assertions required")
    g=json.loads((P/"execution-gate.json").read_text())
    require(g.get("released") is True and g["session"]=="2026-10-08T20:26:35Z-81e5","Root gate closed")
    for key in ("rootAcceptance","councilAcceptance","controlAcceptance","finalSource"):
        document(g,key)
    names={"gate.py","supervise.py","packages.py","review-installed.py","commands.json","build-output-inventory.json","build-reuse.json"}
    require(set(g["preparationSha256"])==names,"Incomplete preparation pins")
    for n,h in g["preparationSha256"].items():require(sha((P/n).read_bytes())==h,"Preparation changed: "+n)
    require(sha(COMMANDS.read_bytes())==COMMANDS_SHA and sha(GUARD.read_bytes())==GUARD_SHA,"Commands/guard changed")
    cfg=json.loads(COMMANDS.read_text());order=cfg["executionOrder"]
    require(name in order and name in g["approvedCommandIds"],"Stage not released")
    require(g["approvedCommandIds"]==order[:len(g["approvedCommandIds"])],"Release must be sequential prefix")
    require(re.fullmatch("[0-9a-f]{40}",g.get("finalHead") or "") is not None and g["finalHead"]!=PARENT,"Final commit pending")
    require(git("rev-parse","HEAD").decode().strip()==g["finalHead"] and git("rev-parse","HEAD^{tree}").decode().strip()==g["finalTree"],"Final identity changed")
    require(git("show","-s","--format=%P","HEAD").decode().strip()==PARENT,"Ordinary direct parent required")
    final=document(g,"finalSource")
    require(final["head"]==g["finalHead"] and final["tree"]==g["finalTree"] and final["sourceBindingSha256"]==BINDING_SHA,"Final source document differs")
    source=source_hashes()
    require(set(git("diff","--name-only",PARENT,"HEAD").decode().splitlines())==set(source),"Eight-file scope differs")
    require(sha((W/"pnpm-lock.yaml").read_bytes())==g["sourceLockSha256"],"Source lock differs")
    require(g["node"]=="/Users/dbox/.nvm/versions/node/v24.16.0/bin/node" and g["pnpm"]=="/Users/dbox/.cache/node/corepack/v1/pnpm/11.20.0/bin/pnpm.cjs","Tool paths changed")
    verify_build(g)
    spec=dict(cfg["commands"][name]);require(spec["supervisorSeconds"]==120,"Wrong time bound")
    for prior in order[:order.index(name)]:completed(prior)
    if spec["phase"] in ("installed","review"):
        p=document(g,"installedProvenance")
        require(p["sourceCommit"]==g["finalHead"] and p["sourceTree"]==g["finalTree"],"Installed source differs")
        entry=Path(p["entryRealpath"]).resolve(strict=True)
        require(entry==(E/"consumer/node_modules/@bdp/server/dist/transaction-installed-process.js").resolve(strict=True) and entry.is_relative_to((E/"consumer").resolve()),"Wrong fresh entry")
        require(sha(entry.read_bytes())==p["entrySha256"],"Entry changed")
        require(sha((E/"consumer/pnpm-lock.yaml").read_bytes())==g["consumerLockSha256"]==p["consumerLockSha256"],"Consumer lock differs")
        if spec["phase"]=="installed":spec["validatedBdpOverrides"]={"BDP_TRANSACTIONAL_INSTALLED_ENTRY":str(entry),"BDP_TRANSACTIONAL_SOURCE_COMMIT":g["finalHead"]}
    g={**g,"sourceSha256":source,"sourceDiffSha256":source_patch_hash(),"preparationSha256":sha((P/"supervise.py").read_bytes())}
    return g,spec
