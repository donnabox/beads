# PR120 W1b CI failure: bounded diagnosis and proposed reproduction

**Merge remains blocked. This is a real unexplained test failure, not a job timeout or an established flake.** The existing CI log proves that the link-metadata counterfactual did not increment; it does not preserve the returned scenario report needed to identify the first failed prerequisite. Do not rerun CI or change behavior on that evidence alone. One instrumented, exactly selected local diagnostic execution is proposed below; nothing has been executed for this audit.

## Exact evidence

Frozen source: `/Users/dbox/repos/wt/vickie/vickie-comparison-timeout-review-repair`, `70ac666002708202d2f93195eed90dfe8541300d`, tree `1353e8f4fb0b1149b3f4ff158335804bf3f1cecc`. CI run 38615422801/job 116029153719 ended in failure. Full raw log `transaction-comparison-timeout-repair-ci-first-full.log` independently hashes to `070cd1f21125d65b97d043488e34621328c0c2f2f4d44536eda4898c96225c31`.

Concise retained facts:

- 18:33:11.445Z: `detects W1b link-metadata assertion-input corruption after actual HTTP` failed after 1435 ms, with counter 0 rather than 1.
- Final failure points to read-update-matrix.test.ts:1642, immediately after `await run(...)`; row/action assertions below it were never evaluated by the test.
- Final totals: 9255 passed, 1 failed, 1 skipped; test duration 485.16 s; pnpm test exited 1. This is not the 30-second case timeout.
- The prior30035 raw CI log records this exact case passing at 18:07:36.866Z in 1604 ms. It records a different service-test failure overall, so that run is not a green baseline. Current sibling stale-unowned-link and stale-link controls passed. Those observations show successful executions on related paths, not why this one failed or whether another run will pass.

Read-update matrix test, exact client adapter, runner, manifest and fixture are byte-identical to30035. Their current SHA256 values are respectively `0f2d05a40007868ad59c5fe09f1a2fd3573932b76815bdc7700994e7dee2d4a9`, `fe7568b3fe7e443873f29c4d38b2115d3763190092f0b084671e39d0e97dbf8d`, `ac20c7fbfe15a5e42a6559e89cd3467562efe9bc74d462478414414ed3efede1`, `c86b33535f25a1d117ce021a3bb70164f5635bdaee50758e3428c130ccc3227f`, and `f89932d68e3bf5867d79eb3de680af4bc67b761b3c3bce5da127f6ca75e27765`. Unchanged source alone does not excuse the failure or prove environmental causation.

## What the zero counter actually establishes

The parameter row at read-update-matrix.test.ts:1623 selects control `link-metadata`, scenario `read-update.singleton.update-link`, action `owned-metadata`, expected failed assertion `metadata`. The hook at 854–880 requires POST with idempotencyKey `owned-metadata` and counter still zero. The manifest has exactly that input at action **36 of 63**. There is no selector/key mismatch apparent in frozen source.

`actionExecutor` first awaits the real client at 529. The counter increments at 880, before metadata mutation at 899–902. Therefore an exception while modifying the returned metadata cannot itself explain a final counter of zero. Either execution never reached this action/hook, or execution failed before reaching the hook (including the awaited client or preceding processing). A returned wrong HTTP status alone would still increment if the client returned normally and the matching action was reached.

The runner executes actions sequentially at runner.ts:638–665 and stops the row on a thrown assertion, capture or execution failure. Programmatic failures are recorded at 1074–1095; failed assertions stop before captures at 1118–1128. The exact client can throw on transport/exact serializer/JSON processing before it returns (read-update-exact-client.ts:118–180). Earlier HTTP assertions, capability/preparation/prerequisite failure, binding capture failure or an action execution error can therefore prevent action36. The scenario has no prerequisiteScenario references and is selected alone; it does not automatically run the whole matrix.

Crucially, `run` returns genuine result/actions/observations at matrix.test.ts:931–976, with cleanup handling through1022. The test currently discards that diagnostic detail by first asserting the bare counter. CI did not print it. The exact blocking action, status, assertion or error category cannot be recovered from the retained log. CPU pressure, random port collision, deadline, revision collision, regression and flake are **not established causes**. The next step should expose the first actual failure rather than choose one of those hypotheses.

## Proposed single diagnostic execution, subject to root release

Prepare an isolated ordinary diagnostic checkout of exact70ac; preserve the frozen reviewed checkout. Permit only observational instrumentation of this existing parameterized test for control==link-metadata. Keep its input/control mutation and every existing assertion unchanged. Immediately after `await run`, retain the actual returned report before the counter assertion, whether the test will pass or fail. Record the selected scenario result (state/reason/category, ordered exchanges, action IDs/assertion outcomes/execution errors), real `observations`, setup/native observations, counter, cleanupComplete, source/entry/configuration/payload identities and sessions. Do not replace failures with synthetic rows or relabel assertion-input corruption as actual wire corruption.

Use the existing optional lifecycle callbacks (run signature at170–178) only to retain launched child PIDs, returned report and cleaned status. The returned callback occurs before outer cleanup, so preserve that snapshot separately; record the post-run cleanupComplete value afterward, not a fabricated true flag. Capture the top-level thrown error if run cannot return. Do not write credential values/configuration credentials or raw authorization headers; the exact client already exposes safe synthetic request/output observations. Save diagnostic JSON locally outside source. Pin the instrumentation diff and its hash; mark the run as diagnostic-source instrumentation, not unchanged-source conformance.

Select only:

`apps/bdptest/src/read-update-matrix.test.ts -t '^detects W1b link-metadata assertion-input corruption after actual HTTP$' --maxWorkers=1 --reporter=verbose --reporter=json`

Use the repository's pinned Node/pnpm and existing emitted dependencies. Before release, verify default entry binding and clear/record any BDP_CUMULATIVE_INSTALLED_ENTRY override; hash the actual emitted entry/payload and diagnostic inputs. Any needed offline setup/build is a separately enumerated prerequisite, not an implicit package run. No pack, new installed candidate, full suite, load simulation, automatic repetition or timeout increase is proposed.

## Execution and cleanup budget

One selected Vitest case, one selected matrix scenario. The manifest has six raw HTTP actions and57 installed.exact-request client actions, each producing at most one exchange, with no redirects/retries. This row is not in fixture.readSeedScenarios, so there are no hidden seed HTTP requests. Bootstrap performs ready/identity checks via IPC only.

Expected successful counterfactual execution stops at action36: **36 HTTP exchanges**, then the intended metadata assertion failure inside the matrix makes the outer test pass. A diagnostic failure can stop earlier. Hard source-derived ceiling: **63 HTTP dispatches**, if the hook unexpectedly does not fire and the complete row runs. Record attempts separately from successful observations, since a client throw can occur before its observation callback. The diagnostic should adjudicate the limit rather than silently omit failed attempts.

Exactly two sequential real child launches at most: bootstrap, then scenario. Two configured persistent store directories beneath one temporary root; two child listener lifetimes, plus the existing short-lived availablePort reservation probe (no HTTP). No child-per-action launch or additional scenario. Existing raw HTTP timeout is5000 ms, existing child startup/close bound10000 ms, selected Vitest timeout30000 ms. Keep all unchanged; propose120-second outer supervision with owned-descendant cleanup. Request body/header caps remain131072/16384 from the existing exact mode; do not claim these are measured request sizes. Preflight should bind actual planned routes/body sizes and diagnostic-output limits before release.

Retain both actual child PID/exit/signal records and independently verify absence. Existing close requires exit0 or SIGTERM; forced SIGKILL does not satisfy graceful cleanup. Record directory-removal evidence or its exact limitation, not merely report cleanupComplete. On unexpected timeout/failure, stop and preserve artifacts; no second run is automatically authorized.

## Decision after the one run

If counter remains zero, use the retained row's first failure and last real observation to identify what prevented action36, then propose the smallest evidence-based repair/next diagnostic. If counter becomes one and the exact metadata assertion is the only failure, the selected reproduction passed; the original CI cause remains unresolved. That outcome alone is not proof of flakiness and does not clear merge. Root decides any further investigation or retry from the preserved evidence. No production, protocol, maintenance-plan or PR120 source change is proposed by this initial audit.

Read-only source/log/manifest inspection only. No source edits, tests, builds, installs, services or network actions were performed. Only this diagnosis-plan artifact was written.
