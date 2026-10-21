"""Print actual final source binding only; root must inspect/save with session guard."""
import sys
sys.dont_write_bytecode=True
from gate import W,FILE,WITNESS,INTEGRATION,git,sha,require
import json
head=git('rev-parse','HEAD').decode().strip();git('merge-base','--is-ancestor',INTEGRATION,head)
require(sha((W/FILE).read_bytes())==WITNESS,'Witness changed')
require(git('diff','--name-only','HEAD').decode().splitlines()==[FILE] and not git('ls-files','--others','--exclude-standard').strip(),'Unexpected source scope')
files={p:sha((W/p).read_bytes()) for p in git('ls-files').decode().splitlines()}
require(all(sha(git('show','HEAD:'+p))==v for p,v in files.items() if p!=FILE),'Runtime source differs from HEAD')
print(json.dumps({'worktree':str(W),'head':head,'tree':git('rev-parse','HEAD^{tree}').decode().strip(),'parents':git('show','-s','--format=%P','HEAD').decode().split(),'requiredIntegration':INTEGRATION,'dirtyPaths':[FILE],'witnessFile':FILE,'witnessSha256':WITNESS,'binaryDiffSha256':sha(git('diff','--binary','HEAD')),'plainDiffSha256':sha(git('diff','HEAD')),'trackedSourceFiles':files},indent=2))
