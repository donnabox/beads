"""Separate single emitted Snapshot-comparator counterfactual; no positive/source mutation."""
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
 target=C/'packages/server/dist/transaction-snapshot.js';original=target.read_text()
 old='jsonValueEqual(link, records.get(link.id.slice(scope.length))?.value)'
 new='isDeepStrictEqual(link, records.get(link.id.slice(scope.length))?.value)'
 require('import { isDeepStrictEqual } from "node:util";' in original,'Existing shallow equality import missing')
 require(original.count(old)==1,'Exact emitted receiving function changed; stop before mutation')
 target.write_text(original.replace(old,new))
 after={r:sha((C/r).read_bytes()) for r in before};changed=[r for r in before if before[r]!=after[r]]
 require(changed==['packages/server/dist/transaction-snapshot.js'],'Mutation scope differs')
 record={'combinedHead':g['finalHead'],'positiveFiles':before,'copyFiles':after,'links':linkproof,'changedFiles':changed,'old':old,'new':new}
 with (E/'control-mirror.json').open('x') as f:f.write(json.dumps(record,indent=2)+'\n')
 verify_mirror();print(json.dumps({'copiedFiles':len(before),'links':len(links),'changedFiles':changed}))
elif mode=='review':
 x=completed('strict-control');r=json.loads((E/'strict-control.results.json').read_text())
 cases=[c for f in r['testResults'] for c in f.get('assertionResults',[]) if c['status'] not in ['pending','skipped','todo']]
 require(len(cases)==1 and cases[0]['fullName']=='captures and reopens deep owned Links without losing JSON value equality' and cases[0]['status']=='failed','Wrong negative callback')
 errors=cases[0].get('failureMessages',[])
 require(len(errors)==1 and all(s in errors[0] for s in ['fresh-depth:snapshot-capture','RangeError','Maximum call stack size exceeded']),'Missing actual named capture RangeError; generic failure is not evidence')
 require(r['numFailedTests']==1 and r['numPassedTests']==0 and not r.get('unhandledErrors') and not r.get('errors') and not r.get('numRuntimeErrorTestSuites',0),'Unrelated negative failure')
 verify_mirror();record={'accepted':True,'actualRawExit':1,'http':0,'httpAttribution':'Pinned selected callback uses explicit emitted owner and no HTTP; not independent packet capture','resultSha256':sha((E/'strict-control.results.json').read_bytes()),'logSha256':sha((E/'strict-control.log').read_bytes()),'limitation':'Named catch surrounds capture only after successful deep Link commit/assertions. Exact RangeError and supervisor cleanup mandatory. Source aliases remain copied source; only owner and comparator closure explicitly emitted.'}
 with (E/'strict-control.adjudication.json').open('x') as f:f.write(json.dumps(record,indent=2)+'\n')
 print(json.dumps({'acceptedExpectedFailure':True,'http':0}))
else:raise ValueError(mode)
