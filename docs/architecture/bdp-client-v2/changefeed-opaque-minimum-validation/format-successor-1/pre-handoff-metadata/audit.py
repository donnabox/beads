"""Closed Changefeed opaque proof auditor. Standard library only; no product imports.
Actual Changefeed opaque report schemas reconciled; root release and final source binding required.
"""
import sys
sys.dont_write_bytecode=True
from gate import *
import re,math
from urllib.parse import urlsplit,parse_qsl
mode=sys.argv[1]
stage={'source':'focused-source','history':'changefeed-history-external','changefeed':'changefeed-opaque-external'}.get(mode)
g,_=load_gate({'tools':'bind-workspace-tools','build':'bind-fresh-build','source':'review-focused-source','history':'review-changefeed-history-external','changefeed':'review-changefeed-opaque-external'}[mode],child=True)
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
 require(ex['finalized'] is True,'Changefeed contracts not finalized')
 cases=[x for suite in result['testResults'] for x in suite['assertionResults']]
 chosen=[x for x in cases if x['fullName'] in config['selectedFullNames']]
 require(len(chosen)==len(config['selectedFullNames']) and sorted(x['fullName'] for x in chosen)==sorted(config['selectedFullNames']) and all(x['status']=='passed' for x in chosen) and result.get('success') is True,'Exact changefeed callbacks did not pass')
 require(all(x['status'] in ['pending','skipped','todo'] for x in cases if x not in chosen),'Unexpected callback executed')
 execution=complete(stage,g);observations=[]
 for line in (E/(stage+'.log')).read_text().splitlines():
  try:x=json.loads(line.strip())
  except json.JSONDecodeError:continue
  if isinstance(x,dict) and 'observation' in x:observations.append(x)
 report={'sourceIdentity':g['_sourceIdentity'],'sourceBindingSha256':g['sourceBinding']['sha256'],'qualification':False,'selectedCallbacks':config['selectedFullNames'],'resultsSha256':sha((E/(stage+'.results.json')).read_bytes()),'logSha256':sha((E/(stage+'.log')).read_bytes())}
 def fingerprint(x):require(isinstance(x,str) and re.fullmatch('[a-f0-9]{64}',x),'Invalid SHA256 field')
 def number(x):require(type(x) is int and x>=0,'Invalid octet/counter')
 def context(x):
  require(isinstance(x,dict) and set(x)=={'scopeEpoch','authorizationView','scopePosition'},'Owner context shape')
  require(all(isinstance(y,str) and re.fullmatch('[A-Za-z0-9_-]{1,256}',y) for y in x.values()),'Owner context token')
 def scalars(o,expected):require(all(k in o and type(o[k]) is type(v) and o[k]==v for k,v in expected.items()),'Report scalar mismatch '+str(expected))
 def rows(o,key,n):
  value=o.get(key);require(isinstance(value,list) and len(value)==n,'Report array count '+key);return value
 names=['bdp-scope-epoch','bdp-authorization-view','bdp-scope-position']
 def header_context(h):return [h[n] for n in names]
 def seed_context(value):require(isinstance(value,list) and len(value)==3 and all(isinstance(v,str) and re.fullmatch('[A-Za-z0-9_-]{1,256}',v) for v in value) and value[2]=='pos-1','Actual seed context')
 def directory_absent(value,prefix):
  tmp=Path(read(E/(stage+'.started.json'))['temporaryDirectory']);d=Path(value)
  require(tmp.parent==Path('/private/tmp') and d.is_absolute() and d.is_relative_to(tmp) and d.name.startswith(prefix) and not d.exists(),'Store ownership/absence')
 def wire_rows(rr,spec,scope):
  for i,x in enumerate(rr):
   require(set(x)==set(spec['fields']) and x['method']==spec['methods'][i] and x['status']==spec['statuses'][i],'Response order/shape')
   if 'ordinal' in x:require(x['ordinal']==i+1,'Response ordinal')
   if 'suffixes' in spec:require(x['url']==scope+spec['suffixes'][i],'Response URL')
   number(x['bodyOctets'])
   for key in ['bodySha256','wireSha256']:
    if key in x:fingerprint(x[key])
   if 'headers' in spec:
    hh=x['headers'];require(set(hh)==set(spec['headers']) and all(v is None or isinstance(v,str) for v in hh.values()),'Safe response headers')
   if x['method']=='HEAD':
    require(x['bodyOctets']==0,'HEAD body nonempty')
    if 'bodySha256' in x:require(x['bodySha256']==sha(b''),'HEAD digest')
 def timeout_pairs(rr,pairs):
  for a,b in pairs:
   for x in [rr[a],rr[b]]:
    h=x['headers'];require(h['content-type']=='application/problem+json' and h['retry-after']=='1' and h['cache-control']=='private, no-store' and h['etag'] is None and h['x-bdp-development-limit'] is None,'Timeout metadata')
   require(rr[a]['bodyOctets']>0 and rr[a]['headers']['content-length']==str(rr[a]['bodyOctets']) and rr[b]['headers']['content-length']==rr[a]['headers']['content-length'],'Timeout GET/HEAD length parity')
 def historical_routes(rr,scope,future_position):
  future=rr[2]['url'];require(re.fullmatch(re.escape(scope)+r'changes/\?after=cg_[A-Za-z0-9_-]{43}_[A-Za-z0-9_-]{43}_'+str(future_position),future),'Historical future checkpoint URL')
  expected=[scope+'operations/batch',scope+'changes/?after=genesis',future,future,future,future,scope+'changes/?after=genesis',scope+'changes/?after=genesis',future,scope+'changes/?after=genesis&start=now']
  if len(rr)==12:expected.extend([scope+'operations/batch',future])
  require([x['url'] for x in rr]==expected,'Historical route sequence')
 if mode=='source':
  expected=[('changefeed-opaque-capture',r) for r in ['unknown','pos-01']]+[('changefeed-opaque-preflight',r) for r in ['already-aborted','invalid-policy','throwing-clock','nonfinite-clock']]+[(n,None) for n in ['changefeed-opaque-abort','changefeed-opaque-boundaries','changefeed-opaque-http','private-source-changefeed-parameter-context']]
  require(len(observations)==10 and sorted((o['observation'],str(o.get('row'))) for o in observations)==sorted((n,str(r)) for n,r in expected),'Expected nine new and one historical report')
  for o in observations:
   scalars(o,{'outcome':'passed'});kind=o['observation']
   if kind!='private-source-changefeed-parameter-context':scalars(o,{'failureCount':0})
   if kind=='changefeed-opaque-capture':
    scalars(o,{'commits':2,'captures':2,'waits':1,'clocks':4,'closes':1});aa=rows(o,'anchors',2);rr=rows(o,'reads',4)
    require(aa[0]==aa[1]==o['positiveAfter'] and re.fullmatch(r'cg_[A-Za-z0-9_-]{43}_[A-Za-z0-9_-]{43}_1',aa[0]),'Fixed admission checkpoint')
    for x in rr:
     require(set(x)=={'after','minimum','minimumFrozen','admission','outcome','context'} and x['after']==aa[0],'Read invocation shape/anchor');context(x['context'])
    c=rr[0]['context'];current={**c,'scopePosition':'pos-2'}
    require(c['scopePosition']=='pos-1' and [x['context'] for x in rr]==[c,c,current,current],'Per-entry owner contexts')
    mm={'scopeEpoch':c['scopeEpoch'],'authorizationView':c['authorizationView'],'minimumPosition':o['row']}
    require([x['minimum'] for x in rr]==[mm,mm,mm,{**mm,'minimumPosition':'pos-2'}] and [x['minimumFrozen'] for x in rr]==[False,True,True,True],'Captured tuple values/frozen status')
    require([x['admission'] for x in rr]==['finite-start-now']*3+[None] and [x['outcome'] for x in rr]==['unproved']*3+['page'],'Direct/service/known read outcomes')
   elif kind=='changefeed-opaque-abort':
    scalars(o,{'captures':1,'reads':1,'waits':1,'clocks':2,'closes':1,'postCloseProbes':1,'order':['wait-entered','abort','wait-released','owner-close','close-resolved']})
   elif kind=='changefeed-opaque-preflight':
    clocks,error={'already-aborted':(1,'ScopeServerOperationAbortedError'),'invalid-policy':(0,'TypeError'),'throwing-clock':(1,'Error'),'nonfinite-clock':(1,'TypeError')}[o['row']]
    scalars(o,{'captures':1,'reads':0,'waits':0,'clocks':clocks,'errorName':error})
   elif kind=='changefeed-opaque-boundaries':
    scalars(o,{'captures':2,'reads':4,'waits':0,'clocks':2,'outcomes':['default-owner-TypeError','default-service-TypeError','after-TypeError','malformed-TypeError','unsupported-sentinel','foreign-unsupported','valid-closed','malformed-before-closed','owner-closed']})
   elif kind=='changefeed-opaque-http':
    scalars(o,{'attempted':10,'exchanges':10,'deadlineExceeded':False,'cleanupDeadlineExceeded':False,'closed':True,'directoryAbsent':True});directory_absent(o['directory'],'bdp-feed-opaque-http-')
    rr=rows(o,'records',10);wire_rows(rr,ex['sourceNewHttp'],'https://example.test/feed/');seed=header_context(rr[0]['headers']);seed_context(seed)
    require(all(header_context(x['headers'])==seed for x in rr[:7]),'Finite response current context')
    require(rr[1]['bodyOctets']>0 and rr[1]['headers']['content-length']==str(rr[1]['bodyOctets']) and rr[2]['headers']['content-length']==rr[1]['headers']['content-length'],'Known finite GET/HEAD lengths')
    timeout_pairs(rr,[(3,4),(5,6)])
    for x in rr[7:9]:require(x['headers']['x-bdp-development-limit']=='minimum-position' and x['headers']['content-type']!='text/event-stream','Strict after/SSE private fence')
    require(header_context(rr[9]['headers'])==[None]*3 and rr[9]['headers']['x-bdp-development-limit'] is None,'Receiving malformed fence')
   else:
    scalars(o,{'attempted':12,'exchanges':12,'contextFailures':[],'hardFailureCount':0,'deadlineExceeded':False,'startupCompleted':True,'closed':True,'directoryAbsent':True,'claimEligible':False});directory_absent(o['directory'],'bdp-feed-http-')
    rr=rows(o,'completedResponses',12);wire_rows(rr,ex['sourceHistorical'],'https://example.test/feed/');historical_routes(rr,'https://example.test/feed/',3)
    seed=rr[1]['context'];seed_context(seed)
    require(all(x['context']==seed for x in rr[2:8]) and all(x['context']==[None]*3 for x in rr[8:10]) and rr[11]['context']==seed[:2]+['pos-2'],'Historical current/refusal context')
    require(all(x['contentType']=='application/problem+json' for x in rr[2:8]+rr[11:12]),'Historical parameter Problem media')
  report.update({'actualInstrumentedObservations':observations,'newObservationRecords':9,'historicalObservationRecords':1,'prospectiveWorkload':ex['prospectiveWorkload'],'limitations':'17 exact passed callbacks; nine new records and one unchanged parameter-context report. Remaining legacy workload is source-derived. HTTP does not count owner attempts. Body/schema/conditional/raw-input assertions remain in pinned callbacks. Historical lifecycle claims stay narrower than new cleanup.'})
 else:
  require(len(observations)==1,'Expected one installed report');o=observations[0];spec=ex['installedHistorical' if mode=='history' else 'installedNew']
  scalars(o,{'observation':spec['observation'],'outcome':'passed','claimEligible':False,'workspaceDirty':True,'declaredSourceCommit':g['finalHead'],'exchanges':spec['exchanges']})
  b=o['binding'];served=b['served'];require(served['sourceCommit']==g['finalHead'] and served['profile']=='transactional' and served['history'] is False and served['scope']=='https://reference.invalid/acme/' and b['authorityMode']=='canonical','Runtime binding')
  for key in ['installedPayloadDigest','installedContractsDigest','policyDigest','authenticationDigest']:fingerprint(served[key])
  fingerprint(b['runtimeConfigurationIdentity']);fixture=W/'packages/conformance/test-support/transactional/fixture-v1.json';fj=read(fixture)
  require(served['fixtureDigest']==sha(fixture.read_bytes()) and served['fixtureId']==fj['id'] and served['capabilities']==sorted(fj['capabilities']),'Fixture binding')
  require(b['operationTargets']=={n:served['scope']+'operations/'+s for n,s in ex['operationTargets'].items()},'Operation target binding')
  proof=verify_consumer(g);require(Path(o['entry']).resolve(strict=True)==Path(proof['entryRealpath']) and o['entrySha256']==proof['entrySha256'] and served['installedPayloadDigest']==proof['payloadIdentity'],'External entry/payload differs')
  rr=rows(o,'completedResponses',spec['exchanges']);scope=served['scope'];wire_rows(rr,spec,scope)
  if mode=='history':
   scalars(o,{'artifactBytes':ex['artifactBytes'],'attempted':10,'deadlineExceeded':False,'closed':True,'startupCompleted':True,'childExitCode':0,'childSignalCode':None,'directoryAbsent':True,'failureCount':0})
   require(type(o['childPid']) is int and o['childPid'] in execution['sampledPids'],'Historical child PID not observed');directory_absent(o['directory'],'bdp-changefeed-context-installed-');historical_routes(rr,scope,2)
   seed=header_context(rr[1]['headers']);seed_context(seed)
   require(all(header_context(x['headers'])==seed for x in rr[2:8]) and all(header_context(x['headers'])==[None]*3 for x in rr[8:]),'Historical installed parameter context')
   require(all(x['headers']['content-type']=='application/problem+json' for x in rr[2:8]),'Historical installed Problem media')
   report['lifecycleLimit']='One actual sampled child PID/zero exit, closed flag and absent owned directory plus supervisor absence; historical callback has no full close-event census or the new cleanup guarantee.'
  else:
   scalars(o,{'attempted':4,'assertionsCompleted':True,'deadlineFired':False,'failureCategories':[],'artifactSha256':ex['artifactSha256']})
   life=o['lifecycleAtReport'];scalars(life,spec['successRequires']);require(type(life['pid']) is int and life['pid'] in execution['sampledPids'],'Actual new child unobserved');directory_absent(life['directory'],'bdp-changefeed-opaque-installed-')
   seed=header_context(rr[0]['headers']);seed_context(seed);require(all(header_context(x['headers'])==seed for x in rr),'Installed current context')
   require(all(x['headers']['x-bdp-development']=='non-attesting' for x in rr),'Private runtime disclosure')
   require(rr[1]['bodyOctets']>0 and rr[1]['headers']['content-length']==str(rr[1]['bodyOctets']),'Known finite page bytes')
   timeout_pairs(rr,[(2,3)]);require(all(x['headers']['location'] is None for x in rr),'No redirects')
   report['lifecycleLimit']='Actual child exit+close, timed cleanup and exchange settlement; no socket-close census or source attestation. Unknown/late cleanup is failure with retained directory, never late success.'
  report.update({'observation':o,'attribution':'Verified reviewed dirty source → fresh build → archives → consumer → actual child payload. Body/schema validations remain pinned passing callback assertions; no public qualification.'})
 write(stage+'.accepted-observation.json',report)
