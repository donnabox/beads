# Native code review: Transactional sequence alias core

## Findings

No actionable correctness or security findings in the bounded core delta. This is acceptance of the reviewed core implementation, not HTTP integration readiness or profile qualification. The known obsolete HTTP alias-gap expectation remains a separate, explicitly planned integration change.

## Exact source and scope

Reviewed frozen commit `0de38a3ec55d263e5e5c44ac116dcc87acc74f45`, tree `1f7f77ee8c568206b56dbae19a3da8a3e82fb77c`, against direct parent `581bad0de6f5eb098b920147f6a7e09809c553ea`. The author checkout `/Users/dbox/repos/wt/vickie/vickie-tx-sequence-alias` remained clean and unchanged. The full five-file delta consists of `packages/server/src/transaction-batch.ts`, `transaction-dependencies.ts`, `transaction-store.ts`, `transaction-sequence-alias.test.ts`, and `transaction-sequence.test.ts` (883 insertions, 41 deletions).

Read the accepted v2 plan (`575a2f104f3ece1412645099a7b6839ffa278c3a0f877f5d1b528d37314b492d`), dispositions, root acceptance, core report, relevant normative sections, and existing admission/normalization/execution/recovery seams. No external code-review findings were used to form this review.

## Contract and implementation assessment

1. **Turn-time normalization and unchanged boundaries.** The preparation inventory defers only permitted Bead-subject and create-Link endpoint alias slots. Existing wrong-kind Link-subject and alias-target restrictions remain member failures. `transaction-batch.ts` retains the existing normalizer and prior alias target/miss witnesses. `transaction-store.ts:1734–1850` uses the immutable prepared template at admission, resolves creator dependencies at the member turn, then normalizes aliases under the current owner's fence. It does not introduce alias reads at reservation time. Nondeferred members retain their existing boundary rather than acquiring an unnecessary normalization transaction. This matches `docs/specs/bdp.md:3062–3080` and `3202–3243`.

2. **Retry correlations across both endpoint slots.** `transaction-dependencies.ts` distinguishes original symbolic locators from new comparison-only locators, masks only URI-bearing reference identities, and retains operation shape, documents, pins, guards, and attribution in structural comparison. Correlation constraints preserve repeated original aliases and dependencies across slots; conflicting fixed targets fail immediately, while satisfiable unresolved aliases remain pending rather than becoming an equality result prematurely. Canonical Bead/miss domains exclude Link and external identities. Existing creator-only templates retain their historical shape. The finite-wait behavior preserves the existing contract at `docs/specs/bdp.md:3994–4025`; it does not return a pending owner's receipt before equality is established.

3. **Ownership, atomicity, and durable transition.** Store admission still enforces all-key reservation and existing foreign-owner precedence. The owned turn verifies the exact durable template against the immutable prepared template before updating identity/witnesses and clearing dependency state atomically. Execution still requires a normalized owned claim. Fault and cancellation cases preserve committed prefix effects while retracting only owned pending work. There is no asynchronous gap between alias normalization and execution in this synchronous owner path.

4. **Recovery and compatibility.** Pending alias templates are additive state rather than a store-format or identity-policy change. Startup's existing exclusive-owner recovery retracts owner-bearing pending receipts before maintenance. New tests exercise a real child crash with committed prefix and normalization fault, then recovery and retry. Separately inspected the author's actual predecessor-emitted-owner proof described below; it covers old-reader retraction and terminal receipt compatibility, not simulated version metadata.

5. **Test independence and remaining integration boundary.** The new suite covers turn-time repointing, original-versus-new locator comparison, correlated endpoints, pins, wrong-kind references, missing aliases, creator bindings, retained success/failure, cross-carrier replay, finite waits, ownership, restart, and malformed private templates. Removing the two formerly forbidden alias-preparation cases does not remove the remaining malformed-sequence/key-capture guards. No HTTP files changed. Installed HTTP proof, shared C1 integration, and replacement of the acknowledged stale HTTP expectation remain outside this core review; no public qualification is inferred.

## Independent validation

Used a disposable archive of the exact commit at `transaction-sequence-alias-native/source`, with copied emitted outputs and dependency links, and a review-local Vitest cache. No author source or build output was modified. Pinned Node was `/Users/dbox/.nvm/versions/node/v24.16.0/bin/node`.

Command, run from the disposable source:

```text
node node_modules/vitest/vitest.mjs run packages/server/src/transaction-sequence-alias.test.ts packages/server/src/transaction-sequence.test.ts packages/server/src/transaction-dependencies.test.ts --maxWorkers=1 --reporter=verbose
```

Actual exit **0**: **59 tests passed in 3 files**, 1.90 seconds. Log: `transaction-sequence-alias-native/focused.log`. These comprise the new alias cases plus existing sequence/dependency regressions. No full suite or author packaging run was duplicated.

An independent bounded probe, `transaction-sequence-alias-native/correlation-probe.mjs`, compared **720** original/candidate endpoint combinations drawn from two aliases, two fixed Beads, one Link, and one external URI. For each original alias, the probe enumerated representative eventual target/miss assignments and compared satisfiability against the existing final normalizer with retained witnesses. It also made any lookup of an original symbolic alias throw. Actual exit **0**, all 720 matched; log `correlation-probe.log`. This is a finite discriminating check of correlation/domain handling, not exhaustive proof for arbitrary strings, creator graphs, or durable execution.

The three copied emitted modules and corresponding source files were byte-compared with the frozen author checkout. No compiler was run in the author checkout, and no test or child process remains owned by this review.

## Inspected author evidence and limits

Read `transaction-sequence-alias-core/old-reader-frozen.mjs`, `old-reader-frozen-proof.json`, and `predecessor-provenance.json`; did not independently rerun that compatibility script. It records clean frozen writer `0de38a3…`, actual predecessor `61ed530ac6a460598e0e312f5d28ec0e6fa2a33d`, writer crash exit 79, old-reader exit 0, one recovered pending reservation, exact terminal success/miss receipts, and a fresh old-writer effect replayed by the new owner. The predecessor provenance compares actual emitted server/protocol bytes with prior frozen packed bytes. This supports the specific predecessor interoperability claim; it does not prove arbitrary power-loss or hot-journal behavior.

The author's broader 487-test/static validation is recorded in the core report but is not counted as this review's independent execution. Public profile gates and broader installed qualification remain unchanged.
