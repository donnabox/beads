"""Data-only actual build inventory; records dirty/clean attribution without relabeling."""
import json
from gate import E,W,load_gate,completed,source_hashes,input_files,build_outputs,emitted_files,sha,require
g,_=load_gate("build-inventory");completed("build")
start=json.loads((E/"build.started.json").read_text())
inputs=input_files();require(inputs==start["buildInputFiles"],"Actual build inputs changed")
pin=lambda p:{"path":str(p),"sha256":sha(p.read_bytes())}
record={"buildHead":start["head"],"workspaceDirty":start["workspaceDirty"],"sourceSha256":source_hashes(),"buildStart":pin(E/"build.started.json"),"buildExit":pin(E/"build.exit.json"),"buildLog":pin(E/"build.log"),"inputFiles":inputs,"files":build_outputs(),"emittedFiles":emitted_files(),"sourceAttribution":"Actual caller checkout build; package-ready clean source requires later exact bridge if identity/status differs. Not child attestation."}
with (E/"build-output-inventory.json").open("x") as f:f.write(json.dumps(record,indent=2)+"\n")
print(json.dumps({"inputFiles":len(inputs),"packageOutputs":len(record["files"]),"emittedFiles":len(record["emittedFiles"]),"inventorySha256":sha((E/"build-output-inventory.json").read_bytes())}))
