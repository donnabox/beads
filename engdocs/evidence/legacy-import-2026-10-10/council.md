# Stage 1 review council

The operator requested Codex and Claude and explicitly excluded Gemini. Native
Codex delegation and the actual external `claude -p` CLI reviewed the same core
prompt against integration base `76fde9c9cb0865d430bf12d76fc473490f51128a`.
The initial tree changed during review; Claude pinned its findings to `0087b490f`.
The full original Claude findings are in [claude-initial-review.txt](claude-initial-review.txt).
The [bounded repair review](claude-repair-review.txt) targets immutable repair
commit `ba235baae` and reports **no confirmed blocker**. Claude ran no tests in
that follow-up; the qualification receipt records author-run checks.

## High-confidence findings

Both seats found compatibility/preservation risks in the original implementation:
strict JSON field aliases and surrogate validation, native label-set ordering,
legacy headers, cycle refusal classification, and external/cross-prefix routing.
Repairs preserve exact admitted data or refuse atomically. Native review also
added fingerprints for labels, comments, child counters and writer fencing.

## Single-reviewer findings and dispositions

| Finding | Reviewer | Disposition |
| --- | --- | --- |
| New creates/dependencies after import use the wrong prefix | Claude, High | Fixed: require one nonempty source prefix, adopt it in the fenced import transaction. Both-mode tests create new Issues and add dependencies in both directions. Fingerprints now include `issue_prefix`; dry-run and faults restore it. |
| Capability guard fails on `legacyImport` | Claude, High | Fixed the two-way guard; exact installed qualification includes that existing test. |
| Deadline classification hides uncertain commit | Native, High (repair review) | Fixed ordering: `ErrOutcomeUnknown` wins over cancellation. Four joined-error cases assert code, exit and no automatic retry. Native independently confirmed repair. |
| 16 MiB input does not guarantee persisted data fits 16 MiB read budget | Claude, Medium | Retain independent engine admission bound rather than invent an inaccurate preflight estimate. Docs call these rejection ceilings, explain staging/rollback, and the final error identifies the stored representation. Both-mode 9 MiB input test proves a late read-budget refusal leaves the complete fingerprint unchanged. |
| Historical header accepts only one schema and named provenance | Claude, Medium | Intentional supported-subset refusal. Exact schema/provenance allowlist is documented; accepting unknown future schemas could discard unsupported data. Classic importer remains unchanged. |
| Bond/formula lineage parses but cannot persist | Claude, Medium | Explicit preparation refusal for nonempty lineage with storage tests. Historical aliases remain parser-compatible so refusal identifies data rather than treating valid old spelling as an unknown field. |
| Size limits report inconsistent codes | Claude, Medium | Fixed: input, resource and current-read ceilings use `capability_unavailable`/5. Parser sentinel has no graphstore dependency. CLI input-limit and both-mode resource/read-limit tests cover it. |
| Field coverage limited to two Issues | Claude, Medium | Added both-mode exact hydration/retained validation for assignment, schedules, close attribution, pinned/template, compaction, molecule/work/wisp classification, gate and event families. Explicit storage-class markers refuse because the native writer normalizes them away. |
| Empty source succeeds silently | Claude, Low | Empty/header-only input explicitly refuses, including installed CLI. |
| Version fence comment omits import | Claude, Low | Updated caller explanation. |
| Bulk import uses interactive deadline | Claude, Low | Import uses five minutes through a shared store-budget helper; cancellation has an explicit refusal and uncertain commit retains exit6. |
| Untested CLI selectors/flags | Claude, Low | Added multiple positional files, input+positional, global flag and oversize input branches to both installed modes. |
| BUILD ordering | Claude, Low | Sorted changed src/dependency lists; Bazel unavailable locally, CI sync remains required. |
| History sentence and graph capability names | Claude, Low | Retained descriptive history disposition in this bounded preview contract. Graph import/export remain false; protocol amendment belongs to Janet. No graph-format reader or exporter is implemented. |
| Repeated Memory coordination updates | Claude, Low | Existing create writer retains its fence; no performance-only writer bypass. Added coordination fault coverage. Bounded batch limits apply. |

## Suggested updates and disagreements

All confirmed High blockers were repaired and both seats confirmed disposition.
Claude's residual non-blocking items are recorded in its full repair receipt:
unusual source prefix grammar and defensive unreachable SQL-error classification
remain follow-up hygiene; comment grammar and depth-refusal documentation were
clarified. The existing Dependency durable workflow and close-conflict tests passed in
embedded mode ([receipt](dependency-regression.jsonl)); optional server variants
skipped because the already-qualified import server had been cleaned up. The
95-case required import qualification separately exercised post-import
dependencies on both engines without skips. The conservative header policy and
independent persisted-byte check deliberately differ from Claude's suggested
permissive-header/lowered-input alternatives; supported/refused data and atomic
rollback are explicit. Full release acceptance still needs repository gates,
exact-head CI and Janet integration under Donna's cutline decision.
