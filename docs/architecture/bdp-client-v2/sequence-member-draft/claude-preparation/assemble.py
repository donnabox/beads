"""Assemble supplied-context review input only; never launches Claude or product code."""
from pathlib import Path
import json, hashlib, subprocess
E=Path(__file__).resolve().parent
D=E.parent
W=Path("/Users/dbox/repos/wt/vickie/vickie-sequence-member-evidence")
BASE="18104a03b850553641f24fe83e7318493fa102ca"
OWNED=["packages/conformance/test-support/transactional/manifest-v1.json","packages/conformance/test-support/transactional/coverage-v1.json","packages/conformance/src/transactional-cumulative.test.ts","apps/bdptest/src/transactional-matrix.test.ts"]
sha=lambda b:hashlib.sha256(b).hexdigest()
def git(*a):return subprocess.check_output(["git",*a],cwd=W)
def load(p):return json.loads(p.read_text())
def checked(p,h):
 b=p.read_bytes()
 assert sha(b)==h,str(p)
 return b
request=load(E/"assembly-request.json") # absent by default; example is not authority
assert request.get("released") is True,"Root assembly release required"
assert not (E/"prompt.txt").exists() and not (E/"final-launch-binding.json").exists()
for name,key in [("context-inventory.json","contextInventorySha256"),("accepted-context.json","acceptedContextSha256"),("prompt.template.txt","promptTemplateSha256")]:
 checked(E/name,request[key])
inventory=load(E/"context-inventory.json")
assert inventory["base"]==BASE
artifacts={}
for key in ["sourceBinding","fullPatch","measuredAcceptance","measurementSummary","nativeReport"]:
 p=Path(request[key+"Path"]).resolve()
 assert p.is_relative_to(D.resolve()) and p.is_file(),key
 artifacts[key]=(p,checked(p,request[key+"Sha256"]))
source=load(artifacts["sourceBinding"][0])
assert source["workspace"]==str(W) and source["head"]==BASE
assert source["branch"]=="codex/vickie-sequence-member-transactions"
assert set(source["files"])==set(OWNED)
assert git("rev-parse","HEAD").decode().strip()==BASE
assert git("branch","--show-current").decode().strip()==source["branch"]
assert git("rev-parse","HEAD^{tree}").decode().strip()==source["tree"]
assert not git("ls-files","--others","--exclude-standard").strip()
assert set(git("diff","--name-only","HEAD","--").decode().splitlines())==set(OWNED)
diff=git("diff","--binary","HEAD","--")
assert diff==artifacts["fullPatch"][1]
assert sha(diff)==source["diffSha256"]
source_hashes={}
for p in OWNED:
 b=checked(W/p,source["files"][p]["sha256"])
 assert not source["files"][p].get("new",False),p
 source_hashes[p]=sha(b)
plans={str(p):sha(b) for p,b in artifacts.values()}
parts=[(E/"prompt.template.txt").read_text()]
def section(title,text):parts.extend(["\n\n===== "+title+" =====\n",text])
for p,h in load(E/"accepted-context.json")["files"].items():
 b=checked(Path(p),h);plans[p]=h;section(p+" SHA256 "+h,b.decode())
for key in ["sourceBinding","measuredAcceptance","measurementSummary","nativeReport","fullPatch"]:
 p,b=artifacts[key];section(str(p)+" SHA256 "+sha(b),b.decode())
def blocks(p,b,entry):
 text=b.decode();header=p+" whole-file SHA256 "+sha(b)
 if "schemaClosure" in entry:
  doc=json.loads(text);selected={}
  def visit(value):
   if isinstance(value,dict):
    ref=value.get("$ref")
    if ref:
     assert ref.startswith("#/$defs/") and "/" not in ref[len("#/$defs/"):],ref
     include(ref[len("#/$defs/"):])
    for v in value.values():visit(v)
   elif isinstance(value,list):
    for v in value:visit(v)
  def include(name):
   if name not in selected:
    selected[name]=doc["$defs"][name];visit(selected[name])
  for name in entry["schemaClosure"]:include(name)
  section(header+" transitive local schema definitions",json.dumps({"$id":doc.get("$id"),"requestedRoots":entry["schemaClosure"],"$defs":selected},indent=2));return
 if "lineExcerpts" in entry:
  lines=text.splitlines()
  for a,z in entry["lineExcerpts"]:
   assert 0<a<=z<=len(lines)
   section(header+" explicit excerpt lines "+str(a)+"-"+str(z),"\n".join(lines[a-1:z]))
  return
 if not entry.get("blocks"):section(header+" complete file",text);return
 for start,end in entry["blocks"]:
  assert start is None or text.count(start)==1,(p,start)
  assert end is None or text.count(end)==1,(p,end)
  a=0 if start is None else text.index(start)
  z=len(text) if end is None else text.index(end)
  assert a<z,(p,start,end)
  section(header+" complete named block, starts line "+str(text[:a].count("\n")+1),text[a:z])
for entry in inventory["files"]:
 p=entry["path"];b=git("show",BASE+":"+p)
 assert sha(b)==entry["committedSha256"]
 checked(W/p,sha(b));source_hashes[p]=sha(b)
 blocks(p,b,entry)
for entry in inventory["ownedCriticalBlocks"]:
 p=entry["path"];blocks(p,checked(W/p,source_hashes[p]),entry)
prompt="".join(parts).encode()
assert len(prompt)<=request["maximumPromptBytes"],"Prompt too large; root must inspect, never truncate"
# Recheck mutable inputs after assembly, before publishing a frozen prompt.
assert git("diff","--binary","HEAD","--")==diff
for p,h in source_hashes.items():checked(W/p,h)
for p,h in plans.items():checked(Path(p),h)
for p in ["assembly-request.json","context-inventory.json","accepted-context.json","prompt.template.txt"]:
 plans[str(E/p)]=sha((E/p).read_bytes())
binding={"status":"frozen-code-review","head":BASE,"tree":source["tree"],"statusPorcelain":git("status","--porcelain=v1","--untracked-files=all").decode(),"trackedDiffSha256":sha(diff),"untrackedFiles":[],"sourceHashes":source_hashes,"plans":plans,"ownedPaths":OWNED,"expectedOwnedPaths":OWNED,"frozenSourceManifest":str(artifacts["sourceBinding"][0]),"frozenSourceManifestSha256":sha(artifacts["sourceBinding"][1]),"fullPatch":str(artifacts["fullPatch"][0]),"fullPatchSha256":sha(diff),"promptSha256":sha(prompt),"promptBytes":len(prompt),"launchReleased":False,"contextLimitation":"Supplied-context code review; no unchanged full fixture or execution logs. Exact diff, critical definitions, schema closure, accepted plan and measured decision supplied."}
with (E/"prompt.txt").open("xb") as f:f.write(prompt)
with (E/"final-launch-binding.json").open("x") as f:json.dump(binding,f,indent=2);f.write("\n")
print(json.dumps({"promptBytes":len(prompt),"promptSha256":sha(prompt),"bindingSha256":sha((E/"final-launch-binding.json").read_bytes()),"launchReleased":False}))
