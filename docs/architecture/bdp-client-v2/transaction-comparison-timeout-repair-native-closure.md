# Comparison timeout repair: independent native source and evidence closure

**Clean bounded closure: no actionable findings or remaining repair blocker.** Frozen repair `70ac666002708202d2f93195eed90dfe8541300d` closes the accepted coverage and service-test failures without production changes. The evidence supports four distinct positive cases and three distinct control cases, with the two authorized service repeats accurately retained. Exact-head CI and root publication/merge acceptance remain separate gates.

## Frozen source and reviewed evidence

Reviewed the complete ordinary `30035b88380b7cbb3abd5359787db236b70345c7..70ac666002708202d2f93195eed90dfe8541300d` diff in `/Users/dbox/repos/wt/vickie/vickie-comparison-timeout-review-repair`. HEAD tree is `1353e8f4fb0b1149b3f4ff158335804bf3f1cecc`, parent is exactly 30035, and the checkout is clean. Four authorized test/doc paths change, 159 additions/21 deletions. No production or configuration change is hidden in the repair.

Read final report/source/summary/cleanup, preservation manifests, accepted preflight/native closure and root release, CI amendment, lint correction, execution scripts, selected test results, named failure records and command exits. Final artifacts under `transaction-comparison-timeout-review-repair/` have independently verified hashes:

- report.md: `a2e5721747b428b01bdd9d72009ead52316a67b8385127bb5357b350b5426cf3`.
- source.json: `68524854724fb5b0da026be9c5dec32ea0e722698b621d27ffc1a8450c73e58b`.
- summary.json: `cf1cfd25bb95b97983d08f070236f13a287a812ad843d32fdc13d993ac756cf9`.
- cleanup.json: `f9187884b3fcb0fbd8335fd461d2a1d5b194fc6cd8a6158464731ee3e26dff2d`.
- package-preservation.json: `b8a1abff6a0534214a165e3124089557abad2d217957d92163569399672552de`.
- source-preservation-final.json: `351c4604c218bccb73eaa17e69e5c5684c84d85b077be66781a895fd192c8132`.

All four source hashes/lengths match source.json. The final committed diff matches final-source.diff SHA `f122290a6b1e0ae519fdc1f8ae6527fec1b3dcb886d6400e0852fa7ed4db7716`. Independently byte-compared all 659 other tracked files against 30035 and the final preservation manifest; every file matches. Store/submission test and progress bytes remain identical to accepted preflight. The only later source adjustment is the authorized service-test line described below.

## Coverage and cleanup closure

F1 is closed by actual owner behavior, not a fabricated HTTP result. transaction-store.test.ts:3314–3355 extends the fenced case through explicit close, then adds the real existing-fixture maintenance callback failure with one 25 ms fake-time advance. It verifies timer cancellation, the retained failure message and original cause after maintenance closes the owner. The normal-closed guard in production was already correct; these tests cover both fault terms. No hidden production defect was found. The separately acknowledged earlier comparisonEpoch maintenance-failure shortcut is unchanged.

F2 is closed at the accepted narrow submission seam. transaction-submission.test.ts:1105–1143 uses a private full-owner double that returns stopped only at the reached final observation. Strict contextless shape, one wait, one observed admission and one callback with actual epoch are asserted. Both existing reservations are cancelled in finally. The literal discriminant and method typing are valid, and actual final typecheck passes. This is defensive composition coverage, not a live epoch-rotation claim.

The service CI amendment is correct. Production service.submit deliberately uses observed submission; the original30035 CI expectation of legacy refused was stale. transaction-service.test.ts:439–525 now pins the complete observed comparison-refusal/deadline tuple using trusted initial epoch/view and independently expected pos-0. Original finite-budget waits, empty pre-turn graph, two scheduled member executions, retained retry at pos-2 and exactly two groups remain. Immediately capturing the settled sequence outcome prevents abandoned rejected delivery. Finally releases remaining callbacks, drains and closes while retaining primary and cleanup errors.

The initial lint failure is preserved and truthfully reported. At line 504, the authorized change from immediately caught `throw outcome.error` to `failures.push(outcome.error)` has exactly the same accumulation behavior and avoids noUnsafeFinally. Independently compared preserved pre-lint source with final bytes: that is the sole change. No assertion, deadline or production path changes. The outer AggregateError still preserves the intended inner assertion, as shown by the actual final control failureMessages.

F3 is paragraph reflow only; historical 18-case/five-control results remain attributed to 30035. F4 remains informational. The failed full CI and its cleanup timeout are not relabelled as successful or erased by local checks.

## Executed checks and sensitivity

Independently reconciled selected JSON results and summary hashes, all 24 command exit records, executed mutant source hashes and adjudication artifacts. Store positives are 2 passed/138 filtered; submission is 1 passed/36 filtered; service is 1 passed/64 filtered both before and after the authorized line correction. Unchanged store/submission evidence remains attributable to file bytes now committed; it is not falsely described as a new post-correction whole-source run.

The closed-guard derivative built successfully and its two independently selected cases each exit 1 at the intended assertion: `closed fenced deadline preserves failure` and `closed maintenance deadline preserves failure`. The maintenance failure is the named Error assertion, not a masked later property error. The old-shape service derivative, before and after correction, builds and exits 1 specifically at `observed service comparison deadline shape`; actual reporter output retains that inner assertion. These expected nonzero outcomes are control sensitivity evidence, not ordinary passing tests.

Directly parsed all four service raw logs. Each contains exactly one cleanup record. Both positives report drained=true, closed=true, cleanupTurns=0, pendingCallbacks=0; both controls report the same fields with cleanupTurns=2. No hook/test timeout substitutes for this evidence. The final derivative source changes only the expected shape; corrected cleanup remains intact.

Actual workload is nine case executions: five positive executions covering four cases, four expected-failure executions covering three cases. Seven original executions plus two root-authorized affected repeats account for nine temporary stores and four in-process service instances. There are zero HTTP requests, listeners or installed children. MaxWorkers1, 120-second command supervision and unchanged inner deadlines match the release. No additional suite or silent retry appears in recorded execution.

Final typecheck, build, format, lint, dependency-boundary check and historical Read evidence verification exit 0. The final lint log retains 666 warnings and 43 infos, and format retains its existing file-size warning; no warning-free claim is made. Original failed lint and diagnostic checks remain visible. Full workspace and repaired-head CI success are not claimed here.

## Preservation, cleanup and limits

Independently reopened the historical30035 protocol/server archives and compared their 313 regular-file members with current rebuilt output. Protocol 62 files and server 249 files are byte-equal. The remaining two package manifests are decoded-equal: protocol directly, server after the existing workspace dependency version substitution. Archive hashes match `4683cb6c6333c332a4bcb7efc3e28f32d1d8d06a7264913f65c31d3a826d4227` and `b9ebfae7b6fd2f5602688fadf0110da3749d5af012fdf4599867c6af2f84bf35`. This establishes file/member preservation, not new tar identity or installed execution. Historical 27 requests/two children remain bound to 30035 and its original payload.

Both derivative source files are restored to the corresponding final source bytes. Independently queried all 24 recorded command PIDs and found them absent. Fixture teardown completion and source establish disposable-directory cleanup behavior; paths were not retained for a second independent directory-absence check, and the report accurately states that limit. Service cleanup is supported by actual records, not merely process exit fields. Fake timers restore in finally and the fenced SQLite spy restores before close.

No extra execution is needed to close this repair review. Root retains publication, exact-head CI and merge gates. Default installed timeout sensitivity, minimum/Event decisions, the earlier maintenance-failure shortcut, live epoch rotation and public profile qualification remain unchanged; claimEligible is false. Reviewer performed no tests/build/install/network/service execution or repository writes; only this report was written.
