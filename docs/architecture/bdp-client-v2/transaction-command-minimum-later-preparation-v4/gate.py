"""Default-closed concrete control/package gate; no release is supplied."""
from pathlib import Path
import hashlib, json, re, subprocess
P=Path(__file__).resolve().parent
W=Path("/Users/dbox/repos/wt/vickie/vickie-command-minimum")
E=P/"run"
DRAFT=P.parent/"transaction-command-minimum-draft"
PARENT="122800511ee6306e4105e1f73bf3658f35c90a14"
sha=lambda data:hashlib.sha256(data).hexdigest()
def require(ok,message):
    if not ok: raise RuntimeError(message)
def git(*args):return subprocess.check_output(["git",*args],cwd=W)
def binding():
    p=P/"prospective-source-binding.json"
    require(sha(p.read_bytes())=="455372e6767b204095b3884c0c25cc3baa182ec9368b668d5e336c9816cd4dad","Prospective binding changed")
    return json.loads(p.read_text())
def source_hashes():
    return {f["path"]:sha((W/f["path"]).read_bytes()) for f in binding()["files"]}
def document(g,key):
    obj=g[key];p=Path(obj["path"])
    require(p.is_absolute() and sha(p.read_bytes())==obj["sha256"],key+" changed")
def completed(name,expected=0):
    p=E/(name+".exit.json");x=json.loads(p.read_text())
    require(x["exit"]==expected and x["supervisorExit"]==expected,"Wrong prior exit: "+name)
    for key in ("timeout","forcedCleanup","errors","remainingAfterCleanup","sampledPidsStillPresent","runtimeDirectoriesStillPresent"):
        require(not x[key],"Unclean prior result: "+name+" "+key)
    require(x["sourceUnchanged"],"Prior source changed")
    return x
def validate_actual_manifest(g):
    p=P/"prospective-source-binding.json"
    require(sha(p.read_bytes())=="455372e6767b204095b3884c0c25cc3baa182ec9368b668d5e336c9816cd4dad","Prospective binding changed")
    expected={f["path"]:f["sha256"] for f in json.loads(p.read_text())["files"]}
    record=g["actualSourceManifest"]
    actual=Path(record["path"])
    require(actual.is_absolute() and actual.resolve()!=p.resolve(),"Actual source manifest required")
    require(sha(actual.read_bytes())==record["sha256"],"Actual source manifest changed")
    manifest=json.loads(actual.read_text())
    require(manifest["workspace"]==str(W),"Wrong actual source workspace")
    actual_files=manifest["files"]
    require(len(actual_files)==7 and {f["path"]:f["sha256"] for f in actual_files}==expected,"Actual manifest differs from prospective candidate")
    require(g["sourceSha256"]==expected,"Gate differs from prospective candidate")
    require(source_hashes()==expected,"Candidate has not been applied exactly")
    return expected

def load_release(phase):
    g=json.loads((P/"execution-gate.json").read_text()) # absent until root releases
    require(g.get("released") is True and g.get("authorizationKind")=="measured-later-execution","Gate closed")
    document(g,"rootAcceptance");document(g,"sourceChecksAcceptance")
    names={"gate.py","supervise.py","controls.py","packages.py","adjudicate.py","commands.json"}
    require(set(g["preparationSha256"])==names,"Incomplete preparation binding")
    for name,h in g["preparationSha256"].items():require(sha((P/name).read_bytes())==h,"Preparation changed: "+name)
    require(git("rev-parse","HEAD").decode().strip()==g["sourceHead"],"Source HEAD changed")
    require(git("rev-parse","HEAD^{tree}").decode().strip()==g["sourceTree"],"Source tree changed")
    expected=validate_actual_manifest(g)
    require(g["sourceSha256"]==expected and source_hashes()==expected,"Reviewed seven-file bytes changed")
    require(sha(git("diff","--binary","HEAD"))==g["sourceDiffSha256"],"Source diff changed")
    require(set(git("diff","--name-only","HEAD").decode().splitlines())<=set(expected),"Unowned change")
    require(not git("ls-files","--others","--exclude-standard").strip(),"Untracked source")
    if phase=="control":
        require(g["sourceHead"]==PARENT or git("show","-s","--format=%P","HEAD").decode().strip()==PARENT,"Not ordinary source")
    else:
        require(re.fullmatch("[0-9a-f]{40}",g.get("finalHead") or "") is not None,"No final commit")
        require(g["finalHead"]==g["sourceHead"] and g["finalTree"]==g["sourceTree"],"Final binding mismatch")
        require(not git("status","--porcelain").strip(),"Final source dirty")
        require(git("show","-s","--format=%P","HEAD").decode().strip()==PARENT,"Final must be ordinary child")
        require(set(git("diff","--name-only",PARENT,"HEAD").decode().splitlines())==set(expected),"Final footprint changed")
        document(g,"controlsAcceptance")
        b=g["buildEvidence"]
        start=json.loads(Path(b["startedPath"]).read_text());end=json.loads(Path(b["exitPath"]).read_text())
        require(sha(Path(b["startedPath"]).read_bytes())==b["startedSha256"] and sha(Path(b["exitPath"]).read_bytes())==b["exitSha256"],"Build evidence changed")
        require(start["sourceSha256"]==expected and end["exit"]==0 and end["supervisorExit"]==0 and end["sourceUnchanged"],"Build not bound")
        require(not end["timeout"] and not end["forcedCleanup"],"Build not clean")
        require(sha((W/"pnpm-lock.yaml").read_bytes())==g["sourceLockSha256"],"Source lock changed")
    return g
def verify_derivative(control,g):
    snapshot=E/(control+".snapshot.json")
    require(sha(snapshot.read_bytes())==g["controls"][control]["snapshotSha256"],"Derivative snapshot not accepted")
    record=json.loads(snapshot.read_text())
    root=E/control
    require(record["sourceSha256"]==g["sourceSha256"],"Derivative input changed")
    require(record["patchSha256"]==g["controls"][control]["patchSha256"],"Derivative patch changed")
    for rel,info in record["files"].items():
        p=root/rel
        require(p.is_file() and not p.is_symlink() and sha(p.read_bytes())==info["sha256"],"Derivative changed: "+rel)
        require(bool(p.stat().st_mode&0o111)==info["executable"],"Derivative mode changed")
    require(record["changedPaths"]==["packages/server/src/transaction-http.ts"],"Derivative scope changed")
    require(sha((root/"packages/server/src/transaction-http.ts").read_bytes())==g["controls"][control]["afterSha256"],"Derivative final hash changed")
def load_gate(name):
    spec=json.loads((P/"commands.json").read_text())["commands"][name]
    g=load_release(spec["phase"])
    require(name in g["approvedCommandIds"],"Command not released")
    require(spec["supervisorSeconds"]==120,"Budget changed")
    for dep in spec.get("requires",[]):completed(dep["id"],dep.get("exit",0))
    control=spec.get("control")
    if control and not name.endswith("-prepare"):verify_derivative(control,g)
    if name=="fresh-tx":
        document(g,"installedProvenance")
        provenance=json.loads(Path(g["installedProvenance"]["path"]).read_text())
        require(provenance["sourceCommit"]==g["finalHead"] and provenance["sourceTree"]==g["finalTree"],"Wrong installed source")
        entry=Path(provenance["entryRealpath"]).resolve(strict=True)
        require(entry.is_relative_to((E/"consumer").resolve()) and sha(entry.read_bytes())==provenance["entrySha256"],"Wrong installed entry")
        spec["validatedBdpOverrides"]={"BDP_TRANSACTIONAL_INSTALLED_ENTRY":str(entry),"BDP_TRANSACTIONAL_SOURCE_COMMIT":g["finalHead"]}
    return g,spec
