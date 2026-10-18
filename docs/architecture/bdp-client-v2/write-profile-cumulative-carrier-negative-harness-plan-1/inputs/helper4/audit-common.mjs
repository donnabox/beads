/** Private validation file/control primitives; WIP, never evaluated during preparation. */
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";
import { createHash } from "node:crypto";
export const P = path.dirname(fileURLToPath(import.meta.url));
export const E = path.join(P, "run");
export const hash = (bytes) => createHash("sha256").update(bytes).digest("hex");
export function need(ok, code) { if (!ok) throw Object.assign(new Error("private validation refused"), { validationCode: code }); }
export function exact(value, keys) { need(value !== null && typeof value === "object" && !Array.isArray(value) && Object.keys(value).length === keys.length && keys.every((key) => Object.hasOwn(value, key)), "closed-keys"); }
export function same(a, b) { if (a === null || b === null || typeof a !== "object" || typeof b !== "object") return Object.is(a, b); if (Array.isArray(a) !== Array.isArray(b)) return false; const ak=Object.keys(a),bk=Object.keys(b);return ak.length===bk.length&&ak.every(k=>Object.hasOwn(b,k)&&same(a[k],b[k])); }
export function inside(root, value) { const relative=path.relative(root,value);return relative!==""&&!relative.startsWith(".."+path.sep)&&relative!==".."&&!path.isAbsolute(relative); }
export function raw(filename, maximum) {
 const before=fs.lstatSync(filename,{bigint:true});need(before.isFile()&&!before.isSymbolicLink()&&before.size<=BigInt(maximum),"ordinary-bounded-file");
 const fd=fs.openSync(filename,fs.constants.O_RDONLY|fs.constants.O_NOFOLLOW);const chunks=[];let total=0;
 const signature=s=>[s.dev,s.ino,s.mode,s.uid,s.size,s.mtimeNs,s.ctimeNs].map(String).join(":");
 try {const opened=fs.fstatSync(fd,{bigint:true});need(signature(opened)===signature(before),"file-open-identity");
  for (;;) {const buffer=Buffer.allocUnsafe(65536);const bytes=fs.readSync(fd,buffer,0,buffer.length,null);if(!bytes)break;total+=bytes;need(total<=maximum,"file-capacity");chunks.push(buffer.subarray(0,bytes));}
  need(signature(fs.fstatSync(fd,{bigint:true}))===signature(opened)&&signature(fs.lstatSync(filename,{bigint:true}))===signature(opened)&&BigInt(total)===opened.size,"file-read-identity");return Buffer.concat(chunks,total);
 } finally {fs.closeSync(fd);}
}
export function json(filename, maximum=4194304) { return JSON.parse(new TextDecoder("utf-8",{fatal:true}).decode(raw(filename,maximum))); }
export function pin(record, maximum=4194304) {exact(record,["path","sha256"]);need(path.isAbsolute(record.path)&&/^[a-f0-9]{64}$/.test(record.sha256),"document-pin");const bytes=raw(record.path,maximum);need(hash(bytes)===record.sha256,"document-pin-changed");return JSON.parse(new TextDecoder("utf-8",{fatal:true}).decode(bytes));}
export function output(filename, value) {need(path.dirname(filename)===E,"owned-output-path");const encoded=Buffer.from(JSON.stringify(value,null,2)+"\n");need(encoded.length<=4194304,"tool-output-capacity");const fd=fs.openSync(filename,fs.constants.O_WRONLY|fs.constants.O_CREAT|fs.constants.O_EXCL,0o600);try{fs.writeFileSync(fd,encoded);}finally{fs.closeSync(fd);}}
export function context(stage, controlName) {
 need(process.argv.length===4&&process.argv[2]==="--contract"&&process.argv[3]===path.join(P,controlName),"exact-tool-argv");
 const control=json(process.argv[3]);const gate=json(path.join(P,"execution-gate.json"));
 need(gate.released===true&&gate.runtimeValidationAuthorized===true&&gate.session==="2026-10-18T02:16:24Z-549f"&&gate.mode==="final"&&gate.approvedStages.includes(stage),"root-runtime-prefix");
 const startedPath=path.join(E,stage+".started.json"),startedBytes=raw(startedPath,1048576),started=JSON.parse(new TextDecoder("utf-8",{fatal:true}).decode(startedBytes));
 need(process.env.CARRIER_VALIDATION_STAGE===stage&&process.env.CARRIER_VALIDATION_ATTEMPT===hash(startedBytes)&&started.rootGateSha256===hash(raw(path.join(P,"execution-gate.json"),4194304)),"supervised-attempt");
 const binding=pin(gate.sourceBinding);need(binding.worktree===control.sourceRoot&&same(binding.trackedSourceFiles,json(path.join(P,"expected-source-map.json")))&&started.sourceBindingSha256===gate.sourceBinding.sha256,"exact-source-map");
 need(started.sourceIdentity==="78e11962df1a744c685c9911d13c0bff3fa42fcdec9310b75ec8f83a17996c7f","known-source-content");
 const implementation=pin({path:path.join(P,"implementation-hashes.json"),sha256:gate.implementationIndexSha256});
 for(const [relative,sha256]of Object.entries(implementation.files)){need(hash(raw(path.join(P,relative),67108864))===sha256,"frozen-helper-input");}
 const workspace=pin(gate.workspaceTools);need(workspace.sourceIdentity===started.sourceIdentity&&workspace.sourceLockSha256===binding.trackedSourceFiles["pnpm-lock.yaml"],"fresh-workspace-tools");
 const proof=pin(gate.installedProvenance);need(proof.sourceIdentity===started.sourceIdentity&&proof.sourceBindingSha256===gate.sourceBinding.sha256&&proof.sourceCommit===binding.head&&proof.sourceTree===binding.tree&&proof.consumerLockSha256===gate.consumerLockSha256,"fresh-product-proof");
 return {control,gate,binding,started,workspace,proof,sourceIdentity:started.sourceIdentity};
}
export function verifyFiles(records) {need(Array.isArray(records)&&records.length>0&&records.length<=1034,"proof-file-count");const names=new Set();for(const record of records){need(!names.has(record.path)&&path.isAbsolute(record.path)&&fs.realpathSync(record.path)===record.realpath,"proof-file-identity");names.add(record.path);const bytes=raw(record.path,67108864);need(bytes.length===record.bytes&&hash(bytes)===record.sha256,"proof-file-bytes");}}
export { fs, path, fileURLToPath, pathToFileURL };
