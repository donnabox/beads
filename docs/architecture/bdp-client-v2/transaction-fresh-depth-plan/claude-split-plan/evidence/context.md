# Exact current committed source excerpts

## packages/server/src/transaction-snapshot.test.ts:27–120
Full source SHA256 0fbdd2594147041689d81ddb7734532d01c1f4956ea0a975d10ef8c0f3458f2d
```
  beadType = `${scope}types/bead`,
  linkType = `${scope}types/link`;
const types = [
  {
    id: beadType,
    name: "Owner",
    describes: "bead",
    conformsTo: [],
    ownsOutgoing: { [linkType]: { max: 32 } },
  },
  {
    id: linkType,
    name: "Link",
    describes: "link",
    conformsTo: [],
    source: { conformsTo: [], external: "none" },
    target: { conformsTo: [], external: "none" },
  },
].map((value) => parseTypeDescriptor(value));
const cleanup: (() => void)[] = [];
afterEach(() => {
  vi.restoreAllMocks();
  for (const fn of cleanup.splice(0).reverse()) fn();
});
const create = (id: string) => ({
  operation: "createBead",
  id: `beads/${id}`,
  type: beadType,
  properties: { n: 0, precise: 0.30000000000000004 },
});
const edge = (id: string, source = "a", target = "b") => ({
  operation: "createLink",
  id: `links/${id}`,
  type: linkType,
  source: `beads/${source}`,
  target: `beads/${target}`,
  properties: { n: 0 },
});
function fixture(fault?: TransactionStoreOptions["fault"]) {
  const directory = mkdtempSync(path.join(tmpdir(), "bdp-snapshot-"));
  let now = 1000000;
  const options = {
    file: path.join(directory, "store.sqlite"),
    scope,
    types,
    create: true,
    clock: () => now,
    ...(fault ? { fault } : {}),
  };
  let owner = openTransactionStore(options),
    key = 0;
  cleanup.push(() => {
    owner.close();
    rmSync(directory, { recursive: true, force: true });
  });
  return {
    options,
    get owner() {
      return owner;
    },
    setTime(value: number) {
      now = value;
    },
    reopen() {
      owner.close();
      owner = openTransactionStore({ ...options, create: false });
    },
    offline(run: (db: DatabaseSync) => void) {
      owner.close();
      const db = new DatabaseSync(options.file);
      try {
        run(db);
      } finally {
        db.close();
      }
      owner = openTransactionStore({ ...options, create: false });
    },
    run(operations: unknown[]) {
      const admitted = owner.admit(
        "p",
        `key_${key++}`,
        prepareTransactionSetCore(scope, JSON.stringify({ operations })),
      );
      if (admitted.kind !== "owned") throw Error("admission");
      const receipt = owner.execute(admitted.claim);
      expect(receipt.status).toBe("completed");
      return receipt;
    },
  };
}
const handle = (manifest: TransactionSnapshotManifest) =>
  new URL(manifest.id).searchParams.get("snapshot")!;
function collect(
  f: ReturnType<typeof fixture>,
```

## packages/server/src/transaction-snapshot.test.ts:300–335
Full source SHA256 0fbdd2594147041689d81ddb7734532d01c1f4956ea0a975d10ef8c0f3458f2d
```
    f.reopen();
    f.offline((db) => {
      expect(db.prepare("SELECT count(*) AS n FROM snapshots").get()?.n).toBe(
        stage === "after-snapshot-copy" ? 0 : 1,
      );
      expect(db.prepare("SELECT count(*) AS n FROM resources").get()?.n).toBe(1);
      expect(db.prepare("SELECT count(*) AS n FROM groups").get()?.n).toBe(1);
      expect(db.prepare("SELECT count(*) AS n FROM receipts").get()?.n).toBe(1);
    });
    expect(f.owner.captureSnapshot().manifest.beads.items).toHaveLength(1);
  },
);
it.each([
  "missing-record",
  "changed-record",
  "oversized-record",
  "missing-range",
  "wrong-kind",
  "wrong-context",
  "wrong-anchor",
  "wrong-expiry",
])("refuses retained corruption: %s rather than rebuilding current graph", (variant) => {
  const f = fixture();
  f.run([create("a"), create("b"), edge("e")]);
  const saved = f.owner.captureSnapshot().manifest,
    id = handle(saved);
  f.offline((db) => {
    if (variant === "missing-record")
      db.prepare("DELETE FROM snapshot_records WHERE snapshot=? AND kind='link'").run(id);
    if (variant === "changed-record")
      db.prepare(
        "UPDATE snapshot_records SET body=replace(body,'0.30000000000000004','0.30000000000000005') WHERE snapshot=? AND kind='bead'",
      ).run(id);
    if (variant === "oversized-record")
      db.prepare("UPDATE snapshot_records SET body=? WHERE snapshot=? AND kind='link'").run(
        "x".repeat(262145),
```

## packages/server/src/transaction-snapshot.ts:19–34
Full source SHA256 d74aad937b0a07ea2878ef2c275cec6dbdb033cf112006e2569d64537520b15c
```
import { TRANSACTION_RUNTIME_LIMITS } from "./transaction-batch.js";
import type { TransactionReadContext } from "./transaction-store.js";

export const TRANSACTION_SNAPSHOT_LIMITS = Object.freeze({
  lifetimeMs: 5 * 60 * 1000,
  resources: 4096,
  recordBytes: 32 * 1024 * 1024,
  liveHandles: 8,
  liveRecordBytes: 64 * 1024 * 1024,
  issuedHandles: 65536,
  dispositionBytes: 1024,
  snapshotPageItems: 32,
  pageBytes: 4 * 1024 * 1024,
  manifestBytes: 10 * 1024 * 1024,
});
export const TRANSACTION_SNAPSHOT_TOKEN_POLICY = "hmac-sha256-ordinal-1";
```

## packages/server/src/transaction-snapshot.ts:585–613
Full source SHA256 d74aad937b0a07ea2878ef2c275cec6dbdb033cf112006e2569d64537520b15c
```
  function manifest(
    id: string,
    now: number,
    context: TransactionReadContext,
  ): TransactionSnapshotManifest {
    const { d, ranges } = resolve(id, now, context);
    return manifestValue(id, d, context, page(id, ranges, "bead"), page(id, ranges, "link"));
  }
  function manifestValue(
    id: string,
    d: Disposition,
    context: TransactionReadContext,
    beads: TransactionSnapshotPage,
    links: TransactionSnapshotPage,
  ): TransactionSnapshotManifest {
    const result: TransactionSnapshotManifest = {
      id: transactionSnapshotUrl(scope, id),
      scope,
      scopeEpoch: d.epoch,
      authorizationView: d.view,
      scopePosition: `pos-${d.position}`,
      checkpoint: transactionCheckpoint(scope, context, d.position),
      expiresAt: new Date(d.expires).toISOString(),
      erasures: [],
      beads,
      links,
    };
    integrity(bytes(text(result)) <= limits.manifestBytes, "serialized manifest bound");
    return freeze(result);
```

## packages/server/src/transaction-batch.ts:766–792
Full source SHA256 0f67235a7fdadf3ae342695b6f8eee795ad51b70e6507dcb7d70ec8848ba9836
```
      options,
    );
    if (result.effect === "failure")
      return {
        ok: false,
        problem: {
          ...receiptProblem(result.outcome),
          operationIndex,
          ...(op.name ? { operationName: op.name } : {}),
        },
      };
    const outcome = result.outcome;
    results.push({ operationIndex, ...(op.name ? { operationName: op.name } : {}), ...outcome });
    const resource = outcome.resource;
    if (op.name && outcome.outcome === "created" && resource) bindings.set(op.name, resource.id);
    // A retained revision is an observed no-op, not a committed transition.
    if (result.changed.length === 0 && result.deleted.length === 0) return;
    const subject = resource ?? before;
    if (!subject) throw new Error("successful transition lacks Resource state");
    // Evaluator IDs are local lookup paths; group identity is always canonical.
    // One owned operation changes at most the Link and its source Bead.
    for (const id of result.changed) {
      const row = tx.resource(id);
      if (!row) throw new Error("changed Resource lacks staged postimage");
      const record = readStoredResource(row, options.scope);
      changes.set(record.id, {
        operation: "upsert",
```

## packages/server/src/transaction-batch.ts:808–832
Full source SHA256 0f67235a7fdadf3ae342695b6f8eee795ad51b70e6507dcb7d70ec8848ba9836
```
    }
    if (outcome.outcome === "created") {
      scopeEvent(subject, "created", {
        revision: subject.revision,
        properties: subject.properties,
        metadata: subject.metadata ?? {},
        ...(subject.revisionProvenance ? { revisionProvenance: subject.revisionProvenance } : {}),
        ...("source" in subject ? { source: subject.source, target: subject.target } : {}),
      });
    } else if (outcome.outcome === "updated") {
      if (!before) throw new Error("update lacks previous state");
      scopeEvent(subject, "updated", {
        previousRevision: before.revision,
        revision: subject.revision,
        ...(input.propertiesChange !== undefined ? { change: input.propertiesChange } : {}),
        ...(input.metadataChange !== undefined ? { metadataChange: input.metadataChange } : {}),
        ...(subject.revisionProvenance ? { revisionProvenance: subject.revisionProvenance } : {}),
      });
    } else {
      scopeEvent(subject, "deleted", { revision: subject.revision });
    }
    if ("source" in subject && outcome.outcome !== "updated") {
      for (const endpoint of ["source", "target"] as const) {
        const uri = referenceUri(subject[endpoint]);
        if (!uri.startsWith(options.scope)) continue;
```

## packages/server/src/transaction-events.ts:244–272
Full source SHA256 6c5ab4482f0b00110a71a96b16d44ca5fe713cb5f397ff78778cc6646f632a3a
```
      event.source !== `${scope}events/` ||
      event.transaction !== group.transaction ||
      !["created", "updated", "deleted", "linked", "unlinked"].includes(event.type)
    )
      throw new Error("invalid retained Event framing");
    const data = event.data as Record<string, unknown>;
    let include = source.kind === "scope";
    if (visible && source.kind !== "scope") {
      include =
        event.subject === source.subject && ["created", "updated", "deleted"].includes(event.type);
      if (source.kind === "bead" && (event.type === "linked" || event.type === "unlinked")) {
        if (data.endpoint !== "source" && data.endpoint !== "target")
          throw new Error("invalid endpoint fact");
        include = reference(data[data.endpoint]) === source.subject;
      }
      if (source.kind === "bead" && event.type === "updated" && linkIds.has(event.subject)) {
        const pair = endpoints.get(event.subject);
        if (!pair) throw new Error("incident Link update lacks same-group endpoints");
        include = pair.some((endpoint) => reference(endpoint) === source.subject);
        // Literal incident rule currently mentions properties only. Do not silently
        // decide metadata-only membership while the spec disposition is pending.
        if (include && data.change === undefined && data.metadataChange !== undefined)
          throw new TransactionEventProjectionGap("incident-link-metadata");
      }
    }
    if (visible && include)
      result.push(renderTransactionEvent(context, source.url, group.checkpoint, event));
  }
  return result;
```

## packages/server/src/transaction-runtime-configuration.ts:90–125
Full source SHA256 57cd5d38c7199011d05589bc2b8ac1a0e2a4a9dc7585d019f9de5801a0a4d808
```
  mutation: TRANSACTION_RUNTIME_LIMITS,
  scopeSnapshot: {
    limits: TRANSACTION_SNAPSHOT_LIMITS,
    tokenPolicy: TRANSACTION_SNAPSHOT_TOKEN_POLICY,
  },
  receiptMaximumItems: transactionReceiptMaximum(),
  readDefaultItems: 2,
  selector: { bytes: 4096, depth: 32, nodes: 256 },
  pagination: {
    cursorTtlMs: 1000,
    retainedStateCapacity: 100,
    maxRetainedCursorPositionsPerSnapshot: 20,
    retainedSnapshotByteCapacity: 1048576,
    retainedSnapshotNodeCapacity: 100000,
    maxOpaqueTokenLength: 64,
    tokenGenerationAttempts: 3,
    idleCleanup: "on-demand" as const,
  },
  token: { bytes: 24, encoding: "base64url" as const },
  maintenanceIntervalMs: 1000,
  receiptDetailMs: TRANSACTION_RECEIPT_DETAIL_MS,
  comparisonBudgetMs: 1000,
  minimumReadBudgetMs: 1000,
  changefeedStream: TRANSACTION_SSE_DEFAULTS,
  // Owner scan/packet bounds only. Event and changefeed delivery share the
  // cadence and live-slot capacity captured in changefeedStream above.
  eventStream: TRANSACTION_EVENT_STREAM,
  transport: {
    host: "127.0.0.1",
    targetBytes: TRANSACTION_HTTP_TARGET_BYTES,
    maxHeaderSize: 16384,
    requestTimeout: 10000,
    headersTimeout: 10000,
    timeout: 10000,
    keepAliveTimeout: 1000,
    cors: "disabled",
```

## packages/server/src/transaction-runtime.ts:86–98
Full source SHA256 48cb5411407c456d7b3f301a4b67970061be9f51c94201de832ba5deff6ea2d4
```
  }
}
function deliver(response: ServerResponse, selected: HttpResponse, raw?: string): void {
  if (response.destroyed) return;
  const bytes =
    raw ?? (selected.body === undefined ? undefined : stringifyJsonValue(selected.body));
  selected.headers.set("x-bdp-development", "non-attesting");
  if (bytes !== undefined) selected.headers.set("content-length", String(Buffer.byteLength(bytes)));
  response.writeHead(selected.status, Object.fromEntries(selected.headers));
  response.end(response.req.method === "HEAD" ? undefined : bytes);
}
export async function startTransactionRuntime(
  options: TransactionRuntimeOptions,
```

## packages/server/src/development-contracts.ts:77–122
Full source SHA256 182a1a6d06b09f08ca3c83ad0fb27a09a419ad3e99729528381f210973b2c40a
```
  const captured = ingestContractArtifacts(input);
  // Reconstruct solely from captured bytes: caller mutation/accessors cannot alter compilation.
  const artifact = (kind: "descriptor" | "schema") =>
    captured.artifacts
      .filter((a) => a.kind === kind)
      .map((a) => ({
        retrievalUri: a.retrievalUri,
        utf8Bytes: Buffer.from(a.bytesBase64, "base64"),
      }));
  const bundle = {
    descriptors: artifact("descriptor"),
    schemas: artifact("schema"),
    limits: captured.receipt.limits,
  };
  const types: Record<string, string> = Object.create(null);
  const contracts = new Map<string, InstalledResourceContract>();
  for (const descriptor of captured.descriptors) {
    if (Buffer.byteLength(JSON.stringify(descriptor.id)) > 4096)
      throw new RangeError("Type diagnostic identifier bound");
    types[descriptor.id] = stringifyJsonValue(descriptor);
    let validateProperties: PropertyValidator | undefined;
    if (descriptor.propertiesSchema !== undefined) {
      const compiled = compilePrivateSchemaEvaluator({
        bundle,
        entry: { kind: "descriptor", typeId: descriptor.id },
        graphLimits: SCHEMA_GRAPH_CEILINGS,
        limits: EVALUATOR_CEILINGS,
      });
      if (compiled.kind !== "compiled")
        throw new TypeError(`administrative schema refused: ${compiled.reason}`);
      validateProperties = (properties, emitter) => {
        const outcome = compiled.evaluateUtf8(Buffer.from(stringifyJsonValue(properties)));
        if (outcome.kind !== "evaluated")
          throw new Error("installed schema evaluation did not complete");
        if (outcome.valid) {
          if (!outcome.diagnosticsComplete)
            throw new Error("installed schema evaluation did not complete");
          return { valid: true };
        }
        if (!outcome.diagnostics.length) throw new Error("invalid schema result lacks diagnostics");
        for (const diagnostic of outcome.diagnostics) {
          if (
            Buffer.byteLength(JSON.stringify(diagnostic.schemaLocation)) > 16384 ||
            Buffer.byteLength(JSON.stringify(diagnostic.message)) > 1024
          )
            throw new Error("installed diagnostic metadata exceeds qualified bound");
```

## packages/server/src/installed-schema-value.ts:1–30
Full source SHA256 ba14d1a1ef75b13ac14ea20c3066b3f8b742a05d50ac8899fd7b456fa94b067c
```
import { isProxy, isUint8Array } from "node:util/types";
import { decodeJsonDocument, JsonNumberLiteral, type LosslessJsonValue } from "@bdp/protocol";
import { decimal, compareDecimal, type Decimal } from "./installed-schema-decimal.js";
export const EVALUATOR_CEILINGS = Object.freeze({
  inputBytes: 1048576,
  valueNodes: 262144,
  states: 1048576,
  frames: 1048576,
  logicalBytes: 16777216,
  // Leading-zero-stripped coefficient width before trailing-zero normalization.
  coefficientDigits: 4096,
  normalizedCoefficientDigits: 4096,
  work: 16777216,
  occurrences: 65536,
  annotations: 1024,
  annotationBytes: 1048576,
  diagnostics: 1024,
  diagnosticBytes: 1048576,
});
export type EvaluatorLimits = { readonly [K in keyof typeof EVALUATOR_CEILINGS]: number };
export class EvaluationRefusal extends Error {
  constructor(readonly code: string) {
    super(`private schema evaluation refused: ${code}`);
  }
}
/** logicalBytes is cumulative logical allocation, never reclaimed: 64-byte value/
 * call records, 96-byte facts, 16-byte name/failure slots, 32-byte links/annotations/memo entries,
 * 64-byte coverage/candidate containers, 32-byte unique coverage entries,
 * 16-byte candidate slots, and two bytes per retained JS string unit
 * (or actual encoded diagnostic bytes).
```

## schemas/bdp-v0.schema.json:766–773
Full source SHA256 0baa5d5ef67144f8e6513b4cde31168effc8df6d52ef3141b0f144b87f8dc227
```
    "properties": {
      "type": "object",
      "additionalProperties": true
    },
    "metadata": {
      "type": "object",
      "additionalProperties": true
    },
```

## schemas/bdp-v0.schema.json:1290–1332
Full source SHA256 0baa5d5ef67144f8e6513b4cde31168effc8df6d52ef3141b0f144b87f8dc227
```
    "jsonPointer": {
      "type": "string",
      "pattern": "^(/([^/~]|~[01])*)*$"
    },
    "propertyChange": {
      "type": "array",
      "minItems": 1,
      "items": {
        "oneOf": [
          {
            "type": "object",
            "required": ["op", "path", "value"],
            "properties": {
              "op": {
                "enum": ["add", "replace"]
              },
              "path": {
                "$ref": "#/$defs/jsonPointer"
              },
              "value": true
            },
            "additionalProperties": false
          },
          {
            "type": "object",
            "required": ["op", "path"],
            "properties": {
              "op": {
                "const": "remove"
              },
              "path": {
                "$ref": "#/$defs/jsonPointer"
              }
            },
            "additionalProperties": false
          }
        ]
      }
    },
    "createBeadMembers": {
      "type": "object",
      "required": ["type"],
      "properties": {
```

## schemas/bdp-v0.schema.json:3896–4029
Full source SHA256 0baa5d5ef67144f8e6513b4cde31168effc8df6d52ef3141b0f144b87f8dc227
```
    "mutationReceipt": {
      "title": "BDP Mutation Receipt",
      "type": "object",
      "allOf": [
        {
          "$ref": "#/$defs/receiptCore"
        }
      ],
      "oneOf": [
        {
          "type": "object",
          "required": [
            "id",
            "status",
            "idempotencyKey",
            "scopeEpoch",
            "authorizationView",
            "transaction"
          ],
          "properties": {
            "id": true,
            "status": {
              "const": "pending"
            },
            "idempotencyKey": true,
            "scopeEpoch": true,
            "authorizationView": true,
            "transaction": true,
            "detail": false,
            "requiredPosition": false,
            "effectPosition": false,
            "results": false,
            "next": false,
            "allocated": false,
            "problem": false,
            "expiresAt": false
          }
        },
        {
          "type": "object",
          "required": [
            "id",
            "status",
            "detail",
            "idempotencyKey",
            "scopeEpoch",
            "authorizationView",
            "transaction",
            "requiredPosition",
            "results",
            "next",
            "expiresAt"
          ],
          "properties": {
            "id": true,
            "status": {
              "const": "completed"
            },
            "detail": {
              "const": "available"
            },
            "idempotencyKey": true,
            "scopeEpoch": true,
            "authorizationView": true,
            "transaction": true,
            "requiredPosition": true,
            "effectPosition": true,
            "results": true,
            "next": true,
            "expiresAt": true,
            "allocated": false,
            "problem": false
          },
          "allOf": [
            {
              "if": {
                "properties": {
                  "next": {
                    "type": "null"
                  }
                },
                "required": ["next"]
              },
              "then": {
                "properties": {
                  "results": {
                    "type": "array",
                    "minItems": 1
                  }
                }
              }
            }
          ]
        },
        {
          "type": "object",
          "required": [
            "id",
            "status",
            "detail",
            "idempotencyKey",
            "scopeEpoch",
            "authorizationView",
            "transaction",
            "requiredPosition",
            "problem",
            "expiresAt"
          ],
          "properties": {
            "id": true,
            "status": {
              "const": "failed"
            },
            "detail": {
              "const": "available"
            },
            "idempotencyKey": true,
            "scopeEpoch": true,
            "authorizationView": true,
            "transaction": true,
            "requiredPosition": true,
            "problem": true,
            "expiresAt": true,
            "effectPosition": false,
            "results": false,
            "next": false,
            "allocated": false
          }
        },
        {
          "type": "object",
          "required": [
            "id",
            "status",
```

## schemas/bdp-v0.schema.json:4784–4856
Full source SHA256 0baa5d5ef67144f8e6513b4cde31168effc8df6d52ef3141b0f144b87f8dc227
```
    "snapshotManifest": {
      "title": "BDP snapshot manifest",
      "type": "object",
      "required": [
        "id",
        "scope",
        "scopeEpoch",
        "authorizationView",
        "scopePosition",
        "checkpoint",
        "expiresAt",
        "erasures",
        "beads",
        "links"
      ],
      "properties": {
        "id": {
          "$ref": "#/$defs/absoluteHttpUrl"
        },
        "scope": {
          "$ref": "#/$defs/absoluteHttpUrl"
        },
        "scopeEpoch": {
          "$ref": "#/$defs/wireToken"
        },
        "authorizationView": {
          "$ref": "#/$defs/wireToken"
        },
        "scopePosition": {
          "$ref": "#/$defs/wireToken"
        },
        "checkpoint": {
          "$ref": "#/$defs/wireToken"
        },
        "expiresAt": {
          "$ref": "#/$defs/dateTime"
        },
        "erasures": {
          "type": "array",
          "items": {
            "$ref": "#/$defs/erasureRecord"
          }
        },
        "beads": {
          "allOf": [
            {
              "$ref": "#/$defs/beadCollection"
            },
            {
              "type": "object",
              "properties": {
                "items": {
                  "type": "array",
                  "items": {
                    "type": "object",
                    "properties": {
                      "links": false
                    }
                  }
                }
              }
            }
          ]
        },
        "links": {
          "$ref": "#/$defs/linkCollection"
        }
      },
      "additionalProperties": false
    },
    "transactionalAdvertisedLimits": {
      "type": "object",
      "properties": {
```

## apps/bdptest/test-support/transactional-execution-binding.ts:1–91
Full source SHA256 cde506c72a8710eafe58522900f6db70ed3701cf333560252dc8e7c5e706cf89
```
import { isDeepStrictEqual } from "node:util";
import { installedContractsDigest } from "./installed-contract-identity.js";
/** Source-only installed binding. Runner retains sole receipt observation authority. */
import { createHash } from "node:crypto";
import {
  createRawHttpScenarioTarget,
  type ExactHttpModeOptions,
  type ConformanceFixture,
  type ExecutableScenario,
} from "@bdp/conformance";
import {
  bindServedProfileConfiguration,
  assertSameServedProfileConfiguration,
  type TransactionalRunContext,
  type TransactionalSession,
} from "@bdp/conformance/testing";
export interface InstalledTransactionChild {
  readonly ready: Record<string, unknown>;
  close(): Promise<unknown>;
}
export function installedTransactionBinding(
  ready: Record<string, unknown>,
  sourceCommit: string,
  fixture: { bytes: Uint8Array },
) {
  if (
    ready.profile !== "transactional" ||
    ready.history !== false ||
    ready.claimEligible !== false ||
    ready.authorityMode !== "canonical" ||
    typeof ready.scope !== "string" ||
    typeof ready.runtimeConfigurationIdentity !== "string" ||
    !/^[a-f0-9]{64}$/.test(ready.runtimeConfigurationIdentity)
  )
    throw new Error("installed Transactional metadata unavailable");
  if (typeof ready.listenUrl !== "string")
    throw new Error("installed physical session unavailable");
  const url = new URL(ready.listenUrl);
  if (
    url.protocol !== "http:" ||
    url.hostname !== "127.0.0.1" ||
    !url.port ||
    url.username ||
    url.password ||
    url.pathname !== "/" ||
    url.search ||
    url.hash
  )
    throw new Error("invalid installed dial binding");
  const targetNames = {
    createBead: "create-bead",
    updateBead: "update-bead",
    deleteBead: "delete-bead",
    createLink: "create-link",
    updateLink: "update-link",
    deleteLink: "delete-link",
    putAlias: "put-alias",
    deleteAlias: "delete-alias",
    updateWhere: "update-where",
    deleteWhere: "delete-where",
    batch: "batch",
    sequence: "sequence",
  };
  const operationTargets = Object.freeze(
    Object.fromEntries(
      Object.entries(targetNames).map(([name, suffix]) => [
        name,
        `${ready.scope}operations/${suffix}`,
      ]),
    ),
  );
  if (!isDeepStrictEqual(ready.targets, operationTargets))
    throw new Error("installed operation target inventory mismatch");
  const served = bindServedProfileConfiguration({
    scope: ready.scope,
    profile: "transactional",
    history: false,
    sourceCommit,
    installedPayloadDigest: ready.payloadIdentity as string,
    installedContractsDigest: installedContractsDigest(ready.artifactIdentity),
    policyDigest: ready.policyIdentity as string,
    authenticationDigest: ready.authenticationIdentity as string,
    fixture,
  });
  return Object.freeze({
    served,
    operationTargets,
    runtimeConfigurationIdentity: ready.runtimeConfigurationIdentity,
    authorityMode: "canonical" as const,
  });
}
```

## apps/bdptest/src/transactional-sequence-depth-installed.test.ts:100–222
Full source SHA256 8ad2c796bfb22375584fc6b348d494ece512490f8a59ef72a74fb34de3f2fca1
```
  expect(validate(value), JSON.stringify(validate.errors)).toBe(true);
  return value;
}
const contextNames = ["bdp-scope-epoch", "bdp-authorization-view", "bdp-scope-position"] as const;
const retainedHeaders = [
  ...contextNames,
  "content-type",
  "content-length",
  "cache-control",
  "vary",
  "retry-after",
  "www-authenticate",
  "x-bdp-development",
  "x-bdp-development-limit",
];
// Independently measured literal requests; metadata has 41 nested objects.
const requests = [
  {
    index: 1,
    label: "seed",
    method: "POST",
    url: "https://reference.invalid/acme/operations/create-bead",
    key: "sequence-retained",
    accept: "application/json",
    contentType: "application/json",
    body: '{"id":"beads/sequence-a","type":"https://reference.invalid/acme/types/task","properties":{"title":"sequence-a"},"metadata":{}}',
    bodyBytes: 126,
    bodySha256: "764bc2f11265c9581533352015d0313596277a928c88556eedb32ef732ceadd1",
    status: 200,
  },
  {
    index: 2,
    label: "retained-deep",
    method: "POST",
    url: "https://reference.invalid/acme/operations/sequence",
    key: null,
    accept: "application/json",
    contentType: "application/json",
    body: '{"operations":[{"operation":"createBead","id":"beads/sequence-a","type":"https://reference.invalid/acme/types/task","properties":{"title":"sequence-a"},"metadata":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":"leaf"}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}},"idempotencyKey":"sequence-retained","name":"present"}]}',
    bodyBytes: 677,
    bodySha256: "7fd1a9710303612edcccee2ebc29d408193283f02ddc3804e75b128d30fb2d23",
    status: 200,
  },
  {
    index: 3,
    label: "original-replay",
    method: "POST",
    url: "https://reference.invalid/acme/operations/create-bead",
    key: "sequence-retained",
    accept: "application/json",
    contentType: "application/json",
    body: '{"id":"beads/sequence-a","type":"https://reference.invalid/acme/types/task","properties":{"title":"sequence-a"},"metadata":{}}',
    bodyBytes: 126,
    bodySha256: "764bc2f11265c9581533352015d0313596277a928c88556eedb32ef732ceadd1",
    status: 200,
  },
  {
    index: 4,
    label: "original-resource",
    method: "GET",
    url: "https://reference.invalid/acme/beads/sequence-a",
    key: null,
    accept: "application/json",
    contentType: null,
    body: null,
    bodyBytes: 0,
    bodySha256: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
    status: 200,
  },
  {
    index: 5,
    label: "fresh-deep",
    method: "POST",
    url: "https://reference.invalid/acme/operations/sequence",
    key: null,
    accept: "application/json",
    contentType: "application/json",
    body: '{"operations":[{"operation":"createBead","id":"beads/sequence-b","type":"https://reference.invalid/acme/types/task","properties":{"title":"sequence-b"},"metadata":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":"leaf"}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}},"idempotencyKey":"sequence-fresh","name":"present"}]}',
    bodyBytes: 674,
    bodySha256: "66f3c980ebb072c775200bda8b73f5d64176d00196eecf4de0fa69240958735e",
    status: 500,
  },
  {
    index: 6,
    label: "fresh-key-reuse",
    method: "POST",
    url: "https://reference.invalid/acme/operations/sequence",
    key: null,
    accept: "application/json",
    contentType: "application/json",
    body: '{"operations":[{"operation":"createBead","id":"beads/sequence-b","type":"https://reference.invalid/acme/types/task","properties":{"title":"sequence-b"},"metadata":{},"idempotencyKey":"sequence-fresh","name":"present"}]}',
    bodyBytes: 219,
    bodySha256: "6542f995ee8619607353fd01b5c2f3b257c9ae85ef43f0357e2393bffd81fd62",
    status: 200,
  },
  {
    index: 7,
    label: "retained-deep-negotiation",
    method: "POST",
    url: "https://reference.invalid/acme/operations/sequence",
    key: null,
    accept: "text/plain",
    contentType: "application/json",
    body: '{"operations":[{"operation":"createBead","id":"beads/sequence-a","type":"https://reference.invalid/acme/types/task","properties":{"title":"sequence-a"},"metadata":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":"leaf"}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}},"idempotencyKey":"sequence-retained","name":"present"}]}',
    bodyBytes: 677,
    bodySha256: "7fd1a9710303612edcccee2ebc29d408193283f02ddc3804e75b128d30fb2d23",
    status: 406,
  },
  {
    index: 8,
    label: "fresh-deep-negotiation",
    method: "POST",
    url: "https://reference.invalid/acme/operations/sequence",
    key: null,
    accept: "text/plain",
    contentType: "application/json",
    body: '{"operations":[{"operation":"createBead","id":"beads/sequence-c","type":"https://reference.invalid/acme/types/task","properties":{"title":"sequence-c"},"metadata":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":{"nested":"leaf"}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}},"idempotencyKey":"sequence-negotiation","name":"present"}]}',
    bodyBytes: 680,
    bodySha256: "37f093585358ce052ca89c3d8c7f4ce32e3945ff5406c0c17f942688a082be3e",
    status: 406,
  },
] as const;

```

## apps/bdptest/src/transactional-sequence-depth-installed.test.ts:319–390
Full source SHA256 8ad2c796bfb22375584fc6b348d494ece512490f8a59ef72a74fb34de3f2fca1
```
    async function exchange(index: number) {
      const row = requests[index - 1];
      if (!row) throw new Error("unbound installed sequence depth request");
      controller.signal.throwIfAborted();
      expect(attempted + 1, "installed sequence depth exact dispatch order").toBe(index);
      expect(++attempted, "installed sequence depth dispatch bound").toBeLessThanOrEqual(8);
      expect(Buffer.byteLength(row.url)).toBeLessThanOrEqual(8192);
      const response = await execute({
        method: row.method,
        url: row.url,
        signal: controller.signal,
        credentialRef: "operator",
        raw: {
          headerLines: [
            { name: "accept", value: row.accept },
            ...(row.contentType === null ? [] : [{ name: "content-type", value: row.contentType }]),
            ...(row.key === null ? [] : [{ name: "idempotency-key", value: row.key }]),
          ],
          ...(row.body === null ? {} : { bodyBytes: Buffer.from(row.body) }),
        },
      });
      // Safe response projection is retained before any status/schema assertion.
      responses.push({
        ordinal: attempted,
        label: row.label,
        method: row.method,
        url: row.url,
        accept: row.accept,
        key: row.key,
        requestBodyBytes: row.bodyBytes,
        requestBodySha256: row.bodySha256,
        status: response.status,
        bodyOctets: response.bodyOctets,
        bodySha256: digest(response.bodyText),
        wireSha256: response.wireResponseBytes ? digest(response.wireResponseBytes) : null,
        headers: Object.fromEntries(
          retainedHeaders.map((name) => [name, response.headers[name] ?? null]),
        ),
      });
      expect(response.exactRequest?.source).toBe("raw-http1-serializer");
      expect(response.exactRequest?.writeState).toBe("complete");
      expect(response.url).toBe(row.url);
      expect(response.exactRequest?.bodyPresent).toBe(row.body !== null);
      expect(response.exactRequest?.bodyOctets).toBe(row.bodyBytes);
      expect(response.exactRequest?.bodyDigest).toBe(
        row.body === null ? undefined : row.bodySha256,
      );
      const headers = response.exactRequest?.headerLines ?? [];
      const values = (name: string) =>
        headers.filter((h) => h.name.toLowerCase() === name).map((h) => h.value);
      expect(values("host")).toStrictEqual(["reference.invalid"]);
      expect(values("accept")).toStrictEqual([row.accept]);
      expect(values("content-type")).toStrictEqual(
        row.contentType === null ? [] : [row.contentType],
      );
      expect(values("idempotency-key")).toStrictEqual(row.key === null ? [] : [row.key]);
      expect(values("authorization").length).toBe(1);
      expect(
        values("authorization")[0] === `Bearer ${operator}`,
        "exact credential serialization",
      ).toBe(true);
      expect(response.wireResponseBytes?.length).toBeGreaterThan(0);
      expect(Number.isSafeInteger(response.bodyOctets)).toBe(true);
      expect(response.bodyOctets).toBeGreaterThanOrEqual(0);
      expect(response.bodyOctets).toBeLessThanOrEqual(131072);
      expect(Buffer.byteLength(response.bodyText)).toBe(response.bodyOctets);
      expect(response.headers["x-bdp-development"]).toBe("non-attesting");
      expect(response.headers["cache-control"]).toBe("private, no-store");
      if (row.method === "POST") expect(response.headers.vary).toBe("Accept");
      return response;
    }
    type Wire = Awaited<ReturnType<typeof exchange>>;
```

## apps/bdptest/src/transactional-sequence-depth-installed.test.ts:539–629
Full source SHA256 8ad2c796bfb22375584fc6b348d494ece512490f8a59ef72a74fb34de3f2fca1
```
  } finally {
    controller.abort();
    const cleanupReason = new Error("installed sequence depth cleanup deadline");
    let cleanupReject: ((reason: Error) => void) | undefined;
    const cleanupTimer = setTimeout(() => {
      cleanupExceeded = true;
      child?.child.kill("SIGKILL");
      cleanupReject?.(cleanupReason);
    }, 10000);
    try {
      if (child) {
        await Promise.race([
          child.close(),
          new Promise<never>((_, reject) => {
            // The cleanup timer also forces the owned child to stop on overrun.
            cleanupReject = reject;
          }),
        ]);
        closed = true;
      }
    } catch (error) {
      failures.push(error);
    } finally {
      clearTimeout(cleanupTimer);
    }
    const stopped =
      !launchAttempted ||
      (child !== undefined && (child.child.exitCode !== null || child.child.signalCode !== null));
    if (!stopped)
      failures.push(
        new Error(
          child === undefined
            ? "installed sequence depth child stop unproven"
            : "installed sequence depth child remains active",
        ),
      );
    try {
      if (directory && stopped) rmSync(directory, { recursive: true, force: true });
    } catch (error) {
      failures.push(error);
    }
    const directoryAbsent = directory === undefined || !existsSync(directory);
    if (stopped && !directoryAbsent)
      failures.push(new Error("installed sequence depth directory remains"));
    clearTimeout(deadline);
    if (deadlineExceeded && !failures.includes(deadlineReason)) failures.push(deadlineReason);
    try {
      console.log(
        JSON.stringify({
          observation: "private-installed-sequence-depth",
          outcome: failures.length === 0 ? "passed" : "failed",
          // Supervisor attribution to archives/consumer, never child-attested Git provenance.
          declaredSourceCommit: sourceCommit,
          workspaceDirty,
          entry: resolvedEntry,
          entrySha256,
          binding: binding ?? null,
          routeBindings,
          artifactBytes:
            artifacts.descriptors.reduce((n, d) => n + Buffer.byteLength(d.utf8), 0) +
            artifacts.schemas.reduce((n, s) => n + Buffer.byteLength(s.utf8), 0),
          attempted,
          exchanges: responses.length,
          completedResponses: responses,
          deadlineExceeded,
          cleanupExceeded,
          closed,
          launchAttempted,
          startupCompleted: child !== undefined,
          stopKnown: stopped,
          childPid: child?.child.pid ?? null,
          childExitCode: child?.child.exitCode ?? null,
          childSignalCode: child?.child.signalCode ?? null,
          directory: directory ?? null,
          directoryAbsent,
          failureCount: failures.length,
          claimEligible: false,
          retention:
            "response fingerprints and selected safe headers; full receipt/resource equality assertions ran in memory; no raw Authorization retained",
          startupLimitation:
            "Before ready, no child handle is returned; startup cleanup is launcher-owned and independent supervisor process evidence remains required",
        }),
      );
    } catch (error) {
      failures.push(error);
    }
  }
  if (failures.length === 1) throw failures[0];
  if (failures.length > 1)
    throw new AggregateError(failures, "installed sequence depth observation and cleanup failed");
}, 60000);
```

## docs/specs/bdp.md:706–720
Full source SHA256 67772742d7d08f48c12196e9789f1859ad2d0e19f133ea65e197b049361f84f8
```
derives from content verifies against bytes another canonicalized, BDP
makes the interoperability rule of
[RFC 7493 Section 2.2](https://www.rfc-editor.org/rfc/rfc7493.html#section-2.2)
mandatory: a number literal is **admissible** if and only if converting
its exact decimal value to the nearest IEEE 754 binary64 value and
serializing that value back in its shortest round-trip form — the
ECMAScript `Number::toString` form that
[RFC 8785 Section 3.2.2.3](https://www.rfc-editor.org/rfc/rfc8785.html#section-3.2.2.3)
adopts — yields a literal denoting the same exact decimal value. `1e300`
and its expanded form are admissible; `-0.0` denotes zero, is admissible,
and is the same value as `0`; `9007199254740993` and a decimal of twenty
significant digits are not admissible. On every admitted value the exact-decimal model and
the binary64 model agree by construction. A revision-token scheme that
derives tokens from content MUST declare, by name, the number model it
serializes under: `sha256-jcs` means RFC 8785 serialization with numbers
```

## docs/specs/bdp.md:6718–6728
Full source SHA256 67772742d7d08f48c12196e9789f1859ad2d0e19f133ea65e197b049361f84f8
```
source citations. Metadata alone is not executable evidence; claims require
the applicable runs under `packages/conformance/matrices/README.md`.

| Row | Obligation |
| --- | --- |
| `read-update.numeric-model.same-value-no-op` | A Property Change that rewrites a number as another admissible spelling of the same exact decimal value — `1.0` as `1` or `1e0`, `0` as `-0.0`, `1e300` as `1E300` — is a no-op that retains the revision and emits no `updated` Event, while an admissible literal of a different exact value mints a fresh revision |
| `read-update.numeric-model.round-trip-admission` | A number literal whose exact decimal value does not round-trip through binary64 — an integer past 2^53, a decimal of twenty significant digits, a magnitude beyond binary64 — is refused at admission with `validation-failed` and a diagnostic naming the member, changing nothing, while `1.0`, `-0.0`, `1e300`, and its expanded form are admitted |
| `read-update.numeric-model.nested-refusal` | The refusal applies at any depth within `properties`, whether the document is supplied whole or as a Property Change `value`, and the diagnostic names the offending input; an `instanceLocation` within an available failing document is document-relative, while an unapplied `properties` or `metadata` patch may omit it |
| `read.numeric-model.declared-token-model` | A revision-token scheme that derives tokens from content declares its number model by name; a target declaring `sha256-jcs` produces one token for every admissible spelling of the same value, over RFC 8785 bytes with numbers as binary64; honestly not applicable to a target declaring no content-derived scheme |

#### History conformance rows
```
