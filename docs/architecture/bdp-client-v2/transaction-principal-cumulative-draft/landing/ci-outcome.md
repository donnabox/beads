# PR136 exact-head CI audit

CI run38817728355 / job116706490120 completed **success** on candidate 1f0cb9455361cef2cc793558793e4a190be6f4fe. All27 recorded steps completed successfully, including downstream preflight, smoke and Read evidence gates. This is CI evidence only; it is not a council verdict or merge/publication authorization.

The actual log checkout is synthetic f418b8c5412983d50a8078ce6eb332a172ab4956. Its fetched Git object has tree 22176e75848cfd6f0a54f799aeb1e62998206edb and ordered parents [5d7fa3f049fa0b4db3aa8dae20e20d8e47fb3854, 1f0cb9455361cef2cc793558793e4a190be6f4fe]. The candidate has sole parent 5d7fa3f049fa0b4db3aa8dae20e20d8e47fb3854; both candidate and synthetic tree are identical. I verified exact four-file scope and all four frozen source hashes, with the local candidate still clean. Terminal run/PR metadata binds the same head/base.

Full suite: **9315 passed / 1 skipped / 181 files passed**,837.57s reported duration (job test step838s). No failing test summary or GitHub error annotation appears in the full raw log. Preflight: **7 passed / 1 file**,1.47s. Smoke logs confirm installed help/version/diagnostics and clean lifecycle, public Read admission/fail-closed profile checks, and private installed-reference observation with claimEligible:false. Read gate actually emits evidence.verified with89rows; it did not merely take the exit2 allowance.

New principal registrations are present and passed:
- 2026-10-15T22:36:01.8271176Z  ✓ packages/conformance/src/transactional-cumulative.test.ts > pins the partial principal namespace row, full parent and exact credential roles 22ms
- 2026-10-15T22:49:25.3617239Z  ✓ apps/bdptest/src/transactional-matrix.test.ts > observes authenticated principal isolation and rotation through selected cumulative HTTP 1302ms
- 2026-10-15T22:49:26.6098510Z  ✓ apps/bdptest/src/transactional-matrix.test.ts > discriminates principal isolation against a separately bound collapsed namespace 1255ms

Job started 2026-10-15T22:31:50Z, completed 2026-10-15T22:49:42Z: **1072s**, leaving 128s against the existing20-minute job limit. Setup through test start took218s; Go bd installation alone took172s. These are this run's observed timings, not a future throughput/headroom guarantee.

Format reported one oversized Read+Update JSON warning. Lint reported897warnings/56infos. Post-job output also records punycode deprecation and the warning about older action Node20 targets being forced onto Node24. Success is not warning-free. The full raw log preserves all diagnostics.

The one skipped test:
- 2026-10-15T22:43:16.3564689Z  ↓ scripts/demo-read.test.mjs > runnable Read walkthrough > matches real bd and preserves the established seed command recipe

| Step | Name | Conclusion | Metadata seconds |
|---|---|---|---|
| 1 | Set up job | success | 3 |
| 2 | Run actions/checkout@11d5960a326750d5838078e36cf38b85af677262 | success | 3 |
| 3 | Run pnpm/action-setup@b906affcce14559ad1aafd4ab0e942779e9f58b1 | success | 1 |
| 4 | Run actions/setup-node@49933ea5288caeca8642d1e84afbd3f7d6820020 | success | 8 |
| 5 | Run actions/setup-go@924ae3a1cded613372ab5595356fb5720e22ba16 | success | 10 |
| 6 | Run test "$(node --version)" = "v24.16.0" | success | 0 |
| 7 | Run test "$(pnpm --version)" = "11.20.0" | success | 1 |
| 8 | Run sudo apt-get update | success | 4 |
| 9 | Run sudo apt-get install --yes libicu-dev | success | 1 |
| 10 | Run mkdir -p "$RUNNER_TEMP/bdp-tools" | success | 0 |
| 11 | Run go install github.com/steveyegge/beads/cmd/bd@v1.0.5 | success | 172 |
| 12 | Verify the Go-installed source-build bd identity | success | 0 |
| 13 | Run pnpm install --frozen-lockfile | success | 2 |
| 14 | Run pnpm format:check | success | 1 |
| 15 | Run pnpm lint | success | 3 |
| 16 | Run pnpm check:boundaries | success | 1 |
| 17 | Run pnpm typecheck | success | 4 |
| 18 | Run pnpm build | success | 3 |
| 19 | Run pnpm test | success | 838 |
| 20 | Run pnpm e2e:ready:preflight | success | 3 |
| 21 | Run pnpm smoke | success | 8 |
| 22 | Read-cohort evidence gate | success | 1 |
| 41 | Post Run actions/setup-go@924ae3a1cded613372ab5595356fb5720e22ba16 | success | 0 |
| 42 | Post Run actions/setup-node@49933ea5288caeca8642d1e84afbd3f7d6820020 | success | 0 |
| 43 | Post Run pnpm/action-setup@b906affcce14559ad1aafd4ab0e942779e9f58b1 | success | 0 |
| 44 | Post Run actions/checkout@11d5960a326750d5838078e36cf38b85af677262 | success | 0 |
| 45 | Complete job | success | 0 |

Artifacts:
- ci-full.log: 7f842941de93a1749d1915ddc965f2c0e549e4009e9e92cf0e7378b364f41a89,3429454bytes, downloaded through the complete raw jobs/116706490120/logs API.
- ci-terminal.json: 2e54d6a82c687fcec47d82222134669d9ebb9b19b9a662683684f653c7416459.
- ci-source-identity.json: 6e3393adcc0f1ad2cf6d405c8fbc728377173056bd5fbf48da852a2141872280.
- Source binding: 0dbae7cb6139263cbfd3931e80d8dec2c628d7e982deecdd6c8242573a35c6e2.

No retries, cancellation, source edits, repository mutation, PR readiness change or merge occurred during this audit. Remote runner cleanup is reported by Actions; local installed cleanup evidence remains separately attributed. Earlier local format failure remains preserved.
