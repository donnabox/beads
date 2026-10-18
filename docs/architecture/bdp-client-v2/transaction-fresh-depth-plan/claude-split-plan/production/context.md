# Exact current committed source excerpts

## packages/server/src/transaction-set.ts:1–65
Full source SHA256 03b1ca9427c6963f5b6ce712a7f70540501e081ed635c742015d2bc49cc9524a
```
/** Private valid-set core. This is not a transport parser or admission policy. */
import {
  JsonNumberLiteral,
  prepareReadUpdateSingleton,
  type LosslessJsonValue,
} from "@bdp/protocol";
import { ReadSelectorError, analyzeSelectorGrammar } from "./read-selector.js";

export class TransactionDevelopmentLimitError extends Error {}
export type TransactionSetOperation = "updateWhere" | "deleteWhere";
export interface PreparedTransactionSet {
  readonly operation: TransactionSetOperation;
  readonly input: Readonly<Record<string, LosslessJsonValue>>;
  readonly name?: never;
}
export const TRANSACTION_SELECTOR_LIMITS = Object.freeze({ bytes: 4096, depth: 32, nodes: 256 });
/** Serialize already decoded, byte-bounded syntax without prematurely applying
 * the executable depth limit. Preserve original numeric lexemes and key order. */
export function transactionInputText(value: LosslessJsonValue): string {
  const pending: (string | { value: LosslessJsonValue })[] = [{ value }];
  const chunks: string[] = [];
  while (pending.length) {
    const task = pending.pop()!;
    if (typeof task === "string") {
      chunks.push(task);
      continue;
    }
    const item = task.value;
    if (item instanceof JsonNumberLiteral) chunks.push(item.literal);
    else if (Array.isArray(item)) {
      chunks.push("[");
      pending.push("]");
      for (let i = item.length - 1; i >= 0; i--) {
        pending.push({ value: item[i]! });
        if (i > 0) pending.push(",");
      }
    } else if (item && typeof item === "object") {
      chunks.push("{");
      pending.push("}");
      const entries = Object.entries(item);
      for (let i = entries.length - 1; i >= 0; i--) {
        const [key, child] = entries[i]!;
        pending.push({ value: child }, ":", JSON.stringify(key));
        if (i > 0) pending.push(",");
      }
    } else chunks.push(JSON.stringify(item));
  }
  return chunks.join("");
}
/** Private execution capacity only, after complete known carrier syntax. The
 * depth convention is unchanged: the batch root is 0, singleton input is 2. */
export function assertTransactionInputDepth(value: LosslessJsonValue, depth = 0): void {
  const pending = [{ value, depth }];
  while (pending.length) {
    const item = pending.pop()!;
    if (item.depth > 32)
      throw new TransactionDevelopmentLimitError("development batch depth limit");
    if (item.value && typeof item.value === "object" && !(item.value instanceof JsonNumberLiteral))
      for (const child of Object.values(item.value))
        pending.push({ value: child, depth: item.depth + 1 });
  }
}
function nonnegativeInteger(value: LosslessJsonValue): boolean {
  if (!(value instanceof JsonNumberLiteral)) return false;
  const match = /^(-?)(\d+)(?:\.(\d+))?(?:[eE]([+-]?\d+))?$/.exec(value.literal);
```

## packages/server/src/resource-evaluator.ts:383–532
Full source SHA256 89606787e455f4ed555b807d91c5331a376431807a9a581de1f94e538fcab4ff
```
function equal(left: unknown, right: unknown): boolean {
  const pending: Array<readonly [unknown, unknown]> = [[left, right]];
  while (pending.length) {
    const pair = pending.pop();
    if (!pair) break;
    const [a, b] = pair;
    if (a === b) continue;
    if (
      !a ||
      !b ||
      typeof a !== "object" ||
      typeof b !== "object" ||
      Array.isArray(a) !== Array.isArray(b)
    )
      return false;
    const aa = a as Record<string, unknown>;
    const bb = b as Record<string, unknown>;
    const keys = Object.keys(aa);
    if (keys.length !== Object.keys(bb).length) return false;
    for (const key of keys) {
      if (!Object.hasOwn(bb, key)) return false;
      pending.push([aa[key], bb[key]]);
    }
  }
  return true;
}
function copy(value: unknown): unknown {
  if (value === null || typeof value !== "object") return value;
  const result: unknown[] | Record<string, unknown> = Array.isArray(value)
    ? []
    : (Object.create(null) as Record<string, unknown>);
  const pending = [{ source: value, target: result }];
  while (pending.length) {
    const entry = pending.pop();
    if (!entry) break;
    for (const [key, child] of Object.entries(entry.source)) {
      const next =
        child && typeof child === "object"
          ? Array.isArray(child)
            ? []
            : (Object.create(null) as Record<string, unknown>)
          : child;
      Object.defineProperty(entry.target, key, {
        value: next,
        enumerable: true,
        configurable: true,
        writable: true,
      });
      if (child && typeof child === "object")
        pending.push({ source: child, target: next as unknown[] | Record<string, unknown> });
    }
  }
  return result;
}
function freeze<T>(value: T): T {
  const pending: unknown[] = [value];
  while (pending.length) {
    const item = pending.pop();
    if (item && typeof item === "object" && !Object.isFrozen(item)) {
      for (const child of Object.values(item)) pending.push(child);
      Object.freeze(item);
    }
  }
  return value;
}
function patch(
  properties: Readonly<Record<string, unknown>>,
  changes: PropertyChange,
  document: "properties" | "metadata",
  options: ResourceEvaluationOptions,
): Readonly<Record<string, unknown>> {
  function invalid(message: string, pointer: string): never {
    return bad(
      document === "metadata" ? `metadata: ${message}` : message,
      options,
      pointer,
      document === "metadata" ? document : undefined,
    );
  }
  let result = copy(properties);
  for (const change of changes) {
    const tokens =
      change.path === ""
        ? []
        : change.path
            .slice(1)
            .split("/")
            .map((token) => token.replace(/~1/g, "/").replace(/~0/g, "~"));
    if (tokens.length === 0) {
      if (change.op !== "add" && result === undefined)
        invalid("Property Change target does not exist", change.path);
      if (change.op === "remove") result = undefined;
      else result = copy(change.value);
      continue;
    }
    let parent = result;
    for (const token of tokens.slice(0, -1)) {
      if (
        !parent ||
        typeof parent !== "object" ||
        !Object.hasOwn(parent, token) ||
        (Array.isArray(parent) && !/^(0|[1-9][0-9]*)$/.test(token))
      )
        invalid("Property Change parent does not exist", change.path);
      parent = (parent as Record<string, unknown>)[token];
    }
    const key = tokens[tokens.length - 1];
    if (key === undefined || !parent || typeof parent !== "object")
      invalid("Property Change target does not exist", change.path);
    if (Array.isArray(parent)) {
      const index =
        key === "-" && change.op === "add"
          ? parent.length
          : /^(0|[1-9][0-9]*)$/.test(key)
            ? Number(key)
            : NaN;
      if (
        !Number.isSafeInteger(index) ||
        index < 0 ||
        index > parent.length ||
        (change.op !== "add" && index === parent.length)
      )
        invalid("Property Change array index does not exist", change.path);
      if (change.op === "add") parent.splice(index, 0, copy(change.value));
      else if (change.op === "remove") parent.splice(index, 1);
      else parent[index] = copy(change.value);
    } else {
      if (change.op !== "add" && !Object.hasOwn(parent, key))
        invalid("Property Change target does not exist", change.path);
      if (change.op === "remove") delete (parent as Record<string, unknown>)[key];
      else
        Object.defineProperty(parent, key, {
          value: copy(change.value),
          enumerable: true,
          configurable: true,
          writable: true,
        });
    }
  }
  if (!result || typeof result !== "object" || Array.isArray(result))
    bad(
      `resulting ${document} must be an object`,
      options,
      "",
      document === "metadata" ? document : undefined,
    );
  return freeze(result as Record<string, unknown>);
}

/** One synchronous member; the owning transaction commits returned outcome and
```

## packages/server/src/resource-evaluator.ts:1027–1060
Full source SHA256 89606787e455f4ed555b807d91c5331a376431807a9a581de1f94e538fcab4ff
```
    fail("revision-mismatch");
  if (deleting && resourceKind === "bead" && tx.incidentLinks(id).length)
    fail("incident-links-exist");
  if ("propertiesChange" in input || "metadataChange" in input) {
    const limits = options.limits;
    const changes = [
      ...("propertiesChange" in input ? (input.propertiesChange ?? []) : []),
      ...("metadataChange" in input ? (input.metadataChange ?? []) : []),
    ];
    if (limits.patchOperations !== undefined && changes.length > limits.patchOperations)
      fail("limit-exceeded");
    for (const change of changes)
      if (
        (limits.patchPathBytes !== undefined &&
          Buffer.byteLength(change.path) > limits.patchPathBytes) ||
        (limits.patchPathDepth !== undefined &&
          (change.path === "" ? 0 : change.path.split("/").length - 1) > limits.patchPathDepth)
      )
        fail("limit-exceeded");
  }
  const writes: EvaluatorStoredResource[] = [];
  for (const record of [after, sourceAfter]) {
    if (!record) continue;
    const bodyJson = stringifyJsonValue(record);
    if (
      (options.limits.propertiesBytes !== undefined &&
        Buffer.byteLength(stringifyJsonValue(record.properties)) >
          options.limits.propertiesBytes) ||
      (options.limits.representationBytes !== undefined &&
        Buffer.byteLength(bodyJson) > options.limits.representationBytes)
    )
      fail("limit-exceeded");
    if (!noOp) {
      const stored = { id: local(scope, record.id) as string, bodyJson };
```

## packages/server/src/transaction-batch.ts:64–77
Full source SHA256 0f67235a7fdadf3ae342695b6f8eee795ad51b70e6507dcb7d70ec8848ba9836
```
  maximumOperations: 32,
  inducedEvents: 256,
  matchedResources: 32,
  representationBytes: 256 * 1024,
  scopeSourceBytes: 64 * 1024,
  requestBodyBytes: 64 * 1024,
  propertiesBytes: 64 * 1024,
  diagnosticBytes: 512 * 1024,
  diagnosticCount: 16,
});
export const TRANSACTION_MAXIMUM_OPERATIONS: number = TRANSACTION_RUNTIME_LIMITS.maximumOperations;
snapshotReadUpdateRuntimeLimits(TRANSACTION_RUNTIME_LIMITS);

export type TransactionSingletonOperation = ReadUpdateOperation | TransactionSetOperation;
```

## packages/server/src/transaction-batch.ts:259–273
Full source SHA256 0f67235a7fdadf3ae342695b6f8eee795ad51b70e6507dcb7d70ec8848ba9836
```
export function prepareTransactionSequence(
  scope: string,
  text: string,
): PreparedTransactionSequence {
  const sequence = prepareSequence(scope, text);
  for (const member of sequence.members) requireTransactionDepth(member.request);
  return sequence;
}
/** Private receiving syntax; retained member identity precedes fresh capacity. */
export function prepareTransactionSequenceForReceiving(
  scope: string,
  text: string,
): PreparedTransactionSequence {
  return prepareSequence(scope, text);
}
```

## packages/server/src/transaction-batch.ts:339–387
Full source SHA256 0f67235a7fdadf3ae342695b6f8eee795ad51b70e6507dcb7d70ec8848ba9836
```
      deferred: dependencies.length > 0 || aliasOperands.length > 0,
    });
  });
  const result = Object.freeze({ scope: carrier.scope, members: Object.freeze(members) });
  sequences.add(result);
  return result;
}
export function assertPreparedTransactionSequence(
  sequence: PreparedTransactionSequence,
  scope: string,
): void {
  if (!sequences.has(sequence) || sequence.scope !== scope)
    throw new TypeError("expected prepared transaction sequence for this Scope");
}
function prepareCarrier(carrier: PreparedTransaction["carrier"]): PreparedTransaction {
  const result = Object.freeze({ carrier });
  prepared.add(result);
  return result;
}
function prepareDepthCarrier(
  carrier: PreparedTransaction["carrier"],
  input: LosslessJsonValue,
  depth = 0,
): PreparedTransaction {
  const result = prepareCarrier(carrier);
  try {
    assertTransactionInputDepth(input, depth);
  } catch (error) {
    if (!(error instanceof TransactionDevelopmentLimitError)) throw error;
    depthRefusals.set(result, error.message);
  }
  return result;
}
/** Source-private, read-only fact; only this module can create or annotate its brand. */
export function transactionDepthRefusal(batch: PreparedTransaction): string | undefined {
  if (!prepared.has(batch)) throw new TypeError("expected prepared transaction");
  return depthRefusals.get(batch);
}
function requireTransactionDepth(batch: PreparedTransaction): PreparedTransaction {
  const message = transactionDepthRefusal(batch);
  if (message !== undefined) throw new TransactionDevelopmentLimitError(message);
  return batch;
}
export function assertPreparedTransaction(batch: PreparedTransaction, scope: string): void {
  if (!prepared.has(batch) || batch.carrier.scope !== scope)
    throw new TypeError("expected prepared transaction for this Scope");
}
/** Source-private permitted alias slots shared by preparation and pending validation. */
export function transactionAliasFields(operation: string): readonly string[] {
```

## packages/server/src/transaction-store.ts:1133–1198
Full source SHA256 bc10df9d02316e395b8c4ea7db6134ab2646c8e8abd83ea785b0ea09f2f0c82c
```
  function admitKey(
    principal: string,
    key: string,
    batch: PreparedTransaction,
    pendingWins = false,
    references?: ReferenceBindings,
    template?: DependencyTemplate,
    controls?: TransactionReceivingControls,
  ): TransactionReceivingAdmission {
    const stored = row(principal, key);
    const prior = stored ? expireRow(stored, instant()).row : undefined;
    if (pendingWins && prior?.owner !== null && prior !== undefined)
      return { kind: "existing", receipt: parseReceipt(prior.receipt) };
    if (prior?.dependency) {
      const template = parseDependencyTemplate(prior.dependency, scope);
      const comparison = compareDependencyTemplate(template, batch, resolveAlias, (index) => {
        const stored = db
          .prepare("SELECT fact FROM creator_facts WHERE attempt=? AND member=?")
          .get(template.attempt, index);
        if (!stored) throw new Error("unresolved receipt lost creator fact");
        return typeof stored.fact === "string"
          ? (JSON.parse(stored.fact) as CreatorFact)
          : undefined;
      });
      return comparison === "conflict"
        ? { kind: "conflict" }
        : comparison === "equal"
          ? { kind: "existing", receipt: parseReceipt(prior.receipt) }
          : { kind: "comparison-pending" };
    }
    const depthRefusal = transactionDepthRefusal(batch);
    if (!prior && depthRefusal !== undefined) {
      if (controls === undefined) throw new TransactionDevelopmentLimitError(depthRefusal);
      return { kind: "development-limit", message: depthRefusal };
    }
    const normalized = template
      ? undefined
      : normalizeTransaction(batch, resolveAlias, prior?.witnesses, references);
    if (prior) {
      if (!normalized) throw new Error("existing key cannot acquire a new unresolved template");
      return prior.identity !== normalized.identity
        ? { kind: "conflict" }
        : { kind: "existing", receipt: parseReceipt(prior.receipt) };
    }
    // Prior-key equality/conflict wins above. Only an unknown key pays fresh
    // admission controls, and no refusal ever inserts a receipt or owner.
    if (controls && batch.carrier.operations.length > controls.maximumOperations)
      return { kind: "operation-limit" };
    if (controls) {
      for (const [index, op] of batch.carrier.operations.entries()) {
        if (op.operation !== "updateWhere" && op.operation !== "deleteWhere") continue;
        const limit = transactionSelectorCapacity(op);
        if (limit)
          return {
            kind: "selector-limit",
            limit,
            pointer:
              batch.carrier.kind === "singleton" ? "/selector" : `/operations/${index}/selector`,
          };
      }
    }
    if (controls && !controls.acceptsJson) return { kind: "not-acceptable" };
    const receipt: TransactionReceipt = {
      id: `${scope}receipts/${randomUUID()}`,
      status: "pending",
      idempotencyKey: key,
```

## packages/server/src/transaction-store.ts:1340–1404
Full source SHA256 bc10df9d02316e395b8c4ea7db6134ab2646c8e8abd83ea785b0ea09f2f0c82c
```
            } finally {
              active = false;
            }
          })();
          options.fault?.("after-effects");
          // The minimum starts when evaluation has become terminal, not at
          // admission or before Resource callbacks have consumed execution time.
          const expiresAt = new Date(instant() + day).toISOString();
          let receipt: TransactionReceipt;
          if (!result.ok) {
            db.exec("ROLLBACK TO effects; RELEASE effects");
            receipt = {
              ...core,
              status: "failed",
              detail: "available",
              requiredPosition: `pos-${previous}`,
              problem: result.problem,
              expiresAt,
            };
          } else {
            db.exec("RELEASE effects");
            const group = {
              scopeEpoch: epoch,
              authorizationView: view,
              checkpoint: position,
              position,
              previousPosition: `pos-${previous}`,
              projectionAdvance: false,
              transaction: core.transaction,
              changes: result.changes,
              erasures: [],
              eventCount: result.events.length,
              events: result.events,
            };
            const effective = result.events.length > 0;
            if (effective) {
              db.prepare("INSERT INTO groups VALUES(?,?)").run(previous + 1, json(group));
              db.prepare("UPDATE metadata SET value=? WHERE name='position'").run(
                String(previous + 1),
              );
            }
            receipt = {
              ...core,
              status: "completed",
              detail: "available",
              requiredPosition: effective ? position : `pos-${previous}`,
              ...(effective ? { effectPosition: position } : {}),
              results: result.results,
              next: null,
              expiresAt,
            };
          }
          const changed = db
            .prepare(
              "UPDATE receipts SET receipt=?,owner=NULL WHERE principal=? AND key=? AND owner=?",
            )
            .run(json(receipt), owned.principal, owned.key, owned.owner);
          if (changed.changes !== 1) throw new Error("terminal ownership lost");
          if (owned.creator)
            writeFact(
              owned.creator.attempt,
              owned.creator.member,
              creatorFact(owned.creator.operation, receipt),
            );
          if (receipt.status === "completed") createPagePlan(owned.principal, owned.key, receipt);
```

## packages/server/src/transaction-store.ts:1478–1518
Full source SHA256 bc10df9d02316e395b8c4ea7db6134ab2646c8e8abd83ea785b0ea09f2f0c82c
```
  ): TransactionReceivingAdmission;
  function admit(
    principal: string,
    key: string,
    batch: PreparedTransaction,
    controls: TransactionReceivingControls | undefined,
    observed: true,
  ): TransactionReceivingAdmission;
  function admit(
    principal: string,
    key: string,
    batch: PreparedTransaction,
    controls?: TransactionReceivingControls,
    observed = false,
  ): TransactionReceivingAdmission {
    const receiving = snapshotTransactionReceivingControls(
      controls ?? { acceptsJson: true, maximumOperations: TRANSACTION_MAXIMUM_OPERATIONS },
    );
    return enter(() => {
      assertPreparedTransaction(batch, scope);
      checkPrincipalKey(principal, key);
      if (mutationPause)
        throw new TransactionAuthorizationFailure("temporarily-unavailable", readContext());
      if (!authorization.canMutate)
        throw new TransactionAuthorizationFailure("forbidden", readContext());
      const admission = atomic(() => {
        const result = admitKey(principal, key, batch, false, undefined, undefined, receiving);
        if (result.kind === "development-limit" && controls === undefined)
          throw new TransactionDevelopmentLimitError(result.message);
        if (result.kind === "operation-limit" && controls === undefined)
          throw new TransactionDevelopmentLimitError("development transaction operation limit");
        if (result.kind === "selector-limit" && controls === undefined)
          throw new TransactionDevelopmentLimitError("development transaction Selector limit");
        if (result.kind === "existing") {
          if (!receiving.acceptsJson)
            return {
              kind: "not-acceptable" as const,
              ...(observed ? { context: readContext() } : {}),
            };
          if (observed) return { kind: "existing" as const, ...observeReceipt(result.receipt) };
        }
```

## packages/server/src/transaction-submission.ts:25–60
Full source SHA256 4336e45d7b4f90eec685b70c7ba059219954bb20271aa2c942ee9e140f0743af
```
export interface TransactionSubmission {
  readonly carrier:
    | { readonly kind: "batch" }
    | { readonly kind: "singleton"; readonly operation: TransactionSingletonOperation };
  readonly body: string;
  readonly key: string;
  /** Trusted authentication/Scope permission decision, not client attribution. */
  readonly authorization: Authorization;
  readonly receiving: TransactionReceivingControls;
}
type ObservedComparisonRefusal = {
  readonly kind: "comparison-refusal";
  readonly status: 503;
  readonly code: "temporarily-unavailable";
  readonly retry: "after-delay";
} & (
  | { readonly reason: "deadline"; readonly context: TransactionReadContext }
  | { readonly reason: "owner-stopped" }
);
export type TransactionSubmissionResult =
  | ObservedComparisonRefusal
  | {
      readonly kind: "receipt";
      readonly receipt: TransactionReceipt;
      readonly context?: TransactionReadContext;
    }
  | { readonly kind: "not-acceptable"; readonly context?: TransactionReadContext }
  | { readonly kind: "development-limit"; readonly message: string }
  | {
      readonly kind: "refused";
      readonly status: 400 | 401 | 403 | 409 | 413 | 503;
      readonly code:
        | "malformed-request"
        | "unauthenticated"
        | "forbidden"
        | "idempotency-conflict"
```

## packages/server/src/transaction-submission.ts:125–155
Full source SHA256 4336e45d7b4f90eec685b70c7ba059219954bb20271aa2c942ee9e140f0743af
```
    prepared,
    {
      ...comparison,
      receiving,
      // Override even an untyped caller's extra field; only this owner supplies it.
      observeDeadline: observed ? store.observeComparisonDeadline : undefined,
    },
  );
  if (
    observed &&
    (result.kind === "conflict" ||
      result.kind === "operation-limit" ||
      result.kind === "selector-limit" ||
      result.kind === "not-acceptable") &&
    !result.context
  )
    throw new Error("admission refusal lost same-entry context");
  switch (result.kind) {
    case "development-limit":
      return result;
    case "conflict":
      return {
        kind: "refused",
        status: 409,
        code: "idempotency-conflict",
        ...(result.context ? { context: result.context } : {}),
      };
    case "operation-limit":
      return {
        kind: "refused",
        status: 413,
```

## packages/server/src/transaction-http.ts:320–381
Full source SHA256 4143e8a36021d60d1deba4649b1de5dfbcb3ee2d630486749932eccc1374c602
```
        }
        if (authorization.kind === "unauthenticated")
          return transactionHttpProblem("unauthenticated");
        if (!authorization.canMutate) return transactionHttpProblem("forbidden");
        if (carrier.operations.length > TRANSACTION_MAXIMUM_OPERATIONS) {
          const observed = await service.sequenceRefusalObserved();
          return context(
            problem({ ...readProblem("limit-exceeded"), limit: "sequence.operations" }),
            observed.context,
          );
        }
        const prepared = prepareTransactionSequenceForReceiving(scope, received.text);
        if (!acceptsJson(request.headers.get("accept"))) {
          const observed = await service.sequenceRefusalObserved();
          return context(empty(406), observed.context);
        }
        try {
          const observed = await service.sequenceObserved(authorization.principal, prepared);
          return context(jsonResponse(observed.body), observed.context);
        } catch (error) {
          if (error instanceof TransactionDevelopmentLimitError)
            return developmentGap("unsupported-transaction-composition");
          throw error;
        }
      }
      const decision: TransactionSubmission["authorization"] =
        authorization.kind === "unauthenticated"
          ? authorization
          : authorization.canMutate
            ? { kind: "authorized", principal: authorization.principal }
            : { kind: "forbidden" };
      const result = await service.submit({
        carrier:
          operation === "batch"
            ? { kind: "batch" }
            : { kind: "singleton", operation: operation as TransactionSingletonOperation },
        body: received.text,
        key: request.headers.get("idempotency-key") ?? "",
        authorization: decision,
        receiving: { acceptsJson: acceptsJson(request.headers.get("accept")), maximumOperations },
      });
      if (result.kind === "receipt") return receiptResponse(result, true);
      if (result.kind === "comparison-refusal") {
        const response = problem(readProblem(result.code));
        if (result.reason === "owner-stopped") return response;
        if (!result.context) throw new Error("comparison deadline lost same-entry context");
        return context(response, result.context);
      }
      if (result.kind === "not-acceptable") {
        if (!result.context) throw new Error("negotiation refusal lost same-entry context");
        return context(empty(406), result.context);
      }
      if (result.kind === "development-limit")
        return developmentGap("unsupported-transaction-composition");
      if (result.code === "idempotency-conflict") {
        if (!result.context) throw new Error("conflict refusal lost same-entry context");
        return context(
          problem(
            parseReadUpdateProblem({
              type: "https://github.com/gastownhall/bdp/problems/conflict",
              code: result.code,
              status: 409,
```

## packages/server/src/transaction-snapshot.ts:361–425
Full source SHA256 d74aad937b0a07ea2878ef2c275cec6dbdb033cf112006e2569d64537520b15c
```
  function checkGraph(
    id: string,
    kind: Kind,
    value: ResourceRecord,
    records: ReadonlyMap<string, { body: string; value: ResourceRecord }>,
    outgoing: ReadonlyMap<string, readonly LinkRecord[]>,
  ) {
    const descriptor = descriptors.get(value.type);
    integrity(descriptor?.describes === kind, "unknown or wrong-kind Type");
    if (kind === "bead") {
      const bead = value as BeadRecord;
      integrity(descriptor.describes === "bead", "bead descriptor");
      const expected = new Map<string, string[]>();
      for (const name of Object.keys(descriptor.ownsOutgoing ?? {}))
        if (name !== "*") expected.set(name, []);
      for (const link of descriptor.ownsOutgoing === undefined ? [] : (outgoing.get(id) ?? [])) {
        if (
          descriptor.ownsOutgoing &&
          (Object.hasOwn(descriptor.ownsOutgoing, link.type) ||
            Object.hasOwn(descriptor.ownsOutgoing, "*"))
        ) {
          const list = expected.get(link.type) ?? [];
          list.push(link.id);
          expected.set(link.type, list);
        }
      }
      if (descriptor.ownsOutgoing === undefined)
        integrity(bead.ownedLinks === undefined, "unexpected owned plane");
      else {
        integrity(
          bead.ownedLinks &&
            Object.keys(bead.ownedLinks).sort().join("\n") ===
              [...expected.keys()].sort().join("\n"),
          "owned plane Types",
        );
        for (const [type, ids] of expected) {
          const inline = bead.ownedLinks[type]!;
          integrity(
            isDeepStrictEqual(
              inline.map((link) => link.id),
              ids.sort(compareCanonicalIds),
            ),
            "owned inventory mismatch",
          );
          for (const link of inline)
            integrity(
              isDeepStrictEqual(link, records.get(link.id.slice(scope.length))?.value) &&
                referenceUri(link.source) === bead.id,
              "owned Link differs from first-class record",
            );
        }
      }
    } else {
      const link = value as LinkRecord;
      for (const endpoint of [link.source, link.target])
        if (validateServerEndpoint(endpoint, scope)) {
          integrity(
            records
              .get(referenceUri(endpoint).slice(scope.length))
              ?.value.id.startsWith(`${scope}beads/`),
            "missing in-Scope endpoint",
          );
        }
    }
  }
```

## packages/server/src/transaction-snapshot.ts:646–706
Full source SHA256 d74aad937b0a07ea2878ef2c275cec6dbdb033cf112006e2569d64537520b15c
```
    const ids = visible()
      ? statement("SELECT id,kind FROM resources LIMIT ?").all(limits.resources + 1)
      : [];
    integrity(ids.length === Number(totals.count), "authority inventory changed inside owner");
    ids.sort((a, b) => compareCanonicalIds(String(a.id), String(b.id)));
    // The nonallocating path uses fixed-width layout witnesses, never an issued
    // identity/secret. No record/page/parent row is written in that path.
    const id = retain ? randomUUID() : "00000000-0000-0000-0000-000000000000",
      seed = retain ? randomBytes(32).toString("hex") : "0".repeat(64),
      counts: [number, number] = [0, 0];
    // One bounded decoded/serialized map and outgoing index per capture.
    // This replaces repeated Link codecs and per-Bead full-table scans. The
    // 32MiB encoded ceiling is not a promise about decoded-object heap size.
    const records = new Map<string, { body: string; value: ResourceRecord }>();
    const outgoing = new Map<string, LinkRecord[]>();
    let totalBytes = 0;
    for (const row of ids) {
      integrity(
        typeof row.id === "string" && (row.kind === "bead" || row.kind === "link"),
        "authority identity/kind",
      );
      const captured = record(row.id, row.kind);
      totalBytes += bytes(captured.body);
      if (totalBytes > limits.recordBytes) throw new TransactionSnapshotFailure("limit-exceeded");
      records.set(row.id, captured);
      if (row.kind === "link") {
        const link = captured.value as LinkRecord;
        const source = referenceUri(link.source);
        if (validateServerEndpoint(link.source, scope)) {
          const local = source.slice(scope.length);
          const links = outgoing.get(local) ?? [];
          links.push(link);
          outgoing.set(local, links);
        }
      }
    }
    const streams: Record<Kind, { body: string; value: ResourceRecord }[]> = { bead: [], link: [] };
    for (const row of ids) {
      const localId = String(row.id),
        kind = row.kind as Kind;
      const { body, value } = records.get(localId)!;
      checkGraph(localId, kind, value, records, outgoing);
      const ordinal = counts[kind === "bead" ? 0 : 1]++;
      streams[kind].push({ body, value });
      // Retained capture inserts the sealed parent last, after ranges exist.
      // This entire body is one owner transaction with no integrity reentry;
      // transient detail without a parent can never be externally observed.
      if (retain)
        statement("INSERT INTO snapshot_records VALUES(?,?,?,?,?)").run(
          id,
          kind,
          ordinal,
          value.id,
          body,
        );
    }
    if (live.reduce((n, row) => n + row.d.bytes, 0) + totalBytes > limits.liveRecordBytes)
      throw new TransactionSnapshotFailure(
        "temporarily-unavailable",
        live.length === 0
          ? 0
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

## packages/server/src/installed-schema-evaluator.ts:989–1020
Full source SHA256 20662b8fb5457d4233e2cc22e35152aa43a2033b7db3568d42847a8752774544
```
  const stack: { key: string; iterator: Generator<Request, Fact, Fact>; next?: Fact }[] = [];
  const push = (node: number, instance: number, key: string): void => {
    b.charge("states");
    b.charge("logicalBytes", 64);
    b.bound("frames", stack.length + 1);
    stack.push({ key, iterator: run(node, instance) });
  };
  const entry = specialization?.entry ?? root;
  push(entry, 0, memo.key(entry, 0));
  let result: Fact | undefined;
  while (stack.length) {
    b.work();
    const frame = stack[stack.length - 1];
    if (!frame) throw new Error("evaluation frame");
    const step = frame.iterator.next(frame.next as Fact);
    delete frame.next;
    if (step.done) {
      memo.set(frame.key, step.value);
      stack.pop();
      if (stack.length) (stack[stack.length - 1] as typeof frame).next = step.value;
      else result = step.value;
    } else {
      const key = memo.key(step.value.node, step.value.instance);
      const hit = memo.get(key);
      if (hit) frame.next = hit;
      else push(step.value.node, step.value.instance, key);
    }
  }
  if (!result) throw new Error("evaluation result");
  const diagnostics: Diagnostic[] = [],
    annotations: Annotation[] = [];
  let diagnosticsComplete = true;
```

## packages/protocol/src/read-update-values.ts:156–180
Full source SHA256 d1fdbc660d5e2c009275c152bdcebda46031700920aa03bfea89f39564c89de2
```
/** Internal shared schema/snapshot boundary for RU response parsers. Object
 * inputs have already lost source spelling; raw requests MUST use text APIs.
 * The iterative snapshot adds no wire depth limit; admission owns byte/work bounds.
 */
export function parseReadUpdateShape(
  value: unknown,
  key: keyof typeof READ_UPDATE_VALUE_SCHEMA_REFS,
): Readonly<Record<string, unknown>> {
  const record = snapshotProtocolRecord(value, key);
  const pending: unknown[] = [record];
  while (pending.length) {
    const item = pending.pop();
    if (typeof item === "string" && !isUnicodeScalarString(item))
      throw new ProtocolArtifactValidationError("RU value contains an unpaired surrogate");
    if (item && typeof item === "object")
      for (const [name, value] of Object.entries(item)) {
        if (!isUnicodeScalarString(name))
          throw new ProtocolArtifactValidationError(
            "RU member name contains an unpaired surrogate",
          );
        pending.push(value);
      }
  }
  validate(record, key);
  return record;
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

## docs/specs/bdp.md:1043–1070
Full source SHA256 67772742d7d08f48c12196e9789f1859ad2d0e19f133ea65e197b049361f84f8
```

The Read discovery document's `limits` uses the closed `advertisedLimits`
definition. Its Read-applicable groups are `page`, `request`, `resource`,
`selector`, and `retention.maximumSnapshotLifetime`; the schema admits no
`validation` group. Schema acceptance of a limit does not expose its capability.

The discovery document MAY contain a `limits` object. The object is optional
so that a small implementation can expose a conforming profile without
predicting every operational bound. Omission means only that the bound is not
pre-advertised; it does not mean infinite capacity and does not permit silent
truncation or a non-normative failure response.

When present, `limits` is divided into capability groups. A group is relevant
only when the advertised profile exposes that capability. Each advertised
value is a binding positive integer or ISO 8601 duration:

- `page.defaultItems` and `page.maximumItems` count Resource records;
- `request.targetBytes` counts octets in the encoded HTTP request target, and
  `request.bodyBytes` counts octets in the representation body;
- `resource.representationBytes` and `resource.propertiesBytes` count UTF-8
  bytes in the corresponding JSON serialization; metadata is covered by the
  complete representation bound, not by the properties-only bound;
- `selector.bytes` counts UTF-8 bytes of the decoded Selector string: after
  percent-decoding for query parameters, or JSON string decoding for a
  Selector in a request body, without additional URL decoding;
  `selector.depth` and `selector.nodes` count parsed Selector structure;
- `retention.maximumSnapshotLifetime` bounds the lifetime of a pagination
  continuation when advertised.
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
