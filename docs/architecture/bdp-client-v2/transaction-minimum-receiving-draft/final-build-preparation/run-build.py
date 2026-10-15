"""Closed single clean-candidate build. No setup, tests, pack/install or release.
Adaptation reuses the reviewed v3 process-group supervisor, not product code.
"""
import hashlib, importlib.util, json, re, subprocess, sys
from pathlib import Path
P=Path(__file__).resolve().parent
D=P.parent
L=D.parent
W=Path("/Users/dbox/repos/wt/vickie/vickie-minimum-receiving")
BASE=D/"execution-preparation-v3/run-source.py"
BASE_SHA='468e6f32d9f7c2aa70ff1e769ae313d61782b83ee43443baed268b91d0b2c2b3'
COMMANDS=P/"commands.json"
COMMANDS_SHA='004e61e0e390e52d509f41870994fc0d86f0c181ad9c1ad4250f72660eee807f'
SESSION="2026-10-08T20:26:35Z-81e5"
E=P/"run"
def require(ok,message):
    if not ok:raise RuntimeError(message)
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def read(p):return json.loads(p.read_text())
def write(p,v):
    with p.open("x") as f:json.dump(v,f,indent=2);f.write("\n")
def git(*args):return subprocess.check_output(["git",*args],cwd=W,timeout=10)
def document(g,key):
    item=g[key];p=Path(item["path"])
    require(p.is_relative_to(L) and sha(p)==item["sha256"],"Bound document changed: "+key)
    return read(p)
def input_files():
    names=git("ls-files").decode().splitlines()
    require(all((W/n).is_file() and not (W/n).is_symlink() for n in names),"Unexpected tracked input type")
    return {n:sha(W/n) for n in names}
def build_outputs():
    paths=[W/"LICENSE"]
    for name in ("protocol","server"):
        root=W/"packages"/name;paths.append(root/"package.json")
        for sub in (("dist","schemas") if name=="protocol" else ("dist",)):
            paths.extend(p for p in (root/sub).rglob("*") if p.is_file())
    require(all(not p.is_symlink() for p in paths),"Package output symlink")
    return {p.relative_to(W).as_posix():sha(p) for p in sorted(paths)}
def emitted_files():
    paths=[p for root in (W/"packages").iterdir() if root.is_dir() for p in (root/"dist").rglob("*") if p.is_file()]
    require(all(not p.is_symlink() for p in paths),"Emitted output symlink")
    return {p.relative_to(W).as_posix():sha(p) for p in sorted(paths)}
def inventory(source,g):
    return {"buildHead":g["expectedHead"],"buildTree":g["expectedTree"],"buildWorkspaceDirty":False,"sourceSha256":source,"inputFiles":input_files(),"packageFiles":build_outputs(),"emittedFiles":emitted_files()}
def main():
    require(sys.argv[1:]==["build"],"Only one build command")
    require(sha(BASE)==BASE_SHA and sha(COMMANDS)==COMMANDS_SHA,"Reviewed preparation changed")
    g=read(P/"execution-gate.json")
    require(g["status"]=="released" and g["session"]==SESSION and g["authorizedStages"]==["build"],"Root build gate closed")
    require(g["installedGate"]==g["controlGate"]=="closed","Broader release prohibited")
    require(g["supervisorSha256"]==sha(Path(__file__)) and g["baseSupervisorSha256"]==BASE_SHA and g["commandsSha256"]==COMMANDS_SHA,"Gate/preparation mismatch")
    acceptance=document(g,"rootAcceptance");require(acceptance.get("accepted") is True,"Root acceptance absent")
    b=document(g,"sourceBinding")
    require(len(b["files"])==11 and set(b["files"])==set(read(D/"receiving-test-repair/source-binding.json")["files"]),"Eleven owned paths required")
    require(b["workspace"]==str(W),"Wrong source workspace")
    for key in ("expectedHead","expectedTree","featureBase"):
        require(re.fullmatch(r"[0-9a-f]{40}",g.get(key) or "") is not None,"Actual final identity pending: "+key)
    require(g["expectedHead"]!=g["featureBase"],"Final candidate pending")
    def source_hashes():
        require(sha(Path(g["sourceBinding"]["path"]))==g["sourceBinding"]["sha256"],"Source manifest changed")
        require(git("rev-parse","HEAD").decode().strip()==g["expectedHead"] and git("rev-parse","HEAD^{tree}").decode().strip()==g["expectedTree"],"Final candidate changed")
        require(git("show","-s","--format=%P","HEAD").decode().strip().split()==g["expectedParents"],"Final ancestry changed")
        require(not git("status","--porcelain").strip(),"Clean candidate required")
        git("merge-base","--is-ancestor",g["featureBase"],g["expectedHead"])
        require(set(git("diff","--name-only",g["featureBase"],"HEAD").decode().splitlines())==set(b["files"]),"Final feature footprint differs")
        actual={n:sha(W/n) for n in b["files"]}
        require(actual=={n:v["sha256"] for n,v in b["files"].items()},"Frozen final source differs")
        return actual
    source=source_hashes()
    require(not E.exists(),"Prior build directory exists; no retry")
    E.mkdir()
    before=inventory(source,g);write(E/"build-inventory-before.json",before)
    spec=importlib.util.spec_from_file_location("reviewed_clean_build_supervisor",BASE)
    module=importlib.util.module_from_spec(spec);spec.loader.exec_module(module)
    module.E=E;module.COMMANDS=COMMANDS;module.COMMANDS_SHA=COMMANDS_SHA
    module.source_hashes=source_hashes
    module.source_patch_hash=lambda:hashlib.sha256(git("diff","--binary","HEAD")).hexdigest()
    def load(name):
        require(name=="build" and sha(COMMANDS)==COMMANDS_SHA and sha(module.GUARD)==module.GUARD_SHA,"Build command/session guard changed")
        require(read(P/"execution-gate.json")==g,"Gate changed before dispatch")
        actual=source_hashes();require(input_files()==before["inputFiles"],"Inputs changed before dispatch")
        return {**g,"sourceSha256":actual,"sourceDiffSha256":module.source_patch_hash(),"preparationSha256":sha(Path(__file__))},dict(read(COMMANDS)["commands"][0])
    module.load_gate=load
    code=1
    try:
        code=module.main()
    finally:
        # Retain post-build bytes even on failure; installed gate independently requires success.
        after=inventory({n:sha(W/n) for n in b["files"]},g)
        after["inputsUnchanged"]=after["inputFiles"]==before["inputFiles"]
        after["sourceBinding"]=g["sourceBinding"]
        after["buildWorkspaceDirty"]=bool(git("status","--porcelain").strip())
        after["actualHeadAfter"]=git("rev-parse","HEAD").decode().strip()
        for key,name in [("buildStart","build.started.json"),("buildExit","build.exit.json"),("buildLog","build.log")]:
            p=E/name
            after[key]={"path":str(p),"sha256":sha(p)} if p.exists() else None
        after["supervisorExit"]=code
        write(E/"build-output-inventory.json",after)
    require(after["inputsUnchanged"],"Build changed tracked inputs; retained outcome is not acceptable")
    source_hashes()
    print(json.dumps({"supervisorExit":code,"inventory":str(E/"build-output-inventory.json"),"inventorySha256":sha(E/"build-output-inventory.json"),"inputCount":len(after["inputFiles"]),"packageCount":len(after["packageFiles"]),"emittedCount":len(after["emittedFiles"])}),flush=True)
    return code
if __name__=="__main__":sys.exit(main())
