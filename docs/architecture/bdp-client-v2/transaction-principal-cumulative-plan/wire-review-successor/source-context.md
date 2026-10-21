# Actual omitted-context source — committed5d7

External plan artifacts are not additional owned product files. requests.json in this successor retains every original method/body/header/ref/status field and adds exact URL provenance spellings only. Full independent literals are supplied there.

## F1 handles — packages/conformance/src/http-executor.ts:335–341

```ts
}

const HTTP_FIELD_NAME = /^[!#$%&'*+.^_`|~0-9A-Za-z-]+$/;
const HTTP_FIELD_VALUE = /^[\t\x20-\x7e\x80-\xff]*$/;
const HANDLE_ID = /^[a-z0-9]+(?:[.-][a-z0-9]+)*$/;
const SECRET_FIELD =
  /^(?:authorization|cookie|proxy-authorization|set-cookie|x-api-key|x-auth-token)$|(?:token|secret|password|credential|auth)/i;
```

## F3/F9 complete response and exact observation interfaces — packages/conformance/src/http-executor.ts:36–109

```ts
export interface RawHeaderLine {
  readonly name: string;
  readonly value: string;
}

export type RequestWriteState = "not-started" | "started-completion-unestablished" | "complete";

export interface ExactRequestObservation {
  readonly source: "raw-http1-serializer";
  /** Transient secret-bearing data, never safe to serialize directly into a report. */
  readonly headerLines: readonly RawHeaderLine[];
  /** Full authored entity metadata, not a claim about a transmitted prefix or server receipt. */
  readonly bodyPresent: boolean;
  readonly bodyOctets: number;
  readonly bodyDigest?: string;
  readonly writeState: RequestWriteState;
}

/** Trusted harness configuration, not portable scenario data or target-supplied authority. */
export interface ExactHttpModeOptions {
  readonly scope: string;
  readonly profile: "read" | "read-update" | "transactional";
  readonly routes: readonly { readonly method: HttpMethod; readonly url: string }[];
  readonly maximumRequestHeaderBytes: number;
  readonly maximumRequestBodyBytes: number;
  readonly defaultCredentialRef: string;
  readonly credentialHandles: readonly {
    readonly id: string;
    /** Empty means a deliberately anonymous handle; names are lower-case. */
    readonly headerNames: readonly string[];
  }[];
  /** Synchronous, trusted nonblocking lookup. Promises/thenables are rejected. */
  readonly resolveCredentials: (
    context: Readonly<{
      scope: string;
      url: string;
      method: HttpMethod;
      credentialRef: string;
      signal: AbortSignal;
    }>,
  ) => Readonly<Record<string, string>>;
}

export type RawHttpExchangeExecutor = (
  request: HttpExchangeRequest | ExactHttpExchangeRequest,
) => Promise<HttpExchangeResponse>;

export interface HttpExchangeResponse {
  readonly exactRequest?: ExactRequestObservation;
  readonly url: string;
  readonly status: number;
  readonly headers: Readonly<Record<string, string>>;
  readonly bodyText: string;
  /** Decoded message-body octets observed below a method-aware HTTP client, when available. */
  readonly bodyOctets?: number;
  /**
   * Exact bounded HTTP/1.x response bytes from status line through framed body.
   * Only the socket-level executor can supply this pre-normalization evidence.
   */
  readonly wireResponseBytes?: Uint8Array;
  /**
   * Compatibility view after authorization/runtime decoration. In exact mode,
   * repeated names retain their last value and this view has no wire-fidelity flag;
   * only exactRequest.headerLines preserves authoritative ordered occurrences.
   */
  readonly effectiveRequest?: {
    readonly url: string;
    readonly headers: Readonly<Record<string, string>>;
    /** Present only when the reported headers match bytes serialized onto the wire. */
    readonly headersTransmitted?: true;
  };
}

export type HttpExchangeExecutor = (request: HttpExchangeRequest) => Promise<HttpExchangeResponse>;
```

## F3/F9 actual serializer response construction — packages/conformance/src/http-executor.ts:748–780

```ts
    state = exchanged.writeState;
    requestDeadline(signal, deadline);
    const parsed = parseRawResponse(
      exchanged.bytes,
      request.method,
      maximumHeaderBytes,
      maximumBodyBytes,
    );
    const exactRequest = Object.freeze({
      source: "raw-http1-serializer" as const,
      headerLines: Object.freeze(lines),
      bodyPresent: body !== undefined,
      bodyOctets: body?.length ?? 0,
      ...(body === undefined
        ? {}
        : { bodyDigest: createHash("sha256").update(body).digest("hex") }),
      writeState: exchanged.writeState,
    });
    const response = {
      url: url.href,
      ...parsed,
      wireResponseBytes: Uint8Array.from(exchanged.bytes),
      exactRequest,
      effectiveRequest: {
        url: url.href,
        // Deliberately lossy last-occurrence compatibility view, never exact wire evidence.
        // Ordered exactRequest.headerLines is authoritative for every occurrence.
        headers: Object.freeze(
          Object.fromEntries(lines.map(({ name, value }) => [name.toLowerCase(), value])),
        ),
      },
    };
    requestDeadline(signal, deadline);
```

## F2 report URL redaction — packages/conformance/src/runner.ts:3319–3327

```ts
function redactUrl(value: string): string {
  const url = new URL(value);
  const hadQuery = url.search !== "";
  url.username = "";
  url.password = "";
  url.hash = "";
  url.search = "";
  return `${url.href}${hadQuery ? "?<redacted>" : ""}`;
}
```

## F7 real duplicate-token rejection — packages/server/src/transaction-runtime-configuration.ts:226–254

```ts
  const credentials = administrativeArray(own.credentials, 32).map((value) => {
    const entry = administrativeData(value, ["token", "principal", "canMutate"]);
    const token = boundedString(entry.token, 256),
      principal = boundedString(entry.principal, 16384);
    if (!/^[A-Za-z0-9_-]{32,256}$/.test(token) || typeof entry.canMutate !== "boolean")
      throw new TypeError("invalid private transaction credential");
    return { token, principal, canMutate: entry.canMutate };
  });
  if (!credentials.length || new Set(credentials.map((c) => c.token)).size !== credentials.length)
    throw new TypeError("unique private transaction credentials required");
  const secrets = credentials.map(({ token, ...rest }) => ({
    ...rest,
    credential: Buffer.from(`Bearer ${token}`),
  }));
  const associations = freeze(
    credentials
      .map(({ token, ...rest }) => ({
        // Domain separation binds this credential use; it does not prevent
        // guessing low-entropy tokens. Metadata digests are not secrecy guarantees.
        tokenDigest: transactionRuntimeDigest({ revision: "transaction-credential-1", token }),
        ...rest,
      }))
      .sort((a, b) => (a.tokenDigest < b.tokenDigest ? -1 : a.tokenDigest > b.tokenDigest ? 1 : 0)),
  );
  const authenticationIdentity = transactionRuntimeDigest({
    revision: "transaction-authentication-1",
    scheme: "bearer",
    associations,
  });
```

## F8 JSON-pointer existence — packages/conformance/src/runner.ts:2231–2245

```ts
  if (assertion.kind === "json-pointer") {
    if (body.kind !== "json") return outcome(assertion.id, false, observedBodyIssue(body));
    const value = jsonPointer(body.value, assertion.pointer);
    const exists = value !== undefined;
    if (assertion.equalsBinding !== undefined) {
      const expected = bindings.get(assertion.equalsBinding);
      if (expected === undefined)
        throw new ConformanceRunnerError("JSON assertion capture is unavailable");
      return outcome(
        assertion.id,
        exists && deepEqual(value, expected),
        "JSON pointer did not equal the earlier observed capture",
      );
    }
    const fixtureExpected =
```

## F8 whole-document JSON pointer — packages/conformance/src/runner.ts:2708–2723

```ts
function jsonPointer(value: unknown, pointer: string): unknown {
  if (pointer === "") return value;
  let current: unknown = value;
  for (const part of pointer.slice(1).split("/")) {
    if (typeof current !== "object" || current === null) return undefined;
    const key = part.replace(/~1/g, "/").replace(/~0/g, "~");
    if (Array.isArray(current)) {
      if (!/^(?:0|[1-9]\d*)$/.test(key)) return undefined;
      current = current[Number(key)];
    } else {
      if (!Object.hasOwn(current, key)) return undefined;
      current = (current as Record<string, unknown>)[key];
    }
  }
  return current;
}
```

## F10 full validator format setup — packages/conformance/src/schema-validator.ts:27–49

```ts
export function createJsonSchemaValidator(
  bundle: JSONSchemaType<unknown> | Record<string, unknown>,
): SchemaValidator {
  const ajv = new Ajv2020({ allErrors: false, strict: true });
  ajv.addFormat("uri", { type: "string", validate: isJsonSchemaUri });
  ajv.addFormat("date-time", { type: "string", validate: isJsonSchemaDateTime });
  const root = bundle as Record<string, unknown>;
  if (typeof root.$id !== "string" || root.$id.length === 0)
    throw new SchemaValidatorError("schema bundle must declare a non-empty $id");
  const schemaId = root.$id;
  ajv.addSchema(bundle, schemaId);
  const resolve = (schemaRef: string): ValidateFunction => {
    const ref = schemaRef.startsWith("#") ? `${schemaId}${schemaRef}` : schemaRef;
    let validator: ValidateFunction | undefined;
    try {
      validator = ajv.getSchema(ref);
    } catch (error) {
      throw new SchemaValidatorError("schema reference could not be compiled", { cause: error });
    }
    if (validator === undefined)
      throw new SchemaValidatorError(`unknown schema reference '${schemaRef}'`);
    return validator;
  };
```

## F11 actual absent-body exact assertion — packages/conformance/src/runner.ts:1918–1936

```ts
      assertion.id,
      assertion.equals === observed.writeState,
      "Exact HTTP write state did not match",
    );
  if (observed.writeState !== "complete")
    return outcome(assertion.id, false, "Exact HTTP request write completion was not established");
  const matches =
    assertion.kind === "request-authored-body"
      ? observed.bodyPresent === (body !== undefined) &&
        observed.bodyOctets === (body?.byteLength ?? 0) &&
        observed.bodyDigest ===
          (body === undefined ? undefined : createHash("sha256").update(body).digest("hex"))
      : options.exactExecution !== undefined &&
        exactHeadersMatch(request, url, body, observed, options.exactExecution.configuration);
  return outcome(
    assertion.id,
    matches,
    "Exact HTTP authored request did not match the observation",
  );
```

## F3/F9 actual post-delegation wrapper — apps/bdptest/src/transactional-matrix.test.ts:275–326

```ts
          throw new Error("generic HTTP executor forbidden");
        },
        exactExecution: {
          execute: async (request) => {
            const response = await target!.target.execute(request);
            if (afterFirstExchangeBudgetMs !== undefined && firstExchangeTimer === undefined) {
              firstExchangeTimer = setTimeout(expireCohort, afterFirstExchangeBudgetMs);
            }
            if (
              typeof control?.invalidReceiptCredentials === "string" &&
              response.status === 200 &&
              [
                baseline.operationTargets.updateWhere,
                baseline.operationTargets.deleteWhere,
              ].includes(response.url)
            ) {
              const receipt = JSON.parse(response.bodyText) as {
                id?: unknown;
                idempotencyKey?: unknown;
              };
              if (
                receipt.idempotencyKey === control.invalidReceiptCredentials &&
                typeof receipt.id === "string"
              )
                deniedReceiptUrls.add(receipt.id);
            }
            await control?.afterExchange?.(response, signal);
            signal.throwIfAborted();
            return response;
          },
          configuration: target.target.exactConfiguration!,
        },
        harness: {
          ...target.target.harness,
          cleanup: async (...args) => {
            try {
              await target!.target.harness.cleanup(...args);
            } catch (error) {
              cleanupFailures.push(error);
              throw error;
            }
          },
        },
        schemaValidator: createJsonSchemaValidator(schema),
        declaredTargetLabel: "private-installed-transactional-partial",
        prepareTimeoutMs: 10000,
        cleanupTimeoutMs: 10000,
        requestTimeoutMs: 5000,
      },
      target.context,
    );
    observed = { result, sessions, baseline };
```

