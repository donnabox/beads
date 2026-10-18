"""Closed exact fresh-depth-branch gate. Root supplies actual identities and phase releases."""
from pathlib import Path
import hashlib,json,subprocess,os
P=Path(__file__).resolve().parent;D=P;L=P.parent.parent
W=Path('/Users/dbox/repos/wt/vickie/vickie-fresh-depth-serving');E=P/'run';GUARD=L/'guarded-action.py';COMMANDS=P/'commands.json'
COMMANDS_SHA='e9bb36f2509902d3b34ad844d1e25f9e0182f30b48ccd49a7403849b0dbdbeab'
sha=lambda b:hashlib.sha256(b).hexdigest()
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
 require(b['head']==g['finalHead'] and b['tree']==g['finalTree'] and b['workspace']==str(W),'Fresh-depth identity missing')
 require(git('rev-parse','HEAD').decode().strip()==b['head'] and git('rev-parse','HEAD^{tree}').decode().strip()==b['tree'],'Actual Git identity changed')
 require(not git('status','--porcelain').strip(),'Fresh-depth source must be clean')
 require(git('show','-s','--format=%P',b['head']).decode().split()==b['parents'],'Ordered fresh-depth parents changed')
 prepared_document=json.loads((P/'prepared-source.json').read_text())
 require(isinstance(prepared_document['finalBase'],str) and len(prepared_document['finalBase'])==40 and b['diffBase']==prepared_document['finalBase'],'Actual final branch base pending/changed')
 git('merge-base','--is-ancestor',prepared_document['reviewBase'],b['diffBase'])
 git('merge-base','--is-ancestor',b['diffBase'],b['head'])
 prepared=prepared_document['files']
 require(len(prepared)==18 and all(isinstance(v,str) and len(v)==64 for v in prepared.values()),'Final18 source hashes pending')
 require(set(git('diff','--name-only',b['diffBase'],b['head']).decode().splitlines())==set(prepared)==set(b['files']),'Final18 scope differs')
 require({p:v['sha256'] for p,v in b['files'].items()}==prepared,'Final source differs from reviewed bytes')
 actual={p:sha((W/p).read_bytes()) for p in b['files']}
 require(actual=={p:v['sha256'] for p,v in b['files'].items()},'Fresh-depth source hashes changed')
 require(sha(git('diff','--binary',b['diffBase'],b['head']))==b['diffSha256'],'Fresh-depth patch changed')
 require(document(g,'inputBinding')==input_files(),'Tracked fresh-depth inputs changed')
 if (E/'installed-consumer-inventory.json').exists():
  require(json.loads((E/'installed-consumer-inventory.json').read_text())==consumer_inventory(),'Consumer graph changed')
 return actual
def source_patch_hash():
 b=document(gate(),'sourceBinding');return sha(git('diff','--binary',b['diffBase'],b['head']))
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
 require(i['head']==g['finalHead'] and i['tree']==g['finalTree'] and i['inputFiles']==input_files(),'Fresh depth build attribution differs')
 require(i['packageFiles']==build_outputs() and i['emittedFiles']==emitted_files(),'Fresh-depth outputs changed')
 for k in ['buildStart','buildExit','buildLog']:
  v=i[k];require(sha(Path(v['path']).read_bytes())==v['sha256'],'Build record changed')
def proof_for(mode=None):return E/'installed-provenance.json'
def verify_mirror():
 m=json.loads((E/'control-mirror.json').read_text());C=P/'mirror'
 require(m['combinedHead']==gate()['finalHead'],'Wrong control source')
 require(m['changedFiles']==['packages/server/dist/transaction-snapshot.js'],'Wrong mutation scope')
 for n,h in m['positiveFiles'].items():require(sha((W/n).read_bytes())==h,'Control positive changed '+n)
 for n,h in m['copyFiles'].items():require(sha((C/n).read_bytes())==h,'Control copy changed '+n)
 for n,v in m['links'].items():
  p=C/n;require(p.is_symlink() and os.readlink(p)==v['link'] and str(p.resolve(strict=True))==v['realpath'] and p.resolve().is_relative_to(C),'Control link escape/change')
 return m
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
 spec=dict(cfg['commands'][name]);require(g['phaseRelease'][spec['phase']] is True,'Phase closed')
 if spec['phase']=='source':
  require(a.get('sourceValidationAuthorized') is True and a.get('sourceCouncilAccepted') is True,'Diagnostic source authorization/council absent')
 else:
  require(g['councilAccepted'] is True,'Full council still pending for control/package/installed')
 source=source_hashes()
 for prior in order[:order.index(name)]:completed(prior)
 require(spec['supervisorSeconds']==120 and spec['automaticRetries']==0,'Limits differ')
 require(set(g['toolHashes'])=={'/Users/dbox/.nvm/versions/node/v24.16.0/bin/node','/Users/dbox/.cache/node/corepack/v1/pnpm/11.20.0/bin/pnpm.cjs',str(W/'node_modules/vitest/vitest.mjs'),str(W/'node_modules/@biomejs/biome/bin/biome'),'/opt/homebrew/opt/python@3.14/bin/python3.14'},'Exact tool pins absent')
 for p,h in g['toolHashes'].items():require(sha(Path(p).read_bytes())==h,'Tool changed')
 require(sha((W/'pnpm-lock.yaml').read_bytes())==g['sourceLockSha256'],'Source lock changed')
 if spec['phase']!='source':verify_build(g)
 if name in ['strict-control','review-control']:verify_mirror()
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
 spec['workload']={k:spec[k] for k in ['cases','http','expectedHTTP','maxHTTP','owners','listeners','directories','productChildren'] if k in spec}
 return {**g,'sourceSha256':source,'sourceDiffSha256':source_patch_hash(),'preparationSha256':sha((P/'supervise.py').read_bytes())},spec
