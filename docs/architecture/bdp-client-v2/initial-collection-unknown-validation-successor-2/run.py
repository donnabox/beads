"""Single bounded stage. Usage: run.py final STAGE; existing AW tools are independently pinned. Closed without root gate."""
import os,sys
if len(sys.argv)!=3 or sys.argv[1]!='final':raise SystemExit('usage: run.py final STAGE')
os.environ['COLLECTION_VALIDATION_PHASE']=sys.argv[1]
import sys
sys.dont_write_bytecode=True
from gate import *
import time,signal
name=sys.argv[2];g,spec=load_gate(name);E.mkdir(exist_ok=True)
for suffix in ['.started.json','.log','.exit.json','.processes.jsonl','.tmp']:
 require(not (E/(name+suffix)).exists(),'Attempt exists: preserve, do not retry')
env=os.environ.copy()
for k in list(env):
 if k.startswith('BDP_') or k in ['NODE_OPTIONS','NODE_PATH','NODE_DEBUG','PYTHONOPTIMIZE','BIOME_BINARY','npm_execpath','npm_node_execpath','FORCE_COLOR']:env.pop(k)
env.update({'PATH':'/Users/dbox/.nvm/versions/node/v24.16.0/bin:'+env.get('PATH',''),'CI':'true','NO_COLOR':'1','PYTHONDONTWRITEBYTECODE':'1','COLLECTION_VALIDATION_PHASE':MODE})
env.update(spec.get('environment',{}));tmp=E/(name+'.tmp');tmp.mkdir();env['TMPDIR']=str(tmp);env['npm_config_cache']=str(tmp/'npm-cache')
write(name+'.started.json',{'argv':spec['argv'],'cwd':spec['cwd'],'head':g['finalHead'],'tree':g['finalTree'],'sourceIdentity':g['_sourceIdentity'],'rootGateSha256':sha((P/('bootstrap-gate.json' if MODE=='bootstrap' else 'execution-gate.json')).read_bytes()),'outerSeconds':120,'automaticRetries':0,'overrides':spec.get('environment',{})})
env['COLLECTION_VALIDATION_STAGE']=name;env['COLLECTION_VALIDATION_ATTEMPT']=sha((E/(name+'.started.json')).read_bytes())
start=time.monotonic();p=None;seen=set();forced=False;errors=[]
def rows():
 return [r.split(None,3) for r in subprocess.check_output(['ps','-axo','pid=,ppid=,pgid=,comm='],text=True,timeout=2).splitlines() if len(r.split(None,3))==4]
def members():return [r for r in rows() if p is not None and r[2]==str(p.pid)]
def stop():
 global forced
 if p is None:return
 forced=True;require(p.pid!=os.getpgrp(),'Refuse supervisor group')
 try:os.killpg(p.pid,signal.SIGTERM)
 except ProcessLookupError:return
 end=min(start+119,time.monotonic()+4)
 while time.monotonic()<end:
  p.poll()
  if not members():break
  time.sleep(.05)
 if members():
  try:os.killpg(p.pid,signal.SIGKILL)
  except ProcessLookupError:pass
 try:p.wait(timeout=max(.01,min(1,start+120-time.monotonic())))
 except subprocess.TimeoutExpired:errors.append('leader not reaped')
def interrupted(sig,frame):raise KeyboardInterrupt('Supervisor interrupted')
old={s:signal.signal(s,interrupted) for s in [signal.SIGTERM,signal.SIGINT]}
try:
 with (E/(name+'.log')).open('x') as log,(E/(name+'.processes.jsonl')).open('x') as samples:
  p=subprocess.Popen(spec['argv'],cwd=spec['cwd'],env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True);seen.add(p.pid);next_note=30
  while True:
   group=members();seen.update(int(r[0]) for r in group);elapsed=time.monotonic()-start
   samples.write(json.dumps({'seconds':elapsed,'group':group})+'\n');samples.flush()
   if p.poll() is not None:break
   if elapsed>=115:errors.append('deadline');stop();break
   if elapsed>=next_note:
    print(json.dumps({'stage':name,'seconds':elapsed,'logBytes':(E/(name+'.log')).stat().st_size}),flush=True);next_note+=30
   time.sleep(.05)
except BaseException as e:
 errors.append(type(e).__name__+': '+str(e))
 try:stop()
 except BaseException as x:errors.append('stop:'+type(x).__name__)
finally:
 for s in old:signal.signal(s,signal.SIG_IGN)
 alive=[];unchanged=False;runtime=[]
 try:
  if members():stop()
  alive=[r for r in rows() if int(r[0]) in seen or (p is not None and r[2]==str(p.pid))]
 except BaseException as e:errors.append('process verification:'+type(e).__name__)
 try:
  unchanged=source_check(g['_sourceBinding'])==g['_sourceIdentity']
  if MODE=='final' and name not in ['fresh-build','bind-fresh-build']:verify_build(g)
  if name in ['collection-external','review-collection-external']:verify_consumer(g)
 except BaseException as e:errors.append('final binding:'+str(e))
 try:runtime=[str(x) for x in tmp.rglob('*') if x.is_dir() and x.name.startswith('bdp-')]
 except BaseException as e:errors.append('temp verification:'+type(e).__name__)
 code=p.returncode if p is not None else None
 if time.monotonic()-start>=120:errors.append('outer elapsed exceeded120s')
 accepted=code==0 and not errors and not forced and not alive and not runtime and unchanged
 log=E/(name+'.log')
 record={'exit':code,'accepted':accepted,'sourceIdentity':g['_sourceIdentity'],'sourceUnchanged':unchanged,'elapsedSeconds':time.monotonic()-start,'forcedCleanup':forced,'errors':errors,'sampledPids':sorted(seen),'remainingProcesses':alive,'runtimeDirectoriesStillPresent':runtime,'logSha256':sha(log.read_bytes()) if log.exists() else None}
 write(name+'.exit.json',record)
 print(json.dumps(record));sys.exit(0 if accepted else 1)
