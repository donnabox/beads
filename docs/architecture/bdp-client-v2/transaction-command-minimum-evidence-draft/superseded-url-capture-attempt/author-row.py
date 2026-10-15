import json,hashlib
from pathlib import Path
W=Path('/Users/dbox/repos/wt/vickie/vickie-command-minimum-evidence')
D=Path(__file__).parent
S='https://reference.invalid/acme/'
ID='read-update.http.command-minimum-fields'
def ptr(id,pointer,equals=None,*,binding=None,absent=False):
 d=dict(id=id,kind='json-pointer',pointer=pointer,exists=not absent)
 if not absent:
  if binding is not None:d['equalsBinding']=binding
  else:d['equals']=equals
 return d
def cap(binding,pointer):return dict(binding=binding,**{'from':dict(kind='json-pointer',pointer=pointer)})
def tuples(id,pointer,fields,rows):return dict(id=id,kind='json-array-tuples',pointer=pointer,ordered=True,projections=[dict(pointer=f) for f in fields],equals=rows)
N=['bdp-minimum-scope-position','bdp-scope-epoch','bdp-authorization-view']
C=['bdp-scope-epoch','bdp-authorization-view','bdp-scope-position']
sets=dict(M=list(zip(N,['not a position!','','view-one, view-two'])),F=list(zip(N,['pos-99','foreign-epoch','foreign-view'])),P=[(N[1],'foreign-epoch')],Z=list(zip(N,['','',''])),O=[])
def headers(key,fields,body=True):return [dict(name=n,value=v) for n,v in ([('content-type','application/json')] if body else [])+([('idempotency-key',key)] if key else [])+([('accept','application/json')] if body else [])+sets[fields]]
def common(pos,schema='mutationReceipt',status=200,media='application/json'):
 return [dict(id='http-status',kind='status',equals=status),dict(id='schema',kind='json-schema',schema='#/$defs/'+schema),dict(id='media',kind='header',name='content-type',equals=media)]+[dict(id='present-'+n,kind='header',name=n,absent=False) for n in C]+[dict(id='serving-position',kind='header',name=C[2],equals=f'pos-{pos}'),dict(id='request-headers',kind='request-authored-headers'),dict(id='request-body',kind='request-authored-body')]
def make(id,route,key,fields,body,pos,schema='mutationReceipt',status=200):
 raw=dict(headerLines=headers(key,fields,body is not None))
 if body is not None:raw['body']=dict(encoding='utf8',value=json.dumps(body,separators=(',',':')))
 return dict(id=id,family='http',method='POST' if body is not None else 'GET',target=dict(binding='scope',path=route.split('?')[0],**({'query':dict(limit='32')} if '?' in route else {})),credentialRef='operator',raw=raw,captures=[],assertions=common(pos,schema,status,'application/problem+json' if status!=200 else 'application/json'))
def create(k):return dict(id='beads/'+k,type=S+'types/task',properties=dict(title='minimum-'+k))
def batch(k,title=None):
 b=dict(operation='createBead',**create(k))
 if title:b['properties']['title']=title
 return dict(operations=[b])
def seq():return dict(operations=[dict(operation='createBead',**create('c'),idempotencyKey='minimum-c')])
def receipt(a,k,pos):
 a['assertions'] += [ptr('disposition','/status','completed'),ptr('detail','/detail','available'),ptr('key','/idempotencyKey','minimum-'+k),ptr('no-next','/next',None),ptr('required-position','/requiredPosition','pos-'+str(pos)),ptr('effect-position','/effectPosition','pos-'+str(pos))]
def resource(a,k):
 a['assertions'] += [tuples('literal-result','/results',['/operationIndex','/outcome','/resource/id','/resource/type','/resource/properties','/resource/metadata'],[[0,'created',S+'beads/'+k,S+'types/task',dict(title='minimum-'+k),{}]]),ptr('no-name','/results/0/operationName',absent=True),ptr('no-source','/results/0/source',absent=True),ptr('no-provenance','/results/0/resource/revisionProvenance',absent=True),ptr('no-owned-links','/results/0/resource/ownedLinks',absent=True)]
 a['captures'] += [cap('minimum-'+k+'-body',''),cap('minimum-'+k+'-record','/results/0/resource'),cap('minimum-'+k+'-revision','/results/0/resource/revision')]
 if k!='c':a['captures'] += [cap('minimum-'+k+'-root','/id'),cap('minimum-'+k+'-transaction','/transaction')]
a=make('m-create-a','operations/create-bead','minimum-a','M',create('a'),1);receipt(a,'a',1);resource(a,'a')
ar=make('m-replay-a','operations/create-bead','minimum-a','O',create('a'),1);receipt(ar,'a',1);ar['assertions']+=[ptr('retained-body','',binding='minimum-a-body'),dict(id='same-context',kind='response-metadata-equals',request='m-create-a',headers=C)]
b=make('m-create-b','operations/batch','minimum-b','P',batch('b'),2);receipt(b,'b',2);resource(b,'b')
br=make('m-replay-b','operations/batch','minimum-b','Z',batch('b'),2);receipt(br,'b',2);br['assertions']+=[ptr('retained-body','',binding='minimum-b-body'),dict(id='same-context',kind='response-metadata-equals',request='m-create-b',headers=C)]
c=make('m-create-c','operations/sequence',None,'F',seq(),3,'transactionalSequenceResponse');resource(c,'c');c['assertions'] += [ptr('no-receipt','/receipt',absent=True),ptr('no-id','/id',absent=True),ptr('no-problem','/results/0/problem',absent=True)]
cr=make('m-replay-c','operations/sequence',None,'M',seq(),3,'transactionalSequenceResponse');cr['assertions'] += [ptr('retained-body','',binding='minimum-c-body'),dict(id='same-context',kind='response-metadata-equals',request='m-create-c',headers=C)]
conflict=dict(type='https://github.com/gastownhall/bdp/problems/conflict',code='idempotency-conflict',status=409,retry='never')
f=make('m-body-conflict','operations/batch','minimum-b','F',batch('b','minimum-conflict'),3,'transactionalProblem',409);f['assertions'] += [ptr('whole-problem','',conflict),dict(id='same-context',kind='response-metadata-equals',request='m-create-c',headers=C)]
u=make('m-stale-update','operations/update-bead','minimum-stale','M',dict(bead='beads/a',expectedRevision='minimum-wrong-revision',propertiesChange=[dict(op='replace',path='/title',value='minimum-stale')]),3)
u['assertions'] += [ptr('disposition','/status','failed'),ptr('detail','/detail','available'),ptr('key','/idempotencyKey','minimum-stale'),ptr('required-position','/requiredPosition','pos-3'),ptr('no-effect','/effectPosition',absent=True),ptr('no-results','/results',absent=True),ptr('no-next','/next',absent=True),ptr('whole-problem','/problem',dict(type=conflict['type'],code='revision-mismatch',status=409,retry='after-state-change',operationIndex=0)),dict(id='same-context',kind='response-metadata-equals',request='m-create-c',headers=C)]
r=dict(id='m-retained-a',family='client',operation='transactional.receipt-read',input=dict(binding='minimum-a-root',method='GET'),captures=[],assertions=[ptr('status','/status',200),ptr('body-present','/bodyPresent',True),ptr('media','/headers/content-type','application/json')]+[dict(id='present-'+n,kind='json-pointer',pointer='/headers/'+n,exists=True) for n in C]+[ptr('serving-position','/headers/bdp-scope-position','pos-3'),ptr('private-cache','/headers/cache-control','private, no-store'),ptr('no-etag','/headers/etag',absent=True),ptr('no-last-modified','/headers/last-modified',absent=True),ptr('retained-body','/body',binding='minimum-a-body'),ptr('root','/body/id',binding='minimum-a-root'),ptr('required-position','/body/requiredPosition','pos-1'),ptr('effect-position','/body/effectPosition','pos-1'),ptr('key','/body/idempotencyKey','minimum-a')])
get=make('m-current','beads/?limit=32',None,'O',None,3,'currentBeadCollection')
get['assertions'] += [ptr('no-next','/next',None),tuples('ordered-state','/items',['/id','/type','/properties','/metadata'],[[S+'beads/'+k,S+'types/task',dict(title='minimum-'+k),{}] for k in 'abc'])]+[ptr(k+'-full-record','/items/'+str(i),binding='minimum-'+k+'-record') for i,k in enumerate('abc')]+[dict(id='same-context',kind='response-metadata-equals',request='m-create-c',headers=C)]
e=make('m-events','events/',None,'O',None,3,'eventPage')
e['assertions'] += [ptr('no-next','/next',None),tuples('ordered-facts','/events',['/id','/source','/type','/subject','/subjectType','/ordinal'],[[f'pos-{i+1}-0',S+'events/','created',S+'beads/'+k,S+'types/task',0] for i,k in enumerate('abc')])]
for i,k in enumerate('abc'):
 p='/events/'+str(i)
 e['assertions'] += [ptr(k+'-properties',p+'/data/properties',dict(title='minimum-'+k)),ptr(k+'-metadata',p+'/data/metadata',{}),ptr(k+'-revision',p+'/data/revision',binding='minimum-'+k+'-revision'),ptr(k+'-no-provenance',p+'/data/revisionProvenance',absent=True)]
 if k!='c':e['assertions'] += [ptr(k+'-transaction',p+'/transaction',binding='minimum-'+k+'-transaction')]
e['assertions'] += [dict(id='same-context',kind='response-metadata-equals',request='m-create-c',headers=C)]
row=dict(id=ID,requiredProfile='read-update',setup=dict(fixture='reference-transactional-v1',requires=['installed-transactional-http','transactional-receipt-observation']),applicability=dict(requires=[]),cleanup=dict(resetFixture=True),actions=[a,ar,b,br,c,cr,f,u,r,get,e])
coverage=dict(id=ID,wholeContract=False,reason='Bounded actual Transactional singleton, batch and sequence observations of ignored literal copied minimum fields, retained identity, conflict, revision refusal and authorization control; complete three-Bead and Scope Event page. Whole contract remains unqualified: this row omits receipt/read minimum enforcement, coherent same-context future tuples, other singleton target/kind and carrier-validation/transport/bounds partitions, and actual Read+Update-profile execution. Historical private observations are not relabeled.',bindings=[dict(scenario=ID,assertions=[a['id']+':'+x['id'] for a in row['actions'] for x in a['assertions']])])
for name,key,value in [('manifest','scenarios',row),('coverage','rows',coverage)]:
 p=W/f'packages/conformance/test-support/transactional/{name}-v1.json';old=p.read_text();assert ID not in old
 (D/f'parent-{name}.json').write_text(old)
 fragment='\n'.join('    '+line for line in json.dumps(value,indent=2,ensure_ascii=False).splitlines())
 p.write_text(old.replace(f'  "{key}": [\n',f'  "{key}": [\n'+fragment+',\n',1))
 (D/f'new-{name}-row.json').write_text(json.dumps(value,indent=2)+'\n')
print('assertions',len(coverage['bindings'][0]['assertions']))
