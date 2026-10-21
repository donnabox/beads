"""Standard-library collection validation auditor; no product imports."""
import sys
sys.dont_write_bytecode=True
from gate import *
import re
mode=sys.argv[1];stage={'source':'focused-source','external':'collection-external'}.get(mode)
g,_=load_gate({'build':'bind-fresh-build','source':'review-focused-source','external':'review-collection-external'}[mode],child=True)
if mode=='build':
 write('build-artifacts.json',{'head':g['finalHead'],'tree':g['finalTree'],'sourceIdentity':g['_sourceIdentity'],'artifacts':inventory(),**{key:{'path':str(E/('fresh-build.'+suffix)),'sha256':sha((E/('fresh-build.'+suffix)).read_bytes())} for key,suffix in [('start','started.json'),('exit','exit.json'),('log','log')]}})
else:
 config=read(P/'commands.json')['commands'][stage];result=read(E/(stage+'.results.json'))
 cases=[x for suite in result['testResults'] for x in suite['assertionResults']]
 chosen=[x for x in cases if x['fullName'] in config['selectedFullNames']]
 require(len(chosen)==len(config['selectedFullNames']) and sorted(x['fullName'] for x in chosen)==sorted(config['selectedFullNames']) and all(x['status']=='passed' for x in chosen) and result.get('success') is True,'Exact selected callbacks did not pass')
 require(all(x['status'] in ['pending','skipped','todo'] for x in cases if x not in chosen),'Unexpected callback executed')
 report={'sourceIdentity':g['_sourceIdentity'],'sourceBindingSha256':g['sourceBinding']['sha256'],'qualification':False,'selectedCallbacks':config['selectedFullNames'],'resultsSha256':sha((E/(stage+'.results.json')).read_bytes()),'logSha256':sha((E/(stage+'.log')).read_bytes())}
 if mode=='source':
  report.update({'attribution':'source-private callback checks','limits':'Exact controlled attempts, waits, owner availability, no token allocation, cancellation and HTTP semantics are assertions of the pinned passing callbacks; not fabricated runtime telemetry.'})
 else:
  observations=[]
  for line in (E/(stage+'.log')).read_text().splitlines():
   try:x=json.loads(line.strip())
   except json.JSONDecodeError:continue
   if isinstance(x,dict) and 'observation' in x:observations.append(x)
  matches=[x for x in observations if x['observation']==config['expectedObservation']];require(len(matches)==1,'Missing/duplicate callback report');o=matches[0]
  require(o['claimEligible'] is False and o['outcome']=='passed' and o['aborted'] is False and o['closed'] is True and o['workspaceDirty'] is True and o['declaredSourceCommit']==g['finalHead'],'Outcome/source/cleanup differs')
  require(isinstance(o['childPid'],int) and o['childPid']>0 and o['childExitCode']==0,'Child exit absent')
  execution=complete(stage,g);require(o['childPid'] in execution['sampledPids'],'Reported child was not observed by supervisor')
  require(o['exchanges']==42 and len(o['completedResponses'])==42,'Expected 42 exchanges')
  for response in o['completedResponses']:
   require(response['method'] in ['GET','HEAD','POST'] and isinstance(response['status'],int) and response['url'].startswith('https://reference.invalid/acme/'),'Malformed response inventory')
   for field in ['wireSha256','bodySha256']:require(re.fullmatch('[0-9a-f]{64}',response[field]),'Response fingerprint absent')
  b=o['binding'];served=b['served'];require(served['sourceCommit']==g['finalHead'] and served['profile']=='transactional' and served['history'] is False and served['scope']=='https://reference.invalid/acme/' and b['authorityMode']=='canonical','Binding differs')
  for field in ['installedPayloadDigest','installedContractsDigest','policyDigest','authenticationDigest']:require(re.fullmatch('[0-9a-f]{64}',served[field]),'Binding digest absent')
  require(re.fullmatch('[0-9a-f]{64}',b['runtimeConfigurationIdentity']),'Runtime configuration binding absent')
  fixture_path=W/'packages/conformance/test-support/transactional/fixture-v1.json';fixture=read(fixture_path)
  require(served['fixtureDigest']==sha(fixture_path.read_bytes()) and served['fixtureId']==fixture['id'] and served['capabilities']==sorted(fixture['capabilities']),'Fixture binding differs')
  proof=verify_consumer(g);entry=Path(proof['entryRealpath'])
  require(Path(o['entry']).resolve(strict=True)==entry and o['entrySha256']==proof['entrySha256'] and served['installedPayloadDigest']==proof['payloadIdentity'],'Actual child differs from independently verified external payload')
  expected=[(c,m) for c in ['beads','links'] for m in ['GET','HEAD']]
  unknown=[x for x in observations if x['observation']=='installed-unknown-collection-minimum-timeout']
  require([(x['collection'],x['method']) for x in unknown]==expected and all(x['claimEligible'] is False and 1000<=x['elapsedMs']<5000 for x in unknown),'Unknown-minimum timing observations differ')
  known=[x for x in observations if x['observation']=='installed-minimum-timeout']
  require([(x['collection'],x['method']) for x in known]==expected,'Historical recognized-minimum observations missing')
  require(len([x for x in observations if x['observation']=='installed-position-oracle-control'])==2,'Existing position controls missing')
  report.update({'attribution':'verified external candidate from frozen four-file dirty draft; base HEAD is chronology, not complete source identity','observation':o,'unknownMinimumTimings':unknown,'recognizedMinimumTimings':known,'storeCleanup':'Supervisor observed no surviving bdp-prefixed directory beneath the owned stage TMPDIR; callback reports close after awaited child.close and removal. No store path was emitted by this historical helper.','limits':'Existing helper reports one child PID and normal exit code, not a signal field. Exact 503 Problem, HEAD suppression/content length, context and conditional precedence are assertions of the pinned passing callback. No complete profile or public qualification claim.'})
 write(stage+'.accepted-observation.json',report)
