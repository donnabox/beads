"""Closed Snapshot proof auditor. Standard library only; no product imports.
Actual Snapshot report schemas reconciled; root release and final source binding required.
"""
import sys
sys.dont_write_bytecode=True
from gate import *
import re,math
from urllib.parse import urlsplit,parse_qsl
mode=sys.argv[1]
stage={'source':'focused-source','history':'snapshot-history-external','opaque':'snapshot-opaque-external'}.get(mode)
g,_=load_gate({'tools':'bind-workspace-tools','build':'bind-fresh-build','source':'review-focused-source','history':'review-snapshot-history-external','opaque':'review-snapshot-opaque-external'}[mode],child=True)
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
 require(ex['finalized'] is True,'Snapshot report contracts not reconciled to frozen source')
 cases=[x for suite in result['testResults'] for x in suite['assertionResults']]
 chosen=[x for x in cases if x['fullName'] in config['selectedFullNames']]
 require(len(chosen)==len(config['selectedFullNames']) and sorted(x['fullName'] for x in chosen)==sorted(config['selectedFullNames']) and all(x['status']=='passed' for x in chosen) and result.get('success') is True,'Exact Snapshot selected callbacks did not pass')
 require(all(x['status'] in ['pending','skipped','todo'] for x in cases if x not in chosen),'Unexpected callback executed')
 execution=complete(stage,g)
 observations=[]
 for line in (E/(stage+'.log')).read_text().splitlines():
  try:x=json.loads(line.strip())
  except json.JSONDecodeError:continue
  if isinstance(x,dict) and 'observation' in x:observations.append(x)
 report={'sourceIdentity':g['_sourceIdentity'],'sourceBindingSha256':g['sourceBinding']['sha256'],'qualification':False,'selectedCallbacks':config['selectedFullNames'],'resultsSha256':sha((E/(stage+'.results.json')).read_bytes()),'logSha256':sha((E/(stage+'.log')).read_bytes())}
 def one(selector):
  found=[x for x in observations if all(x.get(k)==v for k,v in selector.items())]
  require(len(found)==1,'Missing/duplicate Snapshot observation '+str(selector));return found[0]
 def finite(value,minimum=0,maximum=None):
  require(type(value) in [int,float] and math.isfinite(value) and value>=minimum and (maximum is None or value<maximum),'Invalid bounded numeric observation')
 def fingerprint(value):require(isinstance(value,str) and re.fullmatch('[0-9a-f]{64}',value),'Missing fingerprint')
 def at(obj,path):
  for part in path:obj=obj[part]
  return obj
 def contract(o,c):
  # Declarative source-pinned exact fields; no eval, product code, or inferred counters.
  require(set(o)==set(c['keys']),'Snapshot observation shape differs')
  for rule in c['equals']:require(at(o,rule['path'])==rule['value'],'Snapshot value differs '+str(rule['path']))
  for left,right in c.get('same',[]):require(at(o,left)==at(o,right),'Snapshot paired state differs')
  for rule in c.get('finite',[]):finite(at(o,rule['path']),rule.get('minimum',0),rule.get('maximumExclusive'))
 if mode=='source':
  require(len(observations)==ex['newSourceObservationCount'],'Unexpected Snapshot source observation count')
  used=[]
  for expected in ex['sourceExactRecords']:
   o=one(expected['selector']);require(o==expected['record'],'Snapshot exact counter/order record differs');used.append(o)
  owner=one({'observation':'fresh-snapshot-opaque-owner'});contract(owner,ex['ownerContract']);used.append(owner)
  require(set(owner['baseline'])=={'snapshots','records','pages','disposition'},'Owner baseline shape differs')
  require(isinstance(owner['baseline']['disposition'],str) and json.loads(owner['baseline']['disposition'])['expired'] is False,'Owner preserved disposition expired')
  http=one({'observation':'fresh-snapshot-opaque-http'});contract(http,ex['httpContract']['record']);used.append(http)
  responses=at(http,ex['httpContract']['responsesPath']);wanted=ex['httpContract']['orderedResponses'];require(len(responses)==len(wanted)==27,'New Snapshot HTTP response count differs')
  handle=None
  for actual,expected in zip(responses,wanted):
   require(set(actual)==set(ex['httpContract']['responseKeys']),'Snapshot HTTP response shape differs')
   route=actual[ex['httpContract']['routeField']]
   target=expected['route']
   if target in ['__MANIFEST__','__PAGE__']:
    u=urlsplit(route);require(not u.scheme and not u.netloc and not u.fragment,'Unexpected absolute source route')
    q=parse_qsl(u.query,keep_blank_values=True);require(u.path==ex['httpContract']['snapshotPath'] and len(q)==(1 if target=='__MANIFEST__' else 3) and q[0][0]=='snapshot' and bool(q[0][1]),'Retained Snapshot route differs')
    if target=='__PAGE__':require(q[1:]==[('stream','beads'),('page','unused')],'Retained page control differs')
    require(handle is None or handle==q[0][1],'Retained handle changed');handle=q[0][1]
   else:require(route==target,'Snapshot HTTP route differs')
   require(actual['method']==expected['method'] and actual['status']==expected['status'],'Snapshot HTTP method/status differs')
  require(len(used)==len(observations) and len({id(x) for x in used})==len(used),'Source observation not uniquely accounted')
  report.update({'actualInstrumentedObservations':observations,'plannedWorkload':config['prospectiveWorkload'],'attribution':'22 exact selected Snapshot callbacks;13 new callback records only. Historical18 network and15 handler counts remain source-derived unless explicitly reported; no inferred measured total45/60. A offline reopen counts are owner connections, not stores/children. B/C/F semantic assertions remain in pinned callbacks.'})
 else:
  require(len(observations)==1,'Unexpected installed Snapshot report count');o=observations[0]
  require(o['observation']==config['expectedObservation'] and (o.get('outcome')=='passed' if mode=='history' else o.get('ok') is True) and o['claimEligible'] is False,'Installed Snapshot outcome differs')
  require(o['workspaceDirty'] is True and o['declaredSourceCommit']==g['finalHead'],'Installed Snapshot source attribution differs')
  b=o['binding'];served=b['served'];require(served['sourceCommit']==g['finalHead'] and served['profile']=='transactional' and served['history'] is False and served['scope']=='https://reference.invalid/acme/' and b['authorityMode']=='canonical','Snapshot runtime binding differs')
  for f in ['installedPayloadDigest','installedContractsDigest','policyDigest','authenticationDigest']:fingerprint(served[f])
  fingerprint(b['runtimeConfigurationIdentity'])
  fixture=W/'packages/conformance/test-support/transactional/fixture-v1.json';fj=read(fixture)
  require(served['fixtureDigest']==sha(fixture.read_bytes()) and served['fixtureId']==fj['id'] and served['capabilities']==sorted(fj['capabilities']),'Snapshot fixture binding differs')
  proof=verify_consumer(g);entry=Path(proof['entryRealpath']);require(Path(o['entry']).resolve(strict=True)==entry and o['entrySha256']==proof['entrySha256'] and served['installedPayloadDigest']==proof['payloadIdentity'],'Snapshot child differs from verified external consumer')
  if mode=='history':
   require(set(o)==set(ex['historicalInstalledReportKeys']),'Historical Snapshot report shape differs')
   require(o['aborted'] is False and o['exchanges']==31 and len(o['completedResponses'])==31,'Historical Snapshot deadline/count differs')
   children=o['children'];require(len(children)==2 and len({x['pid'] for x in children})==2,'Historical Snapshot reopen child count differs')
   for child in children:require(type(child['pid']) is int and child['pid'] in execution['sampledPids'] and child['exitCode']==0 and child['signalCode'] is None,'Historical child clean exit/census absent')
   responses=o['completedResponses'];require([x['status'] for x in responses]==ex['historicalInstalledStatusSequence'],'Historical Snapshot ordered statuses differ')
   scope=served['scope'];fresh=scope+'snapshot';first=None
   for i,r in enumerate(responses,1):
    require(set(r)=={'url','status','method','responseBytes','wireSha256'},'Historical response schema differs');fingerprint(r['wireSha256']);require(type(r['responseBytes']) is int and r['responseBytes']>=0,'Historical body count invalid')
    method='POST' if i in [3,25] else 'HEAD' if i in [5,10,15,17,24,29] else 'GET';require(r['method']==method,'Historical ordered method differs')
    if method=='HEAD':require(r['responseBytes']==0,'Historical HEAD body not empty')
    if i in [22,23,24,27,28,29]:
     u=urlsplit(r['url']);v=urlsplit(fresh);q=parse_qsl(u.query,keep_blank_values=True);require(u.scheme==v.scheme and u.netloc==v.netloc and u.path==v.path and not u.fragment and len(q)==1 and q[0][0]=='snapshot' and bool(q[0][1]),'Historical retained URL invalid')
     require(first is None or first==r['url'],'Historical retained identity changed');first=r['url']
    else:require(r['url']==(scope+'operations/batch' if i in [3,25] else fresh),'Historical ordered URL differs')
   report['cleanupAttribution']='Historical preserved shared helper reports both children clean; supervisor verifies owned-group/sample absence and no bdp store directories. Historical report has no directory identity, so no invented per-store path assertion; its unconditional deletion alone is not exit proof.'
  else:
   c=ex['newInstalledContract'];contract(o,c['record'])
   require(type(o[c['pidField']]) is int and o[c['pidField']]>0 and o[c['pidField']] in execution['sampledPids'],'New Snapshot child not sampled')
   directory=Path(o[c['directoryField']]);require(directory.is_absolute() and directory.is_relative_to(E/(stage+'.tmp')) and directory.name.startswith('bdp-') and not directory.exists(),'New Snapshot owned store remains or escapes')
   responses=o[c['responsesField']];require(len(responses)==3,'New Snapshot response count differs')
   require([{k:r[k] for k in ['method','url','status']} for r in responses]==c['orderedResponses'],'New Snapshot request order differs')
   for r in responses:
    require(set(r)==set(c['responseKeys']),'New Snapshot response shape differs')
    require(type(r[c['bodyOctetsField']]) is int and r[c['bodyOctetsField']]>=0,'New Snapshot octet count invalid')
    for field in c['responseDigestFields']:fingerprint(r[field])
    finite(r['elapsedSeconds']);require(r['elapsedSeconds']<=o['elapsedSeconds'],'Request elapsed exceeds total')
   require(sum(r['elapsedSeconds'] for r in responses)<=o['elapsedSeconds'],'Sequential request elapsed exceeds total')
   seed,get,head=responses
   require(seed[c['bodyOctetsField']]>0 and get[c['bodyOctetsField']]>0 and head[c['bodyOctetsField']]==0,'New Snapshot seed/GET/HEAD body lengths differ')
   # Headers, body/receipt schemas, exact request/body digest, required context,
   # and HEAD metadata are assertions in the pinned callback, not report fields.
   # There is deliberately no optional HEAD media assertion or invented digest.
   report['cleanupAttribution']='Callback-local timely clean close, actual owned directory absent, sampled child exit and supervisor group/census/store checks. Failure/unknown/late-close is never accepted.'
  report.update({'observation':o,'attribution':'Base HEAD plus independently pinned dirty authored source, fresh build/archive/consumer proof; not child-attested source or full profile qualification','limits':'Only retained response fields are audited here. Exact requests/seed body/receipt/problem schema/body and omitted validators remain passing pinned callback assertions. No invented report data, no elapsed-derived waiter counts, no historical evidence relabeling.'})
 write(stage+'.accepted-observation.json',report)
