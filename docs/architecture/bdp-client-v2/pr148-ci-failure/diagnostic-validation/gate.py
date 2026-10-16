from pathlib import Path
import hashlib,json,subprocess
P=Path(__file__).resolve().parent;D=P;L=Path('/Users/dbox/repos/wt/vickie/bdp-integration-logs')
W=Path('/Users/dbox/repos/wt/vickie/vickie-transaction-conditional-reads');E=P/'run';GUARD=L/'guarded-action.py';COMMANDS=P/'commands.json'
COMMANDS_SHA='b330a646399c99d6627a0c39322cd8e3074fa5bb0091677bfeb138bdef01c1af'
sha=lambda b:hashlib.sha256(b).hexdigest()
def require(ok,message):
 if not ok:raise RuntimeError(message)
def git(*args):return subprocess.check_output(['git',*args],cwd=W,timeout=10)
def input_files():return {n:sha((W/n).read_bytes()) for n in git('ls-files').decode().splitlines()}
def source_patch_hash():return sha(git('diff','--binary','HEAD'))
def verify_build():
 b=json.loads((P/'build-binding.json').read_text());f=Path(b['inventory']);require(sha(f.read_bytes())==b['sha256'],'Historical inventory changed');i=json.loads(f.read_text());s=json.loads((P/'source-binding.json').read_text())
 require(i['head']==s['head'] and i['tree']==s['tree'],'Historical build head/tree differs')
 actual=input_files();require(set(actual)==set(i['inputFiles']),'Build input paths changed')
 require([n for n,h in i['inputFiles'].items() if actual[n]!=h]==[b['buildInputsChangedOnly']],'Production build inputs changed')
 require(i['inputFiles'][b['buildInputsChangedOnly']]==b['historicalTestSha256'] and actual[b['buildInputsChangedOnly']]==b['currentTestSha256'],'Test-only exception changed')
 emitted={f.relative_to(W).as_posix():sha(f.read_bytes()) for root in (W/'packages').iterdir() if root.is_dir() for f in (root/'dist').rglob('*') if f.is_file()}
 require(emitted==i['emittedFiles'],'Emitted inventory changed')
 for n,h in i['packageFiles'].items():require(sha((W/n).read_bytes())==h,'Package member changed')
 for k in ['buildStart','buildExit','buildLog']:
  v=i[k];require(sha(Path(v['path']).read_bytes())==v['sha256'],'Historical build evidence changed')
def source_hashes():
 s=json.loads((P/'source-binding.json').read_text());require(git('rev-parse','HEAD').decode().strip()==s['head'],'HEAD changed');require(git('status','--porcelain').decode()==s['status'],'Scope/status changed');require(source_patch_hash()==s['diffSha256'],'Patch changed');require(input_files()==json.loads((P/'input-binding.json').read_text()),'Tracked input changed');verify_build();return {n:sha((W/n).read_bytes()) for n in s['files']}
def load_gate(name):
 require(__debug__,'Assertions required');g=json.loads((P/'execution-gate.json').read_text());require(g['released'] is True,'Execution closed');require(g['session']=='2026-10-08T20:26:35Z-81e5','Wrong session')
 for n,h in g['preparationSha256'].items():require(sha((P/n).read_bytes())==h,'Preparation changed '+n)
 require(sha(COMMANDS.read_bytes())==COMMANDS_SHA,'Commands changed');t=json.loads((P/'tool-binding.json').read_text());require(sha(GUARD.read_bytes())==t['guard']['sha256'],'Guard changed')
 for n,h in t['toolHashes'].items():require(sha(Path(n).read_bytes())==h,'Tool changed '+n)
 require(sha((W/'pnpm-lock.yaml').read_bytes())==t['sourceLockSha256'],'Lock changed')
 a=g['rootAcceptance'];f=Path(a['path']);require(f.is_relative_to(P) and sha(f.read_bytes())==a['sha256'],'Root release missing/changed');a=json.loads(f.read_text());require(a['accepted'] is True and a['nativeCodeReviewed'] is True and a['externalCodeReviewed'] is True and a['diagnosticOnly'] is True,'Code review disposition/release required');require(a['head']==json.loads((P/'source-binding.json').read_text())['head'],'Acceptance head differs')
 order=json.loads(COMMANDS.read_text())['executionOrder'];require(g['approvedCommandIds']==order[:len(g['approvedCommandIds'])] and name in g['approvedCommandIds'],'Stage closed')
 for prior in order[:order.index(name)]:
  x=json.loads((E/(prior+'.exit.json')).read_text());require(x['supervisorExit']==0 and x['exit']==0 and x['sourceUnchanged'],'Prior failure');require(not any(x.get(k) for k in ['timeout','forcedCleanup','errors','remainingAfterCleanup','sampledPidsStillPresent','runtimeDirectoriesStillPresent']),'Prior cleanup failure')
 spec=json.loads(COMMANDS.read_text())['commands'][name];require(spec['supervisorSeconds']==120 and spec['automaticRetries']==0,'Bound changed');g['sourceSha256']=source_hashes();g['sourceDiffSha256']=source_patch_hash();return g,spec
