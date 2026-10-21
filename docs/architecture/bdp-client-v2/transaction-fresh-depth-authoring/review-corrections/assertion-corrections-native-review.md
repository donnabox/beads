# Native closure of six-file assertion corrections

No remaining concrete blocker found in the applied assertion-only delta. All six current files were reconstructed against the actual focused review patches and their pinned77b4 parent; each reconstructed predecessor matched its review binding. Current bytes match the two installed candidates, submission candidate and three other source candidates exactly. No source writes, product imports, tests, compiler probes or reviewer launches were performed.

Authorship overlap: I proposed the installed two-file assertion patch and previously reviewed these source witnesses. Root authored/applied the four source-test corrections and owns execution/council acceptance. This is source-grounded correction review, not independent runtime evidence.

## Actual focused review outcomes

Submission: exit0/448.6482934160158 seconds, no timeout, stdout bac5437b4b9cc886a71032638f70cebad4616985806216d81daea73e76f9aba5; exit record d94a5e412226e4135a75ab362ae2b293b40098ae467b834b520d93eebd0ff64e.
Other source: exit0/539.7089157078881 seconds, no timeout, stdout b37cb29304b00939ce05b8a9b2742f91643b8e4ecbfd254fef9a405bdcdb01e6; exit record1c986b8c4a7f6ae600a2d55a8a4eebc8458aa60149a7a561c6e073f537d90afd.
Both stderr files are empty, source-before/source-after snapshots match, and independent process inspection found neither PID/group71002 nor71000 present. The other still-running review's output was not inspected or modified.

## Submission findings

1. **Failed/expired freeze branches:** closed. The actual failed receipt is read through readReceipt, its failed status asserted, then independently traversed for Object.isFrozen at every object. The expired completed receipt is similarly read, checked and traversed; the previously issued page must return expired. These use observeReceipt/freezeReceiptValue rather than the legacy receipt() diagnostic. No fabricated receipt is supplied. The status/detail assertions prevent an absent receipt from passing the otherwise object-only traversal.

2. **Former-boundary set loop:** closed. Both27/28 depths and both singleton/wrapped carriers remain; the constant kind column is removed and actual receipt status completed plus matched count0 is required. Body/transport inputs are unchanged. A failed receipt can no longer satisfy this oracle.

3. **Position/identity preservation:** closed with the corrected interpretation. The zero-match operation binds its request key, but inspect().identities reads Resource IDs from SELECT id FROM identities (store2290–2294), populated by putResource (1300–1306), not the receipts key table. No resources are created by zero matches. Therefore unchanged position and complete unchanged Resource identities are right; the reviewer's proposed +1 identity assertion would be wrong. The existing subsequent same-key changed-body409 proves retained request ownership independently.

4. **No-op persisted values/key order:** closed without introducing recursive serialization. After no-op, the complete resources are compared using the independent iterative comparator, and each Resource's immediate properties keys are compared in enumeration order with the before snapshot. This detects the specific {deep,copy}→{copy,deep} rewrite, while keeping depth12000 safely outside recursive JSON.stringify/toEqual of whole resource graphs. Existing position/groups assertions remain. It does not pretend to assert arbitrary nested object-key enumeration beyond this concrete root replacement.

5–6. **Maintenance double parse/full-row materialization:** accurately scoped performance observations; deferred, already disclosed. No parsing cache, detail-column migration or new maintenance behavior is added to these assertions. Byte-bounded inputs and finite witnesses remain subject to actual runtime gates; this closure makes no throughput guarantee.

7. **Cycles:** no correction required. Helper inputs are parsed JSON graphs and fresh acyclic receipt/page wrappers. Iterative traversal includes children of already frozen parents. Public arbitrary cyclic-object admission is not part of this private freeze seam.

8. **Claimed native JSON.stringify deep-storage risk:** the premise is incorrect for the inspected storage path. transaction-store.ts imports transactionStorageText as json at67; transaction-batch.ts1065–1067 delegates it to protocol stringifyJsonValue. That function (json-values.ts101–146) snapshots and emits containers iteratively, invoking JSON.stringify only for primitive values and string keys. resource-evaluator.ts1024 likewise uses stringifyJsonValue(record). Thus this is not newly sending a12000-deep container to native JSON.stringify. Actual source/control/installed execution is still necessary to validate all exercised paths and budgets; source inspection is not substituted for it.

## Other-source findings

1. **Nullable owner type query:** rejected as a claimed typecheck failure; unchanged. owner is assigned open(options) before the corrupt conditional. Assignments to undefined occur inside the corrupt arm and do not flow into its else arm, where typeof owner.captureSnapshot appears. The assertion that this type query necessarily ignores the narrowed value is not borne out by the actual pinned compiler result. Root's preserved actual-successor.json records typecheck exit0; typecheck.log SHA91d54ffbee73a3ac72294d090e58d77745f8b93edbf94b862807698273a6505a is the tsc -p tsconfig.check.json run. That config includes packages/*/src/**/*.ts, including this exact unchanged Snapshot test/type query. The installed TypeScript7.0.2 launcher delegates to its native executable, so no fabricated JS-checker explanation or fresh compiler execution is claimed here. Later static validation remains its own successor attribution.

2. **Two-member HTTP processing:** closed. The original requests/keys are deliberately preserved. First deep sequence call creates a; second call processes retained-key conflict then creates b. Added ordered two-element body expectations require exactly that conflict/created shape; position2 remains correct. This demonstrates two members processed despite maximumOperations1 for direct transactions, not two fresh commits on the second call. The existing direct two-operation request still rejects413. No HTTP added.

3. **Snapshot integrity rollback:** closed. After corruption/capture refusal and owner closure, the existing single SQLite inspection checks snapshots, snapshot_records and snapshot_pages are all empty. No extra database handle or Snapshot capture is added.

4. **Stale label:** closed. Only the failure label becomes conflict-before-negotiation, matching the existing409-before406 behavior. No protocol or oracle value changes.

5. **Pure equality branches:** closed by two assertions: array/object mismatch and nested differing primitive. They reuse the existing callback, add no owner/HTTP and leave the numeric-spellings fixture intact.

6. **Idioms/titles:** optional nits deferred. Bare assertions still throw on failure; the adjacent depth values remain literal examples of admitted former-boundary inputs; existing rollback/cancellation callbacks retain their evidence. No new callbacks or renamed selections are necessary.

## Installed delta and preserved scope

The two installed files exactly equal my prior proposed candidates: four checked deep literals per file; direct ordered contextPosition table equality; two sequence200 media types; two absent headers on changed-key direct409. Prior F3 remains explicitly limited: the eight-request cohort does not independently read back its sequence-b41-level metadata. New13 Snapshot/source reopen evidence covers sequence-created deep properties separately. The prior full disposition remains in claude-installed-existing/native-findings-audit.md; no ninth sequence request is added.

Current six hashes:
- apps/bdptest/src/transactional-depth-key-order-installed.test.ts: f4e19ef9b7aba1a12f3376540257b9680e27dcd00936048b066eb7667ebe19fe
- apps/bdptest/src/transactional-sequence-depth-installed.test.ts: 0a2a765cab6137f324cffb7ff7d96d1f3e81fef3a73c14f56adfd36309083aa7
- packages/server/src/transaction-submission.test.ts: 026cf5ccc65026f0f6462deaba9e1e58cd7ab6c6ae341ef5ed3af0becbb0c182
- packages/server/src/transaction-http.test.ts: 9eb452b8e9d80b53f28cd61b40385886c3a61a992aadb1c9c201d9a188081cce
- packages/server/src/transaction-snapshot.test.ts: b6c31d8b66650ec0668c85d1fb803dda1072ef463989299ed377dd18dcbfee98
- packages/server/src/resource-evaluator.test.ts: 2a35e3add8eabe18d27a69c173b709e1d958c16191c049a4877d92a7b27fa7f1

The reviewed store remains f7fe71da918c135eb43cb4659f186df6cd442aed9c42c2835e95cde05c782a9e; unchanged set/sequence source hashes remain aec0ce27/5526903a. Request bodies, route order, selectors, limits, source42cases/47HTTP and installed34HTTP/three children are unchanged. The correction adds two source readReceipt calls and one expired readReceiptPage call, plus two queries on an already-open offline inspection handle; no new owner, listener, child, case or directory. These source operations must be reflected in final measurement attribution rather than described as literally no additional work.

Remaining gates are factual, not new findings: freeze/rebind final source/measurements and preparation, actual static/source/control/fresh installed validation, and disposition of the separate new-installed review after terminal. No current product pass or public qualification is asserted.
