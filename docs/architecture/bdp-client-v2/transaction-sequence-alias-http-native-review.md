# Alias HTTP successor: native code review

## Findings

No actionable production correctness or evidence-claim defect found in this bounded successor. Recommend accepting the HTTP integration subject to the independent council/root gates. This review does not reopen the already reviewed alias core or grant public Transactional admission, cold discovery, or whole-profile qualification.

Exact source: `ea80c77de7cd86244096a5978b65690a7ba637d8`, tree `c74f80cc11378dddbbddc7baaead44375be9d9cd`, compared with ordinary integration parent `70353341c7c51854c490b3c1746c00be13f1d068`. Reviewed all four files: transaction-http.ts, transaction-http.test.ts, new transactional-alias-installed.test.ts, and transaction-implementation-progress.md. The author checkout `/Users/dbox/repos/wt/vickie/vickie-tx-sequence-alias` remained frozen/clean; no author build or source was changed.

## Code and oracle assessment

1. **Production delta is appropriately narrow.** `packages/server/src/transaction-http.ts` removes only the obsolete five-line catch for the no-longer-thrown Resource-alias sequence development error. Existing TransactionDevelopmentLimitError handling, request reception, sequence admission, authentication/media behavior and execution remain intact. The previously accepted core supplies turn-time resolution; this successor removes stale transport logic and verifies its actual effects. Set-in-sequence/private-depth exclusions and their no-key-capture/no-prefix-effect tests remain separate.

2. **Former no-effects assertion is replaced by the actual intended contract.** `transaction-http.test.ts:427–484` requires HTTP200 with no obsolete gap marker, schema-valid positional sequence results, committed Bead a prefix, retained missing-alias failure, and independent Bead b tail. Actual GETs verify both Beads. After assigning the alias, replay must preserve the original full result; listener reopen preserves it again. A fresh-key deletion resolves the now-live alias and a real GET observes the deletion. This is stronger than merely changing the expected status/header and correctly preserves sequence partial-commit semantics.

3. **Mixed allocated and pinned operands have actual HTTP evidence.** The next direct test creates an allocated Bead, assigns an alias through a prior member, and creates a Link with a creator-bound source and opaque pinned alias target. It validates the response schema and canonical endpoint values, then independently GETs the Link. Repointing and listener reopen cannot alter the old sequence response or rerun the retained alias write; actual alias redirect checks establish that boundary.

4. **Installed success/retry oracles have independent expected values.** `transactional-alias-installed.test.ts:265–338` seeds fixed a/b state, observes work→a, then executes a prefix that repoints work→b before update/Link members. Expected B identity, numeric values, canonical endpoints and pin are authored independently, followed by real GET equality. After external repoint to A, sequence and singleton replay retain B's original result; an actual alias redirect proves the old prefix was not rerun. A new locator pointing to B compares equal to the same retained singleton receipt, then returns409 idempotency-conflict after repoint to A. A's literal unchanged state is observed.

5. **Retained miss is distinguished from current target state.** The second installed sequence deletes work, records resource-not-found on its next member, and commits an independent tail. Actual alias404 and tail GET prove those effects. Assigning work again does not change the complete replayed failure envelope. A fresh key then updates B to4, and GET confirms the new state; the old turn-update receipt still returns its original B=1 postimage. After reopening, both complete sequences and the old singleton receipt replay exactly while actual B remains4, A remains0, and the original Link/tail remain present. These are separate real response/poststate observations, not synthesized result-only tests.

6. **Schema and transport boundaries remain honest.** The test uses the existing normative sequence/receipt/Problem/current-record schemas, plus independent literals and full retained-result equality. The closed test-authored target set constrains per-request exact route/method binding; returned aliases do not broaden it. Canonical Host is asserted, actual write completion/wire-response presence is required, and credentials flow only through the resolver. Safe reports contain response digests and target/configuration bindings rather than tokens or request credentials. The launcher-checked source label is not described as target source attestation. The test has bounded fixture-specific coverage, not complete alias/metadata/authorization conformance.

7. **Lifecycle, capacity and count are consistent with scope.** The new cohort uses three Beads, one Link and two aliases, well below existing owner limits. One90s controller covers startup, exchanges and reopen, with existing5s exchange and bounded launcher close behavior;120s is the outer test budget, explicitly compared to the adjacent Set cohort rather than claimed universal. The finally block attempts closure and accumulates observed operational/close failures, clears the timer, removes its disposable store and rejects late-aborted success. On success it requires exactly two zero-exit sessions and36 completed exchanges. The request guard caps the internal ledger at60. Source count is28 requests before reopen and8 after. Reports deliberately call it completedResponses; they do not claim a full failed-attempt ledger. This review did not fault-inject every filesystem/setup-cleanup error, and no universal teardown guarantee is claimed.

8. **Progress text is bounded.** The new section accurately describes committed prefix/independent tail, existing exclusions, retained result/miss/restart observations and36/2 fixture scope. It explicitly excludes public admission, discovery, complete qualification, arbitrary concurrent histories, Snapshot consistency and crash durability. Historical earlier gap descriptions are labelled as historical rather than rewritten as current.

## Independent validation

Applied observable-test-runs. Used an exact disposable git archive at `transaction-sequence-alias-http-native/source`, copied emitted modules and existing tooling dependency links, with a review-local cache. No frozen-source build writes.

Pinned Node24.16 command:

```text
node node_modules/vitest/vitest.mjs run packages/server/src/transaction-http.test.ts packages/server/src/transaction-sequence-alias.test.ts --maxWorkers=1 --reporter=verbose
```

Actual **exit0,94 tests/2 files,3.01s**. Log `transaction-sequence-alias-http-native/focused.log`. This includes the new actual HTTP tests, inherited C1/receiver tests and alias core regressions; no broad workspace suite or installed packaging run was independently duplicated. Runners and synchronous child probes completed with no review-owned process remaining.

## Fresh author packed evidence available during review

The exact-source `transaction-sequence-alias-http/packed-summary.json`, `source.json`, actual `packed-tx-alias.log` and pack artifacts became available during review. This section describes inspected final evidence, not preparation runs relabelled as frozen proof:

- Source ea80c77… / tree c74f80… / parent703533…, clean true.
- Genuine packed alias observation exit0;36 completed HTTP exchanges; two child sessions36391/36392, each exit0; outcome passed; workspaceDirty false; claimEligible false. Supervisor duration1.129s. Actual installed payload digest `495b3d9c1e6d6a078b91b60f6c41cfb4ab3f8928595467ec895d171d668d53a2` and configuration identity `f3883ac9b60b3aa7f64014bc9b483401c9e13ac2c69ecba6780e8d2611bd441b` are recorded alongside captured contracts/policy/authentication/fixture identities. Reopen compares the complete binding.
- Independently recomputed tarball digests match: protocol `2943f90cd4d0c5ac973dc84c75c80b6b0d944e688b0210990438a482a0a61a8c`; server `8fae8608a298a2f89822fcbcdbe2c214941fb16c31a8375cc136b1e87cd6f2ea`.
- Frozen regression/build/typecheck/boundaries/format/Read-seal and two packed RU positive tests are recorded exit0. These are inspected author outcomes, not added to the independent94-test count. Summary reports zero remaining installed children.

The author preparation logs include failed attempts and are separate from the final exact-source summary. This review does not count their early assertions or corrected carrier/schema preparations as clean frozen evidence. No source edits, branch changes, outward publication or profile flag changes were performed.
