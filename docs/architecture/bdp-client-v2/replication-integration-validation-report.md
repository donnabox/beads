# Independent replication integration validation

Exact frozen source **d907a1e1576c5795c17edb5f15528f47edfd7d6b**. Isolated checkout `/Users/dbox/repos/wt/vickie/vickie-replication-integration-validation`, branch `codex/vickie-replication-integration-validation`, GitHub origin. Source was clean before and after all runs; no source edits, commits, pushes or evidence relabeling. Root's original frozen checkout was not used for tests. Executed 2026-10-13 UTC using Node24.16.0/pnpm11.20.0.

## Results

Full suite: **8,398 passed, zero failed, one optional skip, 146 files**, exit0, 100.78 seconds wall time (Vitest99.73s). One skip is reported honestly; not described as complete no-skip coverage. Structured `fulltests.json` and verbose `fulltests.log` retained. MaxWorkers2. Clean build preceded all emitted/process tests.

| Check | Exit | Wall time |
| --- | --- | --- |
| install | 0 | 0.51s |
| typecheck | 0 | 1.39s |
| build | 0 | 0.81s |
| fulltests | 0 | 100.78s |
| boundaries | 0 | 0.63s |
| format | 0 | 1.41s |
| lint | 0 | 0.45s |
| smoke | 0 | 4.07s |
| smoke-development | 0 | 0.63s |
| preflight | 0 | 1.00s |
| evidence | 0 | 0.30s |

Lint exited0 with **45 warnings and one info**, no fixes applied. Both executable and development smokes passed. Preflight passed7 tests. The historical Read evidence verifier passed89 rows while retaining its own source `e5f7d0c43c151426350ceededdc77fe7e372db9b` / evidence commit `dbaf63a24219e1ea3f80394746853598b0d4464a`; that artifact is not an observation of d907 and was not resealed.

## Fresh packed installed RU run

Copied/adapted the prior executor outside source to `pack-executor.py`; exact-head/clean checks added, fixed new checkout and distinct pack/install/log paths. Packed protocol/server from the clean emitted build, installed both fresh tarballs, then ran actual installed entry `/Users/dbox/repos/wt/vickie/bdp-integration-logs/replication-integration-validation/packed/install/node_modules/@bdp/server/dist/installed-reference-process.js`. It does not fall back to a workspace server entry. Suite: **112 tests passed, zero failed, four files**, exit0 (Vitest54.01s). JSON/verbose outputs are `packed/matrix.json` and `packed/matrix.log`; stage exit/time files retained.

Positive cumulative report: **43 passing scenarios, 43 actual sessions, 731 wire observations, 44 setup observations, 21 native Read observations**. Canonical required inventory129; uncovered-ID accounting103; report also retains91 `harness-error/not-implemented` catalog scenarios with reason `catalog scenario has no executable plan`. These are explicit missing coverage, not failed implemented tests and not N/A. Counts are not interchangeable: plans, catalog selections and uncovered-obligation mapping differ. All counterfactual detection tests completed successfully. `claimEligible:false` in report, target readiness and result.

Actual source/configuration/hashes (full captured values in `summary.json` and `packed/progress.json`):

- Scope: `https://reference.invalid/acme/`; private served profile `read-update`; History false.
- Installed payload: `f9dfc6e2fab5aa27906e15eddab75d48d4a6879a6dc884ac4059ac0941c7d7b6`.
- Installed contract closure: `48667497d6e13064316d0364cf8f49dfc247ab9189306985a357d7b558333dfe`.
- Configuration identity: `1720b63b49a7be3299d4b2865de4db9fcbe47560f217a32eea6397db7ea54ca5`.
- Policy: `dbe8dbd5430f26e19e754d38167cd8084a32d702c695550e2ee0063dc2912430`.
- Authentication binding digest: `66941efc779f4f2431d65fdd5b2387b2af67bd7c1559e09dac032f82259ea16e` (no credentials recorded).
- Fixture: `85cab19b9561baac2b9011ff618cd3a12b440d1ae771e1c6a5cfab1834987402`.
- Catalog/manifest/fixture artifact bindings and exact configuration projection are preserved in summary/report; do not confuse projection digest with full served configuration identity.
- protocol.tgz: `2943f90cd4d0c5ac973dc84c75c80b6b0d944e688b0210990438a482a0a61a8c`
- server.tgz: `d9b0b84c22275591089d387091d1da2b92bf2248a32571b398588f66f931445e`

## Cleanup and interpretation

Every command reached an explicit exit; no timeout, interruption or forced termination occurred. External runners use bounded per-stage deadlines and process-group cleanup for failure. Final process inspection found no validation-owned test workers, installed-reference children or runner processes remaining. Full process snapshot retained at `processes-final.txt`. Both isolated candidate and original source remain clean.

This validates integration tests and the existing **partial RU installed cohort**; it is not a packed Transactional cohort or whole-profile certification. Snapshot bootstrap remains planned; SSE, cold Transactional discovery/Directory, unresolved minimum/control surfaces, set/sequence admission variants, mutable authorization/erasure behavior and required harness coverage remain open as recorded in the burndown. Finite feed/Events and M32/128-Event tests are included in the full source suite; no new normative or public capability claim follows. Root's reviewed integration and remaining council findings determine publication readiness.
