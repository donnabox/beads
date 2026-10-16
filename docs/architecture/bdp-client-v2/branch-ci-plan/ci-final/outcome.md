# Exact branch-push CI outcome

**Failed:** run38858242028/job116834421250, attempt1, event push on codex/vickie-sequence-depth-retained. One failing test; no retry or repository mutation.

The log explicitly records checkout77b4cf980ec91f66c52a554c3dd8c9de5aee859f. Its API/local Git commit object gives tree6a8af0f4b40216ce9c0b3f26957d293caf6f6627 and sole parent11742645bc591b3e5affa9fcc2f36c551c6fa803. The only parent delta is .github/workflows/ci.yml. Product bytes are unchanged from117. This tree attribution derives from the exact logged commit and Git objects; it is not an independently measured runner byte inventory or a PR synthetic checkout.

- Tests:9348 passed,1 failed,1 skipped (9350);184 files passed,1 failed (185). Vitest duration787.69s.
- Failure: transaction-service.test.ts2061, “nonexecuting sequence refusal observation > keeps excluded sequence responses unobserved and without refusal context”: expected500, actual406. The source's last table row combines depth40 input with Accept:text/plain. This report records the observation; it does not authorize a repair or infer all other cases will remain green after one.
- Skipped test: scripts/demo-read.test.mjs, “runnable Read walkthrough > matches real bd and preserves the established seed command recipe”.
- First18 steps succeeded, including format/lint/boundaries/typecheck/build. Format reported1warning; lint947warnings/76infos. Fullsuite step failed with exit1.
- Preflight, smoke and Read-cohort evidence gate were skipped. No gate success or Read qualification is claimed. Post setup-go/setup-node skipped; pnpm/checkout cleanup and Complete job succeeded. All27 step outcomes and API timings are retained in inspection.json.
- Job05:54:28Z–06:11:07Z on2026-10-16:999s,201s before the original1200s job limit. It failed an assertion, not the job timeout. Test step05:57:55Z–06:11:04Z; its789s API interval differs from787.69s Vitest duration.

Complete raw log:3,506,447bytes, SHA256 b35426a229dc984dcd0720b27c119a57bff5422d1d819cfd341fbb2dbf07b29b.
Run metadata:f403b10c09345f3d02ec204bd215a224c5404b1292321b11ba8298256c8356b0.
Job metadata:87d2dd3ffd905f4171b5f52cfd15d84c86d30d6b908a1cadbe9e12caa137b430.
Commit API:28634c5d820500830f9cf61ce0da4c660921ae8f6a1ff2cf40450cae67e0ecf9.

Read-only API capture and exact local Git object inspection only. No main query/fetch, test execution, rerun, workflow dispatch, push or merge. Earlier failures and local23-source/40-installed evidence retain their own attribution.
