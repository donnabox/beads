"""Closed clean-candidate gates; data/Git reads only. Root owns all releases."""
from pathlib import Path
import hashlib,json,subprocess,os
P=Path(__file__).resolve().parent
W=Path('/Users/dbox/repos/wt/vickie/vickie-pr150-main-reconciliation')
E=P/'run'
MODE='final'
sha=lambda b:hashlib.sha256(b).hexdigest()
def require(ok,msg):
 if not ok:raise RuntimeError(msg)
def read(p):return json.loads(Path(p).read_text())
def write(n,v):
 with (E/n).open('x') as f:f.write(json.dumps(v,indent=2)+'\n')
def git(*args):return subprocess.check_output(['git',*args],cwd=W,timeout=15)
def document(b):
 require(isinstance(b,dict),'Unbound document')
 p=Path(b['path']);require(p.is_absolute() and sha(p.read_bytes())==b['sha256'],'Document changed '+str(p));return read(p)
def tools_check(t):
 for n,b in t.items():
  p=Path(n);require(str(p.resolve(strict=True))==b['realpath'] and sha(p.read_bytes())==b['sha256'],'Tool changed '+n)
def source_check(b):
 require(b['worktree']==str(W) and b['branch']=='codex/vickie-pr150-main-reconciliation','Wrong source')
 require(git('branch','--show-current').decode().strip()==b['branch'],'Wrong branch')
 require(git('rev-parse','HEAD').decode().strip()==b['head'] and git('rev-parse','HEAD^{tree}').decode().strip()==b['tree'],'Wrong candidate')
 require(git('show','-s','--format=%P','HEAD').decode().split()==['d4f3f4be88c7837b71d6646518ca3589d464ba3d','8d6671ed19b0c0406428ccc2d2161768d01e8b9b'],'Wrong parent order')
 require(not git('status','--porcelain').strip(),'Candidate is not clean')
 actual={p:sha((W/p).read_bytes()) for p in git('ls-files').decode().splitlines()}
 require(actual==b['trackedSourceFiles'],'Tracked source changed')
 return sha(json.dumps(actual,sort_keys=True).encode())
def inventory():
 result={}
 for group in ['packages','apps']:
  for package in sorted((W/group).iterdir()):
   if not package.is_dir():continue
   dist=package/'dist';require(dist.is_dir(),'Missing dist '+str(dist))
   for p in sorted(dist.rglob('*')):
    require(not p.is_symlink(),'Linked build output')
    if p.is_file():result[p.relative_to(W).as_posix()]=sha(p.read_bytes())
 for p in [W/'LICENSE',W/'packages/server/package.json',W/'packages/protocol/package.json',*(W/'packages/protocol/schemas').rglob('*')]:
  require(not p.is_symlink(),'Linked artifact')
  if p.is_file():result[p.relative_to(W).as_posix()]=sha(p.read_bytes())
 require(result,'Empty build inventory');return result
def consumer_inventory():
 C=E/'consumer';files={};links={}
 for directory,dirs,names in os.walk(C,followlinks=False):
  for n in dirs+names:
   p=Path(directory)/n;r=p.relative_to(C).as_posix()
   if p.is_symlink():
    q=p.resolve(strict=True);require(q.is_relative_to(C.resolve()),'Consumer link escapes');links[r]={'link':os.readlink(p),'target':q.relative_to(C.resolve()).as_posix()}
   elif p.is_file():files[r]=sha(p.read_bytes())
 return {'files':files,'links':links}

def complete(n,g):
 r=read(E/(n+'.exit.json'));require(r['accepted'] and r['sourceUnchanged'] and r['sourceIdentity']==g['_sourceIdentity'] and not r['remainingProcesses'] and not r['forcedCleanup'],'Prior stage failed '+n)
 require(sha((E/(n+'.log')).read_bytes())==r['logSha256'],'Prior log changed '+n)
 return r
def verify_build(g):
 b=read(E/'build-artifacts.json');complete('build',g)
 require(b['head']==g['finalHead'] and b['tree']==g['finalTree'] and b['sourceIdentity']==g['_sourceIdentity'] and b['artifacts']==inventory(),'Build binding changed')
 for x in b['records'].values():require(sha(Path(x['path']).read_bytes())==x['sha256'],'Build record changed')
def verify_consumer(g):
 proof=document(g['installedProvenance']);expected=document(g['consumerInventory'])
 require(expected==consumer_inventory(),'Consumer changed')
 require(proof['sourceCommit']==g['finalHead'] and proof['sourceTree']==g['finalTree'] and proof['sourceIdentity']==g['_sourceIdentity'],'Consumer source changed')
 require(sha((E/'consumer/pnpm-lock.yaml').read_bytes())==g['consumerLockSha256']==proof['consumerLockSha256'],'Consumer lock changed')
 require(sha((E/'tar-source-provenance.json').read_bytes())==proof['tarProvenanceSha256'],'Tar provenance changed')
 for x in read(E/'tar-source-provenance.json')['packages'].values():require(sha(Path(x['archive']).read_bytes())==x['archiveSha256'],'Archive changed')
 for x in proof['entries'].values():require(sha(Path(x['realpath']).read_bytes())==x['sha256'],'Entry changed')
 return proof
def load_gate(name,child=False):
 require(__debug__,'Assertions required')
 subprocess.run(['/opt/homebrew/opt/python@3.14/bin/python3.14','/Users/dbox/repos/wt/vickie/bdp-integration-logs/guarded-action.py','/usr/bin/true'],check=True,timeout=30)
 g=read(P/'execution-gate.json');require(g['rootExecutionReleased'] is True and name in g['approvedStages'],'Root gate closed')
 require(sha((P/'preparation-index.json').read_bytes())==g['preparationIndexSha256'],'Preparation not accepted')
 for n,h in read(P/'preparation-index.json')['files'].items():require(sha((P/n).read_bytes())==h,'Preparation changed '+n)
 b=document(g['sourceBinding']);g['_sourceBinding']=b;g['_sourceIdentity']=source_check(b)
 require(g['finalHead']==b['head'] and g['finalTree']==b['tree'] and sha((W/'pnpm-lock.yaml').read_bytes())==g['sourceLockSha256'],'Candidate metadata changed')
 t=read(P/'tools.json');tools_check(t['tools'])
 for n,h in {**t['records'],**t['pythonEnvironmentFiles']}.items():require(sha(Path(n).read_bytes())==h,'Pinned input changed '+n)
 cfg=read(P/'commands.json');order=cfg['finalOrder'];require(name in order,'Unknown stage')
 for n in order[:order.index(name)]:complete(n,g)
 if name!='dependencies':tools_check(document(g['workspaceTools'])['tools'])
 if order.index(name)>order.index('bind-build'):verify_build(g)
 if name=='e2e-ready-preflight':require('full-test' in g['rootResultsAccepted'],'Root full-suite result audit pending')
 if name.startswith('packed-affected-'):
  verify_consumer(g)
  require('full-test' in g['rootResultsAccepted'],'Root full-suite result audit pending')
  for n in order[:order.index(name)]:
   if n.startswith('packed-affected-'):require(n in g['rootResultsAccepted'],'Prior packed result audit pending')
 if name=='evidence-verify':
  baseline=read(P/'parent-evidence-baseline.json')
  require(g['expectedParentEvidenceExit']==baseline['rootExpectedExit']==0,'Root expected evidence baseline not accepted')
  for key in ['ciRecord','ciLog','earlierLocalRawExit']:
   item=baseline[key];require(sha(Path(item['path']).read_bytes())==item['sha256'],'Evidence baseline changed')
 if name=='verify-installed-consumer':require(g['consumerLockSha256'] is not None,'Root consumer lock acceptance pending')
 import copy
 spec=copy.deepcopy(cfg['commands'][name])
 spec['environment']={k:v.replace('<ACTUAL_RECONCILED_COMMIT>',g['finalHead']) for k,v in spec['environment'].items()}
 require('<' not in json.dumps(spec),'Unbound command placeholder')
 return g,spec
