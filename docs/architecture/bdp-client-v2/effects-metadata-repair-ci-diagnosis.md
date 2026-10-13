# PR99 b604 CI failure diagnosis

Read-only investigation of run [38479779925, job115579254122](https://github.com/gastownhall/bdp/actions/runs/38479779925/job/115579254122), exact source `b604320267dd0adcf1ec91854550fb402a1155e7`, tree `04d6757bb4858297dfd5dc5e85576c30d1d1936c`. The frozen checkout remains clean. No tests, reruns, source edits, pushes, workflow mutations, or shared database access were performed.

## Concrete failure

The complete job log identifies a normal test assertion failure, not a job timeout or runner disappearance:

- `apps/bdptest/src/read-update-matrix.test.ts:2217`, test **detects retained-retry 'retained failure gains instance' after actual wire capture**.
- `report.counterfactualApplications` was **0**, expected **1**. The intended corruption was never successfully applied.
- First reported22:52:26.0691882Z; test duration2155ms. Final failure summary22:55:37.8801769Z.
- `pnpm test` completed exit1 at22:55:37.9932767Z, duration465.92s: **8932passed,1failed,1skipped;161passing files,1failed file**.
- Job completed22:55:41Z, before its15minute deadline. Preflight, smoke and Read evidence gate were skipped after the failed test step; they did not fail independently.
- The installed set test visible at the end of the earlier saved log subsequently **passed**, reported22:52:09.4136642Z. Its last visible stdout line was not the failure location.

The check-run annotation is only `Process completed with exit code 1.` Its other annotations are action-runtime/runner-image deprecation notices. No timeout/OOM annotation appears. Check output has no additional test diagnostics; run artifacts count is0.

## Why the original log did not show it

The original `effects-metadata-repair-ci-failed.log` ends22:51:26.6211359Z, before the failing test. It is preserved unchanged:1,416,310bytes,5248lines,sha256`8cd2c2af4a842f2d42b1d1ddb199699114e06f1f0e9798756a78b37457bb53b9`.

GET of the full job-log endpoint recovered2,926,490bytes/10604lines, through22:55:38.4498361Z,sha256`6aac300f42fbc0e03998f823dce6633a0faafdbee2507b5bfcf98e410233385d`. This proves the earlier local extraction was incomplete; its exact truncation mechanism was not established. The full log and an ANSI-stripped copy are retained in `effects-metadata-repair-ci-diagnosis/`.

## Failure path and present limit of diagnosis

The control targets `read-update.idempotency.cross-carrier`, action `g-replay`, zero-based action35of40 in `packages/conformance/matrices/read-update-v1.json`. The real action POSTs updateBead for `beads/retry-cross-g` with retained key`retry-cross-g`; expected response is the retained404Problem. The test clones the genuine response and adds `/body/instance`, then expects the exact deterministic `complete-problem-body` oracle to reject it.

Relevant source:

- App test2042–2048: target control;514–520: await actual client execution, then choose the current client action;545–568: exact action/input match, body-object traversal, injection, increment.
- App test2217: count assertion precedes the selected row/action assertions. It has no diagnostic context.
- Runner642–666/673–787: first action/assertion error stops the scenario and returns a genuine failed/harness-error result. Runner1023–1039 and1071–1093 retain the attempted action and execution-error category;1116–1123 stop on failed assertions. There is no continue-after-error path that would silently advance this row to a later action.
- Exact client helper awaits actual HTTP, parses its body and records a safe observation before returning. Its request timeout is5000ms; generic action default is30000ms. The recorded2155ms test duration is not an explicit timeout failure witness.
- App test925–968 builds and returns the genuine result, including actions/observations, but this counterfactual test does not persist it. The optional `BDP_RETRY_REPORT_PATH` writer at2003 belongs only to the separate positive test.

**The first failing action and its diagnostic are not recoverable from the retained CI evidence.** Count0 can follow an earlier genuine assertion failure, a preparation/transport/execution error, or a failure at the targeted action before increment. These possibilities are distinguishable in the returned result that the assertion did not print. No artifact or log supplies it, so selecting one as root cause would be speculation.

The adjacent title/extension controls and later controls pass in this job, but they are separate runs against new stores and cannot establish the failing run's cause. The new metadata clone guard is gated on `numericControl === "metadata-refusal-success"`; that option is absent here. This proves that added guard does not directly execute in this failing control, not that the entire integration is necessarily fault-free. CI uses the normal Vitest worker configuration rather than local maxWorkers2; load/transport hypotheses are unproven.

## Recommended bounded next action

Authorize a **test-only observability repair in `apps/bdptest/src/read-update-matrix.test.ts`** before attempting to diagnose with another opaque retry:

1. Select the expected row before checking the application count. Attach a small structured failure context: expected scenario/action; returned row state/category; ordered action IDs; first failed assertion IDs; action execution-error category; last safely observed action/route/status; cleanup status. Use fixed/allowlisted fields and no raw credential headers, exception messages, response bodies or child stderr. Preserve both `counterfactualApplications === 1` and the intended `complete-problem-body` failure assertion.
2. Do not convert an upstream row failure into a passing control, drop the test, classify it as a flake, raise a timeout, change workers, or change runtime semantics on this evidence.
3. After scope approval, reproduce the selected control under the pinned toolchain and retain its source-bound diagnostic result; if it does not reproduce, a subsequent ordinary full CI run can expose the first failing action under the original worker configuration. A local pass would be additional evidence only, not an explanation of this failed attempt. If transport failure is identified, add narrowly bounded safe category capture rather than exposing raw causes; if a previous assertion fails, inspect that actual response/afterstate next.

No runtime fix is justified yet. The actionable proven defect is diagnostic suppression at the first counter assertion; the underlying execution defect remains bounded uncertainty. This failed attempt must remain recorded as failure regardless of any later passing run.

## Retained evidence

`effects-metadata-repair-ci-diagnosis/` contains `job.json`, `check-run.json`, `annotations.json`, `artifacts.json`, `full-job.log`, `full-job.clean.log`, and `evidence-digests.json`. All GitHub requests were reads. Local source status was rechecked clean after inspection; no owned test/child/caffeinate processes were launched.
