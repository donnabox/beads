"""Closed Receipt-root proof auditor. Standard library only; no product imports.
Actual Receipt-root report schemas reconciled; root release and final source binding required.
"""
import sys
sys.dont_write_bytecode=True
from gate import *
import re,math
from urllib.parse import urlsplit,parse_qsl
mode=sys.argv[1]
stage={'source':'focused-source','history':'receipt-history-external','root':'receipt-root-external'}.get(mode)
g,_=load_gate({'tools':'bind-workspace-tools','build':'bind-fresh-build','source':'review-focused-source','history':'review-receipt-history-external','root':'review-receipt-root-external'}[mode],child=True)
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
 require(ex['finalized'] is True,'Receipt contracts not finalized')
 cases=[x for suite in result['testResults'] for x in suite['assertionResults']]
 chosen=[x for x in cases if x['fullName'] in config['selectedFullNames']]
 require(len(chosen)==len(config['selectedFullNames']) and sorted(x['fullName'] for x in chosen)==sorted(config['selectedFullNames']) and all(x['status']=='passed' for x in chosen) and result.get('success') is True,'Exact receipt callbacks did not pass')
 require(all(x['status'] in ['pending','skipped','todo'] for x in cases if x not in chosen),'Unexpected callback executed')
 execution=complete(stage,g);observations=[]
 for line in (E/(stage+'.log')).read_text().splitlines():
  try:x=json.loads(line.strip())
  except json.JSONDecodeError:continue
  if isinstance(x,dict) and 'observation' in x:observations.append(x)
 report={'sourceIdentity':g['_sourceIdentity'],'sourceBindingSha256':g['sourceBinding']['sha256'],'qualification':False,'selectedCallbacks':config['selectedFullNames'],'resultsSha256':sha((E/(stage+'.results.json')).read_bytes()),'logSha256':sha((E/(stage+'.log')).read_bytes())}
 def one(name):
  found=[x for x in observations if x.get('observation')==name];require(len(found)==1,'Missing/duplicate receipt observation '+name);return found[0]
 def fingerprint(x):require(isinstance(x,str) and re.fullmatch('[a-f0-9]{64}',x),'Invalid SHA256 field')
 def number(x):require(type(x) is int and x>=0,'Invalid octet/counter')
 def context(x):
  require(isinstance(x,dict) and set(x)=={'scopeEpoch','authorizationView','scopePosition'},'Owner context shape')
  require(all(isinstance(y,str) and re.fullmatch('[A-Za-z0-9_-]{1,256}',y) for y in x.values()),'Owner context token')
 if mode=='source':
  require(len(observations)==3,'Only three actual A/B/D receipt reports expected')
  a=one(ex['capture']['observation']);require(set(a)==set(ex['capture'])|{'ownerObservations'} and all(a[k]==z for k,z in ex['capture'].items()),'Capture counters differ')
  rows=a['ownerObservations'];require(len(rows)==10 and [x['kind'] for x in rows]==ex['captureKinds'],'Real owner observation order differs')
  for x in rows:require(set(x)=={'kind','context'},'Owner report shape')
  require(rows[7]['context'] is None,'Unsupported sentinel fabricated context')
  for i,x in enumerate(rows):
   if i!=7:context(x['context'])
  first=rows[0]['context'];current={**first,'scopePosition':'pos-2'};require(first['scopePosition']=='pos-1','Seed position differs')
  require(all(x['context']==current for i,x in enumerate(rows) if i not in [0,7]),'Retry/absence current contexts differ')
  require(one(ex['drain']['observation'])==ex['drain'],'Drain record differs')
  d=one('receipt-root-minimum-http');require(set(d)=={'observation','requests','responses','waits'} and d['requests']==10 and d['waits']==3,'HTTP record differs')
  spec=ex['http']['responses'];require(len(d['responses'])==10,'HTTP row count')
  for i,x in enumerate(d['responses']):
   require(set(x)==set(spec['fields']) and x['method']==spec['methods'][i] and x['status']==spec['statuses'][i] and x['route']==spec['routeClasses'][i],'HTTP route/order differs');number(x['bodyOctets'])
   if x['method']=='HEAD' or x['status']==500:require(x['bodyOctets']==0,'HEAD/private refusal has body')
  require(all(d['responses'][i]['bodyOctets']>0 for i in [0,1,3,6,8,9]),'Expected represented body missing')
  report.update({'actualInstrumentedObservations':observations,'plannedWorkload':ex['prospectiveWorkload'],'attribution':ex['attribution'],'limits':'Twelve passed selected callbacks; only three new report records. No measured aggregate36 or inferred owner attempts from HTTP; exact wire length/header assertions remain in passing source callback.'})
 else:
  require(len(observations)==1,'Expected one installed report');o=observations[0];spec=ex['historical' if mode=='history' else 'root']
  require(set(o)==set(spec['keys']) and o['observation']==spec['observation'],'Installed report shape differs')
  require(o['outcome']=='passed' and o['claimEligible'] is False and o['workspaceDirty'] is True and o['declaredSourceCommit']==g['finalHead'],'Installed outcome/source differs')
  require(o['attempted']==o['exchanges']==spec['requests'] and o['failureCount']==0 and o['failureCategories']==[],'Incomplete/failed installed work')
  for k in ['deadlineExceeded','cleanupExceeded']:require(o[k] is False,'Deadline/cleanup exceeded')
  for k in ['launchAttempted','startupCompleted','closed','directoryAbsent']:require(o[k] is True,'Child/store lifecycle incomplete')
  require(type(o['childPid']) is int and o['childPid'] in execution['sampledPids'] and o['childExitCode']==0 and o['childSignalCode'] is None,'Actual clean child exit unproved')
  directory=Path(o['directory']);require(directory.is_absolute() and directory.is_relative_to(E/(stage+'.tmp')) and directory.name.startswith('bdp-') and not directory.exists(),'Owned directory remains/escapes')
  if mode=='root':require(o['activeRequests']==0 and o['pendingOperations']==0,'Unsettled owned transport')
  require(o['artifactBytes']==ex['artifactBytes'],'Installed artifact byte total differs')
  b=o['binding'];served=b['served'];require(served['sourceCommit']==g['finalHead'] and served['profile']=='transactional' and served['history'] is False and served['scope']=='https://reference.invalid/acme/' and b['authorityMode']=='canonical','Runtime binding differs')
  for key in ['installedPayloadDigest','installedContractsDigest','policyDigest','authenticationDigest']:fingerprint(served[key])
  fingerprint(b['runtimeConfigurationIdentity']);fixture=W/'packages/conformance/test-support/transactional/fixture-v1.json';fj=read(fixture)
  require(served['fixtureDigest']==sha(fixture.read_bytes()) and served['fixtureId']==fj['id'] and served['capabilities']==sorted(fj['capabilities']),'Fixture binding differs')
  proof=verify_consumer(g);require(Path(o['entry']).resolve(strict=True)==Path(proof['entryRealpath']) and o['entrySha256']==proof['entrySha256'] and served['installedPayloadDigest']==proof['payloadIdentity'],'Child payload differs from verified consumer')
  responses=o['completedResponses'];require(len(responses)==spec['requests'],'Response count differs');scope=served['scope'];root=None;page=None
  def headers_context(headers):return [headers[k] for k in ex['headers'][:3]]
  for i,x in enumerate(responses):
   require(set(x)==set(ex['responseKeys']),'Response shape differs');require(x['ordinal']==i+1 and x['label']==spec['labels'][i] and x['method']==spec['methods'][i] and x['status']==spec['statuses'][i] and x['credential']==spec['credentials'][i] and x['minimumFieldOccurrences']==spec['minimumOccurrences'][i],'Response/request order differs')
   number(x['bodyOctets']);fingerprint(x['bodySha256']);expected=spec['requestBodies'].get(str(i+1));number(x['requestBodyBytes'])
   require(x['requestBodyBytes']==(expected['bytes'] if expected else 0) and x['requestBodySha256']==(expected['sha256'] if expected else None),'Request-body fingerprint differs')
   route=spec['routes'][i];u=x['url']
   if route=='root':
    require(re.fullmatch(re.escape(scope)+'receipts/[A-Za-z0-9_-]+',u) and u!=scope+'receipts/missing','Invalid issued root');require(root is None or root==u,'Root identity changed');root=u
   elif route=='page':
    require(root is not None and re.fullmatch(re.escape(root)+'/pages/[A-Za-z0-9_-]+',u),'Invalid issued page');require(page is None or page==u,'Page identity changed');page=u
   else:require(u==scope+{'batch':'operations/batch','update-where':'operations/update-where','point':'beads/r00','missing-point':'beads/missing','missing-root':'receipts/missing'}[route],'Fixed route differs')
   headers=x['headers'];require(set(headers)==set(ex['headers']) and headers['x-bdp-development']=='non-attesting','Retained headers shape/marker differs')
   require(all(z is None or isinstance(z,str) for z in headers.values()),'Invalid retained header value')
   if x['method']=='HEAD' or x['status'] in [500,405]:require(x['bodyOctets']==0 and x['bodySha256']==sha(b''),'Expected empty body')
   else:require(x['bodyOctets']>0,'Missing represented body')
  seedctx=headers_context(responses[0]['headers']);require(all(isinstance(z,str) and re.fullmatch('[A-Za-z0-9_-]{1,256}',z) for z in seedctx) and seedctx[2]=='pos-1','Actual seed context invalid')
  expectedctx=seedctx if mode=='root' else [*seedctx[:2],'pos-2']
  if mode=='root':
   for x in responses[1:]:require(headers_context(x['headers'])==expectedctx and x['headers']['x-bdp-development-limit'] is None,'Root response context/private fence differs')
   get,head=responses[1:3]
   for x in [get,head]:
    hh=x['headers'];require(hh['content-type']=='application/problem+json' and hh['cache-control']=='private, no-store' and hh['retry-after']=='1' and hh['etag'] is None and hh['last-modified'] is None,'503 metadata differs')
   require(get['headers']['content-length']==str(get['bodyOctets']) and head['headers']['content-length']==get['headers']['content-length'],'503 paired wire length differs')
  else:
   for i,x in enumerate(responses):
    want=seedctx if i==0 else [None,None,None] if x['status'] in [400,401,500,405] else expectedctx
    require(headers_context(x['headers'])==want,'Historical context order differs')
    if x['status']==400:require(x['headers']['content-type']=='application/problem+json' and x['headers']['cache-control']=='private, no-store' and x['headers']['x-bdp-development-limit'] is None,'Malformed metadata differs')
   get,head=responses[3:5];require(get['headers']['content-length']==str(get['bodyOctets']) and head['headers']['content-length']==get['headers']['content-length'],'Historical malformed HEAD length differs')
   own=responses[15]['headers'];require(own['retry-after']=='1' and own['cache-control']=='private, no-store' and own['x-bdp-development-limit'] is None,'Owned root timeout differs')
   require(responses[16]['headers']['x-bdp-development-limit']=='minimum-position','Page strict sentinel relaxed');require(responses[21]['headers']['allow']=='GET, HEAD','Method precedence differs')
  report.update({'observation':o,'attribution':'Independently verified frozen dirty source/build/archive/consumer chain; not child source attestation or profile qualification','limits':'Actual report fields only. In-memory receipt/problem schemas and exact conditional input assertions remain in passing pinned callback. Historical22 cleanup unchanged; actual clean child exit and supervisor absence supplement directory removal. NewE requires no active requests/pending operations and timely known clean close; no elapsed-derived waiter counts.'})
 write(stage+'.accepted-observation.json',report)
