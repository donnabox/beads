from pathlib import Path
import subprocess,json,hashlib,datetime,time,os,signal
W=Path("/Users/dbox/repos/wt/vickie/vickie-sequence-member-evidence")
E=Path("/Users/dbox/repos/wt/vickie/bdp-integration-logs/transaction-receipt-missing-context-draft/ci-budget-repair/claude-supplement")
P=E;T=P/"tmp"
assert (E/"final-launch-binding.json").is_file(), "PREPARATION ONLY: final source binding absent"
binding=json.loads((E/"final-launch-binding.json").read_text())
assert binding.get("rootLaunchReleased") is True, "ROOT LAUNCH RELEASE REQUIRED"
PLAN_PATHS=list(binding["plans"])
sha=lambda b:hashlib.sha256(b).hexdigest()
def git(*a):return subprocess.check_output(["git",*a],cwd=W)
def snap():
 names=sorted(set(git("ls-files","-z").decode().split("\0")) | set(binding["reviewedSourceHashes"]))
 return {"head":git("rev-parse","HEAD").decode().strip(),"tree":git("rev-parse","HEAD^{tree}").decode().strip(),"status":git("status","--porcelain").decode(),"diffSha256":sha(git("diff","--binary","HEAD")),"trackedFiles":{n:sha((W/n).read_bytes()) for n in names if n and (W/n).is_file()},"promptSha256":sha((P/"prompt.txt").read_bytes()),"plans":{p:sha(Path(p).read_bytes()) for p in PLAN_PATHS}}
def write(name,obj):
 with (P/name).open("x") as f:json.dump(obj,f,indent=2);f.write("\n")
def rows():
 return subprocess.check_output(["ps","-axo","pid=,ppid=,pgid=,etime=,state=,comm="],text=True).splitlines()
assert "[FINAL ROOT FREEZE" not in (P/"prompt.txt").read_text()
before=snap()
assert before["head"]==binding["head"]
binding=json.loads((E/"final-launch-binding.json").read_text())
assert before["tree"]==binding["tree"]
assert git("branch","--show-current").decode().strip()==binding["branch"]
assert before["diffSha256"]==binding["diffSha256"]
assert before["promptSha256"]==binding["promptSha256"]
assert before["plans"]==binding["plans"]
assert all(before["trackedFiles"][p]==h for p,h in binding["reviewedSourceHashes"].items())
assert set(line[3:] for line in before["status"].splitlines())==set(binding["changedPaths"])
assert all(before["trackedFiles"][n]==h for n,h in binding["hashes"].items())
write("source-before.json",before)
T.mkdir(exist_ok=False)
args=["/opt/homebrew/bin/claude","--safe-mode","-p","--tools","","--no-session-persistence"]
env=os.environ.copy();env["TMPDIR"]=str(T)
start=time.monotonic();timed=False;forced=False;seen=set();seen_temp=set()
with (P/"stdout.txt").open("xb") as out,(P/"stderr.txt").open("xb") as err,(P/"processes.jsonl").open("x") as samples:
 proc=subprocess.Popen(args,cwd=W,env=env,stdin=subprocess.PIPE,stdout=out,stderr=err,start_new_session=True)
 write("process.json",{"argv":args,"cwd":str(W),"pid":proc.pid,"processGroup":proc.pid,"started":datetime.datetime.now(datetime.timezone.utc).isoformat(),"timeoutSeconds":900,"promptSha256":before["promptSha256"],"sourceBeforeSha256":sha((P/"source-before.json").read_bytes()),"tempDirectory":str(T),"reviewKind":"supplied-context only; no model override or tools"})
 proc.stdin.write((P/"prompt.txt").read_bytes());proc.stdin.close()
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
after=snap();write("source-after.json",after)
actualrows=rows();live={int(r.split(None,5)[0]):r for r in actualrows if len(r.split(None,5))==6}
remaining=[r for r in actualrows if len(r.split(None,5))==6 and r.split(None,5)[2]==str(proc.pid)]
cleanup={"checkedAt":datetime.datetime.now(datetime.timezone.utc).isoformat(),"pid":proc.pid,"remainingGroup":remaining,"sampledPids":sorted(seen),"sampledPidsStillPresent":{str(n):live[n] for n in seen if n in live},"sampledTemporaryEntries":sorted(seen_temp),"temporaryEntriesAfter":sorted(p.relative_to(T).as_posix() for p in T.rglob("*")),"limitation":"One-second process-group/temp samples can miss short-lived or detached descendants. Temporary contents are retained without reading or printing their content; no product runtime/store is launched."}
write("cleanup.json",cleanup)
result={"exitCode":proc.returncode,"timeout":timed,"forcedTermination":forced,"elapsedSeconds":time.monotonic()-start,"sourceAndPlanUnchanged":before==after,"stdoutSha256":sha((P/"stdout.txt").read_bytes()),"stderrSha256":sha((P/"stderr.txt").read_bytes()),"cleanupSha256":sha((P/"cleanup.json").read_bytes()),"sourceAfterSha256":sha((P/"source-after.json").read_bytes())}
write("exit.json",result);print(json.dumps(result),flush=True)
