# P1a: installed mutable reference payload (proposal for root approval)

Author checkout `/Users/dbox/repos/wt/vickie/vickie-update-installed`, branch
`codex/vickie-update-installed`, exact base `2c5fe9e04250ae036e681f46c60d3592e5f1cdf1`,
GitHub origin verified. U2 checkout remains frozen. No source edits yet.

## Existing boundary and proposed shape

`apps/bdptest/src/main.ts` currently calls `createConfiguredBdptestReadServer`,
which obtains the sealed Read brand before constructing the immutable adapter.
The public Node listener accepts only that branded server. Neither brand nor
listener is suitable as a fake write-profile qualification grant. Keep this
entire public Read composition, old evidence artifact, and ordinary write-profile
refusal intact during P1a.

U2's actual mutable owner/compiler/policy code is installed in the server package
but private. Its runtime derives Scope from a loopback ephemeral listener, uses
fixed limits, refuses Origin and absolute-form targets, and obtains canonical
routes from source-only support. P1a must remove those differences for the
installed candidate; merely wrapping the existing test helper is insufficient.

Proposed implementation:

1. Extract the U1 canonical discovery/Directory projection into a private server
   module used by a single installed mutable reference factory. Reuse
   `read-http-routing.ts`, `read-http.ts`, raw target parser and actual U2 owner.
   Do not import test support in emitted code. Closed directory mapping still
   supplies the eight singleton operations plus sequence.
2. Give the private shared runtime a captured deployment transport configuration
   for canonical Scope and separate listener host/port. The installed factory
   supplies it; ordinary development facade cannot. Scope is always configured,
   including during restart; bind address, Host, absolute-target authority and
   forwarding headers never redefine Resource identities. Reuse raw target
   normalization, retain malformed-target rejection, support valid absolute-form
   Read targets. CORS-disabled requests carrying Origin are processed normally
   without grant headers; OPTIONS remains method refusal. Keep development
   behavior unchanged. Existing authentication/body/key/negotiation/lifecycle
   receiving remains shared, not copied into a second server.
3. Private installed configuration captures directory, explicit create/reopen,
   Scope/listener values, immutable contract artifact bytes, trusted credentials,
   policy snapshots and fixed aggregate policy. U2 compiler/registry stays the
   single contract adapter. Initial limits are an explicitly fixed supported
   tuple equal to the actual owner; reject overrides rather than silently
   advertise unenforced values. Future configurable limits require separate
   owner validation, not P1a hidden hardcodes.
4. Supply a concrete reference artifact bundle from current
   `REFERENCE_TYPE_DESCRIPTORS`, instantiated at Scope-local `/types/<name>` URLs
   with all parent/ownership/endpoint Type references remapped consistently.
   These are honestly distinct local identities, not a claim to host
   `https://work.example/...`. Actual Type inventory/descriptor GET/HEAD traverse
   the same owner. Schema-bearing qualification configurations supply pinned
   offline bytes through the same factory; no Type/schema network installation.
   If schema hosting is configured, serve exact captured bytes only from a
   closed, Scope-local manifest, with JSON negotiation/HEAD; arbitrary external
   URLs are not routed. Defaults start an empty durable graph, populated by real
   protocol writes, not synthetic seeded database rows.

## Packaging/bootstrap decision requested from root

P1a must exercise installed bytes without opening an ordinary public profile
bypass. Proposed private qualification process entry in the emitted server
package (not a package export or bin, not a bdptest CLI switch): installed tests
resolve its file from the packed package and start it with an inherited IPC
channel. It calls the exact installed factory/configuration above and reports
`claimEligible:false` plus configuration identity. It accepts only startup and
shutdown control, no synthetic store/clock/outcome/fault APIs. Qualification is
explicitly a candidate observation, not a capability grant or sealed artifact.

The later root-owned evidence gate adds the ordinary bdptest startup call to the
same factory only after profile/target/configuration/source-bound evidence is
verified. It must not turn the bootstrap IPC condition into admission. There is
no user-selectable profile activation flag in P1a, and no resealing of Read.

This proposes private server installed-factory/qualification-entry modules and a
narrow private runtime transport seam, instead of an app import that bypasses
package exports. Alternative packaging needs root arbitration: bdptest cannot
legitimately import U2 private modules through its existing public dependency.
Do not add a publicly exported ungated factory merely to solve that dependency.

## Scope and coordination

Expected files: new private installed-reference config/factory/route/process
modules under `packages/server/src`; narrow `development-runtime.ts` private
transport capture; existing `development-authority.ts` only if needed to expose
captured exact Type descriptor bytes (prefer existing `perform`); installed
process tests/support under apps/bdptest; package output inventory and separate
installed candidate smoke. Public main/server-composition/config/admission remain
unchanged in P1a, with negative ordinary-startup tests. Root approval needed for
runtime/raw HTTP helper changes and packaging allowance. No store/evaluator,
authority-read/pagination, schema/normative, TX or evidence changes.

## Done when

- Packed/installed process cold client follows Scope Link -> descriptor ->
  directory -> all nine targets; real create/update/Link/alias/sequence afterstate
  and Type descriptor navigation work against one mutable owner.
- Stable configured Scope survives listener restart/reopen and yields retained
  exact-key results; mismatched installation refuses without rebinding identity.
- Raw absolute-form target/Host forwarding cannot change canonical identity;
  unsafe target, duplicate credentials/key, method/Accept/conditional/HEAD and
  CORS-disabled Origin cases are directly asserted, including no key capture
  before refusal where applicable.
- SIGTERM drains accepted work, startup failure closes listener/owner, second
  owner/refused reopen cannot corrupt the store, all test processes cleaned.
- Factory/configuration identity is recorded from actual captured bytes,
  principal-policy/limits/provisioning applicability explicit. No secret logged.
- Public Read evidence and ordinary write-profile failclosed tests unchanged;
  installed inventory/deep-import denial, build/type/boundaries/checks/smokes pass.

Separate later prerequisites: executable cumulative RU requirements on this
installed target/configuration, independent evidence artifact/generator/verifier,
then gated main/config/API activation; TX factory composes its own single owner.
No P1a pass is a full-profile claim. No product decision identified; packaging and
shared transport seams are implementation coordination decisions for root.
