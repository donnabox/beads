import json,hashlib,copy,subprocess
from pathlib import Path
W=Path('/Users/dbox/repos/wt/vickie/vickie-cross-carrier-draft');L=Path('/Users/dbox/repos/wt/vickie/bdp-integration-logs/transaction-cross-carrier-draft');S='https://reference.invalid/acme/'
js=lambda x:json.dumps(x,separators=(',',':'),ensure_ascii=False)
normal=lambda x:json.dumps(x,indent=2,ensure_ascii=False)+'\n'
sha=lambda b:hashlib.sha256(b if isinstance(b,bytes) else b.encode()).hexdigest()
def nodes(v):return 1+sum(nodes(x) for x in (v.values() if isinstance(v,dict) else v if isinstance(v,list) else []))
def props(v):
 n=normal(v);c=js(v)
 return dict(normalBytes=len(n.encode()),normalSHA256=sha(n),compactBytes=len(c.encode()),compactSHA256=sha(c),nodes=nodes(v),actions=sum(len(r['actions']) for r in v['scenarios']),assertions=sum(len(a['assertions']) for r in v['scenarios'] for a in r['actions']))
paths={n:W/f'packages/conformance/test-support/transactional/{n}-v1.json' for n in ['manifest','coverage','fixture']}
m=json.loads(paths['manifest'].read_bytes());c=json.loads(paths['coverage'].read_bytes());rowid='transactional.idempotency.cross-carrier';row=next(r for r in m['scenarios'] if r['id']==rowid)
def rawobjects(text,key):
 start=text.index('[',text.index('"'+key+'"'))+1;out={};d=json.JSONDecoder()
 while True:
  while text[start] in ' \n\r\t,':start+=1
  if text[start]==']':return out
  v,end=d.raw_decode(text,start);out[v['id']]=text[start:end];start=end
preservation=[]
for n,k in [('manifest','scenarios'),('coverage','rows')]:
 old=(L/f'parent-{n}.json').read_text();new=paths[n].read_text();a=rawobjects(old,k);b=rawobjects(new,k)
 assert len(a)==12 and len(b)==13
 assert all(b[id]==raw for id,raw in a.items())
 reconstructed=new.replace('    '+b[rowid]+',\n','');assert reconstructed==old
 preservation.append(dict(artifact=n,rawObjects=12,reconstructedSHA256=sha(reconstructed),parentSHA256=sha(old)))
assert c['rows'][next(i for i,r in enumerate(m['scenarios']) if r['id']==rowid)]['bindings']==[dict(scenario=rowid,assertions=[a['id']+':'+q['id'] for a in row['actions'] for q in a['assertions']])]
assert sum(len(a['assertions']) for a in row['actions'])==230 and len(row['actions'])==12
assert len(paths['manifest'].read_bytes())<=720094 and len(normal(m).encode())<=730676 and nodes(m)<=18973
assert len(paths['coverage'].read_bytes())<=106043 and len(normal(c).encode())<=106043
assert sum(len(a['assertions']) for r in m['scenarios'] for a in r['actions'])<=2172
assert sum(len(r['actions']) for r in m['scenarios'])==199
schema=json.loads((W/'schemas/bdp-v0.schema.json').read_bytes());refs=sorted({a['schema'] for r in m['scenarios'] for x in r['actions'] for a in x['assertions'] if a['kind']=='json-schema'})
for ref in refs:
 assert ref.startswith('#/$defs/') and ref[8:] in schema['$defs'],ref
# All AST/data transformations below independently reproduce authored derivatives.
# They do not execute app code, runtime code, tests, loaders or services.
D=[]
def at(v,rid,aid):return next(a for r in v['scenarios'] if r['id']=='transactional.'+rid for a in r['actions'] if a['id']==aid)
def emit(name,mutate,description=None):
 v=copy.deepcopy(m);mutate(v);p=props(v);p.update(name=name,deltaBytes=p['normalBytes']-len(normal(m).encode()),deltaNodes=p['nodes']-nodes(m))
 if description:p.update(description)
 assert p['normalBytes']<=736794 and p['nodes']<=19129,(name,p)
 D.append(p);(L/('derivative-'+name+'.json')).write_text(normal(v))
def changedassert(v,rid,aid,qid,value,unset=False):
 q=next(q for q in at(v,rid,aid)['assertions'] if q['id']==qid)
 if unset:q.pop('equalsBinding',None)
 q['equals']=value
for k in ['binding','method','credential','body-assertion']:
 def change(v,k=k):
  a=next(a for a in v['scenarios'][0]['actions'] if a.get('operation')=='transactional.receipt-read')
  if k=='body-assertion':a['assertions']=[dict(id='deliberate-counterfactual',kind='json-pointer',pointer='/body/results/0/resource/properties/title',exists=True,equals='never-observed')]
  else:a['input']={'binding':'subject','method':'GET'} if k=='binding' else {'binding':'commit-receipt','method':'POST'} if k=='method' else {'binding':'commit-receipt','method':'GET','credentialRef':'reader'}
 emit('receipt-'+k,change)
def ordinary(v,append=False):
 r=v['scenarios'][0];at(v,'batch.atomic-commit','a-point')['captures']=[{'binding':'ordinary-resource-id','from':{'kind':'json-pointer','pointer':'/id'}}]
 if append:r['actions'].append(dict(id='ordinary-id-as-receipt',family='client',operation='transactional.receipt-read',input={'binding':'ordinary-resource-id','method':'GET'},captures=[],assertions=[dict(id='status',kind='json-pointer',pointer='/status',exists=True,equals=200)]))
emit('ordinary-capture',ordinary);emit('forbidden-receipt-use',lambda v:ordinary(v,True))
for k,aid,qid,val in [('matched','set-update','matched',{'operationIndex':0,'outcome':'matched','count':3}),('order','set-update','a-id',S+'beads/b'),('index','set-update','b-index',1),('poststate','target-a-updated','properties',{'title':'never-observed'})]:emit('set-'+k,lambda v,aid=aid,qid=qid,val=val:changedassert(v,'set.singleton-targets',aid,qid,val))
def crosswire(v):
 a=at(v,'set.singleton-targets','set-delete-facts');a['input']={'binding':'set-update-receipt','method':'GET'};a['assertions']=[dict(id='expected-delete-count',kind='json-pointer',pointer='/body/results/0/count',exists=True,equals=1)]
emit('cross-wired-receipt',crosswire)
for name,rid,aid,qid,val in [
 ('changed-poststate','set.mutation','m-updated-b','properties',{'title':'never-observed'}),('rollback-metadata','set.mutation','m-rollback-a','metadata',{'tag':'never-observed'}),('receipt-order','set.mutation','m-update','a-id',S+'beads/b'),('event-order','set.mutation','m-update-events','ordered-facts',[['created',S+'beads/b',0],['created',S+'beads/c',1],['created',S+'beads/a',2],['updated',S+'beads/b',0],['updated',S+'beads/a',1]]),('zero-results','set.zero-match','z-update','exact-results',[{'operationIndex':0,'outcome':'matched','count':1}]),('zero-events','set.zero-match','z-update-events','ordered-facts',[['created',S+'beads/a',0],['updated',S+'beads/a',0]])]:emit('effects-'+name,lambda v,rid=rid,aid=aid,qid=qid,val=val:changedassert(v,rid,aid,qid,val))
def no_effect(v):next(q for q in at(v,'set.zero-match','z-update')['assertions'] if q['id']=='no-effect')['exists']=True
emit('effects-no-effect',no_effect)
for name,aid,qid,val in [('root-next','p-root-read','next-page',None),('inline-boundary','p-update','p30-id',S+'beads/p31'),('page-member','p-page-read','p31-id',S+'beads/p30'),('page-parent','p-page-read','parent',S+'receipts/wrong'),('historical-title','p-page-retained','historical-title','later'),('terminal-next','p-page-read','no-next',S+'receipts/wrong/pages/next')]:emit('pagination-'+name,lambda v,aid=aid,qid=qid,val=val:changedassert(v,'receipt.pagination',aid,qid,val,True))
for name,aid,qid,val in [('code','c-conflict','problem-code','malformed-request'),('reopened-status','c-conflict-reopened','http-status',200),('current-record','c-state','ordered-state',[[S+'beads/b',S+'types/task',{'title':'original'},{'tag':'conflict'}]]),('retained-body','c-retry-reopened','retained-body',{}),('event','c-events-reopened','ordered-facts',[[S+'events/','created',S+'beads/b',0]])]:emit('conflict-'+name,lambda v,aid=aid,qid=qid,val=val:changedassert(v,'idempotency.conflict',aid,qid,val,True))
for k in ['bead-property','link-metadata','group-inventory-and-head']:
 def noop(v,k=k):
  if k=='group-inventory-and-head':
   changedassert(v,'batch.no-op-entries','n-noop-groups','group-inventory',[['pos-1'],['pos-2']]);changedassert(v,'batch.no-op-entries','n-noop-groups','observed-head','pos-2')
  else:
   a=at(v,'batch.no-op-entries','n-noop');b=json.loads(a['raw']['body']['value'])
   if k=='bead-property':b['operations'][0]['propertiesChange'][0]['value']='different'
   else:b['operations'][1]['metadataChange'][0]['value']={'tag':'different'}
   a['raw']['body']['value']=js(b)
 emit('noop-'+k,noop)
def pointer(id,p,e):return dict(id='control-'+id,kind='json-pointer',pointer=p,exists=True,equals=e)
for k in ['incoming-order','outgoing-order','committed-prefix-state']:
 def deletion(v,k=k):
  direction='outgoing' if k=='outgoing-order' else 'incoming';prefix=direction+'-survives-prefix';links=['r','l'] if direction=='outgoing' else ['l','r'];deleted=links+([] if k=='committed-prefix-state' else ['a']);a=at(v,'batch.deletion-safety','d-'+direction+'-survives')
  ops=[dict(operation='updateBead',bead='beads/b',propertiesChange=[dict(op='replace',path='/title',value=prefix)],metadataChange=[dict(op='replace',path='',value={'phase':prefix})])]+[dict(operation='deleteLink',link='links/'+id) for id in links]+([] if k=='committed-prefix-state' else [dict(operation='deleteBead',bead='beads/a')]);a['raw']['body']={'encoding':'utf8','value':js({'operations':ops})}
  positive=[dict(id='control-position',kind='header',name='bdp-scope-position',equals='pos-2'),pointer('completed','/status','completed'),pointer('detail','/detail','available'),pointer('required','/requiredPosition','pos-2'),pointer('effect','/effectPosition','pos-2'),pointer('inline','/next',None),dict(id='control-ordered-results',kind='json-array-tuples',pointer='/results',ordered=True,projections=[{'pointer':'/operationIndex'},{'pointer':'/outcome'}],equals=[[0,'updated']]+[[i+1,'deleted'] for i,id in enumerate(deleted)]),pointer('updated-id','/results/0/resource/id',S+'beads/b'),pointer('updated-type','/results/0/resource/type',S+'types/task'),pointer('updated-properties','/results/0/resource/properties',{'title':prefix}),pointer('updated-metadata','/results/0/resource/metadata',{'phase':prefix})]
  for i,id in enumerate(deleted):
   base='/results/'+str(i+1)+'/deleted/'
   positive += [pointer(id+'-kind',base+'resourceKind','bead' if id=='a' else 'link'),pointer(id+'-id',base+'resource/id',S+('beads/' if id=='a' else 'links/')+id),pointer(id+'-type',base+'resource/type',S+('types/task' if id=='a' else 'types/relationship')),dict(id='control-'+id+'-revision',kind='json-pointer',pointer=base+'resource/revision',exists=True,equalsBinding='d-'+id+'-revision')]
  if k=='committed-prefix-state':a['assertions']=[q for q in a['assertions'] if q['id'] in ['http-status','present-bdp-scope-epoch','present-bdp-authorization-view','present-bdp-scope-position','schema','request-headers','request-body']]
  a['assertions']+=positive
 emit('deletion-'+k,deletion)
for k in ['batch-semantic-conflict','sequence-fresh-key']:
 aid='x-create-batch' if k=='batch-semantic-conflict' else 'x-create-sequence'
 old=next(a for a in row['actions'] if a['id']==aid)
 if k=='batch-semantic-conflict':
  positives=[dict(id='control-status',kind='status',equals=409),dict(id='control-schema',kind='json-schema',schema='#/$defs/transactionalProblem'),dict(id='control-media',kind='header',name='content-type',equals='application/problem+json')]+[dict(id='control-'+('problem-status' if n=='status' else n),kind='json-pointer',pointer='/'+n,exists=True,equals=v) for n,v in [('type','https://github.com/gastownhall/bdp/problems/conflict'),('code','idempotency-conflict'),('status',409),('retry','never')]]
  failing=['http-status','schema','media','serving-context','disposition','detail','key','no-next','required-position','effect-position','literal-result','retained-root','retained-body']
 else:
  positives=[dict(id='control-member',kind='json-array-tuples',pointer='/results',ordered=True,projections=[{'pointer':x} for x in ['/operationIndex','/operationName','/type','/code','/status','/retry']],equals=[[0,'first','https://github.com/gastownhall/bdp/problems/conflict','identity-taken',409,'never']]),dict(id='control-no-resource',kind='json-pointer',pointer='/results/0/resource',exists=False),dict(id='control-no-outcome',kind='json-pointer',pointer='/results/0/outcome',exists=False)];failing=['presentation','created-record']
 def cross(v,k=k,aid=aid,positives=positives):
  a=at(v,'idempotency.cross-carrier',aid);b=json.loads(a['raw']['body']['value'])
  if k=='batch-semantic-conflict':b['operations'][0]['properties']['title']='carrier-conflict'
  else:b['operations'][0]['idempotencyKey']='carrier-create-new'
  a['raw']['body']['value']=js(b);a['assertions']+=positives
 emit('cross-'+k,cross,dict(action=aid,retainedAssertionIds=[q['id'] for q in old['assertions']],removedAssertionIds=[],positiveAssertionIds=[q['id'] for q in positives],expectedFailingIds=failing,expectedPassingIds=[q['id'] for q in old['assertions']+positives if q['id'] not in failing]))
assert len(D)==37,len(D)
routes=[{'method':method,'url':S+url} for method,url in [('GET','beads/a'),('POST','operations/batch'),('GET','links/l'),('GET','beads/?limit=32'),('GET','events/'),('GET','beads/new'),('GET','changes/?after=genesis'),('GET','links/?limit=32'),('GET','beads/b'),('GET','beads/c'),('POST','operations/update-where'),('POST','operations/delete-where'),('POST','operations/create-bead'),('POST','operations/sequence'),('POST','operations/update-bead')]]
preimage={'maximumRequestBodyBytes':65536,'maximumRequestHeaderBytes':16384,'profile':'transactional','routes':routes,'scope':S}
assert sha(js(preimage))=='a9a17908b6fa87674d5f573b584f4a1ce1d41d7909fadf48b65e9b05397ddfd4'
(L/'route-preimage.json').write_text(js(preimage))
report=dict(parent='eff171839fe4f37638284a4c3346428979d26b5d',canonical={**props(m), 'diskBytes':len(paths['manifest'].read_bytes()),'diskSHA256':sha(paths['manifest'].read_bytes())},coverage={'diskBytes':len(paths['coverage'].read_bytes()),'normalBytes':len(normal(c).encode()),'diskSHA256':sha(paths['coverage'].read_bytes()),'normalSHA256':sha(normal(c))},preservation=preservation,row={'id':rowid,'compactSHA256':sha(js(row)),'perAction':[{ 'id':a['id'],'family':a['family'],'assertionIds':[q['id'] for q in a['assertions']],'captures':a['captures']} for a in row['actions']]},derivatives=D,maxDerivativeBytes=max(x['normalBytes'] for x in D),maxDerivativeNodes=max(x['nodes'] for x in D),schemaReferencesResolved=refs,routeDigest=sha(js(preimage)),unchangedCanonicalDerivatives='15 other observe invocations do not change manifest bytes: metadata5, abort/preabort/duplicate4, deniedall1, omittedroutes2, deniedset2, plus positive1.',newAppCases=59,newObserveInvocations=52)
(L/'measurements.json').write_text(normal(report));(L/'authored-row.json').write_text(normal(row))
print(json.dumps({k:v for k,v in report.items() if k not in ['row','derivatives','schemaReferencesResolved']},indent=2))
