"""Closed clean-commit/package gate. Does not release itself or import product."""
from pathlib import Path
import hashlib,json,re,subprocess
P=Path(__file__).resolve().parent
D=P.parent
W=Path("/Users/dbox/repos/wt/vickie/vickie-update-cold-directory")
E=P/"run"
L=D.parent
GUARD=L/"guarded-action.py"
AUTHORING_PARENT="67a2842e3f1918e70978f745587cf11eaf0df5a8"
def source_binding():
    g=json.loads((P/"execution-gate.json").read_text())
    value=document(g,"sourceBinding")
    require(value==json.loads((P/"prepared-source.json").read_text()),"Source differs from reviewed preparation")
    return value
def binding_sha():return json.loads((P/"execution-gate.json").read_text())["sourceBinding"]["sha256"]
COMMANDS=P/"commands.json"
COMMANDS_SHA="42fd87698feb1b0d51914d144837f046e9b5296c4b21c9979baee948d833c6e4"
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
    expected=source_binding()["sourceFiles"]
    require(set(expected)==set(json.loads(COMMANDS.read_text())["ownedPaths"]) and len(expected)==6,"Wrong combined source scope")
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
    return {p.relative_to(W).as_posix():sha(p.read_bytes()) for parent in ("packages",) for root in (W/parent).iterdir() if root.is_dir() for p in (root/"dist").rglob("*") if p.is_file()}
def verify_build(g):
    inventory=document(g,"buildOutputInventory")
    accepted=document(g,"rootBuildAcceptance")
    require(accepted["accepted"] is True and accepted["inventorySha256"]==g["buildOutputInventory"]["sha256"] and accepted["finalHead"]==g["finalHead"],"Build attribution bridge pending")
    for key in ("buildStart","buildExit","buildLog"):
        x=inventory[key];require(sha(Path(x["path"]).read_bytes())==x["sha256"],"Build evidence changed: "+key)
    start=json.loads(Path(inventory["buildStart"]["path"]).read_text());result=json.loads(Path(inventory["buildExit"]["path"]).read_text())
    require(start["sourceSha256"]==inventory["sourceSha256"]==source_hashes(),"Actual build source differs")
    require(result["exit"]==result["supervisorExit"]==0 and result["sourceUnchanged"] and result["sourceSha256After"]==inventory["sourceSha256"],"Build failed/changed")
    require(start["argv"]==[g["node"],g["pnpm"],"build"] and start["cwd"]==str(W),"Wrong build command")
    for key in ("timeout","forcedCleanup","errors","remainingAfterCleanup","sampledPidsStillPresent","runtimeDirectoriesStillPresent"):require(not result[key],"Unclean build")
    require(start["buildInputFiles"]==inventory["inputFiles"]==result["buildInventorySnapshot"]["inputFiles"]==input_files(),"Contemporaneous tracked build inputs differ")
    require(start["head"]==g["finalHead"] and start["workspaceDirty"] is False,"Fresh clean candidate build required")
    require(inventory["packageFiles"]==result["buildInventorySnapshot"]["packageFiles"] and inventory["emittedFiles"]==result["buildInventorySnapshot"]["emittedFiles"],"Build output record differs")
    require(inventory["packageFiles"]==build_outputs(),"Package output bytes differ")
    require(inventory["emittedFiles"]==emitted_files(),"Emitted graph bytes differ")
    require(start["head"]==inventory["buildHead"] and start["workspaceDirty"]==inventory["buildWorkspaceDirty"],"Historical build chronology changed")
def load_gate(name):
    require(__debug__,"Python assertions required")
    g=json.loads((P/"execution-gate.json").read_text())
    require(g.get("released") is True and g["session"]=="2026-10-08T20:26:35Z-81e5","Root gate closed")
    for key in ("rootAcceptance","councilState","sourceValidationAcceptance","finalSource"):
        document(g,key)
    council=document(g,"councilState")
    require(council.get("status")=="accepted" or (council.get("status")=="pending-actual-claude" and document(g,"rootAcceptance").get("validationOnlyWhileCouncilPending") is True),"Council acceptance or explicit local validation-only release absent")
    require(document(g,"rootAcceptance").get("accepted") is True,"Root acceptance absent")
    require(document(g,"sourceValidationAcceptance").get("accepted") is True,"Source validation acceptance absent")
    names={"gate.py","supervise.py","packages.py","commands.json","prepared-source.json","report-contract.json","selection.json"}
    require(set(g["preparationSha256"])==names,"Incomplete preparation pins")
    for n,h in g["preparationSha256"].items():require(sha((P/n).read_bytes())==h,"Preparation changed: "+n)
    require(sha(COMMANDS.read_bytes())==COMMANDS_SHA and sha(GUARD.read_bytes())==GUARD_SHA,"Commands/guard changed")
    cfg=json.loads(COMMANDS.read_text());order=cfg["executionOrder"]
    require(name in order and name in g["approvedCommandIds"],"Stage not released")
    require(g["approvedCommandIds"]==order[:len(g["approvedCommandIds"])],"Release must be sequential prefix")
    require(re.fullmatch("[0-9a-f]{40}",g.get("finalHead") or "") is not None and g["finalHead"]!=AUTHORING_PARENT,"Final commit pending")
    require(git("rev-parse","HEAD").decode().strip()==g["finalHead"] and git("rev-parse","HEAD^{tree}").decode().strip()==g["finalTree"],"Final identity changed")
    integration=document(g,"integrationAcceptance")
    require(integration.get("accepted") is True and integration["authoringParent"]==AUTHORING_PARENT and integration["finalBase"]==g["finalBase"] and integration["sourceBindingSha256"]==binding_sha(),"Integrated parent/source not accepted")
    require(integration["finalHead"]==g["finalHead"] and integration["finalTree"]==g["finalTree"] and integration["finalParents"]==g["finalParents"],"Actual integration not accepted")
    require(git("show","-s","--format=%P","HEAD").decode().strip().split()==g["finalParents"],"Actual ordinary parents differ")
    require(len(g["finalParents"]) in (1,2) and g["finalBase"] in g["finalParents"],"Final base must be ordinary parent")
    git("merge-base","--is-ancestor",AUTHORING_PARENT,g["finalBase"])
    git("merge-base","--is-ancestor",g["finalBase"],g["finalHead"])
    final=document(g,"finalSource")
    require(final["head"]==g["finalHead"] and final["tree"]==g["finalTree"] and final["parents"]==[g["finalBase"]] and final["parents"]==g["finalParents"] and final["sourceBindingSha256"]==binding_sha(),"Final source document differs")
    source=source_hashes()
    require(final["sourceFiles"]==source,"Final six hashes differ")
    require(sha(git("diff","--binary",g["finalBase"],"HEAD"))==final["diffSha256"]==source_binding()["patchSha256"],"Final feature patch differs")
    require(set(git("diff","--name-only",g["finalBase"],"HEAD").decode().splitlines())==set(source),"Six-file scope differs")
    require(sha((W/"pnpm-lock.yaml").read_bytes())==g["sourceLockSha256"],"Source lock differs")
    require(g["node"]=="/Users/dbox/.nvm/versions/node/v24.16.0/bin/node" and g["pnpm"]=="/Users/dbox/.cache/node/corepack/v1/pnpm/11.20.0/bin/pnpm.cjs","Tool paths changed")
    if name != "fresh-build": verify_build(g)
    spec=dict(cfg["commands"][name]);require(spec["supervisorSeconds"]==120,"Wrong time bound")
    for prior in order[:order.index(name)]:completed(prior)
    if spec["phase"] in ("installed","review"):
        p=document(g,"installedProvenance")
        require(p["sourceCommit"]==g["finalHead"] and p["sourceTree"]==g["finalTree"],"Installed source differs")
        entry=Path(p["entryRealpath"]).resolve(strict=True)
        require(entry==(E/"consumer/node_modules/@bdp/server/dist/installed-reference-process.js").resolve(strict=True) and entry.is_relative_to((E/"consumer").resolve()),"Wrong fresh entry")
        require(sha(entry.read_bytes())==p["entrySha256"],"Entry changed")
        require(sha((E/"consumer/pnpm-lock.yaml").read_bytes())==g["consumerLockSha256"]==p["consumerLockSha256"],"Consumer lock differs")
        if spec["phase"]=="installed":
            for value in spec["reportOverrides"].values():
                require(Path(value).is_absolute() and Path(value).parent==E and not Path(value).exists(),"Output must be fresh/external/exclusive")
            require(set(spec["reportOverrides"]) <= {"BDP_COLD_DIRECTORY_REPORT_PATH","BDP_CUMULATIVE_REPORT_PATH"},"Unknown report override")
            require("BDP_COLD_DIRECTORY_REPORT_PATH" in spec["reportOverrides"],"Cold report required")
            spec["validatedBdpOverrides"]={"BDP_CUMULATIVE_INSTALLED_ENTRY":str(entry),**spec["reportOverrides"]}


    g={**g,"sourceSha256":source,"sourceDiffSha256":source_patch_hash(),"preparationSha256":sha((P/"supervise.py").read_bytes())}
    return g,spec
