from pathlib import Path
import subprocess,time,json,hashlib,os,signal,datetime
p=Path(__file__).parent
w=Path("/Users/dbox/repos/wt/vickie/vickie-fresh-depth-serving")
b=json.loads((p/"binding.json").read_text());sha=lambda x:hashlib.sha256(x).hexdigest()
plans=Path("/Users/dbox/repos/wt/vickie/bdp-integration-logs/transaction-fresh-depth-plan/review-repair")
def snap():return {"files":{f:sha((w/f).read_bytes()) for f in b["files"]},"packets":{f:sha((p/f).read_bytes()) for f in ["production.patch","binding.json","prompt.txt"]},"plans":{f:sha((plans/f).read_bytes()) for f in ["plan-v2.md","source-dispositions-v2.md"]},"head":subprocess.check_output(["git","rev-parse","HEAD"],cwd=w,text=True).strip()}
before=snap();assert before["head"]==b["head"]
assert before["files"]=={f:v["sha256"] for f,v in b["files"].items()}
assert before["packets"]["production.patch"]==b["diffSha256"]
subprocess.run(["python3","/Users/dbox/repos/wt/vickie/bdp-integration-logs/guarded-action.py","/usr/bin/true"],check=True)
(p/"source-before.json").write_text(json.dumps(before,indent=2))
start=time.monotonic();timed=False
args=["/opt/homebrew/bin/claude","--safe-mode","-p","--tools","Read,Glob,Grep","--allowedTools","Read,Glob,Grep","--disable-slash-commands","--strict-mcp-config","--mcp-config",'{"mcpServers":{}}',"--add-dir",str(p.parent),"--add-dir",str(plans),"--add-dir",str(Path("/Users/dbox/repos/wt/vickie/bdp-integration-logs/transaction-fresh-depth-draft/installed-author")),"--no-session-persistence"]
with (p/"stdout.txt").open("xb") as out,(p/"stderr.txt").open("xb") as err,(p/"prompt.txt").open("rb") as inp:
 c=subprocess.Popen(args,cwd=w,stdin=inp,stdout=out,stderr=err,start_new_session=True)
 (p/"process.json").write_text(json.dumps({"pid":c.pid,"started":datetime.datetime.now(datetime.timezone.utc).isoformat(),"argv":args,"timeoutSeconds":900},indent=2))
 try:c.wait(timeout=895)
 except subprocess.TimeoutExpired:
  timed=True;os.killpg(c.pid,signal.SIGTERM)
  try:c.wait(timeout=5)
  except subprocess.TimeoutExpired:os.killpg(c.pid,signal.SIGKILL);c.wait()
 after=snap();(p/"source-after.json").write_text(json.dumps(after,indent=2))
 rows=subprocess.check_output(["ps","-axo","pid=,ppid=,pgid=,comm="],text=True).splitlines();group=[r for r in rows if len(r.split(None,3))==4 and r.split(None,3)[2]==str(c.pid)]
 result={"exitCode":c.returncode,"timeout":timed,"elapsedSeconds":time.monotonic()-start,"inputsUnchanged":before==after,"remainingGroup":group,"stdoutSha256":sha((p/"stdout.txt").read_bytes()),"stderrSha256":sha((p/"stderr.txt").read_bytes()),"scope":"three frozen installed test files only; other author scopes excluded"}
 (p/"exit.json").write_text(json.dumps(result,indent=2));print(json.dumps(result))
