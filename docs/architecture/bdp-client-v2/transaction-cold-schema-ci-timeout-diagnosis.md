# PR119 schema CI timeout: bounded native diagnosis

**One unchanged retry is supported.** Run38606671302 attempt1 / job115999221459 reached the existing15-minute job deadline after its complete main test suite passed. The required preflight step was cancelled, and smoke/Read evidence gates were skipped. This run is cancelled, not green. No observed product assertion or build failure supports changing schema code on this evidence. Root owns the retry decision; this diagnosis performs no retry and waives no gate.

## Exact identities and final state

- GitHub run and job APIs both report completed/cancelled. API head is `c555fe01df5cd20454fb0d663431df816c1270aa`.
- Full raw checkout output lines418–421 binds actual synthetic execution commit `9d015ffed1d85320ffc334318923fb3f4821dd89`, merging c555 into `bff81e216595ca09dac36cd681b5fa4e6b1656f4`. Keep this identity distinct from the reviewed branch commit and local installed evidence.
- Green PR118 attempt2: run38602209334 / job115993441472, API head4aa2d0be950f157f3f0617278dd37c40f41ec035; raw checkout0cbee3a9db97fcaee0ab271e757dd73750e51514.
- Earlier green90cc reference: run38600031523 / job115976594627; raw checkout5b98fcceb5c10c0aad3ecc9bece8e8018d47e92e.
- Correct check-run annotations endpoint explicitly reports: “The job has exceeded the maximum execution time of 15m0s.” The other failure annotation is the resulting cancellation. Deprecation/migration notices are not test failures.

Source remained clean at c555 during diagnosis. The c555 delta changes nine files but leaves the workflow, package scripts, Vitest configuration and both long matrix test files byte-unchanged from4aa. Existing `.github/workflows/ci.yml:19` sets timeout-minutes15; all setup and downstream gates share that budget. No role/session mutation occurred; root's required remote session remains `2026-10-08T20:26:35Z-81e5` as the governing identity, not a new session claimed by this review.

## Completed and incomplete gates

All job steps through dependency installation, format, lint, boundaries, typecheck and build succeeded. The `pnpm test` step is explicitly success in the job API. Raw lines10994–10997 independently show **172 files passed;9,245 tests passed,1 skipped;695.33s duration**. There is no ordinary failed-test summary. The green4aa retry had170 files and9,233 passed/1 skipped; green90cc had170 files and9,232 passed/1 skipped.

All12 added schema tests have actual pass records: two service cases, eight adapter cases and two installed-file cases. Their reported per-test durations sum to1.862s (a sum of test durations, not their incremental wall-clock cost). The installed14-request case itself reports953ms. These counts reconcile the12-test increase over4aa; two other timestamp-bearing test names differ between logs without adding cases.

`pnpm e2e:ready:preflight` began17:41:30. Its nested Vitest printed all seven tests passed,1.50s duration, at17:41:33.098; cancellation follows at17:41:33.152. **The step conclusion remains cancelled**, so those pass lines do not establish a completed successful preflight gate. Smoke and Read-cohort evidence are explicitly skipped. Neither package smoke nor a complete CI pass can be inferred from the main-suite result. Cancellation cleanup is recorded, but it is not a substitute for ordinary completion of the interrupted gate.

## Throughput comparison

| Measurement | Green90cc | Green4aa retry | Cancelledc555 |
|---|---:|---:|---:|
|Job start UTC|16:40:12|17:14:40|17:26:32|
|Setup until test step|162s|150s|202s|
|Go source build|124s|116s|162s|
|Main test step wall time|445s|442s|696s|
|Vitest duration|444.61s|441.14s|695.33s|
|TX matrix completed cases|57|57|57|
|Sum of TX matrix case durations|421.782s|418.984s|663.608s|
|Main test result|9232 passed/1 skipped|9233 passed/1 skipped|9245 passed/1 skipped|
|Preflight/smoke/Read gate|all success|all success|cancelled/skipped/skipped|

The schema run consumed52s more before tests than the green4aa retry;46s of that difference was the unchanged Go build. Main-suite duration increased254.19s, while the same57 TX matrix cases increased244.624s in summed duration. The matrix remains the sequential long tail and completes its final committed-prefix-state control before the suite summary; this is not a stalled schema test or an unfinished matrix assertion.

Matched individually named passing cases whose green4aa duration was at least1s:266 pairs, median c555/4aa duration ratio1.579×. Against90cc,263 such pairs have median1.588×. These are duration comparisons, not reconstructed full-suite counts; duplicate display names and timestamp-dependent names are not used to override authoritative summaries. Examples:

|Unchanged test|Green90cc|Green4aa retry|c555|
|---|---:|---:|---:|
|RU cumulative positive|47.358s|49.676s|79.130s|
|Installed valid-set observation|1.373s|1.545s|2.235s|
|Deletion outgoing-order|8.325s|8.142s|12.104s|
|Deletion committed-prefix-state|8.474s|7.741s|12.225s|

Transform/setup/import timings also increase from5.21/1.96/26.72s in4aa to8.08/3.08/41.48s in c555. This widespread slowdown resembles the already diagnosed first4aa timeout, followed by its green unchanged retry. It supports a throughput/headroom explanation rather than a new schema assertion defect. These logs contain no CPU telemetry that identifies CPU contention, scheduling or any specific host mechanism. No such cause is proved, and source effects cannot be excluded merely by timing correlation.

## Bounded next action

Root may choose exactly one unchanged c555 retry, preserving this failure and requiring final successful main tests, preflight, smoke and Read evidence gate at the correctly bound retry execution identity. Do not alter job or request/cohort deadlines, omit controls, relabel the cancelled preflight as passed, or infer public Transactional qualification. If the unchanged retry also times out, first inspect its full raw log and stop automatic retrying; the recurring headroom problem should return to the held independently reviewed CI proposal with both normal and slow regimes. No product change is justified by this completed-suite cancellation alone.

## Preserved evidence and limits

Downloaded complete raw logs using `gh api repos/gastownhall/bdp/actions/jobs/ID/logs`, not `gh run view --log`. The c555 download has11,041 lines /3,171,265 bytes; the green4aa retry has11,065 lines /3,161,961 bytes. The earlier90cc full raw log is preserved unchanged. Relevant retained files:

- `transaction-cold-schema-ci-first-job.json`, `-first-run.json`, `-first-annotations.json` (distinct from root's initial `-first.json`).
- `transaction-cold-schema-ci-first-raw.log`, SHA256 `fe39a2e2fc6103918e3fa36208075f998a8c48e4bf7c136dda69636f67c2ca1f`.
- `transaction-cold-schema-ci-parent-retry-job.json` and `-parent-retry-raw.log`, raw SHA256 `12a52c618b1773f228c0f5beaa33e4f50f8ed7b21ffd5127438b298f4bb0fc64`.
- Existing `transaction-sequence-refusal-ci-original-success-raw.log`, SHA256 `369098baa8eacc1bfa767f7914cbe91b1ebef0219b15a9c764ea047245fcae8d`.

Only read-only GitHub APIs, local source/log reading and duration/hash arithmetic were used. No retry, workflow/config/source edit, test/build/install/service execution or outward mutation occurred. Root owns integration, source approval and public admission.
