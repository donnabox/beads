# Alias core repair: native closure

## Findings

No new actionable findings. Accepted repairs1–6,9–10 are closed in the stated core scope. Finding7 remains the expressly accepted scope-threading deferral. Finding8's boundary alternative is honest and sufficient for the approved no-new-accessor scope; it is not direct execution with the inaccessible real deferred claim. Finding11 remains the known separate HTTP integration handoff. No public/profile readiness follows from this review.

Reviewed frozen `b3a427e23e9e2c5eb25148aa36c376ab0a3fbffb`, tree `7f850f51b6d3554efea85ddd3e14ebc031c06d81`, direct parent `0de38a3ec55d263e5e5c44ac116dcc87acc74f45`, in `/Users/dbox/repos/wt/vickie/vickie-tx-sequence-alias`. Full four-file delta: transaction-batch.ts, transaction-dependencies.ts, transaction-store.ts, transaction-sequence-alias.test.ts; 348 additions/33 deletions. Author checkout remained clean and immutable. Read the external core review, root-code dispositions, repair report, exact source ledger and predecessor evidence.

## Closure assessment

1. **Structural durable agreement retained.** `transaction-store.ts:1833–1838` compares validated templates with `isDeepStrictEqual`. Both sides are ordinary plain data; freeze attributes do not change their values. Object insertion order is irrelevant, but arrays, member order, strings and actual field values remain significant. This is a bounded robustness improvement, not acceptance of the prior review's unsupported claim that reordering a shared constructor alone changes one side. The real-owner interception test reverses the stored object's keys, verifies bytes differ, and successfully executes. The changed-attempt test still rejects an actual value difference. The integrity check was not removed or replaced with mere shape validation.

2. **Normalization UPDATE checks its own fence.** `transaction-store.ts:1844–1856` requires `changes === 1` before assigning owner.normalized or calling the post-resolution fault hook. The test deletes the actual owned row through the intercepted same SQLite connection immediately before the UPDATE. The transaction fails at the intended fence, the hook remains uncalled, Resource state is unchanged, and the key is reusable after cleanup. This establishes the missing stage-local guard without changing synchronous ownership or nondeferred transaction boundaries. Independent predecessor/restoration tests below discriminate it.

3. **Owner integrity and creator parsing tested.** New parser cases exercise extra top-level field, too many dependencies, invalid operation slot, field, kind and negative creator. Real owner tests mutate missing dependency, malformed JSON, different attempt, or foreign owner after a committed prefix. They require owner-integrity throws, unchanged Resource state, committed alias prefix, no tail execution, and key cleanup; a foreign row is retained until recovery. These are deliberately injected internal integrity states, not wire requests or demonstrated races between external writers. No WeakMap/debug accessor was added.

4. **Pending then conflict has a real state transition.** `transaction-sequence-alias.test.ts:696–740` first observes comparison-pending for canonical A. The real comparison wait callback repoints the alias to B and steps the owner, checking B's literal changed Resource. The pending caller then gets conflict; replay preserves the owner result, including after reopen. This covers the missing transition from symbolic comparison to final witness comparison, rather than merely comparing an already terminal receipt.

5. **Single private slot grammar.** `transactionAliasFields` in transaction-batch.ts is used by preparation, prior-witness validation, operand inventory, normalization/evaluation, and pending-template parsing. Accepting an untrusted operation string returns an empty allowed set for unsupported names. No public package reexport was added. Existing upstream Scope and prepared-object branding remain intact as directed.

6. **Reference encoding restriction matches actual admission output.** `transaction-dependencies.ts:92–110` accepts only exact current encoder output for a string or the existing private-unbound sentinel. Admission's reference map contains normalized URI strings for fixed slots and explicit holes for deferred aliases/creator bindings; it does not contain final alias-miss objects. Re-encoding equality validates the full outer tag, complete node shape, escaping and Unicode rules instead of trusting the extracted string or a prefix. Wrong tag, extra elements, semantic object, malformed JSON and invalid Unicode all fail the intended integrity diagnostic. The existing codec and terminal witness format remain unchanged; this is not a new semantic identity scheme.

7. **Finding8 alternative accurately bounded.** The test confirms admitSequence exposes only memberCount, captures a real pending receipt through an existing fixture connection, and proves copying that receipt cannot forge the WeakMap-backed claim. It also proves the reservation remains comparison-pending and only the legitimate turn executes after alias repointing. It does not claim to directly invoke `execute` on the hidden real deferred claim. The load-bearing execution check for missing normalized state/current dependency remains unchanged, and missing-dependency/zero-update tests independently pin the relevant turn integrity boundaries. Root explicitly allowed this alternative if the actual claim is inaccessible.

8. **Creator-only resolver and startup recovery.** Creator-only template construction now throws if an alias lookup unexpectedly reaches it; current legitimate creator-only and mixed cases pass. The added emitted-child malformed-pending test corrupts an actual crashed owner's row, proves it cannot be parsed, and reopens successfully with the pending reservation retracted before interpretation. This pins the startup ordering on which predecessor compatibility relies without vendoring an old implementation.

## Independent validation

Applied observable-test-runs using a disposable exact git archive at `transaction-sequence-alias-repair-native/source`. Copied emitted output, linked existing tooling dependencies, and used a review-local cache. No build, test-cache or source write occurred in the frozen author checkout.

Pinned Node24.16 direct command:

```text
node node_modules/vitest/vitest.mjs run packages/server/src/transaction-sequence-alias.test.ts packages/server/src/transaction-sequence.test.ts packages/server/src/transaction-dependencies.test.ts --maxWorkers=1 --reporter=verbose
```

Actual **exit0,75 tests/3 files,2.00s**; log `transaction-sequence-alias-repair-native/focused.log`. Includes35 alias tests and existing sequence/dependency regressions, with actual emitted crash/reopen cases.

For a bounded sensitivity check in the disposable copy only, replaced transaction-store.ts with exact0de38 source while retaining successor tests, then ran the two names matching `structurally equal reordered|normalization UPDATE`. Actual **exit1,2 failures/33 skipped**: reordered data throws `sequence deferral template differs`; the zero-row case reaches `transaction execution fenced` instead of the newly required stage-local diagnostic. Restored the file in a finally block and reran the same filter: **exit0,2 passed/33 skipped**. Logs `predecessor-guards.log`, `predecessor-guards.exit`, `restored-guards.log`. These mutant checks test source-level guard sensitivity; they are not predecessor installed compatibility proof.

Independently recomputed all four author source-ledger hashes and checked the restored disposable source bytes against the author checkout. All match. Test runners and synchronous child probes completed; no review-owned process remains. No broad full suite or installed HTTP cohort was duplicated.

## Exact predecessor evidence and limits

Inspected the repair's `old-reader-frozen.mjs` and `old-reader-frozen-proof.json`, separate from historical0de38 artifacts. They bind writerBase b3a427e… with writerDirty=false and actual predecessor61ed530ac6a460598e0e312f5d28ec0e6fa2a33d, predecessor store-module SHA256 `2f7d8d911cc6d01d70e13f41ff1b3abd7838d10625645e312403647ac2bcea63`. The recorded new writer crash is79; old reader exit0 retracts one pending row, exactly replays terminal success and miss, performs a fresh effect, and the new owner reopens/replays that effect. I inspected this author's rerun; I did not independently run the cross-build script. It is a bounded compatibility observation, not a universal crash-recovery guarantee or committed cross-version CI suite.

The author's clean frozen503-test/11-file run and static/smoke/Read-seal exits were inspected, not counted as independent executions. Its initial null-narrowing typecheck failure and earlier counterfactual failures are preserved and distinguished from final passing results. Format, identity policy and terminal codec remain unchanged. The dead HTTP alias-gap catch and obsolete no-prefix-effects test remain intentionally deferred until ordinary C1 integration; root owns that change and publication.
