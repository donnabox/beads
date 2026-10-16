HASHES={'base': '9bcb323adbca9743126a94223cfd0c71dbdf5fd5f52dcf545e259cdd7423c5f4', 'binding': '1d163c7b17364bc5ac94ccb853642ba86993ac5bd3639db3066b8106afc9212c', 'commands': 'd5b8db2b8f6730d04e25d87d5694d5722e46e4dc213212cc1327712c8b382f12'}
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
 gate=json.loads((P/'execution-gate.json').read_text())
 assert gate.get('status')=='released' and gate['session']=='2026-10-08T20:26:35Z-81e5'
 assert gate['supervisorSha256']==sha(Path(__file__)) and gate['sourceBindingSha256']==HASHES['binding'] and gate['commandsSha256']==HASHES['commands']
 assert gate['authorizedStages']==order[:len(gate['authorizedStages'])] and name in gate['authorizedStages']
 acceptance=gate['rootAcceptance'];assert sha(Path(acceptance['path']))==acceptance['sha256']
 actual=source();assert m.source_patch_hash()==b['diffSha256']
 for prior in order[:order.index(name)]:
  r=json.loads((m.E/(prior+'.exit.json')).read_text());assert r['supervisorExit']==0 and r['sourceUnchanged']
 return {'sourceSha256':actual,'sourceDiffSha256':b['diffSha256'],'rootAcceptance':acceptance,'preparationSha256':sha(Path(__file__))},next(x for x in commands if x['id']==name)
def adjudicate(s):
 if s['id']!='smoke':return {'kind':'command-only'}
 text=(P/'run/smoke.log').read_text()
 for executable in ['bdp','bdptest','bdpbd']:
  assert executable+': installed --help, --version, and diagnostics passed' in text
  assert executable+': installed clean lifecycle passed' in text
 for executable in ['bdptest','bdpbd']:
  assert executable+': public Read admission serves discovery and stops cleanly' in text
  assert executable+': public profile admission fails closed before bind' in text
 assert 'bdp: role isolation lifecycle ignores invalid bdptest/server values' in text
 events=[]
 for line in text.splitlines():
  if line.startswith('{'):
   v=json.loads(line)
   if v.get('event')=='installed.reference.observed':events.append(v)
 assert len(events)==1 and events[0]['claimEligible'] is False and events[0]['allNineTargets'] is True
 return {'kind':'complete-existing-smoke','installedObservation':events[0],'logSha256':sha(P/'run/smoke.log'),'noConformanceCredit':True}

m.load_gate=load;m.adjudicate_source=adjudicate
sys.exit(m.main())
