"""Root-only data binding/release preparation. No product commands, Git writes or main reads.
Invoke through guarded-action.py after the actual ordinary branch merge, with exact root-supplied commits.
"""
from pathlib import Path
import argparse,datetime,hashlib,json,os,re,subprocess,sys
B=Path(__file__).resolve().parent;P=B.parent;L=P.parent
W=Path('/Users/dbox/repos/wt/vickie/vickie-sequence-depth-retained')
INPUTS_SHA='c53ca7dc4b18b826dd0cd11306ad59c3adc3918ecffaaedc35f6cc05e9230648'
sha=lambda b:hashlib.sha256(b).hexdigest()
def require(ok,message):
 if not ok:raise RuntimeError(message)
def git(*args):return subprocess.check_output(['git',*args],cwd=W,timeout=30)
def read(p):return json.loads(p.read_text())
def check_input(k):
 d=inputs[k];p=Path(d['path']);require(sha(p.read_bytes())==d['sha256'],'Input changed: '+k);return p

def tree(ref):
 result={}
 for entry in git('ls-tree','-rz',ref).split(b'\0'):
  if not entry:continue
  fields,name=entry.split(b'\t',1);mode,kind,oid=fields.decode().split();name=name.decode()
  require(name not in result and kind=='blob' and mode in ['100644','100755'],'Unexpected tree entry '+name)
  result[name]={'mode':mode,'type':kind,'object':oid}
 return result

def blob(ref,path):return git('show',ref+':'+path)
def parents(ref):return git('show','-s','--format=%P',ref).decode().split()
def pin(p):return {'path':str(p),'sha256':sha(p.read_bytes())}
def write(name,value):
 with (P/name).open('x') as f:f.write(json.dumps(value,indent=2)+'\n')

parser=argparse.ArgumentParser()
parser.add_argument('--sequence-author',required=True)
parser.add_argument('--combined-head',required=True)
parser.add_argument('--combined-tree',required=True)
parser.add_argument('--root-authorization',required=True)
parser.add_argument('--root-authorization-sha256',required=True)
a=parser.parse_args()
for v in [a.sequence_author,a.combined_head,a.combined_tree]:require(re.fullmatch('[0-9a-f]{40}',v),'Exact full Git ID required')
require(sha((B/'inputs.json').read_bytes())==INPUTS_SHA,'Binder inputs changed');inputs=read(B/'inputs.json')
for k in inputs:
 if k!='toolExpectedHashes':check_input(k)
# Session proof is read-only and does not inspect main. Root also wraps this script in the guard.
subprocess.run([sys.executable,str(check_input('guard')),sys.executable,'-c','pass'],check=True,timeout=30)
authority=Path(a.root_authorization)
require(authority.is_absolute() and authority.is_relative_to(P) and not authority.is_symlink() and sha(authority.read_bytes())==a.root_authorization_sha256,'Root authorization absent/changed')
permission=read(authority)
require(permission['accepted'] is True and permission['head']==a.combined_head and permission['sourceOnly'] is True,'Root source-only authorization required')
outputs=['final-source.json','input-binding.json','root-release.json','binding-verification.json','execution-gate.json']
require(all(not (P/n).exists() for n in outputs),'Existing binding/release: preserve it; no overwrite')
require(not (P/'run').exists(),'Execution evidence already exists')
m=read(check_input('integrationMap'));base=m['commonBase'];minimum=m['minimum']['head'];seq=read(check_input('sequenceBinding'));expected=m['files']
require(base=='df80f8cb86b7763bc7bb005492e43d8840527045' and minimum=='5ca22ec4e710cc4417a80d27bff69e1d8424c115','Branch basis differs')
require(len(expected)==20 and len(seq['files'])==7 and m['sequence']['bindingSha256']==inputs['sequenceBinding']['sha256'],'Current union/source binding differs')
require(git('rev-parse','HEAD').decode().strip()==a.combined_head and git('rev-parse','HEAD^{tree}').decode().strip()==a.combined_tree,'Actual clean candidate differs')
require(not git('status','--porcelain').strip(),'Working branch must be clean')
require(parents(a.sequence_author)==[base],'Sequence author must be direct ordinary base child')
require(parents(a.combined_head)==[a.sequence_author,minimum],'Combined ordered parents differ')
require(git('rev-parse',minimum+'^{tree}').decode().strip()==m['minimum']['tree'],'Minimum tree differs')
require(set(git('diff','--name-only',base,a.sequence_author).decode().splitlines())==set(seq['files']),'Sequence author scope differs')
for path,item in seq['files'].items():require(sha(blob(a.sequence_author,path))==item['sha256'],'Author seven frozen bytes differ: '+path)
require(set(git('diff','--name-only',base,a.combined_head).decode().splitlines())==set(expected),'Combined20 scope differs')
old=tree(base);current=tree(a.combined_head);union=set(expected)
require({n:v for n,v in old.items() if n not in union}=={n:v for n,v in current.items() if n not in union},'Outside-union object/mode/type drift')
files={}
for path,item in expected.items():
 data=blob(a.combined_head,path);p=W/path
 require(path in current and current[path]['mode']==old.get(path,{'mode':'100644'})['mode'],'Union mode changed '+path)
 require(len(data)==item['bytes'] and sha(data)==item['sha256'],'Combined mapped bytes differ '+path)
 require(not p.is_symlink() and p.read_bytes()==data,'Actual working bytes differ '+path)
 files[path]={'sha256':sha(data),'bytes':len(data)}
# All tracked build inputs, not only authored files, are tied to exact Git blobs.
tracked={}
for path,item in current.items():
 p=W/path;data=p.read_bytes()
 require(not p.is_symlink() and sha(data)==sha(blob(a.combined_head,path)),'Tracked input differs '+path)
 tracked[path]=sha(data)
sc=read(check_input('sequenceCouncil'));mc=read(check_input('minimumCouncil'));committed=read(check_input('minimumCommitted'))
require(sc['codeReviewAccepted'] is True and sc['sourceBindingSha256']==inputs['sequenceBinding']['sha256'],'Sequence council not bound')
require(sc['actualSupplementSha256']==inputs['sequenceClaudeOutput']['sha256'] and sc['nativeFindingsSha256']==inputs['sequenceNative']['sha256'] and sc['rootSupplementDispositionsSha256']==inputs['sequenceDispositions']['sha256'],'Sequence findings chain differs')
require(mc['accepted'] is True and mc['sourceBindingSha256']==inputs['minimumBinding']['sha256'] and mc['nativeAuditSha256']==inputs['minimumNative']['sha256'] and mc['actualClaudeStdoutSha256']==inputs['minimumClaudeOutput']['sha256'],'Minimum council chain differs')
require(committed['head']==minimum and committed['tree']==m['minimum']['tree'] and committed['councilSha256']==inputs['minimumCouncil']['sha256'],'Minimum committed council differs')
for key in ['sequenceClaudeExit','minimumClaudeExit']:
 x=read(check_input(key));require(x['exitCode']==0 and not x['timeout'] and not x['forcedTermination'] and x['sourceAndInputsUnchanged'] and x['cleanupClear'],'Incomplete review '+key)
inv=read(check_input('preparationInventory'));require(len(inv['files'])==15,'Expected15 preparation files')
for name,h in inv['files'].items():require(sha((P/name).read_bytes())==h,'Preparation changed '+name)
commands=read(P/'commands.json');source=commands['phases']['source']
require(source==['typecheck','fresh-build','format-check','lint','boundaries','read-seal','preparation-positive','sequence-positive','http-positive','minimum-http','build-inventory'] and commands['executionOrder'][:len(source)]==source,'Source prefix differs')
require(sum(commands['commands'][n].get('cases',0) for n in source)==23 and sum(commands['commands'][n].get('http',0) for n in source)==193,'Selection/workload differs')
tools={'node':Path('/Users/dbox/.nvm/versions/node/v24.16.0/bin/node'),'pnpm':Path('/Users/dbox/.cache/node/corepack/v1/pnpm/11.20.0/bin/pnpm.cjs'),'vitest':W/'node_modules/vitest/vitest.mjs','biome':W/'node_modules/@biomejs/biome/bin/biome'}
tool_hashes={};tool_details={}
for name,path in tools.items():
 digest=sha(path.read_bytes());require(digest==inputs['toolExpectedHashes'][name],'Pinned tool differs '+name)
 tool_hashes[str(path)]=digest;tool_details[name]={'path':str(path),'realpath':str(path.resolve(strict=True)),'sha256':digest}
lock=sha((W/'pnpm-lock.yaml').read_bytes());require(lock==tracked['pnpm-lock.yaml'],'Lock inventory mismatch')
final={'workspace':str(W),'head':a.combined_head,'tree':a.combined_tree,'parents':parents(a.combined_head),'sequenceAuthor':a.sequence_author,'minimumHead':minimum,'diffBase':base,'diffSha256':sha(git('diff','--binary',base,a.combined_head)),'files':files,'branch':git('branch','--show-current').decode().strip()}
verification={'checkedAt':datetime.datetime.now(datetime.timezone.utc).isoformat(),'mechanicalChecksPassed':True,'independentRootAcceptanceClaimed':False,'sourceOnly':True,'actualHead':a.combined_head,'actualTree':a.combined_tree,'orderedParents':final['parents'],'sequenceAuthor':a.sequence_author,'unionPaths':20,'outsideUnionEntries':len(current)-20,'fullTreeEntries':len(current),'trackedInputs':len(tracked),'integrationMap':inputs['integrationMap'],'boundEvidence':inputs,'toolInventory':tool_details,'lockSha256':lock,'preparationFiles':inv['files'],'scope':'Data-only checks of working branches; no main inspection, product import, setup, build, tests, packaging or service executed.'}
release={'accepted':True,'head':a.combined_head,'tree':a.combined_tree,'sourceOnly':True,'authorizedStages':source,'rootInvocationAuthorization':pin(authority),'basis':'Root-invoked mechanical binding checks plus pinned existing council/audit closures; not a new independent code review. Controls/package/installed remain closed.'}
g=read(P/'execution-gate.example.json');g.update(released=True,finalHead=a.combined_head,finalTree=a.combined_tree,approvedCommandIds=source,phaseRelease={k:k=='source' for k in g['phaseRelease']},preparationSha256=inv['files'],toolHashes=tool_hashes,sourceLockSha256=lock)
g['preparationSha256'].update({'binding-preparation/bind-source.py':sha(Path(__file__).read_bytes()),'binding-preparation/inputs.json':INPUTS_SHA})
# All preflight checks precede writes; live gate is always last and exclusive.
write('final-source.json',final);write('input-binding.json',tracked);write('binding-verification.json',verification);write('root-release.json',release)
g['sourceBinding']=pin(P/'final-source.json');g['inputBinding']=pin(P/'input-binding.json');g['rootAcceptance']=pin(P/'root-release.json')
require(not git('status','--porcelain').strip() and git('rev-parse','HEAD').decode().strip()==a.combined_head,'Source changed during binding')
for k in inputs:
 if k!='toolExpectedHashes':check_input(k)
write('execution-gate.json',g)
print(json.dumps({'releasedSourcePrefixOnly':source,'head':a.combined_head,'tree':a.combined_tree,'gateSha256':sha((P/'execution-gate.json').read_bytes()),'actualExecutionStarted':False}))
