from pathlib import Path
import json,hashlib,subprocess,sys,datetime
L=Path('/Users/dbox/repos/wt/vickie/bdp-integration-logs'); W=Path('/Users/dbox/repos/wt/vickie/vickie-sequence-depth-retained'); V=L/'sequence-minimum-combined-validation'; F=L/'transaction-fresh-depth-plan/claude-review-preparation'; R=L/'sequence-depth-retained-member-draft/review-corrections'
def git(*a):return subprocess.check_output(['git',*a],cwd=W)
def call(*a):subprocess.run(['git',*a],cwd=W,check=True)
def write(n,o):
 with (V/n).open('x') as f:json.dump(o,f,indent=2);f.write('\n')
sha=lambda b:hashlib.sha256(b).hexdigest()
m=json.loads((L/'sequence-minimum-branch-integration-current.json').read_text()); b=json.loads((R/'final-source-binding.json').read_text()); e=json.loads((F/'exit.json').read_text()); c=json.loads((F/'cleanup.json').read_text())
assert e['sourceAndPlanUnchanged'] and e['checkoutUnchanged']
assert not c['remainingGroup'] and not c['sampledPidsStillPresent']
assert json.loads((R/'root-final-code-council.json').read_text())['codeReviewAccepted']
assert git('rev-parse','HEAD').decode().strip()==m['commonBase']
assert not git('diff','--cached').strip()
assert set(git('diff','--name-only').decode().splitlines()+git('ls-files','--others','--exclude-standard').decode().splitlines())==set(b['files'])
for p,v in b['files'].items():assert sha((W/p).read_bytes())==v['sha256'],p
call('add','--',*b['files'])
call('commit','-m','fix: classify retained sequence keys before fresh depth limits')
author=git('rev-parse','HEAD').decode().strip();assert git('show','-s','--format=%P',author).decode().strip()==m['commonBase']
write('author-commit.json',{'head':author,'tree':git('rev-parse','HEAD^{tree}').decode().strip(),'parent':m['commonBase'],'sourceBindingSha256':sha((R/'final-source-binding.json').read_bytes())})
r=subprocess.run(['git','merge','--no-ff','--no-commit',m['minimum']['head']],cwd=W,capture_output=True,text=True)
write('merge-attempt.json',{'exit':r.returncode,'stdout':r.stdout,'stderr':r.stderr,'author':author,'incoming':m['minimum']['head']})
assert r.returncode in (0,1)
print(r.stdout);print(r.stderr);print('Author:',author)
