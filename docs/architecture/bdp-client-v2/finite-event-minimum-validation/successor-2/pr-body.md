Initial finite Scope Event GET/HEAD reads returned the private minimum-position 500 for an unrecognized minimum token. With a complete tuple matching the current epoch and authorization view, they now wait within the existing bound and return contextual 503 `catch-up-timeout` when the minimum remains unproved.

The change applies to canonical initial JSON Event reads with no `after` or `Last-Event-ID`. It captures one tuple after waiter preflight and checks current context before retained Event scans. Known minima, conditional ordering, cancellation and drain behavior are preserved. Resource Event views, continuations, SSE, incompatible media and foreign contexts retain their documented private gaps. This does not advertise complete Update/Transactional support or activate public admission.

Stacked on PR151 at `b76c3b7df77ea9b33611bc9027db651833e20abb`; frozen PR150 remains unchanged. Six files changed. Commit `f6204255d53ef65d4aea56ff5ccc86d72e8cc7b0` preserves all 679 tested source files exactly.

Validation:

- All 19 fresh stages passed: typecheck, format, lint, clean build, selected source tests, package/archive checks and fresh external-consumer validation. Lint retains 19 warnings and one informational diagnostic.
- All 13 selected source callbacks passed, including five preflight controls and four existing regressions. New HTTP observation recorded 32 exchanges; historical source traffic is not represented as a measured combined total.
- Fresh installed consumer passed the new three-exchange GET200/GET503/HEAD503 case and the historical 11-exchange missing-context case. Both children closed cleanly; owned stores and all 326 sampled processes were absent at final audit.
- Native and actual Claude review completed; subsequent mechanical formatting and equivalent cleanup corrections have separate native/root dispositions. Earlier failed static attempts retain their original identities.
- Exact commit [CI38995336477](https://github.com/gastownhall/bdp/actions/runs/38995336477) passed all 27 steps: 9,375 tests across 186 files, one skip; seven preflight tests; static checks/build, package smoke and the historical Read evidence gate. Raw job log and exact head/tree/ancestry verification are retained in the evidence packet. No merge or release requested by this draft.

[Evidence and dispositions](https://github.com/donnabox/beads/tree/codex/vickie-bdp-client-design/docs/architecture/bdp-client-v2/finite-event-minimum-validation/successor-2) · [Sole implementation plan](https://github.com/donnabox/beads/blob/codex/vickie-bdp-client-design/docs/architecture/bdp-client-v2/design.md)
