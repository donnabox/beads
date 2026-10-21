# Transactional atomic repair — independent follow-up and integration

Corrected author/root head reviewed: `2f7566cd18ca84be3a0cb87b9454eeb079fdaa62`, direct child of original author `d0020005ff01d6b444a916b83adf5fd4e376bb77`.

The prior required diagnostic repair is resolved. fence() now creates a fresh wrapper for the richer original+cleanup aggregate instead of ignoring it. It never mutates the prior wrapper referenced by the aggregate, so it does not introduce a cause cycle. enter()/atomic()/cleanupThenThrow guards prevent unrelated later accesses from refencing the owner. Added regression assertions cover subsequent access after singleton admission, execution and sequence-admission cleanup failures. The original external reproduction now traverses both original application cause and cleanup SQLite cause from both the initial exception and later inspect; output transaction-atomic-fixed-cause-repro.log. The recursive traversal terminates. Independent exact-corrected-head build and 93 transaction-store tests passed.

No remaining blocker found in this bounded review. Original before/after COMMIT, rollback failure, engine-ended transaction, retraction, recovery, receipt/effect/group consistency and ordinary retry assessment remains in transaction-atomic-independent-review.md. These are deterministic SQLite API fault injections, not real disk/power-loss qualification. Completed cancelSequence returning zero before the guard is a harmless already-completed no-op; active state access reaches fencing.

## Frozen candidate

HEAD: `486ad62d7c6ecb8e36dad4e596b1597fde425cdd`.
Checkout: `/Users/dbox/repos/wt/vickie/bdp-transaction-atomic-review`.
Branch: `codex/vickie-bdp-atomic-qualified-integration`.
Ordinary merge parents: published `8c0e043f6bfc5dae8b66044721a0aee170d66c97` and corrected `2f7566cd18ca84be3a0cb87b9454eeb079fdaa62`.
No conflicts/resolution edits. Exactly two store/test files differ from the published parent; original author and root follow-up history is preserved. Clean and frozen locally; no outward writes.

## Exact combined validation

Pinned Node24.16.0 / pnpm11.20.0. Build completed before emitted-runtime process tests.
- 375 tests in seven files passed (7.31 seconds): transaction-store, transaction-sequence, recovery-store, member-executor, sequence-executor, actual development matrix, development evidence inventory.
- Existing development report still has 27 plans / 280 actions / 907 passing assertions and claimEligible:false. No new Transactional HTTP/full-profile claim.
- Build, typecheck, dependency boundaries, format, lint, evidence verification, installed executable smoke and development smoke all passed. Lint retains baseline three warnings/one informational finding.
- Read evidence remains verified with constant6142cf73deae687bc3e9668cc014ea408fa02843, observed head4f0ed579, evidence commit72bdaad and89 target rows. No reseal, changed Read observation or schema/format/profile boundary change.
- No surviving owned test processes. Full repository suite not rerun locally; exact new-head CI still required on publication.

Logs and structured results are adjacent with transaction-atomic-fixed-review-* / transaction-atomic-integration-* prefixes. Combined test JSON and complete diagnostic report are retained; checks have explicit exit files.

## Published full-suite verification and subsequent finding

[CI38112180016](https://github.com/gastownhall/bdp/actions/runs/38112180016) at exact `486ad62d7c6ecb8e36dad4e596b1597fde425cdd` passed 115 files / 7,753 tests, one optional skip, seven preflight tests, installed smoke and actual `evidence.verified` with the unchanged Read seal/run head/89 rows.

Subsequent council packet07 found a separate diagnostic path: nextSequence aggregates an original turn failure with cancelSequence cleanup failure but later fenced access retains only the cancellation error. This bounded follow-up is being repaired; storage remains fenced, and this does not establish data corruption. The earlier direct/admission cleanup finding remains fixed.
