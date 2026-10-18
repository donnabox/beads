# SDK JSON-clone omission guard

Adds only TestLegacyImportJSONCloneOmissionCoverage to the existing pure
freshness/error test file. Production code, active installed workflows, SDK Types
and public contracts are unchanged. The existing refusal test now reuses the
shared concrete fixture helper; no production refusal table is changed. The guard inventories exported fields tagged
`json:"-"` through nested legacy batch input, requires explicit classification,
and detects stale policies/unsupported shapes. A shared test fixture helper
connects every refused policy to executed preparation/refusal checks in this
guard and the existing behavioral test. The new control rejects source cache/counter sentinels and checks unchanged
caller input without constraining when derived state is recomputed.

The first canonical no-engine run had2passing tests/1failure/0skips: the new
assertion wrongly required an empty ContentHash. Preparation correctly computes
it. The failed [raw run](initial-tests.jsonl) and [initial diff](initial.diff) are
retained. Repaired [source manifest](source-manifest.json) and
[raw focused run](focused-tests.jsonl) bind the final test source:3passing records,
zero failures/skips, exit0. Explicit no-engine settings disable shared-server and
Dolt startup; no producer or database runtime ran. The exact command appears in
[runner output](focused-tests.stderr). Pinned v2.10.1 affected-package lint
([result](lint.log)), Go1.26.5, reports0issues with --new-from-rev=HEAD against
preparation1a9f3435. Ambient lint2.9.0 was inspected but not used as proof.

[Native final review](native-review.txt) has no actionable finding;
[dispositions](council.md) cover every initial external finding. Actual initial
review completed exit0/empty stderr; final repaired-source review is running: safe mode disables customizations/hooks; restricted mode, Read/Grep/Glob
only, strict empty MCP, skills and persistence disabled. No model override or
third provider. Initial/final review hashes and source scopes stay separate;
current-source final confirmation is still pending. Full Bazel/hosted/installed
checks, genuine corpus and full activated-source review are not implied.

Janet still owns engine handback and shared Git audit. Only the private recovery
clone is used; no unsafe linked Git worktree, core/staging/production store,
migration, default, main or release ref is changed. Existing source-specific
runtime evidence remains historical. Hooks bypassed for source/evidence commits
to avoid ambient bd/broad engines; required quality gates remain open.
