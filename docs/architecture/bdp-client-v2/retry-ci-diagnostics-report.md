# Retained-retry CI diagnostic repair

Frozen ordinary successor **99e28c5155afa31a55868c219f7771fa46bcef1d**, tree `5451341705a7d637b1c1178f637d021155013d79`, parent `b604320267dd0adcf1ec91854550fb402a1155e7`. Checkout `/Users/dbox/repos/wt/vickie/vickie-retry-ci-diagnostics`, branch `codex/vickie-retry-ci-diagnostics`. Clean source; no pushes or other outward mutations.

The original failed CI attempt remains failed and unexplained at the first-action level. See [read-only diagnosis](../effects-metadata-repair-ci-diagnosis.md), full job log and API evidence. This repair improves visibility; it does **not** claim the underlying execution issue is fixed.

Only `apps/bdptest/src/read-update-matrix.test.ts` changes (+138/-2): a source-local diagnostic projection used by the retained-retry counterfactual table, plus one shape/redaction regression. It reports the selected scenario/action, returned state/category, HTTP exchange IDs/status/failed assertion IDs/error categories, client action IDs/failed assertion IDs/execution category, last actual exact-client route/status and cleanup booleans. It excludes raw reason/error messages, response bodies, request/response headers and child output. Both credentials from this closed fixture are additionally redacted even if they appear in a URL. The unit test supplies secret-bearing body/header/error/route data, checks the exact output, and checks absent-row/empty-observation handling.

All application-count expectations, exact failure oracles, manifest/fixture/coverage/runtime bytes and deadlines are unchanged. No runner behavior or public API changed. A prior client execution failure may still carry only the runner's generic `execution` category; the newly visible first attempted action and last successful route/status narrow the next investigation without leaking raw causes.

## Verification

Pinned Node24.16.0/pnpm11.20.0; frozen install exit0. Typecheck, build, touched-file Biome format and lint exit0. Lint reports187warnings/4infos (two additional non-null-assertion warnings in the new fixed-index regression, matching the existing test style); these are not represented as a warning-free check. `git diff --check` passed.

The affected real emitted-installed retained-retry table plus the diagnostic unit test passed **23tests,0failed,169filtered**, first on final precommit bytes and again on exact frozen source `99e28c5155afa31a55868c219f7771fa46bcef1d`. The latter has full verbose/JSON reports at `frozen-focused.log`/`frozen-focused.json` and explicit exit0. This includes the CI-failing `retained failure gains instance` control; its local pass is not a root-cause explanation. The one earlier focused run before adding native HTTP-exchange context is retained separately as `focused-before-http-context.*`; final evidence is the frozen run.

No full workspace suite, fresh package installation/cohort, CI rerun or runtime validation beyond this affected emitted-installed table was performed. No test limit/worker policy was changed in source; local tests used maxWorkers2. CI's normal worker configuration remains unchanged and its next result must be evaluated independently.

## Cleanup and remaining action

All invoked commands exited. No owned installed child/test/caffeinate process remains. Source status is clean. `source.json` records exact source/tree/parent, sole changed file and content digest; `summary.json` records stage exits and frozen report digest. Root owns integration and the next CI run. If the original failure recurs, the assertion now exposes the preserved row/action diagnostics instead of only `expected0toBe1`.
