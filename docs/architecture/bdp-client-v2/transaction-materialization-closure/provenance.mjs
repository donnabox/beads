import assert from 'node:assert/strict';
import {createHash} from 'node:crypto';
import {readFileSync,readdirSync} from 'node:fs';
import {execFileSync} from 'node:child_process';
import path from 'node:path';
import {fileURLToPath} from 'node:url';
const sha=b=>createHash('sha256').update(b).digest('hex');
export const currentRoot='/Users/dbox/repos/wt/vickie/vickie-materialization-closure';
export function recordProvenance(root,tuple,expectedMatches,scriptUrl){
 assert.equal(process.version,'v24.16.0');
 assert.equal(tuple.matchedResources,expectedMatches);
 assert.equal(tuple.inducedEvents,256);
 assert.equal(tuple.maximumOperations,32);
 assert.equal(tuple.representationBytes,262144);
 assert.equal(tuple.requestBodyBytes,65536);
 const emitted={};
 function walk(relative){for(const e of readdirSync(path.join(root,relative),{withFileTypes:true}).sort((a,b)=>a.name.localeCompare(b.name))){const f=path.posix.join(relative,e.name);if(e.isDirectory())walk(f);else if(e.isFile()&&/\.(js|json)$/.test(e.name))emitted[f]=sha(readFileSync(path.join(root,f)));}}
 walk('packages/server/dist');walk('packages/protocol/dist');
 const record={node:process.version,root,sourceCommit:execFileSync('git',['rev-parse','HEAD'],{cwd:root,encoding:'utf8'}).trim(),sourceStatus:execFileSync('git',['status','--short'],{cwd:root,encoding:'utf8'}).trim(),sourceBatchDigest:sha(readFileSync(path.join(root,'packages/server/src/transaction-batch.ts'))),sourceStoreDigest:sha(readFileSync(path.join(root,'packages/server/src/transaction-store.ts'))),emittedBatchDigest:emitted['packages/server/dist/transaction-batch.js'],emittedStoreDigest:emitted['packages/server/dist/transaction-store.js'],emittedTreeDigest:sha(JSON.stringify(emitted)),emittedFiles:emitted,tuple,tupleDigest:sha(JSON.stringify(tuple)),scriptPath:fileURLToPath(scriptUrl),scriptDigest:sha(readFileSync(fileURLToPath(scriptUrl))),provenanceHelperDigest:sha(readFileSync(fileURLToPath(import.meta.url)))};
 console.log(JSON.stringify({stage:'provenance',...record}));return record;
}
