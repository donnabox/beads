# Native trusted-startup PLAN review — completed N1 footprint

**Verdict: revisions required for N1; otherwise bounded architectural acceptance.** This completes the same finding after narrow export-map fanout inspection. Native-review-1 remains unchanged as the initial, incomplete footprint report.

1. **[P1 / High] Include all four existing exact export-map enforcement/regression files.** The seven-file plan adds `./reference-configuration`, but these existing checks admit only `.` and `./development`:

   - `scripts/smoke-safety.mjs:5–13`: shared deep-equality allowlist, invoked by packed smoke at `scripts/smoke-executables.mjs:411`.
   - `scripts/smoke-safety.test.mjs:7–20`: positive old exact-map fixture and broad/private export refusals.
   - `packages/server/src/transaction-snapshot.test.ts:675–680`: exact package export keys plus private Snapshot/json-equality facade checks.
   - `packages/server/src/read-sequence-lifecycle.test.ts:1528–1551`: exact package export map alongside real lifecycle/facet privacy assertions.

   Consequently, the proposed package deterministically fails packed smoke and existing full-suite privacy tests. The smallest correction is an explicit **eleven-file** footprint successor: original seven plus those four files. Add only the new data-loader subpath to the exact maps/key lists, targeting `dist/reference-deployment.d.ts/js`. Preserve the shared checker, strict missing/wildcard/private/candidate/runtime/admission refusals, private Snapshot names and lifecycle/facet privacy assertions. Do not strip the new export before checking, weaken equality, duplicate the shared allowlist or touch Snapshot/lifecycle production behavior.

The bounded non-listening loader architecture is otherwise accepted: one shared config source for Scope/listener/profile/six limits; trusted local credentials/artifact projection into existing RU/TX captures; TX compilation/readOptions before any owner/listener; retained intended listener without claiming current privateport0 runtime consumes it; secret-safe errors; unchanged Read/no-serve behavior; separate later lifecycle/public admission and cumulative evidence. Proposed regression and installed external-loader tests address this slice without claiming ordinary Write startup.

Exact clean planning source remains a56254d682cf4623cab22a552e3d2f695f691795 / tree144f278c7f5a08516c4ac99ba553d4cc25bcf560. All19 recorded input hashes were verified; seven directly relevant package/build helpers/tests were additionally inspected, including the two root-directed assertions. Review.json records complete identities and preserved initial-review hashes.

Root can disposition this concrete footprint correction and bind the exact successor for the required actual Claude PLAN review. This introduces no new product decision or review gate. Original plan/contract/source evidence/index and native-review-1 are preserved. No source/design/launch/running-review-packet changes, tests/builds/installs/services/CLI reviewers/runtime or remote writes occurred.
