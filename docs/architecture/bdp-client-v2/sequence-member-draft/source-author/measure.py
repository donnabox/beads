from pathlib import Path
import json,hashlib,copy,re,subprocess
W=Path('/Users/dbox/repos/wt/vickie/vickie-sequence-member-evidence')
D=Path(__file__).parent
L=D.parents[1]
H=L/'transaction-command-minimum-evidence-draft/event-id-repair'
sha=lambda b:hashlib.sha256(b).hexdigest()
normal=lambda v:(json.dumps(v,ensure_ascii=False,indent=2)+'\n').encode()
compact=lambda v:json.dumps(v,ensure_ascii=False,separators=(',',':')).encode()
def nodes(v):return 1+sum(nodes(x) for x in (v.values() if isinstance(v,dict) else v if isinstance(v,list) else []))
def metrics(v):return dict(normalBytes=len(normal(v)),normalSha256=sha(normal(v)),compactBytes=len(compact(v)),compactSha256=sha(compact(v)),nodes=nodes(v),actions=sum(len(r['actions']) for r in v['scenarios']),assertions=sum(len(a['assertions']) for r in v['scenarios'] for a in r['actions']))
def rawobjects(text,key):
 dec=json.JSONDecoder();i=text.index('[',text.index('"'+key+'"'))+1;out={}
 while True:
  while text[i].isspace() or text[i]==',':i+=1
  if text[i]==']':break
  value,end=dec.raw_decode(text,i);out[value['id']]=text[i:end];i=end
 return out
paths=['packages/conformance/test-support/transactional/manifest-v1.json','packages/conformance/test-support/transactional/coverage-v1.json','packages/conformance/src/transactional-cumulative.test.ts','apps/bdptest/src/transactional-matrix.test.ts']
current=json.loads((W/paths[0]).read_bytes());parent=json.loads((D/'parent'/paths[0]).read_bytes());new=current['scenarios'][-1]
assert new['id']=='transactional.sequence.member-transactions'
assert current['scenarios'][:-1]==parent['scenarios']
preservation=[]
for path,key in zip(paths[:2],['scenarios','rows']):
 before=(D/'parent'/path).read_text();after=(W/path).read_text();oldraw=rawobjects(before,key);raw=rawobjects(after,key)
 assert len(oldraw)==14 and len(raw)==15
 for id,body in oldraw.items():assert raw[id]==body,id
 tail=',\n    '+raw[new['id']]+'\n  ]\n}\n';assert after.endswith(tail)
 reconstructed=after[:-len(tail)]+'\n  ]\n}\n';assert reconstructed==before
 preservation.append(dict(path=path,parentSha256=sha(before.encode()),reconstructedSha256=sha(reconstructed.encode()),oldRawObjects=[dict(id=id,sha256=sha(body.encode()),bytes=len(body.encode())) for id,body in oldraw.items()]))
# Existing transformations are unchanged, independently checked against parent source.
before=(D/'parent'/paths[3]).read_text();after=(W/paths[3]).read_text();functions=[]
for name in ['applyDeletionControl','applyCrossCarrierControl','applyCommandMinimumControl','crossCarrierRequests','minimumRequests','crossCarrierWire']:
 pattern=r'^function '+name+r'\([\s\S]*?^}'
 a=re.search(pattern,before,re.M).group();b=re.search(pattern,after,re.M).group();assert a==b,name
 functions.append(dict(name=name,sha256=sha(a.encode()),bytes=len(a.encode()),unchanged=True))
pattern=r'^(?P<i> +)manifest\(value\) \{[\s\S]*?^(?P=i)\},?$'
a=[m.group() for m in re.finditer(pattern,before,re.M)];b=[m.group() for m in re.finditer(pattern,after,re.M)];assert a==b
mutators=dict(functions=functions,inlineMutators=[dict(index=i,sha256=sha(s.encode()),bytes=len(s.encode())) for i,s in enumerate(a)],allUnchanged=True)
retention=L/'transaction-command-minimum-evidence-draft/event-final/root-growth-retention.json';inventory=json.loads(retention.read_bytes())['derivatives'];assert len(inventory)==40
out=D/'derivatives';out.mkdir(exist_ok=True);records=[]
def emit(name,value,origin):
 data=normal(value);(out/name).write_bytes(data);records.append(dict(name=name,**metrics(value),**origin))
for entry in inventory:
 path=H/'derivative-measurements'/entry['name'];data=path.read_bytes();assert sha(data)==entry['sha256'],path
 value=json.loads(data);assert len(value['scenarios'])==14
 # Current parent canonical is byte-identified by historical measurement sourceHashes.
 value['scenarios'].append(copy.deepcopy(new))
 emit(entry['name'],value,dict(historicalPath=str(path),historicalSha256=sha(data),historicalNormalBytes=entry['normalBytes'],historicalNodes=entry['nodes'],originalTransformPreserved=True))
negative=copy.deepcopy(current);act=negative['scenarios'][-1]['actions'][0];body=json.loads(act['raw']['body']['value']);body['operations'][1]['bead']='beads/a';act['raw']['body']['value']=compact(body).decode()
assert len(act['raw']['body']['value'].encode())==408
emit('sequence-member-successful-delete.json',negative,dict(newSelected=True,changedAction='s-sequence',body=act['raw']['body']['value'],bodyBytes=408,bodySha256=sha(act['raw']['body']['value'].encode()),assertionsUnchanged=True))
assert len(records)==41
base=metrics(parent);now=metrics(current);coverage=(W/paths[1]).read_bytes();basecoverage=(D/'parent'/paths[1]).read_bytes()
actual=dict(**now,diskBytes=(W/paths[0]).stat().st_size,coverageBytes=len(coverage),newDiskBytes=(W/paths[0]).stat().st_size-(D/'parent'/paths[0]).stat().st_size,newNormalBytes=now['normalBytes']-base['normalBytes'],newNodes=now['nodes']-base['nodes'],newCoverageBytes=len(coverage)-len(basecoverage),newAssertions=now['assertions']-base['assertions'])
assert actual['newDiskBytes']<=80000 and actual['newNormalBytes']<=80000 and actual['newNodes']<=2200 and actual['newCoverageBytes']<=15000 and actual['newAssertions']<=260
maxb=max(r['normalBytes'] for r in records);maxn=max(r['nodes'] for r in records);assert maxb<=836548 and maxn<=21926
result=dict(status='DRAFT_UNEXECUTED_STANDARD_LIBRARY_MEASUREMENT',base=base,actual=actual,rawPreservation=preservation,retentionAuthority=dict(path=str(retention),sha256=sha(retention.read_bytes())),mutationSourcePreservation=mutators,derivatives=records,maxDerivativeNormalBytes=maxb,maxDerivativeNodes=maxn,maxDerivativeNames=[r['name'] for r in records if r['normalBytes']==maxb],mutatedInvocations=41,unchangedInvocations=14,totalInvocations=55,expandedAppCases=62,proposedFencesOnly=dict(manifestDisk=actual['diskBytes']+4096,manifestNormal=now['normalBytes']+4096,manifestDiskDelta=actual['diskBytes']+4096-488280,manifestNodes=now['nodes']+64,manifestNodeDelta=now['nodes']+64-13019,coverageBytes=len(coverage)+1024,coverageDelta=len(coverage)+1024-70662,derivativeNormal=maxb+4096,derivativeNodes=maxn+64),fencesEdited=False,executed=False)
(D/'measurements.json').write_bytes(normal(result));print(json.dumps({k:v for k,v in result.items() if k in ['actual','maxDerivativeNormalBytes','maxDerivativeNodes','maxDerivativeNames','proposedFencesOnly']},indent=2))
