"""Closed Event admission/replay proof auditor. Standard library only; no product imports.
Actual Event admission/replay report schemas reconciled; root release and final source binding required.
"""
import sys
sys.dont_write_bytecode=True
from gate import *
import re,math
from urllib.parse import urlsplit,parse_qsl,quote
mode=sys.argv[1]
stage={'source':'focused-source','history':'historical-event-restart-external','replay':'current-event-admission-replay-external'}.get(mode)
g,_=load_gate({'tools':'bind-workspace-tools','build':'bind-fresh-build','source':'review-focused-source','history':'review-historical-event-restart-external','replay':'review-current-event-admission-replay-external'}[mode],child=True)
if mode=='tools':
 tools=dict(document(g['baseTools'])['tools'])
 paths=[W/'node_modules'/n for n in ['@biomejs/biome/bin/biome','vitest/vitest.mjs','typescript/bin/tsc','typescript/lib/tsc.js','typescript/lib/getExePath.js','typescript/package.json']]
 paths += [W/'node_modules/.pnpm/@biomejs+cli-darwin-arm64@2.5.7/node_modules/@biomejs/cli-darwin-arm64/biome',W/'node_modules/.pnpm/@typescript+typescript-darwin-arm64@7.0.2/node_modules/@typescript/typescript-darwin-arm64/lib/tsc']
 for p in paths:tools[str(p)]={'realpath':str(p.resolve(strict=True)),'sha256':sha(p.read_bytes())}
 versions={'@biomejs/biome':'2.5.7','vitest':'4.1.10','typescript':'7.0.2'}
 for name,version in versions.items():require(read(W/'node_modules'/name/'package.json')['version']==version,'Installed tool version differs')
 complete('workspace-install',g)
 write('workspace-tools.json',{'tools':tools,'versions':versions,'head':g['finalHead'],'sourceIdentity':g['_sourceIdentity'],'sourceLockSha256':g['sourceLockSha256'],'installExitSha256':sha((E/'workspace-install.exit.json').read_bytes()),'installLogSha256':sha((E/'workspace-install.log').read_bytes())})
elif mode=='build':
 write('build-artifacts.json',{'head':g['finalHead'],'tree':g['finalTree'],'sourceIdentity':g['_sourceIdentity'],'artifacts':inventory(),**{key:{'path':str(E/('fresh-build.'+suffix)),'sha256':sha((E/('fresh-build.'+suffix)).read_bytes())} for key,suffix in [('start','started.json'),('exit','exit.json'),('log','log')]}})
else:
 config=read(P/'commands.json')['commands'][stage];result=read(E/(stage+'.results.json'));ex=read(P/'audit-expectations.json')
 require(ex['finalized'] is True and sha((P/'report-schemas.json').read_bytes())==ex['reportSchemasSha256'],'Report contracts changed')
 chosen=[]
 for suite in result['testResults']:
  for case in suite['assertionResults']:
   if case['fullName'] in config['selectedFullNames']:
    require(Path(suite['name']).resolve()==(W/config['selectedFilesByFullName'][case['fullName']]).resolve(),'Selected callback file differs')
    chosen.append(case)
   else:require(case['status'] in ['pending','skipped','todo'],'Unexpected callback executed')
 require(len(chosen)==len(config['selectedFullNames']) and sorted(x['fullName'] for x in chosen)==sorted(config['selectedFullNames']) and all(x['status']=='passed' for x in chosen) and result.get('success') is True,'Exact SSE callbacks did not pass')
 execution=complete(stage,g);observations=[]
 for line in (E/(stage+'.log')).read_text().splitlines():
  try:x=json.loads(line.strip())
  except json.JSONDecodeError:continue
  if isinstance(x,dict) and 'observation' in x:observations.append(x)
 report={'sourceIdentity':g['_sourceIdentity'],'sourceBindingSha256':g['sourceBinding']['sha256'],'qualification':False,'selectedCallbacks':config['selectedFullNames'],'resultsSha256':sha((E/(stage+'.results.json')).read_bytes()),'logSha256':sha((E/(stage+'.log')).read_bytes())}
 def fingerprint(x):require(isinstance(x,str) and re.fullmatch('[a-f0-9]{64}',x),'Invalid SHA256')
 def number(x):require(type(x) is int and x>=0,'Invalid nonnegative counter')
 def scalars(o,expected):require(all(k in o and type(o[k]) is type(v) and o[k]==v for k,v in expected.items()),'Report scalar mismatch '+str(expected))
 def rows(o,key,n):
  value=o.get(key);require(isinstance(value,list) and len(value)==n,'Report array count '+key);return value
 def context(c):
  require(isinstance(c,dict) and set(c)=={'scopeEpoch','authorizationView','scopePosition'},'Owner context shape')
  require(all(isinstance(v,str) and re.fullmatch('[A-Za-z0-9_-]{1,256}',v) for v in c.values()),'Owner context token')
 names=['bdp-scope-epoch','bdp-authorization-view','bdp-scope-position']
 def wire_context(h):
  c={k:h[k] for k in names};require(all(isinstance(v,str) and re.fullmatch('[A-Za-z0-9_-]{1,256}',v) for v in c.values()),'Wire context token');return c
 def live(x):
  scalars(x,{'aborted':False,'destroyed':False});require(type(x['elapsedMs']) in [float,int] and 0<=x['elapsedMs']<5000,'Pre-close elapsed')
 def stream_headers(h,c):
  require(h['content-type']=='text/event-stream' and h['cache-control']=='no-store, no-transform' and wire_context(h)==c and 'etag' not in h and 'content-length' not in h,'SSE media/context/no finite validators')
 def decode_frames(frames):
  values=[]
  for frame in frames:
   require(isinstance(frame,str),'Frame text');lines=frame.split('\n');data=[x[6:] for x in lines if x.startswith('data: ')]
   require(len(data)==1,'Single full Event data line');event=json.loads(data[0]);require(isinstance(event,dict),'Event object')
   require([x[4:] for x in lines if x.startswith('id: ')]==[event['id']] and [x[7:] for x in lines if x.startswith('event: ')]==[event['type']],'Frame fields do not identify full Event')
   values.append(event)
  return values
 def events_shape(events,scope,subjects,types):
  require(len(events)==len(subjects) and len(set(e['id'] for e in events))==len(events),'Distinct complete Events')
  for i,e in enumerate(events):
   require(e['ordinal']==i and e['type']==types[i] and e['source']==scope+'events/' and e['subject']==scope+subjects[i],'Event sequence/source/subject')
   require(isinstance(e['id'],str) and e['id'] and isinstance(e['transaction'],str) and e['transaction'] and isinstance(e['time'],str) and isinstance(e['data'],dict),'Event field shape')
  require(len(set(e['transaction'] for e in events))==1,'Events from one real commit')
 def directory_absent(value,prefix):
  tmp=Path(read(E/(stage+'.started.json'))['temporaryDirectory']);d=Path(value)
  require(tmp.parent==Path('/private/tmp') and d.is_absolute() and d.is_relative_to(tmp) and d.name.startswith(prefix) and not d.exists(),'Store ownership/absence')
 if mode=='source':
  from collections import Counter
  require(Counter(o['observation'] for o in observations)==Counter(ex['sourceLabels']),'Exact12new/fourlegacy reports')
  grouped={label:[o for o in observations if o['observation']==label] for label in set(ex['sourceLabels'])}
  owners=grouped['event-replay-minimum-owner'];require(sorted(o['sourceKind'] for o in owners)==['bead','link','scope'],'Owner source coverage')
  ownerLabels=['initial-opaque','issued-opaque','stale-opaque','both-foreign','foreign-view','issued-known','wrong-source','strict-after','malformed','missing']
  outcomes=['TransactionMinimumUnproved']*3+['TransactionMinimumFailure']*2+['success','TransactionEventParameterError','TypeError','TypeError','missing']
  for o in owners:
   scalars(o,{'calls':22,'baselineCalls':2,'baselineGroupPreparations':'unobserved','seedCommits':1,'groupObservation':'matching SQL preparations, not executions or allocation census'})
   rr=rows(o,'rows',20);c=rr[0]['context'];context(c);require(c['scopePosition']=='pos-1','Owner seed context')
   for i,r in enumerate(rr):
    k=i//2;require(r['label']==ownerLabels[k] and r['method']==['finite','sse'][i%2] and r['outcome']==outcomes[k],'Owner ordered method/outcome')
    require(r['intent']==(None if k==7 else 'current-event'),'Owner intent');require(r.get('context')==(None if k in [7,8] else c),'Owner context disclosure')
    require(r['reason']==('cursor-expired' if k==3 else 'foreign-view' if k==4 else None),'Owner refusal reason')
    number(r['groupPrepareCalls']);require((r['groupPrepareCalls']>0) if k==5 else r['groupPrepareCalls']==0,'Retained preparation oracle')
    require(r['source']==(o['source'] if k!=9 else o['source'].split('/acme/')[0]+'/acme/'+('links' if o['sourceKind']=='link' else 'beads')+'/missing?view=events'),'Owner source/absence route')
   require(rr[0]['after'] is None and rr[1]['after'] is None,'Initial after absent')
  captures=grouped['event-replay-minimum-capture'];require(sorted(o['label'] for o in captures)==sorted(['scope-json','bead-json','link-json','scope-sse','bead-sse','link-sse']),'Capture transport/source coverage')
  for o in captures:
   scalars(o,{'serviceCalls':1,'ownerAttempts':2,'baselineCalls':1,'waits':1,'clockCalls':3,'commits':2,'ownerCloseCalls':1,'closed':True,'failureCount':0})
   rr=rows(o,'attempts',2);c=rr[0]['context'];context(c);later={**c,'scopePosition':'pos-2'}
   require(c['scopePosition']=='pos-1' and rr[1]['context']==later,'Fresh retry context');minimum={'scopeEpoch':c['scopeEpoch'],'authorizationView':c['authorizationView'],'minimumPosition':'unknown'}
   require(all(r['minimum']==minimum and r['after']==rr[0]['after'] and r['intent']=='current-event' and r['outcome']=='TransactionMinimumUnproved' and r['method']==('finite' if o['label'].endswith('-json') else 'sse') for r in rr),'Captured immutable tuple/after/intent')
  drains=grouped['event-replay-minimum-drain'];require(sorted(o['transport'] for o in drains)==['json','sse'],'Drain transports')
  for o in drains:
   scalars(o,{'serviceCalls':1,'ownerAttempts':1,'baselineCalls':1,'waits':1,'clockCalls':2,'commits':1,'ownerCloseCalls':1,'postCloseProbes':1,'synchronouslyUnsettled':True,'closed':True,'failureCount':0})
   r=rows(o,'attempts',1)[0];context(r['context']);require(r['context']['scopePosition']=='pos-1' and r['intent']=='current-event' and r['outcome']=='TransactionMinimumUnproved' and r['method']==('finite' if o['transport']=='json' else 'sse'),'Drained actual admission')
  d=grouped['event-replay-minimum-http'][0];scalars(d,{'requests':37,'seedCommits':1,'closed':True,'directoryRemoved':True,'failureCount':0});require(type(d['elapsedMs']) in [int,float] and math.isfinite(d['elapsedMs']) and 0<=d['elapsedMs']<4000,'Cleanup-inclusive HTTP margin')
  directory_absent(d['directory'],'bdp-event-replay-http-');rr=rows(d,'records',37);expected=ex['workload']['D']['orderedRequests'];ctx=wire_context(rr[0]['context']);require(ctx[names[2]]=='pos-1','HTTP actual current seed')
  for i,(r,want) in enumerate(zip(rr,expected)):
   require(set(r)=={'ordinal','method','target','status','bodyBytes','context'} and r['ordinal']==i+1 and r['method']==want['method'] and r['status']==want['status'],'HTTP exact ordinal/method/status');number(r['bodyBytes']);require(r['context']==({} if i==26 else ctx),'Receiving vs observed context')
   if r['method']=='HEAD' or r['status']==406:require(r['bodyBytes']==0,'HEAD or existing empty406 body')
  scope='https://event-sse.invalid/acme/'
  targets=['operations/batch','events/',rr[2]['target'],rr[3]['target'],'beads/a?view=events','links/l?view=events']
  for idx,route in [(6,'events/'),(12,'beads/a?view=events'),(18,'links/l?view=events')]:
   suffix=rr[idx]['target'];u=urlsplit(scope+suffix);q=parse_qsl(u.query);require(u.path=='/acme/'+route.split('?')[0] and [k for k,v in q]==([] if '?' not in route else ['view'])+['after'] and q[-1][1],'Actual issued-after target')
   targets += [suffix,suffix,route,route,route,suffix]
  for index in [2,3]:
   u=urlsplit(scope+rr[index]['target']);q=parse_qsl(u.query);require(u.path=='/acme/events/' and len(q)==1 and q[0][0]=='after' and q[0][1],'Actual Scope continuation target')
  targets += ['beads/a?view=events','links/l?view=events','events/?after=','beads/missing?view=events']+['events/']*7+[rr[35]['target'],'links/l?view=events']
  require(len(targets)==37 and [r['target'] for r in rr]==targets,'Exact HTTP route sequence')
  positives=rows(d,'positives',2)
  for i,p in enumerate(positives):
   scalars(p,{'ready':True,'status':200,'drained':True});scalars(p['liveBeforeClose'],{'aborted':False,'destroyed':False});stream_headers(p['headers'],ctx)
   frames=rows(p,'framesBeforeClose',1);require(p['framesAfterDrain']==frames and p['target']==rr[35+i]['target'],'Exact positive target/frames after drain');ev=decode_frames(frames)[0]
   require(ev['source']==scope+('beads/a?view=events' if i==0 else 'links/l?view=events') and ev['subject']==scope+'links/l' and ev['type']==('linked' if i==0 else 'created'),'Positive source-bound Event')
  # Existing reports retain their original identity and actual field shape.
  for label,count in [('finite-event-minimum-http',32),('bead-event-minimum-http',37),('link-event-minimum-http',35)]:
   o=grouped[label][0];scalars(o,{'calls':count,'productChildren':0,'directoryAbsent':True,'failureCount':0})
   if label!='finite-event-minimum-http':
    records=rows(o,'completedResponses',count);require(sum(r['status']==503 for r in records)==(18 if label.startswith('bead') else 16),'Legacy converted503 aggregate');require(not any(r['status']==500 for r in records) if label.startswith('link') else True,'Legacy Link500 aggregate')
  legacy=grouped['scope-event-sse-minimum-http'][0];scalars(legacy,{'requests':31,'seedCommits':1,'failureCount':0});records=rows(legacy,'records',31);require(dict(Counter(str(r['status']) for r in records))=={'200':6,'503':16,'400':3,'404':1,'401':1,'304':1,'412':1,'409':1,'410':1},'Legacy Scope SSE status distribution')
  report.update({'actualInstrumentedObservations':observations,'newObservationRecords':12,'legacyObservationRecords':4,'prospectiveWorkload':ex['workload'],'limitations':'Exact17 passed source callbacks and16 actual reports; unreported legacy reconnect28 requests remain source-derived, not reconstructed telemetry. SQL preparation counts are not executions/allocation census. Full receipt/page/Problem schema and request-control assertions remain in pinned passing callbacks. Recorder stored errors are not exposed by close; actual positive live/complete-frame and terminal body assertions are bounded proof, not socket census.'})
 else:
  require(len(observations)==1,'One installed report');o=observations[0];history=mode=='history'
  scalars(o,{'observation':'private-event-restart-cursor' if history else 'private-installed-event-admission-replay-minimum','attribution':'external-installed-candidate','installedProvenance':'requires separately verified external package/consumer/source binding','claimEligible':False,'outcome':'passed','declaredSourceCommit':g['finalHead'],'node':'v24.16.0','pendingOperationsAtReport':0,'directoryRemoved':True,'failures':[]})
  witness='apps/bdptest/src/transactional-event-sse-installed.test.ts';require(o['witnessSource']=={'fingerprintScope':'witness file bytes and tracked working-tree diff only; excludes untracked inputs and installed payload provenance','head':g['finalHead'],'fileSha256':g['_sourceBinding']['trackedSourceFiles'][witness],'diffSha256':sha(git('diff','HEAD')),'dirtyPaths':git('diff','--name-only','HEAD').decode().splitlines()},'Witness actual source identity')
  require(o['deadlines']=={'work':45000,'cleanup':10000,'callback':60000,'outerSupervisor':120000,'finite':5000,'recorder':5000},'Installed bounds')
  scope='https://reference.invalid/acme/';binding=o['binding'];served=binding['served'];require(served['sourceCommit']==g['finalHead'] and served['profile']=='transactional' and served['history'] is False and served['scope']==scope and binding['authorityMode']=='canonical','Actual runtime binding')
  for key in ['installedPayloadDigest','installedContractsDigest','policyDigest','authenticationDigest']:fingerprint(served[key])
  fingerprint(binding['runtimeConfigurationIdentity']);fixture=W/'packages/conformance/test-support/transactional/fixture-v1.json';fj=read(fixture)
  require(served['fixtureDigest']==sha(fixture.read_bytes()) and served['fixtureId']==fj['id'] and served['capabilities']==sorted(fj['capabilities']),'Fixture binding');require(binding['operationTargets']=={n:scope+'operations/'+v for n,v in ex['operationTargets'].items()},'Operation target binding')
  proof=verify_consumer(g)
  require(Path(o['entry']).resolve(strict=True)==Path(proof['entryRealpath']) and o['entrySha256']==proof['entrySha256'] and served['installedPayloadDigest']==proof['payloadIdentity'],'External entry/payload')
  if not history:require(o['entryLiteral']==proof['entryRealpath'],'External entry literal')
  operations=[{'operation':'createBead','id':'beads/'+id,'type':scope+'types/task','properties':{'n':0}} for id in ['a','b']]
  if not history:operations.append({'operation':'createLink','id':'links/l','type':scope+'types/link','source':'beads/a','target':'beads/b','properties':{'n':0}})
  seed=json.dumps({'operations':operations},separators=(',',':')).encode();scalars(o,{'seedKey':'event-restart-two-created-v1' if history else 'event-replay-minimum-three-operations-v1','seedBodyBytes':len(seed),'seedBodySha256':sha(seed)})
  ev=rows(o,'events' if history else 'baselineEvents',2 if history else 8)
  if history:events_shape(ev,scope,['beads/a','beads/b'],['created','created'])
  else:
   events_shape(ev[:5],scope,['beads/a','beads/b','links/l','links/l','links/l'],['created','created','created','linked','linked'])
   require([e['source'] for e in ev[5:]]==[scope+'beads/a?view=events']*2+[scope+'links/l?view=events'] and [e['subject'] for e in ev[5:]]==[scope+'beads/a',scope+'links/l',scope+'links/l'] and [e['type'] for e in ev[5:]]==['created','linked','created'],'Genuine per-source baseline Events')
   require(all(e['transaction']==ev[0]['transaction'] for e in ev),'Single seed transaction')
  sessions=rows(o,'sessions',2 if history else 1);require(len(set(s['pid'] for s in sessions))==len(sessions),'Distinct child PIDs')
  for i,s in enumerate(sessions):
   scalars(s,{'ordinal':i+1,'create':i==0,'state':'owned','closed':True,'stopUnknown':False,'exit':{'code':0,'signal':None}});require(type(s['pid']) is int and s['pid'] in execution['sampledPids'] and s['binding']==binding,'Actual sampled child/binding')
   u=urlsplit(s['listenUrl']);require(u.scheme=='http' and u.hostname in ['127.0.0.1','localhost','::1'] and u.port is not None,'Actual child listener')
   if not history:require(s['closeEvent']=={'code':0,'signal':None},'Actual child closeEvent')
  requests=rows(o,'requests',5 if history else 10)
  if history:targets=['operations/batch','events/','events/?after='+quote(ev[0]['id'],safe="~!*'()"),'events/?after='+quote(ev[1]['id'],safe="~!*'()"),'beads/a?view=events&after='+quote(ev[0]['id'],safe="~!*'()")]
  else:targets=['operations/batch','events/','beads/a?view=events','links/l?view=events','beads/a?view=events','links/l?view=events','events/?after='+quote(ev[0]['id'],safe="~!*'()"),'links/l?view=events','beads/a?view=events','beads/a?view=events&after='+quote(ev[0]['id'],safe="~!*'()")]
  for i,r in enumerate(requests):require(r=={'ordinal':i+1,'session':2 if history and i>=2 else 1,'method':'POST' if i==0 else 'GET','target':targets[i]},'Exact installed request order')
  responses=rows(o,'completedFiniteResponses',4 if history else 9);ordinals=[1,2,3,5] if history else list(range(1,10));statuses=[200,200,200,400] if history else [200]*4+[503]*5;ctx=wire_context(responses[0]['headers']);require(ctx[names[2]]=='pos-1','Installed seed context')
  for r,ordinal,status in zip(responses,ordinals,statuses):
   require({k:r[k] for k in ['ordinal','session','method','target']}==requests[ordinal-1] and r['status']==status,'Exact installed terminal record');number(r['bodyOctets']);fingerprint(r['bodySha256']);fingerprint(r['wireSha256']);h=r['headers'];require(wire_context(h)==ctx and h['x-bdp-development']=='non-attesting','Terminal current context/marker')
   require(h['content-type']==('application/problem+json' if status>=400 else 'application/json'),'Terminal media')
   if not history and ordinal>=5:
    require(r['bodyOctets']>0 and h['content-length']==str(r['bodyOctets']) and h['retry-after']=='1' and h['cache-control']=='private, no-store' and 'etag' not in h and 'x-bdp-development-limit' not in h,'Opaque terminal headers');scalars(r,{'completeObservedFrames':[],'drained':True})
  sse=o['sse'];scalars(sse,{'requestOrdinal':4 if history else 10,'status':200,'drained':True});stream_headers(sse['headers'],ctx);frames=rows(sse,'completeObservedFrames',1);require(decode_frames(frames)==[ev[1] if history else ev[6]],'Exact remaining Event frame')
  if history:
   scalars(sse,{'session':2,'target':targets[3],'lastEventId':ev[0]['id']});require(type(sse['beforeIntentionalCloseMs']) in [int,float] and 0<=sse['beforeIntentionalCloseMs']<5000,'Historical live boundary');require(o['localSensitivityControls']==['restored-event-ordinal','frame-event-identity'],'Historical controls')
  else:
   live(sse['liveBeforeClose']);require(sse['framesBeforeClose']==frames and sse['headers']['x-bdp-development']=='non-attesting','Live pre/post drain frames');require(o['localSensitivityControls']==['frame-event-identity'],'Negative frame control');number(sse['wireBytes']);fingerprint(sse['wireSha256'])
  at=o['streamAtReport'];scalars(at,{'status':200,'drained':True,'destroyed':True});stream_headers(at['headers'],ctx);require(at['completeObservedFrames']==frames,'Final drained frames');number(at['wireBytes']);require(at['wireBytes']>0,'Observed stream wire');fingerprint(at['wireSha256'])
  if not history:require(at['wireSha256']==sse['wireSha256'] and at['wireBytes']==sse['wireBytes'],'Final transport identity')
  directory_absent(o['directory'],'bdp-event-restart-installed-' if history else 'bdp-event-replay-minimum-installed-')
  report.update({'observation':o,'attribution':'Reviewed dirty source/fresh build/archive/root-accepted external consumer/actual child payload; no public qualification.','lifecycleLimit':'Historical restart has two sampled children/exit records but no added closeEvent claim. New callback requires one actual exit+closeEvent/drain and directory absence. Promise counts are not socket census. Raw body/schema and exact request sensitivity assertions remain in pinned passing callbacks; no hidden recorder error visibility is claimed.'})
 write(stage+'.accepted-observation.json',report)
