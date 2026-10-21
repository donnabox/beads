# Serving readiness after PR130: retained-key depth refusal

**A concrete settled serving repair is established: allow valid direct-carrier retries to reach authoritative retained-key comparison before the private execution-depth refusal.** This takes priority over adding the deferred sequence witness. It does not establish a public depth policy, arbitrary-depth execution support, or readiness to enable cold Transactional discovery.

## Bound source and new evidence

Inspected clean committed candidate 1f9cb46c42fa69a09ac847f57e47bc96390aea6f, tree 80a6fe0605698e18f9504432407e7943e1f665b4, ordinary parent 77f4b89c117e46f0e73476f69b4c8b377306452b, in /Users/dbox/repos/wt/vickie/vickie-changefeed-parameter-context. PR130 CI/council/merge remain root-owned; this is not a merge assertion. Read the sole design live checkpoint 2026-10-15T14:57:59Z and previous transaction-after-parameter-context-next-audit.md (9ed8fbba29c30d6e2b848eef2a2701eedcb67b9454dc2b5ae8b873e6072de20b).

The previous audit identified private depth refusal but did not trace its interference with an already-bound key. This report strengthens that exact residual; it does not repeat completed context fixes or reclassify diagnostic counts as functionality.

## Concrete failed client scenario and settled law

1. An authenticated client submits a valid shallow createBead with explicit local id, valid configured type, properties {}, metadata {}, and HTTP Idempotency-Key K. It receives a completed receipt.
2. The same client retries that same direct carrier/key, changing only metadata to an object containing forty nested objects. The body remains comfortably below the existing byte limit. This is different normalized input, with ordinary valid carrier syntax; metadata is not subject to the Type's properties schema.
3. Current source returns private bodyless 500/unsupported-transaction-composition before looking up K. The required result is 409/idempotency-conflict, preserving the original outcome. A subsequent original shallow retry must return its retained receipt; the deep conflicting request must create neither a second receipt nor an effect.

This is a **source-established control-flow defect**, not a newly executed observation. The exact proposed metadata variant has not been run. Existing HTTP test transaction-http.test.ts589–621 already constructs forty-deep valid properties and pins the fresh-key private500/no-effect/key-reuse behavior. It does not exercise an already-bound key, so it cannot settle this ordering.

Normative authority: docs/specs/bdp.md4515–4565 orders syntax, principal, then authoritative key comparison, before fresh admission capacity; a bound different normalized request is409, an equal one returns the saved receipt subject to negotiation, and nothing below key comparison applies to a duplicate. The failure occurs after known syntax and principal checks, so treating it as malformed400 would also be wrong. Metadata's absence of a Type schema is explicit at2212–2217. This conclusion requires neither S1-Q1 nor G2-Q1.

## Exact source chain and existing correct outcome paths

- transaction-batch.ts128–178 completes mixed carrier parsing, references/labels and full embedded Selector grammar, then calls assertTransactionInputDepth before returning a branded prepared carrier. Direct Resource singleton preparation does the same at202–213, using wrapper-equivalent depth2.
- transaction-set.ts49–60 labels this as private execution capacity; depth greater than32 throws TransactionDevelopmentLimitError.
- transaction-submission.ts90–109 catches that marker, still validates the key/principal, but121 returns development-limit before compareAndAdmitTransaction130. Thus no authoritative row/identity comparison occurs.
- transaction-http.ts339–357 routes both direct batch and singleton through this path and emits private500. Its existing conflict path358ff already preserves same-entry context.
- transaction-store.ts1057–1085 already implements the right retained-key distinction: lookup/expiry, dependency-template comparison if necessary, otherwise normalization with retained alias witnesses, then conflict versus existing receipt. Fresh controls follow1087ff. The equal dependency path1074–1075 and ordinary equal path1085 must remain intact. Do not add a separate HTTP key lookup, fabricate a receipt, bypass unresolved-reservation comparison, or infer equality from a supplied key.
- The needed ordinary comparison machinery is already depth-independent in relevant source: transactionInputText (transaction-set.ts18–47) serializes decoded values iteratively; normalizeTransaction (transaction-batch.ts412–489) uses the existing tagged encodeSemanticValue; semantic-identity.ts53–63 explicitly uses iterative traversal/output with caller-owned byte bounds. Protocol json-admission.ts/json-values.ts and their existing deep-nesting tests corroborate that parsing/snapshotting has no new wire depth32 rule. This supports a bounded comparison repair, not a blanket proof that every execution/schema/serialization path tolerates arbitrary depth.

A naturally produced deep equal receipt cannot be demonstrated through today's depth-refusing entry point. Do not seed a fictitious installed receipt and claim a supported wire history. Preserve the real shallow equal-replay path and any existing source-owned saved-outcome comparison coverage; the decisive new wire oracle is a shallow retained key versus different deep metadata.

## Smallest safe next plan

Plan a source-private separation of syntax/prepared identity from execution capacity for **direct singleton and batch comparison**. Carry the private depth-capacity fact far enough for the existing authoritative owner to decide retained equality/conflict first, and stop an unknown key before any receipt/claim/effects. Preserve the current fresh-depth private500 marker as an explicitly remaining development limitation; do not invent a dotted limit, change it to413/503, advertise unlimited JSON depth, or remove the guard globally.

Likely production review footprint: transaction-batch.ts (prepared syntax/capacity fact), transaction-set.ts (depth analysis), transaction-submission.ts (remove premature return), transaction-store.ts (fresh-only capacity disposition and existing observed-context typing). Exact paths may shrink or require a small result-type propagation; that must be measured in the plan. HTTP already maps the development-limit and conflict results, so a new public HTTP contract is unnecessary. Keep complete syntax/principal ordering, unknown-key no-effects, finite comparison lifetime and negotiation ordering. Sequence preparation currently shares the depth guard (transaction-http.test.ts1620ff); preserve its existing private refusal/whole-carrier no-partial-effects behavior rather than silently changing sequence admission in this slice.

Required bounded proof: direct batch and singleton shallow success → same-key deep-metadata conflict → original equal replay; full original receipt/resource equality and unchanged position; fresh deep refusal/no key capture followed by shallow use of that fresh key; existing malformed deep shape/Selector, authorization, retained conflict-context and body-size boundary controls. A negative control restoring the early refusal should fail the named409 assertion. Use real installed transport for the final direct-carrier witness after a clean source/build binding. Exact requests, owners, bytes, selection and control counts belong in that future measured plan; none were executed or invented here.

## Cold usability and remaining holds

Real cold Transactional Scope/bdp.json/operations discovery still returns the deliberate private refusal (transaction-http.ts570–577), although owner-observed descriptor/directory construction already exists (transaction-service.ts253–278; transaction-discovery.ts24–69). This is missing product functionality, but simply wiring the builder or relaxing index.ts288–307 cumulative public admission remains unjustified. The depth repair removes one real obstruction toward that boundary; it is not a profile qualification.

Ordinary Update's private composed Read routing already exists (development-runtime.ts377–399,532–550); the uncomposed diagnostic root400–414 is not evidence that the ordinary authority or discovery builder needs replacing. PR129's overlap work is retained, not reopened.

S1-Q1 retained Snapshot/continuation minimum semantics and G2-Q1 incident-Link metadata Event membership remain pending. Existing minimum incompatibility/visibility residuals and configuration-dependent schema reachability remain disclosed as in the prior audit. No new ruling is inferred. Sequence v6 stays accepted plan-only and awaits the actual parameter merge reanchor; it remains a useful existing-contract evidence slice after this concrete serving repair is prioritized.

No product edits, imports, tests, builds, services, new reviews or publication were performed.

## Committed source SHA256 pins

- packages/server/src/transaction-batch.ts: 5fa7052d95942548727760625e368f3892e1dacced1654f022720af0b5ce899a
- packages/server/src/transaction-set.ts: 03b1ca9427c6963f5b6ce712a7f70540501e081ed635c742015d2bc49cc9524a
- packages/server/src/transaction-submission.ts: 795b45d87b6ee7f727a3727ace64c0e6569d0cfa7ad941e46740e285d04df942
- packages/server/src/transaction-store.ts: c0d1ebcdef1b354aa0806b8106ac0a3b0b24e6c94f50ec49b550decdf36ccf89
- packages/server/src/transaction-http.ts: b51d02801e5924190aba77c377e538db0b8619149d6184e19cb56d440f40c2cf
- packages/server/src/semantic-identity.ts: ab836972c8380cc21c1cca8e6d431faa743b5c31f4f9c578e15d340abec69224
- packages/protocol/src/json-admission.ts: fad4bd5159e7ff9a309d9c094c0c8beb950593a60d69f088971cb9bb6b84e76e
- packages/protocol/src/json-values.ts: 9a1cd0b12ccd41913e2874673fdb2a24ae7925d7e4e880f50cecc717dadec591
- docs/specs/bdp.md: f036103ffdf5f89fa3a2033570e1e987c4cd203cd80b1c633cb44615123f892e
