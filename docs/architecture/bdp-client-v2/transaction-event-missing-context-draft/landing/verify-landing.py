"""Read-only exact-head and actual-merge identity checks; never merges."""
from pathlib import Path
import subprocess,json,sys,datetime,hashlib
P=Path(__file__).resolve().parent
W=Path('/Users/dbox/repos/wt/vickie/vickie-sequence-member-evidence')
HEAD='3df82ac480f50c8f486b4e456336366e7acbdf87';BASE='88c9408e908590ab1a4c64c9b9dee1a8b705999c';TREE='d5f02c936cf080666a8e73bb5f253e52086705f4';RUN=38728209051;JOB=116399188531
assert __debug__ and sys.argv[1:] in [['premerge'],['postmerge']]
mode=sys.argv[1]
def api(path):return json.loads(subprocess.check_output(['gh','api','repos/gastownhall/bdp/'+path]))
def git(*args):return subprocess.check_output(['git',*args],cwd=W,text=True).strip()
assert git('rev-parse','HEAD')==HEAD and git('rev-parse','HEAD^{tree}')==TREE and git('show','-s','--format=%P','HEAD')==BASE
assert not git('status','--porcelain')
binding=json.loads((P.parent/'redirect-repair/source-binding.json').read_text())
assert set(git('diff','--name-only',BASE,HEAD).splitlines())==set(binding['files'])
for name,expected in binding['files'].items():assert hashlib.sha256((W/name).read_bytes()).hexdigest()==expected['sha256'],name
pr=api('pulls/127');run=api(f'actions/runs/{RUN}');job=api(f'actions/jobs/{JOB}');main=api('git/ref/heads/main')['object']['sha']
assert pr['head']['sha']==HEAD and pr['base']['ref']=='main'
assert run['head_sha']==HEAD and run['event']=='pull_request' and run['status']=='completed' and run['conclusion']=='success'
assert job['run_id']==RUN and job['status']=='completed' and job['conclusion']=='success'
checks=api(f'commits/{HEAD}/check-runs')['check_runs'];assert checks and all(x['status']=='completed' and x['conclusion'] in ['success','neutral','skipped'] for x in checks)
assert any(x['id']==JOB and x['conclusion']=='success' for x in checks)
if mode=='premerge':
 assert not pr['merged'] and pr['state']=='open' and pr['base']['sha']==BASE and main==BASE and not pr['draft']
 assert pr['mergeable'] is True and pr['mergeable_state']=='clean'
 tested=api('git/commits/'+pr['merge_commit_sha']);assert tested['tree']['sha']==TREE and [x['sha'] for x in tested['parents']]==[BASE,HEAD]
 result={'testedMerge':tested['sha'],'tree':TREE,'parents':[BASE,HEAD]}
else:
 assert pr['merged'] and pr['state']=='closed' and pr['merge_commit_sha']==main
 actual=api('git/commits/'+main);assert actual['tree']['sha']==TREE and [x['sha'] for x in actual['parents']]==[BASE,HEAD]
 result={'actualMerge':main,'tree':TREE,'parents':[BASE,HEAD],'mergedAt':pr['merged_at']}
result.update(at=datetime.datetime.now(datetime.timezone.utc).isoformat(),head=HEAD,run=RUN,job=JOB,checks=[{'id':x['id'],'name':x['name'],'conclusion':x['conclusion']} for x in checks],mode=mode)
with (P/(mode+'-verification.json')).open('x') as f:json.dump(result,f,indent=2);f.write('\n')
print(json.dumps(result))
