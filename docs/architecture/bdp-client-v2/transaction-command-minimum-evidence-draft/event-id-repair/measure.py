"""Static source/JSON arithmetic only. Does not import or execute product code."""
from pathlib import Path
import hashlib,json,re,subprocess
W=Path('/Users/dbox/repos/wt/vickie/vickie-command-minimum-evidence')
D=Path(__file__).parent
ID='read-update.http.command-minimum-fields'
S='https://reference.invalid/acme/'
PATHS=['packages/conformance/test-support/transactional/manifest-v1.json','packages/conformance/test-support/transactional/coverage-v1.json','packages/conformance/src/transactional-cumulative.test.ts','apps/bdptest/src/transactional-matrix.test.ts','packages/conformance/src/profile-contract-plan.test.ts']
h=lambda b:hashlib.sha256(b).hexdigest()
def write(name,v): (D/name).write_text(json.dumps(v,indent=2,ensure_ascii=False)+'\n')
def normal(v):return (json.dumps(v,indent=2,ensure_ascii=False)+'\n').encode()
def compact(v,sort=False):return json.dumps(v,ensure_ascii=False,separators=(',',':'),sort_keys=sort).encode()
def nodes(v):return 1+sum(nodes(x) for x in (v.values() if isinstance(v,dict) else v if isinstance(v,list) else []))
def slices(s,key):
 pos=s.index('[',s.index('"'+key+'"'))+1;result={};decoder=json.JSONDecoder()
 while True:
  while s[pos].isspace() or s[pos]==',':pos+=1
  if s[pos]==']':return result
  value,end=decoder.raw_decode(s,pos);result[value['id']]=s[pos:end];pos=end
manifest=json.loads((W/PATHS[0]).read_text());coverage=json.loads((W/PATHS[1]).read_text());fixturepath='packages/conformance/test-support/transactional/fixture-v1.json';fb=(W/fixturepath).read_bytes();fixture=json.loads(fb)
proof={};measure={}
for name,key,p in [('manifest','scenarios',PATHS[0]),('coverage','rows',PATHS[1])]:
 b=(W/p).read_bytes();text=b.decode();j=json.loads(text);old=(D/f'parent-{name}.json').read_text();prior=slices(old,key);current=slices(text,key)
 reconstructed=text.replace('    '+current[ID]+',\n','')
 assert reconstructed==old
 assert len(prior)==13 and len(current)==14
 assert all(current[k]==v for k,v in prior.items())
 proof[name]=dict(parentSha256=h(old.encode()),reconstructedSha256=h(reconstructed.encode()),equal=True,objects=[dict(id=k,bytes=len(v.encode()),sha256=h(v.encode()),unchanged=current[k]==v) for k,v in prior.items()])
 before=json.loads(old);measure[name]=dict(diskBytes=len(b),normalBytes=len(normal(j)),compactBytes=len(compact(j)),canonicalBytes=len(compact(j,True)),nodes=nodes(j),sha256=h(b),normalSha256=h(normal(j)),compactSha256=h(compact(j)),canonicalSha256=h(compact(j,True)),delta=dict(diskBytes=len(b)-len(old.encode()),normalBytes=len(normal(j))-len(normal(before)),nodes=nodes(j)-nodes(before)))
row=next(r for r in manifest['scenarios'] if r['id']==ID);cov=next(r for r in coverage['rows'] if r['id']==ID)
mirror=[a['id']+':'+x['id'] for a in row['actions'] for x in a['assertions']];assert cov['bindings']==[dict(scenario=ID,assertions=mirror)]
assert len(row['actions'])==11 and not cov['wholeContract']
measure.update(newActions=11,totalActions=sum(len(r['actions']) for r in manifest['scenarios']),newAssertions=len(mirror),totalAssertions=sum(len(a['assertions']) for r in manifest['scenarios'] for a in r['actions']),newCaptures=sum(len(a['captures']) for a in row['actions']),newRowCompactSha256=h(compact(row)),fixture=dict(path=fixturepath,bytes=len(fb),sha256=h(fb)),planningLimits=dict(manifestAdditionalBytes=122880,nodesAdditional=3000,coverageAdditionalBytes=24576,newAssertions=260),executed=False)
assert measure['manifest']['delta']['diskBytes']<=122880 and measure['manifest']['delta']['normalBytes']<=122880 and measure['manifest']['delta']['nodes']<=3000
assert measure['coverage']['delta']['diskBytes']<=24576 and measure['newAssertions']<=260
# Preserve historical bases and margins; proposals only, no source-allowance edits.
m=measure['manifest'];c=measure['coverage'];measure['proposedCanonicalFences']=dict(manifestDisk=dict(oldAbsolute=688302,newAbsolute=m['diskBytes']+4096,oldDeltaFrom488280=200022,newDeltaFrom488280=m['diskBytes']+4096-488280,margin=4096),manifestNormal=dict(oldAbsolute=699694,newAbsolute=m['normalBytes']+4096,margin=4096),nodes=dict(oldAbsolute=18171,newAbsolute=m['nodes']+64,oldDeltaFrom13019=5152,newDeltaFrom13019=m['nodes']+64-13019,margin=64),coverage=dict(oldAbsolute=100269,newAbsolute=c['diskBytes']+1024,oldDeltaFrom70662=29607,newDeltaFrom70662=c['diskBytes']+1024-70662,margin=1024))
oldRoutes=[('GET','beads/a'),('POST','operations/batch'),('GET','links/l'),('GET','beads/?limit=32'),('GET','events/'),('GET','beads/new'),('GET','changes/?after=genesis'),('GET','links/?limit=32'),('GET','beads/b'),('GET','beads/c'),('POST','operations/update-where'),('POST','operations/delete-where'),('POST','operations/create-bead'),('POST','operations/sequence'),('POST','operations/update-bead')]
newRoutes=[('POST','operations/create-bead'),('POST','operations/batch'),('POST','operations/sequence'),('POST','operations/update-bead'),('GET','beads/?limit=32'),('GET','events/'),('GET','beads/a'),('GET','links/l'),('GET','beads/new'),('GET','changes/?after=genesis'),('GET','links/?limit=32'),('GET','beads/b'),('GET','beads/c'),('POST','operations/update-where'),('POST','operations/delete-where')]
assert set(oldRoutes)==set(newRoutes) and len(set(newRoutes))==15
actual=[]
from urllib.parse import urljoin,urlencode
for r in manifest['scenarios']:
 for a in r['actions']:
  if a['family']!='http':continue
  base=S if a['target']['binding']=='scope' else fixture['bindings'][a['target']['binding']];url=urljoin(base,a['target'].get('path',''))
  if a['target'].get('query'):url+='?'+urlencode(a['target']['query'])
  item=dict(method=a['method'],url=url)
  if item not in actual:actual.append(item)
assert actual==[dict(method=method,url=S+route) for method,route in newRoutes]
configs=[]
for label,routes,expected in [('old',oldRoutes,'a9a17908b6fa87674d5f573b584f4a1ce1d41d7909fadf48b65e9b05397ddfd4'),('new',newRoutes,'02dd7d9f35e7482c07ce1550a6572232cc48ba6858cd66f8dfc8f5384f3628be')]:
 pre=dict(maximumRequestBodyBytes=65536,maximumRequestHeaderBytes=16384,profile='transactional',routes=[dict(method=method,url=S+route) for method,route in routes],scope=S);b=compact(pre,True);assert h(b)==expected
 configs.append(dict(label=label,preimage=b.decode(),bytes=len(b),sha256=h(b)))
write('route-bindings.json',dict(routeSetUnchanged=True,actualOrderedRoutes=actual,preimages=configs,fixtureUnchanged=True,permissionAndRuntimeCapsUnchanged=True))
requests=[]
for a in row['actions']:
 raw=a.get('raw',{});body=raw.get('body',{}).get('value');b=None if body is None else body.encode()
 requests.append(dict(id=a['id'],family=a['family'],method=a.get('method',a.get('input',{}).get('method')),target=a.get('target',a.get('input')),credential=a.get('credentialRef','private-helper-existing'),headers=raw.get('headerLines','private helper empty authored headers'),body=body,bodyBytes=0 if b is None else len(b),bodySha256=None if b is None else h(b),assertionIds=[x['id'] for x in a['assertions']],captures=a['captures']))
write('requests.json',dict(scope=S,actions=requests,postBodyBytes=sum(x['bodyBytes'] for x in requests),maximumBodyBytes=max(x['bodyBytes'] for x in requests),outerRawObservations=10,dynamicReceiptObservations=1,bootstrapHTTP=0,executed=False))
write('measurements.json',measure);write('preservation.json',proof)
base=subprocess.check_output(['git','rev-parse','HEAD'],cwd=W,text=True).strip();assert base=='453b6dab386746bb03aac2c0a293572b8a71c969'
changed=subprocess.check_output(['git','diff','--name-only'],cwd=W,text=True).splitlines();assert set(changed)<=set(PATHS)
write('source-binding.json',dict(stage='UNEXECUTED_DRAFT_MEASUREMENTS',base=base,baseTree=subprocess.check_output(['git','rev-parse','HEAD^{tree}'],cwd=W,text=True).strip(),worktree=str(W),branch=subprocess.check_output(['git','branch','--show-current'],cwd=W,text=True).strip(),files=[dict(path=p,bytes=(W/p).stat().st_size,sha256=h((W/p).read_bytes())) for p in PATHS],changedPaths=changed,allowancesAmended=False,executed=False))
(D/'draft.patch').write_bytes(subprocess.check_output(['git','diff','--binary'],cwd=W))
print(json.dumps(dict(manifest=measure['manifest'],coverage=measure['coverage'],newAssertions=len(mirror),totalAssertions=measure['totalAssertions'],sourceBinding=h((D/'source-binding.json').read_bytes()),diff=h((D/'draft.patch').read_bytes())),indent=2))
