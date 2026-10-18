from pathlib import Path
import subprocess, time, json, hashlib, os, signal
p=Path(__file__).parent
w=Path("/Users/dbox/repos/wt/vickie/vickie-sequence-depth-retained")
names=["plan.md","correction.patch","claude-prompt.txt"]
def pins(): return {n:hashlib.sha256((p/n).read_bytes()).hexdigest() for n in names}
before=pins(); start=time.monotonic(); timed=False
subprocess.run(["python3","/Users/dbox/repos/wt/vickie/bdp-integration-logs/guarded-action.py","/usr/bin/true"],check=True)
with (p/"claude-stdout.txt").open("xb") as out,(p/"claude-stderr.txt").open("xb") as err,(p/"claude-prompt.txt").open("rb") as inp:
 c=subprocess.Popen(["/opt/homebrew/bin/claude","--safe-mode","-p","--tools","","--disable-slash-commands","--strict-mcp-config","--mcp-config",'{"mcpServers":{}}',"--no-session-persistence"],cwd=w,stdin=inp,stdout=out,stderr=err,start_new_session=True)
 (p/"claude-process.json").write_text(json.dumps({"pid":c.pid,"timeoutSeconds":600,"inputHashes":before},indent=2))
 try: c.wait(timeout=595)
 except subprocess.TimeoutExpired:
  timed=True; os.killpg(c.pid,signal.SIGTERM)
  try:c.wait(timeout=5)
  except subprocess.TimeoutExpired:os.killpg(c.pid,signal.SIGKILL);c.wait()
 result={"exit":c.returncode,"timeout":timed,"elapsedSeconds":time.monotonic()-start,"inputHashes":before,"inputsUnchanged":before==pins(),"stdoutSha256":hashlib.sha256((p/"claude-stdout.txt").read_bytes()).hexdigest()}
 (p/"claude-exit.json").write_text(json.dumps(result,indent=2));print(json.dumps(result))
