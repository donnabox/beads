# Receipt-page helper review repair

Frozen clean **55299682e1ecfe7ee84c95975abdb94685607a07**, tree`24d1a731c6cd0f8f6da73f8e8c42d4cf66edafd3`, direct ordinary parent`43e0aff5c5a6142a5f9dc2c50283e437b917d258`. Checkout `/Users/dbox/repos/wt/vickie/vickie-receipt-page-helper-repair`, branch`codex/vickie-receipt-page-helper-repair`. Original43e and transaction evidence checkout untouched. Only `transactional-observation.ts` and its helper test change (+14/-4); runner and all other files are byte-identical to43e. No outward writes.

## Findings closed

Low1: the single-terminal-page fixture bound now applies only while establishing the first accepted page baseline. A later schema/identity-valid page with a changed non-nullnext compares against the retained baseline and raises `TransactionalReceiptFailure` (target failure), rather than misclassifying it as `TransactionalHarnessError`. Wire/schema/parent checks still precede baseline handling. The first nonterminal page remains a harness limitation and cannot initialize a baseline. HEAD still cannot initialize it.

The permanent helper regression pins both error classes, the retained-facts message and successful reread of the original page after refusal. Before the fix it failed exactly as intended: expectedTransactionalReceiptFailure, actualTransactionalHarnessError("receipt page exceeds single-page fixture"). Original red result is retained, not relabeled as validation success.

Low2: each existing capture-order test now explicitly rejects registered root binding→page and registered page binding→root. Both orders pass, alongside existing unknown-name/different-page/revocation/reopen checks. No new authority or runtime behavior is added for this finding.

## Validation

Pinned Node24.16.0/pnpm11.20.0, frozen install. Typecheck/build, touched-file format/lint, boundaries and git diff check exit0. Focused existing helper/runner/cumulative suites pass **334tests,3files,0failures**,1.01s Vitest duration (1.41s process supervision), including real local HTTP runner cases inherited from43e. These tests ran on final precommit bytes pinned by `source.json`; no later edits occurred. Full commands/exits/logs and red/green JSON reports are adjacent. Lint exit0 is not represented as warning-free.

No full packed aggregate, workspace suite, installed receipt evidence or profile claim was performed. This repair closes two helper review findings only; artifact authoring and installed qualification remain separate. No test-owned process remains; no caffeinate was started. Source is frozen clean. Root owns native closure, integration and publication.
