from pathlib import Path
import json,hashlib,sys
P=Path(__file__).resolve().parent; R=P/'run'; name=sys.argv[1]
filename={'cross-sequence':'cross-carrier-sequence-fresh-key'}.get(name,name)+'.json'
d=json.loads((R/filename).read_text()); pos=json.loads((R/'positive.adjudication.json').read_text()); positive=json.loads((R/'positive.observed.json').read_text()); ex=json.loads((R/(name+'.exit.json')).read_text())
summaries=[json.loads(l) for l in (R/(name+'.log')).read_text().splitlines() if l.startswith('{"kind":"installed-tx-')];assert len(summaries)==1;s=summaries[0]
assert ex['exit']==0 and not any(ex[k] for k in ['timeout','forcedCleanup','errors','remainingAfterCleanup','sampledPidsStillPresent','runtimeDirectoriesStillPresent']); assert all(Path(v).name=='node-compile-cache' for v in ex['unclassifiedTempEntriesStillPresent'])
assert d['sourceCommit']==pos['sourceCommit']==d['baseline']['served']['sourceCommit']==s['sourceCommit'];assert d['entry']==positive['entry'];assert not d['result']['claimEligible'] and not s['claimEligible']; assert len(d['sessions'])==20
assert all(t['payloadIdentity']==d['baseline']['served']['installedPayloadDigest']==positive['baseline']['served']['installedPayloadDigest'] and t['runtimeConfigurationIdentity']==d['baseline']['runtimeConfigurationIdentity'] for t in d['sessions'])
assert d['result']['declarations']['exactConfigurationDigest']==s['declarations']['exactConfigurationDigest']=='a9a17908b6fa87674d5f573b584f4a1ce1d41d7909fadf48b65e9b05397ddfd4'
rows=[r for r in d['result']['scenarios'] if r['id'] in pos['selectedRowsPassed']];assert len(rows)==13 and sum(r['state']=='pass' for r in rows)==12
row=next(r for r in rows if r['state']!='pass');assert row['state']=='fail'; assert sum(len(r.get('exchanges',[])) for r in d['result']['scenarios'])==s['exchanges'];assert s['childLaunches']==s['closedChildren']==21
for r in rows:
 if r is not row: assert all(a['passed'] for e in r['exchanges'] for a in e['assertions']) and all(a['passed'] for action in r.get('actions',[]) for a in action['assertions'])
final=row['exchanges'][-1];fail=[a['id'] for a in final['assertions'] if not a['passed']];assert fail==s['failures'];assert final['request']['exact']['writeState']=='complete'
if name=='cross-sequence':
 assert row['id']=='transactional.idempotency.cross-carrier' and fail==['presentation','created-record'] and final['response']['status']==200 and s['exchanges']==184
 assert len([a for a in final['assertions'] if a['id'].startswith('control-') and a['passed']])==3
elif name.startswith('noop-'):
 assert row['id']=='transactional.batch.no-op-entries' and final['response']['status']==200 and all(i in fail for i in s['intended'])
 assert next(a['passed'] for a in final['assertions'] if a['id']=='schema')
 if name.endswith('group-inventory-and-head'): assert sorted(fail)==['group-inventory','observed-head'] and s['exchanges']==186
 else: assert next(a['passed'] for a in final['assertions'] if a['id']=='disposition') and s['exchanges']==181
elif name.startswith('deletion-'):
 assert row['id']=='transactional.batch.deletion-safety'
 post=next(e for e in row['exchanges'] if e['request']['id']==s['actionId']);assert post['response']['status']==200
 assert [a['id'] for a in post['assertions'] if not a['passed']]==s['failedPost']
 assert all(next(a['passed'] for a in post['assertions'] if a['id']==i) for i in s['stableIds']+s['positiveIds'])
 if name.endswith('committed-prefix-state'):
  assert not s['failedPost'] and 'complete-b' in fail and 'ordered-identities' in fail and final['response']['status']==200 and s['exchanges']==177
 else:
  assert sorted(s['failedPost'])==sorted('serving-context disposition required-position no-effect no-results no-next problem'.split());assert s['exchanges']==(176 if name.endswith('incoming-order') else 181)
else: raise AssertionError(name)
assert all(a['passed'] for e in row['exchanges'][:-1] for a in e['assertions']);assert all(a['passed'] for action in row.get('actions',[]) for a in action['assertions'])
a={'accepted':True,'sourceCommit':d['sourceCommit'],'attemptedHttp':s['exchanges'],'childLaunches':21,'closedChildren':21,'reportSha256':hashlib.sha256((R/filename).read_bytes()).hexdigest(),'failures':fail,'otherAuthoredRowsPassed':12,'claimEligible':False,'cleanup':'No owned processes/runtime roots; Node cache retained.','basis':'Root checked complete report against explicit control oracle, passed prefixes/client assertions, source/entry/payload/config identity and runtime cleanup. Log summaries agree. No runtime executed by adjudicator.'}
print(json.dumps({'name':name,'summary':s,'adjudication':a},indent=2))
if '--accept' in sys.argv:
 out=R/(name+'.adjudication.json');assert not out.exists();out.write_text(json.dumps(a,indent=2)+'\n')
