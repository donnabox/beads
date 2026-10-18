from pathlib import Path
import subprocess,time,json,hashlib,os,signal,datetime,runpy
p=Path(__file__).parent
w=Path("/Users/dbox/repos/wt/vickie/vickie-owned-set-retarget-minimum")
b=json.loads((p/"binding.json").read_text());sha=lambda x:hashlib.sha256(x).hexdigest()
def guard(): subprocess.run(["/opt/homebrew/opt/python@3.14/bin/python3.14","/Users/dbox/repos/wt/vickie/bdp-integration-logs/write-profile-D4-owned-set-code-review-1/claude-code-review/guarded-action-session-20261018-021624-549f.py","/usr/bin/true"],check=True)
assert b["rootLaunchReleased"] is True, "Closed: root must release this one review"
assert b["sourceBindingFinalized"] is True, "Closed: actual author source has not been frozen"
outputs=["source-before.json","source-after.json","process.json","stream.jsonl","progress.jsonl","stderr.txt","exit.json","final-review.txt","terminal-results.json"]
assert not any((p/name).exists() for name in outputs), "Preserve prior attempt; no retry"
def git(*args):return subprocess.check_output(["git",*args],cwd=w)
def merge_snapshot():
    head=git("rev-parse","HEAD").decode().strip()
    assert head=="cce884ccea1db97cbc6fd7bff876d4d6091f9081", "Candidate changed"
    assert git("status","--porcelain=v1","--untracked-files=all").decode()==b["expectedGitStatus"], "Frozen dirty source changed"
    assert set(git("diff","--name-only","HEAD").decode().splitlines())==set(b["ownedFiles"]), "Owned dirty scope changed"
    assert sha(git("diff","HEAD"))==b["patchSha256"], "Frozen full patch changed"
    assert set(git("ls-files").decode().splitlines())==set(b["files"]), "Tracked scope changed"
    return {"head":head,"tree":git("rev-parse","HEAD^{tree}").decode().strip(),"parents":git("show","-s","--format=%P","HEAD").decode().split()}
def snap():return {"files":{f:sha((w/f).read_bytes()) for f in b["files"]},"context":{f:sha(Path(f).read_bytes()) for f in b["context"]},"artifacts":{f:sha(Path(f).read_bytes()) for f in b["artifacts"]},"head":git("rev-parse","HEAD").decode().strip(),"mergeState":merge_snapshot(),"promptSha256":sha((p/"prompt.txt").read_bytes()),"runnerSha256":sha((p/"run.py").read_bytes()),"status":git("status","--porcelain=v1","--untracked-files=all").decode(),"bindingSha256":sha((p/"binding.json").read_bytes())}
before=snap()
for key in ["files","context","artifacts","head","mergeState","promptSha256","runnerSha256"]: assert before[key]==b[key], key
assert before["status"]==b["expectedGitStatus"], "Exact frozen one-file dirty status required"
parser=runpy.run_path(str(p/"parse-stream.py"))
args=json.loads((p/"arguments.json").read_text())
guard();os.umask(0o077);(p/"source-before.json").write_text(json.dumps(before,indent=2))
start=time.monotonic();timed=False;sampled=set();results=[];parse_errors=0;records=0;bytes_seen=0;pending=b"";last_event=None;wrapper_error=None;cleanup_required=False;c=None

def processes():
    rows=subprocess.check_output(["ps","-axo","pid=,ppid=,pgid=,%cpu=,time=,comm="],text=True).splitlines()
    return [r.split(None,5) for r in rows if len(r.split(None,5))==6]

def stop_group():
    try: os.killpg(c.pid,signal.SIGTERM)
    except ProcessLookupError: pass
    try:c.wait(timeout=5)
    except subprocess.TimeoutExpired:
        try:os.killpg(c.pid,signal.SIGKILL)
        except ProcessLookupError:pass
        c.wait(timeout=2)

with (p/"stream.jsonl").open("xb") as out,(p/"stderr.txt").open("xb") as err,(p/"prompt.txt").open("rb") as inp,(p/"progress.jsonl").open("x") as progress:
    def log(meta):
        progress.write(json.dumps({"elapsedSeconds":time.monotonic()-start,**meta})+"\n");progress.flush()
    def drain(tail, final=False):
        global pending,bytes_seen,parse_errors,records,last_event
        data=tail.read();bytes_seen+=len(data);pending+=data
        while b"\n" in pending:
            line,pending=pending.split(b"\n",1)
            if not line.strip():continue
            try:
                item=json.loads(line)
                if not isinstance(item,dict):raise ValueError("not-object")
            except (ValueError,UnicodeError):
                parse_errors+=1;log({"kind":"parse_error","lineBytes":len(line)});continue
            records+=1;last_event=time.monotonic()-start
            if item.get("type")=="result":results.append(item)
            log({"kind":"record","recordOrdinal":records,"recordBytes":len(line),**parser["record_metadata"](item)})
        if final and pending.strip():
            try:
                item=json.loads(pending)
                if not isinstance(item,dict):raise ValueError("not-object")
                records+=1;last_event=time.monotonic()-start
                if item.get("type")=="result":results.append(item)
                log({"kind":"record","recordOrdinal":records,"recordBytes":len(pending),**parser["record_metadata"](item)})
            except (ValueError,UnicodeError):
                parse_errors+=1;log({"kind":"partial_or_invalid_final_record","recordBytes":len(pending)})
            pending=b""
    try:
        guard()  # Recheck current remote role session immediately before reviewer execution.
        c=subprocess.Popen(args,cwd=w,stdin=inp,stdout=out,stderr=err,start_new_session=True)
        (p/"process.json").write_text(json.dumps({"pid":c.pid,"pgid":c.pid,"started":datetime.datetime.now(datetime.timezone.utc).isoformat(),"argv":args,"timeoutSeconds":1800,"termAfterSeconds":1790},indent=2))
        with (p/"stream.jsonl").open("rb") as tail:
            next_sample=0
            while c.poll() is None:
                drain(tail)
                elapsed=time.monotonic()-start
                if elapsed>=next_sample:
                    group=[r for r in processes() if r[2]==str(c.pid)];sampled.update(int(r[0]) for r in group)
                    log({"kind":"process_sample","streamBytes":bytes_seen,"records":records,"lastEventElapsed":last_event,"processes":[{"pid":int(r[0]),"ppid":int(r[1]),"pgid":int(r[2]),"cpuPercent":r[3],"cpuTime":r[4],"comm":r[5]} for r in group]});next_sample=elapsed+30
                if elapsed>=1790:timed=True;stop_group();break
                time.sleep(.25)
            drain(tail,final=True)
    except Exception as exc:
        wrapper_error=type(exc).__name__;log({"kind":"wrapper_error","errorType":wrapper_error})
        if c is not None and c.poll() is None:stop_group()
    finally:
        group=[r for r in processes() if c is not None and r[2]==str(c.pid)]
        if group:
            cleanup_required=True
            try:os.killpg(c.pid,signal.SIGTERM)
            except ProcessLookupError:pass
            time.sleep(1)
            try:os.killpg(c.pid,signal.SIGKILL)
            except ProcessLookupError:pass
remaining=[r for r in processes() if (c is not None and r[2]==str(c.pid)) or int(r[0]) in sampled]
after=snap()
source_unchanged=before["mergeState"]==after["mergeState"] and before["files"]==after["files"] and before["context"]==after["context"] and before["head"]==after["head"] and before["status"]==after["status"]==b["expectedGitStatus"]
packet_unchanged=before["artifacts"]==after["artifacts"] and before["promptSha256"]==after["promptSha256"] and before["runnerSha256"]==after["runnerSha256"] and before["bindingSha256"]==after["bindingSha256"]
if time.monotonic()-start>=1800:timed=True
classification=parser["classify"](results,exit_code=c.returncode if c else None,timed_out=timed,parse_errors=parse_errors,source_unchanged=source_unchanged,packet_unchanged=packet_unchanged,remaining=remaining,cleanup_required=cleanup_required,wrapper_error=wrapper_error)
result={"classification":classification,"reviewComplete":classification=="review_complete","approvalInferred":False,"exitCode":c.returncode if c else None,"timeout":timed,"elapsedSeconds":time.monotonic()-start,"sourceInputsUnchanged":source_unchanged,"packetUnchanged":packet_unchanged,"sampledPids":sorted(sampled),"remainingProcesses":remaining,"cleanupRequired":cleanup_required,"wrapperErrorType":wrapper_error,"streamRecords":records,"streamBytes":bytes_seen,"parseErrors":parse_errors,"terminalResultCount":len(results),"lastEventElapsed":last_event,"streamSha256":sha((p/"stream.jsonl").read_bytes()),"stderrSha256":sha((p/"stderr.txt").read_bytes())}
guard();(p/"source-after.json").write_text(json.dumps(after,indent=2));(p/"terminal-results.json").write_text(json.dumps(results,indent=2));(p/"exit.json").write_text(json.dumps(result,indent=2))
if classification=="review_complete":(p/"final-review.txt").write_text(results[0]["result"])
print(json.dumps(result))
