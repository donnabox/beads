# Native production-only code review

No actionable defect found in the eight frozen production changes. The implementation matches the accepted fresh-depth v2 plan's equality extraction, depth-only guard removal and JS receipt-maintenance predicate. This conclusion is scoped to code inspection; unfinished source/installed tests were not treated as frozen or reviewed, and no product execution occurred.

Binding SHA256:4b240cbedf60075c024e589ef0bdce4b2dc842c8192dc2d6c6efe930d8337b3d.
Production patch SHA256:debcb1ba570c4ca6c603aabbd0da1922467650ae9fce10fd9914e23c773865cd.
Author base77b4cf980ec91f66c52a554c3dd8c9de5aee859f, tree6a8af0f4b40216ce9c0b3f26957d293caf6f6627.
Accepted plan3dd6f2be/ledgera885b222/dispositionsa6aa1b7c. I previously reviewed that plan; I did not author this production patch.

## Findings by required boundary

- **Equality domain:** json-value-equality.ts1–28 is a byte-identical function-body extraction of the predecessor resource-evaluator equal function, verified by data-only source comparison. Both original no-op calls now invoke it. For admitted dense arrays/plain JSON records and finite native numbers it preserves key presence/count, array/object distinction, order-sensitive array elements, object-key-order independence and numeric -0/0 equality. It explicitly is not an arbitrary-JavaScript validator; sparse arrays, accessors, symbols, cycles and prototypes are outside the caller domain. Snapshot records come from JSON.parse plus existing Resource parsers/validators at snapshot324–359, not untrusted mutable JS objects supplied directly to this helper.

- **Complete owned-Link integrity:** snapshot406–411 replaces only the recursive full-record comparison. The explicit sorted ID inventory, exact owned-plane Type inventory, source ownership and endpoint/index checks remain. It still compares the entire inline Link against the first-class record; there is no projection that drops properties, metadata or other fields. Missing records remain unequal. The helper is imported only by the evaluator and Snapshot module; it is absent from index/development/package public export surfaces.

- **Receiving syntax and immutable brands:** batch217–230 retains unconditional object(decodeJsonDocument(text)) before singleton branch selection. Resource and set parsers, synthetic singleton discriminator exclusion, body-byte limits and complete carrier validation remain. Removing the second sequence raw decode removes only depth bookkeeping: prepareReadUpdateSequence already performs full lossless root/carrier validation. The prepared WeakSet brands and frozen carrier/member/operation/key arrays remain; protocol parsing already freezes admitted operation inputs and nested decoded values. Source-private preparation API names remain available.

- **Retained-key/sequence ordering:** store1137–1199 still performs retention, foreign-pending precedence and dependency comparison before ordinary normalized retained comparison, then unknown-key operation/Selector/Accept checks before inserting any receipt. Only the depth refusal branch was deleted. Direct owner operation/Selector limits still convert to TransactionDevelopmentLimitError. Bulk reservations and turn-time reclassification keep their actual admitKey calls and atomic ownership cleanup. No extra owner lookup, mutation, scheduler action, receiving reordering or admission bypass was added.

- **Result propagation:** submission removes exactly the impossible development-limit variant/case; current-context checks and all other refusal/receipt paths remain. HTTP removes only that direct mapping. Its sequence instanceof TransactionDevelopmentLimitError catch and generic rethrow are preserved, as planned, without treating the catch as proof of a live remaining HTTP capacity path. Selector depth32 and all other configured capacity constants are unchanged.

- **Receipt maintenance parity:** store545–560 keeps snapshots.expire first, SQL owner-null selection, then parseReceipt(...).detail==='available' before the unchanged expireRow. No new ordering, catch-and-skip, schema/index or format change. Pending owners are excluded; compact receipts stay bound. expireRow retains availability through the exact expiry instant, deletes only expired failed rows, compacts completed receipts preserving original namespace/positions/creation allocations, and updates the same counts. Startup and periodic/manual maintenance still use the same sweep inside existing transaction/cleanup/fencing paths. JS parsing replaces the SQLite JSON function that rejected the pinned deep receipt. Additional JS materialization and duplicate parsing of available receipts are an explicit implementation cost, not a new population/heap qualification claim.

## Verification boundary

The full eight-file patch and relevant complete parser/admission/sweep/expiry/Snapshot graph functions were read, along with protocol parsing/freezing and caller context. No AGENTS.md/CLAUDE.md/GEMINI.md was present in this checkout's file inventory. All eight production hashes matched the supplied freeze at the verification point during review and again immediately before this report; HEAD remained77b4. No test source was accepted as immutable. The author's stated format/typecheck outcomes are historical author evidence, not tests independently run here.

The planned deep receipt maintenance/reopen, owned-Link Snapshot positive/mismatch, full depth-safe oracles, unchanged receiving/retained behavior and emitted control still require actual frozen-code validation. This report does not approve a public profile, arbitrary nesting/resource usage or mutable test work.

## Preserved before/after production pins

Each following hash matched at both verification points:

- packages/server/src/transaction-set.ts: 3085632e897481b048d563cf0ba3ea66dc96a0662cebfde9f978ee00675519e5
- packages/server/src/transaction-batch.ts: 830ad57a680db7148fc75cd9cc10cadfc6a812448fd1704c1b0f47b13611808d
- packages/server/src/transaction-store.ts: 2766e7748a569d44821a0e8481cb5108cfdf1aaa86fca5126c865025a857eb85
- packages/server/src/transaction-submission.ts: 580a5f56075c3f7e2c041354eab75529dd304f86588337f725f4cb33429b33bf
- packages/server/src/transaction-http.ts: 81c130dd18f604ab1d13c03a2f52a98e11a6988f45d62af98b4b8c7499d14529
- packages/server/src/transaction-snapshot.ts: dac7c2f121cfb6bae557b3917f5ec52e47a358ab39da4062f6a5c26f5b9fbd83
- packages/server/src/resource-evaluator.ts: 6784629841a92baaeed648862cfa84d071356ef82e6d29250b254a9eddf793a2
- packages/server/src/json-value-equality.ts: f93e0bc3a090eeda9d40ec873c0024da2fbc368df5babe4f7d18cfa9b4780a66

No source/HEAD changes, tests, builds, imports, new review launch, main queries or Git mutations were performed.
