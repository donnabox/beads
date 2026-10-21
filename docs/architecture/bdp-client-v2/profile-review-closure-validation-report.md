# Integrated review closure: independent validation

Exact frozen source **31aab4db0f61e3d8b1139a7b78ff1ff10f890836**, isolated checkout `/Users/dbox/repos/wt/vickie/vickie-review-closure-validation`, branch `codex/vickie-review-closure-validation`, GitHub origin. Executed 2026-10-13 using Node 24.16.0 / pnpm 11.20.0. Source remained clean; no edits, commits, pushes, seal changes or evidence relabeling. Native merge/root-diff assessment is separately recorded in `native-review.md`: no actionable integration defect found.

## Checks

Full suite: **8,400 passed, zero failed, one optional skip, 146 files**; 102.11s command wall time / 101.10s Vitest duration. MaxWorkers2; clean build completed before emitted/process tests. Logs and structured results are retained.

| Check | Exit | Wall time |
| --- | --- | --- |
| install | 0 | 0.46s |
| typecheck | 0 | 1.23s |
| build | 0 | 0.8s |
| fulltests | 0 | 102.11s |
| boundaries | 0 | 0.58s |
| format | 0 | 1.28s |
| lint | 0 | 0.43s |
| smoke | 0 | 4.23s |
| smoke-development | 0 | 0.64s |
| preflight | 0 | 1.12s |
| evidence | 0 | 0.33s |

Repository lint exited 0 with 44 warnings and one info; no automatic fixes. Formatting, boundaries, both smokes and preflight passed (preflight seven tests). Read verifier passed its existing 89-row seal `674982be2123b9b7aae630bf79c3f774122900cd`, observed at `e5f7d0c43c151426350ceededdc77fe7e372db9b` with evidence commit `dbaf63a24219e1ea3f80394746853598b0d4464a`. It is not a new Read observation at 31aab4d.

## Fresh packed installed cohort

Fresh protocol/server tarballs were built from this clean exact source and installed in this directory's `packed/install`. Actual target: `/Users/dbox/repos/wt/vickie/bdp-integration-logs/profile-review-closure-validation/packed/install/node_modules/@bdp/server/dist/installed-reference-process.js`. The installed-entry environment prevents workspace-server fallback. Packed regression matrix: **112 passed, zero failed, four files**, 60.17s Vitest duration; exit 0. Additional packed `installed-schema-cohort.test.ts`: **one passed**, validating the strengthened canonical-byte capture against the same installed configuration module. This extra test is reported separately, not added to the 112-test matrix count.

The positive cohort reports **43 passing scenarios, 43 sessions, 731 wire observations, 44 setup observations, 21 native Read observations**. Canonical requirement inventory remains 129 IDs with 103 uncovered IDs. The runner also reports 91 explicit `harness-error/not-implemented` catalog scenarios lacking executable plans; these are missing coverage, not N/A or passed obligations. Counts of scenarios, mapped requirements and uncovered IDs are distinct. `claimEligible:false` remains present in target readiness, report and result.

Exact observed identities (also in `summary.json` and raw `packed/progress.json`):

- Installed payload: `cac59fd75a1e406211f4c39fe77e70ed1cfb9c1c4e5d625e1f93e0bc015511c5`.
- Installed contracts: `48667497d6e13064316d0364cf8f49dfc247ab9189306985a357d7b558333dfe`.
- Full served configuration: `1720b63b49a7be3299d4b2865de4db9fcbe47560f217a32eea6397db7ea54ca5`.
- Policy: `dbe8dbd5430f26e19e754d38167cd8084a32d702c695550e2ee0063dc2912430`.
- Authentication binding: `66941efc779f4f2431d65fdd5b2387b2af67bd7c1559e09dac032f82259ea16e`; no credentials printed.
- Fixture: `85cab19b9561baac2b9011ff618cd3a12b440d1ae771e1c6a5cfab1834987402`.
- Exact route/request configuration projection: `bdffd6f957401222f6e5ffc928c7769d23976e8a8a8d5eb16dfa1bb7ce836a07`; this does not hash schema/descriptor bodies.
- Protocol tarball SHA-256: `2943f90cd4d0c5ac973dc84c75c80b6b0d944e688b0210990438a482a0a61a8c`.
- Server tarball SHA-256: `c12831b6b5cc1e62e792b88f6468262db34408393ebdb81bb766ebae12e7b12f`.

Scope remains `https://reference.invalid/acme/`, private Read+Update profile, History false. Catalog/manifest/fixture digests and all observed identities are preserved in the summary. The payload digest changed from d907 while contract/configuration identities remain unchanged; this is a fresh run, not copied observations.

## Observability and cleanup

Every stage reached an explicit exit; no timeout, interrupt or forced termination. External runners record current PID/start time, stage logs and exits, enforce per-stage deadlines and include process-group cleanup on failure. Progress and active child processes were inspected during full and packed runs. The packed child path was confirmed to be the installed tarball entry. Final process snapshot `processes-final.txt` showed no validation-owned test workers or installed-reference children remaining. Source status remained clean in the validation and original checkouts.

`run.py`, `pack-executor.py` and `summarize.py` are retained outside source alongside per-stage results. No full-profile claim follows: these are source integration checks plus the existing partial RU installed cohort, not an installed Transactional qualification. Snapshot, SSE, cold Transactional discovery and all missing contract/policy/control variants remain separate work. No public admission changed.
