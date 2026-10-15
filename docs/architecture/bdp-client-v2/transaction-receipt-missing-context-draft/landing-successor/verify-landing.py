"""Read-only exact receipt landing verification. Root supplies pins.json after final
source/evidence/CI freeze. Never commits, pushes, makes ready, reruns CI or merges.
Invoke through guarded-action.py. No record overwrites; pre/post outputs distinct."""
from pathlib import Path
import subprocess,json,hashlib,sys,datetime,re
P=Path(__file__).resolve().parent
W=Path('/Users/dbox/repos/wt/vickie/vickie-sequence-member-evidence')
BASE='015cac1aa4fa827cc4a2c647dff08454c6c8e6f8'
EXPECTED_PATHS={
 'packages/server/src/transaction-store.ts','packages/server/src/transaction-service.ts',
 'packages/server/src/transaction-http.ts','packages/server/src/transaction-receipt-pages.test.ts',
 'packages/server/src/transaction-service.test.ts','packages/server/src/transaction-http.test.ts',
 'apps/bdptest/src/transactional-receipt-missing-context-installed.test.ts',
 'docs/design/transaction-implementation-progress.md','apps/bdptest/src/transactional-matrix.test.ts'}
REQUIRED_STEPS=['Run pnpm install --frozen-lockfile','Run pnpm format:check','Run pnpm lint',
 'Run pnpm check:boundaries','Run pnpm typecheck','Run pnpm build','Run pnpm test',
 'Run pnpm e2e:ready:preflight','Run pnpm smoke','Read-cohort evidence gate']
assert __debug__ and sys.argv[1:] in [['premerge'],['postmerge']]
pins=json.loads((P/'pins.json').read_text());mode=sys.argv[1]
def sha(b):return hashlib.sha256(b).hexdigest()
def valid(s,n):return isinstance(s,str) and re.fullmatch('[0-9a-f]{'+str(n)+'}',s)
def git(*a):return subprocess.check_output(['git',*a],cwd=W,timeout=60)
def textgit(*a):return git(*a).decode().strip()
def raw(path):return subprocess.check_output(['gh','api','repos/gastownhall/bdp/'+path],timeout=60)
def api(path):return json.loads(raw(path))
def pin(record):
 data=Path(record['path']).read_bytes();assert valid(record['sha256'],64) and sha(data)==record['sha256'],record['path'];return data
assert pins['rootAccepted'] is True and pins['base']==BASE
head,tree,S=pins['head'],pins['tree'],pins['synthetic']
assert all(valid(v,40) for v in [head,tree,S]) and valid(pins['ciLogSha256'],64)
PR,RUN,JOB=pins['pr'],pins['run'],pins['job']
assert all(type(v) is int and v>0 for v in [PR,RUN,JOB])
assert set(pins['files'])==EXPECTED_PATHS and all(valid(v,64) for v in pins['files'].values())
chain=pins['ordinaryChain'];assert len(chain)>=2 and chain[0]==BASE and chain[-1]==head and len(set(chain))==len(chain)
assert all(valid(v,40) for v in chain)
for parent,child in zip(chain,chain[1:]):assert textgit('show','-s','--format=%P',child)==parent
assert textgit('rev-parse',head+'^{tree}')==tree
assert set(textgit('diff','--name-only',BASE,head).splitlines())==EXPECTED_PATHS
for name,digest in pins['files'].items():assert sha(git('show',head+':'+name))==digest,name
assert not textgit('status','--porcelain')
if mode=='premerge':assert textgit('rev-parse','HEAD')==head
else:assert textgit('rev-parse','HEAD^{tree}')==tree
# Hash actual existing acceptances; root supplies final values, not this preparer.
required={'finalSourceBinding','native','claudeExit','claudeStdout','rootCouncil','sourceAcceptance','controlAcceptance','installedAcceptance','rootLandingAcceptance'}
assert set(pins['evidence'])==required
saved={key:pin(record) for key,record in pins['evidence'].items()}
exit_record=json.loads(saved['claudeExit'])
assert exit_record['exitCode']==0 and not exit_record['timeout'] and not exit_record['forcedTermination'] and exit_record['sourceAndPlanUnchanged'] is True
assert exit_record['stdoutSha256']==sha(saved['claudeStdout']) and saved['claudeStdout'].strip()
# Root acceptance explicitly binds this final ordinary source and these exact artifacts.
accept=json.loads(saved['rootLandingAcceptance'])
assert accept['head']==head and accept['tree']==tree and accept['base']==BASE and accept['decision']=='accepted for exact-head landing checks'
assert accept['evidence']=={k:v for k,v in pins['evidence'].items() if k!='rootLandingAcceptance'}
pr=api(f'pulls/{PR}');run=api(f'actions/runs/{RUN}');job=api(f'actions/jobs/{JOB}')
assert pr['head']['sha']==head and pr['base']['ref']=='main'
assert run['head_sha']==head and run['event']=='pull_request' and run['status']=='completed' and run['conclusion']=='success'
assert job['run_id']==RUN and job['head_sha']==head and job['status']=='completed' and job['conclusion']=='success'
assert job['steps'] and all(s['status']=='completed' and s['conclusion']=='success' for s in job['steps'])
for name in REQUIRED_STEPS:assert len([s for s in job['steps'] if s['name']==name])==1,name
checks=api(f'commits/{head}/check-runs?per_page=100');assert checks['total_count']==len(checks['check_runs'])
assert checks['check_runs'] and all(x['status']=='completed' and x['conclusion'] in ['success','neutral','skipped'] for x in checks['check_runs'])
assert any(x['id']==JOB and x['conclusion']=='success' for x in checks['check_runs'])
log=raw(f'actions/jobs/{JOB}/logs');assert sha(log)==pins['ciLogSha256']
lines=log.decode().splitlines();actual=[]
for i,line in enumerate(lines[:-1]):
 if '[command]' in line and 'git log -1 --format=%H' in line:
  value=lines[i+1].split()[-1]
  if valid(value,40):actual.append(value)
assert actual==[S],actual
tested=api('git/commits/'+S);assert tested['tree']['sha']==tree and [x['sha'] for x in tested['parents']]==[BASE,head]
main=api('git/ref/heads/main')['object']['sha']
if mode=='premerge':
 assert not pr['merged'] and pr['state']=='open' and not pr['draft']
 assert pr['base']['sha']==BASE and main==BASE and pr['mergeable'] is True and pr['mergeable_state']=='clean' and pr['merge_commit_sha']==S
 result={'testedMerge':S,'parents':[BASE,head]}
else:
 assert pr['merged'] and pr['state']=='closed' and pr['merge_commit_sha']==main
 merged=api('git/commits/'+main);assert merged['tree']['sha']==tree and [x['sha'] for x in merged['parents']]==[BASE,head]
 result={'actualMerge':main,'parents':[BASE,head],'mergedAt':pr['merged_at'],'testedMerge':S}
result.update(at=datetime.datetime.now(datetime.timezone.utc).isoformat(),mode=mode,head=head,tree=tree,base=BASE,run=RUN,job=JOB,ciLogSha256=sha(log),pinsSha256=sha((P/'pins.json').read_bytes()),verifiedPaths=9,requiredSteps=REQUIRED_STEPS,profileQualification=False)
with (P/(mode+'-verification.json')).open('x') as f:json.dump(result,f,indent=2);f.write('\n')
print(json.dumps(result))
