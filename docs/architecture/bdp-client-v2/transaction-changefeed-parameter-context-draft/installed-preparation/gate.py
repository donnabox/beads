"""Closed clean-commit/package gate. Does not release itself or import product."""
from pathlib import Path
import hashlib,json,re,subprocess
P=Path(__file__).resolve().parent
D=P.parent
W=Path("/Users/dbox/repos/wt/vickie/vickie-changefeed-parameter-context")
E=P/"run"
L=D.parent
GUARD=L/"guarded-action.py"
PARENT="9a42a4531def99e12fa9a1fba02e70558d643265"
BINDING=D/"source-binding.json"
BINDING_SHA="0abed77764ff7763e5e5a6f9b5a767090423fed01b21ea01b162a17b3b743e70"
COMMANDS=P/"commands.json"
COMMANDS_SHA="415c6c2f1b4696a737356d56e8ab771c814153de625654432a5c592e6413a47f"
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
def input_files():
    return {n:sha((W/n).read_bytes()) for n in git("ls-files").decode().splitlines()}
def emitted_files():
    return {p.relative_to(W).as_posix():sha(p.read_bytes()) for root in (W/"packages").iterdir() if root.is_dir() for p in (root/"dist").rglob("*") if p.is_file()}
def verify_build(g):
    inventory=document(g,"buildOutputInventory")
    require(Path(g["buildOutputInventory"]["path"])==E/"build-output-inventory.json","Wrong fresh inventory path")
    require(inventory["sourceCommit"]==g["finalHead"] and inventory["sourceTree"]==g["finalTree"] and inventory["sourceSha256"]==source_hashes(),"Fresh build identity differs")
    for key in ("buildStart","buildExit","buildLog"):
        x=inventory[key];require(sha(Path(x["path"]).read_bytes())==x["sha256"],"Build evidence changed: "+key)
    require(Path(inventory["buildExit"]["path"])==E/"fresh-build.exit.json","Wrong build outcome")
    result=completed("fresh-build")
    start=json.loads((E/"fresh-build.started.json").read_text())
    require(start["head"]==g["finalHead"] and start["workspaceDirty"] is False and start["sourceSha256"]==source_hashes(),"Build attribution changed")
    require(start["argv"]==[g["node"],g["pnpm"],"build"] and start["cwd"]==str(W),"Wrong build command")
    require(start["buildInputFiles"]==inventory["inputFiles"]==input_files(),"Build inputs changed")
    require(inventory["files"]==build_outputs() and len(inventory["files"])==312,"Build output footprint changed")
    require(inventory["emittedFiles"]==emitted_files(),"Emitted graph differs")
def load_gate(name):
    require(__debug__,"Python assertions required")
    g=json.loads((P/"execution-gate.json").read_text())
    require(g.get("released") is True and g["session"]=="2026-10-08T20:26:35Z-81e5","Root gate closed")
    for key in ("rootAcceptance","councilAcceptance","controlAcceptance","finalSource"):
        document(g,key)
    names={"gate.py","supervise.py","packages.py","review-installed.py","commands.json","inventory.py"}
    require(set(g["preparationSha256"])==names,"Incomplete preparation pins")
    for n,h in g["preparationSha256"].items():require(sha((P/n).read_bytes())==h,"Preparation changed: "+n)
    require(sha(COMMANDS.read_bytes())==COMMANDS_SHA and sha(GUARD.read_bytes())==GUARD_SHA,"Commands/guard changed")
    cfg=json.loads(COMMANDS.read_text());order=cfg["executionOrder"]
    require(name in order and name in g["approvedCommandIds"],"Stage not released")
    require(g["approvedCommandIds"]==order[:len(g["approvedCommandIds"])],"Release must be sequential prefix")
    require(re.fullmatch("[0-9a-f]{40}",g.get("finalHead") or "") is not None and g["finalHead"]!=PARENT,"Final commit pending")
    require(git("rev-parse","HEAD").decode().strip()==g["finalHead"] and git("rev-parse","HEAD^{tree}").decode().strip()==g["finalTree"],"Final identity changed")
    require(g["finalBase"]=="77f4b89c117e46f0e73476f69b4c8b377306452b","Wrong integrated base")
    require(git("show","-s","--format=%P","HEAD").decode().strip().split()==g["finalParents"] and len(g["finalParents"]) in (1,2),"Final ordered parents differ")
    git("merge-base","--is-ancestor",PARENT,g["finalBase"])
    git("merge-base","--is-ancestor",g["finalBase"],g["finalHead"])
    final=document(g,"finalSource")
    require(final["head"]==g["finalHead"] and final["tree"]==g["finalTree"] and final["base"]==g["finalBase"] and final["parents"]==g["finalParents"] and final["sourceBindingSha256"]==BINDING_SHA,"Final source document differs")
    source=source_hashes()
    require(set(git("diff","--name-only",g["finalBase"],"HEAD").decode().splitlines())==set(source),"Six-file scope differs")
    require(sha((W/"pnpm-lock.yaml").read_bytes())==g["sourceLockSha256"],"Source lock differs")
    require(g["node"]=="/Users/dbox/.nvm/versions/node/v24.16.0/bin/node" and g["pnpm"]=="/Users/dbox/.cache/node/corepack/v1/pnpm/11.20.0/bin/pnpm.cjs","Tool paths changed")
    if name not in ("fresh-build","build-inventory"):verify_build(g)
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
