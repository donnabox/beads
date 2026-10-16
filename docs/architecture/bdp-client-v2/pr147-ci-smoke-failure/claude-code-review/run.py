from pathlib import Path
import subprocess,time,json,hashlib,os,signal,datetime
p=Path(__file__).parent
w=Path("/Users/dbox/repos/wt/vickie/vickie-fresh-depth-serving")
b=json.loads((p/"binding.json").read_text());sha=lambda x:hashlib.sha256(x).hexdigest()
def guard(): subprocess.run(["python3","/Users/dbox/repos/wt/vickie/bdp-integration-logs/guarded-action.py","/usr/bin/true"],check=True)
assert b['rootLaunchReleased'] is True, 'Closed: root must explicitly release this one review'
for name in ['source-before.json','source-after.json','process.json','stdout.txt','stderr.txt','exit.json']:
 assert not (p/name).exists(), 'Preserve prior attempt; no retry'
def snap():return {"files":{f:sha((w/f).read_bytes()) for f in b["files"]},"context":{f:sha(Path(f).read_bytes()) for f in b["context"]},"artifacts":{f:sha(Path(f).read_bytes()) for f in b["artifacts"]},"head":subprocess.check_output(["git","rev-parse","HEAD"],cwd=w,text=True).strip(),"promptSha256":sha((p/"prompt.txt").read_bytes()),"runnerSha256":sha((p/"run.py").read_bytes())}
before=snap();assert before["files"]==b["files"];assert before["promptSha256"]==b["promptSha256"];assert before["runnerSha256"]==b["runnerSha256"]
assert before["head"]==b["head"] and before["context"]==b["context"] and before["artifacts"]==b["artifacts"]
assert set(subprocess.check_output(["git","diff","--name-only","HEAD"],cwd=w,text=True).splitlines()) == set(b["files"]), "Only reviewed four-line smoke delta allowed"
guard();(p/"source-before.json").write_text(json.dumps(before,indent=2))
args=["/opt/homebrew/bin/claude","--safe-mode","-p","--tools","Read,Glob,Grep","--allowedTools","Read,Glob,Grep","--disable-slash-commands","--strict-mcp-config","--mcp-config",'{"mcpServers":{}}',"--no-session-persistence"]
start=time.monotonic();timed=False;sampled=set()
with (p/"stdout.txt").open("xb") as out,(p/"stderr.txt").open("xb") as err,(p/"prompt.txt").open("rb") as inp:
 c=subprocess.Popen(args,cwd=w,stdin=inp,stdout=out,stderr=err,start_new_session=True)
 (p/"process.json").write_text(json.dumps({"pid":c.pid,"pgid":c.pid,"started":datetime.datetime.now(datetime.timezone.utc).isoformat(),"argv":args,"timeoutSeconds":900,"termAfterSeconds":895,"headIsChronologyOnly":False},indent=2))
 while c.poll() is None:
  rows=subprocess.check_output(["ps","-axo","pid=,ppid=,pgid=,comm="],text=True).splitlines()
  sampled.update(int(r.split()[0]) for r in rows if len(r.split(None,3))==4 and r.split(None,3)[2]==str(c.pid))
  if time.monotonic()-start>=895:
   timed=True
   try:os.killpg(c.pid,signal.SIGTERM)
   except ProcessLookupError:pass
   try:c.wait(timeout=5)
   except subprocess.TimeoutExpired:os.killpg(c.pid,signal.SIGKILL);c.wait()
   break
  time.sleep(2)
 after=snap();rows=subprocess.check_output(["ps","-axo","pid=,ppid=,pgid=,comm="],text=True).splitlines();group=[r for r in rows if len(r.split(None,3))==4 and r.split(None,3)[2]==str(c.pid)]
 # Bound any same-group descendant left behind after the CLI exits.
 if group:
  try:os.killpg(c.pid,signal.SIGTERM)
  except ProcessLookupError:pass
  time.sleep(1)
  try:os.killpg(c.pid,signal.SIGKILL)
  except ProcessLookupError:pass
 rows=subprocess.check_output(["ps","-axo","pid=,ppid=,pgid=,comm="],text=True).splitlines();remaining=[r for r in rows if len(r.split(None,3))==4 and (r.split(None,3)[2]==str(c.pid) or int(r.split()[0]) in sampled)]
 result={"exitCode":c.returncode,"timeout":timed,"elapsedSeconds":time.monotonic()-start,"sourceInputsUnchanged":before["files"]==after["files"] and before["context"]==after["context"] and before["head"]==after["head"],"packetUnchanged":before["artifacts"]==after["artifacts"] and before["promptSha256"]==after["promptSha256"] and before["runnerSha256"]==after["runnerSha256"],"headBefore":before["head"],"headAfter":after["head"],"headIsChronologyOnly":False,"sampledPids":sorted(sampled),"groupAfterCLI":group,"remainingProcesses":remaining,"stdoutSha256":sha((p/"stdout.txt").read_bytes()),"stderrSha256":sha((p/"stderr.txt").read_bytes()),"temporaryDirectoryCreated":False,"noSessionPersistence":True}
 guard();(p/"source-after.json").write_text(json.dumps(after,indent=2));(p/"exit.json").write_text(json.dumps(result,indent=2));print(json.dumps(result))
