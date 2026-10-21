# TX sequence Resource-alias serving plan

Read-only plan, 2026-10-14. Basis: main `581bad0de6f5eb098b920147f6a7e09809c553ea` (PR102); inspected runtime bytes are unchanged from627b/61ed. No source edits or tests were performed. This is a runtime-serving slice, not a coverage increment or profile promotion.

## Recommendation

Implement deferred alias normalization for newly reserved sequence members, reusing TX's existing durable reservation/turn machinery and existing alias witness normalizer. Then remove the specific preparation/HTTP refusal. Do **not** resolve aliases once before admission, merely remove the refusal, or run the RU lifecycle against the TX store.

The smallest safe slice necessarily includes pending-key comparison, not just normal execution: a direct singleton or one-operation batch may present the key reserved for a not-yet-reached alias member. It must not be told that the current alias value is the original execution's final semantic identity.

## Existing law and exact current obstruction

`docs/specs/bdp.md`:

- Alias operands and semantic identity: lines2354ff and3202–3243; table rows6760–6761. A Bead subject or Link endpoint uses the alias's target when its member is reached. Store canonical endpoints. Retained original locators keep their target **or miss** after repoint/delete; previously unseen retry locators resolve afresh in the retry turn, and comparison-only witnesses must not rewrite original metadata. Pins remain opaque and byte-exact after URI resolution.
- Whole sequence syntax, backward bindings and separate commitment:2784–2890; atomic all-key claiming and per-turn classification:3258–3282.
- TX uses the same semantic identities and one key namespace across carriers:3882–3940. Direct callers cannot receive a pending receipt without established equality.
- Existing unresolved-reservation direct comparison:3994–4025, finite single deadline and no DB transaction held while waiting. TX sequence semantics/projection:4047–4116; foreign-pending or transient-creator projection wins without waiting or dependent-key comparison.
- Alias writes remain locator-only, separate commits, not batch operations/history:4189–4214.

Current code:

1. `transaction-batch.ts:229–291` validates the entire carrier and prepares each branded singleton member. At265–268 it rejects Resource alias operands (`bead` on update/delete and `source`/`target` on createLink). `putAlias target:@name` already works; it is a different case.
2. `transaction-http.ts:323–332` maps that specific exception to a nonattesting500 `sequence-resource-alias` development gap. `transaction-http.test.ts:424ff` and `transaction-sequence.test.ts:439ff` explicitly test this limitation.
3. `transaction-store.ts:1720–1789` reserves every unknown member atomically. A member without `@name` dependencies calls `admitKey` immediately; `admitKey:1070` calls `normalizeTransaction`, which performs live alias lookup. This is **too early** for sequence aliases.
4. `nextSequence:1791–1875` already resolves deferred creator bindings at a member's turn, commits normalized identity/witnesses under the exact owner fence, then executes the real evaluator. Retained/non-owned keys are already compared in their eligible turn; foreign-pending keys take precedence.
5. `transaction-dependencies.ts` currently represents creator holes only. Simply feeding it alias-bearing requests is wrong: `dependencyTemplate` calls `normalizeTransaction(..., () => undefined)`, turning an unresolved alias into a recorded miss; `compareDependencyTemplate` later resolves current aliases and would produce false conflicts or premature equality.

## Reuse and bounded design

### A. Describe deferred references; retain carrier validation

Within the existing source-private preparation/dependency modules, record eligible canonical alias locators and their operand slots without store access. Preserve the already branded operation, raw numeric values and pins. A prepared member needs deferred turn normalization if it has creator dependencies **or** eligible alias operands.

Do not widen valid reference slots: an alias as a Link subject remains the existing member-level resource-not-found case; wrong-kind `@name`, forward bindings, malformed references and duplicate keys remain whole-carrier syntax errors before claims. `putAlias.target` must remain canonical Bead or valid backward Bead binding, never alias-chain resolution.

### B. Reserve now, resolve only in the eligible turn

Extend the current private pending reservation template to preserve immutable surrounding semantic fields plus unresolved alias-slot/locator relationships. Reuse the existing receipts/dependency column, exclusive owner fence and all-key admission transaction. No new public fields or storage table is necessary.

At `nextSequence`, first preserve existing transient-creator/foreign-pending precedence. For a newly owned eligible member, resolve stable creator bindings and capture all alias targets/misses at that turn, then persist `normalizeTransaction`'s existing identity and witnesses before real execution. Reuse its per-operation locator map so one locator appearing in both endpoint slots is read consistently. A mixed `source:@made,target:alias/work` member must resolve both using their respective rules.

Normalization and execution must remain within one serialized owner turn, with no await/unlocked external work between capture and evaluator use. Prefer the existing fenced resolution path; if an implementation introduces any asynchronous seam here, move capture into the execution transaction instead. The evaluator must consume the captured alias overlay, not look up the alias again.

### C. Compare pending reservations honestly

Extend `compareDependencyTemplate` rather than add a second comparison loop:

- Mask only unresolved URI values, never operation kind, pin, expectedRevision, properties, metadata, change order or provenance. Reject statically provable mismatches immediately.
- Preserve correlation constraints: if both original endpoints use the same unresolved alias, a direct candidate with two different resolved canonical targets is already inconsistent. Distinct original locators are not assumed equal. Existing creator-kind/supplied-ID/correlation checks remain.
- If any required original alias target/miss remains unresolved and static comparison is compatible, return existing `comparison-pending`; never join the pending receipt merely because locator spelling matches, and never record a current lookup as the owner's future witness.
- Once the original member reaches its turn, the pending row is replaced with ordinary normalized identity and existing witness JSON; the existing finite comparison loop retries against that state. Retraction, epoch change and deadline exhaustion retain their existing handling.
- Sequence duplicate presentation stays nonwaiting and projects foreign-pending even when static fields differ. No new transient code or timeout is introduced.

This applies the existing unresolved-reservation safety law to the alias-induced unresolved state. It is an implementation inference required to reconcile turn-time resolution with atomic sequence admission, not a new wire semantic. Council should explicitly check this inference, especially partial-comparison correlation, before implementation.

### D. Keep final witnesses, receipts and RU untouched

`normalizeTransaction` (`transaction-batch.ts:389–465`) already implements original locator reuse, current lookup for new locators, distinct miss identity, canonical direct comparison and comparison-only witness behavior. Reuse it without replacing it with `normalizePreparedMemberIdentity`: the RU helper resolves committed member bindings, whereas TX batch creator indices remain symbolic.

Final identity/witness encoding, receipt projection and successful tombstone retention should remain unchanged. The new private deferred-template representation requires an explicit compatibility decision in review: prefer a tagged additive pending form with proof that recovery retracts it before interpretation; if old reader safety cannot be established, bump the private store compatibility identity and add refusal-before-maintenance tests. Do not silently claim compatibility or create an unrelated migration. No RU store, evaluator, public SDK/config or normative schema change is planned.

## Proposed file ownership / dependencies

Production files (4):

- `packages/server/src/transaction-batch.ts`: preparation metadata and removal of refusal only after deferred machinery exists; reuse normalizer.
- `packages/server/src/transaction-dependencies.ts`: deferred alias representation and conservative exact partial comparison.
- `packages/server/src/transaction-store.ts`: atomic reservation, fenced per-turn normalization and compatibility decision if required.
- `packages/server/src/transaction-http.ts`: remove only the now-obsolete Resource-alias development-gap branch.

Tests (bounded4): `transaction-sequence.test.ts`, `transaction-dependencies.test.ts`, `transaction-http.test.ts`, plus a focused new `transaction-sequence-alias.test.ts` if the corpus is clearer separately. Prefer existing emitted-runtime fixture/reopen helpers and stepper/scheduler, not new public test hooks. Existing comparison/recovery/runtime suites are regression targets; their source need not change unless a focused assertion belongs there.

No change to `resource-evaluator.ts`, `alias-evaluator.ts`, RU lifecycle, Type compiler, discovery, profile admission, installed configuration or conformance artifacts is expected. Report any required expansion before editing. Coordinate transaction-store ownership with the TX runtime author; cold-discovery author owns runtime/composition and should not concurrently edit this store/comparison path.

## Required discriminating observations

### Actual HTTP and durable reopen

1. SeedA/B; sequence `putAlias(work,A)`, update via work, repoint work→B, second update via work. Assert both canonical Resource IDs and independently expected distinct final A/B state, operation indices and valid sequence schema. This fails an admission-time alias snapshot.
2. Start with work absent; sequence createA + putAlias(target:@made) + createLink(source:alias/work,target:pinned alias/work). Assert canonical endpoints and unchanged opaque pin; real GET of Link and Bead afterstates. Cover relative and absolute equivalent alias spelling without duplicating every transport case.
3. Delete alias, then alias-operand member, then independent tail: member retains resource-not-found and tail commits. Later recreate alias; same-key replay still returns the original failure, fresh key now acts on the new target. Cross-carrier retry of that member returns the same failed receipt; close/reopen and repeat.
4. Successful alias member: repoint and delete original locator after success; retained retry continues to return the old postimage/identity, while current GET proves no extra mutation. Compare direct canonical and newly introduced alias locators pointing to the same target; repoint that new locator and retry again to prove it was not added to original witnesses. A distinct missing locator conflicts with an originally missing locator. Include pin-only mismatch conflict. Repeat core success/miss witnesses after reopen.
5. Existing malformed whole-carrier cases retain no prefix effects or key capture. Alias-as-Link-subject remains resource-not-found rather than becoming supported; putAlias-to-alias remains its existing validation failure. No semantic expansion disguised as gap closure.

### Owner/concurrency/recovery discriminators

6. Existing stepper: admit whole sequence while work→A, externally repoint→B before eligible member step; update must hitB. Alias lookup spy, if available at private normalizer unit level, records zero reads at carrier preparation/admission and exactly captured reads in the execution turn. Do not add runtime hooks solely for this count.
7. Direct same-key singleton and one-op batch race a reserved alias member: immutable-field mismatch409 immediately; compatible request waits without holding DB transaction, then yields same receipt or conflict after the actual turn target is known. A wrong currently matching target must not prematurely obtain pending receipt. Use existing bounded comparison scheduler/deadline, not sleeps; owner can progress/repoint while waiting.
8. Same sequence duplicate remains transient for foreign pending keys, including mismatching immutable fields. Mixed transient `@creator` + alias member must not consult its key or alias, and releases only its own reservation. Preserve all existing creator-attempt reuse/expiry/fencing tests.
9. Real emitted crash after earlier alias repoint/prefix, before later alias member: recovery retracts abandoned pending tail, does not execute it, preserves prefix. Explicit resubmission resolves the later alias in its new turn and succeeds. Existing process fixture pattern at `transaction-sequence.test.ts:469ff` can be reused with disposable files and bounded child cleanup. Crash/fault after turn normalization must neither commit a second mutation nor retain an incorrect witness; exact-owner cleanup/retry is checked.

Counterfactuals must distinguish (a) removing only refusal, (b) snapshotting aliases at whole admission, (c) resolving a recorded locator on every retry, (d) persisting fresh retry witnesses, and (e) equating unresolved direct callers by spelling. These can be permanent positive/negative test sequences, not production mutants or synthetic responses.

## Validation and done-when

Build emitted code before HTTP/process tests, pinned Node24.16/pnpm11.20, maxWorkers2, external logs and bounded child/process-group cleanup. Run focused TX batch/sequence/dependency/comparison/submission/store/recovery/HTTP tests affected by the changes, then typecheck/build/boundaries/format/lint and relevant installed smoke. Reuse the existing source-private installed TX launcher for a fresh packed HTTP alias success/replay/reopen smoke if normal test entry cannot prove packed payload equivalence; no evidence profile relabeling.

Targeted RU regression: `member-identity`, `member-executor`, `sequence-executor` alias/retained-reference tests plus the existing installed RU alias/reference and sequence-retry positive scenarios. Since shared evaluator source is unchanged, a full8k or ever-growing conformance cohort is not the first validation step. Broaden only for an actual shared-file change/failure.

Machine-checkable completion:

- Existing valid alias sequence no longer returns `sequence-resource-alias`500; original malformed cases still fail before writes/key capture.
- Each above semantic discriminator has an executable test with actual afterstate; HTTP tests schema-check response envelopes/receipts and assert positive status before equality.
- All-key reservation remains one atomic step; alias capture occurs in the eligible turn; immutable final identity/witness and comparison-only retry proofs pass before and after reopen.
- Direct unresolved comparison uses the existing one finite budget/no DB lock; sequence projection remains nonwaiting.
- Existing batch, singleton alias, creator dependencies and owned-Link behavior pass regressions; no new public profile or configuration capability is advertised.
- Clean bounded commit/report identifies any private compatibility change and precise remaining scope. No whole-contract coverage credit is inferred from these runtime tests.

## Cold-discovery interaction and limits

This is independent of cold discovery implementation. The current HTTP fixture knows private targets; closing this gap makes `sequence` functional for alias operands but does not prove a cold client can discover it. Cold discovery can point at the same existing sequence handler once its route table is published; it must not advertise a different weaker operation contract. Reuse immutable configured Scope/Host and actual owner. Do not merge stores or hide this new path behind an alternate mock launcher.

No product ruling is needed for turn-time alias resolution, retained witnesses, or separate commitment: these are explicit existing laws. Review is required for the private deferred-template/partial-comparison implementation and compatibility boundary. Authorization remains the currently supported all-visible TX policy; this slice does not claim mutable permission/erasure coverage. Selector timing, minimum-position controls, Event/Snapshot questions, alias enumeration and public activation are outside scope.
