"""External, unexecuted preparation. A root-supplied gate is mandatory."""
from pathlib import Path
import hashlib,json,re,subprocess
P=Path(__file__).resolve().parent
W=Path('/Users/dbox/repos/wt/vickie/vickie-cross-carrier-draft')
E=P/'run'
sha=lambda b:hashlib.sha256(b).hexdigest()
PARENT='eff171839fe4f37638284a4c3346428979d26b5d'
OWNED={'packages/conformance/test-support/transactional/manifest-v1.json','packages/conformance/test-support/transactional/coverage-v1.json','packages/conformance/src/transactional-cumulative.test.ts','apps/bdptest/src/transactional-matrix.test.ts'}
def git(*args): return subprocess.check_output(['git',*args],cwd=W,text=True).strip()
def load_gate(frozen=False):
 g=json.loads((P/'execution-gate.json').read_text()) # Deliberately not provided by preparation.
 assert g['released'] is True and g['rootAcceptancePath']
 assert sha(Path(g['rootAcceptancePath']).read_bytes())==g['rootAcceptanceSha256']
 assert set(g['sourceSha256'])==OWNED
 assert set(git('diff','--name-only','HEAD').splitlines())<=OWNED
 for path,digest in g['sourceSha256'].items():
  assert (W/path).resolve().is_relative_to(W.resolve())
  assert sha((W/path).read_bytes())==digest,path
 assert sha((P/'commands.json').read_bytes())==g['commandsSha256']
 if frozen:
  assert re.fullmatch('[0-9a-f]{40}',g['finalHead'] or '')
  assert re.fullmatch('[0-9a-f]{40}',g['finalTree'] or '')
  assert g['finalHead']!='eff171839fe4f37638284a4c3346428979d26b5d'
  assert git('rev-parse','HEAD')==g['finalHead']
  assert git('rev-parse','HEAD^{tree}')==g['finalTree']
  assert git('status','--porcelain')==''
  subprocess.check_call(['git','merge-base','--is-ancestor',PARENT,g['finalHead']],cwd=W)
  assert set(git('diff','--name-only',PARENT,g['finalHead']).splitlines())==OWNED
  assert g['sourceChecksAccepted'] is True
 return g
