"""Two concrete disposable source derivatives. No worktree/commit or product writes."""
from pathlib import Path
import difflib,json,re,sys
from gate import W,E,DRAFT,load_gate,git,sha,require
name=sys.argv[1]
require(name in ("restore-command","bypass-receipt"),"Unknown derivative")
g,_=load_gate(name+"-prepare")
config=g["controls"][name]
patch=Path(config["patchPath"])
require(patch.parent==DRAFT and sha(patch.read_bytes())==config["patchSha256"],"Wrong reviewed patch")
target="packages/server/src/transaction-http.ts"
text=(W/target).read_text()
lines=patch.read_text().splitlines(keepends=True)
require(lines[:2]==["--- a/"+target+"\n","+++ b/"+target+"\n"],"Wrong patch target")
match=re.fullmatch(r"@@ -(\d+),(\d+) \+(\d+),(\d+) @@\n",lines[2])
require(match is not None and all(line[:1] in (" ","+","-") for line in lines[3:]),"Only one exact hunk accepted")
old=[line[1:] for line in lines[3:] if line[0] in (" ","-")]
new=[line[1:] for line in lines[3:] if line[0] in (" ","+")]
start,count,_,newcount=map(int,match.groups())
require(len(old)==count and len(new)==newcount,"Hunk counts mismatch")
original=text.splitlines(keepends=True)
require(original[start-1:start-1+count]==old,"Hunk input mismatch")
changed="".join(original[:start-1]+new+original[start-1+count:])
require("".join(difflib.unified_diff(original,changed.splitlines(keepends=True),
        fromfile="a/"+target,tofile="b/"+target))==patch.read_text(),"Patch reconstruction mismatch")
require(sha(changed.encode())==config["afterSha256"],"Unapproved derivative bytes")
root=E/name
root.mkdir(exist_ok=False)
records={}
# This exact repository has663 ordinary100644 entries; refuse new file kinds.
for row in git("ls-files","--stage","-z").decode().split("\0"):
    if not row:continue
    meta,rel=row.split("\t",1)
    require(meta.split()[0]=="100644" and meta.split()[2]=="0","Unexpected tracked mode/stage")
    path=Path(rel);require(not path.is_absolute() and ".." not in path.parts,"Unsafe tracked path")
    source=W/path;require(source.is_file() and not source.is_symlink(),"Unexpected source file")
    before=source.read_bytes();data=changed.encode() if rel==target else before
    dest=root/path;dest.parent.mkdir(parents=True,exist_ok=True);dest.write_bytes(data);dest.chmod(0o644)
    records[rel]={"sha256":sha(data),"inputSha256":sha(before),"executable":False}
require(len(records)==663,"Tracked inventory changed; review before copying")
for rel,info in records.items():require(sha((W/rel).read_bytes())==info["inputSha256"],"Source changed during copy")
out={"status":"created-not-built-not-executed","directory":str(root),
     "sourceHead":g["sourceHead"],"sourceSha256":g["sourceSha256"],
     "patchPath":str(patch),"patchSha256":config["patchSha256"],
     "changedPaths":[target],"files":records,
     "attribution":"Plain disposable copy of exact reviewed tracked bytes, not a Git commit; no product checkout writes"}
p=E/(name+".snapshot.json")
with p.open("x") as f:f.write(json.dumps(out,indent=2)+"\n")
print(json.dumps({"snapshot":str(p),"sha256":sha(p.read_bytes()),"files":len(records)}))
