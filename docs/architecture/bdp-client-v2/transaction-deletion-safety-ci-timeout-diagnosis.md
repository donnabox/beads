# PR115 CI timeout: bounded read-only diagnosis

**This was a job-budget exhaustion while tests were still progressing, not the apparent six-minute installed-set hang.** The original `gh run view --log` export truncates the test step at the installed-set stdout header. The direct job-log API contains thousands of later lines, including passing tests ten seconds before cancellation. The same truncation occurs in the successful PR114 comparison, so the apparent silence is not evidence of a new deadlock.

## Exact evidence

PR115 run `38581374707`, job `115911832057`, source `39c160d0d35294003f400cef8d5ab5c6ec9aec9c`, cancelled under the workflow's 15-minute job deadline. Root separately confirmed the explicit timeout annotation. Existing run JSON and source `.github/workflows/ci.yml:19` agree on the bounded job setup; test time is not the entire job time.

Read-only API downloads saved as:

- `transaction-deletion-safety-ci-job-raw.log`: SHA256 `47dccf6bfe8a5597c71bc3194a2056e084882330c37a10f0fbee6b0ffa12c948`.
- `transaction-deletion-safety-ci-pr114-job-raw.log`: SHA256 `e1f1e82caf1b9b0a8f9d57797a23d5cc65b87654583d84a19d0df965fe169ab7`.
- Prior run metadata and the truncated comparison view also retained as `transaction-deletion-safety-ci-pr114-comparison.{json,log}`.

Commands: `gh run view 38572597053 --repo gastownhall/bdp --json headSha,conclusion,jobs`, `gh run view ... --log`, and `gh api repos/gastownhall/bdp/actions/jobs/{115911832057,115881937796}/logs`. All read operations exited 0. No rerun, source edit, timeout change, build, test process or outward mutation was performed. Diagnosis guidance used; reproduction/instrumentation intentionally deferred under the explicit read-only assignment, so this report does not claim a proven host-level performance cause.

|Measure|PR114 successful run38572597053|PR115 cancelled run38581374707|
|---|---|---|
|Source|8d66855ec4e26d3af0b3d7d22d621e1f37005294|39c160d0d35294003f400cef8d5ab5c6ec9aec9c|
|Job|13:33:24–13:42:05 UTC|14:32:10–14:47:27 UTC|
|Go install|108 seconds|188 seconds|
|Tests step|13:35:48–13:41:54, 366 seconds|14:36:01–14:47:24, cancelled after683 seconds|
|Vitest result|168 files;9204 passed/1 skipped;364.92s|No final result; cancellation, not a test assertion failure|
|Installed-set pass|13:39:14;1366ms reported test duration|14:42:06;2390ms reported test duration|
|Last visible passing test in raw log|Final no-op group/head control at13:41:54|No-op Bead-property control at14:47:14,13061ms|

The truncated views stop at13:38:53 and14:41:19 respectively. Raw PR115 has10,947 lines and about3.09MB, versus6,871 lines in the original run-view export. Its longest line is637,707 characters. That large stdout record is a log retrieval/formatting boundary; this audit does not claim the precise gh implementation cause. Importantly the underlying test proceeds past it in both runs. `transactional-set-installed.test.ts:441–453` prints this large report only after child close, deadline clear, directory removal and failure checks. The raw log independently records its later passing test line.

## Throughput and workload

Paired timing extraction from raw passing-test lines (same full test name; prior duration at least100ms) finds554 common tests with median PR115/PR114 ratio1.877. The slowdown is broad:159 server tests median1.952;35 conformance tests1.733;184 Read+Update matrix tests1.856. Their summed durations are overlapping worker time, not wall clock. Unchanged Go installation is also1.74 times slower. These comparisons favor a slower runner/resource-throughput episode over a new isolated deletion algorithm hang, but do not identify CPU contention, I/O or host scheduling conclusively.

The candidate does add real work: one new row is included in repeatedly executed cumulative Transactional cohorts, and three new deletion controls each execute that cohort. Parent-to-candidate delta is exactly four conformance/artifact/test files; runtime, installed-set test, workflow, lockfile and Vitest configuration are unchanged. This additive work compounds the broad slowdown and reduces existing job headroom; it cannot simply be dismissed as infrastructure noise.

The Transactional matrix is the long tail. Prior53 named passing tests sum343.105seconds; current51 completed tests already sum635.637seconds. Examples of unchanged controls rise from roughly6–10seconds to12–21seconds. The canonical positive rises9585→19849ms, with its new row included. At cancellation the Bead-property no-op control had just passed; the two later no-op controls, deletion static pin and three new deletion controls have not produced pass lines. The new deletion controls are therefore not themselves shown hanging or failing. No failure marker or individual-test-timeout appears in the raw cancelled log.

`package.json` runs `vitest run --reporter=verbose`; `vitest.config.ts:62–72` has no configured worker count or file/test concurrency override. Existing per-test order makes the many installed controls in one matrix file a sequential tail; `.each` deletion controls at `transactional-matrix.test.ts:1626–1707` reuse the full observer under unchanged120s test bounds. The installed-set test itself retains90s cooperative/120s outer bounds. The last active passing control completed normally well inside its own bound. Increasing those per-test deadlines would not address this job-wide exhaustion.

## Recommended next action

1. Preserve this cancellation as failed CI evidence and use the **raw job API log**, not the truncated run-view export, for subsequent diagnosis. No frozen source change is justified by the supposed installed-set stall.
2. Root may authorize one unchanged-head CI rerun to distinguish transient broad runner slowness from a repeatable job-capacity failure. This is a proposal, not an action performed here. A green rerun would establish that exact-head CI result, not erase the measured headroom problem.
3. If the slow long tail repeats, make a separate reviewed CI scheduling/workload proposal: isolate the existing long installed-matrix workload from setup/broad-suite time, or partition its independent existing tests while retaining every scenario and negative control with exact-source reports. Measure a proposed schedule before choosing concurrency; indiscriminately adding workers may worsen resource contention. Do not prune required observations or change runtime/HTTP/cohort deadlines. A job-only timeout amendment is a root decision requiring explicit measured rationale, not an automatic repair in this audit.

The logs support continued throughput until cancellation. They do not support a deadlock claim, a specific production regression, a completed PR115 test suite, or proof that all outstanding controls would have passed. Root owns rerun/scheduling disposition and readiness.
