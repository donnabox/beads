"""Closed Alias opaque proof auditor. Standard library only; no product imports.
Actual Alias opaque report schemas reconciled; root release and final source binding required.
"""
import sys
sys.dont_write_bytecode=True
from gate import *
import re,math
from urllib.parse import urlsplit,parse_qsl
mode=sys.argv[1]
stage={'source':'focused-source','history':'alias-history-external','alias':'alias-opaque-external'}.get(mode)
g,_=load_gate({'tools':'bind-workspace-tools','build':'bind-fresh-build','source':'review-focused-source','history':'review-alias-history-external','alias':'review-alias-opaque-external'}[mode],child=True)
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
 require(ex['finalized'] is True,'Alias contracts not finalized')
 cases=[x for suite in result['testResults'] for x in suite['assertionResults']]
 chosen=[x for x in cases if x['fullName'] in config['selectedFullNames']]
 require(len(chosen)==len(config['selectedFullNames']) and sorted(x['fullName'] for x in chosen)==sorted(config['selectedFullNames']) and all(x['status']=='passed' for x in chosen) and result.get('success') is True,'Exact alias callbacks did not pass')
 require(all(x['status'] in ['pending','skipped','todo'] for x in cases if x not in chosen),'Unexpected callback executed')
 execution=complete(stage,g);observations=[]
 for line in (E/(stage+'.log')).read_text().splitlines():
  try:x=json.loads(line.strip())
  except json.JSONDecodeError:continue
  if isinstance(x,dict) and 'observation' in x:observations.append(x)
 report={'sourceIdentity':g['_sourceIdentity'],'sourceBindingSha256':g['sourceBinding']['sha256'],'qualification':False,'selectedCallbacks':config['selectedFullNames'],'resultsSha256':sha((E/(stage+'.results.json')).read_bytes()),'logSha256':sha((E/(stage+'.log')).read_bytes())}
 def one(name):
  found=[x for x in observations if x.get('observation')==name];require(len(found)==1,'Missing/duplicate alias observation '+name);return found[0]
 def fingerprint(x):require(isinstance(x,str) and re.fullmatch('[a-f0-9]{64}',x),'Invalid SHA256 field')
 def number(x):require(type(x) is int and x>=0,'Invalid octet/counter')
 def context(x):
  require(isinstance(x,dict) and set(x)=={'scopeEpoch','authorizationView','scopePosition'},'Owner context shape')
  require(all(isinstance(y,str) and re.fullmatch('[A-Za-z0-9_-]{1,256}',y) for y in x.values()),'Owner context token')
 def scalars(o,expected):
  require(all(o.get(k)==v for k,v in expected.items()),'Report scalar mismatch '+str(expected))
 def rows(o,key,n):
  value=o.get(key);require(isinstance(value,list) and len(value)==n,'Report array count '+key);return value
 def ctxrows(value):
  for x in value:context(x)
 def mappings(value,contexts):
  for x,ctx in zip(value,contexts,strict=True):
   require(set(x) in [{'context'},{'target','context'}] and x['context']==ctx,'Mapping/context shape')
   require('target' not in x or x['target'] in ['beads/a','beads/b'],'Unexpected actual alias target')
 if mode=='source':
  require(len(observations)==10 and all(o.get('outcome')=='passed' for o in observations),'Expected ten passed alias reports')
  expected_keys=[('alias-minimum-capture','unknown'),('alias-minimum-capture','pos-01'),*[('alias-minimum-missing',r) for r in ['absent','deleted','dangling','hidden']],('alias-minimum-transition',None),('alias-minimum-drain',None),('alias-minimum-incident-fence',None),('alias-minimum-http',None)]
  require(sorted((o['observation'],str(o.get('row'))) for o in observations)==sorted((x,str(y)) for x,y in expected_keys),'Unexpected report identities')
  for o in observations:
   kind=o['observation']
   if kind=='alias-minimum-capture':
    scalars(o,{'apiCalls':3,'waits':2,'clockCalls':5,'writes':4});cc=rows(o,'contexts',5);mm=rows(o,'attempts',5);uu=rows(o,'unproved',4);ctxrows(cc);ctxrows(uu);mappings(mm,cc)
    require([m['target'] for m in mm]==['beads/a','beads/a','beads/b','beads/b','beads/b'] and [x['scopePosition'] for x in cc]==['pos-1','pos-1','pos-1','pos-2','pos-2'],'Capture sequence')
    current={**cc[0],'scopePosition':'pos-2'};require(cc==[cc[0],cc[0],cc[0],current,current] and uu==cc[:4],'Owner/unproved context sequence')
    require(o['directContext']==cc[0] and o['timeoutContext']==current and o['positiveContext']==current,'Direct/timeout/positive context')
   elif kind=='alias-minimum-missing':
    r=o['row'];scalars(o,{'apiCalls':3,'probeCalls':1,'conditionedCalls':2,'waits':0,'clockCalls':2,'writes':3 if r in ['deleted','dangling'] else 2,'policyReplacements':1 if r=='hidden' else 0})
    cc=rows(o,'entries',3);mm=rows(o,'mappings',3);ctxrows(cc);mappings(mm,cc);context(o['ordinaryContext']);require(cc==[o['ordinaryContext']]*3 and o['conditionedContexts']==[o['ordinaryContext']]*2,'Missing nondisclosure context')
   elif kind=='alias-minimum-transition':
    scalars(o,{'apiCalls':1,'terminalStatus':404,'waits':1,'clockCalls':2,'writes':3});cc=rows(o,'contexts',2);mm=rows(o,'attempts',2);ctxrows(cc);mappings(mm,cc)
    require(mm[0]['target']=='beads/a' and 'target' not in mm[1] and o['unproved']==[cc[0]] and cc[0]==cc[1]==o['resultContext'] and cc[0]['scopePosition']=='pos-1','Live missing transition/context')
   elif kind=='alias-minimum-drain':
    scalars(o,{'apiCalls':2,'handlerRequests':1,'waits':1,'clockCalls':3,'writes':2,'closed':True,'postCloseProbes':1});cc=rows(o,'entries',1);mm=rows(o,'mappings',1);ctxrows(cc);mappings(mm,cc)
   elif kind=='alias-minimum-incident-fence':
    scalars(o,{'performCalls':2,'aliasLookups':0,'waits':0,'clockCalls':2,'writes':2});cc=rows(o,'entries',2);ctxrows(cc);context(o['missingContext']);require(o['missingContext']==cc[-1],'Incident missing context')
   else:
    scalars(o,{'attempted':28,'waits':6});rr=rows(o,'responses',28);h=next(x for x in ex['source'] if x['observation']==kind)
    methods=['POST','POST','GET','GET','HEAD','HEAD']+['GET','HEAD']*5+['GET']*6+['GET','HEAD']*2+['GET','DELETE']
    routes=['operations/batch','operations/put-alias']+['alias/work']*14+['alias/missing','alias/missing','alias/work?invalid=1','alias/work','alias/missing','alias/work?invalid=1']+['alias/work']*6
    seed=rr[0]['context'];require(set(seed)=={'bdp-scope-epoch','bdp-authorization-view','bdp-scope-position'} and seed['bdp-scope-position']=='pos-1' and all(isinstance(x,str) and x for x in seed.values()),'HTTP seed context')
    for i,x in enumerate(rr):
     require(set(x)==set(h['responseFields']) and x['method']==methods[i] and x['route']==routes[i] and x['status']==h['statusSequence'][i],'HTTP order/shape')
     number(x['bodyOctets']);hh=x['headers'];require(set(hh)==set(h['safeHeaders']) and all(v is None or isinstance(v,str) for v in hh.values()),'Safe HTTP headers')
     if i<18 or i==20:require(x['context']==seed,'Expected actual HTTP context')
     elif i!=27:require(x['context']==dict.fromkeys(seed),'Earlier refusal/private gap unexpectedly exposes context')
     if x['method']=='HEAD' or x['status'] in [307,500]:require(x['bodyOctets']==0,'Empty HTTP response expected')
     if x['status']!=307:require(hh['location'] is None,'Refusal Location')
     if x['status'] in [307,503,409,410,500]:require(hh['etag'] is None,'Unexpected ETag')
     if x['status']==503:require(hh['retry-after']=='1' and hh['cache-control']=='private, no-store','Timeout metadata')
     if x['status']==500:require(hh['x-bdp-development-limit']=='minimum-position','Foreign opaque private fence')
    for i in [6,8,10,12,14]:require(rr[i]['bodyOctets']>0 and rr[i]['headers']['content-length']==str(rr[i]['bodyOctets']) and rr[i+1]['headers']['content-length']==rr[i]['headers']['content-length'],'GET/HEAD length parity')
  report.update({'actualInstrumentedObservations':observations,'prospectiveWorkload':ex['prospectiveWorkload'],'limitations':'20 exact passed cases, ten actual new records; legacy counts remain source-derived. HTTP does not measure owner attempts. Problem/receipt body validators and conditional inputs remain pinned callback assertions.'})
 else:
  require(len(observations)==1,'Expected one installed report');o=observations[0];spec=ex['installedHistorical' if mode=='history' else 'installedNew']
  scalars(o,{'observation':spec['observation'],'outcome':'passed','claimEligible':False,'workspaceDirty':True,'declaredSourceCommit':g['finalHead'],'exchanges':spec['exchanges']})
  b=o['binding'];served=b['served'];require(served['sourceCommit']==g['finalHead'] and served['profile']=='transactional' and served['history'] is False and served['scope']=='https://reference.invalid/acme/' and b['authorityMode']=='canonical','Runtime binding')
  for key in ['installedPayloadDigest','installedContractsDigest','policyDigest','authenticationDigest']:fingerprint(served[key])
  fingerprint(b['runtimeConfigurationIdentity']);fixture=W/'packages/conformance/test-support/transactional/fixture-v1.json';fj=read(fixture)
  require(served['fixtureDigest']==sha(fixture.read_bytes()) and served['fixtureId']==fj['id'] and served['capabilities']==sorted(fj['capabilities']),'Fixture binding')
  targets={n:served['scope']+'operations/'+s for n,s in ex['operationTargets'].items()};require(b['operationTargets']==targets,'Operation target binding')
  proof=verify_consumer(g);require(Path(o['entry']).resolve(strict=True)==Path(proof['entryRealpath']) and o['entrySha256']==proof['entrySha256'] and served['installedPayloadDigest']==proof['payloadIdentity'],'External entry/payload differs')
  rr=rows(o,'completedResponses',spec['exchanges']);scope=served['scope']
  if mode=='history':
   require(o['localSensitivityControls']==['admission-target','stale-postimage'],'Historical controls')
   ss=rows(o,'sessions',2);require(len({s['pid'] for s in ss})==2 and all(type(s['pid']) is int and s['pid'] in execution['sampledPids'] and s['exit']==0 for s in ss),'Historical child sessions')
   for i,(x,want) in enumerate(zip(rr,ex['historicalResponses'],strict=True)):
    require(set(x)=={'method','url','status','bodyOctets','bodySha256','wireSha256','scopePosition'} and [x['method'],x['url'],x['status']]==[want[0],scope+want[1],want[2]],'Historical response order')
    number(x['bodyOctets']);fingerprint(x['bodySha256']);fingerprint(x['wireSha256'])
    if x['status']==200:require(x['scopePosition']=='pos-'+str(1 if i<3 else 3 if i<19 else 4 if i<25 else 5),'Historical current position')
    if x['status']==307:require(x['bodyOctets']==0 and x['bodySha256']==sha(b''),'Historical redirect body')
   report['lifecycleLimit']='Two actual sampled child PIDs/zero exits and supervisor absence; historical callback exposes no directory or full close-event census. Do not relabel as new cleanup guarantee.'
  else:
   scalars(o,{'attempted':6,'assertionsCompleted':True,'deadlineFired':False,'failureCategories':[]});require(o['artifactSha256']==ex['artifactSha256'],'Callback artifact bytes')
   life=o['lifecycleAtReport'];scalars(life,spec['successRequires']);require(life['launchAttempted'] is True and type(life['pid']) is int and life['pid'] in execution['sampledPids'],'Actual new child unobserved')
   start=read(E/(stage+'.started.json'));tmp=Path(start['temporaryDirectory']);directory=Path(life['directory']);require(tmp.parent==Path('/private/tmp') and directory.is_absolute() and directory.is_relative_to(tmp) and directory.name.startswith('bdp-alias-opaque-installed-') and not directory.exists(),'Store ownership/absence')
   for i,x in enumerate(rr):
    require(set(x)==set(spec['responseFields']) and x['method']==spec['methodSequence'][i] and x['url']==scope+spec['suffixSequence'][i] and x['status']==spec['statusSequence'][i],'New response order/shape')
    number(x['bodyOctets']);fingerprint(x['bodySha256']);fingerprint(x['wireSha256']);hh=x['headers'];require(set(hh)==set(ex['installedHeaders']) and all(v is None or isinstance(v,str) for v in hh.values()) and hh['x-bdp-development']=='non-attesting','Safe installed headers')
    if x['method']=='HEAD' or x['status']==307:require(x['bodyOctets']==0 and x['bodySha256']==sha(b''),'Bodyless response expected')
    else:require(x['bodyOctets']>0,'Represented body missing')
   ctx=lambda h:[h[k] for k in ex['installedHeaders'][:3]]
   seed=ctx(rr[0]['headers']);require(all(isinstance(v,str) and re.fullmatch('[A-Za-z0-9_-]{1,256}',v) for v in seed) and seed[2]=='pos-1' and all(ctx(x['headers'])==seed for x in rr),'Installed current context')
   redirect=rr[2]['headers'];require(redirect['location']==scope+'beads/a' and redirect['content-length']=='0' and redirect['cache-control']=='private, no-store' and redirect['etag'] is None,'Known alias redirect')
   for x in rr[3:]:require(x['headers']['location'] is None,'Opaque/missing must not redirect')
   for x in rr[3:5]:
    hh=x['headers'];require(hh['retry-after']=='1' and hh['content-type']=='application/problem+json' and hh['cache-control']=='private, no-store' and hh['etag'] is None,'Opaque timeout metadata')
   require(rr[3]['headers']['content-length']==str(rr[3]['bodyOctets']) and rr[4]['headers']['content-length']==rr[3]['headers']['content-length'],'Installed HEAD length parity')
   report['lifecycleLimit']='Actual child exit+close, timed cleanup and exchange settlement; no socket-close census or source attestation. Unknown/late cleanup is failure with retained directory, never late success.'
  report.update({'observation':o,'attribution':'Verified reviewed dirty source→fresh build→archives→consumer→actual child payload. Body/schema validations remain pinned passing callback assertions; no public profile qualification.'})
 write(stage+'.accepted-observation.json',report)
