"""One ledger command per invocation; adapted from maintenance supervise-v4.py.
Not executed during preparation. No retry loop. Python3 supervise.py COMMAND_ID.
"""
import os,sys,subprocess,time,json,signal,datetime
from pathlib import Path
from gate import P,W,E,sha,load_gate
name=sys.argv[1]
ledger=json.loads((P/'commands.json').read_text());spec=ledger['commands'][name]
g=load_gate(spec['phase'] in ('frozen','installed'))
assert name in g['approvedCommands'],name
E.mkdir(exist_ok=True)
assert not (E/(name+'.started.json')).exists(),'Attempt already exists; no automatic retry'
if spec['phase']=='installed':
 assert g['newControlReportsResolved'] is True,'Complete new-control reports gate unresolved'
 provenance=json.loads((E/'installed-provenance.json').read_text())
 assert sha((E/'installed-provenance.json').read_bytes())==g['installedProvenanceSha256']
 assert provenance['sourceCommit']==g['finalHead']
 entry=Path(provenance['entryRealpath']).resolve(strict=True)
 assert entry.is_relative_to((E/'consumer').resolve())
 assert sha(entry.read_bytes())==provenance['entrySha256']
 # Every previous attempted cohort needs an accepted result, not merely exit zero.
 for earlier in ledger['cohortOrder'][:ledger['cohortOrder'].index(name)]:
  a=json.loads((E/(earlier+'.adjudication.json')).read_text())
  assert a['accepted'] is True and a['sourceCommit']==g['finalHead']
  assert a['attemptedHttp']<=192 and a['childLaunches']<=21
 envextra={'BDP_TRANSACTIONAL_INSTALLED_ENTRY':str(entry),'BDP_TRANSACTIONAL_SOURCE_COMMIT':g['finalHead']}
else: envextra={}
allowed={'BDP_TRANSACTIONAL_OBSERVATION_OUTPUT','BDP_TRANSACTIONAL_NOOP_CONTROL_OUTPUT','BDP_TRANSACTIONAL_DELETION_CONTROL_OUTPUT','BDP_TRANSACTIONAL_DELETION_PREFLIGHT_OUTPUT','BDP_TRANSACTIONAL_CROSS_CARRIER_CONTROL_OUTPUT'}
for key,value in spec.get('outputEnv',{}).items():
 assert key in allowed,key
 assert Path(value).resolve().is_relative_to(E.resolve()),value
 envextra[key]=value
# Exact root-accepted staged writer; source application/final hashes remain a gate.
env=os.environ.copy()
for key in list(env):
 if key.startswith('BDP_'):env.pop(key)
env.update(envextra);env['PATH']='/Users/dbox/.nvm/versions/node/v24.16.0/bin:'+env['PATH'];env['CI']='true'
tmp=E/(name+'.tmp');tmp.mkdir(exist_ok=False);env['TMPDIR']=str(tmp)
cmd=spec['argv'];cwd=spec['cwd'];started=datetime.datetime.now(datetime.timezone.utc).isoformat()
(E/(name+'.started.json')).write_text(json.dumps({'argv':cmd,'cwd':cwd,'started':started,'sourceSha256':g['sourceSha256'],'finalHead':g.get('finalHead'),'bdpOverrides':envextra},indent=2)+'\n')
t=time.monotonic();timed=False;forced=False;seen=set();seenpids=set();errors=[];p=None

def rows():
 return subprocess.check_output(['ps','-axo','pid,ppid,pgid,rss,etime,state,command'],text=True).splitlines()
def group(snapshot):return [r for r in snapshot if len(r.split())>2 and r.split()[2]==str(p.pid)]
def stop_group():
 global forced
 forced=True
 try:os.killpg(p.pid,signal.SIGTERM)
 except ProcessLookupError:return
 end=time.monotonic()+5
 while time.monotonic()<end and group(rows()):time.sleep(.05)
 if group(rows()):
  try:os.killpg(p.pid,signal.SIGKILL)
  except ProcessLookupError:pass
try:
 with (E/(name+'.log')).open('x') as log,(E/(name+'.processes.jsonl')).open('x') as pslog:
  p=subprocess.Popen(cmd,cwd=cwd,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
  next_progress=30
  while p.poll() is None:
   elapsed=time.monotonic()-t;selected=group(rows())
   seenpids.update(int(r.split()[0]) for r in selected)
   pslog.write(json.dumps({'elapsed':elapsed,'rows':selected})+'\n');pslog.flush()
   seen.update(str(x) for x in tmp.iterdir())
   if elapsed>=next_progress:
    print(json.dumps({'command':name,'elapsedSeconds':elapsed,'logBytes':(E/(name+'.log')).stat().st_size,'processes':len(selected)}),flush=True);next_progress+=30
   if elapsed>120:timed=True;stop_group();break
   time.sleep(.05)
  code=p.wait(timeout=5)
except BaseException as ex:
 errors.append(type(ex).__name__);code=1
 if p is not None:stop_group();p.wait(timeout=5)
finally:
 before=group(rows()) if p else []
 if before:stop_group()
 after=group(rows()) if p else []
 snapshot=rows();live={int(r.split()[0]):r for r in snapshot[1:] if r.split()}
 result={'name':name,'argv':cmd,'cwd':cwd,'started':started,'elapsedSeconds':time.monotonic()-t,'exit':code,'timeout':timed,'forcedCleanup':forced,'errors':errors,'processGroup':p.pid if p else None,'remainingBeforeCleanup':before,'remainingAfterCleanup':after,'sampledPids':sorted(seenpids),'sampledPidsStillPresent':{str(i):live[i] for i in seenpids if i in live},'sampledTempEntries':sorted(seen),'tempEntriesAfter':sorted(str(x) for x in tmp.iterdir()),'sampledTempEntriesStillPresent':[x for x in sorted(seen) if Path(x).exists()],'samplingSeconds':.05,'samplingLimitation':'Short-lived children may be missed; test-derived launch/close counts are separately adjudicated.'}
 result['remainingTempEntriesForClassification']=sorted(set(result['sampledTempEntriesStillPresent']) | set(result['tempEntriesAfter']))
 result['runtimeDirectoriesStillPresent']=[x for x in result['remainingTempEntriesForClassification'] if Path(x).name.startswith('bdp-tx-installed-')]
 result['unclassifiedTempEntriesStillPresent']=[x for x in result['remainingTempEntriesForClassification'] if x not in result['runtimeDirectoriesStillPresent']]
 result['tempClassification']='Only the selected observe() mkdtemp prefix is a known service/store root. Other remaining entries are preserved, not deleted or automatically called leaks; review before dependent execution.'
 (E/(name+'.exit.json')).write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result),flush=True)
sys.exit(124 if timed else 1 if forced or errors or after or result['runtimeDirectoriesStillPresent'] else code)
