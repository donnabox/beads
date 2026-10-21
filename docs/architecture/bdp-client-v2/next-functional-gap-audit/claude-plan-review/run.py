"""Root-released single fresh collection minimum plan review. Metadata-only progress; partials never authorize work."""
from pathlib import Path
import subprocess, time, json, hashlib, os, signal, datetime, runpy
p = Path(__file__).resolve().parent
w = Path("/Users/dbox/repos/wt/vickie/vickie-fresh-depth-serving")
sha = lambda data: hashlib.sha256(data).hexdigest()
b = json.loads((p / "binding.json").read_text())
def guard():
    subprocess.run(["python3", "/Users/dbox/repos/wt/vickie/bdp-integration-logs/guarded-action.py", "/usr/bin/true"], check=True, timeout=30)
def git(*args):
    return subprocess.check_output(["git", *args], cwd=w, text=True, timeout=10).strip()
def snap():
    return {"files": {f: sha((w / f).read_bytes()) for f in b["files"]},
            "context": {f: sha(Path(f).read_bytes()) for f in b["context"]},
            "artifacts": {f: sha(Path(f).read_bytes()) for f in b["artifacts"]},
            "head": git("rev-parse", "HEAD"), "patchSha256": sha(subprocess.check_output(["git", "diff", "HEAD"], cwd=w, timeout=10)),
            "promptSha256": sha((p / "prompt.txt").read_bytes()), "runnerSha256": sha((p / "run.py").read_bytes()),
            "bindingSha256": sha((p / "binding.json").read_bytes())}
assert b["rootLaunchReleased"] is True, "Closed: root must release this new review attempt"
outputs = ["attempt.lock", "source-before.json", "source-after.json", "process.json", "stdout.jsonl", "stderr.txt", "progress.jsonl", "stream-summary.json", "final-result.txt", "exit.json"]
assert all(not (p / name).exists() for name in outputs), "Preserve attempt; no retry"
before = snap()
for key in ["files", "context", "artifacts", "head", "patchSha256", "promptSha256", "runnerSha256"]:
    assert before[key] == b[key], "Changed input: " + key
assert set(git("diff", "--name-only", "HEAD").splitlines()) == set(b["files"]), "Only bound clean plan-review checkout allowed"
assert not git("ls-files", "--others", "--exclude-standard"), "Unexpected untracked repo files"
# The parser is pinned above before executing this packet-only helper.
parse_stream = runpy.run_path(str(p / "parse-result.py"))["parse_stream"]
guard()
with (p / "attempt.lock").open("x") as lock:
    lock.write(datetime.datetime.now(datetime.timezone.utc).isoformat())
(p / "source-before.json").write_text(json.dumps(before, indent=2))
args = ["/opt/homebrew/bin/claude", "--safe-mode", "-p", "--output-format", "stream-json", "--verbose", "--include-partial-messages", "--tools", "Read,Glob,Grep", "--allowedTools", "Read,Glob,Grep", "--disable-slash-commands", "--strict-mcp-config", "--mcp-config", '{"mcpServers":{}}', "--no-session-persistence"]
start = time.monotonic()
sampled = set()
errors = []
timed = False
forced = False
c = None
next_note = 0
last_bytes = (0, 0)
last_activity = start
read_offset = 0
fragment = b""
event_count = 0
last_event = None
last_subtype = None

def rows():
    return [r.split(None, 3) for r in subprocess.check_output(["ps", "-axo", "pid=,ppid=,pgid=,comm="], text=True, timeout=5).splitlines() if len(r.split(None, 3)) == 4]

def terminate_group():
    global forced
    if c is None:
        return
    forced = True
    try:
        os.killpg(c.pid, signal.SIGTERM)
    except ProcessLookupError:
        return
    try:
        c.wait(timeout=4)
    except subprocess.TimeoutExpired:
        pass
    # Group children may outlive the leader, so KILL is not conditional on leader state.
    try:
        os.killpg(c.pid, signal.SIGKILL)
    except ProcessLookupError:
        pass
    try:
        c.wait(timeout=1)
    except subprocess.TimeoutExpired:
        errors.append("leader-survived-kill")

def interrupted(signum, frame):
    raise InterruptedError("review supervisor interrupted")

old_signals = {s: signal.signal(s, interrupted) for s in (signal.SIGINT, signal.SIGTERM)}
try:
    with (p / "stdout.jsonl").open("xb") as out, (p / "stderr.txt").open("xb") as err, (p / "prompt.txt").open("rb") as inp, (p / "progress.jsonl").open("x") as progress:
        c = subprocess.Popen(args, cwd=w, stdin=inp, stdout=out, stderr=err, start_new_session=True)
        (p / "process.json").write_text(json.dumps({"pid": c.pid, "pgid": c.pid, "started": datetime.datetime.now(datetime.timezone.utc).isoformat(), "argv": args, "timeoutSeconds": 3600, "termAfterSeconds": 3595, "headIsChronologyOnly": False}, indent=2))
        while True:
            now = time.monotonic()
            group = [r for r in rows() if r[2] == str(c.pid)]
            sampled.update(int(r[0]) for r in group)
            sizes = ((p / "stdout.jsonl").stat().st_size, (p / "stderr.txt").stat().st_size)
            if sizes != last_bytes:
                last_activity = now
                last_bytes = sizes
            if sizes[0] > 64 * 1024 * 1024 or sizes[1] > 4 * 1024 * 1024:
                errors.append("output-bound")
                terminate_group()
                break
            with (p / "stdout.jsonl").open("rb") as reader:
                reader.seek(read_offset)
                chunk = reader.read()
                read_offset += len(chunk)
            fragment += chunk
            lines = fragment.split(b"\n")
            fragment = lines.pop()
            for line in lines:
                if not line.strip():
                    continue
                try:
                    event = json.loads(line)
                    if isinstance(event, dict):
                        event_count += 1
                        last_event = str(event.get("type", "unknown"))[:64]
                        last_subtype = str(event.get("subtype", ""))[:64]
                except (ValueError, TypeError):
                    last_event = "malformed-record"
            if now - start >= next_note or c.poll() is not None:
                item = {"elapsedSeconds": round(now - start, 3), "stdoutBytes": sizes[0], "stderrBytes": sizes[1], "streamRecords": event_count, "lastEventType": last_event, "lastEventSubtype": last_subtype, "secondsSinceOutputActivity": round(now - last_activity, 3), "ownedGroup": [{"pid": int(r[0]), "ppid": int(r[1]), "pgid": int(r[2])} for r in group], "cliRunning": c.poll() is None, "reviewComplete": False}
                progress.write(json.dumps(item) + "\n")
                progress.flush()
                print(json.dumps(item), flush=True)
                next_note = now - start + 30
            if c.poll() is not None:
                break
            if now - start >= 3595:
                timed = True
                terminate_group()
                break
            time.sleep(2)
except BaseException as exc:
    errors.append(type(exc).__name__)
    terminate_group()
finally:
    for s in old_signals:
        signal.signal(s, signal.SIG_IGN)
    final_group = [r for r in rows() if c is not None and r[2] == str(c.pid)]
    if final_group:
        terminate_group()
    remaining = [r for r in rows() if int(r[0]) in sampled or (c is not None and r[2] == str(c.pid))]
    try:
        after = snap()
        inputs_ok = before == after
    except BaseException as exc:
        after = {"snapshotError": type(exc).__name__}
        inputs_ok = False
    raw = (p / "stdout.jsonl").read_bytes() if (p / "stdout.jsonl").exists() else b""
    parsed, final_text = parse_stream(raw)
    code = c.returncode if c is not None else None
    complete = code == 0 and not timed and not forced and not errors and not remaining and inputs_ok and parsed["terminalSuccess"]
    if timed:
        classification = "timeout"
    elif errors:
        classification = "supervisor-error"
    elif not inputs_ok:
        classification = "changed-inputs"
    elif remaining or forced:
        classification = "cleanup-failure-or-intervention"
    elif code != 0:
        classification = "nonzero-cli-exit"
    else:
        classification = parsed["classification"]
    result = {"exitCode": code, "timeout": timed, "forcedCleanup": forced, "elapsedSeconds": time.monotonic() - start, "classification": classification, "streamClassification": parsed["classification"], "reviewComplete": complete, "rootDispositionRequired": True, "approval": False, "inputsUnchanged": inputs_ok, "headBefore": before["head"], "headAfter": after.get("head"), "sampledPids": sorted(sampled), "groupAfterCLI": final_group, "remainingProcesses": remaining, "errors": errors, "stdoutSha256": sha(raw), "stderrSha256": sha((p / "stderr.txt").read_bytes()) if (p / "stderr.txt").exists() else None, "noSessionPersistence": True}
    guard()
    (p / "source-after.json").write_text(json.dumps(after, indent=2))
    (p / "stream-summary.json").write_text(json.dumps(parsed, indent=2))
    if final_text is not None:
        (p / "final-result.txt").write_text(final_text)
    (p / "exit.json").write_text(json.dumps(result, indent=2))
    print(json.dumps(result), flush=True)
    for s, handler in old_signals.items():
        signal.signal(s, handler)
raise SystemExit(0 if complete else 1)
