"""Closed exact conditional-branch gate. Root supplies actual identities and phase releases."""
from pathlib import Path
import hashlib,json,subprocess,os
P=Path(__file__).resolve().parent;D=P;L=Path('/Users/dbox/repos/wt/vickie/bdp-integration-logs')
W=Path('/Users/dbox/repos/wt/vickie/vickie-transaction-conditional-reads');E=P/'run';GUARD=L/'guarded-action.py';COMMANDS=P/'commands.json'
COMMANDS_SHA='3bd95f22807708ef117cc3d5321635d74764bc81857101a23486333924c8c510'
sha=lambda b:hashlib.sha256(b).hexdigest()
def re_full_sha(s):return len(s)==40 and all(c in "0123456789abcdef" for c in s)
def require(ok,message):
 if not ok:raise RuntimeError(message)
def git(*args):return subprocess.check_output(['git',*args],cwd=W,timeout=10)
def document(g,key):
 d=g[key];p=Path(d['path']);require(p.is_absolute() and p.is_relative_to(P) and not p.is_symlink() and sha(p.read_bytes())==d['sha256'],'Missing/changed '+key)
 return json.loads(p.read_text())
def gate():return json.loads((P/'execution-gate.json').read_text())
def input_files():return {p:sha((W/p).read_bytes()) for p in git('ls-files').decode().splitlines()}
def source_hashes():
 g=gate();b=document(g,'sourceBinding')
 require(git('rev-parse','HEAD').decode().strip()==b['head']==g['finalHead'],'Reviewed draft HEAD changed')
 require(git('rev-parse','HEAD^{tree}').decode().strip()==b['tree']==g['finalTree'],'Reviewed draft base tree changed')
 require(set(git('diff','--name-only','HEAD').decode().splitlines())==set(b['files']),'Reviewed draft scope changed')
 require(git('status','--porcelain').decode()==b['status'],'Reviewed draft status changed')
 actual={p:sha((W/p).read_bytes()) for p in b['files']}
 require(actual=={p:v['sha256'] for p,v in b['files'].items()},'Reviewed draft bytes changed')
 require(sha(git('diff','--binary','HEAD'))==b['diffSha256'],'Reviewed draft patch changed')
 require(document(g,'inputBinding')==input_files(),'Tracked preflight input changed')
 return actual
def source_patch_hash():return sha(git('diff','--binary','HEAD'))
def build_outputs():
 files=[W/'LICENSE']
 for name in ['protocol','server']:
  root=W/'packages'/name;files.append(root/'package.json')
  for sub in (['dist','schemas'] if name=='protocol' else ['dist']):files.extend(p for p in (root/sub).rglob('*') if p.is_file())
 require(files and all(not p.is_symlink() for p in files),'Missing/linked package members')
 return {p.relative_to(W).as_posix():sha(p.read_bytes()) for p in sorted(files)}
def emitted_files():return {p.relative_to(W).as_posix():sha(p.read_bytes()) for root in (W/'packages').iterdir() if root.is_dir() for p in (root/'dist').rglob('*') if p.is_file()}
def completed(name):
 x=json.loads((E/(name+'.exit.json')).read_text());cfg=json.loads(COMMANDS.read_text())['commands'][name]
 require(x['exit']==cfg.get('rawExpectedExit',0) and x['supervisorExit']==0 and x['sourceUnchanged'],'Prior stage failed '+name)
 for k in ['timeout','forcedCleanup','errors','remainingAfterCleanup','sampledPidsStillPresent','runtimeDirectoriesStillPresent']:require(not x[k],'Unclean stage '+name+':'+k)
 require(x['sourceSha256After']==source_hashes(),'Prior source mismatch '+name)
 return x
def verify_build(g):
 i=json.loads((E/'build-output-inventory.json').read_text());completed('fresh-build');completed('build-inventory')
 require(i['head']==g['finalHead'] and i['tree']==g['finalTree'] and i['inputFiles']==input_files(),'Conditional build attribution differs')
 require(i['packageFiles']==build_outputs() and i['emittedFiles']==emitted_files(),'Conditional outputs changed')
 for k in ['buildStart','buildExit','buildLog']:
  v=i[k];require(sha(Path(v['path']).read_bytes())==v['sha256'],'Build record changed')
def proof_for(mode=None):return E/'installed-provenance.json'
def consumer_inventory():
 C=E/'consumer';files={};links={}
 for directory,dirs,names in os.walk(C,followlinks=False):
  dirs[:]=[n for n in dirs if n not in ['.vite','.vite-temp']]
  for name in dirs+names:
   p=Path(directory)/name;r=p.relative_to(C).as_posix()
   if p.is_symlink():
    target=p.resolve(strict=True);require(target.is_relative_to(C.resolve()),'Consumer link escapes')
    links[r]={'link':os.readlink(p),'target':target.relative_to(C.resolve()).as_posix()}
   elif p.is_file():files[r]=sha(p.read_bytes())
 return {'files':files,'links':links}
def load_gate(name):
 require(__debug__,'Assertions required');g=gate();cfg=json.loads(COMMANDS.read_text())
 require(g['released'] is True and g['session']=='2026-10-08T20:26:35Z-81e5','Execution closed')
 a=document(g,'rootAcceptance');require(a['accepted'] is True and a['head']==g['finalHead'],'Root acceptance absent')
 require(sha(COMMANDS.read_bytes())==COMMANDS_SHA and sha(GUARD.read_bytes())==g['guardSha256'],'Commands/guard changed')
 for n,h in g['preparationSha256'].items():require(sha((P/n).read_bytes())==h,'Preparation changed '+n)
 order=cfg['executionOrder'];approved=g['approvedCommandIds'];require(approved==order[:len(approved)] and name in approved,'Stage not released')
 # Package verification is a separate root release after actual consumer installation.
 if 'verify-installed' in approved:
  lock=g.get('consumerLockSha256');actualLock=E/'consumer/pnpm-lock.yaml'
  require(isinstance(lock,str) and len(lock)==64 and actualLock.is_file() and sha(actualLock.read_bytes())==lock,'Root must accept generated consumer lock before releasing verify-installed')
 spec=dict(cfg['commands'][name]);require(g['phaseRelease'][spec['phase']] is True,'Phase closed')
 if spec['phase']=='source':
  require(a.get('sourceValidationAuthorized') is True and a.get('nativeCodeReviewed') is True and a.get('externalReviewPending') is True and a.get('diagnosticOnly') is True,'Diagnostic native-review authorization absent')
  require(g.get('councilAccepted') is False and a.get('sourceCouncilAccepted') is not True and a.get('councilAccepted') is not True,'Diagnostic execution must not claim complete council')
 else:
  require(g['councilAccepted'] is True,'Full council still pending for control/package/installed')
 require(spec['phase']=='source' and a.get('reviewedDirtySourceOnly') is True,'Only reviewed dirty preflight permitted')
 source=source_hashes()
 for prior in order[:order.index(name)]:completed(prior)
 require(spec['supervisorSeconds']==120 and spec['automaticRetries']==0,'Limits differ')
 require(set(g['toolHashes'])=={'/Users/dbox/.nvm/versions/node/v24.16.0/bin/node','/Users/dbox/.cache/node/corepack/v1/pnpm/11.20.0/bin/pnpm.cjs',str(W/'node_modules/vitest/vitest.mjs'),str(W/'node_modules/@biomejs/biome/bin/biome'),'/opt/homebrew/opt/python@3.14/bin/python3.14'},'Exact tool pins absent')
 for p,h in g['toolHashes'].items():require(sha(Path(p).read_bytes())==h,'Tool changed')
 require(sha((W/'pnpm-lock.yaml').read_bytes())==g['sourceLockSha256'],'Source lock changed')
 if spec['phase']!='source':verify_build(g)
 if spec['phase'] in ['installed','adjudication']:
  require(sha((E/'installed-consumer-inventory.json').read_bytes())==g['consumerInventorySha256'],'Consumer inventory not bound')
  proof=json.loads(proof_for().read_text());require(sha(proof_for().read_bytes())==g['installedProvenanceSha256'],'Installed provenance not released')
  require(proof['sourceCommit']==g['finalHead'] and proof['sourceTree']==g['finalTree'],'Installed source mismatch')
  C=Path(proof['consumer']).resolve(strict=True)
  require(sha((C/'pnpm-lock.yaml').read_bytes())==g['consumerLockSha256']==proof['consumerLockSha256'],'Consumer lock changed')
  entry=(C/'node_modules/@bdp/server/dist'/'transaction-installed-process.js').resolve(strict=True)
  require(entry.is_relative_to(C),'Workspace fallback')
  tar=json.loads((E/'tar-source-provenance.json').read_text());info=tar['packages']['server']['files']['dist/'+entry.name]
  require(sha(entry.read_bytes())==info['sha256'],'Entry differs from archive')
  if spec['phase']=='installed':
   spec['validatedBdpOverrides']={'BDP_TRANSACTIONAL_INSTALLED_ENTRY':str(entry),'BDP_TRANSACTIONAL_SOURCE_COMMIT':g['finalHead']}
   report=Path(spec['observationPath']);require(report.is_absolute() and report.is_relative_to(E) and not report.exists(),'Fresh exclusive observation required')
   for k,v in spec.get('reportOverrides',{}).items():
    require(k in ['BDP_TRANSACTIONAL_CONDITIONAL_OUTPUT','BDP_TRANSACTIONAL_OBSERVATION_OUTPUT','BDP_TRANSACTIONAL_CROSS_CARRIER_CONTROL_OUTPUT','BDP_TRANSACTIONAL_PRINCIPAL_CUMULATIVE_OUTPUT'] and Path(v).is_relative_to(E),'Unexpected report override')
   spec['validatedBdpOverrides'].update(spec.get('reportOverrides',{}))
 spec.setdefault('workload',{k:spec[k] for k in ['cases','http','expectedHTTP','maxHTTP','owners','listeners','directories','productChildren'] if k in spec})
 return {**g,'sourceSha256':source,'sourceDiffSha256':source_patch_hash(),'preparationSha256':sha((P/'supervise.py').read_bytes())},spec
