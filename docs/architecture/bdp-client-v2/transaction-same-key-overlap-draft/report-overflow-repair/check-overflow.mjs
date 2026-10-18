import {readFileSync,writeFileSync} from 'node:fs';
import {createHash} from 'node:crypto';
import vm from 'node:vm';
import assert from 'node:assert/strict';
const dir='/Users/dbox/repos/wt/vickie/bdp-integration-logs/transaction-same-key-overlap-draft/report-overflow-repair';
const file='/Users/dbox/repos/wt/vickie/vickie-same-key-dispatch/apps/bdptest/src/transactional-same-key-overlap-installed.test.ts';
const digest=s=>createHash('sha256').update(s).digest('hex');
function run(path) {
 const full=readFileSync(path,'utf8');
 const code='{'+full.slice(full.indexOf('    let serialized = JSON.stringify(observation());'),full.lastIndexOf('}, 60000);'));
 let calls=0; const logs=[];const failures=[];
 const context={Buffer,reportOverflow:false,reportWriteFailed:false,reportPath:undefined,mode:'positive',failures,addFailure:e=>failures.push(e),observation:()=>({payload:++calls===1?'small':'x'.repeat(65536)}),console:{log:s=>logs.push(JSON.parse(s))}};
 let error=null;try {vm.runInNewContext(code,context,{timeout:1000});} catch(e) {error=e.message;}
 return {sourceSha256:digest(full),extractedCodeSha256:digest(code),observationCalls:calls,reportOverflow:context.reportOverflow,failureCount:failures.length,error,logs};
}
const before=run(dir+'/test-before.ts.txt'),after=run(file);
assert.equal(before.observationCalls,2);assert.equal(before.error,null);assert.equal(before.failureCount,0);assert.equal(before.logs[0].outcome,'failed');
assert.equal(after.observationCalls,2);assert.equal(after.reportOverflow,true);assert.equal(after.failureCount,1);assert.equal(after.error,'same-key report ceiling');assert.equal(after.logs[0].failureCount,1);
const result={kind:'Focused extracted-source branch check with injected small-first/oversized-second observation; not an installed execution or demonstration that ordinary fixture reaches cap',before,after,passed:true};
writeFileSync(dir+'/overflow-check.json',JSON.stringify(result,null,2)+'\n',{flag:'wx'});console.log(JSON.stringify(result));
