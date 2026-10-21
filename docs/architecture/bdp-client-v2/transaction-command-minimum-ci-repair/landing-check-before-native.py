"""Read-only successor for repaired PR123. Preserves historical runtime attribution."""
from pathlib import Path
import subprocess,json,hashlib,runpy,sys,datetime
R=Path(__file__).resolve().parent;L=R.parent;F=L/'transaction-command-minimum-final-draft'
W=Path('/Users/dbox/repos/wt/vickie/vickie-command-minimum')
B='122800511ee6306e4105e1f73bf3658f35c90a14';O='f2ac3f980fc8d35a3b485db04b813ad1b569d94d';H='727e919c2893bd34d5e72ea4659302a999bc4a6f'
T='734bf9daf40fd17d0fe2bf86b085598589122f2f';OT='74327f052cef6026b3ea10e186523ace9c87a128';S='ac6652f2d235915d87db04b710f4afcff45a6080'
def cmd(*a):return subprocess.check_output(a,cwd=W,text=True,timeout=60).strip()
def gh(*a):return json.loads(cmd('gh',*a))
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def check(v,m):
 if not v:raise RuntimeError(m)
old=runpy.run_path(str(F/'landing-check.py'),run_name='historical_constants')
for p,h in old['REVIEWS'].items():check(sha(L/p)==h,'historical review hash: '+p)
for p,h in old['EVIDENCE'].items():check(sha(F/p)==h,'historical evidence hash: '+p)
check(sha(L/'transaction-command-minimum-later-preparation-v4/native-actual-evidence-audit.md')=='be01d032e47e7acd4efa8ee12a15a79f0e11f763143f5fe8c544ea2f4e790cc6','installed audit changed')
a=json.loads((R/'root-landing-acceptance.json').read_text())
check(a['accepted'] is True and a['head']==H and a['base']==B and a['tree']==T and a['publicQualification'] is False,'missing exact root acceptance')
for p,h in a['repairEvidenceHashes'].items():check(sha(R/p)==h,'repair binding changed: '+p)
c=json.loads((R/'claude-account-switched/exit.json').read_text())
check(c['exitCode']==0 and c['sourceUnchanged'] is True and c['outputSHA256']==sha(R/'claude-account-switched/review.txt'),'actual successful Claude review required')
check(cmd('git','rev-parse','HEAD')==H and cmd('git','rev-parse','HEAD^{tree}')==T and cmd('git','status','--porcelain')=='','clean exact source required')
check(cmd('git','rev-list','--parents','-n','1',H).split()==[H,O],'repair ordinary parent')
check(cmd('git','rev-list','--parents','-n','1',O).split()==[O,B],'original ordinary parent')
check(cmd('git','diff','--name-only',O,H)=='docs/specs/bdp.md','repair scope')
check(cmd('git','diff','--binary',O,H)==(R/'repair.patch').read_text().strip(),'repair patch differs')
f=json.loads((R/'final-source.json').read_text());o=json.loads((F/'final-source.json').read_text())
check((f['head'],f['tree'],f['parent'],f['mainBase'],f['clean'])==(H,T,O,B,True),'repair source manifest')
check((o['head'],o['tree'],o['parent'],o['clean'])==(O,OT,B,True),'historical source manifest')
check(set(cmd('git','diff','--name-only',B,H).splitlines())=={x['path'] for x in f['sevenFiles']}=={x['path'] for x in o['sevenFiles']},'seven-file scope')
prior={x['path']:x['sha256'] for x in o['sevenFiles']}
for x in f['sevenFiles']:
 check(sha(W/x['path'])==x['sha256'],'source file changed: '+x['path'])
 if x['path']!='docs/specs/bdp.md':check(x['sha256']==prior[x['path']],'non-spec delta')
controls=json.loads((F/'root-controls-summary.json').read_text())
check(controls['accepted'] and controls['source']==o['sourceBindingSha256'] and controls['publicQualification'] is False,'historical controls')
check(len(controls['controls'])==2 and {x['name'] for x in controls['controls']}=={'restore-command','bypass-receipt'},'control identity')
for x in controls['controls']:check(x['cleanupIndependentlyVerified'] and x['diagnostic']['status']==(500 if x['name']=='restore-command' else 200),'control result')
i=json.loads((F/'root-installed-summary.json').read_text())
check(i['accepted'] and (i['sourceCommit'],i['sourceTree'],i['actualResponses'])==(O,OT,18) and i['closed'] and i['childExitCode']==0 and i['independentCurrentAbsence'] and i['claimEligible'] is False,'historical installed attribution')
ci=gh('run','view','38648263438','--repo','gastownhall/bdp','--json','status,conclusion,headSha,jobs,event,workflowName')
check((ci['headSha'],ci['status'],ci['conclusion'],ci['event'],ci['workflowName'])==(H,'completed','success','pull_request','CI'),'exact CI')
check(len(ci['jobs'])==1 and ci['jobs'][0]['databaseId']==116140423947 and ci['jobs'][0]['conclusion']=='success','CI job')
for name in ['Run pnpm install --frozen-lockfile','Run pnpm format:check','Run pnpm lint','Run pnpm check:boundaries','Run pnpm typecheck','Run pnpm build','Run pnpm test','Run pnpm e2e:ready:preflight','Run pnpm smoke','Read-cohort evidence gate']:
 check(sum(s['name']==name and s['conclusion']=='success' for s in ci['jobs'][0]['steps'])==1,'CI step: '+name)
check(sha(R/'ci-success.log')=='b823305ca58f91dff21fbe0e508902a8ac24845bca04b0c381cd0a09544213ad','retained CI log')
synthetic=gh('api','repos/gastownhall/bdp/git/commits/'+S)
check(synthetic['tree']['sha']==T and [x['sha'] for x in synthetic['parents']]==[B,H],'tested tree/parents')
fields='state,headRefOid,headRefName,baseRefOid,baseRefName,isDraft,statusCheckRollup,mergeCommit'
p=gh('pr','view','123','--repo','gastownhall/bdp','--json',fields)
check(p['headRefOid']==H and p['headRefName']=='codex/vickie-command-minimum' and p['baseRefName']=='main','PR identity')
check(p['statusCheckRollup'] and all((x.get('status')=='COMPLETED' and x.get('conclusion')=='SUCCESS') or x.get('state')=='SUCCESS' for x in p['statusCheckRollup']),'PR checks')
main=cmd('git','ls-remote','origin','refs/heads/main').split()[0];merged='--verify-merged' in sys.argv
actual=None
if merged:
 check(p['state']=='MERGED' and p['mergeCommit']['oid']==main,'merged main identity')
 actual=gh('api','repos/gastownhall/bdp/git/commits/'+main)
 check(actual['tree']['sha']==T and [x['sha'] for x in actual['parents']]==[B,H],'actual merged tree/parents')
else:
 check(p['state']=='OPEN' and p['baseRefOid']==B and main==B,'base moved or closed')
 check(cmd('git','ls-remote','origin','refs/heads/codex/vickie-command-minimum').split()[0]==H,'published head')
check(gh('pr','view','123','--repo','gastownhall/bdp','--json',fields)==p,'PR state changed during verification')
check(cmd('git','ls-remote','origin','refs/heads/main').split()[0]==main,'main changed')
result={'verifiedAt':datetime.datetime.now(datetime.timezone.utc).isoformat(),'head':H,'base':B,'tree':T,'originalRuntimeEvidenceHead':O,'ciRun':38648263438,'ciJob':116140423947,'testedCommit':S,'actualMain':main,'actualMerge':None if actual is None else {'sha':main,'tree':actual['tree']['sha'],'parents':[x['sha'] for x in actual['parents']]},'acceptanceSHA256':sha(R/'root-landing-acceptance.json'),'publicQualification':False,'releaseAuthorized':False}
(R/('merge-verification.json' if merged else 'merge-readiness.json')).write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result))
