# Independent review of editorial council repairs

Candidate: `e9361d77b862aa0e6ddfa19cf272225f62c32d35` in `bdp-editorial-council-repairs`.
Base: `2535b898`.

**No unresolved blocking finding in this bounded review.** No source edits performed.

## Preserved obligations and profile boundaries

- Compared repair changes with all 17 council dispositions and the base's earlier schema-name/citation repairs. The repairs restore local rule placement/force, actor surfaces, conditional erasure, HTTP ordering, references and clarification wording without changing wire schemas or runtime behavior.
- Across all six catalogs, every original ID, requiredProfile, kind and retires field is preserved. History increases from 21 to 30 cases by adding nine profile-scoped siblings; none retire prior cases. Every prior History source clause occurs in the resulting catalog, after only the intended dangling-pronoun replacement (`it` -> `changeContext`). The context-time title is narrowed to the original case's remaining scope; its atomic-time clause remains on its Transactional sibling.
- Read+Update discovery required/optional closed membership exactly matches unchanged `readUpdateDiscovery`; limit groups and retention members match unchanged `readUpdateAdvertisedLimits`. Existing Read `advertisedLimits` accepts broader capability groups; the repair honestly distinguishes schema acceptance from capability relevance rather than silently tightening it.
- Independent anchor scan: 144 anchors, 185 internal links, all 84 baseline anchors preserved, no missing targets or forward links from Read/Read+Update.

## Test quality

- Problem-table checks now compare each owning section independently against protocol/schema rows. Moving an end row between sections while preserving concatenated order no longer escapes the checks; RU tests no longer depend on reading the TX table merely to slice it away. Existing row/schema assertions remain.
- The new boundary test reads every catalog, validates source quotations/anchors, calculates actual containing profile, and rejects later-profile ownership plus normative citations to the navigation overview. This checks an artifact relationship independently of runtime implementation. It is structural coverage, not proof that every sentence has complete semantics.
- History count updates reflect additive metadata cases, not fabricated executable qualification. No History manifest or shipping admission is added.

## Independent execution

Pinned Node 24.16.0 / pnpm 11.20.0:

`vitest run packages/conformance packages/protocol --maxWorkers=2 --reporter=verbose --reporter=json`

**PASS: 42 files, 1,636 tests, 8.98 seconds; exit 0.**

Evidence in this directory:
- `repair-independent-tests.log`, `.json`, `.exit`
- `repair-independent-inventory.log`
- `repair-independent-anchors.log`

Checkout remained clean at exact reviewed HEAD. No full repository suite, shipping profile conformance, CLI parity or Preview 2 qualification claimed. This native follow-up does not expand Gemini's earlier explicitly partial council coverage. Existing Read limit-schema breadth remains a separate design question, not an editorial blocker created by this repair.
