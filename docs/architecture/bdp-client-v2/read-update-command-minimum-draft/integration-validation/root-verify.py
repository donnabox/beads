from pathlib import Path
import json,hashlib,subprocess,tarfile,datetime
p=Path(__file__).parent; w=Path('/Users/dbox/repos/wt/vickie/vickie-update-cold-directory');h=lambda x:hashlib.sha256(x).hexdigest();git=lambda *a:subprocess.check_output(['git',*a],cwd=w)
f=json.loads((p/'final-source.json').read_text());a=json.loads((p/'actual-results.json').read_text())
assert git('rev-parse','HEAD').decode().strip()==f['head']=='3bb2f2449ce345e118d04c07eadd777363add544'
assert not git('status','--porcelain');assert git('rev-parse','HEAD^{tree}').decode().strip()==f['tree']
assert git('show','-s','--format=%P','HEAD').decode().split()==f['parents']
assert {k:h((w/k).read_bytes()) for k in f['sourceFiles']}==f['sourceFiles']
def tree(ref):return {x.split('\t',1)[1]:x.split('\t',1)[0] for x in git('ls-tree','-r',ref).decode().splitlines()}
base=tree('df80f8cb86b7763bc7bb005492e43d8840527045');actual=tree('HEAD');changed=set(git('diff','--name-only','df80f8cb','HEAD').decode().splitlines());assert len(changed)==5
expected=dict(base);author=tree(f['parents'][0]);expected.update({n:author[n] for n in changed});assert expected==actual
for s in a['stages']:
 r=p/'run';e=json.loads((r/(s['stage']+'.exit.json')).read_text());assert e['exit']==e['supervisorExit']==0 and not e['timeout'] and e['sourceUnchanged'];assert h((r/(s['stage']+'.log')).read_bytes())==s['logSha256'];assert h((r/(s['stage']+'.exit.json')).read_bytes())==s['exitSha256'];assert e['sourceSha256After']==f['sourceFiles']
 if e['adjudication']['kind']=='exact-source-selection':
  rp=next(x.split('=',1)[1] for x in e['argv'] if x.startswith('--outputFile='));v=json.loads(Path(rp).read_text());passed=[t['fullName'] for b in v['testResults'] for t in b['assertionResults'] if t['status']=='passed'];assert sorted(passed)==sorted(e['adjudication']['passedFullNames']);assert v['numFailedTests']==0
for name,count,npass,nfail in [('selected-positive',11,47,0),('selected-minimum-control',2,7,1)]:
 r=json.loads((p/'run'/ (name+'-minimum.json')).read_text());assert r['sourceCommit']==f['head'] and r['cleanupComplete'] and not r['claimEligible'];c=r['commandMinimum'];assert len(c['dispatches'])==len(c['completions'])==count
 assertions=[(e['request']['id'],t['id'],t['passed']) for e in r['result']['exchanges'] for t in e['assertions']];assert sum(x[2] for x in assertions)==npass;bad=[x[:2] for x in assertions if not x[2]];assert bad==([] if nfail==0 else [('singleton-replay','outcome')]);assert all(x['headersMatched'] and x['bodyValidated'] for x in c['completions']);assert c['applications']==nfail
 assert r['baselineReceipt']['payloadIdentity']==a['payloadIdentity']
v=json.loads((p/'run/installed-provenance.json').read_text());members=0
for name,info in v['packages'].items():
 archive=p/'run/packages'/('bdp-'+name+'-0.0.0.tgz');assert h(archive.read_bytes())==info['archiveSha256']
 with tarfile.open(archive) as t:
  for m in t.getmembers():
   if not m.isfile():continue
   rel=m.name.removeprefix('package/');data=t.extractfile(m).read();assert data==(Path(info['realpath'])/rel).read_bytes();src=w/'packages'/name/rel
   if rel=='LICENSE':src=w/'LICENSE'
   if rel=="package.json":
    expectedManifest=json.loads(src.read_bytes())
    if name=="server":
     assert expectedManifest["dependencies"]["@bdp/protocol"]=="workspace:*"
     expectedManifest["dependencies"]["@bdp/protocol"]="0.0.0"
    assert json.loads(data)==expectedManifest
   else: assert data==src.read_bytes(),rel
   members+=1
assert members==317;assert h(Path(v['entryRealpath']).read_bytes())==v['entrySha256']
live={int(x.split()[0]) for x in subprocess.check_output(['ps','-axo','pid='],text=True).splitlines() if x.strip()};assert not live.intersection(a['sampledPids'])
for e in (p/'run').glob('*.exit.json'):
 q=json.loads(e.read_text());assert not q.get('runtimeDirectoriesStillPresent');assert all(not Path(x).exists() for x in q.get('sampledTempEntries',[]) if 'bdp-cumulative-ru-' in x)
out={'checkedAt':datetime.datetime.now(datetime.timezone.utc).isoformat(),'head':f['head'],'tree':f['tree'],'wholeTreeEqualsMainPlusFiveAuthorBlobs':True,'sixReviewedHashesVerified':True,'seventeenStageLogsAndExitHashesVerified':True,'cpuCases':8,'installedPositiveResponses':11,'positiveAssertions':47,'controlResponses':2,'soleExpectedControlFailure':'singleton-replay.outcome','sourceArchiveInstalledMembersVerified':members,'sampledPidsAbsent':len(a['sampledPids']),'limitations':'Full aggregate retained only at historical b895; final exact-head CI required. Private partial evidence, not qualification.'}
(p/'root-integration-acceptance.json').write_text(json.dumps(out,indent=2)+'\n');print(json.dumps(out))
