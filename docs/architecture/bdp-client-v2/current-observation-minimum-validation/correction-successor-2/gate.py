"""Current-observation minimum closed gates; no product imports. Root owns releases and measured pins."""
from pathlib import Path
import hashlib,json,subprocess,os
P=Path(__file__).resolve().parent
W=Path('/Users/dbox/repos/wt/vickie/vickie-current-observation-minimum')
GUARD=Path('/Users/dbox/repos/wt/vickie/bdp-integration-logs/guarded-action.py')
MODE=os.environ.get('CURRENT_OBSERVATION_VALIDATION_PHASE','final')
assert MODE=='final'
E=P/'run'
INTEGRATION='d9c3eb9e61b1982b9bd56aa77e4cdc114384e886'
sha=lambda b:hashlib.sha256(b).hexdigest()
def require(ok,message):
 if not ok:raise RuntimeError(message)
def git(*args):return subprocess.check_output(['git',*args],cwd=W,timeout=10)
def read(p):return json.loads(Path(p).read_text())
def write(n,v):
 with (E/n).open('x') as f:f.write(json.dumps(v,indent=2)+'\n')
def guarded():subprocess.run([str(Path('/opt/homebrew/opt/python@3.14/bin/python3.14')),str(GUARD),'/usr/bin/true'],check=True,timeout=30)
def document(record):
 p=Path(record['path']);require(p.is_absolute() and p.is_file() and not p.is_symlink() and sha(p.read_bytes())==record['sha256'],'Document absent/changed '+str(p));return read(p)
def source_check(b):
 require(b['worktree']==str(W) and b['head']==INTEGRATION and git('branch','--show-current').decode().strip()==b['branch'],'Wrong source binding')
 require(git('rev-parse','HEAD').decode().strip()==b['head'] and git('rev-parse','HEAD^{tree}').decode().strip()==b['tree'],'Git identity changed')
 actual={p:sha((W/p).read_bytes()) for p in git('ls-files').decode().splitlines()}
 require(actual==b['trackedSourceFiles'],'Source files changed')
 require(sorted(git('diff','--name-only','HEAD').decode().splitlines())==b['dirtyPaths']==sorted(b['ownedFiles']) and not git('ls-files','--others','--exclude-standard').strip(),'Source scope differs')
 require(all(actual[p]==v['afterSha256'] and sha(git('show','HEAD:'+p))==v['beforeSha256'] for p,v in b['ownedFiles'].items()),'Owned source binding differs')
 require(sha(git('diff','--binary','HEAD'))==b['binaryDiffSha256'],'Source patch changed')
 require(all(sha(git('show','HEAD:'+f))==v for f,v in actual.items() if f not in b['ownedFiles']),'Other tracked source differs from HEAD')
 return sha(json.dumps(actual,sort_keys=True).encode())
def tools_check(t):
 for name,binding in t.items():
  p=Path(name);require(str(p.resolve(strict=True))==binding['realpath'] and sha(p.read_bytes())==binding['sha256'],'Tool changed '+name)
def inventory():
 result={}
 for group in ['packages','apps']:
  for package in sorted((W/group).iterdir()):
   if not package.is_dir():continue
   dist=package/'dist';require(dist.is_dir(),'Missing dist '+str(dist))
   for p in sorted(dist.rglob('*')):
    require(not p.is_symlink(),'Linked build output')
    if p.is_file():result[p.relative_to(W).as_posix()]=sha(p.read_bytes())
 for p in [W/'LICENSE',W/'packages/server/package.json',W/'packages/protocol/package.json',*(W/'packages/protocol/schemas').rglob('*')]:
  require(not p.is_symlink(),'Linked artifact')
  if p.is_file():result[p.relative_to(W).as_posix()]=sha(p.read_bytes())
 require(result,'Empty build inventory');return result
def consumer_inventory():
 C=E/'consumer';files={};links={}
 for directory,dirs,names in os.walk(C,followlinks=False):
  for n in dirs+names:
   p=Path(directory)/n;r=p.relative_to(C).as_posix()
   if p.is_symlink():
    q=p.resolve(strict=True);require(q.is_relative_to(C.resolve()),'Consumer link escapes');links[r]={'link':os.readlink(p),'target':q.relative_to(C.resolve()).as_posix()}
   elif p.is_file():files[r]=sha(p.read_bytes())
 return {'files':files,'links':links}
def complete(name,g):
 v=read(E/(name+'.exit.json'))
 require(v['accepted'] and v['exit']==0 and not v['forcedCleanup'] and not v['errors'] and not v['remainingProcesses'] and not v['runtimeDirectoriesStillPresent'],'Prior stage failed '+name)
 require(v['sourceIdentity']==g['_sourceIdentity'] and v['sourceUnchanged'],'Prior source differs '+name)
 require(sha((E/(name+'.log')).read_bytes())==v['logSha256'],'Prior log changed '+name)
 return v
def verify_build(g):
 b=read(E/'build-artifacts.json');complete('fresh-build',g)
 require(b['head']==g['finalHead'] and b['tree']==g['finalTree'] and b['sourceIdentity']==g['_sourceIdentity'],'Build source differs')
 require(b['artifacts']==inventory(),'Build outputs changed')
 for key in ['start','exit','log']:
  item=b[key];require(sha(Path(item['path']).read_bytes())==item['sha256'],'Build record changed')
def verify_consumer(g):
 proof=document(g['installedProvenance']);expected=document(g['consumerInventory'])
 require(expected==consumer_inventory(),'Consumer changed')
 require(proof['sourceCommit']==g['finalHead'] and proof['sourceTree']==g['finalTree'] and proof['sourceIdentity']==g['_sourceIdentity'] and proof['sourceBindingSha256']==g['sourceBinding']['sha256'],'Consumer source differs')
 require(sha((E/'consumer/pnpm-lock.yaml').read_bytes())==g['consumerLockSha256']==proof['consumerLockSha256'],'Consumer lock changed')
 tar=read(E/'tar-source-provenance.json');require(sha((E/'tar-source-provenance.json').read_bytes())==proof['tarProvenanceSha256'],'Tar provenance changed')
 for name,data in tar['packages'].items():require(sha(Path(data['archive']).read_bytes())==proof['packages'][name]['archiveSha256'],'Archive changed')
 entry=Path(proof['entryRealpath']);require(entry.resolve(strict=True).is_relative_to((E/'consumer').resolve()) and sha(entry.read_bytes())==proof['entrySha256'],'External entry changed')
 return proof
def load_gate(name,child=False):
 require(__debug__,'Assertions required')
 g=read(P/'execution-gate.json')
 require(read(P/'commands.json').get('configurationFinalized') is True,'Authored Current-observation minimum source/selection not finalized')
 require(read(P/'audit-expectations.json').get('finalized') is True,'Actual Current-observation minimum observation contracts not finalized')
 require(g['released'] is True and g['rootPlanAccepted'] is True and g['rootExactSourceAccepted'] is True and g['session']=='2026-10-08T20:26:35Z-81e5','Root release absent')
 require(g['mode']==MODE,'Wrong phase')
 if child:
  record=E/(name+'.started.json')
  require(os.environ.get('CURRENT_OBSERVATION_VALIDATION_STAGE')==name and record.is_file() and sha(record.read_bytes())==os.environ.get('CURRENT_OBSERVATION_VALIDATION_ATTEMPT'),'Helper requires supervised attempt')
  require(read(record)['rootGateSha256']==sha((P/('bootstrap-gate.json' if MODE=='bootstrap' else 'execution-gate.json')).read_bytes()),'Gate changed during attempt')
 idx=P/'implementation-hashes.json';require(sha(idx.read_bytes())==g['implementationIndexSha256'],'Implementation index changed')
 for f,h in read(idx)['files'].items():require(sha((P/f).read_bytes())==h,'Implementation changed '+f)
 for f,h in read(P/'reference-pins.json').items():require(sha(Path(f).read_bytes())==h,'Reference changed '+f)
 b=document(g['sourceBinding']);sid=source_check(b);g={**g,'finalHead':b['head'],'finalTree':b['tree'],'sourceLockSha256':b['trackedSourceFiles']['pnpm-lock.yaml'],'_sourceIdentity':sid,'_sourceBinding':b}
 disposition=document(g['rootDisposition']);require(disposition['accepted'] is True and disposition['sourceBindingSha256']==g['sourceBinding']['sha256'],'Root disposition differs')
 cfg=read(P/'commands.json');order=cfg[MODE+'Order'];approved=g['approvedStages']
 require(approved==order[:len(approved)] and name in approved,'Stage not released')
 spec=cfg['commands'][name];require(spec['outerSeconds']==120 and spec['automaticRetries']==0,'Limits changed')
 require(not any('PENDING_SUCCESSOR' in arg for arg in spec['argv']),'Unbound command')
 if 'selectedFullNames' in spec:require(isinstance(spec['selectedFullNames'],list) and len(spec['selectedFullNames'])>0 and isinstance(spec['prospectiveWorkload'],dict),'Unbound callback selection/workload')
 base=document(g['baseTools']);tools_check(base['tools'])
 require(base['sourceLockSha256']==g['sourceLockSha256'],'Base lock differs')
 if name in ['workspace-install','bind-workspace-tools']:
  require(disposition.get('dependencyInstallAuthorized') is True,'Dependency preparation not released')
  if name=='workspace-install':require(not (W/'node_modules').exists(),'Workspace install already exists; preserve it and rebind explicitly')
 else:
  tb=document(g['workspaceTools']);tools_check(tb['tools'])
  require(tb['sourceIdentity']==g['_sourceIdentity'] and tb['sourceLockSha256']==g['sourceLockSha256'] and tb['head']==g['finalHead'],'Installed tools source differs')
  require(tb['installExitSha256']==sha((E/'workspace-install.exit.json').read_bytes()) and tb['installLogSha256']==sha((E/'workspace-install.log').read_bytes()),'Install evidence changed')
  require((g['actualClaudeReviewDisposed'] is True and g['nativeCodeReviewDisposed'] is True) or (name in ['typecheck','format-check','lint'] and g['staticValidationAuthorized'] is True),'Code review disposition pending; only explicit static prefix may run')
  require(disposition['staticValidationAuthorized'] is True if name in ['typecheck','format-check','lint'] else disposition['runtimeValidationAuthorized'] is True,'Root disposition does not authorize this stage')
 for prior in order[:order.index(name)]:complete(prior,g)
 if MODE=='final' and order.index(name)>order.index('bind-fresh-build'):verify_build(g)
 if name=='verify-installed-consumer':require(isinstance(g.get('consumerLockSha256'),str) and sha((E/'consumer/pnpm-lock.yaml').read_bytes())==g['consumerLockSha256'],'Consumer lock not root-pinned')
 if name in ['receiving-history-external', 'review-receiving-history-external', 'current-observation-external', 'review-current-observation-external']:
  proof=verify_consumer(g)
  spec={**spec,'environment':{'BDP_TRANSACTIONAL_INSTALLED_ENTRY':proof['entryRealpath'],'BDP_TRANSACTIONAL_SOURCE_COMMIT':b['head']}}
 guarded();return g,spec
