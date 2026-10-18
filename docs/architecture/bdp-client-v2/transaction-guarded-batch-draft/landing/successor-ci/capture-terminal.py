"""Root-released one-shot read-only terminal CI capture; never readiness or merge."""
from pathlib import Path
import subprocess,json,hashlib,re,datetime
O=Path(__file__).resolve().parent;D=O.parent.parent;L=D.parent
W=Path('/Users/dbox/repos/wt/vickie/vickie-guarded-batch')
H='8833c9242b7788e843e9eb97a2c99b2277fe75ef';T='057f879132074166f2c0f643a8b45bbd9dd47c6f'
B='9bf601bfb570adae97ce8cdc99caf2cc7a2b34c1';P='9e5f040edab88be364f683098ead72f595f0ab55'
sha=lambda b:hashlib.sha256(b).hexdigest()
def require(ok,msg):
 if not ok:raise RuntimeError(msg)
def git(*a):return subprocess.check_output(['git','-C',str(W),*a],timeout=15)
release=json.loads((O/'root-terminal-capture-release.json').read_text())
require(release.get('released') is True and release['scriptSha256']==sha(Path(__file__).read_bytes()) and release['preparationSha256']==sha((O/'preparation.json').read_bytes()),'Root capture gate closed/changed')
subprocess.run(['python3',str(L/'guarded-action.py'),'/usr/bin/true'],check=True,timeout=40)
cfg=json.loads((O/'preparation.json').read_text());R=O/'terminal';R.mkdir(exist_ok=False)
def write(name,obj):
 with (R/name).open('x') as f:json.dump(obj,f,indent=2);f.write('\n')
def capture(name,endpoint,raw=False):
 argv=['gh','api','repos/gastownhall/bdp/'+endpoint]
 with (R/name).open('xb') as out,(R/(name+'.stderr')).open('xb') as err:
  result=subprocess.run(argv,stdout=out,stderr=err,timeout=90 if raw else 40)
 write(name+'.capture.json',{'argv':argv,'exit':result.returncode,'sha256':sha((R/name).read_bytes()),'bytes':(R/name).stat().st_size,'at':datetime.datetime.now(datetime.timezone.utc).isoformat()})
 require(result.returncode==0,'Capture failed '+name)
 return None if raw else json.loads((R/name).read_text())
try:
 require(git('show','-s','--format=%H %T %P','HEAD').decode().strip()==f'{H} {T} {P}' and git('status','--porcelain')==b'','Local candidate drift')
 require(git('merge-base',B,H).decode().strip()==B,'Base ancestry drift')
 require(set(git('diff','--name-only',B,H).decode().splitlines())==set(cfg['files']),'Feature scope differs')
 for path,digest in cfg['files'].items():require(sha((W/path).read_bytes())==digest,'Source changed '+path)
 pr=capture('pr.json','pulls/135');run=capture('run.json','actions/runs/38809393968');job=capture('job.json','actions/jobs/116678483498');main=capture('main.json','git/ref/heads/main')
 require(pr['head']['sha']==run['head_sha']==H and pr['base']['ref']=='main' and pr['base']['sha']==main['object']['sha']==B and pr['state']=='open','Current PR head/base/main drift')
 require(run['id']==38809393968 and job['id']==116678483498 and job['run_id']==run['id'] and job['head_sha']==H and run['event']=='pull_request' and run['run_attempt']==1,'Wrong CI identity/attempt')
 require(run['status']==job['status']=='completed','CI not terminal; no log request made')
 capture('job.log','actions/jobs/116678483498/logs',True)
 raw=(R/'job.log').read_text();lines=raw.splitlines();shas=[]
 for i,line in enumerate(lines[:-1]):
  if '[command]/usr/bin/git log -1 --format=%H' in line:
   match=re.search(r'(?:^|\s)([0-9a-f]{40})\s*$',lines[i+1])
   require(match is not None,'Checkout SHA line unavailable');shas.append(match.group(1))
 require(len(shas)==1,'Expected exactly one actual checkout SHA')
 synthetic=capture('synthetic.json','git/commits/'+shas[0])
 require(synthetic['sha']==shas[0] and synthetic['tree']['sha']==T and [p['sha'] for p in synthetic['parents']]==[B,H],'Tested synthetic tree/parents differ')
 checks=capture('checks.json','commits/'+H+'/check-runs?per_page=100')
 require(checks['total_count']==len(checks['check_runs']) and len(checks['check_runs'])>0,'Unresolved check pagination/empty checks')
 require(all(c['status']=='completed' and c['conclusion']=='success' for c in checks['check_runs']),'Checks not all successful')
 require(run['conclusion']==job['conclusion']=='success' and all(s['status']=='completed' and s['conclusion']=='success' for s in job['steps']),'CI/steps not all successful')
 # Full raw log remains authority; extracted lines are review aids, not invented totals.
 interesting=[line for line in lines if any(word in line for word in ['Test Files','Tests  ','Duration  ','preserves one guarded batch winner','e2e:ready:preflight','Read-cohort evidence gate'])]
 write('identity-outcome.json',{'candidate':H,'tree':T,'base':B,'candidateParent':P,'synthetic':shas[0],'syntheticParents':[B,H],'actualSteps':len(job['steps']),'runConclusion':run['conclusion'],'jobConclusion':job['conclusion'],'rawLogSha256':sha((R/'job.log').read_bytes()),'summaryLinesForHumanReview':interesting,'suiteTotals':None,'humanRawLogReviewRequired':True,'finalCouncilPending':True,'ready':False,'mergeAuthorized':False})
except BaseException as error:
 write('capture-failure.json',{'exceptionClass':type(error).__name__,'reason':str(error),'ready':False,'retryAuthorized':False})
 raise
