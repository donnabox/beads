# Final PR83 Read qualification after council repair

Observed source: `e5f7d0c43c151426350ceededdc77fe7e372db9b`.
Evidence-only direct successor: `dbaf63a24219e1ea3f80394746853598b0d4464a`.
Clean checkout: `/Users/dbox/repos/wt/vickie/bdp-provenance-runtime-integration`, branch `codex/vickie-bdp-provenance-runtime-integration`.
Constant: `674982be2123b9b7aae630bf79c3f774122900cd`.

## Fresh observations

Actual generator passed in 203.18 seconds using Node24.16.0, pnpm11.20.0 and explicit Homebrew bd1.0.5. Its baseline verification, both in-process matrices and packaged observations all completed. No baseline recapture, shared workspace, Janet asset or changed pin/policy was used.

49 required definitions split into 38 packaged and 11 self-certifiable definitions. bdptest: 36 packaged plus 11 self-certified =47 passes; 2 capability-gated not-applicable. bdpbd:34 packaged plus8 self-certified =42 passes;7 not-applicable. Total89 applicable passes, no other outcome; existing uncovered variants remain explicit. Self-certified rows are not presented as packaged black-box evidence.

Full artifact SHA256: `674982be2123b9b7aae630bf79c3f774122900cd0e9a50a8daf967bd5a251132`.
Schema Read projection: `bd7da0fc3f8082e0bfb6f4969c04409f2a610e12408dbc6a0bd14b4b351127cc`;47 sealed definitions,46 reachable. No source/policy/schema changes during generation.

## Exact successor validation

- PASS build and evidence:verify; verifier identifies source/evidence pair above and89 rows.
- PASS full suite:7868 passed,1 skipped,117 files,58.60 seconds; maxWorkers2 and actual pinned bd lane required.
- PASS typecheck, dependency boundaries, format, lint.
- PASS installed executable smoke and development smoke.
- PASS readiness preflight:7 tests.
- PASS git diff --check, clean worktree. Evidence commit changes exactly artifact and matching constant, with no other file.
- No test-owned generator, Vitest or packaged-listener processes remained at final process inspection. No timeout/interruption occurred.

Logs are `/Users/dbox/repos/wt/vickie/bdp-integration-logs/pr83-final-*`; generator uses `pr83-final-reseal-generate.log/.exit`, full suite `pr83-final-full-tests.json/.log/.exit`. Individual checks have named .log/.exit files. Pre-generation build/type logs use `pr83-final-reseal-*-before`.

## Remaining boundary

No outward writes. Root owns independent review, PR publication, latest65 ancestry integration, and exact published-head CI. Later ancestry integration must retain this coherent pair and actual observed source relationship; no qualification of another implementation is implied. This Read run does not advertise Update or Transactional profiles or establish current Preview2 CLI parity.
