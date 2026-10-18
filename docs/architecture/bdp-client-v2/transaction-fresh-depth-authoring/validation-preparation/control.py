"""Separate single emitted Snapshot-comparator counterfactual; no positive/source mutation."""
from gate import *
import sys,shutil,re
mode=sys.argv[1];g,spec=load_gate('prepare-control' if mode=='prepare' else 'review-control');C=P/'mirror'
reporter_pins=json.loads((P/'control-reporter-sources.json').read_text())
for relative,h in reporter_pins.items():require(sha((W/relative).read_bytes())==h,'Reporter source changed '+relative)
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
 for relative,h in reporter_pins.items():require(before[relative]==after[relative]==h,'Copied reporter differs '+relative)
 record={'reporterSources':reporter_pins,'combinedHead':g['finalHead'],'positiveFiles':before,'copyFiles':after,'links':linkproof,'changedFiles':changed,'old':old,'new':new}
 with (E/'control-mirror.json').open('x') as f:f.write(json.dumps(record,indent=2)+'\n')
 verify_mirror();print(json.dumps({'copiedFiles':len(before),'links':len(links),'changedFiles':changed}))
elif mode=='review':
 x=completed('strict-control');r=json.loads((E/'strict-control.results.json').read_text())
 cases=[c for f in r['testResults'] for c in f.get('assertionResults',[]) if c['status'] not in ['pending','skipped','todo']]
 require(len(cases)==1 and cases[0]['fullName']=='captures and reopens deep owned Links without losing JSON value equality' and cases[0]['status']=='failed','Wrong negative callback')
 errors=cases[0].get('failureMessages',[])
 require(len(errors)==1 and 'fresh-depth:snapshot-capture' in errors[0],'Missing unique JSON wrapper failure')
 require(all(not f.get('message') and not f.get('failureMessage') for f in r['testResults']),'File/setup failure')
 # JSON omits cause; bind only the named verbose failed-test block, not global text/code frames.
 rawlog=(E/'strict-control.log').read_text();plain=re.sub(r'\x1b\[[0-?]*[ -/]*[@-~]','',rawlog);lines=plain.splitlines()
 title='captures and reopens deep owned Links without losing JSON value equality'
 heading='FAIL  packages/server/src/transaction-snapshot.test.ts > '+title
 starts=[i for i,line in enumerate(lines) if line.strip().startswith('FAIL ')]
 require(len(starts)==1 and lines[starts[0]].strip()==heading,'Wrong/ambiguous verbose failed-test block')
 start=starts[0];ends=[i for i in range(start+1,len(lines)) if re.fullmatch(r'[⎯─━-]+\[1/1\][⎯─━-]+',lines[i].strip())]
 require(len(ends)==1,'Missing unique failed-test block divider');end=ends[0];block=lines[start+1:end]
 wrappers=[i for i,line in enumerate(block) if line=='Error: fresh-depth:snapshot-capture']
 causes=[i for i,line in enumerate(block) if line=='Caused by: RangeError: Maximum call stack size exceeded']
 require(len(wrappers)==len(causes)==1 and wrappers[0]<causes[0],'Missing actual wrapper then nested RangeError in named block')
 require(sum(line.startswith('Caused by: ') for line in block)==1,'Unexpected additional cause')
 direct_path=E/'strict-control.direct-exit.json';direct=json.loads(direct_path.read_text())
 require(direct['stage']=='strict-control' and direct['directReturnCode']==1 and direct['terminationSignal'] is None and not direct['errors'],'Wrong direct Vitest termination')
 require(direct['argv']==json.loads(COMMANDS.read_text())['commands']['strict-control']['directArgv'] and direct['cwd']==str(C),'Wrong direct invocation')
 require(direct['commandsSha256']==sha(COMMANDS.read_bytes()) and direct['mirrorSha256']==sha((E/'control-mirror.json').read_bytes()),'Direct invocation binding changed')
 require(direct['sourceBefore']==direct['sourceAfter']==source_hashes() and x['exit']==0,'Direct source or guarded adapter failure')
 require(r['numFailedTests']==1 and r['numPassedTests']==0 and not r.get('unhandledErrors') and not r.get('errors') and not r.get('numRuntimeErrorTestSuites',0),'Unrelated negative failure')
 verify_mirror();record={'accepted':True,'actualDirectVitestExit':1,'actualGuardExit':x['exit'],'supervisorExit':x['supervisorExit'],'directExitRecordSha256':sha(direct_path.read_bytes()),'verboseFailedBlock':{'startLine':start+1,'endLine':end+1,'sha256':sha(('\n'.join(lines[start:end+1])+'\n').encode())},'http':0,'httpAttribution':'Pinned selected callback uses explicit emitted owner and no HTTP; not independent packet capture','resultSha256':sha((E/'strict-control.results.json').read_bytes()),'logSha256':sha((E/'strict-control.log').read_bytes()),'limitation':'Named catch surrounds capture only after successful deep Link commit/assertions. Exact nested verbose RangeError, direct Vitest exit1 and supervisor cleanup mandatory. Root must inspect the first actual captured named block. Source aliases remain copied source; only owner and comparator closure explicitly emitted.'}
 with (E/'strict-control.adjudication.json').open('x') as f:f.write(json.dumps(record,indent=2)+'\n')
 print(json.dumps({'acceptedExpectedFailure':True,'http':0}))
else:raise ValueError(mode)
