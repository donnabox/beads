"""Read-only single CI snapshot; terminal-only complete raw job log download.
Root supplies released ci-target.json and invokes through guarded-action.py.
No polling loop, retry, cancel, merge or remote write. Repeated snapshots >=90s."""
from pathlib import Path
import subprocess,json,time,datetime,hashlib,re
P=Path(__file__).resolve().parent
p=json.loads((P/'ci-target.json').read_text())
assert __debug__ and p['released'] is True
assert p['base']=='df80f8cb86b7763bc7bb005492e43d8840527045'
assert all(isinstance(p[k],str) and re.fullmatch('[0-9a-f]{40}',p[k]) for k in ['base','head','tree'])
assert all(type(p[k]) is int and p[k]>0 for k in ['pr','run','job'])
assert not (P/'ci-terminal.json').exists(),"Terminal attempt preserved; do not repeat"
history=P/'ci-snapshots.jsonl'
if history.exists():
 last=json.loads(history.read_text().splitlines()[-1])
 assert time.time()-last['capturedUnix']>=90,"Snapshot too soon"
def raw(path):return subprocess.check_output(['gh','api','repos/gastownhall/bdp/'+path],timeout=60)
def save(name,b):
 with (P/name).open('xb') as f:f.write(b)
run=raw(f"actions/runs/{p['run']}");job=raw(f"actions/jobs/{p['job']}");pr=raw(f"pulls/{p['pr']}")
r,j,q=map(json.loads,[run,job,pr])
assert r['head_sha']==j['head_sha']==q['head']['sha']==p['head'] and q['base']['sha']==p['base'],"Head/base drift"
assert j['run_id']==p['run'] and r['event']=='pull_request'
record={'capturedUnix':time.time(),'capturedAt':datetime.datetime.now(datetime.timezone.utc).isoformat(),'run':r,'job':j,'pr':q}
if not history.exists():
 save('ci-initial.json',run);save('ci-job-initial.json',job);save('pr-initial.json',pr)
 synthetic=q.get('merge_commit_sha')
 if isinstance(synthetic,str) and re.fullmatch('[0-9a-f]{40}',synthetic):
  merge=raw('git/commits/'+synthetic);m=json.loads(merge)
  save('synthetic-initial.json',merge)
  assert m['tree']['sha']==p['tree'] and [x['sha'] for x in m['parents']]==[p['base'],p['head']],"Synthetic tree/parents drift"

with history.open('a') as f:f.write(json.dumps(record)+'\n')
if r['status']=='completed' and j['status']=='completed':
 synthetic=q['merge_commit_sha'];merge=raw('git/commits/'+synthetic);m=json.loads(merge)
 assert m['tree']['sha']==p['tree'] and [x['sha'] for x in m['parents']]==[p['base'],p['head']],"Terminal synthetic tree/parents drift"
 save('synthetic-terminal.json',merge)
 log=raw(f"actions/jobs/{p['job']}/logs")
 save('ci-terminal.json',run);save('ci-job-terminal.json',job);save('pr-terminal.json',pr);save('ci-full.log',log)
 print(json.dumps({'conclusion':r['conclusion'],'jobConclusion':j['conclusion'],'head':r['head_sha'],'logBytes':len(log),'logSha256':hashlib.sha256(log).hexdigest()}))
else:print(json.dumps({'status':r['status'],'jobStatus':j['status'],'head':r['head_sha'],'nextSnapshotEarliestSeconds':90}))
