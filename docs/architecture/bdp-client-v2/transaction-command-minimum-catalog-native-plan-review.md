# Native review: command-minimum catalog mapping plan

## Findings

1. **Medium — Required Read+Update test fanout is missing from the seven-path scope.** Plan lines15–29 and39 defer `apps/bdptest/src/read-update-matrix.test.ts` to later executable proof, but the omitted catalog row immediately changes its existing cumulative reports. On the frozen source, uncovered-required counts at lines1266,1359,2021,2442,2872 are96 and become97; omittedRequiredIds at1284 is75 and becomes76. The report computes uncovered IDs from the full required selection minus whole-contract coverage (lines955–961), and progress uses the same canonical inventory (1349–1359). Leaving these assertions unchanged makes existing CI cases fail even though no operation, session or child is added. **Correction:** explicitly include this eighth path, update only these derived inventory counts and appropriate new-ID omission membership checks, preserve authored54/reviewedWhole33 and all actual traffic/child counts, and describe this as immediate metadata fanout rather than future executable work.

2. **Medium — Missing unplanned-obligation pin and selected CPU-only case.** Plan line24 changes only required250→251. `packages/conformance/src/transactional-cumulative.test.ts:189` also pins237 unplanned obligations in “copies source bytes before binding and exposes all missing obligations without executing them” (173–194). The extra required omitted row makes this238; the13 implemented manifest rows remain unchanged. The seven-case selection at plan33 excludes this affected case, so the local gate would miss a deterministic CI failure. **Correction:** include237→238 in the existing owned path and add that existing CPU-only case to the exact selection. This makes at least eight selected cases before the source-growth checks below; recount the final explicit selection rather than retaining seven.

3. **Medium — New coverage row necessarily exceeds an existing source-growth fence.** Plan lines22,33–35 do not account for the inherited evidence-authoring fences in `apps/bdptest/src/read-update-matrix.test.ts`. Current canonical RU coverage is712792 bytes. The metadata-closure fence at4442/4579–4584 allows698614+14336=712950, leaving158 bytes. Even a minimal normally formatted omitted row with the proposed ID, required fields and a one-character reason adds167 bytes (lower bound; the required honest explanatory reason is longer). Therefore the planned insertion necessarily fails that existing static check. The R-closure fence at4962 allows703415+9728=713143, leaving351 bytes and may also be crossed by the actual wording. Other inspected coverage fences permit714717 at4194 and714369 at5105. **Correction:** add a draft-only measurement stop covering all four simultaneous coverage fences; freeze exact normal-form row bytes and totals, then seek a measured root/council-approved amendment only to exceeded authoring allowances. Preserve historical bases/comments and all unrelated bytes; do not shorten old evidence, minify, reset bases, or silently increase caps. Include the affected existing CPU-only growth checks in the bounded selection or explicitly bind equivalent measured preflight checks before release. No runtime budget increase is implied.

## Other conclusions

The proposed single `read-update.http.command-minimum-fields` normative ID is appropriate for the current shared authority at `docs/specs/bdp.md:3579–3586`. Its requiredProfile read-update makes it cumulative in TX and excluded from Read; no duplicate TX ID or retirement is needed. The TX backward reference at5797–5801 expressly retains receipt GET/HEAD read minimum behavior. The plan correctly avoids changing operative semantics or tying the ignore rule only to copied headers.

Omitted/wholeContract:false/bindings:[] in the complete RU coverage inventory is honest and necessary. TX private coverage mirrors implemented manifest rows instead, so not adding a fabricated TX omitted binding is correct. Existing private f2ac runtime evidence cannot be rebound to the new catalog or called cumulative executable coverage. The later proof section appropriately preserves that distinction and raw-empty-header evidence requirements.

The listed seven local cases are six existing registrations plus one proposed new registration. Their selected existing callbacks are static parsing/binding/planning checks; importing a file with other runtime tests does not itself select those tests. I found no HTTP/listener/store/child work in those selected callbacks. This is static inspection, not an observed zero-network run. Findings2–3 require a revised explicit selection and its own static inventory; they do not justify running full installed matrices for metadata changes.

The cumulative artifact digest expectation is recomputed from the current selected catalog digest map, rather than a fixed old catalog hash. Searches for the current RU catalog and coverage hashes in apps/packages/docs/design found no raw old-hash literals to replace. Manifest row/raw-route/configuration hashes must remain unchanged; those artifacts do not change in this slice. Old externally saved reports remain immutable. `docs/design/profile-contract-planning.md:11` describes129RU/250TX as a source snapshot; the revision should either retain it explicitly as historical or update it to the new130/251 snapshot so the newly updated table is not presented with apparently current stale counts. This is editorial, not a fourth blocker.

**Disposition:** revise the plan before authoring release. No semantic/product ruling is needed. Selector ownership and prioritization remain unchanged. There is no demand here for full-profile qualification from a deliberately omitted metadata row.

## Exact basis and limitations

Reviewed plan SHA256 `d0e30e37003a96ac922090e8d90dccdd1dc439cbb16f17e77440bff46cd6221d` in full. Inspected checkout `/Users/dbox/repos/wt/vickie/vickie-command-minimum`; actual HEAD `727e919c2893bd34d5e72ea4659302a999bc4a6f`, tree `734bf9daf40fd17d0fe2bf86b085598589122f2f`, clean status. Root's merged-tree statement is contextual; this review independently binds the local frozen source, not a new remote merge check.

Key inspected SHA256 bindings:

| File | SHA256 |
|---|---|
| catalog/read-update-v1.json | 0ad91810510dd7b688b291863d9b33949bba85fab69d6004ef1423b8c5e47039 |
| test-support/read-update-coverage-v1.json | ee39ca50dc8f48dd062fa8f2e1364c1842aa3bb2d1f732152dd42646ffa154b1 |
| docs/specs/bdp.md | 287a0a64bea0c1c3be51b5bb50d528283185ef7886a9c7402b4d4b5a7d5a0941 |
| docs/design/profile-contract-planning.md | 1a3b31f11180497f28e4bce69047e51ac7ad26a328c191740cef3b783cd8c75a |
| src/profile-contract-plan.test.ts | ca2f0b36bc0685782a2d692f7d5414c1be056df99dd447583db1bf3a96d95449 |
| src/transactional-cumulative.test.ts | 1c361acc6edc460f1a224331d1010da9555a3700066b1ace41f75842aadf7bc8 |
| apps/bdptest/src/transactional-matrix.test.ts | 5420e0bad74cc3906c72af02027b2de58c698e24c209e2fd9f67d013c5ca5ed2 |
| apps/bdptest/src/read-update-matrix.test.ts | 25559baca1589865de4cde257de028b80385a502da543e754de19b899b7c2941 |
| src/read-update-catalog.test.ts | 5ef01c61a35f45568c241f88f18ea752a3933289a00e6621574b2d8598f6d7c0 |
| src/cumulative-artifact.test.ts | 863535a75697ce5cb8fbf07c8096a0e00ae88bc42549be117e74e776490dff5c |

Abbreviated catalog/test-support/src paths in this table are under packages/conformance. Supporting reads included the coverage factory and selection/planning implementation. Read-only Python parsed data and counted bytes; it did not import product code or run tests. No source edits, tests, builds, installs, services, external review, pushes or coordination posts were performed. Only this report was written, after the session guard.
