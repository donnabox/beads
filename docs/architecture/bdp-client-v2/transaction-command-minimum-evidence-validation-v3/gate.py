"""Fixed default-closed command-minimum evidence gate. Root releases stages explicitly."""
from pathlib import Path
import hashlib,json,re,subprocess
P=Path(__file__).resolve().parent
W=Path("/Users/dbox/repos/wt/vickie/vickie-command-minimum-evidence")
E=P/"run"
PARENT="453b6dab386746bb03aac2c0a293572b8a71c969"
sha=lambda b:hashlib.sha256(b).hexdigest()
def require(ok,message):
    if not ok:raise RuntimeError(message)
def git(*args):return subprocess.check_output(["git",*args],cwd=W)
def document(g,key):
    d=g[key];p=Path(d["path"])
    require(p.is_absolute() and sha(p.read_bytes())==d["sha256"],"Changed bound document: "+key)
    return json.loads(p.read_text())
def source_hashes():
    return {n:sha((W/n).read_bytes()) for n in json.loads((P/"commands.json").read_text())["ownedPaths"]}
def completed(name):
    x=json.loads((E/(name+".exit.json")).read_text())
    require(x["exit"]==x["supervisorExit"]==0 and x["sourceUnchanged"],"Prior command failed: "+name)
    for key in ("timeout","forcedCleanup","errors","remainingAfterCleanup","sampledPidsStillPresent","runtimeDirectoriesStillPresent"):
        require(not x[key],"Unclean prior command: "+name+" "+key)
    require(x["sourceSha256After"]==source_hashes(),"Prior command used different source: "+name)
    return x
def verify_derivative(*_):raise RuntimeError("External derivative checkouts are not authorized")
def build_outputs():
    paths=[W/"LICENSE"]
    for name in ("protocol","server"):
        root=W/"packages"/name;paths.append(root/"package.json")
        for sub in (("dist","schemas") if name=="protocol" else ("dist",)):
            paths.extend(p for p in (root/sub).rglob("*") if p.is_file())
    require(all(not p.is_symlink() for p in paths),"Symlink in package outputs")
    return {p.relative_to(W).as_posix():sha(p.read_bytes()) for p in sorted(paths)}
def verify_build(g):
    completed("build")
    start=json.loads((E/"build.started.json").read_text())
    require(start["sourceSha256"]==g["sourceSha256"],"Fresh build source differs")
    require(start["argv"]==[g["node"],g["pnpm"],"build"],"Wrong fresh build command")
    inventory=document(g,"buildOutputInventory")
    require(inventory["sourceSha256"]==g["sourceSha256"] and inventory["files"]==build_outputs(),"Reviewed build output inventory differs")
def load_gate(name):
    require(__debug__,"Python assertions must remain enabled")
    g=json.loads((P/"execution-gate.json").read_text()) # absent until root release
    require(g.get("released") is True and g.get("authorizationKind")=="command-minimum-thirty-source-two-helpers-three-installed","Gate closed")
    for key in ("rootAcceptance","growthAcceptance","sourceManifest","councilAcceptance"):
        d=g[key];p=Path(d["path"])
        require(p.is_absolute() and sha(p.read_bytes())==d["sha256"],"Changed acceptance: "+key)
    names={"gate.py","supervise.py","packages.py","build-inventory.py","adjudicate.py","helper-review.py","commands.json"}
    require(set(g["preparationSha256"])==names,"Preparation binding incomplete")
    for n,h in g["preparationSha256"].items():require(sha((P/n).read_bytes())==h,"Preparation changed: "+n)
    config=json.loads((P/"commands.json").read_text())
    require(sha(Path(config["plan"]["path"]).read_bytes())==config["plan"]["sha256"],"Accepted plan changed")
    require(sha(Path(config["repairPlan"]["path"]).read_bytes())==config["repairPlan"]["sha256"],"Accepted header repair plan changed")
    require(re.fullmatch("[0-9a-f]{40}",g.get("sourceHead") or "") is not None,"Source HEAD not pinned")
    require(git("rev-parse","HEAD").decode().strip()==g["sourceHead"],"Source HEAD differs")
    require(git("rev-parse","HEAD^{tree}").decode().strip()==g["sourceTree"],"Source tree differs")
    require(g["sourceHead"]==PARENT or git("show","-s","--format=%P","HEAD").decode().strip()==PARENT,"Unexpected ancestry")
    require(source_hashes()==g["sourceSha256"] and len(g["sourceSha256"])==12,"Twelve source hashes differ")
    require(sha(git("diff","--binary","HEAD"))==g["sourceDiffSha256"],"Working diff differs")
    require(set(git("diff","--name-only",PARENT).decode().splitlines())==set(config["ownedPaths"]),"Twelve-file footprint differs")
    require(not git("ls-files","--others","--exclude-standard").strip(),"Untracked product files")
    require(sha((W/"pnpm-lock.yaml").read_bytes())==g["sourceLockSha256"],"Source lock differs")
    require(g["node"]=="/Users/dbox/.nvm/versions/node/v24.16.0/bin/node" and g["pnpm"]=="/Users/dbox/.cache/node/corepack/v1/pnpm/11.20.0/bin/pnpm.cjs","Tool paths differ")
    require(name in g["approvedCommandIds"],"Command not released")
    spec=config["commands"][name]
    require(spec["supervisorSeconds"]==120 and "control" not in spec,"Wrong command boundary")
    for dep in spec["requires"]:completed(dep["id"])
    if name=="static-cases":
        start=json.loads((E/"build.started.json").read_text())
        require(start["sourceSha256"]==g["sourceSha256"] and start["sourceDiffSha256"]==g["sourceDiffSha256"],"Matching fresh build required before module import")
    if spec["phase"] in ("package","installed","review"):
        require(g["sourceHead"]!=PARENT and g.get("finalHead")==g["sourceHead"] and g.get("finalTree")==g["sourceTree"],"Final ordinary commit required")
        require(not git("status","--porcelain").strip(),"Final source must be clean")
        verify_build(g)
    if spec["phase"] in ("installed","review"):
        provenance=document(g,"installedProvenance")
        require(provenance["sourceCommit"]==g["finalHead"] and provenance["sourceTree"]==g["finalTree"],"Installed source differs")
        entry=Path(provenance["entryRealpath"]).resolve(strict=True)
        require(entry== (E/"consumer/node_modules/@bdp/server/dist/transaction-installed-process.js").resolve(strict=True) and entry.is_relative_to((E/"consumer").resolve()),"Wrong installed entry")
        require(sha(entry.read_bytes())==provenance["entrySha256"],"Installed entry changed")
        require(sha((E/"consumer/pnpm-lock.yaml").read_bytes())==g["consumerLockSha256"]==provenance["consumerLockSha256"],"Consumer lock changed")
        if spec["phase"]=="installed":
            overrides={"BDP_TRANSACTIONAL_INSTALLED_ENTRY":str(entry),"BDP_TRANSACTIONAL_SOURCE_COMMIT":g["finalHead"]}
            if name=="installed-positive":overrides["BDP_TRANSACTIONAL_OBSERVATION_OUTPUT"]=str(E/(name+".observation.json"))
            elif name in ("installed-batch-control","installed-reader-control"):overrides["BDP_TRANSACTIONAL_MINIMUM_CONTROL_OUTPUT"]=str(E/(name+".observation"))
            else:require(name in ("installed-compatible-minimum","installed-snapshot-wait"),"Unknown installed helper")
            spec["validatedBdpOverrides"]=overrides
    return g,spec
