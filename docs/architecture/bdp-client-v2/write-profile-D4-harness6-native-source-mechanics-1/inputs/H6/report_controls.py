"""Synthetic data controls on the actual auditor, not installed evidence.
Every PID/configuration/body below is authored unit data, not an observed child.
Protocol AJV acceptance is a separate unresolved prerequisite, never inferred.
"""
import base64,copy,hashlib,json
from pathlib import Path
from unittest.mock import patch
from scanner_controls import refusal
SCOPE='https://reference.invalid/acme/'
CATEGORIES=['abort','configuration','timeout','disconnect','header-limit','body-limit','invalid-body','network','unavailable']
def exact_refusal(call,expected):
 try:call()
 except RuntimeError as exc:
  assert str(exc)==expected,('wrong-full-auditor-predicate',expected,str(exc))
  return {'refused':True,'exceptionType':type(exc).__name__,'reason':str(exc),'sameCasePositiveBaselineRequired':True}
 raise AssertionError('full auditor accepted counterfactual')
def wire(row,body):
 text=json.dumps(body,separators=(',',':'));headers={'content-length':str(len(text.encode())),'bdp-scope-epoch':'unit-epoch','bdp-authorization-view':'unit-view','bdp-scope-position':'unit-pos'}
 raw=('HTTP/1.1 '+str(row['status'])+' Unit\r\n'+''.join(k+': '+v+'\r\n'for k,v in headers.items())+'\r\n'+text).encode()
 row.update(headers=headers,body=text,wireResponseBase64=base64.b64encode(raw).decode(),wireSha256=hashlib.sha256(raw).hexdigest())
 return row
def child(generation=1):
 return dict(generation=generation,phase='initial'if generation==1 else'reopen',launchAttempted=True,acquisition='acquired',refusal=None,pid=1000+generation,exitCodeSnapshot=0,signalCodeSnapshot=None,observedExit={'code':0,'signal':None},observedClose={'code':0,'signal':None},closeRequested=True,closePromiseSettled=True,closePromiseSucceeded=True,closure='observed-clean-exit-and-close')
def row(ordinal,target,body,post=None,generation=1,status=200):
 value=dict(ordinal=ordinal,generation=generation,method='POST'if post is not None else'GET',url=SCOPE+target if not target.startswith('https://')else target,settled=True,received=True,validated=True,status=status,headers=None,body=None,wireResponseBase64=None,wireSha256=None,writeState='complete',errorCategory=None)
 if post is not None:value['requestBody']=json.dumps(post,separators=(',',':'))
 return wire(value,body)
def pending(ordinal,target,post=None,settled=False,category=None):
 value=row(ordinal,target,{},post);value.update(settled=settled,received=False,validated=False,status=None,headers=None,body=None,wireResponseBase64=None,wireSha256=None,writeState=None,errorCategory=category);return value
def partial(rows,traversals=None,derived=None):
 observed=dict(attempted=len(rows),settled=sum(r['settled']for r in rows),received=sum(r['received']for r in rows),validated=sum(r['validated']for r in rows),returnedChildHandles=1,receiptPages=[None,None],receiptComplete=False,receiptEntries=None,eventPages=[],events=None,eventTraversalComplete=False,groups=None)
 if derived:observed.update(derived)
 return {'outcome':'failed','observations':rows,'counts':{'observed':observed},'eventTraversals':traversals or {k:{'pages':[],'complete':False,'count':None}for k in ['seed','deletion']},'lifecycle':{'pendingRequests':len(rows)-observed['settled'],'children':[child()],'ownershipUnknown':False,'directory':{'disposition':'retained'}},'source':None,'entry':None,'entrySha256':None,'workspaceDirty':None,'metadata':dict(source=None,entry=None,entrySha256=None,workspaceDirty=None)}
def positive(A,G,P,root):
 # This intentionally implements test DATA only. All checking remains audit_d4.
 fixture=P/'context/write-profile-D4-source-authoring-3/source/packages/conformance/test-support/transactional/fixture-v1.json'
 if not fixture.is_file():fixture=G.W/'packages/conformance/test-support/transactional/fixture-v1.json'
 fixture_bytes=fixture.read_bytes();f=json.loads(fixture_bytes)
 W=root/'unit-worktree';target=W/'packages/conformance/test-support/transactional/fixture-v1.json';target.parent.mkdir(parents=True);target.write_bytes(fixture_bytes)
 entry=root/'unit-entry.mjs';entry.write_text('// unit data, never imported\n');payload='1'*64;artifact='2'*64
 proof={'entryRealpath':str(entry.resolve()),'entrySha256':hashlib.sha256(entry.read_bytes()).hexdigest(),'payloadIdentity':payload}
 cfg=json.loads((P/'commands.json').read_bytes());spec=cfg['commands']['installed-owned-set'];name=spec['selectedFullNames'][0]
 suite=W/spec['selectedFilesByFullName'][name];result={'success':True,'numFailedTests':0,'numFailedTestSuites':0,'numPassedTests':1,'numTotalTests':1,'numPendingTests':0,'numTodoTests':0,'numTotalTestSuites':1,'numPassedTestSuites':1,'numPendingTestSuites':0,'testResults':[{'name':str(suite),'status':'passed','message':'','assertionResults':[{'fullName':name,'status':'passed','duration':1}]}]}
 targets={n:SCOPE+'operations/'+s for n,s in zip(['createBead','updateBead','deleteBead','createLink','updateLink','deleteLink','putAlias','deleteAlias','updateWhere','deleteWhere','batch','sequence'],['create-bead','update-bead','delete-bead','create-link','update-link','delete-link','put-alias','delete-alias','update-where','delete-where','batch','sequence'])}
 ready=dict(event='qualification.ready',authorityMode='canonical',runtimeConfigurationIdentity='3'*64,artifactIdentity='development-schema5-'+artifact,policyIdentity='4'*64,authenticationIdentity='5'*64,scope=SCOPE,profile='transactional',history=False,claimEligible=False,listenUrl='http://127.0.0.1:12345/',targets=targets,payloadIdentity=payload)
 source=G.INTEGRATION;fixture_sha=hashlib.sha256(fixture_bytes).hexdigest()
 metadata=dict(mode='external',source=source,entry=str(entry.resolve()),entrySha256=proof['entrySha256'],fixtureSha256=fixture_sha,workspaceDirty='not-applicable-externally-bound',diffSha256=None,disposition='not-applicable-externally-bound',nullReason=None)
 served=dict(scope=SCOPE,profile='transactional',history=False,sourceCommit=source,installedPayloadDigest=payload,installedContractsDigest=artifact,policyDigest=ready['policyIdentity'],authenticationDigest=ready['authenticationIdentity'],fixtureDigest=fixture_sha,fixtureId=f['id'],capabilities=sorted(f['capabilities']))
 payloads=json.loads((P/'context/semantic-audit-contract/request-payloads.json').read_bytes())['exactPOSTBodiesBySuccessOrdinal']
 ids=[SCOPE+f'links/l{i:02d}'for i in range(32)]
 receipt={'id':SCOPE+'receipts/unit','status':'completed','next':SCOPE+'receipts/unit?next=2','results':[{'operationIndex':0,'outcome':'matched','count':32}]+[{'operationIndex':0,'outcome':'deleted','deleted':{'resource':{'id':id}}}for id in ids[:31]]}
 page={'next':None,'results':[{'operationIndex':0,'outcome':'deleted','deleted':{'resource':{'id':ids[-1]}}}]}
 events=[{'id':f'event-{i}'}for i in range(99)];deleted=[{'type':kind,'subject':id}for id in ids for kind in ['deleted','unlinked','unlinked']]
 before={'id':SCOPE+'beads/owner','revision':'b1','properties':{}};linkbefore={'id':SCOPE+'links/owned','revision':'l1','properties':{'n':0}}
 linkafter={**linkbefore,'revision':'l2','properties':{'n':1}};after={**before,'revision':'b2','ownedLinks':{SCOPE+'types/owned':[linkafter]}}
 updatedentry={'resource':linkafter,'sourceRevision':'b2'};updated={'effectPosition':'p2','requiredPosition':'p2','results':[{},updatedentry]};failed={'status':'failed','problem':{'code':'validation-failed','operationIndex':1,'pointer':'/operations/1/change'}}
 bodies=[{}, {}, *[{'events':events[i*32:(i+1)*32],'next':SCOPE+f'events/?page={i+1}'if i<3 else None}for i in range(4)],{'groups':[],'after':'baseline'},receipt,page,*[{'events':deleted[i*32:(i+1)*32],'next':SCOPE+f'events/?delete={i+1}'if i<2 else None}for i in range(3)],{'groups':[{'events':deleted,'changes':[{'operation':'tombstone','resource':{'id':id}}for id in ids]}],'next':None},{},{},receipt,{'id':SCOPE+'links/later'},{'requiredPosition':'p1'},before,linkbefore,updated,after,linkafter,{'results':[{},updatedentry]},failed,after,linkafter,{},receipt,page,receipt,failed]
 urls=['operations/batch','operations/batch','events/',*['events/?page='+str(i)for i in [1,2,3]],'changes/?start=now','operations/delete-where',receipt['next'],'events/?after=event-98','events/?delete=1','events/?delete=2','changes/?after=baseline','links/l00','operations/batch','operations/batch','links/later','operations/batch','beads/owner','links/owned','operations/update-where','beads/owner','links/owned','operations/update-where','operations/batch','beads/owner','links/owned','beads/tentative',receipt['id'],receipt['next'],'operations/batch','operations/batch']
 assert len(bodies)==len(urls)==32
 rows=[row(i,urls[i-1],bodies[i-1],payloads.get(str(i)),1 if i<=28 else 2,404 if i in [14,28]else 200)for i in range(1,33)]
 expected=dict(requests=32,POST=11,GET=21,children=2,receiptEntries=33,receiptPages=[32,1],eventPages=[32,32,32],events=96,groups=1,staticLocalOracleCount=5)
 observed=dict(attempted=32,settled=32,received=32,validated=32,returnedChildHandles=2,receiptEntries=33,receiptPages=[32,1],receiptComplete=True,eventPages=[32,32,32],events=96,eventTraversalComplete=True,groups=1)
 o=dict(kind='installed-valid-set-observation',reportVersion=2,source=source,workspaceDirty='not-applicable-externally-bound',entry=str(entry.resolve()),entrySha256=proof['entrySha256'],metadata=metadata,binding=dict(served=served,operationTargets=targets,runtimeConfigurationIdentity=ready['runtimeConfigurationIdentity'],authorityMode='canonical'),claimEligible=False,initialReady=ready,reopenedReady={**ready,'listenUrl':'http://127.0.0.1:12346/'},outcome='observed',failureStages=[],counts=dict(expected=expected,observed=observed),observations=rows,oracleOutcomes=[dict(id='O'+str(i),state='pass')for i in range(1,6)],eventTraversals={'seed':dict(pages=[32,32,32,3],count=99,complete=True),'deletion':dict(pages=[32,32,32],count=96,complete=True)},timing=dict(workBudgetMilliseconds=90000,cleanupBudgetMilliseconds=10000,workMilliseconds=1,cleanupMilliseconds=1),lifecycle=dict(pendingRequests=0,ownershipUnknown=False,children=[child(1),child(2)],directory=dict(path=str(root/'bdp-set-installed-unit-absent'),disposition='removed',storeFilePresence='unknown'),descendantOwnership='root-terminal-audit-required'),limitations=['unit-only']*4)
 g=dict(finalHead=source,sourceBinding={'sha256':'6'*64});return o,result,g,spec,proof,W,payloads

def run(case,A,G,P,root):
 o,result,g,spec,proof,W,payloads=positive(A,G,P,root)
 def full(value):
  with patch.object(G,'W',W):return A.audit_d4(value,result,g,spec,root,proof)
 def check(value):
  got=A.audit_partial(value);assert got['positiveCredit']is False and got['claimEligible']is False;return got
 if case=='D01':
  got=full(o);assert got['claimEligible']is False and got['actualExchanges']==32
  return {'actualPythonAuditor':got,'unresolved':['protocol AJV full positive fixture acceptance is not established by Python-only synthetic bodies'],'scope':'pure synthetic crossrow unit; no product children/config truth'}
 if case=='D02':
  seed=[{'id':'unit-last'}];base=[row(1,'operations/batch',{},payloads['1']),row(2,'operations/batch',{},payloads['2']),row(3,'events/',{'events':seed,'next':None})]
  traversals={'seed':dict(pages=[1],count=1,complete=True),'deletion':dict(pages=[],count=None,complete=False)}
  divergent=partial(base+[pending(4,'changes/?start=now',settled=True,category='network')],traversals)
  final=partial(base);unsupported=refusal(lambda:check(final),'failed-unsupported-final-event-boundary')
  baseline=row(4,'changes/?start=now',{'groups':[],'after':'unit'})
  first=partial(base+[baseline,row(5,'operations/delete-where',{'results':[{}]},payloads['8'])],traversals,{'receiptPages':[1,None]})
  second=partial(base+[baseline,row(5,'operations/delete-where',receipt_body:=dict(status='completed',next=SCOPE+'r/next',results=[{}]*32),payloads['8']),row(6,'r/next',{'results':[{}],'next':None})],traversals,{'receiptPages':[32,1],'receiptComplete':True,'receiptEntries':33})
  return {'divergent':check(divergent),'unsupportedFinal':unsupported,'assignmentBeforeAssertion':[check(first),check(second)]}
 if case=='D03':
  baseline=full(o);assert baseline['actualExchanges']==32 and baseline['claimEligible']is False
  outcomes=[]
  for s in [False,True]:
   for r in [False,True]:
    for v in [False,True]:
     value=row(1,'operations/batch',{},payloads['1'])if r else pending(1,'operations/batch',payloads['1'],settled=s,category='network'if s else None)
     value.update(settled=s,received=r,validated=v);p=partial([value]);bits=f'{int(s)}{int(r)}{int(v)}'
     got=check(p)if bits in ['000','100','110','111']else refusal(lambda p=p:check(p))
     outcomes.append({'state':bits,'result':got})
  for category in CATEGORIES:
   p=partial([pending(1,'operations/batch',payloads['1'],category=category)]);outcomes.append({'pendingCategory':category,'result':refusal(lambda p=p:check(p),'pending-category-write-state-null')})
   settled=partial([pending(1,'operations/batch',payloads['1'],settled=True,category=category)]);check(settled)
  for state in ['not-started','started-completion-unestablished','complete']:
   p=partial([pending(1,'operations/batch',payloads['1'])]);p['observations'][0]['writeState']=state;outcomes.append({'pendingWriteState':state,'result':refusal(lambda p=p:check(p),'pending-category-write-state-null')})
  null=partial([pending(1,'operations/batch',payloads['1'],settled=True)]);refusal(lambda:check(null),'settled-rejection-category')
  # All declared derived-field mutations are exercised against actual empty prefix.
  for key,value in [('receiptPages',[0,None]),('receiptPages',[None,0]),('receiptComplete',True),('receiptEntries',0),('groups',0),('eventPages',[0]),('events',0),('eventTraversalComplete',True)]:
   p=partial([]);p['counts']['observed'][key]=value;refusal(lambda p=p:check(p))
  p=partial([]);p['lifecycle']['pendingRequests']=1;refusal(lambda:check(p))
  for kind in ['seed','deletion']:
   for key,value in [('pages',[1]),('count',0),('complete',True)]:
    p=partial([]);p['eventTraversals'][kind][key]=value;refusal(lambda p=p:check(p),'failed-source-event-assignments')
  for mutate in ['wire','body','postimage']:
   p=copy.deepcopy(o)
   if mutate=='wire':p['observations'][0]['wireResponseBase64']=None
   elif mutate=='body':p['observations'][7]['body']='{}'
   else:wire(p['observations'][21],{'revision':'wrong','properties':{}})
   exact_refusal(lambda p=p:full(p),{'wire':'wire-base64-type','body':'wire-body','postimage':'owned-inline-metadata'}[mutate])
  return {'stateControls':outcomes,'declared20StateSubcases':20,'scope':'remaining semantic checks still required; failure classifications never positive'}
 if case=='D04':
  baseline=full(o);assert baseline['actualExchanges']==32 and baseline['claimEligible']is False
  receipts=[]
  for change in [{'observedClose':None},{'closePromiseSucceeded':False},{'pid':None}]:
   p=copy.deepcopy(o);p['lifecycle']['children'][0].update(change);receipts.append(exact_refusal(lambda p=p:full(p),'actual-child-pid'if 'pid'in change else'typed-report-value'))
  return {'subcases':receipts}
 if case=='D05':
  full(o)
  assert all('requestBody'not in r for r in o['observations']if r['method']=='GET')
  assert all(not any(k in r for k in ['authorization','idempotency-key','requestHeaders'])for r in o['observations'])
  p=copy.deepcopy(o);p['observations'][2]['requestBody']='{}';return refusal(lambda:full(p),'actual-request-body-presence')
 if case=='D06':
  baseline=full(o);assert baseline['actualExchanges']==32 and baseline['claimEligible']is False
  receipts=[]
  for path in [('source',),('metadata','fixtureSha256'),('initialReady','payloadIdentity'),('initialReady','runtimeConfigurationIdentity'),('binding','served','sourceCommit')]:
   p=copy.deepcopy(o);parent=p
   for key in path[:-1]:parent=parent[key]
   parent[path[-1]]='f'*64;expected='fixture-proof'if path==('metadata','fixtureSha256')else'same-live-configuration'if path==('initialReady','runtimeConfigurationIdentity')else'typed-report-value';receipts.append(exact_refusal(lambda p=p:full(p),expected))
  return {'bindingRefusals':receipts,'unresolved':['actual archive/build/source before-after verification requires future fresh measured inventory; fake proof is not archive evidence']}
 raise AssertionError('unregistered report control')
