"""One external byte-copy and emitted-only guard removal; no product import/build/install."""
from pathlib import Path
import json,shutil,re,difflib
from gate import E,W,sha,load_gate,document,require

g,_=load_gate("prepare-derivative");p=document(g,"installedProvenance");C=Path(p["consumer"]).resolve(strict=True);T=E/"derivative"
require(C==(E/"consumer").resolve() and not T.exists() and not T.is_relative_to(W),"Wrong derivative roots")
def inventory(root):
 files={};links={}
 for q in sorted(root.rglob("*")):
  r=q.relative_to(root).as_posix()
  if q.is_symlink():
   target=q.resolve(strict=True);require(target.is_relative_to(root),"Escaping symlink: "+r);links[r]=target.relative_to(root).as_posix()
  elif q.is_file():files[r]=sha(q.read_bytes())
 return {"files":files,"links":links}
before=inventory(C);shutil.copytree(C,T,symlinks=True);copied=inventory(T);require(copied==before,"Copy differs")
server=(T/"node_modules/@bdp/server").resolve(strict=True);protocol=(T/"node_modules/@bdp/protocol").resolve(strict=True)
site=server/"dist/resource-evaluator.js";old=site.read_bytes()
needle=b'    if ("expectedRevision" in input &&\n        input.expectedRevision !== undefined &&\n        input.expectedRevision !== before?.revision)\n        fail("revision-mismatch");\n'
require(old.count(needle)==1,"Expected sole emitted guard not found")
new=old.replace(needle,b'',1);site.write_bytes(new);after=inventory(T);rel=site.relative_to(T).as_posix()
require(set(after["files"])==set(before["files"]) and after["links"]==before["links"],"Derivative footprint changed")
require([r for r in before["files"] if before["files"][r]!=after["files"][r]]==[rel],"More than sole mutant changed")
# The copied regular files and symlink topology retain every exact proven consumer edge.
graph=[]
for item in p["installedProductionGraph"]:
 manifest=T/Path(item["manifest"]).relative_to(C);require(sha(manifest.read_bytes())==item["manifestSha256"],"Dependency manifest changed")
 edges={}
 for name,edge in item["resolvedDependencies"].items():
  target=T/Path(edge["manifestRealpath"]).relative_to(C);require(target.resolve().is_relative_to(T) and target.is_file(),"Dependency escape")
  # Resolve each bare dependency through Node's ancestor node_modules search.
  found=None
  for base in [manifest.parent,*manifest.parent.parents]:
   if not base.is_relative_to(T):break
   candidate=base/"node_modules"/name/"package.json"
   if candidate.exists():found=candidate.resolve();break
  require(found==target.resolve(),"Copied bare edge differs")
  edges[name]=str(found)
 graph.append({"name":item["name"],"manifest":str(manifest),"resolvedDependencies":edges})
# Capture every static emitted relative edge; no test alias resolution participates.
relative=[]
for root in [server/"dist",protocol/"dist"]:
 for file in sorted(root.rglob("*.js")):
  text=file.read_text()
  for spec in re.findall(r'(?:from\s*|import\s*\(\s*)["\']([^"\']+)["\']',text):
   if spec.startswith('.'):
    target=(file.parent/spec).resolve(strict=True);require(target.is_relative_to(T),"Relative import escape")
    relative.append({"from":str(file),"specifier":spec,"target":str(target),"sha256":sha(target.read_bytes())})
for src,target in [('installed-reference-process.js','installed-reference.js'),('installed-reference.js','development-runtime.js'),('installed-reference.js','development-composition.js'),('development-runtime.js','development-authority.js'),('development-authority.js','sequence-executor.js'),('sequence-executor.js','member-executor.js'),('member-executor.js','resource-evaluator.js')]:
 require(any(Path(e['from']).name==src and Path(e['target']).name==target for e in relative),"Required chain edge absent: "+src+" -> "+target)
entry=(server/"dist/installed-reference-process.js").resolve(strict=True)
result={"sourceCommit":g["finalHead"],"sourceTree":g["finalTree"],"positiveProvenanceSha256":g["installedProvenance"]["sha256"],"root":str(T),"entry":str(entry),"entrySha256":sha(entry.read_bytes()),"site":str(site),"beforeSha256":sha(old),"afterSha256":sha(new),"removedUtf8":needle.decode(),"before":before,"after":after,"productionGraph":graph,"relativeEdges":relative,"retention":"Retain external derivative through root review; root owns later disposal. Not runtime store cleanup."}
with (E/"derivative-binding.json").open('x') as f:f.write(json.dumps(result,indent=2)+'\n')
with (E/"derivative.patch").open('x') as f:f.write(''.join(difflib.unified_diff(old.decode().splitlines(True),new.decode().splitlines(True),fromfile='positive/'+rel,tofile='derivative/'+rel)))
