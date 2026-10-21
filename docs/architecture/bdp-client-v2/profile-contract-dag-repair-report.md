# PR87 prerequisite DAG repair

Frozen source: `1f2d75fbaa67f631e72bd79384c829d7588b7277`, direct successor of reviewed `28b5da2092e57cac53fd67fab38ab9e95b64df28`.

Checkout: `/Users/dbox/repos/wt/vickie/vickie-profile-contract-dag-repair`; branch `codex/vickie-profile-contract-dag-repair`; origin verified `https://github.com/gastownhall/bdp.git`. Clean after commit. Prior council checkout remains frozen. No outward writes.

## Repair

The native review's valid shared prerequisite DAG caused exponential recursive revisits. The planner now evaluates prerequisite readiness once per scenario in canonical catalog order, reusing resolved boolean readiness. This uses the existing strict earlier-than prerequisite rule. Missing, retired, unready, later, and cyclic dependencies remain blocked; carrier precedence and inspectable blocker lists are unchanged. Dependency evaluation is linear in catalog rows and prerequisite edges; no recursion or new planning surface.

Only `packages/conformance/test-support/profile-contract-plan.ts` and `packages/conformance/src/profile-contract-plan.test.ts` changed. No runtime, spec, catalog, admission, evidence artifact, constant, or package export change.

## Deterministic regression and validation

New regression uses every canonical Read obligation from the actual loader and a valid manifest in which each row depends on its two predecessors. It counts actual requirement Map lookups against a generous linear budget and fails fast if exceeded, without a wall-clock threshold or production instrumentation. It also requires every row ready-unexecuted, no blocked prerequisites, every row unobserved, and claimEligible false.

- RED on unchanged 28b planner: only the new test fails with `shared prerequisite evaluation exceeded linear lookup budget`; existing 53 planner tests pass (`profile-contract-dag-red.log`).
- GREEN: 54 planner tests pass (`profile-contract-dag-focused.log`).
- Full conformance plus bdptest Read matrix: 817 tests / 25 files pass in 15.47 seconds (`profile-contract-dag-conformance.log`, structured `.json`). Existing missing, unready, cycle, retired and carrier-precedence regressions pass.
- Frozen dependency install, workspace build, typecheck, dependency boundaries, focused Biome check, and git diff check pass. Logs: `profile-contract-dag-{install,build,typecheck,boundaries,format}.log`.
- Node24.16.0 and pnpm11.20.0 pinned. No timed-out run or outstanding test process. No full workspace suite, evidence generation/verification, or RU/TX served qualification claimed for this two-file repair.

## Next bounded slice

After review closure, implement actual cumulative RU v2 manifest/fixture and profile-specific artifact admission preserving existing Read admission, then execute against the observed installed P1 configuration. Do not reuse Read v1 as claimed RU readiness or bind provisional P1 evidence. No further planner features proposed.
