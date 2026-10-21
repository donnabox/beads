# Retained-retry diagnostic review closure

Frozen clean **c958e6416f5e9d406abbdacd17f15cecdf789766**, tree`c0fa3c30092476ea56fe21ca9f2da6a1c53e5ab1`, direct ordinary parent`99e28c5155afa31a55868c219f7771fa46bcef1d`. Checkout `/Users/dbox/repos/wt/vickie/vickie-retry-diagnostic-closure`, branch`codex/vickie-retry-diagnostic-closure`. One file only: `apps/bdptest/src/read-update-matrix.test.ts` (+65/-13). Original99e and integration6c23 remain unchanged. No outward writes.

Root accepted the five low findings from external review; the recovered final was later matched by its normally completed exit0 review. This patch closes only those diagnostics issues:

1. Harness error fixture includes a secret message; exact expected output includes only `harnessCategory`.
2. Projection includes bounded writeState, preferring transport state then harness state. Tests independently exercise transport precedence and harness-only fallback; no raw error messages cross the boundary.
3. Context includes observed client-exchange count and returned scenario IDs. A missing selected row omits scenarioCleanupFailed instead of falsely reportingfalse. The empty observation/result test pins this shape.
4. Deep optional route/status reads tolerate missing/null observation/route. Five malformed last-observation shapes execute the helper and verify no diagnostic throw and no invented route/status. This is deliberately not a generalized arbitrary-value or error framework.
5. The two fixed fixture credentials now use the repository's `<redacted>` marker. Existing exact-shape secret exclusion checks continue to cover header/body/assertion/error fields and credentials embedded in URLs.

The helper test uses its literal catalog scenario ID rather than new non-null index assertions. This removes the two warnings added in99e; lint185warnings/4infos remains an exit0 result, not warning-free qualification.

## Verification

Pinned Node24.16.0/pnpm11.20.0, frozen install exit0. Typecheck, build, touched-file Biome format/lint and git diff check pass. Retained-retry real emitted-installed controls/positive plus diagnostic unit test: **23passed,0failed,169unselected**,13.21s Vitest duration (15.17s process supervision). These ran on the final precommit bytes; `source.json` pins their file hash and frozen commit. No later source change was made. No fresh pack, full workspace or CI rerun was performed or claimed.

The original retained-retry control callsite block is byte-identical to99e, independently audited: all expected1 checks, intended failed oracle IDs and30000ms deadline remain unchanged. All non-test files/runtime/manifests/fixtures/coverage are identical to99e. Public admission/whole flags remain unchanged. This does not resolve or reclassify the earlier b604 CI failure; it makes subsequent failures more informative.

All owned commands finished and no owned installed child remains. No caffeinate was used. `summary.json` records stage commands/exits/timings, and `source.json` records exact source/tree/parent/file digest and preservation checks. Root owns native closure, integration and outward actions.
