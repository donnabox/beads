"""Closed Scope Event SSE opaque proof auditor. Standard library only; no product imports.
Actual Scope Event SSE opaque report schemas reconciled; root release and final source binding required.
"""
import sys
sys.dont_write_bytecode=True
from gate import *
import re,math
from urllib.parse import urlsplit,parse_qsl,quote
mode=sys.argv[1]
stage={'source':'focused-source','history':'event-restart-external','scope':'scope-event-sse-opaque-external'}.get(mode)
g,_=load_gate({'tools':'bind-workspace-tools','build':'bind-fresh-build','source':'review-focused-source','history':'review-event-restart-external','scope':'review-scope-event-sse-opaque-external'}[mode],child=True)
if mode=='tools':
 tools=dict(read(P/'base-tools.json')['tools'])
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
  require(len(observations)==5 and sorted(o['observation'] for o in observations)==sorted(ex['sourceLabels']),'Expected four new and one historical observations')
  by={o['observation']:o for o in observations}
  a=by['scope-event-sse-minimum-owner'];scalars(a,{'reads':14,'seedCommits':1});rr=rows(a,'cases',14)
  require([r['label'] for r in rr]==['setup','unknown','pos-01','future','foreign','foreign','default','after','invalid-reconnect','bead','link','missing','malformed','known'],'Owner rows')
  require([r['outcome'] for r in rr]==['packet','TransactionMinimumUnproved','TransactionMinimumUnproved','TransactionMinimumFailure','TransactionMinimumUnsupported','TransactionMinimumUnsupported','TypeError','TypeError','TypeError','TypeError','TypeError','missing','TypeError','packet'],'Owner dispositions')
  scalars(rr[0],{'groupObserverActive':False,'groupPrepareCalls':None})
  for r in rr[1:13]:scalars(r,{'groupObserverActive':True,'groupPrepareCalls':0})
  scalars(rr[13],{'groupObserverActive':True});number(rr[13]['groupPrepareCalls']);require(rr[13]['groupPrepareCalls']>0,'Positive SQL preparation control')
  c=rr[0]['context'];context(c);require(c['scopePosition']=='pos-1','Seed current position')
  for i,r in enumerate(rr):require(r.get('context')==(c if i in [0,1,2,3,11,13] else None),'Owner context disclosure')
  b=by['scope-event-sse-minimum-capture'];scalars(b,{'seedCommits':1,'interveningCommits':1,'serviceCalls':5,'ownerAttempts':3,'waits':1,'clockCalls':4,'ownerCloseCalls':1,'closed':True,'failureCount':0})
  rr=rows(b,'attempts',3);base=rr[0]['context'];context(base);later={**base,'scopePosition':'pos-2'}
  require(base['scopePosition']=='pos-1' and [r['context'] for r in rr]==[base,later,later] and b['finalContext']==later,'Per-attempt current context')
  minimum={'scopeEpoch':base['scopeEpoch'],'authorizationView':base['authorizationView'],'minimumPosition':'unknown'}
  require([r['minimum'] for r in rr]==[minimum,minimum,{**minimum,'minimumPosition':'pos-2'}] and [r['outcome'] for r in rr]==['TransactionMinimumUnproved','TransactionMinimumUnproved','packet'],'Captured tuple/outcomes')
  require(rows(b,'probes',5)==[{'call':i+1,'outcome':v} for i,v in enumerate(['TypeError','promise-returned','promise-returned','TypeError','Error'])],'Capture probes')
  c=by['scope-event-sse-minimum-abort'];scalars(c,{'seedCommits':1,'interveningCommits':0,'serviceCalls':2,'ownerAttempts':1,'waits':1,'clockCalls':2,'ownerCloseCalls':1,'rawPostCloseProbes':1,'operationOutcome':'aborted','synchronouslyUnsettled':True,'closed':True,'failureCount':0})
  require(rows(c,'probes',2)==[{'call':1,'outcome':'promise-returned'},{'call':2,'outcome':'Error'}],'Abort probes')
  d=by['scope-event-sse-minimum-http'];scalars(d,{'requests':31,'seedCommits':1,'failureCount':0});rr=rows(d,'records',31)
  require(type(d['elapsedMs']) in [float,int] and math.isfinite(d['elapsedMs']) and d['elapsedMs']>=0,'HTTP elapsed')
  statuses=[200,200,200,200,200,503,503,503,503,503,503,503,503,400,400,503,503,500,500,500,400,500,500,500,500,500,404,401,304,412,200]
  methods=['POST','GET','GET','GET','HEAD','GET','HEAD','GET','HEAD','GET','HEAD','GET','HEAD','GET','HEAD','GET','HEAD','GET','GET','GET','GET','GET','GET','GET','GET','GET','GET','GET','HEAD','HEAD','GET']
  baseline=rows(d,'baselinePages',3);require([p['eventCount'] for p in baseline]==[2,2,1] and baseline[0]['target']=='events/' and baseline[2]['next'] is None,'Actual three page baseline')
  scope='https://event-sse.invalid/acme/';ctx=wire_context(rr[0]['context']);require(ctx[names[2]]=='pos-1','HTTP seed current position')
  for i,p in enumerate(baseline):
   require(p['context']==ctx and rr[i+1]['target']==p['target'],'Baseline response/context linkage')
   if i<2:
    u=urlsplit(p['next']);require(u.scheme=='https' and u.netloc=='event-sse.invalid' and u.path=='/acme/events/' and not u.fragment and len(parse_qsl(u.query))==1 and parse_qsl(u.query)[0][0]=='after','Canonical issued continuation')
    require(baseline[i+1]['target']==p['next'][len(scope):],'Actual issued next dial target')
  positive=d['positive'];scalars(positive,{'ready':True,'status':200,'drained':True});live(positive['liveBeforeClose']);stream_headers(positive['headers'],ctx)
  frames=rows(positive,'framesBeforeClose',5);require(positive['framesAfterDrain']==frames,'Exact frames after drain');events=decode_frames(frames)
  events_shape(events,scope,['beads/a','beads/b','links/l','links/l','links/l'],['created','created','created','linked','linked'])
  require(parse_qsl(urlsplit(baseline[0]['next']).query)==[('after',events[1]['id'])] and parse_qsl(urlsplit(baseline[1]['next']).query)==[('after',events[3]['id'])],'Issued boundaries match stream Events')
  targets=['operations/batch','events/',baseline[1]['target'],baseline[2]['target']]+['events/']*15+['events/?after='+quote(events[0]['id'],safe="~!*'()"),'events/?after=']+['events/']*3+['beads/a?view=events','links/l?view=events','beads/missing?view=events']+['events/']*4
  require(len(targets)==31,'Bound source target inventory')
  for i,r in enumerate(rr):
   require(set(r)=={'ordinal','method','target','status','bodyBytes','context'} and r['ordinal']==i+1 and r['method']==methods[i] and r['status']==statuses[i],'HTTP row shape/order/status');number(r['bodyBytes'])
   require(r['target']==targets[i],'HTTP route sequence')
   require(r['context']==(ctx if r['status'] in [200,503,404,304,412] else {}),'HTTP current/refusal context')
   if r['method']=='HEAD':require(r['bodyBytes']==0,'HEAD body')
  h=by['finite-event-minimum-http'];scalars(h,{'calls':32,'productChildren':0,'directoryAbsent':True,'failureCount':0})
  report.update({'actualInstrumentedObservations':observations,'newObservationRecords':4,'historicalObservationRecords':1,'prospectiveWorkload':ex['workload'],'limitations':'Eight passed callbacks; only four new and one existing report are counted as runtime instrumentation. Remaining legacy request/commit counts are source-derived. SQL preparations are not executions/allocation counts; row1 is explicitly unobserved. Reported frames are complete Event identities; baseline full-body/schema/conditional/capacity assertions remain in pinned callbacks. Immediate closed-promise flag is only a synchronous observation; completed close/drain is separately required.'})
 else:
  require(len(observations)==1,'One installed report expected');o=observations[0];is_history=mode=='history';expected_label='private-event-restart-cursor' if is_history else 'private-installed-scope-event-sse-opaque-minimum'
  scalars(o,{'observation':expected_label,'attribution':'external-installed-candidate','installedProvenance':'requires separately verified external package/consumer/source binding','claimEligible':False,'outcome':'passed','declaredSourceCommit':g['finalHead'],'node':'v24.16.0','pendingOperationsAtReport':0,'directoryRemoved':True,'failures':[]})
  witness='apps/bdptest/src/transactional-event-sse-installed.test.ts'
  require(o['witnessSource']=={'fingerprintScope':'witness file bytes and tracked working-tree diff only; excludes untracked inputs and installed payload provenance','head':g['finalHead'],'fileSha256':g['_sourceBinding']['trackedSourceFiles'][witness],'diffSha256':sha(git('diff','HEAD')),'dirtyPaths':git('diff','--name-only','HEAD').decode().splitlines()},'Witness attribution changed')
  require(o['deadlines']=={'work':45000,'cleanup':10000,'callback':60000,'outerSupervisor':120000,'finite':5000,'recorder':5000},'Deadlines changed')
  scope='https://reference.invalid/acme/';b=o['binding'];served=b['served']
  require(served['sourceCommit']==g['finalHead'] and served['profile']=='transactional' and served['history'] is False and served['scope']==scope and b['authorityMode']=='canonical','Runtime binding')
  for key in ['installedPayloadDigest','installedContractsDigest','policyDigest','authenticationDigest']:fingerprint(served[key])
  fingerprint(b['runtimeConfigurationIdentity']);fixture=W/'packages/conformance/test-support/transactional/fixture-v1.json';fj=read(fixture)
  require(served['fixtureDigest']==sha(fixture.read_bytes()) and served['fixtureId']==fj['id'] and served['capabilities']==sorted(fj['capabilities']),'Fixture binding')
  require(b['operationTargets']=={n:scope+'operations/'+s for n,s in ex['operationTargets'].items()},'Operation target binding')
  proof=verify_consumer(g);require(Path(o['entry']).resolve(strict=True)==Path(proof['entryRealpath']) and o['entrySha256']==proof['entrySha256'] and served['installedPayloadDigest']==proof['payloadIdentity'],'External entry/payload differs')
  subjects=['beads/a','beads/b'] if is_history else ['beads/a'];seed=json.dumps({'operations':[{'operation':'createBead','id':s,'type':scope+'types/task','properties':{'n':0}} for s in subjects]},separators=(',',':')).encode()
  scalars(o,{'seedKey':'event-restart-two-created-v1' if is_history else 'scope-sse-minimum-one-created-v1','seedBodyBytes':len(seed),'seedBodySha256':sha(seed)})
  events=rows(o,'events' if is_history else 'baselineEvents',len(subjects));events_shape(events,scope,subjects,['created']*len(subjects))
  ss=rows(o,'sessions',2 if is_history else 1);require(len(set(s['pid'] for s in ss))==len(ss),'Distinct child PIDs')
  for i,s in enumerate(ss):
   scalars(s,{'ordinal':i+1,'create':i==0,'state':'owned','closed':True,'stopUnknown':False,'exit':{'code':0,'signal':None}})
   require(type(s['pid']) is int and s['pid'] in execution['sampledPids'] and s['binding']==b,'Actual sampled child/binding')
   u=urlsplit(s['listenUrl']);require(u.scheme=='http' and u.hostname in ['127.0.0.1','localhost','::1'] and u.port is not None,'Actual local child listener')
   if not is_history:require(s['closeEvent']=={'code':0,'signal':None},'Actual child stdio close')
  requests=rows(o,'requests',5 if is_history else 6)
  expected_targets=['operations/batch','events/','events/?after='+quote(events[0]['id'],safe="~!*'()"),'events/?after='+quote(events[1]['id'],safe="~!*'()"),'beads/a?view=events&after='+quote(events[0]['id'],safe="~!*'()")] if is_history else ['operations/batch']+['events/']*5
  expected_methods=['POST','GET','GET','GET','GET'] if is_history else ['POST','GET','HEAD','GET','HEAD','GET']
  for i,r in enumerate(requests):require(r=={'ordinal':i+1,'session':1 if not is_history or i<2 else 2,'method':expected_methods[i],'target':expected_targets[i]},'Exact installed request sequence')
  responses=rows(o,'completedFiniteResponses',4 if is_history else 5);ordinals=[1,2,3,5] if is_history else [1,2,3,4,5];statuses=[200,200,200,400] if is_history else [200,200,200,503,503]
  ctx=wire_context(responses[0]['headers']);require(ctx[names[2]]=='pos-1','Installed seed current position')
  allowed_headers=set(names+['content-type','x-bdp-development']+([] if is_history else ['content-length','cache-control','retry-after']))
  for r,ordinal,status in zip(responses,ordinals,statuses):
   require(set(r)=={'ordinal','session','method','target','status','bodyOctets','bodySha256','wireSha256','headers'} and {k:r[k] for k in ['ordinal','session','method','target']}==requests[ordinal-1] and r['status']==status,'Finite exact response sequence/shape')
   number(r['bodyOctets']);fingerprint(r['bodySha256']);fingerprint(r['wireSha256']);h=r['headers'];require(set(h)<=allowed_headers and all(isinstance(v,str) for v in h.values()) and wire_context(h)==ctx and h['x-bdp-development']=='non-attesting','Finite safe headers/current context')
   if r['method']=='HEAD':require(r['bodyOctets']==0 and r['bodySha256']==sha(b''),'Finite HEAD zero body')
  sse=o['sse'];scalars(sse,{'requestOrdinal':4 if is_history else 6,'status':200,'drained':True});stream_headers(sse['headers'],ctx)
  frames=rows(sse,'completeObservedFrames',1);require(decode_frames(frames)==[events[-1]],'Complete SSE Event equality')
  if is_history:
   scalars(sse,{'session':2,'target':expected_targets[3],'lastEventId':events[0]['id']});require(type(sse['beforeIntentionalCloseMs']) in [float,int] and 0<=sse['beforeIntentionalCloseMs']<5000,'Historical live boundary')
   require(o['localSensitivityControls']==['restored-event-ordinal','frame-event-identity'],'Historical controls')
   require([r['headers']['content-type'] for r in responses]==['application/json']*3+['application/problem+json'],'Historical finite media')
   report['lifecycleLimit']='Two actual sampled children, reported zero exits, clean shutdown before reopen asserted by pinned callback. Historical report has no separate child closeEvent inventory. No new stdio-drain claim is applied retrospectively.'
  else:
   live(sse['liveBeforeClose']);require(sse['framesBeforeClose']==frames and sse['headers']['x-bdp-development']=='non-attesting','Live exact before/after frames')
   number(sse['wireBytes']);require(sse['wireBytes']>0,'Observed SSE wire bytes');fingerprint(sse['wireSha256'])
   require(o['localSensitivityControls']==['frame-event-identity'],'Frame negative control')
   require(responses[0]['headers']['content-type']=='application/json' and responses[1]['headers']['content-type']=='application/json' and responses[2]['headers']['content-type']=='text/event-stream' and 'content-length' not in responses[2]['headers'],'Known finite and HEAD media')
   for r in responses[3:]:
    h=r['headers'];require(h['content-type']=='application/problem+json' and h['cache-control']=='private, no-store' and h['retry-after']=='1','Opaque refusal metadata')
   require(responses[3]['bodyOctets']>0 and responses[3]['headers']['content-length']==str(responses[3]['bodyOctets']) and responses[4]['headers']['content-length']==responses[3]['headers']['content-length'],'Opaque GET/HEAD length parity')
   report['lifecycleLimit']='One reported child exit and closeEvent plus sampled process absence and owned directory removal; one ten-second cleanup budget in pinned callback. Promise settlement is not a socket-close census; unknown/pre-ready/deadline cleanup is failure, never late success.'
  at=o['streamAtReport'];scalars(at,{'status':200,'drained':True,'destroyed':True});stream_headers(at['headers'],ctx);require(at['completeObservedFrames']==frames,'Exact final drained frames');number(at['wireBytes']);require(at['wireBytes']>0,'Reported transport bytes');fingerprint(at['wireSha256'])
  if not is_history:require(at['wireSha256']==sse['wireSha256'] and at['wireBytes']==sse['wireBytes'],'Post-drain wire identity')
  directory_absent(o['directory'],'bdp-event-restart-installed-' if is_history else 'bdp-scope-sse-minimum-installed-')
  report.update({'observation':o,'attribution':'Reviewed dirty source -> fresh build -> archives -> independently verified external consumer -> actual child payload. Complete body/schema/authority/configuration assertions stay in pinned passing callbacks; no public qualification or source attestation.'})
 write(stage+'.accepted-observation.json',report)
