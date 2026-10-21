"""CLOSED availability-only owner; no harness/subject imports or finite control execution.
All auth/hash/file/bootstrap/pipe/receipt work uses the native first clock. The
intact owner and native limiter are assumed; missing receipts fail UNKNOWN.
"""
import time
FIRST=time.monotonic()
import hashlib,json,math,os,select,signal,socket,subprocess,sys
from pathlib import Path
HERE=Path(__file__).resolve().parent
PYTHON='/opt/homebrew/opt/python@3.14/bin/python3.14'
SESSION='2026-10-18T02:16:24Z-549f'

def need(ok,reason):
 if not ok:raise RuntimeError(reason)
def unique(pairs):
 out={}
 for k,v in pairs:need(k not in out,'duplicate-key');out[k]=v
 return out
def read(path,cap):
 path=Path(path);need(path.is_file()and not path.is_symlink()and path.stat().st_size<=cap,'availability-document')
 raw=path.read_bytes();need(len(raw)<=cap,'availability-document-size')
 return json.loads(raw,object_pairs_hook=unique,parse_constant=lambda _:(_ for _ in ()).throw(ValueError('nonfinite')))
def sha(path):return hashlib.sha256(Path(path).read_bytes()).hexdigest()
def pinned(pin):
 need(type(pin)is dict and set(pin)=={'path','sha256'}and Path(pin['path']).is_absolute(),'availability-pin-shape')
 need(sha(pin['path'])==pin['sha256'],'availability-pin');return read(pin['path'],65536)
def emit(value):
 raw=(json.dumps(value,separators=(',',':'),allow_nan=False)+'\n').encode();need(len(raw)<=65536,'availability-receipt-capacity');os.set_blocking(1,False);need(os.write(1,raw)==len(raw),'availability-receipt-write')

def main():
 need(len(sys.argv)==4,'availability-argv');origin=float(sys.argv[1]);release=Path(sys.argv[2]);limiter=int(sys.argv[3]);need(math.isfinite(origin)and origin>0,'availability-origin')
 need(os.getppid()==limiter and os.environ.get('D4_NATIVE_LIMITER')==str(limiter),'availability-native-parent')
 need(0<=FIRST-origin<10,'availability-clock-epoch-or-bootstrap')
 deadline=origin+115;handles=[];receipts=[];captures={};attempt=HERE/'availability-attempt'
 attempt.mkdir() # exclusive; no retry, even after auth/clock failure.
 def left():return deadline-time.monotonic()
 def save(name,value):
  need(left()>0,'availability-before-save-deadline')
  with(attempt/name).open('x')as f:json.dump(value,f,indent=2,allow_nan=False);f.write('\n')
 def capture_snapshot(state,reason):
  streams=[]
  for item in state['streams']:
   identity='UNKNOWN'
   if item['fd']is not None:
    try:
     st=os.fstat(item['fd']);identity={'device':st.st_dev,'inode':st.st_ino,'mode':st.st_mode,'uid':st.st_uid,'bytes':st.st_size,'mtimeNs':st.st_mtime_ns,'ctimeNs':st.st_ctime_ns}
    except OSError as exc:item['identityError']=type(exc).__name__
   elif item['lastIdentity']is not None:identity=item['lastIdentity']
   streams.append({'path':str(item['path']),'receivedBytes':item['received'],'savedBytes':item['saved'],'hashedPrefixBytes':item['hashed'],'savedPrefixSha256':item['hash'].hexdigest()if item['hashKnown']else'UNKNOWN','actualEofObserved':item['eof'],'rawDescriptorClosed':item['closed'],'writeError':item['writeError'],'closeError':item['closeError'],'identityError':item['identityError'],'savedFileIdentity':identity,'fullStreamSha256':'UNKNOWN','fullStreamState':'UNKNOWN until complete actual owner/native/root terminal audit; snapshots are prefixes only','totalReceivedAfterSnapshot':'UNKNOWN if owner killed or later reads/receipt missing'})
  return {'event':'availability-raw-prefix','role':state['role'],'pid':state['process'].pid,'pgid':os.getpgrp(),'reason':reason,'actualExitIfKnown':state['process'].returncode,'atSeconds':time.monotonic()-origin,'streams':streams,'journalError':state['journalError'],'accepted':False,'clockDomainVerified':False,'productCalls':0}
 def checkpoint(state,reason):
  need(left()>0,'availability-capture-deadline');raw=(json.dumps(capture_snapshot(state,reason),separators=(',',':'),allow_nan=False)+'\n').encode()
  need(len(raw)<=65536,'availability-prefix-receipt-capacity')
  try:
   offset=0
   while offset<len(raw):
    need(left()>0,'availability-capture-deadline');count=os.write(state['journal'],raw[offset:]);need(count>0,'availability-prefix-write');offset+=count
  except BaseException as exc:state['journalError']=type(exc).__name__;raise
 def new_capture(p,role):
  state={'role':role,'process':p,'streams':[],'journal':None,'journalError':None};captures[role]=state
  for suffix in ['stdout','stderr']:
   item={'path':attempt/(role+'.'+suffix),'fd':None,'received':0,'saved':0,'hashed':0,'hash':hashlib.sha256(),'hashKnown':True,'eof':False,'closed':False,'writeError':None,'closeError':None,'identityError':None,'lastIdentity':None};state['streams'].append(item)
   item['fd']=os.open(item['path'],os.O_WRONLY|os.O_CREAT|os.O_EXCL|os.O_CLOEXEC,0o600)
  state['journal']=os.open(attempt/(role+'.capture.jsonl'),os.O_WRONLY|os.O_CREAT|os.O_EXCL|os.O_CLOEXEC,0o600)
  checkpoint(state,'exclusive-files-created-before-bootstrap-ACK');return state
 def retain(item,block):
  # Received count precedes fallible output. Unbuffered writes retain exact crossing bytes.
  item['received']+=len(block);offset=0
  try:
   while offset<len(block):
    need(left()>0,'availability-capture-deadline');count=os.write(item['fd'],block[offset:]);need(count>0,'availability-raw-write');item['saved']+=count
    try:item['hash'].update(block[offset:offset+count]);item['hashed']+=count
    except BaseException:item['hashKnown']=False;raise
    offset+=count
  except BaseException as exc:item['writeError']=type(exc).__name__;raise
 def close_capture(state):
  for item in state['streams']:
   if item['fd']is not None:
    try:
     st=os.fstat(item['fd']);item['lastIdentity']={'device':st.st_dev,'inode':st.st_ino,'mode':st.st_mode,'uid':st.st_uid,'bytes':st.st_size,'mtimeNs':st.st_mtime_ns,'ctimeNs':st.st_ctime_ns}
     os.close(item['fd']);item['fd']=None;item['closed']=True
    except OSError as exc:item['closeError']=type(exc).__name__;raise
 def drain(p,seconds,state):
  end=min(deadline,time.monotonic()+seconds);raw=[bytearray(),bytearray()];eof=[False,False];observed_exit=p.returncode
  for stream in [p.stdout,p.stderr]:os.set_blocking(stream.fileno(),False)
  while time.monotonic()<end and(not all(eof)or p.returncode is None):
   for i,stream in enumerate([p.stdout,p.stderr]):
    if eof[i]:continue
    try:block=os.read(stream.fileno(),65536)
    except BlockingIOError:continue
    item=state['streams'][i]
    if not block:eof[i]=True;item['eof']=True;checkpoint(state,'actual-pipe-eof')
    else:
     retain(item,block);checkpoint(state,'received-chunk-retained-before-capacity-check')
     need(item['received']<=65536,'availability-output-capacity');raw[i].extend(block)
   p.poll()
   if p.returncode!=observed_exit:observed_exit=p.returncode;checkpoint(state,'actual-poll-exit')
   select.select([],[],[],.002)
  checkpoint(state,'drain-ended');need(all(eof)and p.returncode is not None,'availability-peer-closure')
  for stream in [p.stdout,p.stderr]:stream.close()
  close_capture(state);checkpoint(state,'raw-files-closed-after-actual-EOF-and-exit')
  need(all(item['writeError']is None and item['closeError']is None and item['identityError']is None and item['hashKnown']and item['received']==item['saved']==item['hashed']and item['lastIdentity']['bytes']==item['saved']for item in state['streams']),'availability-saved-prefix-identity')
  return bytes(raw[0]),bytes(raw[1])
 try:
  save('started.json',{'event':'availability-started','pid':os.getpid(),'pgid':os.getpgrp(),'limiterPid':limiter,'nativeOrigin':origin,'pythonFirst':FIRST,'session':SESSION,'productCalls':0,'noRetry':True})
  r=read(release,65536)
  keys={'session','scope','availabilityOnly','syntheticReleased','productReleased','harnessImplementationSha256','approvedInvocations','nativeHarnessAcceptance','durableRootRelease','noRetry'}
  need(type(r)is dict and set(r)==keys,'availability-release-keys')
  need(r['session']==SESSION and r['scope']=='D4 zero-product clock/bootstrap availability only'and r['availabilityOnly']is True and r['syntheticReleased']is False and r['productReleased']is False and r['approvedInvocations']==['AVAIL--availability']and r['noRetry']is True,'availability-only-release')
  identity=sha(HERE/'implementation-hashes.json');need(r['harnessImplementationSha256']==identity,'availability-implementation')
  native=pinned(r['nativeHarnessAcceptance']);need(native.get('accepted')is True and native.get('availabilityStaticAcceptance')is True and native.get('harnessImplementationSha256')==identity,'availability-other-native')
  durable=pinned(r['durableRootRelease']);need(durable.get('accepted')is True and durable.get('availabilityOnly')is True and durable.get('syntheticReleased')is False and durable.get('productReleased')is False and durable.get('harnessImplementationSha256')==identity,'availability-durable-root')
  gate=read(HERE/'execution-gate.json',65536);need(gate==read(HERE/'execution-gate.closed.json',65536)and gate=={'session':SESSION,'syntheticReleased':False,'productReleased':False,'approvedCases':[],'noRetry':True},'availability-closed-controls')
  index=read(HERE/'implementation-hashes.json',16777216)
  need(type(index)is dict and index and all(type(v)is str and len(v)==64 for v in index.values()),'availability-index')
  for n,d in index.items():
   path=HERE/n;need(path.resolve().is_relative_to(HERE)and not path.is_symlink()and sha(path)==d,'availability-input-changed:'+n);need(left()>0,'availability-hash-deadline')
  for tool in read(HERE/'tool-identity-read.json',65536)['tools']:need(sha(tool['path'])==tool['sha256'],'availability-tool-changed')
  need(time.monotonic()-origin<10,'availability-auth-before10')
  nonce=os.urandom(16).hex()
  for role in ['term','kill']:
   parent,child=socket.socketpair();parent.setblocking(False)
   mask=signal.pthread_sigmask(signal.SIG_BLOCK,{signal.SIGTERM,signal.SIGINT})
   try:
    p=subprocess.Popen([PYTHON,'-I','-B',str(HERE/'bootstrap.py'),str(child.fileno()),str(origin),'115',nonce,PYTHON,'-I','-B',str(HERE/'availability_peer.py'),role,nonce,str(origin)],pass_fds=[child.fileno()],start_new_session=False,stdin=subprocess.DEVNULL,stdout=subprocess.PIPE,stderr=subprocess.PIPE)
    handles.append(p) # FIRST after Popen, before metadata/IO. Same owner group.
   finally:signal.pthread_sigmask(signal.SIG_SETMASK,mask)
   state=new_capture(p,role)
   child.close();before=time.monotonic();need(before<deadline,'availability-before-ack')
   ack=(json.dumps({'event':'execute','nonce':nonce,'pid':p.pid},separators=(',',':'))+'\n').encode();need(parent.send(ack)==len(ack),'availability-ack-write')
   boot=b'';end=min(deadline,time.monotonic()+5)
   while b'\n'not in boot and time.monotonic()<end:
    if select.select([parent],[],[],.01)[0]:
     chunk=parent.recv(4096);need(bool(chunk),'availability-bootstrap-eof');boot+=chunk;need(len(boot)<=4096,'availability-bootstrap-capacity')
   need(boot==(json.dumps({'event':'executing','nonce':nonce,'pid':p.pid},separators=(',',':'))+'\n').encode(),'availability-bootstrap-ack');parent.close()
   out,err=drain(p,5,state)
   need(not err,'availability-peer-stderr')
   if role=='term':
    v=json.loads(out,object_pairs_hook=unique);need(p.returncode==0 and v['event']=='availability-peer'and v['nonce']==nonce and v['pid']==p.pid and v['parent']==os.getpid()and v['pgid']==os.getpgrp()and v['termIntUnblocked']is True and v['observedSignals']==[int(signal.SIGTERM)]and before<=v['firstMonotonic']<=v['lastMonotonic']<=time.monotonic(),'availability-term-or-clock')
   else:need(p.returncode==-int(signal.SIGKILL)and out==b'','availability-kill');v=None
   receipts.append({'role':role,'pid':p.pid,'pgid':os.getpgrp(),'bootstrapAck':True,'rawExit':p.returncode,'stdoutEof':True,'stderrEof':True,'stdoutBytes':len(out),'stderrBytes':len(err),'stdoutSha256':hashlib.sha256(out).hexdigest(),'stderrSha256':hashlib.sha256(err).hexdigest(),'peer':v,'atSeconds':time.monotonic()-origin})
   checkpoint(state,'peer-oracle-observed-prefix-only')
   try:os.close(state['journal'])
   except OSError as exc:state['journalError']='close:'+type(exc).__name__;raise
   state['journal']=None
   save(role+'.json',receipts[-1])
  value={'event':'availability-owner-terminal','pid':os.getpid(),'pgid':os.getpgrp(),'limiterPid':limiter,'nativeOrigin':origin,'pythonFirst':FIRST,'pythonLast':time.monotonic(),'pythonClockInfo':vars(time.get_clock_info('monotonic')),'peers':receipts,'allDirectChildrenReaped':all(p.returncode is not None for p in handles),'productCalls':0,'syntheticReleased':False,'clockDomainVerified':False,'status':'OBSERVED_PENDING_NATIVE_LIMITER_AND_FRESH_ROOT_AUDIT','harnessImplementationSha256':identity}
  save('terminal.json',value);need(time.monotonic()<deadline,'availability-terminal-deadline');emit(value);return 0 if time.monotonic()<deadline else 1
 except BaseException as exc:
  try:save('failure.json',{'event':'availability-failed','type':type(exc).__name__,'reason':str(exc)[:256],'pids':[p.pid for p in handles],'status':'FAILED/UNKNOWN','clockDomainVerified':False,'productCalls':0,'actualExitSnapshots':[{'pid':p.pid,'actualExitIfKnown':p.returncode}for p in handles],'rawCaptures':[capture_snapshot(state,'owner-failure-prefix-only')for state in captures.values()]})
  except BaseException:pass
  return 1
 finally:
  for state in captures.values():
   try:
    close_capture(state)
    if state['journal']is not None and left()>0:checkpoint(state,'owner-finalization-prefix-only')
   except BaseException:pass # Missing/failed receipt stays UNKNOWN, never a successful terminal.
   finally:
    if state['journal']is not None:
     try:os.close(state['journal'])
     except OSError:pass # No positive evidence from a failed/missing journal close.
  for p in handles:
   if p.returncode is None:
    try:p.kill() # unreaped direct-child ownership; never numeric historical groups.
    except ProcessLookupError:pass
  # No blocking postdeadline wait. Missing closure remains FAILED/UNKNOWN.
if __name__=='__main__':raise SystemExit(main())
