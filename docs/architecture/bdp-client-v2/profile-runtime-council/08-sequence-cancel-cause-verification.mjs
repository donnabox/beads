import { mkdtempSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { DatabaseSync } from 'node:sqlite';
import { openTransactionStore } from '/Users/dbox/repos/wt/vickie/bdp-transaction-dependencies-integration-review/packages/server/dist/transaction-store.js';
import { prepareTransactionSequence } from '/Users/dbox/repos/wt/vickie/bdp-transaction-dependencies-integration-review/packages/server/dist/transaction-batch.js';
import { parseTypeDescriptor } from '/Users/dbox/repos/wt/vickie/bdp-transaction-dependencies-integration-review/packages/protocol/dist/index.js';
const dir=mkdtempSync(path.join(tmpdir(),'bdp-council-cancel-'));
const scope='https://example.test/transaction/', type='https://example.test/types/bead';
let armed=false, begins=0;
const actual=DatabaseSync.prototype.exec;
const store=openTransactionStore({file:path.join(dir,'test.sqlite'),scope,create:true,types:[parseTypeDescriptor({id:type,name:'Bead',describes:'bead',conformsTo:[]})],fault(stage){if(armed&&stage==='after-effects'){armed=false;begins=1;throw new Error('original sequence execution failure');}}});
const sequence=prepareTransactionSequence(scope,JSON.stringify({operations:[{operation:'createBead',type,idempotencyKey:'first'},{operation:'createBead',type,idempotencyKey:'second'}]}));
const attempt=store.admitSequence('principal',sequence);
DatabaseSync.prototype.exec=function(sql){if(begins&&sql==='BEGIN IMMEDIATE'&&++begins===3)throw new Error('sequence cancellation storage failure');return actual.call(this,sql);};
function flatten(e,seen=new Set()){if(!(e instanceof Error))return [];if(seen.has(e))throw Error('cycle');seen.add(e);return [e.message,...flatten(e.cause,new Set(seen)),...(e instanceof AggregateError?e.errors.flatMap(x=>flatten(x,new Set(seen))):[])];}
try {armed=true;try{store.nextSequence(attempt);}catch(e){console.log('initial',JSON.stringify(flatten(e)));}try{store.inspect();}catch(e){console.log('later',JSON.stringify(flatten(e)));}}
finally{DatabaseSync.prototype.exec=actual;store.close();rmSync(dir,{recursive:true,force:true});}
