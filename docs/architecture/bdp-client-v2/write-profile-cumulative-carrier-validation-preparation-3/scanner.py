"""Owned worker bounded private carrier data scanner/acquisition. No product imports.
WIP: not independently reviewed and never executed/imported during preparation.
"""
import codecs,hashlib,json,math,os,re,stat
from pathlib import Path
CHUNK=65536
RAW=67108864
RESULTS=16777216
WRAPPER=4194304
SIDECAR=1048576
SUMMARY=1048576
class Refusal(RuntimeError):pass
def need(ok,why):
 if not ok:raise Refusal(why)
def unique(pairs):
 out={}
 for k,v in pairs:
  need(k not in out,'duplicate-json-key');out[k]=v
 return out
def invalid(_):raise Refusal('nonfinite-json')
def strict_json(data,maximum):
 need(type(data)is bytes and len(data)<=maximum,'json-byte-limit')
 text=data.decode('utf-8','strict');n=len(text);i=0;nodes=0;deepest=0;largest=0
 def ws():
  nonlocal i
  while i<n and text[i] in ' \r\n\t':i+=1
 def string(key=False):
  nonlocal i
  need(i<n and text[i]=='"','json-string');start=i;i+=1
  while i<n:
   c=text[i];i+=1
   if c=='"':return json.loads(text[start:i]) if key else None
   need(ord(c)>=32,'json-control')
   if c=='\\':
    need(i<n,'json-escape');c=text[i];i+=1
    if c=='u':need(i+4<=n and re.fullmatch('[a-fA-F0-9]{4}',text[i:i+4]) is not None,'json-unicode');i+=4
    else:need(c in '"\\/bfnrt','json-escape')
  raise Refusal('json-string-incomplete')
 def value(depth):
  nonlocal i,nodes,deepest,largest
  ws();need(i<n,'json-value');nodes+=1;need(nodes<=100000,'json-node-limit');c=text[i]
  if c in '{[':
   current=depth+1;deepest=max(deepest,current);need(current<=128,'json-depth-limit');i+=1;ws();entries=0;keys=set();end='}' if c=='{' else ']'
   if i<n and text[i]==end:i+=1;return
   while True:
    if c=='{':
     ws();key=string(True);need(key not in keys,'duplicate-json-key');keys.add(key);ws();need(i<n and text[i]==':','json-colon');i+=1
    value(current);entries+=1;largest=max(largest,entries);need(entries<=10000,'json-container-limit');ws();need(i<n,'json-container-incomplete')
    if text[i]==end:i+=1;return
    need(text[i]==',','json-delimiter');i+=1
  elif c=='"':string()
  elif text.startswith('true',i):i+=4
  elif text.startswith('false',i):i+=5
  elif text.startswith('null',i):i+=4
  else:
   m=re.compile(r'-?(?:0|[1-9][0-9]*)(?:\.[0-9]+)?(?:[eE][+-]?[0-9]+)?').match(text,i)
   need(m is not None,'json-number');need(math.isfinite(float(m[0])),'nonfinite-json');i=m.end()
 value(0);ws();need(i==n,'json-trailing')
 # Tree allocation occurs only after the full bounded structural scan.
 obj=json.loads(text,object_pairs_hook=unique,parse_constant=invalid)
 actual_nodes=0;actual_depth=0;actual_entries=0
 def count(v,depth):
  nonlocal actual_nodes,actual_depth,actual_entries
  actual_nodes+=1
  if type(v) in [dict,list]:
   actual_depth=max(actual_depth,depth+1);actual_entries=max(actual_entries,len(v))
   for child in (v.values() if type(v)is dict else v):count(child,depth+1)
  elif type(v) in [int,float]:need(math.isfinite(v),'nonfinite-json')
 count(obj,0);need((nodes,deepest,largest)==(actual_nodes,actual_depth,actual_entries),'json-scan-parser-disagreement')
 return obj

class SummaryScanner:
 def __init__(self):
  self.diagnostic_tail=b'';self.unhandled=False;self.offset=0;self.line_start=0;self.line_bytes=0;self.line=bytearray();self.skipped=False;self.prefix=None;self.summaries=[];self.hash=hashlib.sha256();self.utf8=codecs.getincrementaldecoder('utf-8')('strict')
 def feed(self,data):
  need(type(data)is bytes and len(data)<=CHUNK,'log-chunk');probe=self.diagnostic_tail+data;self.unhandled=self.unhandled or re.search(b'Unhandled Errors|Unhandled Rejection|Uncaught Exception|Worker exited unexpectedly',probe)is not None;self.diagnostic_tail=probe[-64:];self.offset+=len(data);need(self.offset<=RAW,'raw-log-limit');self.hash.update(data);self.utf8.decode(data)
  cursor=0
  while cursor<len(data):
   newline=data.find(b'\n',cursor);end=len(data)if newline<0 else newline+1;part=data[cursor:end];cursor=end
   self.line_bytes+=len(part)
   if not self.skipped:
    self.line.extend(part)
    if len(self.line)>SUMMARY:
     need(not bytes(self.line).lstrip(b' \t\r\n').startswith(b'{'),'object-line-limit');self.line.clear();self.skipped=True
   if newline>=0:
    if not self.skipped:self.consume(bytes(self.line),self.line_start)
    self.line_start+=self.line_bytes;self.line_bytes=0;self.line.clear();self.skipped=False
 def consume(self,line,offset):
  raw=line.strip(b' \t\r\n')
  if not raw.startswith(b'{'):return
  value=strict_json(raw,SUMMARY)
  if type(value)is dict and value.get('kind')=='installed-tx-seventeen-partial-summary':
   need(not self.summaries,'duplicate-positive-summary');self.summaries.append({'offset':offset,'bytes':len(line),'sha256':hashlib.sha256(line).hexdigest(),'value':value})
 def finish(self):
  self.utf8.decode(b'',final=True);need(not self.unhandled,'unhandled-runtime-diagnostic');need(not self.line and not self.skipped and self.line_bytes==0,'incomplete-final-log-line');need(len(self.summaries)==1,'positive-summary-missing')

def secure_read(path,root,maximum):
 """Read existing source-written mode0600 evidence; never create/chmod/rewrite it."""
 path=Path(path);root=Path(root);root_stat=root.lstat()
 need(root.is_absolute()and stat.S_ISDIR(root_stat.st_mode)and not root.is_symlink()and stat.S_IMODE(root_stat.st_mode)==0o700 and root_stat.st_uid==os.getuid(),'private-root')
 need(path.is_absolute()and path.parent==root and path.name not in ['','.', '..'],'private-evidence-path')
 initial=path.lstat();need(stat.S_ISREG(initial.st_mode)and not path.is_symlink()and stat.S_IMODE(initial.st_mode)==0o600 and initial.st_uid==os.getuid(),'private-evidence-mode')
 need(initial.st_size<=maximum,'private-evidence-limit');fd=os.open(path,os.O_RDONLY|os.O_NOFOLLOW|os.O_CLOEXEC)
 try:
  before=os.fstat(fd);need((before.st_dev,before.st_ino,before.st_mode,before.st_uid,before.st_size,before.st_mtime_ns)==(initial.st_dev,initial.st_ino,initial.st_mode,initial.st_uid,initial.st_size,initial.st_mtime_ns),'private-evidence-open-identity')
  data=bytearray();digest=hashlib.sha256()
  while True:
   block=os.read(fd,CHUNK)
   if not block:break
   need(len(data)+len(block)<=maximum,'private-evidence-limit');data.extend(block);digest.update(block)
  after=os.fstat(fd);last=path.lstat()
  signature=lambda st:(st.st_dev,st.st_ino,st.st_mode,st.st_uid,st.st_size,st.st_mtime_ns,st.st_ctime_ns)
  need(signature(before)==signature(after)==signature(last)and len(data)==after.st_size,'private-evidence-read-identity')
  need(stat.S_ISREG(after.st_mode)and stat.S_IMODE(after.st_mode)==0o600 and after.st_uid==os.getuid(),'private-evidence-final-mode')
  return bytes(data),{'path':str(path),'bytes':len(data),'sha256':digest.hexdigest(),'device':after.st_dev,'inode':after.st_ino,'mode':'0600','uid':after.st_uid,'mtimeNs':after.st_mtime_ns,'ctimeNs':after.st_ctime_ns}
 finally:os.close(fd)

def bounded_json(path,maximum):
 data=bytearray()
 with Path(path).open('rb')as f:
  while True:
   block=f.read(CHUNK)
   if not block:break
   need(len(data)+len(block)<=maximum,'json-file-limit');data.extend(block)
 return strict_json(bytes(data),maximum)

def file_hash(path,maximum=67108864):
 path=Path(path);before=path.stat();size=0;digest=hashlib.sha256()
 with path.open('rb')as f:
  while True:
   block=f.read(CHUNK)
   if not block:break
   size+=len(block);need(size<=maximum,'file-hash-limit');digest.update(block)
 after=path.stat();need((before.st_dev,before.st_ino,before.st_size,before.st_mtime_ns)==(after.st_dev,after.st_ino,after.st_size,after.st_mtime_ns)and size==after.st_size,'file-hash-changed')
 return {'sha256':digest.hexdigest(),'bytes':size}
