"""Closed external installed validation gate; no product imports or implicit release."""
from pathlib import Path
import hashlib,json,re,subprocess
P=Path(__file__).resolve().parent;D=P.parent;L=D.parent
W=Path('/Users/dbox/repos/wt/vickie/vickie-authorization-transition')
E=P/'run';GUARD=L/'guarded-action.py';COMMANDS=P/'commands.json'
COMMANDS_SHA='4bb4bb7451719eb09843c6907834ca100bcadb3bd6cfaf171c032456bec1861e'
GUARD_SHA='fd81ba029cd4e7c7e60623df89e8292af91141d311d9ce789ed19e2152375e84'
sha=lambda b:hashlib.sha256(b).hexdigest()
def require(ok,message):
 if not ok:raise RuntimeError(message)
def git(*args):return subprocess.check_output(['git',*args],cwd=W,timeout=10)
def document(g,key):
 d=g[key];p=Path(d['path'])
 require(p.is_relative_to(L) and not p.is_symlink() and sha(p.read_bytes())==d['sha256'],'Changed bound document: '+key)
 return json.loads(p.read_text())
def source_binding():
 g=json.loads((P/'execution-gate.json').read_text());b=document(g,'sourceBinding')
 require(b==json.loads((P/'prepared-source.json').read_text()),'Source binding differs')
 return b
def source_hashes():
 expected={p:v['sha256'] for p,v in source_binding()['files'].items()}
 require(len(expected)==15 and set(expected)==set(json.loads(COMMANDS.read_text())['ownedPaths']),'Wrong source scope')
 actual={p:sha((W/p).read_bytes()) for p in expected}
 require(actual==expected,'Source bytes changed')
 require(not git('status','--porcelain').strip(),'Clean committed source required')
 return actual
def source_patch_hash():return sha(git('diff','--binary','HEAD'))
def input_files():return {p:sha((W/p).read_bytes()) for p in git('ls-files').decode().splitlines()}
def build_outputs():
 files=[W/'LICENSE']
 for name in ['protocol','server']:
  root=W/'packages'/name;files.append(root/'package.json')
  for sub in (['dist','schemas'] if name=='protocol' else ['dist']):
   files.extend(p for p in (root/sub).rglob('*') if p.is_file())
 require(all(not p.is_symlink() for p in files),'Package symlink')
 return {p.relative_to(W).as_posix():sha(p.read_bytes()) for p in sorted(files)}
def emitted_files():
 return {p.relative_to(W).as_posix():sha(p.read_bytes()) for root in (W/'packages').iterdir() if root.is_dir() for p in (root/'dist').rglob('*') if p.is_file()}
def verify_build(g):
 inventory=document(g,'buildOutputInventory');accepted=document(g,'rootBuildAcceptance')
 require(accepted.get('accepted') is True and accepted['inventorySha256']==g['buildOutputInventory']['sha256'] and accepted['finalHead']==g['finalHead'],'Build acceptance pending')
 for k in ['buildStart','buildExit','buildLog']:
  v=inventory[k];require(sha(Path(v['path']).read_bytes())==v['sha256'],'Build evidence changed')
 start=json.loads(Path(inventory['buildStart']['path']).read_text());end=json.loads(Path(inventory['buildExit']['path']).read_text())
 require(start['sourceSha256']==inventory['sourceSha256']==source_hashes() and end['sourceSha256After']==inventory['sourceSha256'],'Build source differs')
 require(end['exit']==end['supervisorExit']==0 and end['sourceUnchanged'],'Build failed')
 for k in ['timeout','forcedCleanup','errors','remainingAfterCleanup','sampledPidsStillPresent','runtimeDirectoriesStillPresent']:require(not end[k],'Unclean build: '+k)
 require(start['argv']==[g['node'],g['pnpm'],'run','build'] and start['cwd']==str(W),'Wrong build command')
 require(inventory['inputFiles']==input_files() and inventory['packageFiles']==build_outputs() and inventory['emittedFiles']==emitted_files(),'Current input/output inventory differs')
 require(inventory['buildHead']==start['head'] and inventory['buildWorkspaceDirty']==start['workspaceDirty'],'Build chronology differs')
def completed(name):
 x=json.loads((E/(name+'.exit.json')).read_text())
 require(x['exit']==x['supervisorExit']==0 and x['sourceUnchanged'],'Prior stage failed: '+name)
 for k in ['timeout','forcedCleanup','errors','remainingAfterCleanup','sampledPidsStillPresent','runtimeDirectoriesStillPresent']:require(not x[k],'Unclean prior stage')
 require(x['sourceSha256After']==source_hashes(),'Prior source differs')
 return x
def load_gate(name):
 require(__debug__,'Assertions required')
 g=json.loads((P/'execution-gate.json').read_text())
 require(g.get('released') is True and g['session']=='2026-10-08T20:26:35Z-81e5','Root gate closed')
 root=document(g,'rootAcceptance');council=document(g,'councilState')
 require(root.get('accepted') is True and document(g,'sourceValidationAcceptance').get('accepted') is True,'Source/root acceptance absent')
 require(council.get('status')=='accepted' or (council.get('status')=='pending-actual-claude' and root.get('validationOnlyWhileCouncilPending') is True),'Council pending without validation release')
 names={'gate.py','supervise.py','packages.py','review-installed.py','commands.json','prepared-source.json','ledger.json','authorization-bootstrap.mjs','observe-installed.ts','launcher.mjs','raw-sse-recorder.ts'}
 require(set(g['preparationSha256'])==names,'Preparation pin set differs')
 for n,h in g['preparationSha256'].items():require(sha((P/n).read_bytes())==h,'Preparation changed: '+n)
 require(sha(COMMANDS.read_bytes())==COMMANDS_SHA and sha(GUARD.read_bytes())==GUARD_SHA,'Commands/guard changed')
 cfg=json.loads(COMMANDS.read_text());order=cfg['executionOrder']
 require(name in g['approvedCommandIds'] and g['approvedCommandIds']==order[:len(g['approvedCommandIds'])],'Unreleased/nonprefix stage')
 require(re.fullmatch('[0-9a-f]{40}',g.get('finalHead') or '') is not None,'Final candidate pending')
 require(git('rev-parse','HEAD').decode().strip()==g['finalHead'] and git('rev-parse','HEAD^{tree}').decode().strip()==g['finalTree'],'Final identity differs')
 require(git('show','-s','--format=%P','HEAD').decode().strip().split()==g['finalParents'],'Final ordered parents differ')
 git('merge-base','--is-ancestor','67a2842e3f1918e70978f745587cf11eaf0df5a8',g['finalHead'])
 git('merge-base','--is-ancestor',g['finalBase'],g['finalHead'])
 final=document(g,'finalSource');integration=document(g,'integrationAcceptance')
 require(integration.get('accepted') is True and integration['finalHead']==g['finalHead'] and integration['finalTree']==g['finalTree'] and integration['finalBase']==g['finalBase'] and integration['finalParents']==g['finalParents'],'Integration not accepted')
 require(final['head']==g['finalHead'] and final['tree']==g['finalTree'] and final['sourceBindingSha256']==g['sourceBinding']['sha256'],'Final source differs')
 source=source_hashes()
 require(set(git('diff','--name-only',g['finalBase'],'HEAD').decode().splitlines())==set(source),'Fifteen-path footprint differs')
 require(sha((W/'pnpm-lock.yaml').read_bytes())==g['sourceLockSha256'],'Source lock differs')
 require(g['node']=='/Users/dbox/.nvm/versions/node/v24.16.0/bin/node' and g['pnpm']=='/Users/dbox/.cache/node/corepack/v1/pnpm/11.20.0/bin/pnpm.cjs','Tool paths differ')
 verify_build(g)
 spec=dict(cfg['commands'][name]);require(spec['supervisorSeconds']==120 and spec['automaticRetries']==0,'Bound differs')
 for prior in order[:order.index(name)]:completed(prior)
 if spec['phase'] in ['installed','review']:
  proof=document(g,'installedProvenance')
  require(proof==json.loads((E/'installed-provenance.json').read_text()),'Provenance file differs')
  require(proof['sourceCommit']==g['finalHead'] and proof['sourceTree']==g['finalTree'],'Installed source differs')
  consumer=(E/'consumer').resolve(strict=True);entry=Path(proof['entryRealpath']).resolve(strict=True)
  require(entry==(consumer/'node_modules/@bdp/server/dist/transaction-installed.js').resolve(strict=True) and entry.is_relative_to(consumer),'Installed entry differs')
  require(sha(entry.read_bytes())==proof['entrySha256'] and proof['consumerLockSha256']==g['consumerLockSha256']==sha((consumer/'pnpm-lock.yaml').read_bytes()),'Installed entry/lock changed')
 return {**g,'sourceSha256':source,'sourceDiffSha256':source_patch_hash(),'preparationSha256':sha((P/'supervise.py').read_bytes())},spec
