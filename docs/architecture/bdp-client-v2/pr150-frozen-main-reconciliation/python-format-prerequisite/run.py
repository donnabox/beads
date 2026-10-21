from pathlib import Path
import datetime,hashlib,json,os,signal,subprocess,sys,time
P=Path(__file__).parent;I=P.parent;W=Path('/Users/dbox/repos/wt/vickie/vickie-pr150-main-reconciliation')
PY='/opt/homebrew/opt/python@3.14/bin/python3.14';guard='/Users/dbox/repos/wt/vickie/bdp-integration-logs/guarded-action.py'
sha=lambda p:hashlib.sha256(Path(p).read_bytes()).hexdigest()
def write(path,value):
    with path.open('x') as f:json.dump(value,f,indent=2);f.write('\n')
def inventory(root):
    files={};links={}
    if root.exists():
        for parent,dirs,names in os.walk(root,followlinks=False):
            for name in dirs+names:
                q=Path(parent)/name;r=str(q.relative_to(root))
                if q.is_symlink():
                    links[r]={'target':os.readlink(q),'resolved':str(q.resolve()),'resolvedFileSha256':sha(q) if q.is_file() else None}
                elif q.is_file():files[r]=sha(q)
    return {'root':str(root),'files':files,'symlinks':links}
def git(*args):return subprocess.check_output(['git',*args],cwd=W,text=True).strip()
def snapshot(binding):
    return {'head':git('rev-parse','HEAD'),'tree':git('rev-parse','HEAD^{tree}'),'status':git('status','--porcelain=v1','--untracked-files=all'),'source':{f:sha(W/f) for f in binding['source']},'oldEnvironment':inventory(I/'tools/python-venv'),'basePython':{'path':PY,'realpath':str(Path(PY).resolve()),'sha256':sha(PY)}}
def check_inputs():
    index=json.loads((P/'preparation-index.json').read_text())
    for f,digest in index['files'].items():assert sha(P/f)==digest,f
    binding=json.loads((P/'binding.json').read_text())
    for f,digest in binding['references'].items():assert sha(f)==digest,f
    actual=snapshot(binding)
    for key in ['head','tree','status','source','oldEnvironment','basePython']:assert actual[key]==binding[key],key
    assert set(git('ls-files').splitlines())==set(binding['source'])
    return actual
assert len(sys.argv)==2 and sys.argv[1] in ['provision','probe'];phase=sys.argv[1]
release=json.loads((P/'root-release.json').read_text())
assert release['released'] is True and phase in release['approvedPhases'],'CLOSED: root release required'
assert release['preparationIndexSha256']==sha(P/'preparation-index.json')
before=check_inputs();target=I/'tools/python-format-venv'
if phase=='provision':assert not target.exists() and not target.is_symlink(),'No retry into partial environment'
else:
    prior=json.loads((P/'provision/exit.json').read_text());assert prior['accepted'] is True
    assert release['provisionExitSha256']==sha(P/'provision/exit.json')
    assert release['environmentSha256']==sha(P/'provision/environment.json')
    assert inventory(target)==json.loads((P/'provision/environment.json').read_text())
subprocess.run([PY,guard,'/usr/bin/true'],check=True)
A=P/phase;A.mkdir();os.umask(0o077)
write(A/'source-tool-before.json',before)
command=[PY if phase=='provision' else str(target/'bin/python'),'-I','-B',str(P/(phase+'.py')),'--released-child']
env=os.environ.copy();env.update({'PYTHONDONTWRITEBYTECODE':'1','PIP_DISABLE_PIP_VERSION_CHECK':'1','PIP_NO_INPUT':'1'})
started=time.monotonic();sampled=set();child=None;timed=False;forced=False;error=None
write(A/'started.json',{'phase':phase,'startedAt':datetime.datetime.now(datetime.timezone.utc).isoformat(),'argv':command,'cwd':str(P),'timeoutSeconds':120,'termAfterSeconds':115,'attempts':1,'validationStage':False})
def processes():
    return [line.split(None,3) for line in subprocess.check_output(['ps','-axo','pid=,ppid=,pgid=,comm='],text=True).splitlines() if len(line.split(None,3))==4]
def terminate():
    global forced
    forced=True
    try:os.killpg(child.pid,signal.SIGTERM)
    except ProcessLookupError:pass
    deadline=time.monotonic()+2
    while time.monotonic()<deadline:
        child.poll()
        if not any(int(r[2])==child.pid for r in processes()):break
        time.sleep(.05)
    if any(int(r[2])==child.pid for r in processes()):
        try:os.killpg(child.pid,signal.SIGKILL)
        except ProcessLookupError:pass
    try:child.wait(timeout=1)
    except subprocess.TimeoutExpired:pass
with (A/'stdout.log').open('xb') as out,(A/'stderr.log').open('xb') as err,(A/'processes.jsonl').open('x') as progress:
    try:
        child=subprocess.Popen(command,cwd=P,env=env,stdout=out,stderr=err,start_new_session=True)
        write(A/'process.json',{'pid':child.pid,'pgid':child.pid});sampled.add(child.pid);next_sample=0
        while child.poll() is None:
            elapsed=time.monotonic()-started
            if elapsed>=next_sample:
                owned=[r for r in processes() if int(r[2])==child.pid];sampled.update(int(r[0]) for r in owned)
                progress.write(json.dumps({'elapsedSeconds':elapsed,'owned':owned})+'\n');progress.flush();next_sample=elapsed+10
            if elapsed>=115:timed=True;terminate();break
            time.sleep(.2)
    except Exception as exc:
        error=type(exc).__name__+': '+str(exc)
        if child is not None and child.poll() is None:terminate()
    finally:
        if child is not None and any(int(r[2])==child.pid for r in processes()):terminate()
remaining=[r for r in processes() if int(r[0]) in sampled or child is not None and int(r[2])==child.pid]
try:after=check_inputs();unchanged=after==before
except Exception as exc:after={'verificationError':str(exc)};unchanged=False
elapsed=time.monotonic()-started
newenv=inventory(target)
if phase=='probe' and newenv!=json.loads((P/'provision/environment.json').read_text()):unchanged=False
result={'phase':phase,'exitCode':child.returncode if child else None,'elapsedSeconds':elapsed,'timeout':timed or elapsed>=120,'forcedCleanup':forced,'sampledPids':sorted(sampled),'remainingProcesses':remaining,'sourceAndOldToolsUnchanged':unchanged,'error':error,'accepted':bool(child is not None and child.returncode==0 and not timed and elapsed<120 and not forced and not remaining and unchanged and error is None),'stdoutSha256':sha(A/'stdout.log'),'stderrSha256':sha(A/'stderr.log'),'validationStage':False,'preservePartialState':True}
subprocess.run([PY,guard,'/usr/bin/true'],check=True)
write(A/'source-tool-after.json',after);write(A/'environment.json',newenv);write(A/'exit.json',result)
print(json.dumps(result));sys.exit(0 if result['accepted'] else 1)
