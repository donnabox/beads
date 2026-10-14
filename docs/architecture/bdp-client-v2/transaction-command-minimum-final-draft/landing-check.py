"""Read-only PR123 readiness / actual-merge verification. Never marks ready or merges.
Prepared only; future caller must use the current session guard. Any different base
requires explicit reviewed integration, not a permissive fallback in this script.
"""
from pathlib import Path
import argparse,subprocess,json,hashlib,re,datetime
E=Path(__file__).resolve().parent;L=E.parent
W=Path('/Users/dbox/repos/wt/vickie/vickie-command-minimum')
H='f2ac3f980fc8d35a3b485db04b813ad1b569d94d'
B='122800511ee6306e4105e1f73bf3658f35c90a14'
T='74327f052cef6026b3ea10e186523ace9c87a128'
RUN='38645926223';JOB='116132639415';REPO='gastownhall/bdp';PR='123'
BRANCH='codex/vickie-command-minimum'
REVIEWS={'transaction-command-minimum-native-code-review.md': '2196ad874a3ad1ba0018c364f24425c3a5abedba8a5979154b716956c134164b', 'transaction-command-minimum-native-code-review-packet-closure.md': 'b937d291018e4b6080d5985f99ff83ad26b42b135ad73b8aefa999a3d82a1376', 'transaction-command-minimum-code-review/claude.txt': 'a98fc90fb2911aaafdd0363eb854e859211685cabbebd5e50bbb0d9caf2c87db', 'transaction-command-minimum-code-review/root-dispositions.md': 'd5008f9da345844efbd28ff61fbe38be8b77e02ff53b2975d4ed04195a8cf464', 'transaction-command-minimum-format-repair/native-closure.md': '106815a6aeb536510270cc8c0d021d3035154085a3c6835bd06686ec8d4e15b1', 'transaction-command-minimum-final-draft/native-closure.md': 'a26012d62ebd12311d435670bdfd6ff8ed027ae8cb10d81e14ea5d3ddb434c6d', 'transaction-command-minimum-source-native-audit.md': '940913a7e1e21b0a2a11ea4c1ea40c86d2d9ea51772c4808941bd2c7b83d6eaa', 'transaction-command-minimum-control-repair/native-closure.md': 'd746ea4d87ff7b478ceb4ccab1e46966cbfa7b9d0fe256eba3e9e3b46ecce109', 'transaction-command-minimum-final-draft/native-controls-final-source-audit.md': '3ff2f2b51318c1736d7201b80a11bebf747f0270068184324f65989a90506323', 'transaction-command-minimum-final-draft/root-source-acceptance.md': 'c94c128b22a082d6f8a35b780c918cf467dcaed9f76fbae1054a28a25d7695fc', 'transaction-command-minimum-final-draft/root-controls-acceptance.md': 'ca388b027a8c58cadd1723661fec3cee161f9935537400292523e482d67d75d0', 'transaction-command-minimum-final-draft/root-package-acceptance.md': 'b12ea351f690e74be07889bcdb8b09599b7586d0a3357db54666e2f577f15955'}
EVIDENCE={'final-source.json': '6311aeef96c7e348d364ee700ff9e390011f5a8cc293854e473fa02a77827335', 'source-binding.json': 'ff9c63417fbb348bf8ad9863483096ccc0329beb5c05b532333cf75cd0726a38', 'root-controls-summary.json': '8d762ff434d78cba18e5dc2c7f5f8a0e613cad51a03da491e235f71f64a06266', 'root-installed-summary.json': '06d91c81bf6db7a015efd069272757dad8216a1940321ad6d14bf88c4d3708ce'}
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def cmd(*argv):return subprocess.check_output(argv,cwd=W,text=True,timeout=60).strip()
def gh(*argv):return json.loads(cmd('gh',*argv))
def check(ok,reason):
 if not ok:raise RuntimeError(reason)
def main():
 args=argparse.ArgumentParser(description=__doc__);args.add_argument('--verify-merged',action='store_true');mode=args.parse_args()
 # Pending acceptance/audit is a hard stop, not a manufactured pass or default.
 accept=json.loads((E/'root-landing-acceptance.json').read_text())
 check(accept.get('accepted') is True and accept.get('head')==H and accept.get('tree')==T and accept.get('base')==B and accept.get('pr')==123,'missing exact root landing acceptance')
 check(str(accept.get('ciRun'))==RUN and str(accept.get('ciJob'))==JOB and accept.get('publicQualification') is False,'acceptance CI/qualification mismatch')
 check(accept.get('reviewHashes')==REVIEWS and accept.get('evidenceHashes')==EVIDENCE,'root must bind exact completed review/evidence chain')
 for p,d in REVIEWS.items():check(sha(L/p)==d,'review changed: '+p)
 for p,d in EVIDENCE.items():check(sha(E/p)==d,'evidence changed: '+p)
 audit=accept['finalActualEvidenceAudit'];ap=(L/audit['path']).resolve()
 check(ap.is_relative_to(L.resolve()) and audit.get('accepted') is True and re.fullmatch('[0-9a-f]{64}',audit.get('sha256','')),'invalid or unaccepted final actual audit')
 check(sha(ap)==audit['sha256'],'pending/changed final actual evidence audit')
 check(cmd('git','rev-parse','HEAD')==H and cmd('git','rev-parse','HEAD^{tree}')==T and not cmd('git','status','--porcelain'),'source not clean exact head/tree')
 check(cmd('git','rev-list','--parents','-n','1',H).split()==[H,B],'ordinary single parent changed')
 source=json.loads((E/'final-source.json').read_text())
 check(source['head']==H and source['tree']==T and source['parent']==B and source['clean'] is True,'final source mismatch')
 check(source['sourceBindingSha256']==sha(E/'source-binding.json'),'final source must bind actual draft checks')
 check(set(cmd('git','diff','--name-only',B,H).splitlines())=={f['path'] for f in source['sevenFiles']},'source footprint changed')
 for f in source['sevenFiles']:check(sha(W/f['path'])==f['sha256'],'source byte mismatch: '+f['path'])
 controls=json.loads((E/'root-controls-summary.json').read_text())
 check(controls['accepted'] is True and controls['source']==source['sourceBindingSha256'] and controls['publicQualification'] is False,'control summary mismatch')
 check(len(controls['controls'])==2 and {x['name'] for x in controls['controls']}=={'restore-command','bypass-receipt'},'two exact controls required')
 for c in controls['controls']:
  check(c['cleanupIndependentlyVerified'] is True,'control cleanup unverified')
  check(c['diagnostic']['status']==(500 if c['name']=='restore-command' else 200),'wrong intended control response')
 installed=json.loads((E/'root-installed-summary.json').read_text())
 check(installed['accepted'] is True and installed['sourceCommit']==H and installed['sourceTree']==T and installed['actualResponses']==18,'installed source/count mismatch')
 check(installed['closed'] is True and installed['childExitCode']==0 and installed['independentCurrentAbsence'] is True and installed['claimEligible'] is False,'installed cleanup/claim mismatch')
 check(re.fullmatch('[0-9a-f]{64}',installed['installedPayloadDigest']) is not None,'missing actual payload binding')
 p=gh('pr','view',PR,'--repo',REPO,'--json','state,headRefOid,headRefName,baseRefOid,baseRefName,isDraft,statusCheckRollup,mergeCommit,url')
 check(p['headRefOid']==H and p['headRefName']==BRANCH and p['baseRefName']=='main','PR identity/head/base branch mismatch')
 rollup=p['statusCheckRollup'];check(bool(rollup),'no status checks')
 check(all((x.get('status')=='COMPLETED' and x.get('conclusion')=='SUCCESS') or x.get('state')=='SUCCESS' for x in rollup),'PR check pending/non-success')
 ci=gh('run','view',RUN,'--repo',REPO,'--json','status,conclusion,headSha,jobs,url,event,workflowName')
 check(ci['headSha']==H and ci['status']=='completed' and ci['conclusion']=='success' and ci['event']=='pull_request' and ci['workflowName']=='CI','exact full CI not successful')
 check(len(ci['jobs'])==1 and str(ci['jobs'][0]['databaseId'])==JOB and ci['jobs'][0]['conclusion']=='success','exact CI job mismatch')
 required=['Run pnpm install --frozen-lockfile','Run pnpm format:check','Run pnpm lint','Run pnpm check:boundaries','Run pnpm typecheck','Run pnpm build','Run pnpm test','Run pnpm e2e:ready:preflight','Run pnpm smoke','Read-cohort evidence gate']
 steps=ci['jobs'][0]['steps']
 for name in required:check(sum(s['name']==name and s['conclusion']=='success' for s in steps)==1,'required CI step not successful: '+name)
 log=cmd('gh','api',f'repos/{REPO}/actions/jobs/{JOB}/logs')+'\n'
 check(hashlib.sha256(log.encode()).hexdigest()==accept['ciLogSha256'],'root must read/bind actual complete successful CI log')
 clean=re.sub(r'\x1b\[[0-9;]*m','',log);lines=clean.splitlines();tested=[]
 for i,line in enumerate(lines[:-1]):
  if '/usr/bin/git log -1 --format=%H' in line:
   match=re.search(r'\b[0-9a-f]{40}\b',lines[i+1]);check(match is not None,'missing checked-out CI commit');tested.append(match.group())
 check(len(set(tested))==1,'ambiguous/missing actual tested CI checkout')
 testedCommit=tested[0];commit=gh('api',f'repos/{REPO}/git/commits/{testedCommit}')
 check(commit['tree']['sha']==T and [x['sha'] for x in commit['parents']]==[B,H],'actual CI synthetic tree/parents differ')
 # H is a direct child of B. A normal no-ff merge of B and H has exactly H's tree;
 # no merge-tree --write-tree or fetch is needed (read-only local and remote).
 remoteMain=cmd('git','ls-remote','origin','refs/heads/main').split()[0]
 merged=None
 if mode.verify_merged:
  check(p['state']=='MERGED' and p.get('mergeCommit') and p['mergeCommit']['oid']==remoteMain,'actual main is not this PR merge')
  merged=gh('api',f'repos/{REPO}/git/commits/{remoteMain}')
  check(merged['tree']['sha']==T and [x['sha'] for x in merged['parents']]==[B,H],'actual merge must preserve base/head parents and tested tree')
 else:
  check(p['state']=='OPEN' and p['baseRefOid']==B and remoteMain==B,'base advanced or PR not open: explicit reviewed integration required')
  check(cmd('git','ls-remote','origin','refs/heads/'+BRANCH).split()[0]==H,'published head moved')
 # Final observational readback reduces races; it never substitutes for the merger's
 # separate immediate session/head/base recheck and head-matching merge call.
 again=gh('pr','view',PR,'--repo',REPO,'--json','state,headRefOid,baseRefName,baseRefOid,mergeCommit,statusCheckRollup')
 check(again['headRefOid']==H and again['baseRefName']=='main' and again['state']==p['state'] and again['statusCheckRollup']==rollup,'PR/checks changed during inspection')
 check(cmd('git','ls-remote','origin','refs/heads/main').split()[0]==remoteMain,'main changed during inspection')
 if not mode.verify_merged:check(again['baseRefOid']==B,'metadata base changed')
 else:check(again['mergeCommit']['oid']==remoteMain,'merge readback changed')
 result={'verifiedAt':datetime.datetime.now(datetime.timezone.utc).isoformat(),'mode':'merged-verification' if mode.verify_merged else 'readiness','head':H,'base':B,'tree':T,'pr':123,'ciRun':RUN,'ciJob':JOB,'actualTestedCommit':testedCommit,'actualTestedParents':[B,H],'actualMain':remoteMain,'actualMerge':None if merged is None else {'sha':remoteMain,'tree':merged['tree']['sha'],'parents':[x['sha'] for x in merged['parents']]},'rootAcceptanceSha256':sha(E/'root-landing-acceptance.json'),'finalAuditSha256':audit['sha256'],'releaseAuthorized':False,'publicQualification':False}
 output=E/('actual-merge-verification.json' if mode.verify_merged else 'merge-readiness.json')
 output.write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result))
if __name__=='__main__':main()
