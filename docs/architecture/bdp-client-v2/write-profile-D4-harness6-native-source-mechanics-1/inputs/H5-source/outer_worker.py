"""CLOSED owned authorization/control or IO worker, always under outer clock.
No product program is invoked. All blocking filesystem/hash/guard work is here.
"""
import time
import hashlib,json,os,select,signal,socket,sys,array
from pathlib import Path
HERE=Path(__file__).resolve().parent
sys.dont_write_bytecode=True
sys.path.insert(0,str(HERE))
import owned_spawn as B

def write(path,value):
 with path.open('x')as stream:json.dump(value,stream,indent=2,allow_nan=False);stream.write('\n')
def remaining():return float(os.environ['D4_OUTER_ORIGIN'])+float(os.environ['D4_OUTER_TERM'])-time.monotonic()
def receive(channel,cap=4096):
 raw=b''
 while remaining()>0:
  channel.settimeout(min(.2,max(.001,remaining())))
  try:block=channel.recv(cap)
  except socket.timeout:continue
  if not block:raise RuntimeError('owned-worker-eof')
  raw+=block
  if len(raw)>cap:raise RuntimeError('owned-worker-control-capacity')
  if b'\n'in raw:
   line,tail=raw.split(b'\n',1)
   if tail:raise RuntimeError('owned-worker-control-tail')
   return json.loads(line)
 raise RuntimeError('owned-worker-deadline')

RAW_TOTAL=268435456

def retain_block(stream,block,total):
 # Actual IO and J10 call this same primitive. No caller-supplied capacity.
 # Write and hash every received byte BEFORE reporting threshold overflow.
 stream['file'].write(block);stream['file'].flush();stream['sha'].update(block);stream['bytes']+=len(block)
 total+=len(block)
 return total,total>RAW_TOTAL

def main():
 mode=sys.argv[1];attempt=Path(sys.argv[2]);channel=socket.socket(fileno=int(sys.argv[3]))
 if mode=='control':
  import harness
  case,variant,release=sys.argv[4:7]
  # Same clock as native external limiter, not this module import time.
  harness.ORIGIN=float(os.environ['D4_OUTER_ORIGIN'])
  harness.authorized(case,Path(release),variant)
  B.note('authorized',case=case,variant=variant)
  if receive(channel)!={'event':'execute-control'}:raise RuntimeError('control-execution-not-acknowledged')
  os.environ['D4_SYNTHETIC_ATTEMPT']=str(attempt)
  raw=harness.child(case,Path(release),attempt,variant)
  B.note('control-finished',rawExit=raw)
  return raw
 if mode=='io':
  # Continue owned draining after the owner's overflow TERM. KILL/deadline stays unchanged.
  cleanup_signals=[]
  signal.signal(signal.SIGTERM,lambda sig,frame:cleanup_signals.append(int(sig)))
  signal.signal(signal.SIGINT,lambda sig,frame:cleanup_signals.append(int(sig)))
  attempt.parent.mkdir(exist_ok=True);attempt.mkdir() # sole exclusive attempt, no retries.
  started=json.loads(sys.argv[4]);write(attempt/'started.json',started)
  B.note('io-ready')
  data,anc,flags,_=channel.recvmsg(4096,socket.CMSG_SPACE(2*array.array('i').itemsize))
  if flags or data!=b'stdio\n':raise RuntimeError('io-descriptor-envelope')
  fds=[]
  for level,kind,value in anc:
   if level!=socket.SOL_SOCKET or kind!=socket.SCM_RIGHTS:raise RuntimeError('io-ancillary-type')
   a=array.array('i');a.frombytes(value[:len(value)-len(value)%a.itemsize]);fds.extend(a)
  if len(fds)!=2:raise RuntimeError('io-descriptor-count')
  streams={fd:{'name':name,'file':(attempt/name).open('xb'),'sha':hashlib.sha256(),'bytes':0,'eof':False}for fd,name in zip(fds,['stdout.log','stderr.log'])}
  total=0;overflow=False
  for fd in fds:os.set_blocking(fd,False)
  while any(not s['eof']for s in streams.values()) and remaining()>0:
   for fd in select.select([fd for fd,s in streams.items()if not s['eof']],[],[],.02)[0]:
    block=os.read(fd,65536);s=streams[fd]
    if not block:s['eof']=True;os.close(fd);s['file'].close();continue
    total,crossed=retain_block(s,block,total)
    if crossed and not overflow:
     overflow=True;B.note('io-overflow',aggregateBytes=total,limit=RAW_TOTAL)
    # Keep draining saved bytes on the original clock while the owner stops producers.
  receipt={s['name']:{'bytes':s['bytes'],'sha256':s['sha'].hexdigest(),'eof':s['eof']}for s in streams.values()}
  # Partial prefix hashes are labelled honestly if a descriptor did not reach EOF.
  for stream in streams.values():
   if not stream['file'].closed:stream['file'].close()
  complete=all(s['eof']for s in streams.values())
  write(attempt/'stdio.json',{'streams':receipt,'aggregateBytes':total,'aggregateLimit':RAW_TOTAL,'overflow':overflow,'cleanupSignals':cleanup_signals,'allEof':complete,'hashScope':'complete received stream'if complete else'saved received prefix only; full stream UNKNOWN','accepted':False})
  if not complete:raise RuntimeError('io-eof-incomplete')
  if overflow:raise RuntimeError('io-full-raw-capacity')
  B.note('io-finished',stdio=receipt)
  # Keep IO/receipt worker unreaped until parent sends its final tentative
  # census record. It does all remaining result hashing/publication on-clock.
  record=receive(channel,1048576)
  target=attempt/'control-result.json'
  record['controlResult']=None
  if target.is_file():
   raw=target.read_bytes()
   if len(raw)>1048576:raise RuntimeError('control-result-capacity')
   value=json.loads(raw);record['controlResult']={'sha256':hashlib.sha256(raw).hexdigest(),'bytes':len(raw),'status':value.get('status'),'unresolved':value.get('unresolved')}
  record['publicationSeconds']=time.monotonic()-float(os.environ['D4_OUTER_ORIGIN']);record['requiresExternalLimiterExitAndFreshRootAudit']=True
  if remaining()<=0:raise RuntimeError('publication-deadline')
  write(attempt/'outer-owner-tentative.json',record)
  B.note('publication-finished',at=time.monotonic())
  return 0
 raise RuntimeError('outer-worker-role')
if __name__=='__main__':
 try:code=main()
 except BaseException as exc:
  try:B.note('worker-error',error=type(exc).__name__,reason=str(exc)[:256])
  except BaseException:pass
  raise
 raise SystemExit(code)
