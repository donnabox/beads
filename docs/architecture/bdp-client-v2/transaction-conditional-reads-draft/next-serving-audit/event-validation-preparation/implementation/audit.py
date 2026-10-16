"""Standard-library artifact/result checkpoints; never imports product."""
import sys
sys.dont_write_bytecode=True
from gate import *
import sys,re,urllib.parse
mode=sys.argv[1];names={'tools':'bind-workspace-tools','build':'bind-fresh-build','historical':'review-historical','default':'review-restart-default','external':'review-restart-external'}
g,_=load_gate(names[mode],child=True)
if mode=='tools':
 initial=read(PLAN/'source-binding.json');tools=dict(initial['tools'])
 paths=[W/'node_modules'/n for n in ['@biomejs/biome/bin/biome','vitest/vitest.mjs','typescript/bin/tsc','typescript/lib/tsc.js','typescript/lib/getExePath.js','typescript/package.json']]
 paths += [W/'node_modules/.pnpm/@biomejs+cli-darwin-arm64@2.5.7/node_modules/@biomejs/cli-darwin-arm64/biome',W/'node_modules/.pnpm/@typescript+typescript-darwin-arm64@7.0.2/node_modules/@typescript/typescript-darwin-arm64/lib/tsc']
 for path in map(str,paths):
  p=Path(path);tools[path]={'realpath':str(p.resolve(strict=True)),'sha256':sha(p.read_bytes())}
 versions={'@biomejs/biome':'2.5.7','vitest':'4.1.10','typescript':'7.0.2'}
 for name,version in versions.items():require(read(W/'node_modules'/name/'package.json')['version']==version,'Tool version differs')
 write('workspace-tools.json',{'tools':tools,'versions':versions,'sourceLockSha256':g['sourceLockSha256'],'sourceHeadAtBootstrap':g['finalHead']})
elif mode=='build':
 write('build-artifacts.json',{'head':g['finalHead'],'tree':g['finalTree'],'sourceIdentity':g['_sourceIdentity'],'artifacts':inventory(),**{key:{'path':str(E/('fresh-build.'+suffix)),'sha256':sha((E/('fresh-build.'+suffix)).read_bytes())} for key,suffix in [('start','started.json'),('exit','exit.json'),('log','log')]}})
else:
 stage={'historical':'historical-default','default':'restart-default','external':'restart-external'}[mode]
 config=read(P/'commands.json')['commands'][stage];result=read(E/(stage+'.results.json'))
 cases=[x for suite in result['testResults'] for x in suite['assertionResults']]
 chosen=[x for x in cases if x['fullName'] in config['selectedFullNames']]
 require(len(chosen)==1 and chosen[0]['status']=='passed' and result.get('success') is True,'Selected callback did not pass')
 require(all(x['status'] in ['pending','skipped','todo'] for x in cases if x not in chosen),'Unexpected callback executed')
 observations=[]
 for line in (E/(stage+'.log')).read_text().splitlines():
  try:x=json.loads(line.strip())
  except json.JSONDecodeError:continue
  if isinstance(x,dict) and x.get('observation')==config['expectedObservation']:observations.append(x)
 require(len(observations)==1,'Missing/duplicate callback observation');o=observations[0]
 require(o['claimEligible'] is False and o['node']=='v24.16.0','Claim/node differs')
 b=o['binding'];served=b['served'];require(served['sourceCommit']==g['finalHead'] and served['profile']=='transactional' and served['history'] is False and served['scope']=='https://reference.invalid/acme/' and b['authorityMode']=='canonical','Binding differs')
 for field in ['installedPayloadDigest','installedContractsDigest','policyDigest','authenticationDigest']:
  require(isinstance(served[field],str) and re.fullmatch('[0-9a-f]{64}',served[field]),'Binding digest missing '+field)
 require(re.fullmatch('[0-9a-f]{64}',b['runtimeConfigurationIdentity']),'Runtime binding absent')
 fixture_path=W/'packages/conformance/test-support/transactional/fixture-v1.json';fixture=read(fixture_path)
 require(served['fixtureDigest']==sha(fixture_path.read_bytes()) and served['fixtureId']==fixture['id'] and served['capabilities']==sorted(fixture['capabilities']),'Fixture binding differs')
 expected_entry=(W/'packages/server/dist/transaction-installed-process.js').resolve(strict=True)
 if mode=='external':
  proof=verify_consumer(g);expected_entry=Path(proof['entryRealpath']);require(served['installedPayloadDigest']==proof['payloadIdentity'],'Actual child payload differs from verified consumer')
 if mode=='external':
  earlier=read(E/'restart-default.accepted-observation.json')['observation']
  require(served['installedContractsDigest']==earlier['binding']['served']['installedContractsDigest'],'Same frozen schema artifacts produced different contracts')
 require(Path(o['entry']).resolve(strict=True)==expected_entry and o['entrySha256']==sha(expected_entry.read_bytes()),'Actual entry differs')
 if mode=='historical':
  require(o['source']==g['finalHead'] and o['workspaceDirty'] is True,'Historical source differs')
  require(len(o['observations'])==4 and o['idleMilliseconds']>=15000 and o['comments']>=2 and o['liveEventFrames']==1,'Historical observation incomplete')
 else:
  require(o['outcome']=='passed' and o['failures']==[] and o['pendingOperationsAtReport']==0 and o['directoryRemoved'] is True,'Restart/cleanup failed')
  require(o['attribution']==('external-installed-candidate' if mode=='external' else 'development-local'),'Wrong attribution')
  require(o['declaredSourceCommit']==g['finalHead'],'Declared source differs')
  require(o['witnessSource']=={'head':g['finalHead'],'fileSha256':WITNESS,'diffSha256':sha(git('diff','HEAD')),'dirtyPaths':[FILE]},'Witness identity differs')
  require(o['deadlines']=={'work':45000,'cleanup':10000,'callback':60000,'outerSupervisor':120000,'finite':5000,'recorder':5000},'Deadline changed')
  require(o['seedBodyBytes']==244 and o['seedBodySha256']=='2ffd2dc1c98cf8ec53338127cfad4edba630ac386b22f6c693b3931626559f03' and o['seedKey']=='event-restart-two-created-v1','Seed differs')
  require(o['localSensitivityControls']==['restored-event-ordinal','scope-prefix-frame'],'Controls absent')
  sessions=o['sessions'];require(len(sessions)==2 and [s['create'] for s in sessions]==[True,False],'Wrong sessions')
  require(all(isinstance(s['pid'],int) and s['pid']>0 for s in sessions) and sessions[0]['pid']!=sessions[1]['pid'],'PIDs absent/not distinct')
  for i,s in enumerate(sessions):require(s['ordinal']==i+1 and s['state']=='owned' and s['closed'] is True and s['stopUnknown'] is False and s['exit']=={'code':0,'signal':None} and s['binding']==b,'Child exit/binding differs')
  require([{k:r[k] for k in ['ordinal','session','method']} for r in o['requests']]==[{'ordinal':i+1,'session':1 if i<2 else 2,'method':'POST' if i==0 else 'GET'} for i in range(5)],'Request sequence differs')
  responses=o['completedFiniteResponses'];require([r['ordinal'] for r in responses]==[1,2,3,5] and [r['status'] for r in responses]==[200,200,200,400],'Finite response inventory differs')
  events=o['events'];require(len(events)==2 and events[0]['id']!=events[1]['id'],'Missing distinct Event IDs')
  sse=o['sse'];require(sse['requestOrdinal']==4 and sse['session']==2 and sse['lastEventId']==events[0]['id'] and sse['status']==200 and sse['drained'] is True and 0<=sse['beforeIntentionalCloseMs']<5000,'SSE boundary differs')
  require(urllib.parse.parse_qs(urllib.parse.urlsplit(sse['target']).query).get('after')==[events[1]['id']],'Conflicting query missing')
  frames=sse['completeObservedFrames'];require(len(frames)==1,'Wrong complete frame count')
  frame=frames[0].split('\n');require('id: '+events[1]['id'] in frame and 'event: '+events[1]['type'] in frame,'SSE fields differ')
  data=[x[6:] for x in frame if x.startswith('data: ')];require(len(data)==1 and json.loads(data[0])==events[1],'SSE complete Event differs')
  context=responses[0]['headers'];require(all(r['headers']['bdp-scope-position']=='pos-1' and all(r['headers'][k]==context[k] for k in ['bdp-scope-epoch','bdp-authorization-view']) for r in responses),'Finite context differs')
  require(all(sse['headers'][k]==context[k] for k in ['bdp-scope-epoch','bdp-authorization-view','bdp-scope-position']) and sse['headers']['content-type'].split(';')[0]=='text/event-stream' and sse['headers']['cache-control']=='no-store, no-transform','SSE context/media differs')
  require(o['streamAtReport']['drained'] is True and o['streamAtReport']['destroyed'] is True and o['streamAtReport']['completeObservedFrames']==frames,'Reported drain differs')
  directory=Path(o['directory']);require(directory.is_relative_to(E/(stage+'.tmp')) and not directory.exists(),'Store not removed/owned')
 write(stage+'.accepted-observation.json',{'observation':o,'attribution':'development-local' if mode!='external' else 'verified-external-candidate','qualification':False,'sourceIdentity':g['_sourceIdentity'],'resultsSha256':sha((E/(stage+'.results.json')).read_bytes()),'logSha256':sha((E/(stage+'.log')).read_bytes()),'limits':'Only complete frames observed in bounded window. Complete Resource/Event/schema and synchronous pre-close liveness assertions are established by the pinned callback passing; finite JSON bodies are not fabricated from fingerprints.'})
