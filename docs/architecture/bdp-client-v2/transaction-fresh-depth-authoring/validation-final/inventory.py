"""Data-only fresh combined build inventory, no product imports."""
from gate import *
g,spec=load_gate('build-inventory');x=completed('fresh-build');start=json.loads((E/'fresh-build.started.json').read_text())
require(start['head']==g['finalHead'] and start['workspaceDirty'] is False and start['buildInputFiles']==input_files(),'Fresh build provenance differs')
pin=lambda p:{'path':str(p),'sha256':sha(p.read_bytes())}
r={'head':g['finalHead'],'tree':g['finalTree'],'inputFiles':input_files(),'packageFiles':build_outputs(),'emittedFiles':emitted_files(),'buildStart':pin(E/'fresh-build.started.json'),'buildExit':pin(E/'fresh-build.exit.json'),'buildLog':pin(E/'fresh-build.log')}
with (E/'build-output-inventory.json').open('x') as f:f.write(json.dumps(r,indent=2)+'\n')
print(json.dumps({'inputs':len(r['inputFiles']),'packageMembers':len(r['packageFiles']),'emitted':len(r['emittedFiles'])}))
