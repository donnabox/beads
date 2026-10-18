# Council synthesis — awaiting remaining reviews

Core: `72bdaad02b6fb2b803d4af7a8c4f90ada2f0d2d5` against `a9bcf9ed1d869aaba2cb49fe9cca10e3b8d52a94`.
Alias delta: `59915ea916a210f5ce36611eb6d799f7c3124056` against the core.

See `review-scope.md`, raw outputs and process status JSON for exact source and completion. This is not a readiness declaration.

## High-confidence findings

1. **Optional History discovery clarity — Claude and Gemini, confirmed by native review.** The profile table omits `historicalResolution`, and the old sentence explicitly admitting it in all three profiles was removed. Current cumulative inheritance and executable schema still admit the capability. Restore the optional row and distinguish Scope-history fields without inventing requirements. Claude's High severity is downgraded: there is no actual schema admission loss, and the table omission itself predates this branch.

## Single-provider findings independently checked

2. **Medium: profile ownership guard weakness — Claude.** The test maps shared/end-matter sections to owner zero, allowing a Read scenario to cite Transactional conformance rows. Its explicit-anchor and heading-slug handling also diverges from Markdown. Tighten ownership and allowed shared anchors while preserving valid current citations. This establishes a regression-guard weakness, not a currently misclassified catalog case.
3. **Medium: missing intermediate alias state witnesses — Claude.** Some alias mutations assert only HTTP status before retained-locator retries. Add direct alias readbacks to establish those state preconditions, preserving fresh-locator probes and all existing assertions. No runtime defect is established; Claude's assertion that a universal no-op implementation passes everything is overstated.
4. **Low: allocation observation precision — Claude.** Assert that the sampled allocated ID differs from the reserved live alias segment. Rename the URL serialization check honestly, or actually validate the safe-segment grammar. Deterministic collision skipping remains unqualified.
5. **Low: bounded development HTTP routing — Gemini.** An unknown POST under `/development/` returns 405 with `Allow: GET, HEAD` before the intended 404 branch. A focused routing test/fix is appropriate; this is not a public-profile conformance defect.

## Baseline decisions, not automatic changes

- Successful no-op key retention conflicts with the paragraph's claim that tombstone storage grows only with effects. Keep retention behavior until a product decision: forgetting a formerly no-op patch can let its retry modify subsequently changed state.
- Alias comparison deliberately performs an unfiltered lookup and exposes a bounded equality bit. Record the privacy tradeoff explicitly if desired; changing it would change current semantics.
- The schema admits `changeContext` on `deleteLink` independently of Type. Inert input still binds semantic mutation identity under the explicit rule. Editorial clarification is possible; no new runtime bug was proved.

## Rejected or incomplete claims

- Pinned update/delete Resource subjects are rejected by schema before admission; only endpoint Reference positions admit pin objects.
- Captured alias misses normalize to `undefined`; names on non-creator sequence members are rejected by carrier syntax.
- Fresh Read evidence exists at this core. Documents describing an earlier editorial pass are not the current qualification record.
- A missing expected development catalog should fail its citation audit, rather than be skipped.
- A malformed allocated URL becomes a recorded scenario execution failure; it does not cause an uncaught process crash.
- Pin provenance and endpoint identity differ from semantic request identity. The explicit byte-comparison rule governs idempotency.
- The diagnostic contract does not require distinct wording for every invalid target category.

## Suggested updates and disagreements

The parent assigned alias witnesses, allocation naming, and profile-guard repairs to separate branches. Review their exact deltas before extending council coverage. Treat the low HTTP routing issue separately. Gemini accepted the bounded alias diagnostics as written; Claude identified narrower evidence gaps which native review considers useful strengthening. No provider consensus establishes a runtime correctness failure. A transaction-turn concern is resolved by the inherited rule that foreign key state is classified when the member is reached; see `sequence-turn-admission-disposition.md`. A low follow-up remains to preserve the original exception if execution cleanup also fails; this is diagnostic loss, not a demonstrated state violation.

No public write claim enablement, Type lifecycle changes, wire/schema changes, historical evidence relabeling, or unsolicited helper removal is recommended.

## Coverage limits

This is a selected-packet review, not coverage of all 59 changed files. The current artifact summary was checked, but individual exchanges were omitted from manual review. No tests were run during this council; parent validation is separate. No power-loss, full public Transactional, authorization matrix, History/erasure, replication, or Type lifecycle qualification is claimed.

Later pending/interleaving integration `e538e23` is outside these packets. The timeout-only commit `21654be` received a separate native-only review. Gemini's initial transaction packet omitted the store tail and sequence file. The replacement packet completed for all three reviewers without truncation; the first empty Gemini attempt remains excluded from completed coverage.
