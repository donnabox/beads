"""Data-only derivative reconstruction; no product imports or executable test code."""
from pathlib import Path
import copy, hashlib, json, subprocess
D=Path(__file__).parent
H=Path('/Users/dbox/repos/wt/vickie/bdp-integration-logs/transaction-cross-carrier-draft')
W=Path('/Users/dbox/repos/wt/vickie/vickie-command-minimum-evidence')
sha=lambda b:hashlib.sha256(b).hexdigest()
normal=lambda v:(json.dumps(v,ensure_ascii=False,indent=2)+'\n').encode()
compact=lambda v:json.dumps(v,ensure_ascii=False,separators=(',',':')).encode()
def nodes(v):return 1+sum(nodes(x) for x in (v.values() if isinstance(v,dict) else v if isinstance(v,list) else []))
def metrics(v):
 n=normal(v);c=compact(v)
 return dict(normalBytes=len(n),normalSHA256=sha(n),compactBytes=len(c),compactSHA256=sha(c),nodes=nodes(v),actions=sum(len(r['actions']) for r in v['scenarios']),assertions=sum(len(a['assertions']) for r in v['scenarios'] for a in r['actions']))
def changes(a,b,path=''):
 if type(a)!=type(b):return [dict(pointer=path,before=a,after=b)]
 if isinstance(a,dict):
  out=[]
  for k in dict.fromkeys([*a,*b]):
   p=path+'/'+k.replace('~','~0').replace('/','~1')
   if k not in a:out.append(dict(pointer=p,after=b[k],added=True))
   elif k not in b:out.append(dict(pointer=p,before=a[k],removed=True))
   else:out+=changes(a[k],b[k],p)
  return out
 if isinstance(a,list):
  if len(a)!=len(b):return [dict(pointer=path,before=a,after=b)]
  return [d for i,(x,y) in enumerate(zip(a,b)) for d in changes(x,y,path+'/'+str(i))]
 return [] if a==b else [dict(pointer=path,before=a,after=b)]
paths=['packages/conformance/test-support/transactional/manifest-v1.json','packages/conformance/test-support/transactional/coverage-v1.json','packages/conformance/src/transactional-cumulative.test.ts','apps/bdptest/src/transactional-matrix.test.ts','packages/conformance/src/profile-contract-plan.test.ts']
before={p:sha((W/p).read_bytes()) for p in paths}
assert before[paths[3]]=='e319109ab0e9f7258e6918ae789ed7a5672eb7aabd0ec8725ae38e3634164e68'
parent=json.loads((D/'parent-manifest.json').read_bytes());current=json.loads((W/paths[0]).read_bytes())
row=current['scenarios'][0];assert row['id']=='read-update.http.command-minimum-fields'
without=copy.deepcopy(current);without['scenarios'].pop(0);assert without==parent
historical=json.loads((H/'measurements-v2.json').read_bytes())
assert sha((D/'parent-manifest.json').read_bytes())==historical['canonical']['diskSHA256']
assert metrics(parent)['normalSHA256']==historical['canonical']['normalSHA256']
assert len(historical['derivatives'])==38
out=D/'derivative-measurements';out.mkdir(exist_ok=True)
allrows=[]
def emit(name,v,extra):
 m=metrics(v);m.update(name=name,deltaNormalBytes=m['normalBytes']-metrics(current)['normalBytes'],deltaNodes=m['nodes']-nodes(current),**extra)
 m['changes']=changes(current,v)
 m['changedBodies']=[dict(scenario=r['id'],action=a['id'],utf8Bytes=len(a['raw']['body']['value'].encode()),sha256=sha(a['raw']['body']['value'].encode()),literal=a['raw']['body']['value']) for r in v['scenarios'] for a in r['actions'] if a.get('raw',{}).get('body') and a['raw']['body']!=next((x.get('raw',{}).get('body') for rr in current['scenarios'] if rr['id']==r['id'] for x in rr['actions'] if x['id']==a['id']),None)]
 (out/(name+'.json')).write_bytes(normal(v));allrows.append(m)
for entry in historical['derivatives']:
 name=entry['name'];p=H/('derivative-'+name+'.json');old=json.loads(p.read_bytes())
 for k,v in metrics(old).items():assert entry[k]==v,(name,k)
 changed=copy.deepcopy(old);changed['scenarios'].insert(0,copy.deepcopy(row))
 emit(name,changed,dict(category='existing-manifest-derivative',historicalArtifact=str(p),historicalSHA256=sha(p.read_bytes()),historicalNormalBytes=entry['normalBytes'],historicalNodes=entry['nodes'],oldRowsTransformationPreserved=True,prelaunchRejected=name=='duplicate-receipt-binding'))
for name in ['minimum-batch-body-conflict','minimum-readonly-credential']:
 changed=copy.deepcopy(current);r=changed['scenarios'][0]
 if name=='minimum-batch-body-conflict':
  a=next(a for a in r['actions'] if a['id']=='m-replay-b');body=json.loads(a['raw']['body']['value']);body['operations'][0]['properties']['title']='minimum-conflict';a['raw']['body']['value']=compact(body).decode()
 else:next(a for a in r['actions'] if a['id']=='m-create-a')['credentialRef']='reader'
 emit(name,changed,dict(category='new-selected-manifest-derivative',assertionsUnchanged=True))
assert len(allrows)==40
maxb=max(r['normalBytes'] for r in allrows);maxn=max(r['nodes'] for r in allrows)
result=dict(status='UNEXECUTED_DATA_ONLY_MEASUREMENT',sourceHashes=before,historicalMeasurementsSHA256=sha((H/'measurements-v2.json').read_bytes()),canonical=metrics(current),existingMutatedInvocations=38,existingUnchangedInvocations=14,newSelectedMutatedInvocations=2,totalObserveInvocations=54,derivatives=allrows,maxNormalBytes=maxb,maxNodes=maxn,maxByteNames=[r['name'] for r in allrows if r['normalBytes']==maxb],maxNodeNames=[r['name'] for r in allrows if r['nodes']==maxn],proposedBothAppDerivativeFences=dict(oldNormalBytes=705812,newNormalBytes=maxb+4096,byteMargin=4096,oldNodes=18327,newNodes=maxn+64,nodeMargin=64),runtimeLimitsUnchanged=dict(manifestBytes=1048576,nodes=100000),executed=False)
(D/'derivative-measurements.json').write_bytes(normal(result))
assert before=={p:sha((W/p).read_bytes()) for p in paths}
print(json.dumps({k:v for k,v in result.items() if k not in ['sourceHashes','derivatives']},indent=2))
