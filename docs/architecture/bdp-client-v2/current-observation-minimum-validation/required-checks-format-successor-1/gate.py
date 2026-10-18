from pathlib import Path
import json,hashlib,subprocess,os,importlib.util
P=Path(__file__).resolve().parent
V=Path("/Users/dbox/repos/wt/vickie/bdp-integration-logs/current-observation-minimum-validation/format-successor-1")
W=Path("/Users/dbox/repos/wt/vickie/vickie-current-observation-minimum")
E=P/'run';MODE='final'
sha=lambda b:hashlib.sha256(b).hexdigest()
def require(ok,message):
 if not ok:raise RuntimeError(message)
def read(p):return json.loads(Path(p).read_text())
def write(n,v):
 with (E/n).open('x') as f:f.write(json.dumps(v,indent=2)+'\n')
def document(r):
 p=Path(r['path']);require(p.is_file() and not p.is_symlink() and sha(p.read_bytes())==r['sha256'],'Changed document '+str(p));return read(p)
refs=read(P/'references.json')
for name,h in refs['files'].items():require(sha(Path(name).read_bytes())==h,'Frozen reference changed '+name)
_spec=importlib.util.spec_from_file_location('required_binding_primitives',str(V/'gate.py'))
_v=importlib.util.module_from_spec(_spec);_spec.loader.exec_module(_v)
source_check=_v.source_check

def verify_build(g):
 record=document(g['freshBuild'])
 _v.verify_build(g)
 require(record==read(V/'run/build-artifacts.json'),'Different fresh build proof')
 for name in ['fresh-build','bind-fresh-build']:
  r=read(V/'run'/(name+'.exit.json'))
  require(r['accepted'] and r['exit']==0 and not r['forcedCleanup'] and not r['errors'] and not r['remainingProcesses'] and not r['runtimeDirectoriesStillPresent'] and r['sourceUnchanged'] and r['sourceIdentity']==g['_sourceIdentity'],'Fresh build prerequisite failed')
  require(sha((V/'run'/(name+'.log')).read_bytes())==r['logSha256'],'Build log changed')

def load_gate(name):
 require(__debug__,'Assertions must remain enabled')
 g=read(P/'execution-gate.json');c=read(P/'commands.json')
 require(g['released'] is True and g['session']=='2026-10-08T20:26:35Z-81e5','Required checks closed')
 require(g['actualClaudeReviewDisposed'] is True and g['nativeCodeReviewDisposed'] is True,'CODE dispositions pending')
 require(g['approvedStages']==c['order'][:len(g['approvedStages'])] and name in g['approvedStages'],'Stage unreleased')
 require(c['commands'][name]['outerSeconds']==120 and c['commands'][name]['termAfterSeconds']==115 and c['commands'][name]['automaticRetries']==0 and c['commands'][name]['acceptedRawExitCodes']==[0],'Required command bounds changed')
 for f,h in g['preparedArtifacts'].items():require(sha((P/f).read_bytes())==h,'Prepared artifact changed '+f)
 subprocess.run(['/opt/homebrew/opt/python@3.14/bin/python3.14','/Users/dbox/repos/wt/vickie/bdp-integration-logs/guarded-action.py','/usr/bin/true'],check=True,timeout=30)
 b=document(g['sourceBinding']);g['_sourceBinding']=b;g['_sourceIdentity']=source_check(b);g['finalHead']=b['head'];g['finalTree']=b['tree']
 base=document(g['baseTools']);_v.tools_check(base['tools'])
 g['_sourceBinding']=b;g['sourceLockSha256']=b['trackedSourceFiles']['pnpm-lock.yaml']
 require(base['sourceLockSha256']==g['sourceLockSha256'],'Base lock differs')
 require(g.get('bootstrapReuseAccepted') is True and g.get('bootstrapReuseBinding')=={'path':str(V/'bootstrap-reuse-binding.json'),'sha256':sha((V/'bootstrap-reuse-binding.json').read_bytes())},'Root bootstrap reuse acceptance absent')
 _v.verify_bootstrap_reuse({**g,'sourceBinding':{'path':str(V/'source-binding.json'),'sha256':sha((V/'source-binding.json').read_bytes())}})
 verify_build(g)
 for prior in c['order'][:c['order'].index(name)]:
  r=read(E/(prior+'.exit.json'));require(r['accepted'] and r['exit']==0 and r['sourceIdentity']==g['_sourceIdentity'] and r['sourceUnchanged'] and not r['remainingProcesses'] and not r['runtimeDirectoriesStillPresent'] and not r['forcedCleanup'] and not r['errors'],'Prior required check failed')
  require(sha((E/(prior+'.log')).read_bytes())==r['logSha256'],'Prior required log changed')
 return g,c['commands'][name]

def validate_result(name):
 if name=='preflight':
  d=read(E/'preflight.results.json');require(d['success'] is True and d['numPassedTests']==7 and d['numTotalTests']==7 and d['numFailedTests']==0 and d['numPendingTests']==0,'Expected seven passed preflight tests')
  cases=[r for suite in d['testResults'] for r in suite['assertionResults']]
  expected=read(P/'expectations.json')['preflightFullNames']
  require(sorted(r['fullName'] for r in cases)==sorted(expected) and all(r['status']=='passed' for r in cases),'Preflight callback identity differs')
 elif name in ['package-smoke','historical-read-evidence']:
  rows=[]
  for line in (E/(name+'.log')).read_text().splitlines():
   try:r=json.loads(line)
   except (ValueError,TypeError):continue
   if isinstance(r,dict):rows.append(r)
  event='installed.reference.observed' if name=='package-smoke' else 'evidence.verified'
  found=[r for r in rows if r.get('event')==event];require(len(found)==1,'Missing/duplicate required event '+event)
  if name=='package-smoke':require(found[0]['allNineTargets'] is True and found[0]['claimEligible'] is False,'Smoke observation differs')
  else:
   historical=read(P/'historical-evidence-binding.json')
   require(all(found[0][key]==historical[key] for key in ['artifact','constant','runHead']),'Historical artifact identity changed')
   require(all(sha((W/f).read_bytes())==h for f,h in historical['verificationInputs'].items()),'Historical evidence inputs changed')
