"""Read-only successor landing verifier; never makes a PR ready or merges.
Candidate and root acceptance are frozen; root must still freeze the CI constants.
Run through guarded-action.py only after green. Root acceptance is the existing patch-bound diagnosis decision; its exact bytes,
native report and actual Claude result are checked without inventing a new schema.
No record is overwritten; premerge/postmerge outputs are separate.
"""
from pathlib import Path
import subprocess,json,sys,datetime,hashlib

P=Path(__file__).resolve().parent
D=P.parent
W=Path('/Users/dbox/repos/wt/vickie/vickie-sequence-member-evidence')
BASE='88c9408e908590ab1a4c64c9b9dee1a8b705999c'
PARENT='3df82ac480f50c8f486b4e456336366e7acbdf87'
HEAD='db5a4d6a41814e59704ebef60fda02a2d57e9ad2';TREE='2e16329e195eb8214836de20340a5887236d78f0';RUN=38731731790;JOB=116410895014;SYNTHETIC='52a71dbbfb6a4d9cbf231bc057189dee369c1cf0'
ROOT_ACCEPTANCE=D/'ci-timeout-diagnosis/root-acceptance.json'
ROOT_ACCEPTANCE_SHA256='da560e3a320f54ba54a87d9b8c76366361a2bf71e97136003255b2a224b18af6'
CLAUDE_EXIT_SHA256='2ad32a5e478c10f14c3c3c45f4d39d9bb9825c411e5d925ad808fb4246424262'
TIMER='apps/bdptest/src/development-contracts.test.ts'
TIMER_SHA='c27a4ff9012c85af17bd38e4f06a3b65d26a75a949e57a572a4382fd5bf6e522'
PATCH_SHA='4c355708c6f300d75824e6ac57eaa0e1acb2aa37691c1c59956993b9225d59ac'
NATIVE_SHA='0ad948473f736dcf863adbf257c827d4659e9452ec4bb319b15a4b4acd98b0ca'
BINDING_SHA='acf190e742dfb669fe5607085a3ac60c5ecb66f4f86b1f05025424a7873615b0'
BUILD_SHA='a8da14e9944a1bf9e8e78f12230bd5f60aec1e02e265010dfae15fb0bc4d05e5'
assert __debug__ and sys.argv[1:] in [['premerge'],['postmerge']]
assert all(isinstance(x,str) and len(x)==40 for x in [HEAD,TREE,SYNTHETIC]),'Candidate/CI binding not frozen'
assert type(RUN) is int and type(JOB) is int
assert isinstance(ROOT_ACCEPTANCE_SHA256,str) and len(ROOT_ACCEPTANCE_SHA256)==64,'Root gate not frozen'
mode=sys.argv[1]
def sha(data):return hashlib.sha256(data).hexdigest()
def git(*args):return subprocess.check_output(['git',*args],cwd=W)
def textgit(*args):return git(*args).decode().strip()
def api(path):return json.loads(subprocess.check_output(['gh','api','repos/gastownhall/bdp/'+path]))
def document(path,digest):
 data=Path(path).read_bytes();assert sha(data)==digest,str(path)
 return json.loads(data)
def pinned_text(record):
 path=Path(record['path']);data=path.read_bytes()
 assert sha(data)==record['sha256'] and data.strip(),str(path)
 return data

# Accept only the reviewed ordinary successor and its exact patch, not a rebased equivalent.
assert textgit('rev-parse','HEAD')==HEAD and textgit('rev-parse','HEAD^{tree}')==TREE
assert textgit('show','-s','--format=%P',HEAD)==PARENT
assert textgit('show','-s','--format=%P',PARENT)==BASE
assert not textgit('status','--porcelain')
binding=document(D/'redirect-repair/source-binding.json',BINDING_SHA)
assert len(binding['files'])==6
assert set(textgit('diff','--name-only',BASE,HEAD).splitlines())==set(binding['files'])|{TIMER}
assert textgit('diff','--name-only',PARENT,HEAD)==TIMER
assert sha(git('diff','--binary',PARENT,HEAD,'--',TIMER))==PATCH_SHA
assert sha((W/TIMER).read_bytes())==TIMER_SHA
for name,expected in binding['files'].items():
 assert sha((W/name).read_bytes())==expected['sha256'],name
build=document(D/'installed-preparation/build-output-inventory.json',BUILD_SHA)
for name,digest in build['inputFiles'].items():
 assert sha((W/name).read_bytes())==(TIMER_SHA if name==TIMER else digest),name
for name,digest in build['files'].items():
 assert sha((W/name).read_bytes())==digest,name

# Root acceptance must bind the actual native report and completed actual Claude supplement.
accept=document(ROOT_ACCEPTANCE,ROOT_ACCEPTANCE_SHA256)
assert accept['base']==PARENT and accept['patchSHA256']==PATCH_SHA and accept['nativeSHA256']==NATIVE_SHA
assert accept['decision']=='accept bounded successor for ordinary commit and fresh CI; merge not yet authorized by gates'
assert accept['validation']['selectedPassed']==1 and accept['validation']['notSelected']==14
pinned_text({'path':D/'ci-timeout-diagnosis/native-exact-patch-review.md','sha256':NATIVE_SHA})
claude=pinned_text({'path':D/'claude-ci-timeout-supplement/stdout.txt','sha256':accept['actualClaudeSHA256']})
exit_record=document(D/'claude-ci-timeout-supplement/exit.json',CLAUDE_EXIT_SHA256)
assert exit_record['exitCode']==0 and not exit_record['timeout'] and not exit_record['forcedTermination']
assert exit_record['sourceAndPlanUnchanged'] is True and exit_record['stdoutSha256']==sha(claude)

pr=api('pulls/127');run=api(f'actions/runs/{RUN}');job=api(f'actions/jobs/{JOB}')
main=api('git/ref/heads/main')['object']['sha']
assert pr['head']['sha']==HEAD and pr['base']['ref']=='main'
assert run['head_sha']==HEAD and run['event']=='pull_request' and run['status']=='completed' and run['conclusion']=='success'
assert job['run_id']==RUN and job['head_sha']==HEAD and job['status']=='completed' and job['conclusion']=='success'
required=['Run pnpm format:check','Run pnpm lint','Run pnpm check:boundaries','Run pnpm typecheck','Run pnpm build','Run pnpm test','Run pnpm e2e:ready:preflight','Run pnpm smoke','Read-cohort evidence gate']
for name in required:
 steps=[x for x in job['steps'] if x['name']==name]
 assert len(steps)==1 and steps[0]['status']=='completed' and steps[0]['conclusion']=='success',name
checks=api(f'commits/{HEAD}/check-runs?per_page=100')['check_runs']
assert checks and all(x['status']=='completed' and x['conclusion'] in ['success','neutral','skipped'] for x in checks)
assert any(x['id']==JOB and x['conclusion']=='success' for x in checks)
tested=api('git/commits/'+SYNTHETIC)
assert tested['tree']['sha']==TREE and [x['sha'] for x in tested['parents']]==[BASE,HEAD]
if mode=='premerge':
 assert not pr['merged'] and pr['state']=='open' and pr['base']['sha']==BASE and main==BASE and not pr['draft']
 assert pr['mergeable'] is True and pr['mergeable_state']=='clean'
 assert pr['merge_commit_sha']==SYNTHETIC
 result={'testedMerge':SYNTHETIC,'tree':TREE,'parents':[BASE,HEAD]}
else:
 assert pr['merged'] and pr['state']=='closed' and pr['merge_commit_sha']==main
 actual=api('git/commits/'+main)
 assert actual['tree']['sha']==TREE and [x['sha'] for x in actual['parents']]==[BASE,HEAD]
 result={'actualMerge':main,'tree':TREE,'parents':[BASE,HEAD],'mergedAt':pr['merged_at'],'testedMerge':SYNTHETIC}
result.update(at=datetime.datetime.now(datetime.timezone.utc).isoformat(),head=HEAD,parent=PARENT,run=RUN,job=JOB,mode=mode,rootAcceptanceSha256=ROOT_ACCEPTANCE_SHA256,verifiedEventFiles=6,verifiedCumulativePaths=7,productOutputsPreserved=len(build['files']),requiredSteps=required,checks=[{'id':x['id'],'name':x['name'],'conclusion':x['conclusion']} for x in checks],installedAttribution=PARENT)
with (P/(mode+'-verification.json')).open('x') as f:json.dump(result,f,indent=2);f.write('\n')
print(json.dumps(result))
