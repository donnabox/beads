# Discovery limits and C3 repair integration

Clean ordinary merge `9b95df5dd866dbaf84150404c4db86ac8054b724`, tree `d6749c7da26a5eeb2debce8336af91acc279249a`, parents in order:

1. Limits `fd07bf7ac177d5e53cabcd7666d9f2c77fba6e60`.
2. C3 external repair `a53f4bbf573c64fe900de373b3fef75170002f7c`.

Checkout `/Users/dbox/repos/wt/vickie/vickie-discovery-limits-integration`, branch `codex/vickie-discovery-limits-integration`. Both original frozen checkouts preserved. Git merged the two shared test files automatically without conflicts; no authored resolution or additional source edit. Commit carries the Agent-Signature trailer.

## Scope and preservation

Only four files differ from fd07: transaction-discovery.test.ts, transaction-read.test.ts, transaction-service.test.ts and schema-graph-policy.json. Stable patch ID `a04ae8df10adbba280ea8e5a8507cb9b42425d04` is identical for3f→a53 and fd07→this integration. This preserves the exact reviewed route/policy repair alongside the limits assertions. No production, manifest, schema, public API, package configuration, store or serving change relative to fd07.

## Executed validation

Pinned Node24.16.0/pnpm11.20.0. All stages completed exit0; no failed, interrupted or retried stage.

- Frozen-lockfile install; clean; prebuild typecheck; build: passed.
- Combined affected seven-file cohort: **225 tests passed**, including the schema-interpretation policy tripwire, service/read/HTTP/discovery tests, protocol parser tests and installed-contract tests. Stage3.572s; exact commands and JSON at `exits.json` and `affected.json`.
- Dependency boundaries, full format/lint and existing Read evidence verification: passed.
- `git diff --check`: passed.
- Fresh protocol/server packs from this clean source: passed; no fresh installed cohort was run.

This is bounded integration validation, not a full workspace or complete profile qualification run.

## Exact package identity and historical installed evidence

Fresh archives have identical paths and file bytes to retained fd07 archives, including package manifests: protocol63 files, server246 files. The archive bytes themselves also match:

- Protocol SHA256 `4683cb6c6333c332a4bcb7efc3e28f32d1d8d06a7264913f65c31d3a826d4227`.
- Server SHA256 `2b25f7d284f47e7a9f7c7be1105cf8565b067d181cb7bdf19eaa3ee4b87a6d50`.

`package-identity.json` retains every member hash/length and both archive bindings. The unchanged payload is fd07's `9000444c5d2036133eb8d8d870a760724da49f980d11ca2ec52925336211253a`.

Actual installed proof remains explicitly bound to **fd07** in `../transaction-discovery-limits-implementation/report.md` and `summary.json`: TX12 exchanges/one child and two targeted RU observations. No run is relabeled to this merge commit, and no additional installed execution is claimed. Byte equality justifies avoiding a redundant installed run for this test/policy-only merge.

## Cleanup and publication boundary

No owned child remains: final process inspection found only the inspection shell/rg, all bounded validation supervisors had exited. Working tree clean; no shared database touched. `source.json`/`summary.json` retain exact source, parent, file, test and package bindings.

Guarded branch push used `L/guarded-action.py`, checked remote Vickie session `2026-10-08T20:26:35Z-81e5`, and remote branch readback equals `9b95df5dd866dbaf84150404c4db86ac8054b724`. No PR creation/update or main merge. Root retains publication/readiness decisions; fd07 external review was still active when assigned. Discovery remains internally assembled and publicly refused; no profile claim changes.
