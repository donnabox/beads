/** External observer: exact source oracle bodies; only installed lifecycle adapted. */
import assert from 'node:assert/strict';
import {createHash,randomBytes} from 'node:crypto';
import {request as authorizationHttpRequest} from 'node:http';
import {existsSync,mkdtempSync,readFileSync,realpathSync,rmSync,writeFileSync} from 'node:fs';
import {tmpdir} from 'node:os';
import path from 'node:path';
import {fileURLToPath,pathToFileURL} from 'node:url';
import {createRequire} from 'node:module';
import {recordRawSse} from './raw-sse-recorder.ts';
import {launchInstalledTransaction} from './launcher.mjs';
const P=path.dirname(fileURLToPath(import.meta.url));
const E=path.join(P,'run');
const cohort=process.argv[2],output=path.resolve(process.argv[3]??'');
if(!['visibility-handles','active-event'].includes(cohort)||output!==path.join(E,cohort+'.json'))
 throw Error('exact external report path required');
const provenance=JSON.parse(readFileSync(path.join(E,'installed-provenance.json'),'utf8'));
const consumer=realpathSync(provenance.consumer);
const entry=realpathSync(provenance.entryRealpath);
if(!entry.startsWith(consumer+path.sep))throw Error('installed entry containment');
const hash=(bytes:Uint8Array|string)=>createHash('sha256').update(bytes).digest('hex');
if(hash(readFileSync(entry))!==provenance.entrySha256)throw Error('installed entry hash');
const installedRequire=createRequire(path.join(consumer,'package.json'));
const protocolManifest=realpathSync(path.join(consumer,'node_modules/@bdp/protocol/package.json'));
if(!protocolManifest.startsWith(consumer+path.sep))throw Error('installed protocol containment');
const protocolRequire=createRequire(protocolManifest);
function installedModule(name:string){
 const resolver=name==='@bdp/protocol'?installedRequire:protocolRequire;
 const p=realpathSync(resolver.resolve(name));
 if(!p.startsWith(consumer+path.sep))throw Error('installed dependency escape');
 return pathToFileURL(p).href;
}
const {Ajv2020}=await import(installedModule('ajv/dist/2020.js'));
const {isJsonSchemaUri,isJsonSchemaDateTime}=await import(installedModule('@bdp/protocol'));
const token=randomBytes(32).toString('base64url');
type InstalledRuntime={listenUrl:string;targets:Record<string,string>;replaceAuthorizationPolicy(options:{timeoutMs:number}):Promise<Record<string,unknown>>;close():Promise<void>};
const cases:{name:string;run:()=>Promise<void>}[]=[];
const it=(name:string,run:()=>Promise<void>,_milliseconds:number)=>cases.push({name,run});
function subset(actual:unknown,expected:unknown):void{
 if(expected===null||typeof expected!=='object'){assert.deepStrictEqual(actual,expected);return}
 assert.ok(actual!==null&&typeof actual==='object');
 if(Array.isArray(expected)){assert.ok(Array.isArray(actual));assert.equal(actual.length,expected.length)}
 for(const [key,value]of Object.entries(expected))subset((actual as Record<string,unknown>)[key],value);
}
function expect(actual:unknown,label='authorization assertion'){
 return {
  toBe(expected:unknown){assert.strictEqual(actual,expected,label)},
  toBeNull(){assert.strictEqual(actual,null,label)},
  toBeUndefined(){assert.strictEqual(actual,undefined,label)},
  toEqual(expected:unknown){assert.deepStrictEqual(actual,expected,label)},
  toMatchObject(expected:unknown){subset(actual,expected)},
  toHaveLength(expected:number){assert.equal((actual as {length:number}).length,expected,label)},
  not:{toBe(expected:unknown){assert.notStrictEqual(actual,expected,label)}}
 };
}
let finalReport:Record<string,unknown>|undefined;
async function launchRuntime(directory:string,signal:AbortSignal){
 const wrapper=path.join(P,'authorization-bootstrap.mjs');
 const descriptor={id:'https://event-sse.invalid/acme/types/task',name:'Task',describes:'bead',conformsTo:[]};
 const configuration={scope:'https://event-sse.invalid/acme/',directory,create:true,
  credentials:[{token,principal:'p',canMutate:true}],
  artifacts:{descriptors:[{retrievalUri:descriptor.id,utf8:JSON.stringify(descriptor)}],schemas:[]}};
 const child=await launchInstalledTransaction(wrapper,configuration,signal);
 const r=child.ready;
 const fail=async()=>{await child.close();throw Error('actual installed ready binding mismatch')};
 if(r.profile!=='transactional'||r.history!==false||r.claimEligible!==false||
    r.authorityMode!=='canonical'||r.scope!==configuration.scope||
    r.payloadIdentity!==provenance.payloadIdentity||typeof r.artifactIdentity!=='string'||
    !/^development-schema5-[a-f0-9]{64}$/.test(r.artifactIdentity))return fail();
 for(const field of ['runtimeConfigurationIdentity','policyIdentity','authenticationIdentity'])
  if(typeof r[field]!=='string'||!/^[0-9a-f]{64}$/.test(r[field]))return fail();
 const dial=new URL(r.listenUrl);
 if(dial.protocol!=='http:'||dial.hostname!=='127.0.0.1'||!dial.port||
    dial.username||dial.password||dial.pathname!=='/'||dial.search||dial.hash)return fail();
 const targetNames={createBead:'create-bead',updateBead:'update-bead',deleteBead:'delete-bead',
  createLink:'create-link',updateLink:'update-link',deleteLink:'delete-link',putAlias:'put-alias',
  deleteAlias:'delete-alias',updateWhere:'update-where',deleteWhere:'delete-where',batch:'batch',sequence:'sequence'};
 const targets=Object.fromEntries(Object.entries(targetNames).map(([key,value])=>[key,configuration.scope+'operations/'+value]));
 try{assert.deepStrictEqual(r.targets,targets)}catch{return fail()}
 let published=false;
 const publication=async(options:{timeoutMs:number})=>{
  if(published||options.timeoutMs!==1000)throw Error('one fixed administrative publication');
  published=true;
  return await new Promise<Record<string,unknown>>((resolve,reject)=>{
   const finish=(error?:Error,value?:Record<string,unknown>)=>{
    clearTimeout(timer);child.child.off('message',message);child.child.off('exit',ended);
    error?reject(error):resolve(value!);
   };
   const message=(value:unknown)=>{
    if(!value||typeof value!=='object')return finish(Error('administrative response shape'));
    const v=value as Record<string,unknown>;
    if(Object.keys(v).sort().join(',')!=='event,policy,requestId'||v.event!=='authorization.published'||v.requestId!==1)
     return finish(Error('administrative response identity'));
    const policy=v.policy;
    if(!policy||typeof policy!=='object'||Array.isArray(policy)||
       Object.keys(policy).sort().join(',')!=='authorizationView,bindingKind,canMutate,generation,readPolicy,revision')
     return finish(Error('administrative policy shape'));
    const actual=policy as Record<string,unknown>;
    if(actual.revision!==1||actual.bindingKind!=='scope'||actual.generation!==1||
       actual.readPolicy!=='empty'||actual.canMutate!==false||
       typeof actual.authorizationView!=='string'||
       !/^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/.test(actual.authorizationView))
     return finish(Error('administrative policy values'));
    finish(undefined,actual);
   };
   const ended=()=>finish(Error('child exited before publication'));
   const timer=setTimeout(()=>finish(Error('administrative IPC deadline')),2000);
   child.child.on('message',message);child.child.once('exit',ended);
   child.child.send({command:'publish-empty-policy',requestId:1},error=>{if(error)finish(Error('administrative IPC send'))});
  });
 };
 return {...child,targets,listenUrl:r.listenUrl,replaceAuthorizationPolicy:publication,
  binding:{sourceCommit:provenance.sourceCommit,sourceTree:provenance.sourceTree,
   installedEntry:entry,installedEntrySha256:hash(readFileSync(entry)),
   wrapper:wrapper,wrapperSha256:hash(readFileSync(wrapper)),
   payloadIdentity:r.payloadIdentity,artifactIdentity:r.artifactIdentity,authorityMode:r.authorityMode,scope:r.scope,
   runtimeConfigurationIdentity:r.runtimeConfigurationIdentity,policyIdentity:r.policyIdentity,
   authenticationIdentity:r.authenticationIdentity,
   descriptorSha256:hash(JSON.stringify(descriptor)),canonicalHost:new URL(configuration.scope).host}};
}
// Private trusted administration exercised through a real listener; no wire admin route.
const authorizationScope = "https://event-sse.invalid/acme/";
const authorizationType = `${authorizationScope}types/task`;
const authorizationSeed = (ids: readonly string[]) =>
  JSON.stringify({
    operations: ids.map((id) => ({
      operation: "createBead",
      id: `beads/${id}`,
      type: authorizationType,
      properties: { n: 0 },
    })),
  });
const authorizationValidators = (() => {
  const schema = JSON.parse(
    readFileSync(path.join(consumer,"node_modules/@bdp/protocol/schemas/bdp-v0.schema.json"), "utf8"),
  );
  const ajv = new Ajv2020({ strict: true, allErrors: true });
  ajv.addFormat("uri", { type: "string", validate: isJsonSchemaUri });
  ajv.addFormat("date-time", { type: "string", validate: isJsonSchemaDateTime });
  ajv.addSchema(schema);
  return Object.fromEntries(
    [
      "mutationReceipt",
      "currentBeadRecord",
      "currentBeadCollection",
      "snapshotManifest",
      "eventPage",
      "transactionalProblem",
    ].map((name) => [name, ajv.compile({ $ref: `${schema.$id}#/$defs/${name}` })]),
  );
})();
function authorizationObject(value: unknown): Record<string, unknown> {
  if (!value || typeof value !== "object" || Array.isArray(value))
    throw Error("expected observed object");
  return value as Record<string, unknown>;
}
function authorizationText(value: unknown): string {
  if (typeof value !== "string" || !value) throw Error("expected issued nonempty identity");
  return value;
}
function authorizationRows(value: unknown): Record<string, unknown>[] {
  if (!Array.isArray(value)) throw Error("expected observed array");
  return value.map(authorizationObject);
}
async function authorizationBound<T>(
  promise: Promise<T>,
  label: string,
  milliseconds = 5000,
): Promise<T> {
  let timer: ReturnType<typeof setTimeout> | undefined;
  try {
    return await Promise.race([
      promise,
      new Promise<never>((_resolve, reject) => {
        timer = setTimeout(() => reject(Error(label)), milliseconds);
      }),
    ]);
  } finally {
    clearTimeout(timer);
  }
}
type AuthorizationResponse = {
  status: number;
  headers: Record<string, string>;
  body: Record<string, unknown> | undefined;
  bodyBytes: number;
};
function authorizationContext(headers: Record<string, string>) {
  return {
    scopeEpoch: authorizationText(headers["bdp-scope-epoch"]),
    authorizationView: authorizationText(headers["bdp-authorization-view"]),
    scopePosition: authorizationText(headers["bdp-scope-position"]),
  };
}
function authorizationChecked(response: AuthorizationResponse, status: number, schema: string) {
  expect(response.status, `authorization-transition:status ${schema}`).toBe(status);
  const validate = authorizationValidators[schema]!;
  expect(
    validate(response.body),
    `authorization-transition:schema ${schema} ${JSON.stringify(validate.errors)}`,
  ).toBe(true);
  return authorizationObject(response.body);
}
function authorizationReceipt(
  response: AuthorizationResponse,
  key: string,
  ids: readonly string[],
  position: string,
) {
  const receipt = authorizationChecked(response, 200, "mutationReceipt");
  expect(receipt).toMatchObject({
    status: "completed",
    detail: "available",
    idempotencyKey: key,
    requiredPosition: position,
    effectPosition: position,
    next: null,
  });
  const context = authorizationContext(response.headers);
  expect(receipt.scopeEpoch).toBe(context.scopeEpoch);
  expect(receipt.authorizationView).toBe(context.authorizationView);
  expect(context.scopePosition).toBe(position);
  const results = authorizationRows(receipt.results);
  expect(results).toHaveLength(ids.length);
  const resources = results.map((entry, index) => {
    expect(entry).toMatchObject({ operationIndex: index, outcome: "created" });
    const resource = authorizationObject(entry.resource);
    expect(resource).toEqual({
      id: `${authorizationScope}beads/${ids[index]}`,
      type: authorizationType,
      revision: authorizationText(resource.revision),
      properties: { n: 0 },
      metadata: {},
    });
    return resource;
  });
  return { receipt, resources, context };
}
function authorizationEvents(
  page: Record<string, unknown>,
  receipt: Record<string, unknown>,
  resources: readonly Record<string, unknown>[],
) {
  expect(page.source).toBe(`${authorizationScope}events/`);
  const events = authorizationRows(page.events);
  expect(events).toHaveLength(resources.length);
  for (const [index, event] of events.entries()) {
    expect(event).toEqual({
      id: authorizationText(event.id),
      ordinal: index,
      type: "created",
      source: `${authorizationScope}events/`,
      subject: resources[index]!.id,
      subjectType: authorizationType,
      transaction: receipt.transaction,
      time: authorizationText(event.time),
      data: { revision: resources[index]!.revision, properties: { n: 0 }, metadata: {} },
    });
  }
  return events;
}
async function withAuthorizationRuntime(
  name: string,
  maximumRequests: number,
  use: (fixture: {
    runtime: InstalledRuntime;
    request: (
      target: string,
      method?: "GET" | "HEAD" | "POST",
      body?: string,
      key?: string,
    ) => Promise<AuthorizationResponse>;
    stream: (target: string, headers?: [string, string][]) => ReturnType<typeof recordRawSse>;
    observations: Record<string, unknown>[];
  }) => Promise<void>,
) {
  const directory = mkdtempSync(path.join(tmpdir(), "bdp-authorization-installed-"));
  const observations: Record<string, unknown>[] = [],
    failures: unknown[] = [];
  const streams: ReturnType<typeof recordRawSse>[] = [];
  const controllers: AbortController[] = [];
  let runtime:
    | InstalledRuntime
    | undefined;
  const work = new AbortController();
  const workTimer = setTimeout(() => work.abort(Error("installed authorization work deadline")), 30000);
  let installed: Awaited<ReturnType<typeof launchRuntime>> | undefined;
  let cursorIssuedAt: number | undefined;
  let timingLimited = false;
  let attempts = 0,
    closed = false;
  const reserve = () => {
    if (++attempts > maximumRequests) throw Error("authorization HTTP request bound");
  };
  try {
    installed = await launchRuntime(directory, work.signal);
    runtime = installed;
    const active = runtime;
    const request = async (
      target: string,
      method: "GET" | "HEAD" | "POST" = "GET",
      body?: string,
      key?: string,
    ): Promise<AuthorizationResponse> => {
      reserve();
      work.signal.throwIfAborted();
      if (name === "visibility-handles" && attempts === 4) cursorIssuedAt = performance.now();
      if (name === "visibility-handles" && attempts === 13) {
        const age = performance.now() - (cursorIssuedAt ?? -Infinity);
        observations.push({cursorDispatchAgeMs: age, requiredBelowMs: 500});
        if (!(age < 500)) { timingLimited = true; throw Error("collection cursor timing limit"); }
      }
      const canonical = new URL(target, authorizationScope);
      expect(canonical.origin).toBe(new URL(authorizationScope).origin);
      const dial = new URL(canonical.pathname + canonical.search, active.listenUrl);
      const received = Promise.withResolvers<AuthorizationResponse>();
      void received.promise.catch(() => {});
      let writeComplete = false;
      const req = authorizationHttpRequest(
        dial,
        {
          method,
          headers: {
            host: new URL(authorizationScope).host,
            authorization: `Bearer ${token}`,
            accept: "application/json, application/problem+json",
            connection: "close",
            ...(body === undefined
              ? {}
              : {
                  "content-type": "application/json",
                  "content-length": String(Buffer.byteLength(body)),
                  "idempotency-key": key!,
                }),
          },
        },
        (res) => {
          const chunks: Buffer[] = [];
          let bytes = 0;
          res.on("data", (chunk: Buffer) => {
            bytes += chunk.length;
            if (bytes > 32768) {
              res.destroy(Error("authorization response byte bound"));
              req.destroy();
              return;
            }
            chunks.push(Buffer.from(chunk));
          });
          res.once("error", received.reject);
          res.once("aborted", () => received.reject(Error("authorization response aborted")));
          res.once("end", () => {
            try {
              const text = Buffer.concat(chunks).toString("utf8");
              const headers = Object.fromEntries(
                Object.entries(res.headers).map(([k, v]) => [
                  k,
                  Array.isArray(v) ? v.join(", ") : (v ?? ""),
                ]),
              );
              if (method === "HEAD" && bytes !== 0) throw Error("authorization HEAD delivered body bytes");
              const value = method === "HEAD" ? undefined : authorizationObject(JSON.parse(text));
              observations.push({
                index: attempts,
                method,
                target: canonical.pathname + canonical.search,
                status: res.statusCode,
                bodyBytes: bytes,
                contentType: headers["content-type"],
                contentLength: headers["content-length"],
                writeComplete,
                canonicalHost: new URL(authorizationScope).host,
                bodySha256: createHash("sha256").update(text).digest("hex"),
                requestBodyBytes: body === undefined ? 0 : Buffer.byteLength(body),
                ...(body === undefined
                  ? {}
                  : { requestBodySha256: createHash("sha256").update(body).digest("hex") }),
                context: [
                  headers["bdp-scope-epoch"],
                  headers["bdp-authorization-view"],
                  headers["bdp-scope-position"],
                ],
              });
              if (!writeComplete) throw Error("request write incomplete");
              if (name === "visibility-handles" && attempts === 13) {
                const age = performance.now() - (cursorIssuedAt ?? -Infinity);
                observations.push({cursorResponseAgeMs: age, requiredBelowMs: 500});
                if (!(age < 500)) { timingLimited = true; throw Error("collection cursor timing limit"); }
              }
              received.resolve({ status: res.statusCode!, headers, body: value, bodyBytes: bytes });
            } catch (error) {
              received.reject(error);
            }
          });
        },
      );
      req.once("finish", () => { writeComplete = true; });
      req.once("error", received.reject);
      const socketClosed = new Promise<void>((resolve) =>
        req.once("close", () => {
          received.reject(Error("authorization request closed before complete response"));
          resolve();
        }),
      );
      const timer = setTimeout(
        () => req.destroy(Error("authorization finite request timeout")),
        5000,
      );
      req.end(body);
      try {
        return await received.promise;
      } finally {
        clearTimeout(timer);
        req.destroy();
        await authorizationBound(socketClosed, "authorization finite close timeout");
      }
    };
    const stream = (target: string, headers: [string, string][] = []) => {
      reserve();
      const controller = new AbortController();
      controllers.push(controller);
      const canonical = new URL(target, authorizationScope);
      const result = recordRawSse({
        dial: active.listenUrl,
        host: new URL(authorizationScope).host,
        target: canonical.pathname + canonical.search,
        signal: controller.signal,
        timeoutMs: 5000,
        headers: [
          ["Authorization", `Bearer ${token}`],
          ["Accept", "text/event-stream"],
          ...headers,
        ],
      });
      streams.push(result);
      const observation: Record<string, unknown> = {
        index: attempts,
        method: "GET",
        target: canonical.pathname + canonical.search,
        stream: true,
        canonicalHost: new URL(authorizationScope).host,
      };
      observations.push(observation);
      void result.ready.then(
        () => {
          observation.status = result.status;
          observation.context = [
            result.headers["bdp-scope-epoch"],
            result.headers["bdp-authorization-view"],
            result.headers["bdp-scope-position"],
          ];
        },
        () => {
          observation.failedBeforeHeaders = true;
        },
      );
      void result.closed.then(() => {
        observation.closed = true;
        observation.clientAbortedBeforeClose = controller.signal.aborted;
        observation.wireBytes = result.wireBytes.length;
        observation.completeFrames = result.frames.length;
        observation.idFrames = result.frames.filter(frame => frame.startsWith("id:")).length;
      });
      return result;
    };
    await use({ runtime: active, request, stream, observations });
    work.signal.throwIfAborted();
    expect(attempts).toBe(maximumRequests);
  } catch (error) {
    failures.push(error);
  } finally {
    clearTimeout(workTimer);
    console.info(
      JSON.stringify({
        label: "authorization-transition-before-cleanup",
        name,
        attempts,
        observations,
        failures: failures.length,
      }),
    );
    for (const controller of controllers) controller.abort(Error("authorization fixture cleanup"));
    for (const stream of streams) {
      try {
        await authorizationBound(stream.close(), "authorization stream close timeout");
      } catch (error) {
        failures.push(error);
      }
    }
    if (runtime) {
      try {
        await authorizationBound(runtime.close(), "authorization runtime close timeout", 10000);
        const pid = installed?.child.pid;
        let absent = false;
        if (pid) { try { process.kill(pid, 0); } catch (error) {
          if ((error as NodeJS.ErrnoException).code === "ESRCH") absent = true;
        }}
        if (!absent || installed?.child.exitCode !== 0 || installed?.child.signalCode !== null)
          throw Error("installed child clean exit required");
        closed = true;
      } catch (error) {
        failures.push(error);
      }
    }
    if (closed) {
      try {
        rmSync(directory, { recursive: true, force: true });
      } catch (error) {
        failures.push(error);
      }
    }
    console.info(
      JSON.stringify({
        label: "authorization-transition-cleanup",
        name,
        attempts,
        runtimeReturned: runtime !== undefined,
        closed,
        directory,
        directoryRemoved: !existsSync(directory),
        failures: failures.length,
        claimEligible: false,
      }),
    );
  }
  const pid = installed?.child.pid ?? null;
  let pidAbsent = false;
  if (pid !== null) { try { process.kill(pid, 0); } catch (error) {
    if ((error as NodeJS.ErrnoException).code === "ESRCH") pidAbsent = true;
  }}
  if (!pidAbsent || !closed || existsSync(directory)) failures.push(Error("installed cleanup incomplete"));
  finalReport = {
    label: "private-installed-authorization", name, attempts, maximumRequests, observations,
    result: failures.length ? "failed" : "passed", failureCount: failures.length,
    failureKinds: failures.map(error => error instanceof Error ? error.name : "NonError"),
    timingLimited, child: {pid, exitCode: installed?.child.exitCode ?? null,
      signalCode: installed?.child.signalCode ?? null, pidAbsent},
    cleanup: {closed, directory, directoryRemoved: !existsSync(directory)},
    binding: installed?.binding ?? null, claimEligible: false
  };
  if (failures.length)
    throw new AggregateError(failures, "authorization transport witness and cleanup");
}

it("narrows current Transactional visibility and fences issued handles after policy publication", async () => {
  await withAuthorizationRuntime("visibility-handles", 18, async (f) => {
    const ids = Array.from({ length: 32 }, (_, index) => `b${String(index).padStart(2, "0")}`);
    const firstBody = authorizationSeed(ids);
    const first = authorizationReceipt(
      await f.request(f.runtime.targets.batch, "POST", firstBody, "auth-transition-seed-32"),
      "auth-transition-seed-32",
      ids,
      "pos-1",
    );
    const second = authorizationReceipt(
      await f.request(
        f.runtime.targets.batch,
        "POST",
        authorizationSeed(["b32"]),
        "auth-transition-seed-1",
      ),
      "auth-transition-seed-1",
      ["b32"],
      "pos-2",
    );
    expect(second.context).toEqual({ ...first.context, scopePosition: "pos-2" });
    expect(authorizationChecked(await f.request("beads/b00"), 200, "currentBeadRecord")).toEqual(
      first.resources[0],
    );
    const collection = authorizationChecked(
      await f.request("beads/?limit=2"),
      200,
      "currentBeadCollection",
    );
    expect(collection.items).toEqual(first.resources.slice(0, 2));
    const collectionNext = authorizationText(collection.next);
    const snapshot = authorizationChecked(await f.request("snapshot"), 200, "snapshotManifest");
    expect(snapshot).toMatchObject({
      scope: authorizationScope,
      scopeEpoch: second.context.scopeEpoch,
      authorizationView: second.context.authorizationView,
      scopePosition: "pos-2",
      erasures: [],
      links: { items: [], next: null },
    });
    expect(authorizationObject(snapshot.beads).items).toEqual(first.resources);
    const snapshotNext = authorizationText(authorizationObject(snapshot.beads).next);
    const checkpoint = authorizationText(snapshot.checkpoint);
    expect(checkpoint).not.toBe("genesis");
    const page = authorizationChecked(await f.request("events/"), 200, "eventPage");
    const events = authorizationEvents(page, first.receipt, first.resources);
    authorizationText(page.next);
    const stream = f.stream("changes/?start=now");
    await stream.ready;
    expect(stream.status).toBe(200);
    expect(stream.headers["content-type"]).toBe("text/event-stream");
    expect(authorizationContext(stream.headers)).toEqual(second.context);
    let streamClosed = false;
    void stream.closed.then(() => {
      streamClosed = true;
    });
    const policy = await f.runtime.replaceAuthorizationPolicy({ timeoutMs: 1000 });
    f.observations.push({ publication: { ...policy }, oldStreamClosedAtPublication: streamClosed });
    expect(
      {
        generation: policy.generation,
        readPolicy: policy.readPolicy,
        canMutate: policy.canMutate,
        viewChanged: policy.authorizationView !== second.context.authorizationView,
      },
      "authorization-transition:generation-and-lease",
    ).toEqual({ generation: 1, readPolicy: "empty", canMutate: false, viewChanged: true });
    await authorizationBound(stream.closed, "authorization-transition:generation-and-lease", 1000);
    expect(streamClosed, "authorization-transition:generation-and-lease").toBe(true);
    expect(stream.frames.filter((frame) => frame.startsWith("id:"))).toEqual([]);
    const current = { ...second.context, authorizationView: policy.authorizationView };
    const problem = async (target: string, status: number, code: string) => {
      const response = await f.request(target);
      const value = authorizationChecked(response, status, "transactionalProblem");
      expect(value).toMatchObject({ status, code });
      expect(authorizationContext(response.headers)).toEqual(current);
      return value;
    };
    await problem("beads/b00", 404, "resource-not-found");
    const empty = await f.request("beads/?limit=2");
    expect(authorizationChecked(empty, 200, "currentBeadCollection")).toEqual({
      items: [],
      next: null,
    });
    expect(authorizationContext(empty.headers)).toEqual(current);
    const fresh = await f.request("snapshot");
    const emptySnapshot = authorizationChecked(fresh, 200, "snapshotManifest");
    expect(emptySnapshot).toMatchObject({
      scopeEpoch: current.scopeEpoch,
      authorizationView: current.authorizationView,
      scopePosition: "pos-2",
      beads: { items: [], next: null },
      links: { items: [], next: null },
      erasures: [],
    });
    expect(emptySnapshot.id).not.toBe(snapshot.id);
    expect(authorizationContext(fresh.headers)).toEqual(current);
    await problem(authorizationText(snapshot.id), 409, "foreign-view");
    await problem(snapshotNext, 409, "foreign-view");
    await problem(collectionNext, 409, "foreign-view");
    const reconnect = f.stream("changes/?start=now", [["Last-Event-ID", checkpoint]]);
    await reconnect.ready;
    await reconnect.closed;
    expect(reconnect.status).toBe(409);
    expect(reconnect.frames.filter((frame) => frame.startsWith("id:"))).toEqual([]);
    const reconnectProblem = authorizationObject(JSON.parse(reconnect.body));
    authorizationChecked(
      {
        status: reconnect.status,
        headers: reconnect.headers,
        body: reconnectProblem,
        bodyBytes: Buffer.byteLength(reconnect.body),
      },
      409,
      "transactionalProblem",
    );
    expect(reconnectProblem.code).toBe("foreign-view");
    expect(authorizationContext(reconnect.headers)).toEqual(current);
    await problem(
      `events/?after=${encodeURIComponent(authorizationText(events.at(-1)!.id))}`,
      409,
      "foreign-view",
    );
    const receiptResponse = await f.request(authorizationText(first.receipt.id));
    const retained = authorizationChecked(receiptResponse, 200, "mutationReceipt");
    const projected: Record<string, unknown> = { ...first.receipt, detail: "withheld" };
    delete projected.results;
    delete projected.next;
    delete projected.allocated;
    delete projected.expiresAt;
    expect(retained).toEqual(projected);
    expect(authorizationContext(receiptResponse.headers)).toEqual(current);
    const receiptHead = await f.request(authorizationText(first.receipt.id), "HEAD");
    expect(receiptHead.status).toBe(200);
    expect(receiptHead.body).toBeUndefined();
    expect(receiptHead.bodyBytes).toBe(0);
    expect(receiptResponse.headers["content-length"]).toBe(String(receiptResponse.bodyBytes));
    expect(receiptHead.headers["content-length"]).toBe(receiptResponse.headers["content-length"]);
    expect(authorizationContext(receiptHead.headers)).toEqual(current);
    for (const response of [receiptResponse, receiptHead]) {
      expect(response.headers["cache-control"]).toBe("private, no-store");
      expect(response.headers.etag).toBeUndefined();
      expect(response.headers["last-modified"]).toBeUndefined();
    }
    const replay = await f.request(
      f.runtime.targets.batch,
      "POST",
      firstBody,
      "auth-transition-seed-32",
    );
    expect(authorizationChecked(replay, 403, "transactionalProblem")).toMatchObject({
      status: 403,
      code: "forbidden",
    });
    expect(authorizationContext(replay.headers)).toEqual(current);
  });
}, 30000);

it("cancels an active Event stream after policy publication", async () => {
  await withAuthorizationRuntime("active-event", 4, async (f) => {
    const created = authorizationReceipt(
      await f.request(
        f.runtime.targets.batch,
        "POST",
        authorizationSeed(["a", "b"]),
        "auth-transition-event-seed",
      ),
      "auth-transition-event-seed",
      ["a", "b"],
      "pos-1",
    );
    const finite = authorizationChecked(await f.request("events/"), 200, "eventPage");
    const events = authorizationEvents(finite, created.receipt, created.resources);
    expect(finite.next).toBeNull();
    const stream = f.stream("events/");
    await stream.until(() => stream.frames.filter((frame) => frame.startsWith("id:")).length === 2);
    expect(stream.status).toBe(200);
    expect(authorizationContext(stream.headers)).toEqual(created.context);
    const actual = stream.frames
      .filter((frame) => frame.startsWith("id:"))
      .map((frame) => JSON.parse(frame.split("\ndata: ")[1]!));
    expect(actual).toEqual(events);
    const policy = await f.runtime.replaceAuthorizationPolicy({ timeoutMs: 1000 });
    expect(policy.generation).toBe(1);
    await authorizationBound(stream.closed, "authorization Event stream not cancelled", 1000);
    const after = await f.request("events/");
    expect(authorizationChecked(after, 200, "eventPage")).toEqual({
      source: `${authorizationScope}events/`,
      events: [],
      next: null,
    });
    expect(authorizationContext(after.headers)).toEqual({
      ...created.context,
      authorizationView: policy.authorizationView,
    });
    f.observations.push({
      deliveredFrames: actual.length,
      publishedView: policy.authorizationView,
      oldStreamClosed: stream.destroyed,
    });
  });
}, 30000);

let caught:unknown;
try {
 const chosen=cases[cohort==='visibility-handles'?0:1];
 assert.ok(chosen);await chosen.run();
} catch(error){caught=error}
if(!finalReport)finalReport={label:'private-installed-authorization',name:cohort,result:'failed',
 failureCount:1,failureKinds:['BeforeReport'],claimEligible:false};
let encoded=JSON.stringify(finalReport);
if(Buffer.byteLength(encoded)>1048576){
 caught ??= Error('installed report ceiling');
 encoded=JSON.stringify({label:'private-installed-authorization',name:cohort,result:'failed',
  failureCount:1,reportOverflow:true,claimEligible:false});
}
try{writeFileSync(output,encoded+'\n',{flag:'wx'});process.stdout.write(encoded+'\n')}
catch{caught ??= Error('report write failure')}
if(caught){process.stderr.write('private installed authorization observation failed\n');process.exitCode=1}
