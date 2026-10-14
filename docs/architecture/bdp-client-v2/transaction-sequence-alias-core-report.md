# Transactional sequence Resource aliases — frozen core

Source: `0de38a3ec55d263e5e5c44ac116dcc87acc74f45`; tree `1f7f77ee8c568206b56dbae19a3da8a3e82fb77c`; ordinary parent `581bad0de6f5eb098b920147f6a7e09809c553ea`. Clean checkout `/Users/dbox/repos/wt/vickie/vickie-tx-sequence-alias`, branch `vickie-tx-sequence-alias`. Implements accepted v2 plan `575a2f104f3ece1412645099a7b6839ffa278c3a0f877f5d1b528d37314b492d`. No outward writes, store format/identityPolicy changes, HTTP edits, public admission changes or qualification claim.

## Changes

- `transaction-batch.ts`: immutable preparation records permitted Resource alias operands alongside creator dependencies and one consistent deferral fact. Preparation now accepts the existing alias grammar for update/delete Bead and Link endpoints. A source-private operand inventory shares that grammar with comparison; public exports are unchanged.
- `transaction-dependencies.ts`: additive, validated `transaction-deferred-alias-1` pending-only representation. Partial comparison preserves non-reference structure and pins. Original alias locators remain symbolic on either side and in any eligible slot; fixed/new-locator facts constrain correlated symbols without consulting live original aliases. Compatible templates always remain pending, never equal. Existing creator-only representation and terminal alias witness encoding remain unchanged.
- `transaction-store.ts`: sequence admission reserves without alias lookup. Eligible synchronous turns validate actual durable deferral state, capture live aliases with creator bindings and atomically replace pending identity/witnesses before execution. Nondeferred members retain their previous transaction/fault boundaries. Creator transient precedence, owner fencing and synchronous scheduling remain intact.
- New `transaction-sequence-alias.test.ts`: 19 focused observations cover turn-time target, native cross-slot correlation/repoint counterexample in both directions, pinned/allocated creator mixing, successful deletion, retained success/miss/restart, distinct new candidate alias facts, wrong-kind/put-to-alias member failures, unchanged malformed-template rejection, fixed-field conflicts, direct singleton/batch waiting, bounded deadline, foreign transient creator and actual emitted-process crashes.
- `transaction-sequence.test.ts`: removes only two now-supported alias cases from the obsolete rejection table, accurately renames it, preserves malformed/duplicate rejection and strengthens no-key-capture checks.

## Verification

Pinned Node `24.16.0`, pnpm `11.20.0`; frozen install passed. Final exact-commit regression (`frozen-regression.log/json/exit`) passes **487 tests / 11 files**, 5.20 seconds, maxWorkers=2. It includes new aliases, existing TX sequence/dependencies/store/recovery/submission/service/owned behavior and RU member identity/sequence/alias evaluator regressions. No full workspace suite is claimed.

Build passed (`build-verified.log`, exit 0). Typecheck, formatting, lint, dependency boundaries and unchanged historical Read evidence verification passed again on frozen source (`checks-frozen.json`). Both executable and development smokes passed against unchanged final production bytes (`checks-verified.json`); subsequent edits were only the additional deletion test and its corrected expectations.

### Actual predecessor compatibility

`old-reader-frozen.mjs`, `.log`, `.exit` and `old-reader-frozen-proof.json` record a real emitted writer crash and an actual predecessor owner opening its database. Current frozen writer source is clean. The predecessor is clean `61ed530ac6a460598e0e312f5d28ec0e6fa2a33d` in `/Users/dbox/repos/wt/vickie/vickie-schema-property-change`, whose server/protocol source is identical to parent `581bad0`. No rebuild or source write occurred there. All 240 server and 56 protocol emitted files match its prior frozen c57 packed installation; `predecessor-provenance.json` records that comparison. Old store module SHA256: `2f7d8d911cc6d01d70e13f41ff1b3abd7838d10625645e312403647ac2bcea63`.

The current writer retained a terminal success and a terminal alias miss, committed a prefix repoint, then exited 79 with an alias tail pending. The old owner opened successfully (exit 0), retracted exactly one pending row before interpreting it, reproduced both terminal receipts exactly, accepted the abandoned key fresh, and committed the expected target effect. The new owner then reopened and reproduced that old owner's receipt exactly. Stores are disposable; cleanup removes each directory. This is the accepted additive-pending compatibility witness, not a format migration or power-loss guarantee.

Core tests separately crash after a committed prefix and at the in-transaction dependency-resolution fault. The latter correctly rolls back its pending-row replacement: reopen retracts the original template and does not resume the abandoned tail. Exclusive second-owner refusal and explicit retry after restart are asserted. An active-owner maintenance call preserves pending reservation.

## Failed attempts retained, not relabeled

- `red.*`: four initial tests fail on the original alias-preparation refusal.
- `typecheck-initial.*`: preparation literal inference and test unknown-resource typing corrected. `typecheck-second.*`: imports were unused before the planned process tests were added.
- `core-second.*`: 51 pass / 2 fail because existing emitted-process tests were run before the first build. The missing emitted modules were resolved by building, not by relaxing tests.
- `core-third.*`: 55 pass / 1 fail: the new crash oracle incorrectly expected an uncommitted dependency-resolution row to survive process exit. Corrected to the actual atomic rollback guarantee; `core-fourth.*` passed 56.
- `regression.*`: 484 pass / 2 fail revealed an unnecessary new transaction on nondeferred turns, shifting injected cancellation failures. Production was corrected to inspect under the owner boundary and transact only deferred normalization; both unchanged storage-failure tests pass in `regression-corrected.*` and the final frozen run.
- `build-corrected.*`: TypeScript null narrowing was lost across a closure while making that correction. Parsing the already-validated non-null pending text before the atomic callback fixed compilation; no fallback/default identity was introduced.
- `alias-final.*` and `alias-verified.*`: the extra deletion test first expected the wrong result envelope, then attempted to repoint a dangling alias that existing evaluator law refuses. It now asserts the normative deleted Resource envelope, exact afterstate and retained replay after reopen; no evaluator change or relaxed result check. `alias-complete.*` passes all 19 tests.
- Preliminary dirty-source old-reader evidence is preserved separately as `old-reader-proof.json`; only `old-reader-frozen-proof.json` identifies this final clean commit.

## Boundaries / next handoff

This closes the assigned core implementation slice. HTTP source/tests were deliberately untouched while C1 owned them. The predecessor HTTP alias-refusal assertion is now stale and is **not claimed passing**. Parent must ordinary-merge frozen C1 `71c7aa6f9d81ea32965a1c5d5a15fd4b89bb092c`, then authorize narrow actual HTTP alias/retry/restart observations and replacement of that stale assertion. No installed profile activation, fresh packed full-profile claim, or whole-contract promotion follows from core tests.

All child tests and synchronous crash probes completed; no owned child processes or disposable stores remain. No user/shared database was used. Exact changed-file hashes are in `transaction-sequence-alias-core/source.json`; all logs referenced here are in that directory.
