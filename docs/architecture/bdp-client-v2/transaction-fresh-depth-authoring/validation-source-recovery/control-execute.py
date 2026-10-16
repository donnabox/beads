"""Only the copied Snapshot control: preserve direct Vitest status under session guard."""
from gate import *
import sys,datetime
name='strict-control';g,spec=load_gate(name);require(spec['directExpectedExit']==1,'Wrong direct control exit')
target=E/'strict-control.direct-exit.json';require(not target.exists(),'Direct control record already exists')
argv=spec['directArgv'];before=source_hashes();mirror=verify_mirror();code=None;pid=None;errors=[]
started=datetime.datetime.now(datetime.timezone.utc).isoformat()
try:
 # Inherit the supervisor group and streams; its115/120s limits cover this child.
 child=subprocess.Popen(argv,cwd=spec['cwd']);pid=child.pid;code=child.wait()
except Exception as error:
 errors.append(type(error).__name__)
try:
 after=source_hashes();require(after==before,'Source changed');verify_mirror()
except Exception as error:
 after=None;errors.append('postcheck:'+type(error).__name__)
record={'stage':name,'argv':argv,'cwd':spec['cwd'],'started':started,'pid':pid,'directReturnCode':code,'terminationSignal':-code if isinstance(code,int) and code<0 else None,'sourceBefore':before,'sourceAfter':after,'errors':errors,'commandsSha256':sha(COMMANDS.read_bytes()),'mirrorSha256':sha((E/'control-mirror.json').read_bytes()),'guardMeaning':'Session guard launches this adapter; direct child return code is recorded independently of guard status.'}
with target.open('x') as f:f.write(json.dumps(record,indent=2)+'\n')
sys.exit(0 if code==1 and not errors else 1)
