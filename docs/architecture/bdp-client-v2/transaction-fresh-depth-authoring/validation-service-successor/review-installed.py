"""Read actual safe installed cohort report and exact callback, no product imports."""
from gate import *
import re,sys
name=sys.argv[1];oracle=json.loads((P/'installed-oracles.json').read_text())[name];g,spec=load_gate('review-'+name);x=completed(name);proof=json.loads(proof_for().read_text())
require(oracle['sha256']==sha((W/oracle['file']).read_bytes()),'Installed oracle source binding pending/changed')
log=E/(name+'.log');result=E/(name+'.results.json');rows=oracle['requests'];observations=[]
for line in log.read_text().splitlines():
 if line.startswith('{'):
  try:v=json.loads(line)
  except json.JSONDecodeError:continue
  if v.get('observation')==oracle['observation']:observations.append(v)
require(len(observations)==1,'Missing unique installed observation');o=observations[0]
require(o['outcome']=='passed' and o['failureCount']==0 and o['attempted']==o['exchanges']==oracle['http'] and o['claimEligible'] is False,'Wrong outcome/count')
require(o['workspaceDirty'] is False and o['declaredSourceCommit']==g['finalHead'],'Wrong source attribution')
C=Path(proof['consumer']).resolve();entry=(C/'node_modules/@bdp/server/dist/transaction-installed-process.js').resolve(strict=True)
require(o['entry']==str(entry) and o['entrySha256']==sha(entry.read_bytes()),'Wrong installed executable')
for k in ['closed','launchAttempted','startupCompleted','stopKnown','directoryAbsent']:require(o[k] is True,'Incomplete lifecycle '+k)
require(not o['deadlineExceeded'] and not o['cleanupExceeded'] and o['childExitCode']==0 and o['childSignalCode'] is None,'Unclean child')
require(o['childPid'] not in {int(v) for v in subprocess.check_output(['ps','-axo','pid='],text=True).split()},'Child remains')
require(not Path(o['directory']).exists() and Path(o['directory']).name.startswith(oracle['directoryPrefix']),'Store remains')
served=o['binding']['served'];require(served['sourceCommit']==g['finalHead'] and served['installedPayloadDigest']==proof['payloadIdentity'] and served['scope']=='https://reference.invalid/acme/' and served['profile']=='transactional' and served['history'] is False,'Served identity mismatch')
fixture=W/'packages/conformance/test-support/transactional/fixture-v1.json';require(served['fixtureDigest']==sha(fixture.read_bytes()) and served['fixtureId']==json.loads(fixture.read_text())['id'],'Fixture binding mismatch')
require(len(o['completedResponses'])==oracle['http'],'Missing responses')
for actual,expected in zip(o['completedResponses'],rows,strict=True):
 for a,b in [('ordinal','index'),('label','label'),('method','method'),('accept','accept'),('key','key'),('status','status'),('requestBodyBytes','bodyBytes'),('requestBodySha256','bodySha256')]:require(actual[a]==expected[b],'Ordered request/response differs '+a)
 if name=='fresh-installed':require(actual.get('requestBodyPresent') is (expected['bodyRef'] is not None),'Wrong actual request body presence')
 if expected['url'] is not None:require(actual['url']==expected['url'],'Wrong actual target')
 else:
  require(name=='fresh-installed' and actual['ordinal']==13 and re.fullmatch(r'https://reference\.invalid/acme/receipts/[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}',actual['url']),'Actual receipt target absent/invalid')
  require(actual['targetBytes']==len(actual['url'].encode()) and actual['targetSha256']==sha(actual['url'].encode()),'Actual dynamic target framing changed')
 if expected['contextPosition'] is not None:require(actual['headers']['bdp-scope-position']==expected['contextPosition'],'Wrong response context position')
 require(type(actual['bodyOctets']) is int and 0<=actual['bodyOctets']<=oracle['responseBodyCap'] and re.fullmatch('[a-f0-9]{64}',actual['bodySha256']),'Response framing')
if name=='fresh-installed':
 require(o['qualifiedEntryRequested'] is True and o['failureCategories']==[],'Development fallback or failure')
 require(o['prerequisites']==['genuine-deep-owned-link-created','full-receipt-resource-Event-Snapshot-oracles','no-op-replay-rollback-position-stability'],'Missing nonvacuous oracles')
 sizes=[v['bodyOctets'] for v in o['completedResponses']]
 require(o['responseBytesTotal']==sum(sizes) and o['responseBytesMaximum']==max(sizes),'Unbound response measurements')
r=json.loads(result.read_text());active=[c for f in r['testResults'] for c in f.get('assertionResults',[]) if c['status'] not in ['pending','skipped','todo']]
require(len(active)==1 and active[0]['status']=='passed' and active[0]['fullName']==oracle['fullName'],'Wrong callback')
require(r['success'] and r['numPassedTests']==1 and r['numFailedTests']==0 and not r.get('unhandledErrors') and not r.get('errors') and not r.get('numRuntimeErrorTestSuites',0),'Unrelated runtime failure')
verify_build(g)
record={'accepted':True,'actualHTTP':oracle['http'],'actualChildExit':0,'observation':o,'proofSha256':sha(proof_for().read_bytes()),'resultSha256':sha(result.read_bytes()),'logSha256':sha(log.read_bytes()),'limitation':'Full schema/receipt/Resource equality is asserted in the selected test; persisted report retains fingerprints, not raw sensitive bodies. No public qualification.'}
with (E/(name+'.adjudication.json')).open('x') as f:f.write(json.dumps(record,indent=2)+'\n')
print(json.dumps({'accepted':True,'actualHTTP':oracle['http'],'childPidAbsent':o['childPid']}))
