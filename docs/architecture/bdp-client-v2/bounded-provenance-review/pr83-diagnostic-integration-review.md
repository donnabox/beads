# Frozen PR83 diagnostic integration

Head: `65ff2f00419d7aa94d84d56a7127d172d894b5c4`, clean in `/Users/dbox/repos/wt/vickie/bdp-provenance-runtime-integration`, branch `codex/vickie-bdp-provenance-runtime-integration`.
The previous source freeze remains at branch `codex/vickie-bdp-provenance-runtime-frozen-55ed` (`55ed77e3035647f7ab32030978eba6b01a0dc993`).

History preserved ordinarily:
- Merge `7fd78ffeefd5442e47dbd78d17fef16cb9223332` has parents55ed77e and root PR64 repair1af8d4d3a11f6e90f6528fa78e54ec005c56f8bd. Includes the narrow shared formatter and Transactional regression additions.
- Follow-up65ff2f0 adds only the four built private-helper output filenames to the installed-package smoke inventory. There is no public export.

## Conflict resolutions

Two files conflicted: docs/specs/bdp.md and packages/server/test-support/schema-graph-policy.json. The spec retains the four-part/common revisionProvenance form from55ed, with every incoming c171→1af diagnostic/coherence clause transferred: document-qualified failures, properties-only UpdateWhere, metadata representation-byte scope, retained-upsert metadata convention and historical-digest proof limitation. Policy retains provenance and diagnostic review justifications and binds the combined schema bytes. No graph policy version changed.

The automatically merged Transactional catalog's strengthened owned-Link citation retained old `attribution` wording. Citation tests caught it; only that selectedText was migrated to the actual common `revisionProvenance` clause. IDs, profile tags, counts and requirement strength remain unchanged. All catalog ownership/citation checks pass.

## Runtime change

The reviewed Read+Update numeric diagnostic formatter is extracted unchanged into private member-numeric-diagnostics.ts and invoked by both prepareMemberExecution and evaluateTransaction. The Transactional bypass now applies the same document rules to create/update Bead/Link: create locations are document-relative; metadata always names its document; unapplied metadata update values omit a fabricated location; metadata omits Type/schemaLocation; callers cannot relabel properties as metadata or strip the metadata marker. Closed-field copying happens before the existing byte accounting. Transactional count16/byte65536 and request byte limits remain unchanged.

Ten new actual Transactional owner tests cover both documents across all four operations, late numeric failure after a valid staged prefix, skipped tail and exact unchanged resources/identities/groups/position, failed receipt validation, restart/replay and changed-literal conflict. The metadata byte test uses a shared long parent to reach exactly65536 diagnostic bytes while remaining below the existing request limit; two entries fit exactly, the third truncates, and one extra parent character prevents the second entry fitting. The count test retains16 marked entries and marks the17th omitted. Earlier RU regressions test injected caller mislabeling and stripping Type/schema fields through this same formatter.

## Validation

Pinned Node24.16/pnpm11.20, maxWorkers2. Build completed before tests. Combined targeted suite **518 pass /8files /7.50s** (TX owner, RU identity/executor/evaluator, graph policy, Transactional/RU catalogs, profile citation boundaries). A final owner rerun after replacing unknown-problem test accesses with actual protocol parsing passed **106 tests /1file /5.33s**. Build, final typecheck, boundaries, format, lint, development smoke and installed-package smoke pass. Clean git status and no evidence/capability-file delta from55ed.

The initial run caught test-only use of labels on update operations and a test input exceeding the existing request byte cap; tests were corrected without widening runtime limits. It also found the catalog excerpt drift described above. The first installed smoke caught the new private module missing from the exact packaging inventory; the four-file follow-up fixed that and smoke passed.

Committed evidence:verify still fails with the inherited `current catalog bytes drift for bdptest: re-seal required`. No seal, capability constant, observation counts or evidence metadata was modified. Full suite was deliberately not repeated pending root's final Read enforcement merge and actual fresh integrated qualification. No public profile claim or outward action.

Evidence logs: `pr83-diagnostic-*` in this directory. Root owns further reviewed source merges, independent review, Read generation and publication.
