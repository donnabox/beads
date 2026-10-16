"""Closed read-only exact PR138 postmerge verifier. No fetch, merge, push or tests."""
from pathlib import Path
import subprocess,json,hashlib,datetime,re
P=Path(__file__).resolve().parent
W=Path("/Users/dbox/repos/wt/vickie/vickie-update-cold-directory")
INPUTS="73273cb2ae1486bc848b79f6042ae6825088073814010fae025480f0b84a195c"
OUTCOME="e767e066a2d46ef91915f353990c9aa4241f440e4d0a2aab573bb3a6b7163a62"
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def read(p):return json.loads(p.read_text())
def need(ok,msg):
 if not ok:raise RuntimeError(msg)
def git(*args):return subprocess.check_output(["git",*args],cwd=W,timeout=30)
def main():
 release=read(P/"root-postmerge-release.json")
 need(release.get("released") is True and release.get("scriptSha256")==sha(Path(__file__)),"Postmerge read release closed")
 need(sha(P/"inspection-inputs.json")==INPUTS and sha(P/"inspection/outcome.json")==OUTCOME,"Candidate/inspection changed")
 cfg=read(P/"inspection-inputs.json");outcome=read(P/"inspection/outcome.json")
 need(sha(P/"inspection/job.log")==outcome["rawLogSha256"],"CI log changed")
 merge=release.get("actualMerge")
 need(isinstance(merge,str) and re.fullmatch("[0-9a-f]{40}",merge),"Actual merge missing")
 need(release.get("head")==cfg["head"] and release.get("base")==cfg["base"] and release.get("tree")==cfg["tree"],"Release identity differs")
 subprocess.run(["python3",str(P.parent.parent/"guarded-action.py"),"/usr/bin/true"],check=True,timeout=40)
 R=P/"postmerge";R.mkdir(exist_ok=False)
 def write(n,v):
  with (R/n).open("x") as f:json.dump(v,f,indent=2);f.write("\n")
 def api(n,route):
  argv=["gh","api","--method","GET","repos/gastownhall/bdp/"+route]
  result=subprocess.run(argv,capture_output=True,timeout=60)
  (R/n).write_bytes(result.stdout);(R/(n+".stderr")).write_bytes(result.stderr)
  need(result.returncode==0,"API read failed: "+route)
  return json.loads(result.stdout)
 try:
  pr=api("pr.json","pulls/138");remote=api("main.json","git/ref/heads/main")
  need(pr["merged"] is True and pr["state"]=="closed" and pr["merge_commit_sha"]==merge and pr["head"]["sha"]==cfg["head"] and pr["base"]["ref"]=="main","Merged PR differs")
  need(remote["object"]["sha"]==merge,"Remote main differs from actual merge")
  obj=api("merge.json","git/commits/"+merge)
  need(obj["sha"]==merge and obj["tree"]["sha"]==cfg["tree"] and [x["sha"] for x in obj["parents"]]==[cfg["base"],cfg["head"]],"API merge tree/ordered parents differ")
  # Root fetches the actual object separately; this script never changes Git state.
  need(git("rev-parse",merge+"^{tree}").decode().strip()==cfg["tree"],"Local merge object unavailable/different")
  need(git("show","-s","--format=%P",merge).decode().strip().split()==[cfg["base"],cfg["head"]],"Local merge parents differ")
  need(git("show","-s","--format=%P",cfg["head"]).decode().strip().split()==[cfg["authorCandidate"]] and git("show","-s","--format=%P",cfg["authorCandidate"]).decode().strip().split()==[cfg["base"]],"Ordinary candidate chain differs")
  need(not git("diff","--raw",cfg["head"],merge),"Merged bytes differ from tested candidate")
  tree=api("tree.json","git/trees/"+cfg["tree"]+"?recursive=1")
  need(tree["sha"]==cfg["tree"] and tree.get("truncated") is False,"Incomplete API tree")
  local={}
  for row in git("ls-tree","-r","-z",merge).split(b"\0"):
   if not row:continue
   meta,path=row.split(b"\t",1);mode,kind,oid=meta.decode().split();local[path.decode()]={"mode":mode,"type":kind,"sha":oid}
  actual={x["path"]:{k:x[k] for k in ["mode","type","sha"]} for x in tree["tree"] if x["type"]!="tree"}
  need(local==actual,"Complete API/local blob inventory differs")
  need(set(git("diff","--name-only",cfg["base"],merge).decode().splitlines())==set(cfg["files"]) and len(cfg["files"])==6,"Merged feature scope differs")
  for name,digest in cfg["files"].items():need(hashlib.sha256(git("show",merge+":"+name)).hexdigest()==digest,"Merged file bytes differ: "+name)
  remote2=api("main-final.json","git/ref/heads/main");need(remote2["object"]["sha"]==merge,"Remote main changed during verification")
  write("outcome.json",{"verified":True,"at":datetime.datetime.now(datetime.timezone.utc).isoformat(),"merge":merge,"tree":cfg["tree"],"parents":[cfg["base"],cfg["head"]],"testedSynthetic":outcome["actualSynthetic"],"fullBlobObjects":len(local),"sixFileScopeExact":True,"remoteMain":merge,"noProductExecution":True,"publicQualificationGranted":False})
 except BaseException as e:
  write("failure.json",{"verified":False,"exception":type(e).__name__,"reason":str(e)});raise
if __name__=="__main__":main()
