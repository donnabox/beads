import json, hashlib, copy
from pathlib import Path
L=Path('/Users/dbox/repos/wt/vickie/bdp-integration-logs/transaction-cross-carrier-draft')
W=Path('/Users/dbox/repos/wt/vickie/vickie-cross-carrier-draft')
sha=lambda b:hashlib.sha256(b).hexdigest()
normal=lambda v:json.dumps(v,indent=2,ensure_ascii=False)+'\n'
compact=lambda v:json.dumps(v,separators=(',',':'),ensure_ascii=False)
def nodes(v):
    return 1+sum(nodes(x) for x in (v.values() if isinstance(v,dict) else v if isinstance(v,list) else []))
originals=['measure.py','measurements.json','binding.json','measured-report.md','draft.patch']
before={n:sha((L/n).read_bytes()) for n in originals}
binding=json.loads((L/'binding.json').read_text())
for f in binding['sourceFiles']:
    assert sha((W/f['path']).read_bytes())==f['sha256'],f['path']
report=json.loads((L/'measurements.json').read_text())
assert len(report['derivatives'])==37
manifest=json.loads((W/'packages/conformance/test-support/transactional/manifest-v1.json').read_bytes())
changed=copy.deepcopy(manifest)
row=next(r for r in changed['scenarios'] if r['id']=='transactional.batch.atomic-commit')
action=next(a for a in row['actions'] if a['id']=='a-point')
action['captures']=[{'binding':'commit-receipt','from':{'kind':'json-pointer','pointer':'/id'}}]
n=normal(changed).encode(); c=compact(changed).encode()
d={'name':'duplicate-receipt-binding','normalBytes':len(n),'normalSHA256':sha(n),'compactBytes':len(c),'compactSHA256':sha(c),'nodes':nodes(changed),'actions':sum(len(r['actions']) for r in changed['scenarios']),'assertions':sum(len(a['assertions']) for r in changed['scenarios'] for a in r['actions']),'deltaBytes':len(n)-report['canonical']['normalBytes'],'deltaNodes':nodes(changed)-report['canonical']['nodes'],'expectedAdmission':'scenario manifest: validation failed before installed preparation','expectedServiceChildren':0,'expectedHTTPExchanges':0}
assert (d['normalBytes'],d['nodes'])==(695793,18112)
assert d['normalBytes']<=736794 and d['nodes']<=19129
(L/'derivative-duplicate-receipt-binding.json').write_bytes(n)
report['derivatives'].append(d)
report['unchangedCanonicalDerivatives']='14 other observe invocations do not change manifest bytes: metadata5, abort/preabort3, deniedall1, omittedroutes2, deniedset2, plus positive1. Duplicate-binding is a changed manifest rejected before preparation.'
report['mutatedManifestInvocations']=38
report['unchangedManifestInvocations']=14
report['inventoryCorrection']={'originalMeasurementsSHA256':before['measurements.json'],'originalReportSHA256':before['measured-report.md'],'addedDerivative':'duplicate-receipt-binding','originalMutatedCount':37,'correctedMutatedCount':38,'originalUnchangedCount':15,'correctedUnchangedCount':14,'sourceUnchanged':True,'noApplicationExecution':True}
assert len(report['derivatives'])==38 and 38+14==report['newObserveInvocations']==52
assert max(x['normalBytes'] for x in report['derivatives'])==report['maxDerivativeBytes']==701716
assert max(x['nodes'] for x in report['derivatives'])==report['maxDerivativeNodes']==18263
(L/'measurements-v2.json').write_text(normal(report))
assert before=={n:sha((L/n).read_bytes()) for n in originals}
for f in binding['sourceFiles']:
    assert sha((W/f['path']).read_bytes())==f['sha256'],f['path']
newhash=sha((L/'measurements-v2.json').read_bytes()); script_hash=sha((L/'measure-v2.py').read_bytes())
text=f'''# Supplemental measurement inventory v2

This corrects an inventory omission only. All frozen source files and original reports, measure.py, measurements.json, binding.json and draft.patch remain byte-identical. No guard or source changes were made. No tests, builds, installations, services or application code ran.

## Exact omitted transformation

apps/bdptest/src/transactional-matrix.test.ts:921-935 registers the duplicate receipt-binding control. Its manifest callback replaces atomic-commit/a-point captures with binding commit-receipt, json-pointer /id. This changes the manifest even though the artifact bundle rejects it. observe applies the callback and constructs/validates the cumulative bundle at lines119-127, before temporary installed directory creation at128 and any installed launch. The test expects scenario manifest: validation failed and zero children. Therefore the source model assigns zero service launches and zero HTTP requests to this invalid-bundle invocation; these are not newly executed observations.

The supplemental data-only script independently reproduced exactly that capture replacement. Normal stringify bytes:695793; compact execution bytes:{d['compactBytes']}; nodes:18112; actions:199; assertions:2102. Canonical overhead is195 normal bytes and5 nodes. Normal derivative SHA256:{d['normalSHA256']}; compact SHA256:{d['compactSHA256']}.

## Corrected complete inventory

38 mutated-manifest invocations +14 unchanged-manifest invocations =52 observe invocations. The unchanged14 are metadata5, abort/preabort3, deniedall1, omittedroutes2, deniedset2 and positive1. The original inventory's abort/preabort/duplicate4 group incorrectly counted this invalid manifest among unchanged bytes.

All other37 derivative entries are retained without modification. Joint maxima remain701716 normal stringify bytes and18263 nodes; canonical counts, coverage, route digest,59 expanded app cases and52 observe invocations remain unchanged. The additional derivative is below both736794-byte/19129-node hard stops. No service/HTTP cost is added to the existing source model because this was already one of its zero-launch invalid-bundle invocations.

## Original-to-corrected bindings

- Original measurements.json SHA256:{before['measurements.json']} (preserved).
- Corrected measurements-v2.json SHA256:{newhash}.
- Original measure.py SHA256:{before['measure.py']} (preserved).
- Supplemental measure-v2.py SHA256:{script_hash}.
- Original measured-report.md SHA256:{before['measured-report.md']} (preserved; this supplement supersedes its37/15 inventory statement only).
- Original binding.json SHA256:{before['binding.json']} (preserved).
- Original draft.patch SHA256:{before['draft.patch']} (preserved).

The original source binding still applies unchanged. Both pre/post hash checks passed for the original artifacts and all four source files. New derivative is derivative-duplicate-receipt-binding.json. Root still owns any guard amendment and execution release.

An initial supplemental-script authoring command had a quoting syntax error before writing any file; the immediate attempted script invocation found no file. Both were evidence orchestration failures, not application executions. The corrected script then produced this supplement once.
'''
(L/'measurement-inventory-v2.md').write_text(text)
for name in ['measurement-inventory-v2.md','measurements-v2.json','measure-v2.py','derivative-duplicate-receipt-binding.json']:
    print(name,sha((L/name).read_bytes()))
