import os,sys,json,time,signal,subprocess,hashlib
from pathlib import Path
E=Path(__file__).resolve().parent;W=Path('/Users/dbox/repos/wt/vickie/vickie-fresh-depth-serving');N='/Users/dbox/.nvm/versions/node/v24.16.0/bin/node';G=E.parents[1]/'guarded-action.py'
name=sys.argv[1];assert name in ['red','green','service-suite','typecheck','format-check']
assert not (E/(name+'.exit.json')).exists()
base=[N,str(W/'node_modules/vitest/vitest.mjs'),'run','packages/server/src/transaction-service.test.ts','--reporter=verbose','--reporter=json','--outputFile='+str(E/(name+'.results.json'))]
if name in ['red','green']:base+=['-t','freezes complete deep sequence observations across fresh execution and retained replay']
if name=='typecheck':base=[N,'/Users/dbox/.cache/node/corepack/v1/pnpm/11.20.0/bin/pnpm.cjs','run','typecheck']
if name=='format-check':base=[N,str(W/'node_modules/@biomejs/biome/bin/biome'),'format','--config-path='+str(W/'biome.json'),'packages/server/src/transaction-service.ts','packages/server/src/transaction-service.test.ts']
env=os.environ.copy();env['PATH']=str(Path(N).parent)+':'+env.get('PATH','');env['CI']='true'
for k in list(env):
 if k.startswith('BDP_') or k in ['NODE_OPTIONS','NODE_PATH']:env.pop(k)
tmp=E/(name+'.tmp');tmp.mkdir();env['TMPDIR']=str(tmp)
def hashes():return {p:hashlib.sha256((W/p).read_bytes()).hexdigest() for p in subprocess.check_output(['git','ls-files'],cwd=W,text=True).splitlines()}
before=hashes();(E/(name+'.started.json')).write_text(json.dumps({'argv':base,'cwd':str(W),'source':before,'limitSeconds':120},indent=2))
start=time.monotonic();seen=set();forced=False
with (E/(name+'.log')).open('x') as log:
 p=subprocess.Popen(base,cwd=W,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
 while True:
  rows=[line.split() for line in subprocess.check_output(['ps','-axo','pid=,pgid='],text=True).splitlines()]
  group=[int(x[0]) for x in rows if len(x)==2 and x[1]==str(p.pid)];seen.update(group)
  if p.poll() is not None:break
  if time.monotonic()-start>115:
   forced=True;os.killpg(p.pid,signal.SIGTERM)
   try:p.wait(timeout=4)
   except subprocess.TimeoutExpired:os.killpg(p.pid,signal.SIGKILL);p.wait()
   break
  time.sleep(.05)
 rows=[line.split() for line in subprocess.check_output(['ps','-axo','pid=,pgid='],text=True).splitlines()]
 remaining=[int(x[0]) for x in rows if len(x)==2 and x[1]==str(p.pid)]
 if remaining:
  forced=True;os.killpg(p.pid,signal.SIGKILL)
 after=hashes();present={int(x[0]) for x in rows if x};dirs=[str(x) for x in tmp.rglob('*') if x.is_dir() and x.name.startswith('bdp-')]
 result={'exit':p.returncode,'elapsedSeconds':time.monotonic()-start,'forcedCleanup':forced,'sourceUnchanged':before==after,'sampledPids':sorted(seen),'sampledPidsStillPresent':sorted(seen&present),'remainingGroup':remaining,'runtimeDirectoriesStillPresent':dirs}
 (E/(name+'.exit.json')).write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result))
 sys.exit(0 if p.returncode==(1 if name=='red' else 0) and not forced and before==after and not remaining and not dirs else 1)
