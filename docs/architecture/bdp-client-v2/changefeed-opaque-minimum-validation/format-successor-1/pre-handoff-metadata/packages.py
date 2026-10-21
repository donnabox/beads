"""Read-only provenance modes plus external consumer preparation; never imports product.
Modes tar, consumer, installed. All require the Changefeed opaque dirty-source gate; copied narrowly from pinned packages.py.
"""
from pathlib import Path
import json,tarfile,re,copy
import sys
sys.dont_write_bytecode=True
from gate import P,W,E,sha,load_gate,require,verify_build,consumer_inventory
import sys,hashlib,os
if not __debug__: raise RuntimeError('Python assertions must remain enabled')
mode=sys.argv[1]
g,_=load_gate({'directory':'prepare-package-directory','tar':'verify-archives','consumer':'prepare-consumer','installed':'verify-installed-consumer'}[mode],child=True);C=E/'consumer'
expected={'@bdp/protocol':'0.0.0','@bdp/server':'0.0.0','ajv':'8.17.1','ajv-formats':'3.0.1','fast-deep-equal':'3.1.3','fast-uri':'3.1.5','json-schema-traverse':'1.0.0','require-from-string':'2.0.2'}
assert json.loads((P/'commands.json').read_text())['expectedProductionVersions']==expected
assert sha((W/'pnpm-lock.yaml').read_bytes())==g['sourceLockSha256']
def write(name,value):
 with (E/name).open('x') as f:f.write(json.dumps(value,indent=2)+'\n')
if mode=='directory':
 (E/'packages').mkdir(exist_ok=False)
elif mode=='tar':
 require((E/'node-version.log').read_text().strip()=='v24.16.0','Unexpected actual Node version')
 verify_build(g)
 packages={}
 for name in ['protocol','server']:
  archive=E/'packages'/f'bdp-{name}-0.0.0.tgz';files={};root=W/'packages'/name
  with tarfile.open(archive,'r:gz') as tar:
   for member in tar.getmembers():
    assert member.isfile(),member.name
    assert member.name.startswith('package/')
    rel=member.name[len('package/'):];assert '..' not in Path(rel).parts and not Path(rel).is_absolute()
    assert rel not in files
    b=tar.extractfile(member).read()
    source=W/'LICENSE' if rel=='LICENSE' else root/rel
    original=source.read_bytes()
    if rel=='package.json':
     obj=json.loads(original);adjusted=copy.deepcopy(obj)
     if name=='server':
      assert adjusted['dependencies']['@bdp/protocol']=='workspace:*'
      adjusted['dependencies']['@bdp/protocol']='0.0.0'
     assert json.loads(b)==adjusted,'Unexpected package metadata rewrite'
    else:assert b==original,str(source)
    files[rel]={'sha256':sha(b),'bytes':len(b),'source':str(source),'sourceSha256':sha(original)}
  required={'package.json','LICENSE'}
  for sub in (['dist','schemas'] if name=='protocol' else ['dist']):
   required.update(p.relative_to(root).as_posix() for p in (root/sub).rglob('*') if p.is_file())
  assert set(files)==required,(name,set(files)^required)
  packages[name]={'archive':str(archive),'archiveSha256':sha(archive.read_bytes()),'files':files}
 write('tar-source-provenance.json',{'head':g['finalHead'],'tree':g['finalTree'],'clean':False,'dirtyPaths':g['_sourceBinding']['dirtyPaths'],'ownedSourceFiles':g['_sourceBinding']['ownedFiles'],'sourceIdentity':g['_sourceIdentity'],'sourceBindingSha256':g['sourceBinding']['sha256'],'binaryDiffSha256':g['_sourceBinding']['binaryDiffSha256'],'packages':packages,'sourceLockSha256':g['sourceLockSha256'],'sourceAttribution':'Base HEAD/tree plus independently verified reviewed dirty source identity, fresh build and archives. Base commit does not contain the draft; not child-attested source.'})
elif mode=='consumer':
 basis=json.loads((E/'tar-source-provenance.json').read_text());assert basis['head']==g['finalHead']
 C.mkdir(exist_ok=False)
 for data in basis['packages'].values():assert sha(Path(data['archive']).read_bytes())==data['archiveSha256']
 deps={f'@bdp/{n}':'file:'+basis['packages'][n]['archive'] for n in ['protocol','server']}
 (C/'package.json').write_text(json.dumps({'name':'changefeed-opaque-minimum-external-consumer','private':True,'type':'module','dependencies':deps},indent=2)+'\n')
 # pnpm11 reads workspace overrides here. Frozen lock is independently pinned by root gate.
 (C/'pnpm-workspace.yaml').write_text('overrides:\n  "@bdp/protocol": '+json.dumps(deps['@bdp/protocol'])+'\n  "fast-uri": "3.1.5"\n')
elif mode=='installed':
 basis=json.loads((E/'tar-source-provenance.json').read_text());assert basis['head']==g['finalHead'];packages={}
 for name,data in basis['packages'].items():
  archive=Path(data['archive']);assert sha(archive.read_bytes())==data['archiveSha256']
  root=(C/'node_modules/@bdp'/name).resolve(strict=True);assert root.is_relative_to(C.resolve())
  with tarfile.open(archive,'r:gz') as tar:
   for rel,info in data['files'].items():
    b=(root/rel).read_bytes();original=tar.extractfile('package/'+rel).read()
    assert b==original and sha(b)==info['sha256']
    assert sha(Path(info['source']).read_bytes())==info['sourceSha256']
  actual={p.relative_to(root).as_posix() for p in root.rglob('*') if p.is_file()}
  assert actual==set(data['files'])
  packages[name]={'realpath':str(root),'archiveSha256':data['archiveSha256'],'checkedMembers':len(actual)}
 paths=set(p.resolve(strict=True) for p in list((C/'node_modules/.pnpm').glob('*/node_modules/*/package.json'))+list((C/'node_modules/.pnpm').glob('*/node_modules/@*/*/package.json')))
 actual={};graph=[];byroot={}
 def resolve_dep(manifest,dep):
  for ancestor in [manifest.parent,*manifest.parent.parents]:
   if not ancestor.is_relative_to(C.resolve()):break
   probe=ancestor/'node_modules'/dep/'package.json'
   if probe.exists():
    q=probe.resolve(strict=True);assert q.is_relative_to(C.resolve());return q
  raise AssertionError((str(manifest),dep))
 for manifest in sorted(paths):
  assert manifest.is_relative_to(C.resolve());m=json.loads(manifest.read_text());name=m['name'];version=m['version']
  assert expected[name]==version and name not in actual;actual[name]=version;deps={}
  for dep in sorted(m.get('dependencies',{})):
   q=resolve_dep(manifest,dep);dm=json.loads(q.read_text());assert dm['name']==dep and dm['version']==expected[dep]
   deps[dep]={'manifestRealpath':str(q),'version':dm['version']}
  graph.append({'name':name,'version':version,'manifest':str(manifest),'manifestSha256':sha(manifest.read_bytes()),'resolvedDependencies':deps});byroot[manifest.parent]=m
 assert actual==expected
 require(sha((C/'pnpm-lock.yaml').read_bytes())==g['consumerLockSha256'],'Consumer lock not accepted')
 # Independently reproduce actual installed-payload.ts sorted path-sensitive recursive hash.
 # The fixed closure contains ASCII paths, so Python ordering equals JS's lexical comparison.
 payload=hashlib.sha256();visits=[];filecount=0
 def bind(root,identity,ancestors):
  global filecount
  assert root not in ancestors and len(ancestors)<32
  manifest=root/'package.json';assert not manifest.is_symlink();m=json.loads(manifest.read_bytes());visits.append(identity)
  def emit(rel):
   global filecount
   b=(root/rel).read_bytes();payload.update(f'{identity}/{rel}:{len(b)}:'.encode());payload.update(b);filecount+=1
  def walk(rel):
   p=root/rel;assert not p.is_symlink()
   for f in sorted(p.iterdir(),key=lambda x:x.name):
    assert f.name.isascii()
    if f.name in ('node_modules','.git'):continue
    assert not f.is_symlink(),str(f)
    r=f.relative_to(root).as_posix()
    if f.is_dir():walk(r)
    elif f.is_file() and re.search(r'\.(?:[cm]?js|json|node|wasm)$',f.name):emit(r)
  if m['name'] in ('@bdp/server','@bdp/protocol'):
   emit('package.json');walk('dist')
   if m['name']=='@bdp/protocol':walk('schemas')
  else:walk('')
  for dep in sorted(m.get('dependencies',{})):bind(resolve_dep(manifest,dep).parent,identity+'>'+dep,[*ancestors,root])
 bind((C/'node_modules/@bdp/server').resolve(strict=True),'@bdp/server',[])
 entry=(C/'node_modules/@bdp/server/dist/transaction-installed-process.js').resolve(strict=True);assert entry.is_relative_to(C.resolve())
 write('installed-provenance.json',{'status':'VERIFIED_BEFORE_CHILD','tarProvenanceSha256':sha((E/'tar-source-provenance.json').read_bytes()),'sourceCommit':g['finalHead'],'sourceTree':g['finalTree'],'sourceIdentity':g['_sourceIdentity'],'sourceBindingSha256':g['sourceBinding']['sha256'],'consumer':str(C),'entryRealpath':str(entry),'entrySha256':sha(entry.read_bytes()),'packages':packages,'installedProductionGraph':graph,'expectedProductionVersionsFromFrozenLock':expected,'sourceLockSha256':g['sourceLockSha256'],'consumerLockSha256':sha((C/'pnpm-lock.yaml').read_bytes()),'consumerManifestSha256':sha((C/'package.json').read_bytes()),'consumerSettingsSha256':sha((C/'pnpm-workspace.yaml').read_bytes()),'payloadIdentity':payload.hexdigest(),'payloadRecursiveVisits':visits,'payloadFileVisits':filecount,'limitations':'Read-only independent fixed dependency closure hash; compare against actual child result. Not source attestation or Node/TLS/deployment identity.'})
 write('installed-consumer-inventory.json',consumer_inventory())
else:raise ValueError(mode)
