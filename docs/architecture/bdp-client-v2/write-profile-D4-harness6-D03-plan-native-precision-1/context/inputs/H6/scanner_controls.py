"""Future unit controls: call immutable scanner, never a replacement parser.
No top-level execution. Raw resources use original constants without reduction.
"""
import hashlib,json,io
from pathlib import Path
from unittest.mock import patch
ATTEMPT='a'*32

def line(value):return json.dumps(value,separators=(',',':'),allow_nan=False).encode()+b'\n'
def marker(phase,elapsed=0,attempt=ATTEMPT):
 return {'observation':'transactional-set-validation-phase','version':1,'attemptId':attempt,'phase':phase,'sourceElapsedMilliseconds':elapsed}
def refusal(call,contains=None):
 try:call()
 except (RuntimeError,UnicodeError,ValueError) as exc:
  if contains is not None and contains not in str(exc):raise AssertionError(('wrong-refusal',contains,str(exc)))
  return {'refused':True,'exceptionType':type(exc).__name__,'reason':str(exc)[:128]}
 raise AssertionError('subject accepted counterfactual')
def feed(s,data):
 for start in range(0,len(data),65536):s.feed(data[start:start+65536])
def stream(s,parts):
 for part in parts:feed(s,part)
 s.finish();return s

def run(case,subject,root):
 S=subject;start=line(marker('callback-start'));end=line(marker('callback-end',1));report=line({'kind':'installed-valid-set-observation'})
 valid=start+report+end
 def reject(data,why=None):return refusal(lambda:stream(S.LogScanner(ATTEMPT),[data]),why)
 if case=='M01':
  # Every byte boundary includes token boundaries; identical offsets/hash required.
  for split in range(len(valid)+1):
   got=stream(S.LogScanner(ATTEMPT),[valid[:split],valid[split:]])
   assert got.hash.hexdigest()==hashlib.sha256(valid).hexdigest() and got.offset==len(valid)
   assert got.reports[0]['offset']==len(start) and got.reports[0]['bytes']==len(report)
  return {'checks':len(valid)+1,'scope':'scanner only; no report/result/terminal acceptance'}
 if case=='M02':return reject(line(marker('callback-start',attempt='b'*32))+report+line(marker('callback-end',1,'b'*32)),'marker-attempt')
 if case=='M03':return reject(line(marker('callback-start',attempt=None))+report+end,'marker-attempt')
 if case=='M04':
  # Source env->null behavior is independently UNRESOLVED; collector side tested.
  return {'collector':reject(line(marker('callback-start',attempt='not-hex'))+report+end,'marker-attempt'),'unresolved':['source environment-to-null execution is not a scanner function']}
 if case=='M05':return reject(report+end,'marker-order')
 if case=='M06':return reject(start+report,'missing-phase-or-report')
 if case=='M07':return reject(start+start+report+end,'marker-order')
 if case=='M08':return reject(end+start+report,'marker-order')
 if case=='M09':return reject(valid+end,'marker-order')
 if case=='M10':return reject(line({**marker('callback-start'),'extra':True})+report+end,'marker-keys')
 if case=='M11':return reject(b' '*(4097-len(start))+start+report+end,'marker-byte-limit')
 if case=='M12':
  corrupt=start[:40]+b'\xff'+start[40:]
  receipt=reject(corrupt+report+end);path=root/'corrupt.raw';path.write_bytes(corrupt+report+end)
  assert path.read_bytes()==corrupt+report+end
  return receipt
 if case=='M13':return reject(valid[:-1],'incomplete-final-log-line')
 if case=='M14':
  outcomes=[]
  for value in [-1,'1',True]:outcomes.append(reject(start+report+line(marker('callback-end',value)),'marker-source-elapsed'))
  for value in [b'NaN',b'Infinity',b'1e999']:
   outcomes.append(reject(start+report+line(marker('callback-end',1)).replace(b':1}',b':'+value+b'}')))
  return {'subcases':outcomes}
 if case=='M15':return {'subcases':[reject(line(marker('callback-start',2))+report+line(marker('callback-end',1)),'marker-source-order'),reject(start+report+line(marker('callback-end',120001)),'marker-source-elapsed')]}
 if case=='M16':return reject(start+report+line(marker('callback-end',1,'b'*32)),'marker-attempt')
 if case=='J01':return {'subcases':[refusal(lambda:S.strict_json(v,1000),'duplicate-json-key') for v in [b'{"a":1,"a":2}',b'{"a":1,"\\u0061":2}']]}
 if case=='J02':return refusal(lambda:stream(S.LogScanner(ATTEMPT),[b'\xe2',b'\x28\xa1\n']))
 if case=='J03':return {'subcases':[refusal(lambda v=v:S.strict_json(v,100))for v in [b'NaN',b'Infinity',b'1e999']]}
 if case=='J04':return refusal(lambda:S.strict_json(b'['*129+b'0'+b']'*129,1000),'json-depth-limit')
 if case=='J05':
  # Ten subarrays of 100000 nodes stay under individual container bound.
  piece=b'['+b'0,'*99999+b'0]';data=b'['+b','.join([piece]*10)+b']'
  return refusal(lambda:S.strict_json(data,S.RESULTS),'json-node-limit')
 if case=='J06':
  array=b'['+b'0,'*100000+b'0]';obj=b'{'+b','.join(b'"k'+str(i).encode()+b'":0'for i in range(100001))+b'}'
  return {'subcases':[refusal(lambda d=d:S.strict_json(d,S.RESULTS),'json-container-limit')for d in [array,obj]]}
 if case=='J07':
  assert S.strict_json(b'['*128+b'0'+b']'*128,1000) is not None
  assert len(S.strict_json(b'['+b'0,'*99999+b'0]',S.RESULTS))==100000
  assert S.strict_json(b'[0,-1,1.25,1e3]',100)==[0,-1,1.25,1000]
  return {'checks':3}
 if case=='J08':
  p=root/'results.json';p.write_bytes(b' '*(S.RESULTS+1))
  return refusal(lambda:S.file_json(p,S.RESULTS),'json-file-limit')
 if case=='J09':
  # Full actual64MiB+1 line, no reduced mock limit.
  s=S.LogScanner(ATTEMPT);remaining=S.REPORT+1
  def drive():
   nonlocal remaining
   first=True
   while remaining:
    n=min(S.CHUNK,remaining);chunk=(b'{' if first else b' ')+b' '*(n-1)
    if remaining==n:chunk=chunk[:-1]+b'\n'
    s.feed(chunk);remaining-=n;first=False
  return refusal(drive,'selected-line-limit')
 if case=='J10':
  p=root/'raw.log';s=S.LogScanner(ATTEMPT);written=0;refused=None
  with p.open('xb')as output:
   while written<S.RAW+1:
    data=b'x'*min(S.CHUNK,S.RAW+1-written);output.write(data);written+=len(data)
    try:s.feed(data)
    except RuntimeError as exc:refused=str(exc);break
  assert written==S.RAW+1 and p.stat().st_size==written and refused=='raw-log-limit'
  import outer_worker as O
  split=[]
  for label,aggregate in [('below',S.RAW-1),('equal',S.RAW),('over',S.RAW+1)]:
   directory=root/('split-'+label);directory.mkdir();states=[];total=0;crossed=False
   for streamname,size in [('stdout',S.RAW//2),('stderr',aggregate-S.RAW//2)]:
    path=directory/(streamname+'.log');state={'file':path.open('xb'),'sha':hashlib.sha256(),'bytes':0};states.append((path,state));left=size
    try:
     while left:
      block=b'y'*min(65536,left);total,overflow=O.retain_block(state,block,total);crossed=crossed or overflow;left-=len(block)
    finally:state['file'].close()
   assert total==aggregate and crossed==(label=='over')
   exact=[]
   for path,state in states:
    observed=S.file_hash(path);assert observed['bytes']==state['bytes'] and observed['sha256']==state['sha'].hexdigest()
    assert state['bytes']<S.RAW
    exact.append({'stream':path.name,**observed})
   split.append({'label':label,'aggregateBytes':total,'overflow':crossed,'savedReceivedBytesExact':True,'streams':exact,'scope':'actual shared retain_block filesystem/hash primitive; not pipe scheduling or product output'})
  return {'refused':True,'savedBytes':written,'sha256':S.file_hash(p)['sha256'],'splitStreamSubcases':split,'scope':'actual saved raw and shared aggregate primitive; no product log'}
 if case=='J11':
  p=root/'unrelated.raw';s=S.LogScanner(ATTEMPT);h=hashlib.sha256()
  with p.open('xb')as output:
   for _ in range(S.REPORT//S.CHUNK+1):
    data=b'x'*S.CHUNK;output.write(data);s.feed(data);h.update(data)
   output.write(b'\n'+valid);feed(s,b'\n'+valid);h.update(b'\n'+valid)
  s.finish();assert s.hash.hexdigest()==h.hexdigest()==S.file_hash(p)['sha256'] and len(s.reports)==1
  return {'savedBytes':p.stat().st_size,'scope':'bounded skipped line; original raw retained'}
 if case=='J12':
  p=root/'growing.json';p.write_bytes(b'{}');original=Path.open
  class Growing(io.BytesIO):
   def __init__(self):super().__init__(b'{}');self.once=False
   def read(self,*a):
    data=super().read(*a)
    if data and not self.once:
     self.once=True
     with original(p,'ab')as f:f.write(b' ')
    return data
  def opening(path,*a,**kw):return Growing()if path==p and a and a[0]=='rb'else original(path,*a,**kw)
  with patch.object(Path,'open',opening):receipt=refusal(lambda:S.file_hash(p),'file-grew-or-changed')
  return {**receipt,'scope':'real file growth during subject hash; parse lacks stat-stability check and is unresolved separately','unresolved':['growth during file_json parse has no subject stat-stability predicate']}
 raise AssertionError('unregistered scanner control')
