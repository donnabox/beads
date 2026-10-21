# Independent native runtime audit — fresh private-temp prefix/full suite

PASS for the eight freshly executed stages through full-test. This is not acceptance of later smoke/evidence/package/packed stages or whole-profile qualification. No rerun or product/helper execution was performed by this reviewer.

## Identity and fresh prefix

Candidate `d7197e97a29bff8f075028419afad05388e2e655`, tree `630a94d0ca1545f898cb8a8b614d4b7d2dd6d454`, clean worktree. All 701 tracked source hashes remain exact. Fresh clean build inventory has 477 records, all rehashed against actual current files, and its started/exit/log bindings match. Fresh workspace tool document `b786d3cbd749a07cb938407f09cc62ac46d25f90e7d572c51c47f9ca4591cbd6` matches all 673 tool-file identities. Its post-dependency measurement is separate from original failed validation.

All eight actual stages have raw exit 0, accepted=true, unchanged source, no errors/forced cleanup/reported survivors or residual `bdp-` runtime directories. Actual log hashes match their exit records; every started/exit pair records the same fresh randomized /private/tmp root. No predecessor stage outcome was reused.

| Fresh stage | Supervisor seconds |
|---|---:|
| dependencies | 0.408 |
| format-check | 0.521 |
| lint | 0.640 |
| check-boundaries | 0.692 |
| typecheck | 1.029 |
| build | 1.047 |
| bind-build | 0.917 |
| full-test | 235.529 |

Dependency install reports lock-current pnpm11.20.0. Format reports one oversized read-update matrix warning (465 files checked). Lint reports 1030 warnings and 84 infos, with 1094 diagnostics not shown; this is not a zero-warning or independently established warning-baseline-equivalence claim. Dependency boundaries report 11 workspace packages. The build log executes clean then tsc build; it is not merely a rebinding of old dist bytes.

## Terminal full-suite result

Actual terminal JSON reports success=true, 9375 passed, zero failed and one skipped; the raw log reports **187 passed files**. JSON's 533 test suites include nested suites and are not 533 files. Raw exit is 0; supervisor elapsed235.529s is below the unchanged1800s bound, with no forced cleanup, retry or swallowed failure.

The formerly failing required case `bdpbd Read realization passes every checked-in plan through an isolated public bd CLI workspace (requires local pinned bd or BDP_BD_MATRIX_EXECUTABLE)` passed in46605.514ms. Actual started environment requires BDP_REQUIRE_BD_MATRIX=1 and the unchanged isolated pinned bd executable. Its fixture root is `/private/tmp/vickie-pr150-validation-full-test-9xw0gq54`. The separately skipped runnable Read walkthrough remains skipped; it must not be substituted for this required passing matrix.

All five affected callback full names occur exactly once as passed in terminal JSON:

| Callback | Duration ms |
|---|---:|
| installed Scope/Bead/Link Event SSE matches finite bytes and sparse Bead survives fifteen seconds | 15705.418 |
| installed Event cursors survive restart with Last-Event-ID precedence and source binding | 991.894 |
| serves byte-bounded deep mutations receipts Events and owned-Link Snapshots over packed bytes | 3885.274 |
| serves missing-first current Resources and bounds unproved minimum eligibility | 4519.282 |
| observes retained sequence member conflict and fresh deep execution over packed bytes | 432.274 |

These are full-suite **workspace emitted-child** invocations; later external consumer callbacks are separate. Restart explicitly reports development-local/no installed provenance. Actual entry hashes and declared candidate match; workspace payload is73455528487c841c80c4bc0243e49d8284b2441807b027fc03633d5bdfa5d11f, not an as-yet-unmeasured packed payload.

Raw observations independently parsed: fresh-depth line251 has13 attempted/exchanged/completed responses, minimum-current line318 has32, and sequence-depth line6055 has8. Each reports passed, zero failures/deadlines/cleanup timeout, closed child with exit0/no signal, claimEligible=false and absent owned directory; actual current entry hash matches. Restart line6172 has five requests/four finite responses plus SSE, two closed sessions (53141/53208) each exit0/no signal/no unknown stop, zero pending operations/failures and removed directory. Legacy SSE line5540 reports15120.355ms idle/15143.134ms stream, two comments and one live Event frame; its report lacks a standalone full cleanup ledger, so only its actual assertions/pass plus run-level cleanup are claimed.

## Specification and cumulative checks

Actual spec-layout12, profile-contract-plan59, profile-catalog-boundaries23, schema-bundle14 and transactional-cumulative42 cases all passed. Spec-layout retains legacy anchors, actual document-link targets, authoritative Write/navigation separation and citation boundaries through its existing checks.

The seventeen-partial-contract callback passed. Source at transactional-matrix.test.ts:627 asserts the actual scenario exchange sum234 before printing the terminal summary; line1253 reports241 manifest actions,24 sessions,25 launched/25 closed children, elapsed9431.689ms and claimEligible=false. The raw summary does not contain the complete234-exchange array. This audit establishes the passed actual count assertion plus summary, not an independent reconstruction of missing raw exchanges. The unrelated requirement inventory234 is a different unit; none of these counts is a completion percentage or public qualification.

## Cleanup, failure preservation and limits

A fresh process snapshot found all30 distinct inspected PIDs absent: sampled prefix/full-suite processes plus five child PIDs extracted from the selected actual reports. Four inspected owned store directories are physically absent. The recursive bounded `bdp-` directory scan under each stage's recorded temp root is clear. Parent temporary roots/caches remain deliberately preserved for audit. These are sampled/observed cleanup checks, not a claim about arbitrary detached processes.

Original validation remains immutable FAILED: raw1,9374passed/1failed/1skipped, full-test.log SHA256666318015c4032a98d7f08fd57c3b696dbae082ac8866210164c6e0a09c86f68 and exit SHA256113ba48738d3d0cd7328ca8e02c00a3b8d8c822ac698a3c681eb739c8437a48a reverified. The new success is a distinct reviewed private-temp successor, not relabeled historical evidence. Root retains control of all subsequent stages and external-provenance acceptance.

## Current terminal evidence hashes

- `run/full-test.json`: `5f056cc0b13e97b0d714c1368063f09aa80cb2b7d79e1107017d753b3720f0b0`
- `run/full-test.log`: `8a742a2bc902b5938cbdb0ad93349fe3557b3224851158e9e07552c62c38350d`
- `run/full-test.exit.json`: `894289211d4f34596ccb0ef2d0e2de369928614ceff6f580d0bf3d8fa1aa1869`
- `run/full-test.started.json`: `de9ad9a5f195201fba7240f6d8972a98df687261f029c3f2640f59d56d319202`
- `run/full-test.processes.jsonl`: `942f25c11bea665f5c691d3fa7e210b6c272cc6fa15cc00e49c454d7b58087b2`
- `run/build-artifacts.json`: `8c53e912f90406831be452eb79fcde95bb5ecdb2641298c6a1e43f98d3708c54`
