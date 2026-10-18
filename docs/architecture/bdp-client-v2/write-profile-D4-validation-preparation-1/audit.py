"""D4 actual report audit. All IO runs in owned workers, never watchdog/product."""
import base64,hashlib,json,math,re
from pathlib import Path
from urllib.parse import urlsplit,parse_qs,quote,urljoin
from scanner import need,strict_json,RESULTS,file_json,file_hash
SCOPE='https://reference.invalid/acme/'
def keys(o,want,optional=()):need(type(o)is dict and set(want)<=set(o)<=set(want)|set(optional),'report-keys')
def exact(o,want):
 for k,v in want.items():need(k in o and type(o[k])is type(v) and o[k]==v,'typed-report-value')
def finite(v,maximal):need(type(v)in [int,float] and math.isfinite(v) and 0<=v<=maximal,'report-time')
def digest(v):need(type(v)is str and re.fullmatch('[a-f0-9]{64}',v),'digest')
def strings(o):need(type(o)is dict and all(type(k)is str and type(v)is str for k,v in o.items()),'headers')
def received_wire(row,parse_body=True):
 encoded=row['wireResponseBase64'];need(type(encoded)is str,'wire-base64-type');data=base64.b64decode(encoded,validate=True)
 need(base64.b64encode(data).decode()==encoded and 0<len(data)<=1179648,'canonical-wire-size');need(hashlib.sha256(data).hexdigest()==row['wireSha256'],'wire-digest')
 offset=0
 while True:
  end=data.find(b'\r\n\r\n',offset);need(end>=0 and end+4<=65536,'wire-header');lines=data[offset:end].decode('latin1').split('\r\n');match=re.match(r'HTTP/1\.[01] ([1-5][0-9]{2})(?: |$)',lines[0]);need(match,'wire-status');status=int(match[1]);headers={}
  for line in lines[1:]:
   name,colon,value=line.partition(':');name=name.strip('\t\n\v\f\r \xa0').lower();value=value.strip('\t\n\v\f\r \xa0');need(colon and re.fullmatch(r"[!#$%&'*+.^_`|~0-9a-z-]+",name) and not any(c in value for c in '\0\r\n'),'wire-header-field');headers[name]=headers[name]+', '+value if name in headers else value
  offset=end+4
  if status>=200:break
  need(status!=101 and 'content-length'not in headers and 'transfer-encoding'not in headers,'informational-framing')
 need(status==row['status'] and headers==row['headers'],'wire-response-projection');body=data[offset:];te=headers.get('transfer-encoding');cl=headers.get('content-length');need(te is None or cl is None,'ambiguous-framing')
 if te is not None:
  need(te.lower()=='chunked','transfer-coding');decoded=bytearray();rest=body
  while True:
   size,sep,rest=rest.partition(b'\r\n');number=size.split(b';',1)[0];need(sep and re.fullmatch(b'[0-9a-fA-F]+',number),'wire-chunk');n=int(number,16);need(n<=9007199254740991,'chunk-safe-integer')
   if n==0:
    if rest!=b'\r\n':
     need(rest.endswith(b'\r\n\r\n') and rest.find(b'\r\n\r\n')==len(rest)-4,'wire-trailer-end')
     for line in rest[:-4].decode('latin1').split('\r\n'):
      name,colon,value=line.partition(':');need(colon and re.fullmatch(r"[!#$%&'*+.^_`|~0-9a-z-]+",name.strip('\t\n\v\f\r \xa0').lower()) and not any(c in value for c in '\0\r\n'),'wire-trailer')
    break
   need(len(decoded)+n<=1048576 and len(rest)>=n+2 and rest[n:n+2]==b'\r\n','wire-chunk-size');decoded.extend(rest[:n]);rest=rest[n+2:]
  body=bytes(decoded)
 elif cl is not None:need(re.fullmatch(r'0|[1-9][0-9]*',cl) and int(cl)<=1048576 and len(body)==int(cl),'wire-content-length')
 need(len(body)<=1048576 and body.decode('utf-8','strict')==row['body'],'wire-body');return strict_json(body,1048576) if parse_body else body.decode('utf-8','strict')
def audit_partial(o):
 # This only classifies retained failed evidence. Never invokes a product or
 # substitutes positive32 ordinals for source's two actual-next loops.
 need(o['outcome']=='failed','failed-classification-only');rows=o['observations'];counts=o['counts']['observed'];children=o['lifecycle']['children'];payloads=file_json(Path(__file__).parent/'context/semantic-audit-contract/request-payloads.json',RESULTS)['exactPOSTBodiesBySuccessOrdinal']
 exact(counts,{'attempted':len(rows),'settled':sum(r['settled']is True for r in rows),'received':sum(r['received']is True for r in rows),'validated':sum(r['validated']is True for r in rows),'returnedChildHandles':sum(c['acquisition']=='acquired'for c in children)})
 def program():
  a=yield(1,'operations/batch',1);b=yield(1,'operations/batch',2)
  next='events/';last=None
  for _ in range(12):
   page=yield(1,next,None);last=page['events'][-1]['id'] if page['events'] else last
   if page['next']is None:break
   next=page['next']
  else:return
  baseline=yield(1,'changes/?start=now',None)
  receipt=yield(1,'operations/delete-where',8)
  page=yield(1,receipt['next'],None)
  next='events/?after='+quote(last,safe="~()*!.'-")
  for _ in range(12):
   page=yield(1,next,None)
   if page['next']is None:break
   next=page['next']
  else:return
  yield(1,'changes/?after='+quote(baseline['after'],safe="~()*!.'-"),None)
  yield(1,'links/l00',None);yield(1,'operations/batch',15);yield(1,'operations/batch',16);yield(1,'links/later',None);yield(1,'operations/batch',18)
  for target,post in [('beads/owner',None),('links/owned',None),('operations/update-where',21),('beads/owner',None),('links/owned',None),('operations/update-where',24),('operations/batch',25),('beads/owner',None),('links/owned',None),('beads/tentative',None)]:yield(1,target,post)
  yield(2,receipt['id'],None);yield(2,receipt['next'],None);yield(2,'operations/batch',31);yield(2,'operations/batch',32)
 plan=program();expected=next(plan) if rows else None
 for index,row in enumerate(rows):
  generation,target,post=expected;need(type(target)is str,'failed-dynamic-target');exact(row,{'ordinal':index+1,'generation':generation,'method':'POST'if post else'GET','url':urljoin(SCOPE,target)})
  if post:need(strict_json(row['requestBody'].encode(),65536)==payloads[str(post)],'failed-exact-post')
  need(not row['validated'] or row['received']and row['settled'],'failed-validation-implies-response')
  body=None
  if row['received']:
   need((row['wireResponseBase64']is None)==(row['wireSha256']is None),'partial-wire-availability')
   text=received_wire(row,False) if row['wireResponseBase64']is not None else row['body']
   if row['validated']:
    need(row['wireResponseBase64']is not None,'validated-wire-available');body=strict_json(text.encode(),1048576)
   need(row['errorCategory']is None,'received-no-transport-category')
  else:need(all(row[k]is None for k in ['status','headers','body','wireResponseBase64','wireSha256']),'unreceived-response-fields-null')
  if index+1<len(rows):
   need(row['validated']is True and body is not None,'requests-after-refused-response')
   try:expected=plan.send(body)
   except (StopIteration,KeyError,TypeError,ValueError):raise RuntimeError('failed-source-program-path')
 for kind in ['seed','deletion']:
  traversal=o['eventTraversals'][kind]
  if traversal['complete']:need(traversal['count']==sum(traversal['pages']),'failed-complete-event-count')
  else:need(traversal['count']is None,'failed-incomplete-event-null')
 exact(counts,{'eventPages':o['eventTraversals']['deletion']['pages'],'events':o['eventTraversals']['deletion']['count'],'eventTraversalComplete':o['eventTraversals']['deletion']['complete']})
 for index,c in enumerate(children,1):
  exact(c,{'generation':index,'phase':'initial'if index==1 else'reopen'})
  clean=c['acquisition']=='acquired' and c['pid']is not None and c['observedExit']=={'code':0,'signal':None} and c['observedClose']=={'code':0,'signal':None} and c['closePromiseSettled']is True and c['closePromiseSucceeded']is True
  need((c['closure']=='observed-clean-exit-and-close')==clean,'failed-child-clean-predicate')
  if c['acquisition']=='unknown':need(c['pid']is None and c['observedExit']is None and c['observedClose']is None and c['closePromiseSettled']is None and c['closePromiseSucceeded']is None and c['closeRequested']is False,'failed-unknown-acquisition')
  if not clean:need(o['lifecycle']['ownershipUnknown']is True,'failed-sticky-owner')
 if o['lifecycle']['directory']['disposition']=='removed':need(o['lifecycle']['pendingRequests']==0 and counts['settled']==counts['attempted'] and o['lifecycle']['ownershipUnknown']is False and all(c['closure']=='observed-clean-exit-and-close'for c in children),'failed-removal-owner')
 for key in ['source','entry','entrySha256','workspaceDirty']:need(o[key]==o['metadata'][key],'failed-metadata-top')
 return {'classification':'CONFORMING_FAILED_PARTIAL_NO_CREDIT','actualAttempted':len(rows),'positiveCredit':False,'claimEligible':False,'path':'finite source program; each actual-next loop<=12, dispatch<=32; not success32 prefix'}
def selected_results(result,spec,worktree):
 chosen=[]
 for suite in result['testResults']:
  for case in suite['assertionResults']:
   if case['fullName']in spec['selectedFullNames']:
    need(Path(suite['name']).resolve()==(worktree/spec['selectedFilesByFullName'][case['fullName']]).resolve(),'selected-file');chosen.append(case)
   else:need(case['status']in ['pending','skipped','todo'],'unselected-executed')
 need(len(chosen)==1 and chosen[0]['status']=='passed' and chosen[0]['fullName']==spec['selectedFullNames'][0],'callback-selection')
 need(len(result['testResults'])==1,'one-selected-source-file');exact(result,{'success':True,'numFailedTests':0,'numFailedTestSuites':0,'numPassedTests':1,'numTotalTests':1,'numPendingTests':0,'numTodoTests':0,'numTotalTestSuites':1,'numPassedTestSuites':1,'numPendingTestSuites':0});exact(result['testResults'][0],{'status':'passed','message':''})
 if 'numRuntimeErrorTestSuites'in result:exact(result,{'numRuntimeErrorTestSuites':0})
 finite(chosen[0]['duration'],120000)
def audit_d4(o,result,g,spec,tmp,proof):
 from gate import W,sha
 selected_results(result,spec,W)
 keys(o,['kind','reportVersion','source','workspaceDirty','entry','entrySha256','metadata','binding','claimEligible','initialReady','reopenedReady','outcome','failureStages','counts','observations','oracleOutcomes','eventTraversals','timing','lifecycle','limitations'])
 exact(o,{'kind':'installed-valid-set-observation','reportVersion':2,'source':g['finalHead'],'workspaceDirty':'not-applicable-externally-bound','claimEligible':False,'outcome':'observed','failureStages':[]})
 need(Path(o['entry']).resolve(strict=True)==Path(proof['entryRealpath']) and o['entrySha256']==proof['entrySha256'],'entry-proof')
 metadata=o['metadata'];keys(metadata,['mode','source','entry','entrySha256','fixtureSha256','workspaceDirty','diffSha256','disposition','nullReason']);exact(metadata,{'mode':'external','source':g['finalHead'],'entry':o['entry'],'entrySha256':o['entrySha256'],'workspaceDirty':'not-applicable-externally-bound','diffSha256':None,'disposition':'not-applicable-externally-bound','nullReason':None});fixture=W/'packages/conformance/test-support/transactional/fixture-v1.json';need(metadata['fixtureSha256']==sha(fixture.read_bytes()),'fixture-proof')
 ready_keys=['event','authorityMode','runtimeConfigurationIdentity','artifactIdentity','policyIdentity','authenticationIdentity','scope','profile','history','claimEligible','listenUrl','targets','payloadIdentity'];targets={n:SCOPE+'operations/'+s for n,s in zip(['createBead','updateBead','deleteBead','createLink','updateLink','deleteLink','putAlias','deleteAlias','updateWhere','deleteWhere','batch','sequence'],['create-bead','update-bead','delete-bead','create-link','update-link','delete-link','put-alias','delete-alias','update-where','delete-where','batch','sequence'])}
 for r in [o['initialReady'],o['reopenedReady']]:
  keys(r,ready_keys);exact(r,{'event':'qualification.ready','authorityMode':'canonical','scope':SCOPE,'profile':'transactional','history':False,'claimEligible':False,'payloadIdentity':proof['payloadIdentity'],'targets':targets})
  for k in ['runtimeConfigurationIdentity','policyIdentity','authenticationIdentity','payloadIdentity']:digest(r[k])
  need(re.fullmatch('development-schema5-[a-f0-9]{64}',r['artifactIdentity']),'artifact-identity');dial=urlsplit(r['listenUrl']);need(dial.scheme=='http' and dial.hostname=='127.0.0.1' and dial.port and dial.path=='/' and not dial.query and not dial.fragment and not dial.username and not dial.password,'actual-dial')
 for k in ready_keys:
  if k!='listenUrl':need(o['initialReady'][k]==o['reopenedReady'][k],'same-live-configuration')
 binding=o['binding'];keys(binding,['served','operationTargets','runtimeConfigurationIdentity','authorityMode']);exact(binding,{'operationTargets':targets,'runtimeConfigurationIdentity':o['initialReady']['runtimeConfigurationIdentity'],'authorityMode':'canonical'});served=binding['served'];keys(served,['scope','profile','history','sourceCommit','installedPayloadDigest','installedContractsDigest','policyDigest','authenticationDigest','fixtureDigest','fixtureId','capabilities']);f=strict_json(fixture.read_bytes(),16777216);exact(served,{'scope':SCOPE,'profile':'transactional','history':False,'sourceCommit':g['finalHead'],'installedPayloadDigest':proof['payloadIdentity'],'installedContractsDigest':o['initialReady']['artifactIdentity'].removeprefix('development-schema5-'),'policyDigest':o['initialReady']['policyIdentity'],'authenticationDigest':o['initialReady']['authenticationIdentity'],'fixtureDigest':metadata['fixtureSha256'],'fixtureId':f['id'],'capabilities':sorted(f['capabilities'])})
 keys(o['counts'],['expected','observed']);exact(o['counts']['expected'],{'requests':32,'POST':11,'GET':21,'children':2,'receiptEntries':33,'receiptPages':[32,1],'eventPages':[32,32,32],'events':96,'groups':1,'staticLocalOracleCount':5});exact(o['counts']['observed'],{'attempted':32,'settled':32,'received':32,'validated':32,'returnedChildHandles':2,'receiptEntries':33,'receiptPages':[32,1],'receiptComplete':True,'eventPages':[32,32,32],'events':96,'eventTraversalComplete':True,'groups':1})
 payloads=file_json(Path(__file__).parent/'context/semantic-audit-contract/request-payloads.json',RESULTS)['exactPOSTBodiesBySuccessOrdinal'];rows=o['observations'];need(type(rows)is list and len(rows)==32,'actual-exchanges');bodies=[]
 post={1,2,8,15,16,18,21,24,25,31,32}
 for i,r in enumerate(rows,1):
  keys(r,['ordinal','generation','method','url','settled','received','validated','status','headers','body','wireResponseBase64','wireSha256','writeState','errorCategory'],['requestBody']);exact(r,{'ordinal':i,'generation':1 if i<=28 else 2,'method':'POST' if i in post else 'GET','settled':True,'received':True,'validated':True,'status':404 if i in [14,28] else 200,'writeState':'complete','errorCategory':None});strings(r['headers']);digest(r['wireSha256']);need(type(r['body'])is str and type(r['wireResponseBase64'])is str and r['url'].startswith(SCOPE),'actual-row');need(('requestBody'in r)==(i in post),'actual-request-body-presence')
  if i in post:need(strict_json(r['requestBody'].encode(),65536)==payloads[str(i)],'exact-post-payload')
  if r['status']==200:need(all(type(r['headers'].get(k))is str and r['headers'][k] for k in ['bdp-scope-epoch','bdp-authorization-view','bdp-scope-position']),'actual-context')
  bodies.append(received_wire(r))
 need(sum(len(base64.b64decode(r['wireResponseBase64'],validate=True))for r in rows)<=37748736,'total-wire-cap');
 def at(i):return bodies[i-1]
 def url(i,value):need(rows[i-1]['url']==value,'actual-target-order')
 for i in post:url(i,SCOPE+('operations/delete-where' if i==8 else 'operations/update-where' if i in [21,24] else 'operations/batch'))
 url(3,SCOPE+'events/')
 for i in [4,5,6]:url(i,at(i-1)['next'])
 url(7,SCOPE+'changes/?start=now');url(9,at(8)['next']);seed=[e for i in [3,4,5,6] for e in at(i)['events']];need(len(seed)==99 and [len(at(i)['events'])for i in [3,4,5,6]]==[32,32,32,3] and at(6)['next']is None,'actual-seed-pages');url(10,SCOPE+'events/?after='+quote(seed[-1]['id'],safe="~()*!.'-"))
 for i in [11,12]:url(i,at(i-1)['next'])
 url(13,SCOPE+'changes/?after='+quote(at(7)['after'],safe="~()*!.'-"));need(at(7)['groups']==[],'baseline-feed')
 ids=[SCOPE+f'links/l{i:02d}'for i in range(32)];url(14,ids[0]);need(at(8)['status']=='completed' and at(8)['next']==rows[8]['url'] and at(9)['next']is None,'actual-receipt-page');entries=at(8)['results']+at(9)['results'];need([len(at(8)['results']),len(at(9)['results'])]==[32,1] and entries[0]=={'operationIndex':0,'outcome':'matched','count':32},'actual-33-receipts');need([e['deleted']['resource']['id']for e in entries[1:]]==ids and all(e['operationIndex']==0 and e['outcome']=='deleted'for e in entries[1:]),'actual-deletion-order')
 deletion=[e for i in [10,11,12]for e in at(i)['events']];need([len(at(i)['events'])for i in [10,11,12]]==[32]*3 and at(12)['next']is None,'actual-suffix-pages');need([{'type':e['type'],'subject':e['subject']}for e in deletion]==[{'type':kind,'subject':id}for id in ids for kind in ['deleted','unlinked','unlinked']],'actual96-event-order');need(len(at(13)['groups'])==1 and at(13)['next']is None and len(at(13)['groups'][0]['events'])==96 and [c['resource']['id']for c in at(13)['groups'][0]['changes']]==ids and all(c['operation']=='tombstone'for c in at(13)['groups'][0]['changes']),'actual-group')
 need(at(16)==at(8),'same-key-receipt');url(17,SCOPE+'links/later');need(at(17)['id']==SCOPE+'links/later','later-population-retained')
 for i in [19,22,26]:url(i,SCOPE+'beads/owner')
 for i in [20,23,27]:url(i,SCOPE+'links/owned')
 updated=at(21);entry=updated['results'][1];source_before,link_before=at(19),at(20);source_after,link_after=at(22),at(23)
 need(('metadata'in link_after,link_after.get('metadata'))==('metadata'in link_before,link_before.get('metadata')) and ('metadata'in source_after,source_after.get('metadata'))==('metadata'in source_before,source_before.get('metadata')) and source_after.get('ownedLinks',{}).get(SCOPE+'types/owned')==[link_after],'owned-inline-metadata');need(all(type(x.get('revision'))is str for x in [source_before,link_before,source_after,link_after]),'owned-revision-type');need(link_before['properties']=={'n':0} and link_after['properties']=={'n':1} and source_after['properties']==source_before['properties'],'requested-owned-postimage');need(source_after['revision']!=source_before['revision'] and link_after['revision']!=link_before['revision'] and entry['resource']==link_after and entry['sourceRevision']==source_after['revision'],'owned-revisions');need(type(updated.get('effectPosition'))is str and updated['effectPosition']!=at(18)['requiredPosition'] and updated['requiredPosition']==updated['effectPosition'],'actual-effect');need('effectPosition'not in at(24) and at(24)['results'][1]==entry,'owned-noop');need(at(25)['status']=='failed' and at(25)['problem']['code']=='validation-failed' and at(25)['problem']['operationIndex']==1 and at(25)['problem']['pointer']=='/operations/1/change','atomic-refusal');need(at(26)==source_after and at(27)==link_after,'atomic-restored-postimage');url(28,SCOPE+'beads/tentative');url(29,at(8)['id']);url(30,at(8)['next']);need(at(29)==at(8) and at(30)==at(9) and at(31)==at(8) and at(32)==at(25),'same-store-original-receipts')
 need(o['oracleOutcomes']==[{'id':'O'+str(i),'state':'pass'}for i in range(1,6)],'five-local-clones');exact(o['eventTraversals']['seed'],{'pages':[32,32,32,3],'count':99,'complete':True});exact(o['eventTraversals']['deletion'],{'pages':[32,32,32],'count':96,'complete':True});exact(o['timing'],{'workBudgetMilliseconds':90000,'cleanupBudgetMilliseconds':10000});finite(o['timing']['workMilliseconds'],90000);need(o['timing']['workMilliseconds']<90000,'work-deadline');finite(o['timing']['cleanupMilliseconds'],10000)
 life=o['lifecycle'];keys(life,['pendingRequests','ownershipUnknown','children','directory','descendantOwnership']);exact(life,{'pendingRequests':0,'ownershipUnknown':False,'descendantOwnership':'root-terminal-audit-required'});need(len(life['children'])==2,'two-child-closures')
 for i,c in enumerate(life['children'],1):
  keys(c,['generation','phase','launchAttempted','acquisition','refusal','pid','exitCodeSnapshot','signalCodeSnapshot','observedExit','observedClose','closeRequested','closePromiseSettled','closePromiseSucceeded','closure']);exact(c,{'generation':i,'phase':'initial'if i==1 else'reopen','launchAttempted':True,'acquisition':'acquired','refusal':None,'exitCodeSnapshot':0,'signalCodeSnapshot':None,'observedExit':{'code':0,'signal':None},'observedClose':{'code':0,'signal':None},'closeRequested':True,'closePromiseSettled':True,'closePromiseSucceeded':True,'closure':'observed-clean-exit-and-close'});need(type(c['pid'])is int and c['pid']>0,'actual-child-pid')
 directory=life['directory'];exact(directory,{'disposition':'removed','storeFilePresence':'unknown'});d=Path(directory['path']);need(d.is_absolute() and d.is_relative_to(tmp) and d.name.startswith('bdp-set-installed-') and not d.exists(),'actual-directory-absence');need(type(o['limitations'])is list and len(o['limitations'])==4,'limitations')
 return {'claimEligible':False,'actualExchanges':32,'actualChildren':2,'sourceBindingSha256':g['sourceBinding']['sha256'],'credit':'private D4 slice only; independent root terminal receipt still required'}
def main():
 import sys
 sys.dont_write_bytecode=True
 from gate import load_gate,E,W,document,complete,inventory,sha,write
 mode=sys.argv[1];g,_=load_gate({'tools':'bind-workspace-tools','build':'bind-fresh-build'}[mode],child=True)
 if mode=='build':write('build-artifacts.json',{'head':g['finalHead'],'tree':g['finalTree'],'sourceIdentity':g['_sourceIdentity'],'artifacts':inventory(),**{k:{'path':str(E/('fresh-build.'+s)),'sha256':sha((E/('fresh-build.'+s)).read_bytes())}for k,s in [('start','started.json'),('exit','exit.json'),('log','log')]}})
 else:
  need((E/'node-version.log').read_text().strip()=='v24.16.0' and (E/'pnpm-version.log').read_text().strip()=='11.20.0','actual-tool-versions');tools=dict(document(g['baseTools'])['tools'])
  paths=[W/'node_modules'/n for n in ['@biomejs/biome/bin/biome','vitest/vitest.mjs','typescript/bin/tsc','typescript/lib/tsc.js','typescript/lib/getExePath.js','typescript/package.json']]+[W/'node_modules/.pnpm/@biomejs+cli-darwin-arm64@2.5.7/node_modules/@biomejs/cli-darwin-arm64/biome',W/'node_modules/.pnpm/@typescript+typescript-darwin-arm64@7.0.2/node_modules/@typescript/typescript-darwin-arm64/lib/tsc']
  for name in ['index.UpGiHP7g.js','cli-api.BK8pd4xc.js']:
   f=W/'node_modules/vitest/dist/chunks'/name;need(f.read_bytes()==(Path(__file__).parent/'context/pinned-vitest'/name).read_bytes(),'pinned-reporter-mechanics');paths.append(f)
  for f in paths:tools[str(f)]={'realpath':str(f.resolve(strict=True)),'sha256':sha(f.read_bytes())}
  versions={'@biomejs/biome':'2.5.7','vitest':'4.1.10','typescript':'7.0.2'}
  for name,v in versions.items():need(json.loads((W/'node_modules'/name/'package.json').read_text())['version']==v,'tool-version')
  complete('workspace-install',g);write('workspace-tools.json',{'tools':tools,'versions':versions,'head':g['finalHead'],'sourceIdentity':g['_sourceIdentity'],'sourceLockSha256':g['sourceLockSha256'],'installExitSha256':sha((E/'workspace-install.exit.json').read_bytes()),'installLogSha256':sha((E/'workspace-install.log').read_bytes())})
if __name__=='__main__':main()
