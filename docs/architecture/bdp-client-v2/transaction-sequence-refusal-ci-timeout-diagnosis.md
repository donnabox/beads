# PR118 first repair CI cancellation: bounded diagnosis

**Recommendation: one unchanged retry is justified.** The authoritative raw job log shows continuing passing tests through 17:10:34, nine seconds before the job timeout. The apparent six-minute silence in the saved `gh run view` output is missing log content, not evidence of a hung installed-set test. Broadly slower execution plus longer setup exhausted the existing job budget. Do not change product/test deadlines or the workflow timeout on this evidence alone. A repeat should be diagnosed from full raw logs before another retry or a separately reviewed CI-only amendment.

## Source and run identity

- Failed run 38602209334, attempt 1, job 115984000843; API head `4aa2d0be950f157f3f0617278dd37c40f41ec035`.
- Actual checkout in raw log: synthetic merge `0cbee3a9db97fcaee0ab271e757dd73750e51514`, merging 4aa into `d50f554cf6418ba036072ebf6795cb272cd6cbc1`. Keep this CI execution identity distinct from the branch head and local test evidence.
- Successful parent run 38600031523 / job 115976594627: API head 90cc834, actual checkout synthetic merge5b98fcc of 90cc into the same main. It completed all tests and downstream gates.
- Repair is the verified two-file service-test/progress delta; production and 658 other tracked files, including the matrix and installed scenarios, are unchanged. The four modified service cases completed in the failed CI at 16:59:41 (28,9,5,12 ms); they were not the long tail.

## The tail gap was retrieval loss

The originally saved failed `transaction-sequence-refusal-ci-first.log` has 6,918 lines and ends test output at 17:04:32.5119533, then jumps to cleanup at 17:10:43. The read-only job-log API download has **10,997 lines /3,151,639 bytes**, including **4,033 passing-test records after that apparent cutoff**. The raw log contains the installed-set completion and subsequent suites/controls. The final pass is `discriminates operation-time deletion evidence against outgoing-order`,17:10:34.1427573,12,097 ms. The cancellation error is17:10:43.6594098. There are no observed ordinary failing-test records or completed full-suite summary in this failed run.

The annotation explicitly says the job exceeded 15m0 s. Its cancellation annotation line 9646 is not an absolute line in the whole-job merged download: cancellation is raw whole-job line 10974, following the test-step group at1326. The approximately9,646-line test-step span is consistent with the much larger raw test output, not with the truncated gh-view tail. Exact line numbering depends on the step header included by the viewer; it is not evidence of absent execution.

This settles the alleged six-minute hang. It does **not** prove the last incomplete control would have passed or that no intermittent lifecycle problem can exist.

## Measured normal versus slow run

| Measurement | Parent90cc success | Repair4aa cancellation |
|---|---:|---:|
| Job start |16:40:12|16:55:30|
| Test step start |16:42:54|16:59:08|
| Setup to tests |162 s|218 s|
| Go source build |124 s|175 s|
| Test step |445 s, complete|695 s until cancellation|
| Full Vitest duration |444.61 s|no completion summary|
| Full suite result |9,232 passed /1 skipped,170 files|incomplete; no full pass claim|
| Sequential TX matrix completed cases |57|56|
| Sum of those matrix case durations |421.782 s|651.397 s, still missing last case|
| Preflight / smoke / Read gate |all successful, about 8 s after tests|all skipped|

Common individually named test cases with parent duration at least1 s: 262 matched records, median repair/parent ratio **1.589×**. These are comparable verbose duration records, not a reconstructed total-test count. Examples:

- RU cumulative positive:47.358 s →79.625 s.
- TX cumulative positive:12.037 s →19.225 s.
- Retained-replay-body control:8.303 s →12.138 s.
- Deletion incoming-order:8.539 s →12.142 s.
- Deletion outgoing-order:8.325 s →12.097 s.
- Installed valid-set observation:1.373 s →2.271 s; it completed rather than hanging at the gh-view cutoff.

The only parent matrix completion absent from the failed raw log is the final `committed-prefix-state` control (8.474 s in the successful parent). In the slow run the previous two controls took about 12.1 s each and cancellation arrived9.5 s after the preceding pass. This supports, but does not prove, ordinary in-progress final-control work cut off by the job budget. There is no reason to mark that control passed. GitHub cleanup terminated residual job processes after cancellation, so this run does not establish normal final child cleanup.

The failed setup consumed 56 s more than the parent; 51 s of that increase was the unchanged Go source build. Together with widespread slower tests, this is a throughput/headroom episode rather than evidence that one added service test caused minutes of extra work. CPU/host scheduling/resource contention is plausible, but no runner CPU telemetry identifies the mechanism. The log evidence is sufficient for the bounded retry decision without claiming a specific infrastructure cause.

## Source/configuration checks and next action

Current `.github/workflows/ci.yml:19` still gives the single verify job 15 minutes; setup/build/full test precede required preflight, smoke and Read evidence. Vitest keeps its normal file discovery and the TX matrix controls remain sequential within the file, with each observe using existing bounded harness behavior. The repair does not change that file, runtime, package scripts or any deadlines. The last unfinished deletion control remains present in source; no scenario was silently excluded.

Root may request **one retry of unchanged 4aa**, preserving this failure and synthetic execution identity. Require the retry's full test completion plus all downstream gates and exact source binding; no promotion follows merely from apparent progress. If it times out again, retrieve the complete raw job log first and distinguish steady throughput from a genuinely uncompleted operation. Then revise the held CI-headroom proposal with normal and slow regimes, including the successful identical-matrix PR115 baseline and this parent 90cc success, as already required by `transaction-ci-headroom-held-disposition.md` and its external review. Do not automatically raise 15 to20, change per-request/cohort deadlines, drop controls, or infer a pass for an interrupted final control.

## Retained evidence and review limits

New read-only downloads saved alongside this report:

- `transaction-sequence-refusal-ci-first-job.json`, `-first-run.json`, `-first-annotations.json`.
- `transaction-sequence-refusal-ci-first-raw.log`, SHA256 `70e5367d1410aec7f4b71aff35946032728e45cc375f0a041cccbfa18891ee57`.
- `transaction-sequence-refusal-ci-original-success-raw.log`, SHA256 `369098baa8eacc1bfa767f7914cbe91b1ebef0219b15a9c764ea047245fcae8d`.

Original saved logs and success metadata remain unchanged. Used GitHub read-only job/run/annotation/log APIs, local log filtering and duration/hash arithmetic, and read-only workflow/test/config inspection. No source edit, test, build, service, install, rerun or outward mutation occurred. This report does not authorize a retry; root owns that action.
