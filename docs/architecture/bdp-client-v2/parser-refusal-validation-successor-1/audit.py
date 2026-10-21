"""Standard-library parser refusal validation auditor; no product imports."""
import sys
sys.dont_write_bytecode=True
from gate import *
import re
mode=sys.argv[1];stage={'source':'focused-source','parser':'parser-external','waiter':'waiter-external'}.get(mode)
g,_=load_gate({'tools':'bind-workspace-tools','build':'bind-fresh-build','source':'review-focused-source','parser':'review-parser-external','waiter':'review-waiter-external'}[mode],child=True)
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
 report={'sourceIdentity':g['_sourceIdentity'],'sourceBindingSha256':g['sourceBinding']['sha256'],'qualification':False,'selectedCallbacks':config['selectedFullNames'],'resultsSha256':sha((E/(stage+'.results.json')).read_bytes()),'logSha256':sha((E/(stage+'.log')).read_bytes())}
 if mode=='source':
  report.update({'attribution':'Four source-handler callbacks and three emitted-runtime fixture callbacks on one freshly bound candidate','plannedWorkload':config['prospectiveWorkload'],'limits':'100 handler invocations (67 new+33 existing),91 emitted HTTP exchanges and7 stores are source-derived planned workload, not fabricated observations. Exact refusal/no-owner/wait/token and regression semantics are assertions of the pinned passing callbacks.'})
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
  require(o['exchanges']==config['expectedExchanges'] and len(o['completedResponses'])==config['expectedExchanges'],'Exchange count differs')
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
  unknown=[];known=[]
  if mode=='parser':
   expected=[('https://reference.invalid/acme/beads/','GET',200)]
   for collection in ['beads','links']:
    for query in ['unsupported=x','limit=1&limit=2','limit=0']:
     for method in ['GET','HEAD']:expected.append((f'https://reference.invalid/acme/{collection}/?{query}',method,400))
   for control in ['absent','current']:
    for method in ['GET','HEAD']:expected.append(('https://reference.invalid/acme/beads/?unsupported=x',method,400))
   for suffix in ['types/task?view=properties','beads/a?view=links&direction=sideways','beads/missing?view=bogus']:
    for method in ['GET','HEAD']:expected.append(('https://reference.invalid/acme/'+suffix,method,400))
   expected += [('https://reference.invalid/acme/beads/?unsupported=x','GET',400),('https://reference.invalid/acme/beads/?unsupported=x','HEAD',400)]
   require([(r['url'],r['method'],r['status']) for r in o['completedResponses']]==expected,'Parser request/result sequence differs')
   require(all(r['responseBytes']==0 for r in o['completedResponses'] if r['method']=='HEAD'),'HEAD wire body is not empty')
  else:
   expected=[(c,m) for c in ['beads','links'] for m in ['GET','HEAD']]
   unknown=[x for x in observations if x['observation']=='installed-unknown-collection-minimum-timeout']
   require([(x['collection'],x['method']) for x in unknown]==expected and all(x['claimEligible'] is False and 1000<=x['elapsedMs']<5000 for x in unknown),'Unknown-minimum timing observations differ')
   known=[x for x in observations if x['observation']=='installed-minimum-timeout']
   require([(x['collection'],x['method']) for x in known]==expected,'Historical recognized-minimum observations missing')
   require(len([x for x in observations if x['observation']=='installed-position-oracle-control'])==2,'Existing position controls missing')
  report.update({'attribution':'verified external candidate from frozen three-file dirty draft; base HEAD is chronology, not complete source identity','observation':o,'unknownMinimumTimings':unknown,'recognizedMinimumTimings':known,'storeCleanup':'Supervisor observed no surviving bdp-prefixed directory beneath the owned stage TMPDIR; callback reports close after awaited child.close and removal. No store path was emitted by this historical helper.','limits':'Existing helper reports one child PID and normal exit code, not a signal field. Parser400 schema/retry/context absence and historical503/conditional/retention assertions are established by the separately pinned passing callbacks; source-owner counts are not inferred from installed wire records. No complete profile or public qualification claim.'})
 write(stage+'.accepted-observation.json',report)
