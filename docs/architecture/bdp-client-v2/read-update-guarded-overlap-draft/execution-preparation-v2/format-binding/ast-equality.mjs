import { readFileSync, writeFileSync } from "node:fs";
import { createHash } from "node:crypto";
import { parse } from "/Users/dbox/repos/wt/vickie/vickie-sequence-member-evidence/node_modules/.pnpm/@babel+parser@8.0.4/node_modules/@babel/parser/lib/index.js";
const before = "/Users/dbox/repos/wt/vickie/bdp-integration-logs/read-update-guarded-overlap-draft/review-successor/read-update-guarded-overlap-installed.test.ts";
const after = "/Users/dbox/repos/wt/vickie/vickie-sequence-member-evidence/apps/bdptest/src/read-update-guarded-overlap-installed.test.ts";
const output = "/Users/dbox/repos/wt/vickie/bdp-integration-logs/read-update-guarded-overlap-draft/execution-preparation-v2/format-binding/ast-equality.json";
const omit = new Set(["start", "end", "loc", "leadingComments", "innerComments", "trailingComments", "comments", "extra"]);
function normalize(v) { if (Array.isArray(v)) return v.map(normalize); if (v && typeof v === "object") return Object.fromEntries(Object.entries(v).filter(([k]) => !omit.has(k)).map(([k,x]) => [k, normalize(x)])); return v; }
const sha = v => createHash("sha256").update(v).digest("hex");
const a = readFileSync(before,"utf8"), b = readFileSync(after,"utf8");
const ap = parse(a,{sourceType:"module",plugins:["typescript"],attachComment:false}), bp = parse(b,{sourceType:"module",plugins:["typescript"],attachComment:false});
const an = JSON.stringify(normalize(ap)), bn = JSON.stringify(normalize(bp));
const commentsA = ap.comments.map(x => ({type:x.type,value:x.value})), commentsB = bp.comments.map(x => ({type:x.type,value:x.value}));
const result = {mode:"tooling parser only; no product imports/evaluation",node:process.version,parserVersion:"8.0.4",parserSha256:sha(readFileSync("/Users/dbox/repos/wt/vickie/vickie-sequence-member-evidence/node_modules/.pnpm/@babel+parser@8.0.4/node_modules/@babel/parser/lib/index.js")),before,after,beforeSha256:sha(a),afterSha256:sha(b),normalizedBeforeSha256:sha(an),normalizedAfterSha256:sha(bn),normalizedBytes:Buffer.byteLength(an),equal:an===bn,commentsEqual:JSON.stringify(commentsA)===JSON.stringify(commentsB),ignoredMetadata:[...omit],templateRawAndLiteralValuesRetained:true};
writeFileSync(output,JSON.stringify(result,null,2)+"\n",{flag:"wx"}); console.log(JSON.stringify(result)); if(!result.equal || !result.commentsEqual) process.exitCode=1;
