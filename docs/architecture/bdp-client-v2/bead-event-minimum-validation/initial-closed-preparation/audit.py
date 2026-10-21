"""Standard-library finite Event validation auditor; no product imports."""
import sys
sys.dont_write_bytecode=True
from gate import *
import re
mode=sys.argv[1]
if mode in ['source','bead']:raise RuntimeError('CLOSED skeleton: final authored Bead/source observation schema and checks are not bound')
stage={'source':'focused-source','minimum':'event-minimum-external','missing':'event-missing-context-external'}.get(mode)
g,_=load_gate({'tools':'bind-workspace-tools','build':'bind-fresh-build','source':'review-focused-source','minimum':'review-event-minimum-external','missing':'review-event-missing-context-external'}[mode],child=True)
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
 config=read(P/'commands.json')['commands'][stage];result=read(E/(stage+'.results.json'))
 cases=[x for suite in result['testResults'] for x in suite['assertionResults']]
 chosen=[x for x in cases if x['fullName'] in config['selectedFullNames']]
 require(len(chosen)==len(config['selectedFullNames']) and sorted(x['fullName'] for x in chosen)==sorted(config['selectedFullNames']) and all(x['status']=='passed' for x in chosen) and result.get('success') is True,'Exact selected callbacks did not pass')
 require(all(x['status'] in ['pending','skipped','todo'] for x in cases if x not in chosen),'Unexpected callback executed')
 complete(stage,g)
 report={'sourceIdentity':g['_sourceIdentity'],'sourceBindingSha256':g['sourceBinding']['sha256'],'qualification':False,'selectedCallbacks':config['selectedFullNames'],'resultsSha256':sha((E/(stage+'.results.json')).read_bytes()),'logSha256':sha((E/(stage+'.log')).read_bytes())}
 if mode=='source':
  observations=[]
  for line in (E/(stage+'.log')).read_text().splitlines():
   try:x=json.loads(line.strip())
   except json.JSONDecodeError:continue
   if isinstance(x,dict) and 'observation' in x:observations.append(x)
  def one(label):
   found=[x for x in observations if x['observation']==label];require(len(found)==1,'Missing/duplicate source observation '+label);return found[0]
  owner=one('finite-event-minimum-owner');require(owner=={'observation':'finite-event-minimum-owner','reads':8,'seedMutations':1},'Owner count differs')
  capture=one('finite-event-minimum-capture');require(capture=={'observation':'finite-event-minimum-capture','reads':4,'serviceCalls':2,'waits':2,'probes':2,'groupReads':0,'attemptTimes':[0,0,5,10]},'Capture observation differs')
  drain=one('finite-event-minimum-cancel-drain');require(drain=={'observation':'finite-event-minimum-cancel-drain','reads':5,'serviceCalls':4,'waits':3,'probes':3,'groupReads':0,'attemptTimes':[0,0,5,10,15],'postCloseDirectRefusals':1},'Drain observation differs')
  http=one('finite-event-minimum-http');require(http=={'observation':'finite-event-minimum-http','calls':32,'productChildren':0,'directoryAbsent':True,'failureCount':0},'New emitted HTTP count/cleanup differs')
  labels=read(P/'successor-selection.json')['preflightLabels']
  preflight=[x for x in observations if x['observation']=='finite-event-minimum-preflight']
  require(sorted(preflight,key=lambda x:x['label'])==sorted([{'observation':'finite-event-minimum-preflight','label':label,'serviceCalls':1,'clockCalls':0 if label=='invalid-policy' else 1,'ownerReads':0,'waits':0} for label in labels],key=lambda x:x['label']),'Preflight row reports differ')
  require(len(observations)==9,'Unexpected source observation record')
  report['actualInstrumentedObservations']=observations
  report.update({'attribution':'Selected source owner/service and fresh emitted HTTP callbacks; actual names verified individually','plannedWorkload':config['prospectiveWorkload'],'limits':'Counts are derived from frozen callback code, not fabricated observed owner/HTTP totals; preserve service precedence controls as separate service calls and owner attempts. Assertions of passing pinned callbacks supply semantic evidence.'})
 else:
  observations=[]
  for line in (E/(stage+'.log')).read_text().splitlines():
   try:x=json.loads(line.strip())
   except json.JSONDecodeError:continue
   if isinstance(x,dict) and 'observation' in x:observations.append(x)
  require(len(observations)==1 and observations[0]['observation']==config['expectedObservation'],'Missing/duplicate/unexpected callback report');o=observations[0]
  require(o['claimEligible'] is False and o['outcome']=='passed' and o['closed'] is True and o['deadlineExceeded'] is False and o['failureCount']==0 and o['startupCompleted'] is True,'Outcome/cleanup differs')
  require(o['workspaceDirty'] is True and o['declaredSourceCommit']==g['finalHead'],'Source attribution differs')
  require(isinstance(o['childPid'],int) and not isinstance(o['childPid'],bool) and o['childPid']>0 and o['childExitCode']==0 and o['childSignalCode'] is None,'Clean child exit absent')
  execution=complete(stage,g);require(o['childPid'] in execution['sampledPids'],'Reported child not observed by supervisor')
  require(o['attempted']==o['exchanges']==config['expectedExchanges'] and len(o['completedResponses'])==config['expectedExchanges'],'Exchange count differs')
  directory=Path(o['directory']);require(directory.is_absolute() and directory.is_relative_to(E/(stage+'.tmp')) and directory.name.startswith('bdp-') and o['directoryAbsent'] is True and not directory.exists(),'Owned store cleanup differs')
  for response in o['completedResponses']:
   require(response['method'] in ['GET','HEAD','POST'] and isinstance(response['status'],int) and isinstance(response['bodyOctets'],int) and response['bodyOctets']>=0,'Malformed response inventory')
   for field in ['wireSha256','bodySha256']:require(isinstance(response[field],str) and re.fullmatch('[0-9a-f]{64}',response[field]),'Response fingerprint absent')
  b=o['binding'];served=b['served'];require(served['sourceCommit']==g['finalHead'] and served['profile']=='transactional' and served['history'] is False and served['scope']=='https://reference.invalid/acme/' and b['authorityMode']=='canonical','Binding differs')
  for field in ['installedPayloadDigest','installedContractsDigest','policyDigest','authenticationDigest']:require(re.fullmatch('[0-9a-f]{64}',served[field]),'Binding digest absent')
  require(re.fullmatch('[0-9a-f]{64}',b['runtimeConfigurationIdentity']),'Runtime configuration binding absent')
  fixture_path=W/'packages/conformance/test-support/transactional/fixture-v1.json';fixture=read(fixture_path)
  require(served['fixtureDigest']==sha(fixture_path.read_bytes()) and served['fixtureId']==fixture['id'] and served['capabilities']==sorted(fixture['capabilities']),'Fixture binding differs')
  proof=verify_consumer(g);entry=Path(proof['entryRealpath'])
  require(Path(o['entry']).resolve(strict=True)==entry and o['entrySha256']==proof['entrySha256'] and served['installedPayloadDigest']==proof['payloadIdentity'],'Child differs from independently verified external payload')
  responses=o['completedResponses'];contextNames=['bdp-scope-epoch','bdp-authorization-view','bdp-scope-position']
  if mode=='minimum':
   require(o['launchStarted'] is True and o['closeAttempted'] is True and o['ownership']=='closed','Owned clean close absent')
   require([(r['method'],r['status']) for r in responses]==[('GET',200),('GET',503),('HEAD',503)],'Initial finite Event exchange sequence differs')
   setup,get,head=responses;context=[setup['headers'][n] for n in contextNames];require(all(context),'Real setup context absent')
   for r in [get,head]:
    h=r['headers'];require([h[n] for n in contextNames]==context and h['cache-control']=='private, no-store' and h['retry-after']=='1' and h['x-bdp-development-limit'] is None and h['x-bdp-development']=='non-attesting','Timeout context/retry differs')
   require(get['bodyOctets']>0 and head['bodyOctets']==0 and head['bodySha256']==sha(b''),'GET/HEAD body differs')
   require(get['headers']['content-type'].split(';',1)[0]=='application/problem+json' and head['headers']['content-type']==get['headers']['content-type'] and head['headers']['content-length']==str(get['bodyOctets']),'HEAD paired GET metadata differs')
  else:
   scope='https://reference.invalid/acme/'
   expected=[('POST',scope+'operations/batch','application/json',200),('GET',scope+'events/','application/json',200)]
   expected += [(method,scope+collection+'/missing?view=events',accept,404) for collection in ['beads','links'] for accept in ['application/json','text/event-stream'] for method in ['GET','HEAD']]
   expected += [('GET',scope+'beads/missing?view=events','text/event-stream',401)]
   require([(r['method'],r['url'],r['accept'],r['status']) for r in responses]==expected,'Historical request matrix differs')
   context=[responses[1]['headers'][n] for n in contextNames];require(all(context),'Historical context absent')
   for r in responses[2:10]:
    h=r['headers'];require([h[n] for n in contextNames]==context and h['content-type']=='application/problem+json' and h['cache-control']=='private, no-store' and all(h[n] is None for n in ['retry-after','www-authenticate','x-bdp-development-limit']),'Historical missing metadata differs')
    if r['method']=='HEAD':require(r['bodyOctets']==0 and r['bodySha256']==sha(b''),'Historical HEAD body differs')
   require(all(responses[-1]['headers'][n] is None for n in contextNames),'Unauthenticated context exposed')
  report.update({'attribution':'Verified external candidate; base HEAD plus exact dirty source/build/archive/consumer binding, not child source attestation','observation':o,'storeCleanup':'Reported clean exit, actual stage-owned directory absent, and supervisor PID/store census accepted. Historical unconditional deletion is not independently sufficient for clean exit.','limits':'Problem schema/exact-body and non-retained validator assertions ran inside the pinned callback. No invented elapsed-wait or raw request census beyond reported responses; no conformance/profile credit.'})
 write(stage+'.accepted-observation.json',report)
