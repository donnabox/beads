"""Root-invoked one-shot gate transition; runs no package or product commands."""
from pathlib import Path
import datetime, hashlib, importlib.util, json, sys
P=Path(__file__).resolve().parent
D=P.parent/"transaction-command-minimum-evidence-draft"
HEAD="a912c23fbd80e70ead382a27e700a70123bf3f21"
TREE="501791c6372c3d69b50ef38f777e79de5807f10a"
DIFF="06cf7eb8ca3d4e0a4caaeee81ee5688a9e44f24895809414b164833068d207bb"
COUNCIL="eeb207ba8bab4d9e53654c86ef20099d23254957f63f28b594b9c8824d5e7d12"
INVENTORY="7bde8c94657b1b48d7000e7c96bc71d9af11d34f1a93f0b55a9f7ef93ea9878d"
STATIC_GATE="b79ff6bb96c15cab34f3961c463720b628843c72e45b0271a6c64fb29ee54d22"
PACKAGE=["node-version","package-directory","pack-protocol","pack-server","verify-tar","prepare-consumer","consumer-install"]
STATIC=["offline-setup","typecheck","build","format-check","lint","boundaries","read-seal","static-cases","build-inventory"]
sha=lambda b:hashlib.sha256(b).hexdigest()
def require(ok,message):
    if not ok:raise RuntimeError(message)
def encode(x):return (json.dumps(x,indent=2)+"\n").encode()
def main():
    require(__debug__ and sys.argv[1:]==["--release-package-only"],"Explicit root invocation required")
    raw=(P/"execution-gate.json").read_bytes()
    require(sha(raw)==STATIC_GATE,"Static gate changed")
    g=json.loads(raw)
    require(g["approvedCommandIds"]==STATIC and g["finalHead"] is None,"Wrong prior stage")
    for n,h in g["preparationSha256"].items():require(sha((P/n).read_bytes())==h,"Preparation changed: "+n)
    spec=importlib.util.spec_from_file_location("bound_package_release_gate",P/"gate.py")
    gate=importlib.util.module_from_spec(spec);spec.loader.exec_module(gate)
    for key in ("rootAcceptance","growthAcceptance","sourceManifest","councilAcceptance"):gate.document(g,key)
    require(g["sourceManifest"]["sha256"]=="90ced9c641febb7e5cc5d494c7dc837aa55e16b2f538a852066b7b66c5491961","Wrong source manifest")
    require(gate.git("rev-parse","HEAD").decode().strip()==HEAD,"Wrong clean commit")
    require(gate.git("rev-parse","HEAD^{tree}").decode().strip()==TREE,"Wrong tree")
    require(gate.git("show","-s","--format=%P","HEAD").decode().strip()==gate.PARENT,"Wrong ordinary parent")
    require(not gate.git("status","--porcelain").strip(),"Product checkout is not clean")
    require(sha(gate.git("diff","--binary",gate.PARENT,"HEAD"))==DIFF,"Full parent patch changed")
    require(not gate.git("diff","--binary","HEAD"),"Working diff is not empty")
    require(gate.source_hashes()==g["sourceSha256"] and len(g["sourceSha256"])==12,"Twelve source hashes changed")
    require(set(gate.git("diff","--name-only",gate.PARENT,"HEAD").decode().splitlines())==set(g["sourceSha256"]),"Footprint changed")
    require(sha((gate.W/"pnpm-lock.yaml").read_bytes())==g["sourceLockSha256"],"Lock changed")
    councilPath=D/"root-final-context-council.json"
    require(sha(councilPath.read_bytes())==COUNCIL,"Final council changed")
    council=json.loads(councilPath.read_text())
    require(council["accepted"] is True and council["sourceBinding"]["hashes"]==g["sourceSha256"] and council["sourceBinding"]["diffSha256"]==DIFF,"Council is not for this source")
    for doc in council["documents"]:require(sha(Path(doc["path"]).read_bytes())==doc["sha256"],"Council evidence changed")
    for name in STATIC:
        gate.completed(name)
        start=json.loads((gate.E/(name+".started.json")).read_text())
        require(start["sourceSha256"]==g["sourceSha256"] and start["sourceDiffSha256"]==DIFF,"Historical static source changed: "+name)
    summary=json.loads((P/"root-static-summary.json").read_text())
    require(summary["tests"]=={"passed":30,"skipped":601,"failed":0,"selectionMatches":True},"Static selection not accepted")
    inventory=gate.E/"build-output-inventory.json"
    require(sha(inventory.read_bytes())==INVENTORY,"Build inventory changed")
    g["buildOutputInventory"]={"path":str(inventory),"sha256":INVENTORY}
    gate.verify_build(g)
    for name in PACKAGE:
        require(not list(gate.E.glob(name+".*")),"Package stage already attempted: "+name)
    require(not (gate.E/"packages").exists() and not (gate.E/"consumer").exists(),"Fresh package/consumer directories required")
    snapshot=P/"execution-gate-static-completed.json"
    releasePath=P/"root-package-release.json"
    finalPath=D/"final-source.json"
    for path in (snapshot,releasePath,finalPath,P/"execution-gate.package-next.json"):
        require(not path.exists(),"Refusing overwrite: "+str(path))
    at=datetime.datetime.now(datetime.timezone.utc).isoformat()
    final={"at":at,"head":HEAD,"tree":TREE,"parents":[gate.PARENT],"worktree":str(gate.W),"workspaceDirty":False,"hashes":g["sourceSha256"],"parentDiffSha256":DIFF,"workingDiffSha256":sha(b""),"council":{"path":str(councilPath),"sha256":COUNCIL},"historicalBuild":{"path":str(inventory),"sha256":INVENTORY,"recordedHead":gate.PARENT,"recordedWorkspaceDirty":True,"sourceBytesEqualFinal":True},"installedExecuted":False,"profileQualification":False}
    release={"at":at,"rootInvoked":True,"scope":"Only seven package stages through consumer-install; no verify-installed or product execution","approvedCommandIds":PACKAGE,"finalHead":HEAD,"finalTree":TREE,"sourceDiffSha256":DIFF,"staticGateSha256":STATIC_GATE,"councilSha256":COUNCIL,"buildInventorySha256":INVENTORY,"finalSourceSha256":sha(encode(final))}
    g.update(sourceHead=HEAD,sourceTree=TREE,sourceDiffSha256=sha(b""),finalHead=HEAD,finalTree=TREE,approvedCommandIds=PACKAGE,rootAcceptance={"path":str(releasePath),"sha256":sha(encode(release))},councilAcceptance={"path":str(councilPath),"sha256":COUNCIL})
    # All verification precedes external artifact writes; existing run records remain untouched.
    with snapshot.open("xb") as f:f.write(raw)
    with finalPath.open("xb") as f:f.write(encode(final))
    with releasePath.open("xb") as f:f.write(encode(release))
    nextPath=P/"execution-gate.package-next.json"
    with nextPath.open("xb") as f:f.write(encode(g))
    nextPath.replace(P/"execution-gate.json")
    print(json.dumps({"released":PACKAGE,"finalHead":HEAD,"finalTree":TREE,"gateSha256":sha(encode(g)),"finalSource":str(finalPath)}))
if __name__=="__main__":main()
