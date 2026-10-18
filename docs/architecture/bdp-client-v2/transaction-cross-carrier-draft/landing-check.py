"""Read-only fail-closed landing preflight; does not mark ready or merge."""
from pathlib import Path
import subprocess,json,hashlib,re,datetime
E=Path(__file__).resolve().parent;W=Path('/Users/dbox/repos/wt/vickie/vickie-cross-carrier-draft')
H='c02ccc3706f2e258d380805162a8c47dbb73f4e1'; B='eff171839fe4f37638284a4c3346428979d26b5d';T='088290751951c353797b984f690b7d9fd96deb80';RUN='38635297329'
def cmd(*a):return subprocess.check_output(a,cwd=W,text=True).strip()
def gh(*a):return json.loads(cmd('gh',*a))
assert cmd('git','rev-parse','HEAD')==H and cmd('git','rev-parse','HEAD^{tree}')==T and not cmd('git','status','--porcelain')
source=json.loads((E/'final-source.json').read_text());assert source['head']==H and source['tree']==T and source['parent']==B
assert set(cmd('git','diff','--name-only',B,H).splitlines())==set(source['sourceSha256'])
for p,d in source['sourceSha256'].items():assert hashlib.sha256((W/p).read_bytes()).hexdigest()==d
assert cmd('git','rev-list','--parents','-n','1',H).split()==[H,B]
assert cmd('git','ls-remote','origin','refs/heads/main').split()[0]==B
assert cmd('git','ls-remote','origin','refs/heads/codex/vickie-cross-carrier-draft').split()[0]==H
p=gh('pr','view','122','--repo','gastownhall/bdp','--json','state,headRefOid,baseRefOid,baseRefName,isDraft,statusCheckRollup');assert p['state']=='OPEN' and p['headRefOid']==H and p['baseRefOid']==B and p['baseRefName']=='main'
assert p['statusCheckRollup'] and all(x.get('conclusion')=='SUCCESS' or x.get('state')=='SUCCESS' for x in p['statusCheckRollup'])
ci=gh('run','view',RUN,'--repo','gastownhall/bdp','--json','status,conclusion,headSha,jobs,url');assert ci['headSha']==H and ci['status']=='completed' and ci['conclusion']=='success'
assert len(ci['jobs'])==1 and ci['jobs'][0]['conclusion']=='success'
for n in ['Run pnpm test','Run pnpm e2e:ready:preflight','Run pnpm smoke','Read-cohort evidence gate']:assert any(s['name']==n and s['conclusion']=='success' for s in ci['jobs'][0]['steps']),n
log=cmd('gh','api','repos/gastownhall/bdp/actions/jobs/'+str(ci['jobs'][0]['databaseId'])+'/logs');(E/'ci-success.log').write_text(log+'\n'); clean=re.sub(r'\x1b\[[0-9;]*m','',log);lines=clean.splitlines();c=None
for i,line in enumerate(lines):
 if '/usr/bin/git log -1 --format=%H' in line:
  m=re.search(r'\b[0-9a-f]{40}\b',lines[i+1]);assert m;c=m.group()
assert c; commit=gh('api','repos/gastownhall/bdp/git/commits/'+c);assert commit['tree']['sha']==T and [v['sha'] for v in commit['parents']]==[B,H]
assert cmd('git','merge-tree','--write-tree',B,H).splitlines()[0]==T
reviews={
'claude-review.txt':'a85ec7d1d9691293bbbd823f10c4d030e9c10809fd6e482f74c0018eb813c2ab',
'claude-root-dispositions.md':'055bdc8cce2ecddd9be95669931efd4222db2e93d98f35110ae44f76c1f88fa9',
'review-repair/native-closure.md':'801b55c181795d29db0d9b22d080538fac7d3cea04a4ca65bfa97f6caf7b3add',
'native-execution-preparation-review.md':'72cf024b83d680e58a1df28b077a522ae76b896d029509366bf0c06a4baef50d'}
for n,d in reviews.items():assert hashlib.sha256((E/n).read_bytes()).hexdigest()==d
# Root must read the actual evidence review and CI log, bind them here, and explicitly accept.
accept=json.loads((E/'root-landing-acceptance.json').read_text());assert accept['accepted'] is True and accept['head']==H and accept['tree']==T
assert accept['ciLogSha256']==hashlib.sha256((E/'ci-success.log').read_bytes()).hexdigest()
assert accept['nativeActualEvidenceSha256']==hashlib.sha256((E/'native-actual-evidence-review.md').read_bytes()).hexdigest()
installed=json.loads((E/'installed-summary.json').read_text());assert installed['sourceCommit']==H and installed['tree']==T and len(installed['cohorts'])==9 and installed['attemptedHttp']==1641 and installed['closedChildren']==installed['childLaunches']==189 and installed['claimEligible'] is False
assert accept['installedSummarySha256']==hashlib.sha256((E/'installed-summary.json').read_bytes()).hexdigest()
expected={'positive','cross-batch','cross-sequence','noop-bead-property','noop-link-metadata','noop-group-inventory-and-head','deletion-incoming-order','deletion-outgoing-order','deletion-committed-prefix-state'}
assert {c['name'] for c in installed['cohorts']}==expected
for cohort in installed['cohorts']:
 assert cohort['accepted'] is True and cohort['sourceCommit']==H and cohort['closedChildren']==cohort['childLaunches']==21 and cohort['claimEligible'] is False
 original=json.loads((E/'execution-preparation/run'/(cohort['name']+'.adjudication.json')).read_text())
 assert {k:v for k,v in cohort.items() if k!='name'}==original
r={'verifiedAt':datetime.datetime.now(datetime.timezone.utc).isoformat(),'head':H,'base':B,'tree':T,'pr':p,'ci':ci,'testedCommit':c,'testedParents':[v['sha'] for v in commit['parents']],'reviews':reviews,'rootAcceptance':accept,'releaseAuthorized':False}
(E/'merge-readiness.json').write_text(json.dumps(r,indent=2)+'\n'); print(json.dumps(r))
