from pathlib import Path
import json,hashlib,difflib,re
L=Path('/Users/dbox/repos/wt/vickie/bdp-integration-logs');D=Path('/Users/dbox/repos/wt/vickie/bdp-client-design/checkout/docs/architecture/bdp-client-v2');P=L/'write-trusted-startup-lifecycle-source-authoring-3';M=D/P.name
sha=lambda b:hashlib.sha256(b).hexdigest();load=lambda f:json.loads(f.read_text());C={}
def ck(k,v):C[k]=v;assert v,k
idx=load(P/'index.json');ck('authorIndex',sha((P/'index.json').read_bytes())=='1cb3933ec44789b2cf5ee693f518517f2f4d7a4bf6fa2bee4ce7b0cf83460d78');physical={str(f.relative_to(P)) for f in P.rglob('*') if f.is_file()};ck('770physical',len(physical)==770);ck('completeIndex',set(idx['files'])|{'index.json'}==physical)
for rel,r in idx['files'].items():
 b=(P/rel).read_bytes();ck('index:'+rel,sha(b)==r['sha256'] and len(b)==r['bytes'])
ck('mirrorComplete',{str(f.relative_to(M)) for f in M.rglob('*') if f.is_file()}==physical)
for rel in physical:ck('mirror:'+rel,(P/rel).read_bytes()==(M/rel).read_bytes())
orig=load(P/'input-origins.json');ck('736origins',len(orig)==736)
for rel,r in orig.items():
 b=(P/rel).read_bytes();ck('origin:'+rel,len(b)==r['bytes'] and sha(b)==r['sha256'] and Path(r['origin']).read_bytes()==b)
sb=load(P/'source-binding.json');R=Path(sb['worktree']);B=Path(sb['reviewedSourceCheckout']);before=load(P/'source-before-map.json');after=load(P/'source-after-map.json');base=load(P/'source-original-fdd-map.json');protected=load(P/'protected-other679-map.json')
ck('686maps',len(before)==len(after)==686 and before.keys()==after.keys());changed={f for f in before if before[f]!=after[f]};ck('only7changed',changed==set(sb['changedPaths']) and len(changed)==7);ck('protected679',len(protected)==679 and protected=={f:h for f,h in before.items() if f not in changed})
for rel,h in before.items():ck('original686:'+rel,sha((B/rel).read_bytes())==h)
for rel,h in after.items():ck('current686:'+rel,sha((R/rel).read_bytes())==h)
for rel,h in protected.items():ck('protected:'+rel,after[rel]==h)
ck('677baseprotected',sum(1 for f,h in base.items() if f not in sb['originalNineFootprint'] and after[f]==h)==677)
for f in sb['originalNineFootprint']:
 ck('afterCopy:'+f,(P/'after'/f).read_bytes()==(R/f).read_bytes());ck('beforeCopy:'+f,sha((P/'before'/f).read_bytes())==before[f]) if f in changed else ck('protectedBefore:'+f,before[f]==after[f]);
 if f in base:ck('fddCopy:'+f,sha((P/'baseline-fdd'/f).read_bytes())==base[f])
for prefix,folder1,folder2,files in [('correction3-final','before','after',sb['changedPaths']),('correction3-semantic','before','preformat',sb['changedPaths']),('format-only','preformat','after',sb['changedPaths']),('cumulative-fdd','baseline-fdd','after',sb['originalNineFootprint'])]:
 fw=[];iv=[]
 for rel in sorted(files):
  a=P/folder1/rel;b=P/folder2/rel;aa=a.read_text().splitlines(True) if a.exists() else [];bb=b.read_text().splitlines(True) if b.exists() else []
  if aa==bb:continue
  fw.extend(difflib.unified_diff(aa,bb,fromfile='a/'+rel if a.exists() else '/dev/null',tofile='b/'+rel));iv.extend(difflib.unified_diff(bb,aa,fromfile='a/'+rel,tofile='b/'+rel if a.exists() else '/dev/null'))
 ck('forward:'+prefix,''.join(fw).encode()==(P/(prefix+'.patch')).read_bytes());ck('inverse:'+prefix,''.join(iv).encode()==(P/(prefix+'.inverse.patch')).read_bytes())
# Exact lifecycle hot-path structure and unchanged close cache bodies, without evaluating source.
ru=(R/'packages/server/src/development-runtime.ts').read_text();tx=(R/'packages/server/src/transaction-runtime.ts').read_text();oldru=(B/'packages/server/src/development-runtime.ts').read_text();oldtx=(B/'packages/server/src/transaction-runtime.ts').read_text()
for name,new,old in [('RU',ru,oldru),('TX',tx,oldtx)]:
 ck(name+'factsInsideTry',new.index('  try {\n    if (prepared)',new.index('const closeListener'))<new.index('      preparedFacts = Object.freeze(')<new.index('    published = true;')<new.index('  let completion:'))
 ck(name+'noAwaitAfterCapture',not re.search(r'\bawait\b',new[new.index('      preparedFacts = Object.freeze('):new.index('  let completion:')].split('  } catch (error)')[0]))
 a=new[new.index('  function close(): Promise<void>'):new.index('  if (!prepared) return Object.freeze(legacy);') if name=='TX' else new.index('  const legacy = { scope, close };')];b=old[old.index('  function close(): Promise<void>'):old.index('  if (!prepared) return Object.freeze(legacy);') if name=='TX' else old.index('  const legacy = { scope, close };')];ck(name+'closeBodyExact',a==b)
 ck(name+'readConfigurationCallCount',new.count('.readConfiguration()')==old.count('.readConfiguration()'))
ck('TXreadOptionsCallCount',tx.count('captured.readOptions(scope)')==oldtx.count('captured.readOptions(scope)'))
ck('RUlegacyShape', '  const legacy = { scope, close };\n  if (!prepared) return Object.freeze(legacy);' in ru)
installed=(R/'apps/bdptest/src/reference-deployment-lifecycle-installed.test.ts').read_text();ck('I21oldOrder',installed.index('await f.publicRefusal();',installed.index('"I21 ordinary'))<installed.index('await f.guard("invalid-ipc");',installed.index('"I21 ordinary'))<installed.index('const session = await f.launch(true);',installed.index('"I21 ordinary'))<installed.index('await f.invalidPostReadyControl(session);',installed.index('"I21 ordinary'))<installed.index('await f.ownerProbe(false);',installed.index('"I21 ordinary')))
ck('wireGateBeforeOutput','unsafeLifecycleWire(f.records, f.controls, f.secrets)' in installed and installed.index('unsafeLifecycleWire(f.records, f.controls, f.secrets)')<installed.index('else console.log(text)'))
contract=load(P/'contract/installed-witnesses.json');ck('21Callbacks',len(contract['callbacksTable'])==21);ck('126HTTP',sum(r['scheduledHttp'] for r in contract['callbacksTable'])==126);ck('33Ordinary',sum(r['candidateGenerations'] for r in contract['callbacksTable'])==33);ck('4Guards',sum(r.get('guardHelperChildren',0) for r in contract['callbacksTable'])==4);ck('51source',sum(r['callbacks'] for r in load(P/'contract/source-witnesses.json')['finiteCallbacks'])==51)
fmt=load(P/'formatter-invocation.json');fmr=load(P/'formatter-result.json');ck('oneExactFormat6',fmt['argv'][1:3]==['format','--write'] and len(fmt['argv'][3:])==6 and set(fmt['argv'][3:])==set(sb['changedPaths'])-{'docs/design/startup-configuration.md'} and fmr['exit']==0 and fmr['invocations']==1 and fmr['formatCheckRan'] is False);ck('formatterBinaryPin',sha(Path(fmt['argv'][0]).read_bytes())==fmt['executableSha256']);ck('formatterRaw',sha((P/'formatter.stdout.log').read_bytes())==fmr['stdoutSha256'] and sha((P/'formatter.stderr.log').read_bytes())==fmr['stderrSha256'])
root=L/'write-trusted-startup-lifecycle-root-source-inspection-3';ri=load(root/'inspection.json')
for key in ['actualSevenBinaryPatch','actualNineBinaryPatch']:ck('rootBinary:'+key,sha(Path(ri[key]['path']).read_bytes())==ri[key]['sha256'])
ck('allFlagsFalse',all(x is False for k,x in load(P/'controls.json').items() if k!='sourceAuthoringComplete'))
v={'state':'Independent full corrected native source static inspection; no source/helper/runtime/Git execution','accepted':True,'checks':len(C),'allPassed':all(C.values()),'authorIndexSha256':sha((P/'index.json').read_bytes()),'sourceBindingSha256':sha((P/'source-binding.json').read_bytes()),'afterMapSha256':sha((P/'source-after-map.json').read_bytes()),'sourceIdentity':sb['derivedSourceIdentityJsonDefaultSortedMap'],'rootInspectionSha256':sha((root/'inspection.json').read_bytes()),'actualSevenBinaryPatchSha256':ri['actualSevenBinaryPatch']['sha256'],'actualNineBinaryPatchSha256':ri['actualNineBinaryPatch']['sha256'],'counts':{'physical':770,'origins':736,'source':686,'changed':7,'protectedCandidate':679,'protectedFdd':677,'sourceCallbacks':51,'persisted':21,'sourceHTTP':2,'installedCallbacks':21,'installedHTTP':126,'CONNECT':2,'ordinaryGenerations':33,'guardChildren':4,'children':37},'genuinePatchPairs':4,'runtimeValidation':'UNRUN','secondReviewAllowed':False}
Path('/tmp/lifecycle-source3-native-verification.json').write_text(json.dumps(v,indent=2)+'\n');print(json.dumps(v,indent=2))
