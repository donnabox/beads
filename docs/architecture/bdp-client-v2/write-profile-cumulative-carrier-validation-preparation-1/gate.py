"""CLOSED carrier681 gate. Root owns all releases and fresh measured proofs. WIP."""
from pathlib import Path
import hashlib,json,subprocess,os,math,shutil
P=Path(__file__).resolve().parent
W=Path('/Users/dbox/repos/wt/vickie/vickie-cumulative-evidence-carrier-correction-minimum')
GUARD=P/'guarded-action-session-20261018-021624-549f.py'
MODE=os.environ.get('CARRIER_VALIDATION_PHASE','final')
assert MODE=='final'
E=P/'run'
REVIEWED_BASE_HEAD='2834151544ca6bdc105191665f466c58eee7930f'
sha=lambda b:hashlib.sha256(b).hexdigest()
def require(ok,message):
 if not ok:raise RuntimeError(message)
def git(*args):return subprocess.check_output(['/opt/homebrew/bin/git',*args],cwd=W,timeout=10)
def unique_object(pairs):
 d={}
 for k,v in pairs:
  if k in d:raise RuntimeError('Duplicate JSON key '+k)
  d[k]=v
 return d
def invalid_constant(value):raise RuntimeError('Invalid JSON constant '+value)
def read(p):return json.loads(Path(p).read_text(),object_pairs_hook=unique_object,parse_constant=invalid_constant)
def write(n,v):
 fd=os.open(E/n,os.O_WRONLY|os.O_CREAT|os.O_EXCL|os.O_CLOEXEC,0o600)
 with os.fdopen(fd,'w') as f:f.write(json.dumps(v,indent=2)+'\n')
def guarded():subprocess.run([str(Path('/opt/homebrew/opt/python@3.14/bin/python3.14')),str(GUARD),'/usr/bin/true'],check=True,timeout=30)
def document(record):
 require(type(record) is dict and set(record)=={'path','sha256'},'Unknown/missing document pin key')
 p=Path(record['path']);require(p.is_absolute() and p.is_file() and not p.is_symlink() and sha(p.read_bytes())==record['sha256'],'Document absent/changed '+str(p));return read(p)
def source_check(b):
 expected=read(P/'expected-source-map.json');semantic=read(P/'semantic-source-scope.json')
 require(b['worktree']==str(W)and git('branch','--show-current').decode().strip()==b['branch'],'Wrong source binding')
 require(git('rev-parse','HEAD').decode().strip()==b['head']and git('rev-parse','HEAD^{tree}').decode().strip()==b['tree'],'Actual Git identity changed')
 paths=git('ls-files','-z').decode().split('\0');paths=[n for n in paths if n]
 require(len(paths)==len(set(paths))==681 and set(paths)==set(expected),'Exact tracked source681')
 bytes_by_path={n:(W/n).read_bytes()for n in paths};actual={n:sha(data)for n,data in bytes_by_path.items()};require(actual==expected==b['trackedSourceFiles'],'Exact reviewed source bytes differ')
 require(sha(json.dumps(actual,sort_keys=True).encode())=='78e11962df1a744c685c9911d13c0bff3fa42fcdec9310b75ec8f83a17996c7f','Canonical source identity changed')
 dirty=sorted(git('diff','--name-only','HEAD').decode().splitlines());require(dirty==b['dirtyPaths']==sorted(b['ownedFiles'])and (dirty==[]or dirty==sorted(semantic['allowedPaths'])),'Only exact reviewed source scope is permitted')
 require(not git('ls-files','--others','--exclude-standard').strip()and not git('diff','--cached','--binary').strip(),'Untracked or staged source forbidden')
 require(sha(git('diff','--binary','HEAD'))==b['binaryDiffSha256'],'Actual root-bound source binary patch changed')
 if dirty:require(b['head']==REVIEWED_BASE_HEAD,'Dirty candidate must retain exact reviewed base')
 tree={}
 for item in git('ls-tree','-r','-z','HEAD').split(b'\0'):
  if not item:continue
  meta,path=item.split(b'\t',1);mode,kind,identity=meta.split();require(kind==b'blob'and mode in [b'100644',b'100755'],'Nonordinary tracked source')
  tree[path.decode()]={'gitBlob':identity.decode(),'mode':mode.decode()}
 require(set(tree)==set(actual),'HEAD source inventory differs')
 for n,data in bytes_by_path.items():
  if n in b['ownedFiles']:
   owned=b['ownedFiles'][n];require(owned=={'beforeSha256':semantic['before'][n],'afterSha256':expected[n]}and sha(git('show','HEAD:'+n))==owned['beforeSha256'],'Owned before/after differs')
  else:
   blob=hashlib.sha1(b'blob '+str(len(data)).encode()+b'\0'+data).hexdigest();require(blob==tree[n]['gitBlob'],'Protected source differs from actual HEAD')
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
 watched=read(E/(name+'.watchdog.json'));require(watched.get('accepted') is True and watched.get('attemptId')==v['attemptId'] and watched.get('ownershipUnknown') is False and watched.get('sourceBindingSha256')==g['sourceBinding']['sha256'],'Watchdog terminal refused '+name)
 outer=document(g['outerExitReceipts'].get(name,{}))
 elapsed=outer.get('wholeParentElapsedSeconds');require(type(elapsed)in [int,float] and math.isfinite(elapsed) and 0<=elapsed<115,'Whole parent terminal clock absent/late '+name)
 require(outer.get('accepted') is True and outer.get('rawExit')==0 and outer.get('stage')==name and outer.get('attemptId')==v['attemptId'] and outer.get('sourceBindingSha256')==g['sourceBinding']['sha256'] and outer.get('stageExitSha256')==sha((E/(name+'.exit.json')).read_bytes()) and outer.get('watchdogSha256')==sha((E/(name+'.watchdog.json')).read_bytes()),'Actual outer receipt absent/mismatch '+name)
 require(v.get('stdioEOF')is True,'Child stdio EOF missing '+name)
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
def verify_private_emit(g):
 proof=document(g['privateEmitProof']);accepted=document(g['rootPrivateEmitAcceptance'])
 require(proof['sourceIdentity']==g['_sourceIdentity']and proof['sourceBindingSha256']==g['sourceBinding']['sha256'],'Private compiler proof source differs')
 complete('private-verifier-emit',g);complete('prove-private-graph',g)
 require(accepted.get('accepted')is True and accepted.get('privateEmitProof')==g['privateEmitProof']and accepted.get('sourceBindingSha256')==g['sourceBinding']['sha256'],'Private emit/graph not root accepted')
 for record in proof['files']:
  q=Path(record['path']);require(q.is_file()and not q.is_symlink()and str(q.resolve(strict=True))==record['realpath']and sha(q.read_bytes())==record['sha256'],'Private emit dependency/catalog changed')
 return proof
def load_gate(name,child=False):
 require(__debug__,'Assertions required')
 g=read(P/'execution-gate.json');expected=set(read(P/'gate-keys.json'))
 require(type(g)is dict and set(g)==expected,'Missing/extra root execution gate key')
 require(g['released']is True and g['rootExactSourceAccepted']is True and g['staticValidationAuthorized']is True and g['session']=='2026-10-18T02:16:24Z-549f'and g['mode']=='final','Root execution release absent')
 cfg=read(P/'commands.json');require(cfg.get('configurationFinalized')is True,'Incomplete draft configuration');order=cfg['finalOrder'];approved=g['approvedStages'];require(type(approved)is list and approved==order[:len(approved)]and name in approved,'Fresh ordered prefix not released')
 require(g['bootstrapReuse']==[],'No bootstrap/old execution reuse')
 idx=P/'implementation-hashes.json';require(sha(idx.read_bytes())==g['implementationIndexSha256'],'Implementation index differs')
 for n,h in read(idx)['files'].items():require(sha((P/n).read_bytes())==h,'Implementation changed')
 for n,h in read(P/'reference-pins.json').items():require(sha(Path(n).read_bytes())==h,'Reference input changed')
 b=document(g['sourceBinding']);sid=source_check(b);g={**g,'finalHead':b['head'],'finalTree':b['tree'],'sourceLockSha256':b['trackedSourceFiles']['pnpm-lock.yaml'],'_sourceIdentity':sid,'_sourceBinding':b}
 root=document(g['rootDisposition']);require(root.get('accepted')is True and root.get('sourceBindingSha256')==g['sourceBinding']['sha256']and root.get('implementationIndexSha256')==g['implementationIndexSha256']and root.get('freshPipelineRequired')is True,'Root current source/mechanics disposition differs')
 require(root.get('originalActualCodeReviewDispositionAccepted')is True and root.get('laterCorrectedSourceNativeAccepted')is True,'Original actual CODE + corrected-source native/root dispositions required; no future-byte review inference')
 native=document(g['nativeHelperDisposition']);require(native.get('accepted')is True and native.get('implementationIndexSha256')==g['implementationIndexSha256'],'Independent full helper mechanics absent')
 spec=cfg['commands'][name]
 if spec.get('evaluatesSourceOrProduct')is True:require(g['runtimeValidationAuthorized']is True and root.get('runtimeValidationAuthorized')is True,'Source/private/product evaluator stage not released')
 require((spec['outerSeconds'],spec['termAtSeconds'],spec['killAtSeconds'],spec['workCutoffSeconds'],spec['automaticRetries'])==(120,115,119,105,0),'Fixed clock/retry policy differs')
 require(all(type(a)is str and 'UNRESOLVED'not in a and '${'not in a for a in spec['argv']),'Command argv not concrete')
 if child:
  started=E/(name+'.started.json');require(os.environ.get('CARRIER_VALIDATION_STAGE')==name and started.is_file()and sha(started.read_bytes())==os.environ.get('CARRIER_VALIDATION_ATTEMPT'),'Helper requires current supervised attempt')
  require(read(started)['rootGateSha256']==sha((P/'execution-gate.json').read_bytes()),'Gate changed during attempt')
 base=document(g['baseTools']);tools_check(base['tools']);require(str(Path(shutil.which('gh')or'').resolve(strict=True))==base['tools']['/opt/homebrew/bin/gh']['realpath'],'Guard must resolve exact pinned gh');require(base['sourceLockSha256']==g['sourceLockSha256'],'Base tool source lock differs')
 if order.index(name)>order.index('bind-workspace-tools'):
  fresh=document(g['workspaceTools']);tools_check(fresh['tools']);require(fresh['sourceIdentity']==sid and fresh['sourceLockSha256']==g['sourceLockSha256'],'Fresh workspace tool source differs')
 for prior in order[:order.index(name)]:complete(prior,g)
 if order.index(name)>order.index('bind-fresh-build'):
  verify_build(g);a=document(g['rootBuildAcceptance']);require(a.get('accepted')is True and a.get('sourceBindingSha256')==g['sourceBinding']['sha256']and a.get('buildMapSha256')==sha((E/'build-artifacts.json').read_bytes()),'Root fresh build acceptance absent')
 if order.index(name)>=order.index('prepare-consumer'):
  a=document(g['rootArchiveAcceptance']);require(a.get('accepted')is True and a.get('sourceBindingSha256')==g['sourceBinding']['sha256']and a.get('tarProvenanceSha256')==sha((E/'tar-source-provenance.json').read_bytes()),'Root actual archive acceptance absent')
 if name=='verify-installed-consumer':
  a=document(g['rootConsumerLockAcceptance']);require(a.get('accepted')is True and a.get('sourceBindingSha256')==g['sourceBinding']['sha256'],'Root actual generated lock acceptance absent')
  require(sha((E/'consumer/pnpm-lock.yaml').read_bytes())==g['consumerLockSha256']==a['consumerLockSha256']and a['consumerManifestSha256']==sha((E/'consumer/package.json').read_bytes())and a['consumerSettingsSha256']==sha((E/'consumer/pnpm-workspace.yaml').read_bytes()),'Actual consumer lock/manifest/settings differ')
 if spec.get('requiresConsumerProof')is True:
  proof=verify_consumer(g);a=document(g['rootConsumerProofAcceptance']);require(a.get('accepted')is True and a.get('sourceBindingSha256')==g['sourceBinding']['sha256']and a.get('installedProvenance')==g['installedProvenance']and a.get('consumerInventory')==g['consumerInventory'],'Root actual consumer proof absent')
  spec={**spec,'environment':{**spec.get('environment',{}),'BDP_TRANSACTIONAL_INSTALLED_ENTRY':proof['entryRealpath'],'BDP_TRANSACTIONAL_SOURCE_COMMIT':b['head']}}
 if spec.get('requiresPrivateEmitProof')is True:verify_private_emit(g)
 guarded();return g,spec
