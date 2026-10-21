"""Record actual clean integrated candidate build inputs/output, not installed proof."""
from pathlib import Path
import json,hashlib,subprocess
P=Path(__file__).resolve().parent;E=P/"run";D=P.parent
W=Path('/Users/dbox/repos/wt/vickie/vickie-minimum-current-resource')
sha=lambda b:hashlib.sha256(b).hexdigest()
def require(ok,text):
 if not ok: raise RuntimeError(text)
start=json.loads((E/"build.started.json").read_text())
result=json.loads((E/"build.exit.json").read_text())
require(result["supervisorExit"]==result["exit"]==0 and result["sourceUnchanged"],"Build did not complete")
inputs=json.loads((P/"input-binding.json").read_text())
require(inputs==start["buildInputFiles"],"Build input attribution differs")
require({n:sha((W/n).read_bytes()) for n in inputs}==inputs,"Build inputs changed")
paths=[W/"LICENSE"]
for name in ("protocol","server"):
 root=W/"packages"/name;paths.append(root/"package.json")
 for sub in (("dist","schemas") if name=="protocol" else ("dist",)):
  require((root/sub).is_dir(),"Expected built directory absent")
  paths.extend(p for p in (root/sub).rglob("*") if p.is_file())
require(all(not p.is_symlink() for p in paths),"Unexpected package member symlink")
outputs={p.relative_to(W).as_posix():sha(p.read_bytes()) for p in sorted(paths)}
emitted={p.relative_to(W).as_posix():sha(p.read_bytes()) for root in (W/"packages").iterdir() if root.is_dir() for p in (root/"dist").rglob("*") if p.is_file()}
require(bool(outputs) and bool(emitted),"Empty build inventory")
pin=lambda p:{"path":str(p),"sha256":sha(p.read_bytes())}
record={"buildHead":start["head"],"workspaceDirty":start["workspaceDirty"],"sourceSha256":start["sourceSha256"],"inputFiles":inputs,"files":outputs,"emittedFiles":emitted,"buildStart":pin(E/"build.started.json"),"buildExit":pin(E/"build.exit.json"),"buildLog":pin(E/"build.log"),"sourceAttribution":"Actual clean integrated candidate build; archive/member/consumer/installed proof remains separate."}
with (E/"build-output-inventory.json").open("x") as out:out.write(json.dumps(record,indent=2)+"\n")
print(json.dumps({"inputFiles":len(inputs),"packageFiles":len(outputs),"emittedFiles":len(emitted),"sha256":sha((E/"build-output-inventory.json").read_bytes())}))
