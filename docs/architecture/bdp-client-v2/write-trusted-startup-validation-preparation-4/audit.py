"""Strict startup validation auditor; reused actual tools/build and exact callback grammar. No product imports."""
import sys
sys.dont_write_bytecode=True
from gate import *
import re
from collections import Counter
mode=sys.argv[1]
stages={'source':'focused-source'}
audit_stages={'tools':'bind-workspace-tools','build':'bind-fresh-build','source':'audit-focused-source','installed':'audit-bound-consumer-loader','smoke':'audit-required-smoke'}
g,_=load_gate(audit_stages[mode],child=True)
if mode=='tools':
 require((E/'node-version.log').read_text().strip()=='v24.16.0' and (E/'pnpm-version.log').read_text().strip()=='11.20.0','Actual runtime/package-manager versions differ')
 tools=dict(document(g['baseTools'])['tools'])
 paths=[W/'node_modules'/n for n in ['@biomejs/biome/bin/biome','vitest/vitest.mjs','typescript/bin/tsc','typescript/lib/tsc.js','typescript/lib/getExePath.js','typescript/package.json']]
 paths += [W/'node_modules/.pnpm/@biomejs+cli-darwin-arm64@2.5.7/node_modules/@biomejs/cli-darwin-arm64/biome',W/'node_modules/.pnpm/@typescript+typescript-darwin-arm64@7.0.2/node_modules/@typescript/typescript-darwin-arm64/lib/tsc']
 for p in paths:tools[str(p)]={'realpath':str(p.resolve(strict=True)),'sha256':sha(p.read_bytes())}
 versions={'@biomejs/biome':'2.5.7','vitest':'4.1.10','typescript':'7.0.2'}
 for name,version in versions.items():require(read(W/'node_modules'/name/'package.json')['version']==version,'Installed tool version differs')
 complete('workspace-install',g)
 write('workspace-tools.json',{'tools':tools,'versions':versions,'head':g['finalHead'],'sourceIdentity':g['_sourceIdentity'],'sourceLockSha256':g['sourceLockSha256'],'installExitSha256':sha((E/'workspace-install.exit.json').read_bytes()),'installLogSha256':sha((E/'workspace-install.log').read_bytes())})
elif mode=='build':
 write('build-artifacts.json',{'head':g['finalHead'],'tree':g['finalTree'],'sourceIdentity':g['_sourceIdentity'],'artifacts':inventory(),**{key:{'path':str(E/('fresh-build.'+suffix)),'sha256':sha((E/('fresh-build.'+suffix)).read_bytes())} for key,suffix in [('start','started.json'),('exit','exit.json'),('log','log')]}})
elif mode=='source':
 stage='focused-source';config=read(P/'commands.json')['commands'][stage];result=read(E/(stage+'.results.json'));chosen=[]
 for suite in result['testResults']:
  for case in suite['assertionResults']:
   if case['fullName'] in config['selectedFullNames']:
    require(Path(suite['name']).resolve()==(W/config['selectedFilesByFullName'][case['fullName']]).resolve(),'Selected callback file differs');chosen.append(case)
   else:require(case['status'] in ['pending','skipped','todo'],'Unselected callback executed')
 require(Counter(x['fullName'] for x in chosen)==Counter(config['selectedFullNames']) and all(x['status']=='passed' for x in chosen) and result.get('success') is True,'Exact callbacks did not pass')
 for key in ['numFailedTests','numFailedTestSuites']:
  require(type(result[key]) is int and result[key]==0,'Failed result '+key)
 if 'numRuntimeErrorTestSuites' in result:
  require(type(result['numRuntimeErrorTestSuites']) is int and result['numRuntimeErrorTestSuites']==0,'Runtime suite errors')
 require(type(result['numPassedTests']) is int and result['numPassedTests']==len(chosen),'Passed count differs')
 complete(stage,g);text=(E/(stage+'.log')).read_text();require(not re.search(r'Unhandled Errors|Unhandled Rejection|Uncaught Exception|Worker exited unexpectedly',text),'Unhandled runtime diagnostic')
 write(stage+'.accepted-observation.json',{'sourceIdentity':g['_sourceIdentity'],'selectedFullNames':config['selectedFullNames'],'actualPassedCallbacks':len(chosen),'resultsSha256':sha((E/(stage+'.results.json')).read_bytes()),'logSha256':sha((E/(stage+'.log')).read_bytes()),'qualification':False,'limitations':'Passing callback assertions, source forbidden-call/FD mocks and store-path checks; no independent kernel FD/socket census, no installed behavior implied.'})
elif mode=='installed':
 complete('bound-consumer-loader',g);proof=verify_consumer(g);lines=(E/'bound-consumer-loader.log').read_text().splitlines();reports=[]
 for line in lines:
  try:o=json.loads(line)
  except json.JSONDecodeError:continue
  if isinstance(o,dict) and o.get('event')=='startup.bound-consumer-loader':reports.append(o)
 require(len(reports)==1,'Exactly one actual installed loader report');o=reports[0]
 require(o['outcome']=='passed' and o['entryRealpath']==proof['entryRealpath'] and o['entrySha256']==proof['entrySha256'] and o['provenanceSha256']==g['installedProvenance']['sha256'],'Actual external loader binding differs')
 require(set(o)==set(read(P/'report-schemas.json')['installed']['exactReportFields']),'Unexpected installed report fields')
 require(o['workTimedOut'] is False and o['cleanupUnknown'] is False and o['temporaryDirectoryRemoved'] is True and not Path(o['temporaryDirectory']).exists(),'Installed witness closure unknown')
 started=read(E/'bound-consumer-loader.started.json');tmp=Path(o['temporaryDirectory']);require(tmp.is_relative_to(Path(started['temporaryDirectory'])) and tmp.name.startswith('bdp-bound-loader-'),'Installed temp ownership differs')
 require(o['publicExports']==['ReferenceDeploymentError','loadReferenceDeployment'] and o['profiles']==['read-update','transactional'] and o['loaderCalls']==6 and o['accepted']==2 and o['contractRefusals']==4,'Exact copied loader witness totals')
 require(o['configuredPortAcquisitionExcluded'] is True and o['storeDirectoriesAbsent'] is True and type(o['privateSpecifierCount']) is int and o['privateSpecifierCount']==71,'Installed pure boundary observation missing')
 write('bound-consumer-loader.accepted-observation.json',{'sourceIdentity':g['_sourceIdentity'],'observation':o,'qualification':False,'limitations':'Owned occupied intended port excludes successful acquisition, not attempted binds; no installed per-FD census or ordinary profile/service activation. Payload digest is independent provenance, not loader-attested.'})
elif mode=='smoke':
 complete('required-smoke',g);text=(E/'required-smoke.log').read_text()
 require(text.count('installed pure reference deployment loader: RU/TX, refusals and inert boundary passed')==1,'Actual stock loader smoke completion missing')
 require(not re.search(r'Unhandled Errors|Unhandled Rejection|Uncaught Exception',text),'Unhandled smoke error')
 write('required-smoke.accepted-observation.json',{'sourceIdentity':g['_sourceIdentity'],'logSha256':sha((E/'required-smoke.log').read_bytes()),'qualification':False,'limitations':'Stock smoke creates another consumer then removes it; root-pinned consumer proof is NOT attributed to its loader executions. No retained internal archive/lock proof unless separately actually captured.'})
else:raise ValueError(mode)
