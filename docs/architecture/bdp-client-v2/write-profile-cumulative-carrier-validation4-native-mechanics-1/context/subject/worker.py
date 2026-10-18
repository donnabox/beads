"""Owned blocking attempt worker, supervised under the parent's single monotonic origin."""
import sys,os,time,json,socket,subprocess,tempfile,secrets,hashlib,signal
signal.pthread_sigmask(signal.SIG_SETMASK,[]) # parent registered ownership before unmask
sys.dont_write_bytecode=True
from pathlib import Path
from scanner import SummaryScanner,bounded_json,file_hash,secure_read,RESULTS,RAW,WRAPPER,SIDECAR,need
if len(sys.argv)>1 and sys.argv[1]=='publish':
 # Small private publisher is still an owned deadline-supervised direct child.
 channel=socket.socket(fileno=int(sys.argv[2]));stage=sys.argv[3];data=b''
 while b'\n'not in data:
  part=channel.recv(65536);need(part and len(data)+len(part)<=1048576,'publication-bound');data+=part
 need(data.endswith(b'\n')and data.count(b'\n')==1,'publication-record');record=json.loads(data);need(record['stage']==stage and record['watchdogPid']==os.getppid(),'publication-identity');root=Path(__file__).parent/'run';need(stage and all(c.isalnum()or c=='-'for c in stage)and root.is_dir(),'publication-path')
 with (root/(stage+'.watchdog.json')).open('xb')as f:f.write(data)
 channel.close();raise SystemExit(0)
channel=socket.socket(fileno=int(sys.argv[1]));name=sys.argv[2];origin=float(sys.argv[3]);inbox=b''
def note(event,**data):
 raw=(json.dumps({'event':event,**data},separators=(',',':'))+'\n').encode();need(len(raw)<=4096,'private-control-bound');channel.sendall(raw)
def await_control(expected):
 global inbox
 while True:
  if b'\n'in inbox:
   line,inbox=inbox.split(b'\n',1);value=json.loads(line);need(value.get('event')==expected,'parent-control-order');return value
  block=channel.recv(4096);need(block and len(inbox)+len(block)<=8192,'parent-control-bound');inbox+=block
failure_phase='gate'
try:
 os.environ['CARRIER_VALIDATION_PHASE']='final'
 from gate import *
 g,spec=load_gate(name);failure_phase='prepare'
 if not E.exists():E.mkdir(mode=0o700,exist_ok=False)
 require(E.is_dir()and not E.is_symlink()and E.stat().st_mode&0o777==0o700 and E.stat().st_uid==os.getuid(),'Owned execution root')
 for suffix in ['.started.json','.log','.exit.json','.results.json','.launch.json','.worker-verified.json','.acquisition.json']:
  require(not(E/(name+suffix)).exists(),'Preserve previous stage; no retry')
 attempt=secrets.token_hex(16);owned=E/('attempt-'+name);owned.mkdir(mode=0o700,exist_ok=False);tmp=owned/'scratch';tmp.mkdir(mode=0o700,exist_ok=False)
 env=os.environ.copy()
 for k in list(env):
  if k.startswith(('BDP_','VITEST_','CARRIER_')) or k in ['NODE_OPTIONS','NODE_PATH','NODE_DEBUG','PYTHONOPTIMIZE','BIOME_BINARY','npm_execpath','npm_node_execpath','FORCE_COLOR']:env.pop(k)
 env.update({'PATH':'/Users/dbox/.nvm/versions/node/v24.16.0/bin:'+env.get('PATH',''),'CI':'true','NO_COLOR':'1','PYTHONDONTWRITEBYTECODE':'1','CARRIER_VALIDATION_PHASE':'final','TMPDIR':str(tmp),'npm_config_cache':str(tmp/'npm-cache')});env.update(spec.get('environment',{}))
 destinations=[]
 if name=='single-installed-positive17':
  destinations=[owned/'original-legacy-wrapper.json',owned/'carrier.transactional.profile.attempt.json',owned/'carrier.transactional.profile.json']
  require(all(not f.exists()and not f.is_symlink()for f in destinations),'Source output destination must be absent before launch')
  env.update({'BDP_PRIVATE_PROFILE_EVIDENCE_PATH':str(owned/'carrier'),'BDP_TRANSACTIONAL_OBSERVATION_OUTPUT':str(destinations[0])})
 if name=='private-verifier-emit':
  from private_emit import prepare_emit
  prepare_emit(g,owned)
 if name=='IMMEDIATE-positive-carrier-audit':
  from audit import prepare_carrier_audit
  validated_acquisition=prepare_carrier_audit(g,owned);env['CARRIER_AUDIT_ACQUISITION_SHA256']=validated_acquisition['sha256']
 started={'argv':spec['argv'],'cwd':spec['cwd'],'head':g['finalHead'],'tree':g['finalTree'],'sourceIdentity':g['_sourceIdentity'],'sourceBindingSha256':g['sourceBinding']['sha256'],'rootGateSha256':sha((P/'execution-gate.json').read_bytes()),'temporaryDirectory':str(tmp),'ownedAttemptRoot':str(owned),'attemptId':attempt,'singleMonotonicOrigin':origin,'outerSeconds':120,'workCutoffSeconds':105,'automaticRetries':0,'childUmask':0o077,'outputDestinationsAbsent':True if destinations else None,'outputDestinations':[str(f)for f in destinations]};write(name+'.started.json',started)
 env['CARRIER_VALIDATION_STAGE']=name;env['CARRIER_VALIDATION_ATTEMPT']=sha((E/(name+'.started.json')).read_bytes());note('attempt',attemptId=attempt,sourceBindingSha256=g['sourceBinding']['sha256'])
 require(time.monotonic()<origin+105,'Admission exhausted; no child spawn')
 logpath=E/(name+'.log');logfd=os.open(logpath,os.O_WRONLY|os.O_CREAT|os.O_EXCL|os.O_CLOEXEC,0o600)
 with os.fdopen(logfd,'wb')as log:
  failure_phase='spawn'
  child=subprocess.Popen(spec['argv'],cwd=spec['cwd'],env=env,stdin=subprocess.DEVNULL,stdout=subprocess.PIPE,stderr=subprocess.STDOUT,umask=0o077)
  created=time.time();note('created',pid=child.pid,pgid=os.getpgrp(),ppid=os.getpid(),executable=Path(spec['argv'][0]).name,createdEpoch=created)
  launch={'stage':name,'attemptId':attempt,'pid':child.pid,'ppid':os.getpid(),'pgid':os.getpgrp(),'createdEpoch':created,'argv':spec['argv'],'cwd':spec['cwd'],'sourceBindingSha256':g['sourceBinding']['sha256'],'startedSha256':sha((E/(name+'.started.json')).read_bytes()),'actualOwnedLaunchSetting':{'mechanism':'subprocess.Popen umask argument installed before exec','umask':0o077,'inheritedByVitestWorkersAndInstalledDescendants':True},'outputDestinationsAbsentBeforePopen':True if destinations else None};write(name+'.launch.json',launch)
  await_control('creation-ack');failure_phase='collect'
  scanner=SummaryScanner()if name=='single-installed-positive17'else None;logbytes=0;stdio_eof=False
  os.set_blocking(child.stdout.fileno(),False)
  while True:
   try:block=os.read(child.stdout.fileno(),65536)
   except BlockingIOError:block=None
   if block==b'':stdio_eof=True
   elif block:
    log.write(block);log.flush();logbytes+=len(block);require(logbytes<=min(RAW,spec.get('logBytesMax',RAW)),'Combined raw log capacity exceeded; retain file, no truncation/pass')
    if scanner:scanner.feed(block)
   if child.poll()is not None and stdio_eof:break
   if not block:time.sleep(.01)
  child.stdout.close()
 note('leader-exit',pid=child.pid,code=child.returncode);note('need-quiet');await_control('quiet')
 if scanner:scanner.finish()
 require(child.returncode in spec.get('acceptedRawExitCodes',[0]),'Child exit refused')
 failure_phase='binding'
 unchanged=source_check(g['_sourceBinding'])==g['_sourceIdentity'];require(unchanged,'Final source differs')
 if name not in ['node-version','pnpm-version','workspace-install','bind-workspace-tools','typecheck','format-check','lint','required-check-boundaries','fresh-build','bind-fresh-build']:verify_build(g)
 proof=verify_consumer(g)if spec.get('requiresConsumerProof')is True else None
 if spec.get('requiresPrivateEmitProof')is True:verify_private_emit(g)
 runtime=[str(x)for x in tmp.rglob('*')if x.is_dir()and x.name.startswith('bdp-')];require(not runtime,'Runtime directories remain')
 semantic=None;failure_phase='data-audit'
 if spec.get('selectedFullNames'):
  from audit import selected_results
  semantic=selected_results(bounded_json(E/(name+'.results.json'),RESULTS),spec)
 if scanner:
  facts=[];captured_bytes=[]
  for f,cap in zip(destinations,[WRAPPER,SIDECAR,SIDECAR]):
   data,fact=secure_read(f,owned,cap);facts.append(fact);captured_bytes.append(data)
  require(captured_bytes[1]==captured_bytes[2],'Actual final attempt/summary byte arrays differ')
  require({f.name for f in owned.iterdir()}=={'scratch',*(f.name for f in destinations)},'Unexpected source-output root residue; preserve evidence')
  summary=scanner.summaries[0];require(scanner.hash.hexdigest()==file_hash(logpath)['sha256'],'Full scanned positive log differs')
  write(name+'.acquisition.json',{'attemptId':attempt,'sourceBindingSha256':g['sourceBinding']['sha256'],'ownedAttemptRoot':str(owned),'originalSourceWrittenFiles':facts,'uniqueActualSummary':{k:v for k,v in summary.items()if k!='value'},'stdioEOF':stdio_eof,'productExit':child.returncode,'groupQuietObservedByParent':True,'selectedResults':semantic,'launchReceiptSha256':sha((E/(name+'.launch.json')).read_bytes()),'credit':'PENDING actual whole-stage terminal/root receipt and immediate independent full-byte audit; this is acquisition only','claimEligible':False})
 if name=='private-verifier-emit':
  from private_emit import bind_emit
  bind_emit(g,owned)
 if name=='prove-private-graph':
  emitted=bounded_json(E/'private-emit-proof.json',4194304);require(emitted['sourceIdentity']==g['_sourceIdentity']and emitted['sourceBindingSha256']==g['sourceBinding']['sha256']and emitted['files']and emitted['qualification']is False,'Actual private graph output absent/refused')
 if name=='IMMEDIATE-positive-carrier-audit':
  audited=bounded_json(E/'positive-carrier-audit.json',4194304);require(audited['sourceIdentity']==g['_sourceIdentity']and audited['sourceBindingSha256']==g['sourceBinding']['sha256']and audited['exactFactoryVerifierAccepted']is True and audited['actualModeAll0600']is True and audited['preExecUmask077PrerequisiteVerified']is True and audited['required']==251 and audited['partialOwnerRows']==17 and audited['missingOwnerRows']==234 and audited['wholeClauseEstablished']==0 and audited['claimEligible']is False and audited['qualificationReady']is False,'Actual full-byte verifier output absent/refused')
 require(time.monotonic()<origin+105,'Verification exceeded original work window')
 failure_phase='publish';logfile=file_hash(logpath);verified={'sourceIdentity':g['_sourceIdentity'],'sourceBindingSha256':g['sourceBinding']['sha256'],'attemptId':attempt,'sourceUnchanged':unchanged,'productExit':child.returncode,'stdioEOF':stdio_eof,'log':logfile,'runtimeDirectoriesStillPresent':runtime,'semantic':semantic,'verifiedAtSeconds':time.monotonic()-origin};write(name+'.worker-verified.json',verified);note('verified',logSha256=logfile['sha256']);final=await_control('finalize')
 record={'exit':child.returncode,'accepted':True,'acceptanceRequiresOuterReceipt':True,'sourceIdentity':g['_sourceIdentity'],'sourceBindingSha256':g['sourceBinding']['sha256'],'sourceUnchanged':True,'attemptId':attempt,'elapsedSeconds':time.monotonic()-origin,'forcedCleanup':False,'errors':[],'remainingProcesses':[],'runtimeDirectoriesStillPresent':[],'temporaryDirectory':str(tmp),'stdioEOF':stdio_eof,'logSha256':logfile['sha256'],'watchdogObservations':final};write(name+'.exit.json',record);note('published');channel.close()
except BaseException as error:
 try:
  if 'E'in globals()and E.is_dir()and 'attempt'in globals():
   write(name+'.partial.json',{'attemptId':attempt,'errorClass':type(error).__name__,'failureStage':failure_phase,'atElapsedSeconds':time.monotonic()-origin,'sourceBindingSha256':g['sourceBinding']['sha256'],'childPid':child.pid if 'child'in globals()else None,'actualChildExitIfKnown':child.returncode if 'child'in globals()else None,'stdioEOF':stdio_eof if 'stdio_eof'in globals()else None,'logFullHash':'UNKNOWN','resultsFullHash':'UNKNOWN','observedSummaryCount':len(scanner.summaries)if 'scanner'in globals()and scanner else None,'claimEligible':False,'positiveCredit':False})
 except BaseException:pass
 try:note('failed',errorClass=type(error).__name__,failureStage=failure_phase)
 except BaseException:pass
 raise SystemExit(1)
