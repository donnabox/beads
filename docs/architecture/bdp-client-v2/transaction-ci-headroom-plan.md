# Transactional CI headroom: conditional, measured repair plan

**Recommendation:** if the unchanged PR116 retry repeats the job timeout, propose a separate reviewed change to `.github/workflows/ci.yml:19` only: `jobs.verify.timeout-minutes: 15` → `20`. Keep every command, test, scenario, environment variable, runtime/HTTP/cohort deadline and final evidence gate unchanged. This is the smallest source change, not a claim that twenty minutes is a mathematically minimal or universally sufficient limit. Do not delay PR116 if its unchanged retry goes green. This document authorizes no source change or rerun.

## Basis and observed cost

Read-only source: `/Users/dbox/repos/wt/vickie/vickie-start-now-resume`, HEAD `8c78cba5d2be5d5612c6a9324ee36ed690fdeb19`, tree `f2bc6fd3750bb7b453984f9927f75e64554c777a`. HEAD/tree verified during this audit. Workflow has one `verify` job, fifteen-minute timeout, frozen Node24.16.0/pnpm11.20.0, Go1.26.2/source-built bd1.0.5, frozen dependency installation, static checks/build, the full Vitest run, then preflight, smoke and Read evidence verification. Those last three checks are required even when all tests pass.

| Measure | Successful PR114 comparison | PR116 first attempt |
|---|---:|---:|
| Run / job | 38572597053 / 115881937796 | 38587341773 / 115933355104 |
| Source | 8d66855ec4e26d3af0b3d7d22d621e1f37005294 | 8c78cba5d2be5d5612c6a9324ee36ed690fdeb19 |
| Job start → test-step start | 144s | 211s |
| Go installation, included above | 108s | 168s |
| Vitest wall duration | 364.92s | 696.14s |
| Final full-suite result | 168 files; 9204 passed, 1 skipped | 168 files; 9216 passed, 1 skipped |
| Sequential Transactional matrix passing-test duration sum | 53 cases; 343.105s | 57 cases; 662.639s |
| Test-step end → job end | 11s | Not completed normally |
| Final gates | Preflight ~1s, smoke ~6s, Read gate below API one-second resolution | All three skipped |
| Job conclusion | Success | Cancelled: explicit fifteen-minute annotation |

PR116 started at15:14:06UTC, tests at15:17:37, and printed its last passing deletion control and full summary at15:29:14. The cancellation immediately followed. Setup211s + Vitest696.14s =907.14s, already7.14s beyond900s before the remaining gates. Rounded step timestamps give a697s test step. Job cleanup ended15:29:19. The cancellation does not negate the actual passing test output, but it is **not successful CI**: preflight, smoke and the Read gate did not run.

The matrix sum is approximately95.2% of the suite wall time, not additional elapsed time to add to that wall time. Other files run concurrently. Its first pass was15:18:30 and last15:29:14; its test durations sum662.639s. Vitest's1623.29s aggregate test-duration statistic is likewise overlapping work, not job elapsed time.

The prior PR115 timeout diagnosis independently found broad slowdown, including unchanged controls and Go compilation, rather than the apparent installed-set hang in the truncated log view. Raw logs continued to show progress. That diagnosis does not establish a specific CPU/I/O/host cause. New cumulative rows and controls also add real work; transient runner slowness is not an excuse to ignore growing consumption. PR116 completed all tests, so there is no evidence here of a hung last test or need to weaken a per-test timeout.

## Option A — one bounded outer-job amendment

Using this attempt's setup and test duration plus the successful comparison's11s downstream cost gives an estimated completed job of918.14s (15m18.14s). This is a cross-run estimate, not an observed PR116 completed runtime. At1200s, remaining headroom would be281.86s (~4m42s; about30.7% of estimated consumption). A sixteen-minute cap would leave only41.86s and provides little protection against the observed setup variation. An eighteen-minute cap would leave161.86s; it is possible but less conservative. Twenty minutes is a rounded, bounded scheduling allowance, not a runtime execution-policy change.

As a sensitivity calculation only, substituting the231s setup observed on PR115 and reserving30s for all final gates/cleanup yields231+696.14+30=957.14s, leaving242.86s under twenty minutes. The30s is an explicit proposed allowance, not a measured gate duration. Neither estimate promises future host performance or permits further automatic increases.

Scope: exactly one workflow literal, optionally a short rationale comment. Preserve job name/required-check identity and every step's success/failure behavior, including `pnpm evidence:verify || test $? -eq 2`. No package script, Vitest configuration, test selection, fixture, artifact, source-runtime or installed-harness edits. All current HTTP, service, child-close, aggregate-cohort and individual-test deadlines stay intact.

Risks: a genuinely wedged command can occupy a runner up to five additional minutes; this does not make tests faster or eliminate continued growth. Existing bounded test/child cleanup remains the primary local failure containment. Record stage durations and the final gates on the exact repaired head. If that head still exhausts the new limit, stop automatic retry/timeout growth and investigate actual progress before proposing a partition.

## Option B — isolate the existing long matrix

Moving it to another sequential step in the **same** fifteen-minute job does not create headroom. A meaningful isolation would use a separate job for the complete `apps/bdptest/src/transactional-matrix.test.ts` file, and an explicit complementary full-suite selection for the remaining files. Do not split individual controls or replace their cumulative scenarios. Preserve all57 current matrix tests, every other selected file, the existing skip, negative controls, and exact source/package/configuration provenance.

This is a larger CI-only change requiring a measured trial. The observed matrix duration662.639s leaves237.361s under a900s job clock before setup, import and cleanup. Repeating all current211s setup would leave only26.361s before that remaining overhead; blindly duplicating setup could reproduce the problem. A separately measured matrix job may omit the Go/bd setup only after verifying its transitive test/support dependencies do not require it; lack of a direct `BDP_BD` string is not that proof. The complementary suite still requires the bd executable and its existing identity check. Alternatively share exact built artifacts, but that introduces transfer/provenance validation and is not the smallest repair.

Before acceptance, prove disjoint/exhaustive test-file selection from the configured discovery list (currently168 files), and reconcile each test inventory and skip rather than merely summing pass counts. Preserve clean-source build-before-emitted execution and actual installed children. Keep preflight/smoke/Read checks attached to successful prerequisites; a final required aggregate check must fail if either job fails or is cancelled. Existing branch-protection check naming needs root verification. Measure actual parallel runner costs, matrix setup/import/cleanup, and total workflow time; do not infer isolated speed from overlapping full-suite timing. No concurrency change inside the matrix is proposed.

This option could remove setup and broad-suite contention from the long tail and improve wall time, but this audit has not measured it. It is a follow-on if bounded job headroom remains inadequate, not a prerequisite for an unchanged green PR116 retry.

## Review, validation and stopping predicates

1. Root records the unchanged retry's actual result separately. A green result ends the immediate PR116 gate; it does not rewrite this failed attempt. A repeat timeout triggers consideration of Option A, not automatic authoring or repeated retries.
2. For Option A, review the exact workflow-only diff against this source and confirm no test or execution deadline changes. No local9216-test rerun is needed merely to verify a YAML timeout literal. Validate workflow syntax and inspect the actual configured job limit on the subsequent exact-head CI run.
3. Require that exact run to complete the full discovered suite **and** preflight, smoke and Read evidence gate with their existing semantics. Retain raw logs, annotations, stage durations, source SHA and any failures. A cancelled run with all tests passed is still not CI success. No additional public profile/conformance claim follows from green CI.
4. Stop on real test failures, incomplete cleanup, continued cancellation, or an altered test inventory. Diagnose that evidence rather than changing runtime/test deadlines or excluding scenarios. Root owns final acceptance/publication.

## Evidence and limits

Local inputs read without network or test execution:

- `transaction-start-now-resume-ci-first.json`, SHA256 `93700749627d392d9423c512dcffdc67b51328c5540443d9918f45582736e6fd`.
- `transaction-start-now-resume-ci-first-raw.log`, SHA256 `d0c70ea7d8c621df281e69cf9f96bc742201dfde6ee36bd0ba9626a35ab767b6`.
- `transaction-start-now-resume-ci-first-annotations.json`, SHA256 `c0e2ec5812e20545eb03bd0ce2bd15970e0515598cfc7382d7254690152ba69d`.
- `transaction-start-now-resume-ci-timeout-disposition.md`, SHA256 `ff5123c9c0e900776f8caaa7d609739d3a2353b0f28601e52870ac41f65ec3ec`.
- `transaction-deletion-safety-ci-pr114-job-raw.log`, SHA256 `e1f1e82caf1b9b0a8f9d57797a23d5cc65b87654583d84a19d0df965fe169ab7` and its saved comparison metadata.
- `transaction-deletion-safety-ci-timeout-diagnosis.md`, SHA256 `9d2dbdc57668905e63840362bc05b03f6f707ac48d311c51b978de6001f4564e`.

Local inspection used Git/read commands and Python to strip ANSI, omit huge JSON lines from display, extract matrix pass durations, hash evidence and check arithmetic. No source edits, tests, builds, installs, network calls, workflow reruns, or outward mutations were performed. This plan is the only written artifact; no owned long-running processes were started.
