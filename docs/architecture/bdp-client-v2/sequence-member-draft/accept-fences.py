from pathlib import Path
import json, hashlib, subprocess, datetime
D=Path(__file__).resolve().parent;W=Path('/Users/dbox/repos/wt/vickie/vickie-sequence-member-evidence');SA=D/'source-author';EP=D/'execution-preparation'
sha=lambda b:hashlib.sha256(b).hexdigest()
def write(p,v):
 with p.open('x') as f:json.dump(v,f,indent=2);f.write('\n')
def artifact(p):return {'path':str(p),'sha256':sha(p.read_bytes())}
s=json.loads((SA/'source-binding.json').read_text());m=json.loads((SA/'measurements.json').read_text())
assert sha((SA/'source-binding.json').read_bytes())=='6dc92a67f17acdac93c4b2d6c27ec832d67ad6614c7a70a00ff7ebe407fda71c'
assert sha((SA/'measurements.json').read_bytes())=='04d0c9577be30231dcbb62ee9eaac039a60a6153ee29b99a3be01940a2b67522'
assert all(sha((W/p).read_bytes())==v['sha256'] for p,v in s['files'].items())
assert sha(subprocess.check_output(['git','diff','--binary','HEAD','--'],cwd=W))==s['diffSha256']
a=m['actual'];assert [a[k] for k in ['diskBytes','normalBytes','nodes','coverageBytes','assertions']]==[790555,802361,20956,117702,2503]
assert a['newDiskBytes']<=80000 and a['newNormalBytes']<=80000 and a['newNodes']<=2200 and a['newCoverageBytes']<=15000 and a['newAssertions']<=260
assert m['maxDerivativeNormalBytes']==808479 and m['maxDerivativeNodes']==21112 and len(m['derivatives'])==41
# Deliberately only historical source-authoring fences; runtime files are outside exact changed paths.
replacements={
'packages/conformance/src/transactional-cumulative.test.ts':{'254599':'306371','742879':'794651','754685':'806457','38453':'48064','109115':'118726','6620':'8001','19639':'21020'},
'apps/bdptest/src/transactional-matrix.test.ts':{'760803':'812575','19795':'21176'}}
changed={}
for p,rs in replacements.items():
 old=(W/p).read_text();new=old
 for before,after in rs.items():
  count=new.count(before);assert count==(2 if before in ['109115','760803','19795'] else 1),(before,count)
  new=new.replace(before,after)
 if 'cumulative' in p:
  anchor='  // These measured source authoring fences do not change runtime/parser limits.'
  assert new.count(anchor)==1
  new=new.replace(anchor,'  // Root-accepted sequence-member: disk790555/normal802361/nodes20956/coverage117702.\n'+anchor)
 (W/p).write_text(new);changed[p]={'before':sha(old.encode()),'after':sha(new.encode()),'replacements':rs}
patch=subprocess.check_output(['git','diff','--binary','HEAD','--'],cwd=W)
(D/'draft.patch').write_bytes(patch)
s['status']='FROZEN_MEASURED_FENCES_NOT_EXECUTED';s['diffSha256']=sha(patch);s['preFenceBindingSha256']=sha((SA/'source-binding.json').read_bytes())
for p in s['files']:s['files'][p]['sha256']=sha((W/p).read_bytes())
write(D/'source-binding.json',s)
accept={'status':'accepted-measured-source-fences','at':datetime.datetime.now(datetime.timezone.utc).isoformat(),'sourceBindingSha256':sha((D/'source-binding.json').read_bytes()),'measurement':artifact(SA/'measurements.json'),'preFenceSource':artifact(SA/'source-binding.json'),'runtimeLimitsUnchanged':True,'simultaneousGrowthStopsSatisfied':True,'margins':{'bytes':4096,'nodes':64,'coverageBytes':1024},'changes':changed,'reason':'One appended ten-action partial witness; 28 raw parent objects retained, all41 derivatives measured, original runtime limits and execution bounds unchanged. Exact source fences retain previous margins, not arbitrary growth allowance.'}
write(D/'root-fence-acceptance.json',accept)
print(json.dumps({'source':artifact(D/'source-binding.json'),'acceptance':artifact(D/'root-fence-acceptance.json'),'patch':artifact(D/'draft.patch')}))
