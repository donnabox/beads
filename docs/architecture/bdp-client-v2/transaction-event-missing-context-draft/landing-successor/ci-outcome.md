# PR127 successor exact-head CI outcome

**Success.** Run38731731790/job116410895014 completed2026-10-15T12:14:35Z successfully on exact head `db5a4d6a41814e59704ebef60fda02a2d57e9ad2`. All27 recorded job steps completed successfully, including downstream ready preflight, installed smoke and Read-cohort evidence gate. This monitor made no rerun, cancellation, merge, source edit or remote mutation.

## Identity and actual checkout

Captured terminal PR base `88c9408e908590ab1a4c64c9b9dee1a8b705999c` and headdb5a4d6a match the run. Ordinary candidate chain is88c9408e →3df82ac4 →db5a4d6a; final candidate commit changes only timeout/comment in development-contracts.test.ts. Candidate tree is `2e16329e195eb8214836de20340a5887236d78f0`. Actual raw log94/401/429 records fetching/checking out synthetic `52a71dbbfb6a4d9cbf231bc057189dee369c1cf0`. Its Git API record has ordered parents[88c9408e,db5a4d6a] and the exact candidate tree. CI uses the synthetic checkout; historical local installed observations keep their original source attribution.

During report assembly root independently landed the branch: current local cleanHEAD `015cac1aa4fa827cc4a2c647dff08454c6c8e6f8`, same tree2e16329e and ordered parents[88c9408e,db5a4d6a]. This is a separate local Git observation, not this monitor performing a merge or relabeling the CI checkout. A stale currentHEAD==candidate assertion stopped the first report assembly before output; immutable candidate checking replaced that read-only assumption.

## Actual callback and suite outcomes

174files passed;9296tests passed,1skipped,0failed,9297total. Main suite512.80s (raw11079–11082). Previously failing callback passed2213ms at5478: `keeps an actual evaluator work refusal local and permits correction using the same unclaimed key`. Original failed5029ms/default5000ms evidence remains in predecessor landing/. Success does not establish the original delay cause.

Four immutable owner-context expansions passed: finite Bead9ms, finite Link8ms, SSE Bead12ms, SSE Link11ms (4119–4122). `serves contextual missing Event responses for Bead and Link JSON and SSE GET and HEAD` passed322ms (4126). `observes genuine installed missing Event context across Bead and Link JSON and SSE GET and HEAD` passed789ms (7165). Exact full names are in ci-summary.json.

The skipped case remains `scripts/demo-read.test.mjs > runnable Read walkthrough > matches real bd and preserves the established seed command recipe` (10456), not executed credit. Complete raw console output is preserved. Actions may redact response values; no repaired console JSON is presented as a complete wire witness. Passing CI installed-case rows are workspace/synthetic-checkout evidence, not another independently packaged-consumer witness.

## Downstream gates and timing

Ready preflight passed7cases/1file in1.04s (11095–11106). Smoke passed installed help/version/diagnostics/lifecycles, configuration round-trips and fail-closed admission checks (11116–11133); its installed reference observation remains claimEligible:false. Read-cohort gate emitted evidence.verified with89rows (11142), preserving its historical evidenceCommit/runHead labels. No release, Preview2 promotion or Transactional qualification is inferred.

Job12:03:04–12:14:35 =691s,509s below configured1200s as an observation, not future guaranteed headroom. Main test step12:05:49–12:14:22 =513s; setup/static before tests165s; downstream/completion13s. All step durations:

| Step | Outcome | Seconds |
|---|---|---:|
| Set up job | success | 2 |
| Run actions/checkout@11d5960a326750d5838078e36cf38b85af677262 | success | 1 |
| Run pnpm/action-setup@b906affcce14559ad1aafd4ab0e942779e9f58b1 | success | 2 |
| Run actions/setup-node@49933ea5288caeca8642d1e84afbd3f7d6820020 | success | 6 |
| Run actions/setup-go@924ae3a1cded613372ab5595356fb5720e22ba16 | success | 11 |
| Run test "$(node --version)" = "v24.16.0" | success | 0 |
| Run test "$(pnpm --version)" = "11.20.0" | success | 0 |
| Run sudo apt-get update | success | 6 |
| Run sudo apt-get install --yes libicu-dev | success | 1 |
| Run mkdir -p "$RUNNER_TEMP/bdp-tools" | success | 0 |
| Run go install github.com/steveyegge/beads/cmd/bd@v1.0.5 | success | 126 |
| Verify the Go-installed source-build bd identity | success | 0 |
| Run pnpm install --frozen-lockfile | success | 2 |
| Run pnpm format:check | success | 1 |
| Run pnpm lint | success | 1 |
| Run pnpm check:boundaries | success | 2 |
| Run pnpm typecheck | success | 2 |
| Run pnpm build | success | 2 |
| Run pnpm test | success | 513 |
| Run pnpm e2e:ready:preflight | success | 2 |
| Run pnpm smoke | success | 6 |
| Read-cohort evidence gate | success | 1 |
| Post Run actions/setup-go@924ae3a1cded613372ab5595356fb5720e22ba16 | success | 0 |
| Post Run actions/setup-node@49933ea5288caeca8642d1e84afbd3f7d6820020 | success | 0 |
| Post Run pnpm/action-setup@b906affcce14559ad1aafd4ab0e942779e9f58b1 | success | 0 |
| Post Run actions/checkout@11d5960a326750d5838078e36cf38b85af677262 | success | 0 |
| Complete job | success | 0 |

## Artifact hashes

- `ci-initial.json`: `d86d8d26be442dc4e023109ef71a8c8596bcd2422d9ad12c8355c0d282582b29`.
- `ci-jobs-initial.json`: `996c84fa05475694c28c3176c7a213b8de594bd7293d5900b34d1511cc9327be`.
- `pr-initial.json`: `35f0be1017e70b8c5a78112d5a6c1e91f6273ea5178cc3c12fea0ceaffc100d5`.
- `ci-terminal.json`: `cc27ae109b20c8bf21d395905526dbe0b56aad5657a2c999f02ff3054d7ce714`.
- `ci-jobs-terminal.json`: `13a2f8ba76ab2c2f39ae4f0a83ccd498c423df4ef0b60e590e2ad661e52fb9e0`.
- `pr-terminal.json`: `9611d92c61ba1e67185b726320824c1520b422a085c5c22e36d386a17cca63d5`.
- `synthetic-commit.json`: `a3a3f51f5f91b8cdd7adfe95a57bbf6a9651db477f88015747c2225a24cf2088`.
- `ci-full.log`: `df650156fef49157ca889509fd3ad703fc47c051a475dc59e7b7be63bd9310b6`.
- `ci-summary.json`: `ade8df6f83624a3abfd9def367c594cc3465fee6a2e028dfa4cbed8d8fd35f16`.

Full raw job log:3,312,512bytes from actual job logs endpoint after completion. Snapshots spaced90seconds. Root owns guarded landing/remote merged-tree verification; this report records exact CI and separately observed local merged tree.
