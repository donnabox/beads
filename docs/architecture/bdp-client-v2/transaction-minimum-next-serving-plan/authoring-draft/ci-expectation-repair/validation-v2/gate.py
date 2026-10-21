from pathlib import Path
import hashlib,json,subprocess
P=Path(__file__).resolve().parent;D=P.parent.parent;L=D.parent.parent
W=Path('/Users/dbox/repos/wt/vickie/vickie-minimum-current-resource');E=P/'run';GUARD=L/'guarded-action.py';COMMANDS=P/'commands.json'
COMMANDS_SHA='5a694b3220ffe6d2e0aec50568c622e699b8aa28ded56cb5e03301e313916142'
BINDING_SHA='c681bd235bc018f4b094178b29ce27c1b02341ee60f49be666395293d5668bfd'
RELEASE_SHA='11bec7d1f795884d9c8a35f6887eccd0bc17e76a592801af9c24f38ff497ccf6'
sha=lambda b:hashlib.sha256(b).hexdigest()
def require(ok,message):
 if not ok:raise RuntimeError(message)
def git(*a):return subprocess.check_output(['git',*a],cwd=W,timeout=10)
def input_files():return {p:sha((W/p).read_bytes()) for p in git('ls-files').decode().splitlines()}
def source_patch_hash():return sha(git('diff','--binary'))
def binding():
 require(sha((P/'execution-binding.json').read_bytes())==BINDING_SHA,'Binding changed')
 return json.loads((P/'execution-binding.json').read_text())
def source_hashes():
 b=binding();s=b['source'];require(git('rev-parse','HEAD').decode().strip()==s['head'],'Head drift');require(source_patch_hash()==s['diffSha256'],'Patch drift');require(input_files()==b['inputFiles'],'Tracked input drift')
 return {p:sha((W/p).read_bytes()) for p in s['files']}
def load_gate(name):
 require(__debug__,'Assertions required');b=binding();require(sha((P/'delegated-release.json').read_bytes())==RELEASE_SHA,'Release changed');r=json.loads((P/'delegated-release.json').read_text());require(r['released'],'Closed');require(sha(COMMANDS.read_bytes())==COMMANDS_SHA,'Commands changed');cfg=json.loads(COMMANDS.read_text());require(name in r['approvedCommandIds']==cfg['executionOrder'],'Stage unauthorized')
 require(sha((P/'supervise.py').read_bytes())==r['supervisorSha256'],'Supervisor changed')
 for path,h in {**b['toolPins'],**b['evidencePins'],**b['archives']}.items():require(sha(Path(path).read_bytes())==h,'Artifact drift '+path)
 for path,h in b['outputFiles'].items():require(sha((W/path).read_bytes())==h,'Output drift '+path)
 c=Path(b['consumer']);files={p.relative_to(c).as_posix():sha(p.read_bytes()) for p in c.rglob('*') if p.is_file() and not p.is_symlink()};links={p.relative_to(c).as_posix():p.resolve(strict=True).relative_to(c).as_posix() for p in c.rglob('*') if p.is_symlink()};require(files==b['consumerFiles'] and links==b['consumerLinks'],'Consumer drift')
 hashes=source_hashes()
 for prior in cfg['executionOrder'][:cfg['executionOrder'].index(name)]:
  x=json.loads((E/(prior+'.exit.json')).read_text());require(x['exit']==x['supervisorExit']==0 and x['sourceUnchanged'],'Prior failure')
  for k in ['timeout','forcedCleanup','errors','remainingAfterCleanup','sampledPidsStillPresent','runtimeDirectoriesStillPresent']:require(not x[k],'Prior unclean')
 return {'sourceSha256':hashes,'sourceDiffSha256':source_patch_hash(),'rootAcceptance':{'sha256':RELEASE_SHA},'preparationSha256':r['supervisorSha256']},cfg['commands'][name]
