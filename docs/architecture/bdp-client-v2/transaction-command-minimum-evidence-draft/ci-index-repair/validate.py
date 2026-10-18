from pathlib import Path
import subprocess,json,hashlib,os,time,signal,re,datetime
p=Path(__file__).resolve().parent;w=Path('/Users/dbox/repos/wt/vickie/vickie-command-minimum-evidence');node='/Users/dbox/.nvm/versions/node/v24.16.0/bin/node';pnpm='/Users/dbox/.cache/node/corepack/v1/pnpm/11.20.0/bin/pnpm.cjs'
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest();f=w/'packages/conformance/src/transactional-cumulative.test.ts';assert sha(f)=='22a568820ab2abc187d37c27658fdcf6e5ad39e74e2172374e3658c180f7668e'
names=['pins the one-page matched-plus-updated witness without advertised-limit or other-outcome credit','binds the canonical cumulative TX selection and exact source digests without qualification','pins the inherited command-minimum manifest facts, read-header absence and all thirteen prior objects','preserves every original TX scenario and coverage object while adding one bounded partial cross-carrier row']
commands=[('typecheck',[node,pnpm,'typecheck']),('format-check',[node,pnpm,'format:check']),('four-static',[node,str(w/'node_modules/vitest/vitest.mjs'),'run',str(f.relative_to(w)),'--testNamePattern','^(?:'+'|'.join(re.escape(x) for x in names)+')$','--maxWorkers=1','--reporter=verbose','--reporter=json','--outputFile.json='+str(p/'four-static.results.json')])]
env={k:v for k,v in os.environ.items() if not k.startswith('BDP_') and k not in ['NODE_OPTIONS','NODE_PATH']};env['PATH']=str(Path(node).parent)+':'+env.get('PATH','');tmp=p/'tmp';tmp.mkdir(exist_ok=False);env['TMPDIR']=str(tmp)
summary=[]
for name,argv in commands:
 start=time.monotonic()
 with (p/(name+'.log')).open('xb') as out:
  proc=subprocess.Popen(argv,cwd=w,env=env,stdout=out,stderr=subprocess.STDOUT,start_new_session=True)
  (p/(name+'.started.json')).write_text(json.dumps({'argv':argv,'pid':proc.pid,'source':sha(f),'at':datetime.datetime.now(datetime.timezone.utc).isoformat()},indent=2)+'\n')
  timeout=False
  try:code=proc.wait(timeout=120)
  except subprocess.TimeoutExpired:
   timeout=True;os.killpg(proc.pid,signal.SIGTERM)
   try:code=proc.wait(timeout=5)
   except subprocess.TimeoutExpired:os.killpg(proc.pid,signal.SIGKILL);code=proc.wait()
 group=[x for x in subprocess.check_output(['ps','-axo','pid=,pgid='],text=True).splitlines() if len(x.split())==2 and x.split()[1]==str(proc.pid)]
 result={'command':name,'exit':code,'timeout':timeout,'seconds':time.monotonic()-start,'remainingGroup':group,'sourceUnchanged':sha(f)=='22a568820ab2abc187d37c27658fdcf6e5ad39e74e2172374e3658c180f7668e','logSha256':sha(p/(name+'.log'))};(p/(name+'.exit.json')).write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result),flush=True);summary.append(result)
 assert code==0 and not timeout and not group and result['sourceUnchanged']
r=json.loads((p/'four-static.results.json').read_text());actual=[a['fullName'] for s in r['testResults'] for a in s['assertionResults'] if a['status']=='passed'];assert sorted(actual)==sorted(names) and r['success'] and r['numFailedTests']==0
(p/'validation-summary.json').write_text(json.dumps({'commands':summary,'passed':actual,'skipped':r['numPendingTests'],'source':sha(f),'installedRerun':False,'fullCiPending':True},indent=2)+'\n')
