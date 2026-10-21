"""Read-only bounded CI capture; no rerun/cancel/merge. Root fills ci-target.json.
Run through guarded-action.py.300s snapshot interval,30minute monitor bound.
Refuses existing capture outputs; save any repeat attempt in a distinct directory."""
from pathlib import Path
import subprocess,json,time,datetime,hashlib
P=Path(__file__).resolve().parent
p=json.loads((P/'ci-target.json').read_text())
assert p['base']=='9a42a4531def99e12fa9a1fba02e70558d643265'
assert all(isinstance(p[k],str) and len(p[k])==40 for k in ['base','head','tree'])
assert all(type(p[k]) is int and p[k]>0 for k in ['pr','run','job'])
assert not any((P/n).exists() for n in ['ci-initial.json','ci-snapshots.jsonl','ci-terminal.json','ci-full.log'])
def raw(path):return subprocess.check_output(['gh','api','repos/gastownhall/bdp/'+path],timeout=60)
def save(name,b):
 with (P/name).open('xb') as f:f.write(b)
start=time.monotonic();first=True
while time.monotonic()-start<1800:
 run=raw(f"actions/runs/{p['run']}");job=raw(f"actions/jobs/{p['job']}");pr=raw(f"pulls/{p['pr']}")
 r,j,q=map(json.loads,[run,job,pr]);record={'capturedAt':datetime.datetime.now(datetime.timezone.utc).isoformat(),'run':r,'job':j,'pr':q}
 assert r['head_sha']==j['head_sha']==q['head']['sha']==p['head'] and q['base']['sha']==p['base'],'HEAD/BASE DRIFT; stop'
 assert j['run_id']==p['run']
 if first:
  save('ci-initial.json',run);save('ci-job-initial.json',job);save('pr-initial.json',pr);first=False
 with (P/'ci-snapshots.jsonl').open('a') as f:f.write(json.dumps(record)+'\n')
 if r['status']=='completed':
  save('ci-terminal.json',run);save('ci-job-terminal.json',job);save('pr-terminal.json',pr)
  log=raw(f"actions/jobs/{p['job']}/logs");save('ci-full.log',log)
  print(json.dumps({'conclusion':r['conclusion'],'head':r['head_sha'],'logBytes':len(log),'logSha256':hashlib.sha256(log).hexdigest()}),flush=True);break
 time.sleep(300)
else:raise SystemExit('Monitor bound reached without terminal result; no CI action taken')
