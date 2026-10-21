"""Actual dirty-source build inventory. No product import or package creation."""
from pathlib import Path
import hashlib, json, subprocess
P=Path(__file__).resolve().parent
W=Path('/Users/dbox/repos/wt/vickie/vickie-guarded-batch')
E=P/"run"
sha=lambda b:hashlib.sha256(b).hexdigest()
def require(ok,message):
    if not ok:raise RuntimeError(message)
def git(*args):return subprocess.check_output(["git",*args],cwd=W,timeout=15)
b=json.loads((P.parent/"root-fence-amendment/source-binding.json").read_text())
start=json.loads((E/"build.started.json").read_text())
finish=json.loads((E/"build.exit.json").read_text())
require(finish["exit"]==finish["supervisorExit"]==0 and finish["sourceUnchanged"],"Build did not pass")
names=sorted(set(git("ls-files").decode().splitlines())|set(b["files"]))
inputs={name:sha((W/name).read_bytes()) for name in names}
require(inputs==start["buildInputFiles"],"Inputs changed since actual build start")
require(git("rev-parse","HEAD").decode().strip()==start["head"]==b["head"],"Build HEAD changed")
source={name:sha((W/name).read_bytes()) for name in b["files"]}
require(source=={name:item["sha256"] for name,item in b["files"].items()}==start["sourceSha256"],"Build source changed")
paths=[W/"LICENSE"]
for name in ("protocol","server"):
    root=W/"packages"/name
    paths.append(root/"package.json")
    for sub in (("dist","schemas") if name=="protocol" else ("dist",)):
        require((root/sub).is_dir(),"Build output absent")
        paths.extend(p for p in (root/sub).rglob("*") if p.is_file())
require(all(not p.is_symlink() for p in paths),"Package output symlink")
outputs={p.relative_to(W).as_posix():sha(p.read_bytes()) for p in sorted(paths)}
emitted={p.relative_to(W).as_posix():sha(p.read_bytes()) for root in sorted((W/"packages").iterdir()) if root.is_dir() for p in sorted((root/"dist").rglob("*")) if p.is_file()}
pin=lambda p:{"path":str(p),"sha256":sha(p.read_bytes())}
record={"buildHead":start["head"],"workspaceDirty":start["workspaceDirty"],"sourceSha256":source,"buildStart":pin(E/"build.started.json"),"buildExit":pin(E/"build.exit.json"),"buildLog":pin(E/"build.log"),"inputFiles":inputs,"files":outputs,"emittedFiles":emitted,"sourceAttribution":"Actual dirty caller checkout build. Later clean-candidate packaging requires an explicit byte bridge; no installed claim or child attestation."}
with (E/"build-output-inventory.json").open("x") as f:f.write(json.dumps(record,indent=2)+"\n")
print(json.dumps({"inputFiles":len(inputs),"packageOutputs":len(outputs),"emittedFiles":len(emitted),"inventorySha256":sha((E/"build-output-inventory.json").read_bytes())}))
