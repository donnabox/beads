import assert from 'node:assert/strict';
import { constants } from 'node:buffer';
import { DatabaseSync } from 'node:sqlite';
import { decodeJsonDocument, admitJsonNumbers, stringifyJsonValue } from '/Users/dbox/repos/wt/vickie/vickie-transaction-owned/packages/protocol/dist/index.js';
const R=262144,Q=65536,N=7*Q,A=8192,H=1048576,O=32;
const db=new DatabaseSync(':memory:');
console.log(JSON.stringify({node:process.version,maxStringLength:constants.MAX_STRING_LENGTH,sqliteCompileOptions:db.prepare('PRAGMA compile_options').all().filter(x=>x.compile_options.startsWith('MAX_LENGTH='))}));db.close();
for(const S of [Q,4096]) for(const [M,E] of [[32,4096],[64,256],[128,256],[256,128],[256,256],[256,384],[256,512]]) {
 const K=M+O,perInvocation=Math.max(9*R+4*S,5*R+2*N+2*S),perEvent=2.5*R+N+S;
 const eventBound=E*perEvent+K*A+H,invocationBound=K*(perInvocation+A)+H;
 const receipt=K*(2*R+Q+A)+O*128+S+H;
 console.log(JSON.stringify({S,M,E,K,group:Math.min(eventBound,invocationBound),eventBound,invocationBound,receipt,groupFits:Math.min(eventBound,invocationBound)<constants.MAX_STRING_LENGTH,receiptFits:receipt<constants.MAX_STRING_LENGTH}));
 for(const [label,e,g] of [['bead lifecycle',1,2*R+S],['unowned update',1,2*R+N+S],['owned update',2,5*R+2*N+2*S],...[0,1,2].map(t=>['unowned Link lifecycle '+t,1+t,(2+2*t)*R+(1+t)*S]),...[1,2].map(t=>['owned Link lifecycle '+t,2+t,(5+2*t)*R+(2+t)*S])]) assert.ok(g<=e*perEvent,label);
}
for(const input of ['1e20','9e9','-0.0000010000000000000002','"\\ud800"','"\\udfff"','"\\ud83d\\ude00"']) {
 if (input==='"\\ud800"'||input==='"\\udfff"') {assert.throws(()=>decodeJsonDocument(input),/unpaired surrogate/);console.log(JSON.stringify({input,rejected:'unpaired surrogate'}));continue;}
 const result=admitJsonNumbers(decodeJsonDocument(input),{diagnostics:16,diagnosticBytes:524288,diagnostic:()=>({message:'unsupported'})});assert.ok(result.ok);
 const output=stringifyJsonValue(result.value),inputBytes=Buffer.byteLength(input),outputBytes=Buffer.byteLength(output);
 assert.ok(outputBytes<=7*inputBytes);console.log(JSON.stringify({input,output,inputBytes,outputBytes,ratio:outputBytes/inputBytes}));
}
