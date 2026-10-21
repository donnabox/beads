"""Record actual fresh build inputs and output bytes, with no product imports."""
import json
from pathlib import Path
from gate import E,W,load_gate,completed,source_hashes,input_files,build_outputs,emitted_files,sha,require
g,_=load_gate("build-inventory")
result=completed("fresh-build")
start=json.loads((E/"fresh-build.started.json").read_text())
require(start["head"]==g["finalHead"] and start["workspaceDirty"] is False,"Fresh clean build required")
inputs=input_files();require(inputs==start["buildInputFiles"],"Inputs changed during build")
def pin(p):return {"path":str(p),"sha256":sha(p.read_bytes())}
record={"sourceCommit":g["finalHead"],"sourceTree":g["finalTree"],"sourceSha256":source_hashes(),"buildStart":pin(E/"fresh-build.started.json"),"buildExit":pin(E/"fresh-build.exit.json"),"buildLog":pin(E/"fresh-build.log"),"inputFiles":inputs,"files":build_outputs(),"emittedFiles":emitted_files(),"basis":"Actual final clean integrated source and build; complete tracked input hashes at start and inventory, actual build exit, fresh output inventory. Not child-attested Git identity."}
require(len(record["files"])==312,"Unexpected package footprint")
with (E/"build-output-inventory.json").open("x") as f:f.write(json.dumps(record,indent=2)+"\n")
print(json.dumps({"inputFiles":len(inputs),"packageOutputs":len(record["files"]),"emittedFiles":len(record["emittedFiles"]),"inventorySha256":sha((E/"build-output-inventory.json").read_bytes())}))
