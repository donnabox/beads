"""Closed guarded-batch static/package/cohort gate. No product imports or self-release."""
from pathlib import Path
import hashlib,json,re,subprocess
P=Path(__file__).resolve().parent;D=P.parent;L=D.parent;E=P/"run"
W=Path("/Users/dbox/repos/wt/vickie/vickie-guarded-batch")
GUARD=L/"guarded-action.py";COMMANDS=P/"commands.json";COMMANDS_SHA="afd02c125f9675e0af26740b7fd032a3f7c50c52b0703d51f60d2993baa8e653"
sha=lambda b:hashlib.sha256(b).hexdigest()
def require(ok,message):
 if not ok:raise RuntimeError(message)
def git(*a):return subprocess.check_output(["git",*a],cwd=W,timeout=15)
def document(g,key):
 r=g[key];p=Path(r["path"]);require(p.is_relative_to(L) and sha(p.read_bytes())==r["sha256"],"Bound document changed: "+key);return json.loads(p.read_text())
def current_gate():return json.loads((P/"execution-gate.json").read_text())
def source_hashes():
 g=current_gate();b=document(g,"sourceBinding");files={p:x["sha256"] for p,x in b["files"].items()}
 require(set(files)==set(json.loads(COMMANDS.read_text())["ownedPaths"]) and len(files)==2,"Two-path scope required")
 actual={p:sha((W/p).read_bytes()) for p in files};require(actual==files,"Source changed");return actual
def source_patch_hash():return sha(git("diff","--binary","HEAD","--"))
def input_files():
 names=set(git("ls-files").decode().splitlines())|set(source_hashes())
 return {p:sha((W/p).read_bytes()) for p in sorted(names)}
def build_outputs():
 paths=[W/"LICENSE"]
 for n in ("protocol","server"):
  root=W/"packages"/n;paths.append(root/"package.json")
  for sub in (("dist","schemas") if n=="protocol" else ("dist",)):paths.extend(p for p in (root/sub).rglob("*") if p.is_file())
 require(all(not p.is_symlink() for p in paths),"Package output symlink")
 return {p.relative_to(W).as_posix():sha(p.read_bytes()) for p in sorted(paths)}
def emitted_files():return {p.relative_to(W).as_posix():sha(p.read_bytes()) for root in (W/"packages").iterdir() if root.is_dir() for p in (root/"dist").rglob("*") if p.is_file()}
def completed(name):
 x=json.loads((E/(name+".exit.json")).read_text());expected=json.loads(COMMANDS.read_text())["commands"][name]["expectedExit"]
 require(x["exit"]==x["supervisorExit"]==expected and x["sourceUnchanged"],"Prior stage failed: "+name)
 for k in ("timeout","forcedCleanup","errors","remainingAfterCleanup","sampledPidsStillPresent","runtimeDirectoriesStillPresent"):require(not x[k],"Unclean prior: "+name+" "+k)
 require(x["sourceSha256After"]==source_hashes(),"Prior source differs")
 return x
def verify_build(g):
 inv=document(g,"buildOutputInventory")
 require(Path(g["buildOutputInventory"]["path"])==E/"build-output-inventory.json","Wrong inventory")
 require(inv["sourceSha256"]==source_hashes() and inv["inputFiles"]==input_files(),"Build inputs changed")
 require(inv["files"]==build_outputs() and inv["emittedFiles"]==emitted_files(),"Build outputs changed")
 for key in ("buildStart","buildExit","buildLog"):
  r=inv[key];require(sha(Path(r["path"]).read_bytes())==r["sha256"],"Build evidence changed")
 require(Path(inv["buildExit"]["path"])==E/"build.exit.json","Wrong actual build")
 completed("build")
 # Never relabel a dirty/pre-integration build as a clean candidate execution.
 if inv["buildHead"]!=g["finalHead"] or inv["workspaceDirty"]:
  bridge=document(g,"buildBridge")
  require(bridge["accepted"] is True and bridge["buildHead"]==inv["buildHead"] and bridge["buildWorkspaceDirty"]==inv["workspaceDirty"] and bridge["finalHead"]==g["finalHead"] and bridge["finalTree"]==g["finalTree"] and bridge["buildInventorySha256"]==g["buildOutputInventory"]["sha256"],"Explicit actual build bridge required")
def load_gate(name):
 require(__debug__,"Assertions required");g=current_gate()
 require(g.get("released") is True and g["session"]=="2026-10-08T20:26:35Z-81e5","Root gate closed")
 require(sha(COMMANDS.read_bytes())==COMMANDS_SHA and sha(GUARD.read_bytes())==g["guardSha256"],"Commands/guard changed")
 required={"gate.py","supervise.py","packages.py","inventory.py","review-cohort.py","commands.json"}
 require(set(g["preparationSha256"])==required,"Incomplete preparation hashes")
 for n,h in g["preparationSha256"].items():require(sha((P/n).read_bytes())==h,"Preparation changed: "+n)
 acceptance=document(g,"rootAcceptance");require(acceptance["released"] is True,"Root acceptance closed")
 cfg=json.loads(COMMANDS.read_text());require(cfg.get("frozenDraftReady") is True,"Measured draft/source/argv binding pending");order=cfg["executionOrder"]
 require(g["approvedCommandIds"]==order[:len(g["approvedCommandIds"])] and name in g["approvedCommandIds"],"Command not in released prefix")
 spec=dict(cfg["commands"][name]);require(isinstance(spec.get("argv"),list) and all(isinstance(a,str) for a in spec["argv"]),"Actual argv pending");require(spec["supervisorSeconds"]==120,"Time bound changed")
 require(g["node"]=="/Users/dbox/.nvm/versions/node/v24.16.0/bin/node" and g["pnpm"]=="/Users/dbox/.cache/node/corepack/v1/pnpm/11.20.0/bin/pnpm.cjs","Toolchain changed")
 for tool,h in cfg["toolFiles"].items():require(sha(Path(tool).read_bytes())==h,"Tool changed: "+tool)
 source=source_hashes()
 require(git("rev-parse","HEAD").decode().strip()==g["sourceHead"] and git("rev-parse","HEAD^{tree}").decode().strip()==g["sourceTree"],"Snapshot identity changed")
 require(git("status","--porcelain=v1","--untracked-files=all").decode()==g["sourceStatus"] and source_patch_hash()==g["sourceDiffSha256"],"Snapshot status/diff changed")
 for prior in order[:order.index(name)]:completed(prior)
 if spec["phase"] in ("package","installed","review"):
  require(re.fullmatch("[0-9a-f]{40}",g.get("finalHead") or "") is not None and not git("status","--porcelain").strip(),"Clean final candidate required")
  require(g["finalHead"]==g["sourceHead"] and g["finalTree"]==g["sourceTree"],"Final source mismatch")
  require(git("show","-s","--format=%P","HEAD").decode().strip()==g["finalBase"],"Ordinary final-parent reanchor required")
  require(g["footprintBase"]=="9bf601bfb570adae97ce8cdc99caf2cc7a2b34c1" and git("merge-base",g["footprintBase"],g["finalBase"]).decode().strip()==g["footprintBase"],"Feature footprint ancestry differs")
  require(set(git("diff","--name-only",g["footprintBase"],g["finalHead"]).decode().splitlines())==set(source),"Final feature scope differs")
  require(g["finalBase"]=="9e5f040edab88be364f683098ead72f595f0ab55","Ordinary repair successor parent required")
  final=document(g,"finalSource");integration=document(g,"finalIntegrationAcceptance");review=document(g,"codeReviewState");require(review["validationAllowed"] is True,"Local validation not released")
  require(final["head"]==integration["head"]==g["finalHead"] and final["tree"]==integration["tree"]==g["finalTree"] and final["base"]==integration["base"]==g["finalBase"] and integration["accepted"] is True,"Final integration not accepted")
  require(sha((W/"pnpm-lock.yaml").read_bytes())==g["sourceLockSha256"],"Lock changed")
  verify_build(g)
 if spec["phase"] in ("installed","review"):
  contract=document(g,"reportContract");require(contract["ready"] is True and contract["sourceBindingSha256"]==g["sourceBinding"]["sha256"] and contract["reviewerSha256"]==sha((P/"review-cohort.py").read_bytes()),"Frozen report adapter pending")
  p=document(g,"installedProvenance");require(p["sourceCommit"]==g["finalHead"] and p["sourceTree"]==g["finalTree"],"Installed attribution differs")
  entry=Path(p["entryRealpath"]).resolve(strict=True)
  require(entry==(E/"consumer/node_modules/@bdp/server/dist/transaction-installed-process.js").resolve(strict=True) and entry.is_relative_to((E/"consumer").resolve()),"External entry required")
  require(sha(entry.read_bytes())==p["entrySha256"] and sha((E/"consumer/pnpm-lock.yaml").read_bytes())==g["consumerLockSha256"]==p["consumerLockSha256"],"External consumer changed")
  if spec["phase"]=="installed":
   t=g["installedTestTool"];require(t["path"]==spec["argv"][1] and sha(Path(t["path"]).read_bytes())==t["sha256"],"Post-setup Vitest binding required")
   spec["validatedBdpOverrides"]={"BDP_TRANSACTIONAL_INSTALLED_ENTRY":str(entry),"BDP_TRANSACTIONAL_SOURCE_COMMIT":g["finalHead"],"BDP_TRANSACTIONAL_GUARDED_BATCH_MODE":spec["mode"],"BDP_TRANSACTIONAL_GUARDED_BATCH_OUTPUT":str(E/(name+".observation.json"))}
 return {**g,"sourceSha256":source,"sourceDiffSha256":source_patch_hash(),"preparationSha256":sha((P/"supervise.py").read_bytes())},spec
