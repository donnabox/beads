"""Delegated bounded fresh-package gate; source and build identities remain separately pinned."""
from pathlib import Path
import hashlib,json,re,subprocess
P=Path(__file__).resolve().parent;D=P.parent;L=D.parent.parent
W=Path('/Users/dbox/repos/wt/vickie/vickie-minimum-current-resource')
E=P/'run';GUARD=L/'guarded-action.py';COMMANDS=P/'commands.json'
COMMANDS_SHA='baf0af780504fde6b94854dc2f2e3d5129b04c0797228373a9c49ae24c344157'
sha=lambda b:hashlib.sha256(b).hexdigest()
def require(ok,message):
 if not ok:raise RuntimeError(message)
def git(*args):return subprocess.check_output(['git',*args],cwd=W,timeout=10)
def document(g,key):
 d=g[key];p=Path(d['path']);require(p.is_relative_to(L) and not p.is_symlink() and sha(p.read_bytes())==d['sha256'],'Changed document '+key)
 return json.loads(p.read_text())
def source_hashes():
 b=json.loads((P/'prepared-source.json').read_text());g=json.loads((P/'execution-gate.json').read_text())
 require(sha((P/'prepared-source.json').read_bytes())==g['sourceBinding']['sha256'],'Source binding changed')
 actual={p:sha((W/p).read_bytes()) for p in b['files']}
 require(actual=={p:v['sha256'] for p,v in b['files'].items()} and len(actual)==10,'Ten source hashes differ')
 require(bool(git('status','--porcelain').strip())==b['workspaceDirty'],'Source dirty state differs')
 require(git('rev-parse','HEAD').decode().strip()==b['head'] and git('rev-parse','HEAD^{tree}').decode().strip()==b['tree'],'Candidate identity differs')
 require(set(git('diff','--name-only',b['diffBase']).decode().splitlines())==set(actual),'Ten-path scope differs')
 require(sha(git('diff','--binary',b['diffBase']))==b['diffSha256'],'Candidate patch differs')
 require(git('show','-s','--format=%P',b['head']).decode().strip()==b['integrationCommit'],'Final doc parent')
 require(git('show','-s','--format=%P',b['integrationCommit']).decode().split()==b['integrationParents'],'Ordered merge parents')
 require(git('show','-s','--format=%P',b['authorCommit']).decode().strip()==b['authorParent'],'Author parent')
 git('merge-base','--is-ancestor',b['diffBase'],b['head'])
 return actual
def source_patch_hash():return sha(git('diff','--binary','HEAD'))
def input_files():return {p:sha((W/p).read_bytes()) for p in git('ls-files').decode().splitlines()}
def build_outputs():
 files=[W/'LICENSE']
 for name in ['protocol','server']:
  root=W/'packages'/name;files.append(root/'package.json')
  for sub in (['dist','schemas'] if name=='protocol' else ['dist']):files.extend(p for p in (root/sub).rglob('*') if p.is_file())
 require(all(not p.is_symlink() for p in files),'Package symlink')
 return {p.relative_to(W).as_posix():sha(p.read_bytes()) for p in sorted(files)}
def emitted_files():return {p.relative_to(W).as_posix():sha(p.read_bytes()) for root in (W/'packages').iterdir() if root.is_dir() for p in (root/'dist').rglob('*') if p.is_file()}
def verify_build(g):
 i=document(g,'buildOutputInventory');a=document(g,'rootAcceptance');require(a['accepted'] and a['head']==g['finalHead'],'Delegated acceptance mismatch')
 for k in ['buildStart','buildExit','buildLog','originalInventory']:
  v=i[k];require(sha(Path(v['path']).read_bytes())==v['sha256'],'Build evidence changed')
 s=json.loads(Path(i['buildStart']['path']).read_text());x=json.loads(Path(i['buildExit']['path']).read_text())
 require(s['sourceSha256']==i['sourceSha256']==x['sourceSha256After']=={p:v['sha256'] for p,v in document(g,'buildSourceBinding')['files'].items()},'Original build source differs')
 changes={p for p in i['sourceSha256'] if i['sourceSha256'][p]!=source_hashes()[p]};require(changes==set(g['changedInputPaths']),'Build/test-only bridge differs')
 require(x['exit']==x['supervisorExit']==0 and x['sourceUnchanged'],'Build failed')
 for k in ['timeout','forcedCleanup','errors','remainingAfterCleanup','sampledPidsStillPresent','runtimeDirectoriesStillPresent']:require(not x[k],'Unclean build')
 require(s['argv']==[g['node'],g['pnpm'],'run','build'] and s['cwd']==str(W),'Wrong build')
 current=input_files()
 require(document(g,'currentInputBinding')==current and {p for p in i['inputFiles'] if i['inputFiles'][p]!=current[p]}==set(g['changedInputPaths']) and i['packageFiles']==build_outputs() and i['emittedFiles']==emitted_files(),'Actual input/output bytes differ')
 require(i['buildHead']==s['head']==g['finalHead'] and i['buildWorkspaceDirty']==s['workspaceDirty']==False,'Build chronology differs')
def completed(name):
 cfg=json.loads(COMMANDS.read_text());x=json.loads((E/(name+'.exit.json')).read_text())
 require(x['exit']==cfg['commands'][name].get('rawExpectedExit',0) and x['supervisorExit']==0 and x['sourceUnchanged'],'Prior stage failed '+name)
 for k in ['timeout','forcedCleanup','errors','remainingAfterCleanup','sampledPidsStillPresent','runtimeDirectoriesStillPresent']:require(not x[k],'Unclean prior stage')
 require(x['sourceSha256After']==source_hashes(),'Prior source differs')
 return x
def proof_for(mode):return E/('installed-provenance.json' if mode=='positive' else mode+'.provenance.json')
def load_gate(name):
 require(__debug__,'Assertions required');g=json.loads((P/'execution-gate.json').read_text())
 require(g['released'] and g['session']=='2026-10-08T20:26:35Z-81e5','Gate closed')
 require(document(g,'rootAcceptance')['accepted'],'Delegated authorization absent')
 for n,h in g['preparationSha256'].items():require(sha((P/n).read_bytes())==h,'Preparation changed '+n)
 require(sha(COMMANDS.read_bytes())==COMMANDS_SHA and sha(GUARD.read_bytes())==g['guardSha256'],'Commands/guard changed')
 cfg=json.loads(COMMANDS.read_text());order=cfg['executionOrder'];require(name in g['approvedCommandIds'] and g['approvedCommandIds']==order,'Stage not released')
 source=source_hashes();verify_build(g)
 require(sha((W/'pnpm-lock.yaml').read_bytes())==g['sourceLockSha256'],'Source lock changed')
 spec=dict(cfg['commands'][name]);require(spec['supervisorSeconds']==120 and spec['automaticRetries']==0,'Bounds changed')
 for prior in order[:order.index(name)]:completed(prior)
 if spec['phase'] in ['installed','adjudication']:
  proof=json.loads(proof_for(spec['mode']).read_text())
  require(proof['sourceCommit']==g['finalHead'] and proof['sourceTree']==g['finalTree'],'Proof source differs')
  consumer=Path(proof['consumer']).resolve(strict=True);entry=Path(proof['entryRealpath']).resolve(strict=True)
  require(entry==(consumer/'node_modules/@bdp/server/dist/transaction-installed.js').resolve(strict=True) and entry.is_relative_to(consumer),'Entry provenance')
  require(sha(entry.read_bytes())==proof['entrySha256'] and sha((consumer/'pnpm-lock.yaml').read_bytes())==proof['consumerLockSha256']==g['consumerLockSha256'],'Entry/lock differs')
  if spec['phase']=='installed':
   spec['validatedBdpOverrides']={'BDP_TRANSACTIONAL_MINIMUM_CURRENT_MODE':spec['mode'],'BDP_TRANSACTIONAL_INSTALLED_ENTRY':str(entry),'BDP_TRANSACTIONAL_SOURCE_COMMIT':g['finalHead']}
 return {**g,'sourceSha256':source,'sourceDiffSha256':source_patch_hash(),'preparationSha256':sha((P/'supervise.py').read_bytes())},spec
