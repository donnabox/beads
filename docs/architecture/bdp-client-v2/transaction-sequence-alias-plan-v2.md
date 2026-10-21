# TX sequence Resource-alias serving plan v2

Read-only plan, 2026-10-14. Basis: main `581bad0de6f5eb098b920147f6a7e09809c553ea` (PR102); inspected runtime bytes are unchanged from627b/61ed. Revision of preserved plan72b247 following native P2 and all seven external findings. Root will assign the exact implementation base after C1 lands. No source edits or tests were performed. This is a runtime-serving slice, not a coverage increment or profile promotion.

## Recommendation

Implement deferred alias normalization for newly reserved sequence members, reusing TX's existing durable reservation/turn machinery and existing alias witness normalizer. Then remove the specific preparation/HTTP refusal. Do **not** resolve aliases once before admission, merely remove the refusal, or run the RU lifecycle against the TX store.

The smallest safe slice necessarily includes pending-key comparison, not just normal execution: a direct singleton or one-operation batch may present the key reserved for a not-yet-reached alias member. It must not be told that the current alias value is the original execution's final semantic identity.

## Existing law and exact current obstruction

`docs/specs/bdp.md`:

- Alias operands and semantic identity: operative lines3062–3080 and3202–3243; table rows6760–6761 (2354 is only a forward pointer). A Bead subject or Link endpoint uses the alias's target when its member is reached. Store canonical endpoints. Retained original locators keep their target **or miss** after repoint/delete; previously unseen retry locators resolve afresh in the retry turn, and comparison-only witnesses must not rewrite original metadata. Pins remain opaque and byte-exact after URI resolution.
- Whole sequence syntax, backward bindings and separate commitment:2784–2890; atomic all-key claiming and per-turn classification:3258–3282.
- TX uses the same semantic identities and one key namespace across carriers:3882–3940. Direct callers cannot receive a pending receipt without established equality.
- Existing unresolved-reservation direct comparison:3994–4025, finite single deadline and no DB transaction held while waiting. TX sequence semantics/projection:4047–4116; foreign-pending or transient-creator projection wins without waiting or dependent-key comparison.
- Alias writes remain locator-only, separate commits, not batch operations/history:4189–4214.

Current code:

1. `transaction-batch.ts:229–293` validates the entire carrier and prepares each branded singleton member. At264–268 it rejects Resource alias operands (`bead` on update/delete and `source`/`target` on createLink). `putAlias target:@name` already works; it is a different case.
2. `transaction-http.ts:323–332` maps that specific exception to a nonattesting500 `sequence-resource-alias` development gap. `transaction-http.test.ts:424ff` and `transaction-sequence.test.ts:439ff` explicitly test this limitation.
3. `transaction-store.ts:1720–1789` reserves every unknown member atomically. A member without `@name` dependencies calls `admitKey` immediately; `admitKey:1070` calls `normalizeTransaction`, which performs live alias lookup. This is **too early** for sequence aliases.
4. `nextSequence:1791–1913` already resolves deferred creator bindings at a member's turn, commits normalized identity/witnesses under the exact owner fence, then executes the real evaluator. Retained/non-owned keys are already compared in their eligible turn; foreign-pending keys take precedence.
5. `transaction-dependencies.ts` currently represents creator holes only. Simply feeding it alias-bearing requests is wrong: `dependencyTemplate` calls `normalizeTransaction(..., () => undefined)`, turning an unresolved alias into a recorded miss; `compareDependencyTemplate` later resolves current aliases and would produce false conflicts or premature equality.

## Reuse and bounded design

### A. Describe deferred references; retain carrier validation

Within `prepareTransactionSequence`, publish one immutable prepared deferral fact: existing creator dependencies plus canonical alias operands `{ field, locator }`. Both admission and turn execution consume this same prepared fact; do not independently recompute `member.dependencies.length` at the two store sites. Preserve branded operation, lossless numbers and pins without store access.

The exact permitted alias slots are `updateBead.bead`, `deleteBead.bead`, and `createLink.source/target` (bare or Pinned Reference URI), per3062–3080 and existing `aliasFields`. An alias as `link` subject remains member-level `resource-not-found`; `putAlias.target` does not admit an alias and remains `validation-failed`. Do not confuse this with already supported `putAlias.target:@made`. Wrong-kind/forward `@name`, malformed shapes and duplicate keys retain whole-carrier rejection before effects.

### B. Reserve now, normalize every deferred owned row in its eligible turn

A member is deferred when its shared prepared fact has creator dependencies **or** alias operands. Admission writes the existing pending receipt/owner and a tagged pending-only template in the existing dependency column, with `identity=""`, `witnesses=""` and in-memory `normalized===undefined`. Reserve all unknown keys in the same existing atomic pass. No alias target/miss lookup is permitted here.

At `nextSequence`, first resolve creator facts from the prepared member to preserve transient-creator precedence **before consulting the dependent key**. If transient, preserve the existing release-only-owned-reservation behavior. For an eligible newly owned claim, read its exact fenced durable row: `dependency!==null` requires turn normalization, regardless of creator-dependency count. Validate the stored form and its consistency with the immutable prepared fact; do not silently skip a deferred row. Missing/malformed/inconsistent form is an owner-integrity failure, not a mutation failure receipt.

The eligible turn resolves stable creators and captures every alias target/miss through the existing `normalizeTransaction`. Persist final identity/witnesses and clear dependency under the exact owner fence. The existing execute precondition remains: no owned claim with non-null dependency or absent normalized state may reach successful `executeClaim`. Add the alias-only store test that failed under the old creator-only predicate, plus mixed creator+alias cases. Preserve alias-bearing createLink creator-fact writeback through the end of `nextSequence:1791–1913`.

Keep the existing synchronous path from fenced normalization through `executeClaim`; introduce **no await or asynchronous fallback**. Existing exclusive SQLite ownership, synchronous calls and `enter()` reentrancy checks preserve one owner turn. The evaluator consumes captured aliases rather than resolving them a second time. A crash at existing `after-dependency-resolution` retracts the still-pending receipt on reopen; no operation is resumed automatically. Do not restructure shared direct-carrier execution to move normalization into it.

### C. Partial comparison uses correlated symbols before any lookup

Extend the existing private partial comparer and reuse its existing finite comparison loop; no second wait loop or new wire state.

1. Build the set of canonical locators recorded as unresolved aliases in the **original template**. Each locator denotes one correlated unknown turn-time target-or-miss, not its current live target. On the candidate side, **any eligible operand occurrence of any locator in this original set uses that same symbol, including different reference slots and equivalent relative/absolute spellings**. Never call `resolveAlias` for such a locator, even when it appears only as a candidate occurrence in that slot. This restriction precedes normalization/masking/correlation checks.
2. Direct canonical references are fixed facts. A genuinely new candidate alias locator (not in the original set) resolves live for this comparison attempt only. Its canonical target or locator-specific miss is a fixed comparison fact for the attempt, never an owner witness. The final retry normalizer will re-resolve unrecorded locators on later attempts as existing law requires. Original creator dependencies retain their attempt/slot facts, kinds and supplied-ID constraints.
3. Compare all immutable non-URI structure exactly: operation, property/metadata values, patch order, provenance, expectedRevision, endpoint representation shape and every pin. A hole masks **only** its unresolved URI value. Preserve exact remaining fixed Resource references. No masked identity is persisted as the final identity.
4. Unify corresponding URI constraints rather than comparing eagerly resolved strings. Repeated symbols correlate across **both** original and candidate slots. One symbol cannot equal two distinct fixed canonical targets. Distinct symbols are not assumed equal or unequal: equality constraints may be satisfiable by a common future Bead. Respect each symbol's domain: an alias may resolve to a canonical Bead or its own locator-specific miss, never a Link, failed-creator unbound marker, opaque external URI, or another locator's missing identity. Reject a conflict only when these constraints or immutable fields already prove it. Existing creator correlations remain part of the same bounded slot comparison; do not introduce a generic expression/solver framework.
5. While the durable template contains any alias operands, they remain unresolved **by definition**, even if current live state or a satisfiable constraint suggests a value. Seed the unresolved flag from `aliasOperands.length>0`; after static checks the comparer can return only `conflict` or `pending`, **never equal**, for such a template. Current alias lookup cannot clear that flag. Direct callers must not receive the owner's pending receipt before ordinary final identity comparison becomes possible.
6. Once the owner reaches its turn, replace the template with existing final identity/witness JSON. The existing direct comparison loop rereads that state and uses ordinary witness-aware normalization. Maintain its one finite deadline, no DB transaction while waiting, epoch fence and retraction semantics. A foreign-pending sequence projection stays nonwaiting and wins even over a statically different candidate; transient-creator projection still precedes any dependent-key access.

**Required native counterexample:** work currently points toA. Reserve original `createLink(source:work,target:work)` underK. A direct singleton/one-op batch presentsK with `source:work,target:beads/B`. Partial comparison is `pending`: work is symbolic in both requests, yielding constraint work=B. Repoint work→B before the original turn; original executes B→B, and the candidate compares equal through the original turn witness. A current A/B comparison would falsely reject it. Repeat with source/target swapped to exercise cross-slot reuse. By contrast candidate fixedA/fixedB against original work/work is provably inconsistent and must conflict immediately. Add a swapped two-original-locator case to ensure symbols correlate by canonical locator, not slot number.

The external review F2 same-locator example correctly identifies the prohibition on handing out an unproven pending receipt, but its predicted later conflict is not generally correct: a later presentation of the same original locator inherits the **original turn's** recorded witness. The case must wait, then may be equal. A direct fixed canonicalA candidate instead conflicts if the original turn resolvesB.

This is an implementation of the existing turn-time and retained-witness laws, not a new wire semantic. Tests must prove that the comparer never resolves original template locators on either side, and never returns equal while alias holes remain.

### D. Keep final witnesses, receipts and RU untouched

`normalizeTransaction` (`transaction-batch.ts:391–467`) already implements original locator reuse, current lookup for new locators, distinct miss identity, canonical direct comparison and comparison-only witness behavior. Reuse it without replacing it with `normalizePreparedMemberIdentity`: the RU helper resolves committed member bindings, whereas TX batch creator indices remain symbolic.

Final terminal identity/witness encoding, receipt projection and successful tombstones remain unchanged. Keep private `format` and `identityPolicy` unchanged. The extension lives only in tagged pending `dependency` JSON; reject unknown/malformed tags at its private parser as an integrity fault. Preserve support for the existing creator-only pending shape as needed within one owner; do not reinterpret it as a different identity.

Compatibility evidence is required, not assumed. Current startup acquires exclusive ownership and checks exact configuration, then retracts **every** owner-bearing pending receipt and deletes creator facts before the first sweep (`transaction-store.ts:514–557`). Retention sweep selects only owner-null rows (`499–503`); `expireRow:421–492` skips pending rows and never interprets dependency. Receipt disclosure reads receipt JSON rather than the template (`830–843`). Thus a later owner—including old code—must retract pending alias templates before any comparison can parse them.

Required compatibility observations: (i) maintain() with a live unresolved alias reservation leaves it reserved and does not interpret/retract it; (ii) crash with that reservation and reopen with new code reports recoveredPending and leaves prefix state intact, without executing tail; (iii) actual frozen predecessor emitted owner opens a disposable image written/crashed by new code with the same configuration, retracts pending reservations, and can perform fresh work; (iv) old and new owners read/replay terminal success/miss records through already-supported direct singleton/one-operation-batch routes with the unchanged terminal codec; (v) concurrent second-owner opening remains refused. This is an actual old-reader process test, not a rewritten metadata string or assertion about hypothetical behavior.

If this proof fails or a persisted terminal-format change proves necessary, stop and return to root. **No format/identityPolicy bump is authorized by this plan.** No RU store, evaluator, public configuration, schema or migration change is planned.

## Proposed file ownership / dependencies

Production files (4):

- `packages/server/src/transaction-batch.ts`: preparation metadata and removal of refusal only after deferred machinery exists; reuse normalizer.
- `packages/server/src/transaction-dependencies.ts`: deferred alias representation and conservative exact partial comparison.
- `packages/server/src/transaction-store.ts`: atomic reservation, durable-row-consistent fenced turn normalization and additive pending-only compatibility proofs; no format/identityPolicy changes.
- `packages/server/src/transaction-http.ts`: remove only the now-obsolete Resource-alias development-gap branch.

Tests (bounded4): `transaction-sequence.test.ts`, `transaction-dependencies.test.ts`, `transaction-http.test.ts`, plus a focused new `transaction-sequence-alias.test.ts` if the corpus is clearer separately. Prefer existing emitted-runtime fixture/reopen helpers and stepper/scheduler, not new public test hooks. Existing comparison/recovery/runtime suites are regression targets; their source need not change unless a focused assertion belongs there.

**HTTP ownership/order:** C1 goes first. Root will supply its frozen base; alias preparation/dependency/store work may then proceed in an isolated checkout with exact root-assigned scope. Do not independently edit the C1 HTTP file or its tests while that author owns them. Removing the alias-gap branch and adding HTTP tests happens only after ordinary integration of the reviewed C1 successor, preserving its route/method/admission repairs. The alias lane owns no discovery/composition files.

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
7. Direct singleton and one-op batch races must include the complete native work/work vs work/B A→B counterexample, its cross-slot swap, and two-original-locator swap. Assert `pending` before turn, no pending receipt delivered, owner can progress, then equal receipt after witness capture. Separately fixedA/fixedB conflicts immediately; fixedA against eventualB waits then conflicts; immutable-field mismatch conflicts immediately. Alias-only compatible templates must never compare equal while unresolved. Unit resolver throws on any original-template locator, including candidate-side occurrences; genuinely new locators may resolve without persistence. Use existing finite comparison scheduler/deadline, not sleeps, and test timeout without key loss.
8. Same sequence duplicate remains transient for foreign pending keys, including mismatching immutable fields. Mixed transient `@creator` + alias member must not consult its key or alias, and releases only its own reservation. Preserve all existing creator-attempt reuse/expiry/fencing tests.
9. Add maintenance/old-reader/reopen/terminal-codec/exclusive-owner proofs specified in§D. Assert that an owned alias-only deferred claim cannot bypass normalization; preserve creator-only and mixed precedence. Real emitted crash after earlier alias repoint/prefix, before later alias member: recovery retracts abandoned pending tail, does not execute it, preserves prefix. Explicit resubmission resolves the later alias in its new turn and succeeds. Existing process fixture pattern at `transaction-sequence.test.ts:469ff` can be reused with disposable files and bounded child cleanup. Crash/fault after turn normalization must neither commit a second mutation nor retain an incorrect witness; exact-owner cleanup/retry is checked.

Counterfactuals must distinguish (a) removing only refusal, (b) snapshotting aliases at whole admission, (c) resolving a recorded locator on every retry, (d) persisting fresh retry witnesses, and (e) equating unresolved direct callers by spelling. These can be permanent positive/negative test sequences, not production mutants or synthetic responses.

## Validation and done-when

Build emitted code before HTTP/process tests, pinned Node24.16/pnpm11.20, maxWorkers2, external logs and bounded child/process-group cleanup. Run focused TX batch/sequence/dependency/comparison/submission/store/recovery/HTTP tests affected by the changes, then typecheck/build/boundaries/format/lint and relevant installed smoke. Reuse the existing source-private installed TX launcher for a fresh packed HTTP alias success/replay/reopen smoke if normal test entry cannot prove packed payload equivalence; no evidence profile relabeling.

Targeted RU regression: `member-identity`, `member-executor`, `sequence-executor` alias/retained-reference tests plus the existing installed RU alias/reference and sequence-retry positive scenarios. Since shared evaluator source is unchanged, a full8k or ever-growing conformance cohort is not the first validation step. Broaden only for an actual shared-file change/failure.

Machine-checkable completion:

- Existing valid alias sequence no longer returns `sequence-resource-alias`500; original malformed cases still fail before writes/key capture.
- Each above semantic discriminator has an executable test with actual afterstate; HTTP tests schema-check response envelopes/receipts and assert positive status before equality.
- All-key reservation remains one atomic step; alias capture occurs in the eligible turn; immutable final identity/witness and comparison-only retry proofs pass before and after reopen.
- Direct unresolved comparison uses the existing one finite budget/no DB lock; sequence projection remains nonwaiting. Original-template locators are symbolic on either side in every eligible slot; alias-bearing templates never compare equal; the native A→B and fixedA/fixedB discriminators pass.
- Existing batch, singleton alias, creator dependencies and owned-Link behavior pass regressions; no new public profile or configuration capability is advertised.
- Additive pending-only old-reader/recovery/maintenance proofs pass with unchanged format, identityPolicy and terminal witness codec; otherwise stop for root review. Clean bounded commit/report identifies the exact pending-form addition and precise remaining scope. No whole-contract coverage credit is inferred from these runtime tests.

## Cold-discovery interaction and limits

This is independent of cold discovery implementation. The current HTTP fixture knows private targets; closing this gap makes `sequence` functional for alias operands but does not prove a cold client can discover it. Cold discovery can point at the same existing sequence handler once its route table is published; it must not advertise a different weaker operation contract. Reuse immutable configured Scope/Host and actual owner. Do not merge stores or hide this new path behind an alternate mock launcher.

No product ruling is needed for turn-time alias resolution, retained witnesses, or separate commitment: these are explicit existing laws. Review is required for the private deferred-template/partial-comparison implementation and the explicit unchanged-format compatibility proof. No fallback format bump or asynchronous owner redesign is included. Authorization remains the currently supported all-visible TX policy; this slice does not claim mutable permission/erasure coverage. Selector timing, minimum-position controls, Event/Snapshot questions, alias enumeration and public activation are outside scope.
