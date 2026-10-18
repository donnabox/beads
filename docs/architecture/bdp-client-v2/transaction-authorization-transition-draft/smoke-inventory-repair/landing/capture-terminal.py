from pathlib import Path
import subprocess,json,hashlib,re,datetime,base64
P=Path(__file__).resolve().parent
W=Path('/Users/dbox/repos/wt/vickie/vickie-authorization-transition')
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def git(*args):return subprocess.check_output(['git',*args],cwd=W,timeout=15).decode().strip()
c=json.loads((P/'inputs.json').read_text())
assert sha(P/'inputs.json')=='df80c678aac26ea5bc488590453c6cab76b57c5fd041b7845682f5d8aa0dd52c'
assert git('rev-parse','HEAD')==c['head'] and git('rev-parse','HEAD^{tree}')==c['tree'] and not git('status','--porcelain')
assert git('show','-s','--format=%P','HEAD').split()==c['parents']
assert set(git('diff','--name-only',c['base'],'HEAD').splitlines())==set(c['files'])
assert all(sha(W/n)==h for n,h in c['files'].items())
R=P/'terminal';R.mkdir(exist_ok=False)
def capture(name,endpoint):
 argv=['gh','api','--method','GET','repos/gastownhall/bdp/'+endpoint]
 r=subprocess.run(argv,capture_output=True,timeout=90)
 (R/name).write_bytes(r.stdout);(R/(name+'.stderr')).write_bytes(r.stderr)
 (R/(name+'.capture.json')).write_text(json.dumps({'argv':argv,'exit':r.returncode,'at':datetime.datetime.now(datetime.timezone.utc).isoformat(),'sha256':sha(R/name),'bytes':len(r.stdout)},indent=2)+'\n')
 assert r.returncode==0,name
 return r.stdout
run=json.loads(capture('run.json',f"actions/runs/{c['run']}"));job=json.loads(capture('job.json',f"actions/jobs/{c['job']}"))
assert run['id']==c['run'] and job['id']==c['job'] and job['run_id']==c['run'] and run['head_sha']==job['head_sha']==c['head'] and run['run_attempt']==c['attempt'] and run['event']=='pull_request'
assert run['status']==job['status']=='completed','No partial logs'
log=capture('job.log',f"actions/jobs/{c['job']}/logs").decode();lines=log.splitlines();checkouts=[]
for i,line in enumerate(lines[:-1]):
 if '[command]/usr/bin/git log -1 --format=%H' in line:
  m=re.search(r'(?:^|\s)([0-9a-f]{40})\s*$',lines[i+1]);assert m;checkouts.append(m.group(1))
assert len(checkouts)==1
synthetic=json.loads(capture('synthetic.json','git/commits/'+checkouts[0]))
assert synthetic['tree']['sha']==c['tree'] and [p['sha']for p in synthetic['parents']]==[c['base'],c['head']]
wf=json.loads(capture('workflow.json','contents/.github/workflows/ci.yml?ref='+checkouts[0]))
assert hashlib.sha256(base64.b64decode(wf['content'])).hexdigest()==c['workflowSha256']
pr=json.loads(capture('pr.json',f"pulls/{c['pr']}"));main=json.loads(capture('main.json','git/ref/heads/main'))
checks=json.loads(capture('checks.json',f"commits/{c['head']}/check-runs?per_page=100"))
assert pr['head']['sha']==c['head'] and pr['base']['sha']==main['object']['sha']==c['base']
assert sha(Path(c['council']['path']))==c['council']['sha256']
out={'head':c['head'],'tree':c['tree'],'base':c['base'],'synthetic':checkouts[0],'syntheticParents':[p['sha']for p in synthetic['parents']],'runConclusion':run['conclusion'],'jobConclusion':job['conclusion'],'steps':[{k:x.get(k)for k in ['name','status','conclusion','started_at','completed_at']}for x in job['steps']],'rawLogSha256':sha(R/'job.log'),'humanFullLogReviewRequired':True,'mergeAuthorized':False}
(R/'capture-outcome.json').write_text(json.dumps(out,indent=2)+'\n')
print(json.dumps(out))
