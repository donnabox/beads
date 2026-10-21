# Proposed internal reference deployment-file contract v1

This is a reviewable proposal for a non-normative implementation input. It adds no BDP schema/URI/HTTP behavior. See plan.md for ownership, bounds, validation and delivery gates. All objects are closed at their stated layer. Inner contracts remain delegated to existing installed capture/compile modules.

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

The structural notation above does not replace runtime validation or inner semantics. Explicit empty arrays/objects remain meaningful only where existing capture permits them. RU descriptor[] selects its shipped reference closure; TX requires nonempty descriptors. RU principal policy and TX canMutate remain different existing policy contracts.

## Projection to existing captures

| Shared startup input | Loader behavior |
| --- | --- |
| bdptest.deploymentFile | Absolute local file; bounded single-handle read. Not stored inline with secrets in config diagnostics. |
| server.advertisedProfile | Explicit write token; must equal artifact profile. Config token alone grants nothing. |
| scope.url | Inject as captured scope. No competing Scope field allowed in envelope. |
| server.host/port | Validate supported intended listener; inject RU listener. Retain TX intended listener separately from its current private port0 capture. |
| server.limits | Require exact current enforced tuple. No envelope override or implicit substitute. |
| server.internalFaultResource / bdptest.fixture | Refuse if present when consuming a write deployment, because their write behavior is not implemented here. |

RU projection is `{scope: startup.scope.url, listener:{host,port}, ...configuration}` with exact keys validated before projection. TX projection is `{scope: startup.scope.url, ...configuration}` into installed capture, then existing seam-free runtime preflight. Avoid spread order that permits unrecognized supplied scope/listener fields to override injected values; closed union validation precedes projection.

## Supported effective Read tuples

| Profile | Page default/max | Selector bytes/depth/nodes | Cursor milliseconds |
| --- | --- | --- | --- |
| read-update | 20 / 100 | 4096 / 32 / 256 | 60000 |
| transactional | 2 / 32 | 4096 / 32 / 256 | 1000 |

These describe inspected current source, not constants to duplicate in the new loader. Import existing implementation constants or derive TX tuple from captured readOptions; tests should reject mismatch while preserving existing Read tuning. Inner artifact limits retain their current 1MiB/artifact and16MiB total ceilings; outer file32MiB, depth16,100000 nodes is a separate explicit input transport bound.

## Binding and limitations

RU capture produces its established configuration/artifact/policy/authentication identities. TX preflight produces compiled artifacts/settings/authenticator but its existing live metadata function requires actual owner readConfiguration before producing policy/view/runtime identities. The loader must keep those unavailable rather than claim owner proof. File hash and prepared-data identities are provenance, not evidence eligibility.

Storage existence, compatible private store format, persisted Scope, installed live Type inventory, current policy/view and recovery are validated by the actual owner in the later lifecycle slice. A pure loader must not claim to have checked those or open a store as a hidden validation side effect. The immutable prepared values are passed forward without rereading the file.
