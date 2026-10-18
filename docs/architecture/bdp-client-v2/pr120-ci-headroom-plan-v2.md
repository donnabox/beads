# PR120 bounded CI headroom plan v2

## Proposed decision and exact scope

Approve, after independent review, a separate ordinary successor of frozen `427cfef6decbb422b393024e36218affc26007fa` (tree `9075897867d88344a8477e1f4a8e1a00a50f1228`) changing **only `.github/workflows/ci.yml:19`, `jobs.verify.timeout-minutes: 15` to `20`**. One file, one literal; no additional comment is needed. This document authorizes no source edit, push or run. Root owns those subsequent gates.

Keep the `verify` job ID, required-check identity, runner, action pins, versions, Go/bd source build and identity check, frozen dependency installation, all commands/environment, full test selection, all per-test/HTTP/service/child/cohort deadlines and final checks exactly unchanged. Preserve `pnpm evidence:verify || test $? -eq 2`, including failure on errors other than the existing no-claim exit2. No workload is added, omitted, partitioned or retried by this patch. No production/runtime or test edits accompany it.

The reason is demonstrated outer-budget exhaustion during continuing work, with a wide observed throughput range for the unchanged long matrix. Twenty minutes is a bounded scheduling allowance supported by the measured range, not a promise of future performance, a fix for the unknown original70ac W1b cause, or permission to relax any test assertion or inner deadline.

## Evidence and attribution

PR120 run38620066822/job116044976221, head427, ended cancelled. Its check-run annotation explicitly reports the15-minute maximum. GitHub checked out merge `088fec9496fc11ee00ac4ee19b5757fe2daf7d4d` of427 into main `7e46cdc93b2fd02094ba7e15e57e743d2ef81daa`; raw installed sourceCommit fields identify that merge. The run is associated with427 but is not a standalone427 checkout attestation. Full diagnosis: `pr120-427-ci-diagnosis.md`.

Current setup/static/build all succeeded. Test step consumed703s before cancellation; no final suite summary. ANSI-stripped log has9255 completed test pass rows, one existing skip, no failed row. W1b link-metadata passed2630ms and diagnostic secrecy unit passed3ms. The three final gates were skipped. This is unsuccessful, incomplete CI. The last reported matrix pass was outgoing-order at19:15:58.796Z; cancellation19:16:10.822Z interrupted before any completion report for committed-prefix-state. No final-case result or cleanup proof is inferred.

| Observation | Setup to test step | Go install included in setup | Full suite / outcome | Transactional matrix |
| --- | ---: | ---: | --- | ---: |
| PR115 retry,39c160d, job115925110459 |142s |103s |384.25s Vitest;9213 pass/1 skip; final gates green;538s job |57 cases,364.607s |
| PR116 first,8c78cba, job115933355104 |211s |168s |696.14s Vitest;9216 pass/1 skip printed; job timeout; final gates skipped |same57 cases,662.639s |
| PR119 retry,c555fe0, job116007773987 |156s |123s |409.53s Vitest/410s test step;9245 pass/1 skip; final gates green;578s job |same57 cases,389.570s |
| PR12030035 first |See retained raw log |Not used for budget estimate |425.44s Vitest;9253 pass/1 fail/1 skip; real service test failure |same57 cases,401.519s |
| PR12070ac first |See retained raw log |Not used for budget estimate |485.16s Vitest;9255 pass/1 fail/1 skip; real W1b failure |same57 cases,460.450s |
| PR120427 first |211s |168s |703s test step before cancellation; no final summary; final gates skipped |56 completed cases,658.206s |

The matrix file, workflow, Vitest configuration and package.json are byte-identical from39c160d through427; verified by Git diff, including c555 and70ac comparisons. The full suite has grown outside the matrix, so the older full-suite totals are historical baselines rather than current inventory claims. The diagnostic-only427 patch adds no cases. Current70ac/427 inventory remains172 files/9257 cases as established by the completed70ac run; expected all-pass totals are9256 passed/1 existing skip, plus the deliberately repeated downstream seven-case preflight.

For a like-for-like incomplete-run comparison, the56 completed matrix cases cost658.206s now versus451.450s at70ac and382.556s at green c555: summed ratios1.458 and1.721, median individual nonzero-duration ratios1.486 and1.739. Earlier, the identical57-case PR116 run cost1.817×PR115's green sum. This directly distinguishes throughput variation from added scenario work. The original PR114→PR115 change from53 to57 cases measured343.105→364.607s (+21.502s,6.3%); it does not explain the later near-doubling of identical work. No CPU, I/O, host-contention or specific runner-cause conclusion is established. “Runner variance” here means observed execution-time variation across runs, not a diagnosed infrastructure fault.

Matrix case-duration sums are sequential within that file and useful for comparisons. They overlap other suite files, so they are not additional time to add to suite wall duration. Aggregate Vitest test-duration totals likewise must not be added to wall time.

## Headroom calculation and its limits

Use step-clock accounting and an explicit30s proposed reserve for all downstream gates and cleanup. The reserve is conservative planning allowance, not an observed duration: green PR115 and PR119 consumed11s and12s respectively from test-step end to job completion. Their required downstream commands all ran. Current cancellation cannot supply a completed downstream measurement.

Current setup211s leaves only689s under900s before tests and downstream work, or659s after the30s reserve. The cancelled test step already consumed703s and remains incomplete. Thus the present cap demonstrably intersects the slow regime. At1200s, the same setup/reserve leaves959s for tests,256s beyond the already-consumed portion. Using the independently verified historical worst setup231s (PR115 first attempt,188s Go install) leaves939s,236s beyond the consumed portion. These margins are not evidence the unfinished run would have succeeded.

To express the variance bound against a completed green baseline, use c555's410s step wall: with231s setup and30s reserve,900s tolerates639/410=1.56× that baseline, whereas1200s tolerates939/410=2.29×. At142s setup the corresponding multipliers are1.78× and2.51×. Observed identical-matrix episodes include1.817× (PR116/PR115) and1.721× (current completed56/c55556), so15minutes intersects the observed band while20minutes extends outside it. These are sizing sensitivities: full-suite growth and matrix/suite differences prevent treating them as guaranteed bounds.

A second current-inventory sensitivity uses the completed70ac suite485.16s and the observed current56/70ac56 ratio1.458: estimated707.4s suite, plus231s setup and30s reserve gives968.4s, leaving231.6s under1200. This is a cross-run estimate, not a measured completion time;70ac had a real W1b failure and the multiplier comes from the matrix. It is corroborating sizing arithmetic only. The direct evidence remains interrupted ongoing work and a confirmed outer timeout.

The proposed increase permits a truly wedged job to occupy a runner for up to five more minutes. Existing inner deadlines and cleanup assertions remain the containment mechanisms. Twenty minutes is not mathematically minimal; it gives materially more margin than16minutes while staying bounded. If the revised job exhausts20minutes, stop automatic retries/increases and inspect progress rather than escalating again by default.

## Why not retry unchanged or partition now

An unchanged retry could go green, as earlier retries did, but would leave the measured budget collision intact. This recurrence supports a small reviewed engineering change rather than another run selected for a favorable timing episode. This recommendation does not convert the earlier actual70ac failure into a timeout or claim it fixed. The safe diagnostic messages remain in place for any recurrence.

Moving the matrix to another step in the same15-minute job does not create headroom. A separate job would be larger, requiring provenance and test-selection design. Present both regimes: identical matrix364.607s +142s setup leaves393.393s under900; slow matrix662.639s +211s setup leaves26.361s before import/finalization overhead. Those case sums are proxies for a hypothetical isolated job, not measurements of one. Partitioning offers ample normal-regime margin but remains exposed to the same observed timing variation in the slow regime; splitting alone does not establish faster execution. No partition is proposed here.

Any future partition must prove exhaustive/disjoint current test-file selection and per-case inventories, preserve package build/provenance and bd requirements, and retain an aggregate required check named`verify` or explicitly migrate branch protection. The current literal-only change preserves that name automatically. The existing downstream `scripts/e2e-ready.test.mjs` execution intentionally repeats seven cases already included in normal discovery; retain both executions. Preserve the existing build-dependent demo-read skip, named “matches real bd and preserves the established seed command recipe,” and identical build inputs; do not manipulate build state to create or remove skips. These points close the earlier optional review concerns without adding partition work to this patch.

## Authoring and verification gates

1. Independent native/Claude plan review and root acceptance precede authoring. Prior headroom plan remains preserved; this coherent revision incorporates required findings1–3 and optional4–7 rather than treating the held plan as permission.
2. Author on an isolated ordinary successor of427, one file/one literal only. Verify exact parent, clean initial state, and guarded mutation. Preserve all other tracked object/mode entries. Do not combine maintenance-fault or other queued source work.
3. Perform only static local validation appropriate to YAML: parse with an already available parser, inspect the exact scalar/path and unchanged jobs/steps/environment, and whitespace/diff checks. No fresh9257-case local run, child/service/HTTP execution, build or dependency installation is needed for a timeout literal. If existing parser availability is missing, report it and use reviewed text/structure inspection rather than fetching a new dependency. No test selection or timeout workaround.
4. Freeze the exact source and obtain independent code review of the tiny diff. Root owns publication and a single subsequent exact-head CI attempt. Retain run/head/actual checkout identity, raw untruncated logs, annotations, stage times and conclusion.
5. Require full discovered suite completion with the expected current inventory and all three final gates successful under their unchanged semantics. A printed pass summary with cancelled job is insufficient. Reconcile any legitimate discovery change before acceptance; do not silently sum or waive skipped cases. Do not infer installed child cleanup from runner cancellation/orphan-cleanup text.
6. On real test failure, cancellation, incomplete cleanup or altered workload, preserve evidence and diagnose; no automatic retry, further cap increase or weakened check. No merge/public profile admission follows from this plan alone. Green CI remains one required integration gate and does not erase historical failures or unresolved minimum/Event/public-profile decisions.

## Evidence retained and review disposition

Current artifacts: `pr120-427-ci-first-full.log` SHA256 `752187cef60ddcf05d3149c7f67a9611962126ac58ca81d4a0be63924538c935`; status `0ab28254e8e414f2d0fbba8b1e909a7fa888e5fa8a7d6cf059b880da62340f56`; annotations `6e5d1267716d31073330ff299a82994f3b9523312b7655ca93bdef1d198fd159`.

Independently reread PR115 successful metadata/log: `ea78f2cbc0f0bd3bc2a8fe788eebaa832ff9999495b563ae8bc8da6a54eb6170` / `c66ae06264be2a7b5a6f8bb89d9b1bee57f304ffa0f7a455737d5a8573c583a0`; PR116 first metadata/log: `93700749627d392d9423c512dcffdc67b51328c5540443d9918f45582736e6fd` / `d0c70ea7d8c621df281e69cf9f96bc742201dfde6ee36bd0ba9626a35ab767b6`. Read PR115 first metadata for231s setup, PR119 successful metadata/log,30035 and70ac raw logs. Their hashes and exact observations are in the new diagnosis. Prior PR11453-case comparison is retained in the original plan and completed Claude review, not presented as a newly executed benchmark.

Prior dispositions are explicit: required1 closes through the identical57-case successful baseline and separate growth/throughput attribution; required2 through variance-sensitive step/reserve calculations; required3 through both partition regimes. Optional4 closes through consistent accounting and explicit reserve; optional5 through deliberate repeated preflight; optional6 through named build-dependent skip; optional7 through required-check preservation/migration. No source edits, test/build/service execution, installs, network calls, CI mutation or retry occurred while preparing this plan. Only this new Markdown plan was written.
