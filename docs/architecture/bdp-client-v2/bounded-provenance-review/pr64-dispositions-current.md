# PR64 review dispositions — current checkpoint

This is a bounded review-disposition record, not a readiness or qualification claim. Source repairs are integrated locally at `2a4ede9b17b29c9686dbfd8ae4ee32b9a6adecac`. Original reviews cover `182f1fcf` → `b5a5b415`; their earlier seal verification does not qualify later repaired sources. Original outputs remain unchanged beside this report.

## Findings and dispositions

| Review finding | Disposition and evidence |
|---|---|
| Native P2: synthesized legacy metadata mutable | Closed by `0172c16fea5f29b51628d25d43f3b014cbe04a63`; independent seven-path freeze reproduction passed. `native-repair-addendum.md`. Source-pin follow-up `3bfe7d6` records graph interpretation unchanged. |
| Native P3 / Claude #3: patch limit definition | Closed by `0172c16`: combined operation count, per-path byte/depth bounds, existing set-member interpretation retained. |
| Native supplemental P2: owned-Link no-op/fold omits metadata | Closed by `c1712f2d26f889f723ffb79c01146c87883a65df`; compare both objects, apply each supplied delta, preserve omitted plane. `native-event-coherence-addendum.md`. |
| Claude #1: metadata diagnostic ambiguity | Closed in PR64 by `aeb19286269f922bc9f514bba39703c3cce67608`: optional closed `document` discriminator, omission means properties, metadata marked, document-relative pointers, complete diagnostic budget accounting. Native independently passed 574 tests / 8 files. **PR83 integration still needs its private transaction numeric formatter adapted and tested**; see below. |
| Claude #2: current Read metadata obligation not enforced by executable matrix | Repaired by `21522d4573fcc3afd815a8fe68f86ff5a8f6fbbe`: 25 manifest refs and controlled pages use current schemas; internal observer validates bounded parsed wire before compatibility projection. Five actual matrix omission controls isolate schema failure; legacy parsers still accept those bodies. Finite harness binding includes the new wrapper. Native independently passed 81 tests / 3 files; the unrelated required-check blocker on this commit was corrected by `3e7b6724fac41277c1d7d0e7220f854930d6af0d`. Fresh qualification is pending. |
| Claude #4: metadata selector silently missing | Repaired in `21522d4`: both candidate shapes and root membership include metadata, with Bead/Link collision, escaping, missing/null and immutability tests. Native source review found no further defect. |
| Claude #5: weakened Event catalog excerpts | Closed by `8f1d0596ff4c04358ca246f3f77603f8029ca3e7`: actual owned-Link alternative and full Link-update shape cited. Native coherence check: 158 tests / 2 files. |
| Claude #6: UpdateWhere/upsert coherence | Dispositioned by `8f1d059`: existing UpdateWhere targets properties only; newly committed upserts require metadata, retained pre-metadata groups remain accepted. No bulk metadata API or blanket schema tightening added. |
| Claude #7: historical digest cannot be proved from projected record alone | Explicit limitation retained by `8f1d059`: original representation is required; projected content alone is insufficient. This does **not** add an original-form accessor/discriminator or claim every SDK value can prove a historical digest. Native accepted the bounded clarification; broader proof capability is not silently claimed. |
| Claude #8: illustrative validation messages | Corrected in `aeb1928`: properties and metadata example messages match the respective evaluator messages; metadata example carries `document`. Illustrative `detail` remains illustrative, not captured execution evidence. |
| Claude #9/#10: duplicate projection / optional TypeScript metadata | Optional cleanup, not repaired or asserted closed. Legacy representation support remains intentional. No runtime defect beyond the separately fixed mutability issue established. |
| Claude #11: historical packet/README gating | `8f1d059` adds explicit historical/subordinate authority note to packet. README and remaining historical narrative were not comprehensively rewritten; optional editorial residue remains. |
| Claude #12: fixture reformat noise | Acknowledged reviewability concern; no semantic correction needed. |
| Claude #13: metadata size bound | `8f1d059` states metadata counts toward full representation bytes, not properties-only bytes. No new limit key added. |

## Verification-order defect and exact correction

The authored `21522d4` did **not** pass required typecheck. After successful checks, an import consolidation used an unrestricted text replacement; it also inserted `createCurrentReadWireValidator` into `IN_PROCESS_ONLY_CAPABILITIES: Set<string>`. Only formatting ran after that edit. The author's handoff incorrectly attributed earlier successful typecheck/build to the committed source. This was a verification-order error, not a flaky check.

Native reproduced TS2769 (`native-read-typecheck.log`). Commit `3e7b672` deletes only the accidental set entry, retaining import and observer wiring. Exact pinned Node 24.16.0 / pnpm 11.20.0 typecheck passed with explicit exit 0 in `../read-metadata-typecheck-correction.log`. Root integrated the correction as `2a4ede9`; `repaired-typecheck.log` and `repaired-build.log` record successful exact-source checks, also confirmed by root. Author's 210-test affected run and native's 81-test run remain accurately scoped; neither substitutes for fresh packaged evidence.

The attempted cohort at `c862d10616bb57c918be32d940e1a5a1c41d4a16` contains the rejected source. Its preserved `c862-attempt-unpublished.json` is an **unsealed, unqualified attempt** and must not be reused or relabeled. Root is generating a new cohort from corrected source; no success is inferred from a running log.

## Outstanding review and integration coverage

- At this checkpoint `claude-repair-review.txt` is empty. Its supplied prompt covers frozen `1af8d4d3a11f6e90f6528fa78e54ec005c56f8bd`, the immutability/Event/coherence/diagnostics repairs. It expressly excludes the separately authored Read enforcement and selector repairs. A completed response cannot be extended to `21522d4`/`3e7b672` or their integration without an explicit supplement. Original Claude review and native addenda are not a completed dual-provider review of the corrected Read pair.
- Native reviewed all 13 files of `21522d4` and independently reproduced its typecheck defect. The one-line correction and root's successful exact checks resolve that defect; this record does not represent a new native full-suite or packaged-cohort run. Corrected-source selected-pair external review and final combined source review remain separately tracked by root.
- Fresh corrected-source Read observations, two-file artifact/constant seal, binding verification, and final integration checks remain root-owned. Existing schema projection closure is still 47 definitions with unchanged digest; direct roots shrink 20→16. Manifest/observer binding changes still require fresh observations.
- `native-diagnostic-addendum.md` identifies PR83 `packages/server/src/transaction-batch.ts:439–442` at frozen `8c78d3e198561c242286a185cd8cb0424f9b5b83`: the direct numeric-admission formatter bypasses repaired member formatting. On integration, preserve document-relative create/update metadata diagnostics and test exact byte budgets. This is a downstream integration dependency, not a PR64 defect or already verified fix.
- PR65 provenance and four-part editorial plan findings in the original Claude report are separate lanes. This report does not certify their source/evidence migrations, citation ownership, retained-form rules, identity fences or final PR83 composition. No write-profile advertisement or CLI parity claim follows from Read qualification.

No source checkout, evidence artifact, capability constant, PR or coordination record was changed to produce this report.
