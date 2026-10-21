# Proposed internal reference deployment-file contract v1

This is a reviewable proposal for a non-normative implementation input. It adds no BDP schema/URI/HTTP behavior. Corrected successor3 follows the one completed startup PLAN review and exact root F1–F8 disposition; predecessor contracts stay preserved. See plan.md for twelve-file ownership, bounds, validation and delivery gates. No actual PLAN retry or source release follows from this artifact. All objects are closed at their stated layer. Inner contracts remain delegated to existing installed capture/compile modules.

## Envelope union

```typescript
type ReferenceDeploymentFile =
  | {
      format: "bdptest-reference-deployment-1";
      profile: "read-update";
      configuration: {
        directory: string; // absolute, existing RU bounds
        create: boolean;
        credentials: readonly { token: string; principal: string }[];
        policies: Record<string, {
          create: boolean;
          write: boolean;
          hidden: readonly string[];
          readOnly: readonly string[];
        }>;
        contracts: {
          descriptors: readonly { retrievalUri: string; utf8: string }[];
          schemas: readonly { retrievalUri: string; utf8: string }[];
        };
        maximumEndpointMultiplicity: readonly {
          endpoint: "source" | "target";
          linkConformsTo: string;
          max: number;
        }[];
      };
    }
  | {
      format: "bdptest-reference-deployment-1";
      profile: "transactional";
      configuration: {
        directory: string; // absolute, existing TX bounds
        create: boolean;
        credentials: readonly {
          token: string;
          principal: string;
          canMutate: boolean;
        }[];
        artifacts: {
          descriptors: readonly { retrievalUri: string; utf8: string }[];
          schemas: readonly { retrievalUri: string; utf8: string }[];
        };
      };
    };
```

The structural notation above does not replace runtime validation or inner semantics. Explicit empty arrays/objects remain meaningful only where existing capture permits them. RU descriptor[] selects its shipped reference closure; TX requires nonempty descriptors. Credential tokens must be unique under the existing captures; duplicate principals remain accepted. RU principal policy and TX canMutate remain different existing policy contracts.

## Projection to existing captures

| Shared startup input | Loader behavior |
| --- | --- |
| bdptest.deploymentFile | Absolute local file; bounded single-handle read. Fixed validator messages name field/rule only and never echo rejected input. Resolved path is emitted in startup.config like bd.workspace and must not contain secrets; contents/tokens/payload/cause never enter diagnostics. |
| server.advertisedProfile | Explicit write token; must equal artifact profile. Config token alone grants nothing. |
| scope.url | Inject as captured scope. No competing Scope field allowed in envelope. |
| server.host/port | Validate supported intended listener; inject RU listener. Retain TX intended listener separately from its current private port0 capture. |
| server.limits | Match exact seam-free code/capture tuple as preflightReadLimits only. No envelope override, implicit substitute or claim of persisted-owner enforcement. |
| server.internalFaultResource / bdptest.fixture | Refuse if present when consuming a write deployment, because their write behavior is not implemented here. |

RU projection is `{scope: startup.scope.url, listener:{host,port}, ...configuration}` with exact keys validated before projection. TX projection is `{scope: startup.scope.url, ...configuration}` into installed capture, then existing seam-free runtime preflight. Avoid spread order that permits unrecognized supplied scope/listener fields to override injected values; closed union validation precedes projection.

## Supported code/capture preflight Read tuples

| Profile | Page default/max | Selector bytes/depth/nodes | Cursor milliseconds |
| --- | --- | --- | --- |
| read-update | 20 / 100 | 4096 / 32 / 256 | 60000 |
| transactional | 2 / 32 | 4096 / 32 / 256 | 1000 |

These describe inspected code/capture constants, not persisted-owner proof or constants to duplicate in the loader. Public summary calls this tuple preflightReadLimits. Import existing constants or derive TX tuple privately from captured readOptions; reject mismatch while preserving existing Read tuning. Inner artifacts retain1MiB/artifact and16MiB total ceilings. Independently retain32MiB encoded outer-file ceiling as the sole PRE-decode bound on trusted operator input: full decoder allocation/work occurs before depth/node refusal and is accepted within this byte bound. Decode with the existing strict decoder, then iteratively bound the LOSSLESS outer tree at depth16/100000nodes BEFORE numeric materialization; JsonNumberLiteral is atomic. admitJsonNumbers uses diagnostics:1 and a fixed sanitized formatter; refusal pointers/literals/causes never escape. These depth/node/diagnostic bounds are POST-decode, not pre-decode memory/work guarantees. Do not add a second JSON prescanner. Escaping and envelope overhead mean inner-admissible data, even below16MiB, may exceed32MiB; no guaranteed full-inner-budget fit or universal derating ratio. Isolate the outer overflow test with a valid small-inner envelope plus legal JSON whitespace, without raising either ceiling.

## Public inert handle and private captured state

The package data subpath exposes only an explicitly typed frozen inert prepared handle/summary: format/profile/canonicalScope, intended listener, preflightReadLimits, artifact/file identities and permitted existing RU identities. No captures, projected credential objects/nested tokens, authenticate/metadata/readOptions functions or private runtime types occur in public values or emitted declarations. A PRIVATE reference-deployment-state.ts WeakMap stores the exact immutable installed input/capture and TX preflight keyed by actual handle identity. Only an internal, non-exported-by-package accessor retrieves registered state; forged/foreign handles fail with a fixed local error. No root or data-entry re-export of registry/accessor, no bare/deep registry package import, no file reread or public start/admit operation. Loader and registry emitted files both belong to exact packaging inventory; package exports add only ./reference-configuration. Later internal composition consumes the stored capture and retains the existing deterministic re-capture allowance where needed.

## Binding and limitations

RU capture produces its established configuration/artifact/policy/authentication identities. TX preflight produces compiled artifacts/settings/authenticator but its existing live metadata function requires actual owner readConfiguration before producing policy/view/runtime identities. The loader keeps those unavailable rather than claim owner proof. Actual TX/RU captures, nested credentials and capability functions stay only in private registered state, not on the public handle. File hash and prepared-data identities are provenance, not evidence eligibility.

Storage existence/compatible private store format and persisted receiptMaximumItems, snapshot limits/token policy, runtime limits, contract identity, Scope/installed Type set and current policy/view/recovery are validated by the actual owner BEFORE listen in the later lifecycle slice. A pure loader must not claim to have checked those or open a store as a hidden validation side effect. The immutable prepared values are passed forward without rereading the file.
