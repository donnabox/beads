from pathlib import Path
import subprocess,os,json,time,hashlib
p=Path(__file__).parent
w=Path('/Users/dbox/repos/wt/vickie/vickie-sequence-depth-retained')
b=json.loads((p/'binding.json').read_text())
assert hashlib.sha256((w/b['file']).read_bytes()).hexdigest()==b['sha256']
env=dict(os.environ);env['PATH']='/Users/dbox/.nvm/versions/node/v24.16.0/bin:'+env['PATH']
node='/Users/dbox/.nvm/versions/node/v24.16.0/bin/node';pnpm='/Users/dbox/.cache/node/corepack/v1/pnpm/11.20.0/bin/pnpm.cjs'
steps=[('format',['exec','biome','format',b['file']]),('typecheck',['typecheck']),('selected',['exec','vitest','run',b['file'],'-t','keeps early sequence refusals unobserved and observes deep negotiation refusal','--reporter=verbose'])]
results=[]
for name,args in steps:
 start=time.monotonic()
 with (p/(name+'.log')).open('xb') as out:
  c=subprocess.run([node,pnpm]+args,cwd=w,env=env,stdout=out,stderr=subprocess.STDOUT,timeout=120)
 result={'stage':name,'exitCode':c.returncode,'elapsedSeconds':time.monotonic()-start};results.append(result);print(json.dumps(result),flush=True)
 (p/'validation.json').write_text(json.dumps(results,indent=2))
 if c.returncode:break
assert hashlib.sha256((w/b['file']).read_bytes()).hexdigest()==b['sha256']
