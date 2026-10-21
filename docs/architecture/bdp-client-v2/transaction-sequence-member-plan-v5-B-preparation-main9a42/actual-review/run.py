from pathlib import Path
import subprocess,json,hashlib,datetime,time,os,signal
W=Path("/Users/dbox/repos/wt/vickie/vickie-changefeed-parameter-context")
E=Path("/Users/dbox/repos/wt/vickie/bdp-integration-logs/transaction-sequence-member-plan-v5-B-preparation-main9a42/actual-review")
T=E/"tmp"
sha=lambda b:hashlib.sha256(b).hexdigest()
binding=json.loads((E/"launch-binding.json").read_text())
def git(*a):return subprocess.check_output(["git",*a],cwd=W)
def write(name,obj):
 with (E/name).open("x") as f:json.dump(obj,f,indent=2);f.write("\n")
def snapshot():
 return {"head":binding["head"],"tree":git("rev-parse",binding["head"]+"^{tree}").decode().strip(),"sourceHashes":{p:sha(git("show",binding["head"]+":"+p)) for p in binding["sourceHashes"]},"artifacts":{p:sha(Path(p).read_bytes()) for p in binding["artifacts"]},"finalPromptSha256":sha((E/"prompt.txt").read_bytes())}
def rows():return subprocess.check_output(["ps","-axo","pid=,ppid=,pgid=,etime=,state=,comm="],text=True).splitlines()
before=snapshot()
assert before["tree"]==binding["tree"]
assert before["sourceHashes"]==binding["sourceHashes"]
assert before["artifacts"]==binding["artifacts"]
assert before["finalPromptSha256"]==binding["finalPromptSha256"]
write("source-before.json",before)
write("workspace-before.json",{"head":git("rev-parse","HEAD").decode().strip(),"status":git("status","--porcelain").decode(),"contextOnly":True})
T.mkdir(exist_ok=False)
args=["/opt/homebrew/bin/claude","--safe-mode","-p","--tools","","--no-session-persistence"]
env=os.environ.copy();env["TMPDIR"]=str(T)
start=time.monotonic();timed=False;forced=False;seen=set();seen_temp=set()
with (E/"stdout.txt").open("xb") as out,(E/"stderr.txt").open("xb") as err,(E/"processes.jsonl").open("x") as samples:
 proc=subprocess.Popen(args,cwd=W,env=env,stdin=subprocess.PIPE,stdout=out,stderr=err,start_new_session=True)
 write("process.json",{"argv":args,"cwd":str(W),"pid":proc.pid,"processGroup":proc.pid,"started":datetime.datetime.now(datetime.timezone.utc).isoformat(),"timeoutSeconds":900,"originalPromptSha256":binding["originalPromptSha256"],"finalPromptSha256":binding["finalPromptSha256"],"sourceBeforeSha256":sha((E/"source-before.json").read_bytes()),"sourceMode":binding["sourceMode"],"tempDirectory":str(T),"reviewKind":"supplied-context B-only; no model override or tools"})
 proc.stdin.write((E/"prompt.txt").read_bytes());proc.stdin.close()
 while proc.poll() is None:
  elapsed=time.monotonic()-start
  group=[r for r in rows() if len(r.split(None,5))==6 and r.split(None,5)[2]==str(proc.pid)]
  seen.update(int(r.split(None,5)[0]) for r in group)
  seen_temp.update(p.relative_to(T).as_posix() for p in T.rglob("*"))
  samples.write(json.dumps({"elapsedSeconds":elapsed,"rows":group})+"\n");samples.flush()
  if elapsed>=900:
   timed=True;forced=True
   try:os.killpg(proc.pid,signal.SIGTERM)
   except ProcessLookupError:pass
   try:proc.wait(timeout=5)
   except subprocess.TimeoutExpired:
    try:os.killpg(proc.pid,signal.SIGKILL)
    except ProcessLookupError:pass
    proc.wait(timeout=5)
   break
  time.sleep(1)
 proc.wait()
after=snapshot();write("source-after.json",after)
write("workspace-after.json",{"head":git("rev-parse","HEAD").decode().strip(),"status":git("status","--porcelain").decode(),"contextOnly":True})
actualrows=rows();live={int(r.split(None,5)[0]):r for r in actualrows if len(r.split(None,5))==6}
remaining=[r for r in actualrows if len(r.split(None,5))==6 and r.split(None,5)[2]==str(proc.pid)]
cleanup={"checkedAt":datetime.datetime.now(datetime.timezone.utc).isoformat(),"pid":proc.pid,"remainingGroup":remaining,"sampledPids":sorted(seen),"sampledPidsStillPresent":{str(n):live[n] for n in seen if n in live},"sampledTemporaryEntries":sorted(seen_temp),"temporaryEntriesAfter":sorted(p.relative_to(T).as_posix() for p in T.rglob("*")),"limitation":"One-second process-group/temp samples can miss short-lived or detached descendants. Temporary content retained without reading or printing. No product runtime/store launched."}
write("cleanup.json",cleanup)
result={"exitCode":proc.returncode,"timeout":timed,"forcedTermination":forced,"elapsedSeconds":time.monotonic()-start,"committedSourceAndArtifactsUnchanged":before==after,"stdoutSha256":sha((E/"stdout.txt").read_bytes()),"stderrSha256":sha((E/"stderr.txt").read_bytes()),"cleanupSha256":sha((E/"cleanup.json").read_bytes()),"sourceAfterSha256":sha((E/"source-after.json").read_bytes())}
write("exit.json",result);print(json.dumps(result),flush=True)
