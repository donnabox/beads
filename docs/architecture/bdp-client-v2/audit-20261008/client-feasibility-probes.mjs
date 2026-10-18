/** Audit evidence only. No HTTP endpoint, DB, production files or full BDP engine. */
import assert from 'node:assert/strict';
import { readFileSync, writeFileSync } from 'node:fs';
import { createRequire } from 'node:module';
import { selectReadResources } from './bdp/packages/server/src/read-selector.ts';
import { isJsonSchemaUri, isJsonSchemaDateTime } from './bdp/packages/protocol/src/schema-formats.ts';
const require = createRequire(new URL('./bdp/package.json', import.meta.url));
const { Ajv2020 } = require('ajv/dist/2020.js');
const schema = JSON.parse(readFileSync(new URL('./bdp/schemas/bdp-v0.schema.json', import.meta.url)));
const ajv = new Ajv2020({ strict: true, allErrors: true });
ajv.addFormat('uri',{type:'string',validate:isJsonSchemaUri});
ajv.addFormat('date-time',{type:'string',validate:isJsonSchemaDateTime});
ajv.addSchema(schema);
const batch = ajv.compile({$ref: schema.$id+'#/$defs/batchRequest'});
const limits={bytes:16384,depth:256,nodes:2048};
const scope='https://beads.example/a/';
const bead=(id,properties)=>({id:scope+'beads/'+id,type:scope+'types/issue',revision:'r1',properties});
const choose=(s,x)=>selectReadResources(s,limits,x);
let passed=0;
function test(name,fn){ fn(); ++passed; console.log('PASS '+name); }
const recipes={
 createWithLink:{operations:[{operation:'createBead',name:'new',type:scope+'types/issue',properties:{title:'Task',status:'open'}},{operation:'createLink',type:scope+'types/parent',source:'@new',target:'beads/parent',properties:{}}]},
 guardedClaim:{operations:[{operation:'updateBeadProperties',bead:'beads/a',expectedRevision:'r1',change:[{op:'replace',path:'/status',value:'in_progress'},{op:'add',path:'/assignee',value:'worker-a'}]}]},
 zeroMatchThenClose:{operations:[{operation:'deleteWhere',collection:'links',selector:'$[?@.target == "'+scope+'beads/p"]',cardinality:{max:0}},{operation:'updateBeadProperties',bead:'beads/p',expectedRevision:'r1',change:[{op:'replace',path:'/status',value:'closed'}]}]},
 emptyEffectsGuard:{operations:[{operation:'updateBeadProperties',bead:'beads/blocker',expectedRevision:'r1',change:[{op:'replace',path:'/status',value:'closed'}]}]}
};
test('four generic batch recipes conform to pinned wire schema (not execution)',()=>{for(const [name,r] of Object.entries(recipes))assert(batch(r),name+JSON.stringify(batch.errors));});
test('schema rejects a batch read/result-binding operation',()=>assert.equal(batch({operations:[{operation:'query',selector:'$[?@.properties.status == "open"]',name:'next'}]}),false));
test('schema rejects partial-success batch option',()=>assert.equal(batch({...recipes.guardedClaim,continueOnError:true}),false));
test('actual selector supports scalar Issue filters and exact map label membership',()=>{
 const rows=[bead('a',{status:'open',priority:1,labels:{bug:true}}),bead('b',{status:'closed',priority:0,labels:{bug:true}})];
 assert.equal(choose('$[?@.properties.status == "open" && @.properties.priority <= 1 && @.properties.labels.bug == true]',rows)[0].id,rows[0].id);
});
test('actual selector rejects contains search and array wildcard membership',()=>{
 for(const s of ['$[?contains(@.properties.title, "needle")]','$[?@.properties.labels[*] == "bug"]']) assert.throws(()=>choose(s,[]));
});
const count=50000, pageSize=100, text='x'.repeat(1024);
const rows=Array.from({length:count},(_,i)=>bead(String(i).padStart(6,'0'),{status:'open',priority:2,title:text}));
rows.at(-1).properties.title+='needle'; rows.at(-1).properties.priority=0;
let cost;
test('actual selector retains URI order; first page misses best-priority and only substring match',()=>{
 const candidates=choose('$[?@.properties.status == "open"]',rows);
 assert.equal(candidates.length,count);
 assert.equal(candidates.slice(0,pageSize).some(r=>r.properties.priority===0),false);
 assert.equal(candidates.slice(0,pageSize).some(r=>r.properties.title.includes('needle')),false);
 assert.equal(candidates.filter(r=>r.properties.title.includes('needle')).length,1);
 cost={records:count,pageSize,pages:Math.ceil(count/pageSize),resourceJsonBytes:Buffer.byteLength(JSON.stringify(candidates)),matches:1,interpretation:'Measured bytes for coarse-filter/full-scan fallback on synthetic corpus. Not HTTP traffic, throughput, universal lower bound, or a latency benchmark.'};
});
// Minimal explicit model: revisions stand in for a serialized expectedRevision CAS.
function cas(state,id,revision,patch){const r=state.get(id);if(r.revision!==revision)return false;r.properties={...r.properties,...patch};r.revision+='x';return true;}
test('CAS model gives exactly one winner for eight stale claim attempts',()=>{
 const state=new Map([['a',bead('a',{status:'open'})]]);
 assert.equal(Array.from({length:8},(_,i)=>cas(state,'a','r1',{status:'in_progress',assignee:String(i)})).filter(Boolean).length,1);
});
test('CAS model detects lost-append race and permits reread/recompute retry',()=>{
 const state=new Map([['a',bead('a',{notes:''})]]);assert(cas(state,'a','r1',{notes:'A'}));assert.equal(cas(state,'a','r1',{notes:'B'}),false);
 const r=state.get('a');assert(cas(state,'a',r.revision,{notes:r.properties.notes+'B'}));assert.equal(r.properties.notes,'AB');
});
function link(id){return {id:scope+'links/'+id,type:scope+'types/parent',revision:'l1',source:scope+'beads/child',target:scope+'beads/p',properties:{}};}
test('model counterexample: incoming unowned edge leaves target revision unchanged, allowing stale close',()=>{
 const state=new Map([['p',bead('p',{status:'open'})]]);const seen=state.get('p').revision;const links=[link('late')];
 assert.equal(choose(recipes.zeroMatchThenClose.operations[0].selector,links).length,1);
 assert(cas(state,'p',seen,{status:'closed'}));
});
test('zero-match predicate model rejects phantom before staging close',()=>{
 const selector=recipes.zeroMatchThenClose.operations[0].selector;
 assert.equal(choose(selector,[]).length<=0,true);assert.equal(choose(selector,[link('late')]).length<=0,false);
});
test('equal cardinality cannot detect membership replacement; explicit exclusion can',()=>{
 const old=link('old'), replacement=link('replacement');assert.equal([old].length,[replacement].length);
 const s='$[?@.target == "'+scope+'beads/p" && @.id != "'+old.id+'"]';
 assert.equal(choose(s,[old]).length,0);assert.equal(choose(s,[replacement]).length,1);
});
test('large enumerated guard hits real negotiated selector byte limit',()=>{
 const clauses=Array.from({length:1000},(_,i)=>'@.id != "'+scope+'links/'+i+'"');
 assert.throws(()=>choose('$[?'+clauses.join(' && ')+']',[]),e=>e.code==='source-bytes-limit-exceeded');
});
writeFileSync(new URL('./batch-recipes.json',import.meta.url),JSON.stringify(recipes,null,2)+'\n');
writeFileSync(new URL('./client-query-cost.json',import.meta.url),JSON.stringify(cost,null,2)+'\n');
console.log(JSON.stringify(cost));
console.log(`${passed} probes passed. Parser/schema checks and small explicit models only; no Transactional provider behavior established.`);
