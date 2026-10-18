"""Finite two-test repair validation. Closed until exact root-bound successor release."""
from pathlib import Path
import os,sys,json,time,signal,subprocess,hashlib,re
E=Path(__file__).resolve().parent
W=Path('/Users/dbox/repos/wt/vickie/vickie-fresh-depth-serving')
G=E.parents[1]/'guarded-action.py'
sha=lambda b:hashlib.sha256(b).hexdigest()
def need(ok,message):
 if not ok:raise RuntimeError(message)
def git(*a):return subprocess.check_output(['git',*a],cwd=W,text=True,timeout=10)
def sources():return {p:sha((W/p).read_bytes()) for p in git('ls-files').splitlines()}
name=sys.argv[1];gate=json.loads((E/'execution-gate.json').read_text());cfg=json.loads((E/'commands.json').read_text());order=cfg['order']
need(gate['released'] is True and gate['rootCouncilAccepted'] is True,'Root release/council absent')
need(name in order and gate['approvedStages']==order[:len(gate['approvedStages'])] and name in gate['approvedStages'],'Stage not released')
need(set(gate['preparationHashes'])==set(json.loads((E/'preparation-hashes.json').read_text())['files']),'Complete preparation pins required')
for p,h in gate['preparationHashes'].items():need(sha((E/p).read_bytes())==h,'Preparation changed '+p)
for p,h in gate['toolHashes'].items():need(sha(Path(p).read_bytes())==h,'Tool changed '+p)
need(git('rev-parse','HEAD').strip()==gate['head'],'HEAD changed')
baseline=json.loads((E/'baseline-source.json').read_text());owned=set(gate['files']);need(owned=={'packages/server/src/transaction-set.test.ts','packages/server/src/transaction-submission.test.ts'},'Scope differs')
before=sources();need(set(before)==set(baseline),'Tracked inventory changed')
need(all(before[p]==(gate['files'][p] if p in owned else baseline[p]) for p in before),'Unexpected source delta')
need(not git('ls-files','--others','--exclude-standard').strip(),'Unexpected untracked repository files')
need(set(git('diff','--name-only','HEAD').splitlines())<=owned,'Unexpected dirty scope')
for p,h in json.loads((E/'payload-bridge.json').read_text())['priorArtifacts'].items():need(sha(Path(p).read_bytes())==h,'Historical bridge artifact changed')
for prior in order[:order.index(name)]:
 x=json.loads((E/(prior+'.exit.json')).read_text());need(x['accepted'] and x['sourceAfter']==before,'Prior stage failed/source differs')
helper=gate['helperTitle'];need(isinstance(helper,str) and helper and helper not in cfg['focusExisting'],'Helper title pending')
argv=list(cfg['commands'][name])
if name=='focused-three':argv[-1]='^(?:'+'|'.join(re.escape(x) for x in [*cfg['focusExisting'],helper])+')$'
for suffix in ['.started.json','.log','.results.json','.exit.json','.processes.jsonl','.tmp']:need(not (E/(name+suffix)).exists(),'Attempt exists; preserve, do not retry')
# Guard remote role before any output mutation or product invocation.
subprocess.run([sys.executable,str(G),'/usr/bin/true'],check=True)
env=os.environ.copy()
for k in list(env):
 if k.startswith('BDP_') or k in ['NODE_OPTIONS','NODE_PATH','NODE_DEBUG','PYTHONOPTIMIZE']:env.pop(k)
env['PATH']='/Users/dbox/.nvm/versions/node/v24.16.0/bin:'+env.get('PATH','');env['CI']='true'
tmp=E/(name+'.tmp');tmp.mkdir();env['TMPDIR']=str(tmp)
(E/(name+'.started.json')).write_text(json.dumps({'argv':argv,'cwd':str(W),'head':gate['head'],'sourceBefore':before,'rootReleaseSha256':sha((E/'execution-gate.json').read_bytes()),'timeoutSeconds':120},indent=2)+'\n')
start=time.monotonic();seen=set();forced=False;errors=[];p=None

def rows():return [r.split(None,3) for r in subprocess.check_output(['ps','-axo','pid=,ppid=,pgid=,comm='],text=True,timeout=5).splitlines() if len(r.split(None,3))==4]
def stop():
 global forced
 forced=True
 if p is None:return
 try:os.killpg(p.pid,signal.SIGTERM)
 except ProcessLookupError:return
 try:p.wait(timeout=max(.1,min(4,120-(time.monotonic()-start))))
 except subprocess.TimeoutExpired:
  try:os.killpg(p.pid,signal.SIGKILL)
  except ProcessLookupError:pass
  p.wait(timeout=1)
def interrupted(_signal,_frame):raise KeyboardInterrupt('Supervisor interrupted')
old_signals={sig:signal.signal(sig,interrupted) for sig in (signal.SIGTERM,signal.SIGINT)}
try:
 with (E/(name+'.log')).open('x') as log,(E/(name+'.processes.jsonl')).open('x') as samples:
  p=subprocess.Popen(argv,cwd=W,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True);seen.add(p.pid);next_note=30
  while True:
   group=[r for r in rows() if r[2]==str(p.pid)];seen.update(int(r[0]) for r in group);elapsed=time.monotonic()-start
   samples.write(json.dumps({'seconds':elapsed,'group':group})+'\n');samples.flush()
   if p.poll() is not None:break
   if elapsed>=115:stop();errors.append('deadline');break
   if elapsed>=next_note:print(json.dumps({'stage':name,'seconds':elapsed,'logBytes':(E/(name+'.log')).stat().st_size,'sampledGroup':len(group)}),flush=True);next_note+=30
   time.sleep(.05)
except BaseException as exc:errors.append(type(exc).__name__);stop()
finally:
 for sig in old_signals:signal.signal(sig,signal.SIG_IGN)
 current=rows();remaining=[r for r in current if p is not None and r[2]==str(p.pid)]
 if remaining:stop()
 current=rows();alive=[r for r in current if int(r[0]) in seen or (p is not None and r[2]==str(p.pid))]
 after=sources();runtime=[str(x) for x in tmp.rglob('*') if x.is_dir() and x.name.startswith('bdp-')]
 selected=[];skipped=None
 try:
  if name in ['focused-three','full-two-files']:
   result=json.loads((E/(name+'.results.json')).read_text());tests=[(str(Path(f['name']).resolve().relative_to(W)),a) for f in result['testResults'] for a in f.get('assertionResults',[])]
   executed=[(f,a) for f,a in tests if a['status'] not in ['pending','skipped','todo']];selected=[(f,a['fullName']) for f,a in executed];skipped=len(tests)-len(executed)
   expected=[(x['file'],x['fullName']) for x in json.loads((E/'original-full-selection.json').read_text())]+[('packages/server/src/transaction-submission.test.ts',helper)]
   if name=='focused-three':expected=[x for x in expected if x[1] in [*cfg['focusExisting'],helper]]
   need(sorted(selected)==sorted(expected) and all(a['status']=='passed' for _,a in executed),'Exact selected cases failed/missing')
   need(result['success'] and result['numFailedTests']==0 and result['numPassedTests']==len(expected) and not result.get('unhandledErrors') and not result.get('errors') and not result.get('numRuntimeErrorTestSuites',0),'Test aggregate failed')
 except BaseException as exc:errors.append('adjudication:'+type(exc).__name__)
 code=p.returncode if p is not None else None
 accepted=code==0 and not forced and not errors and not alive and not runtime and before==after and git('rev-parse','HEAD').strip()==gate['head']
 record={'exit':code,'accepted':accepted,'elapsedSeconds':time.monotonic()-start,'forcedCleanup':forced,'errors':errors,'sourceUnchanged':before==after,'sourceAfter':after,'sampledPids':sorted(seen),'remainingProcesses':alive,'runtimeDirectoriesStillPresent':runtime,'otherTempEntries':[str(x) for x in tmp.iterdir()],'selectedCases':selected,'skipped':skipped,'logSha256':sha((E/(name+'.log')).read_bytes()) if (E/(name+'.log')).exists() else None}
 with (E/(name+'.exit.json')).open('x') as f:f.write(json.dumps(record,indent=2)+'\n')
 print(json.dumps({k:v for k,v in record.items() if k!='sourceAfter'}));sys.exit(0 if accepted else 1)
