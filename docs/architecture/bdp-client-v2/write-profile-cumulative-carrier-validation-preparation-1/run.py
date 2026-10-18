"""Responsive single-stage watchdog. Every slow operation belongs to owned workers.
Usage run.py final STAGE. No automatic retries; root must bind the outer exit receipt.
"""
import time
ORIGIN=time.monotonic() # first script clock, before all remaining imports/setup
import os,sys,socket,subprocess,signal,select,json,datetime
from pathlib import Path
sys.dont_write_bytecode=True

def main():
 begin=ORIGIN # includes all remaining module imports and watchdog setup.
 if len(sys.argv)!=3 or sys.argv[1]!='final':return 1
 stage=sys.argv[2];product=False;term=115;kill=119;outer=120 # carrier stages retain ordinary strict bounds, no D4 exception
 here=Path(__file__).parent;parent,child=socket.socketpair();parent.setblocking(False)
 blocked=signal.pthread_sigmask(signal.SIG_BLOCK,{signal.SIGTERM,signal.SIGINT})
 worker=subprocess.Popen([sys.executable,str(here/'worker.py'),str(child.fileno()),stage,str(begin)],pass_fds=[child.fileno()],start_new_session=True,stdin=subprocess.DEVNULL,stdout=subprocess.DEVNULL,stderr=subprocess.DEVNULL);child.close()
 checks={};identity_observations={};last_census_at=None
 groups={worker.pid:{'leader':worker.pid,'direct':True,'gone':False,'expectedParent':os.getpid(),'createdEpoch':time.time(),'registeredAt':time.monotonic(),'observed':False}};identities={};seen=set();closed=set();uncertain=False;failure=None;pending=b'';received=b'';census=None;psout=b'';pserr=b'';psbegin=0;next_ps=begin;start=None;end=None;flush=None;attempt=None;source_binding=None;created=False;primary=None;leader_exit=False;need_quiet=False;verified=False;published=False;finalized=False;worker_reaped=False;term_sent=False;kill_sent=False;failure_at=None;last_rows=[];census_generation=0;quiet_generation=None;publisher=None;publish_socket=None;publish_bytes=b'';signal_receipts=[];last_term=begin;last_kill=begin
 def refuse(reason):
  nonlocal failure,failure_at
  if failure is None:failure=reason;failure_at=time.monotonic()
 def enqueue(event,**values):
  nonlocal pending
  raw=(json.dumps({'event':event,**values},separators=(',',':'))+'\n').encode()
  if len(raw)>4096 or len(pending)+len(raw)>8192:refuse('control-capacity');return
  pending+=raw
 def birth_epoch(row):
  parts=row['birth'].split();hour,minute,second=map(int,parts[3].split(':'));month=['Jan','Feb','Mar','Apr','May','Jun','Jul','Aug','Sep','Oct','Nov','Dec'].index(parts[1])+1
  return datetime.datetime(int(parts[4]),month,int(parts[2]),hour,minute,second,tzinfo=datetime.timezone.utc).timestamp()
 def identity_matches(row):
  old=identities.get(row['pid'])
  if old is None or old['birth']!=row['birth'] or old['pgid']!=row['pgid']:return False
  if row['state'].startswith('Z'):return True # terminal comm is not executable evidence
  return not old['terminalSeen'] and old['liveComm'] is not None and old['liveComm']==row['comm']
 def observe(rows):
  nonlocal uncertain,last_rows,census_generation,quiet_generation,last_census_at
  census_generation+=1;last_census_at=time.monotonic();last_rows=rows;byid={r['pid']:r for r in rows};owned=set()
  # Keep direct Popen unreaped through all remaining members. This anchor
  # covers the contained product's spawn, failed IPC and worker leader exit.
  for gid,state in groups.items():
   current=[r for r in rows if r['pgid']==gid]
   if not current:
    if psbegin>=state['registeredAt']:closed.add(gid)
    continue
   if gid in closed:uncertain=True;continue
   live_anchor=any(identity_matches(r)for r in current)
   leader=byid.get(state['leader']);handle=worker if gid==worker.pid else publisher if publisher is not None and gid==publisher.pid else None
   direct_unreaped=state['direct'] and handle is not None and handle.returncode is None
   if leader and direct_unreaped:
    birth=birth_epoch(leader)
    if leader['ppid']!=state['expectedParent'] or leader['pgid']!=gid or abs(birth-state['createdEpoch'])>=2:uncertain=True;continue
    live_anchor=True;state['observed']=True
   elif direct_unreaped:live_anchor=True
   if not live_anchor:uncertain=True;continue
   owned.update(r['pid']for r in current)
  changed=True
  while changed:
   changed=False
   for r in rows:
    if r['pid']not in owned and r['ppid']in owned:owned.add(r['pid']);changed=True
  for r in rows:
   if r['pid']in owned:
    old=identities.get(r['pid']);terminal=r['state'].startswith('Z')
    if old is not None and not identity_matches(r):uncertain=True;continue
    if old is None:identities[r['pid']]={'birth':r['birth'],'pgid':r['pgid'],'liveComm':None if terminal else r['comm'],'terminalSeen':terminal}
    elif terminal:old['terminalSeen']=True
    seen.add(r['pid'])
    if r['pid']not in identity_observations:identity_observations[r['pid']]={**r,'firstSeenSeconds':last_census_at-begin}
    identity_observations[r['pid']].update({'lastSeenSeconds':last_census_at-begin,'lastState':r['state'],'terminalObserved':identities[r['pid']]['terminalSeen']})
    if r['pgid']not in groups:
     if r['pgid']not in owned:uncertain=True;continue
     groups[r['pgid']]={'leader':r['pgid'],'direct':False,'gone':True,'registeredAt':last_census_at,'observed':False}
  if len(identities)>256:uncertain=True;refuse('ownership-capacity')
  live=[r for r in rows if r['pid']in owned]
  # Zombie is not absence. Exclude only the worker PID, never its whole PGID.
  other=[r for r in live if r['pid']!=worker.pid]
  if need_quiet and leader_exit and not other and not uncertain and quiet_generation is None:quiet_generation=census_generation;enqueue('quiet')
  if primary is not None:
   row=byid.get(primary['pid'])
   if row is not None and row['pgid']==worker.pid and row['ppid']==worker.pid and identity_matches(row):
    if abs(birth_epoch(row)-primary['createdEpoch'])>=2:uncertain=True
    else:primary['observed']=True
  return live
 def reap_after_quiet(handle,live,rows):
  # poll() reaps; use only after a fresh post-registration census has no other
  # owned member and its direct leader is terminal/absent. Then require a
  # subsequent fresh group-absence census before publication/acceptance.
  if handle.returncode is not None:return True
  row=next((r for r in rows if r['pid']==handle.pid),None)
  if uncertain or any(r['pid']!=handle.pid for r in live) or (row is not None and not row['state'].startswith('Z')):return False
  if psbegin<groups[handle.pid]['registeredAt']:return False
  handle.poll()
  return handle.returncode is not None
 def signal_owned(sig):
  for gid,state in list(groups.items()):
   if gid==os.getpgrp():refuse('supervisor-group');continue
   anchor=gid not in closed and last_census_at is not None and time.monotonic()-last_census_at<=.2 and any(r['pgid']==gid and identity_matches(r)for r in last_rows)
   handle=worker if gid==worker.pid else publisher if publisher is not None and gid==publisher.pid else None
   direct=gid not in closed and handle is not None and handle.returncode is None
   if not(anchor or direct):continue
   try:
    os.killpg(gid,sig)
    if len(signal_receipts)<32:signal_receipts.append({'group':gid,'signal':int(sig),'atSeconds':time.monotonic()-begin})
    else:refuse('signal-receipt-capacity')
   except ProcessLookupError:pass
   except OSError:refuse('signal-unknown')
 def interrupt(sig,frame):refuse('interrupted')
 old={s:signal.signal(s,interrupt)for s in [signal.SIGTERM,signal.SIGINT]}
 signal.pthread_sigmask(signal.SIG_SETMASK,blocked)
 try:
  while time.monotonic()-begin<outer:
   now=time.monotonic();elapsed=now-begin
   if publish_socket is not None and publish_bytes:
    try:n=publish_socket.send(publish_bytes);publish_bytes=publish_bytes[n:]
    except BlockingIOError:pass
    except OSError:refuse('publication-unavailable')
   if pending:
    try:n=parent.send(pending);pending=pending[n:]
    except BlockingIOError:pass
    except OSError:refuse('control-unavailable')
   try:
    block=parent.recv(4096)
    if block:
     received+=block
     if len(received)>8192:refuse('control-overflow')
     while b'\n'in received:
      line,received=received.split(b'\n',1)
      if len(line)>4096:refuse('control-record-overflow');continue
      try:value=json.loads(line)
      except (ValueError,UnicodeError):refuse('control-malformed');continue
      now=time.monotonic();elapsed=now-begin # actual complete control receipt, never pre-read time
      event=value.get('event')
      if event=='attempt':
       if attempt is not None or type(value.get('attemptId'))is not str or len(value['attemptId'])!=32 or any(c not in '0123456789abcdef'for c in value['attemptId']):refuse('attempt-invalid')
       else:
        attempt=value['attemptId'];source_binding=value.get('sourceBindingSha256')
        if type(source_binding)is not str or len(source_binding)!=64 or any(c not in '0123456789abcdef'for c in source_binding):refuse('attempt-source-binding')
      elif event=='created':
       pid=value.get('pid')
       if created or type(pid)is not int or pid<=0 or value.get('pgid')!=worker.pid or value.get('ppid')!=worker.pid:refuse('creation-identity')
       else:
        primary={'pid':pid,'pgid':worker.pid,'parent':worker.pid,'createdEpoch':value['createdEpoch'],'observed':False};created=True;enqueue('creation-ack')
      elif event=='check-created':
       pid=value.get('pid')
       if not leader_exit or quiet_generation is None or type(pid)is not int or pid<=0 or pid in checks or value.get('pgid')!=worker.pid or value.get('ppid')!=worker.pid or value.get('role')!='pinned-schema-audit':refuse('check-identity')
       else:checks[pid]={'pid':pid,'pgid':worker.pid,'parent':worker.pid,'role':value['role'],'createdAtSeconds':now-begin,'exit':None,'settled':False}
      elif event=='check-exit':
       pid=value.get('pid')
       if pid not in checks or checks[pid]['settled'] or type(value.get('code'))is not int:refuse('check-terminal')
       else:checks[pid].update({'exit':value['code'],'settled':True,'exitReceiptSeconds':now-begin})
      elif event=='phase':
       if not product:refuse('unexpected-phase')
       elif value.get('phase')=='callback-start' and start is None:
        start=now
        if elapsed>=10:refuse('start-late')
       elif value.get('phase')=='callback-end' and start is not None and end is None:
        end=now
        if now>=min(start+120,begin+130):refuse('end-late')
       else:refuse('phase-order')
      elif event=='leader-exit':
       if primary is None or value.get('pid')!=primary['pid'] or leader_exit or type(value.get('code'))is not int:refuse('leader-identity')
       else:
        leader_exit=True;primary['exit']=value['code']
        if product and not primary['observed']:refuse('product-leader-never-authenticated')
      elif event=='need-quiet':need_quiet=True
      elif event=='verified':
       if now>=begin+105:refuse('verified-after-work-cutoff') # actual receipt, not worker pre-write timestamp
       if verified or (product and (end is None or len(checks)!=1 or not all(c['settled']and c['exit']==0 for c in checks.values()))):refuse('verified-order')
       verified=True;flush=now
       if product and now>=min(end+5,begin+135):refuse('flush-late')
      elif event=='published':published=True
      elif event=='failed':refuse('owned-worker-failure')
      else:refuse('unknown-control')
   except BlockingIOError:pass
   except OSError:refuse('control-read')
   if product and not failure:
    if start is None and elapsed>=10:refuse('start-missing')
    elif start is not None and end is None and now>=min(start+120,begin+130):refuse('end-missing')
    elif end is not None and not verified and now>=min(end+5,begin+135):refuse('flush-incomplete')
    elif verified and now>=min(flush+10,begin+145):refuse('drain-incomplete')
   if elapsed>=105 and not verified:refuse('work-cutoff') # all gate/setup/hash/proof/parse work is inside this origin
   if elapsed>=term:refuse('TERM-boundary')
   if failure and not kill_sent and (not term_sent or now-last_term>=.2):signal_owned(signal.SIGTERM);term_sent=True;last_term=now
   if failure and now>=min(begin+kill,failure_at+4) and (not kill_sent or now-last_kill>=.2):signal_owned(signal.SIGKILL);kill_sent=True;last_kill=now
   if census is None and now>=next_ps:
    census=subprocess.Popen(['/bin/ps','-axo','pid=,ppid=,pgid=,stat=,lstart=,comm='],stdout=subprocess.PIPE,stderr=subprocess.PIPE,start_new_session=True,env={**os.environ,'LC_ALL':'C','TZ':'UTC'});psbegin=now;psout=b'';pserr=b'';os.set_blocking(census.stdout.fileno(),False);os.set_blocking(census.stderr.fileno(),False)
   if census is not None:
    for stream,iserr in [(census.stdout,False),(census.stderr,True)]:
     try:
      for _ in range(5):
       block=os.read(stream.fileno(),65536)
       if not block:break
       if iserr:pserr+=block
       else:psout+=block
       if len(psout)>262144 or len(pserr)>4096:break
     except BlockingIOError:pass
    if len(psout)>262144 or len(pserr)>4096:refuse('census-capacity')
    if time.monotonic()-psbegin>=.2 and census.poll()is None:
     census.kill();refuse('census-deadline')
    if census.poll()is not None:
     if census.returncode!=0 or pserr:refuse('census-failed')
     else:
      rows=[]
      try:
       for line in psout.decode('utf-8','strict').splitlines():
        fields=line.split(None,9)
        if len(fields)!=10:raise ValueError('row')
        rows.append({'pid':int(fields[0]),'ppid':int(fields[1]),'pgid':int(fields[2]),'state':fields[3],'birth':' '.join(fields[4:9]),'comm':fields[9]})
       live=observe(rows)
       if uncertain:refuse('ownership-unknown')
       if verified and not finalized and not failure and not any(r['pid']!=worker.pid for r in live):
        # The worker itself is the only permitted remaining task group here.
        enqueue('finalize',clockOrigin=begin,flushAt=flush-begin,drainObservationAt=now-begin,ownedPids=sorted(seen),remainingNonWorkerGroups=[]);finalized=True
       if not worker_reaped:worker_reaped=reap_after_quiet(worker,live,rows)
       if worker_reaped:
        if not live and worker.pid in closed and publisher is None:
         accepted=published and worker.returncode==0 and not failure and (not product or flush is not None) and time.monotonic()<(min(flush+10,begin+term) if product else begin+term)
         a,b=socket.socketpair();a.setblocking(False)
         publish_mask=signal.pthread_sigmask(signal.SIG_BLOCK,{signal.SIGTERM,signal.SIGINT})
         publisher=subprocess.Popen([sys.executable,str(here/'worker.py'),'publish',str(b.fileno()),stage],pass_fds=[b.fileno()],start_new_session=True,stdin=subprocess.DEVNULL,stdout=subprocess.DEVNULL,stderr=subprocess.DEVNULL)
         groups[publisher.pid]={'leader':publisher.pid,'direct':True,'gone':False,'expectedParent':os.getpid(),'createdEpoch':time.time(),'registeredAt':time.monotonic(),'observed':False}
         b.close();publish_socket=a
         signal.pthread_sigmask(signal.SIG_SETMASK,publish_mask)
         record={'stage':stage,'watchdogPid':os.getpid(),'accepted':bool(accepted),'attemptId':attempt,'sourceBindingSha256':source_binding,'workerExit':worker.returncode,'failure':failure,'elapsedSeconds':time.monotonic()-begin,'sourcePhaseStartSeconds':None if start is None else start-begin,'sourcePhaseEndSeconds':None if end is None else end-begin,'flushSeconds':None if flush is None else flush-begin,'ownedPids':sorted(seen),'ownedIdentityObservations':list(identity_observations.values()),'knownGroups':groups,'ownedChecks':list(checks.values()),'primaryCreation':primary,'productQuietGeneration':quiet_generation,'finalCensusSeconds':last_census_at-begin,'ownershipUnknown':uncertain,'signalReceipts':signal_receipts,'allKnownNonPublisherGroupsObservedEmpty':True,'outerRawExitStillRequiresRootReceipt':True}
         publish_bytes=(json.dumps(record,separators=(',',':'))+'\n').encode()
         if len(publish_bytes)>1048576:refuse('publication-capacity');publish_bytes=b''
        elif publisher is not None and reap_after_quiet(publisher,live,rows) and not live and publisher.pid in closed:
         return 0 if published and worker.returncode==0 and publisher.returncode==0 and not failure and (not product or flush is not None) and time.monotonic()<(min(flush+10,begin+term) if product else begin+term) else 1
      except (ValueError,UnicodeError):refuse('census-malformed')
     census.stdout.close();census.stderr.close();census=None;next_ps=time.monotonic()+.08
   select.select([parent],[],[],.02) # <=20ms slice; no file/check/hash/parser task on loop.
  refuse('outer-boundary');signal_owned(signal.SIGKILL);return 1
 finally:
  for sig in old:signal.signal(sig,signal.SIG_IGN)
  if census is not None and census.poll()is None:census.kill()
  parent.close()
  if publish_socket is not None:publish_socket.close()
  for sig,handler in old.items():signal.signal(sig,handler)
if __name__=='__main__':raise SystemExit(main())
