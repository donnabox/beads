from pathlib import Path
import json,hashlib,subprocess,re,copy
W=Path('/Users/dbox/repos/wt/vickie/vickie-transaction-conditional-reads')
D=Path(__file__).resolve().parent
H=Path('/Users/dbox/repos/wt/vickie/bdp-integration-logs/transaction-principal-cumulative-draft/source-author')
FILES=['packages/conformance/test-support/transactional/manifest-v1.json','packages/conformance/test-support/transactional/coverage-v1.json','packages/conformance/src/transactional-cumulative.test.ts','apps/bdptest/src/transactional-matrix.test.ts']
def sha(x):return hashlib.sha256(x if isinstance(x,bytes) else x.encode()).hexdigest()
def normal(x):return json.dumps(x,ensure_ascii=False,indent=2)+'\n'
def compact(x):return json.dumps(x,ensure_ascii=False,separators=(',',':'))
def put(name,x): (D/name).write_text(normal(x))
def git(*args):return subprocess.check_output(['git',*args],cwd=W)
def measure(m):
 pending=[m];nodes=0
 while pending:
  v=pending.pop();nodes+=1
  if isinstance(v,dict):pending.extend(v.values())
  elif isinstance(v,list):pending.extend(v)
 return {'normalBytes':len(normal(m).encode()),'normalSha256':sha(normal(m)),'compactBytes':len(compact(m).encode()),'compactSha256':sha(compact(m)),'nodes':nodes,'actions':sum(len(r.get('actions',[])) for r in m['scenarios']),'assertions':sum(len(a['assertions']) for r in m['scenarios'] for a in r.get('actions',[]))}
def slices(text,key):
 idx=text.index('[',text.index('"'+key+'"'))+1;out=[];decoder=json.JSONDecoder()
 while True:
  while text[idx].isspace() or text[idx]==',':idx+=1
  if text[idx]==']':return out
  obj,end=decoder.raw_decode(text,idx);out.append((obj['id'],text[idx:end]));idx=end
m=json.loads((W/FILES[0]).read_text());c=json.loads((W/FILES[1]).read_text());row=next(r for r in m['scenarios'] if r['id']=='transactional.http.conditional-reads')
assert row['id']=='transactional.http.conditional-reads' and row['requiredProfile']=='transactional'
assert len(row['actions'])==14 and sum(len(a['assertions']) for a in row['actions'])==186
assert all(a['captures']==[] for a in row['actions'])
assert all('exists' in q for a in row['actions'] for q in a['assertions'] if q['kind']=='json-pointer')
assert len(c['rows'])==len(m['scenarios'])==17
assert next(r for r in c['rows'] if r['id']==row['id'])['bindings']==[{'scenario':row['id'],'assertions':[a['id']+':'+q['id'] for a in row['actions'] for q in a['assertions']]}]
preservation=[]
for name,key,f in [('manifest','scenarios',FILES[0]),('coverage','rows',FILES[1])]:
 parent=git('show','HEAD:'+f);assert parent==(D/f'parent-{name}.json').read_bytes()
 text=(W/f).read_text();raws=slices(text,key);assert raws[-1][0]==row['id']
 suffix=',\n    '+raws[-1][1]+'\n  ]\n}\n';assert text.endswith(suffix)
 restored=text[:-len(suffix)]+'\n  ]\n}\n';assert restored.encode()==parent
 old=slices(parent.decode(),key);assert raws[:-1]==old
 preservation.append({'path':f,'parentSha256':sha(parent),'parentBytes':len(parent),'reconstructionSha256':sha(restored),'rawObjects':[{'id':id,'bytes':len(raw.encode()),'sha256':sha(raw)} for id,raw in old]})
assert dict(slices((D/'parent-manifest.json').read_text(),'scenarios'))['transactional.batch.atomic-commit'].count('"equals": 401')==1
old=git('show','HEAD:'+FILES[3]).decode();new=(W/FILES[3]).read_text();hm=json.loads((H/'measurements.json').read_text());mutators=[]
for fn in hm['mutatorPreservation']['functions']:
 st=old.index('function '+fn['name']+'(');en=old.index('\n}',st)+2;block=old[st:en]
 assert block in new and sha(block)==fn['sha256']
 mutators.append({'name':fn['name'],'sha256':sha(block),'byteIdenticalToParentAndPriorMeasurement':True})
# Thirteen inline manifest mutation bodies (twelve methods plus one arrow) remain verbatim.
def brace_end(text,start):
 depth=0;quote=None;escaped=False;i=start
 while i<len(text):
  ch=text[i]
  if quote:
   if escaped:escaped=False
   elif ch=='\\':escaped=True
   elif ch==quote:quote=None
  elif ch in '\"\'`':quote=ch
  elif text[i:i+2]=='//':
   j=text.find('\n',i);i=len(text) if j<0 else j;continue
  elif text[i:i+2]=='/*':i=text.index('*/',i+2)+2;continue
  elif ch=='{':depth+=1
  elif ch=='}':
   depth-=1
   if depth==0:return i+1
  i+=1
 raise AssertionError('unterminated source block')
inline=[]
for match in re.finditer(r'manifest(?:\(value\)|: \(value\) =>) \{',old):
 start=old.index('{',match.start());end=brace_end(old,start);block=old[start:end];assert block in new
 inline.append({'index':len(inline),'sha256':sha(block),'byteIdenticalToParent':True})
assert len(inline)==13,len(inline)
parent=json.loads((D/'parent-manifest.json').read_text());pm=measure(parent)
assert pm['normalSha256']==hm['actual']['normalSha256']
# Each historical specimen is bound by the actual retained measurement, not a guessed scalar maximum.
derivatives=[];(D/'derivatives').mkdir(exist_ok=True)
for r in hm['derivatives']:
 original=(H/'derivatives'/r['name']).read_bytes();assert sha(original)==r['sha256'];v=json.loads(original)
 assert v['scenarios'][0]==parent['scenarios'][0]
 v['scenarios'].append(copy.deepcopy(row));assert v['scenarios'][:-1]==json.loads(original)['scenarios']
 data=normal(v);(D/'derivatives'/r['name']).write_text(data)
 derivatives.append({'name':r['name'],'historicalPath':str(H/'derivatives'/r['name']),'historicalSha256':sha(original),'sha256':sha(data),'oldTransformUnchanged':True,**measure(v)})
v=copy.deepcopy(m);lines=next(a for a in next(r for r in v['scenarios'] if r['id']==row['id'])['actions'] if a['id']=='conditional-get')['raw']['headerLines'];matches=[h for h in lines if h['name']=='if-none-match'];assert len(matches)==1 and matches[0]['value']=='*';matches[0]['value']='"conditional-no-match"';data=normal(v);(D/'derivatives/conditional-nonmatching-if-none-match.json').write_text(data);derivatives.append({'name':'conditional-nonmatching-if-none-match.json','sha256':sha(data),'soleMutation':'transactional.http.conditional-reads/conditional-get raw if-none-match * -> quoted conditional-no-match','expectedFailureIds':['conditional-get:status','conditional-get:body-absent'],'expectedHttp':4,'diagnosticMaximumHttp':14,**measure(v)});matches[0]['value']='*';assert v==m
actual=measure(m);actual.update({'diskBytes':(W/FILES[0]).stat().st_size,'diskSha256':sha((W/FILES[0]).read_bytes()),'coverageBytes':(W/FILES[1]).stat().st_size,'coverageNormalBytes':len(normal(c).encode()),'coverageSha256':sha((W/FILES[1]).read_bytes()),'newAssertions':186})
growth={k:actual[k]-hm['actual'][k] for k in ['diskBytes','normalBytes','nodes','coverageBytes','assertions']}
maxbytes=max(x['normalBytes'] for x in derivatives);maxnodes=max(x['nodes'] for x in derivatives)
proposed={'canonicalDisk':actual['diskBytes']+4096,'canonicalDiskDeltaFrom488280':actual['diskBytes']+4096-488280,'canonicalNormal':actual['normalBytes']+4096,'canonicalNodes':actual['nodes']+64,'canonicalNodesDeltaFrom13019':actual['nodes']+64-13019,'coverageDiskAndNormal':max(actual['coverageBytes'],actual['coverageNormalBytes'])+1024,'coverageDeltaFrom70662':max(actual['coverageBytes'],actual['coverageNormalBytes'])+1024-70662,'derivativeNormal':maxbytes+4096,'derivativeNodes':maxnodes+64}
measurements={'status':'DATA_ONLY_UNEXECUTED','base':git('rev-parse','HEAD').decode().strip(),'actual':actual,'growth':growth,'preservation':preservation,'mutatorPreservation':{'functions':mutators,'inline':inline,'allUnchanged':True},'derivatives':derivatives,'maxDerivativeNormalBytes':maxbytes,'maxDerivativeNodes':maxnodes,'proposedGuards':proposed,'guardsAmended':'proposed constants pending independent root acceptance','productExecuted':False,'historicalMeasurement':{'path':str(H/'measurements.json'),'sha256':sha((H/'measurements.json').read_bytes())}}
put('measurements.json',measurements)
patch=git('diff','--binary','HEAD');(D/'draft.patch').write_bytes(patch)
status=git('status','--porcelain').decode();actualFiles=set(git('diff','--name-only','HEAD').decode().splitlines());assert actualFiles==set(FILES)
binding={'workspace':str(W),'branch':git('branch','--show-current').decode().strip(),'head':git('rev-parse','HEAD').decode().strip(),'tree':git('rev-parse','HEAD^{tree}').decode().strip(),'files':{f:{'sha256':sha((W/f).read_bytes()),'new':False} for f in FILES},'diffSha256':sha(patch),'status':status,'scope':'four-file TX rebind successor; no product execution; measured fences pending root acceptance'};put('source-binding.json',binding)
put('preservation.json',preservation)
print(normal({'actual':actual,'growth':growth,'derivatives':len(derivatives),'maxDerivativeNormalBytes':maxbytes,'maxDerivativeNodes':maxnodes,'proposedGuards':proposed,'bindingSha256':sha((D/'source-binding.json').read_bytes()),'patchSha256':sha(patch)}))
