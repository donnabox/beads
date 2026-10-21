from pathlib import Path
import json,hashlib,stat,subprocess
L=Path('/Users/dbox/repos/wt/vickie/bdp-integration-logs');C=Path('/Users/dbox/repos/wt/vickie/bdp-client-design/checkout');D=C/'docs/architecture/bdp-client-v2';P=L/'write-profile-G09-validation-preparation-5';S=L/'write-profile-G09-validation5-consumer2-closed-control-source-1';E=L/'write-profile-G09-validation5-consumer-install-root-tool-evidence-1';R=L/'write-profile-G09-validation5-root-first4-dispatch-1/consumer-install';stage='consumer-install'
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def strict(p):
 def pairs(v):
  d=dict(v);assert len(d)==len(v);return d
 return json.loads(p.read_text(),object_pairs_hook=pairs,parse_constant=lambda v:(_ for _ in ()).throw(ValueError(v)))
def meta(p):return {'bytes':p.stat().st_size,'sha256':sha(p),'mode':oct(stat.S_IMODE(p.stat().st_mode))}
eix=strict(E/'index.json');assert len(eix['files'])==8
for n,m in eix['files'].items():assert meta(E/n)==m and meta(E/n)==meta(D/E.name/n),n
b=strict(E/'binding.json')
for n,m in b['actualRawFiles'].items():assert meta(Path(n))==m and meta(Path(n))==meta(D/Path(n).relative_to(L)),n
state=strict(E/'dispatch-state.json');assert state['completedPhases']==['A','B','C'] and set(state['toolRecords'])=={'A','B','C'} and state['status']=='COMPLETE_CAPTURE_PENDING_INDEPENDENT_AUDIT'
terminals={}
for phase,records in state['toolRecords'].items():
 final=records[-1];assert final['exit_code']==0 and not final.get('session_id');terminals[phase]=final['chunk_id']
 for earlier in records[:-1]:assert earlier.get('session_id')==records[0]['session_id'] and 'exit_code' not in earlier
 if phase!='A':assert len(records)==1
assert [terminals[p] for p in 'ABC']==['8f972f','ab393a','8645a5']
assert len(state['toolRecords']['A'])==3 and state['toolRecords']['A'][0]['session_id']==14031
for phase in 'ABC':
 baseline=(S/f'cells/{stage}.{phase}.txt').read_text();actual=(E/f'actual-{phase}.txt').read_text();assert actual==state['issuedSources'][phase]
 if phase=='A':assert actual==baseline
 else:
  key='ACTUAL_A_TOOL_ID' if phase=='B' else 'ACTUAL_B_TOOL_ID';token='UNRESOLVED_ACTUAL_TOOL_ID' if phase=='B' else 'UNRESOLVED_ACTUAL_B_TOOL_ID';tid=terminals['A' if phase=='B' else 'B'];old=f"{key} = '{token}'\n";new=f"{key} = '{tid}'\n";assert baseline.count(old)==1 and baseline.replace(old,new)==actual and actual.replace(new,old)==baseline
col=strict(R/'collector-provisional.json');cen=strict(R/'postterminal-census.json');wit=strict(R/'final-clock-witness.json');origin=strict(R/'origin.json');inner=strict(P/'run'/f'{stage}.exit.json')
assert type(inner['exit']) is int and inner['exit']==0 and inner['accepted'] is True and inner['sourceUnchanged'] is True and inner['forcedCleanup'] is False
assert inner['errors']==inner['remainingProcesses']==inner['runtimeDirectoriesStillPresent']==[]
assert inner==strict(R/f'{stage}.stdout')
assert sha(P/'run'/f'{stage}.log')==inner['logSha256']=='072b17518f5e1c0de392562d55b3a7c2232b1a7ff9aa0fff6f9353b6229d2602'
assert col['accepted'] is False and col['rawExit']==0 and col['failure'] is None and not col['forcedTermination'] and col['pipeCloseSucceeded']
streams=[]
for record in [col,cen]:
 for s in record['streams']:
  p=Path(s['path']);st=p.stat();f=s['fileIdentity'];raw=p.read_bytes()
  assert (st.st_dev,st.st_ino,st.st_mode,st.st_uid,st.st_size)==(f['device'],f['inode'],f['mode'],f['uid'],f['bytes'])
  assert stat.S_IMODE(st.st_mode)==0o600 and st.st_uid==501
  assert s['actualEOF'] and s['rawDescriptorClosed'] and s['hashKnown'] and s['writeOrCloseError'] is None
  assert s['receivedBytes']==s['savedBytes']==s['hashedPrefixBytes']==f['bytes']==len(raw)
  assert sha(p)==s['fullStreamSha256']==s['savedPrefixSha256'] and s['fullStreamState']=='COMPLETE_ACTUAL_CAPTURE'
  streams.append({'path':str(p),'bytes':len(raw),'sha256':sha(p),'actual0600EOFCloseCountsHash':True})
samples=[json.loads(v) for v in (P/'run'/f'{stage}.processes.jsonl').read_text().splitlines()];rows=[r for s in samples for r in s['group']]
pids={col['collectorPid'],col['creationPid'],*inner['sampledPids'],*(int(r[0]) for r in rows)};pgids={col['collectorPgid'],col['creationPgid'],*(int(r[2]) for r in rows)}
assert sorted(pids)==cen['taskPIDs']==col['knownTaskRegistry']['pids'] and sorted(pgids)==cen['taskPGIDs']==col['knownTaskRegistry']['pgids']
allrows=[];matching=[]
for line in (R/'postterminal-census.stdout').read_text().splitlines():
 f=line.split(None,4);assert len(f)==5;pid,ppid,pgid=map(int,f[:3]);allrows.append(f)
 if pid in pids or pgid in pgids:matching.append(line)
assert not matching and cen['matchingTaskRows']==[] and cen['probeRawExit']==0 and cen['probeBothEOF'] and cen['probeDirectReaped'] and cen['failure'] is None
assert cen['phaseAToolId']==terminals['A'] and wit['phaseBActualToolId']==terminals['B']
assert origin['originMonotonic']==col['originMonotonic']==cen['rootOriginMonotonic']==wit['rootOriginMonotonic']
elapsed=wit['postTerminalWitnessMonotonic']-origin['originMonotonic'];assert 0<elapsed<115 and elapsed==wit['completedInvocationAndObservationUpperBoundSeconds'] and wit['strictUnder115']
source=strict(P/'source-binding.json');W=Path(source['worktree']);assert len(source['trackedSourceFiles'])==679
for n,h in source['trackedSourceFiles'].items():assert sha(W/n)==h,n
impl=strict(P/'implementation-hashes.json');assert len(impl['files'])==105
for n,h in impl['files'].items():assert sha(P/n)==h,n
prior34='34c333aa5b688d7abe944363b558e0bd431dd512ec84a24f8e0440d9d2bdf7ea';active='437edd25c7267d9595a15115db65d50051372f0cb4915f6981cf46aa04be65c6';f16='f16d895808b965918bade0cbe1e7552d828423b96930485ec34f74f051e884c8';live=P/'execution-gate.json';assert sha(live)==active
prep=strict(P/'preparation-index.json');assert len(prep['files'])==110
for n,m in prep['files'].items():
 if n=='execution-gate.json':assert m['sha256']=='4a87b45d714a6bd28a94fd50a389defe4b5582dc5f42ec56fd827c7cf5b01e76'
 else:assert meta(P/n)==m,n
refs={}
for stem,count,original,project in [('write-profile-G09-validation5-root-proposed-authority-1',2206,'4a87b45d714a6bd28a94fd50a389defe4b5582dc5f42ec56fd827c7cf5b01e76',1),('write-profile-G09-validation5-root-fixed-protocol-preparation-1',2082,'4a87b45d714a6bd28a94fd50a389defe4b5582dc5f42ec56fd827c7cf5b01e76',1),('write-profile-G09-validation5-fixed-text-proof-correction-1',2207,'4a87b45d714a6bd28a94fd50a389defe4b5582dc5f42ec56fd827c7cf5b01e76',1),('write-profile-G09-validation-preparation-5',1857,None,0),('write-profile-G09-validation5-next5-closed-control-source-1',2314,'ad5b647e6b2f1aba93981d1a85fdb7ef0930062418012cf4acf90073de202919',1),('write-profile-G09-validation5-source-pairs-closed-control-source-1',3510,'18a440d47c2b4e7aef7e46afc73f627a27d13e51cae8b990cd2feade33c774a4',1),('write-profile-G09-validation5-package4-closed-control-source-1',6892,f16,1),('write-profile-G09-validation5-consumer2-closed-control-source-1',8472,prior34,1)]:
 table=strict(L/stem/'reference-pins.json');assert len(table)==count;changes=[]
 for n,h in table.items():
  current=sha(Path(n))
  if current!=h:assert n==str(live) and h==original and current==active;changes.append(n)
 assert len(changes)==project;refs[stem]={'total':count,'soleExactLiveGateHistoricalPhase':changes,'originalGate':original,'currentGate':active}
ledgerfacts={}
act=strict(L/'write-profile-G09-validation5-consumer2-root-activation-1/activation.json')
known={}
for name in ['write-profile-G09-validation5-source-pairs-root-activation-1/execution-gate.before.json','write-profile-G09-validation5-source-pairs-root-activation-1/execution-gate.active.json','write-profile-G09-validation5-package4-root-activation-1/execution-gate.active.json']:
 q=L/name;known[sha(q)]=meta(q)
for filename,facts in act['nativeLedgersAfter'].items():
 file=Path(filename);table=strict(file);assert len(table)==facts['total'];changes=[]
 for n,m in table.items():
  actual=meta(Path(n));expected={k:m[k] for k in ['bytes','mode','sha256']}
  if actual!=expected:
   assert n==str(live) and expected==known[m['sha256']] and actual==meta(live) and actual['sha256']==active and actual['mode']=='0o644',(filename,n);changes.append(n)
 assert changes==[str(live)]
 ledgerfacts[filename]={'total':len(table),'onlySameLiveGateExactKnownRawPhase':changes,'originalGate':facts['originalGate'],'currentGate':active}
previous22=L/'write-profile-G09-validation5-prepare-consumer-actual-native-audit-1/input-ledger.json';table=strict(previous22);assert len(table)==10049
for n,m in table.items():assert meta(Path(n))=={k:m[k] for k in ['bytes','mode','sha256']},n
ledgerfacts[str(previous22)]={'total':len(table),'strictCurrentNoDifferences':True,'currentGate':active}
toolmap=strict(P/'run/workspace-tools.json');assert len(toolmap['tools'])==11
for n,m in toolmap['tools'].items():assert str(Path(n).resolve(strict=True))==m['realpath'] and sha(Path(n))==m['sha256']
assert toolmap['sourceIdentity']==source['sourceIdentity'] and toolmap['sourceLockSha256']==sha(W/'pnpm-lock.yaml') and toolmap['installExitSha256']==sha(P/'run/workspace-install.exit.json') and toolmap['installLogSha256']==sha(P/'run/workspace-install.log')
build=strict(P/'run/build-artifacts.json');assert len(build['artifacts'])==476
for n,h in build['artifacts'].items():assert sha(W/n)==h,n
pfiles=[p for p in P.rglob('*') if p.is_file()];runfiles=[p for p in (P/'run').rglob('*') if p.is_file()];assert len(pfiles)==1086 and len(runfiles)==975
pk=P/'run/packages';assert pk.is_dir() and not pk.is_symlink() and sorted(p.name for p in pk.iterdir())==['bdp-protocol-0.0.0.tgz','bdp-server-0.0.0.tgz']
assert sorted(p.name for p in (P/'run').iterdir() if p.is_dir())==['consumer','packages']
assert not (P/'run/installed-provenance.json').exists() and not (P/'run/installed-consumer-inventory.json').exists()
# Independently check original consumer branch outputs and retain unchanged archive basis.
consumer=P/'run/consumer';assert consumer.is_dir() and not consumer.is_symlink()
assert sorted(p.name for p in consumer.iterdir())==['node_modules','package.json','pnpm-lock.yaml','pnpm-workspace.yaml']
deps={'@bdp/'+n:'file:'+str(pk/f'bdp-{n}-0.0.0.tgz') for n in ['protocol','server']}
manifestBytes=(json.dumps({'name':'tx-discovery-external-consumer','private':True,'type':'module','dependencies':deps},indent=2)+'\n').encode()
settingsBytes=('overrides:\n  "@bdp/protocol": '+json.dumps(deps['@bdp/protocol'])+'\n  "fast-uri": "3.1.5"\n').encode()
assert (consumer/'package.json').read_bytes()==manifestBytes and (consumer/'pnpm-workspace.yaml').read_bytes()==settingsBytes
consumerOutput={'path':str(consumer),'mode':oct(stat.S_IMODE(consumer.stat().st_mode)),'exactOriginalConsumerBranchManifestAndSettingsReproduced':True,'files':{n:meta(consumer/n) for n in ['package.json','pnpm-workspace.yaml']},'manifestSettingsUnchangedGeneratedNonfrozenLockAndNodeModulesPresentOriginalInstalledProofStillAbsent':True}
def typedEqual(a,b):
 if type(a) is not type(b):return False
 if isinstance(a,dict):return set(a)==set(b) and all(typedEqual(a[k],b[k]) for k in a)
 if isinstance(a,list):return len(a)==len(b) and all(typedEqual(x,y) for x,y in zip(a,b))
 return a==b
# Reproduce entire actual archive proof by independent static reads, no helper invocation.

import tarfile
proofPath=P/'run/tar-source-provenance.json';basis=strict(proofPath)
assert sha(proofPath)=='adb24bc4154d97f48c2acb5c4e21bc67a3f541dfd8696bfaff95bf997e552fea' and proofPath.stat().st_size==131894
assert set(basis)=={'head','tree','clean','dirtyPaths','ownedSourceFiles','sourceIdentity','sourceBindingSha256','binaryDiffSha256','packages','sourceLockSha256','sourceAttribution'}
assert basis['head']==source['head'] and basis['tree']==source['tree'] and basis['clean'] is False
assert basis['dirtyPaths']==source['dirtyPaths'] and basis['ownedSourceFiles']==source['ownedFiles']
assert basis['sourceIdentity']==source['sourceIdentity'] and basis['sourceBindingSha256']==sha(P/'source-binding.json') and basis['binaryDiffSha256']==source['binaryDiffSha256']
assert basis['sourceLockSha256']==sha(W/'pnpm-lock.yaml') and basis['sourceAttribution']=='Base HEAD/tree plus independently verified reviewed dirty source identity, fresh build and archives. Base commit does not contain the draft; not child-attested source.'
expectedArchives={'protocol':('4683cb6c6333c332a4bcb7efc3e28f32d1d8d06a7264913f65c31d3a826d4227',68665,63),'server':('f5cf974efd4e68087399a17619901d8c7ec17438eb154cc28570d1fa8eebf6f8',445045,258)}
reproduced={};archiveInspection={}
for name,(expectedSha,expectedBytes,expectedCount) in expectedArchives.items():
 arc=pk/f'bdp-{name}-0.0.0.tgz';assert sha(arc)==expectedSha and arc.stat().st_size==expectedBytes
 members={};root=W/'packages'/name
 with tarfile.open(arc,'r:gz') as tar:
  for member in tar.getmembers():
   assert member.isfile() and member.name.startswith('package/')
   rel=member.name[len('package/'):];assert '..' not in Path(rel).parts and not Path(rel).is_absolute() and rel not in members
   raw=tar.extractfile(member).read();origin=W/'LICENSE' if rel=='LICENSE' else root/rel
   if rel=='package.json':
    adjusted=strict(origin)
    if name=='server':
     assert adjusted['dependencies']['@bdp/protocol']=='workspace:*';adjusted=json.loads(json.dumps(adjusted));adjusted['dependencies']['@bdp/protocol']='0.0.0'
    assert typedEqual(json.loads(raw),adjusted)
   else:assert raw==origin.read_bytes()
   members[rel]={'bytes':len(raw),'sha256':hashlib.sha256(raw).hexdigest(),'source':str(origin),'sourceSha256':sha(origin)}
 required={'package.json','LICENSE'}
 for sub in (['dist','schemas'] if name=='protocol' else ['dist']):required.update(p.relative_to(root).as_posix() for p in (root/sub).rglob('*') if p.is_file())
 assert set(members)==required and len(members)==expectedCount
 reproduced[name]={'archive':str(arc),'archiveSha256':sha(arc),'files':members}
 archiveInspection[name]={'path':str(arc),'bytes':expectedBytes,'sha256':expectedSha,'members':members,'memberCount':expectedCount,'allSafeRegularUniquePathsFullSourceBuildBytesExact':True}
assert typedEqual(basis['packages'],reproduced) and set(basis['packages'])=={'protocol','server'}
expectedBasis={'head':source['head'],'tree':source['tree'],'clean':False,'dirtyPaths':source['dirtyPaths'],'ownedSourceFiles':source['ownedFiles'],'sourceIdentity':source['sourceIdentity'],'sourceBindingSha256':sha(P/'source-binding.json'),'binaryDiffSha256':source['binaryDiffSha256'],'packages':reproduced,'sourceLockSha256':sha(W/'pnpm-lock.yaml'),'sourceAttribution':'Base HEAD/tree plus independently verified reviewed dirty source identity, fresh build and archives. Base commit does not contain the draft; not child-attested source.'}
assert typedEqual(basis,expectedBasis)
def git(args):return subprocess.check_output(['git',*args],cwd=W)
assert git(['rev-parse','HEAD']).decode().strip()==source['head'] and git(['rev-parse','HEAD^{tree}']).decode().strip()==source['tree']
diff=hashlib.sha256(git(['diff','--binary'])).hexdigest();assert diff==source['binaryDiffSha256'];status=git(['status','--porcelain=v1','--untracked-files=all']).decode().splitlines();assert len(status)==10 and all(v.startswith(' M ') for v in status)
# Static installation DATA inventory and fixed graph checks, not original stage24 helper invocation.
fs=b['actualInstalledConsumerFilesystemDataInventory'];assert len(fs['files'])==870 and len(fs['symlinks'])==17
actualFiles={str(p.relative_to(consumer)):meta(p) for p in consumer.rglob('*') if p.is_file() and not p.is_symlink()}
assert actualFiles==fs['files']
actualLinks={str(p.relative_to(consumer)):{'target':str(p.readlink()),'resolved':str(p.resolve(strict=True)),'targetWithinActualConsumer':p.resolve(strict=True).is_relative_to(consumer.resolve())} for p in consumer.rglob('*') if p.is_symlink()}
assert actualLinks==fs['symlinks'] and all(x['targetWithinActualConsumer'] for x in actualLinks.values())
lock=consumer/'pnpm-lock.yaml';assert sha(lock)=='1ee287fcdfe295bd82001f9a09fe3dec06037f2128ee2c332e11cadb3bfd2ee5' and lock.stat().st_size==2958
assert (consumer/'node_modules/.pnpm/lock.yaml').read_bytes()==lock.read_bytes()
import re,base64
lockText=lock.read_text()
pkgText,snapText=lockText.split('\nsnapshots:\n\n');preText,pkgText=pkgText.split('\npackages:\n\n')
packageKeys=[x.strip("'") for x in re.findall(r'^  (.+):
out={'scope':'ONE_ACTUAL_CONSUMER_INSTALL_FULL_ROOT_MECHANICS_ACTUAL_GENERATED_LOCK_FILESYSTEM_DATA_AND_UNCHANGED_ARCHIVE_PROOF_ONLY','stage':stage,'actualTerminalTools':terminals,'AInitial6826d6Session14031Poll4c1311ThenTerminal8f972fAllThreeResultsRetained':True,'issuedAExactBCOnlyWholeLFActualIDsFullInverse':True,'evidence9AndRaw13FullBytesModesMirrorsExact':True,'fullStreams':streams,'allCensusRowsVerified':len(allrows),'allInnerSamplesVerified':len(samples),'recordedTaskPIDs':sorted(pids),'recordedTaskPGIDs':sorted(pgids),'allRecordedTaskRowsAbsent':True,'probeActualDirectReaped':True,'commonClockSeconds':elapsed,'strictUnder115':True,'innerElapsedSeconds':inner['elapsedSeconds'],'innerExit0AcceptedSourceUnchanged':True,'sourceIdentity':source['sourceIdentity'],'source679Impl105Other110Prepared111Files1086Run975IncludingInstalledPayloadBuild476Tools11GitSourceLockExact':True,'allReferenceDomains':refs,'allPriorNativeLedgersExactKnownSameLiveGateOnly':ledgerfacts,'actualOutputPackagesDirectory':{'path':str(pk),'mode':oct(stat.S_IMODE(pk.stat().st_mode)),'bothProtocolAndServerArchivesAtStaticObservation':True,'archiveInspection':archiveInspection,'actualCombinedTarProvenance':{'path':str(proofPath),'sha256':sha(proofPath),'bytes':131894,'all11KeysAnd321MemberRecordsReproducedExactly':True}},'actualNonemptyConsumerInstallLog':meta(P/'run'/f'{stage}.log'),'scratch':{'path':str(scratch),'existsAtStaticRead':scratch.is_dir(),'entriesAtStaticRead':[p.name for p in scratch.iterdir()] if scratch.is_dir() else None,'cleanupPerformed':False},'finiteRecordedSetOnlyNoGlobalOrObserverSelfAbsenceClaim':True,'wireReceiptsAndFullEnvironment':'UNRETAINED/source-attributed normalized records','preservedConsumerManifestSettings':consumerOutput,'actualInstalledFilesystemAndLockData':installedData,'actualArchivesAcceptedGeneratedNonfrozenLockCapturedOriginalInstalledProofUnknown':True,'nextStageReleased':False,'qualityPass':False,'nativeAudit':'PENDING','rootFinalAcceptance':False}
OUT=L/'root-G09-validation5-consumer-install-independent-actual-inspection-1.json';assert not OUT.exists();OUT.write_text(json.dumps(out,indent=2,sort_keys=True)+'\n');print(json.dumps({'inspection':str(OUT),'sha256':sha(OUT),'censusRows':len(allrows),'innerSamples':len(samples),'pids':len(pids),'pgids':len(pgids),'clock':elapsed,'runtimeAction':False}))
,pkgText,re.M)]
snapshotKeys=[x.strip("'") for x in re.findall(r'^  (.+):
out={'scope':'ONE_ACTUAL_CONSUMER_INSTALL_FULL_ROOT_MECHANICS_ACTUAL_GENERATED_LOCK_FILESYSTEM_DATA_AND_UNCHANGED_ARCHIVE_PROOF_ONLY','stage':stage,'actualTerminalTools':terminals,'AInitial6826d6Session14031Poll4c1311ThenTerminal8f972fAllThreeResultsRetained':True,'issuedAExactBCOnlyWholeLFActualIDsFullInverse':True,'evidence9AndRaw13FullBytesModesMirrorsExact':True,'fullStreams':streams,'allCensusRowsVerified':len(allrows),'allInnerSamplesVerified':len(samples),'recordedTaskPIDs':sorted(pids),'recordedTaskPGIDs':sorted(pgids),'allRecordedTaskRowsAbsent':True,'probeActualDirectReaped':True,'commonClockSeconds':elapsed,'strictUnder115':True,'innerElapsedSeconds':inner['elapsedSeconds'],'innerExit0AcceptedSourceUnchanged':True,'sourceIdentity':source['sourceIdentity'],'source679Impl105Other110Prepared111Files1086Run975IncludingInstalledPayloadBuild476Tools11GitSourceLockExact':True,'allReferenceDomains':refs,'allPriorNativeLedgersExactKnownSameLiveGateOnly':ledgerfacts,'actualOutputPackagesDirectory':{'path':str(pk),'mode':oct(stat.S_IMODE(pk.stat().st_mode)),'bothProtocolAndServerArchivesAtStaticObservation':True,'archiveInspection':archiveInspection,'actualCombinedTarProvenance':{'path':str(proofPath),'sha256':sha(proofPath),'bytes':131894,'all11KeysAnd321MemberRecordsReproducedExactly':True}},'actualNonemptyConsumerInstallLog':meta(P/'run'/f'{stage}.log'),'scratch':{'path':str(scratch),'existsAtStaticRead':scratch.is_dir(),'entriesAtStaticRead':[p.name for p in scratch.iterdir()] if scratch.is_dir() else None,'cleanupPerformed':False},'finiteRecordedSetOnlyNoGlobalOrObserverSelfAbsenceClaim':True,'wireReceiptsAndFullEnvironment':'UNRETAINED/source-attributed normalized records','preservedConsumerManifestSettings':consumerOutput,'actualInstalledFilesystemAndLockData':installedData,'actualArchivesAcceptedGeneratedNonfrozenLockCapturedOriginalInstalledProofUnknown':True,'nextStageReleased':False,'qualityPass':False,'nativeAudit':'PENDING','rootFinalAcceptance':False}
OUT=L/'root-G09-validation5-consumer-install-independent-actual-inspection-1.json';assert not OUT.exists();OUT.write_text(json.dumps(out,indent=2,sort_keys=True)+'\n');print(json.dumps({'inspection':str(OUT),'sha256':sha(OUT),'censusRows':len(allrows),'innerSamples':len(samples),'pids':len(pids),'pgids':len(pgids),'clock':elapsed,'runtimeAction':False}))
,snapText,re.M)]
expected=strict(P/'commands.json')['expectedProductionVersions'];assert len(expected)==8
expectedKeys={n+'@'+v for n,v in expected.items() if not n.startswith('@bdp/')}|{'@bdp/'+n+'@file:../packages/bdp-'+n+'-0.0.0.tgz' for n in ['protocol','server']}
assert len(packageKeys)==len(set(packageKeys))==8 and set(packageKeys)==expectedKeys
assert len(snapshotKeys)==len(set(snapshotKeys))==8 and set(snapshotKeys)==(expectedKeys-{'ajv-formats@3.0.1'})|{'ajv-formats@3.0.1(ajv@8.17.1)'}
assert preText.startswith("lockfileVersion: '9.0'\n") and preText.count('        specifier: file:')==2
for name in ['protocol','server']:
 arc=pk/f'bdp-{name}-0.0.0.tgz';integrity='sha512-'+base64.b64encode(hashlib.sha512(arc.read_bytes()).digest()).decode()
 assert 'integrity: '+integrity in pkgText and '        specifier: file:'+str(arc) in preText and '        version: file:../packages/'+arc.name in preText
manifests={p.resolve(strict=True) for p in list((consumer/'node_modules/.pnpm').glob('*/node_modules/*/package.json'))+list((consumer/'node_modules/.pnpm').glob('*/node_modules/@*/*/package.json'))}
assert len(manifests)==8
graph={};versions={}
for manifest in sorted(manifests):
 assert manifest.is_relative_to(consumer.resolve());doc=strict(manifest);name=doc['name'];assert name not in versions and doc['version']==expected[name];versions[name]=doc['version'];depsResolved={}
 for dep in doc.get('dependencies',{}):
  probes=[a/'node_modules'/dep/'package.json' for a in [manifest.parent,*manifest.parent.parents] if a.is_relative_to(consumer.resolve())]
  resolved=next(p.resolve(strict=True) for p in probes if p.exists());assert resolved in manifests
  dm=strict(resolved);assert dm['name']==dep and dm['version']==expected[dep];depsResolved[dep]={'manifest':str(resolved),'version':dm['version']}
 graph[name]={'manifest':str(manifest),'sha256':sha(manifest),'dependencies':depsResolved}
assert versions==expected
for name in ['protocol','server']:
 installed=(consumer/'node_modules/@bdp'/name).resolve(strict=True)
 assert installed.is_relative_to(consumer.resolve())
 members=reproduced[name]['files'];assert {p.relative_to(installed).as_posix() for p in installed.rglob('*') if p.is_file()}==set(members)
 for n,m in members.items():assert sha(installed/n)==m['sha256'] and (installed/n).stat().st_size==m['bytes']
assert (consumer/'node_modules/@bdp/server/dist/transaction-installed-process.js').resolve(strict=True).is_relative_to(consumer.resolve())
installedData={'scope':'STATIC filesystem, actual generated NONFROZEN lock and fixed dependency graph DATA ONLY; original stage24 installed proof unattempted','regularFiles':870,'symlinks':17,'allFilesMetadataHashesAndSymlinkTargetsReproduced':True,'actualLock':{'path':str(lock),'sha256':sha(lock),'bytes':2958},'lockPackageAndSnapshotKeys':{'packages':packageKeys,'snapshots':snapshotKeys},'actualArchiveSha512IntegrityAndFileImportersExact':True,'actualInternalPnpmLockSame':True,'expectedEightProductionVersions':versions,'installedGraph':graph,'installedProtocolServerAll321ArchiveMemberBytesExact':True,'entryPresentFilesystemOnlyNotInvoked':True,'actualLogReports':'resolved8/reused7/downloaded1/added8','networkObservation':'UNRETAINED; actual offline argv and literal log only, no zero-network claim','originalInstalledProvenanceAndInventoryAbsentFilesystemOnly':True}
scratch=Path(inner['temporaryDirectory'])
out={'scope':'ONE_ACTUAL_CONSUMER_INSTALL_FULL_ROOT_MECHANICS_ACTUAL_GENERATED_LOCK_FILESYSTEM_DATA_AND_UNCHANGED_ARCHIVE_PROOF_ONLY','stage':stage,'actualTerminalTools':terminals,'AInitial6826d6Session14031Poll4c1311ThenTerminal8f972fAllThreeResultsRetained':True,'issuedAExactBCOnlyWholeLFActualIDsFullInverse':True,'evidence9AndRaw13FullBytesModesMirrorsExact':True,'fullStreams':streams,'allCensusRowsVerified':len(allrows),'allInnerSamplesVerified':len(samples),'recordedTaskPIDs':sorted(pids),'recordedTaskPGIDs':sorted(pgids),'allRecordedTaskRowsAbsent':True,'probeActualDirectReaped':True,'commonClockSeconds':elapsed,'strictUnder115':True,'innerElapsedSeconds':inner['elapsedSeconds'],'innerExit0AcceptedSourceUnchanged':True,'sourceIdentity':source['sourceIdentity'],'source679Impl105Other110Prepared111Files1086Run975IncludingInstalledPayloadBuild476Tools11GitSourceLockExact':True,'allReferenceDomains':refs,'allPriorNativeLedgersExactKnownSameLiveGateOnly':ledgerfacts,'actualOutputPackagesDirectory':{'path':str(pk),'mode':oct(stat.S_IMODE(pk.stat().st_mode)),'bothProtocolAndServerArchivesAtStaticObservation':True,'archiveInspection':archiveInspection,'actualCombinedTarProvenance':{'path':str(proofPath),'sha256':sha(proofPath),'bytes':131894,'all11KeysAnd321MemberRecordsReproducedExactly':True}},'actualNonemptyConsumerInstallLog':meta(P/'run'/f'{stage}.log'),'scratch':{'path':str(scratch),'existsAtStaticRead':scratch.is_dir(),'entriesAtStaticRead':[p.name for p in scratch.iterdir()] if scratch.is_dir() else None,'cleanupPerformed':False},'finiteRecordedSetOnlyNoGlobalOrObserverSelfAbsenceClaim':True,'wireReceiptsAndFullEnvironment':'UNRETAINED/source-attributed normalized records','preservedConsumerManifestSettings':consumerOutput,'actualInstalledFilesystemAndLockData':installedData,'actualArchivesAcceptedGeneratedNonfrozenLockCapturedOriginalInstalledProofUnknown':True,'nextStageReleased':False,'qualityPass':False,'nativeAudit':'PENDING','rootFinalAcceptance':False}
OUT=L/'root-G09-validation5-consumer-install-independent-actual-inspection-1.json';assert not OUT.exists();OUT.write_text(json.dumps(out,indent=2,sort_keys=True)+'\n');print(json.dumps({'inspection':str(OUT),'sha256':sha(OUT),'censusRows':len(allrows),'innerSamples':len(samples),'pids':len(pids),'pgids':len(pgids),'clock':elapsed,'runtimeAction':False}))
