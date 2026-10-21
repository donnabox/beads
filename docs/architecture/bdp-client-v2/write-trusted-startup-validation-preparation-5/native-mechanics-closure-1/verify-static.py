from pathlib import Path
import json,hashlib,re,ast
L=Path('/Users/dbox/repos/wt/vickie/bdp-integration-logs')
D=Path('/Users/dbox/repos/wt/vickie/bdp-client-design/checkout/docs/architecture/bdp-client-v2')
p=L/'write-trusted-startup-validation-preparation-5';old=L/'write-trusted-startup-validation-preparation-4';out=p/'native-mechanics-closure-1'
def sha(x):return hashlib.sha256(x).hexdigest()
def j(f):return json.loads(f.read_text())
def file(f,h):assert sha(f.read_bytes())==h,str(f)
idx=j(p/'preparation-index.json');file(p/'preparation-index.json','ad48922a2c18662c906d08945cc5565dc51f6b1830a53b0f07e674e0e51eba42')
checks={}
for f,m in idx['files'].items():
 b=(p/f).read_bytes();assert sha(b)==m['sha256'] and len(b)==m['bytes'],f
 assert b==(D/p.name/f).read_bytes(),f
assert (p/'preparation-index.json').read_bytes()==(D/p.name/'preparation-index.json').read_bytes()
checks['frozenFilesAndMirror']=len(idx['files'])
impl=j(p/'implementation-hashes.json');file(p/'implementation-hashes.json','f7abdd4090565e51ea608f3f4d2858ff8e916f421e38012d2870ea6c6da071ca')
for f,h in impl['files'].items():file(p/f,h)
checks['implementationPins']=len(impl['files'])
refs=j(p/'reference-pins.json')
for f,h in refs.items():file(Path(f),h)
checks['referencePins']=len(refs)
orig=j(p/'successor5-origins.json')['origins'];by={r.get('currentPath'):r for r in orig if 'currentPath'in r}
indirect=0
for r in orig:
 if 'currentPath'in r:
  file(p/r['currentPath'],r['currentSha256'])
  if 'sealed'in r.get('derivation',''):
   h=j(p/r['preservationReference'])['files'][r['currentPath']]['sha256'];assert h==r['originSha256'];indirect+=1
  else:file(p/r['preservationReference'],r['originSha256'])
 else:file(p/r['copy'],r['originSha256'])
checks['origins']={'all':len(orig),'indirectSealedHistoricalAggregates':indirect}
def apply(before,patch):
 lines=before.splitlines(keepends=True);ps=patch.splitlines(keepends=True);o=[];cursor=0;i=0
 while i<len(ps):
  m=re.match(r'^@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@',ps[i])
  if not m:i+=1;continue
  n=int(m[1])-1 if int(m[2] or '1') else int(m[1]);assert n>=cursor;o+=lines[cursor:n];cursor=n;i+=1
  while i<len(ps) and not ps[i].startswith('@@ '):
   q=ps[i]
   if q.startswith(('--- ','+++ ')):break
   if q.startswith('\\'):i+=1;continue
   if q[:1] in (' ','-'):assert lines[cursor]==q[1:],(cursor,q);cursor+=1
   if q[:1] in (' ','+'):o.append(q[1:])
   i+=1
 o+=lines[cursor:];return ''.join(o)
man=j(p/'successor5-delta-manifest.json')
for r in man['pairs']:
 before=(p/by[r['path']]['preservationReference']).read_text();after=(p/r['path']).read_text()
 assert sha(before.encode())==r['beforeSha256'] and sha(after.encode())==r['afterSha256']
 for typ in ('forward','inverse'):file(p/r[typ],r[typ+'Sha256'])
 assert apply(before,(p/r['forward']).read_text())==after,r['path']
 assert apply(after,(p/r['inverse']).read_text())==before,r['path']
checks['genuineForwardAndInversePairs']=len(man['pairs'])
helpers=['run.py','gate.py','audit.py','packages.py','guard.py']
# Guard filename is recorded rather than assumed.
helpers=[f for f in impl['files'] if '/'not in f and f.endswith('.py')]
assert len(helpers)==5,helpers
for f in helpers:assert (p/f).read_bytes()==(old/f).read_bytes();ast.parse((p/f).read_text())
checks['byteIdenticalPythonHelpers']=helpers
mjs=(old/'bound-consumer-loader.mjs').read_text().replace(str(old)+'/run',str(p)+'/run')
assert mjs==(p/'bound-consumer-loader.mjs').read_text()
checks['mjsDelta']='sole executionRoot literal4-to5'
oldsel=j(old/'selection.json');sel=j(p/'selection.json');test='packages/server/src/reference-deployment.test.ts'
oldhash=oldsel['files'][test];newhash=sel['files'][test];oldsel['files'][test]=newhash;assert oldsel==sel
assert len(sel['selectedFullNames'])==len(set(sel['selectedFullNames']))==sel['callbackCount']==44
checks['selection']={'callbacks':44,'onlyMetadataFile':test,'before':oldhash,'after':newhash}
oldcmd=(old/'commands.json').read_text().replace(str(old),str(p));assert oldcmd==(p/'commands.json').read_text()
commands=j(p/'commands.json')['commands'];assert len(commands)==25
checks['commands']='25 unchanged command specifications and deadlines after path substitution'
for f in ['installed-audit-contract.json','source-audit-contract.json']:
 t=(old/f).read_text().replace(str(old),str(p)).replace('7e856365e701e96ddf233e13d6883c6d74464dd755c0ff59f1b185c37c7b19db','0c53a6bb33d57133c8fbb43b8787d9251ca4276abcdb18b2a9f434279bfe364b').replace(oldhash,newhash)
 assert t==(p/f).read_text(),f
checks['auditContracts']='only executionRoot/sourceIdentity/fixture metadata substitutions'
binding=j(p/'source-binding.json');file(p/'source-binding.json','0572475970aee5a5466530ad68c4e84b5725947a0d7f3740da9321b6a23f75de')
smap=j(p/'source-map.json');assert smap==binding['trackedSourceFiles'];assert len(smap)==682
for f,h in smap.items():file(Path(binding['worktree'])/f,h)
assert sha(json.dumps(smap,sort_keys=True).encode())=='0c53a6bb33d57133c8fbb43b8787d9251ca4276abcdb18b2a9f434279bfe364b'
oldmap=j(old/'source-map.json');assert [f for f in smap if smap[f]!=oldmap[f]]==[test]
for f in sel['files']:file(p/'context/source'/f,smap[f])
checks['currentSource']={'files':682,'unchangedFrom4':681,'identity':'0c53a6bb33d57133c8fbb43b8787d9251ca4276abcdb18b2a9f434279bfe364b'}
gate=j(p/'execution-gate.json');assert (p/'execution-gate.json').read_bytes()==(p/'execution-gate.closed.json').read_bytes()
for k in ['released','rootPlanAccepted','rootExactSourceAccepted','actualClaudeReviewDisposed','nativeCodeReviewDisposed','staticValidationAuthorized']:assert gate[k] is False,k
assert gate['approvedStages']==[] and not (p/'run').exists()
checks['gates']='identical CLOSED; no approved stages; no run directory; wrappers unresolved'
native=L/'write-trusted-startup-focused-source-correction-1/native-precision-closure-1';file(native/'closure.json','6052ffc66a3e53be0eb2265a8ef8d52f74597c5137c9a3e1c0eb22cc277b34f5')
ni=j(native/'index.json')
for f,m in ni['entries'].items():
 h=m['sha256'] if isinstance(m,dict) else m;file(native/f,h);assert (native/f).read_bytes()==(D/native.relative_to(L)/f).read_bytes()
checks['additiveSourceNativeClosure']='6052ffc66a3e53be0eb2265a8ef8d52f74597c5137c9a3e1c0eb22cc277b34f5'
epath=L/'root-packaging-command-error-1/error.json';file(epath,'6e62aba61963198de52e44b03ec7925bb6efa23ca34428e1142e7ae6b97cb525');err=j(epath)
file(old/'packages.py',err['helperSourceSha256'])
assert len(err['actualArgv'])==2 and err['actualExit']==1
for f,h in err['runFiles'].items():
 file(old/'run'/f,h);assert (old/'run'/f).read_bytes()==(D/old.name/'run'/f).read_bytes()
assert len(err['runFiles'])==43 and not (old/'run/packages').exists() and not (old/'run/consumer').exists()
og=j(old/'execution-gate.json');assert not og['released']
checks['rootCommandError']={'recordSha256':sha(epath.read_bytes()),'oldRunFilesUnchanged':43,'noPackagingConsumerDirectories':True,'oldGateClosed':True,'historicalApprovedStageCount':len(og['approvedStages']),'notValidationAttempt':True,'stderrNotReconstructed':True}
(out/'verification.json').write_text(json.dumps(checks,indent=2)+'\n');print(json.dumps(checks,indent=2))
