"""Root-released supplied-context review only. Never imports product modules."""
from pathlib import Path
import argparse,datetime,hashlib,json,os,signal,subprocess,time
parser=argparse.ArgumentParser()
parser.add_argument("scope",choices=("source-tests",))
args=parser.parse_args()
C=Path(__file__).resolve().parent
P=C/args.scope
W=Path("/Users/dbox/repos/wt/vickie/vickie-minimum-current-resource")
G=Path("/Users/dbox/repos/wt/vickie/bdp-integration-logs/guarded-action.py")
sha=lambda b:hashlib.sha256(b).hexdigest()
def git(*a):return subprocess.check_output(["git",*a],cwd=W)
def write(name,value):
 with (P/name).open("x") as stream:json.dump(value,stream,indent=2);stream.write("\n")
def rows():
 return subprocess.check_output(["ps","-axo","pid=,ppid=,pgid=,etime=,state=,comm="],text=True).splitlines()
def patch():
 value=git("diff","--binary",binding["base"],"--",*binding["ownedPaths"])
 for name in binding["newPaths"]:
  if subprocess.run(["git","cat-file","-e","HEAD:"+name],cwd=W,stdout=subprocess.DEVNULL,stderr=subprocess.DEVNULL).returncode:
   result=subprocess.run(["git","diff","--no-index","--binary","--","/dev/null",name],cwd=W,stdout=subprocess.PIPE)
   if result.returncode!=1:raise RuntimeError("new file diff failed")
   value+=result.stdout
 return sha(value)
def snapshot():
 return {"sourceHashes":{name:sha((W/name).read_bytes()) for name in binding["sourceHashes"]},
         "dependencyHashes":{name:sha((W/name).read_bytes()) for name in binding["dependencyHashes"]},
         "inputHashes":{name:sha(Path(name).read_bytes()) for name in binding["inputHashes"]},
         "patchSha256":patch(),"promptSha256":sha((P/"prompt.txt").read_bytes())}
def context():
 return {"head":git("rev-parse","HEAD").decode().strip(),"tree":git("rev-parse","HEAD^{tree}").decode().strip(),
         "status":git("status","--porcelain").decode(),"note":"Git chronology is separate from immutable source/input bytes."}
assert __debug__, "optimized Python disables release assertions"
binding=json.loads((P/"binding.json").read_text())
release=json.loads((P/"root-release.json").read_text())
assert release.get("released") is True
assert release["scope"]==args.scope
assert release["bindingSha256"]==sha((P/"binding.json").read_bytes())
assert release["runnerSha256"]==sha(Path(__file__).read_bytes())
assert release["promptSha256"]==binding["promptSha256"]
assert binding["timeoutSeconds"]==900
assert not any((P/name).exists() for name in ("process.json","exit.json","stdout.txt","stderr.txt","source-before.json","tmp"))
before=snapshot()
assert before["sourceHashes"]==binding["sourceHashes"]
assert before["dependencyHashes"]==binding["dependencyHashes"]
assert before["inputHashes"]==binding["inputHashes"]
assert before["patchSha256"]==binding["patchSha256"]
assert before["promptSha256"]==binding["promptSha256"]
subprocess.run(["python3",str(G),"/usr/bin/true"],check=True)
write("source-before.json",before);write("head-before.json",context())
T=P/"tmp";T.mkdir()
argv=["/opt/homebrew/bin/claude","--safe-mode","-p","--tools","Read,Glob,Grep","--allowedTools","Read,Glob,Grep","--no-session-persistence"]
env={k:v for k,v in os.environ.items() if not k.startswith("BDP_")}
env["TMPDIR"]=str(T)
start=time.monotonic();timed=False;forced=False;seen=set();temps=set()
with (P/"stdout.txt").open("xb") as out,(P/"stderr.txt").open("xb") as err,(P/"processes.jsonl").open("x") as samples,(P/"prompt.txt").open("rb") as prompt:
 proc=subprocess.Popen(argv,cwd=W,env=env,stdin=prompt,stdout=out,stderr=err,start_new_session=True)
 seen.add(proc.pid)
 write("process.json",{"argv":argv,"cwd":str(W),"pid":proc.pid,"processGroup":proc.pid,
                      "started":datetime.datetime.now(datetime.timezone.utc).isoformat(),"timeoutSeconds":900,
                      "promptSha256":before["promptSha256"],"scope":args.scope,"tempDirectory":str(T),
                      "kind":"supplied-context plus read-only Read/Glob/Grep code review; no model override or product execution"})
 while proc.poll() is None:
  elapsed=time.monotonic()-start
  group=[r for r in rows() if len(r.split(None,5))==6 and r.split(None,5)[2]==str(proc.pid)]
  seen.update(int(r.split(None,5)[0]) for r in group)
  temps.update(p.relative_to(T).as_posix() for p in T.rglob("*"))
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
actualrows=rows()
live={int(r.split(None,5)[0]):r for r in actualrows if len(r.split(None,5))==6}
remaining=[r for r in actualrows if len(r.split(None,5))==6 and r.split(None,5)[2]==str(proc.pid)]
cleanup={"checkedAt":datetime.datetime.now(datetime.timezone.utc).isoformat(),"pid":proc.pid,
         "remainingGroup":remaining,"sampledPids":sorted(seen),
         "sampledPidsStillPresent":{str(n):live[n] for n in seen if n in live},
         "sampledTemporaryEntries":sorted(temps),
         "temporaryEntriesAfter":sorted(p.relative_to(T).as_posix() for p in T.rglob("*")),
         "limitation":"One-second group/temp samples can miss short-lived or detached processes. Temp content is retained, not inspected or printed. No product runtime is invoked."}
write("cleanup.json",cleanup)
try:
 after=snapshot();write("source-after.json",after);unchanged=before==after;sourceError=None
except Exception as error:
 unchanged=False;sourceError=type(error).__name__;write("source-after-error.json",{"errorType":sourceError})
write("head-after.json",context())
result={"exitCode":proc.returncode,"shellEquivalentExit":128-proc.returncode if proc.returncode<0 else proc.returncode,
        "timeout":timed,"forcedTermination":forced,"elapsedSeconds":time.monotonic()-start,
        "sourceAndInputsUnchanged":unchanged,"sourceCheckError":sourceError,
        "stdoutSha256":sha((P/"stdout.txt").read_bytes()),"stderrSha256":sha((P/"stderr.txt").read_bytes()),
        "stdoutBytes":(P/"stdout.txt").stat().st_size,"stderrBytes":(P/"stderr.txt").stat().st_size,
        "cleanupSha256":sha((P/"cleanup.json").read_bytes()),"cleanupClear":not remaining and not cleanup["sampledPidsStillPresent"]}
write("exit.json",result);print(json.dumps(result),flush=True)
