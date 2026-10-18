# Preserved full-suite failure audit

**FAILED; later stages remain blocked.** Exact candidate `d7197e97a29bff8f075028419afad05388e2e655`, tree `630a94d0ca1545f898cb8a8b614d4b7d2dd6d454`. The terminal run reports9374passed,1failed,1skipped across187files (186passed/1failed); supervisor elapsed258.730s, Vitest reported257.26s. Raw process exit1 and accepted=false are preserved. The verifier's additional "No successful structured test result" is the expected rejection of the failed suite, not a second product defect. JSON's533suite total counts nested suites and must not be substituted for187files. This was not a supervisor timeout or a hidden retry.

The sole failed case is `bdpbd Read realization passes every checked-in plan through an isolated public bd CLI workspace (requires local pinned bd or BDP_BD_MATRIX_EXECUTABLE)` in apps/bdpbd/src/read-matrix.test.ts. The actual required env was BDP_REQUIRE_BD_MATRIX=1 and the isolated pinned I/tools/bin/bd. Observed error is **bd command failed with exit1** at packages/conformance/test-support/bd-workspace.ts:131. That helper collects stderr but the nonzero branch rejects only the generic exit error; stderr is returned only on success. Thus this record does not establish the underlying CLI cause, command or remedy. Root owns separate diagnosis; no source fix or rerun is inferred here. The skipped case is `runnable Read walkthrough matches real bd and preserves the established seed command recipe` in scripts/demo-read.test.mjs; it is distinct from the required failed matrix case.

## Completed affected checks inside this failed run

All five intended callbacks were independently found exactly once as passed in terminal JSON. Their later packed-affected stage identifiers are used below only to map selected names; **those external-consumer stages have not run**. These full-suite invocations used the workspace emitted child entry; restart explicitly labels itself development-local with no installed provenance. Do not relabel this as the planned external packed witness.

| Selection | Actual passed callback | Duration ms |
|---|---|---:|
| packed-affected-1 | installed Scope/Bead/Link Event SSE matches finite bytes and sparse Bead survives fifteen seconds | 15736.367 |
| packed-affected-2 | installed Event cursors survive restart with Last-Event-ID precedence and source binding | 537.820 |
| packed-affected-3 | serves byte-bounded deep mutations receipts Events and owned-Link Snapshots over packed bytes | 4220.770 |
| packed-affected-4 | serves missing-first current Resources and bounds unproved minimum eligibility | 4484.862 |
| packed-affected-5 | observes retained sequence member conflict and fresh deep execution over packed bytes | 390.980 |

Actual parsed observations:

- `private-installed-sequence-depth` at raw log line4004: 8 attempted/completed responses; passed, correct declared candidate, clean workspace, entry hash verified, no deadline/cleanup failure, child exit0, directory absent, claimEligible=false.
- `private-installed-fresh-depth` at raw log line4020: 13 attempted/completed responses; passed, correct declared candidate, clean workspace, entry hash verified, no deadline/cleanup failure, child exit0, directory absent, claimEligible=false.
- `private-installed-minimum-current-resource` at raw log line4648: 32 attempted/completed responses; passed, correct declared candidate, clean workspace, entry hash verified, no deadline/cleanup failure, child exit0, directory absent, claimEligible=false.
- Legacy SSE at line9425: actual idle15115.311ms, stream15127.547ms, 2comments and one live Event frame; declared candidate/entry hash verified. This callback's record does not contain a complete child/store cleanup ledger; its pass and run-level cleanup must not be inflated into a stronger standalone claim.
- Restart cursor at line9429: five requests (four finite responses plus drained SSE), two sessions closed with exit0/no unknown stop, zero pending operations/failures and removed directory. Actual attribution remains development-local.

## Cumulative and specification boundaries

`observes seventeen partial Transactional contracts through packed-capable actual installed HTTP` passed. The exact source asserts the actual scenario exchange sum234 immediately before printing the terminal summary (apps/bdptest/src/transactional-matrix.test.ts:627), with <=63000ms constraint. Raw log line5438 contains its summary at10649.977ms:241manifest actions,24sessions,25launched/25closed children and claimEligible=false. The summary does not serialize the complete234exchange array; this audit verifies the passed actual exchange-count assertion plus summary, not an independent re-sum of unavailable raw exchanges. Requirement inventory234 is a separate unit, and none of these counts measures profile completion.

| Completed file | Passed cases |
|---|---:|
| spec-layout.test.ts | 12 |
| profile-contract-plan.test.ts | 59 |
| profile-catalog-boundaries.test.ts | 23 |
| schema-bundle.test.ts | 14 |
| transactional-cumulative.test.ts | 42 |

The12 spec-layout cases include preserving every legacy anchor, all eight actual document link targets (including authoritative Write and old navigation), historical Read+Update heading alias, test citation boundary and navigation-only former monolith. Catalog/contract/schema/cumulative passes verify their existing bounded assertions; they do not override the required CLI failure or confer public profile admission.

## Preservation and cleanup

Independently matched raw log hash, terminal JSON/exit, exact launched argv/env, clean candidate status and all701 current source bytes. Supervisor reports no forced cleanup, survivors or runtime directories; actual full-test temp tree has no retained bdp- runtime directories. All16 sampled run PIDs plus five additional child PIDs from the inspected observations are absent in a fresh process snapshot (21distinct). Four recorded targeted store directories are physically absent. This is bounded observed/sampled cleanup evidence, not proof about arbitrary unobserved detached processes.

No later indexed stage has a started record, and rootResultsAccepted does not include full-test. Existing ordered gate rejects every later stage because full-test accepted=false. No gate/source/index mutation, product/helper execution or rerun occurred in this audit. The failed original is preserved for root diagnosis; completed checks are partial results, never whole-suite success.

Pinned original evidence:

- `validation/run/full-test.json` SHA256 `c0765855828309ea59ebd3c7c1695c2d2d3ad0fe118fd9814b6f15e2ec941d3f`.
- `validation/run/full-test.log` SHA256 `666318015c4032a98d7f08fd57c3b696dbae082ac8866210164c6e0a09c86f68`.
- `validation/run/full-test.exit.json` SHA256 `113ba48738d3d0cd7328ca8e02c00a3b8d8c822ac698a3c681eb739c8437a48a`.
- `validation/run/full-test.started.json` SHA256 `9991443f781eee78d699ad95d42020edaa0b4051c212ac09781fadbfe7c520ea`.
- `validation/run/full-test.processes.jsonl` SHA256 `93678dd50713fd4c2879dd96e274b011e55c8db8ec2008032c437acae99c67f0`.

Audit timestamp 2026-10-17T11:07:18.165613+00:00.
