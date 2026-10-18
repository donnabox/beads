"""Exact safe observations and actual Vitest verdicts; never accepts generic exit1."""
from pathlib import Path
import json,re,sys,os,subprocess
from gate import P,D,W,E,sha,require,load_gate,proof_for,verify_build
mode='positive';g,spec=load_gate('review-minimum-installed');proof=json.loads(proof_for(mode).read_text())
log=E/'minimum-installed.log';result=E/'minimum-installed.results.json';x=json.loads((E/'minimum-installed.exit.json').read_text())
require(x['exit']==(0 if mode=='positive' else 1) and x['supervisorExit']==0,'Wrong raw/supervisor exit')
observations=[]
for line in log.read_text().splitlines():
 if line.startswith('{'):
  try:v=json.loads(line)
  except json.JSONDecodeError:continue
  if v.get('observation')=='private-installed-minimum-current-resource':observations.append(v)
require(len(observations)==1,'Exactly one real observation required');o=observations[0]
keys='observation mode controlPrerequisites publishedView wrapper wrapperSha256 outcome declaredSourceCommit workspaceDirty entry entrySha256 binding artifactBytes attempted exchanges completedResponses deadlineExceeded cleanupExceeded launchAttempted startupCompleted closed childPid childExitCode childSignalCode directory directoryAbsent failureCount failureCategories claimEligible retention startupLimitation'.split()
require(set(o)==set(keys),'Unexpected observation fields')
require(o['mode']==mode and o['declaredSourceCommit']==g['finalHead'] and o['workspaceDirty']==False and o['claimEligible']==False,'Source/mode/claim mismatch')
require(o['entry']==proof['entryRealpath'] and o['entrySha256']==proof['entrySha256'],'Entry mismatch')
require(all(o[k] is True for k in ['launchAttempted','startupCompleted','closed','directoryAbsent']),'Lifecycle incomplete')
require(o['childExitCode']==0 and o['childSignalCode'] is None and not o['deadlineExceeded'] and not o['cleanupExceeded'],'Unclean child')
require(isinstance(o['childPid'],int) and o['childPid']>1,'Missing child PID')
live={int(v.strip()) for v in subprocess.check_output(['ps','-axo','pid='],text=True).splitlines()}
require(o['childPid'] not in live,'Child still present')
directory=Path(o['directory']);require(directory.name.startswith('bdp-minimum-current-installed-') and not directory.exists(),'Owned store remains')
require(Path(o['wrapper'])==directory/'private-bootstrap.mjs','Wrapper location mismatch')
test=W/'apps/bdptest/src/transactional-minimum-current-resource-installed.test.ts'
source=test.read_text();encoded=re.search(r'const bootstrap =\s*("(?:[^"\\]|\\.)*")\s*\.replace',source).group(1)
bootstrap=json.loads(encoded).replace('__ENTRY__',json.dumps(proof['entryRealpath'],separators=(',',':'))).replace('__SHA__',json.dumps(proof['entrySha256']))
require(sha(bootstrap.encode())==o['wrapperSha256'],'Wrapper source/entry binding mismatch')
binding=o['binding'];require(set(binding)==set(['served','operationTargets','runtimeConfigurationIdentity','authorityMode']),'Binding field mismatch')
served=binding['served'];require(set(served)==set('scope profile history sourceCommit installedPayloadDigest installedContractsDigest policyDigest authenticationDigest fixtureDigest fixtureId capabilities'.split()),'Unexpected served fields')
scope='https://event-sse.invalid/acme/'
require(served['scope']==scope and served['profile']=='transactional' and served['history']==False and served['sourceCommit']==g['finalHead'] and served['installedPayloadDigest']==proof['payloadIdentity'],'Served payload/source mismatch')
fixture=W/'packages/conformance/test-support/transactional/fixture-v1.json';fv=json.loads(fixture.read_text())
require(served['fixtureDigest']==sha(fixture.read_bytes()) and served['fixtureId']==fv['id'] and served['capabilities']==sorted(fv['capabilities']),'Fixture differs')
for k in ['installedContractsDigest','policyDigest','authenticationDigest']:require(isinstance(served[k],str) and re.fullmatch('[a-f0-9]{64}',served[k]) is not None,'Invalid safe identity')
require(binding['authorityMode']=='canonical' and re.fullmatch('[a-f0-9]{64}',binding['runtimeConfigurationIdentity']) is not None,'Runtime binding')
targets={'createBead':'create-bead','updateBead':'update-bead','deleteBead':'delete-bead','createLink':'create-link','updateLink':'update-link','deleteLink':'delete-link','putAlias':'put-alias','deleteAlias':'delete-alias','updateWhere':'update-where','deleteWhere':'delete-where','batch':'batch','sequence':'sequence'}
require(binding['operationTargets']=={k:scope+'operations/'+v for k,v in targets.items()},'Targets differ')
ledger=json.loads((Path(g['requestsPath'])).read_text());rows=ledger['rows']
selected=rows if mode=='positive' else [rows[0],rows[1 if mode=='restore-missing-gap' else 10]]
n=len(selected);require(o['attempted']==o['exchanges']==n and len(o['completedResponses'])==n,'Exchange count')
require(o['artifactBytes']==378,'Artifact byte count')
headers='bdp-scope-epoch bdp-authorization-view bdp-scope-position content-type content-length cache-control retry-after www-authenticate x-bdp-development x-bdp-development-limit etag last-modified allow location link'.split()
responseKeys='ordinal label method url credential requestBodyBytes requestBodySha256 minimumFieldOccurrences status bodyOctets bodySha256 headers'.split()
contexts=[]
for ordinal,(r,row) in enumerate(zip(o['completedResponses'],selected),1):
 require(set(r)==set(responseKeys) and set(r['headers'])==set(headers),'Unexpected response fields')
 h=r['headers'];require(all(v is None or isinstance(v,str) for v in h.values()),'Header shape')
 require(r['ordinal']==ordinal and r['method']==row['method'] and r['url']==scope+row['target'],'Ordered actual request mismatch')
 require(r['credential']==('anonymous' if row['index']==30 else 'writer'),'Safe credential role')
 require(r['requestBodyBytes']==(610 if ordinal==1 else 0) and r['requestBodySha256']==(ledger['seed']['bodySha256'] if ordinal==1 else None),'Actual request fingerprint')
 require(r['minimumFieldOccurrences']==([0,0,0] if ordinal==1 else [1,0,0] if row['index']==32 else [1,1,1]),'Minimum occurrence mismatch')
 require(isinstance(r['bodyOctets'],int) and 0<=r['bodyOctets']<=32768 and re.fullmatch('[a-f0-9]{64}',r['bodySha256']) is not None,'Body bound/hash')
 context=[h[k] for k in headers[:3]];contexts.append(context)
 if ordinal==1:
  require(r['status']==200 and r['bodyOctets']>0 and context[2]=='pos-3' and re.fullmatch('[a-f0-9-]{36}',context[0]) and context[1]=='development-all-visible','Actual seed context')
  continue
 expected=row['status'] if mode=='positive' else 500;require(r['status']==expected,'Status differs')
 expectedLabel='minimum-current:missing-before-minimum' if row['index']==2 else 'minimum-current:unknown-catch-up-timeout' if row['index']==11 else 'minimum-current:'+str(row['index'])
 require(r['label']==expectedLabel,'Named oracle differs')
 if mode!='positive':
  require(context==[None]*3 and r['bodyOctets']==0 and r['bodySha256']==sha(b'') and h['x-bdp-development-limit']==('minimum-position-visibility' if mode=='restore-missing-gap' else 'minimum-position'),'Control prerequisite differs')
 elif expected in (404,503):
  require(context==[contexts[0][0],o['publishedView'] if row['index']>=18 else contexts[0][1],'pos-3'],'Actual response context')
  problem={'type':'https://github.com/gastownhall/bdp/problems/'+('not-found' if expected==404 else 'unavailable'),'code':'resource-not-found' if expected==404 else 'catch-up-timeout','retry':'after-state-change' if expected==404 else 'after-delay','status':expected}
  body=json.dumps(problem,separators=(',',':')).encode()
  require(h['content-type']=='application/problem+json' and h['cache-control']=='private, no-store' and h['content-length']==str(len(body)),'Problem representation headers')
  require(r['bodyOctets']==(0 if row['method']=='HEAD' else len(body)) and r['bodySha256']==sha(b'' if row['method']=='HEAD' else body),'Exact Problem body fingerprint')
  require(all(h[k] is None for k in ['etag','last-modified','location','x-bdp-development-limit']),'Disclosure/validator header')
  if expected==503:require(h['retry-after']=='1','Retry header')
 elif expected==200:require(context==contexts[0] and r['bodyOctets']>0,'Known Resource context')
 else:
  require(context==[None]*3,'Pre-owner context leaked')
  if expected==405:require(r['bodyOctets']==0 and h['allow']=='GET, HEAD','Method refusal')
  if expected==401:require(h['www-authenticate']=='Bearer realm="bdp-transaction-development"','Challenge literal')
report=json.loads(result.read_text());cases=[c for f in report['testResults'] for c in f.get('assertionResults',[])];active=[c for c in cases if c['status'] not in ['pending','skipped','todo']]
require(len(cases)==len(active)==1 and active[0]['fullName']=='serves missing-first current Resources and bounds unproved minimum eligibility','Exact callback mismatch')
require(not report.get('unhandledErrors') and not report.get('errors') and not report.get('numRuntimeErrorTestSuites',0),'Unrelated Vitest error')
if mode=='positive':
 require(report['success'] and report['numPassedTests']==1 and report['numFailedTests']==0 and active[0]['status']=='passed','Positive test failed')
 require(o['outcome']=='passed' and o['failureCount']==0 and o['failureCategories']==[] and o['controlPrerequisites']==False,'Positive failure')
 require(isinstance(o['publishedView'],str) and re.fullmatch('[a-f0-9-]{36}',o['publishedView']) is not None and o['publishedView']!=contexts[0][1],'No actual publication')
 require(o['completedResponses'][17]['bodySha256']==o['completedResponses'][28]['bodySha256'],'Hidden/missing body differs')
else:
 label='minimum-current:missing-before-minimum' if mode=='restore-missing-gap' else 'minimum-current:unknown-catch-up-timeout'
 require(report['success']==False and report['numPassedTests']==0 and report['numFailedTests']==1 and active[0]['status']=='failed','Wrong negative verdict')
 require(len(active[0]['failureMessages'])==1 and label in active[0]['failureMessages'][0] and 'expected 500 to be '+('404' if mode=='restore-missing-gap' else '503') in active[0]['failureMessages'][0],'Wrong named negative failure')
 require(o['outcome']=='failed' and o['failureCount']==1 and o['failureCategories']==[label] and o['controlPrerequisites']==True and o['publishedView'] is None,'Negative prerequisite or error multiplicity')
verify_build(g)
record={'accepted':True,'mode':mode,'actualRawExit':x['exit'],'adapterExitExpected':0,'observedHTTP':n,'productChildren':1,'observation':o,'resultSha256':sha(result.read_bytes()),'logSha256':sha(log.read_bytes()),'exitSha256':sha((E/'minimum-installed.exit.json').read_bytes()),'proofSha256':sha(proof_for(mode).read_bytes()),'independentChildAbsent':o['childPid'],'storeAbsent':str(directory),'limitation':'Source test performs full schema/body/raw-header oracles in memory; this adapter checks actual safe records and independent package binding. Author/preparation overlap disclosed; no qualification.'}
with (E/'minimum-installed.adjudication.json').open('x') as f:f.write(json.dumps(record,indent=2)+'\n')
print(json.dumps({'accepted':True,'mode':mode,'actualHTTP':n,'childPidAbsent':o['childPid']}))
