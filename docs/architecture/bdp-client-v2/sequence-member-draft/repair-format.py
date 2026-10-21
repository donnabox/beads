from pathlib import Path
import json,hashlib,subprocess,shutil
D=Path(__file__).resolve().parent;W=Path('/Users/dbox/repos/wt/vickie/vickie-sequence-member-evidence');sha=lambda b:hashlib.sha256(b).hexdigest();F=D/'format-repair';F.mkdir(exist_ok=False)
for n in ['source-binding.json','draft.patch','root-fence-acceptance.json','root-source-release.json']:shutil.copyfile(D/n,F/('before-'+n))
p='packages/conformance/test-support/transactional/manifest-v1.json';old=(D/'source-author/source'/p).read_bytes();new=(W/p).read_bytes();assert json.loads(old)==json.loads(new);i=old.index(b'"id": "transactional.sequence.member-transactions"');assert old[:i]==new[:i];assert len(new)==790009
cp=W/'packages/conformance/src/transactional-cumulative.test.ts';t=cp.read_text()
for a,b in [('disk790555/normal','disk790009/normal'),('306371','305825'),('794651','794105')]:assert t.count(a)==1;t=t.replace(a,b)
cp.write_text(t)
s=json.loads((D/'source-binding.json').read_text());s['preFormatBindingSha256']=sha((F/'before-source-binding.json').read_bytes());s['status']='FROZEN_FORMAT_REPAIR_NOT_EXECUTED'
patch=subprocess.check_output(['git','diff','--binary','HEAD','--'],cwd=W);s['diffSha256']=sha(patch)
for p in s['files']:s['files'][p]['sha256']=sha((W/p).read_bytes())
(D/'source-binding.json').write_text(json.dumps(s,indent=2)+'\n');(D/'draft.patch').write_bytes(patch)
m=json.loads((D/'source-author/measurements.json').read_text());m['status']='FORMAT_ONLY_REMEASUREMENT_WITH_UNCHANGED_NORMAL_AND_DERIVATIVES';m['actual']['diskBytes']=790009;m['actual']['newDiskBytes']=51385;m['proposedFencesOnly']['manifestDisk']=794105;m['proposedFencesOnly']['manifestDiskDelta']=305825
(F/'measurements.json').write_text(json.dumps(m,indent=2)+'\n')
a=json.loads((F/'before-root-fence-acceptance.json').read_text());a.update(sourceBindingSha256=sha((D/'source-binding.json').read_bytes()),measurement={'path':str(F/'measurements.json'),'sha256':sha((F/'measurements.json').read_bytes())},formatOnlyRepair={'oldDisk':790555,'newDisk':790009,'semanticEquality':True,'parentPrefixUnchanged':True,'normalNodesCoverageDerivativesUnchanged':True,'diskMarginStill4096':True,'beforeAcceptanceSha256':sha((F/'before-root-fence-acceptance.json').read_bytes())})
(D/'root-fence-acceptance.json').write_text(json.dumps(a,indent=2)+'\n')
E=D/'execution-preparation-v2';E.mkdir(exist_ok=False);oldE=D/'execution-preparation'
c=(oldE/'commands.json').read_text().replace(str(oldE),str(E));(E/'commands.json').write_text(c)
r=(oldE/'run-source.py').read_text().replace('931723b766fe85a3dbbb7d4af22c261df6b0056a36c48d8240e6606a35433267',sha(c.encode()));(E/'run-source.py').write_text(r)
g=json.loads((oldE/'execution-gate.json').read_text());release=json.loads((F/'before-root-source-release.json').read_text());release.update(sourceBindingSha256=sha((D/'source-binding.json').read_bytes()),measuredFenceAcceptanceSha256=sha((D/'root-fence-acceptance.json').read_bytes()),repair='Only formatting within new manifest row, disk-fence decrease/comment. Original format failure preserved; fresh stages.')
(D/'root-source-release.json').write_text(json.dumps(release,indent=2)+'\n')
g.update(sourceBindingSha256=release['sourceBindingSha256'],supervisorSha256=sha(r.encode()),commandsSha256=sha(c.encode()),rootAcceptance={'path':str(D/'root-source-release.json'),'sha256':sha((D/'root-source-release.json').read_bytes())},measuredFenceAcceptance={'path':str(D/'root-fence-acceptance.json'),'sha256':sha((D/'root-fence-acceptance.json').read_bytes())});(E/'execution-gate.json').write_text(json.dumps(g,indent=2)+'\n')
print(json.dumps({'source':sha((D/'source-binding.json').read_bytes()),'patch':sha(patch),'measurement':sha((F/'measurements.json').read_bytes())}))
