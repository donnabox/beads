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
it. The failed [raw run](initial-tests.jsonl) and [initial diff](initial.diff.gz) are
retained. Repaired [source manifest](source-manifest.json) and
[raw focused run](focused-tests.jsonl) bind the final test source:3passing records,
zero failures/skips, exit0. Explicit no-engine settings disable shared-server and
Dolt startup; no producer or database runtime ran. The exact command appears in
[quality receipt](quality-receipt.json). Pinned v2.10.1 affected-package lint
([result](lint.log)), Go1.26.5, reports0issues with --new-from-rev=HEAD against
preparation1a9f3435. Ambient lint2.9.0 was inspected but not used as proof.

[Native final review](native-review.txt) has no actionable finding;
[dispositions](council.md) cover every initial external finding. Actual initial
review completed exit0/empty stderr; repaired-source review completed; its one Low clean/seeded cache comparison
suggestion is implemented; final tiny confirmation completed exit0/empty
stderr with [no actionable finding](close-external-review.txt): safe mode disables customizations/hooks; restricted mode, Read/Grep/Glob
only, strict empty MCP, skills and persistence disabled. No model override or
third provider. Initial/final review hashes and source scopes stay separate;
current-source final confirmation is complete for this bounded test-only unit. Full Bazel/hosted/installed
checks, genuine corpus and full activated-source review are not implied.

Janet still owns engine handback and shared Git audit. Only the private recovery
clone is used; no unsafe linked Git worktree, core/staging/production store,
migration, default, main or release ref is changed. Existing source-specific
runtime evidence remains historical. Hooks bypassed for source/evidence commits
to avoid ambient bd/broad engines; required quality gates remain open.

Raw patch context and runner stderr contain intentional whitespace. These
proofs are losslessly gzip-archived rather than editing their recorded bytes;
[archive hashes](raw-archives.json) verify exact decompression. The initial
staged whitespace check correctly flagged only those raw artifacts; the final
source/prose/evidence delta is checked after archiving. This changes no reviewed
Go source or test result.
