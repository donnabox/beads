HASHES={'base': '9bcb323adbca9743126a94223cfd0c71dbdf5fd5f52dcf545e259cdd7423c5f4', 'binding': '61afa5ea5adbdaddece495c3983bbe20723c09d687f89f86f4804d55fef3ad1f', 'commands': '9539cbe56d86b1de781818084458a29961266dc22cee18af99d125be6911b1e8'}
from pathlib import Path
import importlib.util,json,hashlib,subprocess,sys
P=Path(__file__).resolve().parent
W=Path('/Users/dbox/repos/wt/vickie/vickie-authorization-transition')
BASE=Path('/Users/dbox/repos/wt/vickie/bdp-integration-logs/transaction-authorization-transition-plan/source-validation-v4/run-source.py')
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
assert sha(BASE)==HASHES['base']
assert sha(P/'source-binding.json')==HASHES['binding'] and sha(P/'commands.json')==HASHES['commands']
b=json.loads((P/'source-binding.json').read_text());before=json.loads((P/'preservation-before.json').read_text());commands=json.loads((P/'commands.json').read_text())['commands'];order=[x['id']for x in commands]
def git(*args):return subprocess.check_output(['git',*args],cwd=W,timeout=10)
def source():
 assert git('rev-parse','HEAD').decode().strip()==b['head'] and git('rev-parse','HEAD^{tree}').decode().strip()==b['tree']
 assert not git('diff','--cached','--name-only').strip() and not git('ls-files','--others','--exclude-standard').strip()
 assert set(git('diff','HEAD','--name-only').decode().splitlines())==set(b['files'])
 actual={n:sha(W/n)for n in b['files']};assert actual=={n:x['sha256']for n,x in b['files'].items()}
 assert all(sha(W/n)==h for n,h in before['tracked'].items()if n not in b['files'])
 assert all(sha(W/n)==h for n,h in before['emitted'].items())
 return actual
spec=importlib.util.spec_from_file_location('bounded_supervisor',BASE);m=importlib.util.module_from_spec(spec);spec.loader.exec_module(m)
m.E=P/'run';m.COMMANDS=P/'commands.json';m.COMMANDS_SHA=HASHES['commands'];m.source_hashes=source;m.source_patch_hash=lambda:hashlib.sha256(git('diff','--binary','HEAD')).hexdigest()
def load(name):
 assert name in order and sha(P/'commands.json')==HASHES['commands']
 actual=source();assert m.source_patch_hash()==b['diffSha256']
 for prior in order[:order.index(name)]:
  r=json.loads((m.E/(prior+'.exit.json')).read_text());assert r['supervisorExit']==0 and r['sourceUnchanged']
 return {'sourceSha256':actual,'sourceDiffSha256':b['diffSha256'],'rootAcceptance':{'sha256':sha(P/'authorization-record.json')},'preparationSha256':sha(Path(__file__))},next(x for x in commands if x['id']==name)
def adjudicate(s):
 if 'fullFiles'not in s:return {'kind':'command-only'}
 p=P/'run/affected-files.results.json';r=json.loads(p.read_text());assert r['success']and r['numFailedTests']==0 and not r.get('numRuntimeErrorTestSuites',0)and not r.get('unhandledErrors')and not r.get('errors')
 assert set(x['name']for x in r['testResults'])=={str(W/n)for n in s['fullFiles']}
 cases=[x for f in r['testResults']for x in f['assertionResults']];assert all(x['status']=='passed'for x in cases)and len(cases)==r['numPassedTests']and len(cases)>3
 assert all(not f.get('message')and not f.get('failureMessage')for f in r['testResults'])
 return {'kind':'two-complete-files','passed':len(cases),'fullNames':[x['fullName']for x in cases],'jsonSha256':sha(p)}
m.load_gate=load;m.adjudicate_source=adjudicate
sys.exit(m.main())
