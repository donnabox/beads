# PR135 successor actual CI outcome

Exact repaired-head CI succeeded. This audit reads root-retained complete terminal evidence; it performs no additional API capture, test, retry, readiness, source/gate mutation or merge. It is an evidence audit by the scenario author, separate from independent code council.

Run38809393968 attempt1 / job116678483498 completed success for `8833c9242b7788e843e9eb97a2c99b2277fe75ef`. All27 actual job steps and captured check runs succeeded, including preflight, smoke, Read evidence and runner post actions. Complete raw log:3,420,544bytes/11,227lines, SHA`856d5c3259216e47714fbbc08d4de0037fcf82afd328ea2c63cca9829dc06e18`. Downloader exit0 and full retained byte count/hash match capture record. Whole-log structural scan found no GitHub error annotations; targeted full tail and new-test records were read directly rather than inferring green from summaries alone.

Actual results:

- Full suite:9312passed,1skipped (9313total),181files passed;790.62s wall duration. The1755.08s summed test duration is not wall time. The skip is scripts/demo-read.test.mjs / runnable Read walkthrough / matches real bd and preserves the established seed command recipe (line10495).
- New callback `preserves one guarded batch winner and rolls back the losing marker through installed HTTP` passed1239ms (line4437). Its actual safe report line4435 records defaultpositive,7completed responses, synthetic source2e96758f, clean workspace, workspace-emitted execution, child17907 exited0/closed/store absent, no primary or cleanup failures, and all repaired replay/full-oracle markers completed. No separate guardless-control CI execution is claimed; fresh EP2 positive7/control3 is separately audited and attributed.
- E2E readiness preflight:7passed/1file,1.43s (lines11157–11168), including descendant supervision and packaged CLI cases.
- Smoke passed all three installed executable help/version/diagnostic and lifecycle checks, config round trips, fail-closed serve behavior, public Read discovery clean stops, installed reference observation and role isolation (lines11178–11195). This does not admit public Transactional claims.
- Read evidence actually emitted evidence.verified with89rows,47projectiondefinitions/46reachable (line11204). This was a positive verification record, not merely the workflow's allowed no-claim exit2 fallback.

Workflow remains byte-identical to base with20minute job cap and pinned Node24.16.0/pnpm11.20.0. Job metadata21:22:46Z→21:39:37Z gives1011s (16m51s),189s (3m09s) observed headroom. pnpm test step spans791s; setup/build and downstream time are included in job duration. This single observed margin is not a future performance guarantee.

Warnings retained: format-check one3.5MiB read-update matrix over1MiB formatter-size warning (447files checked, no fixes); lint851warnings/54infos; action tooling url.parse/punycode deprecations and final Node20-action-runtime deprecation warning. Successful exit does not mean warning-free. No warning was edited/suppressed in this audit.

Checkout evidence and immutable synthetic commit agree: `2e96758fcdcc0801522e60040bc229075963e680`, tree `057f879132074166f2c0f643a8b45bbd9dd47c6f`, ordered parents[`9bf601bfb570adae97ce8cdc99caf2cc7a2b34c1`,`8833c9242b7788e843e9eb97a2c99b2277fe75ef`]. Candidate itself is the ordinary repair child of `9e5f040edab88be364f683098ead72f595f0ab55`; feature footprint from9bf contains exactly the new installed test and progress append. Captured PR head8833/base-main9bf and remote main9bf match; these are observation-time facts, not a substitute for root's immediately-premerge freshness check. No merge identity is claimed here.

Hosted cleanup evidence consists of report childexit0/closed/storeabsence, successful post-job actions, credential extraheader removal and runner 'Cleaning up orphan processes' message (lines11205–11226). Remote PID17907 absence or global hosted filesystem/process absence cannot be independently established from this local audit. Local EP2's86sampled/returnedPID absence and owned-store absence remain separate evidence.

Root final council is now accepted for8833 in root-final-council.json SHA`a4ccf51ce99c5c907b6df0527b060461b90ca21c4e98f7bc3525ed3d60d89d55`; all eight cited evidence hashes rechecked. The earlier terminal helper's finalCouncilPending:true is a historical capture field and was not overwritten. Accepted council coverage is original full9e5 review plus actual F1/F2 repair supplement/dispositions; no required repair remains. Root still owns final current head/base/checks and ordinary merged-tree verification. No readiness/merge action or public qualification follows automatically from this report.

Artifact bindings:
- terminal/run.json: `018b5f66c3ce0916611c4580eb10ae9b2fc9b6e3db27d3f07cb81a5390d159f3`
- terminal/job.json: `2f73d2bcbff4afe5eb8774b2124feb525ab2e2ca6ef855cc0bb3f9f8b7bfc77b`
- terminal/pr.json: `5cb786b4c87c71625574242292769a18dff512d4617fc9ef0c7d8a1d2686eaf1`
- terminal/main.json: `e2d24b45fdd7d9e3f1f7370f611188e56e48832c8715e98a660e5b1c85fe2c29`
- terminal/checks.json: `3b69e565c113068177ed6aa4ade0a374253ca373477aef315eff2dc0605816f0`
- terminal/synthetic.json: `8b28bac313743d030de073819627f3dbfcaa0bd47b0d2b1a65ddad6da278aa1f`
- terminal/job.log: `856d5c3259216e47714fbbc08d4de0037fcf82afd328ea2c63cca9829dc06e18`
- terminal/job.log.capture.json: `9419523a586cd2cc6c9c27d9dcf363c0080a8e59aec27c04e89c2ce294166fc8`

Audit timestamp 2026-10-15T21:42:02.973165+00:00.
