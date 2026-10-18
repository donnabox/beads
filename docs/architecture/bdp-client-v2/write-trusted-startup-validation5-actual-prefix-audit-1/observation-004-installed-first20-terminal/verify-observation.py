from pathlib import Path
import os,json,hashlib,tarfile,re,subprocess,datetime
L=Path('/Users/dbox/repos/wt/vickie/bdp-integration-logs');D=Path('/Users/dbox/repos/wt/vickie/bdp-client-design/checkout/docs/architecture/bdp-client-v2');P=L/'write-trusted-startup-validation-preparation-5';R=P/'run';O=Path(__file__).parent
sha=lambda b:hashlib.sha256(b).hexdigest()
j=lambda f:json.loads(f.read_text())
def save(n,b):
 p=O/n;p.parent.mkdir(parents=True,exist_ok=True);p.write_bytes(b)
def record(n,x):save(n,(json.dumps(x,indent=2)+'\n').encode())
observedUtc=datetime.datetime.now(datetime.UTC).isoformat();config=j(P/'commands.json');stages={};pids=set();groups=set();scratch=[];prefix=config['finalOrder'][:20];assert prefix[-1]=='audit-bound-consumer-loader'
for n in prefix:
 st=j(R/(n+'.started.json'));ex=j(R/(n+'.exit.json'));assert ex['accepted'] is True and ex['exit']==0 and ex['sourceUnchanged'] is True and not ex['forcedCleanup'] and not ex['errors'] and not ex['remainingProcesses'] and not ex['runtimeDirectoriesStillPresent']
 assert sha((R/(n+'.log')).read_bytes())==ex['logSha256'];assert st['argv']==config['commands'][n]['argv'] and st['cwd']==config['commands'][n]['cwd'] and st['automaticRetries']==0;pids.update(ex['sampledPids'])
 for ln in (R/(n+'.processes.jsonl')).read_text().splitlines():
  for row in json.loads(ln)['group']:pids.add(int(row[0]));groups.add(int(row[2]))
 if n in prefix[15:]:
  for suffix in ['started.json','exit.json','log','processes.jsonl']:save('actual/'+n+'.'+suffix,(R/(n+'.'+suffix)).read_bytes())
 tmp=Path(ex['temporaryDirectory']);bdp=[str(x) for x in tmp.rglob('*') if x.is_dir() and x.name.startswith('bdp-')] if tmp.exists() else [];assert not bdp;scratch.append({'stage':n,'path':str(tmp),'exists':tmp.exists(),'bdpDirectories':bdp})
 stages[n]={'exit':0,'accepted':True,'elapsedSeconds':ex['elapsedSeconds'],'exitSha256':sha((R/(n+'.exit.json')).read_bytes()),'logSha256':ex['logSha256'],'startedGateSha256':st['rootGateSha256']}
proof=j(R/'installed-provenance.json');C=Path(proof['consumer']);files={};links={}
for directory,dirs,names in os.walk(C,followlinks=False):
 for name in dirs+names:
  p=Path(directory)/name;rel=p.relative_to(C).as_posix()
  if p.is_symlink():q=p.resolve(strict=True);assert q.is_relative_to(C.resolve());links[rel]={'link':os.readlink(p),'target':q.relative_to(C.resolve()).as_posix()}
  elif p.is_file():files[rel]=sha(p.read_bytes())
assert {'files':files,'links':links}==j(R/'installed-consumer-inventory.json') and len(files)==878 and len(links)==17
basis=j(R/'tar-source-provenance.json');membercount=0
for name,a in basis['packages'].items():
 archive=Path(a['archive']);assert sha(archive.read_bytes())==a['archiveSha256'];installed=(C/'node_modules/@bdp'/name).resolve(strict=True);assert installed.is_relative_to(C.resolve())
 with tarfile.open(archive,'r:gz') as tf:
  for rel,m in a['files'].items():
   z=tf.extractfile('package/'+rel).read();assert (installed/rel).read_bytes()==z and sha(z)==m['sha256'];assert sha(Path(m['source']).read_bytes())==m['sourceSha256'];membercount+=1
 assert {p.relative_to(installed).as_posix() for p in installed.rglob('*') if p.is_file()}==set(a['files'])
 assert str(installed)==proof['packages'][name]['realpath'] and proof['packages'][name]['checkedMembers']==len(a['files'])
assert membercount==329
expected=proof['expectedProductionVersionsFromFrozenLock'];graph=proof['installedProductionGraph'];assert len(graph)==8
actual={}
def resolve(manifest,dep):
 for ancestor in [manifest.parent,*manifest.parent.parents]:
  if not ancestor.is_relative_to(C.resolve()):break
  q=ancestor/'node_modules'/dep/'package.json'
  if q.exists():q=q.resolve(strict=True);assert q.is_relative_to(C.resolve());return q
 raise AssertionError(dep)
for row in graph:
 manifest=Path(row['manifest']);assert sha(manifest.read_bytes())==row['manifestSha256'];obj=j(manifest);assert obj['name']==row['name'] and obj['version']==row['version']==expected[row['name']];actual[row['name']]=row['version'];deps={}
 for dep in sorted(obj.get('dependencies',{})):
  target=resolve(manifest,dep);m=j(target);assert m['name']==dep and m['version']==expected[dep];deps[dep]={'manifestRealpath':str(target),'version':m['version']}
 assert deps==row['resolvedDependencies']
assert actual==expected
h=hashlib.sha256();visits=[];count=[0]
def payload(root,identity,ancestors):
 assert root not in ancestors and len(ancestors)<32;manifest=root/'package.json';assert not manifest.is_symlink();obj=j(manifest);visits.append(identity)
 def emit(rel):
  z=(root/rel).read_bytes();h.update(f'{identity}/{rel}:{len(z)}:'.encode());h.update(z);count[0]+=1
 def walk(rel):
  p=root/rel;assert not p.is_symlink()
  for q in sorted(p.iterdir(),key=lambda q:q.name):
   assert q.name.isascii()
   if q.name in ['node_modules','.git']:continue
   assert not q.is_symlink();name=q.relative_to(root).as_posix()
   if q.is_dir():walk(name)
   elif q.is_file() and re.search(r'\.(?:[cm]?js|json|node|wasm)$',q.name):emit(name)
 if obj['name'] in ['@bdp/server','@bdp/protocol']:
  emit('package.json');walk('dist')
  if obj['name']=='@bdp/protocol':walk('schemas')
 else:walk('')
 for dep in sorted(obj.get('dependencies',{})):payload(resolve(manifest,dep).parent,identity+'>'+dep,[*ancestors,root])
payload((C/'node_modules/@bdp/server').resolve(),'@bdp/server',[]);assert visits==proof['payloadRecursiveVisits'] and len(visits)==14 and count[0]==proof['payloadFileVisits']==477 and h.hexdigest()==proof['payloadIdentity']
assert sha((C/'pnpm-lock.yaml').read_bytes())==proof['consumerLockSha256'] and sha(Path(proof['entryRealpath']).read_bytes())==proof['entrySha256']
reports=[]
for line in (R/'bound-consumer-loader.log').read_text().splitlines():
 try:x=json.loads(line)
 except json.JSONDecodeError:continue
 if isinstance(x,dict) and x.get('event')=='startup.bound-consumer-loader':reports.append(x)
assert len(reports)==1;report=reports[0];accepted=j(R/'bound-consumer-loader.accepted-observation.json');assert accepted['observation']==report and accepted['qualification'] is False
assert set(report)==set(j(P/'report-schemas.json')['installed']['exactReportFields'])
for k,v in {'outcome':'passed','profiles':['read-update','transactional'],'loaderCalls':6,'accepted':2,'contractRefusals':4,'privateSpecifierCount':71,'publicExports':['ReferenceDeploymentError','loadReferenceDeployment'],'configuredPortAcquisitionExcluded':True,'storeDirectoriesAbsent':True,'workTimedOut':False,'cleanupUnknown':False,'temporaryDirectoryRemoved':True}.items():assert type(report[k]) is type(v) and report[k]==v,k
assert report['entryRealpath']==proof['entryRealpath'] and report['entrySha256']==proof['entrySha256'] and report['provenanceSha256']==sha((R/'installed-provenance.json').read_bytes())
assert not Path(report['temporaryDirectory']).exists() and Path(report['temporaryDirectory']).is_relative_to(Path(j(R/'bound-consumer-loader.started.json')['temporaryDirectory']))
b=j(P/'source-binding.json');assert len(b['trackedSourceFiles'])==682
for f,expectedhash in b['trackedSourceFiles'].items():assert sha((Path(b['worktree'])/f).read_bytes())==expectedhash
for f,m in j(P/'preparation-index.json')['files'].items():
 if f=='execution-gate.json':continue
 z=(P/f).read_bytes();assert sha(z)==m['sha256'] and z==(D/P.name/f).read_bytes()
ps=subprocess.run(['/bin/ps','-axo','pid=,ppid=,pgid=,stat=,lstart=,comm='],capture_output=True,check=True,timeout=10);save('process-inventory.txt',ps.stdout)
assert not [x for x in ps.stdout.decode().splitlines() if len(x.split(None,3))>=3 and (int(x.split(None,3)[0]) in pids or int(x.split(None,3)[2]) in groups)]
for n in ['installed-provenance.json','installed-consumer-inventory.json','bound-consumer-loader.accepted-observation.json']:save('actual/'+n,(R/n).read_bytes())
for n in ['pnpm-lock.yaml','package.json','pnpm-workspace.yaml']:save('consumer-metadata/'+n,(C/n).read_bytes())
gate=(P/'execution-gate.json').read_bytes();save('active-gate.observed.json',gate);g=json.loads(gate)
for n in ['rootConsumerLockAcceptance','rootConsumerProofAcceptance']:
 v=g[n];z=Path(v['path']).read_bytes();assert sha(z)==v['sha256'];save('root/'+n+'.json',z)
observation={'observedUtc':observedUtc,'acceptedPrefixStages':20,'stages':stages,'consumerInventoryFiles':878,'consumerInventoryLinks':17,'installedArchiveMembersVerified':329,'graphVersions':actual,'payloadDigest':h.hexdigest(),'payloadVisits':14,'payloadFileVisits':477,'installedProvenanceSha256':sha((R/'installed-provenance.json').read_bytes()),'consumerInventorySha256':sha((R/'installed-consumer-inventory.json').read_bytes()),'consumerLockSha256':proof['consumerLockSha256'],'entrySha256':proof['entrySha256'],'loaderObservation':report,'sourceFilesVerified':682,'immutablePreparationFilesVerified':465,'sourceIdentity':proof['sourceIdentity'],'sourcePatchSha256':basis['binaryDiffSha256'],'sampledPids':sorted(pids),'sampledGroups':sorted(groups),'sampledPidsGroupsAbsent':True,'scratchDirectories':scratch,'observedActiveGateSha256':sha(gate),'observedActiveApprovedStages':g['approvedStages'],'laterStartedAtObservation':sorted(f.name for f in R.glob('*.started.json') if f.name.removesuffix('.started.json') not in prefix),'qualification':False,'noServiceSQLiteListenerAdmissionClaim':True,'stockSmokeConsumerSeparate':True,'findings':[],'auditorRuntimeExecuted':False}
record('observation.json',observation);print(json.dumps({'epoch':observedUtc,'pids':len(pids),'groups':len(groups),'consumer':len(files),'links':len(links),'payload':h.hexdigest(),'laterStarted':observation['laterStartedAtObservation']},indent=2))
