# Profile integration dispositions — 2026-10-13

## H1b ee947 closure supplement

Claude64976 exited0 with substantive static report; native closure accepted with315 focused tests and two installed probes. Both initial native P2s and enumerated external faults are closed. Root accepts a bounded successor before publication for the remaining maintainability/evidence/lifecycle issues. Preserve ee947 source and existing positive packed observations.

1. Preserve exact safe reasons and correct failure category throughout receipt evidence. Wire/parsed-response self-disagreement is a harness fault; an actual target malformed/different retained receipt is a target failure. Audit source schema invalid separately: missing validator is harness, but real invalid target schema must not become harness merely because a reviewer proposed that classification. Avoid importing runner classes into its helper and creating a cycle. Add discriminating tests for both categories and safe reason retention. Do not expose arbitrary server bodies/secrets as reasons.
2. Remove unused retained wire copy. Keep the original observed response digest as provenance; do not replace it with strict equality of subsequent response bytes (headers/current context may legitimately differ). Existing body agreement remains meaningful.
3. Add explicit lifecycle cleanup budget to raw session target, preserving legacy default and per-request5s. TX cleanup can use10s within the unchanged aggregate/reopen budget; document nested end-to-end budget, do not claim each phase separately gets10s. Test real close seam without a10s sleep and preserve abort/child drain.
4. Drop unreachable retainedSource authority extension and align prose to canonical mutation response authority. Dynamic receipt reads do not capture/regrant authority in this harness. Ordinary Resource capture remains valid and revokes reused binding authority.
5. Delete redundant late servedSession attachment; construction-time attachment remains authoritative including failure paths.
6. Response constant centralization is optional; current fail-closed guard is not a demonstrated defect. Do not widen public interfaces or add broad refactor. If a tiny private shared constant materially removes duplication, ensure exact bindings and tests remain unchanged; otherwise record deferred.
7. Preserve original failure context when cleanup also fails and assert AggregateError causes, not only top-level string. No swallowed cleanup failure.
8. Correct documentation: abort may concurrently initiate child close and request cancellation; pending settles before final cleanup resolves, but strict drain-before-signal ordering is only non-abort. No new artificial ordering claim.

This successor is harness/test/docs only, no Snapshot/SSE/runtime semantics. Exact source tests and fresh pack follow repair; prior evidence stays attached to ee947/a88.

## SSE plan revision

Root read revised4e000306d9b20dc9a47565a489a7dc47edbfeb80f5b26425ed12d6e672bef2dd in full. It incorporates council/root dispositions without new semantics or wider implementation authority. Accepted as the next G2a plan. Source authoring remains serialized behind accepted Snapshot HTTP integration; start=now plus Last-Event-ID remains explicitly deferred, erasure/Event-source/set/discovery gaps remain.

## RU key/media review

Native394d review accepts whole credit for the two exact selected clauses, no flag changes. Root checked actual normative3172–3178 and3432–3434 and agrees scope excludes the neighboring normalization/namespace/Accept laws. External30701 is still pending; no promotion until reviewed. Actual manifest hash ends7285d5d23aa..., as independently rehashed; prior conversational shorthand ending7285d5b23aa was inaccurate, not different source.

## Provisional integration

Ordinary a88b1b50a5fbdf6a05933dac22504cd0e05b7bfe merges f725 + ee947 +394d without conflict. Full8571 tests/153files/one optional skip passed, all pinned build/type/static/smokes/preflight/evidence passed. Fresh installed TX/RU worker running. This compatibility run does not close above review residues or imply readiness.

Ordinary2124adb31ff3406052231fa67502d83793bb649d adds frozen Snapshot HTTP18c2 to a88, no conflict. Independent native+Claude code review and provisional integration tests active. No publication yet. Frozen checkouts remain unchanged.
