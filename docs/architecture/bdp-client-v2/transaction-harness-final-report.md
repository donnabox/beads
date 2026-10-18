# Transactional harness final supplement — frozen

Source `9d379c5497df6f74f32ab565424e9174cf14f5c4`, direct parent `ee947ae919472d937625d196e2b5e5b67aa809d1`.
Checkout `/Users/dbox/repos/wt/vickie/vickie-transaction-harness-final`; branch `codex/vickie-transaction-harness-final`; origin GitHub; clean and frozen. Nine files,367 insertions/49 deletions. No outward writes. ee947 and its evidence remain unchanged.

## Result

Eight council residues are dispositioned in `dispositions.md`. Safe private error types distinguish actual invalid/different target receipt facts (fail) from evidence self-disagreement or missing validator/session infrastructure (harness-error), preserving fixed reasons without raw body disclosure or runner/helper import cycles. Source authority is canonical mutation-response only. Removed unused retained wire state, unreachable retainedSource authority and redundant late session attachment. Original response digest and retained JSON agreement remain; changed consistent serving headers are explicitly allowed.

Added the authorized optional cleanupTimeoutMs to the existing raw scenario target: omission preserves the legacy request-timeout/30s fallback. TX sets10s cleanup, while requests remain5s. Parent reopen/preparation10s remains an end-to-end work budget; emergency cleanup may finish after an aborted work deadline. No fresh per-phase10s promise. Existing signal/child drain remains tested.

The actual failed-child abort control found a real cause-retention path: runner cleanup converts child-close errors to row diagnostics before observe's final emergency close. The cohort now retains those errors separately and aggregates them with the original abort. The test uses a disposable actual installed child terminated with SIGKILL, then asserts original Error identity and child-close error among AggregateError causes and verifies children terminated. Strict drain-before-child-signal wording now applies only to ordinary cleanup; abort can initiate close and cancellation concurrently.

Constant consolidation is explicitly deferred. No serving/store/Snapshot/SSE code, protocol/catalog/schema changes or profile admission. No new public exports; the existing harness options type has the specifically authorized additive cleanup bound.

## Validation

Pinned Node24.16.0/pnpm11.20.0, maxWorkers2. Clean build, typecheck, boundaries, format, lint, both smokes,7 preflight tests, unchanged Read evidence verification and git diff check passed.

Final full run: **8,507 passed, one skipped,152 files**,113.29s (`full-final.json/log`). Focused substantive supplement run: **85 passed,five files** before two additional receipt unit controls, both included in the final full run. No timeout or interrupted run is used as proof.

Fresh packed protocol/server installed into a new directory with npm scripts disabled. Packed TX **22 passed**,17.72s; packed RU **112 passed,four files**. Both exit0; executor exits0 and source remains clean. Exact source, package/entry hashes, Node version, configuration and observations are in `packed/source.json`, `packed/transactional-observation.json` and `packed/summary.json`; file hashes in `hashes.json`.

Positive actual installed observation remains four partial rows,35 HTTP exchanges,10 dynamic exchanges,five physical sessions. Catalog250 required/five optional;251 other rows remain not-implemented harness errors. Whole-contract flags false; `claimEligible:false`. Runtime payload `09768af5b56fe2db91ff53bb4254e08b4983734cc547b1e549de34385008f464` is unchanged because this supplement only edits harness code. Actual fresh runtime configuration identity `6bbafc79034f6fd5cb4727d422303fa1ee952323a15e7486b0e7d6dab9dacf8b` is bound to generated credential associations for this run, not copied from earlier evidence.

All test processes finished. `cleanup.txt` confirms no owned Node child or runner remains. The deliberate SIGKILL is confined to its disposable negative-test child and observed as failure; no unrelated process was terminated.

## Diagnostic runs and limits

First focused run exposed the cause-loss path plus a new test fixture using the wrong body field; code and fixture repaired separately. One typecheck found readonly assignments in a new response-header unit control; corrected by constructing a new object. The final full run and typecheck use the corrected source. Earlier logs are retained, not counted as successful verification.

Synthetic raw HTTP/server and harness-self-inconsistency controls demonstrate classifications; they are not additional installed authority conformance rows. Four installed observations remain partial. General receipt paging, dependencies/aliases, catch-up, Snapshot/changefeed/owned qualification and other required contracts remain outside this slice. Historical Read proof is unchanged. Independent review/integration belongs to root; source stays frozen. Accepted SSE plan remains unimplemented pending root-assigned integrated S1b/H1b base and exclusive ownership.
