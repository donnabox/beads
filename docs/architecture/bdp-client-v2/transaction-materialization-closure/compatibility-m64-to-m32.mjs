import assert from 'node:assert/strict';
import {mkdtempSync,rmSync,readFileSync} from 'node:fs';
import {createHash} from 'node:crypto';
import {tmpdir} from 'node:os';
import path from 'node:path';
import {recordProvenance,currentRoot} from './provenance.mjs';
const oldRoot='/Users/dbox/repos/wt/vickie/vickie-events-native-review';
const previous=await import(oldRoot+'/packages/server/dist/transaction-store.js');
const oldBatch=await import(oldRoot+'/packages/server/dist/transaction-batch.js');
const next=await import(currentRoot+'/packages/server/dist/transaction-store.js');
const newBatch=await import(currentRoot+'/packages/server/dist/transaction-batch.js');
recordProvenance(oldRoot,oldBatch.TRANSACTION_RUNTIME_LIMITS,64,import.meta.url);
recordProvenance(currentRoot,newBatch.TRANSACTION_RUNTIME_LIMITS,32,import.meta.url);
const dir=mkdtempSync(path.join(tmpdir(),'bdp-closure-compat-'));let owner;
const sha=b=>createHash('sha256').update(b).digest('hex');
try {
 const scope='https://example.test/compat/',type=scope+'types/bead';
 const options={file:path.join(dir,'owner.sqlite'),scope,types:[{id:type,name:'B',describes:'bead',conformsTo:[]}]};
 const prepared=oldBatch.prepareTransaction(scope,JSON.stringify({operations:[{operation:'createBead',id:'beads/a',type,properties:{n:0}}]}));
 owner=previous.openTransactionStore({...options,create:true});const admitted=owner.admit('p','key',prepared);assert.equal(admitted.kind,'owned');const receipt=owner.execute(admitted.claim);owner.close();owner=undefined;
 const before=readFileSync(options.file);assert.throws(()=>next.openTransactionStore({...options,create:false}),/installation mismatch/);const after=readFileSync(options.file);assert.deepEqual(after,before);
 owner=previous.openTransactionStore({...options,create:false});const replay=owner.admit('p','key',prepared);assert.equal(replay.kind,'existing');assert.equal(JSON.stringify(replay.receipt),JSON.stringify(receipt));
 console.log(JSON.stringify({stage:'verified',oldMatched:oldBatch.TRANSACTION_RUNTIME_LIMITS.matchedResources,newMatched:newBatch.TRANSACTION_RUNTIME_LIMITS.matchedResources,beforeDatabaseDigest:sha(before),afterRefusalDatabaseDigest:sha(after),receiptDigest:sha(JSON.stringify(receipt)),replayedReceiptDigest:sha(JSON.stringify(replay.receipt))}));
}finally{owner?.close();rmSync(dir,{recursive:true,force:true});}
