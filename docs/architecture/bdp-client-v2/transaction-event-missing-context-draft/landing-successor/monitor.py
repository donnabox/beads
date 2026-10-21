from pathlib import Path
import subprocess,json,time,datetime,hashlib
E=Path('/Users/dbox/repos/wt/vickie/bdp-integration-logs/transaction-event-missing-context-draft/landing-successor')
head="db5a4d6a41814e59704ebef60fda02a2d57e9ad2";base="88c9408e908590ab1a4c64c9b9dee1a8b705999c"
start=time.monotonic()
def api(path):return subprocess.check_output(["gh","api",path],timeout=60)
def save(name,b):
 with (E/name).open("xb") as f:f.write(b)
# Initial snapshot was captured separately; every subsequent snapshot starts at least90s later.
time.sleep(90)
while time.monotonic()-start<1500:
 run=api("repos/gastownhall/bdp/actions/runs/38731731790");jobs=api("repos/gastownhall/bdp/actions/runs/38731731790/jobs");pr=api("repos/gastownhall/bdp/pulls/127")
 r=json.loads(run);p=json.loads(pr);j=json.loads(jobs)
 stamp=datetime.datetime.now(datetime.timezone.utc).isoformat()
 with (E/"ci-snapshots.jsonl").open("a") as f:f.write(json.dumps({"capturedAt":stamp,"run":r,"jobs":j,"pr":p})+"\n")
 if r["head_sha"]!=head or p["head"]["sha"]!=head or p["base"]["sha"]!=base:
  save("ci-drift.json",json.dumps({"at":stamp,"runHead":r["head_sha"],"prHead":p["head"]["sha"],"prBase":p["base"]["sha"]},indent=2).encode());raise SystemExit("HEAD/BASE DRIFT; stopped")
 if r["status"]=="completed":
  save("ci-terminal.json",run);save("ci-jobs-terminal.json",jobs);save("pr-terminal.json",pr)
  log=api("repos/gastownhall/bdp/actions/jobs/116410895014/logs");save("ci-full.log",log)
  print(json.dumps({"status":r["status"],"conclusion":r["conclusion"],"logBytes":len(log),"logSha256":hashlib.sha256(log).hexdigest()}),flush=True);break
 time.sleep(90)
else:raise SystemExit("Monitoring bound reached; run not terminal, no action taken")
