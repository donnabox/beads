import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
import { DatabaseSync } from 'node:sqlite';
import path from 'node:path';
import { openTransactionStore } from '/Users/dbox/repos/wt/vickie/vickie-materialization-closure/packages/server/dist/transaction-store.js';
import { TRANSACTION_RUNTIME_LIMITS, prepareTransactionSetCore } from '/Users/dbox/repos/wt/vickie/vickie-materialization-closure/packages/server/dist/transaction-batch.js';
import { recordProvenance,currentRoot } from './provenance.mjs';
const mode='dense', scope='https://example.test/materialization-stress/', bead=scope+'types/bead',link=scope+'types/link';
const options={file:path.join(process.env.BDP_STRESS_DIRECTORY,'owner.sqlite'),scope,types:[{id:bead,name:'Owner',describes:'bead',conformsTo:[],ownsOutgoing:{[link]:{max:1}}},{id:link,name:'Link',describes:'link',conformsTo:[],source:{conformsTo:[bead],external:'none'},target:{conformsTo:[bead],external:'none'}}],receiptMaximumItems:32};
let owner, largestInput=0, serial=0;
const bytes=x=>Buffer.byteLength(JSON.stringify(x));
const digest=x=>createHash('sha256').update(JSON.stringify(x)).digest('hex');
const log=(stage,data={})=>console.log(JSON.stringify({stage,mode,maxRSSKiB:process.resourceUsage().maxRSS,...data}));
const prepare=text=>{largestInput=Math.max(largestInput,Buffer.byteLength(text));assert.ok(Buffer.byteLength(text)<=65536);return prepareTransactionSetCore(scope,text)};
const run=(ops,key='seed'+serial++,principal='seed')=>{const p=typeof ops==='string'?prepare(ops):prepare(JSON.stringify({operations:ops}));const a=owner.admit(principal,key,p);assert.equal(a.kind,'owned');const r=owner.execute(a.claim);assert.equal(r.status,'completed',JSON.stringify(r.problem));return r};
const point=id=>{const r=owner.withRead(reader=>JSON.parse(reader.resource(id).bodyJson));assert.equal(r.kind,'read');return r.value};
const meta=(id,name,value)=>({operation:'updateBead',bead:id,metadataChange:[{op:'add',path:'/'+name,value}]});
const provenance=recordProvenance(currentRoot,TRANSACTION_RUNTIME_LIMITS,32,import.meta.url);
try {
 owner=openTransactionStore({...options,create:true});log('open');
 const matchMaximum=TRANSACTION_RUNTIME_LIMITS.matchedResources;
 const count=mode==='noop'?matchMaximum/32:Math.min(matchMaximum,128);
 const sizes=[];
 for(let i=0;i<count;i++) {
  const id='beads/a'+i, lid='links/l'+i;
  run([{operation:'createBead',id,type:bead,properties:{n:0}}]);
  run([{operation:'createLink',id:lid,type:link,source:id,target:id,properties:{n:0,payload:'p'.repeat(60000)}}]);
  for(let j=0;j<3;j++)run([{operation:'updateLink',link:lid,metadataChange:[{op:'add',path:'/chunk'+j,value:'m'.repeat(60000)}]}]);
  // Source embeds the complete owned Link. Reserve only the exact source /last
  // empty-string wrapper, grow Link into all other available bytes, then fill
  // the source residual (which may be zero). No arbitrary 1KiB headroom.
  const oldLink=point(lid), oldSource=point(id);
  const linkFieldOverhead=bytes({...oldLink,metadata:{...oldLink.metadata,last:''}})-bytes(oldLink);
  const room=262144-bytes({...oldSource,metadata:{...oldSource.metadata,last:''}})-linkFieldOverhead;
  assert.ok(room>0&&room<60000);
  run([{operation:'updateLink',link:lid,metadataChange:[{op:'add',path:'/last',value:'z'.repeat(room)}]}]);
  const current=point(id), remaining=262144-bytes({...current,metadata:{...current.metadata,last:''}});
  assert.ok(remaining>=0&&remaining<60000);run([meta(id,'last','s'.repeat(remaining))]);
  const sourceBytes=bytes(point(id)), linkBytes=bytes(point(lid));assert.equal(sourceBytes,262144);
  sizes.push({sourceBytes,linkBytes});
  if(i%16===15||i===count-1)log('seed',{completed:i+1,total:count,sourceBytes,linkBytes});
 }
 const eventSource=scope+'links/l0?view=events';
 let beforeCursor;
 for (;;) {const p=owner.readEvents(eventSource,beforeCursor);assert.equal(p.kind,'page');beforeCursor=p.page.events.at(-1)?.id??beforeCursor;if(!p.page.next)break;}
 assert.ok(beforeCursor);
 const prefix='{"operations":[{"operation":"updateWhere","collection":"links","selector":"$[?@.properties.n == 0]","change":[{"op":"add","path":"/transient","value":[';
 const suffix=']},{"op":"remove","path":"/transient"},{"op":"replace","path":"/n","value":1}]}]}';
 const numbers=Math.floor((65536-Buffer.byteLength(prefix+suffix)+1)/5);
 let text=prefix+Array(numbers).fill('1e20').join(',')+suffix;
 const padding=65536-Buffer.byteLength(text), lastToken=text.lastIndexOf('1e20');
 text=text.slice(0,lastToken)+'1e'+'0'.repeat(padding)+'20'+text.slice(lastToken+4);
 assert.equal(Buffer.byteLength(text),65536);
 log('input',{requestBytes:Buffer.byteLength(text),numericValues:numbers,normalizedCombinedPatchBytes:bytes(JSON.parse(text).operations[0].change),sourceBytesRange:[Math.min(...sizes.map(s=>s.sourceBytes)),Math.max(...sizes.map(s=>s.sourceBytes))],linkBytesRange:[Math.min(...sizes.map(s=>s.linkBytes)),Math.max(...sizes.map(s=>s.linkBytes))]});
 log('execute-start');
 let receipt=run(text,'measure','measure');
 const identity=receipt.id, position=receipt.effectPosition, receiptBytes=bytes(receipt), receiptHash=digest(receipt), resultCount=receipt.results.length;
 assert.equal(resultCount,mode==='noop'?matchMaximum+32:count+1);
 assert.equal(Boolean(position),mode!=='noop');
 log('executed',{receiptBytes,resultCount,position,largestInput});
 receipt=undefined;
 const stateHash=digest(point('beads/a0')); assert.equal(bytes(point('beads/a0')),262144);
 if(mode!=='noop')assert.equal(point('links/l0').properties.n,1);
 let retry=owner.admit('measure','measure',prepare(text));assert.equal(retry.kind,'existing');assert.equal(digest(retry.receipt),receiptHash);retry=undefined;
 const observed=owner.readReceipt('measure',identity);assert.equal(observed.receipt.status,'completed');assert.ok(observed.receipt.next);const pageUrl=observed.receipt.next;
 let page=owner.readReceiptPage('measure',pageUrl);assert.equal(page.kind,'page');const pageHash=digest(page.page);page=undefined;
 const finite=owner.readEvents(eventSource,beforeCursor);assert.equal(finite.kind,'page');assert.equal(finite.page.events.length,1);assert.equal(finite.page.events[0].type,'updated');assert.equal(finite.page.next,null);const eventHash=digest(finite.page);log('finite-page',{eventBytes:bytes(finite.page)});
 log('replay-and-page');owner.close();owner=undefined;
 log('owner-closed-before-size-query');
 // Scalar-only query of the one relevant persisted group. No inspect, body
 // transfer, history parsing, writes, or new runtime serving accessor.
 const db=new DatabaseSync(options.file,{readOnly:true});
 const groupBytes=position?db.prepare('SELECT length(CAST(body AS BLOB)) AS bytes FROM groups WHERE position=?').get(Number(position.slice(4))).bytes:0;db.close();
 log('size-query',{groupBytes});
 owner=openTransactionStore({...options,create:false});log('reopened');
 retry=owner.admit('measure','measure',prepare(text));assert.equal(retry.kind,'existing');assert.equal(digest(retry.receipt),receiptHash);retry=undefined;
 assert.equal(digest(point('beads/a0')),stateHash);
 page=owner.readReceiptPage('measure',pageUrl);assert.equal(page.kind,'page');assert.equal(digest(page.page),pageHash);
 const eventReopen=owner.readEvents(eventSource,beforeCursor);assert.equal(eventReopen.kind,'page');assert.equal(digest(eventReopen.page),eventHash);
 log('verified',{receiptBytes,groupBytes,largestInput,resultCount});
} finally {owner?.close();}
