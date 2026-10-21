"""Read actual safe installed sequence report and exact callback, no product imports."""
from gate import *
import re
name='sequence-installed';g,spec=load_gate('review-sequence-installed');x=completed(name);proof=json.loads(proof_for().read_text())
log=E/(name+'.log');result=E/(name+'.results.json');rows=json.loads((P/'sequence-requests.json').read_text());observations=[]
for line in log.read_text().splitlines():
 if line.startswith('{'):
  try:v=json.loads(line)
  except json.JSONDecodeError:continue
  if v.get('observation')=='private-installed-sequence-depth':observations.append(v)
require(len(observations)==1,'Missing unique installed sequence observation');o=observations[0]
require(o['outcome']=='passed' and o['failureCount']==0 and o['attempted']==o['exchanges']==8 and o['claimEligible'] is False,'Wrong outcome/count')
require(o['workspaceDirty'] is False and o['declaredSourceCommit']==g['finalHead'],'Wrong source attribution')
C=Path(proof['consumer']).resolve();entry=(C/'node_modules/@bdp/server/dist/transaction-installed-process.js').resolve(strict=True)
require(o['entry']==str(entry) and o['entrySha256']==sha(entry.read_bytes()),'Wrong installed executable')
for k in ['closed','launchAttempted','startupCompleted','stopKnown','directoryAbsent']:require(o[k] is True,'Incomplete lifecycle '+k)
require(not o['deadlineExceeded'] and not o['cleanupExceeded'] and o['childExitCode']==0 and o['childSignalCode'] is None,'Unclean child')
require(o['childPid'] not in {int(v) for v in subprocess.check_output(['ps','-axo','pid='],text=True).split()},'Child remains')
require(not Path(o['directory']).exists() and Path(o['directory']).name.startswith('bdp-sequence-depth-installed-'),'Store remains')
served=o['binding']['served'];require(served['sourceCommit']==g['finalHead'] and served['installedPayloadDigest']==proof['payloadIdentity'] and served['scope']=='https://reference.invalid/acme/' and served['profile']=='transactional' and served['history'] is False,'Served identity mismatch')
fixture=W/'packages/conformance/test-support/transactional/fixture-v1.json';require(served['fixtureDigest']==sha(fixture.read_bytes()) and served['fixtureId']==json.loads(fixture.read_text())['id'],'Fixture binding mismatch')
require(len(o['completedResponses'])==8,'Missing responses')
for actual,expected in zip(o['completedResponses'],rows,strict=True):
 for a,b in [('ordinal','index'),('label','label'),('method','method'),('url','url'),('accept','accept'),('key','key'),('status','status'),('requestBodyBytes','bodyBytes'),('requestBodySha256','bodySha256')]:require(actual[a]==expected[b],'Ordered request/response differs '+a)
 require(actual['bodyOctets']<=65536 and re.fullmatch('[a-f0-9]{64}',actual['bodySha256']),'Response framing')
r=json.loads(result.read_text());active=[c for f in r['testResults'] for c in f.get('assertionResults',[]) if c['status'] not in ['pending','skipped','todo']]
require(len(active)==1 and active[0]['status']=='passed' and active[0]['fullName']=='observes retained sequence member conflict before private depth refusal over packed bytes','Wrong callback')
require(r['success'] and r['numPassedTests']==1 and r['numFailedTests']==0 and not r.get('unhandledErrors') and not r.get('errors') and not r.get('numRuntimeErrorTestSuites',0),'Unrelated runtime failure')
verify_build(g)
record={'accepted':True,'actualHTTP':8,'actualChildExit':0,'observation':o,'proofSha256':sha(proof_for().read_bytes()),'resultSha256':sha(result.read_bytes()),'logSha256':sha(log.read_bytes()),'limitation':'Full schema/receipt/Resource equality is asserted in the selected test; persisted report retains fingerprints, not raw sensitive bodies. No public qualification.'}
with (E/'sequence-installed.adjudication.json').open('x') as f:f.write(json.dumps(record,indent=2)+'\n')
print(json.dumps({'accepted':True,'actualHTTP':8,'childPidAbsent':o['childPid']}))
