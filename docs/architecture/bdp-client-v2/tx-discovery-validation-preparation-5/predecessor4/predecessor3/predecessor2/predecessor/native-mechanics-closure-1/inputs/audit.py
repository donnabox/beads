"""D3a exact private callback/report audit; source-only preparation, root gate required."""
import sys
sys.dont_write_bytecode=True
from gate import *
import re,math
from collections import Counter
mode=sys.argv[1]
g,_=load_gate({'tools':'bind-workspace-tools','build':'bind-fresh-build'}.get(mode,'audit-'+mode),child=True)
if mode=='tools':
 require((E/'node-version.log').read_text().strip()=='v24.16.0' and (E/'pnpm-version.log').read_text().strip()=='11.20.0','Actual runtime/package-manager versions differ')
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
 stage=mode;cfg=read(P/'commands.json')['commands'][stage];sel=read(P/'selection.json')[stage]
 result=read(E/(stage+'.results.json'));complete(stage,g)
 def keys(o,required,optional=()):
  require(type(o) is dict and set(required)<=set(o)<=set(required)|set(optional),'Missing/extra report key '+str(required))
 def exact(o,want):
  for k,v in want.items():require(k in o and type(o[k]) is type(v) and o[k]==v,'Typed actual field '+k)
 def integer(v):require(type(v) is int and v>=0,'Actual nonnegative integer')
 def finite(v,limit=None):require(type(v) in [int,float] and math.isfinite(v) and v>=0 and (limit is None or v<limit),'Actual finite bounded value')
 def digest(v):require(type(v) is str and re.fullmatch('[a-f0-9]{64}',v),'Actual SHA256')
 def strings(o):require(type(o) is dict and all(type(k) is str and type(v) is str for k,v in o.items()),'Header map')
 def absent_pid(pid):
  require(type(pid) is int and pid>0,'Actual child PID')
  try:os.kill(pid,0)
  except ProcessLookupError:return
  except OSError:raise RuntimeError('PID absence unknown')
  raise RuntimeError('Child remains')
 def absent_dir(value,prefix):
  tmp=Path(read(E/(stage+'.started.json'))['temporaryDirectory']);p=Path(value)
  require(p.is_absolute() and p.is_relative_to(tmp) and p.name.startswith(prefix) and not p.exists(),'Owned BDP directory absence')
 chosen=[]
 for suite in result['testResults']:
  for case in suite['assertionResults']:
   if case['fullName'] in cfg['selectedFullNames']:
    require(Path(suite['name']).resolve()==(W/cfg['selectedFilesByFullName'][case['fullName']]).resolve(),'Selected callback file')
    chosen.append(case)
   else:require(case['status'] in ['pending','skipped','todo'],'Unselected callback executed')
 require(Counter(x['fullName'] for x in chosen)==Counter(cfg['selectedFullNames']) and all(x['status']=='passed' for x in chosen),'Exact callbacks pass once')
 exact(result,{'success':True,'numFailedTests':0,'numFailedTestSuites':0,'numPassedTests':len(chosen)})
 if 'numRuntimeErrorTestSuites' in result:exact(result,{'numRuntimeErrorTestSuites':0})
 log=(E/(stage+'.log')).read_text();require(not re.search('Unhandled Errors|Unhandled Rejection|Uncaught Exception|Worker exited unexpectedly',log),'Unhandled runtime diagnostic')
 reports=[]
 for line in log.splitlines():
  try:o=json.loads(line.strip(),object_pairs_hook=unique_object,parse_constant=invalid_constant)
  except json.JSONDecodeError:
   require('\"observation\"' not in line,'Malformed observation JSON')
   continue
  if type(o) is dict and 'observation' in o:reports.append(o)
 schemas=read(P/'report-schemas.json');kind=sel['kind']
 def named_fields(fields):
  require(type(fields) is list,'Field occurrence array')
  for item in fields:
   keys(item,['name','value']);require(type(item['name']) is str and type(item['value']) is str,'Actual field values')
 def headers_record(r):
  strings(r['headers']);integer(r['status']);require(100<=r['status']<=599,'HTTP status')
 def source_report(o):
  keys(o,['observation','callback','elapsedMs','workSettled','cleanupKnown','directories','observations','failureCount']);exact(o,{'observation':'D3-source','workSettled':True,'cleanupKnown':True,'failureCount':0});finite(o['elapsedMs'],4000)
  require(o['callback'] in cfg['selectedFullNames'],'Report callback identity')
  for d in o['directories']:
   keys(d,['directory','absent']);exact(d,{'absent':True});absent_dir(d['directory'],'bdp-d3-')
  counters=[];raw=[];clones=[]
  for r in o['observations']:
   if 'baseline' in r:
    keys(r,['baseline']);keys(r['baseline'],['calls','entries','assemblies']);integer(r['baseline']['entries']);integer(r['baseline']['assemblies']);require(type(r['baseline']['calls']) is list and all(x=='readConfiguration' for x in r['baseline']['calls']),'Construction evidence')
   elif 'counters' in r:
    keys(r,['counters','records']);keys(r['counters'],['direct','discovery','entries','assemblies','waits','clocks','commits'])
    for v in r['counters'].values():integer(v)
    for rr in r['records']:
     keys(rr,['method','url','credential','status','headers'],['body']);headers_record(rr)
    counters.append(r['counters'])
   elif 'counterfactual' in r:
    keys(r,['counterfactual','rejected','wireRequests']);exact(r,{'rejected':True,'wireRequests':0});clones.append(r['counterfactual'])
   else:
    keys(r,['transport','method','target','credential','fieldOccurrences','status','headers','bodyOctets','bodySha256','wireSha256']);exact(r,{'transport':'source-local-raw'});headers_record(r);integer(r['bodyOctets']);digest(r['bodySha256']);digest(r['wireSha256']);require(r['bodyOctets']<=131072,'Raw body bound');raw.append(r)
  want=schemas['sourceExpected'][o['callback']]
  require(len(raw)==want['wire'] and len(clones)==want.get('clones',0),'Exact source raw/clone counts')
  require([r['status'] for r in raw]==want['rawStatuses'],'Exact actual raw status order')
  if clones:require(Counter(clones)==Counter(schemas['sourceCounterfactuals']),'Exact counterfactual identities')
  require(len(counters)==(0 if want['direct']==0 else 1),'Actual source counter census')
  if counters:exact(counters[0],want['counters'])
 if kind=='D3-source':
  require(Counter(o.get('callback') for o in reports)==Counter(cfg['selectedFullNames']) and all(o.get('observation')=='D3-source' for o in reports),'Exactly21 source reports')
  for o in reports:source_report(o)
 elif kind in ['legacy-installed-contracts-suite','schema-zero-wire']:
  require(reports==[],'Unexpected observation from no-report suite')
 elif kind=='legacy-changefeed-suite':
  require(Counter(o['observation'] for o in reports)==Counter(schemas['legacyChangefeedExpectedLabels']),'Exact historical report census')
  for o in reports:
   shape=schemas['legacyChangefeedShapes'][o['observation']];keys(o,shape['required'],shape.get('optional',[]))
   if 'outcome' in o:exact(o,{'outcome':'passed'})
   for field in ['failureCount','hardFailureCount']:
    if field in o:exact(o,{field:0})
   if 'claimEligible' in o:exact(o,{'claimEligible':False})
   if 'directoryAbsent' in o:exact(o,{'directoryAbsent':True})
   for field in ['deadlineExceeded','cleanupDeadlineExceeded']:
    if field in o:exact(o,{field:False})
   if 'cleanupState' in o:
    state=o['cleanupState'];exact(state,{'closed':True,'stopUnknown':False,'directoryAbsent':True});absent_dir(state['directory'],'bdp-current-feed-')
   if 'directory' in o and o['directory'] is not None:absent_dir(o['directory'],'bdp-')
   if 'contextFailures' in o:require(o['contextFailures']==[],'Actual context failures')
 else:
  require(len(reports)==1,'Exactly one installed observation');o=reports[0]
  require(o['observation']==cfg['expectedObservation'],'Installed observation identity');exact(o,{'outcome':'passed','claimEligible':False,'workspaceDirty':True,'declaredSourceCommit':g['finalHead']})
  proof=verify_consumer(g);require(Path(o['entry']).resolve(strict=True)==Path(proof['entryRealpath']) and o['entrySha256']==proof['entrySha256'],'Actual external entry bytes')
  if kind=='D3-installed':
   keys(o,schemas['installedD3Required']);exact(o,{'failureCount':0,'workSettled':True,'cleanupKnown':True,'pending':0,'directoryAbsent':True,'entryLiteral':proof['entryRealpath']});finite(o['elapsedMs'],60000);absent_dir(o['directory'],'bdp-d3-installed-')
   keys(o['counts'],['scheduled','completed','aborts','exact','raw']);exact(o['counts'],cfg['expectedCounts'])
   keys(o['child'],['pid','exitSeen','closeSeen','code','signal']);exact(o['child'],{'exitSeen':True,'closeSeen':True,'code':0,'signal':None});absent_pid(o['child']['pid'])
   keys(o['metadata'],schemas['metadataKeys']);exact(o['metadata'],{'profile':'transactional','history':False,'claimEligible':False,'authorityMode':'canonical'});require(o['metadata']['payloadIdentity']==proof['payloadIdentity'],'Actual child payload identity')
   for k in ['payloadIdentity','policyIdentity','authenticationIdentity','runtimeConfigurationIdentity']:digest(o['metadata'][k])
   require(type(o['metadata']['artifactIdentity']) is str and re.fullmatch('development-schema5-[a-f0-9]{64}',o['metadata']['artifactIdentity']),'Versioned actual artifact identity')
   require(len(o['records'])==cfg['expectedCounts']['scheduled'],'One actual terminal per scheduled request')
   exact_records=[];raw_records=[]
   for r in o['records']:
    if r.get('transport')=='public-exact':
     keys(r,['id','transport','method','url','credential','fieldOccurrences','status','headers','bodyOctets','bodySha256','wireSha256','requestWriteState']);headers_record(r);exact(r,{'requestWriteState':'complete'});digest(r['bodySha256']);digest(r['wireSha256']);integer(r['bodyOctets']);require(r['bodyOctets']<=131072,'Body cap');named_fields(r['fieldOccurrences']);exact_records.append(r)
    else:
     keys(r,['id','transport','method','url','credential','state','terminal','partialWireAvailable','actualWireOctets'],['partialStatus','actualWireSha256']);exact(r,{'transport':'callback-local-raw','method':'GET'});integer(r['actualWireOctets']);require(r['actualWireOctets']<=163840,'Actual raw wire bound');require(type(r['partialWireAvailable']) is bool and r['partialWireAvailable']==(r['actualWireOctets']>0),'Honest raw availability')
     require(('actualWireSha256' in r)==r['partialWireAvailable'],'No invented wire fingerprint')
     if r['partialWireAvailable']:digest(r['actualWireSha256'])
     keys(r['state'],['settled','responseStarted','written','closeRequested'],['writeAtMs','closeRequestedAtMs','responseAtMs']);exact(r['state'],{'settled':True,'written':True})
     if cfg['expectedCounts']['aborts']:
      exact(r,{'terminal':'abort','partialWireAvailable':False,'actualWireOctets':0});exact(r['state'],{'closeRequested':True,'responseStarted':False});require('partialStatus' not in r,'No invented close status');finite(r['state']['writeAtMs']);finite(r['state']['closeRequestedAtMs']);require(r['state']['writeAtMs']<=r['state']['closeRequestedAtMs'],'Actual close chronology')
     else:require(r['terminal']=='response' and r['partialStatus']==400 and r['state']['responseStarted'] is True,'Actual Host refusal response')
     raw_records.append(r)
   require(len(exact_records)==cfg['expectedCounts']['exact'] and len(raw_records)==cfg['expectedCounts']['raw'],'Actual exact/raw census')
  elif kind=='legacy-owner-installed':
   keys(o,schemas['legacyOwnerRequired']);exact(o,{'cleanupKnown':True,'directoryAbsent':True,'exchanges':12});absent_dir(o['directory'],'bdp-discovery-installed-');keys(o['session'],['pid','exit']);exact(o['session'],{'exit':0});absent_pid(o['session']['pid']);keys(o['childTerminal'],['exitSeen','closeSeen','code','signal']);exact(o['childTerminal'],{'exitSeen':True,'closeSeen':True,'code':0,'signal':None});require(len(o['completedResponses'])==12,'Legacy12 actual responses')
   for r in o['completedResponses']:
    keys(r,['method','url','status','bodyOctets','bodySha256','wireSha256']);integer(r['bodyOctets']);digest(r['bodySha256']);digest(r['wireSha256'])
   require([r['status'] for r in o['completedResponses']]==[204,200,200,204,200,200,401,401,401,200,200,200],'Actual legacy12 statuses')
   require(o['binding']['served']['installedPayloadDigest']==proof['payloadIdentity'],'Legacy child payload')
  elif kind=='legacy-schema-installed':
   keys(o,schemas['legacySchemaRequired']);exact(o,{'attemptedRequests':14,'directoryAbsent':True});require(len(o['completedResponses'])==14 and len(o['sessions'])==1,'Schema14 exchanges/one child lifetime')
   require([r['label'] for r in o['completedResponses']]==schemas['schemaLabels'],'Schema request order')
   for session in o['sessions']:
    keys(session,['pid','exit','signal']);exact(session,{'exit':0,'signal':None});absent_pid(session['pid'])
   for r in o['completedResponses']:
    keys(r,['label','method','url','status','requestBodyBytes','requestBodySha256','responseBodyBytes','responseBodySha256','wireSha256','context']);integer(r['requestBodyBytes']);integer(r['responseBodyBytes']);digest(r['requestBodySha256']);digest(r['responseBodySha256']);digest(r['wireSha256'])
   require([r['status'] for r in o['completedResponses']]==[200,200,200,200,200,406,304,412,404,401,405,204,200,200],'Actual schema14 status order')
   for r in o['completedResponses']:
    keys(r['context'],[],['epoch','view','position']);strings(r['context'])
   require(o['binding']['served']['installedPayloadDigest']==proof['payloadIdentity'],'Schema child payload')
  else:raise RuntimeError('Unknown observation grammar')
 write(stage+'.accepted-observation.json',{'sourceIdentity':g['_sourceIdentity'],'sourceBindingSha256':g['sourceBinding']['sha256'],'resultsSha256':sha((E/(stage+'.results.json')).read_bytes()),'logSha256':sha((E/(stage+'.log')).read_bytes()),'selectedFullNames':cfg['selectedFullNames'],'reports':reports,'claimEligible':False,'limitations':'Private exact source/build/archive/consumer only. Historical lanes distinct. Failure/partial records remain raw, never reconstructed; no public Write qualification/admission claim.'})
