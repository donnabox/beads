# Finite Event council closure successor

Frozen clean source `dd3c2c081208ef0da262abbc9d0158ce04c46587`, direct child of `d907a1e1576c5795c17edb5f15528f47edfd7d6b`. Checkout `/Users/dbox/repos/wt/vickie/vickie-transaction-event-closure`; branch `codex/vickie-transaction-event-closure`, GitHub origin. Three files only: `transaction-events.test.ts`, one alias reuse in `transaction-http.ts`, and new `docs/design/transaction-event-review-closure.md`. Shared progress document untouched. No outward writes.

## External finding disposition

1. Closed: real HTTP seed now contains escaped text and non-ASCII data. The corresponding HEAD length equals GET UTF-8 byte length and explicitly differs from JavaScript string length. A code-unit-count regression now fails.
2. Bounded disclosure: documented the exact accounting algebra separately from observations. Existing tests observe linear serialization calls, encoded size below cap and feasibility. No accumulator is exposed and no duplicate implementation is added to pretend a one-byte boundary witness. That precise budget-edge observation remains absent; no correctness defect was identified by the review.
3. Closed bounded gaps: non-string retained body fault refuses at the owner seam, then restored reads equal the original pages. A real middle group is removed from a disposable SQLite store; resumption from a matched earlier cursor rejects the positional hole, and restoring that exact saved row restores the pages. The broader endpoint-witness/corruption space remains explicitly unqualified.
4. Closed: GET/HEAD HTTP cases cover malformed query-bearing and query-free aliases, receipt namespace root, known receipt query and receipt-page query, with both a future minimum and an opaque unsupported minimum. Uniform 404 is asserted before minimum processing. Dedicated prose records the earlier private-gap-to-404 behavior delta; well-formed query-free aliases retain existing behavior, covered by the HTTP regression suite.
5. Closed: reuse the already computed `alias` classification in the later minimum route check. No parsing or protocol semantics change.

## Verification

Pinned Node 24.16.0 / pnpm 11.20.0. Frozen install and clean build pass. Initial full Event file: 21/21 pass. Final regressions after adding the opaque-minimum cases:

`pnpm exec vitest run packages/server/src/transaction-events.test.ts packages/server/src/transaction-http.test.ts packages/server/src/transaction-changefeed.test.ts --maxWorkers=2 --reporter=verbose`

**80 passed, 3 files, 2.18 seconds, exit 0.** Typecheck, package boundaries, touched-TypeScript formatting, repository lint, existing Read evidence verification and diff check pass. Existing lint warnings remain; no claim of warning-free output. Logs are in this directory. No full suite, fresh packed qualification or new Read seal was requested or generated.

Commit is signed off; no active commit hook was installed in this isolated clone, and no hook was bypassed. Source is frozen for root review/integration. No SSE, public admission, metadata-only incident ruling, storage format or capacity changes.
