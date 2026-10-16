/** External private validation IPC adapter; not part of the installed payload. */
import process from 'node:process';
import {readFileSync,realpathSync} from 'node:fs';
import {createHash} from 'node:crypto';
import {fileURLToPath,pathToFileURL} from 'node:url';
import path from 'node:path';
const P=path.dirname(fileURLToPath(import.meta.url));
const hash=b=>createHash('sha256').update(b).digest('hex');
const keys=(v,expected)=>v && typeof v==='object' && !Array.isArray(v) &&
 Object.keys(v).sort().join(',')===[...expected].sort().join(',');
async function main(){
 if(!process.send || process.argv.length!==2) throw Error('inherited IPC required');
 const proof=JSON.parse(readFileSync(path.join(P,'run/installed-provenance.json'),'utf8'));
 const consumer=realpathSync(proof.consumer);
 const entry=realpathSync(path.join(consumer,'node_modules/@bdp/server/dist/transaction-installed.js'));
 if(!entry.startsWith(consumer+path.sep)||entry!==proof.entryRealpath||hash(readFileSync(entry))!==proof.entrySha256)
  throw Error('installed entry binding');
 const input=await new Promise((resolve,reject)=>{
  const timer=setTimeout(()=>reject(Error('configuration deadline')),10000);
  process.once('message',v=>{clearTimeout(timer);resolve(v)});
  process.once('disconnect',()=>{clearTimeout(timer);reject(Error('configuration disconnected'))});
 });
 if(!keys(input,['configuration']))throw Error('configuration envelope');
 let stop;let stopping=false,published=false,chain=Promise.resolve();
 const stopped=new Promise(resolve=>{stop=()=>{stopping=true;resolve()}});
 for(const name of ['SIGTERM','SIGINT','disconnect'])process.once(name,stop);
 let runtime;
 const send=value=>new Promise((resolve,reject)=>{
  if(!process.connected)return reject(Error('IPC disconnected'));
  process.send(value,error=>error?reject(error):resolve());
 });
 process.on('message',message=>{
  if(keys(message,['command'])&&message.command==='close'){stop();return}
  chain=chain.then(async()=>{
   if(!keys(message,['command','requestId'])||message.command!=='publish-empty-policy'||
      message.requestId!==1||published||stopping||!runtime)throw Error('unsupported administration');
   published=true;
   const policy=await runtime.replaceAuthorizationPolicy({timeoutMs:1000});
   await send({event:'authorization.published',requestId:1,policy:{
    revision:policy.revision,bindingKind:policy.bindingKind,generation:policy.generation,
    authorizationView:policy.authorizationView,readPolicy:policy.readPolicy,canMutate:policy.canMutate}});
  }).catch(()=>{process.exitCode=2;stop()});
 });
 try{
  runtime=await (await import(pathToFileURL(entry).href)).startInstalledTransactionCandidate(input.configuration);
  if(!process.connected)stop();
  if(!stopping)await send({event:'qualification.ready',...runtime.configurationMetadata,
   listenUrl:runtime.listenUrl,targets:runtime.targets,payloadIdentity:runtime.payloadIdentity});
  await stopped;await chain;
 }finally{if(runtime)await runtime.close();if(process.connected)process.disconnect()}
}
void main().catch(()=>{
 process.stderr.write('private installed authorization adapter failed\n');
 process.exitCode=2;if(process.connected)process.disconnect();
});
