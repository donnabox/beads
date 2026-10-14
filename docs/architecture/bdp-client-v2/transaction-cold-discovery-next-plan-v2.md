# C3 internal Transactional discovery — revised plan v2

Plan revision only, under C3 of accepted `transaction-cold-client-plan-v2.md` SHA256 `4252bf3e2f951c10ad6bfa351bf7d5212349ea18dc506b770728d949dc7b2dfa` and the sole fork design.md. Root approval of the named scope amendments below is required before implementation. Public serving, discovery advertisement, admission, evidence promotion and C4 cold-client success remain excluded.

Original handoff is preserved as [reviewed-v1](transaction-cold-discovery-next-plan-reviewed-v1.md). [Dispositions](transaction-cold-discovery-next-plan-dispositions.md) account for both reviews and qualify the original native floor conclusion. Root accepted the [replay-floor audit](transaction-discovery-replay-floor-audit.md): position floor `pos-0` and replay checkpoint `genesis` are separate roles; no codec/spec/fixture amendment is needed.

## Ordering and exact inspected basis

Inspected frozen C1 source `20d61119b5faf7251a5d285c5c34a1049866ffe5`, tree `3de0c4b5589423d01eb0dcb7df7ad150d184273d`. Relevant discovery, parser, owner and normative seams are unchanged from original main581bad0. C1 has now merged per root; the original handoff's unmerged-C1 condition is stale. Alias HTTP work is still separately owned; no edits or tests in that checkout. Root orders this C3 implementation **after Snapshot creation-minimum work**, on an exact reviewed ordinary integration base supplied at dispatch. Re-anchor symbols, counts and any precise source lines on that base. This plan does not claim the inspected source already contains later integrations.

C3 produces an internal immutable description from the real owner and one operation map. It neither publishes that description nor makes a new serving-readiness decision. C4 must still wait for truthful required serving behavior; C5 public admission/evidence remains separate. Do not import a broad all-operation installed walk into C3 or treat private profile JSON as permission to advertise it.

## Required scope amendment: authoritative protocol parsers

The accepted C3 says to parse using protocol/schema authority. The original handoff incorrectly replaced that with test-only AJV and forbade the necessary seam. Propose exactly two schema-backed parsers, following `packages/protocol/src/read-update-values.ts` (`validator`, `parseReadUpdateShape`, `parseReadUpdateDiscovery`, `parseReadUpdateOperationDirectory`) and `read-values.ts` reusable helpers:

- New `packages/protocol/src/transaction-discovery-values.ts`: `parseTransactionalDiscovery(value: unknown)` and `parseTransactionalOperationDirectory(value: unknown)`, plus their readonly return types. Keep its two schema-root references private and separate from the Read/RU registries. Bind exactly `#/$defs/transactionalDiscovery` and `#/$defs/transactionalOperationDirectory` in the existing packaged canonical bundle.
- Follow existing lazy AJV2020 setup, strict validation and URI/date-time formats, using `readCanonicalSchemaBundle` and `requireSchemaValidator`. Guard/snapshot JSON through `snapshotProtocolRecord`; reject unsupported host objects/accessors/proxies under that existing boundary, preserve recursive immutability and do not silently strip fields or add a wire depth cap. Use the existing Unicode scalar check pattern for names/string values, canonical URL validation for scope and the nine URL members, and canonical Type-ID validation for optional multiplicity policies as the existing discovery parsers do. Do not introduce a generic exported parse-any-schema helper.
- Add only these two parser and two type reexports to `packages/protocol/src/index.ts`, allowing server consumption through the existing `@bdp/protocol` package root. No package export-map change or deep import. **This is a narrow protocol API addition requiring root approval**, not a new public Transactional server profile or a parser that attests implementation readiness.
- Parser policy is the existing normative schema, including its valid optional members. Do not make this candidate's omission of History/order/multiplicity a global parser prohibition, enforce owner identity in a generic parser, or require all references to use this owner's fixed route spelling. Inputs with a different well-formed Scope may be structurally valid. Candidate-specific constraints are separate builder assertions/tests below.
- Add `transaction-discovery-values.test.ts` positive/negative canonical schema/parser tests, including actual normative discovery fixture and exact directory, missing/unknown members, wrong profile/target, forbidden retention.idempotency, malformed navigation URLs, hostile input and nested immutability. Check valid optional fields remain accepted. This is object-envelope parsing, not a new raw JSON duplicate-key/carrier parser or general Transactional validator.

The server builder must call both parsers in production before returning its immutable values. Test-only AJV is supplemental independent verification, not the production boundary. No schema or illustrative fixture edits. Read schema registry/projection roots and historical evidence remain unchanged; run the existing Read evidence gate after the package-root edit, reporting actual binding effects rather than assuming a pass or resealing silently.

## One internal builder, one owner authority

Move the existing frozen twelve-operation relative mapping from `transaction-http.ts` (`TRANSACTION_HTTP_OPERATIONS`, inspected56–69; dispatch target construction180–190) to new source-private `packages/server/src/transaction-discovery.ts`. HTTP imports that same object; no second production operation list or configurable targets. Independently pin the twelve normative name/value pairs in tests and compare all actual handler targets, not merely length12. Relative directory values remain the schema's exact strings, resolved under `operations/`.

The private module builds the normative description and parsed directory from canonical owner Scope, a same-entry context and frozen actual configuration. Inputs cannot choose profile, targets, floor policy, History flag or extensions. It must not start a second owner, scan groups, use latest receipt/Snapshot as context, or fabricate a second descriptor inventory.

At service construction capture the existing immutable owner configuration once (reuse the current `owner.readConfiguration()` for receipt maximum). Capture the small advertised Read-limit scalar object once and pass the **same captured values** to `createTransactionReadPlane` and the discovery builder. Existing authority validation must still compare advertised page/Selector values with enforced pagination/Selector options and receipt maximum. Do not reread mutable caller options after validation. Copy only the required existing configuration shape; no new dynamic settings API. Mutation-after-construction tests must show both enforcement and description remain bound to the captured values.

Add a source-private service observation method following `sequenceObserved` (service420–437): `run`/`check`, then exactly one synchronous `owner.withRead` entry. Assemble the parsed immutable description and directory from that entry's epoch/view/head plus the captured immutable configuration. Propagate `entry-refused.cause` as refusal/error, never return it as a successful description. No reader, DB object, thenable callback or mutable context escapes. Do not call `owner.readConfiguration()` inside withRead: it enters the owner again and triggers the real reentrancy guard (`transaction-store.ts`, enter340–344, withRead881–957, readConfiguration1706).

### Canonical Scope invariant

Expose only a readonly source-private service `scope` value captured from `owner.scope` on the existing frozen service object. At `createTransactionHttpHandler`, compare `parseCanonicalScope(canonicalScope)` with that service scope and reject mismatch before constructing targets. This small constructor assertion/private property is an explicit part of the proposed C3 amendment; no owner accessor or runtime argument change is needed. Test mismatched handler/owner Scope refusal and normal equality. Derivation twice from unrelated strings must not be treated as identity proof merely because the runtime currently passes the same local variable.

The actual installed configuration uses explicit canonical authority, e.g. a non-root `https://reference.invalid/acme/`, while the listener uses a different loopback origin/port. Only in that explicit canonical fixture assert targets use configured Scope rather than listener origin. Listener-derived authority mode may legitimately contain its listen port; it is not a failing case or prohibited mode.

## Exact route roots and independent integration tests

The builder uses the following closed roots. Tests check the real routing/materialization seams, not a second copy of the builder's constants alone. Some fields are prefixes requiring a subject or query; do not demand bare-root200 or mistake a recognized missing subject for a wrong route.

| Member | Relative root | Existing authority and independent check |
|---|---|---|
| beads | `beads/` | `requestFromUrl` classifies Bead collection; compare existing authority discovery data. |
| links | `links/` | `requestFromUrl` classifies Link collection; compare existing authority discovery data. |
| types | `types/` | `requestFromUrl` classifies Type collection; compare existing authority discovery data. |
| aliases | `alias/` | Existing `authority-read.ts` discovery uses singular root. Append a canonical alias name, then require `classifyAliasPath` returns that name; empty root is intentionally not an alias subject. |
| operations | `operations/` | Resolve each parsed directory value and require equality with every actual handler target. C3 directory GET remains refused. |
| receipts | `receipts/` | Real returned receipt id is under this exact root and routes through the HTTP receipt branch; refetch same-principal receipt. Bare root need not be a receipt. |
| snapshot | `snapshot` | No trailing slash. Actual HTTP Snapshot creation branch recognizes this path; focused source/emitted observation confirms it reaches that branch. |
| changes | `changes/` | Append supported `after=genesis`; `transactionChangefeedRequest` recognizes it. Bare changes URL without after/start is parameter-invalid, not an alternate route. |
| events | `events/` | `transactionEventRequest` recognizes the Scope Event Source. No projection expansion is implied. |

These names cover all nine URL members, in addition to scope. Existing seams: `authority-read.ts` materialize discovery738–750; `read-request.ts` requestFromUrl142 and classifyAliasPath935–952; `transaction-http.ts` Snapshot378, receipts466–467, directory/refusal546–553; `transaction-changefeed.ts` request classifier89–107; `transaction-events.ts` request143–180. Keep literals local to the new private builder and test consistency with existing classifiers; do not move every router into a new generic URL framework. An invented alias plural, Snapshot slash, wrong operation relative value or listener-origin substitution must fail an independent test.

## Floor, limits and optional fields

Current owner is fixed no-pruning/no-erasure. Emit `minimumReplayPosition: "pos-0"` in the **position namespace** used by readContext/head and group.position/previousPosition. Empty head equals this floor; first committed group's previousPosition equals it; later head advances and floor stays fixed. Source no-pruning invariant (`transaction-store.ts` readChanges1672–1675) treats missing positive groups as corruption, not retention.

Separately derive the checkpoint at ordinal0 through `transactionCheckpoint(scope, context, 0)`, which is `genesis`. Use that checkpoint in actual after requests after commits/reopen/receipt expiry. Do **not** send discovery.minimumReplayPosition directly to the feed, require checkpointPosition to decode pos-0, change the codec, or rewrite existing fixtures. Protocol position tokens are opaque to clients; fixed pos-N checks here test this private owner, not a universal prefix law. No production history scan. Future pruning/erasure needs its actual owner floor and is outside this patch.

Build limits member-by-member as exactly:

```text
limits = {
  page: { defaultItems, maximumItems },
  selector: { bytes, depth, nodes }
}
```

All five values come from the same captured/enforcement-validated Read configuration. `cursorTtlMilliseconds` is an internal field and is excluded, including from nested output. No request/resource/patch/sequence/validation/transaction/retention groups are advertised in this candidate; their absence does not remove existing enforcement requirements. No inferred retention duration or retention.idempotency.

Candidate emits the17 required discovery members plus limits only. Omit `order` (canonical-uri baseline applies), `maximumEndpointMultiplicity` (actual installed configuration has empty declarations), and `historicalResolution` (unsupported capability). The generic parser still accepts those optional fields when schema-valid; candidate omissions are checked against its exact allowed-key set. Reject unexpected candidate input rather than stripping it.

Structural parser/schema negatives: wrong profile/version, missing/unknown member, wrong directory name/value, forbidden retention.idempotency, invalid limits shape, malformed URL. Independent candidate negatives: foreign but canonical Scope, stale epoch/view/head, wrong otherwise-valid route roots, History addition, unconsumed limits or mutated configuration. The latter may pass the normative schema and must fail the separately named owner/config/literal oracle. Remove the impossible “no lower-profile members” predicate; Transactional legitimately contains inherited Read/RU members.

## Explicit owner/file inventory

Subject to root approval of the parser/private-scope additions:

- Protocol: new `packages/protocol/src/transaction-discovery-values.ts` and its test; narrow parser/type reexports in `packages/protocol/src/index.ts`.
- Server: new `packages/server/src/transaction-discovery.ts` and its test; narrow observation/config/scope additions in `transaction-service.ts` with focused `transaction-service.test.ts`; map import and scope-invariant assertion in `transaction-http.ts` with focused `transaction-http.test.ts`.
- Existing emitted-boundary test `packages/server/src/transaction-installed-contracts.test.ts`: small relevant route/refusal/reopen checks. `transaction-installed.test.ts` is configuration capture only and gets no launcher graft.
- Packaging `scripts/smoke-executables.mjs`: exactly four server emitted entries for the new private module, deep-import denial, and a small installed protocol-root parser import/positive-negative parse check. Protocol schema bytes remain unchanged. No new package export-map entry.
- One narrow `apps/bdptest/src/transactional-discovery-installed.test.ts` only for the packed cohort below, reusing `scripts/transaction-installed-observation.mjs` and existing credentialed canonical raw transport/lifecycle. No new launcher or ready-metadata seam.
- Narrow `docs/design/transaction-implementation-progress.md` text may record internal assembly complete while serving remains refused, only after implemented and verified.

No store, schema, normative spec, codec, configuration/runtime, public server index/admission, catalog/flags, sealed evidence, runner, RU fixture, Types, erasure, history or Selector changes. Any further necessary seam returns to root. Public protocol parsing exports are the expressly proposed exception to v1's no-parser/no-public-API restriction; public **serving** remains closed.

## Evidence split and bounded installed ledger

Prefer source and existing emitted-boundary tests for owner advancement, two real commits, immutable configuration, one owner entry, close/fence/reentrancy, receipt expiry and reopen, floor/checkpoint relations, route classifiers, scope mismatch and exact discovery refusals. Use existing disposable direct clock seams only for expiry evidence; no installed fake clock. Reuse emitted `transaction-installed-contracts.test.ts`, which already imports dist runtime and exercises credentialed operations/reopen. Its evidence is emitted/in-process, not packed. Add only missing C3 assertions to relevant fixtures; do not replay all C1/alias/Snapshot cohorts to prove JSON assembly.

A small actual fresh packaged check is still needed for module inclusion, parser imports and unchanged discovery refusal. New HTTP cohort: **one child,12 requests**, one Bead, no Link, no reopen, no real wait. Hard ceiling16 requests/one child; unused headroom is not permission for unrelated work. Existing launcher10s start/close, wire5s, explicitly chosen45s aggregate and60s outer, existing raw/body ceilings. These are a named reuse of the Snapshot/C1-style harness bounds, not a claim all installed harnesses use them.

| Requests | Actual observation |
|---|---|
| 1–3 | Credentialed GET Scope, bdp.json, operations/:500 + transactional-discovery marker; no service-desc/profile/directory success. |
| 4–6 | Credentialed HEAD same three: same refusal/status/marker, bodyless. |
| 7–9 | Anonymous GET same three:401, no leaked Scope context/targets. |
| 10 | One actual credentialed create batch at a known existing private target; schema-valid receipt and literal created Resource. |
| 11 | Point GET actual created id; literal state and revision agree with independent intended values/receipt. |
| 12 | Exact same-key/same-body batch retry; retained same receipt/result and no duplicate effect. |

This intentionally removes the original redundant mutation→minimum collections→reopen whole walk from C3 packed acceptance. C1 is merged and its correctness remains a prerequisite/regression responsibility; no new C1 evidence claim is made here. Reopen/receipt/floor authority is directly/emitted tested where the internal method is available. C3 cannot observe its internal builder via a wire descriptor because wire discovery is deliberately refused.

Packing smoke checks separately import the two new parsers from the actual installed protocol root, parse the normative checked-in fixture/directory, and reject a known invalid member; they do not serve that fixture or inject it as candidate authority. They verify private server deep imports remain denied and the expected four emitted files exist. These are zero-wire module/package checks, not extra hidden listener sessions. Pin exact candidate source/tree/payload/config/schema/artifact bindings, process cleanup and all actual requests. No whole-profile/cold-client claim; no self-attested internal description added to IPC ready metadata.

## Done when and unresolved work

After reviewed base and root scope approval: both production parsers validate assembled outputs; directory shares actual route authority; context comes from exactly one live owner entry with immutable validated configuration; owner/handler Scope mismatch fails; floor pos-0 and independent genesis replay are proved without conflation; limits and optional policy are explicit; independent structural/owner oracles fail intended substitutions; lifecycle errors stay intact; required routes remain privately refused over actual packed bytes; canonical routing and module privacy survive packaging.

Run meaningful affected protocol/server/emitted tests, build/typecheck/format/lint/boundaries and packaging smokes plus the existing Read evidence verifier; report exact exits and any source-binding impact. No automatic fullworkspace or broad installed cohort repetition. Freeze source before actual package proof and preserve prior evidence. No source implementation is authorized by this revision.

C3 stays behind Snapshot as root directed. Remaining C2 questions/repairs (retained position comparison, partial/unknown minima, visibility/conditional admission, Event projection etc.) must be tracked individually under current rulings, not solved here or overstated as all product blockers. Retained epoch/view may be separately safe. C4 publication and real Scope-only cold walk, and C5 evidence/admission, remain separately gated.
