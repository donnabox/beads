/** Resolve the actual emitted/private runtime graph using the existing locked parser.
 * WIP: no helper/tool/product module has been evaluated during preparation.
 */
import { createRequire, builtinModules } from "node:module";
import {P,E,fs,path,pathToFileURL,fileURLToPath,need,same,raw,json,pin,hash,output,context,verifyFiles,inside} from "./audit-common.mjs";
async function main(){
 const c=context("prove-private-graph","private-emit-control.json"),control=c.control,root=control.emitRoot;
 const inventory=json(control.emitInventoryPath);need(inventory.sourceIdentity===c.sourceIdentity&&inventory.sourceBindingSha256===c.gate.sourceBinding.sha256&&inventory.root===root,"emit-source");verifyFiles(inventory.files);
 need(fs.lstatSync(path.join(root,"node_modules")).isSymbolicLink()&&fs.realpathSync(path.join(root,"node_modules"))===fs.realpathSync(path.join(control.sourceRoot,"node_modules")),"same-workspace-module-link");
 const parserPath=path.join(control.sourceRoot,control.parserToolRelative),parserBinding=c.workspace.tools[parserPath];need(parserBinding&&fs.realpathSync(parserPath)===parserBinding.realpath&&hash(raw(parserBinding.realpath,16777216))===parserBinding.sha256,"fresh-locked-parser-bytes");
 const parserManifestPath=path.join(control.sourceRoot,"node_modules/@babel/parser/package.json"),parserManifestBinding=c.workspace.tools[parserManifestPath];need(parserManifestBinding&&hash(raw(parserManifestPath,1048576))===parserManifestBinding.sha256&&json(parserManifestPath).version===control.parserVersion,"locked-parser-version");
 const {parse}=await import(pathToFileURL(parserBinding.realpath).href);
 const compilerFiles=new Map(inventory.files.map(f=>[f.realpath,f]));const files=new Map(compilerFiles);const packages=new Map();const installed=new Map(c.proof.installedProductionGraph.map(x=>[x.name,x]));const edges=[];const seen=new Set();const pending=control.rootModules.map(n=>path.join(root,n));let graphBytes=0;
 const builtins=new Set(builtinModules.map(n=>n.replace(/^node:/,"")));
 const isBuiltin=n=>builtins.has(n.replace(/^node:/,""));
 function packageRoot(filename){let dir=path.dirname(filename);for(;;){const meta=path.join(dir,"package.json");if(fs.existsSync(meta))return {root:dir,manifest:meta,data:json(meta)};const parent=path.dirname(dir);need(parent!==dir,"package-owner-absent");dir=parent;}}
 function bind(filename){
  const actual=fs.realpathSync(filename);need(path.isAbsolute(actual),"absolute-runtime-file");
  if(compilerFiles.has(actual))return actual;
  const pkg=packageRoot(actual),item=installed.get(pkg.data.name);need(item&&pkg.data.name!=="@bdp/server"&&pkg.data.version===item.version,"fixed-private-runtime-package");
  const consumer=path.dirname(item.manifest),relative=path.relative(pkg.root,actual);need(inside(pkg.root,actual)&&!relative.startsWith("..")&&fs.realpathSync(consumer).startsWith(fs.realpathSync(c.proof.consumer)+path.sep),"runtime-package-root");
  const bytes=raw(actual,16777216),consumerFile=path.join(consumer,relative);need(hash(bytes)===hash(raw(consumerFile,16777216)),"workspace-runtime-vs-proved-consumer-bytes");
  const workspaceManifest=raw(pkg.manifest,1048576);need(hash(workspaceManifest)===hash(raw(item.manifest,1048576)),"runtime-package-metadata");
  for(const [file,data]of [[actual,bytes],[fs.realpathSync(pkg.manifest),workspaceManifest]])files.set(file,{path:file,realpath:file,sha256:hash(data),bytes:data.length});
  need(!packages.has(pkg.data.name)||packages.get(pkg.data.name).workspaceRoot===pkg.root,"duplicate-private-package-instance");packages.set(pkg.data.name,{name:pkg.data.name,version:pkg.data.version,workspaceRoot:pkg.root,consumerRoot:consumer,manifestSha256:hash(workspaceManifest)});need(packages.size<=control.graphPolicies.runtimePackageInstancesMax,"runtime-package-capacity");return actual;
 }
 function esmBare(specifier,target){
  const pkg=packageRoot(target),parts=specifier.split("/"),name=specifier.startsWith("@")?parts.slice(0,2).join("/"):parts[0],tail=parts.slice(specifier.startsWith("@")?2:1).join("/");need(pkg.data.name===name,"no-runtime-package-alias");let mapped;
  if(pkg.data.exports!==undefined){let chosen=pkg.data.exports;
   if(typeof chosen==="object"&&chosen!==null&&Object.keys(chosen).some(k=>k.startsWith(".")))chosen=chosen[tail?"./"+tail:"."];
   if(chosen&&typeof chosen==="object"){need(Object.keys(chosen).every(k=>k==="types"||k==="default")&&typeof chosen.default==="string","no-unproved-conditional-exports");chosen=chosen.default;}
   need(typeof chosen==="string"&&chosen.startsWith("./"),"exact-runtime-export");mapped=path.resolve(pkg.root,chosen);
  }else if(tail){need(/\.(?:[cm]?js|json)$/.test(tail),"esm-subpath-no-extension-fallback");mapped=path.resolve(pkg.root,tail);}
  else{need(typeof pkg.data.main==="string"&&/\.(?:[cm]?js|json)$/.test(pkg.data.main),"exact-legacy-main");mapped=path.resolve(pkg.root,pkg.data.main);}
  need(fs.realpathSync(mapped)===target,"same-node-import-and-require-target");
 }
 function resolve(specifier,from,mode){
  need(typeof specifier==="string"&&specifier.length>0&&Buffer.byteLength(specifier)<=8192,"literal-bounded-module-specifier");if(isBuiltin(specifier))return {builtin:specifier};
  let target;
  if(mode==="import"&&(specifier.startsWith(".")||specifier.startsWith("/"))){const u=new URL(specifier,pathToFileURL(from));need(u.protocol==="file:"&&!u.search&&!u.hash&&/\.(?:[cm]?js|json)$/.test(u.pathname),"exact-file-module-url");target=fs.realpathSync(fileURLToPath(u));}
  else{need(!specifier.startsWith("#")&&!specifier.includes(":")&&!specifier.includes("?"),"no-internal-or-custom-module-resolver");target=fs.realpathSync(createRequire(pathToFileURL(from)).resolve(specifier));if(mode==="import")esmBare(specifier,target);}
  return {path:bind(target)};
 }
 while(pending.length){
  const file=bind(pending.shift());if(seen.has(file))continue;seen.add(file);need(seen.size<=control.graphPolicies.runtimeFilesMax,"runtime-graph-file-capacity");const bytes=raw(file,control.graphPolicies.runtimeCodeFileBytesMax);graphBytes+=bytes.length;need(graphBytes<=control.graphPolicies.runtimeBytesMax,"runtime-graph-byte-capacity");
  if(file.endsWith(".json"))continue;
  need(/\.[cm]?js$/.test(file),"runtime-file-format");const tree=parse(new TextDecoder("utf-8",{fatal:true}).decode(bytes),{sourceType:"unambiguous",allowReturnOutsideFunction:true,attachComment:false,errorRecovery:false,createImportExpressions:true});
  const stack=[{node:tree,parent:null,key:null}];let nodes=0;
  while(stack.length){const {node,parent,key}=stack.pop();if(!node||typeof node!=="object")continue;if(Array.isArray(node)){for(const child of node)stack.push({node:child,parent,key});continue;}if(typeof node.type!=="string")continue;need(++nodes<=control.graphPolicies.runtimeASTNodesPerFileMax,"runtime-ast-node-capacity");let value=null,mode=null;
   if((node.type==="ImportDeclaration"||node.type==="ExportNamedDeclaration"||node.type==="ExportAllDeclaration")&&node.source){value=node.source.value;mode="import";}
   if(node.type==="ImportExpression"){need(node.source.type==="StringLiteral","no-unresolved-dynamic-import");value=node.source.value;mode="import";}
   if(node.type==="CallExpression"){
    const callee=node.callee;const direct=callee.type==="Identifier"&&callee.name==="require",member=callee.type==="MemberExpression"&&!callee.computed&&callee.property.type==="Identifier"&&((callee.object.type==="Identifier"&&callee.object.name==="module"&&callee.property.name==="require")||(callee.object.type==="Identifier"&&callee.object.name==="require"&&callee.property.name==="resolve"));
    if(direct||member){need(node.arguments.length===1&&node.arguments[0].type==="StringLiteral","no-unresolved-require");value=node.arguments[0].value;mode="require";}
   }
   if(node.type==="Identifier"&&node.name==="require"){
    const permitted=(parent?.type==="CallExpression"&&key==="callee")||(parent?.type==="MemberExpression"&&!parent.computed&&((key==="object"&&parent.property.name==="resolve")||(key==="property"&&parent.object.name==="module")))||(parent?.type==="UnaryExpression"&&parent.operator==="typeof")||(parent?.type==="ObjectProperty"&&key==="key"&&!parent.computed);
    need(permitted,"unresolved-require-alias");
   }
   if(mode){const target=resolve(value,file,mode);edges.push({from:file,specifier:value,mode,...target});need(edges.length<=control.graphPolicies.runtimeEdgesMax,"runtime-edge-capacity");if(target.path)pending.push(target.path);}
   for(const [field,child]of Object.entries(node))if(field!=="loc"&&field!=="tokens"&&field!=="comments"&&child&&typeof child==="object")stack.push({node:child,parent:node,key:field});
  }
 }
 // All static edges have an actual Node-equivalent confined target before imports.
 const entry=path.join(root,control.rootModules[0]),factory=path.join(root,control.rootModules[1]),bundle=path.join(root,control.rootModules[2]);
 const verifier=await import(pathToFileURL(entry).href),factories=await import(pathToFileURL(factory).href),artifacts=await import(pathToFileURL(bundle).href);
 need(typeof verifier.verifyProfileEvidence==="function"&&typeof factories.bindServedProfileConfiguration==="function"&&typeof artifacts.createCumulativeTransactionalArtifactBundle==="function","actual-emitted-apis");
 const selection=factories.loadProfileContractSelection("transactional",false);need(selection.profile==="transactional"&&selection.history===false&&selection.required.length===251&&selection.retiredIds.length===13,"ordinary-complete-selection");
 const cache=Object.keys(createRequire(import.meta.url).cache).map(f=>fs.realpathSync(f));for(const file of cache)need(seen.has(file),"unproved-actual-cjs-module");
 for(const binding of [parserBinding,parserManifestBinding]){const bytes=raw(binding.realpath,16777216);files.set(binding.realpath,{path:binding.realpath,realpath:binding.realpath,sha256:hash(bytes),bytes:bytes.length});}
 verifyFiles([...files.values()]);
 const proof={sourceIdentity:c.sourceIdentity,sourceBindingSha256:c.gate.sourceBinding.sha256,sourceCommit:c.binding.head,sourceTree:c.binding.tree,privateEmitInventorySha256:hash(raw(control.emitInventoryPath,4194304)),installedProvenance:c.gate.installedProvenance,consumerInventory:c.gate.consumerInventory,workspaceTools:c.gate.workspaceTools,entryModules:[entry,factory,bundle],files:[...files.values()].sort((a,b)=>a.path.localeCompare(b.path)),resolvedEdges:edges,runtimePackages:[...packages.values()],actualCJSCacheAfterFactoryImports:cache,selection:{profile:selection.profile,history:selection.history,required:selection.required.length,retired:selection.retiredIds.length,catalogDigests:selection.catalogDigests},limits:control.graphPolicies,qualification:false,limitations:"Fixed source/emitted/compiler and resolved locked-package graph; actual CJS cache checked, ESM edges statically bound to Node-equivalent unconditional targets. No custom loader or product invocation. Late audit repeats actual CJS-cache confinement after verifier calls. Auth/policy metadata remains same-run source attestation, not token regeneration or public admission."};
 const encoded=Buffer.from(JSON.stringify(proof));need(encoded.length<=control.graphPolicies.toolProofBytesMax,"private-proof-capacity");output(path.join(E,"private-emit-proof.json"),proof);console.log(JSON.stringify({kind:"private-graph-proof",proofSha256:hash(raw(path.join(E,"private-emit-proof.json"),4194304)),files:files.size,edges:edges.length,claimEligible:false}));
}
main().catch(error=>{console.error(JSON.stringify({kind:"private-graph-refused",errorClass:error?.constructor?.name??"Error",claimEligible:false}));process.exitCode=1;});
