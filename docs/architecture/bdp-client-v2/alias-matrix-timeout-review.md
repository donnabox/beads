Root publication note: reviewed and pushed21654be to PR83 after this worker report. CI38109646998 is in progress. The worker’s “No push” below describes its handoff boundary.

# Development matrix aggregate timeout diagnosis

Frozen fix: `21654be3cae51d50e75e817ef8ac57735b8ee4ea`, direct child of `59915ea916a210f5ce36611eb6d799f7c3124056`. Checkout `/Users/dbox/repos/wt/vickie/bdp-alias-matrix-timeout`, branch `codex/vickie-bdp-alias-matrix-timeout`. No push.

CI run 38109034182 failed only the complete diagnostic matrix at 5,007 ms with Vitest's default 5,000-ms outer timeout. Other negative-control tests passed; the CI log reports no inner HTTP, action, IPC or cleanup failure. This supports an insufficient aggregate budget but does not conclusively establish the original runner's scheduling cause.

Original-budget local matrix measurements: 1,090 ms, 1,154 ms, 1,131 ms. The latter two runs included the full six-test matrix file plus the three standalone disconnect/recovery tests, all passing. A temporary 25-ms delay after each completed actual action reproduced the exact outer timeout. With only the aggregate budget changed to 30 seconds, the same delayed workload passed all 23 scenarios / 221 actions / 705 assertions in 7,160 ms. This is a controlled cumulative-budget demonstration, not a reproduction of CI load.

Removed the delay entirely. Three final fixed-budget runs each passed all nine focused tests; full-matrix durations were 1,321 ms, 1,445 ms and 1,187 ms. Assertion-failure cleanup, abort cleanup, digest-change negative control, disconnect/drain/SIGKILL and normal replay remained exercised. Typecheck, whole-repository formatting and committed evidence:verify passed. Initial build passed. No redundant full suite run; exact-head full CI remains required after publication.

The final diff changes one test only: a 30,000-ms timeout plus two explanatory comment lines. No global timeout, action/HTTP/IPC/cleanup deadline, assertion, runtime, manifest, policy or evidence changes. Read seal ancestry and constant remain valid. No hooks bypassed. Checkout clean; no test-owned child remains. The deliberate old-budget timeout left one disposable fixture directory; it was identified by exact run birth timestamp and scenario inventory, then removed. Temporary instrumentation is absent.

Raw CI log: `alias-matrix-ci-failure.log`. Reproduction, repeat test JSON/logs, typecheck, formatting, evidence and cleanup logs: `alias-matrix-timeout-*` beside this report.
