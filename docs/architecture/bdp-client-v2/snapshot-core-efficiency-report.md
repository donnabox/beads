# Snapshot core review repair

Frozen source: `390321d788b6d469d3e23b931be7f83ef32aa1b9`, direct successor of `a0a068c689f5588701c2b7a350bb677dc4f114ae`.
Checkout: `/Users/dbox/repos/wt/vickie/vickie-snapshot-core-efficiency-repair`.
Branch: `codex/vickie-snapshot-core-efficiency-repair`. Clean local tree; no push or outward action.

This is an author repair report, not an independent review. It addresses the verified findings in `transaction-snapshot-core-claude-review.txt`. Root retains independent council, integration, packaging and publication ownership.

## Dispositions

1. **Bounded ledger work:** count and compact encoded-size gates precede capture disposition decoding. The live query returns at most `liveHandles + 1` IDs and rejects impossible excess before decoding; valid capture decodes at most eight live dispositions. SQL still scans bounded historical rows for count/size/expiry/classification. It is not an O(1) database query or population-performance qualification. The 65,536 permanent-cap test now proves a repeated exhausted capture does no JSON parsing, and a 1,000-expired/one-live fixture proves only two disposition parses (live plus new manifest).
2. **View persistence:** format becomes `bdp-private-transaction-snapshot-10`; configuration stores the actual fixed `authorizationView`. Both differing-view and real format-9-shaped/missing-view fixtures fail installation matching before pending recovery, preserving ordinary file bytes and the claimed row. Restoring the valid configuration recovers the pending claim. No migration/reseed is added. SQLite hot-journal recovery remains a necessary precondition to readable metadata, outside the no-BDP-maintenance claim.
3. **Graph lookup work:** one bounded validated Resource map and outgoing-Link index replace each Bead's SQL scan of Links. A Bead without ownsOutgoing skips outgoing traversal. The map is limited to 4,096 Resources/32 MiB encoded record bodies, with additional decoded-object/inline-owned memory; there is no universal heap claim.
4. **Duplicate codecs:** the same map supplies endpoint and first-class/inline owned-Link equality validation. A permanent parse counter observes one authority decode and one independent retained first-page readback per Resource in a small owned topology, rather than repeated Link decodes during topology checks. Exact full inline equality is retained.
5. **Hot SQL:** finite literal statements are cached lazily per support/owner. Lazy creation is required because support is constructed before first provisioning creates tables. A regression verifies no duplicate Snapshot SQL preparation during capture and absence of the old per-Bead Link query. The surrounding owner's metadata helper is intentionally outside this cache.
6. **Ordering witness:** real create/capture/reopen covers percent-encoded U+FFFD, U+10000 and U+1F600 Bead/Link IDs with explicit expected order and matching owned Link arrays. Canonical HTTP URL identity requires encoded canonical spellings; the review's suggestion of raw supplementary versus BMP UTF-16 ordering is not a valid canonical wire-identity counterexample. No comparator semantics changed.
7. **Finite capacity:** unchanged, documented in the committed closure note. Expiry does not replenish the 65,536 permanent issuance cap. Eight live handles and encoded byte limits remain. No new rate policy, forgetting behavior, or longevity claim.
8. **Shared context:** startup sweep and reads call the same canonical-position context constructor. Startup malformed `01` position refuses and rolls back pending retraction; later corrected position recovers it. The existing atomic receipt test now checks receipt, position and epoch SQL observations are all transactional (three reads instead of two).

The external review's assertion that the old page-width arithmetic was correct is rejected by the previously recorded native counterexample. The a0a068c successor-token reserve and both 9→10 and 99→100 regressions are preserved unchanged and pass.

## Corruption timing and integrity

Bounded SQL checks reject malformed/nonboolean expiry flags and an `expired:true` row retaining captured records/pages. Such corruption fails startup before any committed pending recovery; tests preserve exact ordinary file bytes and the pending claim. Malformed JSON raises SQLite failure and owner open fails. Reclassification cannot release retained payload or reduce the permanent issued count.

A detail-free cold expired row's commitment is no longer SHA/JSON-validated on every unrelated capture. Direct manifest/page lookup still fully verifies the disposition commitment/context before returned recognition. This deliberate detection-timing limitation is pinned by a test: unrelated capture succeeds, lookup of the corrupt expired commitment fails integrity. It does not claim adversarial tamper resistance or instantaneous background integrity detection; an actor coherently rewriting durable state was never excluded by an unkeyed commitment. No persistent derived counter or unaudited cleanup path was added.

## Files

- `packages/server/src/transaction-snapshot.ts`: bounded ledger checks/query, lazy SQL cache, per-capture map/index.
- `packages/server/src/transaction-store.ts`: private format/config view and shared context.
- `packages/server/src/transaction-snapshot.test.ts`: 11 additional cases plus targeted existing bound/assertion updates; 40 total Snapshot tests.
- `packages/server/src/transaction-store.test.ts`: one guarded-read observation count/comment.
- `docs/design/transaction-snapshot-core-review-closure.md`: limitations, compatibility and finite-capacity disclosure.

No protocol/schema/public export/evidence/capability constant/readConfiguration tuple/packaging files changed. Root's separate Snapshot settings/hash and package inventory integration is still required. No HTTP Snapshot serving or Q1 interpretation is introduced.

## Actual validation

Pinned Node24.16.0/pnpm11.20.0; immutable source following commit.

- Frozen install: passed (`install.log`).
- Build: passed (`build-final.log`).
- Typecheck, boundaries and format: passed (`typecheck-final.log`, `boundaries.log`, `format.log`). Format reports the existing 1.9 MiB matrix size warning.
- Lint: exit0, 72 warnings/1 info including existing matrix size warning (`lint.log`); not a warning-free claim.
- Focused 7-file suite: **282 passed, 0 failed/skipped**, 7.23s (`focused-final.log`, structured `focused-final.json`). Snapshot40, store132, receipts13, Events21, changefeed20, owned17, HTTP39. Build completed before tests.
- Development smoke: passed (`smoke-development.log`), explicitly nonattesting.
- Read evidence verifier: passed 89 historical rows (`evidence-verify.log`); artifact, constant and original run head are untouched, not relabeled as Snapshot observations.
- Installed smoke: **expected dependency failure**, `smoke.log`: Snapshot emitted `.js/.js.map/.d.ts/.d.ts.map` missing from this base's allowlist. Root separately owns the reviewed packaging inventory repair (5208); this branch deliberately does not duplicate it. The smoke stopped before application execution and cleaned its temporary installation.
- `git diff --check`: passed before commit. Post-run process inspection found no remaining child of these test/smoke commands.

Earlier diagnostics are retained honestly: initial typecheck caught a test-only ResourceRecord union access; initial Snapshot test overcounted the surrounding metadata helper as cached Snapshot SQL; initial seven-file suite had 281 passes/one inherited expected-count failure caused by the extra guarded epoch read. These assertions were corrected, all final checks above rerun. No test deadlines or semantic assertions were weakened. No broad/full suite, fresh installed cohort, or public qualification run was performed for this bounded repair.
