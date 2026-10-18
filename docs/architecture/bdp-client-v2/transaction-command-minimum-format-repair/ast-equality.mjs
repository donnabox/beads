import { readFileSync, writeFileSync } from 'node:fs';
import { createHash } from 'node:crypto';
import { parse } from '/Users/dbox/repos/wt/vickie/vickie-command-minimum/node_modules/@babel/parser/lib/index.js';
import { fileURLToPath } from 'node:url';
import path from 'node:path';
const E = path.dirname(fileURLToPath(import.meta.url));
const paths = JSON.parse(readFileSync(path.join(E,'paths.json'),'utf8'));
const omit = new Set(['start','end','loc','leadingComments','innerComments','trailingComments','comments','extra']);
function normalized(value) {
  if (Array.isArray(value)) return value.map(normalized);
  if (value && typeof value === 'object') return Object.fromEntries(Object.entries(value).filter(([k]) => !omit.has(k)).map(([k,v]) => [k,normalized(v)]));
  return value;
}
const hash = x => createHash('sha256').update(x).digest('hex');
const rows = paths.map(p => {
  const before = readFileSync(path.join(E,'input',p),'utf8');
  const after = readFileSync(path.join(E,'output',p),'utf8');
  const a = JSON.stringify(normalized(parse(before,{sourceType:'module',plugins:['typescript'],attachComment:false})));
  const b = JSON.stringify(normalized(parse(after,{sourceType:'module',plugins:['typescript'],attachComment:false})));
  return {path:p,beforeSha256:hash(before),afterSha256:hash(after),normalizedBeforeSha256:hash(a),normalizedAfterSha256:hash(b),normalizedBytes:Buffer.byteLength(a),equal:a===b};
});
const result = {node:process.version,parser:'@babel/parser8.0.4',mode:'parse only; no product import or evaluation',removedMetadata:[...omit],templateRawValuesPreserved:true,literalsAndOperatorsPreserved:true,rows,allEqual:rows.every(x=>x.equal)};
writeFileSync(path.join(E,'ast-equality.json'),JSON.stringify(result,null,2)+'\n');
console.log(JSON.stringify(result,null,2));
if(!result.allEqual) process.exitCode=1;
