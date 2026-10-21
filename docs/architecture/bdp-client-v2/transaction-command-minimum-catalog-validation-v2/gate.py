"""Default-closed static catalog validation; source pins must be supplied by root."""
from pathlib import Path
import hashlib,json,re,subprocess
P=Path(__file__).resolve().parent
W=Path("/Users/dbox/repos/wt/vickie/vickie-command-minimum-catalog")
E=P/"run"
PARENT="bd2178dd6f4b9f7c2d239624d7edaa3d961c3e26"
sha=lambda b:hashlib.sha256(b).hexdigest()
def require(ok,message):
    if not ok:raise RuntimeError(message)
def git(*args):return subprocess.check_output(["git",*args],cwd=W)
def source_hashes():
    names=json.loads((P/"commands.json").read_text())["ownedPaths"]
    return {name:sha((W/name).read_bytes()) for name in names}
def verify_derivative(*_):raise RuntimeError("No derivatives in this static packet")
def completed(name):
    x=json.loads((E/(name+".exit.json")).read_text())
    require(x["exit"]==x["supervisorExit"]==0 and x["sourceUnchanged"],"Prior command failed: "+name)
    for key in ("timeout","forcedCleanup","errors","remainingAfterCleanup","sampledPidsStillPresent","runtimeDirectoriesStillPresent"):
        require(not x[key],"Unclean prior command: "+name+" "+key)
    return x
def load_gate(name):
    g=json.loads((P/"execution-gate.json").read_text()) # absent until root releases
    require(g.get("released") is True and g.get("authorizationKind")=="catalog12-static-validation","Gate closed")
    for key in ("rootAcceptance","growthAcceptance","sourceManifest"):
        d=g[key];path=Path(d["path"])
        require(path.is_absolute() and sha(path.read_bytes())==d["sha256"],"Changed acceptance/binding: "+key)
    names={"gate.py","supervise.py","commands.json"}
    require(set(g["preparationSha256"])==names,"Preparation binding incomplete")
    for n,h in g["preparationSha256"].items():require(sha((P/n).read_bytes())==h,"Preparation changed")
    config=json.loads((P/"commands.json").read_text())
    require(sha(Path(config["plan"]["path"]).read_bytes())==config["plan"]["sha256"],"Accepted plan changed")
    require(re.fullmatch("[0-9a-f]{40}",g.get("sourceHead") or "") is not None,"Final measured source pin absent")
    require(git("rev-parse","HEAD").decode().strip()==g["sourceHead"],"Source HEAD differs")
    require(git("rev-parse","HEAD^{tree}").decode().strip()==g["sourceTree"],"Source tree differs")
    require(g["sourceHead"]==PARENT or git("show","-s","--format=%P","HEAD").decode().strip()==PARENT,"Unexpected source ancestry")
    require(source_hashes()==g["sourceSha256"] and len(g["sourceSha256"])==11,"Eleven-file source binding differs")
    require(sha(git("diff","--binary","HEAD"))==g["sourceDiffSha256"],"Frozen diff differs")
    require(set(git("diff","--name-only","HEAD").decode().splitlines())<=set(config["ownedPaths"]),"Unowned changes")
    require(not git("ls-files","--others","--exclude-standard").strip(),"Untracked product files")
    require(sha((W/"pnpm-lock.yaml").read_bytes())==g["sourceLockSha256"],"Source lock differs")
    require(name in g["approvedCommandIds"],"Command not released")
    spec=config["commands"][name]
    require(spec["supervisorSeconds"]==120 and "control" not in spec,"Unexpected command boundary")
    for dep in spec["requires"]:completed(dep["id"])
    if name=="static-cases":
        start=json.loads((E/"build.started.json").read_text())
        require(start["sourceSha256"]==g["sourceSha256"] and start["sourceDiffSha256"]==g["sourceDiffSha256"],"Current matching build required before module import")
    return g,spec
