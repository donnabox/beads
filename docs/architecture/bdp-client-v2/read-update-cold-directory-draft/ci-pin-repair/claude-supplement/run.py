from pathlib import Path
import subprocess,json,hashlib,datetime,time,os,signal
W=Path("/Users/dbox/repos/wt/vickie/vickie-update-cold-directory")
E=Path("/Users/dbox/repos/wt/vickie/bdp-integration-logs/read-update-cold-directory-draft/ci-pin-repair/claude-supplement")
P=E;T=P/"tmp"
sha=lambda b:hashlib.sha256(b).hexdigest()
def git(*a):return subprocess.check_output(["git",*a],cwd=W)
def snap():
 return {"committedSourceHashes":{p:sha(git("show",binding["reviewHead"]+":"+p)) for p in binding["sourceHashes"]},"workingSourceHashes":{p:sha((W/p).read_bytes()) for p in binding["sourceHashes"]},"head":git("rev-parse","HEAD").decode().strip(),"status":git("status","--porcelain").decode(),"promptSha256":sha((P/"prompt.txt").read_bytes()),"plans":{p:sha(Path(p).read_bytes()) for p in binding["plans"]}}
def write(name,obj):
 with (P/name).open("x") as f:json.dump(obj,f,indent=2);f.write("\n")
def rows():
 return subprocess.check_output(["ps","-axo","pid=,ppid=,pgid=,etime=,state=,comm="],text=True).splitlines()
binding=json.loads((E/"final-launch-binding.json").read_text())
release=json.loads((E/"root-release.json").read_text())
assert binding.get("rootLaunchReleased") is True and release.get("released") is True
assert release["promptSha256"] == binding["promptSha256"]
assert release["bindingSha256"] == sha((E/"final-launch-binding.json").read_bytes())
assert release["runnerSha256"] == sha(Path(__file__).read_bytes())
assert release["reviewHead"] == binding["reviewHead"]
assert release["reviewTree"] == binding["reviewTree"]
assert git("rev-parse",binding["reviewHead"]+"^{tree}").decode().strip() == binding["reviewTree"]
assert git("rev-parse","HEAD").decode().strip() == binding["reviewHead"]
assert set(git("diff","--name-only",binding["reviewHead"]).decode().splitlines()) == {"apps/bdptest/src/read-update-matrix.test.ts"}
assert sha(git("diff","--binary",binding["reviewHead"])) == binding["fullDiffSha256"]
assert not (E/"process.json").exists() and not (E/"exit.json").exists()
subprocess.run(["python3","/Users/dbox/repos/wt/vickie/bdp-integration-logs/guarded-action.py","/usr/bin/true"],check=True)
before=snap()
assert before["committedSourceHashes"]==binding["sourceHashes"]
assert before["workingSourceHashes"]==binding["workingSourceHashes"]
assert before["plans"]==binding["plans"]
assert before["promptSha256"]==binding["promptSha256"]
write("source-before.json",before)
write("head-before.json",{"head":git("rev-parse","HEAD").decode().strip(),"tree":git("rev-parse","HEAD^{tree}").decode().strip(),"status":git("status","--porcelain").decode(),"limitation":"Exact original committed candidate plus one bound working-tree CPU expectation repair; ignored outputs are not product source."})
T.mkdir(exist_ok=False)
args=["/opt/homebrew/bin/claude","--safe-mode","-p","--tools","","--no-session-persistence"]
env=os.environ.copy();env["TMPDIR"]=str(T)
start=time.monotonic();timed=False;forced=False;seen=set();seen_temp=set()
with (P/"stdout.txt").open("xb") as out,(P/"stderr.txt").open("xb") as err,(P/"processes.jsonl").open("x") as samples,(P/"prompt.txt").open("rb") as prompt:
 proc=subprocess.Popen(args,cwd=W,env=env,stdin=prompt,stdout=out,stderr=err,start_new_session=True)
 write("process.json",{"argv":args,"cwd":str(W),"pid":proc.pid,"processGroup":proc.pid,"started":datetime.datetime.now(datetime.timezone.utc).isoformat(),"timeoutSeconds":900,"promptSha256":before["promptSha256"],"sourceBeforeSha256":sha((P/"source-before.json").read_bytes()),"tempDirectory":str(T),"reviewKind":"supplied-context one-hash test supplement; no model override or tools"})
 while proc.poll() is None:
  elapsed=time.monotonic()-start
  group=[r for r in rows() if len(r.split(None,5))==6 and r.split(None,5)[2]==str(proc.pid)]
  seen.update(int(r.split(None,5)[0]) for r in group)
  seen_temp.update(p.relative_to(T).as_posix() for p in T.rglob("*"))
  samples.write(json.dumps({"elapsedSeconds":elapsed,"rows":group})+"\n");samples.flush()
  if elapsed>=895:
   timed=True;forced=True
   try:os.killpg(proc.pid,signal.SIGTERM)
   except ProcessLookupError:pass
   try:proc.wait(timeout=max(0.1,900-(time.monotonic()-start)))
   except subprocess.TimeoutExpired:
    try:os.killpg(proc.pid,signal.SIGKILL)
    except ProcessLookupError:pass
    proc.wait(timeout=5)
   break
  time.sleep(1)
 proc.wait()
after=snap();write("source-after.json",after)
write("head-after.json",{"head":git("rev-parse","HEAD").decode().strip(),"tree":git("rev-parse","HEAD^{tree}").decode().strip(),"status":git("status","--porcelain").decode()})
actualrows=rows();live={int(r.split(None,5)[0]):r for r in actualrows if len(r.split(None,5))==6}
remaining=[r for r in actualrows if len(r.split(None,5))==6 and r.split(None,5)[2]==str(proc.pid)]
cleanup={"checkedAt":datetime.datetime.now(datetime.timezone.utc).isoformat(),"pid":proc.pid,"remainingGroup":remaining,"sampledPids":sorted(seen),"sampledPidsStillPresent":{str(n):live[n] for n in seen if n in live},"sampledTemporaryEntries":sorted(seen_temp),"temporaryEntriesAfter":sorted(p.relative_to(T).as_posix() for p in T.rglob("*")),"limitation":"One-second process-group/temp samples can miss short-lived or detached descendants. Temporary contents are retained without reading or printing their content; no product runtime/store is launched."}
write("cleanup.json",cleanup)
result={"exitCode":proc.returncode,"timeout":timed,"forcedTermination":forced,"elapsedSeconds":time.monotonic()-start,"sourceAndPlanUnchanged":before==after,"stdoutSha256":sha((P/"stdout.txt").read_bytes()),"stderrSha256":sha((P/"stderr.txt").read_bytes()),"cleanupSha256":sha((P/"cleanup.json").read_bytes()),"sourceAfterSha256":sha((P/"source-after.json").read_bytes())}
write("exit.json",result);print(json.dumps(result),flush=True)
