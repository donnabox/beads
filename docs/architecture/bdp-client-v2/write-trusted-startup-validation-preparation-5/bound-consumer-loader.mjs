// Copied bounded external-consumer adaptation of scripts/smoke-executables.mjs
// 133-461 exact inventory/public entry/declarations, 483-505 root private values,
// 554-616 private specifiers (omit absent @bdp/conformance),1193-1353 pure loader.
// Shared exact export checker copied from scripts/smoke-safety.mjs. No source imports.
import { createRequire } from "node:module";
import { createHash } from "node:crypto";
import { isDeepStrictEqual } from "node:util";
import { createServer } from "node:net";
import { mkdtemp, readFile, readdir, realpath, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import path from "node:path";
import { pathToFileURL } from "node:url";
const executionRoot = "/Users/dbox/repos/wt/vickie/bdp-integration-logs/write-trusted-startup-validation-preparation-5/run";
const installRoot = path.join(executionRoot, "consumer");
const proofPath = path.join(executionRoot, "installed-provenance.json");
let proofBytes, proof, entryRealpath, temporaryRoot;
const digest = (bytes) => createHash("sha256").update(bytes).digest("hex");
const stats = {profiles:[],loaderCalls:0,accepted:0,contractRefusals:0,privateSpecifierCount:0,
  publicExports:[],configuredPortAcquisitionExcluded:false,storeDirectoriesAbsent:false};
let cleanupDeadline;
const cleanupRemaining=()=>{cleanupDeadline ??= performance.now()+10000;return Math.max(1,cleanupDeadline-performance.now());};
let guardCloseAttempted=false, guardCloseSettled=false, guardCloseSucceeded=false;
let witnessGuard, workSettled=false, workTimedOut=false, cleanupUnknown=false, temporaryDirectoryRemoved=false, failed=false;
// Fixed guard fails the helper; it never turns unknown closure into success.
const totalGuard=setTimeout(()=>{console.error("bounded pure-loader process guard expired; preserve evidence");process.exit(1);},60000);
async function within(promise, milliseconds, phase) {
 let timer;
 try {return await Promise.race([promise,new Promise((_,reject)=>{timer=setTimeout(()=>{
   const error=new Error("bounded pure-loader "+phase+" deadline");error.phase=phase;reject(error);
 },milliseconds);})]);} finally {clearTimeout(timer);}
}
function assertServerExports(manifest) {
  const expected = {
    ".": { types: "./dist/index.d.ts", default: "./dist/index.js" },
    "./development": { types: "./dist/development.d.ts", default: "./dist/development.js" },
    "./reference-configuration": {
      types: "./dist/reference-deployment.d.ts",
      default: "./dist/reference-deployment.js",
    },
  };
  if (!isDeepStrictEqual(manifest.exports, expected))
    throw new Error(
      "packed server exports differ from the approved root and explicit development surface",
    );
}

async function runWitness() {
proofBytes = await readFile(proofPath);
proof = JSON.parse(proofBytes);
if (process.env.TRUSTED_STARTUP_PROVENANCE_SHA256 !== digest(proofBytes))
  throw new Error("root-bound consumer provenance differs");
entryRealpath = await realpath(proof.entryRealpath);
if (entryRealpath !== process.env.TRUSTED_STARTUP_LOADER_ENTRY ||
    !entryRealpath.startsWith(await realpath(installRoot) + path.sep) ||
    digest(await readFile(entryRealpath)) !== proof.entrySha256)
  throw new Error("root-bound consumer loader entry differs");
temporaryRoot = await mkdtemp(path.join(tmpdir(), "bdp-bound-loader-"));
 const installedServerRoot = path.join(installRoot, "node_modules", "@bdp", "server");
 const installedServerOutput = await readdir(path.join(installedServerRoot, "dist"));
    // Intentional package-surface tripwire: adding or removing a shipped server
    // module is a deliberate reviewed change, not an incidental build artifact.
    const expectedServerOutput = new Set([
      "installed-payload.d.ts",
      "installed-payload.d.ts.map",
      "installed-payload.js",
      "installed-payload.js.map",
      "transaction-installed-configuration.d.ts",
      "transaction-installed-configuration.d.ts.map",
      "transaction-installed-configuration.js",
      "transaction-installed-configuration.js.map",
      "transaction-installed.d.ts",
      "transaction-installed.d.ts.map",
      "transaction-installed.js",
      "transaction-installed.js.map",
      "transaction-installed-process.d.ts",
      "transaction-installed-process.d.ts.map",
      "transaction-installed-process.js",
      "transaction-installed-process.js.map",

      "alias-evaluator.d.ts",
      "alias-evaluator.d.ts.map",
      "alias-evaluator.js",
      "alias-evaluator.js.map",
      "development-composition.d.ts",
      "development-composition.d.ts.map",
      "development-composition.js",
      "development-composition.js.map",
      "read-http-routing.d.ts",
      "read-http-routing.d.ts.map",
      "read-http-routing.js",
      "read-http-routing.js.map",
      "authority-read.d.ts",
      "authority-read.d.ts.map",
      "authority-read.js",
      "authority-read.js.map",
      "installed-schema-pattern.d.ts",
      "installed-schema-pattern.d.ts.map",
      "installed-schema-pattern.js",
      "installed-schema-pattern.js.map",
      "installed-schema-decimal.d.ts",
      "installed-schema-decimal.d.ts.map",
      "installed-schema-decimal.js",
      "installed-schema-decimal.js.map",
      "installed-schema-recursion.d.ts",
      "installed-schema-recursion.d.ts.map",
      "installed-schema-recursion.js",
      "installed-schema-recursion.js.map",
      "installed-schema-evaluator.d.ts",
      "installed-schema-evaluator.d.ts.map",
      "installed-schema-evaluator.js",
      "installed-schema-evaluator.js.map",
      "installed-schema-value.d.ts",
      "installed-schema-value.d.ts.map",
      "installed-schema-value.js",
      "installed-schema-value.js.map",
      "installed-schema-builtins.d.ts",
      "installed-schema-builtins.d.ts.map",
      "installed-schema-builtins.js",
      "installed-schema-builtins.js.map",
      "installed-schema-graph.d.ts",
      "installed-schema-graph.d.ts.map",
      "installed-schema-graph.js",
      "installed-schema-graph.js.map",
      "installed-schema-shape.d.ts",
      "installed-schema-shape.d.ts.map",
      "installed-schema-shape.js",
      "installed-schema-shape.js.map",
      "installed-schema-uri.d.ts",
      "installed-schema-uri.d.ts.map",
      "installed-schema-uri.js",
      "installed-schema-uri.js.map",
      "installed-contract-artifact.d.ts",
      "installed-contract-artifact.d.ts.map",
      "installed-contract-artifact.js",
      "installed-contract-artifact.js.map",
      "member-executor.d.ts",
      "member-executor.d.ts.map",
      "member-executor.js",
      "member-executor.js.map",
      "member-identity.d.ts",
      "member-identity.d.ts.map",
      "member-identity.js",
      "member-identity.js.map",
      "member-numeric-diagnostics.d.ts",
      "member-numeric-diagnostics.d.ts.map",
      "member-numeric-diagnostics.js",
      "member-numeric-diagnostics.js.map",
      "recovery-store.d.ts",
      "recovery-store.d.ts.map",
      "recovery-store.js",
      "recovery-store.js.map",
      "resource-evaluator.d.ts",
      "resource-evaluator.d.ts.map",
      "resource-evaluator.js",
      "resource-evaluator.js.map",
      "json-value-equality.d.ts",
      "json-value-equality.d.ts.map",
      "json-value-equality.js",
      "json-value-equality.js.map",
      "retained-disclosure.d.ts",
      "retained-disclosure.d.ts.map",
      "retained-disclosure.js",
      "retained-disclosure.js.map",
      "semantic-identity.d.ts",
      "semantic-identity.d.ts.map",
      "semantic-identity.js",
      "semantic-identity.js.map",
      "transaction-set.d.ts",
      "transaction-set.d.ts.map",
      "transaction-set.js",
      "transaction-set.js.map",
      "transaction-batch.d.ts",
      "transaction-batch.d.ts.map",
      "transaction-batch.js",
      "transaction-batch.js.map",
      "transaction-comparison.d.ts",
      "transaction-comparison.d.ts.map",
      "transaction-comparison.js",
      "transaction-comparison.js.map",
      "transaction-dependencies.d.ts",
      "transaction-dependencies.d.ts.map",
      "transaction-dependencies.js",
      "transaction-dependencies.js.map",
      "transaction-sse.d.ts",
      "transaction-sse.d.ts.map",
      "transaction-sse.js",
      "transaction-sse.js.map",
      "node-http-disconnect.d.ts",
      "node-http-disconnect.d.ts.map",
      "node-http-disconnect.js",
      "node-http-disconnect.js.map",
      "transaction-authorization.d.ts",
      "transaction-authorization.d.ts.map",
      "transaction-authorization.js",
      "transaction-authorization.js.map",
      "transaction-changefeed.d.ts",
      "transaction-changefeed.d.ts.map",
      "transaction-changefeed.js",
      "transaction-changefeed.js.map",
      "transaction-events.d.ts",
      "transaction-events.d.ts.map",
      "transaction-events.js",
      "transaction-events.js.map",
      "transaction-snapshot.d.ts",
      "transaction-snapshot.d.ts.map",
      "transaction-snapshot.js",
      "transaction-snapshot.js.map",
      "transaction-read.d.ts",
      "transaction-read.d.ts.map",
      "transaction-read.js",
      "transaction-read.js.map",
      "transaction-sequence.d.ts",
      "transaction-sequence.d.ts.map",
      "transaction-sequence.js",
      "transaction-sequence.js.map",
      "transaction-store.d.ts",
      "transaction-store.d.ts.map",
      "transaction-store.js",
      "transaction-store.js.map",
      "transaction-schema-http.d.ts",
      "transaction-schema-http.d.ts.map",
      "transaction-schema-http.js",
      "transaction-schema-http.js.map",
      "transaction-http.d.ts",
      "transaction-http.d.ts.map",
      "transaction-http.js",
      "transaction-http.js.map",
      "transaction-runtime-configuration.d.ts",
      "transaction-runtime-configuration.d.ts.map",
      "transaction-runtime-configuration.js",
      "transaction-runtime-configuration.js.map",
      "transaction-runtime.d.ts",
      "transaction-runtime.d.ts.map",
      "transaction-runtime.js",
      "transaction-runtime.js.map",
      "transaction-consistency.d.ts",
      "transaction-consistency.d.ts.map",
      "transaction-consistency.js",
      "transaction-consistency.js.map",
      "transaction-discovery.d.ts",
      "transaction-discovery.d.ts.map",
      "transaction-discovery.js",
      "transaction-discovery.js.map",
      "transaction-service.d.ts",
      "transaction-service.d.ts.map",
      "transaction-service.js",
      "transaction-service.js.map",
      "transaction-submission.d.ts",
      "transaction-submission.d.ts.map",
      "transaction-submission.js",
      "transaction-submission.js.map",
      "sequence-executor.d.ts",
      "sequence-executor.d.ts.map",
      "sequence-executor.js",
      "sequence-executor.js.map",
      "development.d.ts",
      "development.d.ts.map",
      "development.js",
      "development.js.map",
      "installed-reference-configuration.d.ts",
      "installed-reference-configuration.d.ts.map",
      "installed-reference-configuration.js",
      "installed-reference-configuration.js.map",
      "installed-reference.d.ts",
      "installed-reference.d.ts.map",
      "installed-reference.js",
      "installed-reference.js.map",
      "installed-reference-process.d.ts",
      "installed-reference-process.d.ts.map",
      "installed-reference-process.js",
      "installed-reference-process.js.map",
      "development-contracts.d.ts",
      "development-contracts.d.ts.map",
      "development-contracts.js",
      "development-contracts.js.map",
      "development-runtime.d.ts",
      "development-runtime.d.ts.map",
      "development-runtime.js",
      "development-runtime.js.map",
      "development-authority.d.ts",
      "development-authority.d.ts.map",
      "development-authority.js",
      "development-authority.js.map",
      "index.d.ts",
      "index.d.ts.map",
      "index.js",
      "index.js.map",
      "read-conformance-capability.d.ts",
      "read-conformance-capability.d.ts.map",
      "read-conformance-capability.js",
      "read-conformance-capability.js.map",
      "node-http-target.d.ts",
      "node-http-target.d.ts.map",
      "node-http-target.js",
      "node-http-target.js.map",
      "read-http.d.ts",
      "read-http.d.ts.map",
      "read-http.js",
      "read-http.js.map",
      "read-pagination.d.ts",
      "read-pagination.d.ts.map",
      "read-pagination.js",
      "read-pagination.js.map",
      "reference-deployment.d.ts",
      "reference-deployment.d.ts.map",
      "reference-deployment.js",
      "reference-deployment.js.map",
      "reference-deployment-state.d.ts",
      "reference-deployment-state.d.ts.map",
      "reference-deployment-state.js",
      "reference-deployment-state.js.map",
      "read-request.d.ts",
      "read-request.d.ts.map",
      "read-request.js",
      "read-request.js.map",
      "read-sequence-lifecycle.d.ts",
      "read-sequence-lifecycle.d.ts.map",
      "read-sequence-lifecycle.js",
      "read-sequence-lifecycle.js.map",
      "read-update-runtime-limits.d.ts",
      "read-update-runtime-limits.d.ts.map",
      "read-update-runtime-limits.js",
      "read-update-runtime-limits.js.map",
      "read-selector.d.ts",
      "read-selector.d.ts.map",
      "read-selector.js",
      "read-selector.js.map",
    ]);
    if (
      installedServerOutput.length !== expectedServerOutput.size ||
      installedServerOutput.some((entry) => !expectedServerOutput.has(entry))
    ) {
      const missing = [...expectedServerOutput].filter(
        (entry) => !installedServerOutput.includes(entry),
      );
      const unexpected = installedServerOutput.filter((entry) => !expectedServerOutput.has(entry));
      throw new Error(
        `packed server output mismatch; missing=[${missing.join(", ")}], unexpected=[${unexpected.join(", ")}]`,
      );
    }
    const installedServerFiles = await readdir(installedServerRoot, { recursive: true });
    if (installedServerFiles.some((entry) => entry.split(path.sep).includes("test-support")))
      throw new Error("packed server must not contain test-support files");
    const installedServerManifest = JSON.parse(
      await readFile(path.join(installedServerRoot, "package.json"), "utf8"),
    );
    assertServerExports(installedServerManifest);
    const consumer = createRequire(path.join(installRoot, "consumer.cjs"));
    if (
      (await realpath(consumer.resolve("@bdp/server"))) !==
      (await realpath(path.join(installedServerRoot, "dist", "index.js")))
    )
      throw new Error("packed server public root does not resolve to its approved entry");
    if (
      (await realpath(consumer.resolve("@bdp/server/development"))) !==
      (await realpath(path.join(installedServerRoot, "dist", "development.js")))
    )
      throw new Error("packed development facade does not resolve to its explicit entry");
    const deploymentEntry = consumer.resolve("@bdp/server/reference-configuration");
    if (
      (await realpath(deploymentEntry)) !==
      (await realpath(path.join(installedServerRoot, "dist", "reference-deployment.js")))
    )
      throw new Error("packed deployment loader does not resolve to its approved entry");
    const deployment = await import(pathToFileURL(deploymentEntry).href);
    if (
      Object.keys(deployment).sort().join(",") !==
      "ReferenceDeploymentError,loadReferenceDeployment"
    )
      throw new Error("packed deployment entry exposes unexpected capabilities");
    const publicDeclaration = await readFile(
      path.join(installedServerRoot, "dist", "reference-deployment.d.ts"),
      "utf8",
    );
    for (const name of [
      "reference-deployment-state",
      "installed-reference-configuration",
      "transaction-runtime",
      "TransactionRuntime",
      "credentials",
      "projectedInput",
      "authenticate",
      "metadata(",
      "readOptions(",
    ])
      if (publicDeclaration.includes(name))
        throw new Error("packed deployment declaration exposes private capture types");
    stats.publicExports = Object.keys(deployment).sort();
    await observePackedDeploymentLoader(deployment, temporaryRoot);
    const publicServer = await import(pathToFileURL(consumer.resolve("@bdp/server")).href);
    if (
      "assembleTransactionDiscovery" in publicServer ||
      "TRANSACTION_HTTP_OPERATIONS" in publicServer ||
      "createTransactionService" in publicServer ||
      "createTransactionHttpHandler" in publicServer ||
      "startTransactionRuntime" in publicServer ||
      "snapshotTransactionMinimum" in publicServer ||
      "waitForTransactionMinimum" in publicServer ||
      "checkTransactionMinimum" in publicServer ||
      "transactionReceiptMaximum" in publicServer ||
      "prepareTransactionSetCore" in publicServer ||
      "prepareSetCore" in publicServer
    )
      throw new Error("packed root exposes private transaction service");
    if (
      "loadReferenceDeployment" in publicServer ||
      "referenceDeploymentState" in publicServer ||
      "registerReferenceDeployment" in publicServer
    )
      throw new Error("packed root exposes deployment preparation or private registry");
    if ("createContextualReadPagination" in publicServer)
      throw new Error("packed root exposes private contextual pagination factory");
    for (const module of installedServerOutput.filter(
      (name) =>
        (name.startsWith("installed-schema-") || name.startsWith("transaction-")) &&
        name.endsWith(".js"),
    )) {
      try {
        consumer.resolve(`@bdp/server/dist/${module}`);
        throw new Error("packed private implementation module unexpectedly resolves");
      } catch (error) {
        if (error.code !== "ERR_PACKAGE_PATH_NOT_EXPORTED") throw error;
        stats.privateSpecifierCount++;
      }
    }

    for (const specifier of [
      "@bdp/server/reference-deployment",
      "@bdp/server/dist/reference-deployment.js",
      "@bdp/server/reference-deployment-state",
      "@bdp/server/dist/reference-deployment-state.js",
      "@bdp/server/development-composition",
      "@bdp/server/dist/development-composition.js",
      "@bdp/server/installed-payload",
      "@bdp/server/dist/installed-payload.js",
      "@bdp/server/transaction-installed-configuration",
      "@bdp/server/dist/transaction-installed-configuration.js",
      "@bdp/server/transaction-installed",
      "@bdp/server/dist/transaction-installed.js",
      "@bdp/server/transaction-installed-process",
      "@bdp/server/dist/transaction-installed-process.js",
      "@bdp/server/transaction-runtime-configuration",
      "@bdp/server/dist/transaction-runtime-configuration.js",
      "@bdp/server/transaction-snapshot",
      "@bdp/server/dist/transaction-snapshot.js",
      "@bdp/server/transaction-sse",
      "@bdp/server/dist/transaction-sse.js",
      "@bdp/server/node-http-disconnect",
      "@bdp/server/dist/node-http-disconnect.js",
      "@bdp/server/transaction-changefeed",
      "@bdp/server/dist/transaction-changefeed.js",
      "@bdp/server/read-http-routing",
      "@bdp/server/dist/read-http-routing.js",
      "@bdp/server/read-pagination",
      "@bdp/server/dist/read-pagination.js",
      "@bdp/server/installed-reference-configuration",
      "@bdp/server/dist/installed-reference-configuration.js",
      "@bdp/server/installed-reference",
      "@bdp/server/dist/installed-reference.js",
      "@bdp/server/installed-reference-process",
      "@bdp/server/dist/installed-reference-process.js",
      "@bdp/server/development-contracts",
      "@bdp/server/dist/development-contracts.js",
      "@bdp/server/development-runtime",
      "@bdp/server/dist/development-runtime.js",
      "@bdp/server/development-authority",
    ]) {
      try {
        consumer.resolve(specifier);
        throw new Error("packed private development runtime unexpectedly resolves");
      } catch (error) {
        if (error.code !== "ERR_PACKAGE_PATH_NOT_EXPORTED") throw error;
        stats.privateSpecifierCount++;
      }
    }

 if (await realpath(deploymentEntry) !== entryRealpath) throw new Error("public resolution differs from root entry");
}
let workError;
const work = runWitness().catch(error=>{workError=error;}).finally(()=>{workSettled=true;});
try {
 await within(work,45000,"work");
 if (workError) throw workError;
} catch(error) {
 failed=true; workTimedOut=error?.phase==="work";
 // One remaining cleanup allowance; no cancellation or FD-close fiction for a pending file operation.
 try {await within(work,cleanupRemaining(),"late-work-cleanup");} catch {cleanupUnknown=true;}
} finally {
 cleanupUnknown ||= !workSettled || Boolean(witnessGuard?.listening) ||
  (guardCloseAttempted && (!guardCloseSettled || !guardCloseSucceeded));
 if (temporaryRoot && workSettled && !cleanupUnknown) {
  try {await within(rm(temporaryRoot,{recursive:true,force:true}),cleanupRemaining(),"owned-removal");temporaryDirectoryRemoved=true;}
  catch {failed=true;cleanupUnknown=true;}
 }
 if (workSettled && !cleanupUnknown) clearTimeout(totalGuard);
 if (!workSettled || cleanupUnknown) process.exitCode=1;
}
const report = {event:"startup.bound-consumer-loader",outcome:failed||cleanupUnknown?"failed":"passed",
 ...stats,entryRealpath:entryRealpath ?? null,entrySha256:proof?.entrySha256 ?? null,provenanceSha256:proofBytes ? digest(proofBytes) : null,
 workTimedOut,cleanupUnknown,temporaryDirectory:temporaryRoot ?? null,temporaryDirectoryRemoved,
 sourceLimitation:"Root-bound dirty source/build/archive/lock/entry provenance; loader does not attest source or payload digest",
 observationLimitation:"Occupied intended port excludes successful acquisition, not bind attempts; no installed per-FD or kernel socket census; no service/public profile claim"};
console.log(JSON.stringify(report));
if (failed || cleanupUnknown) process.exitCode=1;
/** Pure installed-loader witness; this guard owns the occupied intended port.
 * Success excludes acquisition of that configured listener. Source forbidden-call
 * instrumentation separately observes attempted startup calls. Store paths stay
 * absent. This is not installed service evidence. */
async function observePackedDeploymentLoader(deployment, temporaryRoot) {
  const guard = createServer();
  witnessGuard = guard;
  await new Promise((resolve, reject) => {
    guard.once("error", reject);
    guard.listen(0, "127.0.0.1", resolve);
  });
  try {
    const address = guard.address();
    if (address === null || typeof address === "string")
      throw new Error("missing owned loader witness port");
    const scope = "https://packed-deployment.test/s/",
      token = "packed_deployment_secret_01234567890123456789";
    const file = path.join(temporaryRoot, "pure-deployment.json");
    const descriptor = `${scope}types/item`,
      schema = `${scope}schemas/item`;
    const artifacts = {
      descriptors: [
        {
          retrievalUri: descriptor,
          utf8: JSON.stringify({
            id: descriptor,
            name: "Item",
            describes: "bead",
            conformsTo: [],
            propertiesSchema: schema,
          }),
        },
      ],
      schemas: [
        {
          retrievalUri: schema,
          utf8: JSON.stringify({
            $id: schema,
            type: "object",
            properties: { n: { type: "integer" } },
            additionalProperties: false,
          }),
        },
      ],
    };
    const inert = (value) => {
      if (typeof value === "function") throw new Error("packed deployment returned a capability");
      if (value !== null && typeof value === "object") {
        if (!Object.isFrozen(value)) throw new Error("packed deployment returned mutable data");
        for (const [key, child] of Object.entries(value)) {
          if (
            [
              "credentials",
              "token",
              "authenticate",
              "metadata",
              "readOptions",
              "projectedInput",
              "installed",
              "preflight",
            ].includes(key)
          )
            throw new Error("packed deployment exposes private state");
          inert(child);
        }
      }
    };
    for (const profile of ["read-update", "transactional"]) {
      const directory = path.join(temporaryRoot, `absent-${profile}-store`);
      const configuration =
        profile === "read-update"
          ? {
              directory,
              create: true,
              credentials: [{ token, principal: "writer" }],
              policies: { writer: { create: true, write: true, hidden: [], readOnly: [] } },
              contracts: { descriptors: [], schemas: [] },
              maximumEndpointMultiplicity: [],
            }
          : {
              directory,
              create: true,
              credentials: [{ token, principal: "writer", canMutate: true }],
              artifacts,
            };
      const input = {
        deploymentFile: file,
        scope: { url: scope },
        server: {
          host: "127.0.0.1",
          port: address.port,
          advertisedProfile: profile,
          limits: {
            page: {
              defaultItems: profile === "read-update" ? 20 : 2,
              maximumItems: profile === "read-update" ? 100 : 32,
            },
            selector: { bytes: 4096, depth: 32, nodes: 256 },
            cursorTtlMilliseconds: profile === "read-update" ? 60000 : 1000,
          },
        },
      };
      const body = { format: "bdptest-reference-deployment-1", profile, configuration };
      await writeFile(file, JSON.stringify(body));
      stats.loaderCalls++;
      const result = await deployment.loadReferenceDeployment(input);
      stats.accepted++;
      inert(result);
      if (
        result.profile !== profile ||
        result.canonicalScope !== scope ||
        result.listener.port !== address.port ||
        JSON.stringify(result).includes(token)
      )
        throw new Error("packed deployment returned incorrect or secret-bearing summary");
      for (const candidate of [
        { ...body, profile: profile === "read-update" ? "transactional" : "read-update" },
        {
          ...body,
          configuration: {
            ...configuration,
            [profile === "read-update" ? "contracts" : "artifacts"]: {
              descriptors: [{ retrievalUri: descriptor, utf8: "packed_payload_sentinel" }],
              schemas: [],
            },
          },
        },
      ]) {
        await writeFile(file, JSON.stringify(candidate));
        let refused = false;
        try {
          stats.loaderCalls++;
          await deployment.loadReferenceDeployment(input);
        } catch (error) {
          if (
            !(error instanceof deployment.ReferenceDeploymentError) ||
            error.code !== "contract" ||
            "cause" in error ||
            [token, "packed_payload_sentinel"].some((secret) =>
              `${error} ${JSON.stringify(error)}`.includes(secret),
            )
          )
            throw error;
          refused = true;
        }
        if (!refused) throw new Error("packed deployment accepted an invalid contract");
        stats.contractRefusals++;
      }
      try {
        await realpath(directory);
        throw new Error("packed loader created a store directory");
      } catch (error) {
        if (error.code !== "ENOENT") throw error;
      }
      if (!guard.listening || guard.address().port !== address.port)
        throw new Error("packed loader disturbed the owned witness listener");
      stats.profiles.push(profile);
      stats.configuredPortAcquisitionExcluded = true;
      stats.storeDirectoriesAbsent = true;
    }
    console.log(
      "installed pure reference deployment loader: RU/TX, refusals and inert boundary passed",
    );
  } finally {
    guardCloseAttempted = true;
    try {
      await within(new Promise((resolve, reject) =>
        guard.close((error) => {
          guardCloseSettled = true;
          guardCloseSucceeded = !error;
          if (error) reject(error);
          else resolve();
        }),
      ), cleanupRemaining(), "cleanup");
    } catch (error) {
      cleanupUnknown = true;
      throw error;
    }
  }
}
