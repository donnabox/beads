"""Separate single emitted strict-preparation counterfactual; no positive/source mutation."""
from gate import *
import sys,shutil,re
mode=sys.argv[1];g,spec=load_gate('prepare-control' if mode=='prepare' else 'review-control');C=P/'mirror'
if mode=='prepare':
 require(not C.exists(),'Control copy exists; no retry');C.mkdir()
 files=set(git('ls-files').decode().splitlines());links={}
 # Copy actual full dependency/tool and emitted trees, not a workspace fallback.
 roots=[W/'node_modules']
 for parent in ['packages','apps']:
  for root in (W/parent).iterdir():
   if root.is_dir():roots.extend([root/'node_modules',root/'dist'])
 for root in roots:
  if not root.exists():continue
  for directory,dirs,names in os.walk(root,followlinks=False):
   # Vitest rewrites these generated caches; retain actual tool/product/link pins.
   if Path(directory).name == "node_modules":
    dirs[:]=[name for name in dirs if name not in (".vite", ".vite-temp") or (Path(directory)/name).is_symlink()]
   for name in dirs+names:
    q=Path(directory)/name;r=q.relative_to(W).as_posix()
    if q.is_symlink():links[r]=q
    elif q.is_file():files.add(r)
 before={}
 for r in sorted(files):
  q=W/r;require(q.is_file() and not q.is_symlink(),'Unexpected source file link '+r)
  out=C/r;out.parent.mkdir(parents=True,exist_ok=True);shutil.copy2(q,out);before[r]=sha(q.read_bytes())
 linkproof={}
 for r,q in sorted(links.items()):
  target=q.resolve(strict=True);require(target.is_relative_to(W),'External dependency link '+r)
  dest=C/target.relative_to(W);require(dest.exists(),'Missing copied link target '+r)
  out=C/r;out.parent.mkdir(parents=True,exist_ok=True);out.symlink_to(os.path.relpath(dest,out.parent),target_is_directory=dest.is_dir())
  linkproof[r]={'link':os.readlink(out),'realpath':str(out.resolve(strict=True))}
 target=C/'packages/server/dist/transaction-batch.js';original=target.read_text()
 old='export function prepareTransactionSequenceForReceiving(scope, text) {\n    return prepareSequence(scope, text);\n}'
 new='export function prepareTransactionSequenceForReceiving(scope, text) {\n    return prepareTransactionSequence(scope, text);\n}'
 require(original.count(old)==1,'Exact emitted receiving function changed; stop before mutation')
 target.write_text(original.replace(old,new))
 after={r:sha((C/r).read_bytes()) for r in before};changed=[r for r in before if before[r]!=after[r]]
 require(changed==['packages/server/dist/transaction-batch.js'],'Mutation scope differs')
 record={'combinedHead':g['finalHead'],'positiveFiles':before,'copyFiles':after,'links':linkproof,'changedFiles':changed,'old':old,'new':new}
 with (E/'control-mirror.json').open('x') as f:f.write(json.dumps(record,indent=2)+'\n')
 verify_mirror();print(json.dumps({'copiedFiles':len(before),'links':len(links),'changedFiles':changed}))
elif mode=='review':
 x=completed('strict-control');r=json.loads((E/'strict-control.results.json').read_text())
 cases=[c for f in r['testResults'] for c in f.get('assertionResults',[]) if c['status'] not in ['pending','skipped','todo']]
 require(len(cases)==1 and cases[0]['fullName']=='compares retained sequence members before private depth refusal and preserves receiving failures' and cases[0]['status']=='failed','Wrong negative callback')
 errors=cases[0].get('failureMessages',[]);require(len(errors)==1 and 'sequence-depth:retained-member-conflict' in errors[0] and 'AssertionError' in errors[0],'Wrong named negative failure')
 require(r['numFailedTests']==1 and r['numPassedTests']==0 and not r.get('unhandledErrors') and not r.get('errors') and not r.get('numRuntimeErrorTestSuites',0),'Unrelated negative failure')
 obs=[]
 for line in (E/'strict-control.log').read_text().splitlines():
  if line.startswith('{'):
   try:v=json.loads(line)
   except json.JSONDecodeError:continue
   if v.get('observation')=='private-source-sequence-depth':obs.append(v)
 require(len(obs)==1,'Missing unique actual source observation');o=obs[0]
 require(o['attempts']==2 and o['outcome']=='failed' and o['failureCount']==1 and o['cleanupCompleted'] and not o['deadlineExceeded'] and not o['cleanupExceeded'] and o['claimEligible'] is False,'Wrong failure/lifecycle')
 responses=o['completedResponses'];require(len(responses)==2,'Wrong actual responses');seed,fail=responses
 require(seed['status']==200 and seed['headers']['bdp-scope-position']=='pos-1' and seed['responseBodyBytes']>0,'Missing seed')
 require(fail['status']==500 and fail['responseBodyBytes']==0 and fail['responseBodySha256']==sha(b''),'Wrong strict refusal')
 require(all(fail['headers'][k] is None for k in ['bdp-scope-epoch','bdp-authorization-view','bdp-scope-position','x-bdp-development-limit']),'Strict refusal discloses context/marker')
 require(fail['method']=='POST' and fail['url'].endswith('/acme/operations/sequence') and seed['method']=='POST' and seed['url'].endswith('/acme/operations/create-bead'),'Wrong actual routes')
 verify_mirror();record={'accepted':True,'actualRawExit':1,'actualHTTP':2,'observation':o,'resultSha256':sha((E/'strict-control.results.json').read_bytes()),'logSha256':sha((E/'strict-control.log').read_bytes()),'limitations':'Exactly one expected source counterfactual, not passing positive execution or product qualification.'}
 with (E/'strict-control.adjudication.json').open('x') as f:f.write(json.dumps(record,indent=2)+'\n')
 print(json.dumps({'acceptedExpectedFailure':True,'actualHTTP':2}))
else:raise ValueError(mode)
