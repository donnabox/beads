"""Closed singleton lint-error observation gates; no product imports or quality/pipeline authority."""
from pathlib import Path
import hashlib,json,subprocess,os
P=Path(__file__).resolve().parent
W=Path('/Users/dbox/repos/wt/vickie/vickie-schema-minimum')
GUARD=P/'guarded-action-session-20261018-021624-549f.py'
MODE=os.environ.get('TX_DISCOVERY_VALIDATION_PHASE','final')
assert MODE=='final'
E=P/'run'
INTEGRATION='cce884ccea1db97cbc6fd7bff876d4d6091f9081'
sha=lambda b:hashlib.sha256(b).hexdigest()
def require(ok,message):
 if not ok:raise RuntimeError(message)
def git(*args):return subprocess.check_output(['git',*args],cwd=W,timeout=10)
def unique_object(pairs):
 d={}
 for k,v in pairs:
  if k in d:raise RuntimeError('Duplicate JSON key '+k)
  d[k]=v
 return d
def invalid_constant(value):raise RuntimeError('Invalid JSON constant '+value)
def read(p):return json.loads(Path(p).read_text(),object_pairs_hook=unique_object,parse_constant=invalid_constant)
def write(n,v):
 with (E/n).open('x') as f:f.write(json.dumps(v,indent=2)+'\n')
def guarded():subprocess.run([str(Path('/opt/homebrew/opt/python@3.14/bin/python3.14')),str(GUARD),'/usr/bin/true'],check=True,timeout=30)
def document(record):
 require(type(record) is dict and set(record)=={'path','sha256'},'Unknown/missing document pin key')
 p=Path(record['path']);require(p.is_absolute() and p.is_file() and not p.is_symlink() and sha(p.read_bytes())==record['sha256'],'Document absent/changed '+str(p));return read(p)
def source_check(b):
 require(b['worktree']==str(W) and b['head']==INTEGRATION and git('branch','--show-current').decode().strip()==b['branch'],'Wrong source binding')
 require(git('rev-parse','HEAD').decode().strip()==b['head'] and git('rev-parse','HEAD^{tree}').decode().strip()==b['tree'],'Git identity changed')
 actual={p:sha((W/p).read_bytes()) for p in git('ls-files').decode().splitlines()}
 require(actual==b['trackedSourceFiles'],'Source files changed')
 require(sorted(git('diff','--name-only','HEAD').decode().splitlines())==b['dirtyPaths']==sorted(b['ownedFiles']) and not git('ls-files','--others','--exclude-standard').strip(),'Source scope differs')
 require(all(actual[p]==v['afterSha256'] and sha(git('show','HEAD:'+p))==v['beforeSha256'] for p,v in b['ownedFiles'].items()),'Owned source binding differs')
 require(not git('diff','--cached','--binary').strip(),'Staged source is forbidden')
 require(len(actual)==679 and len(b['ownedFiles'])==10,'Exact tracked/owned counts')
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
 require(v['accepted'] and v['exit'] in read(P/'commands.json')['commands'][name].get('acceptedRawExitCodes',[0]) and not v['forcedCleanup'] and not v['errors'] and not v['remainingProcesses'] and not v['runtimeDirectoriesStillPresent'],'Prior stage failed '+name)
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
KEYSET={'rootArchiveAcceptance', 'rootConsumerProofAcceptance', 'rootBuildAcceptance', 'knownOriginalActualCodeTerminal', 'session', 'rootPlanAccepted', 'released', 'actualClaudeReviewDisposed', 'baseTools', 'bootstrapReuse', 'nativeCodeDisposition', 'sourceBinding', 'staticValidationAuthorized', 'nativeCodeReviewDisposed', 'installedProvenance', 'workspaceTools', 'consumerInventory', 'actualCodeDisposition', 'consumerLockSha256', 'approvedStages', 'implementationIndexSha256', 'mode', 'knownCorrectedSourceFreeze', 'rootConsumerLockAcceptance', 'rootDisposition', 'rootExactSourceAccepted'}
EXPECTED_SPEC={'argv': ['/Users/dbox/.nvm/versions/node/v24.16.0/bin/node', '/Users/dbox/repos/wt/vickie/vickie-schema-minimum/node_modules/@biomejs/biome/bin/biome', 'lint', '--diagnostic-level=error', '--config-path=/Users/dbox/repos/wt/vickie/vickie-schema-minimum/biome.json', 'apps/bdptest/src/transactional-discovery-installed.test.ts', 'apps/bdptest/src/transactional-schema-installed.test.ts', 'packages/server/src/transaction-changefeed.test.ts', 'packages/server/src/transaction-http.test.ts', 'packages/server/src/transaction-http.ts', 'packages/server/src/transaction-installed-contracts.test.ts', 'packages/server/src/transaction-schema-http.test.ts', 'packages/server/src/transaction-schema-http.ts', 'packages/server/src/transaction-service.test.ts', 'packages/server/src/transaction-service.ts'], 'automaticRetries': 0, 'configurationFinalized': True, 'cwd': '/Users/dbox/repos/wt/vickie/vickie-schema-minimum', 'id': 'lint-error-observation', 'kind': 'fixed-diagnostic-command', 'mutatesSource': False, 'outerSeconds': 120, 'termAtSeconds': 115, 'expectedObservedRawExit': 1, 'qualityAcceptance': False}
COMMANDS_SHA='7850e90410ac64eb109fbbe51650bee67bb5a7afa934edfa069c6e5c8aaadc94'
WORKSPACE_SHA='30468c88674a41a54f8a366ca78c84dffc07f3e035a0d9bf32430300b6503552'
KNOWN_ORIGINAL={'path': '/Users/dbox/repos/wt/vickie/bdp-integration-logs/write-profile-G09-code-review-1/actual-review-terminal-1/closure.json', 'sha256': '440e88bd386604b67f6a27ce83f54d03d5b090e3262b11d6488161b45a6b6057'}
KNOWN_CORRECTED={'path': '/Users/dbox/repos/wt/vickie/bdp-integration-logs/write-profile-G09-source-correction-authoring-2/index.json', 'sha256': '784402dcfdd70603886928b49fb56f9ccc9134aa7a3229d29c0923dcc48edc14'}
def load_gate(name,child=False):
 require(__debug__,'Assertions required')
 require(name=='lint-error-observation' and child is False,'Diagnostic singleton only; no child helper entry')
 g=read(P/'execution-gate.json')
 require(set(g)==KEYSET,'Unknown/missing gate key')
 require(all(type(g[k]) is bool for k in ['released','rootPlanAccepted','rootExactSourceAccepted','actualClaudeReviewDisposed','nativeCodeReviewDisposed','staticValidationAuthorized']),'Gate boolean type')
 require(g['released'] is True and g['rootPlanAccepted'] is True and g['rootExactSourceAccepted'] is True and g['staticValidationAuthorized'] is True and g['session']=='2026-10-18T02:16:24Z-549f','Root diagnostic release absent')
 require(g['mode']==MODE=='final','Wrong phase')
 idx=P/'implementation-hashes.json';require(sha(idx.read_bytes())==g['implementationIndexSha256'],'Implementation index changed')
 for f,h in read(idx)['files'].items():require(sha((P/f).read_bytes())==h,'Implementation changed '+f)
 for f,h in read(P/'reference-pins.json').items():require(sha(Path(f).read_bytes())==h,'Reference changed '+f)
 b=document(g['sourceBinding']);sid=source_check(b);g={**g,'finalHead':b['head'],'finalTree':b['tree'],'sourceLockSha256':b['trackedSourceFiles']['pnpm-lock.yaml'],'_sourceIdentity':sid,'_sourceBinding':b}
 require(g['sourceBinding']=={'path':str(P/'source-binding.json'),'sha256':'712c3c5a8e1543bf5005cbb54d321c6874b7443afb7e01805f393434fe5cd714'},'Exact diagnostic source binding required')
 cfg=read(P/'commands.json')
 require(set(cfg)=={'configurationFinalized','diagnosticOnly','noPipelineCredit','finalOrder','commands'} and cfg['configurationFinalized'] is True and cfg['diagnosticOnly'] is True and cfg['noPipelineCredit'] is True,'Diagnostic command scope differs')
 require(cfg['finalOrder']==['lint-error-observation'] and type(g['approvedStages']) is list and g['approvedStages']==['lint-error-observation'] and set(cfg['commands'])=={'lint-error-observation'},'Only one diagnostic is released')
 spec=cfg['commands'][name]
 require(spec==EXPECTED_SPEC,'Diagnostic command/argv/limits changed')
 require(spec['outerSeconds']==120 and spec['termAtSeconds']==115 and spec['automaticRetries']==0 and spec['expectedObservedRawExit']==1 and type(spec['expectedObservedRawExit']) is int and spec['qualityAcceptance'] is False,'Diagnostic limits/outcome changed')
 require(not any('UNRESOLVED' in arg or ('$'+'{') in arg or 'PENDING_SUCCESSOR' in arg for arg in spec['argv']),'Unbound command')
 disposition=document(g['rootDisposition'])
 require(disposition.get('accepted') is True and disposition.get('sourceBindingSha256')==g['sourceBinding']['sha256'] and disposition.get('commandsSha256')==COMMANDS_SHA,'Root diagnostic disposition differs')
 require(disposition.get('diagnosticOnly') is True and disposition.get('noPipelineCredit') is True and disposition.get('approvedStages')==['lint-error-observation'] and disposition.get('qualityAcceptance') is False,'No pipeline/quality authority allowed')
 require(disposition.get('staticValidationAuthorized') is True and disposition.get('runtimeValidationAuthorized') is False and g['bootstrapReuse']==[],'Only isolated static diagnostic authority is allowed')
 base=document(g['baseTools']);tools_check(base['tools'])
 require(base['sourceLockSha256']==g['sourceLockSha256'],'Base lock differs')
 require(g['workspaceTools']=={'path':str(P/'workspace-tools.json'),'sha256':WORKSPACE_SHA},'Exact previously observed workspace tool identity required')
 workspace=document(g['workspaceTools']);tools_check(workspace['tools'])
 require(workspace['sourceIdentity']==g['_sourceIdentity'] and workspace['sourceLockSha256']==g['sourceLockSha256'],'Workspace tool source/lock differs')
 require(g['actualClaudeReviewDisposed'] is True and g['nativeCodeReviewDisposed'] is True,'Both existing CODE disposition wrappers remain required')
 for label in ['nativeCodeDisposition','actualCodeDisposition']:
  d=document(g[label]);require(d.get('accepted') is True and d.get('sourceBindingSha256')==g['sourceBinding']['sha256'],'CODE disposition differs '+label)
 require(g['knownOriginalActualCodeTerminal']==KNOWN_ORIGINAL and g['knownCorrectedSourceFreeze']==KNOWN_CORRECTED,'Original/corrected review attribution changed')
 document(g['knownOriginalActualCodeTerminal']);document(g['knownCorrectedSourceFreeze'])
 require(all(g[k]=={'path':'UNRESOLVED','sha256':'UNRESOLVED'} for k in ['rootBuildAcceptance','rootArchiveAcceptance','rootConsumerLockAcceptance','rootConsumerProofAcceptance','installedProvenance','consumerInventory']) and g['consumerLockSha256']=='UNRESOLVED','Build/consumer authority is not part of this diagnostic')
 # No prior-stage completion, build, archive, consumer or next-stage branch exists here.
 guarded();return g,spec
