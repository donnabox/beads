"""Closed clean-commit/package gate. Does not release itself or import product."""
from pathlib import Path
import hashlib,json,re,subprocess
P=Path(__file__).resolve().parent
D=P.parent
W=Path("/Users/dbox/repos/wt/vickie/vickie-sequence-member-evidence")
E=P/"run"
L=D.parent
GUARD=L/"guarded-action.py"
PARENT="0d63cf59d447eca2e7ff8b2e0e6a5f31513b9a3a"
BINDING=D/"lint-repair/source-binding.json"
INSTALLED_TEST="apps/bdptest/src/transactional-depth-key-order-installed.test.ts"
INSTALLED_TEST_SHA="48c6a84c84e470c7ce79f4d978bb09e2099daa5df4b4a6ef8629e1b7bf1b1700"
BINDING_SHA="5fbe5c9b0c57f2c253fb16d114675ccb4929a9e4e18b21a1efbf1ca18385d778"
def binding_sha():return BINDING_SHA
COMMANDS=P/"commands.json"
COMMANDS_SHA="ca35819eae510e11344afca5fa844daed33ca474965d1417b59d9f8c69ab6352"
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
    require(sha(BINDING.read_bytes())==binding_sha(),"Source binding changed")
    expected={k:v["sha256"] for k,v in json.loads(BINDING.read_text())["files"].items()}
    require(set(expected)==set(json.loads(COMMANDS.read_text())["ownedPaths"]) and len(expected)==8,"Wrong combined source scope")
    require(expected[INSTALLED_TEST]==INSTALLED_TEST_SHA,"Installed witness changed")
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
    return {p.relative_to(W).as_posix():sha(p.read_bytes()) for parent in ("packages","apps") for root in (W/parent).iterdir() if root.is_dir() for p in (root/"dist").rglob("*") if p.is_file()}
def verify_build(g):
    inventory=document(g,"buildOutputInventory")
    require(Path(g["buildOutputInventory"]["path"])==P/"build-attribution-bridge.json","Wrong reuse inventory")
    require(inventory["status"]=="ACTUAL_SUCCESSFUL_DIRTY_SOURCE_BUILD_PRESERVED_CANDIDATE_PENDING","Historical attribution changed")
    for key in ("sourceBinding","buildStart","buildExit","buildLog","controlBuildManifest","controlCopyManifest","rootControlAcceptance","controlMutation","controlMutationPatch","controlPacket"):
        d=inventory[key];require(sha(Path(d["path"]).read_bytes())==d["sha256"],"Historical evidence changed: "+key)
    require(inventory["sourceBinding"]["path"]==str(BINDING) and inventory["sourceBinding"]["sha256"]==BINDING_SHA,"Wrong historical source")
    start=json.loads(Path(inventory["buildStart"]["path"]).read_text())
    result=json.loads(Path(inventory["buildExit"]["path"]).read_text())
    require(start["head"]==PARENT and start["workspaceDirty"] is True and start["sourceSha256"]==inventory["sourceSha256"]==source_hashes(),"Build source attribution differs")
    require(start["argv"]==[g["node"],g["pnpm"],"build"] and start["cwd"]==str(W),"Wrong historical build")
    require(result["exit"]==result["supervisorExit"]==0 and result["sourceUnchanged"] and result["sourceSha256After"]==inventory["sourceSha256"],"Historical build failed/changed")
    for key in ("timeout","forcedCleanup","errors","remainingAfterCleanup","sampledPidsStillPresent","runtimeDirectoriesStillPresent"):require(not result[key],"Unclean historical build")
    require(inventory["inputFiles"]==input_files(),"Candidate tracked build inputs differ")
    require(inventory["packageFiles"]==build_outputs() and len(inventory["packageFiles"])==312,"Package outputs changed")
    require(inventory["emittedFiles"]==emitted_files() and len(inventory["emittedFiles"])==464,"Emitted outputs changed")
    bm=json.loads(Path(inventory["controlBuildManifest"]["path"]).read_text());cm=json.loads(Path(inventory["controlCopyManifest"]["path"]).read_text())
    require(bm["files"]==inventory["emittedFiles"] and bm["buildExit"]["sha256"]==inventory["buildExit"]["sha256"],"Build manifest mismatch")
    require(cm["sourceBindingSha256"]==BINDING_SHA and cm["buildManifestSha256"]==inventory["controlBuildManifest"]["sha256"],"Copy manifest basis mismatch")
    allfiles={**inventory["inputFiles"],**inventory["emittedFiles"]}
    require(set(allfiles)==set(cm["files"]) and all(cm["files"][n]["positiveSha256"]==h for n,h in allfiles.items()),"Post-build positive/source equality differs")
    mutation=json.loads(Path(inventory["controlMutation"]["path"]).read_text())
    packet=json.loads(Path(inventory["controlPacket"]["path"]).read_text())
    accepted=json.loads(Path(inventory["rootControlAcceptance"]["path"]).read_text())
    require(accepted["status"]=="accepted-single-early-depth-control" and accepted["sourceBindingSha256"]==BINDING_SHA and accepted["graphVerifiedAgain"]["packetSha256"]==inventory["controlPacket"]["sha256"],"Control acceptance binding differs")
    for filename,key in (("copy-manifest.json","controlCopyManifest"),("build-manifest.json","controlBuildManifest"),("mutation.json","controlMutation"),("mutation.patch","controlMutationPatch")):
        require(packet["files"][filename]==inventory[key]["sha256"],"Control packet differs: "+filename)
    target="packages/server/dist/transaction-store.js"
    require(mutation["path"]==target and mutation["patchSha256"]==inventory["controlMutationPatch"]["sha256"],"Wrong negative mutation")
    different={n for n,v in cm["files"].items() if v["copySha256"]!=v["positiveSha256"]}
    require(different=={target},"Unexpected negative copy divergence")
    positive=(W/target).read_bytes();before=mutation["exactBeforeText"].encode();after=mutation["exactAfterText"].encode()
    require(sha(positive)==mutation["positiveSha256"]==allfiles[target] and positive.count(before)==mutation["occurrencesRequired"]==1,"Mutation positive basis differs")
    require(sha(positive.replace(before,after,1))==mutation["derivativeSha256"]==cm["files"][target]["copySha256"],"Negative copy is not exact accepted mutation")
def load_gate(name):
    require(__debug__,"Python assertions required")
    g=json.loads((P/"execution-gate.json").read_text())
    require(g.get("released") is True and g["session"]=="2026-10-08T20:26:35Z-81e5","Root gate closed")
    for key in ("rootAcceptance","councilAcceptance","controlAcceptance","finalSource"):
        document(g,key)
    names={"gate.py","supervise.py","packages.py","review-installed.py","commands.json"}
    require(set(g["preparationSha256"])==names,"Incomplete preparation pins")
    for n,h in g["preparationSha256"].items():require(sha((P/n).read_bytes())==h,"Preparation changed: "+n)
    require(sha(COMMANDS.read_bytes())==COMMANDS_SHA and sha(GUARD.read_bytes())==GUARD_SHA,"Commands/guard changed")
    cfg=json.loads(COMMANDS.read_text());order=cfg["executionOrder"]
    require(name in order and name in g["approvedCommandIds"],"Stage not released")
    require(g["approvedCommandIds"]==order[:len(g["approvedCommandIds"])],"Release must be sequential prefix")
    require(re.fullmatch("[0-9a-f]{40}",g.get("finalHead") or "") is not None and g["finalHead"]!=PARENT,"Final commit pending")
    require(git("rev-parse","HEAD").decode().strip()==g["finalHead"] and git("rev-parse","HEAD^{tree}").decode().strip()==g["finalTree"],"Final identity changed")
    require(g["finalBase"]==PARENT,"Wrong accepted authoring base")
    require(git("show","-s","--format=%P","HEAD").decode().strip().split()==g["finalParents"] and g["finalParents"]==[PARENT],"Final ordered parents differ")
    git("merge-base","--is-ancestor",PARENT,g["finalBase"])
    git("merge-base","--is-ancestor",g["finalBase"],g["finalHead"])
    final=document(g,"finalSource")
    require(final["head"]==g["finalHead"] and final["tree"]==g["finalTree"] and final["base"]==g["finalBase"] and final["parents"]==g["finalParents"] and final["sourceBindingSha256"]==binding_sha(),"Final source document differs")
    source=source_hashes()
    require(set(git("diff","--name-only",g["finalBase"],"HEAD").decode().splitlines())==set(source),"Eight-file scope differs")
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
