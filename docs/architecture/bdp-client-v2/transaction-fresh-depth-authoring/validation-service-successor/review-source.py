"""Require actual emitted numeric identity measurements; never import product."""
from gate import *
g,spec=load_gate('review-source-measurements');stage='transaction-submission-positive';completed(stage)
log=E/(stage+'.log');observations=[]
for line in log.read_text().splitlines():
 if line.startswith('{'):
  try:value=json.loads(line)
  except json.JSONDecodeError:continue
  if value.get('witness')=='fresh-depth-source-identities':observations.append(value)
require(len(observations)==1,'Actual identity measurements absent/duplicated')
rows=observations[0]['measured'];expected=json.loads((P/'source-body-measurements.json').read_text())
labels=['single','seed','patch','noop','set','made','linked','linked-resolved','failed','failed','seed','failed']
require([r['label'] for r in rows]==labels,'Unexpected identity measurement order')
for index,r in enumerate(rows):
 key='sequence' if r['label'] in ['made','linked','linked-resolved'] else 'after-expiry' if index==11 else r['label']
 require(r['bodyBytes']==expected[key]['bytes'] and type(r['identityBytes']) is int and r['identityBytes']>0,'Unbound/invalid actual identity measurement')
record={'head':g['finalHead'],'sourceHashes':source_hashes(),'logSha256':sha(log.read_bytes()),'measurements':rows,'limitation':'Actual runtime identity sizes, not source-precomputed storage bytes or a new identity cap.'}
with (E/'source-identity-measurements.json').open('x') as f:f.write(json.dumps(record,indent=2)+'\n')
print(json.dumps({'accepted':True,'measurements':len(rows)}))
