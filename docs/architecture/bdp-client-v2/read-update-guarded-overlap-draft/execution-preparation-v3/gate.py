"""Default-closed overlap gate; no product imports or automatic release."""
from pathlib import Path
import json,hashlib,subprocess,re
P=Path(__file__).resolve().parent;D=P.parent;L=D.parent;E=P/"run"
W=Path("/Users/dbox/repos/wt/vickie/vickie-sequence-member-evidence")
GUARD=L/"guarded-action.py";COMMANDS=P/"commands.json"
COMMANDS_SHA="450c609ddfe4943b39fbad8883d1b91e9ef606688ae9e3116f496e388f826f47"
# Actual lint-repair source; release and repair council remain explicit separate gates.
BINDING_SHA="792d7e813bc153a7bcf1d7bf4c5750522644873c831d8812b42301702b512f6e"
sha=lambda b:hashlib.sha256(b).hexdigest()
def require(ok,message):
 if not ok:raise RuntimeError(message)
def git(*a):return subprocess.check_output(["git",*a],cwd=W,timeout=10)
def document(g,key):
 x=g[key];p=Path(x["path"]);require(p.is_relative_to(L) and sha(p.read_bytes())==x["sha256"],"Document differs: "+key);return json.loads(p.read_text())
def source_hashes():
 g=json.loads((P/"execution-gate.json").read_text());b=document(g,"sourceBinding")
 require(g["sourceBinding"]["sha256"]==BINDING_SHA,"Source rebind pending")
 require(set(b["files"])==set(json.loads(COMMANDS.read_text())["ownedPaths"]),"Two-path scope differs")
 expected={p:x["sha256"] for p,x in b["files"].items()};actual={p:sha((W/p).read_bytes()) for p in expected};require(actual==expected,"Frozen source differs");return actual
def source_patch_hash():return sha(git("diff","--binary","HEAD"))
def completed(name):
 x=json.loads((E/(name+".exit.json")).read_text());spec=json.loads(COMMANDS.read_text())["commands"][name]
 require(x["exit"]==spec.get("expectedExit",0) and x["supervisorExit"]==0 and x["sourceUnchanged"],"Prior stage failed: "+name)
 for k in ("timeout","forcedCleanup","errors","remainingAfterCleanup","sampledPidsStillPresent","runtimeDirectoriesStillPresent"):require(not x[k],"Unclean prior stage: "+name)
 require(x["sourceSha256After"]==source_hashes(),"Prior source differs");return x
def build_outputs():
 paths=[W/"LICENSE"]
 for n in ("protocol","server"):
  root=W/"packages"/n;paths.append(root/"package.json")
  for sub in (("dist","schemas") if n=="protocol" else ("dist",)):
   paths.extend(p for p in (root/sub).rglob("*") if p.is_file())
 require(all(not p.is_symlink() for p in paths),"Output symlink")
 return {p.relative_to(W).as_posix():sha(p.read_bytes()) for p in sorted(paths)}
def verify_build(g):
 i=document(g,"buildOutputInventory");require(i["sourceSha256"]==source_hashes(),"Build source differs")
 completed("build");require(i["buildExitSha256"]==sha((E/"build.exit.json").read_bytes()),"Build record changed")
 names=git("ls-files","--cached","--others","--exclude-standard").decode().splitlines()
 require({n:sha((W/n).read_bytes()) for n in names}==i["inputFiles"],"Build inputs differ")
 require(i["files"]==build_outputs(),"Build outputs differ")
def load_gate(name):
 require(__debug__,"Assertions required");g=json.loads((P/"execution-gate.json").read_text());require(g.get("released") is True,"Gate closed")
 require(g["session"]=="2026-10-08T20:26:35Z-81e5","Wrong session")
 for k in ("rootAcceptance","reanchorAcceptance","sourceCouncil"):document(g,k)
 require(sha(COMMANDS.read_bytes())==COMMANDS_SHA,"Commands differ")
 require(set(g["preparationSha256"])=={"gate.py","supervise.py","inventory.py","packages.py","review.py","derivative.py","derivative-payload.mjs","commands.json"},"Incomplete preparation pins")
 for n,h in g["preparationSha256"].items():require(sha((P/n).read_bytes())==h,"Preparation differs: "+n)
 cfg=json.loads(COMMANDS.read_text());order=cfg["executionOrder"];require(name in g["approvedCommandIds"] and g["approvedCommandIds"]==order[:len(g["approvedCommandIds"])],"Stage not released")
 source=source_hashes();spec=dict(cfg["commands"][name]);require(spec["supervisorSeconds"]==120,"Wrong deadline")
 require(re.fullmatch("[a-f0-9]{40}",g["mergedParent"]) is not None,"Actual merge pending")
 for p,h in g["reviewedRuHashes"].items():require(sha((W/p).read_bytes())==h,"RU basis changed")
 if spec["phase"]=="static":
  require(git("rev-parse","HEAD").decode().strip()==g["mergedParent"],"Static parent changed")
  changed=set(git("diff","--name-only","HEAD").decode().splitlines())|set(git("ls-files","--others","--exclude-standard").decode().splitlines())
  require(changed==set(source),"Static two-path footprint differs")
 else:
  require(not git("status","--porcelain").strip(),"Clean source required")
  require(git("rev-parse","HEAD").decode().strip()==g["finalHead"] and git("rev-parse","HEAD^{tree}").decode().strip()==g["finalTree"],"Final source changed")
  require(git("show","-s","--format=%P","HEAD").decode().strip()==g["mergedParent"],"Ordinary ancestry differs")
  require(set(git("diff","--name-only",g["mergedParent"],"HEAD").decode().splitlines())==set(source),"Scope differs")
  final=document(g,"finalSource");require(final["head"]==g["finalHead"] and final["sourceBindingSha256"]==BINDING_SHA,"Final source document differs")
  verify_build(g)
 for prior in order[:order.index(name)]:completed(prior)
 require(sha((W/"pnpm-lock.yaml").read_bytes())==g["sourceLockSha256"],"Source lock changed")
 if name in ("prepare-derivative","derivative-payload","negative","negative-review"):document(g,"positiveAcceptance")
 if name in ("derivative-payload","negative","negative-review"):document(g,"derivativeAcceptance")
 if spec["phase"]=="installed":
  require(sha((E/"consumer/pnpm-lock.yaml").read_bytes())==g["consumerLockSha256"],"Consumer lock changed")
  if name=="positive":
   p=document(g,"installedProvenance");entry=Path(p["entryRealpath"]).resolve(strict=True)
   require(entry.is_relative_to((E/"consumer").resolve()) and sha(entry.read_bytes())==p["entrySha256"],"Wrong consumer entry")
  else:
   p=document(g,"derivativeBinding");entry=Path(p["entry"]).resolve(strict=True)
   require(entry.is_relative_to((E/"derivative").resolve()) and sha(entry.read_bytes())==p["entrySha256"],"Wrong derivative entry")
   document(g,"derivativePayload")
  spec["validatedBdpOverrides"]={"BDP_RU_GUARDED_OVERLAP_ENTRY":str(entry),"BDP_RU_GUARDED_OVERLAP_MODE":spec["mode"],"BDP_RU_GUARDED_OVERLAP_OUTPUT":str(E/(name+"-observation.json"))}
 g={**g,"sourceSha256":source,"sourceDiffSha256":source_patch_hash(),"preparationSha256":sha((P/"supervise.py").read_bytes())};return g,spec
