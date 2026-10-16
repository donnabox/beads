# Fresh-depth serving audit

Recommendation: make bounded fresh deep-value Transactional execution the next serving slice after the separately reviewed retained-sequence ordering repair. In the inspected paths I found **no recursive arbitrary-JSON traversal that explains the fixed depth 32 refusal**. The guard is an explicit private execution-capacity fence, not an existing protocol JSON-depth limit. This supports a focused capacity implementation/evidence plan, not immediate unvalidated guard deletion or a generic 413 substitution.

Basis: committed `df80f8cb86b7763bc7bb005492e43d8840527045`, plus the frozen, unmerged seven-path retained-sequence draft in `/Users/dbox/repos/wt/vickie/vickie-sequence-depth-retained`. Line references below are to these actual files. All 29 review context hashes were rechecked unchanged before writing. No product imports, tests, services or source writes were performed for this audit. Existing tests cited below were read, not rerun. Initial isolated validation stopped at lint; it supplies no fresh-depth execution evidence.

## What the guard actually protects today

`transaction-set.ts:52–61` walks input iteratively and throws the private development-depth exception above 32, including scalar leaves. `transaction-batch.ts:358–380` records that fact on a branded carrier; strict preparation requires it immediately. Singleton receiving uses offset 2 (`:213–237`), batch parses the actual carrier, and the pending sequence repair uses the original member at offset 2. The number is not derived from a serializer's stack capacity.

`transaction-store.ts:1133–1193` checks retained/dependent state before a fresh-key depth refusal; comparison uses the iterative semantic encoder. No new receipt is inserted on that refusal. Sequence bulk admission (`:1957–2031`) is atomic; eligible turns (`:2034` onward) resolve dependencies and reclassify state. These boundaries, retained expiry behavior, and labels/aliases must survive any later capacity change. Do not add a second preflight or turn a private exception into a protocol refusal without the correct admission/execution phase.

## Complete value path examined

| Path | Source evidence and consequence |
|---|---|
| Parse, numeric admission, carrier shape | `protocol/json-admission.ts:98`, `:231–299` use explicit parser/work stacks. `protocol/read-update-values.ts:203–215` maps numeric placeholders then uses AJV for carrier structure. The canonical schema's `properties`/`metadata` (`schemas/bdp-v0.schema.json:766–773`) permit arbitrary members; patch value is `true` (`:1294–1311`). Thus this structural AJV boundary does not recursively validate arbitrary property payloads. |
| Normalization and pending dependencies | `transaction-batch.ts:479–558` delegates arbitrary values to `semantic-identity.ts:59` onward, an explicit work stack. Native JSON.stringify there encodes scalar tokens or shallow witness records. `transaction-dependencies.ts:118–171,173–218,220–354` reuses normalization and stores the deep identity as a string, not a nested template. Store turn-time `isDeepStrictEqual` compares that shallow template, not the Resource document. |
| Create, update, patch append/copy, no-op | All Resource operations reach `evaluateResourceMutation` via `transaction-batch.ts:728–775`. `resource-evaluator.ts:383–447` equality/copy/freeze use work lists; `:448–532` walks pointer segments, copies inserted/replaced values iteratively, and handles array append/splice. `:797–812` shares this for properties and metadata. The recursive `materializeVersion` at `:1124–1166` follows owning-Bead→Link recipes, not arbitrary JSON nesting; Link recipes do not recursively own further Links. |
| Sets and batches | `transaction-batch.ts:899–1003` admits numbers, selects current Resources, checks cardinality/match capacity and invokes that same Resource evaluator per match; Selector evaluation itself uses frames (`read-selector.ts:771–810`). Ordered batch operations share the owner's effects SAVEPOINT. No separate recursive deep-payload set evaluator was found. |
| Errors and schema validation | `resource-evaluator.ts:180–225` retains bounded closed diagnostics (strings and fixed fields), so its JSON.stringify does not traverse the original deep value. Numeric diagnostics use the iterative admission walker. The *installed Type* evaluator is separate from structural AJV: `development-contracts.ts:98–118` serializes with stringifyJsonValue and calls the explicit generator-stack evaluator (`installed-schema-evaluator.ts:989–1018`). It can still refuse finite work/frames/diagnostic budgets; see caveat below. |
| Persistence, receipt and restart | `transaction-batch.ts:1103–1105` storage serializer delegates to iterative `protocol/json-values.ts:35–141`. Store Resource bodies, group and receipt writes (`transaction-store.ts:1300–1311,1376,1396`) use this path; native JSON.parse without a reviver restores bodies/receipts (`:463–464`, `resource-evaluator.ts:353–380`). Receipt page plans slice/count result entries (`store:794–861`), not nested JSON. The effects SAVEPOINT is rolled back for normal failed outcomes, then the failed receipt is retained (`:1346–1360`); arbitrary thrown faults take the separate rollback/claim-cleanup path. |
| Read/projection/Event/stream output | Authorization projection maps result/allocated envelopes (`transaction-authorization.ts:79–118`). Event and Snapshot deep freezes use work lists (`transaction-events.ts:279–289`, `transaction-snapshot.ts:130–141`); retained Event equality uses stringifyJsonValue (`events:219`). HTTP delivery and HEAD measurement use it (`transaction-runtime.ts:88–95`, `read-http.ts:40–44`), as do SSE frames (`transaction-sse.ts:39–44`). No native JSON.stringify of the arbitrary Resource body was found on these paths. |

Existing `resource-evaluator.test.ts:1439–1564` includes actual depth-12,000 Resource persistence/reopen, no-op and deep path updates; it explicitly demonstrates native JSON.stringify throwing while the shared evaluator succeeds. This is strong evidence that depth-safe infrastructure already exists, but it is not a fresh Transactional receipt/Event/restart witness and does not prove every installed contract fits its work budget.

## Limits and unresolved proof, not a new depth ruling

The normative spec `docs/specs/bdp.md:1043–1070` distinguishes request bytes, Resource properties/representation bytes and **Selector** depth/nodes. Omitted bounds do not permit non-normative failures. Property-change law (`:2200–2217`) permits only add/replace/remove, validates the whole resulting properties object, gives metadata no Type schema, and distinguishes patch path depth from value nesting. A root replacement or array append can carry deeply nested JSON with path depth zero/one. Reusing `selector.depth` or patch-path limits to reject that value would be wrong.

Receiving order (`:4515–4560`) already specifies pre-parse request-too-large, syntax/authentication, retained comparison, fresh admission controls, and negotiation. Resource limits and validation are checked in the shared evaluator (`resource-evaluator.ts:1030–1070`), producing the existing failed outcome through the effects rollback, not a fresh-key preflight refusal. The private runtime currently has 64 KiB request/properties and 256 KiB representation limits (`transaction-batch.ts:63–73`); deep values can fit those byte bounds. Do not replace private depth500 with request-too-large for a body under its byte bound, or consume a key on a receiving refusal.

Independent caveat: installed schema ceilings (`installed-schema-value.ts:4–19`) bound nodes, frames, logical bytes, work and diagnostics. `development-contracts.ts:108–118` turns an unevaluated/incomplete result into a private error. A complex contract or a large invalid-instance diagnostic can hit that limit even at shallow depth; the 32 fence neither derives from nor solves it. Removing depth32 must not be advertised as solving all Type execution capacity or full-profile qualification. Ordinary protocol shape checks do not have the same instance-recursion issue because their open JSON payload fields are unconstrained. This audit did not execute schema worst cases, prove maximum heap/time for every 64 KiB input, or establish arbitrary custom-validator safety.

## Smallest sound next proposal

Plan one bounded fresh-depth support change using the existing iterative machinery and existing byte/patch/resource limits. Inventory every strict/receiving entry and fresh admission depth fact before choosing final hunks; do not leave a singleton/set/sequence exception path that still rejects valid deep content. Keep the current retained-sequence repair separate. Acceptance should cover actual deep fresh creation, an update including append/root replacement and semantic no-op, a selected-set update and dependent sequence using that same deep value; bind retained replay/conflict, receipt/page/current Resource/Event output and durable reopen to the original content. Include a late existing validation/resource-limit failure after an earlier staged effect to prove rollback plus correctly retained failed outcome. Reuse existing fixtures and avoid raising capacity fences, broad evaluator rewrites or new semantics.

This is a next-slice recommendation, not an executable plan or removal authorization. No unavoidable product ruling was identified for supporting valid deep values within existing bounds. Introducing an arbitrary JSON nesting limit or relabeling its refusal would require a separate normative decision. Final plan must measure a small representative witness and include schema-budget limitations honestly; source inspection alone does not justify claiming the guard can already be removed safely.

## Exact inspected-file pins

Paths below are relative to the workspace above; SHA-256 binds the cited bytes (two production files include the pending sequence patch).

- `docs/specs/bdp.md`: `63e93be562334c84c1e967da2fce212e8ef515aa3c9448924c5f9a272ea7b80b`
- `packages/server/src/transaction-set.ts`: `03b1ca9427c6963f5b6ce712a7f70540501e081ed635c742015d2bc49cc9524a`
- `packages/server/src/transaction-batch.ts`: `95f3ce3957503cba240f18e911fe3d2c8027b4accbba8f1128ea47e3d2b0c362`
- `packages/server/src/transaction-store.ts`: `bc10df9d02316e395b8c4ea7db6134ab2646c8e8abd83ea785b0ea09f2f0c82c`
- `packages/server/src/transaction-dependencies.ts`: `ca79b4be097d221560157c5c4eaac4601dab343de54dda1be15d92c44b2b09ba`
- `packages/server/src/semantic-identity.ts`: `ab836972c8380cc21c1cca8e6d431faa743b5c31f4f9c578e15d340abec69224`
- `packages/server/src/resource-evaluator.ts`: `89606787e455f4ed555b807d91c5331a376431807a9a581de1f94e538fcab4ff`
- `packages/server/src/resource-evaluator.test.ts`: `b06fa79f691449573d63a5273191c7df06a21c89418c6080895926bfac6bf68e`
- `packages/server/src/development-contracts.ts`: `182a1a6d06b09f08ca3c83ad0fb27a09a419ad3e99729528381f210973b2c40a`
- `packages/server/src/installed-schema-evaluator.ts`: `20662b8fb5457d4233e2cc22e35152aa43a2033b7db3568d42847a8752774544`
- `packages/server/src/installed-schema-value.ts`: `ba14d1a1ef75b13ac14ea20c3066b3f8b742a05d50ac8899fd7b456fa94b067c`
- `packages/server/src/transaction-events.ts`: `6c5ab4482f0b00110a71a96b16d44ca5fe713cb5f397ff78778cc6646f632a3a`
- `packages/server/src/transaction-snapshot.ts`: `d74aad937b0a07ea2878ef2c275cec6dbdb033cf112006e2569d64537520b15c`
- `packages/server/src/transaction-authorization.ts`: `c87a7bfa8840aee2f097bf683fe3f3c97fcae80782020c6d5b1e41364d16b6d2`
- `packages/server/src/transaction-runtime.ts`: `48cb5411407c456d7b3f301a4b67970061be9f51c94201de832ba5deff6ea2d4`
- `packages/server/src/transaction-sse.ts`: `1a414cb252b1eb2b5c3695d16aecb59382cbd417876837351e1d80b91b45784c`
- `packages/protocol/src/json-admission.ts`: `fad4bd5159e7ff9a309d9c094c0c8beb950593a60d69f088971cb9bb6b84e76e`
- `packages/protocol/src/json-values.ts`: `9a1cd0b12ccd41913e2874673fdb2a24ae7925d7e4e880f50cecc717dadec591`
- `packages/protocol/src/read-update-values.ts`: `d1fdbc660d5e2c009275c152bdcebda46031700920aa03bfea89f39564c89de2`
- `packages/protocol/schemas/bdp-v0.schema.json`: `0baa5d5ef67144f8e6513b4cde31168effc8df6d52ef3141b0f144b87f8dc227`
